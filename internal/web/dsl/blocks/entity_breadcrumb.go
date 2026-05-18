package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// EntityBreadcrumbBlock renders a hierarchy path for a given entity.
// The breadcrumb data is expected in ${breadcrumbs} as an array of {label, url}.
func EntityBreadcrumbBlock(_ ui.UISessionContext) ast.Node {
	return ast.StatNode{
		Label:    "breadcrumb",
		ValueKey: "breadcrumbs",
	}
}
