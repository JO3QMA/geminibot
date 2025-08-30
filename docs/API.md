# API仕様書

## 概要

このドキュメントは、Discord-Gemini連携Botのリファクタリング後のAPI仕様を説明します。

## アーキテクチャ概要

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Discord API   │    │   Gemini API    │    │   Application   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Presentation    │    │ Infrastructure  │    │ Domain          │
│ Layer           │    │ Layer           │    │ Layer           │
│                 │    │                 │    │                 │
│ • handler_new   │    │ • discord       │    │ • service       │
│ • processor     │    │ • gemini        │    │ • context_mgr   │
│ • sender        │    │ • config        │    │ • value_objects │
│ • formatter     │    │                 │    │                 │
│ • error_handler │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## プレゼンテーション層

### DiscordHandlerNew

新しい統括ハンドラー。他のコンポーネントを統合してDiscordイベントを処理します。

```go
type DiscordHandlerNew struct {
    session        *discordgo.Session
    mentionService *application.MentionApplicationService
    processor      *MessageProcessor
    sender         *MessageSender
    errorHandler   *ErrorHandler
    formatter      *MessageFormatter
    botID          string
    botUsername    string
}

func NewDiscordHandlerNew(
    session *discordgo.Session,
    mentionService *application.MentionApplicationService,
    botID string,
) *DiscordHandlerNew
```

#### メソッド

- `SetupHandlers()` - Discordイベントハンドラーを設定
- `handleReady()` - Bot準備完了時の処理
- `handleMessageCreate()` - メッセージ作成時の処理

### MessageProcessor

メッセージ処理ロジックを担当します。

```go
type MessageProcessor struct {
    mentionService *application.MentionApplicationService
    sender         *MessageSender
    errorHandler   *ErrorHandler
}

func NewMessageProcessor(
    mentionService *application.MentionApplicationService,
    sender *MessageSender,
    errorHandler *ErrorHandler,
) *MessageProcessor
```

#### メソッド

- `ProcessMentionAsync(s, m, mention)` - メンションを非同期で処理
- `ProcessMentionNormal(s, m, mention)` - 通常のメンション処理
- `CreateBotMention(s, m)` - BotMentionオブジェクトを作成
- `IsMentioned(m, botID, botUsername)` - メンション判定
- `extractUserContent(m)` - ユーザーコンテンツを抽出
- `getDisplayName(s, m)` - 表示名を取得

### MessageSender

メッセージ送信ロジックを担当します。

```go
type MessageSender struct {
    formatter *MessageFormatter
}

func NewMessageSender(formatter *MessageFormatter) *MessageSender
```

#### メソッド

- `SendThreadResponse(session, threadID, response)` - スレッド内に応答を送信
- `SendSplitResponse(session, m, response)` - 分割して応答を送信
- `SendNormalReply(session, m, response)` - 通常のリプライを送信
- `SendThinkingMessage(session, m)` - 処理中メッセージを送信
- `SendThinkingMessageToThread(session, threadID)` - スレッド内に処理中メッセージを送信
- `sendAsFileToThread(session, threadID, content, filename)` - ファイルとして送信（スレッド）
- `sendAsFile(session, m, content, filename)` - ファイルとして送信（通常）
- `splitMessage(message)` - メッセージを分割

### MessageFormatter

Markdown変換とフォーマットを担当します。

```go
type MessageFormatter struct{}

func NewMessageFormatter() *MessageFormatter
```

#### メソッド

- `FormatForDiscord(response)` - Discord用にフォーマット
- `convertCodeBlocks(text)` - コードブロックを変換
- `convertInlineCode(text)` - インラインコードを変換
- `convertBold(text)` - 太字を変換
- `convertItalic(text)` - 斜体を変換
- `convertLists(text)` - リストを変換

### ErrorHandler

エラー処理とフォーマットを担当します。

```go
type ErrorHandler struct{}

func NewErrorHandler() *ErrorHandler
```

#### メソッド

- `IsTimeoutError(err)` - タイムアウトエラー判定
- `FormatError(err)` - エラーを適切なメッセージにフォーマット

## アプリケーション層

### MentionApplicationService

メンション処理のユースケースを実装します。

```go
type MentionApplicationService struct {
    conversationRepo domain.ConversationRepository
    promptGenerator  *domain.PromptGenerator
    geminiClient     GeminiClient
    contextManager   *domain.ContextManager
    config           *config.BotConfig
}

func NewMentionApplicationService(
    conversationRepo domain.ConversationRepository,
    geminiClient GeminiClient,
    botConfig *config.BotConfig,
) *MentionApplicationService
```

#### メソッド

- `HandleMention(ctx, mention)` - メンションを処理
- `getConversationHistory(ctx, mention)` - 会話履歴を取得

## ドメイン層

### PromptGenerator

プロンプト生成のビジネスロジックを担当します。

```go
type PromptGenerator struct {
    systemPrompt string
}

func NewPromptGenerator(systemPrompt string) *PromptGenerator
```

#### メソッド

- `GeneratePrompt(history, userQuestion)` - プロンプトを生成
- `GeneratePromptWithMention(history, userQuestion, mentionerName, mentionerID)` - メンション情報を含めてプロンプトを生成
- `GeneratePromptWithContext(history, userQuestion, additionalContext)` - 追加コンテキストを含めてプロンプトを生成

### ContextManager

コンテキストの長さを管理します。

```go
type ContextManager struct {
    maxContextLength int
    maxHistoryLength int
}

func NewContextManager(maxContextLength, maxHistoryLength int) *ContextManager
```

#### メソッド

- `TruncateConversationHistory(history)` - 会話履歴を切り詰め
- `TruncateSystemPrompt(systemPrompt)` - システムプロンプトを切り詰め
- `TruncateUserQuestion(userQuestion)` - ユーザー質問を切り詰め
- `GetContextStats(systemPrompt, history, userQuestion)` - コンテキスト統計を取得
- `calculateHistoryLength(messages)` - 履歴長を計算

### 値オブジェクト

#### BotMention

```go
type BotMention struct {
    ChannelID domain.ChannelID
    User      domain.User
    Content   string
    MessageID string
}

func NewBotMention(channelID, user, content, messageID) BotMention
```

#### Message

```go
type Message struct {
    ID        string
    User      domain.User
    Content   string
    Timestamp time.Time
}

func NewMessage(id, user, content, timestamp) Message
```

#### ConversationHistory

```go
type ConversationHistory struct {
    messages []Message
}

func NewConversationHistory(messages) ConversationHistory
```

## インフラストラクチャ層

### GeminiAPIClient

Gemini APIとの通信を担当します。

```go
type GeminiAPIClient struct {
    client *genai.Client
    config *config.GeminiConfig
}

func NewGeminiAPIClient(apiKey string, config *config.GeminiConfig) (*GeminiAPIClient, error)
```

#### メソッド

- `GenerateText(ctx, prompt)` - テキストを生成
- `GenerateTextWithOptions(ctx, prompt, options)` - オプション付きでテキストを生成
- `GenerateTextWithStructuredContext(ctx, systemPrompt, conversationHistory, userQuestion)` - 構造化コンテキストでテキストを生成
- `formatConversationHistory(messages)` - 会話履歴をフォーマット
- `Close()` - クライアントを閉じる

### DiscordConversationRepository

Discord APIを使用して会話履歴を取得します。

```go
type DiscordConversationRepository struct {
    session *discordgo.Session
}

func NewDiscordConversationRepository(session *discordgo.Session) *DiscordConversationRepository
```

#### メソッド

- `GetRecentMessages(ctx, channelID, limit)` - 最近のメッセージを取得
- `GetThreadMessages(ctx, threadID)` - スレッドメッセージを取得
- `GetMessagesBefore(ctx, channelID, messageID, limit)` - 指定メッセージ以前のメッセージを取得

## 設定

### BotConfig

```go
type BotConfig struct {
    MaxContextLength int
    MaxHistoryLength int
    RequestTimeout   time.Duration
    SystemPrompt     string
}
```

### GeminiConfig

```go
type GeminiConfig struct {
    APIKey      string
    ModelName   string
    MaxTokens   int32
    Temperature float32
    TopP        float32
    TopK        int32
}
```

### DiscordConfig

```go
type DiscordConfig struct {
    BotToken string
}
```

## エラーハンドリング

### エラー型

- `TimeoutError` - タイムアウトエラー
- `APIError` - API呼び出しエラー
- `ValidationError` - バリデーションエラー

### エラーフォーマット

エラーは適切なメッセージにフォーマットされ、ユーザーに分かりやすく表示されます：

- タイムアウトエラー: "⏰ **タイムアウトしました**"
- レート制限エラー: "⚠️ **レート制限を超過しました**"
- スパム検出: "🚫 **スパムが検出されました**"

## テスト

各コンポーネントは独立してテスト可能です：

```bash
# ドメイン層のテスト
go test ./internal/domain/... -v

# アプリケーション層のテスト
go test ./internal/application/... -v

# インフラストラクチャ層のテスト
go test ./internal/infrastructure/... -v

# プレゼンテーション層のテスト
go test ./internal/presentation/... -v
```

## パフォーマンス

### 最適化ポイント

1. **非同期処理**: メンション処理は非同期で実行
2. **コンテキスト制限**: 文字数制限による効率的なトークン使用
3. **メッセージ分割**: 長い応答の適切な分割
4. **エラーハンドリング**: 適切なエラー処理による安定性向上

### 制限値

- **最大コンテキスト長**: 8000文字
- **最大履歴長**: 4000文字
- **Discordメッセージ制限**: 2000文字
- **リクエストタイムアウト**: 30秒
