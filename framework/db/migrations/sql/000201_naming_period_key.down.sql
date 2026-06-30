-- Revert period_key addition to awo_naming_sequences.
-- WARNING: this drops all period-keyed sequence rows (yearly/monthly counters).
-- Non-keyed rows (period_key = '') are preserved but their unique constraint changes.

ALTER TABLE awo_naming_sequences
    DROP CONSTRAINT IF EXISTS awo_naming_sequences_tenant_entity_period_key;

ALTER TABLE awo_naming_sequences
    ADD CONSTRAINT awo_naming_sequences_tenant_id_entity_key
    UNIQUE (tenant_id, entity);

ALTER TABLE awo_naming_sequences
    DROP COLUMN IF EXISTS period_key;
