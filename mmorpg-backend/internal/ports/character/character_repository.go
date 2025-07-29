package character

import (
	"context"

	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/character"
)

// CharacterRepository defines the interface for character data persistence
type CharacterRepository interface {
	// Character CRUD operations
	Create(ctx context.Context, char *character.Character) error
	GetByID(ctx context.Context, id uuid.UUID) (*character.Character, error)
	GetByName(ctx context.Context, name string) (*character.Character, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*character.Character, error)
	GetByUserIDAndSlot(ctx context.Context, userID uuid.UUID, slot int) (*character.Character, error)
	Update(ctx context.Context, char *character.Character) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Soft delete operations
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error
	CleanupDeleted(ctx context.Context) error
	
	// Validation
	NameExists(ctx context.Context, name string) (bool, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)
}

// AppearanceRepository defines the interface for character appearance persistence
type AppearanceRepository interface {
	Create(ctx context.Context, appearance *character.Appearance) error
	GetByCharacterID(ctx context.Context, characterID uuid.UUID) (*character.Appearance, error)
	Update(ctx context.Context, appearance *character.Appearance) error
	Delete(ctx context.Context, characterID uuid.UUID) error
}

// StatsRepository defines the interface for character stats persistence
type StatsRepository interface {
	Create(ctx context.Context, stats *character.Stats) error
	GetByCharacterID(ctx context.Context, characterID uuid.UUID) (*character.Stats, error)
	Update(ctx context.Context, stats *character.Stats) error
	Delete(ctx context.Context, characterID uuid.UUID) error
	
	// Bulk operations for performance
	GetMultipleByCharacterIDs(ctx context.Context, characterIDs []uuid.UUID) (map[uuid.UUID]*character.Stats, error)
}

// PositionRepository defines the interface for character position persistence
type PositionRepository interface {
	Create(ctx context.Context, position *character.Position) error
	GetByCharacterID(ctx context.Context, characterID uuid.UUID) (*character.Position, error)
	Update(ctx context.Context, position *character.Position) error
	Delete(ctx context.Context, characterID uuid.UUID) error
	
	// Spatial queries
	FindNearbyCharacters(ctx context.Context, characterID uuid.UUID, maxDistance float64) ([]*NearbyCharacter, error)
	GetCharactersInZone(ctx context.Context, worldID, zoneID string) ([]*character.Position, error)
	GetCharactersInInstance(ctx context.Context, instanceID uuid.UUID) ([]*character.Position, error)
	
	// Safe position operations
	SaveSafePosition(ctx context.Context, characterID uuid.UUID) error
	TeleportToSafePosition(ctx context.Context, characterID uuid.UUID) error
}

// NearbyCharacter represents a character near another character
type NearbyCharacter struct {
	CharacterID   uuid.UUID
	CharacterName string
	Distance      float64
	Position      character.Vector3
}