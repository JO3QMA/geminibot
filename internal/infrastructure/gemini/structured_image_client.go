package gemini

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"geminibot/internal/domain"

	"google.golang.org/genai"
)

// GenerateImage は、プロンプトを受け取ってGemini APIから画像を生成します
func (g *StructuredGeminiClient) GenerateImage(ctx context.Context, prompt domain.ImagePrompt) (*domain.ImageGenerationResult, error) {
	log.Printf("構造化Geminiクライアントで画像生成をリクエスト中: %d文字", len(prompt.Content))
	log.Printf("プロンプト内容: %s", prompt.Content)

	// 画像生成用のコンテンツを作成
	contents := genai.Text(prompt.Content)

	// 画像生成用の設定を作成
	config := g.createImageGenerateConfig()

	// nano bananaモデルを使用
	modelName := "gemini-2.5-flash-image"
	if g.config.ModelName != "" {
		modelName = g.config.ModelName
	}

	resp, err := g.client.Models.GenerateContent(ctx, modelName, contents, config)
	if err != nil {
		return nil, fmt.Errorf("Gemini APIからの画像生成応答取得に失敗: %w", err)
	}

	// 画像生成結果を処理
	return g.processImageResponse(resp, prompt.Content, modelName)
}

// GenerateImageWithOptions は、オプション付きで画像を生成します
func (g *StructuredGeminiClient) GenerateImageWithOptions(ctx context.Context, prompt domain.ImagePrompt, options domain.ImageGenerationOptions) (*domain.ImageGenerationResult, error) {
	log.Printf("構造化Geminiクライアントでオプション付き画像生成をリクエスト中: %d文字", len(prompt.Content))
	log.Printf("プロンプト内容: %s", prompt.Content)
	log.Printf("オプション: %+v", options)

	// 画像生成用のコンテンツを作成
	contents := genai.Text(prompt.Content)

	// オプションに基づいて画像生成設定を作成
	config := g.createImageGenerateConfigWithOptions(options)

	// モデル名を決定
	modelName := options.Model
	if modelName == "" {
		modelName = "gemini-2.5-flash-image"
	}
	if g.config.ModelName != "" {
		modelName = g.config.ModelName
	}

	resp, err := g.client.Models.GenerateContent(ctx, modelName, contents, config)
	if err != nil {
		return nil, fmt.Errorf("Gemini APIからの画像生成応答取得に失敗: %w", err)
	}

	// 画像生成結果を処理
	return g.processImageResponse(resp, prompt.Content, modelName)
}

// createImageGenerateConfig は、画像生成用の設定を作成します
func (g *StructuredGeminiClient) createImageGenerateConfig() *genai.GenerateContentConfig {
	return &genai.GenerateContentConfig{
		MaxOutputTokens: g.config.MaxTokens,
		Temperature:     &g.config.Temperature,
		TopP:            &g.config.TopP,
		SafetySettings: []*genai.SafetySetting{
			{
				Category:  genai.HarmCategoryHarassment,
				Threshold: genai.HarmBlockThresholdBlockMediumAndAbove,
			},
			{
				Category:  genai.HarmCategoryHateSpeech,
				Threshold: genai.HarmBlockThresholdBlockMediumAndAbove,
			},
			{
				Category:  genai.HarmCategorySexuallyExplicit,
				Threshold: genai.HarmBlockThresholdBlockMediumAndAbove,
			},
			{
				Category:  genai.HarmCategoryDangerousContent,
				Threshold: genai.HarmBlockThresholdBlockMediumAndAbove,
			},
		},
	}
}

// createImageGenerateConfigWithOptions は、オプション付きで画像生成設定を作成します
func (g *StructuredGeminiClient) createImageGenerateConfigWithOptions(options domain.ImageGenerationOptions) *genai.GenerateContentConfig {
	config := &genai.GenerateContentConfig{
		SafetySettings: []*genai.SafetySetting{
			{
				Category:  genai.HarmCategoryHarassment,
				Threshold: genai.HarmBlockThresholdBlockMediumAndAbove,
			},
			{
				Category:  genai.HarmCategoryHateSpeech,
				Threshold: genai.HarmBlockThresholdBlockMediumAndAbove,
			},
			{
				Category:  genai.HarmCategorySexuallyExplicit,
				Threshold: genai.HarmBlockThresholdBlockMediumAndAbove,
			},
			{
				Category:  genai.HarmCategoryDangerousContent,
				Threshold: genai.HarmBlockThresholdBlockMediumAndAbove,
			},
		},
	}

	// オプションから設定値を適用
	if options.MaxTokens > 0 {
		config.MaxOutputTokens = options.MaxTokens
	} else {
		config.MaxOutputTokens = g.config.MaxTokens
	}

	if options.Temperature > 0 {
		config.Temperature = &options.Temperature
	} else {
		config.Temperature = &g.config.Temperature
	}

	if options.TopP > 0 {
		config.TopP = &options.TopP
	} else {
		config.TopP = &g.config.TopP
	}

	return config
}

// processImageResponse は、画像生成レスポンスを処理します
func (g *StructuredGeminiClient) processImageResponse(resp *genai.GenerateContentResponse, prompt, modelName string) (*domain.ImageGenerationResult, error) {
	if resp == nil {
		return &domain.ImageGenerationResult{
			Success: false,
			Error:   "レスポンスが空です",
		}, fmt.Errorf("レスポンスが空です")
	}

	// 安全フィルターのチェック
	if len(resp.Candidates) == 0 {
		return &domain.ImageGenerationResult{
			Success: false,
			Error:   "安全フィルターにより生成がブロックされました",
		}, fmt.Errorf("安全フィルターにより生成がブロックされました")
	}

	candidate := resp.Candidates[0]

	// 安全フィルターの詳細チェック
	if candidate.FinishReason == genai.FinishReasonSafety {
		safetyRatings := g.formatSafetyRatings(candidate.SafetyRatings)
		return &domain.ImageGenerationResult{
			Success: false,
			Error:   fmt.Sprintf("安全フィルターにより生成がブロックされました: %s", safetyRatings),
		}, fmt.Errorf("安全フィルターにより生成がブロックされました: %s", safetyRatings)
	}

	// 画像URLを抽出
	var imageURL string
	if len(candidate.Content.Parts) > 0 {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				// テキストから画像URLを抽出する処理
				imageURL = g.extractImageURLFromText(part.Text)
				if imageURL != "" {
					break
				}
			}
		}
	}

	if imageURL == "" {
		return &domain.ImageGenerationResult{
			Success: false,
			Error:   "画像URLが見つかりませんでした",
		}, fmt.Errorf("画像URLが見つかりませんでした")
	}

	return &domain.ImageGenerationResult{
		ImageURL:    imageURL,
		Prompt:      prompt,
		Model:       modelName,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Success:     true,
	}, nil
}

// formatSafetyRatings は、安全フィルターの評価をフォーマットします
func (g *StructuredGeminiClient) formatSafetyRatings(ratings []*genai.SafetyRating) string {
	if len(ratings) == 0 {
		return "評価なし"
	}

	var result []string
	for _, rating := range ratings {
		category := g.translateSafetyCategory(rating.Category)
		probability := g.translateSafetyProbability(rating.Probability)
		result = append(result, fmt.Sprintf("%s: %s", category, probability))
	}

	return strings.Join(result, ", ")
}

// translateSafetyCategory は、安全フィルターのカテゴリを日本語に翻訳します
func (g *StructuredGeminiClient) translateSafetyCategory(category genai.HarmCategory) string {
	switch category {
	case genai.HarmCategoryHarassment:
		return "ハラスメント"
	case genai.HarmCategoryHateSpeech:
		return "ヘイトスピーチ"
	case genai.HarmCategorySexuallyExplicit:
		return "性的表現"
	case genai.HarmCategoryDangerousContent:
		return "危険なコンテンツ"
	default:
		return "不明"
	}
}

// translateSafetyProbability は、安全フィルターの確率を日本語に翻訳します
func (g *StructuredGeminiClient) translateSafetyProbability(probability genai.HarmProbability) string {
	switch probability {
	case genai.HarmProbabilityNegligible:
		return "無視できる"
	case genai.HarmProbabilityLow:
		return "低"
	case genai.HarmProbabilityMedium:
		return "中"
	case genai.HarmProbabilityHigh:
		return "高"
	default:
		return "不明"
	}
}

// extractImageURLFromText は、テキストから画像URLを抽出します
func (g *StructuredGeminiClient) extractImageURLFromText(text string) string {
	// 基本的なURL抽出ロジック
	// 実際の実装では、より複雑なパターンマッチングが必要かもしれません
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "http") && (strings.Contains(line, ".jpg") || strings.Contains(line, ".png") || strings.Contains(line, ".jpeg") || strings.Contains(line, ".gif")) {
			return line
		}
	}
	return ""
}
