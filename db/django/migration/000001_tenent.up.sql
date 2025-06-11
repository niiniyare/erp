-- =====================================================
-- ERP/ACCOUNTING SYSTEM DATABASE SCHEMA (PostgreSQL)
-- =====================================================
-- This schema represents a comprehensive accounting and project management system
-- with entities, customers, vendors, estimates, invoices, bills, and financial tracking

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- CREATE EXTENSION IF NOT EXISTS "ltree";

-- Tenant management (your existing structure is good)
CREATE TABLE tenants (
    id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    subdomain VARCHAR(63) UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'active' 
        CHECK (status IN ('active', 'suspended', 'pending')),
    industry VARCHAR(50), -- For future industry-specific modules
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Tenant configurations
CREATE TABLE tenant_configurations (
    tenant_id INT PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    max_users INT NOT NULL DEFAULT 100,
    storage_quota BIGINT NOT NULL DEFAULT 1073741824, -- 1GB
    features JSONB NOT NULL DEFAULT '{}'::jsonb,
    modules_enabled JSONB NOT NULL DEFAULT '["accounting", "inventory"]'::jsonb
);

