---
title: "Chapter 12: Server Setup"
part: "Part III — The API Layer"
chapter: 12
section: "12-server-setup"
related:
  - "[Chapter 13: Middleware Pipeline](13-middleware-pipeline.md)"
  - "[Chapter 14: Multi-Tenancy Middleware](14-multitenancy-middleware.md)"
---

# Chapter 12: Server Setup

The Awo server is a Fiber v2 application bootstrapped via Google Wire dependency injection. Understanding the startup sequence, configuration model, and health endpoint design is essential for both deployment and debugging.

---

## 12.1. Entry Point and Process Structure

### 12.1.1. `main.go` — Bootstrap Sequence

```go
func main() {
    // 1. Load and validate config from environment
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("config error: %v", err)
    }

    // 2. Initialize structured logger
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: cfg.LogLevel,
    }))
    slog.SetDefault(logger)

    // 3. Wire dependencies (database, Redis, Temporal, repositories, services)
    app, cleanup, err := wire.InitializeApp(cfg)
    if err != nil {
        log.Fatalf("wire initialization failed: %v", err)
    }
    defer cleanup()

    // 4. Start the Fiber HTTP server in a goroutine
    go func() {
        addr := fmt.Sprintf(":%d", cfg.Port)
        if err := app.Fiber.Listen(addr); err != nil {
            slog.Error("server stopped", "error", err)
        }
    }()

    // 5. Block until OS signal, then graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    slog.Info("shutting down server")

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := app.Fiber.ShutdownWithContext(ctx); err != nil {
        slog.Error("shutdown error", "error", err)
    }
}
```

### 12.1.2. Dependency Injection — What Is Wired at Startup

Wire generates `wire_gen.go` from provider functions declared in `internal/platform/wire/`. The full dependency graph includes:

| Layer | Components |
|---|---|
| Infrastructure | pgxpool, Redis (rueidis), Temporal client |
| Data | ent client, all entity repositories |
| Services | All core services (IAM, tenant, feature flags, finance, etc.) |
| API | Fiber app, all route handlers, middleware instances |
| Workers | Temporal worker registrations |

Wire ensures every dependency is created exactly once, in the correct order, with all transitive dependencies satisfied.

### 12.1.3. Graceful Shutdown — Signal Handling, Drain Timeout

On `SIGINT` or `SIGTERM`:
1. Fiber stops accepting new connections
2. In-flight requests are given 30 seconds to complete
3. Temporal worker stops polling for new tasks
4. Database connection pool drains (in-flight queries complete)
5. Redis connection closes
6. Process exits 0

The 30-second drain timeout matches Kubernetes' default `terminationGracePeriodSeconds`. Requests that don't complete within 30 seconds receive a 503 and the connection is forcefully closed.

---

## 12.2. Configuration Loading

### 12.2.1. Environment Variables — The `AWO_*` Namespace

All Awo configuration is loaded from environment variables with the `AWO_` prefix:

```bash
AWO_PORT=8080
AWO_DATABASE_URL=postgres://user:pass@localhost/awo
AWO_REDIS_URL=redis://localhost:6379
AWO_TEMPORAL_HOST=localhost:7233
AWO_TEMPORAL_NAMESPACE=awo
AWO_SESSION_SECRET=32-byte-random-hex
AWO_LOG_LEVEL=info
AWO_ENV=production        # development | staging | production
AWO_TRUSTED_PROXIES=10.0.0.0/8
```

### 12.2.2. Config Struct — Typed, Validated at Startup

```go
type Config struct {
    Port            int           `env:"AWO_PORT,required"`
    DatabaseURL     string        `env:"AWO_DATABASE_URL,required"`
    RedisURL        string        `env:"AWO_REDIS_URL,required"`
    TemporalHost    string        `env:"AWO_TEMPORAL_HOST,required"`
    TemporalNS      string        `env:"AWO_TEMPORAL_NAMESPACE" envDefault:"awo"`
    SessionSecret   string        `env:"AWO_SESSION_SECRET,required"`
    LogLevel        slog.Level    `env:"AWO_LOG_LEVEL" envDefault:"info"`
    Env             string        `env:"AWO_ENV" envDefault:"development"`
    TrustedProxies  []string      `env:"AWO_TRUSTED_PROXIES" envSeparator:","`
    MaxConns        int           `env:"AWO_DB_MAX_CONNS" envDefault:"20"`
    SessionTTL      time.Duration `env:"AWO_SESSION_TTL" envDefault:"8h"`
}
```

Configuration is validated during `config.Load()` — missing required values fail fast before any dependency is created. This is the correct behaviour: a server that boots with an invalid config and silently falls back to defaults creates subtle production bugs.

### 12.2.3. Environment-Specific Overrides

```bash
# development: relaxed TLS, verbose logging, shorter session TTL
AWO_ENV=development
AWO_LOG_LEVEL=debug
AWO_SESSION_TTL=24h
AWO_TRUSTED_PROXIES=127.0.0.1

# production: strict settings enforced in config.Validate()
AWO_ENV=production
AWO_LOG_LEVEL=warn
AWO_SESSION_TTL=8h
```

`config.Validate()` enforces production-specific rules:
- `AWO_SESSION_SECRET` must be at least 32 bytes
- `AWO_DATABASE_URL` must use SSL (`sslmode=require`)
- `AWO_ENV=development` cannot be set in a production build (detected via build tag)

### 12.2.4. Secrets — Never in Config Files

Secrets (`AWO_SESSION_SECRET`, `AWO_DATABASE_URL` password component, API keys) are injected via environment variables from a secret manager (HashiCorp Vault, AWS Secrets Manager, Kubernetes Secrets). They are never:
- Committed to version control
- Logged (the config struct's `String()` method redacts secret fields)
- Included in error messages

---

## 12.3. TLS and Reverse Proxy

### 12.3.1. Caddy as the Recommended Reverse Proxy

Caddy handles TLS termination automatically via Let's Encrypt ACME. Awo runs on HTTP internally; Caddy proxies HTTPS → HTTP:

```caddyfile
{tenant}.awo.app {
    reverse_proxy localhost:8080 {
        header_up X-Forwarded-For {remote_host}
        header_up X-Real-IP {remote_host}
    }
}
```

### 12.3.2. X-Forwarded-For and Real IP Extraction

Fiber extracts the real client IP from `X-Forwarded-For` when trusted proxies are configured:

```go
app.Fiber.Use(func(c *fiber.Ctx) error {
    // Only trust X-Forwarded-For from known proxy IP ranges
    c.IP() // returns real IP after proxy configuration
    return c.Next()
})
```

### 12.3.3. Trusted Proxy Configuration — Security Implications

If `AWO_TRUSTED_PROXIES` is too broad (e.g. `0.0.0.0/0`), any client can spoof their IP by setting `X-Forwarded-For`. This defeats:
- Rate limiting (spoofed IP bypasses per-IP limits)
- Login attempt tracking
- Audit log IP addresses

Always set `AWO_TRUSTED_PROXIES` to only the known CIDR range of your load balancer/proxy fleet.

---

## 12.4. Health and Readiness Endpoints

### 12.4.1. `GET /health/live` — Process Is Alive

Returns 200 if the process is running. No dependencies checked. Used by Kubernetes liveness probe to detect process crashes.

```json
{ "status": "ok", "timestamp": "2025-07-04T12:00:00Z" }
```

### 12.4.2. `GET /health/ready` — Process Is Ready to Serve Traffic

Returns 200 only if all critical dependencies are reachable. Used by Kubernetes readiness probe to route traffic only to healthy instances.

```go
func healthReady(c *fiber.Ctx) error {
    checks := map[string]error{
        "database": db.Ping(c.Context()),
        "redis":    redis.Ping(c.Context()),
    }

    failed := map[string]string{}
    for name, err := range checks {
        if err != nil {
            failed[name] = err.Error()
        }
    }

    if len(failed) > 0 {
        return c.Status(503).JSON(fiber.Map{
            "status": "degraded",
            "failed": failed,
        })
    }
    return c.JSON(fiber.Map{"status": "ready"})
}
```

### 12.4.3. Deep Health Check — Database, Redis, Temporal

`GET /health/deep` (authenticated, for monitoring systems):

```json
{
  "status": "ok",
  "checks": {
    "database": { "status": "ok", "latency_ms": 2 },
    "redis":    { "status": "ok", "latency_ms": 1 },
    "temporal": { "status": "ok", "latency_ms": 15 },
    "migrations": { "status": "ok", "pending": 0 }
  }
}
```

### 12.4.4. Health Check Authentication

- `/health/live` — unauthenticated (Kubernetes probes cannot authenticate)
- `/health/ready` — unauthenticated (same reason)
- `/health/deep` — requires `Bearer {internal_token}` — not exposed to the public internet

Never expose the deep health check without authentication. It reveals internal infrastructure topology to anyone who queries it.
