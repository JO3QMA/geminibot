package gemini

import (
	"context"
	"fmt"
	"log"

	"geminibot/internal/application"
	"geminibot/internal/domain"

	"google.golang.org/genai"
)

// GeminiAPIClient は、Gemini APIとの通信を行うクライアントです
type GeminiAPIClient struct {
	client *genai.Client
	config *Config
}

// Config は、Gemini APIクライアントの設定を定義します
type Config struct {
	APIKey      string
	ModelName   string
	MaxTokens   int32
	Temperature float32
	TopP        float32
	TopK        int32
}

// DefaultConfig は、デフォルトの設定を返します
func DefaultConfig() *Config {
	return &Config{
		ModelName:   "gemini-1.5-flash",
		MaxTokens:   1000,
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
	}
}

// NewGeminiAPIClient は新しいGeminiAPIClientインスタンスを作成します
func NewGeminiAPIClient(apiKey string, config *Config) (*GeminiAPIClient, error) {
	if config == nil {
		config = DefaultConfig()
	}

	ctx := context.Background()
	clientConfig := &genai.ClientConfig{
		APIKey: apiKey,
	}

	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("Gemini APIクライアントの作成に失敗: %w", err)
	}

	return &GeminiAPIClient{
		client: client,
		config: config,
	}, nil
}

// GenerateText は、プロンプトを受け取ってGemini APIからテキストを生成します
func (g *GeminiAPIClient) GenerateText(ctx context.Context, prompt domain.Prompt) (string, error) {
	log.Printf("Gemini APIにテキスト生成をリクエスト中: %d文字", len(prompt.Content()))

	// 新しいGemini APIライブラリの仕様に合わせて実装
	contents := genai.Text(prompt.Content())

	// 生成設定を作成
	config := &genai.GenerateContentConfig{
		MaxOutputTokens: g.config.MaxTokens,
		Temperature:     &g.config.Temperature,
		TopP:            &g.config.TopP,
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.config.ModelName, contents, config)
	if err != nil {
		return "", fmt.Errorf("Gemini APIからの応答取得に失敗: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("Gemini APIから有効な応答が得られませんでした")
	}

	candidate := resp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("Gemini APIの応答にコンテンツが含まれていません")
	}

	// テキスト部分を抽出
	var result string
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			result += part.Text
		}
	}

	log.Printf("Gemini APIから応答を取得: %d文字", len(result))
	return result, nil
}

// GenerateTextWithOptions は、オプション付きでテキストを生成します
func (g *GeminiAPIClient) GenerateTextWithOptions(ctx context.Context, prompt domain.Prompt, options application.TextGenerationOptions) (string, error) {
	log.Printf("Gemini APIにオプション付きテキスト生成をリクエスト中: %d文字", len(prompt.Content()))

	// 新しいGemini APIライブラリの仕様に合わせて実装
	contents := genai.Text(prompt.Content())

	// オプションを適用
	maxTokens := g.config.MaxTokens
	if options.MaxTokens > 0 {
		maxTokens = int32(options.MaxTokens)
	}

	temperature := g.config.Temperature
	if options.Temperature > 0 {
		temperature = float32(options.Temperature)
	}

	topP := g.config.TopP
	if options.TopP > 0 {
		topP = float32(options.TopP)
	}

	// 生成設定を作成
	config := &genai.GenerateContentConfig{
		MaxOutputTokens: maxTokens,
		Temperature:     &temperature,
		TopP:            &topP,
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.config.ModelName, contents, config)
	if err != nil {
		return "", fmt.Errorf("Gemini APIからの応答取得に失敗: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("Gemini APIから有効な応答が得られませんでした")
	}

	candidate := resp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("Gemini APIの応答にコンテンツが含まれていません")
	}

	// テキスト部分を抽出
	var result string
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			result += part.Text
		}
	}

	log.Printf("Gemini APIから応答を取得: %d文字", len(result))
	return result, nil
}

// Close は、クライアントのリソースを解放します
func (g *GeminiAPIClient) Close() error {
	// 新しいGemini APIライブラリではCloseメソッドが不要
	return nil
}
