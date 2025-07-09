package organization

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
)

// Organization represents an organizational unit
type Organization struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	ParentID    *uuid.UUID             `json:"parent_id,omitempty"`
	Name        string                 `json:"name"`
	Code        string                 `json:"code,omitempty"`
	Type        OrganizationType       `json:"type"`
	Description string                 `json:"description,omitempty"`
	ManagerID   *uuid.UUID             `json:"manager_id,omitempty"`
	IsActive    bool                   `json:"is_active"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`

	// Relations (loaded separately)
	Parent   *Organization   `json:"parent,omitempty"`
	Children []*Organization `json:"children,omitempty"`
	// Manager  *User          `json:"manager,omitempty"` // From user domain
}

// OrganizationType defines types of organizational units
type OrganizationType string

const (
	OrgTypeDepartment OrganizationType = "department"
	OrgTypeRegion     OrganizationType = "region"
	OrgTypeDivision   OrganizationType = "division"
	OrgTypeTeam       OrganizationType = "team"
	OrgTypeOffice     OrganizationType = "office"
)

// OrganizationUser represents user assignment to organization
type OrganizationUser struct {
	ID             uuid.UUID `json:"id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	UserID         uuid.UUID `json:"user_id"`
	Role           string    `json:"role"`
	IsPrimary      bool      `json:"is_primary"`
	JoinedAt       time.Time `json:"joined_at"`
}

func FromSQLCEntity(e db.Entity) *Organization {
	var metadata map[string]interface{}
	if e.Settings != nil {
		_ = json.Unmarshal(e.Settings, &metadata)
	}

	var code string
	if e.Code != nil {
		code = *e.Code
	}

	return &Organization{
		ID:        e.Uuid,
		TenantID:  e.TenantID,
		ParentID:  e.ParentID,
		Name:      e.Name,
		Code:      code,
		Type:      OrganizationType(e.Type),
		IsActive:  e.IsActive,
		Metadata:  metadata,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}
