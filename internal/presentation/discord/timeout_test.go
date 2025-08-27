package discord

import (
	"strings"
	"testing"
)

func TestDiscordHandler_IsTimeoutError(t *testing.T) {
	handler := &DiscordHandler{}

	// タイムアウトエラーのテストケース
	timeoutErrors := []string{
		"context deadline exceeded",
		"timeout",
		"タイムアウト",
		"deadline exceeded",
		"context deadline",
		"request timeout",
		"TIMEOUT",
		"タイムアウトしました",
		"Request timeout",
	}

	for _, errMsg := range timeoutErrors {
		err := &timeoutError{message: errMsg}
		if !handler.isTimeoutError(err) {
			t.Errorf("タイムアウトエラーとして認識されるべき: %s", errMsg)
		}
	}

	// 非タイムアウトエラーのテストケース
	nonTimeoutErrors := []string{
		"network error",
		"permission denied",
		"invalid token",
		"エラーが発生しました",
		"something went wrong",
	}

	for _, errMsg := range nonTimeoutErrors {
		err := &timeoutError{message: errMsg}
		if handler.isTimeoutError(err) {
			t.Errorf("タイムアウトエラーとして認識されるべきではない: %s", errMsg)
		}
	}

	// nilエラーのテスト
	if handler.isTimeoutError(nil) {
		t.Error("nilエラーはタイムアウトエラーとして認識されるべきではない")
	}
}

func TestDiscordHandler_FormatError(t *testing.T) {
	handler := &DiscordHandler{}

	// タイムアウトエラーのフォーマットテスト
	timeoutErr := &timeoutError{message: "context deadline exceeded"}
	formatted := handler.formatError(timeoutErr)

	if !strings.Contains(formatted, "⏰ **タイムアウトしました**") {
		t.Error("タイムアウトエラーメッセージが正しくフォーマットされていません")
	}

	if !strings.Contains(formatted, "質問を短くしてみる") {
		t.Error("タイムアウトエラーメッセージに対処法が含まれていません")
	}

	// 荒らし対策エラーのフォーマットテスト
	spamErr := &timeoutError{message: "レート制限を超過しました"}
	formatted = handler.formatError(spamErr)

	if !strings.Contains(formatted, "⚠️ **レート制限を超過しました**") {
		t.Error("レート制限エラーメッセージが正しくフォーマットされていません")
	}

	// 一般的なエラーのフォーマットテスト
	generalErr := &timeoutError{message: "一般的なエラー"}
	formatted = handler.formatError(generalErr)

	if !strings.Contains(formatted, "❌ **エラーが発生しました**") {
		t.Error("一般的なエラーメッセージが正しくフォーマットされていません")
	}
}

// timeoutError は、テスト用のエラー型です
type timeoutError struct {
	message string
}

func (e *timeoutError) Error() string {
	return e.message
}
