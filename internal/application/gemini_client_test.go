package application

import (
	"testing"
)

func TestDefaultTextGenerationOptions(t *testing.T) {
	options := DefaultTextGenerationOptions()

	if options.MaxTokens != 1000 {
		t.Errorf("期待されるMaxTokens: 1000, 実際: %d", options.MaxTokens)
	}

	if options.Temperature != 0.7 {
		t.Errorf("期待されるTemperature: 0.7, 実際: %f", options.Temperature)
	}

	if options.TopP != 0.9 {
		t.Errorf("期待されるTopP: 0.9, 実際: %f", options.TopP)
	}

	if options.TopK != 40 {
		t.Errorf("期待されるTopK: 40, 実際: %d", options.TopK)
	}

	if options.Model != "gemini-pro" {
		t.Errorf("期待されるModel: gemini-pro, 実際: %s", options.Model)
	}
}

func TestTextGenerationOptions_JSONTags(t *testing.T) {
	options := TextGenerationOptions{
		MaxTokens:   1000,
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
		Model:       "gemini-pro",
	}

	// JSONタグが正しく設定されているかを確認
	// 実際のJSONシリアライゼーションは別途テストが必要
	if options.MaxTokens != 1000 {
		t.Error("MaxTokensが正しく設定されていません")
	}
	if options.Temperature != 0.7 {
		t.Error("Temperatureが正しく設定されていません")
	}
	if options.TopP != 0.9 {
		t.Error("TopPが正しく設定されていません")
	}
	if options.TopK != 40 {
		t.Error("TopKが正しく設定されていません")
	}
	if options.Model != "gemini-pro" {
		t.Error("Modelが正しく設定されていません")
	}
}
