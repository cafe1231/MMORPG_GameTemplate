// Copyright (c) 2024 MMORPG Template Project

#pragma once

#include "CoreMinimal.h"
#include "Blueprint/UserWidget.h"
#include "MMORPGConsoleWidget.generated.h"

/**
 * Console widget for the MMORPG Developer Console
 * This should be implemented in Blueprint
 */
UCLASS()
class MMORPGCORE_API UMMORPGConsoleWidget : public UUserWidget
{
    GENERATED_BODY()

public:
    /** Set the console instance */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    void SetConsole(class UMMORPGDeveloperConsole* InConsole) { Console = InConsole; }

    /** Add a line to the output */
    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|Console")
    void AddOutputLine(const FString& Text, const FColor& Color);

    /** Clear all output */
    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|Console")
    void ClearOutput();

    /** Set keyboard focus to input field */
    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|Console")
    void SetKeyboardFocus();

    /** Called when the user submits a command */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    void OnCommandSubmitted(const FString& Command);

    /** Called when user navigates history (up/down arrows) */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    FString NavigateHistory(bool bUp);

    /** Get autocomplete suggestions for current input */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    TArray<FString> GetAutocompleteSuggestions(const FString& PartialCommand);

protected:
    /** Reference to the console */
    UPROPERTY(BlueprintReadOnly, Category = "MMORPG|Console")
    class UMMORPGDeveloperConsole* Console;

    /** Current history navigation index */
    int32 HistoryIndex;
};