#pragma once

#include "CoreMinimal.h"
#include "GameFramework/GameModeBase.h"
#include "ACharacterTestGameMode.generated.h"

/**
 * Test game mode for character system
 */
UCLASS()
class MMORPGUI_API ACharacterTestGameMode : public AGameModeBase
{
    GENERATED_BODY()

public:
    ACharacterTestGameMode();

    virtual void BeginPlay() override;

protected:
    // Widget class to spawn
    UPROPERTY(EditDefaultsOnly, Category = "UI")
    TSubclassOf<class UUserWidget> CharacterCreateWidgetClass;

    // Widget instance
    UPROPERTY()
    class UUserWidget* CharacterCreateWidget;
};