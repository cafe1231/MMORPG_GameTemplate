package character

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mmorpg-template/backend/internal/domain/auth"
	"github.com/mmorpg-template/backend/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTMiddleware_Validate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		setupAuth      func() string
		expectedStatus int
		expectedUserID string
		expectedError  string
	}{
		{
			name: "valid token",
			setupAuth: func() string {
				claims := &auth.Claims{
					UserID:    "test-user-123",
					SessionID: "session-456",
					Email:     "test@example.com",
					Username:  "testuser",
					Roles:     []string{"player"},
					DeviceID:  "device-789",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
						NotBefore: jwt.NewNumericDate(time.Now()),
						Issuer:    "mmorpg-auth",
						Subject:   "test-user-123",
					},
				}
				
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("test-secret"))
				return "Bearer " + tokenString
			},
			expectedStatus: http.StatusOK,
			expectedUserID: "test-user-123",
		},
		{
			name: "missing authorization header",
			setupAuth: func() string {
				return ""
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "missing authorization header",
		},
		{
			name: "invalid authorization format",
			setupAuth: func() string {
				return "InvalidFormat token"
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid authorization header format",
		},
		{
			name: "expired token",
			setupAuth: func() string {
				claims := &auth.Claims{
					UserID:    "test-user-123",
					SessionID: "session-456",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
						Issuer:    "mmorpg-auth",
					},
				}
				
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("test-secret"))
				return "Bearer " + tokenString
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "token expired",
		},
		{
			name: "wrong issuer",
			setupAuth: func() string {
				claims := &auth.Claims{
					UserID:    "test-user-123",
					SessionID: "session-456",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
						Issuer:    "wrong-issuer",
					},
				}
				
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("test-secret"))
				return "Bearer " + tokenString
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid token issuer",
		},
		{
			name: "invalid signature",
			setupAuth: func() string {
				claims := &auth.Claims{
					UserID:    "test-user-123",
					SessionID: "session-456",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
						Issuer:    "mmorpg-auth",
					},
				}
				
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("wrong-secret"))
				return "Bearer " + tokenString
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid token signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			config := &JWTConfig{
				AccessSecret: "test-secret",
				Issuer:       "mmorpg-auth",
			}
			middleware := NewJWTMiddleware(config, logger.NewNoop())
			
			// Create test router
			router := gin.New()
			router.Use(middleware.Validate())
			router.GET("/test", func(c *gin.Context) {
				userID, exists := c.Get("userID")
				if exists {
					c.JSON(http.StatusOK, gin.H{"userID": userID})
				} else {
					c.JSON(http.StatusOK, gin.H{"message": "ok"})
				}
			})
			
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if authHeader := tt.setupAuth(); authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}
			
			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedUserID != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedUserID, response["userID"])
			}
			
			if tt.expectedError != "" {
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				require.NoError(t, err)
				errorDetail := errorResponse["error"].(map[string]interface{})
				assert.Contains(t, errorDetail["message"], tt.expectedError)
			}
		})
	}
}

func TestJWTMiddleware_OptionalAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Setup
	config := &JWTConfig{
		AccessSecret: "test-secret",
		Issuer:       "mmorpg-auth",
	}
	middleware := NewJWTMiddleware(config, logger.NewNoop())
	
	// Create test router
	router := gin.New()
	router.Use(middleware.OptionalAuth())
	router.GET("/test", func(c *gin.Context) {
		authenticated := c.GetBool("authenticated")
		userID, _ := c.Get("userID")
		c.JSON(http.StatusOK, gin.H{
			"authenticated": authenticated,
			"userID":        userID,
		})
	})
	
	t.Run("with valid token", func(t *testing.T) {
		// Create valid token
		claims := &auth.Claims{
			UserID:    "test-user-123",
			SessionID: "session-456",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
				Issuer:    "mmorpg-auth",
			},
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte("test-secret"))
		
		// Create request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		
		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["authenticated"].(bool))
		assert.Equal(t, "test-user-123", response["userID"])
	})
	
	t.Run("without token", func(t *testing.T) {
		// Create request without auth header
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		
		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response["authenticated"].(bool))
		assert.Nil(t, response["userID"])
	})
	
	t.Run("with invalid token", func(t *testing.T) {
		// Create request with invalid token
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		
		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert - should still be OK but not authenticated
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response["authenticated"].(bool))
		assert.Nil(t, response["userID"])
	})
}

func TestJWTMiddleware_RequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Setup
	config := &JWTConfig{
		AccessSecret: "test-secret",
		Issuer:       "mmorpg-auth",
	}
	middleware := NewJWTMiddleware(config, logger.NewNoop())
	
	// Create test router
	router := gin.New()
	router.Use(middleware.OptionalAuth())
	router.GET("/admin", middleware.RequireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})
	
	t.Run("with admin role", func(t *testing.T) {
		// Create token with admin role
		claims := &auth.Claims{
			UserID:    "admin-user",
			SessionID: "session-456",
			Roles:     []string{"player", "admin"},
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
				Issuer:    "mmorpg-auth",
			},
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte("test-secret"))
		
		// Create request
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		
		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("without admin role", func(t *testing.T) {
		// Create token without admin role
		claims := &auth.Claims{
			UserID:    "regular-user",
			SessionID: "session-456",
			Roles:     []string{"player"},
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
				Issuer:    "mmorpg-auth",
			},
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte("test-secret"))
		
		// Create request
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		
		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	
	t.Run("without authentication", func(t *testing.T) {
		// Create request without auth
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		
		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}