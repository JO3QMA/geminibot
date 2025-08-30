package database

import (
	"context"
	"fmt"
	"log"

	"geminibot/internal/domain"
)

// HybridConversationRepository は、PostgreSQLとRedisを組み合わせた会話履歴リポジトリです
type HybridConversationRepository struct {
	postgres *PostgresConversationRepository
	redis    *RedisConversationCache
}

// NewHybridConversationRepository は新しいHybridConversationRepositoryインスタンスを作成します
func NewHybridConversationRepository(postgres *PostgresConversationRepository, redis *RedisConversationCache) *HybridConversationRepository {
	return &HybridConversationRepository{
		postgres: postgres,
		redis:    redis,
	}
}

// SaveMessage は、メッセージをPostgreSQLとRedisの両方に保存します
func (r *HybridConversationRepository) SaveMessage(ctx context.Context, message domain.Message, channelID domain.ChannelID) error {
	// PostgreSQLに永続化
	if err := r.postgres.SaveMessage(ctx, message, channelID); err != nil {
		log.Printf("PostgreSQL保存に失敗: %v", err)
		// PostgreSQL保存に失敗してもRedisには保存を試行
	}

	// Redisにキャッシュ
	if err := r.redis.CacheMessage(ctx, message, channelID); err != nil {
		log.Printf("Redisキャッシュに失敗: %v", err)
		// Redis保存に失敗してもPostgreSQLには保存済みなので続行
	}

	return nil
}

// GetRecentMessages は、Redisから取得を試行し、失敗した場合はPostgreSQLから取得します
func (r *HybridConversationRepository) GetRecentMessages(ctx context.Context, channelID domain.ChannelID, limit int) (domain.ConversationHistory, error) {
	// まずRedisからキャッシュを取得
	history, err := r.redis.GetRecentMessages(ctx, channelID, limit)
	if err == nil && !history.IsEmpty() {
		log.Printf("Redisキャッシュからメッセージを取得: %d件", history.Count())
		return history, nil
	}

	// Redisから取得できない場合はPostgreSQLから取得
	log.Printf("Redisキャッシュから取得できないため、PostgreSQLから取得します")
	history, err = r.postgres.GetRecentMessages(ctx, channelID, limit)
	if err != nil {
		return domain.ConversationHistory{}, fmt.Errorf("メッセージ取得に失敗: %w", err)
	}

	// PostgreSQLから取得したデータをRedisにキャッシュ
	if !history.IsEmpty() {
		for _, message := range history.Messages() {
			if err := r.redis.CacheMessage(ctx, message, channelID); err != nil {
				log.Printf("Redisキャッシュ更新に失敗: %v", err)
			}
		}
	}

	return history, nil
}

// GetThreadMessages は、スレッドの全メッセージを取得します
func (r *HybridConversationRepository) GetThreadMessages(ctx context.Context, threadID domain.ChannelID) (domain.ConversationHistory, error) {
	// スレッドの場合はPostgreSQLから直接取得（キャッシュは使用しない）
	return r.postgres.GetThreadMessages(ctx, threadID)
}

// GetMessagesBefore は、指定されたメッセージIDより前のメッセージを取得します
func (r *HybridConversationRepository) GetMessagesBefore(ctx context.Context, channelID domain.ChannelID, messageID string, limit int) (domain.ConversationHistory, error) {
	// この操作はPostgreSQLから直接取得
	return r.postgres.GetMessagesBefore(ctx, channelID, messageID, limit)
}

// SearchMessages は、メッセージ内容で検索します
func (r *HybridConversationRepository) SearchMessages(ctx context.Context, channelID domain.ChannelID, query string, limit int) (domain.ConversationHistory, error) {
	// 検索はPostgreSQLから直接取得
	return r.postgres.SearchMessages(ctx, channelID, query, limit)
}

// GetUserHistory は、ユーザーの会話履歴を取得します
func (r *HybridConversationRepository) GetUserHistory(ctx context.Context, userID domain.UserID, limit int) (domain.ConversationHistory, error) {
	// ユーザー履歴はRedisから取得
	return r.redis.GetUserHistory(ctx, userID, limit)
}

// Close は、両方のデータベース接続を閉じます
func (r *HybridConversationRepository) Close() error {
	var errors []error

	if err := r.postgres.Close(); err != nil {
		errors = append(errors, fmt.Errorf("PostgreSQL接続終了エラー: %w", err))
	}

	if err := r.redis.Close(); err != nil {
		errors = append(errors, fmt.Errorf("Redis接続終了エラー: %w", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("データベース接続終了エラー: %v", errors)
	}

	return nil
}
