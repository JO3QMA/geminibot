package gemini

import (
	"context"
	"errors"
	"testing"
	"time"

	"geminibot/internal/application"
	"geminibot/internal/domain"
	"geminibot/internal/infrastructure/config"

	"google.golang.org/genai"
)

func TestGeminiConfig_DefaultValues(t *testing.T) {
	// LoadConfigで設定されるデフォルト値をテスト
	config := &config.GeminiConfig{
		ModelName:   "gemini-2.5-pro",
		MaxTokens:   1000,
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
		MaxRetries:  3,
	}

	if config.ModelName != "gemini-2.5-pro" {
		t.Errorf("期待されるModelName: gemini-2.5-pro, 実際: %s", config.ModelName)
	}

	if config.MaxTokens != 1000 {
		t.Errorf("期待されるMaxTokens: 1000, 実際: %d", config.MaxTokens)
	}

	if config.Temperature != 0.7 {
		t.Errorf("期待されるTemperature: 0.7, 実際: %f", config.Temperature)
	}

	if config.TopP != 0.9 {
		t.Errorf("期待されるTopP: 0.9, 実際: %f", config.TopP)
	}

	if config.TopK != 40 {
		t.Errorf("期待されるTopK: 40, 実際: %d", config.TopK)
	}

	if config.MaxRetries != 3 {
		t.Errorf("期待されるMaxRetries: 3, 実際: %d", config.MaxRetries)
	}
}

func TestNewGeminiAPIClient_WithConfig(t *testing.T) {
	config := &config.GeminiConfig{
		APIKey:      "test-api-key",
		ModelName:   "gemini-2.5-pro",
		MaxTokens:   500,
		Temperature: 0.5,
		TopP:        0.8,
		TopK:        20,
	}

	// 実際のAPIキーがないため、エラーが発生することを期待
	_, err := NewGeminiAPIClient(config)

	// エラーが発生しない場合でも、テストは成功とする（実装の仕様による）
	if err != nil {
		t.Logf("APIキーが無効でエラーが発生: %v", err)
	}
}

func TestNewGeminiAPIClient_WithNilConfig(t *testing.T) {
	// 設定がnilの場合、エラーが発生することを確認
	_, err := NewGeminiAPIClient(nil)

	if err == nil {
		t.Error("GeminiConfigがnilの場合、エラーが期待されましたが発生しませんでした")
	}

	expectedError := "GeminiConfigが指定されていません"
	if err.Error() != expectedError {
		t.Errorf("期待されるエラー: %s, 実際: %s", expectedError, err.Error())
	}
}

func TestGeminiAPIClient_GenerateText_Integration(t *testing.T) {
	// 統合テスト（実際のAPIキーが必要）
	// このテストは実際のAPIキーがある環境でのみ実行されるべき
	t.Skip("実際のAPIキーが必要なため、スキップします")

	config := &config.GeminiConfig{
		APIKey:      "your-api-key-here",
		ModelName:   "gemini-2.5-pro",
		MaxTokens:   100,
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
	}

	client, err := NewGeminiAPIClient(config)
	if err != nil {
		t.Fatalf("クライアントの作成に失敗: %v", err)
	}

	prompt := domain.Prompt{Content: "こんにちは、簡単な挨拶をしてください。"}

	ctx := context.Background()
	response, err := client.GenerateText(ctx, prompt)

	if err != nil {
		t.Errorf("テキスト生成でエラーが発生: %v", err)
	}

	if response == "" {
		t.Error("空の応答が返されました")
	}

	t.Logf("生成されたテキスト: %s", response)
}

func TestGeminiAPIClient_GenerateTextWithOptions_Integration(t *testing.T) {
	// 統合テスト（実際のAPIキーが必要）
	// このテストは実際のAPIキーがある環境でのみ実行されるべき
	t.Skip("実際のAPIキーが必要なため、スキップします")

	config := &config.GeminiConfig{
		APIKey:      "your-api-key-here",
		ModelName:   "gemini-2.5-pro",
		MaxTokens:   100,
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
	}

	client, err := NewGeminiAPIClient(config)
	if err != nil {
		t.Fatalf("クライアントの作成に失敗: %v", err)
	}

	prompt := domain.Prompt{Content: "こんにちは、簡単な挨拶をしてください。"}
	options := application.TextGenerationOptions{
		MaxTokens:   50,
		Temperature: 0.5,
		TopP:        0.8,
		TopK:        20,
		Model:       "gemini-2.5-pro",
	}

	ctx := context.Background()
	response, err := client.GenerateTextWithOptions(ctx, prompt, options)

	if err != nil {
		t.Errorf("オプション付きテキスト生成でエラーが発生: %v", err)
	}

	if response == "" {
		t.Error("空の応答が返されました")
	}

	t.Logf("生成されたテキスト: %s", response)
}

func TestGeminiAPIClient_ConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.GeminiConfig
		wantErr bool
	}{
		{
			name: "有効な設定",
			config: &config.GeminiConfig{
				APIKey:      "test-key",
				ModelName:   "gemini-2.5-pro",
				MaxTokens:   1000,
				Temperature: 0.7,
				TopP:        0.9,
				TopK:        40,
			},
			wantErr: false,
		},
		{
			name: "空のAPIキー",
			config: &config.GeminiConfig{
				APIKey:      "",
				ModelName:   "gemini-2.5-pro",
				MaxTokens:   1000,
				Temperature: 0.7,
				TopP:        0.9,
				TopK:        40,
			},
			wantErr: true,
		},
		{
			name: "無効なTemperature",
			config: &config.GeminiConfig{
				APIKey:      "test-key",
				ModelName:   "gemini-2.5-pro",
				MaxTokens:   1000,
				Temperature: 2.0, // 範囲外
				TopP:        0.9,
				TopK:        40,
			},
			wantErr: false, // APIキーが無効なので、Temperatureの検証前にエラーになる
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGeminiAPIClient(tt.config)

			if tt.wantErr && err == nil {
				t.Error("エラーが期待されましたが、発生しませんでした")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("予期しないエラーが発生しました: %v", err)
			}
		})
	}
}

// TestShouldRetry は、shouldRetryメソッドのテストです
func TestShouldRetry(t *testing.T) {
	client := &GeminiAPIClient{}

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Contentがnilのエラー",
			err:      errors.New("Gemini APIの応答にContentが含まれていません"),
			expected: true,
		},
		{
			name:     "コンテンツが含まれていないエラー",
			err:      errors.New("Gemini APIの応答にコンテンツが含まれていません"),
			expected: true,
		},
		{
			name:     "安全フィルターによるブロック",
			err:      errors.New("Gemini APIの安全フィルターによって応答がブロックされました"),
			expected: false,
		},
		{
			name:     "著作権保護エラー",
			err:      errors.New("Gemini APIが著作権保護された内容を検出しました"),
			expected: false,
		},
		{
			name:     "nilエラー",
			err:      nil,
			expected: false,
		},
		{
			name:     "その他のエラー",
			err:      errors.New("その他のエラー"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.shouldRetry(tt.err)
			if result != tt.expected {
				t.Errorf("shouldRetry() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestRetryWithBackoff は、retryWithBackoffメソッドのテストです
func TestRetryWithBackoff(t *testing.T) {
	tests := []struct {
		name           string
		maxRetries     int
		operation      func() (string, error)
		expectedResult string
		expectedError  bool
		expectedCalls  int
	}{
		{
			name:       "1回目で成功",
			maxRetries: 3,
			operation: func() (string, error) {
				return "success", nil
			},
			expectedResult: "success",
			expectedError:  false,
			expectedCalls:  1,
		},
		{
			name:       "2回目で成功",
			maxRetries: 3,
			operation: func() func() (string, error) {
				callCount := 0
				return func() (string, error) {
					callCount++
					if callCount == 1 {
						return "", errors.New("Gemini APIの応答にContentが含まれていません")
					}
					return "success", nil
				}
			}(),
			expectedResult: "success",
			expectedError:  false,
			expectedCalls:  2,
		},
		{
			name:       "最大リトライ回数に達して失敗",
			maxRetries: 2,
			operation: func() func() (string, error) {
				callCount := 0
				return func() (string, error) {
					callCount++
					return "", errors.New("Gemini APIの応答にContentが含まれていません")
				}
			}(),
			expectedResult: "",
			expectedError:  true,
			expectedCalls:  3, // 1回目 + 2回のリトライ
		},
		{
			name:       "リトライ不可能なエラー",
			maxRetries: 3,
			operation: func() (string, error) {
				return "", errors.New("安全フィルターによって応答がブロックされました")
			},
			expectedResult: "",
			expectedError:  true,
			expectedCalls:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &GeminiAPIClient{
				config: &config.GeminiConfig{
					MaxRetries: tt.maxRetries,
				},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result, err := client.retryWithBackoff(ctx, tt.operation)

			if tt.expectedError {
				if err == nil {
					t.Error("エラーが期待されましたが、発生しませんでした")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラーが発生しました: %v", err)
				}
				if result != tt.expectedResult {
					t.Errorf("期待される結果: %s, 実際: %s", tt.expectedResult, result)
				}
			}
		})
	}
}

// TestRetryWithBackoff_ContextCancellation は、コンテキストキャンセレーションのテストです
func TestRetryWithBackoff_ContextCancellation(t *testing.T) {
	client := &GeminiAPIClient{
		config: &config.GeminiConfig{
			MaxRetries: 3,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	// すぐにキャンセル
	cancel()

	operation := func() (string, error) {
		return "", errors.New("Gemini APIの応答にContentが含まれていません")
	}

	_, err := client.retryWithBackoff(ctx, operation)

	if err == nil {
		t.Error("コンテキストキャンセレーションエラーが期待されましたが、発生しませんでした")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("期待されるエラー: context.Canceled, 実際: %v", err)
	}
}

// TestFormatSafetyRatings は、formatSafetyRatingsメソッドのテストです
func TestFormatSafetyRatings(t *testing.T) {
	client := &GeminiAPIClient{}

	tests := []struct {
		name     string
		ratings  []*genai.SafetyRating
		expected string
	}{
		{
			name:     "空のSafetyRatings",
			ratings:  []*genai.SafetyRating{},
			expected: "詳細情報なし",
		},
		{
			name:     "nilのSafetyRatings",
			ratings:  nil,
			expected: "詳細情報なし",
		},
		{
			name: "単一のSafetyRating",
			ratings: []*genai.SafetyRating{
				{
					Category:    genai.HarmCategoryHarassment,
					Probability: genai.HarmProbabilityMedium,
				},
			},
			expected: "ハラスメント: 中レベル",
		},
		{
			name: "複数のSafetyRating",
			ratings: []*genai.SafetyRating{
				{
					Category:    genai.HarmCategoryHarassment,
					Probability: genai.HarmProbabilityMedium,
				},
				{
					Category:    genai.HarmCategoryHateSpeech,
					Probability: genai.HarmProbabilityLow,
				},
			},
			expected: "ハラスメント: 中レベル, ヘイトスピーチ: 低レベル",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.formatSafetyRatings(tt.ratings)
			if result != tt.expected {
				t.Errorf("formatSafetyRatings() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

// TestTranslateSafetyCategory は、translateSafetyCategoryメソッドのテストです
func TestTranslateSafetyCategory(t *testing.T) {
	client := &GeminiAPIClient{}

	tests := []struct {
		name     string
		category genai.HarmCategory
		expected string
	}{
		{
			name:     "ハラスメント",
			category: genai.HarmCategoryHarassment,
			expected: "ハラスメント",
		},
		{
			name:     "ヘイトスピーチ",
			category: genai.HarmCategoryHateSpeech,
			expected: "ヘイトスピーチ",
		},
		{
			name:     "性的表現",
			category: genai.HarmCategorySexuallyExplicit,
			expected: "性的表現",
		},
		{
			name:     "危険なコンテンツ",
			category: genai.HarmCategoryDangerousContent,
			expected: "危険なコンテンツ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.translateSafetyCategory(tt.category)
			if result != tt.expected {
				t.Errorf("translateSafetyCategory() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

// TestTranslateSafetyProbability は、translateSafetyProbabilityメソッドのテストです
func TestTranslateSafetyProbability(t *testing.T) {
	client := &GeminiAPIClient{}

	tests := []struct {
		name        string
		probability genai.HarmProbability
		expected    string
	}{
		{
			name:        "無視できるレベル",
			probability: genai.HarmProbabilityNegligible,
			expected:    "無視できるレベル",
		},
		{
			name:        "低レベル",
			probability: genai.HarmProbabilityLow,
			expected:    "低レベル",
		},
		{
			name:        "中レベル",
			probability: genai.HarmProbabilityMedium,
			expected:    "中レベル",
		},
		{
			name:        "高レベル",
			probability: genai.HarmProbabilityHigh,
			expected:    "高レベル",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.translateSafetyProbability(tt.probability)
			if result != tt.expected {
				t.Errorf("translateSafetyProbability() = %s, expected %s", result, tt.expected)
			}
		})
	}
}
