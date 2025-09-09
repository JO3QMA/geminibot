package application

import (
	"context"
	"fmt"
	"log"

	"geminibot/internal/domain"
)

// ImageGenerationService は、画像生成に関するビジネスロジックを担当するサービスです
type ImageGenerationService struct {
	geminiClient GeminiClient
}

// NewImageGenerationService は新しいImageGenerationServiceインスタンスを作成します
func NewImageGenerationService(geminiClient GeminiClient) *ImageGenerationService {
	return &ImageGenerationService{
		geminiClient: geminiClient,
	}
}

// GenerateImage は、プロンプトから画像を生成します
func (s *ImageGenerationService) GenerateImage(ctx context.Context, prompt string, options ...domain.ImageGenerationOptions) (*domain.ImageGenerationResponse, error) {
	log.Printf("画像生成サービス: プロンプト=%s", prompt)

	// プロンプトの検証
	if err := s.validatePrompt(prompt); err != nil {
		return nil, fmt.Errorf("プロンプトの検証に失敗: %w", err)
	}

	// 画像生成
	response, err := s.geminiClient.GenerateImage(ctx, prompt, options...)
	if err != nil {
		return nil, fmt.Errorf("画像生成に失敗: %w", err)
	}

	log.Printf("画像生成サービス: 生成完了, 画像数=%d", len(response.Images))
	return response, nil
}

// validatePrompt は、プロンプトの妥当性を検証します
func (s *ImageGenerationService) validatePrompt(prompt string) error {
	if prompt == "" {
		return fmt.Errorf("プロンプトが空です")
	}

	if len(prompt) > 1000 {
		return fmt.Errorf("プロンプトが長すぎます (最大1000文字)")
	}

	if len(prompt) < 3 {
		return fmt.Errorf("プロンプトが短すぎます (最小3文字)")
	}

	return nil
}

// GetSupportedStyles は、サポートされているスタイルのリストを返します
func (s *ImageGenerationService) GetSupportedStyles() []string {
	return []string{
		"photographic",
		"anime",
		"illustration",
		"oil_painting",
		"watercolor",
		"digital_art",
		"sketch",
		"cartoon",
	}
}

// GetSupportedQualities は、サポートされている品質のリストを返します
func (s *ImageGenerationService) GetSupportedQualities() []string {
	return []string{
		"standard",
		"high",
	}
}

// GetSupportedSizes は、サポートされているサイズのリストを返します
func (s *ImageGenerationService) GetSupportedSizes() []string {
	return []string{
		"512x512",
		"1024x1024",
		"1024x768",
		"768x1024",
	}
}
