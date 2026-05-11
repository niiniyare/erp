//go:build ignore

package conditional

import (
	"time"

	"github.com/google/uuid"
)

// ConditionalAccessPolicy represents a policy for conditional access
type ConditionalAccessPolicy struct {
	ID          uuid.UUID      `json:"id"`
	TenantID    uuid.UUID      `json:"tenant_id"`
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Priority    int            `json:"priority"`
	IsActive    bool           `json:"is_active"`
	Conditions  map[string]any `json:"conditions"` // e.g., IP range, time of day, device compliance
	Actions     map[string]any `json:"actions"`    // e.g., BLOCK, MFA_REQUIRED, LOG_ONLY, STEP_UP_AUTH
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// PolicyEvaluationResult represents the result of a conditional access policy evaluation
type PolicyEvaluationResult struct {
	PolicyID          uuid.UUID      `json:"policy_id"`
	PolicyName        string         `json:"policy_name"`
	Decision          string         `json:"decision"` // ALLOW, BLOCK, MFA_REQUIRED, LOG_ONLY
	MatchedConditions []string       `json:"matched_conditions,omitempty"`
	ActionsTaken      []string       `json:"actions_taken,omitempty"`
	Details           map[string]any `json:"details,omitempty"`
	ExecutedAt        time.Time      `json:"executed_at"`
}
