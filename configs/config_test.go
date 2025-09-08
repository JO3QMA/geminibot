package configs

import (
	"os"
	"testing"
	"time"

	"geminibot/internal/infrastructure/config"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "有効な設定",
			config: &Config{
				Discord: config.DiscordConfig{
					BotToken: "test-token",
				},
				Gemini: config.GeminiConfig{
					APIKey:      "test-api-key",
					ModelName:   "gemini-2.5-pro",
					MaxTokens:   1000,
					Temperature: 0.7,
					TopP:        0.9,
					TopK:        40,
					MaxRetries:  3,
				},
				Bot: config.BotConfig{
					MaxContextLength: 8000,
					MaxHistoryLength: 4000,
					RequestTimeout:   30 * time.Second,
					SystemPrompt:     "test prompt",
				},
			},
			wantErr: false,
		},
		{
			name: "Discord BotTokenが空",
			config: &Config{
				Discord: config.DiscordConfig{
					BotToken: "",
				},
				Gemini: config.GeminiConfig{
					APIKey:      "test-api-key",
					ModelName:   "gemini-2.5-pro",
					MaxTokens:   1000,
					Temperature: 0.7,
					TopP:        0.9,
					TopK:        40,
					MaxRetries:  3,
				},
				Bot: config.BotConfig{
					MaxContextLength: 8000,
					MaxHistoryLength: 4000,
					RequestTimeout:   30 * time.Second,
					SystemPrompt:     "test prompt",
				},
			},
			wantErr: true,
			errMsg:  "DISCORD_BOT_TOKEN が設定されていません",
		},
		{
			name: "Gemini APIKeyが空",
			config: &Config{
				Discord: config.DiscordConfig{
					BotToken: "test-token",
				},
				Gemini: config.GeminiConfig{
					APIKey:      "",
					ModelName:   "gemini-2.5-pro",
					MaxTokens:   1000,
					Temperature: 0.7,
					TopP:        0.9,
					TopK:        40,
					MaxRetries:  3,
				},
				Bot: config.BotConfig{
					MaxContextLength: 8000,
					MaxHistoryLength: 4000,
					RequestTimeout:   30 * time.Second,
					SystemPrompt:     "test prompt",
				},
			},
			wantErr: true,
			errMsg:  "GEMINI_API_KEY が設定されていません",
		},
		{
			name: "MaxTokensが0以下",
			config: &Config{
				Discord: config.DiscordConfig{
					BotToken: "test-token",
				},
				Gemini: config.GeminiConfig{
					APIKey:      "test-api-key",
					ModelName:   "gemini-2.5-pro",
					MaxTokens:   0,
					Temperature: 0.7,
					TopP:        0.9,
					TopK:        40,
					MaxRetries:  3,
				},
				Bot: config.BotConfig{
					MaxContextLength: 8000,
					MaxHistoryLength: 4000,
					RequestTimeout:   30 * time.Second,
					SystemPrompt:     "test prompt",
				},
			},
			wantErr: true,
			errMsg:  "GEMINI_MAX_TOKENS は正の整数である必要があります",
		},
		{
			name: "Temperatureが範囲外（負の値）",
			config: &Config{
				Discord: config.DiscordConfig{
					BotToken: "test-token",
				},
				Gemini: config.GeminiConfig{
					APIKey:      "test-api-key",
					ModelName:   "gemini-2.5-pro",
					MaxTokens:   1000,
					Temperature: -0.1,
					TopP:        0.9,
					TopK:        40,
					MaxRetries:  3,
				},
				Bot: config.BotConfig{
					MaxContextLength: 8000,
					MaxHistoryLength: 4000,
					RequestTimeout:   30 * time.Second,
					SystemPrompt:     "test prompt",
				},
			},
			wantErr: true,
			errMsg:  "GEMINI_TEMPERATURE は0以上2以下の値である必要があります",
		},
		{
			name: "Temperatureが範囲外（2を超える）",
			config: &Config{
				Discord: config.DiscordConfig{
					BotToken: "test-token",
				},
				Gemini: config.GeminiConfig{
					APIKey:      "test-api-key",
					ModelName:   "gemini-2.5-pro",
					MaxTokens:   1000,
					Temperature: 2.1,
					TopP:        0.9,
					TopK:        40,
					MaxRetries:  3,
				},
				Bot: config.BotConfig{
					MaxContextLength: 8000,
					MaxHistoryLength: 4000,
					RequestTimeout:   30 * time.Second,
					SystemPrompt:     "test prompt",
				},
			},
			wantErr: true,
			errMsg:  "GEMINI_TEMPERATURE は0以上2以下の値である必要があります",
		},
		{
			name: "TopPが範囲外（負の値）",
			config: &Config{
				Discord: config.DiscordConfig{
					BotToken: "test-token",
				},
				Gemini: config.GeminiConfig{
					APIKey:      "test-api-key",
					ModelName:   "gemini-2.5-pro",
					MaxTokens:   1000,
					Temperature: 0.7,
					TopP:        -0.1,
					TopK:        40,
					MaxRetries:  3,
				},
				Bot: config.BotConfig{
					MaxContextLength: 8000,
					MaxHistoryLength: 4000,
					RequestTimeout:   30 * time.Second,
					SystemPrompt:     "test prompt",
				},
			},
			wantErr: true,
			errMsg:  "GEMINI_TOP_P は0以上1以下の値である必要があります",
		},
		{
			name: "TopPが範囲外（1を超える）",
			config: &Config{
				Discord: config.DiscordConfig{
					BotToken: "test-token",
				},
				Gemini: config.GeminiConfig{
					APIKey:      "test-api-key",
					ModelName:   "gemini-2.5-pro",
					MaxTokens:   1000,
					Temperature: 0.7,
					TopP:        1.1,
					TopK:        40,
					MaxRetries:  3,
				},
				Bot: config.BotConfig{
					MaxContextLength: 8000,
					MaxHistoryLength: 4000,
					RequestTimeout:   30 * time.Second,
					SystemPrompt:     "test prompt",
				},
			},
			wantErr: true,
			errMsg:  "GEMINI_TOP_P は0以上1以下の値である必要があります",
		},
		{
			name: "TopKが0以下",
			config: &Config{
				Discord: config.DiscordConfig{
					BotToken: "test-token",
				},
				Gemini: config.GeminiConfig{
					APIKey:      "test-api-key",
					ModelName:   "gemini-2.5-pro",
					MaxTokens:   1000,
					Temperature: 0.7,
					TopP:        0.9,
					TopK:        0,
					MaxRetries:  3,
				},
				Bot: config.BotConfig{
					MaxContextLength: 8000,
					MaxHistoryLength: 4000,
					RequestTimeout:   30 * time.Second,
					SystemPrompt:     "test prompt",
				},
			},
			wantErr: true,
			errMsg:  "GEMINI_TOP_K は正の整数である必要があります",
		},
		{
			name: "MaxRetriesが負の値",
			config: &Config{
				Discord: config.DiscordConfig{
					BotToken: "test-token",
				},
				Gemini: config.GeminiConfig{
					APIKey:      "test-api-key",
					ModelName:   "gemini-2.5-pro",
					MaxTokens:   1000,
					Temperature: 0.7,
					TopP:        0.9,
					TopK:        40,
					MaxRetries:  -1,
				},
				Bot: config.BotConfig{
					MaxContextLength: 8000,
					MaxHistoryLength: 4000,
					RequestTimeout:   30 * time.Second,
					SystemPrompt:     "test prompt",
				},
			},
			wantErr: true,
			errMsg:  "GEMINI_MAX_RETRIES は0以上の整数である必要があります",
		},
		{
			name: "MaxContextLengthが0以下",
			config: &Config{
				Discord: config.DiscordConfig{
					BotToken: "test-token",
				},
				Gemini: config.GeminiConfig{
					APIKey:      "test-api-key",
					ModelName:   "gemini-2.5-pro",
					MaxTokens:   1000,
					Temperature: 0.7,
					TopP:        0.9,
					TopK:        40,
					MaxRetries:  3,
				},
				Bot: config.BotConfig{
					MaxContextLength: 0,
					MaxHistoryLength: 4000,
					RequestTimeout:   30 * time.Second,
					SystemPrompt:     "test prompt",
				},
			},
			wantErr: true,
			errMsg:  "MAX_CONTEXT_LENGTH は正の整数である必要があります",
		},
		{
			name: "MaxHistoryLengthが0以下",
			config: &Config{
				Discord: config.DiscordConfig{
					BotToken: "test-token",
				},
				Gemini: config.GeminiConfig{
					APIKey:      "test-api-key",
					ModelName:   "gemini-2.5-pro",
					MaxTokens:   1000,
					Temperature: 0.7,
					TopP:        0.9,
					TopK:        40,
					MaxRetries:  3,
				},
				Bot: config.BotConfig{
					MaxContextLength: 8000,
					MaxHistoryLength: 0,
					RequestTimeout:   30 * time.Second,
					SystemPrompt:     "test prompt",
				},
			},
			wantErr: true,
			errMsg:  "MAX_HISTORY_LENGTH は正の整数である必要があります",
		},
		{
			name: "MaxHistoryLengthがMaxContextLengthを超える",
			config: &Config{
				Discord: config.DiscordConfig{
					BotToken: "test-token",
				},
				Gemini: config.GeminiConfig{
					APIKey:      "test-api-key",
					ModelName:   "gemini-2.5-pro",
					MaxTokens:   1000,
					Temperature: 0.7,
					TopP:        0.9,
					TopK:        40,
					MaxRetries:  3,
				},
				Bot: config.BotConfig{
					MaxContextLength: 4000,
					MaxHistoryLength: 8000,
					RequestTimeout:   30 * time.Second,
					SystemPrompt:     "test prompt",
				},
			},
			wantErr: true,
			errMsg:  "MAX_HISTORY_LENGTH は MAX_CONTEXT_LENGTH 以下である必要があります",
		},
		{
			name: "RequestTimeoutが0以下",
			config: &Config{
				Discord: config.DiscordConfig{
					BotToken: "test-token",
				},
				Gemini: config.GeminiConfig{
					APIKey:      "test-api-key",
					ModelName:   "gemini-2.5-pro",
					MaxTokens:   1000,
					Temperature: 0.7,
					TopP:        0.9,
					TopK:        40,
					MaxRetries:  3,
				},
				Bot: config.BotConfig{
					MaxContextLength: 8000,
					MaxHistoryLength: 4000,
					RequestTimeout:   0,
					SystemPrompt:     "test prompt",
				},
			},
			wantErr: true,
			errMsg:  "REQUEST_TIMEOUT は正の値である必要があります",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("エラーが期待されましたが、発生しませんでした")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("期待されるエラーメッセージ: %s, 実際: %s", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラーが発生しました: %v", err)
				}
			}
		})
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	// 環境変数をクリア
	os.Unsetenv("TEST_ENV_VAR")

	// デフォルト値のテスト
	result := getEnvOrDefault("TEST_ENV_VAR", "default")
	if result != "default" {
		t.Errorf("期待される値: default, 実際: %s", result)
	}

	// 環境変数が設定されている場合のテスト
	os.Setenv("TEST_ENV_VAR", "test-value")
	defer os.Unsetenv("TEST_ENV_VAR")

	result = getEnvOrDefault("TEST_ENV_VAR", "default")
	if result != "test-value" {
		t.Errorf("期待される値: test-value, 実際: %s", result)
	}
}

func TestGetEnvAsIntOrDefault(t *testing.T) {
	// 環境変数をクリア
	os.Unsetenv("TEST_INT_VAR")

	// デフォルト値のテスト
	result := getEnvAsIntOrDefault("TEST_INT_VAR", 42)
	if result != 42 {
		t.Errorf("期待される値: 42, 実際: %d", result)
	}

	// 有効な整数値のテスト
	os.Setenv("TEST_INT_VAR", "123")
	defer os.Unsetenv("TEST_INT_VAR")

	result = getEnvAsIntOrDefault("TEST_INT_VAR", 42)
	if result != 123 {
		t.Errorf("期待される値: 123, 実際: %d", result)
	}

	// 無効な値のテスト
	os.Setenv("TEST_INT_VAR", "invalid")
	result = getEnvAsIntOrDefault("TEST_INT_VAR", 42)
	if result != 42 {
		t.Errorf("無効な値の場合、デフォルト値が返されるべきです。期待: 42, 実際: %d", result)
	}
}

func TestGetEnvAsFloatOrDefault(t *testing.T) {
	// 環境変数をクリア
	os.Unsetenv("TEST_FLOAT_VAR")

	// デフォルト値のテスト
	result := getEnvAsFloatOrDefault("TEST_FLOAT_VAR", 3.14)
	if result != 3.14 {
		t.Errorf("期待される値: 3.14, 実際: %f", result)
	}

	// 有効な浮動小数点値のテスト
	os.Setenv("TEST_FLOAT_VAR", "2.71")
	defer os.Unsetenv("TEST_FLOAT_VAR")

	result = getEnvAsFloatOrDefault("TEST_FLOAT_VAR", 3.14)
	if result != 2.71 {
		t.Errorf("期待される値: 2.71, 実際: %f", result)
	}

	// 無効な値のテスト
	os.Setenv("TEST_FLOAT_VAR", "invalid")
	result = getEnvAsFloatOrDefault("TEST_FLOAT_VAR", 3.14)
	if result != 3.14 {
		t.Errorf("無効な値の場合、デフォルト値が返されるべきです。期待: 3.14, 実際: %f", result)
	}
}

func TestGetEnvAsDurationOrDefault(t *testing.T) {
	// 環境変数をクリア
	os.Unsetenv("TEST_DURATION_VAR")

	// デフォルト値のテスト
	defaultDuration := 30 * time.Second
	result := getEnvAsDurationOrDefault("TEST_DURATION_VAR", defaultDuration)
	if result != defaultDuration {
		t.Errorf("期待される値: %v, 実際: %v", defaultDuration, result)
	}

	// 有効な時間値のテスト
	os.Setenv("TEST_DURATION_VAR", "60s")
	defer os.Unsetenv("TEST_DURATION_VAR")

	expectedDuration := 60 * time.Second
	result = getEnvAsDurationOrDefault("TEST_DURATION_VAR", defaultDuration)
	if result != expectedDuration {
		t.Errorf("期待される値: %v, 実際: %v", expectedDuration, result)
	}

	// 無効な値のテスト
	os.Setenv("TEST_DURATION_VAR", "invalid")
	result = getEnvAsDurationOrDefault("TEST_DURATION_VAR", defaultDuration)
	if result != defaultDuration {
		t.Errorf("無効な値の場合、デフォルト値が返されるべきです。期待: %v, 実際: %v", defaultDuration, result)
	}
}
