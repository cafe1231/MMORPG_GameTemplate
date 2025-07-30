#include "Types/FCharacterTypes.h"
#include "Dom/JsonObject.h"
#include "Serialization/JsonSerializer.h"
#include "Serialization/JsonReader.h"
#include "Serialization/JsonWriter.h"

// FCharacterAppearance implementation
FString FCharacterAppearance::ToJSON() const
{
    TSharedPtr<FJsonObject> JsonObject = MakeShareable(new FJsonObject);
    
    JsonObject->SetStringField(TEXT("gender"), CharacterGenderToString(Gender));
    JsonObject->SetNumberField(TEXT("face_id"), FaceID);
    JsonObject->SetNumberField(TEXT("hair_id"), HairID);
    JsonObject->SetStringField(TEXT("skin_color"), SkinColor);
    JsonObject->SetStringField(TEXT("hair_color"), HairColor);
    JsonObject->SetStringField(TEXT("eye_color"), EyeColor);
    JsonObject->SetNumberField(TEXT("height"), Height);
    JsonObject->SetNumberField(TEXT("build"), Build);

    FString OutputString;
    TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&OutputString);
    FJsonSerializer::Serialize(JsonObject.ToSharedRef(), Writer);
    
    return OutputString;
}

bool FCharacterAppearance::ParseFromJSON(const TSharedPtr<FJsonObject>& JsonObject)
{
    if (!JsonObject.IsValid())
        return false;

    FString GenderStr;
    if (JsonObject->TryGetStringField(TEXT("gender"), GenderStr))
        Gender = StringToCharacterGender(GenderStr);

    JsonObject->TryGetNumberField(TEXT("face_id"), FaceID);
    JsonObject->TryGetNumberField(TEXT("hair_id"), HairID);
    JsonObject->TryGetStringField(TEXT("skin_color"), SkinColor);
    JsonObject->TryGetStringField(TEXT("hair_color"), HairColor);
    JsonObject->TryGetStringField(TEXT("eye_color"), EyeColor);
    
    double HeightValue = 1.0;
    if (JsonObject->TryGetNumberField(TEXT("height"), HeightValue))
        Height = HeightValue;

    double BuildValue = 1.0;
    if (JsonObject->TryGetNumberField(TEXT("build"), BuildValue))
        Build = BuildValue;

    return true;
}

// FCharacterInfo implementation
bool FCharacterInfo::ParseFromJSON(const TSharedPtr<FJsonObject>& JsonObject)
{
    if (!JsonObject.IsValid())
        return false;

    JsonObject->TryGetStringField(TEXT("id"), ID);
    JsonObject->TryGetStringField(TEXT("user_id"), UserID);
    JsonObject->TryGetStringField(TEXT("name"), Name);
    JsonObject->TryGetStringField(TEXT("class"), Class);
    
    FString RaceStr;
    if (JsonObject->TryGetStringField(TEXT("race"), RaceStr))
        Race = StringToCharacterRace(RaceStr);

    JsonObject->TryGetNumberField(TEXT("level"), Level);
    JsonObject->TryGetNumberField(TEXT("experience_points"), ExperiencePoints);

    // Parse appearance
    const TSharedPtr<FJsonObject>* AppearanceObj;
    if (JsonObject->TryGetObjectField(TEXT("appearance"), AppearanceObj))
    {
        Appearance.ParseFromJSON(*AppearanceObj);
    }

    // Parse stats
    const TSharedPtr<FJsonObject>* StatsObj;
    if (JsonObject->TryGetObjectField(TEXT("stats"), StatsObj))
    {
        (*StatsObj)->TryGetNumberField(TEXT("health"), Stats.Health);
        (*StatsObj)->TryGetNumberField(TEXT("max_health"), Stats.MaxHealth);
        (*StatsObj)->TryGetNumberField(TEXT("mana"), Stats.Mana);
        (*StatsObj)->TryGetNumberField(TEXT("max_mana"), Stats.MaxMana);
        (*StatsObj)->TryGetNumberField(TEXT("strength"), Stats.Strength);
        (*StatsObj)->TryGetNumberField(TEXT("intelligence"), Stats.Intelligence);
        (*StatsObj)->TryGetNumberField(TEXT("agility"), Stats.Agility);
        (*StatsObj)->TryGetNumberField(TEXT("stamina"), Stats.Stamina);
    }

    // Parse position
    const TSharedPtr<FJsonObject>* PositionObj;
    if (JsonObject->TryGetObjectField(TEXT("position"), PositionObj))
    {
        (*PositionObj)->TryGetStringField(TEXT("world"), Position.World);
        
        double X = 0, Y = 0, Z = 0;
        (*PositionObj)->TryGetNumberField(TEXT("x"), X);
        (*PositionObj)->TryGetNumberField(TEXT("y"), Y);
        (*PositionObj)->TryGetNumberField(TEXT("z"), Z);
        Position.Location = FVector(X, Y, Z);

        double Pitch = 0, Yaw = 0, Roll = 0;
        (*PositionObj)->TryGetNumberField(TEXT("pitch"), Pitch);
        (*PositionObj)->TryGetNumberField(TEXT("yaw"), Yaw);
        (*PositionObj)->TryGetNumberField(TEXT("roll"), Roll);
        Position.Rotation = FRotator(Pitch, Yaw, Roll);
    }

    // Parse timestamps
    FString CreatedAtStr;
    if (JsonObject->TryGetStringField(TEXT("created_at"), CreatedAtStr))
    {
        FDateTime::ParseIso8601(*CreatedAtStr, CreatedAt);
    }

    FString LastPlayedAtStr;
    if (JsonObject->TryGetStringField(TEXT("last_played_at"), LastPlayedAtStr))
    {
        FDateTime::ParseIso8601(*LastPlayedAtStr, LastPlayedAt);
    }

    JsonObject->TryGetBoolField(TEXT("is_deleted"), bIsDeleted);

    return true;
}

// FCharacterCreateRequest implementation
FString FCharacterCreateRequest::ToJSON() const
{
    TSharedPtr<FJsonObject> JsonObject = MakeShareable(new FJsonObject);
    
    JsonObject->SetStringField(TEXT("name"), Name);
    JsonObject->SetStringField(TEXT("class"), Class.ToLower());
    JsonObject->SetStringField(TEXT("race"), CharacterRaceToString(Race).ToLower());
    
    // Add appearance as nested object
    TSharedPtr<FJsonObject> AppearanceJson = MakeShareable(new FJsonObject);
    AppearanceJson->SetStringField(TEXT("gender"), CharacterGenderToString(Appearance.Gender).ToLower());
    AppearanceJson->SetNumberField(TEXT("face_id"), Appearance.FaceID);
    AppearanceJson->SetNumberField(TEXT("hair_id"), Appearance.HairID);
    AppearanceJson->SetStringField(TEXT("skin_color"), Appearance.SkinColor);
    AppearanceJson->SetStringField(TEXT("hair_color"), Appearance.HairColor);
    AppearanceJson->SetStringField(TEXT("eye_color"), Appearance.EyeColor);
    AppearanceJson->SetNumberField(TEXT("height"), Appearance.Height);
    AppearanceJson->SetNumberField(TEXT("build"), Appearance.Build);
    
    JsonObject->SetObjectField(TEXT("appearance"), AppearanceJson);

    FString OutputString;
    TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&OutputString);
    FJsonSerializer::Serialize(JsonObject.ToSharedRef(), Writer);
    
    return OutputString;
}

// FCharacterUpdateRequest implementation
FString FCharacterUpdateRequest::ToJSON() const
{
    TSharedPtr<FJsonObject> JsonObject = MakeShareable(new FJsonObject);
    
    if (!Name.IsEmpty())
        JsonObject->SetStringField(TEXT("name"), Name);
    
    // Add appearance as nested object
    TSharedPtr<FJsonObject> AppearanceJson = MakeShareable(new FJsonObject);
    AppearanceJson->SetStringField(TEXT("gender"), CharacterGenderToString(Appearance.Gender).ToLower());
    AppearanceJson->SetNumberField(TEXT("face_id"), Appearance.FaceID);
    AppearanceJson->SetNumberField(TEXT("hair_id"), Appearance.HairID);
    AppearanceJson->SetStringField(TEXT("skin_color"), Appearance.SkinColor);
    AppearanceJson->SetStringField(TEXT("hair_color"), Appearance.HairColor);
    AppearanceJson->SetStringField(TEXT("eye_color"), Appearance.EyeColor);
    AppearanceJson->SetNumberField(TEXT("height"), Appearance.Height);
    AppearanceJson->SetNumberField(TEXT("build"), Appearance.Build);
    
    JsonObject->SetObjectField(TEXT("appearance"), AppearanceJson);

    FString OutputString;
    TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&OutputString);
    FJsonSerializer::Serialize(JsonObject.ToSharedRef(), Writer);
    
    return OutputString;
}

// FCharacterListResponse implementation
bool FCharacterListResponse::ParseFromJSON(const FString& JsonString)
{
    TSharedPtr<FJsonObject> JsonObject;
    TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(JsonString);
    
    if (!FJsonSerializer::Deserialize(Reader, JsonObject) || !JsonObject.IsValid())
        return false;

    JsonObject->TryGetBoolField(TEXT("success"), bSuccess);
    
    // Check for error
    const TSharedPtr<FJsonObject>* ErrorObj;
    if (JsonObject->TryGetObjectField(TEXT("error"), ErrorObj))
    {
        (*ErrorObj)->TryGetStringField(TEXT("message"), ErrorMessage);
    }

    // Parse characters array
    const TArray<TSharedPtr<FJsonValue>>* CharactersArray;
    if (JsonObject->TryGetArrayField(TEXT("data"), CharactersArray))
    {
        for (const TSharedPtr<FJsonValue>& Value : *CharactersArray)
        {
            const TSharedPtr<FJsonObject>* CharObj;
            if (Value->TryGetObject(CharObj))
            {
                FCharacterInfo CharInfo;
                if (CharInfo.ParseFromJSON(*CharObj))
                {
                    Characters.Add(CharInfo);
                }
            }
        }
    }

    return true;
}

// FCharacterResponse implementation
bool FCharacterResponse::ParseFromJSON(const FString& JsonString)
{
    TSharedPtr<FJsonObject> JsonObject;
    TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(JsonString);
    
    if (!FJsonSerializer::Deserialize(Reader, JsonObject) || !JsonObject.IsValid())
        return false;

    JsonObject->TryGetBoolField(TEXT("success"), bSuccess);
    
    // Check for error
    const TSharedPtr<FJsonObject>* ErrorObj;
    if (JsonObject->TryGetObjectField(TEXT("error"), ErrorObj))
    {
        (*ErrorObj)->TryGetStringField(TEXT("message"), ErrorMessage);
    }

    // Parse character data
    const TSharedPtr<FJsonObject>* DataObj;
    if (JsonObject->TryGetObjectField(TEXT("data"), DataObj))
    {
        Character.ParseFromJSON(*DataObj);
    }

    return true;
}

// Helper function implementations
FString CharacterClassToString(ECharacterClass Class)
{
    switch (Class)
    {
        case ECharacterClass::Warrior: return TEXT("Warrior");
        case ECharacterClass::Mage: return TEXT("Mage");
        case ECharacterClass::Archer: return TEXT("Archer");
        case ECharacterClass::Rogue: return TEXT("Rogue");
        case ECharacterClass::Priest: return TEXT("Priest");
        case ECharacterClass::Paladin: return TEXT("Paladin");
        default: return TEXT("None");
    }
}

ECharacterClass StringToCharacterClass(const FString& ClassString)
{
    if (ClassString.Equals(TEXT("Warrior"), ESearchCase::IgnoreCase))
        return ECharacterClass::Warrior;
    else if (ClassString.Equals(TEXT("Mage"), ESearchCase::IgnoreCase))
        return ECharacterClass::Mage;
    else if (ClassString.Equals(TEXT("Archer"), ESearchCase::IgnoreCase))
        return ECharacterClass::Archer;
    else if (ClassString.Equals(TEXT("Rogue"), ESearchCase::IgnoreCase))
        return ECharacterClass::Rogue;
    else if (ClassString.Equals(TEXT("Priest"), ESearchCase::IgnoreCase))
        return ECharacterClass::Priest;
    else if (ClassString.Equals(TEXT("Paladin"), ESearchCase::IgnoreCase))
        return ECharacterClass::Paladin;
    else
        return ECharacterClass::None;
}

FString CharacterRaceToString(ECharacterRace Race)
{
    switch (Race)
    {
        case ECharacterRace::Human: return TEXT("Human");
        case ECharacterRace::Elf: return TEXT("Elf");
        case ECharacterRace::Dwarf: return TEXT("Dwarf");
        case ECharacterRace::Orc: return TEXT("Orc");
        case ECharacterRace::Undead: return TEXT("Undead");
        default: return TEXT("None");
    }
}

ECharacterRace StringToCharacterRace(const FString& RaceString)
{
    if (RaceString.Equals(TEXT("Human"), ESearchCase::IgnoreCase))
        return ECharacterRace::Human;
    else if (RaceString.Equals(TEXT("Elf"), ESearchCase::IgnoreCase))
        return ECharacterRace::Elf;
    else if (RaceString.Equals(TEXT("Dwarf"), ESearchCase::IgnoreCase))
        return ECharacterRace::Dwarf;
    else if (RaceString.Equals(TEXT("Orc"), ESearchCase::IgnoreCase))
        return ECharacterRace::Orc;
    else if (RaceString.Equals(TEXT("Undead"), ESearchCase::IgnoreCase))
        return ECharacterRace::Undead;
    else
        return ECharacterRace::None;
}

FString CharacterGenderToString(ECharacterGender Gender)
{
    switch (Gender)
    {
        case ECharacterGender::Male: return TEXT("Male");
        case ECharacterGender::Female: return TEXT("Female");
        case ECharacterGender::Other: return TEXT("Other");
        default: return TEXT("Male");
    }
}

ECharacterGender StringToCharacterGender(const FString& GenderString)
{
    if (GenderString.Equals(TEXT("Male"), ESearchCase::IgnoreCase))
        return ECharacterGender::Male;
    else if (GenderString.Equals(TEXT("Female"), ESearchCase::IgnoreCase))
        return ECharacterGender::Female;
    else if (GenderString.Equals(TEXT("Other"), ESearchCase::IgnoreCase))
        return ECharacterGender::Other;
    else
        return ECharacterGender::Male;
}