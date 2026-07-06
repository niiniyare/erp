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

	"awo.so/awo/api/authz"
	"awo.so/awo/api/middleware"
	"awo.so/awo/api/openapi"
	"awo.so/awo/api/router"
	"awo.so/awo/bootstrap"
	contrib "awo.so/awo/contrib/pgx"
	contribredis "awo.so/awo/contrib/redis"
	"awo.so/awo/events/outbox"
	"awo.so/awo/observability/health"
	"awo.so/awo/observability/metrics"
	"awo.so/awo/platform/iam"

	// Platform module init() calls — imports drive entity registration.
	// platform/iam is already imported above for iam.NewService; the init()
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

	// Build shared infrastructure clients.
	redisClient := contribredis.New(result.Redis)

	// IAM service — requires the iam_user/iam_session entity schemas.
	iamSchema, ok := result.Schema.ByName["iam_user"]
	if !ok {
		slog.Error("iam_user entity not found in schema — platform/iam not registered")
		os.Exit(1)
	}
	iamSessionSchema, ok := result.Schema.ByName["iam_session"]
	if !ok {
		slog.Error("iam_session entity not found in schema — platform/iam not registered")
		os.Exit(1)
	}
	userRepo := contrib.NewRepository(result.Pool, iamSchema)
	sessionRepo := contrib.NewRepository(result.Pool, iamSessionSchema)
	iamSvc := iam.NewService(userRepo, sessionRepo, redisClient)

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
	iam.RegisterAuthRoutes(app, iamSvc)

	// OpenAPI schema endpoint (no auth required).
	app.Get("/api/openapi.json", func(c *fiber.Ctx) error {
		doc := openapi.Generate(result.Schema, "")
		return c.JSON(doc)
	})

	// Build RBAC enforcer from compiled Casbin policies.
	enforcer, err := authz.NewEnforcer(result.Schema)
	if err != nil {
		slog.Error("casbin enforcer init failed", "err", err)
		os.Exit(1)
	}

	// CRUD routes for all registered entities — full middleware pipeline applied inside.
	router.Register(app, result.Schema, router.RegisterOptions{
		Pool:     result.Pool,
		Redis:    result.Redis,
		IAM:      iamSvc,
		Tenants:  tenantRepo,
		Authz:    enforcer,
		Temporal: nil, // TODO: wire Temporal client when worker is configured
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
