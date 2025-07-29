package character

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/character"
	portsCharacter "github.com/mmorpg-template/backend/internal/ports/character"
	"github.com/mmorpg-template/backend/pkg/logger"
)

// TransactionalCharacterService extends the base CharacterService with transaction support
type TransactionalCharacterService struct {
	*CharacterService
	db *sql.DB
}

// NewTransactionalCharacterService creates a new transactional character service
func NewTransactionalCharacterService(
	db *sql.DB,
	characterRepo portsCharacter.CharacterRepository,
	appearanceRepo portsCharacter.AppearanceRepository,
	statsRepo portsCharacter.StatsRepository,
	positionRepo portsCharacter.PositionRepository,
	cache portsCharacter.CharacterCache,
	eventPublisher portsCharacter.EventPublisher,
	config *Config,
	logger logger.Logger,
) *TransactionalCharacterService {
	return &TransactionalCharacterService{
		CharacterService: NewCharacterService(
			characterRepo,
			appearanceRepo,
			statsRepo,
			positionRepo,
			cache,
			eventPublisher,
			config,
			logger,
		),
		db: db,
	}
}

// CreateCharacterWithTransaction creates a new character within a database transaction
func (s *TransactionalCharacterService) CreateCharacterWithTransaction(ctx context.Context, req *portsCharacter.CreateCharacterRequest) (*character.Character, error) {
	// Validate user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, character.ErrInvalidUserID
	}

	// Pre-transaction validations
	canCreate, err := s.CanCreateCharacter(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check character limit: %w", err)
	}
	if !canCreate {
		return nil, character.ErrCharacterLimitReached
	}

	if err := s.validateCharacterName(req.Name); err != nil {
		return nil, err
	}

	exists, err := s.characterRepo.NameExists(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check name availability: %w", err)
	}
	if exists {
		return nil, character.ErrCharacterNameTaken
	}

	if !character.IsValidClass(req.ClassType) {
		return nil, character.ErrInvalidClass
	}
	if !character.IsValidRace(req.Race) {
		return nil, character.ErrInvalidRace
	}
	if !character.IsValidGender(req.Gender) {
		return nil, character.ErrInvalidGender
	}

	if req.SlotNumber < 1 || req.SlotNumber > 100 {
		return nil, character.ErrInvalidSlotNumber
	}

	// Execute creation in transaction
	var createdChar *character.Character
	
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	// Use transaction-scoped repositories
	txCharRepo := &transactionalCharacterRepo{tx: tx}
	txAppearanceRepo := &transactionalAppearanceRepo{tx: tx}
	txStatsRepo := &transactionalStatsRepo{tx: tx}
	txPositionRepo := &transactionalPositionRepo{tx: tx}

	// Check slot availability within transaction
	existingChar, err := txCharRepo.GetByUserIDAndSlot(ctx, userID, req.SlotNumber)
	if err == nil && existingChar != nil && !existingChar.IsDeleted {
		_ = tx.Rollback()
		return nil, character.ErrSlotOccupied
	}

	// Create character entity
	char := character.NewCharacter(userID, req.Name, req.SlotNumber, req.ClassType, req.Race, req.Gender)

	// Create character
	if err := txCharRepo.Create(ctx, char); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to create character: %w", err)
	}

	// Create appearance
	appearance := character.NewAppearance(char.ID)
	appearance.ApplyRaceDefaults(req.Race)
	appearance.ApplyGenderDefaults(req.Gender)
	
	if req.Appearance != nil {
		s.applyAppearanceOptions(appearance, req.Appearance)
	}

	if err := appearance.Validate(); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("invalid appearance: %w", err)
	}

	if err := txAppearanceRepo.Create(ctx, appearance); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to create appearance: %w", err)
	}

	// Create stats
	stats := character.NewStats(char.ID, req.ClassType)
	if err := txStatsRepo.Create(ctx, stats); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to create stats: %w", err)
	}

	// Create position
	position := character.NewPosition(char.ID)
	position.ApplyClassStartingPosition(req.ClassType)
	if err := txPositionRepo.Create(ctx, position); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to create position: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	createdChar = char

	s.logger.WithFields(map[string]interface{}{
		"character_id": char.ID,
		"user_id":      userID,
		"name":         char.Name,
		"class":        char.ClassType,
	}).Info("Character created successfully with transaction")

	return createdChar, nil
}

// Transactional repository wrappers
type transactionalCharacterRepo struct {
	tx *sql.Tx
}

func (r *transactionalCharacterRepo) Create(ctx context.Context, char *character.Character) error {
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

	_, err := r.tx.ExecContext(ctx, query,
		char.ID, char.UserID, char.Name, char.SlotNumber, char.Level,
		char.Experience, char.ClassType, char.Race, char.Gender,
		char.IsDeleted, char.DeletedAt, char.DeletionScheduledAt,
		char.CreatedAt, char.UpdatedAt, char.LastPlayedAt, char.TotalPlayTime,
	)
	return err
}

func (r *transactionalCharacterRepo) GetByUserIDAndSlot(ctx context.Context, userID uuid.UUID, slot int) (*character.Character, error) {
	query := `
		SELECT 
			id, user_id, name, slot_number, level, experience,
			class_type, race, gender, is_deleted, deleted_at,
			deletion_scheduled_at, created_at, updated_at,
			last_played_at, total_play_time
		FROM characters
		WHERE user_id = $1 AND slot_number = $2`

	var char character.Character
	err := r.tx.QueryRowContext(ctx, query, userID, slot).Scan(
		&char.ID, &char.UserID, &char.Name, &char.SlotNumber,
		&char.Level, &char.Experience, &char.ClassType, &char.Race,
		&char.Gender, &char.IsDeleted, &char.DeletedAt,
		&char.DeletionScheduledAt, &char.CreatedAt, &char.UpdatedAt,
		&char.LastPlayedAt, &char.TotalPlayTime,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &char, nil
}

type transactionalAppearanceRepo struct {
	tx *sql.Tx
}

func (r *transactionalAppearanceRepo) Create(ctx context.Context, appearance *character.Appearance) error {
	query := `
		INSERT INTO character_appearance (
			id, character_id, face_type, skin_color, eye_color,
			hair_style, hair_color, facial_hair_style, facial_hair_color,
			body_type, height, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	_, err := r.tx.ExecContext(ctx, query,
		appearance.ID, appearance.CharacterID, appearance.FaceType,
		appearance.SkinColor, appearance.EyeColor, appearance.HairStyle,
		appearance.HairColor, appearance.FacialHairStyle, appearance.FacialHairColor,
		appearance.BodyType, appearance.Height, appearance.CreatedAt, appearance.UpdatedAt,
	)
	return err
}

type transactionalStatsRepo struct {
	tx *sql.Tx
}

func (r *transactionalStatsRepo) Create(ctx context.Context, stats *character.Stats) error {
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

	_, err := r.tx.ExecContext(ctx, query,
		stats.ID, stats.CharacterID, stats.Strength, stats.Dexterity,
		stats.Intelligence, stats.Wisdom, stats.Constitution, stats.Charisma,
		stats.HealthCurrent, stats.HealthMax, stats.ManaCurrent, stats.ManaMax,
		stats.StaminaCurrent, stats.StaminaMax, stats.AttackPower, stats.SpellPower,
		stats.Defense, stats.CriticalChance, stats.CriticalDamage,
		stats.DodgeChance, stats.BlockChance, stats.MovementSpeed,
		stats.AttackSpeed, stats.CastSpeed, stats.HealthRegen, stats.ManaRegen,
		stats.StaminaRegen, stats.StatPointsAvailable, stats.SkillPointsAvailable,
		stats.CreatedAt, stats.UpdatedAt,
	)
	return err
}

type transactionalPositionRepo struct {
	tx *sql.Tx
}

func (r *transactionalPositionRepo) Create(ctx context.Context, position *character.Position) error {
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

	_, err := r.tx.ExecContext(ctx, query,
		position.ID, position.CharacterID, position.WorldID, position.ZoneID,
		position.MapID, position.PositionX, position.PositionY, position.PositionZ,
		position.RotationPitch, position.RotationYaw, position.RotationRoll,
		position.VelocityX, position.VelocityY, position.VelocityZ,
		position.InstanceID, position.InstanceType,
		position.SafePositionX, position.SafePositionY, position.SafePositionZ,
		position.SafeWorldID, position.SafeZoneID,
		position.LastMovement, position.CreatedAt, position.UpdatedAt,
	)
	return err
}