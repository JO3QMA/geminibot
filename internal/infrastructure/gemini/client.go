package gemini

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

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
		return nil, fmt.Errorf("GeminiConfigが指定されていません")
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

// createSafetySettings は、安全フィルター設定を作成します
func (g *GeminiAPIClient) createSafetySettings() []*genai.SafetySetting {
	return []*genai.SafetySetting{
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
	}
}

// createGenerateConfig は、生成設定を作成します
func (g *GeminiAPIClient) createGenerateConfig() *genai.GenerateContentConfig {
	return &genai.GenerateContentConfig{
		MaxOutputTokens: g.config.MaxTokens,
		Temperature:     &g.config.Temperature,
		TopP:            &g.config.TopP,
		SafetySettings:  g.createSafetySettings(),
	}
}

// createGenerateConfigWithOptions は、オプション付きで生成設定を作成します
func (g *GeminiAPIClient) createGenerateConfigWithOptions(options application.TextGenerationOptions) *genai.GenerateContentConfig {
	temp := float32(options.Temperature)
	topP := float32(options.TopP)
	return &genai.GenerateContentConfig{
		MaxOutputTokens: int32(options.MaxTokens),
		Temperature:     &temp,
		TopP:            &topP,
		SafetySettings:  g.createSafetySettings(),
	}
}

// handleAPIError は、APIエラーを統一して処理します
func (g *GeminiAPIClient) handleAPIError(err error, ctx context.Context) error {
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("Gemini APIへのリクエストがタイムアウトしました。時間を置いて再度お試しください: %w", err)
	}

	// エラーメッセージをより詳細に
	errStr := err.Error()
	if strings.Contains(errStr, "quota") || strings.Contains(errStr, "limit") {
		return fmt.Errorf("Gemini APIの利用制限に達しました。しばらく時間を置いてから再度お試しください: %w", err)
	}
	if strings.Contains(errStr, "permission") || strings.Contains(errStr, "unauthorized") {
		return fmt.Errorf("Gemini APIへのアクセス権限がありません。APIキーを確認してください: %w", err)
	}
	if strings.Contains(errStr, "network") || strings.Contains(errStr, "connection") {
		return fmt.Errorf("ネットワークエラーが発生しました。接続を確認して再度お試しください: %w", err)
	}

	return fmt.Errorf("Gemini APIからの応答取得に失敗しました: %w", err)
}

// logRequestDetails は、リクエスト詳細をログ出力します
func (g *GeminiAPIClient) logRequestDetails(promptLength int, promptContent string) {
	log.Printf("Gemini APIにテキスト生成をリクエスト中: %d文字", promptLength)
	log.Printf("プロンプト内容: %s", promptContent)
}

// logResponseDetails は、レスポンス詳細をログ出力します
func (g *GeminiAPIClient) logResponseDetails(resp *genai.GenerateContentResponse) {
	log.Printf("Gemini APIレスポンス: Candidates数=%d", len(resp.Candidates))
	if len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]
		if candidate.Content != nil {
			log.Printf("Candidate詳細: FinishReason=%s, Parts数=%d", candidate.FinishReason, len(candidate.Content.Parts))
		} else {
			log.Printf("Candidate詳細: FinishReason=%s, Content=nil", candidate.FinishReason)
		}

		if len(candidate.SafetyRatings) > 0 {
			for i, rating := range candidate.SafetyRatings {
				log.Printf("SafetyRating[%d]: Category=%s, Probability=%s", i, rating.Category, rating.Probability)
			}
		}
	}
}

// shouldRetry は、エラーがリトライ可能かどうかを判定します
func (g *GeminiAPIClient) shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Contentがnilの場合やコンテンツが含まれていない場合はリトライ対象
	return strings.Contains(errStr, "Contentが含まれていません") ||
		strings.Contains(errStr, "コンテンツが含まれていません")
}

// retryWithBackoff は、指数バックオフでリトライを実行します
func (g *GeminiAPIClient) retryWithBackoff(ctx context.Context, operation func() (string, error)) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= g.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// 指数バックオフ: 1秒、2秒、4秒...
			backoffDuration := time.Duration(1<<uint(attempt-1)) * time.Second
			log.Printf("リトライ %d/%d 回目: %v 後に再試行します", attempt, g.config.MaxRetries, backoffDuration)

			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoffDuration):
			}
		}

		result, err := operation()
		if err == nil {
			if attempt > 0 {
				log.Printf("リトライ成功: %d回目の試行で成功しました", attempt+1)
			}
			return result, nil
		}

		lastErr = err

		// リトライ可能なエラーかチェック
		if !g.shouldRetry(err) {
			log.Printf("リトライ不可能なエラー: %v", err)
			return "", err
		}

		if attempt < g.config.MaxRetries {
			log.Printf("リトライ可能なエラーが発生: %v", err)
		}
	}

	return "", fmt.Errorf("最大リトライ回数 (%d) に達しました。最後のエラー: %w", g.config.MaxRetries, lastErr)
}

// GenerateText は、プロンプトを受け取ってGemini APIからテキストを生成します
func (g *GeminiAPIClient) GenerateText(ctx context.Context, prompt domain.Prompt) (string, error) {
	g.logRequestDetails(len(prompt.Content), prompt.Content)

	// リトライ機能付きでテキスト生成を実行
	return g.retryWithBackoff(ctx, func() (string, error) {
		// 新しいGemini APIライブラリの仕様に合わせて実装
		contents := genai.Text(prompt.Content)

		// 生成設定を作成
		config := g.createGenerateConfig()

		resp, err := g.client.Models.GenerateContent(ctx, g.config.ModelName, contents, config)
		if err != nil {
			return "", g.handleAPIError(err, ctx)
		}

		// レスポンス詳細をログ出力
		g.logResponseDetails(resp)

		// 統一されたレスポンス処理を使用
		return g.processResponse(resp)
	})
}

// GenerateTextWithOptions は、オプション付きでテキストを生成します
func (g *GeminiAPIClient) GenerateTextWithOptions(ctx context.Context, prompt domain.Prompt, options application.TextGenerationOptions) (string, error) {
	g.logRequestDetails(len(prompt.Content), prompt.Content)

	// リトライ機能付きでテキスト生成を実行
	return g.retryWithBackoff(ctx, func() (string, error) {
		// 新しいGemini APIライブラリの仕様に合わせて実装
		contents := genai.Text(prompt.Content)

		// オプションに基づいて生成設定を作成
		config := g.createGenerateConfigWithOptions(options)

		// モデル名を決定（オプションで指定されていない場合はデフォルトを使用）
		modelName := g.config.ModelName
		if options.Model != "" {
			modelName = options.Model
		}

		resp, err := g.client.Models.GenerateContent(ctx, modelName, contents, config)
		if err != nil {
			return "", g.handleAPIError(err, ctx)
		}

		// レスポンス詳細をログ出力
		g.logResponseDetails(resp)

		// レスポンス処理
		return g.processResponse(resp)
	})
}

// GenerateTextWithStructuredContext は、構造化されたコンテキストを使用してテキストを生成します
func (g *GeminiAPIClient) GenerateTextWithStructuredContext(ctx context.Context, systemPrompt string, conversationHistory []domain.Message, userQuestion string) (string, error) {
	// 統一されたログ出力メソッドを使用
	g.logRequestDetails(len(userQuestion), userQuestion)
	log.Printf("構造化コンテキストでGemini APIにテキスト生成をリクエスト中")
	log.Printf("システムプロンプト: %d文字", len(systemPrompt))
	log.Printf("会話履歴: %d件", len(conversationHistory))

	// リトライ機能付きでテキスト生成を実行
	return g.retryWithBackoff(ctx, func() (string, error) {
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
			return "", g.handleAPIError(err, ctx)
		}

		// レスポンス詳細をログ出力
		g.logResponseDetails(resp)

		// レスポンス処理
		return g.processResponse(resp)
	})
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

// formatSafetyRatings は、SafetyRatingsの詳細情報をフォーマットします
func (g *GeminiAPIClient) formatSafetyRatings(ratings []*genai.SafetyRating) string {
	if len(ratings) == 0 {
		return "詳細情報なし"
	}

	var details []string
	for _, rating := range ratings {
		if rating != nil {
			category := g.translateSafetyCategory(rating.Category)
			probability := g.translateSafetyProbability(rating.Probability)
			details = append(details, fmt.Sprintf("%s: %s", category, probability))
		}
	}

	if len(details) == 0 {
		return "詳細情報なし"
	}

	return strings.Join(details, ", ")
}

// translateSafetyCategory は、SafetyCategoryを日本語に翻訳します
func (g *GeminiAPIClient) translateSafetyCategory(category genai.HarmCategory) string {
	switch category {
	case genai.HarmCategoryHarassment:
		return "ハラスメント"
	case genai.HarmCategoryHateSpeech:
		return "ヘイトスピーチ"
	case genai.HarmCategorySexuallyExplicit:
		return "性的表現"
	case genai.HarmCategoryDangerousContent:
		return "危険なコンテンツ"
	default:
		return string(category)
	}
}

// translateSafetyProbability は、SafetyProbabilityを日本語に翻訳します
func (g *GeminiAPIClient) translateSafetyProbability(probability genai.HarmProbability) string {
	switch probability {
	case genai.HarmProbabilityNegligible:
		return "無視できるレベル"
	case genai.HarmProbabilityLow:
		return "低レベル"
	case genai.HarmProbabilityMedium:
		return "中レベル"
	case genai.HarmProbabilityHigh:
		return "高レベル"
	default:
		return string(probability)
	}
}

// processResponse は、Gemini APIのレスポンスを処理します
func (g *GeminiAPIClient) processResponse(resp *genai.GenerateContentResponse) (string, error) {
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("Gemini APIから有効な応答が得られませんでした")
	}

	candidate := resp.Candidates[0]

	// FinishReasonをチェックして安全フィルターによるブロックを検出
	if candidate.FinishReason == "SAFETY" {
		safetyDetails := g.formatSafetyRatings(candidate.SafetyRatings)
		return "", fmt.Errorf("Gemini APIの安全フィルターによって応答がブロックされました。詳細: %s", safetyDetails)
	}

	if candidate.FinishReason == "RECITATION" {
		return "", fmt.Errorf("Gemini APIが著作権保護された内容を検出しました。著作権で保護されたコンテンツが含まれている可能性があります")
	}

	if candidate.FinishReason == "MAX_TOKENS" {
		return "", fmt.Errorf("Gemini APIの応答が最大トークン数に達しました。より短い質問を試してください")
	}

	if candidate.FinishReason == "STOP" {
		// STOPは正常な終了なので、そのまま処理を続行
	} else if candidate.FinishReason != "" && candidate.FinishReason != "STOP" {
		return "", fmt.Errorf("Gemini APIで予期しない終了理由が発生しました: %s", candidate.FinishReason)
	}

	// Contentがnilの場合のチェックを追加
	if candidate.Content == nil {
		return "", fmt.Errorf("Gemini APIの応答にContentが含まれていません。FinishReason: %s", candidate.FinishReason)
	}

	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("Gemini APIの応答にコンテンツが含まれていません。FinishReason: %s", candidate.FinishReason)
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

// GenerateImage は、プロンプトを受け取ってGemini APIから画像を生成します
// optionsが空の場合はデフォルト設定を使用します
func (g *GeminiAPIClient) GenerateImage(ctx context.Context, request domain.ImageGenerationRequest) (*domain.ImageGenerationResponse, error) {
	log.Printf("Gemini APIに画像生成をリクエスト中: %d文字", len(request.Prompt))
	log.Printf("プロンプト内容: %s", request.Prompt)
	log.Printf("オプション: %+v", request.Options)

	// リトライ機能付きで画像生成を実行
	return g.retryWithBackoffForImage(ctx, func() (*domain.ImageGenerationResponse, error) {
		// 画像生成用のコンテンツを作成
		contents := genai.Text(request.Prompt)

		// オプションに基づいて画像生成設定を作成
		config := g.createImageConfig(request.Options)

		// モデル名を決定
		modelName := request.Options.Model
		if g.config.ModelName != "" {
			modelName = g.config.ModelName
		}

		resp, err := g.client.Models.GenerateContent(ctx, modelName, contents, config)
		if err != nil {
			return nil, g.handleAPIError(err, ctx)
		}

		// レスポンス詳細をログ出力
		g.logResponseDetails(resp)

		// 画像生成結果を処理
		return g.processImageResponse(resp, request.Prompt, modelName)
	})
}
