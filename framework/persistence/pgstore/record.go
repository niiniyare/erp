// Package pgstore provides a pgx-backed implementation of persistence.EntityStore
// driven entirely by EntityDefinition metadata.
package pgstore

import (
	"github.com/google/uuid"
	"awo.so/framework/definition"
)

// mapRecord is a concrete implementation of definition.Record and
// definition.MutableRecord backed by a plain map.
type mapRecord struct {
	id         uuid.UUID
	tenantID   uuid.UUID
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

func (r *mapRecord) Get(field string) any          { return r.data[field] }
func (r *mapRecord) Set(field string, value any)   { r.data[field] = value }
func (r *mapRecord) ID() uuid.UUID                 { return r.id }
func (r *mapRecord) TenantID() uuid.UUID           { return r.tenantID }
func (r *mapRecord) EntityName() string            { return r.entityName }

// ensure interfaces satisfied at compile time
var _ definition.Record        = (*mapRecord)(nil)
var _ definition.MutableRecord = (*mapRecord)(nil)
