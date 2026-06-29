package activities_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"awo.so/internal/core/tenant/activities"
	"awo.so/internal/core/tenant/domain"
	"awo.so/internal/core/tenant/service"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/tracing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ---------------------------------------------------------------------------
// ActivitySuite
// ---------------------------------------------------------------------------

type ActivitySuite struct {
	suite.Suite
	repo *mockRepo
	acts *activities.Activities
}

func TestActivitySuite(t *testing.T) { suite.Run(t, new(ActivitySuite)) }

func (s *ActivitySuite) SetupTest() {
	s.repo = newMockRepo()
	tenantSvc := service.NewTenantService(
		s.repo,
		newMockCache(),
		noopTracer{},
		noopLogger{},
	)
	provSvc := service.NewProvisioningService(s.repo, noopTracer{})
	s.acts = activities.New(activities.Deps{
		TenantService:       tenantSvc,
		ProvisioningService: provSvc,
		Repo:                s.repo,
		Tracer:              noopTracer{},
	})
}

// ---- ProvisionTenantActivity (TN-ACT-001) ---------------------------------

func (s *ActivitySuite) TestProvisionTenantActivity_Success() {
	wantID := uuid.New()
	s.repo.provisionFn = func(_ context.Context, input domain.ProvisioningInput) (*domain.ProvisioningResult, error) {
		return &domain.ProvisioningResult{
			TenantID: wantID,
			Slug:     "acme",
			Status:   "pending",
		}, nil
	}

	result, err := s.acts.ProvisionTenantActivity(context.Background(), domain.ProvisioningInput{
		Name:  "Acme",
		Email: "a@b.com",
	})

	require.NoError(s.T(), err)
	require.NotNil(s.T(), result)
	require.Equal(s.T(), wantID, result.TenantID)
}

func (s *ActivitySuite) TestProvisionTenantActivity_Error() {
	s.repo.provisionFn = func(_ context.Context, _ domain.ProvisioningInput) (*domain.ProvisioningResult, error) {
		return nil, fmt.Errorf("db error")
	}

	result, err := s.acts.ProvisionTenantActivity(context.Background(), domain.ProvisioningInput{
		Name:  "Fail",
		Email: "f@f.com",
	})

	require.Error(s.T(), err)
	require.Nil(s.T(), result)
	require.Contains(s.T(), err.Error(), "provision tenant activity failed")
}

// ---- CreateDefaultConfigActivity (TN-ACT-002) ----------------------------

func (s *ActivitySuite) TestCreateDefaultConfigActivity_Success() {
	err := s.acts.CreateDefaultConfigActivity(context.Background(), uuid.New())
	require.NoError(s.T(), err)
}

func (s *ActivitySuite) TestCreateDefaultConfigActivity_Error() {
	s.repo.createDefaultConfigFn = func(_ context.Context, _ uuid.UUID) error {
		return fmt.Errorf("config creation failed")
	}

	err := s.acts.CreateDefaultConfigActivity(context.Background(), uuid.New())
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "create default config activity failed")
}

// ---- InitUsageActivity ----------------------------------------------------

func (s *ActivitySuite) TestInitUsageActivity_Success() {
	err := s.acts.InitUsageActivity(context.Background(), uuid.New())
	require.NoError(s.T(), err)
}

func (s *ActivitySuite) TestInitUsageActivity_Error() {
	s.repo.initUsageFn = func(_ context.Context, _ uuid.UUID) error {
		return fmt.Errorf("init failed")
	}

	err := s.acts.InitUsageActivity(context.Background(), uuid.New())
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "init usage activity failed")
}

// ---- ActivateTenantActivity (TN-ACT-003) ----------------------------------

func (s *ActivitySuite) TestActivateTenantActivity_Success() {
	t := s.seedTenant(domain.StatusPending)

	err := s.acts.ActivateTenantActivity(context.Background(), t.ID)
	require.NoError(s.T(), err)
}

func (s *ActivitySuite) TestActivateTenantActivity_NotFound() {
	err := s.acts.ActivateTenantActivity(context.Background(), uuid.New())
	require.Error(s.T(), err)
}

// ---- CleanupTenantActivity (TN-ACT-004) -----------------------------------

func (s *ActivitySuite) TestCleanupTenantActivity_Success() {
	t := s.seedTenant(domain.StatusActive)

	err := s.acts.CleanupTenantActivity(context.Background(), t.ID)
	require.NoError(s.T(), err)
}

func (s *ActivitySuite) TestCleanupTenantActivity_NotFound() {
	err := s.acts.CleanupTenantActivity(context.Background(), uuid.New())
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "cleanup tenant activity failed")
}

// ---- SendWelcomeNotificationActivity --------------------------------------

func (s *ActivitySuite) TestSendWelcomeNotificationActivity() {
	err := s.acts.SendWelcomeNotificationActivity(context.Background(), uuid.New())
	require.NoError(s.T(), err) // placeholder always succeeds
}

// ---- BulkUpdateStatusActivity (TN-ACT-005) --------------------------------

func (s *ActivitySuite) TestBulkUpdateStatusActivity_Success() {
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	err := s.acts.BulkUpdateStatusActivity(context.Background(), ids, domain.StatusSuspended)
	require.NoError(s.T(), err)
}

func (s *ActivitySuite) TestBulkUpdateStatusActivity_Error() {
	s.repo.bulkUpdateStatusFn = func(_ context.Context, _ []uuid.UUID, _ domain.TenantStatus) error {
		return fmt.Errorf("bulk fail")
	}

	err := s.acts.BulkUpdateStatusActivity(context.Background(), []uuid.UUID{uuid.New()}, domain.StatusSuspended)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "bulk update status activity failed")
}

// ---- BulkSoftDeleteActivity -----------------------------------------------

func (s *ActivitySuite) TestBulkSoftDeleteActivity_Success() {
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	err := s.acts.BulkSoftDeleteActivity(context.Background(), ids)
	require.NoError(s.T(), err)
}

func (s *ActivitySuite) TestBulkSoftDeleteActivity_Error() {
	s.repo.bulkSoftDeleteFn = func(_ context.Context, _ []uuid.UUID) error {
		return fmt.Errorf("delete fail")
	}

	err := s.acts.BulkSoftDeleteActivity(context.Background(), []uuid.UUID{uuid.New()})
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "bulk soft delete activity failed")
}

// ---- helpers --------------------------------------------------------------

func (s *ActivitySuite) seedTenant(status domain.TenantStatus) *domain.Tenant {
	now := time.Now()
	t := &domain.Tenant{
		ID:           uuid.New(),
		Slug:         "seeded",
		Name:         "Seeded Tenant",
		Email:        "seed@test.com",
		Status:       status,
		Timezone:     "UTC",
		CurrencyCode: "USD",
		Metadata:     make(map[string]any),
		Settings:     make(map[string]any),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.repo.seed(t)
	return t
}

// ===========================================================================
// Mocks — mirrors service_test mocks but in activities_test package
// ===========================================================================

type mockRepo struct {
	mu      sync.Mutex
	tenants map[uuid.UUID]*domain.Tenant

	createFn              func(ctx context.Context, t *domain.Tenant) error
	getByIDFn             func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	updateFn              func(ctx context.Context, id uuid.UUID, req domain.UpdateTenantRequest) error
	softDeleteFn          func(ctx context.Context, id uuid.UUID) error
	subdomainExistsFn     func(ctx context.Context, s string) (bool, error)
	provisionFn           func(ctx context.Context, input domain.ProvisioningInput) (*domain.ProvisioningResult, error)
	createDefaultConfigFn func(ctx context.Context, id uuid.UUID) error
	initUsageFn           func(ctx context.Context, id uuid.UUID) error
	bulkUpdateStatusFn    func(ctx context.Context, ids []uuid.UUID, status domain.TenantStatus) error
	bulkSoftDeleteFn      func(ctx context.Context, ids []uuid.UUID) error
}

func newMockRepo() *mockRepo {
	return &mockRepo{tenants: make(map[uuid.UUID]*domain.Tenant)}
}

func (r *mockRepo) seed(t *domain.Tenant) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[t.ID] = t
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

func (r *mockRepo) GetBySubdomain(_ context.Context, sub string) (*domain.Tenant, error) {
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

func (r *mockRepo) Exists(_ context.Context, id uuid.UUID) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.tenants[id]
	return ok, nil
}

func (r *mockRepo) List(_ context.Context, f domain.TenantFilter) ([]*domain.Tenant, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var all []*domain.Tenant
	for _, t := range r.tenants {
		if t.DeletedAt == nil {
			all = append(all, t)
		}
	}
	return all, int64(len(all)), nil
}

func (r *mockRepo) CreateDefaultConfig(ctx context.Context, id uuid.UUID) error {
	if r.createDefaultConfigFn != nil {
		return r.createDefaultConfigFn(ctx, id)
	}
	return nil
}

func (r *mockRepo) GetConfig(context.Context, uuid.UUID) (*domain.TenantConfiguration, error) {
	return nil, nil
}

func (r *mockRepo) GetUsage(context.Context, uuid.UUID) (*domain.TenantUsage, error) {
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
	return &domain.ProvisioningResult{TenantID: uuid.New(), Status: "pending"}, nil
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

func (r *mockRepo) GetGrowthStats(context.Context, int) ([]domain.GrowthStat, error) {
	return nil, nil
}

func (r *mockRepo) GetStatusDistribution(context.Context) (*domain.StatusCount, error) {
	return nil, nil
}
func (r *mockRepo) GetCurrentTenantID(context.Context) (uuid.UUID, error) { return uuid.Nil, nil }
func (r *mockRepo) GetCurrentTenant(context.Context) (*domain.Tenant, error) {
	return nil, nil
}
func (r *mockRepo) ValidateCurrentTenant(context.Context) error { return nil }

// ---------------------------------------------------------------------------
// Mock Cache
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

func (c *mockCache) Get(_ context.Context, key string, _ any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.data[key]; !ok {
		return cache.ErrCacheMiss
	}
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
// No-op Tracing / Logger
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

func (noopTracer) GetSpanID(context.Context) string { return "" }

func (noopTracer) Shutdown(context.Context) error { return nil }

type noopSpan struct{}

func (*noopSpan) End(...tracing.SpanEndOption)            {}
func (*noopSpan) SetAttributes(...attribute.KeyValue)     {}
func (*noopSpan) SetStatus(codes.Code, string)            {}
func (*noopSpan) RecordError(error, ...trace.EventOption) {}
func (*noopSpan) AddEvent(string, ...attribute.KeyValue)  {}
func (*noopSpan) IsRecording() bool                       { return false }
func (*noopSpan) SpanContext() trace.SpanContext          { return trace.SpanContext{} }
func (*noopSpan) SetName(string)                          {}

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
