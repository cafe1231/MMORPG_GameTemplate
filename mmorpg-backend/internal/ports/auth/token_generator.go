package auth

import (
	"context"

	"github.com/mmorpg-template/backend/internal/domain/auth"
)

// TokenGenerator defines the interface for token generation and validation
type TokenGenerator interface {
	// GenerateTokenPair generates an access and refresh token pair
	GenerateTokenPair(ctx context.Context, user *auth.User, sessionID, deviceID string) (*auth.TokenPair, error)
	
	// ValidateAccessToken validates an access token and returns the claims
	ValidateAccessToken(ctx context.Context, token string) (*auth.Claims, error)
	
	// ValidateRefreshToken validates a refresh token and returns the claims
	ValidateRefreshToken(ctx context.Context, token string) (*auth.RefreshClaims, error)
	
	// GeneratePasswordResetToken generates a password reset token
	GeneratePasswordResetToken(ctx context.Context, userID string) (string, error)
	
	// ValidatePasswordResetToken validates a password reset token
	ValidatePasswordResetToken(ctx context.Context, token string) (string, error)
	
	// HashToken creates a hash of a token for storage
	HashToken(token string) string
}

// PasswordHasher defines the interface for password hashing
type PasswordHasher interface {
	// HashPassword creates a hash from a password
	HashPassword(password string) (string, error)
	
	// ComparePassword compares a password with its hash
	ComparePassword(hash, password string) error
}