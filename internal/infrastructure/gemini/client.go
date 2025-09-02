package gemini

import (
	"context"
	"fmt"
	"log"
	"strings"

	"geminibot/internal/application"
	"geminibot/internal/domain"
	"geminibot/internal/infrastructure/config"

	"google.golang.org/genai"
)

// GeminiAPIClient は、Gemini APIとの通信を行うクライアントです
type GeminiAPIClient struct {
	client *genai.Client
	config *config.GeminiConfig
}

// NewGeminiAPIClient は新しいGeminiAPIClientインスタンスを作成します
func NewGeminiAPIClient(apiKey string, geminiConfig *config.GeminiConfig) (*GeminiAPIClient, error) {
	if geminiConfig == nil {
		geminiConfig = config.DefaultGeminiConfig()
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
		config: geminiConfig,
	}, nil
}

// createGenerateConfig は、生成設定を作成します
func (g *GeminiAPIClient) createGenerateConfig() *genai.GenerateContentConfig {
	return &genai.GenerateContentConfig{
		MaxOutputTokens: g.config.MaxTokens,
		Temperature:     &g.config.Temperature,
		TopP:            &g.config.TopP,
		// 安全フィルターの設定を調整（中程度の制限）
		SafetySettings: []*genai.SafetySetting{
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
		},
	}
}

// GenerateText は、プロンプトを受け取ってGemini APIからテキストを生成します
func (g *GeminiAPIClient) GenerateText(ctx context.Context, prompt domain.Prompt) (string, error) {
	log.Printf("Gemini APIにテキスト生成をリクエスト中: %d文字", len(prompt.Content()))
	log.Printf("プロンプト内容: %s", prompt.Content())

	// 新しいGemini APIライブラリの仕様に合わせて実装
	contents := genai.Text(prompt.Content())

	// 生成設定を作成
	config := g.createGenerateConfig()

	resp, err := g.client.Models.GenerateContent(ctx, g.config.ModelName, contents, config)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("Gemini APIへのリクエストがタイムアウトしました: %w", err)
		}
		return "", fmt.Errorf("Gemini APIからの応答取得に失敗: %w", err)
	}

	// デバッグ用：レスポンスの詳細をログ出力
	log.Printf("Gemini APIレスポンス: Candidates数=%d", len(resp.Candidates))
	if len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]
		log.Printf("Candidate詳細: FinishReason=%s, Parts数=%d", candidate.FinishReason, len(candidate.Content.Parts))

		// SafetyRatingsがある場合はログ出力
		if len(candidate.SafetyRatings) > 0 {
			for i, rating := range candidate.SafetyRatings {
				log.Printf("SafetyRating[%d]: Category=%s, Probability=%s", i, rating.Category, rating.Probability)
			}
		}
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("Gemini APIから有効な応答が得られませんでした")
	}

	candidate := resp.Candidates[0]

	// FinishReasonをチェックして安全フィルターによるブロックを検出
	if candidate.FinishReason == "SAFETY" {
		return "", fmt.Errorf("Gemini APIの安全フィルターによって応答がブロックされました")
	}

	if candidate.FinishReason == "RECITATION" {
		return "", fmt.Errorf("Gemini APIが著作権保護された内容を検出しました")
	}

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
	log.Printf("オプション付きでGemini APIにテキスト生成をリクエスト中: %d文字", len(prompt.Content()))

	// 新しいGemini APIライブラリの仕様に合わせて実装
	contents := genai.Text(prompt.Content())

	// オプションに基づいて生成設定を作成
	temp := float32(options.Temperature)
	topP := float32(options.TopP)
	config := &genai.GenerateContentConfig{
		MaxOutputTokens: int32(options.MaxTokens),
		Temperature:     &temp,
		TopP:            &topP,
		// 安全フィルターの設定を調整（中程度の制限）
		SafetySettings: []*genai.SafetySetting{
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
		},
	}

	// モデル名を決定（オプションで指定されていない場合はデフォルトを使用）
	modelName := g.config.ModelName
	if options.Model != "" {
		modelName = options.Model
	}

	resp, err := g.client.Models.GenerateContent(ctx, modelName, contents, config)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("Gemini APIへのリクエストがタイムアウトしました: %w", err)
		}
		return "", fmt.Errorf("Gemini APIからの応答取得に失敗: %w", err)
	}

	// レスポンス処理
	return g.processResponse(resp)
}

// GenerateTextWithStructuredContext は、構造化されたコンテキストを使用してテキストを生成します
func (g *GeminiAPIClient) GenerateTextWithStructuredContext(ctx context.Context, systemPrompt string, conversationHistory []domain.Message, userQuestion string) (string, error) {
	log.Printf("構造化コンテキストでGemini APIにテキスト生成をリクエスト中")
	log.Printf("システムプロンプト: %d文字", len(systemPrompt))
	log.Printf("会話履歴: %d件", len(conversationHistory))
	log.Printf("ユーザー質問: %d文字", len(userQuestion))

	// 構造化されたコンテンツを作成
	var allContents []*genai.Content

	// システムプロンプトを追加
	allContents = append(allContents, genai.Text(systemPrompt)...)

	// ユーザーの質問を最初に追加（最優先）
	userQuestionText := fmt.Sprintf("## ユーザーの現在の質問\n%s", userQuestion)
	allContents = append(allContents, genai.Text(userQuestionText)...)

	// 会話履歴を最後に追加（参考情報として）
	if len(conversationHistory) > 0 {
		historyText := g.formatConversationHistory(conversationHistory)
		allContents = append(allContents, genai.Text(historyText)...)
	}

	// 生成設定を作成
	config := g.createGenerateConfig()

	resp, err := g.client.Models.GenerateContent(ctx, g.config.ModelName, allContents, config)
	if err != nil {
		return "", fmt.Errorf("Gemini APIからの応答取得に失敗: %w", err)
	}

	// レスポンス処理
	return g.processResponse(resp)
}

// formatConversationHistory は、会話履歴を構造化された形式にフォーマットします
func (g *GeminiAPIClient) formatConversationHistory(messages []domain.Message) string {
	var builder strings.Builder
	builder.WriteString("## 参考情報：過去の会話履歴\n")
	builder.WriteString("※ 以下の会話履歴は参考情報です。ユーザーの現在の質問に直接答えてください。\n\n")

	for _, msg := range messages {
		displayName := msg.User.DisplayName
		builder.WriteString(fmt.Sprintf("%s: %s\n", displayName, msg.Content))
	}

	return builder.String()
}

// processResponse は、Gemini APIのレスポンスを処理します
func (g *GeminiAPIClient) processResponse(resp *genai.GenerateContentResponse) (string, error) {
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("Gemini APIから有効な応答が得られませんでした")
	}

	candidate := resp.Candidates[0]

	// FinishReasonをチェックして安全フィルターによるブロックを検出
	if candidate.FinishReason == "SAFETY" {
		return "", fmt.Errorf("Gemini APIの安全フィルターによって応答がブロックされました")
	}

	if candidate.FinishReason == "RECITATION" {
		return "", fmt.Errorf("Gemini APIが著作権保護された内容を検出しました")
	}

	// Contentがnilの場合のチェックを追加
	if candidate.Content == nil {
		return "", fmt.Errorf("Gemini APIの応答にContentが含まれていません")
	}

	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("Gemini APIの応答にコンテンツが含まれていません")
	}

	// テキスト部分を抽出
	var result string
	for _, part := range candidate.Content.Parts {
		if part != nil && part.Text != "" {
			result += part.Text
		}
	}

	log.Printf("Gemini APIから応答を取得: %d文字", len(result))
	return result, nil
}

// Close は、Gemini APIクライアントを閉じます
func (g *GeminiAPIClient) Close() error {
	// genai.ClientにはCloseメソッドがないため、何もしない
	return nil
}
