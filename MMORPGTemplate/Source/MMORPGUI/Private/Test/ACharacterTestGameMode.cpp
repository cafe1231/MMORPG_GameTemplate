#include "Test/ACharacterTestGameMode.h"
#include "Subsystems/UMMORPGCharacterSubsystem.h"
#include "Blueprint/UserWidget.h"
#include "Kismet/GameplayStatics.h"

ACharacterTestGameMode::ACharacterTestGameMode()
{
    // Default pawn
    DefaultPawnClass = nullptr;
}

void ACharacterTestGameMode::BeginPlay()
{
    Super::BeginPlay();

    // Get character subsystem and enable mock mode
    if (UGameInstance* GameInstance = GetGameInstance())
    {
        if (UMMORPGCharacterSubsystem* CharacterSubsystem = GameInstance->GetSubsystem<UMMORPGCharacterSubsystem>())
        {
            // Enable mock mode for testing without backend
            CharacterSubsystem->SetMockMode(true);
            
            UE_LOG(LogTemp, Warning, TEXT("Character Test Mode: Mock mode enabled"));
        }
    }

    // Create and display the character creation widget
    if (CharacterCreateWidgetClass)
    {
        if (APlayerController* PC = GetWorld()->GetFirstPlayerController())
        {
            CharacterCreateWidget = CreateWidget<UUserWidget>(PC, CharacterCreateWidgetClass);
            if (CharacterCreateWidget)
            {
                CharacterCreateWidget->AddToViewport();
                
                // Set input mode to UI
                FInputModeUIOnly InputMode;
                InputMode.SetWidgetToFocus(CharacterCreateWidget->TakeWidget());
                InputMode.SetLockMouseToViewportBehavior(EMouseLockMode::DoNotLock);
                
                PC->SetInputMode(InputMode);
                PC->bShowMouseCursor = true;
            }
        }
    }
    else
    {
        UE_LOG(LogTemp, Error, TEXT("CharacterCreateWidgetClass not set in CharacterTestGameMode!"));
    }
}