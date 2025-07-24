#pragma once

#include "CoreMinimal.h"
#include "Subsystems/GameInstanceSubsystem.h"
#include "Http/MMORPGHTTPClient.h"
#include "WebSocket/MMORPGWebSocketClient.h"
#include "MMORPGNetworkSubsystem.generated.h"

/**
 * Network configuration
 */
USTRUCT(BlueprintType)
struct MMORPGNETWORK_API FMMORPGNetworkConfig
{
	GENERATED_BODY()

	// Backend base URL
	UPROPERTY(BlueprintReadWrite, EditAnywhere, Category = "Network")
	FString BackendURL = TEXT("http://localhost:8080");

	// WebSocket URL
	UPROPERTY(BlueprintReadWrite, EditAnywhere, Category = "Network")
	FString WebSocketURL = TEXT("ws://localhost:8080/ws");

	// API version
	UPROPERTY(BlueprintReadWrite, EditAnywhere, Category = "Network")
	FString APIVersion = TEXT("v1");

	// Connection timeout in seconds
	UPROPERTY(BlueprintReadWrite, EditAnywhere, Category = "Network", meta = (ClampMin = "1", ClampMax = "60"))
	float ConnectionTimeout = 10.0f;

	// Reconnect attempts
	UPROPERTY(BlueprintReadWrite, EditAnywhere, Category = "Network", meta = (ClampMin = "0", ClampMax = "10"))
	int32 MaxReconnectAttempts = 3;

	// Reconnect delay in seconds
	UPROPERTY(BlueprintReadWrite, EditAnywhere, Category = "Network", meta = (ClampMin = "0.1", ClampMax = "30"))
	float ReconnectDelay = 2.0f;
};

/**
 * Network subsystem for managing all network operations
 */
UCLASS()
class MMORPGNETWORK_API UMMORPGNetworkSubsystem : public UGameInstanceSubsystem
{
	GENERATED_BODY()

public:
	// USubsystem interface
	virtual void Initialize(FSubsystemCollectionBase& Collection) override;
	virtual void Deinitialize() override;

	// Configuration
	
	/**
	 * Get network configuration
	 * @return Current network configuration
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Network")
	const FMMORPGNetworkConfig& GetNetworkConfig() const { return NetworkConfig; }

	/**
	 * Update network configuration
	 * @param NewConfig New configuration to apply
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Network")
	void SetNetworkConfig(const FMMORPGNetworkConfig& NewConfig);

	// HTTP Operations

	/**
	 * Make an HTTP request
	 * @param Path API path (will be appended to base URL)
	 * @param Verb HTTP verb
	 * @param Headers Additional headers
	 * @param Body Request body
	 * @return Async HTTP request
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Make API Request"))
	UMMORPGHTTPRequest* MakeAPIRequest(
		const FString& Path,
		EMMORPGHTTPVerb Verb,
		const TMap<FString, FString>& Headers,
		const FString& Body
	);

	/**
	 * Get full API URL
	 * @param Path API path
	 * @return Complete URL
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Network")
	FString GetAPIURL(const FString& Path) const;

	// WebSocket Operations

	/**
	 * Get or create the main WebSocket client
	 * @return WebSocket client instance
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Network")
	UMMORPGWebSocketClient* GetWebSocketClient();

	/**
	 * Connect WebSocket with current configuration
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Network")
	void ConnectWebSocket();

	/**
	 * Disconnect WebSocket
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Network")
	void DisconnectWebSocket();

	// Authentication

	/**
	 * Get current auth token
	 * @return Auth token if available
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Network")
	FString GetAuthToken() const { return AuthToken; }

	/**
	 * Set auth token
	 * @param Token New auth token
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Network")
	void SetAuthToken(const FString& Token);

	/**
	 * Clear auth token
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Network")
	void ClearAuthToken();

	/**
	 * Check if authenticated
	 * @return True if auth token is set
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Network")
	bool IsAuthenticated() const { return !AuthToken.IsEmpty(); }

	// Utility

	/**
	 * Get default headers for API requests
	 * @return Map of default headers
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Network")
	TMap<FString, FString> GetDefaultHeaders() const;

protected:
	// Network configuration
	UPROPERTY()
	FMMORPGNetworkConfig NetworkConfig;

	// WebSocket client
	UPROPERTY()
	UMMORPGWebSocketClient* WebSocketClient;

	// Auth token
	FString AuthToken;

	// Reconnection
	FTimerHandle ReconnectTimer;
	int32 ReconnectAttempts;

	// WebSocket callbacks
	UFUNCTION()
	void OnWebSocketConnected();

	UFUNCTION()
	void OnWebSocketConnectionError(const FString& Error);

	UFUNCTION()
	void OnWebSocketClosed(int32 StatusCode, const FString& Reason);

	// Reconnection logic
	void ScheduleReconnect();
	void AttemptReconnect();
};