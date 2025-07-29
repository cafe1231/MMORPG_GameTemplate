package character

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/character"
	_ "github.com/lib/pq"
)

// PostgresCharacterRepository implements the CharacterRepository interface using PostgreSQL
type PostgresCharacterRepository struct {
	db DBExecutor
}

// NewPostgresCharacterRepository creates a new PostgreSQL character repository
func NewPostgresCharacterRepository(db *sql.DB) *PostgresCharacterRepository {
	return &PostgresCharacterRepository{
		db: db,
	}
}

// NewTransactionalCharacterRepository creates a new PostgreSQL character repository using a transaction
func NewTransactionalCharacterRepository(tx *sql.Tx) *PostgresCharacterRepository {
	return &PostgresCharacterRepository{
		db: tx,
	}
}

// Create creates a new character in the database
func (r *PostgresCharacterRepository) Create(ctx context.Context, char *character.Character) error {
	query := `
		INSERT INTO characters (
			id, user_id, name, slot_number, level, experience,
			class_type, race, gender, is_deleted, deleted_at,
			deletion_scheduled_at, created_at, updated_at,
			last_played_at, total_play_time
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16
		)`

	_, err := r.db.ExecContext(ctx, query,
		char.ID,
		char.UserID,
		char.Name,
		char.SlotNumber,
		char.Level,
		char.Experience,
		char.ClassType,
		char.Race,
		char.Gender,
		char.IsDeleted,
		char.DeletedAt,
		char.DeletionScheduledAt,
		char.CreatedAt,
		char.UpdatedAt,
		char.LastPlayedAt,
		char.TotalPlayTime,
	)

	if err != nil {
		return fmt.Errorf("failed to create character: %w", err)
	}

	return nil
}

// GetByID retrieves a character by ID
func (r *PostgresCharacterRepository) GetByID(ctx context.Context, id uuid.UUID) (*character.Character, error) {
	query := `
		SELECT 
			id, user_id, name, slot_number, level, experience,
			class_type, race, gender, is_deleted, deleted_at,
			deletion_scheduled_at, created_at, updated_at,
			last_played_at, total_play_time
		FROM characters
		WHERE id = $1`

	var char character.Character
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&char.ID,
		&char.UserID,
		&char.Name,
		&char.SlotNumber,
		&char.Level,
		&char.Experience,
		&char.ClassType,
		&char.Race,
		&char.Gender,
		&char.IsDeleted,
		&char.DeletedAt,
		&char.DeletionScheduledAt,
		&char.CreatedAt,
		&char.UpdatedAt,
		&char.LastPlayedAt,
		&char.TotalPlayTime,
	)

	if err == sql.ErrNoRows {
		return nil, character.ErrCharacterNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	return &char, nil
}

// GetByName retrieves a character by name
func (r *PostgresCharacterRepository) GetByName(ctx context.Context, name string) (*character.Character, error) {
	query := `
		SELECT 
			id, user_id, name, slot_number, level, experience,
			class_type, race, gender, is_deleted, deleted_at,
			deletion_scheduled_at, created_at, updated_at,
			last_played_at, total_play_time
		FROM characters
		WHERE LOWER(name) = LOWER($1)`

	var char character.Character
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&char.ID,
		&char.UserID,
		&char.Name,
		&char.SlotNumber,
		&char.Level,
		&char.Experience,
		&char.ClassType,
		&char.Race,
		&char.Gender,
		&char.IsDeleted,
		&char.DeletedAt,
		&char.DeletionScheduledAt,
		&char.CreatedAt,
		&char.UpdatedAt,
		&char.LastPlayedAt,
		&char.TotalPlayTime,
	)

	if err == sql.ErrNoRows {
		return nil, character.ErrCharacterNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get character by name: %w", err)
	}

	return &char, nil
}

// GetByUserID retrieves all characters for a user
func (r *PostgresCharacterRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*character.Character, error) {
	query := `
		SELECT 
			id, user_id, name, slot_number, level, experience,
			class_type, race, gender, is_deleted, deleted_at,
			deletion_scheduled_at, created_at, updated_at,
			last_played_at, total_play_time
		FROM characters
		WHERE user_id = $1
		ORDER BY slot_number`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query characters: %w", err)
	}
	defer rows.Close()

	var characters []*character.Character
	for rows.Next() {
		var char character.Character
		err := rows.Scan(
			&char.ID,
			&char.UserID,
			&char.Name,
			&char.SlotNumber,
			&char.Level,
			&char.Experience,
			&char.ClassType,
			&char.Race,
			&char.Gender,
			&char.IsDeleted,
			&char.DeletedAt,
			&char.DeletionScheduledAt,
			&char.CreatedAt,
			&char.UpdatedAt,
			&char.LastPlayedAt,
			&char.TotalPlayTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan character: %w", err)
		}
		characters = append(characters, &char)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating characters: %w", err)
	}

	return characters, nil
}

// GetByUserIDAndSlot retrieves a character by user ID and slot number
func (r *PostgresCharacterRepository) GetByUserIDAndSlot(ctx context.Context, userID uuid.UUID, slot int) (*character.Character, error) {
	query := `
		SELECT 
			id, user_id, name, slot_number, level, experience,
			class_type, race, gender, is_deleted, deleted_at,
			deletion_scheduled_at, created_at, updated_at,
			last_played_at, total_play_time
		FROM characters
		WHERE user_id = $1 AND slot_number = $2`

	var char character.Character
	err := r.db.QueryRowContext(ctx, query, userID, slot).Scan(
		&char.ID,
		&char.UserID,
		&char.Name,
		&char.SlotNumber,
		&char.Level,
		&char.Experience,
		&char.ClassType,
		&char.Race,
		&char.Gender,
		&char.IsDeleted,
		&char.DeletedAt,
		&char.DeletionScheduledAt,
		&char.CreatedAt,
		&char.UpdatedAt,
		&char.LastPlayedAt,
		&char.TotalPlayTime,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Slot is available
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get character by slot: %w", err)
	}

	return &char, nil
}

// Update updates a character in the database
func (r *PostgresCharacterRepository) Update(ctx context.Context, char *character.Character) error {
	query := `
		UPDATE characters SET
			name = $2,
			slot_number = $3,
			level = $4,
			experience = $5,
			class_type = $6,
			race = $7,
			gender = $8,
			is_deleted = $9,
			deleted_at = $10,
			deletion_scheduled_at = $11,
			updated_at = $12,
			last_played_at = $13,
			total_play_time = $14
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		char.ID,
		char.Name,
		char.SlotNumber,
		char.Level,
		char.Experience,
		char.ClassType,
		char.Race,
		char.Gender,
		char.IsDeleted,
		char.DeletedAt,
		char.DeletionScheduledAt,
		char.UpdatedAt,
		char.LastPlayedAt,
		char.TotalPlayTime,
	)

	if err != nil {
		return fmt.Errorf("failed to update character: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return character.ErrCharacterNotFound
	}

	return nil
}

// Delete permanently deletes a character
func (r *PostgresCharacterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM characters WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return character.ErrCharacterNotFound
	}

	return nil
}

// SoftDelete soft deletes a character
func (r *PostgresCharacterRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `SELECT soft_delete_character($1)`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete character: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted character
func (r *PostgresCharacterRepository) Restore(ctx context.Context, id uuid.UUID) error {
	query := `SELECT restore_character($1)`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore character: %w", err)
	}

	return nil
}

// CleanupDeleted removes characters past their recovery period
func (r *PostgresCharacterRepository) CleanupDeleted(ctx context.Context) error {
	query := `SELECT cleanup_deleted_characters()`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup deleted characters: %w", err)
	}

	return nil
}

// NameExists checks if a character name already exists
func (r *PostgresCharacterRepository) NameExists(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM characters WHERE LOWER(name) = LOWER($1) AND is_deleted = false)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check name existence: %w", err)
	}

	return exists, nil
}

// CountByUserID counts the number of active characters for a user
func (r *PostgresCharacterRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM characters WHERE user_id = $1 AND is_deleted = false`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count characters: %w", err)
	}

	return count, nil
}