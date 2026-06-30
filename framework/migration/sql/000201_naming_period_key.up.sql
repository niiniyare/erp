-- Add period_key to awo_naming_sequences to support date-based counter resets.
--
-- period_key encodes the reset boundary:
--   ""        → ResetNever  (monotonically increasing, existing behaviour)
--   "2025"    → ResetYearly
--   "2025-06" → ResetMonthly
--
-- The existing unique constraint (tenant_id, entity) covered only one sequence
-- per entity per tenant. The new constraint includes period_key so yearly and
-- monthly sequences are independent rows.
--
-- Migration is non-destructive: existing rows receive period_key = '' which
-- preserves their current counter values and maps to ResetNever behaviour.

-- Step 1: add column with default '' so existing rows are updated atomically.
ALTER TABLE awo_naming_sequences
    ADD COLUMN IF NOT EXISTS period_key text NOT NULL DEFAULT '';

-- Step 2: drop the old unique constraint (tenant_id, entity).
ALTER TABLE awo_naming_sequences
    DROP CONSTRAINT IF EXISTS awo_naming_sequences_tenant_id_entity_key;

-- Step 3: add new unique constraint that includes period_key.
ALTER TABLE awo_naming_sequences
    ADD CONSTRAINT awo_naming_sequences_tenant_entity_period_key
    UNIQUE (tenant_id, entity, period_key);
