#include "Auth/UMMORPGRegisterWidget.h"
#include "Components/EditableTextBox.h"
#include "Components/Button.h"
#include "Components/TextBlock.h"
#include "Subsystems/UMMORPGAuthSubsystem.h"
#include "Engine/GameInstance.h"

void UMMORPGRegisterWidget::NativeConstruct()
{
    Super::NativeConstruct();

    // Get auth subsystem
    if (UGameInstance* GameInstance = GetGameInstance())
    {
        AuthSubsystem = GameInstance->GetSubsystem<UMMORPGAuthSubsystem>();
        if (AuthSubsystem)
        {
            AuthSubsystem->OnRegisterResponse.AddDynamic(this, &UMMORPGRegisterWidget::OnRegisterResponse);
        }
    }

    // Bind button clicks
    if (RegisterButton)
    {
        RegisterButton->OnClicked.AddDynamic(this, &UMMORPGRegisterWidget::OnRegisterClicked);
    }

    if (BackToLoginButton)
    {
        BackToLoginButton->OnClicked.AddDynamic(this, &UMMORPGRegisterWidget::OnBackToLoginClicked);
    }

    // Clear error text
    ClearErrorMessage();
}

void UMMORPGRegisterWidget::OnRegisterClicked()
{
    if (!AuthSubsystem)
    {
        SetErrorMessage(TEXT("Authentication system not available"));
        return;
    }

    FString Email, Username, Password;
    if (!ValidateInputs(Email, Username, Password))
    {
        return;
    }

    // Clear error and attempt registration
    ClearErrorMessage();

    FRegisterRequest Request;
    Request.Email = Email;
    Request.Username = Username;
    Request.Password = Password;

    AuthSubsystem->Register(Request);
}

void UMMORPGRegisterWidget::OnBackToLoginClicked()
{
    SwitchToLoginView();
}

void UMMORPGRegisterWidget::OnRegisterResponse(const FAuthResponse& Response)
{
    if (Response.bSuccess)
    {
        OnRegisterSuccess();
        // Switch back to login view after successful registration
        SwitchToLoginView();
    }
    else
    {
        FString ErrorMessage = Response.Message.IsEmpty() ? TEXT("Registration failed") : Response.Message;
        SetErrorMessage(ErrorMessage);
        OnRegisterFailed(ErrorMessage);
    }
}

bool UMMORPGRegisterWidget::ValidateInputs(FString& OutEmail, FString& OutUsername, FString& OutPassword)
{
    // Get input values
    OutEmail = EmailTextBox ? EmailTextBox->GetText().ToString() : TEXT("");
    OutUsername = UsernameTextBox ? UsernameTextBox->GetText().ToString() : TEXT("");
    OutPassword = PasswordTextBox ? PasswordTextBox->GetText().ToString() : TEXT("");
    FString ConfirmPassword = ConfirmPasswordTextBox ? ConfirmPasswordTextBox->GetText().ToString() : TEXT("");

    // Validate email
    if (OutEmail.IsEmpty())
    {
        SetErrorMessage(TEXT("Please enter your email"));
        return false;
    }

    // Basic email validation
    if (!OutEmail.Contains(TEXT("@")) || !OutEmail.Contains(TEXT(".")))
    {
        SetErrorMessage(TEXT("Please enter a valid email address"));
        return false;
    }

    // Validate username
    if (OutUsername.IsEmpty())
    {
        SetErrorMessage(TEXT("Please enter a username"));
        return false;
    }

    if (OutUsername.Len() < 3)
    {
        SetErrorMessage(TEXT("Username must be at least 3 characters"));
        return false;
    }

    // Validate password
    if (OutPassword.IsEmpty())
    {
        SetErrorMessage(TEXT("Please enter a password"));
        return false;
    }

    if (OutPassword.Len() < 6)
    {
        SetErrorMessage(TEXT("Password must be at least 6 characters"));
        return false;
    }

    // Validate password confirmation
    if (OutPassword != ConfirmPassword)
    {
        SetErrorMessage(TEXT("Passwords do not match"));
        return false;
    }

    return true;
}

void UMMORPGRegisterWidget::SetErrorMessage(const FString& Message)
{
    if (ErrorText)
    {
        ErrorText->SetText(FText::FromString(Message));
        ErrorText->SetVisibility(ESlateVisibility::Visible);
    }
}

void UMMORPGRegisterWidget::ClearErrorMessage()
{
    if (ErrorText)
    {
        ErrorText->SetText(FText::GetEmpty());
        ErrorText->SetVisibility(ESlateVisibility::Collapsed);
    }
}