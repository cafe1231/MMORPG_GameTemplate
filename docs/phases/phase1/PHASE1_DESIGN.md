# Phase 1 - Design - MMORPG Template Architecture

## Architecture Philosophy

### Core Principles
1. **Write Once, Scale Anywhere** - Same codebase from 1 to 1M+ players
2. **Infrastructure Agnostic** - No vendor lock-in, works with any cloud provider
3. **Developer First** - Clear abstractions, extensive documentation, easy customization
4. **Production Ready** - Battle-tested patterns, not experimental technology
5. **Cost Optimized** - Efficient resource usage at every scale

### Design Goals
- **Modularity**: Every component can be replaced or extended
- **Simplicity**: Complex internally, simple externally
- **Performance**: Optimized for real-time multiplayer at scale
- **Security**: Built-in protection against common exploits
- **Flexibility**: Customers can modify anything

## Technology Stack

### Client (Unreal Engine 5.6)
- **Core Language**: C++ for performance-critical systems
- **Blueprint Integration**: Full exposure for designer-friendly development
- **Serialization**: Protocol Buffers for efficient network communication
- **Architecture**: Component-based with clear separation of concerns

### Backend (Go Microservices)
- **Language**: Go 1.21+ for excellent concurrency and performance
- **Architecture**: Hexagonal (Ports & Adapters) for flexibility
- **Communication**: Protocol Buffers over WebSocket for real-time
- **Service Mesh**: Optional Istio/Linkerd support for enterprise deployments

### Data Layer
- **Primary Database**: PostgreSQL 16+ with sharding support
- **Caching**: Redis 7+ with cluster mode
- **Message Queue**: NATS 2.10+ for inter-service communication
- **Search**: Optional Elasticsearch for game analytics

### Infrastructure
- **Containers**: Docker with multi-stage builds
- **Orchestration**: Kubernetes with Helm charts
- **Monitoring**: Prometheus + Grafana dashboards
- **Logging**: Structured logging with ELK stack support

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT LAYER                              │
├─────────────────────────────────────────────────────────────────┤
│  Unreal Engine 5.6 Client                                       │
│  ├── MMORPGPlugin (C++)                                         │
│  ├── Game Logic (Blueprint/C++)                                 │
│  └── Protocol Buffer Serialization                              │
└─────────────────────────────────────────────────────────────────┘
                                │
                                │ WebSocket + Protobuf
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        GATEWAY LAYER                             │
├─────────────────────────────────────────────────────────────────┤
│  API Gateway (Go)                                               │
│  ├── Load Balancing (Consistent Hashing)                        │
│  ├── Rate Limiting                                              │
│  ├── Protocol Version Negotiation                               │
│  └── Regional Routing                                           │
└─────────────────────────────────────────────────────────────────┘
                                │
                    ┌───────────┴───────────┐
                    ▼                       ▼
┌─────────────────────────┐    ┌─────────────────────────┐
│   REAL-TIME SERVICES    │    │      API SERVICES       │
├─────────────────────────┤    ├─────────────────────────┤
│  World Service (Go)     │    │  Auth Service (Go)      │
│  ├── Player Position    │    │  ├── JWT Generation     │
│  ├── Combat Actions     │    │  ├── Session Management │
│  ├── Chat Messages      │    │  └── Account CRUD       │
│  └── Interest Mgmt      │    │                         │
│                         │    │  Character Service (Go)  │
│  Game Service (Go)      │    │  ├── Character CRUD     │
│  ├── Inventory          │    │  ├── Stats Management   │
│  ├── Quests             │    │  └── Equipment          │
│  └── NPCs               │    │                         │
└─────────────────────────┘    └─────────────────────────┘
            │                               │
            └───────────┬───────────────────┘
                        ▼
┌─────────────────────────────────────────────────────────────────┐
│                        DATA LAYER                                │
├─────────────────────────────────────────────────────────────────┤
│  PostgreSQL          Redis              NATS                    │
│  ├── Sharding        ├── Session Cache  ├── Event Bus          │
│  ├── Read Replicas   ├── Game State     └── Service Discovery  │
│  └── Connection Pool └── Leaderboards                           │
└─────────────────────────────────────────────────────────────────┘
```

## Detailed Component Design

### Unreal Engine Plugin Structure

```
Plugins/MMORPGTemplate/
├── Source/
│   ├── MMORPGCore/           # Core systems
│   │   ├── Public/
│   │   │   ├── Network/
│   │   │   │   ├── MMORPGNetworkManager.h
│   │   │   │   ├── MMORPGProtobufSerializer.h
│   │   │   │   └── MMORPGWebSocketClient.h
│   │   │   ├── Authentication/
│   │   │   │   ├── MMORPGAuthManager.h
│   │   │   │   └── MMORPGSessionManager.h
│   │   │   ├── Data/
│   │   │   │   ├── MMORPGDataManager.h
│   │   │   │   └── MMORPGCacheManager.h
│   │   │   └── Gameplay/
│   │   │       ├── MMORPGPlayerController.h
│   │   │       └── MMORPGCharacter.h
│   │   └── Private/
│   │       └── [Implementation files]
│   └── MMORPGEditor/         # Editor extensions
│       └── [Editor tools and utilities]
├── Content/
│   ├── Blueprints/
│   ├── UI/
│   └── Examples/
└── Config/
    └── DefaultMMORPG.ini
```

### Go Backend Structure

```
mmorpg-backend/
├── cmd/
│   ├── gateway/          # API Gateway service
│   ├── auth/             # Authentication service
│   ├── character/        # Character management service
│   ├── world/            # World/real-time service
│   └── game/             # Game logic service
├── internal/
│   ├── domain/           # Business logic (pure Go)
│   │   ├── auth/
│   │   ├── character/
│   │   ├── world/
│   │   └── game/
│   ├── adapters/         # External interfaces
│   │   ├── postgres/
│   │   ├── redis/
│   │   ├── websocket/
│   │   └── nats/
│   └── ports/            # Interface definitions
├── pkg/
│   ├── proto/            # Protocol Buffer definitions
│   │   ├── auth.proto
│   │   ├── character.proto
│   │   ├── world.proto
│   │   └── game.proto
│   ├── middleware/       # Shared middleware
│   ├── logger/           # Structured logging
│   └── metrics/          # Prometheus metrics
├── deployments/
│   ├── docker/
│   ├── kubernetes/
│   └── terraform/
└── scripts/
    ├── setup.sh
    └── test.sh
```

## Protocol Buffer Definitions

### Core Message Structure
```protobuf
syntax = "proto3";
package mmorpg;

// Base message envelope
message GameMessage {
    uint32 version = 1;        // Protocol version
    uint32 sequence = 2;       // Message sequence number
    int64 timestamp = 3;       // Unix timestamp in milliseconds
    MessageType type = 4;      // Message type
    bytes payload = 5;         // Actual message data
}

enum MessageType {
    MESSAGE_TYPE_UNSPECIFIED = 0;
    MESSAGE_TYPE_AUTH = 1;
    MESSAGE_TYPE_CHARACTER = 2;
    MESSAGE_TYPE_WORLD = 3;
    MESSAGE_TYPE_GAME = 4;
    MESSAGE_TYPE_CHAT = 5;
    MESSAGE_TYPE_SYSTEM = 6;
}
```

### Authentication Messages
```protobuf
// Login request
message LoginRequest {
    string email = 1;
    string password = 2;
    string client_version = 3;
}

// Login response
message LoginResponse {
    bool success = 1;
    string access_token = 2;
    string refresh_token = 3;
    string session_id = 4;
    int32 expires_in = 5;
    string error_message = 6;
}
```

### World Synchronization
```protobuf
// Player position update
message PlayerPosition {
    string player_id = 1;
    float x = 2;
    float y = 3;
    float z = 4;
    float rotation_yaw = 5;
    float rotation_pitch = 6;
    float rotation_roll = 7;
    float velocity_x = 8;
    float velocity_y = 9;
    float velocity_z = 10;
    int64 timestamp = 11;
}

// Area update (sent to clients)
message AreaUpdate {
    repeated PlayerPosition players = 1;
    repeated NPCPosition npcs = 2;
    repeated WorldObject objects = 3;
}
```

## Service Implementations

### Authentication Service (Go)
```go
// internal/domain/auth/service.go
package auth

import (
    "context"
    "time"
)

type Service interface {
    Login(ctx context.Context, email, password string) (*User, *Token, error)
    Register(ctx context.Context, req *RegisterRequest) (*User, error)
    ValidateToken(ctx context.Context, token string) (*Claims, error)
    RefreshToken(ctx context.Context, refreshToken string) (*Token, error)
    Logout(ctx context.Context, sessionID string) error
}

type Token struct {
    AccessToken  string
    RefreshToken string
    ExpiresIn    int
    SessionID    string
}

// Implementation with hexagonal architecture
type service struct {
    userRepo    UserRepository
    tokenRepo   TokenRepository
    sessionRepo SessionRepository
    hasher      PasswordHasher
    jwtService  JWTService
}
```

### World Service (Go)
```go
// internal/domain/world/service.go
package world

import (
    "context"
    "sync"
)

type Service interface {
    JoinWorld(ctx context.Context, playerID string, character Character) error
    UpdatePosition(ctx context.Context, playerID string, pos Position) error
    GetNearbyEntities(ctx context.Context, playerID string) (*AreaUpdate, error)
    LeaveWorld(ctx context.Context, playerID string) error
}

// High-performance spatial indexing
type spatialIndex struct {
    mu       sync.RWMutex
    octree   *Octree
    players  map[string]*Player
    gridSize float64
}

func (si *spatialIndex) UpdatePosition(playerID string, pos Position) {
    si.mu.Lock()
    defer si.mu.Unlock()
    
    player := si.players[playerID]
    if player != nil {
        si.octree.Remove(player)
        player.Position = pos
        si.octree.Insert(player)
    }
}
```

## Network Protocol

### WebSocket Connection Flow
```
1. Client → Gateway: HTTP Upgrade to WebSocket
2. Client → Gateway: Handshake with version info
3. Gateway → Client: Accept with session details
4. Client → Gateway: Authentication token
5. Gateway → Client: Authentication result
6. Client ↔ Gateway: Game messages (protobuf)
```

### Message Framing
```
┌─────────────┬─────────────┬─────────────┬─────────────┐
│  Version    │  Length     │  Type       │  Payload    │
│  (2 bytes)  │  (4 bytes)  │  (2 bytes)  │  (variable) │
└─────────────┴─────────────┴─────────────┴─────────────┘
```

## Scalability Patterns

### Horizontal Scaling
```go
// Load balancer with consistent hashing
type LoadBalancer struct {
    ring    *hashring.HashRing
    servers map[string]*ServerInfo
    mu      sync.RWMutex
}

func (lb *LoadBalancer) GetServer(playerID string) *ServerInfo {
    lb.mu.RLock()
    defer lb.mu.RUnlock()
    
    serverID := lb.ring.GetNode(playerID)
    return lb.servers[serverID]
}
```

### Database Sharding
```go
// Sharding strategy
func GetShardID(userID string) int {
    hash := fnv.New32a()
    hash.Write([]byte(userID))
    return int(hash.Sum32() % uint32(shardCount))
}

// Connection routing
func (db *ShardedDB) GetConnection(userID string) *sql.DB {
    shardID := GetShardID(userID)
    return db.shards[shardID]
}
```

## Deployment Configurations

### Local Development (docker-compose.yml)
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: mmorpg
      POSTGRES_USER: dev
      POSTGRES_PASSWORD: dev
    ports:
      - "5432:5432"
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
  
  nats:
    image: nats:2.10-alpine
    ports:
      - "4222:4222"
  
  gateway:
    build: ./cmd/gateway
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - NATS_URL=nats://nats:4222
    depends_on:
      - postgres
      - redis
      - nats
```

### Production Kubernetes (gateway-deployment.yaml)
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: gateway
  template:
    metadata:
      labels:
        app: gateway
    spec:
      containers:
      - name: gateway
        image: mmorpg/gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: host
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

## Security Architecture

### Authentication Flow
```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  Client  │────►│ Gateway  │────►│   Auth   │────►│    DB    │
│          │     │          │     │ Service  │     │          │
└──────────┘     └──────────┘     └──────────┘     └──────────┘
     │                 │                 │                 │
     │   1. Login     │                 │                 │
     │────────────────►                 │                 │
     │                 │  2. Validate   │                 │
     │                 │────────────────►                 │
     │                 │                 │  3. Check User │
     │                 │                 │────────────────►
     │                 │                 │                 │
     │                 │                 │◄────────────────
     │                 │  4. JWT Token  │                 │
     │                 │◄────────────────                 │
     │  5. Token      │                 │                 │
     │◄────────────────                 │                 │
```

### Anti-Cheat Measures
```go
// Server-side validation
type MovementValidator struct {
    maxSpeed     float64
    maxJumpHeight float64
    teleportThreshold float64
}

func (mv *MovementValidator) Validate(oldPos, newPos Position, deltaTime float64) error {
    distance := oldPos.DistanceTo(newPos)
    speed := distance / deltaTime
    
    if speed > mv.maxSpeed {
        return ErrSpeedHack
    }
    
    if distance > mv.teleportThreshold {
        return ErrTeleportHack
    }
    
    return nil
}
```

## Performance Optimizations

### Interest Management
```go
// Only sync entities within view distance
func (w *World) GetRelevantEntities(player *Player) []Entity {
    viewDistance := player.GetViewDistance()
    nearbyEntities := w.spatialIndex.GetWithinRadius(
        player.Position, 
        viewDistance,
    )
    
    // Further filtering based on game logic
    return w.filterByRelevance(player, nearbyEntities)
}
```

### Network Optimization
```go
// Delta compression for position updates
type PositionDelta struct {
    PlayerID  string
    DeltaX    float32 // Using float32 for network efficiency
    DeltaY    float32
    DeltaZ    float32
    Timestamp int64
}

// Batch updates for efficiency
type BatchUpdate struct {
    Positions []PositionDelta
    Actions   []PlayerAction
    Chat      []ChatMessage
}
```

## Monitoring and Observability

### Metrics Collection
```go
// Prometheus metrics
var (
    activeConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "mmorpg_active_connections",
            Help: "Number of active WebSocket connections",
        },
        []string{"region", "server"},
    )
    
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "mmorpg_request_duration_seconds",
            Help: "Request duration in seconds",
        },
        []string{"service", "method"},
    )
)
```

### Health Checks
```go
// Health check endpoint
type HealthChecker struct {
    db    *sql.DB
    redis *redis.Client
    nats  *nats.Conn
}

func (hc *HealthChecker) Check() HealthStatus {
    status := HealthStatus{
        Status: "healthy",
        Checks: make(map[string]CheckResult),
    }
    
    // Check database
    if err := hc.db.Ping(); err != nil {
        status.Status = "unhealthy"
        status.Checks["database"] = CheckResult{
            Status: "down",
            Error:  err.Error(),
        }
    }
    
    // Similar checks for Redis and NATS...
    
    return status
}
```

## Configuration Management

### Environment-based Configuration
```go
// config/config.go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    NATS     NATSConfig
    Security SecurityConfig
    Game     GameConfig
}

func Load() (*Config, error) {
    cfg := &Config{}
    
    // Load from environment with defaults
    cfg.Server.Port = getEnvOrDefault("PORT", "8080")
    cfg.Database.URL = getEnvOrDefault("DATABASE_URL", "postgres://localhost/mmorpg")
    cfg.Redis.URL = getEnvOrDefault("REDIS_URL", "redis://localhost:6379")
    
    // Validate configuration
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    
    return cfg, nil
}
```

## Customer Customization Points

### Extension Interfaces
```go
// Game-specific logic injection
type GameLogic interface {
    OnPlayerJoin(player *Player) error
    OnPlayerLeave(player *Player) error
    OnPlayerAction(player *Player, action Action) error
    CalculateDamage(attacker, target *Character) int
    ValidateItemUse(player *Player, item *Item) bool
}

// Customers implement their own game logic
type CustomGameLogic struct {
    // Custom implementation
}
```

### Event System
```go
// Event-driven architecture for extensibility
type EventBus interface {
    Subscribe(eventType string, handler EventHandler)
    Publish(event Event) error
}

// Customers can subscribe to game events
eventBus.Subscribe("player.levelup", func(e Event) error {
    // Custom logic on level up
    return nil
})
```

This architecture provides a solid foundation that can scale from a single developer testing locally to millions of concurrent players in production, while maintaining clean separation of concerns and easy customization points for game developers.