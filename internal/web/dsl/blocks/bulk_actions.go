package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// BulkActionsBlock returns a slice of permission-filtered bulk action nodes.
// Returns empty slice (not nil) when no permitted actions remain.
func BulkActionsBlock(sess ui.UISessionContext, defs []BulkActionDef) []ast.Node {
	nodes := make([]ast.Node, 0, len(defs))
	for _, d := range defs {
		if d.Permission != "" && !canPerm(sess, d.Permission) {
			continue
		}
		method := d.APIMethod
		if method == "" {
			method = "post"
		}
		level := d.Level
		if level == "" {
			level = "default"
		}
		nodes = append(nodes, ast.ActionNode{
			Label:       d.Label,
			ActionType:  "ajax",
			Level:       level,
			ConfirmText: d.Confirm,
			API:         &ast.APISpec{Method: method, URL: d.APIURL},
		})
	}
	return nodes
}
