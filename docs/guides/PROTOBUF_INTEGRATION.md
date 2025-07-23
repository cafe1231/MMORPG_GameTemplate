# Protocol Buffer Integration Guide

## Overview

The MMORPG Template uses Protocol Buffers for efficient data serialization between the Unreal Engine client and Go backend. This guide explains how to work with protobuf in both environments.

## Architecture

```
mmorpg-backend/pkg/proto/          # Proto definitions
├── base.proto                     # Common types
├── auth.proto                     # Authentication messages
├── character.proto                # Character management
├── game.proto                     # Gameplay messages
├── chat.proto                     # Chat system
└── gen/                          # Generated code
    ├── go/                       # Go generated files
    └── cpp/                      # C++ generated files

MMORPGTemplate/Source/MMORPGTemplate/
├── Public/Proto/
│   ├── MMORPGProtoHelper.h       # Proto utilities
│   └── MMORPGProtoTypes.h        # Type definitions
└── Private/Proto/
    └── MMORPGProtoHelper.cpp     # Implementation
```

## Working with Protocol Buffers

### Blueprint Usage

#### 1. Error Handling
```blueprint
// Check if operation succeeded
Is Success (ErrorCode) → Branch

// Get human-readable error message
Get Error Message (ErrorCode) → Print String
```

#### 2. Type Conversions
```blueprint
// Convert between UE and Proto types
FVector → FVector to Vector3 → Send to Server
Receive from Server → Vector3 to FVector → Set Actor Location

// Transform conversion
Get Actor Transform → FTransform to Transform → Send Position Update
```

#### 3. JSON Operations
```blueprint
// Parse server response
HTTP Response → Parse JSON String → Get Field "user_id"

// Create request
Make Map → Add "username", "password" → Create JSON String → Send Login Request
```

### C++ Usage

#### 1. Including Headers
```cpp
#include "Proto/MMORPGProtoHelper.h"
#include "Proto/MMORPGProtoTypes.h"
```

#### 2. Serialization Example
```cpp
// Create a login request
mmorpg::LoginRequest Request;
Request.username = "player123";
Request.password = "hashed_password";
Request.device_id = "device_uuid";
Request.client_version = "1.0.0";

// Serialize to binary
TArray<uint8> BinaryData = FMMORPGProtoHelper::SerializeProto(Request);

// Or to JSON
TSharedPtr<FJsonObject> JsonData = FMMORPGProtoHelper::ProtoToJson(Request);
```

#### 3. Deserialization Example
```cpp
// From binary data
mmorpg::LoginResponse Response;
if (FMMORPGProtoHelper::DeserializeProto(ReceivedData, Response))
{
    if (FMMORPGProtoHelper::IsSuccess(Response.error_code))
    {
        FString Token = FString(Response.access_token.c_str());
        // Use token for authenticated requests
    }
}
```

#### 4. Type Conversion
```cpp
// Convert Unreal types to Proto
FVector PlayerPos = GetActorLocation();
mmorpg::Vector3 ProtoPos = FMMORPGProtoHelper::VectorToProto(PlayerPos);

// Convert Proto to Unreal types
FTransform WorldTransform = FMMORPGProtoHelper::ProtoToTransform(ProtoTransform);
SetActorTransform(WorldTransform);
```

## Common Message Types

### Authentication Flow
```
Client                          Server
  |                               |
  |-- LoginRequest -------------->|
  |                               |
  |<-- LoginResponse -------------|
  |    (token, user_id)           |
  |                               |
  |-- RefreshTokenRequest ------->|
  |                               |
  |<-- RefreshTokenResponse ------|
```

### Character Management
```
Client                          Server
  |                               |
  |-- CharacterListRequest ------>|
  |                               |
  |<-- CharacterListResponse -----|
  |    (characters[])             |
  |                               |
  |-- CharacterCreateRequest ---->|
  |                               |
  |<-- CharacterCreateResponse ---|
  |    (character)                |
```

### Game State Updates
```
Client                          Server
  |                               |
  |-- PlayerStateUpdate --------->|
  |    (position, rotation)       |
  |                               |
  |<-- WorldStateUpdate ----------|
  |    (nearby_players[])         |
  |                               |
  |-- ActionRequest ------------->|
  |                               |
  |<-- ActionResponse ------------|
```

## Best Practices

### 1. Message Validation
Always validate messages before processing:
```cpp
FString Error;
if (!FMMORPGProtoHelper::ValidateMessage(Message, Error))
{
    MMORPG_LOG(Error, TEXT("Invalid message: %s"), *Error);
    return;
}
```

### 2. Error Handling
Check error codes in responses:
```cpp
if (!FMMORPGProtoHelper::IsSuccess(Response.error_code))
{
    FString ErrorMsg = FMMORPGProtoHelper::GetErrorMessage(Response.error_code);
    // Handle error appropriately
}
```

### 3. Performance Tips
- Use binary serialization for game data
- Use JSON for debugging and configuration
- Pool frequently-used message objects
- Batch small messages when possible

### 4. Debugging
Enable proto logging for development:
```cpp
#if !UE_BUILD_SHIPPING
FMMORPGProtoHelper::LogProtoMessage(Message, TEXT("Sending"));
#endif
```

## Adding New Messages

### 1. Define in Proto File
```protobuf
// In mmorpg-backend/pkg/proto/game.proto
message CustomAction {
    string action_id = 1;
    map<string, string> parameters = 2;
    int64 timestamp = 3;
}
```

### 2. Generate Code
```bash
cd mmorpg-backend
make proto
```

### 3. Copy Generated C++ Files
```bash
cp pkg/proto/gen/cpp/* ../MMORPGTemplate/Source/MMORPGTemplate/Public/Proto/Generated/
```

### 4. Add Wrapper Types (Optional)
```cpp
// In MMORPGProtoTypes.h
USTRUCT(BlueprintType)
struct FMMORPGCustomAction
{
    GENERATED_BODY()
    
    UPROPERTY(BlueprintReadWrite)
    FString ActionId;
    
    UPROPERTY(BlueprintReadWrite)
    TMap<FString, FString> Parameters;
    
    UPROPERTY(BlueprintReadWrite)
    int64 Timestamp;
};
```

## Troubleshooting

### Common Issues

1. **Serialization Fails**
   - Check message is properly initialized
   - Verify all required fields are set
   - Ensure proto versions match

2. **Type Conversion Errors**
   - Verify coordinate system conventions
   - Check for unit conversions (cm vs m)
   - Ensure proper type casting

3. **Network Issues**
   - Verify Content-Type headers
   - Check for proper error handling
   - Monitor message sizes

### Debug Commands
```
// Console commands for testing
mmorpg.proto.test - Run proto tests
mmorpg.proto.log - Toggle proto logging
mmorpg.proto.stats - Show serialization stats
```

## Resources

- [Protocol Buffers Documentation](https://developers.google.com/protocol-buffers)
- [Proto3 Language Guide](https://developers.google.com/protocol-buffers/docs/proto3)
- [C++ Generated Code Guide](https://developers.google.com/protocol-buffers/docs/reference/cpp-generated)