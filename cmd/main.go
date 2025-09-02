package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"geminibot/configs"
	"geminibot/internal/application"
	discordInfra "geminibot/internal/infrastructure/discord"
	"geminibot/internal/infrastructure/gemini"
	discordPres "geminibot/internal/presentation/discord"

	"github.com/bwmarrin/discordgo"
)

func main() {
	log.Println("Discord-Gemini連携Botを起動中...")

	// 設定を読み込み
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗: %v", err)
	}

	// Discordセッションを作成
	session, err := discordgo.New("Bot " + config.Discord.BotToken)
	if err != nil {
		log.Fatalf("Discordセッションの作成に失敗: %v", err)
	}
	defer session.Close()

	// Botの情報を取得
	user, err := session.User("@me")
	if err != nil {
		log.Fatalf("Bot情報の取得に失敗: %v", err)
	}

	log.Printf("Bot情報: %s#%s (ID: %s)", user.Username, user.Discriminator, user.ID)

	// Gemini APIクライアントを作成
	geminiClient, err := gemini.NewGeminiAPIClient(config.Gemini.APIKey, &config.Gemini)
	if err != nil {
		log.Fatalf("Gemini APIクライアントの作成に失敗: %v", err)
	}
	defer geminiClient.Close()

	// リポジトリを作成
	conversationRepo := discordInfra.NewDiscordConversationRepository(session)
	apiKeyRepo := discordInfra.NewDiscordGuildAPIKeyRepository()

	// アプリケーションサービスを作成
	mentionService := application.NewMentionApplicationService(
		conversationRepo,
		geminiClient,
		&config.Bot,
	)
	apiKeyService := application.NewAPIKeyApplicationService(apiKeyRepo)

	// Discordハンドラを作成
	handler := discordPres.NewDiscordHandler(session, mentionService, user.ID)
	handler.SetupHandlers()

	// スラッシュコマンドハンドラを作成
	slashCommandHandler := discordPres.NewSlashCommandHandler(session, apiKeyService, config.Gemini.APIKey)
	
	// スラッシュコマンドを設定
	if err := slashCommandHandler.SetupSlashCommands(); err != nil {
		log.Fatalf("スラッシュコマンドの設定に失敗: %v", err)
	}
	
	// スラッシュコマンドのイベントハンドラーを設定
	slashCommandHandler.SetupSlashCommandHandlers()

	// Discordに接続
	err = session.Open()
	if err != nil {
		log.Fatalf("Discordへの接続に失敗: %v", err)
	}

	log.Println("Discordに接続しました。Botが準備完了しました！")
	log.Println("利用可能なスラッシュコマンド:")
	log.Println("  /set-api - このサーバー用のGemini APIキーを設定")
	log.Println("  /del-api - このサーバー用のGemini APIキーを削除")

	// シグナルハンドリング
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// 終了シグナルを待機
	<-stop
	log.Println("終了シグナルを受信しました。Botを停止中...")

	// クリーンアップ
	if err := session.Close(); err != nil {
		log.Printf("Discordセッションのクローズに失敗: %v", err)
	}

	log.Println("Botが正常に停止しました。")
}
