package configs

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"geminibot/internal/infrastructure/config"

	"github.com/joho/godotenv"
)

// Config は config.AppConfig のエイリアスです（設定構造体の二重定義を避けます）。
type Config = config.AppConfig

// LoadConfig は、環境変数から設定を読み込みます
func LoadConfig() (*config.AppConfig, error) {
	// .envファイルを読み込み（ファイルが存在しない場合は無視）
	if err := godotenv.Load(); err != nil {
		// .envファイルが存在しない場合は警告のみ出力（エラーにはしない）
		fmt.Printf("警告: .envファイルの読み込みに失敗しました: %v\n", err)
	}

	cfg := &config.AppConfig{
		Discord: config.DiscordConfig{
			BotToken: getEnvOrDefault("DISCORD_BOT_TOKEN", ""),
		},
		Gemini: config.GeminiConfig{
			APIKey:         getEnvOrDefault("GEMINI_API_KEY", ""),
			ModelName:      getEnvOrDefault("GEMINI_MODEL_NAME", config.DefaultGeminiTextModel),
			MaxTokens:      int32(getEnvAsIntOrDefault("GEMINI_MAX_TOKENS", 1000)),
			Temperature:    float32(getEnvAsFloatOrDefault("GEMINI_TEMPERATURE", 0.7)),
			TopP:           float32(getEnvAsFloatOrDefault("GEMINI_TOP_P", 0.9)),
			TopK:           int32(getEnvAsIntOrDefault("GEMINI_TOP_K", 40)),
			MaxRetries:     getEnvAsIntOrDefault("GEMINI_MAX_RETRIES", 3),
			EnableImageGen: getEnvAsBoolOrDefault("GEMINI_ENABLE_IMAGE_GEN", true),

			// 画像生成関連の設定
			ImageModelName: getEnvOrDefault("GEMINI_IMAGE_MODEL_NAME", "gemini-2.5-flash-image-preview"),
			ImageStyle:     getEnvOrDefault("GEMINI_IMAGE_STYLE", "photographic"),
			ImageQuality:   getEnvOrDefault("GEMINI_IMAGE_QUALITY", "standard"),
			ImageSize:      getEnvOrDefault("GEMINI_IMAGE_SIZE", "1024x1024"),
			ImageCount:     getEnvAsIntOrDefault("GEMINI_IMAGE_COUNT", 1),
		},
		Bot: config.BotConfig{
			MaxContextLength: getEnvAsIntOrDefault("MAX_CONTEXT_LENGTH", 8000),
			MaxHistoryLength: getEnvAsIntOrDefault("MAX_HISTORY_LENGTH", 4000),
			RequestTimeout:   getEnvAsDurationOrDefault("REQUEST_TIMEOUT", 30*time.Second),
			SystemPrompt:     getEnvOrDefault("SYSTEM_PROMPT", "あなたは親切で役立つAIアシスタントです。最も重要なのは、ユーザーが今送信した質問やリクエストに直接答えることです。会話履歴は参考情報として使用し、ユーザーの現在の質問を最優先で回答してください。有害な内容や不適切な内容については、適切に断るか、代替案を提案してください。"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
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
