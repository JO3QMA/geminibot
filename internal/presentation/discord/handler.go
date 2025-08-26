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

// DiscordHandler は、Discordのイベントハンドラです
type DiscordHandler struct {
	session        *discordgo.Session
	mentionService *application.MentionApplicationService
	botID          string
	botUsername    string
}

// NewDiscordHandler は新しいDiscordHandlerインスタンスを作成します
func NewDiscordHandler(
	session *discordgo.Session,
	mentionService *application.MentionApplicationService,
	botID string,
) *DiscordHandler {
	return &DiscordHandler{
		session:        session,
		mentionService: mentionService,
		botID:          botID,
	}
}

// SetupHandlers は、Discordのイベントハンドラを設定します
func (h *DiscordHandler) SetupHandlers() {
	h.session.AddHandler(h.handleMessageCreate)
	h.session.AddHandler(h.handleReady)
}

// handleReady は、Botが準備完了した際のイベントを処理します
func (h *DiscordHandler) handleReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("Botが準備完了しました: %s#%s", event.User.Username, event.User.Discriminator)
	h.botUsername = event.User.Username
}

// handleMessageCreate は、メッセージ作成イベントを処理します
func (h *DiscordHandler) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Bot自身のメッセージは無視
	if m.Author.ID == h.botID {
		return
	}

	// メンションされているかチェック
	if !h.isMentioned(m) {
		return
	}

	log.Printf("Botへのメンションを検出: %s", m.Content)

	// メンション情報を作成
	mention := h.createBotMention(m)

	// 非同期でメンションを処理
	go h.processMentionAsync(s, m, mention)
}

// isMentioned は、メッセージがBotへのメンションかどうかを判定します
func (h *DiscordHandler) isMentioned(m *discordgo.MessageCreate) bool {
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
func (h *DiscordHandler) createBotMention(m *discordgo.MessageCreate) domain.BotMention {
	// メンション部分を除去したコンテンツを取得
	content := h.extractUserContent(m)

	return domain.NewBotMention(
		domain.NewChannelID(m.ChannelID),
		domain.NewUserID(m.Author.ID),
		content,
		m.ID,
	)
}

// extractUserContent は、メンション部分を除去したユーザーのコンテンツを抽出します
func (h *DiscordHandler) extractUserContent(m *discordgo.MessageCreate) string {
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

// processMentionAsync は、メンションを非同期で処理します
func (h *DiscordHandler) processMentionAsync(s *discordgo.Session, m *discordgo.MessageCreate, mention domain.BotMention) {
	// 処理中メッセージを送信
	thinkingMsg, err := s.ChannelMessageSend(m.ChannelID, "🤔 考え中...")
	if err != nil {
		log.Printf("処理中メッセージの送信に失敗: %v", err)
		return
	}

	// メンションを処理
	ctx := context.Background()
	response, err := h.mentionService.HandleMention(ctx, mention)

	// 処理中メッセージを削除
	s.ChannelMessageDelete(m.ChannelID, thinkingMsg.ID)

	if err != nil {
		log.Printf("メンション処理に失敗: %v", err)
		errorMsg := fmt.Sprintf("❌ エラーが発生しました: %s", err.Error())
		s.ChannelMessageSendReply(m.ChannelID, errorMsg, &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})
		return
	}

	// 応答を送信
	_, err = s.ChannelMessageSendReply(m.ChannelID, response, &discordgo.MessageReference{
		MessageID: m.ID,
		ChannelID: m.ChannelID,
		GuildID:   m.GuildID,
	})

	if err != nil {
		log.Printf("応答メッセージの送信に失敗: %v", err)
	}
}
