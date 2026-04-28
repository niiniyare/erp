package featureflag

//
//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/audit"
	"awo.so/internal/core/tenant"
	"awo.so/internal/shared/convert"
	"github.com/google/uuid"
)

// Service defines the business logic interface for feature flags.
// It exposes three groups of operations:
//   - CRUD management of flag definitions
//   - Runtime evaluation of flags against an evaluation context
//   - Analytics helpers (stats, search, type filtering)
type Service interface {
	// --- Feature Flag Management ---

	// CreateFeatureFlag creates a new flag scoped to the caller's tenant.
	CreateFeatureFlag(ctx context.Context, request *CreateFeatureFlagRequest) (*FeatureFlag, error)

	// GetFeatureFlag retrieves a flag by its unique name within the current tenant.
	GetFeatureFlag(ctx context.Context, name string) (*FeatureFlag, error)

	// GetFeatureFlagByID retrieves a flag by its UUID within the current tenant.
	GetFeatureFlagByID(ctx context.Context, id uuid.UUID) (*FeatureFlag, error)

	// UpdateFeatureFlag applies a partial update to an existing flag.
	// Only non-nil fields in the request are written.
	UpdateFeatureFlag(ctx context.Context, id uuid.UUID, request *UpdateFeatureFlagRequest) (*FeatureFlag, error)

	// DeleteFeatureFlag permanently removes a flag and emits an audit event.
	DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error

	// ListFeatureFlags returns a paginated list of flags, optionally filtered by type.
	ListFeatureFlags(ctx context.Context, request *ListFeatureFlagsRequest) (*ListFeatureFlagsResponse, error)

	// --- Flag Evaluation ---

	// EvaluateFlag resolves the effective value of a single flag for the given
	// evaluation context (user, attributes, etc.).
	EvaluateFlag(ctx context.Context, name string, evalCtx *EvaluationContext) (*EvaluationResult, error)

	// EvaluateFlags resolves multiple flags in one call, returning per-flag
	// results and a map of any per-flag errors.
	EvaluateFlags(ctx context.Context, names []string, evalCtx *EvaluationContext) (*BulkEvaluationResponse, error)

	// IsEnabled is a convenience wrapper around EvaluateFlag that returns only
	// the boolean enabled/disabled state.
	IsEnabled(ctx context.Context, name string, evalCtx *EvaluationContext) (bool, error)

	// --- Analytics and Monitoring ---

	// GetFlagStats returns aggregate counts/metrics across all flags for the tenant.
	GetFlagStats(ctx context.Context) (*FlagStats, error)

	// SearchFlags performs a full-text search over flag names/descriptions.
	SearchFlags(ctx context.Context, query string, limit, offset int32) ([]*FeatureFlag, error)

	// GetFlagsByType returns all flags that match the given flag type string
	// (e.g. FlagTypeBoolean, FlagTypeString, …).
	GetFlagsByType(ctx context.Context, flagType string) ([]*FeatureFlag, error)
}

// service is the concrete implementation of Service.
// It wires together persistence, tenant isolation, auditing, and real-time
// WebSocket notifications.
type service struct {
	repository       Repository
	tenantService    tenant.Service
	store            db.Store
	auditService     audit.Service
	webSocketService WebSocketService // optional; nil-checked before use
}

// NewService constructs a service with all required dependencies injected.
// webSocketService may be nil; real-time notifications are skipped when absent.
func NewService(
	repository Repository,
	tenantService tenant.Service,
	store db.Store,
	auditService audit.Service,
	webSocketService WebSocketService,
) Service {
	return &service{
		repository:       repository,
		tenantService:    tenantService,
		store:            store,
		auditService:     auditService,
		webSocketService: webSocketService,
	}
}

// ListFeatureFlagsRequest carries pagination and optional filtering parameters
// for ListFeatureFlags.
type ListFeatureFlagsRequest struct {
	Page     int     `json:"page"      validate:"min=1"`
	PageSize int     `json:"page_size" validate:"min=1,max=100"`
	FlagType *string `json:"flag_type,omitempty"` // nil means "all types"
}

// ListFeatureFlagsResponse wraps a page of flags together with the pagination
// metadata needed by callers to implement "next page" logic.
type ListFeatureFlagsResponse struct {
	FeatureFlags []*FeatureFlag `json:"feature_flags"`
	Total        int64          `json:"total"` // total flags matching the filter (not just this page)
	Page         int            `json:"page"`
	PageSize     int            `json:"page_size"`
}

// CreateFeatureFlag creates a new feature flag scoped to the current tenant.
//
// It enforces:
//   - A valid tenant context must exist in ctx.
//   - The flag name must be unique within the tenant.
//   - The flag type must be one of the known FlagType constants.
//
// On success it emits an audit event and asynchronously notifies WebSocket
// subscribers of the new flag.
func (s *service) CreateFeatureFlag(ctx context.Context, request *CreateFeatureFlagRequest) (*FeatureFlag, error) {
	// Ensure the caller is operating within a valid tenant scope.
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	// Reject duplicate names early to surface a clear error before hitting the DB.
	existing, err := s.repository.GetFeatureFlagByName(ctx, request.Name)
	if err != nil && err != ErrFeatureFlagNotFound {
		return nil, fmt.Errorf("failed to check existing flag: %w", err)
	}
	if existing != nil {
		return nil, ErrFeatureFlagAlreadyExists
	}

	// Guard against unsupported flag types before persisting.
	switch request.FlagType {
	case FlagTypeBoolean, FlagTypeString, FlagTypeNumber, FlagTypeJSON:
		// valid — fall through
	default:
		return nil, fmt.Errorf("invalid flag type: %s", request.FlagType)
	}

	flag, err := s.repository.CreateFeatureFlag(ctx, request)
	if err != nil {
		return nil, err
	}

	s.auditFlagCreated(ctx, flag, request)

	// Notify connected WebSocket clients asynchronously so we don't block the
	// HTTP response on broadcast latency.
	if s.webSocketService != nil {
		if t, _ := s.tenantService.GetCurrentTenant(ctx); t != nil {
			go func() {
				s.webSocketService.NotifyFlagCreated(context.Background(), t.ID, flag)
			}()
		}
	}

	return flag, nil
}

// GetFeatureFlag retrieves a feature flag by name, enforcing tenant isolation.
func (s *service) GetFeatureFlag(ctx context.Context, name string) (*FeatureFlag, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	return s.repository.GetFeatureFlagByName(ctx, name)
}

// GetFeatureFlagByID retrieves a feature flag by UUID, enforcing tenant isolation.
func (s *service) GetFeatureFlagByID(ctx context.Context, id uuid.UUID) (*FeatureFlag, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	return s.repository.GetFeatureFlagByID(ctx, id)
}

// UpdateFeatureFlag applies a partial update to the flag identified by id.
//
// Side-effects (all best-effort, non-blocking):
//   - Audit event is recorded.
//   - All active sessions for the tenant are invalidated so clients receive
//     the updated flag configuration on their next login.
//   - WebSocket subscribers are notified with a typed change event whose
//     changeType reflects whether the flag was enabled, disabled, had its
//     rollout percentage changed, or had some other configuration updated.
func (s *service) UpdateFeatureFlag(ctx context.Context, id uuid.UUID, request *UpdateFeatureFlagRequest) (*FeatureFlag, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	// Fetch the pre-update snapshot so we can diff old vs. new values in the
	// WebSocket change event and audit log.
	oldFlag, err := s.repository.GetFeatureFlagByID(ctx, id)
	if err != nil {
		return nil, err
	}

	flag, err := s.repository.UpdateFeatureFlag(ctx, id, request)
	if err != nil {
		return nil, err
	}

	s.auditFlagUpdated(ctx, flag, request)

	// Invalidate sessions so every tenant user picks up fresh flag values on
	// next login. Failure here is non-fatal — the flag update itself succeeded.
	if t, _ := s.tenantService.GetCurrentTenant(ctx); t != nil {
		_ = s.store.InvalidateSessionsByTenant(ctx, t.ID)
	}

	// Build and broadcast the change event asynchronously.
	if s.webSocketService != nil {
		if t, _ := s.tenantService.GetCurrentTenant(ctx); t != nil {
			go func() {
				// Determine the semantic change type and capture old/new values
				// for the event payload.
				changeType := "update_config"
				var oldValue any = oldFlag.DefaultValue
				var newValue any = flag.DefaultValue

				if request.DefaultValue != nil && *request.DefaultValue != oldFlag.DefaultValue {
					// The enabled/disabled state flipped — use a more specific label.
					if *request.DefaultValue {
						changeType = "enable"
					} else {
						changeType = "disable"
					}
				}

				if request.RolloutPercentage != nil &&
					*request.RolloutPercentage != *oldFlag.RolloutPercentage {
					changeType = "update_rollout"
					oldValue = oldFlag.RolloutPercentage
					newValue = flag.RolloutPercentage
				}

				event := &FeatureFlagChangeEvent{
					FlagID:     flag.ID,
					FlagName:   flag.Name,
					ChangeType: changeType,
					OldValue:   oldValue,
					NewValue:   newValue,
					ChangedBy:  uuid.Nil, // TODO: extract actor from context
					AppliedAt:  time.Now(),
				}

				s.webSocketService.NotifyFlagChange(context.Background(), t.ID, event)
			}()
		}
	}

	return flag, nil
}

// DeleteFeatureFlag permanently removes the flag identified by id.
// It records a HIGH-severity audit event and asynchronously notifies WebSocket
// subscribers so dashboards can remove the flag from their view.
func (s *service) DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return fmt.Errorf("invalid tenant context: %w", err)
	}

	// Fetch first so we have the flag's metadata available for auditing even
	// after the row is gone.
	flag, err := s.repository.GetFeatureFlagByID(ctx, id)
	if err != nil {
		return err
	}

	if err = s.repository.DeleteFeatureFlag(ctx, id); err != nil {
		return err
	}

	s.auditFlagDeleted(ctx, flag)

	if s.webSocketService != nil {
		if t, _ := s.tenantService.GetCurrentTenant(ctx); t != nil {
			go func() {
				s.webSocketService.NotifyFlagDeleted(context.Background(), t.ID, flag.Name)
			}()
		}
	}

	return nil
}

// ListFeatureFlags returns a paginated list of flags for the current tenant.
//
// Page and PageSize are clamped to safe defaults if missing or out of range.
// NOTE: Total currently reflects only the count of flags in this page, not the
// true table total. A dedicated COUNT query should be added when needed.
func (s *service) ListFeatureFlags(ctx context.Context, request *ListFeatureFlagsRequest) (*ListFeatureFlagsResponse, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	// Apply safe defaults / clamp bounds.
	if request.Page < 1 {
		request.Page = 1
	}
	if request.PageSize < 1 {
		request.PageSize = 20
	}
	if request.PageSize > 100 {
		request.PageSize = 100
	}

	offset := (request.Page - 1) * request.PageSize

	// The repository layer expects int32; convert with overflow protection.
	limit, err := convert.IntToInt32(request.PageSize)
	if err != nil {
		return nil, fmt.Errorf("invalid page size: %w", err)
	}

	offsetInt32, err := convert.IntToInt32(offset)
	if err != nil {
		return nil, fmt.Errorf("invalid offset: %w", err)
	}

	flags, err := s.repository.ListFeatureFlags(ctx, ListFeatureFlagsParams{
		FlagType: request.FlagType,
		Limit:    limit,
		Offset:   offsetInt32,
	})
	if err != nil {
		return nil, err
	}

	// TODO: replace with a proper COUNT query so Total reflects the full
	// result set, not just the current page size.
	total := int64(len(flags))

	response := &ListFeatureFlagsResponse{
		FeatureFlags: flags,
		Total:        total,
		Page:         request.Page,
		PageSize:     request.PageSize,
	}

	s.auditFlagListed(ctx, request, response)

	return response, nil
}

// EvaluateFlag resolves the effective value of the named flag for evalCtx.
//
// If the flag does not exist, a safe "disabled" default is returned rather than
// an error, so callers can treat unknown flags as off without extra nil-checks.
func (s *service) EvaluateFlag(ctx context.Context, name string, evalCtx *EvaluationContext) (*EvaluationResult, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	flag, err := s.repository.GetFeatureFlagByName(ctx, name)
	if err != nil {
		if err == ErrFeatureFlagNotFound {
			// Graceful degradation: treat missing flags as disabled so callers
			// don't need to special-case the "flag doesn't exist yet" scenario.
			return &EvaluationResult{
				FlagName: name,
				Value:    false,
				Enabled:  false,
				Reason:   string(ReasonFlagNotFound),
				Metadata: ResultMetadata{
					EvaluatedAt:   time.Now(),
					CacheHit:      false,
					EvaluationMs:  0.0,
					ConfigVersion: "1.0",
				},
			}, nil
		}
		return nil, err
	}

	result := s.evaluateSimpleFlag(flag, evalCtx)

	s.auditFlagEvaluated(ctx, flag, evalCtx, result)

	return result, nil
}

// EvaluateFlags resolves multiple flags in a single call.
//
// Each flag is evaluated independently; a failure on one flag populates the
// Errors map but does not abort evaluation of the remaining flags.
func (s *service) EvaluateFlags(ctx context.Context, names []string, evalCtx *EvaluationContext) (*BulkEvaluationResponse, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	results := make(map[string]*EvaluationResult)
	errors := make(map[string]string)

	for _, name := range names {
		result, err := s.EvaluateFlag(ctx, name, evalCtx)
		if err != nil {
			errors[name] = err.Error()
		} else {
			results[name] = result
		}
	}

	response := &BulkEvaluationResponse{
		Results: results,
	}

	// Only attach the errors map when there is at least one failure to keep
	// the JSON response clean in the happy path.
	if len(errors) > 0 {
		response.Errors = errors
	}

	return response, nil
}

// IsEnabled is a thin convenience wrapper that returns only the boolean
// Enabled field from EvaluateFlag, useful for simple feature gates.
func (s *service) IsEnabled(ctx context.Context, name string, evalCtx *EvaluationContext) (bool, error) {
	result, err := s.EvaluateFlag(ctx, name, evalCtx)
	if err != nil {
		return false, err
	}
	return result.Enabled, nil
}

// GetFlagStats returns aggregate statistics (total count, enabled count, etc.)
// for all flags in the current tenant.
func (s *service) GetFlagStats(ctx context.Context) (*FlagStats, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	return s.repository.GetFlagStats(ctx)
}

// SearchFlags performs a full-text search over flag names and descriptions
// within the current tenant.
func (s *service) SearchFlags(ctx context.Context, query string, limit, offset int32) ([]*FeatureFlag, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	results, err := s.repository.SearchFlags(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	s.auditFlagSearched(ctx, query, limit, offset, results)

	return results, nil
}

// GetFlagsByType returns all flags whose type matches flagType (e.g.
// FlagTypeBoolean) within the current tenant.
func (s *service) GetFlagsByType(ctx context.Context, flagType string) ([]*FeatureFlag, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	return s.repository.GetFlagsByType(ctx, flagType)
}

// evaluateSimpleFlag implements the core flag evaluation algorithm.
//
// Evaluation order:
//  1. If a rollout percentage is configured, hash the flag name + user ID into
//     a 0-99 bucket and enable the flag only for buckets below the threshold.
//  2. Otherwise, return the flag's DefaultValue.
//
// This is intentionally simple; a rules-engine layer can be added later without
// changing the Service interface.
func (s *service) evaluateSimpleFlag(flag *FeatureFlag, evalCtx *EvaluationContext) *EvaluationResult {
	now := time.Now()

	var enabled bool
	var reason EvaluationReason
	var value any

	if flag.RolloutPercentage != nil && *flag.RolloutPercentage > 0 {
		// Deterministic hash-based bucketing: the same user always lands in the
		// same bucket, giving a stable rollout experience.
		hash := s.calculateSimpleHash(flag.Name, evalCtx.UserID)
		if hash < int(*flag.RolloutPercentage) {
			enabled = flag.DefaultValue
			reason = ReasonPercentRollout
			value = flag.DefaultValue
		} else {
			// User is outside the rollout window — treat as disabled.
			enabled = false
			reason = ReasonPercentRollout
			value = false
		}
	} else {
		// No rollout configured; use the flag's default value directly.
		enabled = flag.DefaultValue
		reason = ReasonDefaultValue
		value = flag.DefaultValue
	}

	return &EvaluationResult{
		FlagName: flag.Name,
		Value:    value,
		Enabled:  enabled,
		Reason:   string(reason),
		Metadata: ResultMetadata{
			EvaluatedAt:   now,
			CacheHit:      false,
			EvaluationMs:  1.0, // placeholder; replace with real timing if needed
			ConfigVersion: "1.0",
		},
	}
}

// calculateSimpleHash hashes a flag name + user ID to a stable integer in [0, 99].
//
// The hash is computed with a basic polynomial rolling hash (multiplier 31) and
// then normalised to the [0, 99] range. Anonymous users (nil userID) fall back
// to the literal string "anonymous" so they are consistently bucketed together.
//
// NOTE: This is intentionally simple. For production-grade distribution
// consider using a cryptographic hash such as FNV-1a or xxHash.
func (s *service) calculateSimpleHash(flagName string, userID *uuid.UUID) int {
	var input string
	if userID != nil {
		input = flagName + ":" + userID.String()
	} else {
		input = flagName + ":anonymous"
	}

	hash := 0
	for _, char := range input {
		hash = hash*31 + int(char)
	}

	// Use double-modulo to handle negative hash values from integer overflow.
	return (hash%100 + 100) % 100
}

// ---------------------------------------------------------------------------
// Audit event type, category, and severity constants
// ---------------------------------------------------------------------------

// Audit event type identifiers — stored verbatim in the audit log.
const (
	AuditEventFeatureFlagCreated   = "feature_flag_created"
	AuditEventFeatureFlagUpdated   = "feature_flag_updated"
	AuditEventFeatureFlagDeleted   = "feature_flag_deleted"
	AuditEventFeatureFlagEvaluated = "feature_flag_evaluated"
	AuditEventFeatureFlagListed    = "feature_flag_listed"
	AuditEventFeatureFlagSearched  = "feature_flag_searched"
)

// Audit category labels group related events for filtering in the audit UI.
const (
	AuditCategoryFeatureManagement = "FEATURE_MANAGEMENT"
	AuditCategoryAccess            = "ACCESS"
	AuditCategoryAdmin             = "ADMIN"
)

// Audit severity levels follow a standard INFO → WARN → HIGH → CRITICAL ladder.
// Deletions are HIGH; value/rollout changes are WARN; reads are INFO.
const (
	AuditSeverityInfo     = "INFO"
	AuditSeverityWarn     = "WARN"
	AuditSeverityHigh     = "HIGH"
	AuditSeverityCritical = "CRITICAL"
)

// auditFlagCreated emits an INFO-level audit event recording the full initial
// state of the newly created flag.
func (s *service) auditFlagCreated(ctx context.Context, flag *FeatureFlag, request *CreateFeatureFlagRequest) {
	contextData, _ := json.Marshal(map[string]any{
		"flag_id":            flag.ID.String(),
		"flag_name":          flag.Name,
		"flag_type":          string(flag.FlagType),
		"default_value":      flag.DefaultValue,
		"rollout_percentage": flag.RolloutPercentage,
		"target_audience":    flag.TargetAudience,
		"metadata":           flag.Metadata,
		"tenant_id":          flag.TenantID.String(),
	})

	decision := "CREATED"
	reason := fmt.Sprintf("Feature flag '%s' created successfully", flag.Name)
	flagEntityID := flag.ID
	s.auditService.CreateAuditEvent(ctx, audit.CreateAuditEventRequest{
		EventType:     AuditEventFeatureFlagCreated,
		EventCategory: AuditCategoryFeatureManagement,
		Severity:      AuditSeverityInfo,
		EntityID:      &flagEntityID,
		Decision:      &decision,
		Reason:        &reason,
		Context:       contextData,
	})
}

// auditFlagUpdated emits an audit event for a flag update.
// Severity is escalated to WARN when the DefaultValue or RolloutPercentage
// changes, as these directly affect end-user behaviour.
func (s *service) auditFlagUpdated(ctx context.Context, flag *FeatureFlag, request *UpdateFeatureFlagRequest) {
	contextData, _ := json.Marshal(map[string]any{
		"flag_id":        flag.ID.String(),
		"flag_name":      flag.Name,
		"updates":        request,
		"tenant_id":      flag.TenantID.String(),
		"updated_fields": s.getUpdatedFields(request),
	})

	severity := AuditSeverityInfo
	if request.DefaultValue != nil || request.RolloutPercentage != nil {
		severity = AuditSeverityWarn // behavioural changes warrant a warning
	}

	decision := "UPDATED"
	reason := fmt.Sprintf("Feature flag '%s' updated successfully", flag.Name)
	flagEntityID := flag.ID
	s.auditService.CreateAuditEvent(ctx, audit.CreateAuditEventRequest{
		EventType:     AuditEventFeatureFlagUpdated,
		EventCategory: AuditCategoryFeatureManagement,
		Severity:      severity,
		EntityID:      &flagEntityID,
		Decision:      &decision,
		Reason:        &reason,
		Context:       contextData,
	})
}

// auditFlagDeleted emits a HIGH-severity audit event.
// Deletions are irreversible, so they receive elevated severity to ensure they
// surface prominently in compliance and security reviews.
func (s *service) auditFlagDeleted(ctx context.Context, flag *FeatureFlag) {
	contextData, _ := json.Marshal(map[string]any{
		"flag_id":    flag.ID.String(),
		"flag_name":  flag.Name,
		"flag_type":  string(flag.FlagType),
		"tenant_id":  flag.TenantID.String(),
		"deleted_at": time.Now(),
	})

	decision := "DELETED"
	reason := fmt.Sprintf("Feature flag '%s' deleted successfully", flag.Name)
	flagEntityID := flag.ID
	s.auditService.CreateAuditEvent(ctx, audit.CreateAuditEventRequest{
		EventType:     AuditEventFeatureFlagDeleted,
		EventCategory: AuditCategoryFeatureManagement,
		Severity:      AuditSeverityHigh,
		EntityID:      &flagEntityID,
		Decision:      &decision,
		Reason:        &reason,
		Context:       contextData,
	})
}

// auditFlagEvaluated records every flag resolution, including the evaluation
// context and result metadata (cache hit, evaluation latency).
// These events feed into usage analytics and debugging workflows.
func (s *service) auditFlagEvaluated(ctx context.Context, flag *FeatureFlag, evalCtx *EvaluationContext, result *EvaluationResult) {
	contextData, _ := json.Marshal(map[string]any{
		"flag_name":          flag.Name,
		"flag_id":            flag.ID.String(),
		"evaluation_result":  result.Value,
		"enabled":            result.Enabled,
		"reason":             result.Reason,
		"rule_matched":       result.RuleMatched,
		"evaluation_context": evalCtx,
		"tenant_id":          flag.TenantID.String(),
		"evaluation_time_ms": result.Metadata.EvaluationMs,
		"cache_hit":          result.Metadata.CacheHit,
	})

	// Encode the boolean outcome in the decision string for easy log filtering.
	decision := fmt.Sprintf("EVALUATED_%t", result.Enabled)
	reason := fmt.Sprintf("Feature flag '%s' evaluated: %v", flag.Name, result.Value)
	flagEntityID := flag.ID
	s.auditService.CreateAuditEvent(ctx, audit.CreateAuditEventRequest{
		EventType:     AuditEventFeatureFlagEvaluated,
		EventCategory: AuditCategoryAccess,
		Severity:      AuditSeverityInfo,
		EntityID:      &flagEntityID,
		Decision:      &decision,
		Reason:        &reason,
		Context:       contextData,
	})
}

// auditFlagListed records pagination parameters and result counts for every
// list operation, supporting access-pattern analysis.
func (s *service) auditFlagListed(ctx context.Context, request *ListFeatureFlagsRequest, response *ListFeatureFlagsResponse) {
	contextData, _ := json.Marshal(map[string]any{
		"page":      request.Page,
		"page_size": request.PageSize,
		"flag_type": request.FlagType,
		"total":     response.Total,
		"returned":  len(response.FeatureFlags),
	})

	decision := "LISTED"
	reason := fmt.Sprintf("Listed %d feature flags (page %d)", len(response.FeatureFlags), request.Page)
	s.auditService.CreateAuditEvent(ctx, audit.CreateAuditEventRequest{
		EventType:     AuditEventFeatureFlagListed,
		EventCategory: AuditCategoryAccess,
		Severity:      AuditSeverityInfo,
		Decision:      &decision,
		Reason:        &reason,
		Context:       contextData,
	})
}

// auditFlagSearched records search queries alongside result counts so
// operators can track which flags are most frequently searched and detect
// potential scraping attempts.
func (s *service) auditFlagSearched(ctx context.Context, query string, limit, offset int32, results []*FeatureFlag) {
	contextData, _ := json.Marshal(map[string]any{
		"query":         query,
		"limit":         limit,
		"offset":        offset,
		"results_found": len(results),
		"searched_at":   time.Now(),
	})

	decision := "SEARCHED"
	reason := fmt.Sprintf("Searched feature flags with query '%s', found %d results", query, len(results))
	s.auditService.CreateAuditEvent(ctx, audit.CreateAuditEventRequest{
		EventType:     AuditEventFeatureFlagSearched,
		EventCategory: AuditCategoryAccess,
		Severity:      AuditSeverityInfo,
		Decision:      &decision,
		Reason:        &reason,
		Context:       contextData,
	})
}

// getUpdatedFields inspects an UpdateFeatureFlagRequest and returns the names
// of fields that were explicitly set (i.e. non-nil). Used to populate the
// "updated_fields" key in audit events for easier change-tracking.
func (s *service) getUpdatedFields(request *UpdateFeatureFlagRequest) []string {
	var fields []string
	if request.Description != nil {
		fields = append(fields, "description")
	}
	if request.DefaultValue != nil {
		fields = append(fields, "default_value")
	}
	if request.RolloutPercentage != nil {
		fields = append(fields, "rollout_percentage")
	}
	if request.TargetAudience != nil {
		fields = append(fields, "target_audience")
	}
	if request.Metadata != nil {
		fields = append(fields, "metadata")
	}
	return fields
}
