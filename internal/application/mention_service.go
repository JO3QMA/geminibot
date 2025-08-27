package application

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"geminibot/internal/domain"
)

// MentionApplicationService は、メンションイベントをトリガーに、一連の処理を制御するアプリケーションサービスです
type MentionApplicationService struct {
	conversationRepo domain.ConversationRepository
	promptGenerator  *domain.PromptGenerator
	geminiClient     GeminiClient
	contextManager   *domain.ContextManager
	config           *Config
}

// Config は、アプリケーションサービスの設定を定義します
type Config struct {
	MaxContextLength     int // 最大コンテキスト長（文字数）
	MaxHistoryLength     int // 最大履歴長（文字数）
	RequestTimeout       time.Duration
	SystemPrompt         string
	UseStructuredContext bool // 構造化コンテキストを使用するかどうか
}

// DefaultConfig は、デフォルトの設定を返します
func DefaultConfig() *Config {
	return &Config{
		MaxContextLength:     8000,
		MaxHistoryLength:     4000,
		RequestTimeout:       30 * time.Second,
		SystemPrompt:         "あなたは優秀なアシスタントです。与えられた会話履歴を参考に、ユーザーのチャット内容に適切に回答してください。",
		UseStructuredContext: true, // デフォルトで構造化コンテキストを使用
	}
}

// NewMentionApplicationService は新しいMentionApplicationServiceインスタンスを作成します
func NewMentionApplicationService(
	conversationRepo domain.ConversationRepository,
	geminiClient GeminiClient,
	config *Config,
) *MentionApplicationService {
	if config == nil {
		config = DefaultConfig()
	}

	return &MentionApplicationService{
		conversationRepo: conversationRepo,
		promptGenerator:  domain.NewPromptGenerator(config.SystemPrompt),
		geminiClient:     geminiClient,
		contextManager:   domain.NewContextManager(config.MaxContextLength, config.MaxHistoryLength),
		config:           config,
	}
}

// HandleMention は、Botへのメンションを処理します
func (s *MentionApplicationService) HandleMention(ctx context.Context, mention domain.BotMention) (string, error) {
	// 設定に基づいて構造化コンテキストを使用するかどうかを決定
	if s.config.UseStructuredContext {
		return s.HandleMentionWithStructuredContext(ctx, mention)
	}

	log.Printf("従来の方法でメンションを処理中: %s", mention.String())

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

	// 2. プロンプトを生成
	prompt := s.promptGenerator.GeneratePromptWithMention(history, mention.Content, mention.User.GetDisplayName(), mention.User.ID.String())

	// 3. Gemini APIにリクエストを送信
	response, err := s.geminiClient.GenerateText(ctx, prompt)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("Gemini APIからの応答取得がタイムアウトしました: %w", err)
		}
		return "", fmt.Errorf("Gemini APIからの応答取得に失敗: %w", err)
	}

	log.Printf("Gemini APIからの応答を取得: %d文字", len(response))
	return response, nil
}

// HandleMentionWithStructuredContext は、構造化されたコンテキストを使用してメンションを処理します
func (s *MentionApplicationService) HandleMentionWithStructuredContext(ctx context.Context, mention domain.BotMention) (string, error) {
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
	log.Printf("コンテキスト統計: 総長=%d, 履歴長=%d, 制限=%d, 切り詰め=%v",
		stats.TotalLength, stats.HistoryLength, stats.MaxContextLength, stats.IsTruncated)

	// 4. 構造化コンテキストを使用してGemini APIにリクエストを送信
	response, err := s.geminiClient.GenerateTextWithStructuredContext(
		ctx,
		truncatedSystemPrompt,
		history.Messages(),
		truncatedQuestion,
	)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("Gemini APIからの応答取得がタイムアウトしました: %w", err)
		}
		return "", fmt.Errorf("Gemini APIからの応答取得に失敗: %w", err)
	}

	log.Printf("構造化コンテキストでGemini APIからの応答を取得: %d文字", len(response))
	return response, nil
}

// getConversationHistory は、メンションの種類に応じて適切なチャット履歴を取得します
func (s *MentionApplicationService) getConversationHistory(ctx context.Context, mention domain.BotMention) (domain.ConversationHistory, error) {
	// 十分な数のメッセージを取得（コンテキスト長制限で調整される）
	const defaultMessageLimit = 100

	if mention.IsThread() {
		// スレッドの場合：全メッセージを取得
		log.Printf("スレッド内の全メッセージを取得中: %s", mention.ChannelID)
		history, err := s.conversationRepo.GetThreadMessages(ctx, mention.ChannelID)
		if err != nil {
			return domain.ConversationHistory{}, err
		}
		// コンテキスト長制限を適用
		return s.contextManager.TruncateConversationHistory(history), nil
	} else {
		// 通常チャンネルの場合：十分な数のメッセージを取得
		log.Printf("通常チャンネルの直近%d件のメッセージを取得中: %s", defaultMessageLimit, mention.ChannelID)
		history, err := s.conversationRepo.GetRecentMessages(ctx, mention.ChannelID, defaultMessageLimit)
		if err != nil {
			return domain.ConversationHistory{}, err
		}
		// コンテキスト長制限を適用
		return s.contextManager.TruncateConversationHistory(history), nil
	}
}

// ExtractUserQuestion は、メンションからユーザーのチャット内容部分を抽出します
func (s *MentionApplicationService) ExtractUserQuestion(mention domain.BotMention) string {
	content := strings.TrimSpace(mention.Content)

	// Botのメンション部分を除去（例: "@Bot " を除去）
	// 実際の実装では、Botの名前やIDに基づいて適切に除去する必要があります
	content = strings.TrimPrefix(content, "@")

	// 最初のスペース以降を取得
	if idx := strings.Index(content, " "); idx != -1 {
		content = strings.TrimSpace(content[idx+1:])
	}

	return content
}
