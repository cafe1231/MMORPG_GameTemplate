#include "Auth/AMMORPGAuthPlayerController.h"

AMMORPGAuthPlayerController::AMMORPGAuthPlayerController()
{
    // Enable mouse cursor by default for UI interaction
    bShowMouseCursor = true;
    bEnableClickEvents = true;
    bEnableMouseOverEvents = true;
}

void AMMORPGAuthPlayerController::BeginPlay()
{
    Super::BeginPlay();
    
    // Set input mode to UI only at start
    FInputModeUIOnly InputMode;
    InputMode.SetLockMouseToViewportBehavior(EMouseLockMode::DoNotLock);
    SetInputMode(InputMode);
}