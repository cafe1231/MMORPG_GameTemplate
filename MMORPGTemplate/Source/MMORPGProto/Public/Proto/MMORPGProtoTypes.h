#pragma once

#include "CoreMinimal.h"
#include "MMORPGProtoTypes.generated.h"

/**
 * Base message class for all protocol buffer messages
 */
USTRUCT(BlueprintType)
struct MMORPGPROTO_API FMMORPGProtoMessage
{
	GENERATED_BODY()

	UPROPERTY(BlueprintReadWrite, Category = "Proto")
	FString Type;

	UPROPERTY(BlueprintReadWrite, Category = "Proto")
	FString Version = TEXT("1.0");

	UPROPERTY(BlueprintReadWrite, Category = "Proto")
	int64 Timestamp = 0;

	FMMORPGProtoMessage()
	{
		Timestamp = FDateTime::Now().ToUnixTimestamp();
	}

	virtual ~FMMORPGProtoMessage() = default;
};

/**
 * Vector3 representation for protocol buffer compatibility
 */
USTRUCT(BlueprintType)
struct MMORPGPROTO_API FMMORPGVector3
{
	GENERATED_BODY()

	UPROPERTY(BlueprintReadWrite, Category = "Proto")
	float X = 0.0f;

	UPROPERTY(BlueprintReadWrite, Category = "Proto")
	float Y = 0.0f;

	UPROPERTY(BlueprintReadWrite, Category = "Proto")
	float Z = 0.0f;

	FMMORPGVector3() = default;

	FMMORPGVector3(float InX, float InY, float InZ)
		: X(InX), Y(InY), Z(InZ)
	{
	}

	FMMORPGVector3(const FVector& Vector)
		: X(Vector.X), Y(Vector.Y), Z(Vector.Z)
	{
	}

	FVector ToFVector() const
	{
		return FVector(X, Y, Z);
	}

	static FMMORPGVector3 FromFVector(const FVector& V)
	{
		return FMMORPGVector3(V.X, V.Y, V.Z);
	}
};

/**
 * Quaternion representation for protocol buffer compatibility
 */
USTRUCT(BlueprintType)
struct MMORPGPROTO_API FMMORPGQuaternion
{
	GENERATED_BODY()

	UPROPERTY(BlueprintReadWrite, Category = "Proto")
	float X = 0.0f;

	UPROPERTY(BlueprintReadWrite, Category = "Proto")
	float Y = 0.0f;

	UPROPERTY(BlueprintReadWrite, Category = "Proto")
	float Z = 0.0f;

	UPROPERTY(BlueprintReadWrite, Category = "Proto")
	float W = 1.0f;

	FMMORPGQuaternion() = default;

	FMMORPGQuaternion(float InX, float InY, float InZ, float InW)
		: X(InX), Y(InY), Z(InZ), W(InW)
	{
	}

	FMMORPGQuaternion(const FQuat& Quat)
		: X(Quat.X), Y(Quat.Y), Z(Quat.Z), W(Quat.W)
	{
	}

	FQuat ToFQuat() const
	{
		return FQuat(X, Y, Z, W);
	}

	static FMMORPGQuaternion FromFQuat(const FQuat& Q)
	{
		return FMMORPGQuaternion(Q.X, Q.Y, Q.Z, Q.W);
	}
};

/**
 * Transform representation for protocol buffer compatibility
 */
USTRUCT(BlueprintType)
struct MMORPGPROTO_API FMMORPGTransform
{
	GENERATED_BODY()

	UPROPERTY(BlueprintReadWrite, Category = "Proto")
	FMMORPGVector3 Position;

	UPROPERTY(BlueprintReadWrite, Category = "Proto")
	FMMORPGQuaternion Rotation;

	UPROPERTY(BlueprintReadWrite, Category = "Proto")
	FMMORPGVector3 Scale = FMMORPGVector3(1.0f, 1.0f, 1.0f);

	FMMORPGTransform() = default;

	FMMORPGTransform(const FTransform& Transform)
		: Position(Transform.GetLocation())
		, Rotation(Transform.GetRotation())
		, Scale(Transform.GetScale3D())
	{
	}

	FTransform ToFTransform() const
	{
		return FTransform(Rotation.ToFQuat(), Position.ToFVector(), Scale.ToFVector());
	}

	static FMMORPGTransform FromFTransform(const FTransform& T)
	{
		return FMMORPGTransform(T);
	}
};