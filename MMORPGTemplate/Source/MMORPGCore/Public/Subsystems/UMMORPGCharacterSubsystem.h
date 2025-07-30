#pragma once

#include "CoreMinimal.h"
#include "Subsystems/GameInstanceSubsystem.h"
#include "Engine/GameInstance.h"
#include "UMMORPGCharacterSubsystem.generated.h"

// Forward declarations
class UMMORPGHTTPClient;
class UMMORPGAuthSubsystem;
struct FCharacterInfo;
struct FCharacterCreateRequest;
struct FCharacterUpdateRequest;
struct FCharacterListResponse;
struct FCharacterResponse;

DECLARE_LOG_CATEGORY_EXTERN(LogMMORPGCharacter, Log, All);

// Delegate declarations for character operations
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnCharacterListReceived, const FCharacterListResponse&, Response);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnCharacterCreated, const FCharacterResponse&, Response);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnCharacterSelected, const FString&, CharacterID);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnCharacterDeleted, const FString&, CharacterID);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnCharacterError, const FString&, ErrorMessage);

/**
 * Character subsystem for managing player characters
 * Handles character creation, selection, deletion, and caching
 */
UCLASS()
class MMORPGCORE_API UMMORPGCharacterSubsystem : public UGameInstanceSubsystem
{
    GENERATED_BODY()

public:
    // UGameInstanceSubsystem interface
    virtual void Initialize(FSubsystemCollectionBase& Collection) override;
    virtual void Deinitialize() override;

    // Character operations
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character")
    void GetCharacterList();

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character")
    void CreateCharacter(const FCharacterCreateRequest& Request);

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character")
    void SelectCharacter(const FString& CharacterID);

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character")
    void DeleteCharacter(const FString& CharacterID);

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character")
    void UpdateCharacter(const FString& CharacterID, const FCharacterUpdateRequest& Request);

    // Character state getters
    UFUNCTION(BlueprintPure, Category = "MMORPG|Character")
    const TArray<FCharacterInfo>& GetCachedCharacterList() const { return CachedCharacters; }

    UFUNCTION(BlueprintPure, Category = "MMORPG|Character")
    FString GetSelectedCharacterID() const { return SelectedCharacterID; }

    UFUNCTION(BlueprintPure, Category = "MMORPG|Character")
    bool HasSelectedCharacter() const { return !SelectedCharacterID.IsEmpty(); }

    UFUNCTION(BlueprintPure, Category = "MMORPG|Character")
    int32 GetCharacterCount() const { return CachedCharacters.Num(); }

    UFUNCTION(BlueprintPure, Category = "MMORPG|Character")
    int32 GetMaxCharacterSlots() const { return MaxCharacterSlots; }

    UFUNCTION(BlueprintPure, Category = "MMORPG|Character")
    bool CanCreateMoreCharacters() const { return CachedCharacters.Num() < MaxCharacterSlots; }

    // Find character by ID
    UFUNCTION(BlueprintPure, Category = "MMORPG|Character")
    bool GetCharacterByID(const FString& CharacterID, FCharacterInfo& OutCharacterInfo) const;

    // Mock mode for testing
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character|Debug")
    void SetMockMode(bool bEnable) { bUseMockMode = bEnable; }

    UFUNCTION(BlueprintPure, Category = "MMORPG|Character|Debug")
    bool IsMockMode() const { return bUseMockMode; }

    // Delegates
    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Character")
    FOnCharacterListReceived OnCharacterListReceived;

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Character")
    FOnCharacterCreated OnCharacterCreated;

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Character")
    FOnCharacterSelected OnCharacterSelected;

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Character")
    FOnCharacterDeleted OnCharacterDeleted;

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Character")
    FOnCharacterError OnCharacterError;

protected:
    // Response handlers
    void HandleCharacterListResponse(const FCharacterListResponse& Response);
    void HandleCharacterCreateResponse(const FCharacterResponse& Response);
    void HandleCharacterSelectResponse(const FString& CharacterID, bool bSuccess, const FString& ErrorMessage);
    void HandleCharacterDeleteResponse(const FString& CharacterID, bool bSuccess, const FString& ErrorMessage);
    void HandleCharacterUpdateResponse(const FCharacterResponse& Response);

    // HTTP response callbacks
    UFUNCTION()
    void OnGetCharacterListResponse(bool bWasSuccessful, int32 ResponseCode, const FString& ResponseContent);
    
    UFUNCTION()
    void OnCreateCharacterResponse(bool bWasSuccessful, int32 ResponseCode, const FString& ResponseContent);
    
    UFUNCTION()
    void OnSelectCharacterResponse(bool bWasSuccessful, int32 ResponseCode, const FString& ResponseContent);
    
    UFUNCTION()
    void OnDeleteCharacterResponse(bool bWasSuccessful, int32 ResponseCode, const FString& ResponseContent);

    // Mock implementations for testing
    void MockGetCharacterList();
    void MockCreateCharacter(const FCharacterCreateRequest& Request);
    void MockSelectCharacter(const FString& CharacterID);
    void MockDeleteCharacter(const FString& CharacterID);

    // Utility functions
    void ClearCharacterCache();
    void SaveSelectedCharacter();
    void LoadSelectedCharacter();
    bool ValidateCharacterName(const FString& Name, FString& OutError) const;

private:
    // Dependencies
    UPROPERTY()
    UMMORPGHTTPClient* HTTPClient;

    UPROPERTY()
    UMMORPGAuthSubsystem* AuthSubsystem;

    // Character data
    UPROPERTY()
    TArray<FCharacterInfo> CachedCharacters;

    UPROPERTY()
    FString SelectedCharacterID;

    // Configuration
    UPROPERTY()
    int32 MaxCharacterSlots = 5;

    UPROPERTY()
    bool bUseMockMode = false;

    // Request tracking
    bool bIsRequestInProgress = false;
    
    // Temporary storage for callback context
    FString PendingCharacterID;
};