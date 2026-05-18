package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// ReportHeaderConfig configures the report header section.
type ReportHeaderConfig struct {
	Title       string
	Description string
}

// ReportHeaderBlock renders the report title, run-by, and period summary.
func ReportHeaderBlock(_ ui.UISessionContext, cfg ReportHeaderConfig) ast.Node {
	title := cfg.Title
	if title == "" {
		title = "Report"
	}
	return ast.CardNode{
		Title: title,
		Body: []ast.Node{
			ast.InputTextNode{Name: "_report_title", Label: "Report", DisabledOn: "true"},
		},
	}
}
