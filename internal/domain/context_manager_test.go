package domain

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestNewContextManager(t *testing.T) {
	manager := NewContextManager(8000, 4000)

	if manager.maxContextLength != 8000 {
		t.Errorf("期待されるmaxContextLength: 8000, 実際の値: %d", manager.maxContextLength)
	}

	if manager.maxHistoryLength != 4000 {
		t.Errorf("期待されるmaxHistoryLength: 4000, 実際の値: %d", manager.maxHistoryLength)
	}
}

func TestContextManager_TruncateConversationHistory_Empty(t *testing.T) {
	manager := NewContextManager(8000, 4000)
	emptyHistory := NewConversationHistory([]Message{})

	result := manager.TruncateConversationHistory(emptyHistory)

	if !result.IsEmpty() {
		t.Error("空の履歴は空のままである必要があります")
	}
}

func TestContextManager_TruncateConversationHistory_WithinLimit(t *testing.T) {
	manager := NewContextManager(8000, 4000)

	messages := []Message{
		{
			ID: "msg1",
			User: User{
				ID:          "user1",
				Username:    "testuser1",
				DisplayName: "TestUser1",
			},
			Content:   "短いメッセージ",
			Timestamp: time.Now(),
		},
		{
			ID: "msg2",
			User: User{
				ID:          "user2",
				Username:    "testuser2",
				DisplayName: "TestUser2",
			},
			Content:   "もう一つの短いメッセージ",
			Timestamp: time.Now(),
		},
	}

	history := NewConversationHistory(messages)
	result := manager.TruncateConversationHistory(history)

	if result.Count() != 2 {
		t.Errorf("制限内の履歴は切り詰められてはいけません。期待される件数: 2, 実際の件数: %d", result.Count())
	}
}

func TestContextManager_TruncateConversationHistory_ExceedsLimit(t *testing.T) {
	manager := NewContextManager(8000, 100) // 小さな制限を設定

	// 長いメッセージを作成
	longMessage := strings.Repeat("これは非常に長いメッセージです。", 50)

	messages := []Message{
		{
			ID: "msg1",
			User: User{
				ID:          "user1",
				Username:    "testuser1",
				DisplayName: "TestUser1",
			},
			Content:   longMessage,
			Timestamp: time.Now().Add(-time.Hour), // 古いメッセージ
		},
		{
			ID: "msg2",
			User: User{
				ID:          "user2",
				Username:    "testuser2",
				DisplayName: "TestUser2",
			},
			Content:   "新しい短いメッセージ",
			Timestamp: time.Now(), // 新しいメッセージ
		},
	}

	history := NewConversationHistory(messages)
	result := manager.TruncateConversationHistory(history)

	// 新しいメッセージが優先的に保持されることを確認
	if result.Count() != 1 {
		t.Errorf("制限を超えた履歴は切り詰められる必要があります。期待される件数: 1, 実際の件数: %d", result.Count())
	}

	resultMessages := result.Messages()
	if len(resultMessages) > 0 && resultMessages[0].Content != "新しい短いメッセージ" {
		t.Error("新しいメッセージが優先的に保持される必要があります")
	}
}

func TestContextManager_TruncateSystemPrompt(t *testing.T) {
	manager := NewContextManager(50, 4000)

	// 制限内のプロンプト
	shortPrompt := "短いシステムプロンプト"
	result := manager.TruncateSystemPrompt(shortPrompt)
	if result != shortPrompt {
		t.Errorf("制限内のプロンプトは変更されません。期待される値: %s, 実際の値: %s", shortPrompt, result)
	}

	// 制限を超えるプロンプト（rune数で制限を超える）
	longPrompt := "これは非常に長いシステムプロンプトです。制限を超える長さのプロンプトです。これは非常に長いシステムプロンプトです。制限を超える長さのプロンプトです。これは非常に長いシステムプロンプトです。制限を超える長さのプロンプトです。"
	result = manager.TruncateSystemPrompt(longPrompt)
	if utf8.RuneCountInString(result) >= utf8.RuneCountInString(longPrompt) {
		t.Error("制限を超えるプロンプトは切り詰められる必要があります")
	}
}

func TestContextManager_TruncateUserQuestion(t *testing.T) {
	manager := NewContextManager(50, 4000)

	// 制限内の質問
	shortQuestion := "短い質問"
	result := manager.TruncateUserQuestion(shortQuestion)
	if result != shortQuestion {
		t.Errorf("制限内の質問は変更されません。期待される値: %s, 実際の値: %s", shortQuestion, result)
	}

	// 制限を超える質問（rune数で制限を超える）
	longQuestion := "これは非常に長い質問です。制限を超える長さの質問です。これは非常に長い質問です。制限を超える長さの質問です。これは非常に長い質問です。制限を超える長さの質問です。"
	result = manager.TruncateUserQuestion(longQuestion)
	if utf8.RuneCountInString(result) >= utf8.RuneCountInString(longQuestion) {
		t.Error("制限を超える質問は切り詰められる必要があります")
	}
}

func TestContextManager_GetContextStats(t *testing.T) {
	manager := NewContextManager(8000, 4000)

	systemPrompt := "システムプロンプト"
	messages := []Message{
		{
			ID: "msg1",
			User: User{
				ID:          "user1",
				Username:    "testuser1",
				DisplayName: "TestUser1",
			},
			Content:   "テストメッセージ",
			Timestamp: time.Now(),
		},
	}
	history := NewConversationHistory(messages)
	userQuestion := "ユーザーの質問"

	stats := manager.GetContextStats(systemPrompt, history, userQuestion)

	if stats.SystemPromptLength <= 0 {
		t.Error("システムプロンプトの長さは正の値である必要があります")
	}

	if stats.HistoryLength <= 0 {
		t.Error("履歴の長さは正の値である必要があります")
	}

	if stats.QuestionLength <= 0 {
		t.Error("質問の長さは正の値である必要があります")
	}

	if stats.TotalLength != stats.SystemPromptLength+stats.HistoryLength+stats.QuestionLength {
		t.Error("総長は各部分の長さの合計である必要があります")
	}

	if stats.MaxContextLength != 8000 {
		t.Errorf("期待されるMaxContextLength: 8000, 実際の値: %d", stats.MaxContextLength)
	}

	if stats.MaxHistoryLength != 4000 {
		t.Errorf("期待されるMaxHistoryLength: 4000, 実際の値: %d", stats.MaxHistoryLength)
	}
}

func TestContextManager_calculateHistoryLength(t *testing.T) {
	manager := NewContextManager(8000, 4000)

	messages := []Message{
		{
			ID: "msg1",
			User: User{
				ID:          "user1",
				Username:    "testuser1",
				DisplayName: "TestUser1",
			},
			Content:   "テストメッセージ",
			Timestamp: time.Now(),
		},
	}

	length := manager.calculateHistoryLength(messages)
	expectedLength := utf8.RuneCountInString("TestUser1") + 2 + utf8.RuneCountInString("テストメッセージ") + 1

	if length != expectedLength {
		t.Errorf("期待される履歴長: %d, 実際の履歴長: %d", expectedLength, length)
	}
}
