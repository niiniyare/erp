-- =============================================================================
-- MIGRATION 008 UP: Configuration Definitions
-- =============================================================================
-- Architecture Decision (ADR-024): config_definitions ownership boundary
--   The Settings module's ConfigurationService owns this table exclusively.
--   No other module should INSERT, UPDATE, or DELETE from it directly.
--   Application code and other services interact with configuration values
--   through the ConfigurationService interface (GetEffectiveConfiguration,
--   ListEffectiveConfigurations) — never via direct SQL against this table.
--
--   This table answers the question: "what keys are valid for a given module?"
--   It is a schema registry for the 3-level inheritance system
--   (System → Tenant → Entity) described in the PRD.
--
-- Architecture Decision (ADR-025): Why this is NOT on the tenants table
--   config_definitions describes what configuration keys EXIST across the
--   whole platform. It is module-scoped, not tenant-scoped. Putting these
--   rows on the tenants table would mean duplicating them once per tenant,
--   which would make schema evolution (adding a new config key) require
--   inserting N rows instead of 1.
--
-- Architecture Decision (ADR-026): FeatureFlag integration column
--   required_feature_flag stores the feature flag key that must be enabled
--   for this configuration to be available. The FeatureFlag service is
--   authoritative for flag evaluation — this column is just a metadata hint
--   so the Settings UI can show/hide configurations correctly and so the
--   ConfigurationService can skip resolution for gated configs without
--   calling the FeatureFlag service on every read.
--   NULL = always available, no feature flag required.
--
-- Architecture Decision (ADR-027): required_iam_permission column
--   Stores the IAM permission key required to read or write this config.
--   The IAM service is authoritative for permission evaluation. This column
--   is metadata for the Settings UI and for the ConfigurationService to
--   return a clear "insufficient permissions" error before attempting a write.
--   NULL = any authenticated user can read/write (internal defaults).
--
-- Architecture Decision (ADR-028): validation_rules JSONB
--   Validation rules are stored as JSONB to allow the ConfigurationService
--   to apply typed validation without a schema migration when new rule types
--   are added. The ConfigurationService is the single point of validation —
--   the database stores the rules, not the logic.
--
--   Supported rule schema (extensible):
--   {
--     "min": 0,                    // numeric minimum
--     "max": 1000000,              // numeric maximum
--     "pattern": "^[A-Z]{2,5}-$", // regex for string values
--     "enum": ["FIFO","LIFO"],     // allowed discrete values
--     "required": true             // whether the value may be null/empty
--   }
-- =============================================================================

CREATE TABLE IF NOT EXISTS config_definitions (
  id  UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,

  -- =========================================================================
  -- KEY IDENTITY
  -- =========================================================================

  module_id UUID NOT NULL references module(id),
  -- Module that owns this configuration key (e.g. 'finance', 'hr', 'inventory').
  -- Must match the module_name used in ConfigurationService calls.
  -- TODO: module is stored at module table to
  -- fetch the name or any info related it can be desided in the future how queried
  module_name  VARCHAR(50) NOT NULL
               CHECK (module_name ~* '^[a-z][a-z0-9_]*$'),

  -- The configuration key within the module (e.g. 'approval_limit', 'invoice_prefix').
  -- Together with module_name, forms the globally unique key.
  config_key   VARCHAR(100) NOT NULL
               CHECK (config_key ~* '^[a-z][a-z0-9_]*$'),

  -- Globally unique composite key: a module cannot have duplicate key names.
  CONSTRAINT config_definitions_module_key_key UNIQUE (module_name, config_key),

  -- =========================================================================
  -- TYPE AND DEFAULT
  -- =========================================================================

  -- Data type of the configuration value.
  -- The ConfigurationService uses this to apply correct type casting and
  -- to populate the correct AsString() / AsDecimal() / AsBool() / AsInt()
  -- method on the ConfigValue value object.
config_type VARCHAR(20) NOT NULL
      CHECK (
    config_type IN ('STRING', 'INTEGER', 'DECIMAL', 'BOOLEAN', 'JSON')),

  -- System-level default value (Level 1 in the 3-level hierarchy).
  -- Stored as JSONB so it can hold any scalar or structured value consistently.
  -- NULL means "no system default exists" — tenant or entity MUST configure it.
  default_value  JSONB,

  -- =========================================================================
  -- INHERITANCE CONTROL
  -- =========================================================================

  -- Whether tenant administrators are permitted to override this setting.
  -- FALSE = system default is locked; tenant sees the default and cannot change it.
  is_tenant_overridable  BOOLEAN NOT NULL DEFAULT TRUE,

  -- Whether entity managers are permitted to override this setting.
  -- FALSE = tenant-level value is locked; entity sees the tenant value and cannot change it.
  -- is_tenant_overridable=FALSE and is_entity_overridable=FALSE means the
  -- system default is authoritative for all levels.
  is_entity_overridable  BOOLEAN NOT NULL DEFAULT TRUE,

  -- =========================================================================
  -- VALIDATION
  -- =========================================================================

  -- JSON schema for value validation (see ADR-028).
  -- NULL = no additional validation beyond type checking.
  -- The ConfigurationService evaluates these rules — not the database.
  validation_rules  JSONB,

  -- =========================================================================
  -- ACCESS CONTROL (metadata hints — see ADR-026, ADR-027)
  -- =========================================================================

  -- Feature flag key from the FeatureFlag service.
  -- NULL = always available (no feature gate).
  -- Example: 'inventory.lot_tracking', 'finance.multi_currency'
  required_feature_flag  VARCHAR(100),

  -- IAM permission key from the IAM service.
  -- NULL = accessible to any authenticated user.
  -- Example: 'finance.admin', 'hr.payroll.write'
  required_iam_permission  VARCHAR(100),

  -- =========================================================================
  -- DOCUMENTATION
  -- =========================================================================

  -- Human-readable label shown in the Settings UI.
  display_name  VARCHAR(255) NOT NULL,

  -- Full description explaining what this setting controls and its impact.
  -- Shown as help text in the Settings UI.
  description  TEXT NOT NULL,

  -- Unit label for numeric values (optional). Shown next to the input field.
  -- Examples: 'days', 'USD', '%', 'hours'
  unit_label  VARCHAR(20),

  -- =========================================================================
  -- LIFECYCLE
  -- =========================================================================

  -- Whether this configuration definition is currently active.
  -- Inactive definitions are hidden from the UI but preserved for
  -- backward compatibility with existing stored values.
  is_active  BOOLEAN NOT NULL DEFAULT TRUE,

  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- -------------------------------------------------------------------------
-- INDEXES
-- -------------------------------------------------------------------------

-- Primary lookup pattern: module_name + config_key (covered by UNIQUE constraint).
-- Additional indexes for list operations.

CREATE INDEX IF NOT EXISTS idx_config_defs_module
  ON config_definitions(module_name)
  WHERE is_active = TRUE;

CREATE INDEX IF NOT EXISTS idx_config_defs_feature_flag
  ON config_definitions(required_feature_flag)
  WHERE required_feature_flag IS NOT NULL AND is_active = TRUE;

-- -------------------------------------------------------------------------
-- TRIGGER: updated_at maintenance
-- -------------------------------------------------------------------------
-- Reuse the generic trigger function from migration 007.
DROP TRIGGER IF EXISTS update_config_definitions_updated_at ON config_definitions;
CREATE TRIGGER update_config_definitions_updated_at
  BEFORE UPDATE ON config_definitions
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

-- -------------------------------------------------------------------------
-- PERMISSIONS
-- -------------------------------------------------------------------------
-- The Settings module (application_role) needs to read config_definitions
-- to resolve configuration hierarchies. Only admin_role may write.
GRANT SELECT                ON config_definitions TO application_role;
GRANT SELECT                ON config_definitions TO readonly_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON config_definitions TO admin_role;

-- -------------------------------------------------------------------------
-- COMMENTS
-- -------------------------------------------------------------------------
COMMENT ON TABLE config_definitions IS
  'Schema registry for all available configuration keys across modules. '
  'Owned exclusively by the Settings module — do not write from other modules directly. '
  'Defines what keys exist, their types, defaults, validation rules, and access control.';

COMMENT ON COLUMN config_definitions.module_name           IS 'Module that owns this key (e.g. finance, hr, inventory). Lowercase snake_case.';
COMMENT ON COLUMN config_definitions.config_key            IS 'Configuration key within the module. Lowercase snake_case. Unique within module.';
COMMENT ON COLUMN config_definitions.config_type           IS 'Value data type used by ConfigurationService for casting and validation.';
COMMENT ON COLUMN config_definitions.default_value         IS 'System-level default (Level 1 in hierarchy). NULL means tenant/entity must configure it explicitly.';
COMMENT ON COLUMN config_definitions.is_tenant_overridable IS 'Whether tenant admins may override the system default.';
COMMENT ON COLUMN config_definitions.is_entity_overridable IS 'Whether entity managers may override the tenant value.';
COMMENT ON COLUMN config_definitions.validation_rules      IS 'JSON validation rules evaluated by ConfigurationService. Supports min/max/pattern/enum/required.';
COMMENT ON COLUMN config_definitions.required_feature_flag IS 'FeatureFlag service key required for this config to be available. NULL = always available. ADR-026.';
COMMENT ON COLUMN config_definitions.required_iam_permission IS 'IAM service permission key required to read/write this config. NULL = any authenticated user. ADR-027.';
COMMENT ON COLUMN config_definitions.display_name          IS 'Human-readable label shown in the Settings UI.';
COMMENT ON COLUMN config_definitions.description           IS 'Full explanation of what this setting controls. Shown as help text.';
COMMENT ON COLUMN config_definitions.unit_label            IS 'Optional unit label for numeric values (days, %, USD). Shown next to input field.';
COMMENT ON COLUMN config_definitions.is_active             IS 'Inactive definitions are hidden from UI but preserved for backward compatibility with stored values.';

-- -------------------------------------------------------------------------
-- SEED DATA: Core module configuration definitions
-- -------------------------------------------------------------------------
-- These are the baseline definitions required at system startup.
-- Module teams should add their own definitions here or via separate seed
-- migrations as modules are built.
-- -------------------------------------------------------------------------
INSERT INTO config_definitions
  (module_name, config_key, config_type, default_value,
   is_tenant_overridable, is_entity_overridable,
   display_name, description, unit_label, validation_rules,
   required_iam_permission)
VALUES

  -- FINANCE MODULE
  ('finance', 'invoice_prefix',        'string',  '"INV-"',
   TRUE, TRUE,
   'Invoice Prefix', 'Prefix applied to all invoice document numbers.', NULL,
   '{"pattern": "^[A-Z0-9]{2,6}-$"}',
   'finance.admin'),

  ('finance', 'payment_terms_default', 'string',  '"NET30"',
   TRUE, TRUE,
   'Default Payment Terms', 'Default payment terms applied to new invoices when not specified.', NULL,
   '{"enum": ["NET15","NET30","NET45","NET60","DUE_ON_RECEIPT"]}',
   'finance.admin'),

  ('finance', 'auto_approval_limit',   'decimal', '1000',
   TRUE, TRUE,
   'Auto-Approval Limit', 'Transactions below this amount are automatically approved without manual review.', 'USD',
   '{"min": 0, "max": 1000000}',
   'finance.admin'),

  ('finance', 'require_po_for_expense', 'boolean', 'false',
   TRUE, TRUE,
   'Require PO for Expenses', 'When enabled, all expense reports must reference a Purchase Order.', NULL,
   NULL,
   'finance.admin'),

  ('finance', 'invoice_pad_length',    'integer', '6',
   TRUE, TRUE,
   'Invoice Number Padding', 'Number of digits in the invoice sequence (e.g. 6 produces INV-000001).', 'digits',
   '{"min": 4, "max": 10}',
   'finance.admin'),

  ('finance', 'invoice_reset_frequency', 'string', '"yearly"',
   TRUE, TRUE,
   'Invoice Sequence Reset', 'How often the invoice number sequence resets to 1.', NULL,
   '{"enum": ["never","yearly","monthly"]}',
   'finance.admin'),

  -- HR MODULE
  ('hr', 'overtime_threshold',         'decimal', '40',
   TRUE, TRUE,
   'Overtime Threshold', 'Hours worked per week beyond which overtime rates apply.', 'hours',
   '{"min": 0, "max": 168}',
   'hr.admin'),

  ('hr', 'overtime_multiplier',        'decimal', '1.5',
   TRUE, TRUE,
   'Overtime Pay Multiplier', 'Multiplier applied to the base hourly rate for overtime hours.', 'x',
   '{"min": 1.0, "max": 5.0}',
   'hr.admin'),

  ('hr', 'default_pay_frequency',      'string',  '"biweekly"',
   TRUE, TRUE,
   'Default Pay Frequency', 'Default payroll frequency applied to new employees.', NULL,
   '{"enum": ["weekly","biweekly","semimonthly","monthly"]}',
   'hr.admin'),

  ('hr', 'probation_period_days',      'integer', '90',
   TRUE, TRUE,
   'Probation Period', 'Default probation period in days for new employees.', 'days',
   '{"min": 0, "max": 365}',
   'hr.admin'),

  -- INVENTORY MODULE
  ('inventory', 'reorder_point_days',      'integer', '30',
   TRUE, TRUE,
   'Reorder Point (Days)', 'Trigger a reorder when current stock falls below this many days of supply.', 'days',
   '{"min": 1, "max": 365}',
   'inventory.admin'),

  ('inventory', 'default_valuation_method', 'string', '"FIFO"',
   TRUE, FALSE,
   'Default Valuation Method', 'Inventory cost flow assumption for new items. Entity-level override is not permitted — tenant policy is authoritative.', NULL,
   '{"enum": ["FIFO","LIFO","WEIGHTED_AVERAGE"]}',
   'inventory.admin'),

  ('inventory', 'allow_negative_stock',     'boolean', 'false',
   TRUE, TRUE,
   'Allow Negative Stock', 'When enabled, stock levels may go below zero (e.g. for back-order scenarios).', NULL,
   NULL,
   'inventory.admin'),

  ('inventory', 'require_lot_tracking',     'boolean', 'false',
   TRUE, TRUE,
   'Require Lot Tracking', 'When enabled, all inventory movements must reference a lot or batch number.', NULL,
   NULL,
   'inventory.admin')

ON CONFLICT (module_name, config_key) DO NOTHING;
