package domain

import (
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	user := NewUser(NewUserID("123456789"), "user123", "User123", "", "", false)
	timestamp := time.Now()
	content := "テストメッセージ"

	msg := NewMessage("msg123", user, content, timestamp)

	if msg.ID != "msg123" {
		t.Errorf("期待されるID: msg123, 実際のID: %s", msg.ID)
	}
	if msg.User.ID != user.ID {
		t.Errorf("期待されるUserID: %v, 実際のUserID: %v", user.ID, msg.User.ID)
	}
	if msg.Content != content {
		t.Errorf("期待されるContent: %s, 実際のContent: %s", content, msg.Content)
	}
	if msg.Timestamp != timestamp {
		t.Errorf("期待されるTimestamp: %v, 実際のTimestamp: %v", timestamp, msg.Timestamp)
	}
}

func TestNewUserID(t *testing.T) {
	id := "123456789"
	userID := NewUserID(id)

	if userID.String() != id {
		t.Errorf("期待されるID: %s, 実際のID: %s", id, userID.String())
	}
}

func TestNewChannelID(t *testing.T) {
	id := "987654321"
	channelID := NewChannelID(id)

	if channelID.String() != id {
		t.Errorf("期待されるID: %s, 実際のID: %s", id, channelID.String())
	}
}

func TestNewConversationHistory(t *testing.T) {
	messages := []Message{
		NewMessage("msg1", NewUser(NewUserID("user1"), "user1", "User1", "", "", false), "こんにちは", time.Now()),
		NewMessage("msg2", NewUser(NewUserID("user2"), "user2", "User2", "", "", false), "こんばんは", time.Now()),
	}

	history := NewConversationHistory(messages)

	if history.Count() != 2 {
		t.Errorf("期待されるメッセージ数: 2, 実際のメッセージ数: %d", history.Count())
	}

	if history.IsEmpty() {
		t.Error("履歴が空と判定されましたが、実際は空ではありません")
	}

	retrievedMessages := history.Messages()
	if len(retrievedMessages) != 2 {
		t.Errorf("期待されるメッセージ数: 2, 実際のメッセージ数: %d", len(retrievedMessages))
	}
}

func TestConversationHistory_IsEmpty(t *testing.T) {
	// 空の履歴
	emptyHistory := NewConversationHistory([]Message{})
	if !emptyHistory.IsEmpty() {
		t.Error("空の履歴が空と判定されませんでした")
	}

	// メッセージがある履歴
	messages := []Message{
		NewMessage("msg1", NewUser(NewUserID("user1"), "user1", "User1", "", "", false), "テスト", time.Now()),
	}
	nonEmptyHistory := NewConversationHistory(messages)
	if nonEmptyHistory.IsEmpty() {
		t.Error("メッセージがある履歴が空と判定されました")
	}
}

func TestNewPrompt(t *testing.T) {
	content := "これはテストプロンプトです"
	prompt := NewPrompt(content)

	if prompt.Content() != content {
		t.Errorf("期待されるContent: %s, 実際のContent: %s", content, prompt.Content())
	}

	if prompt.String() != content {
		t.Errorf("期待されるString: %s, 実際のString: %s", content, prompt.String())
	}
}

func TestNewBotMention(t *testing.T) {
	channelID := NewChannelID("channel123")
	user := NewUser(NewUserID("user123"), "user123", "User123", "", "", false)
	content := "@bot こんにちは"
	messageID := "msg123"

	mention := NewBotMention(channelID, user, content, messageID)

	if mention.ChannelID != channelID {
		t.Errorf("期待されるChannelID: %v, 実際のChannelID: %v", channelID, mention.ChannelID)
	}
	if mention.User.ID != user.ID {
		t.Errorf("期待されるUserID: %v, 実際のUserID: %v", user.ID, mention.User.ID)
	}
	if mention.Content != content {
		t.Errorf("期待されるContent: %s, 実際のContent: %s", content, mention.Content)
	}
	if mention.MessageID != messageID {
		t.Errorf("期待されるMessageID: %s, 実際のMessageID: %s", messageID, mention.MessageID)
	}
}

func TestBotMention_IsThread(t *testing.T) {
	channelID := NewChannelID("channel123")
	user := NewUser(NewUserID("user123"), "user123", "User123", "", "", false)
	content := "テスト"
	messageID := "msg123"

	mention := NewBotMention(channelID, user, content, messageID)

	// 現在の実装では常にfalseを返す
	if mention.IsThread() {
		t.Error("通常のチャンネルがスレッドと判定されました")
	}
}

func TestBotMention_String(t *testing.T) {
	channelID := NewChannelID("channel123")
	user := NewUser(NewUserID("user123"), "user123", "User123", "", "", false)
	content := "テストメッセージ"
	messageID := "msg123"

	mention := NewBotMention(channelID, user, content, messageID)
	expected := "BotMention{ChannelID: channel123, UserID: user123, Content: テストメッセージ, MessageID: msg123}"

	if mention.String() != expected {
		t.Errorf("期待されるString: %s, 実際のString: %s", expected, mention.String())
	}
}
