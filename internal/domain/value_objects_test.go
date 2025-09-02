package domain

import (
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	user := User{
		ID:            "123",
		Username:      "testuser",
		DisplayName:   "Test User",
		Avatar:        "avatar.jpg",
		IsBot:         false,
		Discriminator: "1234",
	}

	now := time.Now()
	message := Message{ID: "msg123", User: user, Content: "Hello, World!", Timestamp: now}

	if message.ID != "msg123" {
		t.Errorf("Expected ID 'msg123', got '%s'", message.ID)
	}

	if message.User.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", message.User.Username)
	}

	if message.Content != "Hello, World!" {
		t.Errorf("Expected Content 'Hello, World!', got '%s'", message.Content)
	}

	if !message.Timestamp.Equal(now) {
		t.Errorf("Expected Timestamp %v, got %v", now, message.Timestamp)
	}
}

func TestNewConversationHistory(t *testing.T) {
	messages := []Message{
		Message{ID: "1", User: User{Username: "user1"}, Content: "Hello", Timestamp: time.Now()},
		Message{ID: "2", User: User{Username: "user2"}, Content: "Hi there", Timestamp: time.Now()},
	}

	history := NewConversationHistory(messages)

	if history.Count() != 2 {
		t.Errorf("Expected count 2, got %d", history.Count())
	}

	if history.IsEmpty() {
		t.Error("Expected history to not be empty")
	}

	if len(history.Messages()) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(history.Messages()))
	}
}

func TestConversationHistoryEmpty(t *testing.T) {
	history := NewConversationHistory([]Message{})

	if !history.IsEmpty() {
		t.Error("Expected history to be empty")
	}

	if history.Count() != 0 {
		t.Errorf("Expected count 0, got %d", history.Count())
	}
}

func TestNewPrompt(t *testing.T) {
	content := "This is a test prompt"
	prompt := NewPrompt(content)

	if prompt.Content() != content {
		t.Errorf("Expected content '%s', got '%s'", content, prompt.Content())
	}

	if prompt.String() != content {
		t.Errorf("Expected string '%s', got '%s'", content, prompt.String())
	}
}

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
