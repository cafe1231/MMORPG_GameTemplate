package ports

import (
	"context"
	"time"
)

// Cache is the interface for cache operations
// This abstraction allows switching between different cache implementations (Redis, Memcached, etc.)
type Cache interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error
	
	// Basic operations
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	
	// Atomic operations
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	Decrement(ctx context.Context, key string, delta int64) (int64, error)
	
	// Expiration management
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	
	// Bulk operations
	GetMultiple(ctx context.Context, keys []string) (map[string][]byte, error)
	SetMultiple(ctx context.Context, items map[string][]byte, expiration time.Duration) error
	DeleteMultiple(ctx context.Context, keys []string) error
	
	// List operations
	ListPush(ctx context.Context, key string, values ...[]byte) error
	ListPop(ctx context.Context, key string) ([]byte, error)
	ListLength(ctx context.Context, key string) (int64, error)
	ListRange(ctx context.Context, key string, start, stop int64) ([][]byte, error)
	
	// Set operations
	SetAdd(ctx context.Context, key string, members ...[]byte) error
	SetRemove(ctx context.Context, key string, members ...[]byte) error
	SetMembers(ctx context.Context, key string) ([][]byte, error)
	SetIsMember(ctx context.Context, key string, member []byte) (bool, error)
	
	// Hash operations
	HashSet(ctx context.Context, key, field string, value []byte) error
	HashGet(ctx context.Context, key, field string) ([]byte, error)
	HashGetAll(ctx context.Context, key string) (map[string][]byte, error)
	HashDelete(ctx context.Context, key string, fields ...string) error
	HashExists(ctx context.Context, key, field string) (bool, error)
	
	// Pub/Sub operations
	Subscribe(ctx context.Context, channels ...string) (Subscription, error)
	Publish(ctx context.Context, channel string, message []byte) error
	
	// Pattern operations
	Keys(ctx context.Context, pattern string) ([]string, error)
	
	// Cache clearing
	FlushDB(ctx context.Context) error
}

// Subscription represents a pub/sub subscription
type Subscription interface {
	// Receive messages
	Receive(ctx context.Context) (Message, error)
	
	// Channel management
	Subscribe(ctx context.Context, channels ...string) error
	Unsubscribe(ctx context.Context, channels ...string) error
	
	// Close subscription
	Close() error
}

// Message represents a pub/sub message
type Message interface {
	Channel() string
	Payload() []byte
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	URL          string
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	Password     string
	DB           int
	
	// Timeouts
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	
	// Default expiration
	DefaultExpiration time.Duration
}

// Common cache errors
var (
	ErrCacheMiss      = NewError("CACHE_MISS", "Key not found in cache")
	ErrCacheSet       = NewError("CACHE_SET", "Failed to set cache value")
	ErrCacheDelete    = NewError("CACHE_DELETE", "Failed to delete cache key")
	ErrCacheExpired   = NewError("CACHE_EXPIRED", "Cache key has expired")
	ErrCacheConnect   = NewError("CACHE_CONNECT", "Failed to connect to cache")
)

// CacheSerializer defines how to serialize/deserialize cache values
type CacheSerializer interface {
	Serialize(v interface{}) ([]byte, error)
	Deserialize(data []byte, v interface{}) error
}

// CacheManager provides high-level cache operations with serialization
type CacheManager interface {
	// Typed operations
	GetObject(ctx context.Context, key string, obj interface{}) error
	SetObject(ctx context.Context, key string, obj interface{}, expiration time.Duration) error
	
	// Cache invalidation patterns
	InvalidatePattern(ctx context.Context, pattern string) error
	InvalidateTags(ctx context.Context, tags []string) error
	
	// Cache warming
	WarmCache(ctx context.Context, keys []string, loader func(string) (interface{}, error)) error
}