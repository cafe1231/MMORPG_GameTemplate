# MMORPG Template for Unreal Engine 5.6

<div align="center">

[![Unreal Engine](https://img.shields.io/badge/Unreal%20Engine-5.6-blue?logo=unrealengine)](https://www.unrealengine.com/)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
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
docker-compose up -d
go run cmd/gateway/main.go

# 3. Open Unreal Engine project
# Open MMORPGTemplate/MMORPGTemplate.uproject in UE 5.6

# 4. Test connection (F1 in-game for console)
mmorpg.connect localhost 8090
```

## 🎯 Current Status

### ✅ Phase 0: Foundation (COMPLETE)
- **Infrastructure**: Go microservices with hexagonal architecture
- **Networking**: HTTP/WebSocket client-server communication
- **Serialization**: Protocol Buffers integration (Go + UE5)
- **Development**: Docker environment with hot-reload
- **CI/CD**: GitHub Actions for automated testing
- **Tools**: In-game developer console
- **Error Handling**: Comprehensive error system with retry logic
- **Documentation**: Complete guides and API references

### 🚧 Phase 1: Authentication (NEXT)
- JWT-based authentication
- Account creation and management
- Character system
- Session handling

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
- **Go 1.21+**
- **Docker Desktop**
- **Git**
- **8GB+ RAM** (16GB recommended)

## 🏗️ Project Structure

```
MMORPG_GameTemplate/
├── mmorpg-backend/          # Go microservices
│   ├── cmd/                 # Service entry points
│   ├── internal/            # Business logic
│   ├── pkg/proto/           # Protocol definitions
│   └── deployments/         # Docker/K8s configs
├── MMORPGTemplate/          # Unreal Engine 5.6 Game Project
│   ├── Source/              # C++ game code
│   ├── Content/             # Game assets
│   ├── Config/              # Configuration files
│   └── Plugins/             # Additional plugins
├── docs/                    # Documentation
│   ├── guides/             # How-to guides
│   ├── phases/             # Development phases
│   └── reports/            # Test reports
├── tools/                   # Development utilities
└── .github/                 # CI/CD workflows
```

## 🧪 Testing

```bash
# Backend tests
cd mmorpg-backend
make test

# Connection test
curl http://localhost:8090/api/v1/test

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

### Development Guides
- [Protocol Buffers Integration](docs/guides/PROTOBUF_INTEGRATION.md)
- [Developer Console](docs/guides/DEVELOPER_CONSOLE.md)
- [Error Handling](docs/guides/ERROR_HANDLING.md)
- [CI/CD Pipeline](docs/guides/CI_CD_GUIDE.md)

### Architecture
- [System Design](docs/phases/phase1/PHASE1_DESIGN.md)
- [Requirements](docs/phases/phase1/PHASE1_REQUIREMENTS.md)
- [Development Tasks](docs/phases/phase1/PHASE1_TASKS.md)

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

- [x] Phase 0: Foundation (Complete)
- [ ] Phase 1: Authentication System
- [ ] Phase 2: Real-time Networking
- [ ] Phase 3: Core Gameplay Systems
- [ ] Phase 4: Production Tools
- [ ] Phase 5: Advanced Features

## 📄 License

This is a commercial template. Usage is subject to the license agreement.

---

<div align="center">
Built with ❤️ for game developers who dream big

⭐ Star us on [GitHub](https://github.com/cafe1231/MMORPG_GameTemplate) if you find this helpful!
</div>