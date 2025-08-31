package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"geminibot/configs"
	"geminibot/internal/infrastructure/container"

	"github.com/bwmarrin/discordgo"
)

func main() {
	log.Println("Discord-Gemini連携Botを起動中...")

	// 設定を読み込み
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗: %v", err)
	}

	// 依存性注入コンテナを作成
	container, err := container.NewContainer(config)
	if err != nil {
		log.Fatalf("コンテナの初期化に失敗: %v", err)
	}

	// クリーンアップを確実に実行
	defer func() {
		if err := container.Close(); err != nil {
			log.Printf("コンテナのクローズに失敗: %v", err)
		}
	}()

	// Discordセッションを取得
	session := container.GetDiscordSession()
	handler := container.GetDiscordHandler()

	// ハンドラーを設定
	handler.SetupHandlers()

	// Discordに接続
	if err := session.Open(); err != nil {
		log.Fatalf("Discordへの接続に失敗: %v", err)
	}

	log.Println("Discordに接続しました。Botが準備完了しました！")

	// シグナルハンドリング
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// 終了シグナルを待機
	<-stop
	log.Println("終了シグナルを受信しました。Botを停止中...")

	// グレースフルシャットダウン
	if err := gracefulShutdown(session); err != nil {
		log.Printf("グレースフルシャットダウンに失敗: %v", err)
	}

	log.Println("Botが正常に停止しました。")
}

// gracefulShutdown はグレースフルシャットダウンを実行します
func gracefulShutdown(session *discordgo.Session) error {
	// コンテキストにタイムアウトを設定
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Discordセッションをクローズ
	if err := session.Close(); err != nil {
		return err
	}

	// 残りの処理を待機
	select {
	case <-ctx.Done():
		log.Println("シャットダウンがタイムアウトしました")
		return ctx.Err()
	default:
		return nil
	}
}
