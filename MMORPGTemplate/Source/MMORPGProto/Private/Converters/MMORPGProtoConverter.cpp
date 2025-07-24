#include "Converters/MMORPGProtoConverter.h"
#include "Dom/JsonValue.h"
#include "Serialization/JsonSerializer.h"
#include "Serialization/JsonWriter.h"
#include "Serialization/JsonReader.h"

FString UMMORPGProtoConverter::SerializeToJson(const FMMORPGProtoMessage& Message)
{
	TSharedPtr<FJsonObject> JsonObject = StructToJsonObject(Message);
	if (!JsonObject.IsValid())
	{
		return TEXT("");
	}

	FString OutputString;
	TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&OutputString);
	FJsonSerializer::Serialize(JsonObject.ToSharedRef(), Writer);

	return OutputString;
}

bool UMMORPGProtoConverter::DeserializeFromJson(const FString& JsonString, FMMORPGProtoMessage& OutMessage)
{
	TSharedPtr<FJsonObject> JsonObject;
	TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(JsonString);

	if (!FJsonSerializer::Deserialize(Reader, JsonObject) || !JsonObject.IsValid())
	{
		return false;
	}

	return JsonObjectToStruct(JsonObject, OutMessage);
}

FVector UMMORPGProtoConverter::ProtoVectorToFVector(const FMMORPGVector3& ProtoVector)
{
	return ProtoVector.ToFVector();
}

FMMORPGVector3 UMMORPGProtoConverter::FVectorToProtoVector(const FVector& Vector)
{
	return FMMORPGVector3::FromFVector(Vector);
}

FQuat UMMORPGProtoConverter::ProtoQuaternionToFQuat(const FMMORPGQuaternion& ProtoQuat)
{
	return ProtoQuat.ToFQuat();
}

FMMORPGQuaternion UMMORPGProtoConverter::FQuatToProtoQuaternion(const FQuat& Quat)
{
	return FMMORPGQuaternion::FromFQuat(Quat);
}

FTransform UMMORPGProtoConverter::ProtoTransformToFTransform(const FMMORPGTransform& ProtoTransform)
{
	return ProtoTransform.ToFTransform();
}

FMMORPGTransform UMMORPGProtoConverter::FTransformToProtoTransform(const FTransform& Transform)
{
	return FMMORPGTransform::FromFTransform(Transform);
}

TSharedPtr<FJsonObject> UMMORPGProtoConverter::StructToJsonObject(const FMMORPGProtoMessage& Message)
{
	TSharedPtr<FJsonObject> JsonObject = MakeShareable(new FJsonObject);

	// Serialize base message fields
	JsonObject->SetStringField(TEXT("Type"), Message.Type);
	JsonObject->SetStringField(TEXT("Version"), Message.Version);
	JsonObject->SetNumberField(TEXT("Timestamp"), Message.Timestamp);

	// Note: In a real implementation, we would use reflection to serialize
	// derived message types. For now, this handles the base fields only.
	// When we implement real protobuf, this will be replaced.

	return JsonObject;
}

bool UMMORPGProtoConverter::JsonObjectToStruct(const TSharedPtr<FJsonObject>& JsonObject, FMMORPGProtoMessage& OutMessage)
{
	if (!JsonObject.IsValid())
	{
		return false;
	}

	// Deserialize base message fields
	FString Type;
	if (JsonObject->TryGetStringField(TEXT("Type"), Type))
	{
		OutMessage.Type = Type;
	}

	FString Version;
	if (JsonObject->TryGetStringField(TEXT("Version"), Version))
	{
		OutMessage.Version = Version;
	}

	double Timestamp;
	if (JsonObject->TryGetNumberField(TEXT("Timestamp"), Timestamp))
	{
		OutMessage.Timestamp = static_cast<int64>(Timestamp);
	}

	// Note: In a real implementation, we would use reflection to deserialize
	// derived message types. For now, this handles the base fields only.
	// When we implement real protobuf, this will be replaced.

	return true;
}