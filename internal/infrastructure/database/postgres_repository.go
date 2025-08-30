package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"geminibot/internal/domain"

	_ "github.com/lib/pq"
)

// PostgresConversationRepository は、PostgreSQLを使用してConversationRepositoryインターフェースを実装します
type PostgresConversationRepository struct {
	db *sql.DB
}

// NewPostgresConversationRepository は新しいPostgresConversationRepositoryインスタンスを作成します
func NewPostgresConversationRepository(connectionString string) (*PostgresConversationRepository, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("PostgreSQL接続に失敗: %w", err)
	}

	// 接続テスト
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("PostgreSQL接続テストに失敗: %w", err)
	}

	return &PostgresConversationRepository{db: db}, nil
}

// SaveMessage は、メッセージをデータベースに保存します
func (r *PostgresConversationRepository) SaveMessage(ctx context.Context, message domain.Message, channelID domain.ChannelID) error {
	// まずユーザーを保存または更新
	if err := r.saveUser(ctx, message.User); err != nil {
		return fmt.Errorf("ユーザー保存に失敗: %w", err)
	}

	// チャンネルを保存または更新
	if err := r.saveChannel(ctx, channelID, "text"); err != nil {
		return fmt.Errorf("チャンネル保存に失敗: %w", err)
	}

	// メッセージを保存
	query := `
		INSERT INTO messages (id, channel_id, user_id, content, timestamp, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			timestamp = EXCLUDED.timestamp,
			metadata = EXCLUDED.metadata
	`

	metadata := map[string]interface{}{
		"username":    message.User.Username,
		"displayName": message.User.DisplayName,
		"avatar":      message.User.Avatar,
		"isBot":       message.User.IsBot,
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("メタデータのJSON変換に失敗: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		message.ID,
		channelID.String(),
		message.User.ID.String(),
		message.Content,
		message.Timestamp,
		metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("メッセージ保存に失敗: %w", err)
	}

	log.Printf("メッセージを保存しました: %s", message.ID)
	return nil
}

// saveUser は、ユーザーを保存または更新します
func (r *PostgresConversationRepository) saveUser(ctx context.Context, user domain.User) error {
	query := `
		INSERT INTO users (id, username, display_name, avatar)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			display_name = EXCLUDED.display_name,
			avatar = EXCLUDED.avatar
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID.String(),
		user.Username,
		user.DisplayName,
		user.Avatar,
	)

	if err != nil {
		return fmt.Errorf("ユーザー保存に失敗: %w", err)
	}

	return nil
}

// saveChannel は、チャンネルを保存または更新します
func (r *PostgresConversationRepository) saveChannel(ctx context.Context, channelID domain.ChannelID, channelType string) error {
	query := `
		INSERT INTO channels (id, name, type)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type
	`

	_, err := r.db.ExecContext(ctx, query,
		channelID.String(),
		"channel-"+channelID.String(), // 実際の実装ではチャンネル名を取得する必要があります
		channelType,
	)

	if err != nil {
		return fmt.Errorf("チャンネル保存に失敗: %w", err)
	}

	return nil
}

// GetRecentMessages は、指定されたチャンネルの直近のメッセージを取得します
func (r *PostgresConversationRepository) GetRecentMessages(ctx context.Context, channelID domain.ChannelID, limit int) (domain.ConversationHistory, error) {
	query := `
		SELECT m.id, m.content, m.timestamp, m.metadata, u.username, u.display_name, u.avatar, u.id as user_id
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.channel_id = $1
		ORDER BY m.timestamp DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, channelID.String(), limit)
	if err != nil {
		return domain.ConversationHistory{}, fmt.Errorf("メッセージ取得に失敗: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var id, content, username, displayName, avatar, userID string
		var timestamp time.Time
		var metadataJSON []byte

		if err := rows.Scan(&id, &content, &timestamp, &metadataJSON, &username, &displayName, &avatar, &userID); err != nil {
			return domain.ConversationHistory{}, fmt.Errorf("メッセージ読み取りに失敗: %w", err)
		}

		user := domain.NewUser(
			domain.NewUserID(userID),
			username,
			displayName,
			avatar,
			"",
			false, // 実際の実装ではユーザー情報から取得
		)

		message := domain.NewMessage(id, user, content, timestamp)
		messages = append(messages, message)
	}

	// 時系列順に並び替え（古い順）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return domain.NewConversationHistory(messages), nil
}

// GetThreadMessages は、指定されたスレッドの全メッセージを取得します
func (r *PostgresConversationRepository) GetThreadMessages(ctx context.Context, threadID domain.ChannelID) (domain.ConversationHistory, error) {
	// スレッドの場合は十分な数のメッセージを取得（コンテキスト長制限で調整される）
	const threadMessageLimit = 200
	return r.GetRecentMessages(ctx, threadID, threadMessageLimit)
}

// GetMessagesBefore は、指定されたメッセージIDより前のメッセージを取得します
func (r *PostgresConversationRepository) GetMessagesBefore(ctx context.Context, channelID domain.ChannelID, messageID string, limit int) (domain.ConversationHistory, error) {
	query := `
		SELECT m.id, m.content, m.timestamp, m.metadata, u.username, u.display_name, u.avatar, u.id as user_id
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.channel_id = $1 AND m.timestamp < (
			SELECT timestamp FROM messages WHERE id = $2
		)
		ORDER BY m.timestamp DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, channelID.String(), messageID, limit)
	if err != nil {
		return domain.ConversationHistory{}, fmt.Errorf("メッセージ取得に失敗: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var id, content, username, displayName, avatar, userID string
		var timestamp time.Time
		var metadataJSON []byte

		if err := rows.Scan(&id, &content, &timestamp, &metadataJSON, &username, &displayName, &avatar, &userID); err != nil {
			return domain.ConversationHistory{}, fmt.Errorf("メッセージ読み取りに失敗: %w", err)
		}

		user := domain.NewUser(
			domain.NewUserID(userID),
			username,
			displayName,
			avatar,
			"",
			false,
		)

		message := domain.NewMessage(id, user, content, timestamp)
		messages = append(messages, message)
	}

	// 時系列順に並び替え（古い順）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return domain.NewConversationHistory(messages), nil
}

// SearchMessages は、メッセージ内容で検索します
func (r *PostgresConversationRepository) SearchMessages(ctx context.Context, channelID domain.ChannelID, query string, limit int) (domain.ConversationHistory, error) {
	searchQuery := `
		SELECT m.id, m.content, m.timestamp, m.metadata, u.username, u.display_name, u.avatar, u.id as user_id
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.channel_id = $1 AND to_tsvector('japanese', m.content) @@ plainto_tsquery('japanese', $2)
		ORDER BY m.timestamp DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, searchQuery, channelID.String(), query, limit)
	if err != nil {
		return domain.ConversationHistory{}, fmt.Errorf("メッセージ検索に失敗: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var id, content, username, displayName, avatar, userID string
		var timestamp time.Time
		var metadataJSON []byte

		if err := rows.Scan(&id, &content, &timestamp, &metadataJSON, &username, &displayName, &avatar, &userID); err != nil {
			return domain.ConversationHistory{}, fmt.Errorf("メッセージ読み取りに失敗: %w", err)
		}

		user := domain.NewUser(
			domain.NewUserID(userID),
			username,
			displayName,
			avatar,
			"",
			false,
		)

		message := domain.NewMessage(id, user, content, timestamp)
		messages = append(messages, message)
	}

	return domain.NewConversationHistory(messages), nil
}

// Close は、データベース接続を閉じます
func (r *PostgresConversationRepository) Close() error {
	return r.db.Close()
}
