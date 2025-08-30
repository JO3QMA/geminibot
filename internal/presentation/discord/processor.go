package discord

import (
	"context"
	"fmt"
	"log"
	"strings"

	"geminibot/internal/application"
	"geminibot/internal/domain"

	"github.com/bwmarrin/discordgo"
)

// MessageProcessor は、Discordメッセージの処理を担当します
type MessageProcessor struct {
	mentionService *application.MentionApplicationService
	sender         *MessageSender
	errorHandler   *ErrorHandler
}

// NewMessageProcessor は新しいMessageProcessorインスタンスを作成します
func NewMessageProcessor(
	mentionService *application.MentionApplicationService,
	sender *MessageSender,
	errorHandler *ErrorHandler,
) *MessageProcessor {
	return &MessageProcessor{
		mentionService: mentionService,
		sender:         sender,
		errorHandler:   errorHandler,
	}
}

// ProcessMentionAsync は、メンションを非同期で処理します
func (p *MessageProcessor) ProcessMentionAsync(s *discordgo.Session, m *discordgo.MessageCreate, mention domain.BotMention) {
	// メッセージからスレッドを作成
	thread, err := s.MessageThreadStart(m.ChannelID, m.ID, "Bot応答", 60) // 60分後にアーカイブ
	if err != nil {
		log.Printf("スレッド作成に失敗: %v", err)
			// スレッド作成に失敗した場合は通常のリプライとして送信
	p.ProcessMentionNormal(s, m, mention)
		return
	}

	// 処理中メッセージをスレッド内に送信
	thinkingMsg, err := p.sender.SendThinkingMessageToThread(s, thread.ID)
	if err != nil {
		log.Printf("処理中メッセージの送信に失敗: %v", err)
		return
	}

	// メンションを処理
	ctx := context.Background()
	response, err := p.mentionService.HandleMention(ctx, mention)

	// 処理中メッセージを削除
	if err := s.ChannelMessageDelete(thread.ID, thinkingMsg.ID); err != nil {
		log.Printf("処理中メッセージの削除に失敗: %v", err)
	}

	if err != nil {
		log.Printf("メンション処理に失敗: %v", err)

		// エラーを適切なメッセージにフォーマット
		errorMsg := p.errorHandler.FormatError(err)
		if _, err := s.ChannelMessageSend(thread.ID, errorMsg); err != nil {
			log.Printf("エラーメッセージの送信に失敗: %v", err)
		}
		return
	}

	// 応答をスレッド内に送信
	p.sender.SendThreadResponse(s, thread.ID, response)
}

// ProcessMentionNormal は、通常のリプライとしてメンションを処理します
func (p *MessageProcessor) ProcessMentionNormal(s *discordgo.Session, m *discordgo.MessageCreate, mention domain.BotMention) {
	// 処理中メッセージを送信
	thinkingMsg, err := p.sender.SendThinkingMessage(s, m.ChannelID, m.ID)
	if err != nil {
		log.Printf("処理中メッセージの送信に失敗: %v", err)
		return
	}

	// メンションを処理
	ctx := context.Background()
	response, err := p.mentionService.HandleMention(ctx, mention)

	// 処理中メッセージを削除
	s.ChannelMessageDelete(m.ChannelID, thinkingMsg.ID)

	if err != nil {
		log.Printf("メンション処理に失敗: %v", err)

		// エラーを適切なメッセージにフォーマット
		errorMsg := p.errorHandler.FormatError(err)
		if _, err := s.ChannelMessageSendReply(m.ChannelID, errorMsg, &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		}); err != nil {
			log.Printf("エラーメッセージの送信に失敗: %v", err)
		}
		return
	}

	// 応答を分割して送信
	p.sender.SendSplitResponse(s, m, response)
}

// CreateBotMention は、DiscordメッセージからBotMentionオブジェクトを作成します
func (p *MessageProcessor) CreateBotMention(s *discordgo.Session, m *discordgo.MessageCreate) domain.BotMention {
	// メンション部分を除去したコンテンツを取得
	content := p.extractUserContent(m)

	// ユーザー情報を作成
	user := domain.NewUser(
		domain.NewUserID(m.Author.ID),
		m.Author.Username,
		p.getDisplayName(s, m),
		m.Author.Avatar,
		m.Author.Discriminator,
		m.Author.Bot,
	)

	return domain.NewBotMention(
		domain.NewChannelID(m.ChannelID),
		user,
		content,
		m.ID,
	)
}

// IsMentioned は、メッセージがBotへのメンションかどうかを判定します
func (p *MessageProcessor) IsMentioned(m *discordgo.MessageCreate, botID string, botUsername string) bool {
	// メンション配列をチェック
	for _, mention := range m.Mentions {
		if mention.ID == botID {
			return true
		}
	}

	// メンション配列が空の場合、コンテンツをチェック
	if len(m.Mentions) == 0 {
		content := strings.ToLower(m.Content)
		botMention := fmt.Sprintf("@%s", strings.ToLower(botUsername))
		return strings.Contains(content, botMention)
	}

	return false
}

// extractUserContent は、メンション部分を除去したユーザーのコンテンツを抽出します
func (p *MessageProcessor) extractUserContent(m *discordgo.MessageCreate) string {
	content := m.Content

	// メンション配列がある場合、それらを除去
	for _, mention := range m.Mentions {
		mentionText := fmt.Sprintf("<@%s>", mention.ID)
		content = strings.ReplaceAll(content, mentionText, "")
	}

	// 先頭と末尾の空白を除去
	content = strings.TrimSpace(content)

	return content
}

// getDisplayName は、Discordメッセージから表示名を取得します
func (p *MessageProcessor) getDisplayName(s *discordgo.Session, m *discordgo.MessageCreate) string {
	// メンバー情報がある場合はニックネームを優先
	if m.Member != nil && m.Member.Nick != "" {
		return m.Member.Nick
	}

	// メンバー情報がない場合は、Discord APIからメンバー情報を取得を試行
	if m.GuildID != "" {
		member, err := s.GuildMember(m.GuildID, m.Author.ID)
		if err == nil && member.Nick != "" {
			return member.Nick
		}
	}

	// ニックネームがない場合はユーザー名を使用
	return m.Author.Username
}
