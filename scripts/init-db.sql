-- ユーザーテーブル
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(20) PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    display_name VARCHAR(100),
    avatar VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- チャンネルテーブル
CREATE TABLE IF NOT EXISTS channels (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL, -- 'text', 'thread'
    parent_id VARCHAR(20), -- スレッドの場合の親チャンネル
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- メッセージテーブル
CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(20) PRIMARY KEY,
    channel_id VARCHAR(20) NOT NULL REFERENCES channels(id),
    user_id VARCHAR(20) NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    metadata JSONB, -- 追加情報
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- インデックス
CREATE INDEX IF NOT EXISTS idx_messages_channel_timestamp ON messages(channel_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_messages_user_timestamp ON messages(user_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_messages_content_gin ON messages USING gin(to_tsvector('japanese', content));

-- 日本語全文検索用の設定（既に作成済みの場合はスキップ）
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_ts_config WHERE cfgname = 'japanese') THEN
        CREATE TEXT SEARCH CONFIGURATION japanese (COPY = pg_catalog.simple);
    END IF;
END
$$;
