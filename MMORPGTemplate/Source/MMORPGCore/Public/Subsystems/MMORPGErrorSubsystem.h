#pragma once

#include "CoreMinimal.h"
#include "Subsystems/GameInstanceSubsystem.h"
#include "CoreTypes.h"
#include "MMORPGErrorSubsystem.generated.h"

/**
 * Centralized error handling subsystem
 * Manages error reporting, history, and propagation throughout the game
 */
UCLASS()
class MMORPGCORE_API UMMORPGErrorSubsystem : public UGameInstanceSubsystem
{
	GENERATED_BODY()

public:
	// Subsystem interface
	virtual void Initialize(FSubsystemCollectionBase& Collection) override;
	virtual void Deinitialize() override;

	/**
	 * Report an error to the system
	 * @param Error The error to report
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Error")
	void ReportError(const FMMORPGError& Error);

	/**
	 * Report an error with basic parameters
	 * @param Code Error code
	 * @param Message Error message
	 * @param Category Error category
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Error", meta = (DisplayName = "Report Error (Simple)"))
	void ReportErrorSimple(int32 Code, const FString& Message, EMMORPGErrorCategory Category);

	/**
	 * Clear all stored errors
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Error")
	void ClearErrors();

	/**
	 * Get recent errors from history
	 * @param Count Number of errors to retrieve
	 * @return Array of recent errors
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Error")
	TArray<FMMORPGError> GetRecentErrors(int32 Count = 10) const;

	/**
	 * Get the last error that occurred
	 * @param OutError The last error (if any)
	 * @return True if there was an error to retrieve
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Error")
	bool GetLastError(FMMORPGError& OutError) const;

	/**
	 * Event fired when an error is reported
	 */
	UPROPERTY(BlueprintAssignable, Category = "MMORPG|Error")
	FOnMMORPGError OnErrorReported;

	/**
	 * Check if we should retry based on error type
	 * @param Error The error to check
	 * @return True if retry is recommended
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Error")
	bool ShouldRetry(const FMMORPGError& Error) const;

private:
	// Error history storage
	TArray<FMMORPGError> ErrorHistory;

	// Thread safety for error operations
	mutable FCriticalSection ErrorMutex;

	// Maximum errors to keep in history
	static constexpr int32 MaxErrorHistory = 100;

	// Log the error to UE log system
	void LogError(const FMMORPGError& Error) const;
};