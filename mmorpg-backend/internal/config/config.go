package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	NATS      NATSConfig
	Security  SecurityConfig
	Game      GameConfig
	Metrics   MetricsConfig
	Auth      AuthConfig
	Character CharacterConfig
}

type ServerConfig struct {
	Port            string
	Host            string
	ReadTimeout     int
	WriteTimeout    int
	ShutdownTimeout int
}

type DatabaseConfig struct {
	URL            string
	MaxConnections int
	MaxIdleConns   int
	ConnMaxLifetime int
}

type RedisConfig struct {
	URL             string
	PoolSize        int
	MinIdleConns    int
	MaxRetries      int
	DB              int
}

type NATSConfig struct {
	URL            string
	ClusterID      string
	ClientID       string
	MaxReconnects  int
	ReconnectWait  int
}

type SecurityConfig struct {
	JWTSecret         string
	JWTExpiry         int
	RefreshExpiry     int
	BcryptCost        int
	RateLimitPerIP    int
	RateLimitPerUser  int
}

type GameConfig struct {
	MaxPlayersPerWorld int
	ViewDistance       float64
	TickRate           int
	MaxInventorySize   int
}

type MetricsConfig struct {
	Port     string
	Enabled  bool
	Endpoint string
}

type AuthConfig struct {
	Port              int
	JWTAccessSecret   string
	JWTRefreshSecret  string
	MaxSessionsPerUser int
	LoginRateLimit    int
	LoginRateLimitWindow int
	MaxLoginAttempts  int
}

type CharacterConfig struct {
	Port                   int
	MaxCharactersPerUser   int
	MaxCharacterNameLength int
	MinCharacterNameLength int
	DefaultStartingLevel   int
	DefaultStartingExp     int64
}


func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.readTimeout", 15)
	viper.SetDefault("server.writeTimeout", 15)
	viper.SetDefault("server.shutdownTimeout", 30)

	// Database defaults
	viper.SetDefault("database.url", "postgres://dev:dev@localhost:5432/mmorpg?sslmode=disable")
	viper.SetDefault("database.maxConnections", 50)
	viper.SetDefault("database.maxIdleConns", 10)
	viper.SetDefault("database.connMaxLifetime", 300)

	// Redis defaults
	viper.SetDefault("redis.url", "redis://localhost:6379")
	viper.SetDefault("redis.poolSize", 100)
	viper.SetDefault("redis.minIdleConns", 10)
	viper.SetDefault("redis.maxRetries", 3)
	viper.SetDefault("redis.db", 0)

	// NATS defaults
	viper.SetDefault("nats.url", "nats://localhost:4222")
	viper.SetDefault("nats.clusterID", "mmorpg-cluster")
	viper.SetDefault("nats.clientID", "mmorpg-gateway")
	viper.SetDefault("nats.maxReconnects", 60)
	viper.SetDefault("nats.reconnectWait", 2)

	// Security defaults
	viper.SetDefault("security.jwtSecret", "change-me-in-production")
	viper.SetDefault("security.jwtExpiry", 3600)
	viper.SetDefault("security.refreshExpiry", 86400)
	viper.SetDefault("security.bcryptCost", 10)
	viper.SetDefault("security.rateLimitPerIP", 100)
	viper.SetDefault("security.rateLimitPerUser", 1000)

	// Game defaults
	viper.SetDefault("game.maxPlayersPerWorld", 1000)
	viper.SetDefault("game.viewDistance", 100.0)
	viper.SetDefault("game.tickRate", 30)
	viper.SetDefault("game.maxInventorySize", 100)

	// Metrics defaults
	viper.SetDefault("metrics.port", "9090")
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.endpoint", "/metrics")

	// Auth defaults
	viper.SetDefault("auth.port", 8081)
	viper.SetDefault("auth.jwtAccessSecret", "change-me-access-secret")
	viper.SetDefault("auth.jwtRefreshSecret", "change-me-refresh-secret")
	viper.SetDefault("auth.maxSessionsPerUser", 10)
	viper.SetDefault("auth.loginRateLimit", 10)
	viper.SetDefault("auth.loginRateLimitWindow", 900) // 15 minutes
	viper.SetDefault("auth.maxLoginAttempts", 5)
	
	// Character defaults
	viper.SetDefault("character.port", 8082)
	viper.SetDefault("character.maxCharactersPerUser", 5)
	viper.SetDefault("character.maxCharacterNameLength", 30)
	viper.SetDefault("character.minCharacterNameLength", 3)
	viper.SetDefault("character.defaultStartingLevel", 1)
	viper.SetDefault("character.defaultStartingExp", 0)
}

func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if c.Database.URL == "" {
		return fmt.Errorf("database URL is required")
	}

	if c.Redis.URL == "" {
		return fmt.Errorf("redis URL is required")
	}

	if c.NATS.URL == "" {
		return fmt.Errorf("NATS URL is required")
	}

	if c.Security.JWTSecret == "change-me-in-production" {
		fmt.Println("WARNING: Using default JWT secret. Please change in production!")
	}

	if c.Game.TickRate < 10 || c.Game.TickRate > 60 {
		return fmt.Errorf("game tick rate must be between 10 and 60")
	}

	return nil
}

func GetEnvOrDefault(key, defaultValue string) string {
	if value := viper.GetString(key); value != "" {
		return value
	}
	return defaultValue
}

// Convenience getters for common values
func (c *Config) DatabaseURL() string {
	return c.Database.URL
}

func (c *Config) RedisURL() string {
	return c.Redis.URL
}

func (c *Config) NATSURL() string {
	return c.NATS.URL
}

// Load is a convenience function that loads config and panics on error
// For services that need to fail fast on config errors
func Load() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}
	return cfg
}

// LoadConfig loads the configuration and returns error
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/mmorpg/")

	viper.SetEnvPrefix("MMORPG")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}