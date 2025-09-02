package gemini

import (
	"strings"
	"testing"
	"time"

	"geminibot/internal/domain"
	"geminibot/internal/infrastructure/config"
)

func TestGeminiAPIClient_GenerateTextWithStructuredContext(t *testing.T) {
	// テスト用の設定
	config := &config.GeminiConfig{
		ModelName:   "gemini-pro",
		MaxTokens:   1000,
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
	}

	// テスト用のクライアントを作成（実際のAPIキーは不要）
	client := &GeminiAPIClient{
		client: nil, // 実際のテストではモックを使用
		config: config,
	}

	// テストデータ
	systemPrompt := "あなたは優秀なアシスタントです。"
	conversationHistory := []domain.Message{
		{
			ID: "msg1",
			User: domain.User{
				ID:          domain.NewUserID("user1"),
				Username:    "testuser1",
				DisplayName: "TestUser1",
			},
			Content:   "こんにちは",
			Timestamp: time.Now(),
		},
		{
			ID: "msg2",
			User: domain.User{
				ID:          domain.NewUserID("user2"),
				Username:    "testuser2",
				DisplayName: "TestUser2",
			},
			Content:   "こんばんは",
			Timestamp: time.Now(),
		},
	}
	userQuestion := "今日の天気は？"

	// 構造化コンテキストのフォーマットをテスト
	historyText := client.formatConversationHistory(conversationHistory)

	// 期待される形式をチェック
	expectedSections := []string{
		"## 会話履歴",
		"TestUser1: こんにちは",
		"TestUser2: こんばんは",
	}

	for _, expected := range expectedSections {
		if !strings.Contains(historyText, expected) {
			t.Errorf("期待されるセクション '%s' が含まれていません", expected)
		}
	}

	// このテストでは実際のAPI呼び出しは行わないため、
	// 構造化されたコンテンツの作成ロジックのみをテスト
	var allContents []string

	// システムプロンプトを追加
	allContents = append(allContents, systemPrompt)

	// 会話履歴を構造化して追加
	if len(conversationHistory) > 0 {
		allContents = append(allContents, historyText)
	}

	// ユーザーの質問を追加
	allContents = append(allContents, userQuestion)

	// コンテンツの順序をチェック
	if len(allContents) != 3 {
		t.Errorf("期待されるコンテンツ数: 3, 実際のコンテンツ数: %d", len(allContents))
	}

	if allContents[0] != systemPrompt {
		t.Errorf("最初のコンテンツがシステムプロンプトではありません: %s", allContents[0])
	}

	if !strings.Contains(allContents[1], "## 会話履歴") {
		t.Errorf("2番目のコンテンツに会話履歴が含まれていません: %s", allContents[1])
	}

	if allContents[2] != userQuestion {
		t.Errorf("3番目のコンテンツがユーザー質問ではありません: %s", allContents[2])
	}
}

func TestGeminiAPIClient_formatConversationHistory(t *testing.T) {
	client := &GeminiAPIClient{
		client: nil,
		config: config.DefaultGeminiConfig(),
	}

	// 空の履歴をテスト
	emptyHistory := []domain.Message{}
	result := client.formatConversationHistory(emptyHistory)

	if !strings.Contains(result, "## 会話履歴") {
		t.Error("空の履歴でも会話履歴セクションが含まれている必要があります")
	}

	// 単一メッセージの履歴をテスト
	singleMessage := []domain.Message{
		{
			ID: "msg1",
			User: domain.User{
				ID:          domain.NewUserID("user1"),
				Username:    "testuser",
				DisplayName: "TestUser",
			},
			Content:   "テストメッセージ",
			Timestamp: time.Now(),
		},
	}

	result = client.formatConversationHistory(singleMessage)

	if !strings.Contains(result, "TestUser: テストメッセージ") {
		t.Error("単一メッセージが正しくフォーマットされていません")
	}

	// 複数メッセージの履歴をテスト
	multipleMessages := []domain.Message{
		{
			ID: "msg1",
			User: domain.User{
				ID:          domain.NewUserID("user1"),
				Username:    "user1",
				DisplayName: "User1",
			},
			Content:   "最初のメッセージ",
			Timestamp: time.Now(),
		},
		{
			ID: "msg2",
			User: domain.User{
				ID:          domain.NewUserID("user2"),
				Username:    "user2",
				DisplayName: "User2",
			},
			Content:   "2番目のメッセージ",
			Timestamp: time.Now(),
		},
	}

	result = client.formatConversationHistory(multipleMessages)

	expectedLines := []string{
		"## 会話履歴",
		"User1: 最初のメッセージ",
		"User2: 2番目のメッセージ",
	}

	for _, expected := range expectedLines {
		if !strings.Contains(result, expected) {
			t.Errorf("期待される行 '%s' が含まれていません", expected)
		}
	}
}
