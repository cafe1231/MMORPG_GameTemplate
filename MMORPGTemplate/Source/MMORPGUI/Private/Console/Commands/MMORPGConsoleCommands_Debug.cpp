#include "Console/Commands/MMORPGConsoleCommands_Debug.h"
#include "Engine/World.h"
#include "Engine/GameViewportClient.h"
#include "Engine/Engine.h"
#include "GameFramework/GameUserSettings.h"
#include "Subsystems/MMORPGNetworkSubsystem.h"
#include "Subsystems/MMORPGErrorSubsystem.h"
#include "HAL/PlatformMemory.h"
#include "ConsoleSettings.h"

// Show FPS Command

UMMORPGConsoleCommand_ShowFPS::UMMORPGConsoleCommand_ShowFPS()
{
	CommandName = TEXT("showfps");
	Aliases = { TEXT("fps") };
	Description = TEXT("Toggle FPS display");
	Category = TEXT("Debug");
	bAvailableInShipping = false;

	FConsoleCommandParam EnableParam;
	EnableParam.Name = TEXT("enable");
	EnableParam.Type = EConsoleParamType::Boolean;
	EnableParam.bOptional = true;
	EnableParam.DefaultValue = TEXT("toggle");
	EnableParam.Description = TEXT("true/false to enable/disable, or omit to toggle");
	Parameters.Add(EnableParam);
}

FString UMMORPGConsoleCommand_ShowFPS::Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext)
{
	if (!GEngine)
	{
		return TEXT("Error: Engine not available");
	}

	bool bNewState = false;
	
	if (Args.Num() > 0)
	{
		// Parse boolean argument
		FString ParsedValue;
		if (ParseArgumentValue(Args[0], EConsoleParamType::Boolean, ParsedValue))
		{
			bNewState = ParsedValue == TEXT("true");
		}
	}
	else
	{
		// Toggle current state
		bNewState = !GEngine->bEnableOnScreenDebugMessages;
	}

	// Execute console command
	if (bNewState)
	{
		GEngine->Exec(WorldContext ? WorldContext->GetWorld() : nullptr, TEXT("stat fps"));
		return TEXT("FPS display enabled");
	}
	else
	{
		GEngine->Exec(WorldContext ? WorldContext->GetWorld() : nullptr, TEXT("stat none"));
		return TEXT("FPS display disabled");
	}
}

// Set Resolution Command

UMMORPGConsoleCommand_SetResolution::UMMORPGConsoleCommand_SetResolution()
{
	CommandName = TEXT("setres");
	Aliases = { TEXT("resolution") };
	Description = TEXT("Set screen resolution");
	Category = TEXT("Graphics");
	bAvailableInShipping = true;

	FConsoleCommandParam WidthParam;
	WidthParam.Name = TEXT("width");
	WidthParam.Type = EConsoleParamType::Integer;
	WidthParam.bOptional = false;
	WidthParam.Description = TEXT("Screen width in pixels");
	Parameters.Add(WidthParam);

	FConsoleCommandParam HeightParam;
	HeightParam.Name = TEXT("height");
	HeightParam.Type = EConsoleParamType::Integer;
	HeightParam.bOptional = false;
	HeightParam.Description = TEXT("Screen height in pixels");
	Parameters.Add(HeightParam);

	FConsoleCommandParam FullscreenParam;
	FullscreenParam.Name = TEXT("fullscreen");
	FullscreenParam.Type = EConsoleParamType::Boolean;
	FullscreenParam.bOptional = true;
	FullscreenParam.DefaultValue = TEXT("false");
	FullscreenParam.Description = TEXT("Fullscreen mode");
	Parameters.Add(FullscreenParam);
}

FString UMMORPGConsoleCommand_SetResolution::Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext)
{
	int32 Width = FCString::Atoi(*Args[0]);
	int32 Height = FCString::Atoi(*Args[1]);

	if (Width < 640 || Height < 480)
	{
		return TEXT("Error: Resolution too small (minimum 640x480)");
	}

	bool bFullscreen = false;
	if (Args.Num() > 2)
	{
		FString ParsedValue;
		if (ParseArgumentValue(Args[2], EConsoleParamType::Boolean, ParsedValue))
		{
			bFullscreen = ParsedValue == TEXT("true");
		}
	}

	// Set resolution
	UGameUserSettings* Settings = UGameUserSettings::GetGameUserSettings();
	if (Settings)
	{
		Settings->SetScreenResolution(FIntPoint(Width, Height));
		Settings->SetFullscreenMode(bFullscreen ? EWindowMode::Fullscreen : EWindowMode::Windowed);
		Settings->ApplySettings(false);

		return FString::Printf(TEXT("Resolution set to %dx%d (%s)"), 
			Width, Height, bFullscreen ? TEXT("Fullscreen") : TEXT("Windowed"));
	}

	return TEXT("Error: Could not access game settings");
}

// Network Status Command

UMMORPGConsoleCommand_NetStatus::UMMORPGConsoleCommand_NetStatus()
{
	CommandName = TEXT("netstatus");
	Aliases = { TEXT("netstat"), TEXT("network") };
	Description = TEXT("Show network connection status");
	Category = TEXT("Network");
	bAvailableInShipping = false;
}

FString UMMORPGConsoleCommand_NetStatus::Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext)
{
	if (!WorldContext)
	{
		return TEXT("Error: No world context");
	}

	UWorld* World = GEngine->GetWorldFromContextObject(WorldContext, EGetWorldErrorMode::LogAndReturnNull);
	if (!World)
	{
		return TEXT("Error: Invalid world");
	}

	UGameInstance* GameInstance = World->GetGameInstance();
	if (!GameInstance)
	{
		return TEXT("Error: No game instance");
	}

	UMMORPGNetworkSubsystem* NetworkSubsystem = GameInstance->GetSubsystem<UMMORPGNetworkSubsystem>();
	if (!NetworkSubsystem)
	{
		return TEXT("Error: Network subsystem not found");
	}

	FString Output = TEXT("=== Network Status ===\n");
	
	// Get network configuration
	const FMMORPGNetworkConfig& Config = NetworkSubsystem->GetNetworkConfig();
	Output += FString::Printf(TEXT("Backend URL: %s\n"), *Config.BackendURL);
	Output += FString::Printf(TEXT("WebSocket URL: %s\n"), *Config.WebSocketURL);
	Output += FString::Printf(TEXT("API Version: %s\n"), *Config.APIVersion);
	Output += FString::Printf(TEXT("Authenticated: %s\n"), 
		NetworkSubsystem->IsAuthenticated() ? TEXT("Yes") : TEXT("No"));

	// Check WebSocket status
	UMMORPGWebSocketClient* WebSocket = NetworkSubsystem->GetWebSocketClient();
	if (WebSocket)
	{
		Output += FString::Printf(TEXT("WebSocket Status: %s\n"),
			*UEnum::GetValueAsString(WebSocket->GetConnectionState()));
		Output += FString::Printf(TEXT("WebSocket URL: %s\n"), *WebSocket->GetServerURL());
	}
	else
	{
		Output += TEXT("WebSocket: Not initialized\n");
	}

	return Output;
}

// Memory Stats Command

UMMORPGConsoleCommand_MemStats::UMMORPGConsoleCommand_MemStats()
{
	CommandName = TEXT("memstats");
	Aliases = { TEXT("memory"), TEXT("mem") };
	Description = TEXT("Show memory statistics");
	Category = TEXT("Debug");
	bAvailableInShipping = false;
}

FString UMMORPGConsoleCommand_MemStats::Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext)
{
	FPlatformMemoryStats Stats = FPlatformMemory::GetStats();

	FString Output = TEXT("=== Memory Statistics ===\n");
	Output += FString::Printf(TEXT("Total Physical: %.2f GB\n"), Stats.TotalPhysical / (1024.0 * 1024.0 * 1024.0));
	Output += FString::Printf(TEXT("Available Physical: %.2f GB\n"), Stats.AvailablePhysical / (1024.0 * 1024.0 * 1024.0));
	Output += FString::Printf(TEXT("Used Physical: %.2f GB\n"), Stats.UsedPhysical / (1024.0 * 1024.0 * 1024.0));
	Output += FString::Printf(TEXT("Peak Used Physical: %.2f GB\n"), Stats.PeakUsedPhysical / (1024.0 * 1024.0 * 1024.0));
	Output += TEXT("\n");
	Output += FString::Printf(TEXT("Total Virtual: %.2f GB\n"), Stats.TotalVirtual / (1024.0 * 1024.0 * 1024.0));
	Output += FString::Printf(TEXT("Available Virtual: %.2f GB\n"), Stats.AvailableVirtual / (1024.0 * 1024.0 * 1024.0));
	Output += FString::Printf(TEXT("Used Virtual: %.2f GB\n"), Stats.UsedVirtual / (1024.0 * 1024.0 * 1024.0));
	Output += FString::Printf(TEXT("Peak Used Virtual: %.2f GB\n"), Stats.PeakUsedVirtual / (1024.0 * 1024.0 * 1024.0));

	return Output;
}

// List CVars Command

UMMORPGConsoleCommand_ListCVars::UMMORPGConsoleCommand_ListCVars()
{
	CommandName = TEXT("listcvars");
	Aliases = { TEXT("cvars") };
	Description = TEXT("List console variables matching a pattern");
	Category = TEXT("Debug");
	bAvailableInShipping = false;

	FConsoleCommandParam PatternParam;
	PatternParam.Name = TEXT("pattern");
	PatternParam.Type = EConsoleParamType::String;
	PatternParam.bOptional = true;
	PatternParam.DefaultValue = TEXT("");
	PatternParam.Description = TEXT("Pattern to match (e.g. 'r.', 'stat.')");
	Parameters.Add(PatternParam);

	FConsoleCommandParam LimitParam;
	LimitParam.Name = TEXT("limit");
	LimitParam.Type = EConsoleParamType::Integer;
	LimitParam.bOptional = true;
	LimitParam.DefaultValue = TEXT("20");
	LimitParam.Description = TEXT("Maximum number of results");
	Parameters.Add(LimitParam);
}

FString UMMORPGConsoleCommand_ListCVars::Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext)
{
	FString Pattern = Args.Num() > 0 ? Args[0] : TEXT("");
	int32 Limit = Args.Num() > 1 ? FCString::Atoi(*Args[1]) : 20;

	if (Pattern.IsEmpty())
	{
		return TEXT("Please specify a pattern to search for (e.g., 'r.' for rendering cvars)");
	}

	FString Output = FString::Printf(TEXT("=== Console Variables matching '%s' ===\n"), *Pattern);
	
	// Note: In a real implementation, we would iterate through registered console variables
	// For now, we'll show a message indicating the feature
	Output += TEXT("Note: Full CVar listing requires engine modification.\n");
	Output += TEXT("Use the built-in console (`) to access CVars directly.\n");
	Output += TEXT("\nCommon patterns:\n");
	Output += TEXT("  r.       - Rendering commands\n");
	Output += TEXT("  stat.    - Statistics commands\n");
	Output += TEXT("  t.       - Threading commands\n");
	Output += TEXT("  net.     - Networking commands\n");

	return Output;
}