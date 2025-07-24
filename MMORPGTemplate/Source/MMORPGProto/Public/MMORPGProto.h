#pragma once

#include "CoreMinimal.h"
#include "Modules/ModuleManager.h"

// Log category for MMORPGProto module
DECLARE_LOG_CATEGORY_EXTERN(LogMMORPGProto, Log, All);

class FMMORPGProtoModule : public IModuleInterface
{
public:
	/** IModuleInterface implementation */
	virtual void StartupModule() override;
	virtual void ShutdownModule() override;
};