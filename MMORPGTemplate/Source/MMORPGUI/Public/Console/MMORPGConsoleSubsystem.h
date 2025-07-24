#pragma once

#include "CoreMinimal.h"
#include "Subsystems/GameInstanceSubsystem.h"
#include "Console/MMORPGConsoleCommand.h"
#include "MMORPGConsoleSubsystem.generated.h"

DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnConsoleOutput, const FString&, Output);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnConsoleCommandExecuted, const FString&, Command);

/**
 * Console history entry
 */
USTRUCT(BlueprintType)
struct MMORPGUI_API FConsoleHistoryEntry
{
	GENERATED_BODY()

	UPROPERTY(BlueprintReadOnly)
	FString Command;

	UPROPERTY(BlueprintReadOnly)
	FString Output;

	UPROPERTY(BlueprintReadOnly)
	FDateTime Timestamp;

	FConsoleHistoryEntry()
	{
		Timestamp = FDateTime::Now();
	}
};

/**
 * Console subsystem for managing developer console
 */
UCLASS()
class MMORPGUI_API UMMORPGConsoleSubsystem : public UGameInstanceSubsystem
{
	GENERATED_BODY()

public:
	// USubsystem interface
	virtual void Initialize(FSubsystemCollectionBase& Collection) override;
	virtual void Deinitialize() override;

	// Events
	UPROPERTY(BlueprintAssignable, Category = "MMORPG|Console")
	FOnConsoleOutput OnConsoleOutput;

	UPROPERTY(BlueprintAssignable, Category = "MMORPG|Console")
	FOnConsoleCommandExecuted OnCommandExecuted;

	// Command Management

	/**
	 * Register a console command
	 * @param Command The command to register
	 * @return True if registration succeeded
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
	bool RegisterCommand(UMMORPGConsoleCommand* Command);

	/**
	 * Unregister a console command
	 * @param CommandName The command name to unregister
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
	void UnregisterCommand(const FString& CommandName);

	/**
	 * Find a command by name
	 * @param CommandName The command name (or alias)
	 * @return The command if found
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Console")
	UMMORPGConsoleCommand* FindCommand(const FString& CommandName) const;

	/**
	 * Get all registered commands
	 * @return Array of all commands
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Console")
	TArray<UMMORPGConsoleCommand*> GetAllCommands() const;

	// Command Execution

	/**
	 * Execute a console command
	 * @param CommandLine The full command line to execute
	 * @param WorldContext World context for the command
	 * @return Command output
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Console", meta = (WorldContext = "WorldContext"))
	FString ExecuteCommand(const FString& CommandLine, UObject* WorldContext);

	/**
	 * Parse command line into command and arguments
	 * @param CommandLine The command line to parse
	 * @param OutCommand The command name
	 * @param OutArgs The arguments
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Console")
	static void ParseCommandLine(const FString& CommandLine, FString& OutCommand, TArray<FString>& OutArgs);

	// History

	/**
	 * Get command history
	 * @return Array of history entries
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Console")
	const TArray<FConsoleHistoryEntry>& GetHistory() const { return History; }

	/**
	 * Get command input history (just the commands, not outputs)
	 * @return Array of previously entered commands
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Console")
	TArray<FString> GetCommandHistory() const;

	/**
	 * Clear history
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
	void ClearHistory();

	// Auto-completion

	/**
	 * Get auto-completion suggestions
	 * @param PartialCommand The partial command to complete
	 * @param MaxSuggestions Maximum number of suggestions
	 * @return Array of suggestions
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Console")
	TArray<FString> GetAutoCompleteSuggestions(const FString& PartialCommand, int32 MaxSuggestions = 10) const;

	// Output

	/**
	 * Write to console output
	 * @param Text The text to write
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
	void WriteOutput(const FString& Text);

	// Settings

	/**
	 * Get maximum history size
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Console")
	int32 GetMaxHistorySize() const { return MaxHistorySize; }

	/**
	 * Set maximum history size
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
	void SetMaxHistorySize(int32 NewSize);

	/**
	 * Check if console is enabled
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Console")
	bool IsConsoleEnabled() const { return bConsoleEnabled; }

	/**
	 * Enable/disable console
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
	void SetConsoleEnabled(bool bEnabled);

protected:
	// Registered commands
	UPROPERTY()
	TMap<FString, UMMORPGConsoleCommand*> Commands;

	// Command aliases
	TMap<FString, FString> Aliases;

	// Command history
	UPROPERTY()
	TArray<FConsoleHistoryEntry> History;

	// Settings
	UPROPERTY()
	int32 MaxHistorySize = 100;

	UPROPERTY()
	bool bConsoleEnabled = true;

	// Built-in commands
	void RegisterBuiltInCommands();

	// Add to history
	void AddToHistory(const FString& Command, const FString& Output);
};