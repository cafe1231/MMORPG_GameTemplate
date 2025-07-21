// Copyright (c) 2024 MMORPG Template Project

#pragma once

#include "CoreMinimal.h"

// Forward declarations for Protocol Buffer types
// These will be replaced with actual generated headers when protobuf is integrated

namespace google {
namespace protobuf {
    class Message;
    class FieldDescriptor;
    class Reflection;
    class Descriptor;
}
}

namespace mmorpg {
    // Basic types (from base.proto)
    enum ErrorCode {
        SUCCESS = 0,
        UNKNOWN_ERROR = 1,
        INVALID_REQUEST = 2,
        UNAUTHORIZED = 3,
        FORBIDDEN = 4,
        NOT_FOUND = 5,
        ALREADY_EXISTS = 6,
        RATE_LIMITED = 7,
        SERVER_ERROR = 8,
        DATABASE_ERROR = 9,
        NETWORK_ERROR = 10
    };
    
    struct Vector3 {
        float x;
        float y;
        float z;
    };
    
    struct Rotation {
        float pitch;
        float yaw;
        float roll;
    };
    
    struct Transform {
        Vector3 position;
        Rotation rotation;
        Vector3 scale;
    };
    
    struct Color {
        float r;
        float g;
        float b;
        float a;
    };
    
    // Auth types (from auth.proto)
    struct LoginRequest {
        std::string username;
        std::string password;
        std::string device_id;
        std::string client_version;
    };
    
    struct LoginResponse {
        ErrorCode error_code;
        std::string error_message;
        std::string access_token;
        std::string refresh_token;
        int64_t expires_at;
        std::string user_id;
    };
    
    // Character types (from character.proto)
    struct Character {
        std::string id;
        std::string name;
        std::string class_id;
        int32_t level;
        int64_t experience;
        Transform world_transform;
        std::string zone_id;
    };
    
    struct CharacterListResponse {
        ErrorCode error_code;
        std::string error_message;
        std::vector<Character> characters;
        int32_t max_characters;
    };
    
    // Game types (from game.proto)
    struct PlayerState {
        std::string player_id;
        std::string character_id;
        Transform transform;
        float health;
        float max_health;
        float mana;
        float max_mana;
        int32_t movement_state;
        int32_t combat_state;
    };
    
    struct WorldState {
        std::string zone_id;
        int64_t server_time;
        std::vector<PlayerState> nearby_players;
        int32_t player_count;
    };
}

// Unreal Engine wrapper types for Protocol Buffers
USTRUCT(BlueprintType)
struct FMMORPGErrorInfo
{
    GENERATED_BODY()
    
    UPROPERTY(BlueprintReadOnly)
    int32 Code;
    
    UPROPERTY(BlueprintReadOnly)
    FString Message;
    
    FMMORPGErrorInfo()
        : Code(0)
    {}
};

USTRUCT(BlueprintType)
struct FMMORPGVector3
{
    GENERATED_BODY()
    
    UPROPERTY(BlueprintReadWrite)
    float X;
    
    UPROPERTY(BlueprintReadWrite)
    float Y;
    
    UPROPERTY(BlueprintReadWrite)
    float Z;
    
    FMMORPGVector3()
        : X(0.0f), Y(0.0f), Z(0.0f)
    {}
    
    FMMORPGVector3(float InX, float InY, float InZ)
        : X(InX), Y(InY), Z(InZ)
    {}
    
    FVector ToFVector() const { return FVector(X, Y, Z); }
    static FMMORPGVector3 FromFVector(const FVector& V) { return FMMORPGVector3(V.X, V.Y, V.Z); }
};

USTRUCT(BlueprintType)
struct FMMORPGRotation
{
    GENERATED_BODY()
    
    UPROPERTY(BlueprintReadWrite)
    float Pitch;
    
    UPROPERTY(BlueprintReadWrite)
    float Yaw;
    
    UPROPERTY(BlueprintReadWrite)
    float Roll;
    
    FMMORPGRotation()
        : Pitch(0.0f), Yaw(0.0f), Roll(0.0f)
    {}
    
    FRotator ToFRotator() const { return FRotator(Pitch, Yaw, Roll); }
    static FMMORPGRotation FromFRotator(const FRotator& R) 
    { 
        FMMORPGRotation Rot;
        Rot.Pitch = R.Pitch;
        Rot.Yaw = R.Yaw;
        Rot.Roll = R.Roll;
        return Rot;
    }
};

USTRUCT(BlueprintType)
struct FMMORPGTransform
{
    GENERATED_BODY()
    
    UPROPERTY(BlueprintReadWrite)
    FMMORPGVector3 Position;
    
    UPROPERTY(BlueprintReadWrite)
    FMMORPGRotation Rotation;
    
    UPROPERTY(BlueprintReadWrite)
    FMMORPGVector3 Scale;
    
    FMMORPGTransform()
    {
        Scale = FMMORPGVector3(1.0f, 1.0f, 1.0f);
    }
    
    FTransform ToFTransform() const 
    { 
        return FTransform(Rotation.ToFRotator(), Position.ToFVector(), Scale.ToFVector()); 
    }
    
    static FMMORPGTransform FromFTransform(const FTransform& T)
    {
        FMMORPGTransform Transform;
        Transform.Position = FMMORPGVector3::FromFVector(T.GetLocation());
        Transform.Rotation = FMMORPGRotation::FromFRotator(T.GetRotation().Rotator());
        Transform.Scale = FMMORPGVector3::FromFVector(T.GetScale3D());
        return Transform;
    }
};