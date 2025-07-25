#include "Auth/UMMORPGLoginWidget.h"
#include "Components/EditableTextBox.h"
#include "Components/Button.h"
#include "Components/TextBlock.h"
#include "Subsystems/UMMORPGAuthSubsystem.h"
#include "Engine/GameInstance.h"

void UMMORPGLoginWidget::NativeConstruct()
{
    Super::NativeConstruct();

    // Get auth subsystem
    if (UGameInstance* GameInstance = GetGameInstance())
    {
        AuthSubsystem = GameInstance->GetSubsystem<UMMORPGAuthSubsystem>();
        if (AuthSubsystem)
        {
            AuthSubsystem->OnLoginResponse.AddDynamic(this, &UMMORPGLoginWidget::OnLoginResponse);
        }
    }

    // Bind button clicks
    if (LoginButton)
    {
        LoginButton->OnClicked.AddDynamic(this, &UMMORPGLoginWidget::OnLoginClicked);
    }

    if (RegisterButton)
    {
        RegisterButton->OnClicked.AddDynamic(this, &UMMORPGLoginWidget::OnRegisterClicked);
    }

    // Clear error text
    ClearErrorMessage();
}

void UMMORPGLoginWidget::OnLoginClicked()
{
    if (!AuthSubsystem)
    {
        SetErrorMessage(TEXT("Authentication system not available"));
        return;
    }

    // Validate inputs
    FString Email = EmailTextBox ? EmailTextBox->GetText().ToString() : TEXT("");
    FString Password = PasswordTextBox ? PasswordTextBox->GetText().ToString() : TEXT("");

    if (Email.IsEmpty() || Password.IsEmpty())
    {
        SetErrorMessage(TEXT("Please enter email and password"));
        return;
    }

    // Clear error and attempt login
    ClearErrorMessage();

    FLoginRequest Request;
    Request.Email = Email;
    Request.Password = Password;

    AuthSubsystem->Login(Request);
}

void UMMORPGLoginWidget::OnRegisterClicked()
{
    SwitchToRegisterView();
}

void UMMORPGLoginWidget::OnLoginResponse(const FAuthResponse& Response)
{
    if (Response.bSuccess)
    {
        OnLoginSuccess();
    }
    else
    {
        FString ErrorMessage = Response.Message.IsEmpty() ? TEXT("Login failed") : Response.Message;
        SetErrorMessage(ErrorMessage);
        OnLoginFailed(ErrorMessage);
    }
}

void UMMORPGLoginWidget::SetErrorMessage(const FString& Message)
{
    if (ErrorText)
    {
        ErrorText->SetText(FText::FromString(Message));
        ErrorText->SetVisibility(ESlateVisibility::Visible);
    }
}

void UMMORPGLoginWidget::ClearErrorMessage()
{
    if (ErrorText)
    {
        ErrorText->SetText(FText::GetEmpty());
        ErrorText->SetVisibility(ESlateVisibility::Collapsed);
    }
}