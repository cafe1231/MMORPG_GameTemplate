#pragma once

#include "CoreMinimal.h"
#include "Blueprint/UserWidget.h"
#include "Types/FAuthTypes.h"
#include "UMMORPGLoginWidget.generated.h"

class UEditableTextBox;
class UButton;
class UTextBlock;
class UMMORPGAuthSubsystem;

UCLASS(Abstract)
class MMORPGUI_API UMMORPGLoginWidget : public UUserWidget
{
    GENERATED_BODY()

public:
    virtual void NativeConstruct() override;

protected:
    // UI Components
    UPROPERTY(meta = (BindWidget))
    UEditableTextBox* EmailTextBox;

    UPROPERTY(meta = (BindWidget))
    UEditableTextBox* PasswordTextBox;

    UPROPERTY(meta = (BindWidget))
    UButton* LoginButton;

    UPROPERTY(meta = (BindWidget))
    UButton* RegisterButton;

    UPROPERTY(meta = (BindWidget))
    UTextBlock* ErrorText;

    // Button click handlers
    UFUNCTION()
    void OnLoginClicked();

    UFUNCTION()
    void OnRegisterClicked();

    // Auth response handler
    UFUNCTION()
    void OnLoginResponse(const FAuthResponse& Response);

    // Blueprint events
    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|Auth")
    void OnLoginSuccess();

    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|Auth")
    void OnLoginFailed(const FString& ErrorMessage);

    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|Auth")
    void SwitchToRegisterView();

private:
    UPROPERTY()
    UMMORPGAuthSubsystem* AuthSubsystem;

    void SetErrorMessage(const FString& Message);
    void ClearErrorMessage();
};