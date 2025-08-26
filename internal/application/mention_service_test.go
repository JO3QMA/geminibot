package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"geminibot/internal/domain"
)

// MockConversationRepository は、テスト用のモックリポジトリです
type MockConversationRepository struct {
	recentMessages domain.ConversationHistory
	threadMessages domain.ConversationHistory
	recentError    error
	threadError    error
}

func (m *MockConversationRepository) GetRecentMessages(ctx context.Context, channelID domain.ChannelID, limit int) (domain.ConversationHistory, error) {
	return m.recentMessages, m.recentError
}

func (m *MockConversationRepository) GetThreadMessages(ctx context.Context, threadID domain.ChannelID) (domain.ConversationHistory, error) {
	return m.threadMessages, m.threadError
}

func (m *MockConversationRepository) GetMessagesBefore(ctx context.Context, channelID domain.ChannelID, messageID string, limit int) (domain.ConversationHistory, error) {
	return domain.NewConversationHistory([]domain.Message{}), nil
}

// MockGeminiClient は、テスト用のモックGeminiクライアントです
type MockGeminiClient struct {
	response string
	error    error
}

func (m *MockGeminiClient) GenerateText(ctx context.Context, prompt domain.Prompt) (string, error) {
	return m.response, m.error
}

func (m *MockGeminiClient) GenerateTextWithOptions(ctx context.Context, prompt domain.Prompt, options TextGenerationOptions) (string, error) {
	return m.response, m.error
}

func TestNewMentionApplicationService(t *testing.T) {
	repo := &MockConversationRepository{}
	client := &MockGeminiClient{}
	config := &Config{
		MaxHistoryMessages: 5,
		RequestTimeout:     10 * time.Second,
		SystemPrompt:       "テストシステムプロンプト",
	}

	service := NewMentionApplicationService(repo, client, config)

	if service.conversationRepo != repo {
		t.Error("リポジトリが正しく設定されていません")
	}
	if service.geminiClient != client {
		t.Error("Geminiクライアントが正しく設定されていません")
	}
	if service.config != config {
		t.Error("設定が正しく設定されていません")
	}
	if service.promptGenerator == nil {
		t.Error("プロンプトジェネレーターが作成されていません")
	}
}

func TestNewMentionApplicationService_DefaultConfig(t *testing.T) {
	repo := &MockConversationRepository{}
	client := &MockGeminiClient{}

	service := NewMentionApplicationService(repo, client, nil)

	if service.config == nil {
		t.Error("デフォルト設定が作成されていません")
	}
	if service.config.MaxHistoryMessages != 10 {
		t.Errorf("期待されるMaxHistoryMessages: 10, 実際: %d", service.config.MaxHistoryMessages)
	}
	if service.config.RequestTimeout != 30*time.Second {
		t.Errorf("期待されるRequestTimeout: 30s, 実際: %v", service.config.RequestTimeout)
	}
}

func TestMentionApplicationService_HandleMention_Success(t *testing.T) {
	// テストデータの準備
	messages := []domain.Message{
		domain.NewMessage("msg1", domain.NewUserID("user1"), "こんにちは", time.Now()),
	}
	history := domain.NewConversationHistory(messages)
	
	repo := &MockConversationRepository{
		recentMessages: history,
	}
	
	client := &MockGeminiClient{
		response: "こんにちは！お元気ですか？",
	}
	
	config := &Config{
		MaxHistoryMessages: 5,
		RequestTimeout:     10 * time.Second,
		SystemPrompt:       "テストシステムプロンプト",
	}
	
	service := NewMentionApplicationService(repo, client, config)
	
	mention := domain.NewBotMention(
		domain.NewChannelID("channel123"),
		domain.NewUserID("user123"),
		"こんにちは",
		"msg123",
	)
	
	ctx := context.Background()
	response, err := service.HandleMention(ctx, mention)
	
	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}
	
	if response != "こんにちは！お元気ですか？" {
		t.Errorf("期待される応答: こんにちは！お元気ですか？, 実際の応答: %s", response)
	}
}

func TestMentionApplicationService_HandleMention_RepositoryError(t *testing.T) {
	repo := &MockConversationRepository{
		recentError: errors.New("リポジトリエラー"),
	}
	
	client := &MockGeminiClient{}
	config := &Config{
		MaxHistoryMessages: 5,
		RequestTimeout:     10 * time.Second,
		SystemPrompt:       "テストシステムプロンプト",
	}
	
	service := NewMentionApplicationService(repo, client, config)
	
	mention := domain.NewBotMention(
		domain.NewChannelID("channel123"),
		domain.NewUserID("user123"),
		"こんにちは",
		"msg123",
	)
	
	ctx := context.Background()
	_, err := service.HandleMention(ctx, mention)
	
	if err == nil {
		t.Error("エラーが発生しませんでした")
	}
	
	if !strings.Contains(err.Error(), "チャット履歴の取得に失敗") {
		t.Errorf("期待されるエラーメッセージに含まれていません: %v", err)
	}
}

func TestMentionApplicationService_HandleMention_GeminiError(t *testing.T) {
	messages := []domain.Message{
		domain.NewMessage("msg1", domain.NewUserID("user1"), "こんにちは", time.Now()),
	}
	history := domain.NewConversationHistory(messages)
	
	repo := &MockConversationRepository{
		recentMessages: history,
	}
	
	client := &MockGeminiClient{
		error: errors.New("Gemini APIエラー"),
	}
	
	config := &Config{
		MaxHistoryMessages: 5,
		RequestTimeout:     10 * time.Second,
		SystemPrompt:       "テストシステムプロンプト",
	}
	
	service := NewMentionApplicationService(repo, client, config)
	
	mention := domain.NewBotMention(
		domain.NewChannelID("channel123"),
		domain.NewUserID("user123"),
		"こんにちは",
		"msg123",
	)
	
	ctx := context.Background()
	_, err := service.HandleMention(ctx, mention)
	
	if err == nil {
		t.Error("エラーが発生しませんでした")
	}
	
	if !strings.Contains(err.Error(), "Gemini APIからの応答取得に失敗") {
		t.Errorf("期待されるエラーメッセージに含まれていません: %v", err)
	}
}

func TestMentionApplicationService_HandleMention_Timeout(t *testing.T) {
	repo := &MockConversationRepository{
		recentMessages: domain.NewConversationHistory([]domain.Message{}),
	}
	
	client := &MockGeminiClient{
		response: "応答",
	}
	
	config := &Config{
		MaxHistoryMessages: 5,
		RequestTimeout:     1 * time.Millisecond, // 非常に短いタイムアウト
		SystemPrompt:       "テストシステムプロンプト",
	}
	
	service := NewMentionApplicationService(repo, client, config)
	
	mention := domain.NewBotMention(
		domain.NewChannelID("channel123"),
		domain.NewUserID("user123"),
		"こんにちは",
		"msg123",
	)
	
	ctx := context.Background()
	_, err := service.HandleMention(ctx, mention)
	
	// タイムアウトが発生する可能性があるが、モックなので通常は成功する
	// このテストは主にタイムアウトコンテキストが正しく設定されているかを確認
	if err != nil && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("予期しないエラー: %v", err)
	}
}

func TestMentionApplicationService_ExtractUserQuestion(t *testing.T) {
	service := NewMentionApplicationService(nil, nil, nil)
	
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "通常のメンション",
			content:  "@bot こんにちは",
			expected: "こんにちは",
		},
		{
			name:     "メンションのみ",
			content:  "@bot",
			expected: "bot",
		},
		{
			name:     "空のコンテンツ",
			content:  "",
			expected: "",
		},
		{
			name:     "スペースのみ",
			content:  "   ",
			expected: "",
		},
		{
			name:     "複数のスペース",
			content:  "@bot    こんにちは",
			expected: "こんにちは",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mention := domain.NewBotMention(
				domain.NewChannelID("channel123"),
				domain.NewUserID("user123"),
				tt.content,
				"msg123",
			)
			
			result := service.ExtractUserQuestion(mention)
			
			if result != tt.expected {
				t.Errorf("期待される結果: %s, 実際の結果: %s", tt.expected, result)
			}
		})
	}
}

func TestMentionApplicationService_GetConversationHistory_Thread(t *testing.T) {
	threadMessages := domain.NewConversationHistory([]domain.Message{
		domain.NewMessage("msg1", domain.NewUserID("user1"), "スレッドメッセージ1", time.Now()),
		domain.NewMessage("msg2", domain.NewUserID("user2"), "スレッドメッセージ2", time.Now()),
	})
	
	repo := &MockConversationRepository{
		recentMessages: threadMessages, // スレッドも通常チャンネルとして扱われるため
	}
	
	client := &MockGeminiClient{}
	config := &Config{
		MaxHistoryMessages: 5,
		RequestTimeout:     10 * time.Second,
		SystemPrompt:       "テストシステムプロンプト",
	}
	
	service := NewMentionApplicationService(repo, client, config)
	
	// スレッドのメンションを作成
	mention := domain.NewBotMention(
		domain.NewChannelID("thread123"),
		domain.NewUserID("user123"),
		"こんにちは",
		"msg123",
	)
	
	// 現在の実装ではIsThread()は常にfalseを返すため、
	// このテストは通常チャンネルとして動作することを確認
	
	ctx := context.Background()
	history, err := service.getConversationHistory(ctx, mention)
	
	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}
	
	if history.Count() != 2 {
		t.Errorf("期待されるメッセージ数: 2, 実際のメッセージ数: %d", history.Count())
	}
}

func TestMentionApplicationService_GetConversationHistory_RegularChannel(t *testing.T) {
	recentMessages := domain.NewConversationHistory([]domain.Message{
		domain.NewMessage("msg1", domain.NewUserID("user1"), "通常メッセージ1", time.Now()),
		domain.NewMessage("msg2", domain.NewUserID("user2"), "通常メッセージ2", time.Now()),
	})
	
	repo := &MockConversationRepository{
		recentMessages: recentMessages,
	}
	
	client := &MockGeminiClient{}
	config := &Config{
		MaxHistoryMessages: 5,
		RequestTimeout:     10 * time.Second,
		SystemPrompt:       "テストシステムプロンプト",
	}
	
	service := NewMentionApplicationService(repo, client, config)
	
	mention := domain.NewBotMention(
		domain.NewChannelID("channel123"),
		domain.NewUserID("user123"),
		"こんにちは",
		"msg123",
	)
	
	ctx := context.Background()
	history, err := service.getConversationHistory(ctx, mention)
	
	if err != nil {
		t.Errorf("エラーが発生しました: %v", err)
	}
	
	if history.Count() != 2 {
		t.Errorf("期待されるメッセージ数: 2, 実際のメッセージ数: %d", history.Count())
	}
}
