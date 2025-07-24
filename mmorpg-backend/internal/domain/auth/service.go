package auth

import (
	"context"
	"regexp"
	"strings"
	"unicode"
)

// Service defines the authentication business logic
type Service interface {
	// User registration and authentication
	Register(ctx context.Context, req *RegisterRequest) (*User, error)
	Login(ctx context.Context, email, password, deviceID, ipAddress, userAgent string) (*TokenPair, *User, error)
	Logout(ctx context.Context, sessionID string) error
	LogoutAllDevices(ctx context.Context, userID string) error
	
	// Token management
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	ValidateToken(ctx context.Context, token string) (*Claims, error)
	RevokeToken(ctx context.Context, token string) error
	
	// User management
	GetUser(ctx context.Context, userID string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	
	// Session management
	GetUserSessions(ctx context.Context, userID string) ([]*Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email        string
	Password     string
	Username     string
	AcceptTerms  bool
	ReferralCode string
}

// Validate validates the registration request
func (r *RegisterRequest) Validate() error {
	// Validate email
	if !isValidEmail(r.Email) {
		return ErrInvalidEmail
	}
	
	// Validate username
	if !isValidUsername(r.Username) {
		return ErrInvalidUsername
	}
	
	// Validate password
	if !isStrongPassword(r.Password) {
		return ErrPasswordTooWeak
	}
	
	// Check terms acceptance
	if !r.AcceptTerms {
		return ErrTermsNotAccepted
	}
	
	return nil
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