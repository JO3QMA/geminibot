package domain

import "errors"

// ドメイン固有のエラー型を定義
var (
	// ErrEmptyConversationHistory は、会話履歴が空の場合のエラーです
	ErrEmptyConversationHistory = errors.New("会話履歴が空です")

	// ErrInvalidMessage は、無効なメッセージの場合のエラーです
	ErrInvalidMessage = errors.New("無効なメッセージです")

	// ErrInvalidPrompt は、無効なプロンプトの場合のエラーです
	ErrInvalidPrompt = errors.New("無効なプロンプトです")

	// ErrInvalidChannelID は、無効なチャンネルIDの場合のエラーです
	ErrInvalidChannelID = errors.New("無効なチャンネルIDです")

	// ErrInvalidUserID は、無効なユーザーIDの場合のエラーです
	ErrInvalidUserID = errors.New("無効なユーザーIDです")
)
