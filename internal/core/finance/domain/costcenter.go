package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CostCenterStatus represents the operational status of a cost centre.
type CostCenterStatus string

const (
	CostCenterStatusActive   CostCenterStatus = "ACTIVE"
	CostCenterStatusInactive CostCenterStatus = "INACTIVE"
	CostCenterStatusClosed   CostCenterStatus = "CLOSED"
)

// IsValid returns true if the CostCenterStatus is a recognised value.
func (s CostCenterStatus) IsValid() bool {
	switch s {
	case CostCenterStatusActive, CostCenterStatusInactive, CostCenterStatusClosed:
		return true
	default:
		return false
	}
}

// String returns the string representation of CostCenterStatus.
func (s CostCenterStatus) String() string { return string(s) }

// CostCenter represents an organisational unit to which costs and revenues are attributed.
// A cost centre has a unique code within a tenant and may form a hierarchy via ParentID.
type CostCenter struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// Identity
	Code        string  `json:"code"`                  // Short mnemonic code, e.g. "CC-001"
	Name        string  `json:"name"`                  // Display name
	Description *string `json:"description,omitempty"` // Optional long description

	// Hierarchy
	ParentID *uuid.UUID `json:"parent_id,omitempty"` // Parent cost centre for roll-up reporting
	Level    int        `json:"level"`               // 0 = root, 1 = child, etc.
	Path     *string    `json:"path,omitempty"`      // Materialised path e.g. "/root-id/child-id/"

	// Classification
	DepartmentID *uuid.UUID `json:"department_id,omitempty"` // Optional link to HR department
	ManagerID    *uuid.UUID `json:"manager_id,omitempty"`    // Responsible person (tenant user ID)

	Status   CostCenterStatus `json:"status"`
	IsActive bool             `json:"is_active"`

	// Audit
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// CanAcceptTransactions returns true when journal entries may reference this cost centre.
func (cc *CostCenter) CanAcceptTransactions() bool {
	return cc.Status == CostCenterStatusActive && cc.IsActive && cc.DeletedAt == nil
}

// IsDescendantOf reports whether this cost centre is a descendant of parentID using
// the materialised path.
func (cc *CostCenter) IsDescendantOf(parentID uuid.UUID) bool {
	if cc.Path == nil {
		return false
	}
	return strings.Contains(*cc.Path, fmt.Sprintf("/%s/", parentID.String()))
}

// Validate returns a slice of ValidationErrors for the CostCenter.
func (cc *CostCenter) Validate() []ValidationError {
	var errs []ValidationError

	if cc.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "tenant_id", Message: "tenant_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if strings.TrimSpace(cc.Code) == "" {
		errs = append(errs, ValidationError{Field: "code", Message: "cost centre code is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	} else if len(cc.Code) > 20 {
		errs = append(errs, ValidationError{Field: "code", Message: "cost centre code must be 20 characters or less", Code: "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError})
	}
	if strings.TrimSpace(cc.Name) == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "cost centre name is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	} else if len(cc.Name) > 200 {
		errs = append(errs, ValidationError{Field: "name", Message: "cost centre name must be 200 characters or less", Code: "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError})
	}
	if !cc.Status.IsValid() {
		errs = append(errs, ValidationError{Field: "status", Message: fmt.Sprintf("invalid cost centre status: %s", cc.Status), Code: "INVALID_VALUE", Severity: ValidationSeverityError})
	}
	if cc.ParentID == nil && cc.Level != 0 {
		errs = append(errs, ValidationError{Field: "level", Message: "root cost centres must have level 0", Code: "INCONSISTENT_VALUE", Severity: ValidationSeverityError})
	}
	if cc.ParentID != nil && cc.Level <= 0 {
		errs = append(errs, ValidationError{Field: "level", Message: "child cost centres must have level > 0", Code: "INCONSISTENT_VALUE", Severity: ValidationSeverityError})
	}

	return errs
}

// CreateCostCenterRequest is the payload for creating a new cost centre.
type CreateCostCenterRequest struct {
	Code        string     `json:"code"         validate:"required,max=20"`
	Name        string     `json:"name"         validate:"required,max=200"`
	Description *string    `json:"description,omitempty" validate:"omitempty,max=1000"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	ManagerID   *uuid.UUID `json:"manager_id,omitempty"`
}

// UpdateCostCenterRequest is the payload for updating an existing cost centre.
type UpdateCostCenterRequest struct {
	Name        *string    `json:"name,omitempty"         validate:"omitempty,max=200"`
	Description *string    `json:"description,omitempty"  validate:"omitempty,max=1000"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	ManagerID   *uuid.UUID `json:"manager_id,omitempty"`
	Status      *CostCenterStatus `json:"status,omitempty"`
	IsActive    *bool      `json:"is_active,omitempty"`
}
