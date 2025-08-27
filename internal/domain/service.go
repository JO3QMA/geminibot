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
		systemPrompt = "あなたは優秀なアシスタントです。与えられた会話履歴を参考に、ユーザーのチャット内容に適切に回答してください。"
	}

	return &PromptGenerator{
		systemPrompt: systemPrompt,
	}
}

// GeneratePrompt は、会話履歴とユーザーのチャット内容からプロンプトを生成します
func (pg *PromptGenerator) GeneratePrompt(history ConversationHistory, userQuestion string) Prompt {
	return pg.GeneratePromptWithMention(history, userQuestion, "", "")
}

// GeneratePromptWithMention は、メンション情報を含めてプロンプトを生成します
func (pg *PromptGenerator) GeneratePromptWithMention(history ConversationHistory, userQuestion string, mentionerName string, mentionerID string) Prompt {
	var builder strings.Builder

	// システムプロンプトを追加
	builder.WriteString(pg.systemPrompt)
	builder.WriteString("\n\n")

	// メンション情報がある場合は追加
	if mentionerName != "" {
		builder.WriteString("## メンション情報\n")
		builder.WriteString(fmt.Sprintf("メンションしたユーザー: %s (ID: %s)\n", mentionerName, mentionerID))
		builder.WriteString("\n")
	}

	// 会話履歴を追加
	if !history.IsEmpty() {
		builder.WriteString("## 会話履歴\n")
		for _, msg := range history.Messages() {
			displayName := msg.User.GetDisplayName()
			builder.WriteString(fmt.Sprintf("%s: %s\n", displayName, msg.Content))
		}
		builder.WriteString("\n")
	}

	// ユーザーのチャット内容を追加
	builder.WriteString("## ユーザーのチャット内容\n")
	builder.WriteString(userQuestion)

	return NewPrompt(builder.String())
}

// GeneratePromptWithContext は、追加のコンテキスト情報を含めてプロンプトを生成します
func (pg *PromptGenerator) GeneratePromptWithContext(history ConversationHistory, userQuestion string, additionalContext string) Prompt {
	return pg.GeneratePromptWithContextAndMention(history, userQuestion, additionalContext, "", "")
}

// GeneratePromptWithContextAndMention は、追加のコンテキスト情報とメンション情報を含めてプロンプトを生成します
func (pg *PromptGenerator) GeneratePromptWithContextAndMention(history ConversationHistory, userQuestion string, additionalContext string, mentionerName string, mentionerID string) Prompt {
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

	// メンション情報がある場合は追加
	if mentionerName != "" {
		builder.WriteString("## メンション情報\n")
		builder.WriteString(fmt.Sprintf("メンションしたユーザー: %s (ID: %s)\n", mentionerName, mentionerID))
		builder.WriteString("\n")
	}

	// 会話履歴を追加
	if !history.IsEmpty() {
		builder.WriteString("## 会話履歴\n")
		for _, msg := range history.Messages() {
			displayName := msg.User.GetDisplayName()
			builder.WriteString(fmt.Sprintf("%s: %s\n", displayName, msg.Content))
		}
		builder.WriteString("\n")
	}

	// ユーザーのチャット内容を追加
	builder.WriteString("## ユーザーのチャット内容\n")
	builder.WriteString(userQuestion)

	return NewPrompt(builder.String())
}
