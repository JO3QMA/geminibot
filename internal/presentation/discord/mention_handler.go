package discord

import (
	"context"
	"fmt"
	"log"
	"strings"

	"geminibot/internal/application"
	"geminibot/internal/domain"

	"github.com/bwmarrin/discordgo"
)

// MentionHandler は、Discordのメンション処理を担当するハンドラーです
type MentionHandler struct {
	session         *discordgo.Session
	mentionService  *application.MentionApplicationService
	botID           string
	botUsername     string
	responseHandler *ResponseHandler
}

// NewMentionHandler は新しいMentionHandlerインスタンスを作成します
func NewMentionHandler(
	session *discordgo.Session,
	mentionService *application.MentionApplicationService,
	botID string,
	responseHandler *ResponseHandler,
) *MentionHandler {
	return &MentionHandler{
		session:         session,
		mentionService:  mentionService,
		botID:           botID,
		responseHandler: responseHandler,
	}
}

// SetupHandlers は、メンション関連のイベントハンドラを設定します
func (h *MentionHandler) SetupHandlers() {
	h.session.AddHandler(h.handleMessageCreate)
	h.session.AddHandler(h.handleReady)
}

// SetBotUsername は、Botのユーザー名を設定します
func (h *MentionHandler) SetBotUsername(username string) {
	h.botUsername = username
}

// handleReady は、Botが準備完了した際のイベントを処理します
func (h *MentionHandler) handleReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("Botが準備完了しました: %s#%s", event.User.Username, event.User.Discriminator)
	h.botUsername = event.User.Username
}

// handleMessageCreate は、メッセージ作成イベントを処理します
func (h *MentionHandler) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Bot自身のメッセージは無視
	if m.Author.ID == h.botID {
		return
	}

	// メンションされているかチェック
	if !h.isMentioned(m) {
		return
	}

	log.Printf("Botへのメンションを検出: %s", m.Content)

	// 画像生成リクエストかどうかをチェック
	if h.isImageGenerationRequest(m.Content) {
		log.Printf("画像生成リクエストを検出: %s", m.Content)
		// 非同期で画像生成を処理
		go h.processImageGenerationAsync(s, m)
		return
	}

	// メンション情報を作成
	mention := h.createBotMention(m)

	// 非同期でメンションを処理
	go h.processMentionAsync(s, m, mention)
}

// isMentioned は、メッセージがBotへのメンションかどうかを判定します
func (h *MentionHandler) isMentioned(m *discordgo.MessageCreate) bool {
	// メンション配列をチェック
	for _, mention := range m.Mentions {
		if mention.ID == h.botID {
			return true
		}
	}

	// メンション配列が空の場合、コンテンツをチェック
	if len(m.Mentions) == 0 {
		content := strings.ToLower(m.Content)
		botMention := fmt.Sprintf("@%s", strings.ToLower(h.botUsername))
		return strings.Contains(content, botMention)
	}

	return false
}

// createBotMention は、DiscordメッセージからBotMentionオブジェクトを作成します
func (h *MentionHandler) createBotMention(m *discordgo.MessageCreate) domain.BotMention {
	// メンション部分を除去したコンテンツを取得
	content := h.extractUserContent(m)

	// ユーザー情報を作成
	user := domain.User{
		ID:            m.Author.ID,
		Username:      m.Author.Username,
		DisplayName:   h.getDisplayName(m),
		Avatar:        m.Author.Avatar,
		Discriminator: m.Author.Discriminator,
		IsBot:         m.Author.Bot,
	}

	return domain.BotMention{
		ChannelID: m.ChannelID,
		GuildID:   m.GuildID,
		User:      user,
		Content:   content,
		MessageID: m.ID,
	}
}

// extractUserContent は、メンション部分を除去したユーザーのコンテンツを抽出します
func (h *MentionHandler) extractUserContent(m *discordgo.MessageCreate) string {
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

// getDisplayName は、Discordメッセージから表示名を取得します
func (h *MentionHandler) getDisplayName(m *discordgo.MessageCreate) string {
	// メンバー情報がある場合はニックネームを優先
	if m.Member != nil && m.Member.Nick != "" {
		return m.Member.Nick
	}

	// メンバー情報がない場合は、Discord APIからメンバー情報を取得を試行
	if m.GuildID != "" {
		member, err := h.session.GuildMember(m.GuildID, m.Author.ID)
		if err == nil && member.Nick != "" {
			return member.Nick
		}
	}

	// ニックネームがない場合はユーザー名を使用
	return m.Author.Username
}

// processMentionAsync は、メンションを非同期で処理します
func (h *MentionHandler) processMentionAsync(s *discordgo.Session, m *discordgo.MessageCreate, mention domain.BotMention) {
	// メッセージからスレッドを作成
	thread, err := s.MessageThreadStart(m.ChannelID, m.ID, "Bot応答", 60) // 60分後にアーカイブ
	if err != nil {
		log.Printf("スレッド作成に失敗: %v", err)
		// スレッド作成に失敗した場合は通常のリプライとして送信
		h.responseHandler.sendNormalReply(s, m, mention, h.mentionService)
		return
	}

	// 処理中メッセージをスレッド内に送信
	thinkingMsg, err := s.ChannelMessageSend(thread.ID, "🤔 考え中...")
	if err != nil {
		log.Printf("処理中メッセージの送信に失敗: %v", err)
		return
	}

	// メンションを処理
	ctx := context.Background()
	response, err := h.mentionService.HandleMention(ctx, mention)

	// 処理中メッセージを削除
	s.ChannelMessageDelete(thread.ID, thinkingMsg.ID)

	if err != nil {
		log.Printf("メンション処理に失敗: %v", err)

		// エラーを適切なメッセージにフォーマット
		errorMsg := h.responseHandler.formatError(err)
		s.ChannelMessageSend(thread.ID, errorMsg)
		return
	}

	// 応答をスレッド内に送信
	h.responseHandler.sendThreadResponse(s, thread.ID, response)
}

// isImageGenerationRequest は、メッセージが画像生成リクエストかどうかを判定します
func (h *MentionHandler) isImageGenerationRequest(content string) bool {
	keywords := []string{
		"画像生成", "画像作成", "絵を描いて", "イラスト作成", "画像を作って",
		"generate image", "create image", "draw", "illustration", "picture",
		"画像", "絵", "イラスト", "ピクチャー", "写真",
	}

	content = strings.ToLower(content)

	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

// processImageGenerationAsync は、画像生成を非同期で処理します
func (h *MentionHandler) processImageGenerationAsync(s *discordgo.Session, m *discordgo.MessageCreate) {
	// メッセージからスレッドを作成
	thread, err := s.MessageThreadStart(m.ChannelID, m.ID, "画像生成中...", 60) // 60分後にアーカイブ
	if err != nil {
		log.Printf("スレッド作成に失敗: %v", err)
		// スレッド作成に失敗した場合は通常のリプライとして送信
		h.responseHandler.sendImageGenerationNormalReply(s, m, h.mentionService)
		return
	}

	// 処理中メッセージをスレッド内に送信
	thinkingMsg, err := s.ChannelMessageSend(thread.ID, "🎨 画像を生成中...")
	if err != nil {
		log.Printf("処理中メッセージの送信に失敗: %v", err)
		return
	}

	// 画像生成を処理
	ctx := context.Background()
	imageResult, err := h.generateImage(ctx, m)

	// 処理中メッセージを削除
	s.ChannelMessageDelete(thread.ID, thinkingMsg.ID)

	if err != nil {
		log.Printf("画像生成に失敗: %v", err)
		errorMsg := h.responseHandler.formatImageGenerationError(err)
		s.ChannelMessageSend(thread.ID, errorMsg)
		return
	}

	// 画像生成結果をスレッド内に送信
	h.responseHandler.sendImageGenerationResult(s, thread.ID, imageResult)
}

// generateImage は、画像生成を実行します
func (h *MentionHandler) generateImage(ctx context.Context, m *discordgo.MessageCreate) (*domain.ImageGenerationResult, error) {
	// メンション部分を除去したコンテンツを取得
	content := h.extractUserContent(m)

	// 画像生成用のプロンプトを作成
	prompt := domain.NewImagePrompt(content)

	// Geminiクライアントを使用して画像生成
	response, err := h.mentionService.GenerateImage(ctx, domain.ImageGenerationRequest{
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
