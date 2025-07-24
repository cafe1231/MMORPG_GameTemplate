package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // seconds until access token expires
}

// Claims represents the JWT claims
type Claims struct {
	UserID    string   `json:"uid"`
	SessionID string   `json:"sid"`
	Email     string   `json:"email"`
	Username  string   `json:"username"`
	Roles     []string `json:"roles"`
	DeviceID  string   `json:"did,omitempty"`
	jwt.RegisteredClaims
}

// Token durations
const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour // 7 days
)

// RefreshClaims represents the claims for a refresh token
type RefreshClaims struct {
	UserID    string `json:"uid"`
	SessionID string `json:"sid"`
	DeviceID  string `json:"did,omitempty"`
	jwt.RegisteredClaims
}

// NewTokenPair creates a new token pair
func NewTokenPair(accessToken, refreshToken string) *TokenPair {
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(AccessTokenDuration.Seconds()),
	}
}

// IsValid checks if the claims are valid
func (c *Claims) IsValid() bool {
	return c.UserID != "" && c.SessionID != ""
}

// HasRole checks if the user has a specific role
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}