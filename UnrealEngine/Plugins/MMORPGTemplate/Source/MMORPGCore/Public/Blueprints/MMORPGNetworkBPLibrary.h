// Copyright (c) 2024 MMORPG Template Project

#pragma once

#include "CoreMinimal.h"
#include "Kismet/BlueprintFunctionLibrary.h"
#include "Engine/EngineTypes.h"
#include "MMORPGNetworkBPLibrary.generated.h"

DECLARE_DYNAMIC_DELEGATE_TwoParams(FOnHttpRequestComplete, bool, bSuccess, const FString&, Response);

/**
 * Blueprint function library for MMORPG networking
 * Provides easy-to-use networking functions for Blueprint
 */
UCLASS()
class MMORPGCORE_API UMMORPGNetworkBPLibrary : public UBlueprintFunctionLibrary
{
    GENERATED_BODY()
    
public:
    /** Connect to the MMORPG backend server */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Connect to Server"))
    static void ConnectToServer(const FString& Host = "localhost", int32 Port = 8090);
    
    /** Disconnect from the server */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Disconnect from Server"))
    static void DisconnectFromServer();
    
    /** Check if connected to server */
    UFUNCTION(BlueprintPure, Category = "MMORPG|Network", meta = (DisplayName = "Is Connected"))
    static bool IsConnected();
    
    /** Test server connection */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Test Connection"))
    static void TestConnection(const FOnHttpRequestComplete& OnComplete);
    
    /** Get server health status */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Get Health Status"))
    static void GetHealthStatus(const FOnHttpRequestComplete& OnComplete);
    
    /** Send GET request */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Send GET Request"))
    static void SendGetRequest(const FString& Endpoint, const FOnHttpRequestComplete& OnComplete);
    
    /** Send POST request with JSON data */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Send POST Request"))
    static void SendPostRequest(const FString& Endpoint, const FString& JsonData, const FOnHttpRequestComplete& OnComplete);
    
    /** Set authentication token */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Set Auth Token"))
    static void SetAuthToken(const FString& Token);
    
    /** Get current server URL */
    UFUNCTION(BlueprintPure, Category = "MMORPG|Network", meta = (DisplayName = "Get Server URL"))
    static FString GetServerURL();
    
    /** Simple test API call */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Test API"))
    static void TestAPI(const FOnHttpRequestComplete& OnComplete);
    
    /** Echo test - sends data and receives it back */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Echo Test"))
    static void EchoTest(const FString& Message, const FOnHttpRequestComplete& OnComplete);
};