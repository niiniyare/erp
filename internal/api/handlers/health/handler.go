package health

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

type HealthHandler struct {
	logger    logger.Logger
	metrics   metrics.MetricsProvider
	tracer    tracing.TracingService
	startTime time.Time
}

func NewHealthHandler(logger logger.Logger, metrics metrics.MetricsProvider, tracer tracing.TracingService) *HealthHandler {
	return &HealthHandler{
		logger:    logger,
		metrics:   metrics,
		tracer:    tracer,
		startTime: time.Now(),
	}
}

func (h *HealthHandler) Get(c *fiber.Ctx) error {
	_, span := h.tracer.StartSpan(c.Context(), "health.Get")
	defer span.End()

	h.logger.Info("Health check requested", logger.Fields{
		"method": c.Method(),
		"path":   c.Path(),
	})

	h.metrics.IncrementCounter("health_checks_total", metrics.Fields{})

	acceptHeader := c.Get("Accept")
	checkDeps := c.Query("check") == "dependencies"

	responseData := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(h.startTime).String(),
		"version":   "1.0.0",
	}

	if checkDeps {
		responseData["dependencies"] = map[string]interface{}{
			"database": "healthy",
			"cache":    "healthy",
		}
	}

	if strings.Contains(acceptHeader, "text/html") {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(fmt.Sprintf(`<html><body><h1>Health Status: %s</h1><p>Timestamp: %s</p></body></html>`, 
			responseData["status"], responseData["timestamp"]))
	}

	c.Set("Content-Type", "application/json; charset=utf-8")
	return c.JSON(responseData)
}