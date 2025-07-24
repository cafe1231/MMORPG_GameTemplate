#pragma once

#include "CoreMinimal.h"
#include "GameFramework/SaveGame.h"
#include "Types/FAuthTypes.h"
#include "UMMORPGAuthSaveGame.generated.h"

/**
 * Save game class for storing authentication data
 */
UCLASS()
class MMORPGCORE_API UMMORPGAuthSaveGame : public USaveGame
{
    GENERATED_BODY()

public:
    UMMORPGAuthSaveGame();

    /**
     * Refresh token for auto-login
     * Only stored if remember me is enabled
     */
    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString RefreshToken;

    /**
     * User information
     * Stored for quick access without needing to refresh
     */
    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FUserInfo UserInfo;

    /**
     * Whether the user chose to be remembered
     */
    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    bool bRememberMe;

    /**
     * Last login timestamp
     */
    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FDateTime LastLoginTime;

    /**
     * Save game version for future migrations
     */
    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    int32 SaveGameVersion;

    /**
     * Clear all saved data
     */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    void ClearData();

    /**
     * Check if save game has valid data for auto-login
     */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth", BlueprintPure)
    bool HasValidAuthData() const;
};