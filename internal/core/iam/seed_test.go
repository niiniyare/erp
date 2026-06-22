package iam_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"awo.so/internal/core/iam"
	"awo.so/internal/shared/logger"
)

// Note: repo is passed as nil to NewInMemoryAuthzService — the service skips
// all persistence calls when repo == nil, which is sufficient for these tests.

// =============================================================================
// SeedDefaultRoles Suite
// =============================================================================

type SeedDefaultRolesSuite struct {
	suite.Suite
	svc iam.AuthzService
	ctx context.Context
}

func TestSeedDefaultRolesSuite(t *testing.T) { suite.Run(t, new(SeedDefaultRolesSuite)) }

func (s *SeedDefaultRolesSuite) SetupTest() {
	log := logger.NewNoOp()
	var err error
	s.svc, err = iam.NewInMemoryAuthzService(nil, log)
	s.Require().NoError(err)
	s.ctx = context.Background()
}

func (s *SeedDefaultRolesSuite) TestSeedDefaultRoles_Succeeds() {
	tenantID := uuid.New().String()
	err := iam.SeedDefaultRoles(s.ctx, s.svc, tenantID)
	s.Require().NoError(err)
}

func (s *SeedDefaultRolesSuite) TestSeedDefaultRoles_Idempotent() {
	tenantID := uuid.New().String()
	s.Require().NoError(iam.SeedDefaultRoles(s.ctx, s.svc, tenantID))
	// Second call must not return an error (ErrPolicyConflict is swallowed)
	s.Require().NoError(iam.SeedDefaultRoles(s.ctx, s.svc, tenantID))
}

func (s *SeedDefaultRolesSuite) TestSeedDefaultRoles_AdminRoleCanReadFinance() {
	tenantID := uuid.New().String()
	s.Require().NoError(iam.SeedDefaultRoles(s.ctx, s.svc, tenantID))

	policies, err := s.svc.GetPolicies(s.ctx, iam.TenantDomain(tenantID))
	s.Require().NoError(err)

	// Expect read + write for each of the 11 objects = 22 policies
	s.GreaterOrEqual(len(policies), 22, "expected at least 22 seeded policies")

	// Verify at least one finance policy and one IAM policy are present
	var hasFinanceRead, hasIAMRead bool
	for _, p := range policies {
		if p.Object == "finance.accounts" && p.Action == "read" {
			hasFinanceRead = true
		}
		if p.Object == "iam.roles" && p.Action == "read" {
			hasIAMRead = true
		}
	}
	s.True(hasFinanceRead, "finance.accounts read must be seeded")
	s.True(hasIAMRead, "iam.roles read must be seeded")
}

func (s *SeedDefaultRolesSuite) TestSeedDefaultRoles_IsolatedPerTenant() {
	tenantA := uuid.New().String()
	tenantB := uuid.New().String()

	s.Require().NoError(iam.SeedDefaultRoles(s.ctx, s.svc, tenantA))

	// Tenant B should have no policies yet
	policiesB, err := s.svc.GetPolicies(s.ctx, iam.TenantDomain(tenantB))
	s.Require().NoError(err)
	s.Empty(policiesB, "policies seeded for tenant A must not leak into tenant B")
}

// =============================================================================
// AssignAdminRole Suite
// =============================================================================

type AssignAdminRoleSuite struct {
	suite.Suite
	svc iam.AuthzService
	ctx context.Context
}

func TestAssignAdminRoleSuite(t *testing.T) { suite.Run(t, new(AssignAdminRoleSuite)) }

func (s *AssignAdminRoleSuite) SetupTest() {
	log := logger.NewNoOp()
	var err error
	s.svc, err = iam.NewInMemoryAuthzService(nil, log)
	s.Require().NoError(err)
	s.ctx = context.Background()
}

func (s *AssignAdminRoleSuite) TestAssignAdminRole_GrantsRole() {
	tenantID := uuid.New().String()
	userID := uuid.New().String()

	s.Require().NoError(iam.AssignAdminRole(s.ctx, s.svc, tenantID, userID))

	subj := iam.TenantSubject(userID)
	dom := iam.TenantDomain(tenantID)
	has, err := s.svc.HasRole(s.ctx, subj, "role:tenant.admin", dom)
	s.Require().NoError(err)
	s.True(has)
}

func (s *AssignAdminRoleSuite) TestAssignAdminRole_Idempotent() {
	tenantID := uuid.New().String()
	userID := uuid.New().String()

	s.Require().NoError(iam.AssignAdminRole(s.ctx, s.svc, tenantID, userID))
	// Re-assigning same role must not error (Casbin ignores duplicate g-rules)
	s.Require().NoError(iam.AssignAdminRole(s.ctx, s.svc, tenantID, userID))
}

func (s *AssignAdminRoleSuite) TestAssignAdminRole_AfterSeed_CanReadFinance() {
	tenantID := uuid.New().String()
	userID := uuid.New().String()

	s.Require().NoError(iam.SeedDefaultRoles(s.ctx, s.svc, tenantID))
	s.Require().NoError(iam.AssignAdminRole(s.ctx, s.svc, tenantID, userID))

	subj := iam.TenantSubject(userID)
	dom := iam.TenantDomain(tenantID)

	for _, obj := range []string{
		"finance.accounts",
		"finance.transactions",
		"finance.periods",
		"iam.policies",
		"iam.roles",
		"contracts",
	} {
		for _, act := range []string{"read", "write"} {
			allowed, err := s.svc.Enforce(s.ctx, iam.Request{
				Subject: subj, Domain: dom, Object: obj, Action: act,
			})
			s.Require().NoError(err)
			s.True(allowed, "admin should be allowed: %s %s", act, obj)
		}
	}
}

func (s *AssignAdminRoleSuite) TestAssignAdminRole_DifferentUsersInSameTenant() {
	tenantID := uuid.New().String()
	userA := uuid.New().String()
	userB := uuid.New().String()

	s.Require().NoError(iam.AssignAdminRole(s.ctx, s.svc, tenantID, userA))
	s.Require().NoError(iam.AssignAdminRole(s.ctx, s.svc, tenantID, userB))

	dom := iam.TenantDomain(tenantID)
	hasA, _ := s.svc.HasRole(s.ctx, iam.TenantSubject(userA), "role:tenant.admin", dom)
	hasB, _ := s.svc.HasRole(s.ctx, iam.TenantSubject(userB), "role:tenant.admin", dom)
	s.True(hasA)
	s.True(hasB)
}
