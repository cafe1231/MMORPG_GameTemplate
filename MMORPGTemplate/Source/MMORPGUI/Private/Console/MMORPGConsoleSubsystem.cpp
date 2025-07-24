#include "Console/MMORPGConsoleSubsystem.h"
#include "Console/Commands/MMORPGConsoleCommands_Debug.h"
#include "MMORPGUI.h"
#include "Engine/World.h"

void UMMORPGConsoleSubsystem::Initialize(FSubsystemCollectionBase& Collection)
{
	Super::Initialize(Collection);

	UE_LOG(LogMMORPGUI, Log, TEXT("MMORPGConsoleSubsystem initialized"));

	// Register built-in commands
	RegisterBuiltInCommands();
}

void UMMORPGConsoleSubsystem::Deinitialize()
{
	Commands.Empty();
	Aliases.Empty();
	History.Empty();

	Super::Deinitialize();

	UE_LOG(LogMMORPGUI, Log, TEXT("MMORPGConsoleSubsystem deinitialized"));
}

bool UMMORPGConsoleSubsystem::RegisterCommand(UMMORPGConsoleCommand* Command)
{
	if (!Command || Command->CommandName.IsEmpty())
	{
		return false;
	}

	// Check if command already exists
	if (Commands.Contains(Command->CommandName))
	{
		UE_LOG(LogMMORPGUI, Warning, TEXT("Command '%s' already registered"), *Command->CommandName);
		return false;
	}

	// Register main command
	Commands.Add(Command->CommandName, Command);

	// Register aliases
	for (const FString& Alias : Command->Aliases)
	{
		if (!Alias.IsEmpty() && !Aliases.Contains(Alias))
		{
			Aliases.Add(Alias, Command->CommandName);
		}
	}

	UE_LOG(LogMMORPGUI, Log, TEXT("Registered console command: %s"), *Command->CommandName);
	return true;
}

void UMMORPGConsoleSubsystem::UnregisterCommand(const FString& CommandName)
{
	UMMORPGConsoleCommand* Command = nullptr;
	if (Commands.RemoveAndCopyValue(CommandName, Command))
	{
		// Remove aliases
		if (Command)
		{
			for (const FString& Alias : Command->Aliases)
			{
				Aliases.Remove(Alias);
			}
		}

		UE_LOG(LogMMORPGUI, Log, TEXT("Unregistered console command: %s"), *CommandName);
	}
}

UMMORPGConsoleCommand* UMMORPGConsoleSubsystem::FindCommand(const FString& CommandName) const
{
	// Check direct command
	if (Commands.Contains(CommandName))
	{
		return Commands[CommandName];
	}

	// Check aliases
	if (Aliases.Contains(CommandName))
	{
		const FString& RealCommand = Aliases[CommandName];
		if (Commands.Contains(RealCommand))
		{
			return Commands[RealCommand];
		}
	}

	return nullptr;
}

TArray<UMMORPGConsoleCommand*> UMMORPGConsoleSubsystem::GetAllCommands() const
{
	TArray<UMMORPGConsoleCommand*> Result;
	Commands.GenerateValueArray(Result);
	return Result;
}

FString UMMORPGConsoleSubsystem::ExecuteCommand(const FString& CommandLine, UObject* WorldContext)
{
	if (!bConsoleEnabled)
	{
		return TEXT("Console is disabled");
	}

	if (CommandLine.IsEmpty())
	{
		return TEXT("");
	}

	// Parse command line
	FString CommandName;
	TArray<FString> Args;
	ParseCommandLine(CommandLine, CommandName, Args);

	// Find command
	UMMORPGConsoleCommand* Command = FindCommand(CommandName);
	if (!Command)
	{
		FString Output = FString::Printf(TEXT("Unknown command: %s"), *CommandName);
		AddToHistory(CommandLine, Output);
		OnConsoleOutput.Broadcast(Output);
		return Output;
	}

	// Validate arguments
	FString ValidationError;
	if (!Command->ValidateArguments(Args, ValidationError))
	{
		FString Output = FString::Printf(TEXT("Error: %s\nUsage: %s"), *ValidationError, *Command->GetUsageString());
		AddToHistory(CommandLine, Output);
		OnConsoleOutput.Broadcast(Output);
		return Output;
	}

	// Execute command
	FString Output = Command->Execute(Args, WorldContext);

	// Handle special commands
	if (Output == TEXT("@CLEAR_CONSOLE@"))
	{
		ClearHistory();
		OnConsoleOutput.Broadcast(TEXT(""));
		return TEXT("");
	}

	// Add to history and broadcast
	AddToHistory(CommandLine, Output);
	OnConsoleOutput.Broadcast(Output);
	OnCommandExecuted.Broadcast(CommandLine);

	return Output;
}

void UMMORPGConsoleSubsystem::ParseCommandLine(const FString& CommandLine, FString& OutCommand, TArray<FString>& OutArgs)
{
	// Simple parsing - split by spaces, respecting quotes
	TArray<FString> Tokens;
	FString CurrentToken;
	bool bInQuotes = false;

	for (int32 i = 0; i < CommandLine.Len(); i++)
	{
		TCHAR Ch = CommandLine[i];

		if (Ch == '"')
		{
			bInQuotes = !bInQuotes;
		}
		else if (Ch == ' ' && !bInQuotes)
		{
			if (!CurrentToken.IsEmpty())
			{
				Tokens.Add(CurrentToken);
				CurrentToken.Empty();
			}
		}
		else
		{
			CurrentToken.AppendChar(Ch);
		}
	}

	if (!CurrentToken.IsEmpty())
	{
		Tokens.Add(CurrentToken);
	}

	// First token is command
	if (Tokens.Num() > 0)
	{
		OutCommand = Tokens[0];
		for (int32 i = 1; i < Tokens.Num(); i++)
		{
			OutArgs.Add(Tokens[i]);
		}
	}
	else
	{
		OutCommand.Empty();
	}
}

TArray<FString> UMMORPGConsoleSubsystem::GetCommandHistory() const
{
	TArray<FString> Result;
	
	// Extract unique commands from history
	TSet<FString> UniqueCommands;
	for (int32 i = History.Num() - 1; i >= 0; i--)
	{
		if (!History[i].Command.IsEmpty() && !UniqueCommands.Contains(History[i].Command))
		{
			UniqueCommands.Add(History[i].Command);
			Result.Add(History[i].Command);
		}
	}

	return Result;
}

void UMMORPGConsoleSubsystem::ClearHistory()
{
	History.Empty();
}

TArray<FString> UMMORPGConsoleSubsystem::GetAutoCompleteSuggestions(const FString& PartialCommand, int32 MaxSuggestions) const
{
	TArray<FString> Suggestions;

	if (PartialCommand.IsEmpty())
	{
		return Suggestions;
	}

	// Parse partial command
	FString CommandName;
	TArray<FString> Args;
	ParseCommandLine(PartialCommand, CommandName, Args);

	// If we're still typing the command name
	if (Args.Num() == 0 && !PartialCommand.EndsWith(TEXT(" ")))
	{
		// Search command names
		for (const auto& Pair : Commands)
		{
			if (Pair.Key.StartsWith(CommandName, ESearchCase::IgnoreCase))
			{
				Suggestions.Add(Pair.Key);
			}
		}

		// Search aliases
		for (const auto& Pair : Aliases)
		{
			if (Pair.Key.StartsWith(CommandName, ESearchCase::IgnoreCase))
			{
				Suggestions.Add(Pair.Key);
			}
		}
	}
	else
	{
		// We're typing arguments - find the command and provide context-specific suggestions
		UMMORPGConsoleCommand* Command = FindCommand(CommandName);
		if (Command)
		{
			// For now, just show the usage
			Suggestions.Add(Command->GetUsageString());
		}
	}

	// Sort and limit
	Suggestions.Sort();
	if (Suggestions.Num() > MaxSuggestions)
	{
		Suggestions.SetNum(MaxSuggestions);
	}

	return Suggestions;
}

void UMMORPGConsoleSubsystem::WriteOutput(const FString& Text)
{
	if (!Text.IsEmpty())
	{
		FConsoleHistoryEntry Entry;
		Entry.Output = Text;
		History.Add(Entry);

		// Trim history if needed
		if (History.Num() > MaxHistorySize)
		{
			History.RemoveAt(0, History.Num() - MaxHistorySize);
		}

		OnConsoleOutput.Broadcast(Text);
	}
}

void UMMORPGConsoleSubsystem::SetMaxHistorySize(int32 NewSize)
{
	MaxHistorySize = FMath::Max(10, NewSize);

	// Trim history if needed
	if (History.Num() > MaxHistorySize)
	{
		History.RemoveAt(0, History.Num() - MaxHistorySize);
	}
}

void UMMORPGConsoleSubsystem::SetConsoleEnabled(bool bEnabled)
{
	bConsoleEnabled = bEnabled;
	UE_LOG(LogMMORPGUI, Log, TEXT("Console %s"), bEnabled ? TEXT("enabled") : TEXT("disabled"));
}

void UMMORPGConsoleSubsystem::RegisterBuiltInCommands()
{
	// System commands
	RegisterCommand(NewObject<UMMORPGConsoleCommand_Help>(this));
	RegisterCommand(NewObject<UMMORPGConsoleCommand_Clear>(this));

	// Debug commands
	RegisterCommand(NewObject<UMMORPGConsoleCommand_ShowFPS>(this));
	RegisterCommand(NewObject<UMMORPGConsoleCommand_SetResolution>(this));
	RegisterCommand(NewObject<UMMORPGConsoleCommand_NetStatus>(this));
	RegisterCommand(NewObject<UMMORPGConsoleCommand_MemStats>(this));
	RegisterCommand(NewObject<UMMORPGConsoleCommand_ListCVars>(this));
}

void UMMORPGConsoleSubsystem::AddToHistory(const FString& Command, const FString& Output)
{
	FConsoleHistoryEntry Entry;
	Entry.Command = Command;
	Entry.Output = Output;
	History.Add(Entry);

	// Trim history if needed
	if (History.Num() > MaxHistorySize)
	{
		History.RemoveAt(0, History.Num() - MaxHistorySize);
	}
}