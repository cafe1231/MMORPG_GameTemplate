// Copyright (c) 2024 MMORPG Template Project

#pragma once

#include "CoreMinimal.h"
#include "HttpModule.h"
#include "Interfaces/IHttpRequest.h"
#include "Interfaces/IHttpResponse.h"
#include "Dom/JsonObject.h"
#include "Serialization/JsonSerializer.h"

DECLARE_MULTICAST_DELEGATE_OneParam(FOnConnectionStatusChanged, bool /* bIsConnected */);
DECLARE_MULTICAST_DELEGATE_OneParam(FOnRequestCompleted, bool /* bSuccess */);
DECLARE_MULTICAST_DELEGATE_TwoParams(FOnRequestError, int32 /* ErrorCode */, const FString& /* ErrorMessage */);

/**
 * Network manager for MMORPG Template
 * Handles HTTP and WebSocket communication with backend services
 */
class MMORPGCORE_API FMMORPGNetworkManager : public TSharedFromThis<FMMORPGNetworkManager>
{
public:
    FMMORPGNetworkManager();
    ~FMMORPGNetworkManager();
    
    /** Initialize the network manager */
    void Initialize();
    
    /** Shutdown the network manager */
    void Shutdown();
    
    /** Connect to the backend server */
    void Connect(const FString& Host, int32 Port);
    
    /** Disconnect from the backend server */
    void Disconnect();
    
    /** Check if connected to backend */
    bool IsConnected() const { return bIsConnected; }
    
    /** Get the server URL */
    FString GetServerURL() const;
    
    /** Send a GET request */
    void SendGetRequest(const FString& Endpoint, 
                       TFunction<void(bool, const FString&)> Callback);
    
    /** Send a POST request */
    void SendPostRequest(const FString& Endpoint, 
                        const TSharedPtr<FJsonObject>& JsonData,
                        TFunction<void(bool, const FString&)> Callback);
    
    /** Send a PUT request */
    void SendPutRequest(const FString& Endpoint, 
                       const TSharedPtr<FJsonObject>& JsonData,
                       TFunction<void(bool, const FString&)> Callback);
    
    /** Send a DELETE request */
    void SendDeleteRequest(const FString& Endpoint,
                          TFunction<void(bool, const FString&)> Callback);
    
    /** Test server connection */
    void TestConnection(TFunction<void(bool, const FString&)> Callback);
    
    /** Get server health status */
    void GetHealthStatus(TFunction<void(bool, const FString&)> Callback);
    
    /** Set authentication token */
    void SetAuthToken(const FString& Token) { AuthToken = Token; }
    
    /** Get authentication token */
    FString GetAuthToken() const { return AuthToken; }
    
    /** Events */
    FOnConnectionStatusChanged OnConnectionStatusChanged;
    FOnRequestCompleted OnRequestCompleted;
    FOnRequestError OnRequestError;
    
protected:
    /** Create an HTTP request with common headers */
    TSharedRef<IHttpRequest, ESPMode::ThreadSafe> CreateHttpRequest(const FString& Verb);
    
    /** Handle HTTP request completion */
    void OnHttpRequestComplete(FHttpRequestPtr Request, 
                              FHttpResponsePtr Response, 
                              bool bSuccess,
                              TFunction<void(bool, const FString&)> Callback);
    
    /** Process JSON response */
    bool ProcessJsonResponse(FHttpResponsePtr Response, 
                           TSharedPtr<FJsonObject>& OutJsonObject);
    
    /** Update connection status */
    void SetConnectionStatus(bool bNewStatus);
    
private:
    /** Server configuration */
    FString ServerHost;
    int32 ServerPort;
    FString ServerProtocol;
    
    /** Connection state */
    bool bIsConnected;
    
    /** Authentication */
    FString AuthToken;
    
    /** Request timeout in seconds */
    float RequestTimeout;
    
    /** Maximum retry attempts */
    int32 MaxRetryAttempts;
    
    /** HTTP module reference */
    FHttpModule* HttpModule;
};