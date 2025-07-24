# Client Module Architecture

## Overview

The MMORPG Template client is built using a modular C++ architecture in Unreal Engine 5.6. This design ensures clean separation of concerns, easier testing, and better code organization.

## Module Overview

### MMORPGCore
**Purpose**: Foundation layer providing core systems and interfaces used by all other modules.

**Key Components**:
- Error handling system
- Core type definitions (enums, structs)
- Game instance subsystem
- Service interfaces

**Dependencies**: None (base module)

### MMORPGProto
**Purpose**: Handles Protocol Buffer integration and message serialization.

**Key Components**:
- Proto to UE5 type converters
- Message builders and parsers
- Serialization helpers
- Blueprint-friendly wrappers

**Dependencies**: MMORPGCore

### MMORPGNetwork
**Purpose**: Manages all network communication with the backend.

**Key Components**:
- HTTP client wrapper
- WebSocket client implementation
- Connection management
- Request/response handling
- Network state management

**Dependencies**: MMORPGCore, MMORPGProto

### MMORPGUI
**Purpose**: Provides UI framework and developer tools.

**Key Components**:
- Developer console system
- Base widget classes
- HUD management
- Error notification UI
- Blueprint utilities

**Dependencies**: MMORPGCore

## Module Loading Order

1. **PreDefault Phase**:
   - MMORPGCore (loads first, no dependencies)
   - MMORPGProto (depends on Core)

2. **Default Phase**:
   - MMORPGNetwork (depends on Core and Proto)
   - MMORPGUI (depends on Core)
   - MMORPGTemplate (main game, depends on all)

## Module Structure

```
Source/
├── MMORPGCore/
│   ├── Public/
│   │   ├── MMORPGCore.h
│   │   ├── CoreTypes.h
│   │   ├── Interfaces/
│   │   │   ├── IMMORPGService.h
│   │   │   └── IMMORPGErrorHandler.h
│   │   └── Subsystems/
│   │       ├── MMORPGGameInstanceSubsystem.h
│   │       └── MMORPGErrorSubsystem.h
│   └── Private/
│       └── [Implementation files]
│
├── MMORPGProto/
│   ├── Public/
│   │   ├── MMORPGProto.h
│   │   ├── Proto/
│   │   │   └── Generated/
│   │   └── Converters/
│   │       └── MMORPGProtoConverters.h
│   └── Private/
│       └── [Implementation files]
│
├── MMORPGNetwork/
│   ├── Public/
│   │   ├── MMORPGNetwork.h
│   │   ├── HTTP/
│   │   │   └── MMORPGHttpClient.h
│   │   ├── WebSocket/
│   │   │   └── MMORPGWebSocketClient.h
│   │   └── NetworkManager/
│   │       └── MMORPGNetworkManager.h
│   └── Private/
│       └── [Implementation files]
│
└── MMORPGUI/
    ├── Public/
    │   ├── MMORPGUI.h
    │   ├── Console/
    │   │   ├── MMORPGConsoleManager.h
    │   │   └── MMORPGConsoleWidget.h
    │   └── Widgets/
    │       └── MMORPGBaseWidget.h
    └── Private/
        └── [Implementation files]
```

## Module Communication

### Service Registration
```cpp
// In MMORPGCore
class IMMORPGService : public UInterface
{
    GENERATED_BODY()
};

class IMMORPGService
{
    GENERATED_IINTERFACE_BODY()
    
    virtual void Initialize() = 0;
    virtual void Shutdown() = 0;
};

// In other modules
class UMMORPGNetworkManager : public UObject, public IMMORPGService
{
    // Implementation
};
```

### Error Propagation
```cpp
// Error flows from any module through Core's error system
NetworkModule → ErrorSubsystem → UI Module → User Notification
```

### Data Flow
```
User Input → UI Module → Network Module → Proto Module → Backend
Backend → Proto Module → Network Module → Game Logic → UI Update
```

## Best Practices

### 1. Module Independence
Each module should only depend on modules loaded before it in the loading phase order.

### 2. Interface Segregation
Define interfaces in MMORPGCore that other modules implement:
```cpp
// Good: Interface in Core
class IMMORPGNetworkService : public UInterface { };

// Bad: Concrete class dependency
class UMMORPGNetworkManager; // Direct reference
```

### 3. Blueprint Exposure
All public functionality should be Blueprint-accessible:
```cpp
UCLASS(BlueprintType)
class MMORPGCORE_API UMMORPGErrorHandler : public UObject
{
    GENERATED_BODY()
    
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Error")
    void HandleError(const FMMORPGError& Error);
};
```

### 4. Error Handling
All modules report errors through MMORPGCore's error system:
```cpp
// In any module
FMMORPGError Error;
Error.Code = 1001;
Error.Message = TEXT("Connection failed");
Error.Category = EMMORPGErrorCategory::Network;

ErrorSubsystem->ReportError(Error);
```

### 5. Logging
Each module has its own log category:
```cpp
// In module header
DECLARE_LOG_CATEGORY_EXTERN(LogMMORPGNetwork, Log, All);

// In module cpp
DEFINE_LOG_CATEGORY(LogMMORPGNetwork);

// Usage
UE_LOG(LogMMORPGNetwork, Warning, TEXT("Connection lost"));
```

## Module Configuration

### Build.cs Dependencies
```csharp
// MMORPGNetwork.Build.cs
PublicDependencyModuleNames.AddRange(new string[] {
    "Core",
    "CoreUObject",
    "Engine",
    "MMORPGCore",      // Our base module
    "MMORPGProto",     // For message handling
    "HTTP",            // UE5 HTTP module
    "WebSockets",      // UE5 WebSocket module
    "Json",            // JSON parsing
    "JsonUtilities"    // JSON helpers
});
```

### Module Initialization
```cpp
void FMMORPGNetworkModule::StartupModule()
{
    // Register with game instance subsystem
    if (UGameInstance* GameInstance = GetGameInstance())
    {
        if (auto* Subsystem = GameInstance->GetSubsystem<UMMORPGGameInstanceSubsystem>())
        {
            Subsystem->RegisterService(NetworkManager);
        }
    }
}
```

## Testing Strategy

### Unit Tests
Each module should have its own test suite:
```
Source/
├── MMORPGCoreTests/
├── MMORPGProtoTests/
├── MMORPGNetworkTests/
└── MMORPGUITests/
```

### Integration Tests
Test module interactions:
- Core ↔ Network communication
- Proto ↔ Network serialization
- UI ↔ Core error handling

### Blueprint Tests
Ensure all exposed functions work correctly from Blueprint.

## Future Considerations

### Additional Modules
As the project grows, consider adding:
- **MMORPGCombat**: Combat system module
- **MMORPGInventory**: Inventory management
- **MMORPGQuest**: Quest system
- **MMORPGAudio**: Audio management

### Module Splitting
If a module grows too large, split it:
- MMORPGNetwork → MMORPGHttp + MMORPGWebSocket
- MMORPGUI → MMORPGWidgets + MMORPGConsole

### Performance
- Keep module boundaries clean for better build times
- Use forward declarations where possible
- Minimize inter-module dependencies