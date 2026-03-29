// Package iam_test contains end-to-end enforcement tests for the AuthzService.
// These tests use an in-memory Casbin enforcer — no database required.
// 
// Phase 16 coverage:
// V4 — Role expiry: expired role is lazily revoked on Enforce() and denies access
// V5 — Tenant isolation: a policy in domain A must not grant access in domain B
package iam_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"awo.so/internal/core/iam"
	"awo.so/internal/shared/logger"
)

// noopLogger

type noopLogger struct{}

func (noopLogger) Debug(_ string, _ ...logger.Fields)                              {}
func (noopLogger) Info(_ string, _ ...logger.Fields)                               {}
func (noopLogger) Warn(_ string, _ ...logger.Fields)                               {}
func (noopLogger) Error(_ string, _ ...logger.Fields)                              {}
func (noopLogger) Fatal(_ string, _ ...logger.Fields)                              {}
func (noopLogger) DebugContext(_ context.Context, _ string, _ ...logger.Fields)   {}
func (noopLogger) InfoContext(_ context.Context, _ string, _ ...logger.Fields)    {}
func (noopLogger) WarnContext(_ context.Context, _ string, _ ...logger.Fields)    {}
func (noopLogger) ErrorContext(_ context.Context, _ string, _ ...logger.Fields)   {}
func (n noopLogger) WithFields(_ logger.Fields) logger.Logger                     { return n }
func (n noopLogger) WithContext(_ context.Context) logger.Logger                  { return n }
func (noopLogger) SetLevel(_ logger.LogLevel)                                      {}
func (noopLogger) Close() error                                                    { return nil }

// mockAuthzRepo
// 
// Controls which roles ListExpiredActiveRoleNames returns on the next call, and
// records which roles were passed to DeactivateRoleAssignment.

type mockAuthzRepo struct {
	expiredRoles []string
	deactivated  []string
}

func (m *mockAuthzRepo) UpsertRoleAssignment(
	_ context.Context, _ uuid.UUID, _, _, _ string, _, _ *string, _ *time.Time,
) error {
	return nil
}

func (m *mockAuthzRepo) DeactivateRoleAssignment(_ context.Context, _, role, _ string) error {
	m.deactivated = append(m.deactivated, role)
	return nil
}

func (m *mockAuthzRepo) ListRoleAssignments(_ context.Context, _, _ string) ([]iam.RoleAssignment, error) {
	return nil, nil
}

func (m *mockAuthzRepo) ListExpiredActiveRoleNames(_ context.Context, _, _ string) ([]string, error) {
	return m.expiredRoles, nil
}

// suite

type AuthzEnforceSuite struct {
	suite.Suite
	repo *mockAuthzRepo
	svc  iam.AuthzService
}

func TestAuthzEnforceSuite(t *testing.T) { suite.Run(t, new(AuthzEnforceSuite)) }

func (s *AuthzEnforceSuite) SetupTest() {
	s.repo = &mockAuthzRepo{}
	var err error
	s.svc, err = iam.NewInMemoryAuthzService(s.repo, noopLogger{})
	require.NoError(s.T(), err)
}

// V4 — Role expiry: lazy revoke
// 
// Scenario: a user holds role:finance-manager which grants invoice/* access.
// The role has expired in the DB. On the next Enforce() call the service reads
// the expired role from the repository, removes it from the in-memory enforcer,
// deactivates the DB row, then re-evaluates — returning false.

func (s *AuthzEnforceSuite) TestV4_ExpiredRole_LazilyCleaned_DeniesAccess() {
	ctx := context.Background()
	tenantID := uuid.New().String()
	subject := "tenant:user-v4"
	role := "role:finance-manager"

	// Assign role and policy — user has access before expiry.
	require.NoError(s.T(), s.svc.AssignRole(ctx, tenantID, subject, role, tenantID))
	require.NoError(s.T(), s.svc.AddPolicy(ctx, iam.Policy{
		Subject: role, Domain: tenantID, Object: "invoice/*", Action: "*", Effect: "allow",
	}))

	// Sanity check: access is granted.
	allowed, err := s.svc.Enforce(ctx, iam.Request{
		Subject: subject, Domain: tenantID, Object: "invoice/123", Action: "read",
	})
	require.NoError(s.T(), err)
	require.True(s.T(), allowed, "role must grant access before expiry")

	// Simulate the role expiring: the repository now reports it as expired.
	s.repo.expiredRoles = []string{role}

	// Enforce again: the expired role is lazily removed → access denied.
	allowed, err = s.svc.Enforce(ctx, iam.Request{
		Subject: subject, Domain: tenantID, Object: "invoice/123", Action: "read",
	})
	require.NoError(s.T(), err)
	require.False(s.T(), allowed, "expired role must not grant access after lazy revoke")

	// The repo's DeactivateRoleAssignment must have been called with the expired role.
	require.Contains(s.T(), s.repo.deactivated, role,
		"expired role must be deactivated in the repository")
}

// V5 — Tenant isolation: cross-tenant policy leak
// 
// Scenario: tenantA grants role:finance-manager to a subject with invoice/* access.
// The same subject attempts to access invoice/* in tenantB's domain — must be denied.
// Casbin enforces domain isolation: r.dom == p.dom prevents cross-domain matches.

func (s *AuthzEnforceSuite) TestV5_CrossTenantPolicyLeak_Denied() {
	ctx := context.Background()
	tenantA := uuid.New().String()
	tenantB := uuid.New().String()
	subject := "tenant:user-v5"
	role := "role:finance-manager"

	// Configure role + policy only in tenantA's domain.
	require.NoError(s.T(), s.svc.AssignRole(ctx, tenantA, subject, role, tenantA))
	require.NoError(s.T(), s.svc.AddPolicy(ctx, iam.Policy{
		Subject: role, Domain: tenantA, Object: "invoice/*", Action: "*", Effect: "allow",
	}))

	// tenantA access: must be allowed.
	allowedA, err := s.svc.Enforce(ctx, iam.Request{
		Subject: subject, Domain: tenantA, Object: "invoice/123", Action: "read",
	})
	require.NoError(s.T(), err)
	require.True(s.T(), allowedA, "tenantA user must access tenantA resources")

	// tenantB access: same subject + object + action, different domain — must be denied.
	allowedB, err := s.svc.Enforce(ctx, iam.Request{
		Subject: subject, Domain: tenantB, Object: "invoice/123", Action: "read",
	})
	require.NoError(s.T(), err)
	require.False(s.T(), allowedB,
		"tenantA policy must not leak into tenantB: cross-domain access must be denied")
}
