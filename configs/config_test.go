package configs

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig_WithValidEnvVars(t *testing.T) {
	// テスト用の環境変数を設定
	os.Setenv("DISCORD_BOT_TOKEN", "test-discord-token")
	os.Setenv("GEMINI_API_KEY", "test-gemini-key")
	os.Setenv("GEMINI_MODEL_NAME", "gemini-pro")
	os.Setenv("GEMINI_MAX_TOKENS", "500")
	os.Setenv("GEMINI_TEMPERATURE", "0.5")
	os.Setenv("GEMINI_TOP_P", "0.8")
	os.Setenv("GEMINI_TOP_K", "20")
	os.Setenv("MAX_HISTORY_MESSAGES", "5")
	os.Setenv("REQUEST_TIMEOUT", "15s")
	os.Setenv("SYSTEM_PROMPT", "テストシステムプロンプト")

	// テスト後に環境変数をクリーンアップ
	defer func() {
		os.Unsetenv("DISCORD_BOT_TOKEN")
		os.Unsetenv("GEMINI_API_KEY")
		os.Unsetenv("GEMINI_MODEL_NAME")
		os.Unsetenv("GEMINI_MAX_TOKENS")
		os.Unsetenv("GEMINI_TEMPERATURE")
		os.Unsetenv("GEMINI_TOP_P")
		os.Unsetenv("GEMINI_TOP_K")
		os.Unsetenv("MAX_HISTORY_MESSAGES")
		os.Unsetenv("REQUEST_TIMEOUT")
		os.Unsetenv("SYSTEM_PROMPT")
	}()

	config, err := LoadConfig()

	if err != nil {
		t.Errorf("設定の読み込みでエラーが発生しました: %v", err)
	}

	// Discord設定の検証
	if config.Discord.BotToken != "test-discord-token" {
		t.Errorf("期待されるDiscord BotToken: test-discord-token, 実際: %s", config.Discord.BotToken)
	}

	// Gemini設定の検証
	if config.Gemini.APIKey != "test-gemini-key" {
		t.Errorf("期待されるGemini APIKey: test-gemini-key, 実際: %s", config.Gemini.APIKey)
	}
	if config.Gemini.ModelName != "gemini-pro" {
		t.Errorf("期待されるGemini ModelName: gemini-pro, 実際: %s", config.Gemini.ModelName)
	}
	if config.Gemini.MaxTokens != 500 {
		t.Errorf("期待されるGemini MaxTokens: 500, 実際: %d", config.Gemini.MaxTokens)
	}
	if config.Gemini.Temperature != 0.5 {
		t.Errorf("期待されるGemini Temperature: 0.5, 実際: %f", config.Gemini.Temperature)
	}
	if config.Gemini.TopP != 0.8 {
		t.Errorf("期待されるGemini TopP: 0.8, 実際: %f", config.Gemini.TopP)
	}
	if config.Gemini.TopK != 20 {
		t.Errorf("期待されるGemini TopK: 20, 実際: %d", config.Gemini.TopK)
	}

	// Bot設定の検証
	if config.Bot.MaxHistoryMessages != 5 {
		t.Errorf("期待されるBot MaxHistoryMessages: 5, 実際: %d", config.Bot.MaxHistoryMessages)
	}
	if config.Bot.RequestTimeout != 15*time.Second {
		t.Errorf("期待されるBot RequestTimeout: 15s, 実際: %v", config.Bot.RequestTimeout)
	}
	if config.Bot.SystemPrompt != "テストシステムプロンプト" {
		t.Errorf("期待されるBot SystemPrompt: テストシステムプロンプト, 実際: %s", config.Bot.SystemPrompt)
	}
}

func TestLoadConfig_WithDefaultValues(t *testing.T) {
	// 環境変数をクリア
	os.Unsetenv("DISCORD_BOT_TOKEN")
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("GEMINI_MODEL_NAME")
	os.Unsetenv("GEMINI_MAX_TOKENS")
	os.Unsetenv("GEMINI_TEMPERATURE")
	os.Unsetenv("GEMINI_TOP_P")
	os.Unsetenv("GEMINI_TOP_K")
	os.Unsetenv("MAX_HISTORY_MESSAGES")
	os.Unsetenv("REQUEST_TIMEOUT")
	os.Unsetenv("SYSTEM_PROMPT")

	_, err := LoadConfig()

	// 必須設定が不足しているためエラーが発生することを期待
	if err == nil {
		t.Error("必須設定が不足しているのにエラーが発生しませんでした")
	}

	// エラーメッセージを確認
	if err.Error() == "" {
		t.Error("エラーメッセージが空です")
	}
}

func TestLoadConfig_WithInvalidValues(t *testing.T) {
	// 無効な値の環境変数を設定
	os.Setenv("DISCORD_BOT_TOKEN", "test-discord-token")
	os.Setenv("GEMINI_API_KEY", "test-gemini-key")
	os.Setenv("MAX_HISTORY_MESSAGES", "-1") // 無効な値
	os.Setenv("REQUEST_TIMEOUT", "invalid") // 無効な値

	defer func() {
		os.Unsetenv("DISCORD_BOT_TOKEN")
		os.Unsetenv("GEMINI_API_KEY")
		os.Unsetenv("MAX_HISTORY_MESSAGES")
		os.Unsetenv("REQUEST_TIMEOUT")
	}()

	_, err := LoadConfig()

	// 無効な値のためエラーが発生することを期待
	if err == nil {
		t.Error("無効な値なのにエラーが発生しませんでした")
	}
}

func TestConfig_Validate_ValidConfig(t *testing.T) {
	config := &Config{
		Discord: DiscordConfig{
			BotToken: "valid-token",
		},
		Gemini: GeminiConfig{
			APIKey: "valid-key",
		},
		Bot: BotConfig{
			MaxHistoryMessages: 10,
			RequestTimeout:     30 * time.Second,
			SystemPrompt:       "テストプロンプト",
		},
	}

	err := config.Validate()

	if err != nil {
		t.Errorf("有効な設定でエラーが発生しました: %v", err)
	}
}

func TestConfig_Validate_MissingDiscordToken(t *testing.T) {
	config := &Config{
		Discord: DiscordConfig{
			BotToken: "", // 空のトークン
		},
		Gemini: GeminiConfig{
			APIKey: "valid-key",
		},
		Bot: BotConfig{
			MaxHistoryMessages: 10,
			RequestTimeout:     30 * time.Second,
			SystemPrompt:       "テストプロンプト",
		},
	}

	err := config.Validate()

	if err == nil {
		t.Error("Discordトークンが不足しているのにエラーが発生しませんでした")
	}

	if err.Error() == "" {
		t.Error("エラーメッセージが空です")
	}
}

func TestConfig_Validate_MissingGeminiKey(t *testing.T) {
	config := &Config{
		Discord: DiscordConfig{
			BotToken: "valid-token",
		},
		Gemini: GeminiConfig{
			APIKey: "", // 空のAPIキー
		},
		Bot: BotConfig{
			MaxHistoryMessages: 10,
			RequestTimeout:     30 * time.Second,
			SystemPrompt:       "テストプロンプト",
		},
	}

	err := config.Validate()

	if err == nil {
		t.Error("Gemini APIキーが不足しているのにエラーが発生しませんでした")
	}

	if err.Error() == "" {
		t.Error("エラーメッセージが空です")
	}
}

func TestConfig_Validate_InvalidMaxHistoryMessages(t *testing.T) {
	config := &Config{
		Discord: DiscordConfig{
			BotToken: "valid-token",
		},
		Gemini: GeminiConfig{
			APIKey: "valid-key",
		},
		Bot: BotConfig{
			MaxHistoryMessages: 0, // 無効な値
			RequestTimeout:     30 * time.Second,
			SystemPrompt:       "テストプロンプト",
		},
	}

	err := config.Validate()

	if err == nil {
		t.Error("無効なMaxHistoryMessagesなのにエラーが発生しませんでした")
	}

	if err.Error() == "" {
		t.Error("エラーメッセージが空です")
	}
}

func TestConfig_Validate_InvalidRequestTimeout(t *testing.T) {
	config := &Config{
		Discord: DiscordConfig{
			BotToken: "valid-token",
		},
		Gemini: GeminiConfig{
			APIKey: "valid-key",
		},
		Bot: BotConfig{
			MaxHistoryMessages: 10,
			RequestTimeout:     0, // 無効な値
			SystemPrompt:       "テストプロンプト",
		},
	}

	err := config.Validate()

	if err == nil {
		t.Error("無効なRequestTimeoutなのにエラーが発生しませんでした")
	}

	if err.Error() == "" {
		t.Error("エラーメッセージが空です")
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	// 環境変数が設定されている場合
	os.Setenv("TEST_KEY", "test-value")
	defer os.Unsetenv("TEST_KEY")

	result := getEnvOrDefault("TEST_KEY", "default")
	if result != "test-value" {
		t.Errorf("期待される値: test-value, 実際: %s", result)
	}

	// 環境変数が設定されていない場合
	result = getEnvOrDefault("NONEXISTENT_KEY", "default")
	if result != "default" {
		t.Errorf("期待される値: default, 実際: %s", result)
	}
}

func TestGetEnvAsIntOrDefault(t *testing.T) {
	// 有効な整数値
	os.Setenv("TEST_INT", "123")
	defer os.Unsetenv("TEST_INT")

	result := getEnvAsIntOrDefault("TEST_INT", 0)
	if result != 123 {
		t.Errorf("期待される値: 123, 実際: %d", result)
	}

	// 無効な整数値
	os.Setenv("TEST_INVALID_INT", "invalid")
	defer os.Unsetenv("TEST_INVALID_INT")

	result = getEnvAsIntOrDefault("TEST_INVALID_INT", 456)
	if result != 456 {
		t.Errorf("期待される値: 456, 実際: %d", result)
	}

	// 環境変数が設定されていない場合
	result = getEnvAsIntOrDefault("NONEXISTENT_INT", 789)
	if result != 789 {
		t.Errorf("期待される値: 789, 実際: %d", result)
	}
}

func TestGetEnvAsFloatOrDefault(t *testing.T) {
	// 有効な浮動小数点数値
	os.Setenv("TEST_FLOAT", "3.14")
	defer os.Unsetenv("TEST_FLOAT")

	result := getEnvAsFloatOrDefault("TEST_FLOAT", 0.0)
	if result != 3.14 {
		t.Errorf("期待される値: 3.14, 実際: %f", result)
	}

	// 無効な浮動小数点数値
	os.Setenv("TEST_INVALID_FLOAT", "invalid")
	defer os.Unsetenv("TEST_INVALID_FLOAT")

	result = getEnvAsFloatOrDefault("TEST_INVALID_FLOAT", 2.71)
	if result != 2.71 {
		t.Errorf("期待される値: 2.71, 実際: %f", result)
	}

	// 環境変数が設定されていない場合
	result = getEnvAsFloatOrDefault("NONEXISTENT_FLOAT", 1.41)
	if result != 1.41 {
		t.Errorf("期待される値: 1.41, 実際: %f", result)
	}
}

func TestGetEnvAsDurationOrDefault(t *testing.T) {
	// 有効な時間値
	os.Setenv("TEST_DURATION", "30s")
	defer os.Unsetenv("TEST_DURATION")

	result := getEnvAsDurationOrDefault("TEST_DURATION", 0)
	if result != 30*time.Second {
		t.Errorf("期待される値: 30s, 実際: %v", result)
	}

	// 無効な時間値
	os.Setenv("TEST_INVALID_DURATION", "invalid")
	defer os.Unsetenv("TEST_INVALID_DURATION")

	result = getEnvAsDurationOrDefault("TEST_INVALID_DURATION", 60*time.Second)
	if result != 60*time.Second {
		t.Errorf("期待される値: 60s, 実際: %v", result)
	}

	// 環境変数が設定されていない場合
	result = getEnvAsDurationOrDefault("NONEXISTENT_DURATION", 120*time.Second)
	if result != 120*time.Second {
		t.Errorf("期待される値: 120s, 実際: %v", result)
	}
}
