package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	// .envファイルを読み込み
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: .envファイルの読み込みに失敗しました: %v", err)
	}

	// Bot Tokenを取得
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN が設定されていません")
	}

	// Discordセッションを作成
	session, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Discordセッションの作成に失敗: %v", err)
	}
	defer session.Close()

	// Botの情報を取得
	user, err := session.User("@me")
	if err != nil {
		log.Fatalf("Bot情報の取得に失敗: %v", err)
	}

	fmt.Printf("🤖 Bot情報:\n")
	fmt.Printf("   名前: %s#%s\n", user.Username, user.Discriminator)
	fmt.Printf("   ID: %s\n", user.ID)
	fmt.Printf("   Client ID: %s\n", user.ID)
	fmt.Println()

	// 招待URLを生成
	inviteURL := fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&permissions=66560&scope=bot", user.ID)

	fmt.Printf("🔗 Bot招待URL:\n")
	fmt.Printf("   %s\n", inviteURL)
	fmt.Println()

	fmt.Printf("📋 必要な権限:\n")
	fmt.Printf("   - Read Messages/View Channels (1024)\n")
	fmt.Printf("   - Send Messages (2048)\n")
	fmt.Printf("   - Read Message History (65536)\n")
	fmt.Printf("   - 合計: 66560\n")
	fmt.Println()

	fmt.Printf("💡 使用方法:\n")
	fmt.Printf("   1. 上記のURLをクリックまたはコピー\n")
	fmt.Printf("   2. 招待したいDiscordサーバーを選択\n")
	fmt.Printf("   3. 権限を確認して「承認」をクリック\n")
	fmt.Printf("   4. Botがサーバーに参加します\n")
	fmt.Println()

	fmt.Printf("🎯 Botの使い方:\n")
	fmt.Printf("   1. チャンネルでBotをメンション: @%s こんにちは\n", user.Username)
	fmt.Printf("   2. Botがチャット履歴を考慮してGemini AIで応答を生成\n")
	fmt.Printf("   3. スレッド内でも同様に使用可能\n")
}
