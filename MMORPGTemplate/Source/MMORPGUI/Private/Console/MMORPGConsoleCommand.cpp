#include "Console/MMORPGConsoleCommand.h"
#include "Console/MMORPGConsoleSubsystem.h"
#include "Engine/World.h"
#include "Kismet/GameplayStatics.h"

FString UMMORPGConsoleCommand::Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext)
{
	return TEXT("Command not implemented");
}

FString UMMORPGConsoleCommand::GetUsageString() const
{
	FString Usage = CommandName;

	for (const FConsoleCommandParam& Param : Parameters)
	{
		if (Param.bOptional)
		{
			Usage += FString::Printf(TEXT(" [%s]"), *Param.Name);
		}
		else
		{
			Usage += FString::Printf(TEXT(" <%s>"), *Param.Name);
		}
	}

	return Usage;
}

bool UMMORPGConsoleCommand::ValidateArguments(const TArray<FString>& Args, FString& OutError) const
{
	int32 RequiredParams = 0;
	for (const FConsoleCommandParam& Param : Parameters)
	{
		if (!Param.bOptional)
		{
			RequiredParams++;
		}
	}

	if (Args.Num() < RequiredParams)
	{
		OutError = FString::Printf(TEXT("Not enough arguments. Expected at least %d, got %d"), RequiredParams, Args.Num());
		return false;
	}

	if (Args.Num() > Parameters.Num())
	{
		OutError = FString::Printf(TEXT("Too many arguments. Expected at most %d, got %d"), Parameters.Num(), Args.Num());
		return false;
	}

	// Validate each argument type
	for (int32 i = 0; i < Args.Num(); i++)
	{
		FString ParsedValue;
		if (!ParseArgumentValue(Args[i], Parameters[i].Type, ParsedValue))
		{
			OutError = FString::Printf(TEXT("Invalid %s value for parameter '%s': %s"),
				*UEnum::GetValueAsString(Parameters[i].Type),
				*Parameters[i].Name,
				*Args[i]);
			return false;
		}
	}

	return true;
}

bool UMMORPGConsoleCommand::ParseArgumentValue(const FString& ArgString, EConsoleParamType ParamType, FString& OutValue)
{
	switch (ParamType)
	{
		case EConsoleParamType::String:
			OutValue = ArgString;
			return true;

		case EConsoleParamType::Integer:
		{
			if (ArgString.IsNumeric())
			{
				OutValue = ArgString;
				return true;
			}
			return false;
		}

		case EConsoleParamType::Float:
		{
			if (ArgString.IsNumeric())
			{
				OutValue = ArgString;
				return true;
			}
			// Check for decimal point
			float TestValue;
			if (FCString::Atof(*ArgString) != 0.0f || ArgString == TEXT("0") || ArgString == TEXT("0.0"))
			{
				OutValue = ArgString;
				return true;
			}
			return false;
		}

		case EConsoleParamType::Boolean:
		{
			FString Lower = ArgString.ToLower();
			if (Lower == TEXT("true") || Lower == TEXT("1") || Lower == TEXT("yes") || Lower == TEXT("on"))
			{
				OutValue = TEXT("true");
				return true;
			}
			else if (Lower == TEXT("false") || Lower == TEXT("0") || Lower == TEXT("no") || Lower == TEXT("off"))
			{
				OutValue = TEXT("false");
				return true;
			}
			return false;
		}
	}

	return false;
}

// Help command implementation

UMMORPGConsoleCommand_Help::UMMORPGConsoleCommand_Help()
{
	CommandName = TEXT("help");
	Aliases = { TEXT("?"), TEXT("h") };
	Description = TEXT("Display help information about console commands");
	Category = TEXT("System");
	bAvailableInShipping = true;

	FConsoleCommandParam CommandParam;
	CommandParam.Name = TEXT("command");
	CommandParam.Type = EConsoleParamType::String;
	CommandParam.bOptional = true;
	CommandParam.Description = TEXT("Specific command to get help for");
	Parameters.Add(CommandParam);
}

FString UMMORPGConsoleCommand_Help::Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext)
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

	UMMORPGConsoleSubsystem* Console = GameInstance->GetSubsystem<UMMORPGConsoleSubsystem>();
	if (!Console)
	{
		return TEXT("Error: Console subsystem not found");
	}

	// Specific command help
	if (Args.Num() > 0)
	{
		UMMORPGConsoleCommand* Command = Console->FindCommand(Args[0]);
		if (!Command)
		{
			return FString::Printf(TEXT("Unknown command: %s"), *Args[0]);
		}

		FString Output = FString::Printf(TEXT("=== %s ===\n"), *Command->CommandName);
		Output += FString::Printf(TEXT("Description: %s\n"), *Command->Description);
		Output += FString::Printf(TEXT("Usage: %s\n"), *Command->GetUsageString());
		
		if (Command->Aliases.Num() > 0)
		{
			Output += TEXT("Aliases: ");
			for (int32 i = 0; i < Command->Aliases.Num(); i++)
			{
				Output += Command->Aliases[i];
				if (i < Command->Aliases.Num() - 1)
				{
					Output += TEXT(", ");
				}
			}
			Output += TEXT("\n");
		}

		if (Command->Parameters.Num() > 0)
		{
			Output += TEXT("\nParameters:\n");
			for (const FConsoleCommandParam& Param : Command->Parameters)
			{
				Output += FString::Printf(TEXT("  %s (%s%s) - %s\n"),
					*Param.Name,
					*UEnum::GetValueAsString(Param.Type),
					Param.bOptional ? TEXT(", optional") : TEXT(""),
					*Param.Description);
			}
		}

		return Output;
	}

	// General help - list all commands
	TArray<UMMORPGConsoleCommand*> Commands = Console->GetAllCommands();
	
	// Group by category
	TMap<FString, TArray<UMMORPGConsoleCommand*>> CommandsByCategory;
	for (UMMORPGConsoleCommand* Command : Commands)
	{
		CommandsByCategory.FindOrAdd(Command->Category).Add(Command);
	}

	FString Output = TEXT("=== Available Commands ===\n\n");
	
	// Sort categories
	TArray<FString> Categories;
	CommandsByCategory.GetKeys(Categories);
	Categories.Sort();

	for (const FString& Category : Categories)
	{
		Output += FString::Printf(TEXT("[%s]\n"), *Category);
		
		for (UMMORPGConsoleCommand* Command : CommandsByCategory[Category])
		{
			Output += FString::Printf(TEXT("  %-20s %s\n"), *Command->CommandName, *Command->Description);
		}
		Output += TEXT("\n");
	}

	Output += TEXT("Type 'help <command>' for detailed information about a specific command.");

	return Output;
}

// Clear command implementation

UMMORPGConsoleCommand_Clear::UMMORPGConsoleCommand_Clear()
{
	CommandName = TEXT("clear");
	Aliases = { TEXT("cls") };
	Description = TEXT("Clear the console output");
	Category = TEXT("System");
	bAvailableInShipping = true;
}

FString UMMORPGConsoleCommand_Clear::Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext)
{
	// Return special marker that console UI will recognize
	return TEXT("@CLEAR_CONSOLE@");
}