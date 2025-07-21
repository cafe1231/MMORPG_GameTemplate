package ports

import (
	"context"
	"time"
)

// Database is the interface for database operations
// This follows the hexagonal architecture pattern to decouple business logic from infrastructure
type Database interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error
	
	// Transaction support
	BeginTx(ctx context.Context) (Transaction, error)
	
	// Query execution
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
	
	// Prepared statements
	Prepare(ctx context.Context, query string) (Statement, error)
}

// Transaction represents a database transaction
type Transaction interface {
	// Transaction control
	Commit() error
	Rollback() error
	
	// Query execution within transaction
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
	
	// Prepared statements within transaction
	Prepare(ctx context.Context, query string) (Statement, error)
}

// Statement represents a prepared statement
type Statement interface {
	// Statement execution
	Query(ctx context.Context, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, args ...interface{}) Row
	Exec(ctx context.Context, args ...interface{}) (Result, error)
	
	// Statement lifecycle
	Close() error
}

// Rows represents the result of a query
type Rows interface {
	// Row iteration
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
	
	// Error handling
	Err() error
	
	// Column information
	Columns() ([]string, error)
}

// Row represents a single row result
type Row interface {
	Scan(dest ...interface{}) error
}

// Result represents the result of an exec command
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL             string
	MaxConnections  int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
}

// Common database errors
var (
	ErrNotFound      = NewError("NOT_FOUND", "Record not found")
	ErrDuplicate     = NewError("DUPLICATE", "Duplicate record")
	ErrConnection    = NewError("CONNECTION", "Database connection error")
	ErrTransaction   = NewError("TRANSACTION", "Transaction error")
	ErrInvalidQuery  = NewError("INVALID_QUERY", "Invalid query")
)

// Error represents a database error
type Error struct {
	Code    string
	Message string
}

func NewError(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func (e *Error) Error() string {
	return e.Message
}