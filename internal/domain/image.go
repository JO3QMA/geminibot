package domain

// ImageGenerationResult は、画像生成の結果を表すドメインオブジェクトです
type ImageGenerationResult struct {
	ImageURL    string `json:"image_url"`
	Prompt      string `json:"prompt"`
	Model       string `json:"model"`
	GeneratedAt string `json:"generated_at"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

// ImageGenerationOptions は、画像生成時のオプションを定義します
type ImageGenerationOptions struct {
	Model       string  `json:"model,omitempty"`
	Quality     string  `json:"quality,omitempty"`
	Style       string  `json:"style,omitempty"`
	MaxTokens   int32   `json:"max_tokens,omitempty"`
	Temperature float32 `json:"temperature,omitempty"`
	TopP        float32 `json:"top_p,omitempty"`
	TopK        int32   `json:"top_k,omitempty"`
}

// DefaultImageGenerationOptions は、デフォルトの画像生成オプションを返します
func DefaultImageGenerationOptions() ImageGenerationOptions {
	return ImageGenerationOptions{
		Model:       "gemini-2.5-flash-image",
		Quality:     "standard",
		Style:       "natural",
		MaxTokens:   1000,
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
	}
}

// ImagePrompt は、画像生成用のプロンプトを表すドメインオブジェクトです
type ImagePrompt struct {
	Content string `json:"content"`
	Type    string `json:"type"` // "text", "image_edit", "image_generation"
}

// NewImagePrompt は、新しいImagePromptを作成します
func NewImagePrompt(content string) ImagePrompt {
	return ImagePrompt{
		Content: content,
		Type:    "image_generation",
	}
}
