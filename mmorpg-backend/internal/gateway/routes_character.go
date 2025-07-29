package gateway

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mmorpg-template/backend/pkg/logger"
)

// CharacterRoutes defines all character service routes
type CharacterRoutes struct {
	characterServiceURL string
	authMiddleware      func(http.HandlerFunc) http.HandlerFunc
	rateLimiter        *RateLimiter
	logger             logger.Logger
}

// NewCharacterRoutes creates a new character routes handler
func NewCharacterRoutes(characterServiceURL string, authMiddleware func(http.HandlerFunc) http.HandlerFunc, rateLimiter *RateLimiter, logger logger.Logger) *CharacterRoutes {
	return &CharacterRoutes{
		characterServiceURL: characterServiceURL,
		authMiddleware:      authMiddleware,
		rateLimiter:        rateLimiter,
		logger:             logger,
	}
}

// RegisterRoutes registers all character routes with the mux
func (cr *CharacterRoutes) RegisterRoutes(mux *http.ServeMux, corsHandler func(http.HandlerFunc) http.HandlerFunc) {
	// Public endpoints (no auth required)
	mux.HandleFunc("/api/v1/characters/check-name", corsHandler(
		cr.rateLimiter.Limit("check-name", 30, 1*Minute)(
			cr.createProxy("/api/v1/characters/check-name"),
		),
	))

	// Character management endpoints (auth required)
	mux.HandleFunc("/api/v1/characters", corsHandler(
		cr.authMiddleware(
			cr.rateLimiter.Limit("characters", 60, 1*Minute)(
				cr.createProxy("/api/v1/characters"),
			),
		),
	))

	mux.HandleFunc("/api/v1/characters/deleted", corsHandler(
		cr.authMiddleware(
			cr.rateLimiter.Limit("characters", 60, 1*Minute)(
				cr.createProxy("/api/v1/characters/deleted"),
			),
		),
	))

	// Character-specific endpoints with ID parameter
	// These need special handling for path parameters
	characterPaths := []struct {
		path      string
		rateLimit RateLimitConfig
	}{
		{"/api/v1/characters/", RateLimitConfig{requests: 60, window: 1 * Minute}},
		{"/api/v1/characters/", RateLimitConfig{requests: 5, window: 1 * Hour}}, // for DELETE
		{"/api/v1/characters/", RateLimitConfig{requests: 10, window: 1 * Hour}}, // for appearance
	}

	// Generic character endpoint handler
	mux.HandleFunc("/api/v1/characters/", corsHandler(
		cr.authMiddleware(
			cr.handleCharacterEndpoint(),
		),
	))
}

// handleCharacterEndpoint handles all /api/v1/characters/{id}/* endpoints
func (cr *CharacterRoutes) handleCharacterEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		// Extract the character ID and endpoint
		parts := strings.Split(strings.TrimPrefix(path, "/api/v1/characters/"), "/")
		if len(parts) < 1 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		characterID := parts[0]
		endpoint := ""
		if len(parts) > 1 {
			endpoint = parts[1]
		}

		// Route to appropriate handler based on endpoint
		switch endpoint {
		case "":
			// GET /api/v1/characters/{id}
			// DELETE /api/v1/characters/{id}
			if method == "DELETE" {
				cr.rateLimiter.Limit("delete-character", 3, 24*Hour)(
					cr.createProxy(path),
				)(w, r)
			} else {
				cr.rateLimiter.Limit("get-character", 60, 1*Minute)(
					cr.createProxy(path),
				)(w, r)
			}

		case "restore":
			// POST /api/v1/characters/{id}/restore
			cr.rateLimiter.Limit("restore-character", 5, 1*Hour)(
				cr.createProxy(path),
			)(w, r)

		case "select":
			// POST /api/v1/characters/{id}/select
			cr.rateLimiter.Limit("select-character", 10, 1*Minute)(
				cr.createProxy(path),
			)(w, r)

		case "appearance":
			// GET/PUT /api/v1/characters/{id}/appearance
			cr.rateLimiter.Limit("appearance", 10, 1*Hour)(
				cr.createProxy(path),
			)(w, r)

		case "stats":
			// GET /api/v1/characters/{id}/stats
			if len(parts) > 2 && parts[2] == "allocate" {
				// POST /api/v1/characters/{id}/stats/allocate
				cr.rateLimiter.Limit("allocate-stats", 100, 1*Hour)(
					cr.createProxy(path),
				)(w, r)
			} else {
				cr.rateLimiter.Limit("get-stats", 60, 1*Minute)(
					cr.createProxy(path),
				)(w, r)
			}

		case "position":
			// GET/PUT /api/v1/characters/{id}/position
			cr.rateLimiter.Limit("position", 100, 1*Second)(
				cr.createProxy(path),
			)(w, r)

		case "permanent":
			// DELETE /api/v1/characters/{id}/permanent
			if method == "DELETE" {
				cr.rateLimiter.Limit("permanent-delete", 1, 24*Hour)(
					cr.createProxy(path),
				)(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}

		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}
}

// createProxy creates a proxy handler for the character service
func (cr *CharacterRoutes) createProxy(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetURL := cr.characterServiceURL + path
		
		// Add query parameters if any
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		// Create proxy request
		proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
		if err != nil {
			cr.logger.WithError(err).Error("Failed to create proxy request")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Copy headers
		for name, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(name, value)
			}
		}

		// Add user ID from auth context
		if userID := r.Context().Value("user_id"); userID != nil {
			proxyReq.Header.Set("X-User-ID", userID.(string))
		}

		// Forward client IP
		proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)
		proxyReq.Header.Set("X-Real-IP", strings.Split(r.RemoteAddr, ":")[0])

		// Make the request
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(proxyReq)
		if err != nil {
			cr.logger.WithError(err).Error("Character service request failed")
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		// Set status code
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		io.Copy(w, resp.Body)
	}
}

// RateLimitConfig defines rate limit configuration
type RateLimitConfig struct {
	requests int
	window   time.Duration
}

// Time constants for rate limiting
const (
	Second = time.Second
	Minute = time.Minute
	Hour   = time.Hour
)

// RateLimiter interface (to be implemented)
type RateLimiter struct {
	// Implementation details
}

func (rl *RateLimiter) Limit(name string, requests int, window time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	// Rate limiting middleware implementation
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement rate limiting logic
			// For now, just pass through
			next(w, r)
		}
	}
}