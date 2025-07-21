package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	
	_ "github.com/lib/pq"
	"github.com/mmorpg-template/backend/internal/ports"
	"github.com/mmorpg-template/backend/pkg/logger"
)

// PostgresDB implements the ports.Database interface for PostgreSQL
type PostgresDB struct {
	db     *sql.DB
	config *ports.DatabaseConfig
	log    logger.Logger
}

// NewPostgresDB creates a new PostgreSQL database adapter
func NewPostgresDB(config *ports.DatabaseConfig, log logger.Logger) *PostgresDB {
	return &PostgresDB{
		config: config,
		log:    log,
	}
}

// Connect establishes a connection to the PostgreSQL database
func (p *PostgresDB) Connect(ctx context.Context) error {
	db, err := sql.Open("postgres", p.config.URL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	
	// Configure connection pool
	db.SetMaxOpenConns(p.config.MaxConnections)
	db.SetMaxIdleConns(p.config.MaxIdleConns)
	db.SetConnMaxLifetime(p.config.ConnMaxLifetime)
	
	// Test connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}
	
	p.db = db
	p.log.Info("Connected to PostgreSQL database")
	return nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Ping checks if the database connection is alive
func (p *PostgresDB) Ping(ctx context.Context) error {
	if p.db == nil {
		return ports.ErrConnection
	}
	return p.db.PingContext(ctx)
}

// BeginTx starts a new transaction
func (p *PostgresDB) BeginTx(ctx context.Context) (ports.Transaction, error) {
	if p.db == nil {
		return nil, ports.ErrConnection
	}
	
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	return &postgresTx{tx: tx}, nil
}

// Query executes a query that returns rows
func (p *PostgresDB) Query(ctx context.Context, query string, args ...interface{}) (ports.Rows, error) {
	if p.db == nil {
		return nil, ports.ErrConnection
	}
	
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	
	return &postgresRows{rows: rows}, nil
}

// QueryRow executes a query that returns at most one row
func (p *PostgresDB) QueryRow(ctx context.Context, query string, args ...interface{}) ports.Row {
	if p.db == nil {
		return &postgresRow{err: ports.ErrConnection}
	}
	
	row := p.db.QueryRowContext(ctx, query, args...)
	return &postgresRow{row: row}
}

// Exec executes a query that doesn't return rows
func (p *PostgresDB) Exec(ctx context.Context, query string, args ...interface{}) (ports.Result, error) {
	if p.db == nil {
		return nil, ports.ErrConnection
	}
	
	result, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	
	return &postgresResult{result: result}, nil
}

// Prepare creates a prepared statement
func (p *PostgresDB) Prepare(ctx context.Context, query string) (ports.Statement, error) {
	if p.db == nil {
		return nil, ports.ErrConnection
	}
	
	stmt, err := p.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	
	return &postgresStmt{stmt: stmt}, nil
}

// Transaction wrapper
type postgresTx struct {
	tx *sql.Tx
}

func (t *postgresTx) Commit() error {
	return t.tx.Commit()
}

func (t *postgresTx) Rollback() error {
	return t.tx.Rollback()
}

func (t *postgresTx) Query(ctx context.Context, query string, args ...interface{}) (ports.Rows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &postgresRows{rows: rows}, nil
}

func (t *postgresTx) QueryRow(ctx context.Context, query string, args ...interface{}) ports.Row {
	row := t.tx.QueryRowContext(ctx, query, args...)
	return &postgresRow{row: row}
}

func (t *postgresTx) Exec(ctx context.Context, query string, args ...interface{}) (ports.Result, error) {
	result, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &postgresResult{result: result}, nil
}

func (t *postgresTx) Prepare(ctx context.Context, query string) (ports.Statement, error) {
	stmt, err := t.tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &postgresStmt{stmt: stmt}, nil
}

// Rows wrapper
type postgresRows struct {
	rows *sql.Rows
}

func (r *postgresRows) Next() bool {
	return r.rows.Next()
}

func (r *postgresRows) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

func (r *postgresRows) Close() error {
	return r.rows.Close()
}

func (r *postgresRows) Err() error {
	return r.rows.Err()
}

func (r *postgresRows) Columns() ([]string, error) {
	return r.rows.Columns()
}

// Row wrapper
type postgresRow struct {
	row *sql.Row
	err error
}

func (r *postgresRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	err := r.row.Scan(dest...)
	if err == sql.ErrNoRows {
		return ports.ErrNotFound
	}
	return err
}

// Result wrapper
type postgresResult struct {
	result sql.Result
}

func (r *postgresResult) LastInsertId() (int64, error) {
	return r.result.LastInsertId()
}

func (r *postgresResult) RowsAffected() (int64, error) {
	return r.result.RowsAffected()
}

// Statement wrapper
type postgresStmt struct {
	stmt *sql.Stmt
}

func (s *postgresStmt) Query(ctx context.Context, args ...interface{}) (ports.Rows, error) {
	rows, err := s.stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	return &postgresRows{rows: rows}, nil
}

func (s *postgresStmt) QueryRow(ctx context.Context, args ...interface{}) ports.Row {
	row := s.stmt.QueryRowContext(ctx, args...)
	return &postgresRow{row: row}
}

func (s *postgresStmt) Exec(ctx context.Context, args ...interface{}) (ports.Result, error) {
	result, err := s.stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	return &postgresResult{result: result}, nil
}

func (s *postgresStmt) Close() error {
	return s.stmt.Close()
}