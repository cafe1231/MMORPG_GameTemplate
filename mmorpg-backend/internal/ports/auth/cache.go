package auth

import (
	"context"
	"time"
)

// TokenCache defines the interface for caching tokens and related data
type TokenCache interface {
	// SetBlacklisted adds a token to the blacklist
	SetBlacklisted(ctx context.Context, tokenHash string, expiration time.Duration) error
	
	// IsBlacklisted checks if a token is blacklisted
	IsBlacklisted(ctx context.Context, tokenHash string) (bool, error)
	
	// SetSession caches a session
	SetSession(ctx context.Context, sessionID string, sessionData []byte, expiration time.Duration) error
	
	// GetSession retrieves a cached session
	GetSession(ctx context.Context, sessionID string) ([]byte, error)
	
	// DeleteSession removes a session from cache
	DeleteSession(ctx context.Context, sessionID string) error
	
	// SetLoginAttempts sets the login attempt count for an identifier (email/IP)
	SetLoginAttempts(ctx context.Context, identifier string, attempts int, expiration time.Duration) error
	
	// GetLoginAttempts gets the login attempt count for an identifier
	GetLoginAttempts(ctx context.Context, identifier string) (int, error)
	
	// IncrementLoginAttempts increments and returns the login attempt count
	IncrementLoginAttempts(ctx context.Context, identifier string, expiration time.Duration) (int, error)
	
	// DeleteLoginAttempts removes the login attempt count for an identifier
	DeleteLoginAttempts(ctx context.Context, identifier string) error
	
	// SetPasswordResetToken caches a password reset token
	SetPasswordResetToken(ctx context.Context, token string, userID string, expiration time.Duration) error
	
	// GetPasswordResetToken retrieves a password reset token's associated user ID
	GetPasswordResetToken(ctx context.Context, token string) (string, error)
	
	// DeletePasswordResetToken removes a password reset token
	DeletePasswordResetToken(ctx context.Context, token string) error
}