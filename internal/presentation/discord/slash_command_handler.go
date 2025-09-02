package discord

import (
	"context"
	"fmt"
	"log"

	"geminibot/internal/application"

	"github.com/bwmarrin/discordgo"
)

// SlashCommandHandler は、Discordのスラッシュコマンドを処理するハンドラーです
type SlashCommandHandler struct {
	session       *discordgo.Session
	apiKeyService *application.APIKeyApplicationService
	defaultAPIKey string
}

// NewSlashCommandHandler は新しいSlashCommandHandlerインスタンスを作成します
func NewSlashCommandHandler(
	session *discordgo.Session,
	apiKeyService *application.APIKeyApplicationService,
	defaultAPIKey string,
) *SlashCommandHandler {
	return &SlashCommandHandler{
		session:       session,
		apiKeyService: apiKeyService,
		defaultAPIKey: defaultAPIKey,
	}
}

// SetupSlashCommands は、スラッシュコマンドを設定します
func (h *SlashCommandHandler) SetupSlashCommands() error {
	// BotのユーザーIDを取得
	user, err := h.session.User("@me")
	if err != nil {
		return fmt.Errorf("Botユーザー情報の取得に失敗: %w", err)
	}

	// スラッシュコマンドの定義
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "set-api",
			Description: "このサーバー用のGemini APIキーを設定します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "api-key",
					Description: "Gemini APIキー",
					Required:    true,
				},
			},
		},
		{
			Name:        "del-api",
			Description: "このサーバー用のGemini APIキーを削除します",
		},
		{
			Name:        "set-model",
			Description: "このサーバーで使用するAIモデルを設定します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "model",
					Description: "使用するAIモデル",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Gemini Pro", Value: "gemini-pro"},
						{Name: "Gemini 1.5 Flash", Value: "gemini-1.5-flash"},
						{Name: "Gemini 1.5 Nano", Value: "gemini-1.5-nano"},
						{Name: "Gemini Pro Vision", Value: "gemini-pro-vision"},
					},
				},
			},
		},
		{
			Name:        "status",
			Description: "このサーバーのGemini APIキー設定状況を表示します",
		},
	}

	// グローバルコマンドとして登録
	for _, command := range commands {
		_, err := h.session.ApplicationCommandCreate(user.ID, "", command)
		if err != nil {
			log.Printf("スラッシュコマンド %s の登録に失敗: %v", command.Name, err)
			return err
		}
		log.Printf("スラッシュコマンド %s を登録しました", command.Name)
	}

	return nil
}

// SetupSlashCommandHandlers は、スラッシュコマンドのハンドラーを設定します
func (h *SlashCommandHandler) SetupSlashCommandHandlers() {
	h.session.AddHandler(h.handleInteractionCreate)
}

// handleInteractionCreate は、インタラクション作成イベントを処理します
func (h *SlashCommandHandler) handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "set-api":
		h.handleSetAPICommand(s, i)
	case "del-api":
		h.handleDelAPICommand(s, i)
	case "set-model":
		h.handleSetModelCommand(s, i)
	case "status":
		h.handleStatusCommand(s, i)
	default:
		log.Printf("未知のスラッシュコマンド: %s", i.ApplicationCommandData().Name)
	}
}

// handleSetAPICommand は、/set-apiコマンドを処理します
func (h *SlashCommandHandler) handleSetAPICommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// 権限チェック（管理者権限が必要）
	if !h.hasAdminPermission(i.Member) {
		h.respondToInteraction(s, i, "❌ このコマンドを実行するには管理者権限が必要です。", true)
		return
	}

	// APIキーを取得
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		h.respondToInteraction(s, i, "❌ APIキーが指定されていません。", true)
		return
	}

	apiKey := options[0].StringValue()
	guildID := i.GuildID
	setBy := i.Member.User.Username

	// APIキーを設定
	ctx := context.Background()
	err := h.apiKeyService.SetGuildAPIKey(ctx, guildID, apiKey, setBy)
	if err != nil {
		log.Printf("APIキーの設定に失敗: %v", err)
		h.respondToInteraction(s, i, fmt.Sprintf("❌ APIキーの設定に失敗しました: %v", err), true)
		return
	}

	// 成功メッセージを送信
	successMsg := fmt.Sprintf("✅ このサーバー用のGemini APIキーを設定しました。\n設定者: %s", setBy)
	h.respondToInteraction(s, i, successMsg, false)
}

// handleDelAPICommand は、/del-apiコマンドを処理します
func (h *SlashCommandHandler) handleDelAPICommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// 権限チェック（管理者権限が必要）
	if !h.hasAdminPermission(i.Member) {
		h.respondToInteraction(s, i, "❌ このコマンドを実行するには管理者権限が必要です。", true)
		return
	}

	guildID := i.GuildID

	// APIキーを削除
	ctx := context.Background()
	err := h.apiKeyService.DeleteGuildAPIKey(ctx, guildID)
	if err != nil {
		log.Printf("APIキーの削除に失敗: %v", err)
		h.respondToInteraction(s, i, fmt.Sprintf("❌ APIキーの削除に失敗しました: %v", err), true)
		return
	}

	// 成功メッセージを送信
	successMsg := "✅ このサーバー用のGemini APIキーを削除しました。\n今後はデフォルトのAPIキーを使用します。"
	h.respondToInteraction(s, i, successMsg, false)
}

// handleSetModelCommand は、/set-modelコマンドを処理します
func (h *SlashCommandHandler) handleSetModelCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// 権限チェック（管理者権限が必要）
	if !h.hasAdminPermission(i.Member) {
		h.respondToInteraction(s, i, "❌ このコマンドを実行するには管理者権限が必要です。", true)
		return
	}

	// モデルを取得
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		h.respondToInteraction(s, i, "❌ 使用するAIモデルが指定されていません。", true)
		return
	}

	model := options[0].StringValue()
	guildID := i.GuildID
	setBy := i.Member.User.Username

	// モデルを設定
	ctx := context.Background()
	err := h.apiKeyService.SetGuildModel(ctx, guildID, model)
	if err != nil {
		log.Printf("モデルの設定に失敗: %v", err)
		h.respondToInteraction(s, i, fmt.Sprintf("❌ モデルの設定に失敗しました: %v", err), true)
		return
	}

	// 成功メッセージを送信
	successMsg := fmt.Sprintf("✅ このサーバーで使用するAIモデルを %s に設定しました。\n設定者: %s", model, setBy)
	h.respondToInteraction(s, i, successMsg, false)
}

// handleStatusCommand は、/statusコマンドを処理します
func (h *SlashCommandHandler) handleStatusCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	ctx := context.Background()

	// APIキーの設定状況を確認
	hasAPIKey, err := h.apiKeyService.HasGuildAPIKey(ctx, guildID)
	if err != nil {
		log.Printf("APIキーの確認に失敗: %v", err)
		h.respondToInteraction(s, i, "❌ 設定状況の確認に失敗しました。", true)
		return
	}

	var statusMessage string

	if hasAPIKey {
		// APIキーが設定されている場合
		apiKeyInfo, err := h.apiKeyService.GetGuildAPIKeyInfo(ctx, guildID)
		if err != nil {
			log.Printf("APIキー情報の取得に失敗: %v", err)
			h.respondToInteraction(s, i, "❌ 設定情報の取得に失敗しました。", true)
			return
		}

		// 設定日時のフォーマット
		setDate := apiKeyInfo.SetAt.Format("2006年1月2日 15:04")

		statusMessage = fmt.Sprintf(`📊 **サーバー設定状況**

✅ **APIキー**: 設定済み
👤 **設定者**: %s
📅 **設定日**: %s
🤖 **使用モデル**: %s`,
			apiKeyInfo.SetBy,
			setDate,
			apiKeyInfo.Model)
	} else {
		// APIキーが未設定の場合
		model, err := h.apiKeyService.GetGuildModel(ctx, guildID)
		if err != nil {
			model = "gemini-pro" // デフォルト
		}

		statusMessage = fmt.Sprintf(`📊 **サーバー設定状況**

❌ **APIキー**: 未設定（デフォルトを使用）
🤖 **使用モデル**: %s（デフォルト）`, model)
	}

	h.respondToInteraction(s, i, statusMessage, false)
}

// hasAdminPermission は、メンバーが管理者権限を持っているかをチェックします
func (h *SlashCommandHandler) hasAdminPermission(member *discordgo.Member) bool {
	if member == nil {
		return false
	}

	// 管理者権限をチェック（Permissionsはint64のビットフラグ）
	return member.Permissions&discordgo.PermissionAdministrator != 0
}

// respondToInteraction は、インタラクションに応答します
func (h *SlashCommandHandler) respondToInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, content string, ephemeral bool) {
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}

	if !ephemeral {
		response.Data.Flags = 0
	}

	err := s.InteractionRespond(i.Interaction, response)
	if err != nil {
		log.Printf("インタラクションへの応答に失敗: %v", err)
	}
}
