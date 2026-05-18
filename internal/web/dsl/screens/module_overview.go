package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// ModuleOverviewConfig configures a per-module landing page.
type ModuleOverviewConfig struct {
	Title            string
	Description      string
	KPIs             []blocks.StatCardConfig
	Actions          []blocks.QuickAction
	ActivityResource string
}

// ModuleOverviewScreen renders a generic module landing page with KPIs, quick actions,
// and recent activity. Covers HR, inventory, tenant admin, and other module entry points.
func ModuleOverviewScreen(sess ui.UISessionContext, cfg ModuleOverviewConfig) ast.Node {
	body := []ast.Node{}
	if len(cfg.KPIs) > 0 {
		body = append(body, blocks.KPIRowBlock(sess, cfg.KPIs))
	}
	if len(cfg.Actions) > 0 {
		body = append(body, blocks.QuickActionsBlock(sess, cfg.Actions))
	}
	if cfg.ActivityResource != "" {
		body = append(body, blocks.ActivityPanelBlock(sess, blocks.ActivityPanelConfig{
			Title: "Recent Activity", Resource: cfg.ActivityResource, Limit: 10,
		}))
	}
	if len(body) == 0 {
		body = append(body, blocks.ActivityPanelBlock(sess, blocks.ActivityPanelConfig{Title: "Recent Activity", Limit: 5}))
	}
	return ast.PageNode{Title: cfg.Title, Body: body}
}
