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

	// ── Invoice edit (param route) ────────────────────────────────────────────
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/finance/invoices/:id",
		Module:      "finance",
		Title:       "Invoice",
		Description: "View or edit an existing sales invoice",
		ASTFn: func(sess ui.UISessionContext) any {
			return InvoiceScreen(sess, InvoiceScreenConfig{
				ShowPaymentTerms:  true,
				ShowAttachments:   true,
				ShowInternalNotes: true,
			})
		},
	})

	// ── Bill edit (param route) ───────────────────────────────────────────────
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/finance/bills/:id",
		Module:      "finance",
		Title:       "Bill",
		Description: "View or edit an existing supplier bill",
		ASTFn: func(sess ui.UISessionContext) any {
			return InvoiceScreen(sess, InvoiceScreenConfig{
				IsPurchase:       true,
				ShowPaymentTerms: true,
				ShowAttachments:  true,
			})
		},
	})

	// ── Journal entry edit (param route) ─────────────────────────────────────
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/finance/journal-entries/:id",
		Module:      "finance",
		Title:       "Journal Entry",
		Description: "View or edit an existing journal entry",
		ASTFn: func(sess ui.UISessionContext) any {
			return JournalEntryScreen(sess, JournalEntryScreenConfig{})
		},
	})

	// ── Sell module ───────────────────────────────────────────────────────────
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/sell/quotations/new",
		Module:      "sell",
		Title:       "New Quotation",
		Description: "Create a new sales quotation",
		ASTFn:       func(sess ui.UISessionContext) any { return QuotationScreen(sess) },
	})
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/sell/orders/new",
		Module:      "sell",
		Title:       "New Sales Order",
		Description: "Create a new sales order",
		ASTFn:       func(sess ui.UISessionContext) any { return SalesOrderScreen(sess) },
	})
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/sell/delivery-notes/new",
		Module:      "sell",
		Title:       "New Delivery Note",
		Description: "Create a goods delivery confirmation",
		ASTFn:       func(sess ui.UISessionContext) any { return DeliveryNoteScreen(sess, DeliveryNoteScreenConfig{}) },
	})

	// ── Buy module ────────────────────────────────────────────────────────────
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/buy/requisitions/new",
		Module:      "buy",
		Title:       "New Requisition",
		Description: "Create an internal purchase requisition",
		ASTFn:       func(sess ui.UISessionContext) any { return RequisitionScreen(sess) },
	})
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/buy/purchase-orders/new",
		Module:      "buy",
		Title:       "New Purchase Order",
		Description: "Create a purchase order to send to a supplier",
		ASTFn:       func(sess ui.UISessionContext) any { return PurchaseOrderScreen(sess) },
	})
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/buy/goods-receipts/new",
		Module:      "buy",
		Title:       "New Goods Receipt",
		Description: "Record goods received against a purchase order",
		ASTFn:       func(sess ui.UISessionContext) any { return GoodsReceiptScreen(sess, GoodsReceiptScreenConfig{}) },
	})

	// ── Inventory module ──────────────────────────────────────────────────────
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/inventory/stock-ledger",
		Module:      "inventory",
		Title:       "Stock Ledger",
		Description: "Paginated stock movement ledger with warehouse and item filters",
		ASTFn:       func(sess ui.UISessionContext) any { return StockLedgerScreen(sess) },
	})
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/inventory/warehouses",
		Module:      "inventory",
		Title:       "Warehouses",
		Description: "Warehouse list with current stock levels",
		ASTFn:       func(sess ui.UISessionContext) any { return WarehouseScreen(sess) },
	})
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/inventory/items",
		Module:      "inventory",
		Title:       "Item Catalog",
		Description: "Item master data: UOM, reorder point, supplier links",
		ASTFn:       func(sess ui.UISessionContext) any { return ItemCatalogScreen(sess) },
	})

	// ── HR module ─────────────────────────────────────────────────────────────
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/hr/employees/new",
		Module:      "hr",
		Title:       "New Employee",
		Description: "Employee detail: personal info, department, position, contract dates",
		ASTFn:       func(sess ui.UISessionContext) any { return EmployeeScreen(sess) },
	})
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/hr/leave-requests/new",
		Module:      "hr",
		Title:       "New Leave Request",
		Description: "Leave request form with type, date range, and approval chain",
		ASTFn:       func(sess ui.UISessionContext) any { return LeaveRequestScreen(sess) },
	})
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/hr/payroll",
		Module:      "hr",
		Title:       "Payroll Summary",
		Description: "Dashboard: headcount, total payroll, pending approvals",
		ASTFn:       func(sess ui.UISessionContext) any { return PayrollSummaryScreen(sess) },
	})

	// ── IAM / AuthZ module ────────────────────────────────────────────────────
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/iam/roles/new",
		Module:      "iam",
		Title:       "New Role",
		Description: "Role detail: name, description, assigned permissions, members",
		ASTFn:       func(sess ui.UISessionContext) any { return RoleScreen(sess) },
	})
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/iam/policies",
		Module:      "iam",
		Title:       "Authorization Policies",
		Description: "Casbin policy list: subject, domain, object, action",
		ASTFn:       func(sess ui.UISessionContext) any { return PolicyScreen(sess) },
	})
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/iam/users/:id",
		Module:      "iam",
		Title:       "User",
		Description: "User detail: profile, active roles, session history",
		ASTFn:       func(sess ui.UISessionContext) any { return UserScreen(sess) },
	})
}
