package gemini

import (
	"context"
	"fmt"
	"log"
	"time"

	"geminibot/internal/domain"

	"google.golang.org/genai"
)

// GenerateImage は、プロンプトを受け取ってGemini APIから画像を生成します
// optionsが空の場合はデフォルト設定を使用します
func (g *StructuredGeminiClient) GenerateImage(ctx context.Context, request domain.ImageGenerationRequest) (*domain.ImageGenerationResponse, error) {
	if request.Options == (domain.ImageGenerationOptions{}) {
		request.Options = domain.DefaultImageGenerationOptions()
	}
	return g.generateImageWithOptions(ctx, request.Prompt, request.Options)
}

// generateImageWithOptions は、オプション付きで画像を生成する内部実装です
func (g *StructuredGeminiClient) generateImageWithOptions(ctx context.Context, prompt string, options domain.ImageGenerationOptions) (*domain.ImageGenerationResponse, error) {
	log.Printf("構造化Geminiクライアントで画像生成をリクエスト中: %d文字", len(prompt))
	log.Printf("プロンプト内容: %s", prompt)
	log.Printf("オプション: %+v", options)

	// 画像生成用のコンテンツを作成
	contents := genai.Text(prompt)

	// オプションに基づいて画像生成設定を作成
	config := g.createImageConfig(options)

	// モデル名を決定
	modelName := options.Model

	resp, err := g.client.Models.GenerateContent(ctx, modelName, contents, config)
	if err != nil {
		return nil, fmt.Errorf("Gemini APIからの画像生成応答取得に失敗: %w", err)
	}

	// 画像生成結果を処理
	return g.processImageResponse(resp, prompt, modelName)
}

// createImageConfig は、画像生成設定を作成します
func (g *StructuredGeminiClient) createImageConfig(options domain.ImageGenerationOptions) *genai.GenerateContentConfig {
	config := &genai.GenerateContentConfig{
		SafetySettings: g.createSafetySettings(),
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
func (g *StructuredGeminiClient) processImageResponse(resp *genai.GenerateContentResponse, prompt string, modelName string) (*domain.ImageGenerationResponse, error) {
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("Gemini APIから有効な画像生成応答が得られませんでした")
	}

	candidate := resp.Candidates[0]

	// FinishReasonをチェックして安全フィルターによるブロックを検出
	if candidate.FinishReason == "SAFETY" {
		safetyDetails := g.formatSafetyRatings(candidate.SafetyRatings)
		return nil, fmt.Errorf("Gemini APIの安全フィルターによって画像生成がブロックされました。詳細: %s", safetyDetails)
	}

	if candidate.FinishReason == "RECITATION" {
		return nil, fmt.Errorf("Gemini APIが著作権保護された内容を検出しました。著作権で保護されたコンテンツが含まれている可能性があります")
	}

	if candidate.FinishReason == "MAX_TOKENS" {
		return nil, fmt.Errorf("Gemini APIの応答が最大トークン数に達しました。より短いプロンプトを試してください")
	}

	// Contentがnilの場合のチェック
	if candidate.Content == nil {
		return nil, fmt.Errorf("Gemini APIの画像生成応答にContentが含まれていません。FinishReason: %s", candidate.FinishReason)
	}

	if len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("Gemini APIの画像生成応答にコンテンツが含まれていません。FinishReason: %s", candidate.FinishReason)
	}

	// 画像データを抽出
	var images []domain.GeneratedImage
	for i, part := range candidate.Content.Parts {
		if part != nil && part.InlineData != nil {
			image := domain.GeneratedImage{
				Data:        part.InlineData.Data,
				MimeType:    part.InlineData.MIMEType,
				Filename:    fmt.Sprintf("generated_image_%d.png", i+1),
				Size:        int64(len(part.InlineData.Data)),
				GeneratedAt: time.Now(),
			}
			images = append(images, image)
		}
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("Gemini APIから画像データが取得できませんでした")
	}

	log.Printf("Gemini APIから画像を生成: %d枚", len(images))

	return &domain.ImageGenerationResponse{
		Images:      images,
		Prompt:      prompt,
		Model:       modelName,
		GeneratedAt: time.Now(),
	}, nil
}

// createSafetySettings は、安全フィルター設定を作成します
func (g *StructuredGeminiClient) createSafetySettings() []*genai.SafetySetting {
	return []*genai.SafetySetting{
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
	}
}

// formatSafetyRatings は、SafetyRatingsの詳細情報をフォーマットします
func (g *StructuredGeminiClient) formatSafetyRatings(ratings []*genai.SafetyRating) string {
	if len(ratings) == 0 {
		return "詳細情報なし"
	}

	var details []string
	for _, rating := range ratings {
		if rating != nil {
			category := g.translateSafetyCategory(rating.Category)
			probability := g.translateSafetyProbability(rating.Probability)
			details = append(details, fmt.Sprintf("%s: %s", category, probability))
		}
	}

	if len(details) == 0 {
		return "詳細情報なし"
	}

	return fmt.Sprintf("%s", details)
}

// translateSafetyCategory は、SafetyCategoryを日本語に翻訳します
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
		return string(category)
	}
}

// translateSafetyProbability は、SafetyProbabilityを日本語に翻訳します
func (g *StructuredGeminiClient) translateSafetyProbability(probability genai.HarmProbability) string {
	switch probability {
	case genai.HarmProbabilityNegligible:
		return "無視できるレベル"
	case genai.HarmProbabilityLow:
		return "低レベル"
	case genai.HarmProbabilityMedium:
		return "中レベル"
	case genai.HarmProbabilityHigh:
		return "高レベル"
	default:
		return string(probability)
	}
}
