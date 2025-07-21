# MMORPG Template Development Setup Guide

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Environment Setup](#environment-setup)
3. [Backend Setup](#backend-setup)
4. [Unreal Engine Setup](#unreal-engine-setup)
5. [Running the Project](#running-the-project)
6. [Development Workflow](#development-workflow)
7. [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Software

#### Windows
- **Windows 10/11** (64-bit)
- **Visual Studio 2022** (Community or higher)
  - Workloads: "Game development with C++" and ".NET desktop development"
  - Individual components: "Windows 10 SDK" and "MSVC v143"
- **Unreal Engine 5.6**
- **Git** (latest version)
- **Docker Desktop for Windows**
- **Go 1.21+**
- **Protocol Buffers Compiler (protoc) 25.1+**
- **Node.js 18+** (for tooling)

#### macOS
- **macOS 12+** (Monterey or later)
- **Xcode 14+**
- **Unreal Engine 5.6**
- **Git** (via Xcode or Homebrew)
- **Docker Desktop for Mac**
- **Go 1.21+**
- **Protocol Buffers Compiler (protoc) 25.1+**
- **Node.js 18+**

#### Linux
- **Ubuntu 22.04 LTS** (or compatible distribution)
- **GCC 11+** or **Clang 14+**
- **Unreal Engine 5.6** (built from source)
- **Git**
- **Docker** and **Docker Compose**
- **Go 1.21+**
- **Protocol Buffers Compiler (protoc) 25.1+**
- **Node.js 18+**

### Hardware Requirements
- **CPU**: 8+ cores recommended
- **RAM**: 32GB minimum, 64GB recommended
- **GPU**: DirectX 12 or Vulkan compatible
- **Storage**: 150GB+ free space (SSD recommended)

## Environment Setup

### 1. Install Git
```bash
# Windows (via Chocolatey)
choco install git

# macOS (via Homebrew)
brew install git

# Linux
sudo apt-get install git
```

### 2. Clone the Repository
```bash
git clone https://github.com/your-org/mmorpg-template.git
cd mmorpg-template
```

### 3. Install Go
```bash
# Windows
choco install golang

# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```

### 4. Install Protocol Buffers
```bash
# Windows
choco install protoc

# macOS
brew install protobuf

# Linux
curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-x86_64.zip
sudo unzip protoc-25.1-linux-x86_64.zip -d /usr/local
```

### 5. Install Docker
Follow the official installation guides:
- [Docker Desktop for Windows](https://docs.docker.com/desktop/install/windows-install/)
- [Docker Desktop for Mac](https://docs.docker.com/desktop/install/mac-install/)
- [Docker Engine for Linux](https://docs.docker.com/engine/install/)

## Backend Setup

### 1. Navigate to Backend Directory
```bash
cd mmorpg-backend
```

### 2. Install Go Dependencies
```bash
go mod download
go mod verify
```

### 3. Install Development Tools
```bash
# Protocol buffer plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Linting tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Migration tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 4. Start Infrastructure Services
```bash
# Start PostgreSQL, Redis, and NATS
docker-compose up -d

# Verify services are running
docker-compose ps
```

### 5. Setup Database
```bash
# Create database
docker exec -it mmorpg-postgres psql -U mmorpg -c "CREATE DATABASE mmorpg_dev;"

# Run migrations (when available)
# migrate -path migrations -database "postgres://mmorpg:mmorpg123@localhost:5432/mmorpg_dev?sslmode=disable" up
```

### 6. Generate Protocol Buffers
```bash
make proto
# or manually:
./scripts/generate_proto.sh
```

### 7. Build Backend Services
```bash
# Build all services
make build

# Or build individually
go build -o bin/gateway cmd/gateway/main.go
go build -o bin/auth cmd/auth/main.go
go build -o bin/character cmd/character/main.go
go build -o bin/game cmd/game/main.go
go build -o bin/chat cmd/chat/main.go
go build -o bin/world cmd/world/main.go
```

## Unreal Engine Setup

### 1. Install Unreal Engine 5.6
- Download and install from [Epic Games Launcher](https://www.unrealengine.com/download)
- Or build from source for Linux

### 2. Generate Project Files
```bash
# Windows
cd UnrealEngine
"C:\Program Files\Epic Games\UE_5.6\Engine\Build\BatchFiles\GenerateProjectFiles.bat" -projectfiles -project="%cd%\MMORPGTemplate.uproject" -game -rocket -progress

# macOS/Linux
cd UnrealEngine
"/Users/Shared/Epic Games/UE_5.6/Engine/Build/BatchFiles/Mac/GenerateProjectFiles.sh" -projectfiles -project="$(pwd)/MMORPGTemplate.uproject" -game -rocket -progress
```

### 3. Open in IDE
- **Windows**: Open `MMORPGTemplate.sln` in Visual Studio 2022
- **macOS**: Open `MMORPGTemplate.xcworkspace` in Xcode
- **Linux**: Use your preferred IDE with CMake support

### 4. Build the Plugin
```bash
# Windows (from VS Developer Command Prompt)
msbuild MMORPGTemplate.sln /p:Configuration=Development /p:Platform=Win64

# macOS/Linux
make MMORPGTemplateEditor
```

### 5. Configure Plugin Settings
Create `UnrealEngine/Config/DefaultMMORPG.ini`:
```ini
[/Script/MMORPGCore.MMORPGSettings]
DefaultServerHost=localhost
DefaultServerPort=8090
ConnectionTimeout=30.0
EnableDebugLogging=true
```

## Running the Project

### 1. Start Backend Services
```bash
# Terminal 1: Infrastructure
cd mmorpg-backend
docker-compose up

# Terminal 2: Gateway Service
cd mmorpg-backend
go run cmd/gateway/main.go

# Terminal 3: Auth Service (if needed)
cd mmorpg-backend
go run cmd/auth/main.go
```

### 2. Launch Unreal Editor
- Open `UnrealEngine/MMORPGTemplate.uproject`
- Wait for shaders to compile
- Open the test map: `Content/MMORPG/Maps/TestMap`

### 3. Test Connection
- Place `BP_ConnectionTest` actor in the level
- Play in Editor (PIE)
- Check Output Log for connection status

## Development Workflow

### 1. Code Style
```bash
# Format Go code
cd mmorpg-backend
go fmt ./...
golangci-lint run

# Format C++ code (requires clang-format)
cd UnrealEngine
find . -name "*.cpp" -o -name "*.h" | xargs clang-format -i
```

### 2. Running Tests
```bash
# Go tests
cd mmorpg-backend
go test ./...
go test -race ./...
go test -cover ./...

# Unreal tests
# Run from Editor: Window > Test Automation
```

### 3. Hot Reload
- **Go**: Services restart automatically with `air` or `reflex`
- **Unreal**: Use Live Coding (Ctrl+Alt+F11)

### 4. Debugging

#### Backend Debugging
```bash
# VS Code launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Gateway",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/mmorpg-backend/cmd/gateway/main.go",
            "env": {
                "CONFIG_PATH": "./config/development.yaml"
            }
        }
    ]
}
```

#### Unreal Debugging
1. Set breakpoints in Visual Studio/Xcode
2. Launch Editor with debugger attached
3. Use Visual Studio diagnostic tools

### 5. Git Workflow
```bash
# Create feature branch
git checkout -b feature/your-feature

# Make changes and commit
git add .
git commit -m "feat: add new feature"

# Push and create PR
git push -u origin feature/your-feature
```

## Troubleshooting

### Common Issues

#### 1. Docker Issues
```bash
# Reset Docker environment
docker-compose down -v
docker system prune -a
docker-compose up --build
```

#### 2. Go Module Issues
```bash
# Clear module cache
go clean -modcache
go mod download
go mod tidy
```

#### 3. Unreal Build Errors
```bash
# Clean and rebuild
# Windows
"C:\Program Files\Epic Games\UE_5.6\Engine\Build\BatchFiles\Clean.bat" MMORPGTemplateEditor Win64 Development
"C:\Program Files\Epic Games\UE_5.6\Engine\Build\BatchFiles\Build.bat" MMORPGTemplateEditor Win64 Development

# Regenerate project files if needed
```

#### 4. Connection Issues
- Check firewall settings
- Verify services are running: `docker-compose ps`
- Check logs: `docker-compose logs -f`
- Test with curl: `curl http://localhost:8090/health`

#### 5. Proto Generation Fails
```bash
# Verify protoc installation
protoc --version

# Check Go plugins
which protoc-gen-go
which protoc-gen-go-grpc

# Reinstall if needed
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Getting Help

1. Check logs:
   - Backend: Console output or log files
   - Unreal: Output Log (Window > Developer Tools > Output Log)

2. Enable debug logging:
   - Backend: Set `LOG_LEVEL=debug`
   - Unreal: Set `EnableDebugLogging=true` in config

3. Community resources:
   - Discord: [Join our Discord](#)
   - Forums: [Community Forums](#)
   - Issues: [GitHub Issues](https://github.com/your-org/mmorpg-template/issues)

## Next Steps

1. Read the [Architecture Overview](ARCHITECTURE.md)
2. Review [Coding Standards](CODING_STANDARDS.md)
3. Check [API Documentation](API_DOCUMENTATION.md)
4. Start with [Tutorial: Creating Your First Feature](TUTORIAL_FIRST_FEATURE.md)

Happy coding! 🎮