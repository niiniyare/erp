package service

import (
	"context"
	"errors"
	"testing"

	"awo.so/internal/core/iam/domain"
	"awo.so/internal/shared/logger"
)

// discardLogger is a no-op logger for unit tests that do not care about log
// output.  All methods are stubs; WithFields returns the same logger so that
// the logger.WithFields("component", ...) chain inside NewInMemoryAuthzService
// does not panic.
type discardLogger struct{}

func (d *discardLogger) Debug(_ string, _ ...logger.Fields)                         {}
func (d *discardLogger) Info(_ string, _ ...logger.Fields)                          {}
func (d *discardLogger) Warn(_ string, _ ...logger.Fields)                          {}
func (d *discardLogger) Error(_ string, _ ...logger.Fields)                         {}
func (d *discardLogger) Fatal(_ string, _ ...logger.Fields)                         {}
func (d *discardLogger) DebugContext(_ context.Context, _ string, _ ...logger.Fields) {}
func (d *discardLogger) InfoContext(_ context.Context, _ string, _ ...logger.Fields)  {}
func (d *discardLogger) WarnContext(_ context.Context, _ string, _ ...logger.Fields)  {}
func (d *discardLogger) ErrorContext(_ context.Context, _ string, _ ...logger.Fields) {}
func (d *discardLogger) WithFields(_ logger.Fields) logger.Logger                   { return d }
func (d *discardLogger) WithContext(_ context.Context) logger.Logger                { return d }
func (d *discardLogger) SetLevel(_ logger.LogLevel)                                 {}
func (d *discardLogger) Close() error                                               { return nil }

func newTestAuthzService(t *testing.T) AuthzService {
	t.Helper()
	svc, err := NewInMemoryAuthzService(nil, &discardLogger{})
	if err != nil {
		t.Fatalf("NewInMemoryAuthzService: %v", err)
	}
	return svc
}

// ---------------------------------------------------------------------------
// AZ-SEC-050 — tenant actor cannot write _platform_ policy
// ---------------------------------------------------------------------------

// TestAddPolicy_PlatformDomainGuard verifies that a non-platform subject cannot
// add a policy in the _platform_ domain, regardless of other fields.
func TestAddPolicy_PlatformDomainGuard(t *testing.T) {
	svc := newTestAuthzService(t)
	ctx := context.Background()

	err := svc.AddPolicy(ctx, domain.Policy{
		Subject: "tenant:some-user-uuid",
		Domain:  domain.DomainPlatform,
		Object:  "system.*",
		Action:  "read",
		Effect:  "allow",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("want ErrForbidden, got %v", err)
	}
}

// TestAddPolicy_PlatformSubjectAllowed verifies that a platform subject CAN
// add a policy in the _platform_ domain.
func TestAddPolicy_PlatformSubjectAllowed(t *testing.T) {
	svc := newTestAuthzService(t)
	ctx := context.Background()

	err := svc.AddPolicy(ctx, domain.Policy{
		Subject: "platform:admin-uuid",
		Domain:  domain.DomainPlatform,
		Object:  "system.*",
		Action:  "read",
		Effect:  "allow",
	})
	if err != nil {
		t.Errorf("platform subject should be allowed, got error: %v", err)
	}
}

// TestRemovePolicy_PlatformDomainGuard verifies that a non-platform subject
// cannot remove a policy from the _platform_ domain.
func TestRemovePolicy_PlatformDomainGuard(t *testing.T) {
	svc := newTestAuthzService(t)
	ctx := context.Background()

	err := svc.RemovePolicy(ctx, domain.Policy{
		Subject: "tenant:attacker-uuid",
		Domain:  domain.DomainPlatform,
		Object:  "system.*",
		Action:  "delete",
		Effect:  "allow",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("want ErrForbidden, got %v", err)
	}
}

// TestAssignRole_PlatformRoleGuard verifies that a non-platform actor cannot
// assign a "role:platform-*" role.
func TestAssignRole_PlatformRoleGuard(t *testing.T) {
	svc := newTestAuthzService(t)
	ctx := context.Background()

	err := svc.AssignRole(ctx,
		"tenant-uuid",
		"tenant:victim-uuid",
		"role:platform-admin",
		domain.DomainPlatform,
		domain.WithAssignedBy("tenant:attacker-uuid"),
	)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("want ErrForbidden, got %v", err)
	}
}

// TestAssignRole_RequiresAssignedBy verifies that AssignedBy is mandatory.
func TestAssignRole_RequiresAssignedBy(t *testing.T) {
	svc := newTestAuthzService(t)
	ctx := context.Background()

	// No WithAssignedBy option — should fail validation.
	err := svc.AssignRole(ctx,
		"tenant-uuid",
		"tenant:some-user",
		"role:accountant",
		"tenant-uuid",
	)
	if err == nil {
		t.Error("expected error when AssignedBy is empty, got nil")
	}
}
