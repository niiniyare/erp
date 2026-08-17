package finance_test

import (
	"slices"
	"testing"

	"awo.so/awo/compiler"
	"awo.so/awo/def"
	"awo.so/awo/registry"

	// Import finance package to trigger init() registration.
	_ "awo.so/modules/finance"
)

// financeEntityDefs returns all finance entity definitions from the global registry.
// Uses def.All() which returns registered definitions before Seal.
func financeEntityDefs() []def.EntityDefinition {
	all := def.All()
	var finance []def.EntityDefinition
	for _, d := range all {
		if d.EntityModule() == "finance" {
			finance = append(finance, d)
		}
	}
	return finance
}

// buildFinanceSchema compiles only finance entities in isolation.
func buildFinanceSchema(t *testing.T) *compiler.CompiledSchema {
	t.Helper()
	defs := financeEntityDefs()
	if len(defs) == 0 {
		t.Fatal("no finance entities found — init() did not run")
	}
	reg, err := registry.BuildFrom(defs)
	if err != nil {
		t.Fatalf("registry.BuildFrom: %v", err)
	}
	cs, err := compiler.Compile(reg)
	if err != nil {
		t.Fatalf("compiler.Compile: %v", err)
	}
	return cs
}

// TestFinanceEntitiesRegister verifies all finance entities are registered.
func TestFinanceEntitiesRegister(t *testing.T) {
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
		if d := def.Lookup(name); d == nil {
			t.Errorf("entity %q not registered", name)
		}
	}
}

// TestFinanceEntitiesCompile verifies finance entities pass schema compilation.
func TestFinanceEntitiesCompile(t *testing.T) {
	cs := buildFinanceSchema(t)

	entities := []string{
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

	for _, name := range entities {
		es, ok := cs.ByName[name]
		if !ok {
			t.Errorf("entity %q missing from compiled schema", name)
			continue
		}
		if es.TableName == "" {
			t.Errorf("entity %q: empty TableName", name)
		}
		if es.RoutePrefix == "" {
			t.Errorf("entity %q: empty RoutePrefix", name)
		}
	}
}

// TestFinanceImmutableFields verifies immutable field declarations.
func TestFinanceImmutableFields(t *testing.T) {
	cs := buildFinanceSchema(t)

	cases := []struct {
		entity string
		field  string
	}{
		{"finance_bank_transaction", "amount"},
		{"finance_bank_transaction", "transaction_date"},
		{"finance_exchange_rate", "rate"},
		{"finance_exchange_rate", "effective_date"},
		{"finance_journal_entry", "journal"},
		{"finance_currency", "code"},
	}

	for _, tc := range cases {
		es, ok := cs.ByName[tc.entity]
		if !ok {
			t.Errorf("entity %q not found", tc.entity)
			continue
		}
		fd, ok := es.FieldsByName[tc.field]
		if !ok {
			t.Errorf("entity %q: field %q not found", tc.entity, tc.field)
			continue
		}
		if !fd.Immutable {
			t.Errorf("entity %q: field %q should be Immutable", tc.entity, tc.field)
		}
	}
}

// TestFinanceStateMachineOptions verifies state machine field options.
func TestFinanceStateMachineOptions(t *testing.T) {
	cs := buildFinanceSchema(t)

	cases := []struct {
		entity  string
		field   string
		wantOpt string
	}{
		{"finance_journal_entry", "status", "draft"},
		{"finance_journal_entry", "status", "submitted"},
		{"finance_journal_entry", "status", "posted"},
		{"finance_journal_entry", "status", "reversed"},
		{"finance_payment", "status", "draft"},
		{"finance_payment", "status", "submitted"},
		{"finance_payment", "status", "processed"},
		{"finance_payment", "status", "reconciled"},
	}

	for _, tc := range cases {
		es, ok := cs.ByName[tc.entity]
		if !ok {
			t.Errorf("entity %q not found", tc.entity)
			continue
		}
		fd, ok := es.FieldsByName[tc.field]
		if !ok {
			t.Errorf("entity %q: field %q not found", tc.entity, tc.field)
			continue
		}
		found := slices.Contains(fd.Options, tc.wantOpt)
		if !found {
			t.Errorf("entity %q: field %q missing option %q (got %v)",
				tc.entity, tc.field, tc.wantOpt, fd.Options)
		}
	}
}

// TestFinancePermissionIdentifiers verifies all finance entities declare
// at least read permissions following the naming convention.
func TestFinancePermissionIdentifiers(t *testing.T) {
	cs := buildFinanceSchema(t)

	for _, es := range cs.Entities {
		if es.Module != "finance" {
			continue
		}
		perms := es.Permissions
		if len(perms.Read) == 0 {
			t.Errorf("entity %q: no Read permissions declared", es.QualifiedName)
			continue
		}
		// Convention: "finance.{entity_local_name}.read"
		for _, p := range perms.Read {
			if len(p) == 0 {
				t.Errorf("entity %q: empty Read permission identifier", es.QualifiedName)
			}
		}
	}
}
