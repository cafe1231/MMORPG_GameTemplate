#include "Subsystems/UMMORPGAuthSubsystem.h"
#include "GameFramework/GameUserSettings.h"
#include "Dom/JsonObject.h"
#include "Serialization/JsonSerializer.h"
#include "Serialization/JsonReader.h"
#include "Serialization/JsonWriter.h"
#include "Engine/Engine.h"
#include "HttpModule.h"
#include "Interfaces/IHttpResponse.h"
#include "Misc/Paths.h"
#include "Misc/ConfigCacheIni.h"

const FString UMMORPGAuthSubsystem::AuthSaveSlotName = TEXT("MMORPGAuthData");

UMMORPGAuthSubsystem::UMMORPGAuthSubsystem()
{
    ServerURL = TEXT("http://localhost:3000");
}

void UMMORPGAuthSubsystem::Initialize(FSubsystemCollectionBase& Collection)
{
    Super::Initialize(Collection);
    
    // Load saved auth data
    LoadAuthData();
}

void UMMORPGAuthSubsystem::Deinitialize()
{
    Super::Deinitialize();
}

void UMMORPGAuthSubsystem::Login(const FLoginRequest& Request)
{
    // MOCK MODE FOR TESTING - Remove this block when backend is ready
    bool bUseMockMode = true; // Set to false when you have a real backend
    if (bUseMockMode)
    {
        FAuthResponse MockResponse;
        
        // Simulate validation
        if (Request.Email.IsEmpty() || Request.Password.IsEmpty())
        {
            MockResponse.bSuccess = false;
            MockResponse.Message = TEXT("Please enter email and password");
        }
        else if (Request.Email == TEXT("test@test.com") && Request.Password == TEXT("password"))
        {
            MockResponse.bSuccess = true;
            MockResponse.Message = TEXT("Login successful!");
            MockResponse.Tokens.AccessToken = TEXT("mock_access_token");
            MockResponse.Tokens.RefreshToken = TEXT("mock_refresh_token");
            MockResponse.Tokens.ExpiresAt = FDateTime::Now() + FTimespan::FromHours(1);
            
            CurrentTokens = MockResponse.Tokens;
            CurrentUserInfo.Email = Request.Email;
            CurrentUserInfo.Username = TEXT("TestUser");
            CurrentUserInfo.UserId = TEXT("12345");
        }
        else
        {
            MockResponse.bSuccess = false;
            MockResponse.Message = TEXT("Invalid email or password");
        }
        
        // Broadcast the response
        OnLoginResponse.Broadcast(MockResponse);
        if (MockResponse.bSuccess)
        {
            OnUserInfoReceived.Broadcast(CurrentUserInfo);
        }
        
        return;
    }
    // END MOCK MODE
    
    // Create JSON request
    TSharedPtr<FJsonObject> JsonObject = MakeShareable(new FJsonObject);
    JsonObject->SetStringField(TEXT("email"), Request.Email);
    JsonObject->SetStringField(TEXT("password"), Request.Password);
    
    FString OutputString;
    TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&OutputString);
    FJsonSerializer::Serialize(JsonObject.ToSharedRef(), Writer);
    
    // Create and send HTTP request
    TSharedRef<IHttpRequest, ESPMode::ThreadSafe> HttpRequest = CreateHttpRequest(TEXT("POST"), TEXT("/api/auth/login"));
    HttpRequest->SetContentAsString(OutputString);
    HttpRequest->OnProcessRequestComplete().BindUObject(this, &UMMORPGAuthSubsystem::HandleLoginResponse);
    HttpRequest->ProcessRequest();
}

void UMMORPGAuthSubsystem::Register(const FRegisterRequest& Request)
{
    // MOCK MODE FOR TESTING
    bool bUseMockMode = true; // Set to false when you have a real backend
    if (bUseMockMode)
    {
        FAuthResponse MockResponse;
        
        // Simulate validation
        if (Request.Email.IsEmpty() || Request.Password.IsEmpty() || Request.Username.IsEmpty())
        {
            MockResponse.bSuccess = false;
            MockResponse.Message = TEXT("All fields are required");
        }
        else if (Request.Email == TEXT("test@test.com"))
        {
            MockResponse.bSuccess = false;
            MockResponse.Message = TEXT("Email already exists");
        }
        else
        {
            MockResponse.bSuccess = true;
            MockResponse.Message = TEXT("Registration successful! Please login.");
        }
        
        // Broadcast the response
        OnRegisterResponse.Broadcast(MockResponse);
        
        return;
    }
    // END MOCK MODE
    
    // Create JSON request
    TSharedPtr<FJsonObject> JsonObject = MakeShareable(new FJsonObject);
    JsonObject->SetStringField(TEXT("email"), Request.Email);
    JsonObject->SetStringField(TEXT("password"), Request.Password);
    JsonObject->SetStringField(TEXT("username"), Request.Username);
    
    FString OutputString;
    TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&OutputString);
    FJsonSerializer::Serialize(JsonObject.ToSharedRef(), Writer);
    
    // Create and send HTTP request
    TSharedRef<IHttpRequest, ESPMode::ThreadSafe> HttpRequest = CreateHttpRequest(TEXT("POST"), TEXT("/api/auth/register"));
    HttpRequest->SetContentAsString(OutputString);
    HttpRequest->OnProcessRequestComplete().BindUObject(this, &UMMORPGAuthSubsystem::HandleRegisterResponse);
    HttpRequest->ProcessRequest();
}

void UMMORPGAuthSubsystem::Logout()
{
    CurrentTokens = FAuthTokens();
    CurrentUserInfo = FUserInfo();
    SaveAuthData();
}

void UMMORPGAuthSubsystem::RefreshToken()
{
    if (CurrentTokens.RefreshToken.IsEmpty())
    {
        return;
    }
    
    // Create JSON request
    TSharedPtr<FJsonObject> JsonObject = MakeShareable(new FJsonObject);
    JsonObject->SetStringField(TEXT("refreshToken"), CurrentTokens.RefreshToken);
    
    FString OutputString;
    TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&OutputString);
    FJsonSerializer::Serialize(JsonObject.ToSharedRef(), Writer);
    
    // Create and send HTTP request
    TSharedRef<IHttpRequest, ESPMode::ThreadSafe> HttpRequest = CreateHttpRequest(TEXT("POST"), TEXT("/api/auth/refresh"));
    HttpRequest->SetContentAsString(OutputString);
    HttpRequest->OnProcessRequestComplete().BindUObject(this, &UMMORPGAuthSubsystem::HandleRefreshResponse);
    HttpRequest->ProcessRequest();
}

bool UMMORPGAuthSubsystem::IsAuthenticated() const
{
    return !CurrentTokens.AccessToken.IsEmpty() && CurrentTokens.ExpiresAt > FDateTime::Now();
}

void UMMORPGAuthSubsystem::SetServerURL(const FString& URL)
{
    ServerURL = URL;
}

TSharedRef<IHttpRequest, ESPMode::ThreadSafe> UMMORPGAuthSubsystem::CreateHttpRequest(const FString& Verb, const FString& Path)
{
    TSharedRef<IHttpRequest, ESPMode::ThreadSafe> Request = FHttpModule::Get().CreateRequest();
    Request->SetURL(ServerURL + Path);
    Request->SetVerb(Verb);
    Request->SetHeader(TEXT("Content-Type"), TEXT("application/json"));
    Request->SetTimeout(10.0f);
    return Request;
}

void UMMORPGAuthSubsystem::HandleLoginResponse(FHttpRequestPtr Request, FHttpResponsePtr Response, bool bWasSuccessful)
{
    FAuthResponse AuthResponse;
    
    if (bWasSuccessful && Response.IsValid())
    {
        const FString ResponseContent = Response->GetContentAsString();
        const int32 ResponseCode = Response->GetResponseCode();
        
        AuthResponse = ParseAuthResponse(ResponseContent);
        
        if (ResponseCode == 200 && AuthResponse.bSuccess)
        {
            CurrentTokens = AuthResponse.Tokens;
            SaveAuthData();
            
            // Get user info
            if (!CurrentTokens.AccessToken.IsEmpty())
            {
                TSharedRef<IHttpRequest, ESPMode::ThreadSafe> HttpRequest = CreateHttpRequest(TEXT("GET"), TEXT("/api/auth/me"));
                HttpRequest->SetHeader(TEXT("Authorization"), TEXT("Bearer ") + CurrentTokens.AccessToken);
                HttpRequest->OnProcessRequestComplete().BindUObject(this, &UMMORPGAuthSubsystem::HandleUserInfoResponse);
                HttpRequest->ProcessRequest();
            }
        }
    }
    else
    {
        AuthResponse.bSuccess = false;
        AuthResponse.Message = TEXT("Network error");
    }
    
    OnLoginResponse.Broadcast(AuthResponse);
}

void UMMORPGAuthSubsystem::HandleRegisterResponse(FHttpRequestPtr Request, FHttpResponsePtr Response, bool bWasSuccessful)
{
    FAuthResponse AuthResponse;
    
    if (bWasSuccessful && Response.IsValid())
    {
        const FString ResponseContent = Response->GetContentAsString();
        const int32 ResponseCode = Response->GetResponseCode();
        
        AuthResponse = ParseAuthResponse(ResponseContent);
        
        if ((ResponseCode == 200 || ResponseCode == 201) && AuthResponse.bSuccess)
        {
            CurrentTokens = AuthResponse.Tokens;
            SaveAuthData();
        }
    }
    else
    {
        AuthResponse.bSuccess = false;
        AuthResponse.Message = TEXT("Network error");
    }
    
    OnRegisterResponse.Broadcast(AuthResponse);
}

void UMMORPGAuthSubsystem::HandleRefreshResponse(FHttpRequestPtr Request, FHttpResponsePtr Response, bool bWasSuccessful)
{
    if (bWasSuccessful && Response.IsValid())
    {
        const FString ResponseContent = Response->GetContentAsString();
        const int32 ResponseCode = Response->GetResponseCode();
        
        FAuthResponse AuthResponse = ParseAuthResponse(ResponseContent);
        
        if (ResponseCode == 200 && AuthResponse.bSuccess)
        {
            CurrentTokens = AuthResponse.Tokens;
            SaveAuthData();
        }
    }
}

void UMMORPGAuthSubsystem::HandleUserInfoResponse(FHttpRequestPtr Request, FHttpResponsePtr Response, bool bWasSuccessful)
{
    if (bWasSuccessful && Response.IsValid() && Response->GetResponseCode() == 200)
    {
        const FString ResponseContent = Response->GetContentAsString();
        CurrentUserInfo = ParseUserInfo(ResponseContent);
        OnUserInfoReceived.Broadcast(CurrentUserInfo);
    }
}

FAuthResponse UMMORPGAuthSubsystem::ParseAuthResponse(const FString& JsonString)
{
    FAuthResponse Response;
    
    TSharedPtr<FJsonObject> JsonObject;
    TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(JsonString);
    
    if (FJsonSerializer::Deserialize(Reader, JsonObject) && JsonObject.IsValid())
    {
        Response.bSuccess = JsonObject->GetBoolField(TEXT("success"));
        JsonObject->TryGetStringField(TEXT("message"), Response.Message);
        
        const TSharedPtr<FJsonObject>* TokensObjectPtr;
        if (JsonObject->TryGetObjectField(TEXT("tokens"), TokensObjectPtr) && TokensObjectPtr)
        {
            const TSharedPtr<FJsonObject>& TokensObject = *TokensObjectPtr;
            TokensObject->TryGetStringField(TEXT("accessToken"), Response.Tokens.AccessToken);
            TokensObject->TryGetStringField(TEXT("refreshToken"), Response.Tokens.RefreshToken);
            
            FString ExpiresAtStr;
            if (TokensObject->TryGetStringField(TEXT("expiresAt"), ExpiresAtStr))
            {
                FDateTime::ParseIso8601(*ExpiresAtStr, Response.Tokens.ExpiresAt);
            }
        }
    }
    
    return Response;
}

FUserInfo UMMORPGAuthSubsystem::ParseUserInfo(const FString& JsonString)
{
    FUserInfo UserInfo;
    
    TSharedPtr<FJsonObject> JsonObject;
    TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(JsonString);
    
    if (FJsonSerializer::Deserialize(Reader, JsonObject) && JsonObject.IsValid())
    {
        JsonObject->TryGetStringField(TEXT("userId"), UserInfo.UserId);
        JsonObject->TryGetStringField(TEXT("email"), UserInfo.Email);
        JsonObject->TryGetStringField(TEXT("username"), UserInfo.Username);
    }
    
    return UserInfo;
}

void UMMORPGAuthSubsystem::SaveAuthData()
{
    // Save auth tokens using config system
    if (!GConfig) return;
    
    FString ConfigPath = FPaths::ProjectSavedDir() / TEXT("Config/WindowsEditor/AuthData.ini");
    
    GConfig->SetString(TEXT("Auth"), TEXT("AccessToken"), *CurrentTokens.AccessToken, ConfigPath);
    GConfig->SetString(TEXT("Auth"), TEXT("RefreshToken"), *CurrentTokens.RefreshToken, ConfigPath);
    GConfig->SetString(TEXT("Auth"), TEXT("ExpiresAt"), *CurrentTokens.ExpiresAt.ToIso8601(), ConfigPath);
    
    GConfig->Flush(false, ConfigPath);
}

void UMMORPGAuthSubsystem::LoadAuthData()
{
    // Load auth tokens from config
    if (!GConfig) return;
    
    FString ConfigPath = FPaths::ProjectSavedDir() / TEXT("Config/WindowsEditor/AuthData.ini");
    
    GConfig->GetString(TEXT("Auth"), TEXT("AccessToken"), CurrentTokens.AccessToken, ConfigPath);
    GConfig->GetString(TEXT("Auth"), TEXT("RefreshToken"), CurrentTokens.RefreshToken, ConfigPath);
    
    FString ExpiresAtStr;
    GConfig->GetString(TEXT("Auth"), TEXT("ExpiresAt"), ExpiresAtStr, ConfigPath);
    
    if (!ExpiresAtStr.IsEmpty())
    {
        FDateTime::ParseIso8601(*ExpiresAtStr, CurrentTokens.ExpiresAt);
    }
    
    // Check if token is expired and refresh if needed
    if (!CurrentTokens.RefreshToken.IsEmpty() && CurrentTokens.ExpiresAt <= FDateTime::Now())
    {
        RefreshToken();
    }
}