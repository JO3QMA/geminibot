package application

import (
	"context"
	"testing"
	"time"

	"geminibot/internal/domain"
	"geminibot/internal/infrastructure/config"
)

// ContextManagementMockGeminiClient は、テスト用のGeminiClientモックです
type ContextManagementMockGeminiClient struct{}

func (m *ContextManagementMockGeminiClient) GenerateText(ctx context.Context, prompt domain.Prompt) (string, error) {
	return "従来の方法での応答", nil
}

func (m *ContextManagementMockGeminiClient) GenerateTextWithOptions(ctx context.Context, prompt domain.Prompt, options TextGenerationOptions) (string, error) {
	return "オプション付きでの応答", nil
}

func (m *ContextManagementMockGeminiClient) GenerateTextWithStructuredContext(ctx context.Context, systemPrompt string, conversationHistory []domain.Message, userQuestion string) (string, error) {
	return "構造化コンテキストでの応答", nil
}

// ContextManagementMockConversationRepository は、テスト用のConversationRepositoryモックです
type ContextManagementMockConversationRepository struct{}

func (m *ContextManagementMockConversationRepository) GetRecentMessages(ctx context.Context, channelID domain.ChannelID, limit int) (domain.ConversationHistory, error) {
	messages := []domain.Message{
		{
			ID: "msg1",
			User: domain.User{
				ID:          "user1",
				Username:    "testuser1",
				DisplayName: "TestUser1",
			},
			Content:   "こんにちは",
			Timestamp: time.Now(),
		},
	}
	return domain.NewConversationHistory(messages), nil
}

func (m *ContextManagementMockConversationRepository) GetThreadMessages(ctx context.Context, threadID domain.ChannelID) (domain.ConversationHistory, error) {
	return m.GetRecentMessages(ctx, threadID, 10)
}

func (m *ContextManagementMockConversationRepository) GetMessagesBefore(ctx context.Context, channelID domain.ChannelID, messageID string, limit int) (domain.ConversationHistory, error) {
	return m.GetRecentMessages(ctx, channelID, limit)
}

func TestMentionApplicationService_ContextManagement(t *testing.T) {
	// テスト用の設定
	config := &config.BotConfig{
		MaxContextLength: 100, // 小さな制限を設定
		MaxHistoryLength: 50,  // 小さな制限を設定
		RequestTimeout:   30 * time.Second,
		SystemPrompt:     "テストシステムプロンプト",
	}

	// モッククライアントとリポジトリを作成
	mockClient := &ContextManagementMockGeminiClient{}
	mockRepo := &ContextManagementMockConversationRepository{}

	// アプリケーションサービスを作成
	service := &MentionApplicationService{
		conversationRepo: mockRepo,
		promptGenerator:  domain.NewPromptGenerator(config.SystemPrompt),
		geminiClient:     mockClient,
		contextManager:   domain.NewContextManager(config.MaxContextLength, config.MaxHistoryLength),
		config:           config,
	}

	// テスト用のメンションを作成
	mention := domain.BotMention{
		User: domain.User{
			ID:            "testuser",
			Username:      "testuser",
			Avatar:        "",
			Discriminator: "",
			IsBot:         false,
			DisplayName:   "TestUser",
		},
		Content:   "テストメッセージ",
		ChannelID: domain.NewChannelID("testchannel"),
		MessageID: "testmessageid",
	}

	// コンテキスト管理機能をテスト
	ctx := context.Background()
	response, err := service.HandleMention(ctx, mention)

	if err != nil {
		t.Errorf("メンション処理でエラーが発生しました: %v", err)
	}

	if response != "構造化コンテキストでの応答" {
		t.Errorf("期待される応答: '構造化コンテキストでの応答', 実際の応答: %s", response)
	}
}

func TestMentionApplicationService_GetConversationHistory(t *testing.T) {
	// テスト用の設定
	config := &config.BotConfig{
		MaxContextLength: 8000,
		MaxHistoryLength: 4000,
		RequestTimeout:   30 * time.Second,
		SystemPrompt:     "テストシステムプロンプト",
	}

	mockClient := &ContextManagementMockGeminiClient{}
	mockRepo := &ContextManagementMockConversationRepository{}

	service := &MentionApplicationService{
		conversationRepo: mockRepo,
		promptGenerator:  domain.NewPromptGenerator(config.SystemPrompt),
		geminiClient:     mockClient,
		contextManager:   domain.NewContextManager(config.MaxContextLength, config.MaxHistoryLength),
		config:           config,
	}

	mention := domain.BotMention{
		User: domain.User{
			ID:            "testuser",
			Username:      "testuser",
			DisplayName:   "TestUser",
			Avatar:        "",
			Discriminator: "",
			IsBot:         false,
		},
		Content:   "テストメッセージ",
		ChannelID: domain.NewChannelID("testchannel"),
		MessageID: "testmessageid",
	}

	// getConversationHistoryメソッドをテスト
	ctx := context.Background()
	history, err := service.getConversationHistory(ctx, mention)

	if err != nil {
		t.Errorf("会話履歴の取得でエラーが発生しました: %v", err)
	}

	if history.IsEmpty() {
		t.Error("会話履歴が空です")
	}

	// コンテキスト長制限が適用されていることを確認
	messages := history.Messages()
	if len(messages) == 0 {
		t.Error("メッセージが取得されていません")
	}
}

func TestMentionApplicationService_ContextTruncation(t *testing.T) {
	// 非常に小さな制限を設定
	config := &config.BotConfig{
		MaxContextLength: 10, // 非常に小さな制限
		MaxHistoryLength: 5,  // 非常に小さな制限
		RequestTimeout:   30 * time.Second,
		SystemPrompt:     "これは非常に長いシステムプロンプトです。制限を超える長さです。",
	}

	mockClient := &ContextManagementMockGeminiClient{}
	mockRepo := &ContextManagementMockConversationRepository{}

	service := &MentionApplicationService{
		conversationRepo: mockRepo,
		promptGenerator:  domain.NewPromptGenerator(config.SystemPrompt),
		geminiClient:     mockClient,
		contextManager:   domain.NewContextManager(config.MaxContextLength, config.MaxHistoryLength),
		config:           config,
	}

	mention := domain.BotMention{
		User: domain.User{
			ID:          "testuser",
			Username:    "testuser",
			DisplayName: "TestUser",
		},
		Content:   "これは非常に長いユーザーの質問です。制限を超える長さです。",
		ChannelID: domain.NewChannelID("testchannel"),
		MessageID: "testmessageid",
	}

	// コンテキスト切り詰め機能をテスト
	ctx := context.Background()
	response, err := service.HandleMention(ctx, mention)

	if err != nil {
		t.Errorf("メンション処理でエラーが発生しました: %v", err)
	}

	if response != "構造化コンテキストでの応答" {
		t.Errorf("期待される応答: '構造化コンテキストでの応答', 実際の応答: %s", response)
	}
}
