package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/mmorpg-template/backend/internal/domain/auth"
	portsAuth "github.com/mmorpg-template/backend/internal/ports/auth"
)

// PostgresUserRepository implements UserRepository using PostgreSQL
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *sql.DB) portsAuth.UserRepository {
	return &PostgresUserRepository{db: db}
}

// Create creates a new user
func (r *PostgresUserRepository) Create(ctx context.Context, user *auth.User) error {
	query := `
		INSERT INTO users (
			id, email, username, password_hash, email_verified, 
			account_status, roles, max_characters, character_count,
			is_premium, premium_expires_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.EmailVerified,
		user.AccountStatus,
		pq.Array(user.Roles),
		user.MaxCharacters,
		user.CharacterCount,
		user.IsPremium,
		user.PremiumExpiresAt,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				if pqErr.Constraint == "users_email_key" {
					return auth.ErrEmailAlreadyTaken
				}
				if pqErr.Constraint == "users_username_key" {
					return auth.ErrUsernameAlreadyTaken
				}
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*auth.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	query := `
		SELECT 
			id, email, username, password_hash, email_verified,
			account_status, roles, max_characters, character_count,
			is_premium, premium_expires_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &auth.User{}
	err = r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.AccountStatus,
		pq.Array(&user.Roles),
		&user.MaxCharacters,
		&user.CharacterCount,
		&user.IsPremium,
		&user.PremiumExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	query := `
		SELECT 
			id, email, username, password_hash, email_verified,
			account_status, roles, max_characters, character_count,
			is_premium, premium_expires_at, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`

	user := &auth.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.AccountStatus,
		pq.Array(&user.Roles),
		&user.MaxCharacters,
		&user.CharacterCount,
		&user.IsPremium,
		&user.PremiumExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*auth.User, error) {
	query := `
		SELECT 
			id, email, username, password_hash, email_verified,
			account_status, roles, max_characters, character_count,
			is_premium, premium_expires_at, created_at, updated_at
		FROM users
		WHERE LOWER(username) = LOWER($1)
	`

	user := &auth.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.AccountStatus,
		pq.Array(&user.Roles),
		&user.MaxCharacters,
		&user.CharacterCount,
		&user.IsPremium,
		&user.PremiumExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

// Update updates a user
func (r *PostgresUserRepository) Update(ctx context.Context, user *auth.User) error {
	query := `
		UPDATE users SET
			email = $2,
			username = $3,
			password_hash = $4,
			email_verified = $5,
			account_status = $6,
			roles = $7,
			max_characters = $8,
			character_count = $9,
			is_premium = $10,
			premium_expires_at = $11,
			updated_at = $12
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.EmailVerified,
		user.AccountStatus,
		pq.Array(user.Roles),
		user.MaxCharacters,
		user.CharacterCount,
		user.IsPremium,
		user.PremiumExpiresAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return auth.ErrUserNotFound
	}

	return nil
}

// Delete deletes a user
func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return auth.ErrUserNotFound
	}

	return nil
}

// ExistsByEmail checks if a user exists with the given email
func (r *PostgresUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email) = LOWER($1))`
	
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists by email: %w", err)
	}

	return exists, nil
}

// ExistsByUsername checks if a user exists with the given username
func (r *PostgresUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(username) = LOWER($1))`
	
	var exists bool
	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists by username: %w", err)
	}

	return exists, nil
}

// IncrementCharacterCount increments the character count for a user
func (r *PostgresUserRepository) IncrementCharacterCount(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	query := `
		UPDATE users 
		SET character_count = character_count + 1, updated_at = $2
		WHERE id = $1 AND character_count < max_characters
	`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to increment character count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("character limit reached or user not found")
	}

	return nil
}

// DecrementCharacterCount decrements the character count for a user
func (r *PostgresUserRepository) DecrementCharacterCount(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	query := `
		UPDATE users 
		SET character_count = GREATEST(character_count - 1, 0), updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to decrement character count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return auth.ErrUserNotFound
	}

	return nil
}