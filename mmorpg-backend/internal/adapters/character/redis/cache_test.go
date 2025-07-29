package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/character"
	portsCharacter "github.com/mmorpg-template/backend/internal/ports/character"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRedisClient is a simple mock for testing
type MockRedisClient struct {
	data map[string]interface{}
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string]interface{}),
	}
}

func TestRedisCharacterCache_Character(t *testing.T) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // This would be mocked in real tests
	})
	
	cache := NewRedisCharacterCache(client, "test", nil)
	
	// Create test character
	char := &character.Character{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		Name:       "TestHero",
		SlotNumber: 1,
		Level:      10,
		Experience: 5000,
		ClassType:  character.ClassWarrior,
		Race:       character.RaceHuman,
		Gender:     character.GenderMale,
		IsDeleted:  false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	// Test Set and Get
	err := cache.SetCharacter(ctx, char, 5*time.Minute)
	assert.NoError(t, err)
	
	retrieved, err := cache.GetCharacter(ctx, char.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, char.Name, retrieved.Name)
	assert.Equal(t, char.Level, retrieved.Level)
	
	// Test Delete
	err = cache.DeleteCharacter(ctx, char.ID)
	assert.NoError(t, err)
	
	// Verify deleted
	retrieved, err = cache.GetCharacter(ctx, char.ID)
	assert.NoError(t, err)
	assert.Nil(t, retrieved) // Should be cache miss
}

func TestRedisCharacterCache_UserCharacters(t *testing.T) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	cache := NewRedisCharacterCache(client, "test", nil)
	userID := uuid.New()
	
	// Create test characters
	chars := []*character.Character{
		{
			ID:         uuid.New(),
			UserID:     userID,
			Name:       "Hero1",
			SlotNumber: 1,
			Level:      10,
			ClassType:  character.ClassWarrior,
			Race:       character.RaceHuman,
			Gender:     character.GenderMale,
		},
		{
			ID:         uuid.New(),
			UserID:     userID,
			Name:       "Hero2",
			SlotNumber: 2,
			Level:      5,
			ClassType:  character.ClassMage,
			Race:       character.RaceElf,
			Gender:     character.GenderFemale,
		},
	}
	
	// Test Set and Get
	err := cache.SetUserCharacters(ctx, userID, chars, 5*time.Minute)
	assert.NoError(t, err)
	
	retrieved, err := cache.GetUserCharacters(ctx, userID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Len(t, retrieved, 2)
	assert.Equal(t, "Hero1", retrieved[0].Name)
	assert.Equal(t, "Hero2", retrieved[1].Name)
}

func TestRedisCharacterCache_CharacterCount(t *testing.T) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	cache := NewRedisCharacterCache(client, "test", nil)
	userID := uuid.New()
	
	// Test Set and Get
	err := cache.SetCharacterCount(ctx, userID, 3, 1*time.Minute)
	assert.NoError(t, err)
	
	count, found, err := cache.GetCharacterCount(ctx, userID)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, 3, count)
	
	// Test Delete
	err = cache.DeleteCharacterCount(ctx, userID)
	assert.NoError(t, err)
	
	// Verify deleted
	count, found, err = cache.GetCharacterCount(ctx, userID)
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Equal(t, 0, count)
}

func TestRedisCharacterCache_SelectedCharacter(t *testing.T) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	cache := NewRedisCharacterCache(client, "test", nil)
	userID := uuid.New()
	characterID := uuid.New()
	
	// Test Set and Get
	err := cache.SetSelectedCharacter(ctx, userID, characterID, 24*time.Hour)
	assert.NoError(t, err)
	
	retrieved, err := cache.GetSelectedCharacter(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, characterID, retrieved)
	
	// Test Delete
	err = cache.DeleteSelectedCharacter(ctx, userID)
	assert.NoError(t, err)
	
	// Verify deleted
	retrieved, err = cache.GetSelectedCharacter(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, uuid.Nil, retrieved)
}

func TestRedisCharacterCache_Appearance(t *testing.T) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	cache := NewRedisCharacterCache(client, "test", nil)
	
	// Create test appearance
	appearance := &character.Appearance{
		CharacterID:     uuid.New(),
		FaceType:        1,
		SkinColor:       "fair",
		EyeColor:        "blue",
		HairStyle:       2,
		HairColor:       "blonde",
		FacialHairStyle: 0,
		FacialHairColor: "",
		BodyType:        "athletic",
		Height:          180,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	
	// Test Set and Get
	err := cache.SetAppearance(ctx, appearance, 10*time.Minute)
	assert.NoError(t, err)
	
	retrieved, err := cache.GetAppearance(ctx, appearance.CharacterID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, appearance.EyeColor, retrieved.EyeColor)
	assert.Equal(t, appearance.Height, retrieved.Height)
}

func TestRedisCharacterCache_Stats(t *testing.T) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	cache := NewRedisCharacterCache(client, "test", nil)
	
	// Create test stats
	stats := &character.Stats{
		CharacterID: uuid.New(),
		Health:      100,
		MaxHealth:   100,
		Mana:        50,
		MaxMana:     50,
		Stamina:     100,
		MaxStamina:  100,
		Strength:    20,
		Agility:     15,
		Intelligence:10,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Test Set and Get
	err := cache.SetStats(ctx, stats, 10*time.Minute)
	assert.NoError(t, err)
	
	retrieved, err := cache.GetStats(ctx, stats.CharacterID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, stats.Health, retrieved.Health)
	assert.Equal(t, stats.Strength, retrieved.Strength)
}

func TestRedisCharacterCache_InvalidateData(t *testing.T) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	cache := NewRedisCharacterCache(client, "test", nil)
	userID := uuid.New()
	characterID := uuid.New()
	
	// Set various cached data
	char := &character.Character{
		ID:     characterID,
		UserID: userID,
		Name:   "TestHero",
	}
	err := cache.SetCharacter(ctx, char, 10*time.Minute)
	require.NoError(t, err)
	
	err = cache.SetUserCharacters(ctx, userID, []*character.Character{char}, 5*time.Minute)
	require.NoError(t, err)
	
	err = cache.SetCharacterCount(ctx, userID, 1, 1*time.Minute)
	require.NoError(t, err)
	
	// Test InvalidateCharacterData
	err = cache.InvalidateCharacterData(ctx, characterID)
	assert.NoError(t, err)
	
	// Verify character data is gone
	retrieved, err := cache.GetCharacter(ctx, characterID)
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
	
	// Test InvalidateUserData
	err = cache.InvalidateUserData(ctx, userID)
	assert.NoError(t, err)
	
	// Verify user data is gone
	chars, err := cache.GetUserCharacters(ctx, userID)
	assert.NoError(t, err)
	assert.Nil(t, chars)
	
	count, found, err := cache.GetCharacterCount(ctx, userID)
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Equal(t, 0, count)
}

func TestRedisCharacterCache_TTL(t *testing.T) {
	ttl := portsCharacter.DefaultCacheTTL()
	
	assert.Equal(t, 10*time.Minute, ttl.Character)
	assert.Equal(t, 5*time.Minute, ttl.CharacterList)
	assert.Equal(t, 1*time.Minute, ttl.CharacterCount)
	assert.Equal(t, 24*time.Hour, ttl.SelectedCharacter)
	assert.Equal(t, 10*time.Minute, ttl.Appearance)
	assert.Equal(t, 10*time.Minute, ttl.Stats)
	assert.Equal(t, 30*time.Second, ttl.Position)
}