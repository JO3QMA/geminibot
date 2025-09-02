# Discord-Gemini連携Bot

Discord上で動作し、Googleの生成AIモデル「Gemini」と連携するBotです。ユーザーからのメンションをトリガーとして、指定された範囲のチャット履歴を文脈（コンテキスト）としてGeminiに渡し、その応答をDiscordに投稿する機能を提供します。

## 🚀 機能

- Discordチャンネルまたはスレッドでのメンションによる起動
- チャット履歴の自動取得（通常チャンネル：直近10件、スレッド：全メッセージ）
- Gemini APIとの連携によるAI応答生成
- **構造化コンテキスト機能**: genaiライブラリの機能を活用した高度なコンテキスト管理
- エラーハンドリングとログ記録

## 🏗️ アーキテクチャ

本Botは**ドメイン駆動設計 (DDD)** のアプローチを採用しています：

- **ドメイン層**: ビジネスロジック（会話履歴、プロンプト生成など）
- **アプリケーション層**: ユースケース実装（メンション処理など）
- **インフラストラクチャ層**: Discord API、Gemini APIとの通信
- **プレゼンテーション層**: Discordイベントハンドリング

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
├── internal/               # 内部パッケージ
│   ├── domain/            # ドメイン層
│   ├── application/       # アプリケーション層
│   ├── infrastructure/    # インフラストラクチャ層
│   └── presentation/      # プレゼンテーション層
├── pkg/                   # 公開パッケージ
├── configs/               # 設定ファイル
├── .devcontainer/         # Devcontainer設定
├── Dockerfile             # 本番用Dockerfile
├── docker-compose.yml     # 開発用Docker Compose
└── README.md
```

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
| `GEMINI_MODEL_NAME` | Geminiモデル名 | `gemini-2.5-pro` |
| `GEMINI_MAX_TOKENS` | 最大トークン数 | `1000` |
| `GEMINI_TEMPERATURE` | 生成の温度パラメータ | `0.7` |
| `GEMINI_TOP_P` | Top-Pサンプリング | `0.9` |
| `GEMINI_TOP_K` | Top-Kサンプリング | `40` |
| `MAX_HISTORY_MESSAGES` | 取得する履歴メッセージ数 | `10` |
| `REQUEST_TIMEOUT` | APIリクエストタイムアウト | `30s` |
| `SYSTEM_PROMPT` | システムプロンプト | デフォルトのアシスタントプロンプト |

## 📝 ライセンス

MIT License

## 🤝 コントリビューション

プルリクエストやイシューの報告を歓迎します！
