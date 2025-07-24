#include "MMORPGProto.h"

#define LOCTEXT_NAMESPACE "FMMORPGProtoModule"

// Define log category
DEFINE_LOG_CATEGORY(LogMMORPGProto);

void FMMORPGProtoModule::StartupModule()
{
	// This code will execute after your module is loaded into memory
	UE_LOG(LogMMORPGProto, Log, TEXT("MMORPGProto Module Started - Protocol buffer types and converters initialized"));
}

void FMMORPGProtoModule::ShutdownModule()
{
	// This function may be called during shutdown to clean up your module
	UE_LOG(LogMMORPGProto, Log, TEXT("MMORPGProto Module Shutdown"));
}

#undef LOCTEXT_NAMESPACE
	
IMPLEMENT_MODULE(FMMORPGProtoModule, MMORPGProto)