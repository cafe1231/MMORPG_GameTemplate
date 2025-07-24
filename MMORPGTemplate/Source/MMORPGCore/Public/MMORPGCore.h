#pragma once

#include "CoreMinimal.h"
#include "Modules/ModuleManager.h"

// Log category for MMORPGCore module
DECLARE_LOG_CATEGORY_EXTERN(LogMMORPGCore, Log, All);

class FMMORPGCoreModule : public IModuleInterface
{
public:
	/** IModuleInterface implementation */
	virtual void StartupModule() override;
	virtual void ShutdownModule() override;
};