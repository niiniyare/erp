// Package migrations registers Finance module schema migrations.
//
// Finance migrations create the double-entry accounting tables:
//   - finance_fiscal_year
//   - finance_accounting_period
//   - finance_currency
//   - finance_exchange_rate
//   - finance_chart_of_accounts
//   - finance_account
//   - finance_cost_center
//   - finance_tax_group / finance_tax
//   - finance_journal / finance_journal_entry / finance_journal_entry_line
//   - finance_ledger_entry
//   - finance_payment_method / finance_payment / finance_allocation
//   - finance_bank_account / finance_bank_transaction
//   - finance_invoice / finance_invoice_line
//   - finance_credit_note / finance_debit_note / finance_receipt
//   - finance_tax_entry
//
// Import with a blank import to activate:
//
//	import _ "awo.so/modules/finance/migrations"
package migrations

import (
	"embed"

	"awo.so/awo/migration"
)

//go:embed *.sql
var sqlFS embed.FS

func init() {
	migration.Register(migration.Source{
		Module:    "finance",
		Priority:  30,
		DependsOn: []string{"bootstrap", "tenant", "iam"},
		FS:        sqlFS,
	})
}
