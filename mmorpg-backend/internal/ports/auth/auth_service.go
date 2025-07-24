package auth

import (
	"context"

	"github.com/mmorpg-template/backend/internal/domain/auth"
)

// AuthService defines the interface for authentication operations
// This is the port that will be used by adapters (HTTP handlers, gRPC, etc.)
type AuthService interface {
	// Register creates a new user account
	Register(ctx context.Context, req *auth.RegisterRequest) (*auth.User, error)
	
	// Login authenticates a user and returns tokens
	Login(ctx context.Context, email, password, deviceID, ipAddress, userAgent string) (*auth.TokenPair, *auth.User, error)
	
	// Logout invalidates a session
	Logout(ctx context.Context, sessionID string) error
	
	// LogoutAllDevices logs out all sessions for a user
	LogoutAllDevices(ctx context.Context, userID string) error
	
	// RefreshToken generates a new token pair from a refresh token
	RefreshToken(ctx context.Context, refreshToken, deviceID, ipAddress, userAgent string) (*auth.TokenPair, error)
	
	// ValidateToken validates an access token
	ValidateToken(ctx context.Context, token string) (*auth.Claims, error)
	
	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, userID string) (*auth.User, error)
	
	// ChangePassword changes a user's password
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	
	// RequestPasswordReset initiates a password reset
	RequestPasswordReset(ctx context.Context, email string) error
	
	// ResetPassword completes a password reset
	ResetPassword(ctx context.Context, token, newPassword string) error
	
	// GetUserSessions retrieves all active sessions for a user
	GetUserSessions(ctx context.Context, userID string) ([]*auth.Session, error)
	
	// RevokeSession revokes a specific session
	RevokeSession(ctx context.Context, sessionID string) error
}