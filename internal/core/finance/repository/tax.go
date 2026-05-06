package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/tracing"
)

type taxRepository struct {
	store   db.Store
	tracing tracing.Service
}

// NewTaxRepository returns a new domain.TaxRepository.
func NewTaxRepository(store db.Store, tracing tracing.Service) domain.TaxRepository {
	return &taxRepository{store: store, tracing: tracing}
}

// ── Tax Authority ─────────────────────────────────────────────────────────────

func (r *taxRepository) CreateAuthority(ctx context.Context, a *domain.TaxAuthority) error {
	ctx, span := r.tracing.StartSpan(ctx, "TaxRepository.CreateAuthority")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
INSERT INTO finance_tax_authorities
  (tenant_id, authority_code, authority_name, authority_type,
   country_code, state_province_code, jurisdiction_level,
   filing_frequency, filing_due_day, payment_due_day,
   supports_e_filing, e_filing_endpoint, is_active, created_by)
VALUES
  (current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING id, created_at, updated_at`

		return tx.QueryRow(ctx, q,
			a.AuthorityCode,
			a.AuthorityName,
			string(a.AuthorityType),
			a.CountryCode,
			nullString(a.StateProvinceCode),
			a.JurisdictionLevel,
			string(a.FilingFrequency),
			a.FilingDueDay,
			a.PaymentDueDay,
			a.SupportsEFiling,
			nullString(a.EFilingEndpoint),
			a.IsActive,
			nullUUID(a.CreatedBy),
		).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	})
}

func (r *taxRepository) GetAuthorityByID(ctx context.Context, id uuid.UUID) (*domain.TaxAuthority, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TaxRepository.GetAuthorityByID")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.TaxAuthority
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, authority_code, authority_name, authority_type,
       country_code, state_province_code, jurisdiction_level,
       filing_frequency, filing_due_day, payment_due_day,
       supports_e_filing, e_filing_endpoint, is_active,
       created_at, updated_at, created_by, updated_by
FROM   finance_tax_authorities
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1`

		a := &domain.TaxAuthority{}
		var authorityType, filingFreq string
		var createdBy *uuid.UUID
		err = tx.QueryRow(ctx, q, id).Scan(
			&a.ID, &a.TenantID, &a.AuthorityCode, &a.AuthorityName, &authorityType,
			&a.CountryCode, &a.StateProvinceCode, &a.JurisdictionLevel,
			&filingFreq, &a.FilingDueDay, &a.PaymentDueDay,
			&a.SupportsEFiling, &a.EFilingEndpoint, &a.IsActive,
			&a.CreatedAt, &a.UpdatedAt, &createdBy, &a.UpdatedBy,
		)
		if err == pgx.ErrNoRows {
			return domain.ErrTaxAuthorityNotFound
		}
		if err != nil {
			return fmt.Errorf("get tax authority: %w", err)
		}
		if createdBy != nil {
			a.CreatedBy = *createdBy
		}
		a.AuthorityType = domain.TaxAuthorityType(authorityType)
		a.FilingFrequency = domain.FilingFrequency(filingFreq)
		result = a
		return nil
	})
	return result, err
}

func (r *taxRepository) GetAuthorityByCode(ctx context.Context, tenantID uuid.UUID, code string) (*domain.TaxAuthority, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TaxRepository.GetAuthorityByCode")
	defer span.End()

	var result *domain.TaxAuthority
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, authority_code, authority_name, authority_type,
       country_code, state_province_code, jurisdiction_level,
       filing_frequency, filing_due_day, payment_due_day,
       supports_e_filing, e_filing_endpoint, is_active,
       created_at, updated_at, created_by, updated_by
FROM   finance_tax_authorities
WHERE  tenant_id      = current_tenant_id()
  AND  authority_code = $1`

		a := &domain.TaxAuthority{}
		var authorityType, filingFreq string
		var createdBy *uuid.UUID
		err = tx.QueryRow(ctx, q, code).Scan(
			&a.ID, &a.TenantID, &a.AuthorityCode, &a.AuthorityName, &authorityType,
			&a.CountryCode, &a.StateProvinceCode, &a.JurisdictionLevel,
			&filingFreq, &a.FilingDueDay, &a.PaymentDueDay,
			&a.SupportsEFiling, &a.EFilingEndpoint, &a.IsActive,
			&a.CreatedAt, &a.UpdatedAt, &createdBy, &a.UpdatedBy,
		)
		if err == pgx.ErrNoRows {
			return domain.ErrTaxAuthorityNotFound
		}
		if err != nil {
			return fmt.Errorf("get tax authority by code: %w", err)
		}
		if createdBy != nil {
			a.CreatedBy = *createdBy
		}
		a.AuthorityType = domain.TaxAuthorityType(authorityType)
		a.FilingFrequency = domain.FilingFrequency(filingFreq)
		result = a
		return nil
	})
	return result, err
}

func (r *taxRepository) ListAuthorities(ctx context.Context, tenantID uuid.UUID, activeOnly bool) ([]*domain.TaxAuthority, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TaxRepository.ListAuthorities")
	defer span.End()

	var results []*domain.TaxAuthority
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}

		var sb strings.Builder
		sb.WriteString(`
SELECT id, tenant_id, authority_code, authority_name, authority_type,
       country_code, state_province_code, jurisdiction_level,
       filing_frequency, filing_due_day, payment_due_day,
       supports_e_filing, e_filing_endpoint, is_active,
       created_at, updated_at, created_by, updated_by
FROM   finance_tax_authorities
WHERE  tenant_id = current_tenant_id()`)
		if activeOnly {
			sb.WriteString(` AND is_active = TRUE`)
		}
		sb.WriteString(` ORDER BY authority_code`)

		rows, err := tx.Query(ctx, sb.String())
		if err != nil {
			return fmt.Errorf("list tax authorities: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			a := &domain.TaxAuthority{}
			var authorityType, filingFreq string
			var createdBy *uuid.UUID
			if err := rows.Scan(
				&a.ID, &a.TenantID, &a.AuthorityCode, &a.AuthorityName, &authorityType,
				&a.CountryCode, &a.StateProvinceCode, &a.JurisdictionLevel,
				&filingFreq, &a.FilingDueDay, &a.PaymentDueDay,
				&a.SupportsEFiling, &a.EFilingEndpoint, &a.IsActive,
				&a.CreatedAt, &a.UpdatedAt, &createdBy, &a.UpdatedBy,
			); err != nil {
				return fmt.Errorf("scan tax authority: %w", err)
			}
			if createdBy != nil {
				a.CreatedBy = *createdBy
			}
			a.AuthorityType = domain.TaxAuthorityType(authorityType)
			a.FilingFrequency = domain.FilingFrequency(filingFreq)
			results = append(results, a)
		}
		return rows.Err()
	})
	return results, err
}

func (r *taxRepository) UpdateAuthority(ctx context.Context, a *domain.TaxAuthority) error {
	ctx, span := r.tracing.StartSpan(ctx, "TaxRepository.UpdateAuthority")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
UPDATE finance_tax_authorities
SET    authority_name       = $2,
       authority_type       = $3,
       country_code         = $4,
       state_province_code  = $5,
       jurisdiction_level   = $6,
       filing_frequency     = $7,
       filing_due_day       = $8,
       payment_due_day      = $9,
       supports_e_filing    = $10,
       e_filing_endpoint    = $11,
       is_active            = $12,
       updated_at           = NOW(),
       updated_by           = $13
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1
RETURNING updated_at`

		return tx.QueryRow(ctx, q,
			a.ID,
			a.AuthorityName,
			string(a.AuthorityType),
			a.CountryCode,
			nullString(a.StateProvinceCode),
			a.JurisdictionLevel,
			string(a.FilingFrequency),
			a.FilingDueDay,
			a.PaymentDueDay,
			a.SupportsEFiling,
			nullString(a.EFilingEndpoint),
			a.IsActive,
			nullUUID2(a.UpdatedBy),
		).Scan(&a.UpdatedAt)
	})
}

// ── Tax Code ──────────────────────────────────────────────────────────────────

func (r *taxRepository) CreateTaxCode(ctx context.Context, tc *domain.TaxCode) error {
	ctx, span := r.tracing.StartSpan(ctx, "TaxRepository.CreateTaxCode")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
INSERT INTO finance_tax_codes
  (tenant_id, code, name, description, tax_type, tax_category, tax_authority_id,
   calculation_method, tax_rate, compound_tax, cascade_order,
   effective_date, expiry_date, minimum_amount, maximum_amount,
   tax_payable_account_id, tax_expense_account_id, tax_receivable_account_id,
   reporting_code, return_line_number, is_active, is_default, created_by)
VALUES
  (current_tenant_id(), $1, $2, $3, $4, $5, $6,
   $7, $8, $9, $10,
   $11, $12, $13, $14,
   $15, $16, $17,
   $18, $19, $20, $21, $22)
RETURNING id, created_at, updated_at`

		return tx.QueryRow(ctx, q,
			tc.Code,
			tc.Name,
			nullString(tc.Description),
			string(tc.TaxType),
			tc.TaxCategory,
			tc.TaxAuthorityID,
			string(tc.CalculationMethod),
			tc.TaxRate.String(),
			tc.CompoundTax,
			tc.CascadeOrder,
			tc.EffectiveDate,
			tc.ExpiryDate,
			nullDecimal(tc.MinimumAmount),
			nullDecimal(tc.MaximumAmount),
			tc.TaxPayableAccountID,
			tc.TaxExpenseAccountID,
			tc.TaxReceivableAccountID,
			nullString(tc.ReportingCode),
			nullString(tc.ReturnLineNumber),
			tc.IsActive,
			tc.IsDefault,
			nullUUID(tc.CreatedBy),
		).Scan(&tc.ID, &tc.CreatedAt, &tc.UpdatedAt)
	})
}

func (r *taxRepository) GetTaxCodeByID(ctx context.Context, id uuid.UUID) (*domain.TaxCode, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TaxRepository.GetTaxCodeByID")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.TaxCode
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, code, name, description, tax_type, tax_category, tax_authority_id,
       calculation_method, tax_rate, compound_tax, cascade_order,
       effective_date, expiry_date, minimum_amount, maximum_amount,
       tax_payable_account_id, tax_expense_account_id, tax_receivable_account_id,
       reporting_code, return_line_number, is_active, is_default,
       created_at, updated_at, created_by, updated_by
FROM   finance_tax_codes
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1`

		tc, scanErr := scanTaxCode(tx.QueryRow(ctx, q, id))
		if scanErr == pgx.ErrNoRows {
			return domain.ErrTaxCodeNotFound
		}
		if scanErr != nil {
			return fmt.Errorf("get tax code: %w", scanErr)
		}
		result = tc
		return nil
	})
	return result, err
}

func (r *taxRepository) GetTaxCodeByCode(ctx context.Context, tenantID uuid.UUID, code string) (*domain.TaxCode, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TaxRepository.GetTaxCodeByCode")
	defer span.End()

	var result *domain.TaxCode
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, code, name, description, tax_type, tax_category, tax_authority_id,
       calculation_method, tax_rate, compound_tax, cascade_order,
       effective_date, expiry_date, minimum_amount, maximum_amount,
       tax_payable_account_id, tax_expense_account_id, tax_receivable_account_id,
       reporting_code, return_line_number, is_active, is_default,
       created_at, updated_at, created_by, updated_by
FROM   finance_tax_codes
WHERE  tenant_id = current_tenant_id()
  AND  code      = $1`

		tc, scanErr := scanTaxCode(tx.QueryRow(ctx, q, code))
		if scanErr == pgx.ErrNoRows {
			return domain.ErrTaxCodeNotFound
		}
		if scanErr != nil {
			return fmt.Errorf("get tax code by code: %w", scanErr)
		}
		result = tc
		return nil
	})
	return result, err
}

func (r *taxRepository) ListTaxCodes(ctx context.Context, tenantID uuid.UUID, taxType *domain.TaxType, activeOnly bool) ([]*domain.TaxCode, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TaxRepository.ListTaxCodes")
	defer span.End()

	var results []*domain.TaxCode
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}

		var sb strings.Builder
		sb.WriteString(`
SELECT id, tenant_id, code, name, description, tax_type, tax_category, tax_authority_id,
       calculation_method, tax_rate, compound_tax, cascade_order,
       effective_date, expiry_date, minimum_amount, maximum_amount,
       tax_payable_account_id, tax_expense_account_id, tax_receivable_account_id,
       reporting_code, return_line_number, is_active, is_default,
       created_at, updated_at, created_by, updated_by
FROM   finance_tax_codes
WHERE  tenant_id = current_tenant_id()`)

		args := []interface{}{}
		if taxType != nil {
			args = append(args, string(*taxType))
			sb.WriteString(fmt.Sprintf(" AND tax_type = $%d", len(args)))
		}
		if activeOnly {
			sb.WriteString(` AND is_active = TRUE`)
		}
		sb.WriteString(` ORDER BY code`)

		rows, err := tx.Query(ctx, sb.String(), args...)
		if err != nil {
			return fmt.Errorf("list tax codes: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			tc, scanErr := scanTaxCode(rows)
			if scanErr != nil {
				return fmt.Errorf("scan tax code: %w", scanErr)
			}
			results = append(results, tc)
		}
		return rows.Err()
	})
	return results, err
}

func (r *taxRepository) UpdateTaxCode(ctx context.Context, tc *domain.TaxCode) error {
	ctx, span := r.tracing.StartSpan(ctx, "TaxRepository.UpdateTaxCode")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
UPDATE finance_tax_codes
SET    name                      = $2,
       description               = $3,
       tax_category              = $4,
       calculation_method        = $5,
       tax_rate                  = $6,
       compound_tax              = $7,
       cascade_order             = $8,
       effective_date            = $9,
       expiry_date               = $10,
       minimum_amount            = $11,
       maximum_amount            = $12,
       tax_payable_account_id    = $13,
       tax_expense_account_id    = $14,
       tax_receivable_account_id = $15,
       reporting_code            = $16,
       return_line_number        = $17,
       is_active                 = $18,
       is_default                = $19,
       updated_at                = NOW(),
       updated_by                = $20
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1
RETURNING updated_at`

		return tx.QueryRow(ctx, q,
			tc.ID,
			tc.Name,
			nullString(tc.Description),
			tc.TaxCategory,
			string(tc.CalculationMethod),
			tc.TaxRate.String(),
			tc.CompoundTax,
			tc.CascadeOrder,
			tc.EffectiveDate,
			tc.ExpiryDate,
			nullDecimal(tc.MinimumAmount),
			nullDecimal(tc.MaximumAmount),
			tc.TaxPayableAccountID,
			tc.TaxExpenseAccountID,
			tc.TaxReceivableAccountID,
			nullString(tc.ReportingCode),
			nullString(tc.ReturnLineNumber),
			tc.IsActive,
			tc.IsDefault,
			nullUUID2(tc.UpdatedBy),
		).Scan(&tc.UpdatedAt)
	})
}

// ── Tax Brackets ──────────────────────────────────────────────────────────────

func (r *taxRepository) CreateBrackets(ctx context.Context, taxCodeID uuid.UUID, brackets []*domain.TaxBracket) error {
	ctx, span := r.tracing.StartSpan(ctx, "TaxRepository.CreateBrackets")
	defer span.End()

	if len(brackets) == 0 {
		return nil
	}

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
INSERT INTO finance_tax_brackets
  (tax_code_id, bracket_number, minimum_amount, maximum_amount, tax_rate, marginal_calculation)
VALUES
  ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at`

		for _, b := range brackets {
			b.TaxCodeID = taxCodeID
			if err := tx.QueryRow(ctx, q,
				taxCodeID,
				b.BracketNumber,
				b.MinimumAmount.String(),
				nullDecimal(b.MaximumAmount),
				b.TaxRate.String(),
				b.MarginalCalculation,
			).Scan(&b.ID, &b.CreatedAt); err != nil {
				return fmt.Errorf("insert tax bracket %d: %w", b.BracketNumber, err)
			}
		}
		return nil
	})
}

func (r *taxRepository) GetBrackets(ctx context.Context, taxCodeID uuid.UUID) ([]*domain.TaxBracket, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TaxRepository.GetBrackets")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var results []*domain.TaxBracket
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tax_code_id, bracket_number, minimum_amount, maximum_amount,
       tax_rate, marginal_calculation, created_at
FROM   finance_tax_brackets
WHERE  tax_code_id = $1
ORDER  BY bracket_number`

		rows, err := tx.Query(ctx, q, taxCodeID)
		if err != nil {
			return fmt.Errorf("get tax brackets: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			b := &domain.TaxBracket{}
			var minStr, rateStr string
			var maxStr *string
			if err := rows.Scan(
				&b.ID, &b.TaxCodeID, &b.BracketNumber,
				&minStr, &maxStr, &rateStr,
				&b.MarginalCalculation, &b.CreatedAt,
			); err != nil {
				return fmt.Errorf("scan tax bracket: %w", err)
			}
			if b.MinimumAmount, err = decimal.NewFromString(minStr); err != nil {
				return fmt.Errorf("parse minimum_amount: %w", err)
			}
			if b.TaxRate, err = decimal.NewFromString(rateStr); err != nil {
				return fmt.Errorf("parse tax_rate: %w", err)
			}
			if maxStr != nil {
				d, parseErr := decimal.NewFromString(*maxStr)
				if parseErr != nil {
					return fmt.Errorf("parse maximum_amount: %w", parseErr)
				}
				b.MaximumAmount = &d
			}
			results = append(results, b)
		}
		return rows.Err()
	})
	return results, err
}

func (r *taxRepository) DeleteBrackets(ctx context.Context, taxCodeID uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "TaxRepository.DeleteBrackets")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx,
			`DELETE FROM finance_tax_brackets WHERE tax_code_id = $1`,
			taxCodeID,
		)
		return err
	})
}

// ── Scan helpers ──────────────────────────────────────────────────────────────

// scannable abstracts pgx.Row and pgx.Rows so we can reuse scanTaxCode.
type scannable interface {
	Scan(dest ...interface{}) error
}

func scanTaxCode(row scannable) (*domain.TaxCode, error) {
	tc := &domain.TaxCode{}
	var taxType, calcMethod, rateStr string
	var createdBy *uuid.UUID
	var minStr, maxStr *string
	err := row.Scan(
		&tc.ID, &tc.TenantID, &tc.Code, &tc.Name, &tc.Description,
		&taxType, &tc.TaxCategory, &tc.TaxAuthorityID,
		&calcMethod, &rateStr, &tc.CompoundTax, &tc.CascadeOrder,
		&tc.EffectiveDate, &tc.ExpiryDate, &minStr, &maxStr,
		&tc.TaxPayableAccountID, &tc.TaxExpenseAccountID, &tc.TaxReceivableAccountID,
		&tc.ReportingCode, &tc.ReturnLineNumber, &tc.IsActive, &tc.IsDefault,
		&tc.CreatedAt, &tc.UpdatedAt, &createdBy, &tc.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	if createdBy != nil {
		tc.CreatedBy = *createdBy
	}
	tc.TaxType = domain.TaxType(taxType)
	tc.CalculationMethod = domain.TaxCalculationMethod(calcMethod)
	if tc.TaxRate, err = decimal.NewFromString(rateStr); err != nil {
		return nil, fmt.Errorf("parse tax_rate: %w", err)
	}
	if minStr != nil {
		d, parseErr := decimal.NewFromString(*minStr)
		if parseErr != nil {
			return nil, fmt.Errorf("parse minimum_amount: %w", parseErr)
		}
		tc.MinimumAmount = &d
	}
	if maxStr != nil {
		d, parseErr := decimal.NewFromString(*maxStr)
		if parseErr != nil {
			return nil, fmt.Errorf("parse maximum_amount: %w", parseErr)
		}
		tc.MaximumAmount = &d
	}
	return tc, nil
}

// nullString converts an empty string to nil for nullable DB columns.
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// nullDecimal converts a nil decimal pointer to nil for nullable DB columns.
func nullDecimal(d *decimal.Decimal) interface{} {
	if d == nil {
		return nil
	}
	return d.String()
}
