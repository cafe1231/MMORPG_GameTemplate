# MMORPG Template - System Architecture

## Overview

The MMORPG Template is built on a modern, scalable architecture designed to support everything from solo development to massive multiplayer deployments. The system uses a microservices backend written in Go and a modular Unreal Engine 5.6 client.

## Architecture Principles

1. **Separation of Concerns**: Clear boundaries between different system components
2. **Scalability**: Horizontal scaling capabilities built-in from the start
3. **Modularity**: Components can be developed, tested, and deployed independently
4. **Type Safety**: Protocol Buffers ensure type-safe communication
5. **Developer Experience**: Tools and patterns that accelerate development

## System Components

### Backend Architecture (Go)

The backend follows a hexagonal (ports and adapters) architecture:

```
mmorpg-backend/
├── cmd/                    # Service entry points
│   ├── gateway/           # API Gateway service
│   ├── auth/              # Authentication service
│   ├── character/         # Character management
│   ├── game/              # Game logic service
│   ├── chat/              # Chat service
│   └── world/             # World state service
├── internal/              # Business logic
│   ├── domain/            # Core business entities
│   ├── ports/             # Interface definitions
│   └── adapters/          # External integrations
└── pkg/                   # Shared packages
    └── proto/             # Protocol Buffer definitions
```

**Key Features:**
- Microservices architecture with clear service boundaries
- NATS for inter-service communication
- PostgreSQL for persistent storage
- Redis for caching and session management
- Docker-based development environment
- Comprehensive error handling and logging

### Client Architecture (Unreal Engine 5.6)

The client uses a modular C++ architecture:

```
MMORPGTemplate/Source/
├── MMORPGCore/            # Foundation layer
├── MMORPGProto/           # Protocol Buffer integration
├── MMORPGNetwork/         # Networking layer
├── MMORPGUI/              # UI framework
└── MMORPGTemplate/        # Main game module
```

See [CLIENT_MODULES.md](CLIENT_MODULES.md) for detailed client architecture.

## Communication Flow

### Client-Server Communication

```
[UE5 Client] <---> [Gateway Service] <---> [Backend Services]
     |                    |                       |
     |                    |                    [NATS]
     |                    |                       |
  HTTP/WS            Load Balancer          Message Bus
```

1. **HTTP**: Used for request-response operations (login, character creation)
2. **WebSocket**: Used for real-time updates (movement, chat, combat)
3. **Protocol Buffers**: All messages are serialized using protobuf

### Service Communication

Backend services communicate via NATS messaging:

```
[Gateway] --> NATS --> [Auth Service]
                   --> [Character Service]
                   --> [Game Service]
                   --> [Chat Service]
                   --> [World Service]
```

## Data Flow

### Authentication Flow
```
1. Client -> Gateway: Login Request (HTTP)
2. Gateway -> Auth Service: Validate Credentials (NATS)
3. Auth Service -> Database: Check User
4. Auth Service -> Redis: Store Session
5. Auth Service -> Gateway: JWT Token
6. Gateway -> Client: Login Response
```

### Game State Flow
```
1. Client -> Gateway: Movement Update (WebSocket)
2. Gateway -> Game Service: Process Movement (NATS)
3. Game Service -> World Service: Update Position
4. World Service -> Gateway: Broadcast Update
5. Gateway -> All Clients: Position Update
```

## Scalability Design

### Horizontal Scaling

Each component can be scaled independently:

- **Gateway**: Multiple instances behind load balancer
- **Services**: Multiple instances per service type
- **Database**: Read replicas and sharding
- **Cache**: Redis cluster
- **Message Bus**: NATS cluster

### Regional Distribution

```
Region A                    Region B
├── Gateway Cluster        ├── Gateway Cluster
├── Service Cluster        ├── Service Cluster
├── Database Master        ├── Database Replica
└── Redis Cluster          └── Redis Cluster
         |                          |
         └──── Global NATS Bus ────┘
```

## Security Architecture

### Network Security
- TLS/SSL for all external communications
- Service mesh for internal communications
- API rate limiting and DDoS protection

### Application Security
- JWT tokens for authentication
- Role-based access control (RBAC)
- Input validation at all entry points
- SQL injection prevention
- XSS protection for web components

## Development vs Production

### Development Environment
```
Docker Compose
├── PostgreSQL (single instance)
├── Redis (single instance)
├── NATS (single instance)
└── All services (hot-reload enabled)
```

### Production Environment
```
Kubernetes Cluster
├── Gateway Pods (auto-scaling)
├── Service Pods (auto-scaling)
├── PostgreSQL (managed service)
├── Redis Cluster
├── NATS Cluster
├── Monitoring Stack
└── Load Balancers
```

## Monitoring and Observability

### Metrics
- Prometheus for metrics collection
- Grafana for visualization
- Custom game-specific metrics

### Logging
- Structured logging (JSON format)
- Centralized log aggregation
- Log levels: Debug, Info, Warn, Error

### Tracing
- OpenTelemetry integration
- Distributed tracing across services
- Performance bottleneck identification

## Technology Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Standard library + minimal dependencies
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Message Bus**: NATS 2.9+
- **Serialization**: Protocol Buffers 3
- **Container**: Docker 20+
- **Orchestration**: Kubernetes 1.25+

### Client
- **Engine**: Unreal Engine 5.6
- **Language**: C++ 17
- **Build**: Unreal Build Tool
- **Modules**: Custom modular architecture
- **UI**: UMG (Unreal Motion Graphics)
- **Networking**: Built on UE5 networking layer

## Future Considerations

### Phase 2-3 Additions
- Real-time combat system
- Spatial indexing for proximity queries
- Voice chat integration
- Advanced anti-cheat systems

### Scaling to Millions
- Database sharding strategy
- CDN for asset delivery
- Edge computing for reduced latency
- Global state synchronization

## Related Documentation

- [CLIENT_MODULES.md](CLIENT_MODULES.md) - Detailed client architecture
- [Phase 1 Design](../phases/phase1/PHASE1_DESIGN.md) - Authentication system design
- [Protocol Buffers Guide](../guides/PROTOBUF_INTEGRATION.md) - Message format details
- [Development Setup](../guides/DEVELOPMENT_SETUP.md) - Getting started guide