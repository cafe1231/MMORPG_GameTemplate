// Copyright (c) 2024 MMORPG Template Project

#pragma once

#include "CoreMinimal.h"
#include "Kismet/BlueprintFunctionLibrary.h"
#include "Proto/MMORPGProtoTypes.h"
#include "MMORPGProtoBPLibrary.generated.h"

/**
 * Blueprint function library for Protocol Buffer operations
 * Provides easy-to-use functions for Blueprint scripting
 */
UCLASS()
class MMORPGCORE_API UMMORPGProtoBPLibrary : public UBlueprintFunctionLibrary
{
    GENERATED_BODY()

public:
    // Error Handling
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Get Error Message"))
    static FString GetErrorMessage(int32 ErrorCode);
    
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Is Success"))
    static bool IsSuccess(int32 ErrorCode);
    
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Make Error Info"))
    static FMMORPGErrorInfo MakeErrorInfo(int32 Code, const FString& Message);
    
    // Type Conversions
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Vector3 to FVector"))
    static FVector Vector3ToFVector(const FMMORPGVector3& Vector3);
    
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "FVector to Vector3"))
    static FMMORPGVector3 FVectorToVector3(const FVector& Vector);
    
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Rotation to FRotator"))
    static FRotator RotationToFRotator(const FMMORPGRotation& Rotation);
    
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "FRotator to Rotation"))
    static FMMORPGRotation FRotatorToRotation(const FRotator& Rotator);
    
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Transform to FTransform"))
    static FTransform TransformToFTransform(const FMMORPGTransform& Transform);
    
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "FTransform to Transform"))
    static FMMORPGTransform FTransformToTransform(const FTransform& Transform);
    
    // Time Conversions
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Timestamp to DateTime"))
    static FDateTime TimestampToDateTime(int64 Timestamp);
    
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "DateTime to Timestamp"))
    static int64 DateTimeToTimestamp(const FDateTime& DateTime);
    
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Get Current Timestamp"))
    static int64 GetCurrentTimestamp();
    
    // JSON Operations
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Proto", meta = (DisplayName = "Parse JSON String"))
    static bool ParseJsonString(const FString& JsonString, TMap<FString, FString>& OutFields);
    
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Create JSON String"))
    static FString CreateJsonString(const TMap<FString, FString>& Fields);
    
    // Validation
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Validate Username"))
    static bool ValidateUsername(const FString& Username, FString& OutError);
    
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Validate Character Name"))
    static bool ValidateCharacterName(const FString& CharacterName, FString& OutError);
    
    // Utility
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Generate UUID"))
    static FString GenerateUUID();
    
    UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Hash String"))
    static FString HashString(const FString& Input);
    
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Proto", meta = (DisplayName = "Log Proto Debug"))
    static void LogProtoDebug(const FString& Message, const FString& Category = TEXT("Proto"));
};