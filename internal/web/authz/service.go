// Package authz provides the UIAuthzService — the single allowed Casbin
// access point for the UI compilation pipeline.
//
// This package is the ONLY place in internal/web/ that may import
// github.com/casbin/casbin/v2. All other UI packages receive pre-resolved
// permission maps (map[string]bool) via UISessionContext.
//
// Contract boundary: UIAuthzService accepts contract.SessionContext and
// returns a plain map — no IAM domain types cross into the UI layer.
package authz

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"

	"awo.so/internal/core/iam/contract"
)

// UIPermission is a canonical UI permission string in "resource.action" format.
type UIPermission string

// AllUIPermissions is the complete set of UI permissions evaluated by BulkEnforce.
// Every permission referenced by any PageFn, NavFn, or UIBlock MUST be listed here.
// Add entries here when adding a new page or action that requires access control.
//
// Failure to list a permission here means BulkEnforce will never evaluate it,
// and UISessionContext.Can() will always return false for that permission.
var AllUIPermissions = []UIPermission{
	// Finance — Invoices
	"invoice.read",
	"invoice.create",
	"invoice.update",
	"invoice.delete",
	"invoice.approve",
	"invoice.void",

	// Finance — Accounts
	"account.read",
	"account.create",
	"account.update",

	// Finance — Payments
	"payment.read",
	"payment.create",
	"payment.approve",

	// Finance — Journal
	"journal.read",
	"journal.create",
	"journal.approve",

	// Finance — Reports
	"report.read",
	"report.export",

	// Inventory
	"inventory.read",
	"inventory.create",
	"inventory.update",
	"inventory.adjust",

	// HR
	"employee.read",
	"employee.create",
	"employee.update",

	// IAM — Users
	"user.read",
	"user.create",
	"user.update",
	"user.delete",

	// IAM — Roles
	"role.read",
	"role.create",
	"role.update",
	"role.delete",

	// Platform
	"tenant.read",
	"tenant.update",
	"settings.read",
	"settings.update",
	"auditlog.read",

	// Dashboard
	"dashboard.read",
}

// UIAuthzService resolves UI permissions for a session.
// Implemented by CasbinUIAuthzService in production and MockUIAuthzService in tests.
type UIAuthzService interface {
	// BulkEnforce resolves all AllUIPermissions for the session.
	// Returns map[UIPermission]bool. An absent key means false.
	// Called once per cache miss by AuthzStage.
	BulkEnforce(ctx context.Context, sc contract.SessionContext) (map[string]bool, error)
}

// PermissionFingerprint returns a stable hash of the true-valued permissions.
// Two sessions with identical role sets produce identical fingerprints.
// Only true-valued entries contribute — false values do not differentiate sessions.
func PermissionFingerprint(perms map[string]bool) string {
	keys := make([]string, 0, len(perms))
	for k, v := range perms {
		if v {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys) // deterministic order regardless of map iteration
	h := sha256.Sum256([]byte(strings.Join(keys, ",")))
	return fmt.Sprintf("%x", h[:8]) // 16-char hex — sufficient for cache keying
}

// FlagFingerprint returns a stable hash of enabled feature flags.
func FlagFingerprint(sc contract.SessionContext, flags []UIPermission) string {
	enabled := make([]string, 0, len(flags))
	for _, f := range flags {
		if sc.FeatureEnabled(string(f)) {
			enabled = append(enabled, string(f))
		}
	}
	sort.Strings(enabled)
	h := sha256.Sum256([]byte(strings.Join(enabled, ",")))
	return fmt.Sprintf("%x", h[:8])
}

// ─── CasbinUIAuthzService ─────────────────────────────────────────────────────

// CasbinUIAuthzService implements UIAuthzService using the Casbin enforcer.
// This is the production implementation.
//
// It constructs the Casbin subject from contract.SessionContext methods only —
// no internal IAM types are used. Subject format:
//   - Platform actor: "platform:<userID>"
//   - Portal actor:   "portal:<userID>"
//   - Tenant actor:   "user:<userID>"
type CasbinUIAuthzService struct {
	enforcer *casbin.Enforcer
}

// NewCasbinUIAuthzService creates a CasbinUIAuthzService.
// The enforcer must be fully initialised (model + adapter loaded).
func NewCasbinUIAuthzService(enforcer *casbin.Enforcer) *CasbinUIAuthzService {
	return &CasbinUIAuthzService{enforcer: enforcer}
}

// BulkEnforce evaluates all AllUIPermissions for the session in one pass.
// Uses Casbin BatchEnforce to minimise enforcer round-trips.
func (s *CasbinUIAuthzService) BulkEnforce(
	_ context.Context,
	sc contract.SessionContext,
) (map[string]bool, error) {
	subject := subjectFor(sc)
	domain := sc.TenantID().String()

	requests := make([][]any, len(AllUIPermissions))
	for i, perm := range AllUIPermissions {
		parts := strings.SplitN(string(perm), ".", 2)
		if len(parts) != 2 {
			continue
		}
		resource := domain + ":" + parts[0]
		action := parts[1]
		requests[i] = []any{subject, resource, action}
	}

	results, err := s.enforcer.BatchEnforce(requests)
	if err != nil {
		return nil, fmt.Errorf("authz: batch enforce failed: %w", err)
	}

	perms := make(map[string]bool, len(AllUIPermissions))
	for i, perm := range AllUIPermissions {
		if i < len(results) {
			perms[string(perm)] = results[i]
		}
	}
	return perms, nil
}

// subjectFor constructs the Casbin subject string from contract.SessionContext.
// Mirrors the logic in iamdomain.ResolvedSession.ToPrincipal() without importing it.
func subjectFor(sc contract.SessionContext) string {
	switch {
	case sc.IsPlatform():
		return "platform:" + sc.UserID().String()
	case sc.IsPortal():
		return "portal:" + sc.UserID().String()
	default:
		return "user:" + sc.UserID().String()
	}
}

// ─── MockUIAuthzService ───────────────────────────────────────────────────────

// MockUIAuthzService is a test double for UIAuthzService.
// Returns a pre-configured permission map for any session.
type MockUIAuthzService struct {
	Perms map[string]bool
	Err   error
}

// BulkEnforce returns the pre-configured permission map.
func (m *MockUIAuthzService) BulkEnforce(_ context.Context, _ contract.SessionContext) (map[string]bool, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	result := make(map[string]bool, len(m.Perms))
	for k, v := range m.Perms {
		result[k] = v
	}
	return result, nil
}

// AllPermsGranted returns a MockUIAuthzService where every AllUIPermission is true.
func AllPermsGranted() *MockUIAuthzService {
	perms := make(map[string]bool, len(AllUIPermissions))
	for _, p := range AllUIPermissions {
		perms[string(p)] = true
	}
	return &MockUIAuthzService{Perms: perms}
}

// NoPermsGranted returns a MockUIAuthzService where every permission is false.
func NoPermsGranted() *MockUIAuthzService {
	return &MockUIAuthzService{Perms: map[string]bool{}}
}

// GrantOnly returns a MockUIAuthzService with only the specified permissions true.
func GrantOnly(perms ...string) *MockUIAuthzService {
	m := make(map[string]bool, len(perms))
	for _, p := range perms {
		m[p] = true
	}
	return &MockUIAuthzService{Perms: m}
}

// Ensure compile-time conformance.
var (
	_ UIAuthzService = (*CasbinUIAuthzService)(nil)
	_ UIAuthzService = (*MockUIAuthzService)(nil)
)

// UIFlagList is the canonical set of feature flags used in fingerprinting.
// Mirrors allKnownFlags in ui/types.go without the circular import.
var UIFlagList = []UIPermission{
	"advanced_reporting",
	"bulk_import",
	"multi_currency",
	"approval_workflow",
	"ai_assist",
	"billing.autopay",
	"hr.payroll_v2",
	"inventory.lot_tracking",
	"finance.auto_reconcile",
}

// SchemaVersion is kept for backward compatibility with existing cached keys.
// New code must use uicache.Key() with CacheVersions instead.
//
// Deprecated: use internal/web/cache.Key with CacheVersions.
const SchemaVersion = "v1"

// CacheKey builds a schema cache key using only static version components.
//
// Deprecated: use internal/web/cache.Key(tenantID, route, permFP, flagFP, versions)
// which includes PolicyGeneration and SchemaGeneration for runtime soft-invalidation.
// This function remains for test helpers and migration path only.
func CacheKey(tenantID, route, permFP, flagFP string) string {
	route = strings.TrimPrefix(route, "/")
	route = strings.ReplaceAll(route, "/", ":")
	return fmt.Sprintf("ui:schema:%s:%s:%s:%s:%s",
		tenantID, route, permFP, flagFP, SchemaVersion)
}

// InvalidationPattern returns the Redis key pattern to delete all schema
// cache entries for a given tenant. Used when roles change.
//
// Deprecated: use internal/web/cache.TenantPattern(tenantID.String()).
func InvalidationPattern(tenantID uuid.UUID) string {
	return fmt.Sprintf("ui:schema:%s:*", tenantID.String())
}
