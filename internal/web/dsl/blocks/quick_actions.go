package blocks

import (
	"strings"

	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// QuickAction defines one shortcut button in the quick actions block.
type QuickAction struct {
	Label      string
	URL        string
	Permission string // raw permission string; empty = always visible
	Icon       string
}

// QuickActionsBlock renders permitted shortcut buttons.
// Actions whose Permission the session lacks are excluded entirely.
func QuickActionsBlock(sess ui.UISessionContext, actions []QuickAction) ast.Node {
	var nodes []ast.Node
	for _, a := range actions {
		if a.Permission != "" && !canPerm(sess, a.Permission) {
			continue
		}
		nodes = append(nodes, ast.ActionNode{
			Label:      a.Label,
			ActionType: "link",
			Target:     a.URL,
			Level:      "default",
			Icon:       a.Icon,
		})
	}
	if len(nodes) == 0 {
		nodes = []ast.Node{ast.ActionNode{Label: "Dashboard", ActionType: "link", Target: "/", Level: "default"}}
	}
	return ast.FlexNode{Items: nodes, Direction: "row", Gap: "sm", Wrap: true}
}

// canPerm checks a raw "resource.action" permission string against the session.
func canPerm(sess ui.UISessionContext, perm string) bool {
	i := strings.LastIndex(perm, ".")
	if i < 0 {
		return false
	}
	return sess.Can(perm[i+1:], perm[:i])
}

// boolExpr returns "true" when b is true (AMIS disabledOn expression), else "".
func boolExpr(b bool) string {
	if b {
		return "true"
	}
	return ""
}

// resourceFromURL extracts a coarse resource hint from a URL for permission checks.
// e.g. "/api/v1/finance/invoices/new" → "finance.invoices"
func resourceFromURL(url string) string {
	parts := strings.Split(strings.TrimPrefix(url, "/api/v1/"), "/")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return ""
}
