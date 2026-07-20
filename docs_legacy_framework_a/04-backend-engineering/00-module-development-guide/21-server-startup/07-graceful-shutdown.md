> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Graceful Shutdown
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, devops]
related:
  - "[Startup Overview](01-startup-overview.md)"
  - "[Startup Sequence](03-startup-sequence.md)"
  - "[Kubernetes Deployment](../../../06-devops/03-kubernetes-deployment.md)"
---

# Graceful Shutdown

## Why Graceful Shutdown Matters

When Kubernetes rolls out a new pod version, it sends `SIGTERM` to the old pod. Without graceful shutdown:
- In-flight requests get connection-reset errors
- Open DB transactions may be left incomplete
- Temporal activities may be abandoned mid-execution

AwoERP's shutdown sequence ensures all of these are handled cleanly.

## Shutdown Sequence

```
SIGTERM received
  │
  ├── 1. Stop accepting new connections (Fiber ShutdownWithTimeout)
  │
  ├── 2. Wait for in-flight HTTP requests to complete (30s timeout)
  │
  ├── 3. Stop Temporal workers (DrainWorkers)
  │         → workers stop polling, wait for running activities to finish
  │
  ├── 4. Flush event outbox (drain pending events)
  │
  ├── 5. Close Redis connection pool
  │
  ├── 6. Close pgxpool (waits for open transactions)
  │
  └── 7. Flush OTel trace exporter
```

## main.go Implementation

```go
func main() {
    cfg, err := loadConfig()
    if err != nil {
        slog.Error("config failed", "error", err)
        os.Exit(1)
    }

    app, cleanup, err := InitializeApp(cfg)
    if err != nil {
        slog.Error("init failed", "error", err)
        os.Exit(1)
    }

    // Start server in background goroutine
    serverErr := make(chan error, 1)
    go func() {
        slog.Info("server starting", "port", cfg.HTTPPort)
        serverErr <- app.fiberApp.Listen(fmt.Sprintf(":%d", cfg.HTTPPort))
    }()

    // Block until signal or server error
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    select {
    case err := <-serverErr:
        slog.Error("server failed", "error", err)
    case sig := <-quit:
        slog.Info("shutdown signal received", "signal", sig)
    }

    // Initiate graceful shutdown
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := app.fiberApp.ShutdownWithContext(shutdownCtx); err != nil {
        slog.Error("fiber shutdown error", "error", err)
    }

    // Stop Temporal workers
    for _, worker := range app.temporalWorkers {
        worker.Stop()
    }

    // Run Wire-generated cleanup (closes pools, flushes exporters)
    cleanup()

    slog.Info("shutdown complete")
}
```

## `preStop` Hook in Kubernetes

Kubernetes needs a brief delay between removing the pod from the load balancer and sending SIGTERM. Configure `preStop`:

```yaml
lifecycle:
  preStop:
    exec:
      command: ["/bin/sh", "-c", "sleep 5"]
```

This 5-second sleep gives the load balancer time to route new requests away from this pod before the server stops accepting connections.

## `terminationGracePeriodSeconds`

Set in the Deployment spec — Kubernetes waits this long after SIGTERM before sending SIGKILL:

```yaml
spec:
  terminationGracePeriodSeconds: 60
```

60 seconds = 5s `preStop` + 30s request drain + 25s buffer for DB/Redis/worker shutdown.

## Temporal Worker Drain

```go
func (w *ContractsWorker) Stop() {
    w.worker.Stop()   // signals worker to stop polling
    // Temporal SDK waits for in-progress activities to complete
    // or until the worker's stop timeout (default: 5s)
}
```

For long-running activities with heartbeat: set `HeartbeatTimeout` and the worker will stop mid-activity gracefully, Temporal reschedules on another worker.

## DB Pool Cleanup

Wire generates cleanup functions for providers that return `(T, func(), error)`:

```go
func NewPool(cfg Config) (*pgxpool.Pool, func(), error) {
    pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
    if err != nil {
        return nil, nil, err
    }
    cleanup := func() {
        pool.Close()   // waits for all connections to return
    }
    return pool, cleanup, nil
}
```

The generated `cleanup()` function in `wire_gen.go` calls all provider cleanups in reverse construction order — pool closes last, after all consumers have stopped.

## Graceful Shutdown Checklist

- [ ] `ShutdownWithContext` called with 30s timeout on Fiber
- [ ] All Temporal workers stopped via `worker.Stop()`
- [ ] pgxpool closed via Wire cleanup
- [ ] Redis pool closed via Wire cleanup
- [ ] OTel exporter flushed (`ForceFlush` before process exit)
- [ ] `terminationGracePeriodSeconds` > drain timeout + buffer
- [ ] `preStop` sleep added to Kubernetes Deployment spec
