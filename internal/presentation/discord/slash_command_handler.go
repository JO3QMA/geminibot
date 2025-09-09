package discord

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"geminibot/internal/application"
	"geminibot/internal/domain"
	"geminibot/internal/infrastructure/config"
	"geminibot/internal/infrastructure/gemini"

	"github.com/bwmarrin/discordgo"
)

// SlashCommandHandler は、Discordのスラッシュコマンドを処理するハンドラーです
type SlashCommandHandler struct {
	session             *discordgo.Session
	apiKeyService       *application.APIKeyApplicationService
	defaultAPIKey       string
	defaultGeminiConfig *config.GeminiConfig
}

// NewSlashCommandHandler は新しいSlashCommandHandlerインスタンスを作成します
func NewSlashCommandHandler(
	session *discordgo.Session,
	apiKeyService *application.APIKeyApplicationService,
	defaultAPIKey string,
	defaultGeminiConfig *config.GeminiConfig,
) *SlashCommandHandler {
	return &SlashCommandHandler{
		session:             session,
		apiKeyService:       apiKeyService,
		defaultAPIKey:       defaultAPIKey,
		defaultGeminiConfig: defaultGeminiConfig,
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
						{Name: "Gemini 2.5 Pro", Value: "gemini-2.5-pro"},
						{Name: "Gemini 2.0 Flash", Value: "gemini-2.0-flash"},
						{Name: "Gemini 2.5 Flash Lite", Value: "gemini-2.5-flash-lite"},
					},
				},
			},
		},
		{
			Name:        "status",
			Description: "このサーバーのGemini APIキー設定状況を表示します",
		},
		{
			Name:        "generate-image",
			Description: "Nano Bananaを使って画像を生成します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "prompt",
					Description: "画像生成用のプロンプト",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "style",
					Description: "画像のスタイル",
					Required:    false,
					Choices: func() []*discordgo.ApplicationCommandOptionChoice {
						styles := domain.AllImageStyles()
						choices := make([]*discordgo.ApplicationCommandOptionChoice, len(styles))
						for i, style := range styles {
							choices[i] = &discordgo.ApplicationCommandOptionChoice{
								Name:  style.DisplayName(),
								Value: style.String(),
							}
						}
						return choices
					}(),
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "quality",
					Description: "画像の品質",
					Required:    false,
					Choices: func() []*discordgo.ApplicationCommandOptionChoice {
						qualities := domain.AllImageQualities()
						choices := make([]*discordgo.ApplicationCommandOptionChoice, len(qualities))
						for i, quality := range qualities {
							choices[i] = &discordgo.ApplicationCommandOptionChoice{
								Name:  quality.DisplayName(),
								Value: quality.String(),
							}
						}
						return choices
					}(),
				},
			},
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
	case "generate-image":
		h.handleGenerateImageCommand(s, i)
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
			model = "gemini-2.5-pro" // デフォルト
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

// handleGenerateImageCommand は、/generate-imageコマンドを処理します
func (h *SlashCommandHandler) handleGenerateImageCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// まず処理中メッセージを送信
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("画像生成コマンドの応答に失敗: %v", err)
		return
	}

	// オプションを取得
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		h.followUpInteraction(s, i, "❌ プロンプトが指定されていません。", true)
		return
	}

	request := domain.ImageGenerationRequest{}
	// 画像生成オプションを作成（設定ファイルの値をベースに、ユーザー指定の値を上書き）
	request.Options = domain.DefaultImageGenerationOptions()

	for _, option := range options {
		switch option.Name {
		case "prompt":
			request.Prompt = option.StringValue()
		case "style":
			request.Options.Style = option.StringValue()
		case "quality":
			request.Options.Quality = option.StringValue()
		}
	}

	if request.Prompt == "" {
		h.followUpInteraction(s, i, "❌ プロンプトが指定されていません。", true)
		return
	}

	// APIキーを取得（ギルド固有のAPIキーがない場合はデフォルトを使用）
	ctx := context.Background()
	var apiKey string

	// ギルド固有のAPIキーがあるかチェック
	hasCustomAPIKey, err := h.apiKeyService.HasGuildAPIKey(ctx, i.GuildID)
	if err != nil {
		log.Printf("ギルド %s のAPIキー確認に失敗: %v, デフォルトのAPIキーを使用", i.GuildID, err)
		apiKey = h.defaultAPIKey
	} else if hasCustomAPIKey {
		// カスタムAPIキーを取得
		customAPIKey, err := h.apiKeyService.GetGuildAPIKey(ctx, i.GuildID)
		if err != nil {
			log.Printf("ギルド %s のカスタムAPIキー取得に失敗: %v, デフォルトのAPIキーを使用", i.GuildID, err)
			apiKey = h.defaultAPIKey
		} else {
			apiKey = customAPIKey
			log.Printf("ギルド %s 用のカスタムAPIキーを使用", i.GuildID)
		}
	} else {
		// デフォルトのAPIキーを使用
		apiKey = h.defaultAPIKey
		log.Printf("ギルド %s のAPIキーが設定されていないため、デフォルトのAPIキーを使用", i.GuildID)
	}

	// Geminiクライアントを作成
	geminiClient, err := gemini.NewStructuredGeminiClientWithAPIKey(apiKey, h.defaultGeminiConfig)
	if err != nil {
		log.Printf("Geminiクライアントの作成に失敗: %v", err)
		h.followUpInteraction(s, i, "❌ Gemini APIクライアントの作成に失敗しました。", true)
		return
	}

	// 画像を生成
	response, err := geminiClient.GenerateImage(ctx, request)
	if err != nil {
		log.Printf("画像生成に失敗: %v", err)
		h.followUpInteraction(s, i, fmt.Sprintf("❌ 画像生成に失敗しました: %v", err), true)
		return
	}

	if len(response.Images) == 0 {
		h.followUpInteraction(s, i, "❌ 画像が生成されませんでした。", true)
		return
	}

	// 生成された画像をDiscordに送信
	image := response.Images[0]
	file := &discordgo.File{
		Name:        image.Filename,
		ContentType: image.MimeType,
		Reader:      bytes.NewReader(image.Data),
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🎨 画像生成完了",
		Description: fmt.Sprintf("**プロンプト:** %s\n**スタイル:** %s\n**品質:** %s", request.Prompt, request.Options.Style, request.Options.Quality),
		Color:       0x00ff00,
		Timestamp:   response.GeneratedAt.Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("生成者: %s | モデル: %s", i.Member.User.Username, response.Model),
		},
	}

	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
		Files:  []*discordgo.File{file},
	})
	if err != nil {
		log.Printf("画像の送信に失敗: %v", err)
		h.followUpInteraction(s, i, "❌ 画像の送信に失敗しました。", true)
		return
	}
}

// followUpInteraction は、フォローアップメッセージを送信します
func (h *SlashCommandHandler) followUpInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, content string, ephemeral bool) {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: content,
		Flags:   flags,
	})
	if err != nil {
		log.Printf("フォローアップメッセージの送信に失敗: %v", err)
	}
}
