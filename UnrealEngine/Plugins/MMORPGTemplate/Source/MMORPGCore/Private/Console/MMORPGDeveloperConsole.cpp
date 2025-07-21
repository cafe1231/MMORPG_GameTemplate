// Copyright (c) 2024 MMORPG Template Project

#include "Console/MMORPGDeveloperConsole.h"
#include "MMORPGCore.h"
#include "Network/MMORPGNetworkManager.h"
#include "Engine/Engine.h"
#include "Engine/World.h"
#include "GameFramework/PlayerController.h"
#include "GameFramework/GameModeBase.h"
#include "GameFramework/GameStateBase.h"
#include "Kismet/KismetSystemLibrary.h"
#include "HAL/PlatformFilemanager.h"
#include "Misc/DateTime.h"

DEFINE_LOG_CATEGORY(LogMMORPGConsole);

UMMORPGDeveloperConsole::UMMORPGDeveloperConsole()
    : bIsVisible(false)
    , MaxHistorySize(100)
    , MaxOutputLines(500)
    , CurrentHistoryIndex(-1)
    , ConsoleWidget(nullptr)
{
}

void UMMORPGDeveloperConsole::Initialize()
{
    RegisterBuiltInCommands();
    
    // Load console widget class
    static ConstructorHelpers::FClassFinder<UMMORPGConsoleWidget> ConsoleWidgetClassFinder(
        TEXT("/MMORPGTemplate/UI/WBP_DeveloperConsole"));
    if (ConsoleWidgetClassFinder.Succeeded())
    {
        ConsoleWidgetClass = ConsoleWidgetClassFinder.Class;
    }
    
    WriteOutput(TEXT("MMORPG Developer Console initialized"), FColor::Green);
    WriteOutput(TEXT("Type 'help' for available commands"), FColor::Yellow);
    
    UE_LOG(LogMMORPGConsole, Log, TEXT("Developer console initialized"));
}

void UMMORPGDeveloperConsole::Shutdown()
{
    RegisteredCommands.Empty();
    
    if (ConsoleWidget && ConsoleWidget->IsInViewport())
    {
        ConsoleWidget->RemoveFromParent();
    }
    
    ConsoleWidget = nullptr;
    
    UE_LOG(LogMMORPGConsole, Log, TEXT("Developer console shutdown"));
}

void UMMORPGDeveloperConsole::ToggleConsole()
{
    if (bIsVisible)
    {
        HideConsole();
    }
    else
    {
        ShowConsole();
    }
}

void UMMORPGDeveloperConsole::ShowConsole()
{
    if (bIsVisible)
    {
        return;
    }
    
    UWorld* World = GEngine->GetWorldFromContextObjectChecked(this);
    if (!World)
    {
        return;
    }
    
    APlayerController* PC = World->GetFirstPlayerController();
    if (!PC)
    {
        return;
    }
    
    // Create console widget if needed
    if (!ConsoleWidget && ConsoleWidgetClass)
    {
        ConsoleWidget = CreateWidget<UMMORPGConsoleWidget>(PC, ConsoleWidgetClass);
        if (ConsoleWidget)
        {
            ConsoleWidget->SetConsole(this);
        }
    }
    
    if (ConsoleWidget && !ConsoleWidget->IsInViewport())
    {
        ConsoleWidget->AddToViewport(1000); // High Z-order
        ConsoleWidget->SetKeyboardFocus();
        
        // Set input mode
        FInputModeGameAndUI InputMode;
        InputMode.SetWidgetToFocus(ConsoleWidget->TakeWidget());
        InputMode.SetLockMouseToViewportBehavior(EMouseLockMode::DoNotLock);
        PC->SetInputMode(InputMode);
        PC->bShowMouseCursor = true;
    }
    
    bIsVisible = true;
    UE_LOG(LogMMORPGConsole, Log, TEXT("Console shown"));
}

void UMMORPGDeveloperConsole::HideConsole()
{
    if (!bIsVisible)
    {
        return;
    }
    
    if (ConsoleWidget && ConsoleWidget->IsInViewport())
    {
        ConsoleWidget->RemoveFromParent();
        
        // Restore input mode
        UWorld* World = GEngine->GetWorldFromContextObjectChecked(this);
        if (World)
        {
            APlayerController* PC = World->GetFirstPlayerController();
            if (PC)
            {
                FInputModeGameOnly InputMode;
                PC->SetInputMode(InputMode);
                PC->bShowMouseCursor = false;
            }
        }
    }
    
    bIsVisible = false;
    CurrentHistoryIndex = -1;
    
    UE_LOG(LogMMORPGConsole, Log, TEXT("Console hidden"));
}

void UMMORPGDeveloperConsole::ExecuteCommand(const FString& Command)
{
    if (Command.IsEmpty())
    {
        return;
    }
    
    // Add to history
    AddToHistory(Command);
    
    // Echo command to output
    WriteOutput(FString::Printf(TEXT("> %s"), *Command), FColor::Cyan);
    
    // Parse command
    FString CommandName;
    TArray<FString> Args;
    ParseCommandLine(Command, CommandName, Args);
    
    // Execute command
    if (RegisteredCommands.Contains(CommandName))
    {
        const FCommandInfo& CmdInfo = RegisteredCommands[CommandName];
        CmdInfo.Handler(Args);
        OnCommandExecuted.Broadcast(CommandName, Args);
    }
    else
    {
        WriteOutput(FString::Printf(TEXT("Unknown command: %s"), *CommandName), FColor::Red);
        WriteOutput(TEXT("Type 'help' for available commands"), FColor::Yellow);
    }
}

void UMMORPGDeveloperConsole::AddToHistory(const FString& Command)
{
    // Don't add duplicate consecutive commands
    if (CommandHistory.Num() > 0 && CommandHistory.Last() == Command)
    {
        return;
    }
    
    CommandHistory.Add(Command);
    
    // Trim history if needed
    if (CommandHistory.Num() > MaxHistorySize)
    {
        CommandHistory.RemoveAt(0);
    }
    
    // Save to config
    SaveConfig();
}

void UMMORPGDeveloperConsole::WriteOutput(const FString& Message, const FColor& Color)
{
    FString Timestamp = FDateTime::Now().ToString(TEXT("[%H:%M:%S]"));
    FString FormattedMessage = FString::Printf(TEXT("%s %s"), *Timestamp, *Message);
    
    ConsoleOutput.Add(FormattedMessage);
    
    // Trim output if needed
    if (ConsoleOutput.Num() > MaxOutputLines)
    {
        ConsoleOutput.RemoveAt(0);
    }
    
    // Update widget if visible
    if (ConsoleWidget)
    {
        ConsoleWidget->AddOutputLine(FormattedMessage, Color);
    }
    
    // Also log to file
    UE_LOG(LogMMORPGConsole, Log, TEXT("%s"), *Message);
}

void UMMORPGDeveloperConsole::ClearOutput()
{
    ConsoleOutput.Empty();
    
    if (ConsoleWidget)
    {
        ConsoleWidget->ClearOutput();
    }
    
    WriteOutput(TEXT("Console cleared"), FColor::Green);
}

void UMMORPGDeveloperConsole::RegisterCommand(const FString& Command, const FString& Description, 
                                             TFunction<void(const TArray<FString>&)> Handler)
{
    if (Command.IsEmpty() || !Handler)
    {
        return;
    }
    
    FCommandInfo CmdInfo;
    CmdInfo.Description = Description;
    CmdInfo.Handler = Handler;
    
    RegisteredCommands.Add(Command.ToLower(), CmdInfo);
    
    UE_LOG(LogMMORPGConsole, Log, TEXT("Registered command: %s"), *Command);
}

void UMMORPGDeveloperConsole::UnregisterCommand(const FString& Command)
{
    RegisteredCommands.Remove(Command.ToLower());
    
    UE_LOG(LogMMORPGConsole, Log, TEXT("Unregistered command: %s"), *Command);
}

TArray<FString> UMMORPGDeveloperConsole::GetRegisteredCommands() const
{
    TArray<FString> Commands;
    RegisteredCommands.GetKeys(Commands);
    Commands.Sort();
    return Commands;
}

FString UMMORPGDeveloperConsole::GetCommandDescription(const FString& Command) const
{
    if (const FCommandInfo* CmdInfo = RegisteredCommands.Find(Command.ToLower()))
    {
        return CmdInfo->Description;
    }
    return FString();
}

void UMMORPGDeveloperConsole::RegisterBuiltInCommands()
{
    // Help command
    RegisterCommand(TEXT("help"), TEXT("Show available commands"),
        [this](const TArray<FString>& Args) { HandleHelpCommand(Args); });
    
    // Clear command
    RegisterCommand(TEXT("clear"), TEXT("Clear console output"),
        [this](const TArray<FString>& Args) { HandleClearCommand(Args); });
    
    // Status command
    RegisterCommand(TEXT("status"), TEXT("Show system status"),
        [this](const TArray<FString>& Args) { HandleStatusCommand(Args); });
    
    // Network commands
    RegisterCommand(TEXT("connect"), TEXT("Connect to server (usage: connect <host> <port>)"),
        [this](const TArray<FString>& Args) { HandleConnectCommand(Args); });
    
    RegisterCommand(TEXT("disconnect"), TEXT("Disconnect from server"),
        [this](const TArray<FString>& Args) { HandleDisconnectCommand(Args); });
    
    RegisterCommand(TEXT("test"), TEXT("Run connection test"),
        [this](const TArray<FString>& Args) { HandleTestCommand(Args); });
    
    // Stats commands
    RegisterCommand(TEXT("netstats"), TEXT("Show network statistics"),
        [this](const TArray<FString>& Args) { HandleNetStatsCommand(Args); });
    
    RegisterCommand(TEXT("memstats"), TEXT("Show memory statistics"),
        [this](const TArray<FString>& Args) { HandleMemStatsCommand(Args); });
    
    RegisterCommand(TEXT("fps"), TEXT("Toggle FPS display"),
        [this](const TArray<FString>& Args) { HandleFPSCommand(Args); });
    
    // System commands
    RegisterCommand(TEXT("quit"), TEXT("Quit the game"),
        [this](const TArray<FString>& Args) { HandleQuitCommand(Args); });
}

void UMMORPGDeveloperConsole::HandleHelpCommand(const TArray<FString>& Args)
{
    if (Args.Num() > 0)
    {
        // Show help for specific command
        FString Command = Args[0].ToLower();
        if (RegisteredCommands.Contains(Command))
        {
            WriteOutput(FString::Printf(TEXT("%s: %s"), *Command, *GetCommandDescription(Command)), FColor::White);
        }
        else
        {
            WriteOutput(FString::Printf(TEXT("Unknown command: %s"), *Command), FColor::Red);
        }
    }
    else
    {
        // Show all commands
        WriteOutput(TEXT("Available commands:"), FColor::Yellow);
        
        TArray<FString> Commands = GetRegisteredCommands();
        for (const FString& Cmd : Commands)
        {
            WriteOutput(FString::Printf(TEXT("  %s - %s"), *Cmd, *GetCommandDescription(Cmd)), FColor::White);
        }
        
        WriteOutput(TEXT(""), FColor::White);
        WriteOutput(TEXT("Use 'help <command>' for more information"), FColor::Gray);
    }
}

void UMMORPGDeveloperConsole::HandleClearCommand(const TArray<FString>& Args)
{
    ClearOutput();
}

void UMMORPGDeveloperConsole::HandleStatusCommand(const TArray<FString>& Args)
{
    WriteOutput(TEXT("=== System Status ==="), FColor::Yellow);
    
    // Game info
    UWorld* World = GEngine->GetWorldFromContextObjectChecked(this);
    if (World)
    {
        WriteOutput(FString::Printf(TEXT("World: %s"), *World->GetMapName()), FColor::White);
        WriteOutput(FString::Printf(TEXT("Net Mode: %s"), *UEnum::GetValueAsString(World->GetNetMode())), FColor::White);
        
        if (AGameStateBase* GameState = World->GetGameState())
        {
            WriteOutput(FString::Printf(TEXT("Game Time: %.2f"), GameState->GetServerWorldTimeSeconds()), FColor::White);
        }
    }
    
    // Network status
    FMMORPGCoreModule& Module = FMMORPGCoreModule::Get();
    if (FMMORPGNetworkManager* NetManager = Module.GetNetworkManager())
    {
        WriteOutput(FString::Printf(TEXT("Network: %s"), NetManager->IsConnected() ? TEXT("Connected") : TEXT("Disconnected")), 
                   NetManager->IsConnected() ? FColor::Green : FColor::Red);
        WriteOutput(FString::Printf(TEXT("Server: %s"), *NetManager->GetServerURL()), FColor::White);
    }
    
    // Memory info
    FPlatformMemoryStats MemStats = FPlatformMemory::GetStats();
    WriteOutput(FString::Printf(TEXT("Memory Used: %.2f MB"), MemStats.UsedPhysical / (1024.0f * 1024.0f)), FColor::White);
    WriteOutput(FString::Printf(TEXT("Memory Available: %.2f MB"), MemStats.AvailablePhysical / (1024.0f * 1024.0f)), FColor::White);
}

void UMMORPGDeveloperConsole::HandleConnectCommand(const TArray<FString>& Args)
{
    if (Args.Num() < 2)
    {
        WriteOutput(TEXT("Usage: connect <host> <port>"), FColor::Red);
        return;
    }
    
    FString Host = Args[0];
    int32 Port = FCString::Atoi(*Args[1]);
    
    if (Port <= 0 || Port > 65535)
    {
        WriteOutput(TEXT("Invalid port number"), FColor::Red);
        return;
    }
    
    FMMORPGCoreModule& Module = FMMORPGCoreModule::Get();
    if (FMMORPGNetworkManager* NetManager = Module.GetNetworkManager())
    {
        WriteOutput(FString::Printf(TEXT("Connecting to %s:%d..."), *Host, Port), FColor::Yellow);
        NetManager->Connect(Host, Port);
    }
}

void UMMORPGDeveloperConsole::HandleDisconnectCommand(const TArray<FString>& Args)
{
    FMMORPGCoreModule& Module = FMMORPGCoreModule::Get();
    if (FMMORPGNetworkManager* NetManager = Module.GetNetworkManager())
    {
        NetManager->Disconnect();
        WriteOutput(TEXT("Disconnected from server"), FColor::Yellow);
    }
}

void UMMORPGDeveloperConsole::HandleTestCommand(const TArray<FString>& Args)
{
    FMMORPGCoreModule& Module = FMMORPGCoreModule::Get();
    if (FMMORPGNetworkManager* NetManager = Module.GetNetworkManager())
    {
        WriteOutput(TEXT("Running connection test..."), FColor::Yellow);
        
        NetManager->TestConnection([this](bool bSuccess, const FString& Response)
        {
            if (bSuccess)
            {
                WriteOutput(TEXT("Connection test successful!"), FColor::Green);
                WriteOutput(FString::Printf(TEXT("Response: %s"), *Response), FColor::White);
            }
            else
            {
                WriteOutput(TEXT("Connection test failed!"), FColor::Red);
                WriteOutput(FString::Printf(TEXT("Error: %s"), *Response), FColor::Red);
            }
        });
    }
}

void UMMORPGDeveloperConsole::HandleNetStatsCommand(const TArray<FString>& Args)
{
    WriteOutput(TEXT("=== Network Statistics ==="), FColor::Yellow);
    
    FMMORPGCoreModule& Module = FMMORPGCoreModule::Get();
    if (FMMORPGNetworkManager* NetManager = Module.GetNetworkManager())
    {
        // TODO: Add actual network statistics
        WriteOutput(TEXT("Network stats not yet implemented"), FColor::Gray);
    }
}

void UMMORPGDeveloperConsole::HandleMemStatsCommand(const TArray<FString>& Args)
{
    WriteOutput(TEXT("=== Memory Statistics ==="), FColor::Yellow);
    
    FPlatformMemoryStats MemStats = FPlatformMemory::GetStats();
    
    WriteOutput(FString::Printf(TEXT("Total Physical: %.2f GB"), MemStats.TotalPhysical / (1024.0f * 1024.0f * 1024.0f)), FColor::White);
    WriteOutput(FString::Printf(TEXT("Used Physical: %.2f GB"), MemStats.UsedPhysical / (1024.0f * 1024.0f * 1024.0f)), FColor::White);
    WriteOutput(FString::Printf(TEXT("Peak Used Physical: %.2f GB"), MemStats.PeakUsedPhysical / (1024.0f * 1024.0f * 1024.0f)), FColor::White);
    WriteOutput(FString::Printf(TEXT("Available Physical: %.2f GB"), MemStats.AvailablePhysical / (1024.0f * 1024.0f * 1024.0f)), FColor::White);
}

void UMMORPGDeveloperConsole::HandleFPSCommand(const TArray<FString>& Args)
{
    UWorld* World = GEngine->GetWorldFromContextObjectChecked(this);
    if (World && World->GetFirstPlayerController())
    {
        World->GetFirstPlayerController()->ConsoleCommand(TEXT("stat fps"));
        WriteOutput(TEXT("FPS display toggled"), FColor::Green);
    }
}

void UMMORPGDeveloperConsole::HandleQuitCommand(const TArray<FString>& Args)
{
    WriteOutput(TEXT("Quitting game..."), FColor::Yellow);
    
    UWorld* World = GEngine->GetWorldFromContextObjectChecked(this);
    if (World && World->GetFirstPlayerController())
    {
        UKismetSystemLibrary::QuitGame(World, World->GetFirstPlayerController(), EQuitPreference::Quit, false);
    }
}

void UMMORPGDeveloperConsole::ParseCommandLine(const FString& CommandLine, FString& OutCommand, TArray<FString>& OutArgs) const
{
    TArray<FString> Tokens;
    CommandLine.ParseIntoArray(Tokens, TEXT(" "), true);
    
    if (Tokens.Num() > 0)
    {
        OutCommand = Tokens[0].ToLower();
        
        for (int32 i = 1; i < Tokens.Num(); ++i)
        {
            OutArgs.Add(Tokens[i]);
        }
    }
}

// Console command registration helper

FMMORPGConsoleCommand::FMMORPGConsoleCommand(const FString& Command, const FString& Description, 
                                           TFunction<void(const TArray<FString>&)> Handler)
    : CommandName(Command)
{
    FMMORPGCoreModule& Module = FMMORPGCoreModule::Get();
    if (UMMORPGDeveloperConsole* Console = Module.GetDeveloperConsole())
    {
        Console->RegisterCommand(Command, Description, Handler);
    }
}

FMMORPGConsoleCommand::~FMMORPGConsoleCommand()
{
    FMMORPGCoreModule& Module = FMMORPGCoreModule::Get();
    if (UMMORPGDeveloperConsole* Console = Module.GetDeveloperConsole())
    {
        Console->UnregisterCommand(CommandName);
    }
}