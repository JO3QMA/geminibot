package configs

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config は、アプリケーション全体の設定を定義します
type Config struct {
	Discord DiscordConfig
	Gemini  GeminiConfig
	Bot     BotConfig
}

// DiscordConfig は、Discord関連の設定を定義します
type DiscordConfig struct {
	BotToken string
}

// GeminiConfig は、Gemini API関連の設定を定義します
type GeminiConfig struct {
	APIKey      string
	ModelName   string
	MaxTokens   int32
	Temperature float32
	TopP        float32
	TopK        int32
}

// BotConfig は、Bot関連の設定を定義します
type BotConfig struct {
	MaxHistoryMessages int
	RequestTimeout     time.Duration
	SystemPrompt       string
}

// LoadConfig は、環境変数から設定を読み込みます
func LoadConfig() (*Config, error) {
	config := &Config{
		Discord: DiscordConfig{
			BotToken: getEnvOrDefault("DISCORD_BOT_TOKEN", ""),
		},
		Gemini: GeminiConfig{
			APIKey:      getEnvOrDefault("GEMINI_API_KEY", ""),
			ModelName:   getEnvOrDefault("GEMINI_MODEL_NAME", "gemini-pro"),
			MaxTokens:   int32(getEnvAsIntOrDefault("GEMINI_MAX_TOKENS", 1000)),
			Temperature: float32(getEnvAsFloatOrDefault("GEMINI_TEMPERATURE", 0.7)),
			TopP:        float32(getEnvAsFloatOrDefault("GEMINI_TOP_P", 0.9)),
			TopK:        int32(getEnvAsIntOrDefault("GEMINI_TOP_K", 40)),
		},
		Bot: BotConfig{
			MaxHistoryMessages: getEnvAsIntOrDefault("MAX_HISTORY_MESSAGES", 10),
			RequestTimeout:     getEnvAsDurationOrDefault("REQUEST_TIMEOUT", 30*time.Second),
			SystemPrompt:       getEnvOrDefault("SYSTEM_PROMPT", "あなたは優秀なアシスタントです。与えられた会話履歴を参考に、ユーザーの質問に適切に回答してください。"),
		},
	}

	// 必須設定の検証
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate は、設定の妥当性を検証します
func (c *Config) Validate() error {
	if c.Discord.BotToken == "" {
		return fmt.Errorf("DISCORD_BOT_TOKEN が設定されていません")
	}

	if c.Gemini.APIKey == "" {
		return fmt.Errorf("GEMINI_API_KEY が設定されていません")
	}

	if c.Bot.MaxHistoryMessages <= 0 {
		return fmt.Errorf("MAX_HISTORY_MESSAGES は正の整数である必要があります")
	}

	if c.Bot.RequestTimeout <= 0 {
		return fmt.Errorf("REQUEST_TIMEOUT は正の値である必要があります")
	}

	return nil
}

// getEnvOrDefault は、環境変数を取得し、存在しない場合はデフォルト値を返します
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsIntOrDefault は、環境変数を整数として取得し、存在しない場合はデフォルト値を返します
func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsFloatOrDefault は、環境変数を浮動小数点数として取得し、存在しない場合はデフォルト値を返します
func getEnvAsFloatOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// getEnvAsDurationOrDefault は、環境変数を時間として取得し、存在しない場合はデフォルト値を返します
func getEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
