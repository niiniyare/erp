package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// ActivityPanelConfig configures the dashboard activity panel.
type ActivityPanelConfig struct {
	Title    string
	Resource string // "transactions"|"invoices"|"events"
	Limit    int
}

// ActivityPanelBlock renders a card with a recent-activity timeline.
func ActivityPanelBlock(_ ui.UISessionContext, cfg ActivityPanelConfig) ast.Node {
	title := cfg.Title
	if title == "" {
		title = "Recent Activity"
	}
	url := "/api/v1/audit/recent"
	if cfg.Resource != "" {
		url = "/api/v1/" + cfg.Resource + "/recent"
	}
	return ast.CardNode{
		Title: title,
		Body: []ast.Node{
			ast.TimelineNode{API: &ast.APISpec{Method: "get", URL: url}},
		},
	}
}
