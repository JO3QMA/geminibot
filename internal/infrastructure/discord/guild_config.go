package discord

import (
	"context"
	"fmt"
	"sync"
	"time"

	"geminibot/internal/domain"
)

// GuildConfigManager は、Discord用のギルド別 API キー／モデル設定のインメモリ実装です。
// 現在はメモリベースですが、将来的にはデータベースやKVストアに変更可能です。
type GuildConfigManager struct {
	apiKeys          map[string]domain.GuildConfig
	mutex            sync.RWMutex
	defaultTextModel string
}

// NewGuildConfigManager は新しい GuildConfigManager を作成します。
// defaultTextModel はギルド未登録時やモデル未設定時に使う既定のテキスト生成モデル名です。
func NewGuildConfigManager(defaultTextModel string) *GuildConfigManager {
	return &GuildConfigManager{
		apiKeys:          make(map[string]domain.GuildConfig),
		defaultTextModel: defaultTextModel,
	}
}

func (r *GuildConfigManager) makeGuildConfig(guildID, apiKey, setBy, model string) domain.GuildConfig {
	if model == "" {
		model = r.defaultTextModel
	}

	return domain.GuildConfig{
		GuildID: guildID,
		APIKey:  apiKey,
		SetBy:   setBy,
		SetAt:   time.Now(),
		Model:   model,
	}
}

// SetAPIKey は、指定されたギルドのAPIキーを設定します
func (r *GuildConfigManager) SetAPIKey(ctx context.Context, guildID, apiKey, setBy string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 既存の設定がある場合は、モデル設定を保持
	model := ""
	if existing, exists := r.apiKeys[guildID]; exists {
		model = existing.Model
	}

	guildAPIKey := r.makeGuildConfig(guildID, apiKey, setBy, model)
	r.apiKeys[guildID] = guildAPIKey

	return nil
}

// GetAPIKey は、指定されたギルドのAPIキーを取得します
func (r *GuildConfigManager) GetAPIKey(ctx context.Context, guildID string) (string, error) {
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
func (r *GuildConfigManager) DeleteAPIKey(ctx context.Context, guildID string) error {
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
func (r *GuildConfigManager) HasAPIKey(ctx context.Context, guildID string) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.apiKeys[guildID]
	return exists, nil
}

// GetGuildAPIKeyInfo は、指定されたギルドのAPIキー情報を取得します（APIキーは含まれません）
func (r *GuildConfigManager) GetGuildAPIKeyInfo(ctx context.Context, guildID string) (domain.GuildConfig, error) {
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
func (r *GuildConfigManager) SetGuildModel(ctx context.Context, guildID, model string) error {
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
		guildAPIKey := r.makeGuildConfig(guildID, "", "", model)
		guildAPIKey.APIKey = ""
		r.apiKeys[guildID] = guildAPIKey
	}

	return nil
}

// GetGuildModel は、指定されたギルドのAIモデルを取得します
func (r *GuildConfigManager) GetGuildModel(ctx context.Context, guildID string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	guildAPIKey, exists := r.apiKeys[guildID]
	if !exists {
		return r.defaultTextModel, nil
	}

	return guildAPIKey.Model, nil
}
