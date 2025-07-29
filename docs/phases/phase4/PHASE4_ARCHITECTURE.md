# 🏗️ Phase 4: Production Tools - Architecture

## 📋 Executive Summary

This document defines the complete technical architecture for Phase 4's production tools and infrastructure. Following hexagonal architecture principles, we separate core business logic from infrastructure concerns, enabling testability, maintainability, and flexibility in our production systems.

**Architecture Principles**:
- Hexagonal (Ports & Adapters) architecture for all services
- Domain-Driven Design for bounded contexts
- Event-driven communication between services
- Infrastructure as Code for all deployments
- Zero-trust security model

---

## 🌐 Production Architecture Overview

### Multi-Tier Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Presentation Tier                      │
├─────────────────────────────────────────────────────────────┤
│  Admin Dashboard   │  GM Client Tools  │  Mobile Support    │
│  (React + Redux)   │  (Unreal Plugin)  │  (React Native)    │
└────────────────────┴──────────────────┴────────────────────┘
                               │
                          HTTPS/WSS
                               │
┌─────────────────────────────────────────────────────────────┐
│                      API Gateway Tier                         │
├─────────────────────────────────────────────────────────────┤
│  Kong API Gateway  │  Rate Limiting    │  JWT Validation    │
│  Request Routing   │  API Versioning   │  Request Transform │
└────────────────────┴──────────────────┴────────────────────┘
                               │
                         Service Mesh
                               │
┌─────────────────────────────────────────────────────────────┐
│                    Application Services Tier                  │
├─────────────────────────────────────────────────────────────┤
│  Admin Service     │  Analytics Service│  Support Service   │
│  Content Service   │  Monitoring Svc   │  Deployment Svc    │
└────────────────────┴──────────────────┴────────────────────┘
                               │
                          Data Layer
                               │
┌─────────────────────────────────────────────────────────────┐
│                        Data Tier                              │
├─────────────────────────────────────────────────────────────┤
│  PostgreSQL        │  Redis Cache      │  S3 Object Store   │
│  TimescaleDB       │  Kafka Streams    │  Elasticsearch     │
└────────────────────┴──────────────────┴────────────────────┘
```

### Service Mesh Implementation (Istio)

```yaml
# istio-service-mesh.yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: mmorpg-istio-control-plane
spec:
  values:
    global:
      meshID: mmorpg-mesh
      multiCluster:
        clusterName: production
      network: mmorpg-network
  components:
    base:
      enabled: true
    pilot:
      enabled: true
      k8s:
        resources:
          requests:
            cpu: 500m
            memory: 2048Mi
    ingressGateways:
    - name: istio-ingressgateway
      enabled: true
      k8s:
        service:
          type: LoadBalancer
          ports:
          - port: 80
            targetPort: 8080
            name: http
          - port: 443
            targetPort: 8443
            name: https
    egressGateways:
    - name: istio-egressgateway
      enabled: true
  meshConfig:
    accessLogFile: /dev/stdout
    defaultConfig:
      proxyStatsMatcher:
        inclusionRegexps:
        - ".*outlier_detection.*"
        - ".*circuit_breakers.*"
        - ".*upstream_rq_retry.*"
        - ".*upstream_rq_pending.*"
```

### Observability Stack Architecture

```go
// pkg/observability/telemetry.go
package observability

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/prometheus"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/metric"
    "go.opentelemetry.io/otel/sdk/trace"
)

type TelemetryConfig struct {
    ServiceName      string
    JaegerEndpoint   string
    PrometheusPort   int
    SamplingRate     float64
}

type Telemetry struct {
    TracerProvider *trace.TracerProvider
    MeterProvider  *metric.MeterProvider
    Shutdown       func(context.Context) error
}

func InitTelemetry(cfg TelemetryConfig) (*Telemetry, error) {
    // Initialize Jaeger exporter for distributed tracing
    jaegerExporter, err := jaeger.New(
        jaeger.WithCollectorEndpoint(
            jaeger.WithEndpoint(cfg.JaegerEndpoint),
        ),
    )
    if err != nil {
        return nil, err
    }

    // Initialize tracer provider
    tp := trace.NewTracerProvider(
        trace.WithBatcher(jaegerExporter),
        trace.WithSampler(trace.TraceIDRatioBased(cfg.SamplingRate)),
        trace.WithResource(newResource(cfg.ServiceName)),
    )
    otel.SetTracerProvider(tp)

    // Initialize Prometheus exporter for metrics
    promExporter, err := prometheus.New()
    if err != nil {
        return nil, err
    }

    // Initialize meter provider
    mp := metric.NewMeterProvider(
        metric.WithReader(promExporter),
        metric.WithResource(newResource(cfg.ServiceName)),
    )
    otel.SetMeterProvider(mp)

    return &Telemetry{
        TracerProvider: tp,
        MeterProvider:  mp,
        Shutdown: func(ctx context.Context) error {
            if err := tp.Shutdown(ctx); err != nil {
                return err
            }
            return mp.Shutdown(ctx)
        },
    }, nil
}
```

### Security Layers and Boundaries

```go
// pkg/security/layers.go
package security

import (
    "crypto/rsa"
    "github.com/golang-jwt/jwt/v4"
    "golang.org/x/crypto/bcrypt"
)

// SecurityLayer defines the security boundaries
type SecurityLayer int

const (
    LayerPublic SecurityLayer = iota
    LayerAuthenticated
    LayerAuthorized
    LayerPrivileged
    LayerSystem
)

type SecurityConfig struct {
    JWTSigningKey   *rsa.PrivateKey
    JWTVerifyKey    *rsa.PublicKey
    PasswordCost    int
    TokenExpiration time.Duration
    MFARequired     bool
}

// SecurityMiddleware implements layered security checks
type SecurityMiddleware struct {
    config *SecurityConfig
    rbac   *RBACService
    audit  *AuditService
}

func (sm *SecurityMiddleware) Authenticate(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return sm.config.JWTVerifyKey, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, ErrInvalidToken
    }
    
    // Audit authentication attempt
    sm.audit.LogAuthentication(claims.Subject, true)
    
    return claims, nil
}

func (sm *SecurityMiddleware) Authorize(claims *Claims, resource string, action string) error {
    allowed, err := sm.rbac.CheckPermission(claims.Subject, resource, action)
    if err != nil {
        return err
    }
    
    if !allowed {
        sm.audit.LogAuthorizationFailure(claims.Subject, resource, action)
        return ErrUnauthorized
    }
    
    sm.audit.LogAuthorization(claims.Subject, resource, action)
    return nil
}
```

---

## 🔐 Admin Services Architecture

### Admin API Gateway Design

```go
// cmd/admin-gateway/main.go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/mmorpg/admin/pkg/gateway"
    "github.com/mmorpg/admin/pkg/middleware"
)

func main() {
    cfg := gateway.LoadConfig()
    
    // Initialize gateway with middleware chain
    gw := gateway.New(cfg,
        middleware.RequestID(),
        middleware.Logger(),
        middleware.RateLimiter(cfg.RateLimit),
        middleware.Authentication(cfg.Auth),
        middleware.Authorization(cfg.RBAC),
        middleware.Audit(cfg.Audit),
    )
    
    // Setup routes
    gw.SetupRoutes()
    
    // Start gateway
    gw.Run()
}
```

```go
// pkg/gateway/gateway.go
package gateway

import (
    "net/http/httputil"
    "net/url"
)

type Gateway struct {
    router      *gin.Engine
    services    map[string]*ServiceEndpoint
    middlewares []gin.HandlerFunc
}

type ServiceEndpoint struct {
    Name        string
    URL         *url.URL
    HealthCheck string
    Proxy       *httputil.ReverseProxy
}

func (g *Gateway) SetupRoutes() {
    api := g.router.Group("/api/v1")
    
    // Admin routes
    admin := api.Group("/admin")
    {
        admin.GET("/users", g.proxyTo("admin-service", "/users"))
        admin.PUT("/users/:id", g.proxyTo("admin-service", "/users/:id"))
        admin.POST("/users/:id/ban", g.proxyTo("admin-service", "/users/:id/ban"))
        admin.GET("/audit-logs", g.proxyTo("admin-service", "/audit-logs"))
    }
    
    // Analytics routes
    analytics := api.Group("/analytics")
    {
        analytics.GET("/metrics", g.proxyTo("analytics-service", "/metrics"))
        analytics.GET("/reports", g.proxyTo("analytics-service", "/reports"))
        analytics.POST("/custom-query", g.proxyTo("analytics-service", "/custom-query"))
    }
    
    // Content management routes
    content := api.Group("/content")
    {
        content.GET("/items", g.proxyTo("content-service", "/items"))
        content.POST("/items", g.proxyTo("content-service", "/items"))
        content.PUT("/items/:id", g.proxyTo("content-service", "/items/:id"))
        content.POST("/deploy", g.proxyTo("content-service", "/deploy"))
    }
}
```

### Authentication/Authorization for Admins

```go
// internal/admin/domain/auth.go
package domain

type AdminUser struct {
    ID          string
    Email       string
    Username    string
    Roles       []Role
    Permissions []Permission
    MFAEnabled  bool
    LastLogin   time.Time
}

type Role struct {
    ID          string
    Name        string
    Description string
    Permissions []Permission
}

type Permission struct {
    ID       string
    Resource string
    Action   string
}

// Hexagonal architecture - Port definition
type AuthRepository interface {
    GetUserByEmail(ctx context.Context, email string) (*AdminUser, error)
    ValidateCredentials(ctx context.Context, email, password string) error
    UpdateLastLogin(ctx context.Context, userID string) error
    GetUserPermissions(ctx context.Context, userID string) ([]Permission, error)
}

type AuthService interface {
    Login(ctx context.Context, email, password, mfaToken string) (*TokenPair, error)
    RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
    ValidateToken(ctx context.Context, token string) (*Claims, error)
    Logout(ctx context.Context, token string) error
}
```

```go
// internal/admin/application/auth_service.go
package application

type authService struct {
    repo      domain.AuthRepository
    tokenGen  TokenGenerator
    mfaVerify MFAVerifier
    cache     Cache
}

func (s *authService) Login(ctx context.Context, email, password, mfaToken string) (*TokenPair, error) {
    user, err := s.repo.GetUserByEmail(ctx, email)
    if err != nil {
        return nil, err
    }
    
    if err := s.repo.ValidateCredentials(ctx, email, password); err != nil {
        return nil, err
    }
    
    if user.MFAEnabled {
        if err := s.mfaVerify.Verify(user.ID, mfaToken); err != nil {
            return nil, ErrInvalidMFA
        }
    }
    
    permissions, err := s.repo.GetUserPermissions(ctx, user.ID)
    if err != nil {
        return nil, err
    }
    
    claims := &Claims{
        UserID:      user.ID,
        Email:       user.Email,
        Roles:       user.Roles,
        Permissions: permissions,
    }
    
    tokens, err := s.tokenGen.GenerateTokenPair(claims)
    if err != nil {
        return nil, err
    }
    
    // Cache session
    s.cache.Set(ctx, "session:"+user.ID, tokens, 24*time.Hour)
    
    // Update last login
    s.repo.UpdateLastLogin(ctx, user.ID)
    
    return tokens, nil
}
```

### Admin Database Schema

```sql
-- Admin user management schema
CREATE SCHEMA admin;

-- Admin users table
CREATE TABLE admin.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    mfa_secret VARCHAR(255),
    mfa_enabled BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP,
    failed_attempts INT DEFAULT 0,
    locked_until TIMESTAMP
);

-- Roles table
CREATE TABLE admin.roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Permissions table
CREATE TABLE admin.permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    UNIQUE(resource, action)
);

-- User-Role mapping
CREATE TABLE admin.user_roles (
    user_id UUID REFERENCES admin.users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES admin.roles(id) ON DELETE CASCADE,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    granted_by UUID REFERENCES admin.users(id),
    PRIMARY KEY (user_id, role_id)
);

-- Role-Permission mapping
CREATE TABLE admin.role_permissions (
    role_id UUID REFERENCES admin.roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES admin.permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- Audit logs table
CREATE TABLE admin.audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES admin.users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_audit_logs_user_id ON admin.audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON admin.audit_logs(created_at);
CREATE INDEX idx_audit_logs_action ON admin.audit_logs(action);

-- Default roles
INSERT INTO admin.roles (name, description, is_system) VALUES
('super_admin', 'Full system access', true),
('admin', 'Administrative access', true),
('moderator', 'Content moderation access', true),
('support', 'Customer support access', true),
('analyst', 'Read-only analytics access', true);
```

### Audit Logging Architecture

```go
// internal/admin/domain/audit.go
package domain

type AuditLog struct {
    ID           string
    UserID       string
    Action       string
    ResourceType string
    ResourceID   string
    OldValue     interface{}
    NewValue     interface{}
    IPAddress    string
    UserAgent    string
    Timestamp    time.Time
}

type AuditLogger interface {
    LogAction(ctx context.Context, log AuditLog) error
    GetLogs(ctx context.Context, filter AuditFilter) ([]AuditLog, error)
}

// internal/admin/infrastructure/audit_logger.go
package infrastructure

type postgresAuditLogger struct {
    db *sql.DB
}

func (l *postgresAuditLogger) LogAction(ctx context.Context, log domain.AuditLog) error {
    query := `
        INSERT INTO admin.audit_logs 
        (user_id, action, resource_type, resource_id, old_value, new_value, ip_address, user_agent)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
    
    oldValueJSON, _ := json.Marshal(log.OldValue)
    newValueJSON, _ := json.Marshal(log.NewValue)
    
    _, err := l.db.ExecContext(ctx, query,
        log.UserID,
        log.Action,
        log.ResourceType,
        log.ResourceID,
        oldValueJSON,
        newValueJSON,
        log.IPAddress,
        log.UserAgent,
    )
    
    return err
}

// Audit middleware for automatic logging
func AuditMiddleware(logger domain.AuditLogger) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Capture request state
        bodyBytes, _ := ioutil.ReadAll(c.Request.Body)
        c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
        
        // Process request
        c.Next()
        
        // Log if successful modification
        if c.Writer.Status() < 400 && isModifyingRequest(c.Request.Method) {
            log := domain.AuditLog{
                UserID:       c.GetString("user_id"),
                Action:       fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
                ResourceType: extractResourceType(c.Request.URL.Path),
                ResourceID:   c.Param("id"),
                IPAddress:    c.ClientIP(),
                UserAgent:    c.Request.UserAgent(),
                Timestamp:    time.Now(),
            }
            
            logger.LogAction(c.Request.Context(), log)
        }
    }
}
```

---

## 📝 Content Management Architecture

### CMS Backend Services

```go
// internal/cms/domain/content.go
package domain

// Domain entities following DDD
type Item struct {
    ID          string
    Name        string
    Description string
    Type        ItemType
    Rarity      Rarity
    Stats       ItemStats
    Version     int
    ValidFrom   time.Time
    ValidUntil  *time.Time
}

type Quest struct {
    ID           string
    Name         string
    Description  string
    Requirements []QuestRequirement
    Objectives   []QuestObjective
    Rewards      []Reward
    Version      int
}

type NPC struct {
    ID          string
    Name        string
    Type        NPCType
    Level       int
    SpawnPoints []SpawnPoint
    Dialogue    []DialogueNode
    LootTable   *LootTable
    Version     int
}

// Hexagonal architecture - Port definitions
type ContentRepository interface {
    // Items
    GetItem(ctx context.Context, id string) (*Item, error)
    ListItems(ctx context.Context, filter ItemFilter) ([]Item, error)
    CreateItem(ctx context.Context, item *Item) error
    UpdateItem(ctx context.Context, item *Item) error
    
    // Quests
    GetQuest(ctx context.Context, id string) (*Quest, error)
    ListQuests(ctx context.Context, filter QuestFilter) ([]Quest, error)
    CreateQuest(ctx context.Context, quest *Quest) error
    UpdateQuest(ctx context.Context, quest *Quest) error
    
    // NPCs
    GetNPC(ctx context.Context, id string) (*NPC, error)
    ListNPCs(ctx context.Context, filter NPCFilter) ([]NPC, error)
    CreateNPC(ctx context.Context, npc *NPC) error
    UpdateNPC(ctx context.Context, npc *NPC) error
}

type ContentService interface {
    // Item management
    CreateItem(ctx context.Context, req CreateItemRequest) (*Item, error)
    UpdateItem(ctx context.Context, id string, req UpdateItemRequest) (*Item, error)
    ValidateItem(ctx context.Context, item *Item) error
    
    // Quest management
    CreateQuest(ctx context.Context, req CreateQuestRequest) (*Quest, error)
    UpdateQuest(ctx context.Context, id string, req UpdateQuestRequest) (*Quest, error)
    ValidateQuest(ctx context.Context, quest *Quest) error
    
    // Deployment
    DeployContent(ctx context.Context, deploymentID string) error
    PreviewChanges(ctx context.Context) (*ContentChangeset, error)
}
```

### Version Control Integration

```go
// internal/cms/infrastructure/version_control.go
package infrastructure

import (
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing/object"
)

type GitVersionControl struct {
    repo     *git.Repository
    worktree *git.Worktree
    basePath string
}

func (g *GitVersionControl) CommitContent(ctx context.Context, content interface{}, message string) (string, error) {
    // Serialize content to JSON
    data, err := json.MarshalIndent(content, "", "  ")
    if err != nil {
        return "", err
    }
    
    // Determine file path based on content type
    filePath := g.getFilePath(content)
    
    // Write to file
    fullPath := filepath.Join(g.basePath, filePath)
    if err := os.WriteFile(fullPath, data, 0644); err != nil {
        return "", err
    }
    
    // Stage file
    _, err = g.worktree.Add(filePath)
    if err != nil {
        return "", err
    }
    
    // Commit
    commit, err := g.worktree.Commit(message, &git.CommitOptions{
        Author: &object.Signature{
            Name:  "CMS System",
            Email: "cms@mmorpg.com",
            When:  time.Now(),
        },
    })
    
    if err != nil {
        return "", err
    }
    
    return commit.String(), nil
}

func (g *GitVersionControl) GetContentHistory(ctx context.Context, contentType string, contentID string) ([]ContentVersion, error) {
    filePath := g.getFilePathByID(contentType, contentID)
    
    // Get commit history for file
    commits, err := g.repo.Log(&git.LogOptions{
        FileName: &filePath,
    })
    if err != nil {
        return nil, err
    }
    
    var versions []ContentVersion
    err = commits.ForEach(func(c *object.Commit) error {
        // Get file content at this commit
        file, err := c.File(filePath)
        if err != nil {
            return nil // File might not exist in this commit
        }
        
        content, err := file.Contents()
        if err != nil {
            return err
        }
        
        versions = append(versions, ContentVersion{
            CommitID:  c.Hash.String(),
            Timestamp: c.Author.When,
            Author:    c.Author.Name,
            Message:   c.Message,
            Content:   content,
        })
        
        return nil
    })
    
    return versions, err
}
```

### Hot-Reload Mechanisms

```go
// internal/cms/application/hot_reload.go
package application

import (
    "github.com/nats-io/nats.go"
)

type HotReloadService struct {
    nats         *nats.Conn
    contentCache *ContentCache
    validator    ContentValidator
}

func (h *HotReloadService) DeployContent(ctx context.Context, changeset *ContentChangeset) error {
    // Validate all changes
    for _, change := range changeset.Changes {
        if err := h.validator.Validate(change.Content); err != nil {
            return fmt.Errorf("validation failed for %s: %w", change.ID, err)
        }
    }
    
    // Prepare deployment package
    deployment := &ContentDeployment{
        ID:        uuid.New().String(),
        Timestamp: time.Now(),
        Changes:   changeset.Changes,
    }
    
    // Encode deployment
    data, err := proto.Marshal(deployment)
    if err != nil {
        return err
    }
    
    // Publish to all game servers
    if err := h.nats.Publish("content.deploy", data); err != nil {
        return err
    }
    
    // Update local cache
    for _, change := range changeset.Changes {
        h.contentCache.Set(change.ID, change.Content)
    }
    
    return nil
}

// Game server side hot-reload handler
type ContentReloadHandler struct {
    contentMgr *ContentManager
    validator  ContentValidator
}

func (c *ContentReloadHandler) HandleContentUpdate(msg *nats.Msg) {
    var deployment ContentDeployment
    if err := proto.Unmarshal(msg.Data, &deployment); err != nil {
        log.Printf("Failed to unmarshal deployment: %v", err)
        return
    }
    
    // Apply changes atomically
    tx := c.contentMgr.BeginTransaction()
    defer tx.Rollback()
    
    for _, change := range deployment.Changes {
        switch change.Type {
        case "item":
            var item Item
            if err := json.Unmarshal(change.Content, &item); err != nil {
                log.Printf("Failed to unmarshal item: %v", err)
                return
            }
            tx.UpdateItem(&item)
            
        case "quest":
            var quest Quest
            if err := json.Unmarshal(change.Content, &quest); err != nil {
                log.Printf("Failed to unmarshal quest: %v", err)
                return
            }
            tx.UpdateQuest(&quest)
            
        case "npc":
            var npc NPC
            if err := json.Unmarshal(change.Content, &npc); err != nil {
                log.Printf("Failed to unmarshal NPC: %v", err)
                return
            }
            tx.UpdateNPC(&npc)
        }
    }
    
    if err := tx.Commit(); err != nil {
        log.Printf("Failed to commit content update: %v", err)
        return
    }
    
    // Acknowledge successful update
    msg.Ack()
    
    log.Printf("Successfully applied content deployment %s", deployment.ID)
}
```

### Content Delivery Pipeline

```yaml
# kubernetes/content-delivery-pipeline.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: content-pipeline-config
data:
  pipeline.yaml: |
    stages:
      - name: validate
        steps:
          - validate-schema
          - check-references
          - test-balance
      
      - name: preview
        steps:
          - deploy-to-staging
          - run-integration-tests
          - generate-preview-report
      
      - name: approve
        steps:
          - require-approval
          - audit-log-approval
      
      - name: deploy
        steps:
          - create-backup
          - deploy-to-production
          - verify-deployment
          - rollback-on-failure

---
apiVersion: batch/v1
kind: Job
metadata:
  name: content-deployment-job
spec:
  template:
    spec:
      containers:
      - name: content-deployer
        image: mmorpg/content-deployer:latest
        env:
        - name: DEPLOYMENT_ID
          value: "{{ .Values.deploymentId }}"
        - name: CONTENT_BUCKET
          value: "s3://mmorpg-content"
        - name: NATS_URL
          value: "nats://nats-cluster:4222"
        command: ["/app/deploy-content"]
      restartPolicy: OnFailure
```

---

## 📊 Monitoring & Analytics Architecture

### Metrics Collection (Prometheus)

```go
// internal/monitoring/metrics/collector.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

type GameMetrics struct {
    // Player metrics
    PlayersOnline prometheus.Gauge
    PlayerLogins  prometheus.Counter
    PlayerLogouts prometheus.Counter
    
    // Performance metrics
    RequestDuration *prometheus.HistogramVec
    RequestTotal    *prometheus.CounterVec
    
    // Game metrics
    ItemsCreated    prometheus.Counter
    QuestsCompleted prometheus.Counter
    PvPMatches      prometheus.Counter
    ChatMessages    prometheus.Counter
    
    // Economic metrics
    GoldInCirculation prometheus.Gauge
    ItemsTraded       prometheus.Counter
    AuctionVolume     prometheus.Gauge
    
    // System metrics
    DBConnections   prometheus.Gauge
    CacheHitRatio   prometheus.Gauge
    MessageQueueLag prometheus.Gauge
}

func NewGameMetrics() *GameMetrics {
    return &GameMetrics{
        PlayersOnline: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "mmorpg_players_online_total",
            Help: "Current number of players online",
        }),
        
        PlayerLogins: promauto.NewCounter(prometheus.CounterOpts{
            Name: "mmorpg_player_logins_total",
            Help: "Total number of player logins",
        }),
        
        RequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
            Name:    "mmorpg_request_duration_seconds",
            Help:    "Request duration in seconds",
            Buckets: prometheus.DefBuckets,
        }, []string{"service", "method", "status"}),
        
        RequestTotal: promauto.NewCounterVec(prometheus.CounterOpts{
            Name: "mmorpg_requests_total",
            Help: "Total number of requests",
        }, []string{"service", "method", "status"}),
        
        ItemsCreated: promauto.NewCounter(prometheus.CounterOpts{
            Name: "mmorpg_items_created_total",
            Help: "Total number of items created",
        }),
        
        GoldInCirculation: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "mmorpg_gold_circulation_total",
            Help: "Total gold in circulation",
        }),
    }
}

// Custom game metrics collector
type PlayerStatsCollector struct {
    gameDB *sql.DB
}

func (c *PlayerStatsCollector) Describe(ch chan<- *prometheus.Desc) {
    prometheus.DescribeByCollect(c, ch)
}

func (c *PlayerStatsCollector) Collect(ch chan<- prometheus.Metric) {
    // Collect player distribution by level
    rows, err := c.gameDB.Query(`
        SELECT level, COUNT(*) as count 
        FROM players 
        WHERE last_login > NOW() - INTERVAL '7 days'
        GROUP BY level
    `)
    if err != nil {
        return
    }
    defer rows.Close()
    
    for rows.Next() {
        var level int
        var count int
        if err := rows.Scan(&level, &count); err != nil {
            continue
        }
        
        ch <- prometheus.MustNewConstMetric(
            prometheus.NewDesc(
                "mmorpg_player_level_distribution",
                "Distribution of active players by level",
                []string{"level"},
                nil,
            ),
            prometheus.GaugeValue,
            float64(count),
            fmt.Sprintf("%d", level),
        )
    }
}
```

### Log Aggregation (ELK Stack)

```yaml
# kubernetes/elk-stack.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: filebeat-config
data:
  filebeat.yml: |
    filebeat.inputs:
    - type: container
      paths:
        - /var/log/containers/*.log
      processors:
        - add_kubernetes_metadata:
            host: ${NODE_NAME}
            matchers:
            - logs_path:
                logs_path: "/var/log/containers/"
      multiline.pattern: '^[[:space:]]'
      multiline.negate: false
      multiline.match: after
    
    output.logstash:
      hosts: ['logstash:5044']

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: logstash-config
data:
  logstash.conf: |
    input {
      beats {
        port => 5044
      }
    }
    
    filter {
      if [kubernetes][labels][app] == "game-server" {
        grok {
          match => { 
            "message" => "%{TIMESTAMP_ISO8601:timestamp} %{LOGLEVEL:level} %{GREEDYDATA:message}" 
          }
        }
        
        if [message] =~ /player_action/ {
          grok {
            match => {
              "message" => "player_action: player_id=%{UUID:player_id} action=%{WORD:action} %{GREEDYDATA:details}"
            }
          }
          
          mutate {
            add_tag => [ "player_action" ]
          }
        }
      }
      
      date {
        match => [ "timestamp", "ISO8601" ]
        target => "@timestamp"
      }
    }
    
    output {
      elasticsearch {
        hosts => ["elasticsearch:9200"]
        index => "mmorpg-%{[kubernetes][labels][app]}-%{+YYYY.MM.dd}"
      }
    }
```

### Distributed Tracing (Jaeger)

```go
// internal/monitoring/tracing/middleware.go
package tracing

import (
    "github.com/gin-gonic/gin"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

func TracingMiddleware(serviceName string) gin.HandlerFunc {
    tracer := otel.Tracer(serviceName)
    
    return func(c *gin.Context) {
        ctx, span := tracer.Start(c.Request.Context(), 
            fmt.Sprintf("%s %s", c.Request.Method, c.FullPath()),
            trace.WithAttributes(
                attribute.String("http.method", c.Request.Method),
                attribute.String("http.url", c.Request.URL.String()),
                attribute.String("http.user_agent", c.Request.UserAgent()),
                attribute.String("user.id", c.GetString("user_id")),
            ),
        )
        defer span.End()
        
        c.Request = c.Request.WithContext(ctx)
        c.Next()
        
        span.SetAttributes(
            attribute.Int("http.status_code", c.Writer.Status()),
            attribute.Int64("http.response_size", int64(c.Writer.Size())),
        )
        
        if c.Writer.Status() >= 400 {
            span.RecordError(fmt.Errorf("HTTP %d", c.Writer.Status()))
        }
    }
}

// Database tracing wrapper
type TracedDB struct {
    db     *sql.DB
    tracer trace.Tracer
}

func (t *TracedDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    ctx, span := t.tracer.Start(ctx, "db.query",
        trace.WithAttributes(
            attribute.String("db.statement", query),
            attribute.Int("db.args_count", len(args)),
        ),
    )
    defer span.End()
    
    start := time.Now()
    rows, err := t.db.QueryContext(ctx, query, args...)
    
    span.SetAttributes(
        attribute.Int64("db.duration_ms", time.Since(start).Milliseconds()),
    )
    
    if err != nil {
        span.RecordError(err)
    }
    
    return rows, err
}
```

### Custom Game Metrics Design

```go
// internal/analytics/domain/metrics.go
package domain

type GameMetric struct {
    ID        string
    Type      MetricType
    PlayerID  string
    Value     float64
    Metadata  map[string]interface{}
    Timestamp time.Time
}

type MetricType string

const (
    MetricPlayerLevel      MetricType = "player_level"
    MetricQuestCompleted   MetricType = "quest_completed"
    MetricItemAcquired     MetricType = "item_acquired"
    MetricPvPMatch         MetricType = "pvp_match"
    MetricDungeonCompleted MetricType = "dungeon_completed"
    MetricGoldEarned       MetricType = "gold_earned"
    MetricDeathOccurred    MetricType = "death_occurred"
)

// internal/analytics/application/metrics_aggregator.go
package application

type MetricsAggregator struct {
    store     MetricStore
    publisher EventPublisher
}

func (a *MetricsAggregator) AggregatePlayerMetrics(ctx context.Context, playerID string) (*PlayerMetricsSummary, error) {
    metrics, err := a.store.GetPlayerMetrics(ctx, playerID, time.Now().Add(-30*24*time.Hour))
    if err != nil {
        return nil, err
    }
    
    summary := &PlayerMetricsSummary{
        PlayerID:         playerID,
        TotalPlaytime:    a.calculatePlaytime(metrics),
        QuestsCompleted:  a.countMetricType(metrics, MetricQuestCompleted),
        ItemsAcquired:    a.countMetricType(metrics, MetricItemAcquired),
        PvPWinRate:       a.calculatePvPWinRate(metrics),
        AverageGoldEarned: a.calculateAverageGold(metrics),
        DeathCount:       a.countMetricType(metrics, MetricDeathOccurred),
    }
    
    return summary, nil
}

// Real-time metrics streaming
type MetricsStreamer struct {
    kafka *kafka.Writer
}

func (s *MetricsStreamer) StreamMetric(ctx context.Context, metric GameMetric) error {
    data, err := json.Marshal(metric)
    if err != nil {
        return err
    }
    
    return s.kafka.WriteMessages(ctx, kafka.Message{
        Topic: "game-metrics",
        Key:   []byte(metric.PlayerID),
        Value: data,
        Headers: []kafka.Header{
            {Key: "metric_type", Value: []byte(string(metric.Type))},
            {Key: "timestamp", Value: []byte(metric.Timestamp.Format(time.RFC3339))},
        },
    })
}
```

---

## 🛠️ GM Tools Architecture

### GM Client Integration

```go
// internal/gm/domain/commands.go
package domain

type GMCommand interface {
    Execute(ctx context.Context, args []string) (string, error)
    Validate(args []string) error
    GetHelp() string
    GetRequiredPermission() string
}

type GMCommandRegistry struct {
    commands map[string]GMCommand
}

func NewGMCommandRegistry() *GMCommandRegistry {
    registry := &GMCommandRegistry{
        commands: make(map[string]GMCommand),
    }
    
    // Register all GM commands
    registry.Register("teleport", &TeleportCommand{})
    registry.Register("spawn", &SpawnCommand{})
    registry.Register("give", &GiveItemCommand{})
    registry.Register("ban", &BanCommand{})
    registry.Register("kick", &KickCommand{})
    registry.Register("announce", &AnnounceCommand{})
    
    return registry
}

// Example GM command implementation
type TeleportCommand struct {
    worldService WorldService
}

func (c *TeleportCommand) Execute(ctx context.Context, args []string) (string, error) {
    if len(args) < 4 {
        return "", fmt.Errorf("usage: /teleport <player> <x> <y> <z>")
    }
    
    player := args[0]
    x, _ := strconv.ParseFloat(args[1], 64)
    y, _ := strconv.ParseFloat(args[2], 64)
    z, _ := strconv.ParseFloat(args[3], 64)
    
    err := c.worldService.TeleportPlayer(ctx, player, Vector3{X: x, Y: y, Z: z})
    if err != nil {
        return "", err
    }
    
    return fmt.Sprintf("Teleported %s to (%.2f, %.2f, %.2f)", player, x, y, z), nil
}

func (c *TeleportCommand) GetRequiredPermission() string {
    return "gm.command.teleport"
}
```

### Privilege Escalation System

```go
// internal/gm/application/privilege_manager.go
package application

type PrivilegeManager struct {
    rbac   RBACService
    audit  AuditLogger
    cache  Cache
}

type GMPrivilege struct {
    Level       int
    Permissions []string
    Restrictions []string
}

var GMPrivilegeLevels = map[string]GMPrivilege{
    "gm_tier1": {
        Level: 1,
        Permissions: []string{
            "gm.command.kick",
            "gm.command.mute",
            "gm.command.teleport_self",
            "gm.view.player_info",
        },
        Restrictions: []string{
            "no_item_spawn",
            "no_permanent_bans",
        },
    },
    "gm_tier2": {
        Level: 2,
        Permissions: []string{
            "gm.command.*",
            "gm.view.*",
            "gm.modify.player_stats",
        },
        Restrictions: []string{
            "no_economy_manipulation",
        },
    },
    "gm_tier3": {
        Level: 3,
        Permissions: []string{
            "gm.*",
            "admin.server.restart",
            "admin.economy.modify",
        },
        Restrictions: []string{},
    },
}

func (p *PrivilegeManager) EscalatePrivileges(ctx context.Context, userID string, reason string) error {
    user, err := p.rbac.GetUser(ctx, userID)
    if err != nil {
        return err
    }
    
    // Check if user has GM role
    if !user.HasRole("gm_tier1") && !user.HasRole("gm_tier2") {
        return ErrNotGM
    }
    
    // Create temporary escalation
    escalation := &PrivilegeEscalation{
        UserID:    userID,
        FromLevel: user.GetGMLevel(),
        ToLevel:   user.GetGMLevel() + 1,
        Reason:    reason,
        ExpiresAt: time.Now().Add(2 * time.Hour),
    }
    
    // Store escalation
    key := fmt.Sprintf("escalation:%s", userID)
    p.cache.Set(ctx, key, escalation, 2*time.Hour)
    
    // Audit log
    p.audit.LogAction(ctx, AuditLog{
        UserID:       userID,
        Action:       "privilege_escalation",
        ResourceType: "gm_privileges",
        NewValue:     escalation,
    })
    
    return nil
}
```

### Action Logging and Replay

```go
// internal/gm/domain/action_log.go
package domain

type GMAction struct {
    ID          string
    GMID        string
    GMUsername  string
    Command     string
    Arguments   []string
    TargetType  string
    TargetID    string
    Result      string
    Error       string
    Timestamp   time.Time
    IPAddress   string
    Reversible  bool
    ReversedBy  *string
}

type GMActionLogger interface {
    LogAction(ctx context.Context, action GMAction) error
    GetActions(ctx context.Context, filter GMActionFilter) ([]GMAction, error)
    ReplayAction(ctx context.Context, actionID string) error
    ReverseAction(ctx context.Context, actionID string, reversedBy string) error
}

// internal/gm/infrastructure/action_logger.go
package infrastructure

type postgresGMActionLogger struct {
    db       *sql.DB
    executor CommandExecutor
}

func (l *postgresGMActionLogger) LogAction(ctx context.Context, action domain.GMAction) error {
    query := `
        INSERT INTO gm_actions 
        (id, gm_id, gm_username, command, arguments, target_type, target_id, 
         result, error, timestamp, ip_address, reversible)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `
    
    argsJSON, _ := json.Marshal(action.Arguments)
    
    _, err := l.db.ExecContext(ctx, query,
        action.ID,
        action.GMID,
        action.GMUsername,
        action.Command,
        argsJSON,
        action.TargetType,
        action.TargetID,
        action.Result,
        action.Error,
        action.Timestamp,
        action.IPAddress,
        action.Reversible,
    )
    
    return err
}

func (l *postgresGMActionLogger) ReplayAction(ctx context.Context, actionID string) error {
    // Fetch original action
    var action domain.GMAction
    query := `SELECT * FROM gm_actions WHERE id = $1`
    
    row := l.db.QueryRowContext(ctx, query, actionID)
    if err := scanGMAction(row, &action); err != nil {
        return err
    }
    
    // Check if action is replayable
    if !isReplayableCommand(action.Command) {
        return ErrActionNotReplayable
    }
    
    // Execute command again
    result, err := l.executor.Execute(ctx, action.Command, action.Arguments)
    
    // Log the replay
    replayAction := action
    replayAction.ID = uuid.New().String()
    replayAction.Result = result
    replayAction.Error = ""
    if err != nil {
        replayAction.Error = err.Error()
    }
    
    return l.LogAction(ctx, replayAction)
}

// Action reversal for certain commands
func (l *postgresGMActionLogger) ReverseAction(ctx context.Context, actionID string, reversedBy string) error {
    var action domain.GMAction
    query := `SELECT * FROM gm_actions WHERE id = $1 AND reversible = true AND reversed_by IS NULL`
    
    row := l.db.QueryRowContext(ctx, query, actionID)
    if err := scanGMAction(row, &action); err != nil {
        return err
    }
    
    // Generate reverse command
    reverseCmd, reverseArgs := generateReverseCommand(action.Command, action.Arguments)
    if reverseCmd == "" {
        return ErrActionNotReversible
    }
    
    // Execute reverse command
    result, err := l.executor.Execute(ctx, reverseCmd, reverseArgs)
    if err != nil {
        return err
    }
    
    // Mark original action as reversed
    updateQuery := `UPDATE gm_actions SET reversed_by = $1 WHERE id = $2`
    _, err = l.db.ExecContext(ctx, updateQuery, reversedBy, actionID)
    
    // Log the reversal
    reversalAction := domain.GMAction{
        ID:         uuid.New().String(),
        GMID:       reversedBy,
        Command:    reverseCmd,
        Arguments:  reverseArgs,
        Result:     result,
        Timestamp:  time.Now(),
    }
    
    return l.LogAction(ctx, reversalAction)
}
```

### Real-time Game Manipulation

```go
// internal/gm/application/realtime_tools.go
package application

type RealtimeGMTools struct {
    gameState   GameStateManager
    wsHub       *WebSocketHub
    permissions PermissionChecker
}

func (t *RealtimeGMTools) InspectPlayer(ctx context.Context, gmID, playerID string) (*PlayerInspection, error) {
    // Check permissions
    if !t.permissions.Can(gmID, "gm.inspect.player") {
        return nil, ErrUnauthorized
    }
    
    player, err := t.gameState.GetPlayer(playerID)
    if err != nil {
        return nil, err
    }
    
    inspection := &PlayerInspection{
        PlayerID:    playerID,
        Username:    player.Username,
        Level:       player.Level,
        Location:    player.Location,
        Stats:       player.Stats,
        Inventory:   player.Inventory,
        ActiveBuffs: player.Buffs,
        QuestLog:    player.QuestLog,
        RecentChat:  t.gameState.GetRecentChat(playerID, 50),
    }
    
    return inspection, nil
}

func (t *RealtimeGMTools) ModifyPlayerStats(ctx context.Context, gmID, playerID string, mods StatModifications) error {
    // Validate permissions
    if !t.permissions.Can(gmID, "gm.modify.stats") {
        return ErrUnauthorized
    }
    
    // Apply modifications with validation
    player, err := t.gameState.GetPlayer(playerID)
    if err != nil {
        return err
    }
    
    // Create backup for potential rollback
    backup := player.Stats.Clone()
    
    // Apply modifications
    if mods.Level != nil {
        player.Stats.Level = *mods.Level
    }
    if mods.Health != nil {
        player.Stats.Health = *mods.Health
    }
    if mods.Experience != nil {
        player.Stats.Experience = *mods.Experience
    }
    
    // Validate new stats
    if err := player.Stats.Validate(); err != nil {
        player.Stats = backup // Rollback
        return err
    }
    
    // Apply to game state
    if err := t.gameState.UpdatePlayer(player); err != nil {
        return err
    }
    
    // Notify player client
    t.wsHub.SendToPlayer(playerID, &StatsUpdateMessage{
        Stats:  player.Stats,
        Source: "gm_modification",
    })
    
    return nil
}

// Real-time monitoring dashboard
type GMMonitoringService struct {
    metrics  MetricsCollector
    gameState GameStateManager
    wsHub    *WebSocketHub
}

func (s *GMMonitoringService) StreamServerMetrics(ctx context.Context, gmID string) error {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    conn := s.wsHub.GetConnection(gmID)
    if conn == nil {
        return ErrNotConnected
    }
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
            
        case <-ticker.C:
            metrics := &ServerMetrics{
                Timestamp:     time.Now(),
                PlayersOnline: s.gameState.GetOnlinePlayerCount(),
                ServerFPS:     s.metrics.GetServerFPS(),
                MemoryUsage:   s.metrics.GetMemoryUsage(),
                CPUUsage:      s.metrics.GetCPUUsage(),
                ActiveZones:   s.gameState.GetActiveZones(),
                QueuedEvents:  s.gameState.GetEventQueueSize(),
            }
            
            if err := conn.WriteJSON(metrics); err != nil {
                return err
            }
        }
    }
}
```

---

## 🚀 Deployment Architecture

### Kubernetes Cluster Design

```yaml
# kubernetes/cluster-architecture.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: mmorpg-prod
  labels:
    environment: production
    
---
apiVersion: v1
kind: Namespace
metadata:
  name: mmorpg-staging
  labels:
    environment: staging

---
# Production node pool configuration
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-config
  namespace: mmorpg-prod
data:
  node-pools.yaml: |
    nodePools:
      - name: game-servers
        machineType: n2-standard-8
        minNodes: 3
        maxNodes: 20
        labels:
          workload: game-server
        taints:
          - key: game-server
            value: "true"
            effect: NoSchedule
            
      - name: admin-services
        machineType: n2-standard-4
        minNodes: 2
        maxNodes: 5
        labels:
          workload: admin
          
      - name: database
        machineType: n2-highmem-4
        minNodes: 3
        maxNodes: 3
        labels:
          workload: database
        taints:
          - key: database
            value: "true"
            effect: NoSchedule
```

```go
// internal/deployment/k8s/manager.go
package k8s

import (
    "k8s.io/client-go/kubernetes"
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
)

type DeploymentManager struct {
    client    kubernetes.Interface
    namespace string
}

func (m *DeploymentManager) DeployService(ctx context.Context, spec ServiceSpec) error {
    // Create deployment
    deployment := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      spec.Name,
            Namespace: m.namespace,
            Labels:    spec.Labels,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &spec.Replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: spec.Labels,
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: spec.Labels,
                    Annotations: map[string]string{
                        "prometheus.io/scrape": "true",
                        "prometheus.io/port":   "9090",
                    },
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  spec.Name,
                            Image: spec.Image,
                            Ports: spec.Ports,
                            Env:   spec.EnvVars,
                            Resources: corev1.ResourceRequirements{
                                Requests: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse(spec.CPURequest),
                                    corev1.ResourceMemory: resource.MustParse(spec.MemoryRequest),
                                },
                                Limits: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse(spec.CPULimit),
                                    corev1.ResourceMemory: resource.MustParse(spec.MemoryLimit),
                                },
                            },
                            LivenessProbe: &corev1.Probe{
                                Handler: corev1.Handler{
                                    HTTPGet: &corev1.HTTPGetAction{
                                        Path: "/health",
                                        Port: intstr.FromInt(8080),
                                    },
                                },
                                InitialDelaySeconds: 30,
                                PeriodSeconds:       10,
                            },
                            ReadinessProbe: &corev1.Probe{
                                Handler: corev1.Handler{
                                    HTTPGet: &corev1.HTTPGetAction{
                                        Path: "/ready",
                                        Port: intstr.FromInt(8080),
                                    },
                                },
                                InitialDelaySeconds: 5,
                                PeriodSeconds:       5,
                            },
                        },
                    },
                },
            },
        },
    }
    
    _, err := m.client.AppsV1().Deployments(m.namespace).Create(ctx, deployment, metav1.CreateOptions{})
    return err
}

func (m *DeploymentManager) CreateHorizontalPodAutoscaler(ctx context.Context, name string, minReplicas, maxReplicas int32) error {
    hpa := &autoscalingv2.HorizontalPodAutoscaler{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name + "-hpa",
            Namespace: m.namespace,
        },
        Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
            ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
                APIVersion: "apps/v1",
                Kind:       "Deployment",
                Name:       name,
            },
            MinReplicas: &minReplicas,
            MaxReplicas: maxReplicas,
            Metrics: []autoscalingv2.MetricSpec{
                {
                    Type: autoscalingv2.ResourceMetricSourceType,
                    Resource: &autoscalingv2.ResourceMetricSource{
                        Name: corev1.ResourceCPU,
                        Target: autoscalingv2.MetricTarget{
                            Type:               autoscalingv2.UtilizationMetricType,
                            AverageUtilization: &[]int32{70}[0],
                        },
                    },
                },
                {
                    Type: autoscalingv2.PodsMetricSourceType,
                    Pods: &autoscalingv2.PodsMetricSource{
                        Metric: autoscalingv2.MetricIdentifier{
                            Name: "mmorpg_players_per_pod",
                        },
                        Target: autoscalingv2.MetricTarget{
                            Type:         autoscalingv2.AverageValueMetricType,
                            AverageValue: resource.NewQuantity(100, resource.DecimalSI),
                        },
                    },
                },
            },
        },
    }
    
    _, err := m.client.AutoscalingV2().HorizontalPodAutoscalers(m.namespace).Create(ctx, hpa, metav1.CreateOptions{})
    return err
}
```

### CI/CD Pipeline Architecture

```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - security
  - deploy-staging
  - integration-tests
  - deploy-production

variables:
  DOCKER_REGISTRY: gcr.io/mmorpg-project
  GO_VERSION: "1.21"
  
test:unit:
  stage: test
  image: golang:${GO_VERSION}
  script:
    - go test -v -coverprofile=coverage.out ./...
    - go tool cover -html=coverage.out -o coverage.html
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
    paths:
      - coverage.html

test:integration:
  stage: test
  services:
    - postgres:14
    - redis:7
  script:
    - go test -v -tags=integration ./...

build:services:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  script:
    - |
      for service in admin-api analytics-api content-api gm-api; do
        docker build -t ${DOCKER_REGISTRY}/${service}:${CI_COMMIT_SHA} -f cmd/${service}/Dockerfile .
        docker push ${DOCKER_REGISTRY}/${service}:${CI_COMMIT_SHA}
      done

security:scan:
  stage: security
  image: aquasec/trivy:latest
  script:
    - |
      for service in admin-api analytics-api content-api gm-api; do
        trivy image --severity HIGH,CRITICAL ${DOCKER_REGISTRY}/${service}:${CI_COMMIT_SHA}
      done

deploy:staging:
  stage: deploy-staging
  image: google/cloud-sdk:latest
  environment:
    name: staging
    url: https://staging.admin.mmorpg.com
  script:
    - gcloud auth activate-service-account --key-file=$GCP_SERVICE_KEY
    - gcloud container clusters get-credentials staging-cluster --zone=us-central1-a
    - helm upgrade --install mmorpg-admin ./helm/admin-services
      --namespace=mmorpg-staging
      --set image.tag=${CI_COMMIT_SHA}
      --values=helm/values/staging.yaml
  only:
    - develop

integration:tests:
  stage: integration-tests
  script:
    - npm install -g @playwright/test
    - playwright test --config=e2e/playwright.config.ts
  environment:
    name: staging

deploy:production:
  stage: deploy-production
  image: google/cloud-sdk:latest
  environment:
    name: production
    url: https://admin.mmorpg.com
  script:
    - gcloud auth activate-service-account --key-file=$GCP_SERVICE_KEY_PROD
    - gcloud container clusters get-credentials prod-cluster --zone=us-central1-a
    - helm upgrade --install mmorpg-admin ./helm/admin-services
      --namespace=mmorpg-prod
      --set image.tag=${CI_COMMIT_SHA}
      --values=helm/values/production.yaml
      --atomic
      --timeout=10m
  only:
    - master
  when: manual
```

### Blue-Green Deployment Strategy

```go
// internal/deployment/bluegreen/strategy.go
package bluegreen

type BlueGreenDeployer struct {
    k8s          kubernetes.Interface
    loadBalancer LoadBalancerManager
    healthCheck  HealthChecker
}

func (d *BlueGreenDeployer) Deploy(ctx context.Context, spec DeploymentSpec) error {
    // Determine current active color
    activeColor, err := d.getActiveColor(spec.Service)
    if err != nil {
        return err
    }
    
    inactiveColor := "blue"
    if activeColor == "blue" {
        inactiveColor = "green"
    }
    
    // Deploy to inactive environment
    deploymentName := fmt.Sprintf("%s-%s", spec.Service, inactiveColor)
    
    deployment := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      deploymentName,
            Namespace: spec.Namespace,
            Labels: map[string]string{
                "app":   spec.Service,
                "color": inactiveColor,
            },
        },
        Spec: spec.DeploymentSpec,
    }
    
    // Create or update deployment
    _, err = d.k8s.AppsV1().Deployments(spec.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
    if err != nil {
        _, err = d.k8s.AppsV1().Deployments(spec.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
        if err != nil {
            return err
        }
    }
    
    // Wait for deployment to be ready
    if err := d.waitForDeployment(ctx, deploymentName, spec.Namespace); err != nil {
        return err
    }
    
    // Run health checks
    if err := d.healthCheck.CheckDeployment(ctx, deploymentName, spec.Namespace); err != nil {
        return fmt.Errorf("health check failed: %w", err)
    }
    
    // Switch traffic
    if err := d.switchTraffic(ctx, spec.Service, inactiveColor); err != nil {
        return err
    }
    
    // Wait for traffic to stabilize
    time.Sleep(30 * time.Second)
    
    // Final health check
    if err := d.healthCheck.CheckService(ctx, spec.Service, spec.Namespace); err != nil {
        // Rollback
        d.switchTraffic(ctx, spec.Service, activeColor)
        return fmt.Errorf("post-switch health check failed, rolled back: %w", err)
    }
    
    // Scale down old deployment
    oldDeployment := fmt.Sprintf("%s-%s", spec.Service, activeColor)
    scale := &autoscalingv1.Scale{
        Spec: autoscalingv1.ScaleSpec{
            Replicas: 0,
        },
    }
    
    _, err = d.k8s.AppsV1().Deployments(spec.Namespace).UpdateScale(ctx, oldDeployment, scale, metav1.UpdateOptions{})
    
    return err
}

func (d *BlueGreenDeployer) switchTraffic(ctx context.Context, service, color string) error {
    svc, err := d.k8s.CoreV1().Services(service).Get(ctx, service, metav1.GetOptions{})
    if err != nil {
        return err
    }
    
    // Update selector to point to new color
    svc.Spec.Selector = map[string]string{
        "app":   service,
        "color": color,
    }
    
    _, err = d.k8s.CoreV1().Services(service).Update(ctx, svc, metav1.UpdateOptions{})
    return err
}
```

### Database Migration Patterns

```go
// internal/deployment/migrations/manager.go
package migrations

import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

type MigrationManager struct {
    migrate     *migrate.Migrate
    lockManager LockManager
}

func (m *MigrationManager) RunMigrations(ctx context.Context) error {
    // Acquire distributed lock
    lock, err := m.lockManager.AcquireLock(ctx, "db-migration", 5*time.Minute)
    if err != nil {
        return fmt.Errorf("failed to acquire migration lock: %w", err)
    }
    defer lock.Release()
    
    // Check current version
    version, dirty, err := m.migrate.Version()
    if err != nil && err != migrate.ErrNilVersion {
        return err
    }
    
    if dirty {
        return fmt.Errorf("database is in dirty state at version %d", version)
    }
    
    // Run migrations
    if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    
    newVersion, _, _ := m.migrate.Version()
    log.Printf("Migrated database from version %d to %d", version, newVersion)
    
    return nil
}

// Zero-downtime migration patterns
type ZeroDowntimeMigrator struct {
    db          *sql.DB
    schemaCheck SchemaChecker
}

func (z *ZeroDowntimeMigrator) AddColumnWithDefault(ctx context.Context, table, column, dataType string, defaultValue interface{}) error {
    // Step 1: Add column as nullable
    if _, err := z.db.ExecContext(ctx, 
        fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s", table, column, dataType),
    ); err != nil {
        return err
    }
    
    // Step 2: Backfill existing rows in batches
    batchSize := 1000
    offset := 0
    
    for {
        result, err := z.db.ExecContext(ctx,
            fmt.Sprintf(`UPDATE %s SET %s = $1 
                        WHERE %s IS NULL 
                        AND id IN (SELECT id FROM %s WHERE %s IS NULL LIMIT $2)`,
                table, column, column, table, column),
            defaultValue, batchSize,
        )
        
        if err != nil {
            return err
        }
        
        affected, _ := result.RowsAffected()
        if affected == 0 {
            break
        }
        
        // Small delay to reduce load
        time.Sleep(100 * time.Millisecond)
    }
    
    // Step 3: Add NOT NULL constraint
    if _, err := z.db.ExecContext(ctx,
        fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL", table, column),
    ); err != nil {
        return err
    }
    
    return nil
}

// Migration validation
func (z *ZeroDowntimeMigrator) ValidateMigration(ctx context.Context, migration Migration) error {
    // Check if migration is backwards compatible
    if !migration.IsBackwardsCompatible() {
        return fmt.Errorf("migration %s is not backwards compatible", migration.Name)
    }
    
    // Validate against current schema
    currentSchema, err := z.schemaCheck.GetCurrentSchema(ctx)
    if err != nil {
        return err
    }
    
    if err := migration.Validate(currentSchema); err != nil {
        return fmt.Errorf("migration validation failed: %w", err)
    }
    
    return nil
}
```

---

## 📚 Infrastructure as Code Examples

### Terraform Configuration

```hcl
# terraform/production/main.tf
terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
  
  backend "gcs" {
    bucket = "mmorpg-terraform-state"
    prefix = "production"
  }
}

module "gke_cluster" {
  source = "../modules/gke"
  
  project_id     = var.project_id
  region         = var.region
  cluster_name   = "mmorpg-production"
  
  node_pools = [
    {
      name               = "game-servers"
      machine_type       = "n2-standard-8"
      min_count          = 3
      max_count          = 20
      disk_size_gb       = 100
      disk_type          = "pd-ssd"
      preemptible        = false
      auto_repair        = true
      auto_upgrade       = true
    },
    {
      name               = "admin-services"
      machine_type       = "n2-standard-4"
      min_count          = 2
      max_count          = 5
      disk_size_gb       = 50
      disk_type          = "pd-standard"
      preemptible        = true
      auto_repair        = true
      auto_upgrade       = true
    }
  ]
  
  master_authorized_networks = [
    {
      display_name = "Office"
      cidr_block   = var.office_cidr
    }
  ]
}

module "cloudsql" {
  source = "../modules/cloudsql"
  
  project_id      = var.project_id
  region          = var.region
  instance_name   = "mmorpg-production-db"
  database_version = "POSTGRES_14"
  tier            = "db-custom-8-32768"
  
  high_availability = true
  backup_enabled    = true
  backup_start_time = "03:00"
  
  databases = [
    "mmorpg_game",
    "mmorpg_admin",
    "mmorpg_analytics"
  ]
}

module "monitoring" {
  source = "../modules/monitoring"
  
  project_id = var.project_id
  
  alert_email = var.ops_team_email
  
  alerts = [
    {
      name        = "high-cpu-usage"
      condition   = "cpu_usage > 0.8"
      duration    = "5m"
      severity    = "warning"
    },
    {
      name        = "database-connection-pool-exhausted"
      condition   = "db_connections_available < 5"
      duration    = "1m"
      severity    = "critical"
    }
  ]
}
```

### Helm Charts

```yaml
# helm/admin-services/values.yaml
global:
  image:
    registry: gcr.io/mmorpg-project
    pullPolicy: IfNotPresent
  
  ingress:
    enabled: true
    className: nginx
    tls:
      enabled: true
      issuer: letsencrypt-prod

adminApi:
  replicaCount: 3
  image:
    repository: admin-api
    tag: latest
  
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: 1000m
      memory: 1Gi
  
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70
  
  env:
    - name: DATABASE_URL
      valueFrom:
        secretKeyRef:
          name: database-credentials
          key: admin-db-url
    - name: REDIS_URL
      value: redis://redis-master:6379
    - name: JWT_SECRET
      valueFrom:
        secretKeyRef:
          name: jwt-secrets
          key: secret

monitoring:
  prometheus:
    enabled: true
    serviceMonitor:
      enabled: true
      interval: 30s
  
  grafana:
    enabled: true
    dashboards:
      - admin-overview
      - api-performance
      - error-rates
```

---

## 🔍 Summary

This architecture document provides a comprehensive blueprint for Phase 4's production tools and infrastructure. Key architectural decisions include:

1. **Hexagonal Architecture**: Clear separation of business logic from infrastructure
2. **Event-Driven Design**: Loose coupling between services via message queues
3. **Zero-Trust Security**: Multiple layers of authentication and authorization
4. **Observable Systems**: Comprehensive monitoring, logging, and tracing
5. **Automated Operations**: CI/CD pipelines and infrastructure as code

The architecture supports the scalability, reliability, and maintainability requirements of a production MMORPG service while providing powerful tools for operations teams to manage the live game effectively.

---

*This architecture document serves as the technical blueprint for implementing Phase 4's production tools and infrastructure, ensuring a robust and scalable live service operation.*