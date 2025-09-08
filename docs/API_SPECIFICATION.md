# GeminiBot API仕様書

## 概要

GeminiBotのAPI仕様書です。Discord Bot API、Gemini API、および内部APIの仕様を定義します。

## Discord Bot API

### 1. メンション機能

#### 1.1 メンション検出

**イベント**: `MessageCreate`

**条件**: メッセージにBotのメンションが含まれている

**処理フロー**:
1. メンション検出
2. チャット履歴取得
3. コンテキスト構築
4. Gemini API呼び出し
5. 応答メッセージ送信

#### 1.2 チャット履歴取得

**取得範囲**:
- 通常チャンネル: 直近10件のメッセージ
- スレッド: 全メッセージ

**制限事項**:
- 最大コンテキスト長: 8000文字
- 最大履歴長: 4000文字

### 2. スラッシュコマンド

#### 2.1 `/set-api`

**説明**: サーバー用のGemini APIキーを設定

**権限**: 管理者権限必須

**パラメータ**:
- `api-key` (string, 必須): Gemini APIキー

**レスポンス**:
- 成功: "✅ このサーバー用のGemini APIキーを設定しました。"
- 失敗: "❌ APIキーの設定に失敗しました: {エラー詳細}"

**エラーケース**:
- 管理者権限不足
- 無効なAPIキー形式
- データベース接続エラー

#### 2.2 `/del-api`

**説明**: サーバー用のGemini APIキーを削除

**権限**: 管理者権限必須

**パラメータ**: なし

**レスポンス**:
- 成功: "✅ このサーバー用のGemini APIキーを削除しました。"
- 失敗: "❌ APIキーの削除に失敗しました: {エラー詳細}"

#### 2.3 `/set-model`

**説明**: 使用するAIモデルを設定

**権限**: 管理者権限必須

**パラメータ**:
- `model` (string, 必須): 使用するAIモデル
  - `gemini-2.5-pro`
  - `gemini-2.0-flash`
  - `gemini-2.5-flash-lite`

**レスポンス**:
- 成功: "✅ このサーバーで使用するAIモデルを {model} に設定しました。"
- 失敗: "❌ モデルの設定に失敗しました: {エラー詳細}"

#### 2.4 `/status`

**説明**: APIキー設定状況を表示

**権限**: 全ユーザー

**パラメータ**: なし

**レスポンス**:

**APIキー設定済みの場合**:
```
📊 **サーバー設定状況**

✅ **APIキー**: 設定済み
👤 **設定者**: {username}
📅 **設定日**: {date}
🤖 **使用モデル**: {model}
```

**APIキー未設定の場合**:
```
📊 **サーバー設定状況**

❌ **APIキー**: 未設定（デフォルトを使用）
🤖 **使用モデル**: {model}（デフォルト）
```

## Gemini API

### 1. 生成リクエスト

#### 1.1 エンドポイント

**URL**: `https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent`

**メソッド**: POST

**認証**: API Key

#### 1.2 リクエストヘッダー

```
Content-Type: application/json
X-Goog-Api-Key: {API_KEY}
```

#### 1.3 リクエストボディ

```json
{
  "contents": [
    {
      "parts": [
        {
          "text": "{system_prompt}"
        }
      ],
      "role": "user"
    },
    {
      "parts": [
        {
          "text": "{conversation_history}"
        }
      ],
      "role": "user"
    },
    {
      "parts": [
        {
          "text": "{user_question}"
        }
      ],
      "role": "user"
    }
  ],
  "generationConfig": {
    "maxOutputTokens": 1000,
    "temperature": 0.7,
    "topP": 0.9,
    "topK": 40
  }
}
```

#### 1.4 レスポンス

**成功時**:
```json
{
  "candidates": [
    {
      "content": {
        "parts": [
          {
            "text": "{generated_text}"
          }
        ],
        "role": "model"
      },
      "finishReason": "STOP"
    }
  ],
  "usageMetadata": {
    "promptTokenCount": 100,
    "candidatesTokenCount": 50,
    "totalTokenCount": 150
  }
}
```

**エラー時**:
```json
{
  "error": {
    "code": 400,
    "message": "Invalid API key",
    "status": "INVALID_ARGUMENT"
  }
}
```

### 2. パラメータ

| パラメータ | 説明 | デフォルト値 | 範囲 |
|-----------|------|-------------|------|
| `maxOutputTokens` | 最大出力トークン数 | 1000 | 1-8192 |
| `temperature` | 生成の温度パラメータ | 0.7 | 0.0-2.0 |
| `topP` | Top-Pサンプリング | 0.9 | 0.0-1.0 |
| `topK` | Top-Kサンプリング | 40 | 1-100 |

## 内部API

### 1. ドメインサービス

#### 1.1 ContextManager

**目的**: コンテキストの構築と管理

**主要メソッド**:

```go
type ContextManager struct {
    maxContextLength int
    maxHistoryLength int
}

func (cm *ContextManager) TruncateConversationHistory(history []Message) []Message
func (cm *ContextManager) TruncateSystemPrompt(systemPrompt string) string
func (cm *ContextManager) TruncateUserQuestion(userQuestion string) string
func (cm *ContextManager) GetContextStats(systemPrompt string, history []Message, userQuestion string) ContextStats
```

**TruncateConversationHistory**:
- 入力: メッセージ配列
- 出力: 切り詰められたメッセージ配列
- 機能: 新しいメッセージから優先的に保持

**TruncateSystemPrompt**:
- 入力: システムプロンプト文字列
- 出力: 切り詰められたシステムプロンプト
- 機能: 完全な文で終わるように調整

**GetContextStats**:
- 入力: システムプロンプト、履歴、質問
- 出力: ContextStats構造体
- 機能: コンテキストの統計情報を提供

#### 1.2 PromptGenerator

**目的**: プロンプトの生成

**主要メソッド**:

```go
type PromptGenerator struct {
    systemPrompt string
}

func (pg *PromptGenerator) GeneratePrompt(history []Message, userQuestion string) Prompt
func (pg *PromptGenerator) GeneratePromptWithMention(history []Message, userQuestion string, mentionerName string, mentionerID string) Prompt
func (pg *PromptGenerator) GeneratePromptWithContext(history []Message, userQuestion string, additionalContext string) Prompt
```

**GeneratePrompt**:
- 入力: 履歴、ユーザー質問
- 出力: Prompt構造体
- 機能: 基本的なプロンプト生成

**GeneratePromptWithMention**:
- 入力: 履歴、ユーザー質問、メンション情報
- 出力: Prompt構造体
- 機能: メンション情報を含むプロンプト生成

### 2. アプリケーションサービス

#### 2.1 MentionApplicationService

**目的**: メンション処理のユースケース実装

**主要メソッド**:

```go
type MentionApplicationService struct {
    conversationRepo    domain.ConversationRepository
    promptGenerator     *domain.PromptGenerator
    geminiClient        GeminiClient
    contextManager      *domain.ContextManager
    config              *config.BotConfig
    apiKeyService       *APIKeyApplicationService
    defaultGeminiConfig *config.GeminiConfig
    geminiClientFactory func(apiKey string) (GeminiClient, error)
}

func (s *MentionApplicationService) HandleMention(ctx context.Context, mention domain.BotMention) (string, error)
```

**HandleMention**:
- 入力: コンテキスト、BotMention
- 出力: 生成されたテキスト
- 機能: メンション処理の全体的な制御

#### 2.2 APIKeyApplicationService

**目的**: APIキー管理のユースケース実装

**主要メソッド**:

```go
type APIKeyApplicationService struct {
    apiKeyRepo domain.GuildAPIKeyRepository
}

func (s *APIKeyApplicationService) SetGuildAPIKey(ctx context.Context, guildID, apiKey, setBy string) error
func (s *APIKeyApplicationService) DeleteGuildAPIKey(ctx context.Context, guildID string) error
func (s *APIKeyApplicationService) GetGuildAPIKey(ctx context.Context, guildID string) (string, error)
func (s *APIKeyApplicationService) SetGuildModel(ctx context.Context, guildID, model string) error
func (s *APIKeyApplicationService) GetGuildModel(ctx context.Context, guildID string) (string, error)
func (s *APIKeyApplicationService) HasGuildAPIKey(ctx context.Context, guildID string) (bool, error)
func (s *APIKeyApplicationService) GetGuildAPIKeyInfo(ctx context.Context, guildID string) (domain.GuildAPIKey, error)
```

**SetGuildAPIKey**:
- 入力: コンテキスト、ギルドID、APIキー、設定者
- 出力: エラー
- 機能: ギルド固有のAPIキーを設定

**GetGuildModel**:
- 入力: コンテキスト、ギルドID
- 出力: モデル名、エラー
- 機能: ギルド固有のモデル設定を取得

### 3. リポジトリ

#### 3.1 GuildAPIKeyRepository

**目的**: APIキーデータの永続化

**主要メソッド**:

```go
type GuildAPIKeyRepository interface {
    SetAPIKey(ctx context.Context, guildID, apiKey, setBy string) error
    DeleteAPIKey(ctx context.Context, guildID string) error
    GetAPIKey(ctx context.Context, guildID string) (string, error)
    HasAPIKey(ctx context.Context, guildID string) (bool, error)
    GetGuildAPIKeyInfo(ctx context.Context, guildID string) (GuildAPIKey, error)
    SetGuildModel(ctx context.Context, guildID, model string) error
    GetGuildModel(ctx context.Context, guildID string) (string, error)
}
```

#### 3.2 ConversationRepository

**目的**: 会話履歴データの取得

**主要メソッド**:

```go
type ConversationRepository interface {
    GetRecentMessages(ctx context.Context, channelID string, limit int) ([]Message, error)
    GetThreadMessages(ctx context.Context, threadID string) ([]Message, error)
    GetMessagesBefore(ctx context.Context, channelID string, messageID string, limit int) ([]Message, error)
}
```

## エラーハンドリング

### 1. エラーコード

| コード | 説明 | HTTPステータス |
|-------|------|---------------|
| `INVALID_API_KEY` | 無効なAPIキー | 401 |
| `RATE_LIMIT_EXCEEDED` | レート制限超過 | 429 |
| `CONTEXT_TOO_LONG` | コンテキストが長すぎる | 400 |
| `PERMISSION_DENIED` | 権限不足 | 403 |
| `INTERNAL_ERROR` | 内部エラー | 500 |

### 2. エラーレスポンス形式

```json
{
  "error": {
    "code": "INVALID_API_KEY",
    "message": "指定されたAPIキーが無効です",
    "details": "APIキーの形式を確認してください"
  }
}
```

## レート制限

### 1. Discord API

- **メッセージ送信**: 5回/秒
- **チャット履歴取得**: 50回/秒
- **スラッシュコマンド**: 制限なし

### 2. Gemini API

- **生成リクエスト**: 60回/分
- **トークン数**: 1,000,000トークン/日

## セキュリティ

### 1. 認証

- Discord Bot Tokenによる認証
- Gemini API Keyによる認証
- 管理者権限によるアクセス制御

### 2. データ保護

- APIキーの暗号化保存
- ログからの機密情報除外
- HTTPS通信の強制

### 3. 入力検証

- APIキー形式の検証
- モデル名の検証
- コンテキスト長の制限

## 監視・ログ

### 1. ログレベル

- **INFO**: 通常の処理ログ
- **WARN**: 警告レベルのログ
- **ERROR**: エラーレベルのログ

### 2. ログ出力項目

- タイムスタンプ
- ログレベル
- メッセージ
- エラー詳細
- ユーザーID
- サーバーID
- チャンネルID

### 3. メトリクス

- API呼び出し回数
- レスポンス時間
- エラー率
- トークン使用量
