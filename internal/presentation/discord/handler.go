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

// DiscordHandler は、Discordのイベントハンドラです
type DiscordHandler struct {
	session        *discordgo.Session
	mentionService *application.MentionApplicationService
	botID          string
	botUsername    string
}

// DiscordMessageLimit は、Discordのメッセージ長制限です
const DiscordMessageLimit = 2000

// NewDiscordHandler は新しいDiscordHandlerインスタンスを作成します
func NewDiscordHandler(
	session *discordgo.Session,
	mentionService *application.MentionApplicationService,
	botID string,
) *DiscordHandler {
	return &DiscordHandler{
		session:        session,
		mentionService: mentionService,
		botID:          botID,
	}
}

// SetupHandlers は、Discordのイベントハンドラを設定します
func (h *DiscordHandler) SetupHandlers() {
	h.session.AddHandler(h.handleMessageCreate)
	h.session.AddHandler(h.handleReady)
}

// handleReady は、Botが準備完了した際のイベントを処理します
func (h *DiscordHandler) handleReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("Botが準備完了しました: %s#%s", event.User.Username, event.User.Discriminator)
	h.botUsername = event.User.Username
}

// handleMessageCreate は、メッセージ作成イベントを処理します
func (h *DiscordHandler) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Bot自身のメッセージは無視
	if m.Author.ID == h.botID {
		return
	}

	// メンションされているかチェック
	if !h.isMentioned(m) {
		return
	}

	log.Printf("Botへのメンションを検出: %s", m.Content)

	// メンション情報を作成
	mention := h.createBotMention(m)

	// 非同期でメンションを処理
	go h.processMentionAsync(s, m, mention)
}

// isMentioned は、メッセージがBotへのメンションかどうかを判定します
func (h *DiscordHandler) isMentioned(m *discordgo.MessageCreate) bool {
	// メンション配列をチェック
	for _, mention := range m.Mentions {
		if mention.ID == h.botID {
			return true
		}
	}

	// メンション配列が空の場合、コンテンツをチェック
	if len(m.Mentions) == 0 {
		content := strings.ToLower(m.Content)
		botMention := fmt.Sprintf("@%s", strings.ToLower(h.botUsername))
		return strings.Contains(content, botMention)
	}

	return false
}

// createBotMention は、DiscordメッセージからBotMentionオブジェクトを作成します
func (h *DiscordHandler) createBotMention(m *discordgo.MessageCreate) domain.BotMention {
	// メンション部分を除去したコンテンツを取得
	content := h.extractUserContent(m)

	return domain.NewBotMention(
		domain.NewChannelID(m.ChannelID),
		domain.NewUserID(m.Author.ID),
		content,
		m.ID,
	)
}

// extractUserContent は、メンション部分を除去したユーザーのコンテンツを抽出します
func (h *DiscordHandler) extractUserContent(m *discordgo.MessageCreate) string {
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

// processMentionAsync は、メンションを非同期で処理します
func (h *DiscordHandler) processMentionAsync(s *discordgo.Session, m *discordgo.MessageCreate, mention domain.BotMention) {
	// 処理中メッセージを送信
	thinkingMsg, err := s.ChannelMessageSend(m.ChannelID, "🤔 考え中...")
	if err != nil {
		log.Printf("処理中メッセージの送信に失敗: %v", err)
		return
	}

	// メンションを処理
	ctx := context.Background()
	response, err := h.mentionService.HandleMention(ctx, mention)

	// 処理中メッセージを削除
	s.ChannelMessageDelete(m.ChannelID, thinkingMsg.ID)

	if err != nil {
		log.Printf("メンション処理に失敗: %v", err)
		errorMsg := fmt.Sprintf("❌ エラーが発生しました: %s", err.Error())
		s.ChannelMessageSendReply(m.ChannelID, errorMsg, &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})
		return
	}

	// 応答を分割して送信
	h.sendSplitResponse(s, m, response)
}

// sendSplitResponse は、長い応答を複数のメッセージに分割して送信します
func (h *DiscordHandler) sendSplitResponse(s *discordgo.Session, m *discordgo.MessageCreate, response string) {
	// 応答をDiscord用にフォーマット
	formattedResponse := h.formatForDiscord(response)
	
	// 応答をDiscordの制限に合わせて分割
	chunks := h.splitMessage(formattedResponse)
	
	if len(chunks) == 1 {
		// 単一メッセージの場合
		_, err := s.ChannelMessageSendReply(m.ChannelID, chunks[0], &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})
		if err != nil {
			log.Printf("応答メッセージの送信に失敗: %v", err)
		}
		return
	}

	// 複数メッセージの場合
	for i, chunk := range chunks {
		var err error
		if i == 0 {
			// 最初のメッセージはリプライとして送信
			_, err = s.ChannelMessageSendReply(m.ChannelID, chunk, &discordgo.MessageReference{
				MessageID: m.ID,
				ChannelID: m.ChannelID,
				GuildID:   m.GuildID,
			})
		} else {
			// 2番目以降は通常のメッセージとして送信
			_, err = s.ChannelMessageSend(m.ChannelID, chunk)
		}
		
		if err != nil {
			log.Printf("応答メッセージの送信に失敗 (チャンク %d): %v", i+1, err)
			break
		}
	}
}

// formatForDiscord は、Geminiからの応答をDiscord用にフォーマットします
func (h *DiscordHandler) formatForDiscord(response string) string {
	// markdownのコードブロックをDiscord用に変換
	formatted := h.convertCodeBlocks(response)
	
	// markdownのインラインコードをDiscord用に変換
	formatted = h.convertInlineCode(formatted)
	
	// markdownの太字をDiscord用に変換
	formatted = h.convertBold(formatted)
	
	// markdownの斜体をDiscord用に変換
	formatted = h.convertItalic(formatted)
	
	// markdownのリストをDiscord用に変換
	formatted = h.convertLists(formatted)
	
	return formatted
}

// convertCodeBlocks は、markdownのコードブロックをDiscord用に変換します
func (h *DiscordHandler) convertCodeBlocks(text string) string {
	// ```で囲まれたコードブロックを```に変換
	// 言語指定がある場合は除去
	lines := strings.Split(text, "\n")
	var result []string
	inCodeBlock := false
	codeBlockContent := []string{}
	
	for _, line := range lines {
		if strings.HasPrefix(line, "```") && !inCodeBlock {
			// コードブロック開始
			inCodeBlock = true
			codeBlockContent = []string{}
		} else if strings.HasPrefix(line, "```") && inCodeBlock {
			// コードブロック終了
			inCodeBlock = false
			if len(codeBlockContent) > 0 {
				result = append(result, "```")
				result = append(result, codeBlockContent...)
				result = append(result, "```")
			}
		} else if inCodeBlock {
			// コードブロック内の内容
			codeBlockContent = append(codeBlockContent, line)
		} else {
			// 通常の行
			result = append(result, line)
		}
	}
	
	return strings.Join(result, "\n")
}

// convertInlineCode は、markdownのインラインコードをDiscord用に変換します
func (h *DiscordHandler) convertInlineCode(text string) string {
	// `で囲まれたインラインコードを`に変換
	// ただし、コードブロック内は除外
	lines := strings.Split(text, "\n")
	var result []string
	
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			// コードブロックの境界はそのまま
			result = append(result, line)
		} else {
			// インラインコードを変換
			converted := h.convertInlineCodeInLine(line)
			result = append(result, converted)
		}
	}
	
	return strings.Join(result, "\n")
}

// convertInlineCodeInLine は、1行内のインラインコードを変換します
func (h *DiscordHandler) convertInlineCodeInLine(line string) string {
	// バッククォートのペアを`に変換
	// ただし、コードブロック内は除外
	var result strings.Builder
	inInlineCode := false
	codeContent := strings.Builder{}
	
	for i := 0; i < len(line); i++ {
		if line[i] == '`' && !inInlineCode {
			// インラインコード開始
			inInlineCode = true
			codeContent.Reset()
		} else if line[i] == '`' && inInlineCode {
			// インラインコード終了
			inInlineCode = false
			result.WriteString("`")
			result.WriteString(codeContent.String())
			result.WriteString("`")
		} else if inInlineCode {
			// インラインコード内の内容
			codeContent.WriteByte(line[i])
		} else {
			// 通常の文字
			result.WriteByte(line[i])
		}
	}
	
	return result.String()
}

// convertBold は、markdownの太字をDiscord用に変換します
func (h *DiscordHandler) convertBold(text string) string {
	// **で囲まれた太字を**に変換
	// ただし、コードブロック内は除外
	lines := strings.Split(text, "\n")
	var result []string
	
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			// コードブロックの境界はそのまま
			result = append(result, line)
		} else {
			// 太字を変換
			converted := h.convertBoldInLine(line)
			result = append(result, converted)
		}
	}
	
	return strings.Join(result, "\n")
}

// convertBoldInLine は、1行内の太字を変換します
func (h *DiscordHandler) convertBoldInLine(line string) string {
	// **で囲まれた太字を**に変換
	// ただし、インラインコード内は除外
	var result strings.Builder
	inInlineCode := false
	inBold := false
	boldContent := strings.Builder{}
	
	for i := 0; i < len(line); i++ {
		if line[i] == '`' {
			// インラインコードの境界
			if inBold {
				// 太字を終了してからインラインコードを処理
				inBold = false
				result.WriteString("**")
				result.WriteString(boldContent.String())
				result.WriteString("**")
				boldContent.Reset()
			}
			inInlineCode = !inInlineCode
			result.WriteByte(line[i])
		} else if !inInlineCode && i+1 < len(line) && line[i] == '*' && line[i+1] == '*' {
			// **の検出
			if !inBold {
				// 太字開始
				inBold = true
				boldContent.Reset()
			} else {
				// 太字終了
				inBold = false
				result.WriteString("**")
				result.WriteString(boldContent.String())
				result.WriteString("**")
				boldContent.Reset()
			}
			i++ // 次の*をスキップ
		} else if inBold {
			// 太字内の内容
			boldContent.WriteByte(line[i])
		} else {
			// 通常の文字
			result.WriteByte(line[i])
		}
	}
	
	// 未終了の太字があれば終了
	if inBold {
		result.WriteString("**")
		result.WriteString(boldContent.String())
		result.WriteString("**")
	}
	
	return result.String()
}

// convertItalic は、markdownの斜体をDiscord用に変換します
func (h *DiscordHandler) convertItalic(text string) string {
	// *で囲まれた斜体を*に変換（ただし、太字の**は除外）
	// ただし、コードブロック内は除外
	lines := strings.Split(text, "\n")
	var result []string
	
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			// コードブロックの境界はそのまま
			result = append(result, line)
		} else {
			// 斜体を変換
			converted := h.convertItalicInLine(line)
			result = append(result, converted)
		}
	}
	
	return strings.Join(result, "\n")
}

// convertItalicInLine は、1行内の斜体を変換します
func (h *DiscordHandler) convertItalicInLine(line string) string {
	// *で囲まれた斜体を*に変換（ただし、太字の**は除外）
	// ただし、インラインコード内は除外
	var result strings.Builder
	inInlineCode := false
	inItalic := false
	italicContent := strings.Builder{}
	
	for i := 0; i < len(line); i++ {
		if line[i] == '`' {
			// インラインコードの境界
			if inItalic {
				// 斜体を終了してからインラインコードを処理
				inItalic = false
				result.WriteString("*")
				result.WriteString(italicContent.String())
				result.WriteString("*")
				italicContent.Reset()
			}
			inInlineCode = !inInlineCode
			result.WriteByte(line[i])
		} else if !inInlineCode && line[i] == '*' {
			// *の検出
			if i+1 < len(line) && line[i+1] == '*' {
				// **の場合は太字なのでスキップ
				result.WriteString("**")
				i++
			} else if !inItalic {
				// 斜体開始
				inItalic = true
				italicContent.Reset()
			} else {
				// 斜体終了
				inItalic = false
				result.WriteString("*")
				result.WriteString(italicContent.String())
				result.WriteString("*")
				italicContent.Reset()
			}
		} else if inItalic {
			// 斜体内の内容
			italicContent.WriteByte(line[i])
		} else {
			// 通常の文字
			result.WriteByte(line[i])
		}
	}
	
	// 未終了の斜体があれば終了
	if inItalic {
		result.WriteString("*")
		result.WriteString(italicContent.String())
		result.WriteString("*")
	}
	
	return result.String()
}

// convertLists は、markdownのリストをDiscord用に変換します
func (h *DiscordHandler) convertLists(text string) string {
	// リストの変換（基本的にはそのまま、必要に応じて調整）
	// Discordは基本的なリスト表示をサポートしているので、
	// 主に番号付きリストの形式を調整
	lines := strings.Split(text, "\n")
	var result []string
	
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			// コードブロックの境界はそのまま
			result = append(result, line)
		} else {
			// リストを変換
			converted := h.convertListInLine(line)
			result = append(result, converted)
		}
	}
	
	return strings.Join(result, "\n")
}

// convertListInLine は、1行内のリストを変換します
func (h *DiscordHandler) convertListInLine(line string) string {
	// 番号付きリストの形式を調整
	// 1. の形式を1) に変換（Discordの表示を改善）
	trimmed := strings.TrimSpace(line)
	if len(trimmed) >= 2 && trimmed[1] == '.' {
		// 番号付きリストの可能性
		if trimmed[0] >= '0' && trimmed[0] <= '9' {
			// 数字. の形式を数字) に変換
			return strings.Replace(line, ". ", ") ", 1)
		}
	}
	
	return line
}

// splitMessage は、長いメッセージをDiscordの制限に合わせて分割します
func (h *DiscordHandler) splitMessage(message string) []string {
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
