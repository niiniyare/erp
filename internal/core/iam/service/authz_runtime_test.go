package service_test

// Runtime hardening tests: Phases 2 (distributed Casbin), 3 (session invalidation), 5 (compensation).
//
// These tests use in-memory Casbin (no DB required) with test doubles for the
// watcher and session invalidator. They verify the contracts:
//   - Every successful mutation calls PolicyWatcher.Notify (AUTHZ-DIST-*)
//   - Watcher error is non-fatal (AUTHZ-DIST-5)
//   - RevokeRole calls SessionInvalidator.InvalidateByUser (SES-INV-1)
//   - RemovePolicy calls SessionInvalidator.InvalidateByTenant for tenant domains (SES-INV-2)
//   - Session invalidation error is non-fatal (SES-INV-3)
//   - Platform domain removal skips tenant session invalidation (SES-INV-4)
//   - AssignRole compensates in-memory state if repo fails (AUTHZ-TXN-1)
//   - RevokeRole compensates in-memory state if repo fails (AUTHZ-TXN-2)

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"awo.so/internal/core/iam/domain"
	"awo.so/internal/core/iam/repository"
	"awo.so/internal/core/iam/service"
	"awo.so/internal/core/iam/watcher"
	"awo.so/internal/shared/logger"
)

// ─── Test doubles ─────────────────────────────────────────────────────────────

// countingWatcher records Notify invocations.
type countingWatcher struct {
	calls     atomic.Int32
	notifyErr error
}

func (w *countingWatcher) Notify(_ context.Context) error {
	w.calls.Add(1)
	return w.notifyErr
}
func (w *countingWatcher) Watch(_ context.Context, _ func()) error { return nil }
func (w *countingWatcher) Close() error                            { return nil }

var _ watcher.PolicyWatcher = (*countingWatcher)(nil)

// trackInv records InvalidateByUser / InvalidateByTenant calls.
type trackInv struct {
	users   []uuid.UUID
	tenants []uuid.UUID
	err     error
}

func (s *trackInv) InvalidateByUser(_ context.Context, id uuid.UUID) error {
	s.users = append(s.users, id)
	return s.err
}

func (s *trackInv) InvalidateByTenant(_ context.Context, id uuid.UUID) error {
	s.tenants = append(s.tenants, id)
	return s.err
}

var _ service.SessionInvalidator = (*trackInv)(nil)

// failRepo wraps a nil repository but returns controlled errors on writes.
type failRepo struct {
	upsertErr     error
	deactivateErr error
}

func (r *failRepo) UpsertRoleAssignment(_ context.Context, _ uuid.UUID, _, _, _ string, _, _ *string, _ *time.Time) error {
	return r.upsertErr
}

func (r *failRepo) DeactivateRoleAssignment(_ context.Context, _, _, _ string) error {
	return r.deactivateErr
}

func (r *failRepo) ListRoleAssignments(_ context.Context, _, _ string) ([]domain.RoleAssignment, error) {
	return nil, nil
}

func (r *failRepo) ListExpiredActiveRoleNames(_ context.Context, _, _ string) ([]string, error) {
	return nil, nil
}

var _ repository.AuthzRepository = (*failRepo)(nil)

// ─── Helpers ──────────────────────────────────────────────────────────────────

func newSvcFull(t *testing.T, repo repository.AuthzRepository, w watcher.PolicyWatcher, inv service.SessionInvalidator) service.AuthzService {
	t.Helper()
	svc, err := service.NewInMemoryAuthzServiceFull(repo, logger.NewNoOp(), w, inv)
	require.NoError(t, err)
	return svc
}

func newSvc(t *testing.T, w watcher.PolicyWatcher, inv service.SessionInvalidator) service.AuthzService {
	return newSvcFull(t, nil, w, inv)
}

// tenantCtx builds a tenant subject + domain for test role operations.
func tenantCtx() (tenantID, userID uuid.UUID, subject, domainName string) {
	tenantID = uuid.New()
	userID = uuid.New()
	subject = domain.TenantSubject(userID.String())
	domainName = domain.TenantDomain(tenantID.String())
	return tenantID, userID, subject, domainName
}

// ─── AUTHZ-DIST: Watcher is notified after every mutation ─────────────────────

func TestAuthzDist1_AddPolicyNotifiesWatcher(t *testing.T) {
	w := &countingWatcher{}
	svc := newSvc(t, w, nil)

	err := svc.AddPolicy(context.Background(), domain.Policy{
		Subject: "role:editor",
		Domain:  "tenant_" + uuid.New().String(),
		Object:  "invoices",
		Action:  "read",
		Effect:  "allow",
	})
	require.NoError(t, err)
	assert.EqualValues(t, 1, w.calls.Load())
}

func TestAuthzDist2_RevokeRoleNotifiesWatcher(t *testing.T) {
	w := &countingWatcher{}
	svc := newSvc(t, w, nil)

	tenantID, _, subject, dom := tenantCtx()
	_ = svc.AssignRole(context.Background(), tenantID.String(), subject, "viewer", dom,
		domain.WithAssignedBy("platform:system"))
	w.calls.Store(0) // reset

	err := svc.RevokeRole(context.Background(), subject, "viewer", dom)
	require.NoError(t, err)
	assert.EqualValues(t, 1, w.calls.Load())
}

func TestAuthzDist3_RemovePolicyNotifiesWatcher(t *testing.T) {
	w := &countingWatcher{}
	svc := newSvc(t, w, nil)

	p := domain.Policy{
		Subject: "role:viewer",
		Domain:  "tenant_" + uuid.New().String(),
		Object:  "reports",
		Action:  "read",
		Effect:  "allow",
	}
	_ = svc.AddPolicy(context.Background(), p)
	w.calls.Store(0)

	err := svc.RemovePolicy(context.Background(), p)
	require.NoError(t, err)
	assert.EqualValues(t, 1, w.calls.Load())
}

func TestAuthzDist4_AssignRoleNotifiesWatcher(t *testing.T) {
	w := &countingWatcher{}
	svc := newSvc(t, w, nil)

	tenantID, _, subject, dom := tenantCtx()
	err := svc.AssignRole(context.Background(), tenantID.String(), subject, "editor", dom,
		domain.WithAssignedBy("platform:system"))
	require.NoError(t, err)
	assert.EqualValues(t, 1, w.calls.Load())
}

// AUTHZ-DIST-5: Watcher.Notify error must not fail the mutation.
func TestAuthzDist5_WatcherErrorNonFatal(t *testing.T) {
	w := &countingWatcher{notifyErr: errors.New("pg notify failed")}
	svc := newSvc(t, w, nil)

	err := svc.AddPolicy(context.Background(), domain.Policy{
		Subject: "role:op",
		Domain:  "tenant_" + uuid.New().String(),
		Object:  "*",
		Action:  "*",
		Effect:  "allow",
	})
	require.NoError(t, err, "mutation must succeed despite watcher error")
	assert.EqualValues(t, 1, w.calls.Load())
}

// ─── SES-INV: Session invalidation bridge ─────────────────────────────────────

// SES-INV-1: RevokeRole evicts user sessions.
func TestSesInv1_RevokeRoleInvalidatesUser(t *testing.T) {
	inv := &trackInv{}
	svc := newSvc(t, watcher.NoopWatcher{}, inv)

	tenantID, userID, subject, dom := tenantCtx()
	_ = svc.AssignRole(context.Background(), tenantID.String(), subject, "viewer", dom,
		domain.WithAssignedBy("platform:system"))

	err := svc.RevokeRole(context.Background(), subject, "viewer", dom)
	require.NoError(t, err)

	require.Len(t, inv.users, 1)
	assert.Equal(t, userID, inv.users[0])
}

// SES-INV-2: RemovePolicy for a tenant domain invalidates all tenant sessions.
func TestSesInv2_RemovePolicyInvalidatesTenant(t *testing.T) {
	inv := &trackInv{}
	svc := newSvc(t, watcher.NoopWatcher{}, inv)

	tenantID := uuid.New()
	p := domain.Policy{
		Subject: "role:editor",
		Domain:  domain.TenantDomain(tenantID.String()),
		Object:  "invoices",
		Action:  "write",
		Effect:  "allow",
	}
	_ = svc.AddPolicy(context.Background(), p)

	err := svc.RemovePolicy(context.Background(), p)
	require.NoError(t, err)

	require.Len(t, inv.tenants, 1)
	assert.Equal(t, tenantID, inv.tenants[0])
}

// SES-INV-3: Session invalidation failure is non-fatal.
func TestSesInv3_InvalidationErrorNonFatal(t *testing.T) {
	inv := &trackInv{err: errors.New("redis timeout")}
	svc := newSvc(t, watcher.NoopWatcher{}, inv)

	tenantID, _, subject, dom := tenantCtx()
	_ = svc.AssignRole(context.Background(), tenantID.String(), subject, "viewer", dom,
		domain.WithAssignedBy("platform:system"))

	err := svc.RevokeRole(context.Background(), subject, "viewer", dom)
	require.NoError(t, err)
}

// SES-INV-4: Removing a platform-domain policy does NOT call tenant invalidation.
func TestSesInv4_PlatformRemovePolicy_NoTenantInvalidation(t *testing.T) {
	inv := &trackInv{}
	svc := newSvc(t, watcher.NoopWatcher{}, inv)

	p := domain.Policy{
		Subject: "platform:sysop",
		Domain:  domain.DomainPlatform,
		Object:  "tenants",
		Action:  "read",
		Effect:  "allow",
	}
	_ = svc.AddPolicy(context.Background(), p)

	err := svc.RemovePolicy(context.Background(), p)
	require.NoError(t, err)

	assert.Empty(t, inv.tenants, "platform domain must not trigger tenant session invalidation")
}

// ─── AUTHZ-TXN: Compensation on partial failure ───────────────────────────────

// AUTHZ-TXN-1: AssignRole compensates Casbin state when repo.Upsert fails.
func TestAuthzTxn1_AssignRoleCompensatesOnRepoFailure(t *testing.T) {
	repo := &failRepo{upsertErr: errors.New("unique constraint")}
	svc := newSvcFull(t, repo, watcher.NoopWatcher{}, nil)

	tenantID, _, subject, dom := tenantCtx()
	err := svc.AssignRole(context.Background(), tenantID.String(), subject, "editor", dom,
		domain.WithAssignedBy("platform:system"))
	require.Error(t, err)

	// Casbin in-memory must be rolled back — role must NOT exist.
	has, _ := svc.HasRole(context.Background(), subject, "editor", dom)
	assert.False(t, has, "compensation must remove Casbin grouping policy after repo failure")
}

// AUTHZ-TXN-2: RevokeRole compensates Casbin state when repo.Deactivate fails.
func TestAuthzTxn2_RevokeRoleCompensatesOnRepoFailure(t *testing.T) {
	// Assign succeeds (no error on upsert yet).
	repo := &failRepo{}
	svc := newSvcFull(t, repo, watcher.NoopWatcher{}, nil)

	tenantID, _, subject, dom := tenantCtx()
	_ = svc.AssignRole(context.Background(), tenantID.String(), subject, "editor", dom,
		domain.WithAssignedBy("platform:system"))

	// Now make deactivate fail.
	repo.deactivateErr = errors.New("db deadlock")
	err := svc.RevokeRole(context.Background(), subject, "editor", dom)
	require.Error(t, err)

	// Casbin in-memory must be restored — role MUST still exist.
	has, _ := svc.HasRole(context.Background(), subject, "editor", dom)
	assert.True(t, has, "compensation must re-add Casbin grouping policy after repo failure")
}
