// Command ui starts the AMIS UI server.
//
// This is a lightweight standalone server — no database, no Redis, no Temporal.
// It serves:
//
//	GET /              → web/pages/index.html  (the shell)
//	GET /sdk/*         → web/sdk/              (AMIS JS/CSS)
//	GET /public/*      → web/public/           (logo, fonts, custom CSS)
//	GET /schema/*      → Go schema registry    (AMIS page schemas as JSON)
//
// API calls made by AMIS components (e.g. GET /api/v1/users) are proxied
// to the main API server via the fetcher in index.html. The UI server itself
// never touches the database.
//
// Port: UI_PORT env var, default 8081.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/fiber/v2/middleware/recover"

	webHandler "awo.so/internal/web/handler"

	// Each blank import runs init() which registers the schema into the registry.
	// Add one line here per new page — nothing else changes.
	_ "awo.so/internal/web/pages/dashboard"
	_ "awo.so/internal/web/pages/finance/accounts"
	_ "awo.so/internal/web/pages/finance/transactions"
	_ "awo.so/internal/web/pages/organizations"
	_ "awo.so/internal/web/pages/settings"
	_ "awo.so/internal/web/pages/users"
)

func main() {
	port := os.Getenv("UI_PORT")
	if port == "" {
		port = "8081"
	}

	apiBase := os.Getenv("API_BASE_URL")
	if apiBase == "" {
		apiBase = "http://localhost:8080"
	}

	app := fiber.New(fiber.Config{
		ServerHeader: "Awo-UI",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"status": code, "msg": err.Error()})
		},
	})

	app.Use(recover.New())
	app.Use(compress.New())

	// ── Static assets ──
	app.Static("/sdk", "./web/sdk", fiber.Static{Compress: true, MaxAge: 3600})
	app.Static("/utils", "./web/utils", fiber.Static{Compress: true, MaxAge: 3600})
	app.Static("/public", "./web/public", fiber.Static{Compress: true, MaxAge: 3600})

	// ── Schema API — no auth in dev; schemas contain no sensitive data ──
	// Auth is enforced by the API server on actual data endpoints.
	schemaH := webHandler.NewSchemaHandler()
	app.Get("/schema/*", schemaH.Handle)

	// ── API proxy — forward /api/* to the main API server ──
	// This keeps browser same-origin so cookies & CORS are not an issue.
	app.All("/api/*", proxy.Balancer(proxy.Config{
		Servers: []string{apiBase},
	}))

	// ── Shell — serve index.html for all other routes (SPA pattern) ──
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./web/pages/index.html")
	})

	// ── Graceful shutdown ──
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		addr := "0.0.0.0:" + port
		fmt.Printf("  UI server  →  http://localhost:%s\n", port)
		fmt.Printf("  API proxy  →  %s\n", apiBase)
		fmt.Printf("  Schemas    →  http://localhost:%s/schema/<route>\n", port)
		if err := app.Listen(addr); err != nil {
			log.Printf("UI server error: %v", err)
			cancel()
		}
	}()

	select {
	case sig := <-sigChan:
		fmt.Printf("\nReceived %v — shutting down UI server...\n", sig)
	case <-ctx.Done():
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("UI server shutdown error: %v", err)
	}
	fmt.Println("UI server stopped.")
}
