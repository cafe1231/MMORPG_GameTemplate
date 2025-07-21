# MMORPG Template Plugin for Unreal Engine 5.6

## Overview

This is the Unreal Engine plugin component of the MMORPG Template, providing a complete networking and gameplay framework for building massively multiplayer online games that can scale from 1 to 1M+ players.

## Features

- **WebSocket-based Networking**: Real-time bidirectional communication
- **Protocol Buffer Serialization**: Efficient binary message format
- **Authentication System**: JWT-based secure authentication
- **Character Management**: Complete character creation and selection
- **World Synchronization**: Interest management and spatial optimization
- **Modular Architecture**: Easy to extend and customize
- **Blueprint Support**: Full Blueprint exposure for all systems
- **Editor Tools**: Custom editor tools for testing and debugging

## Requirements

- Unreal Engine 5.6 or later
- Visual Studio 2022 (Windows) or Xcode (Mac)
- Protocol Buffers compiler (protoc)
- C++17 compatible compiler

## Installation

1. Copy the `MMORPGTemplate` folder to your project's `Plugins` directory
2. Regenerate project files
3. Compile your project
4. Enable the plugin in Edit > Plugins > Networking > MMORPG Template

## Quick Start

### 1. Configure Backend Connection

Edit `Config/DefaultMMORPG.ini`:

```ini
[/Script/MMORPGCore.MMORPGSettings]
DefaultServerHost=localhost
DefaultServerPort=8080
```

### 2. Basic Setup in Blueprint

1. Create a new GameInstance blueprint
2. Add the MMORPG Network Component
3. Set up authentication callbacks
4. Handle connection events

### 3. C++ Example

```cpp
#include "MMORPGCore.h"

// Get the network manager
auto NetworkManager = FMMORPGCoreModule::Get().GetNetworkManager();

// Connect to server
NetworkManager->Connect("localhost", 8080);

// Login
auto AuthManager = FMMORPGCoreModule::Get().GetAuthManager();
AuthManager->Login("user@example.com", "password");
```

## Module Structure

```
MMORPGTemplate/
├── Source/
│   ├── MMORPGCore/          # Runtime module
│   │   ├── Public/
│   │   │   ├── Network/     # Networking classes
│   │   │   ├── Authentication/  # Auth system
│   │   │   ├── Data/        # Data management
│   │   │   ├── Gameplay/    # Gameplay systems
│   │   │   ├── UI/          # UI components
│   │   │   └── Proto/       # Protocol Buffers
│   │   └── Private/         # Implementation
│   └── MMORPGEditor/        # Editor module
│       ├── Public/          # Editor tools API
│       └── Private/         # Editor implementation
├── Content/                 # Blueprint content
│   ├── Blueprints/         # Example blueprints
│   ├── UI/                 # UI widgets
│   └── Examples/           # Example content
└── Config/                 # Configuration files
```

## Core Components

### Network Manager
Handles all network communication with the backend:
- WebSocket connection management
- Message serialization/deserialization
- Connection state handling
- Automatic reconnection

### Authentication Manager
Manages user authentication and sessions:
- Login/Register functionality
- JWT token management
- Session persistence
- Auto-refresh tokens

### Data Manager
Handles game data and caching:
- Character data management
- Inventory synchronization
- Quest progress tracking
- Local data caching

### World Manager
Manages world state and synchronization:
- Player position updates
- Interest management
- Entity spawning/despawning
- Physics reconciliation

## Blueprint Nodes

### Network Nodes
- **Connect to Server**: Establish connection to backend
- **Disconnect**: Close connection cleanly
- **Send Message**: Send custom messages
- **Is Connected**: Check connection status

### Authentication Nodes
- **Login**: Authenticate with email/password
- **Register**: Create new account
- **Logout**: End current session
- **Get Current User**: Get logged-in user info

### Character Nodes
- **Get Character List**: Retrieve user's characters
- **Create Character**: Create new character
- **Select Character**: Choose character to play
- **Delete Character**: Remove character

### World Nodes
- **Join World**: Enter game world
- **Leave World**: Exit game world
- **Update Position**: Send position updates
- **Get Nearby Players**: Get players in range

## Configuration

### Network Settings
```ini
[/Script/MMORPGCore.MMORPGSettings]
ConnectionTimeout=30.0
ReconnectAttempts=3
HeartbeatInterval=30.0
MaxMessageSize=65536
```

### Performance Settings
```ini
NetworkTickRate=30
EnableCompression=true
CompressionThreshold=1024
EnableDeltaCompression=true
```

### Debug Settings
```ini
EnableDebugLogging=true
EnableNetworkLogging=true
ShowNetworkStats=true
```

## Editor Tools

### MMORPG Dashboard
Access via Window > MMORPG Tools > Dashboard
- Connection status
- Quick actions
- Documentation links
- System information

### Connection Test Tool
Test backend connectivity:
- Ping/latency test
- Authentication test
- Protocol version check
- Service health check

### Protocol Viewer
Inspect Protocol Buffer definitions:
- Message structure
- Field types
- Serialization test
- Size calculator

## Console Commands

### Development Commands (non-shipping builds)
- `mmorpg.status` - Show plugin status
- `mmorpg.connect <host> <port>` - Connect to server
- `mmorpg.disconnect` - Disconnect from server
- `mmorpg.stats` - Show network statistics

## Troubleshooting

### Connection Issues
1. Verify backend is running
2. Check firewall settings
3. Confirm correct host/port
4. Check logs for errors

### Authentication Failures
1. Verify credentials
2. Check JWT configuration
3. Ensure backend auth service is running
4. Check token expiration

### Performance Issues
1. Enable network statistics
2. Check message frequency
3. Verify interest management
4. Profile network usage

## Best Practices

1. **Message Batching**: Combine multiple updates into single messages
2. **Interest Management**: Only sync entities within view distance
3. **Delta Compression**: Send only changed values
4. **Object Pooling**: Reuse network objects
5. **Async Operations**: Use callbacks for network operations

## Support

- Documentation: https://docs.mmorpg-template.com
- Discord: https://discord.gg/mmorpg-template
- Issues: https://github.com/mmorpg-template/issues

## License

This plugin is part of the commercial MMORPG Template product. See LICENSE.md for details.