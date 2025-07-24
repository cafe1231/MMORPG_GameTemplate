#pragma once

#include "CoreMinimal.h"
#include "CoreTypes.h"
#include "FAuthTypes.generated.h"

/**
 * User information structure
 */
USTRUCT(BlueprintType)
struct MMORPGCORE_API FUserInfo
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString Id;

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString Email;

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString Username;

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FDateTime CreatedAt;

    FUserInfo()
    {
        Id = "";
        Email = "";
        Username = "";
        CreatedAt = FDateTime::MinValue();
    }
};

/**
 * Login request structure
 */
USTRUCT(BlueprintType)
struct MMORPGCORE_API FLoginRequest
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString Email;

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString Password;

    FLoginRequest()
    {
        Email = "";
        Password = "";
    }
};

/**
 * Login response structure
 */
USTRUCT(BlueprintType)
struct MMORPGCORE_API FLoginResponse
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString AccessToken;

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString RefreshToken;

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FUserInfo User;

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    int32 ExpiresIn;

    FLoginResponse()
    {
        AccessToken = "";
        RefreshToken = "";
        ExpiresIn = 0;
    }
};

/**
 * Register request structure
 */
USTRUCT(BlueprintType)
struct MMORPGCORE_API FRegisterRequest
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString Email;

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString Username;

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString Password;

    FRegisterRequest()
    {
        Email = "";
        Username = "";
        Password = "";
    }
};

/**
 * Register response structure
 */
USTRUCT(BlueprintType)
struct MMORPGCORE_API FRegisterResponse
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FUserInfo User;

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString Message;

    FRegisterResponse()
    {
        Message = "";
    }
};

/**
 * Token refresh request structure
 */
USTRUCT(BlueprintType)
struct MMORPGCORE_API FRefreshTokenRequest
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString RefreshToken;

    FRefreshTokenRequest()
    {
        RefreshToken = "";
    }
};

/**
 * Token refresh response structure
 */
USTRUCT(BlueprintType)
struct MMORPGCORE_API FRefreshTokenResponse
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString AccessToken;

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString RefreshToken;

    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    int32 ExpiresIn;

    FRefreshTokenResponse()
    {
        AccessToken = "";
        RefreshToken = "";
        ExpiresIn = 0;
    }
};

/**
 * Delegate signatures for authentication callbacks
 */
DECLARE_DELEGATE_OneParam(FOnLoginComplete, const FLoginResponse&);
DECLARE_DELEGATE_OneParam(FOnLoginFailed, const FMMORPGError&);
DECLARE_DELEGATE_OneParam(FOnRegisterComplete, const FRegisterResponse&);
DECLARE_DELEGATE_OneParam(FOnRegisterFailed, const FMMORPGError&);
DECLARE_DELEGATE_OneParam(FOnRefreshTokenComplete, const FRefreshTokenResponse&);
DECLARE_DELEGATE_OneParam(FOnRefreshTokenFailed, const FMMORPGError&);
DECLARE_DELEGATE(FOnLogoutComplete);

/**
 * Blueprint-friendly delegate signatures
 */
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnLoginCompleteBP, const FLoginResponse&, Response);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnLoginFailedBP, const FMMORPGError&, Error);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnRegisterCompleteBP, const FRegisterResponse&, Response);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnRegisterFailedBP, const FMMORPGError&, Error);
DECLARE_DYNAMIC_MULTICAST_DELEGATE(FOnLogoutCompleteBP);