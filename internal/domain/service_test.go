package domain

import (
	"strings"
	"testing"
	"time"
)

func TestNewPromptGenerator(t *testing.T) {
	// システムプロンプトを指定した場合
	systemPrompt := "あなたは優秀なアシスタントです。"
	generator := NewPromptGenerator(systemPrompt)

	if generator.systemPrompt != systemPrompt {
		t.Errorf("期待されるシステムプロンプト: %s, 実際のシステムプロンプト: %s", systemPrompt, generator.systemPrompt)
	}

	// 空のシステムプロンプトを指定した場合
	emptyGenerator := NewPromptGenerator("")
	expectedDefault := "あなたは優秀なアシスタントです。与えられた会話履歴を参考に、ユーザーのチャット内容に適切に回答してください。"

	if emptyGenerator.systemPrompt != expectedDefault {
		t.Errorf("期待されるデフォルトシステムプロンプト: %s, 実際のシステムプロンプト: %s", expectedDefault, emptyGenerator.systemPrompt)
	}
}

func TestPromptGenerator_GeneratePrompt_EmptyHistory(t *testing.T) {
	generator := NewPromptGenerator("テストシステムプロンプト")
	history := NewConversationHistory([]Message{})
	userQuestion := "こんにちは"

	prompt := generator.GeneratePrompt(history, userQuestion)
	content := prompt.Content()

	// システムプロンプトが含まれているかチェック
	if !strings.Contains(content, "テストシステムプロンプト") {
		t.Error("生成されたプロンプトにシステムプロンプトが含まれていません")
	}

	// 会話履歴セクションが含まれていないかチェック（空なので）
	if strings.Contains(content, "## 会話履歴") {
		t.Error("空の履歴なのに会話履歴セクションが含まれています")
	}

	// ユーザーのチャット内容が含まれているかチェック
	if !strings.Contains(content, "## ユーザーのチャット内容") {
		t.Error("生成されたプロンプトにユーザーのチャット内容セクションが含まれていません")
	}
	if !strings.Contains(content, userQuestion) {
		t.Error("生成されたプロンプトにユーザーのチャット内容が含まれていません")
	}
}

func TestPromptGenerator_GeneratePrompt_WithHistory(t *testing.T) {
	generator := NewPromptGenerator("テストシステムプロンプト")

	messages := []Message{
		NewMessage("msg1", NewUserID("user1"), "こんにちは", time.Now()),
		NewMessage("msg2", NewUserID("user2"), "こんばんは", time.Now()),
	}
	history := NewConversationHistory(messages)
	userQuestion := "今日の天気は？"

	prompt := generator.GeneratePrompt(history, userQuestion)
	content := prompt.Content()

	// システムプロンプトが含まれているかチェック
	if !strings.Contains(content, "テストシステムプロンプト") {
		t.Error("生成されたプロンプトにシステムプロンプトが含まれていません")
	}

	// 会話履歴セクションが含まれているかチェック
	if !strings.Contains(content, "## 会話履歴") {
		t.Error("生成されたプロンプトに会話履歴セクションが含まれていません")
	}

	// 履歴メッセージが含まれているかチェック
	if !strings.Contains(content, "ユーザーuser1: こんにちは") {
		t.Error("生成されたプロンプトに最初の履歴メッセージが含まれていません")
	}
	if !strings.Contains(content, "ユーザーuser2: こんばんは") {
		t.Error("生成されたプロンプトに2番目の履歴メッセージが含まれていません")
	}

	// ユーザーのチャット内容が含まれているかチェック
	if !strings.Contains(content, "## ユーザーのチャット内容") {
		t.Error("生成されたプロンプトにユーザーのチャット内容セクションが含まれていません")
	}
	if !strings.Contains(content, userQuestion) {
		t.Error("生成されたプロンプトにユーザーのチャット内容が含まれていません")
	}
}

func TestPromptGenerator_GeneratePromptWithContext_EmptyHistory(t *testing.T) {
	generator := NewPromptGenerator("テストシステムプロンプト")
	history := NewConversationHistory([]Message{})
	userQuestion := "こんにちは"
	additionalContext := "今日は晴れです"

	prompt := generator.GeneratePromptWithContext(history, userQuestion, additionalContext)
	content := prompt.Content()

	// システムプロンプトが含まれているかチェック
	if !strings.Contains(content, "テストシステムプロンプト") {
		t.Error("生成されたプロンプトにシステムプロンプトが含まれていません")
	}

	// 追加コンテキストセクションが含まれているかチェック
	if !strings.Contains(content, "## 追加コンテキスト") {
		t.Error("生成されたプロンプトに追加コンテキストセクションが含まれていません")
	}
	if !strings.Contains(content, additionalContext) {
		t.Error("生成されたプロンプトに追加コンテキストが含まれていません")
	}

	// 会話履歴セクションが含まれていないかチェック（空なので）
	if strings.Contains(content, "## 会話履歴") {
		t.Error("空の履歴なのに会話履歴セクションが含まれています")
	}

	// ユーザーのチャット内容が含まれているかチェック
	if !strings.Contains(content, "## ユーザーのチャット内容") {
		t.Error("生成されたプロンプトにユーザーのチャット内容セクションが含まれていません")
	}
	if !strings.Contains(content, userQuestion) {
		t.Error("生成されたプロンプトにユーザーのチャット内容が含まれていません")
	}
}

func TestPromptGenerator_GeneratePromptWithContext_WithHistory(t *testing.T) {
	generator := NewPromptGenerator("テストシステムプロンプト")

	messages := []Message{
		NewMessage("msg1", NewUserID("user1"), "こんにちは", time.Now()),
		NewMessage("msg2", NewUserID("user2"), "こんばんは", time.Now()),
	}
	history := NewConversationHistory(messages)
	userQuestion := "今日の天気は？"
	additionalContext := "今日は晴れです"

	prompt := generator.GeneratePromptWithContext(history, userQuestion, additionalContext)
	content := prompt.Content()

	// システムプロンプトが含まれているかチェック
	if !strings.Contains(content, "テストシステムプロンプト") {
		t.Error("生成されたプロンプトにシステムプロンプトが含まれていません")
	}

	// 追加コンテキストセクションが含まれているかチェック
	if !strings.Contains(content, "## 追加コンテキスト") {
		t.Error("生成されたプロンプトに追加コンテキストセクションが含まれていません")
	}
	if !strings.Contains(content, additionalContext) {
		t.Error("生成されたプロンプトに追加コンテキストが含まれていません")
	}

	// 会話履歴セクションが含まれているかチェック
	if !strings.Contains(content, "## 会話履歴") {
		t.Error("生成されたプロンプトに会話履歴セクションが含まれていません")
	}

	// 履歴メッセージが含まれているかチェック
	if !strings.Contains(content, "ユーザーuser1: こんにちは") {
		t.Error("生成されたプロンプトに最初の履歴メッセージが含まれていません")
	}
	if !strings.Contains(content, "ユーザーuser2: こんばんは") {
		t.Error("生成されたプロンプトに2番目の履歴メッセージが含まれていません")
	}

	// ユーザーのチャット内容が含まれているかチェック
	if !strings.Contains(content, "## ユーザーのチャット内容") {
		t.Error("生成されたプロンプトにユーザーのチャット内容セクションが含まれていません")
	}
	if !strings.Contains(content, userQuestion) {
		t.Error("生成されたプロンプトにユーザーのチャット内容が含まれていません")
	}
}

func TestPromptGenerator_GeneratePromptWithContext_EmptyContext(t *testing.T) {
	generator := NewPromptGenerator("テストシステムプロンプト")
	history := NewConversationHistory([]Message{})
	userQuestion := "こんにちは"
	additionalContext := ""

	prompt := generator.GeneratePromptWithContext(history, userQuestion, additionalContext)
	content := prompt.Content()

	// 追加コンテキストセクションが含まれていないかチェック（空なので）
	if strings.Contains(content, "## 追加コンテキスト") {
		t.Error("空のコンテキストなのに追加コンテキストセクションが含まれています")
	}

	// ユーザーのチャット内容が含まれているかチェック
	if !strings.Contains(content, "## ユーザーのチャット内容") {
		t.Error("生成されたプロンプトにユーザーのチャット内容セクションが含まれていません")
	}
	if !strings.Contains(content, userQuestion) {
		t.Error("生成されたプロンプトにユーザーのチャット内容が含まれていません")
	}
}

func TestPromptGenerator_GeneratePrompt_Order(t *testing.T) {
	generator := NewPromptGenerator("システムプロンプト")
	history := NewConversationHistory([]Message{})
	userQuestion := "質問"

	prompt := generator.GeneratePrompt(history, userQuestion)
	content := prompt.Content()

	// 順序をチェック: システムプロンプト -> 会話履歴 -> ユーザーのチャット内容
	parts := strings.Split(content, "\n\n")

	if len(parts) < 2 {
		t.Error("プロンプトの構造が正しくありません")
	}

	// 最初の部分にシステムプロンプトが含まれているか
	if !strings.Contains(parts[0], "システムプロンプト") {
		t.Error("最初の部分にシステムプロンプトが含まれていません")
	}

	// 最後の部分にユーザーのチャット内容が含まれているか
	lastPart := parts[len(parts)-1]
	if !strings.Contains(lastPart, "## ユーザーのチャット内容") {
		t.Error("最後の部分にユーザーのチャット内容セクションが含まれていません")
	}
}
