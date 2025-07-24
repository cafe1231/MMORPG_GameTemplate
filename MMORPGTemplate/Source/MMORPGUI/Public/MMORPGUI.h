#pragma once

#include "CoreMinimal.h"
#include "Modules/ModuleManager.h"

// Log category for MMORPGUI module
DECLARE_LOG_CATEGORY_EXTERN(LogMMORPGUI, Log, All);

class FMMORPGUIModule : public IModuleInterface
{
public:
	/** IModuleInterface implementation */
	virtual void StartupModule() override;
	virtual void ShutdownModule() override;
};