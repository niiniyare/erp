package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// InternalNotesBlock renders a collapsed rich-text notes section.
// Internal notes are not customer-visible and are gated on write permission.
func InternalNotesBlock(sess ui.UISessionContext) ast.Node {
	return ast.SectionNode{
		Title:     "Internal Notes",
		Collapsed: true,
		Body: []ast.Node{
			ast.InputTextNode{
				Name:        "internal_notes",
				Label:       "",
				Placeholder: "Add internal notes…",
				MaxLength:   2000,
			},
		},
	}
}
