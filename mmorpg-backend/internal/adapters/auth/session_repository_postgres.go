package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/auth"
	portsAuth "github.com/mmorpg-template/backend/internal/ports/auth"
)

// PostgresSessionRepository implements SessionRepository using PostgreSQL
type PostgresSessionRepository struct {
	db *sql.DB
}

// NewPostgresSessionRepository creates a new PostgreSQL session repository
func NewPostgresSessionRepository(db *sql.DB) portsAuth.SessionRepository {
	return &PostgresSessionRepository{db: db}
}

// Create creates a new session
func (r *PostgresSessionRepository) Create(ctx context.Context, session *auth.Session) error {
	query := `
		INSERT INTO sessions (
			id, user_id, token_hash, device_id, ip_address,
			user_agent, expires_at, created_at, last_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.TokenHash,
		session.DeviceID,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
		session.CreatedAt,
		session.LastActive,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByID retrieves a session by ID
func (r *PostgresSessionRepository) GetByID(ctx context.Context, id string) (*auth.Session, error) {
	sessionID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	query := `
		SELECT 
			id, user_id, token_hash, device_id, ip_address,
			user_agent, expires_at, created_at, last_active
		FROM sessions
		WHERE id = $1
	`

	session := &auth.Session{}
	err = r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.DeviceID,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, auth.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by ID: %w", err)
	}

	if session.IsExpired() {
		return nil, auth.ErrSessionExpired
	}

	return session, nil
}

// GetByTokenHash retrieves a session by token hash
func (r *PostgresSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*auth.Session, error) {
	query := `
		SELECT 
			id, user_id, token_hash, device_id, ip_address,
			user_agent, expires_at, created_at, last_active
		FROM sessions
		WHERE token_hash = $1 AND expires_at > NOW()
	`

	session := &auth.Session{}
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.DeviceID,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, auth.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by token hash: %w", err)
	}

	return session, nil
}

// GetByUserID retrieves all sessions for a user
func (r *PostgresSessionRepository) GetByUserID(ctx context.Context, userID string) ([]*auth.Session, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	query := `
		SELECT 
			id, user_id, token_hash, device_id, ip_address,
			user_agent, expires_at, created_at, last_active
		FROM sessions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY last_active DESC
	`

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by user ID: %w", err)
	}
	defer rows.Close()

	var sessions []*auth.Session
	for rows.Next() {
		session := &auth.Session{}
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.TokenHash,
			&session.DeviceID,
			&session.IPAddress,
			&session.UserAgent,
			&session.ExpiresAt,
			&session.CreatedAt,
			&session.LastActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// Update updates a session
func (r *PostgresSessionRepository) Update(ctx context.Context, session *auth.Session) error {
	query := `
		UPDATE sessions SET
			token_hash = $2,
			device_id = $3,
			ip_address = $4,
			user_agent = $5,
			expires_at = $6,
			last_active = $7
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.TokenHash,
		session.DeviceID,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
		session.LastActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return auth.ErrSessionNotFound
	}

	return nil
}

// Delete deletes a session
func (r *PostgresSessionRepository) Delete(ctx context.Context, id string) error {
	sessionID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	query := `DELETE FROM sessions WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return auth.ErrSessionNotFound
	}

	return nil
}

// DeleteByUserID deletes all sessions for a user
func (r *PostgresSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err = r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete sessions by user ID: %w", err)
	}

	return nil
}

// DeleteExpired deletes all expired sessions
func (r *PostgresSessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}

// CountByUserID counts active sessions for a user
func (r *PostgresSessionRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID: %w", err)
	}

	query := `SELECT COUNT(*) FROM sessions WHERE user_id = $1 AND expires_at > NOW()`
	
	var count int
	err = r.db.QueryRowContext(ctx, query, id).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	return count, nil
}

// UpdateLastActive updates the last active timestamp for a session
func (r *PostgresSessionRepository) UpdateLastActive(ctx context.Context, sessionID string, lastActive time.Time) error {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	query := `UPDATE sessions SET last_active = $2 WHERE id = $1 AND expires_at > NOW()`
	
	result, err := r.db.ExecContext(ctx, query, id, lastActive)
	if err != nil {
		return fmt.Errorf("failed to update last active: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return auth.ErrSessionNotFound
	}

	return nil
}