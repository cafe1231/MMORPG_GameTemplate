package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/auth"
	portsAuth "github.com/mmorpg-template/backend/internal/ports/auth"
)

// JWTGenerator implements the TokenGenerator interface using JWT
type JWTGenerator struct {
	accessSecret  string
	refreshSecret string
	issuer        string
}

// NewJWTGenerator creates a new JWT generator
func NewJWTGenerator(accessSecret, refreshSecret, issuer string) portsAuth.TokenGenerator {
	return &JWTGenerator{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		issuer:        issuer,
	}
}

// GenerateTokenPair generates an access and refresh token pair
func (j *JWTGenerator) GenerateTokenPair(ctx context.Context, user *auth.User, sessionID, deviceID string) (*auth.TokenPair, error) {
	// Generate access token
	accessClaims := &auth.Claims{
		UserID:    user.ID.String(),
		SessionID: sessionID,
		Email:     user.Email,
		Username:  user.Username,
		Roles:     user.Roles,
		DeviceID:  deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(auth.AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
			Subject:   user.ID.String(),
			ID:        uuid.New().String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(j.accessSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &auth.RefreshClaims{
		UserID:    user.ID.String(),
		SessionID: sessionID,
		DeviceID:  deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(auth.RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
			Subject:   user.ID.String(),
			ID:        uuid.New().String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(j.refreshSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return auth.NewTokenPair(accessTokenString, refreshTokenString), nil
}

// ValidateAccessToken validates an access token and returns the claims
func (j *JWTGenerator) ValidateAccessToken(ctx context.Context, tokenString string) (*auth.Claims, error) {
	claims := &auth.Claims{}
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.accessSecret), nil
	})

	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, auth.ErrTokenExpired
		}
		if err == jwt.ErrTokenMalformed {
			return nil, auth.ErrTokenMalformed
		}
		if err == jwt.ErrTokenSignatureInvalid {
			return nil, auth.ErrTokenSignatureInvalid
		}
		return nil, auth.ErrInvalidToken
	}

	if !token.Valid {
		return nil, auth.ErrInvalidToken
	}

	if !claims.IsValid() {
		return nil, auth.ErrInvalidToken
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns the claims
func (j *JWTGenerator) ValidateRefreshToken(ctx context.Context, tokenString string) (*auth.RefreshClaims, error) {
	claims := &auth.RefreshClaims{}
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.refreshSecret), nil
	})

	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, auth.ErrTokenExpired
		}
		if err == jwt.ErrTokenMalformed {
			return nil, auth.ErrTokenMalformed
		}
		if err == jwt.ErrTokenSignatureInvalid {
			return nil, auth.ErrTokenSignatureInvalid
		}
		return nil, auth.ErrRefreshTokenInvalid
	}

	if !token.Valid {
		return nil, auth.ErrRefreshTokenInvalid
	}

	return claims, nil
}

// GeneratePasswordResetToken generates a password reset token
func (j *JWTGenerator) GeneratePasswordResetToken(ctx context.Context, userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    j.issuer,
		Subject:   userID,
		ID:        uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.accessSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign password reset token: %w", err)
	}

	return tokenString, nil
}

// ValidatePasswordResetToken validates a password reset token
func (j *JWTGenerator) ValidatePasswordResetToken(ctx context.Context, tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.accessSecret), nil
	})

	if err != nil || !token.Valid {
		return "", auth.ErrInvalidToken
	}

	return claims.Subject, nil
}

// HashToken creates a hash of a token for storage
func (j *JWTGenerator) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}