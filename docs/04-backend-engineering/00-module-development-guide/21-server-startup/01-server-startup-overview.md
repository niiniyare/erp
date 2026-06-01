---
title: Server Startup Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead, devops]
related:
  - "[Wire Registration](../09-wire-registration/01-wire-registration-overview.md)"
  - "[DevOps Overview](../../../../06-devops/01-devops-overview.md)"
  - "[Monitoring and Alerting](../../../../06-devops/06-monitoring.md)"
---

# Server Startup Overview

## Startup Sequence

```
main()
  │
  ├── parse flags (migrate | serve)
  │
  ├── [migrate] → run golang-migrate, exit 0 or 1
  │
  └── [serve]
        │
        ├── load Config from env
        ├── InitializeApp() via Wire
        │     ├── connect PostgreSQL pool
        │     ├── connect Redis
        │     ├── connect Temporal client
        │     ├── build all service + handler trees
        │     └── return *App
        │
        ├── start health server :8081
        ├── start metrics server :9090
        ├── start Temporal worker goroutines
        ├── start Fiber HTTP server :8080
        │
        └── wait for SIGINT / SIGTERM
              │
              └── graceful shutdown (5s timeout)
```

## main.go

```go
// cmd/server/main.go
package main

import (
    "context"
    "flag"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/rs/zerolog/log"
)

func main() {
    migrate := flag.Bool("migrate", false, "run migrations and exit")
    flag.Parse()

    cfg := loadConfig()

    if *migrate {
        if err := runMigrations(cfg.DatabaseURL); err != nil {
            log.Fatal().Err(err).Msg("migration failed")
        }
        os.Exit(0)
    }

    app, cleanup, err := InitializeApp(cfg)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to initialize application")
    }
    defer cleanup()

    app.Start()
    waitForShutdown(app)
}

func waitForShutdown(app *App) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := app.Shutdown(ctx); err != nil {
        log.Error().Err(err).Msg("server forced to shutdown")
    }
}
```

## App struct

```go
// cmd/server/app.go
package main

type App struct {
    fiber       *fiber.App
    healthSrv   *http.Server
    metricsSrv  *http.Server
    workers     []*temporalWorker
    logger      zerolog.Logger
}

func (a *App) Start() {
    // 1. Health server (port 8081) — must be up before HTTP server accepts traffic
    go func() {
        if err := a.healthSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            a.logger.Fatal().Err(err).Msg("health server failed")
        }
    }()

    // 2. Metrics server (port 9090)
    go func() {
        if err := a.metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            a.logger.Fatal().Err(err).Msg("metrics server failed")
        }
    }()

    // 3. Temporal workers
    for _, w := range a.workers {
        go func(worker *temporalWorker) {
            if err := worker.Start(); err != nil {
                a.logger.Error().Err(err).Str("queue", worker.taskQueue).Msg("temporal worker failed")
            }
        }(w)
    }

    // 4. HTTP server (port 8080) — blocks until shutdown
    go func() {
        if err := a.fiber.Listen(":8080"); err != nil {
            a.logger.Fatal().Err(err).Msg("HTTP server failed")
        }
    }()
}

func (a *App) Shutdown(ctx context.Context) error {
    a.logger.Info().Msg("shutting down server")

    // Stop accepting new connections
    if err := a.fiber.ShutdownWithContext(ctx); err != nil {
        return err
    }

    // Stop Temporal workers (drain in-progress activities)
    for _, w := range a.workers {
        w.Stop()
    }

    // Stop health + metrics servers
    a.healthSrv.Shutdown(ctx)
    a.metricsSrv.Shutdown(ctx)

    return nil
}
```

## Health Checks

```go
// cmd/server/health.go
func newHealthServer(db *pgxpool.Pool, redis *redis.Client, tc temporalClient.Client) *http.Server {
    mux := http.NewServeMux()

    mux.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
        w.Write([]byte(`{"status":"live"}`))
    })

    mux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
        checks := map[string]string{}
        status := 200

        if err := db.Ping(r.Context()); err != nil {
            checks["database"] = "error: " + err.Error()
            status = 503
        } else {
            checks["database"] = "ok"
        }

        if err := redis.Ping(r.Context()).Err(); err != nil {
            checks["redis"] = "error: " + err.Error()
            status = 503
        } else {
            checks["redis"] = "ok"
        }

        if err := tc.CheckHealth(r.Context()); err != nil {
            checks["temporal"] = "error: " + err.Error()
            status = 503
        } else {
            checks["temporal"] = "ok"
        }

        b, _ := json.Marshal(map[string]any{
            "status": map[int]string{200: "ready", 503: "not_ready"}[status],
            "checks": checks,
        })
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(status)
        w.Write(b)
    })

    mux.HandleFunc("/health/startup", func(w http.ResponseWriter, r *http.Request) {
        // Used by K8s startup probe — same as readiness
        // Distinguishes startup failures from runtime failures
        w.WriteHeader(200)
        w.Write([]byte(`{"status":"started"}`))
    })

    return &http.Server{Addr: ":8081", Handler: mux}
}
```

## Fiber App Setup

```go
// cmd/server/fiber.go
func newFiberApp(cfg Config) *fiber.App {
    app := fiber.New(fiber.Config{
        // Never expose internal errors to clients
        ErrorHandler: middleware.ErrorHandler,
        // ReadTimeout prevents slow-loris attacks
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  120 * time.Second,
    })

    // Global middleware
    app.Use(
        middleware.RequestID(),
        middleware.Logger(),
        middleware.Recover(),
        middleware.CORS(cfg.AllowedOrigins),
        middleware.RateLimit(cfg.RateLimit),
    )

    return app
}
```

## Config Loading

All configuration from environment variables — no config files:

```go
// cmd/server/config.go
type Config struct {
    DatabaseURL       string
    RedisURL          string
    TemporalAddress   string
    TemporalNamespace string
    JWTSecret         string
    AllowedOrigins    []string
    RateLimit         int
    LogLevel          string
}

func loadConfig() Config {
    return Config{
        DatabaseURL:       mustEnv("DATABASE_URL"),
        RedisURL:          mustEnv("REDIS_URL"),
        TemporalAddress:   envOrDefault("TEMPORAL_ADDRESS", "temporal:7233"),
        TemporalNamespace: envOrDefault("TEMPORAL_NAMESPACE", "default"),
        JWTSecret:         mustEnv("JWT_SECRET"),
        AllowedOrigins:    strings.Split(envOrDefault("ALLOWED_ORIGINS", ""), ","),
        RateLimit:         envInt("RATE_LIMIT_RPM", 600),
        LogLevel:          envOrDefault("LOG_LEVEL", "info"),
    }
}

func mustEnv(key string) string {
    v := os.Getenv(key)
    if v == "" {
        log.Fatal().Str("key", key).Msg("required environment variable not set")
    }
    return v
}
```

## Migration Subcommand

The server binary handles its own migrations via the `migrate` flag:

```bash
# Run migrations only (used as Kubernetes Job before deployment)
/server -migrate

# Start server (default)
/server
```

```go
func runMigrations(dbURL string) error {
    m, err := migrate.New(
        "file:///db/migration",
        dbURL,
    )
    if err != nil {
        return fmt.Errorf("creating migrator: %w", err)
    }
    if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
        return fmt.Errorf("running migrations: %w", err)
    }
    return nil
}
```

## Startup Ordering Rules

**Mandatory order: migrations → deps → health server → HTTP server**

1. Migrations must complete before serving traffic — health/ready returns 503 until done
2. Health server starts before HTTP server so K8s doesn't route traffic too early
3. Temporal workers start after HTTP server — worker startup failure must not block HTTP
4. Never start the HTTP server if database pool connection fails

## Port Summary

| Port | Service | Protocol |
|------|---------|----------|
| 8080 | HTTP API | HTTP/1.1 |
| 8081 | Health checks | HTTP/1.1 |
| 9090 | Prometheus metrics | HTTP/1.1 |

All three expose HTTP — no gRPC or TCP for external traffic.
