package application

import (
	"context"
	"geminibot/internal/domain"
)

// GeminiClient は、Gemini APIとの通信を行うクライアントのインターフェースです
type GeminiClient interface {
	// GenerateText は、プロンプトを受け取ってGemini APIからテキストを生成します
	GenerateText(ctx context.Context, prompt domain.Prompt) (string, error)

	// GenerateTextWithOptions は、オプション付きでテキストを生成します
	GenerateTextWithOptions(ctx context.Context, prompt domain.Prompt, options TextGenerationOptions) (string, error)

	// GenerateTextWithStructuredContext は、構造化されたコンテキストを使用してテキストを生成します
	GenerateTextWithStructuredContext(ctx context.Context, systemPrompt string, conversationHistory []domain.Message, userQuestion string) (string, error)

	// GenerateImage は、プロンプトを受け取ってGemini APIから画像を生成します
	GenerateImage(ctx context.Context, prompt string) (*domain.ImageGenerationResponse, error)

	// GenerateImageWithOptions は、オプション付きで画像を生成します
	GenerateImageWithOptions(ctx context.Context, prompt string, options domain.ImageGenerationOptions) (*domain.ImageGenerationResponse, error)
}

// TextGenerationOptions は、テキスト生成時のオプションを定義します
type TextGenerationOptions struct {
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
	Model       string  `json:"model,omitempty"`
}

// DefaultTextGenerationOptions は、デフォルトのテキスト生成オプションを返します
func DefaultTextGenerationOptions() TextGenerationOptions {
	return TextGenerationOptions{
		MaxTokens:   1000,
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
		Model:       "gemini-2.5-pro",
	}
}
