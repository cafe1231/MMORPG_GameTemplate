#pragma once

#include "CoreMinimal.h"
#include "Blueprint/UserWidget.h"
#include "Types/FAuthTypes.h"
#include "UMMORPGRegisterWidget.generated.h"

class UEditableTextBox;
class UButton;
class UTextBlock;
class UMMORPGAuthSubsystem;

UCLASS(Abstract)
class MMORPGUI_API UMMORPGRegisterWidget : public UUserWidget
{
    GENERATED_BODY()

public:
    virtual void NativeConstruct() override;

protected:
    // UI Components
    UPROPERTY(meta = (BindWidget))
    UEditableTextBox* EmailTextBox;

    UPROPERTY(meta = (BindWidget))
    UEditableTextBox* UsernameTextBox;

    UPROPERTY(meta = (BindWidget))
    UEditableTextBox* PasswordTextBox;

    UPROPERTY(meta = (BindWidget))
    UEditableTextBox* ConfirmPasswordTextBox;

    UPROPERTY(meta = (BindWidget))
    UButton* RegisterButton;

    UPROPERTY(meta = (BindWidget))
    UButton* BackToLoginButton;

    UPROPERTY(meta = (BindWidget))
    UTextBlock* ErrorText;

    // Button click handlers
    UFUNCTION()
    void OnRegisterClicked();

    UFUNCTION()
    void OnBackToLoginClicked();

    // Auth response handler
    UFUNCTION()
    void OnRegisterResponse(const FAuthResponse& Response);

    // Blueprint events
    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|Auth")
    void OnRegisterSuccess();

    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|Auth")
    void OnRegisterFailed(const FString& ErrorMessage);

    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|Auth")
    void SwitchToLoginView();

private:
    UPROPERTY()
    UMMORPGAuthSubsystem* AuthSubsystem;

    void SetErrorMessage(const FString& Message);
    void ClearErrorMessage();
    bool ValidateInputs(FString& OutEmail, FString& OutUsername, FString& OutPassword);
};