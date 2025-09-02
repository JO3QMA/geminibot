package discord

import (
	"context"
	"fmt"
	"sync"

	"geminibot/internal/domain"
)

// DiscordGuildAPIKeyRepository は、Discord用のAPIキー管理リポジトリの実装です
// 現在はメモリベースですが、将来的にはデータベースやKVストアに変更可能です
type DiscordGuildAPIKeyRepository struct {
	apiKeys map[string]domain.GuildAPIKey
	mutex   sync.RWMutex
}

// NewDiscordGuildAPIKeyRepository は新しいDiscordGuildAPIKeyRepositoryインスタンスを作成します
func NewDiscordGuildAPIKeyRepository() *DiscordGuildAPIKeyRepository {
	return &DiscordGuildAPIKeyRepository{
		apiKeys: make(map[string]domain.GuildAPIKey),
	}
}

// SetAPIKey は、指定されたギルドのAPIキーを設定します
func (r *DiscordGuildAPIKeyRepository) SetAPIKey(ctx context.Context, guildID, apiKey, setBy string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	guildAPIKey := domain.NewGuildAPIKey(guildID, apiKey, setBy)
	r.apiKeys[guildID] = guildAPIKey

	return nil
}

// GetAPIKey は、指定されたギルドのAPIキーを取得します
func (r *DiscordGuildAPIKeyRepository) GetAPIKey(ctx context.Context, guildID string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	guildAPIKey, exists := r.apiKeys[guildID]
	if !exists {
		return "", fmt.Errorf("ギルド %s のAPIキーが設定されていません", guildID)
	}

	return guildAPIKey.APIKey, nil
}

// DeleteAPIKey は、指定されたギルドのAPIキーを削除します
func (r *DiscordGuildAPIKeyRepository) DeleteAPIKey(ctx context.Context, guildID string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.apiKeys[guildID]; !exists {
		return fmt.Errorf("ギルド %s のAPIキーが設定されていません", guildID)
	}

	delete(r.apiKeys, guildID)
	return nil
}

// HasAPIKey は、指定されたギルドにAPIキーが設定されているかを確認します
func (r *DiscordGuildAPIKeyRepository) HasAPIKey(ctx context.Context, guildID string) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.apiKeys[guildID]
	return exists, nil
}
