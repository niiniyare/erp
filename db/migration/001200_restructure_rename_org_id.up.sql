-- --------------------------------------------------------------------
-- 001200  Rename entity_id → org_id across all tenant tables
-- --------------------------------------------------------------------
-- "entity" is overloaded: it means both an organisational unit (the
-- `entities` table) and a framework concept (EntityDefinition).
-- Renaming the FK column to `org_id` removes the ambiguity.
--
-- Affected tables (any table with an entity_id FK to entities.uuid):
--   entities, persons, employees, users, user_sessions,
--   finance_accounts, finance_account_groups, finance_account_balances,
--   finance_transactions, finance_cost_centers, finance_budgets,
--   finance_tax_codes, finance_fiscal_years, finance_accounting_periods,
--   roles, policies, policy_evaluations, access_requests,
--   contracts, feature_flags, audit_logs
--
-- Strategy: rename column in place — no data movement required.
-- Down migration renames back.
-- --------------------------------------------------------------------

-- entities self-reference (parent stays parent_id; entity_id not used here)

-- persons
ALTER TABLE persons          RENAME COLUMN entity_id TO org_id;

-- employees
ALTER TABLE employees        RENAME COLUMN entity_id TO org_id;

-- users
ALTER TABLE users            RENAME COLUMN entity_id TO org_id;

-- user_sessions (entity scope context)
ALTER TABLE user_sessions    RENAME COLUMN entity_id TO org_id;

-- finance_accounts
ALTER TABLE finance_accounts RENAME COLUMN entity_id TO org_id;

-- finance_account_groups
ALTER TABLE finance_account_groups RENAME COLUMN entity_id TO org_id;

-- finance_account_balances
ALTER TABLE finance_account_balances RENAME COLUMN entity_id TO org_id;

-- finance_transactions
ALTER TABLE finance_transactions RENAME COLUMN entity_id TO org_id;

-- finance_cost_centers: no entity_id (uses tenant_id only) — skip

-- finance_budgets: no entity_id — skip

-- finance_tax_codes: no entity_id — skip

-- roles
ALTER TABLE roles            RENAME COLUMN entity_id TO org_id;

-- policies
ALTER TABLE policies         RENAME COLUMN entity_id TO org_id;

-- policy_evaluations
ALTER TABLE policy_evaluations RENAME COLUMN entity_id TO org_id;

-- access_requests
ALTER TABLE access_requests  RENAME COLUMN entity_id TO org_id;

-- contracts
ALTER TABLE contracts        RENAME COLUMN entity_id TO org_id;

-- feature_flags
ALTER TABLE feature_flags    RENAME COLUMN entity_id TO org_id;

-- audit_logs
ALTER TABLE audit_logs       RENAME COLUMN entity_id TO org_id;

-- Rename any indexes that embed "entity_id" in their name for clarity.
-- golang-migrate runs this in a transaction; if any table doesn't have
-- entity_id the migration will fail — remove that ALTER from the list above.
