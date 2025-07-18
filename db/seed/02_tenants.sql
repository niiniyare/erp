-- Seed tenant data
INSERT INTO tenants (slug, name, email, subdomain, status) 
VALUES 
('acme-corp', 'ACME Corporation', 'admin@acme-corp.com', 'acme', 'active'),
('globex', 'Globex Corporation', 'info@globex.com', 'globex', 'active'),
('stark-ind', 'Stark Industries', 'contact@stark.com', 'stark', 'active')
ON CONFLICT (slug) DO NOTHING;
