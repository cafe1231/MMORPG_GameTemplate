package auth

import (
	"context"

	"github.com/mmorpg-template/backend/internal/domain/auth"
)

// UserRepository defines the interface for user persistence
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *auth.User) error
	
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id string) (*auth.User, error)
	
	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*auth.User, error)
	
	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*auth.User, error)
	
	// Update updates a user
	Update(ctx context.Context, user *auth.User) error
	
	// Delete deletes a user
	Delete(ctx context.Context, id string) error
	
	// ExistsByEmail checks if a user exists with the given email
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	
	// ExistsByUsername checks if a user exists with the given username
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	
	// IncrementCharacterCount increments the character count for a user
	IncrementCharacterCount(ctx context.Context, userID string) error
	
	// DecrementCharacterCount decrements the character count for a user
	DecrementCharacterCount(ctx context.Context, userID string) error
}