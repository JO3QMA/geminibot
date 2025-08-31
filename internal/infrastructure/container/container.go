package container

import (
	"fmt"
	"log"

	"geminibot/configs"
	"geminibot/internal/application"
	"geminibot/internal/domain"
	"geminibot/internal/infrastructure/database"
	"geminibot/internal/infrastructure/gemini"
	discordPres "geminibot/internal/presentation/discord"

	"github.com/bwmarrin/discordgo"
)

// Container は依存性注入コンテナです
type Container struct {
	config           *configs.Config
	discordSession   *discordgo.Session
	geminiClient     application.GeminiClient
	conversationRepo domain.ConversationRepository
	mentionService   *application.MentionApplicationService
	discordHandler   *discordPres.DiscordHandlerNew
}

// NewContainer は新しいContainerインスタンスを作成します
func NewContainer(config *configs.Config) (*Container, error) {
	container := &Container{
		config: config,
	}

	// 依存関係を順次初期化
	if err := container.initializeDiscordSession(); err != nil {
		return nil, fmt.Errorf("Discordセッション初期化エラー: %w", err)
	}

	if err := container.initializeGeminiClient(); err != nil {
		return nil, fmt.Errorf("Geminiクライアント初期化エラー: %w", err)
	}

	if err := container.initializeDatabase(); err != nil {
		return nil, fmt.Errorf("データベース初期化エラー: %w", err)
	}

	if err := container.initializeServices(); err != nil {
		return nil, fmt.Errorf("サービス初期化エラー: %w", err)
	}

	if err := container.initializeHandlers(); err != nil {
		return nil, fmt.Errorf("ハンドラー初期化エラー: %w", err)
	}

	return container, nil
}

// initializeDiscordSession はDiscordセッションを初期化します
func (c *Container) initializeDiscordSession() error {
	session, err := discordgo.New("Bot " + c.config.Discord.BotToken)
	if err != nil {
		return fmt.Errorf("Discordセッション作成エラー: %w", err)
	}

	// Botの情報を取得
	user, err := session.User("@me")
	if err != nil {
		return fmt.Errorf("Bot情報取得エラー: %w", err)
	}

	log.Printf("Bot情報: %s#%s (ID: %s)", user.Username, user.Discriminator, user.ID)
	c.discordSession = session
	return nil
}

// initializeGeminiClient はGeminiクライアントを初期化します
func (c *Container) initializeGeminiClient() error {
	geminiClient, err := gemini.NewGeminiAPIClient(c.config.Gemini.APIKey, &c.config.Gemini)
	if err != nil {
		return fmt.Errorf("Gemini APIクライアント作成エラー: %w", err)
	}
	c.geminiClient = geminiClient
	return nil
}

// initializeDatabase はデータベース接続を初期化します
func (c *Container) initializeDatabase() error {
	// PostgreSQLリポジトリを作成
	postgresConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.config.Database.PostgresHost,
		c.config.Database.PostgresPort,
		c.config.Database.PostgresUser,
		c.config.Database.PostgresPassword,
		c.config.Database.PostgresDB,
	)

	postgresRepo, err := database.NewPostgresConversationRepository(postgresConnStr)
	if err != nil {
		return fmt.Errorf("PostgreSQLリポジトリ作成エラー: %w", err)
	}

	// Redisキャッシュを作成
	redisAddr := fmt.Sprintf("%s:%d", c.config.Database.RedisHost, c.config.Database.RedisPort)
	redisCache, err := database.NewRedisConversationCache(redisAddr, c.config.Database.RedisPassword, c.config.Database.RedisDB)
	if err != nil {
		return fmt.Errorf("Redisキャッシュ作成エラー: %w", err)
	}

	// ハイブリッドリポジトリを作成
	c.conversationRepo = database.NewHybridConversationRepository(postgresRepo, redisCache)
	return nil
}

// initializeServices はアプリケーションサービスを初期化します
func (c *Container) initializeServices() error {
	c.mentionService = application.NewMentionApplicationService(
		c.conversationRepo,
		c.geminiClient,
		&c.config.Bot,
	)
	return nil
}

// initializeHandlers はプレゼンテーション層のハンドラーを初期化します
func (c *Container) initializeHandlers() error {
	user, err := c.discordSession.User("@me")
	if err != nil {
		return fmt.Errorf("Bot情報取得エラー: %w", err)
	}

	c.discordHandler = discordPres.NewDiscordHandlerNew(c.discordSession, c.mentionService, user.ID)
	return nil
}

// GetDiscordSession はDiscordセッションを取得します
func (c *Container) GetDiscordSession() *discordgo.Session {
	return c.discordSession
}

// GetDiscordHandler はDiscordハンドラーを取得します
func (c *Container) GetDiscordHandler() *discordPres.DiscordHandlerNew {
	return c.discordHandler
}

// Close はすべてのリソースを適切にクローズします
func (c *Container) Close() error {
	var errors []error

	if c.discordSession != nil {
		if err := c.discordSession.Close(); err != nil {
			errors = append(errors, fmt.Errorf("Discordセッションクローズエラー: %w", err))
		}
	}

	// GeminiClientとConversationRepositoryはインターフェースなので、
	// 具体的な実装でCloseメソッドが実装されている場合のみクローズ
	if closer, ok := c.geminiClient.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("Geminiクライアントクローズエラー: %w", err))
		}
	}

	if closer, ok := c.conversationRepo.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("データベース接続クローズエラー: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("リソースクローズエラー: %v", errors)
	}

	return nil
}
