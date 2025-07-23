# MMORPG Template - Quick Start Guide

## Prerequisites

### Backend Development
- Go 1.21 or higher
- Docker Desktop
- Protocol Buffers compiler (protoc)
- Make (optional but recommended)

### Unreal Engine Development
- Unreal Engine 5.6
- Visual Studio 2022 (Windows) or Xcode (Mac)
- Git

## Backend Setup (5 minutes)

1. **Start Infrastructure Services**
   ```bash
   cd mmorpg-backend
   docker-compose up -d
   ```
   This starts PostgreSQL, Redis, and NATS.

2. **Verify Services**
   ```bash
   docker ps
   ```
   You should see 3 running containers:
   - mmorpg-postgres
   - mmorpg-redis
   - mmorpg-nats

3. **Build the Gateway Service**
   ```bash
   make build-gateway
   # or
   go build -o bin/gateway ./cmd/gateway/main.go
   ```

4. **Run the Gateway**
   ```bash
   ./bin/gateway
   ```
   The gateway will start on port 8080.

## Unreal Engine Setup (10 minutes)

1. **Open the Game Template Project**
   ```bash
   # Navigate to the game template directory
   cd MMORPGTemplate
   ```

2. **Generate Project Files**
   - Right-click MMORPGTemplate.uproject
   - Select "Generate Visual Studio project files" (Windows)
   - Or use UnrealBuildTool on Mac/Linux

3. **Open and Compile**
   - Open MMORPGTemplate.uproject in Unreal Engine 5.6
   - The project will compile automatically
   - Wait for compilation to complete

4. **Verify Setup**
   - The game template should open in the editor
   - Check that all systems are loaded
   - (Note: Client networking features pending implementation)

5. **Access MMORPG Tools**
   - Window > MMORPG Tools > Dashboard
   - Or click the MMORPG button in the toolbar

## Testing the Setup

### Backend Health Check
```bash
curl http://localhost:8080/health
```
Should return: `{"status": "healthy"}`

### Database Connection
```bash
docker exec -it mmorpg-postgres psql -U dev -d mmorpg
\dt
```
Should show the database tables.

### Protocol Buffer Compilation
```bash
cd mmorpg-backend
# Windows
scripts\compile_proto.bat
# Mac/Linux
./scripts/compile_proto.sh
```

## Common Issues

### Port Already in Use
If ports are already in use, edit `docker-compose.yml` to change port mappings:
```yaml
postgres:
  ports:
    - "5433:5432"  # Changed from 5432
```

### Docker Not Running
Make sure Docker Desktop is running before using docker-compose.

### Project Compilation Errors
- Ensure you're using Unreal Engine 5.6
- Check that all dependencies in the .Build.cs files are available
- Try deleting Intermediate and Binaries folders and recompiling

## Next Steps

1. **Explore the Code**
   - Backend: Check `mmorpg-backend/cmd/gateway/main.go`
   - Game Template: Look at `MMORPGTemplate.h` and `MMORPGTemplate.cpp`
   - Protos: Review the message definitions in `pkg/proto/`

2. **Run with Hot Reload (Backend)**
   ```bash
   cd mmorpg-backend
   docker-compose -f docker-compose.dev.yml up gateway
   ```

3. **Create a Test Blueprint**
   - Create a new GameInstance Blueprint
   - Add initialization logic to connect to backend
   - Test in PIE (Play In Editor)

## Useful Commands

### Backend
```bash
make help          # Show all available commands
make test          # Run tests
make docker-up     # Start services
make docker-down   # Stop services
make docker-logs   # View logs
```

### Database
```bash
# Connect to database
docker exec -it mmorpg-postgres psql -U dev -d mmorpg

# View tables
\dt

# Describe a table
\d users
```

### Monitoring
- NATS Dashboard: http://localhost:8222
- Add Prometheus: http://localhost:9090 (if enabled)
- Add Grafana: http://localhost:3000 (if enabled)

## Development Workflow

1. Make backend changes
2. Run `make build` or use hot reload
3. Test with curl or Postman
4. Update Unreal Engine code
5. Compile in editor
6. Test in PIE

## Support

- Check `PHASE0_SUMMARY.md` for architecture details
- Review individual README files in each directory
- Look at the design documents in the root directory

Happy coding! 🎮