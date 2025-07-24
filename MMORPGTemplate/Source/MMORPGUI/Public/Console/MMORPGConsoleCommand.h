#pragma once

#include "CoreMinimal.h"
#include "Engine/DataAsset.h"
#include "MMORPGConsoleCommand.generated.h"

/**
 * Console command parameter type
 */
UENUM(BlueprintType)
enum class EConsoleParamType : uint8
{
	String		UMETA(DisplayName = "String"),
	Integer		UMETA(DisplayName = "Integer"),
	Float		UMETA(DisplayName = "Float"),
	Boolean		UMETA(DisplayName = "Boolean")
};

/**
 * Console command parameter definition
 */
USTRUCT(BlueprintType)
struct MMORPGUI_API FConsoleCommandParam
{
	GENERATED_BODY()

	// Parameter name
	UPROPERTY(EditAnywhere, BlueprintReadOnly, Category = "Console")
	FString Name;

	// Parameter type
	UPROPERTY(EditAnywhere, BlueprintReadOnly, Category = "Console")
	EConsoleParamType Type = EConsoleParamType::String;

	// Is this parameter optional?
	UPROPERTY(EditAnywhere, BlueprintReadOnly, Category = "Console")
	bool bOptional = false;

	// Default value (as string)
	UPROPERTY(EditAnywhere, BlueprintReadOnly, Category = "Console")
	FString DefaultValue;

	// Parameter description
	UPROPERTY(EditAnywhere, BlueprintReadOnly, Category = "Console")
	FString Description;
};

/**
 * Base class for console commands
 */
UCLASS(Abstract, Blueprintable)
class MMORPGUI_API UMMORPGConsoleCommand : public UDataAsset
{
	GENERATED_BODY()

public:
	// Command name (what user types)
	UPROPERTY(EditAnywhere, BlueprintReadOnly, Category = "Console")
	FString CommandName;

	// Command aliases
	UPROPERTY(EditAnywhere, BlueprintReadOnly, Category = "Console")
	TArray<FString> Aliases;

	// Command description
	UPROPERTY(EditAnywhere, BlueprintReadOnly, Category = "Console", meta = (MultiLine = true))
	FString Description;

	// Command category
	UPROPERTY(EditAnywhere, BlueprintReadOnly, Category = "Console")
	FString Category = TEXT("General");

	// Required permission level
	UPROPERTY(EditAnywhere, BlueprintReadOnly, Category = "Console")
	int32 RequiredPermissionLevel = 0;

	// Command parameters
	UPROPERTY(EditAnywhere, BlueprintReadOnly, Category = "Console")
	TArray<FConsoleCommandParam> Parameters;

	// Is this command available in shipping builds?
	UPROPERTY(EditAnywhere, BlueprintReadOnly, Category = "Console")
	bool bAvailableInShipping = false;

	/**
	 * Execute the command
	 * @param Args Command arguments
	 * @param WorldContext World context
	 * @return Command output
	 */
	UFUNCTION(BlueprintNativeEvent, Category = "Console")
	FString Execute(const TArray<FString>& Args, UObject* WorldContext);
	virtual FString Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext);

	/**
	 * Get command usage string
	 * @return Usage string
	 */
	UFUNCTION(BlueprintPure, Category = "Console")
	FString GetUsageString() const;

	/**
	 * Validate arguments
	 * @param Args Arguments to validate
	 * @param OutError Error message if validation fails
	 * @return True if arguments are valid
	 */
	UFUNCTION(BlueprintPure, Category = "Console")
	bool ValidateArguments(const TArray<FString>& Args, FString& OutError) const;

	/**
	 * Parse argument value
	 * @param ArgString Argument string
	 * @param ParamType Expected parameter type
	 * @param OutValue Parsed value
	 * @return True if parsing succeeded
	 */
	UFUNCTION(BlueprintPure, Category = "Console")
	static bool ParseArgumentValue(const FString& ArgString, EConsoleParamType ParamType, FString& OutValue);
};

/**
 * Built-in help command
 */
UCLASS()
class MMORPGUI_API UMMORPGConsoleCommand_Help : public UMMORPGConsoleCommand
{
	GENERATED_BODY()

public:
	UMMORPGConsoleCommand_Help();

	virtual FString Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext) override;
};

/**
 * Built-in clear command
 */
UCLASS()
class MMORPGUI_API UMMORPGConsoleCommand_Clear : public UMMORPGConsoleCommand
{
	GENERATED_BODY()

public:
	UMMORPGConsoleCommand_Clear();

	virtual FString Execute_Implementation(const TArray<FString>& Args, UObject* WorldContext) override;
};