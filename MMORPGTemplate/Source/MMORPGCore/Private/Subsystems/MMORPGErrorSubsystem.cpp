#include "Subsystems/MMORPGErrorSubsystem.h"
#include "Engine/Engine.h"

// Define log category for errors
DEFINE_LOG_CATEGORY_STATIC(LogMMORPGError, Log, All);

void UMMORPGErrorSubsystem::Initialize(FSubsystemCollectionBase& Collection)
{
	Super::Initialize(Collection);
	
	UE_LOG(LogMMORPGError, Log, TEXT("MMORPGErrorSubsystem initialized"));
}

void UMMORPGErrorSubsystem::Deinitialize()
{
	ClearErrors();
	
	Super::Deinitialize();
}

void UMMORPGErrorSubsystem::ReportError(const FMMORPGError& Error)
{
	FScopeLock Lock(&ErrorMutex);
	
	// Add to history
	ErrorHistory.Add(Error);
	
	// Maintain max history size
	if (ErrorHistory.Num() > MaxErrorHistory)
	{
		ErrorHistory.RemoveAt(0, ErrorHistory.Num() - MaxErrorHistory);
	}
	
	// Log the error
	LogError(Error);
	
	// Fire event (must be on game thread)
	if (IsInGameThread())
	{
		OnErrorReported.Broadcast(Error);
	}
	else
	{
		FMMORPGError ErrorCopy = Error;
		AsyncTask(ENamedThreads::GameThread, [this, ErrorCopy]()
		{
			OnErrorReported.Broadcast(ErrorCopy);
		});
	}
}

void UMMORPGErrorSubsystem::ReportErrorSimple(int32 Code, const FString& Message, EMMORPGErrorCategory Category)
{
	FMMORPGError Error(Code, Message, Category);
	ReportError(Error);
}

void UMMORPGErrorSubsystem::ClearErrors()
{
	FScopeLock Lock(&ErrorMutex);
	ErrorHistory.Empty();
}

TArray<FMMORPGError> UMMORPGErrorSubsystem::GetRecentErrors(int32 Count) const
{
	FScopeLock Lock(&ErrorMutex);
	
	TArray<FMMORPGError> Result;
	
	int32 StartIndex = FMath::Max(0, ErrorHistory.Num() - Count);
	for (int32 i = StartIndex; i < ErrorHistory.Num(); i++)
	{
		Result.Add(ErrorHistory[i]);
	}
	
	return Result;
}

bool UMMORPGErrorSubsystem::GetLastError(FMMORPGError& OutError) const
{
	FScopeLock Lock(&ErrorMutex);
	
	if (ErrorHistory.Num() > 0)
	{
		OutError = ErrorHistory.Last();
		return true;
	}
	
	return false;
}

bool UMMORPGErrorSubsystem::ShouldRetry(const FMMORPGError& Error) const
{
	// Network errors are generally retryable
	if (Error.Category == EMMORPGErrorCategory::Network)
	{
		// Timeout and connection errors are retryable
		if (Error.Code >= 1000 && Error.Code < 1100)
		{
			return true;
		}
	}
	
	// Auth errors might be retryable (token refresh)
	if (Error.Category == EMMORPGErrorCategory::Auth)
	{
		// Token expired is retryable
		if (Error.Code == 2001)
		{
			return true;
		}
	}
	
	// Most other errors are not retryable
	return false;
}

void UMMORPGErrorSubsystem::LogError(const FMMORPGError& Error) const
{
	const FString SeverityStr = UEnum::GetValueAsString(Error.Severity);
	const FString CategoryStr = UEnum::GetValueAsString(Error.Category);
	
	switch (Error.Severity)
	{
	case EMMORPGErrorSeverity::Info:
		UE_LOG(LogMMORPGError, Log, TEXT("[%s] %s - Code: %d, Message: %s"),
			*CategoryStr, *SeverityStr, Error.Code, *Error.Message);
		break;
		
	case EMMORPGErrorSeverity::Warning:
		UE_LOG(LogMMORPGError, Warning, TEXT("[%s] %s - Code: %d, Message: %s"),
			*CategoryStr, *SeverityStr, Error.Code, *Error.Message);
		break;
		
	case EMMORPGErrorSeverity::Error:
		UE_LOG(LogMMORPGError, Error, TEXT("[%s] %s - Code: %d, Message: %s"),
			*CategoryStr, *SeverityStr, Error.Code, *Error.Message);
		break;
		
	case EMMORPGErrorSeverity::Critical:
		UE_LOG(LogMMORPGError, Fatal, TEXT("[%s] %s - Code: %d, Message: %s"),
			*CategoryStr, *SeverityStr, Error.Code, *Error.Message);
		break;
	}
	
	// Also log context if available
	if (!Error.Context.IsEmpty())
	{
		UE_LOG(LogMMORPGError, Log, TEXT("  Context: %s"), *Error.Context);
	}
}