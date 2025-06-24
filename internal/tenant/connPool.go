// Package tenant provides tenant-aware database connection management
// using PostgreSQL Row-Level Security (RLS) and pgx/v5 connection pooling.
package tenant

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantContextKey is the type for context keys to avoid collisions
type TenantContextKey struct{}

var (
	tenantCtxKey = &TenantContextKey{}
	poolCache    = make(map[int]*pgxpool.Pool)
	poolMutex    sync.RWMutex
)

// WithTenant returns a new context with the tenant ID set
func WithTenant(ctx context.Context, tenantID int) context.Context {
	return context.WithValue(ctx, tenantCtxKey, tenantID)
}

// FromContext retrieves the tenant ID from the context
func FromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(tenantCtxKey).(int)
	return id, ok
}

// ResetContext returns a context without tenant information
func ResetContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, tenantCtxKey, nil)
}

// PoolConfig holds configuration for tenant-aware connection pools
type PoolConfig struct {
	ConnString  string // Base connection string without tenant context
	MinConns    int32
	MaxConns    int32
	MaxIdleTime string // e.g., "5m"
}

// TenantAwarePool manages tenant-specific connection pools
type TenantAwarePool struct {
	config PoolConfig
}

// NewTenantAwarePool creates a new tenant-aware pool manager
func NewTenantAwarePool(config PoolConfig) *TenantAwarePool {
	return &TenantAwarePool{config: config}
}

// GetPool retrieves or creates a connection pool for a specific tenant
func (tap *TenantAwarePool) GetPool(ctx context.Context, tenantID int) (*pgxpool.Pool, error) {
	poolMutex.RLock()
	pool, exists := poolCache[tenantID]
	poolMutex.RUnlock()

	if exists {
		return pool, nil
	}

	return tap.createPool(tenantID)
}

// createPool creates a new connection pool with tenant context
func (tap *TenantAwarePool) createPool(tenantID int) (*pgxpool.Pool, error) {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	// Double-check in case of race condition
	if pool, exists := poolCache[tenantID]; exists {
		return pool, nil
	}

	// Configure connection pool
	config, err := pgxpool.ParseConfig(tap.config.ConnString)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %w", err)
	}

	// Set pool size parameters
	config.MinConns = tap.config.MinConns
	// config.MaxConns = tap.config.MaxConns

	if tap.config.MaxIdleTime != "" {
		duration, err := time.ParseDuration(tap.config.MaxIdleTime)
		if err != nil {
			return nil, fmt.Errorf("invalid max idle time: %w", err)
		}
		config.MaxConnIdleTime = duration
	}

	// Add hook to set tenant context for new connections
	config.BeforeConnect = func(ctx context.Context, connConfig *pgx.ConnConfig) error {
		connConfig.RuntimeParams["application_name"] = fmt.Sprintf("tenant-%d", tenantID)
		return nil
	}

	// Create the pool
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	// Add to cache
	poolCache[tenantID] = pool
	return pool, nil
}

// Exec executes SQL with tenant context
func (tap *TenantAwarePool) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	tenantID, ok := FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant context missing")
	}

	pool, err := tap.GetPool(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Set tenant context for this operation
	_, err = pool.Exec(ctx, "SET app.current_tenant_id = $1", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	return pool.Exec(ctx, sql, args...)
}

// Query executes a query with tenant context
func (tap *TenantAwarePool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	tenantID, ok := FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant context missing")
	}

	pool, err := tap.GetPool(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Set tenant context for this operation
	_, err = pool.Exec(ctx, "SET app.current_tenant_id = $1", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	return pool.Query(ctx, sql, args...)
}

// QueryRow executes a query that returns a single row with tenant context
func (tap *TenantAwarePool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	tenantID, ok := FromContext(ctx)
	if !ok {
		return &errorRow{err: fmt.Errorf("tenant context missing")}
	}

	pool, err := tap.GetPool(ctx, tenantID)
	if err != nil {
		return &errorRow{err: err}
	}

	// Set tenant context for this operation
	_, err = pool.Exec(ctx, "SET app.current_tenant_id = $1", tenantID)
	if err != nil {
		return &errorRow{err: fmt.Errorf("failed to set tenant context: %w", err)}
	}

	return pool.QueryRow(ctx, sql, args...)
}

// Begin starts a transaction with tenant context
func (tap *TenantAwarePool) Begin(ctx context.Context) (pgx.Tx, error) {
	tenantID, ok := FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant context missing")
	}

	pool, err := tap.GetPool(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Set tenant context for this transaction
	_, err = pool.Exec(ctx, "SET app.current_tenant_id = $1", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	return pool.Begin(ctx)
}

// CloseAll closes all connection pools
func (tap *TenantAwarePool) CloseAll() {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	for tenantID, pool := range poolCache {
		pool.Close()
		delete(poolCache, tenantID)
	}
}

// Helper for handling errors in QueryRow
type errorRow struct {
	err error
}

func (r *errorRow) Scan(dest ...interface{}) error {
	return r.err
}
