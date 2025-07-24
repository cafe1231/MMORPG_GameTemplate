package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mmorpg-template/backend/internal/domain/auth"
	portsAuth "github.com/mmorpg-template/backend/internal/ports/auth"
	"github.com/redis/go-redis/v9"
)

// RedisTokenCache implements TokenCache using Redis
type RedisTokenCache struct {
	client *redis.Client
	prefix string
}

// NewRedisTokenCache creates a new Redis token cache
func NewRedisTokenCache(client *redis.Client, prefix string) portsAuth.TokenCache {
	return &RedisTokenCache{
		client: client,
		prefix: prefix,
	}
}

// SetBlacklisted adds a token to the blacklist
func (c *RedisTokenCache) SetBlacklisted(ctx context.Context, tokenHash string, expiration time.Duration) error {
	key := fmt.Sprintf("%s:blacklist:%s", c.prefix, tokenHash)
	err := c.client.Set(ctx, key, true, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (c *RedisTokenCache) IsBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	key := fmt.Sprintf("%s:blacklist:%s", c.prefix, tokenHash)
	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}
	return exists > 0, nil
}

// SetSession caches a session
func (c *RedisTokenCache) SetSession(ctx context.Context, sessionID string, sessionData []byte, expiration time.Duration) error {
	key := fmt.Sprintf("%s:session:%s", c.prefix, sessionID)
	err := c.client.Set(ctx, key, sessionData, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to cache session: %w", err)
	}
	return nil
}

// GetSession retrieves a cached session
func (c *RedisTokenCache) GetSession(ctx context.Context, sessionID string) ([]byte, error) {
	key := fmt.Sprintf("%s:session:%s", c.prefix, sessionID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, auth.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get cached session: %w", err)
	}
	return data, nil
}

// DeleteSession removes a session from cache
func (c *RedisTokenCache) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("%s:session:%s", c.prefix, sessionID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cached session: %w", err)
	}
	return nil
}

// SetLoginAttempts sets the login attempt count for an identifier
func (c *RedisTokenCache) SetLoginAttempts(ctx context.Context, identifier string, attempts int, expiration time.Duration) error {
	key := fmt.Sprintf("%s:login_attempts:%s", c.prefix, identifier)
	err := c.client.Set(ctx, key, attempts, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set login attempts: %w", err)
	}
	return nil
}

// GetLoginAttempts gets the login attempt count for an identifier
func (c *RedisTokenCache) GetLoginAttempts(ctx context.Context, identifier string) (int, error) {
	key := fmt.Sprintf("%s:login_attempts:%s", c.prefix, identifier)
	attempts, err := c.client.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get login attempts: %w", err)
	}
	return attempts, nil
}

// IncrementLoginAttempts increments and returns the login attempt count
func (c *RedisTokenCache) IncrementLoginAttempts(ctx context.Context, identifier string, expiration time.Duration) (int, error) {
	key := fmt.Sprintf("%s:login_attempts:%s", c.prefix, identifier)
	
	// Use a pipeline to increment and set expiration atomically
	pipe := c.client.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment login attempts: %w", err)
	}
	
	return int(incrCmd.Val()), nil
}

// DeleteLoginAttempts removes the login attempt count for an identifier
func (c *RedisTokenCache) DeleteLoginAttempts(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("%s:login_attempts:%s", c.prefix, identifier)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete login attempts: %w", err)
	}
	return nil
}

// SetPasswordResetToken caches a password reset token
func (c *RedisTokenCache) SetPasswordResetToken(ctx context.Context, token string, userID string, expiration time.Duration) error {
	key := fmt.Sprintf("%s:password_reset:%s", c.prefix, token)
	err := c.client.Set(ctx, key, userID, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to cache password reset token: %w", err)
	}
	return nil
}

// GetPasswordResetToken retrieves a password reset token's associated user ID
func (c *RedisTokenCache) GetPasswordResetToken(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("%s:password_reset:%s", c.prefix, token)
	userID, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", auth.ErrInvalidToken
		}
		return "", fmt.Errorf("failed to get password reset token: %w", err)
	}
	return userID, nil
}

// DeletePasswordResetToken removes a password reset token
func (c *RedisTokenCache) DeletePasswordResetToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("%s:password_reset:%s", c.prefix, token)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete password reset token: %w", err)
	}
	return nil
}

// CacheSession is a helper method to cache a session struct
func (c *RedisTokenCache) CacheSession(ctx context.Context, session *auth.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return nil // Don't cache expired sessions
	}
	
	return c.SetSession(ctx, session.ID.String(), data, ttl)
}

// GetCachedSession is a helper method to retrieve a cached session struct
func (c *RedisTokenCache) GetCachedSession(ctx context.Context, sessionID string) (*auth.Session, error) {
	data, err := c.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	
	var session auth.Session
	err = json.Unmarshal(data, &session)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}
	
	if session.IsExpired() {
		// Clean up expired session from cache
		_ = c.DeleteSession(ctx, sessionID)
		return nil, auth.ErrSessionExpired
	}
	
	return &session, nil
}