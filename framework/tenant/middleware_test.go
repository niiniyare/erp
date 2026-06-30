package tenant_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/framework/tenant"
)

func newResolutionApp() *fiber.App {
	app := fiber.New()
	app.Use(tenant.ResolutionMiddleware())
	app.Get("/api/v1/entities/invoice", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"tenant_slug": c.Locals("tenant_slug"),
		})
	})
	app.Get("/health/live", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})
	return app
}

func TestResolutionMiddleware_AwoTenantHeader(t *testing.T) {
	app := newResolutionApp()
	id := uuid.New()
	req := httptest.NewRequest("GET", "/api/v1/entities/invoice", nil)
	req.Header.Set("X-Awo-Tenant", id.String())
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestResolutionMiddleware_FallbackXTenantID(t *testing.T) {
	app := newResolutionApp()
	req := httptest.NewRequest("GET", "/api/v1/entities/invoice", nil)
	req.Header.Set("X-Tenant-ID", uuid.New().String())
	resp, _ := app.Test(req)
	// Deprecated header still works — returns 200 and logs warning.
	if resp.StatusCode != 200 {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestResolutionMiddleware_QueryParam(t *testing.T) {
	app := newResolutionApp()
	req := httptest.NewRequest("GET", "/api/v1/entities/invoice?tenant="+uuid.New().String(), nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestResolutionMiddleware_MissingTenant_Returns400(t *testing.T) {
	app := newResolutionApp()
	req := httptest.NewRequest("GET", "/api/v1/entities/invoice", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 400 {
		t.Errorf("want 400, got %d", resp.StatusCode)
	}
}

func TestResolutionMiddleware_GlobalPath_Skips(t *testing.T) {
	app := newResolutionApp()
	req := httptest.NewRequest("GET", "/health/live", nil)
	resp, _ := app.Test(req)
	// /health is a global route — no tenant required.
	if resp.StatusCode != 200 {
		t.Errorf("want 200 for health route, got %d", resp.StatusCode)
	}
}
