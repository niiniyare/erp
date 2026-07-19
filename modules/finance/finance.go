// Package finance is the canonical Awo Finance module.
//
// Finance is the reference implementation for all ERP modules — it exercises
// every framework capability: entity definitions, hooks, workflow triggers,
// custom actions, RBAC policies, SDUI page builders, migrations, and tests.
//
// Entity registry (finance_* qualified names):
//
//	Core:      currency, exchange_rate
//	Periods:   fiscal_year, accounting_period
//	COA:       chart_of_accounts, account, cost_center
//	Journal:   journal, journal_entry (mandatory), journal_entry_line, ledger_entry (mandatory)
//	Tax:       tax_group, tax, tax_entry (mandatory)
//	Payment:   payment_method, payment (mandatory), allocation
//	Documents: invoice, invoice_line, credit_note, debit_note, receipt
//	Banking:   bank_account, bank_transaction
package finance

import "awo.so/awo/def"

func init() {
	// Core
	def.Register(&CurrencyDefinition)
	def.Register(&ExchangeRateDefinition)
	// Periods
	def.Register(&FiscalYearDefinition)
	def.Register(&AccountingPeriodDefinition)
	// Chart of Accounts
	def.Register(&ChartOfAccountsDefinition)
	def.Register(&AccountDefinition)
	def.Register(&CostCenterDefinition)
	// Journal (mandatory: journal_entry, ledger_entry)
	def.Register(&JournalDefinition)
	def.Register(&JournalEntryDefinition)
	def.Register(&JournalEntryLineDefinition)
	def.Register(&LedgerEntryDefinition)
	// Tax (mandatory: tax_entry)
	def.Register(&TaxGroupDefinition)
	def.Register(&TaxDefinition)
	def.Register(&TaxEntryDefinition)
	// Payment (mandatory: payment)
	def.Register(&PaymentMethodDefinition)
	def.Register(&PaymentDefinition)
	def.Register(&AllocationDefinition)
	// Documents
	def.Register(&InvoiceDefinition)
	def.Register(&InvoiceLineDefinition)
	def.Register(&CreditNoteDefinition)
	def.Register(&DebitNoteDefinition)
	def.Register(&ReceiptDefinition)
	// Banking
	def.Register(&BankAccountDefinition)
	def.Register(&BankTransactionDefinition)
}
