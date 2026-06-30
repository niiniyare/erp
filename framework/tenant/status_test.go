package tenant_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/framework/tenant"
)

// ── StatusMiddleware unit tests ────────────────────────────────────────────────

type mockStatusSvc struct {
	status tenant.TenantStatus
	err    error
}

func (m *mockStatusSvc) GetStatus(_ context.Context, _ uuid.UUID) (tenant.TenantStatus, error) {
	return m.status, m.err
}
func (m *mockStatusSvc) SetStatus(_ context.Context, _ uuid.UUID, _ tenant.TenantStatus) error {
	return nil
}

func newTestApp(status tenant.TenantStatus, err error) *fiber.App {
	app := fiber.New()
	svc := &mockStatusSvc{status: status, err: err}
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("tenant_id", uuid.New())
		c.Locals("request_id", "test-req-id")
		return c.Next()
	})
	app.Use(tenant.StatusMiddleware(svc))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(200) })
	return app
}

func TestStatusMiddleware_Active(t *testing.T) {
	app := newTestApp(tenant.StatusActive, nil)
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestStatusMiddleware_Pending(t *testing.T) {
	app := newTestApp(tenant.StatusPending, nil)
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 503 {
		t.Errorf("want 503, got %d", resp.StatusCode)
	}
	if v := resp.Header.Get("Retry-After"); v != "60" {
		t.Errorf("want Retry-After: 60, got %q", v)
	}
}

func TestStatusMiddleware_Suspended(t *testing.T) {
	app := newTestApp(tenant.StatusSuspended, nil)
	req := httptest.NewRequest("GET", "/", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 402 {
		t.Errorf("want 402, got %d", resp.StatusCode)
	}
}

func TestStatusMiddleware_Archived(t *testing.T) {
	app := newTestApp(tenant.StatusArchived, nil)
	req := httptest.NewRequest("GET", "/", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 410 {
		t.Errorf("want 410, got %d", resp.StatusCode)
	}
}

func TestStatusMiddleware_NoTenantLocals_Skips(t *testing.T) {
	// No tenant_id in Locals — middleware should pass through.
	app := fiber.New()
	app.Use(tenant.StatusMiddleware(&mockStatusSvc{status: tenant.StatusPending}))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	req := httptest.NewRequest("GET", "/", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("expected pass-through for no-tenant route, got %d", resp.StatusCode)
	}
}

func TestStatusMiddleware_NotFound(t *testing.T) {
	app := newTestApp("", tenant.ErrTenantNotFound)
	req := httptest.NewRequest("GET", "/", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 404 {
		t.Errorf("want 404, got %d", resp.StatusCode)
	}
}

// ── RedisStatusService unit tests ─────────────────────────────────────────────

func TestRedisStatusService_CacheHit(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()

	loaderCalled := 0
	svc := tenant.NewRedisStatusService(rdb, func(_ context.Context, _ uuid.UUID) (tenant.TenantStatus, error) {
		loaderCalled++
		return tenant.StatusActive, nil
	})

	id := uuid.New()

	// First call → cache miss → loader.
	s, err := svc.GetStatus(ctx, id)
	if err != nil || s != tenant.StatusActive {
		t.Fatalf("first call: want Active/nil, got %s/%v", s, err)
	}
	if loaderCalled != 1 {
		t.Fatalf("expected loader called once, got %d", loaderCalled)
	}

	// Second call → cache hit → loader NOT called again.
	s, err = svc.GetStatus(ctx, id)
	if err != nil || s != tenant.StatusActive {
		t.Fatalf("second call: want Active/nil, got %s/%v", s, err)
	}
	if loaderCalled != 1 {
		t.Fatalf("expected loader still called once, got %d", loaderCalled)
	}
}

func TestRedisStatusService_SetStatusInvalidatesCache(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()

	callCount := 0
	statuses := []tenant.TenantStatus{tenant.StatusActive, tenant.StatusSuspended}
	svc := tenant.NewRedisStatusService(rdb, func(_ context.Context, _ uuid.UUID) (tenant.TenantStatus, error) {
		s := statuses[callCount]
		callCount++
		return s, nil
	})

	id := uuid.New()

	// Populate cache with Active.
	s, _ := svc.GetStatus(ctx, id)
	if s != tenant.StatusActive {
		t.Fatalf("want Active, got %s", s)
	}

	// SetStatus → Suspended (write-through, resets TTL).
	if err := svc.SetStatus(ctx, id, tenant.StatusSuspended); err != nil {
		t.Fatal(err)
	}

	// Next read must return Suspended from cache (loader should NOT be called again).
	s, _ = svc.GetStatus(ctx, id)
	if s != tenant.StatusSuspended {
		t.Fatalf("want Suspended after SetStatus, got %s", s)
	}
	if callCount != 1 {
		t.Fatalf("loader called %d times, expected 1", callCount)
	}
}
