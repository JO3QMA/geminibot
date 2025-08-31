FROM mcr.microsoft.com/devcontainers/go:1.24

# 追加のツールをインストール
RUN apt-get update && apt-get install -y \
    postgresql-client \
    redis-tools \
    && rm -rf /var/lib/apt/lists/*

# Goツールのインストール
RUN go install golang.org/x/tools/cmd/goimports@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
    && go install golang.org/x/tools/gopls@latest

# 作業ディレクトリを設定
WORKDIR /workspaces/geminibot

# ユーザーをvscodeに変更
USER vscode
