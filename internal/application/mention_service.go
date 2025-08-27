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
	config           *Config
}

// Config は、アプリケーションサービスの設定を定義します
type Config struct {
	MaxHistoryMessages int
	RequestTimeout     time.Duration
	SystemPrompt       string
}

// DefaultConfig は、デフォルトの設定を返します
func DefaultConfig() *Config {
	return &Config{
		MaxHistoryMessages: 10,
		RequestTimeout:     30 * time.Second,
		SystemPrompt:       "あなたは優秀なアシスタントです。与えられた会話履歴を参考に、ユーザーの質問に適切に回答してください。",
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
		config:           config,
	}
}

// HandleMention は、Botへのメンションを処理します
func (s *MentionApplicationService) HandleMention(ctx context.Context, mention domain.BotMention) (string, error) {
	log.Printf("メンションを処理中: %s", mention.String())

	// コンテキストにタイムアウトを設定
	ctx, cancel := context.WithTimeout(ctx, s.config.RequestTimeout)
	defer cancel()

	// 1. チャット履歴を取得
	history, err := s.getConversationHistory(ctx, mention)
	if err != nil {
		return "", fmt.Errorf("チャット履歴の取得に失敗: %w", err)
	}

	// 2. プロンプトを生成
	prompt := s.promptGenerator.GeneratePromptWithMention(history, mention.Content, mention.DisplayName, mention.UserID.String())

	// 3. Gemini APIにリクエストを送信
	response, err := s.geminiClient.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("Gemini APIからの応答取得に失敗: %w", err)
	}

	log.Printf("Gemini APIからの応答を取得: %d文字", len(response))
	return response, nil
}

// getConversationHistory は、メンションの種類に応じて適切なチャット履歴を取得します
func (s *MentionApplicationService) getConversationHistory(ctx context.Context, mention domain.BotMention) (domain.ConversationHistory, error) {
	if mention.IsThread() {
		// スレッドの場合：全メッセージを取得
		log.Printf("スレッド内の全メッセージを取得中: %s", mention.ChannelID)
		return s.conversationRepo.GetThreadMessages(ctx, mention.ChannelID)
	} else {
		// 通常チャンネルの場合：直近のメッセージを取得
		log.Printf("通常チャンネルの直近%d件のメッセージを取得中: %s", s.config.MaxHistoryMessages, mention.ChannelID)
		return s.conversationRepo.GetRecentMessages(ctx, mention.ChannelID, s.config.MaxHistoryMessages)
	}
}

// ExtractUserQuestion は、メンションからユーザーの質問部分を抽出します
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
