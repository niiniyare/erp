package stages

import (
	"fmt"

	"awo.so/internal/core/iam/contract"
	"awo.so/internal/pipeline"
	"awo.so/internal/web/authz"
	"awo.so/internal/web/ui"
)

// ─── TASK 2 — AUTHZ STAGE ────────────────────────────────────────────────────
//
// AuthzStage bulk-resolves all UI permissions and feature flags for the session.
// It is the ONLY place in the UI pipeline that calls the IAM authz service.
//
// DESCRIPTION:
// Calls UIAuthzService.BulkEnforce(ctx, sc) to evaluate every permission in
// authz.AllUIPermissions via Casbin in one batch. Computes stable fingerprints
// for the permission set and feature flag set. Constructs UISessionContext and
// stores it in opCtx.Data for the CompileStage.
//
// WHY:
// Centralises all permission resolution into one auditable, observable stage.
// The resulting permission map never leaves the pipeline — PageFn receives it
// only through UISessionContext.Can(), never as a raw map.
//
// IMPLEMENTATION:
// UIAuthzService is injected at wire time. The stage reads contract.SessionContext
// from opCtx.Ctx (already validated by SessionStage). BulkEnforce returns
// map[string]bool. Fingerprints are computed by authz.PermissionFingerprint() and
// authz.FlagFingerprint() using deterministic sorted-key SHA256.
//
// RISKS:
// Casbin call latency. Mitigate with connection pooling on the enforcer.
// UIAuthzService failure returns ErrPermissionResolution — pipeline aborts.
// Do not cache the permission map here; CacheStage handles schema caching.

// AuthzStage is Priority 20, Required true.
type AuthzStage struct {
	pipeline.BaseStage
	authzSvc authz.UIAuthzService
}

// NewAuthzStage constructs an AuthzStage with the given UIAuthzService.
// authzSvc must not be nil.
func NewAuthzStage(svc authz.UIAuthzService) *AuthzStage {
	return &AuthzStage{
		BaseStage: pipeline.BaseStage{
			StageName:       "ui.authz",
			StageOperations: []string{ui.OperationKey, ui.AppOperationKey},
			StagePriority:   ui.PriorityAuthz,
			StageRequired:   true,
			StageDependsOn:  []string{"ui.session"},
		},
		authzSvc: svc,
	}
}

// Execute resolves permissions, computes fingerprints, and builds UISessionContext.
func (s *AuthzStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
	sc, ok := contract.FromContext(opCtx.Ctx)
	if !ok || sc.IsZero() {
		// SessionStage should have caught this — defensive check.
		return pipeline.StageResult{}, fmt.Errorf("%w: authz stage reached without session", ui.ErrUnauthenticated)
	}

	perms, err := s.authzSvc.BulkEnforce(opCtx.Ctx, sc)
	if err != nil {
		return pipeline.StageResult{}, fmt.Errorf("%w: %v", ui.ErrPermissionResolution, err)
	}

	permFP := authz.PermissionFingerprint(perms)
	flagFP := authz.FlagFingerprint(sc, authz.UIFlagList)

	uiSess := ui.NewUISessionContext(sc, perms)

	return pipeline.StageResult{
		Status: "completed",
		Message: fmt.Sprintf(
			"resolved %d permissions for user %s (permFP=%s flagFP=%s)",
			len(perms), sc.UserID(), permFP, flagFP,
		),
		Outputs: map[string]any{
			ui.DataKeyPermissions:     perms,
			ui.DataKeyPermFingerprint: permFP,
			ui.DataKeyFlagFingerprint: flagFP,
			ui.DataKeySessionCtx:      uiSess,
		},
	}, nil
}

var _ pipeline.Stage = (*AuthzStage)(nil)
