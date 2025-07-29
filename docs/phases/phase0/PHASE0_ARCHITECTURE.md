# 🏗️ Phase 0: Foundation - Architecture Document

## 📋 Executive Summary

This document details the technical architecture for Phase 0's foundation implementation. Phase 0 establishes the core infrastructure, networking clients, modular code organization, and development tools that all subsequent phases build upon. The architecture emphasizes modularity, extensibility, and developer experience.

**Key Architectural Decisions:**
- Hexagonal architecture for backend services
- Modular C++ architecture for Unreal Engine client
- Protocol Buffers for cross-platform communication
- Docker-based development environment
- Comprehensive error handling and logging

---

## 🎮 1. Client Architecture (Unreal Engine 5.6)

### 1.1 Module Structure

```
MMORPGTemplate/
├── Source/
│   ├── MMORPGCore/          # Core systems and types
│   ├── MMORPGNetwork/       # HTTP/WebSocket clients
│   ├── MMORPGProto/         # Protocol Buffer integration
│   └── MMORPGUI/            # UI and console systems
```

### 1.2 Core Module Implementation

```cpp
// Source/MMORPGCore/Public/Core/MMORPGGameInstance.h
#pragma once

#include "CoreMinimal.h"
#include "Engine/GameInstance.h"
#include "MMORPGGameInstance.generated.h"

UCLASS()
class MMORGCORE_API UMMORPGGameInstance : public UGameInstance
{
    GENERATED_BODY()

public:
    virtual void Init() override;
    virtual void Shutdown() override;

    // Subsystem accessors
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Core")
    class UMMORPGErrorSubsystem* GetErrorSubsystem() const;

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Core")
    class UMMORPGHTTPClient* GetHTTPClient() const;

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Core")
    class UMMORPGWebSocketClient* GetWebSocketClient() const;

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Core")
    class UMMORPGConsoleManager* GetConsoleManager() const;

private:
    void InitializeSubsystems();
    void RegisterConsoleCommands();
};
```

### 1.3 Error Handling Architecture

```cpp
// Source/MMORPGCore/Public/Errors/MMORPGErrorSubsystem.h
UENUM(BlueprintType)
enum class EMMORPGErrorSeverity : uint8
{
    Info        UMETA(DisplayName = "Info"),
    Warning     UMETA(DisplayName = "Warning"),
    Error       UMETA(DisplayName = "Error"),
    Critical    UMETA(DisplayName = "Critical")
};

USTRUCT(BlueprintType)
struct MMORGCORE_API FMMORPGError
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadOnly)
    int32 Code = 0;

    UPROPERTY(BlueprintReadOnly)
    FString Message;

    UPROPERTY(BlueprintReadOnly)
    FString Context;

    UPROPERTY(BlueprintReadOnly)
    EMMORPGErrorSeverity Severity = EMMORPGErrorSeverity::Error;

    UPROPERTY(BlueprintReadOnly)
    FDateTime Timestamp;

    // Helper functions
    bool IsNetworkError() const;
    bool IsAuthError() const;
    FString ToUserFriendlyMessage() const;
};

DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnMMORPGError, const FMMORPGError&, Error);

UCLASS()
class MMORGCORE_API UMMORPGErrorSubsystem : public UGameInstanceSubsystem
{
    GENERATED_BODY()

public:
    // Blueprint events
    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Errors")
    FOnMMORPGError OnErrorOccurred;

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Errors")
    FOnMMORPGError OnCriticalError;

    // Error reporting
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Errors")
    void ReportError(const FMMORPGError& Error);

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Errors")
    void ReportErrorFromCode(int32 ErrorCode, const FString& Context);

    // Error history
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Errors")
    TArray<FMMORPGError> GetRecentErrors(int32 Count = 10) const;

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Errors")
    void ClearErrorHistory();

private:
    TArray<FMMORPGError> ErrorHistory;
    FCriticalSection ErrorHistoryLock;
    
    static constexpr int32 MaxErrorHistory = 100;
};
```

### 1.4 Network Module Architecture

```cpp
// Source/MMORPGNetwork/Public/Http/MMORPGHTTPClient.h
DECLARE_DYNAMIC_MULTICAST_DELEGATE_TwoParams(
    FOnHTTPRequestComplete,
    const FString&, Response,
    bool, bSuccess
);

UCLASS(BlueprintType)
class MMORGNETWORK_API UMMORPGHTTPRequest : public UObject
{
    GENERATED_BODY()

public:
    UPROPERTY(BlueprintAssignable)
    FOnHTTPRequestComplete OnComplete;

    UFUNCTION(BlueprintCallable, Category = "MMORPG|HTTP")
    void Send();

    UFUNCTION(BlueprintCallable, Category = "MMORPG|HTTP")
    void Cancel();

    // Builder pattern for Blueprint
    UFUNCTION(BlueprintCallable, Category = "MMORPG|HTTP", meta = (DisplayName = "Set Header"))
    UMMORPGHTTPRequest* K2_SetHeader(const FString& HeaderName, const FString& HeaderValue);

    UFUNCTION(BlueprintCallable, Category = "MMORPG|HTTP", meta = (DisplayName = "Set Body"))
    UMMORPGHTTPRequest* K2_SetBody(const FString& Body);

private:
    friend class UMMORPGHTTPClient;
    
    TSharedPtr<IHttpRequest, ESPMode::ThreadSafe> Request;
    void ProcessResponse(FHttpResponsePtr Response, bool bSuccess);
};

UCLASS()
class MMORGNETWORK_API UMMORPGHTTPClient : public UGameInstanceSubsystem
{
    GENERATED_BODY()

public:
    // Configuration
    UPROPERTY(EditDefaultsOnly, Category = "MMORPG|HTTP")
    FString DefaultBaseURL = "http://localhost:8080";

    UPROPERTY(EditDefaultsOnly, Category = "MMORPG|HTTP")
    float DefaultTimeout = 30.0f;

    // Request creation
    UFUNCTION(BlueprintCallable, Category = "MMORPG|HTTP")
    UMMORPGHTTPRequest* CreateRequest(
        const FString& Verb,
        const FString& URL
    );

    UFUNCTION(BlueprintCallable, Category = "MMORPG|HTTP")
    UMMORPGHTTPRequest* GET(const FString& URL);

    UFUNCTION(BlueprintCallable, Category = "MMORPG|HTTP")
    UMMORPGHTTPRequest* POST(const FString& URL, const FString& Body);

    // Authentication
    UFUNCTION(BlueprintCallable, Category = "MMORPG|HTTP")
    void SetAuthToken(const FString& Token);

    UFUNCTION(BlueprintCallable, Category = "MMORPG|HTTP")
    void ClearAuthToken();

private:
    FString AuthToken;
    FCriticalSection AuthTokenLock;
};
```

### 1.5 WebSocket Architecture

```cpp
// Source/MMORPGNetwork/Public/WebSocket/MMORPGWebSocketClient.h
UENUM(BlueprintType)
enum class EWebSocketState : uint8
{
    Disconnected    UMETA(DisplayName = "Disconnected"),
    Connecting      UMETA(DisplayName = "Connecting"),
    Connected       UMETA(DisplayName = "Connected"),
    Reconnecting    UMETA(DisplayName = "Reconnecting"),
    Error           UMETA(DisplayName = "Error")
};

DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnWebSocketStateChanged, EWebSocketState, NewState);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnWebSocketMessage, const FString&, Message);

UCLASS()
class MMORGNETWORK_API UMMORPGWebSocketClient : public UGameInstanceSubsystem, public FTickableGameObject
{
    GENERATED_BODY()

public:
    // Events
    UPROPERTY(BlueprintAssignable)
    FOnWebSocketStateChanged OnStateChanged;

    UPROPERTY(BlueprintAssignable)
    FOnWebSocketMessage OnMessageReceived;

    // Connection management
    UFUNCTION(BlueprintCallable, Category = "MMORPG|WebSocket")
    void Connect(const FString& URL);

    UFUNCTION(BlueprintCallable, Category = "MMORPG|WebSocket")
    void Disconnect();

    UFUNCTION(BlueprintCallable, Category = "MMORPG|WebSocket")
    void SendMessage(const FString& Message);

    UFUNCTION(BlueprintCallable, Category = "MMORPG|WebSocket")
    bool IsConnected() const;

    // Auto-reconnection
    UPROPERTY(EditDefaultsOnly, Category = "MMORPG|WebSocket")
    bool bAutoReconnect = true;

    UPROPERTY(EditDefaultsOnly, Category = "MMORPG|WebSocket")
    float InitialReconnectDelay = 1.0f;

    UPROPERTY(EditDefaultsOnly, Category = "MMORPG|WebSocket")
    float MaxReconnectDelay = 30.0f;

    UPROPERTY(EditDefaultsOnly, Category = "MMORPG|WebSocket")
    float ReconnectBackoffMultiplier = 2.0f;

    // FTickableGameObject interface
    virtual void Tick(float DeltaTime) override;
    virtual bool IsTickable() const override { return true; }
    virtual TStatId GetStatId() const override;

private:
    TSharedPtr<IWebSocket> WebSocket;
    FString CurrentURL;
    EWebSocketState State = EWebSocketState::Disconnected;
    
    // Reconnection state
    float CurrentReconnectDelay = 1.0f;
    float TimeUntilReconnect = 0.0f;
    int32 ReconnectAttempts = 0;

    // Thread safety
    FCriticalSection StateLock;
    TQueue<FString> OutgoingMessages;
    TQueue<FString> IncomingMessages;

    void OnConnected();
    void OnConnectionError(const FString& Error);
    void OnClosed(int32 StatusCode, const FString& Reason, bool bWasClean);
    void OnMessage(const FString& Message);
    void OnRawMessage(const void* Data, SIZE_T Size, SIZE_T BytesRemaining);

    void ProcessIncomingMessages();
    void AttemptReconnection();
    void ResetReconnectionState();
};
```

### 1.6 Console System Architecture

```cpp
// Source/MMORPGUI/Public/Console/MMORPGConsoleManager.h
DECLARE_DELEGATE_RetVal_OneParam(bool, FMMORPGConsoleCommandDelegate, const TArray<FString>&);

USTRUCT()
struct MMORPGUI_API FMMORPGConsoleCommand
{
    GENERATED_BODY()

    FString Command;
    FString Description;
    FString Usage;
    FMMORPGConsoleCommandDelegate Delegate;
};

UCLASS()
class MMORPGUI_API UMMORPGConsoleManager : public UGameInstanceSubsystem
{
    GENERATED_BODY()

public:
    virtual void Initialize(FSubsystemCollectionBase& Collection) override;
    virtual void Deinitialize() override;

    // Command registration
    void RegisterCommand(
        const FString& Command,
        const FString& Description,
        const FString& Usage,
        const FMMORPGConsoleCommandDelegate& Delegate
    );

    void UnregisterCommand(const FString& Command);

    // Command execution
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    bool ExecuteCommand(const FString& CommandLine, FString& OutResult);

    // Command discovery
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    TArray<FString> GetAllCommands() const;

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    FString GetCommandHelp(const FString& Command) const;

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Console")
    TArray<FString> GetCommandSuggestions(const FString& Partial) const;

private:
    TMap<FString, FMMORPGConsoleCommand> Commands;
    
    void RegisterBuiltInCommands();
    
    // Built-in commands
    bool Cmd_Help(const TArray<FString>& Args);
    bool Cmd_Clear(const TArray<FString>& Args);
    bool Cmd_ShowFPS(const TArray<FString>& Args);
    bool Cmd_SetResolution(const TArray<FString>& Args);
    bool Cmd_MemStats(const TArray<FString>& Args);
    bool Cmd_NetStatus(const TArray<FString>& Args);
    bool Cmd_MMORPGTest(const TArray<FString>& Args);
};
```

---

## 🖥️ 2. Backend Architecture (Go)

### 2.1 Hexagonal Architecture

```
mmorpg-backend/
├── cmd/                    # Application entry points
│   ├── gateway/           # API Gateway service
│   ├── auth/             # Authentication service (Phase 1)
│   └── ...               # Other services
├── internal/              # Private application code
│   ├── domain/           # Business logic and entities
│   ├── application/      # Use cases and services
│   ├── adapters/         # External interfaces
│   └── ports/            # Interface definitions
├── pkg/                   # Public packages
│   ├── proto/            # Protocol Buffer definitions
│   ├── logger/           # Logging utilities
│   └── errors/           # Error handling
└── deployments/          # Deployment configurations
```

### 2.2 Gateway Service Implementation

```go
// cmd/gateway/main.go
package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/mmorpg-template/backend/internal/config"
    "github.com/mmorpg-template/backend/pkg/logger"
    "github.com/mmorpg-template/backend/pkg/metrics"
)

func main() {
    ctx := context.Background()
    
    // Initialize logger
    log := logger.NewWithService("gateway")
    log.Info("Starting MMORPG Gateway Service...")
    
    // Load configuration
    cfg := config.Load()
    
    // Setup metrics server
    metricsServer := metrics.NewServer(cfg.Metrics.Port)
    go func() {
        if err := metricsServer.Start(); err != nil {
            log.Errorf("Metrics server error: %v", err)
        }
    }()
    
    // Setup HTTP server
    srv := &http.Server{
        Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
        Handler:      setupRoutes(cfg, log),
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    // Start server
    go func() {
        log.Infof("Gateway server listening on %s", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Info("Shutting down gateway service...")
    
    // Graceful shutdown
    shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Errorf("Server forced to shutdown: %v", err)
    }
    
    log.Info("Gateway service stopped")
}
```

### 2.3 Protocol Buffer Definitions

```protobuf
// pkg/proto/base.proto
syntax = "proto3";

package mmorpg;

option go_package = "github.com/mmorpg-template/backend/pkg/proto";

import "google/protobuf/timestamp.proto";

// Base message envelope
message GameMessage {
    uint32 version = 1;
    uint32 sequence = 2;
    google.protobuf.Timestamp timestamp = 3;
    MessageType type = 4;
    bytes payload = 5;
}

// Message types
enum MessageType {
    MESSAGE_TYPE_UNSPECIFIED = 0;
    
    // System messages (500-599)
    MESSAGE_TYPE_SYSTEM_PING = 500;
    MESSAGE_TYPE_SYSTEM_PONG = 501;
    MESSAGE_TYPE_SYSTEM_ERROR = 502;
    MESSAGE_TYPE_SYSTEM_NOTIFICATION = 503;
}

// Error codes
enum ErrorCode {
    ERROR_CODE_UNSPECIFIED = 0;
    ERROR_CODE_INVALID_REQUEST = 1;
    ERROR_CODE_UNAUTHORIZED = 2;
    ERROR_CODE_SERVER_ERROR = 7;
    ERROR_CODE_SERVICE_UNAVAILABLE = 8;
}

// Common data structures
message Vector3 {
    float x = 1;
    float y = 2;
    float z = 3;
}

message ErrorResponse {
    ErrorCode code = 1;
    string message = 2;
    map<string, string> details = 3;
}
```

### 2.4 Error Handling System

```go
// pkg/errors/errors.go
package errors

import (
    "fmt"
    "runtime"
)

type ErrorCode int

const (
    CodeUnknown ErrorCode = iota
    CodeInvalidRequest
    CodeUnauthorized
    CodeForbidden
    CodeNotFound
    CodeConflict
    CodeInternal
    CodeServiceUnavailable
)

type MMORPGError struct {
    Code       ErrorCode
    Message    string
    Details    map[string]interface{}
    Err        error
    StackTrace string
}

func New(code ErrorCode, message string) *MMORPGError {
    return &MMORPGError{
        Code:       code,
        Message:    message,
        Details:    make(map[string]interface{}),
        StackTrace: captureStackTrace(),
    }
}

func Wrap(err error, code ErrorCode, message string) *MMORPGError {
    return &MMORPGError{
        Code:       code,
        Message:    message,
        Err:        err,
        Details:    make(map[string]interface{}),
        StackTrace: captureStackTrace(),
    }
}

func (e *MMORPGError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

func (e *MMORPGError) WithDetail(key string, value interface{}) *MMORPGError {
    e.Details[key] = value
    return e
}

func captureStackTrace() string {
    buf := make([]byte, 4096)
    n := runtime.Stack(buf, false)
    return string(buf[:n])
}
```

### 2.5 Logging Infrastructure

```go
// pkg/logger/logger.go
package logger

import (
    "os"
    
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)
    
    With(fields ...Field) Logger
    WithError(err error) Logger
}

type zapLogger struct {
    logger *zap.Logger
}

func New() Logger {
    config := zap.NewProductionConfig()
    
    // Configure based on environment
    if os.Getenv("GO_ENV") == "development" {
        config = zap.NewDevelopmentConfig()
        config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    }
    
    // Add custom fields
    config.InitialFields = map[string]interface{}{
        "app": "mmorpg-backend",
    }
    
    logger, err := config.Build()
    if err != nil {
        panic(err)
    }
    
    return &zapLogger{logger: logger}
}

func NewWithService(service string) Logger {
    return New().With(String("service", service))
}

func (l *zapLogger) Debug(msg string, fields ...Field) {
    l.logger.Debug(msg, convertFields(fields)...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
    l.logger.Info(msg, convertFields(fields)...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
    l.logger.Error(msg, convertFields(fields)...)
}

func (l *zapLogger) WithError(err error) Logger {
    return &zapLogger{
        logger: l.logger.With(zap.Error(err)),
    }
}
```

---

## 🐳 3. Development Environment

### 3.1 Docker Compose Configuration

```yaml
# docker-compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: mmorpg
      POSTGRES_PASSWORD: mmorpg_dev
      POSTGRES_DB: mmorpg_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U mmorpg"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  nats:
    image: nats:2.10-alpine
    command: "-js -sd /data"
    ports:
      - "4222:4222"
      - "8222:8222"
    volumes:
      - nats_data:/data
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8222/healthz"]
      interval: 10s
      timeout: 5s
      retries: 5

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./deployments/docker/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'

volumes:
  postgres_data:
  redis_data:
  nats_data:
  prometheus_data:
```

### 3.2 Development Workflow

```makefile
# Makefile
.PHONY: help dev test build clean

help:
	@echo "Available commands:"
	@echo "  make dev    - Start development environment"
	@echo "  make test   - Run tests"
	@echo "  make build  - Build all services"
	@echo "  make clean  - Clean up containers and volumes"

dev:
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	go run cmd/gateway/main.go

test:
	go test -v -race ./...

build:
	CGO_ENABLED=0 go build -o bin/gateway cmd/gateway/main.go

clean:
	docker-compose down -v
	rm -rf bin/

# Database migrations
migrate-up:
	migrate -path migrations -database "postgresql://mmorpg:mmorpg_dev@localhost:5432/mmorpg_db?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgresql://mmorpg:mmorpg_dev@localhost:5432/mmorpg_db?sslmode=disable" down

# Protocol buffer generation
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/proto/*.proto
```

---

## 📋 4. CI/CD Pipeline

### 4.1 GitHub Actions Workflow

```yaml
# .github/workflows/backend.yml
name: Backend CI

on:
  push:
    branches: [main, develop]
    paths:
      - 'mmorpg-backend/**'
      - '.github/workflows/backend.yml'
  pull_request:
    branches: [main, develop]
    paths:
      - 'mmorpg-backend/**'

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: test_db
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
          
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-
            
      - name: Install dependencies
        working-directory: ./mmorpg-backend
        run: go mod download
        
      - name: Run tests
        working-directory: ./mmorpg-backend
        run: go test -v -race -coverprofile=coverage.out ./...
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/test_db?sslmode=disable
          REDIS_URL: redis://localhost:6379
          
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./mmorpg-backend/coverage.out
          
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          working-directory: ./mmorpg-backend
```

---

## 🎯 Summary

Phase 0's architecture establishes a robust foundation with:

1. **Client Architecture**
   - Modular C++ design with clear separation of concerns
   - Comprehensive error handling and logging
   - Thread-safe networking with auto-reconnection
   - Extensible console system for development
   - Full Blueprint support for all systems

2. **Backend Architecture**
   - Clean hexagonal architecture
   - Microservices-ready structure
   - Protocol Buffer support
   - Comprehensive error handling
   - Structured logging

3. **Development Environment**
   - Docker-based infrastructure
   - Hot-reload support
   - Automated testing
   - CI/CD pipeline

4. **Key Features**
   - Type-safe communication
   - Async operations throughout
   - Developer-friendly tools
   - Production-ready patterns

This foundation ensures that all subsequent phases can build upon a solid, scalable, and maintainable architecture.