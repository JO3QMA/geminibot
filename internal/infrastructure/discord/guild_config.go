package discord

import (
	"context"
	"fmt"
	"sync"
	"time"

	"geminibot/configs"
	"geminibot/internal/domain"
)

func newGuildConfig(guildID, apiKey, setBy, model string) domain.GuildConfig {
	// モデルが空の場合はデフォルトモデルを設定
	if model == "" {
		config, err := configs.LoadConfig()
		if err != nil {
			// 設定の読み込みに失敗した場合はデフォルト値を返す
			model = "gemini-2.5-pro"
		} else {
			model = config.Gemini.ModelName
		}
	}

	return domain.GuildConfig{
		GuildID: guildID,
		APIKey:  apiKey,
		SetBy:   setBy,
		SetAt:   time.Now(),
		Model:   model,
	}
}

// DiscordGuildAPIKeyRepository は、Discord用のAPIキー管理リポジトリの実装です
// 現在はメモリベースですが、将来的にはデータベースやKVストアに変更可能です
type DiscordGuildConfigManager struct {
	apiKeys map[string]domain.GuildConfig
	mutex   sync.RWMutex
}

// NewDiscordGuildAPIKeyRepository は新しいDiscordGuildAPIKeyRepositoryインスタンスを作成します
func NewDiscordGuildConfigManager() *DiscordGuildConfigManager {
	return &DiscordGuildConfigManager{
		apiKeys: make(map[string]domain.GuildConfig),
	}
}

// SetAPIKey は、指定されたギルドのAPIキーを設定します
func (r *DiscordGuildConfigManager) SetAPIKey(ctx context.Context, guildID, apiKey, setBy string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 既存の設定がある場合は、モデル設定を保持
	model := "" // デフォルトモデル（newGuildConfigで設定される）
	if existing, exists := r.apiKeys[guildID]; exists {
		model = existing.Model
	}

	guildAPIKey := newGuildConfig(guildID, apiKey, setBy, model)
	r.apiKeys[guildID] = guildAPIKey

	return nil
}

// GetAPIKey は、指定されたギルドのAPIキーを取得します
func (r *DiscordGuildConfigManager) GetAPIKey(ctx context.Context, guildID string) (string, error) {
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
func (r *DiscordGuildConfigManager) DeleteAPIKey(ctx context.Context, guildID string) error {
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
func (r *DiscordGuildConfigManager) HasAPIKey(ctx context.Context, guildID string) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.apiKeys[guildID]
	return exists, nil
}

// GetGuildAPIKeyInfo は、指定されたギルドのAPIキー情報を取得します（APIキーは含まれません）
func (r *DiscordGuildConfigManager) GetGuildAPIKeyInfo(ctx context.Context, guildID string) (domain.GuildConfig, error) {
	if ctx.Err() != nil {
		return domain.GuildConfig{}, ctx.Err()
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	guildAPIKey, exists := r.apiKeys[guildID]
	if !exists {
		return domain.GuildConfig{}, fmt.Errorf("ギルド %s のAPIキーが設定されていません", guildID)
	}

	// APIキーを空文字にして返す（セキュリティのため）
	info := guildAPIKey
	info.APIKey = ""
	return info, nil
}

// SetGuildModel は、指定されたギルドのAIモデルを設定します
func (r *DiscordGuildConfigManager) SetGuildModel(ctx context.Context, guildID string, model string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 既存の設定がある場合は更新、ない場合は新規作成
	if existing, exists := r.apiKeys[guildID]; exists {
		// 既存の設定を更新
		existing.Model = model
		r.apiKeys[guildID] = existing
	} else {
		// 新規作成（APIキーは空文字）
		guildAPIKey := newGuildConfig(guildID, "", "", model)
		guildAPIKey.APIKey = "" // APIキーは空文字
		r.apiKeys[guildID] = guildAPIKey
	}

	return nil
}

// GetGuildModel は、指定されたギルドのAIモデルを取得します
func (r *DiscordGuildConfigManager) GetGuildModel(ctx context.Context, guildID string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	guildAPIKey, exists := r.apiKeys[guildID]
	if !exists {
		// デフォルトモデルを返す
		config, _ := configs.LoadConfig()

		return config.Gemini.ModelName, nil
	}

	return guildAPIKey.Model, nil
}
