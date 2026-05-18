package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// OperationsDashboardScreen renders the operations / procurement dashboard.
func OperationsDashboardScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Operations Dashboard",
		Body: []ast.Node{
			blocks.KPIRowBlock(sess, []blocks.StatCardConfig{
				{Label: "Open POs", ValueKey: "open_po_count"},
				{Label: "Pending GRNs", ValueKey: "pending_grn_count"},
				{Label: "Low Stock Items", ValueKey: "low_stock_count", Trend: blocks.TrendDown},
				{Label: "Pending Approvals", ValueKey: "pending_approvals_count", Trend: blocks.TrendDown},
			}),
			ast.GridNode{Columns: []ast.GridColumn{
				{Body: []ast.Node{blocks.ChartPanelBlock(sess, blocks.ChartPanelConfig{
					Title: "Purchase Orders by Status", ChartType: blocks.ChartTypePie,
					APIURL: "/api/v1/procurement/dashboard/po-status-chart",
				})}, MD: 6},
				{Body: []ast.Node{blocks.ActivityPanelBlock(sess, blocks.ActivityPanelConfig{
					Title: "Recent Activity", Resource: "procurement/orders", Limit: 8,
				})}, MD: 6},
			}},
			blocks.QuickActionsBlock(sess, []blocks.QuickAction{
				{Label: "New Purchase Order", URL: "/procurement/pos/new", Permission: "procurement.purchase_orders.create", Icon: "fa fa-shopping-cart"},
				{Label: "Receive Goods", URL: "/procurement/grn/new", Permission: "procurement.grn.create", Icon: "fa fa-truck"},
				{Label: "Manage Suppliers", URL: "/procurement/suppliers", Permission: "procurement.suppliers.read", Icon: "fa fa-building"},
			}),
		},
	}
}
