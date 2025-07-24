package auth

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID               uuid.UUID
	Email            string
	Username         string
	PasswordHash     string
	EmailVerified    bool
	AccountStatus    AccountStatus
	Roles            []string
	MaxCharacters    int
	CharacterCount   int
	IsPremium        bool
	PremiumExpiresAt *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// AccountStatus represents the status of a user account
type AccountStatus int

const (
	AccountStatusActive AccountStatus = iota + 1
	AccountStatusSuspended
	AccountStatusBanned
	AccountStatusPendingVerification
	AccountStatusDeleted
)

// NewUser creates a new user with default values
func NewUser(email, username, passwordHash string) *User {
	now := time.Now()
	return &User{
		ID:               uuid.New(),
		Email:            email,
		Username:         username,
		PasswordHash:     passwordHash,
		EmailVerified:    false,
		AccountStatus:    AccountStatusPendingVerification,
		Roles:            []string{"player"},
		MaxCharacters:    5,
		CharacterCount:   0,
		IsPremium:        false,
		PremiumExpiresAt: nil,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// CanCreateCharacter checks if the user can create more characters
func (u *User) CanCreateCharacter() bool {
	return u.CharacterCount < u.MaxCharacters
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.AccountStatus == AccountStatusActive
}

// HasRole checks if the user has a specific role
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// UpdatePremiumStatus updates the user's premium status
func (u *User) UpdatePremiumStatus(isPremium bool, expiresAt *time.Time) {
	u.IsPremium = isPremium
	u.PremiumExpiresAt = expiresAt
	if isPremium && u.MaxCharacters < 10 {
		u.MaxCharacters = 10 // Premium users get more character slots
	} else if !isPremium && u.MaxCharacters > 5 {
		u.MaxCharacters = 5
	}
	u.UpdatedAt = time.Now()
}