package domain

import "context"

// ConversationRepository は、指定された条件に基づき、DiscordからConversationHistoryを取得するためのインターフェースです
type ConversationRepository interface {
	// GetRecentMessages は、指定されたチャンネルの直近のメッセージを取得します
	GetRecentMessages(ctx context.Context, channelID string, limit int) (ConversationHistory, error)

	// GetThreadMessages は、指定されたスレッドの全メッセージを取得します
	GetThreadMessages(ctx context.Context, threadID string) (ConversationHistory, error)

	// GetMessagesBefore は、指定されたメッセージIDより前のメッセージを取得します
	GetMessagesBefore(ctx context.Context, channelID string, messageID string, limit int) (ConversationHistory, error)
}
