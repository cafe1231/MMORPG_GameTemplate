package auth

import "errors"

// Domain-specific errors for authentication
var (
	// User errors
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrAccountNotActive      = errors.New("account is not active")
	ErrAccountSuspended      = errors.New("account is suspended")
	ErrAccountBanned         = errors.New("account is banned")
	ErrEmailNotVerified      = errors.New("email not verified")
	ErrUsernameAlreadyTaken  = errors.New("username already taken")
	ErrEmailAlreadyTaken     = errors.New("email already taken")
	
	// Session errors
	ErrSessionNotFound       = errors.New("session not found")
	ErrSessionExpired        = errors.New("session expired")
	ErrSessionInvalid        = errors.New("session invalid")
	ErrTooManySessions       = errors.New("too many active sessions")
	
	// Token errors
	ErrInvalidToken          = errors.New("invalid token")
	ErrTokenExpired          = errors.New("token expired")
	ErrTokenMalformed        = errors.New("token malformed")
	ErrTokenSignatureInvalid = errors.New("token signature invalid")
	ErrRefreshTokenInvalid   = errors.New("refresh token invalid")
	
	// Password errors
	ErrPasswordTooWeak       = errors.New("password too weak")
	ErrPasswordMismatch      = errors.New("password mismatch")
	
	// Rate limiting errors
	ErrTooManyAttempts       = errors.New("too many login attempts")
	
	// Validation errors
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrInvalidUsername       = errors.New("invalid username format")
	ErrTermsNotAccepted      = errors.New("terms of service not accepted")
)

// IsAuthError checks if an error is an authentication error
func IsAuthError(err error) bool {
	switch err {
	case ErrUserNotFound, ErrInvalidCredentials, ErrAccountNotActive,
		ErrAccountSuspended, ErrAccountBanned, ErrEmailNotVerified,
		ErrSessionExpired, ErrSessionInvalid, ErrInvalidToken,
		ErrTokenExpired, ErrTokenMalformed, ErrTokenSignatureInvalid:
		return true
	default:
		return false
	}
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	switch err {
	case ErrInvalidEmail, ErrInvalidUsername, ErrPasswordTooWeak,
		ErrTermsNotAccepted, ErrUsernameAlreadyTaken, ErrEmailAlreadyTaken:
		return true
	default:
		return false
	}
}