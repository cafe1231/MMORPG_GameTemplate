# MMORPG Template for Unreal Engine 5.6

<div align="center">

[![Unreal Engine](https://img.shields.io/badge/Unreal%20Engine-5.6-blue?logo=unrealengine)](https://www.unrealengine.com/)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://golang.org/)
[![Protocol Buffers](https://img.shields.io/badge/Protocol%20Buffers-3.0-green)](https://protobuf.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-Commercial-red)](LICENSE)

A professional, production-ready MMORPG template that scales from local development to millions of concurrent players.

[Documentation](docs/) • [Quick Start](docs/guides/QUICKSTART.md) • [Architecture](docs/phases/phase1/PHASE1_DESIGN.md) • [GitHub](https://github.com/cafe1231/MMORPG_GameTemplate)

</div>

## 🚀 Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/cafe1231/MMORPG_GameTemplate.git
cd MMORPG_GameTemplate

# 2. Start backend services
cd mmorpg-backend
docker-compose -f docker-compose.dev.yml up -d
# All services start automatically including Gateway and Auth

# 3. Open Unreal Engine project
# Open MMORPGTemplate/MMORPGTemplate.uproject in UE 5.6
# Compile the project first

# 4. Play in editor and create an account/login!
# The authentication system is fully functional
```

## 🎯 Current Status

### ✅ Phase 0: Foundation (100% COMPLETE)
- **Infrastructure**: Go microservices with hexagonal architecture
- **Networking**: HTTP/WebSocket client-server communication
- **Serialization**: Protocol Buffers integration (Go + UE5)
- **Development**: Docker environment with hot-reload
- **CI/CD**: GitHub Actions for automated testing
- **Tools**: In-game developer console
- **Error Handling**: Comprehensive error system with retry logic
- **Documentation**: Complete guides and API references

### ✅ Phase 1: Authentication (100% COMPLETE)

#### Phase 1A - Backend (✅ COMPLETE)
- JWT-based authentication with access/refresh tokens
- User registration and login endpoints
- Session management with PostgreSQL + Redis
- Password hashing with bcrypt
- NATS event publishing

#### Phase 1B - Frontend (✅ COMPLETE - Fully Tested!)
- Login/Register UI widgets with UMG
- UMMORPGAuthSubsystem with JWT token management
- Blueprint-friendly authentication types
- Widget Switcher navigation between views
- Full integration with backend API
- Auto-token refresh on startup
- Error handling and validation
- Game mode and player controller setup
- Accept terms checkbox implementation
- JSON parsing for multiple response formats
- Rate limiting handling

### 📋 Phase 1.5: Character System (PLANNED - Next Phase)
- Character creation and customization
- Multiple characters per account
- Character selection after login
- Persistent character data storage
- 3D preview system
- Character slots management
- Name validation and uniqueness
- Class and race selection

## 🛠️ Key Features

### For Solo Developers
- **One-command setup** - Get running in < 10 minutes
- **Blueprint-friendly** - Full Blueprint API exposure
- **Built-in debugging** - Developer console and error tracking
- **Local development** - Everything runs on your machine

### For Studios
- **Production architecture** - Battle-tested patterns
- **Horizontal scaling** - Microservices that scale independently
- **Multi-region ready** - Deploy globally with ease
- **Enterprise patterns** - SOLID principles, clean architecture

### Technical Highlights
- **Protocol Buffers** - Type-safe, efficient serialization
- **Hexagonal Architecture** - Clean separation of concerns
- **Event-driven** - NATS messaging for service communication
- **Observable** - Prometheus metrics + Grafana dashboards
- **Kubernetes-ready** - Helm charts included

## 📋 Prerequisites

- **Unreal Engine 5.6+**
- **Visual Studio 2022** (Windows) or Xcode 14+ (macOS)
- **Go 1.23+**
- **Docker Desktop**
- **Git**
- **8GB+ RAM** (16GB recommended)

## 🏗️ Project Structure

```
MMORPG_GameTemplate/
├── mmorpg-backend/              # Go microservices backend (✅ Complete)
│   ├── cmd/                     # Service entry points
│   │   ├── gateway/             # API Gateway service
│   │   └── auth/                # Authentication service (✅ Phase 1A)
│   ├── internal/                # Business logic (hexagonal architecture)
│   ├── pkg/proto/               # Protocol Buffer definitions
│   └── deployments/             # Docker/K8s configurations
├── MMORPGTemplate/              # Unreal Engine 5.6 client (✅ Phase 1 Complete)
│   ├── Source/                  # C++ source code
│   │   ├── MMORPGTemplate/      # Main game module
│   │   ├── MMORPGCore/          # Core systems module (✅ Auth subsystem)
│   │   ├── MMORPGNetwork/       # Networking module (✅ HTTP client)
│   │   └── MMORPGUI/            # UI module (✅ Auth widgets)
│   ├── Content/                 # Game assets
│   ├── Config/                  # Configuration files
│   └── MMORPGTemplate.uproject  # Project file
├── docs/                        # Comprehensive documentation
│   ├── guides/                  # Development guides
│   ├── phases/                  # Phase tracking
│   │   ├── phase0/              # Foundation
│   │   ├── phase1/              # Authentication
│   │   ├── phase1_5/            # Character System (NEW)
│   │   ├── phase2/              # Networking (Split)
│   │   ├── phase3/              # Gameplay
│   │   └── phase4/              # Production
│   ├── architecture/            # System design
│   ├── project/                 # Project management
│   └── reports/                 # Test reports
├── scripts/                     # Helper scripts
│   ├── db/                      # Database scripts
│   └── unreal/                  # Build scripts
├── tools/                       # Development utilities
└── .github/                     # CI/CD workflows
```

## 🧪 Testing

```bash
# Backend tests
cd mmorpg-backend
make test

# Connection test
curl http://localhost:8090/api/v1/test

# Authentication API tests
# Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"testuser","password":"Password123!","accept_terms":true}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Password123!"}'

# Test in Unreal Engine
# 1. Launch the game in editor
# 2. Create account with UI
# 3. Login and enjoy!

# View data in database (Adminer)
# http://localhost:8091
# Server: localhost, User: dev, Password: dev, Database: mmorpg

# In-game console commands
mmorpg.status        # Check system status
mmorpg.test          # Run connection test
help                 # List all commands
```

## 📊 Performance Targets

| Scale | Players | Infrastructure | Monthly Cost |
|-------|---------|----------------|--------------|
| Dev | 1-10 | Local Docker | $0 |
| Small | 100-1K | 3-5 servers | ~$100 |
| Medium | 1K-10K | 10-50 servers | ~$1,000 |
| Large | 10K-100K | 50+ servers | ~$10,000 |
| Massive | 100K-1M+ | Multi-region | $10,000+ |

## 📚 Documentation

### Getting Started
- [Quick Start Guide](docs/guides/QUICKSTART.md)
- [Development Setup](docs/guides/DEVELOPMENT_SETUP.md)
- [Phase 0 Summary](docs/phases/phase0/PHASE0_SUMMARY.md)
- [Phase 1A Completion Report](docs/phases/phase1/PHASE1A_COMPLETION_REPORT.md)
- [Phase 1B Completion Report](docs/phases/phase1/PHASE1B_COMPLETION_REPORT.md)
- [Phase 1B Quick Start](docs/phases/phase1/PHASE1B_QUICKSTART.md)
- [Phase 1B Implementation Summary](docs/phases/phase1/PHASE1B_IMPLEMENTATION_SUMMARY.md)
- [Phase 1.5 Character System](docs/phases/phase1_5/PHASE1_5_SUMMARY.md)

### Development Guides
- [Protocol Buffers Integration](docs/guides/PROTOBUF_INTEGRATION.md)
- [Developer Console](docs/guides/DEVELOPER_CONSOLE.md)
- [Error Handling](docs/guides/ERROR_HANDLING.md)
- [CI/CD Pipeline](docs/guides/CI_CD_GUIDE.md)
- [Database Migration Strategy](docs/architecture/DATABASE_MIGRATION_STRATEGY.md)

### Architecture
- [System Design](docs/phases/phase1/PHASE1_DESIGN.md)
- [Requirements](docs/phases/phase1/PHASE1_REQUIREMENTS.md)
- [Development Tasks](docs/phases/phase1/PHASE1_TASKS.md)
- [Phase 1.5 Overview](docs/phases/phase1_5/PHASE1_5_OVERVIEW.md)
- [Phase 2 Restructuring](docs/phases/phase2/PHASE2_RESTRUCTURING.md)
- [Project Timeline](docs/project/REVISED_PROJECT_TIMELINE.md)

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

See [Git Workflow Guide](docs/guides/CI_CD_GUIDE.md) for detailed contribution guidelines.

## 🐛 Known Issues

- Console widget needs to be created manually in UE5 (see [Developer Console Guide](docs/guides/DEVELOPER_CONSOLE.md))
- Windows line endings warnings during git operations (normal, handled by .gitattributes)

## 📞 Support

- **Discord**: [Join our community](#) (coming soon)
- **Issues**: [GitHub Issues](https://github.com/cafe1231/MMORPG_GameTemplate/issues)
- **Email**: support@example.com (coming soon)

## 🚀 Roadmap

- [x] Phase 0: Foundation (✅ Complete)
- [x] Phase 1: Authentication System (✅ Complete)
  - [x] Phase 1A: Backend Auth (JWT, Login, Register)
  - [x] Phase 1B: Frontend Integration (UI, Auth Subsystem)
- [ ] Phase 1.5: Character System Foundation (🚧 Next - 3-4 weeks)
- [ ] Phase 2: Real-time Networking (Split into 2A & 2B)
  - [ ] Phase 2A: Core Infrastructure (3-4 weeks)
  - [ ] Phase 2B: Advanced Features (3-4 weeks)
- [ ] Phase 3: Core Gameplay Systems (8-10 weeks)
- [ ] Phase 4: Production & Polish (4-6 weeks)

**Total Timeline**: 28-38 weeks (7-9 months)

## 📄 License

This is a commercial template. Usage is subject to the license agreement.

---

<div align="center">
Built with ❤️ for game developers who dream big

⭐ Star us on [GitHub](https://github.com/cafe1231/MMORPG_GameTemplate) if you find this helpful!
</div>