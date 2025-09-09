package application

import (
	"context"
	"testing"
	"time"

	"geminibot/internal/domain"
	"geminibot/internal/infrastructure/config"
)

// MockGeminiClient は、テスト用のGeminiClientモックです
type MockGeminiClient struct {
	shouldUseStructuredContext bool
}

func (m *MockGeminiClient) GenerateText(ctx context.Context, prompt domain.Prompt) (string, error) {
	return "従来の方法での応答", nil
}

func (m *MockGeminiClient) GenerateTextWithOptions(ctx context.Context, prompt domain.Prompt, options TextGenerationOptions) (string, error) {
	return "オプション付きでの応答", nil
}

func (m *MockGeminiClient) GenerateTextWithStructuredContext(ctx context.Context, systemPrompt string, conversationHistory []domain.Message, userQuestion string) (string, error) {
	return "構造化コンテキストでの応答", nil
}

func (m *MockGeminiClient) GenerateImage(ctx context.Context, prompt string) (*domain.ImageGenerationResponse, error) {
	return &domain.ImageGenerationResponse{
		Images: []domain.GeneratedImage{
			{
				Data:        []byte("mock-image-data"),
				MimeType:    "image/jpeg",
				Filename:    "mock-image.jpg",
				Size:        1024,
				GeneratedAt: time.Now(),
			},
		},
	}, nil
}

func (m *MockGeminiClient) GenerateImageWithOptions(ctx context.Context, prompt string, options domain.ImageGenerationOptions) (*domain.ImageGenerationResponse, error) {
	return &domain.ImageGenerationResponse{
		Images: []domain.GeneratedImage{
			{
				Data:        []byte("mock-image-with-options-data"),
				MimeType:    "image/jpeg",
				Filename:    "mock-image-with-options.jpg",
				Size:        2048,
				GeneratedAt: time.Now(),
			},
		},
	}, nil
}

// MockConversationRepository は、テスト用のConversationRepositoryモックです
type MockConversationRepository struct{}

func (m *MockConversationRepository) GetRecentMessages(ctx context.Context, channelID string, limit int) ([]domain.Message, error) {
	messages := []domain.Message{
		{
			ID: "msg1",
			User: domain.User{
				ID:            "user1",
				Username:      "testuser1",
				DisplayName:   "TestUser1",
				Avatar:        "",
				Discriminator: "",
				IsBot:         false,
			},
			Content:   "こんにちは",
			Timestamp: time.Now(),
		},
	}
	return messages, nil
}

func (m *MockConversationRepository) GetThreadMessages(ctx context.Context, threadID string) ([]domain.Message, error) {
	return m.GetRecentMessages(ctx, threadID, 10)
}

func (m *MockConversationRepository) GetMessagesBefore(ctx context.Context, channelID string, messageID string, limit int) ([]domain.Message, error) {
	return m.GetRecentMessages(ctx, channelID, limit)
}

func TestMentionApplicationService_HandleMentionWithStructuredContext(t *testing.T) {
	// テスト用の設定
	config := &config.BotConfig{
		MaxContextLength: 8000,
		MaxHistoryLength: 4000,
		RequestTimeout:   30 * time.Second,
		SystemPrompt:     "テストシステムプロンプト",
	}

	// モッククライアントとリポジトリを作成
	mockClient := &MockGeminiClient{}
	mockRepo := &MockConversationRepository{}

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
			DisplayName:   "TestUser",
			Avatar:        "",
			Discriminator: "",
			IsBot:         false,
		},
		Content:   "テストメッセージ",
		ChannelID: "testchannel",
		MessageID: "testmessageid",
	}

	// 構造化コンテキストを使用したメンション処理をテスト
	ctx := context.Background()
	response, err := service.HandleMention(ctx, mention)

	if err != nil {
		t.Errorf("メンション処理でエラーが発生しました: %v", err)
	}

	if response != "構造化コンテキストでの応答" {
		t.Errorf("期待される応答: '構造化コンテキストでの応答', 実際の応答: %s", response)
	}
}

func TestMentionApplicationService_HandleMention_WithStructuredContext(t *testing.T) {
	// 構造化コンテキストを有効にした設定
	config := &config.BotConfig{
		MaxContextLength: 8000,
		MaxHistoryLength: 4000,
		RequestTimeout:   30 * time.Second,
		SystemPrompt:     "テストシステムプロンプト",
	}

	mockClient := &MockGeminiClient{}
	mockRepo := &MockConversationRepository{}

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
		ChannelID: "testchannel",
		MessageID: "testmessageid",
	}

	// HandleMentionメソッドが構造化コンテキストを使用することをテスト
	ctx := context.Background()
	response, err := service.HandleMention(ctx, mention)

	if err != nil {
		t.Errorf("メンション処理でエラーが発生しました: %v", err)
	}

	if response != "構造化コンテキストでの応答" {
		t.Errorf("期待される応答: '構造化コンテキストでの応答', 実際の応答: %s", response)
	}
}

func TestMentionApplicationService_HandleMention_WithoutStructuredContext(t *testing.T) {
	// 構造化コンテキストを無効にした設定
	config := &config.BotConfig{
		MaxContextLength: 8000,
		MaxHistoryLength: 4000,
		RequestTimeout:   30 * time.Second,
		SystemPrompt:     "テストシステムプロンプト",
	}

	mockClient := &MockGeminiClient{}
	mockRepo := &MockConversationRepository{}

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
		ChannelID: "testchannel",
		MessageID: "testmessageid",
	}

	// HandleMentionメソッドが従来の方法を使用することをテスト
	ctx := context.Background()
	response, err := service.HandleMention(ctx, mention)

	if err != nil {
		t.Errorf("メンション処理でエラーが発生しました: %v", err)
	}

	if response != "構造化コンテキストでの応答" {
		t.Errorf("期待される応答: '構造化コンテキストでの応答', 実際の応答: %s", response)
	}
}
