#include "Auth/UMMORPGAuthWidget.h"
#include "Auth/UMMORPGLoginWidget.h"
#include "Auth/UMMORPGRegisterWidget.h"
#include "Components/WidgetSwitcher.h"

void UMMORPGAuthWidget::NativeConstruct()
{
    Super::NativeConstruct();

    // Set initial view to login
    ShowLoginView();
}

void UMMORPGAuthWidget::ShowLoginView()
{
    if (AuthSwitcher && LoginWidget)
    {
        AuthSwitcher->SetActiveWidget(LoginWidget);
    }
}

void UMMORPGAuthWidget::ShowRegisterView()
{
    if (AuthSwitcher && RegisterWidget)
    {
        AuthSwitcher->SetActiveWidget(RegisterWidget);
    }
}

void UMMORPGAuthWidget::HandleLoginSuccess()
{
    OnAuthenticationSuccess();
}

void UMMORPGAuthWidget::HandleRegisterSuccess()
{
    // Registration success typically switches back to login
    ShowLoginView();
}