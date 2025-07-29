package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/character"
	portsCharacter "github.com/mmorpg-template/backend/internal/ports/character"
	"github.com/redis/go-redis/v9"
)

// RedisCharacterCache implements CharacterCache using Redis
type RedisCharacterCache struct {
	client *redis.Client
	prefix string
	ttl    *portsCharacter.CacheTTL
}

// NewRedisCharacterCache creates a new Redis character cache
func NewRedisCharacterCache(client *redis.Client, prefix string, ttl *portsCharacter.CacheTTL) portsCharacter.CharacterCache {
	if ttl == nil {
		ttl = portsCharacter.DefaultCacheTTL()
	}
	return &RedisCharacterCache{
		client: client,
		prefix: prefix,
		ttl:    ttl,
	}
}

// SetCharacter caches a character
func (c *RedisCharacterCache) SetCharacter(ctx context.Context, char *character.Character, expiration time.Duration) error {
	if char == nil {
		return fmt.Errorf("character cannot be nil")
	}

	key := c.characterKey(char.ID)
	data, err := json.Marshal(char)
	if err != nil {
		return fmt.Errorf("failed to marshal character: %w", err)
	}

	err = c.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to cache character: %w", err)
	}

	return nil
}

// GetCharacter retrieves a cached character
func (c *RedisCharacterCache) GetCharacter(ctx context.Context, characterID uuid.UUID) (*character.Character, error) {
	key := c.characterKey(characterID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get cached character: %w", err)
	}

	var char character.Character
	err = json.Unmarshal(data, &char)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal character: %w", err)
	}

	return &char, nil
}

// DeleteCharacter removes a character from cache
func (c *RedisCharacterCache) DeleteCharacter(ctx context.Context, characterID uuid.UUID) error {
	key := c.characterKey(characterID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cached character: %w", err)
	}
	return nil
}

// SetUserCharacters caches a user's character list
func (c *RedisCharacterCache) SetUserCharacters(ctx context.Context, userID uuid.UUID, characters []*character.Character, expiration time.Duration) error {
	key := c.userCharactersKey(userID)
	data, err := json.Marshal(characters)
	if err != nil {
		return fmt.Errorf("failed to marshal character list: %w", err)
	}

	err = c.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to cache character list: %w", err)
	}

	return nil
}

// GetUserCharacters retrieves a cached character list
func (c *RedisCharacterCache) GetUserCharacters(ctx context.Context, userID uuid.UUID) ([]*character.Character, error) {
	key := c.userCharactersKey(userID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get cached character list: %w", err)
	}

	var characters []*character.Character
	err = json.Unmarshal(data, &characters)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal character list: %w", err)
	}

	return characters, nil
}

// DeleteUserCharacters removes a user's character list from cache
func (c *RedisCharacterCache) DeleteUserCharacters(ctx context.Context, userID uuid.UUID) error {
	key := c.userCharactersKey(userID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cached character list: %w", err)
	}
	return nil
}

// SetCharacterCount caches a user's character count
func (c *RedisCharacterCache) SetCharacterCount(ctx context.Context, userID uuid.UUID, count int, expiration time.Duration) error {
	key := c.characterCountKey(userID)
	err := c.client.Set(ctx, key, count, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to cache character count: %w", err)
	}
	return nil
}

// GetCharacterCount retrieves a cached character count
func (c *RedisCharacterCache) GetCharacterCount(ctx context.Context, userID uuid.UUID) (int, bool, error) {
	key := c.characterCountKey(userID)
	count, err := c.client.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, false, nil // Cache miss
		}
		return 0, false, fmt.Errorf("failed to get cached character count: %w", err)
	}
	return count, true, nil
}

// DeleteCharacterCount removes a user's character count from cache
func (c *RedisCharacterCache) DeleteCharacterCount(ctx context.Context, userID uuid.UUID) error {
	key := c.characterCountKey(userID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cached character count: %w", err)
	}
	return nil
}

// SetSelectedCharacter caches the selected character for a user
func (c *RedisCharacterCache) SetSelectedCharacter(ctx context.Context, userID uuid.UUID, characterID uuid.UUID, expiration time.Duration) error {
	key := c.selectedCharacterKey(userID)
	err := c.client.Set(ctx, key, characterID.String(), expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to cache selected character: %w", err)
	}
	return nil
}

// GetSelectedCharacter retrieves the cached selected character
func (c *RedisCharacterCache) GetSelectedCharacter(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	key := c.selectedCharacterKey(userID)
	charIDStr, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return uuid.Nil, nil // Cache miss
		}
		return uuid.Nil, fmt.Errorf("failed to get cached selected character: %w", err)
	}

	charID, err := uuid.Parse(charIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse character ID: %w", err)
	}

	return charID, nil
}

// DeleteSelectedCharacter removes the selected character from cache
func (c *RedisCharacterCache) DeleteSelectedCharacter(ctx context.Context, userID uuid.UUID) error {
	key := c.selectedCharacterKey(userID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cached selected character: %w", err)
	}
	return nil
}

// SetAppearance caches character appearance
func (c *RedisCharacterCache) SetAppearance(ctx context.Context, appearance *character.Appearance, expiration time.Duration) error {
	if appearance == nil {
		return fmt.Errorf("appearance cannot be nil")
	}

	key := c.appearanceKey(appearance.CharacterID)
	data, err := json.Marshal(appearance)
	if err != nil {
		return fmt.Errorf("failed to marshal appearance: %w", err)
	}

	err = c.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to cache appearance: %w", err)
	}

	return nil
}

// GetAppearance retrieves cached character appearance
func (c *RedisCharacterCache) GetAppearance(ctx context.Context, characterID uuid.UUID) (*character.Appearance, error) {
	key := c.appearanceKey(characterID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get cached appearance: %w", err)
	}

	var appearance character.Appearance
	err = json.Unmarshal(data, &appearance)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal appearance: %w", err)
	}

	return &appearance, nil
}

// DeleteAppearance removes character appearance from cache
func (c *RedisCharacterCache) DeleteAppearance(ctx context.Context, characterID uuid.UUID) error {
	key := c.appearanceKey(characterID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cached appearance: %w", err)
	}
	return nil
}

// SetStats caches character stats
func (c *RedisCharacterCache) SetStats(ctx context.Context, stats *character.Stats, expiration time.Duration) error {
	if stats == nil {
		return fmt.Errorf("stats cannot be nil")
	}

	key := c.statsKey(stats.CharacterID)
	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	err = c.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to cache stats: %w", err)
	}

	return nil
}

// GetStats retrieves cached character stats
func (c *RedisCharacterCache) GetStats(ctx context.Context, characterID uuid.UUID) (*character.Stats, error) {
	key := c.statsKey(characterID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get cached stats: %w", err)
	}

	var stats character.Stats
	err = json.Unmarshal(data, &stats)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	return &stats, nil
}

// DeleteStats removes character stats from cache
func (c *RedisCharacterCache) DeleteStats(ctx context.Context, characterID uuid.UUID) error {
	key := c.statsKey(characterID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cached stats: %w", err)
	}
	return nil
}

// SetPosition caches character position
func (c *RedisCharacterCache) SetPosition(ctx context.Context, position *character.Position, expiration time.Duration) error {
	if position == nil {
		return fmt.Errorf("position cannot be nil")
	}

	key := c.positionKey(position.CharacterID)
	data, err := json.Marshal(position)
	if err != nil {
		return fmt.Errorf("failed to marshal position: %w", err)
	}

	err = c.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to cache position: %w", err)
	}

	return nil
}

// GetPosition retrieves cached character position
func (c *RedisCharacterCache) GetPosition(ctx context.Context, characterID uuid.UUID) (*character.Position, error) {
	key := c.positionKey(characterID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get cached position: %w", err)
	}

	var position character.Position
	err = json.Unmarshal(data, &position)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal position: %w", err)
	}

	return &position, nil
}

// DeletePosition removes character position from cache
func (c *RedisCharacterCache) DeletePosition(ctx context.Context, characterID uuid.UUID) error {
	key := c.positionKey(characterID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cached position: %w", err)
	}
	return nil
}

// WarmCharacterCache pre-loads all character data into cache
func (c *RedisCharacterCache) WarmCharacterCache(ctx context.Context, characterID uuid.UUID) error {
	// This method would typically be called after character selection
	// to pre-load all character data for faster access during gameplay
	// The actual implementation would fetch data from repositories
	// and cache it using the Set methods above
	
	// For now, we'll just invalidate to ensure fresh data
	return c.InvalidateCharacterData(ctx, characterID)
}

// InvalidateCharacterData removes all cached data for a character
func (c *RedisCharacterCache) InvalidateCharacterData(ctx context.Context, characterID uuid.UUID) error {
	pipe := c.client.Pipeline()
	
	// Delete all character-related cache entries
	pipe.Del(ctx, c.characterKey(characterID))
	pipe.Del(ctx, c.appearanceKey(characterID))
	pipe.Del(ctx, c.statsKey(characterID))
	pipe.Del(ctx, c.positionKey(characterID))
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to invalidate character data: %w", err)
	}
	
	return nil
}

// InvalidateUserData removes all cached data for a user
func (c *RedisCharacterCache) InvalidateUserData(ctx context.Context, userID uuid.UUID) error {
	pipe := c.client.Pipeline()
	
	// Delete user-related cache entries
	pipe.Del(ctx, c.userCharactersKey(userID))
	pipe.Del(ctx, c.characterCountKey(userID))
	pipe.Del(ctx, c.selectedCharacterKey(userID))
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to invalidate user data: %w", err)
	}
	
	return nil
}

// Cache key helper methods
func (c *RedisCharacterCache) characterKey(characterID uuid.UUID) string {
	return fmt.Sprintf("%s:character:%s", c.prefix, characterID.String())
}

func (c *RedisCharacterCache) userCharactersKey(userID uuid.UUID) string {
	return fmt.Sprintf("%s:user_characters:%s", c.prefix, userID.String())
}

func (c *RedisCharacterCache) characterCountKey(userID uuid.UUID) string {
	return fmt.Sprintf("%s:character_count:%s", c.prefix, userID.String())
}

func (c *RedisCharacterCache) selectedCharacterKey(userID uuid.UUID) string {
	return fmt.Sprintf("%s:selected_character:%s", c.prefix, userID.String())
}

func (c *RedisCharacterCache) appearanceKey(characterID uuid.UUID) string {
	return fmt.Sprintf("%s:appearance:%s", c.prefix, characterID.String())
}

func (c *RedisCharacterCache) statsKey(characterID uuid.UUID) string {
	return fmt.Sprintf("%s:stats:%s", c.prefix, characterID.String())
}

func (c *RedisCharacterCache) positionKey(characterID uuid.UUID) string {
	return fmt.Sprintf("%s:position:%s", c.prefix, characterID.String())
}