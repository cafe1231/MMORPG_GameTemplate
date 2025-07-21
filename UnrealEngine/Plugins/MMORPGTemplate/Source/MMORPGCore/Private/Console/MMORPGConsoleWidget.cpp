// Copyright (c) 2024 MMORPG Template Project

#include "Console/MMORPGConsoleWidget.h"
#include "Console/MMORPGDeveloperConsole.h"

void UMMORPGConsoleWidget::OnCommandSubmitted(const FString& Command)
{
    if (Console && !Command.IsEmpty())
    {
        Console->ExecuteCommand(Command);
        HistoryIndex = -1; // Reset history navigation
    }
}

FString UMMORPGConsoleWidget::NavigateHistory(bool bUp)
{
    if (!Console)
    {
        return FString();
    }
    
    TArray<FString> History = Console->GetCommandHistory();
    if (History.Num() == 0)
    {
        return FString();
    }
    
    if (bUp)
    {
        // Navigate up (older commands)
        if (HistoryIndex == -1)
        {
            HistoryIndex = History.Num() - 1;
        }
        else if (HistoryIndex > 0)
        {
            HistoryIndex--;
        }
    }
    else
    {
        // Navigate down (newer commands)
        if (HistoryIndex >= 0 && HistoryIndex < History.Num() - 1)
        {
            HistoryIndex++;
        }
        else
        {
            HistoryIndex = -1;
            return FString(); // Return empty for current input
        }
    }
    
    if (HistoryIndex >= 0 && HistoryIndex < History.Num())
    {
        return History[HistoryIndex];
    }
    
    return FString();
}

TArray<FString> UMMORPGConsoleWidget::GetAutocompleteSuggestions(const FString& PartialCommand)
{
    TArray<FString> Suggestions;
    
    if (!Console || PartialCommand.IsEmpty())
    {
        return Suggestions;
    }
    
    FString LowerPartial = PartialCommand.ToLower();
    TArray<FString> Commands = Console->GetRegisteredCommands();
    
    for (const FString& Cmd : Commands)
    {
        if (Cmd.StartsWith(LowerPartial))
        {
            Suggestions.Add(Cmd);
        }
    }
    
    // Sort by length (shorter commands first)
    Suggestions.Sort([](const FString& A, const FString& B)
    {
        return A.Len() < B.Len();
    });
    
    return Suggestions;
}