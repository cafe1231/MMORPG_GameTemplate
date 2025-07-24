package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Gateway metrics
	ActiveConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mmorpg_active_connections",
			Help: "Number of active WebSocket connections",
		},
		[]string{"region", "server"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mmorpg_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "status"},
	)

	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mmorpg_requests_total",
			Help: "Total number of requests",
		},
		[]string{"service", "method", "status"},
	)

	// Game metrics
	PlayersOnline = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mmorpg_players_online",
			Help: "Number of players currently online",
		},
		[]string{"world", "region"},
	)

	MessagesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mmorpg_messages_processed_total",
			Help: "Total number of game messages processed",
		},
		[]string{"type", "direction"},
	)

	MessageSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mmorpg_message_size_bytes",
			Help:    "Size of game messages in bytes",
			Buckets: []float64{64, 128, 256, 512, 1024, 2048, 4096, 8192},
		},
		[]string{"type", "direction"},
	)

	// Database metrics
	DatabaseQueries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mmorpg_database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"service", "operation", "status"},
	)

	DatabaseDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mmorpg_database_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"service", "operation"},
	)

	// Cache metrics
	CacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mmorpg_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"service", "cache_type"},
	)

	CacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mmorpg_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"service", "cache_type"},
	)

	// Authentication metrics
	LoginAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mmorpg_login_attempts_total",
			Help: "Total number of login attempts",
		},
		[]string{"status", "reason"},
	)

	TokensGenerated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mmorpg_tokens_generated_total",
			Help: "Total number of JWT tokens generated",
		},
		[]string{"type"},
	)

	// Performance metrics
	TickDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mmorpg_tick_duration_seconds",
			Help:    "Game tick processing duration",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
		[]string{"world"},
	)

	EntityCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mmorpg_entity_count",
			Help: "Number of entities in the game world",
		},
		[]string{"world", "type"},
	)
)

func init() {
	// Register all metrics
	prometheus.MustRegister(
		ActiveConnections,
		RequestDuration,
		RequestsTotal,
		PlayersOnline,
		MessagesProcessed,
		MessageSize,
		DatabaseQueries,
		DatabaseDuration,
		CacheHits,
		CacheMisses,
		LoginAttempts,
		TokensGenerated,
		TickDuration,
		EntityCount,
	)
}

// Init initializes metrics (for explicit initialization)
func Init() {
	// Metrics are already registered in init()
	// This function exists for explicit initialization if needed
}

type Server struct {
	port string
}

func NewServer(port string) *Server {
	if port == "" {
		port = "9090"
	}
	return &Server{port: port}
}

func (s *Server) Start() error {
	http.Handle("/metrics", promhttp.Handler())
	
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := fmt.Sprintf(":%s", s.port)
	return http.ListenAndServe(addr, nil)
}

// Helper functions for common metric operations

func RecordRequestDuration(service, method, status string, duration float64) {
	RequestDuration.WithLabelValues(service, method, status).Observe(duration)
	RequestsTotal.WithLabelValues(service, method, status).Inc()
}

func RecordDatabaseQuery(service, operation, status string, duration float64) {
	DatabaseQueries.WithLabelValues(service, operation, status).Inc()
	DatabaseDuration.WithLabelValues(service, operation).Observe(duration)
}

func RecordCacheHit(service, cacheType string) {
	CacheHits.WithLabelValues(service, cacheType).Inc()
}

func RecordCacheMiss(service, cacheType string) {
	CacheMisses.WithLabelValues(service, cacheType).Inc()
}

func RecordLoginAttempt(status, reason string) {
	LoginAttempts.WithLabelValues(status, reason).Inc()
}

func RecordMessage(messageType, direction string, size float64) {
	MessagesProcessed.WithLabelValues(messageType, direction).Inc()
	MessageSize.WithLabelValues(messageType, direction).Observe(size)
}

func UpdatePlayerCount(world, region string, count float64) {
	PlayersOnline.WithLabelValues(world, region).Set(count)
}

func UpdateEntityCount(world, entityType string, count float64) {
	EntityCount.WithLabelValues(world, entityType).Set(count)
}

// GinMiddleware returns a gin middleware for recording HTTP metrics
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		
		// Process request
		c.Next()
		
		// Record metrics
		duration := time.Since(startTime).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		
		service := c.GetString("service")
		if service == "" {
			service = "gateway"
		}
		
		RecordRequestDuration(service, method, statusCode, duration)
	}
}