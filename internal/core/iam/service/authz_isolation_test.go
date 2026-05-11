package service

import (
	"context"
	"testing"

	"awo.so/internal/core/iam/domain"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func seedPolicy(t *testing.T, svc AuthzService, sub, dom, obj, act string) {
	t.Helper()
	err := svc.AddPolicy(context.Background(), domain.Policy{
		Subject: sub, Domain: dom, Object: obj, Action: act, Effect: "allow",
	})
	if err != nil && err != domain.ErrPolicyConflict {
		t.Fatalf("seed policy: %v", err)
	}
}

func mustEnforce(t *testing.T, svc AuthzService, sub, dom, obj, act string) bool {
	t.Helper()
	allowed, err := svc.Enforce(context.Background(), domain.Request{
		Subject: sub, Domain: dom, Object: obj, Action: act,
	})
	if err != nil {
		t.Fatalf("Enforce error: %v", err)
	}
	return allowed
}

// ---------------------------------------------------------------------------
// AZ-ISO-001 — Policy does not cross domain
// ---------------------------------------------------------------------------

func TestDomainIsolation_PolicyDoesNotCrossDomain(t *testing.T) {
	svc := newTestAuthzService(t)

	tenantA := "tenant-a-uuid"
	tenantB := "tenant-b-uuid"
	sub := domain.TenantSubject("user-1")

	seedPolicy(t, svc, sub, domain.TenantDomain(tenantA), "invoice", "read")

	if !mustEnforce(t, svc, sub, domain.TenantDomain(tenantA), "invoice", "read") {
		t.Error("should be allowed in tenant A")
	}
	if mustEnforce(t, svc, sub, domain.TenantDomain(tenantB), "invoice", "read") {
		t.Error("policy in tenant A must not bleed into tenant B")
	}
}

// ---------------------------------------------------------------------------
// AZ-ISO-002 — Role does not cross domain
// ---------------------------------------------------------------------------

func TestDomainIsolation_RoleDoesNotCrossDomain(t *testing.T) {
	svc := newTestAuthzService(t)
	ctx := context.Background()

	tenantA := "tenant-a-uuid"
	tenantB := "tenant-b-uuid"
	role := "accountant"
	sub := domain.TenantSubject("user-2")

	seedPolicy(t, svc, role, domain.TenantDomain(tenantA), "ledger", "read")
	if err := svc.AssignRole(ctx, tenantA, sub, role, domain.TenantDomain(tenantA),
		domain.WithAssignedBy("platform:system")); err != nil {
		t.Fatalf("AssignRole: %v", err)
	}

	if !mustEnforce(t, svc, sub, domain.TenantDomain(tenantA), "ledger", "read") {
		t.Error("should be allowed in tenant A via role")
	}
	if mustEnforce(t, svc, sub, domain.TenantDomain(tenantB), "ledger", "read") {
		t.Error("role in tenant A must not grant access in tenant B")
	}
}

// ---------------------------------------------------------------------------
// AZ-ISO-003 — Platform domain vs tenant domain
// ---------------------------------------------------------------------------

func TestDomainIsolation_PlatformVsTenant(t *testing.T) {
	svc := newTestAuthzService(t)

	platformSub := domain.PlatformSubject("sysadmin-uuid")
	tenantSub := domain.TenantSubject("user-uuid")
	tenantDom := domain.TenantDomain("some-tenant")

	seedPolicy(t, svc, platformSub, domain.DomainPlatform, "tenants", "manage")
	seedPolicy(t, svc, tenantSub, tenantDom, "invoice", "read")

	// Platform policy must not grant access in tenant domain.
	if mustEnforce(t, svc, platformSub, tenantDom, "tenants", "manage") {
		t.Error("platform policy must not work in tenant domain")
	}
	// Tenant policy must not grant access in platform domain.
	if mustEnforce(t, svc, tenantSub, domain.DomainPlatform, "invoice", "read") {
		t.Error("tenant policy must not work in platform domain")
	}
}

// ---------------------------------------------------------------------------
// AZ-ISO-004 — Tenant domain vs portal domain
// ---------------------------------------------------------------------------

func TestDomainIsolation_TenantVsPortal(t *testing.T) {
	svc := newTestAuthzService(t)

	tid := "tenant-uuid"
	internalSub := domain.TenantSubject("employee-uuid")
	portalSub := domain.PortalSubject("customer-uuid")

	seedPolicy(t, svc, internalSub, domain.TenantDomain(tid), "payroll", "read")
	seedPolicy(t, svc, portalSub, domain.PortalDomain(tid), "orders", "read")

	// Internal policy must not work in portal domain.
	if mustEnforce(t, svc, internalSub, domain.PortalDomain(tid), "payroll", "read") {
		t.Error("internal policy must not bleed into portal domain")
	}
	// Portal policy must not work in internal tenant domain.
	if mustEnforce(t, svc, portalSub, domain.TenantDomain(tid), "orders", "read") {
		t.Error("portal policy must not bleed into tenant domain")
	}
}

// ---------------------------------------------------------------------------
// AZ-ISO-005 — Tenant domain vs API domain
// ---------------------------------------------------------------------------

func TestDomainIsolation_TenantVsAPI(t *testing.T) {
	svc := newTestAuthzService(t)

	tid := "tenant-uuid"
	tenantSub := domain.TenantSubject("user-uuid")
	apiSub := domain.APISubject("client-id")

	seedPolicy(t, svc, tenantSub, domain.TenantDomain(tid), "report", "export")
	seedPolicy(t, svc, apiSub, domain.APIDomain(tid), "webhook", "receive")

	if mustEnforce(t, svc, tenantSub, domain.APIDomain(tid), "report", "export") {
		t.Error("tenant policy must not bleed into API domain")
	}
	if mustEnforce(t, svc, apiSub, domain.TenantDomain(tid), "webhook", "receive") {
		t.Error("API policy must not bleed into tenant domain")
	}
}

// ---------------------------------------------------------------------------
// AZ-ISO-010 — Cross-tenant isolation
// ---------------------------------------------------------------------------

func TestDomainIsolation_CrossTenant(t *testing.T) {
	svc := newTestAuthzService(t)

	subA := domain.TenantSubject("alice-uuid")
	subB := domain.TenantSubject("bob-uuid")
	domA := domain.TenantDomain("company-a")
	domB := domain.TenantDomain("company-b")

	seedPolicy(t, svc, subA, domA, "finance.*", "read")

	if !mustEnforce(t, svc, subA, domA, "finance.*", "read") {
		t.Error("alice should have access in company-a")
	}
	if mustEnforce(t, svc, subA, domB, "finance.*", "read") {
		t.Error("alice must not have access in company-b")
	}
	if mustEnforce(t, svc, subB, domA, "finance.*", "read") {
		t.Error("bob must not have access in company-a")
	}
}

// ---------------------------------------------------------------------------
// AZ-ISO-020 — Wildcard subject is domain-scoped
// ---------------------------------------------------------------------------

func TestDomainIsolation_WildcardSubjectDomainScoped(t *testing.T) {
	svc := newTestAuthzService(t)

	domA := domain.TenantDomain("company-a")
	domB := domain.TenantDomain("company-b")

	// Wildcard subject in domain A.
	seedPolicy(t, svc, "*", domA, "public.announcements", "read")

	anyUser := domain.TenantSubject("random-uuid")

	if !mustEnforce(t, svc, anyUser, domA, "public.announcements", "read") {
		t.Error("wildcard should grant any subject in domain A")
	}
	if mustEnforce(t, svc, anyUser, domB, "public.announcements", "read") {
		t.Error("wildcard in domain A must not apply to domain B")
	}
}

// ---------------------------------------------------------------------------
// T-SEC: TestPolicyCountLimit_Enforced (AUTHZ-5)
// ---------------------------------------------------------------------------

func TestPolicyCountLimit_Enforced(t *testing.T) {
	svc, err := NewInMemoryAuthzServiceWithLimit(nil, &discardLogger{}, 3)
	if err != nil {
		t.Fatalf("NewInMemoryAuthzServiceWithLimit: %v", err)
	}
	ctx := context.Background()
	dom := domain.TenantDomain("test-tenant")

	// Seed 3 policies — should all succeed.
	for i := range 3 {
		err := svc.AddPolicy(ctx, domain.Policy{
			Subject: domain.TenantSubject("role-x"),
			Domain:  dom,
			Object:  "resource-" + string(rune('a'+i)),
			Action:  "read",
			Effect:  "allow",
		})
		if err != nil {
			t.Fatalf("policy %d: unexpected error: %v", i+1, err)
		}
	}

	// 4th policy must be rejected.
	err = svc.AddPolicy(ctx, domain.Policy{
		Subject: domain.TenantSubject("role-x"),
		Domain:  dom,
		Object:  "resource-d",
		Action:  "read",
		Effect:  "allow",
	})
	if err != domain.ErrPolicyLimitExceeded {
		t.Errorf("want ErrPolicyLimitExceeded, got %v", err)
	}
}
