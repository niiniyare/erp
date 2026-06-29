package contract_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"awo.so/internal/core/iam"
	"awo.so/internal/core/iam/contract"
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

// resolvedSession builds a minimal *iam.ResolvedSession for tests.
func resolvedSession(userID, tenantID uuid.UUID) *iam.ResolvedSession {
	return &iam.ResolvedSession{
		UserID:   userID,
		TenantID: tenantID,
		UserType: "INTERNAL",
		Configuration: iam.Configuration{
			Flags:    map[string]bool{"billing.autopay": true, "hr.payroll_v2": false},
			Settings: map[string]string{"iam.session_ttl_hours": "4", "ui.max_items": "50"},
			Prefs:    map[string]string{"ui.theme": "dark", "locale": "en-GB"},
		},
	}
}

// platformSession returns a *iam.ResolvedSession with UserType = "SYSADMIN" (platform actor).
func platformSession(userID, tenantID uuid.UUID) *iam.ResolvedSession {
	s := resolvedSession(userID, tenantID)
	s.UserType = "SYSADMIN"
	return s
}

// getStatus sends a GET request to the Fiber app and returns the status code.
func getStatus(app *fiber.App, path string) int {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		panic(err)
	}
	return resp.StatusCode
}

// ─── SessionContext: Identity ─────────────────────────────────────────────────

func TestSessionContext_UserID(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	sc := contract.ExportNewSessionContext(resolvedSession(userID, tenantID))

	assert.Equal(t, userID, sc.UserID())
}

func TestSessionContext_TenantID(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	sc := contract.ExportNewSessionContext(resolvedSession(userID, tenantID))

	assert.Equal(t, tenantID, sc.TenantID())
}

func TestSessionContext_IsZero_WhenEmpty(t *testing.T) {
	var sc contract.SessionContext
	assert.True(t, sc.IsZero())
	assert.Equal(t, uuid.Nil, sc.UserID())
	assert.Equal(t, uuid.Nil, sc.TenantID())
}

func TestSessionContext_IsPlatform(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()

	platform := contract.ExportNewSessionContext(platformSession(userID, tenantID))
	tenant := contract.ExportNewSessionContext(resolvedSession(userID, tenantID))

	assert.True(t, platform.IsPlatform())
	assert.False(t, tenant.IsPlatform())
}

func TestSessionContext_IsPortal(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()

	portal := &iam.ResolvedSession{UserID: userID, TenantID: tenantID, UserType: "CUSTOMER"}
	internal := resolvedSession(userID, tenantID)

	assert.True(t, contract.ExportNewSessionContext(portal).IsPortal())
	assert.False(t, contract.ExportNewSessionContext(internal).IsPortal())
}

// ─── SessionContext: Feature Flags ───────────────────────────────────────────

func TestSessionContext_FeatureEnabled_Present(t *testing.T) {
	sc := contract.ExportNewSessionContext(resolvedSession(uuid.New(), uuid.New()))

	assert.True(t, sc.FeatureEnabled("billing.autopay"))
	assert.False(t, sc.FeatureEnabled("hr.payroll_v2")) // explicitly false
	assert.False(t, sc.FeatureEnabled("nonexistent"))   // absent
}

func TestSessionContext_FeatureEnabled_ZeroContext(t *testing.T) {
	var sc contract.SessionContext
	assert.False(t, sc.FeatureEnabled("any.flag"))
}

// ─── SessionContext: Tenant Settings ─────────────────────────────────────────

func TestSessionContext_Setting(t *testing.T) {
	sc := contract.ExportNewSessionContext(resolvedSession(uuid.New(), uuid.New()))

	assert.Equal(t, "4", sc.Setting("iam.session_ttl_hours", "8"))
	assert.Equal(t, "default", sc.Setting("missing.key", "default"))
}

func TestSessionContext_SettingBool(t *testing.T) {
	s := resolvedSession(uuid.New(), uuid.New())
	s.Configuration.Settings["feature.bool"] = "true"
	sc := contract.ExportNewSessionContext(s)

	assert.True(t, sc.SettingBool("feature.bool", false))
	assert.False(t, sc.SettingBool("missing", false))
}

func TestSessionContext_SettingInt(t *testing.T) {
	sc := contract.ExportNewSessionContext(resolvedSession(uuid.New(), uuid.New()))

	assert.Equal(t, 50, sc.SettingInt("ui.max_items", 0))
	assert.Equal(t, 99, sc.SettingInt("missing.key", 99))
}

func TestSessionContext_Settings_ZeroContext(t *testing.T) {
	var sc contract.SessionContext
	assert.Equal(t, "fallback", sc.Setting("any", "fallback"))
	assert.Equal(t, true, sc.SettingBool("any", true))
	assert.Equal(t, 42, sc.SettingInt("any", 42))
}

// ─── SessionContext: User Preferences ────────────────────────────────────────

func TestSessionContext_Preference(t *testing.T) {
	sc := contract.ExportNewSessionContext(resolvedSession(uuid.New(), uuid.New()))

	assert.Equal(t, "dark", sc.Preference("ui.theme", "light"))
	assert.Equal(t, "en-GB", sc.Preference("locale", "en-US"))
	assert.Equal(t, "default", sc.Preference("missing", "default"))
}

func TestSessionContext_Preference_ZeroContext(t *testing.T) {
	var sc contract.SessionContext
	assert.Equal(t, "fallback", sc.Preference("any", "fallback"))
}

// ─── SessionContext: Entity Scope ────────────────────────────────────────────

func TestSessionContext_EntityScope(t *testing.T) {
	s := resolvedSession(uuid.New(), uuid.New())
	s.EntityScope = iam.EntityScope{
		Type:     iam.EntityScopeSubtree,
		EntityID: uuid.New().String(),
	}
	sc := contract.ExportNewSessionContext(s)

	scope := sc.EntityScope()
	assert.Equal(t, iam.EntityScopeSubtree, scope.Type)
	assert.NotEmpty(t, scope.EntityID)
}

// ─── SessionContext: Immutability ─────────────────────────────────────────────

// TestSessionContext_IsValueType verifies that SessionContext is a value type
// with no exported setters. Consumers cannot mutate the session context.
func TestSessionContext_IsValueType(t *testing.T) {
	t.Parallel()
	typ := reflect.TypeOf(contract.SessionContext{})

	// Value type (not pointer receiver for mutation)
	assert.Equal(t, reflect.Struct, typ.Kind())

	// No exported fields (all access through methods)
	for i := range typ.NumField() {
		f := typ.Field(i)
		assert.False(t, f.IsExported(),
			"SessionContext field %q must be unexported — consumers use methods only", f.Name)
	}
}

// TestSessionContext_NoCanNoEnforce verifies that SessionContext exposes no
// authorization methods. Consumers must never evaluate permissions themselves.
func TestSessionContext_NoCanNoEnforce(t *testing.T) {
	t.Parallel()
	typ := reflect.TypeOf(contract.SessionContext{})

	forbidden := []string{"Can", "CanDo", "Enforce", "HasRole", "GetRoles", "Permissions"}
	for _, name := range forbidden {
		_, found := typ.MethodByName(name)
		assert.False(t, found,
			"SessionContext must not have method %q — authorization is IAM-only", name)
	}
}

// ─── Go Context Helpers ───────────────────────────────────────────────────────

func TestWithContext_FromContext_RoundTrip(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	sc := contract.ExportNewSessionContext(resolvedSession(userID, tenantID))

	ctx := contract.WithContext(context.Background(), sc)
	got, ok := contract.FromContext(ctx)

	require.True(t, ok, "FromContext must find the session after WithContext")
	assert.Equal(t, userID, got.UserID())
	assert.Equal(t, tenantID, got.TenantID())
}

func TestFromContext_Missing(t *testing.T) {
	_, ok := contract.FromContext(context.Background())
	assert.False(t, ok, "FromContext must return false for empty context")
}

func TestFromContext_ZeroSessionContext(t *testing.T) {
	// Storing a zero SessionContext must also return false.
	ctx := contract.WithContext(context.Background(), contract.SessionContext{})
	_, ok := contract.FromContext(ctx)
	assert.False(t, ok, "FromContext must return false for zero SessionContext")
}

func TestWithContext_DoesNotMutateParent(t *testing.T) {
	parent := context.Background()
	sc := contract.ExportNewSessionContext(resolvedSession(uuid.New(), uuid.New()))
	child := contract.WithContext(parent, sc)

	_, okParent := contract.FromContext(parent)
	_, okChild := contract.FromContext(child)

	assert.False(t, okParent, "parent context must not be mutated by WithContext")
	assert.True(t, okChild, "child context must contain the session")
}

// ─── Middleware: InjectSessionContext ─────────────────────────────────────────

func makeApp(middlewares ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	for _, mw := range middlewares {
		app.Use(mw)
	}
	return app
}

// injectLocals simulates what middleware.Authenticate does: puts a
// ResolvedSession into Fiber locals.
func injectLocals(s *iam.ResolvedSession) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(iam.LocalsKeySession, s)
		return c.Next()
	}
}

func TestInjectSessionContext_InjectsIntoGoContext(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	sess := resolvedSession(userID, tenantID)

	var capturedSC contract.SessionContext
	capture := func(c *fiber.Ctx) error {
		sc, ok := contract.FromContext(c.UserContext())
		if !ok {
			return fiber.NewError(fiber.StatusInternalServerError, "session missing from Go context")
		}
		capturedSC = sc
		return c.SendStatus(fiber.StatusOK)
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test",
		injectLocals(sess),
		contract.InjectSessionContext(),
		capture,
	)

	status := getStatus(app, "/test")
	require.Equal(t, fiber.StatusOK, status)
	assert.Equal(t, userID, capturedSC.UserID())
	assert.Equal(t, tenantID, capturedSC.TenantID())
}

func TestInjectSessionContext_Returns401_WhenNoSession(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test", contract.InjectSessionContext(), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	status := getStatus(app, "/test")
	assert.Equal(t, fiber.StatusUnauthorized, status)
}

func TestInjectSessionContext_PreservesExistingContextValues(t *testing.T) {
	type otherKey struct{}
	userID := uuid.New()
	tenantID := uuid.New()
	sess := resolvedSession(userID, tenantID)

	var capturedValue any
	check := func(c *fiber.Ctx) error {
		capturedValue = c.UserContext().Value(otherKey{})
		return c.SendStatus(fiber.StatusOK)
	}

	preFill := func(c *fiber.Ctx) error {
		ctx := context.WithValue(c.UserContext(), otherKey{}, "marker")
		c.SetUserContext(ctx)
		return c.Next()
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test", preFill, injectLocals(sess), contract.InjectSessionContext(), check)

	getStatus(app, "/test")
	assert.Equal(t, "marker", capturedValue, "InjectSessionContext must not wipe existing context values")
}

// ─── AuthService: SessionServiceAdapter ──────────────────────────────────────

// stubSessionService is a minimal in-memory SessionService stub for adapter tests.
// Compile-time assertion that it satisfies the full iam.SessionService interface.
var _ iam.SessionService = (*stubSessionService)(nil)

type stubSessionService struct {
	loginResult  *iam.ResolvedSession
	loginToken   string
	loginErr     error
	logoutErr    error
	validateSess *iam.ResolvedSession
	validateErr  error
}

func (s *stubSessionService) Login(_ context.Context, _, _ string) (*iam.ResolvedSession, string, error) {
	return s.loginResult, s.loginToken, s.loginErr
}
func (s *stubSessionService) Logout(_ context.Context, _ string) error { return s.logoutErr }
func (s *stubSessionService) ValidateSession(_ context.Context, _ string) (*iam.ResolvedSession, error) {
	return s.validateSess, s.validateErr
}

func (s *stubSessionService) CompleteMFALogin(_ context.Context, _, _ string) (*iam.ResolvedSession, string, error) {
	return nil, "", nil
}
func (s *stubSessionService) LogoutAllForUser(_ context.Context, _ uuid.UUID) error   { return nil }
func (s *stubSessionService) LogoutAllForTenant(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubSessionService) LoginWithSSO(_ context.Context, _ *iam.User) (*iam.ResolvedSession, string, error) {
	return nil, "", nil
}

func TestServiceAdapter_Login_Success(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	stub := &stubSessionService{
		loginResult: resolvedSession(userID, tenantID),
		loginToken:  "tok-abc123",
	}
	svc := contract.NewServiceAdapter(stub)

	result, err := svc.Login(context.Background(), "user@example.com", "password")

	require.NoError(t, err)
	assert.Equal(t, "tok-abc123", result.Token)
	assert.Equal(t, userID, result.Session.UserID())
	assert.Equal(t, tenantID, result.Session.TenantID())
	assert.False(t, result.Session.IsZero())
}

func TestServiceAdapter_Login_Error(t *testing.T) {
	stub := &stubSessionService{loginErr: fmt.Errorf("invalid credentials")}
	svc := contract.NewServiceAdapter(stub)

	result, err := svc.Login(context.Background(), "bad@example.com", "wrong")

	require.Error(t, err)
	assert.True(t, result.Session.IsZero())
}

func TestServiceAdapter_Logout_Delegates(t *testing.T) {
	stub := &stubSessionService{logoutErr: nil}
	svc := contract.NewServiceAdapter(stub)

	err := svc.Logout(context.Background(), "some-token")
	assert.NoError(t, err)
}

func TestServiceAdapter_Logout_PropagatesError(t *testing.T) {
	stub := &stubSessionService{logoutErr: fmt.Errorf("session not found")}
	svc := contract.NewServiceAdapter(stub)

	err := svc.Logout(context.Background(), "bad-token")
	assert.Error(t, err)
}

func TestServiceAdapter_ValidateSession_Valid(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	stub := &stubSessionService{validateSess: resolvedSession(userID, tenantID)}
	svc := contract.NewServiceAdapter(stub)

	sc, ok := svc.ValidateSession(context.Background(), "valid-token")

	require.True(t, ok)
	assert.Equal(t, userID, sc.UserID())
}

func TestServiceAdapter_ValidateSession_Invalid(t *testing.T) {
	stub := &stubSessionService{validateErr: fmt.Errorf("token expired")}
	svc := contract.NewServiceAdapter(stub)

	sc, ok := svc.ValidateSession(context.Background(), "expired-token")

	assert.False(t, ok)
	assert.True(t, sc.IsZero())
}

// ─── Anti-Bypass Guards ───────────────────────────────────────────────────────

// TestContract_SessionContext_NoPermissionsField verifies that SessionContext
// carries no permissions map. Authorization must remain in Casbin only.
func TestContract_SessionContext_NoPermissionsField(t *testing.T) {
	t.Parallel()
	typ := reflect.TypeOf(contract.SessionContext{})
	for i := range typ.NumField() {
		f := typ.Field(i)
		assert.NotContains(t, f.Name, "Permission",
			"SessionContext must carry no permission data")
		assert.NotContains(t, f.Name, "Roles",
			"SessionContext must carry no role data")
	}
}

// TestContract_AuthService_NoEnforce verifies that AuthService exposes no
// Casbin-level methods. Enforcement belongs in IAM middleware only.
func TestContract_AuthService_NoEnforce(t *testing.T) {
	t.Parallel()
	typ := reflect.TypeOf((*contract.AuthService)(nil)).Elem()
	forbidden := []string{"Enforce", "EnforceBatch", "AddPolicy", "GetRoles", "HasRole"}
	for _, name := range forbidden {
		_, found := typ.MethodByName(name)
		assert.False(t, found,
			"AuthService must not expose method %q — authz is IAM middleware only", name)
	}
}

// TestContract_MFARequired_Propagated verifies that ErrMFARequired is not
// swallowed by the adapter — callers need to detect it via errors.Is.
func TestContract_MFARequired_Propagated(t *testing.T) {
	mfaErr := errors.New("mfa required") // approximation — real code uses sentinel
	stub := &stubSessionService{loginErr: mfaErr, loginToken: "pending-mfa-tok"}
	svc := contract.NewServiceAdapter(stub)

	result, err := svc.Login(context.Background(), "user@example.com", "password")

	require.Error(t, err)
	assert.Equal(t, mfaErr, err, "adapter must propagate MFA error unchanged for errors.Is detection")
	// Token is passed through even on MFA error (pending token for MFA flow).
	assert.Equal(t, "pending-mfa-tok", result.Token)
}
