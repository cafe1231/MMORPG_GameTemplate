#pragma once

#include "CoreMinimal.h"
#include "GameFramework/PlayerController.h"
#include "AMMORPGAuthPlayerController.generated.h"

UCLASS()
class MMORPGUI_API AMMORPGAuthPlayerController : public APlayerController
{
    GENERATED_BODY()

public:
    AMMORPGAuthPlayerController();

protected:
    virtual void BeginPlay() override;
};