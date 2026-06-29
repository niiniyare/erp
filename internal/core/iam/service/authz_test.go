package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"awo.so/internal/core/iam/domain"
	"awo.so/internal/core/iam/service"
	"awo.so/internal/shared/logger"
)

// =============================================================================
// Stub AuthzRepository (in-memory no-op)
// =============================================================================

type stubAuthzRepo struct {
	upsertErr       error
	deactivateErr   error
	listAssignments []domain.RoleAssignment
	listExpiredErr  error
}

func (r *stubAuthzRepo) UpsertRoleAssignment(
	_ context.Context, _ uuid.UUID, _, _, _ string,
	_, _ *string, _ *time.Time,
) error {
	return r.upsertErr
}

func (r *stubAuthzRepo) DeactivateRoleAssignment(_ context.Context, _, _, _ string) error {
	return r.deactivateErr
}

func (r *stubAuthzRepo) ListRoleAssignments(_ context.Context, _, _ string) ([]domain.RoleAssignment, error) {
	return r.listAssignments, nil
}

func (r *stubAuthzRepo) ListExpiredActiveRoleNames(_ context.Context, _, _ string) ([]string, error) {
	return nil, r.listExpiredErr
}

// =============================================================================
// Stub SessionInvalidator
// =============================================================================

type stubSessionInv struct {
	invalidateUserCalls   []uuid.UUID
	invalidateTenantCalls []uuid.UUID
	invalidateUserErr     error
}

func (s *stubSessionInv) InvalidateByUser(_ context.Context, userID uuid.UUID) error {
	s.invalidateUserCalls = append(s.invalidateUserCalls, userID)
	return s.invalidateUserErr
}

func (s *stubSessionInv) InvalidateByTenant(_ context.Context, tenantID uuid.UUID) error {
	s.invalidateTenantCalls = append(s.invalidateTenantCalls, tenantID)
	return nil
}

// =============================================================================
// Helper: build in-memory AuthzService
// =============================================================================

func newTestAuthzService(t *testing.T, repo *stubAuthzRepo) service.AuthzService {
	t.Helper()
	log := logger.NewNoOp()
	svc, err := service.NewInMemoryAuthzService(repo, log)
	if err != nil {
		t.Fatalf("NewInMemoryAuthzService: %v", err)
	}
	return svc
}

func newTestAuthzServiceWithInv(t *testing.T, repo *stubAuthzRepo, inv *stubSessionInv) service.AuthzService {
	t.Helper()
	log := logger.NewNoOp()
	svc, err := service.NewInMemoryAuthzServiceWithSessionInv(repo, log, inv)
	if err != nil {
		t.Fatalf("NewInMemoryAuthzServiceWithSessionInv: %v", err)
	}
	return svc
}

func newTestAuthzServiceWithLimit(t *testing.T, repo *stubAuthzRepo, maxPolicies int) service.AuthzService {
	t.Helper()
	log := logger.NewNoOp()
	svc, err := service.NewInMemoryAuthzServiceWithLimit(repo, log, maxPolicies)
	if err != nil {
		t.Fatalf("NewInMemoryAuthzServiceWithLimit: %v", err)
	}
	return svc
}

// =============================================================================
// AuthzService — Enforce Suite
// =============================================================================

type EnforceSuite struct {
	suite.Suite
	svc  service.AuthzService
	repo *stubAuthzRepo
	ctx  context.Context
}

func TestEnforceSuite(t *testing.T) { suite.Run(t, new(EnforceSuite)) }

func (s *EnforceSuite) SetupTest() {
	s.repo = &stubAuthzRepo{}
	s.svc = newTestAuthzService(s.T(), s.repo)
	s.ctx = context.Background()
}

func (s *EnforceSuite) TestEnforce_EmptyFields_ReturnsInvalidRequest() {
	_, err := s.svc.Enforce(s.ctx, domain.Request{Subject: "", Domain: "dom", Object: "obj", Action: "read"})
	s.Require().Error(err)
	s.True(errors.Is(err, domain.ErrInvalidRequest))
}

func (s *EnforceSuite) TestEnforce_EmptyDomain() {
	_, err := s.svc.Enforce(s.ctx, domain.Request{Subject: "tenant:x", Domain: "", Object: "obj", Action: "read"})
	s.Require().Error(err)
	s.True(errors.Is(err, domain.ErrInvalidRequest))
}

func (s *EnforceSuite) TestEnforce_EmptyObject() {
	_, err := s.svc.Enforce(s.ctx, domain.Request{Subject: "tenant:x", Domain: "dom", Object: "", Action: "read"})
	s.Require().Error(err)
	s.True(errors.Is(err, domain.ErrInvalidRequest))
}

func (s *EnforceSuite) TestEnforce_EmptyAction() {
	_, err := s.svc.Enforce(s.ctx, domain.Request{Subject: "tenant:x", Domain: "dom", Object: "obj", Action: ""})
	s.Require().Error(err)
	s.True(errors.Is(err, domain.ErrInvalidRequest))
}

func (s *EnforceSuite) TestEnforce_PlatformSubject_PlatformDomain_Allowed() {
	allowed, err := s.svc.Enforce(s.ctx, domain.Request{
		Subject: domain.PlatformSubject("admin-uuid"),
		Domain:  domain.DomainPlatform,
		Object:  "anything",
		Action:  "delete",
	})
	s.Require().NoError(err)
	s.True(allowed)
}

func (s *EnforceSuite) TestEnforce_TenantSubject_PlatformDomain_Denied() {
	// A non-platform subject must never bypass the platform domain boundary.
	allowed, err := s.svc.Enforce(s.ctx, domain.Request{
		Subject: domain.TenantSubject("some-user"),
		Domain:  domain.DomainPlatform,
		Object:  "admin/settings",
		Action:  "write",
	})
	s.Require().NoError(err)
	s.False(allowed, "tenant subject must not access platform domain")
}

func (s *EnforceSuite) TestEnforce_NoPolicy_Denied() {
	allowed, err := s.svc.Enforce(s.ctx, domain.Request{
		Subject: domain.TenantSubject(uuid.New().String()),
		Domain:  uuid.New().String(),
		Object:  "invoice/*",
		Action:  "read",
	})
	s.Require().NoError(err)
	s.False(allowed)
}

func (s *EnforceSuite) TestEnforce_AllowPolicy_Permits() {
	tenantID := uuid.New().String()
	role := "role:accountant"
	subj := domain.TenantSubject(uuid.New().String())
	dom := domain.TenantDomain(tenantID)

	// Add allow policy for role
	s.Require().NoError(s.svc.AddPolicy(s.ctx, domain.Policy{
		Subject: role, Domain: dom, Object: "finance.accounts", Action: "read", Effect: "allow",
	}))
	// Assign role to subject
	s.Require().NoError(s.svc.AssignRole(s.ctx, tenantID, subj, role, dom,
		domain.WithAssignedBy("platform:system"),
	))

	allowed, err := s.svc.Enforce(s.ctx, domain.Request{
		Subject: subj, Domain: dom, Object: "finance.accounts", Action: "read",
	})
	s.Require().NoError(err)
	s.True(allowed)
}

func (s *EnforceSuite) TestEnforce_DenyOverridesAllow() {
	tenantID := uuid.New().String()
	role := "role:accountant"
	subj := domain.TenantSubject(uuid.New().String())
	dom := domain.TenantDomain(tenantID)

	// Allow read
	s.Require().NoError(s.svc.AddPolicy(s.ctx, domain.Policy{
		Subject: role, Domain: dom, Object: "finance.accounts", Action: "read", Effect: "allow",
	}))
	// Deny read (deny-override model: deny wins)
	s.Require().NoError(s.svc.AddPolicy(s.ctx, domain.Policy{
		Subject: role, Domain: dom, Object: "finance.accounts", Action: "read", Effect: "deny",
	}))
	s.Require().NoError(s.svc.AssignRole(s.ctx, tenantID, subj, role, dom,
		domain.WithAssignedBy("platform:system"),
	))

	allowed, err := s.svc.Enforce(s.ctx, domain.Request{
		Subject: subj, Domain: dom, Object: "finance.accounts", Action: "read",
	})
	s.Require().NoError(err)
	s.False(allowed, "deny policy should override allow")
}

func (s *EnforceSuite) TestEnforce_WildcardAction_Matches() {
	tenantID := uuid.New().String()
	role := "role:admin"
	subj := domain.TenantSubject(uuid.New().String())
	dom := domain.TenantDomain(tenantID)

	s.Require().NoError(s.svc.AddPolicy(s.ctx, domain.Policy{
		Subject: role, Domain: dom, Object: "iam.roles", Action: "*", Effect: "allow",
	}))
	s.Require().NoError(s.svc.AssignRole(s.ctx, tenantID, subj, role, dom,
		domain.WithAssignedBy("platform:system"),
	))

	for _, act := range []string{"read", "write", "delete"} {
		allowed, err := s.svc.Enforce(s.ctx, domain.Request{
			Subject: subj, Domain: dom, Object: "iam.roles", Action: act,
		})
		s.Require().NoError(err)
		s.True(allowed, "action=%q should match wildcard", act)
	}
}

func (s *EnforceSuite) TestEnforce_TenantIsolation() {
	tenantA := uuid.New().String()
	tenantB := uuid.New().String()
	role := "role:accountant"
	subjA := domain.TenantSubject(uuid.New().String())
	domA := domain.TenantDomain(tenantA)
	domB := domain.TenantDomain(tenantB)

	// Policy only in tenant A
	s.Require().NoError(s.svc.AddPolicy(s.ctx, domain.Policy{
		Subject: role, Domain: domA, Object: "finance.accounts", Action: "read", Effect: "allow",
	}))
	s.Require().NoError(s.svc.AssignRole(s.ctx, tenantA, subjA, role, domA,
		domain.WithAssignedBy("platform:system"),
	))

	// subjA cannot access tenant B's domain even with same role name
	allowed, err := s.svc.Enforce(s.ctx, domain.Request{
		Subject: subjA, Domain: domB, Object: "finance.accounts", Action: "read",
	})
	s.Require().NoError(err)
	s.False(allowed, "cross-tenant access must be denied")
}

// =============================================================================
// AuthzService — EnforceBatch Suite
// =============================================================================

type EnforceBatchSuite struct {
	suite.Suite
	svc service.AuthzService
	ctx context.Context
}

func TestEnforceBatchSuite(t *testing.T) { suite.Run(t, new(EnforceBatchSuite)) }

func (s *EnforceBatchSuite) SetupTest() {
	s.svc = newTestAuthzService(s.T(), &stubAuthzRepo{})
	s.ctx = context.Background()
}

func (s *EnforceBatchSuite) TestEnforceBatch_Nil_ReturnsNil() {
	res, err := s.svc.EnforceBatch(s.ctx, nil)
	s.Require().NoError(err)
	s.Nil(res)
}

func (s *EnforceBatchSuite) TestEnforceBatch_EmptyFields_Error() {
	_, err := s.svc.EnforceBatch(s.ctx, []domain.Request{
		{Subject: "", Domain: "d", Object: "o", Action: "a"},
	})
	s.Require().Error(err)
}

func (s *EnforceBatchSuite) TestEnforceBatch_PlatformBypass() {
	reqs := []domain.Request{
		{Subject: "platform:x", Domain: domain.DomainPlatform, Object: "any", Action: "delete"},
		{Subject: "platform:y", Domain: domain.DomainPlatform, Object: "foo", Action: "write"},
	}
	res, err := s.svc.EnforceBatch(s.ctx, reqs)
	s.Require().NoError(err)
	s.Require().Len(res, 2)
	s.True(res[0])
	s.True(res[1])
}

func (s *EnforceBatchSuite) TestEnforceBatch_MixedPlatformAndTenant() {
	tenantID := uuid.New().String()
	dom := domain.TenantDomain(tenantID)
	role := "role:accountant"
	subj := domain.TenantSubject(uuid.New().String())

	s.Require().NoError(s.svc.AddPolicy(s.ctx, domain.Policy{
		Subject: role, Domain: dom, Object: "finance.accounts", Action: "read", Effect: "allow",
	}))
	s.Require().NoError(s.svc.AssignRole(s.ctx, tenantID, subj, role, dom,
		domain.WithAssignedBy("platform:system"),
	))

	reqs := []domain.Request{
		{Subject: "platform:admin", Domain: domain.DomainPlatform, Object: "any", Action: "read"},
		{Subject: subj, Domain: dom, Object: "finance.accounts", Action: "read"},
		{Subject: subj, Domain: dom, Object: "finance.accounts", Action: "delete"},
	}
	res, err := s.svc.EnforceBatch(s.ctx, reqs)
	s.Require().NoError(err)
	s.Require().Len(res, 3)
	s.True(res[0], "platform actor must be allowed")
	s.True(res[1], "allowed action must pass")
	s.False(res[2], "non-allowed action must be denied")
}

// =============================================================================
// AuthzService — AssignRole Suite
// =============================================================================

type AssignRoleSuite struct {
	suite.Suite
	svc      service.AuthzService
	repo     *stubAuthzRepo
	ctx      context.Context
	tenantID string
	userID   string
	dom      string
	subj     string
}

func TestAssignRoleSuite(t *testing.T) { suite.Run(t, new(AssignRoleSuite)) }

func (s *AssignRoleSuite) SetupTest() {
	s.repo = &stubAuthzRepo{}
	s.svc = newTestAuthzService(s.T(), s.repo)
	s.ctx = context.Background()
	s.tenantID = uuid.New().String()
	s.userID = uuid.New().String()
	s.dom = domain.TenantDomain(s.tenantID)
	s.subj = domain.TenantSubject(s.userID)
}

func (s *AssignRoleSuite) TestAssignRole_EmptySubject() {
	err := s.svc.AssignRole(s.ctx, s.tenantID, "", "role:admin", s.dom,
		domain.WithAssignedBy("platform:system"))
	s.Require().Error(err)
	s.True(errors.Is(err, domain.ErrInvalidRequest))
}

func (s *AssignRoleSuite) TestAssignRole_EmptyRole() {
	err := s.svc.AssignRole(s.ctx, s.tenantID, s.subj, "", s.dom,
		domain.WithAssignedBy("platform:system"))
	s.Require().Error(err)
	s.True(errors.Is(err, domain.ErrInvalidRequest))
}

func (s *AssignRoleSuite) TestAssignRole_EmptyDomain() {
	err := s.svc.AssignRole(s.ctx, s.tenantID, s.subj, "role:admin", "",
		domain.WithAssignedBy("platform:system"))
	s.Require().Error(err)
	s.True(errors.Is(err, domain.ErrInvalidRequest))
}

func (s *AssignRoleSuite) TestAssignRole_MissingAssignedBy() {
	err := s.svc.AssignRole(s.ctx, s.tenantID, s.subj, "role:admin", s.dom)
	s.Require().Error(err)
	s.True(errors.Is(err, domain.ErrInvalidRequest))
}

func (s *AssignRoleSuite) TestAssignRole_PlatformRoleByNonPlatform_Forbidden() {
	err := s.svc.AssignRole(s.ctx, s.tenantID, s.subj, "role:platform-admin", s.dom,
		domain.WithAssignedBy("tenant:manager"))
	s.Require().Error(err)
	s.True(errors.Is(err, domain.ErrForbidden))
}

func (s *AssignRoleSuite) TestAssignRole_PlatformRoleByPlatform_Allowed() {
	err := s.svc.AssignRole(s.ctx, s.tenantID, s.subj, "role:platform-admin", s.dom,
		domain.WithAssignedBy("platform:system"))
	s.Require().NoError(err)
}

func (s *AssignRoleSuite) TestAssignRole_Success_RoleVisible() {
	role := "role:accountant"
	s.Require().NoError(s.svc.AssignRole(s.ctx, s.tenantID, s.subj, role, s.dom,
		domain.WithAssignedBy("platform:system"),
	))

	roles, err := s.svc.GetRoles(s.ctx, s.subj, s.dom)
	s.Require().NoError(err)
	s.Contains(roles, role)
}

func (s *AssignRoleSuite) TestAssignRole_WithExpiry() {
	future := time.Now().Add(24 * time.Hour)
	err := s.svc.AssignRole(s.ctx, s.tenantID, s.subj, "role:temp", s.dom,
		domain.WithAssignedBy("platform:system"),
		domain.WithExpiry(future),
	)
	s.Require().NoError(err)
}

func (s *AssignRoleSuite) TestAssignRole_RepoUpsertError_CompensatesEnforcer() {
	s.repo.upsertErr = errors.New("db down")
	err := s.svc.AssignRole(s.ctx, s.tenantID, s.subj, "role:accountant", s.dom,
		domain.WithAssignedBy("platform:system"),
	)
	s.Require().Error(err)

	// After compensation, role should NOT be visible in the enforcer
	roles, _ := s.svc.GetRoles(s.ctx, s.subj, s.dom)
	s.NotContains(roles, "role:accountant")
}

// =============================================================================
// AuthzService — RevokeRole Suite
// =============================================================================

type RevokeRoleSuite struct {
	suite.Suite
	svc      service.AuthzService
	repo     *stubAuthzRepo
	ctx      context.Context
	tenantID string
	dom      string
	subj     string
}

func TestRevokeRoleSuite(t *testing.T) { suite.Run(t, new(RevokeRoleSuite)) }

func (s *RevokeRoleSuite) SetupTest() {
	s.repo = &stubAuthzRepo{}
	s.svc = newTestAuthzService(s.T(), s.repo)
	s.ctx = context.Background()
	s.tenantID = uuid.New().String()
	s.dom = domain.TenantDomain(s.tenantID)
	s.subj = domain.TenantSubject(uuid.New().String())
}

func (s *RevokeRoleSuite) TestRevokeRole_EmptyFields() {
	err := s.svc.RevokeRole(s.ctx, "", "role:admin", s.dom)
	s.Require().Error(err)
	s.True(errors.Is(err, domain.ErrInvalidRequest))
}

func (s *RevokeRoleSuite) TestRevokeRole_RemovesRoleFromEnforcer() {
	role := "role:accountant"
	s.Require().NoError(s.svc.AssignRole(s.ctx, s.tenantID, s.subj, role, s.dom,
		domain.WithAssignedBy("platform:system"),
	))

	s.Require().NoError(s.svc.RevokeRole(s.ctx, s.subj, role, s.dom))

	roles, err := s.svc.GetRoles(s.ctx, s.subj, s.dom)
	s.Require().NoError(err)
	s.NotContains(roles, role)
}

func (s *RevokeRoleSuite) TestRevokeRole_RepoDeactivateError_Compensates() {
	role := "role:accountant"
	s.Require().NoError(s.svc.AssignRole(s.ctx, s.tenantID, s.subj, role, s.dom,
		domain.WithAssignedBy("platform:system"),
	))

	s.repo.deactivateErr = errors.New("db error")
	err := s.svc.RevokeRole(s.ctx, s.subj, role, s.dom)
	s.Require().Error(err)

	// AUTHZ-TXN-2: enforcer must be compensated back — role restored
	roles, _ := s.svc.GetRoles(s.ctx, s.subj, s.dom)
	s.Contains(roles, role)
}

// =============================================================================
// AuthzService — Role Query Suite (GetRoles / GetImplicitRoles / HasRole)
// =============================================================================

type RoleQuerySuite struct {
	suite.Suite
	svc      service.AuthzService
	ctx      context.Context
	tenantID string
	dom      string
	subj     string
}

func TestRoleQuerySuite(t *testing.T) { suite.Run(t, new(RoleQuerySuite)) }

func (s *RoleQuerySuite) SetupTest() {
	s.svc = newTestAuthzService(s.T(), &stubAuthzRepo{})
	s.ctx = context.Background()
	s.tenantID = uuid.New().String()
	s.dom = domain.TenantDomain(s.tenantID)
	s.subj = domain.TenantSubject(uuid.New().String())
}

func (s *RoleQuerySuite) assignRole(role string) {
	s.T().Helper()
	s.Require().NoError(s.svc.AssignRole(s.ctx, s.tenantID, s.subj, role, s.dom,
		domain.WithAssignedBy("platform:system"),
	))
}

func (s *RoleQuerySuite) TestGetRoles_Empty() {
	roles, err := s.svc.GetRoles(s.ctx, s.subj, s.dom)
	s.Require().NoError(err)
	s.Empty(roles)
}

func (s *RoleQuerySuite) TestGetRoles_AfterAssign() {
	s.assignRole("role:accountant")
	s.assignRole("role:manager")

	roles, err := s.svc.GetRoles(s.ctx, s.subj, s.dom)
	s.Require().NoError(err)
	s.ElementsMatch([]string{"role:accountant", "role:manager"}, roles)
}

func (s *RoleQuerySuite) TestHasRole_True() {
	s.assignRole("role:accountant")
	ok, err := s.svc.HasRole(s.ctx, s.subj, "role:accountant", s.dom)
	s.Require().NoError(err)
	s.True(ok)
}

func (s *RoleQuerySuite) TestHasRole_False() {
	ok, err := s.svc.HasRole(s.ctx, s.subj, "role:nonexistent", s.dom)
	s.Require().NoError(err)
	s.False(ok)
}

func (s *RoleQuerySuite) TestGetImplicitRoles_IncludesDirectRole() {
	s.assignRole("role:accountant")
	roles, err := s.svc.GetImplicitRoles(s.ctx, s.subj, s.dom)
	s.Require().NoError(err)
	s.Contains(roles, "role:accountant")
}

// =============================================================================
// AuthzService — AddPolicy Suite
// =============================================================================

type AddPolicySuite struct {
	suite.Suite
	svc  service.AuthzService
	ctx  context.Context
	dom  string
	role string
}

func TestAddPolicySuite(t *testing.T) { suite.Run(t, new(AddPolicySuite)) }

func (s *AddPolicySuite) SetupTest() {
	s.svc = newTestAuthzService(s.T(), &stubAuthzRepo{})
	s.ctx = context.Background()
	s.dom = domain.TenantDomain(uuid.New().String())
	s.role = "role:accountant"
}

func basePolicy(role, dom string) domain.Policy {
	return domain.Policy{Subject: role, Domain: dom, Object: "finance.accounts", Action: "read", Effect: "allow"}
}

func (s *AddPolicySuite) TestAddPolicy_MissingSubject() {
	p := basePolicy("", s.dom)
	err := s.svc.AddPolicy(s.ctx, p)
	s.True(errors.Is(err, domain.ErrInvalidRequest))
}

func (s *AddPolicySuite) TestAddPolicy_MissingDomain() {
	p := basePolicy(s.role, "")
	err := s.svc.AddPolicy(s.ctx, p)
	s.True(errors.Is(err, domain.ErrInvalidRequest))
}

func (s *AddPolicySuite) TestAddPolicy_InvalidEffect() {
	p := basePolicy(s.role, s.dom)
	p.Effect = "maybe"
	err := s.svc.AddPolicy(s.ctx, p)
	s.Require().Error(err)
	s.NotEqual(domain.ErrInvalidRequest, err)
}

func (s *AddPolicySuite) TestAddPolicy_PlatformDomainByTenantSubject_Forbidden() {
	p := domain.Policy{
		Subject: "tenant:user",
		Domain:  domain.DomainPlatform,
		Object:  "anything",
		Action:  "read",
		Effect:  "allow",
	}
	err := s.svc.AddPolicy(s.ctx, p)
	s.Require().Error(err)
	s.True(errors.Is(err, domain.ErrForbidden))
}

func (s *AddPolicySuite) TestAddPolicy_PlatformDomainByPlatformSubject_Allowed() {
	p := domain.Policy{
		Subject: "platform:admin",
		Domain:  domain.DomainPlatform,
		Object:  "anything",
		Action:  "read",
		Effect:  "allow",
	}
	s.Require().NoError(s.svc.AddPolicy(s.ctx, p))
}

func (s *AddPolicySuite) TestAddPolicy_PlatformDomainByRoleSubject_Allowed() {
	p := domain.Policy{
		Subject: "role:platform-admin",
		Domain:  domain.DomainPlatform,
		Object:  "anything",
		Action:  "read",
		Effect:  "allow",
	}
	s.Require().NoError(s.svc.AddPolicy(s.ctx, p))
}

func (s *AddPolicySuite) TestAddPolicy_Duplicate_ErrPolicyConflict() {
	p := basePolicy(s.role, s.dom)
	s.Require().NoError(s.svc.AddPolicy(s.ctx, p))

	err := s.svc.AddPolicy(s.ctx, p)
	s.Require().Error(err)
	s.True(errors.Is(err, domain.ErrPolicyConflict))
}

func (s *AddPolicySuite) TestAddPolicy_PolicyLimitExceeded() {
	svc := newTestAuthzServiceWithLimit(s.T(), &stubAuthzRepo{}, 2)

	dom := domain.TenantDomain(uuid.New().String())
	s.Require().NoError(svc.AddPolicy(s.ctx, domain.Policy{Subject: "role:a", Domain: dom, Object: "obj1", Action: "read", Effect: "allow"}))
	s.Require().NoError(svc.AddPolicy(s.ctx, domain.Policy{Subject: "role:a", Domain: dom, Object: "obj2", Action: "read", Effect: "allow"}))

	// Third policy in same domain must exceed limit
	err := svc.AddPolicy(s.ctx, domain.Policy{Subject: "role:a", Domain: dom, Object: "obj3", Action: "read", Effect: "allow"})
	s.Require().Error(err)
	s.True(errors.Is(err, domain.ErrPolicyLimitExceeded))
}

func (s *AddPolicySuite) TestAddPolicy_LimitPerDomain_NotCrossContaminated() {
	svc := newTestAuthzServiceWithLimit(s.T(), &stubAuthzRepo{}, 1)

	domA := domain.TenantDomain(uuid.New().String())
	domB := domain.TenantDomain(uuid.New().String())

	s.Require().NoError(svc.AddPolicy(s.ctx, domain.Policy{Subject: "role:a", Domain: domA, Object: "obj", Action: "read", Effect: "allow"}))
	// Different domain should have its own counter
	s.Require().NoError(svc.AddPolicy(s.ctx, domain.Policy{Subject: "role:a", Domain: domB, Object: "obj", Action: "read", Effect: "allow"}))
}

// =============================================================================
// AuthzService — GetPolicies / RemovePolicy Suite
// =============================================================================

type PolicyManagementSuite struct {
	suite.Suite
	svc service.AuthzService
	ctx context.Context
	dom string
}

func TestPolicyManagementSuite(t *testing.T) { suite.Run(t, new(PolicyManagementSuite)) }

func (s *PolicyManagementSuite) SetupTest() {
	s.svc = newTestAuthzService(s.T(), &stubAuthzRepo{})
	s.ctx = context.Background()
	s.dom = domain.TenantDomain(uuid.New().String())
}

func (s *PolicyManagementSuite) TestGetPolicies_Empty() {
	policies, err := s.svc.GetPolicies(s.ctx, s.dom)
	s.Require().NoError(err)
	s.Empty(policies)
}

func (s *PolicyManagementSuite) TestGetPolicies_AfterAdd() {
	p := domain.Policy{Subject: "role:a", Domain: s.dom, Object: "finance.accounts", Action: "read", Effect: "allow"}
	s.Require().NoError(s.svc.AddPolicy(s.ctx, p))

	policies, err := s.svc.GetPolicies(s.ctx, s.dom)
	s.Require().NoError(err)
	s.Require().Len(policies, 1)
	s.Equal("role:a", policies[0].Subject)
	s.Equal("read", policies[0].Action)
}

func (s *PolicyManagementSuite) TestRemovePolicy_PlatformByTenantSubject_Forbidden() {
	p := domain.Policy{Subject: "tenant:x", Domain: domain.DomainPlatform, Object: "obj", Action: "read", Effect: "allow"}
	err := s.svc.RemovePolicy(s.ctx, p)
	s.Require().Error(err)
	s.True(errors.Is(err, domain.ErrForbidden))
}

func (s *PolicyManagementSuite) TestRemovePolicy_Success() {
	p := domain.Policy{Subject: "role:a", Domain: s.dom, Object: "finance.accounts", Action: "read", Effect: "allow"}
	s.Require().NoError(s.svc.AddPolicy(s.ctx, p))

	s.Require().NoError(s.svc.RemovePolicy(s.ctx, p))

	policies, err := s.svc.GetPolicies(s.ctx, s.dom)
	s.Require().NoError(err)
	s.Empty(policies)
}

// =============================================================================
// AuthzService — Session Invalidation Suite
// =============================================================================

type SessionInvSuite struct {
	suite.Suite
	inv      *stubSessionInv
	svc      service.AuthzService
	ctx      context.Context
	tenantID string
	dom      string
}

func TestSessionInvSuite(t *testing.T) { suite.Run(t, new(SessionInvSuite)) }

func (s *SessionInvSuite) SetupTest() {
	s.inv = &stubSessionInv{}
	s.svc = newTestAuthzServiceWithInv(s.T(), &stubAuthzRepo{}, s.inv)
	s.ctx = context.Background()
	s.tenantID = uuid.New().String()
	s.dom = domain.TenantDomain(s.tenantID)
}

func (s *SessionInvSuite) TestRevokeRole_EvictsUserSession() {
	userID := uuid.New()
	subj := domain.TenantSubject(userID.String())
	role := "role:accountant"

	s.Require().NoError(s.svc.AssignRole(s.ctx, s.tenantID, subj, role, s.dom,
		domain.WithAssignedBy("platform:system"),
	))
	s.Require().NoError(s.svc.RevokeRole(s.ctx, subj, role, s.dom))

	s.Require().Len(s.inv.invalidateUserCalls, 1)
	s.Equal(userID, s.inv.invalidateUserCalls[0])
}

func (s *SessionInvSuite) TestRemovePolicy_EvictsTenantSessions() {
	tenantID, err := uuid.Parse(s.tenantID)
	s.Require().NoError(err)

	p := domain.Policy{Subject: "role:a", Domain: s.dom, Object: "iam.roles", Action: "read", Effect: "allow"}
	s.Require().NoError(s.svc.AddPolicy(s.ctx, p))
	s.Require().NoError(s.svc.RemovePolicy(s.ctx, p))

	s.Require().Len(s.inv.invalidateTenantCalls, 1)
	s.Equal(tenantID, s.inv.invalidateTenantCalls[0])
}

// =============================================================================
// AuthzService — BootstrapTenantAdmin Suite
// =============================================================================

type BootstrapSuite struct {
	suite.Suite
	svc service.AuthzService
	ctx context.Context
}

func TestBootstrapSuite(t *testing.T) { suite.Run(t, new(BootstrapSuite)) }

func (s *BootstrapSuite) SetupTest() {
	s.svc = newTestAuthzService(s.T(), &stubAuthzRepo{})
	s.ctx = context.Background()
}

func (s *BootstrapSuite) TestBootstrapTenantAdmin_CreatesAdminRole() {
	tenantID := uuid.New()
	userID := uuid.New()

	s.Require().NoError(s.svc.BootstrapTenantAdmin(s.ctx, tenantID, userID))

	// Subject should have the admin role in the tenant domain
	subj := domain.TenantSubject(userID.String())
	dom := domain.TenantDomain(tenantID.String())

	has, err := s.svc.HasRole(s.ctx, subj, "role:tenant.admin", dom)
	s.Require().NoError(err)
	s.True(has, "user should have role:tenant.admin after bootstrap")
}

func (s *BootstrapSuite) TestBootstrapTenantAdmin_Idempotent() {
	tenantID := uuid.New()
	userID := uuid.New()

	s.Require().NoError(s.svc.BootstrapTenantAdmin(s.ctx, tenantID, userID))
	// Second call must not error (duplicate policy/assignment silently skipped)
	s.Require().NoError(s.svc.BootstrapTenantAdmin(s.ctx, tenantID, userID))
}

func (s *BootstrapSuite) TestBootstrapTenantAdmin_AdminCanReadFinance() {
	tenantID := uuid.New()
	userID := uuid.New()
	s.Require().NoError(s.svc.BootstrapTenantAdmin(s.ctx, tenantID, userID))

	subj := domain.TenantSubject(userID.String())
	dom := domain.TenantDomain(tenantID.String())

	allowed, err := s.svc.Enforce(s.ctx, domain.Request{
		Subject: subj, Domain: dom, Object: "finance.accounts", Action: "read",
	})
	s.Require().NoError(err)
	s.True(allowed, "bootstrapped admin should be able to read finance.accounts")
}
