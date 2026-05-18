package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// KPIRowBlock renders a horizontal row of StatCard KPIs.
// Each card gets an equal-width grid column.
func KPIRowBlock(sess ui.UISessionContext, cards []StatCardConfig) ast.Node {
	if len(cards) == 0 {
		return ast.FlexNode{Items: []ast.Node{ast.StatNode{Label: "—", ValueKey: "_empty"}}}
	}
	md := 12 / len(cards)
	if md < 1 {
		md = 1
	}
	cols := make([]ast.GridColumn, 0, len(cards))
	for _, c := range cards {
		cols = append(cols, ast.GridColumn{
			Body: []ast.Node{StatCardBlock(sess, c)},
			MD:   md,
		})
	}
	return ast.GridNode{Columns: cols}
}
