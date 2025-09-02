package domain

import (
	"testing"
)

func TestNewBotMention(t *testing.T) {
	channelID := "channel123"
	user := User{Username: "testuser"}

	mention := BotMention{
		ChannelID: channelID,
		GuildID:   "guild123",
		User:      user,
		Content:   "Hello bot!",
		MessageID: "msg123",
	}

	if mention.ChannelID != channelID {
		t.Errorf("Expected ChannelID %v, got %v", channelID, mention.ChannelID)
	}

	if mention.GuildID != "guild123" {
		t.Errorf("Expected GuildID 'guild123', got '%s'", mention.GuildID)
	}

	if mention.User.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", mention.User.Username)
	}

	if mention.Content != "Hello bot!" {
		t.Errorf("Expected Content 'Hello bot!', got '%s'", mention.Content)
	}

	if mention.MessageID != "msg123" {
		t.Errorf("Expected MessageID 'msg123', got '%s'", mention.MessageID)
	}
}

func TestBotMentionString(t *testing.T) {
	channelID := "channel123"
	user := User{Username: "testuser"}

	mention := BotMention{
		ChannelID: channelID,
		GuildID:   "guild123",
		User:      user,
		Content:   "Hello bot!",
		MessageID: "msg123",
	}

	expected := "BotMention{ChannelID: channel123, GuildID: guild123, User: testuser, Content: Hello bot!, MessageID: msg123}"
	if mention.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, mention.String())
	}
}

func TestChannelID(t *testing.T) {
	channelID := "channel123"

	if channelID != "channel123" {
		t.Errorf("Expected 'channel123', got '%s'", channelID)
	}
}
