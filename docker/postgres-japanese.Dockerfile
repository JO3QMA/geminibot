FROM postgres:15-alpine

# 日本語全文検索機能を追加するためのパッケージをインストール
RUN apk add --no-cache \
    postgresql-contrib \
    && rm -rf /var/cache/apk/*

# 日本語全文検索の設定ファイルをコピー
COPY ./docker/postgresql.conf /etc/postgresql/postgresql.conf

# 初期化スクリプトをコピー
COPY ./docker/init-japanese.sh /docker-entrypoint-initdb.d/01-init-japanese.sh

# 実行権限を付与
RUN chmod +x /docker-entrypoint-initdb.d/01-init-japanese.sh
