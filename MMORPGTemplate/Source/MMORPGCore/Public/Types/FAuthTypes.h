#pragma once

#include "CoreMinimal.h"
#include "FAuthTypes.generated.h"

USTRUCT(BlueprintType)
struct FLoginRequest
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

USTRUCT(BlueprintType)
struct FRegisterRequest
{
    GENERATED_BODY()
    
    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString Email;
    
    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString Password;
    
    UPROPERTY(BlueprintReadWrite, Category = "MMORPG|Auth")
    FString Username;

    FRegisterRequest()
    {
        Email = "";
        Password = "";
        Username = "";
    }
};

USTRUCT(BlueprintType)
struct FAuthTokens
{
    GENERATED_BODY()
    
    UPROPERTY(BlueprintReadOnly, Category = "MMORPG|Auth")
    FString AccessToken;
    
    UPROPERTY(BlueprintReadOnly, Category = "MMORPG|Auth")
    FString RefreshToken;
    
    UPROPERTY(BlueprintReadOnly, Category = "MMORPG|Auth")
    FDateTime ExpiresAt;

    FAuthTokens()
    {
        AccessToken = "";
        RefreshToken = "";
        ExpiresAt = FDateTime::MinValue();
    }
};

USTRUCT(BlueprintType)
struct FAuthResponse
{
    GENERATED_BODY()
    
    UPROPERTY(BlueprintReadOnly, Category = "MMORPG|Auth")
    bool bSuccess;
    
    UPROPERTY(BlueprintReadOnly, Category = "MMORPG|Auth")
    FString Message;
    
    UPROPERTY(BlueprintReadOnly, Category = "MMORPG|Auth")
    FAuthTokens Tokens;

    FAuthResponse()
    {
        bSuccess = false;
        Message = "";
    }
};

USTRUCT(BlueprintType)
struct FUserInfo
{
    GENERATED_BODY()
    
    UPROPERTY(BlueprintReadOnly, Category = "MMORPG|Auth")
    FString UserId;
    
    UPROPERTY(BlueprintReadOnly, Category = "MMORPG|Auth")
    FString Email;
    
    UPROPERTY(BlueprintReadOnly, Category = "MMORPG|Auth")
    FString Username;

    FUserInfo()
    {
        UserId = "";
        Email = "";
        Username = "";
    }
};