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

// createImageGenerateConfig は、画像生成用の設定を作成します
func (g *GeminiAPIClient) createImageGenerateConfig() *genai.GenerateContentConfig {
	return &genai.GenerateContentConfig{
		MaxOutputTokens: g.config.MaxTokens,
		Temperature:     &g.config.Temperature,
		TopP:            &g.config.TopP,
		SafetySettings:  g.createSafetySettings(),
	}
}

// createImageGenerateConfigWithOptions は、オプション付きで画像生成設定を作成します
func (g *GeminiAPIClient) createImageGenerateConfigWithOptions(options domain.ImageGenerationOptions) *genai.GenerateContentConfig {
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

// retryWithBackoffForImage は、画像生成用のリトライ機能付きで関数を実行します
func (g *GeminiAPIClient) retryWithBackoffForImage(ctx context.Context, fn func() (*domain.ImageGenerationResult, error)) (*domain.ImageGenerationResult, error) {
	var lastErr error

	for attempt := 0; attempt < g.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// 指数バックオフで待機
			backoffDuration := time.Duration(attempt*attempt) * time.Second
			log.Printf("画像生成リトライ %d/%d: %v後に再試行", attempt+1, g.config.MaxRetries, backoffDuration)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoffDuration):
			}
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err
		log.Printf("画像生成試行 %d/%d 失敗: %v", attempt+1, g.config.MaxRetries, err)

		// 致命的なエラーの場合はリトライしない
		if g.isFatalErrorForImage(err) {
			break
		}
	}

	return nil, fmt.Errorf("画像生成が %d 回の試行後に失敗しました: %w", g.config.MaxRetries, lastErr)
}

// processImageResponse は、画像生成レスポンスを処理します
func (g *GeminiAPIClient) processImageResponse(resp *genai.GenerateContentResponse, prompt, modelName string) (*domain.ImageGenerationResult, error) {
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

// extractImageURLFromText は、テキストから画像URLを抽出します
func (g *GeminiAPIClient) extractImageURLFromText(text string) string {
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

// isFatalErrorForImage は、画像生成で致命的なエラーかどうかを判定します
func (g *GeminiAPIClient) isFatalErrorForImage(err error) bool {
	if err == nil {
		return false
	}

	errorMsg := err.Error()
	
	// 認証エラー
	if strings.Contains(errorMsg, "authentication") || strings.Contains(errorMsg, "unauthorized") {
		return true
	}
	
	// APIキーエラー
	if strings.Contains(errorMsg, "API key") || strings.Contains(errorMsg, "invalid key") {
		return true
	}
	
	// レート制限エラー（一時的なものなのでリトライする）
	if strings.Contains(errorMsg, "rate limit") || strings.Contains(errorMsg, "quota") {
		return false
	}
	
	// その他のエラーはリトライする
	return false
}
