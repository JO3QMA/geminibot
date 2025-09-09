# GeminiBot 開発者ガイド

## 概要

GeminiBotの開発に参加する開発者向けのガイドです。開発環境の構築から、コントリビューション方法まで説明します。

## 開発環境構築

### 1. 前提条件

- **Go**: 1.23以上
- **Docker**: 最新版
- **Docker Compose**: 最新版
- **Visual Studio Code**: Devcontainer対応版
- **Git**: 最新版

### 2. リポジトリのクローン

```bash
git clone https://github.com/your-org/geminibot.git
cd geminibot
```

### 3. 開発環境の起動

#### Devcontainerを使用する場合（推奨）

1. Visual Studio Codeでプロジェクトを開く
2. コマンドパレット（Ctrl+Shift+P）を開く
3. "Dev Containers: Reopen in Container"を選択
4. コンテナのビルドと起動を待つ

#### ローカル環境を使用する場合

```bash
# 依存関係のインストール
go mod download

# 環境変数の設定
cp env.example .env
# .envファイルを編集して認証情報を設定

# アプリケーションの起動
go run cmd/main.go
```

### 4. 認証情報の設定

#### Discord Bot Token

1. [Discord Developer Portal](https://discord.com/developers/applications)にアクセス
2. 新しいアプリケーションを作成
3. BotセクションでBotを作成
4. Tokenをコピーして`.env`ファイルに設定

#### Gemini API Key

1. [Google AI Studio](https://makersuite.google.com/app/apikey)にアクセス
2. API Keyを作成
3. キーをコピーして`.env`ファイルに設定

## プロジェクト構造

### ディレクトリ構成

```
geminibot/
├── cmd/                    # アプリケーションエントリーポイント
│   └── main.go
├── internal/               # 内部パッケージ
│   ├── domain/            # ドメイン層
│   ├── application/       # アプリケーション層
│   ├── infrastructure/    # インフラストラクチャ層
│   └── presentation/      # プレゼンテーション層
├── pkg/                   # 公開パッケージ
├── configs/               # 設定ファイル
├── docs/                  # ドキュメント
├── logs/                  # ログファイル
├── scripts/               # スクリプト
└── tests/                 # テストファイル
```

### 命名規則

#### ファイル名
- スネークケース: `api_key_service.go`
- テストファイル: `*_test.go`

#### パッケージ名
- 小文字: `domain`, `application`
- 単数形: `service`, `repository`

#### 構造体名
- パスカルケース: `ContextManager`, `GeminiService`

#### メソッド名
- パスカルケース（公開）: `BuildContext`, `GenerateContent`
- 小文字（非公開）: `buildContext`, `generateContent`

## 開発ワークフロー

### 1. ブランチ戦略

- **main**: 本番環境用ブランチ
- **develop**: 開発用ブランチ
- **feature/***: 新機能開発用ブランチ
- **bugfix/***: バグ修正用ブランチ
- **hotfix/***: 緊急修正用ブランチ

### 2. コミットメッセージ

[Conventional Commits](https://www.conventionalcommits.org/)の形式に従います。

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### タイプ

| タイプ | 説明 |
|-------|------|
| `feat` | 新機能 |
| `fix` | バグ修正 |
| `docs` | ドキュメント |
| `style` | コードスタイル |
| `refactor` | リファクタリング |
| `test` | テスト |
| `chore` | その他 |

#### 例

```
feat(mention): メンション機能の実装

- チャット履歴の取得機能を追加
- コンテキスト構築機能を実装
- Gemini APIとの連携機能を追加

Closes #123
```

### 3. プルリクエスト

#### 作成手順

1. 機能ブランチを作成
```bash
git checkout -b feature/new-feature
```

2. 変更をコミット
```bash
git add .
git commit -m "feat: 新機能の実装"
```

3. ブランチをプッシュ
```bash
git push origin feature/new-feature
```

4. プルリクエストを作成

#### プルリクエストテンプレート

```markdown
## 概要
<!-- 変更内容の概要を記述 -->

## 変更内容
<!-- 具体的な変更内容を記述 -->

## テスト
<!-- テスト内容を記述 -->

## チェックリスト
- [ ] コードレビューを依頼
- [ ] テストが通ることを確認
- [ ] ドキュメントを更新
- [ ] 破壊的変更がないことを確認
```

## テスト

### 1. テストの実行

```bash
# 全テストの実行
go test ./...

# 特定のパッケージのテスト
go test ./internal/domain

# カバレッジ付きテスト
go test -cover ./...

# ベンチマークテスト
go test -bench=. ./...
```

### 2. テストの書き方

#### 単体テスト

```go
func TestContextManager_BuildContext(t *testing.T) {
    // Arrange
    manager := NewContextManager()
    messages := []*discordgo.Message{
        {Content: "Hello", Author: &discordgo.User{Username: "user1"}},
    }
    
    // Act
    content, err := manager.BuildContext(messages, "What's up?")
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, content)
}
```

#### 統合テスト

```go
func TestMentionService_ProcessMention_Integration(t *testing.T) {
    // テスト用のDiscordセッションを作成
    session := createTestSession()
    defer session.Close()
    
    // テスト用のGeminiクライアントを作成
    client := createTestGeminiClient()
    
    // サービスの作成
    service := NewMentionService(session, client)
    
    // テスト実行
    err := service.ProcessMention(context.Background(), testMessage)
    
    // 検証
    assert.NoError(t, err)
}
```

### 3. モックの使用

```go
// モックの定義
type MockGeminiService struct {
    mock.Mock
}

func (m *MockGeminiService) GenerateContent(ctx context.Context, content *genai.Content) (string, error) {
    args := m.Called(ctx, content)
    return args.String(0), args.Error(1)
}

// テストでの使用
func TestMentionService_ProcessMention(t *testing.T) {
    // モックの作成
    mockGemini := new(MockGeminiService)
    mockGemini.On("GenerateContent", mock.Anything, mock.Anything).Return("Test response", nil)
    
    // サービスの作成
    service := NewMentionService(nil, mockGemini)
    
    // テスト実行
    err := service.ProcessMention(context.Background(), testMessage)
    
    // 検証
    assert.NoError(t, err)
    mockGemini.AssertExpectations(t)
}
```

## コードレビュー

### 1. レビューの観点

#### 機能性
- 要件を満たしているか
- エラーハンドリングが適切か
- エッジケースを考慮しているか

#### コード品質
- 可読性が高いか
- 命名が適切か
- 重複コードがないか

#### アーキテクチャ
- レイヤー分離が適切か
- 依存関係が正しいか
- テスタビリティが高いか

#### パフォーマンス
- 非同期処理が適切か
- メモリリークがないか
- 効率的なアルゴリズムを使用しているか

### 2. レビューのコメント

#### 良いコメント例

```go
// コンテキストが長すぎる場合は、新しいメッセージから優先的に保持する
func (m *ContextManager) TruncateContext(content string, maxLength int) string {
    if len(content) <= maxLength {
        return content
    }
    
    // 完全な文で終わるように調整
    truncated := content[:maxLength]
    lastSentenceEnd := strings.LastIndex(truncated, "。")
    if lastSentenceEnd > 0 {
        return truncated[:lastSentenceEnd+1]
    }
    
    return truncated
}
```

#### 改善が必要なコメント例

```go
// この関数は何かをする
func doSomething() {
    // 何かの処理
}
```

## デバッグ

### 1. ログの活用

```go
import "log"

func (s *MentionService) ProcessMention(ctx context.Context, message *discordgo.Message) error {
    log.Printf("メンション処理開始: チャンネルID=%s, ユーザーID=%s", 
        message.ChannelID, message.Author.ID)
    
    // 処理...
    
    log.Printf("メンション処理完了: チャンネルID=%s", message.ChannelID)
    return nil
}
```

### 2. デバッグツール

#### Delveデバッガー

```bash
# デバッグモードで起動
dlv debug cmd/main.go

# ブレークポイントの設定
(dlv) break internal/application/mention_service.go:25

# 実行
(dlv) continue

# 変数の確認
(dlv) print message.ChannelID
```

#### プロファイリング

```bash
# CPUプロファイリング
go run -cpuprofile=cpu.prof cmd/main.go

# メモリプロファイリング
go run -memprofile=mem.prof cmd/main.go

# プロファイルの確認
go tool pprof cpu.prof
```

## パフォーマンス最適化

### 1. 並行処理

```go
func (s *MentionService) ProcessMentions(ctx context.Context, messages []*discordgo.Message) error {
    // 並行処理でメンションを処理
    var wg sync.WaitGroup
    errChan := make(chan error, len(messages))
    
    for _, message := range messages {
        wg.Add(1)
        go func(msg *discordgo.Message) {
            defer wg.Done()
            if err := s.ProcessMention(ctx, msg); err != nil {
                errChan <- err
            }
        }(message)
    }
    
    wg.Wait()
    close(errChan)
    
    // エラーの確認
    for err := range errChan {
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

### 2. キャッシュの活用

```go
type CachedGeminiService struct {
    service GeminiService
    cache   map[string]string
    mutex   sync.RWMutex
}

func (s *CachedGeminiService) GenerateContent(ctx context.Context, content *genai.Content) (string, error) {
    // キャッシュキーの生成
    key := generateCacheKey(content)
    
    // キャッシュの確認
    s.mutex.RLock()
    if cached, exists := s.cache[key]; exists {
        s.mutex.RUnlock()
        return cached, nil
    }
    s.mutex.RUnlock()
    
    // キャッシュにない場合は生成
    result, err := s.service.GenerateContent(ctx, content)
    if err != nil {
        return "", err
    }
    
    // キャッシュに保存
    s.mutex.Lock()
    s.cache[key] = result
    s.mutex.Unlock()
    
    return result, nil
}
```

## セキュリティ

### 1. 入力検証

```go
func ValidateAPIKey(apiKey string) error {
    if apiKey == "" {
        return errors.New("APIキーが空です")
    }
    
    if len(apiKey) < 32 {
        return errors.New("APIキーが短すぎます")
    }
    
    // 形式の検証
    if !regexp.MustCompile(`^[A-Za-z0-9_-]+$`).MatchString(apiKey) {
        return errors.New("APIキーの形式が無効です")
    }
    
    return nil
}
```

### 2. 機密情報の保護

```go
func LogMessage(message string) {
    // 機密情報をマスク
    masked := regexp.MustCompile(`(?i)(api[_-]?key|token|password)\s*[:=]\s*[^\s]+`).
        ReplaceAllString(message, "$1: ***")
    
    log.Printf("メッセージ: %s", masked)
}
```

## トラブルシューティング

### 1. よくある問題

#### Discord Bot Tokenが無効

```
エラー: 401 Unauthorized
```

**解決方法**:
1. Discord Developer PortalでTokenを確認
2. Botの権限を確認
3. Tokenの再生成

#### Gemini API Keyが無効

```
エラー: 400 Bad Request - Invalid API key
```

**解決方法**:
1. Google AI StudioでAPI Keyを確認
2. API Keyの権限を確認
3. 新しいAPI Keyを生成

#### コンテキストが長すぎる

```
エラー: Context too long
```

**解決方法**:
1. `MAX_CONTEXT_LENGTH`の値を調整
2. コンテキストの切り詰め処理を確認

### 2. ログの確認

```bash
# ログファイルの確認
tail -f logs/geminibot.log

# エラーログの確認
grep "ERROR" logs/geminibot.log

# 特定のユーザーのログを確認
grep "user123" logs/geminibot.log
```

## コントリビューション

### 1. コントリビューションの流れ

1. イシューの作成または確認
2. 機能ブランチの作成
3. 実装とテスト
4. プルリクエストの作成
5. コードレビュー
6. マージ

### 2. コントリビューションガイドライン

- 既存のコードスタイルに従う
- 適切なテストを書く
- ドキュメントを更新する
- 破壊的変更を避ける
- セキュリティを考慮する

### 3. イシューの報告

#### バグレポート

```markdown
## バグの概要
<!-- バグの簡潔な説明 -->

## 再現手順
1. 
2. 
3. 

## 期待される動作
<!-- 期待される動作を記述 -->

## 実際の動作
<!-- 実際の動作を記述 -->

## 環境
- OS: 
- Go Version: 
- Bot Version: 

## ログ
<!-- 関連するログを貼り付け -->
```

#### 機能要求

```markdown
## 機能の概要
<!-- 要求する機能の簡潔な説明 -->

## 動機
<!-- なぜこの機能が必要なのか -->

## 詳細
<!-- 機能の詳細な説明 -->

## 代替案
<!-- 検討した代替案があれば -->
```

## 参考資料

- [Go公式ドキュメント](https://golang.org/doc/)
- [Discord API Documentation](https://discord.com/developers/docs)
- [Google Generative AI Documentation](https://ai.google.dev/docs)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
