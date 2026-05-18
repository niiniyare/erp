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
}

// DetailCardBlock renders a read-only field group (label: value pairs).
func DetailCardBlock(sess ui.UISessionContext, cfg DetailCardConfig) ast.Node {
	cols := make([]ast.TableColumn, 0, len(cfg.Fields))
	for _, f := range cfg.Fields {
		colType := f.Format
		if colType == "" {
			colType = "text"
		}
		cols = append(cols, ast.TableColumn{Name: f.Key, Label: f.Label, Type: colType})
	}
	return ast.CardNode{
		Title: cfg.Title,
		Body: []ast.Node{
			ast.TableNode{Source: "${record}", Columns: cols},
		},
	}
}
