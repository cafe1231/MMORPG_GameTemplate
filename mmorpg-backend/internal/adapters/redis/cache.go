package redis

import (
	"context"
	"fmt"
	"time"
	
	"github.com/redis/go-redis/v9"
	"github.com/mmorpg-template/backend/internal/ports"
	"github.com/mmorpg-template/backend/pkg/logger"
)

// RedisCache implements the ports.Cache interface for Redis
type RedisCache struct {
	client *redis.Client
	config *ports.CacheConfig
	log    logger.Logger
}

// NewRedisCache creates a new Redis cache adapter
func NewRedisCache(config *ports.CacheConfig, log logger.Logger) *RedisCache {
	return &RedisCache{
		config: config,
		log:    log,
	}
}

// Connect establishes a connection to Redis
func (r *RedisCache) Connect(ctx context.Context) error {
	opt, err := redis.ParseURL(r.config.URL)
	if err != nil {
		return fmt.Errorf("failed to parse redis URL: %w", err)
	}
	
	// Apply additional configuration
	opt.PoolSize = r.config.PoolSize
	opt.MinIdleConns = r.config.MinIdleConns
	opt.MaxRetries = r.config.MaxRetries
	opt.DialTimeout = r.config.DialTimeout
	opt.ReadTimeout = r.config.ReadTimeout
	opt.WriteTimeout = r.config.WriteTimeout
	
	if r.config.Password != "" {
		opt.Password = r.config.Password
	}
	if r.config.DB != 0 {
		opt.DB = r.config.DB
	}
	
	r.client = redis.NewClient(opt)
	
	// Test connection
	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}
	
	r.log.Info("Connected to Redis cache")
	return nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// Ping checks if the Redis connection is alive
func (r *RedisCache) Ping(ctx context.Context) error {
	if r.client == nil {
		return ports.ErrCacheConnect
	}
	return r.client.Ping(ctx).Err()
}

// Get retrieves a value from cache
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ports.ErrCacheMiss
	}
	return val, err
}

// Set stores a value in cache
func (r *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Delete removes a value from cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in cache
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

// Increment atomically increments a value
func (r *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return r.client.IncrBy(ctx, key, delta).Result()
}

// Decrement atomically decrements a value
func (r *RedisCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return r.client.DecrBy(ctx, key, delta).Result()
}

// Expire sets expiration on a key
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL gets the time to live for a key
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if ttl < 0 {
		return 0, ports.ErrCacheExpired
	}
	return ttl, nil
}

// GetMultiple retrieves multiple values
func (r *RedisCache) GetMultiple(ctx context.Context, keys []string) (map[string][]byte, error) {
	values, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	
	result := make(map[string][]byte)
	for i, val := range values {
		if val != nil {
			if str, ok := val.(string); ok {
				result[keys[i]] = []byte(str)
			}
		}
	}
	return result, nil
}

// SetMultiple stores multiple values
func (r *RedisCache) SetMultiple(ctx context.Context, items map[string][]byte, expiration time.Duration) error {
	pipe := r.client.Pipeline()
	for key, value := range items {
		pipe.Set(ctx, key, value, expiration)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// DeleteMultiple removes multiple values
func (r *RedisCache) DeleteMultiple(ctx context.Context, keys []string) error {
	return r.client.Del(ctx, keys...).Err()
}

// ListPush adds values to a list
func (r *RedisCache) ListPush(ctx context.Context, key string, values ...[]byte) error {
	interfaces := make([]interface{}, len(values))
	for i, v := range values {
		interfaces[i] = v
	}
	return r.client.LPush(ctx, key, interfaces...).Err()
}

// ListPop removes and returns a value from a list
func (r *RedisCache) ListPop(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.LPop(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ports.ErrCacheMiss
	}
	return val, err
}

// ListLength gets the length of a list
func (r *RedisCache) ListLength(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

// ListRange gets a range of values from a list
func (r *RedisCache) ListRange(ctx context.Context, key string, start, stop int64) ([][]byte, error) {
	values, err := r.client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	
	result := make([][]byte, len(values))
	for i, val := range values {
		result[i] = []byte(val)
	}
	return result, nil
}

// SetAdd adds members to a set
func (r *RedisCache) SetAdd(ctx context.Context, key string, members ...[]byte) error {
	interfaces := make([]interface{}, len(members))
	for i, m := range members {
		interfaces[i] = m
	}
	return r.client.SAdd(ctx, key, interfaces...).Err()
}

// SetRemove removes members from a set
func (r *RedisCache) SetRemove(ctx context.Context, key string, members ...[]byte) error {
	interfaces := make([]interface{}, len(members))
	for i, m := range members {
		interfaces[i] = m
	}
	return r.client.SRem(ctx, key, interfaces...).Err()
}

// SetMembers gets all members of a set
func (r *RedisCache) SetMembers(ctx context.Context, key string) ([][]byte, error) {
	values, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	
	result := make([][]byte, len(values))
	for i, val := range values {
		result[i] = []byte(val)
	}
	return result, nil
}

// SetIsMember checks if a member exists in a set
func (r *RedisCache) SetIsMember(ctx context.Context, key string, member []byte) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}

// HashSet sets a field in a hash
func (r *RedisCache) HashSet(ctx context.Context, key, field string, value []byte) error {
	return r.client.HSet(ctx, key, field, value).Err()
}

// HashGet gets a field from a hash
func (r *RedisCache) HashGet(ctx context.Context, key, field string) ([]byte, error) {
	val, err := r.client.HGet(ctx, key, field).Bytes()
	if err == redis.Nil {
		return nil, ports.ErrCacheMiss
	}
	return val, err
}

// HashGetAll gets all fields from a hash
func (r *RedisCache) HashGetAll(ctx context.Context, key string) (map[string][]byte, error) {
	values, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	
	result := make(map[string][]byte)
	for k, v := range values {
		result[k] = []byte(v)
	}
	return result, nil
}

// HashDelete removes fields from a hash
func (r *RedisCache) HashDelete(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// HashExists checks if a field exists in a hash
func (r *RedisCache) HashExists(ctx context.Context, key, field string) (bool, error) {
	return r.client.HExists(ctx, key, field).Result()
}

// Subscribe creates a subscription to channels
func (r *RedisCache) Subscribe(ctx context.Context, channels ...string) (ports.Subscription, error) {
	sub := r.client.Subscribe(ctx, channels...)
	return &redisSubscription{sub: sub}, nil
}

// Publish publishes a message to a channel
func (r *RedisCache) Publish(ctx context.Context, channel string, message []byte) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// Keys returns keys matching a pattern
func (r *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

// FlushDB removes all keys from the database
func (r *RedisCache) FlushDB(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// redisSubscription wraps Redis pub/sub
type redisSubscription struct {
	sub *redis.PubSub
}

func (s *redisSubscription) Receive(ctx context.Context) (ports.Message, error) {
	msg, err := s.sub.ReceiveMessage(ctx)
	if err != nil {
		return nil, err
	}
	return &redisMessage{msg: msg}, nil
}

func (s *redisSubscription) Subscribe(ctx context.Context, channels ...string) error {
	return s.sub.Subscribe(ctx, channels...)
}

func (s *redisSubscription) Unsubscribe(ctx context.Context, channels ...string) error {
	return s.sub.Unsubscribe(ctx, channels...)
}

func (s *redisSubscription) Close() error {
	return s.sub.Close()
}

// redisMessage wraps Redis message
type redisMessage struct {
	msg *redis.Message
}

func (m *redisMessage) Channel() string {
	return m.msg.Channel
}

func (m *redisMessage) Payload() []byte {
	return []byte(m.msg.Payload)
}