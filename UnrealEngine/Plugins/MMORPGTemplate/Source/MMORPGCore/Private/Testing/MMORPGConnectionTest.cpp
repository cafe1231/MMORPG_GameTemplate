// Copyright (c) 2024 MMORPG Template Project

#include "Testing/MMORPGConnectionTest.h"
#include "MMORPGCore.h"
#include "Blueprints/MMORPGNetworkBPLibrary.h"
#include "Engine/Engine.h"
#include "TimerManager.h"

AMMORPGConnectionTest::AMMORPGConnectionTest()
{
    PrimaryActorTick.bCanEverTick = false;
    
    bAutoRunOnBeginPlay = true;
    ServerHost = TEXT("localhost");
    ServerPort = 8090;
    bIsRunningTests = false;
}

void AMMORPGConnectionTest::BeginPlay()
{
    Super::BeginPlay();
    
    if (bAutoRunOnBeginPlay)
    {
        GetWorldTimerManager().SetTimer(
            FTimerHandle(),
            this,
            &AMMORPGConnectionTest::RunConnectionTests,
            2.0f,
            false
        );
    }
}

void AMMORPGConnectionTest::RunConnectionTests()
{
    if (bIsRunningTests)
    {
        MMORPG_LOG(Warning, TEXT("Tests already running"));
        return;
    }
    
    bIsRunningTests = true;
    TestResults.Empty();
    
    MMORPG_LOG(Log, TEXT("Starting connection tests to %s:%d"), *ServerHost, ServerPort);
    if (GEngine)
    {
        GEngine->AddOnScreenDebugMessage(-1, 5.0f, FColor::Yellow, 
            FString::Printf(TEXT("Starting MMORPG Connection Tests to %s:%d"), *ServerHost, ServerPort));
    }
    
    UMMORPGNetworkBPLibrary::ConnectToServer(ServerHost, ServerPort);
    
    FTimerHandle TimerHandle;
    GetWorldTimerManager().SetTimer(TimerHandle, [this]()
    {
        TestBasicConnection();
    }, 0.5f, false);
}

void AMMORPGConnectionTest::TestBasicConnection()
{
    MMORPG_LOG(Log, TEXT("Testing basic connection..."));
    
    FOnHttpRequestComplete Delegate;
    Delegate.BindLambda([this](bool bSuccess, const FString& Response)
    {
        OnTestComplete(TEXT("Basic Connection Test"), bSuccess, Response);
        
        if (bSuccess)
        {
            FTimerHandle TimerHandle;
            GetWorldTimerManager().SetTimer(TimerHandle, [this]()
            {
                TestHealthCheck();
            }, 0.5f, false);
        }
        else
        {
            bIsRunningTests = false;
        }
    });
    
    UMMORPGNetworkBPLibrary::TestAPI(Delegate);
}

void AMMORPGConnectionTest::TestEcho(const FString& Message)
{
    MMORPG_LOG(Log, TEXT("Testing echo with message: %s"), *Message);
    
    FOnHttpRequestComplete Delegate;
    Delegate.BindLambda([this, Message](bool bSuccess, const FString& Response)
    {
        OnTestComplete(FString::Printf(TEXT("Echo Test: %s"), *Message), bSuccess, Response);
    });
    
    UMMORPGNetworkBPLibrary::EchoTest(Message, Delegate);
}

void AMMORPGConnectionTest::TestHealthCheck()
{
    MMORPG_LOG(Log, TEXT("Testing health check..."));
    
    FOnHttpRequestComplete Delegate;
    Delegate.BindLambda([this](bool bSuccess, const FString& Response)
    {
        OnTestComplete(TEXT("Health Check"), bSuccess, Response);
        
        FTimerHandle TimerHandle;
        GetWorldTimerManager().SetTimer(TimerHandle, [this]()
        {
            TestEcho(TEXT("Hello from Unreal Engine 5.6!"));
        }, 0.5f, false);
        
        FTimerHandle CompleteHandle;
        GetWorldTimerManager().SetTimer(CompleteHandle, [this]()
        {
            bIsRunningTests = false;
            
            MMORPG_LOG(Log, TEXT("Connection tests completed. Results:"));
            for (const FString& Result : TestResults)
            {
                MMORPG_LOG(Log, TEXT("  %s"), *Result);
            }
            
            if (GEngine)
            {
                GEngine->AddOnScreenDebugMessage(-1, 10.0f, FColor::Green, 
                    FString::Printf(TEXT("MMORPG Connection Tests Completed - %d tests run"), TestResults.Num()));
            }
        }, 1.5f, false);
    });
    
    UMMORPGNetworkBPLibrary::GetHealthStatus(Delegate);
}

void AMMORPGConnectionTest::AddTestResult(const FString& TestName, bool bSuccess, const FString& Details)
{
    FString Result = FString::Printf(TEXT("[%s] %s: %s"), 
        bSuccess ? TEXT("PASS") : TEXT("FAIL"), 
        *TestName, 
        *Details);
    
    TestResults.Add(Result);
    
    if (GEngine)
    {
        FColor Color = bSuccess ? FColor::Green : FColor::Red;
        GEngine->AddOnScreenDebugMessage(-1, 5.0f, Color, Result);
    }
}

void AMMORPGConnectionTest::OnTestComplete(const FString& TestName, bool bSuccess, const FString& Response)
{
    FString Details = bSuccess ? TEXT("Success") : Response;
    if (bSuccess && Response.Len() > 100)
    {
        Details = Response.Left(100) + TEXT("...");
    }
    else if (bSuccess)
    {
        Details = Response;
    }
    
    AddTestResult(TestName, bSuccess, Details);
}