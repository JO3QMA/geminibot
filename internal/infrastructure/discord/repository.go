package discord

import (
	"context"
	"fmt"
	"log"

	"geminibot/internal/domain"

	"github.com/bwmarrin/discordgo"
)

// DiscordConversationRepository は、Discord APIを使用してConversationRepositoryインターフェースを実装します
type DiscordConversationRepository struct {
	session *discordgo.Session
}

// NewDiscordConversationRepository は新しいDiscordConversationRepositoryインスタンスを作成します
func NewDiscordConversationRepository(session *discordgo.Session) *DiscordConversationRepository {
	return &DiscordConversationRepository{
		session: session,
	}
}

// GetRecentMessages は、指定されたチャンネルの直近のメッセージを取得します
func (r *DiscordConversationRepository) GetRecentMessages(ctx context.Context, channelID domain.ChannelID, limit int) (domain.ConversationHistory, error) {
	log.Printf("Discordから直近%d件のメッセージを取得中: %s", limit, channelID)

	messages, err := r.session.ChannelMessages(channelID.String(), limit, "", "", "")
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return domain.ConversationHistory{}, fmt.Errorf("Discord APIからのメッセージ取得がタイムアウトしました: %w", err)
		}
		return domain.ConversationHistory{}, fmt.Errorf("Discord APIからメッセージ取得に失敗: %w", err)
	}

	domainMessages := make([]domain.Message, 0, len(messages))
	for _, msg := range messages {
		// Botのメッセージは除外
		if msg.Author.Bot {
			continue
		}

		timestamp := msg.Timestamp

		// ユーザー情報を作成
		user := domain.NewUser(
			domain.NewUserID(msg.Author.ID),
			msg.Author.Username,
			r.getDisplayName(msg),
			msg.Author.Avatar,
			msg.Author.Discriminator,
			msg.Author.Bot,
		)

		domainMessage := domain.NewMessage(
			msg.ID,
			user,
			msg.Content,
			timestamp,
		)
		domainMessages = append(domainMessages, domainMessage)
	}

	return domain.NewConversationHistory(domainMessages), nil
}

// GetThreadMessages は、指定されたスレッドの全メッセージを取得します
func (r *DiscordConversationRepository) GetThreadMessages(ctx context.Context, threadID domain.ChannelID) (domain.ConversationHistory, error) {
	log.Printf("Discordからスレッドの全メッセージを取得中: %s", threadID)

	// スレッドの場合は十分な数のメッセージを取得（コンテキスト長制限で調整される）
	const threadMessageLimit = 200
	return r.GetRecentMessages(ctx, threadID, threadMessageLimit)
}

// GetMessagesBefore は、指定されたメッセージIDより前のメッセージを取得します
func (r *DiscordConversationRepository) GetMessagesBefore(ctx context.Context, channelID domain.ChannelID, messageID string, limit int) (domain.ConversationHistory, error) {
	log.Printf("DiscordからメッセージID %s より前の%d件のメッセージを取得中: %s", messageID, limit, channelID)

	messages, err := r.session.ChannelMessages(channelID.String(), limit, messageID, "", "")
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return domain.ConversationHistory{}, fmt.Errorf("Discord APIからのメッセージ取得がタイムアウトしました: %w", err)
		}
		return domain.ConversationHistory{}, fmt.Errorf("Discord APIからメッセージ取得に失敗: %w", err)
	}

	domainMessages := make([]domain.Message, 0, len(messages))
	for _, msg := range messages {
		// Botのメッセージは除外
		if msg.Author.Bot {
			continue
		}

		timestamp := msg.Timestamp

		// ユーザー情報を作成
		user := domain.NewUser(
			domain.NewUserID(msg.Author.ID),
			msg.Author.Username,
			r.getDisplayName(msg),
			msg.Author.Avatar,
			msg.Author.Discriminator,
			msg.Author.Bot,
		)

		domainMessage := domain.NewMessage(
			msg.ID,
			user,
			msg.Content,
			timestamp,
		)
		domainMessages = append(domainMessages, domainMessage)
	}

	return domain.NewConversationHistory(domainMessages), nil
}

// getDisplayName は、Discordメッセージから表示名を取得します
func (r *DiscordConversationRepository) getDisplayName(msg *discordgo.Message) string {
	// メンバー情報がある場合はニックネームを優先
	if msg.Member != nil && msg.Member.Nick != "" {
		return msg.Member.Nick
	}

	// メンバー情報がない場合は、Discord APIからメンバー情報を取得を試行
	if msg.GuildID != "" {
		member, err := r.session.GuildMember(msg.GuildID, msg.Author.ID)
		if err == nil && member.Nick != "" {
			return member.Nick
		}
	}

	// ニックネームがない場合はユーザー名を使用
	return msg.Author.Username
}
