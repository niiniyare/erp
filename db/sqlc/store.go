package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store defines all functions to execute db queries and transactions
type Store interface {
	Querier
	// Tenant context methods
	SetTenantContext(ctx context.Context, tenantID uuid.UUID) error
	WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, Store) error) error
	BeginTxWithTenant(ctx context.Context, tenantID uuid.UUID) (pgx.Tx, Store, error)
	WithTx(ctx context.Context, fn func(context.Context, Store) error) error
	// Connection management
	Close()
	GetPool() *pgxpool.Pool
}

// SQLStore provides all functions to execute SQL queries and transactions
type SQLStore struct {
	connPool *pgxpool.Pool
	*Queries
}

// NewStore creates a new store
func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}

// NewDB creates a new database connection and returns a Store
func NewDB(databaseURL string) (Store, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool
	config.MaxConns = 25
	config.MinConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return NewStore(pool), nil
}

// Close closes the database connection pool
func (s *SQLStore) Close() {
	s.connPool.Close()
}

// GetPool returns the underlying pgxpool.Pool
func (s *SQLStore) GetPool() *pgxpool.Pool {
	return s.connPool
}

func (s *SQLStore) SetTenantContext(ctx context.Context, tenantID uuid.UUID) error {
	// Use false instead of true to make it session-level, not transaction-level
	_, err := s.connPool.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, false)", tenantID.String())
	return err
}

// WithTenant executes a function with tenant context set
func (s *SQLStore) WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, Store) error) error {
	tx, err := s.connPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Set tenant context
	_, err = tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID.String())
	if err != nil {
		return err
	}

	// Create store instance with transaction
	txStore := &SQLStore{
		connPool: s.connPool,
		Queries:  s.Queries.WithTx(tx),
	}

	// Execute function
	if err := fn(ctx, txStore); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// BeginTxWithTenant starts a transaction with tenant context
func (s *SQLStore) BeginTxWithTenant(ctx context.Context, tenantID uuid.UUID) (pgx.Tx, Store, error) {
	tx, err := s.connPool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Set tenant context
	_, err = tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID.String())
	if err != nil {
		tx.Rollback(ctx)
		return nil, nil, err
	}

	txStore := &SQLStore{
		connPool: s.connPool,
		Queries:  s.Queries.WithTx(tx),
	}

	return tx, txStore, nil
}

// WithTx executes a function within a transaction
func (s *SQLStore) WithTx(ctx context.Context, fn func(context.Context, Store) error) error {
	tx, err := s.connPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txStore := &SQLStore{
		connPool: s.connPool,
		Queries:  s.Queries.WithTx(tx),
	}

	if err := fn(ctx, txStore); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
