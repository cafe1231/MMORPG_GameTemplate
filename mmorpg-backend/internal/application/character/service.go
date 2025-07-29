package character

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/character"
	portsCharacter "github.com/mmorpg-template/backend/internal/ports/character"
	"github.com/mmorpg-template/backend/pkg/logger"
)

// Config holds the configuration for the character service
type Config struct {
	MaxCharactersPerUser      int
	MaxCharacterNameLength    int
	MinCharacterNameLength    int
	DefaultStartingLevel      int
	DefaultStartingExperience int64
}

// CharacterService implements the character service interface
type CharacterService struct {
	characterRepo  portsCharacter.CharacterRepository
	appearanceRepo portsCharacter.AppearanceRepository
	statsRepo      portsCharacter.StatsRepository
	positionRepo   portsCharacter.PositionRepository
	cache          portsCharacter.CharacterCache
	cacheTTL       *portsCharacter.CacheTTL
	eventPublisher portsCharacter.EventPublisher
	config         *Config
	logger         logger.Logger
}

// NewCharacterService creates a new character service instance
func NewCharacterService(
	characterRepo portsCharacter.CharacterRepository,
	appearanceRepo portsCharacter.AppearanceRepository,
	statsRepo portsCharacter.StatsRepository,
	positionRepo portsCharacter.PositionRepository,
	cache portsCharacter.CharacterCache,
	eventPublisher portsCharacter.EventPublisher,
	config *Config,
	logger logger.Logger,
) *CharacterService {
	return &CharacterService{
		characterRepo:  characterRepo,
		appearanceRepo: appearanceRepo,
		statsRepo:      statsRepo,
		positionRepo:   positionRepo,
		cache:          cache,
		cacheTTL:       portsCharacter.DefaultCacheTTL(),
		eventPublisher: eventPublisher,
		config:         config,
		logger:         logger,
	}
}

// CreateCharacter creates a new character for a user
func (s *CharacterService) CreateCharacter(ctx context.Context, req *portsCharacter.CreateCharacterRequest) (*character.Character, error) {
	// Validate user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, character.ErrInvalidUserID
	}

	// Check if user can create more characters
	canCreate, err := s.CanCreateCharacter(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check character limit: %w", err)
	}
	if !canCreate {
		return nil, character.ErrCharacterLimitReached
	}

	// Validate character name
	if err := s.validateCharacterName(req.Name); err != nil {
		return nil, err
	}

	// Check if name is already taken
	exists, err := s.characterRepo.NameExists(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check name availability: %w", err)
	}
	if exists {
		return nil, character.ErrCharacterNameTaken
	}

	// Validate class, race, and gender
	if !character.IsValidClass(req.ClassType) {
		return nil, character.ErrInvalidClass
	}
	if !character.IsValidRace(req.Race) {
		return nil, character.ErrInvalidRace
	}
	if !character.IsValidGender(req.Gender) {
		return nil, character.ErrInvalidGender
	}

	// Validate slot number
	if req.SlotNumber < 1 || req.SlotNumber > 100 {
		return nil, character.ErrInvalidSlotNumber
	}

	// Check if slot is occupied
	existingChar, err := s.characterRepo.GetByUserIDAndSlot(ctx, userID, req.SlotNumber)
	if err == nil && existingChar != nil && !existingChar.IsDeleted {
		return nil, character.ErrSlotOccupied
	}

	// Create character entity
	char := character.NewCharacter(userID, req.Name, req.SlotNumber, req.ClassType, req.Race, req.Gender)

	// Create character in database
	if err := s.characterRepo.Create(ctx, char); err != nil {
		return nil, fmt.Errorf("failed to create character: %w", err)
	}

	// Create appearance
	appearance := character.NewAppearance(char.ID)
	appearance.ApplyRaceDefaults(req.Race)
	appearance.ApplyGenderDefaults(req.Gender)
	
	// Apply custom appearance options if provided
	if req.Appearance != nil {
		s.applyAppearanceOptions(appearance, req.Appearance)
	}

	if err := appearance.Validate(); err != nil {
		// Rollback character creation
		s.characterRepo.Delete(ctx, char.ID)
		return nil, fmt.Errorf("invalid appearance: %w", err)
	}

	if err := s.appearanceRepo.Create(ctx, appearance); err != nil {
		// Rollback character creation
		s.characterRepo.Delete(ctx, char.ID)
		return nil, fmt.Errorf("failed to create appearance: %w", err)
	}

	// Create stats
	stats := character.NewStats(char.ID, req.ClassType)
	if err := s.statsRepo.Create(ctx, stats); err != nil {
		// Rollback
		s.appearanceRepo.Delete(ctx, char.ID)
		s.characterRepo.Delete(ctx, char.ID)
		return nil, fmt.Errorf("failed to create stats: %w", err)
	}

	// Create position
	position := character.NewPosition(char.ID)
	position.ApplyClassStartingPosition(req.ClassType)
	if err := s.positionRepo.Create(ctx, position); err != nil {
		// Rollback
		s.statsRepo.Delete(ctx, char.ID)
		s.appearanceRepo.Delete(ctx, char.ID)
		s.characterRepo.Delete(ctx, char.ID)
		return nil, fmt.Errorf("failed to create position: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"character_id": char.ID,
		"user_id":      userID,
		"name":         char.Name,
		"class":        char.ClassType,
	}).Info("Character created successfully")

	// Invalidate user caches
	if s.cache != nil {
		if err := s.cache.InvalidateUserData(ctx, userID); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate user cache after character creation")
		}
	}

	// Publish character created event
	if s.eventPublisher != nil {
		event := &character.CharacterCreatedEvent{
			BaseEvent: character.BaseEvent{
				EventType:   character.EventCharacterCreated,
				CharacterID: char.ID.String(),
				UserID:      userID.String(),
			},
			Name:       char.Name,
			ClassType:  char.ClassType,
			Race:       char.Race,
			Gender:     char.Gender,
			Level:      char.Level,
			SlotNumber: char.SlotNumber,
		}
		
		if err := s.eventPublisher.PublishCharacterCreated(ctx, event); err != nil {
			s.logger.WithError(err).Warn("Failed to publish character created event")
		}
	}

	return char, nil
}

// GetCharacter retrieves a character by ID
func (s *CharacterService) GetCharacter(ctx context.Context, characterID string) (*character.Character, error) {
	charID, err := uuid.Parse(characterID)
	if err != nil {
		return nil, character.ErrInvalidCharacterID
	}

	// Try to get from cache first
	if s.cache != nil {
		char, err := s.cache.GetCharacter(ctx, charID)
		if err == nil && char != nil {
			if char.IsDeleted {
				return nil, character.ErrCharacterDeleted
			}
			return char, nil
		}
		// Log cache miss but continue
		if err != nil {
			s.logger.WithError(err).Warn("Failed to get character from cache")
		}
	}

	// Get from repository
	char, err := s.characterRepo.GetByID(ctx, charID)
	if err != nil {
		return nil, character.ErrCharacterNotFound
	}

	if char.IsDeleted {
		return nil, character.ErrCharacterDeleted
	}

	// Cache the character
	if s.cache != nil {
		if err := s.cache.SetCharacter(ctx, char, s.cacheTTL.Character); err != nil {
			s.logger.WithError(err).Warn("Failed to cache character")
		}
	}

	return char, nil
}

// GetCharacterByName retrieves a character by name
func (s *CharacterService) GetCharacterByName(ctx context.Context, name string) (*character.Character, error) {
	char, err := s.characterRepo.GetByName(ctx, name)
	if err != nil {
		return nil, character.ErrCharacterNotFound
	}

	if char.IsDeleted {
		return nil, character.ErrCharacterDeleted
	}

	return char, nil
}

// ListCharactersByUser lists all characters for a user
func (s *CharacterService) ListCharactersByUser(ctx context.Context, userID string) ([]*character.Character, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, character.ErrInvalidUserID
	}

	// Try to get from cache first
	if s.cache != nil {
		characters, err := s.cache.GetUserCharacters(ctx, uid)
		if err == nil && characters != nil {
			return characters, nil
		}
		// Log cache miss but continue
		if err != nil {
			s.logger.WithError(err).Warn("Failed to get user characters from cache")
		}
	}

	// Get from repository
	characters, err := s.characterRepo.GetByUserID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to list characters: %w", err)
	}

	// Filter out deleted characters
	activeChars := make([]*character.Character, 0, len(characters))
	for _, char := range characters {
		if !char.IsDeleted {
			activeChars = append(activeChars, char)
		}
	}

	// Cache the character list
	if s.cache != nil {
		if err := s.cache.SetUserCharacters(ctx, uid, activeChars, s.cacheTTL.CharacterList); err != nil {
			s.logger.WithError(err).Warn("Failed to cache user characters")
		}
	}

	return activeChars, nil
}

// DeleteCharacter soft deletes a character
func (s *CharacterService) DeleteCharacter(ctx context.Context, characterID string, userID string) error {
	// Validate ownership
	if err := s.ValidateCharacterOwnership(ctx, characterID, userID); err != nil {
		return err
	}

	charID, _ := uuid.Parse(characterID)
	uid, _ := uuid.Parse(userID)
	
	if err := s.characterRepo.SoftDelete(ctx, charID); err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}

	// Get character info for event
	char, _ := s.characterRepo.GetByID(ctx, charID)
	
	s.logger.WithFields(map[string]interface{}{
		"character_id": characterID,
		"user_id":      userID,
	}).Info("Character soft deleted")

	// Publish character deleted event
	if s.eventPublisher != nil && char != nil {
		event := &character.CharacterDeletedEvent{
			BaseEvent: character.BaseEvent{
				EventType:   character.EventCharacterDeleted,
				CharacterID: characterID,
				UserID:      userID,
			},
			Name:       char.Name,
			SoftDelete: true,
		}
		
		if err := s.eventPublisher.PublishCharacterDeleted(ctx, event); err != nil {
			s.logger.WithError(err).Warn("Failed to publish character deleted event")
		}
	}

	// Invalidate caches
	if s.cache != nil {
		// Invalidate both character and user data
		if err := s.cache.InvalidateCharacterData(ctx, charID); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate character cache after deletion")
		}
		if err := s.cache.InvalidateUserData(ctx, uid); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate user cache after deletion")
		}
	}

	return nil
}

// RestoreCharacter restores a soft-deleted character
func (s *CharacterService) RestoreCharacter(ctx context.Context, characterID string, userID string) error {
	// Validate ownership
	if err := s.ValidateCharacterOwnership(ctx, characterID, userID); err != nil {
		return err
	}

	charID, _ := uuid.Parse(characterID)
	char, err := s.characterRepo.GetByID(ctx, charID)
	if err != nil {
		return character.ErrCharacterNotFound
	}

	if !char.CanBeRestored() {
		return character.ErrCharacterCannotBeRestored
	}

	if err := s.characterRepo.Restore(ctx, charID); err != nil {
		return fmt.Errorf("failed to restore character: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"character_id": characterID,
		"user_id":      userID,
	}).Info("Character restored")

	// Publish character restored event
	if s.eventPublisher != nil {
		event := &character.CharacterRestoredEvent{
			BaseEvent: character.BaseEvent{
				EventType:   character.EventCharacterRestored,
				CharacterID: characterID,
				UserID:      userID,
			},
			Name: char.Name,
		}
		
		if err := s.eventPublisher.PublishCharacterRestored(ctx, event); err != nil {
			s.logger.WithError(err).Warn("Failed to publish character restored event")
		}
	}

	return nil
}

// GetAppearance retrieves character appearance
func (s *CharacterService) GetAppearance(ctx context.Context, characterID string) (*character.Appearance, error) {
	charID, err := uuid.Parse(characterID)
	if err != nil {
		return nil, character.ErrInvalidCharacterID
	}

	// Try to get from cache first
	if s.cache != nil {
		appearance, err := s.cache.GetAppearance(ctx, charID)
		if err == nil && appearance != nil {
			return appearance, nil
		}
		// Log cache miss but continue
		if err != nil {
			s.logger.WithError(err).Warn("Failed to get appearance from cache")
		}
	}

	// Verify character exists and is not deleted
	char, err := s.characterRepo.GetByID(ctx, charID)
	if err != nil {
		return nil, character.ErrCharacterNotFound
	}
	if char.IsDeleted {
		return nil, character.ErrCharacterDeleted
	}

	appearance, err := s.appearanceRepo.GetByCharacterID(ctx, charID)
	if err != nil {
		return nil, character.ErrAppearanceNotFound
	}

	// Cache the appearance
	if s.cache != nil {
		if err := s.cache.SetAppearance(ctx, appearance, s.cacheTTL.Appearance); err != nil {
			s.logger.WithError(err).Warn("Failed to cache appearance")
		}
	}

	return appearance, nil
}

// UpdateAppearance updates character appearance
func (s *CharacterService) UpdateAppearance(ctx context.Context, characterID string, req *portsCharacter.UpdateAppearanceRequest) (*character.Appearance, error) {
	charID, err := uuid.Parse(characterID)
	if err != nil {
		return nil, character.ErrInvalidCharacterID
	}

	// Get existing appearance
	appearance, err := s.appearanceRepo.GetByCharacterID(ctx, charID)
	if err != nil {
		return nil, character.ErrAppearanceNotFound
	}

	// Apply updates
	if req.FaceType != nil {
		appearance.FaceType = *req.FaceType
	}
	if req.SkinColor != nil {
		appearance.SkinColor = *req.SkinColor
	}
	if req.EyeColor != nil {
		appearance.EyeColor = *req.EyeColor
	}
	if req.HairStyle != nil {
		appearance.HairStyle = *req.HairStyle
	}
	if req.HairColor != nil {
		appearance.HairColor = *req.HairColor
	}
	if req.FacialHairStyle != nil {
		appearance.FacialHairStyle = *req.FacialHairStyle
	}
	if req.FacialHairColor != nil {
		appearance.FacialHairColor = *req.FacialHairColor
	}
	if req.BodyType != nil {
		appearance.BodyType = *req.BodyType
	}
	if req.Height != nil {
		appearance.Height = *req.Height
	}
	if req.BodyProportions != nil {
		appearance.BodyProportions = *req.BodyProportions
	}
	if req.Scars != nil {
		appearance.Scars = req.Scars
	}
	if req.Tattoos != nil {
		appearance.Tattoos = req.Tattoos
	}
	if req.Accessories != nil {
		appearance.Accessories = req.Accessories
	}

	// Validate
	if err := appearance.Validate(); err != nil {
		return nil, err
	}

	appearance.UpdatedAt = time.Now()

	// Update in database
	if err := s.appearanceRepo.Update(ctx, appearance); err != nil {
		return nil, fmt.Errorf("failed to update appearance: %w", err)
	}

	// Invalidate appearance cache
	if s.cache != nil {
		if err := s.cache.DeleteAppearance(ctx, charID); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate appearance cache after update")
		}
	}

	// Publish appearance updated event
	if s.eventPublisher != nil {
		// Get character info for event
		char, _ := s.characterRepo.GetByID(ctx, charID)
		if char != nil {
			// Track which fields changed
			changedFields := []string{}
			if req.FaceType != nil { changedFields = append(changedFields, "face_type") }
			if req.SkinColor != nil { changedFields = append(changedFields, "skin_color") }
			if req.EyeColor != nil { changedFields = append(changedFields, "eye_color") }
			if req.HairStyle != nil { changedFields = append(changedFields, "hair_style") }
			if req.HairColor != nil { changedFields = append(changedFields, "hair_color") }
			if req.FacialHairStyle != nil { changedFields = append(changedFields, "facial_hair_style") }
			if req.FacialHairColor != nil { changedFields = append(changedFields, "facial_hair_color") }
			if req.BodyType != nil { changedFields = append(changedFields, "body_type") }
			if req.Height != nil { changedFields = append(changedFields, "height") }
			if req.BodyProportions != nil { changedFields = append(changedFields, "body_proportions") }
			if req.Scars != nil { changedFields = append(changedFields, "scars") }
			if req.Tattoos != nil { changedFields = append(changedFields, "tattoos") }
			if req.Accessories != nil { changedFields = append(changedFields, "accessories") }
			
			event := &character.CharacterAppearanceUpdatedEvent{
				BaseEvent: character.BaseEvent{
					EventType:   character.EventCharacterAppearanceUpdated,
					CharacterID: characterID,
					UserID:      char.UserID.String(),
				},
				ChangedFields: changedFields,
			}
			
			if err := s.eventPublisher.PublishCharacterAppearanceUpdated(ctx, event); err != nil {
				s.logger.WithError(err).Warn("Failed to publish appearance updated event")
			}
		}
	}

	return appearance, nil
}

// GetStats retrieves character stats
func (s *CharacterService) GetStats(ctx context.Context, characterID string) (*character.Stats, error) {
	charID, err := uuid.Parse(characterID)
	if err != nil {
		return nil, character.ErrInvalidCharacterID
	}

	// Try to get from cache first
	if s.cache != nil {
		stats, err := s.cache.GetStats(ctx, charID)
		if err == nil && stats != nil {
			return stats, nil
		}
		// Log cache miss but continue
		if err != nil {
			s.logger.WithError(err).Warn("Failed to get stats from cache")
		}
	}

	// Verify character exists and is not deleted
	char, err := s.characterRepo.GetByID(ctx, charID)
	if err != nil {
		return nil, character.ErrCharacterNotFound
	}
	if char.IsDeleted {
		return nil, character.ErrCharacterDeleted
	}

	stats, err := s.statsRepo.GetByCharacterID(ctx, charID)
	if err != nil {
		return nil, character.ErrStatsNotFound
	}

	// Cache the stats
	if s.cache != nil {
		if err := s.cache.SetStats(ctx, stats, s.cacheTTL.Stats); err != nil {
			s.logger.WithError(err).Warn("Failed to cache stats")
		}
	}

	return stats, nil
}

// AllocateStatPoint allocates a stat point to a primary stat
func (s *CharacterService) AllocateStatPoint(ctx context.Context, characterID string, stat string) (*character.Stats, error) {
	charID, err := uuid.Parse(characterID)
	if err != nil {
		return nil, character.ErrInvalidCharacterID
	}

	// Get character to verify it exists and get class type
	char, err := s.characterRepo.GetByID(ctx, charID)
	if err != nil {
		return nil, character.ErrCharacterNotFound
	}
	if char.IsDeleted {
		return nil, character.ErrCharacterDeleted
	}

	// Get stats
	stats, err := s.statsRepo.GetByCharacterID(ctx, charID)
	if err != nil {
		return nil, character.ErrStatsNotFound
	}

	// Allocate point
	if err := stats.AllocateStatPoint(stat); err != nil {
		return nil, err
	}

	// Recalculate derived stats
	stats.CalculateDerivedStats(char.ClassType)

	// Update in database
	if err := s.statsRepo.Update(ctx, stats); err != nil {
		return nil, fmt.Errorf("failed to update stats: %w", err)
	}

	// Invalidate stats cache
	if s.cache != nil {
		if err := s.cache.DeleteStats(ctx, charID); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate stats cache after update")
		}
	}

	// Publish stats updated event
	if s.eventPublisher != nil {
		// Create a map of previous stats (before allocation)
		previousStats := map[string]int{
			"strength":     stats.Strength - 1, // We allocated 1 point
			"dexterity":    stats.Dexterity,
			"intelligence": stats.Intelligence,
			"wisdom":       stats.Wisdom,
			"constitution": stats.Constitution,
		}
		// Adjust the previous stat that was changed
		if stat == "strength" { previousStats["strength"] = stats.Strength - 1 }
		if stat == "dexterity" { previousStats["dexterity"] = stats.Dexterity - 1 }
		if stat == "intelligence" { previousStats["intelligence"] = stats.Intelligence - 1 }
		if stat == "wisdom" { previousStats["wisdom"] = stats.Wisdom - 1 }
		if stat == "constitution" { previousStats["constitution"] = stats.Constitution - 1 }
		
		newStats := map[string]int{
			"strength":     stats.Strength,
			"dexterity":    stats.Dexterity,
			"intelligence": stats.Intelligence,
			"wisdom":       stats.Wisdom,
			"constitution": stats.Constitution,
		}
		
		changes := map[string]int{stat: 1}
		
		event := &character.CharacterStatsUpdatedEvent{
			BaseEvent: character.BaseEvent{
				EventType:   character.EventCharacterStatsUpdated,
				CharacterID: characterID,
				UserID:      char.UserID.String(),
			},
			UpdateType:    "stat_allocation",
			PreviousStats: previousStats,
			NewStats:      newStats,
			Changes:       changes,
		}
		
		if err := s.eventPublisher.PublishCharacterStatsUpdated(ctx, event); err != nil {
			s.logger.WithError(err).Warn("Failed to publish stats updated event")
		}
	}

	return stats, nil
}

// GetPosition retrieves character position
func (s *CharacterService) GetPosition(ctx context.Context, characterID string) (*character.Position, error) {
	charID, err := uuid.Parse(characterID)
	if err != nil {
		return nil, character.ErrInvalidCharacterID
	}

	// Try to get from cache first
	if s.cache != nil {
		position, err := s.cache.GetPosition(ctx, charID)
		if err == nil && position != nil {
			return position, nil
		}
		// Log cache miss but continue
		if err != nil {
			s.logger.WithError(err).Warn("Failed to get position from cache")
		}
	}

	// Verify character exists and is not deleted
	char, err := s.characterRepo.GetByID(ctx, charID)
	if err != nil {
		return nil, character.ErrCharacterNotFound
	}
	if char.IsDeleted {
		return nil, character.ErrCharacterDeleted
	}

	position, err := s.positionRepo.GetByCharacterID(ctx, charID)
	if err != nil {
		return nil, character.ErrPositionNotFound
	}

	// Cache the position with a shorter TTL for real-time updates
	if s.cache != nil {
		if err := s.cache.SetPosition(ctx, position, s.cacheTTL.Position); err != nil {
			s.logger.WithError(err).Warn("Failed to cache position")
		}
	}

	return position, nil
}

// UpdatePosition updates character position
func (s *CharacterService) UpdatePosition(ctx context.Context, characterID string, req *portsCharacter.UpdatePositionRequest) (*character.Position, error) {
	charID, err := uuid.Parse(characterID)
	if err != nil {
		return nil, character.ErrInvalidCharacterID
	}

	// Get existing position
	position, err := s.positionRepo.GetByCharacterID(ctx, charID)
	if err != nil {
		return nil, character.ErrPositionNotFound
	}

	// Apply updates
	if req.WorldID != nil {
		position.WorldID = *req.WorldID
	}
	if req.ZoneID != nil {
		position.ZoneID = *req.ZoneID
	}
	if req.MapID != nil {
		position.MapID = *req.MapID
	}
	if req.PositionX != nil {
		position.PositionX = *req.PositionX
	}
	if req.PositionY != nil {
		position.PositionY = *req.PositionY
	}
	if req.PositionZ != nil {
		position.PositionZ = *req.PositionZ
	}
	if req.RotationPitch != nil {
		position.RotationPitch = *req.RotationPitch
	}
	if req.RotationYaw != nil {
		position.RotationYaw = *req.RotationYaw
	}
	if req.RotationRoll != nil {
		position.RotationRoll = *req.RotationRoll
	}
	if req.VelocityX != nil {
		position.VelocityX = *req.VelocityX
	}
	if req.VelocityY != nil {
		position.VelocityY = *req.VelocityY
	}
	if req.VelocityZ != nil {
		position.VelocityZ = *req.VelocityZ
	}

	// Validate
	if err := position.Validate(); err != nil {
		return nil, err
	}

	position.LastMovement = time.Now()
	position.UpdatedAt = time.Now()

	// Update in database
	if err := s.positionRepo.Update(ctx, position); err != nil {
		return nil, fmt.Errorf("failed to update position: %w", err)
	}

	// Update cache with the new position
	if s.cache != nil {
		if err := s.cache.SetPosition(ctx, position, s.cacheTTL.Position); err != nil {
			s.logger.WithError(err).Warn("Failed to update position cache")
		}
	}

	// Publish position updated event
	if s.eventPublisher != nil {
		// Get character info for event
		char, _ := s.characterRepo.GetByID(ctx, charID)
		if char != nil {
			// Get previous position for comparison (if needed)
			previousPos, _ := s.positionRepo.GetByCharacterID(ctx, charID)
			
			event := &character.CharacterPositionUpdatedEvent{
				BaseEvent: character.BaseEvent{
					EventType:   character.EventCharacterPositionUpdated,
					CharacterID: characterID,
					UserID:      char.UserID.String(),
				},
				PreviousPosition: previousPos,
				NewPosition:      position,
				MovementType:     "walk", // This could be enhanced based on velocity/distance
			}
			
			if err := s.eventPublisher.PublishCharacterPositionUpdated(ctx, event); err != nil {
				s.logger.WithError(err).Warn("Failed to publish position updated event")
			}
		}
	}

	return position, nil
}

// TeleportToSafePosition teleports character to their safe position
func (s *CharacterService) TeleportToSafePosition(ctx context.Context, characterID string) (*character.Position, error) {
	charID, err := uuid.Parse(characterID)
	if err != nil {
		return nil, character.ErrInvalidCharacterID
	}

	if err := s.positionRepo.TeleportToSafePosition(ctx, charID); err != nil {
		return nil, fmt.Errorf("failed to teleport: %w", err)
	}

	// Get updated position
	position, err := s.positionRepo.GetByCharacterID(ctx, charID)
	if err != nil {
		return nil, character.ErrPositionNotFound
	}

	// Publish position updated event (teleport)
	if s.eventPublisher != nil {
		// Get character info for event
		char, _ := s.characterRepo.GetByID(ctx, charID)
		if char != nil {
			event := &character.CharacterPositionUpdatedEvent{
				BaseEvent: character.BaseEvent{
					EventType:   character.EventCharacterPositionUpdated,
					CharacterID: characterID,
					UserID:      char.UserID.String(),
				},
				NewPosition:  position,
				MovementType: "teleport",
			}
			
			if err := s.eventPublisher.PublishCharacterPositionUpdated(ctx, event); err != nil {
				s.logger.WithError(err).Warn("Failed to publish position updated event")
			}
		}
	}

	return position, nil
}

// ValidateCharacterOwnership validates that a character belongs to a user
func (s *CharacterService) ValidateCharacterOwnership(ctx context.Context, characterID string, userID string) error {
	charID, err := uuid.Parse(characterID)
	if err != nil {
		return character.ErrInvalidCharacterID
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return character.ErrInvalidUserID
	}

	char, err := s.characterRepo.GetByID(ctx, charID)
	if err != nil {
		return character.ErrCharacterNotFound
	}

	if char.UserID != uid {
		return character.ErrCharacterBelongsToOther
	}

	return nil
}

// CanCreateCharacter checks if a user can create more characters
func (s *CharacterService) CanCreateCharacter(ctx context.Context, userID string) (bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, character.ErrInvalidUserID
	}

	var count int
	
	// Try to get count from cache first
	if s.cache != nil {
		cachedCount, found, err := s.cache.GetCharacterCount(ctx, uid)
		if err == nil && found {
			return cachedCount < s.config.MaxCharactersPerUser, nil
		}
		// Log cache miss but continue
		if err != nil {
			s.logger.WithError(err).Warn("Failed to get character count from cache")
		}
	}

	// Get from repository
	count, err = s.characterRepo.CountByUserID(ctx, uid)
	if err != nil {
		return false, fmt.Errorf("failed to count characters: %w", err)
	}

	// Cache the count
	if s.cache != nil {
		if err := s.cache.SetCharacterCount(ctx, uid, count, s.cacheTTL.CharacterCount); err != nil {
			s.logger.WithError(err).Warn("Failed to cache character count")
		}
	}

	return count < s.config.MaxCharactersPerUser, nil
}

// SelectCharacter selects a character for gameplay
func (s *CharacterService) SelectCharacter(ctx context.Context, characterID string, userID string, sessionID string) error {
	// Validate ownership
	if err := s.ValidateCharacterOwnership(ctx, characterID, userID); err != nil {
		return err
	}

	charID, _ := uuid.Parse(characterID)
	
	// Get character details
	char, err := s.characterRepo.GetByID(ctx, charID)
	if err != nil {
		return character.ErrCharacterNotFound
	}
	
	if char.IsDeleted {
		return character.ErrCharacterDeleted
	}

	// Update last selected time
	char.LastSelectedAt = time.Now()
	if err := s.characterRepo.Update(ctx, char); err != nil {
		s.logger.WithError(err).Warn("Failed to update last selected time")
	}

	// Publish character selected event
	if s.eventPublisher != nil {
		event := &character.CharacterSelectedEvent{
			BaseEvent: character.BaseEvent{
				EventType:   character.EventCharacterSelected,
				CharacterID: characterID,
				UserID:      userID,
			},
			Name:      char.Name,
			SessionID: sessionID,
		}
		
		if err := s.eventPublisher.PublishCharacterSelected(ctx, event); err != nil {
			s.logger.WithError(err).Warn("Failed to publish character selected event")
		}
		
		// Also publish character online event
		position, _ := s.positionRepo.GetByCharacterID(ctx, charID)
		if position != nil {
			onlineEvent := &character.CharacterOnlineEvent{
				BaseEvent: character.BaseEvent{
					EventType:   character.EventCharacterOnline,
					CharacterID: characterID,
					UserID:      userID,
				},
				Name:      char.Name,
				SessionID: sessionID,
				WorldID:   position.WorldID,
				ZoneID:    position.ZoneID,
			}
			
			if err := s.eventPublisher.PublishCharacterOnline(ctx, onlineEvent); err != nil {
				s.logger.WithError(err).Warn("Failed to publish character online event")
			}
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"character_id": characterID,
		"user_id":      userID,
		"session_id":   sessionID,
		"character_name": char.Name,
	}).Info("Character selected for gameplay")

	return nil
}

// validateCharacterName validates a character name
func (s *CharacterService) validateCharacterName(name string) error {
	// Check length
	if len(name) < s.config.MinCharacterNameLength || len(name) > s.config.MaxCharacterNameLength {
		return character.ErrInvalidCharacterName
	}

	// Check for valid characters (letters, numbers, spaces, hyphens)
	validName := regexp.MustCompile(`^[a-zA-Z0-9\s\-]+$`)
	if !validName.MatchString(name) {
		return character.ErrInvalidCharacterName
	}

	// Check for offensive words (simplified - in production would use a more comprehensive filter)
	offensiveWords := []string{"admin", "gm", "gamemaster", "moderator"}
	nameLower := strings.ToLower(name)
	for _, word := range offensiveWords {
		if strings.Contains(nameLower, word) {
			return character.ErrInvalidCharacterName
		}
	}

	return nil
}

// applyAppearanceOptions applies custom appearance options to an appearance entity
func (s *CharacterService) applyAppearanceOptions(appearance *character.Appearance, options *portsCharacter.CharacterAppearanceOptions) {
	if options.FaceType != nil {
		appearance.FaceType = *options.FaceType
	}
	if options.SkinColor != nil {
		appearance.SkinColor = *options.SkinColor
	}
	if options.EyeColor != nil {
		appearance.EyeColor = *options.EyeColor
	}
	if options.HairStyle != nil {
		appearance.HairStyle = *options.HairStyle
	}
	if options.HairColor != nil {
		appearance.HairColor = *options.HairColor
	}
	if options.FacialHairStyle != nil {
		appearance.FacialHairStyle = *options.FacialHairStyle
	}
	if options.FacialHairColor != nil {
		appearance.FacialHairColor = *options.FacialHairColor
	}
	if options.BodyType != nil {
		appearance.BodyType = *options.BodyType
	}
	if options.Height != nil {
		appearance.Height = *options.Height
	}
}


// GetSelectedCharacter retrieves the currently selected character for a user
func (s *CharacterService) GetSelectedCharacter(ctx context.Context, userID string) (*character.Character, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, character.ErrInvalidUserID
	}

	// Get selected character ID from cache
	if s.cache != nil {
		charID, err := s.cache.GetSelectedCharacter(ctx, uid)
		if err == nil && charID != uuid.Nil {
			// Get the character data
			return s.GetCharacter(ctx, charID.String())
		}
	}

	return nil, character.ErrNoCharacterSelected
}

// DeselectCharacter removes the character selection for a user
func (s *CharacterService) DeselectCharacter(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return character.ErrInvalidUserID
	}

	// Remove selected character from cache
	if s.cache != nil {
		if err := s.cache.DeleteSelectedCharacter(ctx, uid); err != nil {
			return fmt.Errorf("failed to deselect character: %w", err)
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("Character deselected")

	return nil
}