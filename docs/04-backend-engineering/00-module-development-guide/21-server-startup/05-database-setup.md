---
title: Database Setup
portal: 4 — Backend Engineering
section: 00-module-development-guide/21-server-startup
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-startup-overview.md
    title: Startup Overview
  - path: ../03-database-design/01-schema-overview.md
    title: Schema Overview
---

# Database Setup

## pgxpool Provider

```go
// internal/platform/db/db.go
package db

import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewPool)

// NewPool creates and verifies the database connection pool.
func NewPool(cfg Config) (*pgxpool.Pool, error) {
    poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("parse database URL: %w", err)
    }

    // Pool configuration
    poolCfg.MaxConns          = 25
    poolCfg.MinConns          = 5
    poolCfg.MaxConnLifetime   = 30 * time.Minute
    poolCfg.MaxConnIdleTime   = 5 * time.Minute
    poolCfg.HealthCheckPeriod = 1 * time.Minute

    pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
    if err != nil {
        return nil, fmt.Errorf("create pool: %w", err)
    }

    // Verify connectivity
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := pool.Ping(ctx); err != nil {
        pool.Close()
        return nil, fmt.Errorf("ping database: %w", err)
    }

    return pool, nil
}
```

## Migration Runner

Migrations run at startup using golang-migrate before the Wire graph is built:

```go
// internal/platform/db/migrate.go
package db

import (
    "fmt"

    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(databaseURL, migrationsDir string) error {
    m, err := migrate.New(
        "file://"+migrationsDir,
        databaseURL,
    )
    if err != nil {
        return fmt.Errorf("init migrate: %w", err)
    }
    defer m.Close()

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("run migrations: %w", err)
    }

    return nil
}
```

```go
// In main.go — before InitializeApp
if err := db.RunMigrations(cfg.DatabaseURL, "db/migration"); err != nil {
    log.Fatal().Err(err).Msg("database migration failed")
}
```

## RLS Session Variable

The `Store.WithTenant` helper sets the RLS GUC before every tenant-scoped query:

```go
// internal/platform/db/store.go
package db

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/google/uuid"
)

type Store interface {
    WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(q *sqlc.Queries) error) error
}

type pgStore struct {
    pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) Store {
    return &pgStore{pool: pool}
}

func (s *pgStore) WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(q *sqlc.Queries) error) error {
    return s.pool.AcquireFunc(ctx, func(conn *pgxpool.Conn) error {
        tx, err := conn.Begin(ctx)
        if err != nil {
            return err
        }
        defer tx.Rollback(ctx)

        // Set RLS context variable
        if _, err := tx.Exec(ctx,
            "SET LOCAL app.tenant_id = $1",
            tenantID.String(),
        ); err != nil {
            return fmt.Errorf("set tenant context: %w", err)
        }

        if err := fn(sqlc.New(tx)); err != nil {
            return err
        }

        return tx.Commit(ctx)
    })
}
```

`SET LOCAL` scopes the GUC to the current transaction. It is automatically cleared when the transaction ends.
