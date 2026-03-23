package iam

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"

	db "awo.so/db/sqlc"
)

// ---------------------------------------------------------------------------
// ServiceSuite — DB-backed integration tests for New(), Enforce,
// EnforceBatch, AddPolicy, RemovePolicy, GetPolicies, InvalidateCache.
// Requires DATABASE_URL environment variable.
// ---------------------------------------------------------------------------

type ServiceSuite struct {
	suite.Suite
	pool *pgxpool.Pool
	svc  Service
	ctx  context.Context
}

func TestServiceSuite(t *testing.T) { suite.Run(t, new(ServiceSuite)) }

func (s *ServiceSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pool = testPool(s.T())
	s.svc = newTestService(s.T(), s.pool)
	seedTestTenant(s.T(), s.pool)
}

func (s *ServiceSuite) SetupTest() {
	cleanTables(s.T(), s.pool)
}

func (s *ServiceSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// ---- Constructor -----------------------------------------------------------

func (s *ServiceSuite) TestNew_NilStore_ReturnsError() {
	_, err := New(Config{Store: nil, Cache: noopCache{}, Logger: noopLogger{}})
	s.Require().Error(err)
	s.Contains(err.Error(), "store is required")
}

func (s *ServiceSuite) TestNew_NilLogger_ReturnsError() {
	_, err := New(Config{Store: db.NewStore(s.pool), Cache: noopCache{}, Logger: nil})
	s.Require().Error(err)
	s.Contains(err.Error(), "logger is required")
}

func (s *ServiceSuite) TestNew_Success() {
	svc, err := New(Config{Store: db.NewStore(s.pool), Cache: noopCache{}, Logger: noopLogger{}})
	s.Require().NoError(err)
	s.NotNil(svc)
}

// ---- Enforce: input validation ---------------------------------------------

func (s *ServiceSuite) TestEnforce_EmptySubject_ReturnsInvalidRequest() {
	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: "", Domain: "dom", Object: "obj", Action: "act",
	})
	s.ErrorIs(err, ErrInvalidRequest)
	s.False(ok)
}

func (s *ServiceSuite) TestEnforce_EmptyDomain_ReturnsInvalidRequest() {
	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: "tenant:usr", Domain: "", Object: "obj", Action: "act",
	})
	s.ErrorIs(err, ErrInvalidRequest)
	s.False(ok)
}

func (s *ServiceSuite) TestEnforce_EmptyObject_ReturnsInvalidRequest() {
	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: "tenant:usr", Domain: "dom", Object: "", Action: "act",
	})
	s.ErrorIs(err, ErrInvalidRequest)
	s.False(ok)
}

func (s *ServiceSuite) TestEnforce_EmptyAction_ReturnsInvalidRequest() {
	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: "tenant:usr", Domain: "dom", Object: "obj", Action: "",
	})
	s.ErrorIs(err, ErrInvalidRequest)
	s.False(ok)
}

// ---- Enforce: default deny ------------------------------------------------

func (s *ServiceSuite) TestEnforce_NoPolicy_DefaultDeny() {
	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: testSubject,
		Domain:  testDomain,
		Object:  "invoice/123",
		Action:  "read",
	})
	s.NoError(err)
	s.False(ok, "no policies → default deny")
}

// ---- Enforce: allow --------------------------------------------------------

func (s *ServiceSuite) TestEnforce_Allow_ExactMatch() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole,
		Domain:  testDomain,
		Object:  "invoice/inv_001",
		Action:  "read",
		Effect:  "allow",
	}))
	// Assign the role to the subject (Casbin g-rule).
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain))

	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: testSubject,
		Domain:  testDomain,
		Object:  "invoice/inv_001",
		Action:  "read",
	})
	s.NoError(err)
	s.True(ok)
}

func (s *ServiceSuite) TestEnforce_Allow_WildcardObject() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole,
		Domain:  testDomain,
		Object:  "invoice/*",
		Action:  "read",
		Effect:  "allow",
	}))
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain))

	for _, obj := range []string{"invoice/inv_001", "invoice/inv_999", "invoice/xyz"} {
		s.Run(obj, func() {
			ok, err := s.svc.Enforce(s.ctx, Request{
				Subject: testSubject,
				Domain:  testDomain,
				Object:  obj,
				Action:  "read",
			})
			s.NoError(err)
			s.True(ok, "wildcard invoice/* must match %s", obj)
		})
	}
}

func (s *ServiceSuite) TestEnforce_Allow_WildcardAction() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole,
		Domain:  testDomain,
		Object:  "invoice/*",
		Action:  "*",
		Effect:  "allow",
	}))
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain))

	for _, act := range []string{"read", "create", "delete", "approve", "export"} {
		s.Run(act, func() {
			ok, err := s.svc.Enforce(s.ctx, Request{
				Subject: testSubject,
				Domain:  testDomain,
				Object:  "invoice/123",
				Action:  act,
			})
			s.NoError(err)
			s.True(ok, "wildcard action * must cover %s", act)
		})
	}
}

// ---- Enforce: deny-override -----------------------------------------------

func (s *ServiceSuite) TestEnforce_DenyOverridesAllow() {
	// Allow via role
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole,
		Domain:  testDomain,
		Object:  "invoice/*",
		Action:  "*",
		Effect:  "allow",
	}))
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain))

	// Blanket deny on the subject itself
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testSubject,
		Domain:  testDomain,
		Object:  "*",
		Action:  "*",
		Effect:  "deny",
	}))

	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: testSubject,
		Domain:  testDomain,
		Object:  "invoice/123",
		Action:  "read",
	})
	s.NoError(err)
	s.False(ok, "deny must override allow")
}

// ---- Enforce: domain isolation ---------------------------------------------

func (s *ServiceSuite) TestEnforce_PolicyDoesNotCrossDomains() {
	dom1 := testDomain
	dom2 := "bbbbbbbb-cccc-dddd-eeee-ffffffffffff"

	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole,
		Domain:  dom1,
		Object:  "invoice/*",
		Action:  "read",
		Effect:  "allow",
	}))
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, dom1))

	// dom2 has no policies at all
	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: testSubject,
		Domain:  dom2,
		Object:  "invoice/123",
		Action:  "read",
	})
	s.NoError(err)
	s.False(ok, "policy in dom1 must not bleed into dom2")
}

// ---- EnforceBatch ----------------------------------------------------------

func (s *ServiceSuite) TestEnforceBatch_EmptyInput_ReturnsNil() {
	results, err := s.svc.EnforceBatch(s.ctx, []Request{})
	s.NoError(err)
	s.Nil(results)
}

func (s *ServiceSuite) TestEnforceBatch_InvalidRequest_ReturnsError() {
	reqs := []Request{
		{Subject: testSubject, Domain: testDomain, Object: "invoice/1", Action: "read"},
		{Subject: "", Domain: testDomain, Object: "invoice/2", Action: "read"}, // empty subject
	}
	results, err := s.svc.EnforceBatch(s.ctx, reqs)
	s.ErrorIs(err, ErrInvalidRequest)
	s.Nil(results)
}

func (s *ServiceSuite) TestEnforceBatch_MixedResults() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole,
		Domain:  testDomain,
		Object:  "invoice/*",
		Action:  "read",
		Effect:  "allow",
	}))
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain))

	reqs := []Request{
		{Subject: testSubject, Domain: testDomain, Object: "invoice/1", Action: "read"},    // allow
		{Subject: testSubject, Domain: testDomain, Object: "invoice/1", Action: "delete"},  // deny (no policy)
		{Subject: "tenant:other", Domain: testDomain, Object: "invoice/1", Action: "read"}, // deny (no role)
	}
	results, err := s.svc.EnforceBatch(s.ctx, reqs)
	s.Require().NoError(err)
	s.Require().Len(results, 3)
	s.True(results[0])
	s.False(results[1])
	s.False(results[2])
}

// ---- AddPolicy validation --------------------------------------------------

func (s *ServiceSuite) TestAddPolicy_InvalidEffect_Empty() {
	err := s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole, Domain: testDomain, Object: "obj", Action: "act", Effect: "",
	})
	s.Error(err)
}

func (s *ServiceSuite) TestAddPolicy_InvalidEffect_Uppercase() {
	err := s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole, Domain: testDomain, Object: "obj", Action: "act", Effect: "ALLOW",
	})
	s.Error(err)
}

func (s *ServiceSuite) TestAddPolicy_InvalidEffect_Unknown() {
	err := s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole, Domain: testDomain, Object: "obj", Action: "act", Effect: "permit",
	})
	s.Error(err)
}

func (s *ServiceSuite) TestAddPolicy_EmptySubject_ReturnsInvalidRequest() {
	err := s.svc.AddPolicy(s.ctx, Policy{
		Subject: "", Domain: testDomain, Object: "obj", Action: "act", Effect: "allow",
	})
	s.ErrorIs(err, ErrInvalidRequest)
}

func (s *ServiceSuite) TestAddPolicy_EmptyDomain_ReturnsInvalidRequest() {
	err := s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole, Domain: "", Object: "obj", Action: "act", Effect: "allow",
	})
	s.ErrorIs(err, ErrInvalidRequest)
}

func (s *ServiceSuite) TestAddPolicy_Duplicate_ReturnsConflict() {
	pol := Policy{
		Subject: testRole, Domain: testDomain,
		Object: "invoice/*", Action: "read", Effect: "allow",
	}
	s.Require().NoError(s.svc.AddPolicy(s.ctx, pol))

	err := s.svc.AddPolicy(s.ctx, pol)
	s.ErrorIs(err, ErrPolicyConflict)
}

// ---- RemovePolicy ----------------------------------------------------------

func (s *ServiceSuite) TestRemovePolicy_RemovesMatchingRule() {
	pol := Policy{
		Subject: testRole, Domain: testDomain,
		Object: "invoice/*", Action: "read", Effect: "allow",
	}
	s.Require().NoError(s.svc.AddPolicy(s.ctx, pol))
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain))

	// Verify allow before removal.
	ok, _ := s.svc.Enforce(s.ctx, Request{
		Subject: testSubject, Domain: testDomain, Object: "invoice/1", Action: "read",
	})
	s.True(ok)

	s.Require().NoError(s.svc.RemovePolicy(s.ctx, pol))

	// After removal — default deny.
	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: testSubject, Domain: testDomain, Object: "invoice/1", Action: "read",
	})
	s.NoError(err)
	s.False(ok, "removed policy must no longer grant access")
}

func (s *ServiceSuite) TestRemovePolicy_EmptyFields_ReturnsInvalidRequest() {
	err := s.svc.RemovePolicy(s.ctx, Policy{
		Subject: "", Domain: testDomain, Object: "obj", Action: "act",
	})
	s.ErrorIs(err, ErrInvalidRequest)
}

// ---- GetPolicies -----------------------------------------------------------

func (s *ServiceSuite) TestGetPolicies_ReturnsDomainPolicies() {
	policies := []Policy{
		{Subject: "role:a", Domain: testDomain, Object: "invoice/*", Action: "read", Effect: "allow"},
		{Subject: "role:b", Domain: testDomain, Object: "payment/*", Action: "*", Effect: "allow"},
	}
	for _, p := range policies {
		s.Require().NoError(s.svc.AddPolicy(s.ctx, p))
	}

	// Add a policy in a different domain — must not appear.
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: "role:c", Domain: "other-dom", Object: "anything", Action: "read", Effect: "allow",
	}))

	result, err := s.svc.GetPolicies(s.ctx, testDomain)
	s.Require().NoError(err)
	s.Len(result, 2, "only testDomain policies should be returned")
	for _, p := range result {
		s.Equal(testDomain, p.Domain)
	}
}

func (s *ServiceSuite) TestGetPolicies_EmptyDomain_ReturnsEmpty() {
	result, err := s.svc.GetPolicies(s.ctx, "no-such-domain")
	s.NoError(err)
	s.Empty(result)
}

func (s *ServiceSuite) TestGetPolicies_FieldsCorrectlyMapped() {
	pol := Policy{
		Subject: testRole,
		Domain:  testDomain,
		Object:  "invoice/*",
		Action:  "*",
		Effect:  "allow",
	}
	s.Require().NoError(s.svc.AddPolicy(s.ctx, pol))

	result, err := s.svc.GetPolicies(s.ctx, testDomain)
	s.Require().NoError(err)
	s.Require().Len(result, 1)

	got := result[0]
	s.Equal(pol.Subject, got.Subject)
	s.Equal(pol.Domain, got.Domain)
	s.Equal(pol.Object, got.Object)
	s.Equal(pol.Action, got.Action)
	s.Equal(pol.Effect, got.Effect)
}

// ---- InvalidateCache -------------------------------------------------------

func (s *ServiceSuite) TestInvalidateCache_ReloadsDirectSQLWrite() {
	// Insert a rule directly via raw SQL (bypassing the Service).
	_, err := s.pool.Exec(s.ctx, `
		INSERT INTO casbin_rule(ptype,v0,v1,v2,v3,v4,v5)
		VALUES('p',$1,$2,'invoice/*','*','allow','')`,
		testRole, testDomain,
	)
	s.Require().NoError(err)

	// Assign role via Casbin directly through Service (in-memory + DB).
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain))

	// Before reload: in-memory model was NOT updated by raw SQL insert, so
	// GetPolicies still reflects only previously loaded state. After reload it must appear.
	s.Require().NoError(s.svc.InvalidateCache(s.ctx))

	result, err := s.svc.GetPolicies(s.ctx, testDomain)
	s.Require().NoError(err)
	s.NotEmpty(result, "policy inserted via raw SQL must appear after InvalidateCache")
}
