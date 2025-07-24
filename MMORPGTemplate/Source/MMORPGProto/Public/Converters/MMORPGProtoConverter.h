#pragma once

#include "CoreMinimal.h"
#include "Kismet/BlueprintFunctionLibrary.h"
#include "Dom/JsonObject.h"
#include "Proto/MMORPGProtoTypes.h"
#include "MMORPGProtoConverter.generated.h"

/**
 * Protocol buffer converter utility functions
 * Currently uses JSON serialization as a temporary solution
 * Will be replaced with actual protobuf later
 */
UCLASS()
class MMORPGPROTO_API UMMORPGProtoConverter : public UBlueprintFunctionLibrary
{
	GENERATED_BODY()

public:
	// JSON Serialization (temporary until real protobuf)
	
	/**
	 * Serialize a message to JSON string
	 * @param Message The message to serialize
	 * @return JSON string representation
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Proto", meta = (DisplayName = "Serialize To JSON"))
	static FString SerializeToJson(const FMMORPGProtoMessage& Message);

	/**
	 * Deserialize a JSON string to a message
	 * @param JsonString The JSON string to deserialize
	 * @param OutMessage The resulting message
	 * @return True if deserialization was successful
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Proto", meta = (DisplayName = "Deserialize From JSON"))
	static bool DeserializeFromJson(const FString& JsonString, FMMORPGProtoMessage& OutMessage);

	// Type converters

	/**
	 * Convert Proto vector to Unreal vector
	 * @param ProtoVector The proto vector to convert
	 * @return Unreal FVector
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Proto Vector to FVector"))
	static FVector ProtoVectorToFVector(const FMMORPGVector3& ProtoVector);

	/**
	 * Convert Unreal vector to Proto vector
	 * @param Vector The Unreal vector to convert
	 * @return Proto vector
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "FVector to Proto Vector"))
	static FMMORPGVector3 FVectorToProtoVector(const FVector& Vector);

	/**
	 * Convert Proto quaternion to Unreal quaternion
	 * @param ProtoQuat The proto quaternion to convert
	 * @return Unreal FQuat
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Proto Quaternion to FQuat"))
	static FQuat ProtoQuaternionToFQuat(const FMMORPGQuaternion& ProtoQuat);

	/**
	 * Convert Unreal quaternion to Proto quaternion
	 * @param Quat The Unreal quaternion to convert
	 * @return Proto quaternion
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "FQuat to Proto Quaternion"))
	static FMMORPGQuaternion FQuatToProtoQuaternion(const FQuat& Quat);

	/**
	 * Convert Proto transform to Unreal transform
	 * @param ProtoTransform The proto transform to convert
	 * @return Unreal FTransform
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "Proto Transform to FTransform"))
	static FTransform ProtoTransformToFTransform(const FMMORPGTransform& ProtoTransform);

	/**
	 * Convert Unreal transform to Proto transform
	 * @param Transform The Unreal transform to convert
	 * @return Proto transform
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Proto", meta = (DisplayName = "FTransform to Proto Transform"))
	static FMMORPGTransform FTransformToProtoTransform(const FTransform& Transform);

	// Helper functions for JSON object manipulation
	static TSharedPtr<FJsonObject> StructToJsonObject(const FMMORPGProtoMessage& Message);
	static bool JsonObjectToStruct(const TSharedPtr<FJsonObject>& JsonObject, FMMORPGProtoMessage& OutMessage);
};