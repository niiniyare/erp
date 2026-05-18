package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// ProfitLossScreen renders the Profit & Loss Statement report.
func ProfitLossScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Profit & Loss Statement",
		Body: []ast.Node{
			blocks.ReportHeaderBlock(sess, blocks.ReportHeaderConfig{Title: "Profit & Loss Statement"}),
			blocks.ReportFilterBlock(sess, blocks.ReportFilterConfig{
				ShowPeriod: true, ShowEntityPicker: true,
				EntityURL:    "/api/v1/tenant/entities/options",
				ShowCurrency: true, ShowComparison: true,
			}),
			blocks.ReportTableBlock(sess, blocks.ReportTableConfig{
				Columns:        plColumns(),
				GroupBy:        []string{"section", "account_type"},
				ShowSubtotals:  true,
				ShowGrandTotal: true,
				Exportable:     true,
			}),
			blocks.ReportChartBlock(sess, blocks.ReportChartConfig{
				Type:   blocks.ChartTypeBar,
				Series: []string{"income", "expenses", "net_profit"},
			}),
		},
	}
}

func plColumns() []blocks.ReportColumnDef {
	return []blocks.ReportColumnDef{
		{Name: "account_name", Label: "Account", Type: "text"},
		{Name: "current_period", Label: "Current Period", Type: "currency"},
		{Name: "prior_period", Label: "Prior Period", Type: "currency"},
		{Name: "variance", Label: "Variance", Type: "currency"},
		{Name: "variance_pct", Label: "Var %", Type: "number"},
	}
}
