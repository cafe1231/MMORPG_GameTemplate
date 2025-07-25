#pragma once

#include "CoreMinimal.h"
#include "Subsystems/GameInstanceSubsystem.h"
#include "Types/FAuthTypes.h"
#include "Http.h"
#include "UMMORPGAuthSubsystem.generated.h"

DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnAuthResponseDelegate, const FAuthResponse&, Response);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnUserInfoReceivedDelegate, const FUserInfo&, UserInfo);

UCLASS(BlueprintType, Blueprintable)
class MMORPGCORE_API UMMORPGAuthSubsystem : public UGameInstanceSubsystem
{
    GENERATED_BODY()

public:
    UMMORPGAuthSubsystem();

    // Subsystem interface
    virtual void Initialize(FSubsystemCollectionBase& Collection) override;
    virtual void Deinitialize() override;

    // Authentication methods
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth", meta = (DisplayName = "Login"))
    void Login(const FLoginRequest& Request);

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth", meta = (DisplayName = "Register"))
    void Register(const FRegisterRequest& Request);

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth", meta = (DisplayName = "Logout"))
    void Logout();

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth", meta = (DisplayName = "Refresh Token"))
    void RefreshToken();

    // Check if user is authenticated
    UFUNCTION(BlueprintPure, Category = "MMORPG|Auth")
    bool IsAuthenticated() const;

    // Get current auth tokens
    UFUNCTION(BlueprintPure, Category = "MMORPG|Auth")
    FAuthTokens GetAuthTokens() const { return CurrentTokens; }

    // Get current user info
    UFUNCTION(BlueprintPure, Category = "MMORPG|Auth")
    FUserInfo GetUserInfo() const { return CurrentUserInfo; }

    // Set server URL
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    void SetServerURL(const FString& URL);

    // Events
    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Auth")
    FOnAuthResponseDelegate OnLoginResponse;

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Auth")
    FOnAuthResponseDelegate OnRegisterResponse;

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Auth")
    FOnUserInfoReceivedDelegate OnUserInfoReceived;

protected:
    // HTTP response handlers
    void HandleLoginResponse(FHttpRequestPtr Request, FHttpResponsePtr Response, bool bWasSuccessful);
    void HandleRegisterResponse(FHttpRequestPtr Request, FHttpResponsePtr Response, bool bWasSuccessful);
    void HandleRefreshResponse(FHttpRequestPtr Request, FHttpResponsePtr Response, bool bWasSuccessful);
    void HandleUserInfoResponse(FHttpRequestPtr Request, FHttpResponsePtr Response, bool bWasSuccessful);

    // Parse responses
    FAuthResponse ParseAuthResponse(const FString& JsonString);
    FUserInfo ParseUserInfo(const FString& JsonString);

    // Save/Load auth data
    void SaveAuthData();
    void LoadAuthData();

    // HTTP helpers
    TSharedRef<IHttpRequest, ESPMode::ThreadSafe> CreateHttpRequest(const FString& Verb, const FString& Path);

private:
    FAuthTokens CurrentTokens;
    FUserInfo CurrentUserInfo;
    FString ServerURL;

    // Save game slot name
    static const FString AuthSaveSlotName;
};