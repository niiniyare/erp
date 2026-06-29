package migration

import "fmt"

// FrameworkTables returns the up+down migration Files for all internal
// framework tables that are not tied to a specific EntityDefinition.
//
// Call this before GenerateAll so framework tables are created first.
// Recommended sequence numbers: 000001–000099 (reserved for framework).
//
//	files := migration.FrameworkTables(1)
//	files = append(files, migration.GenerateAll(100)...)
func FrameworkTables(seq int) []File {
	up := File{
		Name:    pad6(seq) + "_framework_tables.up.sql",
		Content: frameworkUp,
	}
	down := File{
		Name:    pad6(seq) + "_framework_tables.down.sql",
		Content: frameworkDown,
	}
	return []File{up, down}
}

func pad6(n int) string {
	return fmt.Sprintf("%06d", n)
}

const frameworkUp = `-- -----------------------------------------------------------------------
-- Framework internal tables (awo.so/framework)
-- Generated automatically — do not edit by hand.
-- -----------------------------------------------------------------------

-- ── Naming sequences ────────────────────────────────────────────────────────
-- Stores the current counter for each entity's document numbering series.
-- The atomic upsert in naming.Next guarantees no gaps under concurrent load.
CREATE TABLE IF NOT EXISTS naming_sequences (
    tenant_id   UUID   NOT NULL,
    entity      TEXT   NOT NULL,
    current_seq BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (tenant_id, entity)
);

GRANT SELECT, INSERT, UPDATE ON naming_sequences TO application_role, admin_role;

-- ── Custom field definitions ─────────────────────────────────────────────────
-- Stores per-tenant, per-entity custom field definitions as a JSONB array of
-- definition.FieldDef objects. The customfield.Registry caches these in-memory
-- and must be invalidated on change.
CREATE TABLE IF NOT EXISTS custom_fields (
    tenant_id  UUID        NOT NULL,
    entity     TEXT        NOT NULL,
    fields     JSONB       NOT NULL DEFAULT '[]',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, entity)
);

GRANT SELECT, INSERT, UPDATE ON custom_fields TO application_role, admin_role;

-- ── Audit log ────────────────────────────────────────────────────────────────
-- Immutable audit trail for Create / Update / Delete on Audited entities.
-- Rows are never updated or deleted — append-only.
CREATE TABLE IF NOT EXISTS audit_log (
    id         UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id  UUID,
    entity     TEXT        NOT NULL,
    record_id  UUID        NOT NULL,
    op         TEXT        NOT NULL CHECK (op IN ('create','update','delete')),
    actor_id   TEXT        NOT NULL,
    changes    JSONB       NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS audit_log_record_idx
    ON audit_log (entity, record_id);
CREATE INDEX IF NOT EXISTS audit_log_tenant_idx
    ON audit_log (tenant_id, created_at DESC);

-- Audit log is INSERT-only for the application role.
GRANT INSERT ON audit_log TO application_role;
GRANT SELECT ON audit_log TO application_role, admin_role, readonly_role;
GRANT INSERT, SELECT ON audit_log TO admin_role;
`

const frameworkDown = `-- Drop framework tables in reverse dependency order.
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS custom_fields;
DROP TABLE IF EXISTS naming_sequences;
`
