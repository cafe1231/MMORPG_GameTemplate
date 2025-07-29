package character

import (
	"context"
	"database/sql"
	"fmt"
)

// TransactionManager handles database transactions for character operations
type TransactionManager struct {
	db *sql.DB
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{
		db: db,
	}
}

// ExecuteInTransaction executes a function within a database transaction
func (tm *TransactionManager) ExecuteInTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // Re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TransactionalRepositories holds repositories that can work within a transaction
type TransactionalRepositories struct {
	Character  *PostgresCharacterRepository
	Appearance *PostgresAppearanceRepository
	Stats      *PostgresStatsRepository
	Position   *PostgresPositionRepository
}

// NewTransactionalRepositories creates repositories that use the given transaction
func NewTransactionalRepositories(tx *sql.Tx) *TransactionalRepositories {
	return &TransactionalRepositories{
		Character:  NewTransactionalCharacterRepository(tx),
		Appearance: NewTransactionalAppearanceRepository(tx),
		Stats:      NewTransactionalStatsRepository(tx),
		Position:   NewTransactionalPositionRepository(tx),
	}
}

// txDBWrapper wraps a transaction to implement the subset of sql.DB interface we need
type txDBWrapper struct {
	tx *sql.Tx
}

func (w *txDBWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return w.tx.ExecContext(ctx, query, args...)
}

func (w *txDBWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return w.tx.QueryContext(ctx, query, args...)
}

func (w *txDBWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return w.tx.QueryRowContext(ctx, query, args...)
}