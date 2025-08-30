# Discord-Gemini連携Bot

Discord上で動作し、Googleの生成AIモデル「Gemini」と連携するBotです。ユーザーからのメンションをトリガーとして、指定された範囲のチャット履歴を文脈（コンテキスト）としてGeminiに渡し、その応答をDiscordに投稿する機能を提供します。

## 🚀 機能

- Discordチャンネルまたはスレッドでのメンションによる起動
- チャット履歴の自動取得（通常チャンネル：直近10件、スレッド：全メッセージ）
- Gemini APIとの連携によるAI応答生成
- **構造化コンテキスト機能**: genaiライブラリの機能を活用した高度なコンテキスト管理
- **リファクタリング済みアーキテクチャ**: 単一責任の原則に基づく責務分離
- エラーハンドリングとログ記録
- Markdown変換とメッセージ分割機能

## 🏗️ アーキテクチャ

本Botは**ドメイン駆動設計 (DDD)** のアプローチを採用し、**単一責任の原則**に基づいてリファクタリングされています：

### レイヤー構造
- **ドメイン層**: ビジネスロジック（会話履歴、プロンプト生成、コンテキスト管理など）
- **アプリケーション層**: ユースケース実装（メンション処理サービスなど）
- **インフラストラクチャ層**: Discord API、Gemini APIとの通信
- **プレゼンテーション層**: Discordイベントハンドリング（責務別に分割）

### リファクタリング後の構造
```
internal/presentation/discord/
├── handler_new.go      # 新しい統括ハンドラー
├── processor.go        # メッセージ処理ロジック
├── sender.go          # メッセージ送信ロジック
├── formatter.go       # Markdown変換とフォーマット
├── error_handler.go   # エラー処理とフォーマット
└── handler.go         # 旧ハンドラー（後方互換性）
```

### 責務分離の利点
- **可読性の向上**: 巨大な747行から複数の小さなファイルに分割
- **テスタビリティ**: 各コンポーネントを独立してテスト可能
- **再利用性**: コンポーネントが独立して再利用可能
- **メンテナンス性**: バグの特定と修正が容易
- **依存注入パターン**: 疎結合な設計を実現

## 🛠️ 技術スタック

- **言語**: Go (Golang) 1.21
- **開発環境**: Devcontainer (Visual Studio Code)
- **実行環境**: Docker
- **主要ライブラリ**:
  - `discordgo`: Discord APIクライアント
  - `generative-ai-go`: Google Generative AI SDK

## 📋 セットアップ

### 1. 前提条件

- Docker & Docker Compose
- Visual Studio Code (Devcontainer対応)
- Discord Developer Portal アカウント
- Google Cloud Platform アカウント

### 2. 環境構築

1. リポジトリをクローン
```bash
git clone <repository-url>
cd geminibot
```

2. 環境変数を設定
```bash
cp env.example .env
# .envファイルを編集して認証情報を設定
```

3. Devcontainerで開発環境を起動
```bash
# VS Codeで Devcontainer: Reopen in Container を実行
```

**注意**: Devcontainer内では`.env`ファイルが自動的に読み込まれます。アプリケーション起動時に環境変数が正しく設定されているか確認してください。

## 🔧 コンテキスト管理機能

本Botは、**構造化コンテキスト機能**と**コンテキスト長制限機能**を提供します。

### 構造化コンテキスト機能

本Botは、genaiライブラリの構造化機能を活用した**構造化コンテキスト機能**を採用しています：

- **システムプロンプト**: Botの役割や行動指針を個別のContentパーツとして送信
- **会話履歴**: 過去のメッセージを構造化された形式で送信
- **ユーザー質問**: メンション時のメッセージ内容を個別に送信
- **効率的なコンテキスト管理**: 各部分を分離することで、Geminiがより適切に理解

### コンテキスト長制限機能

**問題の解決**:
- コンテキストが長すぎて支離滅裂な応答を防止
- 効率的なトークン使用
- 重要な情報の優先保持

**制限方法**:
- 新しいメッセージから優先的に保持
- 制限を超えた古いメッセージを自動削除
- 完全な文で終わるように調整

### 設定方法

```bash
# コンテキスト長制限（文字数）
MAX_CONTEXT_LENGTH=8000    # 最大コンテキスト長
MAX_HISTORY_LENGTH=4000    # 最大履歴長
```

### メリット

1. **より自然な会話フロー**: AIが会話の文脈をより自然に理解
2. **構造化されたコンテキスト**: システムプロンプト、履歴、質問を明確に分離
3. **効率的なトークン使用**: genaiライブラリの最適化機能を活用
4. **支離滅裂な応答の防止**: コンテキスト長制限による品質向上
5. **重要な情報の保持**: 新しいメッセージを優先的に保持
6. **設定の簡素化**: メッセージ数制限から文字数制限への統一

### 3. 認証情報の取得

#### Discord Bot Token
1. [Discord Developer Portal](https://discord.com/developers/applications) にアクセス
2. 新しいアプリケーションを作成
3. BotセクションでBotを作成
4. Tokenをコピーして`.env`ファイルに設定

#### Gemini API Key
1. [Google AI Studio](https://makersuite.google.com/app/apikey) にアクセス
2. API Keyを作成
3. キーをコピーして`.env`ファイルに設定

## 🚀 実行

### 開発環境
```bash
go run cmd/main.go
```

### Docker環境
```bash
docker-compose up --build
```

## 📁 プロジェクト構造

```
geminibot/
├── cmd/                    # アプリケーションエントリーポイント
│   └── main.go            # メインアプリケーション（新しいハンドラー使用）
├── internal/               # 内部パッケージ
│   ├── domain/            # ドメイン層
│   │   ├── service.go     # プロンプト生成サービス
│   │   ├── context_manager.go # コンテキスト管理
│   │   └── value_objects.go   # 値オブジェクト
│   ├── application/       # アプリケーション層
│   │   └── mention_service.go # メンション処理サービス
│   ├── infrastructure/    # インフラストラクチャ層
│   │   ├── discord/       # Discord API連携
│   │   ├── gemini/        # Gemini API連携
│   │   └── config/        # 設定管理
│   └── presentation/      # プレゼンテーション層
│       └── discord/       # Discordイベントハンドリング
│           ├── handler_new.go      # 新しい統括ハンドラー
│           ├── processor.go        # メッセージ処理ロジック
│           ├── sender.go          # メッセージ送信ロジック
│           ├── formatter.go       # Markdown変換とフォーマット
│           ├── error_handler.go   # エラー処理とフォーマット
│           └── handler.go         # 旧ハンドラー（後方互換性）
├── configs/               # 設定ファイル
├── .devcontainer/         # Devcontainer設定
├── Dockerfile             # 本番用Dockerfile
├── docker-compose.yml     # 開発用Docker Compose
└── README.md
```

## 🔄 リファクタリング詳細

### リファクタリングの背景
元のDiscordハンドラーは747行の巨大なファイルで、以下の問題がありました：
- 単一責任の原則違反
- テストの困難さ
- コードの可読性の低下
- メンテナンス性の悪化

### リファクタリングの成果
**ハンドラーの責務分離**により、以下の改善を実現：

#### 分割されたコンポーネント
1. **`handler_new.go`** (78行) - 新しい統括ハンドラー
2. **`processor.go`** (182行) - メッセージ処理ロジック
3. **`sender.go`** (242行) - メッセージ送信ロジック
4. **`formatter.go`** (315行) - Markdown変換とフォーマット
5. **`error_handler.go`** (73行) - エラー処理とフォーマット

#### 改善された点
- **可読性**: 巨大なファイルから理解しやすい小さなファイルに分割
- **テスタビリティ**: 各コンポーネントを独立してテスト可能
- **再利用性**: コンポーネントが独立して再利用可能
- **メンテナンス性**: バグの特定と修正が容易
- **依存注入**: 疎結合な設計を実現

### 後方互換性
旧ハンドラー（`handler.go`）は後方互換性のために残されており、必要に応じて段階的に移行できます。

## 🔧 設定

### 環境変数の設定

アプリケーションは`.env`ファイルから環境変数を自動的に読み込みます。開発時は`env.example`をコピーして`.env`ファイルを作成し、必要な認証情報を設定してください。

```bash
cp env.example .env
# .envファイルを編集して認証情報を設定
```

### 環境変数一覧

| 環境変数 | 説明 | デフォルト値 |
|---------|------|-------------|
| `DISCORD_BOT_TOKEN` | Discord Bot Token | - |
| `GEMINI_API_KEY` | Gemini API Key | - |
| `GEMINI_MODEL_NAME` | Geminiモデル名 | `gemini-pro` |
| `GEMINI_MAX_TOKENS` | 最大トークン数 | `1000` |
| `GEMINI_TEMPERATURE` | 生成の温度パラメータ | `0.7` |
| `GEMINI_TOP_P` | Top-Pサンプリング | `0.9` |
| `GEMINI_TOP_K` | Top-Kサンプリング | `40` |
| `MAX_CONTEXT_LENGTH` | 最大コンテキスト長（文字数） | `8000` |
| `MAX_HISTORY_LENGTH` | 最大履歴長（文字数） | `4000` |
| `REQUEST_TIMEOUT` | APIリクエストタイムアウト | `30s` |
| `SYSTEM_PROMPT` | システムプロンプト | デフォルトのアシスタントプロンプト |

## 📝 ライセンス

MIT License

## 🧪 開発・テスト

### テスト実行
```bash
# 全体のテスト
go test ./... -v

# 特定のパッケージのテスト
go test ./internal/domain/... -v
go test ./internal/application/... -v
go test ./internal/infrastructure/... -v
go test ./internal/presentation/... -v
```

### Linter実行
```bash
golangci-lint run
```

### ビルド確認
```bash
go build ./cmd/main.go
```

## 🤝 コントリビューション

プルリクエストやイシューの報告を歓迎します！

### 開発ガイドライン
- **単一責任の原則**に従ったコード設計
- **テストカバレッジ**の維持
- **Linterエラー**の解消
- **ドキュメント**の更新
