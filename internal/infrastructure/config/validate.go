package config

import "fmt"

// Validate は、アプリケーション設定の妥当性を検証します。
func (c *AppConfig) Validate() error {
	if c.Discord.BotToken == "" {
		return fmt.Errorf("DISCORD_BOT_TOKEN が設定されていません")
	}

	if c.Gemini.APIKey == "" {
		return fmt.Errorf("GEMINI_API_KEY が設定されていません")
	}

	if c.Gemini.MaxTokens <= 0 {
		return fmt.Errorf("GEMINI_MAX_TOKENS は正の整数である必要があります")
	}

	if c.Gemini.Temperature < 0 || c.Gemini.Temperature > 2 {
		return fmt.Errorf("GEMINI_TEMPERATURE は0以上2以下の値である必要があります")
	}

	if c.Gemini.TopP < 0 || c.Gemini.TopP > 1 {
		return fmt.Errorf("GEMINI_TOP_P は0以上1以下の値である必要があります")
	}

	if c.Gemini.TopK <= 0 {
		return fmt.Errorf("GEMINI_TOP_K は正の整数である必要があります")
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

	if c.Gemini.MaxRetries < 0 {
		return fmt.Errorf("GEMINI_MAX_RETRIES は0以上の整数である必要があります")
	}

	return nil
}
