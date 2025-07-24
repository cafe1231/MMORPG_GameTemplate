#pragma once

#include "CoreMinimal.h"
#include "Modules/ModuleManager.h"

// Log category for MMORPGNetwork module
DECLARE_LOG_CATEGORY_EXTERN(LogMMORPGNetwork, Log, All);

class FMMORPGNetworkModule : public IModuleInterface
{
public:
	/** IModuleInterface implementation */
	virtual void StartupModule() override;
	virtual void ShutdownModule() override;
};