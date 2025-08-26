package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestNewDiscordConversationRepository(t *testing.T) {
	session := &discordgo.Session{}
	repo := NewDiscordConversationRepository(session)

	if repo.session != session {
		t.Error("セッションが正しく設定されていません")
	}
}
