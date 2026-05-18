package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// CashFlowScreen renders the Cash Flow Statement report.
func CashFlowScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Cash Flow Statement",
		Body: []ast.Node{
			blocks.ReportHeaderBlock(sess, blocks.ReportHeaderConfig{Title: "Cash Flow Statement"}),
			blocks.ReportFilterBlock(sess, blocks.ReportFilterConfig{
				ShowPeriod: true, ShowCurrency: true, ShowComparison: true,
			}),
			blocks.ReportTableBlock(sess, blocks.ReportTableConfig{
				Columns: []blocks.ReportColumnDef{
					{Name: "activity_name", Label: "Activity", Type: "text"},
					{Name: "amount", Label: "Amount", Type: "currency"},
				},
				GroupBy:        []string{"activity_type"},
				ShowSubtotals:  true,
				ShowGrandTotal: true,
				Exportable:     true,
			}),
			blocks.ReportChartBlock(sess, blocks.ReportChartConfig{Type: blocks.ChartTypeLine}),
		},
	}
}
