package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestNewDiscordHandler(t *testing.T) {
	session := &discordgo.Session{}
	botID := "bot123"

	handler := NewDiscordHandler(session, nil, botID)

	if handler.session != session {
		t.Error("セッションが正しく設定されていません")
	}
	if handler.botID != botID {
		t.Error("BotIDが正しく設定されていません")
	}
}

func TestDiscordHandler_IsMentioned_WithMentions(t *testing.T) {
	handler := &DiscordHandler{
		botID:       "bot123",
		botUsername: "TestBot",
	}

	// メンション配列がある場合
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content: "<@bot123> こんにちは",
			Mentions: []*discordgo.User{
				{ID: "bot123"},
			},
		},
	}

	if !handler.isMentioned(message) {
		t.Error("メンション配列での判定が失敗しました")
	}
}

func TestDiscordHandler_IsMentioned_WithUsername(t *testing.T) {
	handler := &DiscordHandler{
		botID:       "bot123",
		botUsername: "TestBot",
	}

	// メンション配列が空で、ユーザー名でのメンション
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content:  "@testbot こんにちは",
			Mentions: []*discordgo.User{},
		},
	}

	if !handler.isMentioned(message) {
		t.Error("ユーザー名でのメンション判定が失敗しました")
	}
}

func TestDiscordHandler_IsMentioned_NotMentioned(t *testing.T) {
	handler := &DiscordHandler{
		botID:       "bot123",
		botUsername: "TestBot",
	}

	// メンションされていない場合
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content:  "こんにちは",
			Mentions: []*discordgo.User{},
		},
	}

	if handler.isMentioned(message) {
		t.Error("メンションされていないのに判定されました")
	}
}

func TestDiscordHandler_IsMentioned_DifferentBot(t *testing.T) {
	handler := &DiscordHandler{
		botID:       "bot123",
		botUsername: "TestBot",
	}

	// 別のBotへのメンション
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content: "<@otherbot> こんにちは",
			Mentions: []*discordgo.User{
				{ID: "otherbot"},
			},
		},
	}

	if handler.isMentioned(message) {
		t.Error("別のBotへのメンションが誤って判定されました")
	}
}

func TestDiscordHandler_ExtractUserContent_WithMention(t *testing.T) {
	handler := &DiscordHandler{
		botID:       "bot123",
		botUsername: "TestBot",
	}

	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content: "<@bot123> こんにちは、お元気ですか？",
			Mentions: []*discordgo.User{
				{ID: "bot123"},
			},
		},
	}

	content := handler.extractUserContent(message)
	expected := "こんにちは、お元気ですか？"

	if content != expected {
		t.Errorf("期待されるコンテンツ: %s, 実際: %s", expected, content)
	}
}

func TestDiscordHandler_ExtractUserContent_WithoutMention(t *testing.T) {
	handler := &DiscordHandler{
		botID:       "bot123",
		botUsername: "TestBot",
	}

	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content:  "こんにちは",
			Mentions: []*discordgo.User{},
		},
	}

	content := handler.extractUserContent(message)
	expected := "こんにちは"

	if content != expected {
		t.Errorf("期待されるコンテンツ: %s, 実際: %s", expected, content)
	}
}

func TestDiscordHandler_ExtractUserContent_WithSpaces(t *testing.T) {
	handler := &DiscordHandler{
		botID:       "bot123",
		botUsername: "TestBot",
	}

	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content: "  <@bot123>  こんにちは  ",
			Mentions: []*discordgo.User{
				{ID: "bot123"},
			},
		},
	}

	content := handler.extractUserContent(message)
	expected := "こんにちは"

	if content != expected {
		t.Errorf("期待されるコンテンツ: %s, 実際: %s", expected, content)
	}
}

func TestDiscordHandler_HandleReady(t *testing.T) {
	handler := &DiscordHandler{
		botID: "bot123",
	}

	event := &discordgo.Ready{
		User: &discordgo.User{
			Username:    "TestBot",
			Discriminator: "1234",
		},
	}

	handler.handleReady(nil, event)

	if handler.botUsername != "TestBot" {
		t.Errorf("期待されるBotUsername: TestBot, 実際: %s", handler.botUsername)
	}
}
