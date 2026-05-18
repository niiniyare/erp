// Package ui defines the canonical types for the Go-first UI compilation system.
//
// All non-UI packages interact with the UI layer exclusively through these types.
// No UI package may define its own parallel type for PageFn, Schema, or UIBlock.
package ui

import "awo.so/internal/core/iam/contract"

// M is the AMIS schema map. Alias for map[string]any.
// Every AMIS node is an M. Use this instead of the raw type to make
// intent clear and to allow future type-level migration.
type M = map[string]any

// A is the AMIS schema array. Alias for []any.
type A = []any

// Schema is the root compiled AMIS schema returned by every PageFn.
// It is M by definition; the alias signals "this is the document root."
type Schema = M

// PageFn is the signature every page builder function must satisfy.
//
// Invariants (enforced by pipeline, not compiler):
//   - Pure function: same UISessionContext produces identical Schema.
//   - No I/O: no DB calls, no HTTP calls, no file reads.
//   - No IAM access: UISessionContext is the only allowed identity source.
//   - No goroutines: PageFn executes synchronously within the pipeline.
//
// Violations are caught by SchemaValidator at compile time (CI) or
// at request time (NormalizeStage aborts the pipeline).
type PageFn func(sess UISessionContext) Schema

// NavFn builds the permission-filtered navigation tree for the app shell.
// Same purity invariants as PageFn.
type NavFn func(sess UISessionContext) []M

// UIBlock is a reusable, session-aware schema fragment.
// Blocks compose inside page functions to build ERP dashboard widgets,
// approval inboxes, summary panels, etc.
type UIBlock func(sess UISessionContext) M

// UISessionContext is the read-only view of session + resolved IAM state
// that PageFn, NavFn, and UIBlock functions receive.
//
// It is constructed exclusively by the AuthzStage after permission resolution.
// It MUST NOT be constructed anywhere else.
//
// Contract enforcements:
//   - No Can() call here leads to Casbin — permissions are pre-resolved.
//   - No reference to *iam.ResolvedSession or any IAM internal type.
//   - Feature flags come from contract.SessionContext (pre-resolved at login).
//   - Permissions come from UIAuthzService.BulkEnforce() (Casbin, once per request).
type UISessionContext struct {
	// Identity — sourced from contract.SessionContext
	UserID      string
	TenantID    string
	DisplayName string
	IsPlatform  bool
	IsPortal    bool

	// Locale settings — sourced from contract.SessionContext.Preference() / Setting()
	Locale   string // e.g. "en-GB"
	Timezone string // e.g. "Africa/Nairobi"
	Currency string // e.g. "KES"

	// private: populated by AuthzStage only
	permissions  map[string]bool
	featureFlags map[string]bool
	prefs        map[string]string
}

// Can returns true iff the session has the resolved permission "resource.action".
// Permissions are pre-resolved by AuthzStage via UIAuthzService.BulkEnforce().
// This is NOT a Casbin call — it reads the pre-computed map.
func (u UISessionContext) Can(action, resource string) bool {
	return u.permissions[resource+"."+action]
}

// CanAny returns true iff the session has the named permission for any resource.
func (u UISessionContext) CanAny(action string, resources ...string) bool {
	for _, r := range resources {
		if u.Can(action, r) {
			return true
		}
	}
	return false
}

// CanAll returns true iff the session has the named permission for every resource.
func (u UISessionContext) CanAll(action string, resources ...string) bool {
	for _, r := range resources {
		if !u.Can(action, r) {
			return false
		}
	}
	return true
}

// Flag returns true iff the named feature flag is enabled for this session.
// Source: contract.SessionContext.FeatureEnabled() — pre-resolved at login.
// No feature flag service is called here.
func (u UISessionContext) Flag(name string) bool {
	return u.featureFlags[name]
}

// Pref returns a user UI preference string.
// Source: contract.SessionContext.Preference() — pre-resolved at login.
func (u UISessionContext) Pref(key, fallback string) string {
	if v, ok := u.prefs[key]; ok {
		return v
	}
	return fallback
}

// NewUISessionContext constructs a UISessionContext from a resolved contract session
// and a pre-computed permissions map.
//
// This is the ONLY constructor. Called exclusively by AuthzStage.
// All other callers should receive UISessionContext as a function argument.
//
// flags and prefs are copied to avoid holding a reference to the contract session
// beyond the AuthzStage — UISessionContext may be passed into cached schemas.
func NewUISessionContext(
	sc contract.SessionContext,
	permissions map[string]bool,
) UISessionContext {
	// Deep-copy flags — contract.SessionContext.FeatureEnabled iterates the
	// underlying map; we snapshot here so UISessionContext is self-contained.
	flags := make(map[string]bool)
	for _, flag := range allKnownFlags {
		flags[flag] = sc.FeatureEnabled(flag)
	}

	prefs := map[string]string{
		"ui.theme":    sc.Preference("ui.theme", "light"),
		"ui.locale":   sc.Preference("ui.locale", "en"),
		"ui.timezone": sc.Preference("ui.timezone", "UTC"),
	}

	return UISessionContext{
		UserID:      sc.UserID().String(),
		TenantID:    sc.TenantID().String(),
		DisplayName: sc.DisplayName(),
		IsPlatform:  sc.IsPlatform(),
		IsPortal:    sc.IsPortal(),
		Locale:      sc.Preference("ui.locale", "en"),
		Timezone:    sc.Preference("ui.timezone", "UTC"),
		Currency:    sc.Setting("tenant.currency", "USD"),
		permissions:  permissions,
		featureFlags: flags,
		prefs:        prefs,
	}
}

// allKnownFlags is used by NewUISessionContext to snapshot the session flags.
// Add entries here when adding a new feature-gated UI behaviour.
// This list is separate from AllUIFlags in the authz package to avoid circular imports.
var allKnownFlags = []string{
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

