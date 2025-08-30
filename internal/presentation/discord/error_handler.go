package discord

import (
	"fmt"
	"strings"
)

// ErrorHandler は、エラーの処理とフォーマットを担当します
type ErrorHandler struct{}

// NewErrorHandler は新しいErrorHandlerインスタンスを作成します
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// IsTimeoutError は、エラーがタイムアウトエラーかどうかを判定します
func (h *ErrorHandler) IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	// context.DeadlineExceeded エラーの検出
	if err.Error() == "context deadline exceeded" {
		return true
	}

	// タイムアウト関連のエラーメッセージを検出
	errorMsg := err.Error()
	timeoutKeywords := []string{
		"timeout",
		"タイムアウト",
		"deadline exceeded",
		"context deadline",
		"request timeout",
	}

	for _, keyword := range timeoutKeywords {
		if strings.Contains(strings.ToLower(errorMsg), strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

// FormatError は、エラーを適切なメッセージにフォーマットします
func (h *ErrorHandler) FormatError(err error) string {
	// タイムアウトエラーの場合
	if h.IsTimeoutError(err) {
		return "⏰ **タイムアウトしました**\n\n処理に時間がかかりすぎました。以下の対処法をお試しください：\n\n" +
			"• 質問を短くしてみる\n" +
			"• 複雑な質問を分割する\n" +
			"• しばらく待ってから再度お試しください\n\n" +
			"ご不便をおかけして申し訳ございません。"
	}

	// 荒らし対策エラーの場合
	switch err.Error() {
	case "レート制限を超過しました":
		return "⚠️ **レート制限を超過しました**\nしばらく待ってから再度お試しください。"
	case "スパムが検出されました":
		return "🚫 **スパムが検出されました**\n短時間での大量メッセージは禁止されています。"
	case "不適切なコンテンツが検出されました":
		return "🚫 **不適切なコンテンツが検出されました**\n禁止ワードが含まれています。"
	case "メッセージが長すぎます":
		return "📏 **メッセージが長すぎます**\n2000文字以内でお願いします。"
	case "重複メッセージが検出されました":
		return "🔄 **重複メッセージが検出されました**\n同じ内容のメッセージを連続で送信しないでください。"
	default:
		return fmt.Sprintf("❌ **エラーが発生しました**\n%s", err.Error())
	}
}
