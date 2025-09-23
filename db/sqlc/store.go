package db

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"
import (
	"context"
	"fmt"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niiniyare/erp/internal/shared"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// Store defines all functions to execute db queries and transactions
type Store interface {
	Querier
	// Modern tenant context methods (recommended)
	SetTenantContextFromCtx(ctx context.Context) error
	WithTenantFromCtx(ctx context.Context, fn func(context.Context, Store) error) error
	BeginTxWithTenantFromCtx(ctx context.Context) (pgx.Tx, Store, error)

	// Legacy tenant context methods (deprecated, kept for backward compatibility)
	// Deprecated: Use SetTenantContextFromCtx instead
	SetTenantContext(ctx context.Context, tenantID uuid.UUID) error
	// Deprecated: Use WithTenantFromCtx instead
	WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, Store) error) error
	// Deprecated: Use BeginTxWithTenantFromCtx instead
	BeginTxWithTenant(ctx context.Context, tenantID uuid.UUID) (pgx.Tx, Store, error)

	// General transaction methods
	WithTx(ctx context.Context, fn func(context.Context, Store) error) error
	WithTxOptions(ctx context.Context, opts pgx.TxOptions, fn func(context.Context, Store) error) error

	// Connection management
	Close()
	GetPool() *pgxpool.Pool
	HealthCheck(ctx context.Context) error
}

// TxStore represents a store bound to a specific transaction
type TxStore interface {
	Store
	GetTx() pgx.Tx
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// SQLStore provides all functions to execute SQL queries and transactions
type SQLStore struct {
	connPool *pgxpool.Pool
	*Queries
	logger logger.Logger
}

// SQLTxStore represents a store bound to a transaction
type SQLTxStore struct {
	*SQLStore
	tx pgx.Tx
}

// NewStore creates a new store
func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
		logger:   logger.WithFields(logger.Fields{"component": "db"}),
	}
}

// NewStoreWithLogger creates a new store with custom logger
func NewStoreWithLogger(connPool *pgxpool.Pool, log logger.Logger) Store {
	// Handle nil logger case - create a default logger
	var storeLogger logger.Logger
	if log == nil {
		// Create a default logger if none is provided
		storeLogger = logger.WithFields(logger.Fields{"component": "db"})
	} else {
		storeLogger = log.WithFields(logger.Fields{"component": "db"})
	}

	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
		logger:   storeLogger,
	}
}

// NewDB creates a new database connection and returns a Store with retry mechanism
func NewDB(databaseURL string) (Store, error) {
	return NewDBWithConfig(databaseURL, nil)
}

// DBConfig holds database configuration options
type DBConfig struct {
	MaxConns       int32
	MinConns       int32
	Logger         logger.Logger
	EnableTracer   bool
	MaxRetries     int
	RetryDelay     time.Duration
	ConnectTimeout time.Duration
	HealthTimeout  time.Duration
}

// DefaultDBConfig returns default database configuration
func DefaultDBConfig() *DBConfig {
	return &DBConfig{
		MaxConns:       25,
		MinConns:       5,
		EnableTracer:   true,
		MaxRetries:     3,
		RetryDelay:     5 * time.Second,
		ConnectTimeout: 10 * time.Second,
		HealthTimeout:  5 * time.Second,
		Logger:         nil, // Will be handled by NewStoreWithLogger
	}
}

// NewDBWithConfig creates a new database connection with custom configuration and retry logic
func NewDBWithConfig(databaseURL string, cfg *DBConfig) (Store, error) {
	if cfg == nil {
		cfg = DefaultDBConfig()
	}

	var pool *pgxpool.Pool
	var lastErr error

	// Create a default logger if none is provided in config
	var configLogger logger.Logger
	if cfg.Logger == nil {
		// Use a basic logger if none provided - you might want to adjust this based on your logger implementation
		configLogger = logger.WithFields(logger.Fields{"component": "db-config"})
	} else {
		configLogger = cfg.Logger
	}

	// Retry loop for database connection
	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		configLogger.InfoContext(context.Background(), "Attempting database connection",
			logger.Fields{
				"attempt":     attempt,
				"max_retries": cfg.MaxRetries,
			},
		)

		var err error
		pool, err = createConnectionPool(databaseURL, cfg)
		if err != nil {
			lastErr = err
			configLogger.ErrorContext(context.Background(), "Database connection attempt failed",
				logger.Fields{
					"attempt": attempt,
					"error":   err.Error(),
				},
			)

			// If this is not the last attempt, wait before retrying
			if attempt < cfg.MaxRetries {
				configLogger.InfoContext(context.Background(), "Retrying database connection",
					logger.Fields{
						"retry_delay": cfg.RetryDelay.String(),
					},
				)
				time.Sleep(cfg.RetryDelay)
				continue
			}
			break
		}

		// Test connection with timeout
		ctx, cancel := context.WithTimeout(context.Background(), cfg.HealthTimeout)
		err = pool.Ping(ctx)
		cancel()

		if err != nil {
			lastErr = err
			pool.Close()
			configLogger.ErrorContext(context.Background(), "Database ping failed",
				logger.Fields{
					"attempt": attempt,
					"error":   err.Error(),
				},
			)

			if attempt < cfg.MaxRetries {
				time.Sleep(cfg.RetryDelay)
				continue
			}
			break
		}

		// Connection successful
		configLogger.InfoContext(context.Background(), "Database connection established successfully",
			logger.Fields{
				"attempt":   attempt,
				"max_conns": cfg.MaxConns,
				"min_conns": cfg.MinConns,
			},
		)

		// Pass the logger to NewStoreWithLogger, which will handle nil case
		return NewStoreWithLogger(pool, cfg.Logger), nil
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts, last error: %w", cfg.MaxRetries, lastErr)
}

// createConnectionPool creates a new connection pool with the given configuration
func createConnectionPool(databaseURL string, cfg *DBConfig) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool
	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns

	// Set connection timeouts
	if cfg.ConnectTimeout > 0 {
		config.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	}

	// Add tracer if enabled
	if cfg.EnableTracer {
		config.ConnConfig.Tracer = otelpgx.NewTracer()
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return pool, nil
}

// Close closes the database connection pool and resets tenant context
func (s *SQLStore) Close() {
	if s.connPool == nil {
		return
	}

	// Reset tenant context before closing
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.resetTenantContext(ctx); err != nil {
		s.logger.ErrorContext(ctx, "Failed to reset tenant context during close",
			logger.Fields{"error": err.Error()},
		)
	}

	s.connPool.Close()
	s.logger.InfoContext(context.Background(), "Database connection pool closed")
}

// GetPool returns the underlying pgxpool.Pool
func (s *SQLStore) GetPool() *pgxpool.Pool {
	return s.connPool
}

// HealthCheck performs a health check on the database
func (s *SQLStore) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.connPool.Ping(ctx); err != nil {
		s.logger.ErrorContext(ctx, "Database health check failed",
			logger.Fields{"error": err.Error()},
		)
		return fmt.Errorf("database health check failed: %w", err)
	}

	s.logger.DebugContext(ctx, "Database health check successful")
	return nil
}

// resetTenantContext resets the tenant context
func (s *SQLStore) resetTenantContext(ctx context.Context) error {
	query := "SELECT set_config('app.current_tenant_id', NULL, false)"
	_, err := s.connPool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to reset tenant context: %w", err)
	}
	return nil
}

// setTenantContext is a helper method to set tenant context
func (s *SQLStore) setTenantContext(ctx context.Context, exec DBTX, tenantID uuid.UUID, isTransaction bool) error {
	query := "SELECT set_config('app.current_tenant_id', $1, $2)"
	_, err := exec.Exec(ctx, query, tenantID.String(), isTransaction)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to set tenant context",
			logger.Fields{
				"tenant_id":      tenantID.String(),
				"is_transaction": isTransaction,
				"error":          err.Error(),
			},
		)
		return fmt.Errorf("failed to set tenant context for tenant %s: %w", tenantID.String(), err)
	}

	s.logger.DebugContext(ctx, "Tenant context set successfully",
		logger.Fields{
			"tenant_id":      tenantID.String(),
			"is_transaction": isTransaction,
		},
	)
	return nil
}

// getTenantIDFromContext is a helper to extract tenant ID from context
func (s *SQLStore) getTenantIDFromContext(ctx context.Context) (uuid.UUID, error) {
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return uuid.Nil, fmt.Errorf("tenant ID not found in context")
	}
	return tenantID, nil
}

// SetTenantContextFromCtx sets tenant context from the provided context
func (s *SQLStore) SetTenantContextFromCtx(ctx context.Context) error {
	tenantID, err := s.getTenantIDFromContext(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get tenant ID from context", logger.Fields{"error": err.Error()})
		return err
	}
	return s.setTenantContext(ctx, s.connPool, tenantID, false)
}

// SetTenantContext sets tenant context (deprecated)
func (s *SQLStore) SetTenantContext(ctx context.Context, tenantID uuid.UUID) error {
	return s.setTenantContext(ctx, s.connPool, tenantID, false)
}

// WithTenant executes a function with tenant context set (deprecated)
func (s *SQLStore) WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, Store) error) error {
	return s.withTenantTransaction(ctx, tenantID, fn)
}

// WithTenantFromCtx executes a function with tenant context set from context
func (s *SQLStore) WithTenantFromCtx(ctx context.Context, fn func(context.Context, Store) error) error {
	tenantID, err := s.getTenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	return s.withTenantTransaction(ctx, tenantID, fn)
}

// withTenantTransaction is a helper method for tenant-scoped transactions
func (s *SQLStore) withTenantTransaction(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, Store) error) error {
	s.logger.DebugContext(ctx, "Starting tenant-scoped transaction",
		logger.Fields{"tenant_id": tenantID.String()},
	)

	tx, err := s.connPool.Begin(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to begin tenant transaction",
			logger.Fields{
				"tenant_id": tenantID.String(),
				"error":     err.Error(),
			},
		)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Always rollback on function exit, commit will override this
	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && rbErr != pgx.ErrTxClosed {
			s.logger.ErrorContext(ctx, "Failed to rollback tenant transaction",
				logger.Fields{
					"tenant_id": tenantID.String(),
					"error":     rbErr.Error(),
				},
			)
		}
	}()

	// Set tenant context
	if err := s.setTenantContext(ctx, tx, tenantID, true); err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Create store instance with transaction
	txStore := &SQLStore{
		connPool: s.connPool,
		Queries:  s.Queries.WithTx(tx),
		logger:   s.logger.WithFields(logger.Fields{"tenant_id": tenantID.String()}),
	}

	// Execute function
	if err := fn(ctx, txStore); err != nil {
		s.logger.ErrorContext(ctx, "Function execution failed in tenant transaction",
			logger.Fields{
				"tenant_id": tenantID.String(),
				"error":     err.Error(),
			},
		)
		return err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		s.logger.ErrorContext(ctx, "Failed to commit tenant transaction",
			logger.Fields{
				"tenant_id": tenantID.String(),
				"error":     err.Error(),
			},
		)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.DebugContext(ctx, "Tenant-scoped transaction completed successfully",
		logger.Fields{"tenant_id": tenantID.String()},
	)
	return nil
}

// BeginTxWithTenant starts a transaction with tenant context (deprecated)
func (s *SQLStore) BeginTxWithTenant(ctx context.Context, tenantID uuid.UUID) (pgx.Tx, Store, error) {
	return s.beginTxWithTenant(ctx, tenantID, pgx.TxOptions{})
}

// BeginTxWithTenantFromCtx starts a transaction with tenant context from context
func (s *SQLStore) BeginTxWithTenantFromCtx(ctx context.Context) (pgx.Tx, Store, error) {
	tenantID, err := s.getTenantIDFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	return s.beginTxWithTenant(ctx, tenantID, pgx.TxOptions{})
}

// beginTxWithTenant is a helper method for beginning tenant-scoped transactions
func (s *SQLStore) beginTxWithTenant(ctx context.Context, tenantID uuid.UUID, opts pgx.TxOptions) (pgx.Tx, Store, error) {
	s.logger.DebugContext(ctx, "Beginning tenant-scoped transaction",
		logger.Fields{
			"tenant_id":       tenantID.String(),
			"isolation_level": opts.IsoLevel,
			"access_mode":     opts.AccessMode,
			"deferrable_mode": opts.DeferrableMode,
		},
	)

	tx, err := s.connPool.BeginTx(ctx, opts)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to begin tenant transaction",
			logger.Fields{
				"tenant_id": tenantID.String(),
				"error":     err.Error(),
			},
		)
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Set tenant context - prefer stored procedure if available, fallback to set_config
	if err := s.trySetTenantContextStoredProc(ctx, tx, tenantID); err != nil {
		// Fallback to set_config
		if setErr := s.setTenantContext(ctx, tx, tenantID, true); setErr != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				s.logger.ErrorContext(ctx, "Failed to rollback after tenant context error",
					logger.Fields{
						"tenant_id": tenantID.String(),
						"error":     rbErr.Error(),
					},
				)
			}
			return nil, nil, fmt.Errorf("failed to set tenant context: %w", setErr)
		}
	}

	txStore := &SQLStore{
		connPool: s.connPool,
		Queries:  s.Queries.WithTx(tx),
		logger:   s.logger.WithFields(logger.Fields{"tenant_id": tenantID.String()}),
	}

	s.logger.DebugContext(ctx, "Tenant-scoped transaction started successfully",
		logger.Fields{"tenant_id": tenantID.String()},
	)

	return tx, txStore, nil
}

// trySetTenantContextStoredProc attempts to use stored procedure for setting tenant context
func (s *SQLStore) trySetTenantContextStoredProc(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) error {
	_, err := tx.Exec(ctx, "SELECT set_tenant_context($1)", tenantID.String())
	if err != nil {
		s.logger.DebugContext(ctx, "Stored procedure set_tenant_context not available, using fallback",
			logger.Fields{
				"tenant_id": tenantID.String(),
				"error":     err.Error(),
			},
		)
	}
	return err
}

// WithTx executes a function within a transaction
func (s *SQLStore) WithTx(ctx context.Context, fn func(context.Context, Store) error) error {
	return s.WithTxOptions(ctx, pgx.TxOptions{}, fn)
}

// WithTxOptions executes a function within a transaction with custom options
func (s *SQLStore) WithTxOptions(ctx context.Context, opts pgx.TxOptions, fn func(context.Context, Store) error) error {
	s.logger.DebugContext(ctx, "Starting transaction",
		logger.Fields{
			"isolation_level": opts.IsoLevel,
			"access_mode":     opts.AccessMode,
			"deferrable_mode": opts.DeferrableMode,
		},
	)

	tx, err := s.connPool.BeginTx(ctx, opts)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to begin transaction",
			logger.Fields{"error": err.Error()},
		)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Always rollback on function exit, commit will override this
	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && rbErr != pgx.ErrTxClosed {
			s.logger.ErrorContext(ctx, "Failed to rollback transaction",
				logger.Fields{"error": rbErr.Error()},
			)
		}
	}()

	txStore := &SQLStore{
		connPool: s.connPool,
		Queries:  s.Queries.WithTx(tx),
		logger:   s.logger,
	}

	if err := fn(ctx, txStore); err != nil {
		s.logger.ErrorContext(ctx, "Function execution failed in transaction",
			logger.Fields{"error": err.Error()},
		)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.ErrorContext(ctx, "Failed to commit transaction",
			logger.Fields{"error": err.Error()},
		)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.DebugContext(ctx, "Transaction completed successfully")
	return nil
}

// GetTx returns the transaction if this is a transaction store
func (ts *SQLTxStore) GetTx() pgx.Tx {
	return ts.tx
}

// Commit commits the transaction
func (ts *SQLTxStore) Commit(ctx context.Context) error {
	err := ts.tx.Commit(ctx)
	if err != nil {
		ts.logger.ErrorContext(ctx, "Failed to commit transaction",
			logger.Fields{"error": err.Error()},
		)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	ts.logger.DebugContext(ctx, "Transaction committed successfully")
	return nil
}

// Rollback rolls back the transaction
func (ts *SQLTxStore) Rollback(ctx context.Context) error {
	err := ts.tx.Rollback(ctx)
	if err != nil && err != pgx.ErrTxClosed {
		ts.logger.ErrorContext(ctx, "Failed to rollback transaction",
			logger.Fields{"error": err.Error()},
		)
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	ts.logger.DebugContext(ctx, "Transaction rolled back successfully")
	return nil
}

// Example usage patterns:
/*
// Initialize database with custom configuration
dbConfig := &DBConfig{
	MaxConns:       50,
	MinConns:       10,
	Logger:         myLogger, // Can be nil, will be handled gracefully
	MaxRetries:     5,
	RetryDelay:     3 * time.Second,
	ConnectTimeout: 15 * time.Second,
}

store, err := NewDBWithConfig(databaseURL, dbConfig)
if err != nil {
	log.Fatal("Failed to connect to database:", err)
}
defer store.Close() // This will reset tenant context before closing

// Modern approach with context-based tenant ID
err := store.WithTenantFromCtx(ctx, func(ctx context.Context, s Store) error {
    // All queries here will be tenant-scoped
    return s.CreateUser(ctx, params)
})

// Health check with timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := store.HealthCheck(ctx); err != nil {
    log.Error("Database health check failed:", err)
}
*/
