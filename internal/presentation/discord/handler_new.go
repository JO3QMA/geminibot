package discord

import (
	"log"

	"geminibot/internal/application"

	"github.com/bwmarrin/discordgo"
)

// DiscordHandlerNew は、リファクタリングされたDiscordのイベントハンドラです
type DiscordHandlerNew struct {
	session        *discordgo.Session
	mentionService *application.MentionApplicationService
	processor      *MessageProcessor
	sender         *MessageSender
	errorHandler   *ErrorHandler
	formatter      *MessageFormatter
	botID          string
	botUsername    string
}

// NewDiscordHandlerNew は新しいDiscordHandlerNewインスタンスを作成します
func NewDiscordHandlerNew(
	session *discordgo.Session,
	mentionService *application.MentionApplicationService,
	botID string,
) *DiscordHandlerNew {
	// コンポーネントを作成
	formatter := NewMessageFormatter()
	sender := NewMessageSender(formatter)
	errorHandler := NewErrorHandler()
	processor := NewMessageProcessor(mentionService, sender, errorHandler)

	return &DiscordHandlerNew{
		session:        session,
		mentionService: mentionService,
		processor:      processor,
		sender:         sender,
		errorHandler:   errorHandler,
		formatter:      formatter,
		botID:          botID,
	}
}

// SetupHandlers は、Discordのイベントハンドラを設定します
func (h *DiscordHandlerNew) SetupHandlers() {
	h.session.AddHandler(h.handleMessageCreate)
	h.session.AddHandler(h.handleReady)
}

// handleReady は、Botが準備完了した際のイベントを処理します
func (h *DiscordHandlerNew) handleReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("Botが準備完了しました: %s#%s", event.User.Username, event.User.Discriminator)
	h.botUsername = event.User.Username
}

// handleMessageCreate は、メッセージ作成イベントを処理します
func (h *DiscordHandlerNew) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Bot自身のメッセージは無視
	if m.Author.ID == h.botID {
		return
	}

	// メンションされているかチェック
	if !h.processor.IsMentioned(m, h.botID, h.botUsername) {
		return
	}

	log.Printf("Botへのメンションを検出: %s", m.Content)

	// メンション情報を作成
	mention := h.processor.CreateBotMention(s, m)

	// 非同期でメンションを処理
	go h.processor.ProcessMentionAsync(s, m, mention)
}
