// Copyright (c) 2024 MMORPG Template Project

#include "Network/MMORPGNetworkManager.h"
#include "MMORPGCore.h"
#include "HttpModule.h"
#include "Interfaces/IHttpRequest.h"
#include "Interfaces/IHttpResponse.h"
#include "Dom/JsonObject.h"
#include "Serialization/JsonSerializer.h"
#include "Serialization/JsonReader.h"
#include "Misc/ConfigCacheIni.h"

FMMORPGNetworkManager::FMMORPGNetworkManager()
    : ServerHost(TEXT("localhost"))
    , ServerPort(8090)
    , ServerProtocol(TEXT("http"))
    , bIsConnected(false)
    , RequestTimeout(30.0f)
    , MaxRetryAttempts(3)
{
    HttpModule = &FHttpModule::Get();
}

FMMORPGNetworkManager::~FMMORPGNetworkManager()
{
    Shutdown();
}

void FMMORPGNetworkManager::Initialize()
{
    // Load configuration
    FString ConfigFile = FPaths::ProjectConfigDir() / TEXT("DefaultMMORPG.ini");
    if (GConfig)
    {
        FString ConfigHost;
        int32 ConfigPort;
        
        if (GConfig->GetString(TEXT("/Script/MMORPGCore.MMORPGSettings"), TEXT("DefaultServerHost"), ConfigHost, ConfigFile))
        {
            ServerHost = ConfigHost;
        }
        
        if (GConfig->GetInt(TEXT("/Script/MMORPGCore.MMORPGSettings"), TEXT("DefaultServerPort"), ConfigPort, ConfigFile))
        {
            ServerPort = ConfigPort;
        }
        
        float ConfigTimeout;
        if (GConfig->GetFloat(TEXT("/Script/MMORPGCore.MMORPGSettings"), TEXT("ConnectionTimeout"), ConfigTimeout, ConfigFile))
        {
            RequestTimeout = ConfigTimeout;
        }
    }
    
    MMORPG_LOG_NET(Log, TEXT("Network Manager initialized - Server: %s:%d"), *ServerHost, ServerPort);
}

void FMMORPGNetworkManager::Shutdown()
{
    if (bIsConnected)
    {
        Disconnect();
    }
    
    MMORPG_LOG_NET(Log, TEXT("Network Manager shutdown"));
}

void FMMORPGNetworkManager::Connect(const FString& Host, int32 Port)
{
    ServerHost = Host;
    ServerPort = Port;
    
    MMORPG_LOG_NET(Log, TEXT("Connecting to server %s:%d"), *ServerHost, ServerPort);
    
    // Test connection with health check
    TestConnection([this](bool bSuccess, const FString& Response)
    {
        if (bSuccess)
        {
            SetConnectionStatus(true);
            MMORPG_LOG_NET(Log, TEXT("Successfully connected to server"));
        }
        else
        {
            SetConnectionStatus(false);
            MMORPG_LOG_NET(Error, TEXT("Failed to connect to server: %s"), *Response);
        }
    });
}

void FMMORPGNetworkManager::Disconnect()
{
    MMORPG_LOG_NET(Log, TEXT("Disconnecting from server"));
    SetConnectionStatus(false);
}

FString FMMORPGNetworkManager::GetServerURL() const
{
    return FString::Printf(TEXT("%s://%s:%d"), *ServerProtocol, *ServerHost, ServerPort);
}

void FMMORPGNetworkManager::SendGetRequest(const FString& Endpoint, 
                                          TFunction<void(bool, const FString&)> Callback)
{
    TSharedRef<IHttpRequest, ESPMode::ThreadSafe> Request = CreateHttpRequest(TEXT("GET"));
    Request->SetURL(GetServerURL() + Endpoint);
    
    Request->OnProcessRequestComplete().BindSP(
        this, 
        &FMMORPGNetworkManager::OnHttpRequestComplete,
        Callback
    );
    
    Request->ProcessRequest();
}

void FMMORPGNetworkManager::SendPostRequest(const FString& Endpoint, 
                                           const TSharedPtr<FJsonObject>& JsonData,
                                           TFunction<void(bool, const FString&)> Callback)
{
    TSharedRef<IHttpRequest, ESPMode::ThreadSafe> Request = CreateHttpRequest(TEXT("POST"));
    Request->SetURL(GetServerURL() + Endpoint);
    Request->SetHeader(TEXT("Content-Type"), TEXT("application/json"));
    
    if (JsonData.IsValid())
    {
        FString OutputString;
        TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&OutputString);
        FJsonSerializer::Serialize(JsonData.ToSharedRef(), Writer);
        Request->SetContentAsString(OutputString);
    }
    
    Request->OnProcessRequestComplete().BindSP(
        this, 
        &FMMORPGNetworkManager::OnHttpRequestComplete,
        Callback
    );
    
    Request->ProcessRequest();
}

void FMMORPGNetworkManager::SendPutRequest(const FString& Endpoint, 
                                          const TSharedPtr<FJsonObject>& JsonData,
                                          TFunction<void(bool, const FString&)> Callback)
{
    TSharedRef<IHttpRequest, ESPMode::ThreadSafe> Request = CreateHttpRequest(TEXT("PUT"));
    Request->SetURL(GetServerURL() + Endpoint);
    Request->SetHeader(TEXT("Content-Type"), TEXT("application/json"));
    
    if (JsonData.IsValid())
    {
        FString OutputString;
        TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&OutputString);
        FJsonSerializer::Serialize(JsonData.ToSharedRef(), Writer);
        Request->SetContentAsString(OutputString);
    }
    
    Request->OnProcessRequestComplete().BindSP(
        this, 
        &FMMORPGNetworkManager::OnHttpRequestComplete,
        Callback
    );
    
    Request->ProcessRequest();
}

void FMMORPGNetworkManager::SendDeleteRequest(const FString& Endpoint,
                                             TFunction<void(bool, const FString&)> Callback)
{
    TSharedRef<IHttpRequest, ESPMode::ThreadSafe> Request = CreateHttpRequest(TEXT("DELETE"));
    Request->SetURL(GetServerURL() + Endpoint);
    
    Request->OnProcessRequestComplete().BindSP(
        this, 
        &FMMORPGNetworkManager::OnHttpRequestComplete,
        Callback
    );
    
    Request->ProcessRequest();
}

void FMMORPGNetworkManager::TestConnection(TFunction<void(bool, const FString&)> Callback)
{
    SendGetRequest(TEXT("/"), Callback);
}

void FMMORPGNetworkManager::GetHealthStatus(TFunction<void(bool, const FString&)> Callback)
{
    SendGetRequest(TEXT("/health"), Callback);
}

TSharedRef<IHttpRequest, ESPMode::ThreadSafe> FMMORPGNetworkManager::CreateHttpRequest(const FString& Verb)
{
    TSharedRef<IHttpRequest, ESPMode::ThreadSafe> Request = HttpModule->CreateRequest();
    Request->SetVerb(Verb);
    Request->SetTimeout(RequestTimeout);
    
    // Add common headers
    Request->SetHeader(TEXT("User-Agent"), TEXT("MMORPG-Template-UE5/1.0"));
    Request->SetHeader(TEXT("Accept"), TEXT("application/json"));
    
    // Add auth token if available
    if (!AuthToken.IsEmpty())
    {
        Request->SetHeader(TEXT("Authorization"), FString::Printf(TEXT("Bearer %s"), *AuthToken));
    }
    
    return Request;
}

void FMMORPGNetworkManager::OnHttpRequestComplete(FHttpRequestPtr Request, 
                                                 FHttpResponsePtr Response, 
                                                 bool bSuccess,
                                                 TFunction<void(bool, const FString&)> Callback)
{
    FString ResponseString;
    bool bRequestSuccess = false;
    
    if (bSuccess && Response.IsValid())
    {
        int32 ResponseCode = Response->GetResponseCode();
        ResponseString = Response->GetContentAsString();
        
        MMORPG_LOG_NET(Verbose, TEXT("HTTP Response [%d]: %s"), ResponseCode, *ResponseString);
        
        if (ResponseCode >= 200 && ResponseCode < 300)
        {
            bRequestSuccess = true;
        }
        else
        {
            // Try to parse error message from JSON
            TSharedPtr<FJsonObject> JsonObject;
            if (ProcessJsonResponse(Response, JsonObject))
            {
                FString ErrorMessage;
                if (JsonObject->TryGetStringField(TEXT("error"), ErrorMessage))
                {
                    ResponseString = ErrorMessage;
                }
            }
            
            OnRequestError.Broadcast(ResponseCode, ResponseString);
        }
    }
    else
    {
        ResponseString = TEXT("Network request failed");
        MMORPG_LOG_NET(Error, TEXT("HTTP Request failed: %s"), *Request->GetURL());
        OnRequestError.Broadcast(0, ResponseString);
    }
    
    OnRequestCompleted.Broadcast(bRequestSuccess);
    
    if (Callback)
    {
        Callback(bRequestSuccess, ResponseString);
    }
}

bool FMMORPGNetworkManager::ProcessJsonResponse(FHttpResponsePtr Response, 
                                               TSharedPtr<FJsonObject>& OutJsonObject)
{
    if (!Response.IsValid())
    {
        return false;
    }
    
    FString ContentType = Response->GetContentType();
    if (!ContentType.Contains(TEXT("application/json")))
    {
        return false;
    }
    
    FString ResponseString = Response->GetContentAsString();
    TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(ResponseString);
    
    return FJsonSerializer::Deserialize(Reader, OutJsonObject);
}

void FMMORPGNetworkManager::SetConnectionStatus(bool bNewStatus)
{
    if (bIsConnected != bNewStatus)
    {
        bIsConnected = bNewStatus;
        OnConnectionStatusChanged.Broadcast(bIsConnected);
    }
}