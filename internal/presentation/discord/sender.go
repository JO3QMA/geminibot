package discord

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// MessageSender は、Discordメッセージの送信を担当します
type MessageSender struct {
	formatter *MessageFormatter
}

// NewMessageSender は新しいMessageSenderインスタンスを作成します
func NewMessageSender(formatter *MessageFormatter) *MessageSender {
	return &MessageSender{
		formatter: formatter,
	}
}

// SendThreadResponse は、スレッド内に応答を送信します
func (s *MessageSender) SendThreadResponse(session *discordgo.Session, threadID string, response string) {
	// 応答をDiscord用にフォーマット
	formattedResponse := s.formatter.FormatForDiscord(response)

	// 応答が非常に長い場合はファイルとして送信
	if len(formattedResponse) > DiscordMessageLimit*5 {
		s.sendAsFileToThread(session, threadID, formattedResponse, "response.txt")
		return
	}

	// 応答をDiscordの制限に合わせて分割
	chunks := s.splitMessage(formattedResponse)

	// すべてのチャンクをスレッド内に送信
	for i, chunk := range chunks {
		_, err := session.ChannelMessageSend(threadID, chunk)
		if err != nil {
			log.Printf("スレッド内メッセージの送信に失敗 (チャンク %d): %v", i+1, err)
			break
		}
	}
}

// SendSplitResponse は、長い応答を複数のメッセージに分割して送信します
func (s *MessageSender) SendSplitResponse(session *discordgo.Session, m *discordgo.MessageCreate, response string) {
	// 応答をDiscord用にフォーマット
	formattedResponse := s.formatter.FormatForDiscord(response)

	// 応答が非常に長い場合はファイルとして送信
	if len(formattedResponse) > DiscordMessageLimit*5 {
		s.sendAsFile(session, m, formattedResponse, "response.txt")
		return
	}

	// 応答をDiscordの制限に合わせて分割
	chunks := s.splitMessage(formattedResponse)

	if len(chunks) == 1 {
		// 単一メッセージの場合
		_, err := session.ChannelMessageSendReply(m.ChannelID, chunks[0], &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})
		if err != nil {
			log.Printf("応答メッセージの送信に失敗: %v", err)
		}
		return
	}

	// 複数メッセージの場合 - すべてスレッド返信として送信
	for i, chunk := range chunks {
		_, err := session.ChannelMessageSendReply(m.ChannelID, chunk, &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})

		if err != nil {
			log.Printf("応答メッセージの送信に失敗 (チャンク %d): %v", i+1, err)
			break
		}
	}
}

// SendNormalReply は、通常のリプライ送信を行います
func (s *MessageSender) SendNormalReply(session *discordgo.Session, m *discordgo.MessageCreate, response string) {
	// 応答をDiscord用にフォーマット
	formattedResponse := s.formatter.FormatForDiscord(response)

	// 応答が非常に長い場合はファイルとして送信
	if len(formattedResponse) > DiscordMessageLimit*5 {
		s.sendAsFile(session, m, formattedResponse, "response.txt")
		return
	}

	// 応答をDiscordの制限に合わせて分割
	chunks := s.splitMessage(formattedResponse)

	if len(chunks) == 1 {
		// 単一メッセージの場合
		_, err := session.ChannelMessageSendReply(m.ChannelID, chunks[0], &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})
		if err != nil {
			log.Printf("応答メッセージの送信に失敗: %v", err)
		}
		return
	}

	// 複数メッセージの場合 - すべてスレッド返信として送信
	for i, chunk := range chunks {
		_, err := session.ChannelMessageSendReply(m.ChannelID, chunk, &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})

		if err != nil {
			log.Printf("応答メッセージの送信に失敗 (チャンク %d): %v", i+1, err)
			break
		}
	}
}

// SendThinkingMessage は、処理中メッセージを送信します
func (s *MessageSender) SendThinkingMessage(session *discordgo.Session, channelID string, messageID string) (*discordgo.Message, error) {
	return session.ChannelMessageSendReply(channelID, "🤔 考え中...", &discordgo.MessageReference{
		MessageID: messageID,
		ChannelID: channelID,
	})
}

// SendThinkingMessageToThread は、スレッド内に処理中メッセージを送信します
func (s *MessageSender) SendThinkingMessageToThread(session *discordgo.Session, threadID string) (*discordgo.Message, error) {
	return session.ChannelMessageSend(threadID, "🤔 考え中...")
}

// sendAsFileToThread は、長い応答をファイルとしてスレッド内に送信します
func (s *MessageSender) sendAsFileToThread(session *discordgo.Session, threadID string, content, filename string) {
	// ファイルデータを作成
	fileData := strings.NewReader(content)

	// ファイルを添付してメッセージを送信
	_, err := session.ChannelFileSend(threadID, filename, fileData)

	if err != nil {
		log.Printf("ファイル送信に失敗: %v", err)
		// ファイル送信に失敗した場合は通常の分割送信にフォールバック
		s.SendThreadResponse(session, threadID, content)
		return
	}

	// ファイル送信成功のメッセージを送信
	fileMsg := fmt.Sprintf("📄 **応答が長いため、ファイルとして送信しました**\nファイル名: `%s`", filename)
	if _, err := session.ChannelMessageSend(threadID, fileMsg); err != nil {
		log.Printf("ファイル送信メッセージの送信に失敗: %v", err)
	}
}

// sendAsFile は、長い応答をファイルとして送信します
func (s *MessageSender) sendAsFile(session *discordgo.Session, m *discordgo.MessageCreate, content, filename string) {
	// ファイルデータを作成
	fileData := strings.NewReader(content)

	// ファイルを添付してメッセージを送信
	_, err := session.ChannelFileSend(
		m.ChannelID,
		filename,
		fileData,
	)

	if err != nil {
		log.Printf("ファイル送信に失敗: %v", err)
		// ファイル送信に失敗した場合は通常の分割送信にフォールバック
		s.SendSplitResponse(session, m, content)
		return
	}

	// ファイル送信成功のメッセージをスレッド返信として送信
	fileMsg := fmt.Sprintf("📄 **応答が長いため、ファイルとして送信しました**\nファイル名: `%s`", filename)
	if _, err := session.ChannelMessageSendReply(m.ChannelID, fileMsg, &discordgo.MessageReference{
		MessageID: m.ID,
		ChannelID: m.ChannelID,
		GuildID:   m.GuildID,
	}); err != nil {
		log.Printf("ファイル送信メッセージの送信に失敗: %v", err)
	}
}

// splitMessage は、長いメッセージをDiscordの制限に合わせて分割します
func (s *MessageSender) splitMessage(message string) []string {
	if len(message) <= DiscordMessageLimit {
		return []string{message}
	}

	var chunks []string
	remaining := message

	for len(remaining) > 0 {
		if len(remaining) <= DiscordMessageLimit {
			chunks = append(chunks, remaining)
			break
		}

		// 2000文字以内で最も近い改行位置を探す
		splitIndex := DiscordMessageLimit
		for i := DiscordMessageLimit; i > 0; i-- {
			if remaining[i-1] == '\n' {
				splitIndex = i
				break
			}
		}

		// 改行が見つからない場合は、単語の境界で分割
		if splitIndex == DiscordMessageLimit {
			for i := DiscordMessageLimit; i > 0; i-- {
				if remaining[i-1] == ' ' {
					splitIndex = i
					break
				}
			}
		}

		// それでも見つからない場合は強制的に分割
		if splitIndex == DiscordMessageLimit {
			splitIndex = DiscordMessageLimit
		}

		chunk := remaining[:splitIndex]
		remaining = remaining[splitIndex:]

		// 先頭の空白を除去
		remaining = strings.TrimLeft(remaining, " \n")

		chunks = append(chunks, chunk)
	}

	return chunks
}
