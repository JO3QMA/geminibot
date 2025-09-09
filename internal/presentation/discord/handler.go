package discord

import (
	"geminibot/internal/application"

	"github.com/bwmarrin/discordgo"
)

// DiscordHandler は、Discordのイベントハンドラです
type DiscordHandler struct {
	session             *discordgo.Session
	mentionService      *application.MentionApplicationService
	botID               string
	mentionHandler      *MentionHandler
	slashCommandHandler *SlashCommandHandler
}

// DiscordMessageLimit は、Discordのメッセージ長制限です
const DiscordMessageLimit = 2000

// NewDiscordHandler は新しいDiscordHandlerインスタンスを作成します
func NewDiscordHandler(
	session *discordgo.Session,
	mentionService *application.MentionApplicationService,
	botID string,
	slashCommandHandler *SlashCommandHandler,
) *DiscordHandler {
	// ResponseHandlerを作成
	responseHandler := NewResponseHandler()

	// MentionHandlerを作成
	mentionHandler := NewMentionHandler(session, mentionService, botID, responseHandler)

	return &DiscordHandler{
		session:             session,
		mentionService:      mentionService,
		botID:               botID,
		mentionHandler:      mentionHandler,
		slashCommandHandler: slashCommandHandler,
	}
}

// SetupHandlers は、Discordのイベントハンドラを設定します
func (h *DiscordHandler) SetupHandlers() {
	// メンションハンドラーを設定
	h.mentionHandler.SetupHandlers()

	// スラッシュコマンドハンドラーを設定
	if h.slashCommandHandler != nil {
		h.slashCommandHandler.SetupSlashCommandHandlers()
	}
}
