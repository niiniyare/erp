package blocks

import (
	"fmt"

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

// colorClass converts a StatusMapping Color to a Bootstrap badge CSS class.
func colorClass(color string) string {
	switch color {
	case "success":
		return "badge-success"
	case "warning":
		return "badge-warning"
	case "danger":
		return "badge-danger"
	case "info":
		return "badge-info"
	default:
		return "badge-default"
	}
}

// buildMap converts StatusMappings to the AMIS mapping map[string]string.
// Each value maps to an HTML badge span. A "*" fallback renders the raw value.
func buildMap(mappings []StatusMapping) map[string]string {
	m := make(map[string]string, len(mappings)+1)
	for _, sm := range mappings {
		m[sm.Value] = fmt.Sprintf(
			`<span class="badge %s">%s</span>`,
			colorClass(sm.Color), sm.Label,
		)
	}
	// Fallback: render raw value when no mapping matches.
	m["*"] = `<span class="badge badge-default">${value}</span>`
	return m
}

// StatusBadgeColumnDef returns a ColumnDef for use in DataTableConfig.Columns.
// Equivalent to StatusBadgeColumn but fits DataTableBlock's Columns slice type.
func StatusBadgeColumnDef(cfg StatusBadgeConfig) ColumnDef {
	return ColumnDef{
		Name:  cfg.FieldName,
		Label: cfg.Label,
		Type:  "mapping",
		Map:   buildMap(cfg.Mappings),
	}
}

// StatusBadgeColumn returns a TableColumn configured as a colour-mapped status badge.
// Use inside DataTableConfig.Columns rather than creating raw TableColumns.
func StatusBadgeColumn(cfg StatusBadgeConfig) ast.TableColumn {
	return ast.TableColumn{
		Name:  cfg.FieldName,
		Label: cfg.Label,
		Type:  "mapping",
		Map:   buildMap(cfg.Mappings),
	}
}

// StatusBadgeBlock renders a standalone status indicator (not in a table).
// Emits an AMIS mapping node — colour-coded badge, read-only display.
func StatusBadgeBlock(_ ui.UISessionContext, cfg StatusBadgeConfig) ast.Node {
	return ast.MappingNode{
		Name:  cfg.FieldName,
		Label: cfg.Label,
		Map:   buildMap(cfg.Mappings),
	}
}
