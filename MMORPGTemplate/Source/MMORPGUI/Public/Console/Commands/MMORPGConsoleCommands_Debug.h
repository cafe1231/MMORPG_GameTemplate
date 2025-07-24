#pragma once

#include "CoreMinimal.h"
#include "Console/MMORPGConsoleCommand.h"
#include "MMORPGConsoleCommands_Debug.generated.h"

/**
 * Show FPS command
 */
UCLASS()
class MMORPGUI_API UMMORPGConsoleCommand_ShowFPS : public UMMORPGConsoleCommand
{
	GENERATED_BODY()

public:
	UMMORPGConsoleCommand_ShowFPS();
	virtual FString Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext) override;
};

/**
 * Set screen resolution command
 */
UCLASS()
class MMORPGUI_API UMMORPGConsoleCommand_SetResolution : public UMMORPGConsoleCommand
{
	GENERATED_BODY()

public:
	UMMORPGConsoleCommand_SetResolution();
	virtual FString Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext) override;
};

/**
 * Network status command
 */
UCLASS()
class MMORPGUI_API UMMORPGConsoleCommand_NetStatus : public UMMORPGConsoleCommand
{
	GENERATED_BODY()

public:
	UMMORPGConsoleCommand_NetStatus();
	virtual FString Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext) override;
};

/**
 * Memory stats command
 */
UCLASS()
class MMORPGUI_API UMMORPGConsoleCommand_MemStats : public UMMORPGConsoleCommand
{
	GENERATED_BODY()

public:
	UMMORPGConsoleCommand_MemStats();
	virtual FString Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext) override;
};

/**
 * List console variables command
 */
UCLASS()
class MMORPGUI_API UMMORPGConsoleCommand_ListCVars : public UMMORPGConsoleCommand
{
	GENERATED_BODY()

public:
	UMMORPGConsoleCommand_ListCVars();
	virtual FString Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext) override;
};