package org

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ScopeLevel defines how an entity's data rows are scoped within the org hierarchy.
type ScopeLevel string

const (
	ScopeLevelGlobal ScopeLevel = "global"
	ScopeLevelTenant ScopeLevel = "tenant"
	ScopeLevelUnit   ScopeLevel = "unit"
)

// Scope is the organisational context carried on every authenticated request.
type Scope struct {
	TenantID uuid.UUID
	UnitID   uuid.UUID
}

func TenantOnly(tenantID uuid.UUID) Scope { return Scope{TenantID: tenantID} }
func WithUnit(tenantID, unitID uuid.UUID) Scope {
	return Scope{TenantID: tenantID, UnitID: unitID}
}
func (s Scope) IsTenantWide() bool { return s.UnitID == uuid.Nil }
func (s Scope) Validate() error {
	if s.TenantID == uuid.Nil {
		return errors.New("org scope: tenant_id is required")
	}
	return nil
}

func (s Scope) String() string {
	if s.IsTenantWide() {
		return fmt.Sprintf("tenant:%s", s.TenantID)
	}
	return fmt.Sprintf("tenant:%s/unit:%s", s.TenantID, s.UnitID)
}

// NodeType is an alias for UnitType retained for migration compatibility.
type NodeType = UnitType

var (
	ErrCrossTenantAccess = errors.New("cross-tenant data access is forbidden")
	ErrCrossUnitAccess   = errors.New("cross-unit data access is forbidden")
	ErrScopeInsufficient = errors.New("organisational scope is insufficient for this operation")
)
