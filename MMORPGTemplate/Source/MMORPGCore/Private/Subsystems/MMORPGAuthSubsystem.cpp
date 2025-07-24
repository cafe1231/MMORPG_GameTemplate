#include "Subsystems/MMORPGAuthSubsystem.h"

// NOTE: This implementation temporarily disables HTTP functionality to resolve circular dependencies.
// The NetworkSubsystem needs to be updated with proper HTTP client methods before this can be fully implemented.
// TODO: Implement proper HTTP requests once NetworkSubsystem API is finalized.

#include "SaveGame/MMORPGAuthSaveGame.h"
#include "Engine/GameInstance.h"
#include "TimerManager.h"
#include "Kismet/GameplayStatics.h"
#include "Dom/JsonObject.h"
#include "Serialization/JsonSerializer.h"
#include "Serialization/JsonWriter.h"

// API Endpoints
const FString UMMORPGAuthSubsystem::LoginEndpoint = TEXT("/api/v1/auth/login");
const FString UMMORPGAuthSubsystem::RegisterEndpoint = TEXT("/api/v1/auth/register");
const FString UMMORPGAuthSubsystem::LogoutEndpoint = TEXT("/api/v1/auth/logout");
const FString UMMORPGAuthSubsystem::RefreshEndpoint = TEXT("/api/v1/auth/refresh");

void UMMORPGAuthSubsystem::Initialize(FSubsystemCollectionBase& Collection)
{
    Super::Initialize(Collection);
    
    UE_LOG(LogTemp, Log, TEXT("MMORPGAuthSubsystem: Initializing"));
    
    // Get network subsystem
    // TODO: Re-enable once circular dependency is resolved
    NetworkSubsystem = nullptr; // GetGameInstance()->GetSubsystem<UMMORPGNetworkSubsystem>();
    if (!NetworkSubsystem)
    {
        UE_LOG(LogTemp, Warning, TEXT("MMORPGAuthSubsystem: NetworkSubsystem integration disabled - HTTP functionality pending"));
    }
    
    // Initialize state
    bIsLoggedIn = false;
    bRememberMe = false;
    
    // Load saved auth data
    LoadAuthData();
}

void UMMORPGAuthSubsystem::Deinitialize()
{
    CancelTokenRefresh();
    Super::Deinitialize();
}

void UMMORPGAuthSubsystem::Login(const FString& Email, const FString& Password, const FOnLoginComplete& OnComplete, const FOnLoginFailed& OnFailed)
{
    if (!NetworkSubsystem)
    {
        FMMORPGError Error(1001, "Network subsystem not available", EMMORPGErrorCategory::Network);
        OnFailed.ExecuteIfBound(Error);
        OnLoginFailedBP.Broadcast(Error);
        return;
    }
    
    // Create login request
    FLoginRequest Request;
    Request.Email = Email;
    Request.Password = Password;
    
    // Convert to JSON
    FString JsonBody = LoginRequestToJson(Request);
    
    // TODO: Implement HTTP request using NetworkSubsystem->MakeAPIRequest
    // For now, we'll return an error to avoid build issues
    FMMORPGError Error(5001, "Auth system not fully implemented - HTTP client integration pending", EMMORPGErrorCategory::System);
    OnFailed.ExecuteIfBound(Error);
    OnLoginFailedBP.Broadcast(Error);
    
    // Original code to be reimplemented:
    // HTTPClient->Post(LoginEndpoint, JsonBody, ...);
}

void UMMORPGAuthSubsystem::Register(const FString& Email, const FString& Username, const FString& Password, const FOnRegisterComplete& OnComplete, const FOnRegisterFailed& OnFailed)
{
    if (!NetworkSubsystem)
    {
        FMMORPGError Error(1001, "Network subsystem not available", EMMORPGErrorCategory::Network);
        OnFailed.ExecuteIfBound(Error);
        OnRegisterFailedBP.Broadcast(Error);
        return;
    }
    
    // Create register request
    FRegisterRequest Request;
    Request.Email = Email;
    Request.Username = Username;
    Request.Password = Password;
    
    // Convert to JSON
    FString JsonBody = RegisterRequestToJson(Request);
    
    // TODO: Implement HTTP request using NetworkSubsystem->MakeAPIRequest
    // For now, we'll return an error to avoid build issues
    FMMORPGError Error(5001, "Auth system not fully implemented - HTTP client integration pending", EMMORPGErrorCategory::System);
    OnFailed.ExecuteIfBound(Error);
    OnRegisterFailedBP.Broadcast(Error);
    
    // Original code to be reimplemented:
    // HTTPClient->Post(RegisterEndpoint, JsonBody, ...);
}

void UMMORPGAuthSubsystem::Logout(const FOnLogoutComplete& OnComplete)
{
    if (!bIsLoggedIn)
    {
        OnComplete.ExecuteIfBound();
        OnLogoutCompleteBP.Broadcast();
        return;
    }
    
    if (!NetworkSubsystem)
    {
        // Even if network is unavailable, clear local auth data
        ClearAuthData();
        OnComplete.ExecuteIfBound();
        OnLogoutCompleteBP.Broadcast();
        return;
    }
    
    // TODO: Implement HTTP request using NetworkSubsystem->MakeAPIRequest
    // For now, just clear local data
    ClearAuthData();
    OnComplete.ExecuteIfBound();
    OnLogoutCompleteBP.Broadcast();
    
    // Original code to be reimplemented:
    // HTTPClient->Post(LogoutEndpoint, TEXT("{}"), ...);
}

void UMMORPGAuthSubsystem::RefreshToken(const FOnRefreshTokenComplete& OnComplete, const FOnRefreshTokenFailed& OnFailed)
{
    if (!bIsLoggedIn || CurrentAuthData.RefreshToken.IsEmpty())
    {
        FMMORPGError Error(2001, "No refresh token available", EMMORPGErrorCategory::Auth);
        OnFailed.ExecuteIfBound(Error);
        return;
    }
    
    if (!NetworkSubsystem)
    {
        FMMORPGError Error(1001, "Network subsystem not available", EMMORPGErrorCategory::Network);
        OnFailed.ExecuteIfBound(Error);
        return;
    }
    
    // Create refresh request
    FRefreshTokenRequest Request;
    Request.RefreshToken = CurrentAuthData.RefreshToken;
    
    // Convert to JSON
    FString JsonBody = RefreshRequestToJson(Request);
    
    // TODO: Implement HTTP request using NetworkSubsystem->MakeAPIRequest
    // For now, we'll return an error to avoid build issues
    FMMORPGError Error(5001, "Auth system not fully implemented - HTTP client integration pending", EMMORPGErrorCategory::System);
    OnFailed.ExecuteIfBound(Error);
    
    // Original code to be reimplemented:
    // HTTPClient->Post(RefreshEndpoint, JsonBody, ...);
}

bool UMMORPGAuthSubsystem::IsLoggedIn() const
{
    return bIsLoggedIn && !IsTokenExpired();
}

FUserInfo UMMORPGAuthSubsystem::GetCurrentUser() const
{
    return CurrentUser;
}

FString UMMORPGAuthSubsystem::GetAccessToken() const
{
    return CurrentAuthData.AccessToken;
}

void UMMORPGAuthSubsystem::SetRememberMe(bool bRemember)
{
    bRememberMe = bRemember;
    if (bIsLoggedIn)
    {
        SaveAuthData();
    }
}

void UMMORPGAuthSubsystem::TryAutoLogin(const FOnLoginComplete& OnComplete, const FOnLoginFailed& OnFailed)
{
    if (!AuthSaveGame || AuthSaveGame->RefreshToken.IsEmpty())
    {
        FMMORPGError Error(2002, "No saved credentials found", EMMORPGErrorCategory::Auth);
        OnFailed.ExecuteIfBound(Error);
        return;
    }
    
    // Use refresh token to login
    CurrentAuthData.RefreshToken = AuthSaveGame->RefreshToken;
    RefreshToken(
        [this, OnComplete](const FRefreshTokenResponse& Response)
        {
            // Convert refresh response to login response
            FLoginResponse LoginResponse;
            LoginResponse.AccessToken = Response.AccessToken;
            LoginResponse.RefreshToken = Response.RefreshToken;
            LoginResponse.ExpiresIn = Response.ExpiresIn;
            LoginResponse.User = CurrentUser; // Use saved user info
            
            OnComplete.ExecuteIfBound(LoginResponse);
            OnLoginCompleteBP.Broadcast(LoginResponse);
        },
        [OnFailed, this](const FMMORPGError& Error)
        {
            OnFailed.ExecuteIfBound(Error);
            OnLoginFailedBP.Broadcast(Error);
        }
    );
}

void UMMORPGAuthSubsystem::HandleLoginResponse(const FString& Response, const FOnLoginComplete& OnComplete, const FOnLoginFailed& OnFailed)
{
    FLoginResponse LoginResponse;
    if (!ParseLoginResponse(Response, LoginResponse))
    {
        FMMORPGError Error(3001, "Failed to parse login response", EMMORPGErrorCategory::Protocol);
        OnFailed.ExecuteIfBound(Error);
        OnLoginFailedBP.Broadcast(Error);
        return;
    }
    
    // Update auth state
    CurrentAuthData = LoginResponse;
    CurrentUser = LoginResponse.User;
    bIsLoggedIn = true;
    TokenExpiryTime = FDateTime::Now() + FTimespan::FromSeconds(LoginResponse.ExpiresIn);
    
    // Update network subsystem with token
    if (NetworkSubsystem)
    {
        // TODO: NetworkSubsystem->SetAuthToken(LoginResponse.AccessToken);
    }
    
    // Save if remember me is enabled
    if (bRememberMe)
    {
        SaveAuthData();
    }
    
    // Schedule token refresh
    ScheduleTokenRefresh();
    
    // Notify success
    OnComplete.ExecuteIfBound(LoginResponse);
    OnLoginCompleteBP.Broadcast(LoginResponse);
}

void UMMORPGAuthSubsystem::HandleRegisterResponse(const FString& Response, const FOnRegisterComplete& OnComplete, const FOnRegisterFailed& OnFailed)
{
    FRegisterResponse RegisterResponse;
    if (!ParseRegisterResponse(Response, RegisterResponse))
    {
        FMMORPGError Error(3002, "Failed to parse register response", EMMORPGErrorCategory::Protocol);
        OnFailed.ExecuteIfBound(Error);
        OnRegisterFailedBP.Broadcast(Error);
        return;
    }
    
    // Notify success
    OnComplete.ExecuteIfBound(RegisterResponse);
    OnRegisterCompleteBP.Broadcast(RegisterResponse);
}

void UMMORPGAuthSubsystem::HandleRefreshResponse(const FString& Response, const FOnRefreshTokenComplete& OnComplete, const FOnRefreshTokenFailed& OnFailed)
{
    FRefreshTokenResponse RefreshResponse;
    if (!ParseRefreshResponse(Response, RefreshResponse))
    {
        FMMORPGError Error(3003, "Failed to parse refresh response", EMMORPGErrorCategory::Protocol);
        OnFailed.ExecuteIfBound(Error);
        
        // Clear auth data on refresh failure
        ClearAuthData();
        return;
    }
    
    // Update tokens
    CurrentAuthData.AccessToken = RefreshResponse.AccessToken;
    CurrentAuthData.RefreshToken = RefreshResponse.RefreshToken;
    CurrentAuthData.ExpiresIn = RefreshResponse.ExpiresIn;
    TokenExpiryTime = FDateTime::Now() + FTimespan::FromSeconds(RefreshResponse.ExpiresIn);
    
    // Update network subsystem with new token
    if (NetworkSubsystem)
    {
        // TODO: NetworkSubsystem->SetAuthToken(RefreshResponse.AccessToken);
    }
    
    // Save if remember me is enabled
    if (bRememberMe)
    {
        SaveAuthData();
    }
    
    // Reschedule token refresh
    ScheduleTokenRefresh();
    
    // Notify success
    OnComplete.ExecuteIfBound(RefreshResponse);
}

void UMMORPGAuthSubsystem::SaveAuthData()
{
    if (!AuthSaveGame)
    {
        AuthSaveGame = Cast<UMMORPGAuthSaveGame>(UGameplayStatics::CreateSaveGameObject(UMMORPGAuthSaveGame::StaticClass()));
    }
    
    if (AuthSaveGame)
    {
        AuthSaveGame->RefreshToken = bRememberMe ? CurrentAuthData.RefreshToken : TEXT("");
        AuthSaveGame->UserInfo = CurrentUser;
        AuthSaveGame->bRememberMe = bRememberMe;
        
        UGameplayStatics::SaveGameToSlot(AuthSaveGame, TEXT("AuthSaveGame"), 0);
    }
}

void UMMORPGAuthSubsystem::LoadAuthData()
{
    AuthSaveGame = Cast<UMMORPGAuthSaveGame>(UGameplayStatics::LoadGameFromSlot(TEXT("AuthSaveGame"), 0));
    if (AuthSaveGame)
    {
        bRememberMe = AuthSaveGame->bRememberMe;
        CurrentUser = AuthSaveGame->UserInfo;
    }
}

void UMMORPGAuthSubsystem::ClearAuthData()
{
    bIsLoggedIn = false;
    CurrentAuthData = FLoginResponse();
    CurrentUser = FUserInfo();
    TokenExpiryTime = FDateTime::MinValue();
    
    // Clear network token
    if (NetworkSubsystem)
    {
        // TODO: NetworkSubsystem->ClearAuthToken();
    }
    
    // Clear saved data
    if (AuthSaveGame)
    {
        AuthSaveGame->RefreshToken = TEXT("");
        AuthSaveGame->UserInfo = FUserInfo();
        UGameplayStatics::SaveGameToSlot(AuthSaveGame, TEXT("AuthSaveGame"), 0);
    }
    
    // Cancel token refresh
    CancelTokenRefresh();
}

bool UMMORPGAuthSubsystem::IsTokenExpired() const
{
    return FDateTime::Now() >= TokenExpiryTime;
}

void UMMORPGAuthSubsystem::ScheduleTokenRefresh()
{
    CancelTokenRefresh();
    
    if (GetGameInstance() && GetGameInstance()->GetTimerManager().IsTimerActive(TokenRefreshTimer))
    {
        return;
    }
    
    // Schedule refresh 1 minute before expiry
    float RefreshDelay = FMath::Max(1.0f, CurrentAuthData.ExpiresIn - 60.0f);
    
    if (GetGameInstance())
    {
        GetGameInstance()->GetTimerManager().SetTimer(TokenRefreshTimer,
            [this]()
            {
                RefreshToken(
                    [](const FRefreshTokenResponse& Response)
                    {
                        UE_LOG(LogTemp, Log, TEXT("MMORPGAuthSubsystem: Token refreshed successfully"));
                    },
                    [](const FMMORPGError& Error)
                    {
                        UE_LOG(LogTemp, Warning, TEXT("MMORPGAuthSubsystem: Token refresh failed: %s"), *Error.Message);
                    }
                );
            },
            RefreshDelay,
            false
        );
    }
}

void UMMORPGAuthSubsystem::CancelTokenRefresh()
{
    if (GetGameInstance() && TokenRefreshTimer.IsValid())
    {
        GetGameInstance()->GetTimerManager().ClearTimer(TokenRefreshTimer);
    }
}

FString UMMORPGAuthSubsystem::LoginRequestToJson(const FLoginRequest& Request) const
{
    TSharedPtr<FJsonObject> JsonObject = MakeShareable(new FJsonObject);
    JsonObject->SetStringField(TEXT("email"), Request.Email);
    JsonObject->SetStringField(TEXT("password"), Request.Password);
    
    FString OutputString;
    TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&OutputString);
    FJsonSerializer::Serialize(JsonObject.ToSharedRef(), Writer);
    
    return OutputString;
}

FString UMMORPGAuthSubsystem::RegisterRequestToJson(const FRegisterRequest& Request) const
{
    TSharedPtr<FJsonObject> JsonObject = MakeShareable(new FJsonObject);
    JsonObject->SetStringField(TEXT("email"), Request.Email);
    JsonObject->SetStringField(TEXT("username"), Request.Username);
    JsonObject->SetStringField(TEXT("password"), Request.Password);
    
    FString OutputString;
    TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&OutputString);
    FJsonSerializer::Serialize(JsonObject.ToSharedRef(), Writer);
    
    return OutputString;
}

FString UMMORPGAuthSubsystem::RefreshRequestToJson(const FRefreshTokenRequest& Request) const
{
    TSharedPtr<FJsonObject> JsonObject = MakeShareable(new FJsonObject);
    JsonObject->SetStringField(TEXT("refresh_token"), Request.RefreshToken);
    
    FString OutputString;
    TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&OutputString);
    FJsonSerializer::Serialize(JsonObject.ToSharedRef(), Writer);
    
    return OutputString;
}

bool UMMORPGAuthSubsystem::ParseLoginResponse(const FString& JsonString, FLoginResponse& OutResponse) const
{
    TSharedPtr<FJsonObject> JsonObject;
    TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(JsonString);
    
    if (!FJsonSerializer::Deserialize(Reader, JsonObject) || !JsonObject.IsValid())
    {
        return false;
    }
    
    JsonObject->TryGetStringField(TEXT("access_token"), OutResponse.AccessToken);
    JsonObject->TryGetStringField(TEXT("refresh_token"), OutResponse.RefreshToken);
    JsonObject->TryGetNumberField(TEXT("expires_in"), OutResponse.ExpiresIn);
    
    // Parse user object
    const TSharedPtr<FJsonObject>* UserObject;
    if (JsonObject->TryGetObjectField(TEXT("user"), UserObject))
    {
        ParseUserInfo(*UserObject, OutResponse.User);
    }
    
    return !OutResponse.AccessToken.IsEmpty();
}

bool UMMORPGAuthSubsystem::ParseRegisterResponse(const FString& JsonString, FRegisterResponse& OutResponse) const
{
    TSharedPtr<FJsonObject> JsonObject;
    TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(JsonString);
    
    if (!FJsonSerializer::Deserialize(Reader, JsonObject) || !JsonObject.IsValid())
    {
        return false;
    }
    
    JsonObject->TryGetStringField(TEXT("message"), OutResponse.Message);
    
    // Parse user object
    const TSharedPtr<FJsonObject>* UserObject;
    if (JsonObject->TryGetObjectField(TEXT("user"), UserObject))
    {
        ParseUserInfo(*UserObject, OutResponse.User);
    }
    
    return true;
}

bool UMMORPGAuthSubsystem::ParseRefreshResponse(const FString& JsonString, FRefreshTokenResponse& OutResponse) const
{
    TSharedPtr<FJsonObject> JsonObject;
    TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(JsonString);
    
    if (!FJsonSerializer::Deserialize(Reader, JsonObject) || !JsonObject.IsValid())
    {
        return false;
    }
    
    JsonObject->TryGetStringField(TEXT("access_token"), OutResponse.AccessToken);
    JsonObject->TryGetStringField(TEXT("refresh_token"), OutResponse.RefreshToken);
    JsonObject->TryGetNumberField(TEXT("expires_in"), OutResponse.ExpiresIn);
    
    return !OutResponse.AccessToken.IsEmpty();
}

bool UMMORPGAuthSubsystem::ParseUserInfo(const TSharedPtr<FJsonObject>& JsonObject, FUserInfo& OutUserInfo) const
{
    if (!JsonObject.IsValid())
    {
        return false;
    }
    
    JsonObject->TryGetStringField(TEXT("id"), OutUserInfo.Id);
    JsonObject->TryGetStringField(TEXT("email"), OutUserInfo.Email);
    JsonObject->TryGetStringField(TEXT("username"), OutUserInfo.Username);
    
    // Parse created_at timestamp
    FString CreatedAtString;
    if (JsonObject->TryGetStringField(TEXT("created_at"), CreatedAtString))
    {
        FDateTime::ParseIso8601(*CreatedAtString, OutUserInfo.CreatedAt);
    }
    
    return !OutUserInfo.Id.IsEmpty();
}