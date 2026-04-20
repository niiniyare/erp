-- Finance Tax Module
-- Tax authorities, codes, and brackets with multi-jurisdiction support.

CREATE TABLE IF NOT EXISTS finance_tax_authorities (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        UUID         NOT NULL,
    authority_code   VARCHAR(20)  NOT NULL,
    authority_name   VARCHAR(255) NOT NULL,
    authority_type   VARCHAR(20)  NOT NULL
                         CHECK (authority_type IN ('federal','state','local','municipal','vat','customs')),

    -- Jurisdiction
    country_code          CHAR(2)      NOT NULL,
    state_province_code   VARCHAR(10),
    jurisdiction_level    VARCHAR(20)  NOT NULL DEFAULT 'national',

    -- Filing rules
    filing_frequency VARCHAR(20)  NOT NULL DEFAULT 'monthly'
                         CHECK (filing_frequency IN ('weekly','monthly','quarterly','annual')),
    filing_due_day   INT          NOT NULL DEFAULT 20,
    payment_due_day  INT          NOT NULL DEFAULT 20,

    -- E-filing
    supports_e_filing   BOOLEAN      NOT NULL DEFAULT FALSE,
    e_filing_endpoint   TEXT,

    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_by UUID,

    CONSTRAINT uq_tax_authority_code UNIQUE (tenant_id, authority_code)
);

CREATE INDEX idx_finance_tax_authorities_tenant  ON finance_tax_authorities(tenant_id);
CREATE INDEX idx_finance_tax_authorities_country ON finance_tax_authorities(country_code);

ALTER TABLE finance_tax_authorities ENABLE ROW LEVEL SECURITY;

CREATE POLICY finance_tax_authorities_tenant_isolation
    ON finance_tax_authorities
    USING (tenant_id = current_tenant_id());

-- Tax codes — one rate rule per authority + type + effective-date window
CREATE TABLE IF NOT EXISTS finance_tax_codes (
    id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID         NOT NULL,
    code              VARCHAR(50)  NOT NULL,
    name              VARCHAR(255) NOT NULL,
    description       TEXT,

    tax_type              VARCHAR(20)  NOT NULL
                              CHECK (tax_type IN ('VAT','GST','WHT','PAYE','NSSF','NHIF','SHIF','SDL','NITA','EXCISE','CUSTOMS','CIT')),
    tax_category          VARCHAR(30)  NOT NULL DEFAULT 'standard',
    tax_authority_id      UUID         NOT NULL REFERENCES finance_tax_authorities(id) ON DELETE RESTRICT,
    calculation_method    VARCHAR(20)  NOT NULL
                              CHECK (calculation_method IN ('PERCENTAGE','FIXED_AMOUNT','PROGRESSIVE','LOOKUP_TABLE')),

    tax_rate          NUMERIC(10,6) NOT NULL DEFAULT 0 CHECK (tax_rate >= 0),

    compound_tax      BOOLEAN      NOT NULL DEFAULT FALSE,
    cascade_order     INT          NOT NULL DEFAULT 0,

    -- Effective dating
    effective_date    DATE         NOT NULL,
    expiry_date       DATE,

    -- Optional thresholds
    minimum_amount    NUMERIC(18,4),
    maximum_amount    NUMERIC(18,4),

    -- GL account mappings (set after chart of accounts configured)
    tax_payable_account_id     UUID,
    tax_expense_account_id     UUID,
    tax_receivable_account_id  UUID,

    -- Reporting
    reporting_code      VARCHAR(50),
    return_line_number  VARCHAR(20),

    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  UUID,
    updated_by  UUID,

    CONSTRAINT uq_tax_code UNIQUE (tenant_id, code)
);

CREATE INDEX idx_finance_tax_codes_tenant    ON finance_tax_codes(tenant_id);
CREATE INDEX idx_finance_tax_codes_type      ON finance_tax_codes(tenant_id, tax_type);
CREATE INDEX idx_finance_tax_codes_authority ON finance_tax_codes(tax_authority_id);
CREATE INDEX idx_finance_tax_codes_effective ON finance_tax_codes(effective_date, expiry_date);

ALTER TABLE finance_tax_codes ENABLE ROW LEVEL SECURITY;

CREATE POLICY finance_tax_codes_tenant_isolation
    ON finance_tax_codes
    USING (tenant_id = current_tenant_id());

-- Progressive tax brackets (PAYE, CIT)
CREATE TABLE IF NOT EXISTS finance_tax_brackets (
    id                   UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tax_code_id          UUID          NOT NULL REFERENCES finance_tax_codes(id) ON DELETE CASCADE,
    bracket_number       INT           NOT NULL,
    minimum_amount       NUMERIC(18,4) NOT NULL DEFAULT 0,
    maximum_amount       NUMERIC(18,4),
    tax_rate             NUMERIC(10,6) NOT NULL CHECK (tax_rate >= 0),
    marginal_calculation BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_tax_bracket UNIQUE (tax_code_id, bracket_number)
);

CREATE INDEX idx_finance_tax_brackets_code ON finance_tax_brackets(tax_code_id);
