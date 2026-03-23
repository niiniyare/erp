package ui

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"

	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// UIHandler serves the AMIS-based frontend pages and schemas.
type UIHandler struct {
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.Service
}

// NewUIHandler creates a new UI handler.
func NewUIHandler(logger logger.Logger, metrics metrics.MetricsProvider, tracer tracing.Service) *UIHandler {
	return &UIHandler{
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// ServeDemo serves the main AMIS SPA shell page.
func (h *UIHandler) ServeDemo(c *fiber.Ctx) error {
	h.metrics.IncrementCounter("ui_page_views_total", metrics.Fields{"page": "demo"})
	return c.SendFile("./web/pages/index.html")
}

// ServeComponents serves the components demo page.
func (h *UIHandler) ServeComponents(c *fiber.Ctx) error {
	h.metrics.IncrementCounter("ui_page_views_total", metrics.Fields{"page": "components"})
	return c.SendFile("./web/pages/index.html")
}

// ServeForms serves the forms demo page.
func (h *UIHandler) ServeForms(c *fiber.Ctx) error {
	h.metrics.IncrementCounter("ui_page_views_total", metrics.Fields{"page": "forms"})
	return c.SendFile("./web/pages/index.html")
}

// ServeLogin serves the login page.
func (h *UIHandler) ServeLogin(c *fiber.Ctx) error {
	h.metrics.IncrementCounter("ui_page_views_total", metrics.Fields{"page": "login"})
	return c.SendFile("./web/pages/login.html")
}

// ServeSchema serves AMIS JSON schema files from web/schemas/.
func (h *UIHandler) ServeSchema(c *fiber.Ctx) error {
	schemaPath := c.Params("*")
	if schemaPath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": 1,
			"msg":    "schema path required",
		})
	}

	// Prevent directory traversal
	clean := filepath.Clean(schemaPath)
	if strings.Contains(clean, "..") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status": 1,
			"msg":    "invalid path",
		})
	}

	fullPath := filepath.Join("./web/schemas", clean)
	return c.SendFile(fullPath)
}

// ServeSDK serves AMIS SDK static assets from web/sdk/.
func (h *UIHandler) ServeSDK(c *fiber.Ctx) error {
	assetPath := c.Params("*")
	if assetPath == "" {
		return c.Status(fiber.StatusNotFound).SendString("not found")
	}

	clean := filepath.Clean(assetPath)
	if strings.Contains(clean, "..") {
		return c.Status(fiber.StatusForbidden).SendString("forbidden")
	}

	fullPath := filepath.Join("./web/sdk", clean)
	return c.SendFile(fullPath)
}

// RegisterRoutes registers all UI-related routes on the given router group.
// This is a convenience method for extended route setup beyond what routes.go defines.
func (h *UIHandler) RegisterRoutes(group fiber.Router) {
	// Schema API - serves JSON page definitions
	group.Get("/schemas/*", h.ServeSchema)

	// SDK assets
	group.Get("/sdk/*", h.ServeSDK)

	// Utils (schema-loader.js etc.)
	group.Static("/utils", "./web/utils")
}

// AMISResponse wraps backend data into AMIS-compatible response format.
// AMIS expects: {"status": 0, "msg": "", "data": {...}}
// Backend returns: {"success": true, "data": {...}, "request_id": "...", "timestamp": "..."}
func AMISResponse(data interface{}, msg string) fiber.Map {
	return fiber.Map{
		"status": 0,
		"msg":    msg,
		"data":   data,
	}
}

// AMISErrorResponse returns an AMIS-compatible error response.
func AMISErrorResponse(msg string) fiber.Map {
	return fiber.Map{
		"status": 1,
		"msg":    msg,
	}
}

// AMISListResponse wraps a list result for AMIS CRUD components.
// AMIS expects: {"status": 0, "data": {"items": [...], "total": N}}
func AMISListResponse(items interface{}, total int64) fiber.Map {
	return fiber.Map{
		"status": 0,
		"msg":    "",
		"data": fiber.Map{
			"items": items,
			"total": total,
		},
	}
}

// AMISAdapterMiddleware converts standard API responses to AMIS format.
// Apply this middleware to API routes that AMIS pages call.
func AMISAdapterMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only adapt if the request comes from AMIS (check header or query param)
		if c.Get("X-Requested-With") != "AMIS" && c.Query("_amis") == "" {
			return c.Next()
		}

		// Execute the handler
		err := c.Next()
		if err != nil {
			return err
		}

		// If response is JSON, try to adapt it
		if strings.Contains(string(c.Response().Header.ContentType()), "application/json") {
			body := c.Response().Body()

			var original map[string]interface{}
			if jsonErr := json.Unmarshal(body, &original); jsonErr == nil {
				// Check if it's already AMIS format
				if _, hasStatus := original["status"]; hasStatus {
					return nil
				}

				// Convert backend format to AMIS format
				amisResp := fiber.Map{
					"status": 0,
					"msg":    "",
				}

				if success, ok := original["success"].(bool); ok && !success {
					amisResp["status"] = 1
					if msg, ok := original["message"].(string); ok {
						amisResp["msg"] = msg
					}
				}

				if data, ok := original["data"]; ok {
					amisResp["data"] = data
				}

				adapted, _ := json.Marshal(amisResp)
				c.Response().SetBody(adapted)
			}
		}

		return nil
	}
}
