package config

import "geminibot/internal/domain"

// ImageGenerationDefaults は画像生成リクエスト用のデフォルトオプションを返します（g が nil の場合はゼロ値）。
func (g *GeminiConfig) ImageGenerationDefaults() domain.ImageGenerationOptions {
	if g == nil {
		return domain.ImageGenerationOptions{}
	}
	return domain.ImageGenerationOptions{
		Model:       g.ImageModelName,
		Style:       domain.ImageStyleFromString(g.ImageStyle),
		Quality:     domain.ImageQualityFromString(g.ImageQuality),
		Size:        domain.ImageSizeFromString(g.ImageSize),
		Count:       g.ImageCount,
		MaxTokens:   g.MaxTokens,
		Temperature: g.Temperature,
		TopP:        g.TopP,
		TopK:        g.TopK,
	}
}
