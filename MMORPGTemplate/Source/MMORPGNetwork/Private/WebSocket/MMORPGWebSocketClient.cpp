#include "WebSocket/MMORPGWebSocketClient.h"
#include "WebSocketsModule.h"
#include "MMORPGNetwork.h"

UMMORPGWebSocketClient::UMMORPGWebSocketClient()
	: ConnectionState(EWebSocketState::Disconnected)
{
}

UMMORPGWebSocketClient::~UMMORPGWebSocketClient()
{
	Cleanup();
}

void UMMORPGWebSocketClient::Connect(const FString& URL, const FString& Protocol, const TMap<FString, FString>& Headers)
{
	if (ConnectionState != EWebSocketState::Disconnected)
	{
		UE_LOG(LogMMORPGNetwork, Warning, TEXT("WebSocket already connected or connecting"));
		return;
	}

	ServerURL = URL;
	ConnectionState = EWebSocketState::Connecting;

	// Create WebSocket
	if (!FModuleManager::Get().IsModuleLoaded("WebSockets"))
	{
		FModuleManager::Get().LoadModule("WebSockets");
	}

	WebSocket = FWebSocketsModule::Get().CreateWebSocket(URL, Protocol, Headers);

	if (!WebSocket.IsValid())
	{
		ConnectionState = EWebSocketState::Disconnected;
		OnConnectionError.Broadcast(TEXT("Failed to create WebSocket"));
		return;
	}

	// Bind callbacks
	WebSocket->OnConnected().AddUObject(this, &UMMORPGWebSocketClient::HandleOnConnected);
	WebSocket->OnConnectionError().AddUObject(this, &UMMORPGWebSocketClient::HandleOnConnectionError);
	WebSocket->OnClosed().AddUObject(this, &UMMORPGWebSocketClient::HandleOnClosed);
	WebSocket->OnMessage().AddUObject(this, &UMMORPGWebSocketClient::HandleOnMessage);
	WebSocket->OnBinaryMessage().AddUObject(this, &UMMORPGWebSocketClient::HandleOnBinaryMessage);
	WebSocket->OnMessageSent().AddUObject(this, &UMMORPGWebSocketClient::HandleOnMessageSent);

	// Connect
	WebSocket->Connect();

	UE_LOG(LogMMORPGNetwork, Log, TEXT("WebSocket connecting to: %s"), *URL);
}

void UMMORPGWebSocketClient::Disconnect(int32 Code, const FString& Reason)
{
	if (ConnectionState == EWebSocketState::Disconnected)
	{
		return;
	}

	ConnectionState = EWebSocketState::Closing;

	if (WebSocket.IsValid())
	{
		WebSocket->Close(Code, Reason);
	}

	UE_LOG(LogMMORPGNetwork, Log, TEXT("WebSocket disconnecting: Code=%d, Reason=%s"), Code, *Reason);
}

bool UMMORPGWebSocketClient::SendMessage(const FString& Message)
{
	if (!IsConnected() || !WebSocket.IsValid())
	{
		UE_LOG(LogMMORPGNetwork, Warning, TEXT("Cannot send message: WebSocket not connected"));
		return false;
	}

	WebSocket->Send(Message);
	return true;
}

bool UMMORPGWebSocketClient::SendBinaryMessage(const TArray<uint8>& Data)
{
	if (!IsConnected() || !WebSocket.IsValid())
	{
		UE_LOG(LogMMORPGNetwork, Warning, TEXT("Cannot send binary message: WebSocket not connected"));
		return false;
	}

	WebSocket->Send(Data.GetData(), Data.Num(), true);
	return true;
}

void UMMORPGWebSocketClient::HandleOnConnected()
{
	ConnectionState = EWebSocketState::Connected;
	UE_LOG(LogMMORPGNetwork, Log, TEXT("WebSocket connected successfully"));
	OnConnected.Broadcast();
}

void UMMORPGWebSocketClient::HandleOnConnectionError(const FString& Error)
{
	ConnectionState = EWebSocketState::Disconnected;
	UE_LOG(LogMMORPGNetwork, Error, TEXT("WebSocket connection error: %s"), *Error);
	OnConnectionError.Broadcast(Error);
	Cleanup();
}

void UMMORPGWebSocketClient::HandleOnClosed(int32 StatusCode, const FString& Reason, bool bWasClean)
{
	ConnectionState = EWebSocketState::Disconnected;
	UE_LOG(LogMMORPGNetwork, Log, TEXT("WebSocket closed: Code=%d, Reason=%s, Clean=%s"), 
		StatusCode, *Reason, bWasClean ? TEXT("Yes") : TEXT("No"));
	OnClosed.Broadcast(StatusCode, Reason);
	Cleanup();
}

void UMMORPGWebSocketClient::HandleOnMessage(const FString& Message)
{
	UE_LOG(LogMMORPGNetwork, VeryVerbose, TEXT("WebSocket message received: %s"), *Message);
	OnMessageReceived.Broadcast(Message);
}

void UMMORPGWebSocketClient::HandleOnBinaryMessage(const void* Data, SIZE_T Size, bool bIsLastFragment)
{
	TArray<uint8> BinaryData;
	BinaryData.Append(static_cast<const uint8*>(Data), Size);
	
	UE_LOG(LogMMORPGNetwork, VeryVerbose, TEXT("WebSocket binary message received: %d bytes"), Size);
	OnBinaryMessageReceived.Broadcast(BinaryData);
}

void UMMORPGWebSocketClient::HandleOnMessageSent(const FString& Message)
{
	UE_LOG(LogMMORPGNetwork, VeryVerbose, TEXT("WebSocket message sent: %s"), *Message);
}

void UMMORPGWebSocketClient::Cleanup()
{
	if (WebSocket.IsValid())
	{
		WebSocket->OnConnected().Clear();
		WebSocket->OnConnectionError().Clear();
		WebSocket->OnClosed().Clear();
		WebSocket->OnMessage().Clear();
		WebSocket->OnBinaryMessage().Clear();
		WebSocket->OnMessageSent().Clear();
		
		if (WebSocket->IsConnected())
		{
			WebSocket->Close();
		}
		
		WebSocket.Reset();
	}

	ConnectionState = EWebSocketState::Disconnected;
}

// UMMORPGWebSocketManager implementation

UMMORPGWebSocketClient* UMMORPGWebSocketManager::CreateWebSocketClient(UObject* Outer)
{
	if (!Outer)
	{
		return nullptr;
	}

	return NewObject<UMMORPGWebSocketClient>(Outer);
}

bool UMMORPGWebSocketManager::ParseWebSocketURL(const FString& URL, FString& OutProtocol, FString& OutHost, int32& OutPort, FString& OutPath)
{
	// Parse WebSocket URL format: ws://host:port/path or wss://host:port/path
	FString TempURL = URL;

	// Extract protocol
	int32 ProtocolEnd = TempURL.Find(TEXT("://"));
	if (ProtocolEnd == INDEX_NONE)
	{
		return false;
	}

	OutProtocol = TempURL.Left(ProtocolEnd);
	if (OutProtocol != TEXT("ws") && OutProtocol != TEXT("wss"))
	{
		return false;
	}

	TempURL = TempURL.Mid(ProtocolEnd + 3);

	// Extract host and port
	int32 PathStart = TempURL.Find(TEXT("/"));
	FString HostPort;
	
	if (PathStart != INDEX_NONE)
	{
		HostPort = TempURL.Left(PathStart);
		OutPath = TempURL.Mid(PathStart);
	}
	else
	{
		HostPort = TempURL;
		OutPath = TEXT("/");
	}

	// Split host and port
	int32 PortStart = HostPort.Find(TEXT(":"));
	if (PortStart != INDEX_NONE)
	{
		OutHost = HostPort.Left(PortStart);
		FString PortStr = HostPort.Mid(PortStart + 1);
		OutPort = FCString::Atoi(*PortStr);
	}
	else
	{
		OutHost = HostPort;
		OutPort = (OutProtocol == TEXT("wss")) ? 443 : 80;
	}

	return !OutHost.IsEmpty();
}