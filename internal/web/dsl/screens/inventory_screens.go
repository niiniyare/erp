package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// StockLedgerScreen renders the paginated stock movements listing.
func StockLedgerScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Stock Ledger",
		Body: []ast.Node{
			ast.CRUDNode{
				API: ast.APISpec{Method: "get", URL: "/api/v1/inventory/stock-movements"},
				Filter: ast.FilterBarNode{
					Body: []ast.Node{
						ast.SelectNode{
							Name:       "warehouse_id",
							Label:      "Warehouse",
							Source:     &ast.APISpec{Method: "get", URL: "/api/v1/inventory/warehouses/options"},
							Searchable: true,
						},
						ast.SelectNode{
							Name:       "item_id",
							Label:      "Item",
							Source:     &ast.APISpec{Method: "get", URL: "/api/v1/inventory/items/options"},
							Searchable: true,
						},
						ast.InputDateRangeNode{Name: "date_range", Label: "Date Range"},
					},
				},
				Columns: []ast.TableColumn{
					{Name: "date", Label: "Date", Type: "date", Sortable: true},
					{Name: "item_code", Label: "Item Code"},
					{Name: "item_name", Label: "Item"},
					{Name: "warehouse", Label: "Warehouse"},
					{Name: "movement_type", Label: "Type", Type: "mapping", Map: stockMovementTypeMap()},
					{Name: "qty", Label: "Qty", Type: "number"},
					{Name: "balance", Label: "Balance", Type: "number"},
					{Name: "reference", Label: "Reference"},
				},
				PageSize: 50,
			},
		},
	}
}

func stockMovementTypeMap() map[string]string {
	return map[string]string{
		"in":       `<span class="label label-success">IN</span>`,
		"out":      `<span class="label label-danger">OUT</span>`,
		"transfer": `<span class="label label-info">TRANSFER</span>`,
		"*":        `${value}`,
	}
}

// WarehouseScreen renders the warehouse listing and current stock levels.
func WarehouseScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Warehouses",
		Body: []ast.Node{
			ast.CRUDNode{
				API: ast.APISpec{Method: "get", URL: "/api/v1/inventory/warehouses"},
				Toolbar: []ast.Node{
					ast.ActionNode{
						Label:      "New Warehouse",
						ActionType: "link",
						Target:     "/inventory/warehouses/new",
						Level:      "primary",
						VisibleOn:  `${permissions["inventory.warehouses.create"]}`,
					},
				},
				Columns: []ast.TableColumn{
					{Name: "code", Label: "Code", Sortable: true},
					{Name: "name", Label: "Warehouse Name", Sortable: true},
					{Name: "location", Label: "Location"},
					{Name: "total_items", Label: "Items", Type: "number"},
					{Name: "total_value", Label: "Stock Value", Type: "number"},
					{Name: "status", Label: "Status", Type: "mapping", Map: map[string]string{
						"active":   `<span class="label label-success">Active</span>`,
						"inactive": `<span class="label label-default">Inactive</span>`,
						"*":        `${value}`,
					}},
				},
				RowActions: []ast.ActionNode{
					{Label: "View Stock", ActionType: "link", Target: "/inventory/stock-ledger?warehouse_id=${id}"},
				},
				PageSize: 20,
			},
		},
	}
}

// ItemCatalogScreen renders the item master data listing.
func ItemCatalogScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Item Catalog",
		Body: []ast.Node{
			blocks.KPIRowBlock(sess, []blocks.StatCardConfig{
				{Label: "Total Items", ValueKey: "total_items", Format: "number"},
				{Label: "Active Items", ValueKey: "active_items", Format: "number"},
				{Label: "Low Stock", ValueKey: "low_stock_count", Format: "number", Trend: blocks.TrendDown},
			}),
			ast.CRUDNode{
				API: ast.APISpec{Method: "get", URL: "/api/v1/inventory/items"},
				Filter: ast.FilterBarNode{
					Body: []ast.Node{
						ast.InputTextNode{Name: "search", Label: "Search", Placeholder: "Code or name…"},
						ast.SelectNode{
							Name:   "category_id",
							Label:  "Category",
							Source: &ast.APISpec{Method: "get", URL: "/api/v1/inventory/item-categories/options"},
						},
					},
				},
				Toolbar: []ast.Node{
					ast.ActionNode{
						Label:      "New Item",
						ActionType: "link",
						Target:     "/inventory/items/new",
						Level:      "primary",
						VisibleOn:  `${permissions["inventory.items.create"]}`,
					},
				},
				Columns: []ast.TableColumn{
					{Name: "code", Label: "Code", Sortable: true},
					{Name: "name", Label: "Item Name", Sortable: true},
					{Name: "category", Label: "Category"},
					{Name: "uom", Label: "UOM"},
					{Name: "reorder_point", Label: "Reorder Pt", Type: "number"},
					{Name: "current_stock", Label: "Stock", Type: "number"},
					{Name: "unit_cost", Label: "Unit Cost", Type: "number"},
				},
				RowActions: []ast.ActionNode{
					{Label: "Edit", ActionType: "link", Target: "/inventory/items/${id}"},
					{Label: "Stock Movements", ActionType: "link", Target: "/inventory/stock-ledger?item_id=${id}"},
				},
				PageSize: 30,
			},
		},
	}
}
