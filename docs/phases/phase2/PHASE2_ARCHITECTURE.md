# 🏗️ Phase 2: Real-time Networking - Architecture Document

## 📋 Executive Summary

This document details the technical architecture for Phase 2's real-time networking implementation. Building on the hexagonal architecture established in Phase 1, we'll add WebSocket capabilities to the gateway service, implement real-time message routing, state synchronization, and presence tracking - all while maintaining clean separation of concerns and scalability.

**Key Architectural Decisions:**
- WebSocket upgrade in existing gateway service (no new services)
- JSON-RPC 2.0 over WebSocket with Protocol Buffer payloads
- Redis for session state and presence tracking
- NATS for inter-service real-time events
- Server-authoritative state with client prediction

---

## 🌐 1. WebSocket Service Architecture

### 1.1 Gateway Service WebSocket Upgrade

The gateway service will be enhanced to support WebSocket connections alongside HTTP:

```go
// internal/adapters/websocket/handler.go
package websocket

import (
    "context"
    "net/http"
    "time"
    
    "github.com/gorilla/websocket"
    "github.com/mmorpg-template/backend/internal/ports"
    "github.com/mmorpg-template/backend/pkg/logger"
)

type WebSocketHandler struct {
    upgrader        websocket.Upgrader
    sessionManager  ports.SessionManager
    messageRouter   ports.MessageRouter
    authService     ports.AuthService
    eventBus        ports.EventBus
    logger          logger.Logger
}

func NewWebSocketHandler(
    sessionManager ports.SessionManager,
    messageRouter ports.MessageRouter,
    authService ports.AuthService,
    eventBus ports.EventBus,
    logger logger.Logger,
) *WebSocketHandler {
    return &WebSocketHandler{
        upgrader: websocket.Upgrader{
            ReadBufferSize:  1024,
            WriteBufferSize: 1024,
            CheckOrigin: func(r *http.Request) bool {
                // TODO: Implement proper origin checking
                return true
            },
        },
        sessionManager: sessionManager,
        messageRouter:  messageRouter,
        authService:    authService,
        eventBus:       eventBus,
        logger:         logger,
    }
}

func (h *WebSocketHandler) HandleUpgrade(w http.ResponseWriter, r *http.Request) {
    // Extract JWT from query params or Authorization header
    token := extractToken(r)
    if token == "" {
        http.Error(w, "Missing authentication token", http.StatusUnauthorized)
        return
    }
    
    // Validate JWT
    claims, err := h.authService.ValidateToken(r.Context(), token)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }
    
    // Upgrade connection
    conn, err := h.upgrader.Upgrade(w, r, nil)
    if err != nil {
        h.logger.WithError(err).Error("Failed to upgrade connection")
        return
    }
    
    // Create player session
    session := &PlayerSession{
        ID:         generateSessionID(),
        UserID:     claims.UserID,
        Connection: conn,
        State:      SessionStateConnected,
        CreatedAt:  time.Now(),
        LastPing:   time.Now(),
    }
    
    // Register session
    if err := h.sessionManager.Register(session); err != nil {
        conn.Close()
        h.logger.WithError(err).Error("Failed to register session")
        return
    }
    
    // Start connection handler
    go h.handleConnection(session)
}
```

### 1.2 Connection Management Patterns

```go
// internal/domain/realtime/session.go
package realtime

import (
    "sync"
    "time"
    
    "github.com/gorilla/websocket"
)

type SessionState int

const (
    SessionStateConnected SessionState = iota
    SessionStateAuthenticated
    SessionStateInWorld
    SessionStateDisconnecting
    SessionStateDisconnected
)

type PlayerSession struct {
    ID           string
    UserID       string
    CharacterID  string // Set when character selected
    Connection   *websocket.Conn
    State        SessionState
    
    // Connection health
    LastPing     time.Time
    LastPong     time.Time
    PingInterval time.Duration
    PongTimeout  time.Duration
    
    // Message handling
    IncomingQueue  chan []byte
    OutgoingQueue  chan []byte
    
    // State management
    mu            sync.RWMutex
    attributes    map[string]interface{}
    subscriptions map[string]bool
    
    // Metrics
    MessagesReceived uint64
    MessagesSent     uint64
    BytesReceived    uint64
    BytesSent        uint64
    
    CreatedAt    time.Time
    LastActivity time.Time
}

// Connection handler with heartbeat
func (h *WebSocketHandler) handleConnection(session *PlayerSession) {
    defer func() {
        h.sessionManager.Unregister(session.ID)
        session.Connection.Close()
    }()
    
    // Start heartbeat
    pingTicker := time.NewTicker(30 * time.Second)
    defer pingTicker.Stop()
    
    // Start message processors
    go h.readMessages(session)
    go h.writeMessages(session)
    
    // Main connection loop
    for {
        select {
        case <-pingTicker.C:
            if err := h.sendPing(session); err != nil {
                h.logger.WithError(err).Error("Ping failed")
                return
            }
            
        case <-session.Done():
            return
        }
    }
}
```

### 1.3 Load Balancing Strategies

```go
// internal/adapters/websocket/loadbalancer.go
package websocket

import (
    "hash/fnv"
    "sync"
    "sync/atomic"
)

type LoadBalancer interface {
    SelectGateway(userID string) string
    RegisterGateway(id string, capacity int)
    UnregisterGateway(id string)
    UpdateLoad(gatewayID string, currentLoad int)
}

type ConsistentHashBalancer struct {
    mu        sync.RWMutex
    gateways  map[string]*GatewayNode
    ring      []string
    loads     map[string]*atomic.Int32
}

type GatewayNode struct {
    ID           string
    Endpoint     string
    Capacity     int
    CurrentLoad  int32
    HealthStatus bool
}

func (b *ConsistentHashBalancer) SelectGateway(userID string) string {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    if len(b.ring) == 0 {
        return ""
    }
    
    // Use consistent hashing for sticky sessions
    h := fnv.New32a()
    h.Write([]byte(userID))
    hash := h.Sum32()
    
    idx := int(hash) % len(b.ring)
    gatewayID := b.ring[idx]
    
    // Check if gateway is healthy and not overloaded
    gateway := b.gateways[gatewayID]
    if !gateway.HealthStatus || atomic.LoadInt32(&gateway.CurrentLoad) >= int32(gateway.Capacity) {
        // Find next available gateway
        for i := 1; i < len(b.ring); i++ {
            idx = (idx + i) % len(b.ring)
            gatewayID = b.ring[idx]
            gateway = b.gateways[gatewayID]
            if gateway.HealthStatus && atomic.LoadInt32(&gateway.CurrentLoad) < int32(gateway.Capacity) {
                break
            }
        }
    }
    
    return gateway.Endpoint
}
```

### 1.4 Failover and Redundancy

```go
// internal/application/realtime/failover.go
package realtime

import (
    "context"
    "time"
)

type FailoverManager struct {
    sessionStore    ports.SessionStore
    stateStore      ports.StateStore
    eventBus        ports.EventBus
    gatewayRegistry ports.GatewayRegistry
    logger          logger.Logger
}

func (fm *FailoverManager) HandleGatewayFailure(ctx context.Context, failedGatewayID string) error {
    // 1. Mark gateway as unhealthy
    if err := fm.gatewayRegistry.SetHealth(failedGatewayID, false); err != nil {
        return err
    }
    
    // 2. Get all sessions from failed gateway
    sessions, err := fm.sessionStore.GetByGateway(ctx, failedGatewayID)
    if err != nil {
        return err
    }
    
    // 3. Redistribute sessions
    for _, session := range sessions {
        // Get new gateway
        newGateway := fm.gatewayRegistry.SelectHealthyGateway(session.UserID)
        if newGateway == "" {
            fm.logger.Errorf("No healthy gateway available for user %s", session.UserID)
            continue
        }
        
        // Save session state for recovery
        state := &SessionRecoveryState{
            SessionID:    session.ID,
            UserID:       session.UserID,
            CharacterID:  session.CharacterID,
            LastPosition: session.LastKnownPosition,
            Subscriptions: session.Subscriptions,
            Timestamp:    time.Now(),
        }
        
        if err := fm.stateStore.SaveRecoveryState(ctx, state); err != nil {
            fm.logger.WithError(err).Errorf("Failed to save recovery state for session %s", session.ID)
        }
        
        // Notify client to reconnect
        fm.eventBus.Publish(ctx, "session.reconnect", map[string]interface{}{
            "sessionID":   session.ID,
            "newGateway":  newGateway,
            "recoveryToken": generateRecoveryToken(session),
        })
    }
    
    return nil
}

// Client-side reconnection with state recovery
func (fm *FailoverManager) RecoverSession(ctx context.Context, recoveryToken string) (*PlayerSession, error) {
    // Validate recovery token
    claims, err := validateRecoveryToken(recoveryToken)
    if err != nil {
        return nil, err
    }
    
    // Load recovery state
    state, err := fm.stateStore.GetRecoveryState(ctx, claims.SessionID)
    if err != nil {
        return nil, err
    }
    
    // Restore session
    session := &PlayerSession{
        ID:          state.SessionID,
        UserID:      state.UserID,
        CharacterID: state.CharacterID,
        State:       SessionStateConnected,
    }
    
    // Restore subscriptions
    for _, sub := range state.Subscriptions {
        session.Subscribe(sub)
    }
    
    // Notify other services
    fm.eventBus.Publish(ctx, "session.recovered", map[string]interface{}{
        "sessionID":  session.ID,
        "userID":     session.UserID,
        "characterID": session.CharacterID,
    })
    
    return session, nil
}
```

---

## 📨 2. Message Protocol Design

### 2.1 Protocol Buffer Definitions for Real-time Messages

```protobuf
// pkg/proto/realtime.proto
syntax = "proto3";

package mmorpg.realtime;

option go_package = "github.com/mmorpg-template/backend/pkg/proto/realtime";

import "google/protobuf/timestamp.proto";
import "google/protobuf/any.proto";

// WebSocket message wrapper using JSON-RPC 2.0
message WebSocketMessage {
    string jsonrpc = 1;  // Always "2.0"
    oneof id_type {
        string id_string = 2;
        int64 id_number = 3;
    }
    
    oneof message_type {
        Request request = 4;
        Response response = 5;
        Notification notification = 6;
    }
}

message Request {
    string method = 1;
    google.protobuf.Any params = 2;
}

message Response {
    oneof result_type {
        google.protobuf.Any result = 1;
        Error error = 2;
    }
}

message Notification {
    string method = 1;
    google.protobuf.Any params = 2;
}

message Error {
    int32 code = 1;
    string message = 2;
    google.protobuf.Any data = 3;
}

// Real-time specific messages
message PositionUpdate {
    string character_id = 1;
    Vector3 position = 2;
    Rotation rotation = 3;
    Vector3 velocity = 4;
    google.protobuf.Timestamp timestamp = 5;
    uint32 sequence = 6;  // For ordering and interpolation
}

message StateSync {
    string entity_id = 1;
    string entity_type = 2;
    map<string, google.protobuf.Any> properties = 3;
    uint32 version = 4;  // For conflict resolution
    google.protobuf.Timestamp timestamp = 5;
}

message PresenceUpdate {
    string user_id = 1;
    string character_id = 2;
    PresenceStatus status = 3;
    string zone_id = 4;
    map<string, string> metadata = 5;
}

enum PresenceStatus {
    PRESENCE_STATUS_UNSPECIFIED = 0;
    PRESENCE_STATUS_ONLINE = 1;
    PRESENCE_STATUS_AWAY = 2;
    PRESENCE_STATUS_BUSY = 3;
    PRESENCE_STATUS_OFFLINE = 4;
}

// Batch message for efficiency
message BatchMessage {
    repeated WebSocketMessage messages = 1;
    bool compressed = 2;
    uint32 original_size = 3;  // If compressed
}
```

### 2.2 JSON-RPC 2.0 Wrapper Format

```go
// internal/adapters/websocket/protocol.go
package websocket

import (
    "encoding/json"
    "fmt"
    
    "google.golang.org/protobuf/proto"
    "google.golang.org/protobuf/types/known/anypb"
)

type JSONRPCMessage struct {
    Version string      `json:"jsonrpc"`
    ID      interface{} `json:"id,omitempty"`
    Method  string      `json:"method,omitempty"`
    Params  interface{} `json:"params,omitempty"`
    Result  interface{} `json:"result,omitempty"`
    Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// Protocol converter
type ProtocolConverter struct {
    logger logger.Logger
}

func (pc *ProtocolConverter) EncodeMessage(msg proto.Message, method string, id interface{}) ([]byte, error) {
    // Marshal protobuf to Any
    any, err := anypb.New(msg)
    if err != nil {
        return nil, fmt.Errorf("failed to create Any: %w", err)
    }
    
    // Create JSON-RPC message
    jsonRPC := &JSONRPCMessage{
        Version: "2.0",
        Method:  method,
        Params: map[string]interface{}{
            "@type": any.TypeUrl,
            "value": any.Value, // Base64 encoded
        },
    }
    
    if id != nil {
        jsonRPC.ID = id
    }
    
    return json.Marshal(jsonRPC)
}

func (pc *ProtocolConverter) DecodeMessage(data []byte) (proto.Message, string, interface{}, error) {
    var jsonRPC JSONRPCMessage
    if err := json.Unmarshal(data, &jsonRPC); err != nil {
        return nil, "", nil, fmt.Errorf("invalid JSON-RPC: %w", err)
    }
    
    if jsonRPC.Version != "2.0" {
        return nil, "", nil, fmt.Errorf("unsupported JSON-RPC version: %s", jsonRPC.Version)
    }
    
    // Extract protobuf from params
    params, ok := jsonRPC.Params.(map[string]interface{})
    if !ok {
        return nil, jsonRPC.Method, jsonRPC.ID, fmt.Errorf("invalid params format")
    }
    
    typeURL, _ := params["@type"].(string)
    value, _ := params["value"].([]byte)
    
    // Create Any and unmarshal
    any := &anypb.Any{
        TypeUrl: typeURL,
        Value:   value,
    }
    
    // Dynamic message creation based on type URL
    msg, err := pc.createMessageFromType(typeURL)
    if err != nil {
        return nil, jsonRPC.Method, jsonRPC.ID, err
    }
    
    if err := any.UnmarshalTo(msg); err != nil {
        return nil, jsonRPC.Method, jsonRPC.ID, fmt.Errorf("failed to unmarshal: %w", err)
    }
    
    return msg, jsonRPC.Method, jsonRPC.ID, nil
}
```

### 2.3 Message Types and Routing

```go
// internal/domain/realtime/router.go
package realtime

import (
    "context"
    "fmt"
    "sync"
    
    "google.golang.org/protobuf/proto"
)

type MessageHandler func(ctx context.Context, session *PlayerSession, msg proto.Message) error

type MessageRouter struct {
    mu       sync.RWMutex
    handlers map[string]MessageHandler
    logger   logger.Logger
}

func NewMessageRouter(logger logger.Logger) *MessageRouter {
    return &MessageRouter{
        handlers: make(map[string]MessageHandler),
        logger:   logger,
    }
}

func (r *MessageRouter) RegisterHandler(method string, handler MessageHandler) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.handlers[method] = handler
}

func (r *MessageRouter) Route(ctx context.Context, session *PlayerSession, method string, msg proto.Message) error {
    r.mu.RLock()
    handler, exists := r.handlers[method]
    r.mu.RUnlock()
    
    if !exists {
        return fmt.Errorf("no handler for method: %s", method)
    }
    
    // Add context enrichment
    ctx = context.WithValue(ctx, "sessionID", session.ID)
    ctx = context.WithValue(ctx, "userID", session.UserID)
    
    // Execute handler with panic recovery
    return r.safeExecute(ctx, session, msg, handler)
}

func (r *MessageRouter) safeExecute(ctx context.Context, session *PlayerSession, msg proto.Message, handler MessageHandler) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("handler panic: %v", r)
            r.logger.WithError(err).Error("Handler panicked")
        }
    }()
    
    return handler(ctx, session, msg)
}

// Route registration
func (r *MessageRouter) RegisterGameRoutes() {
    // Position updates
    r.RegisterHandler("world.position.update", handlePositionUpdate)
    r.RegisterHandler("world.position.batch", handlePositionBatch)
    
    // State synchronization
    r.RegisterHandler("state.sync.request", handleStateSyncRequest)
    r.RegisterHandler("state.sync.delta", handleStateDelta)
    
    // Presence
    r.RegisterHandler("presence.update", handlePresenceUpdate)
    r.RegisterHandler("presence.subscribe", handlePresenceSubscribe)
    
    // Chat (foundation for Phase 3)
    r.RegisterHandler("chat.message", handleChatMessage)
    r.RegisterHandler("chat.channel.join", handleChannelJoin)
    
    // System
    r.RegisterHandler("system.ping", handlePing)
    r.RegisterHandler("system.subscribe", handleSubscribe)
}
```

### 2.4 Error Handling and Recovery

```go
// internal/domain/realtime/errors.go
package realtime

import (
    "fmt"
)

// Standard JSON-RPC 2.0 error codes
const (
    ErrorCodeParseError     = -32700
    ErrorCodeInvalidRequest = -32600
    ErrorCodeMethodNotFound = -32601
    ErrorCodeInvalidParams  = -32602
    ErrorCodeInternalError  = -32603
    
    // Application-specific error codes
    ErrorCodeUnauthorized      = -32000
    ErrorCodeRateLimited       = -32001
    ErrorCodeSessionExpired    = -32002
    ErrorCodeInvalidState      = -32003
    ErrorCodeSubscriptionLimit = -32004
)

type ErrorHandler struct {
    logger logger.Logger
}

func (eh *ErrorHandler) HandleError(session *PlayerSession, err error, requestID interface{}) {
    // Classify error
    code, message, data := eh.classifyError(err)
    
    // Create error response
    errorResp := &JSONRPCMessage{
        Version: "2.0",
        ID:      requestID,
        Error: &JSONRPCError{
            Code:    code,
            Message: message,
            Data:    data,
        },
    }
    
    // Send error response
    if sendErr := session.SendMessage(errorResp); sendErr != nil {
        eh.logger.WithError(sendErr).Error("Failed to send error response")
    }
    
    // Log error for monitoring
    eh.logger.WithFields(logger.Fields{
        "sessionID": session.ID,
        "userID":    session.UserID,
        "errorCode": code,
        "error":     err.Error(),
    }).Error("WebSocket error")
}

func (eh *ErrorHandler) classifyError(err error) (code int, message string, data interface{}) {
    switch e := err.(type) {
    case *ValidationError:
        return ErrorCodeInvalidParams, "Invalid parameters", e.Details
    case *AuthorizationError:
        return ErrorCodeUnauthorized, "Unauthorized", nil
    case *RateLimitError:
        return ErrorCodeRateLimited, "Rate limit exceeded", map[string]interface{}{
            "retryAfter": e.RetryAfter.Seconds(),
        }
    default:
        return ErrorCodeInternalError, "Internal error", nil
    }
}

// Message retry mechanism
type RetryManager struct {
    maxRetries   int
    retryDelay   time.Duration
    backoffMultiplier float64
}

func (rm *RetryManager) RetryMessage(session *PlayerSession, msg *JSONRPCMessage, attempt int) error {
    if attempt >= rm.maxRetries {
        return fmt.Errorf("max retries exceeded")
    }
    
    // Calculate backoff
    delay := time.Duration(float64(rm.retryDelay) * math.Pow(rm.backoffMultiplier, float64(attempt)))
    
    // Schedule retry
    time.AfterFunc(delay, func() {
        if err := session.SendMessage(msg); err != nil {
            rm.RetryMessage(session, msg, attempt+1)
        }
    })
    
    return nil
}
```

---

## 🗄️ 3. State Management Architecture

### 3.1 Session State in Redis

```go
// internal/adapters/redis/session_store.go
package redis

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type SessionStore struct {
    client *redis.Client
    ttl    time.Duration
}

func NewSessionStore(client *redis.Client, ttl time.Duration) *SessionStore {
    return &SessionStore{
        client: client,
        ttl:    ttl,
    }
}

// Session data structure in Redis
type SessionData struct {
    ID           string                 `json:"id"`
    UserID       string                 `json:"user_id"`
    CharacterID  string                 `json:"character_id,omitempty"`
    GatewayID    string                 `json:"gateway_id"`
    State        string                 `json:"state"`
    ConnectedAt  time.Time              `json:"connected_at"`
    LastActivity time.Time              `json:"last_activity"`
    Attributes   map[string]interface{} `json:"attributes"`
    
    // Presence data
    Status       string    `json:"status"`
    ZoneID       string    `json:"zone_id,omitempty"`
    Position     *Vector3  `json:"position,omitempty"`
    
    // Subscriptions
    Subscriptions []string  `json:"subscriptions"`
}

func (s *SessionStore) Save(ctx context.Context, session *SessionData) error {
    key := fmt.Sprintf("session:%s", session.ID)
    
    data, err := json.Marshal(session)
    if err != nil {
        return fmt.Errorf("failed to marshal session: %w", err)
    }
    
    // Save with TTL
    if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
        return fmt.Errorf("failed to save session: %w", err)
    }
    
    // Add to user's session set
    userKey := fmt.Sprintf("user:sessions:%s", session.UserID)
    if err := s.client.SAdd(ctx, userKey, session.ID).Err(); err != nil {
        return fmt.Errorf("failed to add to user sessions: %w", err)
    }
    
    // Add to gateway's session set
    gatewayKey := fmt.Sprintf("gateway:sessions:%s", session.GatewayID)
    if err := s.client.SAdd(ctx, gatewayKey, session.ID).Err(); err != nil {
        return fmt.Errorf("failed to add to gateway sessions: %w", err)
    }
    
    // Update presence index
    if session.Status != "" && session.Status != "offline" {
        presenceKey := "presence:online"
        if err := s.client.ZAdd(ctx, presenceKey, &redis.Z{
            Score:  float64(time.Now().Unix()),
            Member: session.UserID,
        }).Err(); err != nil {
            return fmt.Errorf("failed to update presence: %w", err)
        }
    }
    
    return nil
}

func (s *SessionStore) Get(ctx context.Context, sessionID string) (*SessionData, error) {
    key := fmt.Sprintf("session:%s", sessionID)
    
    data, err := s.client.Get(ctx, key).Bytes()
    if err != nil {
        if err == redis.Nil {
            return nil, fmt.Errorf("session not found")
        }
        return nil, fmt.Errorf("failed to get session: %w", err)
    }
    
    var session SessionData
    if err := json.Unmarshal(data, &session); err != nil {
        return nil, fmt.Errorf("failed to unmarshal session: %w", err)
    }
    
    // Refresh TTL on access
    s.client.Expire(ctx, key, s.ttl)
    
    return &session, nil
}

// Atomic session state update
func (s *SessionStore) UpdateState(ctx context.Context, sessionID string, newState string) error {
    script := `
        local key = KEYS[1]
        local new_state = ARGV[1]
        local ttl = ARGV[2]
        
        local session = redis.call('GET', key)
        if not session then
            return redis.error_reply('session not found')
        end
        
        local data = cjson.decode(session)
        data.state = new_state
        data.last_activity = ARGV[3]
        
        redis.call('SET', key, cjson.encode(data), 'EX', ttl)
        return 'OK'
    `
    
    return s.client.Eval(ctx, script, []string{
        fmt.Sprintf("session:%s", sessionID),
    }, newState, int(s.ttl.Seconds()), time.Now().Unix()).Err()
}
```

### 3.2 Presence Tracking System

```go
// internal/application/realtime/presence.go
package realtime

import (
    "context"
    "fmt"
    "time"
)

type PresenceService struct {
    store      ports.PresenceStore
    eventBus   ports.EventBus
    logger     logger.Logger
}

func (ps *PresenceService) UpdatePresence(ctx context.Context, update *PresenceUpdate) error {
    // Validate status transition
    if err := ps.validateStatusTransition(ctx, update.UserID, update.Status); err != nil {
        return err
    }
    
    // Save presence data
    presence := &PresenceData{
        UserID:      update.UserID,
        CharacterID: update.CharacterID,
        Status:      update.Status,
        ZoneID:      update.ZoneID,
        LastUpdate:  time.Now(),
        Metadata:    update.Metadata,
    }
    
    if err := ps.store.SavePresence(ctx, presence); err != nil {
        return fmt.Errorf("failed to save presence: %w", err)
    }
    
    // Notify subscribers
    ps.notifyPresenceChange(ctx, presence)
    
    return nil
}

func (ps *PresenceService) GetPresenceBulk(ctx context.Context, userIDs []string) (map[string]*PresenceData, error) {
    return ps.store.GetMultiple(ctx, userIDs)
}

func (ps *PresenceService) SubscribeToPresence(ctx context.Context, subscriberID string, targetUserIDs []string) error {
    // Limit subscriptions per user
    const maxSubscriptions = 200
    
    currentSubs, err := ps.store.GetSubscriptions(ctx, subscriberID)
    if err != nil {
        return err
    }
    
    if len(currentSubs)+len(targetUserIDs) > maxSubscriptions {
        return fmt.Errorf("subscription limit exceeded")
    }
    
    // Add subscriptions
    for _, targetID := range targetUserIDs {
        if err := ps.store.AddSubscription(ctx, subscriberID, targetID); err != nil {
            return err
        }
    }
    
    // Send initial presence data
    presenceData, err := ps.GetPresenceBulk(ctx, targetUserIDs)
    if err != nil {
        return err
    }
    
    // Notify subscriber
    ps.eventBus.PublishToUser(ctx, subscriberID, "presence.bulk", presenceData)
    
    return nil
}

func (ps *PresenceService) notifyPresenceChange(ctx context.Context, presence *PresenceData) {
    // Get all subscribers
    subscribers, err := ps.store.GetSubscribers(ctx, presence.UserID)
    if err != nil {
        ps.logger.WithError(err).Error("Failed to get subscribers")
        return
    }
    
    // Batch notify
    event := map[string]interface{}{
        "userID":      presence.UserID,
        "characterID": presence.CharacterID,
        "status":      presence.Status,
        "zoneID":      presence.ZoneID,
        "timestamp":   presence.LastUpdate,
    }
    
    for _, subscriberID := range subscribers {
        ps.eventBus.PublishToUser(ctx, subscriberID, "presence.update", event)
    }
}

// Presence aggregation for friend lists
func (ps *PresenceService) GetFriendPresences(ctx context.Context, userID string) ([]*FriendPresence, error) {
    // Get friend list (would come from social service in full implementation)
    friends, err := ps.getFriendList(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // Get presence data
    presences, err := ps.GetPresenceBulk(ctx, friends)
    if err != nil {
        return nil, err
    }
    
    // Build friend presence list
    result := make([]*FriendPresence, 0, len(friends))
    for _, friendID := range friends {
        if presence, ok := presences[friendID]; ok {
            result = append(result, &FriendPresence{
                UserID:       friendID,
                CharacterID:  presence.CharacterID,
                Status:       presence.Status,
                ZoneID:       presence.ZoneID,
                LastOnline:   presence.LastUpdate,
                CharacterName: ps.getCharacterName(presence.CharacterID),
            })
        }
    }
    
    return result, nil
}
```

### 3.3 State Synchronization Patterns

```go
// internal/domain/realtime/state_sync.go
package realtime

import (
    "context"
    "sync"
    "time"
)

type StateManager struct {
    mu           sync.RWMutex
    states       map[string]*EntityState
    deltaBuffer  *DeltaBuffer
    snapshots    ports.SnapshotStore
    eventBus     ports.EventBus
    syncInterval time.Duration
}

type EntityState struct {
    ID         string
    Type       string
    Version    uint32
    Properties map[string]interface{}
    LastUpdate time.Time
    Dirty      bool
}

func (sm *StateManager) UpdateState(ctx context.Context, entityID string, updates map[string]interface{}) error {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    state, exists := sm.states[entityID]
    if !exists {
        state = &EntityState{
            ID:         entityID,
            Type:       extractEntityType(entityID),
            Version:    0,
            Properties: make(map[string]interface{}),
        }
        sm.states[entityID] = state
    }
    
    // Apply updates and track changes
    delta := &StateDelta{
        EntityID:  entityID,
        Version:   state.Version + 1,
        Changes:   make(map[string]*PropertyChange),
        Timestamp: time.Now(),
    }
    
    for key, newValue := range updates {
        oldValue := state.Properties[key]
        if !reflect.DeepEqual(oldValue, newValue) {
            delta.Changes[key] = &PropertyChange{
                OldValue: oldValue,
                NewValue: newValue,
            }
            state.Properties[key] = newValue
        }
    }
    
    if len(delta.Changes) > 0 {
        state.Version++
        state.LastUpdate = time.Now()
        state.Dirty = true
        
        // Buffer delta for batching
        sm.deltaBuffer.Add(delta)
    }
    
    return nil
}

// Delta compression and batching
type DeltaBuffer struct {
    mu       sync.Mutex
    deltas   []*StateDelta
    maxSize  int
    flushInterval time.Duration
    flushFunc func([]*StateDelta)
}

func (db *DeltaBuffer) Start() {
    ticker := time.NewTicker(db.flushInterval)
    go func() {
        for range ticker.C {
            db.Flush()
        }
    }()
}

func (db *DeltaBuffer) Add(delta *StateDelta) {
    db.mu.Lock()
    defer db.mu.Unlock()
    
    db.deltas = append(db.deltas, delta)
    
    if len(db.deltas) >= db.maxSize {
        db.flushNow()
    }
}

func (db *DeltaBuffer) flushNow() {
    if len(db.deltas) == 0 {
        return
    }
    
    // Compress deltas
    compressed := db.compressDeltas(db.deltas)
    
    // Send batch
    db.flushFunc(compressed)
    
    // Clear buffer
    db.deltas = db.deltas[:0]
}

func (db *DeltaBuffer) compressDeltas(deltas []*StateDelta) []*StateDelta {
    // Group by entity
    entityDeltas := make(map[string][]*StateDelta)
    for _, delta := range deltas {
        entityDeltas[delta.EntityID] = append(entityDeltas[delta.EntityID], delta)
    }
    
    // Compress each entity's deltas
    compressed := make([]*StateDelta, 0, len(entityDeltas))
    for entityID, deltas := range entityDeltas {
        if len(deltas) == 1 {
            compressed = append(compressed, deltas[0])
            continue
        }
        
        // Merge multiple deltas
        merged := &StateDelta{
            EntityID:  entityID,
            Version:   deltas[len(deltas)-1].Version,
            Changes:   make(map[string]*PropertyChange),
            Timestamp: deltas[len(deltas)-1].Timestamp,
        }
        
        // Keep only final state for each property
        for _, delta := range deltas {
            for key, change := range delta.Changes {
                merged.Changes[key] = change
            }
        }
        
        compressed = append(compressed, merged)
    }
    
    return compressed
}

// Interest management for state sync
type InterestManager struct {
    mu            sync.RWMutex
    subscriptions map[string]map[string]bool // sessionID -> entityIDs
    proximity     map[string]*ProximityArea   // entityID -> area
}

func (im *InterestManager) UpdateInterest(sessionID string, position Vector3, radius float32) {
    im.mu.Lock()
    defer im.mu.Unlock()
    
    // Find entities within radius
    nearbyEntities := im.findNearbyEntities(position, radius)
    
    // Update subscriptions
    currentSubs := im.subscriptions[sessionID]
    if currentSubs == nil {
        currentSubs = make(map[string]bool)
        im.subscriptions[sessionID] = currentSubs
    }
    
    // Add new subscriptions
    for _, entityID := range nearbyEntities {
        if !currentSubs[entityID] {
            currentSubs[entityID] = true
            // Notify about new entity entering interest
        }
    }
    
    // Remove old subscriptions
    for entityID := range currentSubs {
        if !contains(nearbyEntities, entityID) {
            delete(currentSubs, entityID)
            // Notify about entity leaving interest
        }
    }
}
```

### 3.4 Delta Compression Implementation

```go
// internal/domain/realtime/compression.go
package realtime

import (
    "bytes"
    "compress/zlib"
    "encoding/binary"
    "fmt"
)

type DeltaCompressor struct {
    compressionThreshold int
}

func NewDeltaCompressor(threshold int) *DeltaCompressor {
    return &DeltaCompressor{
        compressionThreshold: threshold,
    }
}

func (dc *DeltaCompressor) CompressMessage(msg []byte) ([]byte, bool, error) {
    // Don't compress small messages
    if len(msg) < dc.compressionThreshold {
        return msg, false, nil
    }
    
    var buf bytes.Buffer
    
    // Write uncompressed size
    binary.Write(&buf, binary.LittleEndian, uint32(len(msg)))
    
    // Compress
    w := zlib.NewWriter(&buf)
    if _, err := w.Write(msg); err != nil {
        return nil, false, err
    }
    if err := w.Close(); err != nil {
        return nil, false, err
    }
    
    compressed := buf.Bytes()
    
    // Only use compression if it reduces size by at least 20%
    if float64(len(compressed)) > float64(len(msg))*0.8 {
        return msg, false, nil
    }
    
    return compressed, true, nil
}

func (dc *DeltaCompressor) DecompressMessage(data []byte) ([]byte, error) {
    buf := bytes.NewReader(data)
    
    // Read uncompressed size
    var uncompressedSize uint32
    if err := binary.Read(buf, binary.LittleEndian, &uncompressedSize); err != nil {
        return nil, err
    }
    
    // Limit size to prevent memory attacks
    if uncompressedSize > 10*1024*1024 { // 10MB limit
        return nil, fmt.Errorf("uncompressed size too large: %d", uncompressedSize)
    }
    
    // Decompress
    r, err := zlib.NewReader(buf)
    if err != nil {
        return nil, err
    }
    defer r.Close()
    
    decompressed := make([]byte, uncompressedSize)
    if _, err := io.ReadFull(r, decompressed); err != nil {
        return nil, err
    }
    
    return decompressed, nil
}

// Delta encoding for position updates
type PositionDeltaEncoder struct {
    lastPositions map[string]*Vector3
    mu            sync.RWMutex
}

func (pde *PositionDeltaEncoder) EncodePosition(entityID string, pos *Vector3) *PositionDelta {
    pde.mu.Lock()
    defer pde.mu.Unlock()
    
    lastPos, exists := pde.lastPositions[entityID]
    if !exists {
        // First position, send absolute
        pde.lastPositions[entityID] = pos
        return &PositionDelta{
            EntityID: entityID,
            Absolute: pos,
        }
    }
    
    // Calculate delta
    delta := &Vector3{
        X: pos.X - lastPos.X,
        Y: pos.Y - lastPos.Y,
        Z: pos.Z - lastPos.Z,
    }
    
    // Use absolute if delta is too large
    if abs(delta.X) > 100 || abs(delta.Y) > 100 || abs(delta.Z) > 100 {
        pde.lastPositions[entityID] = pos
        return &PositionDelta{
            EntityID: entityID,
            Absolute: pos,
        }
    }
    
    // Update last position
    pde.lastPositions[entityID] = pos
    
    return &PositionDelta{
        EntityID: entityID,
        Delta:    delta,
    }
}
```

---

## 🔌 4. Integration Patterns

### 4.1 JWT Authentication for WebSocket

```go
// internal/adapters/websocket/auth.go
package websocket

import (
    "context"
    "fmt"
    "net/http"
    "strings"
    "time"
)

type WebSocketAuthenticator struct {
    authService ports.AuthService
    logger      logger.Logger
}

func (wa *WebSocketAuthenticator) AuthenticateUpgrade(r *http.Request) (*AuthClaims, error) {
    // Try multiple token sources
    token := wa.extractToken(r)
    if token == "" {
        return nil, fmt.Errorf("no authentication token provided")
    }
    
    // Validate token
    claims, err := wa.authService.ValidateToken(r.Context(), token)
    if err != nil {
        return nil, fmt.Errorf("invalid token: %w", err)
    }
    
    // Check token expiry with grace period
    if time.Now().After(claims.ExpiresAt.Add(-30 * time.Second)) {
        return nil, fmt.Errorf("token expired or expiring soon")
    }
    
    return claims, nil
}

func (wa *WebSocketAuthenticator) extractToken(r *http.Request) string {
    // 1. Try Authorization header
    if auth := r.Header.Get("Authorization"); auth != "" {
        parts := strings.Split(auth, " ")
        if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
            return parts[1]
        }
    }
    
    // 2. Try query parameter (for browser WebSocket API)
    if token := r.URL.Query().Get("token"); token != "" {
        return token
    }
    
    // 3. Try cookie (for web clients)
    if cookie, err := r.Cookie("auth_token"); err == nil {
        return cookie.Value
    }
    
    return ""
}

// Continuous auth validation during connection
func (wa *WebSocketAuthenticator) ValidateSession(ctx context.Context, session *PlayerSession) error {
    // Periodic revalidation
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Check if token is still valid
            if err := wa.authService.ValidateSession(ctx, session.UserID); err != nil {
                session.Close(CloseReasonAuthExpired, "Authentication expired")
                return err
            }
            
        case <-session.Done():
            return nil
        }
    }
}

// Token refresh over WebSocket
func (wa *WebSocketAuthenticator) HandleTokenRefresh(ctx context.Context, session *PlayerSession, refreshToken string) error {
    // Generate new access token
    newToken, err := wa.authService.RefreshToken(ctx, refreshToken)
    if err != nil {
        return fmt.Errorf("failed to refresh token: %w", err)
    }
    
    // Send new token to client
    response := &TokenRefreshResponse{
        AccessToken: newToken.AccessToken,
        ExpiresIn:   newToken.ExpiresIn,
    }
    
    return session.SendMessage("auth.token.refreshed", response)
}
```

### 4.2 Hybrid HTTP/WebSocket Approach

```go
// internal/adapters/gateway/hybrid_handler.go
package gateway

import (
    "net/http"
)

type HybridHandler struct {
    httpHandler      http.Handler
    websocketHandler *websocket.WebSocketHandler
    logger           logger.Logger
}

func (h *HybridHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Check for WebSocket upgrade
    if r.Header.Get("Upgrade") == "websocket" {
        h.handleWebSocketUpgrade(w, r)
        return
    }
    
    // Regular HTTP request
    h.httpHandler.ServeHTTP(w, r)
}

func (h *HybridHandler) handleWebSocketUpgrade(w http.ResponseWriter, r *http.Request) {
    // Apply rate limiting
    if !h.checkRateLimit(r) {
        http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
        return
    }
    
    // Authenticate
    claims, err := h.websocketHandler.Authenticate(r)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Upgrade connection
    h.websocketHandler.HandleUpgrade(w, r, claims)
}

// Seamless protocol switching
type ProtocolSwitcher struct {
    httpClient   ports.HTTPClient
    wsConnection ports.WebSocketConnection
}

func (ps *ProtocolSwitcher) Request(ctx context.Context, method string, data interface{}) (interface{}, error) {
    // Use WebSocket if connected and method supports it
    if ps.wsConnection != nil && ps.wsConnection.IsConnected() && supportsWebSocket(method) {
        return ps.wsConnection.Request(ctx, method, data)
    }
    
    // Fall back to HTTP
    return ps.httpClient.Request(ctx, method, data)
}

// Method routing configuration
var methodProtocols = map[string][]string{
    "auth.login":         {"http"},          // Always HTTP
    "auth.register":      {"http"},          // Always HTTP
    "world.position":     {"websocket"},     // Always WebSocket
    "character.list":     {"http", "websocket"}, // Either
    "chat.message":       {"websocket"},     // Always WebSocket
}

func supportsWebSocket(method string) bool {
    protocols, exists := methodProtocols[method]
    if !exists {
        return true // Default to supporting both
    }
    
    for _, p := range protocols {
        if p == "websocket" {
            return true
        }
    }
    return false
}
```

### 4.3 Event Bus Integration with NATS

```go
// internal/adapters/nats/realtime_events.go
package nats

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/nats-io/nats.go"
)

type RealtimeEventBus struct {
    conn   *nats.Conn
    js     nats.JetStreamContext
    logger logger.Logger
}

func NewRealtimeEventBus(conn *nats.Conn, logger logger.Logger) (*RealtimeEventBus, error) {
    js, err := conn.JetStream()
    if err != nil {
        return nil, err
    }
    
    // Create streams for real-time events
    if err := createRealtimeStreams(js); err != nil {
        return nil, err
    }
    
    return &RealtimeEventBus{
        conn:   conn,
        js:     js,
        logger: logger,
    }, nil
}

func createRealtimeStreams(js nats.JetStreamContext) error {
    // Presence stream
    _, err := js.AddStream(&nats.StreamConfig{
        Name:      "PRESENCE",
        Subjects:  []string{"presence.>"},
        Retention: nats.WorkQueuePolicy,
        MaxAge:    time.Hour,
    })
    if err != nil && err != nats.ErrStreamNameAlreadyInUse {
        return err
    }
    
    // State sync stream
    _, err = js.AddStream(&nats.StreamConfig{
        Name:      "STATE_SYNC",
        Subjects:  []string{"state.>"},
        Retention: nats.WorkQueuePolicy,
        MaxAge:    time.Minute * 5,
    })
    if err != nil && err != nats.ErrStreamNameAlreadyInUse {
        return err
    }
    
    // Position updates stream (high frequency)
    _, err = js.AddStream(&nats.StreamConfig{
        Name:      "POSITION",
        Subjects:  []string{"position.>"},
        Retention: nats.LimitsPolicy,
        MaxMsgs:   100000,
        MaxAge:    time.Second * 30,
    })
    if err != nil && err != nats.ErrStreamNameAlreadyInUse {
        return err
    }
    
    return nil
}

// Publish position update to specific zone
func (eb *RealtimeEventBus) PublishPositionUpdate(ctx context.Context, zoneID string, update *PositionUpdate) error {
    subject := fmt.Sprintf("position.zone.%s", zoneID)
    
    data, err := json.Marshal(update)
    if err != nil {
        return err
    }
    
    // Use Core NATS for low latency
    return eb.conn.Publish(subject, data)
}

// Subscribe to zone position updates
func (eb *RealtimeEventBus) SubscribeToZonePositions(ctx context.Context, zoneID string, handler func(*PositionUpdate)) error {
    subject := fmt.Sprintf("position.zone.%s", zoneID)
    
    sub, err := eb.conn.Subscribe(subject, func(msg *nats.Msg) {
        var update PositionUpdate
        if err := json.Unmarshal(msg.Data, &update); err != nil {
            eb.logger.WithError(err).Error("Failed to unmarshal position update")
            return
        }
        
        handler(&update)
    })
    
    if err != nil {
        return err
    }
    
    // Handle context cancellation
    go func() {
        <-ctx.Done()
        sub.Unsubscribe()
    }()
    
    return nil
}

// Cross-gateway message routing
func (eb *RealtimeEventBus) RouteToGateway(ctx context.Context, gatewayID string, sessionID string, message interface{}) error {
    subject := fmt.Sprintf("gateway.%s.session.%s", gatewayID, sessionID)
    
    envelope := &GatewayMessage{
        SessionID: sessionID,
        Message:   message,
        Timestamp: time.Now(),
    }
    
    data, err := json.Marshal(envelope)
    if err != nil {
        return err
    }
    
    // Request-reply pattern for confirmation
    msg, err := eb.conn.RequestWithContext(ctx, subject, data)
    if err != nil {
        return fmt.Errorf("failed to route message: %w", err)
    }
    
    // Check response
    var resp GatewayResponse
    if err := json.Unmarshal(msg.Data, &resp); err != nil {
        return fmt.Errorf("invalid gateway response: %w", err)
    }
    
    if !resp.Success {
        return fmt.Errorf("gateway rejected message: %s", resp.Error)
    }
    
    return nil
}
```

### 4.4 Service Mesh Considerations

```go
// internal/infrastructure/mesh/sidecar.go
package mesh

import (
    "context"
    "net/http"
    "time"
)

// Service mesh sidecar proxy for WebSocket
type WebSocketSidecar struct {
    serviceName string
    meshConfig  *MeshConfig
    metrics     *MetricsCollector
    tracing     *TracingProvider
}

type MeshConfig struct {
    CircuitBreaker   *CircuitBreakerConfig
    RetryPolicy      *RetryConfig
    LoadBalancing    *LoadBalancerConfig
    SecurityPolicy   *SecurityConfig
}

func (ws *WebSocketSidecar) WrapHandler(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Start trace
        span := ws.tracing.StartSpan("websocket.request", r.Header)
        defer span.End()
        
        // Apply security policies
        if err := ws.applySecurityPolicy(r); err != nil {
            http.Error(w, "Security policy violation", http.StatusForbidden)
            return
        }
        
        // Circuit breaker
        if !ws.meshConfig.CircuitBreaker.Allow() {
            http.Error(w, "Circuit breaker open", http.StatusServiceUnavailable)
            return
        }
        
        // Metrics
        start := time.Now()
        defer func() {
            ws.metrics.RecordRequest(ws.serviceName, time.Since(start))
        }()
        
        // Delegate to actual handler
        handler.ServeHTTP(w, r)
    })
}

// Envoy proxy integration for WebSocket
type EnvoyIntegration struct {
    config *EnvoyConfig
}

func (ei *EnvoyIntegration) GenerateWebSocketConfig() string {
    return `
static_resources:
  listeners:
  - name: websocket_listener
    address:
      socket_address:
        address: 0.0.0.0
        port_value: 8080
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          upgrade_configs:
          - upgrade_type: websocket
          route_config:
            name: local_route
            virtual_hosts:
            - name: websocket_service
              domains: ["*"]
              routes:
              - match:
                  prefix: "/ws"
                  headers:
                  - name: upgrade
                    exact_match: websocket
                route:
                  cluster: websocket_cluster
                  timeout: 0s
                  upgrade_configs:
                  - upgrade_type: websocket
                    enabled: true
`
}

// Istio WebSocket configuration
func generateIstioWebSocketPolicy() string {
    return `
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: websocket-service
spec:
  host: websocket-service
  trafficPolicy:
    connectionPool:
      http:
        http2MaxRequests: 10000
        h2UpgradePolicy: UPGRADE
    outlierDetection:
      consecutiveErrors: 5
      interval: 30s
      baseEjectionTime: 30s
`
}
```

---

## ⚡ 5. Performance Architecture

### 5.1 Connection Pooling

```go
// internal/infrastructure/pool/connection_pool.go
package pool

import (
    "context"
    "sync"
    "time"
)

type ConnectionPool struct {
    mu          sync.RWMutex
    connections map[string]*PooledConnection
    factory     ConnectionFactory
    config      *PoolConfig
    metrics     *PoolMetrics
}

type PoolConfig struct {
    MaxConnections      int
    MaxIdleConnections  int
    ConnectionTimeout   time.Duration
    IdleTimeout         time.Duration
    HealthCheckInterval time.Duration
}

type PooledConnection struct {
    conn         *websocket.Conn
    id           string
    createdAt    time.Time
    lastUsed     time.Time
    inUse        bool
    healthy      bool
}

func (p *ConnectionPool) Get(ctx context.Context) (*PooledConnection, error) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    // Find idle connection
    for id, conn := range p.connections {
        if !conn.inUse && conn.healthy {
            conn.inUse = true
            conn.lastUsed = time.Now()
            p.metrics.ConnectionBorrowed()
            return conn, nil
        }
    }
    
    // Create new if under limit
    if len(p.connections) < p.config.MaxConnections {
        conn, err := p.factory.Create(ctx)
        if err != nil {
            return nil, err
        }
        
        pooled := &PooledConnection{
            conn:      conn,
            id:        generateID(),
            createdAt: time.Now(),
            lastUsed:  time.Now(),
            inUse:     true,
            healthy:   true,
        }
        
        p.connections[pooled.id] = pooled
        p.metrics.ConnectionCreated()
        return pooled, nil
    }
    
    // Wait for available connection
    return p.waitForConnection(ctx)
}

func (p *ConnectionPool) Return(conn *PooledConnection) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    conn.inUse = false
    p.metrics.ConnectionReturned()
    
    // Check if should close
    if !conn.healthy || time.Since(conn.createdAt) > p.config.IdleTimeout {
        delete(p.connections, conn.id)
        conn.conn.Close()
        p.metrics.ConnectionClosed()
    }
}

// Health checking
func (p *ConnectionPool) healthCheck() {
    ticker := time.NewTicker(p.config.HealthCheckInterval)
    defer ticker.Stop()
    
    for range ticker.C {
        p.mu.Lock()
        for id, conn := range p.connections {
            if !conn.inUse {
                if err := p.ping(conn); err != nil {
                    conn.healthy = false
                    p.metrics.HealthCheckFailed()
                }
            }
        }
        p.mu.Unlock()
    }
}
```

### 5.2 Message Batching

```go
// internal/application/realtime/batcher.go
package realtime

import (
    "context"
    "sync"
    "time"
)

type MessageBatcher struct {
    mu            sync.Mutex
    batches       map[string]*Batch
    flushInterval time.Duration
    maxBatchSize  int
    sender        BatchSender
}

type Batch struct {
    SessionID string
    Messages  []interface{}
    Size      int
    Created   time.Time
}

func NewMessageBatcher(flushInterval time.Duration, maxBatchSize int, sender BatchSender) *MessageBatcher {
    mb := &MessageBatcher{
        batches:       make(map[string]*Batch),
        flushInterval: flushInterval,
        maxBatchSize:  maxBatchSize,
        sender:        sender,
    }
    
    go mb.periodicFlush()
    return mb
}

func (mb *MessageBatcher) Add(sessionID string, message interface{}, size int) {
    mb.mu.Lock()
    defer mb.mu.Unlock()
    
    batch, exists := mb.batches[sessionID]
    if !exists {
        batch = &Batch{
            SessionID: sessionID,
            Messages:  make([]interface{}, 0, mb.maxBatchSize),
            Created:   time.Now(),
        }
        mb.batches[sessionID] = batch
    }
    
    batch.Messages = append(batch.Messages, message)
    batch.Size += size
    
    // Flush if batch is full
    if len(batch.Messages) >= mb.maxBatchSize || batch.Size > 64*1024 {
        mb.flushBatch(sessionID)
    }
}

func (mb *MessageBatcher) flushBatch(sessionID string) {
    batch, exists := mb.batches[sessionID]
    if !exists || len(batch.Messages) == 0 {
        return
    }
    
    // Send batch
    go mb.sender.SendBatch(batch)
    
    // Reset batch
    delete(mb.batches, sessionID)
}

func (mb *MessageBatcher) periodicFlush() {
    ticker := time.NewTicker(mb.flushInterval)
    defer ticker.Stop()
    
    for range ticker.C {
        mb.mu.Lock()
        
        now := time.Now()
        for sessionID, batch := range mb.batches {
            if now.Sub(batch.Created) >= mb.flushInterval {
                mb.flushBatch(sessionID)
            }
        }
        
        mb.mu.Unlock()
    }
}

// Intelligent batching based on message type
type SmartBatcher struct {
    *MessageBatcher
    priorities map[string]int
}

func (sb *SmartBatcher) Add(sessionID string, method string, message interface{}, size int) {
    priority := sb.priorities[method]
    
    // High priority messages bypass batching
    if priority >= 10 {
        sb.sender.SendImmediate(sessionID, message)
        return
    }
    
    // Low priority messages are batched more aggressively
    if priority <= 3 {
        sb.MessageBatcher.Add(sessionID, message, size)
        return
    }
    
    // Medium priority: batch with shorter timeout
    sb.AddWithTimeout(sessionID, message, size, 50*time.Millisecond)
}
```

### 5.3 Rate Limiting

```go
// internal/application/realtime/ratelimit.go
package realtime

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    mu         sync.RWMutex
    limiters   map[string]*UserLimiter
    config     *RateLimitConfig
    store      RateLimitStore
}

type RateLimitConfig struct {
    // Global limits
    GlobalRPS        int
    GlobalBurstSize  int
    
    // Per-user limits
    UserRPS          int
    UserBurstSize    int
    
    // Per-method limits
    MethodLimits     map[string]*MethodLimit
    
    // Premium user multiplier
    PremiumMultiplier float64
}

type UserLimiter struct {
    global      *rate.Limiter
    perMethod   map[string]*rate.Limiter
    lastAccess  time.Time
    isPremium   bool
}

func (rl *RateLimiter) Allow(ctx context.Context, userID string, method string) error {
    rl.mu.Lock()
    limiter, exists := rl.limiters[userID]
    if !exists {
        limiter = rl.createUserLimiter(userID)
        rl.limiters[userID] = limiter
    }
    limiter.lastAccess = time.Now()
    rl.mu.Unlock()
    
    // Check global limit
    if !limiter.global.Allow() {
        return &RateLimitError{
            Type:       "global",
            RetryAfter: rl.calculateRetryAfter(limiter.global),
        }
    }
    
    // Check method-specific limit
    methodLimiter, exists := limiter.perMethod[method]
    if exists && !methodLimiter.Allow() {
        return &RateLimitError{
            Type:       "method",
            Method:     method,
            RetryAfter: rl.calculateRetryAfter(methodLimiter),
        }
    }
    
    return nil
}

func (rl *RateLimiter) createUserLimiter(userID string) *UserLimiter {
    // Check if premium user
    isPremium := rl.store.IsPremium(userID)
    
    // Calculate limits
    userRPS := rl.config.UserRPS
    userBurst := rl.config.UserBurstSize
    
    if isPremium {
        userRPS = int(float64(userRPS) * rl.config.PremiumMultiplier)
        userBurst = int(float64(userBurst) * rl.config.PremiumMultiplier)
    }
    
    limiter := &UserLimiter{
        global:     rate.NewLimiter(rate.Limit(userRPS), userBurst),
        perMethod:  make(map[string]*rate.Limiter),
        lastAccess: time.Now(),
        isPremium:  isPremium,
    }
    
    // Create method-specific limiters
    for method, limit := range rl.config.MethodLimits {
        methodRPS := limit.RPS
        methodBurst := limit.Burst
        
        if isPremium {
            methodRPS = int(float64(methodRPS) * rl.config.PremiumMultiplier)
            methodBurst = int(float64(methodBurst) * rl.config.PremiumMultiplier)
        }
        
        limiter.perMethod[method] = rate.NewLimiter(rate.Limit(methodRPS), methodBurst)
    }
    
    return limiter
}

// Distributed rate limiting with Redis
type DistributedRateLimiter struct {
    redis  *redis.Client
    config *RateLimitConfig
}

func (drl *DistributedRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
    now := time.Now()
    windowStart := now.Add(-window).Unix()
    
    pipe := drl.redis.Pipeline()
    
    // Remove old entries
    pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
    
    // Count current window
    count := pipe.ZCard(ctx, key)
    
    // Add current request
    pipe.ZAdd(ctx, key, &redis.Z{
        Score:  float64(now.Unix()),
        Member: now.UnixNano(),
    })
    
    // Set expiry
    pipe.Expire(ctx, key, window)
    
    _, err := pipe.Exec(ctx)
    if err != nil {
        return false, err
    }
    
    return count.Val() < int64(limit), nil
}
```

### 5.4 Monitoring and Metrics

```go
// internal/infrastructure/metrics/websocket_metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

type WebSocketMetrics struct {
    // Connection metrics
    ActiveConnections      prometheus.Gauge
    ConnectionsTotal       prometheus.Counter
    ConnectionDuration     prometheus.Histogram
    ConnectionErrors       prometheus.Counter
    
    // Message metrics
    MessagesReceived       prometheus.Counter
    MessagesSent           prometheus.Counter
    MessageSize            prometheus.Histogram
    MessageLatency         prometheus.Histogram
    
    // Protocol metrics
    ProtocolErrors         prometheus.Counter
    CompressionRatio       prometheus.Histogram
    
    // Performance metrics
    QueueDepth             prometheus.Gauge
    ProcessingTime         prometheus.Histogram
    
    // Business metrics
    ActiveUsers            prometheus.Gauge
    ZonePopulation         *prometheus.GaugeVec
}

func NewWebSocketMetrics() *WebSocketMetrics {
    return &WebSocketMetrics{
        ActiveConnections: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "websocket_active_connections",
            Help: "Current number of active WebSocket connections",
        }),
        
        ConnectionsTotal: promauto.NewCounter(prometheus.CounterOpts{
            Name: "websocket_connections_total",
            Help: "Total number of WebSocket connections",
        }),
        
        ConnectionDuration: promauto.NewHistogram(prometheus.HistogramOpts{
            Name:    "websocket_connection_duration_seconds",
            Help:    "Duration of WebSocket connections",
            Buckets: prometheus.ExponentialBuckets(1, 2, 15),
        }),
        
        MessageLatency: promauto.NewHistogram(prometheus.HistogramOpts{
            Name:    "websocket_message_latency_seconds",
            Help:    "Latency of message processing",
            Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
        }),
        
        ZonePopulation: promauto.NewGaugeVec(prometheus.GaugeOpts{
            Name: "game_zone_population",
            Help: "Number of players in each zone",
        }, []string{"zone_id"}),
    }
}

// Custom metrics collector
type MetricsCollector struct {
    metrics *WebSocketMetrics
}

func (mc *MetricsCollector) RecordConnection(session *PlayerSession) {
    mc.metrics.ActiveConnections.Inc()
    mc.metrics.ConnectionsTotal.Inc()
    
    // Track connection duration
    go func() {
        <-session.Done()
        duration := time.Since(session.CreatedAt).Seconds()
        mc.metrics.ConnectionDuration.Observe(duration)
        mc.metrics.ActiveConnections.Dec()
    }()
}

func (mc *MetricsCollector) RecordMessage(method string, size int, latency time.Duration) {
    labels := prometheus.Labels{
        "method": method,
    }
    
    mc.metrics.MessagesReceived.With(labels).Inc()
    mc.metrics.MessageSize.With(labels).Observe(float64(size))
    mc.metrics.MessageLatency.With(labels).Observe(latency.Seconds())
}

// Real-time dashboard data
type DashboardCollector struct {
    metrics *WebSocketMetrics
    store   MetricsStore
}

func (dc *DashboardCollector) GetRealtimeStats() *RealtimeStats {
    return &RealtimeStats{
        ActiveConnections:  int(dc.metrics.ActiveConnections.Get()),
        MessagesPerSecond:  dc.calculateMessageRate(),
        AverageLatency:     dc.calculateAverageLatency(),
        ActiveZones:        dc.getActiveZones(),
        SystemHealth:       dc.calculateSystemHealth(),
    }
}
```

---

## 🔒 6. Security Architecture

### 6.1 WebSocket Security Best Practices

```go
// internal/infrastructure/security/websocket_security.go
package security

import (
    "context"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "net/http"
    "time"
)

type WebSocketSecurity struct {
    config     *SecurityConfig
    validator  *MessageValidator
    firewall   *WebSocketFirewall
}

type SecurityConfig struct {
    // Origin validation
    AllowedOrigins      []string
    StrictOriginCheck   bool
    
    // Message security
    MaxMessageSize      int
    MessageSigningKey   []byte
    RequireEncryption   bool
    
    // Connection security
    MaxConnectionsPerIP int
    ConnectionTimeout   time.Duration
    
    // DDoS protection
    RateLimitPerIP      int
    BanDuration         time.Duration
}

func (ws *WebSocketSecurity) ValidateUpgradeRequest(r *http.Request) error {
    // Origin validation
    if ws.config.StrictOriginCheck {
        origin := r.Header.Get("Origin")
        if !ws.isAllowedOrigin(origin) {
            return fmt.Errorf("origin not allowed: %s", origin)
        }
    }
    
    // IP-based connection limit
    clientIP := extractClientIP(r)
    if ws.firewall.GetConnectionCount(clientIP) >= ws.config.MaxConnectionsPerIP {
        return fmt.Errorf("connection limit exceeded for IP: %s", clientIP)
    }
    
    // Check if IP is banned
    if ws.firewall.IsBanned(clientIP) {
        return fmt.Errorf("IP is banned: %s", clientIP)
    }
    
    // Validate upgrade headers
    if r.Header.Get("Upgrade") != "websocket" {
        return fmt.Errorf("invalid upgrade header")
    }
    
    return nil
}

// Message signing for integrity
func (ws *WebSocketSecurity) SignMessage(message []byte) ([]byte, error) {
    h := hmac.New(sha256.New, ws.config.MessageSigningKey)
    h.Write(message)
    signature := h.Sum(nil)
    
    // Append signature to message
    signed := append(message, signature...)
    return signed, nil
}

func (ws *WebSocketSecurity) VerifyMessage(signed []byte) ([]byte, error) {
    if len(signed) < 32 {
        return nil, fmt.Errorf("message too short")
    }
    
    // Split message and signature
    message := signed[:len(signed)-32]
    signature := signed[len(signed)-32:]
    
    // Verify signature
    h := hmac.New(sha256.New, ws.config.MessageSigningKey)
    h.Write(message)
    expected := h.Sum(nil)
    
    if !hmac.Equal(signature, expected) {
        return nil, fmt.Errorf("invalid message signature")
    }
    
    return message, nil
}
```

### 6.2 Message Validation

```go
// internal/infrastructure/security/message_validator.go
package security

import (
    "encoding/json"
    "fmt"
    "regexp"
    "strings"
)

type MessageValidator struct {
    rules      map[string]*ValidationRule
    sanitizer  *Sanitizer
}

type ValidationRule struct {
    MaxSize        int
    RequiredFields []string
    FieldRules     map[string]*FieldRule
}

type FieldRule struct {
    Type       string
    MinLength  int
    MaxLength  int
    Pattern    *regexp.Regexp
    Sanitize   bool
}

func (mv *MessageValidator) Validate(method string, data []byte) error {
    rule, exists := mv.rules[method]
    if !exists {
        return nil // No validation rules defined
    }
    
    // Check message size
    if len(data) > rule.MaxSize {
        return fmt.Errorf("message too large: %d > %d", len(data), rule.MaxSize)
    }
    
    // Parse message
    var message map[string]interface{}
    if err := json.Unmarshal(data, &message); err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }
    
    // Check required fields
    for _, field := range rule.RequiredFields {
        if _, exists := message[field]; !exists {
            return fmt.Errorf("missing required field: %s", field)
        }
    }
    
    // Validate fields
    for fieldName, fieldRule := range rule.FieldRules {
        value, exists := message[fieldName]
        if !exists {
            continue
        }
        
        if err := mv.validateField(fieldName, value, fieldRule); err != nil {
            return err
        }
        
        // Sanitize if needed
        if fieldRule.Sanitize {
            message[fieldName] = mv.sanitizer.Sanitize(value)
        }
    }
    
    return nil
}

// Input sanitization
type Sanitizer struct {
    htmlPolicy     *bluemonday.Policy
    sqlPattern     *regexp.Regexp
    scriptPattern  *regexp.Regexp
}

func NewSanitizer() *Sanitizer {
    return &Sanitizer{
        htmlPolicy:    bluemonday.StrictPolicy(),
        sqlPattern:    regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop)`),
        scriptPattern: regexp.MustCompile(`(?i)<script|javascript:|onerror=`),
    }
}

func (s *Sanitizer) Sanitize(input interface{}) interface{} {
    switch v := input.(type) {
    case string:
        // Remove HTML
        cleaned := s.htmlPolicy.Sanitize(v)
        
        // Check for SQL injection patterns
        if s.sqlPattern.MatchString(cleaned) {
            cleaned = s.sqlPattern.ReplaceAllString(cleaned, "")
        }
        
        // Check for XSS patterns
        if s.scriptPattern.MatchString(cleaned) {
            cleaned = s.scriptPattern.ReplaceAllString(cleaned, "")
        }
        
        return strings.TrimSpace(cleaned)
        
    case map[string]interface{}:
        // Recursively sanitize object
        for k, v := range v {
            v[k] = s.Sanitize(v)
        }
        return v
        
    case []interface{}:
        // Recursively sanitize array
        for i, item := range v {
            v[i] = s.Sanitize(item)
        }
        return v
        
    default:
        return v
    }
}
```

### 6.3 DDoS Protection

```go
// internal/infrastructure/security/ddos_protection.go
package security

import (
    "context"
    "sync"
    "time"
)

type DDoSProtection struct {
    mu              sync.RWMutex
    connectionStats map[string]*ConnectionStats
    suspiciousIPs   map[string]*SuspiciousActivity
    config          *DDoSConfig
}

type DDoSConfig struct {
    // Detection thresholds
    MaxConnectionsPerSecond   int
    MaxMessagesPerSecond      int
    MaxBandwidthPerSecond     int64
    
    // Pattern detection
    SuspiciousPatterns        []string
    AnomalyThreshold          float64
    
    // Response actions
    AutoBanEnabled            bool
    BanDuration               time.Duration
    AlertThreshold            int
}

type ConnectionStats struct {
    Connections      int
    Messages         int
    Bandwidth        int64
    LastReset        time.Time
    SuspiciousCount  int
}

func (dp *DDoSProtection) CheckConnection(ctx context.Context, clientIP string) error {
    dp.mu.Lock()
    defer dp.mu.Unlock()
    
    stats, exists := dp.connectionStats[clientIP]
    if !exists {
        stats = &ConnectionStats{
            LastReset: time.Now(),
        }
        dp.connectionStats[clientIP] = stats
    }
    
    // Reset stats if window expired
    if time.Since(stats.LastReset) > time.Second {
        stats.Connections = 0
        stats.Messages = 0
        stats.Bandwidth = 0
        stats.LastReset = time.Now()
    }
    
    stats.Connections++
    
    // Check thresholds
    if stats.Connections > dp.config.MaxConnectionsPerSecond {
        dp.handleSuspiciousActivity(clientIP, "excessive_connections")
        return fmt.Errorf("connection rate exceeded")
    }
    
    return nil
}

func (dp *DDoSProtection) handleSuspiciousActivity(clientIP string, reason string) {
    activity, exists := dp.suspiciousIPs[clientIP]
    if !exists {
        activity = &SuspiciousActivity{
            FirstSeen: time.Now(),
            Reasons:   make(map[string]int),
        }
        dp.suspiciousIPs[clientIP] = activity
    }
    
    activity.Count++
    activity.Reasons[reason]++
    activity.LastSeen = time.Now()
    
    // Auto-ban if threshold reached
    if dp.config.AutoBanEnabled && activity.Count >= dp.config.AlertThreshold {
        dp.banIP(clientIP, dp.config.BanDuration)
    }
}

// Pattern-based detection
type PatternDetector struct {
    patterns []AttackPattern
    ml       *AnomalyDetector
}

type AttackPattern struct {
    Name        string
    Indicators  []string
    Threshold   int
    TimeWindow  time.Duration
}

func (pd *PatternDetector) Analyze(session *PlayerSession, message []byte) *ThreatAssessment {
    assessment := &ThreatAssessment{
        Timestamp: time.Now(),
        SessionID: session.ID,
    }
    
    // Check known patterns
    for _, pattern := range pd.patterns {
        if pd.matchesPattern(message, pattern) {
            assessment.Threats = append(assessment.Threats, &Threat{
                Type:     pattern.Name,
                Severity: "high",
                Action:   "block",
            })
        }
    }
    
    // ML-based anomaly detection
    anomalyScore := pd.ml.Score(session, message)
    if anomalyScore > 0.8 {
        assessment.Threats = append(assessment.Threats, &Threat{
            Type:     "anomaly",
            Severity: "medium",
            Score:    anomalyScore,
            Action:   "monitor",
        })
    }
    
    return assessment
}
```

### 6.4 Encryption Considerations

```go
// internal/infrastructure/security/encryption.go
package security

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "encoding/base64"
    "io"
)

type EncryptionManager struct {
    privateKey  *rsa.PrivateKey
    publicKeys  map[string]*rsa.PublicKey
    symmetricKey []byte
}

// TLS configuration for WebSocket
func GetTLSConfig() *tls.Config {
    return &tls.Config{
        MinVersion:               tls.VersionTLS12,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
        },
        PreferServerCipherSuites: true,
        CurvePreferences: []tls.CurveID{
            tls.CurveP256,
            tls.X25519,
        },
    }
}

// End-to-end encryption for sensitive data
func (em *EncryptionManager) EncryptMessage(plaintext []byte, recipientID string) ([]byte, error) {
    // Get recipient's public key
    publicKey, exists := em.publicKeys[recipientID]
    if !exists {
        return nil, fmt.Errorf("no public key for recipient: %s", recipientID)
    }
    
    // Generate symmetric key for this message
    sessionKey := make([]byte, 32)
    if _, err := io.ReadFull(rand.Reader, sessionKey); err != nil {
        return nil, err
    }
    
    // Encrypt symmetric key with RSA
    encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, sessionKey, nil)
    if err != nil {
        return nil, err
    }
    
    // Encrypt message with AES-GCM
    block, err := aes.NewCipher(sessionKey)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    
    ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
    
    // Combine encrypted key, nonce, and ciphertext
    result := append(encryptedKey, nonce...)
    result = append(result, ciphertext...)
    
    return result, nil
}

// Key rotation
type KeyRotationManager struct {
    currentKey  []byte
    previousKey []byte
    rotationInterval time.Duration
    lastRotation    time.Time
}

func (krm *KeyRotationManager) RotateKeys() error {
    // Generate new key
    newKey := make([]byte, 32)
    if _, err := io.ReadFull(rand.Reader, newKey); err != nil {
        return err
    }
    
    // Keep previous key for decryption
    krm.previousKey = krm.currentKey
    krm.currentKey = newKey
    krm.lastRotation = time.Now()
    
    // Notify all connected clients
    // Implementation depends on your notification system
    
    return nil
}

func (krm *KeyRotationManager) GetDecryptionKey(timestamp time.Time) []byte {
    if timestamp.Before(krm.lastRotation) && krm.previousKey != nil {
        return krm.previousKey
    }
    return krm.currentKey
}
```

---

## 📊 Implementation Examples

### Complete WebSocket Handler Example

```go
// cmd/gateway/websocket/handler.go
package main

import (
    "context"
    "net/http"
    
    "github.com/mmorpg-template/backend/internal/adapters/websocket"
    "github.com/mmorpg-template/backend/internal/application/realtime"
    "github.com/mmorpg-template/backend/internal/config"
    "github.com/mmorpg-template/backend/pkg/logger"
)

func setupWebSocketHandler(cfg *config.Config, log logger.Logger) http.Handler {
    // Initialize dependencies
    sessionManager := realtime.NewSessionManager(cfg.Redis, log)
    messageRouter := realtime.NewMessageRouter(log)
    stateManager := realtime.NewStateManager(cfg.Redis, log)
    presenceService := realtime.NewPresenceService(sessionManager, log)
    
    // Create WebSocket handler
    wsHandler := websocket.NewWebSocketHandler(
        sessionManager,
        messageRouter,
        cfg.Auth,
        cfg.EventBus,
        log,
    )
    
    // Register message handlers
    messageRouter.RegisterGameRoutes()
    
    // Setup middleware chain
    handler := websocket.ChainMiddleware(
        wsHandler,
        websocket.RateLimitMiddleware(cfg.RateLimit),
        websocket.SecurityMiddleware(cfg.Security),
        websocket.MetricsMiddleware(cfg.Metrics),
    )
    
    return handler
}

// Enhanced gateway main with WebSocket support
func main() {
    ctx := context.Background()
    log := logger.NewWithService("gateway")
    cfg := config.Load()
    
    // Setup HTTP routes (existing)
    httpMux := setupHTTPRoutes(cfg, log)
    
    // Setup WebSocket handler
    wsMux := http.NewServeMux()
    wsMux.Handle("/ws", setupWebSocketHandler(cfg, log))
    
    // Combine handlers
    mux := http.NewServeMux()
    mux.Handle("/ws", wsMux)
    mux.Handle("/", httpMux)
    
    // Start server
    srv := &http.Server{
        Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
        Handler: mux,
    }
    
    log.Infof("Gateway with WebSocket support listening on %s", srv.Addr)
    if err := srv.ListenAndServe(); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

---

## 🎯 Summary

This architecture provides a robust foundation for Phase 2's real-time networking implementation:

1. **WebSocket Integration**: Clean integration into existing gateway service
2. **Protocol Design**: JSON-RPC 2.0 with Protocol Buffer payloads for efficiency
3. **State Management**: Redis-based session and presence tracking
4. **Performance**: Connection pooling, message batching, and compression
5. **Security**: Comprehensive protection against common WebSocket vulnerabilities
6. **Scalability**: Designed to handle thousands of concurrent connections

The architecture maintains consistency with Phase 1's hexagonal pattern while adding the real-time capabilities needed for Phase 3's gameplay features.

### Next Steps

1. Implement core WebSocket handler in gateway service
2. Set up Redis for session management
3. Create Protocol Buffer definitions for real-time messages
4. Implement basic message routing and state synchronization
5. Add monitoring and metrics collection
6. Load test with simulated clients

This architecture serves as the technical blueprint for Phase 2 implementation, ensuring we build a scalable, secure, and performant real-time networking layer for the MMORPG template.