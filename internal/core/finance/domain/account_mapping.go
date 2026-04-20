package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// ReportingScheme
// =============================================================================

// ReportingScheme is the top-level container for a complete, self-consistent
// reporting hierarchy. Each scheme owns its own tree of ReportingGroups and its
// own set of AccountMappings.
//
// Multiple schemes can coexist for the same tenant, allowing the same ledger
// accounts to be presented differently in different regulatory or management
// contexts without any changes to the chart of accounts.
//
// # Schemes in use at Shell Maanzoni
//
//   - INTERNAL: Management accounts used for day-to-day operational reporting.
//   - EPRA: Energy and Petroleum Regulatory Authority format for fuel-grade revenue.
//   - KRA: Kenya Revenue Authority format for VAT and corporate tax submissions.
//   - IFRS: International Financial Reporting Standards presentation.
type ReportingScheme struct {
	ID       uuid.UUID  `json:"id"`
	TenantID uuid.UUID  `json:"tenant_id"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"`

	// Code is a short unique mnemonic, e.g. "EPRA", "KRA", "INTERNAL".
	Code string `json:"code"`

	// Name is the human-readable label, e.g. "EPRA Regulatory Reporting".
	Name string `json:"name"`

	// Description explains the scheme's purpose and applicable standards.
	Description *string `json:"description,omitempty"`

	// DefaultConsolidationMethod is applied to all ReportingGroups in this
	// scheme that do not override it explicitly. Defaults to SUM.
	DefaultConsolidationMethod ConsolidationMethod `json:"default_consolidation_method"`

	// IsSystemDefined marks schemes provisioned by the seed (EPRA, KRA).
	// Tenant users may not delete or rename system-defined schemes.
	IsSystemDefined bool `json:"is_system_defined"`

	// IsActive controls whether this scheme is available for report generation.
	IsActive bool `json:"is_active"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// =============================================================================
// AccountMapping
// =============================================================================

// AccountMapping is the explicit join between a ledger Account and a
// ReportingGroup within a specific ReportingScheme.
//
// # Why this entity exists
//
// Without AccountMapping the only way to put an account on a report would be to
// add a foreign key from Account to AccountGroup (or ReportingGroup). That
// creates two problems:
//
//  1. An account can only belong to one group per scheme in that model.
//     AccountMapping allows the same account to appear in multiple groups within
//     different schemes — e.g. "Fuel Revenue (PMS)" appears under
//     "Motor Spirit Revenue" in the EPRA scheme and under "Petroleum Sales" in
//     the internal management account scheme.
//
//  2. Changing the report layout (reordering groups, renaming sections) would
//     require touching Account records. AccountMapping lets the reporting layer
//     evolve independently of the ledger.
//
// # Override fields
//
// AccountMapping carries three presentation overrides that take precedence over
// the Account's own fields when rendering within this group:
//
//   - DisplayLabel: print a different label in the report (e.g. "Petrol (PMS)")
//     instead of the ledger account name.
//   - DisplayOrder: control sort position within the group independently of
//     Account.DisplayOrder.
//   - SignFlip: negate the account's effective balance for presentation. Used
//     when a credit-normal account (e.g. Accumulated Depreciation) appears in a
//     section where positive values are expected (e.g. under Fixed Assets as a
//     deduction).
//
// # One-to-many
//
// One Account may have multiple AccountMappings across different schemes. It
// must not have two mappings to the same ReportingGroup (that would double-count
// the balance). The repository enforces a unique constraint on
// (account_id, reporting_group_id).
type AccountMapping struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// AccountID references the ledger account whose balance is being mapped.
	AccountID uuid.UUID `json:"account_id"`

	// ReportingGroupID references the group in which this account's balance
	// should appear. The group's ReportingSchemeID determines which scheme this
	// mapping contributes to.
	ReportingGroupID uuid.UUID `json:"reporting_group_id"`

	// -------------------------------------------------------------------------
	// Presentation overrides
	// -------------------------------------------------------------------------

	// DisplayLabel overrides the account name in the report output.
	// Nil means use the account's AccountName.
	DisplayLabel *string `json:"display_label,omitempty"`

	// DisplayOrder overrides the account's default display order within this
	// group. Nil means use Account.DisplayOrder.
	DisplayOrder *int `json:"display_order,omitempty"`

	// SignFlip negates the account's effective balance for presentation in this
	// group. Use when the balance direction convention of the group differs from
	// that of the account's NormalBalance.
	//
	// Example: "Accumulated Depreciation" is a credit-normal contra-asset.
	// In a balance sheet section showing Asset values, its contribution should
	// be shown as a positive deduction, so SignFlip = true flips the credit
	// balance to a positive presentation number.
	SignFlip bool `json:"sign_flip"`

	// -------------------------------------------------------------------------
	// Lifecycle
	// -------------------------------------------------------------------------

	// IsActive controls whether this mapping is included in report generation.
	// Setting false hides the account from the group without deleting history.
	IsActive bool `json:"is_active"`

	// ValidFrom and ValidTo allow time-bounded mappings. Nil ValidFrom means
	// the mapping is effective from the account's creation. Nil ValidTo means
	// no expiry. The reporting engine uses these to reconstruct historical
	// reports correctly after remapping.
	ValidFrom *time.Time `json:"valid_from,omitempty"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`

	// -------------------------------------------------------------------------
	// Audit
	// -------------------------------------------------------------------------

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// IsEffectiveAt returns true when this mapping is in effect at the given point
// in time, considering ValidFrom and ValidTo.
func (m *AccountMapping) IsEffectiveAt(t time.Time) bool {
	if !m.IsActive {
		return false
	}
	if m.ValidFrom != nil && t.Before(*m.ValidFrom) {
		return false
	}
	if m.ValidTo != nil && t.After(*m.ValidTo) {
		return false
	}
	return true
}

// Validate checks the internal consistency of a single AccountMapping.
func (m *AccountMapping) Validate() []ValidationError {
	var errs []ValidationError

	if m.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "tenant_id", Message: "tenant_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}
	if m.AccountID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "account_id", Message: "account_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}
	if m.ReportingGroupID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "reporting_group_id", Message: "reporting_group_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}
	if m.DisplayOrder != nil && *m.DisplayOrder < 0 {
		errs = append(errs, ValidationError{
			Field: "display_order", Message: "display order must be non-negative",
			Code: "OUT_OF_RANGE", Severity: ValidationSeverityError,
		})
	}
	// ValidTo must be after ValidFrom when both are set.
	if m.ValidFrom != nil && m.ValidTo != nil && !m.ValidTo.After(*m.ValidFrom) {
		errs = append(errs, ValidationError{
			Field:    "valid_to",
			Message:  "valid_to must be after valid_from",
			Code:     "INCONSISTENT_VALUE",
			Severity: ValidationSeverityError,
		})
	}

	return errs
}

// =============================================================================
// Request DTOs
// =============================================================================

// CreateAccountMappingRequest is the payload for mapping an account to a
// reporting group.
type CreateAccountMappingRequest struct {
	AccountID        uuid.UUID `json:"account_id"         validate:"required"`
	ReportingGroupID uuid.UUID `json:"reporting_group_id" validate:"required"`

	DisplayLabel *string `json:"display_label,omitempty"  validate:"omitempty,max=200"`
	DisplayOrder *int    `json:"display_order,omitempty"  validate:"omitempty,min=0"`
	SignFlip     bool    `json:"sign_flip"`
	IsActive     bool    `json:"is_active"`

	ValidFrom *time.Time `json:"valid_from,omitempty"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`
}

// Validate checks the create request for self-consistency.
func (r *CreateAccountMappingRequest) Validate() []ValidationError {
	var errs []ValidationError

	if r.AccountID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "account_id", Message: "account_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}
	if r.ReportingGroupID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "reporting_group_id", Message: "reporting_group_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}
	if r.DisplayOrder != nil && *r.DisplayOrder < 0 {
		errs = append(errs, ValidationError{
			Field: "display_order", Message: "display order must be non-negative",
			Code: "OUT_OF_RANGE", Severity: ValidationSeverityError,
		})
	}
	if r.ValidFrom != nil && r.ValidTo != nil && !r.ValidTo.After(*r.ValidFrom) {
		errs = append(errs, ValidationError{
			Field:    "valid_to",
			Message:  "valid_to must be after valid_from",
			Code:     "INCONSISTENT_VALUE",
			Severity: ValidationSeverityError,
		})
	}

	return errs
}

// UpdateAccountMappingRequest is the partial-update payload.
//
// AccountID and ReportingGroupID are excluded: changing either requires
// deleting the old mapping and creating a new one so history is preserved.
type UpdateAccountMappingRequest struct {
	DisplayLabel *string    `json:"display_label,omitempty" validate:"omitempty,max=200"`
	DisplayOrder *int       `json:"display_order,omitempty" validate:"omitempty,min=0"`
	SignFlip     *bool      `json:"sign_flip,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
	ValidFrom    *time.Time `json:"valid_from,omitempty"`
	ValidTo      *time.Time `json:"valid_to,omitempty"`
}

// =============================================================================
// Mapping set validation
// =============================================================================

// ValidateMappingSet checks a collection of AccountMappings that all belong to
// the same ReportingGroup. Rules:
//   - The set must be non-empty.
//   - No two mappings may reference the same AccountID (double-count guard).
//   - All mappings must be for the same ReportingGroupID.
func ValidateMappingSet(mappings []AccountMapping) []ValidationError {
	var errs []ValidationError

	if len(mappings) == 0 {
		return append(errs, ValidationError{
			Field:    "mappings",
			Message:  "a reporting group must have at least one account mapping to produce output",
			Code:     "REQUIRED_FIELD",
			Severity: ValidationSeverityWarning, // Warning, not error: empty groups are allowed but unusual
		})
	}

	// All mappings must share the same ReportingGroupID.
	firstGroupID := mappings[0].ReportingGroupID
	seen := make(map[uuid.UUID]struct{}, len(mappings))

	for i, m := range mappings {
		if m.ReportingGroupID != firstGroupID {
			errs = append(errs, ValidationError{
				Field: "reporting_group_id",
				Message: fmt.Sprintf(
					"mapping %d has reporting_group_id %s; expected %s",
					i, m.ReportingGroupID, firstGroupID,
				),
				Code:     "INCONSISTENT_VALUE",
				Severity: ValidationSeverityError,
			})
		}
		if _, dup := seen[m.AccountID]; dup {
			errs = append(errs, ValidationError{
				Field: "account_id",
				Message: fmt.Sprintf(
					"account %s appears more than once in reporting group %s; this would double-count its balance",
					m.AccountID, firstGroupID,
				),
				Code:     "DUPLICATE_VALUE",
				Severity: ValidationSeverityError,
			})
		}
		seen[m.AccountID] = struct{}{}
	}

	return errs
}
