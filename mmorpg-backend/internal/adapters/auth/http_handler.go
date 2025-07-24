package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mmorpg-template/backend/internal/domain/auth"
	portsAuth "github.com/mmorpg-template/backend/internal/ports/auth"
	"github.com/mmorpg-template/backend/pkg/logger"
	"github.com/mmorpg-template/backend/pkg/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// HTTPHandler handles HTTP requests for authentication
type HTTPHandler struct {
	authService portsAuth.AuthService
	logger      logger.Logger
}

// NewHTTPHandler creates a new HTTP handler for auth
func NewHTTPHandler(authService portsAuth.AuthService, logger logger.Logger) *HTTPHandler {
	return &HTTPHandler{
		authService: authService,
		logger:      logger,
	}
}

// Register handles user registration
func (h *HTTPHandler) Register(c *gin.Context) {
	var req proto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, proto.ErrorCode_ERROR_CODE_INVALID_REQUEST, "Invalid request format")
		return
	}

	// Convert proto request to domain request
	domainReq := &auth.RegisterRequest{
		Email:        req.Email,
		Password:     req.Password,
		Username:     req.Username,
		AcceptTerms:  req.AcceptTerms,
		ReferralCode: req.GetReferralCode(),
	}

	// Call service
	user, err := h.authService.Register(c.Request.Context(), domainReq)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	// Build response
	resp := &proto.RegisterResponse{
		Success: true,
		UserId:  user.ID.String(),
	}

	c.JSON(http.StatusCreated, resp)
}

// Login handles user login
func (h *HTTPHandler) Login(c *gin.Context) {
	var req proto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, proto.ErrorCode_ERROR_CODE_INVALID_REQUEST, "Invalid request format")
		return
	}

	// Get device info
	deviceID := req.DeviceId
	if deviceID == "" {
		deviceID = c.GetHeader("X-Device-ID")
	}
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Call service
	tokenPair, user, err := h.authService.Login(
		c.Request.Context(),
		req.Email,
		req.Password,
		deviceID,
		ipAddress,
		userAgent,
	)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	// Build user info
	userInfo := &proto.UserInfo{
		UserId:         user.ID.String(),
		Email:          user.Email,
		Username:       user.Username,
		EmailVerified:  user.EmailVerified,
		AccountStatus:  h.mapAccountStatus(user.AccountStatus),
		Roles:          user.Roles,
		MaxCharacters:  int32(user.MaxCharacters),
		CharacterCount: int32(user.CharacterCount),
		IsPremium:      user.IsPremium,
	}

	if user.CreatedAt.Unix() > 0 {
		userInfo.CreatedAt = h.timeToProto(user.CreatedAt)
	}

	// Build response
	resp := &proto.LoginResponse{
		Success:      true,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		SessionId:    "", // Session ID is embedded in the token
		ExpiresIn:    int32(tokenPair.ExpiresIn),
		UserInfo:     userInfo,
	}

	c.JSON(http.StatusOK, resp)
}

// Logout handles user logout
func (h *HTTPHandler) Logout(c *gin.Context) {
	var req proto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, proto.ErrorCode_ERROR_CODE_INVALID_REQUEST, "Invalid request format")
		return
	}

	// Get session ID from token if not provided
	sessionID := req.SessionId
	if sessionID == "" {
		claims, _ := h.getClaimsFromContext(c)
		if claims != nil {
			sessionID = claims.SessionID
		}
	}

	if sessionID == "" {
		h.respondWithError(c, http.StatusBadRequest, proto.ErrorCode_ERROR_CODE_INVALID_REQUEST, "Session ID required")
		return
	}

	// Handle logout all devices
	if req.LogoutAllDevices {
		claims, ok := h.getClaimsFromContext(c)
		if !ok {
			h.respondWithError(c, http.StatusUnauthorized, proto.ErrorCode_ERROR_CODE_UNAUTHORIZED, "Unauthorized")
			return
		}
		err := h.authService.LogoutAllDevices(c.Request.Context(), claims.UserID)
		if err != nil {
			h.handleAuthError(c, err)
			return
		}
	} else {
		// Logout single session
		err := h.authService.Logout(c.Request.Context(), sessionID)
		if err != nil {
			h.handleAuthError(c, err)
			return
		}
	}

	resp := &proto.LogoutResponse{
		Success: true,
		Message: "Logged out successfully",
	}

	c.JSON(http.StatusOK, resp)
}

// RefreshToken handles token refresh
func (h *HTTPHandler) RefreshToken(c *gin.Context) {
	var req proto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, proto.ErrorCode_ERROR_CODE_INVALID_REQUEST, "Invalid request format")
		return
	}

	// Get device info
	deviceID := c.GetHeader("X-Device-ID")
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Call service
	tokenPair, err := h.authService.RefreshToken(
		c.Request.Context(),
		req.RefreshToken,
		deviceID,
		ipAddress,
		userAgent,
	)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	// Build response
	resp := &proto.RefreshTokenResponse{
		Success:      true,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int32(tokenPair.ExpiresIn),
	}

	c.JSON(http.StatusOK, resp)
}

// VerifyToken validates a token (used by other services)
func (h *HTTPHandler) VerifyToken(c *gin.Context) {
	// Get token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		h.respondWithError(c, http.StatusUnauthorized, proto.ErrorCode_ERROR_CODE_UNAUTHORIZED, "Missing authorization header")
		return
	}

	// Extract token
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		h.respondWithError(c, http.StatusUnauthorized, proto.ErrorCode_ERROR_CODE_UNAUTHORIZED, "Invalid authorization format")
		return
	}

	// Validate token
	claims, err := h.authService.ValidateToken(c.Request.Context(), token)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	// Return claims as JSON
	c.JSON(http.StatusOK, gin.H{
		"valid":     true,
		"user_id":   claims.UserID,
		"session_id": claims.SessionID,
		"email":     claims.Email,
		"username":  claims.Username,
		"roles":     claims.Roles,
	})
}

// Middleware provides JWT authentication middleware
func (h *HTTPHandler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			h.respondWithError(c, http.StatusUnauthorized, proto.ErrorCode_ERROR_CODE_UNAUTHORIZED, "Missing authorization header")
			c.Abort()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			h.respondWithError(c, http.StatusUnauthorized, proto.ErrorCode_ERROR_CODE_UNAUTHORIZED, "Invalid authorization format")
			c.Abort()
			return
		}

		// Validate token
		claims, err := h.authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			h.handleAuthError(c, err)
			c.Abort()
			return
		}

		// Store claims in context
		c.Set("claims", claims)
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// Helper methods

func (h *HTTPHandler) handleAuthError(c *gin.Context, err error) {
	switch err {
	case auth.ErrUserNotFound, auth.ErrInvalidCredentials:
		h.respondWithError(c, http.StatusUnauthorized, proto.ErrorCode_ERROR_CODE_INVALID_CREDENTIALS, "Invalid credentials")
	case auth.ErrUserAlreadyExists, auth.ErrEmailAlreadyTaken:
		h.respondWithError(c, http.StatusConflict, proto.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "Email already registered")
	case auth.ErrUsernameAlreadyTaken:
		h.respondWithError(c, http.StatusConflict, proto.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "Username already taken")
	case auth.ErrAccountSuspended:
		h.respondWithError(c, http.StatusForbidden, proto.ErrorCode_ERROR_CODE_FORBIDDEN, "Account suspended")
	case auth.ErrAccountBanned:
		h.respondWithError(c, http.StatusForbidden, proto.ErrorCode_ERROR_CODE_FORBIDDEN, "Account banned")
	case auth.ErrEmailNotVerified:
		h.respondWithError(c, http.StatusForbidden, proto.ErrorCode_ERROR_CODE_FORBIDDEN, "Email not verified")
	case auth.ErrTokenExpired:
		h.respondWithError(c, http.StatusUnauthorized, proto.ErrorCode_ERROR_CODE_UNAUTHORIZED, "Token expired")
	case auth.ErrInvalidToken, auth.ErrTokenMalformed, auth.ErrTokenSignatureInvalid:
		h.respondWithError(c, http.StatusUnauthorized, proto.ErrorCode_ERROR_CODE_UNAUTHORIZED, "Invalid token")
	case auth.ErrTooManyAttempts:
		h.respondWithError(c, http.StatusTooManyRequests, proto.ErrorCode_ERROR_CODE_RATE_LIMITED, "Too many login attempts")
	case auth.ErrPasswordTooWeak:
		h.respondWithError(c, http.StatusBadRequest, proto.ErrorCode_ERROR_CODE_INVALID_REQUEST, "Password too weak")
	case auth.ErrInvalidEmail:
		h.respondWithError(c, http.StatusBadRequest, proto.ErrorCode_ERROR_CODE_INVALID_REQUEST, "Invalid email format")
	case auth.ErrInvalidUsername:
		h.respondWithError(c, http.StatusBadRequest, proto.ErrorCode_ERROR_CODE_INVALID_REQUEST, "Invalid username format")
	case auth.ErrTermsNotAccepted:
		h.respondWithError(c, http.StatusBadRequest, proto.ErrorCode_ERROR_CODE_INVALID_REQUEST, "Terms of service must be accepted")
	default:
		h.logger.WithError(err).Error("Unhandled auth error")
		h.respondWithError(c, http.StatusInternalServerError, proto.ErrorCode_ERROR_CODE_SERVER_ERROR, "Internal server error")
	}
}

func (h *HTTPHandler) respondWithError(c *gin.Context, statusCode int, errorCode proto.ErrorCode, message string) {
	// For auth endpoints, use the specific response types
	path := c.Request.URL.Path
	
	if strings.Contains(path, "/login") {
		resp := &proto.LoginResponse{
			Success:      false,
			ErrorMessage: message,
			ErrorCode:    errorCode,
		}
		c.JSON(statusCode, resp)
	} else if strings.Contains(path, "/register") {
		resp := &proto.RegisterResponse{
			Success:      false,
			ErrorMessage: message,
			ErrorCode:    errorCode,
		}
		c.JSON(statusCode, resp)
	} else if strings.Contains(path, "/refresh") {
		resp := &proto.RefreshTokenResponse{
			Success:      false,
			ErrorMessage: message,
			ErrorCode:    errorCode,
		}
		c.JSON(statusCode, resp)
	} else {
		// Generic error response
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    errorCode,
				"message": message,
			},
		})
	}
}

func (h *HTTPHandler) getClaimsFromContext(c *gin.Context) (*auth.Claims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}
	authClaims, ok := claims.(*auth.Claims)
	return authClaims, ok
}

func (h *HTTPHandler) mapAccountStatus(status auth.AccountStatus) proto.AccountStatus {
	switch status {
	case auth.AccountStatusActive:
		return proto.AccountStatus_ACCOUNT_STATUS_ACTIVE
	case auth.AccountStatusSuspended:
		return proto.AccountStatus_ACCOUNT_STATUS_SUSPENDED
	case auth.AccountStatusBanned:
		return proto.AccountStatus_ACCOUNT_STATUS_BANNED
	case auth.AccountStatusPendingVerification:
		return proto.AccountStatus_ACCOUNT_STATUS_PENDING_VERIFICATION
	case auth.AccountStatusDeleted:
		return proto.AccountStatus_ACCOUNT_STATUS_DELETED
	default:
		return proto.AccountStatus_ACCOUNT_STATUS_UNSPECIFIED
	}
}

func (h *HTTPHandler) timeToProto(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}