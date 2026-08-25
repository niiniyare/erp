package finance_test

import (
	"context"
	"strings"
	"testing"

	_ "awo.so/modules/finance"

	"awo.so/awo/compiler"
	"awo.so/awo/def"
	"awo.so/awo/generator"
	"awo.so/awo/registry"
	testdb "awo.so/awo/testutil/db"
)

// platformTenantStub creates a minimal platform_tenant table so FK constraints
// in tenant-scoped finance entities resolve without importing the platform module.
const platformTenantStub = `
CREATE TABLE IF NOT EXISTS platform_tenant (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    CONSTRAINT platform_tenant_pkey PRIMARY KEY (id)
);
GRANT SELECT ON platform_tenant TO awo_app;
`

// buildFinancePlan registers all finance entities, compiles, and generates SQL.
// Called independently in each test; BuildFrom never seals the global def registry.
func buildFinancePlan(t *testing.T) *generator.Plan {
	t.Helper()
	reg, err := registry.BuildFrom(def.All())
	if err != nil {
		t.Fatalf("registry.BuildFrom: %v", err)
	}
	schema, err := compiler.Compile(reg)
	if err != nil {
		t.Fatalf("compiler.Compile: %v", err)
	}
	plan, err := generator.Generate(schema, generator.Options{})
	if err != nil {
		t.Fatalf("generator.Generate: %v", err)
	}
	return plan
}

// ── Pure unit tests (no PG required) ─────────────────────────────────────────

func TestFinanceMigration_PlanHasCorrectFileCount(t *testing.T) {
	plan := buildFinancePlan(t)
	// 1 infrastructure file + 14 entity files = 15 total.
	if len(plan.Files) != 15 {
		t.Errorf("expected 15 migration files (1 infra + 14 entities), got %d", len(plan.Files))
	}
}

func TestFinanceMigration_FirstFileIsInfrastructure(t *testing.T) {
	plan := buildFinancePlan(t)
	if !strings.HasSuffix(plan.Files[0].Name, "_awo_infrastructure") {
		t.Errorf("first file must be infrastructure, got %q", plan.Files[0].Name)
	}
}

func TestFinanceMigration_AllEntityFilesPresent(t *testing.T) {
	plan := buildFinancePlan(t)

	want := []string{
		"finance_currency", "finance_exchange_rate",
		"finance_chart_of_accounts", "finance_account",
		"finance_bank_account", "finance_bank_transaction",
		"finance_journal", "finance_journal_entry",
		"finance_payment", "finance_payment_method",
		"finance_tax_group", "finance_tax",
		"finance_fiscal_year", "finance_accounting_period",
	}
	nameSet := make(map[string]bool, len(plan.Files))
	for _, f := range plan.Files {
		nameSet[f.Name] = true
	}
	for _, entity := range want {
		found := false
		for name := range nameSet {
			if strings.Contains(name, entity) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("migration file for %q not found in plan", entity)
		}
	}
}

func TestFinanceMigration_SystemEntitySQL_NoTenantIDColumn(t *testing.T) {
	plan := buildFinancePlan(t)
	for _, f := range plan.Files {
		if strings.Contains(f.Name, "finance_currency") && !strings.Contains(f.Name, "exchange") {
			if strings.Contains(f.SQL, "tenant_id") {
				t.Errorf("finance_currency (ScopeSystem) SQL must not contain tenant_id column")
			}
			if strings.Contains(f.SQL, "ROW LEVEL SECURITY") {
				t.Errorf("finance_currency (ScopeSystem) SQL must not contain ROW LEVEL SECURITY")
			}
			return
		}
	}
	t.Error("finance_currency migration file not found")
}

func TestFinanceMigration_TenantScopedEntitySQL_HasRLS(t *testing.T) {
	plan := buildFinancePlan(t)
	for _, f := range plan.Files {
		if strings.Contains(f.Name, "finance_fiscal_year") {
			if !strings.Contains(f.SQL, "tenant_id") {
				t.Error("finance_fiscal_year SQL must have tenant_id column")
			}
			if !strings.Contains(f.SQL, "ROW LEVEL SECURITY") {
				t.Error("finance_fiscal_year SQL must enable ROW LEVEL SECURITY")
			}
			if !strings.Contains(f.SQL, "tenant_isolation") {
				t.Error("finance_fiscal_year SQL must declare tenant_isolation RLS policy")
			}
			return
		}
	}
	t.Error("finance_fiscal_year migration file not found")
}

// ── PostgreSQL integration tests ──────────────────────────────────────────────

func TestFinanceMigration_AllTablesExist(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	plan := buildFinancePlan(t)

	testdb.ApplySQL(t, pool, platformTenantStub)
	for _, f := range plan.Files {
		testdb.ApplySQL(t, pool, f.SQL)
	}

	want := []string{
		"finance_currency", "finance_exchange_rate",
		"finance_chart_of_accounts", "finance_account",
		"finance_bank_account", "finance_bank_transaction",
		"finance_journal", "finance_journal_entry",
		"finance_payment", "finance_payment_method",
		"finance_tax_group", "finance_tax",
		"finance_fiscal_year", "finance_accounting_period",
	}
	for _, tbl := range want {
		if !testdb.TableExists(t, pool, tbl) {
			t.Errorf("table %q missing after migration", tbl)
		}
	}
}

func TestFinanceMigration_SystemEntityHasNoTenantIDColumn(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	plan := buildFinancePlan(t)

	testdb.ApplySQL(t, pool, platformTenantStub)
	for _, f := range plan.Files {
		testdb.ApplySQL(t, pool, f.SQL)
	}

	var hasTenantID bool
	if err := pool.QueryRow(context.Background(), `
		SELECT EXISTS(
			SELECT 1 FROM information_schema.columns
			WHERE table_name   = 'finance_currency'
			  AND column_name  = 'tenant_id'
			  AND table_schema = current_schema()
		)`).Scan(&hasTenantID); err != nil {
		t.Fatalf("query columns: %v", err)
	}
	if hasTenantID {
		t.Error("finance_currency (ScopeSystem) must NOT have tenant_id column")
	}
}

func TestFinanceMigration_TenantScopedEntityHasTenantIDColumn(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	plan := buildFinancePlan(t)

	testdb.ApplySQL(t, pool, platformTenantStub)
	for _, f := range plan.Files {
		testdb.ApplySQL(t, pool, f.SQL)
	}

	var hasTenantID bool
	if err := pool.QueryRow(context.Background(), `
		SELECT EXISTS(
			SELECT 1 FROM information_schema.columns
			WHERE table_name   = 'finance_fiscal_year'
			  AND column_name  = 'tenant_id'
			  AND table_schema = current_schema()
		)`).Scan(&hasTenantID); err != nil {
		t.Fatalf("query columns: %v", err)
	}
	if !hasTenantID {
		t.Error("finance_fiscal_year (ScopeTenant) must have tenant_id column")
	}
}

func TestFinanceMigration_RLSAbsentOnSystemEntity(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	plan := buildFinancePlan(t)

	testdb.ApplySQL(t, pool, platformTenantStub)
	for _, f := range plan.Files {
		testdb.ApplySQL(t, pool, f.SQL)
	}

	var hasPolicy bool
	if err := pool.QueryRow(context.Background(), `
		SELECT EXISTS(
			SELECT 1 FROM pg_policies
			WHERE tablename  = 'finance_currency'
			  AND schemaname = current_schema()
		)`).Scan(&hasPolicy); err != nil {
		t.Fatalf("query pg_policies: %v", err)
	}
	if hasPolicy {
		t.Error("finance_currency (ScopeSystem) must NOT have an RLS policy")
	}
}

func TestFinanceMigration_RLSPresentOnTenantScopedEntity(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	plan := buildFinancePlan(t)

	testdb.ApplySQL(t, pool, platformTenantStub)
	for _, f := range plan.Files {
		testdb.ApplySQL(t, pool, f.SQL)
	}

	var hasPolicy bool
	if err := pool.QueryRow(context.Background(), `
		SELECT EXISTS(
			SELECT 1 FROM pg_policies
			WHERE tablename  = 'finance_fiscal_year'
			  AND schemaname = current_schema()
		)`).Scan(&hasPolicy); err != nil {
		t.Fatalf("query pg_policies: %v", err)
	}
	if !hasPolicy {
		t.Error("finance_fiscal_year (ScopeTenant) must have an RLS policy")
	}
}
