-- Revert org_id → entity_id

ALTER TABLE persons               RENAME COLUMN org_id TO entity_id;
ALTER TABLE employees             RENAME COLUMN org_id TO entity_id;
ALTER TABLE users                 RENAME COLUMN org_id TO entity_id;
ALTER TABLE user_sessions         RENAME COLUMN org_id TO entity_id;
ALTER TABLE finance_accounts      RENAME COLUMN org_id TO entity_id;
ALTER TABLE finance_account_groups RENAME COLUMN org_id TO entity_id;
ALTER TABLE finance_account_balances RENAME COLUMN org_id TO entity_id;
ALTER TABLE finance_transactions  RENAME COLUMN org_id TO entity_id;
ALTER TABLE roles                 RENAME COLUMN org_id TO entity_id;
ALTER TABLE policies              RENAME COLUMN org_id TO entity_id;
ALTER TABLE policy_evaluations    RENAME COLUMN org_id TO entity_id;
ALTER TABLE access_requests       RENAME COLUMN org_id TO entity_id;
ALTER TABLE contracts             RENAME COLUMN org_id TO entity_id;
ALTER TABLE feature_flags         RENAME COLUMN org_id TO entity_id;
ALTER TABLE audit_logs            RENAME COLUMN org_id TO entity_id;
