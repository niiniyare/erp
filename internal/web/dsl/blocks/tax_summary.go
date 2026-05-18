package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// TaxSummaryBlock renders a read-only card summarising tax breakdown by rate.
// Sourced from the document's computed tax lines — no API call needed at render.
func TaxSummaryBlock(sess ui.UISessionContext) ast.Node {
	return ast.CardNode{
		Title: "Tax Summary",
		Body: []ast.Node{
			ast.TableNode{
				Source: "${tax_lines}",
				Columns: []ast.TableColumn{
					{Name: "tax_name", Label: "Tax"},
					{Name: "taxable_amount", Label: "Taxable Amount", Type: "number"},
					{Name: "tax_amount", Label: "Tax Amount", Type: "number"},
					{Name: "rate_pct", Label: "Rate %", Type: "number"},
				},
			},
		},
	}
}
