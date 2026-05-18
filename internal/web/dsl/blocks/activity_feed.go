package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// ActivityFeedConfig configures the activity timeline.
type ActivityFeedConfig struct {
	Title        string
	Resource     string
	ResourceID   string
	ShowComments bool
	Limit        int
}

// ActivityFeedBlock renders a chronological list of events for a document or resource.
func ActivityFeedBlock(sess ui.UISessionContext, cfg ActivityFeedConfig) ast.Node {
	url := "/api/v1/audit/activity"
	if cfg.Resource != "" {
		url = "/api/v1/audit/" + cfg.Resource + "/activity"
	}
	title := cfg.Title
	if title == "" {
		title = "Activity"
	}
	return ast.CardNode{
		Title: title,
		Body: []ast.Node{
			ast.TimelineNode{
				API: &ast.APISpec{Method: "get", URL: url},
			},
		},
	}
}
