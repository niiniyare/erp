package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// TransactionsScreen builds the finance transaction history listing page.
// Wires to /api/v1/finance/transactions with full filter bar:
// search, type, status, date range, currency, amount range.
func TransactionsScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Transactions",
		Body: []ast.Node{
			blocks.DataTableBlock(sess, blocks.DataTableConfig{
				APIURL:           "/api/v1/finance/transactions",
				Title:            "Transaction History",
				PrimaryKey:       "id",
				PageSize:         25,
				AllowCreate:      true,
				CreateURL:        "/finance/transactions/new",
				CreatePermission: "finance.transactions.create",
				AllowExport:      true,
				Filter: blocks.FilterBarBlock(sess, blocks.FilterBarConfig{
					ShowSearch:        true,
					SearchPlaceholder: "Search by reference…",
					ShowTypeFilter:    true,
					TypeFieldName:     "transaction_type",
					TypeLabel:         "Type",
					TypeOptions: []ast.SelectOption{
						{Label: "Journal Entry", Value: "JOURNAL"},
						{Label: "Invoice", Value: "INVOICE"},
						{Label: "Payment", Value: "PAYMENT"},
						{Label: "Receipt", Value: "RECEIPT"},
						{Label: "Transfer", Value: "TRANSFER"},
					},
					ShowStatus: true,
					StatusOptions: []ast.SelectOption{
						{Label: "Draft", Value: "DRAFT"},
						{Label: "Posted", Value: "POSTED"},
						{Label: "Void", Value: "VOID"},
					},
					ShowDateRange:   true,
					ShowCurrency:    true,
					ShowAmountRange: true,
				}),
				Columns: []blocks.ColumnDef{
					{Name: "reference_number", Label: "Reference", Sortable: true},
					{
						Name:  "transaction_type",
						Label: "Type",
						Type:  "mapping",
						Map: map[string]string{
							"JOURNAL":  `<span class="badge badge-default">Journal</span>`,
							"INVOICE":  `<span class="badge badge-primary">Invoice</span>`,
							"PAYMENT":  `<span class="badge badge-success">Payment</span>`,
							"RECEIPT":  `<span class="badge badge-info">Receipt</span>`,
							"TRANSFER": `<span class="badge badge-warning">Transfer</span>`,
							"*":        `<span class="badge badge-default">${value}</span>`,
						},
					},
					{Name: "transaction_date", Label: "Date", Type: "date", Sortable: true},
					{Name: "description", Label: "Description"},
					{Name: "amount", Label: "Amount", Type: "currency"},
					blocks.StatusBadgeColumnDef(blocks.StatusBadgeConfig{
						FieldName: "status",
						Label:     "Status",
						Mappings: []blocks.StatusMapping{
							{Value: "DRAFT", Label: "Draft", Color: "default"},
							{Value: "POSTED", Label: "Posted", Color: "success"},
							{Value: "VOID", Label: "Void", Color: "danger"},
						},
					}),
				},
				RowActions: []ast.ActionNode{
					{
						Label:      "View",
						ActionType: "link",
						Target:     "/finance/transactions/${id}",
						Icon:       "fa fa-eye",
					},
				},
			}),
		},
	}
}
