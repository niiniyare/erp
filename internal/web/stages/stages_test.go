package stages_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"awo.so/internal/core/iam"
	"awo.so/internal/core/iam/contract"
	"awo.so/internal/pipeline"
	"awo.so/internal/platform/cache"
	"awo.so/internal/web/authz"
	"awo.so/internal/web/stages"
	"awo.so/internal/web/ui"
)

// ─── Test Helpers ─────────────────────────────────────────────────────────────

func newOpCtx(ctx context.Context) *pipeline.OperationContext {
	opCtx := pipeline.AcquireOperationContext()
	opCtx.Ctx = ctx
	opCtx.OperationKey = ui.OperationKey
	opCtx.TenantID = uuid.New()
	opCtx.UserID = uuid.New()
	return opCtx
}

// buildContractContext injects a contract.SessionContext into ctx.
func buildContractContext(userID, tenantID uuid.UUID) context.Context {
	resolved := &iam.ResolvedSession{
		UserID:   userID,
		TenantID: tenantID,
		UserType: "INTERNAL",
		Configuration: iam.Configuration{
			Flags:    map[string]bool{"advanced_reporting": true},
			Settings: map[string]string{"tenant.currency": "KES"},
			Prefs:    map[string]string{"ui.locale": "en-GB"},
		},
	}
	// Use the contract.InjectSessionContext path: build sc via the adapter helper.
	// We use WithContext directly since we're in a test (no Fiber).
	_ = resolved // suppress unused warning during compilation checks
	// Build SessionContext through the exported test helper.
	// In the real app, InjectSessionContext() middleware does this.
	sc := newSessionContextForTest(resolved)
	return contract.WithContext(context.Background(), sc)
}

// newSessionContextForTest constructs a contract.SessionContext for tests.
// Uses the same path as InjectSessionContext() middleware.
func newSessionContextForTest(resolved *iam.ResolvedSession) contract.SessionContext {
	stub := &stubSessionSvc{sess: resolved}
	adapter := contract.NewServiceAdapter(stub)
	sc, ok := adapter.ValidateSession(context.Background(), "test-token")
	if !ok {
		panic("stubSessionSvc.ValidateSession returned false — test setup error")
	}
	return sc
}

// stubSessionSvc satisfies iam.SessionService for test purposes.
type stubSessionSvc struct {
	sess *iam.ResolvedSession
}

func (s *stubSessionSvc) Login(_ context.Context, _, _ string) (*iam.ResolvedSession, string, error) {
	return s.sess, "tok", nil
}
func (s *stubSessionSvc) Logout(_ context.Context, _ string) error { return nil }
func (s *stubSessionSvc) ValidateSession(_ context.Context, _ string) (*iam.ResolvedSession, error) {
	return s.sess, nil
}
func (s *stubSessionSvc) CompleteMFALogin(_ context.Context, _, _ string) (*iam.ResolvedSession, string, error) {
	return nil, "", nil
}
func (s *stubSessionSvc) LogoutAllForUser(_ context.Context, _ uuid.UUID) error  { return nil }
func (s *stubSessionSvc) LogoutAllForTenant(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubSessionSvc) LoginWithSSO(_ context.Context, _ *iam.User) (*iam.ResolvedSession, string, error) {
	return nil, "", nil
}

// ─── TASK 1 — SessionStage Tests ─────────────────────────────────────────────

func TestSessionStage_PassesWithValidSession(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := buildContractContext(userID, tenantID)

	opCtx := newOpCtx(ctx)
	defer pipeline.ReleaseOperationContext(opCtx)

	stage := stages.NewSessionStage()
	result, err := stage.Execute(opCtx)

	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
}

func TestSessionStage_AbortsWithNoSession(t *testing.T) {
	opCtx := newOpCtx(context.Background()) // no session injected
	defer pipeline.ReleaseOperationContext(opCtx)

	stage := stages.NewSessionStage()
	_, err := stage.Execute(opCtx)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ui.ErrUnauthenticated))
}

func TestSessionStage_AbortsWithZeroSession(t *testing.T) {
	// Zero SessionContext should also fail.
	ctx := contract.WithContext(context.Background(), contract.SessionContext{})
	opCtx := newOpCtx(ctx)
	defer pipeline.ReleaseOperationContext(opCtx)

	stage := stages.NewSessionStage()
	_, err := stage.Execute(opCtx)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ui.ErrUnauthenticated))
}

func TestSessionStage_Priority(t *testing.T) {
	s := stages.NewSessionStage()
	assert.Equal(t, ui.PrioritySession, s.Priority())
	assert.True(t, s.Required())
}

// ─── TASK 2 — AuthzStage Tests ────────────────────────────────────────────────

func TestAuthzStage_ResolvesPermissions(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := buildContractContext(userID, tenantID)

	opCtx := newOpCtx(ctx)
	opCtx.TenantID = tenantID
	defer pipeline.ReleaseOperationContext(opCtx)

	svc := authz.GrantOnly("invoice.read", "invoice.create", "payment.read")
	stage := stages.NewAuthzStage(svc)

	result, err := stage.Execute(opCtx)

	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)

	// Permissions must be stored in outputs.
	perms, ok := result.Outputs[ui.DataKeyPermissions].(map[string]bool)
	require.True(t, ok, "DataKeyPermissions must be map[string]bool")
	assert.True(t, perms["invoice.read"])
	assert.True(t, perms["invoice.create"])
	assert.False(t, perms["invoice.delete"])

	// Fingerprints must be set.
	assert.NotEmpty(t, result.Outputs[ui.DataKeyPermFingerprint])
	assert.NotEmpty(t, result.Outputs[ui.DataKeyFlagFingerprint])

	// UISessionContext must be set.
	sess, ok := result.Outputs[ui.DataKeySessionCtx].(ui.UISessionContext)
	require.True(t, ok)
	assert.True(t, sess.Can("read", "invoice"))
	assert.False(t, sess.Can("delete", "invoice"))
}

func TestAuthzStage_FailsOnPermissionResolutionError(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := buildContractContext(userID, tenantID)

	opCtx := newOpCtx(ctx)
	defer pipeline.ReleaseOperationContext(opCtx)

	svc := &authz.MockUIAuthzService{Err: errors.New("casbin unavailable")}
	stage := stages.NewAuthzStage(svc)

	_, err := stage.Execute(opCtx)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ui.ErrPermissionResolution))
}

func TestAuthzStage_PermFingerprintIsStable(t *testing.T) {
	// Two users with identical permissions → identical fingerprint.
	svc := authz.GrantOnly("invoice.read", "report.read")

	userID1, tenantID1 := uuid.New(), uuid.New()
	userID2, tenantID2 := uuid.New(), uuid.New()

	ctx1 := buildContractContext(userID1, tenantID1)
	ctx2 := buildContractContext(userID2, tenantID2)

	op1 := newOpCtx(ctx1)
	op2 := newOpCtx(ctx2)
	defer pipeline.ReleaseOperationContext(op1)
	defer pipeline.ReleaseOperationContext(op2)

	r1, _ := stages.NewAuthzStage(svc).Execute(op1)
	r2, _ := stages.NewAuthzStage(svc).Execute(op2)

	fp1 := r1.Outputs[ui.DataKeyPermFingerprint].(string)
	fp2 := r2.Outputs[ui.DataKeyPermFingerprint].(string)

	assert.Equal(t, fp1, fp2, "identical permission sets must produce identical fingerprints")
}

func TestAuthzStage_Priority(t *testing.T) {
	s := stages.NewAuthzStage(authz.NoPermsGranted())
	assert.Equal(t, ui.PriorityAuthz, s.Priority())
	assert.True(t, s.Required())
	// Authz stage MUST run after SessionStage.
	assert.Greater(t, s.Priority(), stages.NewSessionStage().Priority())
}

// ─── TASK 3 — CacheLookupStage Tests ─────────────────────────────────────────

func TestCacheLookupStage_SkipsWhenFingerprintsAbsent(t *testing.T) {
	opCtx := newOpCtx(context.Background())
	opCtx.Input = ui.UISchemaInput{Route: "/finance/invoices"}
	defer pipeline.ReleaseOperationContext(opCtx)
	// No fingerprints in Data — simulate missing AuthzStage.

	svc := &mockCacheService{}
	stage := stages.NewCacheLookupStage(svc)
	result, err := stage.Execute(opCtx)

	require.NoError(t, err)
	assert.Equal(t, "skipped", result.Status)
	assert.Zero(t, svc.getCalls)
}

func TestCacheLookupStage_HitJumpsToResponse(t *testing.T) {
	schema := ui.Schema{"type": "page", "title": "Invoices"}
	svc := &mockCacheService{hitSchema: schema}

	opCtx := newOpCtx(context.Background())
	opCtx.TenantID = uuid.New()
	opCtx.Input = ui.UISchemaInput{Route: "/finance/invoices"}
	opCtx.Data[ui.DataKeyPermFingerprint] = "abc123"
	opCtx.Data[ui.DataKeyFlagFingerprint] = "def456"
	defer pipeline.ReleaseOperationContext(opCtx)

	stage := stages.NewCacheLookupStage(svc)
	result, err := stage.Execute(opCtx)

	require.NoError(t, err)
	assert.Equal(t, "ui.response", result.NextStageID, "cache hit must jump to response stage")
	assert.True(t, result.Outputs[ui.DataKeyCacheHit].(bool))
}

func TestCacheLookupStage_MissReturnsKey(t *testing.T) {
	svc := &mockCacheService{miss: true}

	opCtx := newOpCtx(context.Background())
	opCtx.TenantID = uuid.New()
	opCtx.Input = ui.UISchemaInput{Route: "/finance/invoices"}
	opCtx.Data[ui.DataKeyPermFingerprint] = "abc"
	opCtx.Data[ui.DataKeyFlagFingerprint] = "def"
	defer pipeline.ReleaseOperationContext(opCtx)

	stage := stages.NewCacheLookupStage(svc)
	result, err := stage.Execute(opCtx)

	require.NoError(t, err)
	assert.Empty(t, result.NextStageID, "cache miss must not jump")
	assert.NotEmpty(t, result.Outputs[ui.DataKeyCacheKey])
	assert.False(t, result.Outputs[ui.DataKeyCacheHit].(bool))
}

// ─── TASK 5 — CompileStage Tests ─────────────────────────────────────────────

func TestCompileStage_ExecutesPageFn(t *testing.T) {
	sess := buildUISession()
	pageFn := func(s ui.UISessionContext) ui.Schema {
		return ui.Schema{"type": "page", "title": "Test", "can_read": s.Can("read", "invoice")}
	}

	opCtx := newOpCtx(context.Background())
	opCtx.Data[ui.DataKeyPageFn] = ui.PageFn(pageFn)
	opCtx.Data[ui.DataKeySessionCtx] = sess
	defer pipeline.ReleaseOperationContext(opCtx)

	stage := stages.NewCompileStage()
	result, err := stage.Execute(opCtx)

	require.NoError(t, err)
	schema := result.Outputs[ui.DataKeySchema].(ui.Schema)
	assert.Equal(t, "page", schema["type"])
}

func TestCompileStage_RecoversPanic(t *testing.T) {
	sess := buildUISession()
	panicFn := ui.PageFn(func(_ ui.UISessionContext) ui.Schema {
		panic("intentional panic in PageFn")
	})

	opCtx := newOpCtx(context.Background())
	opCtx.Data[ui.DataKeyPageFn] = panicFn
	opCtx.Data[ui.DataKeySessionCtx] = sess
	defer pipeline.ReleaseOperationContext(opCtx)

	stage := stages.NewCompileStage()
	_, err := stage.Execute(opCtx)

	require.Error(t, err, "panic must be converted to error")
	assert.Contains(t, err.Error(), "panic")
}

func TestCompileStage_FailsWhenPageFnMissing(t *testing.T) {
	opCtx := newOpCtx(context.Background())
	defer pipeline.ReleaseOperationContext(opCtx)

	stage := stages.NewCompileStage()
	_, err := stage.Execute(opCtx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "DataKeyPageFn missing")
}

// ─── TASK 6/7 — NormalizeStage + ValidateStage Tests ─────────────────────────

func TestNormalizeStage_PassesCompliantSchema(t *testing.T) {
	schema := ui.Schema{
		"type": "page",
		"body": ui.Schema{
			"type":         "crud",
			"api":          "get:/api/v1/invoices",
			"syncLocation": true,
		},
	}
	opCtx := newOpCtx(context.Background())
	opCtx.Data[ui.DataKeySchema] = schema
	defer pipeline.ReleaseOperationContext(opCtx)

	_, err := stages.NewNormalizeStage().Execute(opCtx)
	assert.NoError(t, err)
}

func TestNormalizeStage_FailsOnMissingSyncLocation(t *testing.T) {
	schema := ui.Schema{
		"type": "page",
		"body": ui.Schema{
			"type": "crud",
			"api":  "get:/api/v1/invoices",
			// syncLocation missing
		},
	}
	opCtx := newOpCtx(context.Background())
	opCtx.Data[ui.DataKeySchema] = schema
	defer pipeline.ReleaseOperationContext(opCtx)

	_, err := stages.NewNormalizeStage().Execute(opCtx)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ui.ErrSchemaInvalid))
}

func TestNormalizeStage_FailsOnMissingAPIMethodPrefix(t *testing.T) {
	schema := ui.Schema{
		"type": "page",
		"body": ui.Schema{
			"type":         "crud",
			"api":          "/api/v1/invoices", // missing method prefix
			"syncLocation": true,
		},
	}
	opCtx := newOpCtx(context.Background())
	opCtx.Data[ui.DataKeySchema] = schema
	defer pipeline.ReleaseOperationContext(opCtx)

	_, err := stages.NewNormalizeStage().Execute(opCtx)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ui.ErrSchemaInvalid))
}

func TestValidateStage_PassesWhenNoIAMExpressions(t *testing.T) {
	schema := ui.Schema{
		"type":      "button",
		"visibleOn": "${can_approve_invoice}", // boolean set by Go
	}
	opCtx := newOpCtx(context.Background())
	opCtx.Data[ui.DataKeySchema] = schema
	defer pipeline.ReleaseOperationContext(opCtx)

	_, err := stages.NewValidateStage().Execute(opCtx)
	assert.NoError(t, err)
}

func TestValidateStage_RejectsRoleExpression(t *testing.T) {
	schema := ui.Schema{
		"type":      "button",
		"visibleOn": "${user.role === 'ADMIN'}", // IAM expression — forbidden
	}
	opCtx := newOpCtx(context.Background())
	opCtx.Data[ui.DataKeySchema] = schema
	defer pipeline.ReleaseOperationContext(opCtx)

	_, err := stages.NewValidateStage().Execute(opCtx)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ui.ErrSchemaInvalid))
}

func TestValidateStage_RejectsPermissionsField(t *testing.T) {
	schema := ui.Schema{
		"type":      "button",
		"visibleOn": "${user.permissions.includes('ADMIN')}", // IAM field — forbidden
	}
	opCtx := newOpCtx(context.Background())
	opCtx.Data[ui.DataKeySchema] = schema
	defer pipeline.ReleaseOperationContext(opCtx)

	_, err := stages.NewValidateStage().Execute(opCtx)
	require.Error(t, err)
}

// ─── TASK 8 — ResponseStage Tests ────────────────────────────────────────────

func TestResponseStage_AssemblesOutput(t *testing.T) {
	schema := ui.Schema{"type": "page", "title": "Dashboard"}

	opCtx := newOpCtx(context.Background())
	opCtx.Input = ui.UISchemaInput{Route: "/dashboard"}
	opCtx.Data[ui.DataKeySchema] = schema
	opCtx.Data[ui.DataKeyCacheHit] = false
	defer pipeline.ReleaseOperationContext(opCtx)

	stage := stages.NewResponseStage()
	result, err := stage.Execute(opCtx)

	require.NoError(t, err)
	out, ok := result.Outputs[ui.DataKeyResponse].(ui.UISchemaOutput)
	require.True(t, ok)
	assert.Equal(t, "page", out.Schema["type"])
	assert.Equal(t, "/dashboard", out.Route)
	assert.False(t, out.CacheHit)
}

func TestResponseStage_FailsWhenSchemaMissing(t *testing.T) {
	opCtx := newOpCtx(context.Background())
	defer pipeline.ReleaseOperationContext(opCtx)

	_, err := stages.NewResponseStage().Execute(opCtx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DataKeySchema missing")
}

// ─── TASK — Priority Ordering Tests (critical architectural invariant) ────────

func TestPipelinePriorityOrdering(t *testing.T) {
	// Enforce: cache lookup runs AFTER authz (fingerprints must be ready).
	assert.Greater(t, stages.NewCacheLookupStage(nil).Priority(),
		stages.NewAuthzStage(nil).Priority(),
		"CacheLookupStage must run after AuthzStage — cache key requires fingerprints")

	// Enforce: registry runs after cache (skip if cache hit).
	assert.Greater(t, stages.NewRegistryStage().Priority(),
		stages.NewCacheLookupStage(nil).Priority())

	// Enforce: compile runs after registry.
	assert.Greater(t, stages.NewCompileStage().Priority(),
		stages.NewRegistryStage().Priority())

	// Enforce: normalize runs after compile.
	assert.Greater(t, stages.NewNormalizeStage().Priority(),
		stages.NewCompileStage().Priority())

	// Enforce: validate runs after normalize.
	assert.Greater(t, stages.NewValidateStage().Priority(),
		stages.NewNormalizeStage().Priority())

	// Enforce: cache store runs after validate.
	assert.Greater(t, stages.NewCacheStoreStage(nil).Priority(),
		stages.NewValidateStage().Priority())

	// Enforce: response runs last.
	assert.Greater(t, stages.NewResponseStage().Priority(),
		stages.NewCacheStoreStage(nil).Priority())
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func buildUISession() ui.UISessionContext {
	sc := contract.SessionContext{} // zero — use NewUISessionContext with empty perms
	return ui.NewUISessionContext(sc, map[string]bool{
		"invoice.read":   true,
		"invoice.create": true,
	})
}

// mockCacheService satisfies cache.Service for stage tests.
// Only implements Get and Set — other methods not needed.
type mockCacheService struct {
	hitSchema ui.Schema
	miss      bool
	getCalls  int
	setCalls  int
}

func (m *mockCacheService) Get(_ context.Context, _ string, dest any) error {
	m.getCalls++
	if m.miss || m.hitSchema == nil {
		return cache.ErrCacheMiss
	}
	// Simulate cache hit by writing to dest.
	if ptr, ok := dest.(*ui.Schema); ok {
		*ptr = m.hitSchema
	}
	return nil
}

func (m *mockCacheService) Set(_ context.Context, _ string, _ any, _ time.Duration) error {
	m.setCalls++
	return nil
}

// Remaining cache.Service methods — not used in these tests.
func (m *mockCacheService) GetAndDelete(_ context.Context, _ string, _ any) error { return cache.ErrCacheMiss }
func (m *mockCacheService) Delete(_ context.Context, _ string) error               { return nil }
func (m *mockCacheService) Flush(_ context.Context) error                          { return nil }
func (m *mockCacheService) MGet(_ context.Context, _ []string) ([]cache.Result, error) {
	return nil, nil
}
func (m *mockCacheService) MSet(_ context.Context, _ map[string]any, _ time.Duration) error {
	return nil
}
func (m *mockCacheService) MDelete(_ context.Context, _ []string) error            { return nil }
func (m *mockCacheService) DeletePattern(_ context.Context, _ string) error        { return nil }
func (m *mockCacheService) Keys(_ context.Context, _ string) ([]string, error)     { return nil, nil }
func (m *mockCacheService) Exists(_ context.Context, _ string) (bool, error)       { return false, nil }
func (m *mockCacheService) TTL(_ context.Context, _ string) (time.Duration, error) { return 0, nil }
func (m *mockCacheService) Expire(_ context.Context, _ string, _ time.Duration) error { return nil }
func (m *mockCacheService) GetMemory(_ context.Context, _ string, _ any) error        { return cache.ErrCacheMiss }
func (m *mockCacheService) SetMemory(_ context.Context, _ string, _ any, _ time.Duration) error {
	return nil
}
func (m *mockCacheService) DeleteMemory(_ context.Context, _ string) error    { return nil }
func (m *mockCacheService) GetGlobalMemory(_ string, _ any) error             { return cache.ErrCacheMiss }
func (m *mockCacheService) SetGlobalMemory(_ string, _ any, _ time.Duration) error {
	return nil
}
func (m *mockCacheService) DeleteGlobalMemory(_ string) error { return nil }
func (m *mockCacheService) Ping(_ context.Context) error      { return nil }
func (m *mockCacheService) Stats() cache.CacheStats           { return cache.CacheStats{} }
func (m *mockCacheService) Reset()                            {}
func (m *mockCacheService) Close() error                      { return nil }

var _ cache.Service = (*mockCacheService)(nil)
