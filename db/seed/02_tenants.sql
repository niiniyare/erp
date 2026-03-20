-- =====================================================================
-- TENANTS
-- Migration: 000053 (tenants table)
-- Notes:
--   - "Status" column is quoted (capital S) — required by schema definition
--   - company_size values must be uppercase: STARTUP/SMALL/MEDIUM/LARGE/ENTERPRISE
--   - plan_tier: STARTER/GROWTH/PROFESSIONAL/ENTERPRISE
--   - Inserting a tenant auto-creates tenant_configurations via trigger (000102)
-- =====================================================================

INSERT INTO tenants (
  slug,
  name,
  email,
  billing_email,
  billing_contact_name,
  subdomain,
  "Status",
  plan_tier,
  timezone,
  currency_code,
  industry,
  company_size,
  tax_id,
  registration_number,
  legal_entity_type,
  metadata,
  settings
)
VALUES
  (
    'acme-corp',
    'ACME Corporation',
    'admin@acme-corp.com',
    'billing@acme-corp.com',
    'Finance Department',
    'acme',
    'ACTIVE',
    'ENTERPRISE',
    'America/New_York',
    'USD',
    'Technology',
    'LARGE',
    'TAX-ACME-12345',
    'REG-ACME-001',
    'LLC',
    '{"crm_id": "CRM-ACME-001", "stripe_customer_id": "cus_acme_001"}',
    '{"theme": "light", "notifications": true}'
  ),
  (
    'globex',
    'Globex Corporation',
    'info@globex.com',
    'ap@globex.com',
    'Accounts Payable',
    'globex',
    'ACTIVE',
    'ENTERPRISE',
    'Asia/Tokyo',
    'JPY',
    'Manufacturing',
    'ENTERPRISE',
    'TAX-GLOBEX-67890',
    'REG-GLOBEX-002',
    'K.K.',
    '{"crm_id": "CRM-GLOBEX-002"}',
    '{"theme": "light", "notifications": true}'
  ),
  (
    'stark-ind',
    'Stark Industries',
    'contact@stark.com',
    'billing@stark.com',
    'Pepper Potts',
    'stark',
    'ACTIVE',
    'ENTERPRISE',
    'America/Los_Angeles',
    'USD',
    'Defense',
    'ENTERPRISE',
    'TAX-STARK-45678',
    'REG-STARK-003',
    'Inc.',
    '{"crm_id": "CRM-STARK-003", "cleared_vendor": true}',
    '{"theme": "dark", "notifications": true, "security_level": "high"}'
  )
ON CONFLICT (slug) DO NOTHING;
