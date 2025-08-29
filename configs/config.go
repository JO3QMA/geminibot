package configs

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"geminibot/internal/infrastructure/config"

	"github.com/joho/godotenv"
)

// Config は、アプリケーション全体の設定を定義します
type Config struct {
	Discord config.DiscordConfig
	Gemini  config.GeminiConfig
	Bot     config.BotConfig
}

// LoadConfig は、環境変数から設定を読み込みます
func LoadConfig() (*Config, error) {
	// .envファイルを読み込み（ファイルが存在しない場合は無視）
	if err := godotenv.Load(); err != nil {
		// .envファイルが存在しない場合は警告のみ出力（エラーにはしない）
		fmt.Printf("警告: .envファイルの読み込みに失敗しました: %v\n", err)
	}

	config := &Config{
		Discord: config.DiscordConfig{
			BotToken: getEnvOrDefault("DISCORD_BOT_TOKEN", ""),
		},
		Gemini: config.GeminiConfig{
			APIKey:      getEnvOrDefault("GEMINI_API_KEY", ""),
			ModelName:   getEnvOrDefault("GEMINI_MODEL_NAME", "gemini-pro"),
			MaxTokens:   int32(getEnvAsIntOrDefault("GEMINI_MAX_TOKENS", 1000)),
			Temperature: float32(getEnvAsFloatOrDefault("GEMINI_TEMPERATURE", 0.7)),
			TopP:        float32(getEnvAsFloatOrDefault("GEMINI_TOP_P", 0.9)),
			TopK:        int32(getEnvAsIntOrDefault("GEMINI_TOP_K", 40)),
		},
		Bot: config.BotConfig{
			MaxContextLength:     getEnvAsIntOrDefault("MAX_CONTEXT_LENGTH", 8000),
			MaxHistoryLength:     getEnvAsIntOrDefault("MAX_HISTORY_LENGTH", 4000),
			RequestTimeout:       getEnvAsDurationOrDefault("REQUEST_TIMEOUT", 30*time.Second),
			SystemPrompt:         getEnvOrDefault("SYSTEM_PROMPT", "あなたは親切で役立つAIアシスタントです。ユーザーのチャット内容に対して、安全で適切な回答を提供してください。有害な内容や不適切な内容については、適切に断るか、代替案を提案してください。"),
			UseStructuredContext: getEnvAsBoolOrDefault("USE_STRUCTURED_CONTEXT", true),
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

	if c.Bot.MaxContextLength <= 0 {
		return fmt.Errorf("MAX_CONTEXT_LENGTH は正の整数である必要があります")
	}

	if c.Bot.MaxHistoryLength <= 0 {
		return fmt.Errorf("MAX_HISTORY_LENGTH は正の整数である必要があります")
	}

	if c.Bot.MaxHistoryLength > c.Bot.MaxContextLength {
		return fmt.Errorf("MAX_HISTORY_LENGTH は MAX_CONTEXT_LENGTH 以下である必要があります")
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

// getEnvAsBoolOrDefault は、環境変数を真偽値として取得し、存在しない場合はデフォルト値を返します
func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
