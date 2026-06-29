// Package pgstore provides a pgx-backed implementation of persistence.EntityStore
// driven entirely by EntityDefinition metadata.
package pgstore

import (
	"awo.so/framework/def"
	"github.com/google/uuid"
)

// mapRecord is a concrete implementation of def.Record and
// def.MutableRecord backed by a plain map.
type mapRecord struct {
	id         uuid.UUID
	tenantID   uuid.UUID
	orgUnitID  uuid.UUID // uuid.Nil for non-unit-scoped entities
	entityName string
	data       map[string]any
}

func newMapRecord(entityName string, tenantID uuid.UUID) *mapRecord {
	return &mapRecord{
		entityName: entityName,
		tenantID:   tenantID,
		data:       make(map[string]any),
	}
}

func (r *mapRecord) Get(field string) any        { return r.data[field] }
func (r *mapRecord) Set(field string, value any) { r.data[field] = value }
func (r *mapRecord) ID() uuid.UUID               { return r.id }
func (r *mapRecord) TenantID() uuid.UUID         { return r.tenantID }
func (r *mapRecord) OrgUnitID() uuid.UUID        { return r.orgUnitID }
func (r *mapRecord) EntityName() string          { return r.entityName }

// RecordOrgUnitID implements def.OrgScoped so mapRecord works directly
// with AllowWithinOrgScope policy without any wrapping.
func (r *mapRecord) RecordOrgUnitID() uuid.UUID { return r.orgUnitID }

// ensure interfaces satisfied at compile time
var (
	_ def.Record        = (*mapRecord)(nil)
	_ def.MutableRecord = (*mapRecord)(nil)
	_ def.OrgScoped     = (*mapRecord)(nil)
)
