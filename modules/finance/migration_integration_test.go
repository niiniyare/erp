package finance_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"

	"awo.so/awo/def"
	"awo.so/awo/generator"
	testdb "awo.so/awo/testutil/db"
)

// financeMigrationSuite applies all generated finance migrations to a real
// PostgreSQL database and validates schema correctness.
type financeMigrationSuite struct {
	suite.Suite
	pool *pgxpool.Pool
	plan *generator.Plan
}

func TestFinanceMigrationSuite(t *testing.T) {
	suite.Run(t, new(financeMigrationSuite))
}

// SetupTest resets to superuser role before each test so tests are isolated.
func (s *financeMigrationSuite) SetupTest() {
	testdb.ResetRole(s.T(), s.pool)
}

func (s *financeMigrationSuite) SetupSuite() {
	s.pool = testdb.SetupTestDB(s.T())

	cs := buildFinanceSchema(s.T())
	plan, err := generator.Generate(cs, generator.Options{})
	s.Require().NoError(err)
	s.plan = plan

	// platform_tenant stub: needed because finance tenant-scoped entities
	// reference it as a FK. Create before applying finance migrations.
	platformTenantSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS platform_tenant (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    CONSTRAINT platform_tenant_pkey PRIMARY KEY (id)
);
GRANT SELECT, INSERT, UPDATE, DELETE ON platform_tenant TO %s;
`, testdb.AppRole)
	testdb.ApplySQL(s.T(), s.pool, platformTenantSQL)

	// Apply infra file first (functions, extensions).
	testdb.ApplySQL(s.T(), s.pool, plan.Files[0].SQL)

	// Apply entity migration files.
	for _, f := range plan.Files[1:] {
		// Add GRANT after each CREATE TABLE so AppRole can access it.
		entityName := extractTableName(f.SQL)
		sql := f.SQL
		if entityName != "" {
			sql += fmt.Sprintf("\nGRANT SELECT, INSERT, UPDATE, DELETE ON %q TO %s;\n",
				entityName, testdb.AppRole)
		}
		testdb.ApplySQL(s.T(), s.pool, sql)
	}
}

// extractTableName pulls the table name from "CREATE TABLE IF NOT EXISTS "name"" in DDL.
func extractTableName(sql string) string {
	const marker = "CREATE TABLE IF NOT EXISTS \""
	idx := strings.Index(sql, marker)
	if idx < 0 {
		return ""
	}
	rest := sql[idx+len(marker):]
	end := strings.Index(rest, "\"")
	if end < 0 {
		return ""
	}
	return rest[:end]
}

// TestAllTablesExist verifies all 14 finance entity tables were created.
func (s *financeMigrationSuite) TestAllTablesExist() {
	expected := []string{
		"finance_currency",
		"finance_exchange_rate",
		"finance_chart_of_accounts",
		"finance_account",
		"finance_bank_account",
		"finance_bank_transaction",
		"finance_journal",
		"finance_journal_entry",
		"finance_payment_method",
		"finance_payment",
		"finance_tax_group",
		"finance_tax",
		"finance_fiscal_year",
		"finance_accounting_period",
	}
	for _, name := range expected {
		s.True(testdb.TableExists(s.T(), s.pool, name), "table %q should exist", name)
	}
}

// TestScopeSystem_NoPlatformTenantFK verifies system-scoped entities
// (finance_currency) have no tenant_id column.
func (s *financeMigrationSuite) TestScopeSystem_NoPlatformTenantFK() {
	systemEntities := []string{
		"finance_currency",
	}
	for _, name := range systemEntities {
		exists := s.columnExists(name, "tenant_id")
		s.False(exists, "ScopeSystem entity %q must NOT have tenant_id column", name)
	}
}

// TestScopeTenant_HasTenantID verifies tenant-scoped entities have tenant_id.
func (s *financeMigrationSuite) TestScopeTenant_HasTenantID() {
	tenantEntities := []string{
		"finance_exchange_rate",
		"finance_chart_of_accounts",
		"finance_account",
		"finance_bank_account",
		"finance_bank_transaction",
		"finance_journal",
		"finance_journal_entry",
		"finance_payment_method",
		"finance_payment",
		"finance_tax_group",
		"finance_tax",
		"finance_fiscal_year",
		"finance_accounting_period",
	}
	for _, name := range tenantEntities {
		exists := s.columnExists(name, "tenant_id")
		s.True(exists, "ScopeTenant entity %q must have tenant_id column", name)
	}
}

// TestRLS_TenantIsolation verifies RLS on a tenant-scoped table isolates rows.
// Uses finance_fiscal_year which only requires tenant_id (no other FKs).
func (s *financeMigrationSuite) TestRLS_TenantIsolation() {
	tenantA := testdb.RawTenantID()
	tenantB := testdb.RawTenantID()
	ctx := context.Background()

	// Insert tenant rows into platform_tenant so FK constraint passes.
	testdb.ResetRole(s.T(), s.pool)
	for _, tid := range []uuid.UUID{tenantA, tenantB} {
		_, err := s.pool.Exec(ctx, `INSERT INTO platform_tenant (id) VALUES ($1) ON CONFLICT DO NOTHING`, tid)
		s.Require().NoError(err)
	}

	// Insert one finance_fiscal_year row per tenant as superuser (bypasses RLS).
	insertFY := func(tenantID uuid.UUID, name string) {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO finance_fiscal_year (tenant_id, name, start_date, end_date, status)
			VALUES ($1, $2, '2026-01-01', '2026-12-31', 'draft')
		`, tenantID, name)
		s.Require().NoError(err, "insert finance_fiscal_year for tenant %s", tenantID)
	}
	testdb.ResetRole(s.T(), s.pool)
	insertFY(tenantA, "FY2026-A")
	insertFY(tenantB, "FY2026-B")

	// Activate tenant A → RLS should show only A's row.
	testdb.ActivateTenant(s.T(), s.pool, tenantA)
	countA := testdb.RowCount(s.T(), s.pool, "finance_fiscal_year")
	s.Equal(1, countA, "tenant A should see only its own finance_fiscal_year rows")

	// Switch to tenant B → RLS should show only B's row.
	testdb.ActivateTenant(s.T(), s.pool, tenantB)
	countB := testdb.RowCount(s.T(), s.pool, "finance_fiscal_year")
	s.Equal(1, countB, "tenant B should see only its own finance_fiscal_year rows")
}

// TestRLS_ScopeSystem_NoRLSPolicy verifies system-scoped tables have no RLS.
func (s *financeMigrationSuite) TestRLS_ScopeSystem_NoRLSPolicy() {
	// finance_currency is ScopeSystem. Any user (incl AppRole) should see all rows.
	ctx := context.Background()

	// Insert a currency row as superuser.
	testdb.ResetRole(s.T(), s.pool)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO finance_currency (code, name, symbol, decimal_places)
		VALUES ('KES', 'Kenyan Shilling', 'KSh', 2)
		ON CONFLICT DO NOTHING
	`)
	s.Require().NoError(err)

	// Activate AppRole without a specific tenant — system table should still be readable.
	testdb.ResetRole(s.T(), s.pool)
	_, err = s.pool.Exec(ctx, fmt.Sprintf("SET ROLE %s", testdb.AppRole))
	s.Require().NoError(err)

	count := testdb.RowCount(s.T(), s.pool, "finance_currency")
	s.GreaterOrEqual(count, 1, "finance_currency (ScopeSystem) should be readable without tenant context")
}

// TestStandardColumns verifies all tables have the required standard columns.
func (s *financeMigrationSuite) TestStandardColumns() {
	tables := []string{
		"finance_currency",
		"finance_account",
		"finance_journal_entry",
		"finance_payment",
	}
	for _, tbl := range tables {
		for _, col := range []string{"id", "created_at", "updated_at", "deleted_at"} {
			s.True(s.columnExists(tbl, col), "table %q missing standard column %q", tbl, col)
		}
	}
}

// TestInfraFunctions verifies framework functions exist after migration.
func (s *financeMigrationSuite) TestInfraFunctions() {
	ctx := context.Background()
	// current_tenant_id() should exist (defined in infra migration).
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM pg_proc p
			JOIN pg_namespace n ON n.oid = p.pronamespace
			WHERE p.proname = 'current_tenant_id'
			  AND n.nspname = current_schema()
		)
	`).Scan(&exists)
	s.Require().NoError(err)
	s.True(exists, "current_tenant_id() function should exist")
}

// TestPlanFileCount verifies plan has 1 infra + 14 entity files.
func (s *financeMigrationSuite) TestPlanFileCount() {
	// 1 infra + 14 finance entities
	s.Equal(15, len(s.plan.Files), "expected 15 migration files (1 infra + 14 entities)")
}

// TestEntityScopeAnnotation verifies compiler correctly sets Scope on entities.
func (s *financeMigrationSuite) TestEntityScopeAnnotation() {
	cs := buildFinanceSchema(s.T())

	systemEntities := map[string]bool{
		"finance_currency": true,
	}

	for _, es := range cs.Entities {
		if es.Module != "finance" {
			continue
		}
		if systemEntities[es.QualifiedName] {
			s.Equal(def.ScopeSystem, es.Scope,
				"entity %q should have ScopeSystem", es.QualifiedName)
		} else {
			s.NotEqual(def.ScopeSystem, es.Scope,
				"entity %q should NOT have ScopeSystem", es.QualifiedName)
		}
	}
}

// columnExists checks whether a column exists on a table in the current schema.
func (s *financeMigrationSuite) columnExists(table, column string) bool {
	s.T().Helper()
	var exists bool
	err := s.pool.QueryRow(context.Background(), `
		SELECT EXISTS(
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = current_schema()
			  AND table_name = $1
			  AND column_name = $2
		)
	`, table, column).Scan(&exists)
	s.Require().NoError(err)
	return exists
}
