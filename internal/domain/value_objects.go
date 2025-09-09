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

// ImageGenerationRequest は、画像生成リクエストを表現する値オブジェクトです
type ImageGenerationRequest struct {
	Prompt  string
	Model   string
	Style   string
	Quality string
	Size    string
	Count   int
}

// ImageGenerationResponse は、画像生成レスポンスを表現する値オブジェクトです
type ImageGenerationResponse struct {
	Images      []GeneratedImage
	Prompt      string
	Model       string
	GeneratedAt time.Time
}

// GeneratedImage は、生成された画像の情報を表現する値オブジェクトです
type GeneratedImage struct {
	Data        []byte
	MimeType    string
	Filename    string
	Size        int64
	GeneratedAt time.Time
}

// ImageGenerationOptions は、画像生成時のオプションを定義します
type ImageGenerationOptions struct {
	Model       string  `json:"model,omitempty"`
	Style       string  `json:"style,omitempty"`
	Quality     string  `json:"quality,omitempty"`
	Size        string  `json:"size,omitempty"`
	Count       int     `json:"count,omitempty"`
	MaxTokens   int32   `json:"max_tokens,omitempty"`
	Temperature float32 `json:"temperature,omitempty"`
	TopP        float32 `json:"top_p,omitempty"`
	TopK        int32   `json:"top_k,omitempty"`
}

// ImageGenerationResult は、画像生成の結果を表現する値オブジェクトです
type ImageGenerationResult struct {
	Images      []GeneratedImage
	Prompt      string
	Model       string
	GeneratedAt time.Time
	Success     bool
	Error       string
	ImageURL    string
}

// NewImagePrompt は、画像生成用のプロンプトを作成します
func NewImagePrompt(content string) string {
	return content
}

// DefaultImageGenerationOptions は、デフォルトの画像生成オプションを返します
func DefaultImageGenerationOptions() ImageGenerationOptions {
	return ImageGenerationOptions{
		Model:       "gemini-2.5-flash-image",
		Style:       "photographic",
		Quality:     "standard",
		Size:        "1024x1024",
		Count:       1,
		MaxTokens:   1000,
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
	}
}
