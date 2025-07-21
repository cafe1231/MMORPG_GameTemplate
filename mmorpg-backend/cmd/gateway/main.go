package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mmorpg-template/backend/internal/config"
	"github.com/mmorpg-template/backend/pkg/logger"
	"github.com/mmorpg-template/backend/pkg/metrics"
)

func main() {
	ctx := context.Background()

	log := logger.New("gateway")
	log.Info("Starting MMORPG Gateway Service...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	metricsServer := metrics.NewServer(cfg.Metrics.Port)
	go func() {
		if err := metricsServer.Start(); err != nil {
			log.Errorf("Metrics server error: %v", err)
		}
	}()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      setupRoutes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Infof("Gateway server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down gateway service...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Gateway service stopped")
}

func setupRoutes() http.Handler {
	mux := http.NewServeMux()
	
	// Enable CORS for development
	handler := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, User-Agent")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			next(w, r)
		}
	}
	
	mux.HandleFunc("/health", handler(func(w http.ResponseWriter, r *http.Request) {
		health := map[string]interface{}{
			"status": "healthy",
			"timestamp": time.Now().Unix(),
			"service": "gateway",
			"version": "0.1.0",
		}
		json.NewEncoder(w).Encode(health)
	}))

	mux.HandleFunc("/", handler(func(w http.ResponseWriter, r *http.Request) {
		info := map[string]interface{}{
			"service": "mmorpg-gateway",
			"version": "0.1.0",
			"endpoints": []string{
				"/health",
				"/api/v1/test",
				"/api/v1/echo",
				"/api/v1/auth/login",
				"/api/v1/auth/register",
				"/api/v1/characters",
			},
		}
		json.NewEncoder(w).Encode(info)
	}))
	
	// Test endpoint for client connection
	mux.HandleFunc("/api/v1/test", handler(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"success": true,
			"message": "Connection test successful",
			"timestamp": time.Now().Unix(),
			"method": r.Method,
			"client": r.Header.Get("User-Agent"),
		}
		json.NewEncoder(w).Encode(response)
	}))
	
	// Echo endpoint for testing
	mux.HandleFunc("/api/v1/echo", handler(func(w http.ResponseWriter, r *http.Request) {
		var data map[string]interface{}
		if r.Method == "POST" || r.Method == "PUT" {
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Invalid JSON",
				})
				return
			}
		}
		
		response := map[string]interface{}{
			"echo": data,
			"headers": r.Header,
			"method": r.Method,
		}
		json.NewEncoder(w).Encode(response)
	}))

	return mux
}