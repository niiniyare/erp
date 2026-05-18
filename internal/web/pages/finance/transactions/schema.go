package transactions

import (
	"awo.so/internal/web/amis"
	"awo.so/internal/web/registry"
	"awo.so/internal/web/ui"
)

func init() {
	registry.Register("/finance/transactions", Schema)
}

var mockTransactions = amis.A{
	amis.M{"id": "txn-001", "reference_number": "JNL-2026-001", "transaction_type": "JOURNAL", "transaction_date": "2026-04-01", "description": "Monthly payroll allocation", "amount": 450000, "currency": "KES", "status": "POSTED"},
	amis.M{"id": "txn-002", "reference_number": "INV-2026-042", "transaction_type": "INVOICE", "transaction_date": "2026-04-03", "description": "Invoice — Savanna Tech Ltd", "amount": 125000, "currency": "KES", "status": "PENDING"},
	amis.M{"id": "txn-003", "reference_number": "PMT-2026-011", "transaction_type": "PAYMENT", "transaction_date": "2026-04-05", "description": "Supplier payment — Baobab Supplies", "amount": 87500, "currency": "KES", "status": "POSTED"},
	amis.M{"id": "txn-004", "reference_number": "RCP-2026-008", "transaction_type": "RECEIPT", "transaction_date": "2026-04-07", "description": "Customer receipt — Kilimanjaro Finance", "amount": 200000, "currency": "KES", "status": "POSTED"},
	amis.M{"id": "txn-005", "reference_number": "JNL-2026-002", "transaction_type": "JOURNAL", "transaction_date": "2026-04-08", "description": "Depreciation entry Q1", "amount": 35000, "currency": "KES", "status": "DRAFT"},
	amis.M{"id": "txn-006", "reference_number": "INV-2026-043", "transaction_type": "INVOICE", "transaction_date": "2026-04-09", "description": "Invoice — Serengeti Farms", "amount": 62000, "currency": "KES", "status": "DRAFT"},
	amis.M{"id": "txn-007", "reference_number": "TRF-2026-003", "transaction_type": "TRANSFER", "transaction_date": "2026-04-10", "description": "Bank transfer — savings to operations", "amount": 500000, "currency": "KES", "status": "POSTED"},
	amis.M{"id": "txn-008", "reference_number": "PMT-2026-012", "transaction_type": "PAYMENT", "transaction_date": "2026-04-11", "description": "Utility bills payment", "amount": 28500, "currency": "KES", "status": "VOIDED"},
}

func Schema(_ ui.UISessionContext) ui.Schema {
	return amis.Page("").
		Toolbar(
			amis.M{
				"type": "button", "label": "New Transaction", "icon": "fa fa-plus",
				"level": "primary", "actionType": "dialog",
				"dialog": amis.M{
					"title": "New Transaction",
					"size":  "md",
					"body": amis.Form("post:/api/v1/finance/transactions").Fields(
						amis.Required(amis.SelectField("transaction_type", "Type",
							amis.SelectOpt("Journal Entry", "JOURNAL"),
							amis.SelectOpt("Invoice", "INVOICE"),
							amis.SelectOpt("Payment", "PAYMENT"),
							amis.SelectOpt("Receipt", "RECEIPT"),
							amis.SelectOpt("Transfer", "TRANSFER"),
						)),
						amis.Required(amis.DateField("transaction_date", "Date")),
						amis.Required(amis.TextField("reference_number", "Reference #")),
						amis.Required(amis.NumberField("amount", "Amount")),
						amis.SelectField("currency", "Currency",
							amis.SelectOpt("KES", "KES"),
							amis.SelectOpt("USD", "USD"),
							amis.SelectOpt("EUR", "EUR"),
						),
						amis.Optional(amis.TextAreaField("description", "Description")),
					).Build(),
				},
			},
			amis.M{"type": "button", "label": "Export CSV", "icon": "fa fa-download"},
		).
		Body([]any{
			amis.Breadcrumb(
				amis.BC("Home", "#dashboard"),
				amis.BC("Finance"),
				amis.BC("Transactions"),
			),
			amis.CRUD("get:/api/v1/finance/transactions").
				StaticData(mockTransactions).
				Filter(amis.Form("").Fields(
					amis.TextField("reference_number", "Reference #"),
					amis.SelectField("transaction_type", "Type",
						amis.SelectOpt("All", ""),
						amis.SelectOpt("Journal", "JOURNAL"),
						amis.SelectOpt("Invoice", "INVOICE"),
						amis.SelectOpt("Payment", "PAYMENT"),
						amis.SelectOpt("Receipt", "RECEIPT"),
						amis.SelectOpt("Transfer", "TRANSFER"),
					),
					amis.SelectField("status", "Status",
						amis.SelectOpt("All", ""),
						amis.SelectOpt("Draft", "DRAFT"),
						amis.SelectOpt("Pending", "PENDING"),
						amis.SelectOpt("Posted", "POSTED"),
						amis.SelectOpt("Voided", "VOIDED"),
					),
					amis.DateRangeField("date_range", "Date Range"),
				).Build()).
				Columns(
					amis.Column("reference_number", "Reference").Sortable(),
					amis.Column("transaction_type", "Type").Map(amis.M{
						"JOURNAL":  "Journal",
						"INVOICE":  "Invoice",
						"PAYMENT":  "Payment",
						"RECEIPT":  "Receipt",
						"TRANSFER": "Transfer",
					}),
					amis.Column("transaction_date", "Date").Type("date").Sortable(),
					amis.Column("description", "Description"),
					amis.Column("amount", "Amount (KES)").Type("number").Align("right").Sortable(),
					amis.Column("status", "Status").Map(amis.M{
						"DRAFT":   "<span class='label label-default'>Draft</span>",
						"PENDING": "<span class='label label-warning'>Pending</span>",
						"POSTED":  "<span class='label label-success'>Posted</span>",
						"VOIDED":  "<span class='label label-danger'>Voided</span>",
					}),
					amis.Column("", "Actions").Buttons(
						amis.ViewBtn(txDetail()),
					),
				).
				DefaultSort("transaction_date", "desc").
				PerPage(25).
				Build(),
		}).Build()
}

func txDetail() amis.Schema {
	return amis.Descriptions("Transaction Detail").
		Item("Reference", "reference_number").
		Item("Type", "transaction_type").
		Item("Date", "transaction_date").
		Item("Amount", "amount").
		Item("Currency", "currency").
		Item("Status", "status").
		Item("Description", "description").
		Build()
}
