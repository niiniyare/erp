// Package main is the entry point for the Awo ERP API server (framework edition).
//
// Configuration via environment variables:
//
//	DATABASE_URL   — PostgreSQL DSN (required)
//	REDIS_URL      — Redis URL, e.g. redis://localhost:6379 (required for IAM)
//	PORT           — HTTP listen port (default: 8080)
//	HOST           — HTTP listen host (default: 0.0.0.0)
//	LOG_LEVEL      — trace|debug|info|warn|error (default: info)
//	LOG_FORMAT     — json|pretty (default: json)
//	ALLOWED_ORIGINS— CORS origins, comma-separated (default: *)
//	TEMPORAL_HOST  — Temporal server host:port (optional; disables workflows if absent)
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tclient "go.temporal.io/sdk/client"

	// Register all platform EntityDefinitions via init():
	//   tenant_user, role, permission, user_role_assignment
	iamPkg "awo.so/framework/platform/iam"

	"awo.so/framework/api"
	"awo.so/framework/bootstrap"
)

func main() {
	setupLogger()

	pool, err := buildPool()
	if err != nil {
		log.Fatal().Err(err).Msg("database connect failed")
	}
	defer pool.Close()

	redisClient, err := buildRedis()
	if err != nil {
		log.Fatal().Err(err).Msg("redis connect failed")
	}

	var temporalClient tclient.Client
	if host := env("TEMPORAL_HOST", ""); host != "" {
		c, err := tclient.Dial(tclient.Options{HostPort: host})
		if err != nil {
			log.Warn().Err(err).Msg("temporal unavailable — workflow triggers disabled")
		} else {
			temporalClient = c
			defer temporalClient.Close()
		}
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: jsonErrorHandler,
	})

	app.Use(recover.New())
	allowOrigins := env("ALLOWED_ORIGINS", "")
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowCredentials: allowOrigins != "" && allowOrigins != "*",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type,Authorization,X-Awo-Tenant",
	}))
	app.Use(tenantContextMiddleware())

	// Mount IAM auth routes first so they take precedence over entity wildcard routes.
	// bootstrap.Mount is called with ViewerFn set manually to avoid double-mounting IAM.
	iamSvc := iamPkg.NewAuthService(pool, redisClient)
	iamPkg.NewHandler(iamSvc).Mount(app)
	app.Use(iamPkg.AuthMiddleware(redisClient))

	// Wire entity CRUD routes, SDUI, and workflow triggers.
	bootstrap.Mount(app, bootstrap.Options{
		Pool:           pool,
		RedisClient:    nil, // IAM already mounted above; pass nil to skip duplicate mount.
		APIPrefix:      "/api",
		ViewerFn:       iamPkg.ViewerFromCtx,
		TenantResolver: slugTenantResolver(pool),
		TemporalClient: temporalClient,
	})

	// Platform admin endpoint — list all tenants (no RLS; platform-plane token required).
	app.Get("/api/tenants", func(c *fiber.Ctx) error {
		type row struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Slug      string `json:"slug"`
			Status    string `json:"status"`
			CreatedAt string `json:"created_at"`
		}
		rows, err := pool.Query(c.Context(),
			`SELECT id, name, slug, status, created_at FROM tenants ORDER BY created_at DESC`)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": 1, "msg": err.Error()})
		}
		defer rows.Close()
		var out []row
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.ID, &r.Name, &r.Slug, &r.Status, &r.CreatedAt); err == nil {
				out = append(out, r)
			}
		}
		if out == nil {
			out = []row{}
		}
		return c.JSON(fiber.Map{"status": 0, "data": out})
	})

	if webDir := env("WEB_DIR", ""); webDir != "" {
		app.Static("/sdk", webDir+"/sdk")
		app.Static("/schemas", webDir+"/schemas")
		app.Static("/pages", webDir+"/pages")
		app.Get("/", func(c *fiber.Ctx) error { return c.SendFile(webDir + "/shell.html") })
	} else {
		app.Get("/", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"app": "awo-erp", "status": "ok"})
		})
	}

	app.Get("/health/live", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Get("/health/ready", func(c *fiber.Ctx) error {
		if err := pool.Ping(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(
				fiber.Map{"status": "unhealthy", "db": err.Error()})
		}
		if err := redisClient.Ping(c.Context()).Err(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(
				fiber.Map{"status": "unhealthy", "redis": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	addr := env("HOST", "0.0.0.0") + ":" + env("PORT", "8080")
	go func() {
		log.Info().Str("addr", addr).Msg("server listening")
		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-quit
	log.Info().Msg("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}
}

func buildPool() (*pgxpool.Pool, error) {
	dsn := env("DATABASE_URL", "")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return pool, nil
}

func buildRedis() (*redis.Client, error) {
	rawURL := env("REDIS_URL", "redis://localhost:6379")
	opt, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse REDIS_URL: %w", err)
	}
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}

func setupLogger() {
	level, err := zerolog.ParseLevel(env("LOG_LEVEL", "info"))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	if env("LOG_FORMAT", "json") == "pretty" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}

func tenantContextMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if t := c.Get("X-Awo-Tenant"); t != "" {
			c.Locals("tenant_slug", t)
		}
		return c.Next()
	}
}

func slugTenantResolver(pool *pgxpool.Pool) api.TenantResolver {
	return func(ctx context.Context, slugOrID string) (uuid.UUID, error) {
		if id, err := uuid.Parse(slugOrID); err == nil {
			return id, nil
		}
		var id uuid.UUID
		if err := pool.QueryRow(ctx,
			"SELECT id FROM tenants WHERE slug = $1 LIMIT 1", slugOrID,
		).Scan(&id); err != nil {
			return uuid.Nil, fmt.Errorf("unknown tenant %q", slugOrID)
		}
		return id, nil
	}
}

func jsonErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	msg := "internal server error"
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		msg = e.Message
	}
	log.Error().Err(err).Int("status", code).Str("path", c.Path()).Msg("request error")
	return c.Status(code).JSON(fiber.Map{"status": code, "msg": msg})
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
