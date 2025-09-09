package gemini

import (
	"context"
	"fmt"
	"log"
	"time"

	"geminibot/internal/domain"

	"google.golang.org/genai"
)

// retryWithBackoffForImage は、画像生成用の指数バックオフでリトライを実行します
func (g *GeminiAPIClient) retryWithBackoffForImage(ctx context.Context, operation func() (*domain.ImageGenerationResponse, error)) (*domain.ImageGenerationResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= g.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// 指数バックオフ: 1秒、2秒、4秒...
			backoffDuration := time.Duration(1<<uint(attempt-1)) * time.Second
			log.Printf("画像生成リトライ %d/%d 回目: %v 後に再試行します", attempt, g.config.MaxRetries, backoffDuration)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoffDuration):
			}
		}

		result, err := operation()
		if err == nil {
			if attempt > 0 {
				log.Printf("画像生成リトライ成功: %d回目の試行で成功しました", attempt+1)
			}
			return result, nil
		}

		lastErr = err

		// リトライ可能なエラーかチェック
		if !g.shouldRetry(err) {
			log.Printf("画像生成リトライ不可能なエラー: %v", err)
			return nil, err
		}

		if attempt < g.config.MaxRetries {
			log.Printf("画像生成リトライ可能なエラーが発生: %v", err)
		}
	}

	return nil, fmt.Errorf("画像生成で最大リトライ回数 (%d) に達しました。最後のエラー: %w", g.config.MaxRetries, lastErr)
}

// createImageConfig は、画像生成設定を作成します
func (g *GeminiAPIClient) createImageConfig(options domain.ImageGenerationOptions) *genai.GenerateContentConfig {
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
func (g *GeminiAPIClient) processImageResponse(resp *genai.GenerateContentResponse, prompt string, modelName string) (*domain.ImageGenerationResponse, error) {
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
