#pragma once

#include "CoreMinimal.h"
#include "Blueprint/UserWidget.h"
#include "UMMORPGAuthWidget.generated.h"

class UWidgetSwitcher;
class UMMORPGLoginWidget;
class UMMORPGRegisterWidget;

UCLASS(Abstract)
class MMORPGUI_API UMMORPGAuthWidget : public UUserWidget
{
    GENERATED_BODY()

public:
    virtual void NativeConstruct() override;

    // Switch between login and register views
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    void ShowLoginView();

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    void ShowRegisterView();

protected:
    // UI Components
    UPROPERTY(meta = (BindWidget))
    UWidgetSwitcher* AuthSwitcher;

    UPROPERTY(meta = (BindWidget))
    UMMORPGLoginWidget* LoginWidget;

    UPROPERTY(meta = (BindWidget))
    UMMORPGRegisterWidget* RegisterWidget;

    // Blueprint events
    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|Auth")
    void OnAuthenticationSuccess();

private:
    UFUNCTION()
    void HandleLoginSuccess();

    UFUNCTION()
    void HandleRegisterSuccess();
};