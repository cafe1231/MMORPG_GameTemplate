package character

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/character"
)

// PostgresStatsRepository implements the StatsRepository interface using PostgreSQL
type PostgresStatsRepository struct {
	db DBExecutor
}

// NewPostgresStatsRepository creates a new PostgreSQL stats repository
func NewPostgresStatsRepository(db *sql.DB) *PostgresStatsRepository {
	return &PostgresStatsRepository{
		db: db,
	}
}

// NewTransactionalStatsRepository creates a new PostgreSQL stats repository using a transaction
func NewTransactionalStatsRepository(tx *sql.Tx) *PostgresStatsRepository {
	return &PostgresStatsRepository{
		db: tx,
	}
}

// Create creates new character stats in the database
func (r *PostgresStatsRepository) Create(ctx context.Context, stats *character.Stats) error {
	query := `
		INSERT INTO character_stats (
			id, character_id, strength, dexterity, intelligence, wisdom,
			constitution, charisma, health_current, health_max, mana_current,
			mana_max, stamina_current, stamina_max, attack_power, spell_power,
			defense, critical_chance, critical_damage, dodge_chance, block_chance,
			movement_speed, attack_speed, cast_speed, health_regen, mana_regen,
			stamina_regen, stat_points_available, skill_points_available,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28,
			$29, $30, $31
		)`

	_, err := r.db.ExecContext(ctx, query,
		stats.ID,
		stats.CharacterID,
		stats.Strength,
		stats.Dexterity,
		stats.Intelligence,
		stats.Wisdom,
		stats.Constitution,
		stats.Charisma,
		stats.HealthCurrent,
		stats.HealthMax,
		stats.ManaCurrent,
		stats.ManaMax,
		stats.StaminaCurrent,
		stats.StaminaMax,
		stats.AttackPower,
		stats.SpellPower,
		stats.Defense,
		stats.CriticalChance,
		stats.CriticalDamage,
		stats.DodgeChance,
		stats.BlockChance,
		stats.MovementSpeed,
		stats.AttackSpeed,
		stats.CastSpeed,
		stats.HealthRegen,
		stats.ManaRegen,
		stats.StaminaRegen,
		stats.StatPointsAvailable,
		stats.SkillPointsAvailable,
		stats.CreatedAt,
		stats.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create stats: %w", err)
	}

	return nil
}

// GetByCharacterID retrieves stats by character ID
func (r *PostgresStatsRepository) GetByCharacterID(ctx context.Context, characterID uuid.UUID) (*character.Stats, error) {
	query := `
		SELECT 
			id, character_id, strength, dexterity, intelligence, wisdom,
			constitution, charisma, health_current, health_max, mana_current,
			mana_max, stamina_current, stamina_max, attack_power, spell_power,
			defense, critical_chance, critical_damage, dodge_chance, block_chance,
			movement_speed, attack_speed, cast_speed, health_regen, mana_regen,
			stamina_regen, stat_points_available, skill_points_available,
			created_at, updated_at
		FROM character_stats
		WHERE character_id = $1`

	var stats character.Stats
	err := r.db.QueryRowContext(ctx, query, characterID).Scan(
		&stats.ID,
		&stats.CharacterID,
		&stats.Strength,
		&stats.Dexterity,
		&stats.Intelligence,
		&stats.Wisdom,
		&stats.Constitution,
		&stats.Charisma,
		&stats.HealthCurrent,
		&stats.HealthMax,
		&stats.ManaCurrent,
		&stats.ManaMax,
		&stats.StaminaCurrent,
		&stats.StaminaMax,
		&stats.AttackPower,
		&stats.SpellPower,
		&stats.Defense,
		&stats.CriticalChance,
		&stats.CriticalDamage,
		&stats.DodgeChance,
		&stats.BlockChance,
		&stats.MovementSpeed,
		&stats.AttackSpeed,
		&stats.CastSpeed,
		&stats.HealthRegen,
		&stats.ManaRegen,
		&stats.StaminaRegen,
		&stats.StatPointsAvailable,
		&stats.SkillPointsAvailable,
		&stats.CreatedAt,
		&stats.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, character.ErrStatsNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return &stats, nil
}

// Update updates character stats in the database
func (r *PostgresStatsRepository) Update(ctx context.Context, stats *character.Stats) error {
	query := `
		UPDATE character_stats SET
			strength = $2,
			dexterity = $3,
			intelligence = $4,
			wisdom = $5,
			constitution = $6,
			charisma = $7,
			health_current = $8,
			health_max = $9,
			mana_current = $10,
			mana_max = $11,
			stamina_current = $12,
			stamina_max = $13,
			attack_power = $14,
			spell_power = $15,
			defense = $16,
			critical_chance = $17,
			critical_damage = $18,
			dodge_chance = $19,
			block_chance = $20,
			movement_speed = $21,
			attack_speed = $22,
			cast_speed = $23,
			health_regen = $24,
			mana_regen = $25,
			stamina_regen = $26,
			stat_points_available = $27,
			skill_points_available = $28,
			updated_at = $29
		WHERE character_id = $1`

	result, err := r.db.ExecContext(ctx, query,
		stats.CharacterID,
		stats.Strength,
		stats.Dexterity,
		stats.Intelligence,
		stats.Wisdom,
		stats.Constitution,
		stats.Charisma,
		stats.HealthCurrent,
		stats.HealthMax,
		stats.ManaCurrent,
		stats.ManaMax,
		stats.StaminaCurrent,
		stats.StaminaMax,
		stats.AttackPower,
		stats.SpellPower,
		stats.Defense,
		stats.CriticalChance,
		stats.CriticalDamage,
		stats.DodgeChance,
		stats.BlockChance,
		stats.MovementSpeed,
		stats.AttackSpeed,
		stats.CastSpeed,
		stats.HealthRegen,
		stats.ManaRegen,
		stats.StaminaRegen,
		stats.StatPointsAvailable,
		stats.SkillPointsAvailable,
		stats.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update stats: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return character.ErrStatsNotFound
	}

	return nil
}

// Delete deletes character stats
func (r *PostgresStatsRepository) Delete(ctx context.Context, characterID uuid.UUID) error {
	query := `DELETE FROM character_stats WHERE character_id = $1`

	result, err := r.db.ExecContext(ctx, query, characterID)
	if err != nil {
		return fmt.Errorf("failed to delete stats: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return character.ErrStatsNotFound
	}

	return nil
}

// GetMultipleByCharacterIDs retrieves stats for multiple characters
func (r *PostgresStatsRepository) GetMultipleByCharacterIDs(ctx context.Context, characterIDs []uuid.UUID) (map[uuid.UUID]*character.Stats, error) {
	if len(characterIDs) == 0 {
		return make(map[uuid.UUID]*character.Stats), nil
	}

	// Build query with placeholders
	query := `
		SELECT 
			id, character_id, strength, dexterity, intelligence, wisdom,
			constitution, charisma, health_current, health_max, mana_current,
			mana_max, stamina_current, stamina_max, attack_power, spell_power,
			defense, critical_chance, critical_damage, dodge_chance, block_chance,
			movement_speed, attack_speed, cast_speed, health_regen, mana_regen,
			stamina_regen, stat_points_available, skill_points_available,
			created_at, updated_at
		FROM character_stats
		WHERE character_id = ANY($1)`

	// Convert UUIDs to interface slice
	ids := make([]interface{}, len(characterIDs))
	for i, id := range characterIDs {
		ids[i] = id
	}

	rows, err := r.db.QueryContext(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to query stats: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]*character.Stats)
	for rows.Next() {
		var stats character.Stats
		err := rows.Scan(
			&stats.ID,
			&stats.CharacterID,
			&stats.Strength,
			&stats.Dexterity,
			&stats.Intelligence,
			&stats.Wisdom,
			&stats.Constitution,
			&stats.Charisma,
			&stats.HealthCurrent,
			&stats.HealthMax,
			&stats.ManaCurrent,
			&stats.ManaMax,
			&stats.StaminaCurrent,
			&stats.StaminaMax,
			&stats.AttackPower,
			&stats.SpellPower,
			&stats.Defense,
			&stats.CriticalChance,
			&stats.CriticalDamage,
			&stats.DodgeChance,
			&stats.BlockChance,
			&stats.MovementSpeed,
			&stats.AttackSpeed,
			&stats.CastSpeed,
			&stats.HealthRegen,
			&stats.ManaRegen,
			&stats.StaminaRegen,
			&stats.StatPointsAvailable,
			&stats.SkillPointsAvailable,
			&stats.CreatedAt,
			&stats.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}
		result[stats.CharacterID] = &stats
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stats: %w", err)
	}

	return result, nil
}