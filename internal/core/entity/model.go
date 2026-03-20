package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	db "awo/db/sqlc"
)

// Entity represents a business entity in the system (domain model)
type Entity struct {
	ID        uuid.UUID      `json:"id"`
	ParentID  *uuid.UUID     `json:"parent_id,omitempty"`
	Name      string         `json:"name"`
	Code      string         `json:"code"`
	Type      EntityType     `json:"type"`
	IsActive  bool           `json:"is_active"`
	IsHidden  bool           `json:"is_hidden"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt *time.Time     `json:"deleted_at,omitempty"`
}

// EntityType represents the type of entity
type EntityType string

const (
	EntityTypeCompany    EntityType = "COMPANY"
	EntityTypeSubsidiary EntityType = "SUBSIDIARY"
	EntityTypeRegion     EntityType = "REGION"
	EntityTypeBranch     EntityType = "BRANCH"
	EntityTypeLocation   EntityType = "LOCATION"
	EntityTypeDepartment EntityType = "DEPARTMENT"
	EntityTypeDivision   EntityType = "DIVISION"
	EntityTypeCostCenter EntityType = "COST_CENTER"
	EntityTypeProject    EntityType = "PROJECT"
	EntityTypeBudgetUnit EntityType = "BUDGET_UNIT"
)

// EntityWithHierarchy represents entity with hierarchy information
type EntityWithHierarchy struct {
	Entity
	Level       int                   `json:"level"`
	Path        string                `json:"path"`
	HasChildren bool                  `json:"has_children"`
	Children    []EntityWithHierarchy `json:"children,omitempty"`
}

// EntityState represents state information for an entity
type EntityState struct {
	ID         uuid.UUID `json:"id"`
	EntityID   uuid.UUID `json:"entity_id"`
	Key        string    `json:"key"`
	Sequence   int64     `json:"sequence"`
	FiscalYear int       `json:"fiscal_year"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// HierarchyPath represents the hierarchy path between entities
type HierarchyPath struct {
	AncestorID   uuid.UUID `json:"ancestor_id"`
	DescendantID uuid.UUID `json:"descendant_id"`
	Depth        int       `json:"depth"`
}

// CreateEntityRequest represents entity creation request
type CreateEntityRequest struct {
	ParentID      *uuid.UUID     `json:"parent_id,omitempty"`
	Name          string         `json:"name" validate:"required,min=2,max=100"`
	Code          string         `json:"code" validate:"required,min=2,max=50"`
	Type          EntityType     `json:"type" validate:"required"`
	IsActive      bool           `json:"is_active"`
	IsHidden      bool           `json:"is_hidden"`
	AccrualMethod bool           `json:"accrual_method"` // TRUE = Accrual, FALSE = Cash
	FYStartMonth  int            `json:"fy_start_month"` // Fiscal year start month (1-12)
	Address       map[string]any `json:"address,omitempty"`
	Picture       string         `json:"picture,omitempty"`
	Settings      map[string]any `json:"settings,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

// UpdateEntityRequest represents entity update request
type UpdateEntityRequest struct {
	ParentID *uuid.UUID     `json:"parent_id,omitempty"`
	Name     *string        `json:"name,omitempty"`
	Code     *string        `json:"code,omitempty"`
	Type     *EntityType    `json:"type,omitempty"`
	IsActive *bool          `json:"is_active,omitempty"`
	IsHidden *bool          `json:"is_hidden,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ListEntitiesRequest represents entity list request with filters
type ListEntitiesRequest struct {
	Type     *EntityType `json:"type,omitempty"`
	ParentID *uuid.UUID  `json:"parent_id,omitempty"`
	IsActive *bool       `json:"is_active,omitempty"`
	IsHidden *bool       `json:"is_hidden,omitempty"`
	Limit    int         `json:"limit"`
	Offset   int         `json:"offset"`
}

// EntitySequenceRequest represents sequence number request
type EntitySequenceRequest struct {
	EntityID   uuid.UUID `json:"entity_id" validate:"required"`
	Key        string    `json:"key" validate:"required"`
	FiscalYear int       `json:"fiscal_year" validate:"required"`
}

// FromSQLCEntity converts SQLC Entity to domain Entity
func FromSQLCEntity(sqlcEntity *db.Entity) (*Entity, error) {
	var metadata map[string]any
	if len(sqlcEntity.Metadata) > 0 {
		if err := json.Unmarshal(sqlcEntity.Metadata, &metadata); err != nil {
			return nil, err
		}
	}

	var deletedAt *time.Time
	if sqlcEntity.DeletedAt.Valid {
		deletedAt = &sqlcEntity.DeletedAt.Time
	}

	var parentID *uuid.UUID
	if sqlcEntity.ParentID == nil {
		parentID = sqlcEntity.ParentID
	}

	return &Entity{
		ID:        sqlcEntity.Uuid,
		ParentID:  parentID,
		Name:      sqlcEntity.Name,
		Code:      *sqlcEntity.Code,
		Type:      EntityType(sqlcEntity.Type),
		IsActive:  sqlcEntity.IsActive,
		IsHidden:  sqlcEntity.Hidden,
		Metadata:  metadata,
		CreatedAt: sqlcEntity.CreatedAt,
		UpdatedAt: sqlcEntity.UpdatedAt,
		DeletedAt: deletedAt,
	}, nil
}

// ToSQLCCreateParams converts domain CreateEntityRequest to SQLC params
func (req *CreateEntityRequest) ToSQLCCreateParams() (db.CreateEntityParams, error) {
	var metadata []byte
	if req.Metadata != nil {
		var err error
		metadata, err = json.Marshal(req.Metadata)
		if err != nil {
			return db.CreateEntityParams{}, err
		}
	}

	var settings []byte
	if req.Settings != nil {
		var err error
		settings, err = json.Marshal(req.Settings)
		if err != nil {
			return db.CreateEntityParams{}, err
		}
	}

	var address []byte
	if req.Address != nil {
		var err error
		address, err = json.Marshal(req.Address)
		if err != nil {
			return db.CreateEntityParams{}, err
		}
	}

	params := db.CreateEntityParams{
		Name:          req.Name,
		Code:          &req.Code,
		Type:          string(req.Type),
		IsActive:      req.IsActive,
		Hidden:        req.IsHidden,
		AccrualMethod: req.AccrualMethod,
		FyStartMonth:  int32(req.FYStartMonth),
		Address:       address,
		Picture:       &req.Picture,
		Settings:      settings,
		Metadata:      metadata,
	}

	if req.ParentID != nil {
		params.ParentID = req.ParentID
	}

	return params, nil
}

// FromSQLCEntityState converts SQLC EntityState to domain EntityState
func FromSQLCEntityState(sqlcState *db.Entitystate) *EntityState {
	return &EntityState{
		ID:         sqlcState.EntityID,
		EntityID:   sqlcState.EntityID,
		Key:        sqlcState.Key,
		Sequence:   sqlcState.Sequence,
		FiscalYear: int(*sqlcState.FiscalYear),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// FromSQLCHierarchyPath converts SQLC HierarchyPath to domain HierarchyPath
func FromSQLCHierarchyPath(sqlcPath *db.HierarchyPath) *HierarchyPath {
	return &HierarchyPath{
		AncestorID:   sqlcPath.AncestorID,
		DescendantID: sqlcPath.DescendantID,
		Depth:        int(sqlcPath.Depth),
	}
}

// IsValid checks if the entity type is valid
func (et EntityType) IsValid() bool {
	switch et {
	case EntityTypeCompany, EntityTypeSubsidiary, EntityTypeRegion, EntityTypeBranch,
		EntityTypeLocation, EntityTypeDepartment, EntityTypeDivision, EntityTypeCostCenter,
		EntityTypeProject, EntityTypeBudgetUnit:
		return true
	default:
		return false
	}
}

// String returns the string representation of EntityType
func (et EntityType) String() string {
	return string(et)
}
