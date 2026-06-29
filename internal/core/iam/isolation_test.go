package iam

// isolation_test.go — AZ-ISO-001 through AZ-ISO-020.
//
// Cross-tenant isolation unit tests. All tests use the pure in-memory enforcer
// (newMemService / memRole) — no DATABASE_URL required.
//
// Invariant: a g-rule or p-rule scoped to domain D MUST NOT grant access,
// expose data, or mutate state in any other domain D'.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

// ===========================================================================
// IsolationSuite
// ===========================================================================

type IsolationSuite struct {
	suite.Suite
	svc Service
	ctx context.Context
}

func TestIsolationSuite(t *testing.T) { suite.Run(t, new(IsolationSuite)) }

func (s *IsolationSuite) SetupTest() {
	s.ctx = context.Background()
	s.svc = newMemService(s.T())
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func (s *IsolationSuite) allow(subject, domain, object, action string) {
	s.T().Helper()
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: subject, Domain: domain, Object: object, Action: action, Effect: "allow",
	}))
}

func (s *IsolationSuite) deny(subject, domain, object, action string) {
	s.T().Helper()
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: subject, Domain: domain, Object: object, Action: action, Effect: "deny",
	}))
}

func (s *IsolationSuite) enforce(subject, domain, object, action string) bool {
	s.T().Helper()
	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: subject, Domain: domain, Object: object, Action: action,
	})
	s.Require().NoError(err)
	return ok
}

// ---------------------------------------------------------------------------
// AZ-ISO-001: p-rule in tenantA domain does not grant access in tenantB domain
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_001_PolicyInTenantA_DeniedInTenantB() {
	domA := "tenant-aaaa-0001"
	domB := "tenant-bbbb-0002"

	s.allow("role:finance", domA, "invoice/*", "read")
	memRole(s.T(), s.svc, "tenant:usr_alice", "role:finance", domA)

	s.False(s.enforce("tenant:usr_alice", domB, "invoice/1", "read"),
		"p-rule in domA must not bleed into domB")
}

// ---------------------------------------------------------------------------
// AZ-ISO-002: g-rule in tenantA does not apply in tenantB
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_002_RoleInTenantA_NotInTenantB() {
	domA := "tenant-aaaa-0001"
	domB := "tenant-bbbb-0002"

	// Policy exists in both domains.
	s.allow("role:finance", domA, "invoice/*", "read")
	s.allow("role:finance", domB, "invoice/*", "read")

	// Role assigned only in domA.
	memRole(s.T(), s.svc, "tenant:usr_bob", "role:finance", domA)

	s.True(s.enforce("tenant:usr_bob", domA, "invoice/1", "read"), "should be allowed in domA")
	s.False(s.enforce("tenant:usr_bob", domB, "invoice/1", "read"),
		"g-rule in domA must not apply in domB")
}

// ---------------------------------------------------------------------------
// AZ-ISO-003: platform domain policy does not bleed into tenant domain
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_003_PlatformPolicy_NotInTenantDomain() {
	tenantDom := "tenant-cccc-0003"

	s.allow("role:platform-admin", DomainPlatform, "*", "*")
	memRole(s.T(), s.svc, "platform:usr_ops", "role:platform-admin", DomainPlatform)

	s.False(s.enforce("platform:usr_ops", tenantDom, "invoice/1", "read"),
		"platform-domain policy must not bleed into tenant domain")
}

// ---------------------------------------------------------------------------
// AZ-ISO-004: tenant policy does not bleed into platform domain
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_004_TenantPolicy_NotInPlatformDomain() {
	tenantDom := "tenant-dddd-0004"

	s.allow("role:finance", tenantDom, "*", "*")
	memRole(s.T(), s.svc, "tenant:usr_carol", "role:finance", tenantDom)

	s.False(s.enforce("tenant:usr_carol", DomainPlatform, "admin/settings", "write"),
		"tenant-domain policy must not bleed into platform domain")
}

// ---------------------------------------------------------------------------
// AZ-ISO-005: EnforceBatch requests isolated per domain
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_005_EnforceBatch_IsolatedPerDomain() {
	domA := "tenant-aaaa-0005a"
	domB := "tenant-bbbb-0005b"

	s.allow("role:r", domA, "res", "read")
	memRole(s.T(), s.svc, "tenant:usr", "role:r", domA)

	reqs := []Request{
		{Subject: "tenant:usr", Domain: domA, Object: "res", Action: "read"}, // true
		{Subject: "tenant:usr", Domain: domB, Object: "res", Action: "read"}, // false
	}
	results, err := s.svc.EnforceBatch(s.ctx, reqs)
	s.Require().NoError(err)
	s.Require().Len(results, 2)
	s.True(results[0], "allowed in domA")
	s.False(results[1], "must be denied in domB")
}

// ---------------------------------------------------------------------------
// AZ-ISO-006: GetPolicies scoped to requested domain only
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_006_GetPolicies_ScopedToDomain() {
	domA := "tenant-aaaa-0006a"
	domB := "tenant-bbbb-0006b"

	s.allow("role:a", domA, "invoice/*", "read")
	s.allow("role:b", domB, "payment/*", "write")

	pols, err := s.svc.GetPolicies(s.ctx, domA)
	s.Require().NoError(err)
	for _, p := range pols {
		s.Equal(domA, p.Domain, "GetPolicies must return only domA policies")
	}
	s.Len(pols, 1)
}

// ---------------------------------------------------------------------------
// AZ-ISO-007: GetRoles scoped to requested domain only
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_007_GetRoles_ScopedToDomain() {
	domA := "tenant-aaaa-0007a"
	domB := "tenant-bbbb-0007b"

	memRole(s.T(), s.svc, "tenant:usr", "role:finance", domA)
	memRole(s.T(), s.svc, "tenant:usr", "role:sales", domB)

	rolesA, err := s.svc.GetRoles(s.ctx, "tenant:usr", domA)
	s.Require().NoError(err)
	s.ElementsMatch([]string{"role:finance"}, rolesA)

	rolesB, err := s.svc.GetRoles(s.ctx, "tenant:usr", domB)
	s.Require().NoError(err)
	s.ElementsMatch([]string{"role:sales"}, rolesB)
}

// ---------------------------------------------------------------------------
// AZ-ISO-008: HasRole is domain-scoped
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_008_HasRole_DomainScoped() {
	domA := "tenant-aaaa-0008a"
	domB := "tenant-bbbb-0008b"

	memRole(s.T(), s.svc, "tenant:usr", "role:admin", domA)

	hasA, err := s.svc.HasRole(s.ctx, "tenant:usr", "role:admin", domA)
	s.Require().NoError(err)
	s.True(hasA)

	hasB, err := s.svc.HasRole(s.ctx, "tenant:usr", "role:admin", domB)
	s.Require().NoError(err)
	s.False(hasB, "HasRole must return false for a different domain")
}

// ---------------------------------------------------------------------------
// AZ-ISO-009: explicit deny in tenantA does not affect tenantB
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_009_DenyInTenantA_NotInTenantB() {
	domA := "tenant-aaaa-0009a"
	domB := "tenant-bbbb-0009b"

	// User allowed via role in both domains.
	s.allow("role:finance", domA, "invoice/*", "read")
	s.allow("role:finance", domB, "invoice/*", "read")
	memRole(s.T(), s.svc, "tenant:usr_dave", "role:finance", domA)
	memRole(s.T(), s.svc, "tenant:usr_dave", "role:finance", domB)

	// Explicit deny only in domA.
	s.deny("tenant:usr_dave", domA, "*", "*")

	s.False(s.enforce("tenant:usr_dave", domA, "invoice/1", "read"), "denied in domA")
	s.True(s.enforce("tenant:usr_dave", domB, "invoice/1", "read"),
		"deny in domA must not affect domB")
}

// ---------------------------------------------------------------------------
// AZ-ISO-010: wildcard subject policy does not cross domains
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_010_WildcardSubject_ScopedToDomain() {
	domA := "tenant-aaaa-0010a"
	domB := "tenant-bbbb-0010b"

	// Wildcard allow in domA only.
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "*", Domain: domA, Object: "report/*", Action: "read", Effect: "allow",
	}))

	s.True(s.enforce("tenant:any_user", domA, "report/q1", "read"),
		"wildcard should match in domA")
	s.False(s.enforce("tenant:any_user", domB, "report/q1", "read"),
		"wildcard policy in domA must not apply in domB")
}

// ---------------------------------------------------------------------------
// AZ-ISO-011: role hierarchy does not cross domains
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_011_RoleHierarchy_NotCrossDomain() {
	domA := "tenant-aaaa-0011a"
	domB := "tenant-bbbb-0011b"

	// In domA: role:viewer allows read, role:admin inherits role:viewer.
	s.allow("role:viewer", domA, "report/*", "read")
	memRole(s.T(), s.svc, "tenant:usr_eve", "role:viewer", domA)

	// Same role in domB has no policy assigned.
	// User has no g-rule in domB.

	s.True(s.enforce("tenant:usr_eve", domA, "report/q1", "read"), "allowed via role in domA")
	s.False(s.enforce("tenant:usr_eve", domB, "report/q1", "read"),
		"role hierarchy in domA must not cross into domB")
}

// ---------------------------------------------------------------------------
// AZ-ISO-012: RemovePolicy in tenantA does not affect tenantB
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_012_RemovePolicy_IsolatedToDomain() {
	domA := "tenant-aaaa-0012a"
	domB := "tenant-bbbb-0012b"

	polA := Policy{Subject: "role:finance", Domain: domA, Object: "invoice/*", Action: "read", Effect: "allow"}
	polB := Policy{Subject: "role:finance", Domain: domB, Object: "invoice/*", Action: "read", Effect: "allow"}

	s.Require().NoError(s.svc.AddPolicy(s.ctx, polA))
	s.Require().NoError(s.svc.AddPolicy(s.ctx, polB))
	memRole(s.T(), s.svc, "tenant:usr", "role:finance", domA)
	memRole(s.T(), s.svc, "tenant:usr", "role:finance", domB)

	// Remove from domA only.
	s.Require().NoError(s.svc.RemovePolicy(s.ctx, polA))

	s.False(s.enforce("tenant:usr", domA, "invoice/1", "read"), "policy removed from domA")
	s.True(s.enforce("tenant:usr", domB, "invoice/1", "read"),
		"RemovePolicy in domA must not affect domB")
}

// ---------------------------------------------------------------------------
// AZ-ISO-013: subject with wrong prefix type denied in correct domain
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_013_WrongSubjectPrefix_DeniedInDomain() {
	dom := "tenant-aaaa-0013"

	// Policy for tenant: subject.
	s.allow("role:finance", dom, "invoice/*", "read")
	memRole(s.T(), s.svc, "tenant:usr_frank", "role:finance", dom)

	// platform: subject presenting in tenant domain — no g-rule exists for it.
	s.False(s.enforce("platform:usr_frank", dom, "invoice/1", "read"),
		"platform: subject must not match tenant: g-rule in same domain")
}

// ---------------------------------------------------------------------------
// AZ-ISO-014: portal domain isolated from tenant domain
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_014_PortalDomain_IsolatedFromTenantDomain() {
	portalDom := "portal"
	tenantDom := "tenant-cccc-0014"

	s.allow("role:portal-user", portalDom, "dashboard/*", "read")
	memRole(s.T(), s.svc, "portal:usr_grace", "role:portal-user", portalDom)

	s.True(s.enforce("portal:usr_grace", portalDom, "dashboard/home", "read"),
		"portal user allowed in portal domain")
	s.False(s.enforce("portal:usr_grace", tenantDom, "dashboard/home", "read"),
		"portal domain policy must not bleed into tenant domain")
}

// ---------------------------------------------------------------------------
// AZ-ISO-015: API key subject isolated from tenant domain
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_015_APIKeySubject_IsolatedFromTenantDomain() {
	apiDom := "api-key-domain-0015"
	tenantDom := "tenant-eeee-0015"

	s.allow("role:integrator", apiDom, "webhook/*", "create")
	memRole(s.T(), s.svc, "api:key_abc123", "role:integrator", apiDom)

	s.True(s.enforce("api:key_abc123", apiDom, "webhook/events", "create"),
		"api key allowed in its own domain")
	s.False(s.enforce("api:key_abc123", tenantDom, "invoice/1", "read"),
		"api key must not bleed into tenant domain")
}

// ---------------------------------------------------------------------------
// AZ-ISO-016: platform role cannot grant access in tenant domain
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_016_PlatformRole_NoGrantInTenantDomain() {
	tenantDom := "tenant-ffff-0016"

	// Platform admin role exists with broad platform policy.
	s.allow("role:platform-admin", DomainPlatform, "*", "*")
	memRole(s.T(), s.svc, "platform:usr_hector", "role:platform-admin", DomainPlatform)

	// Verify platform role grants access in platform domain.
	s.True(s.enforce("platform:usr_hector", DomainPlatform, "anything", "delete"),
		"platform admin allowed in platform domain")

	// Same user must not gain access in tenant domain.
	s.False(s.enforce("platform:usr_hector", tenantDom, "invoice/1", "delete"),
		"platform role must not grant access in tenant domain")
}

// ---------------------------------------------------------------------------
// AZ-ISO-017: multiple tenants simultaneously isolated
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_017_MultipleTenants_SimultaneouslyIsolated() {
	tenants := []struct {
		domain  string
		subject string
		role    string
		object  string
	}{
		{"tenant-t1-0017", "tenant:usr_1", "role:finance", "ledger/1"},
		{"tenant-t2-0017", "tenant:usr_2", "role:sales", "order/2"},
		{"tenant-t3-0017", "tenant:usr_3", "role:hr", "employee/3"},
	}

	for _, tc := range tenants {
		s.allow(tc.role, tc.domain, tc.object, "read")
		memRole(s.T(), s.svc, tc.subject, tc.role, tc.domain)
	}

	// Each user can access their own resource.
	for _, tc := range tenants {
		s.True(s.enforce(tc.subject, tc.domain, tc.object, "read"),
			"%s must be allowed in their own domain", tc.subject)
	}

	// No user can access another tenant's resource.
	for i, tc := range tenants {
		for j, other := range tenants {
			if i == j {
				continue
			}
			s.False(s.enforce(tc.subject, other.domain, other.object, "read"),
				"%s must not access %s's resource in %s", tc.subject, other.subject, other.domain)
		}
	}
}

// ---------------------------------------------------------------------------
// AZ-ISO-018: wildcard action policy does not cross domains
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_018_WildcardAction_ScopedToDomain() {
	domA := "tenant-aaaa-0018a"
	domB := "tenant-bbbb-0018b"

	// Wildcard action in domA only.
	s.allow("role:superuser", domA, "invoice/*", "*")
	memRole(s.T(), s.svc, "tenant:usr_ivan", "role:superuser", domA)

	for _, action := range []string{"read", "write", "delete", "approve"} {
		s.True(s.enforce("tenant:usr_ivan", domA, "invoice/1", action),
			"wildcard action must cover %s in domA", action)
		s.False(s.enforce("tenant:usr_ivan", domB, "invoice/1", action),
			"wildcard action in domA must not apply in domB for action %s", action)
	}
}

// ---------------------------------------------------------------------------
// AZ-ISO-019: AssignRole in one domain does not affect GetRoles in another
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_019_AssignRole_IsolatedToTargetDomain() {
	domA := "tenant-aaaa-0019a"
	domB := "tenant-bbbb-0019b"

	memRole(s.T(), s.svc, "tenant:usr_judy", "role:manager", domA)

	rolesA, err := s.svc.GetRoles(s.ctx, "tenant:usr_judy", domA)
	s.Require().NoError(err)
	s.Contains(rolesA, "role:manager")

	rolesB, err := s.svc.GetRoles(s.ctx, "tenant:usr_judy", domB)
	s.Require().NoError(err)
	s.Empty(rolesB, "AssignRole in domA must not appear in GetRoles for domB")
}

// ---------------------------------------------------------------------------
// AZ-ISO-020: empty domain returns ErrInvalidRequest (defense-in-depth)
// ---------------------------------------------------------------------------

func (s *IsolationSuite) TestAZ_ISO_020_EmptyDomain_ReturnsErrInvalidRequest() {
	// Enforce with empty domain must be rejected before any Casbin lookup.
	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: "tenant:usr_kate", Domain: "", Object: "invoice/1", Action: "read",
	})
	s.ErrorIs(err, ErrInvalidRequest, "empty domain must return ErrInvalidRequest")
	s.False(ok)

	// AddPolicy with empty domain must also be rejected.
	addErr := s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:x", Domain: "", Object: "obj", Action: "read", Effect: "allow",
	})
	s.ErrorIs(addErr, ErrInvalidRequest, "AddPolicy with empty domain must return ErrInvalidRequest")

	// GetRoles with empty domain — service may return empty or error; must not panic.
	_, _ = s.svc.GetRoles(s.ctx, "tenant:usr_kate", "")
}
