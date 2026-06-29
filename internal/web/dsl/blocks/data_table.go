package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// ColumnDef defines one column in a DataTableBlock.
type ColumnDef struct {
	Name  string
	Label string
	Type  string // "text"|"date"|"number"|"currency"|"mapping"|"link"|"image"
	// Map holds value→HTML entries for Type "mapping" columns.
	// Prefer StatusBadgeColumn() to build this correctly.
	Map      map[string]string
	Width    int
	Sortable bool
	Fixed    string // ""|"left"|"right"
}

// BulkActionDef defines one bulk action button.
type BulkActionDef struct {
	Label      string
	Permission string // raw permission string e.g. "finance.invoices.delete"
	APIURL     string
	APIMethod  string
	Level      string // "danger"|"warning"|"default"
	Confirm    string
}

// DataTableConfig configures a DataTableBlock.
type DataTableConfig struct {
	APIURL      string
	Columns     []ColumnDef
	RowActions  []ast.ActionNode
	BulkActions []BulkActionDef
	Filter      ast.Node
	PrimaryKey  string
	PageSize    int
	Title       string
	AllowCreate bool
	CreateURL   string
	// CreatePermission is the dot-notation permission checked before showing the
	// "New" button. Format: "resource.action" e.g. "finance.invoices.create".
	// Required when AllowCreate is true — without it the button is always hidden.
	CreatePermission string
	AllowExport      bool
}

// DataTableBlock builds the paginated, filterable CRUD table for listing pages.
func DataTableBlock(sess ui.UISessionContext, cfg DataTableConfig) ast.Node {
	cols := make([]ast.TableColumn, 0, len(cfg.Columns))
	for _, c := range cfg.Columns {
		cols = append(cols, ast.TableColumn{
			Name:     c.Name,
			Label:    c.Label,
			Type:     c.Type,
			Map:      c.Map,
			Width:    c.Width,
			Sortable: c.Sortable,
			Fixed:    c.Fixed,
		})
	}

	var toolbar []ast.Node
	if cfg.AllowCreate && cfg.CreateURL != "" && (cfg.CreatePermission == "" || canPerm(sess, cfg.CreatePermission)) {
		toolbar = append(toolbar, ast.ActionNode{
			Label:      "New",
			ActionType: "link",
			Target:     cfg.CreateURL,
			Level:      "primary",
			Icon:       "fa fa-plus",
		})
	}
	if cfg.AllowExport {
		toolbar = append(toolbar, ast.ActionNode{
			Label: "Export", ActionType: "ajax", Level: "default", Icon: "fa fa-download",
			API: &ast.APISpec{Method: "get", URL: cfg.APIURL + "/export"},
		})
	}

	var bulkNodes []ast.Node
	for _, b := range cfg.BulkActions {
		if b.Permission != "" && !canPerm(sess, b.Permission) {
			continue
		}
		method := b.APIMethod
		if method == "" {
			method = "post"
		}
		bulkNodes = append(bulkNodes, ast.ActionNode{
			Label:       b.Label,
			ActionType:  "ajax",
			Level:       b.Level,
			ConfirmText: b.Confirm,
			API:         &ast.APISpec{Method: method, URL: b.APIURL},
		})
	}

	return ast.CRUDNode{
		API:         ast.APISpec{Method: "get", URL: cfg.APIURL},
		Columns:     cols,
		Filter:      cfg.Filter,
		Toolbar:     toolbar,
		BulkActions: bulkNodes,
		RowActions:  cfg.RowActions,
		PrimaryKey:  cfg.PrimaryKey,
		PageSize:    cfg.PageSize,
		Title:       cfg.Title,
	}
}
