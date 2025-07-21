// Copyright (c) 2024 MMORPG Template Project

#pragma once

#include "CoreMinimal.h"
#include "Engine/Console.h"
#include "Engine/DebugCameraController.h"
#include "MMORPGDeveloperConsole.generated.h"

DECLARE_LOG_CATEGORY_EXTERN(LogMMORPGConsole, Log, All);

DECLARE_MULTICAST_DELEGATE_TwoParams(FOnConsoleCommand, const FString&, const TArray<FString>&);

/**
 * MMORPG Developer Console System
 * Provides an in-game console for debugging and development commands
 */
UCLASS(Config=Game)
class MMORPGCORE_API UMMORPGDeveloperConsole : public UObject
{
    GENERATED_BODY()

public:
    UMMORPGDeveloperConsole();

    /** Initialize the console system */
    void Initialize();

    /** Shutdown the console system */
    void Shutdown();

    /** Toggle console visibility */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    void ToggleConsole();

    /** Show the console */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    void ShowConsole();

    /** Hide the console */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    void HideConsole();

    /** Check if console is visible */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    bool IsConsoleVisible() const { return bIsVisible; }

    /** Execute a console command */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    void ExecuteCommand(const FString& Command);

    /** Add a command to history */
    void AddToHistory(const FString& Command);

    /** Get command history */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    TArray<FString> GetCommandHistory() const { return CommandHistory; }

    /** Clear command history */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    void ClearHistory() { CommandHistory.Empty(); }

    /** Write output to console */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    void WriteOutput(const FString& Message, const FColor& Color = FColor::White);

    /** Clear console output */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    void ClearOutput();

    /** Get console output */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    TArray<FString> GetOutput() const { return ConsoleOutput; }

    /** Register a custom command handler */
    void RegisterCommand(const FString& Command, const FString& Description, TFunction<void(const TArray<FString>&)> Handler);

    /** Unregister a command */
    void UnregisterCommand(const FString& Command);

    /** Get all registered commands */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    TArray<FString> GetRegisteredCommands() const;

    /** Get command description */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    FString GetCommandDescription(const FString& Command) const;

    /** Delegate for command execution */
    FOnConsoleCommand OnCommandExecuted;

protected:
    /** Register built-in commands */
    void RegisterBuiltInCommands();

    /** Handle help command */
    void HandleHelpCommand(const TArray<FString>& Args);

    /** Handle clear command */
    void HandleClearCommand(const TArray<FString>& Args);

    /** Handle status command */
    void HandleStatusCommand(const TArray<FString>& Args);

    /** Handle connect command */
    void HandleConnectCommand(const TArray<FString>& Args);

    /** Handle disconnect command */
    void HandleDisconnectCommand(const TArray<FString>& Args);

    /** Handle test command */
    void HandleTestCommand(const TArray<FString>& Args);

    /** Handle network stats command */
    void HandleNetStatsCommand(const TArray<FString>& Args);

    /** Handle memory stats command */
    void HandleMemStatsCommand(const TArray<FString>& Args);

    /** Handle FPS command */
    void HandleFPSCommand(const TArray<FString>& Args);

    /** Handle quit command */
    void HandleQuitCommand(const TArray<FString>& Args);

    /** Parse command line into command and arguments */
    void ParseCommandLine(const FString& CommandLine, FString& OutCommand, TArray<FString>& OutArgs) const;

private:
    /** Console visibility state */
    UPROPERTY()
    bool bIsVisible;

    /** Command history */
    UPROPERTY(Config)
    TArray<FString> CommandHistory;

    /** Maximum history size */
    UPROPERTY(Config)
    int32 MaxHistorySize;

    /** Console output buffer */
    UPROPERTY()
    TArray<FString> ConsoleOutput;

    /** Maximum output lines */
    UPROPERTY(Config)
    int32 MaxOutputLines;

    /** Command info structure */
    struct FCommandInfo
    {
        FString Description;
        TFunction<void(const TArray<FString>&)> Handler;
    };

    /** Registered commands */
    TMap<FString, FCommandInfo> RegisteredCommands;

    /** Current history index for navigation */
    int32 CurrentHistoryIndex;

    /** Console widget instance */
    UPROPERTY()
    class UMMORPGConsoleWidget* ConsoleWidget;

    /** Widget class to spawn */
    UPROPERTY()
    TSubclassOf<class UMMORPGConsoleWidget> ConsoleWidgetClass;
};

/**
 * Console command registration helper
 */
class MMORPGCORE_API FMMORPGConsoleCommand
{
public:
    FMMORPGConsoleCommand(const FString& Command, const FString& Description, TFunction<void(const TArray<FString>&)> Handler);
    ~FMMORPGConsoleCommand();

private:
    FString CommandName;
};

/**
 * Macro for easy command registration
 */
#define MMORPG_CONSOLE_COMMAND(Command, Description, Handler) \
    static FMMORPGConsoleCommand Command##_Registration(TEXT(#Command), TEXT(Description), Handler);