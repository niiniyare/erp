package organization

import (
	"context"

	"github.com/google/uuid"
)

// Service defines organization business logic interface
type Service interface {
	CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*Organization, error)
	GetOrganization(ctx context.Context, id uuid.UUID) (*Organization, error)
	UpdateOrganization(ctx context.Context, id uuid.UUID, req UpdateOrganizationRequest) error
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
	AssignManager(ctx context.Context, orgID, userID uuid.UUID) error
}

// CreateOrganizationRequest represents organization creation request
type CreateOrganizationRequest struct {
	TenantID    uuid.UUID  `json:"tenant_id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Name        string     `json:"name"`
	Code        string     `json:"code,omitempty"`
	Type        string     `json:"type"`
	Description string     `json:"description,omitempty"`
	ManagerID   *uuid.UUID `json:"manager_id,omitempty"`
}

// UpdateOrganizationRequest represents organization update request
type UpdateOrganizationRequest struct {
	Name        *string    `json:"name,omitempty"`
	Code        *string    `json:"code,omitempty"`
	Description *string    `json:"description,omitempty"`
	ManagerID   *uuid.UUID `json:"manager_id,omitempty"`
}
