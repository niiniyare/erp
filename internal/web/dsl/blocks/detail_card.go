package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// FieldDef defines one field in a DetailCardBlock.
type FieldDef struct {
	Label  string
	Key    string
	Format string // ""|"currency"|"date"|"percent"
}

// DetailCardConfig configures a read-only field group card.
type DetailCardConfig struct {
	Title  string
	Fields []FieldDef
	// Column controls how many label-value pairs appear per row (default: 3).
	Column int
}

// fieldContent converts a FieldDef to an AMIS expression string using format filters.
func fieldContent(f FieldDef) string {
	switch f.Format {
	case "currency":
		// number filter adds thousand separators and decimal places.
		return "${" + f.Key + "|number}"
	case "date":
		return "${" + f.Key + "|date:YYYY-MM-DD}"
	case "percent":
		return "${" + f.Key + "|percent}"
	default:
		return "${" + f.Key + "}"
	}
}

// DetailCardBlock renders a read-only field group (label: value pairs).
// Emits an AMIS property node — a description list, not a table.
func DetailCardBlock(_ ui.UISessionContext, cfg DetailCardConfig) ast.Node {
	items := make([]ast.PropertyItem, 0, len(cfg.Fields))
	for _, f := range cfg.Fields {
		items = append(items, ast.PropertyItem{
			Label:   f.Label,
			Content: fieldContent(f),
		})
	}
	return ast.CardNode{
		Title: cfg.Title,
		Body: []ast.Node{
			ast.PropertyNode{
				Column: cfg.Column,
				Items:  items,
			},
		},
	}
}
