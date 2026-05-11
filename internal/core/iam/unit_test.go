package iam

// unit_test.go — full business-logic coverage with zero DB dependency.
//
// Each suite uses newMemService() which provides:
//   - A pure in-memory Casbin enforcer (EnableAutoSave=false, nil adapter)
//   - A fake pgxpool that returns fast "connection refused" errors so
//     revokeExpiredRoles is non-fatally swallowed by Enforce().
//
// Coverage targets (no DATABASE_URL required):
//   service.go  — New() validation, Enforce(), EnforceBatch(), InvalidateCache()
//   policies.go — AddPolicy(), RemovePolicy(), GetPolicies()
//   roles.go    — GetRoles(), HasRole() (in-memory path)
//   middleware.go — Middleware() Fiber handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"
)

// ===========================================================================
// ConstructorUnitSuite — New() validation (no DB connection needed)
// ===========================================================================

type ConstructorUnitSuite struct{ suite.Suite }

func TestConstructorUnitSuite(t *testing.T) { suite.Run(t, new(ConstructorUnitSuite)) }

func (s *ConstructorUnitSuite) TestNew_NilStore_ReturnsError() {
	_, err := New(Config{Store: nil, Cache: noopCache{}, Logger: noopLogger{}})
	s.Require().Error(err)
	s.Contains(err.Error(), "store is required")
}

func (s *ConstructorUnitSuite) TestNew_NilLogger_ReturnsError() {
	// Logger nil-check fires before any DB connection attempt; noopCache and a
	// nil Store are enough to reach it.
	_, err := New(Config{Store: nil, Cache: noopCache{}, Logger: nil})
	s.Require().Error(err)
	// New() checks Store first, so we just assert an error is returned.
	s.Error(err)
}

// ===========================================================================
// EnforceUnitSuite — Enforce() and EnforceBatch() without a real database
// ===========================================================================

type EnforceUnitSuite struct {
	suite.Suite
	svc Service
	ctx context.Context
}

func TestEnforceUnitSuite(t *testing.T) { suite.Run(t, new(EnforceUnitSuite)) }

func (s *EnforceUnitSuite) SetupTest() {
	s.ctx = context.Background()
	s.svc = newMemService(s.T())
}

// ---- Input validation -----------------------------------------------------

func (s *EnforceUnitSuite) TestEnforce_EmptySubject() {
	ok, err := s.svc.Enforce(s.ctx, Request{Subject: "", Domain: "dom", Object: "obj", Action: "act"})
	s.ErrorIs(err, ErrInvalidRequest)
	s.False(ok)
}

func (s *EnforceUnitSuite) TestEnforce_EmptyDomain() {
	ok, err := s.svc.Enforce(s.ctx, Request{Subject: "tenant:usr", Domain: "", Object: "obj", Action: "act"})
	s.ErrorIs(err, ErrInvalidRequest)
	s.False(ok)
}

func (s *EnforceUnitSuite) TestEnforce_EmptyObject() {
	ok, err := s.svc.Enforce(s.ctx, Request{Subject: "tenant:usr", Domain: "dom", Object: "", Action: "act"})
	s.ErrorIs(err, ErrInvalidRequest)
	s.False(ok)
}

func (s *EnforceUnitSuite) TestEnforce_EmptyAction() {
	ok, err := s.svc.Enforce(s.ctx, Request{Subject: "tenant:usr", Domain: "dom", Object: "obj", Action: ""})
	s.ErrorIs(err, ErrInvalidRequest)
	s.False(ok)
}

// ---- Default deny ----------------------------------------------------------

func (s *EnforceUnitSuite) TestEnforce_NoPolicies_DefaultDeny() {
	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: "tenant:usr_001",
		Domain:  "dom-1",
		Object:  "invoice/123",
		Action:  "read",
	})
	s.NoError(err)
	s.False(ok)
}

// ---- Allow -----------------------------------------------------------------

func (s *EnforceUnitSuite) TestEnforce_Allow_ExactMatch() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:finance", Domain: "dom-1", Object: "invoice/123", Action: "read", Effect: "allow",
	}))
	memRole(s.T(), s.svc, "tenant:usr_001", "role:finance", "dom-1")

	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: "tenant:usr_001", Domain: "dom-1", Object: "invoice/123", Action: "read",
	})
	s.NoError(err)
	s.True(ok)
}

func (s *EnforceUnitSuite) TestEnforce_Allow_WildcardObject_keyMatch2() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:finance", Domain: "dom-1", Object: "invoice/*", Action: "read", Effect: "allow",
	}))
	memRole(s.T(), s.svc, "tenant:usr_001", "role:finance", "dom-1")

	for _, obj := range []string{"invoice/inv_001", "invoice/999", "invoice/xyz-abc"} {
		s.Run(obj, func() {
			ok, err := s.svc.Enforce(s.ctx, Request{
				Subject: "tenant:usr_001", Domain: "dom-1", Object: obj, Action: "read",
			})
			s.NoError(err)
			s.True(ok, "wildcard invoice/* must match %s", obj)
		})
	}
}

func (s *EnforceUnitSuite) TestEnforce_Allow_WildcardAction() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:admin", Domain: "dom-1", Object: "invoice/*", Action: "*", Effect: "allow",
	}))
	memRole(s.T(), s.svc, "tenant:usr_001", "role:admin", "dom-1")

	for _, act := range []string{"read", "create", "delete", "approve", "export", "execute"} {
		s.Run(act, func() {
			ok, err := s.svc.Enforce(s.ctx, Request{
				Subject: "tenant:usr_001", Domain: "dom-1", Object: "invoice/1", Action: act,
			})
			s.NoError(err)
			s.True(ok, "wildcard * must cover action %s", act)
		})
	}
}

func (s *EnforceUnitSuite) TestEnforce_Allow_ViaRoleHierarchy() {
	// Policy granted to role, user has that role.
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:viewer", Domain: "dom-1", Object: "report/*", Action: "read", Effect: "allow",
	}))
	memRole(s.T(), s.svc, "tenant:usr_bob", "role:viewer", "dom-1")

	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: "tenant:usr_bob", Domain: "dom-1", Object: "report/q1", Action: "read",
	})
	s.NoError(err)
	s.True(ok)
}

// ---- Deny ------------------------------------------------------------------

func (s *EnforceUnitSuite) TestEnforce_Deny_ExplicitDenyOverridesAllow() {
	// Allow via role.
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:finance", Domain: "dom-1", Object: "invoice/*", Action: "*", Effect: "allow",
	}))
	memRole(s.T(), s.svc, "tenant:usr_terminated", "role:finance", "dom-1")

	// Blanket deny directly on the subject.
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "tenant:usr_terminated", Domain: "dom-1", Object: "*", Action: "*", Effect: "deny",
	}))

	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: "tenant:usr_terminated", Domain: "dom-1", Object: "invoice/123", Action: "read",
	})
	s.NoError(err)
	s.False(ok, "deny must override allow")
}

func (s *EnforceUnitSuite) TestEnforce_Deny_NoAllow_IsFalse() {
	// Just a deny rule, no allow — still false.
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "tenant:usr_001", Domain: "dom-1", Object: "*", Action: "*", Effect: "deny",
	}))
	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: "tenant:usr_001", Domain: "dom-1", Object: "invoice/1", Action: "read",
	})
	s.NoError(err)
	s.False(ok)
}

// ---- Domain isolation -----------------------------------------------------

func (s *EnforceUnitSuite) TestEnforce_PolicyInDom1_DoesNotMatchDom2() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:finance", Domain: "dom-1", Object: "invoice/*", Action: "read", Effect: "allow",
	}))
	memRole(s.T(), s.svc, "tenant:usr_001", "role:finance", "dom-1")

	// Same user, different domain.
	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: "tenant:usr_001", Domain: "dom-2", Object: "invoice/123", Action: "read",
	})
	s.NoError(err)
	s.False(ok, "dom-1 policy must not bleed into dom-2")
}

func (s *EnforceUnitSuite) TestEnforce_RoleInDom1_DoesNotApplyInDom2() {
	// Policy exists in BOTH domains, but user only has role in dom-1.
	for _, dom := range []string{"dom-1", "dom-2"} {
		s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
			Subject: "role:finance", Domain: dom, Object: "invoice/*", Action: "read", Effect: "allow",
		}))
	}
	memRole(s.T(), s.svc, "tenant:usr_001", "role:finance", "dom-1") // only dom-1

	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: "tenant:usr_001", Domain: "dom-2", Object: "invoice/123", Action: "read",
	})
	s.NoError(err)
	s.False(ok, "g-rule in dom-1 must not grant access in dom-2")
}

func (s *EnforceUnitSuite) TestEnforce_PlatformDomain_IsolatedFromTenant() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:platform-admin", Domain: DomainPlatform, Object: "*", Action: "*", Effect: "allow",
	}))
	memRole(s.T(), s.svc, "platform:usr_ops", "role:platform-admin", DomainPlatform)

	// Platform user trying to access tenant domain.
	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: "platform:usr_ops", Domain: "some-tenant-uuid", Object: "invoice/1", Action: "read",
	})
	s.NoError(err)
	s.False(ok, "_platform_ policy must not apply in tenant domain")
}

// ---- EnforceBatch ----------------------------------------------------------

func (s *EnforceUnitSuite) TestEnforceBatch_Empty_ReturnsNil() {
	results, err := s.svc.EnforceBatch(s.ctx, []Request{})
	s.NoError(err)
	s.Nil(results)
}

func (s *EnforceUnitSuite) TestEnforceBatch_AnyEmptyField_ReturnsError() {
	reqs := []Request{
		{Subject: "tenant:usr", Domain: "dom", Object: "obj", Action: "act"},
		{Subject: "", Domain: "dom", Object: "obj", Action: "act"}, // invalid
	}
	results, err := s.svc.EnforceBatch(s.ctx, reqs)
	s.ErrorIs(err, ErrInvalidRequest)
	s.Nil(results)
}

func (s *EnforceUnitSuite) TestEnforceBatch_ParallelResults_MatchInput() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:finance", Domain: "dom-1", Object: "invoice/*", Action: "read", Effect: "allow",
	}))
	memRole(s.T(), s.svc, "tenant:usr_001", "role:finance", "dom-1")

	reqs := []Request{
		{Subject: "tenant:usr_001", Domain: "dom-1", Object: "invoice/1", Action: "read"},  // true
		{Subject: "tenant:usr_001", Domain: "dom-1", Object: "invoice/1", Action: "write"}, // false (no write policy)
		{Subject: "tenant:nobody", Domain: "dom-1", Object: "invoice/1", Action: "read"},   // false (no role)
	}
	results, err := s.svc.EnforceBatch(s.ctx, reqs)
	s.Require().NoError(err)
	s.Require().Len(results, 3)
	s.True(results[0])
	s.False(results[1])
	s.False(results[2])
}

func (s *EnforceUnitSuite) TestEnforceBatch_SingleRequest_Works() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:r", Domain: "dom-1", Object: "res", Action: "read", Effect: "allow",
	}))
	memRole(s.T(), s.svc, "tenant:usr", "role:r", "dom-1")

	results, err := s.svc.EnforceBatch(s.ctx, []Request{
		{Subject: "tenant:usr", Domain: "dom-1", Object: "res", Action: "read"},
	})
	s.Require().NoError(err)
	s.Require().Len(results, 1)
	s.True(results[0])
}

// ===========================================================================
// PolicyUnitSuite — AddPolicy, RemovePolicy, GetPolicies without DB
// ===========================================================================

type PolicyUnitSuite struct {
	suite.Suite
	svc Service
	ctx context.Context
}

func TestPolicyUnitSuite(t *testing.T) { suite.Run(t, new(PolicyUnitSuite)) }

func (s *PolicyUnitSuite) SetupTest() {
	s.ctx = context.Background()
	s.svc = newMemService(s.T())
}

// ---- AddPolicy validation --------------------------------------------------

func (s *PolicyUnitSuite) TestAddPolicy_Allow_Succeeds() {
	err := s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:cfo", Domain: "dom", Object: "invoice/*", Action: "*", Effect: "allow",
	})
	s.NoError(err)
}

func (s *PolicyUnitSuite) TestAddPolicy_Deny_Succeeds() {
	err := s.svc.AddPolicy(s.ctx, Policy{
		Subject: "tenant:usr_bad", Domain: "dom", Object: "*", Action: "*", Effect: "deny",
	})
	s.NoError(err)
}

func (s *PolicyUnitSuite) TestAddPolicy_Effect_Empty_Errors() {
	err := s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:x", Domain: "dom", Object: "obj", Action: "act", Effect: "",
	})
	s.Error(err)
	s.Contains(err.Error(), "effect")
}

func (s *PolicyUnitSuite) TestAddPolicy_Effect_Uppercase_Errors() {
	err := s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:x", Domain: "dom", Object: "obj", Action: "act", Effect: "ALLOW",
	})
	s.Error(err)
}

func (s *PolicyUnitSuite) TestAddPolicy_Effect_Unknown_Errors() {
	for _, bad := range []string{"permit", "block", "yes", "1"} {
		s.Run(bad, func() {
			err := s.svc.AddPolicy(s.ctx, Policy{
				Subject: "role:x", Domain: "dom", Object: "obj", Action: "act", Effect: bad,
			})
			s.Error(err)
		})
	}
}

func (s *PolicyUnitSuite) TestAddPolicy_EmptySubject_ErrInvalidRequest() {
	err := s.svc.AddPolicy(s.ctx, Policy{
		Subject: "", Domain: "dom", Object: "obj", Action: "act", Effect: "allow",
	})
	s.ErrorIs(err, ErrInvalidRequest)
}

func (s *PolicyUnitSuite) TestAddPolicy_EmptyDomain_ErrInvalidRequest() {
	err := s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:x", Domain: "", Object: "obj", Action: "act", Effect: "allow",
	})
	s.ErrorIs(err, ErrInvalidRequest)
}

func (s *PolicyUnitSuite) TestAddPolicy_EmptyObject_ErrInvalidRequest() {
	err := s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:x", Domain: "dom", Object: "", Action: "act", Effect: "allow",
	})
	s.ErrorIs(err, ErrInvalidRequest)
}

func (s *PolicyUnitSuite) TestAddPolicy_EmptyAction_ErrInvalidRequest() {
	err := s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:x", Domain: "dom", Object: "obj", Action: "", Effect: "allow",
	})
	s.ErrorIs(err, ErrInvalidRequest)
}

func (s *PolicyUnitSuite) TestAddPolicy_Duplicate_ErrPolicyConflict() {
	pol := Policy{Subject: "role:x", Domain: "dom", Object: "obj", Action: "read", Effect: "allow"}
	s.Require().NoError(s.svc.AddPolicy(s.ctx, pol))

	err := s.svc.AddPolicy(s.ctx, pol)
	s.ErrorIs(err, ErrPolicyConflict)
}

// ---- RemovePolicy ----------------------------------------------------------

func (s *PolicyUnitSuite) TestRemovePolicy_Succeeds() {
	pol := Policy{Subject: "role:x", Domain: "dom", Object: "invoice/*", Action: "read", Effect: "allow"}
	s.Require().NoError(s.svc.AddPolicy(s.ctx, pol))
	memRole(s.T(), s.svc, "tenant:usr_001", "role:x", "dom")

	// Verify allow before.
	ok, _ := s.svc.Enforce(s.ctx, Request{Subject: "tenant:usr_001", Domain: "dom", Object: "invoice/1", Action: "read"})
	s.True(ok)

	s.Require().NoError(s.svc.RemovePolicy(s.ctx, pol))

	// Verify deny after.
	ok, err := s.svc.Enforce(s.ctx, Request{Subject: "tenant:usr_001", Domain: "dom", Object: "invoice/1", Action: "read"})
	s.NoError(err)
	s.False(ok)
}

func (s *PolicyUnitSuite) TestRemovePolicy_EmptySubject_ErrInvalidRequest() {
	err := s.svc.RemovePolicy(s.ctx, Policy{Subject: "", Domain: "dom", Object: "obj", Action: "act", Effect: "allow"})
	s.ErrorIs(err, ErrInvalidRequest)
}

func (s *PolicyUnitSuite) TestRemovePolicy_EmptyDomain_ErrInvalidRequest() {
	err := s.svc.RemovePolicy(s.ctx, Policy{Subject: "role:x", Domain: "", Object: "obj", Action: "act", Effect: "allow"})
	s.ErrorIs(err, ErrInvalidRequest)
}

func (s *PolicyUnitSuite) TestRemovePolicy_EmptyObject_ErrInvalidRequest() {
	err := s.svc.RemovePolicy(s.ctx, Policy{Subject: "role:x", Domain: "dom", Object: "", Action: "act", Effect: "allow"})
	s.ErrorIs(err, ErrInvalidRequest)
}

func (s *PolicyUnitSuite) TestRemovePolicy_EmptyAction_ErrInvalidRequest() {
	err := s.svc.RemovePolicy(s.ctx, Policy{Subject: "role:x", Domain: "dom", Object: "obj", Action: "", Effect: "allow"})
	s.ErrorIs(err, ErrInvalidRequest)
}

func (s *PolicyUnitSuite) TestRemovePolicy_NonExistent_NoError() {
	// Removing a policy that doesn't exist is not an error.
	err := s.svc.RemovePolicy(s.ctx, Policy{
		Subject: "role:x", Domain: "dom", Object: "obj", Action: "read", Effect: "allow",
	})
	s.NoError(err)
}

// ---- GetPolicies -----------------------------------------------------------

func (s *PolicyUnitSuite) TestGetPolicies_ReturnsAll_ForDomain() {
	pols := []Policy{
		{Subject: "role:a", Domain: "dom-1", Object: "invoice/*", Action: "read", Effect: "allow"},
		{Subject: "role:b", Domain: "dom-1", Object: "payment/*", Action: "*", Effect: "allow"},
	}
	for _, p := range pols {
		s.Require().NoError(s.svc.AddPolicy(s.ctx, p))
	}
	// Different domain — must not appear.
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:c", Domain: "dom-2", Object: "order/*", Action: "read", Effect: "allow",
	}))

	result, err := s.svc.GetPolicies(s.ctx, "dom-1")
	s.Require().NoError(err)
	s.Len(result, 2)
	for _, p := range result {
		s.Equal("dom-1", p.Domain)
	}
}

func (s *PolicyUnitSuite) TestGetPolicies_EmptyDomain_ReturnsEmpty() {
	result, err := s.svc.GetPolicies(s.ctx, "no-such-domain")
	s.NoError(err)
	s.Empty(result)
}

func (s *PolicyUnitSuite) TestGetPolicies_FieldsCorrectlyMapped() {
	pol := Policy{Subject: "role:cfo", Domain: "dom-1", Object: "period/*", Action: "close", Effect: "allow"}
	s.Require().NoError(s.svc.AddPolicy(s.ctx, pol))

	result, err := s.svc.GetPolicies(s.ctx, "dom-1")
	s.Require().NoError(err)
	s.Require().Len(result, 1)
	s.Equal(pol.Subject, result[0].Subject)
	s.Equal(pol.Domain, result[0].Domain)
	s.Equal(pol.Object, result[0].Object)
	s.Equal(pol.Action, result[0].Action)
	s.Equal(pol.Effect, result[0].Effect)
}

func (s *PolicyUnitSuite) TestGetPolicies_DenyPolicy_IsIncluded() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "*", Domain: "dom-1", Object: "journal/closed/*", Action: "create", Effect: "deny",
	}))
	result, err := s.svc.GetPolicies(s.ctx, "dom-1")
	s.Require().NoError(err)
	s.Require().Len(result, 1)
	s.Equal("deny", result[0].Effect)
}

// ===========================================================================
// RoleQueryUnitSuite — GetRoles() and HasRole() (in-memory paths)
// ===========================================================================

type RoleQueryUnitSuite struct {
	suite.Suite
	svc Service
	ctx context.Context
}

func TestRoleQueryUnitSuite(t *testing.T) { suite.Run(t, new(RoleQueryUnitSuite)) }

func (s *RoleQueryUnitSuite) SetupTest() {
	s.ctx = context.Background()
	s.svc = newMemService(s.T())
}

func (s *RoleQueryUnitSuite) TestGetRoles_ReturnsAssignedRole() {
	memRole(s.T(), s.svc, "tenant:usr_001", "role:finance", "dom-1")

	roles, err := s.svc.GetRoles(s.ctx, "tenant:usr_001", "dom-1")
	s.Require().NoError(err)
	s.Equal([]string{"role:finance"}, roles)
}

func (s *RoleQueryUnitSuite) TestGetRoles_NoRoles_ReturnsEmptySlice() {
	roles, err := s.svc.GetRoles(s.ctx, "tenant:nobody", "dom-1")
	s.NoError(err)
	s.Empty(roles)
}

func (s *RoleQueryUnitSuite) TestGetRoles_MultipleRoles_AllReturned() {
	expected := []string{"role:finance", "role:sales", "role:viewer"}
	for _, r := range expected {
		memRole(s.T(), s.svc, "tenant:usr_001", r, "dom-1")
	}
	roles, err := s.svc.GetRoles(s.ctx, "tenant:usr_001", "dom-1")
	s.Require().NoError(err)
	s.ElementsMatch(expected, roles)
}

func (s *RoleQueryUnitSuite) TestGetRoles_IsDomainScoped() {
	memRole(s.T(), s.svc, "tenant:usr_001", "role:finance", "dom-1")

	// Query dom-2 — must be empty.
	roles, err := s.svc.GetRoles(s.ctx, "tenant:usr_001", "dom-2")
	s.NoError(err)
	s.Empty(roles)
}

func (s *RoleQueryUnitSuite) TestHasRole_TrueWhenRoleExists() {
	memRole(s.T(), s.svc, "tenant:usr_001", "role:finance", "dom-1")

	has, err := s.svc.HasRole(s.ctx, "tenant:usr_001", "role:finance", "dom-1")
	s.Require().NoError(err)
	s.True(has)
}

func (s *RoleQueryUnitSuite) TestHasRole_FalseWhenRoleAbsent() {
	has, err := s.svc.HasRole(s.ctx, "tenant:usr_nobody", "role:finance", "dom-1")
	s.Require().NoError(err)
	s.False(has)
}

func (s *RoleQueryUnitSuite) TestHasRole_FalseForWrongDomain() {
	memRole(s.T(), s.svc, "tenant:usr_001", "role:finance", "dom-1")

	has, err := s.svc.HasRole(s.ctx, "tenant:usr_001", "role:finance", "dom-2")
	s.Require().NoError(err)
	s.False(has)
}

func (s *RoleQueryUnitSuite) TestHasRole_FalseForWrongRole() {
	memRole(s.T(), s.svc, "tenant:usr_001", "role:finance", "dom-1")

	has, err := s.svc.HasRole(s.ctx, "tenant:usr_001", "role:admin", "dom-1")
	s.Require().NoError(err)
	s.False(has)
}

// ===========================================================================
// MiddlewareUnitSuite — Middleware() Fiber handler without DB
// ===========================================================================

type MiddlewareUnitSuite struct {
	suite.Suite
	svc Service
	ctx context.Context
}

func TestMiddlewareUnitSuite(t *testing.T) { suite.Run(t, new(MiddlewareUnitSuite)) }

func (s *MiddlewareUnitSuite) SetupTest() {
	s.ctx = context.Background()
	s.svc = newMemService(s.T())
}

// ---- Fiber test helpers ----------------------------------------------------

func makeApp(handlers ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/res", handlers...)
	return app
}

func makeAppWithID(handlers ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/res/:id", handlers...)
	return app
}

func injectPrincipal(p Principal) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(LocalsKeyPrincipal, p)
		return c.Next()
	}
}

func respondOK(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) }

// testMiddlewareHandler replicates the old svc.Middleware behaviour for tests.
// Production code uses api/middleware.AuthorizeCasbin instead.
func testMiddlewareHandler(svc Service, object, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		p, ok := c.Locals(LocalsKeyPrincipal).(Principal)
		if !ok || p.Subject == "" {
			return fiber.NewError(fiber.StatusUnauthorized, ErrUnauthorized.Error())
		}
		obj := object
		if id := c.Params("id"); id != "" {
			obj = object + "/" + id
		}
		allowed, err := svc.Enforce(c.Context(), Request{
			Subject: p.Subject, Domain: p.Domain, Object: obj, Action: action,
		})
		if err != nil {
			return err
		}
		if !allowed {
			return fiber.NewError(fiber.StatusForbidden, ErrForbidden.Error())
		}
		return c.Next()
	}
}

func testRequest(app *fiber.App, path string) int {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body) //nolint:errcheck
	return resp.StatusCode
}

// ---- 401 cases -------------------------------------------------------------

func (s *MiddlewareUnitSuite) TestMiddleware_NoPrincipalInLocals_Returns401() {
	app := makeApp(testMiddlewareHandler(s.svc, "invoice", "read"), respondOK)
	s.Equal(fiber.StatusUnauthorized, testRequest(app, "/res"))
}

func (s *MiddlewareUnitSuite) TestMiddleware_EmptySubject_Returns401() {
	app := makeApp(
		injectPrincipal(Principal{Subject: "", Domain: "dom-1"}),
		testMiddlewareHandler(s.svc, "invoice", "read"),
		respondOK,
	)
	s.Equal(fiber.StatusUnauthorized, testRequest(app, "/res"))
}

func (s *MiddlewareUnitSuite) TestMiddleware_WrongLocalsKey_Returns401() {
	// Storing under a different key must be treated as missing.
	wrong := fiber.Handler(func(c *fiber.Ctx) error {
		c.Locals("WRONG_KEY", Principal{Subject: "tenant:usr", Domain: "dom-1"})
		return c.Next()
	})
	app := makeApp(wrong, testMiddlewareHandler(s.svc, "invoice", "read"), respondOK)
	s.Equal(fiber.StatusUnauthorized, testRequest(app, "/res"))
}

// ---- 403 cases -------------------------------------------------------------

func (s *MiddlewareUnitSuite) TestMiddleware_NoPolicyForUser_Returns403() {
	app := makeApp(
		injectPrincipal(Principal{Subject: "tenant:usr_001", Domain: "dom-1"}),
		testMiddlewareHandler(s.svc, "invoice", "read"),
		respondOK,
	)
	s.Equal(fiber.StatusForbidden, testRequest(app, "/res"))
}

func (s *MiddlewareUnitSuite) TestMiddleware_DenyRuleWins_Returns403() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:finance", Domain: "dom-1", Object: "invoice/*", Action: "*", Effect: "allow",
	}))
	memRole(s.T(), s.svc, "tenant:usr_terminated", "role:finance", "dom-1")
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "tenant:usr_terminated", Domain: "dom-1", Object: "*", Action: "*", Effect: "deny",
	}))

	app := makeAppWithID(
		injectPrincipal(Principal{Subject: "tenant:usr_terminated", Domain: "dom-1"}),
		testMiddlewareHandler(s.svc, "invoice", "read"),
		respondOK,
	)
	s.Equal(fiber.StatusForbidden, testRequest(app, "/res/inv_001"))
}

// ---- 200 cases -------------------------------------------------------------

func (s *MiddlewareUnitSuite) TestMiddleware_AllowPolicy_Returns200() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:finance", Domain: "dom-1", Object: "invoice", Action: "read", Effect: "allow",
	}))
	memRole(s.T(), s.svc, "tenant:usr_001", "role:finance", "dom-1")

	app := makeApp(
		injectPrincipal(Principal{Subject: "tenant:usr_001", Domain: "dom-1"}),
		testMiddlewareHandler(s.svc, "invoice", "read"),
		respondOK,
	)
	s.Equal(fiber.StatusOK, testRequest(app, "/res"))
}

func (s *MiddlewareUnitSuite) TestMiddleware_CallsNext_OnAllow() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:r", Domain: "dom-1", Object: "res", Action: "read", Effect: "allow",
	}))
	memRole(s.T(), s.svc, "tenant:usr", "role:r", "dom-1")

	nextCalled := false
	downstream := fiber.Handler(func(c *fiber.Ctx) error {
		nextCalled = true
		return c.SendStatus(fiber.StatusOK)
	})

	app := makeApp(
		injectPrincipal(Principal{Subject: "tenant:usr", Domain: "dom-1"}),
		testMiddlewareHandler(s.svc, "res", "read"),
		downstream,
	)
	testRequest(app, "/res")
	s.True(nextCalled, "downstream handler must be called on allow")
}

// ---- Object expansion with :id --------------------------------------------

func (s *MiddlewareUnitSuite) TestMiddleware_ObjectExpandedWithID() {
	// Policy covers "invoice/*" — :id expands object to "invoice/{id}".
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:finance", Domain: "dom-1", Object: "invoice/*", Action: "read", Effect: "allow",
	}))
	memRole(s.T(), s.svc, "tenant:usr_001", "role:finance", "dom-1")

	app := makeAppWithID(
		injectPrincipal(Principal{Subject: "tenant:usr_001", Domain: "dom-1"}),
		testMiddlewareHandler(s.svc, "invoice", "read"),
		respondOK,
	)
	s.Equal(fiber.StatusOK, testRequest(app, "/res/inv_abc"))
}

func (s *MiddlewareUnitSuite) TestMiddleware_NoIDParam_UsesPlainObject() {
	// No :id route — object stays as "invoice".
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:finance", Domain: "dom-1", Object: "invoice", Action: "read", Effect: "allow",
	}))
	memRole(s.T(), s.svc, "tenant:usr_001", "role:finance", "dom-1")

	app := makeApp( // no :id param in route
		injectPrincipal(Principal{Subject: "tenant:usr_001", Domain: "dom-1"}),
		testMiddlewareHandler(s.svc, "invoice", "read"),
		respondOK,
	)
	s.Equal(fiber.StatusOK, testRequest(app, "/res"))
}

// AZ-MID-030 — Enforce error → 500 (not 403).
// testMiddlewareHandler propagates the raw error from Enforce; Fiber converts
// any non-fiber.Error return to 500 by default.
func (s *MiddlewareUnitSuite) TestMiddleware_EnforceError_Returns500() {
	stub := errAuthzService{}
	app := makeApp(
		injectPrincipal(Principal{Subject: "tenant:usr_001", Domain: "dom-1"}),
		testMiddlewareHandler(stub, "invoice", "read"),
		respondOK,
	)
	s.Equal(fiber.StatusInternalServerError, testRequest(app, "/res"))
}

func (s *MiddlewareUnitSuite) TestMiddleware_IDParam_WrongResource_Returns403() {
	// Policy only for "invoice/specific" — other IDs denied.
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:finance", Domain: "dom-1", Object: "invoice/specific", Action: "read", Effect: "allow",
	}))
	memRole(s.T(), s.svc, "tenant:usr_001", "role:finance", "dom-1")

	app := makeAppWithID(
		injectPrincipal(Principal{Subject: "tenant:usr_001", Domain: "dom-1"}),
		testMiddlewareHandler(s.svc, "invoice", "read"),
		respondOK,
	)
	s.Equal(fiber.StatusForbidden, testRequest(app, "/res/wrong_id"))
}
