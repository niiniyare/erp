package health

import (
	"context"
	"html/template"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// HealthStatus represents the overall health status
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

// DependencyChecker defines the interface for checking dependencies
type DependencyChecker interface {
	CheckHealth(ctx context.Context) (bool, error)
	Name() string
}

// HealthHandler handles health check endpoints
type HealthHandler struct {
	logger       logger.Logger
	metrics      metrics.MetricsProvider
	tracer       tracing.Service
	startTime    time.Time
	version      string
	buildTime    string
	dependencies map[string]DependencyChecker
	htmlTemplate *template.Template
}

// Config holds configuration for the health handler
type Config struct {
	Version      string
	BuildTime    string
	Dependencies map[string]DependencyChecker
}

// NewHealthHandler creates a new health handler with configuration
func NewHealthHandler(
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
	cfg *Config,
) *HealthHandler {
	if cfg == nil {
		cfg = &Config{
			Version:   "unknown",
			BuildTime: time.Now().Format(time.RFC3339),
		}
	}

	return &HealthHandler{
		logger:       logger,
		metrics:      metrics,
		tracer:       tracer,
		startTime:    time.Now(),
		version:      cfg.Version,
		buildTime:    cfg.BuildTime,
		dependencies: cfg.Dependencies,
		htmlTemplate: parseHealthTemplate(),
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status       HealthStatus                `json:"status"`
	Timestamp    string                      `json:"timestamp"`
	Uptime       string                      `json:"uptime"`
	Version      string                      `json:"version"`
	BuildTime    string                      `json:"build_time,omitempty"`
	Dependencies map[string]DependencyStatus `json:"dependencies,omitempty"`
}

// DependencyStatus represents the status of a single dependency
type DependencyStatus struct {
	Status  HealthStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Latency string       `json:"latency,omitempty"`
}

// Get handles the main health check endpoint
func (h *HealthHandler) Get(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "health.Get")
	defer span.End()

	h.logger.InfoContext(ctx, "health check requested", logger.Fields{
		"method":      c.Method(),
		"path":        c.Path(),
		"user_agent":  c.Get("User-Agent"),
		"remote_addr": c.IP(),
	})

	h.metrics.IncrementCounter("health_checks_total", metrics.Fields{
		"endpoint": "health",
	})

	checkDeps := c.QueryBool("check_deps", false)
	acceptHeader := c.Get("Accept")

	response := h.buildHealthResponse(ctx, checkDeps)

	// Set appropriate HTTP status code
	statusCode := h.getStatusCode(response.Status)
	c.Status(statusCode)

	// Content negotiation
	if strings.Contains(acceptHeader, "text/html") {
		return h.renderHTML(c, response)
	}

	c.Set("Content-Type", "application/json; charset=utf-8")
	return c.JSON(response)
}

// Live handles Kubernetes liveness probe
// Returns 200 if the application is alive (not deadlocked)
func (h *HealthHandler) Live(c *fiber.Ctx) error {
	_, span := h.tracer.StartSpan(c.UserContext(), "health.Live")
	defer span.End()

	h.metrics.IncrementCounter("health_checks_total", metrics.Fields{
		"endpoint": "liveness",
	})

	response := fiber.Map{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	c.Set("Content-Type", "application/json; charset=utf-8")
	return c.JSON(response)
}

// Ready handles Kubernetes readiness probe
// Returns 200 if the application is ready to serve traffic
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "health.Ready")
	defer span.End()

	h.metrics.IncrementCounter("health_checks_total", metrics.Fields{
		"endpoint": "readiness",
	})

	// Check critical dependencies
	response := h.buildHealthResponse(ctx, true)

	statusCode := h.getStatusCode(response.Status)
	c.Status(statusCode)

	c.Set("Content-Type", "application/json; charset=utf-8")
	return c.JSON(response)
}

// Startup handles Kubernetes startup probe
// Returns 200 once the application has started successfully
func (h *HealthHandler) Startup(c *fiber.Ctx) error {
	_, span := h.tracer.StartSpan(c.UserContext(), "health.Startup")
	defer span.End()

	h.metrics.IncrementCounter("health_checks_total", metrics.Fields{
		"endpoint": "startup",
	})

	// Check if application has finished initialization
	if time.Since(h.startTime) < 5*time.Second {
		c.Status(fiber.StatusServiceUnavailable)
		return c.JSON(fiber.Map{
			"status":  "starting",
			"message": "application is still starting up",
			"uptime":  time.Since(h.startTime).String(),
		})
	}

	c.Set("Content-Type", "application/json; charset=utf-8")
	return c.JSON(fiber.Map{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(h.startTime).String(),
	})
}

// buildHealthResponse constructs the health response
func (h *HealthHandler) buildHealthResponse(ctx context.Context, checkDeps bool) HealthResponse {
	response := HealthResponse{
		Status:    StatusHealthy,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    time.Since(h.startTime).String(),
		Version:   h.version,
		BuildTime: h.buildTime,
	}

	if checkDeps && len(h.dependencies) > 0 {
		depStatuses, overallStatus := h.checkDependencies(ctx)
		response.Dependencies = depStatuses
		response.Status = overallStatus
	}

	return response
}

// checkDependencies checks all registered dependencies
func (h *HealthHandler) checkDependencies(ctx context.Context) (map[string]DependencyStatus, HealthStatus) {
	statuses := make(map[string]DependencyStatus)
	overallStatus := StatusHealthy
	unhealthyCount := 0

	for name, checker := range h.dependencies {
		start := time.Now()
		healthy, err := checker.CheckHealth(ctx)
		latency := time.Since(start)

		status := DependencyStatus{
			Latency: latency.String(),
		}

		if healthy {
			status.Status = StatusHealthy
		} else {
			status.Status = StatusUnhealthy
			unhealthyCount++
			if err != nil {
				status.Message = err.Error()
			}
		}

		statuses[name] = status

		// Record metrics
		h.metrics.IncrementCounter("dependency_checks_total", metrics.Fields{
			"dependency": name,
			"status":     string(status.Status),
		})

		h.metrics.ObserveHistogram("dependency_check_duration_seconds",
			latency.Seconds(),
			metrics.Fields{"dependency": name},
		)
	}

	// Determine overall status
	if unhealthyCount > 0 {
		if unhealthyCount == len(h.dependencies) {
			overallStatus = StatusUnhealthy
		} else {
			overallStatus = StatusDegraded
		}
	}

	return statuses, overallStatus
}

// getStatusCode returns the appropriate HTTP status code
func (h *HealthHandler) getStatusCode(status HealthStatus) int {
	switch status {
	case StatusHealthy:
		return fiber.StatusOK
	case StatusDegraded:
		return fiber.StatusOK // Still serving traffic
	case StatusUnhealthy:
		return fiber.StatusServiceUnavailable
	default:
		return fiber.StatusOK
	}
}

// renderHTML renders the health check as HTML
func (h *HealthHandler) renderHTML(c *fiber.Ctx, response HealthResponse) error {
	c.Set("Content-Type", "text/html; charset=utf-8")

	var buf strings.Builder
	if err := h.htmlTemplate.Execute(&buf, response); err != nil {
		h.logger.Error("failed to render health HTML", logger.Fields{
			"error": err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).SendString("Error rendering health page")
	}

	return c.SendString(buf.String())
}

// parseHealthTemplate parses the HTML template for health checks
func parseHealthTemplate() *template.Template {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Health Status</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        .status { font-size: 24px; font-weight: bold; margin: 20px 0; }
        .healthy { color: green; }
        .degraded { color: orange; }
        .unhealthy { color: red; }
        .info { background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 10px 0; }
        .dependencies { margin-top: 30px; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background: #f5f5f5; font-weight: bold; }
    </style>
</head>
<body>
    <h1>System Health Status</h1>
    <div class="status {{.Status}}">Status: {{.Status}}</div>
    
    <div class="info">
        <p><strong>Version:</strong> {{.Version}}</p>
        <p><strong>Build Time:</strong> {{.BuildTime}}</p>
        <p><strong>Uptime:</strong> {{.Uptime}}</p>
        <p><strong>Timestamp:</strong> {{.Timestamp}}</p>
    </div>

    {{if .Dependencies}}
    <div class="dependencies">
        <h2>Dependencies</h2>
        <table>
            <thead>
                <tr>
                    <th>Service</th>
                    <th>Status</th>
                    <th>Latency</th>
                    <th>Message</th>
                </tr>
            </thead>
            <tbody>
                {{range $name, $dep := .Dependencies}}
                <tr>
                    <td>{{$name}}</td>
                    <td class="{{$dep.Status}}">{{$dep.Status}}</td>
                    <td>{{$dep.Latency}}</td>
                    <td>{{$dep.Message}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
    {{end}}
</body>
</html>`

	return template.Must(template.New("health").Parse(tmpl))
}
