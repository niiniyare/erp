package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// StatusMapping maps a status value to a display label and colour.
type StatusMapping struct {
	Value string
	Label string
	Color string // "success"|"warning"|"danger"|"info"|"default"
}

// StatusBadgeConfig configures a status badge column or display.
type StatusBadgeConfig struct {
	FieldName string
	Label     string
	Mappings  []StatusMapping
}

// StatusBadgeColumn returns a TableColumn configured as a colour-mapped status badge.
// Use this inside DataTableConfig.Columns rather than creating raw TableColumns.
func StatusBadgeColumn(cfg StatusBadgeConfig) ast.TableColumn {
	return ast.TableColumn{
		Name:  cfg.FieldName,
		Label: cfg.Label,
		Type:  "status",
	}
}

// StatusBadgeBlock renders a standalone status indicator (not in a table).
func StatusBadgeBlock(_ ui.UISessionContext, cfg StatusBadgeConfig) ast.Node {
	opts := make([]ast.SelectOption, 0, len(cfg.Mappings))
	for _, m := range cfg.Mappings {
		opts = append(opts, ast.SelectOption{Label: m.Label, Value: m.Value})
	}
	return ast.SelectNode{
		Name:       cfg.FieldName,
		Label:      cfg.Label,
		Options:    opts,
		DisabledOn: "true",
	}
}
