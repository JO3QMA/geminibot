package domain

import (
	"fmt"
	"time"
)

// Message は、Discordのメッセージを表現する値オブジェクトです
type Message struct {
	ID        string
	User      User
	Content   string
	Timestamp time.Time
}

// User は、Discordのユーザー情報を表現する値オブジェクトです
type User struct {
	ID            string
	Username      string
	DisplayName   string
	Avatar        string
	IsBot         bool
	Discriminator string
}

// Prompt は、Gemini APIに送信するために整形されたテキストを表現する値オブジェクトです
type Prompt struct {
	Content string
}

// BotMention は、Botへのメンション情報を表現する値オブジェクトです
type BotMention struct {
	ChannelID string
	GuildID   string
	User      User
	Content   string
	MessageID string
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
	return fmt.Sprintf("BotMention{ChannelID: %s, GuildID: %s, User: %s, Content: %s, MessageID: %s}",
		bm.ChannelID, bm.GuildID, bm.User.Username, bm.Content, bm.MessageID)
}
