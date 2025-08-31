package discord

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"geminibot/internal/domain"
)

// ErrorType はエラーの種類を定義します
type ErrorType string

const (
	ErrorTypeTimeout        ErrorType = "timeout"
	ErrorTypeNetwork        ErrorType = "network"
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypeRateLimit      ErrorType = "rate_limit"
	ErrorTypeValidation     ErrorType = "validation"
	ErrorTypeInternal       ErrorType = "internal"
	ErrorTypeUnknown        ErrorType = "unknown"
)

// ErrorInfo はエラー情報を格納します
type ErrorInfo struct {
	Type        ErrorType
	Message     string
	UserMessage string
	Retryable   bool
	LogLevel    string
}

// ErrorHandler は、エラーの処理とフォーマットを担当します
type ErrorHandler struct {
	errorMap map[ErrorType]ErrorInfo
}

// NewErrorHandler は新しいErrorHandlerインスタンスを作成します
func NewErrorHandler() *ErrorHandler {
	handler := &ErrorHandler{
		errorMap: make(map[ErrorType]ErrorInfo),
	}

	// エラータイプごとの処理を定義
	handler.errorMap[ErrorTypeTimeout] = ErrorInfo{
		Type:        ErrorTypeTimeout,
		Message:     "リクエストがタイムアウトしました",
		UserMessage: "申し訳ございません。処理に時間がかかりすぎました。しばらく時間をおいてから再度お試しください。",
		Retryable:   true,
		LogLevel:    "warn",
	}

	handler.errorMap[ErrorTypeNetwork] = ErrorInfo{
		Type:        ErrorTypeNetwork,
		Message:     "ネットワークエラーが発生しました",
		UserMessage: "申し訳ございません。ネットワーク接続に問題が発生しました。しばらく時間をおいてから再度お試しください。",
		Retryable:   true,
		LogLevel:    "error",
	}

	handler.errorMap[ErrorTypeAuthentication] = ErrorInfo{
		Type:        ErrorTypeAuthentication,
		Message:     "認証エラーが発生しました",
		UserMessage: "申し訳ございません。認証に問題が発生しました。管理者にお問い合わせください。",
		Retryable:   false,
		LogLevel:    "error",
	}

	handler.errorMap[ErrorTypeRateLimit] = ErrorInfo{
		Type:        ErrorTypeRateLimit,
		Message:     "レート制限に達しました",
		UserMessage: "申し訳ございません。リクエスト制限に達しました。しばらく時間をおいてから再度お試しください。",
		Retryable:   true,
		LogLevel:    "warn",
	}

	handler.errorMap[ErrorTypeValidation] = ErrorInfo{
		Type:        ErrorTypeValidation,
		Message:     "入力値が無効です",
		UserMessage: "申し訳ございません。入力内容に問題があります。内容を確認してから再度お試しください。",
		Retryable:   false,
		LogLevel:    "info",
	}

	handler.errorMap[ErrorTypeInternal] = ErrorInfo{
		Type:        ErrorTypeInternal,
		Message:     "内部エラーが発生しました",
		UserMessage: "申し訳ございません。システムエラーが発生しました。しばらく時間をおいてから再度お試しください。",
		Retryable:   true,
		LogLevel:    "error",
	}

	handler.errorMap[ErrorTypeUnknown] = ErrorInfo{
		Type:        ErrorTypeUnknown,
		Message:     "予期しないエラーが発生しました",
		UserMessage: "申し訳ございません。予期しないエラーが発生しました。しばらく時間をおいてから再度お試しください。",
		Retryable:   false,
		LogLevel:    "error",
	}

	return handler
}

// ClassifyError はエラーを分類します
func (h *ErrorHandler) ClassifyError(err error) ErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}

	errStr := err.Error()

	// コンテキストエラー（タイムアウト）
	if err == context.DeadlineExceeded || err == context.Canceled {
		return ErrorTypeTimeout
	}

	// ネットワークエラー
	if strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "timeout") {
		return ErrorTypeNetwork
	}

	// 認証エラー
	if strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "forbidden") ||
		strings.Contains(errStr, "invalid token") {
		return ErrorTypeAuthentication
	}

	// レート制限エラー
	if strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "too many requests") {
		return ErrorTypeRateLimit
	}

	// バリデーションエラー
	if strings.Contains(errStr, "validation") ||
		strings.Contains(errStr, "invalid") {
		return ErrorTypeValidation
	}

	// ドメインエラー
	if err == domain.ErrEmptyConversationHistory ||
		err == domain.ErrInvalidMessage ||
		err == domain.ErrInvalidPrompt ||
		err == domain.ErrInvalidChannelID ||
		err == domain.ErrInvalidUserID {
		return ErrorTypeValidation
	}

	return ErrorTypeInternal
}

// FormatError はエラーをユーザー向けメッセージにフォーマットします
func (h *ErrorHandler) FormatError(err error) string {
	errorType := h.ClassifyError(err)
	errorInfo, exists := h.errorMap[errorType]

	if !exists {
		errorInfo = h.errorMap[ErrorTypeUnknown]
	}

	// ログ出力
	switch errorInfo.LogLevel {
	case "error":
		log.Printf("エラー発生 [%s]: %v", errorType, err)
	case "warn":
		log.Printf("警告発生 [%s]: %v", errorType, err)
	case "info":
		log.Printf("情報 [%s]: %v", errorType, err)
	}

	return errorInfo.UserMessage
}

// IsRetryable はエラーが再試行可能かどうかを判定します
func (h *ErrorHandler) IsRetryable(err error) bool {
	errorType := h.ClassifyError(err)
	errorInfo, exists := h.errorMap[errorType]

	if !exists {
		return false
	}

	return errorInfo.Retryable
}

// GetRetryDelay は再試行までの待機時間を取得します
func (h *ErrorHandler) GetRetryDelay(err error) time.Duration {
	errorType := h.ClassifyError(err)

	switch errorType {
	case ErrorTypeTimeout:
		return 5 * time.Second
	case ErrorTypeNetwork:
		return 10 * time.Second
	case ErrorTypeRateLimit:
		return 30 * time.Second
	case ErrorTypeInternal:
		return 15 * time.Second
	default:
		return 5 * time.Second
	}
}

// LogError はエラーを適切なログレベルで出力します
func (h *ErrorHandler) LogError(err error, context string) {
	errorType := h.ClassifyError(err)
	errorInfo, exists := h.errorMap[errorType]

	if !exists {
		errorInfo = h.errorMap[ErrorTypeUnknown]
	}

	logMessage := fmt.Sprintf("[%s] %s: %v", errorType, context, err)

	switch errorInfo.LogLevel {
	case "error":
		log.Printf("ERROR: %s", logMessage)
	case "warn":
		log.Printf("WARN: %s", logMessage)
	case "info":
		log.Printf("INFO: %s", logMessage)
	}
}
