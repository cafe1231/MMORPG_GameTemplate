#pragma once

#include "CoreMinimal.h"
#include "GameFramework/GameModeBase.h"
#include "AMMORPGAuthGameMode.generated.h"

class UMMORPGAuthWidget;

UCLASS()
class MMORPGUI_API AMMORPGAuthGameMode : public AGameModeBase
{
    GENERATED_BODY()

public:
    AMMORPGAuthGameMode();

    virtual void BeginPlay() override;

protected:
    // Widget class to spawn
    UPROPERTY(EditDefaultsOnly, BlueprintReadOnly, Category = "MMORPG|Auth")
    TSubclassOf<UMMORPGAuthWidget> AuthWidgetClass;

    // Blueprint event called when authentication is successful
    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|Auth")
    void OnAuthenticationSuccess();

private:
    UPROPERTY()
    UMMORPGAuthWidget* AuthWidget;

    UFUNCTION()
    void HandleAuthenticationSuccess();
};