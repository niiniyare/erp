package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// BalanceSheetScreen renders the Balance Sheet report.
func BalanceSheetScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Balance Sheet",
		Body: []ast.Node{
			blocks.ReportHeaderBlock(sess, blocks.ReportHeaderConfig{Title: "Balance Sheet"}),
			blocks.ReportFilterBlock(sess, blocks.ReportFilterConfig{
				ShowPeriod: true, ShowEntityPicker: true,
				EntityURL:    "/api/v1/tenant/entities/options",
				ShowCurrency: true, ShowComparison: true,
			}),
			blocks.ReportTableBlock(sess, blocks.ReportTableConfig{
				Columns: []blocks.ReportColumnDef{
					{Name: "account_name", Label: "Account", Type: "text"},
					{Name: "balance", Label: "Balance", Type: "currency"},
					{Name: "prior_balance", Label: "Prior Balance", Type: "currency"},
				},
				GroupBy:        []string{"section", "account_type"},
				ShowSubtotals:  true,
				ShowGrandTotal: true,
				Exportable:     true,
			}),
		},
	}
}
