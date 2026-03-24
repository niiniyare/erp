// Package entity is the bounded context for organizational entity management.
//
// An "entity" in this system is an organisational unit within a tenant —
// for example a Company, Subsidiary, Region, Branch, or Department. Entities
// form a tree hierarchy that determines each user's data-visibility scope
// (EntityScope) at login time.
//
// External code should import only this package, never the sub-packages
// (domain, repository, service) directly.
package entity

import (
	db "awo.so/db/sqlc"
	"awo.so/internal/core/entity/domain"
	"awo.so/internal/core/entity/repository"
	"awo.so/internal/core/entity/service"
)

// ─── Domain type aliases ───────────────────────────────────────────────────

type (
	EntityNode          = domain.EntityNode
	EntityType          = domain.EntityType
	CreateEntityRequest = domain.CreateEntityRequest
)

// Entity type constants.
const (
	TypeCompany    = domain.EntityTypeCompany
	TypeSubsidiary = domain.EntityTypeSubsidiary
	TypeDepartment = domain.EntityTypeDepartment
	TypeRegion     = domain.EntityTypeRegion
	TypeBranch     = domain.EntityTypeBranch
)

// ─── Service interface ─────────────────────────────────────────────────────

// Service is the entity service interface.
type Service = service.Service

// ─── Constructors ──────────────────────────────────────────────────────────

// NewRepository returns a Postgres-backed entity repository.
func NewRepository(store db.Store) repository.Repository {
	return repository.NewPostgres(store)
}

// NewService returns a fully wired entity service.
func NewService(store db.Store) service.Service {
	repo := repository.NewPostgres(store)
	return service.NewService(repo, store)
}
