#pragma once

#include "CoreMinimal.h"
#include "IWebSocket.h"
#include "Engine/GameInstance.h"
#include "MMORPGWebSocketClient.generated.h"

// Forward declarations
class IWebSocket;

DECLARE_DYNAMIC_MULTICAST_DELEGATE(FOnWebSocketConnected);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnWebSocketConnectionError, const FString&, Error);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_TwoParams(FOnWebSocketClosed, int32, StatusCode, const FString&, Reason);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnWebSocketMessageReceived, const FString&, Message);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnWebSocketBinaryMessageReceived, const TArray<uint8>&, Data);

/**
 * WebSocket connection state
 */
UENUM(BlueprintType)
enum class EWebSocketState : uint8
{
	Disconnected	UMETA(DisplayName = "Disconnected"),
	Connecting		UMETA(DisplayName = "Connecting"),
	Connected		UMETA(DisplayName = "Connected"),
	Closing			UMETA(DisplayName = "Closing")
};

/**
 * WebSocket client for real-time communication
 */
UCLASS(BlueprintType)
class MMORPGNETWORK_API UMMORPGWebSocketClient : public UObject
{
	GENERATED_BODY()

public:
	// Constructor
	UMMORPGWebSocketClient();

	// Destructor
	virtual ~UMMORPGWebSocketClient();

	// Events
	UPROPERTY(BlueprintAssignable, Category = "MMORPG|WebSocket")
	FOnWebSocketConnected OnConnected;

	UPROPERTY(BlueprintAssignable, Category = "MMORPG|WebSocket")
	FOnWebSocketConnectionError OnConnectionError;

	UPROPERTY(BlueprintAssignable, Category = "MMORPG|WebSocket")
	FOnWebSocketClosed OnClosed;

	UPROPERTY(BlueprintAssignable, Category = "MMORPG|WebSocket")
	FOnWebSocketMessageReceived OnMessageReceived;

	UPROPERTY(BlueprintAssignable, Category = "MMORPG|WebSocket")
	FOnWebSocketBinaryMessageReceived OnBinaryMessageReceived;

	/**
	 * Connect to WebSocket server
	 * @param URL The WebSocket URL (ws:// or wss://)
	 * @param Protocol Optional protocol
	 * @param Headers Optional headers
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|WebSocket")
	void Connect(const FString& URL, const FString& Protocol = TEXT(""), const TMap<FString, FString>& Headers = TMap<FString, FString>());

	/**
	 * Disconnect from WebSocket server
	 * @param Code Close status code
	 * @param Reason Close reason
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|WebSocket")
	void Disconnect(int32 Code = 1000, const FString& Reason = TEXT("Normal Closure"));

	/**
	 * Send text message
	 * @param Message The message to send
	 * @return True if message was sent
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|WebSocket")
	bool SendMessage(const FString& Message);

	/**
	 * Send binary message
	 * @param Data The data to send
	 * @return True if message was sent
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|WebSocket")
	bool SendBinaryMessage(const TArray<uint8>& Data);

	/**
	 * Get connection state
	 * @return Current connection state
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|WebSocket")
	EWebSocketState GetConnectionState() const { return ConnectionState; }

	/**
	 * Check if connected
	 * @return True if connected
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|WebSocket")
	bool IsConnected() const { return ConnectionState == EWebSocketState::Connected; }

	/**
	 * Get server URL
	 * @return The server URL
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|WebSocket")
	FString GetServerURL() const { return ServerURL; }

protected:
	// WebSocket instance
	TSharedPtr<IWebSocket> WebSocket;

	// Connection state
	EWebSocketState ConnectionState;

	// Server URL
	FString ServerURL;

	// Internal callbacks
	void HandleOnConnected();
	void HandleOnConnectionError(const FString& Error);
	void HandleOnClosed(int32 StatusCode, const FString& Reason, bool bWasClean);
	void HandleOnMessage(const FString& Message);
	void HandleOnBinaryMessage(const void* Data, SIZE_T Size, bool bIsLastFragment);
	void HandleOnMessageSent(const FString& Message);

	// Cleanup
	void Cleanup();
};

/**
 * WebSocket manager for Blueprint use
 */
UCLASS()
class MMORPGNETWORK_API UMMORPGWebSocketManager : public UBlueprintFunctionLibrary
{
	GENERATED_BODY()

public:
	/**
	 * Create a new WebSocket client
	 * @param Outer The outer object
	 * @return New WebSocket client instance
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|WebSocket", meta = (DisplayName = "Create WebSocket Client"))
	static UMMORPGWebSocketClient* CreateWebSocketClient(UObject* Outer);

	/**
	 * Parse WebSocket URL
	 * @param URL The URL to parse
	 * @param OutProtocol Output protocol (ws or wss)
	 * @param OutHost Output host
	 * @param OutPort Output port
	 * @param OutPath Output path
	 * @return True if URL is valid
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|WebSocket", meta = (DisplayName = "Parse WebSocket URL"))
	static bool ParseWebSocketURL(const FString& URL, FString& OutProtocol, FString& OutHost, int32& OutPort, FString& OutPath);
};