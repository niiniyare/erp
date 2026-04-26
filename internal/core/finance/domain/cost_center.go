// Package domain contains the core business entities and rules for the Awo ERP
// system. Types in this package are persistence-agnostic; they are consumed by
// service and repository layers but carry no infrastructure dependencies.
package domain

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Field-length & precision constants
// =============================================================================

// These constants are the single source of truth for field constraints.
// They are referenced in both Validate() methods and DTO struct tags.
const (
	CostCenterCodeMaxLen        = 20
	CostCenterNameMaxLen        = 200
	CostCenterDescriptionMaxLen = 1000

	// allocationPercentageTolerance is the maximum allowed deviation from 100
	// when validating that a set of PERCENTAGE allocations sums correctly.
	// Floating-point arithmetic means an exact equality check is unreliable.
	allocationPercentageTolerance = 0.0001
)

// =============================================================================
// CostCenterStatus
// =============================================================================

// CostCenterStatus represents the operational lifecycle state of a cost centre.
// The state machine is:
//
//	ACTIVE  ──► INACTIVE  ──► ACTIVE   (re-activation is permitted)
//	ACTIVE  ──► CLOSED               (terminal — cannot be re-opened)
//	INACTIVE ──► CLOSED              (terminal — cannot be re-opened)
//
// Only ACTIVE centres may accept journal postings.
type CostCenterStatus string

const (
	// CostCenterStatusActive means the centre is fully operational.
	// It may accept direct journal postings and participate in cost distributions.
	CostCenterStatusActive CostCenterStatus = "ACTIVE"

	// CostCenterStatusInactive means the centre is temporarily suspended.
	// Existing data is preserved; no new transactions are permitted until
	// the centre is re-activated.
	CostCenterStatusInactive CostCenterStatus = "INACTIVE"

	// CostCenterStatusClosed is a terminal state. The centre has been
	// permanently shut down and is retained for historical reporting only.
	// Transition to CLOSED cannot be reversed.
	CostCenterStatusClosed CostCenterStatus = "CLOSED"
)

// IsValid returns true if s is one of the recognised status constants.
func (s CostCenterStatus) IsValid() bool {
	switch s {
	case CostCenterStatusActive, CostCenterStatusInactive, CostCenterStatusClosed:
		return true
	default:
		return false
	}
}

// CanTransitionTo reports whether a transition from the current status to next
// is permitted by the state machine.
//
// Permitted transitions:
//   - ACTIVE   → INACTIVE
//   - ACTIVE   → CLOSED
//   - INACTIVE → ACTIVE
//   - INACTIVE → CLOSED
//
// CLOSED is terminal: no outbound transitions are allowed.
func (s CostCenterStatus) CanTransitionTo(next CostCenterStatus) bool {
	switch s {
	case CostCenterStatusActive:
		return next == CostCenterStatusInactive || next == CostCenterStatusClosed
	case CostCenterStatusInactive:
		return next == CostCenterStatusActive || next == CostCenterStatusClosed
	case CostCenterStatusClosed:
		return false // terminal state
	default:
		return false
	}
}

// String implements fmt.Stringer.
func (s CostCenterStatus) String() string { return string(s) }

// =============================================================================
// AllocationMethod
// =============================================================================

// AllocationMethod defines the algorithm used to split costs from a source
// cost centre across one or more target centres during period-end allocation.
type AllocationMethod string

const (
	// AllocationMethodPercentage distributes costs by a fixed percentage split.
	// Each CostCenterAllocation for the source must carry a Percentage value,
	// and the sum across all targets must equal exactly 100 (within tolerance).
	AllocationMethodPercentage AllocationMethod = "PERCENTAGE"

	// AllocationMethodHeadcount distributes costs proportionally to employee
	// headcount in each target centre. Each allocation must carry a DriverValue
	// representing the headcount for that target. The engine computes each
	// target's share as: DriverValue / sum(all DriverValues).
	AllocationMethodHeadcount AllocationMethod = "HEADCOUNT"

	// AllocationMethodSquareFootage distributes costs proportionally to the
	// occupied floor area (m²) of each target centre. DriverValue holds the
	// area for that target.
	AllocationMethodSquareFootage AllocationMethod = "SQUARE_FOOTAGE"

	// AllocationMethodActivityBased distributes costs proportionally to a
	// user-defined activity driver (e.g. machine-hours, number of service
	// requests). DriverValue holds the raw activity quantity for the target.
	AllocationMethodActivityBased AllocationMethod = "ACTIVITY_BASED"
)

// IsValid returns true if m is one of the recognised method constants.
func (m AllocationMethod) IsValid() bool {
	switch m {
	case AllocationMethodPercentage, AllocationMethodHeadcount,
		AllocationMethodSquareFootage, AllocationMethodActivityBased:
		return true
	default:
		return false
	}
}

// RequiresDriverValue reports whether this method uses DriverValue (as opposed
// to Percentage) on each CostCenterAllocation target leg.
// All non-PERCENTAGE methods are driver-based.
func (m AllocationMethod) RequiresDriverValue() bool {
	return m != AllocationMethodPercentage
}

// String implements fmt.Stringer.
func (m AllocationMethod) String() string { return string(m) }

// =============================================================================
// CostCenter
// =============================================================================

// CostCenter is an organisational unit to which costs and revenues are
// attributed for management accounting purposes.
//
// # Hierarchy
//
// Cost centres may form a tree. ParentID links a centre to its parent.
// Level tracks depth (0 = root). Path is a materialised path string
// (e.g. "/root-id/parent-id/this-id/") maintained by the repository layer to
// enable efficient subtree queries without recursive CTEs.
//
// IsGroup marks a centre that exists only to aggregate its children in reports.
// Group centres must not receive direct journal postings.
//
// # Lifecycle
//
// Status and IsActive together control whether a centre accepts transactions.
// Both must be affirmative (Status == ACTIVE and IsActive == true) before a
// journal entry may reference this centre. DeletedAt provides a soft-delete
// escape hatch; deleted centres are invisible to active queries.
//
// # Cost Distribution
//
// When IsDistributed is true the period-end allocation engine redistributes
// all costs accumulated in this centre to one or more target centres according
// to AllocationMethod. The individual target legs are stored as
// CostCenterAllocation records keyed by this centre's ID.
type CostCenter struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// -------------------------------------------------------------------------
	// Identity
	// -------------------------------------------------------------------------

	// Code is a short mnemonic, unique within the tenant (case-insensitive
	// comparison is enforced at the repository layer), e.g. "CC-OPS-001".
	// Maximum CostCenterCodeMaxLen characters. Immutable after creation.
	Code string `json:"code"`

	// Name is the human-readable display label shown in reports and selectors.
	// Maximum CostCenterNameMaxLen characters.
	Name string `json:"name"`

	// Description is an optional long-form explanation of the centre's purpose,
	// ownership, or scope. Maximum CostCenterDescriptionMaxLen characters.
	Description *string `json:"description,omitempty"`

	// -------------------------------------------------------------------------
	// Hierarchy
	// -------------------------------------------------------------------------

	// ParentID links this centre to its parent in the reporting tree.
	// Nil for root centres (Level == 0).
	ParentID *uuid.UUID `json:"parent_id,omitempty"`

	// Level is the zero-based depth in the tree (0 = root, 1 = child, …).
	// The invariant is: ParentID == nil ↔ Level == 0.
	Level int `json:"level"`

	// Path is the materialised ancestor path, e.g. "/a-uuid/b-uuid/this-uuid/".
	// Slashes wrap every segment so substring search suffices for subtree
	// membership tests. Maintained exclusively by the repository; never set
	// directly by callers.
	Path *string `json:"path,omitempty"`

	// IsGroup marks a pure grouping node used only for roll-up reporting.
	// Group centres aggregate their children's balances but must not receive
	// direct postings. CanAcceptTransactions returns false for group centres.
	IsGroup bool `json:"is_group"`

	// -------------------------------------------------------------------------
	// Classification & ownership
	// -------------------------------------------------------------------------

	// DepartmentID optionally links this centre to an HR department record,
	// enabling headcount-based allocation drivers and org-chart reporting.
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`

	// ManagerID is the UUID of the tenant user accountable for this centre's
	// budget. Used for approval workflows and budget-vs-actual reports.
	ManagerID *uuid.UUID `json:"manager_id,omitempty"`

	// -------------------------------------------------------------------------
	// Cost distribution
	// -------------------------------------------------------------------------

	// IsDistributed marks this centre for automatic period-end cost
	// redistribution. When true:
	//   1. AllocationMethod must be set.
	//   2. At least one CostCenterAllocation record must exist with this
	//      centre as SourceCenterID.
	//   3. For PERCENTAGE method, allocation percentages must sum to 100.
	// The allocation engine drains the centre's balance to zero each period.
	IsDistributed bool `json:"is_distributed"`

	// AllocationMethod determines how the allocation engine splits costs across
	// target centres. Required when IsDistributed is true; ignored otherwise.
	AllocationMethod *AllocationMethod `json:"allocation_method,omitempty"`

	// -------------------------------------------------------------------------
	// Lifecycle
	// -------------------------------------------------------------------------

	// Status is the explicit operational state. See CostCenterStatus for the
	// permitted state machine transitions.
	Status CostCenterStatus `json:"status"`

	// IsActive is a fast-path administrative toggle. Setting IsActive = false
	// suspends the centre without triggering a formal status transition.
	// A centre may only accept transactions when both Status == ACTIVE and
	// IsActive == true.
	IsActive bool `json:"is_active"`

	// -------------------------------------------------------------------------
	// Audit trail
	// -------------------------------------------------------------------------

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`

	// DeletedAt is populated on soft-delete. Deleted centres are excluded from
	// all active queries and cannot accept transactions. Hard deletion is not
	// supported; use status CLOSED for permanent retirement.
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// -----------------------------------------------------------------------------
// Behaviour methods
// -----------------------------------------------------------------------------

// IsRoot returns true when this centre sits at the top of the hierarchy
// (no parent, level 0).
func (cc *CostCenter) IsRoot() bool {
	return cc.ParentID == nil && cc.Level == 0
}

// IsDeleted returns true when the centre has been soft-deleted.
func (cc *CostCenter) IsDeleted() bool {
	return cc.DeletedAt != nil
}

// CanAcceptTransactions returns true when journal entries may reference this
// cost centre directly. All of the following must hold:
//   - Status == ACTIVE
//   - IsActive == true
//   - IsGroup == false  (group centres aggregate children; no direct postings)
//   - DeletedAt == nil  (centre has not been soft-deleted)
func (cc *CostCenter) CanAcceptTransactions() bool {
	return cc.Status == CostCenterStatusActive &&
		cc.IsActive &&
		!cc.IsGroup &&
		cc.DeletedAt == nil
}

// CanBeDeleted returns true when it is safe to soft-delete the centre.
// A centre may not be deleted if it is still active or has pending distributions
// configured (those allocation records should be removed first).
func (cc *CostCenter) CanBeDeleted() bool {
	return cc.Status != CostCenterStatusActive && !cc.IsDistributed
}

// IsDescendantOf reports whether this cost centre is a descendant of the centre
// identified by parentID, using the materialised Path for an O(1) check.
// Returns false when Path has not been set (e.g. for an unsaved entity).
func (cc *CostCenter) IsDescendantOf(parentID uuid.UUID) bool {
	if cc.Path == nil {
		return false
	}
	return strings.Contains(*cc.Path, fmt.Sprintf("/%s/", parentID.String()))
}

// ApplyUpdate applies the non-nil fields from req onto the cost centre and
// stamps UpdatedBy/UpdatedAt. The caller is responsible for persisting the
// result. Status transitions are validated against the state machine; an error
// is returned if the requested transition is invalid.
//
// Note: Code and ParentID are intentionally absent from UpdateCostCenterRequest
// because changing them requires dedicated service-layer operations (code
// uniqueness re-check and hierarchy re-path respectively).
func (cc *CostCenter) ApplyUpdate(req UpdateCostCenterRequest, updatedBy uuid.UUID) error {
	if req.Name != nil {
		cc.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		cc.Description = req.Description
	}
	if req.DepartmentID != nil {
		cc.DepartmentID = req.DepartmentID
	}
	if req.ManagerID != nil {
		cc.ManagerID = req.ManagerID
	}
	if req.IsActive != nil {
		cc.IsActive = *req.IsActive
	}
	if req.IsGroup != nil {
		cc.IsGroup = *req.IsGroup
	}
	if req.IsDistributed != nil {
		cc.IsDistributed = *req.IsDistributed
	}
	if req.AllocationMethod != nil {
		cc.AllocationMethod = req.AllocationMethod
	}
	if req.Status != nil {
		if !cc.Status.CanTransitionTo(*req.Status) {
			return fmt.Errorf(
				"invalid status transition: %s → %s", cc.Status, *req.Status,
			)
		}
		cc.Status = *req.Status
	}

	now := time.Now().UTC()
	cc.UpdatedAt = now
	cc.UpdatedBy = &updatedBy
	return nil
}

// Validate performs domain-level consistency checks and returns all discovered
// errors. An empty (nil) slice means the entity is valid.
// Validate does not check cross-entity constraints (e.g. parent existence,
// code uniqueness); those are enforced by the service layer.
func (cc *CostCenter) Validate() []ValidationError {
	var errs []ValidationError

	// -- Tenant ---------------------------------------------------------------
	if cc.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "tenant_id", Message: "tenant_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}

	// -- Code -----------------------------------------------------------------
	switch {
	case strings.TrimSpace(cc.Code) == "":
		errs = append(errs, ValidationError{
			Field: "code", Message: "cost centre code is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	case len(cc.Code) > CostCenterCodeMaxLen:
		errs = append(errs, ValidationError{
			Field: "code",
			Message: fmt.Sprintf(
				"cost centre code must be %d characters or less", CostCenterCodeMaxLen,
			),
			Code: "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}

	// -- Name -----------------------------------------------------------------
	switch {
	case strings.TrimSpace(cc.Name) == "":
		errs = append(errs, ValidationError{
			Field: "name", Message: "cost centre name is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	case len(cc.Name) > CostCenterNameMaxLen:
		errs = append(errs, ValidationError{
			Field: "name",
			Message: fmt.Sprintf(
				"cost centre name must be %d characters or less", CostCenterNameMaxLen,
			),
			Code: "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}

	// -- Description ----------------------------------------------------------
	if cc.Description != nil && len(*cc.Description) > CostCenterDescriptionMaxLen {
		errs = append(errs, ValidationError{
			Field: "description",
			Message: fmt.Sprintf(
				"description must be %d characters or less", CostCenterDescriptionMaxLen,
			),
			Code: "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}

	// -- Status ---------------------------------------------------------------
	if !cc.Status.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "status",
			Message:  fmt.Sprintf("invalid cost centre status: %q", cc.Status),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}

	// -- Hierarchy consistency ------------------------------------------------
	// Invariant: a centre is a root if and only if it has no parent AND level 0.
	if cc.ParentID == nil && cc.Level != 0 {
		errs = append(errs, ValidationError{
			Field:    "level",
			Message:  "root cost centre (no parent) must have level 0",
			Code:     "INCONSISTENT_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if cc.ParentID != nil && cc.Level <= 0 {
		errs = append(errs, ValidationError{
			Field:    "level",
			Message:  "child cost centre (has parent) must have level > 0",
			Code:     "INCONSISTENT_VALUE",
			Severity: ValidationSeverityError,
		})
	}

	// -- Distribution consistency --------------------------------------------
	if cc.IsDistributed && cc.AllocationMethod == nil {
		errs = append(errs, ValidationError{
			Field:    "allocation_method",
			Message:  "allocation_method is required when is_distributed is true",
			Code:     "REQUIRED_FIELD",
			Severity: ValidationSeverityError,
		})
	}
	if cc.AllocationMethod != nil && !cc.AllocationMethod.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "allocation_method",
			Message:  fmt.Sprintf("invalid allocation method: %q", *cc.AllocationMethod),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	// Warn when an allocation method is set but distribution is not enabled,
	// as the orphaned value will be silently ignored by the engine.
	if !cc.IsDistributed && cc.AllocationMethod != nil {
		errs = append(errs, ValidationError{
			Field:    "allocation_method",
			Message:  "allocation_method has no effect when is_distributed is false",
			Code:     "SUPERFLUOUS_VALUE",
			Severity: ValidationSeverityWarning,
		})
	}

	return errs
}

// =============================================================================
// CostCenterAllocation
// =============================================================================

// CostCenterAllocation defines a single target leg in a cost distribution rule.
//
// Each distributed cost centre (CostCenter.IsDistributed == true) must have one
// or more CostCenterAllocation records with SourceCenterID pointing to it. At
// period-end the allocation engine:
//
//  1. Reads all allocations for the source centre.
//  2. Computes each target's share using Method (percentage or driver-ratio).
//  3. Reverses the source centre's balance to zero.
//  4. Posts a corresponding credit to each target centre.
//
// For PERCENTAGE allocations, the sum of all Percentage values across the
// sibling legs (same SourceCenterID) must equal 100. Validate this at the set
// level using ValidateAllocationSet.
type CostCenterAllocation struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// SourceCenterID is the cost centre whose accumulated costs are being
	// redistributed. Must reference a CostCenter with IsDistributed == true.
	SourceCenterID uuid.UUID `json:"source_center_id"`

	// TargetCenterID is the cost centre that receives the distributed share.
	// Must be different from SourceCenterID (self-allocation is not permitted).
	TargetCenterID uuid.UUID `json:"target_center_id"`

	// Method is the distribution algorithm for this leg. It must match
	// CostCenter.AllocationMethod of the source centre. Stored here to make
	// each leg self-describing and to allow future per-leg method overrides.
	Method AllocationMethod `json:"method"`

	// Percentage is the fixed share (0 < x ≤ 100) assigned to this target.
	// Required when Method == PERCENTAGE; ignored for driver-based methods.
	// The sum of Percentage across all legs for the same source must equal 100.
	Percentage *float64 `json:"percentage,omitempty"`

	// DriverValue is the raw driver quantity attributed to this target centre
	// (e.g. headcount = 12, floor area = 450.5). Required for all non-PERCENTAGE
	// methods. The engine normalises each value against the total across all
	// sibling legs to derive a share percentage.
	// Must be > 0.
	DriverValue *float64 `json:"driver_value,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// Validate checks the internal consistency of a single allocation leg.
// Cross-leg rules (e.g. percentages summing to 100) require the full sibling
// set and are checked by ValidateAllocationSet.
func (a *CostCenterAllocation) Validate() []ValidationError {
	var errs []ValidationError

	if a.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "tenant_id", Message: "tenant_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}
	if a.SourceCenterID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "source_center_id", Message: "source_center_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}
	if a.TargetCenterID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "target_center_id", Message: "target_center_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}
	if a.SourceCenterID != uuid.Nil && a.TargetCenterID != uuid.Nil &&
		a.SourceCenterID == a.TargetCenterID {
		errs = append(errs, ValidationError{
			Field:    "target_center_id",
			Message:  "target_center_id must differ from source_center_id",
			Code:     "SELF_REFERENCE",
			Severity: ValidationSeverityError,
		})
	}
	if !a.Method.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "method",
			Message:  fmt.Sprintf("invalid allocation method: %q", a.Method),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}

	// Validate value fields based on method.
	if a.Method == AllocationMethodPercentage {
		switch {
		case a.Percentage == nil:
			errs = append(errs, ValidationError{
				Field:    "percentage",
				Message:  "percentage is required for PERCENTAGE allocation method",
				Code:     "REQUIRED_FIELD",
				Severity: ValidationSeverityError,
			})
		case *a.Percentage <= 0 || *a.Percentage > 100:
			errs = append(errs, ValidationError{
				Field:    "percentage",
				Message:  "percentage must be in the range (0, 100]",
				Code:     "OUT_OF_RANGE",
				Severity: ValidationSeverityError,
			})
		}
	} else if a.Method.RequiresDriverValue() {
		switch {
		case a.DriverValue == nil:
			errs = append(errs, ValidationError{
				Field: "driver_value",
				Message: fmt.Sprintf(
					"driver_value is required for %s allocation method", a.Method,
				),
				Code:     "REQUIRED_FIELD",
				Severity: ValidationSeverityError,
			})
		case *a.DriverValue <= 0:
			errs = append(errs, ValidationError{
				Field:    "driver_value",
				Message:  "driver_value must be greater than zero",
				Code:     "OUT_OF_RANGE",
				Severity: ValidationSeverityError,
			})
		}
	}

	return errs
}

// ValidateAllocationSet performs cross-leg validation for a complete set of
// allocation records that all share the same SourceCenterID.
//
// Rules checked:
//   - The set must be non-empty.
//   - All legs must share the same Method.
//   - For PERCENTAGE allocations the percentages must sum to 100 (within
//     allocationPercentageTolerance).
//
// Individual-leg rules are not re-checked here; call Validate() on each leg
// separately before calling this function.
func ValidateAllocationSet(allocations []CostCenterAllocation) []ValidationError {
	var errs []ValidationError

	if len(allocations) == 0 {
		return append(errs, ValidationError{
			Field:    "allocations",
			Message:  "at least one allocation leg is required for a distributed cost centre",
			Code:     "REQUIRED_FIELD",
			Severity: ValidationSeverityError,
		})
	}

	// All legs must use the same method (method is set on the source centre).
	firstMethod := allocations[0].Method
	for i, a := range allocations[1:] {
		if a.Method != firstMethod {
			errs = append(errs, ValidationError{
				Field: "method",
				Message: fmt.Sprintf(
					"allocation leg %d has method %s; expected %s to match all other legs",
					i+1, a.Method, firstMethod,
				),
				Code:     "INCONSISTENT_VALUE",
				Severity: ValidationSeverityError,
			})
		}
	}

	// For percentage-based allocations, verify the sum.
	if firstMethod == AllocationMethodPercentage {
		var total float64
		for _, a := range allocations {
			if a.Percentage != nil {
				total += *a.Percentage
			}
		}
		if math.Abs(total-100.0) > allocationPercentageTolerance {
			errs = append(errs, ValidationError{
				Field: "percentage",
				Message: fmt.Sprintf(
					"allocation percentages must sum to 100; current sum is %.4f", total,
				),
				Code:     "INVALID_SUM",
				Severity: ValidationSeverityError,
			})
		}
	}

	return errs
}

// =============================================================================
// Request DTOs
// =============================================================================

// CreateCostCenterRequest is the inbound payload for creating a new cost centre
// via the service layer.
//
// Level and Path are derived by the service from ParentID and must not be set
// by the caller. The new centre is always created with Status == ACTIVE.
type CreateCostCenterRequest struct {
	// Code must be unique within the tenant. Immutable after creation.
	Code string `json:"code" validate:"required,max=20"`
	Name string `json:"name" validate:"required,max=200"`

	Description  *string    `json:"description,omitempty"  validate:"omitempty,max=1000"`
	ParentID     *uuid.UUID `json:"parent_id,omitempty"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	ManagerID    *uuid.UUID `json:"manager_id,omitempty"`

	// IsGroup marks this centre as a reporting-only grouping node.
	// Group centres cannot accept direct journal postings.
	IsGroup bool `json:"is_group"`

	// IsDistributed marks this centre for period-end cost redistribution.
	// If true, AllocationMethod must be provided.
	IsDistributed    bool              `json:"is_distributed"`
	AllocationMethod *AllocationMethod `json:"allocation_method,omitempty"`
}

// Validate checks that the request itself is self-consistent before the service
// layer applies cross-entity rules (parent existence, code uniqueness, etc.).
func (r *CreateCostCenterRequest) Validate() []ValidationError {
	var errs []ValidationError

	if strings.TrimSpace(r.Code) == "" {
		errs = append(errs, ValidationError{
			Field: "code", Message: "code is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	} else if len(r.Code) > CostCenterCodeMaxLen {
		errs = append(errs, ValidationError{
			Field:    "code",
			Message:  fmt.Sprintf("code must be %d characters or less", CostCenterCodeMaxLen),
			Code:     "MAX_LENGTH_EXCEEDED",
			Severity: ValidationSeverityError,
		})
	}
	if strings.TrimSpace(r.Name) == "" {
		errs = append(errs, ValidationError{
			Field: "name", Message: "name is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	} else if len(r.Name) > CostCenterNameMaxLen {
		errs = append(errs, ValidationError{
			Field:    "name",
			Message:  fmt.Sprintf("name must be %d characters or less", CostCenterNameMaxLen),
			Code:     "MAX_LENGTH_EXCEEDED",
			Severity: ValidationSeverityError,
		})
	}
	if r.Description != nil && len(*r.Description) > CostCenterDescriptionMaxLen {
		errs = append(errs, ValidationError{
			Field: "description",
			Message: fmt.Sprintf(
				"description must be %d characters or less", CostCenterDescriptionMaxLen,
			),
			Code:     "MAX_LENGTH_EXCEEDED",
			Severity: ValidationSeverityError,
		})
	}
	if r.IsDistributed && r.AllocationMethod == nil {
		errs = append(errs, ValidationError{
			Field:    "allocation_method",
			Message:  "allocation_method is required when is_distributed is true",
			Code:     "REQUIRED_FIELD",
			Severity: ValidationSeverityError,
		})
	}
	if r.AllocationMethod != nil && !r.AllocationMethod.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "allocation_method",
			Message:  fmt.Sprintf("invalid allocation method: %q", *r.AllocationMethod),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}

	return errs
}

// UpdateCostCenterRequest is the inbound payload for a partial update.
// Only non-nil fields are applied; nil fields are left unchanged.
//
// Code and ParentID are excluded intentionally:
//   - Code changes require a uniqueness re-check and are handled by a dedicated
//     service operation.
//   - ParentID changes require hierarchy re-pathing (Level, Path, all
//     descendants) and are handled by a dedicated Reparent operation.
type UpdateCostCenterRequest struct {
	Name         *string    `json:"name,omitempty"         validate:"omitempty,max=200"`
	Description  *string    `json:"description,omitempty"  validate:"omitempty,max=1000"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	ManagerID    *uuid.UUID `json:"manager_id,omitempty"`

	// Status triggers a state machine transition. The service layer rejects
	// transitions that are not permitted by CostCenterStatus.CanTransitionTo.
	Status *CostCenterStatus `json:"status,omitempty"`

	IsActive      *bool `json:"is_active,omitempty"`
	IsGroup       *bool `json:"is_group,omitempty"`
	IsDistributed *bool `json:"is_distributed,omitempty"`

	// Setting AllocationMethod to non-nil implicitly requires IsDistributed to
	// be (or remain) true. The service layer enforces this.
	AllocationMethod *AllocationMethod `json:"allocation_method,omitempty"`
}
