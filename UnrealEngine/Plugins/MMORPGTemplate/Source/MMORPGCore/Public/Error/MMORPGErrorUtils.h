// Copyright (c) 2024 MMORPG Template Project

#pragma once

#include "CoreMinimal.h"
#include "Error/MMORPGErrorHandler.h"
#include "Kismet/BlueprintFunctionLibrary.h"
#include "MMORPGErrorUtils.generated.h"

/**
 * Blueprint utilities for error handling
 */
UCLASS()
class MMORPGCORE_API UMMORPGErrorUtils : public UBlueprintFunctionLibrary
{
    GENERATED_BODY()

public:
    /** Report an error with simple parameters */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error", meta = (DisplayName = "Report Error"))
    static void ReportError(int32 ErrorCode, const FString& Message, EMMORPGErrorSeverity Severity = EMMORPGErrorSeverity::Error);

    /** Report a network error */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error")
    static void ReportNetworkError(int32 ErrorCode, const FString& Message, bool bCanRetry = true);

    /** Report an authentication error */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error")
    static void ReportAuthError(int32 ErrorCode, const FString& Message);

    /** Report a game logic error */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error")
    static void ReportGameError(int32 ErrorCode, const FString& Message);

    /** Create error from network response */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error")
    static FMMORPGError CreateErrorFromResponse(int32 ResponseCode, const FString& ResponseBody);

    /** Get error handler instance */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error", meta = (DisplayName = "Get Error Handler"))
    static UMMORPGErrorHandler* GetErrorHandler();

    /** Check if error is retryable */
    UFUNCTION(BlueprintPure, Category = "MMORPG|Error")
    static bool IsRetryableError(const FMMORPGError& Error);

    /** Get localized error message */
    UFUNCTION(BlueprintPure, Category = "MMORPG|Error")
    static FText GetLocalizedErrorMessage(const FMMORPGError& Error);

    /** Show error dialog */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error", meta = (CallInEditor = "true"))
    static void ShowErrorDialog(const FMMORPGError& Error, bool bAllowRetry = false);

    /** Log error to console */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error")
    static void LogError(const FMMORPGError& Error);
};

/**
 * Retry policy for handling retryable errors
 */
USTRUCT(BlueprintType)
struct MMORPGCORE_API FMMORPGRetryPolicy
{
    GENERATED_BODY()

    /** Maximum number of retry attempts */
    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    int32 MaxAttempts = 3;

    /** Initial delay before first retry (in seconds) */
    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    float InitialDelay = 1.0f;

    /** Maximum delay between retries (in seconds) */
    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    float MaxDelay = 30.0f;

    /** Backoff multiplier for exponential backoff */
    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    float BackoffMultiplier = 2.0f;

    /** Whether to use exponential backoff */
    UPROPERTY(BlueprintReadWrite, EditAnywhere)
    bool bUseExponentialBackoff = true;

    /** Calculate delay for given attempt number */
    float GetDelayForAttempt(int32 AttemptNumber) const;
};

/**
 * Retry handler for managing retry logic
 */
UCLASS(BlueprintType)
class MMORPGCORE_API UMMORPGRetryHandler : public UObject
{
    GENERATED_BODY()

public:
    /** Execute action with retry logic */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error", meta = (DisplayName = "Execute With Retry"))
    void ExecuteWithRetry(const FMMORPGRetryPolicy& Policy, TFunction<bool()> Action, TFunction<void(bool)> OnComplete);

    /** Cancel ongoing retry */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error")
    void CancelRetry();

    /** Check if retry is in progress */
    UFUNCTION(BlueprintPure, Category = "MMORPG|Error")
    bool IsRetrying() const { return bIsRetrying; }

private:
    /** Handle retry timer */
    void HandleRetryTimer();

    /** Current retry state */
    bool bIsRetrying = false;
    int32 CurrentAttempt = 0;
    FMMORPGRetryPolicy CurrentPolicy;
    TFunction<bool()> CurrentAction;
    TFunction<void(bool)> CurrentCallback;
    FTimerHandle RetryTimerHandle;
};