package auth

import (
	"github.com/mmorpg-template/backend/internal/domain/auth"
	portsAuth "github.com/mmorpg-template/backend/internal/ports/auth"
	"golang.org/x/crypto/bcrypt"
)

// BcryptPasswordHasher implements PasswordHasher using bcrypt
type BcryptPasswordHasher struct {
	cost int
}

// NewBcryptPasswordHasher creates a new bcrypt password hasher
func NewBcryptPasswordHasher(cost int) portsAuth.PasswordHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &BcryptPasswordHasher{cost: cost}
}

// HashPassword creates a hash from a password
func (h *BcryptPasswordHasher) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ComparePassword compares a password with its hash
func (h *BcryptPasswordHasher) ComparePassword(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return auth.ErrPasswordMismatch
		}
		return err
	}
	return nil
}