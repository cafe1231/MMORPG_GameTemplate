// Copyright (c) 2024 MMORPG Template Project

#include "Blueprints/MMORPGProtoBPLibrary.h"
#include "MMORPGCore.h"
#include "Proto/MMORPGProtoHelper.h"
#include "Dom/JsonObject.h"
#include "Serialization/JsonSerializer.h"
#include "Serialization/JsonReader.h"
#include "Misc/Guid.h"
#include "Misc/SecureHash.h"

FString UMMORPGProtoBPLibrary::GetErrorMessage(int32 ErrorCode)
{
    return FMMORPGProtoHelper::GetErrorMessage(static_cast<mmorpg::ErrorCode>(ErrorCode));
}

bool UMMORPGProtoBPLibrary::IsSuccess(int32 ErrorCode)
{
    return ErrorCode == static_cast<int32>(mmorpg::ErrorCode::SUCCESS);
}

FMMORPGErrorInfo UMMORPGProtoBPLibrary::MakeErrorInfo(int32 Code, const FString& Message)
{
    FMMORPGErrorInfo Info;
    Info.Code = Code;
    Info.Message = Message;
    return Info;
}

FVector UMMORPGProtoBPLibrary::Vector3ToFVector(const FMMORPGVector3& Vector3)
{
    return Vector3.ToFVector();
}

FMMORPGVector3 UMMORPGProtoBPLibrary::FVectorToVector3(const FVector& Vector)
{
    return FMMORPGVector3::FromFVector(Vector);
}

FRotator UMMORPGProtoBPLibrary::RotationToFRotator(const FMMORPGRotation& Rotation)
{
    return Rotation.ToFRotator();
}

FMMORPGRotation UMMORPGProtoBPLibrary::FRotatorToRotation(const FRotator& Rotator)
{
    return FMMORPGRotation::FromFRotator(Rotator);
}

FTransform UMMORPGProtoBPLibrary::TransformToFTransform(const FMMORPGTransform& Transform)
{
    return Transform.ToFTransform();
}

FMMORPGTransform UMMORPGProtoBPLibrary::FTransformToTransform(const FTransform& Transform)
{
    return FMMORPGTransform::FromFTransform(Transform);
}

FDateTime UMMORPGProtoBPLibrary::TimestampToDateTime(int64 Timestamp)
{
    return FMMORPGProtoHelper::ProtoToDateTime(Timestamp);
}

int64 UMMORPGProtoBPLibrary::DateTimeToTimestamp(const FDateTime& DateTime)
{
    return FMMORPGProtoHelper::DateTimeToProto(DateTime);
}

int64 UMMORPGProtoBPLibrary::GetCurrentTimestamp()
{
    return FDateTime::UtcNow().ToUnixTimestamp();
}

bool UMMORPGProtoBPLibrary::ParseJsonString(const FString& JsonString, TMap<FString, FString>& OutFields)
{
    OutFields.Empty();
    
    TSharedPtr<FJsonObject> JsonObject;
    TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(JsonString);
    
    if (!FJsonSerializer::Deserialize(Reader, JsonObject) || !JsonObject.IsValid())
    {
        return false;
    }
    
    for (const auto& Field : JsonObject->Values)
    {
        FString Value;
        if (Field.Value->TryGetString(Value))
        {
            OutFields.Add(Field.Key, Value);
        }
        else if (Field.Value->TryGetNumber(Value))
        {
            OutFields.Add(Field.Key, Value);
        }
        else if (Field.Value->TryGetBool(Value))
        {
            OutFields.Add(Field.Key, Value);
        }
        else
        {
            // Convert complex types to string
            TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&Value);
            FJsonSerializer::Serialize(Field.Value.ToSharedRef(), Writer);
            OutFields.Add(Field.Key, Value);
        }
    }
    
    return true;
}

FString UMMORPGProtoBPLibrary::CreateJsonString(const TMap<FString, FString>& Fields)
{
    TSharedPtr<FJsonObject> JsonObject = MakeShareable(new FJsonObject);
    
    for (const auto& Field : Fields)
    {
        JsonObject->SetStringField(Field.Key, Field.Value);
    }
    
    FString OutputString;
    TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&OutputString);
    FJsonSerializer::Serialize(JsonObject.ToSharedRef(), Writer);
    
    return OutputString;
}

bool UMMORPGProtoBPLibrary::ValidateUsername(const FString& Username, FString& OutError)
{
    if (Username.IsEmpty())
    {
        OutError = TEXT("Username cannot be empty");
        return false;
    }
    
    if (Username.Len() < 3)
    {
        OutError = TEXT("Username must be at least 3 characters long");
        return false;
    }
    
    if (Username.Len() > 20)
    {
        OutError = TEXT("Username cannot be longer than 20 characters");
        return false;
    }
    
    // Check for valid characters (alphanumeric and underscore)
    FRegexPattern Pattern(TEXT("^[a-zA-Z0-9_]+$"));
    FRegexMatcher Matcher(Pattern, Username);
    
    if (!Matcher.FindNext())
    {
        OutError = TEXT("Username can only contain letters, numbers, and underscores");
        return false;
    }
    
    OutError = TEXT("");
    return true;
}

bool UMMORPGProtoBPLibrary::ValidateCharacterName(const FString& CharacterName, FString& OutError)
{
    if (CharacterName.IsEmpty())
    {
        OutError = TEXT("Character name cannot be empty");
        return false;
    }
    
    if (CharacterName.Len() < 3)
    {
        OutError = TEXT("Character name must be at least 3 characters long");
        return false;
    }
    
    if (CharacterName.Len() > 16)
    {
        OutError = TEXT("Character name cannot be longer than 16 characters");
        return false;
    }
    
    // Check for valid characters (letters only, with single spaces allowed)
    FRegexPattern Pattern(TEXT("^[a-zA-Z]+( [a-zA-Z]+)*$"));
    FRegexMatcher Matcher(Pattern, CharacterName);
    
    if (!Matcher.FindNext())
    {
        OutError = TEXT("Character name can only contain letters and single spaces");
        return false;
    }
    
    // Check for inappropriate content (basic filter)
    TArray<FString> BlockedWords = {
        TEXT("admin"), TEXT("gm"), TEXT("moderator"), TEXT("dev"), TEXT("system")
    };
    
    FString LowerName = CharacterName.ToLower();
    for (const FString& BlockedWord : BlockedWords)
    {
        if (LowerName.Contains(BlockedWord))
        {
            OutError = TEXT("Character name contains restricted words");
            return false;
        }
    }
    
    OutError = TEXT("");
    return true;
}

FString UMMORPGProtoBPLibrary::GenerateUUID()
{
    return FGuid::NewGuid().ToString();
}

FString UMMORPGProtoBPLibrary::HashString(const FString& Input)
{
    return FMD5::HashAnsiString(*Input);
}

void UMMORPGProtoBPLibrary::LogProtoDebug(const FString& Message, const FString& Category)
{
    if (Category.Equals(TEXT("Proto"), ESearchCase::IgnoreCase))
    {
        MMORPG_LOG(Log, TEXT("%s"), *Message);
    }
    else
    {
        UE_LOG(LogTemp, Log, TEXT("[%s] %s"), *Category, *Message);
    }
}