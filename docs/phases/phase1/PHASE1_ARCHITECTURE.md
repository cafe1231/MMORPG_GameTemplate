# 🏗️ Phase 1: Authentication System - Architecture Document

## 📋 Executive Summary

This document details the technical architecture for Phase 1's authentication system implementation. Building on Phase 0's foundation, we implement a secure, scalable authentication system using JWT tokens, hexagonal architecture on the backend, and a robust subsystem architecture on the frontend. The system supports user registration, login, session management, and character selection.

**Key Architectural Decisions:**
- JWT-based authentication with refresh tokens
- Hexagonal architecture for the auth service
- Redis for session caching with PostgreSQL persistence
- UE5 subsystem pattern for client-side auth management
- Event-driven architecture with NATS messaging

---

## 🔐 1. Backend Authentication Architecture

### 1.1 Hexagonal Architecture Implementation

```go
// internal/domain/auth/user.go
package auth

import (
    "time"
    "regexp"
    "golang.org/x/crypto/bcrypt"
)

// Domain entity
type User struct {
    ID           string
    Username     string
    Email        string
    PasswordHash string
    CreatedAt    time.Time
    UpdatedAt    time.Time
    LastLoginAt  *time.Time
    IsActive     bool
    IsVerified   bool
}

// Domain value objects
type Password struct {
    value string
}

func NewPassword(plaintext string) (*Password, error) {
    if err := validatePassword(plaintext); err != nil {
        return nil, err
    }
    return &Password{value: plaintext}, nil
}

func (p *Password) Hash() (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(p.value), 12)
    return string(bytes), err
}

type Email struct {
    value string
}

func NewEmail(email string) (*Email, error) {
    if !isValidEmail(email) {
        return nil, ErrInvalidEmail
    }
    return &Email{value: email}, nil
}

// Domain service
type UserService struct {
    repo UserRepository
}

func (s *UserService) CreateUser(username, email, password string) (*User, error) {
    // Validate username uniqueness
    if exists, err := s.repo.UsernameExists(username); err != nil {
        return nil, err
    } else if exists {
        return nil, ErrUsernameExists
    }
    
    // Validate email
    emailVO, err := NewEmail(email)
    if err != nil {
        return nil, err
    }
    
    // Hash password
    passwordVO, err := NewPassword(password)
    if err != nil {
        return nil, err
    }
    
    hash, err := passwordVO.Hash()
    if err != nil {
        return nil, err
    }
    
    // Create user
    user := &User{
        ID:           generateUserID(),
        Username:     username,
        Email:        emailVO.value,
        PasswordHash: hash,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
        IsActive:     true,
        IsVerified:   false,
    }
    
    return s.repo.Create(user)
}
```

### 1.2 JWT Token Service

```go
// internal/domain/auth/token.go
package auth

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
    accessSecret  []byte
    refreshSecret []byte
    accessTTL     time.Duration
    refreshTTL    time.Duration
}

type Claims struct {
    jwt.RegisteredClaims
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Type     string `json:"type"` // "access" or "refresh"
}

type TokenPair struct {
    AccessToken  string
    RefreshToken string
    ExpiresIn    int64
}

func (s *TokenService) GenerateTokenPair(user *User) (*TokenPair, error) {
    // Generate access token
    accessClaims := &Claims{
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   user.ID,
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "mmorpg-auth",
        },
        UserID:   user.ID,
        Username: user.Username,
        Type:     "access",
    }
    
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessString, err := accessToken.SignedString(s.accessSecret)
    if err != nil {
        return nil, err
    }
    
    // Generate refresh token
    refreshClaims := &Claims{
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   user.ID,
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshTTL)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "mmorpg-auth",
        },
        UserID:   user.ID,
        Username: user.Username,
        Type:     "refresh",
    }
    
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshString, err := refreshToken.SignedString(s.refreshSecret)
    if err != nil {
        return nil, err
    }
    
    return &TokenPair{
        AccessToken:  accessString,
        RefreshToken: refreshString,
        ExpiresIn:    int64(s.accessTTL.Seconds()),
    }, nil
}

func (s *TokenService) ValidateAccessToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, ErrInvalidSigningMethod
        }
        return s.accessSecret, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, ErrInvalidToken
    }
    
    if claims.Type != "access" {
        return nil, ErrWrongTokenType
    }
    
    return claims, nil
}
```

### 1.3 Session Management

```go
// internal/domain/auth/session.go
package auth

import (
    "time"
    "encoding/json"
)

type Session struct {
    ID           string
    UserID       string
    RefreshToken string
    UserAgent    string
    IPAddress    string
    CreatedAt    time.Time
    ExpiresAt    time.Time
    LastUsedAt   time.Time
}

type SessionService struct {
    repo  SessionRepository
    cache SessionCache
}

func (s *SessionService) CreateSession(userID, refreshToken, userAgent, ip string) (*Session, error) {
    session := &Session{
        ID:           generateSessionID(),
        UserID:       userID,
        RefreshToken: hashToken(refreshToken),
        UserAgent:    userAgent,
        IPAddress:    ip,
        CreatedAt:    time.Now(),
        ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
        LastUsedAt:   time.Now(),
    }
    
    // Save to database
    if err := s.repo.Create(session); err != nil {
        return nil, err
    }
    
    // Cache in Redis
    if err := s.cache.Set(session); err != nil {
        // Log error but don't fail
        logger.Error("Failed to cache session", "error", err)
    }
    
    return session, nil
}

func (s *SessionService) ValidateSession(sessionID string) (*Session, error) {
    // Try cache first
    session, err := s.cache.Get(sessionID)
    if err == nil && session != nil {
        return session, nil
    }
    
    // Fall back to database
    session, err = s.repo.FindByID(sessionID)
    if err != nil {
        return nil, err
    }
    
    // Check expiration
    if time.Now().After(session.ExpiresAt) {
        return nil, ErrSessionExpired
    }
    
    // Update last used
    session.LastUsedAt = time.Now()
    s.repo.Update(session)
    
    // Re-cache
    s.cache.Set(session)
    
    return session, nil
}

// Redis cache implementation
type RedisSessionCache struct {
    client *redis.Client
    ttl    time.Duration
}

func (c *RedisSessionCache) Set(session *Session) error {
    data, err := json.Marshal(session)
    if err != nil {
        return err
    }
    
    key := fmt.Sprintf("session:%s", session.ID)
    return c.client.Set(context.Background(), key, data, c.ttl).Err()
}

func (c *RedisSessionCache) Get(sessionID string) (*Session, error) {
    key := fmt.Sprintf("session:%s", sessionID)
    data, err := c.client.Get(context.Background(), key).Bytes()
    if err != nil {
        if err == redis.Nil {
            return nil, nil
        }
        return nil, err
    }
    
    var session Session
    if err := json.Unmarshal(data, &session); err != nil {
        return nil, err
    }
    
    return &session, nil
}
```

### 1.4 HTTP API Implementation

```go
// internal/adapters/auth/http_handler.go
package auth

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type HTTPHandler struct {
    authService    ports.AuthService
    tokenService   ports.TokenService
    sessionService ports.SessionService
    rateLimiter    ports.RateLimiter
}

func (h *HTTPHandler) RegisterRoutes(router *gin.RouterGroup) {
    router.POST("/register", h.rateLimiter.Limit("register"), h.Register)
    router.POST("/login", h.rateLimiter.Limit("login"), h.Login)
    router.POST("/refresh", h.Refresh)
    router.POST("/logout", h.AuthMiddleware(), h.Logout)
    router.GET("/verify", h.AuthMiddleware(), h.Verify)
}

// Register endpoint
func (h *HTTPHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Error: "Invalid request format",
        })
        return
    }
    
    // Validate input
    if err := h.validateRegisterRequest(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Error: err.Error(),
        })
        return
    }
    
    // Create user
    user, err := h.authService.Register(c.Request.Context(), &auth.RegisterInput{
        Username: req.Username,
        Email:    req.Email,
        Password: req.Password,
    })
    
    if err != nil {
        status, message := h.mapError(err)
        c.JSON(status, ErrorResponse{
            Error: message,
        })
        return
    }
    
    // Generate tokens
    tokens, err := h.tokenService.GenerateTokenPair(user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error: "Failed to generate tokens",
        })
        return
    }
    
    // Create session
    session, err := h.sessionService.CreateSession(
        user.ID,
        tokens.RefreshToken,
        c.GetHeader("User-Agent"),
        c.ClientIP(),
    )
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error: "Failed to create session",
        })
        return
    }
    
    c.JSON(http.StatusCreated, RegisterResponse{
        User: UserDTO{
            ID:       user.ID,
            Username: user.Username,
            Email:    user.Email,
        },
        Tokens: TokenDTO{
            AccessToken:  tokens.AccessToken,
            RefreshToken: tokens.RefreshToken,
            ExpiresIn:    tokens.ExpiresIn,
        },
        SessionID: session.ID,
    })
}

// Login endpoint
func (h *HTTPHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Error: "Invalid request format",
        })
        return
    }
    
    // Authenticate user
    user, err := h.authService.Login(c.Request.Context(), &auth.LoginInput{
        Username: req.Username,
        Password: req.Password,
    })
    
    if err != nil {
        // Track failed attempts
        h.rateLimiter.TrackFailedLogin(c.ClientIP(), req.Username)
        
        status, message := h.mapError(err)
        c.JSON(status, ErrorResponse{
            Error: message,
        })
        return
    }
    
    // Check if account is locked
    if locked, until := h.rateLimiter.IsAccountLocked(user.ID); locked {
        c.JSON(http.StatusTooManyRequests, ErrorResponse{
            Error: fmt.Sprintf("Account locked until %v", until),
        })
        return
    }
    
    // Generate tokens and create session (same as register)
    // ...
}

// Auth middleware
func (h *HTTPHandler) AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token from header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, ErrorResponse{
                Error: "Missing authorization header",
            })
            c.Abort()
            return
        }
        
        // Validate Bearer prefix
        const bearerPrefix = "Bearer "
        if !strings.HasPrefix(authHeader, bearerPrefix) {
            c.JSON(http.StatusUnauthorized, ErrorResponse{
                Error: "Invalid authorization format",
            })
            c.Abort()
            return
        }
        
        token := authHeader[len(bearerPrefix):]
        
        // Validate token
        claims, err := h.tokenService.ValidateAccessToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, ErrorResponse{
                Error: "Invalid or expired token",
            })
            c.Abort()
            return
        }
        
        // Set user context
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Next()
    }
}
```

### 1.5 Database Schema

```sql
-- migrations/002_create_users_table.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMP,
    
    CONSTRAINT username_valid CHECK (username ~ '^[a-zA-Z0-9_]{3,50}$'),
    CONSTRAINT email_valid CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_last_login ON users(last_login_at);

-- migrations/003_create_sessions_table.sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL,
    user_agent TEXT,
    ip_address INET,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    last_used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_expiry CHECK (expires_at > created_at)
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token_hash);

-- migrations/004_create_characters_table.sql
CREATE TABLE characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(50) UNIQUE NOT NULL,
    level INT NOT NULL DEFAULT 1,
    class VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_played_at TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    
    CONSTRAINT name_valid CHECK (name ~ '^[a-zA-Z]{3,50}$'),
    CONSTRAINT level_valid CHECK (level >= 1 AND level <= 100)
);

CREATE INDEX idx_characters_user_id ON characters(user_id);
CREATE INDEX idx_characters_name ON characters(name);
```

---

## 🎮 2. Frontend Authentication Architecture

### 2.1 Auth Subsystem Implementation

```cpp
// Source/MMORPGCore/Public/Auth/MMORPGAuthSubsystem.h
#pragma once

#include "CoreMinimal.h"
#include "Subsystems/GameInstanceSubsystem.h"
#include "MMORPGAuthTypes.h"
#include "MMORPGAuthSubsystem.generated.h"

DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnAuthStateChanged, EAuthState, NewState);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnLoginSuccess, const FAuthTokens&, Tokens);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnLoginFailure, const FMMORPGError&, Error);

UENUM(BlueprintType)
enum class EAuthState : uint8
{
    NotAuthenticated    UMETA(DisplayName = "Not Authenticated"),
    Authenticating      UMETA(DisplayName = "Authenticating"),
    Authenticated       UMETA(DisplayName = "Authenticated"),
    RefreshingToken     UMETA(DisplayName = "Refreshing Token"),
    Error               UMETA(DisplayName = "Error")
};

USTRUCT(BlueprintType)
struct FAuthTokens
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadOnly)
    FString AccessToken;

    UPROPERTY(BlueprintReadOnly)
    FString RefreshToken;

    UPROPERTY(BlueprintReadOnly)
    int32 ExpiresIn = 0;

    UPROPERTY(BlueprintReadOnly)
    FDateTime ExpiresAt;
};

UCLASS()
class MMORGCORE_API UMMORPGAuthSubsystem : public UGameInstanceSubsystem
{
    GENERATED_BODY()

public:
    // Events
    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Auth")
    FOnAuthStateChanged OnAuthStateChanged;

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Auth")
    FOnLoginSuccess OnLoginSuccess;

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Auth")
    FOnLoginFailure OnLoginFailure;

    // Authentication methods
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    void Login(const FString& Username, const FString& Password);

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    void Register(const FString& Username, const FString& Email, const FString& Password);

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    void Logout();

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    void RefreshToken();

    // State queries
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    bool IsAuthenticated() const;

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    FString GetUsername() const;

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    FString GetUserID() const;

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    EAuthState GetAuthState() const { return CurrentState; }

    // Token management
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    FString GetAccessToken() const;

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    bool IsTokenExpired() const;

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Auth")
    float GetTimeUntilTokenExpiry() const;

protected:
    virtual void Initialize(FSubsystemCollectionBase& Collection) override;
    virtual void Deinitialize() override;

private:
    UPROPERTY()
    EAuthState CurrentState = EAuthState::NotAuthenticated;

    UPROPERTY()
    FAuthTokens CurrentTokens;

    UPROPERTY()
    FString CurrentUsername;

    UPROPERTY()
    FString CurrentUserID;

    UPROPERTY()
    FString SessionID;

    FTimerHandle TokenRefreshTimer;

    void SetAuthState(EAuthState NewState);
    void HandleLoginResponse(const FString& Response, bool bSuccess);
    void HandleRegisterResponse(const FString& Response, bool bSuccess);
    void HandleRefreshResponse(const FString& Response, bool bSuccess);
    
    void StartTokenRefreshTimer();
    void StopTokenRefreshTimer();
    void OnTokenRefreshTimerFired();
    
    void SaveTokensToLocalStorage(const FAuthTokens& Tokens);
    bool LoadTokensFromLocalStorage(FAuthTokens& OutTokens);
    void ClearLocalStorage();
    
    void ParseJWTClaims(const FString& Token);
};
```

### 2.2 Auth UI Implementation

```cpp
// Source/MMORPGUI/Public/Auth/MMORPGLoginWidget.h
#pragma once

#include "CoreMinimal.h"
#include "Blueprint/UserWidget.h"
#include "MMORPGAuthTypes.h"
#include "MMORPGLoginWidget.generated.h"

UCLASS()
class MMORPGUI_API UMMORPGLoginWidget : public UUserWidget
{
    GENERATED_BODY()

public:
    // UI Bindings
    UPROPERTY(meta = (BindWidget))
    class UEditableTextBox* UsernameTextBox;

    UPROPERTY(meta = (BindWidget))
    class UEditableTextBox* PasswordTextBox;

    UPROPERTY(meta = (BindWidget))
    class UButton* LoginButton;

    UPROPERTY(meta = (BindWidget))
    class UButton* RegisterButton;

    UPROPERTY(meta = (BindWidget))
    class UTextBlock* ErrorText;

    UPROPERTY(meta = (BindWidget))
    class UThrobber* LoadingThrobber;

    // Events
    DECLARE_DYNAMIC_MULTICAST_DELEGATE(FOnLoginRequested);
    DECLARE_DYNAMIC_MULTICAST_DELEGATE(FOnRegisterRequested);

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|UI")
    FOnLoginRequested OnLoginRequested;

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|UI")
    FOnRegisterRequested OnRegisterRequested;

protected:
    virtual void NativeConstruct() override;
    virtual void NativeDestruct() override;

private:
    UFUNCTION()
    void OnLoginButtonClicked();

    UFUNCTION()
    void OnRegisterButtonClicked();

    UFUNCTION()
    void OnUsernameTextChanged(const FText& Text);

    UFUNCTION()
    void OnPasswordTextChanged(const FText& Text);

    UFUNCTION()
    void OnAuthStateChanged(EAuthState NewState);

    UFUNCTION()
    void OnLoginSuccess(const FAuthTokens& Tokens);

    UFUNCTION()
    void OnLoginFailure(const FMMORPGError& Error);

    void SetUIEnabled(bool bEnabled);
    void ShowError(const FString& ErrorMessage);
    void ClearError();
    void UpdateLoginButtonState();

    UPROPERTY()
    class UMMORPGAuthSubsystem* AuthSubsystem;

    FString CurrentUsername;
    FString CurrentPassword;
};
```

### 2.3 Character System Integration

```cpp
// Source/MMORPGUI/Public/Character/MMORPGCharacterSubsystem.h
#pragma once

#include "CoreMinimal.h"
#include "Subsystems/GameInstanceSubsystem.h"
#include "MMORPGCharacterTypes.h"
#include "MMORPGCharacterSubsystem.generated.h"

USTRUCT(BlueprintType)
struct FCharacterInfo
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadOnly)
    FString CharacterID;

    UPROPERTY(BlueprintReadOnly)
    FString Name;

    UPROPERTY(BlueprintReadOnly)
    int32 Level = 1;

    UPROPERTY(BlueprintReadOnly)
    FString Class;

    UPROPERTY(BlueprintReadOnly)
    FDateTime LastPlayed;
};

DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnCharacterListReceived, const TArray<FCharacterInfo>&, Characters);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnCharacterCreated, const FCharacterInfo&, Character);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnCharacterSelected, const FCharacterInfo&, Character);

UCLASS()
class MMORGCORE_API UMMORPGCharacterSubsystem : public UGameInstanceSubsystem
{
    GENERATED_BODY()

public:
    // Events
    UPROPERTY(BlueprintAssignable)
    FOnCharacterListReceived OnCharacterListReceived;

    UPROPERTY(BlueprintAssignable)
    FOnCharacterCreated OnCharacterCreated;

    UPROPERTY(BlueprintAssignable)
    FOnCharacterSelected OnCharacterSelected;

    // Character management
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character")
    void FetchCharacterList();

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character")
    void CreateCharacter(const FString& Name, const FString& Class);

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character")
    void SelectCharacter(const FString& CharacterID);

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character")
    void DeleteCharacter(const FString& CharacterID);

    // Character queries
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character")
    TArray<FCharacterInfo> GetCachedCharacterList() const { return CachedCharacters; }

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character")
    FCharacterInfo GetSelectedCharacter() const { return SelectedCharacter; }

    UFUNCTION(BlueprintCallable, Category = "MMORPG|Character")
    bool HasSelectedCharacter() const { return !SelectedCharacter.CharacterID.IsEmpty(); }

private:
    UPROPERTY()
    TArray<FCharacterInfo> CachedCharacters;

    UPROPERTY()
    FCharacterInfo SelectedCharacter;

    UPROPERTY()
    class UMMORPGAuthSubsystem* AuthSubsystem;

    void HandleCharacterListResponse(const FString& Response, bool bSuccess);
    void HandleCreateCharacterResponse(const FString& Response, bool bSuccess);
    void HandleSelectCharacterResponse(const FString& Response, bool bSuccess);
};
```

### 2.4 Token Storage and Security

```cpp
// Source/MMORPGCore/Private/Auth/MMORPGTokenStorage.cpp
#include "Auth/MMORPGTokenStorage.h"
#include "HAL/PlatformFilemanager.h"
#include "Misc/Paths.h"
#include "Misc/FileHelper.h"
#include "Dom/JsonObject.h"
#include "Serialization/JsonSerializer.h"

#if PLATFORM_WINDOWS
#include "Windows/AllowWindowsPlatformTypes.h"
#include <dpapi.h>
#include "Windows/HideWindowsPlatformTypes.h"
#endif

class FMMORPGTokenStorage
{
public:
    static bool SaveTokens(const FAuthTokens& Tokens)
    {
        // Create JSON object
        TSharedPtr<FJsonObject> JsonObject = MakeShareable(new FJsonObject);
        JsonObject->SetStringField("AccessToken", Tokens.AccessToken);
        JsonObject->SetStringField("RefreshToken", Tokens.RefreshToken);
        JsonObject->SetNumberField("ExpiresIn", Tokens.ExpiresIn);
        JsonObject->SetStringField("ExpiresAt", Tokens.ExpiresAt.ToString());
        
        // Serialize to string
        FString JsonString;
        TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&JsonString);
        FJsonSerializer::Serialize(JsonObject.ToSharedRef(), Writer);
        
        // Encrypt data
        FString EncryptedData = EncryptString(JsonString);
        
        // Save to file
        FString FilePath = GetTokenFilePath();
        return FFileHelper::SaveStringToFile(EncryptedData, *FilePath);
    }
    
    static bool LoadTokens(FAuthTokens& OutTokens)
    {
        FString FilePath = GetTokenFilePath();
        
        // Check if file exists
        if (!FPlatformFileManager::Get().GetPlatformFile().FileExists(*FilePath))
        {
            return false;
        }
        
        // Load encrypted data
        FString EncryptedData;
        if (!FFileHelper::LoadFileToString(EncryptedData, *FilePath))
        {
            return false;
        }
        
        // Decrypt data
        FString JsonString = DecryptString(EncryptedData);
        
        // Parse JSON
        TSharedPtr<FJsonObject> JsonObject;
        TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(JsonString);
        if (!FJsonSerializer::Deserialize(Reader, JsonObject))
        {
            return false;
        }
        
        // Extract tokens
        OutTokens.AccessToken = JsonObject->GetStringField("AccessToken");
        OutTokens.RefreshToken = JsonObject->GetStringField("RefreshToken");
        OutTokens.ExpiresIn = JsonObject->GetIntegerField("ExpiresIn");
        
        FString ExpiresAtStr = JsonObject->GetStringField("ExpiresAt");
        FDateTime::Parse(ExpiresAtStr, OutTokens.ExpiresAt);
        
        return true;
    }
    
    static void ClearTokens()
    {
        FString FilePath = GetTokenFilePath();
        FPlatformFileManager::Get().GetPlatformFile().DeleteFile(*FilePath);
    }
    
private:
    static FString GetTokenFilePath()
    {
        FString SaveDir = FPaths::ProjectSavedDir() / "Auth";
        FPlatformFileManager::Get().GetPlatformFile().CreateDirectory(*SaveDir);
        return SaveDir / "tokens.dat";
    }
    
    static FString EncryptString(const FString& PlainText)
    {
#if PLATFORM_WINDOWS
        // Use Windows DPAPI for encryption
        DATA_BLOB DataIn;
        DATA_BLOB DataOut;
        
        // Convert string to bytes
        TArray<uint8> PlainBytes;
        FTCHARToUTF8 Converter(*PlainText);
        PlainBytes.Append((uint8*)Converter.Get(), Converter.Length());
        
        DataIn.pbData = PlainBytes.GetData();
        DataIn.cbData = PlainBytes.Num();
        
        // Encrypt
        if (CryptProtectData(&DataIn, nullptr, nullptr, nullptr, nullptr, 0, &DataOut))
        {
            // Convert to base64
            FString Result = FBase64::Encode(DataOut.pbData, DataOut.cbData);
            LocalFree(DataOut.pbData);
            return Result;
        }
#endif
        // Fallback: Simple XOR encryption (not secure, for development only)
        return SimpleXOREncrypt(PlainText);
    }
    
    static FString DecryptString(const FString& EncryptedText)
    {
#if PLATFORM_WINDOWS
        // Decode from base64
        TArray<uint8> EncryptedBytes;
        FBase64::Decode(EncryptedText, EncryptedBytes);
        
        DATA_BLOB DataIn;
        DATA_BLOB DataOut;
        
        DataIn.pbData = EncryptedBytes.GetData();
        DataIn.cbData = EncryptedBytes.Num();
        
        // Decrypt
        if (CryptUnprotectData(&DataIn, nullptr, nullptr, nullptr, nullptr, 0, &DataOut))
        {
            // Convert to string
            FUTF8ToTCHAR Converter((char*)DataOut.pbData, DataOut.cbData);
            FString Result(Converter.Get());
            LocalFree(DataOut.pbData);
            return Result;
        }
#endif
        // Fallback
        return SimpleXORDecrypt(EncryptedText);
    }
    
    // Simple XOR for development (NOT SECURE)
    static FString SimpleXOREncrypt(const FString& Text)
    {
        const uint8 Key = 0x42;
        FString Result;
        for (TCHAR Char : Text)
        {
            Result += FString::Printf(TEXT("%02X"), (uint8)(Char ^ Key));
        }
        return Result;
    }
    
    static FString SimpleXORDecrypt(const FString& Text)
    {
        const uint8 Key = 0x42;
        FString Result;
        for (int32 i = 0; i < Text.Len(); i += 2)
        {
            FString Hex = Text.Mid(i, 2);
            uint8 Byte = (uint8)FCString::Strtoi(*Hex, nullptr, 16);
            Result += (TCHAR)(Byte ^ Key);
        }
        return Result;
    }
};
```

### 2.5 Auto Token Refresh

```cpp
// Source/MMORPGCore/Private/Auth/MMORPGAuthSubsystem.cpp
void UMMORPGAuthSubsystem::StartTokenRefreshTimer()
{
    StopTokenRefreshTimer();
    
    if (!IsTokenExpired() && IsAuthenticated())
    {
        // Calculate when to refresh (80% of token lifetime)
        float TimeUntilExpiry = GetTimeUntilTokenExpiry();
        float RefreshTime = TimeUntilExpiry * 0.8f;
        
        // Minimum 30 seconds
        RefreshTime = FMath::Max(30.0f, RefreshTime);
        
        UE_LOG(LogMMORPG, Log, TEXT("Starting token refresh timer for %f seconds"), RefreshTime);
        
        GetWorld()->GetTimerManager().SetTimer(
            TokenRefreshTimer,
            this,
            &UMMORPGAuthSubsystem::OnTokenRefreshTimerFired,
            RefreshTime,
            false
        );
    }
}

void UMMORPGAuthSubsystem::OnTokenRefreshTimerFired()
{
    UE_LOG(LogMMORPG, Log, TEXT("Token refresh timer fired"));
    
    if (IsAuthenticated() && !CurrentTokens.RefreshToken.IsEmpty())
    {
        RefreshToken();
    }
}

void UMMORPGAuthSubsystem::RefreshToken()
{
    if (CurrentState == EAuthState::RefreshingToken)
    {
        UE_LOG(LogMMORPG, Warning, TEXT("Already refreshing token"));
        return;
    }
    
    SetAuthState(EAuthState::RefreshingToken);
    
    // Create refresh request
    TSharedPtr<FJsonObject> RequestBody = MakeShareable(new FJsonObject);
    RequestBody->SetStringField("refresh_token", CurrentTokens.RefreshToken);
    
    FString BodyString;
    TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&BodyString);
    FJsonSerializer::Serialize(RequestBody.ToSharedRef(), Writer);
    
    // Send request
    UMMORPGHTTPClient* HTTPClient = GetGameInstance()->GetSubsystem<UMMORPGHTTPClient>();
    UMMORPGHTTPRequest* Request = HTTPClient->POST("/api/v1/auth/refresh", BodyString);
    
    Request->OnComplete.AddUObject(this, &UMMORPGAuthSubsystem::HandleRefreshResponse);
    Request->Send();
}

void UMMORPGAuthSubsystem::HandleRefreshResponse(const FString& Response, bool bSuccess)
{
    if (!bSuccess)
    {
        UE_LOG(LogMMORPG, Error, TEXT("Token refresh failed"));
        
        // If refresh fails, user needs to login again
        Logout();
        
        FMMORPGError Error;
        Error.Code = 401;
        Error.Message = "Session expired. Please login again.";
        Error.Severity = EMMORPGErrorSeverity::Warning;
        
        OnLoginFailure.Broadcast(Error);
        return;
    }
    
    // Parse response
    TSharedPtr<FJsonObject> JsonObject;
    TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(Response);
    if (!FJsonSerializer::Deserialize(Reader, JsonObject))
    {
        UE_LOG(LogMMORPG, Error, TEXT("Failed to parse refresh response"));
        SetAuthState(EAuthState::Error);
        return;
    }
    
    // Extract new tokens
    FAuthTokens NewTokens;
    JsonObject->TryGetStringField("access_token", NewTokens.AccessToken);
    JsonObject->TryGetStringField("refresh_token", NewTokens.RefreshToken);
    JsonObject->TryGetNumberField("expires_in", NewTokens.ExpiresIn);
    
    // Update stored tokens
    CurrentTokens = NewTokens;
    CurrentTokens.ExpiresAt = FDateTime::Now() + FTimespan::FromSeconds(NewTokens.ExpiresIn);
    
    // Parse JWT claims
    ParseJWTClaims(NewTokens.AccessToken);
    
    // Save to storage
    SaveTokensToLocalStorage(CurrentTokens);
    
    // Update HTTP client auth header
    UMMORPGHTTPClient* HTTPClient = GetGameInstance()->GetSubsystem<UMMORPGHTTPClient>();
    HTTPClient->SetAuthToken(CurrentTokens.AccessToken);
    
    // Restart refresh timer
    StartTokenRefreshTimer();
    
    SetAuthState(EAuthState::Authenticated);
    
    UE_LOG(LogMMORPG, Log, TEXT("Token refresh successful"));
}
```

---

## 🔒 3. Security Architecture

### 3.1 Rate Limiting Implementation

```go
// internal/adapters/auth/rate_limiter.go
package auth

import (
    "context"
    "fmt"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type RateLimiter struct {
    redis  *redis.Client
    config *RateLimitConfig
}

type RateLimitConfig struct {
    // Request limits
    RegisterLimit    int           // per hour per IP
    LoginLimit       int           // per hour per IP
    RefreshLimit     int           // per hour per user
    
    // Failed login tracking
    MaxFailedLogins  int           // before lockout
    LockoutDuration  time.Duration // account lockout time
    
    // IP-based limits
    MaxRequestsPerIP int           // per hour
    BanDuration      time.Duration // IP ban duration
}

func NewRateLimiter(redis *redis.Client, config *RateLimitConfig) *RateLimiter {
    return &RateLimiter{
        redis:  redis,
        config: config,
    }
}

func (rl *RateLimiter) Limit(operation string) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        
        // Check if IP is banned
        if banned, err := rl.isIPBanned(clientIP); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit check failed"})
            c.Abort()
            return
        } else if banned {
            c.JSON(http.StatusTooManyRequests, gin.H{"error": "IP address banned"})
            c.Abort()
            return
        }
        
        // Check operation-specific limit
        key := fmt.Sprintf("ratelimit:%s:%s", operation, clientIP)
        limit := rl.getLimit(operation)
        
        count, err := rl.incrementCounter(key, time.Hour)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit check failed"})
            c.Abort()
            return
        }
        
        if count > limit {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
                "retry_after": 3600,
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

func (rl *RateLimiter) TrackFailedLogin(ip, username string) error {
    // Track by IP
    ipKey := fmt.Sprintf("failed_login:ip:%s", ip)
    ipCount, err := rl.incrementCounter(ipKey, time.Hour)
    if err != nil {
        return err
    }
    
    // Ban IP if too many failures
    if ipCount > rl.config.MaxFailedLogins * 3 {
        return rl.banIP(ip)
    }
    
    // Track by username
    if username != "" {
        userKey := fmt.Sprintf("failed_login:user:%s", username)
        userCount, err := rl.incrementCounter(userKey, 30*time.Minute)
        if err != nil {
            return err
        }
        
        // Lock account if too many failures
        if userCount >= rl.config.MaxFailedLogins {
            return rl.lockAccount(username)
        }
    }
    
    return nil
}

func (rl *RateLimiter) IsAccountLocked(userID string) (bool, time.Time) {
    key := fmt.Sprintf("account_locked:%s", userID)
    
    result, err := rl.redis.Get(context.Background(), key).Result()
    if err == redis.Nil {
        return false, time.Time{}
    } else if err != nil {
        // Log error but don't block login
        return false, time.Time{}
    }
    
    until, err := time.Parse(time.RFC3339, result)
    if err != nil {
        return false, time.Time{}
    }
    
    if time.Now().After(until) {
        // Lock expired, clean up
        rl.redis.Del(context.Background(), key)
        return false, time.Time{}
    }
    
    return true, until
}

func (rl *RateLimiter) incrementCounter(key string, window time.Duration) (int, error) {
    pipe := rl.redis.Pipeline()
    
    incr := pipe.Incr(context.Background(), key)
    pipe.Expire(context.Background(), key, window)
    
    _, err := pipe.Exec(context.Background())
    if err != nil {
        return 0, err
    }
    
    return int(incr.Val()), nil
}

func (rl *RateLimiter) banIP(ip string) error {
    key := fmt.Sprintf("ip_banned:%s", ip)
    return rl.redis.Set(
        context.Background(),
        key,
        time.Now().Add(rl.config.BanDuration).Format(time.RFC3339),
        rl.config.BanDuration,
    ).Err()
}

func (rl *RateLimiter) lockAccount(username string) error {
    // Get user ID from username
    // This would normally query the database
    
    key := fmt.Sprintf("account_locked:%s", username)
    until := time.Now().Add(rl.config.LockoutDuration)
    
    return rl.redis.Set(
        context.Background(),
        key,
        until.Format(time.RFC3339),
        rl.config.LockoutDuration,
    ).Err()
}
```

### 3.2 Input Validation

```go
// internal/adapters/auth/validation.go
package auth

import (
    "regexp"
    "strings"
    "unicode"
)

var (
    usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)
    emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

func ValidateUsername(username string) error {
    if len(username) < 3 || len(username) > 50 {
        return ErrInvalidUsername
    }
    
    if !usernameRegex.MatchString(username) {
        return ErrInvalidUsername
    }
    
    // Check for reserved names
    reserved := []string{"admin", "moderator", "gm", "gamemaster", "system"}
    lower := strings.ToLower(username)
    for _, r := range reserved {
        if strings.Contains(lower, r) {
            return ErrUsernameReserved
        }
    }
    
    return nil
}

func ValidatePassword(password string) error {
    if len(password) < 8 || len(password) > 128 {
        return ErrPasswordTooWeak
    }
    
    var (
        hasUpper   bool
        hasLower   bool
        hasNumber  bool
        hasSpecial bool
    )
    
    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsDigit(char):
            hasNumber = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):
            hasSpecial = true
        }
    }
    
    // Require at least 3 of 4 character types
    strength := 0
    if hasUpper {
        strength++
    }
    if hasLower {
        strength++
    }
    if hasNumber {
        strength++
    }
    if hasSpecial {
        strength++
    }
    
    if strength < 3 {
        return ErrPasswordTooWeak
    }
    
    // Check for common passwords
    if isCommonPassword(password) {
        return ErrPasswordTooCommon
    }
    
    return nil
}

func ValidateEmail(email string) error {
    if len(email) > 255 {
        return ErrInvalidEmail
    }
    
    if !emailRegex.MatchString(email) {
        return ErrInvalidEmail
    }
    
    // Check for disposable email domains
    domain := strings.Split(email, "@")[1]
    if isDisposableEmailDomain(domain) {
        return ErrDisposableEmail
    }
    
    return nil
}
```

---

## 🔄 4. Event-Driven Architecture

### 4.1 NATS Integration

```go
// internal/adapters/nats/auth_events.go
package nats

import (
    "context"
    "encoding/json"
    
    "github.com/nats-io/nats.go"
)

type AuthEventPublisher struct {
    nc     *nats.Conn
    logger logger.Logger
}

func NewAuthEventPublisher(nc *nats.Conn, logger logger.Logger) *AuthEventPublisher {
    return &AuthEventPublisher{
        nc:     nc,
        logger: logger,
    }
}

func (p *AuthEventPublisher) PublishUserRegistered(ctx context.Context, event *UserRegisteredEvent) error {
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    subject := "auth.user.registered"
    if err := p.nc.Publish(subject, data); err != nil {
        p.logger.WithError(err).Error("Failed to publish user registered event")
        return err
    }
    
    p.logger.WithFields(logger.Fields{
        "user_id": event.UserID,
        "subject": subject,
    }).Debug("Published user registered event")
    
    return nil
}

func (p *AuthEventPublisher) PublishUserLoggedIn(ctx context.Context, event *UserLoggedInEvent) error {
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    subject := "auth.user.logged_in"
    if err := p.nc.Publish(subject, data); err != nil {
        p.logger.WithError(err).Error("Failed to publish user logged in event")
        return err
    }
    
    // Track for analytics
    metricsSubject := "metrics.auth.login"
    p.nc.Publish(metricsSubject, data)
    
    return nil
}

// Event types
type UserRegisteredEvent struct {
    UserID    string    `json:"user_id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Timestamp time.Time `json:"timestamp"`
}

type UserLoggedInEvent struct {
    UserID    string    `json:"user_id"`
    SessionID string    `json:"session_id"`
    IP        string    `json:"ip"`
    UserAgent string    `json:"user_agent"`
    Timestamp time.Time `json:"timestamp"`
}

type UserLoggedOutEvent struct {
    UserID    string    `json:"user_id"`
    SessionID string    `json:"session_id"`
    Reason    string    `json:"reason"` // "manual", "timeout", "forced"
    Timestamp time.Time `json:"timestamp"`
}
```

---

## 🎯 Summary

Phase 1's authentication architecture provides:

1. **Backend Architecture**
   - Clean hexagonal architecture with domain-driven design
   - JWT-based authentication with refresh tokens
   - Secure password handling with bcrypt
   - Redis session caching with PostgreSQL persistence
   - Comprehensive rate limiting and security

2. **Frontend Architecture**
   - Centralized auth subsystem in UE5
   - Secure token storage with platform-specific encryption
   - Automatic token refresh mechanism
   - Clean UI widget architecture
   - Full Blueprint support

3. **Security Features**
   - Multi-layered rate limiting
   - Account lockout protection
   - Input validation and sanitization
   - Secure token storage
   - Event-driven audit logging

4. **Integration**
   - NATS event publishing for microservices
   - Clean API design with proper error handling
   - Extensible character system
   - Ready for Phase 2 WebSocket authentication

This architecture ensures secure, scalable authentication that can grow with the game while maintaining clean separation of concerns and excellent developer experience.