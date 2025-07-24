package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/auth"
	"github.com/mmorpg-template/backend/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *auth.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*auth.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*auth.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *mockUserRepository) Update(ctx context.Context, user *auth.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepository) IncrementCharacterCount(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockUserRepository) DecrementCharacterCount(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type mockSessionRepository struct {
	mock.Mock
}

func (m *mockSessionRepository) Create(ctx context.Context, session *auth.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockSessionRepository) GetByID(ctx context.Context, id string) (*auth.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Session), args.Error(1)
}

func (m *mockSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*auth.Session, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Session), args.Error(1)
}

func (m *mockSessionRepository) GetByUserID(ctx context.Context, userID string) ([]*auth.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*auth.Session), args.Error(1)
}

func (m *mockSessionRepository) Update(ctx context.Context, session *auth.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockSessionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockSessionRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockSessionRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *mockSessionRepository) UpdateLastActive(ctx context.Context, sessionID string, lastActive time.Time) error {
	args := m.Called(ctx, sessionID, lastActive)
	return args.Error(0)
}

type mockTokenGenerator struct {
	mock.Mock
}

func (m *mockTokenGenerator) GenerateTokenPair(ctx context.Context, user *auth.User, sessionID, deviceID string) (*auth.TokenPair, error) {
	args := m.Called(ctx, user, sessionID, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenPair), args.Error(1)
}

func (m *mockTokenGenerator) ValidateAccessToken(ctx context.Context, token string) (*auth.Claims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

func (m *mockTokenGenerator) ValidateRefreshToken(ctx context.Context, token string) (*auth.RefreshClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.RefreshClaims), args.Error(1)
}

func (m *mockTokenGenerator) GeneratePasswordResetToken(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *mockTokenGenerator) ValidatePasswordResetToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *mockTokenGenerator) HashToken(token string) string {
	args := m.Called(token)
	return args.String(0)
}

type mockPasswordHasher struct {
	mock.Mock
}

func (m *mockPasswordHasher) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *mockPasswordHasher) ComparePassword(hash, password string) error {
	args := m.Called(hash, password)
	return args.Error(0)
}

type mockTokenCache struct {
	mock.Mock
}

func (m *mockTokenCache) SetBlacklisted(ctx context.Context, tokenHash string, expiration time.Duration) error {
	args := m.Called(ctx, tokenHash, expiration)
	return args.Error(0)
}

func (m *mockTokenCache) IsBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	args := m.Called(ctx, tokenHash)
	return args.Bool(0), args.Error(1)
}

func (m *mockTokenCache) SetSession(ctx context.Context, sessionID string, sessionData []byte, expiration time.Duration) error {
	args := m.Called(ctx, sessionID, sessionData, expiration)
	return args.Error(0)
}

func (m *mockTokenCache) GetSession(ctx context.Context, sessionID string) ([]byte, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockTokenCache) DeleteSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *mockTokenCache) SetLoginAttempts(ctx context.Context, identifier string, attempts int, expiration time.Duration) error {
	args := m.Called(ctx, identifier, attempts, expiration)
	return args.Error(0)
}

func (m *mockTokenCache) GetLoginAttempts(ctx context.Context, identifier string) (int, error) {
	args := m.Called(ctx, identifier)
	return args.Int(0), args.Error(1)
}

func (m *mockTokenCache) IncrementLoginAttempts(ctx context.Context, identifier string, expiration time.Duration) (int, error) {
	args := m.Called(ctx, identifier, expiration)
	return args.Int(0), args.Error(1)
}

func (m *mockTokenCache) DeleteLoginAttempts(ctx context.Context, identifier string) error {
	args := m.Called(ctx, identifier)
	return args.Error(0)
}

func (m *mockTokenCache) SetPasswordResetToken(ctx context.Context, token string, userID string, expiration time.Duration) error {
	args := m.Called(ctx, token, userID, expiration)
	return args.Error(0)
}

func (m *mockTokenCache) GetPasswordResetToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *mockTokenCache) DeletePasswordResetToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// Tests

func TestRegister(t *testing.T) {
	ctx := context.Background()
	
	userRepo := new(mockUserRepository)
	sessionRepo := new(mockSessionRepository)
	tokenGen := new(mockTokenGenerator)
	passHasher := new(mockPasswordHasher)
	tokenCache := new(mockTokenCache)
	
	config := &auth.Config{
		MaxSessionsPerUser:   10,
		LoginRateLimit:       10,
		LoginRateLimitWindow: 15 * time.Minute,
		SessionDuration:      7 * 24 * time.Hour,
		MaxLoginAttempts:     5,
	}
	
	logger := logger.NewNoop()
	
	service := auth.NewAuthService(userRepo, sessionRepo, tokenGen, passHasher, tokenCache, config, logger)
	
	t.Run("successful registration", func(t *testing.T) {
		req := &auth.RegisterRequest{
			Email:       "test@example.com",
			Password:    "StrongPass123!",
			Username:    "testuser",
			AcceptTerms: true,
		}
		
		userRepo.On("ExistsByEmail", ctx, req.Email).Return(false, nil)
		userRepo.On("ExistsByUsername", ctx, req.Username).Return(false, nil)
		passHasher.On("HashPassword", req.Password).Return("hashed_password", nil)
		userRepo.On("Create", ctx, mock.AnythingOfType("*auth.User")).Return(nil)
		
		user, err := service.Register(ctx, req)
		
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, req.Email, user.Email)
		assert.Equal(t, req.Username, user.Username)
		assert.Equal(t, "hashed_password", user.PasswordHash)
		
		userRepo.AssertExpectations(t)
		passHasher.AssertExpectations(t)
	})
	
	t.Run("email already exists", func(t *testing.T) {
		req := &auth.RegisterRequest{
			Email:       "existing@example.com",
			Password:    "StrongPass123!",
			Username:    "newuser",
			AcceptTerms: true,
		}
		
		userRepo.On("ExistsByEmail", ctx, req.Email).Return(true, nil)
		
		user, err := service.Register(ctx, req)
		
		assert.Error(t, err)
		assert.Equal(t, auth.ErrEmailAlreadyTaken, err)
		assert.Nil(t, user)
	})
	
	t.Run("weak password", func(t *testing.T) {
		req := &auth.RegisterRequest{
			Email:       "test@example.com",
			Password:    "weak",
			Username:    "testuser",
			AcceptTerms: true,
		}
		
		user, err := service.Register(ctx, req)
		
		assert.Error(t, err)
		assert.Equal(t, auth.ErrPasswordTooWeak, err)
		assert.Nil(t, user)
	})
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	
	userRepo := new(mockUserRepository)
	sessionRepo := new(mockSessionRepository)
	tokenGen := new(mockTokenGenerator)
	passHasher := new(mockPasswordHasher)
	tokenCache := new(mockTokenCache)
	
	config := &auth.Config{
		MaxSessionsPerUser:   10,
		LoginRateLimit:       10,
		LoginRateLimitWindow: 15 * time.Minute,
		SessionDuration:      7 * 24 * time.Hour,
		MaxLoginAttempts:     5,
	}
	
	logger := logger.NewNoop()
	
	service := auth.NewAuthService(userRepo, sessionRepo, tokenGen, passHasher, tokenCache, config, logger)
	
	t.Run("successful login", func(t *testing.T) {
		email := "test@example.com"
		password := "StrongPass123!"
		deviceID := "device123"
		ipAddress := "192.168.1.1"
		userAgent := "TestAgent"
		
		user := &auth.User{
			ID:            uuid.New(),
			Email:         email,
			Username:      "testuser",
			PasswordHash:  "hashed_password",
			AccountStatus: auth.AccountStatusActive,
			Roles:         []string{"player"},
		}
		
		tokenPair := &auth.TokenPair{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			ExpiresIn:    900,
		}
		
		tokenCache.On("IncrementLoginAttempts", ctx, "login:"+ipAddress, config.LoginRateLimitWindow).Return(1, nil)
		userRepo.On("GetByEmail", ctx, email).Return(user, nil)
		passHasher.On("ComparePassword", user.PasswordHash, password).Return(nil)
		sessionRepo.On("CountByUserID", ctx, user.ID.String()).Return(0, nil)
		tokenGen.On("GenerateTokenPair", ctx, user, mock.AnythingOfType("string"), deviceID).Return(tokenPair, nil)
		tokenGen.On("HashToken", tokenPair.RefreshToken).Return("hashed_refresh_token")
		sessionRepo.On("Create", ctx, mock.AnythingOfType("*auth.Session")).Return(nil)
		tokenCache.On("DeleteLoginAttempts", ctx, "login:"+ipAddress).Return(nil)
		userRepo.On("Update", ctx, user).Return(nil)
		
		resultTokenPair, resultUser, err := service.Login(ctx, email, password, deviceID, ipAddress, userAgent)
		
		assert.NoError(t, err)
		assert.NotNil(t, resultTokenPair)
		assert.NotNil(t, resultUser)
		assert.Equal(t, tokenPair, resultTokenPair)
		assert.Equal(t, user, resultUser)
		
		userRepo.AssertExpectations(t)
		passHasher.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
		tokenGen.AssertExpectations(t)
		tokenCache.AssertExpectations(t)
	})
	
	t.Run("invalid credentials", func(t *testing.T) {
		email := "test@example.com"
		password := "wrongpassword"
		deviceID := "device123"
		ipAddress := "192.168.1.1"
		userAgent := "TestAgent"
		
		user := &auth.User{
			ID:            uuid.New(),
			Email:         email,
			Username:      "testuser",
			PasswordHash:  "hashed_password",
			AccountStatus: auth.AccountStatusActive,
		}
		
		tokenCache.On("IncrementLoginAttempts", ctx, "login:"+ipAddress, config.LoginRateLimitWindow).Return(1, nil)
		userRepo.On("GetByEmail", ctx, email).Return(user, nil)
		passHasher.On("ComparePassword", user.PasswordHash, password).Return(auth.ErrPasswordMismatch)
		
		resultTokenPair, resultUser, err := service.Login(ctx, email, password, deviceID, ipAddress, userAgent)
		
		assert.Error(t, err)
		assert.Equal(t, auth.ErrInvalidCredentials, err)
		assert.Nil(t, resultTokenPair)
		assert.Nil(t, resultUser)
	})
}