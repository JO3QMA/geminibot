package domain

import (
	"sort"
	"strings"
	"unicode/utf8"
)

// ContextManager は、コンテキストの長さを管理するドメインサービスです
type ContextManager struct {
	maxContextLength int // 最大コンテキスト長（文字数）
	maxHistoryLength int // 最大履歴長（文字数）
}

// NewContextManager は新しいContextManagerインスタンスを作成します
func NewContextManager(maxContextLength, maxHistoryLength int) *ContextManager {
	return &ContextManager{
		maxContextLength: maxContextLength,
		maxHistoryLength: maxHistoryLength,
	}
}

// TruncateConversationHistory は、会話履歴を指定された長さに制限します
func (cm *ContextManager) TruncateConversationHistory(history ConversationHistory) ConversationHistory {
	if history.IsEmpty() {
		return history
	}

	messages := history.Messages()
	if len(messages) == 0 {
		return history
	}

	// 現在の履歴の総文字数を計算
	totalLength := cm.calculateHistoryLength(messages)

	// 制限内に収まっている場合はそのまま返す
	if totalLength <= cm.maxHistoryLength {
		return history
	}

	// 制限を超えている場合、新しいメッセージから優先的に保持
	truncatedMessages := cm.truncateMessagesFromNewest(messages)

	return NewConversationHistory(truncatedMessages)
}

// TruncateSystemPrompt は、システムプロンプトを指定された長さに制限します
func (cm *ContextManager) TruncateSystemPrompt(systemPrompt string) string {
	if utf8.RuneCountInString(systemPrompt) <= cm.maxContextLength {
		return systemPrompt
	}

	// 文字数制限を超えている場合、末尾を切り詰める
	runes := []rune(systemPrompt)
	if len(runes) > cm.maxContextLength {
		runes = runes[:cm.maxContextLength]
		// 完全な文で終わるように調整
		lastPeriod := strings.LastIndex(string(runes), "。")
		if lastPeriod > 0 && lastPeriod < len(runes)-50 {
			runes = runes[:lastPeriod+1]
		}
	}

	return string(runes)
}

// TruncateUserQuestion は、ユーザーの質問を指定された長さに制限します
func (cm *ContextManager) TruncateUserQuestion(userQuestion string) string {
	if utf8.RuneCountInString(userQuestion) <= cm.maxContextLength {
		return userQuestion
	}

	// 文字数制限を超えている場合、末尾を切り詰める
	runes := []rune(userQuestion)
	if len(runes) > cm.maxContextLength {
		runes = runes[:cm.maxContextLength]
		// 完全な文で終わるように調整
		lastPeriod := strings.LastIndex(string(runes), "。")
		if lastPeriod > 0 && lastPeriod < len(runes)-30 {
			runes = runes[:lastPeriod+1]
		}
	}

	return string(runes)
}

// calculateHistoryLength は、会話履歴の総文字数を計算します
func (cm *ContextManager) calculateHistoryLength(messages []Message) int {
	totalLength := 0
	for _, msg := range messages {
		// ユーザー名 + ": " + メッセージ内容 + 改行
		displayName := msg.User.DisplayName
		totalLength += utf8.RuneCountInString(displayName) + 2 + utf8.RuneCountInString(msg.Content) + 1
	}
	return totalLength
}

// truncateMessagesFromNewest は、新しいメッセージから優先的に保持して履歴を切り詰めます
func (cm *ContextManager) truncateMessagesFromNewest(messages []Message) []Message {
	// メッセージを時系列順にソート（新しい順）
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp.After(messages[j].Timestamp)
	})

	var truncatedMessages []Message
	currentLength := 0

	// 新しいメッセージから順に追加
	for _, msg := range messages {
		messageLength := utf8.RuneCountInString(msg.User.DisplayName) + 2 + utf8.RuneCountInString(msg.Content) + 1

		// このメッセージを追加しても制限内に収まる場合
		if currentLength+messageLength <= cm.maxHistoryLength {
			truncatedMessages = append(truncatedMessages, msg)
			currentLength += messageLength
		} else {
			// 制限を超える場合は終了
			break
		}
	}

	// 時系列順に戻す（古い順）
	sort.Slice(truncatedMessages, func(i, j int) bool {
		return truncatedMessages[i].Timestamp.Before(truncatedMessages[j].Timestamp)
	})

	return truncatedMessages
}

// GetContextStats は、コンテキストの統計情報を返します
func (cm *ContextManager) GetContextStats(systemPrompt string, history ConversationHistory, userQuestion string) ContextStats {
	systemLength := utf8.RuneCountInString(systemPrompt)
	historyLength := cm.calculateHistoryLength(history.Messages())
	questionLength := utf8.RuneCountInString(userQuestion)
	totalLength := systemLength + historyLength + questionLength

	return ContextStats{
		SystemPromptLength: systemLength,
		HistoryLength:      historyLength,
		QuestionLength:     questionLength,
		TotalLength:        totalLength,
		MaxContextLength:   cm.maxContextLength,
		MaxHistoryLength:   cm.maxHistoryLength,
		IsTruncated:        totalLength > cm.maxContextLength || historyLength > cm.maxHistoryLength,
	}
}

// ContextStats は、コンテキストの統計情報を表現します
type ContextStats struct {
	SystemPromptLength int
	HistoryLength      int
	QuestionLength     int
	TotalLength        int
	MaxContextLength   int
	MaxHistoryLength   int
	IsTruncated        bool
}
