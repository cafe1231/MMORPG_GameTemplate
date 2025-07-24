#include "MMORPGNetwork.h"

#define LOCTEXT_NAMESPACE "FMMORPGNetworkModule"

// Define log category
DEFINE_LOG_CATEGORY(LogMMORPGNetwork);

void FMMORPGNetworkModule::StartupModule()
{
	// This code will execute after your module is loaded into memory
	UE_LOG(LogMMORPGNetwork, Log, TEXT("MMORPGNetwork Module Started - HTTP and WebSocket clients initialized"));
}

void FMMORPGNetworkModule::ShutdownModule()
{
	// This function may be called during shutdown to clean up your module
	UE_LOG(LogMMORPGNetwork, Log, TEXT("MMORPGNetwork Module Shutdown"));
}

#undef LOCTEXT_NAMESPACE
	
IMPLEMENT_MODULE(FMMORPGNetworkModule, MMORPGNetwork)