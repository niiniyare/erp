package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// ReportColumnDef defines one column in a ReportTableBlock.
type ReportColumnDef struct {
	Name     string
	Label    string
	Type     string // "text"|"number"|"currency"
	Sortable bool
}

// ReportTableConfig configures the financial report data table.
type ReportTableConfig struct {
	Columns        []ReportColumnDef
	GroupBy        []string
	ShowSubtotals  bool
	ShowGrandTotal bool
	Exportable     bool
	Currency       string
}

// ReportTableBlock renders the data table for financial reports with grouping and totals.
func ReportTableBlock(sess ui.UISessionContext, cfg ReportTableConfig) ast.Node {
	cols := make([]ast.TableColumn, 0, len(cfg.Columns))
	for _, c := range cfg.Columns {
		cols = append(cols, ast.TableColumn{Name: c.Name, Label: c.Label, Type: c.Type, Sortable: c.Sortable})
	}
	var toolbar []ast.Node
	if cfg.Exportable {
		toolbar = append(toolbar,
			ast.ActionNode{Label: "CSV", ActionType: "ajax", Level: "default", API: &ast.APISpec{Method: "get", URL: "${api_url}/export?format=csv"}},
			ast.ActionNode{Label: "PDF", ActionType: "ajax", Level: "default", API: &ast.APISpec{Method: "get", URL: "${api_url}/export?format=pdf"}},
		)
	}
	return ast.CRUDNode{
		API:     ast.APISpec{Method: "post", URL: "${report_api_url}"},
		Columns: cols,
		Toolbar: toolbar,
	}
}
