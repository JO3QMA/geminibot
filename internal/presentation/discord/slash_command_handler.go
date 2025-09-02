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
	session           *discordgo.Session
	apiKeyService     *application.APIKeyApplicationService
	defaultAPIKey     string
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

// SetupSlashCommandHandlers は、スラッシュコマンドのイベントハンドラーを設定します
func (h *SlashCommandHandler) SetupSlashCommandHandlers() {
	h.session.AddHandler(h.handleInteractionCreate)
}

// handleInteractionCreate は、インタラクション作成イベントを処理します
func (h *SlashCommandHandler) handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	// コマンド名に基づいて処理を分岐
	switch i.ApplicationCommandData().Name {
	case "set-api":
		h.handleSetAPICommand(s, i)
	case "del-api":
		h.handleDelAPICommand(s, i)
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

	// APIキーが設定されているかチェック
	ctx := context.Background()
	hasAPIKey, err := h.apiKeyService.HasGuildAPIKey(ctx, guildID)
	if err != nil {
		log.Printf("APIキーの存在確認に失敗: %v", err)
		h.respondToInteraction(s, i, "❌ APIキーの確認に失敗しました。", true)
		return
	}

	if !hasAPIKey {
		h.respondToInteraction(s, i, "❌ このサーバーにはAPIキーが設定されていません。", true)
		return
	}

	// APIキーを削除
	err = h.apiKeyService.DeleteGuildAPIKey(ctx, guildID)
	if err != nil {
		log.Printf("APIキーの削除に失敗: %v", err)
		h.respondToInteraction(s, i, fmt.Sprintf("❌ APIキーの削除に失敗しました: %v", err), true)
		return
	}

	// 成功メッセージを送信
	successMsg := "✅ このサーバー用のGemini APIキーを削除しました。\nデフォルトのAPIキーが使用されます。"
	h.respondToInteraction(s, i, successMsg, false)
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
			Content:  content,
			Flags:    discordgo.MessageFlagsEphemeral,
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
