package discord

import (
	"fmt"
	"log"
	"strings"

	"geminibot/internal/domain"

	"github.com/bwmarrin/discordgo"
)

// ResponseHandler は、Discordのレスポンス送信・フォーマット処理を担当するハンドラーです
type ResponseHandler struct{}

// DiscordMessageLimit は、Discordのメッセージ文字数制限です
const DiscordMessageLimit = 2000

// NewResponseHandler は新しいResponseHandlerインスタンスを作成します
func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

// SendUnifiedResponse は、統一レスポンスを送信します（ThreadIDに基づいてスレッドまたはリプライで送信）
func (h *ResponseHandler) SendUnifiedResponse(s *discordgo.Session, m *discordgo.MessageCreate, response *domain.UnifiedResponse) {
	// エラーレスポンスの場合は直接リプライで送信
	if !response.Success {
		errorMsg := h.formatUnifiedError(response)
		s.ChannelMessageSendReply(m.ChannelID, errorMsg, &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})
		return
	}

	var targetChannelID string
	var isReply bool

	// ThreadIDが設定されている場合はスレッド内に送信
	if response.ThreadID != "" {
		targetChannelID = response.ThreadID
		isReply = false
	} else {
		// ThreadIDが空の場合はスレッド作成を試行
		threadID, err := h.createThreadForResponse(s, m, response)
		if err != nil {
			log.Printf("スレッド作成に失敗、リプライで送信します: %v", err)
			// スレッド作成に失敗した場合はリプライで送信
			targetChannelID = m.ChannelID
			isReply = true
		} else {
			// スレッド内に送信
			targetChannelID = threadID
			isReply = false
		}
	}

	// テキストコンテンツがある場合は送信
	if response.Content != "" {
		if isReply {
			h.sendTextContentToChannel(s, m, response.Content)
		} else {
			h.sendTextContentToThread(s, targetChannelID, response.Content)
		}
	}

	// 添付ファイルがある場合は送信
	if response.HasAttachments() {
		if isReply {
			h.sendAttachmentsToChannel(s, m, response.Attachments, response.Metadata)
		} else {
			h.sendAttachmentsToThread(s, targetChannelID, response.Attachments, response.Metadata)
		}
	}
}

// createThreadForResponse は、レスポンス用のスレッドを作成します
func (h *ResponseHandler) createThreadForResponse(s *discordgo.Session, m *discordgo.MessageCreate, response *domain.UnifiedResponse) (string, error) {
	// 既にスレッド内の場合はスレッド作成をスキップ
	if h.isInThread(s, m.ChannelID) {
		return "", fmt.Errorf("既にスレッド内です")
	}

	// スレッド名を生成
	threadName := h.generateThreadName(m, response)

	// スレッドを作成
	thread, err := s.MessageThreadStartComplex(m.ChannelID, m.ID, &discordgo.ThreadStart{
		Name:                threadName,
		AutoArchiveDuration: 60, // 1時間後に自動アーカイブ
		Invitable:           false,
	})
	if err != nil {
		return "", fmt.Errorf("スレッド作成に失敗: %w", err)
	}

	log.Printf("スレッドを作成しました: %s (ID: %s)", threadName, thread.ID)
	return thread.ID, nil
}

// isInThread は、指定されたチャンネルがスレッドかどうかを判定します
func (h *ResponseHandler) isInThread(s *discordgo.Session, channelID string) bool {
	// DiscordのスレッドチャンネルIDは通常のチャンネルIDと異なる形式を持つ場合があります
	// 実際の実装では、Discord APIの仕様に基づいて判定ロジックを調整する必要があります
	// ここでは簡易的な実装として、チャンネル情報を取得して判定
	channel, err := s.Channel(channelID)
	if err != nil {
		log.Printf("チャンネル情報の取得に失敗: %v", err)
		return false
	}

	// スレッドの場合はParentIDが設定されている
	return channel.ParentID != ""
}

// generateThreadName は、スレッド名を生成します
func (h *ResponseHandler) generateThreadName(_ *discordgo.MessageCreate, response *domain.UnifiedResponse) string {
	// レスポンスタイプに基づいてスレッド名を生成
	content := response.Content
	if len(content) > 20 {
		content = content[:20] + "..."
	}
	switch response.Metadata.Type {
	case "image":
		return "🎨 " + content
	case "text":
		return "💬 " + content
	default:
		return "🤖 Bot応答"
	}
}

// sendTextContentToThread は、テキストコンテンツをスレッド内に送信します
func (h *ResponseHandler) sendTextContentToThread(s *discordgo.Session, threadID string, content string) {
	// 応答が非常に長い場合はファイルとして送信
	if len(content) > DiscordMessageLimit*5 {
		h.sendAsFileToThread(s, threadID, content, "response.txt")
		return
	}

	// 応答をDiscordの制限に合わせて分割
	chunks := h.splitMessage(content)

	// すべてのチャンクをスレッド内に送信
	for i, chunk := range chunks {
		_, err := s.ChannelMessageSend(threadID, chunk)
		if err != nil {
			log.Printf("スレッド内メッセージの送信に失敗 (チャンク %d): %v", i+1, err)
			break
		}
	}
}

// sendTextContentToChannel は、テキストコンテンツをチャンネルにリプライ付きで送信します
func (h *ResponseHandler) sendTextContentToChannel(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	// 応答が非常に長い場合はファイルとして送信
	if len(content) > DiscordMessageLimit*5 {
		h.sendAsFile(s, m, content, "response.txt")
		return
	}

	// 応答をDiscordの制限に合わせて分割
	chunks := h.splitMessage(content)

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

// sendAttachmentsToThread は、添付ファイルをスレッド内に送信します
func (h *ResponseHandler) sendAttachmentsToThread(s *discordgo.Session, threadID string, attachments []domain.Attachment, metadata domain.ResponseMetadata) {
	// 画像添付がある場合のメッセージを作成
	if len(attachments) > 0 {
		message := h.createAttachmentMessage(metadata)
		if message != "" {
			_, err := s.ChannelMessageSend(threadID, message)
			if err != nil {
				log.Printf("添付ファイルメッセージの送信に失敗: %v", err)
			}
		}
	}

	// 各添付ファイルを送信
	for i, attachment := range attachments {
		if attachment.IsImage {
			err := h.uploadAttachmentToThread(s, threadID, attachment, i+1)
			if err != nil {
				log.Printf("添付ファイルのアップロードに失敗 (ファイル %d): %v", i+1, err)
			}
		}
	}
}

// sendAttachmentsToChannel は、添付ファイルをチャンネルにリプライ付きで送信します
func (h *ResponseHandler) sendAttachmentsToChannel(s *discordgo.Session, m *discordgo.MessageCreate, attachments []domain.Attachment, metadata domain.ResponseMetadata) {
	// 画像添付がある場合のメッセージを作成
	if len(attachments) > 0 {
		message := h.createAttachmentMessage(metadata)
		if message != "" {
			_, err := s.ChannelMessageSendReply(m.ChannelID, message, &discordgo.MessageReference{
				MessageID: m.ID,
				ChannelID: m.ChannelID,
				GuildID:   m.GuildID,
			})
			if err != nil {
				log.Printf("添付ファイルメッセージの送信に失敗: %v", err)
			}
		}
	}

	// 各添付ファイルを送信
	for i, attachment := range attachments {
		if attachment.IsImage {
			err := h.uploadAttachmentToChannel(s, m, attachment, i+1)
			if err != nil {
				log.Printf("添付ファイルのアップロードに失敗 (ファイル %d): %v", i+1, err)
			}
		}
	}
}

// createAttachmentMessage は、添付ファイル用のメッセージを作成します
func (h *ResponseHandler) createAttachmentMessage(metadata domain.ResponseMetadata) string {
	switch metadata.Type {
	case "image":
		return fmt.Sprintf("🎨 **画像生成完了！**\n\n**プロンプト:** %s\n**モデル:** %s\n**生成時刻:** %s",
			metadata.Prompt, metadata.Model, metadata.GeneratedAt.Format("2006-01-02 15:04:05"))
	default:
		return ""
	}
}

// uploadAttachmentToThread は、添付ファイルをスレッド内にアップロードします
func (h *ResponseHandler) uploadAttachmentToThread(s *discordgo.Session, threadID string, attachment domain.Attachment, index int) error {
	// ファイル名を生成
	filename := attachment.Filename
	if filename == "" {
		filename = fmt.Sprintf("attachment_%d", index)
		if attachment.MimeType == "image/png" {
			filename += ".png"
		} else if attachment.MimeType == "image/jpeg" {
			filename += ".jpg"
		} else if attachment.MimeType == "image/gif" {
			filename += ".gif"
		} else if attachment.MimeType == "image/webp" {
			filename += ".webp"
		}
	}

	// Discordにファイルをアップロード
	_, err := s.ChannelFileSend(threadID, filename, strings.NewReader(string(attachment.Data)))
	if err != nil {
		return fmt.Errorf("Discordへのファイルアップロードに失敗: %w", err)
	}

	log.Printf("添付ファイルのアップロードが完了しました: %s", filename)
	return nil
}

// uploadAttachmentToChannel は、添付ファイルをチャンネルにリプライ付きでアップロードします
func (h *ResponseHandler) uploadAttachmentToChannel(s *discordgo.Session, m *discordgo.MessageCreate, attachment domain.Attachment, index int) error {
	// ファイル名を生成
	filename := attachment.Filename
	if filename == "" {
		filename = fmt.Sprintf("attachment_%d", index)
		if attachment.MimeType == "image/png" {
			filename += ".png"
		} else if attachment.MimeType == "image/jpeg" {
			filename += ".jpg"
		} else if attachment.MimeType == "image/gif" {
			filename += ".gif"
		} else if attachment.MimeType == "image/webp" {
			filename += ".webp"
		}
	}

	// Discordにファイルをアップロード（リプライ付き）
	_, err := s.ChannelFileSendWithMessage(m.ChannelID, "", filename, strings.NewReader(string(attachment.Data)))
	if err != nil {
		return fmt.Errorf("Discordへのファイルアップロードに失敗: %w", err)
	}

	log.Printf("添付ファイルのアップロードが完了しました: %s", filename)
	return nil
}

// formatUnifiedError は、統一レスポンスのエラーを適切なメッセージにフォーマットします
func (h *ResponseHandler) formatUnifiedError(response *domain.UnifiedResponse) string {
	if response.Error == "" {
		return "❌ **不明なエラーが発生しました**"
	}

	errorMsg := response.Error

	// タイムアウトエラーの場合
	if h.isTimeoutError(fmt.Errorf(errorMsg)) {
		return "⏰ **タイムアウトしました**\n\n処理に時間がかかりすぎました。以下の対処法をお試しください：\n\n" +
			"• 質問を短くしてみる\n" +
			"• 複雑な質問を分割する\n" +
			"• しばらく待ってから再度お試しください\n\n" +
			"ご不便をおかけして申し訳ございません。"
	}

	// 画像生成関連のエラー
	if response.Metadata.Type == "image" {
		// 安全フィルターエラーの場合
		if strings.Contains(errorMsg, "安全フィルター") {
			return "🚫 **安全フィルターにより画像生成がブロックされました**\n\n" +
				"プロンプトに不適切な内容が含まれている可能性があります。\n" +
				"より適切な表現で再度お試しください。"
		}

		// 画像生成タイムアウトエラーの場合
		if h.isTimeoutError(fmt.Errorf(errorMsg)) {
			return "⏰ **画像生成がタイムアウトしました**\n\n" +
				"処理に時間がかかりすぎました。以下の対処法をお試しください：\n\n" +
				"• プロンプトを短くしてみる\n" +
				"• しばらく待ってから再度お試しください\n\n" +
				"ご不便をおかけして申し訳ございません。"
		}

		return fmt.Sprintf("❌ **画像生成エラー**\n%s", errorMsg)
	}

	// テキスト生成関連のエラー
	switch errorMsg {
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
		return fmt.Sprintf("❌ **エラーが発生しました**\n%s", errorMsg)
	}
}

// convertImageResultToUnifiedResponse は、画像生成結果を統一レスポンスに変換します
func (h *ResponseHandler) convertImageResultToUnifiedResponse(imageResult *domain.ImageGenerationResult, m *discordgo.MessageCreate) *domain.UnifiedResponse {
	if !imageResult.Success {
		return domain.NewErrorResponse(fmt.Errorf(imageResult.Error), "image")
	}

	// メンション部分を除去したコンテンツを取得
	content := h.extractUserContent(m)

	// 画像生成レスポンスから統一レスポンスを作成
	if imageResult.Response != nil && len(imageResult.Response.Images) > 0 {
		return domain.NewImageResponse("", imageResult.Response.Images, content, imageResult.Response.Model)
	}

	// 画像データがない場合はテキストレスポンスとして処理
	return domain.NewTextResponse(imageResult.ImageURL, content, imageResult.Response.Model)
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

// sendAsFileToThread は、長い応答をファイルとしてスレッド内に送信します
func (h *ResponseHandler) sendAsFileToThread(s *discordgo.Session, threadID string, content, filename string) {
	// ファイルデータを作成
	fileData := strings.NewReader(content)

	// ファイルを添付してメッセージを送信
	_, err := s.ChannelFileSend(threadID, filename, fileData)

	if err != nil {
		log.Printf("ファイル送信に失敗: %v", err)
		// ファイル送信に失敗した場合は通常の分割送信にフォールバック
		h.sendTextContentToThread(s, threadID, content)
		return
	}

	// ファイル送信成功のメッセージを送信
	fileMsg := fmt.Sprintf("📄 **応答が長いため、ファイルとして送信しました**\nファイル名: `%s`", filename)
	s.ChannelMessageSend(threadID, fileMsg)
}

// sendSplitResponse は、長い応答を複数のメッセージに分割して送信します（後方互換性のため残す）
func (h *ResponseHandler) sendSplitResponse(s *discordgo.Session, m *discordgo.MessageCreate, response string) {
	// テキストレスポンスを作成
	textResponse := domain.NewTextResponse(response, "", "gemini-pro")
	h.SendUnifiedResponse(s, m, textResponse)
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
			"- 質問を短くしてみる\n" +
			"- 複雑な質問を分割する\n" +
			"- しばらく待ってから再度お試しください\n\n" +
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
