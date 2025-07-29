# Phase 2 - Design - Real-time Networking Architecture

## Architecture Philosophy

### Core Principles
1. **Real-time First** - Every design decision optimizes for low latency
2. **Graceful Degradation** - System remains functional under poor network conditions
3. **State Consistency** - Server authoritative with client prediction
4. **Scalable by Design** - No architectural limits on concurrent connections
5. **Developer Friendly** - Complex networking made simple through abstractions

### Design Goals
- **Reliability**: Automatic recovery from any network disruption
- **Performance**: Sub-50ms latency for critical game events
- **Efficiency**: Minimal bandwidth usage through compression and delta updates
- **Extensibility**: Easy to add new message types and handlers
- **Observability**: Built-in debugging and monitoring capabilities

## Component Design

### Gateway WebSocket Upgrade

#### Architecture
```
┌─────────────────────────────────────────────────────────┐
│                   Gateway Service                        │
├─────────────────────────────────────────────────────────┤
│  HTTP Handler           │    WebSocket Handler          │
│  ├─ Health Check        │    ├─ Upgrade Logic          │
│  ├─ Metrics             │    ├─ JWT Validation         │
│  └─ Admin API           │    ├─ Protocol Negotiation   │
│                         │    ├─ Message Router         │
│                         │    └─ Connection Manager     │
└─────────────────────────┴────────────────────────────────┘
```

#### Implementation Details
```go
// internal/gateway/websocket/upgrader.go
type Upgrader struct {
    jwtValidator JWTValidator
    connManager  ConnectionManager
    rateLimiter  RateLimiter
    metrics      MetricsCollector
}

func (u *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request) error {
    // Extract JWT from Authorization header or query param
    token := u.extractToken(r)
    
    // Validate JWT before upgrade
    claims, err := u.jwtValidator.Validate(token)
    if err != nil {
        return ErrUnauthorized
    }
    
    // Check rate limits
    if !u.rateLimiter.Allow(claims.UserID) {
        return ErrRateLimited
    }
    
    // Perform WebSocket upgrade
    conn, err := websocket.Upgrade(w, r, nil)
    if err != nil {
        return err
    }
    
    // Create managed connection
    managedConn := u.connManager.Register(conn, claims)
    
    // Start connection handlers
    go managedConn.ReadPump()
    go managedConn.WritePump()
    go managedConn.HeartbeatPump()
    
    return nil
}
```

#### Connection Lifecycle
```
1. HTTP Request with Upgrade header
2. JWT validation from cookie/header/query
3. Rate limit check
4. WebSocket protocol negotiation
5. Connection registration
6. Start read/write/heartbeat pumps
7. Message routing begins
```

### Client Connection Manager

#### Architecture
```
┌─────────────────────────────────────────────────────────┐
│              Client Connection Manager                   │
├─────────────────────────────────────────────────────────┤
│  Connection Pool        │    State Manager              │
│  ├─ Active Connections  │    ├─ Connection State        │
│  ├─ Reconnect Queue     │    ├─ Message Queue           │
│  └─ Connection Stats    │    └─ Pending Operations      │
├─────────────────────────┼────────────────────────────────┤
│  Reconnection Logic     │    Health Monitor             │
│  ├─ Exponential Backoff │    ├─ Heartbeat Sender       │
│  ├─ Jitter Algorithm    │    ├─ Latency Tracker        │
│  └─ Max Retry Config    │    └─ Timeout Handler        │
└─────────────────────────┴────────────────────────────────┘
```

#### Unreal Engine Implementation
```cpp
// MMORPGConnectionManager.h
UCLASS()
class MMORPGCORE_API UMMORPGConnectionManager : public UGameInstanceSubsystem
{
    GENERATED_BODY()

public:
    // Connection management
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network")
    void Connect(const FString& ServerURL, const FString& AuthToken);
    
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network")
    void Disconnect();
    
    // Message handling
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network")
    void SendMessage(const FMMORPGMessage& Message);
    
    // Events
    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Network")
    FOnConnectionStateChanged OnConnectionStateChanged;
    
    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Network")
    FOnMessageReceived OnMessageReceived;

private:
    // WebSocket connection
    TSharedPtr<IWebSocket> WebSocket;
    
    // Reconnection logic
    FTimerHandle ReconnectTimer;
    int32 ReconnectAttempts;
    float GetReconnectDelay() const;
    
    // Message queue for offline/reconnecting
    TQueue<FMMORPGMessage> PendingMessages;
    
    // Connection health
    FTimerHandle HeartbeatTimer;
    FDateTime LastPongReceived;
};
```

#### Reconnection Strategy
```cpp
float UMMORPGConnectionManager::GetReconnectDelay() const
{
    // Exponential backoff with jitter
    float BaseDelay = 1.0f; // 1 second
    float MaxDelay = 60.0f; // 60 seconds
    
    // Calculate exponential delay
    float Delay = FMath::Min(
        BaseDelay * FMath::Pow(2.0f, ReconnectAttempts),
        MaxDelay
    );
    
    // Add jitter (±25%)
    float Jitter = FMath::RandRange(0.75f, 1.25f);
    
    return Delay * Jitter;
}
```

### Message Routing System

#### Message Format
```protobuf
syntax = "proto3";
package mmorpg.network;

// Base message envelope
message NetworkMessage {
    uint32 version = 1;         // Protocol version
    uint64 sequence = 2;        // Message sequence number
    int64 timestamp = 3;        // Unix timestamp (ms)
    string session_id = 4;      // Session identifier
    MessageType type = 5;       // Message type
    bytes payload = 6;          // Serialized message data
    bool requires_ack = 7;      // Acknowledgment required
    uint32 priority = 8;        // Message priority (0-10)
}

enum MessageType {
    MESSAGE_TYPE_UNKNOWN = 0;
    
    // System messages (1-99)
    MESSAGE_TYPE_HEARTBEAT = 1;
    MESSAGE_TYPE_PONG = 2;
    MESSAGE_TYPE_ACK = 3;
    MESSAGE_TYPE_ERROR = 4;
    
    // Player messages (100-199)
    MESSAGE_TYPE_PLAYER_MOVE = 100;
    MESSAGE_TYPE_PLAYER_ACTION = 101;
    MESSAGE_TYPE_PLAYER_STATE = 102;
    
    // World messages (200-299)
    MESSAGE_TYPE_WORLD_STATE = 200;
    MESSAGE_TYPE_ZONE_UPDATE = 201;
    MESSAGE_TYPE_TIME_SYNC = 202;
    
    // Social messages (300-399)
    MESSAGE_TYPE_PRESENCE_UPDATE = 300;
    MESSAGE_TYPE_FRIEND_STATUS = 301;
    
    // Reserved for Phase 3 (400+)
}
```

#### Server-side Router
```go
// internal/gateway/router/message_router.go
type MessageRouter struct {
    handlers map[MessageType]MessageHandler
    metrics  MetricsCollector
    logger   Logger
}

type MessageHandler func(ctx context.Context, 
    conn *Connection, 
    msg *NetworkMessage) error

func (r *MessageRouter) Route(ctx context.Context, 
    conn *Connection, 
    data []byte) error {
    // Parse message envelope
    var msg NetworkMessage
    if err := proto.Unmarshal(data, &msg); err != nil {
        return r.handleMalformedMessage(conn, err)
    }
    
    // Record metrics
    r.metrics.RecordMessage(msg.Type, len(data))
    
    // Get handler
    handler, exists := r.handlers[msg.Type]
    if !exists {
        return r.handleUnknownMessage(conn, &msg)
    }
    
    // Execute handler with timeout
    handlerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    if err := handler(handlerCtx, conn, &msg); err != nil {
        r.logger.Error("Handler error", 
            "type", msg.Type, 
            "error", err)
        return r.sendError(conn, msg.Sequence, err)
    }
    
    // Send acknowledgment if required
    if msg.RequiresAck {
        return r.sendAck(conn, msg.Sequence)
    }
    
    return nil
}
```

#### Client-side Dispatcher
```cpp
// MMORPGMessageDispatcher.cpp
void UMMORPGMessageDispatcher::DispatchMessage(
    const FMMORPGNetworkMessage& Message)
{
    // Find registered handlers
    auto* Handlers = MessageHandlers.Find(Message.Type);
    if (!Handlers || Handlers->Num() == 0)
    {
        UE_LOG(LogMMORPG, Warning, 
            TEXT("No handlers for message type: %d"), 
            (int32)Message.Type);
        return;
    }
    
    // Dispatch to all handlers
    for (const auto& Handler : *Handlers)
    {
        if (Handler.IsBound())
        {
            Handler.Execute(Message);
        }
    }
    
    // Update metrics
    NetworkMetrics->RecordMessageReceived(
        Message.Type, 
        Message.Payload.Num()
    );
}
```

### State Synchronization Framework

#### State Types and Updates
```protobuf
// Player state updates
message PlayerStateUpdate {
    string player_id = 1;
    
    // Position and movement
    Vector3 position = 2;
    Quaternion rotation = 3;
    Vector3 velocity = 4;
    
    // State flags
    uint32 state_flags = 5;  // Bitfield for states
    
    // Timestamp for interpolation
    int64 timestamp = 6;
    uint32 tick = 7;         // Server tick number
}

// Delta compression
message PlayerStateDelta {
    string player_id = 1;
    uint32 changed_fields = 2;  // Bitfield of changes
    
    // Only included if changed
    optional Vector3 position = 3;
    optional Quaternion rotation = 4;
    optional Vector3 velocity = 5;
    optional uint32 state_flags = 6;
    
    int64 timestamp = 7;
    uint32 base_tick = 8;    // Base state tick
    uint32 delta_tick = 9;   // Current tick
}

// World state snapshot
message WorldStateSnapshot {
    uint32 tick = 1;
    int64 timestamp = 2;
    string zone_id = 3;
    
    repeated PlayerStateUpdate players = 4;
    repeated NPCStateUpdate npcs = 5;
    repeated ObjectStateUpdate objects = 6;
    
    EnvironmentState environment = 7;
}
```

#### State Manager Implementation
```go
// internal/state/manager.go
type StateManager struct {
    mu         sync.RWMutex
    states     map[string]*EntityState
    history    *StateHistory
    compressor *DeltaCompressor
}

func (sm *StateManager) UpdatePlayerState(
    playerID string, 
    update *PlayerStateUpdate) (*PlayerStateDelta, error) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    // Get or create state
    state, exists := sm.states[playerID]
    if !exists {
        state = NewEntityState(playerID)
        sm.states[playerID] = state
    }
    
    // Calculate delta
    delta := sm.compressor.CalculateDelta(
        state.LastUpdate, 
        update
    )
    
    // Validate update (anti-cheat)
    if err := sm.validateUpdate(state, update); err != nil {
        return nil, err
    }
    
    // Apply update
    state.ApplyUpdate(update)
    
    // Store in history
    sm.history.Record(playerID, update)
    
    return delta, nil
}

// Anti-cheat validation
func (sm *StateManager) validateUpdate(
    state *EntityState, 
    update *PlayerStateUpdate) error {
    // Check movement speed
    if state.LastUpdate != nil {
        distance := calculateDistance(
            state.LastUpdate.Position, 
            update.Position
        )
        timeDelta := float64(update.Timestamp - 
            state.LastUpdate.Timestamp) / 1000.0
        speed := distance / timeDelta
        
        if speed > MaxAllowedSpeed {
            return ErrSpeedHack
        }
    }
    
    // Check position bounds
    if !sm.isValidPosition(update.Position) {
        return ErrInvalidPosition
    }
    
    return nil
}
```

#### Client-side Prediction
```cpp
// MMORPGMovementComponent.cpp
void UMMORPGMovementComponent::TickComponent(
    float DeltaTime, 
    ELevelTick TickType, 
    FActorComponentTickFunction* ThisTickFunction)
{
    Super::TickComponent(DeltaTime, TickType, ThisTickFunction);
    
    if (IsLocallyControlled())
    {
        // Client-side prediction
        PredictMovement(DeltaTime);
        
        // Send update to server
        if (ShouldSendUpdate())
        {
            SendMovementUpdate();
        }
    }
    else
    {
        // Interpolate remote players
        InterpolateMovement(DeltaTime);
    }
    
    // Reconcile with server state
    if (HasServerCorrection())
    {
        ReconcileWithServer();
    }
}

void UMMORPGMovementComponent::ReconcileWithServer()
{
    // Get server state
    const FPlayerState& ServerState = GetLatestServerState();
    
    // Calculate error
    FVector PositionError = ServerState.Position - 
        GetActorLocation();
    
    // Apply correction if significant
    if (PositionError.Size() > ReconciliationThreshold)
    {
        // Smooth correction over time
        CorrectionAlpha = 0.0f;
        CorrectionStart = GetActorLocation();
        CorrectionTarget = ServerState.Position;
        bIsReconciling = true;
        
        // Replay unacknowledged inputs
        ReplayInputsSince(ServerState.Timestamp);
    }
}
```

## Network Protocol Design

### Message Format Specifications

#### Wire Format
```
┌─────────────┬─────────────┬─────────────┬─────────────┐
│   Magic     │   Version   │   Length    │   Payload   │
│  (2 bytes)  │  (2 bytes)  │  (4 bytes)  │ (variable)  │
│   0x4D4D    │    0x0001   │  uint32_be  │  protobuf   │
└─────────────┴─────────────┴─────────────┴─────────────┘

Magic: 0x4D4D ('MM' for MMORPG)
Version: Protocol version (currently 1)
Length: Big-endian uint32 of payload size
Payload: Protocol Buffer encoded message
```

#### Message Priorities
```go
const (
    PrioritySystem    = 10  // Heartbeat, auth
    PriorityCritical  = 9   // Player actions
    PriorityHigh      = 7   // State updates
    PriorityNormal    = 5   // Chat, social
    PriorityLow       = 3   // Analytics
    PriorityDeferred  = 1   // Non-critical
)
```

### Event Types and Payloads

#### System Events
```protobuf
// Connection established
message ConnectionEstablished {
    string session_id = 1;
    string server_version = 2;
    int64 server_time = 3;
    uint32 heartbeat_interval = 4;
    map<string, string> features = 5;  // Feature flags
}

// Authentication result
message AuthenticationResult {
    bool success = 1;
    string user_id = 2;
    string character_id = 3;
    string error_message = 4;
    repeated string permissions = 5;
}

// Error notification
message ErrorNotification {
    uint32 code = 1;
    string message = 2;
    string details = 3;
    bool fatal = 4;
    string request_id = 5;
}
```

#### Player Events
```protobuf
// Player movement
message PlayerMovement {
    Vector3 position = 1;
    Quaternion rotation = 2;
    Vector3 velocity = 3;
    uint32 movement_flags = 4;  // Walking, running, etc.
    uint32 client_tick = 5;
    int64 client_time = 6;
}

// Player action
message PlayerAction {
    string action_id = 1;
    string target_id = 2;
    map<string, string> parameters = 3;
    uint32 client_tick = 4;
}

// Player status change
message PlayerStatusChange {
    string player_id = 1;
    PlayerStatus old_status = 2;
    PlayerStatus new_status = 3;
    string reason = 4;
}
```

#### World Events
```protobuf
// Zone update
message ZoneUpdate {
    string zone_id = 1;
    repeated EntitySpawn spawns = 2;
    repeated string despawns = 3;
    repeated EntityUpdate updates = 4;
    EnvironmentState environment = 5;
}

// Time synchronization
message TimeSync {
    int64 server_time = 1;
    uint32 server_tick = 2;
    int64 client_time = 3;  // Filled by client
    uint32 latency_ms = 4;  // Calculated RTT
}

// Interest management update
message InterestUpdate {
    repeated string entered_view = 1;
    repeated string exited_view = 2;
    float view_distance = 3;
}
```

### Compression Strategies

#### Message Batching
```go
// internal/network/batcher.go
type MessageBatcher struct {
    messages  []NetworkMessage
    size      int
    maxSize   int
    maxDelay  time.Duration
    timer     *time.Timer
}

func (b *MessageBatcher) Add(msg NetworkMessage) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    msgSize := proto.Size(&msg)
    
    // Send immediately if too large
    if b.size + msgSize > b.maxSize {
        b.flush()
    }
    
    b.messages = append(b.messages, msg)
    b.size += msgSize
    
    // Start timer on first message
    if len(b.messages) == 1 {
        b.timer = time.AfterFunc(b.maxDelay, b.flush)
    }
    
    return nil
}
```

#### Delta Compression
```go
// internal/compression/delta.go
type DeltaCompressor struct {
    baseStates map[string]interface{}
    schemas    map[reflect.Type]*Schema
}

func (dc *DeltaCompressor) Compress(
    entityID string, 
    newState interface{}) ([]byte, error) {
    baseState, exists := dc.baseStates[entityID]
    if !exists {
        // First update, send full state
        dc.baseStates[entityID] = newState
        return proto.Marshal(newState)
    }
    
    // Calculate differences
    delta := dc.calculateDelta(baseState, newState)
    
    // Update base state
    dc.baseStates[entityID] = newState
    
    return proto.Marshal(delta)
}
```

#### Bandwidth Optimization
```cpp
// Bit packing for common values
struct PackedMovement {
    // Position (3 floats -> 3 fixed point)
    int16_t PosX : 16;  // -32768 to 32767 (world units)
    int16_t PosY : 16;
    int16_t PosZ : 16;
    
    // Rotation (4 floats -> 3 components)
    int8_t RotPitch : 8;  // -128 to 127 (degrees/2)
    int8_t RotYaw : 8;
    int8_t RotRoll : 8;
    
    // Velocity (3 floats -> magnitude + direction)
    uint8_t VelMagnitude : 8;  // 0-255 (units/s)
    uint8_t VelDirection : 8;  // 0-255 (degrees*256/360)
    
    // Flags
    uint8_t MovementFlags : 8;
};
```

## Client-Server Flow

### Connection Sequence Diagrams

#### Initial Connection
```
┌──────────┐         ┌──────────┐         ┌──────────┐
│  Client  │         │ Gateway  │         │   Auth   │
└────┬─────┘         └────┬─────┘         └────┬─────┘
     │                    │                    │
     │ 1. HTTP Upgrade    │                    │
     │───────────────────►│                    │
     │                    │                    │
     │                    │ 2. Validate JWT    │
     │                    │───────────────────►│
     │                    │                    │
     │                    │ 3. JWT Valid       │
     │                    │◄───────────────────│
     │                    │                    │
     │ 4. 101 Switching   │                    │
     │◄───────────────────│                    │
     │                    │                    │
     │ 5. WS Connected    │                    │
     │◄──────────────────►│                    │
     │                    │                    │
     │ 6. Handshake Msg   │                    │
     │───────────────────►│                    │
     │                    │                    │
     │ 7. Session Created │                    │
     │◄───────────────────│                    │
     │                    │                    │
```

#### Reconnection Flow
```
┌──────────┐         ┌──────────┐         ┌──────────┐
│  Client  │         │ Gateway  │         │ Session  │
└────┬─────┘         └────┬─────┘         └────┬─────┘
     │                    │                    │
     │ 1. Connection Lost │                    │
     ├─ ─ ─ ─ ─ ─ ─ ─ ─ ─┤                    │
     │                    │                    │
     │ 2. Wait (backoff)  │                    │
     │                    │                    │
     │ 3. Reconnect       │                    │
     │───────────────────►│                    │
     │                    │                    │
     │                    │ 4. Check Session   │
     │                    │───────────────────►│
     │                    │                    │
     │                    │ 5. Session Valid   │
     │                    │◄───────────────────│
     │                    │                    │
     │ 6. Resume Session  │                    │
     │◄───────────────────│                    │
     │                    │                    │
     │ 7. Replay Messages │                    │
     │◄───────────────────│                    │
     │                    │                    │
```

### Message Flow Diagrams

#### Real-time State Sync
```
┌──────────┐         ┌──────────┐         ┌──────────┐
│ Client A │         │  Server  │         │ Client B │
└────┬─────┘         └────┬─────┘         └────┬─────┘
     │                    │                    │
     │ 1. Move Input     │                    │
     │───────────────────►│                    │
     │                    │                    │
     │                    │ 2. Validate        │
     │                    │                    │
     │                    │ 3. Update State    │
     │                    │                    │
     │ 4. Ack + Correct  │                    │
     │◄───────────────────│                    │
     │                    │                    │
     │                    │ 5. Broadcast       │
     │                    │───────────────────►│
     │                    │                    │
     │ 6. Area Update    │                    │
     │◄───────────────────│                    │
     │                    │                    │
```

#### Presence Update Flow
```
┌──────────┐      ┌──────────┐      ┌──────────┐      ┌──────────┐
│  Client  │      │ Gateway  │      │ Presence │      │  Redis   │
└────┬─────┘      └────┬─────┘      └────┬─────┘      └────┬─────┘
     │                 │                 │                 │
     │ 1. Status Change│                 │                 │
     │────────────────►│                 │                 │
     │                 │                 │                 │
     │                 │ 2. Update       │                 │
     │                 │────────────────►│                 │
     │                 │                 │                 │
     │                 │                 │ 3. Store        │
     │                 │                 │────────────────►│
     │                 │                 │                 │
     │                 │                 │ 4. Publish      │
     │                 │                 │◄────────────────│
     │                 │                 │                 │
     │                 │ 5. Notify Subscribers             │
     │                 │◄─────────────────────────────────│
     │                 │                 │                 │
     │ 6. Friend Update│                 │                 │
     │◄────────────────│                 │                 │
     │                 │                 │                 │
```

### State Sync Patterns

#### Client-side Prediction Pattern
```cpp
class PredictionSystem {
    // Local state (predicted)
    Transform predictedTransform;
    
    // Server state (authoritative)
    Transform serverTransform;
    uint32 lastServerTick;
    
    // Input buffer for replay
    CircularBuffer<Input> inputBuffer;
    
    void OnInputGenerated(const Input& input) {
        // Apply input locally immediately
        predictedTransform = ApplyInput(
            predictedTransform, 
            input
        );
        
        // Store for replay
        inputBuffer.Add(input);
        
        // Send to server
        SendInputToServer(input);
    }
    
    void OnServerUpdate(const ServerState& state) {
        // Store server state
        serverTransform = state.transform;
        lastServerTick = state.tick;
        
        // Find input at server tick
        auto serverInput = inputBuffer.FindByTick(
            lastServerTick
        );
        
        // Replay from server state
        Transform replayed = serverTransform;
        for (auto& input : inputBuffer.GetAfter(serverInput)) {
            replayed = ApplyInput(replayed, input);
        }
        
        // Check prediction error
        float error = Distance(replayed, predictedTransform);
        if (error > threshold) {
            // Correct with interpolation
            StartCorrection(replayed);
        }
    }
};
```

#### Interest Management Pattern
```go
// Spatial partitioning for scalability
type SpatialGrid struct {
    cellSize float64
    cells    map[CellID]*Cell
    entities map[EntityID]*Entity
}

func (sg *SpatialGrid) GetNearbyEntities(
    pos Position, 
    radius float64) []Entity {
    // Calculate affected cells
    minCell := sg.positionToCell(Position{
        X: pos.X - radius,
        Y: pos.Y - radius,
        Z: pos.Z - radius,
    })
    maxCell := sg.positionToCell(Position{
        X: pos.X + radius,
        Y: pos.Y + radius,
        Z: pos.Z + radius,
    })
    
    // Collect entities from cells
    var nearby []Entity
    for x := minCell.X; x <= maxCell.X; x++ {
        for y := minCell.Y; y <= maxCell.Y; y++ {
            for z := minCell.Z; z <= maxCell.Z; z++ {
                cellID := CellID{x, y, z}
                if cell, exists := sg.cells[cellID]; exists {
                    // Check actual distance
                    for _, entity := range cell.entities {
                        if entity.Position.DistanceTo(pos) <= radius {
                            nearby = append(nearby, entity)
                        }
                    }
                }
            }
        }
    }
    
    return nearby
}
```

## Error Handling

### Disconnection Recovery

#### Client-side Recovery
```cpp
void UMMORPGConnectionManager::HandleDisconnection(
    const FString& Reason)
{
    UE_LOG(LogMMORPG, Warning, 
        TEXT("Disconnected: %s"), *Reason);
    
    // Update connection state
    SetConnectionState(EConnectionState::Disconnected);
    
    // Save current game state
    SaveLocalState();
    
    // Check if we should reconnect
    if (ShouldAttemptReconnection(Reason))
    {
        // Queue messages during reconnection
        bIsQueueingMessages = true;
        
        // Start reconnection timer
        float Delay = GetReconnectDelay();
        GetWorld()->GetTimerManager().SetTimer(
            ReconnectTimer,
            this,
            &UMMORPGConnectionManager::AttemptReconnection,
            Delay,
            false
        );
        
        // Notify UI
        OnConnectionStateChanged.Broadcast(
            EConnectionState::Reconnecting
        );
    }
    else
    {
        // Fatal disconnection
        ShowDisconnectionDialog(Reason);
    }
}

void UMMORPGConnectionManager::AttemptReconnection()
{
    ReconnectAttempts++;
    
    UE_LOG(LogMMORPG, Log, 
        TEXT("Reconnection attempt %d"), 
        ReconnectAttempts);
    
    // Create new WebSocket with resume token
    FString ResumeURL = FString::Printf(
        TEXT("%s?resume=%s"),
        *ServerURL,
        *ResumeToken
    );
    
    Connect(ResumeURL, AuthToken);
}
```

#### Server-side Recovery
```go
func (s *SessionManager) HandleReconnection(
    conn *Connection, 
    resumeToken string) error {
    // Validate resume token
    session, err := s.validateResumeToken(resumeToken)
    if err != nil {
        return ErrInvalidResumeToken
    }
    
    // Check if session is still valid
    if time.Since(session.LastSeen) > SessionTimeout {
        return ErrSessionExpired
    }
    
    // Restore connection
    session.Connection = conn
    session.LastSeen = time.Now()
    
    // Send missed messages
    missedMessages := s.messageQueue.GetSince(
        session.PlayerID, 
        session.LastAck
    )
    
    for _, msg := range missedMessages {
        if err := conn.Send(msg); err != nil {
            return err
        }
    }
    
    // Send state snapshot
    snapshot := s.stateManager.GetSnapshot(session.PlayerID)
    return conn.Send(snapshot)
}
```

### Message Delivery Guarantees

#### At-least-once Delivery
```go
type ReliableMessageSender struct {
    pending  map[uint64]*PendingMessage
    mu       sync.Mutex
    timeout  time.Duration
    maxRetry int
}

type PendingMessage struct {
    Message   *NetworkMessage
    Sent      time.Time
    Retries   int
    Timer     *time.Timer
}

func (r *ReliableMessageSender) Send(
    conn *Connection, 
    msg *NetworkMessage) error {
    if !msg.RequiresAck {
        // Best effort delivery
        return conn.Send(msg)
    }
    
    // Store for acknowledgment
    r.mu.Lock()
    pending := &PendingMessage{
        Message: msg,
        Sent:    time.Now(),
        Retries: 0,
    }
    
    // Set retry timer
    pending.Timer = time.AfterFunc(r.timeout, func() {
        r.handleTimeout(conn, msg.Sequence)
    })
    
    r.pending[msg.Sequence] = pending
    r.mu.Unlock()
    
    return conn.Send(msg)
}

func (r *ReliableMessageSender) handleAck(sequence uint64) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if pending, exists := r.pending[sequence]; exists {
        pending.Timer.Stop()
        delete(r.pending, sequence)
    }
}
```

#### Message Ordering
```cpp
class MessageSequencer {
private:
    uint64_t NextExpectedSequence = 1;
    std::map<uint64_t, FNetworkMessage> OutOfOrderMessages;
    
public:
    void ProcessMessage(const FNetworkMessage& Message) {
        if (Message.Sequence == NextExpectedSequence) {
            // Process in order
            DeliverMessage(Message);
            NextExpectedSequence++;
            
            // Check for queued messages
            ProcessQueuedMessages();
        }
        else if (Message.Sequence > NextExpectedSequence) {
            // Store out of order
            OutOfOrderMessages[Message.Sequence] = Message;
        }
        // Ignore old messages (Message.Sequence < NextExpected)
    }
    
private:
    void ProcessQueuedMessages() {
        while (OutOfOrderMessages.count(NextExpectedSequence) > 0) {
            DeliverMessage(OutOfOrderMessages[NextExpectedSequence]);
            OutOfOrderMessages.erase(NextExpectedSequence);
            NextExpectedSequence++;
        }
    }
};
```

### Timeout Handling

#### Connection Timeout
```go
type TimeoutManager struct {
    connections map[string]*TimedConnection
    mu          sync.RWMutex
    interval    time.Duration
}

type TimedConnection struct {
    Conn         *Connection
    LastActivity time.Time
    TimeoutTimer *time.Timer
}

func (tm *TimeoutManager) UpdateActivity(connID string) {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    
    if tc, exists := tm.connections[connID]; exists {
        tc.LastActivity = time.Now()
        
        // Reset timeout timer
        tc.TimeoutTimer.Reset(tm.interval)
    }
}

func (tm *TimeoutManager) handleTimeout(connID string) {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    
    tc, exists := tm.connections[connID]
    if !exists {
        return
    }
    
    // Check if truly timed out
    if time.Since(tc.LastActivity) >= tm.interval {
        // Send final ping
        if err := tc.Conn.Ping(); err != nil {
            // Connection dead, clean up
            tc.Conn.Close()
            delete(tm.connections, connID)
            
            // Notify session manager
            tm.notifyDisconnection(connID, "timeout")
        } else {
            // Got response, reset timer
            tc.TimeoutTimer.Reset(tm.interval)
        }
    }
}
```

## Integration Design

### With Phase 1 Auth

#### JWT Integration
```go
// WebSocket authentication using Phase 1 JWT
func (g *Gateway) authenticateWebSocket(
    r *http.Request) (*auth.Claims, error) {
    // Try Authorization header first
    token := r.Header.Get("Authorization")
    if token != "" {
        token = strings.TrimPrefix(token, "Bearer ")
    }
    
    // Fall back to query parameter
    if token == "" {
        token = r.URL.Query().Get("token")
    }
    
    // Fall back to cookie
    if token == "" {
        if cookie, err := r.Cookie("auth_token"); err == nil {
            token = cookie.Value
        }
    }
    
    if token == "" {
        return nil, ErrNoToken
    }
    
    // Validate using Phase 1 auth service
    return g.authService.ValidateToken(r.Context(), token)
}
```

#### Session Continuity
```go
// Maintain session across HTTP and WebSocket
type UnifiedSession struct {
    SessionID    string
    UserID       string
    CharacterID  string
    HTTPToken    string
    WSConnection *Connection
    LastActivity time.Time
    Data         map[string]interface{}
}

func (s *SessionService) UpgradeToWebSocket(
    httpSession *HTTPSession, 
    conn *Connection) (*UnifiedSession, error) {
    // Create unified session
    unified := &UnifiedSession{
        SessionID:    httpSession.ID,
        UserID:       httpSession.UserID,
        CharacterID:  httpSession.CharacterID,
        HTTPToken:    httpSession.Token,
        WSConnection: conn,
        LastActivity: time.Now(),
        Data:         httpSession.Data,
    }
    
    // Register WebSocket connection
    s.registerConnection(unified)
    
    // Notify other services
    s.publishSessionUpgrade(unified)
    
    return unified, nil
}
```

### Preparing for Phase 3

#### Message Type Reservations
```protobuf
enum MessageType {
    // ... Phase 2 types (1-399) ...
    
    // Reserved for Phase 3 gameplay
    MESSAGE_TYPE_COMBAT_ACTION = 400;
    MESSAGE_TYPE_INVENTORY_UPDATE = 401;
    MESSAGE_TYPE_QUEST_PROGRESS = 402;
    MESSAGE_TYPE_CHAT_MESSAGE = 403;
    MESSAGE_TYPE_TRADE_REQUEST = 404;
    MESSAGE_TYPE_PARTY_INVITE = 405;
    MESSAGE_TYPE_GUILD_EVENT = 406;
    
    // Reserved for future phases (500+)
}
```

#### Extension Points
```go
// Gameplay hook interface for Phase 3
type GameplayHooks interface {
    // Called before processing player input
    PreProcessInput(player *Player, input Input) error
    
    // Called after state update
    PostStateUpdate(player *Player, oldState, newState State)
    
    // Custom validation rules
    ValidateAction(player *Player, action Action) error
    
    // Interest management customization
    GetInterestRadius(player *Player) float64
    
    // State serialization hooks
    SerializeCustomState(player *Player) []byte
    DeserializeCustomState(player *Player, data []byte) error
}

// Register Phase 3 gameplay systems
func (s *Server) RegisterGameplay(hooks GameplayHooks) {
    s.gameplayHooks = hooks
}
```

#### Metrics Foundation
```go
// Metrics that Phase 3 will build upon
var (
    // Player metrics
    PlayersOnline = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "mmorpg_players_online",
            Help: "Current online players",
        },
        []string{"region", "zone"},
    )
    
    // Performance metrics
    StateUpdateLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "mmorpg_state_update_latency_seconds",
            Help: "State update processing time",
        },
        []string{"update_type"},
    )
    
    // Network metrics
    BandwidthUsage = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mmorpg_bandwidth_bytes_total",
            Help: "Total bandwidth usage",
        },
        []string{"direction", "message_type"},
    )
)
```

## Extension Points

### Custom Message Types

#### Registration System
```go
// Message type registry for custom messages
type MessageRegistry struct {
    types    map[MessageType]MessageDescriptor
    handlers map[MessageType]MessageHandler
    mu       sync.RWMutex
}

type MessageDescriptor struct {
    Type     MessageType
    Name     string
    Schema   proto.Message
    Priority uint32
    Compress bool
}

func (r *MessageRegistry) RegisterType(
    desc MessageDescriptor, 
    handler MessageHandler) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Validate type not already registered
    if _, exists := r.types[desc.Type]; exists {
        return ErrTypeAlreadyRegistered
    }
    
    // Validate type range
    if desc.Type < 1000 {
        return ErrReservedTypeRange
    }
    
    r.types[desc.Type] = desc
    r.handlers[desc.Type] = handler
    
    return nil
}
```

#### Client-side Registration
```cpp
// Blueprint-friendly message registration
UCLASS()
class MMORPGCORE_API UMMORPGMessageRegistry : public UObject
{
    GENERATED_BODY()
    
public:
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network")
    void RegisterCustomMessage(
        int32 MessageType,
        TSubclassOf<UMMORPGMessage> MessageClass,
        const FMMORPGMessageHandler& Handler
    );
    
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network")
    void SendCustomMessage(
        int32 MessageType,
        const TMap<FString, FString>& Payload
    );
    
protected:
    TMap<int32, FMessageRegistration> CustomMessages;
};
```

### Event Handlers

#### Dynamic Handler System
```cpp
// Event handler interface
UINTERFACE(BlueprintType)
class UMMORPGEventHandler : public UInterface
{
    GENERATED_BODY()
};

class IMMORPGEventHandler
{
    GENERATED_BODY()
    
public:
    UFUNCTION(BlueprintNativeEvent, Category = "MMORPG|Events")
    void OnNetworkEvent(const FMMORPGEvent& Event);
    
    UFUNCTION(BlueprintNativeEvent, Category = "MMORPG|Events")
    bool CanHandleEvent(const FMMORPGEvent& Event);
    
    UFUNCTION(BlueprintNativeEvent, Category = "MMORPG|Events")
    int32 GetHandlerPriority() const;
};

// Event bus for registration
UCLASS()
class MMORPGCORE_API UMMORPGEventBus : public UGameInstanceSubsystem
{
    GENERATED_BODY()
    
public:
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Events")
    void RegisterHandler(
        const FString& EventType,
        TScriptInterface<IMMORPGEventHandler> Handler
    );
    
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Events")
    void UnregisterHandler(
        const FString& EventType,
        TScriptInterface<IMMORPGEventHandler> Handler
    );
    
    void DispatchEvent(const FMMORPGEvent& Event);
    
private:
    TMap<FString, TArray<TScriptInterface<IMMORPGEventHandler>>> Handlers;
};
```

### State Providers

#### Custom State Interface
```go
// State provider for custom game systems
type StateProvider interface {
    // Get state type identifier
    GetStateType() string
    
    // Serialize current state
    GetState(playerID string) ([]byte, error)
    
    // Apply state update
    ApplyState(playerID string, state []byte) error
    
    // Calculate state delta
    GetDelta(playerID string, since time.Time) ([]byte, error)
    
    // Handle player connection
    OnPlayerConnected(playerID string) error
    OnPlayerDisconnected(playerID string) error
}

// State provider registry
type StateProviderRegistry struct {
    providers map[string]StateProvider
    mu        sync.RWMutex
}

func (r *StateProviderRegistry) Register(
    stateType string, 
    provider StateProvider) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if _, exists := r.providers[stateType]; exists {
        return ErrProviderExists
    }
    
    r.providers[stateType] = provider
    
    // Initialize provider
    return provider.OnInitialize()
}
```

#### Blueprint State Provider
```cpp
// Blueprint-compatible state provider
UCLASS(Abstract, Blueprintable)
class MMORPGCORE_API UMMORPGStateProvider : public UObject
{
    GENERATED_BODY()
    
public:
    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|State")
    FString GetStateType() const;
    
    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|State")
    FMMORPGStateData GetCurrentState() const;
    
    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|State")
    void ApplyStateUpdate(const FMMORPGStateData& NewState);
    
    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|State")
    bool ShouldSyncState() const;
    
    UFUNCTION(BlueprintImplementableEvent, Category = "MMORPG|State")
    float GetSyncPriority() const;
};
```

This architecture provides a robust foundation for real-time networking that can scale from development to production while maintaining low latency and high reliability. The modular design allows developers to extend and customize the system for their specific game requirements.