package domain

import (
	"fmt"
	"strings"
)

// PromptGenerator は、ConversationHistoryとユーザーからの質問から、Geminiに最適なPromptを生成するビジネスロジックを担当します
type PromptGenerator struct {
	systemPrompt string
}

// NewPromptGenerator は新しいPromptGeneratorインスタンスを作成します
func NewPromptGenerator(systemPrompt string) *PromptGenerator {
	if systemPrompt == "" {
		systemPrompt = "あなたは優秀なアシスタントです。与えられた会話履歴を参考に、ユーザーの質問に適切に回答してください。"
	}

	return &PromptGenerator{
		systemPrompt: systemPrompt,
	}
}

// GeneratePrompt は、会話履歴とユーザーの質問からプロンプトを生成します
func (pg *PromptGenerator) GeneratePrompt(history ConversationHistory, userQuestion string) Prompt {
	var builder strings.Builder

	// システムプロンプトを追加
	builder.WriteString(pg.systemPrompt)
	builder.WriteString("\n\n")

	// 会話履歴を追加
	if !history.IsEmpty() {
		builder.WriteString("## 会話履歴\n")
		for _, msg := range history.Messages() {
			builder.WriteString(fmt.Sprintf("ユーザー%s: %s\n", msg.UserID, msg.Content))
		}
		builder.WriteString("\n")
	}

	// ユーザーの質問を追加
	builder.WriteString("## ユーザーの質問\n")
	builder.WriteString(userQuestion)

	return NewPrompt(builder.String())
}

// GeneratePromptWithContext は、追加のコンテキスト情報を含めてプロンプトを生成します
func (pg *PromptGenerator) GeneratePromptWithContext(history ConversationHistory, userQuestion string, additionalContext string) Prompt {
	var builder strings.Builder

	// システムプロンプトを追加
	builder.WriteString(pg.systemPrompt)
	builder.WriteString("\n\n")

	// 追加コンテキストがある場合は追加
	if additionalContext != "" {
		builder.WriteString("## 追加コンテキスト\n")
		builder.WriteString(additionalContext)
		builder.WriteString("\n\n")
	}

	// 会話履歴を追加
	if !history.IsEmpty() {
		builder.WriteString("## 会話履歴\n")
		for _, msg := range history.Messages() {
			builder.WriteString(fmt.Sprintf("ユーザー%s: %s\n", msg.UserID, msg.Content))
		}
		builder.WriteString("\n")
	}

	// ユーザーの質問を追加
	builder.WriteString("## ユーザーの質問\n")
	builder.WriteString(userQuestion)

	return NewPrompt(builder.String())
}
