// Package main is the entry point for the AWO ERP API server.
package frameworkserver

// package main

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"
//
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/gofiber/fiber/v2/middleware/cors"
// 	"github.com/gofiber/fiber/v2/middleware/recover"
// 	"github.com/google/uuid"
// 	"github.com/jackc/pgx/v5/pgxpool"
// 	"github.com/rs/zerolog"
// 	"github.com/rs/zerolog/log"
// 	tclient "go.temporal.io/sdk/client"
//
// 	// Register all built-in EntityDefinitions via init().
// 	_ "awo.so/internal/modules"
//
// 	"awo.so/framework/api"
// 	"awo.so/framework/bootstrap"
// 	"awo.so/internal/framework"
// 	"awo.so/internal/platform/config"
// 	"awo.so/internal/platform/temporal"
// 	"awo.so/internal/shared/logger"
// )
//
// func main() {
// 	cfg := config.Load()
// 	setupLogger(cfg)
// 	log.Info().Str("port", cfg.Server.Port).Msg("starting AWO ERP server")
//
// 	pool, err := buildPool(cfg)
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("failed to connect to database")
// 	}
// 	defer pool.Close()
//
// 	var temporalClient tclient.Client
// 	if tm, err := temporal.NewClientManager(&cfg.Temporal, logger.NewNoOp()); err != nil {
// 		log.Warn().Err(err).Msg("temporal unavailable — workflow triggers disabled")
// 	} else {
// 		temporalClient = tm.GetClient()
// 		defer temporalClient.Close()
// 	}
//
// 	app := fiber.New(fiber.Config{
// 		ReadTimeout:  cfg.Server.ReadTimeout,
// 		WriteTimeout: cfg.Server.WriteTimeout,
// 		IdleTimeout:  cfg.Server.IdleTimeout,
// 		ErrorHandler: jsonErrorHandler,
// 	})
//
// 	app.Use(recover.New())
// 	app.Use(cors.New(cors.Config{
// 		AllowOrigins:     cfg.Server.AllowedOrigins,
// 		AllowCredentials: true,
// 		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
// 		AllowHeaders:     "Content-Type,Authorization,X-Awo-Tenant",
// 	}))
// 	app.Use(tenantContextMiddleware())
//
// 	// Mount the entire Awo Framework (CRUD + SDUI routes) in one call.
// 	bootstrap.Mount(app, bootstrap.Options{
// 		Pool:           pool,
// 		APIPrefix:      "/api",
// 		ViewerFn:       framework.ViewerFromFiber(),
// 		TenantResolver: slugTenantResolver(pool),
// 		TemporalClient: temporalClient,
// 	})
//
// 	// Static shell UI
// 	app.Static("/sdk", "./framework/web/sdk")
// 	app.Static("/schemas", "./web/schemas")
// 	app.Get("/", func(c *fiber.Ctx) error {
// 		return c.SendFile("./framework/web/shell.html")
// 	})
// 	app.Get("/shell", func(c *fiber.Ctx) error {
// 		return c.SendFile("./framework/web/shell.html")
// 	})
//
// 	app.Get("/health", func(c *fiber.Ctx) error {
// 		if err := pool.Ping(c.Context()); err != nil {
// 			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
// 				"status": "unhealthy", "db": err.Error(),
// 			})
// 		}
// 		return c.JSON(fiber.Map{"status": "ok"})
// 	})
//
// 	quit := make(chan os.Signal, 1)
// 	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
//
// 	go func() {
// 		addr := cfg.Server.Host + ":" + cfg.Server.Port
// 		log.Info().Str("addr", addr).Msg("server listening")
// 		if err := app.Listen(addr); err != nil {
// 			log.Fatal().Err(err).Msg("server error")
// 		}
// 	}()
//
// 	<-quit
// 	log.Info().Msg("shutting down…")
//
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	if err := app.ShutdownWithContext(ctx); err != nil {
// 		log.Error().Err(err).Msg("shutdown error")
// 	}
// 	log.Info().Msg("server stopped")
// }
//
// func buildPool(cfg *config.Config) (*pgxpool.Pool, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
//
// 	poolCfg, err := pgxpool.ParseConfig(cfg.Database.GetDatabaseURL())
// 	if err != nil {
// 		return nil, fmt.Errorf("parse db url: %w", err)
// 	}
// 	poolCfg.MaxConns = int32(cfg.Database.MaxOpenConns)
// 	poolCfg.MinConns = int32(cfg.Database.MaxIdleConns)
// 	poolCfg.MaxConnLifetime = cfg.Database.ConnMaxLifetime
//
// 	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
// 	if err != nil {
// 		return nil, fmt.Errorf("create pool: %w", err)
// 	}
// 	if err := pool.Ping(ctx); err != nil {
// 		pool.Close()
// 		return nil, fmt.Errorf("ping db: %w", err)
// 	}
// 	return pool, nil
// }
//
// func setupLogger(cfg *config.Config) {
// 	level, err := zerolog.ParseLevel(cfg.Logger.Level)
// 	if err != nil {
// 		level = zerolog.InfoLevel
// 	}
// 	zerolog.SetGlobalLevel(level)
// 	if cfg.Logger.Format == "pretty" {
// 		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
// 	}
// }
//
// func tenantContextMiddleware() fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		if tenant := c.Get("X-Awo-Tenant"); tenant != "" {
// 			c.Locals("tenant_slug", tenant)
// 		}
// 		return c.Next()
// 	}
// }
//
// // slugTenantResolver looks up a tenant UUID by slug from the tenants table.
// // Used when the X-Awo-Tenant header contains a slug instead of a UUID.
// func slugTenantResolver(pool *pgxpool.Pool) api.TenantResolver {
// 	return func(ctx context.Context, slugOrID string) (uuid.UUID, error) {
// 		if id, err := uuid.Parse(slugOrID); err == nil {
// 			return id, nil
// 		}
// 		var id uuid.UUID
// 		err := pool.QueryRow(ctx,
// 			"SELECT id FROM tenants WHERE slug = $1 AND deleted_at IS NULL LIMIT 1",
// 			slugOrID,
// 		).Scan(&id)
// 		if err != nil {
// 			return uuid.Nil, fmt.Errorf("unknown tenant %q", slugOrID)
// 		}
// 		return id, nil
// 	}
// }
//
// func jsonErrorHandler(c *fiber.Ctx, err error) error {
// 	code := fiber.StatusInternalServerError
// 	msg := "internal server error"
// 	if e, ok := err.(*fiber.Error); ok {
// 		code = e.Code
// 		msg = e.Message
// 	}
// 	log.Error().Err(err).Int("status", code).Str("path", c.Path()).Msg("request error")
// 	return c.Status(code).JSON(fiber.Map{"status": code, "msg": msg})
// }
