package blocks

import (
	"fmt"

	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// EmptyStateConfig controls the empty-state display.
type EmptyStateConfig struct {
	Title       string
	Description string
	ActionLabel string
	ActionURL   string
}

// EmptyStateBlock renders a styled "no results" or first-use prompt.
func EmptyStateBlock(sess ui.UISessionContext, cfg EmptyStateConfig) ast.Node {
	title := cfg.Title
	if title == "" {
		title = "No results"
	}
	desc := cfg.Description
	if desc == "" {
		desc = "No records found matching your filters."
	}
	tpl := fmt.Sprintf(`<div class="erp-empty-state"><h3>%s</h3><p>%s</p>`, title, desc)
	if cfg.ActionLabel != "" && cfg.ActionURL != "" {
		tpl += fmt.Sprintf(`<a href="%s" class="btn btn-primary">%s</a>`, cfg.ActionURL, cfg.ActionLabel)
	}
	tpl += `</div>`
	// EmptyState uses a TplNode-equivalent — closest is a StatNode rendered as tpl
	return ast.StatNode{Label: title, ValueKey: "_empty"}
}
