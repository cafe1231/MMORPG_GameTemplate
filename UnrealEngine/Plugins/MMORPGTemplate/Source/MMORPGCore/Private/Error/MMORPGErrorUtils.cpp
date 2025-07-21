// Copyright (c) 2024 MMORPG Template Project

#include "Error/MMORPGErrorUtils.h"
#include "MMORPGCore.h"
#include "Console/MMORPGDeveloperConsole.h"
#include "Engine/World.h"
#include "TimerManager.h"
#include "Framework/Application/SlateApplication.h"
#include "Widgets/Layout/SBox.h"
#include "Widgets/Text/STextBlock.h"
#include "Widgets/Input/SButton.h"

void UMMORPGErrorUtils::ReportError(int32 ErrorCode, const FString& Message, EMMORPGErrorSeverity Severity)
{
    if (UMMORPGErrorHandler* Handler = GetErrorHandler())
    {
        Handler->ReportErrorSimple(ErrorCode, Message, Severity);
    }
}

void UMMORPGErrorUtils::ReportNetworkError(int32 ErrorCode, const FString& Message, bool bCanRetry)
{
    if (UMMORPGErrorHandler* Handler = GetErrorHandler())
    {
        Handler->CreateError(ErrorCode)
            .WithMessage(Message)
            .WithCategory(EMMORPGErrorCategory::Network)
            .WithSeverity(EMMORPGErrorSeverity::Error)
            .CanRetry(bCanRetry)
            .WithUserAction(TEXT("Please check your internet connection"))
            .Report();
    }
}

void UMMORPGErrorUtils::ReportAuthError(int32 ErrorCode, const FString& Message)
{
    if (UMMORPGErrorHandler* Handler = GetErrorHandler())
    {
        FString UserAction;
        switch (ErrorCode)
        {
        case MMORPGErrorCodes::AuthInvalidCredentials:
            UserAction = TEXT("Please check your username and password");
            break;
        case MMORPGErrorCodes::AuthTokenExpired:
            UserAction = TEXT("Please log in again");
            break;
        case MMORPGErrorCodes::AuthAccountLocked:
            UserAction = TEXT("Please contact support");
            break;
        default:
            UserAction = TEXT("Please try logging in again");
            break;
        }
        
        Handler->CreateError(ErrorCode)
            .WithMessage(Message)
            .WithCategory(EMMORPGErrorCategory::Auth)
            .WithSeverity(EMMORPGErrorSeverity::Error)
            .WithUserAction(UserAction)
            .Report();
    }
}

void UMMORPGErrorUtils::ReportGameError(int32 ErrorCode, const FString& Message)
{
    if (UMMORPGErrorHandler* Handler = GetErrorHandler())
    {
        Handler->CreateError(ErrorCode)
            .WithMessage(Message)
            .WithCategory(EMMORPGErrorCategory::Game)
            .WithSeverity(EMMORPGErrorSeverity::Warning)
            .Report();
    }
}

FMMORPGError UMMORPGErrorUtils::CreateErrorFromResponse(int32 ResponseCode, const FString& ResponseBody)
{
    FMMORPGError Error;
    
    // Map HTTP status codes to error codes
    switch (ResponseCode)
    {
    case 400:
        Error.ErrorCode = MMORPGErrorCodes::ProtocolInvalidMessage;
        Error.Message = TEXT("Bad request");
        Error.Category = EMMORPGErrorCategory::Protocol;
        break;
    case 401:
        Error.ErrorCode = MMORPGErrorCodes::AuthPermissionDenied;
        Error.Message = TEXT("Unauthorized");
        Error.Category = EMMORPGErrorCategory::Auth;
        break;
    case 403:
        Error.ErrorCode = MMORPGErrorCodes::AuthPermissionDenied;
        Error.Message = TEXT("Forbidden");
        Error.Category = EMMORPGErrorCategory::Auth;
        break;
    case 404:
        Error.ErrorCode = MMORPGErrorCodes::GameResourceNotFound;
        Error.Message = TEXT("Resource not found");
        Error.Category = EMMORPGErrorCategory::Network;
        break;
    case 500:
    case 502:
    case 503:
        Error.ErrorCode = MMORPGErrorCodes::NetworkServerUnreachable;
        Error.Message = TEXT("Server error");
        Error.Category = EMMORPGErrorCategory::Network;
        Error.bCanRetry = true;
        break;
    case 504:
        Error.ErrorCode = MMORPGErrorCodes::NetworkTimeout;
        Error.Message = TEXT("Gateway timeout");
        Error.Category = EMMORPGErrorCategory::Network;
        Error.bCanRetry = true;
        break;
    default:
        Error.ErrorCode = ResponseCode;
        Error.Message = FString::Printf(TEXT("HTTP Error %d"), ResponseCode);
        Error.Category = EMMORPGErrorCategory::Network;
        break;
    }
    
    Error.Details = ResponseBody;
    Error.Severity = (ResponseCode >= 500) ? EMMORPGErrorSeverity::Critical : EMMORPGErrorSeverity::Error;
    Error.Timestamp = FDateTime::Now();
    
    return Error;
}

UMMORPGErrorHandler* UMMORPGErrorUtils::GetErrorHandler()
{
    if (FMMORPGCoreModule::IsAvailable())
    {
        return FMMORPGCoreModule::Get().GetErrorHandler();
    }
    return nullptr;
}

bool UMMORPGErrorUtils::IsRetryableError(const FMMORPGError& Error)
{
    if (UMMORPGErrorHandler* Handler = GetErrorHandler())
    {
        return Handler->CanRetryError(Error);
    }
    return Error.bCanRetry;
}

FText UMMORPGErrorUtils::GetLocalizedErrorMessage(const FMMORPGError& Error)
{
    // TODO: Implement proper localization
    FString Message = Error.Message;
    
    if (!Error.UserAction.IsEmpty())
    {
        Message += TEXT("\n\n") + Error.UserAction;
    }
    
    return FText::FromString(Message);
}

void UMMORPGErrorUtils::ShowErrorDialog(const FMMORPGError& Error, bool bAllowRetry)
{
#if WITH_EDITOR
    // Create simple error dialog
    TSharedRef<SWindow> ErrorWindow = SNew(SWindow)
        .Title(FText::FromString(TEXT("Error")))
        .SizingRule(ESizingRule::Autosized)
        .SupportsMinimize(false)
        .SupportsMaximize(false);
    
    TSharedPtr<SBox> DialogContent;
    
    SAssignNew(DialogContent, SBox)
        .Padding(FMargin(20))
        .MinDesiredWidth(400)
        [
            SNew(SVerticalBox)
            + SVerticalBox::Slot()
            .AutoHeight()
            .Padding(FMargin(0, 0, 0, 10))
            [
                SNew(STextBlock)
                .Text(GetLocalizedErrorMessage(Error))
                .WrapTextAt(380)
            ]
            + SVerticalBox::Slot()
            .AutoHeight()
            .HAlign(HAlign_Right)
            [
                SNew(SHorizontalBox)
                + SHorizontalBox::Slot()
                .AutoWidth()
                .Padding(FMargin(0, 0, 10, 0))
                [
                    SNew(SButton)
                    .Text(FText::FromString(TEXT("OK")))
                    .OnClicked_Lambda([ErrorWindow]()
                    {
                        ErrorWindow->RequestDestroyWindow();
                        return FReply::Handled();
                    })
                ]
            ]
        ];
    
    ErrorWindow->SetContent(DialogContent.ToSharedRef());
    
    FSlateApplication::Get().AddModalWindow(ErrorWindow, nullptr);
#endif
}

void UMMORPGErrorUtils::LogError(const FMMORPGError& Error)
{
    // Log to developer console if available
    if (UMMORPGDeveloperConsole* Console = FMMORPGCoreModule::Get().GetDeveloperConsole())
    {
        FColor Color;
        switch (Error.Severity)
        {
        case EMMORPGErrorSeverity::Info:
            Color = FColor::White;
            break;
        case EMMORPGErrorSeverity::Warning:
            Color = FColor::Yellow;
            break;
        case EMMORPGErrorSeverity::Error:
            Color = FColor::Red;
            break;
        case EMMORPGErrorSeverity::Critical:
        case EMMORPGErrorSeverity::Fatal:
            Color = FColor::Red;
            break;
        default:
            Color = FColor::White;
            break;
        }
        
        FString Output = FString::Printf(TEXT("[ERROR %d] %s"), Error.ErrorCode, *Error.Message);
        Console->WriteOutput(Output, Color);
    }
}

// Retry Policy Implementation

float FMMORPGRetryPolicy::GetDelayForAttempt(int32 AttemptNumber) const
{
    if (AttemptNumber <= 0)
    {
        return 0.0f;
    }
    
    float Delay = InitialDelay;
    
    if (bUseExponentialBackoff)
    {
        Delay = InitialDelay * FMath::Pow(BackoffMultiplier, AttemptNumber - 1);
    }
    
    return FMath::Min(Delay, MaxDelay);
}

// Retry Handler Implementation

void UMMORPGRetryHandler::ExecuteWithRetry(const FMMORPGRetryPolicy& Policy, TFunction<bool()> Action, TFunction<void(bool)> OnComplete)
{
    if (bIsRetrying)
    {
        UE_LOG(LogMMORPGError, Warning, TEXT("Retry already in progress"));
        return;
    }
    
    CurrentPolicy = Policy;
    CurrentAction = Action;
    CurrentCallback = OnComplete;
    CurrentAttempt = 0;
    bIsRetrying = true;
    
    // Execute first attempt immediately
    HandleRetryTimer();
}

void UMMORPGRetryHandler::CancelRetry()
{
    if (bIsRetrying)
    {
        if (UWorld* World = GetWorld())
        {
            World->GetTimerManager().ClearTimer(RetryTimerHandle);
        }
        
        bIsRetrying = false;
        CurrentAttempt = 0;
        
        if (CurrentCallback)
        {
            CurrentCallback(false);
        }
    }
}

void UMMORPGRetryHandler::HandleRetryTimer()
{
    if (!bIsRetrying || !CurrentAction)
    {
        return;
    }
    
    CurrentAttempt++;
    
    // Execute the action
    bool bSuccess = CurrentAction();
    
    if (bSuccess)
    {
        // Success!
        bIsRetrying = false;
        if (CurrentCallback)
        {
            CurrentCallback(true);
        }
    }
    else if (CurrentAttempt < CurrentPolicy.MaxAttempts)
    {
        // Schedule retry
        float Delay = CurrentPolicy.GetDelayForAttempt(CurrentAttempt);
        
        UE_LOG(LogMMORPGError, Log, TEXT("Retrying in %.1f seconds (attempt %d/%d)"), 
            Delay, CurrentAttempt + 1, CurrentPolicy.MaxAttempts);
        
        if (UWorld* World = GetWorld())
        {
            World->GetTimerManager().SetTimer(RetryTimerHandle, this, &UMMORPGRetryHandler::HandleRetryTimer, Delay, false);
        }
    }
    else
    {
        // Max attempts reached
        bIsRetrying = false;
        UE_LOG(LogMMORPGError, Error, TEXT("Max retry attempts reached (%d)"), CurrentPolicy.MaxAttempts);
        
        if (CurrentCallback)
        {
            CurrentCallback(false);
        }
    }
}