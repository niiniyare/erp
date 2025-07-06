package tenant

import (
    "time"
    "github.com/google/uuid"
)

// Tenant represents a tenant in the system
type Tenant struct {
    ID        uuid.UUID          `json:"id"`
    Name      string             `json:"name"`
    Subdomain string             `json:"subdomain"`
    PlanType  PlanType           `json:"plan_type"`
    Status    Status             `json:"status"`
    Settings  map[string]interface{} `json:"settings"`
    CreatedAt time.Time          `json:"created_at"`
    UpdatedAt time.Time          `json:"updated_at"`
}

// PlanType represents subscription plans
type PlanType string

const (
    PlanTypeBasic      PlanType = "basic"
    PlanTypeProfessional PlanType = "professional"
    PlanTypeEnterprise PlanType = "enterprise"
)

// Status represents tenant status
type Status string

const (
    StatusActive    Status = "active"
    StatusSuspended Status = "suspended"
    StatusTrialI    Status = "trial"
)

// CreateTenantRequest represents tenant creation request
type CreateTenantRequest struct {
    Name      string             `json:"name" validate:"required,min=2,max=100"`
    Subdomain string             `json:"subdomain" validate:"required,min=3,max=50,alphanum"`
    PlanType  PlanType           `json:"plan_type"`
    Settings  map[string]interface{} `json:"settings,omitempty"`
}

// UpdateTenantRequest represents tenant update request  
type UpdateTenantRequest struct {
    Name     *string            `json:"name,omitempty"`
    PlanType *PlanType          `json:"plan_type,omitempty"`
    Status   *Status            `json:"status,omitempty"`
    Settings map[string]interface{} `json:"settings,omitempty"`
}