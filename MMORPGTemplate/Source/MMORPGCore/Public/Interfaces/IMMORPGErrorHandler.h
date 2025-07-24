#pragma once

#include "CoreMinimal.h"
#include "UObject/Interface.h"
#include "CoreTypes.h"
#include "IMMORPGErrorHandler.generated.h"

UINTERFACE(BlueprintType)
class UMMORPGErrorHandler : public UInterface
{
	GENERATED_BODY()
};

/**
 * Interface for objects that can handle errors
 * Implement this to create custom error handling behavior
 */
class MMORPGCORE_API IMMORPGErrorHandler
{
	GENERATED_BODY()

public:
	/**
	 * Handle an error
	 * @param Error The error to handle
	 */
	UFUNCTION(BlueprintCallable, BlueprintNativeEvent, Category = "MMORPG|Error")
	void HandleError(const FMMORPGError& Error);

	/**
	 * Check if this handler can retry the operation that caused the error
	 * @param Error The error to check
	 * @return True if retry is possible
	 */
	UFUNCTION(BlueprintCallable, BlueprintNativeEvent, Category = "MMORPG|Error")
	bool CanRetry(const FMMORPGError& Error);

	/**
	 * Get user-friendly message for the error
	 * @param Error The error to get message for
	 * @return User-friendly error message
	 */
	UFUNCTION(BlueprintCallable, BlueprintNativeEvent, Category = "MMORPG|Error")
	FString GetUserMessage(const FMMORPGError& Error);
};