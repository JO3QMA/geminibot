# ========================================
# 🚀 Discord-Gemini Bot Makefile
# ========================================

# 変数定義
BINARY_NAME=geminibot
MAIN_PATH=cmd/main.go
BUILD_DIR=build
DOCKER_IMAGE=geminibot

# Go関連の変数
GO=go
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
GO_VERSION=$(shell go version | awk '{print $$3}')

# デフォルトターゲット
.DEFAULT_GOAL := help

# ========================================
# 🧪 テスト関連
# ========================================

.PHONY: test
test: ## すべてのテストを実行
	@echo "🧪 テストを実行中..."
	$(GO) test -v ./...

.PHONY: test-short
test-short: ## 短時間のテストを実行
	@echo "⚡ 短時間テストを実行中..."
	$(GO) test -v -short ./...

.PHONY: test-race
test-race: ## レースコンディション検出付きでテストを実行
	@echo "🏃 レースコンディション検出付きテストを実行中..."
	$(GO) test -race ./...

.PHONY: test-coverage
test-coverage: ## テストカバレッジを測定
	@echo "📊 テストカバレッジを測定中..."
	$(GO) test -cover ./...
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "📈 カバレッジレポート: coverage.html"

.PHONY: test-benchmark
test-benchmark: ## ベンチマークテストを実行
	@echo "⚡ ベンチマークテストを実行中..."
	$(GO) test -bench=. ./...

# ========================================
# 🏗️ ビルド関連
# ========================================

.PHONY: build
build: ## アプリケーションをビルド
	@echo "🔨 アプリケーションをビルド中..."
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

.PHONY: build-all
build-all: ## 全プラットフォーム向けにビルド
	@echo "🌍 全プラットフォーム向けにビルド中..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

.PHONY: clean
clean: ## ビルドファイルを削除
	@echo "🧹 ビルドファイルを削除中..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# ========================================
# 🔧 開発関連
# ========================================

.PHONY: run
run: ## アプリケーションを実行
	@echo "🚀 アプリケーションを実行中..."
	$(GO) run $(MAIN_PATH)

.PHONY: run-dev
run-dev: ## 開発モードでアプリケーションを実行
	@echo "🔧 開発モードでアプリケーションを実行中..."
	$(GO) run -race $(MAIN_PATH)

.PHONY: mod-tidy
mod-tidy: ## Goモジュールの依存関係を整理
	@echo "📦 Goモジュールの依存関係を整理中..."
	$(GO) mod tidy
	$(GO) mod download

.PHONY: mod-verify
mod-verify: ## Goモジュールの整合性を検証
	@echo "✅ Goモジュールの整合性を検証中..."
	$(GO) mod verify

.PHONY: lint
lint: ## コードの静的解析を実行
	@echo "🔍 コードの静的解析を実行中..."
	$(GO) vet ./...
	$(GO) fmt ./...

.PHONY: install-tools
install-tools: ## 開発ツールをインストール
	@echo "🛠️ 開発ツールをインストール中..."
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install golang.org/x/tools/gopls@latest

# ========================================
# 🐳 Docker関連
# ========================================

.PHONY: docker-build
docker-build: ## Dockerイメージをビルド
	@echo "🐳 Dockerイメージをビルド中..."
	docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-run
docker-run: ## Dockerコンテナを実行
	@echo "🐳 Dockerコンテナを実行中..."
	docker run --rm -p 8080:8080 $(DOCKER_IMAGE)

.PHONY: docker-clean
docker-clean: ## Dockerイメージを削除
	@echo "🧹 Dockerイメージを削除中..."
	docker rmi $(DOCKER_IMAGE) || true

# ========================================
# 📋 ヘルプ
# ========================================

.PHONY: help
help: ## 利用可能なコマンドを表示
	@echo "🚀 Discord-Gemini Bot - 利用可能なコマンド"
	@echo ""
	@echo "🧪 テスト関連:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "🏗️ ビルド関連:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "🔧 開発関連:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "🐳 Docker関連:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ========================================
# 🚀 クイックスタート
# ========================================

.PHONY: dev-setup
dev-setup: install-tools mod-tidy ## 開発環境をセットアップ

.PHONY: ci
ci: test-coverage lint build ## CI/CDパイプライン用のコマンド

.PHONY: all
all: clean test-coverage lint build ## すべてのタスクを実行
