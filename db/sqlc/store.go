package db

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"
import (
	"context"
	"fmt"
	"time"

	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"github.com/exaring/otelpgx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store defines all functions to execute db queries and transactions.
type Store interface {
	Querier

	// Tenant context — context-based (preferred).
	SetTenantContextFromCtx(ctx context.Context) error
	WithTenantFromCtx(ctx context.Context, fn func(context.Context, Store) error) error
	BeginTxWithTenantFromCtx(ctx context.Context) (pgx.Tx, Store, error)

	// ValidateTenantContext re-validates the current transaction-local tenant
	// context against the live tenants table. Use in long-running background
	// jobs where tenant status may change after the context was first set.
	// Returns the active tenant UUID on success; returns an error on any failure.
	ValidateTenantContext(ctx context.Context) (uuid.UUID, error)

	// Tenant context — explicit UUID (deprecated, kept for backward compatibility).
	// Deprecated: Use SetTenantContextFromCtx instead.
	SetTenantContext(ctx context.Context, tenantID uuid.UUID) error
	// Deprecated: Use WithTenantFromCtx instead.
	WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, Store) error) error
	// Deprecated: Use BeginTxWithTenantFromCtx instead.
	BeginTxWithTenant(ctx context.Context, tenantID uuid.UUID) (pgx.Tx, Store, error)

	// General transactions.
	WithTx(ctx context.Context, fn func(context.Context, Store) error) error
	WithTxOptions(ctx context.Context, opts pgx.TxOptions, fn func(context.Context, Store) error) error

	// Connection management.
	Close()
	GetPool() *pgxpool.Pool
	HealthCheck(ctx context.Context) error
}

// TxStore is a Store bound to an explicit transaction.
// Returned by BeginTxWithTenant / BeginTxWithTenantFromCtx so callers can
// commit or roll back manually when the WithTenant* helpers are too coarse.
type TxStore interface {
	Store
	GetTx() pgx.Tx
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// SQLStore provides all functions to execute SQL queries and transactions.
type SQLStore struct {
	connPool *pgxpool.Pool
	*Queries
	logger logger.Logger
}

// SQLTxStore is an SQLStore bound to a single transaction.
type SQLTxStore struct {
	*SQLStore
	tx pgx.Tx
}

// ------------------------------------------------------------------------------------------------
// Constructors
// ------------------------------------------------------------------------------------------------

// NewStore creates a new store backed by an existing pool.
func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
		logger:   logger.WithFields(logger.Fields{"component": "db"}),
	}
}

// NewStoreWithLogger creates a new store with a custom logger.
// A nil logger falls back to the default structured logger.
func NewStoreWithLogger(connPool *pgxpool.Pool, log logger.Logger) Store {
	var storeLogger logger.Logger
	if log == nil {
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

// NewDB creates a new database connection using default configuration.
func NewDB(databaseURL string) (Store, error) {
	return NewDBWithConfig(databaseURL, nil)
}

// ------------------------------------------------------------------------------------------------
// Configuration
// ------------------------------------------------------------------------------------------------

// DBConfig holds database connection and pool configuration.
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

// DefaultDBConfig returns conservative, production-safe defaults.
func DefaultDBConfig() *DBConfig {
	return &DBConfig{
		MaxConns:       25,
		MinConns:       5,
		EnableTracer:   true,
		MaxRetries:     3,
		RetryDelay:     5 * time.Second,
		ConnectTimeout: 10 * time.Second,
		HealthTimeout:  5 * time.Second,
	}
}

// NewDBWithConfig creates a database connection with retry logic and custom configuration.
// cfg may be nil, in which case DefaultDBConfig is used.
func NewDBWithConfig(databaseURL string, cfg *DBConfig) (Store, error) {
	if cfg == nil {
		cfg = DefaultDBConfig()
	}

	// Ensure we always have a logger for the connection phase, even if cfg.Logger is nil.
	// NewStoreWithLogger handles the nil case for the returned Store.
	cfgLogger := cfg.Logger
	if cfgLogger == nil {
		cfgLogger = logger.WithFields(logger.Fields{"component": "db-config"})
	}

	var (
		pool    *pgxpool.Pool
		lastErr error
	)

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		cfgLogger.InfoContext(context.Background(), "Attempting database connection",
			logger.Fields{"attempt": attempt, "max_retries": cfg.MaxRetries},
		)

		var err error
		pool, err = createConnectionPool(databaseURL, cfg)
		if err != nil {
			lastErr = err
			cfgLogger.ErrorContext(context.Background(), "Connection pool creation failed",
				logger.Fields{"attempt": attempt, "error": err.Error()},
			)
			if attempt < cfg.MaxRetries {
				time.Sleep(cfg.RetryDelay)
			}
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), cfg.HealthTimeout)
		err = pool.Ping(ctx)
		cancel()

		if err != nil {
			lastErr = err
			pool.Close()
			cfgLogger.ErrorContext(context.Background(), "Database ping failed",
				logger.Fields{"attempt": attempt, "error": err.Error()},
			)
			if attempt < cfg.MaxRetries {
				time.Sleep(cfg.RetryDelay)
			}
			continue
		}

		cfgLogger.InfoContext(context.Background(), "Database connection established",
			logger.Fields{"attempt": attempt, "max_conns": cfg.MaxConns, "min_conns": cfg.MinConns},
		)

		return NewStoreWithLogger(pool, cfg.Logger), nil
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", cfg.MaxRetries, lastErr)
}

// createConnectionPool builds a pgxpool.Pool from a URL and config.
func createConnectionPool(databaseURL string, cfg *DBConfig) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns

	if cfg.ConnectTimeout > 0 {
		config.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	}

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

// ------------------------------------------------------------------------------------------------
// Connection management
// ------------------------------------------------------------------------------------------------

// Close resets tenant context and closes the connection pool.
//
// Note: set_tenant_context() uses set_config(..., TRUE) — transaction-local
// scope — so context is already cleared on every COMMIT/ROLLBACK.
// The call to clear_tenant_context() here is a best-effort hygiene measure
// before pool shutdown and may target any idle connection in the pool.
func (s *SQLStore) Close() {
	if s.connPool == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.clearTenantContext(ctx); err != nil {
		s.logger.ErrorContext(ctx, "Failed to clear tenant context during close",
			logger.Fields{"error": err.Error()},
		)
	}

	s.connPool.Close()
	s.logger.InfoContext(context.Background(), "Database connection pool closed")
}

// GetPool returns the underlying pgxpool.Pool.
func (s *SQLStore) GetPool() *pgxpool.Pool {
	return s.connPool
}

// HealthCheck pings the database with a 5-second timeout.
func (s *SQLStore) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.connPool.Ping(ctx); err != nil {
		s.logger.ErrorContext(ctx, "Database health check failed", logger.Fields{"error": err.Error()})
		return fmt.Errorf("database health check failed: %w", err)
	}

	s.logger.DebugContext(ctx, "Database health check successful")
	return nil
}

// ------------------------------------------------------------------------------------------------
// Tenant context helpers (internal)
// ------------------------------------------------------------------------------------------------

// setTenantContext calls the set_tenant_context() stored procedure, which:
//   - Verifies the tenant exists and is not soft-deleted.
//   - Verifies the tenant Status is ACTIVE.
//   - Writes to app.current_tenant_id with transaction-local scope (is_local = TRUE).
//
// There is intentionally no set_config fallback — bypassing the stored procedure
// would skip existence and status validation, which is a security regression.
func (s *SQLStore) setTenantContext(ctx context.Context, exec DBTX, tenantID uuid.UUID) error {
	_, err := exec.Exec(ctx, "SELECT set_tenant_context($1)", tenantID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to set tenant context",
			logger.Fields{"tenant_id": tenantID.String(), "error": err.Error()},
		)
		return fmt.Errorf("failed to set tenant context for tenant %s: %w", tenantID.String(), err)
	}

	s.logger.DebugContext(ctx, "Tenant context set", logger.Fields{"tenant_id": tenantID.String()})
	return nil
}

// clearTenantContext calls clear_tenant_context() which resets app.current_tenant_id
// to ” (empty string). current_tenant_id() treats ” as NULL via NULLIF.
func (s *SQLStore) clearTenantContext(ctx context.Context) error {
	_, err := s.connPool.Exec(ctx, "SELECT clear_tenant_context()")
	if err != nil {
		return fmt.Errorf("failed to clear tenant context: %w", err)
	}
	return nil
}

// getTenantIDFromContext extracts the tenant UUID from ctx.
func (s *SQLStore) getTenantIDFromContext(ctx context.Context) (uuid.UUID, error) {
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok || tenantID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("tenant ID not found in context")
	}
	return tenantID, nil
}

// ------------------------------------------------------------------------------------------------
// Tenant context — public API (context-based, preferred)
// ------------------------------------------------------------------------------------------------

// SetTenantContextFromCtx sets the transaction-local tenant context from ctx.
func (s *SQLStore) SetTenantContextFromCtx(ctx context.Context) error {
	tenantID, err := s.getTenantIDFromContext(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get tenant ID from context", logger.Fields{"error": err.Error()})
		return err
	}
	return s.setTenantContext(ctx, s.connPool, tenantID)
}

// WithTenantFromCtx runs fn inside a transaction with tenant context set from ctx.
// fn receives a tenant-scoped Store — all queries inside fn are RLS-filtered.
func (s *SQLStore) WithTenantFromCtx(ctx context.Context, fn func(context.Context, Store) error) error {
	tenantID, err := s.getTenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	return s.withTenantTransaction(ctx, tenantID, fn)
}

// BeginTxWithTenantFromCtx starts a transaction with tenant context from ctx.
// Returns the raw pgx.Tx alongside a tenant-scoped Store (which also implements TxStore).
// The caller is responsible for committing or rolling back the transaction.
func (s *SQLStore) BeginTxWithTenantFromCtx(ctx context.Context) (pgx.Tx, Store, error) {
	tenantID, err := s.getTenantIDFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	return s.beginTxWithTenant(ctx, tenantID, pgx.TxOptions{})
}

// ValidateTenantContext re-validates the current transaction-local tenant context
// against the live tenants table. Delegates to the validate_tenant_context()
// stored procedure, which checks both existence and ACTIVE status.
//
// Use this in long-running background jobs where tenant status may change after
// the context was first set.
func (s *SQLStore) ValidateTenantContext(ctx context.Context) (uuid.UUID, error) {
	var tenantID uuid.UUID
	err := s.connPool.QueryRow(ctx, "SELECT validate_tenant_context()").Scan(&tenantID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Tenant context validation failed", logger.Fields{"error": err.Error()})
		return uuid.Nil, fmt.Errorf("tenant context validation failed: %w", err)
	}
	return tenantID, nil
}

// ------------------------------------------------------------------------------------------------
// Tenant context — public API (explicit UUID, deprecated)
// ------------------------------------------------------------------------------------------------

// SetTenantContext sets the transaction-local tenant context from an explicit UUID.
// Deprecated: Use SetTenantContextFromCtx instead.
func (s *SQLStore) SetTenantContext(ctx context.Context, tenantID uuid.UUID) error {
	return s.setTenantContext(ctx, s.connPool, tenantID)
}

// WithTenant runs fn inside a transaction with an explicit tenant UUID.
// Deprecated: Use WithTenantFromCtx instead.
func (s *SQLStore) WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, Store) error) error {
	return s.withTenantTransaction(ctx, tenantID, fn)
}

// BeginTxWithTenant starts a transaction with an explicit tenant UUID.
// Deprecated: Use BeginTxWithTenantFromCtx instead.
func (s *SQLStore) BeginTxWithTenant(ctx context.Context, tenantID uuid.UUID) (pgx.Tx, Store, error) {
	return s.beginTxWithTenant(ctx, tenantID, pgx.TxOptions{})
}

// ------------------------------------------------------------------------------------------------
// Tenant transaction internals
// ------------------------------------------------------------------------------------------------

// withTenantTransaction runs fn inside a transaction with tenant context set.
// Rolls back automatically on fn error; commits on success.
func (s *SQLStore) withTenantTransaction(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, Store) error) error {
	s.logger.DebugContext(ctx, "Starting tenant-scoped transaction",
		logger.Fields{"tenant_id": tenantID.String()},
	)

	tx, err := s.connPool.Begin(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to begin tenant transaction",
			logger.Fields{"tenant_id": tenantID.String(), "error": err.Error()},
		)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && rbErr != pgx.ErrTxClosed {
			s.logger.ErrorContext(ctx, "Failed to rollback tenant transaction",
				logger.Fields{"tenant_id": tenantID.String(), "error": rbErr.Error()},
			)
		}
	}()

	if err := s.setTenantContext(ctx, tx, tenantID); err != nil {
		return err
	}

	txStore := &SQLStore{
		connPool: s.connPool,
		Queries:  s.Queries.WithTx(tx),
		logger:   s.logger.WithFields(logger.Fields{"tenant_id": tenantID.String()}),
	}

	if err := fn(ctx, txStore); err != nil {
		s.logger.ErrorContext(ctx, "Function execution failed in tenant transaction",
			logger.Fields{"tenant_id": tenantID.String(), "error": err.Error()},
		)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.ErrorContext(ctx, "Failed to commit tenant transaction",
			logger.Fields{"tenant_id": tenantID.String(), "error": err.Error()},
		)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.DebugContext(ctx, "Tenant-scoped transaction completed",
		logger.Fields{"tenant_id": tenantID.String()},
	)
	return nil
}

// beginTxWithTenant starts a transaction, sets tenant context, and returns a
// TxStore (wrapped as Store) for manual commit/rollback by the caller.
//
// set_tenant_context() is the only path for setting tenant context — there is
// no fallback to raw set_config because that would bypass tenant validation.
func (s *SQLStore) beginTxWithTenant(ctx context.Context, tenantID uuid.UUID, opts pgx.TxOptions) (pgx.Tx, Store, error) {
	s.logger.DebugContext(ctx, "Beginning tenant-scoped transaction",
		logger.Fields{
			"tenant_id":       tenantID.String(),
			"isolation_level": opts.IsoLevel,
			"access_mode":     opts.AccessMode,
		},
	)

	tx, err := s.connPool.BeginTx(ctx, opts)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to begin tenant transaction",
			logger.Fields{"tenant_id": tenantID.String(), "error": err.Error()},
		)
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := s.setTenantContext(ctx, tx, tenantID); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil && rbErr != pgx.ErrTxClosed {
			s.logger.ErrorContext(ctx, "Failed to rollback after tenant context error",
				logger.Fields{"tenant_id": tenantID.String(), "error": rbErr.Error()},
			)
		}
		return nil, nil, err
	}

	txStore := &SQLTxStore{
		SQLStore: &SQLStore{
			connPool: s.connPool,
			Queries:  s.Queries.WithTx(tx),
			logger:   s.logger.WithFields(logger.Fields{"tenant_id": tenantID.String()}),
		},
		tx: tx,
	}

	s.logger.DebugContext(ctx, "Tenant-scoped transaction started",
		logger.Fields{"tenant_id": tenantID.String()},
	)

	return tx, txStore, nil
}

// ------------------------------------------------------------------------------------------------
// General transactions
// ------------------------------------------------------------------------------------------------

// WithTx runs fn inside a transaction with default options.
func (s *SQLStore) WithTx(ctx context.Context, fn func(context.Context, Store) error) error {
	return s.WithTxOptions(ctx, pgx.TxOptions{}, fn)
}

// WithTxOptions runs fn inside a transaction with custom pgx.TxOptions.
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
		s.logger.ErrorContext(ctx, "Failed to begin transaction", logger.Fields{"error": err.Error()})
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && rbErr != pgx.ErrTxClosed {
			s.logger.ErrorContext(ctx, "Failed to rollback transaction", logger.Fields{"error": rbErr.Error()})
		}
	}()

	txStore := &SQLStore{
		connPool: s.connPool,
		Queries:  s.Queries.WithTx(tx),
		logger:   s.logger,
	}

	if err := fn(ctx, txStore); err != nil {
		s.logger.ErrorContext(ctx, "Function execution failed in transaction", logger.Fields{"error": err.Error()})
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.ErrorContext(ctx, "Failed to commit transaction", logger.Fields{"error": err.Error()})
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.DebugContext(ctx, "Transaction completed successfully")
	return nil
}

// ------------------------------------------------------------------------------------------------
// SQLTxStore — manual transaction control
// ------------------------------------------------------------------------------------------------

// GetTx returns the underlying pgx.Tx.
func (ts *SQLTxStore) GetTx() pgx.Tx {
	return ts.tx
}

// Commit commits the transaction.
func (ts *SQLTxStore) Commit(ctx context.Context) error {
	if err := ts.tx.Commit(ctx); err != nil {
		ts.logger.ErrorContext(ctx, "Failed to commit transaction", logger.Fields{"error": err.Error()})
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	ts.logger.DebugContext(ctx, "Transaction committed successfully")
	return nil
}

// Rollback rolls back the transaction. pgx.ErrTxClosed is treated as a no-op
// because it means the transaction was already committed or rolled back.
func (ts *SQLTxStore) Rollback(ctx context.Context) error {
	if err := ts.tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
		ts.logger.ErrorContext(ctx, "Failed to rollback transaction", logger.Fields{"error": err.Error()})
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	ts.logger.DebugContext(ctx, "Transaction rolled back successfully")
	return nil
}
