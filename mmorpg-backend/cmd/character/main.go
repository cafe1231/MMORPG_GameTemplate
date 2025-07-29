package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/mmorpg-template/backend/internal/adapters/character"
	redisCharacter "github.com/mmorpg-template/backend/internal/adapters/character/redis"
	natsCharacter "github.com/mmorpg-template/backend/internal/adapters/character/nats"
	natsAdapter "github.com/mmorpg-template/backend/internal/adapters/nats"
	appCharacter "github.com/mmorpg-template/backend/internal/application/character"
	"github.com/mmorpg-template/backend/internal/ports"
	portsCharacter "github.com/mmorpg-template/backend/internal/ports/character"
	"github.com/mmorpg-template/backend/internal/config"
	"github.com/mmorpg-template/backend/pkg/db"
	"github.com/mmorpg-template/backend/pkg/logger"
	"github.com/mmorpg-template/backend/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	redisClient "github.com/redis/go-redis/v9"
)

func main() {
	// Initialize logger
	log := logger.New()
	log.Info("Starting MMORPG Character Service")

	// Load configuration
	cfg := config.Load()
	if cfg == nil {
		log.Fatal("Failed to load configuration")
	}

	// Initialize metrics
	metrics.Init()

	// Connect to PostgreSQL
	database, err := initDatabase(cfg.DatabaseURL())
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer database.Close()

	// Run migrations
	migrator, err := db.NewMigrator(database, log)
	if err != nil {
		log.WithError(err).Fatal("Failed to create migrator")
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil {
		log.WithError(err).Fatal("Failed to run migrations")
	}

	// Connect to Redis
	redisClient, err := initRedis(cfg.RedisURL())
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Initialize NATS message queue
	mqConfig := &ports.MessageQueueConfig{
		URL:           cfg.NATSURL(),
		ClientID:      "character-service",
		MaxReconnects: 10,
		ReconnectWait: 2 * time.Second,
		PingInterval:  30 * time.Second,
		MaxPingsOut:   5,
	}
	
	mq := natsAdapter.NewNATSMessageQueue(mqConfig, log)
	if err := mq.Connect(context.Background()); err != nil {
		log.WithError(err).Fatal("Failed to connect to NATS")
	}
	defer mq.Close()
	
	// Initialize event publisher
	eventPublisher := natsCharacter.NewEventPublisher(mq, log)
	if err := eventPublisher.Initialize(context.Background()); err != nil {
		log.WithError(err).Fatal("Failed to initialize event publisher")
	}

	// Initialize repositories
	characterRepo := character.NewPostgresCharacterRepository(database)
	appearanceRepo := character.NewPostgresAppearanceRepository(database)
	statsRepo := character.NewPostgresStatsRepository(database)
	positionRepo := character.NewPostgresPositionRepository(database)

	// Initialize cache
	characterCache := redisCharacter.NewRedisCharacterCache(
		redisClient,
		"character",
		nil, // Use default TTL configuration
	)

	// Initialize character service
	characterConfig := &appCharacter.Config{
		MaxCharactersPerUser: 5,
		MaxCharacterNameLength: 30,
		MinCharacterNameLength: 3,
		DefaultStartingLevel: 1,
		DefaultStartingExperience: 0,
	}

	characterService := appCharacter.NewCharacterService(
		characterRepo,
		appearanceRepo,
		statsRepo,
		positionRepo,
		characterCache,
		eventPublisher,
		characterConfig,
		log,
	)

	// Initialize JWT middleware
	jwtConfig := &character.JWTConfig{
		AccessSecret: cfg.Auth.JWTAccessSecret,
		Issuer:       "mmorpg-auth",
	}
	jwtMiddleware := character.NewJWTMiddleware(jwtConfig, log)

	// Initialize HTTP handler
	httpHandler := character.NewHTTPHandler(characterService, jwtMiddleware, log)

	// Setup HTTP server
	router := setupRouter(httpHandler)

	// Start HTTP server
	// Character service runs on port 8082 by default
	port := 8082
	if cfg.Server.Port != "" {
		port = 8082 // Override with character-specific port if needed
	}
	
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.WithField("port", port).Info("Character service listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Setup NATS subscriptions
	setupNATSSubscriptions(mq, characterService, log)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down character service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Failed to gracefully shutdown HTTP server")
	}

	log.Info("Character service stopped")
}

func initDatabase(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func initRedis(redisURL string) (*redisClient.Client, error) {
	opts, err := redisClient.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redisClient.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}

func setupRouter(handler *character.HTTPHandler) *gin.Engine {
	// Set gin mode based on environment
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	
	// Middleware
	router.Use(gin.Recovery())
	router.Use(logger.GinLogger())
	router.Use(metrics.GinMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Register character routes
	handler.RegisterRoutes(router)

	return router
}

func setupNATSSubscriptions(mq ports.MessageQueue, characterService portsCharacter.CharacterService, log logger.Logger) {
	// Subscribe to character validation requests
	mq.Subscribe(context.Background(), "character.validate", func(msg *ports.QueueMessage) error {
		// Parse character ID from message
		characterID := string(msg.Data)
		
		// Validate character exists
		character, err := characterService.GetCharacter(context.Background(), characterID)
		if err != nil {
			// Would need to implement reply mechanism
			return fmt.Errorf("character not found: %w", err)
		}

		// Return character info
		response := fmt.Sprintf(`{"valid":true,"character_id":"%s","name":"%s","level":%d,"class":"%s"}`,
			character.ID, character.Name, character.Level, character.ClassType)
		// Would need to implement reply mechanism
		log.Infof("Validated character: %s", response)
		return nil
	})

	// Subscribe to character list requests for a user
	mq.Subscribe(context.Background(), "character.list.byuser", func(msg *ports.QueueMessage) error {
		userID := string(msg.Data)
		
		characters, err := characterService.ListCharactersByUser(context.Background(), userID)
		if err != nil {
			return fmt.Errorf("failed to list characters: %w", err)
		}

		// Simple JSON response for now
		response := fmt.Sprintf(`{"count":%d}`, len(characters))
		log.Infof("Listed characters for user %s: %s", userID, response)
		return nil
	})

	log.Info("NATS subscriptions established for character service")
}