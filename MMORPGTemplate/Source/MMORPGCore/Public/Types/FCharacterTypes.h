#pragma once

#include "CoreMinimal.h"
#include "Engine/DataTable.h"
#include "Dom/JsonObject.h"
#include "FCharacterTypes.generated.h"

// Character class enum
UENUM(BlueprintType)
enum class ECharacterClass : uint8
{
    None        UMETA(DisplayName = "None"),
    Warrior     UMETA(DisplayName = "Warrior"),
    Mage        UMETA(DisplayName = "Mage"),
    Archer      UMETA(DisplayName = "Archer"),
    Rogue       UMETA(DisplayName = "Rogue"),
    Priest      UMETA(DisplayName = "Priest"),
    Paladin     UMETA(DisplayName = "Paladin")
};

// Character race enum
UENUM(BlueprintType)
enum class ECharacterRace : uint8
{
    None        UMETA(DisplayName = "None"),
    Human       UMETA(DisplayName = "Human"),
    Elf         UMETA(DisplayName = "Elf"),
    Dwarf       UMETA(DisplayName = "Dwarf"),
    Orc         UMETA(DisplayName = "Orc"),
    Undead      UMETA(DisplayName = "Undead")
};

// Character gender enum
UENUM(BlueprintType)
enum class ECharacterGender : uint8
{
    Male        UMETA(DisplayName = "Male"),
    Female      UMETA(DisplayName = "Female"),
    Other       UMETA(DisplayName = "Other")
};

// Character appearance data
USTRUCT(BlueprintType)
struct MMORPGCORE_API FCharacterAppearance
{
    GENERATED_BODY()

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Appearance")
    ECharacterGender Gender = ECharacterGender::Male;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Appearance")
    int32 FaceID = 1;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Appearance")
    int32 HairID = 1;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Appearance")
    FString SkinColor = TEXT("#FFD4B2");

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Appearance")
    FString HairColor = TEXT("#4A3728");

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Appearance")
    FString EyeColor = TEXT("#0066CC");

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Appearance")
    float Height = 1.0f;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Appearance")
    float Build = 1.0f;

    FCharacterAppearance() {}

    // Convert to JSON for API
    FString ToJSON() const;
    bool ParseFromJSON(const TSharedPtr<FJsonObject>& JsonObject);
};

// Character statistics
USTRUCT(BlueprintType)
struct MMORPGCORE_API FCharacterStats
{
    GENERATED_BODY()

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Stats")
    int32 Health = 100;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Stats")
    int32 MaxHealth = 100;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Stats")
    int32 Mana = 50;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Stats")
    int32 MaxMana = 50;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Stats")
    int32 Strength = 10;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Stats")
    int32 Intelligence = 10;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Stats")
    int32 Agility = 10;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Stats")
    int32 Stamina = 10;

    FCharacterStats() {}
};

// Character position in world
USTRUCT(BlueprintType)
struct MMORPGCORE_API FCharacterPosition
{
    GENERATED_BODY()

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Position")
    FString World = TEXT("DefaultWorld");

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Position")
    FVector Location = FVector::ZeroVector;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Position")
    FRotator Rotation = FRotator::ZeroRotator;

    FCharacterPosition() {}
};

// Character info (main data structure)
USTRUCT(BlueprintType)
struct MMORPGCORE_API FCharacterInfo
{
    GENERATED_BODY()

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    FString ID;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    FString UserID;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    FString Name;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    FString Class;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    ECharacterRace Race = ECharacterRace::Human;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    int32 Level = 1;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    int32 ExperiencePoints = 0;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    FCharacterAppearance Appearance;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    FCharacterStats Stats;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    FCharacterPosition Position;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    FDateTime CreatedAt;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    FDateTime LastPlayedAt;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    bool bIsDeleted = false;

    FCharacterInfo() {}

    bool IsValid() const { return !ID.IsEmpty() && !Name.IsEmpty(); }
    bool ParseFromJSON(const TSharedPtr<FJsonObject>& JsonObject);
};

// Character creation request
USTRUCT(BlueprintType)
struct MMORPGCORE_API FCharacterCreateRequest
{
    GENERATED_BODY()

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    FString Name;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    FString Class;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    ECharacterRace Race = ECharacterRace::Human;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    FCharacterAppearance Appearance;

    FCharacterCreateRequest() {}

    FString ToJSON() const;
};

// Character update request
USTRUCT(BlueprintType)
struct MMORPGCORE_API FCharacterUpdateRequest
{
    GENERATED_BODY()

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    FString Name;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Character")
    FCharacterAppearance Appearance;

    FCharacterUpdateRequest() {}

    FString ToJSON() const;
};

// Character list response
USTRUCT(BlueprintType)
struct MMORPGCORE_API FCharacterListResponse
{
    GENERATED_BODY()

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Response")
    bool bSuccess = false;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Response")
    FString ErrorMessage;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Response")
    TArray<FCharacterInfo> Characters;

    FCharacterListResponse() {}

    bool ParseFromJSON(const FString& JsonString);
};

// Character response (single character)
USTRUCT(BlueprintType)
struct MMORPGCORE_API FCharacterResponse
{
    GENERATED_BODY()

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Response")
    bool bSuccess = false;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Response")
    FString ErrorMessage;

    UPROPERTY(EditAnywhere, BlueprintReadWrite, Category = "Response")
    FCharacterInfo Character;

    FCharacterResponse() {}

    bool ParseFromJSON(const FString& JsonString);
};

// Helper functions for enum conversions
MMORPGCORE_API FString CharacterClassToString(ECharacterClass Class);
MMORPGCORE_API ECharacterClass StringToCharacterClass(const FString& ClassString);
MMORPGCORE_API FString CharacterRaceToString(ECharacterRace Race);
MMORPGCORE_API ECharacterRace StringToCharacterRace(const FString& RaceString);
MMORPGCORE_API FString CharacterGenderToString(ECharacterGender Gender);
MMORPGCORE_API ECharacterGender StringToCharacterGender(const FString& GenderString);