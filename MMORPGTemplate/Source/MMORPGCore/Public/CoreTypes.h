#pragma once

#include "CoreMinimal.h"
#include "CoreTypes.generated.h"

/**
 * Error severity levels
 */
UENUM(BlueprintType)
enum class EMMORPGErrorSeverity : uint8
{
	Info         UMETA(DisplayName = "Info"),
	Warning      UMETA(DisplayName = "Warning"),
	Error        UMETA(DisplayName = "Error"),
	Critical     UMETA(DisplayName = "Critical")
};

/**
 * Error categories matching backend error codes
 */
UENUM(BlueprintType)
enum class EMMORPGErrorCategory : uint8
{
	Network      UMETA(DisplayName = "Network (1000-1999)"),
	Auth         UMETA(DisplayName = "Authentication (2000-2999)"),
	Protocol     UMETA(DisplayName = "Protocol (3000-3999)"),
	GameLogic    UMETA(DisplayName = "Game Logic (4000-4999)"),
	System       UMETA(DisplayName = "System (5000-5999)")
};

/**
 * Error structure for unified error handling
 */
USTRUCT(BlueprintType)
struct MMORPGCORE_API FMMORPGError
{
	GENERATED_BODY()

	UPROPERTY(BlueprintReadWrite, Category = "Error")
	int32 Code = 0;

	UPROPERTY(BlueprintReadWrite, Category = "Error")
	FString Message;

	UPROPERTY(BlueprintReadWrite, Category = "Error")
	EMMORPGErrorSeverity Severity = EMMORPGErrorSeverity::Error;

	UPROPERTY(BlueprintReadWrite, Category = "Error")
	EMMORPGErrorCategory Category = EMMORPGErrorCategory::System;

	UPROPERTY(BlueprintReadWrite, Category = "Error")
	FDateTime Timestamp;

	UPROPERTY(BlueprintReadWrite, Category = "Error")
	FString Context;

	FMMORPGError()
	{
		Timestamp = FDateTime::Now();
	}

	FMMORPGError(int32 InCode, const FString& InMessage, EMMORPGErrorCategory InCategory)
		: Code(InCode)
		, Message(InMessage)
		, Category(InCategory)
	{
		Timestamp = FDateTime::Now();
		
		// Determine severity based on code
		if (Code >= 5000)
		{
			Severity = EMMORPGErrorSeverity::Critical;
		}
		else if (Code >= 4000)
		{
			Severity = EMMORPGErrorSeverity::Error;
		}
		else
		{
			Severity = EMMORPGErrorSeverity::Warning;
		}
	}
};

/**
 * Delegate for error events
 */
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnMMORPGError, const FMMORPGError&, Error);