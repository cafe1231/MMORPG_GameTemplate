#include "Subsystems/MMORPGNetworkSubsystem.h"
#include "Engine/World.h"
#include "TimerManager.h"
#include "MMORPGNetwork.h"

void UMMORPGNetworkSubsystem::Initialize(FSubsystemCollectionBase& Collection)
{
	Super::Initialize(Collection);

	UE_LOG(LogMMORPGNetwork, Log, TEXT("MMORPGNetworkSubsystem initialized"));

	// Load configuration from settings if available
	// For now, use defaults
}

void UMMORPGNetworkSubsystem::Deinitialize()
{
	// Clean up WebSocket
	DisconnectWebSocket();

	// Clear timers
	if (GetGameInstance() && GetGameInstance()->GetWorld())
	{
		GetGameInstance()->GetWorld()->GetTimerManager().ClearTimer(ReconnectTimer);
	}

	Super::Deinitialize();

	UE_LOG(LogMMORPGNetwork, Log, TEXT("MMORPGNetworkSubsystem deinitialized"));
}

void UMMORPGNetworkSubsystem::SetNetworkConfig(const FMMORPGNetworkConfig& NewConfig)
{
	NetworkConfig = NewConfig;
	
	// If WebSocket is connected with old config, reconnect with new
	if (WebSocketClient && WebSocketClient->IsConnected())
	{
		DisconnectWebSocket();
		ConnectWebSocket();
	}

	UE_LOG(LogMMORPGNetwork, Log, TEXT("Network configuration updated: Backend=%s, WebSocket=%s"), 
		*NetworkConfig.BackendURL, *NetworkConfig.WebSocketURL);
}

UMMORPGHTTPRequest* UMMORPGNetworkSubsystem::MakeAPIRequest(
	const FString& Path,
	EMMORPGHTTPVerb Verb,
	const TMap<FString, FString>& Headers,
	const FString& Body)
{
	if (!GetGameInstance())
	{
		return nullptr;
	}

	FString FullURL = GetAPIURL(Path);
	TMap<FString, FString> AllHeaders = GetDefaultHeaders();
	
	// Merge custom headers
	for (const auto& Header : Headers)
	{
		AllHeaders.Add(Header.Key, Header.Value);
	}

	return UMMORPGHTTPRequest::MakeHTTPRequest(
		GetGameInstance(),
		FullURL,
		Verb,
		AllHeaders,
		Body
	);
}

FString UMMORPGNetworkSubsystem::GetAPIURL(const FString& Path) const
{
	FString CleanPath = Path;
	
	// Ensure path starts with /
	if (!CleanPath.StartsWith(TEXT("/")))
	{
		CleanPath = TEXT("/") + CleanPath;
	}

	// Build full URL
	return FString::Printf(TEXT("%s/api/%s%s"), 
		*NetworkConfig.BackendURL, 
		*NetworkConfig.APIVersion,
		*CleanPath);
}

UMMORPGWebSocketClient* UMMORPGNetworkSubsystem::GetWebSocketClient()
{
	if (!WebSocketClient)
	{
		WebSocketClient = NewObject<UMMORPGWebSocketClient>(this);
		
		// Bind callbacks
		WebSocketClient->OnConnected.AddDynamic(this, &UMMORPGNetworkSubsystem::OnWebSocketConnected);
		WebSocketClient->OnConnectionError.AddDynamic(this, &UMMORPGNetworkSubsystem::OnWebSocketConnectionError);
		WebSocketClient->OnClosed.AddDynamic(this, &UMMORPGNetworkSubsystem::OnWebSocketClosed);
	}

	return WebSocketClient;
}

void UMMORPGNetworkSubsystem::ConnectWebSocket()
{
	if (!GetWebSocketClient())
	{
		return;
	}

	// Reset reconnect attempts
	ReconnectAttempts = 0;

	// Build headers
	TMap<FString, FString> Headers = GetDefaultHeaders();

	// Connect
	WebSocketClient->Connect(NetworkConfig.WebSocketURL, TEXT(""), Headers);
}

void UMMORPGNetworkSubsystem::DisconnectWebSocket()
{
	// Clear reconnect timer
	if (GetGameInstance() && GetGameInstance()->GetWorld())
	{
		GetGameInstance()->GetWorld()->GetTimerManager().ClearTimer(ReconnectTimer);
	}

	// Disconnect
	if (WebSocketClient)
	{
		WebSocketClient->Disconnect();
	}
}

void UMMORPGNetworkSubsystem::SetAuthToken(const FString& Token)
{
	AuthToken = Token;
	UE_LOG(LogMMORPGNetwork, Log, TEXT("Auth token updated"));

	// Reconnect WebSocket with new auth
	if (WebSocketClient && WebSocketClient->IsConnected())
	{
		DisconnectWebSocket();
		ConnectWebSocket();
	}
}

void UMMORPGNetworkSubsystem::ClearAuthToken()
{
	AuthToken.Empty();
	UE_LOG(LogMMORPGNetwork, Log, TEXT("Auth token cleared"));
}

TMap<FString, FString> UMMORPGNetworkSubsystem::GetDefaultHeaders() const
{
	TMap<FString, FString> Headers;
	
	// Content type
	Headers.Add(TEXT("Content-Type"), TEXT("application/json"));
	
	// API version
	Headers.Add(TEXT("X-API-Version"), NetworkConfig.APIVersion);
	
	// Auth token if available
	if (!AuthToken.IsEmpty())
	{
		Headers.Add(TEXT("Authorization"), UMMORPGHTTPClient::CreateAuthHeader(AuthToken));
	}

	return Headers;
}

void UMMORPGNetworkSubsystem::OnWebSocketConnected()
{
	ReconnectAttempts = 0;
	UE_LOG(LogMMORPGNetwork, Log, TEXT("WebSocket connected successfully"));
}

void UMMORPGNetworkSubsystem::OnWebSocketConnectionError(const FString& Error)
{
	UE_LOG(LogMMORPGNetwork, Error, TEXT("WebSocket connection error: %s"), *Error);
	ScheduleReconnect();
}

void UMMORPGNetworkSubsystem::OnWebSocketClosed(int32 StatusCode, const FString& Reason)
{
	UE_LOG(LogMMORPGNetwork, Log, TEXT("WebSocket closed: Code=%d, Reason=%s"), StatusCode, *Reason);
	
	// Schedule reconnect unless it was a normal closure
	if (StatusCode != 1000)
	{
		ScheduleReconnect();
	}
}

void UMMORPGNetworkSubsystem::ScheduleReconnect()
{
	if (ReconnectAttempts >= NetworkConfig.MaxReconnectAttempts)
	{
		UE_LOG(LogMMORPGNetwork, Warning, TEXT("Max reconnect attempts reached (%d)"), NetworkConfig.MaxReconnectAttempts);
		return;
	}

	if (!GetGameInstance() || !GetGameInstance()->GetWorld())
	{
		return;
	}

	ReconnectAttempts++;
	
	float Delay = NetworkConfig.ReconnectDelay * FMath::Pow(2.0f, ReconnectAttempts - 1);
	
	UE_LOG(LogMMORPGNetwork, Log, TEXT("Scheduling reconnect attempt %d/%d in %.1f seconds"),
		ReconnectAttempts, NetworkConfig.MaxReconnectAttempts, Delay);

	GetGameInstance()->GetWorld()->GetTimerManager().SetTimer(
		ReconnectTimer,
		this,
		&UMMORPGNetworkSubsystem::AttemptReconnect,
		Delay,
		false
	);
}

void UMMORPGNetworkSubsystem::AttemptReconnect()
{
	UE_LOG(LogMMORPGNetwork, Log, TEXT("Attempting to reconnect WebSocket"));
	ConnectWebSocket();
}