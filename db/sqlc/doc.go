// Package db provides a PostgreSQL-backed store with multi-tenant support,
// connection pooling, retry logic, and OpenTelemetry tracing.
//
// # Basic Usage
//
//	store, err := db.NewDBWithConfig(databaseURL, &db.DBConfig{
//	    MaxConns:   50,
//	    MaxRetries: 5,
//	})
//
// # Tenant-Scoped Queries
//
//	err := store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
//	    return s.CreateUser(ctx, params)
//	})
package db
