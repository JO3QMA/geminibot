package application

import (
	"context"
	"fmt"
	"log"
	"strings"

	"geminibot/internal/domain"
	"geminibot/internal/infrastructure/config"
)

// MentionApplicationService は、メンションイベントをトリガーに、一連の処理を制御するアプリケーションサービスです
type MentionApplicationService struct {
	conversationRepo    domain.ConversationRepository
	promptGenerator     *domain.PromptGenerator
	geminiClient        GeminiClient
	contextManager      *domain.ContextManager
	config              *config.BotConfig
	apiKeyService       *APIKeyApplicationService
	defaultGeminiConfig *config.GeminiConfig
	geminiClientFactory func(apiKey string) (GeminiClient, error)
}

// NewMentionApplicationService は新しいMentionApplicationServiceインスタンスを作成します
func NewMentionApplicationService(
	conversationRepo domain.ConversationRepository,
	geminiClient GeminiClient,
	botConfig *config.BotConfig,
	apiKeyService *APIKeyApplicationService,
	defaultGeminiConfig *config.GeminiConfig,
	geminiClientFactory func(apiKey string) (GeminiClient, error),
) (*MentionApplicationService, error) {
	if botConfig == nil {
		return nil, fmt.Errorf("BotConfigが指定されていません")
	}

	return &MentionApplicationService{
		conversationRepo:    conversationRepo,
		promptGenerator:     domain.NewPromptGenerator(botConfig.SystemPrompt),
		geminiClient:        geminiClient,
		contextManager:      domain.NewContextManager(botConfig.MaxContextLength, botConfig.MaxHistoryLength),
		config:              botConfig,
		apiKeyService:       apiKeyService,
		defaultGeminiConfig: defaultGeminiConfig,
		geminiClientFactory: geminiClientFactory,
	}, nil
}

// HandleMention は、Botへのメンションを処理します
func (s *MentionApplicationService) HandleMention(ctx context.Context, mention domain.BotMention) (string, error) {
	log.Printf("構造化コンテキストでメンションを処理中: %s", mention.String())

	// コンテキストにタイムアウトを設定
	ctx, cancel := context.WithTimeout(ctx, s.config.RequestTimeout)
	defer cancel()

	// 1. チャット履歴を取得
	history, err := s.getConversationHistory(ctx, mention)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("チャット履歴の取得がタイムアウトしました: %w", err)
		}
		return "", fmt.Errorf("チャット履歴の取得に失敗: %w", err)
	}

	// 2. コンテキスト長制限を適用
	truncatedSystemPrompt := s.contextManager.TruncateSystemPrompt(s.config.SystemPrompt)
	truncatedQuestion := s.contextManager.TruncateUserQuestion(mention.Content)

	// 3. 統計情報をログ出力
	stats := s.contextManager.GetContextStats(truncatedSystemPrompt, history, truncatedQuestion)
	log.Printf("コンテキスト統計: システム=%d文字, 履歴=%d文字, 質問=%d文字, 合計=%d文字, 制限=%d文字, 切り詰め=%v",
		stats.SystemPromptLength, stats.HistoryLength, stats.QuestionLength, stats.TotalLength, stats.MaxContextLength, stats.IsTruncated)

	// 4. サーバー別のAPIキーを使用してGemini APIにリクエストを送信
	response, err := s.generateResponseWithGuildAPIKey(ctx, mention, truncatedSystemPrompt, history, truncatedQuestion)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("Gemini APIからの応答取得がタイムアウトしました: %w", err)
		}
		return "", fmt.Errorf("Gemini APIからの応答取得に失敗: %w", err)
	}

	log.Printf("Gemini APIからの応答を取得: %d文字", len(response))
	return response, nil
}

// GenerateImage は、画像生成を実行します
func (s *MentionApplicationService) GenerateImage(ctx context.Context, request domain.ImageGenerationRequest) (*domain.ImageGenerationResponse, error) {
	log.Printf("MentionApplicationService: 画像生成を開始")
	log.Printf("プロンプト: %s", request.Prompt)

	// デフォルトのGeminiクライアントを使用して画像生成
	result, err := s.geminiClient.GenerateImage(ctx, request)
	if err != nil {
		log.Printf("画像生成に失敗: %v", err)
		return nil, fmt.Errorf("画像生成に失敗: %w", err)
	}

	log.Printf("画像生成完了: %+v", result)
	return result, nil
}

// generateResponseWithGuildAPIKey は、サーバー別のAPIキーを使用してGemini APIにリクエストを送信します
func (s *MentionApplicationService) generateResponseWithGuildAPIKey(
	ctx context.Context,
	mention domain.BotMention,
	systemPrompt string,
	conversationHistory []domain.Message,
	userQuestion string,
) (string, error) {
	// ギルドIDを取得
	guildID := mention.GuildID

	if guildID == "" {
		log.Printf("ギルドIDが取得できないため、デフォルトのAPIキーとモデルを使用")
		return s.geminiClient.GenerateTextWithStructuredContext(ctx, systemPrompt, conversationHistory, userQuestion)
	}

	// ギルド固有のモデル設定を取得
	guildModel, err := s.apiKeyService.GetGuildModel(ctx, guildID)
	if err != nil {
		log.Printf("ギルド %s のモデル設定取得に失敗: %v, デフォルト設定を使用", guildID, err)
		guildModel = "gemini-2.5-pro" // デフォルト
	}

	// ギルド固有のAPIキーがあるかチェック
	hasCustomAPIKey, err := s.apiKeyService.HasGuildAPIKey(ctx, guildID)
	if err != nil {
		log.Printf("ギルド %s のAPIキー確認に失敗: %v, デフォルトのAPIキーを使用", guildID, err)
		return s.geminiClient.GenerateTextWithStructuredContext(ctx, systemPrompt, conversationHistory, userQuestion)
	}

	if hasCustomAPIKey {
		// カスタムAPIキーを使用
		customAPIKey, err := s.apiKeyService.GetGuildAPIKey(ctx, guildID)
		if err != nil {
			log.Printf("ギルド %s のカスタムAPIキー取得に失敗: %v, デフォルトのAPIキーを使用", guildID, err)
			return s.geminiClient.GenerateTextWithStructuredContext(ctx, systemPrompt, conversationHistory, userQuestion)
		}

		log.Printf("ギルド %s 用のカスタムAPIキーとモデル %s を使用", guildID, guildModel)

		// カスタムAPIキーでGeminiクライアントを作成
		customClient, err := s.createGeminiClientWithAPIKey(customAPIKey)
		if err != nil {
			log.Printf("カスタムAPIキーでのGeminiクライアント作成に失敗: %v, デフォルトのAPIキーを使用", err)
			return s.geminiClient.GenerateTextWithStructuredContext(ctx, systemPrompt, conversationHistory, userQuestion)
		}

		return customClient.GenerateTextWithStructuredContext(ctx, systemPrompt, conversationHistory, userQuestion)
	}

	// デフォルトのAPIキーを使用、ただしモデル設定がある場合はそれを使用
	if guildModel != "gemini-2.5-pro" && guildModel != "" {
		log.Printf("デフォルトAPIキーとカスタムモデル %s を使用", guildModel)
		// TODO: 将来的にモデル設定を反映したい場合は、ここでGeminiクライアントの設定を変更
	} else {
		log.Printf("デフォルトAPIキーを使用")
	}
	return s.geminiClient.GenerateTextWithStructuredContext(ctx, systemPrompt, conversationHistory, userQuestion)
}

// createGeminiClientWithAPIKey は、指定されたAPIキーでGeminiクライアントを作成します
func (s *MentionApplicationService) createGeminiClientWithAPIKey(apiKey string) (GeminiClient, error) {
	// ファクトリー関数を使用してカスタムAPIキーでGeminiクライアントを作成
	if s.geminiClientFactory != nil {
		return s.geminiClientFactory(apiKey)
	}
	return nil, fmt.Errorf("Geminiクライアントファクトリーが設定されていません")
}

// getConversationHistory は、メンションに基づいて会話履歴を取得します
func (s *MentionApplicationService) getConversationHistory(ctx context.Context, mention domain.BotMention) ([]domain.Message, error) {
	// スレッドかどうかを判定（簡易的な判定）
	if mention.IsThread() {
		log.Printf("スレッド内のメンションを検出: %s", mention.ChannelID)
		// スレッドの場合は全メッセージを取得
		return s.conversationRepo.GetThreadMessages(ctx, mention.ChannelID)
	} else {
		log.Printf("通常チャンネル内のメンションを検出: %s", mention.ChannelID)
		// 通常チャンネルの場合は直近のメッセージを取得
		return s.conversationRepo.GetRecentMessages(ctx, mention.ChannelID, 10)
	}
}

// truncateResponse は、Discordのメッセージ長制限に合わせて応答を切り詰めます
func (s *MentionApplicationService) truncateResponse(response string) string {
	const DiscordMessageLimit = 2000

	if len(response) <= DiscordMessageLimit {
		return response
	}

	// 文字数制限を超えている場合、完全な文で終わるように調整
	truncated := response[:DiscordMessageLimit]
	lastPeriod := strings.LastIndex(truncated, "。")
	if lastPeriod > 0 && lastPeriod > DiscordMessageLimit-100 {
		truncated = truncated[:lastPeriod+1]
	} else {
		// 句点がない場合は、最後の完全な単語で終わるように調整
		lastSpace := strings.LastIndex(truncated, " ")
		if lastSpace > 0 && lastSpace > DiscordMessageLimit-50 {
			truncated = truncated[:lastSpace]
		}
	}

	return truncated + "\n\n（文字数制限により省略されました）"
}
