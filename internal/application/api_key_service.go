package application

import (
	"context"
	"fmt"
	"geminibot/internal/domain"
)

// APIKeyApplicationService は、APIキーの管理を行うアプリケーションサービスです
type APIKeyApplicationService struct {
	apiKeyRepo domain.GuildAPIKeyRepository
}

// NewAPIKeyApplicationService は新しいAPIKeyApplicationServiceインスタンスを作成します
func NewAPIKeyApplicationService(apiKeyRepo domain.GuildAPIKeyRepository) *APIKeyApplicationService {
	return &APIKeyApplicationService{
		apiKeyRepo: apiKeyRepo,
	}
}

// SetGuildAPIKey は、指定されたギルドのAPIキーを設定します
func (s *APIKeyApplicationService) SetGuildAPIKey(ctx context.Context, guildID, apiKey, setBy string) error {
	// APIキーの形式を検証（基本的な検証）
	if apiKey == "" {
		return fmt.Errorf("APIキーが空です")
	}
	
	if len(apiKey) < 10 {
		return fmt.Errorf("APIキーが短すぎます")
	}
	
	// リポジトリに保存
	return s.apiKeyRepo.SetAPIKey(ctx, guildID, apiKey, setBy)
}

// GetGuildAPIKey は、指定されたギルドのAPIキーを取得します
func (s *APIKeyApplicationService) GetGuildAPIKey(ctx context.Context, guildID string) (string, error) {
	return s.apiKeyRepo.GetAPIKey(ctx, guildID)
}

// DeleteGuildAPIKey は、指定されたギルドのAPIキーを削除します
func (s *APIKeyApplicationService) DeleteGuildAPIKey(ctx context.Context, guildID string) error {
	return s.apiKeyRepo.DeleteAPIKey(ctx, guildID)
}

// HasGuildAPIKey は、指定されたギルドにAPIキーが設定されているかを確認します
func (s *APIKeyApplicationService) HasGuildAPIKey(ctx context.Context, guildID string) (bool, error) {
	return s.apiKeyRepo.HasAPIKey(ctx, guildID)
}
