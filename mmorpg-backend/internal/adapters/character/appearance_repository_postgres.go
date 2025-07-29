package character

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/character"
	"github.com/lib/pq"
)

// PostgresAppearanceRepository implements the AppearanceRepository interface using PostgreSQL
type PostgresAppearanceRepository struct {
	db DBExecutor
}

// NewPostgresAppearanceRepository creates a new PostgreSQL appearance repository
func NewPostgresAppearanceRepository(db *sql.DB) *PostgresAppearanceRepository {
	return &PostgresAppearanceRepository{
		db: db,
	}
}

// NewTransactionalAppearanceRepository creates a new PostgreSQL appearance repository using a transaction
func NewTransactionalAppearanceRepository(tx *sql.Tx) *PostgresAppearanceRepository {
	return &PostgresAppearanceRepository{
		db: tx,
	}
}

// Create creates a new character appearance in the database
func (r *PostgresAppearanceRepository) Create(ctx context.Context, appearance *character.Appearance) error {
	// Convert body proportions to JSON
	bodyPropsJSON, err := json.Marshal(appearance.BodyProportions)
	if err != nil {
		return fmt.Errorf("failed to marshal body proportions: %w", err)
	}

	query := `
		INSERT INTO character_appearance (
			id, character_id, face_type, skin_color, eye_color,
			hair_style, hair_color, facial_hair_style, facial_hair_color,
			body_type, height, body_proportions, scars, tattoos,
			accessories, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)`

	_, err = r.db.ExecContext(ctx, query,
		appearance.ID,
		appearance.CharacterID,
		appearance.FaceType,
		appearance.SkinColor,
		appearance.EyeColor,
		appearance.HairStyle,
		appearance.HairColor,
		appearance.FacialHairStyle,
		appearance.FacialHairColor,
		appearance.BodyType,
		appearance.Height,
		bodyPropsJSON,
		pq.Array(appearance.Scars),
		pq.Array(appearance.Tattoos),
		pq.Array(appearance.Accessories),
		appearance.CreatedAt,
		appearance.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create appearance: %w", err)
	}

	return nil
}

// GetByCharacterID retrieves appearance by character ID
func (r *PostgresAppearanceRepository) GetByCharacterID(ctx context.Context, characterID uuid.UUID) (*character.Appearance, error) {
	query := `
		SELECT 
			id, character_id, face_type, skin_color, eye_color,
			hair_style, hair_color, facial_hair_style, facial_hair_color,
			body_type, height, body_proportions, scars, tattoos,
			accessories, created_at, updated_at
		FROM character_appearance
		WHERE character_id = $1`

	var appearance character.Appearance
	var bodyPropsJSON []byte

	err := r.db.QueryRowContext(ctx, query, characterID).Scan(
		&appearance.ID,
		&appearance.CharacterID,
		&appearance.FaceType,
		&appearance.SkinColor,
		&appearance.EyeColor,
		&appearance.HairStyle,
		&appearance.HairColor,
		&appearance.FacialHairStyle,
		&appearance.FacialHairColor,
		&appearance.BodyType,
		&appearance.Height,
		&bodyPropsJSON,
		pq.Array(&appearance.Scars),
		pq.Array(&appearance.Tattoos),
		pq.Array(&appearance.Accessories),
		&appearance.CreatedAt,
		&appearance.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, character.ErrAppearanceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get appearance: %w", err)
	}

	// Parse body proportions JSON
	if err := json.Unmarshal(bodyPropsJSON, &appearance.BodyProportions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal body proportions: %w", err)
	}

	return &appearance, nil
}

// Update updates a character appearance in the database
func (r *PostgresAppearanceRepository) Update(ctx context.Context, appearance *character.Appearance) error {
	// Convert body proportions to JSON
	bodyPropsJSON, err := json.Marshal(appearance.BodyProportions)
	if err != nil {
		return fmt.Errorf("failed to marshal body proportions: %w", err)
	}

	query := `
		UPDATE character_appearance SET
			face_type = $2,
			skin_color = $3,
			eye_color = $4,
			hair_style = $5,
			hair_color = $6,
			facial_hair_style = $7,
			facial_hair_color = $8,
			body_type = $9,
			height = $10,
			body_proportions = $11,
			scars = $12,
			tattoos = $13,
			accessories = $14,
			updated_at = $15
		WHERE character_id = $1`

	result, err := r.db.ExecContext(ctx, query,
		appearance.CharacterID,
		appearance.FaceType,
		appearance.SkinColor,
		appearance.EyeColor,
		appearance.HairStyle,
		appearance.HairColor,
		appearance.FacialHairStyle,
		appearance.FacialHairColor,
		appearance.BodyType,
		appearance.Height,
		bodyPropsJSON,
		pq.Array(appearance.Scars),
		pq.Array(appearance.Tattoos),
		pq.Array(appearance.Accessories),
		appearance.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update appearance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return character.ErrAppearanceNotFound
	}

	return nil
}

// Delete deletes a character appearance
func (r *PostgresAppearanceRepository) Delete(ctx context.Context, characterID uuid.UUID) error {
	query := `DELETE FROM character_appearance WHERE character_id = $1`

	result, err := r.db.ExecContext(ctx, query, characterID)
	if err != nil {
		return fmt.Errorf("failed to delete appearance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return character.ErrAppearanceNotFound
	}

	return nil
}