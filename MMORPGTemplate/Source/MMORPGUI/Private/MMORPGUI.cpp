#include "MMORPGUI.h"

#define LOCTEXT_NAMESPACE "FMMORPGUIModule"

// Define log category
DEFINE_LOG_CATEGORY(LogMMORPGUI);

void FMMORPGUIModule::StartupModule()
{
	// This code will execute after your module is loaded into memory
	UE_LOG(LogMMORPGUI, Log, TEXT("MMORPGUI Module Started - Console subsystem initialized"));
}

void FMMORPGUIModule::ShutdownModule()
{
	// This function may be called during shutdown to clean up your module
	UE_LOG(LogMMORPGUI, Log, TEXT("MMORPGUI Module Shutdown"));
}

#undef LOCTEXT_NAMESPACE
	
IMPLEMENT_MODULE(FMMORPGUIModule, MMORPGUI)