package character

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mmorpg-template/backend/internal/domain/auth"
	"github.com/mmorpg-template/backend/pkg/logger"
)

// JWTConfig holds JWT validation configuration
type JWTConfig struct {
	AccessSecret string
	Issuer       string
}

// JWTMiddleware provides JWT validation middleware
type JWTMiddleware struct {
	config *JWTConfig
	logger logger.Logger
}

// NewJWTMiddleware creates a new JWT middleware instance
func NewJWTMiddleware(config *JWTConfig, logger logger.Logger) *JWTMiddleware {
	return &JWTMiddleware{
		config: config,
		logger: logger,
	}
}

// Validate returns a gin middleware function that validates JWT tokens
func (m *JWTMiddleware) Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.respondWithAuthError(c, "missing authorization header")
			return
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.respondWithAuthError(c, "invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims, err := m.validateToken(c.Request.Context(), tokenString)
		if err != nil {
			m.handleTokenError(c, err)
			return
		}

		// Set user information in context
		c.Set("userID", claims.UserID)
		c.Set("sessionID", claims.SessionID)
		c.Set("email", claims.Email)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Set("deviceID", claims.DeviceID)

		// Add claims to request context for downstream use
		ctx := context.WithValue(c.Request.Context(), "claims", claims)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// validateToken validates a JWT token and returns the claims
func (m *JWTMiddleware) validateToken(ctx context.Context, tokenString string) (*auth.Claims, error) {
	claims := &auth.Claims{}
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.config.AccessSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, auth.ErrInvalidToken
	}

	// Validate issuer
	if claims.Issuer != m.config.Issuer {
		return nil, auth.ErrInvalidToken
	}

	// Check if claims are valid (user ID and session ID must be present)
	if !claims.IsValid() {
		return nil, auth.ErrInvalidToken
	}

	return claims, nil
}

// handleTokenError handles JWT validation errors and returns appropriate responses
func (m *JWTMiddleware) handleTokenError(c *gin.Context, err error) {
	switch err {
	case jwt.ErrTokenExpired:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": ErrorDetail{
				Code:    ErrorCodeUnauthorized,
				Message: "token expired",
				Details: map[string]interface{}{
					"expired": true,
				},
			},
			"timestamp": time.Now().Format(time.RFC3339),
		})
	case jwt.ErrTokenMalformed:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": ErrorDetail{
				Code:    ErrorCodeUnauthorized,
				Message: "malformed token",
			},
			"timestamp": time.Now().Format(time.RFC3339),
		})
	case jwt.ErrTokenSignatureInvalid:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": ErrorDetail{
				Code:    ErrorCodeUnauthorized,
				Message: "invalid token signature",
			},
			"timestamp": time.Now().Format(time.RFC3339),
		})
	case auth.ErrInvalidToken:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": ErrorDetail{
				Code:    ErrorCodeUnauthorized,
				Message: "invalid token",
			},
			"timestamp": time.Now().Format(time.RFC3339),
		})
	default:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": ErrorDetail{
				Code:    ErrorCodeUnauthorized,
				Message: "invalid token",
			},
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
	c.Abort()
}

// respondWithAuthError sends an authentication error response
func (m *JWTMiddleware) respondWithAuthError(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"error": ErrorDetail{
			Code:    ErrorCodeUnauthorized,
			Message: message,
		},
		"timestamp": time.Now().Format(time.RFC3339),
		"request_id": c.GetString("request_id"),
	})
	c.Abort()
}

// OptionalAuth returns a middleware that validates JWT if present but doesn't require it
func (m *JWTMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			c.Next()
			return
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format, continue without authentication
			c.Next()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims, err := m.validateToken(c.Request.Context(), tokenString)
		if err != nil {
			// Log the error but continue without authentication
			m.logger.WithError(err).Debug("Optional auth token validation failed")
			c.Next()
			return
		}

		// Set user information in context
		c.Set("userID", claims.UserID)
		c.Set("sessionID", claims.SessionID)
		c.Set("email", claims.Email)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Set("deviceID", claims.DeviceID)
		c.Set("authenticated", true)

		// Add claims to request context
		ctx := context.WithValue(c.Request.Context(), "claims", claims)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequireRole returns a middleware that requires a specific role
func (m *JWTMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated
		if !c.GetBool("authenticated") {
			m.respondWithAuthError(c, "authentication required")
			return
		}

		// Get roles from context
		roles, exists := c.Get("roles")
		if !exists {
			m.respondWithAuthError(c, "no roles found")
			return
		}

		// Check if user has the required role
		userRoles, ok := roles.([]string)
		if !ok {
			m.respondWithAuthError(c, "invalid roles format")
			return
		}

		hasRole := false
		for _, r := range userRoles {
			if r == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": ErrorDetail{
					Code:    "INSUFFICIENT_PERMISSIONS",
					Message: "insufficient permissions",
					Details: map[string]interface{}{
						"required_role": role,
					},
				},
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserIDFromContext extracts the user ID from the request context
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	
	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

// GetClaimsFromContext extracts the full claims from the request context
func GetClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value("claims").(*auth.Claims)
	return claims, ok
}