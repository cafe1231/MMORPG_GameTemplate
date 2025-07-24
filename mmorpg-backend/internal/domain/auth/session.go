package auth

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an authenticated user session
type Session struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string
	DeviceID   string
	IPAddress  string
	UserAgent  string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	LastActive time.Time
}

// NewSession creates a new session
func NewSession(userID uuid.UUID, tokenHash, deviceID, ipAddress, userAgent string, expiresAt time.Time) *Session {
	now := time.Now()
	return &Session{
		ID:         uuid.New(),
		UserID:     userID,
		TokenHash:  tokenHash,
		DeviceID:   deviceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		ExpiresAt:  expiresAt,
		CreatedAt:  now,
		LastActive: now,
	}
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// UpdateActivity updates the last active timestamp
func (s *Session) UpdateActivity() {
	s.LastActive = time.Now()
}

// IsStale checks if the session has been inactive for too long
func (s *Session) IsStale(staleDuration time.Duration) bool {
	return time.Since(s.LastActive) > staleDuration
}