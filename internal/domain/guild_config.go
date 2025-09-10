package domain

import (
	"context"
	"time"
)

// GuildAPIKey は、Discordサーバー（ギルド）固有のAPIキーを表します
type GuildConfig struct {
	GuildID string
	APIKey  string
	SetBy   string
	SetAt   time.Time
	Model   string
}

// NewGuildConfig は新しいGuildConfigインスタンスを作成します
func NewGuildConfig(guildID, apiKey, setBy, model string) GuildConfig {
	if model == "" {
		model = "gemini-2.5-pro" // デフォルトモデル
	}

	return GuildConfig{
		GuildID: guildID,
		APIKey:  apiKey,
		SetBy:   setBy,
		SetAt:   time.Now(),
		Model:   model,
	}
}

// GuildConfigManager は、ギルド固有のAPIキーの永続化を行うインターフェースです
type GuildConfigManager interface {
	// SetAPIKey は、指定されたギルドのAPIキーを設定します
	SetAPIKey(ctx context.Context, guildID string, apiKey string, setBy string) error

	// GetAPIKey は、指定されたギルドのAPIキーを取得します
	GetAPIKey(ctx context.Context, guildID string) (string, error)

	// DeleteAPIKey は、指定されたギルドのAPIキーを削除します
	DeleteAPIKey(ctx context.Context, guildID string) error

	// HasAPIKey は、指定されたギルドにAPIキーが設定されているかを確認します
	HasAPIKey(ctx context.Context, guildID string) (bool, error)

	// GetGuildAPIKeyInfo は、指定されたギルドのAPIキー情報を取得します（APIキーは含まれません）
	GetGuildAPIKeyInfo(ctx context.Context, guildID string) (GuildConfig, error)

	// SetGuildModel は、指定されたギルドのAIモデルを設定します
	SetGuildModel(ctx context.Context, guildID string, model string) error

	// GetGuildModel は、指定されたギルドのAIモデルを取得します
	GetGuildModel(ctx context.Context, guildID string) (string, error)
}
