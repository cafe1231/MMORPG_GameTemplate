# 🗃️ Phase 1.5: Character System - Data Models Specification

## Overview

This document defines the complete data models for the character system, including database schemas, API contracts, domain entities, and client-side structures.

---

## Backend Data Models

### Domain Entities (Go)

```go
package character

import (
    "time"
    "github.com/google/uuid"
)

// CharacterClass enum
type CharacterClass int32

const (
    CharacterClassUnspecified CharacterClass = iota
    CharacterClassWarrior
    CharacterClassMage
    CharacterClassArcher
    CharacterClassRogue
    CharacterClassPriest
    CharacterClassPaladin
    CharacterClassWarlock
    CharacterClassDruid
)

// CharacterRace enum
type CharacterRace int32

const (
    CharacterRaceUnspecified CharacterRace = iota
    CharacterRaceHuman
    CharacterRaceElf
    CharacterRaceDwarf
    CharacterRaceOrc
    CharacterRaceUndead
    CharacterRaceTauren
    CharacterRaceGnome
    CharacterRaceTroll
)

// Gender enum
type Gender int32

const (
    GenderUnspecified Gender = iota
    GenderMale
    GenderFemale
    GenderOther
)

// Character represents a player's game character
type Character struct {
    ID              uuid.UUID              `json:"id"`
    UserID          uuid.UUID              `json:"user_id"`
    Name            string                 `json:"name"`
    Class           CharacterClass         `json:"class"`
    Race            CharacterRace          `json:"race"`
    Gender          Gender                 `json:"gender"`
    Level           int                    `json:"level"`
    Experience      int64                  `json:"experience"`
    Appearance      CharacterAppearance    `json:"appearance"`
    Stats           CharacterStats         `json:"stats"`
    Position        WorldPosition          `json:"position"`
    Equipment       CharacterEquipment     `json:"equipment"`
    Inventory       CharacterInventory     `json:"inventory"`
    CreatedAt       time.Time              `json:"created_at"`
    UpdatedAt       time.Time              `json:"updated_at"`
    LastPlayedAt    *time.Time             `json:"last_played_at,omitempty"`
    DeletedAt       *time.Time             `json:"deleted_at,omitempty"`
    IsSelected      bool                   `json:"is_selected"`
}

// CharacterAppearance defines character visual customization
type CharacterAppearance struct {
    HairStyle       int                    `json:"hair_style"`
    HairColor       string                 `json:"hair_color"`
    FaceType        int                    `json:"face_type"`
    SkinTone        string                 `json:"skin_tone"`
    BodyType        int                    `json:"body_type"`
    Height          float32                `json:"height"`
    EyeColor        string                 `json:"eye_color"`
    FacialHair      int                    `json:"facial_hair"`
    Scars           []int                  `json:"scars"`
    Tattoos         []int                  `json:"tattoos"`
    CustomFeatures  map[string]interface{} `json:"custom_features"`
}

// CharacterStats represents character attributes
type CharacterStats struct {
    // Core stats
    Health          int                    `json:"health"`
    MaxHealth       int                    `json:"max_health"`
    Mana            int                    `json:"mana"`
    MaxMana         int                    `json:"max_mana"`
    Stamina         int                    `json:"stamina"`
    MaxStamina      int                    `json:"max_stamina"`
    
    // Primary attributes
    Strength        int                    `json:"strength"`
    Intelligence    int                    `json:"intelligence"`
    Dexterity       int                    `json:"dexterity"`
    Vitality        int                    `json:"vitality"`
    Wisdom          int                    `json:"wisdom"`
    Charisma        int                    `json:"charisma"`
    
    // Secondary attributes
    AttackPower     int                    `json:"attack_power"`
    SpellPower      int                    `json:"spell_power"`
    Defense         int                    `json:"defense"`
    CritChance      float32                `json:"crit_chance"`
    CritDamage      float32                `json:"crit_damage"`
    AttackSpeed     float32                `json:"attack_speed"`
    MoveSpeed       float32                `json:"move_speed"`
    
    // Resistances
    PhysicalResist  int                    `json:"physical_resist"`
    MagicalResist   int                    `json:"magical_resist"`
    FireResist      int                    `json:"fire_resist"`
    IceResist       int                    `json:"ice_resist"`
    LightningResist int                    `json:"lightning_resist"`
    
    // Class-specific stats
    ClassStats      map[string]int         `json:"class_stats"`
}

// WorldPosition tracks character location in the game world
type WorldPosition struct {
    Zone            string                 `json:"zone"`
    X               float64                `json:"x"`
    Y               float64                `json:"y"`
    Z               float64                `json:"z"`
    Rotation        float64                `json:"rotation"`
    MapLayer        int                    `json:"map_layer"`
}

// CharacterEquipment represents equipped items (placeholder for Phase 3)
type CharacterEquipment struct {
    Head            *uuid.UUID             `json:"head,omitempty"`
    Shoulders       *uuid.UUID             `json:"shoulders,omitempty"`
    Chest           *uuid.UUID             `json:"chest,omitempty"`
    Hands           *uuid.UUID             `json:"hands,omitempty"`
    Legs            *uuid.UUID             `json:"legs,omitempty"`
    Feet            *uuid.UUID             `json:"feet,omitempty"`
    MainHand        *uuid.UUID             `json:"main_hand,omitempty"`
    OffHand         *uuid.UUID             `json:"off_hand,omitempty"`
    Rings           []uuid.UUID            `json:"rings"`
    Trinkets        []uuid.UUID            `json:"trinkets"`
}

// CharacterInventory represents inventory state (placeholder for Phase 3)
type CharacterInventory struct {
    Slots           int                    `json:"slots"`
    UsedSlots       int                    `json:"used_slots"`
    Items           []uuid.UUID            `json:"items"`
    Gold            int64                  `json:"gold"`
}
```

### API Request/Response Models

```go
// CreateCharacterRequest for character creation
type CreateCharacterRequest struct {
    Name            string                 `json:"name" validate:"required,min=3,max=32,alphanum"`
    Class           CharacterClass         `json:"class" validate:"required,min=1,max=8"`
    Race            CharacterRace          `json:"race" validate:"required,min=1,max=8"`
    Gender          Gender                 `json:"gender" validate:"required,min=1,max=3"`
    Appearance      CharacterAppearance    `json:"appearance" validate:"required"`
    StartingZone    string                 `json:"starting_zone,omitempty"`
}

// UpdateCharacterRequest for character updates
type UpdateCharacterRequest struct {
    Appearance      *CharacterAppearance   `json:"appearance,omitempty"`
    CustomFeatures  map[string]interface{} `json:"custom_features,omitempty"`
}

// CharacterListResponse returns user's characters
type CharacterListResponse struct {
    Success         bool                   `json:"success"`
    Characters      []CharacterSummary     `json:"characters"`
    CharacterSlots  CharacterSlotInfo      `json:"character_slots"`
    ErrorMessage    string                 `json:"error_message,omitempty"`
    ErrorCode       string                 `json:"error_code,omitempty"`
}

// CharacterSummary provides overview info for character selection
type CharacterSummary struct {
    ID              uuid.UUID              `json:"id"`
    Name            string                 `json:"name"`
    Class           CharacterClass         `json:"class"`
    Race            CharacterRace          `json:"race"`
    Gender          Gender                 `json:"gender"`
    Level           int                    `json:"level"`
    Zone            string                 `json:"zone"`
    LastPlayedAt    *time.Time             `json:"last_played_at,omitempty"`
    IsSelected      bool                   `json:"is_selected"`
    ThumbnailURL    string                 `json:"thumbnail_url,omitempty"`
}

// CharacterSlotInfo tracks character slots
type CharacterSlotInfo struct {
    Used            int                    `json:"used"`
    Max             int                    `json:"max"`
    IsPremium       bool                   `json:"is_premium"`
}

// CharacterDetailsResponse for full character data
type CharacterDetailsResponse struct {
    Success         bool                   `json:"success"`
    Character       *Character             `json:"character,omitempty"`
    ErrorMessage    string                 `json:"error_message,omitempty"`
    ErrorCode       string                 `json:"error_code,omitempty"`
}

// ValidateNameRequest checks name availability
type ValidateNameRequest struct {
    Name            string                 `json:"name" validate:"required,min=3,max=32"`
}

// ValidateNameResponse returns validation result
type ValidateNameResponse struct {
    Available       bool                   `json:"available"`
    Reason          string                 `json:"reason,omitempty"`
    Suggestions     []string               `json:"suggestions,omitempty"`
}
```

### Database Schema Extensions

```sql
-- Add character-related columns to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS selected_character_id UUID REFERENCES characters(id);

-- Character name reservations (for preventing abuse)
CREATE TABLE character_name_reservations (
    name VARCHAR(32) PRIMARY KEY,
    reserved_by UUID REFERENCES users(id),
    reserved_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() + INTERVAL '30 minutes'
);

-- Character deletion log for recovery
CREATE TABLE character_deletion_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL,
    user_id UUID NOT NULL,
    character_data JSONB NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    recovery_token VARCHAR(64) UNIQUE,
    recovered_at TIMESTAMP WITH TIME ZONE,
    permanent_delete_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() + INTERVAL '30 days'
);

-- Character class/race combinations (for validation)
CREATE TABLE valid_class_race_combinations (
    class INTEGER NOT NULL,
    race INTEGER NOT NULL,
    PRIMARY KEY (class, race)
);

-- Insert valid combinations
INSERT INTO valid_class_race_combinations (class, race) VALUES
-- Warriors can be any race
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5), (1, 6), (1, 7), (1, 8),
-- Mages cannot be Orc or Tauren
(2, 1), (2, 2), (2, 3), (2, 5), (2, 7), (2, 8),
-- Add more combinations as needed
;

-- Starting stats by class
CREATE TABLE class_starting_stats (
    class INTEGER PRIMARY KEY,
    health INTEGER NOT NULL,
    mana INTEGER NOT NULL,
    stamina INTEGER NOT NULL,
    strength INTEGER NOT NULL,
    intelligence INTEGER NOT NULL,
    dexterity INTEGER NOT NULL,
    vitality INTEGER NOT NULL,
    wisdom INTEGER NOT NULL,
    charisma INTEGER NOT NULL
);

-- Starting positions by race
CREATE TABLE race_starting_positions (
    race INTEGER PRIMARY KEY,
    zone VARCHAR(64) NOT NULL,
    x DOUBLE PRECISION NOT NULL,
    y DOUBLE PRECISION NOT NULL,
    z DOUBLE PRECISION NOT NULL,
    rotation DOUBLE PRECISION NOT NULL
);
```

---

## Frontend Data Models

### Unreal Engine Structures (C++)

```cpp
// CharacterTypes.h
#pragma once

#include "CoreMinimal.h"
#include "CharacterTypes.generated.h"

UENUM(BlueprintType)
enum class ECharacterClass : uint8
{
    Unspecified = 0,
    Warrior = 1,
    Mage = 2,
    Archer = 3,
    Rogue = 4,
    Priest = 5,
    Paladin = 6,
    Warlock = 7,
    Druid = 8
};

UENUM(BlueprintType)
enum class ECharacterRace : uint8
{
    Unspecified = 0,
    Human = 1,
    Elf = 2,
    Dwarf = 3,
    Orc = 4,
    Undead = 5,
    Tauren = 6,
    Gnome = 7,
    Troll = 8
};

UENUM(BlueprintType)
enum class EGender : uint8
{
    Unspecified = 0,
    Male = 1,
    Female = 2,
    Other = 3
};

USTRUCT(BlueprintType)
struct FCharacterAppearance
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    int32 HairStyle = 0;

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    FLinearColor HairColor = FLinearColor::Black;

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    int32 FaceType = 0;

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    FLinearColor SkinTone = FLinearColor(0.96f, 0.87f, 0.70f);

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    int32 BodyType = 0;

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    float Height = 1.0f;

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    FLinearColor EyeColor = FLinearColor(0.0f, 0.5f, 1.0f);

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    int32 FacialHair = 0;

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    TArray<int32> Scars;

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    TArray<int32> Tattoos;
};

USTRUCT(BlueprintType)
struct FCharacterStats
{
    GENERATED_BODY()

    // Core stats
    UPROPERTY(BlueprintReadOnly)
    int32 Health = 100;

    UPROPERTY(BlueprintReadOnly)
    int32 MaxHealth = 100;

    UPROPERTY(BlueprintReadOnly)
    int32 Mana = 100;

    UPROPERTY(BlueprintReadOnly)
    int32 MaxMana = 100;

    UPROPERTY(BlueprintReadOnly)
    int32 Stamina = 100;

    UPROPERTY(BlueprintReadOnly)
    int32 MaxStamina = 100;

    // Primary attributes
    UPROPERTY(BlueprintReadOnly)
    int32 Strength = 10;

    UPROPERTY(BlueprintReadOnly)
    int32 Intelligence = 10;

    UPROPERTY(BlueprintReadOnly)
    int32 Dexterity = 10;

    UPROPERTY(BlueprintReadOnly)
    int32 Vitality = 10;

    UPROPERTY(BlueprintReadOnly)
    int32 Wisdom = 10;

    UPROPERTY(BlueprintReadOnly)
    int32 Charisma = 10;
};

USTRUCT(BlueprintType)
struct FWorldPosition
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadOnly)
    FString Zone = "starter_zone";

    UPROPERTY(BlueprintReadOnly)
    FVector Location = FVector::ZeroVector;

    UPROPERTY(BlueprintReadOnly)
    float Rotation = 0.0f;

    UPROPERTY(BlueprintReadOnly)
    int32 MapLayer = 0;
};

USTRUCT(BlueprintType)
struct FCharacterInfo
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadOnly)
    FString CharacterID;

    UPROPERTY(BlueprintReadOnly)
    FString Name;

    UPROPERTY(BlueprintReadOnly)
    ECharacterClass Class = ECharacterClass::Unspecified;

    UPROPERTY(BlueprintReadOnly)
    ECharacterRace Race = ECharacterRace::Unspecified;

    UPROPERTY(BlueprintReadOnly)
    EGender Gender = EGender::Unspecified;

    UPROPERTY(BlueprintReadOnly)
    int32 Level = 1;

    UPROPERTY(BlueprintReadOnly)
    int64 Experience = 0;

    UPROPERTY(BlueprintReadOnly)
    FCharacterAppearance Appearance;

    UPROPERTY(BlueprintReadOnly)
    FCharacterStats Stats;

    UPROPERTY(BlueprintReadOnly)
    FWorldPosition Position;

    UPROPERTY(BlueprintReadOnly)
    FDateTime LastPlayedAt;

    UPROPERTY(BlueprintReadOnly)
    bool bIsSelected = false;
};

USTRUCT(BlueprintType)
struct FCharacterCreateRequest
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    FString Name;

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    ECharacterClass Class = ECharacterClass::Warrior;

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    ECharacterRace Race = ECharacterRace::Human;

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    EGender Gender = EGender::Male;

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    FCharacterAppearance Appearance;

    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    FString StartingZone = "starter_zone";
};

USTRUCT(BlueprintType)
struct FCharacterListResponse
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadOnly)
    bool bSuccess = false;

    UPROPERTY(BlueprintReadOnly)
    TArray<FCharacterInfo> Characters;

    UPROPERTY(BlueprintReadOnly)
    int32 UsedSlots = 0;

    UPROPERTY(BlueprintReadOnly)
    int32 MaxSlots = 5;

    UPROPERTY(BlueprintReadOnly)
    bool bIsPremium = false;

    UPROPERTY(BlueprintReadOnly)
    FString ErrorMessage;

    UPROPERTY(BlueprintReadOnly)
    FString ErrorCode;
};
```

### Character Validation Rules

```cpp
// CharacterValidation.h
#pragma once

#include "CoreMinimal.h"

class FCharacterValidation
{
public:
    // Name validation
    static bool IsValidCharacterName(const FString& Name, FString& OutError)
    {
        // Length check
        if (Name.Len() < 3 || Name.Len() > 32)
        {
            OutError = "Character name must be between 3 and 32 characters";
            return false;
        }

        // Alphanumeric check
        for (TCHAR C : Name)
        {
            if (!FChar::IsAlnum(C))
            {
                OutError = "Character name can only contain letters and numbers";
                return false;
            }
        }

        // Reserved names check
        static const TArray<FString> ReservedNames = {
            "Admin", "Moderator", "GM", "GameMaster", "System", "Server"
        };

        for (const FString& Reserved : ReservedNames)
        {
            if (Name.Equals(Reserved, ESearchCase::IgnoreCase))
            {
                OutError = "This name is reserved";
                return false;
            }
        }

        return true;
    }

    // Class/Race combination validation
    static bool IsValidClassRaceCombination(ECharacterClass Class, ECharacterRace Race)
    {
        // Define invalid combinations
        if (Class == ECharacterClass::Mage && 
            (Race == ECharacterRace::Orc || Race == ECharacterRace::Tauren))
        {
            return false;
        }

        // Add more rules as needed
        return true;
    }

    // Appearance validation
    static bool IsValidAppearance(const FCharacterAppearance& Appearance, 
                                  ECharacterRace Race, EGender Gender)
    {
        // Validate ranges
        if (Appearance.HairStyle < 0 || Appearance.HairStyle > 20) return false;
        if (Appearance.FaceType < 0 || Appearance.FaceType > 10) return false;
        if (Appearance.BodyType < 0 || Appearance.BodyType > 3) return false;
        if (Appearance.Height < 0.8f || Appearance.Height > 1.2f) return false;

        // Race-specific validations
        if (Race == ECharacterRace::Undead && Appearance.SkinTone.A > 0.5f)
        {
            return false; // Undead should have pale skin
        }

        return true;
    }
};
```

---

## Cache Models (Redis)

```json
// Character list cache
// Key: character:list:{user_id}
{
    "characters": [
        {
            "id": "123e4567-e89b-12d3-a456-426614174000",
            "name": "Gandalf",
            "class": 2,
            "race": 1,
            "gender": 1,
            "level": 60,
            "zone": "eastern_kingdoms",
            "last_played_at": "2025-07-29T10:00:00Z",
            "is_selected": true
        }
    ],
    "slots": {
        "used": 1,
        "max": 5
    },
    "updated_at": "2025-07-29T10:00:00Z"
}

// Selected character cache
// Key: character:selected:{user_id}
{
    "character_id": "123e4567-e89b-12d3-a456-426614174000",
    "character_name": "Gandalf",
    "session_id": "456e7890-e89b-12d3-a456-426614174000"
}

// Name availability cache
// Key: character:name:{normalized_name}
{
    "taken": true,
    "character_id": "123e4567-e89b-12d3-a456-426614174000",
    "checked_at": "2025-07-29T10:00:00Z"
}
```

---

## Event Models (NATS)

```json
// Character created event
{
    "event_type": "character.created",
    "timestamp": "2025-07-29T10:00:00Z",
    "data": {
        "user_id": "789e0123-e89b-12d3-a456-426614174000",
        "character_id": "123e4567-e89b-12d3-a456-426614174000",
        "character_name": "Gandalf",
        "class": 2,
        "race": 1,
        "level": 1
    }
}

// Character selected event
{
    "event_type": "character.selected",
    "timestamp": "2025-07-29T10:00:00Z",
    "data": {
        "user_id": "789e0123-e89b-12d3-a456-426614174000",
        "character_id": "123e4567-e89b-12d3-a456-426614174000",
        "previous_character_id": null,
        "session_id": "456e7890-e89b-12d3-a456-426614174000"
    }
}

// Character deleted event
{
    "event_type": "character.deleted",
    "timestamp": "2025-07-29T10:00:00Z",
    "data": {
        "user_id": "789e0123-e89b-12d3-a456-426614174000",
        "character_id": "123e4567-e89b-12d3-a456-426614174000",
        "character_name": "Gandalf",
        "recovery_token": "abc123def456",
        "permanent_delete_at": "2025-08-28T10:00:00Z"
    }
}
```

---

## Migration Path

### From Phase 1 to Phase 1.5

1. **Database Migration**:
   - Run character table creation scripts
   - Update users table with character fields
   - Populate class/race data

2. **Backend Updates**:
   - Deploy character service
   - Update gateway routing
   - Configure Redis keys

3. **Frontend Updates**:
   - Add character subsystem
   - Update auth flow to include character selection
   - Deploy new UI widgets

### Data Model Versioning

- All API responses include version field
- Proto definitions use proper versioning
- Database migrations are numbered sequentially
- Client maintains backward compatibility

---

## Validation Constants

```go
// Character name constraints
const (
    MinCharacterNameLength = 3
    MaxCharacterNameLength = 32
    CharacterNamePattern   = "^[a-zA-Z0-9]+$"
)

// Character limits
const (
    DefaultMaxCharacters    = 5
    PremiumMaxCharacters    = 10
    MaxCharacterLevel       = 100
    StartingLevel           = 1
    StartingGold            = 0
)

// Appearance ranges
const (
    MaxHairStyles          = 20
    MaxFaceTypes           = 10
    MaxBodyTypes           = 3
    MinHeight              = 0.8
    MaxHeight              = 1.2
)

// Time constraints
const (
    CharacterDeletionGrace = 30 * 24 * time.Hour
    NameReservationTimeout = 30 * time.Minute
)
```

This comprehensive data model specification provides all the necessary structures for implementing the character system across the entire stack.