#pragma once

#include "CoreMinimal.h"
#include "GameFramework/PlayerController.h"
#include "ACharacterTestPlayerController.generated.h"

/**
 * Test player controller for character system
 */
UCLASS()
class MMORPGUI_API ACharacterTestPlayerController : public APlayerController
{
    GENERATED_BODY()

public:
    ACharacterTestPlayerController();
};