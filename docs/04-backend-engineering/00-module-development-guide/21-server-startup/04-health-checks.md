---
title: Health Checks
portal: 4 — Backend Engineering
section: 00-module-development-guide/21-server-startup
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-startup-overview.md
    title: Startup Overview
  - path: ../08-instrumentation/05-health-and-readiness.md
    title: Health and Readiness
---

# Health Checks

## Health Probe Server

Health probes run on a separate port (default 8081) so they remain accessible even when the main server is overloaded or shutting down:

```go
// internal/server/health.go
package server

import (
    "context"
    "net/http"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

func startHealthServer(app *App, port int) {
    mux := http.NewServeMux()

    mux.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    })

    mux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
        if err := checkReadiness(r.Context(), app); err != nil {
            http.Error(w, `{"status":"not_ready","error":"`+err.Error()+`"}`, http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ready"}`))
    })

    srv := &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: mux,
    }
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Error().Err(err).Msg("health server error")
    }
}

func checkReadiness(ctx context.Context, app *App) error {
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    if err := app.DBPool.Ping(ctx); err != nil {
        return fmt.Errorf("database: %w", err)
    }

    if err := app.RedisClient.Ping(ctx).Err(); err != nil {
        return fmt.Errorf("redis: %w", err)
    }

    return nil
}
```

## Liveness vs. Readiness

| Probe | Endpoint | Returns 200 when |
|-------|----------|-----------------|
| Liveness | `/health/live` | Process is running (always 200) |
| Readiness | `/health/ready` | DB + Redis reachable, migrations complete |

Kubernetes uses liveness to decide whether to restart the pod and readiness to decide whether to send traffic.

Never fail the liveness probe for transient issues (DB slow, cache miss) — this causes unnecessary restarts.

## Startup Probe

During startup, before dependencies are ready, return 503 for readiness:

```go
var ready atomic.Bool // set to true after all deps initialized

mux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
    if !ready.Load() {
        http.Error(w, `{"status":"starting"}`, http.StatusServiceUnavailable)
        return
    }
    // ... full readiness check ...
})

// At end of startup sequence:
ready.Store(true)
log.Info().Msg("server ready")
```

## Metrics Endpoint

Prometheus scrapes `/metrics` on the metrics port:

```go
func startMetricsServer(port int) {
    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())

    srv := &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: mux,
    }
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Error().Err(err).Msg("metrics server error")
    }
}
```
