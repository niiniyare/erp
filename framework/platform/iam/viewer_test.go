package iam

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func makeViewer(roles, perms []string) *SessionViewer {
	return newSessionViewer(&Session{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		TenantID:    uuid.New(),
		OrgUnitID:   uuid.Nil,
		Roles:       roles,
		Permissions: perms,
		IssuedAt:    time.Now(),
		ExpiresAt:   time.Now().Add(8 * time.Hour),
	})
}

func TestSessionViewer_HasRole(t *testing.T) {
	v := makeViewer([]string{"tenant-admin", "tenant-manager"}, nil)
	if !v.HasRole("tenant-admin") {
		t.Error("want true for tenant-admin")
	}
	if v.HasRole("nonexistent") {
		t.Error("want false for nonexistent role")
	}
}

func TestSessionViewer_HasPermission(t *testing.T) {
	v := makeViewer(nil, []string{"fms.shifts.read", "fms.shifts.submit"})
	if !v.HasPermission("fms.shifts.read") {
		t.Error("want true for fms.shifts.read")
	}
	if v.HasPermission("finance.accounts.write") {
		t.Error("want false for permission not in snapshot")
	}
}

func TestSessionViewer_WildcardPermission(t *testing.T) {
	v := makeViewer(nil, []string{"*.*.*"})
	if !v.HasPermission("anything.goes.here") {
		t.Error("wildcard *.*.* should grant HasPermission for any string")
	}
}

func TestSessionViewer_IsNotSystem(t *testing.T) {
	v := makeViewer(nil, nil)
	if v.IsSystem() {
		t.Error("SessionViewer must not be system")
	}
}

func TestSessionViewer_OrgUnitNil(t *testing.T) {
	v := makeViewer(nil, nil)
	if v.OrgUnitID() != uuid.Nil {
		t.Error("want uuid.Nil for tenant-wide viewer")
	}
}
