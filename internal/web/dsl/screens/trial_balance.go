package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// TrialBalanceScreen renders the Trial Balance report.
func TrialBalanceScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Trial Balance",
		Body: []ast.Node{
			blocks.ReportHeaderBlock(sess, blocks.ReportHeaderConfig{Title: "Trial Balance"}),
			blocks.ReportFilterBlock(sess, blocks.ReportFilterConfig{
				ShowPeriod: true, ShowCurrency: true,
			}),
			blocks.ReportTableBlock(sess, blocks.ReportTableConfig{
				Columns: []blocks.ReportColumnDef{
					{Name: "account_code", Label: "Code", Type: "text"},
					{Name: "account_name", Label: "Account", Type: "text"},
					{Name: "debit", Label: "Debit", Type: "currency"},
					{Name: "credit", Label: "Credit", Type: "currency"},
				},
				ShowGrandTotal: true,
				Exportable:     true,
			}),
		},
	}
}
