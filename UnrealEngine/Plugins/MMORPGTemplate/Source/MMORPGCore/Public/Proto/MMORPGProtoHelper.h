// Copyright (c) 2024 MMORPG Template Project

#pragma once

#include "CoreMinimal.h"
#include "Dom/JsonObject.h"
#include "Proto/MMORPGProtoTypes.h"

/**
 * Helper class for Protocol Buffer serialization/deserialization in Unreal Engine
 * Provides utilities to convert between Protobuf messages and Unreal types
 */
class MMORPGCORE_API FMMORPGProtoHelper
{
public:
    // JSON Conversion
    static TSharedPtr<FJsonObject> ProtoToJson(const google::protobuf::Message& Message);
    static bool JsonToProto(const TSharedPtr<FJsonObject>& JsonObject, google::protobuf::Message& OutMessage);
    
    // Binary Serialization
    static TArray<uint8> SerializeProto(const google::protobuf::Message& Message);
    static bool DeserializeProto(const TArray<uint8>& Data, google::protobuf::Message& OutMessage);
    
    // String Serialization
    static FString ProtoToString(const google::protobuf::Message& Message);
    static bool StringToProto(const FString& Data, google::protobuf::Message& OutMessage);
    
    // Type Conversions
    static FVector ProtoToVector(const mmorpg::Vector3& ProtoVector);
    static mmorpg::Vector3 VectorToProto(const FVector& Vector);
    
    static FRotator ProtoToRotator(const mmorpg::Rotation& ProtoRotation);
    static mmorpg::Rotation RotatorToProto(const FRotator& Rotator);
    
    static FTransform ProtoToTransform(const mmorpg::Transform& ProtoTransform);
    static mmorpg::Transform TransformToProto(const FTransform& Transform);
    
    static FDateTime ProtoToDateTime(int64 Timestamp);
    static int64 DateTimeToProto(const FDateTime& DateTime);
    
    // Error Code Handling
    static FString GetErrorMessage(mmorpg::ErrorCode Code);
    static bool IsSuccess(mmorpg::ErrorCode Code) { return Code == mmorpg::ErrorCode::SUCCESS; }
    
    // Validation
    static bool ValidateMessage(const google::protobuf::Message& Message, FString& OutError);
    
    // Debug
    static void LogProtoMessage(const google::protobuf::Message& Message, const FString& Prefix = TEXT(""));
    
private:
    static void ConvertFieldToJson(const google::protobuf::Message& Message, 
                                  const google::protobuf::FieldDescriptor* Field,
                                  TSharedPtr<FJsonObject>& OutJsonObject);
    
    static void ConvertJsonToField(const TSharedPtr<FJsonObject>& JsonObject,
                                  google::protobuf::Message& Message,
                                  const google::protobuf::FieldDescriptor* Field);
};