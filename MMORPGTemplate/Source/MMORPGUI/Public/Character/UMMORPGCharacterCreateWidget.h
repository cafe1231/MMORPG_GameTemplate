#pragma once

#include "CoreMinimal.h"
#include "Blueprint/UserWidget.h"
#include "Types/FCharacterTypes.h"
#include "UMMORPGCharacterCreateWidget.generated.h"

// Forward declarations
class UButton;
class UEditableTextBox;
class UComboBoxString;
class UTextBlock;
class USlider;
class UImage;
class UMMORPGCharacterSubsystem;

/**
 * Character creation widget for MMORPG
 * Allows players to create a new character with name, class, race, and appearance customization
 */
UCLASS()
class MMORPGUI_API UMMORPGCharacterCreateWidget : public UUserWidget
{
    GENERATED_BODY()

public:
    // Widget initialization
    virtual void NativeConstruct() override;
    virtual void NativeDestruct() override;

    // Initialize the widget with default values
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character")
    void InitializeWidget();

    // Reset form to default values
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character")
    void ResetForm();

    // Validate form before submission
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character")
    bool ValidateForm(FString& OutError);

protected:
    // UI Components - Name
    UPROPERTY(meta = (BindWidget))
    UEditableTextBox* NameTextBox;

    // UI Components - Class/Race Selection
    UPROPERTY(meta = (BindWidget))
    UComboBoxString* ClassComboBox;

    UPROPERTY(meta = (BindWidget))
    UComboBoxString* RaceComboBox;

    // UI Components - Appearance
    UPROPERTY(meta = (BindWidget))
    UComboBoxString* GenderComboBox;

    UPROPERTY(meta = (BindWidget))
    USlider* HeightSlider;

    UPROPERTY(meta = (BindWidget))
    USlider* BuildSlider;

    UPROPERTY(meta = (BindWidget))
    UTextBlock* HeightValueText;

    UPROPERTY(meta = (BindWidget))
    UTextBlock* BuildValueText;

    // UI Components - Color Pickers (simplified for now)
    UPROPERTY(meta = (BindWidget))
    UComboBoxString* SkinColorComboBox;

    UPROPERTY(meta = (BindWidget))
    UComboBoxString* HairColorComboBox;

    UPROPERTY(meta = (BindWidget))
    UComboBoxString* EyeColorComboBox;

    // UI Components - Face/Hair Selection
    UPROPERTY(meta = (BindWidget))
    UComboBoxString* FaceStyleComboBox;

    UPROPERTY(meta = (BindWidget))
    UComboBoxString* HairStyleComboBox;

    // UI Components - Actions
    UPROPERTY(meta = (BindWidget))
    UButton* CreateButton;

    UPROPERTY(meta = (BindWidget))
    UButton* CancelButton;

    UPROPERTY(meta = (BindWidget))
    UButton* RandomizeButton;

    // UI Components - Feedback
    UPROPERTY(meta = (BindWidget))
    UTextBlock* ErrorMessageText;

    UPROPERTY(meta = (BindWidget))
    UTextBlock* CharacterCountText;

    // Button click handlers
    UFUNCTION()
    void OnCreateClicked();

    UFUNCTION()
    void OnCancelClicked();

    UFUNCTION()
    void OnRandomizeClicked();

    // Value change handlers
    UFUNCTION()
    void OnNameChanged(const FText& Text);

    UFUNCTION()
    void OnClassChanged(FString SelectedItem, ESelectInfo::Type SelectionType);

    UFUNCTION()
    void OnRaceChanged(FString SelectedItem, ESelectInfo::Type SelectionType);

    UFUNCTION()
    void OnGenderChanged(FString SelectedItem, ESelectInfo::Type SelectionType);

    UFUNCTION()
    void OnHeightChanged(float Value);

    UFUNCTION()
    void OnBuildChanged(float Value);

    // Character subsystem callbacks
    UFUNCTION()
    void OnCharacterCreated(const FCharacterResponse& Response);

    UFUNCTION()
    void OnCharacterError(const FString& ErrorMessage);

    UFUNCTION()
    void OnCharacterListReceived(const FCharacterListResponse& Response);

    // Helper functions
    void PopulateDropdowns();
    void UpdateCharacterCount();
    void ShowError(const FString& Message);
    void ClearError();
    void SetFormEnabled(bool bEnabled);
    FCharacterCreateRequest BuildCreateRequest() const;
    FString GetRandomName() const;
    void RandomizeAppearance();

    // Events for Blueprint
    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|Character")
    void OnCharacterCreationStarted();

    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|Character")
    void OnCharacterCreationCompleted(const FCharacterInfo& Character);

    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|Character")
    void OnCharacterCreationCancelled();

private:
    // Subsystem reference
    UPROPERTY()
    UMMORPGCharacterSubsystem* CharacterSubsystem;

    // State tracking
    bool bIsCreating = false;
    int32 CurrentCharacterCount = 0;
    int32 MaxCharacterSlots = 5;

    // Predefined options
    TArray<FString> PredefinedSkinColors;
    TArray<FString> PredefinedHairColors;
    TArray<FString> PredefinedEyeColors;
};