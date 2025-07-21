// Copyright (c) 2024 MMORPG Template Project

#pragma once

#include "CoreMinimal.h"
#include "GameFramework/Actor.h"
#include "MMORPGConnectionTest.generated.h"

/**
 * Simple test actor for verifying client-server connection
 * Spawns in the world and performs various network tests
 */
UCLASS(Blueprintable, Category = "MMORPG|Testing")
class MMORPGCORE_API AMMORPGConnectionTest : public AActor
{
    GENERATED_BODY()

public:
    AMMORPGConnectionTest();

    /** Run all connection tests */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Testing")
    void RunConnectionTests();

    /** Test basic connection */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Testing")
    void TestBasicConnection();

    /** Test echo functionality */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Testing")
    void TestEcho(const FString& Message);

    /** Test health endpoint */
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Testing")
    void TestHealthCheck();

protected:
    virtual void BeginPlay() override;

    UPROPERTY(BlueprintReadOnly, Category = "MMORPG|Testing")
    bool bIsRunningTests;

    UPROPERTY(BlueprintReadOnly, Category = "MMORPG|Testing")
    TArray<FString> TestResults;

    UPROPERTY(EditDefaultsOnly, BlueprintReadOnly, Category = "MMORPG|Testing")
    bool bAutoRunOnBeginPlay;

    UPROPERTY(EditDefaultsOnly, BlueprintReadOnly, Category = "MMORPG|Testing")
    FString ServerHost;

    UPROPERTY(EditDefaultsOnly, BlueprintReadOnly, Category = "MMORPG|Testing")
    int32 ServerPort;

private:
    void AddTestResult(const FString& TestName, bool bSuccess, const FString& Details);
    void OnTestComplete(const FString& TestName, bool bSuccess, const FString& Response);
};