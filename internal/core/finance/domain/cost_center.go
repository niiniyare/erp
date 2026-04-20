package domain

//
// import (
// 	"fmt"
// 	"time"
//
// 	"github.com/google/uuid"
// )
//
// // AllocationMethod defines how costs are distributed across cost centres.
// type AllocationMethod string
//
// const (
// 	AllocationMethodPercentage    AllocationMethod = "PERCENTAGE"
// 	AllocationMethodHeadcount     AllocationMethod = "HEADCOUNT"
// 	AllocationMethodSquareFootage AllocationMethod = "SQUARE_FOOTAGE"
// 	AllocationMethodActivityBased AllocationMethod = "ACTIVITY_BASED"
// )
//
// // IsValid returns true if the AllocationMethod is recognised.
// func (m AllocationMethod) IsValid() bool {
// 	switch m {
// 	case AllocationMethodPercentage, AllocationMethodHeadcount,
// 		AllocationMethodSquareFootage, AllocationMethodActivityBased:
// 		return true
// 	default:
// 		return false
// 	}
// }
//
// // CostCenter represents a cost centre used for expense tracking and allocation.
// // Cost centres may be hierarchical (parent → child) and can distribute their
// // costs to other centres via AllocationTargets.
// type CostCenter struct {
// 	ID          uuid.UUID  `json:"id"`
// 	TenantID    uuid.UUID  `json:"tenant_id"`
// 	Code        string     `json:"code"`
// 	Name        string     `json:"name"`
// 	Description string     `json:"description,omitempty"`
// 	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
//
// 	// IsGroup marks this as a grouping node — it should not receive direct postings.
// 	IsGroup bool `json:"is_group"`
//
// 	// IsDistributed marks this centre for automatic cost distribution.
// 	// When true, AllocationMethod must be set and AllocationTargets should be configured.
// 	IsDistributed    bool              `json:"is_distributed"`
// 	AllocationMethod *AllocationMethod `json:"allocation_method,omitempty"`
//
// 	IsActive bool `json:"is_active"`
//
// 	CreatedAt time.Time  `json:"created_at"`
// 	UpdatedAt time.Time  `json:"updated_at"`
// 	CreatedBy uuid.UUID  `json:"created_by"`
// 	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
// }
//
// // Validate returns validation errors for the CostCenter.
// func (cc *CostCenter) Validate() []ValidationError {
// 	var errs []ValidationError
//
// 	if cc.TenantID == uuid.Nil {
// 		errs = append(errs, ValidationError{
// 			Field: "tenant_id", Message: "tenant_id is required",
// 			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
// 		})
// 	}
// 	if cc.Code == "" {
// 		errs = append(errs, ValidationError{
// 			Field: "code", Message: "cost centre code is required",
// 			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
// 		})
// 	}
// 	if cc.Name == "" {
// 		errs = append(errs, ValidationError{
// 			Field: "name", Message: "cost centre name is required",
// 			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
// 		})
// 	}
// 	if cc.IsDistributed && cc.AllocationMethod == nil {
// 		errs = append(errs, ValidationError{
// 			Field: "allocation_method", Message: "allocation_method is required for distributed cost centres",
// 			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
// 		})
// 	}
// 	if cc.AllocationMethod != nil && !cc.AllocationMethod.IsValid() {
// 		errs = append(errs, ValidationError{
// 			Field: "allocation_method", Message: fmt.Sprintf("invalid allocation method: %s", *cc.AllocationMethod),
// 			Code: "INVALID_VALUE", Severity: ValidationSeverityError,
// 		})
// 	}
//
// 	return errs
// }
//
// // CostCenterAllocation defines a single target in a cost distribution rule.
// type CostCenterAllocation struct {
// 	ID             uuid.UUID        `json:"id"`
// 	TenantID       uuid.UUID        `json:"tenant_id"`
// 	SourceCenterID uuid.UUID        `json:"source_center_id"`
// 	TargetCenterID uuid.UUID        `json:"target_center_id"`
// 	Method         AllocationMethod `json:"method"`
// 	// Percentage is used when Method is PERCENTAGE (0–100).
// 	Percentage *float64 `json:"percentage,omitempty"`
// 	// DriverValue holds the raw driver quantity (headcount, sq-footage, etc.)
// 	// used for non-percentage methods.
// 	DriverValue *float64 `json:"driver_value,omitempty"`
//
// 	CreatedAt time.Time `json:"created_at"`
// 	UpdatedAt time.Time `json:"updated_at"`
// }
