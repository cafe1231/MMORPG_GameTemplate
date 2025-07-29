# 🎭 Phase 1.5: Character System Foundation - Technical Architecture

## Executive Summary

Phase 1.5 bridges the authentication system (Phase 1) and real-time networking (Phase 2) by implementing a comprehensive character management system. This phase enables players to create, customize, select, and manage game characters, establishing the player's in-game identity before they enter the networked game world.

**Duration**: Estimated 3-4 weeks  
**Prerequisites**: Phase 1 (Authentication) complete  
**Dependencies**: JWT auth, user management, PostgreSQL, Redis

---

## System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Client (Unreal Engine)                     │
├─────────────────────────────────────────────────────────────┤
│  Character UI       │  Character Manager │  Data Models      │
│  ├─ Creation Form   │  ├─ CRUD Operations │  ├─ Character    │
│  ├─ Selection Grid  │  ├─ Validation      │  ├─ Appearance   │
│  ├─ Preview System  │  ├─ Caching         │  ├─ Stats        │
│  └─ Delete Confirm  │  └─ State Manager   │  └─ Equipment    │
└────────────────────┴────────────────────┴───────────────────┘
                              │
                         HTTP/HTTPS
                              │
┌─────────────────────────────────────────────────────────────┐
│                      Gateway Service                         │
│                  (Routes to Character Service)               │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────┴──────────────────────────────────────┐
│                   Character Service                          │
├──────────────────────────────────────────────────────────────┤
│  Domain Layer      │  Application Layer │  Adapter Layer     │
│  ├─ Character      │  ├─ Create Char    │  ├─ HTTP Handler  │
│  ├─ Appearance     │  ├─ Update Char    │  ├─ Validators    │
│  ├─ Stats          │  ├─ Delete Char    │  ├─ Repositories │
│  └─ Constraints    │  └─ List Chars     │  └─ Proto Serial │
└────────────────────┴────────────────────┴───────────────────┘
                              │
                    ┌─────────┴─────────┐
                    │                   │
┌───────────────────┴──┐    ┌──────────┴────────────┐
│    PostgreSQL        │    │      Redis            │
│  ├─ Characters Table │    │  ├─ Character Cache   │
│  ├─ Appearance Data  │    │  ├─ Selection State   │
│  └─ Character Stats  │    │  └─ Validation Cache  │
└──────────────────────┘    └───────────────────────┘
```

### Component Design

#### 1. Backend Character Service

**Domain Entities**:
```go
// Character represents a player character
type Character struct {
    ID              uuid.UUID
    UserID          uuid.UUID
    Name            string
    Class           CharacterClass
    Race            CharacterRace
    Gender          Gender
    Level           int
    Experience      int64
    Appearance      CharacterAppearance
    Stats           CharacterStats
    Position        WorldPosition
    CreatedAt       time.Time
    UpdatedAt       time.Time
    LastPlayedAt    *time.Time
    DeletedAt       *time.Time
    IsSelected      bool
}

// CharacterAppearance stores customization data
type CharacterAppearance struct {
    HairStyle       int
    HairColor       string
    FaceType        int
    SkinTone        string
    BodyType        int
    Height          float32
    CustomFeatures  map[string]interface{}
}

// CharacterStats represents base character attributes
type CharacterStats struct {
    Health          int
    MaxHealth       int
    Mana            int
    MaxMana         int
    Strength        int
    Intelligence    int
    Dexterity       int
    Vitality        int
    // Class-specific stats
    ClassStats      map[string]int
}

// WorldPosition tracks character location
type WorldPosition struct {
    Zone            string
    X               float64
    Y               float64
    Z               float64
    Rotation        float64
}
```

**Service Interface**:
```go
type CharacterService interface {
    // Character CRUD operations
    CreateCharacter(ctx context.Context, userID uuid.UUID, req CreateCharacterRequest) (*Character, error)
    GetCharacter(ctx context.Context, characterID uuid.UUID) (*Character, error)
    GetUserCharacters(ctx context.Context, userID uuid.UUID) ([]*Character, error)
    UpdateCharacter(ctx context.Context, characterID uuid.UUID, updates CharacterUpdates) (*Character, error)
    DeleteCharacter(ctx context.Context, characterID uuid.UUID) error
    
    // Character selection
    SelectCharacter(ctx context.Context, userID uuid.UUID, characterID uuid.UUID) (*Character, error)
    GetSelectedCharacter(ctx context.Context, userID uuid.UUID) (*Character, error)
    
    // Validation
    ValidateCharacterName(ctx context.Context, name string) error
    IsCharacterOwnedBy(ctx context.Context, characterID, userID uuid.UUID) bool
}
```

#### 2. Database Schema

```sql
-- Characters table
CREATE TABLE characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(32) UNIQUE NOT NULL,
    class INTEGER NOT NULL,
    race INTEGER NOT NULL,
    gender INTEGER NOT NULL,
    level INTEGER DEFAULT 1,
    experience BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_played_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    is_selected BOOLEAN DEFAULT FALSE,
    CONSTRAINT check_level CHECK (level >= 1 AND level <= 100),
    CONSTRAINT check_name_length CHECK (char_length(name) >= 3 AND char_length(name) <= 32)
);

-- Character appearance table
CREATE TABLE character_appearance (
    character_id UUID PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    hair_style INTEGER DEFAULT 0,
    hair_color VARCHAR(7) DEFAULT '#000000',
    face_type INTEGER DEFAULT 0,
    skin_tone VARCHAR(7) DEFAULT '#F5DEB3',
    body_type INTEGER DEFAULT 0,
    height REAL DEFAULT 1.0,
    custom_features JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Character stats table
CREATE TABLE character_stats (
    character_id UUID PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    health INTEGER DEFAULT 100,
    max_health INTEGER DEFAULT 100,
    mana INTEGER DEFAULT 100,
    max_mana INTEGER DEFAULT 100,
    strength INTEGER DEFAULT 10,
    intelligence INTEGER DEFAULT 10,
    dexterity INTEGER DEFAULT 10,
    vitality INTEGER DEFAULT 10,
    class_stats JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Character position table
CREATE TABLE character_positions (
    character_id UUID PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    zone VARCHAR(64) DEFAULT 'starter_zone',
    x DOUBLE PRECISION DEFAULT 0,
    y DOUBLE PRECISION DEFAULT 0,
    z DOUBLE PRECISION DEFAULT 0,
    rotation DOUBLE PRECISION DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_characters_user_id ON characters(user_id);
CREATE INDEX idx_characters_name_lower ON characters(LOWER(name));
CREATE INDEX idx_characters_deleted_at ON characters(deleted_at);
CREATE INDEX idx_characters_is_selected ON characters(user_id, is_selected) WHERE is_selected = TRUE;

-- Constraints
ALTER TABLE characters ADD CONSTRAINT unique_selected_per_user 
    EXCLUDE (user_id WITH =) WHERE (is_selected = TRUE AND deleted_at IS NULL);
```

#### 3. API Endpoints

**Character Management**:
- `POST /api/v1/characters` - Create new character
- `GET /api/v1/characters` - List user's characters
- `GET /api/v1/characters/{id}` - Get specific character
- `PUT /api/v1/characters/{id}` - Update character (limited fields)
- `DELETE /api/v1/characters/{id}` - Soft delete character
- `POST /api/v1/characters/{id}/select` - Select character for gameplay
- `GET /api/v1/characters/selected` - Get currently selected character
- `POST /api/v1/characters/validate-name` - Check name availability

**Request/Response Examples**:

```json
// Create Character Request
POST /api/v1/characters
{
    "name": "Gandalf",
    "class": "mage",
    "race": "human",
    "gender": "male",
    "appearance": {
        "hair_style": 3,
        "hair_color": "#FFFFFF",
        "face_type": 2,
        "skin_tone": "#F5DEB3",
        "body_type": 1,
        "height": 1.8
    }
}

// Character List Response
GET /api/v1/characters
{
    "success": true,
    "characters": [
        {
            "id": "123e4567-e89b-12d3-a456-426614174000",
            "name": "Gandalf",
            "class": "mage",
            "race": "human",
            "gender": "male",
            "level": 1,
            "last_played_at": "2025-07-29T10:00:00Z",
            "is_selected": true
        }
    ],
    "character_slots": {
        "used": 1,
        "max": 5
    }
}
```

#### 4. Frontend Integration

**Character Subsystem**:
```cpp
UCLASS(BlueprintType)
class MMORPGCORE_API UMMORPGCharacterSubsystem : public UGameInstanceSubsystem
{
    GENERATED_BODY()

public:
    // Character operations
    UFUNCTION(BlueprintCallable)
    void CreateCharacter(const FCharacterCreateRequest& Request);
    
    UFUNCTION(BlueprintCallable)
    void GetCharacterList();
    
    UFUNCTION(BlueprintCallable)
    void SelectCharacter(const FString& CharacterID);
    
    UFUNCTION(BlueprintCallable)
    void DeleteCharacter(const FString& CharacterID);
    
    // Character data
    UFUNCTION(BlueprintPure)
    TArray<FCharacterInfo> GetCachedCharacters() const;
    
    UFUNCTION(BlueprintPure)
    FCharacterInfo GetSelectedCharacter() const;
    
    // Events
    UPROPERTY(BlueprintAssignable)
    FOnCharacterListReceived OnCharacterListReceived;
    
    UPROPERTY(BlueprintAssignable)
    FOnCharacterCreated OnCharacterCreated;
    
    UPROPERTY(BlueprintAssignable)
    FOnCharacterSelected OnCharacterSelected;
};
```

**UI Widgets**:
- `UMMORPGCharacterCreationWidget` - Character creation form
- `UMMORPGCharacterSelectionWidget` - Character grid selection
- `UMMORPGCharacterPreviewWidget` - 3D character preview
- `UMMORPGCharacterInfoWidget` - Character details display

### Integration Points

#### With Phase 1 (Authentication)
- JWT tokens required for all character API calls
- User ID extracted from JWT for ownership validation
- Character count tracked in user table
- Session required for character operations

#### With Phase 2 (Networking)
- Selected character ID included in WebSocket auth
- Character data needed for initial world spawn
- Character state updates via real-time connection
- Position synchronization preparation

#### With Phase 3 (Gameplay)
- Character stats ready for combat system
- Inventory system character binding
- Guild/party character references
- Achievement character tracking

### Security Considerations

1. **Authorization**:
   - All operations verify JWT and user ownership
   - Character names validated for profanity/reserved words
   - Rate limiting on character creation
   - Soft delete with recovery period

2. **Data Validation**:
   - Server-side validation of all inputs
   - Enum validation for class/race/gender
   - Appearance values within allowed ranges
   - Name uniqueness across all characters

3. **Anti-Cheat Preparation**:
   - Stats validated against class constraints
   - Position sanity checks
   - Creation timestamp verification
   - Audit log for character modifications

### Performance Optimizations

1. **Caching Strategy**:
   - Redis cache for character list per user
   - Selected character in Redis for fast access
   - Name availability cache with TTL
   - Appearance data compressed in cache

2. **Database Optimization**:
   - Composite indexes for common queries
   - Partial indexes for active characters
   - JSONB for flexible custom data
   - Connection pooling configuration

3. **API Optimization**:
   - Batch character loading
   - Selective field queries
   - Pagination for admin interfaces
   - Async character deletion

### Error Handling

**Error Codes**:
- `CHARACTER_NAME_TAKEN` - Name already in use
- `CHARACTER_LIMIT_REACHED` - Max characters exceeded
- `INVALID_CHARACTER_NAME` - Name validation failed
- `CHARACTER_NOT_FOUND` - Character doesn't exist
- `CHARACTER_NOT_OWNED` - Ownership validation failed
- `CHARACTER_CREATION_FAILED` - Generic creation error

### Monitoring and Analytics

1. **Metrics**:
   - Character creation rate
   - Popular class/race combinations
   - Average characters per user
   - Character deletion rate
   - API response times

2. **Events**:
   - Character created
   - Character selected
   - Character deleted
   - Character restored
   - Name validation failures

---

## Implementation Phases

### Phase 1.5A: Backend Character Service (Week 1-2)
1. Create character service with hexagonal architecture
2. Implement database schema and migrations
3. Build character CRUD operations
4. Add validation and business rules
5. Integrate with auth service
6. Create API endpoints and tests

### Phase 1.5B: Frontend Integration (Week 2-3)
1. Create character data structures
2. Build character subsystem
3. Implement UI widgets
4. Add 3D character preview
5. Connect to backend APIs
6. Handle error states

### Phase 1.5C: Polish and Testing (Week 3-4)
1. Performance optimization
2. Security hardening
3. UI/UX improvements
4. Integration testing
5. Documentation
6. Migration guides

---

## Success Criteria

1. **Functional Requirements**:
   - Players can create multiple characters
   - Character names are unique and validated
   - Characters can be selected for gameplay
   - Soft deletion with recovery option
   - All operations require authentication

2. **Performance Requirements**:
   - Character creation < 1 second
   - Character list load < 200ms
   - Name validation < 100ms
   - Support 10k+ concurrent users

3. **Quality Requirements**:
   - 90%+ code coverage
   - Zero critical security issues
   - Comprehensive error handling
   - Full API documentation
   - Migration path defined

---

## Conclusion

Phase 1.5 provides the essential character management layer between authentication and real-time gameplay. By implementing a robust character system with proper data models, security, and performance optimizations, we create a solid foundation for the networked gameplay features in Phase 2 and beyond.