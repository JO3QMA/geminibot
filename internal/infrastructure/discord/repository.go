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
		return domain.ConversationHistory{}, fmt.Errorf("Discord APIからメッセージ取得に失敗: %w", err)
	}

	domainMessages := make([]domain.Message, 0, len(messages))
	for _, msg := range messages {
		// Botのメッセージは除外
		if msg.Author.Bot {
			continue
		}

		timestamp := msg.Timestamp

		// 表示名を取得（ニックネームがある場合はニックネーム、ない場合はユーザー名）
		displayName := r.getDisplayName(msg)

		domainMessage := domain.NewMessage(
			msg.ID,
			domain.NewUserID(msg.Author.ID),
			displayName,
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

	// Discord APIでは、スレッドのメッセージを取得する際に特別な処理が必要
	// 実際の実装では、スレッドの特性に応じて適切に実装する必要があります

	// 仮の実装：通常のチャンネルメッセージ取得と同じ処理
	// 実際には、スレッドの場合は異なるAPIエンドポイントを使用する必要があります
	return r.GetRecentMessages(ctx, threadID, 100) // スレッドの場合はより多くのメッセージを取得
}

// GetMessagesBefore は、指定されたメッセージIDより前のメッセージを取得します
func (r *DiscordConversationRepository) GetMessagesBefore(ctx context.Context, channelID domain.ChannelID, messageID string, limit int) (domain.ConversationHistory, error) {
	log.Printf("DiscordからメッセージID %s より前の%d件のメッセージを取得中: %s", messageID, limit, channelID)

	messages, err := r.session.ChannelMessages(channelID.String(), limit, messageID, "", "")
	if err != nil {
		return domain.ConversationHistory{}, fmt.Errorf("Discord APIからメッセージ取得に失敗: %w", err)
	}

	domainMessages := make([]domain.Message, 0, len(messages))
	for _, msg := range messages {
		// Botのメッセージは除外
		if msg.Author.Bot {
			continue
		}

		timestamp := msg.Timestamp

		// 表示名を取得（ニックネームがある場合はニックネーム、ない場合はユーザー名）
		displayName := r.getDisplayName(msg)

		domainMessage := domain.NewMessage(
			msg.ID,
			domain.NewUserID(msg.Author.ID),
			displayName,
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
