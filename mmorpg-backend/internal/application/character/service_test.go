package character_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mmorpg-template/backend/internal/application/character"
	domainCharacter "github.com/mmorpg-template/backend/internal/domain/character"
	portsCharacter "github.com/mmorpg-template/backend/internal/ports/character"
	"github.com/mmorpg-template/backend/pkg/logger"
)

// Mock repositories
type MockCharacterRepo struct {
	mock.Mock
}

func (m *MockCharacterRepo) Create(ctx context.Context, char *domainCharacter.Character) error {
	args := m.Called(ctx, char)
	return args.Error(0)
}

func (m *MockCharacterRepo) GetByID(ctx context.Context, id uuid.UUID) (*domainCharacter.Character, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainCharacter.Character), args.Error(1)
}

func (m *MockCharacterRepo) GetByName(ctx context.Context, name string) (*domainCharacter.Character, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainCharacter.Character), args.Error(1)
}

func (m *MockCharacterRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domainCharacter.Character, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainCharacter.Character), args.Error(1)
}

func (m *MockCharacterRepo) GetByUserIDAndSlot(ctx context.Context, userID uuid.UUID, slot int) (*domainCharacter.Character, error) {
	args := m.Called(ctx, userID, slot)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainCharacter.Character), args.Error(1)
}

func (m *MockCharacterRepo) Update(ctx context.Context, char *domainCharacter.Character) error {
	args := m.Called(ctx, char)
	return args.Error(0)
}

func (m *MockCharacterRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCharacterRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCharacterRepo) Restore(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCharacterRepo) CleanupDeleted(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCharacterRepo) NameExists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockCharacterRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

// Mock appearance repository
type MockAppearanceRepo struct {
	mock.Mock
}

func (m *MockAppearanceRepo) Create(ctx context.Context, appearance *domainCharacter.Appearance) error {
	args := m.Called(ctx, appearance)
	return args.Error(0)
}

func (m *MockAppearanceRepo) GetByCharacterID(ctx context.Context, characterID uuid.UUID) (*domainCharacter.Appearance, error) {
	args := m.Called(ctx, characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainCharacter.Appearance), args.Error(1)
}

func (m *MockAppearanceRepo) Update(ctx context.Context, appearance *domainCharacter.Appearance) error {
	args := m.Called(ctx, appearance)
	return args.Error(0)
}

func (m *MockAppearanceRepo) Delete(ctx context.Context, characterID uuid.UUID) error {
	args := m.Called(ctx, characterID)
	return args.Error(0)
}

// Mock stats repository
type MockStatsRepo struct {
	mock.Mock
}

func (m *MockStatsRepo) Create(ctx context.Context, stats *domainCharacter.Stats) error {
	args := m.Called(ctx, stats)
	return args.Error(0)
}

func (m *MockStatsRepo) GetByCharacterID(ctx context.Context, characterID uuid.UUID) (*domainCharacter.Stats, error) {
	args := m.Called(ctx, characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainCharacter.Stats), args.Error(1)
}

func (m *MockStatsRepo) Update(ctx context.Context, stats *domainCharacter.Stats) error {
	args := m.Called(ctx, stats)
	return args.Error(0)
}

func (m *MockStatsRepo) Delete(ctx context.Context, characterID uuid.UUID) error {
	args := m.Called(ctx, characterID)
	return args.Error(0)
}

func (m *MockStatsRepo) GetMultipleByCharacterIDs(ctx context.Context, characterIDs []uuid.UUID) (map[uuid.UUID]*domainCharacter.Stats, error) {
	args := m.Called(ctx, characterIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]*domainCharacter.Stats), args.Error(1)
}

// Mock position repository
type MockPositionRepo struct {
	mock.Mock
}

func (m *MockPositionRepo) Create(ctx context.Context, position *domainCharacter.Position) error {
	args := m.Called(ctx, position)
	return args.Error(0)
}

func (m *MockPositionRepo) GetByCharacterID(ctx context.Context, characterID uuid.UUID) (*domainCharacter.Position, error) {
	args := m.Called(ctx, characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainCharacter.Position), args.Error(1)
}

func (m *MockPositionRepo) Update(ctx context.Context, position *domainCharacter.Position) error {
	args := m.Called(ctx, position)
	return args.Error(0)
}

func (m *MockPositionRepo) Delete(ctx context.Context, characterID uuid.UUID) error {
	args := m.Called(ctx, characterID)
	return args.Error(0)
}

func (m *MockPositionRepo) FindNearbyCharacters(ctx context.Context, characterID uuid.UUID, maxDistance float64) ([]*portsCharacter.NearbyCharacter, error) {
	args := m.Called(ctx, characterID, maxDistance)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*portsCharacter.NearbyCharacter), args.Error(1)
}

func (m *MockPositionRepo) GetCharactersInZone(ctx context.Context, worldID, zoneID string) ([]*domainCharacter.Position, error) {
	args := m.Called(ctx, worldID, zoneID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainCharacter.Position), args.Error(1)
}

func (m *MockPositionRepo) GetCharactersInInstance(ctx context.Context, instanceID uuid.UUID) ([]*domainCharacter.Position, error) {
	args := m.Called(ctx, instanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainCharacter.Position), args.Error(1)
}

func (m *MockPositionRepo) SaveSafePosition(ctx context.Context, characterID uuid.UUID) error {
	args := m.Called(ctx, characterID)
	return args.Error(0)
}

func (m *MockPositionRepo) TeleportToSafePosition(ctx context.Context, characterID uuid.UUID) error {
	args := m.Called(ctx, characterID)
	return args.Error(0)
}

// Test cases

func TestCharacterService_CreateCharacter(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	
	// Setup
	charRepo := new(MockCharacterRepo)
	appearanceRepo := new(MockAppearanceRepo)
	statsRepo := new(MockStatsRepo)
	positionRepo := new(MockPositionRepo)
	
	config := &character.Config{
		MaxCharactersPerUser:      5,
		MaxCharacterNameLength:    30,
		MinCharacterNameLength:    3,
		DefaultStartingLevel:      1,
		DefaultStartingExperience: 0,
	}
	
	log := logger.NewNoop()
	
	service := character.NewCharacterService(
		charRepo,
		appearanceRepo,
		statsRepo,
		positionRepo,
		config,
		log,
	)
	
	t.Run("successful character creation", func(t *testing.T) {
		// Setup mocks
		charRepo.On("CountByUserID", ctx, userID).Return(2, nil).Once()
		charRepo.On("NameExists", ctx, "TestHero").Return(false, nil).Once()
		charRepo.On("GetByUserIDAndSlot", ctx, userID, 1).Return(nil, sql.ErrNoRows).Once()
		charRepo.On("Create", ctx, mock.AnythingOfType("*character.Character")).Return(nil).Once()
		appearanceRepo.On("Create", ctx, mock.AnythingOfType("*character.Appearance")).Return(nil).Once()
		statsRepo.On("Create", ctx, mock.AnythingOfType("*character.Stats")).Return(nil).Once()
		positionRepo.On("Create", ctx, mock.AnythingOfType("*character.Position")).Return(nil).Once()
		
		req := &portsCharacter.CreateCharacterRequest{
			UserID:     userID.String(),
			Name:       "TestHero",
			SlotNumber: 1,
			ClassType:  domainCharacter.ClassWarrior,
			Race:       domainCharacter.RaceHuman,
			Gender:     domainCharacter.GenderMale,
		}
		
		char, err := service.CreateCharacter(ctx, req)
		
		require.NoError(t, err)
		assert.NotNil(t, char)
		assert.Equal(t, "TestHero", char.Name)
		assert.Equal(t, domainCharacter.ClassWarrior, char.ClassType)
		assert.Equal(t, domainCharacter.RaceHuman, char.Race)
		assert.Equal(t, 1, char.Level)
		assert.Equal(t, int64(0), char.Experience)
		
		// Verify all mocks were called
		charRepo.AssertExpectations(t)
		appearanceRepo.AssertExpectations(t)
		statsRepo.AssertExpectations(t)
		positionRepo.AssertExpectations(t)
	})
	
	t.Run("character limit reached", func(t *testing.T) {
		charRepo.On("CountByUserID", ctx, userID).Return(5, nil).Once()
		
		req := &portsCharacter.CreateCharacterRequest{
			UserID:     userID.String(),
			Name:       "TestHero2",
			SlotNumber: 2,
			ClassType:  domainCharacter.ClassMage,
			Race:       domainCharacter.RaceElf,
			Gender:     domainCharacter.GenderFemale,
		}
		
		_, err := service.CreateCharacter(ctx, req)
		
		assert.ErrorIs(t, err, domainCharacter.ErrCharacterLimitReached)
		charRepo.AssertExpectations(t)
	})
	
	t.Run("name already taken", func(t *testing.T) {
		charRepo.On("CountByUserID", ctx, userID).Return(2, nil).Once()
		charRepo.On("NameExists", ctx, "TakenName").Return(true, nil).Once()
		
		req := &portsCharacter.CreateCharacterRequest{
			UserID:     userID.String(),
			Name:       "TakenName",
			SlotNumber: 3,
			ClassType:  domainCharacter.ClassRogue,
			Race:       domainCharacter.RaceDwarf,
			Gender:     domainCharacter.GenderMale,
		}
		
		_, err := service.CreateCharacter(ctx, req)
		
		assert.ErrorIs(t, err, domainCharacter.ErrCharacterNameTaken)
		charRepo.AssertExpectations(t)
	})
}

func TestCharacterService_GetCharacter(t *testing.T) {
	ctx := context.Background()
	charID := uuid.New()
	userID := uuid.New()
	
	// Setup
	charRepo := new(MockCharacterRepo)
	appearanceRepo := new(MockAppearanceRepo)
	statsRepo := new(MockStatsRepo)
	positionRepo := new(MockPositionRepo)
	
	config := &character.Config{
		MaxCharactersPerUser:      5,
		MaxCharacterNameLength:    30,
		MinCharacterNameLength:    3,
		DefaultStartingLevel:      1,
		DefaultStartingExperience: 0,
	}
	
	log := logger.NewNoop()
	
	service := character.NewCharacterService(
		charRepo,
		appearanceRepo,
		statsRepo,
		positionRepo,
		config,
		log,
	)
	
	t.Run("successful retrieval", func(t *testing.T) {
		expectedChar := &domainCharacter.Character{
			ID:         charID,
			UserID:     userID,
			Name:       "TestHero",
			SlotNumber: 1,
			Level:      10,
			Experience: 5000,
			ClassType:  domainCharacter.ClassWarrior,
			Race:       domainCharacter.RaceHuman,
			Gender:     domainCharacter.GenderMale,
			IsDeleted:  false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		
		charRepo.On("GetByID", ctx, charID).Return(expectedChar, nil).Once()
		
		char, err := service.GetCharacter(ctx, charID.String())
		
		require.NoError(t, err)
		assert.Equal(t, expectedChar, char)
		charRepo.AssertExpectations(t)
	})
	
	t.Run("character not found", func(t *testing.T) {
		charRepo.On("GetByID", ctx, charID).Return(nil, domainCharacter.ErrCharacterNotFound).Once()
		
		_, err := service.GetCharacter(ctx, charID.String())
		
		assert.ErrorIs(t, err, domainCharacter.ErrCharacterNotFound)
		charRepo.AssertExpectations(t)
	})
	
	t.Run("deleted character", func(t *testing.T) {
		deletedChar := &domainCharacter.Character{
			ID:        charID,
			UserID:    userID,
			Name:      "DeletedHero",
			IsDeleted: true,
		}
		
		charRepo.On("GetByID", ctx, charID).Return(deletedChar, nil).Once()
		
		_, err := service.GetCharacter(ctx, charID.String())
		
		assert.ErrorIs(t, err, domainCharacter.ErrCharacterDeleted)
		charRepo.AssertExpectations(t)
	})
}

func TestCharacterService_DeleteAndRestore(t *testing.T) {
	ctx := context.Background()
	charID := uuid.New()
	userID := uuid.New()
	
	// Setup
	charRepo := new(MockCharacterRepo)
	appearanceRepo := new(MockAppearanceRepo)
	statsRepo := new(MockStatsRepo)
	positionRepo := new(MockPositionRepo)
	
	config := &character.Config{
		MaxCharactersPerUser:      5,
		MaxCharacterNameLength:    30,
		MinCharacterNameLength:    3,
		DefaultStartingLevel:      1,
		DefaultStartingExperience: 0,
	}
	
	log := logger.NewNoop()
	
	service := character.NewCharacterService(
		charRepo,
		appearanceRepo,
		statsRepo,
		positionRepo,
		config,
		log,
	)
	
	t.Run("successful soft delete", func(t *testing.T) {
		char := &domainCharacter.Character{
			ID:     charID,
			UserID: userID,
			Name:   "TestHero",
		}
		
		charRepo.On("GetByID", ctx, charID).Return(char, nil).Once()
		charRepo.On("SoftDelete", ctx, charID).Return(nil).Once()
		
		err := service.DeleteCharacter(ctx, charID.String(), userID.String())
		
		require.NoError(t, err)
		charRepo.AssertExpectations(t)
	})
	
	t.Run("successful restore", func(t *testing.T) {
		deletionTime := time.Now().Add(29 * 24 * time.Hour) // Within 30 day window
		char := &domainCharacter.Character{
			ID:                  charID,
			UserID:              userID,
			Name:                "TestHero",
			IsDeleted:           true,
			DeletionScheduledAt: &deletionTime,
		}
		
		charRepo.On("GetByID", ctx, charID).Return(char, nil).Once()
		charRepo.On("Restore", ctx, charID).Return(nil).Once()
		
		err := service.RestoreCharacter(ctx, charID.String(), userID.String())
		
		require.NoError(t, err)
		charRepo.AssertExpectations(t)
	})
}