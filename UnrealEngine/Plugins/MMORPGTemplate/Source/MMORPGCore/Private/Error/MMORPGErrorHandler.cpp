// Copyright (c) 2024 MMORPG Template Project

#include "Error/MMORPGErrorHandler.h"
#include "MMORPGCore.h"
#include "Engine/Engine.h"
#include "Misc/DateTime.h"
#include "HAL/PlatformFilemanager.h"
#include "Misc/FileHelper.h"
#include "Misc/Paths.h"
#include "Framework/Notifications/NotificationManager.h"
#include "Widgets/Notifications/SNotificationList.h"

DEFINE_LOG_CATEGORY(LogMMORPGError);

UMMORPGErrorHandler::UMMORPGErrorHandler()
    : MaxErrorHistory(100)
    , bShowNotifications(true)
{
}

void UMMORPGErrorHandler::Initialize()
{
    InitializeErrorMessages();
    
    UE_LOG(LogMMORPGError, Log, TEXT("Error handler initialized"));
}

void UMMORPGErrorHandler::Shutdown()
{
    // Save error log if needed
    if (ErrorHistory.Num() > 0)
    {
        FString LogPath = FPaths::ProjectLogDir() / TEXT("MMORPGErrors.log");
        FString LogContent;
        
        for (const FMMORPGError& Error : ErrorHistory)
        {
            LogContent += FString::Printf(TEXT("[%s] [%s] %s: %s\n"),
                *Error.Timestamp.ToString(),
                *UEnum::GetValueAsString(Error.Severity),
                *Error.Context,
                *Error.Message
            );
        }
        
        FFileHelper::SaveStringToFile(LogContent, *LogPath);
    }
    
    ErrorHistory.Empty();
    ErrorMessages.Empty();
    
    UE_LOG(LogMMORPGError, Log, TEXT("Error handler shutdown"));
}

void UMMORPGErrorHandler::ReportError(const FMMORPGError& Error)
{
    // Add to history
    ErrorHistory.Add(Error);
    
    // Trim history if needed
    if (ErrorHistory.Num() > MaxErrorHistory)
    {
        ErrorHistory.RemoveAt(0, ErrorHistory.Num() - MaxErrorHistory);
    }
    
    // Handle based on severity
    HandleErrorBySeverity(Error);
    
    // Log to file
    LogErrorToFile(Error);
    
    // Show notification if enabled
    if (bShowNotifications)
    {
        ShowErrorNotification(Error);
    }
    
    // Send telemetry
    SendErrorTelemetry(Error);
    
    // Broadcast delegates
    OnErrorOccurred.Broadcast(Error);
    OnErrorOccurredBP.Broadcast(Error);
}

void UMMORPGErrorHandler::ReportErrorSimple(int32 ErrorCode, const FString& Message, EMMORPGErrorSeverity Severity)
{
    FMMORPGError Error;
    Error.ErrorCode = ErrorCode;
    Error.Message = Message;
    Error.Severity = Severity;
    Error.Category = EMMORPGErrorCategory::Unknown;
    Error.Timestamp = FDateTime::Now();
    
    // Try to determine category from error code
    if (ErrorCode >= 1000 && ErrorCode < 2000)
    {
        Error.Category = EMMORPGErrorCategory::Network;
    }
    else if (ErrorCode >= 2000 && ErrorCode < 3000)
    {
        Error.Category = EMMORPGErrorCategory::Auth;
    }
    else if (ErrorCode >= 3000 && ErrorCode < 4000)
    {
        Error.Category = EMMORPGErrorCategory::Protocol;
    }
    else if (ErrorCode >= 4000 && ErrorCode < 5000)
    {
        Error.Category = EMMORPGErrorCategory::Game;
    }
    else if (ErrorCode >= 5000 && ErrorCode < 6000)
    {
        Error.Category = EMMORPGErrorCategory::System;
    }
    
    ReportError(Error);
}

void UMMORPGErrorHandler::ReportException(const FString& Context, const FString& Exception)
{
    FMMORPGError Error;
    Error.ErrorCode = 0;
    Error.Severity = EMMORPGErrorSeverity::Critical;
    Error.Category = EMMORPGErrorCategory::System;
    Error.Message = TEXT("Exception occurred");
    Error.Details = Exception;
    Error.Context = Context;
    Error.Timestamp = FDateTime::Now();
    Error.UserAction = TEXT("Please restart the application");
    
    ReportError(Error);
}

FString UMMORPGErrorHandler::GetErrorMessage(int32 ErrorCode) const
{
    if (const FString* Message = ErrorMessages.Find(ErrorCode))
    {
        return *Message;
    }
    
    return FString::Printf(TEXT("Unknown error (Code: %d)"), ErrorCode);
}

FString UMMORPGErrorHandler::GetUserFriendlyMessage(const FMMORPGError& Error) const
{
    FString Message;
    
    switch (Error.Category)
    {
    case EMMORPGErrorCategory::Network:
        Message = TEXT("Network error: ");
        break;
    case EMMORPGErrorCategory::Auth:
        Message = TEXT("Authentication error: ");
        break;
    case EMMORPGErrorCategory::Game:
        Message = TEXT("Game error: ");
        break;
    default:
        Message = TEXT("Error: ");
        break;
    }
    
    Message += Error.Message;
    
    if (!Error.UserAction.IsEmpty())
    {
        Message += TEXT("\n") + Error.UserAction;
    }
    
    return Message;
}

bool UMMORPGErrorHandler::CanRetryError(const FMMORPGError& Error) const
{
    // Network errors are usually retryable
    if (Error.Category == EMMORPGErrorCategory::Network && Error.bCanRetry)
    {
        return true;
    }
    
    // Some auth errors can be retried
    if (Error.ErrorCode == MMORPGErrorCodes::AuthTokenExpired)
    {
        return true;
    }
    
    return Error.bCanRetry;
}

TArray<FMMORPGError> UMMORPGErrorHandler::GetRecentErrors(int32 Count) const
{
    TArray<FMMORPGError> RecentErrors;
    
    int32 StartIndex = FMath::Max(0, ErrorHistory.Num() - Count);
    for (int32 i = StartIndex; i < ErrorHistory.Num(); i++)
    {
        RecentErrors.Add(ErrorHistory[i]);
    }
    
    return RecentErrors;
}

void UMMORPGErrorHandler::ClearErrorHistory()
{
    ErrorHistory.Empty();
    UE_LOG(LogMMORPGError, Log, TEXT("Error history cleared"));
}

void UMMORPGErrorHandler::HandleErrorBySeverity(const FMMORPGError& Error)
{
    switch (Error.Severity)
    {
    case EMMORPGErrorSeverity::Info:
        UE_LOG(LogMMORPGError, Display, TEXT("[%s] %s"), *Error.Context, *Error.Message);
        break;
        
    case EMMORPGErrorSeverity::Warning:
        UE_LOG(LogMMORPGError, Warning, TEXT("[%s] %s"), *Error.Context, *Error.Message);
        break;
        
    case EMMORPGErrorSeverity::Error:
        UE_LOG(LogMMORPGError, Error, TEXT("[%s] %s"), *Error.Context, *Error.Message);
        if (!Error.Details.IsEmpty())
        {
            UE_LOG(LogMMORPGError, Error, TEXT("Details: %s"), *Error.Details);
        }
        break;
        
    case EMMORPGErrorSeverity::Critical:
        UE_LOG(LogMMORPGError, Error, TEXT("CRITICAL: [%s] %s"), *Error.Context, *Error.Message);
        if (!Error.Details.IsEmpty())
        {
            UE_LOG(LogMMORPGError, Error, TEXT("Details: %s"), *Error.Details);
        }
        break;
        
    case EMMORPGErrorSeverity::Fatal:
        UE_LOG(LogMMORPGError, Fatal, TEXT("FATAL: [%s] %s"), *Error.Context, *Error.Message);
        break;
    }
}

void UMMORPGErrorHandler::LogErrorToFile(const FMMORPGError& Error)
{
    // Only log errors and above to file
    if (Error.Severity < EMMORPGErrorSeverity::Error)
    {
        return;
    }
    
    FString LogDir = FPaths::ProjectLogDir() / TEXT("MMORPG");
    IPlatformFile& PlatformFile = FPlatformFileManager::Get().GetPlatformFile();
    
    if (!PlatformFile.DirectoryExists(*LogDir))
    {
        PlatformFile.CreateDirectory(*LogDir);
    }
    
    FString DateStr = FDateTime::Now().ToString(TEXT("%Y%m%d"));
    FString LogPath = LogDir / FString::Printf(TEXT("Errors_%s.log"), *DateStr);
    
    FString LogEntry = FString::Printf(TEXT("[%s] [%s] [%s] Code:%d - %s"),
        *Error.Timestamp.ToString(),
        *UEnum::GetValueAsString(Error.Severity),
        *UEnum::GetValueAsString(Error.Category),
        Error.ErrorCode,
        *Error.Message
    );
    
    if (!Error.Details.IsEmpty())
    {
        LogEntry += FString::Printf(TEXT("\n  Details: %s"), *Error.Details);
    }
    
    if (!Error.Context.IsEmpty())
    {
        LogEntry += FString::Printf(TEXT("\n  Context: %s"), *Error.Context);
    }
    
    LogEntry += TEXT("\n");
    
    FFileHelper::SaveStringArrayToFile(TArray<FString>{LogEntry}, *LogPath, 
        FFileHelper::EEncodingOptions::AutoDetect, &IFileManager::Get(), EFileWrite::FILEWRITE_Append);
}

void UMMORPGErrorHandler::ShowErrorNotification(const FMMORPGError& Error)
{
#if WITH_EDITOR
    if (Error.Severity >= EMMORPGErrorSeverity::Error)
    {
        FNotificationInfo Info(FText::FromString(GetUserFriendlyMessage(Error)));
        Info.ExpireDuration = 5.0f;
        
        switch (Error.Severity)
        {
        case EMMORPGErrorSeverity::Warning:
            Info.Image = FCoreStyle::Get().GetBrush(TEXT("MessageLog.Warning"));
            break;
        case EMMORPGErrorSeverity::Error:
        case EMMORPGErrorSeverity::Critical:
            Info.Image = FCoreStyle::Get().GetBrush(TEXT("MessageLog.Error"));
            break;
        default:
            Info.Image = FCoreStyle::Get().GetBrush(TEXT("MessageLog.Note"));
            break;
        }
        
        FSlateNotificationManager::Get().AddNotification(Info);
    }
#endif
}

void UMMORPGErrorHandler::SendErrorTelemetry(const FMMORPGError& Error)
{
    // TODO: Implement telemetry sending
    // This would send error data to analytics service
}

void UMMORPGErrorHandler::InitializeErrorMessages()
{
    // Network errors
    ErrorMessages.Add(MMORPGErrorCodes::NetworkConnectionFailed, TEXT("Failed to connect to server"));
    ErrorMessages.Add(MMORPGErrorCodes::NetworkTimeout, TEXT("Connection timed out"));
    ErrorMessages.Add(MMORPGErrorCodes::NetworkDisconnected, TEXT("Disconnected from server"));
    ErrorMessages.Add(MMORPGErrorCodes::NetworkInvalidResponse, TEXT("Invalid response from server"));
    ErrorMessages.Add(MMORPGErrorCodes::NetworkServerUnreachable, TEXT("Server is unreachable"));
    
    // Authentication errors
    ErrorMessages.Add(MMORPGErrorCodes::AuthInvalidCredentials, TEXT("Invalid username or password"));
    ErrorMessages.Add(MMORPGErrorCodes::AuthTokenExpired, TEXT("Session has expired"));
    ErrorMessages.Add(MMORPGErrorCodes::AuthAccountLocked, TEXT("Account is locked"));
    ErrorMessages.Add(MMORPGErrorCodes::AuthSessionInvalid, TEXT("Invalid session"));
    ErrorMessages.Add(MMORPGErrorCodes::AuthPermissionDenied, TEXT("Permission denied"));
    
    // Protocol errors
    ErrorMessages.Add(MMORPGErrorCodes::ProtocolVersionMismatch, TEXT("Client version incompatible with server"));
    ErrorMessages.Add(MMORPGErrorCodes::ProtocolInvalidMessage, TEXT("Invalid message format"));
    ErrorMessages.Add(MMORPGErrorCodes::ProtocolSerializationFailed, TEXT("Failed to serialize message"));
    ErrorMessages.Add(MMORPGErrorCodes::ProtocolDeserializationFailed, TEXT("Failed to deserialize message"));
    
    // Game logic errors
    ErrorMessages.Add(MMORPGErrorCodes::GameInvalidOperation, TEXT("Invalid operation"));
    ErrorMessages.Add(MMORPGErrorCodes::GameStateCorrupted, TEXT("Game state is corrupted"));
    ErrorMessages.Add(MMORPGErrorCodes::GameResourceNotFound, TEXT("Resource not found"));
    ErrorMessages.Add(MMORPGErrorCodes::GameActionNotAllowed, TEXT("Action not allowed"));
    
    // System errors
    ErrorMessages.Add(MMORPGErrorCodes::SystemOutOfMemory, TEXT("Out of memory"));
    ErrorMessages.Add(MMORPGErrorCodes::SystemFileNotFound, TEXT("File not found"));
    ErrorMessages.Add(MMORPGErrorCodes::SystemPermissionDenied, TEXT("Permission denied"));
    ErrorMessages.Add(MMORPGErrorCodes::SystemInitializationFailed, TEXT("Initialization failed"));
}