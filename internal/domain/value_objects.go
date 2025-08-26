package domain

import (
	"fmt"
	"time"
)

// Message は、Discordのメッセージを表現する値オブジェクトです
type Message struct {
	ID        string
	UserID    UserID
	Content   string
	Timestamp time.Time
}

// NewMessage は新しいMessageインスタンスを作成します
func NewMessage(id string, userID UserID, content string, timestamp time.Time) Message {
	return Message{
		ID:        id,
		UserID:    userID,
		Content:   content,
		Timestamp: timestamp,
	}
}

// UserID は、DiscordのユーザーIDを表現する値オブジェクトです
type UserID string

// NewUserID は新しいUserIDインスタンスを作成します
func NewUserID(id string) UserID {
	return UserID(id)
}

// String はUserIDを文字列として返します
func (u UserID) String() string {
	return string(u)
}

// ChannelID は、DiscordのチャンネルIDを表現する値オブジェクトです
type ChannelID string

// NewChannelID は新しいChannelIDインスタンスを作成します
func NewChannelID(id string) ChannelID {
	return ChannelID(id)
}

// String はChannelIDを文字列として返します
func (c ChannelID) String() string {
	return string(c)
}

// ConversationHistory は、複数のMessageを内包する、コンテキストを表すオブジェクトです
type ConversationHistory struct {
	messages []Message
}

// NewConversationHistory は新しいConversationHistoryインスタンスを作成します
func NewConversationHistory(messages []Message) ConversationHistory {
	return ConversationHistory{
		messages: messages,
	}
}

// Messages は履歴メッセージのスライスを返します
func (ch ConversationHistory) Messages() []Message {
	return ch.messages
}

// Count は履歴メッセージの数を返します
func (ch ConversationHistory) Count() int {
	return len(ch.messages)
}

// IsEmpty は履歴が空かどうかを判定します
func (ch ConversationHistory) IsEmpty() bool {
	return len(ch.messages) == 0
}

// Prompt は、Gemini APIに送信するために整形されたテキストを表現する値オブジェクトです
type Prompt struct {
	content string
}

// NewPrompt は新しいPromptインスタンスを作成します
func NewPrompt(content string) Prompt {
	return Prompt{
		content: content,
	}
}

// Content はプロンプトの内容を返します
func (p Prompt) Content() string {
	return p.content
}

// String はPromptを文字列として返します
func (p Prompt) String() string {
	return p.content
}

// BotMention は、Botへのメンション情報を表現する値オブジェクトです
type BotMention struct {
	ChannelID ChannelID
	UserID    UserID
	Content   string
	MessageID string
}

// NewBotMention は新しいBotMentionインスタンスを作成します
func NewBotMention(channelID ChannelID, userID UserID, content, messageID string) BotMention {
	return BotMention{
		ChannelID: channelID,
		UserID:    userID,
		Content:   content,
		MessageID: messageID,
	}
}

// IsThread は、このメンションがスレッド内で発生したかどうかを判定します
// この判定は、チャンネルIDの形式に基づいて行われます
func (bm BotMention) IsThread() bool {
	// DiscordのスレッドチャンネルIDは通常のチャンネルIDと異なる形式を持つ場合があります
	// 実際の実装では、Discord APIの仕様に基づいて判定ロジックを調整する必要があります
	return false // 仮の実装
}

// String はBotMentionの文字列表現を返します
func (bm BotMention) String() string {
	return fmt.Sprintf("BotMention{ChannelID: %s, UserID: %s, Content: %s, MessageID: %s}",
		bm.ChannelID, bm.UserID, bm.Content, bm.MessageID)
}
