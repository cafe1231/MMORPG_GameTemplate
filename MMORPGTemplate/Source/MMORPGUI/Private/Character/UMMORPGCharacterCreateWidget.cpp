#include "Character/UMMORPGCharacterCreateWidget.h"
#include "Subsystems/UMMORPGCharacterSubsystem.h"
#include "Components/Button.h"
#include "Components/EditableTextBox.h"
#include "Components/ComboBoxString.h"
#include "Components/TextBlock.h"
#include "Components/Slider.h"
#include "Kismet/GameplayStatics.h"

void UMMORPGCharacterCreateWidget::NativeConstruct()
{
    Super::NativeConstruct();

    // Get character subsystem
    if (UGameInstance* GameInstance = GetGameInstance())
    {
        CharacterSubsystem = GameInstance->GetSubsystem<UMMORPGCharacterSubsystem>();
        if (CharacterSubsystem)
        {
            // Bind to subsystem events
            CharacterSubsystem->OnCharacterCreated.AddDynamic(this, &UMMORPGCharacterCreateWidget::OnCharacterCreated);
            CharacterSubsystem->OnCharacterError.AddDynamic(this, &UMMORPGCharacterCreateWidget::OnCharacterError);
            CharacterSubsystem->OnCharacterListReceived.AddDynamic(this, &UMMORPGCharacterCreateWidget::OnCharacterListReceived);
            
            // Get current character count
            MaxCharacterSlots = CharacterSubsystem->GetMaxCharacterSlots();
            CharacterSubsystem->GetCharacterList();
        }
    }

    // Initialize predefined colors
    PredefinedSkinColors = {
        TEXT("#FFE0BD"), // Light
        TEXT("#FFD4B2"), // Fair
        TEXT("#F0C8A0"), // Medium
        TEXT("#D4A76A"), // Tan
        TEXT("#8D5524"), // Brown
        TEXT("#5D4037")  // Dark
    };

    PredefinedHairColors = {
        TEXT("#000000"), // Black
        TEXT("#4A3728"), // Dark Brown
        TEXT("#8B4513"), // Brown
        TEXT("#D2691E"), // Light Brown
        TEXT("#FFD700"), // Blonde
        TEXT("#DC143C"), // Red
        TEXT("#808080")  // Gray
    };

    PredefinedEyeColors = {
        TEXT("#8B4513"), // Brown
        TEXT("#0066CC"), // Blue
        TEXT("#228B22"), // Green
        TEXT("#708090"), // Gray
        TEXT("#FFD700"), // Amber
        TEXT("#8B008B")  // Violet
    };

    // Bind button events
    if (CreateButton)
    {
        CreateButton->OnClicked.AddDynamic(this, &UMMORPGCharacterCreateWidget::OnCreateClicked);
    }

    if (CancelButton)
    {
        CancelButton->OnClicked.AddDynamic(this, &UMMORPGCharacterCreateWidget::OnCancelClicked);
    }

    if (RandomizeButton)
    {
        RandomizeButton->OnClicked.AddDynamic(this, &UMMORPGCharacterCreateWidget::OnRandomizeClicked);
    }

    // Bind input events
    if (NameTextBox)
    {
        NameTextBox->OnTextChanged.AddDynamic(this, &UMMORPGCharacterCreateWidget::OnNameChanged);
    }

    if (ClassComboBox)
    {
        ClassComboBox->OnSelectionChanged.AddDynamic(this, &UMMORPGCharacterCreateWidget::OnClassChanged);
    }

    if (RaceComboBox)
    {
        RaceComboBox->OnSelectionChanged.AddDynamic(this, &UMMORPGCharacterCreateWidget::OnRaceChanged);
    }

    if (GenderComboBox)
    {
        GenderComboBox->OnSelectionChanged.AddDynamic(this, &UMMORPGCharacterCreateWidget::OnGenderChanged);
    }

    if (HeightSlider)
    {
        HeightSlider->OnValueChanged.AddDynamic(this, &UMMORPGCharacterCreateWidget::OnHeightChanged);
        HeightSlider->SetMinValue(0.8f);
        HeightSlider->SetMaxValue(1.2f);
        HeightSlider->SetValue(1.0f);
    }

    if (BuildSlider)
    {
        BuildSlider->OnValueChanged.AddDynamic(this, &UMMORPGCharacterCreateWidget::OnBuildChanged);
        BuildSlider->SetMinValue(0.8f);
        BuildSlider->SetMaxValue(1.2f);
        BuildSlider->SetValue(1.0f);
    }

    InitializeWidget();
}

void UMMORPGCharacterCreateWidget::NativeDestruct()
{
    // Unbind from subsystem events
    if (CharacterSubsystem)
    {
        CharacterSubsystem->OnCharacterCreated.RemoveDynamic(this, &UMMORPGCharacterCreateWidget::OnCharacterCreated);
        CharacterSubsystem->OnCharacterError.RemoveDynamic(this, &UMMORPGCharacterCreateWidget::OnCharacterError);
        CharacterSubsystem->OnCharacterListReceived.RemoveDynamic(this, &UMMORPGCharacterCreateWidget::OnCharacterListReceived);
    }

    Super::NativeDestruct();
}

void UMMORPGCharacterCreateWidget::InitializeWidget()
{
    PopulateDropdowns();
    ResetForm();
    UpdateCharacterCount();
}

void UMMORPGCharacterCreateWidget::ResetForm()
{
    if (NameTextBox)
    {
        NameTextBox->SetText(FText::GetEmpty());
    }

    if (ClassComboBox && ClassComboBox->GetOptionCount() > 0)
    {
        ClassComboBox->SetSelectedIndex(0);
    }

    if (RaceComboBox && RaceComboBox->GetOptionCount() > 0)
    {
        RaceComboBox->SetSelectedIndex(0);
    }

    if (GenderComboBox && GenderComboBox->GetOptionCount() > 0)
    {
        GenderComboBox->SetSelectedIndex(0);
    }

    if (HeightSlider)
    {
        HeightSlider->SetValue(1.0f);
    }

    if (BuildSlider)
    {
        BuildSlider->SetValue(1.0f);
    }

    if (SkinColorComboBox && SkinColorComboBox->GetOptionCount() > 0)
    {
        SkinColorComboBox->SetSelectedIndex(1); // Fair skin default
    }

    if (HairColorComboBox && HairColorComboBox->GetOptionCount() > 0)
    {
        HairColorComboBox->SetSelectedIndex(2); // Brown hair default
    }

    if (EyeColorComboBox && EyeColorComboBox->GetOptionCount() > 0)
    {
        EyeColorComboBox->SetSelectedIndex(1); // Blue eyes default
    }

    if (FaceStyleComboBox && FaceStyleComboBox->GetOptionCount() > 0)
    {
        FaceStyleComboBox->SetSelectedIndex(0);
    }

    if (HairStyleComboBox && HairStyleComboBox->GetOptionCount() > 0)
    {
        HairStyleComboBox->SetSelectedIndex(0);
    }

    ClearError();
}

bool UMMORPGCharacterCreateWidget::ValidateForm(FString& OutError)
{
    // Check character name
    if (NameTextBox)
    {
        FString Name = NameTextBox->GetText().ToString().TrimStartAndEnd();
        if (Name.IsEmpty())
        {
            OutError = TEXT("Please enter a character name");
            return false;
        }

        if (Name.Len() < 3)
        {
            OutError = TEXT("Character name must be at least 3 characters long");
            return false;
        }

        if (Name.Len() > 16)
        {
            OutError = TEXT("Character name must be 16 characters or less");
            return false;
        }

        // Check for valid characters
        for (TCHAR Char : Name)
        {
            if (!FChar::IsAlnum(Char))
            {
                OutError = TEXT("Character name can only contain letters and numbers");
                return false;
            }
        }
    }

    // Check if we can create more characters
    if (!CharacterSubsystem || !CharacterSubsystem->CanCreateMoreCharacters())
    {
        OutError = TEXT("Maximum character limit reached");
        return false;
    }

    return true;
}

void UMMORPGCharacterCreateWidget::PopulateDropdowns()
{
    // Populate class dropdown
    if (ClassComboBox)
    {
        ClassComboBox->ClearOptions();
        ClassComboBox->AddOption(TEXT("Warrior"));
        ClassComboBox->AddOption(TEXT("Mage"));
        ClassComboBox->AddOption(TEXT("Archer"));
        ClassComboBox->AddOption(TEXT("Rogue"));
        ClassComboBox->AddOption(TEXT("Priest"));
        ClassComboBox->AddOption(TEXT("Paladin"));
        ClassComboBox->SetSelectedIndex(0);
    }

    // Populate race dropdown
    if (RaceComboBox)
    {
        RaceComboBox->ClearOptions();
        RaceComboBox->AddOption(TEXT("Human"));
        RaceComboBox->AddOption(TEXT("Elf"));
        RaceComboBox->AddOption(TEXT("Dwarf"));
        RaceComboBox->AddOption(TEXT("Orc"));
        RaceComboBox->AddOption(TEXT("Undead"));
        RaceComboBox->SetSelectedIndex(0);
    }

    // Populate gender dropdown
    if (GenderComboBox)
    {
        GenderComboBox->ClearOptions();
        GenderComboBox->AddOption(TEXT("Male"));
        GenderComboBox->AddOption(TEXT("Female"));
        GenderComboBox->AddOption(TEXT("Other"));
        GenderComboBox->SetSelectedIndex(0);
    }

    // Populate color dropdowns with display names
    if (SkinColorComboBox)
    {
        SkinColorComboBox->ClearOptions();
        SkinColorComboBox->AddOption(TEXT("Light"));
        SkinColorComboBox->AddOption(TEXT("Fair"));
        SkinColorComboBox->AddOption(TEXT("Medium"));
        SkinColorComboBox->AddOption(TEXT("Tan"));
        SkinColorComboBox->AddOption(TEXT("Brown"));
        SkinColorComboBox->AddOption(TEXT("Dark"));
        SkinColorComboBox->SetSelectedIndex(1);
    }

    if (HairColorComboBox)
    {
        HairColorComboBox->ClearOptions();
        HairColorComboBox->AddOption(TEXT("Black"));
        HairColorComboBox->AddOption(TEXT("Dark Brown"));
        HairColorComboBox->AddOption(TEXT("Brown"));
        HairColorComboBox->AddOption(TEXT("Light Brown"));
        HairColorComboBox->AddOption(TEXT("Blonde"));
        HairColorComboBox->AddOption(TEXT("Red"));
        HairColorComboBox->AddOption(TEXT("Gray"));
        HairColorComboBox->SetSelectedIndex(2);
    }

    if (EyeColorComboBox)
    {
        EyeColorComboBox->ClearOptions();
        EyeColorComboBox->AddOption(TEXT("Brown"));
        EyeColorComboBox->AddOption(TEXT("Blue"));
        EyeColorComboBox->AddOption(TEXT("Green"));
        EyeColorComboBox->AddOption(TEXT("Gray"));
        EyeColorComboBox->AddOption(TEXT("Amber"));
        EyeColorComboBox->AddOption(TEXT("Violet"));
        EyeColorComboBox->SetSelectedIndex(1);
    }

    // Populate style dropdowns
    if (FaceStyleComboBox)
    {
        FaceStyleComboBox->ClearOptions();
        for (int32 i = 1; i <= 5; ++i)
        {
            FaceStyleComboBox->AddOption(FString::Printf(TEXT("Face %d"), i));
        }
        FaceStyleComboBox->SetSelectedIndex(0);
    }

    if (HairStyleComboBox)
    {
        HairStyleComboBox->ClearOptions();
        for (int32 i = 1; i <= 10; ++i)
        {
            HairStyleComboBox->AddOption(FString::Printf(TEXT("Hair %d"), i));
        }
        HairStyleComboBox->SetSelectedIndex(0);
    }
}

void UMMORPGCharacterCreateWidget::OnCreateClicked()
{
    if (bIsCreating)
        return;

    // Validate form
    FString ValidationError;
    if (!ValidateForm(ValidationError))
    {
        ShowError(ValidationError);
        return;
    }

    // Build character creation request
    FCharacterCreateRequest Request = BuildCreateRequest();

    // Start creation process
    bIsCreating = true;
    SetFormEnabled(false);
    ClearError();

    OnCharacterCreationStarted();

    // Submit to subsystem
    if (CharacterSubsystem)
    {
        CharacterSubsystem->CreateCharacter(Request);
    }
}

void UMMORPGCharacterCreateWidget::OnCancelClicked()
{
    OnCharacterCreationCancelled();
}

void UMMORPGCharacterCreateWidget::OnRandomizeClicked()
{
    // Generate random name
    if (NameTextBox)
    {
        NameTextBox->SetText(FText::FromString(GetRandomName()));
    }

    // Randomize appearance
    RandomizeAppearance();
}

void UMMORPGCharacterCreateWidget::OnNameChanged(const FText& Text)
{
    ClearError();
}

void UMMORPGCharacterCreateWidget::OnClassChanged(FString SelectedItem, ESelectInfo::Type SelectionType)
{
    // Could update preview based on class
}

void UMMORPGCharacterCreateWidget::OnRaceChanged(FString SelectedItem, ESelectInfo::Type SelectionType)
{
    // Could update available customization options based on race
}

void UMMORPGCharacterCreateWidget::OnGenderChanged(FString SelectedItem, ESelectInfo::Type SelectionType)
{
    // Could update preview based on gender
}

void UMMORPGCharacterCreateWidget::OnHeightChanged(float Value)
{
    if (HeightValueText)
    {
        HeightValueText->SetText(FText::FromString(FString::Printf(TEXT("%.0f%%"), Value * 100)));
    }
}

void UMMORPGCharacterCreateWidget::OnBuildChanged(float Value)
{
    if (BuildValueText)
    {
        BuildValueText->SetText(FText::FromString(FString::Printf(TEXT("%.0f%%"), Value * 100)));
    }
}

void UMMORPGCharacterCreateWidget::OnCharacterCreated(const FCharacterResponse& Response)
{
    bIsCreating = false;
    SetFormEnabled(true);

    if (Response.bSuccess)
    {
        OnCharacterCreationCompleted(Response.Character);
    }
    else
    {
        ShowError(Response.ErrorMessage);
    }
}

void UMMORPGCharacterCreateWidget::OnCharacterError(const FString& ErrorMessage)
{
    bIsCreating = false;
    SetFormEnabled(true);
    ShowError(ErrorMessage);
}

void UMMORPGCharacterCreateWidget::OnCharacterListReceived(const FCharacterListResponse& Response)
{
    CurrentCharacterCount = Response.Characters.Num();
    UpdateCharacterCount();
}

void UMMORPGCharacterCreateWidget::UpdateCharacterCount()
{
    if (CharacterCountText)
    {
        CharacterCountText->SetText(FText::FromString(
            FString::Printf(TEXT("Characters: %d/%d"), CurrentCharacterCount, MaxCharacterSlots)
        ));
    }

    // Disable create button if at max
    if (CreateButton && CharacterSubsystem)
    {
        CreateButton->SetIsEnabled(CharacterSubsystem->CanCreateMoreCharacters());
    }
}

void UMMORPGCharacterCreateWidget::ShowError(const FString& Message)
{
    if (ErrorMessageText)
    {
        ErrorMessageText->SetText(FText::FromString(Message));
        ErrorMessageText->SetVisibility(ESlateVisibility::Visible);
    }
}

void UMMORPGCharacterCreateWidget::ClearError()
{
    if (ErrorMessageText)
    {
        ErrorMessageText->SetText(FText::GetEmpty());
        ErrorMessageText->SetVisibility(ESlateVisibility::Collapsed);
    }
}

void UMMORPGCharacterCreateWidget::SetFormEnabled(bool bEnabled)
{
    if (NameTextBox) NameTextBox->SetIsEnabled(bEnabled);
    if (ClassComboBox) ClassComboBox->SetIsEnabled(bEnabled);
    if (RaceComboBox) RaceComboBox->SetIsEnabled(bEnabled);
    if (GenderComboBox) GenderComboBox->SetIsEnabled(bEnabled);
    if (HeightSlider) HeightSlider->SetIsEnabled(bEnabled);
    if (BuildSlider) BuildSlider->SetIsEnabled(bEnabled);
    if (SkinColorComboBox) SkinColorComboBox->SetIsEnabled(bEnabled);
    if (HairColorComboBox) HairColorComboBox->SetIsEnabled(bEnabled);
    if (EyeColorComboBox) EyeColorComboBox->SetIsEnabled(bEnabled);
    if (FaceStyleComboBox) FaceStyleComboBox->SetIsEnabled(bEnabled);
    if (HairStyleComboBox) HairStyleComboBox->SetIsEnabled(bEnabled);
    if (CreateButton) CreateButton->SetIsEnabled(bEnabled && CharacterSubsystem && CharacterSubsystem->CanCreateMoreCharacters());
    if (RandomizeButton) RandomizeButton->SetIsEnabled(bEnabled);
}

FCharacterCreateRequest UMMORPGCharacterCreateWidget::BuildCreateRequest() const
{
    FCharacterCreateRequest Request;

    // Set name
    if (NameTextBox)
    {
        Request.Name = NameTextBox->GetText().ToString().TrimStartAndEnd();
    }

    // Set class
    if (ClassComboBox)
    {
        Request.Class = ClassComboBox->GetSelectedOption();
    }

    // Set race
    if (RaceComboBox)
    {
        Request.Race = StringToCharacterRace(RaceComboBox->GetSelectedOption());
    }

    // Set appearance
    if (GenderComboBox)
    {
        Request.Appearance.Gender = StringToCharacterGender(GenderComboBox->GetSelectedOption());
    }

    if (HeightSlider)
    {
        Request.Appearance.Height = HeightSlider->GetValue();
    }

    if (BuildSlider)
    {
        Request.Appearance.Build = BuildSlider->GetValue();
    }

    // Set colors
    if (SkinColorComboBox)
    {
        int32 Index = SkinColorComboBox->GetSelectedIndex();
        if (PredefinedSkinColors.IsValidIndex(Index))
        {
            Request.Appearance.SkinColor = PredefinedSkinColors[Index];
        }
    }

    if (HairColorComboBox)
    {
        int32 Index = HairColorComboBox->GetSelectedIndex();
        if (PredefinedHairColors.IsValidIndex(Index))
        {
            Request.Appearance.HairColor = PredefinedHairColors[Index];
        }
    }

    if (EyeColorComboBox)
    {
        int32 Index = EyeColorComboBox->GetSelectedIndex();
        if (PredefinedEyeColors.IsValidIndex(Index))
        {
            Request.Appearance.EyeColor = PredefinedEyeColors[Index];
        }
    }

    // Set styles
    if (FaceStyleComboBox)
    {
        Request.Appearance.FaceID = FaceStyleComboBox->GetSelectedIndex() + 1;
    }

    if (HairStyleComboBox)
    {
        Request.Appearance.HairID = HairStyleComboBox->GetSelectedIndex() + 1;
    }

    return Request;
}

FString UMMORPGCharacterCreateWidget::GetRandomName() const
{
    // Simple random name generator
    TArray<FString> Prefixes = {
        TEXT("Aether"), TEXT("Storm"), TEXT("Shadow"), TEXT("Fire"), TEXT("Ice"),
        TEXT("Thunder"), TEXT("Dragon"), TEXT("Phoenix"), TEXT("Wolf"), TEXT("Eagle"),
        TEXT("Raven"), TEXT("Lion"), TEXT("Tiger"), TEXT("Bear"), TEXT("Falcon")
    };

    TArray<FString> Suffixes = {
        TEXT("blade"), TEXT("heart"), TEXT("soul"), TEXT("fist"), TEXT("eye"),
        TEXT("claw"), TEXT("wing"), TEXT("tail"), TEXT("mane"), TEXT("bane"),
        TEXT("walker"), TEXT("runner"), TEXT("hunter"), TEXT("seeker"), TEXT("keeper")
    };

    int32 PrefixIndex = FMath::RandRange(0, Prefixes.Num() - 1);
    int32 SuffixIndex = FMath::RandRange(0, Suffixes.Num() - 1);

    return Prefixes[PrefixIndex] + Suffixes[SuffixIndex];
}

void UMMORPGCharacterCreateWidget::RandomizeAppearance()
{
    // Randomize all appearance options
    if (ClassComboBox && ClassComboBox->GetOptionCount() > 0)
    {
        ClassComboBox->SetSelectedIndex(FMath::RandRange(0, ClassComboBox->GetOptionCount() - 1));
    }

    if (RaceComboBox && RaceComboBox->GetOptionCount() > 0)
    {
        RaceComboBox->SetSelectedIndex(FMath::RandRange(0, RaceComboBox->GetOptionCount() - 1));
    }

    if (GenderComboBox && GenderComboBox->GetOptionCount() > 0)
    {
        GenderComboBox->SetSelectedIndex(FMath::RandRange(0, GenderComboBox->GetOptionCount() - 1));
    }

    if (HeightSlider)
    {
        HeightSlider->SetValue(FMath::RandRange(0.8f, 1.2f));
    }

    if (BuildSlider)
    {
        BuildSlider->SetValue(FMath::RandRange(0.8f, 1.2f));
    }

    if (SkinColorComboBox && SkinColorComboBox->GetOptionCount() > 0)
    {
        SkinColorComboBox->SetSelectedIndex(FMath::RandRange(0, SkinColorComboBox->GetOptionCount() - 1));
    }

    if (HairColorComboBox && HairColorComboBox->GetOptionCount() > 0)
    {
        HairColorComboBox->SetSelectedIndex(FMath::RandRange(0, HairColorComboBox->GetOptionCount() - 1));
    }

    if (EyeColorComboBox && EyeColorComboBox->GetOptionCount() > 0)
    {
        EyeColorComboBox->SetSelectedIndex(FMath::RandRange(0, EyeColorComboBox->GetOptionCount() - 1));
    }

    if (FaceStyleComboBox && FaceStyleComboBox->GetOptionCount() > 0)
    {
        FaceStyleComboBox->SetSelectedIndex(FMath::RandRange(0, FaceStyleComboBox->GetOptionCount() - 1));
    }

    if (HairStyleComboBox && HairStyleComboBox->GetOptionCount() > 0)
    {
        HairStyleComboBox->SetSelectedIndex(FMath::RandRange(0, HairStyleComboBox->GetOptionCount() - 1));
    }
}