package discord

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"geminibot/internal/domain"

	"github.com/bwmarrin/discordgo"
)

// ResponseHandler は、Discordのレスポンス送信・フォーマット処理を担当するハンドラーです
type ResponseHandler struct{}

// NewResponseHandler は新しいResponseHandlerインスタンスを作成します
func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

// sendNormalReply は、スレッド作成に失敗した場合の通常のリプライ送信を行います
func (h *ResponseHandler) sendNormalReply(s *discordgo.Session, m *discordgo.MessageCreate, mention domain.BotMention, mentionService interface{}) {
	// 処理中メッセージを送信
	thinkingMsg, err := s.ChannelMessageSendReply(m.ChannelID, "🤔 考え中...", &discordgo.MessageReference{
		MessageID: m.ID,
		ChannelID: m.ChannelID,
		GuildID:   m.GuildID,
	})
	if err != nil {
		log.Printf("処理中メッセージの送信に失敗: %v", err)
		return
	}

	// メンションを処理
	ctx := context.Background()
	response, err := h.handleMentionWithService(ctx, mention, mentionService)

	// 処理中メッセージを削除
	s.ChannelMessageDelete(m.ChannelID, thinkingMsg.ID)

	if err != nil {
		log.Printf("メンション処理に失敗: %v", err)

		// エラーを適切なメッセージにフォーマット
		errorMsg := h.formatError(err)
		s.ChannelMessageSendReply(m.ChannelID, errorMsg, &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})
		return
	}

	// 応答を通常のリプライとして送信
	h.sendSplitResponse(s, m, response)
}

// ProcessImageGenerationWithoutThread は、スレッド作成に失敗した場合の画像生成処理を行います
func (h *ResponseHandler) sendImageGenerationNormalReply(s *discordgo.Session, m *discordgo.MessageCreate, mentionService interface{}) {
	// 処理中メッセージを送信
	thinkingMsg, err := s.ChannelMessageSendReply(m.ChannelID, "🎨 画像を生成中...", &discordgo.MessageReference{
		MessageID: m.ID,
		ChannelID: m.ChannelID,
		GuildID:   m.GuildID,
	})
	if err != nil {
		log.Printf("処理中メッセージの送信に失敗: %v", err)
		return
	}

	// 画像生成を処理
	ctx := context.Background()
	imageResult, err := h.generateImageWithService(ctx, m, mentionService)

	// 処理中メッセージを削除
	s.ChannelMessageDelete(m.ChannelID, thinkingMsg.ID)

	if err != nil {
		log.Printf("画像生成に失敗: %v", err)
		errorMsg := h.formatImageGenerationError(err)
		s.ChannelMessageSendReply(m.ChannelID, errorMsg, &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})
		return
	}

	// 画像生成結果を通常のリプライとして送信
	h.sendImageGenerationResultToChannel(s, m, imageResult)
}

// handleMentionWithService は、mentionServiceを使用してメンションを処理します
func (h *ResponseHandler) handleMentionWithService(ctx context.Context, mention domain.BotMention, mentionService interface{}) (string, error) {
	// mentionServiceの型を確認して適切なメソッドを呼び出す
	if service, ok := mentionService.(interface {
		HandleMention(ctx context.Context, mention domain.BotMention) (string, error)
	}); ok {
		return service.HandleMention(ctx, mention)
	}
	return "", fmt.Errorf("mentionServiceがHandleMentionメソッドを実装していません")
}

// generateImageWithService は、mentionServiceを使用して画像生成を実行します
func (h *ResponseHandler) generateImageWithService(ctx context.Context, m *discordgo.MessageCreate, mentionService interface{}) (*domain.ImageGenerationResult, error) {
	// メンション部分を除去したコンテンツを取得
	content := h.extractUserContent(m)

	// 画像生成用のプロンプトを作成
	prompt := domain.NewImagePrompt(content)

	// mentionServiceの型を確認して適切なメソッドを呼び出す
	if service, ok := mentionService.(interface {
		GenerateImage(ctx context.Context, request domain.ImageGenerationRequest) (*domain.ImageGenerationResponse, error)
	}); ok {
		response, err := service.GenerateImage(ctx, domain.ImageGenerationRequest{
			Prompt:  prompt,
			Options: domain.DefaultImageGenerationOptions(),
		})
		if err != nil {
			return &domain.ImageGenerationResult{
				Success: false,
				Error:   err.Error(),
			}, nil
		}

		// ImageGenerationResponseをImageGenerationResultに変換
		result := &domain.ImageGenerationResult{
			Response: response,
			Success:  true,
			Error:    "",
			ImageURL: "", // 必要に応じて設定
		}

		return result, nil
	}
	return &domain.ImageGenerationResult{
		Success: false,
		Error:   "mentionServiceがGenerateImageメソッドを実装していません",
	}, nil
}

// extractUserContent は、メンション部分を除去したユーザーのコンテンツを抽出します
func (h *ResponseHandler) extractUserContent(m *discordgo.MessageCreate) string {
	content := m.Content

	// メンション配列がある場合、それらを除去
	for _, mention := range m.Mentions {
		mentionText := fmt.Sprintf("<@%s>", mention.ID)
		content = strings.ReplaceAll(content, mentionText, "")
	}

	// 先頭と末尾の空白を除去
	content = strings.TrimSpace(content)

	return content
}

// sendThreadResponse は、スレッド内に応答を送信します
func (h *ResponseHandler) sendThreadResponse(s *discordgo.Session, threadID string, response string) {
	// 応答をDiscord用にフォーマット
	formattedResponse := h.formatForDiscord(response)

	// 応答が非常に長い場合はファイルとして送信
	if len(formattedResponse) > DiscordMessageLimit*5 {
		h.sendAsFileToThread(s, threadID, formattedResponse, "response.txt")
		return
	}

	// 応答をDiscordの制限に合わせて分割
	chunks := h.splitMessage(formattedResponse)

	// すべてのチャンクをスレッド内に送信
	for i, chunk := range chunks {
		_, err := s.ChannelMessageSend(threadID, chunk)
		if err != nil {
			log.Printf("スレッド内メッセージの送信に失敗 (チャンク %d): %v", i+1, err)
			break
		}
	}
}

// sendAsFileToThread は、長い応答をファイルとしてスレッド内に送信します
func (h *ResponseHandler) sendAsFileToThread(s *discordgo.Session, threadID string, content, filename string) {
	// ファイルデータを作成
	fileData := strings.NewReader(content)

	// ファイルを添付してメッセージを送信
	_, err := s.ChannelFileSend(threadID, filename, fileData)

	if err != nil {
		log.Printf("ファイル送信に失敗: %v", err)
		// ファイル送信に失敗した場合は通常の分割送信にフォールバック
		h.sendThreadResponse(s, threadID, content)
		return
	}

	// ファイル送信成功のメッセージを送信
	fileMsg := fmt.Sprintf("📄 **応答が長いため、ファイルとして送信しました**\nファイル名: `%s`", filename)
	s.ChannelMessageSend(threadID, fileMsg)
}

// sendSplitResponse は、長い応答を複数のメッセージに分割して送信します
func (h *ResponseHandler) sendSplitResponse(s *discordgo.Session, m *discordgo.MessageCreate, response string) {
	// 応答をDiscord用にフォーマット
	formattedResponse := h.formatForDiscord(response)

	// 応答が非常に長い場合はファイルとして送信
	if len(formattedResponse) > DiscordMessageLimit*5 {
		h.sendAsFile(s, m, formattedResponse, "response.txt")
		return
	}

	// 応答をDiscordの制限に合わせて分割
	chunks := h.splitMessage(formattedResponse)

	if len(chunks) == 1 {
		// 単一メッセージの場合
		_, err := s.ChannelMessageSendReply(m.ChannelID, chunks[0], &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})
		if err != nil {
			log.Printf("応答メッセージの送信に失敗: %v", err)
		}
		return
	}

	// 複数メッセージの場合 - すべてスレッド返信として送信
	for i, chunk := range chunks {
		_, err := s.ChannelMessageSendReply(m.ChannelID, chunk, &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})

		if err != nil {
			log.Printf("応答メッセージの送信に失敗 (チャンク %d): %v", i+1, err)
			break
		}
	}
}

// sendAsFile は、長い応答をファイルとして送信します
func (h *ResponseHandler) sendAsFile(s *discordgo.Session, m *discordgo.MessageCreate, content, filename string) {
	// ファイルデータを作成
	fileData := strings.NewReader(content)

	// ファイルを添付してメッセージを送信
	_, err := s.ChannelFileSend(
		m.ChannelID,
		filename,
		fileData,
	)

	if err != nil {
		log.Printf("ファイル送信に失敗: %v", err)
		// ファイル送信に失敗した場合は通常の分割送信にフォールバック
		h.sendSplitResponse(s, m, content)
		return
	}

	// ファイル送信成功のメッセージをスレッド返信として送信
	fileMsg := fmt.Sprintf("📄 **応答が長いため、ファイルとして送信しました**\nファイル名: `%s`", filename)
	s.ChannelMessageSendReply(m.ChannelID, fileMsg, &discordgo.MessageReference{
		MessageID: m.ID,
		ChannelID: m.ChannelID,
		GuildID:   m.GuildID,
	})
}

// formatForDiscord は、Geminiからの応答をDiscord用にフォーマットします
func (h *ResponseHandler) formatForDiscord(response string) string {
	// markdownのコードブロックをDiscord用に変換
	formatted := h.convertCodeBlocks(response)

	// markdownのインラインコードをDiscord用に変換
	formatted = h.convertInlineCode(formatted)

	// markdownの太字をDiscord用に変換
	formatted = h.convertBold(formatted)

	// markdownの斜体をDiscord用に変換
	formatted = h.convertItalic(formatted)

	// markdownのリストをDiscord用に変換
	formatted = h.convertLists(formatted)

	return formatted
}

// convertCodeBlocks は、markdownのコードブロックをDiscord用に変換します
func (h *ResponseHandler) convertCodeBlocks(text string) string {
	// ```で囲まれたコードブロックを```に変換
	// 言語指定がある場合は除去
	lines := strings.Split(text, "\n")
	var result []string
	inCodeBlock := false
	codeBlockContent := []string{}

	for _, line := range lines {
		if strings.HasPrefix(line, "```") && !inCodeBlock {
			// コードブロック開始
			inCodeBlock = true
			codeBlockContent = []string{}
		} else if strings.HasPrefix(line, "```") && inCodeBlock {
			// コードブロック終了
			inCodeBlock = false
			if len(codeBlockContent) > 0 {
				result = append(result, "```")
				result = append(result, codeBlockContent...)
				result = append(result, "```")
			}
		} else if inCodeBlock {
			// コードブロック内の内容
			codeBlockContent = append(codeBlockContent, line)
		} else {
			// 通常の行
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// convertInlineCode は、markdownのインラインコードをDiscord用に変換します
func (h *ResponseHandler) convertInlineCode(text string) string {
	// `で囲まれたインラインコードを`に変換
	// ただし、コードブロック内は除外
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			// コードブロックの境界はそのまま
			result = append(result, line)
		} else {
			// インラインコードを変換
			converted := h.convertInlineCodeInLine(line)
			result = append(result, converted)
		}
	}

	return strings.Join(result, "\n")
}

// convertInlineCodeInLine は、1行内のインラインコードを変換します
func (h *ResponseHandler) convertInlineCodeInLine(line string) string {
	// バッククォートのペアを`に変換
	// ただし、コードブロック内は除外
	var result strings.Builder
	inInlineCode := false
	codeContent := strings.Builder{}

	for i := 0; i < len(line); i++ {
		if line[i] == '`' && !inInlineCode {
			// インラインコード開始
			inInlineCode = true
			codeContent.Reset()
		} else if line[i] == '`' && inInlineCode {
			// インラインコード終了
			inInlineCode = false
			result.WriteString("`")
			result.WriteString(codeContent.String())
			result.WriteString("`")
		} else if inInlineCode {
			// インラインコード内の内容
			codeContent.WriteByte(line[i])
		} else {
			// 通常の文字
			result.WriteByte(line[i])
		}
	}

	return result.String()
}

// convertBold は、markdownの太字をDiscord用に変換します
func (h *ResponseHandler) convertBold(text string) string {
	// **で囲まれた太字を**に変換
	// ただし、コードブロック内は除外
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			// コードブロックの境界はそのまま
			result = append(result, line)
		} else {
			// 太字を変換
			converted := h.convertBoldInLine(line)
			result = append(result, converted)
		}
	}

	return strings.Join(result, "\n")
}

// convertBoldInLine は、1行内の太字を変換します
func (h *ResponseHandler) convertBoldInLine(line string) string {
	// **で囲まれた太字を**に変換
	// ただし、インラインコード内は除外
	var result strings.Builder
	inInlineCode := false
	inBold := false
	boldContent := strings.Builder{}

	for i := 0; i < len(line); i++ {
		if line[i] == '`' {
			// インラインコードの境界
			if inBold {
				// 太字を終了してからインラインコードを処理
				inBold = false
				result.WriteString("**")
				result.WriteString(boldContent.String())
				result.WriteString("**")
				boldContent.Reset()
			}
			inInlineCode = !inInlineCode
			result.WriteByte(line[i])
		} else if !inInlineCode && i+1 < len(line) && line[i] == '*' && line[i+1] == '*' {
			// **の検出
			if !inBold {
				// 太字開始
				inBold = true
				boldContent.Reset()
			} else {
				// 太字終了
				inBold = false
				result.WriteString("**")
				result.WriteString(boldContent.String())
				result.WriteString("**")
				boldContent.Reset()
			}
			i++ // 次の*をスキップ
		} else if inBold {
			// 太字内の内容
			boldContent.WriteByte(line[i])
		} else {
			// 通常の文字
			result.WriteByte(line[i])
		}
	}

	// 未終了の太字があれば終了
	if inBold {
		result.WriteString("**")
		result.WriteString(boldContent.String())
		result.WriteString("**")
	}

	return result.String()
}

// convertItalic は、markdownの斜体をDiscord用に変換します
func (h *ResponseHandler) convertItalic(text string) string {
	// *で囲まれた斜体を*に変換（ただし、太字の**は除外）
	// ただし、コードブロック内は除外
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			// コードブロックの境界はそのまま
			result = append(result, line)
		} else {
			// 斜体を変換
			converted := h.convertItalicInLine(line)
			result = append(result, converted)
		}
	}

	return strings.Join(result, "\n")
}

// convertItalicInLine は、1行内の斜体を変換します
func (h *ResponseHandler) convertItalicInLine(line string) string {
	// *で囲まれた斜体を*に変換（ただし、太字の**は除外）
	// ただし、インラインコード内は除外
	var result strings.Builder
	inInlineCode := false
	inItalic := false
	italicContent := strings.Builder{}

	for i := 0; i < len(line); i++ {
		if line[i] == '`' {
			// インラインコードの境界
			if inItalic {
				// 斜体を終了してからインラインコードを処理
				inItalic = false
				result.WriteString("*")
				result.WriteString(italicContent.String())
				result.WriteString("*")
				italicContent.Reset()
			}
			inInlineCode = !inInlineCode
			result.WriteByte(line[i])
		} else if !inInlineCode && line[i] == '*' {
			// *の検出
			if i+1 < len(line) && line[i+1] == '*' {
				// **の場合は太字なのでスキップ
				result.WriteString("**")
				i++
			} else if !inItalic {
				// 斜体開始
				inItalic = true
				italicContent.Reset()
			} else {
				// 斜体終了
				inItalic = false
				result.WriteString("*")
				result.WriteString(italicContent.String())
				result.WriteString("*")
				italicContent.Reset()
			}
		} else if inItalic {
			// 斜体内の内容
			italicContent.WriteByte(line[i])
		} else {
			// 通常の文字
			result.WriteByte(line[i])
		}
	}

	// 未終了の斜体があれば終了
	if inItalic {
		result.WriteString("*")
		result.WriteString(italicContent.String())
		result.WriteString("*")
	}

	return result.String()
}

// convertLists は、markdownのリストをDiscord用に変換します
func (h *ResponseHandler) convertLists(text string) string {
	// リストの変換（基本的にはそのまま、必要に応じて調整）
	// Discordは基本的なリスト表示をサポートしているので、
	// 主に番号付きリストの形式を調整
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			// コードブロックの境界はそのまま
			result = append(result, line)
		} else {
			// リストを変換
			converted := h.convertListInLine(line)
			result = append(result, converted)
		}
	}

	return strings.Join(result, "\n")
}

// convertListInLine は、1行内のリストを変換します
func (h *ResponseHandler) convertListInLine(line string) string {
	// 番号付きリストの形式を調整
	// 1. の形式を1) に変換（Discordの表示を改善）
	trimmed := strings.TrimSpace(line)
	if len(trimmed) >= 2 && trimmed[1] == '.' {
		// 番号付きリストの可能性
		if trimmed[0] >= '0' && trimmed[0] <= '9' {
			// 数字. の形式を数字) に変換
			return strings.Replace(line, ". ", ") ", 1)
		}
	}

	return line
}

// splitMessage は、長いメッセージをDiscordの制限に合わせて分割します
func (h *ResponseHandler) splitMessage(message string) []string {
	if len(message) <= DiscordMessageLimit {
		return []string{message}
	}

	var chunks []string
	remaining := message

	for len(remaining) > 0 {
		if len(remaining) <= DiscordMessageLimit {
			chunks = append(chunks, remaining)
			break
		}

		// 2000文字以内で最も近い改行位置を探す
		splitIndex := DiscordMessageLimit
		for i := DiscordMessageLimit; i > 0; i-- {
			if remaining[i-1] == '\n' {
				splitIndex = i
				break
			}
		}

		// 改行が見つからない場合は、単語の境界で分割
		if splitIndex == DiscordMessageLimit {
			for i := DiscordMessageLimit; i > 0; i-- {
				if remaining[i-1] == ' ' {
					splitIndex = i
					break
				}
			}
		}

		// それでも見つからない場合は強制的に分割
		if splitIndex == DiscordMessageLimit {
			splitIndex = DiscordMessageLimit
		}

		chunk := remaining[:splitIndex]
		remaining = remaining[splitIndex:]

		// 先頭の空白を除去
		remaining = strings.TrimLeft(remaining, " \n")

		chunks = append(chunks, chunk)
	}

	return chunks
}

// isTimeoutError は、エラーがタイムアウトエラーかどうかを判定します
func (h *ResponseHandler) isTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	// context.DeadlineExceeded エラーの検出
	if err.Error() == "context deadline exceeded" {
		return true
	}

	// タイムアウト関連のエラーメッセージを検出
	errorMsg := err.Error()
	timeoutKeywords := []string{
		"timeout",
		"タイムアウト",
		"deadline exceeded",
		"context deadline",
		"request timeout",
	}

	for _, keyword := range timeoutKeywords {
		if strings.Contains(strings.ToLower(errorMsg), strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

// formatError は、エラーを適切なメッセージにフォーマットします
func (h *ResponseHandler) formatError(err error) string {
	// タイムアウトエラーの場合
	if h.isTimeoutError(err) {
		return "⏰ **タイムアウトしました**\n\n処理に時間がかかりすぎました。以下の対処法をお試しください：\n\n" +
			"• 質問を短くしてみる\n" +
			"• 複雑な質問を分割する\n" +
			"• しばらく待ってから再度お試しください\n\n" +
			"ご不便をおかけして申し訳ございません。"
	}

	// 荒らし対策エラーの場合
	switch err.Error() {
	case "レート制限を超過しました":
		return "⚠️ **レート制限を超過しました**\nしばらく待ってから再度お試しください。"
	case "スパムが検出されました":
		return "🚫 **スパムが検出されました**\n短時間での大量メッセージは禁止されています。"
	case "不適切なコンテンツが検出されました":
		return "🚫 **不適切なコンテンツが検出されました**\n禁止ワードが含まれています。"
	case "メッセージが長すぎます":
		return "📏 **メッセージが長すぎます**\n2000文字以内でお願いします。"
	case "重複メッセージが検出されました":
		return "🔄 **重複メッセージが検出されました**\n同じ内容のメッセージを連続で送信しないでください。"
	default:
		return fmt.Sprintf("❌ **エラーが発生しました**\n%s", err.Error())
	}
}

// sendImageGenerationResult は、画像生成結果をスレッド内に送信します
func (h *ResponseHandler) sendImageGenerationResult(s *discordgo.Session, threadID string, result *domain.ImageGenerationResult) {
	if !result.Success {
		errorMsg := h.formatImageGenerationError(fmt.Errorf(result.Error))
		s.ChannelMessageSend(threadID, errorMsg)
		return
	}

	// 画像URLかテキストかを判定
	if h.isImageURL(result.ImageURL) {
		// 実際の画像URLの場合
		message := fmt.Sprintf("🎨 **画像生成完了！**\n\n**プロンプト:** %s\n**モデル:** %s\n**生成時刻:** %s",
			result.Response.Prompt, result.Response.Model, result.Response.GeneratedAt)

		// 画像生成結果メッセージを送信
		_, err := s.ChannelMessageSend(threadID, message)
		if err != nil {
			log.Printf("画像生成結果メッセージの送信に失敗: %v", err)
		}

		// 画像をダウンロードしてDiscordにアップロード
		err = h.uploadImageToDiscord(s, threadID, result.ImageURL)
		if err != nil {
			log.Printf("画像のアップロードに失敗: %v", err)
			// フォールバック: 画像情報とURLを送信
			fallbackMsg := fmt.Sprintf("📷 **画像生成完了（URL表示）**\n\n**画像URL:**\n%s\n\n*注: 画像の直接表示に失敗しました。上記URLをブラウザで開いてご確認ください。*", result.ImageURL)
			_, err = s.ChannelMessageSend(threadID, fallbackMsg)
			if err != nil {
				log.Printf("フォールバックメッセージの送信に失敗: %v", err)
			}
		}
	} else {
		// テキストレスポンスの場合（nano bananaの説明文など）
		message := fmt.Sprintf("🎨 **画像生成レスポンス**\n\n**プロンプト:** %s\n**モデル:** %s\n**生成時刻:** %s\n\n**レスポンス:**\n%s",
			result.Response.Prompt, result.Response.Model, result.Response.GeneratedAt, result.ImageURL)

		// テキストレスポンスを送信
		_, err := s.ChannelMessageSend(threadID, message)
		if err != nil {
			log.Printf("画像生成テキストレスポンスの送信に失敗: %v", err)
		}
	}
}

// sendImageGenerationResultToChannel は、画像生成結果をチャンネルに送信します
func (h *ResponseHandler) sendImageGenerationResultToChannel(s *discordgo.Session, m *discordgo.MessageCreate, result *domain.ImageGenerationResult) {
	if !result.Success {
		errorMsg := h.formatImageGenerationError(fmt.Errorf(result.Error))
		s.ChannelMessageSendReply(m.ChannelID, errorMsg, &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})
		return
	}

	// 画像URLかテキストかを判定
	if h.isImageURL(result.ImageURL) {
		// 実際の画像URLの場合
		message := fmt.Sprintf("🎨 **画像生成完了！**\n\n**プロンプト:** %s\n**モデル:** %s\n**生成時刻:** %s",
			result.Response.Prompt, result.Response.Model, result.Response.GeneratedAt)

		// 画像生成結果メッセージを送信
		_, err := s.ChannelMessageSendReply(m.ChannelID, message, &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})
		if err != nil {
			log.Printf("画像生成結果メッセージの送信に失敗: %v", err)
		}

		// 画像をダウンロードしてDiscordにアップロード
		err = h.uploadImageToDiscordWithReply(s, m, result.ImageURL)
		if err != nil {
			log.Printf("画像のアップロードに失敗: %v", err)
			// フォールバック: 画像情報とURLを送信
			fallbackMsg := fmt.Sprintf("📷 **画像生成完了（URL表示）**\n\n**画像URL:**\n%s\n\n*注: 画像の直接表示に失敗しました。上記URLをブラウザで開いてご確認ください。*", result.ImageURL)
			_, err = s.ChannelMessageSendReply(m.ChannelID, fallbackMsg, &discordgo.MessageReference{
				MessageID: m.ID,
				ChannelID: m.ChannelID,
				GuildID:   m.GuildID,
			})
			if err != nil {
				log.Printf("フォールバックメッセージの送信に失敗: %v", err)
			}
		}
	} else {
		// テキストレスポンスの場合（nano bananaの説明文など）
		message := fmt.Sprintf("🎨 **画像生成レスポンス**\n\n**プロンプト:** %s\n**モデル:** %s\n**生成時刻:** %s\n\n**レスポンス:**\n%s",
			result.Response.Prompt, result.Response.Model, result.Response.GeneratedAt, result.ImageURL)

		// テキストレスポンスを送信
		_, err := s.ChannelMessageSendReply(m.ChannelID, message, &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})
		if err != nil {
			log.Printf("画像生成テキストレスポンスの送信に失敗: %v", err)
		}
	}
}

// formatImageGenerationError は、画像生成エラーを適切なメッセージにフォーマットします
func (h *ResponseHandler) formatImageGenerationError(err error) string {
	if err == nil {
		return "❌ **不明なエラーが発生しました**"
	}

	errorMsg := err.Error()

	// 安全フィルターエラーの場合
	if strings.Contains(errorMsg, "安全フィルター") {
		return "🚫 **安全フィルターにより画像生成がブロックされました**\n\n" +
			"プロンプトに不適切な内容が含まれている可能性があります。\n" +
			"より適切な表現で再度お試しください。"
	}

	// タイムアウトエラーの場合
	if h.isTimeoutError(err) {
		return "⏰ **画像生成がタイムアウトしました**\n\n" +
			"処理に時間がかかりすぎました。以下の対処法をお試しください：\n\n" +
			"• プロンプトを短くしてみる\n" +
			"• しばらく待ってから再度お試しください\n\n" +
			"ご不便をおかけして申し訳ございません。"
	}

	// その他のエラー
	return fmt.Sprintf("❌ **画像生成エラー**\n%s", err.Error())
}

// isImageURL は、文字列が画像URLかどうかを判定します
func (h *ResponseHandler) isImageURL(text string) bool {
	// HTTP/HTTPSで始まるかチェック
	if !strings.HasPrefix(text, "http://") && !strings.HasPrefix(text, "https://") {
		return false
	}

	// 画像ファイル拡張子をチェック
	lowerText := strings.ToLower(text)
	if strings.Contains(lowerText, ".jpg") || strings.Contains(lowerText, ".png") ||
		strings.Contains(lowerText, ".jpeg") || strings.Contains(lowerText, ".gif") ||
		strings.Contains(lowerText, ".webp") || strings.Contains(lowerText, ".bmp") {
		return true
	}

	// 画像ホスティングサービスのURLパターンをチェック
	if strings.Contains(lowerText, "imgur.com") || strings.Contains(lowerText, "i.imgur.com") ||
		strings.Contains(lowerText, "drive.google.com") || strings.Contains(lowerText, "photos.google.com") ||
		strings.Contains(lowerText, "cloudinary.com") || strings.Contains(lowerText, "unsplash.com") ||
		strings.Contains(lowerText, "files.oaiusercontent.com") {
		return true
	}

	return false
}

// uploadImageToDiscord は、画像URLから画像をダウンロードしてDiscordにアップロードします
func (h *ResponseHandler) uploadImageToDiscord(s *discordgo.Session, channelID, imageURL string) error {
	log.Printf("画像をダウンロード中: %s", imageURL)

	// HTTPクライアントを作成（タイムアウト設定）
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// リクエストを作成（User-Agentヘッダーを追加）
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return fmt.Errorf("リクエストの作成に失敗: %w", err)
	}

	// User-Agentを設定（ブラウザとして認識させる）
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// 画像をダウンロード
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("画像のダウンロードに失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("画像のダウンロードに失敗: HTTP %d", resp.StatusCode)
	}

	// 画像データを読み込み
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("画像データの読み込みに失敗: %w", err)
	}

	// ファイル名を生成
	filename := "generated_image.png"
	if strings.Contains(imageURL, ".jpg") || strings.Contains(imageURL, ".jpeg") {
		filename = "generated_image.jpg"
	} else if strings.Contains(imageURL, ".gif") {
		filename = "generated_image.gif"
	} else if strings.Contains(imageURL, ".webp") {
		filename = "generated_image.webp"
	}

	// Discordにファイルをアップロード
	_, err = s.ChannelFileSend(channelID, filename, strings.NewReader(string(imageData)))
	if err != nil {
		return fmt.Errorf("Discordへの画像アップロードに失敗: %w", err)
	}

	log.Printf("画像のアップロードが完了しました: %s", filename)
	return nil
}

// uploadImageToDiscordWithReply は、画像URLから画像をダウンロードしてDiscordにリプライ付きでアップロードします
func (h *ResponseHandler) uploadImageToDiscordWithReply(s *discordgo.Session, m *discordgo.MessageCreate, imageURL string) error {
	log.Printf("画像をダウンロード中: %s", imageURL)

	// HTTPクライアントを作成（タイムアウト設定）
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// リクエストを作成（User-Agentヘッダーを追加）
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return fmt.Errorf("リクエストの作成に失敗: %w", err)
	}

	// User-Agentを設定（ブラウザとして認識させる）
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// 画像をダウンロード
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("画像のダウンロードに失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("画像のダウンロードに失敗: HTTP %d", resp.StatusCode)
	}

	// 画像データを読み込み
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("画像データの読み込みに失敗: %w", err)
	}

	// ファイル名を生成
	filename := "generated_image.png"
	if strings.Contains(imageURL, ".jpg") || strings.Contains(imageURL, ".jpeg") {
		filename = "generated_image.jpg"
	} else if strings.Contains(imageURL, ".gif") {
		filename = "generated_image.gif"
	} else if strings.Contains(imageURL, ".webp") {
		filename = "generated_image.webp"
	}

	// Discordにファイルをアップロード（リプライ付き）
	_, err = s.ChannelFileSendWithMessage(m.ChannelID, "", filename, strings.NewReader(string(imageData)))
	if err != nil {
		return fmt.Errorf("Discordへの画像アップロードに失敗: %w", err)
	}

	log.Printf("画像のアップロードが完了しました: %s", filename)
	return nil
}
