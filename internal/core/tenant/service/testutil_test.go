package service_test

import (
	"context"
	"net/http"
	"sync"
	"time"

	"awo.so/internal/core/tenant/domain"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ---------------------------------------------------------------------------
// Mock Repository
// ---------------------------------------------------------------------------

type mockRepo struct {
	mu      sync.Mutex
	tenants map[uuid.UUID]*domain.Tenant

	// hooks for custom behaviour per test
	createFn              func(ctx context.Context, t *domain.Tenant) error
	getByIDFn             func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	getBySubdomainFn      func(ctx context.Context, s string) (*domain.Tenant, error)
	updateFn              func(ctx context.Context, id uuid.UUID, req domain.UpdateTenantRequest) error
	softDeleteFn          func(ctx context.Context, id uuid.UUID) error
	subdomainExistsFn     func(ctx context.Context, s string) (bool, error)
	existsFn              func(ctx context.Context, id uuid.UUID) (bool, error)
	listFn                func(ctx context.Context, f domain.TenantFilter) ([]*domain.Tenant, int64, error)
	provisionFn           func(ctx context.Context, input domain.ProvisioningInput) (*domain.ProvisioningResult, error)
	createDefaultConfigFn func(ctx context.Context, id uuid.UUID) error
	initUsageFn           func(ctx context.Context, id uuid.UUID) error
	bulkUpdateStatusFn    func(ctx context.Context, ids []uuid.UUID, status domain.TenantStatus) error
	bulkSoftDeleteFn      func(ctx context.Context, ids []uuid.UUID) error
	getGrowthStatsFn      func(ctx context.Context, days int) ([]domain.GrowthStat, error)
	getStatusDistFn       func(ctx context.Context) (*domain.StatusCount, error)
	getConfigFn           func(ctx context.Context, id uuid.UUID) (*domain.TenantConfiguration, error)
	getUsageFn            func(ctx context.Context, id uuid.UUID) (*domain.TenantUsage, error)
}

func newMockRepo() *mockRepo {
	return &mockRepo{tenants: make(map[uuid.UUID]*domain.Tenant)}
}

func (r *mockRepo) Create(ctx context.Context, t *domain.Tenant) error {
	if r.createFn != nil {
		return r.createFn(ctx, t)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[t.ID] = t
	return nil
}

func (r *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	if r.getByIDFn != nil {
		return r.getByIDFn(ctx, id)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tenants[id]
	if !ok {
		return nil, domain.ErrTenantNotFound
	}
	return t, nil
}

func (r *mockRepo) GetBySubdomain(ctx context.Context, sub string) (*domain.Tenant, error) {
	if r.getBySubdomainFn != nil {
		return r.getBySubdomainFn(ctx, sub)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.tenants {
		if t.Subdomain != nil && *t.Subdomain == sub {
			return t, nil
		}
	}
	return nil, domain.ErrTenantNotFound
}

func (r *mockRepo) Update(ctx context.Context, id uuid.UUID, req domain.UpdateTenantRequest) error {
	if r.updateFn != nil {
		return r.updateFn(ctx, id, req)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tenants[id]
	if !ok {
		return domain.ErrTenantNotFound
	}
	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Status != nil {
		t.Status = *req.Status
	}
	t.UpdatedAt = time.Now()
	return nil
}

func (r *mockRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if r.softDeleteFn != nil {
		return r.softDeleteFn(ctx, id)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tenants[id]
	if !ok {
		return domain.ErrTenantNotFound
	}
	now := time.Now()
	t.DeletedAt = &now
	return nil
}

func (r *mockRepo) SubdomainExists(ctx context.Context, sub string) (bool, error) {
	if r.subdomainExistsFn != nil {
		return r.subdomainExistsFn(ctx, sub)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.tenants {
		if t.Subdomain != nil && *t.Subdomain == sub && t.DeletedAt == nil {
			return true, nil
		}
	}
	return false, nil
}

func (r *mockRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	if r.existsFn != nil {
		return r.existsFn(ctx, id)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.tenants[id]
	return ok, nil
}

func (r *mockRepo) List(ctx context.Context, f domain.TenantFilter) ([]*domain.Tenant, int64, error) {
	if r.listFn != nil {
		return r.listFn(ctx, f)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	var all []*domain.Tenant
	for _, t := range r.tenants {
		if t.DeletedAt == nil {
			all = append(all, t)
		}
	}
	total := int64(len(all))
	end := int(f.Offset + f.Limit)
	if end > len(all) {
		end = len(all)
	}
	start := int(f.Offset)
	if start > len(all) {
		start = len(all)
	}
	return all[start:end], total, nil
}

// Stubs for remaining interface methods — all support optional hooks.

func (r *mockRepo) CreateDefaultConfig(ctx context.Context, id uuid.UUID) error {
	if r.createDefaultConfigFn != nil {
		return r.createDefaultConfigFn(ctx, id)
	}
	return nil
}

func (r *mockRepo) GetConfig(ctx context.Context, id uuid.UUID) (*domain.TenantConfiguration, error) {
	if r.getConfigFn != nil {
		return r.getConfigFn(ctx, id)
	}
	return nil, nil
}

func (r *mockRepo) GetUsage(ctx context.Context, id uuid.UUID) (*domain.TenantUsage, error) {
	if r.getUsageFn != nil {
		return r.getUsageFn(ctx, id)
	}
	return nil, nil
}

func (r *mockRepo) InitUsage(ctx context.Context, id uuid.UUID) error {
	if r.initUsageFn != nil {
		return r.initUsageFn(ctx, id)
	}
	return nil
}

func (r *mockRepo) Provision(ctx context.Context, input domain.ProvisioningInput) (*domain.ProvisioningResult, error) {
	if r.provisionFn != nil {
		return r.provisionFn(ctx, input)
	}
	return &domain.ProvisioningResult{
		TenantID:  uuid.New(),
		Slug:      "provisioned",
		Subdomain: input.Subdomain,
		Status:    "pending",
	}, nil
}

func (r *mockRepo) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status domain.TenantStatus) error {
	if r.bulkUpdateStatusFn != nil {
		return r.bulkUpdateStatusFn(ctx, ids, status)
	}
	return nil
}

func (r *mockRepo) BulkSoftDelete(ctx context.Context, ids []uuid.UUID) error {
	if r.bulkSoftDeleteFn != nil {
		return r.bulkSoftDeleteFn(ctx, ids)
	}
	return nil
}

func (r *mockRepo) GetGrowthStats(ctx context.Context, days int) ([]domain.GrowthStat, error) {
	if r.getGrowthStatsFn != nil {
		return r.getGrowthStatsFn(ctx, days)
	}
	return nil, nil
}

func (r *mockRepo) GetStatusDistribution(ctx context.Context) (*domain.StatusCount, error) {
	if r.getStatusDistFn != nil {
		return r.getStatusDistFn(ctx)
	}
	return nil, nil
}

func (r *mockRepo) GetCurrentTenantID(context.Context) (uuid.UUID, error)    { return uuid.Nil, nil }
func (r *mockRepo) GetCurrentTenant(context.Context) (*domain.Tenant, error) { return nil, nil }
func (r *mockRepo) ValidateCurrentTenant(context.Context) error              { return nil }

// seed adds a tenant to the mock store for testing.
func (r *mockRepo) seed(t *domain.Tenant) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[t.ID] = t
}

// ---------------------------------------------------------------------------
// Mock Cache — simple in-memory for tests
// ---------------------------------------------------------------------------

type mockCache struct {
	mu   sync.Mutex
	data map[string]any
}

func newMockCache() *mockCache { return &mockCache{data: make(map[string]any)} }

func (c *mockCache) GetAndDelete(_ context.Context, key string, _ any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.data[key]; !ok {
		return cache.ErrCacheMiss
	}
	delete(c.data, key)
	return nil
}

func (c *mockCache) Get(_ context.Context, key string, dest any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.data[key]
	if !ok {
		return cache.ErrCacheMiss
	}
	// For test simplicity, we skip actual deserialization
	return nil
}

func (c *mockCache) Set(_ context.Context, key string, value any, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	return nil
}

func (c *mockCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	return nil
}

func (c *mockCache) Flush(context.Context) error                                 { return nil }
func (c *mockCache) MGet(context.Context, []string) ([]cache.Result, error)      { return nil, nil }
func (c *mockCache) MSet(context.Context, map[string]any, time.Duration) error   { return nil }
func (c *mockCache) MDelete(context.Context, []string) error                     { return nil }
func (c *mockCache) DeletePattern(context.Context, string) error                 { return nil }
func (c *mockCache) Keys(context.Context, string) ([]string, error)              { return nil, nil }
func (c *mockCache) Exists(context.Context, string) (bool, error)                { return false, nil }
func (c *mockCache) TTL(context.Context, string) (time.Duration, error)          { return 0, nil }
func (c *mockCache) Expire(context.Context, string, time.Duration) error         { return nil }
func (c *mockCache) GetMemory(context.Context, string, any) error                { return cache.ErrCacheMiss }
func (c *mockCache) SetMemory(context.Context, string, any, time.Duration) error { return nil }
func (c *mockCache) DeleteMemory(context.Context, string) error                  { return nil }
func (c *mockCache) GetGlobalMemory(string, any) error                           { return cache.ErrCacheMiss }
func (c *mockCache) SetGlobalMemory(string, any, time.Duration) error            { return nil }
func (c *mockCache) DeleteGlobalMemory(string) error                             { return nil }
func (c *mockCache) Ping(context.Context) error                                  { return nil }
func (c *mockCache) Stats() cache.CacheStats                                     { return cache.CacheStats{} }
func (c *mockCache) Reset()                                                      {}
func (c *mockCache) Close() error                                                { return nil }

// ---------------------------------------------------------------------------
// No-op Tracing
// ---------------------------------------------------------------------------

type noopTracer struct{}

func (noopTracer) StartSpan(ctx context.Context, _ string, _ ...tracing.SpanOption) (context.Context, tracing.Span) {
	return ctx, &noopSpan{}
}
func (noopTracer) SpanFromContext(context.Context) tracing.Span                          { return &noopSpan{} }
func (noopTracer) InjectHTTPHeaders(context.Context, http.Header)                        {}
func (noopTracer) ExtractHTTPHeaders(ctx context.Context, _ http.Header) context.Context { return ctx }
func (noopTracer) SetAttributes(context.Context, ...attribute.KeyValue)                  {}
func (noopTracer) RecordError(context.Context, error, ...tracing.ErrorOption)            {}
func (noopTracer) AddEvent(context.Context, string, ...attribute.KeyValue)               {}
func (noopTracer) GetTraceID(context.Context) string                                     { return "" }
func (noopTracer) GetSpanID(context.Context) string                                      { return "" }
func (noopTracer) Shutdown(context.Context) error                                        { return nil }

type noopSpan struct{}

func (*noopSpan) End(...tracing.SpanEndOption)            {}
func (*noopSpan) SetAttributes(...attribute.KeyValue)     {}
func (*noopSpan) SetStatus(codes.Code, string)            {}
func (*noopSpan) RecordError(error, ...trace.EventOption) {}
func (*noopSpan) AddEvent(string, ...attribute.KeyValue)  {}
func (*noopSpan) IsRecording() bool                       { return false }
func (*noopSpan) SpanContext() trace.SpanContext          { return trace.SpanContext{} }
func (*noopSpan) SetName(string)                          {}

// ---------------------------------------------------------------------------
// No-op Logger
// ---------------------------------------------------------------------------

type noopLogger struct{}

func (noopLogger) Debug(string, ...logger.Fields)                         {}
func (noopLogger) Info(string, ...logger.Fields)                          {}
func (noopLogger) Warn(string, ...logger.Fields)                          {}
func (noopLogger) Error(string, ...logger.Fields)                         {}
func (noopLogger) Fatal(string, ...logger.Fields)                         {}
func (noopLogger) DebugContext(context.Context, string, ...logger.Fields) {}
func (noopLogger) InfoContext(context.Context, string, ...logger.Fields)  {}
func (noopLogger) WarnContext(context.Context, string, ...logger.Fields)  {}
func (noopLogger) ErrorContext(context.Context, string, ...logger.Fields) {}
func (l noopLogger) WithFields(logger.Fields) logger.Logger               { return l }
func (l noopLogger) WithContext(context.Context) logger.Logger            { return l }
func (noopLogger) SetLevel(logger.LogLevel)                               {}
func (noopLogger) Close() error                                           { return nil }
