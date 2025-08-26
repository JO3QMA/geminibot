# Discord-Gemini連携Bot

Discord上で動作し、Googleの生成AIモデル「Gemini」と連携するBotです。ユーザーからのメンションをトリガーとして、指定された範囲のチャット履歴を文脈（コンテキスト）としてGeminiに渡し、その応答をDiscordに投稿する機能を提供します。

## 🚀 機能

- Discordチャンネルまたはスレッドでのメンションによる起動
- チャット履歴の自動取得（通常チャンネル：直近10件、スレッド：全メッセージ）
- Gemini APIとの連携によるAI応答生成
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

| 環境変数 | 説明 | デフォルト値 |
|---------|------|-------------|
| `DISCORD_BOT_TOKEN` | Discord Bot Token | - |
| `GEMINI_API_KEY` | Gemini API Key | - |
| `BOT_PREFIX` | Botのメンション接頭辞 | `@` |
| `MAX_HISTORY_MESSAGES` | 取得する履歴メッセージ数 | `10` |
| `REQUEST_TIMEOUT` | APIリクエストタイムアウト | `30s` |

## 📝 ライセンス

MIT License

## 🤝 コントリビューション

プルリクエストやイシューの報告を歓迎します！
