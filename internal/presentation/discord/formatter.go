package discord

import (
	"strings"
)

// MessageFormatter は、Geminiからの応答をDiscord用にフォーマットする責務を担当します
type MessageFormatter struct{}

// NewMessageFormatter は新しいMessageFormatterインスタンスを作成します
func NewMessageFormatter() *MessageFormatter {
	return &MessageFormatter{}
}

// FormatForDiscord は、Geminiからの応答をDiscord用にフォーマットします
func (f *MessageFormatter) FormatForDiscord(response string) string {
	// markdownのコードブロックをDiscord用に変換
	formatted := f.convertCodeBlocks(response)

	// markdownのインラインコードをDiscord用に変換
	formatted = f.convertInlineCode(formatted)

	// markdownの太字をDiscord用に変換
	formatted = f.convertBold(formatted)

	// markdownの斜体をDiscord用に変換
	formatted = f.convertItalic(formatted)

	// markdownのリストをDiscord用に変換
	formatted = f.convertLists(formatted)

	return formatted
}

// convertCodeBlocks は、markdownのコードブロックをDiscord用に変換します
func (f *MessageFormatter) convertCodeBlocks(text string) string {
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
func (f *MessageFormatter) convertInlineCode(text string) string {
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
			converted := f.convertInlineCodeInLine(line)
			result = append(result, converted)
		}
	}

	return strings.Join(result, "\n")
}

// convertInlineCodeInLine は、1行内のインラインコードを変換します
func (f *MessageFormatter) convertInlineCodeInLine(line string) string {
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
func (f *MessageFormatter) convertBold(text string) string {
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
			converted := f.convertBoldInLine(line)
			result = append(result, converted)
		}
	}

	return strings.Join(result, "\n")
}

// convertBoldInLine は、1行内の太字を変換します
func (f *MessageFormatter) convertBoldInLine(line string) string {
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
func (f *MessageFormatter) convertItalic(text string) string {
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
			converted := f.convertItalicInLine(line)
			result = append(result, converted)
		}
	}

	return strings.Join(result, "\n")
}

// convertItalicInLine は、1行内の斜体を変換します
func (f *MessageFormatter) convertItalicInLine(line string) string {
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
func (f *MessageFormatter) convertLists(text string) string {
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
			converted := f.convertListInLine(line)
			result = append(result, converted)
		}
	}

	return strings.Join(result, "\n")
}

// convertListInLine は、1行内のリストを変換します
func (f *MessageFormatter) convertListInLine(line string) string {
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
