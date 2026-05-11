package iam

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

//
// RolesSuite — DB-backed integration tests for AssignRole, RevokeRole,
// GetRoles, HasRole, GetAssignments, and temporal role expiry.
// Requires DATABASE_URL environment variable.
//

type RolesSuite struct {
	suite.Suite
	pool *pgxpool.Pool
	svc  Service
	ctx  context.Context
}

func TestRolesSuite(t *testing.T) { suite.Run(t, new(RolesSuite)) }

func (s *RolesSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pool = testPool(s.T())
	s.svc = newTestService(s.T(), s.pool)
	seedTestTenant(s.T(), s.pool)
}

func (s *RolesSuite) SetupTest() {
	cleanTables(s.T(), s.pool)
}

func (s *RolesSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// ---- AssignRole ------------------------------------------------------------

func (s *RolesSuite) TestAssignRole_PersistsToDB() {
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("platform:system")))

	var count int
	s.Require().NoError(s.pool.QueryRow(s.ctx, `
		SELECT COUNT(*) FROM role_assignments
		WHERE subject=$1 AND role_name=$2 AND domain=$3 AND is_active=TRUE`,
		testSubject, testRole, testDomain,
	).Scan(&count))
	s.Equal(1, count)
}

func (s *RolesSuite) TestAssignRole_AddsCasbinGRule() {
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("platform:system")))

	has, err := s.svc.HasRole(s.ctx, testSubject, testRole, testDomain)
	s.Require().NoError(err)
	s.True(has)
}

func (s *RolesSuite) TestAssignRole_Idempotent() {
	// Call twice with same params — must not error or duplicate.
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("platform:system")))
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("platform:system")))

	var count int
	s.Require().NoError(s.pool.QueryRow(s.ctx, `
		SELECT COUNT(*) FROM role_assignments
		WHERE subject=$1 AND role_name=$2 AND domain=$3`,
		testSubject, testRole, testDomain,
	).Scan(&count))
	s.Equal(1, count, "idempotent UPSERT must not create duplicate rows")
}

func (s *RolesSuite) TestAssignRole_WithExpiry_StoredCorrectly() {
	expiry := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)

	s.Require().NoError(s.svc.AssignRole(
		s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithExpiry(expiry),
		WithAssignedBy("platform:system"),
	))

	var storedExpiry time.Time
	s.Require().NoError(s.pool.QueryRow(s.ctx, `
		SELECT expires_at FROM role_assignments
		WHERE subject=$1 AND role_name=$2 AND domain=$3`,
		testSubject, testRole, testDomain,
	).Scan(&storedExpiry))

	s.WithinDuration(expiry, storedExpiry.UTC(), time.Second)
}

func (s *RolesSuite) TestAssignRole_WithAssignedBy_StoredCorrectly() {
	s.Require().NoError(s.svc.AssignRole(
		s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("tenant:usr_ceo"),
	))

	var assignedBy string
	s.Require().NoError(s.pool.QueryRow(s.ctx, `
		SELECT COALESCE(assigned_by,'') FROM role_assignments
		WHERE subject=$1 AND role_name=$2 AND domain=$3`,
		testSubject, testRole, testDomain,
	).Scan(&assignedBy))

	s.Equal("tenant:usr_ceo", assignedBy)
}

func (s *RolesSuite) TestAssignRole_WithDelegatedBy_StoredCorrectly() {
	s.Require().NoError(s.svc.AssignRole(
		s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("tenant:usr_ceo"),
		WithDelegatedBy("tenant:usr_cfo"),
	))

	var delegatedBy string
	s.Require().NoError(s.pool.QueryRow(s.ctx, `
		SELECT COALESCE(delegated_by,'') FROM role_assignments
		WHERE subject=$1 AND role_name=$2 AND domain=$3`,
		testSubject, testRole, testDomain,
	).Scan(&delegatedBy))

	s.Equal("tenant:usr_cfo", delegatedBy)
}

func (s *RolesSuite) TestAssignRole_ReactivatesPreviouslyRevoked() {
	// Assign, then revoke, then re-assign.
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("platform:system")))
	s.Require().NoError(s.svc.RevokeRole(s.ctx, testSubject, testRole, testDomain))

	has, _ := s.svc.HasRole(s.ctx, testSubject, testRole, testDomain)
	s.False(has, "must be gone after revoke")

	// Re-assign
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("platform:system")))

	has, err := s.svc.HasRole(s.ctx, testSubject, testRole, testDomain)
	s.Require().NoError(err)
	s.True(has, "must be active again after re-assign")

	var isActive bool
	s.Require().NoError(s.pool.QueryRow(s.ctx, `
		SELECT is_active FROM role_assignments
		WHERE subject=$1 AND role_name=$2 AND domain=$3`,
		testSubject, testRole, testDomain,
	).Scan(&isActive))
	s.True(isActive)
}

// ---- RevokeRole ------------------------------------------------------------

func (s *RolesSuite) TestRevokeRole_MarksInactiveAndRemovesGRule() {
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("platform:system")))
	s.Require().NoError(s.svc.RevokeRole(s.ctx, testSubject, testRole, testDomain))

	// Casbin in-memory: role gone.
	has, err := s.svc.HasRole(s.ctx, testSubject, testRole, testDomain)
	s.Require().NoError(err)
	s.False(has)

	// DB row: preserved for audit but inactive.
	var isActive bool
	var rowCount int
	s.Require().NoError(s.pool.QueryRow(s.ctx, `
		SELECT COUNT(*), BOOL_OR(is_active) FROM role_assignments
		WHERE subject=$1 AND role_name=$2 AND domain=$3`,
		testSubject, testRole, testDomain,
	).Scan(&rowCount, &isActive))

	s.Equal(1, rowCount, "audit row must be preserved")
	s.False(isActive, "is_active must be FALSE after revoke")
}

func (s *RolesSuite) TestRevokeRole_NonExistent_NoError() {
	err := s.svc.RevokeRole(s.ctx, "tenant:nobody", "role:nobody", testDomain)
	s.NoError(err, "revoking a non-existent role must not error")
}

// ---- GetRoles --------------------------------------------------------------

func (s *RolesSuite) TestGetRoles_ReturnsSingleRole() {
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("platform:system")))

	roles, err := s.svc.GetRoles(s.ctx, testSubject, testDomain)
	s.Require().NoError(err)
	s.Equal([]string{testRole}, roles)
}

func (s *RolesSuite) TestGetRoles_ReturnsMultipleRoles() {
	roles := []string{"role:finance-manager", "role:sales-rep", "role:viewer"}
	for _, r := range roles {
		s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, r, testDomain,
			WithAssignedBy("platform:system")))
	}

	got, err := s.svc.GetRoles(s.ctx, testSubject, testDomain)
	s.Require().NoError(err)
	s.Len(got, len(roles))
	s.ElementsMatch(roles, got)
}

func (s *RolesSuite) TestGetRoles_NoRoles_ReturnsEmptySlice_NotError() {
	roles, err := s.svc.GetRoles(s.ctx, "tenant:has_no_roles", testDomain)
	s.NoError(err)
	s.Empty(roles)
}

func (s *RolesSuite) TestGetRoles_IsDomainScoped() {
	dom2 := "cccccccc-dddd-eeee-ffff-000000000000"
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("platform:system")))

	roles, err := s.svc.GetRoles(s.ctx, testSubject, dom2)
	s.Require().NoError(err)
	s.Empty(roles, "roles in testDomain must not appear in dom2")
}

// ---- HasRole ---------------------------------------------------------------

func (s *RolesSuite) TestHasRole_TrueAfterAssign() {
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("platform:system")))

	has, err := s.svc.HasRole(s.ctx, testSubject, testRole, testDomain)
	s.Require().NoError(err)
	s.True(has)
}

func (s *RolesSuite) TestHasRole_FalseAfterRevoke() {
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("platform:system")))
	s.Require().NoError(s.svc.RevokeRole(s.ctx, testSubject, testRole, testDomain))

	has, err := s.svc.HasRole(s.ctx, testSubject, testRole, testDomain)
	s.Require().NoError(err)
	s.False(has)
}

func (s *RolesSuite) TestHasRole_FalseForUnknownSubject() {
	has, err := s.svc.HasRole(s.ctx, "tenant:unknown", testRole, testDomain)
	s.Require().NoError(err)
	s.False(has)
}

func (s *RolesSuite) TestHasRole_IsDomainScoped() {
	dom2 := "cccccccc-dddd-eeee-ffff-000000000000"
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("platform:system")))

	has, err := s.svc.HasRole(s.ctx, testSubject, testRole, dom2)
	s.Require().NoError(err)
	s.False(has, "role in testDomain must not appear in dom2")
}

// ---- GetAssignments --------------------------------------------------------

func (s *RolesSuite) TestGetAssignments_ReturnsRow() {
	s.Require().NoError(s.svc.AssignRole(
		s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("tenant:usr_ceo"),
	))

	assignments, err := s.svc.GetAssignments(s.ctx, testSubject, testDomain)
	s.Require().NoError(err)
	s.Require().Len(assignments, 1)

	a := assignments[0]
	s.NotEmpty(a.ID)
	s.Equal(testSubject, a.Subject)
	s.Equal(testRole, a.Role)
	s.Equal(testDomain, a.Domain)
	s.Equal(testTenantID, a.TenantID)
	s.Equal("tenant:usr_ceo", a.AssignedBy)
	s.True(a.IsActive)
	s.False(a.CreatedAt.IsZero())
}

func (s *RolesSuite) TestGetAssignments_IncludesInactiveRows() {
	// Assign and then revoke — audit row must still be present.
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("platform:system")))
	s.Require().NoError(s.svc.RevokeRole(s.ctx, testSubject, testRole, testDomain))

	assignments, err := s.svc.GetAssignments(s.ctx, testSubject, testDomain)
	s.Require().NoError(err)
	s.Require().Len(assignments, 1)
	s.False(assignments[0].IsActive, "is_active must be FALSE in audit row")
}

func (s *RolesSuite) TestGetAssignments_EmptyForUnknownSubject() {
	assignments, err := s.svc.GetAssignments(s.ctx, "tenant:nobody", testDomain)
	s.NoError(err)
	s.Empty(assignments)
}

func (s *RolesSuite) TestGetAssignments_WithExpiry_PopulatesExpiresAt() {
	expiry := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Second)

	s.Require().NoError(s.svc.AssignRole(
		s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithExpiry(expiry),
		WithAssignedBy("platform:system"),
	))

	assignments, err := s.svc.GetAssignments(s.ctx, testSubject, testDomain)
	s.Require().NoError(err)
	s.Require().Len(assignments, 1)
	s.Require().NotNil(assignments[0].ExpiresAt)
	s.WithinDuration(expiry, assignments[0].ExpiresAt.UTC(), time.Second)
}

func (s *RolesSuite) TestGetAssignments_PermanentRole_NilExpiresAt() {
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("platform:system")))

	assignments, err := s.svc.GetAssignments(s.ctx, testSubject, testDomain)
	s.Require().NoError(err)
	s.Require().Len(assignments, 1)
	s.Nil(assignments[0].ExpiresAt, "permanent role must have nil ExpiresAt")
}

// ---- Temporal roles — lazy expiry -----------------------------------------

func (s *RolesSuite) TestEnforce_ExpiredRole_IsRevoked() {
	// Assign with a past expiry to simulate an already-expired role.
	past := time.Now().UTC().Add(-1 * time.Hour)
	s.Require().NoError(s.svc.AssignRole(
		s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithExpiry(past),
		WithAssignedBy("platform:system"),
	))

	// Add allow policy for the role.
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole, Domain: testDomain, Object: "invoice/*", Action: "read", Effect: "allow",
	}))

	// Before lazy revoke: verify the role is in Casbin memory.
	has, _ := s.svc.HasRole(s.ctx, testSubject, testRole, testDomain)
	s.True(has, "role must be in memory before Enforce triggers lazy cleanup")

	// Enforce triggers revokeExpiredRoles.
	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: testSubject,
		Domain:  testDomain,
		Object:  "invoice/123",
		Action:  "read",
	})
	s.NoError(err)
	s.False(ok, "expired role must cause deny")

	// Casbin in-memory state updated.
	has, _ = s.svc.HasRole(s.ctx, testSubject, testRole, testDomain)
	s.False(has, "role must be gone from Casbin after lazy revoke")

	// DB row marked inactive.
	var isActive bool
	s.Require().NoError(s.pool.QueryRow(s.ctx, `
		SELECT is_active FROM role_assignments
		WHERE subject=$1 AND role_name=$2 AND domain=$3`,
		testSubject, testRole, testDomain,
	).Scan(&isActive))
	s.False(isActive)
}

func (s *RolesSuite) TestEnforce_ActiveRole_NotExpired() {
	// Assign with a far-future expiry.
	future := time.Now().UTC().Add(24 * time.Hour)
	s.Require().NoError(s.svc.AssignRole(
		s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithExpiry(future),
		WithAssignedBy("platform:system"),
	))
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole, Domain: testDomain, Object: "invoice/*", Action: "read", Effect: "allow",
	}))

	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: testSubject,
		Domain:  testDomain,
		Object:  "invoice/123",
		Action:  "read",
	})
	s.NoError(err)
	s.True(ok, "unexpired role must still grant access")

	// DB row stays active.
	var isActive bool
	s.Require().NoError(s.pool.QueryRow(s.ctx, `
		SELECT is_active FROM role_assignments
		WHERE subject=$1 AND role_name=$2 AND domain=$3`,
		testSubject, testRole, testDomain,
	).Scan(&isActive))
	s.True(isActive)
}

func (s *RolesSuite) TestEnforce_PermanentRole_NeverExpires() {
	// No expires_at — permanent role.
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain,
		WithAssignedBy("platform:system")))
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole, Domain: testDomain, Object: "invoice/*", Action: "read", Effect: "allow",
	}))

	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: testSubject,
		Domain:  testDomain,
		Object:  "invoice/123",
		Action:  "read",
	})
	s.NoError(err)
	s.True(ok, "permanent role must never be revoked by lazy cleanup")
}

func (s *RolesSuite) TestEnforce_MultipleExpiredRoles_AllRevoked() {
	rolesToExpire := []string{"role:expired-a", "role:expired-b", "role:expired-c"}
	permanentRole := "role:permanent"
	past := time.Now().UTC().Add(-1 * time.Hour)

	for _, r := range rolesToExpire {
		s.Require().NoError(s.svc.AssignRole(
			s.ctx, testTenantID, testSubject, r, testDomain,
			WithExpiry(past),
			WithAssignedBy("platform:system"),
		))
		s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
			Subject: r, Domain: testDomain, Object: "res/*", Action: "read", Effect: "allow",
		}))
	}
	// One permanent role.
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, permanentRole, testDomain,
		WithAssignedBy("platform:system")))
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: permanentRole, Domain: testDomain, Object: "res/*", Action: "read", Effect: "allow",
	}))

	// Enforce triggers mass expiry.
	ok, err := s.svc.Enforce(s.ctx, Request{
		Subject: testSubject, Domain: testDomain, Object: "res/1", Action: "read",
	})
	s.NoError(err)
	s.True(ok, "permanent role must still allow access")

	// All expired roles must be gone.
	for _, r := range rolesToExpire {
		has, _ := s.svc.HasRole(s.ctx, testSubject, r, testDomain)
		s.False(has, "expired role %s must be revoked", r)
	}

	// Permanent role untouched.
	has, _ := s.svc.HasRole(s.ctx, testSubject, permanentRole, testDomain)
	s.True(has, "permanent role must remain")
}
