# MMORPG Template for Unreal Engine 5.6

A professional, scalable MMORPG template designed to grow from local development to millions of concurrent players.

## 🚀 Quick Start

```bash
# 1. Start backend services
cd mmorpg-backend
docker-compose up -d
go run cmd/gateway/main.go

# 2. Open Unreal project
# Open UnrealEngine/MMORPGTemplate.uproject in UE 5.6
```

## 📚 Documentation

All documentation is organized in the [`docs/`](docs/) directory:

- **[Getting Started](docs/guides/QUICKSTART.md)** - Quick start guide
- **[Development Setup](docs/guides/DEVELOPMENT_SETUP.md)** - Complete setup instructions
- **[Documentation Index](docs/README.md)** - Full documentation structure

## 🏗️ Project Structure

```
Plugin_mmorpg/
├── mmorpg-backend/          # Go microservices backend
├── UnrealEngine/            # UE5.6 plugin and example project
├── docs/                    # All documentation
│   ├── phases/             # Development phase docs
│   ├── guides/             # How-to guides
│   └── reports/            # Test reports
├── tools/                   # Development tools
└── .github/                 # CI/CD workflows
```

## 🎯 Current Status

**Phase 0: Foundation** ✅ COMPLETE
- Core infrastructure
- Development environment
- Basic client-server connection
- Protocol Buffer integration
- Developer console
- Error handling framework

**Phase 1: Authentication** 🚧 NEXT
- JWT authentication
- Account management
- Character creation
- Session handling

## 🛠️ Key Features

- **Scalable Architecture** - From 1 to 1M+ concurrent players
- **Protocol Buffers** - Efficient binary serialization
- **Hexagonal Architecture** - Clean, maintainable backend
- **Blueprint Support** - Full UE5 Blueprint integration
- **Developer Tools** - In-game console, error handling, monitoring
- **Production Ready** - Docker, Kubernetes, CI/CD pipelines

## 📖 Learn More

- [Architecture Overview](docs/phases/phase1/PHASE1_DESIGN.md)
- [API Documentation](docs/guides/PROTOBUF_INTEGRATION.md)
- [Error Handling](docs/guides/ERROR_HANDLING.md)
- [Developer Console](docs/guides/DEVELOPER_CONSOLE.md)

## 🤝 Contributing

See [Git Workflow](docs/guides/CI_CD_GUIDE.md) for contribution guidelines.

## 📄 License

This is a commercial template. See LICENSE for details.

---

Built with ❤️ for game developers who dream big.