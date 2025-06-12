-- =============================================================================
-- CUSTOMER AND VENDOR MANAGEMENT TABLES
-- =============================================================================

/*
 * Table: customer
 */
CREATE TABLE IF NOT EXISTS customer (
-- Audit fields
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier
  tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE, 

-- Core customer information
  customer_name VARCHAR(100) NOT NULL, -- Customer business/personal name
  customer_number VARCHAR(30) NOT NULL, -- Internal customer reference number
  description TEXT NOT NULL, -- Customer notes/description
  active BOOLEAN NOT NULL, -- Whether customer is currently active
  hidden BOOLEAN NOT NULL, -- Whether to hide from customer lists
  entity__id UUID NOT NULL, -- Reference to owning entity

-- Contact information
  address JSONB DEFAULT '{}'::jsonb, -- Entity-specific address
  -- address_1 VARCHAR(70) NOT NULL, -- Primary billing address
  -- address_2 VARCHAR(70) NULL, -- Secondary address line
  -- city VARCHAR(70) NULL, -- City name
  -- state VARCHAR(70) NULL, -- State/province
  -- zip_code VARCHAR(20) NULL, -- Postal/ZIP code
  -- country VARCHAR(70) NULL, -- Country name
  email VARCHAR(254) NULL, -- Primary email for invoices
  website VARCHAR(200) NULL, -- Customer website
  phone VARCHAR(30) NULL, -- Primary contact phone

-- Business settings
  sales_tax_rate REAL NULL, -- Default sales tax rate (decimal)
  additional_info JSONB NULL -- Additional custom fields (JSON)
);


COMMENT ON TABLE customer IS ' * Purpose: Stores customer information and contact details
 * Description: Maintains customer database with contact information, billing
 *              details, and sales tax rates for invoicing purposes';


