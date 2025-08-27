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
	geminiConfig := &gemini.Config{
		APIKey:      config.Gemini.APIKey,
		ModelName:   config.Gemini.ModelName,
		MaxTokens:   config.Gemini.MaxTokens,
		Temperature: config.Gemini.Temperature,
		TopP:        config.Gemini.TopP,
		TopK:        config.Gemini.TopK,
	}

	geminiClient, err := gemini.NewGeminiAPIClient(config.Gemini.APIKey, geminiConfig)
	if err != nil {
		log.Fatalf("Gemini APIクライアントの作成に失敗: %v", err)
	}
	defer geminiClient.Close()

	// リポジトリを作成
	conversationRepo := discordInfra.NewDiscordConversationRepository(session)

	// アプリケーションサービスの設定を作成
	appConfig := &application.Config{
		MaxContextLength:     config.Bot.MaxContextLength,
		MaxHistoryLength:     config.Bot.MaxHistoryLength,
		RequestTimeout:       config.Bot.RequestTimeout,
		SystemPrompt:         config.Bot.SystemPrompt,
		UseStructuredContext: config.Bot.UseStructuredContext,
	}

	// アプリケーションサービスを作成
	mentionService := application.NewMentionApplicationService(
		conversationRepo,
		geminiClient,
		appConfig,
	)

	// Discordハンドラを作成
	handler := discordPres.NewDiscordHandler(session, mentionService, user.ID)
	handler.SetupHandlers()

	// Discordに接続
	err = session.Open()
	if err != nil {
		log.Fatalf("Discordへの接続に失敗: %v", err)
	}

	log.Println("Discordに接続しました。Botが準備完了しました！")

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
