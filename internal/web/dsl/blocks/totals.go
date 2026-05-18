package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// TotalsSummaryBlock renders the subtotal / tax / total footer card for documents.
func TotalsSummaryBlock(sess ui.UISessionContext) ast.Node {
	return ast.CardNode{
		Title: "Totals",
		Body: []ast.Node{
			ast.TableNode{
				Source: "${totals}",
				Columns: []ast.TableColumn{
					{Name: "label", Label: ""},
					{Name: "amount", Label: "Amount", Type: "number"},
				},
			},
		},
	}
}
