-- =====================================================================
--  ENTITY STATE MANAGEMENT QUERIES
--  Complete query collection for document sequence tracking
-- =====================================================================

-- =====================================================================
--  EXISTING QUERIES (DO NOT MODIFY)
-- =====================================================================

-- name: GetEntityStateWithLocking :one
-- Usage: Retrieves entity state with row-level locking for atomic sequence operations
-- Use case: When you need to get and immediately update a sequence number safely
-- NOTE: Handles query timeout to prevent deadlocks
DO $$
BEGIN
  SET LOCAL lock_timeout = '3s'; -- adjust as appropriate (e.g. '500ms', '5s')

  -- Actual SELECT with FOR UPDATE locking
  PERFORM * FROM entitystate
  WHERE entity_id = $1 AND key = $2 AND fiscal_year = $3 AND tenant_id = current_tenant_id()
  FOR UPDATE;
END $$;

-- -- name: GetEntityStateWithLocking :one
-- -- Usage: Retrieves entity state with row-level locking for atomic sequence operations
-- -- Use case: When you need to get and immediately update a sequence number safely
-- -- NOTE: Consider adding query timeout handling for deadlock scenarios
-- SELECT * FROM entitystate
-- WHERE entity_id = $1 AND key = $2 AND fiscal_year = $3 AND tenant_id = current_tenant_id()
-- FOR UPDATE;

-- name: BulkCreateEntityStates :exec
-- Usage: Creates multiple entity states in batch for different document types
-- Use case: Initial setup of document sequences for a new entity
-- NOTE: Missing tenant_id assignment and UUID generation - should be addressed
INSERT INTO entitystate (entity_id, key, sequence_number, fiscal_year, tenant_id)
SELECT $1, unnest($2::VARCHAR[]), 1, $3, current_tenant_id()
ON CONFLICT (entity_id, key, fiscal_year) DO NOTHING;

-- name: GetEntityStateHistory :many
-- Usage: Retrieves entity state history with optional filtering by key and fiscal year
-- Use case: Audit trails, reporting, and historical sequence analysis
-- NOTE: Consider adding pagination (LIMIT/OFFSET) for large datasets
SELECT es.*, e.name as entity_name
FROM entitystate es
JOIN entities e ON es.entity_id = e.uuid
WHERE es.entity_id = $1 AND e.tenant_id = current_tenant_id()
AND ($2::VARCHAR IS NULL OR es.key = $2)
AND ($3::SMALLINT IS NULL OR es.fiscal_year = $3)
ORDER BY es.fiscal_year DESC, es.key;

-- name: GetHighestSequenceNumber :one
-- Usage: Gets the highest sequence number for a specific entity/key/fiscal year combination
-- Use case: Finding the current maximum sequence before manual adjustments
SELECT COALESCE(MAX(sequence), 0) AS sequence
FROM entitystate
WHERE entity_id = $1 AND key = $2 AND fiscal_year = $3 AND tenant_id = current_tenant_id();

-- name: ResetAllEntitySequences :exec
-- Usage: Resets all document sequences to 1 for an entity's fiscal year
-- Use case: New fiscal year initialization or sequence resets
UPDATE entitystate
SET sequence = 1
WHERE entity_id = $1 AND fiscal_year = $2 AND tenant_id = current_tenant_id();

-- name: GetEntityStatesByFiscalYear :many
-- Usage: Retrieves all entity states for a specific fiscal year with optional entity unit filtering
-- Use case: Cross-entity reporting, fiscal year analysis, bulk operations
-- Parameters: $1=fiscal_year, $2=entity_unit_id (nullable for filtering by specific unit)
SELECT es.*, e.name as entity_name, eu.name as unit_name
FROM entitystate es
JOIN entities e ON es.entity_id = e.uuid
LEFT JOIN entities eu ON es.entity_unit_id = eu.uuid
WHERE es.fiscal_year = $1
AND e.tenant_id = current_tenant_id()
AND es.tenant_id = current_tenant_id()
AND ($2::UUID IS NULL OR es.entity_unit_id = $2)
ORDER BY e.name, COALESCE(eu.name, ''), es.key;

-- =====================================================================
--  NEW SEQUENCE MANAGEMENT QUERIES
-- =====================================================================

-- name: UpdateEntitySequence :one
-- Usage: Atomically increments sequence number and returns the new value
-- Use case: Getting next sequence number for document creation (most common operation)
UPDATE entitystate 
SET sequence = sequence + 1, updated_at = NOW()
WHERE entity_id = $1 AND key = $2 AND fiscal_year = $3 AND tenant_id = current_tenant_id()
RETURNING sequence;

-- name: GetOrCreateEntityState :one
-- Usage: Gets existing entity state or creates new one with sequence = 1
-- Use case: Lazy initialization of sequences when first document is created
INSERT INTO entitystate (entity_id, key, fiscal_year, sequence, tenant_id)
VALUES ($1, $2, $3, 1, current_tenant_id())
ON CONFLICT (tenant_id, entity_id, key, fiscal_year) 
DO UPDATE SET updated_at = NOW()
RETURNING *;

-- name: SetEntitySequence :exec
-- Usage: Manually sets a specific sequence number (with validation)
-- Use case: Data migration, manual sequence adjustments, importing from other systems
UPDATE entitystate 
SET sequence = $4, updated_at = NOW()
WHERE entity_id = $1 AND key = $2 AND fiscal_year = $3 AND tenant_id = current_tenant_id()
AND $4 > 0; -- Ensures positive sequence number

-- =====================================================================
--  BULK OPERATIONS QUERIES
-- =====================================================================

-- name: GetMultipleEntityStates :many
-- Usage: Retrieves multiple entity states by document types in one query
-- Use case: Dashboard displays, bulk document creation, batch processing
SELECT * FROM entitystate
WHERE entity_id = $1 
AND key = ANY($2::VARCHAR[])
AND fiscal_year = $3 
AND tenant_id = current_tenant_id()
ORDER BY key;

-- name: BulkUpdateSequences :exec
-- Usage: Updates multiple sequences in a single transaction
-- Use case: Batch sequence adjustments, data synchronization, bulk imports
UPDATE entitystate 
SET sequence = data.new_sequence, updated_at = NOW()
FROM (
    SELECT 
        unnest($1::UUID[]) as entity_id, 
        unnest($2::VARCHAR[]) as key, 
        unnest($3::BIGINT[]) as new_sequence
) as data
WHERE entitystate.entity_id = data.entity_id 
AND entitystate.key = data.key 
AND entitystate.fiscal_year = $4
AND entitystate.tenant_id = current_tenant_id();

-- name: BulkCreateEntityStatesFixed :exec
-- Usage: Improved bulk creation with proper tenant_id and UUID handling
-- Use case: Initial entity setup, adding new document types to existing entities
INSERT INTO entitystate (uuid, tenant_id, entity_id, key, fiscal_year, sequence, entity_unit_id)
SELECT 
    gen_random_uuid(),
    current_tenant_id(),
    $1,
    unnest($2::VARCHAR[]),
    $3,
    1,
    $4
ON CONFLICT (tenant_id, entity_id, key, fiscal_year) DO NOTHING;

-- =====================================================================
--  ANALYTICS & REPORTING QUERIES
-- =====================================================================

-- name: GetEntityStateStats :many
-- Usage: Provides statistical overview of sequence usage by document type
-- Use case: Usage analytics, capacity planning, identifying heavily used document types
SELECT 
    key,
    COUNT(*) as entity_count,
    AVG(sequence)::BIGINT as avg_sequence,
    MAX(sequence) as max_sequence,
    MIN(sequence) as min_sequence,
    SUM(sequence) as total_sequences_used
FROM entitystate 
WHERE fiscal_year = $1 AND tenant_id = current_tenant_id()
GROUP BY key
ORDER BY total_sequences_used DESC;

-- name: GetSequenceGaps :many
-- Usage: Identifies missing sequence numbers (gaps in numbering)
-- Use case: Audit compliance, finding deleted/voided documents, sequence integrity checks
WITH sequence_range AS (
    SELECT generate_series(1, (
        SELECT MAX(es1.sequence) FROM entitystate es1
        WHERE es1.entity_id = $1 AND es1.key = $2 AND es1.fiscal_year = $3 AND es1.tenant_id = current_tenant_id()
    )) as seq_num
)
SELECT seq_num as missing_sequence
FROM sequence_range
WHERE seq_num NOT IN (
    SELECT es2.sequence FROM entitystate es2
    WHERE es2.entity_id = $1 AND es2.key = $2 AND es2.fiscal_year = $3 AND es2.tenant_id = current_tenant_id()
);

-- name: GetEntitySequenceSummary :many
-- Usage: Summary view of all sequences for an entity across fiscal years
-- Use case: Entity overview dashboards, year-over-year comparisons
SELECT 
    fiscal_year,
    key,
    sequence,
    updated_at,
    (sequence - 1) as documents_created
FROM entitystate
WHERE entity_id = $1 AND tenant_id = current_tenant_id()
ORDER BY fiscal_year DESC, key;

-- =====================================================================
--  ENTITY UNIT (SUB-ENTITY) QUERIES
-- =====================================================================

-- name: GetEntityUnitStates :many
-- Usage: Retrieves all sequences for a specific entity unit/department
-- Use case: Department-level reporting, unit-specific sequence management
SELECT es.*, e1.name as entity_name, e2.name as unit_name
FROM entitystate es
JOIN entities e1 ON es.entity_id = e1.uuid
LEFT JOIN entities e2 ON es.entity_unit_id = e2.uuid
WHERE es.entity_unit_id = $1 AND es.tenant_id = current_tenant_id()
ORDER BY es.key, es.fiscal_year;

-- name: GetEntityWithUnitStates :many
-- Usage: Hierarchical view of entity and all its unit sequences
-- Use case: Complete entity structure analysis, hierarchical reporting
SELECT 
    es.*,
    e1.name as entity_name,
    e2.name as unit_name,
    CASE WHEN es.entity_unit_id IS NULL THEN 'Main Entity' ELSE 'Unit' END as level_type
FROM entitystate es
JOIN entities e1 ON es.entity_id = e1.uuid
LEFT JOIN entities e2 ON es.entity_unit_id = e2.uuid
WHERE es.entity_id = $1 AND es.tenant_id = current_tenant_id()
ORDER BY level_type, COALESCE(e2.name, e1.name), es.key, es.fiscal_year;

-- name: GetEntityUnitSequence :one
-- Usage: Gets specific sequence for an entity unit with locking
-- Use case: Unit-specific sequence generation with concurrency safety
SELECT * FROM entitystate
WHERE entity_id = $1 AND entity_unit_id = $2 AND key = $3 AND fiscal_year = $4 
AND tenant_id = current_tenant_id()
FOR UPDATE;

-- =====================================================================
--  MAINTENANCE & CLEANUP QUERIES
-- =====================================================================

-- name: DeleteUnusedEntityStates :exec
-- Usage: Removes entity states for old fiscal years or inactive entities
-- Use case: Data retention policy enforcement, database cleanup
DELETE FROM entitystate
WHERE entity_id = $1 
AND fiscal_year < $2 
AND tenant_id = current_tenant_id();

-- name: GetStaleEntityStates :many
-- Usage: Finds entity states that haven't been updated recently
-- Use case: Identifying inactive sequences, cleanup candidate identification
SELECT es.*, e.name as entity_name,
       NOW() - es.updated_at as time_since_update
FROM entitystate es
JOIN entities e ON es.entity_id = e.uuid
WHERE es.updated_at < $1 
AND es.tenant_id = current_tenant_id()
ORDER BY es.updated_at ASC;

-- name: ArchiveOldEntityStates :exec
-- Usage: Soft delete or archive old entity states
-- Use case: Long-term data archival while maintaining referential integrity
UPDATE entitystate 
SET updated_at = NOW()
WHERE fiscal_year < $1 
AND tenant_id = current_tenant_id()
AND updated_at < (NOW() - INTERVAL '$2 months');

-- =====================================================================
--  VALIDATION & HEALTH CHECK QUERIES
-- =====================================================================

-- name: ValidateSequenceIntegrity :many
-- Usage: Checks for sequence numbering issues across all entities
-- Use case: Data integrity audits, troubleshooting sequence problems
SELECT 
    entity_id,
    key,
    fiscal_year,
    sequence,
    CASE 
        WHEN sequence <= 0 THEN 'Invalid: Non-positive sequence'
        WHEN sequence = 1 THEN 'OK: Initial state'
        ELSE 'OK: In use'
    END as status
FROM entitystate
WHERE tenant_id = current_tenant_id()
AND (sequence <= 0 OR sequence IS NULL)
ORDER BY entity_id, key, fiscal_year;

-- name: GetDuplicateSequenceCheck :many
-- Usage: Identifies potential duplicate sequence configurations
-- Use case: Data integrity verification, migration validation
SELECT 
    tenant_id, entity_id, key, fiscal_year, COUNT(*)
FROM entitystate
WHERE tenant_id = current_tenant_id()
GROUP BY tenant_id, entity_id, key, fiscal_year
HAVING COUNT(*) > 1;

-- name: GetEntityStateHealthCheck :many
-- Usage: Comprehensive health check of entity state configuration
-- Use case: System health monitoring, pre-deployment validation
SELECT 
    es.entity_id,
    e.name as entity_name,
    COUNT(DISTINCT es.key) as document_types_count,
    COUNT(DISTINCT es.fiscal_year) as fiscal_years_count,
    MIN(es.created_at) as oldest_sequence,
    MAX(es.updated_at) as last_activity,
    SUM(es.sequence - 1) as total_documents_created
FROM entitystate es
JOIN entities e ON es.entity_id = e.uuid
WHERE es.tenant_id = current_tenant_id()
GROUP BY es.entity_id, e.name
ORDER BY total_documents_created DESC;
