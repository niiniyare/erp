---
title: "Chapter 13: Middleware Pipeline"
part: "Part III — The API Layer"
chapter: 13
section: "13-middleware-pipeline"
related:
  - "[Chapter 12: Server Setup](12-server-setup.md)"
  - "[Chapter 14: Multi-Tenancy Middleware](14-multitenancy-middleware.md)"
  - "[Chapter 15: Authentication and Session Management](15-auth-sessions.md)"
---

# Chapter 13: Middleware Pipeline

Every Awo HTTP request passes through an ordered middleware pipeline before reaching a route handler. The order is not arbitrary — each middleware depends on data set by the preceding one. Getting the order wrong produces subtle bugs: logs missing request IDs, rate limits applied to the wrong IP, auth bypassed by panics.

---

## 13.1. Middleware Execution Order

### 13.1.1. Why Order Matters

Consider: if panic recovery middleware runs before request ID middleware, a panicking handler logs a stack trace without a request ID — impossible to correlate with the user's reported error. If rate limiting runs before IP trust is established, a user behind NAT is rate-limited as a single entity.

### 13.1.2. The Canonical Order

```go
func SetupMiddleware(app *fiber.App, cfg *config.Config, deps *Dependencies) {
    // 1. Panic recovery — must be first so all errors produce structured responses
    app.Use(middleware.Recovery(deps.Logger))

    // 2. Request ID — set before logging so log entries have correlation IDs
    app.Use(middleware.RequestID())

    // 3. Structured request logging — after RequestID, before auth
    app.Use(middleware.Logger(deps.Logger))

    // 4. CORS — before auth so preflight OPTIONS requests resolve without a session
    app.Use(middleware.CORS(cfg.AllowedOrigins))

    // 5. Tenant resolution — before auth (auth is tenant-scoped)
    app.Use(middleware.Tenant(deps.TenantService))

    // 6. Session validation — after tenant (session is validated in tenant context)
    app.Use(middleware.Session(deps.SessionService))

    // 7. Rate limiting — after identity is established
    app.Use(middleware.RateLimit(deps.Redis, cfg.RateLimit))

    // 8. Feature flags — after tenant, before route handlers
    app.Use(middleware.FeatureFlags(deps.FeatureFlagService))
}
```

---

## 13.2. Request ID Middleware

### 13.2.1. Generating a UUID Request ID

Every request receives a UUID v4 request ID:

```go
func RequestID() fiber.Handler {
    return func(c *fiber.Ctx) error {
        id := c.Get("X-Request-ID")
        if id == "" {
            id = uuid.New().String()
        }
        c.Locals("request_id", id)
        c.Set("X-Request-ID", id)
        return c.Next()
    }
}
```

The request ID is:
- Set as `X-Request-ID` on the response (so callers can correlate errors)
- Stored in `c.Locals("request_id")` for downstream middleware
- Injected into the logger context via the logging middleware

### 13.2.2. Honouring Incoming `X-Request-ID` — Trust Rules

Accept `X-Request-ID` from callers only when it arrives from a trusted internal service (identified by IP CIDR match). For external requests, always generate a new ID — a malicious caller should not be able to inject an ID that collides with another user's request in your logs.

For the Awo SDUI client (running in the browser), the browser generates a UUID per page load and includes it in all API requests. This is an exception to the above: the browser is not trusted, but accepting its request ID improves traceability for support teams.

### 13.2.3. Propagating to All Downstream Logs and Traces

The request ID must appear in every log line for a request. Inject it into the slog context:

```go
func Logger(logger *slog.Logger) fiber.Handler {
    return func(c *fiber.Ctx) error {
        reqID := c.Locals("request_id").(string)
        // Attach to every log from this request
        ctx := context.WithValue(c.UserContext(), logContextKey{}, logger.With("request_id", reqID))
        c.SetUserContext(ctx)

        start := time.Now()
        err := c.Next()
        duration := time.Since(start)

        log := logFromContext(c.UserContext())
        log.Info("request",
            "method", c.Method(),
            "path", c.Path(),
            "status", c.Response().StatusCode(),
            "duration_ms", duration.Milliseconds(),
            "tenant_id", c.Locals("tenant_id"),
            "actor_id", c.Locals("actor_id"),
            "ip", c.IP(),
        )
        return err
    }
}
```

---

## 13.3. Structured Logging Middleware

### 13.3.1. `slog` Integration

Awo uses Go's standard `log/slog` with a JSON handler for production. Every request log line is a single JSON object:

```json
{
  "time": "2025-07-04T12:00:00Z",
  "level": "INFO",
  "msg": "request",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "POST",
  "path": "/api/v1/invoices",
  "status": 201,
  "duration_ms": 47,
  "tenant_id": "a1b2c3d4-...",
  "actor_id": "e5f6g7h8-...",
  "ip": "41.206.3.1"
}
```

### 13.3.2. What Is Always Logged

- HTTP method and path (without query string — query strings may contain sensitive filter values)
- Response status code
- Request duration in milliseconds
- Request ID
- Tenant ID (from middleware)
- Actor ID (from session, if authenticated)
- Client IP address

### 13.3.3. What Is Never Logged

- Request or response bodies (may contain PII, credentials, sensitive business data)
- `Sensitive` field values (see Chapter 5)
- Session tokens or API keys
- Passwords in any form
- Credit card numbers, bank account numbers

The logging middleware has an explicit blocklist of header names that are scrubbed: `Authorization`, `Cookie`, `X-API-Key`.

### 13.3.4. Slow Request Threshold

Requests exceeding 500ms are logged at `WARN` level with additional fields:

```go
if duration > 500*time.Millisecond {
    log.Warn("slow request",
        "duration_ms", duration.Milliseconds(),
        "path", c.Path(),
        "tenant_id", c.Locals("tenant_id"),
        "query", c.Request().URI().QueryString(),  // safe for slow-request diagnostics
    )
}
```

---

## 13.4. Panic Recovery Middleware

### 13.4.1. Converting Panics to 500 Responses

```go
func Recovery(logger *slog.Logger) fiber.Handler {
    return func(c *fiber.Ctx) error {
        defer func() {
            if r := recover(); r != nil {
                err, ok := r.(error)
                if !ok {
                    err = fmt.Errorf("%v", r)
                }
                logger.Error("panic recovered",
                    "error", err,
                    "stack", string(debug.Stack()),
                    "request_id", c.Locals("request_id"),
                    "path", c.Path(),
                    "method", c.Method(),
                )
                _ = c.Status(500).JSON(fiber.Map{
                    "status":  500,
                    "message": "an internal error occurred",
                    "request_id": c.Locals("request_id"),
                })
            }
        }()
        return c.Next()
    }
}
```

### 13.4.2. Stack Trace in Logs, Not in Response

The stack trace is logged server-side with the request ID. The response contains only the request ID — enough for support to correlate with server logs, not enough to expose implementation details.

### 13.4.3. Panic Rate Alerting

Monitor `count(level=ERROR, msg="panic recovered")` in your metrics. More than 1 panic per minute triggers an alert. Panics indicate bugs (nil pointer dereferences, type assertion failures) that should be fixed, not silently recovered.

---

## 13.5. CORS Middleware

### 13.5.1. Allowed Origins

```go
func CORS(allowedOrigins []string) fiber.Handler {
    return cors.New(cors.Config{
        AllowOrigins:     strings.Join(allowedOrigins, ","),
        AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
        AllowHeaders:     "Content-Type,Authorization,X-Request-ID,X-Tenant-ID",
        ExposeHeaders:    "X-Request-ID,X-RateLimit-Remaining",
        AllowCredentials: true,
        MaxAge:           86400,
    })
}
```

Allowed origins come from `AWO_CORS_ORIGINS` config, typically:
- The CDN domain for the amis frontend: `https://app.awo.so`
- Tenant subdomains: `https://*.awo.app`

### 13.5.2. Why `AllowCredentials` Requires Explicit Origin

When `AllowCredentials: true`, `AllowOrigins` cannot be `*`. The browser enforces this. A wildcard origin with credentials would allow any malicious website to make authenticated requests to your API with the user's session cookie — a CSRF attack.

Always enumerate specific origins. Use a regex or wildcard pattern for subdomain matching where supported.

---

## 13.6. Rate Limiting Middleware

### 13.6.1. Per-Tenant Rate Limiting — Redis Sliding Window

```go
type RateLimiter struct {
    redis    rueidis.Client
    limit    int           // requests per window
    window   time.Duration // window size
}

func (rl *RateLimiter) Check(ctx context.Context, key string) (remaining int, err error) {
    now := time.Now().UnixMilli()
    windowStart := now - rl.window.Milliseconds()

    // Sliding window using Redis sorted sets
    pipe := rl.redis.Pipeline()
    pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(windowStart, 10))
    pipe.ZAdd(ctx, key, rueidis.ZAddArgs{Members: []rueidis.Z{{Score: float64(now), Member: strconv.FormatInt(now, 10)}}})
    pipe.ZCard(ctx, key)
    pipe.Expire(ctx, key, rl.window)
    results, err := pipe.Exec(ctx)
    // ...
    count := results[2].AsInt64()
    return rl.limit - int(count), nil
}
```

Rate limit keys:
- Per-tenant: `rl:tenant:{tenant_id}` — 1000 req/min default
- Per-user: `rl:user:{user_id}` — 100 req/min default
- Per-IP (unauthenticated): `rl:ip:{ip}` — 30 req/min

### 13.6.2. Rate Limit Headers

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 856
X-RateLimit-Reset: 1720000060
Retry-After: 30   (only when rate limited — 429 response)
```

### 13.6.3. Rate Limit Bypass for Internal Calls

Service-to-service calls (Temporal activities calling the API, internal cron jobs) carry an `X-Internal-Token` header. When this header matches `AWO_INTERNAL_TOKEN`, rate limiting is bypassed entirely.

Never document `AWO_INTERNAL_TOKEN` publicly. Rotate it on security incidents.
