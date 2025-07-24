package auth

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/auth"
	portsAuth "github.com/mmorpg-template/backend/internal/ports/auth"
	"github.com/mmorpg-template/backend/pkg/logger"
)

// AuthServiceImpl implements the auth service business logic
type AuthServiceImpl struct {
	userRepo       portsAuth.UserRepository
	sessionRepo    portsAuth.SessionRepository
	tokenGenerator portsAuth.TokenGenerator
	passwordHasher portsAuth.PasswordHasher
	tokenCache     portsAuth.TokenCache
	config         *Config
	logger         logger.Logger
}

// Config holds auth service configuration
type Config struct {
	MaxSessionsPerUser   int
	LoginRateLimit       int
	LoginRateLimitWindow time.Duration
	SessionDuration      time.Duration
	MaxLoginAttempts     int
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo portsAuth.UserRepository,
	sessionRepo portsAuth.SessionRepository,
	tokenGenerator portsAuth.TokenGenerator,
	passwordHasher portsAuth.PasswordHasher,
	tokenCache portsAuth.TokenCache,
	config *Config,
	logger logger.Logger,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		tokenGenerator: tokenGenerator,
		passwordHasher: passwordHasher,
		tokenCache:     tokenCache,
		config:         config,
		logger:         logger,
	}
}

// Register creates a new user account
func (s *AuthServiceImpl) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.User, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if email already exists
	emailExists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check email existence")
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if emailExists {
		return nil, auth.ErrEmailAlreadyTaken
	}

	// Check if username already exists
	usernameExists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check username existence")
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if usernameExists {
		return nil, auth.ErrUsernameAlreadyTaken
	}

	// Hash password
	passwordHash, err := s.passwordHasher.HashPassword(req.Password)
	if err != nil {
		s.logger.WithError(err).Error("Failed to hash password")
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := auth.NewUser(req.Email, req.Username, passwordHash)
	
	// Set initial account status based on configuration
	// In production, you might want email verification
	user.AccountStatus = auth.AccountStatusActive
	user.EmailVerified = true // For now, auto-verify

	// Save user
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.WithError(err).Error("Failed to create user")
		return nil, err
	}

	s.logger.WithField("userID", user.ID).Info("User registered successfully")
	return user, nil
}

// Login authenticates a user and returns tokens
func (s *AuthServiceImpl) Login(ctx context.Context, email, password, deviceID, ipAddress, userAgent string) (*auth.TokenPair, *auth.User, error) {
	// Check rate limiting
	identifier := fmt.Sprintf("login:%s", ipAddress)
	attempts, err := s.tokenCache.IncrementLoginAttempts(ctx, identifier, s.config.LoginRateLimitWindow)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check login attempts")
	}
	
	if attempts > s.config.MaxLoginAttempts {
		return nil, nil, auth.ErrTooManyAttempts
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == auth.ErrUserNotFound {
			return nil, nil, auth.ErrInvalidCredentials
		}
		s.logger.WithError(err).Error("Failed to get user by email")
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check password
	if err := s.passwordHasher.ComparePassword(user.PasswordHash, password); err != nil {
		return nil, nil, auth.ErrInvalidCredentials
	}

	// Check account status
	switch user.AccountStatus {
	case auth.AccountStatusActive:
		// Continue with login
	case auth.AccountStatusSuspended:
		return nil, nil, auth.ErrAccountSuspended
	case auth.AccountStatusBanned:
		return nil, nil, auth.ErrAccountBanned
	case auth.AccountStatusPendingVerification:
		return nil, nil, auth.ErrEmailNotVerified
	default:
		return nil, nil, auth.ErrAccountNotActive
	}

	// Check session limit
	sessionCount, err := s.sessionRepo.CountByUserID(ctx, user.ID.String())
	if err != nil {
		s.logger.WithError(err).Error("Failed to count user sessions")
	}
	
	if sessionCount >= s.config.MaxSessionsPerUser {
		// Optional: Delete oldest session
		s.logger.WithField("userID", user.ID).Warn("Max sessions reached")
		return nil, nil, auth.ErrTooManySessions
	}

	// Generate tokens
	sessionID := uuid.New().String()
	tokenPair, err := s.tokenGenerator.GenerateTokenPair(ctx, user, sessionID, deviceID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate token pair")
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	tokenHash := s.tokenGenerator.HashToken(tokenPair.RefreshToken)
	session := auth.NewSession(
		user.ID,
		tokenHash,
		deviceID,
		ipAddress,
		userAgent,
		time.Now().Add(auth.RefreshTokenDuration),
	)
	session.ID = uuid.MustParse(sessionID)

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logger.WithError(err).Error("Failed to create session")
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Cache session for faster access
	// Note: This is a type assertion to a concrete type from adapters layer
	// In a pure hexagonal architecture, we'd use an interface method instead

	// Clear login attempts on successful login
	_ = s.tokenCache.DeleteLoginAttempts(ctx, identifier)

	// Update user last login
	user.UpdatedAt = time.Now()
	_ = s.userRepo.Update(ctx, user)

	s.logger.WithFields(map[string]interface{}{
		"userID":    user.ID,
		"sessionID": sessionID,
		"deviceID":  deviceID,
	}).Info("User logged in successfully")

	return tokenPair, user, nil
}

// Logout invalidates a session
func (s *AuthServiceImpl) Logout(ctx context.Context, sessionID string) error {
	// Delete session
	if err := s.sessionRepo.Delete(ctx, sessionID); err != nil {
		if err == auth.ErrSessionNotFound {
			return nil // Already logged out
		}
		s.logger.WithError(err).Error("Failed to delete session")
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Delete cached session
	_ = s.tokenCache.DeleteSession(ctx, sessionID)

	s.logger.WithField("sessionID", sessionID).Info("User logged out successfully")
	return nil
}

// LogoutAllDevices logs out all sessions for a user
func (s *AuthServiceImpl) LogoutAllDevices(ctx context.Context, userID string) error {
	// Get all sessions for cleanup
	sessions, err := s.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user sessions")
	}

	// Delete all sessions from database
	if err := s.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		s.logger.WithError(err).Error("Failed to delete user sessions")
		return fmt.Errorf("failed to delete sessions: %w", err)
	}

	// Clean up cached sessions
	for _, session := range sessions {
		_ = s.tokenCache.DeleteSession(ctx, session.ID.String())
	}

	s.logger.WithField("userID", userID).Info("All devices logged out successfully")
	return nil
}

// RefreshToken generates a new token pair from a refresh token
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, refreshToken, deviceID, ipAddress, userAgent string) (*auth.TokenPair, error) {
	// Validate refresh token
	claims, err := s.tokenGenerator.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	// Get session
	tokenHash := s.tokenGenerator.HashToken(refreshToken)
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, auth.ErrSessionNotFound
	}

	// Verify session belongs to the right user
	if session.UserID.String() != claims.UserID {
		return nil, auth.ErrSessionInvalid
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is still active
	if !user.IsActive() {
		return nil, auth.ErrAccountNotActive
	}

	// Generate new token pair
	newTokenPair, err := s.tokenGenerator.GenerateTokenPair(ctx, user, session.ID.String(), deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update session with new token hash
	session.TokenHash = s.tokenGenerator.HashToken(newTokenPair.RefreshToken)
	session.IPAddress = ipAddress
	session.UserAgent = userAgent
	session.UpdateActivity()

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"userID":    user.ID,
		"sessionID": session.ID,
	}).Debug("Token refreshed successfully")

	return newTokenPair, nil
}

// ValidateToken validates an access token
func (s *AuthServiceImpl) ValidateToken(ctx context.Context, token string) (*auth.Claims, error) {
	// Check if token is blacklisted
	tokenHash := s.tokenGenerator.HashToken(token)
	blacklisted, err := s.tokenCache.IsBlacklisted(ctx, tokenHash)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check token blacklist")
	}
	if blacklisted {
		return nil, auth.ErrInvalidToken
	}

	// Validate token
	claims, err := s.tokenGenerator.ValidateAccessToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Update session activity
	go func() {
		ctx := context.Background()
		_ = s.sessionRepo.UpdateLastActive(ctx, claims.SessionID, time.Now())
	}()

	return claims, nil
}

// GetUser retrieves a user by ID
func (s *AuthServiceImpl) GetUser(ctx context.Context, userID string) (*auth.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// ChangePassword changes a user's password
func (s *AuthServiceImpl) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if err := s.passwordHasher.ComparePassword(user.PasswordHash, currentPassword); err != nil {
		return auth.ErrPasswordMismatch
	}

	// Validate new password
	if !isStrongPassword(newPassword) {
		return auth.ErrPasswordTooWeak
	}

	// Hash new password
	newPasswordHash, err := s.passwordHasher.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user
	user.PasswordHash = newPasswordHash
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Optionally: Invalidate all sessions except current
	// This is a security measure to log out potential attackers

	s.logger.WithField("userID", userID).Info("Password changed successfully")
	return nil
}

// RequestPasswordReset initiates a password reset
func (s *AuthServiceImpl) RequestPasswordReset(ctx context.Context, email string) error {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists or not
		s.logger.WithField("email", email).Debug("Password reset requested for non-existent email")
		return nil
	}

	// Generate reset token
	resetToken, err := s.tokenGenerator.GeneratePasswordResetToken(ctx, user.ID.String())
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Store reset token in cache
	if err := s.tokenCache.SetPasswordResetToken(ctx, resetToken, user.ID.String(), 1*time.Hour); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// TODO: Send email with reset token
	// This would integrate with an email service

	s.logger.WithField("userID", user.ID).Info("Password reset requested")
	return nil
}

// ResetPassword completes a password reset
func (s *AuthServiceImpl) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Validate reset token
	userID, err := s.tokenGenerator.ValidatePasswordResetToken(ctx, token)
	if err != nil {
		return auth.ErrInvalidToken
	}

	// Check if token is in cache
	cachedUserID, err := s.tokenCache.GetPasswordResetToken(ctx, token)
	if err != nil || cachedUserID != userID {
		return auth.ErrInvalidToken
	}

	// Validate new password
	if !isStrongPassword(newPassword) {
		return auth.ErrPasswordTooWeak
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Hash new password
	newPasswordHash, err := s.passwordHasher.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user
	user.PasswordHash = newPasswordHash
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Delete reset token
	_ = s.tokenCache.DeletePasswordResetToken(ctx, token)

	// Invalidate all sessions for security
	_ = s.LogoutAllDevices(ctx, userID)

	s.logger.WithField("userID", userID).Info("Password reset successfully")
	return nil
}

// GetUserSessions retrieves all active sessions for a user
func (s *AuthServiceImpl) GetUserSessions(ctx context.Context, userID string) ([]*auth.Session, error) {
	return s.sessionRepo.GetByUserID(ctx, userID)
}

// RevokeSession revokes a specific session
func (s *AuthServiceImpl) RevokeSession(ctx context.Context, sessionID string) error {
	return s.Logout(ctx, sessionID)
}

// Helper functions for validation
var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(strings.ToLower(email))
}

func isValidUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

func isStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
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
	
	// Require at least 3 of the 4 character types
	count := 0
	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasNumber {
		count++
	}
	if hasSpecial {
		count++
	}
	
	return count >= 3
}