# 開発者ガイド

## 概要

このドキュメントは、Discord-Gemini連携Botの開発者向けガイドです。リファクタリング後の新しいアーキテクチャに基づいて開発を行うための情報を提供します。

## 開発環境のセットアップ

### 前提条件

- Go 1.21以上
- Docker & Docker Compose
- Visual Studio Code (推奨)
- Git

### 環境構築

1. **リポジトリのクローン**
```bash
git clone <repository-url>
cd geminibot
```

2. **Devcontainerでの開発**
```bash
# VS Codeで Devcontainer: Reopen in Container を実行
```

3. **環境変数の設定**
```bash
cp env.example .env
# .envファイルを編集して認証情報を設定
```

## アーキテクチャ理解

### レイヤー構造

本プロジェクトは**ドメイン駆動設計 (DDD)** に基づいて設計されています：

```
┌─────────────────────────────────────────────────────────┐
│                    Presentation Layer                    │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐        │
│  │ handler_new │ │ processor   │ │ sender      │        │
│  └─────────────┘ └─────────────┘ └─────────────┘        │
│  ┌─────────────┐ ┌─────────────┐                        │
│  │ formatter   │ │error_handler│                        │
│  └─────────────┘ └─────────────┘                        │
└─────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────┐
│                  Application Layer                       │
│  ┌─────────────────────────────────────────────────┐    │
│  │           MentionApplicationService             │    │
│  └─────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────┐
│                     Domain Layer                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐        │
│  │   service   │ │context_mgr  │ │value_objects│        │
│  └─────────────┘ └─────────────┘ └─────────────┘        │
└─────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────┐
│                Infrastructure Layer                      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐        │
│  │   discord   │ │   gemini    │ │   config    │        │
│  └─────────────┘ └─────────────┘ └─────────────┘        │
└─────────────────────────────────────────────────────────┘
```

### 責務分離の原則

各コンポーネントは**単一責任の原則**に従って設計されています：

#### Presentation Layer
- **handler_new.go**: 統括ハンドラー（78行）
- **processor.go**: メッセージ処理ロジック（182行）
- **sender.go**: メッセージ送信ロジック（242行）
- **formatter.go**: Markdown変換とフォーマット（315行）
- **error_handler.go**: エラー処理とフォーマット（73行）

#### Application Layer
- **mention_service.go**: メンション処理のユースケース実装

#### Domain Layer
- **service.go**: プロンプト生成のビジネスロジック
- **context_manager.go**: コンテキスト管理
- **value_objects.go**: 値オブジェクト（BotMention、Message等）

#### Infrastructure Layer
- **discord/**: Discord API連携
- **gemini/**: Gemini API連携
- **config/**: 設定管理

## 開発ガイドライン

### コード設計原則

1. **単一責任の原則 (SRP)**
   - 各コンポーネントは1つの責任のみを持つ
   - 巨大なファイルは避け、適切に分割する

2. **依存注入パターン**
   - コンポーネント間の依存関係を明確にする
   - テスト可能な設計を心がける

3. **エラーハンドリング**
   - 適切なエラー処理を実装する
   - ユーザーに分かりやすいエラーメッセージを提供する

4. **テスト駆動開発 (TDD)**
   - 各コンポーネントにテストを書く
   - テストカバレッジを維持する

### 命名規則

- **ファイル名**: スネークケース（例: `message_processor.go`）
- **構造体名**: パスカルケース（例: `MessageProcessor`）
- **メソッド名**: パスカルケース（例: `ProcessMention`）
- **変数名**: キャメルケース（例: `userMessage`）
- **定数名**: パスカルケース（例: `DiscordMessageLimit`）

### コメント規則

```go
// MessageProcessor は、Discordメッセージの処理を担当します
type MessageProcessor struct {
    mentionService *application.MentionApplicationService
    sender         *MessageSender
    errorHandler   *ErrorHandler
}

// NewMessageProcessor は新しいMessageProcessorインスタンスを作成します
func NewMessageProcessor(
    mentionService *application.MentionApplicationService,
    sender *MessageSender,
    errorHandler *ErrorHandler,
) *MessageProcessor {
    return &MessageProcessor{
        mentionService: mentionService,
        sender:         sender,
        errorHandler:   errorHandler,
    }
}

// ProcessMentionAsync は、メンションを非同期で処理します
func (p *MessageProcessor) ProcessMentionAsync(s *discordgo.Session, m *discordgo.MessageCreate, mention domain.BotMention) {
    // 実装...
}
```

## テスト

### テスト実行

```bash
# 全体のテスト
go test ./... -v

# 特定のパッケージのテスト
go test ./internal/domain/... -v
go test ./internal/application/... -v
go test ./internal/infrastructure/... -v
go test ./internal/presentation/... -v

# テストカバレッジの確認
go test ./... -cover
```

### テストファイルの命名

- テストファイル名: `*_test.go`
- テスト関数名: `Test*`

### テスト例

```go
func TestMessageProcessor_ProcessMentionAsync(t *testing.T) {
    // テストのセットアップ
    mockService := &MockMentionService{}
    mockSender := &MockMessageSender{}
    mockErrorHandler := &MockErrorHandler{}
    
    processor := NewMessageProcessor(mockService, mockSender, mockErrorHandler)
    
    // テストケース
    t.Run("正常なメンション処理", func(t *testing.T) {
        // テスト実装
    })
    
    t.Run("エラー時の処理", func(t *testing.T) {
        // テスト実装
    })
}
```

### モックの使用

```go
// MockGeminiClient は、テスト用のGeminiClientモックです
type MockGeminiClient struct{}

func (m *MockGeminiClient) GenerateText(ctx context.Context, prompt domain.Prompt) (string, error) {
    return "テスト応答", nil
}

func (m *MockGeminiClient) GenerateTextWithStructuredContext(ctx context.Context, systemPrompt string, conversationHistory []domain.Message, userQuestion string) (string, error) {
    return "構造化コンテキストでの応答", nil
}
```

## デバッグ

### ログ出力

```go
import "log"

// デバッグ情報の出力
log.Printf("メンション処理開始: %s", mention.String())

// エラー情報の出力
log.Printf("エラーが発生しました: %v", err)
```

### 環境変数でのデバッグ制御

```bash
# デバッグモードの有効化
DEBUG=true go run cmd/main.go
```

## パフォーマンス最適化

### ベンチマークテスト

```go
func BenchmarkMessageProcessor_ProcessMention(b *testing.B) {
    // ベンチマークのセットアップ
    processor := NewMessageProcessor(mockService, mockSender, mockErrorHandler)
    
    for i := 0; i < b.N; i++ {
        // ベンチマーク対象の処理
        processor.ProcessMentionAsync(session, message, mention)
    }
}
```

### ベンチマーク実行

```bash
go test -bench=. ./internal/presentation/discord/
```

## リファクタリング

### リファクタリングの指針

1. **小さな変更を繰り返す**
   - 一度に大きな変更を避ける
   - テストを維持しながら段階的に変更

2. **テストの維持**
   - リファクタリング前後でテストが通ることを確認
   - 新しいテストケースを追加

3. **後方互換性の維持**
   - 既存のAPIとの互換性を保つ
   - 段階的な移行を可能にする

### リファクタリング例

```go
// リファクタリング前
func (h *DiscordHandler) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    // 747行の巨大な関数
}

// リファクタリング後
func (h *DiscordHandlerNew) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    // 統括処理のみ
    if h.processor.IsMentioned(m, h.botID, h.botUsername) {
        mention := h.processor.CreateBotMention(s, m)
        h.processor.ProcessMentionAsync(s, m, mention)
    }
}
```

## デプロイメント

### 開発環境

```bash
# 開発用の実行
go run cmd/main.go
```

### 本番環境

```bash
# Dockerでの実行
docker-compose up --build

# バイナリでの実行
go build ./cmd/main.go
./main
```

## トラブルシューティング

### よくある問題

1. **環境変数の読み込みエラー**
   ```bash
   # .envファイルの存在確認
   ls -la .env
   
   # 環境変数の確認
   echo $DISCORD_BOT_TOKEN
   ```

2. **テストエラー**
   ```bash
   # テストの詳細実行
   go test -v ./internal/domain/
   
   # テストカバレッジの確認
   go test -cover ./internal/domain/
   ```

3. **Linterエラー**
   ```bash
   # Linterの実行
   golangci-lint run
   
   # 特定のファイルのLinter実行
   golangci-lint run internal/presentation/discord/
   ```

### ログの確認

```bash
# アプリケーションログの確認
docker-compose logs -f

# 特定のサービスのログ確認
docker-compose logs -f app
```

## 貢献ガイド

### プルリクエストの作成

1. **ブランチの作成**
   ```bash
   git checkout -b feature/new-feature
   ```

2. **変更の実装**
   - 単一責任の原則に従った実装
   - 適切なテストの追加
   - ドキュメントの更新

3. **テストの実行**
   ```bash
   go test ./... -v
   golangci-lint run
   ```

4. **プルリクエストの作成**
   - 変更内容の詳細な説明
   - テスト結果の添付
   - スクリーンショット（必要に応じて）

### コードレビュー

- **可読性**: コードが理解しやすいか
- **テスタビリティ**: テストが適切に書かれているか
- **パフォーマンス**: パフォーマンスに問題がないか
- **セキュリティ**: セキュリティ上の問題がないか

## 参考資料

- [Go言語公式ドキュメント](https://golang.org/doc/)
- [Discord API ドキュメント](https://discord.com/developers/docs)
- [Google Generative AI ドキュメント](https://ai.google.dev/docs)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
