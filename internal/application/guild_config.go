package application

import (
	"context"
	"fmt"
	"geminibot/internal/domain"
)

// APIKeyApplicationService は、APIキーの管理を行うアプリケーションサービスです
type APIKeyApplicationService struct {
	apiKeyRepo domain.GuildConfigManager
}

// NewAPIKeyApplicationService は新しいAPIKeyApplicationServiceインスタンスを作成します
func NewAPIKeyApplicationService(apiKeyRepo domain.GuildConfigManager) *APIKeyApplicationService {
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

// GetGuildAPIKeyInfo は、指定されたギルドのAPIキー情報を取得します（APIキーは含まれません）
func (s *APIKeyApplicationService) GetGuildAPIKeyInfo(ctx context.Context, guildID string) (domain.GuildConfig, error) {
	return s.apiKeyRepo.GetGuildAPIKeyInfo(ctx, guildID)
}

// SetGuildModel は、指定されたギルドのAIモデルを設定します
func (s *APIKeyApplicationService) SetGuildModel(ctx context.Context, guildID string, model string) error {
	// モデルの妥当性を検証
	if !s.isValidModel(model) {
		return fmt.Errorf("無効なモデルです: %s", model)
	}

	// リポジトリに保存
	return s.apiKeyRepo.SetGuildModel(ctx, guildID, model)
}

// GetGuildModel は、指定されたギルドのAIモデルを取得します
func (s *APIKeyApplicationService) GetGuildModel(ctx context.Context, guildID string) (string, error) {
	return s.apiKeyRepo.GetGuildModel(ctx, guildID)
}

// isValidModel は、指定されたモデルが有効かどうかを検証します
func (s *APIKeyApplicationService) isValidModel(model string) bool {
	validModels := []string{
		"gemini-2.5-pro",
		"gemini-2.0-flash",
		"gemini-2.5-flash-lite",
	}

	for _, validModel := range validModels {
		if model == validModel {
			return true
		}
	}
	return false
}
