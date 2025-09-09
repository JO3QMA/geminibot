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
func (s *ImageGenerationService) GenerateImage(ctx context.Context, request domain.ImageGenerationRequest) (*domain.ImageGenerationResponse, error) {
	log.Printf("画像生成サービス: プロンプト=%s", request.Prompt)

	// プロンプトの検証
	if err := s.validatePrompt(request); err != nil {
		return nil, fmt.Errorf("プロンプトの検証に失敗: %w", err)
	}

	// 画像生成
	response, err := s.geminiClient.GenerateImage(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("画像生成に失敗: %w", err)
	}

	log.Printf("画像生成サービス: 生成完了, 画像数=%d", len(response.Images))
	return response, nil
}

// validatePrompt は、プロンプトの妥当性を検証します
func (s *ImageGenerationService) validatePrompt(request domain.ImageGenerationRequest) error {
	if request.Prompt == "" {
		return fmt.Errorf("プロンプトが空です")
	}

	if len(request.Prompt) > 1000 {
		return fmt.Errorf("プロンプトが長すぎます (最大1000文字)")
	}

	if len(request.Prompt) < 3 {
		return fmt.Errorf("プロンプトが短すぎます (最小3文字)")
	}

	return nil
}

// GetSupportedStyles は、サポートされているスタイルのリストを返します
func (s *ImageGenerationService) GetSupportedStyles() []string {
	styles := domain.AllImageStyles()
	result := make([]string, len(styles))
	for i, style := range styles {
		result[i] = style.String()
	}
	return result
}

// GetSupportedQualities は、サポートされている品質のリストを返します
func (s *ImageGenerationService) GetSupportedQualities() []string {
	qualities := domain.AllImageQualities()
	result := make([]string, len(qualities))
	for i, quality := range qualities {
		result[i] = quality.String()
	}
	return result
}

// GetSupportedSizes は、サポートされているサイズのリストを返します
func (s *ImageGenerationService) GetSupportedSizes() []string {
	sizes := domain.AllImageSizes()
	result := make([]string, len(sizes))
	for i, size := range sizes {
		result[i] = size.String()
	}
	return result
}
