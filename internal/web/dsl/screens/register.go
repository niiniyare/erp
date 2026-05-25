package screens

// register.go wires all dsl/screens to the page registry via ASTFn.
// This file is the single source of truth for DSL screen routes.
//
// Edit routes (/:id) are not registered here — they require prefix/param
// matching in the registry which is not yet implemented. They will be
// added once the registry supports route parameters.
//
// Import this package (via blank import) in cmd/server to trigger init().

import (
	"awo.so/internal/web/registry"
	"awo.so/internal/web/ui"
)

func init() {
	// ── Finance Dashboard ────────────────────────────────────────────────────
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/finance/dashboard",
		Module:      "finance",
		Title:       "Finance Dashboard",
		Description: "KPI summary, revenue chart, and quick actions for the finance module",
		ASTFn: func(sess ui.UISessionContext) any {
			return FinanceDashboardScreen(sess)
		},
	})

	// ── Invoices ─────────────────────────────────────────────────────────────
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/finance/invoices/new",
		Module:      "finance",
		Title:       "New Invoice",
		Description: "Create a new sales invoice",
		ASTFn: func(sess ui.UISessionContext) any {
			return InvoiceScreen(sess, InvoiceScreenConfig{
				ShowPaymentTerms:  true,
				ShowAttachments:   true,
				ShowInternalNotes: true,
			})
		},
	})

	// ── Bills (purchase invoices) ─────────────────────────────────────────────
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/finance/bills/new",
		Module:      "finance",
		Title:       "New Bill",
		Description: "Record a new supplier bill (purchase invoice)",
		ASTFn: func(sess ui.UISessionContext) any {
			return InvoiceScreen(sess, InvoiceScreenConfig{
				IsPurchase:       true,
				ShowPaymentTerms: true,
				ShowAttachments:  true,
			})
		},
	})

	// ── Journal Entries ───────────────────────────────────────────────────────
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/finance/journal-entries/new",
		Module:      "finance",
		Title:       "New Journal Entry",
		Description: "Create a new manual journal entry",
		ASTFn: func(sess ui.UISessionContext) any {
			return JournalEntryScreen(sess, JournalEntryScreenConfig{})
		},
	})

	// ── Reports ──────────────────────────────────────────────────────────────
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/finance/reports/trial-balance",
		Module:      "finance",
		Title:       "Trial Balance",
		Description: "Trial balance report with period and currency filters",
		ASTFn: func(sess ui.UISessionContext) any {
			return TrialBalanceScreen(sess)
		},
	})
}
