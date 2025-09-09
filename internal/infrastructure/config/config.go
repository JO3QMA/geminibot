package config

import "time"

// GeminiConfig は、Gemini API関連の設定を定義します
type GeminiConfig struct {
	APIKey           string
	ModelName        string
	ImageModelName   string  // 画像生成用モデル名
	MaxTokens        int32
	Temperature      float32
	TopP             float32
	TopK             int32
	MaxRetries       int     // 最大リトライ回数
	EnableImageGen   bool    // 画像生成機能の有効/無効
}

// BotConfig は、Bot関連の設定を定義します
type BotConfig struct {
	MaxContextLength int // 最大コンテキスト長（文字数）
	MaxHistoryLength int // 最大履歴長（文字数）
	RequestTimeout   time.Duration
	SystemPrompt     string
}

// DiscordConfig は、Discord関連の設定を定義します
type DiscordConfig struct {
	BotToken string
}

// AppConfig は、アプリケーション全体の設定を定義します
type AppConfig struct {
	Discord DiscordConfig
	Gemini  GeminiConfig
	Bot     BotConfig
}
