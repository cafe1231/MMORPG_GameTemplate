# Testing Client-Server Connection

## Quick Start

### 1. Start Backend Services

```bash
cd mmorpg-backend
docker-compose up -d
go run cmd/gateway/main.go
```

### 2. Test from Command Line

```bash
# Test basic connection
curl http://localhost:8090/api/v1/test

# Test echo endpoint
curl -X POST http://localhost:8090/api/v1/echo \
  -H "Content-Type: application/json" \
  -d '{"message":"Hello from client"}'

# Check health status
curl http://localhost:8090/health
```

### 3. Test from Unreal Engine

#### Option A: Using Blueprint Nodes
1. Create a new Blueprint Actor
2. Add nodes from **MMORPG|Network** category:
   - **Connect to Server** (localhost, 8090)
   - **Test Connection**
   - **Echo Test**
   - **Get Health Status**

#### Option B: Using Test Actor
1. In Content Browser, create Blueprint from `AMMORPGConnectionTest`
2. Place in level
3. Set properties:
   - Server Host: localhost
   - Server Port: 8090
   - Auto Run On Begin Play: true
4. Play in Editor - tests run automatically

#### Option C: Using Console Commands
```
mmorpg.status
mmorpg.connect localhost 8090
```

## Network Manager API

### Blueprint Functions
- `ConnectToServer(Host, Port)` - Connect to backend
- `DisconnectFromServer()` - Disconnect
- `IsConnected()` - Check connection status
- `TestAPI(OnComplete)` - Test basic API
- `EchoTest(Message, OnComplete)` - Echo test
- `SendGetRequest(Endpoint, OnComplete)` - GET request
- `SendPostRequest(Endpoint, JsonData, OnComplete)` - POST request
- `SetAuthToken(Token)` - Set auth token

### C++ API
```cpp
// Get network manager
auto NetworkManager = FMMORPGCoreModule::Get().GetNetworkManager();

// Connect
NetworkManager->Connect("localhost", 8090);

// Send requests
NetworkManager->SendGetRequest("/api/v1/test", 
    [](bool bSuccess, const FString& Response) {
        // Handle response
    });
```

## Troubleshooting

### Connection Failed
1. Check gateway service is running: `ps aux | grep gateway`
2. Check port 8090 is available: `netstat -an | grep 8090`
3. Check Docker services: `docker-compose ps`
4. Check firewall settings

### CORS Issues
Gateway has CORS enabled by default for development.

### Logs
- Backend: Check console output
- Unreal: Check Output Log (LogMMORPG, LogMMORPGNetwork categories)