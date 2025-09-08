package config

import "time"

// GeminiConfig は、Gemini API関連の設定を定義します
type GeminiConfig struct {
	APIKey      string
	ModelName   string
	MaxTokens   int32
	Temperature float32
	TopP        float32
	TopK        int32
	MaxRetries  int // 最大リトライ回数
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

// DefaultGeminiConfig は、デフォルトのGemini設定を返します
func DefaultGeminiConfig() *GeminiConfig {
	return &GeminiConfig{
		ModelName:   "gemini-2.5-pro",
		MaxTokens:   1000,
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
		MaxRetries:  3, // デフォルトで3回リトライ
	}
}

// DefaultBotConfig は、デフォルトのBot設定を返します
func DefaultBotConfig() *BotConfig {
	return &BotConfig{
		MaxContextLength: 8000,
		MaxHistoryLength: 4000,
		RequestTimeout:   30 * time.Second,
		SystemPrompt:     "あなたは親切で役立つAIアシスタントです。最も重要なのは、ユーザーが今送信した質問やリクエストに直接答えることです。会話履歴は参考情報として使用し、ユーザーの現在の質問を最優先で回答してください。",
	}
}
