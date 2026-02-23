-- =============================================================================
-- MIGRATION 002 UP: Reserved Subdomains
-- =============================================================================
-- Architecture Decision (ADR-002):
--   Reserved subdomains are stored in a table rather than hard-coded in a
--   CHECK constraint for two reasons:
--
--   1. CHECK constraints in PostgreSQL CANNOT reliably query other tables.
--      The SQL standard technically allows it, but PostgreSQL only re-validates
--      CHECK constraints on the modified row, not on concurrent inserts. This
--      means a table-querying CHECK is silently unreliable. Enforcement is
--      done via a BEFORE INSERT OR UPDATE trigger (migration 007) which fires
--      correctly under all isolation levels.
--
--   2. Operational flexibility: ops can add/remove reserved names (e.g., for
--      new infrastructure routes) with a plain INSERT/DELETE — no DDL, no
--      deployment, no migration required.
--
-- Seeded with common platform-level names. Audit before go-live and extend
-- to match your actual DNS / reverse-proxy routing rules.
-- =============================================================================

CREATE TABLE IF NOT EXISTS reserved_subdomains (
  -- subdomain must obey DNS label rules: lowercase alphanumeric, hyphens
  -- allowed in the middle, max 63 chars (RFC 1035 §2.3.4).
  subdomain VARCHAR(63) PRIMARY KEY
    CHECK (subdomain ~* '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$'),

  -- Human-readable note explaining WHY this subdomain is reserved.
  -- Invaluable when someone asks "why can't I register 'status'?"
  reason    TEXT NOT NULL DEFAULT 'Platform infrastructure',

  -- Who added this entry and when — useful in incident post-mortems.
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE reserved_subdomains IS
  'Subdomains tenants may not register. '
  'Enforcement is via trigger on tenants table (see migration 007). '
  'Ops can add rows here without a schema migration.';

COMMENT ON COLUMN reserved_subdomains.subdomain IS 'DNS label — lowercase alphanumeric plus hyphens, max 63 chars.';
COMMENT ON COLUMN reserved_subdomains.reason    IS 'Why this subdomain is reserved — required for audit clarity.';
COMMENT ON COLUMN reserved_subdomains.created_at IS 'When this reservation was added.';

-- -------------------------------------------------------------------------
-- SEED DATA
-- -------------------------------------------------------------------------
-- ON CONFLICT DO NOTHING makes this block safe to re-run.
-- Add your actual infrastructure subdomains here before go-live.
-- -------------------------------------------------------------------------
INSERT INTO reserved_subdomains (subdomain, reason) VALUES
  ('www',         'Primary web entry point'),
  ('api',         'REST / GraphQL API gateway'),
  ('admin',       'Internal administration dashboard'),
  ('app',         'Main application entry point'),
  ('mail',        'Email MX routing'),
  ('smtp',        'Outbound mail server'),
  ('pop',         'POP3 mail retrieval'),
  ('imap',        'IMAP mail retrieval'),
  ('ftp',         'File transfer (legacy)'),
  ('sftp',        'Secure file transfer'),
  ('static',      'Static asset CDN origin'),
  ('assets',      'Compiled asset serving'),
  ('cdn',         'Content delivery network'),
  ('media',       'User media uploads'),
  ('status',      'Platform status page'),
  ('help',        'Help center / knowledge base'),
  ('docs',        'Developer documentation'),
  ('blog',        'Company blog'),
  ('support',     'Customer support portal'),
  ('billing',     'Billing and invoicing portal'),
  ('auth',        'Authentication service (OAuth / SAML)'),
  ('login',       'Login redirect endpoint'),
  ('signup',      'Self-serve registration flow'),
  ('dashboard',   'Platform-level dashboard'),
  ('portal',      'Customer portal'),
  ('health',      'Service health check endpoint'),
  ('metrics',     'Internal Prometheus / StatsD endpoint'),
  ('monitoring',  'Observability and alerting UI'),
  ('internal',    'Internal tooling namespace'),
  ('intranet',    'Internal employee tools'),
  ('dev',         'Development / preview environment namespace'),
  ('staging',     'Staging environment namespace'),
  ('sandbox',     'Sandbox / demo environment'),
  ('test',        'Test environment namespace'),
  ('preview',     'Branch preview deployments'),
  ('system',      'System-level operations namespace'),
  ('platform',    'Platform management tools'),
  ('security',    'Security and compliance tooling'),
  ('compliance',  'Regulatory compliance portal'),
  ('webhooks',    'Incoming webhook receiver'),
  ('events',      'Event streaming endpoint')
ON CONFLICT (subdomain) DO NOTHING;
