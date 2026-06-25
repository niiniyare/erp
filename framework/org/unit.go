// Package org defines the organisational hierarchy enforced by the Awo Framework.
//
// The hierarchy is a tree of OrgUnits rooted at a tenant. Every business record
// is scoped to exactly one OrgUnit via an org_unit_id foreign key. Access control
// is tree-based: a viewer at a given node can read and write records belonging to
// that node and any of its descendants.
//
// # Tree storage
//
// Two complementary structures are maintained in the database:
//
//   - org_units.org_unit_path — materialized ancestor path (/root/parent/self/).
//     Enables O(1) subtree membership checks via LIKE '/ancestor/%'.
//
//   - org_unit_paths closure table — every ancestor–descendant pair with depth.
//     Enables O(1) depth lookup and efficient ancestor/descendant list queries.
//
// Both structures are maintained by the application layer (not DB triggers) via
// InsertPaths and RebuildPaths on the Tree interface.
//
// # Unit types
//
// A tenant's root unit is always type COMPANY. Arbitrary child types (REGION,
// DEPARTMENT, COST_CENTER, etc.) may be nested up to 8 levels deep. The framework
// places no semantic meaning on the type beyond what policies choose to enforce.
package org

import (
	"time"

	"github.com/google/uuid"
)

// UnitType is the discriminator for the organisational level and purpose of a unit.
// Values must match the CHECK constraint on org_units.type.
type UnitType string

const (
	// UnitTypeCompany is the root legal entity within a tenant.
	// Every tenant has at least one COMPANY unit; it is always a root node (parent_id = NULL).
	UnitTypeCompany UnitType = "COMPANY"

	// UnitTypeSubsidiary is a legally distinct subsidiary under a parent company.
	UnitTypeSubsidiary UnitType = "SUBSIDIARY"

	// UnitTypeRegion is a geographic grouping of branches or locations.
	UnitTypeRegion UnitType = "REGION"

	// UnitTypeBranch is a physical branch office or retail outlet.
	UnitTypeBranch UnitType = "BRANCH"

	// UnitTypeLocation is a specific physical address within a branch or company.
	UnitTypeLocation UnitType = "LOCATION"

	// UnitTypeDepartment is a functional grouping of staff within a region or company.
	UnitTypeDepartment UnitType = "DEPARTMENT"

	// UnitTypeDivision is a business division (e.g. sales, engineering).
	UnitTypeDivision UnitType = "DIVISION"

	// UnitTypeCostCenter is an accounting cost centre tracked for budget and reporting.
	UnitTypeCostCenter UnitType = "COST_CENTER"

	// UnitTypeProject is a time-bounded initiative scoped under a department or division.
	UnitTypeProject UnitType = "PROJECT"

	// UnitTypeBudgetUnit is a financial planning unit for budget allocation.
	UnitTypeBudgetUnit UnitType = "BUDGET_UNIT"
)

// IsValid reports whether t is a recognised UnitType constant.
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
//
// Note: the database PK column is named "uuid", not "id". All repository
// operations must use "uuid" as the column name.
type Unit struct {
	// ID is the primary key (maps to org_units.uuid).
	ID uuid.UUID

	// TenantID scopes this unit to a tenant (RLS key).
	TenantID uuid.UUID

	// ParentID is the immediate parent unit. nil for root COMPANY units.
	ParentID *uuid.UUID

	// Name is the display name, unique within a tenant.
	Name string

	// Code is an optional short reference code (e.g. "NAI-SALES"), unique per tenant when set.
	Code *string

	// Type classifies the unit's level and purpose in the hierarchy.
	Type UnitType

	// IsActive indicates whether the unit is operationally live.
	IsActive bool

	// Hidden hides the unit from standard listings without deactivating it.
	Hidden bool

	// AccrualMethod: true = accrual accounting, false = cash basis.
	AccrualMethod bool

	// FYStartMonth is the fiscal year start month (1 = Jan … 12 = Dec).
	FYStartMonth int

	// Address is a flexible JSON object (street, city, postcode, country, etc.).
	Address map[string]any

	// Picture is a logo file path or URL for the unit.
	Picture *string

	// OrgUnitPath is the materialized ancestor path: /root_uuid/parent_uuid/this_uuid/.
	// Maintained by the application layer on create/reparent.
	// Used for O(1) subtree membership: WHERE org_unit_path LIKE '/ancestor_uuid/%'.
	OrgUnitPath *string

	// OrgLevel is the hierarchy depth: 1 = root COMPANY, max 8.
	OrgLevel int

	// Settings holds per-unit configuration overrides (document numbering, feature flags, etc.).
	Settings map[string]any

	// Metadata holds arbitrary key-value data for integrations and extensions.
	Metadata map[string]any

	// Version is the optimistic-locking counter; incremented on every UPDATE.
	Version int

	// ValidationStatus is the outcome of the last validation job.
	ValidationStatus ValidationStatus

	// ValidationErrors holds structured validation error objects from the last run.
	ValidationErrors []any

	LastValidationRun *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

// IsRoot reports whether this unit is a root node (no parent).
func (u *Unit) IsRoot() bool { return u.ParentID == nil }

// IsDeleted reports whether this unit has been soft-deleted.
func (u *Unit) IsDeleted() bool { return u.DeletedAt != nil }

// PathPrefix returns the LIKE pattern that matches this unit and all its descendants.
// Example: "/abc123/" → "/abc123/%"
//
// Returns ("", false) when OrgUnitPath has not been set.
func (u *Unit) PathPrefix() (string, bool) {
	if u.OrgUnitPath == nil || *u.OrgUnitPath == "" {
		return "", false
	}
	return *u.OrgUnitPath + "%", true
}

// UnitPath is a row from the org_unit_paths closure table.
// It records every ancestor–descendant pair including self-references at depth 0.
type UnitPath struct {
	TenantID     uuid.UUID
	AncestorID   uuid.UUID
	DescendantID uuid.UUID
	// Depth is 0 for self-references, 1 for direct parent-child, 2+ for transitive.
	Depth     int
	CreatedAt time.Time
	UpdatedAt time.Time
}
