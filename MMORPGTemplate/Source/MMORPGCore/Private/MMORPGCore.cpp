#include "MMORPGCore.h"

#define LOCTEXT_NAMESPACE "FMMORPGCoreModule"

// Define log category
DEFINE_LOG_CATEGORY(LogMMORPGCore);

void FMMORPGCoreModule::StartupModule()
{
	// This code will execute after your module is loaded into memory
	UE_LOG(LogMMORPGCore, Log, TEXT("MMORPGCore Module Started - Error handling system initialized"));
}

void FMMORPGCoreModule::ShutdownModule()
{
	// This function may be called during shutdown to clean up your module
	UE_LOG(LogMMORPGCore, Log, TEXT("MMORPGCore Module Shutdown"));
}

#undef LOCTEXT_NAMESPACE
	
IMPLEMENT_MODULE(FMMORPGCoreModule, MMORPGCore)