# 🎮 Phase 1.5: Character System Developer Guide

## Table of Contents

1. [Introduction](#introduction)
2. [Quick Start](#quick-start)
3. [Architecture Overview](#architecture-overview)
4. [Using the Character System](#using-the-character-system)
5. [API Reference](#api-reference)
6. [Customization Guide](#customization-guide)
7. [Best Practices](#best-practices)
8. [Common Patterns](#common-patterns)
9. [Troubleshooting](#troubleshooting)
10. [FAQ](#faq)

---

## Introduction

The character system in Phase 1.5 provides a complete solution for player character management in your MMORPG. This guide will help you understand how to use, extend, and customize the character system for your specific game needs.

### What You'll Learn

- How to integrate the character system into your game
- How to customize character classes, races, and appearance
- How to extend the system with your own features
- Best practices for performance and security
- Common implementation patterns

### Prerequisites

Before using this guide, ensure you have:
- Completed Phase 1 (Authentication) setup
- Basic understanding of Unreal Engine C++ and Blueprints
- Familiarity with Go microservices (for backend customization)
- PostgreSQL and Redis running

---

## Quick Start

### 1. Enable the Character System

After your player authenticates (Phase 1), redirect them to character selection:

**C++ Example**:
```cpp
void AMyGameMode::OnPlayerAuthenticated(const FAuthResponse& Response)
{
    // Get the character subsystem
    UMMORPGCharacterSubsystem* CharSubsystem = GetGameInstance()->GetSubsystem<UMMORPGCharacterSubsystem>();
    
    // Load the player's characters
    CharSubsystem->GetCharacterList();
    
    // Listen for the response
    CharSubsystem->OnCharacterListReceived.AddDynamic(this, &AMyGameMode::HandleCharacterListReceived);
}

void AMyGameMode::HandleCharacterListReceived(const FCharacterListResponse& Response)
{
    if (Response.bSuccess)
    {
        if (Response.Characters.Num() == 0)
        {
            // No characters, show creation screen
            ShowCharacterCreationUI();
        }
        else
        {
            // Show character selection screen
            ShowCharacterSelectionUI(Response.Characters);
        }
    }
}
```

**Blueprint Example**:
1. Get Character Subsystem from Game Instance
2. Call "Get Character List" node
3. Bind to "On Character List Received" event
4. Check if Characters array is empty
5. Show appropriate UI

### 2. Create a Character

**C++ Example**:
```cpp
FCharacterCreateRequest CreateRequest;
CreateRequest.Name = "Gandalf";
CreateRequest.Class = ECharacterClass::Mage;
CreateRequest.Race = ECharacterRace::Human;
CreateRequest.Gender = EGender::Male;

// Set appearance
CreateRequest.Appearance.HairStyle = 3;
CreateRequest.Appearance.HairColor = FLinearColor::White;
CreateRequest.Appearance.FaceType = 2;

// Create the character
CharSubsystem->CreateCharacter(CreateRequest);
```

### 3. Select and Enter Game

```cpp
void UMyCharacterSelectionWidget::OnCharacterClicked(const FString& CharacterID)
{
    CharSubsystem->SelectCharacter(CharacterID);
    CharSubsystem->OnCharacterSelected.AddDynamic(this, &UMyCharacterSelectionWidget::HandleCharacterSelected);
}

void UMyCharacterSelectionWidget::HandleCharacterSelected(const FCharacterInfo& Character)
{
    // Character selected, transition to game world
    UGameplayStatics::OpenLevel(this, "GameWorld");
}
```

---

## Architecture Overview

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (UE5)                         │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │ Character       │  │ Character UI │  │ Character 3D  │  │
│  │ Subsystem       │  │ Widgets      │  │ Preview       │  │
│  │                 │  │              │  │               │  │
│  │ - CRUD Ops      │  │ - Creation   │  │ - Model Load  │  │
│  │ - Validation    │  │ - Selection  │  │ - Customizer  │  │
│  │ - Caching       │  │ - Management │  │ - Animation   │  │
│  └─────────────────┘  └──────────────┘  └───────────────┘  │
│                                                               │
└───────────────────────────┬───────────────────────────────────┘
                            │ HTTP/REST
┌───────────────────────────┴───────────────────────────────────┐
│                        Backend (Go)                            │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │ Character       │  │ Database     │  │ Cache Layer   │  │
│  │ Service         │  │ Repository   │  │ (Redis)       │  │
│  │                 │  │              │  │               │  │
│  │ - Business Logic│  │ - PostgreSQL │  │ - User Lists  │  │
│  │ - Validation    │  │ - Migrations │  │ - Selected    │  │
│  │ - Events        │  │ - Queries    │  │ - Name Cache  │  │
│  └─────────────────┘  └──────────────┘  └───────────────┘  │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

### Data Flow

1. **Character Creation Flow**:
   ```
   UI Input → Validation → API Request → Backend Validation → 
   Database Insert → Cache Update → Event Emission → UI Update
   ```

2. **Character Selection Flow**:
   ```
   Character List → User Selection → API Request → 
   Session Update → Cache Update → Game World Load
   ```

---

## Using the Character System

### Character Subsystem

The `UMMORPGCharacterSubsystem` is your main interface for character operations:

```cpp
// Get the subsystem
UMMORPGCharacterSubsystem* CharSubsystem = GetGameInstance()->GetSubsystem<UMMORPGCharacterSubsystem>();

// Available operations
CharSubsystem->CreateCharacter(CreateRequest);
CharSubsystem->GetCharacterList();
CharSubsystem->GetCharacterDetails(CharacterID);
CharSubsystem->SelectCharacter(CharacterID);
CharSubsystem->DeleteCharacter(CharacterID);
CharSubsystem->ValidateCharacterName(Name);
```

### Character Data Structures

**FCharacterInfo** - Complete character data:
```cpp
USTRUCT(BlueprintType)
struct FCharacterInfo
{
    UPROPERTY(BlueprintReadOnly)
    FString CharacterID;
    
    UPROPERTY(BlueprintReadOnly)
    FString Name;
    
    UPROPERTY(BlueprintReadOnly)
    ECharacterClass Class;
    
    UPROPERTY(BlueprintReadOnly)
    ECharacterRace Race;
    
    UPROPERTY(BlueprintReadOnly)
    int32 Level;
    
    UPROPERTY(BlueprintReadOnly)
    FCharacterAppearance Appearance;
    
    UPROPERTY(BlueprintReadOnly)
    FCharacterStats Stats;
    
    // ... more fields
};
```

### Events and Delegates

Subscribe to character system events:

```cpp
// Success events
CharSubsystem->OnCharacterCreated.AddDynamic(this, &AMyClass::HandleCharacterCreated);
CharSubsystem->OnCharacterListReceived.AddDynamic(this, &AMyClass::HandleCharacterList);
CharSubsystem->OnCharacterSelected.AddDynamic(this, &AMyClass::HandleCharacterSelected);
CharSubsystem->OnCharacterDeleted.AddDynamic(this, &AMyClass::HandleCharacterDeleted);

// Error events
CharSubsystem->OnCharacterError.AddDynamic(this, &AMyClass::HandleCharacterError);
```

### UI Widget Integration

Base character widgets you can extend:

```cpp
// Character Creation Widget
class MMORPGUI_API UMMORPGCharacterCreationWidget : public UMMORPGBaseWidget
{
public:
    // Override these in Blueprint
    UFUNCTION(BlueprintImplementableEvent)
    void OnCreationStepChanged(int32 Step);
    
    UFUNCTION(BlueprintImplementableEvent)
    void OnCharacterPreviewUpdated();
    
protected:
    UFUNCTION(BlueprintCallable)
    void CreateCharacter();
    
    UFUNCTION(BlueprintCallable)
    void UpdatePreview();
};
```

---

## API Reference

### REST Endpoints

#### Create Character
```http
POST /api/v1/characters
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
    "name": "Aragorn",
    "class": "warrior",
    "race": "human",
    "gender": "male",
    "appearance": {
        "hair_style": 5,
        "hair_color": "#4A4A4A",
        "face_type": 3,
        "skin_tone": "#F5DEB3",
        "body_type": 2,
        "height": 1.85
    }
}

Response:
{
    "success": true,
    "character": {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "name": "Aragorn",
        "class": "warrior",
        "race": "human",
        "level": 1,
        "created_at": "2025-07-29T10:00:00Z"
    }
}
```

#### Get Character List
```http
GET /api/v1/characters
Authorization: Bearer {jwt_token}

Response:
{
    "success": true,
    "characters": [
        {
            "id": "123e4567-e89b-12d3-a456-426614174000",
            "name": "Aragorn",
            "class": "warrior",
            "race": "human",
            "level": 60,
            "zone": "rivendell",
            "last_played_at": "2025-07-29T10:00:00Z",
            "is_selected": true
        }
    ],
    "character_slots": {
        "used": 1,
        "max": 5,
        "is_premium": false
    }
}
```

#### Select Character
```http
POST /api/v1/characters/{character_id}/select
Authorization: Bearer {jwt_token}

Response:
{
    "success": true,
    "character": { /* full character data */ }
}
```

### Error Responses

All errors follow this format:
```json
{
    "success": false,
    "error_code": "CHARACTER_NAME_TAKEN",
    "error_message": "This character name is already in use",
    "suggestions": ["Aragorn123", "AragornTheKing", "Aragorn_"]
}
```

Common error codes:
- `CHARACTER_NAME_TAKEN` - Name already exists
- `CHARACTER_NAME_INVALID` - Name contains invalid characters
- `CHARACTER_LIMIT_REACHED` - No available character slots
- `CHARACTER_NOT_FOUND` - Character ID doesn't exist
- `CHARACTER_NOT_OWNED` - Character belongs to another user

---

## Customization Guide

### Adding New Character Classes

1. **Update the Enum** (C++):
```cpp
// CharacterTypes.h
UENUM(BlueprintType)
enum class ECharacterClass : uint8
{
    // Existing classes...
    
    // Add your new class
    Necromancer = 9,
    Monk = 10
};
```

2. **Update Backend Enum** (Go):
```go
// character/types.go
const (
    // Existing classes...
    
    CharacterClassNecromancer CharacterClass = 9
    CharacterClassMonk        CharacterClass = 10
)
```

3. **Add Starting Stats**:
```sql
INSERT INTO class_starting_stats (class, health, mana, strength, intelligence)
VALUES 
    (9, 80, 120, 8, 15),  -- Necromancer
    (10, 100, 80, 12, 10); -- Monk
```

4. **Update UI**:
   - Add class icons to `Content/UI/Icons/Classes/`
   - Update class selection widget
   - Add class descriptions

### Customizing Appearance Options

1. **Extend Appearance Structure**:
```cpp
// In your game's character types
USTRUCT(BlueprintType)
struct FMyGameCharacterAppearance : public FCharacterAppearance
{
    GENERATED_BODY()
    
    // Add custom fields
    UPROPERTY(BlueprintReadWrite)
    int32 AccessoryType;
    
    UPROPERTY(BlueprintReadWrite)
    FLinearColor AccessoryColor;
    
    UPROPERTY(BlueprintReadWrite)
    TArray<int32> Piercings;
};
```

2. **Update Backend Model**:
```go
type MyGameCharacterAppearance struct {
    CharacterAppearance
    
    AccessoryType  int      `json:"accessory_type"`
    AccessoryColor string   `json:"accessory_color"`
    Piercings      []int    `json:"piercings"`
}
```

### Adding Custom Validation

1. **Name Validation**:
```cpp
bool UMyGameCharacterValidator::ValidateCharacterName(const FString& Name, FString& OutError)
{
    // Call base validation first
    if (!Super::ValidateCharacterName(Name, OutError))
    {
        return false;
    }
    
    // Add custom rules
    if (Name.Contains("GM") || Name.Contains("Admin"))
    {
        OutError = "Names cannot contain GM or Admin";
        return false;
    }
    
    // Check against bad words list
    if (IsProfanity(Name))
    {
        OutError = "This name contains inappropriate content";
        return false;
    }
    
    return true;
}
```

2. **Server-Side Validation**:
```go
func (v *MyGameCharacterValidator) ValidateName(name string) error {
    // Base validation
    if err := v.CharacterValidator.ValidateName(name); err != nil {
        return err
    }
    
    // Custom validation
    if strings.Contains(strings.ToLower(name), "gm") {
        return ErrNameContainsReservedWord
    }
    
    return nil
}
```

### Extending Character Stats

1. **Add Custom Stats**:
```cpp
USTRUCT(BlueprintType)
struct FMyGameCharacterStats : public FCharacterStats
{
    GENERATED_BODY()
    
    // Survival game stats
    UPROPERTY(BlueprintReadOnly)
    int32 Hunger = 100;
    
    UPROPERTY(BlueprintReadOnly)
    int32 Thirst = 100;
    
    UPROPERTY(BlueprintReadOnly)
    float Temperature = 37.0f;
    
    // Reputation system
    UPROPERTY(BlueprintReadOnly)
    TMap<FString, int32> FactionReputation;
};
```

2. **Update Database Schema**:
```sql
ALTER TABLE character_stats ADD COLUMN hunger INTEGER DEFAULT 100;
ALTER TABLE character_stats ADD COLUMN thirst INTEGER DEFAULT 100;
ALTER TABLE character_stats ADD COLUMN temperature REAL DEFAULT 37.0;

CREATE TABLE character_faction_reputation (
    character_id UUID REFERENCES characters(id),
    faction_id VARCHAR(64),
    reputation INTEGER DEFAULT 0,
    PRIMARY KEY (character_id, faction_id)
);
```

---

## Best Practices

### Performance Optimization

1. **Character List Caching**:
```cpp
void UMMORPGCharacterSubsystem::GetCharacterList()
{
    // Check cache first
    if (CachedCharacters.Num() > 0 && 
        FDateTime::Now() - LastCacheTime < FTimespan::FromMinutes(5))
    {
        // Use cached data
        OnCharacterListReceived.Broadcast(CreateCachedResponse());
        return;
    }
    
    // Otherwise fetch from server
    FetchCharacterListFromServer();
}
```

2. **Lazy Loading Character Details**:
```cpp
// Only load full character data when needed
void LoadCharacterForGameplay(const FString& CharacterID)
{
    // Basic info already cached from list
    FCharacterInfo* BasicInfo = GetCachedCharacter(CharacterID);
    
    // Load full details including inventory, achievements, etc.
    CharSubsystem->GetCharacterDetails(CharacterID);
}
```

3. **3D Preview Optimization**:
```cpp
// Use LODs for preview models
void UCharacterPreviewWidget::SetupPreviewMesh()
{
    PreviewMesh->SetForcedLodModel(2); // Use lower LOD
    PreviewMesh->bUseAsyncCooking = true;
    PreviewMesh->SetCullDistance(5000.0f); // Don't need far rendering
}
```

### Security Considerations

1. **Always Validate Server-Side**:
```go
func (s *CharacterService) CreateCharacter(ctx context.Context, userID uuid.UUID, req CreateCharacterRequest) (*Character, error) {
    // Verify user hasn't exceeded character limit
    count, err := s.repo.GetCharacterCount(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    maxChars := s.getMaxCharacters(ctx, userID)
    if count >= maxChars {
        return nil, ErrCharacterLimitReached
    }
    
    // Validate all inputs
    if err := s.validator.ValidateCreateRequest(req); err != nil {
        return nil, err
    }
    
    // Additional security checks...
}
```

2. **Rate Limiting**:
```cpp
// Client-side rate limiting (server enforces as well)
bool UMMORPGCharacterSubsystem::CanCreateCharacter() const
{
    if (LastCharacterCreationTime.IsValid())
    {
        float TimeSinceLastCreation = (FDateTime::Now() - LastCharacterCreationTime).GetTotalSeconds();
        if (TimeSinceLastCreation < MinCharacterCreationInterval)
        {
            return false;
        }
    }
    
    return true;
}
```

### Error Handling

1. **Graceful Degradation**:
```cpp
void UMMORPGCharacterSubsystem::HandleCharacterListError(const FString& Error)
{
    UE_LOG(LogMMORPG, Error, TEXT("Failed to load character list: %s"), *Error);
    
    // Try to use cached data if available
    if (CachedCharacters.Num() > 0)
    {
        FCharacterListResponse Response;
        Response.bSuccess = true;
        Response.Characters = CachedCharacters;
        Response.ErrorMessage = "Using cached data - server temporarily unavailable";
        
        OnCharacterListReceived.Broadcast(Response);
    }
    else
    {
        // Show error UI
        OnCharacterError.Broadcast("CHARACTER_LIST_FAILED", Error);
    }
}
```

2. **User-Friendly Messages**:
```cpp
FString GetUserFriendlyError(const FString& ErrorCode)
{
    static TMap<FString, FString> ErrorMessages = {
        {"CHARACTER_NAME_TAKEN", "This name is already taken. Please choose another."},
        {"CHARACTER_LIMIT_REACHED", "You've reached the maximum number of characters. Delete one to create a new character."},
        {"CHARACTER_NAME_INVALID", "Character names must be 3-32 characters and contain only letters and numbers."},
        {"NETWORK_ERROR", "Connection error. Please check your internet and try again."}
    };
    
    if (ErrorMessages.Contains(ErrorCode))
    {
        return ErrorMessages[ErrorCode];
    }
    
    return "An unexpected error occurred. Please try again.";
}
```

---

## Common Patterns

### Pattern 1: Character Creation Wizard

```cpp
UCLASS()
class UCharacterCreationWizard : public UUserWidget
{
public:
    UPROPERTY(BlueprintReadWrite, meta=(BindWidget))
    UWidgetSwitcher* StepSwitcher;
    
    UFUNCTION(BlueprintCallable)
    void NextStep()
    {
        int32 CurrentIndex = StepSwitcher->GetActiveWidgetIndex();
        
        // Validate current step
        if (!ValidateCurrentStep(CurrentIndex))
        {
            return;
        }
        
        // Move to next step
        if (CurrentIndex < StepSwitcher->GetNumWidgets() - 1)
        {
            StepSwitcher->SetActiveWidgetIndex(CurrentIndex + 1);
            OnStepChanged(CurrentIndex + 1);
        }
        else
        {
            // Final step - create character
            SubmitCharacterCreation();
        }
    }
    
private:
    bool ValidateCurrentStep(int32 StepIndex)
    {
        switch (StepIndex)
        {
            case 0: // Name step
                return ValidateCharacterName();
            case 1: // Class/Race step
                return ValidateClassRaceSelection();
            case 2: // Appearance step
                return true; // Always valid
            default:
                return true;
        }
    }
};
```

### Pattern 2: Character Quick Switch

```cpp
// Allow quick character switching without returning to menu
void UQuickSwitchComponent::ShowQuickSwitchUI()
{
    // Get character list from cache
    auto* CharSubsystem = GetGameInstance()->GetSubsystem<UMMORPGCharacterSubsystem>();
    TArray<FCharacterInfo> Characters = CharSubsystem->GetCachedCharacters();
    
    // Filter out current character
    FString CurrentCharID = CharSubsystem->GetSelectedCharacter().CharacterID;
    Characters.RemoveAll([CurrentCharID](const FCharacterInfo& Char) {
        return Char.CharacterID == CurrentCharID;
    });
    
    // Show UI with other characters
    QuickSwitchWidget->SetCharacterList(Characters);
    QuickSwitchWidget->AddToViewport();
}

void UQuickSwitchComponent::OnCharacterSelected(const FString& CharacterID)
{
    // Save current character state
    SaveCurrentCharacterState();
    
    // Switch characters
    CharSubsystem->SelectCharacter(CharacterID);
}
```

### Pattern 3: Character Templates

```cpp
// Provide pre-configured character templates for quick creation
USTRUCT(BlueprintType)
struct FCharacterTemplate
{
    GENERATED_BODY()
    
    UPROPERTY(EditAnywhere)
    FString TemplateName;
    
    UPROPERTY(EditAnywhere)
    FString Description;
    
    UPROPERTY(EditAnywhere)
    ECharacterClass Class;
    
    UPROPERTY(EditAnywhere)
    ECharacterRace Race;
    
    UPROPERTY(EditAnywhere)
    EGender Gender;
    
    UPROPERTY(EditAnywhere)
    FCharacterAppearance Appearance;
    
    UPROPERTY(EditAnywhere)
    UTexture2D* PreviewImage;
};

// In your character creation widget
void ApplyTemplate(const FCharacterTemplate& Template)
{
    CreateRequest.Class = Template.Class;
    CreateRequest.Race = Template.Race;
    CreateRequest.Gender = Template.Gender;
    CreateRequest.Appearance = Template.Appearance;
    
    // Update UI
    UpdateAllFields();
    UpdatePreview();
}
```

---

## Troubleshooting

### Common Issues and Solutions

#### Issue: Character list not loading
**Symptoms**: Empty character list, infinite loading
**Solutions**:
1. Check authentication token is valid
2. Verify character service is running
3. Check network connectivity
4. Look for errors in console logs
5. Clear cache and retry

```cpp
// Debug helper
void DebugCharacterList()
{
    auto* CharSubsystem = GetGameInstance()->GetSubsystem<UMMORPGCharacterSubsystem>();
    
    UE_LOG(LogMMORPG, Warning, TEXT("=== Character Debug Info ==="));
    UE_LOG(LogMMORPG, Warning, TEXT("Authenticated: %s"), 
        CharSubsystem->IsAuthenticated() ? TEXT("Yes") : TEXT("No"));
    UE_LOG(LogMMORPG, Warning, TEXT("Cache Size: %d"), 
        CharSubsystem->GetCachedCharacters().Num());
    UE_LOG(LogMMORPG, Warning, TEXT("Last Error: %s"), 
        *CharSubsystem->GetLastError());
}
```

#### Issue: Character creation fails
**Symptoms**: Error message, character not created
**Solutions**:
1. Validate name meets requirements
2. Check character slot availability
3. Verify class/race combination is valid
4. Ensure appearance values in range
5. Check server logs for details

#### Issue: 3D preview not updating
**Symptoms**: Preview shows default character
**Solutions**:
1. Ensure preview actor is spawned
2. Check material instances are created
3. Verify mesh components are valid
4. Force preview refresh

```cpp
void ForcePreviewRefresh()
{
    if (PreviewActor)
    {
        // Destroy and recreate
        PreviewActor->Destroy();
        PreviewActor = nullptr;
        
        // Recreate with delay
        GetWorld()->GetTimerManager().SetTimerForNextTick([this]() {
            CreatePreviewActor();
            UpdatePreviewAppearance();
        });
    }
}
```

#### Issue: Character selection not persisting
**Symptoms**: Wrong character loaded, selection lost
**Solutions**:
1. Check Redis cache is working
2. Verify session management
3. Ensure selection API call succeeds
4. Check for race conditions

### Debug Console Commands

Enable character system debugging:

```cpp
// Register console commands
void UMMORPGCharacterSubsystem::Initialize(FSubsystemCollectionBase& Collection)
{
    Super::Initialize(Collection);
    
    // Debug commands
    IConsoleManager::Get().RegisterConsoleCommand(
        TEXT("mmorpg.character.list"),
        TEXT("List all cached characters"),
        FConsoleCommandDelegate::CreateUObject(this, &UMMORPGCharacterSubsystem::DebugListCharacters)
    );
    
    IConsoleManager::Get().RegisterConsoleCommand(
        TEXT("mmorpg.character.refresh"),
        TEXT("Force refresh character list"),
        FConsoleCommandDelegate::CreateUObject(this, &UMMORPGCharacterSubsystem::ForceRefreshCharacters)
    );
    
    IConsoleManager::Get().RegisterConsoleCommand(
        TEXT("mmorpg.character.clear"),
        TEXT("Clear character cache"),
        FConsoleCommandDelegate::CreateUObject(this, &UMMORPGCharacterSubsystem::ClearCharacterCache)
    );
}
```

### Logging

Enable detailed logging for troubleshooting:

```ini
; In DefaultEngine.ini
[Core.Log]
LogMMORPGCharacter=Verbose
LogMMORPGUI=Verbose
LogHTTP=Warning
```

Backend logging:
```go
// Enable debug logging
logger.SetLevel(zapcore.DebugLevel)

// Log all character operations
func (s *CharacterService) CreateCharacter(ctx context.Context, userID uuid.UUID, req CreateCharacterRequest) (*Character, error) {
    logger.Debug("Creating character", 
        zap.String("user_id", userID.String()),
        zap.String("name", req.Name),
        zap.Int("class", int(req.Class)),
        zap.Int("race", int(req.Race)))
    
    // ... implementation
}
```

---

## FAQ

### General Questions

**Q: How many characters can a player have?**
A: By default, 5 characters. Premium accounts can have 10. The absolute maximum is 50.

**Q: Can character names be changed?**
A: Not in the base implementation. This feature can be added as a premium service.

**Q: What happens to deleted characters?**
A: Characters are soft-deleted and can be recovered within 30 days. After that, they're permanently removed.

**Q: Can characters be transferred between accounts?**
A: No, this is not supported for security reasons.

### Technical Questions

**Q: How do I add a new character class?**
A: Update the enums in both C++ and Go, add starting stats to the database, and update UI assets. See the [Customization Guide](#customization-guide).

**Q: Can I use my own character models?**
A: Yes! The system is model-agnostic. Update the preview system to load your models based on class/race/appearance.

**Q: How do I implement character progression?**
A: Character level and experience are included. Extend the stats system for your progression mechanics (skills, talents, etc.).

**Q: Is the character system compatible with dedicated servers?**
A: Yes, the system is designed for both listen servers and dedicated servers.

### Performance Questions

**Q: How many concurrent character creations can the system handle?**
A: The default configuration supports 10,000 character creations per hour. This can be scaled horizontally.

**Q: What's the character list loading time?**
A: Average loading time is under 200ms for up to 50 characters, using Redis caching.

**Q: How much bandwidth does character sync use?**
A: Character selection uses about 2-5KB. Full character data with inventory (Phase 3) will use 10-20KB.

### Integration Questions

**Q: Can I use this with Steam authentication?**
A: Yes, as long as Phase 1 authentication provides a valid JWT token, the character system will work.

**Q: Does this work with cloud saves?**
A: The backend already provides cloud storage. Client-side preferences can be added.

**Q: Can I integrate with PlayFab or other BaaS?**
A: Yes, you can replace the backend implementation while keeping the same API contract.

---

## Conclusion

The Phase 1.5 character system provides a solid foundation for your MMORPG's character management needs. By following this guide, you can:

- Implement character creation and management quickly
- Customize the system for your game's unique requirements
- Follow best practices for performance and security
- Troubleshoot common issues effectively

Remember to:
- Always validate on the server side
- Cache data appropriately for performance
- Handle errors gracefully
- Keep the user experience smooth
- Document your customizations

For additional support, refer to the other Phase 1.5 documentation files or reach out to the community.

Happy developing!