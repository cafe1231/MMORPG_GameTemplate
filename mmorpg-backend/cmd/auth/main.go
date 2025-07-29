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
	"github.com/mmorpg-template/backend/internal/adapters/auth"
	appAuth "github.com/mmorpg-template/backend/internal/application/auth"
	"github.com/mmorpg-template/backend/internal/config"
	portsAuth "github.com/mmorpg-template/backend/internal/ports/auth"
	"github.com/mmorpg-template/backend/pkg/db"
	"github.com/mmorpg-template/backend/pkg/logger"
	"github.com/mmorpg-template/backend/pkg/metrics"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	redisClient "github.com/redis/go-redis/v9"
)

func main() {
	// Initialize logger
	log := logger.New()
	log.Info("Starting MMORPG Auth Service")

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

	// Connect to NATS
	nc, err := nats.Connect(cfg.NATSURL())
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to NATS")
	}
	defer nc.Close()

	// Initialize repositories
	userRepo := auth.NewPostgresUserRepository(database)
	sessionRepo := auth.NewPostgresSessionRepository(database)

	// Initialize adapters
	tokenGenerator := auth.NewJWTGenerator(
		cfg.Auth.JWTAccessSecret,
		cfg.Auth.JWTRefreshSecret,
		"mmorpg-auth",
	)
	passwordHasher := auth.NewBcryptPasswordHasher(12)
	tokenCache := auth.NewRedisTokenCache(redisClient, "auth")

	// Initialize auth service
	authConfig := &appAuth.Config{
		MaxSessionsPerUser:   10,
		LoginRateLimit:       10,
		LoginRateLimitWindow: 15 * time.Minute,
		SessionDuration:      7 * 24 * time.Hour,
		MaxLoginAttempts:     5,
	}

	authService := appAuth.NewAuthService(
		userRepo,
		sessionRepo,
		tokenGenerator,
		passwordHasher,
		tokenCache,
		authConfig,
		log,
	)

	// Initialize HTTP handler
	httpHandler := auth.NewHTTPHandler(authService, log)

	// Setup HTTP server
	router := setupRouter(httpHandler)

	// Start HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Auth.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.WithField("port", cfg.Auth.Port).Info("Auth service listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Setup NATS subscriptions
	setupNATSSubscriptions(nc, authService, log)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down auth service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Failed to gracefully shutdown HTTP server")
	}

	log.Info("Auth service stopped")
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

func setupRouter(handler *auth.HTTPHandler) *gin.Engine {
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

	// Auth routes (public)
	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handler.Register)
			auth.POST("/login", handler.Login)
			auth.POST("/refresh", handler.RefreshToken)
			
			// Protected routes
			protected := auth.Group("")
			protected.Use(handler.Middleware())
			{
				protected.POST("/logout", handler.Logout)
				protected.GET("/verify", handler.VerifyToken)
			}
		}
	}

	return router
}

func setupNATSSubscriptions(nc *nats.Conn, authService portsAuth.AuthService, log logger.Logger) {
	// Subscribe to auth validation requests from other services
	nc.Subscribe("auth.validate", func(m *nats.Msg) {
		// Parse token from message
		token := string(m.Data)
		
		// Validate token
		claims, err := authService.ValidateToken(context.Background(), token)
		if err != nil {
			m.Respond([]byte(fmt.Sprintf(`{"valid":false,"error":"%s"}`, err.Error())))
			return
		}

		// Return claims
		response := fmt.Sprintf(`{"valid":true,"user_id":"%s","session_id":"%s","roles":%v}`,
			claims.UserID, claims.SessionID, claims.Roles)
		m.Respond([]byte(response))
	})

	// Subscribe to user info requests
	nc.Subscribe("auth.user.get", func(m *nats.Msg) {
		userID := string(m.Data)
		
		user, err := authService.GetUser(context.Background(), userID)
		if err != nil {
			m.Respond([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
			return
		}

		// Return user info (simplified for now)
		response := fmt.Sprintf(`{"id":"%s","email":"%s","username":"%s","roles":%v}`,
			user.ID, user.Email, user.Username, user.Roles)
		m.Respond([]byte(response))
	})

	log.Info("NATS subscriptions established")
}