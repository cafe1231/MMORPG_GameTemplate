# Phase 2A: Core Real-time Infrastructure

## Overview

Phase 2A establishes the fundamental WebSocket connectivity and basic messaging infrastructure needed for real-time gameplay. This phase focuses on creating a stable, simple foundation that can be extended in Phase 2B.

**Duration**: 3-4 weeks
**Prerequisites**: Phase 1 (Authentication) complete
**Deliverable**: Basic real-time connectivity with presence tracking

---

## Technical Scope

### 1. WebSocket Foundation

#### Backend Components
```go
// Gateway WebSocket Handler
type WebSocketHandler struct {
    upgrader     websocket.Upgrader
    connections  *ConnectionManager
    auth         AuthValidator
    eventRouter  *EventRouter
}

// Core connection management
type Connection struct {
    ID           string
    UserID       string
    Socket       *websocket.Conn
    LastPing     time.Time
    SendQueue    chan Message
}
```

#### Frontend Components
```cpp
// Unreal WebSocket Manager
UCLASS()
class MMORPGNETWORK_API UMMORPGWebSocketManager : public UGameInstanceSubsystem
{
    GENERATED_BODY()
    
public:
    // Connection management
    UFUNCTION(BlueprintCallable)
    void Connect(const FString& URL, const FString& AuthToken);
    
    UFUNCTION(BlueprintCallable)
    void Disconnect();
    
    // Connection state
    UPROPERTY(BlueprintReadOnly)
    EWebSocketState ConnectionState;
    
    // Events
    UPROPERTY(BlueprintAssignable)
    FOnWebSocketConnected OnConnected;
    
    UPROPERTY(BlueprintAssignable)
    FOnWebSocketDisconnected OnDisconnected;
};
```

### 2. Core Messaging System

#### Message Protocol
```json
{
    "jsonrpc": "2.0",
    "method": "player.updatePresence",
    "params": {
        "status": "online",
        "location": "lobby"
    },
    "id": 1
}
```

#### Message Types
- **System Messages**: Connection status, server announcements
- **Player Messages**: Presence updates, basic player events
- **Error Messages**: Connection errors, validation failures

### 3. Basic Presence System

#### Features
- Online/offline status tracking
- Last seen timestamps
- Simple location tracking (lobby, in-game)
- Presence change notifications

#### Data Model
```sql
-- Redis keys for presence
presence:online:<user_id> = {
    "status": "online",
    "location": "lobby",
    "last_seen": "2024-01-20T10:30:00Z",
    "session_id": "ws_123456"
}

-- Online users set
presence:online:users = ["user1", "user2", "user3"]
```

### 4. Infrastructure Components

#### Redis Integration
- Session storage with TTL
- Presence data caching
- Fast user lookups

#### NATS Basic Setup
- Simple event publishing
- Service-to-service communication
- Presence update broadcasting

#### Monitoring
- Connection count metrics
- Message throughput tracking
- Basic error rates

---

## Implementation Tasks

### Week 1: Backend Foundation
1. **WebSocket Upgrade Handler**
   - HTTP to WebSocket upgrade
   - JWT validation
   - Connection registration

2. **Connection Manager**
   - Connection pool
   - Heartbeat mechanism
   - Graceful disconnect

3. **Basic Message Router**
   - Message parsing
   - Method routing
   - Error responses

### Week 2: Frontend Foundation
1. **WebSocket Manager Subsystem**
   - Connection lifecycle
   - Auto-reconnect (simple)
   - Event delegates

2. **Message Handler System**
   - Message parsing
   - Event dispatching
   - Error handling

3. **Connection UI Widget**
   - Status indicator
   - Reconnect button
   - Error messages

### Week 3: Presence System
1. **Backend Presence Service**
   - Status tracking
   - Redis integration
   - NATS broadcasting

2. **Frontend Presence Display**
   - Online users list
   - Status indicators
   - Update handling

3. **Integration Testing**
   - End-to-end flow
   - Multiple clients
   - Disconnect scenarios

### Week 4: Polish and Documentation
1. **Error Handling**
   - Graceful failures
   - User feedback
   - Retry logic

2. **Performance Testing**
   - Load testing
   - Latency measurements
   - Resource usage

3. **Documentation**
   - API reference
   - Integration guide
   - Troubleshooting

---

## Technical Decisions

### Why JSON-RPC 2.0?
- Simple, well-defined protocol
- Easy debugging and logging
- Good library support
- Extensible for Phase 2B

### Why Redis for Presence?
- Fast read/write operations
- TTL support for automatic cleanup
- Set operations for user lists
- Pub/sub for updates

### Why Simple Reconnect?
- Exponential backoff added in 2B
- Focus on core functionality first
- Easier testing and debugging
- Reduces initial complexity

---

## Success Criteria

### Functional Requirements
✅ WebSocket connections establish and authenticate
✅ Messages route correctly between client and server
✅ Presence updates broadcast to connected clients
✅ Basic reconnection works after disconnect
✅ Connection status visible in UI

### Performance Requirements
✅ Connection establishment < 2 seconds
✅ Message latency < 100ms (same region)
✅ Support 100+ concurrent connections
✅ Presence updates within 2 seconds
✅ Memory usage stable over time

### Quality Requirements
✅ No message loss during normal operation
✅ Graceful handling of network errors
✅ Clear error messages for users
✅ Clean disconnection on exit
✅ No resource leaks

---

## API Examples

### Connect to Server
```cpp
// Unreal Blueprint
WebSocketManager->Connect("ws://localhost:8090/ws", AuthToken);
```

### Send Message
```cpp
// C++ Example
FMMORPGMessage Message;
Message.Method = "player.updateStatus";
Message.Params.Add("status", "away");
WebSocketManager->SendMessage(Message);
```

### Handle Presence Update
```cpp
// Event handler
void OnPresenceUpdate(const FString& UserID, const FPresenceData& Data)
{
    // Update UI
    PresenceWidget->UpdateUserStatus(UserID, Data.Status);
}
```

---

## Testing Strategy

### Unit Tests
- Message parsing and serialization
- Connection state management
- Presence data operations

### Integration Tests
- Full connection flow
- Message round trips
- Multi-client scenarios

### Load Tests
- 100 concurrent connections
- 1000 messages/second
- Sustained 1-hour sessions

---

## Migration to Phase 2B

Phase 2A provides these foundations for 2B:
1. **Stable WebSocket connections** - Ready for advanced features
2. **Message routing system** - Extensible for new message types
3. **Basic presence tracking** - Foundation for complex state
4. **Monitoring infrastructure** - Metrics for optimization

---

## Risks and Mitigations

### Risk 1: Connection Stability
**Mitigation**: Simple heartbeat mechanism, basic reconnect

### Risk 2: Message Ordering
**Mitigation**: Sequential processing, no complex state yet

### Risk 3: Scalability Limits
**Mitigation**: Design for 100 users now, optimize in 2B

---

## Deliverables Checklist

- [ ] WebSocket server implementation
- [ ] Client WebSocket manager
- [ ] Basic message protocol
- [ ] Simple presence system
- [ ] Connection status UI
- [ ] API documentation
- [ ] Integration tests
- [ ] Performance benchmarks

---

*Phase 2A focuses on simplicity and stability, providing a solid foundation for the advanced features in Phase 2B.*