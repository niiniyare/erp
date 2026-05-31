package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// AccountsScreen builds the chart of accounts listing page.
// Wires to /api/v1/finance/accounts with permission-gated create.
func AccountsScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Chart of Accounts",
		Body: []ast.Node{
			blocks.DataTableBlock(sess, blocks.DataTableConfig{
				APIURL:           "/api/v1/finance/accounts",
				Title:            "Chart of Accounts",
				PrimaryKey:       "id",
				PageSize:         50,
				AllowCreate:      true,
				CreateURL:        "/finance/accounts/new",
				CreatePermission: "finance.accounts.create",
				AllowExport:      true,
				Filter: blocks.FilterBarBlock(sess, blocks.FilterBarConfig{
					ShowSearch:        true,
					SearchPlaceholder: "Search accounts…",
					ShowTypeFilter:    true,
					TypeFieldName:     "account_type",
					TypeLabel:         "Type",
					TypeOptions: []ast.SelectOption{
						{Label: "Asset", Value: "ASSET"},
						{Label: "Liability", Value: "LIABILITY"},
						{Label: "Equity", Value: "EQUITY"},
						{Label: "Revenue", Value: "REVENUE"},
						{Label: "Expense", Value: "EXPENSE"},
					},
					ShowStatus: true,
					StatusOptions: []ast.SelectOption{
						{Label: "Active", Value: "true"},
						{Label: "Inactive", Value: "false"},
					},
				}),
				Columns: []blocks.ColumnDef{
					{Name: "code", Label: "Code", Sortable: true, Width: 120},
					{Name: "name", Label: "Account Name", Sortable: true},
					{
						Name:  "account_type",
						Label: "Type",
						Type:  "mapping",
						Map: map[string]string{
							"ASSET":     `<span class="badge badge-info">Asset</span>`,
							"LIABILITY": `<span class="badge badge-warning">Liability</span>`,
							"EQUITY":    `<span class="badge badge-primary">Equity</span>`,
							"REVENUE":   `<span class="badge badge-success">Revenue</span>`,
							"EXPENSE":   `<span class="badge badge-danger">Expense</span>`,
							"*":         `<span class="badge badge-default">${value}</span>`,
						},
					},
					{Name: "currency", Label: "Currency", Width: 100},
					{Name: "balance", Label: "Balance", Type: "number"},
					{
						Name:  "is_active",
						Label: "Active",
						Type:  "mapping",
						Map: map[string]string{
							"true":  `<i class="fa fa-check-circle text-success"></i>`,
							"false": `<i class="fa fa-times-circle text-danger"></i>`,
							"*":     `<span>${value}</span>`,
						},
						Width: 80,
					},
				},
				RowActions: []ast.ActionNode{
					{
						Label:      "Edit",
						ActionType: "link",
						Target:     "/finance/accounts/${id}",
						Icon:       "fa fa-edit",
					},
					{
						Label:       "Delete",
						ActionType:  "ajax",
						Level:       "danger",
						Icon:        "fa fa-trash",
						ConfirmText: "Delete this account? This cannot be undone.",
						API:         &ast.APISpec{Method: "delete", URL: "/api/v1/finance/accounts/${id}"},
					},
				},
			}),
		},
	}
}
