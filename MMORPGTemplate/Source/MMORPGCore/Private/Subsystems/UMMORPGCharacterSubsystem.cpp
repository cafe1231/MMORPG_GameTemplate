#include "Subsystems/UMMORPGCharacterSubsystem.h"
#include "Subsystems/UMMORPGAuthSubsystem.h"
#include "Http/MMORPGHTTPClient.h"
#include "Types/FCharacterTypes.h"
#include "Types/FAuthTypes.h"
#include "Engine/World.h"
#include "Misc/ConfigCacheIni.h"
#include "TimerManager.h"

DEFINE_LOG_CATEGORY(LogMMORPGCharacter);

void UMMORPGCharacterSubsystem::Initialize(FSubsystemCollectionBase& Collection)
{
    Super::Initialize(Collection);

    // Get dependencies
    AuthSubsystem = GetGameInstance()->GetSubsystem<UMMORPGAuthSubsystem>();
    if (!AuthSubsystem)
    {
        UE_LOG(LogMMORPGCharacter, Error, TEXT("Failed to get AuthSubsystem"));
        return;
    }

    // Create HTTP client
    HTTPClient = NewObject<UMMORPGHTTPClient>(this);
    if (!HTTPClient)
    {
        UE_LOG(LogMMORPGCharacter, Error, TEXT("Failed to create HTTPClient"));
        return;
    }

    // Configure HTTP client
    FString ServerURL;
    if (GConfig->GetString(TEXT("MMORPG"), TEXT("ServerURL"), ServerURL, GGameIni))
    {
        HTTPClient->SetBaseURL(ServerURL);
    }
    else
    {
        // Default to localhost for development
        HTTPClient->SetBaseURL(TEXT("http://localhost:8090"));
    }

    // Load configuration
    GConfig->GetInt(TEXT("MMORPG.Character"), TEXT("MaxCharacterSlots"), MaxCharacterSlots, GGameIni);

    // Load previously selected character
    LoadSelectedCharacter();

    UE_LOG(LogMMORPGCharacter, Log, TEXT("Character subsystem initialized. Max slots: %d"), MaxCharacterSlots);
}

void UMMORPGCharacterSubsystem::Deinitialize()
{
    SaveSelectedCharacter();
    ClearCharacterCache();
    
    Super::Deinitialize();
}

void UMMORPGCharacterSubsystem::GetCharacterList()
{
    if (bIsRequestInProgress)
    {
        UE_LOG(LogMMORPGCharacter, Warning, TEXT("Character request already in progress"));
        return;
    }

    if (bUseMockMode)
    {
        MockGetCharacterList();
        return;
    }

    if (!AuthSubsystem || !AuthSubsystem->IsAuthenticated())
    {
        OnCharacterError.Broadcast(TEXT("Not authenticated"));
        return;
    }

    bIsRequestInProgress = true;

    // Create request
    FString URL = HTTPClient->GetBaseURL() + TEXT("/api/v1/characters");
    TMap<FString, FString> Headers;
    Headers.Add(TEXT("Authorization"), FString::Printf(TEXT("Bearer %s"), *AuthSubsystem->GetAuthTokens().AccessToken));

    // Set up response handler
    HTTPClient->OnRequestComplete.Clear();
    HTTPClient->OnRequestComplete.AddDynamic(this, &UMMORPGCharacterSubsystem::OnGetCharacterListResponse);

    // Send request
    HTTPClient->SendGetRequestWithHeaders(URL, Headers);
}

void UMMORPGCharacterSubsystem::OnGetCharacterListResponse(bool bWasSuccessful, int32 ResponseCode, const FString& ResponseContent)
{
    bIsRequestInProgress = false;
    
    if (bWasSuccessful && ResponseCode == 200)
    {
        FCharacterListResponse ListResponse;
        if (ListResponse.ParseFromJSON(ResponseContent))
        {
            HandleCharacterListResponse(ListResponse);
        }
        else
        {
            OnCharacterError.Broadcast(TEXT("Failed to parse character list response"));
        }
    }
    else
    {
        FString ErrorMsg = FString::Printf(TEXT("Failed to get character list. Response code: %d"), ResponseCode);
        OnCharacterError.Broadcast(ErrorMsg);
    }
}

void UMMORPGCharacterSubsystem::CreateCharacter(const FCharacterCreateRequest& Request)
{
    if (bIsRequestInProgress)
    {
        UE_LOG(LogMMORPGCharacter, Warning, TEXT("Character request already in progress"));
        return;
    }

    // Validate character name
    FString ValidationError;
    if (!ValidateCharacterName(Request.Name, ValidationError))
    {
        OnCharacterError.Broadcast(ValidationError);
        return;
    }

    // Check character limit
    if (!CanCreateMoreCharacters())
    {
        OnCharacterError.Broadcast(TEXT("Maximum character limit reached"));
        return;
    }

    if (bUseMockMode)
    {
        MockCreateCharacter(Request);
        return;
    }

    if (!AuthSubsystem || !AuthSubsystem->IsAuthenticated())
    {
        OnCharacterError.Broadcast(TEXT("Not authenticated"));
        return;
    }

    bIsRequestInProgress = true;

    // Create request
    FString URL = HTTPClient->GetBaseURL() + TEXT("/api/v1/characters");
    FString Body = Request.ToJSON();
    TMap<FString, FString> Headers;
    Headers.Add(TEXT("Authorization"), FString::Printf(TEXT("Bearer %s"), *AuthSubsystem->GetAuthTokens().AccessToken));
    Headers.Add(TEXT("Content-Type"), TEXT("application/json"));

    // Set up response handler
    HTTPClient->OnRequestComplete.Clear();
    HTTPClient->OnRequestComplete.AddDynamic(this, &UMMORPGCharacterSubsystem::OnCreateCharacterResponse);

    // Send request
    HTTPClient->SendPostRequestWithHeaders(URL, Body, Headers);
}

void UMMORPGCharacterSubsystem::OnCreateCharacterResponse(bool bWasSuccessful, int32 ResponseCode, const FString& ResponseContent)
{
    bIsRequestInProgress = false;
    
    if (bWasSuccessful && (ResponseCode == 200 || ResponseCode == 201))
    {
        FCharacterResponse CharResponse;
        if (CharResponse.ParseFromJSON(ResponseContent))
        {
            HandleCharacterCreateResponse(CharResponse);
        }
        else
        {
            OnCharacterError.Broadcast(TEXT("Failed to parse character creation response"));
        }
    }
    else
    {
        FString ErrorMsg = FString::Printf(TEXT("Failed to create character. Response code: %d"), ResponseCode);
        OnCharacterError.Broadcast(ErrorMsg);
    }
}

void UMMORPGCharacterSubsystem::SelectCharacter(const FString& CharacterID)
{
    if (CharacterID.IsEmpty())
    {
        OnCharacterError.Broadcast(TEXT("Invalid character ID"));
        return;
    }

    if (bUseMockMode)
    {
        MockSelectCharacter(CharacterID);
        return;
    }

    if (!AuthSubsystem || !AuthSubsystem->IsAuthenticated())
    {
        OnCharacterError.Broadcast(TEXT("Not authenticated"));
        return;
    }

    // Create request
    FString URL = HTTPClient->GetBaseURL() + FString::Printf(TEXT("/api/v1/characters/%s/select"), *CharacterID);
    TMap<FString, FString> Headers;
    Headers.Add(TEXT("Authorization"), FString::Printf(TEXT("Bearer %s"), *AuthSubsystem->GetAuthTokens().AccessToken));

    // Store character ID for callback
    PendingCharacterID = CharacterID;
    
    // Set up response handler
    HTTPClient->OnRequestComplete.Clear();
    HTTPClient->OnRequestComplete.AddDynamic(this, &UMMORPGCharacterSubsystem::OnSelectCharacterResponse);

    // Send request
    HTTPClient->SendPostRequestWithHeaders(URL, TEXT("{}"), Headers);
}

void UMMORPGCharacterSubsystem::OnSelectCharacterResponse(bool bWasSuccessful, int32 ResponseCode, const FString& ResponseContent)
{
    if (bWasSuccessful && ResponseCode == 200)
    {
        HandleCharacterSelectResponse(PendingCharacterID, true, TEXT(""));
    }
    else
    {
        FString ErrorMsg = FString::Printf(TEXT("Failed to select character. Response code: %d"), ResponseCode);
        HandleCharacterSelectResponse(PendingCharacterID, false, ErrorMsg);
    }
    PendingCharacterID.Empty();
}

void UMMORPGCharacterSubsystem::DeleteCharacter(const FString& CharacterID)
{
    if (CharacterID.IsEmpty())
    {
        OnCharacterError.Broadcast(TEXT("Invalid character ID"));
        return;
    }

    if (bUseMockMode)
    {
        MockDeleteCharacter(CharacterID);
        return;
    }

    if (!AuthSubsystem || !AuthSubsystem->IsAuthenticated())
    {
        OnCharacterError.Broadcast(TEXT("Not authenticated"));
        return;
    }

    // Create request
    FString URL = HTTPClient->GetBaseURL() + FString::Printf(TEXT("/api/v1/characters/%s"), *CharacterID);
    TMap<FString, FString> Headers;
    Headers.Add(TEXT("Authorization"), FString::Printf(TEXT("Bearer %s"), *AuthSubsystem->GetAuthTokens().AccessToken));

    // Note: UE4/5 HTTP module doesn't support DELETE verb in simplified interface
    // We'll need to create a custom request or modify the HTTPClient class
    // For now, we'll use POST with a delete action
    FString DeleteURL = HTTPClient->GetBaseURL() + FString::Printf(TEXT("/api/v1/characters/%s/delete"), *CharacterID);
    
    // Store character ID for callback
    PendingCharacterID = CharacterID;
    
    // Set up response handler
    HTTPClient->OnRequestComplete.Clear();
    HTTPClient->OnRequestComplete.AddDynamic(this, &UMMORPGCharacterSubsystem::OnDeleteCharacterResponse);

    // Send request as POST with delete endpoint
    HTTPClient->SendPostRequestWithHeaders(DeleteURL, TEXT("{}"), Headers);
}

void UMMORPGCharacterSubsystem::OnDeleteCharacterResponse(bool bWasSuccessful, int32 ResponseCode, const FString& ResponseContent)
{
    if (bWasSuccessful && (ResponseCode == 200 || ResponseCode == 204))
    {
        HandleCharacterDeleteResponse(PendingCharacterID, true, TEXT(""));
    }
    else
    {
        FString ErrorMsg = FString::Printf(TEXT("Failed to delete character. Response code: %d"), ResponseCode);
        HandleCharacterDeleteResponse(PendingCharacterID, false, ErrorMsg);
    }
    PendingCharacterID.Empty();
}

void UMMORPGCharacterSubsystem::UpdateCharacter(const FString& CharacterID, const FCharacterUpdateRequest& Request)
{
    // TODO: Implement character update functionality
    OnCharacterError.Broadcast(TEXT("Character update not yet implemented"));
}

bool UMMORPGCharacterSubsystem::GetCharacterByID(const FString& CharacterID, FCharacterInfo& OutCharacterInfo) const
{
    for (const FCharacterInfo& Character : CachedCharacters)
    {
        if (Character.ID == CharacterID)
        {
            OutCharacterInfo = Character;
            return true;
        }
    }
    return false;
}

void UMMORPGCharacterSubsystem::HandleCharacterListResponse(const FCharacterListResponse& Response)
{
    CachedCharacters = Response.Characters;
    OnCharacterListReceived.Broadcast(Response);
    
    UE_LOG(LogMMORPGCharacter, Log, TEXT("Received %d characters"), CachedCharacters.Num());
}

void UMMORPGCharacterSubsystem::HandleCharacterCreateResponse(const FCharacterResponse& Response)
{
    if (Response.bSuccess && Response.Character.IsValid())
    {
        CachedCharacters.Add(Response.Character);
        OnCharacterCreated.Broadcast(Response);
        
        UE_LOG(LogMMORPGCharacter, Log, TEXT("Character created: %s"), *Response.Character.Name);
    }
    else
    {
        OnCharacterError.Broadcast(Response.ErrorMessage);
    }
}

void UMMORPGCharacterSubsystem::HandleCharacterSelectResponse(const FString& CharacterID, bool bSuccess, const FString& ErrorMessage)
{
    if (bSuccess)
    {
        SelectedCharacterID = CharacterID;
        SaveSelectedCharacter();
        OnCharacterSelected.Broadcast(CharacterID);
        
        UE_LOG(LogMMORPGCharacter, Log, TEXT("Character selected: %s"), *CharacterID);
    }
    else
    {
        OnCharacterError.Broadcast(ErrorMessage);
    }
}

void UMMORPGCharacterSubsystem::HandleCharacterDeleteResponse(const FString& CharacterID, bool bSuccess, const FString& ErrorMessage)
{
    if (bSuccess)
    {
        // Remove from cache
        CachedCharacters.RemoveAll([CharacterID](const FCharacterInfo& Character)
        {
            return Character.ID == CharacterID;
        });

        // Clear selection if deleted character was selected
        if (SelectedCharacterID == CharacterID)
        {
            SelectedCharacterID.Empty();
            SaveSelectedCharacter();
        }

        OnCharacterDeleted.Broadcast(CharacterID);
        
        UE_LOG(LogMMORPGCharacter, Log, TEXT("Character deleted: %s"), *CharacterID);
    }
    else
    {
        OnCharacterError.Broadcast(ErrorMessage);
    }
}

void UMMORPGCharacterSubsystem::HandleCharacterUpdateResponse(const FCharacterResponse& Response)
{
    // TODO: Implement update response handling
}

void UMMORPGCharacterSubsystem::MockGetCharacterList()
{
    // Create mock characters for testing
    FCharacterListResponse MockResponse;
    MockResponse.bSuccess = true;

    // Add some test characters
    FCharacterInfo Character1;
    Character1.ID = TEXT("mock_char_1");
    Character1.Name = TEXT("TestWarrior");
    Character1.Class = TEXT("warrior");
    Character1.Level = 10;
    Character1.CreatedAt = FDateTime::Now();
    MockResponse.Characters.Add(Character1);

    FCharacterInfo Character2;
    Character2.ID = TEXT("mock_char_2");
    Character2.Name = TEXT("TestMage");
    Character2.Class = TEXT("mage");
    Character2.Level = 5;
    Character2.CreatedAt = FDateTime::Now();
    MockResponse.Characters.Add(Character2);

    // Simulate async response
    FTimerHandle TimerHandle;
    GetGameInstance()->GetWorld()->GetTimerManager().SetTimer(TimerHandle, [this, MockResponse]()
    {
        HandleCharacterListResponse(MockResponse);
    }, 0.5f, false);
}

void UMMORPGCharacterSubsystem::MockCreateCharacter(const FCharacterCreateRequest& Request)
{
    // Create mock response
    FCharacterResponse MockResponse;
    MockResponse.bSuccess = true;
    
    FCharacterInfo NewCharacter;
    NewCharacter.ID = FString::Printf(TEXT("mock_char_%d"), FMath::RandRange(1000, 9999));
    NewCharacter.Name = Request.Name;
    NewCharacter.Class = Request.Class;
    NewCharacter.Level = 1;
    NewCharacter.CreatedAt = FDateTime::Now();
    
    MockResponse.Character = NewCharacter;

    // Simulate async response
    FTimerHandle TimerHandle;
    GetGameInstance()->GetWorld()->GetTimerManager().SetTimer(TimerHandle, [this, MockResponse]()
    {
        HandleCharacterCreateResponse(MockResponse);
    }, 0.5f, false);
}

void UMMORPGCharacterSubsystem::MockSelectCharacter(const FString& CharacterID)
{
    // Check if character exists in cache
    bool bFound = false;
    for (const FCharacterInfo& Character : CachedCharacters)
    {
        if (Character.ID == CharacterID)
        {
            bFound = true;
            break;
        }
    }

    // Simulate async response
    FTimerHandle TimerHandle;
    GetGameInstance()->GetWorld()->GetTimerManager().SetTimer(TimerHandle, [this, CharacterID, bFound]()
    {
        if (bFound)
        {
            HandleCharacterSelectResponse(CharacterID, true, TEXT(""));
        }
        else
        {
            HandleCharacterSelectResponse(CharacterID, false, TEXT("Character not found"));
        }
    }, 0.3f, false);
}

void UMMORPGCharacterSubsystem::MockDeleteCharacter(const FString& CharacterID)
{
    // Simulate async response
    FTimerHandle TimerHandle;
    GetGameInstance()->GetWorld()->GetTimerManager().SetTimer(TimerHandle, [this, CharacterID]()
    {
        HandleCharacterDeleteResponse(CharacterID, true, TEXT(""));
    }, 0.3f, false);
}

void UMMORPGCharacterSubsystem::ClearCharacterCache()
{
    CachedCharacters.Empty();
    SelectedCharacterID.Empty();
}

void UMMORPGCharacterSubsystem::SaveSelectedCharacter()
{
    GConfig->SetString(TEXT("MMORPG.Character"), TEXT("SelectedCharacterID"), *SelectedCharacterID, GGameIni);
    GConfig->Flush(false, GGameIni);
}

void UMMORPGCharacterSubsystem::LoadSelectedCharacter()
{
    GConfig->GetString(TEXT("MMORPG.Character"), TEXT("SelectedCharacterID"), SelectedCharacterID, GGameIni);
}

bool UMMORPGCharacterSubsystem::ValidateCharacterName(const FString& Name, FString& OutError) const
{
    // Check length
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

    // Check for valid characters (alphanumeric only)
    for (TCHAR Char : Name)
    {
        if (!FChar::IsAlnum(Char))
        {
            OutError = TEXT("Character name can only contain letters and numbers");
            return false;
        }
    }

    // Check for duplicate names in cache (client-side check only)
    for (const FCharacterInfo& Character : CachedCharacters)
    {
        if (Character.Name.Equals(Name, ESearchCase::IgnoreCase))
        {
            OutError = TEXT("You already have a character with this name");
            return false;
        }
    }

    return true;
}