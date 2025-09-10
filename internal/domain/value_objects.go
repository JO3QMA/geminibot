package domain

import (
	"fmt"
	"geminibot/configs"
	"log"
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
	Options ImageGenerationOptions
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
	Model       string       `json:"model,omitempty"`
	Style       ImageStyle   `json:"style,omitempty"`
	Quality     ImageQuality `json:"quality,omitempty"`
	Size        ImageSize    `json:"size,omitempty"`
	Count       int          `json:"count,omitempty"`
	MaxTokens   int32        `json:"max_tokens,omitempty"`
	Temperature float32      `json:"temperature,omitempty"`
	TopP        float32      `json:"top_p,omitempty"`
	TopK        int32        `json:"top_k,omitempty"`
}

// ImageGenerationResult は、画像生成の結果を表現する値オブジェクトです
type ImageGenerationResult struct {
	Response *ImageGenerationResponse // 画像生成レスポンスを内包
	Success  bool                     // 成功/失敗の状態
	Error    string                   // エラーメッセージ
	ImageURL string                   // 画像URL（必要に応じて設定）
}

// NewImagePrompt は、画像生成用のプロンプトを作成します
func NewImagePrompt(content string) string {
	return content
}

// DefaultImageGenerationOptions は、デフォルトの画像生成オプションを返します
func DefaultImageGenerationOptions() ImageGenerationOptions {
	geminiConfig, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return ImageGenerationOptions{
		Model:       geminiConfig.Gemini.ImageModelName,
		Style:       ImageStyleFromString(geminiConfig.Gemini.ImageStyle),
		Quality:     ImageQualityFromString(geminiConfig.Gemini.ImageQuality),
		Size:        ImageSizeFromString(geminiConfig.Gemini.ImageSize),
		Count:       geminiConfig.Gemini.ImageCount,
		MaxTokens:   geminiConfig.Gemini.MaxTokens,
		Temperature: geminiConfig.Gemini.Temperature,
		TopP:        geminiConfig.Gemini.TopP,
		TopK:        geminiConfig.Gemini.TopK,
	}
}

// Attachment は、添付ファイルの情報を表現する値オブジェクトです
type Attachment struct {
	Data        []byte    // ファイルデータ
	MimeType    string    // MIMEタイプ
	Filename    string    // ファイル名
	Size        int64     // ファイルサイズ
	IsImage     bool      // 画像かどうか
	GeneratedAt time.Time // 生成時刻
}

// ResponseMetadata は、レスポンスのメタデータを表現する値オブジェクトです
type ResponseMetadata struct {
	Prompt      string    // プロンプト
	Model       string    // 使用モデル
	GeneratedAt time.Time // 生成時刻
	Type        string    // レスポンスタイプ（text, image, mixed）
}

// UnifiedResponse は、テキスト生成と画像生成を統合したレスポンスを表現する値オブジェクトです
type UnifiedResponse struct {
	Content     string           // テキストコンテンツ
	Attachments []Attachment     // 添付ファイル（画像など）
	Metadata    ResponseMetadata // メタデータ（プロンプト、モデルなど）
	Success     bool             // 成功/失敗
	Error       string           // エラーメッセージ
	ThreadID    string           // スレッドID（空の場合はリプライで送信）
}

// NewTextResponse は、テキストレスポンスを作成します
func NewTextResponse(content, prompt, model string) *UnifiedResponse {
	return &UnifiedResponse{
		Content:     content,
		Attachments: []Attachment{},
		Metadata: ResponseMetadata{
			Prompt:      prompt,
			Model:       model,
			GeneratedAt: time.Now(),
			Type:        "text",
		},
		Success:  true,
		Error:    "",
		ThreadID: "",
	}
}

// NewImageResponse は、画像レスポンスを作成します
func NewImageResponse(content string, images []GeneratedImage, prompt, model string) *UnifiedResponse {
	attachments := make([]Attachment, len(images))
	for i, img := range images {
		attachments[i] = Attachment{
			Data:        img.Data,
			MimeType:    img.MimeType,
			Filename:    img.Filename,
			Size:        img.Size,
			IsImage:     true,
			GeneratedAt: img.GeneratedAt,
		}
	}

	return &UnifiedResponse{
		Content:     content,
		Attachments: attachments,
		Metadata: ResponseMetadata{
			Prompt:      prompt,
			Model:       model,
			GeneratedAt: time.Now(),
			Type:        "image",
		},
		Success:  true,
		Error:    "",
		ThreadID: "",
	}
}

// NewErrorResponse は、エラーレスポンスを作成します
func NewErrorResponse(err error, responseType string) *UnifiedResponse {
	return &UnifiedResponse{
		Content:     "",
		Attachments: []Attachment{},
		Metadata: ResponseMetadata{
			Prompt:      "",
			Model:       "",
			GeneratedAt: time.Now(),
			Type:        responseType,
		},
		Success:  false,
		Error:    err.Error(),
		ThreadID: "",
	}
}

// HasAttachments は、添付ファイルがあるかどうかを判定します
func (ur *UnifiedResponse) HasAttachments() bool {
	return len(ur.Attachments) > 0
}

// HasImages は、画像添付があるかどうかを判定します
func (ur *UnifiedResponse) HasImages() bool {
	for _, attachment := range ur.Attachments {
		if attachment.IsImage {
			return true
		}
	}
	return false
}

// NewTextResponseForThread は、スレッド用のテキストレスポンスを作成します
func NewTextResponseForThread(content, prompt, model, threadID string) *UnifiedResponse {
	response := NewTextResponse(content, prompt, model)
	response.ThreadID = threadID
	return response
}

// NewImageResponseForThread は、スレッド用の画像レスポンスを作成します
func NewImageResponseForThread(content string, images []GeneratedImage, prompt, model, threadID string) *UnifiedResponse {
	response := NewImageResponse(content, images, prompt, model)
	response.ThreadID = threadID
	return response
}

// NewErrorResponseForThread は、スレッド用のエラーレスポンスを作成します
func NewErrorResponseForThread(err error, responseType, threadID string) *UnifiedResponse {
	response := NewErrorResponse(err, responseType)
	response.ThreadID = threadID
	return response
}
