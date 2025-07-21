// Copyright (c) 2024 MMORPG Template Project

#include "MMORPGCore.h"
#include "Core.h"
#include "Modules/ModuleManager.h"
#include "Engine/Engine.h"
#include "Misc/ConfigCacheIni.h"
#include "Network/MMORPGNetworkManager.h"
#include "Console/MMORPGDeveloperConsole.h"
#include "Error/MMORPGErrorHandler.h"
#include "GameFramework/GameplayStatics.h"

// Define log categories
DEFINE_LOG_CATEGORY(LogMMORPG);
DEFINE_LOG_CATEGORY(LogMMORPGNetwork);
DEFINE_LOG_CATEGORY(LogMMORPGAuth);
DEFINE_LOG_CATEGORY(LogMMORPGGame);

#define LOCTEXT_NAMESPACE "FMMORPGCoreModule"

void FMMORPGCoreModule::StartupModule()
{
    MMORPG_LOG(Log, TEXT("MMORPG Template Plugin Starting - Version %s"), *PluginVersion);
    
    // Load configuration
    FString ConfigFile = FPaths::ProjectConfigDir() / TEXT("DefaultMMORPG.ini");
    if (FPaths::FileExists(ConfigFile))
    {
        MMORPG_LOG(Log, TEXT("Loading MMORPG configuration from: %s"), *ConfigFile);
        // Load custom config settings here
    }
    
    // Log system information
    MMORPG_LOG(Log, TEXT("Platform: %s"), *UGameplayStatics::GetPlatformName());
    MMORPG_LOG(Log, TEXT("Engine Version: %s"), *FEngineVersion::Current().ToString());
    MMORPG_LOG(Log, TEXT("Protocol Version: %d"), ProtocolVersion);
    
    // Initialize subsystems
    InitializeManagers();
    
    // Register console commands for development
#if !UE_BUILD_SHIPPING
    if (!IsRunningDedicatedServer())
    {
        // Register console commands
        IConsoleManager::Get().RegisterConsoleCommand(
            TEXT("mmorpg.status"),
            TEXT("Show MMORPG plugin status"),
            FConsoleCommandDelegate::CreateLambda([]()
            {
                if (FMMORPGCoreModule::IsAvailable())
                {
                    FMMORPGCoreModule& Module = FMMORPGCoreModule::Get();
                    UE_LOG(LogMMORPG, Display, TEXT("MMORPG Plugin Status:"));
                    UE_LOG(LogMMORPG, Display, TEXT("  Version: %s"), *Module.GetPluginVersion());
                    UE_LOG(LogMMORPG, Display, TEXT("  Protocol: %d"), Module.GetProtocolVersion());
                    UE_LOG(LogMMORPG, Display, TEXT("  Network Manager: %s"), 
                        Module.GetNetworkManager().IsValid() ? TEXT("Active") : TEXT("Inactive"));
                    UE_LOG(LogMMORPG, Display, TEXT("  Auth Manager: %s"), 
                        Module.GetAuthManager().IsValid() ? TEXT("Active") : TEXT("Inactive"));
                    UE_LOG(LogMMORPG, Display, TEXT("  Data Manager: %s"), 
                        Module.GetDataManager().IsValid() ? TEXT("Active") : TEXT("Inactive"));
                }
            }),
            ECVF_Default
        );
        
        IConsoleManager::Get().RegisterConsoleCommand(
            TEXT("mmorpg.connect"),
            TEXT("Connect to MMORPG server (usage: mmorpg.connect <host> <port>)"),
            FConsoleCommandWithArgsDelegate::CreateLambda([](const TArray<FString>& Args)
            {
                if (Args.Num() < 2)
                {
                    UE_LOG(LogMMORPG, Warning, TEXT("Usage: mmorpg.connect <host> <port>"));
                    return;
                }
                
                FString Host = Args[0];
                int32 Port = FCString::Atoi(*Args[1]);
                
                if (FMMORPGCoreModule::IsAvailable())
                {
                    auto NetworkManager = FMMORPGCoreModule::Get().GetNetworkManager();
                    if (NetworkManager.IsValid())
                    {
                        NetworkManager->Connect(Host, Port);
                        UE_LOG(LogMMORPG, Log, TEXT("Connecting to %s:%d..."), *Host, Port);
                    }
                }
            }),
            ECVF_Default
        );
        
        IConsoleManager::Get().RegisterConsoleCommand(
            TEXT("mmorpg.console"),
            TEXT("Toggle MMORPG developer console"),
            FConsoleCommandDelegate::CreateLambda([]() 
            {
                if (FMMORPGCoreModule::IsAvailable())
                {
                    UMMORPGDeveloperConsole* Console = FMMORPGCoreModule::Get().GetDeveloperConsole();
                    if (Console)
                    {
                        Console->ToggleConsole();
                    }
                }
            }),
            ECVF_Default
        );
    }
#endif
    
    MMORPG_LOG(Log, TEXT("MMORPG Template Plugin Started Successfully"));
}

void FMMORPGCoreModule::ShutdownModule()
{
    MMORPG_LOG(Log, TEXT("MMORPG Template Plugin Shutting Down"));
    
    // Shutdown managers in reverse order
    ShutdownManagers();
    
    // Unregister console commands
#if !UE_BUILD_SHIPPING
    IConsoleManager::Get().UnregisterConsoleObject(TEXT("mmorpg.status"));
    IConsoleManager::Get().UnregisterConsoleObject(TEXT("mmorpg.connect"));
    IConsoleManager::Get().UnregisterConsoleObject(TEXT("mmorpg.console"));
#endif
    
    MMORPG_LOG(Log, TEXT("MMORPG Template Plugin Shutdown Complete"));
}

void FMMORPGCoreModule::InitializeManagers()
{
    if (bManagersInitialized)
    {
        MMORPG_LOG(Warning, TEXT("Managers already initialized"));
        return;
    }
    
    MMORPG_LOG(Log, TEXT("Initializing MMORPG Managers"));
    
    // Create manager instances
    NetworkManager = MakeShareable(new FMMORPGNetworkManager());
    // AuthManager = MakeShareable(new FMMORPGAuthManager());
    // DataManager = MakeShareable(new FMMORPGDataManager());
    
    // Create error handler
    ErrorHandler = NewObject<UMMORPGErrorHandler>(GetTransientPackage(), UMMORPGErrorHandler::StaticClass());
    if (ErrorHandler)
    {
        ErrorHandler->AddToRoot(); // Prevent garbage collection
        ErrorHandler->Initialize();
    }
    
    // Create developer console
    DeveloperConsole = NewObject<UMMORPGDeveloperConsole>(GetTransientPackage(), UMMORPGDeveloperConsole::StaticClass());
    if (DeveloperConsole)
    {
        DeveloperConsole->AddToRoot(); // Prevent garbage collection
        DeveloperConsole->Initialize();
    }
    
    // Initialize managers
    if (NetworkManager.IsValid())
    {
        NetworkManager->Initialize();
    }
    
    // if (AuthManager.IsValid())
    // {
    //     AuthManager->Initialize();
    // }
    
    // if (DataManager.IsValid())
    // {
    //     DataManager->Initialize();
    // }
    
    bManagersInitialized = true;
    MMORPG_LOG(Log, TEXT("MMORPG Managers Initialized"));
}

void FMMORPGCoreModule::ShutdownManagers()
{
    if (!bManagersInitialized)
    {
        return;
    }
    
    MMORPG_LOG(Log, TEXT("Shutting down MMORPG Managers"));
    
    // Shutdown developer console
    if (DeveloperConsole)
    {
        DeveloperConsole->Shutdown();
        DeveloperConsole->RemoveFromRoot();
        DeveloperConsole = nullptr;
    }
    
    // Shutdown error handler
    if (ErrorHandler)
    {
        ErrorHandler->Shutdown();
        ErrorHandler->RemoveFromRoot();
        ErrorHandler = nullptr;
    }
    
    // Shutdown in reverse order
    // if (DataManager.IsValid())
    // {
    //     DataManager->Shutdown();
    //     DataManager.Reset();
    // }
    
    // if (AuthManager.IsValid())
    // {
    //     AuthManager->Shutdown();
    //     AuthManager.Reset();
    // }
    
    if (NetworkManager.IsValid())
    {
        NetworkManager->Shutdown();
        NetworkManager.Reset();
    }
    
    bManagersInitialized = false;
    MMORPG_LOG(Log, TEXT("MMORPG Managers Shutdown Complete"));
}

#undef LOCTEXT_NAMESPACE

IMPLEMENT_MODULE(FMMORPGCoreModule, MMORPGCore)