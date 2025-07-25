/*
 * Table: vendor
 */
CREATE TABLE IF NOT EXISTS vendor (
-- Audit fields
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier

-- Core vendor information
  vendor_name VARCHAR(100) NOT NULL, -- Vendor business name
  vendor_number VARCHAR(30) NULL, -- Internal vendor reference number
  description TEXT NOT NULL, -- Vendor notes/description
  active BOOLEAN NOT NULL, -- Whether vendor is currently active
  hidden BOOLEAN NOT NULL, -- Whether to hide from vendor lists
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE, 

-- Contact information
    address JSONB DEFAULT '{}'::jsonb, -- Entity-specific address
  -- address_1 VARCHAR(70) NOT NULL, -- Primary address
  -- address_2 VARCHAR(70) NULL, -- Secondary address line
  -- city VARCHAR(70) NULL, -- City name
  -- state VARCHAR(70) NULL, -- State/province
  -- zip_code VARCHAR(20) NULL, -- Postal/ZIP code
  -- country VARCHAR(70) NULL, -- Country name
  contact  JSONB DEFAULT '{}'::jsonb, -- Entity-specific address

  -- email VARCHAR(254) NULL, -- Primary contact email
  -- website VARCHAR(200) NULL, -- Vendor website
  -- phone VARCHAR(30) NULL, -- Primary contact phone

-- Banking and payment information
  account_number VARCHAR(30) NULL, -- Bank account number for payments
  routing_number VARCHAR(30) NULL, -- Bank routing number
  aba_number VARCHAR(30) NULL, -- ABA routing number
  swift_number VARCHAR(30) NULL, -- SWIFT code for international transfers
  tax_id_number VARCHAR(30) NULL, -- Tax ID/EIN for 1099 reporting
  account_type VARCHAR(20) NOT NULL, -- Account type (CHECKING, SAVINGS, etc.)
  additional_info JSONB NULL -- Additional custom fields (JSON)
);

COMMENT ON TABLE vendor IS ' Purpose: Stores vendor/supplier information and payment details
 * Description:
Maintains vendor database with contact information and banking details for bill payments and purchase orders';
