# Character Service Implementation

This package implements the character management system for the MMORPG backend, providing complete CRUD operations with hexagonal architecture.

## Features

### Core Operations
- **Create Character**: Full character creation with appearance, stats, and position initialization
- **Get Character**: Retrieve character by ID or name
- **List Characters**: Get all characters for a user
- **Update Character**: Modify character attributes (limited fields)
- **Delete Character**: Soft delete with 30-day recovery window
- **Restore Character**: Recover soft-deleted characters

### Additional Features
- **Appearance Management**: Customizable character appearance with race/gender defaults
- **Stats Management**: Class-based stat initialization and stat point allocation
- **Position Tracking**: World position management with safe position teleportation
- **Validation**: Comprehensive business rule validation
- **Transaction Support**: Atomic operations for data consistency

## Architecture

### Service Layer (`service.go`)
- Implements business logic and validation
- Orchestrates repository operations
- Handles error mapping and logging

### Repository Layer
- `character_repository_postgres.go`: Character CRUD operations
- `appearance_repository_postgres.go`: Appearance data management
- `stats_repository_postgres.go`: Character statistics
- `position_repository_postgres.go`: Position and movement tracking

### Transaction Support (`service_transactional.go`)
- Enhanced service with database transaction support
- Ensures atomic operations for character creation
- Prevents partial data states

## Usage Example

```go
// Create service
config := &character.Config{
    MaxCharactersPerUser:      5,
    MaxCharacterNameLength:    30,
    MinCharacterNameLength:    3,
    DefaultStartingLevel:      1,
    DefaultStartingExperience: 0,
}

service := character.NewCharacterService(
    characterRepo,
    appearanceRepo,
    statsRepo,
    positionRepo,
    config,
    logger,
)

// Create character
req := &portsCharacter.CreateCharacterRequest{
    UserID:     "user-uuid",
    Name:       "HeroName",
    SlotNumber: 1,
    ClassType:  character.ClassWarrior,
    Race:       character.RaceHuman,
    Gender:     character.GenderMale,
    Appearance: &portsCharacter.CharacterAppearanceOptions{
        HairStyle: &hairStyle,
        HairColor: &hairColor,
    },
}

char, err := service.CreateCharacter(ctx, req)
```

## Business Rules

### Character Creation
- User must not exceed character limit (configurable, default: 5)
- Character name must be unique (case-insensitive)
- Name must be 3-30 characters, alphanumeric with spaces/hyphens
- Slot number must be 1-100 and unique per user
- Valid class, race, and gender required

### Character Deletion
- Soft delete with 30-day recovery window
- Deleted characters free up their slot
- Name becomes available after permanent deletion
- Related data (appearance, stats, position) preserved during soft delete

### Stat Allocation
- Characters receive stat points on level up
- Points can be allocated to primary attributes
- Derived stats automatically recalculated
- Class-specific stat modifiers applied

## Database Schema

### Tables
- `characters`: Core character data
- `character_appearance`: Visual customization
- `character_stats`: Attributes and combat stats
- `character_position`: World location and movement
- `character_name_history`: Name change audit log

### Key Indexes
- User-based queries: `idx_characters_user_active`
- Name uniqueness: `idx_characters_name_lower`
- Slot management: `idx_characters_user_slot_active`
- Performance queries: `idx_characters_level_experience`

### Stored Procedures
- `soft_delete_character()`: Marks character as deleted
- `restore_character()`: Restores soft-deleted character
- `cleanup_deleted_characters()`: Removes expired deletions
- `find_nearby_characters()`: Spatial queries for networking
- `save_safe_position()`: Saves current as safe position
- `teleport_to_safe_position()`: Returns to safe location

## Error Handling

### Domain Errors
- `ErrCharacterNotFound`: Character doesn't exist
- `ErrCharacterNameTaken`: Name already in use
- `ErrCharacterLimitReached`: User at max characters
- `ErrSlotOccupied`: Slot number in use
- `ErrCharacterDeleted`: Accessing deleted character
- `ErrCharacterCannotBeRestored`: Outside recovery window

### Validation Errors
- `ErrInvalidCharacterName`: Name format invalid
- `ErrInvalidClass/Race/Gender`: Invalid enum value
- `ErrInvalidSlotNumber`: Slot out of range
- `ErrInsufficientStatPoints`: Not enough points to allocate

## Testing

The service includes comprehensive unit tests (`service_test.go`) covering:
- Character creation scenarios
- Validation edge cases
- Soft delete/restore operations
- Concurrent access handling
- Transaction rollback scenarios

Run tests:
```bash
go test ./internal/application/character/...
```

## Performance Considerations

1. **Indexed Queries**: All common queries use appropriate indexes
2. **Batch Operations**: `GetMultipleByCharacterIDs` for efficient bulk fetching
3. **Materialized Views**: `mv_character_statistics` for analytics
4. **Connection Pooling**: Repository uses shared DB connection pool
5. **Transaction Isolation**: Read-committed level for optimal performance

## Future Enhancements

- Character transfer between servers
- Character appearance presets
- Achievement-based stat bonuses
- Character templates for quick creation
- Advanced name filtering system