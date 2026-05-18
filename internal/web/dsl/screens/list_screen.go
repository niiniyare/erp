package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// ListScreenConfig configures the generic listing screen.
type ListScreenConfig struct {
	Title          string
	Resource       string
	APIURL         string
	Columns        []blocks.ColumnDef
	FilterBar      blocks.FilterBarConfig
	AllowCreate    bool
	CreateURL      string
	AllowExport    bool
	BulkActions    []blocks.BulkActionDef
	RowClickAction string
}

// GenericListScreen composes FilterBarBlock + DataTableBlock into a listing page.
// Covers invoice list, customer list, product list, transaction history, audit log, and more.
func GenericListScreen(sess ui.UISessionContext, cfg ListScreenConfig) ast.Node {
	filter := blocks.FilterBarBlock(sess, cfg.FilterBar)
	table := blocks.DataTableBlock(sess, blocks.DataTableConfig{
		APIURL:      cfg.APIURL,
		Columns:     cfg.Columns,
		Filter:      filter,
		BulkActions: cfg.BulkActions,
		AllowCreate: cfg.AllowCreate,
		CreateURL:   cfg.CreateURL,
		AllowExport: cfg.AllowExport,
	})
	return ast.PageNode{Title: cfg.Title, Body: []ast.Node{table}}
}
