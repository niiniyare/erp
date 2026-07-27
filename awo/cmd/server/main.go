// Command server starts the Awo API server and Temporal worker.
//
// The server starts in two phases:
//  1. Bootstrap: PostgreSQL + Redis + entity registry + schema compilation.
//  2. Serve: Fiber HTTP server starts; Temporal worker starts concurrently.
//
// Configuration is loaded from environment variables. All required variables
// must be set; the process exits with a clear error message if any are missing.
//
// Environment variables:
//
//	DATABASE_URL   — PostgreSQL DSN (required)
//	REDIS_URL      — Redis URL (required)
//	PORT           — HTTP listen port (default: 8080)
//	TEMPORAL_HOST  — Temporal server host:port (default: localhost:7233)
//	LOG_LEVEL      — debug|info|warn|error (default: info)
//	APP_NAME       — service name for observability labels (default: awo)
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"awo.so/awo/api/middleware"
	"awo.so/awo/api/openapi"
	"awo.so/awo/api/router"
	"awo.so/awo/audit"
	"awo.so/awo/auth"
	"awo.so/awo/bootstrap"
	"awo.so/awo/cache"
	contrib "awo.so/awo/contrib/pgx"
	contribredis "awo.so/awo/contrib/redis"
	"awo.so/awo/events/outbox"
	"awo.so/awo/observability/health"
	"awo.so/awo/observability/metrics"
	"awo.so/awo/platform/iam"
	"awo.so/awo/sdui"

	// Platform module init() calls — imports drive entity registration.
	// platform/iam is already imported above for iam.New; the init()
	// side effect (entity registration) is included via that import.
	_ "awo.so/awo/platform/audit"
	_ "awo.so/awo/platform/flags"
	_ "awo.so/awo/platform/metadata"
	_ "awo.so/awo/platform/registry"
	_ "awo.so/awo/platform/settings"
	_ "awo.so/awo/platform/tenant"
)

func main() {
	setupLogging()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Bootstrap infrastructure.
	result, err := bootstrap.Run(ctx, bootstrap.Config{
		DatabaseURL: mustEnv("DATABASE_URL"),
		RedisURL:    mustEnv("REDIS_URL"),
		AppName:     getEnv("APP_NAME", "awo"),
	})
	if err != nil {
		slog.Error("bootstrap failed", "err", err)
		os.Exit(1)
	}
	defer bootstrap.Shutdown(context.Background(), result)

	// Audit writer — writes AuditRecord to platform_audit_log inside the entity
	// transaction. contrib.NewPoolQuerier provides the fallback connection for
	// standalone auth writes that occur outside an entity transaction.
	//
	// Sanitizer strips sensitive fields before any INSERT; RiskScorer computes
	// the 0-100 risk score and derives the Severity for each record.
	auditSanitizer := audit.NewSanitizer()
	// SensitiveEntityLoader queries audit_sensitive_entities (Phase 2 migration).
	// When the table does not yet exist, the query fails and the loader returns
	// nil, nil — the scorer falls back to default rules (no risk premium).
	pool := result.Pool
	auditScorer := audit.NewRiskScorer().WithLoader(func(ctx context.Context) ([]string, error) {
		rows, err := pool.Query(ctx, "SELECT entity_name FROM audit_sensitive_entities")
		if err != nil {
			// Table may not exist pre-Phase-2 migration; treat as empty.
			slog.WarnContext(ctx, "audit: sensitive entity table not available; using defaults", "err", err)
			return nil, nil
		}
		defer rows.Close()
		var entities []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return nil, err
			}
			entities = append(entities, name)
		}
		return entities, rows.Err()
	})
	if err := auditScorer.Warm(ctx); err != nil {
		slog.Warn("audit scorer warm failed; using default scoring rules", "err", err)
	}
	// FlagLoader reads the feature flag from platform_audit_config.
	// When the table does not exist (pre-migration), the query fails and the
	// loader returns false — legacy audit system remains active (ADR-019 Phase 2).
	auditFlagLoader := audit.FlagLoader(func(ctx context.Context) bool {
		var value string
		err := pool.QueryRow(ctx,
			"SELECT value FROM platform_audit_config WHERE key = 'feature.unified_audit.enabled'",
		).Scan(&value)
		if err != nil {
			slog.Warn("audit: feature flag read failed; unified audit disabled", "err", err)
			return false
		}
		return value == "true"
	})

	auditWriter := audit.NewTransactionalWriter(contrib.NewPoolQuerier(result.Pool), auditSanitizer, auditScorer).
		WithFlagLoader(auditFlagLoader)

	// IAM module — authenticates users and manages sessions.
	// Construct the SessionStore and token cache from the shared Redis client,
	// then inject them into iam.New as abstract interfaces.
	sessions := contribredis.NewSessionStore(result.Redis)
	tokenCache := contribredis.New(result.Redis)
	iamModule := iam.New(result.Pool, sessions, tokenCache).WithAuditWriter(auditWriter)

	// Tenant entity repository for TenantResolver middleware.
	tenantSchema, ok := result.Schema.ByName["platform_tenant"]
	if !ok {
		slog.Error("platform_tenant entity not found in schema — platform/tenant not registered")
		os.Exit(1)
	}
	tenantRepo := contrib.NewRepository(result.Pool, tenantSchema)

	// Outbox relay — delivers domain events from the transactional outbox.
	relay := outbox.New(result.Pool)
	go func() {
		if err := relay.Start(ctx); err != nil && ctx.Err() == nil {
			slog.Error("outbox relay stopped unexpectedly", "err", err)
		}
	}()

	// Build Fiber app.
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
	})

	// Global middleware pipeline — request ID, logger, recovery, CORS.
	router.RegisterMiddleware(app, middleware.CORSConfig{
		BaseDomain:     getEnv("BASE_DOMAIN", ""),
		AllowedOrigins: splitComma(getEnv("CORS_ALLOWED_ORIGINS", "")),
	})

	// Metrics middleware.
	app.Use(metrics.Middleware())

	// Health checks (no auth required).
	checker := health.New(result.Pool, result.Redis)
	app.Get("/health/live", checker.Live)
	app.Get("/health/ready", checker.Ready)
	app.Get("/metrics", metrics.Handler())

	// Auth routes (login/logout/me — no upstream RequireAuth middleware).
	iamModule.RegisterRoutes(app)

	// OpenAPI schema endpoint (no auth required).
	app.Get("/api/openapi.json", func(c *fiber.Ctx) error {
		doc := openapi.Generate(result.Schema, "")
		return c.JSON(doc)
	})

	// Build RBAC evaluator from compiled capability grants + IAM role-permission bindings.
	rolePerms, err := iamModule.Auth.LoadRolePermissions(ctx)
	if err != nil {
		slog.Error("load role permissions failed", "err", err)
		os.Exit(1)
	}
	evaluator, err := auth.NewCasbinEvaluator(result.Schema.CapabilityGrants, rolePerms)
	if err != nil {
		slog.Error("casbin evaluator init failed", "err", err)
		os.Exit(1)
	}

	// SDUI generator — produces amis JSON page schemas from EntitySchema.
	// Cache backed by Redis when available; nil disables caching (dev mode).
	var sduiCache cache.Cache
	if result.Redis != nil {
		sduiCache = contribredis.New(result.Redis)
	}
	sduiGen := sdui.New(result.Schema, evaluator, sduiCache)

	// CRUD routes for all registered entities — full middleware pipeline applied inside.
	router.Register(app, result.Schema, router.RegisterOptions{
		Pool:               result.Pool,
		Redis:              result.Redis,
		IAM:                iamModule.Auth,
		Tenants:            tenantRepo,
		Authz:              evaluator,
		Temporal:           nil, // TODO: wire Temporal client when worker is configured
		AuditWriter:        auditWriter,
		AuditSigningSecret: getEnv("AUDIT_SIGNING_SECRET", ""),
		SDUIGenerator:      sduiGen,
	})

	// Static file serving — amis SDK assets.
	// Serve /static/* from awo/web/static/ relative to the binary working dir.
	app.Static("/static", "./awo/web/static", fiber.Static{
		Compress: true,
		MaxAge:   86400, // 1 day — SDK files are pinned and content-stable
	})

	// Web UI fallback — serve index.html for all /ui/* paths (SPA routing).
	app.Get("/ui/*", func(c *fiber.Ctx) error {
		return c.SendFile("./awo/web/pages/index.html")
	})

	// Start server.
	port := getEnv("PORT", "8080")
	slog.Info("server starting", "port", port, "app", getEnv("APP_NAME", "awo"))

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- app.Listen(":" + port)
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			slog.Error("graceful shutdown failed", "err", err)
		}
	case err := <-serverErr:
		if err != nil {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}
}

func setupLogging() {
	level := slog.LevelInfo
	if l := os.Getenv("LOG_LEVEL"); l != "" {
		_ = level.UnmarshalText([]byte(l))
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("required environment variable is not set", "var", key)
		os.Exit(1)
	}
	return v
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func splitComma(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
