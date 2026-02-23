CREATE TABLE casbin_rule (
  id    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  ptype VARCHAR(10) NOT NULL,
  v0    VARCHAR(256) NOT NULL DEFAULT '',
  v1    VARCHAR(256) NOT NULL DEFAULT '',
  v2    VARCHAR(256) NOT NULL DEFAULT '',
  v3    VARCHAR(256) NOT NULL DEFAULT '',
  v4    VARCHAR(256) NOT NULL DEFAULT '',
  v5    VARCHAR(256) NOT NULL DEFAULT ''
);

-- Deduplicate: Casbin assumes unique rows
CREATE UNIQUE INDEX idx_casbin_rule_unique
  ON casbin_rule(ptype, v0, v1, v2, v3, v4, v5);
CREATE INDEX idx_casbin_rule_ptype ON casbin_rule(ptype);

-- No tenant-level RLS: domain value in v1 enforces isolation at app level
ALTER TABLE casbin_rule ENABLE ROW LEVEL SECURITY;
CREATE POLICY casbin_rule_app   ON casbin_rule FOR ALL TO application_role USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY casbin_rule_admin ON casbin_rule FOR ALL TO admin_role        USING (TRUE) WITH CHECK (TRUE);
GRANT SELECT, INSERT, UPDATE, DELETE ON casbin_rule TO application_role;
GRANT ALL ON casbin_rule TO admin_role;
