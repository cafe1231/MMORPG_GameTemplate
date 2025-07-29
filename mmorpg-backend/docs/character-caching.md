# Character Service Redis Caching Implementation

## Overview

The character service now includes a comprehensive Redis caching layer to improve performance and reduce database load. The caching implementation follows the hexagonal architecture pattern with clean separation between the cache interface (port) and Redis implementation (adapter).

## Architecture

### Cache Interface (Port)
- **Location**: `internal/ports/character/cache.go`
- **Interface**: `CharacterCache`
- Defines all caching operations for character-related data

### Redis Implementation (Adapter)
- **Location**: `internal/adapters/character/redis/cache.go`
- **Implementation**: `RedisCharacterCache`
- Implements the `CharacterCache` interface using Redis

### Integration
- **Service**: `internal/application/character/service.go`
- The character service uses the cache for read operations and invalidates cache on updates

## Cached Data Types

### 1. Individual Character Data
- **Key**: `character:character:{characterID}`
- **TTL**: 10 minutes
- **Usage**: Frequently accessed character details
- **Invalidated on**: Character update, deletion

### 2. User Character List
- **Key**: `character:user_characters:{userID}`
- **TTL**: 5 minutes
- **Usage**: Character selection screen
- **Invalidated on**: Character creation, deletion

### 3. Character Count
- **Key**: `character:character_count:{userID}`
- **TTL**: 1 minute
- **Usage**: Slot validation for character creation
- **Invalidated on**: Character creation, deletion

### 4. Selected Character
- **Key**: `character:selected_character:{userID}`
- **TTL**: 24 hours (session duration)
- **Usage**: Current gameplay session
- **Invalidated on**: Character deselection, logout

### 5. Character Appearance
- **Key**: `character:appearance:{characterID}`
- **TTL**: 10 minutes
- **Usage**: Character rendering
- **Invalidated on**: Appearance update

### 6. Character Stats
- **Key**: `character:stats:{characterID}`
- **TTL**: 10 minutes
- **Usage**: Combat calculations, UI display
- **Invalidated on**: Stat updates, level up

### 7. Character Position
- **Key**: `character:position:{characterID}`
- **TTL**: 30 seconds
- **Usage**: Real-time position tracking
- **Updated on**: Position changes

## Caching Strategies

### Read-Through Cache
- Service first checks cache for data
- On cache miss, fetches from database
- Caches the result for future requests

### Cache Invalidation
- Updates and deletes invalidate related cache entries
- Bulk invalidation for user-level changes
- Atomic operations using Redis pipelines

### Cache Warming
- When a character is selected, related data is pre-loaded
- Asynchronous warming to avoid blocking

### Error Handling
- Cache failures don't break functionality
- Service falls back to database on cache errors
- Errors are logged but not propagated

## Usage Examples

### Character Selection
```go
// Select character for gameplay
err := characterService.SelectCharacter(ctx, characterID, userID)

// Get currently selected character
char, err := characterService.GetSelectedCharacter(ctx, userID)

// Deselect character
err := characterService.DeselectCharacter(ctx, userID)
```

### Efficient Data Access
```go
// First call hits database, subsequent calls use cache
char, err := characterService.GetCharacter(ctx, characterID)

// Character list cached for 5 minutes
characters, err := characterService.ListCharactersByUser(ctx, userID)

// Count cached for slot validation
canCreate, err := characterService.CanCreateCharacter(ctx, userID)
```

## Configuration

### Default TTL Values
```go
Character:         10 * time.Minute
CharacterList:     5 * time.Minute
CharacterCount:    1 * time.Minute
SelectedCharacter: 24 * time.Hour
Appearance:        10 * time.Minute
Stats:             10 * time.Minute
Position:          30 * time.Second
```

### Custom TTL Configuration
```go
cache := redis.NewRedisCharacterCache(client, "character", &portsCharacter.CacheTTL{
    Character:         15 * time.Minute,
    CharacterList:     10 * time.Minute,
    // ... other custom values
})
```

## Performance Benefits

1. **Reduced Database Load**: Frequently accessed data served from memory
2. **Faster Response Times**: Sub-millisecond cache hits vs database queries
3. **Session Persistence**: Selected character cached for entire session
4. **Real-time Updates**: Position caching with short TTL for movement

## Testing

### Unit Tests
- **Location**: `internal/adapters/character/redis/cache_test.go`
- Tests all cache operations
- Validates TTL behavior
- Tests invalidation patterns

### Integration Testing
```bash
# Run Redis locally
docker run -d -p 6379:6379 redis:alpine

# Run tests
go test ./internal/adapters/character/redis/...
```

## Monitoring

### Cache Metrics to Track
- Cache hit/miss ratio
- Operation latency
- Memory usage
- Key count by pattern

### Redis Commands for Monitoring
```bash
# Check all character keys
redis-cli --scan --pattern "character:*"

# Monitor cache operations
redis-cli monitor

# Check memory usage
redis-cli info memory
```

## Best Practices

1. **Always check cache availability**: Service should work without cache
2. **Use appropriate TTLs**: Balance between freshness and performance
3. **Invalidate on updates**: Keep cache consistent with database
4. **Monitor cache performance**: Track hit rates and adjust TTLs
5. **Handle cache stampede**: Use cache warming for popular data