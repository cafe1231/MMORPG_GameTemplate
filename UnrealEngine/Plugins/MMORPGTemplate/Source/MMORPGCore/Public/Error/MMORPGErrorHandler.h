// Copyright (c) 2024 MMORPG Template Project

#pragma once

#include "CoreMinimal.h"
#include "MMORPGErrorHandler.generated.h"

DECLARE_LOG_CATEGORY_EXTERN(LogMMORPGError, Log, All);

/**
 * Error severity levels
 */
UENUM(BlueprintType)
enum class EMMORPGErrorSeverity : uint8
{
    Info         UMETA(DisplayName = "Info"),
    Warning      UMETA(DisplayName = "Warning"),
    Error        UMETA(DisplayName = "Error"),
    Critical     UMETA(DisplayName = "Critical"),
    Fatal        UMETA(DisplayName = "Fatal")
};

/**
 * Error categories for grouping related errors
 */
UENUM(BlueprintType)
enum class EMMORPGErrorCategory : uint8
{
    Network      UMETA(DisplayName = "Network"),
    Auth         UMETA(DisplayName = "Authentication"),
    Game         UMETA(DisplayName = "Game Logic"),
    Protocol     UMETA(DisplayName = "Protocol"),
    Database     UMETA(DisplayName = "Database"),
    System       UMETA(DisplayName = "System"),
    Unknown      UMETA(DisplayName = "Unknown")
};

/**
 * Error structure containing all error information
 */
USTRUCT(BlueprintType)
struct MMORPGCORE_API FMMORPGError
{
    GENERATED_BODY()

    /** Unique error code */
    UPROPERTY(BlueprintReadOnly)
    int32 ErrorCode;

    /** Error severity */
    UPROPERTY(BlueprintReadOnly)
    EMMORPGErrorSeverity Severity;

    /** Error category */
    UPROPERTY(BlueprintReadOnly)
    EMMORPGErrorCategory Category;

    /** Human-readable error message */
    UPROPERTY(BlueprintReadOnly)
    FString Message;

    /** Technical details for debugging */
    UPROPERTY(BlueprintReadOnly)
    FString Details;

    /** Timestamp when error occurred */
    UPROPERTY(BlueprintReadOnly)
    FDateTime Timestamp;

    /** Context information (e.g., function name, file) */
    UPROPERTY(BlueprintReadOnly)
    FString Context;

    /** Suggested action for the user */
    UPROPERTY(BlueprintReadOnly)
    FString UserAction;

    /** Whether this error can be retried */
    UPROPERTY(BlueprintReadOnly)
    bool bCanRetry;

    /** Maximum retry attempts */
    UPROPERTY(BlueprintReadOnly)
    int32 MaxRetries;

    FMMORPGError()
    {
        ErrorCode = 0;
        Severity = EMMORPGErrorSeverity::Error;
        Category = EMMORPGErrorCategory::Unknown;
        Timestamp = FDateTime::Now();
        bCanRetry = false;
        MaxRetries = 3;
    }
};

/**
 * Delegate for error notifications
 */
DECLARE_MULTICAST_DELEGATE_OneParam(FOnMMORPGError, const FMMORPGError&);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnMMORPGErrorBP, const FMMORPGError&, Error);

/**
 * Error handler system for the MMORPG Template
 */
UCLASS(BlueprintType)
class MMORPGCORE_API UMMORPGErrorHandler : public UObject
{
    GENERATED_BODY()

public:
    UMMORPGErrorHandler();

    /** Initialize the error handler */
    void Initialize();

    /** Shutdown the error handler */
    void Shutdown();

    /** Report an error */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error", meta = (CallInEditor = "true"))
    void ReportError(const FMMORPGError& Error);

    /** Create and report an error */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error", meta = (CallInEditor = "true"))
    void ReportErrorSimple(int32 ErrorCode, const FString& Message, EMMORPGErrorSeverity Severity = EMMORPGErrorSeverity::Error);

    /** Create error from exception */
    void ReportException(const FString& Context, const FString& Exception);

    /** Get error message for error code */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error")
    FString GetErrorMessage(int32 ErrorCode) const;

    /** Get user-friendly message for error */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error")
    FString GetUserFriendlyMessage(const FMMORPGError& Error) const;

    /** Check if error can be retried */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error")
    bool CanRetryError(const FMMORPGError& Error) const;

    /** Get recent errors */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error")
    TArray<FMMORPGError> GetRecentErrors(int32 Count = 10) const;

    /** Clear error history */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error")
    void ClearErrorHistory();

    /** Set whether to show error notifications */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error")
    void SetShowNotifications(bool bShow) { bShowNotifications = bShow; }

    /** Error occurred delegate */
    FOnMMORPGError OnErrorOccurred;

    /** Blueprint-friendly error delegate */
    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Error")
    FOnMMORPGErrorBP OnErrorOccurredBP;

    /** Create error builder for fluent API */
    class FErrorBuilder
    {
    public:
        FErrorBuilder(UMMORPGErrorHandler* InHandler, int32 InCode)
            : Handler(InHandler)
        {
            Error.ErrorCode = InCode;
        }

        FErrorBuilder& WithMessage(const FString& Message)
        {
            Error.Message = Message;
            return *this;
        }

        FErrorBuilder& WithDetails(const FString& Details)
        {
            Error.Details = Details;
            return *this;
        }

        FErrorBuilder& WithSeverity(EMMORPGErrorSeverity Severity)
        {
            Error.Severity = Severity;
            return *this;
        }

        FErrorBuilder& WithCategory(EMMORPGErrorCategory Category)
        {
            Error.Category = Category;
            return *this;
        }

        FErrorBuilder& WithContext(const FString& Context)
        {
            Error.Context = Context;
            return *this;
        }

        FErrorBuilder& WithUserAction(const FString& Action)
        {
            Error.UserAction = Action;
            return *this;
        }

        FErrorBuilder& CanRetry(bool bRetry = true, int32 MaxAttempts = 3)
        {
            Error.bCanRetry = bRetry;
            Error.MaxRetries = MaxAttempts;
            return *this;
        }

        void Report()
        {
            if (Handler)
            {
                Handler->ReportError(Error);
            }
        }

    private:
        UMMORPGErrorHandler* Handler;
        FMMORPGError Error;
    };

    /** Create error builder */
    FErrorBuilder CreateError(int32 ErrorCode)
    {
        return FErrorBuilder(this, ErrorCode);
    }

protected:
    /** Handle different error severities */
    void HandleErrorBySeverity(const FMMORPGError& Error);

    /** Log error to file */
    void LogErrorToFile(const FMMORPGError& Error);

    /** Show error notification */
    void ShowErrorNotification(const FMMORPGError& Error);

    /** Send error telemetry */
    void SendErrorTelemetry(const FMMORPGError& Error);

private:
    /** Error history */
    UPROPERTY()
    TArray<FMMORPGError> ErrorHistory;

    /** Maximum errors to keep in history */
    UPROPERTY()
    int32 MaxErrorHistory;

    /** Whether to show notifications */
    UPROPERTY()
    bool bShowNotifications;

    /** Error message lookup table */
    TMap<int32, FString> ErrorMessages;

    /** Initialize error messages */
    void InitializeErrorMessages();
};

/**
 * Error codes used throughout the system
 */
namespace MMORPGErrorCodes
{
    // Network errors (1000-1999)
    constexpr int32 NetworkConnectionFailed = 1001;
    constexpr int32 NetworkTimeout = 1002;
    constexpr int32 NetworkDisconnected = 1003;
    constexpr int32 NetworkInvalidResponse = 1004;
    constexpr int32 NetworkServerUnreachable = 1005;

    // Authentication errors (2000-2999)
    constexpr int32 AuthInvalidCredentials = 2001;
    constexpr int32 AuthTokenExpired = 2002;
    constexpr int32 AuthAccountLocked = 2003;
    constexpr int32 AuthSessionInvalid = 2004;
    constexpr int32 AuthPermissionDenied = 2005;

    // Protocol errors (3000-3999)
    constexpr int32 ProtocolVersionMismatch = 3001;
    constexpr int32 ProtocolInvalidMessage = 3002;
    constexpr int32 ProtocolSerializationFailed = 3003;
    constexpr int32 ProtocolDeserializationFailed = 3004;

    // Game logic errors (4000-4999)
    constexpr int32 GameInvalidOperation = 4001;
    constexpr int32 GameStateCorrupted = 4002;
    constexpr int32 GameResourceNotFound = 4003;
    constexpr int32 GameActionNotAllowed = 4004;

    // System errors (5000-5999)
    constexpr int32 SystemOutOfMemory = 5001;
    constexpr int32 SystemFileNotFound = 5002;
    constexpr int32 SystemPermissionDenied = 5003;
    constexpr int32 SystemInitializationFailed = 5004;
}

/**
 * Helper macros for error reporting
 */
#define MMORPG_ERROR(Code, Message) \
    if (UMMORPGErrorHandler* Handler = FMMORPGCoreModule::Get().GetErrorHandler()) \
    { \
        Handler->CreateError(Code).WithMessage(Message).WithContext(FString::Printf(TEXT("%s:%d"), TEXT(__FILE__), __LINE__)).Report(); \
    }

#define MMORPG_ERROR_DETAILED(Code, Message, Details) \
    if (UMMORPGErrorHandler* Handler = FMMORPGCoreModule::Get().GetErrorHandler()) \
    { \
        Handler->CreateError(Code).WithMessage(Message).WithDetails(Details).WithContext(FString::Printf(TEXT("%s:%d"), TEXT(__FILE__), __LINE__)).Report(); \
    }