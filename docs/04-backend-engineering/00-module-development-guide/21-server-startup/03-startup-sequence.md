---
title: Startup Sequence
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Startup Overview](01-startup-overview.md)"
  - "[Wire App Initialization](02-wire-app.md)"
  - "[Graceful Shutdown](07-graceful-shutdown.md)"
---

# Startup Sequence

## Full main.go

```go
// cmd/server/main.go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func main() {
    // 1. Configure logger
    log.Logger = zerolog.New(os.Stdout).
        With().
        Timestamp().
        Str("service", "awoerp").
        Logger()

    // 2. Load config
    cfg, err := loadConfig()
    if err != nil {
        log.Fatal().Err(err).Msg("failed to load config")
    }

    // 3. Initialize tracer
    tracerProvider, err := initTracer(cfg.OTel)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to initialize tracer")
    }
    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        _ = tracerProvider.Shutdown(ctx)
    }()

    // 4. Build full DI graph
    app, err := InitializeApp(cfg)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to initialize application")
    }
    defer app.DBPool.Close()

    // 5. Register event subscriptions
    for _, sub := range app.Subscribers {
        if err := sub.Register(app.EventBus); err != nil {
            log.Fatal().Err(err).Msg("failed to register event subscriptions")
        }
    }

    // 6. Register Temporal schedules
    if err := registerSchedules(context.Background(), app.TemporalClient, cfg); err != nil {
        log.Warn().Err(err).Msg("failed to register some Temporal schedules — continuing")
    }

    // 7. Start Temporal workers
    for _, w := range app.TemporalWorkers {
        go func(w worker.Worker) {
            if err := w.Start(); err != nil {
                log.Error().Err(err).Msg("temporal worker failed")
            }
        }(w)
    }

    // 8. Start health probe server on separate port
    go startHealthServer(app, cfg.HealthPort)

    // 9. Start metrics server
    go startMetricsServer(cfg.MetricsPort)

    // 10. Start HTTP server (blocking)
    go func() {
        log.Info().Str("addr", cfg.HTTPAddr).Msg("starting HTTP server")
        if err := app.Fiber.Listen(cfg.HTTPAddr); err != nil {
            log.Error().Err(err).Msg("fiber server error")
        }
    }()

    // 11. Wait for shutdown signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    <-quit

    log.Info().Msg("shutting down gracefully (30s timeout)")

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := app.Fiber.ShutdownWithContext(ctx); err != nil {
        log.Error().Err(err).Msg("graceful shutdown error")
    }

    for _, w := range app.TemporalWorkers {
        w.Stop()
    }

    log.Info().Msg("shutdown complete")
}
```

## Config Struct

```go
// cmd/server/config.go
package main

import (
    "fmt"
    "os"
    "strconv"
)

type Config struct {
    HTTPAddr    string
    HealthPort  int
    MetricsPort int
    DatabaseURL string
    RedisURL    string
    TemporalURL string
    OTel        OTelConfig
}

type OTelConfig struct {
    Endpoint    string
    ServiceName string
}

func loadConfig() (Config, error) {
    cfg := Config{
        HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
        HealthPort:  getEnvInt("HEALTH_PORT", 8081),
        MetricsPort: getEnvInt("METRICS_PORT", 9090),
        DatabaseURL: os.Getenv("DATABASE_URL"),
        RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
        TemporalURL: getEnv("TEMPORAL_URL", "localhost:7233"),
        OTel: OTelConfig{
            Endpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
            ServiceName: getEnv("OTEL_SERVICE_NAME", "awoerp"),
        },
    }

    if cfg.DatabaseURL == "" {
        return cfg, fmt.Errorf("DATABASE_URL is required")
    }
    return cfg, nil
}

func getEnv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func getEnvInt(key string, def int) int {
    if v := os.Getenv(key); v != "" {
        if i, err := strconv.Atoi(v); err == nil {
            return i
        }
    }
    return def
}
```
