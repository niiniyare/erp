package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// FinanceDashboardScreen renders the finance module dashboard.
func FinanceDashboardScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Finance Dashboard",
		Body: []ast.Node{
			blocks.KPIRowBlock(sess, []blocks.StatCardConfig{
				{Label: "Total Revenue", ValueKey: "revenue", Format: "currency", Trend: blocks.TrendUp},
				{Label: "Outstanding AR", ValueKey: "ar_balance", Format: "currency", Trend: blocks.TrendDown},
				{Label: "Outstanding AP", ValueKey: "ap_balance", Format: "currency", Trend: blocks.TrendDown},
				{Label: "Cash & Bank", ValueKey: "cash_balance", Format: "currency"},
			}),
			ast.GridNode{Columns: []ast.GridColumn{
				{Body: []ast.Node{blocks.ChartPanelBlock(sess, blocks.ChartPanelConfig{
					Title: "Revenue vs Expenses", ChartType: blocks.ChartTypeLine, PeriodPicker: true,
					APIURL: "/api/v1/finance/dashboard/revenue-chart",
				})}, MD: 8},
				{Body: []ast.Node{blocks.ActivityPanelBlock(sess, blocks.ActivityPanelConfig{
					Title: "Recent Transactions", Resource: "finance/transactions", Limit: 10,
				})}, MD: 4},
			}},
			blocks.QuickActionsBlock(sess, []blocks.QuickAction{
				{Label: "New Invoice", URL: "/finance/invoices/new", Permission: "finance.invoices.create", Icon: "fa fa-file-invoice"},
				{Label: "Record Payment", URL: "/finance/payments/new", Permission: "finance.payments.create", Icon: "fa fa-credit-card"},
				{Label: "View Reports", URL: "/finance/reports", Permission: "finance.reports.read", Icon: "fa fa-chart-bar"},
			}),
		},
	}
}
