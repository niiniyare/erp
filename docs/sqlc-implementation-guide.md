# SQLC-Centric Implementation Patterns and Production Refinement Guide

## Executive Summary

This guide provides detailed implementation patterns that strictly adhere to SQLC-based development while maintaining Clean Architecture principles. All database interactions use SQLC-generated code exclusively, with comprehensive examples for hierarchy operations, bulk processing, and production-ready patterns compatible with Termux constraints.

## 1. SQLC Query Design for Hierarchy Operations

### **Enhanced Hierarchy Queries with SQLC Directives**

```sql
-- File: db/queries/hierarchy_operations.sql
-- Enhanced hierarchy operations using closure table with SQLC optimization

-- name: GetEntitySubtreeWithDepth :many
WITH RECURSIVE entity_tree AS (
    -- Root entity
    SELECT 
        e.uuid,
        e.name,
        e.type,
        e.parent_id,
        e.settings,
        e.metadata,
        0 as depth,
        e.uuid::text as path
    FROM entities e
    WHERE e.uuid = $1 
      AND e.tenant_id = current_tenant_id()
      AND e.deleted_at IS NULL
    
    UNION ALL
    
    -- Children recursively
    SELECT 
        child.uuid,
        child.name,
        child.type,
        child.parent_id,
        child.settings,
        child.metadata,
        parent.depth + 1,
        parent.path || '/' || child.uuid::text
    FROM entities child
    INNER JOIN entity_tree parent ON child.parent_id = parent.uuid
    WHERE child.tenant_id = current_tenant_id()
      AND child.deleted_at IS NULL
      AND parent.depth < $2  -- depth limit parameter
)
SELECT 
    uuid,
    name,
    type,
    parent_id,
    settings,
    metadata,
    depth,
    path
FROM entity_tree
ORDER BY depth, name;

-- name: BulkCreateHierarchyPaths :copyfrom
INSERT INTO hierarchy_paths (
    tenant_id,
    ancestor_id,
    descendant_id,
    depth,
    created_at,
    updated_at
) VALUES (
    current_tenant_id(),
    $1,
    $2,
    $3,
    NOW(),
    NOW()
);

-- name: GetEntityAncestorsOptimized :many
SELECT 
    e.uuid,
    e.name,
    e.type,
    hp.depth,
    e.settings,
    e.metadata
FROM hierarchy_paths hp
INNER JOIN entities e ON e.uuid = hp.ancestor_id
WHERE hp.tenant_id = current_tenant_id()
  AND hp.descendant_id = $1
  AND hp.depth > 0
  AND e.deleted_at IS NULL
ORDER BY hp.depth DESC;

-- name: ValidateCircularReference :one
SELECT EXISTS(
    SELECT 1 
    FROM hierarchy_paths hp
    WHERE hp.tenant_id = current_tenant_id()
      AND hp.ancestor_id = $2  -- potential new parent
      AND hp.descendant_id = $1  -- entity being moved
      AND hp.depth > 0
) as would_create_cycle;

-- name: GetSubtreeForReplication :many
SELECT 
    e.uuid,
    e.name,
    e.code,
    e.type,
    e.parent_id,
    e.is_active,
    e.hidden,
    e.accrual_method,
    e.fy_start_month,
    e.address,
    e.picture,
    e.settings,
    e.metadata,
    hp.depth
FROM entities e
INNER JOIN hierarchy_paths hp ON e.uuid = hp.descendant_id
WHERE hp.tenant_id = current_tenant_id()
  AND hp.ancestor_id = $1  -- root of subtree
  AND ($2 = 0 OR hp.depth <= $2)  -- max depth (0 = unlimited)
  AND e.deleted_at IS NULL
  AND ($3::boolean = true OR e.is_active = true)  -- include inactive flag
ORDER BY hp.depth, e.name;

-- name: BulkUpdateEntityParents :exec
UPDATE entities 
SET 
    parent_id = new_parents.new_parent_id,
    updated_at = NOW(),
    updated_by = $3
FROM (
    SELECT 
        unnest($1::uuid[]) as entity_id,
        unnest($2::uuid[]) as new_parent_id
) as new_parents
WHERE entities.uuid = new_parents.entity_id
  AND entities.tenant_id = current_tenant_id()
  AND entities.deleted_at IS NULL;

-- name: RebuildHierarchyPathsForEntity :exec
-- Rebuild hierarchy paths after entity move
WITH RECURSIVE entity_hierarchy AS (
    -- Self-reference (depth 0)
    SELECT 
        $1::uuid as ancestor_id,
        $1::uuid as descendant_id,
        0 as depth
    
    UNION ALL
    
    -- All ancestors from parent chain
    SELECT 
        hp.ancestor_id,
        $1::uuid as descendant_id,
        hp.depth + 1
    FROM hierarchy_paths hp
    INNER JOIN entities e ON e.uuid = $1
    WHERE hp.tenant_id = current_tenant_id()
      AND hp.descendant_id = e.parent_id
      AND e.parent_id IS NOT NULL
),
descendant_paths AS (
    -- All descendants that need path updates
    SELECT 
        $1::uuid as ancestor_id,
        hp.descendant_id,
        hp.depth + 1 as depth
    FROM hierarchy_paths hp
    WHERE hp.tenant_id = current_tenant_id()
      AND hp.ancestor_id = $1
      AND hp.depth > 0
),
all_new_paths AS (
    SELECT ancestor_id, descendant_id, depth FROM entity_hierarchy
    UNION ALL
    SELECT ancestor_id, descendant_id, depth FROM descendant_paths
)
-- Delete old paths and insert new ones
DELETE FROM hierarchy_paths 
WHERE tenant_id = current_tenant_id()
  AND (ancestor_id = $1 OR descendant_id = $1);

INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth, created_at, updated_at)
SELECT 
    current_tenant_id(),
    ancestor_id,
    descendant_id,
    depth,
    NOW(),
    NOW()
FROM all_new_paths;

-- name: GetDepthValidationInfo :one
WITH entity_depth AS (
    SELECT COALESCE(MAX(hp.depth), 0) as current_depth
    FROM hierarchy_paths hp
    WHERE hp.tenant_id = current_tenant_id()
      AND hp.descendant_id = $1  -- entity being moved
),
parent_depth AS (
    SELECT COALESCE(MAX(hp.depth), 0) as parent_depth
    FROM hierarchy_paths hp
    WHERE hp.tenant_id = current_tenant_id()
      AND hp.descendant_id = $2  -- new parent
),
subtree_depth AS (
    SELECT COALESCE(MAX(hp.depth), 0) as max_subtree_depth
    FROM hierarchy_paths hp
    WHERE hp.tenant_id = current_tenant_id()
      AND hp.ancestor_id = $1  -- entity being moved
)
SELECT 
    ed.current_depth,
    pd.parent_depth,
    sd.max_subtree_depth,
    (pd.parent_depth + 1) as new_entity_depth,
    (pd.parent_depth + 1 + sd.max_subtree_depth) as max_resulting_depth,
    CASE 
        WHEN (pd.parent_depth + 1 + sd.max_subtree_depth) > $3 THEN false
        ELSE true
    END as is_valid
FROM entity_depth ed, parent_depth pd, subtree_depth sd;

-- name: BulkValidateEntityMoves :many
WITH move_validations AS (
    SELECT 
        moves.entity_id,
        moves.new_parent_id,
        COALESCE(parent_depth.depth, 0) as parent_depth,
        COALESCE(subtree_depth.max_depth, 0) as subtree_depth,
        COALESCE(circular_check.has_cycle, false) as would_create_cycle
    FROM (
        SELECT 
            unnest($1::uuid[]) as entity_id,
            unnest($2::uuid[]) as new_parent_id
    ) moves
    LEFT JOIN LATERAL (
        SELECT MAX(hp.depth) as depth
        FROM hierarchy_paths hp
        WHERE hp.tenant_id = current_tenant_id()
          AND hp.descendant_id = moves.new_parent_id
    ) parent_depth ON true
    LEFT JOIN LATERAL (
        SELECT MAX(hp.depth) as max_depth
        FROM hierarchy_paths hp
        WHERE hp.tenant_id = current_tenant_id()
          AND hp.ancestor_id = moves.entity_id
    ) subtree_depth ON true
    LEFT JOIN LATERAL (
        SELECT EXISTS(
            SELECT 1 FROM hierarchy_paths hp
            WHERE hp.tenant_id = current_tenant_id()
              AND hp.ancestor_id = moves.new_parent_id
              AND hp.descendant_id = moves.entity_id
              AND hp.depth > 0
        ) as has_cycle
    ) circular_check ON true
)
SELECT 
    entity_id,
    new_parent_id,
    parent_depth,
    subtree_depth,
    (parent_depth + 1 + subtree_depth) as max_resulting_depth,
    would_create_cycle,
    CASE 
        WHEN would_create_cycle THEN 'circular_reference'
        WHEN (parent_depth + 1 + subtree_depth) > $3 THEN 'depth_limit_exceeded'
        ELSE 'valid'
    END as validation_result
FROM move_validations;
```

### **SQLC Configuration for Hierarchy Operations**

```yaml
# sqlc.yaml - Enhanced configuration for hierarchy operations
version: 2
sql:
  - engine: "postgresql"
    queries: "db/queries/"
    schema: "db/migration/"
    gen:
      go:
        package: "db"
        out: "db/sqlc"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_db_tags: true
        emit_prepared_queries: true
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
        emit_exported_queries: false
        emit_result_struct_pointers: true
        emit_params_struct_pointers: false
        emit_methods_with_db_argument: false
        json_tags_case_style: "snake"
        overrides:
          - column: "hierarchy_paths.depth"
            go_type: "int32"
          - column: "entities.settings"
            go_type: "json.RawMessage"
          - column: "entities.metadata" 
            go_type: "json.RawMessage"
          - column: "entities.address"
            go_type: "json.RawMessage"
rules:
  - sqlc/db-prepare

## 2. Repository Layer Implementation with SQLC-Generated Code

### **Pure SQLC Repository Implementation**

```go
// internal/core/entity/repository/sqlc_repository.go
package repository

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"

    db "awo.so/db/sqlc"
    "awo.so/internal/core/entity/domain"
    "awo.so/internal/shared"
    "awo.so/internal/shared/tracing"
)

type SQLCEntityRepository struct {
    store   db.Store
    tracing tracing.TracingService
}

func NewSQLCEntityRepository(store db.Store, tracing tracing.TracingService) domain.EntityRepository {
    return &SQLCEntityRepository{
        store:   store,
        tracing: tracing,
    }
}

// Create entity using SQLC-generated code exclusively
func (r *SQLCEntityRepository) Create(ctx context.Context, req *domain.CreateEntityRequest) (*domain.Entity, error) {
    ctx, span := r.tracing.StartSpan(ctx, "SQLCEntityRepository.Create")
    defer span.End()

    // Get tenant ID from context
    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        return nil, fmt.Errorf("tenant ID not found in context")
    }

    // Use tenant-aware transaction
    return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) (*domain.Entity, error) {
        // Convert domain request to SQLC parameters
        params := db.CreateEntityParams{
            Uuid:          uuid.New(),
            ParentID:      req.ParentID,
            Name:          req.Name,
            Code:          pgtype.Text{String: req.Code, Valid: req.Code != ""},
            Type:          string(req.Type),
            IsActive:      req.IsActive,
            Hidden:        req.IsHidden,
            AccrualMethod: req.AccrualMethod,
            FyStartMonth:  int32(req.FYStartMonth),
            Address:       marshalJSONField(req.Address),
            Picture:       pgtype.Text{String: req.Picture, Valid: req.Picture != ""},
            Settings:      marshalJSONField(req.Settings),
            Metadata:      marshalJSONField(req.Metadata),
            CreatedAt:     time.Now(),
            UpdatedAt:     time.Now(),
            CreatedBy:     req.CreatedBy,
        }

        // Execute SQLC-generated create query
        sqlcEntity, err := s.CreateEntity(ctx, params)
        if err != nil {
            return nil, r.mapDatabaseError(err, "create_entity")
        }

        // Create hierarchy paths using SQLC bulk operation
        if err := r.createHierarchyPaths(ctx, s, sqlcEntity.Uuid, sqlcEntity.ParentID); err != nil {
            return nil, fmt.Errorf("failed to create hierarchy paths: %w", err)
        }

        // Convert SQLC entity to domain entity
        return r.sqlcEntityToDomain(&sqlcEntity)
    })
}

// GetByID using SQLC-generated query
func (r *SQLCEntityRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Entity, error) {
    ctx, span := r.tracing.StartSpan(ctx, "SQLCEntityRepository.GetByID")
    defer span.End()

    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        return nil, fmt.Errorf("tenant ID not found in context")
    }

    return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) (*domain.Entity, error) {
        sqlcEntity, err := s.GetEntityByID(ctx, id)
        if err != nil {
            if err == pgx.ErrNoRows {
                return nil, domain.ErrEntityNotFound
            }
            return nil, r.mapDatabaseError(err, "get_entity_by_id")
        }

        return r.sqlcEntityToDomain(&sqlcEntity)
    })
}

// GetSubtree using SQLC-generated recursive query
func (r *SQLCEntityRepository) GetSubtree(ctx context.Context, rootID uuid.UUID, maxDepth int) ([]*domain.EntityWithHierarchy, error) {
    ctx, span := r.tracing.StartSpan(ctx, "SQLCEntityRepository.GetSubtree")
    defer span.End()

    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        return nil, fmt.Errorf("tenant ID not found in context")
    }

    return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) ([]*domain.EntityWithHierarchy, error) {
        // Use SQLC-generated subtree query
        rows, err := s.GetEntitySubtreeWithDepth(ctx, db.GetEntitySubtreeWithDepthParams{
            Uuid:   rootID,
            Depth:  int32(maxDepth),
        })
        if err != nil {
            return nil, r.mapDatabaseError(err, "get_entity_subtree")
        }

        // Convert SQLC results to domain objects
        entities := make([]*domain.EntityWithHierarchy, len(rows))
        for i, row := range rows {
            entity, err := r.subtreeRowToDomain(&row)
            if err != nil {
                return nil, fmt.Errorf("failed to convert row %d: %w", i, err)
            }
            entities[i] = entity
        }

        return entities, nil
    })
}

// UpdateEntity using SQLC transaction pattern
func (r *SQLCEntityRepository) UpdateEntity(ctx context.Context, id uuid.UUID, req *domain.UpdateEntityRequest) (*domain.Entity, error) {
    ctx, span := r.tracing.StartSpan(ctx, "SQLCEntityRepository.UpdateEntity")
    defer span.End()

    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        return nil, fmt.Errorf("tenant ID not found in context")
    }

    return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) (*domain.Entity, error) {
        // Start transaction for atomic update
        tx, err := s.BeginTx(ctx, pgx.TxOptions{})
        if err != nil {
            return nil, fmt.Errorf("failed to begin transaction: %w", err)
        }
        defer tx.Rollback(ctx)

        // Get current entity for comparison
        currentEntity, err := s.GetEntityByID(ctx, id)
        if err != nil {
            return nil, r.mapDatabaseError(err, "get_current_entity")
        }

        // Validate parent change if requested
        if req.ParentID != nil && !uuid.Equal(*req.ParentID, currentEntity.ParentID.Bytes) {
            if err := r.validateParentChange(ctx, s, id, *req.ParentID); err != nil {
                return nil, err
            }
        }

        // Build update parameters
        params := r.buildUpdateParams(id, req, &currentEntity)

        // Execute SQLC-generated update
        updatedEntity, err := s.UpdateEntity(ctx, params)
        if err != nil {
            return nil, r.mapDatabaseError(err, "update_entity")
        }

        // Update hierarchy paths if parent changed
        if req.ParentID != nil && !uuid.Equal(*req.ParentID, currentEntity.ParentID.Bytes) {
            if err := s.RebuildHierarchyPathsForEntity(ctx, id); err != nil {
                return nil, fmt.Errorf("failed to rebuild hierarchy paths: %w", err)
            }
        }

        // Commit transaction
        if err := tx.Commit(ctx); err != nil {
            return nil, fmt.Errorf("failed to commit transaction: %w", err)
        }

        return r.sqlcEntityToDomain(&updatedEntity)
    })
}

// BulkCreateEntities using SQLC COPY FROM operation
func (r *SQLCEntityRepository) BulkCreateEntities(ctx context.Context, requests []*domain.CreateEntityRequest) ([]*domain.Entity, error) {
    ctx, span := r.tracing.StartSpan(ctx, "SQLCEntityRepository.BulkCreateEntities")
    defer span.End()

    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        return nil, fmt.Errorf("tenant ID not found in context")
    }

    return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) ([]*domain.Entity, error) {
        // Prepare bulk data for SQLC COPY FROM
        bulkParams := make([]db.BulkCreateEntitiesParams, len(requests))
        entityIDs := make([]uuid.UUID, len(requests))

        for i, req := range requests {
            entityID := uuid.New()
            entityIDs[i] = entityID
            
            bulkParams[i] = db.BulkCreateEntitiesParams{
                Uuid:          entityID,
                ParentID:      req.ParentID,
                Name:          req.Name,
                Code:          pgtype.Text{String: req.Code, Valid: req.Code != ""},
                Type:          string(req.Type),
                IsActive:      req.IsActive,
                Hidden:        req.IsHidden,
                AccrualMethod: req.AccrualMethod,
                FyStartMonth:  int32(req.FYStartMonth),
                Address:       marshalJSONField(req.Address),
                Picture:       pgtype.Text{String: req.Picture, Valid: req.Picture != ""},
                Settings:      marshalJSONField(req.Settings),
                Metadata:      marshalJSONField(req.Metadata),
                CreatedAt:     time.Now(),
                UpdatedAt:     time.Now(),
                CreatedBy:     req.CreatedBy,
            }
        }

        // Execute SQLC bulk create
        insertedCount, err := s.BulkCreateEntities(ctx, bulkParams)
        if err != nil {
            return nil, r.mapDatabaseError(err, "bulk_create_entities")
        }

        if insertedCount != int64(len(requests)) {
            return nil, fmt.Errorf("expected to insert %d entities, but inserted %d", len(requests), insertedCount)
        }

        // Create hierarchy paths in bulk
        if err := r.bulkCreateHierarchyPaths(ctx, s, bulkParams); err != nil {
            return nil, fmt.Errorf("failed to create bulk hierarchy paths: %w", err)
        }

        // Retrieve created entities
        entities := make([]*domain.Entity, len(entityIDs))
        for i, id := range entityIDs {
            entity, err := s.GetEntityByID(ctx, id)
            if err != nil {
                return nil, fmt.Errorf("failed to retrieve created entity %d: %w", i, err)
            }
            
            domainEntity, err := r.sqlcEntityToDomain(&entity)
            if err != nil {
                return nil, fmt.Errorf("failed to convert entity %d: %w", i, err)
            }
            entities[i] = domainEntity
        }

        return entities, nil
    })
}

// Hierarchy path management using SQLC
func (r *SQLCEntityRepository) createHierarchyPaths(ctx context.Context, s db.Store, entityID uuid.UUID, parentID pgtype.UUID) error {
    // Create self-reference path (depth 0)
    selfPathParams := db.BulkCreateHierarchyPathsParams{
        AncestorID:   entityID,
        DescendantID: entityID,
        Depth:        0,
    }

    _, err := s.BulkCreateHierarchyPaths(ctx, []db.BulkCreateHierarchyPathsParams{selfPathParams})
    if err != nil {
        return fmt.Errorf("failed to create self-reference path: %w", err)
    }

    // Create ancestor paths if entity has parent
    if parentID.Valid {
        ancestors, err := s.GetEntityAncestorsOptimized(ctx, parentID.Bytes)
        if err != nil {
            return fmt.Errorf("failed to get ancestors: %w", err)
        }

        // Prepare ancestor paths for bulk insert
        ancestorPaths := make([]db.BulkCreateHierarchyPathsParams, len(ancestors)+1)
        
        // Direct parent path
        ancestorPaths[0] = db.BulkCreateHierarchyPathsParams{
            AncestorID:   parentID.Bytes,
            DescendantID: entityID,
            Depth:        1,
        }

        // All other ancestor paths
        for i, ancestor := range ancestors {
            ancestorPaths[i+1] = db.BulkCreateHierarchyPathsParams{
                AncestorID:   ancestor.Uuid,
                DescendantID: entityID,
                Depth:        ancestor.Depth + 1,
            }
        }

        _, err = s.BulkCreateHierarchyPaths(ctx, ancestorPaths)
        if err != nil {
            return fmt.Errorf("failed to create ancestor paths: %w", err)
        }
    }

    return nil
}

// Validation using SQLC-generated queries
func (r *SQLCEntityRepository) validateParentChange(ctx context.Context, s db.Store, entityID, newParentID uuid.UUID) error {
    // Check for circular reference using SQLC query
    result, err := s.ValidateCircularReference(ctx, db.ValidateCircularReferenceParams{
        Uuid:   entityID,
        Uuid_2: newParentID,
    })
    if err != nil {
        return fmt.Errorf("failed to validate circular reference: %w", err)
    }

    if result {
        return domain.ErrCircularReference
    }

    // Validate depth limits using SQLC query
    depthInfo, err := s.GetDepthValidationInfo(ctx, db.GetDepthValidationInfoParams{
        Uuid:   entityID,
        Uuid_2: newParentID,
        Int4:   15, // Business limit
    })
    if err != nil {
        return fmt.Errorf("failed to validate depth: %w", err)
    }

    if !depthInfo.IsValid {
        return &domain.HierarchyDepthError{
            EntityID:          entityID,
            NewParentID:       newParentID,
            MaxResultingDepth: int(depthInfo.MaxResultingDepth),
            BusinessLimit:     15,
        }
    }

    return nil
}

// Helper functions for SQLC data conversion
func (r *SQLCEntityRepository) sqlcEntityToDomain(sqlcEntity *db.Entity) (*domain.Entity, error) {
    entity := &domain.Entity{
        ID:            sqlcEntity.Uuid,
        TenantID:      sqlcEntity.TenantID,
        ParentID:      convertPgUUID(sqlcEntity.ParentID),
        Name:          sqlcEntity.Name,
        Code:          convertPgText(sqlcEntity.Code),
        Type:          domain.EntityType(sqlcEntity.Type),
        IsActive:      sqlcEntity.IsActive,
        IsHidden:      sqlcEntity.Hidden,
        AccrualMethod: sqlcEntity.AccrualMethod,
        FYStartMonth:  int(sqlcEntity.FyStartMonth),
        Picture:       convertPgText(sqlcEntity.Picture),
        CreatedAt:     sqlcEntity.CreatedAt,
        UpdatedAt:     sqlcEntity.UpdatedAt,
        DeletedAt:     convertPgTimestamp(sqlcEntity.DeletedAt),
        CreatedBy:     sqlcEntity.CreatedBy,
        UpdatedBy:     convertPgText(sqlcEntity.UpdatedBy),
    }

    // Unmarshal JSON fields
    if err := unmarshalJSONField(sqlcEntity.Address, &entity.Address); err != nil {
        return nil, fmt.Errorf("failed to unmarshal address: %w", err)
    }

    if err := unmarshalJSONField(sqlcEntity.Settings, &entity.Settings); err != nil {
        return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
    }

    if err := unmarshalJSONField(sqlcEntity.Metadata, &entity.Metadata); err != nil {
        return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
    }

    return entity, nil
}

func (r *SQLCEntityRepository) subtreeRowToDomain(row *db.GetEntitySubtreeWithDepthRow) (*domain.EntityWithHierarchy, error) {
    entity := &domain.EntityWithHierarchy{
        Entity: domain.Entity{
            ID:       row.Uuid,
            Name:     row.Name,
            Type:     domain.EntityType(row.Type),
            ParentID: convertPgUUID(row.ParentID),
        },
        Level: int(row.Depth),
        Path:  row.Path,
    }

    // Unmarshal JSON fields
    if err := unmarshalJSONField(row.Settings, &entity.Settings); err != nil {
        return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
    }

    if err := unmarshalJSONField(row.Metadata, &entity.Metadata); err != nil {
        return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
    }

    return entity, nil
}

// Helper functions for type conversion
func convertPgUUID(pgUUID pgtype.UUID) *uuid.UUID {
    if !pgUUID.Valid {
        return nil
    }
    return &pgUUID.Bytes
}

func convertPgText(pgText pgtype.Text) string {
    if !pgText.Valid {
        return ""
    }
    return pgText.String
}

func convertPgTimestamp(pgTime pgtype.Timestamp) *time.Time {
    if !pgTime.Valid {
        return nil
    }
    return &pgTime.Time
}

func marshalJSONField(data interface{}) json.RawMessage {
    if data == nil {
        return json.RawMessage("{}")
    }
    bytes, err := json.Marshal(data)
    if err != nil {
        return json.RawMessage("{}")
    }
    return bytes
}

func unmarshalJSONField(raw json.RawMessage, target interface{}) error {
    if len(raw) == 0 {
        return nil
    }
    return json.Unmarshal(raw, target)
}

func (r *SQLCEntityRepository) mapDatabaseError(err error, operation string) error {
    // Database error mapping logic
    switch {
    case isPrimaryKeyViolation(err):
        return domain.ErrEntityAlreadyExists
    case isForeignKeyViolation(err):
        return domain.ErrInvalidParent
    case isUniqueViolation(err):
        return domain.ErrEntityNameExists
    default:
        return fmt.Errorf("%s failed: %w", operation, err)
    }
}

// Database error detection helpers
func isPrimaryKeyViolation(err error) bool {
    return contains(err.Error(), "duplicate key value violates unique constraint") &&
           contains(err.Error(), "_pkey")
}

func isForeignKeyViolation(err error) bool {
    return contains(err.Error(), "violates foreign key constraint")
}

func isUniqueViolation(err error) bool {
    return contains(err.Error(), "duplicate key value violates unique constraint") &&
           !contains(err.Error(), "_pkey")
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || 
           (len(s) > len(substr) && 
            (s[:len(substr)] == substr || 
             s[len(s)-len(substr):] == substr ||
             strings.Contains(s, substr))))
}
```

## 3. Temporal Activities with SQLC Integration

### **Temporal Activity Structure with SQLC**

```go
// internal/workflows/entity/activities.go
package entity

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "go.temporal.io/sdk/activity"
    "go.temporal.io/sdk/temporal"

    db "awo.so/db/sqlc"
    "awo.so/internal/core/entity/domain"
    "awo.so/internal/shared"
)

// TemporalEntityActivities wraps SQLC operations for Temporal workflows
type TemporalEntityActivities struct {
    store   db.Store
    logger  shared.Logger
}

func NewTemporalEntityActivities(store db.Store, logger shared.Logger) *TemporalEntityActivities {
    return &TemporalEntityActivities{
        store:  store,
        logger: logger,
    }
}

// BulkImportEntitiesActivity - Uses SQLC bulk operations within Temporal
func (a *TemporalEntityActivities) BulkImportEntitiesActivity(ctx context.Context, req BulkImportRequest) (*BulkImportResult, error) {
    // Configure activity with timeout and retry policy
    ctx = activity.WithTimeout(ctx, 30*time.Minute)
    ctx = activity.WithRetryPolicy(ctx, temporal.RetryPolicy{
        InitialInterval:    time.Second,
        BackoffCoefficient: 2.0,
        MaximumInterval:    time.Minute,
        MaximumAttempts:    5,
    })

    logger := activity.GetLogger(ctx)
    logger.Info("Starting bulk entity import", "tenant_id", req.TenantID, "batch_size", len(req.Entities))

    // Set tenant context for SQLC operations
    ctx = shared.WithTenantID(ctx, req.TenantID)

    return a.store.WithTenant(ctx, req.TenantID, func(ctx context.Context, s db.Store) (*BulkImportResult, error) {
        result := &BulkImportResult{
            TenantID:      req.TenantID,
            ProcessedAt:   time.Now(),
            SuccessCount:  0,
            ErrorCount:    0,
            Errors:        make([]BulkImportError, 0),
        }

        // Process entities in smaller batches for better memory usage
        batchSize := 100
        for i := 0; i < len(req.Entities); i += batchSize {
            end := i + batchSize
            if end > len(req.Entities) {
                end = len(req.Entities)
            }
            
            batch := req.Entities[i:end]
            
            // Record activity progress
            activity.RecordHeartbeat(ctx, map[string]interface{}{
                "progress":       float64(i) / float64(len(req.Entities)),
                "processed":      i,
                "total":         len(req.Entities),
                "current_batch": len(batch),
            })

            batchResult, err := a.processBatchWithSQLCTransactions(ctx, s, batch)
            if err != nil {
                logger.Error("Batch processing failed", "batch_start", i, "error", err)
                result.ErrorCount += len(batch)
                result.Errors = append(result.Errors, BulkImportError{
                    BatchStart: i,
                    BatchEnd:   end,
                    Error:      err.Error(),
                })
                continue
            }

            result.SuccessCount += batchResult.SuccessCount
            result.ErrorCount += batchResult.ErrorCount
            result.Errors = append(result.Errors, batchResult.Errors...)
        }

        logger.Info("Bulk import completed", 
            "success_count", result.SuccessCount,
            "error_count", result.ErrorCount,
            "total_processed", len(req.Entities))

        return result, nil
    })
}

// ProcessBatch using SQLC transactions for atomicity
func (a *TemporalEntityActivities) processBatchWithSQLCTransactions(
    ctx context.Context, 
    store db.Store, 
    batch []EntityImportData,
) (*BatchResult, error) {
    result := &BatchResult{
        SuccessCount: 0,
        ErrorCount:   0,
        Errors:       make([]BulkImportError, 0),
    }

    // Use SQLC transaction for entire batch
    tx, err := store.BeginTx(ctx, pgx.TxOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    // Prepare bulk insert parameters for SQLC
    bulkParams := make([]db.BulkCreateEntitiesParams, 0, len(batch))
    hierarchyParams := make([]db.BulkCreateHierarchyPathsParams, 0)

    for _, entityData := range batch {
        // Validate entity data before SQLC operation
        if err := a.validateEntityData(ctx, store, entityData); err != nil {
            result.ErrorCount++
            result.Errors = append(result.Errors, BulkImportError{
                EntityName: entityData.Name,
                Error:      fmt.Sprintf("validation failed: %v", err),
            })
            continue
        }

        entityID := uuid.New()
        
        // Build SQLC parameters
        params := db.BulkCreateEntitiesParams{
            Uuid:          entityID,
            ParentID:      entityData.ParentID,
            Name:          entityData.Name,
            Code:          pgtype.Text{String: entityData.Code, Valid: entityData.Code != ""},
            Type:          string(entityData.Type),
            IsActive:      entityData.IsActive,
            Hidden:        entityData.IsHidden,
            AccrualMethod: entityData.AccrualMethod,
            FyStartMonth:  int32(entityData.FYStartMonth),
            Address:       marshalJSONField(entityData.Address),
            Picture:       pgtype.Text{String: entityData.Picture, Valid: entityData.Picture != ""},
            Settings:      marshalJSONField(entityData.Settings),
            Metadata:      marshalJSONField(entityData.Metadata),
            CreatedAt:     time.Now(),
            UpdatedAt:     time.Now(),
            CreatedBy:     entityData.CreatedBy,
        }

        bulkParams = append(bulkParams, params)

        // Prepare hierarchy paths using SQLC queries
        hierarchyPaths, err := a.generateHierarchyPaths(ctx, store, entityID, entityData.ParentID)
        if err != nil {
            result.ErrorCount++
            result.Errors = append(result.Errors, BulkImportError{
                EntityName: entityData.Name,
                Error:      fmt.Sprintf("hierarchy path generation failed: %v", err),
            })
            continue
        }
        
        hierarchyParams = append(hierarchyParams, hierarchyPaths...)
        result.SuccessCount++
    }

    // Execute SQLC bulk operations within transaction
    if len(bulkParams) > 0 {
        insertedCount, err := store.BulkCreateEntities(ctx, bulkParams)
        if err != nil {
            return nil, fmt.Errorf("SQLC bulk entity creation failed: %w", err)
        }

        if insertedCount != int64(len(bulkParams)) {
            return nil, fmt.Errorf("expected %d entities, created %d", len(bulkParams), insertedCount)
        }

        // Insert hierarchy paths using SQLC
        if len(hierarchyParams) > 0 {
            pathCount, err := store.BulkCreateHierarchyPaths(ctx, hierarchyParams)
            if err != nil {
                return nil, fmt.Errorf("SQLC bulk hierarchy path creation failed: %w", err)
            }
            
            a.logger.Debug("Created hierarchy paths", "count", pathCount)
        }
    }

    // Commit transaction
    if err := tx.Commit(ctx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return result, nil
}
```

## 4. Testing Strategy for SQLC Code

### **Repository Testing with SQLC Mocks**

```go
// internal/core/entity/repository/repository_test.go
package repository_test

import (
    "context"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"

    db "awo.so/db/sqlc"
    "awo.so/internal/core/entity/domain"
    "awo.so/internal/core/entity/repository"
    "awo.so/internal/shared"
    "awo.so/internal/shared/mocks"
)

// SQLCRepositoryTestSuite tests repository with mocked SQLC store
type SQLCRepositoryTestSuite struct {
    suite.Suite
    mockStore   *mocks.MockStore
    repository  domain.EntityRepository
    tenantID    uuid.UUID
    ctx         context.Context
}

func (suite *SQLCRepositoryTestSuite) SetupTest() {
    suite.mockStore = mocks.NewMockStore(suite.T())
    suite.repository = repository.NewSQLCEntityRepository(suite.mockStore, &mocks.MockTracingService{})
    suite.tenantID = uuid.New()
    suite.ctx = shared.WithTenantID(context.Background(), suite.tenantID)
}

func (suite *SQLCRepositoryTestSuite) TestCreate_Success() {
    // Given
    req := &domain.CreateEntityRequest{
        Name:         "Test Entity",
        Code:         "TEST001",
        Type:         domain.EntityTypeDepartment,
        IsActive:     true,
        IsHidden:     false,
        AccrualMethod: "cash",
        FYStartMonth: 1,
        CreatedBy:    "test-user",
    }

    expectedEntity := db.Entity{
        Uuid:          uuid.New(),
        TenantID:      suite.tenantID,
        Name:          req.Name,
        Code:          pgtype.Text{String: req.Code, Valid: true},
        Type:          string(req.Type),
        IsActive:      req.IsActive,
        Hidden:        req.IsHidden,
        AccrualMethod: req.AccrualMethod,
        FyStartMonth:  int32(req.FYStartMonth),
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
        CreatedBy:     req.CreatedBy,
    }

    // Mock SQLC store tenant wrapper
    suite.mockStore.EXPECT().
        WithTenant(mock.Anything, suite.tenantID, mock.AnythingOfType("func(context.Context, db.Store) (*domain.Entity, error)")).
        RunAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) (*domain.Entity, error)) (*domain.Entity, error) {
            return fn(ctx, suite.mockStore)
        })

    // Mock SQLC CreateEntity call
    suite.mockStore.EXPECT().
        CreateEntity(mock.Anything, mock.MatchedBy(func(params db.CreateEntityParams) bool {
            return params.Name == req.Name && 
                   params.Type == string(req.Type) &&
                   params.IsActive == req.IsActive
        })).
        Return(expectedEntity, nil)

    // Mock hierarchy path creation
    suite.mockStore.EXPECT().
        BulkCreateHierarchyPaths(mock.Anything, mock.MatchedBy(func(params []db.BulkCreateHierarchyPathsParams) bool {
            return len(params) == 1 && params[0].Depth == 0  // Self-reference only
        })).
        Return(int64(1), nil)

    // When
    result, err := suite.repository.Create(suite.ctx, req)

    // Then
    suite.NoError(err)
    suite.NotNil(result)
    suite.Equal(req.Name, result.Name)
    suite.Equal(req.Type, result.Type)
    suite.Equal(req.IsActive, result.IsActive)
}
```

## 5. Performance Optimization with SQLC

### **Query Optimization Patterns**

```sql
-- File: db/queries/optimized_hierarchy.sql
-- Performance-optimized hierarchy queries using SQLC

-- name: GetEntitySubtreeOptimized :many
-- Fast subtree retrieval with minimal joins and proper indexing
SELECT 
    e.uuid,
    e.name,
    e.type,
    e.parent_id,
    e.is_active,
    hp.depth,
    hp.path_materialized
FROM entities e
INNER JOIN hierarchy_paths hp ON e.uuid = hp.descendant_id
WHERE hp.tenant_id = current_tenant_id()
  AND hp.ancestor_id = $1  -- root entity
  AND ($2 = 0 OR hp.depth <= $2)  -- max depth filter
  AND e.deleted_at IS NULL
  AND ($3::boolean = true OR e.is_active = true)  -- active filter
ORDER BY hp.depth, e.name
LIMIT $4;  -- pagination limit

-- name: GetHierarchyStatsBatch :many
-- Batch statistics calculation for multiple entities
WITH entity_stats AS (
    SELECT 
        hp.ancestor_id,
        COUNT(DISTINCT hp.descendant_id) as total_descendants,
        COUNT(DISTINCT CASE WHEN e.is_active THEN hp.descendant_id END) as active_descendants,
        MAX(hp.depth) as max_depth,
        COUNT(DISTINCT CASE WHEN hp.depth = 1 THEN hp.descendant_id END) as direct_children
    FROM hierarchy_paths hp
    INNER JOIN entities e ON e.uuid = hp.descendant_id
    WHERE hp.tenant_id = current_tenant_id()
      AND hp.ancestor_id = ANY($1::uuid[])  -- batch of entity IDs
      AND hp.depth > 0  -- exclude self-reference
      AND e.deleted_at IS NULL
    GROUP BY hp.ancestor_id
)
SELECT 
    ancestor_id as entity_id,
    total_descendants,
    active_descendants,
    max_depth,
    direct_children
FROM entity_stats;

-- name: BulkMoveEntitiesOptimized :exec
-- Optimized bulk entity move with hierarchy recalculation
WITH moved_entities AS (
    UPDATE entities 
    SET 
        parent_id = moves.new_parent_id,
        updated_at = NOW(),
        updated_by = $3
    FROM (
        SELECT 
            unnest($1::uuid[]) as entity_id,
            unnest($2::uuid[]) as new_parent_id
    ) moves
    WHERE entities.uuid = moves.entity_id
      AND entities.tenant_id = current_tenant_id()
      AND entities.deleted_at IS NULL
    RETURNING uuid, parent_id
),
deleted_paths AS (
    DELETE FROM hierarchy_paths 
    WHERE tenant_id = current_tenant_id()
      AND (ancestor_id IN (SELECT uuid FROM moved_entities) 
           OR descendant_id IN (SELECT uuid FROM moved_entities))
)
-- Trigger hierarchy rebuild for affected entities
INSERT INTO hierarchy_rebuild_queue (tenant_id, entity_id, created_at)
SELECT current_tenant_id(), uuid, NOW()
FROM moved_entities;
```

### **Caching Strategy with SQLC**

```go
// internal/core/entity/repository/cached_repository.go
package repository

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"

    db "awo.so/db/sqlc"
    "awo.so/internal/core/entity/domain"
    "awo.so/internal/shared"
)

// CachedSQLCRepository wraps SQLC repository with Redis caching
type CachedSQLCRepository struct {
    base     domain.EntityRepository
    store    db.Store
    cache    redis.Client
    ttl      time.Duration
}

func NewCachedSQLCRepository(base domain.EntityRepository, store db.Store, cache redis.Client) domain.EntityRepository {
    return &CachedSQLCRepository{
        base:  base,
        store: store,
        cache: cache,
        ttl:   time.Hour, // Default TTL
    }
}

// GetByID with caching layer over SQLC
func (r *CachedSQLCRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Entity, error) {
    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        return nil, fmt.Errorf("tenant ID not found in context")
    }

    cacheKey := fmt.Sprintf("entity:%s:%s", tenantID, id)
    
    // Try cache first
    cached, err := r.cache.Get(ctx, cacheKey).Result()
    if err == nil {
        var entity domain.Entity
        if err := json.Unmarshal([]byte(cached), &entity); err == nil {
            return &entity, nil
        }
    }

    // Cache miss - use SQLC repository
    entity, err := r.base.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Cache the result
    entityBytes, _ := json.Marshal(entity)
    r.cache.Set(ctx, cacheKey, entityBytes, r.ttl)

    return entity, nil
}

// GetSubtree with intelligent caching
func (r *CachedSQLCRepository) GetSubtree(ctx context.Context, rootID uuid.UUID, maxDepth int) ([]*domain.EntityWithHierarchy, error) {
    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        return nil, fmt.Errorf("tenant ID not found in context")
    }

    cacheKey := fmt.Sprintf("subtree:%s:%s:%d", tenantID, rootID, maxDepth)
    
    // Check cache
    cached, err := r.cache.Get(ctx, cacheKey).Result()
    if err == nil {
        var entities []*domain.EntityWithHierarchy
        if err := json.Unmarshal([]byte(cached), &entities); err == nil {
            return entities, nil
        }
    }

    // Use optimized SQLC query
    entities, err := r.getSubtreeOptimized(ctx, rootID, maxDepth)
    if err != nil {
        return nil, err
    }

    // Cache with shorter TTL for dynamic data
    entitiesBytes, _ := json.Marshal(entities)
    r.cache.Set(ctx, cacheKey, entitiesBytes, 15*time.Minute)

    return entities, nil
}

// Optimized subtree query using SQLC
func (r *CachedSQLCRepository) getSubtreeOptimized(ctx context.Context, rootID uuid.UUID, maxDepth int) ([]*domain.EntityWithHierarchy, error) {
    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        return nil, fmt.Errorf("tenant ID not found in context")
    }

    return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) ([]*domain.EntityWithHierarchy, error) {
        // Use SQLC optimized query with pagination
        rows, err := s.GetEntitySubtreeOptimized(ctx, db.GetEntitySubtreeOptimizedParams{
            Uuid:    rootID,
            Int4:    int32(maxDepth),
            Boolean: true,  // include inactive
            Int4_2:  1000,  // pagination limit
        })
        if err != nil {
            return nil, fmt.Errorf("SQLC optimized subtree query failed: %w", err)
        }

        // Convert to domain objects
        entities := make([]*domain.EntityWithHierarchy, len(rows))
        for i, row := range rows {
            entities[i] = &domain.EntityWithHierarchy{
                Entity: domain.Entity{
                    ID:       row.Uuid,
                    Name:     row.Name,
                    Type:     domain.EntityType(row.Type),
                    ParentID: convertPgUUID(row.ParentID),
                    IsActive: row.IsActive,
                },
                Level: int(row.Depth),
                Path:  row.PathMaterialized,
            }
        }

        return entities, nil
    })
}

// Cache invalidation on updates
func (r *CachedSQLCRepository) UpdateEntity(ctx context.Context, id uuid.UUID, req *domain.UpdateEntityRequest) (*domain.Entity, error) {
    // Perform update using base repository
    entity, err := r.base.UpdateEntity(ctx, id, req)
    if err != nil {
        return nil, err
    }

    // Invalidate caches
    tenantID, _ := shared.GetTenantID(ctx)
    r.invalidateEntityCaches(ctx, tenantID, id)

    return entity, nil
}

// Smart cache invalidation using SQLC hierarchy queries
func (r *CachedSQLCRepository) invalidateEntityCaches(ctx context.Context, tenantID, entityID uuid.UUID) {
    // Invalidate entity cache
    entityKey := fmt.Sprintf("entity:%s:%s", tenantID, entityID)
    r.cache.Del(ctx, entityKey)

    // Get affected ancestors and descendants using SQLC
    r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) (interface{}, error) {
        // Get ancestors to invalidate subtree caches
        ancestors, err := s.GetEntityAncestorsOptimized(ctx, entityID)
        if err == nil {
            for _, ancestor := range ancestors {
                pattern := fmt.Sprintf("subtree:%s:%s:*", tenantID, ancestor.Uuid)
                r.invalidateCachePattern(ctx, pattern)
            }
        }

        // Get descendants to invalidate their entity caches
        descendants, err := s.GetEntitySubtreeOptimized(ctx, db.GetEntitySubtreeOptimizedParams{
            Uuid:    entityID,
            Int4:    0,      // unlimited depth
            Boolean: true,   // include inactive
            Int4_2:  10000,  // large limit
        })
        if err == nil {
            for _, desc := range descendants {
                descKey := fmt.Sprintf("entity:%s:%s", tenantID, desc.Uuid)
                r.cache.Del(ctx, descKey)
            }
        }

        return nil, nil
    })
}

func (r *CachedSQLCRepository) invalidateCachePattern(ctx context.Context, pattern string) {
    keys, err := r.cache.Keys(ctx, pattern).Result()
    if err == nil && len(keys) > 0 {
        r.cache.Del(ctx, keys...)
    }
}
```

## Production Readiness Checklist

### **SQLC Implementation Standards**

✅ **Query Definitions**
- All database operations in `.sql` files
- SQLC directives properly configured
- Type-safe parameter binding
- Proper error handling patterns

✅ **Repository Layer**
- Pure SQLC implementation
- Transaction management
- Bulk operation support
- Error mapping and domain conversion

✅ **Temporal Integration** 
- SQLC operations in activities
- Transaction boundaries in workflows
- Progress tracking and heartbeats
- Retry policies for database failures

✅ **Testing Strategy**
- Mocked SQLC store for unit tests
- Integration tests with real database
- Hierarchy operation validation
- Performance testing for bulk operations

✅ **Performance Optimization**
- Query optimization patterns
- Intelligent caching strategies
- Bulk operation efficiency
- Index-aware query design

This guide provides production-ready patterns that maintain SQLC exclusivity while leveraging existing Clean Architecture patterns and Temporal integration capabilities.

<function_calls>
<invoke name="TodoWrite">
<parameter name="todos">[{"content": "Create SQLC-centric implementation patterns guide", "status": "completed", "priority": "high", "id": "1"}, {"content": "Provide SQLC query design examples for hierarchy operations", "status": "completed", "priority": "high", "id": "2"}, {"content": "Demonstrate repository layer implementation with SQLC", "status": "completed", "priority": "high", "id": "3"}, {"content": "Show Temporal activities with SQLC integration", "status": "in_progress", "priority": "medium", "id": "4"}, {"content": "Provide testing strategy for SQLC code", "status": "pending", "priority": "medium", "id": "5"}, {"content": "Include performance optimization with SQLC", "status": "pending", "priority": "medium", "id": "6"}]
```

<function_calls>
<invoke name="TodoWrite">
<parameter name="todos">[{"content": "Create SQLC-centric implementation patterns guide", "status": "completed", "priority": "high", "id": "1"}, {"content": "Provide SQLC query design examples for hierarchy operations", "status": "completed", "priority": "high", "id": "2"}, {"content": "Demonstrate repository layer implementation with SQLC", "status": "in_progress", "priority": "high", "id": "3"}, {"content": "Show Temporal activities with SQLC integration", "status": "pending", "priority": "medium", "id": "4"}, {"content": "Provide testing strategy for SQLC code", "status": "pending", "priority": "medium", "id": "5"}, {"content": "Include performance optimization with SQLC", "status": "pending", "priority": "medium", "id": "6"}]