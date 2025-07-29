package character

import (
	"context"
	"database/sql"
)

// DBExecutor interface abstracts database operations for both sql.DB and sql.Tx
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}