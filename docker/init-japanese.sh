#!/bin/bash
set -e

# 日本語全文検索機能の初期化
echo "Initializing Japanese text search configuration..."

# PostgreSQLに接続して日本語全文検索設定を作成
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- 日本語全文検索設定を作成
    CREATE TEXT SEARCH CONFIGURATION japanese (COPY = simple);
    
    -- 日本語の形態素解析器を設定
    ALTER TEXT SEARCH CONFIGURATION japanese ALTER MAPPING FOR asciiword, asciihword, hword_asciipart, word, hword, hword_part WITH simple;
    
    -- 日本語全文検索設定をデフォルトに設定
    ALTER DATABASE $POSTGRES_DB SET default_text_search_config = 'japanese';
    
    -- 設定を確認
    SELECT cfgname, cfgnamespace::regnamespace, cfgparser::regproc FROM pg_ts_config WHERE cfgname = 'japanese';
EOSQL

echo "Japanese text search configuration initialized successfully."
