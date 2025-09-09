package gemini

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"geminibot/internal/domain"

	"google.golang.org/genai"
)

// createImageGenerateConfig は、画像生成用の設定を作成します
func (g *GeminiAPIClient) createImageGenerateConfig() *genai.GenerateContentConfig {
	// 画像生成用はMaxTokensを増加（複数画像生成に対応）
	maxTokens := g.config.MaxTokens * 2
	if maxTokens < 2000 {
		maxTokens = 2000
	}

	return &genai.GenerateContentConfig{
		MaxOutputTokens: maxTokens,
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

	// 詳細なログ出力
	log.Printf("画像生成レスポンス詳細:")
	log.Printf("  FinishReason: %v", candidate.FinishReason)
	log.Printf("  Parts数: %d", len(candidate.Content.Parts))

	for i, part := range candidate.Content.Parts {
		log.Printf("  Part[%d]: Text長=%d", i, len(part.Text))
		if len(part.Text) > 0 {
			log.Printf("  Part[%d]内容: %s", i, part.Text)
		}
	}

	// 安全フィルターの詳細チェック
	if candidate.FinishReason == genai.FinishReasonSafety {
		safetyRatings := g.formatSafetyRatings(candidate.SafetyRatings)
		return &domain.ImageGenerationResult{
			Success: false,
			Error:   fmt.Sprintf("安全フィルターにより生成がブロックされました: %s", safetyRatings),
		}, fmt.Errorf("安全フィルターにより生成がブロックされました: %s", safetyRatings)
	}

	// MAX_TOKENSの場合は、生成されたテキストをそのまま返す
	if candidate.FinishReason == genai.FinishReasonMaxTokens {
		log.Printf("MAX_TOKENSで終了 - 生成されたテキストを返します")
		if len(candidate.Content.Parts) > 0 && candidate.Content.Parts[0].Text != "" {
			// テキスト生成として処理
			return &domain.ImageGenerationResult{
				ImageURL:    candidate.Content.Parts[0].Text,
				Prompt:      prompt,
				Model:       modelName,
				GeneratedAt: time.Now().Format(time.RFC3339),
				Success:     true,
			}, nil
		}
	}

	// 画像URLを抽出
	var imageURL string
	if len(candidate.Content.Parts) > 0 {
		for i, part := range candidate.Content.Parts {
			if part.Text != "" {
				log.Printf("Part[%d]から画像URLを抽出中: %s", i, part.Text)
				// テキストから画像URLを抽出する処理
				imageURL = g.extractImageURLFromText(part.Text)
				if imageURL != "" {
					log.Printf("画像URLを発見: %s", imageURL)
					break
				}
			}
		}
	}

	if imageURL == "" {
		// 画像URLが見つからない場合、生成されたテキストをそのまま返す
		if len(candidate.Content.Parts) > 0 && candidate.Content.Parts[0].Text != "" {
			log.Printf("画像URLが見つからないため、生成されたテキストを返します: %s", candidate.Content.Parts[0].Text)
			return &domain.ImageGenerationResult{
				ImageURL:    candidate.Content.Parts[0].Text,
				Prompt:      prompt,
				Model:       modelName,
				GeneratedAt: time.Now().Format(time.RFC3339),
				Success:     true,
			}, nil
		}

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

// processImageResponseWithMultipleImages は、複数画像生成レスポンスを処理します
func (g *GeminiAPIClient) processImageResponseWithMultipleImages(resp *genai.GenerateContentResponse, prompt, modelName string) (*domain.ImageGenerationResult, error) {
	if resp == nil {
		return &domain.ImageGenerationResult{
			Success: false,
			Error:   "レスポンスが空です",
		}, nil
	}

	if len(resp.Candidates) == 0 {
		return &domain.ImageGenerationResult{
			Success: false,
			Error:   "候補がありません",
		}, nil
	}

	candidate := resp.Candidates[0]
	log.Printf("画像生成レスポンス詳細:")
	log.Printf("  FinishReason: %s", candidate.FinishReason)
	log.Printf("  Parts数: %d", len(candidate.Content.Parts))

	// すべての画像URLを抽出
	var allImageURLs []string
	var fullText strings.Builder

	for i, part := range candidate.Content.Parts {
		if part.Text != "" {
			log.Printf("  Part[%d]: Text長=%d", i, len(part.Text))
			log.Printf("  Part[%d]内容: %s", i, part.Text)
			
			fullText.WriteString(part.Text)
			fullText.WriteString("\n")
			
			// このPartから画像URLを抽出
			imageURLs := g.extractAllImageURLsFromText(part.Text)
			allImageURLs = append(allImageURLs, imageURLs...)
		}
	}

	// 画像URLが見つからない場合
	if len(allImageURLs) == 0 {
		log.Printf("画像URLが見つかりませんでした。テキストレスポンスを返します。")
		return &domain.ImageGenerationResult{
			ImageURL:    fullText.String(),
			Prompt:      prompt,
			Model:       modelName,
			GeneratedAt: time.Now().Format(time.RFC3339),
			Success:     true,
		}, nil
	}

	// 最初の画像URLを返す（後で複数対応を拡張可能）
	firstImageURL := allImageURLs[0]
	log.Printf("複数画像から最初の画像URLを選択: %s", firstImageURL)
	log.Printf("合計 %d 個の画像URLを発見", len(allImageURLs))

	return &domain.ImageGenerationResult{
		ImageURL:    firstImageURL,
		Prompt:      prompt,
		Model:       modelName,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Success:     true,
	}, nil
}

// extractImageURLFromText は、テキストから画像URLを抽出します
func (g *GeminiAPIClient) extractImageURLFromText(text string) string {
	log.Printf("テキストから画像URLを抽出中: %s", text)

	// Markdown形式の画像URLを抽出: ![alt](url)
	markdownPattern := `!\[.*?\]\((https?://[^)]+)\)`
	re := regexp.MustCompile(markdownPattern)
	matches := re.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) > 1 {
			url := match[1]
			log.Printf("Markdown形式の画像URLを発見: %s", url)
			return url
		}
	}

	// 通常のURL抽出ロジック
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// HTTP/HTTPSで始まるURLを探す
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			// 画像ファイル拡張子をチェック
			lowerLine := strings.ToLower(line)
			if strings.Contains(lowerLine, ".jpg") || strings.Contains(lowerLine, ".png") ||
				strings.Contains(lowerLine, ".jpeg") || strings.Contains(lowerLine, ".gif") ||
				strings.Contains(lowerLine, ".webp") || strings.Contains(lowerLine, ".bmp") {
				log.Printf("画像URLを発見: %s", line)
				return line
			}

			// 画像ホスティングサービスのURLパターンをチェック
			if strings.Contains(lowerLine, "imgur.com") || strings.Contains(lowerLine, "i.imgur.com") ||
				strings.Contains(lowerLine, "drive.google.com") || strings.Contains(lowerLine, "photos.google.com") ||
				strings.Contains(lowerLine, "cloudinary.com") || strings.Contains(lowerLine, "unsplash.com") ||
				strings.Contains(lowerLine, "files.oaiusercontent.com") {
				log.Printf("画像ホスティングサービスURLを発見: %s", line)
				return line
			}
		}
	}

	log.Printf("画像URLが見つかりませんでした")
	return ""
}

// extractAllImageURLsFromText は、テキストからすべての画像URLを抽出します
func (g *GeminiAPIClient) extractAllImageURLsFromText(text string) []string {
	var urls []string
	log.Printf("テキストからすべての画像URLを抽出中: %s", text)

	// Markdown形式の画像URLを抽出: ![alt](url)
	markdownPattern := `!\[.*?\]\((https?://[^)]+)\)`
	re := regexp.MustCompile(markdownPattern)
	matches := re.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) > 1 {
			url := match[1]
			urls = append(urls, url)
			log.Printf("Markdown形式の画像URLを発見: %s", url)
		}
	}

	// 通常のURL抽出ロジック
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// HTTP/HTTPSで始まるURLを探す
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			// 画像ファイル拡張子をチェック
			lowerLine := strings.ToLower(line)
			if strings.Contains(lowerLine, ".jpg") || strings.Contains(lowerLine, ".png") ||
				strings.Contains(lowerLine, ".jpeg") || strings.Contains(lowerLine, ".gif") ||
				strings.Contains(lowerLine, ".webp") || strings.Contains(lowerLine, ".bmp") {
				urls = append(urls, line)
				log.Printf("画像URLを発見: %s", line)
			}

			// 画像ホスティングサービスのURLパターンをチェック
			if strings.Contains(lowerLine, "imgur.com") || strings.Contains(lowerLine, "i.imgur.com") ||
				strings.Contains(lowerLine, "drive.google.com") || strings.Contains(lowerLine, "photos.google.com") ||
				strings.Contains(lowerLine, "cloudinary.com") || strings.Contains(lowerLine, "unsplash.com") ||
				strings.Contains(lowerLine, "files.oaiusercontent.com") {
				urls = append(urls, line)
				log.Printf("画像ホスティングサービスURLを発見: %s", line)
			}
		}
	}

	log.Printf("合計 %d 個の画像URLを発見", len(urls))
	return urls
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
