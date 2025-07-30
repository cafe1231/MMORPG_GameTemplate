# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a professional MMORPG template combining:
- **Frontend**: Unreal Engine 5.6 client with modular C++ architecture
- **Backend**: Go microservices using hexagonal architecture
- **Communication**: HTTP/WebSocket with Protocol Buffers serialization

The project has completed Phase 0 (Foundation) and Phase 1 (Authentication) with both backend JWT auth and frontend UI widgets fully implemented and tested. Phase 1.5 (Character System) backend is complete, and frontend implementation is in progress.

## Common Commands

### Frontend (Unreal Engine)

**Build the project:**
```bash
# Use batch scripts in scripts/unreal/
./scripts/unreal/BuildProject.bat          # Build the UE5 project
./scripts/unreal/CheckBuildErrors.bat      # Check for build errors
./scripts/unreal/CleanAndRebuild.bat       # Clean rebuild
```

**Test in Unreal Editor:**
1. Open `MMORPGTemplate/MMORPGTemplate.uproject` in UE 5.6
2. Compile the project first (Editor will prompt)
3. Play in Editor to test authentication UI

**In-game console commands:**
```
mmorpg.status        # Check system status
mmorpg.test          # Run connection test
help                 # List all commands
```

### Backend (Go Microservices)

**Start backend services:**
```bash
cd mmorpg-backend
docker-compose -f docker-compose.dev.yml up -d    # Start all services
docker-compose -f docker-compose.dev.yml logs -f  # View logs
docker-compose -f docker-compose.dev.yml down     # Stop services
```

**Backend development:**
```bash
cd mmorpg-backend
make test           # Run all tests
make lint           # Run linter
make fmt            # Format code
make build          # Build all services
make build-auth     # Build specific service
make coverage       # Generate coverage report
make help           # See all commands
```

**Test backend endpoints:**
```bash
# Health check
curl http://localhost:8090/api/v1/test

# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"testuser","password":"Password123!","accept_terms":true}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Password123!"}'
```

**Database access:**
- Adminer UI: http://localhost:8091
- Credentials: Server: localhost, User: dev, Password: dev, Database: mmorpg

## High-Level Architecture

### Frontend Architecture (Unreal Engine)

The UE5 client uses a modular architecture with 4 C++ modules:

1. **MMORPGCore** - Foundation layer
   - `UMMORPGAuthSubsystem`: JWT token management, auto-refresh, auth state
   - `UMMORPGCharacterSubsystem`: Character management, creation, selection, caching
   - Core types and interfaces shared across modules
   - Blueprint-exposed authentication types (FAuthResponse, FLoginRequest, etc.)
   - Character data types (FCharacterInfo, FCharacterCreateRequest, etc.)

2. **MMORPGNetwork** - Networking layer
   - HTTP client with retry logic and error handling
   - Protocol Buffers integration for serialization
   - Request/response handling infrastructure

3. **MMORPGUI** - UI framework
   - `UMMORPGAuthWidget`: Main auth UI with widget switcher
   - `UMMORPGLoginWidget`: Login form with validation
   - `UMMORPGRegisterWidget`: Registration with terms acceptance
   - `UMMORPGCharacterCreateWidget`: Character creation with appearance customization
   - All widgets are Blueprint-extendable

4. **MMORPGTemplate** - Main game module
   - `AMMORPGAuthGameMode`: Spawns auth UI on start
   - `AMMORPGAuthPlayerController`: Handles input for auth screens
   - Game-specific implementations

### Backend Architecture (Go Microservices)

The backend follows hexagonal architecture with clean separation:

**Services:**
- **Gateway Service** (port 8090): API gateway, routes requests to microservices
- **Auth Service** (port 8080): JWT auth, user management, session handling

**Infrastructure:**
- PostgreSQL: User data, persistent storage
- Redis: Session management, caching
- NATS: Event-driven communication between services
- Docker Compose: Local development environment

**Code Organization:**
```
mmorpg-backend/
├── cmd/           # Service entry points
├── internal/      # Business logic (hexagonal architecture)
│   ├── core/      # Domain entities and ports
│   ├── adapters/  # Infrastructure implementations
│   └── services/  # Application services
└── pkg/proto/     # Protocol Buffers definitions
```

### Key Design Patterns

1. **Authentication Flow**:
   - Client sends login/register request to Gateway
   - Gateway forwards to Auth Service
   - Auth Service validates, creates JWT tokens
   - Tokens stored in UE5 AuthSubsystem
   - Auto-refresh on startup if valid refresh token exists

2. **Error Handling**:
   - Structured error responses with codes
   - Client-side retry logic for transient failures
   - UI feedback for validation errors

3. **Blueprint Integration**:
   - All core types exposed to Blueprint
   - Event delegates for async operations
   - Widget base classes handle C++ complexity

4. **Development Workflow**:
   - Hot-reload enabled for Go services
   - Mock mode in UE5 for testing without backend
   - Docker Compose for one-command setup

## Key Files to Know

**Frontend:**
- `MMORPGTemplate/Source/MMORPGCore/Public/Auth/MMORPGAuthSubsystem.h` - Auth state management
- `MMORPGTemplate/Source/MMORPGCore/Public/Subsystems/UMMORPGCharacterSubsystem.h` - Character management
- `MMORPGTemplate/Source/MMORPGCore/Public/Types/FCharacterTypes.h` - Character data types
- `MMORPGTemplate/Source/MMORPGUI/Public/Widgets/Auth/MMORPGAuthWidget.h` - Main auth UI
- `MMORPGTemplate/Source/MMORPGUI/Public/Character/UMMORPGCharacterCreateWidget.h` - Character creation UI
- `MMORPGTemplate/Source/MMORPGTemplate/Public/GameModes/MMORPGAuthGameMode.h` - Auth game mode

**Backend:**
- `mmorpg-backend/internal/auth/core/ports/auth.go` - Auth service interface
- `mmorpg-backend/internal/auth/adapters/http/handlers.go` - HTTP handlers
- `mmorpg-backend/cmd/gateway/main.go` - Gateway entry point
- `mmorpg-backend/config.yaml` - Service configuration

**Documentation:**
- `docs/phases/phase1/PHASE1B_QUICKSTART.md` - Frontend setup guide
- `docs/phases/phase1/PHASE1B_REAL_AUTH_TEST_GUIDE.md` - Testing guide
- `docs/phases/phase1_5/PHASE1_5_FRONTEND_PROGRESS.md` - Character system progress
- `docs/architecture/ARCHITECTURE.md` - System design