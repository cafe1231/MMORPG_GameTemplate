package character

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/character"
	portsCharacter "github.com/mmorpg-template/backend/internal/ports/character"
)

// PostgresPositionRepository implements the PositionRepository interface using PostgreSQL
type PostgresPositionRepository struct {
	db DBExecutor
}

// NewPostgresPositionRepository creates a new PostgreSQL position repository
func NewPostgresPositionRepository(db *sql.DB) *PostgresPositionRepository {
	return &PostgresPositionRepository{
		db: db,
	}
}

// NewTransactionalPositionRepository creates a new PostgreSQL position repository using a transaction
func NewTransactionalPositionRepository(tx *sql.Tx) *PostgresPositionRepository {
	return &PostgresPositionRepository{
		db: tx,
	}
}

// Create creates a new character position in the database
func (r *PostgresPositionRepository) Create(ctx context.Context, position *character.Position) error {
	query := `
		INSERT INTO character_position (
			id, character_id, world_id, zone_id, map_id,
			position_x, position_y, position_z,
			rotation_pitch, rotation_yaw, rotation_roll,
			velocity_x, velocity_y, velocity_z,
			instance_id, instance_type,
			safe_position_x, safe_position_y, safe_position_z,
			safe_world_id, safe_zone_id,
			last_movement, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21, $22, $23, $24
		)`

	_, err := r.db.ExecContext(ctx, query,
		position.ID,
		position.CharacterID,
		position.WorldID,
		position.ZoneID,
		position.MapID,
		position.PositionX,
		position.PositionY,
		position.PositionZ,
		position.RotationPitch,
		position.RotationYaw,
		position.RotationRoll,
		position.VelocityX,
		position.VelocityY,
		position.VelocityZ,
		position.InstanceID,
		position.InstanceType,
		position.SafePositionX,
		position.SafePositionY,
		position.SafePositionZ,
		position.SafeWorldID,
		position.SafeZoneID,
		position.LastMovement,
		position.CreatedAt,
		position.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create position: %w", err)
	}

	return nil
}

// GetByCharacterID retrieves position by character ID
func (r *PostgresPositionRepository) GetByCharacterID(ctx context.Context, characterID uuid.UUID) (*character.Position, error) {
	query := `
		SELECT 
			id, character_id, world_id, zone_id, map_id,
			position_x, position_y, position_z,
			rotation_pitch, rotation_yaw, rotation_roll,
			velocity_x, velocity_y, velocity_z,
			instance_id, instance_type,
			safe_position_x, safe_position_y, safe_position_z,
			safe_world_id, safe_zone_id,
			last_movement, created_at, updated_at
		FROM character_position
		WHERE character_id = $1`

	var position character.Position
	err := r.db.QueryRowContext(ctx, query, characterID).Scan(
		&position.ID,
		&position.CharacterID,
		&position.WorldID,
		&position.ZoneID,
		&position.MapID,
		&position.PositionX,
		&position.PositionY,
		&position.PositionZ,
		&position.RotationPitch,
		&position.RotationYaw,
		&position.RotationRoll,
		&position.VelocityX,
		&position.VelocityY,
		&position.VelocityZ,
		&position.InstanceID,
		&position.InstanceType,
		&position.SafePositionX,
		&position.SafePositionY,
		&position.SafePositionZ,
		&position.SafeWorldID,
		&position.SafeZoneID,
		&position.LastMovement,
		&position.CreatedAt,
		&position.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, character.ErrPositionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get position: %w", err)
	}

	return &position, nil
}

// Update updates a character position in the database
func (r *PostgresPositionRepository) Update(ctx context.Context, position *character.Position) error {
	query := `
		UPDATE character_position SET
			world_id = $2,
			zone_id = $3,
			map_id = $4,
			position_x = $5,
			position_y = $6,
			position_z = $7,
			rotation_pitch = $8,
			rotation_yaw = $9,
			rotation_roll = $10,
			velocity_x = $11,
			velocity_y = $12,
			velocity_z = $13,
			instance_id = $14,
			instance_type = $15,
			safe_position_x = $16,
			safe_position_y = $17,
			safe_position_z = $18,
			safe_world_id = $19,
			safe_zone_id = $20,
			last_movement = $21,
			updated_at = $22
		WHERE character_id = $1`

	result, err := r.db.ExecContext(ctx, query,
		position.CharacterID,
		position.WorldID,
		position.ZoneID,
		position.MapID,
		position.PositionX,
		position.PositionY,
		position.PositionZ,
		position.RotationPitch,
		position.RotationYaw,
		position.RotationRoll,
		position.VelocityX,
		position.VelocityY,
		position.VelocityZ,
		position.InstanceID,
		position.InstanceType,
		position.SafePositionX,
		position.SafePositionY,
		position.SafePositionZ,
		position.SafeWorldID,
		position.SafeZoneID,
		position.LastMovement,
		position.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return character.ErrPositionNotFound
	}

	return nil
}

// Delete deletes a character position
func (r *PostgresPositionRepository) Delete(ctx context.Context, characterID uuid.UUID) error {
	query := `DELETE FROM character_position WHERE character_id = $1`

	result, err := r.db.ExecContext(ctx, query, characterID)
	if err != nil {
		return fmt.Errorf("failed to delete position: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return character.ErrPositionNotFound
	}

	return nil
}

// FindNearbyCharacters finds characters near a given character
func (r *PostgresPositionRepository) FindNearbyCharacters(ctx context.Context, characterID uuid.UUID, maxDistance float64) ([]*portsCharacter.NearbyCharacter, error) {
	query := `
		SELECT * FROM find_nearby_characters($1, $2)`

	rows, err := r.db.QueryContext(ctx, query, characterID, maxDistance)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby characters: %w", err)
	}
	defer rows.Close()

	var nearby []*portsCharacter.NearbyCharacter
	for rows.Next() {
		var nc portsCharacter.NearbyCharacter
		err := rows.Scan(
			&nc.CharacterID,
			&nc.CharacterName,
			&nc.Distance,
			&nc.Position.X,
			&nc.Position.Y,
			&nc.Position.Z,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan nearby character: %w", err)
		}
		nearby = append(nearby, &nc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating nearby characters: %w", err)
	}

	return nearby, nil
}

// GetCharactersInZone retrieves all characters in a specific zone
func (r *PostgresPositionRepository) GetCharactersInZone(ctx context.Context, worldID, zoneID string) ([]*character.Position, error) {
	query := `
		SELECT 
			p.id, p.character_id, p.world_id, p.zone_id, p.map_id,
			p.position_x, p.position_y, p.position_z,
			p.rotation_pitch, p.rotation_yaw, p.rotation_roll,
			p.velocity_x, p.velocity_y, p.velocity_z,
			p.instance_id, p.instance_type,
			p.safe_position_x, p.safe_position_y, p.safe_position_z,
			p.safe_world_id, p.safe_zone_id,
			p.last_movement, p.created_at, p.updated_at
		FROM character_position p
		JOIN characters c ON c.id = p.character_id
		WHERE p.world_id = $1 AND p.zone_id = $2 AND c.is_deleted = false`

	rows, err := r.db.QueryContext(ctx, query, worldID, zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to query characters in zone: %w", err)
	}
	defer rows.Close()

	var positions []*character.Position
	for rows.Next() {
		var position character.Position
		err := rows.Scan(
			&position.ID,
			&position.CharacterID,
			&position.WorldID,
			&position.ZoneID,
			&position.MapID,
			&position.PositionX,
			&position.PositionY,
			&position.PositionZ,
			&position.RotationPitch,
			&position.RotationYaw,
			&position.RotationRoll,
			&position.VelocityX,
			&position.VelocityY,
			&position.VelocityZ,
			&position.InstanceID,
			&position.InstanceType,
			&position.SafePositionX,
			&position.SafePositionY,
			&position.SafePositionZ,
			&position.SafeWorldID,
			&position.SafeZoneID,
			&position.LastMovement,
			&position.CreatedAt,
			&position.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		positions = append(positions, &position)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating positions: %w", err)
	}

	return positions, nil
}

// GetCharactersInInstance retrieves all characters in a specific instance
func (r *PostgresPositionRepository) GetCharactersInInstance(ctx context.Context, instanceID uuid.UUID) ([]*character.Position, error) {
	query := `
		SELECT 
			p.id, p.character_id, p.world_id, p.zone_id, p.map_id,
			p.position_x, p.position_y, p.position_z,
			p.rotation_pitch, p.rotation_yaw, p.rotation_roll,
			p.velocity_x, p.velocity_y, p.velocity_z,
			p.instance_id, p.instance_type,
			p.safe_position_x, p.safe_position_y, p.safe_position_z,
			p.safe_world_id, p.safe_zone_id,
			p.last_movement, p.created_at, p.updated_at
		FROM character_position p
		JOIN characters c ON c.id = p.character_id
		WHERE p.instance_id = $1 AND c.is_deleted = false`

	rows, err := r.db.QueryContext(ctx, query, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query characters in instance: %w", err)
	}
	defer rows.Close()

	var positions []*character.Position
	for rows.Next() {
		var position character.Position
		err := rows.Scan(
			&position.ID,
			&position.CharacterID,
			&position.WorldID,
			&position.ZoneID,
			&position.MapID,
			&position.PositionX,
			&position.PositionY,
			&position.PositionZ,
			&position.RotationPitch,
			&position.RotationYaw,
			&position.RotationRoll,
			&position.VelocityX,
			&position.VelocityY,
			&position.VelocityZ,
			&position.InstanceID,
			&position.InstanceType,
			&position.SafePositionX,
			&position.SafePositionY,
			&position.SafePositionZ,
			&position.SafeWorldID,
			&position.SafeZoneID,
			&position.LastMovement,
			&position.CreatedAt,
			&position.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		positions = append(positions, &position)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating positions: %w", err)
	}

	return positions, nil
}

// SaveSafePosition saves the current position as the safe position
func (r *PostgresPositionRepository) SaveSafePosition(ctx context.Context, characterID uuid.UUID) error {
	query := `SELECT save_safe_position($1)`

	_, err := r.db.ExecContext(ctx, query, characterID)
	if err != nil {
		return fmt.Errorf("failed to save safe position: %w", err)
	}

	return nil
}

// TeleportToSafePosition teleports the character to their safe position
func (r *PostgresPositionRepository) TeleportToSafePosition(ctx context.Context, characterID uuid.UUID) error {
	query := `SELECT teleport_to_safe_position($1)`

	_, err := r.db.ExecContext(ctx, query, characterID)
	if err != nil {
		return fmt.Errorf("failed to teleport to safe position: %w", err)
	}

	return nil
}