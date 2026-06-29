// Package org defines the organisational hierarchy enforced by the Awo Framework.
package org

import (
	"time"

	"github.com/google/uuid"
)

// UnitType is the discriminator for the organisational level and purpose of a unit.
type UnitType string

const (
	UnitTypeCompany    UnitType = "COMPANY"
	UnitTypeSubsidiary UnitType = "SUBSIDIARY"
	UnitTypeRegion     UnitType = "REGION"
	UnitTypeBranch     UnitType = "BRANCH"
	UnitTypeLocation   UnitType = "LOCATION"
	UnitTypeDepartment UnitType = "DEPARTMENT"
	UnitTypeDivision   UnitType = "DIVISION"
	UnitTypeCostCenter UnitType = "COST_CENTER"
	UnitTypeProject    UnitType = "PROJECT"
	UnitTypeBudgetUnit UnitType = "BUDGET_UNIT"
)

func (t UnitType) IsValid() bool {
	switch t {
	case UnitTypeCompany, UnitTypeSubsidiary, UnitTypeRegion, UnitTypeBranch,
		UnitTypeLocation, UnitTypeDepartment, UnitTypeDivision,
		UnitTypeCostCenter, UnitTypeProject, UnitTypeBudgetUnit:
		return true
	}
	return false
}

// ValidationStatus is the outcome of the last validation job run on a unit.
type ValidationStatus string

const (
	ValidationStatusPending ValidationStatus = "PENDING"
	ValidationStatusValid   ValidationStatus = "VALID"
	ValidationStatusWarning ValidationStatus = "WARNING"
	ValidationStatusError   ValidationStatus = "ERROR"
)

// Unit is the in-memory representation of an org_units row.
type Unit struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
	ParentID          *uuid.UUID
	Name              string
	Code              *string
	Type              UnitType
	IsActive          bool
	Hidden            bool
	AccrualMethod     bool
	FYStartMonth      int
	Address           map[string]any
	Picture           *string
	OrgUnitPath       *string
	OrgLevel          int
	Settings          map[string]any
	Metadata          map[string]any
	Version           int
	ValidationStatus  ValidationStatus
	ValidationErrors  []any
	LastValidationRun *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

func (u *Unit) IsRoot() bool    { return u.ParentID == nil }
func (u *Unit) IsDeleted() bool { return u.DeletedAt != nil }

func (u *Unit) PathPrefix() (string, bool) {
	if u.OrgUnitPath == nil || *u.OrgUnitPath == "" {
		return "", false
	}
	return *u.OrgUnitPath + "%", true
}

// UnitPath is a row from the org_unit_paths closure table.
type UnitPath struct {
	TenantID     uuid.UUID
	AncestorID   uuid.UUID
	DescendantID uuid.UUID
	Depth        int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
