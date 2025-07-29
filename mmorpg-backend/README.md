# MMORPG Backend

## Overview

This is the Go backend for the MMORPG template, implementing a microservices architecture with hexagonal design patterns. The backend is designed to scale from local development to millions of concurrent players.

## Architecture

The backend follows hexagonal architecture (Ports & Adapters) principles:

```
mmorpg-backend/
├── cmd/                    # Service entry points
│   ├── gateway/           # API Gateway service
│   ├── auth/             # Authentication service
│   ├── character/        # Character management service
│   ├── world/           # World/real-time service
│   └── game/            # Game logic service
├── internal/              # Private application code
│   ├── domain/           # Business logic (pure Go)
│   │   ├── auth/        # Authentication domain
│   │   ├── character/   # Character domain
│   │   ├── world/       # World domain
│   │   └── game/        # Game domain
│   ├── adapters/        # External interfaces
│   │   ├── postgres/    # PostgreSQL adapter
│   │   ├── redis/       # Redis adapter
│   │   ├── websocket/   # WebSocket adapter
│   │   └── nats/        # NATS adapter
│   └── ports/           # Interface definitions
├── pkg/                   # Public packages
│   ├── proto/           # Protocol Buffer definitions
│   ├── middleware/      # HTTP/gRPC middleware
│   ├── logger/          # Structured logging
│   └── metrics/         # Prometheus metrics
├── deployments/          # Deployment configurations
│   ├── docker/          # Docker files
│   ├── kubernetes/      # K8s manifests
│   └── terraform/       # Infrastructure as code
├── scripts/              # Utility scripts
├── migrations/           # Database migrations
├── go.mod               # Go module definition
├── go.sum               # Go dependencies lock
├── Makefile             # Build automation
└── README.md            # This file
```

## Services

### Gateway Service
- **Port**: 8080
- **Purpose**: API gateway, load balancing, rate limiting
- **Features**:
  - WebSocket connection management
  - Protocol version negotiation
  - Request routing
  - Authentication validation

### Auth Service
- **Port**: 8081
- **Purpose**: Authentication and authorization
- **Features**:
  - JWT token generation/validation
  - User registration/login
  - Session management
  - Password hashing (bcrypt)

### Character Service
- **Port**: 8082
- **Purpose**: Character management with hexagonal architecture
- **Features**:
  - Character CRUD operations with soft delete
  - Appearance and stats management
  - Position tracking
  - Redis caching for performance
  - NATS event publishing
  - JWT authentication
  - 30-day recovery for deleted characters
- **API Endpoints**:
  - POST `/api/v1/characters` - Create character
  - GET `/api/v1/characters` - List characters
  - GET `/api/v1/characters/{id}` - Get character
  - PUT `/api/v1/characters/{id}` - Update character
  - DELETE `/api/v1/characters/{id}` - Soft delete
  - POST `/api/v1/characters/{id}/select` - Select character

### World Service
- **Port**: 8083
- **Purpose**: Real-time world state
- **Features**:
  - Player position tracking
  - Spatial indexing (Octree)
  - Interest management
  - Entity synchronization

### Game Service
- **Port**: 8084
- **Purpose**: Game logic processing
- **Features**:
  - Inventory management
  - Quest system
  - Combat calculations
  - NPC interactions

## Database Migrations

The project uses SQL migrations for database schema management:

### Current Migrations
- **001-003**: Core tables (users, sessions)
- **004-009**: Character system tables
  - `004_create_characters_table.sql` - Core character data
  - `005_create_character_appearance_table.sql` - Visual customization
  - `006_create_character_stats_table.sql` - RPG statistics
  - `007_create_character_position_table.sql` - World location
  - `008_create_character_initialization_triggers.sql` - Auto-initialization
  - `009_create_character_performance_indexes.sql` - Query optimization

### Running Migrations
```bash
# Run all migrations
make migrate-up

# Rollback last migration
make migrate-down

# Reset database
make migrate-reset
```

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Docker and Docker Compose
- Make (optional but recommended)
- Protocol Buffers compiler (protoc)

### Development Setup

1. Clone the repository:
```bash
cd mmorpg-backend
```

2. Install dependencies:
```bash
make deps
```

3. Start the development environment:
```bash
make dev-setup
```

This will:
- Start PostgreSQL, Redis, and NATS in Docker
- Run database migrations
- Set up the development environment

4. Build all services:
```bash
make build
```

5. Run a specific service:
```bash
make run-gateway
# or
./bin/gateway
```

## Makefile Commands

- `make all` - Clean, format, lint, test, and build
- `make build` - Build all services
- `make build-<service>` - Build specific service
- `make test` - Run tests
- `make coverage` - Generate test coverage report
- `make fmt` - Format code
- `make lint` - Run linter
- `make proto` - Generate protobuf code
- `make docker-up` - Start Docker environment
- `make docker-down` - Stop Docker environment
- `make help` - Show all available commands

## Configuration

Configuration is managed through:
1. Environment variables (prefixed with `MMORPG_`)
2. Configuration files (`config.yaml`)
3. Command-line flags

Example environment variables:
```bash
MMORPG_SERVER_PORT=8080
MMORPG_DATABASE_URL=postgres://user:pass@localhost:5432/mmorpg
MMORPG_REDIS_URL=redis://localhost:6379
MMORPG_NATS_URL=nats://localhost:4222
MMORPG_SECURITY_JWT_SECRET=your-secret-key
```

## Development

### Running Tests
```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run benchmarks
make bench
```

### Code Style
The project uses standard Go formatting and linting rules:
```bash
# Format code
make fmt

# Run linter
make lint

# Run security scan
make security
```

### Adding a New Service

1. Create service directory:
```bash
mkdir -p cmd/newservice
```

2. Create main.go following the existing pattern
3. Add service to Makefile SERVICES variable
4. Update docker-compose.yml if needed
5. Document the service in this README

## Monitoring

The backend exposes Prometheus metrics on port 9090 for each service:
- `http://localhost:9090/metrics` - Gateway metrics
- `http://localhost:9091/metrics` - Auth metrics
- etc.

Key metrics:
- `mmorpg_active_connections` - Active WebSocket connections
- `mmorpg_request_duration_seconds` - Request latency
- `mmorpg_players_online` - Online player count
- `mmorpg_database_duration_seconds` - Database query performance

## Deployment

### Docker
```bash
# Build Docker images
make docker-build

# Run with Docker Compose
make docker-up
```

### Kubernetes
```bash
# Apply Kubernetes manifests
kubectl apply -f deployments/kubernetes/
```

### Production Considerations
- Use environment-specific configuration
- Enable TLS for all communications
- Set up proper monitoring and alerting
- Configure autoscaling policies
- Implement proper secret management
- Set up database backups

## Troubleshooting

### Service won't start
- Check if ports are already in use
- Verify database connection
- Check logs: `docker-compose logs <service>`

### Protocol Buffer issues
- Ensure protoc is installed
- Regenerate: `make proto`

### Database connection errors
- Verify PostgreSQL is running: `docker ps`
- Check connection string in environment

## Contributing

1. Follow the coding standards
2. Write tests for new features
3. Update documentation
4. Run `make all` before committing
5. Create descriptive commit messages

## License

Copyright (c) 2024 MMORPG Template Project