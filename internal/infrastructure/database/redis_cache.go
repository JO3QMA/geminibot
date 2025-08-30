package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"geminibot/internal/domain"

	"github.com/redis/go-redis/v9"
)

// RedisConversationCache は、Redisを使用した会話履歴キャッシュです
type RedisConversationCache struct {
	client *redis.Client
}

// NewRedisConversationCache は新しいRedisConversationCacheインスタンスを作成します
func NewRedisConversationCache(addr string, password string, db int) (*RedisConversationCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 接続テスト
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis接続に失敗: %w", err)
	}

	return &RedisConversationCache{client: client}, nil
}

// CacheMessage は、メッセージをキャッシュに保存します
func (r *RedisConversationCache) CacheMessage(ctx context.Context, message domain.Message, channelID domain.ChannelID) error {
	// メッセージ詳細をHashで保存
	messageKey := fmt.Sprintf("message:%s", message.ID)
	messageData := map[string]interface{}{
		"id":           message.ID,
		"content":      message.Content,
		"user_id":      message.User.ID.String(),
		"channel_id":   channelID.String(),
		"timestamp":    message.Timestamp.Format(time.RFC3339),
		"username":     message.User.Username,
		"display_name": message.User.DisplayName,
		"avatar":       message.User.Avatar,
		"is_bot":       message.User.IsBot,
	}

	// Hashでメッセージ詳細を保存
	if err := r.client.HSet(ctx, messageKey, messageData).Err(); err != nil {
		return fmt.Errorf("メッセージ詳細保存に失敗: %w", err)
	}

	// 24時間でTTL設定
	r.client.Expire(ctx, messageKey, 24*time.Hour)

	// チャンネル別の最新メッセージリストに追加
	channelKey := fmt.Sprintf("channel:%s:recent", channelID.String())
	messageJSON, _ := json.Marshal(messageData)

	// Listの先頭に追加
	if err := r.client.LPush(ctx, channelKey, messageJSON).Err(); err != nil {
		return fmt.Errorf("チャンネル履歴保存に失敗: %w", err)
	}

	// 最新50件に制限
	r.client.LTrim(ctx, channelKey, 0, 49)

	// ユーザー別の履歴に追加（Sorted Set）
	userKey := fmt.Sprintf("user:%s:history", message.User.ID.String())
	timestamp := float64(message.Timestamp.Unix())

	if err := r.client.ZAdd(ctx, userKey, redis.Z{
		Score:  timestamp,
		Member: message.ID,
	}).Err(); err != nil {
		return fmt.Errorf("ユーザー履歴保存に失敗: %w", err)
	}

	// 最新100件に制限
	r.client.ZRemRangeByRank(ctx, userKey, 0, -101)

	// セッション情報を更新
	sessionKey := fmt.Sprintf("session:user:%s", message.User.ID.String())
	sessionData := map[string]interface{}{
		"last_activity":  timestamp,
		"channel_id":     channelID.String(),
		"message_count":  r.client.ZCard(ctx, userKey).Val(),
		"context_length": r.getContextLength(ctx, userKey),
	}

	r.client.HSet(ctx, sessionKey, sessionData)
	r.client.Expire(ctx, sessionKey, 1*time.Hour)

	log.Printf("メッセージをキャッシュに保存しました: %s", message.ID)
	return nil
}

// GetRecentMessages は、チャンネルの最新メッセージをキャッシュから取得します
func (r *RedisConversationCache) GetRecentMessages(ctx context.Context, channelID domain.ChannelID, limit int) (domain.ConversationHistory, error) {
	channelKey := fmt.Sprintf("channel:%s:recent", channelID.String())

	// Listから最新メッセージを取得
	messageJSONs, err := r.client.LRange(ctx, channelKey, 0, int64(limit-1)).Result()
	if err != nil {
		return domain.ConversationHistory{}, fmt.Errorf("キャッシュからメッセージ取得に失敗: %w", err)
	}

	var messages []domain.Message
	for _, messageJSON := range messageJSONs {
		var messageData map[string]interface{}
		if err := json.Unmarshal([]byte(messageJSON), &messageData); err != nil {
			continue
		}

		// メッセージオブジェクトに変換
		message := r.convertToMessage(messageData)
		messages = append(messages, message)
	}

	// 時系列順に並び替え（古い順）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return domain.NewConversationHistory(messages), nil
}

// GetUserHistory は、ユーザーの会話履歴をキャッシュから取得します
func (r *RedisConversationCache) GetUserHistory(ctx context.Context, userID domain.UserID, limit int) (domain.ConversationHistory, error) {
	userKey := fmt.Sprintf("user:%s:history", userID.String())

	// Sorted Setから最新メッセージIDを取得
	messageIDs, err := r.client.ZRevRange(ctx, userKey, 0, int64(limit-1)).Result()
	if err != nil {
		return domain.ConversationHistory{}, fmt.Errorf("ユーザー履歴取得に失敗: %w", err)
	}

	var messages []domain.Message
	for _, messageID := range messageIDs {
		messageKey := fmt.Sprintf("message:%s", messageID)
		messageData, err := r.client.HGetAll(ctx, messageKey).Result()
		if err != nil {
			continue
		}

		// メッセージオブジェクトに変換
		message := r.convertFromHash(messageData)
		messages = append(messages, message)
	}

	return domain.NewConversationHistory(messages), nil
}

// convertToMessage は、JSONデータをMessageオブジェクトに変換します
func (r *RedisConversationCache) convertToMessage(data map[string]interface{}) domain.Message {
	id, _ := data["id"].(string)
	content, _ := data["content"].(string)
	userID, _ := data["user_id"].(string)
	timestampStr, _ := data["timestamp"].(string)
	username, _ := data["username"].(string)
	displayName, _ := data["display_name"].(string)
	avatar, _ := data["avatar"].(string)
	isBot, _ := data["is_bot"].(bool)

	timestamp, _ := time.Parse(time.RFC3339, timestampStr)

	user := domain.NewUser(
		domain.NewUserID(userID),
		username,
		displayName,
		avatar,
		"",
		isBot,
	)

	return domain.NewMessage(id, user, content, timestamp)
}

// convertFromHash は、HashデータをMessageオブジェクトに変換します
func (r *RedisConversationCache) convertFromHash(data map[string]string) domain.Message {
	id := data["id"]
	content := data["content"]
	userID := data["user_id"]
	timestampStr := data["timestamp"]
	username := data["username"]
	displayName := data["display_name"]
	avatar := data["avatar"]
	isBotStr := data["is_bot"]

	timestamp, _ := time.Parse(time.RFC3339, timestampStr)
	isBot := isBotStr == "true"

	user := domain.NewUser(
		domain.NewUserID(userID),
		username,
		displayName,
		avatar,
		"",
		isBot,
	)

	return domain.NewMessage(id, user, content, timestamp)
}

// getContextLength は、ユーザーのコンテキスト長を計算します
func (r *RedisConversationCache) getContextLength(ctx context.Context, userKey string) int {
	// 最新10件のメッセージの文字数を合計
	messageIDs, _ := r.client.ZRevRange(ctx, userKey, 0, 9).Result()

	totalLength := 0
	for _, messageID := range messageIDs {
		messageKey := fmt.Sprintf("message:%s", messageID)
		content, _ := r.client.HGet(ctx, messageKey, "content").Result()
		totalLength += len(content)
	}

	return totalLength
}

// Close は、Redis接続を閉じます
func (r *RedisConversationCache) Close() error {
	return r.client.Close()
}
