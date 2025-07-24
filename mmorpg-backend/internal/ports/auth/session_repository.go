package auth

import (
	"context"
	"time"

	"github.com/mmorpg-template/backend/internal/domain/auth"
)

// SessionRepository defines the interface for session persistence
type SessionRepository interface {
	// Create creates a new session
	Create(ctx context.Context, session *auth.Session) error
	
	// GetByID retrieves a session by ID
	GetByID(ctx context.Context, id string) (*auth.Session, error)
	
	// GetByTokenHash retrieves a session by token hash
	GetByTokenHash(ctx context.Context, tokenHash string) (*auth.Session, error)
	
	// GetByUserID retrieves all sessions for a user
	GetByUserID(ctx context.Context, userID string) ([]*auth.Session, error)
	
	// Update updates a session
	Update(ctx context.Context, session *auth.Session) error
	
	// Delete deletes a session
	Delete(ctx context.Context, id string) error
	
	// DeleteByUserID deletes all sessions for a user
	DeleteByUserID(ctx context.Context, userID string) error
	
	// DeleteExpired deletes all expired sessions
	DeleteExpired(ctx context.Context) error
	
	// CountByUserID counts active sessions for a user
	CountByUserID(ctx context.Context, userID string) (int, error)
	
	// UpdateLastActive updates the last active timestamp for a session
	UpdateLastActive(ctx context.Context, sessionID string, lastActive time.Time) error
}