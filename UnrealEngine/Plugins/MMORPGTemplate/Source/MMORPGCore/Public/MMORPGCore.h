// Copyright (c) 2024 MMORPG Template Project

#pragma once

#include "CoreMinimal.h"
#include "Modules/ModuleManager.h"

DECLARE_LOG_CATEGORY_EXTERN(LogMMORPG, Log, All);
DECLARE_LOG_CATEGORY_EXTERN(LogMMORPGNetwork, Log, All);
DECLARE_LOG_CATEGORY_EXTERN(LogMMORPGAuth, Log, All);
DECLARE_LOG_CATEGORY_EXTERN(LogMMORPGGame, Log, All);

// Forward declarations
class FMMORPGNetworkManager;
class FMMORPGAuthManager;
class FMMORPGDataManager;
class UMMORPGSettings;
class UMMORPGDeveloperConsole;
class UMMORPGErrorHandler;

/**
 * Main module class for MMORPG Template plugin
 */
class MMORPGCORE_API FMMORPGCoreModule : public IModuleInterface
{
public:
    /** IModuleInterface implementation */
    virtual void StartupModule() override;
    virtual void ShutdownModule() override;
    
    /** Get the singleton instance of this module */
    static inline FMMORPGCoreModule& Get()
    {
        return FModuleManager::LoadModuleChecked<FMMORPGCoreModule>("MMORPGCore");
    }
    
    /** Check if the module is loaded */
    static inline bool IsAvailable()
    {
        return FModuleManager::Get().IsModuleLoaded("MMORPGCore");
    }
    
    /** Get the network manager instance */
    TSharedPtr<FMMORPGNetworkManager> GetNetworkManager() const { return NetworkManager; }
    
    /** Get the authentication manager instance */
    TSharedPtr<FMMORPGAuthManager> GetAuthManager() const { return AuthManager; }
    
    /** Get the data manager instance */
    TSharedPtr<FMMORPGDataManager> GetDataManager() const { return DataManager; }
    
    /** Get the developer console instance */
    UMMORPGDeveloperConsole* GetDeveloperConsole() const { return DeveloperConsole; }
    
    /** Get the error handler instance */
    UMMORPGErrorHandler* GetErrorHandler() const { return ErrorHandler; }
    
    /** Get plugin version */
    FString GetPluginVersion() const { return PluginVersion; }
    
    /** Get protocol version */
    uint32 GetProtocolVersion() const { return ProtocolVersion; }
    
    /** Initialize managers - called after module startup */
    void InitializeManagers();
    
    /** Shutdown managers - called before module shutdown */
    void ShutdownManagers();
    
private:
    /** Network manager instance */
    TSharedPtr<FMMORPGNetworkManager> NetworkManager;
    
    /** Authentication manager instance */
    TSharedPtr<FMMORPGAuthManager> AuthManager;
    
    /** Data manager instance */
    TSharedPtr<FMMORPGDataManager> DataManager;
    
    /** Developer console instance */
    UPROPERTY()
    UMMORPGDeveloperConsole* DeveloperConsole;
    
    /** Error handler instance */
    UPROPERTY()
    UMMORPGErrorHandler* ErrorHandler;
    
    /** Plugin version string */
    FString PluginVersion = TEXT("0.1.0");
    
    /** Protocol version for client-server communication */
    uint32 ProtocolVersion = 1;
    
    /** Flag to track if managers are initialized */
    bool bManagersInitialized = false;
};

// Helper macros for logging
#define MMORPG_LOG(Verbosity, Format, ...) \
    UE_LOG(LogMMORPG, Verbosity, Format, ##__VA_ARGS__)

#define MMORPG_LOG_NET(Verbosity, Format, ...) \
    UE_LOG(LogMMORPGNetwork, Verbosity, Format, ##__VA_ARGS__)

#define MMORPG_LOG_AUTH(Verbosity, Format, ...) \
    UE_LOG(LogMMORPGAuth, Verbosity, Format, ##__VA_ARGS__)

#define MMORPG_LOG_GAME(Verbosity, Format, ...) \
    UE_LOG(LogMMORPGGame, Verbosity, Format, ##__VA_ARGS__)

// Conditional logging for verbose messages
#if !UE_BUILD_SHIPPING
    #define MMORPG_LOG_VERBOSE(Format, ...) \
        UE_LOG(LogMMORPG, Verbose, Format, ##__VA_ARGS__)
#else
    #define MMORPG_LOG_VERBOSE(Format, ...)
#endif