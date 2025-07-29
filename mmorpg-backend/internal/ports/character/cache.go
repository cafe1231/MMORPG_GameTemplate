package character

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/character"
)

// CharacterCache defines the interface for caching character-related data
type CharacterCache interface {
	// Character caching
	SetCharacter(ctx context.Context, char *character.Character, expiration time.Duration) error
	GetCharacter(ctx context.Context, characterID uuid.UUID) (*character.Character, error)
	DeleteCharacter(ctx context.Context, characterID uuid.UUID) error
	
	// Character list caching
	SetUserCharacters(ctx context.Context, userID uuid.UUID, characters []*character.Character, expiration time.Duration) error
	GetUserCharacters(ctx context.Context, userID uuid.UUID) ([]*character.Character, error)
	DeleteUserCharacters(ctx context.Context, userID uuid.UUID) error
	
	// Character count caching (for slot validation)
	SetCharacterCount(ctx context.Context, userID uuid.UUID, count int, expiration time.Duration) error
	GetCharacterCount(ctx context.Context, userID uuid.UUID) (int, bool, error)
	DeleteCharacterCount(ctx context.Context, userID uuid.UUID) error
	
	// Selected character caching (for gameplay session)
	SetSelectedCharacter(ctx context.Context, userID uuid.UUID, characterID uuid.UUID, expiration time.Duration) error
	GetSelectedCharacter(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	DeleteSelectedCharacter(ctx context.Context, userID uuid.UUID) error
	
	// Appearance caching
	SetAppearance(ctx context.Context, appearance *character.Appearance, expiration time.Duration) error
	GetAppearance(ctx context.Context, characterID uuid.UUID) (*character.Appearance, error)
	DeleteAppearance(ctx context.Context, characterID uuid.UUID) error
	
	// Stats caching
	SetStats(ctx context.Context, stats *character.Stats, expiration time.Duration) error
	GetStats(ctx context.Context, characterID uuid.UUID) (*character.Stats, error)
	DeleteStats(ctx context.Context, characterID uuid.UUID) error
	
	// Position caching
	SetPosition(ctx context.Context, position *character.Position, expiration time.Duration) error
	GetPosition(ctx context.Context, characterID uuid.UUID) (*character.Position, error)
	DeletePosition(ctx context.Context, characterID uuid.UUID) error
	
	// Cache warming
	WarmCharacterCache(ctx context.Context, characterID uuid.UUID) error
	
	// Bulk invalidation
	InvalidateCharacterData(ctx context.Context, characterID uuid.UUID) error
	InvalidateUserData(ctx context.Context, userID uuid.UUID) error
}

// CacheTTL defines cache time-to-live values for different data types
type CacheTTL struct {
	Character         time.Duration // Individual character data (10 minutes)
	CharacterList     time.Duration // User's character list (5 minutes)
	CharacterCount    time.Duration // Character count for slot validation (1 minute)
	SelectedCharacter time.Duration // Selected character for gameplay (session duration - 24 hours)
	Appearance        time.Duration // Character appearance (10 minutes)
	Stats             time.Duration // Character stats (10 minutes)
	Position          time.Duration // Character position (30 seconds for real-time updates)
}

// DefaultCacheTTL returns the default cache TTL configuration
func DefaultCacheTTL() *CacheTTL {
	return &CacheTTL{
		Character:         10 * time.Minute,
		CharacterList:     5 * time.Minute,
		CharacterCount:    1 * time.Minute,
		SelectedCharacter: 24 * time.Hour,
		Appearance:        10 * time.Minute,
		Stats:             10 * time.Minute,
		Position:          30 * time.Second,
	}
}