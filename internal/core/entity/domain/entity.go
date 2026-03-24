package domain

import (
	"time"

	"github.com/google/uuid"
)

// EntityType classifies the role of an organizational unit.
type EntityType string

const (
	EntityTypeCompany    EntityType = "company"
	EntityTypeSubsidiary EntityType = "subsidiary"
	EntityTypeDepartment EntityType = "department"
	EntityTypeRegion     EntityType = "region"
	EntityTypeBranch     EntityType = "branch"
)

// EntityNode is the domain representation of an organisational entity.
// It maps to the `entities` DB table with the fields relevant to IAM and
// data-scoping concerns. More detailed fields (address, picture, accrual
// settings) live only in the DB layer and are not needed here.
type EntityNode struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	ParentID    *uuid.UUID
	Name        string
	Code        *string
	Type        EntityType
	IsActive    bool
	EntityPath  string // materialized path: /uuid1/uuid2/this_uuid/
	EntityLevel int32  // 1 = root (company); increments per level, max 8
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateEntityRequest is the input for inserting a new entity.
type CreateEntityRequest struct {
	// ParentID is nil for root (company-level) entities.
	ParentID *uuid.UUID
	Name     string
	Type     EntityType
	// Code is an optional short reference code (e.g. "HQ", "NORTH").
	Code     *string
	IsActive bool
}
