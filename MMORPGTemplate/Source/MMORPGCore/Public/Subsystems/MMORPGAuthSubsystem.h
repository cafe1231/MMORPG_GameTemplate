#pragma once

#include "CoreMinimal.h"
#include "Subsystems/GameInstanceSubsystem.h"
#include "Types/FAuthTypes.h"
#include "CoreTypes.h"
#include "MMORPGAuthSubsystem.generated.h"

class UMMORPGHTTPClient;
class UMMORPGNetworkSubsystem;
class UMMORPGAuthSaveGame;

/**
 * Authentication subsystem for managing user login, registration, and token management
 */
UCLASS()
class MMORPGCORE_API UMMORPGAuthSubsystem : public UGameInstanceSubsystem
{
    GENERATED_BODY()

public:
    // USubsystem implementation
    virtual void Initialize(FSubsystemCollectionBase& Collection) override;
    virtual void Deinitialize() override;

    /**
     * Login with email and password
     * @param Email User's email address
     * @param Password User's password
     * @param OnComplete Callback when login succeeds
     * @param OnFailed Callback when login fails
     */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth", meta = (DisplayName = "Login"))
    void Login(const FString& Email, const FString& Password, const FOnLoginComplete& OnComplete, const FOnLoginFailed& OnFailed);

    /**
     * Register a new user account
     * @param Email User's email address
     * @param Username User's display name
     * @param Password User's password
     * @param OnComplete Callback when registration succeeds
     * @param OnFailed Callback when registration fails
     */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth", meta = (DisplayName = "Register"))
    void Register(const FString& Email, const FString& Username, const FString& Password, const FOnRegisterComplete& OnComplete, const FOnRegisterFailed& OnFailed);

    /**
     * Logout the current user
     * @param OnComplete Callback when logout completes
     */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth", meta = (DisplayName = "Logout"))
    void Logout(const FOnLogoutComplete& OnComplete);

    /**
     * Refresh the authentication token
     * @param OnComplete Callback when refresh succeeds
     * @param OnFailed Callback when refresh fails
     */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth", meta = (DisplayName = "Refresh Token"))
    void RefreshToken(const FOnRefreshTokenComplete& OnComplete, const FOnRefreshTokenFailed& OnFailed);

    /**
     * Check if user is currently logged in
     * @return True if user has valid authentication token
     */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth", BlueprintPure)
    bool IsLoggedIn() const;

    /**
     * Get the current user's information
     * @return Current user info if logged in, empty struct otherwise
     */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth", BlueprintPure)
    FUserInfo GetCurrentUser() const;

    /**
     * Get the current access token
     * @return Current access token or empty string if not logged in
     */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth", BlueprintPure)
    FString GetAccessToken() const;

    /**
     * Set whether to remember login credentials
     * @param bRemember True to save credentials for auto-login
     */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    void SetRememberMe(bool bRemember);

    /**
     * Try to auto-login with saved credentials
     * @param OnComplete Callback when auto-login succeeds
     * @param OnFailed Callback when auto-login fails
     */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    void TryAutoLogin(const FOnLoginComplete& OnComplete, const FOnLoginFailed& OnFailed);

    // Blueprint events
    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Auth")
    FOnLoginCompleteBP OnLoginCompleteBP;

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Auth")
    FOnLoginFailedBP OnLoginFailedBP;

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Auth")
    FOnRegisterCompleteBP OnRegisterCompleteBP;

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Auth")
    FOnRegisterFailedBP OnRegisterFailedBP;

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Auth")
    FOnLogoutCompleteBP OnLogoutCompleteBP;

protected:
    // Internal functions
    void HandleLoginResponse(const FString& Response, const FOnLoginComplete& OnComplete, const FOnLoginFailed& OnFailed);
    void HandleRegisterResponse(const FString& Response, const FOnRegisterComplete& OnComplete, const FOnRegisterFailed& OnFailed);
    void HandleRefreshResponse(const FString& Response, const FOnRefreshTokenComplete& OnComplete, const FOnRefreshTokenFailed& OnFailed);
    
    void SaveAuthData();
    void LoadAuthData();
    void ClearAuthData();
    
    bool IsTokenExpired() const;
    void ScheduleTokenRefresh();
    void CancelTokenRefresh();

    // Convert structs to/from JSON
    FString LoginRequestToJson(const FLoginRequest& Request) const;
    FString RegisterRequestToJson(const FRegisterRequest& Request) const;
    FString RefreshRequestToJson(const FRefreshTokenRequest& Request) const;
    
    bool ParseLoginResponse(const FString& JsonString, FLoginResponse& OutResponse) const;
    bool ParseRegisterResponse(const FString& JsonString, FRegisterResponse& OutResponse) const;
    bool ParseRefreshResponse(const FString& JsonString, FRefreshTokenResponse& OutResponse) const;
    bool ParseUserInfo(const TSharedPtr<FJsonObject>& JsonObject, FUserInfo& OutUserInfo) const;

private:
    // Cached subsystems
    UPROPERTY()
    UMMORPGNetworkSubsystem* NetworkSubsystem;

    // Current authentication state
    FLoginResponse CurrentAuthData;
    FUserInfo CurrentUser;
    FDateTime TokenExpiryTime;
    bool bIsLoggedIn;
    bool bRememberMe;

    // Token refresh timer
    FTimerHandle TokenRefreshTimer;

    // Save game
    UPROPERTY()
    UMMORPGAuthSaveGame* AuthSaveGame;

    // API endpoints
    static const FString LoginEndpoint;
    static const FString RegisterEndpoint;
    static const FString LogoutEndpoint;
    static const FString RefreshEndpoint;
};