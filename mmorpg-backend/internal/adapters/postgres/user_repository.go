package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	
	"github.com/mmorpg-template/backend/internal/ports"
	"github.com/google/uuid"
)

// UserRepository implements ports.UserRepository for PostgreSQL
type UserRepository struct {
	db ports.Database
	tx ports.Transaction
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db ports.Database) *UserRepository {
	return &UserRepository{db: db}
}

// WithTx returns a new repository instance with a transaction
func (r *UserRepository) WithTx(tx ports.Transaction) ports.Repository {
	return &UserRepository{db: r.db, tx: tx}
}

// getDB returns the transaction if available, otherwise the database
func (r *UserRepository) getDB() interface {
	Query(ctx context.Context, query string, args ...interface{}) (ports.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) ports.Row
	Exec(ctx context.Context, query string, args ...interface{}) (ports.Result, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *ports.User) error {
	query := `
		INSERT INTO users (
			id, email, username, password_hash, account_status,
			email_verified, is_premium, premium_expires_at, max_characters,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)`
	
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	
	_, err := r.getDB().Exec(ctx, query,
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.AccountStatus,
		user.EmailVerified,
		user.IsPremium,
		user.PremiumExpiresAt,
		user.MaxCharacters,
		user.CreatedAt,
		user.UpdatedAt,
	)
	
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			if strings.Contains(err.Error(), "email") {
				return fmt.Errorf("email already exists: %w", ports.ErrDuplicate)
			}
			if strings.Contains(err.Error(), "username") {
				return fmt.Errorf("username already exists: %w", ports.ErrDuplicate)
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*ports.User, error) {
	query := `
		SELECT 
			id, email, username, password_hash, account_status,
			email_verified, is_premium, premium_expires_at, max_characters,
			created_at, updated_at, last_login_at, deleted_at
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL`
	
	user := &ports.User{}
	err := r.getDB().QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.AccountStatus,
		&user.EmailVerified,
		&user.IsPremium,
		&user.PremiumExpiresAt,
		&user.MaxCharacters,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
		&user.DeletedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ports.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	
	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*ports.User, error) {
	query := `
		SELECT 
			id, email, username, password_hash, account_status,
			email_verified, is_premium, premium_expires_at, max_characters,
			created_at, updated_at, last_login_at, deleted_at
		FROM users 
		WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL`
	
	user := &ports.User{}
	err := r.getDB().QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.AccountStatus,
		&user.EmailVerified,
		&user.IsPremium,
		&user.PremiumExpiresAt,
		&user.MaxCharacters,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
		&user.DeletedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ports.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	
	return user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*ports.User, error) {
	query := `
		SELECT 
			id, email, username, password_hash, account_status,
			email_verified, is_premium, premium_expires_at, max_characters,
			created_at, updated_at, last_login_at, deleted_at
		FROM users 
		WHERE LOWER(username) = LOWER($1) AND deleted_at IS NULL`
	
	user := &ports.User{}
	err := r.getDB().QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.AccountStatus,
		&user.EmailVerified,
		&user.IsPremium,
		&user.PremiumExpiresAt,
		&user.MaxCharacters,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
		&user.DeletedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ports.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	
	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *ports.User) error {
	query := `
		UPDATE users SET
			email = $2,
			username = $3,
			account_status = $4,
			email_verified = $5,
			is_premium = $6,
			premium_expires_at = $7,
			max_characters = $8,
			updated_at = $9
		WHERE id = $1 AND deleted_at IS NULL`
	
	user.UpdatedAt = time.Now()
	
	result, err := r.getDB().Exec(ctx, query,
		user.ID,
		user.Email,
		user.Username,
		user.AccountStatus,
		user.EmailVerified,
		user.IsPremium,
		user.PremiumExpiresAt,
		user.MaxCharacters,
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
		return ports.ErrNotFound
	}
	
	return nil
}

// Delete soft deletes a user
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE users SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	
	result, err := r.getDB().Exec(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return ports.ErrNotFound
	}
	
	return nil
}

// List lists users with filtering and pagination
func (r *UserRepository) List(ctx context.Context, filter ports.UserFilter, pagination ports.Pagination) ([]*ports.User, int64, error) {
	// Build WHERE clause
	conditions := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argCount := 0
	
	if filter.AccountStatus != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("account_status = $%d", argCount))
		args = append(args, *filter.AccountStatus)
	}
	
	if filter.IsPremium != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("is_premium = $%d", argCount))
		args = append(args, *filter.IsPremium)
	}
	
	if filter.EmailVerified != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("email_verified = $%d", argCount))
		args = append(args, *filter.EmailVerified)
	}
	
	if filter.CreatedAfter != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("created_at > $%d", argCount))
		args = append(args, *filter.CreatedAfter)
	}
	
	if filter.CreatedBefore != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("created_at < $%d", argCount))
		args = append(args, *filter.CreatedBefore)
	}
	
	whereClause := strings.Join(conditions, " AND ")
	
	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause)
	var total int64
	err := r.getDB().QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}
	
	// Build ORDER BY clause
	orderClause := "created_at DESC" // default
	if pagination.Sort != "" {
		order := "ASC"
		if pagination.Order == "DESC" {
			order = "DESC"
		}
		orderClause = fmt.Sprintf("%s %s", pagination.Sort, order)
	}
	
	// Query with pagination
	query := fmt.Sprintf(`
		SELECT 
			id, email, username, password_hash, account_status,
			email_verified, is_premium, premium_expires_at, max_characters,
			created_at, updated_at, last_login_at, deleted_at
		FROM users 
		WHERE %s 
		ORDER BY %s 
		LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, argCount+1, argCount+2)
	
	args = append(args, pagination.Limit, pagination.Offset)
	
	rows, err := r.getDB().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()
	
	users := []*ports.User{}
	for rows.Next() {
		user := &ports.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.PasswordHash,
			&user.AccountStatus,
			&user.EmailVerified,
			&user.IsPremium,
			&user.PremiumExpiresAt,
			&user.MaxCharacters,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastLoginAt,
			&user.DeletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}
	
	return users, total, nil
}

// Exists checks if a user exists with the given email or username
func (r *UserRepository) Exists(ctx context.Context, email, username string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users 
			WHERE (LOWER(email) = LOWER($1) OR LOWER(username) = LOWER($2)) 
			AND deleted_at IS NULL
		)`
	
	var exists bool
	err := r.getDB().QueryRow(ctx, query, email, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	
	return exists, nil
}

// UpdateLastLogin updates the user's last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string, timestamp time.Time) error {
	query := `UPDATE users SET last_login_at = $2 WHERE id = $1`
	
	_, err := r.getDB().Exec(ctx, query, id, timestamp)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	
	return nil
}

// UpdatePassword updates the user's password hash
func (r *UserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = $3 WHERE id = $1`
	
	_, err := r.getDB().Exec(ctx, query, id, passwordHash, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	
	return nil
}