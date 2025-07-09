
please impiliment this feature repo methods following [dataflow pattern](docs/data-flow-pattern.md) 

// ==============================================
// CREATE ENTITY
// ==============================================
function createEntity(tenantId, entityData) {
    // Validate unique name and code
    if validateEntityName(entityData.name) exists 
        throw "Entity name already exists"
    if validateEntityCode(entityData.code) exists 
        throw "Entity code already exists"
    
    // Validate parent if provided
    if entityData.parentId {
        if !validateEntityParent(entityData.parentId) 
            throw "Invalid parent entity"
        if checkCircularReference(entityData.parentId, newUuid)
            throw "Circular reference detected"
    }
    
    // Generate UUID and create entity
    newUuid = generateUUID()
    entity = execute CreateEntity(
        uuid: newUuid,
        parent_id: entityData.parentId,
        name: entityData.name,
        code: entityData.code,
        type: entityData.type,
        is_active: true,
        ...otherFields
    )
    
    // Build hierarchy paths
    if entityData.parentId {
        execute CreateHierarchyPath(
            ancestor_id: entityData.parentId,
            descendant_id: newUuid,
            depth: 1
        )
        execute UpdateHierarchyPaths(descendant_id: newUuid)
    }
    
    return entity
}

// ==============================================
// READ ENTITY
// ==============================================
function getEntity(tenantId, identifier) {
    if isUUID(identifier) {
        return execute GetEntity(uuid: identifier)
    } 
    else if isCode(identifier) {
        return execute GetEntityByCode(code: identifier)
    }
    else {
        return execute GetEntityByName(name: identifier)
    }
}

// Get entity with hierarchy info
function getEntityWithHierarchy(tenantId, uuid) {
    return execute GetEntityWithHierarchyInfo(
        entityId: uuid, 
        tenantId: tenantId
    )
}

// List entities with filtering
function listEntities(tenantId, filters) {
    return execute ListEntitiesWithPagination(
        type: filters.type,
        is_active: filters.active,
        hidden: filters.hidden,
        limit: filters.limit,
        offset: filters.offset
    )
}

// ==============================================
// UPDATE ENTITY
// ==============================================
function updateEntity(tenantId, uuid, updateData) {
    // Validate unique constraints
    if updateData.name && validateEntityName(updateData.name, uuid) 
        throw "New entity name already exists"
    if updateData.code && validateEntityCode(updateData.code, uuid)
        throw "New entity code already exists"
    
    // Handle parent changes
    if updateData.parentId {
        oldParent = execute GetEntityParent(uuid)
        if oldParent.uuid !== updateData.parentId {
            if !validateEntityParent(updateData.parentId)
                throw "Invalid parent entity"
            if checkCircularReference(updateData.parentId, uuid)
                throw "Circular reference detected"
            
            execute MoveEntityToNewParent(
                entity_id: uuid,
                new_parent_id: updateData.parentId
            )
        }
    }
    
    // Perform update
    return execute UpdateEntity(
        uuid: uuid,
        name: updateData.name,
        code: updateData.code,
        ...otherFields
    )
}

// ==============================================
// DELETE ENTITY
// ==============================================
function deleteEntity(tenantId, uuid, permanent = false) {
    if permanent {
        // Hard delete
        execute DeleteHierarchyPaths(uuid)
        execute DeleteEntityStatesForEntity(uuid)
        execute HardDeleteEntity(uuid)
    } else {
        // Soft delete
        execute SoftDeleteEntity(uuid)
        // Optional: Cleanup orphaned hierarchy paths
        execute CleanupOrphanedHierarchyPaths()
    }
}

// Restore soft-deleted entity
function restoreEntity(tenantId, uuid) {
    execute RestoreEntity(uuid)
    execute RebuildHierarchyPaths(tenantId)
}

// ==============================================
// ENTITY STATE MANAGEMENT
// ==============================================
function getNextSequence(tenantId, entityId, key, fiscalYear) {
    return execute GetNextSequenceNumber(
        entity_id: entityId,
        key: key,
        fiscal_year: fiscalYear
    )
}

function resetEntitySequence(entityId, key, fiscalYear) {
    execute ResetEntityStateSequence(
        entity_id: entityId,
        key: key,
        sequence: 1,
        fiscal_year: fiscalYear
    )
}

// ==============================================
// HIERARCHY OPERATIONS
// ==============================================
function getEntityChildren(tenantId, uuid) {
    return execute GetEntityChildren(ancestor_id: uuid)
}

function getEntityAncestors(tenantId, uuid) {
    return execute GetEntityAncestors(descendant_id: uuid)
}

function getEntityTree(tenantId) {
    return execute GetEntityTreeStructure(tenant_id: tenantId)
}

// ==============================================
// VALIDATION FUNCTIONS
// ==============================================
function validateEntityName(name, excludeUuid = null) {
    return execute ValidateEntityName(name, excludeUuid)
}

function validateEntityCode(code, excludeUuid = null) {
    return execute ValidateEntityCode(code, excludeUuid)
}

function validateEntityParent(parentId, childId) {
    return execute ValidateEntityParent(parentId, childId)
}

function checkCircularReference(parentId, childId) {
    return execute IsEntityAncestor(
        ancestor_id: childId,
        descendant_id: parentId
    )
}
