#include "Auth/AMMORPGAuthGameMode.h"
#include "Auth/UMMORPGAuthWidget.h"
#include "Auth/AMMORPGAuthPlayerController.h"
#include "Blueprint/UserWidget.h"
#include "Engine/World.h"
#include "Kismet/GameplayStatics.h"

AMMORPGAuthGameMode::AMMORPGAuthGameMode()
{
    // Set default player controller class (will be created in Blueprint)
    PlayerControllerClass = AMMORPGAuthPlayerController::StaticClass();
}

void AMMORPGAuthGameMode::BeginPlay()
{
    Super::BeginPlay();

    // Create and display auth widget
    if (AuthWidgetClass)
    {
        if (APlayerController* PC = UGameplayStatics::GetPlayerController(this, 0))
        {
            AuthWidget = CreateWidget<UMMORPGAuthWidget>(PC, AuthWidgetClass);
            if (AuthWidget)
            {
                AuthWidget->AddToViewport();
                
                // Set input mode to UI only
                FInputModeUIOnly InputMode;
                InputMode.SetWidgetToFocus(AuthWidget->TakeWidget());
                InputMode.SetLockMouseToViewportBehavior(EMouseLockMode::DoNotLock);
                
                PC->SetInputMode(InputMode);
                PC->bShowMouseCursor = true;
            }
        }
    }
}

void AMMORPGAuthGameMode::HandleAuthenticationSuccess()
{
    // Remove auth widget
    if (AuthWidget)
    {
        AuthWidget->RemoveFromParent();
        AuthWidget = nullptr;
    }

    // Call blueprint event
    OnAuthenticationSuccess();
}