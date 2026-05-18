package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// ChartPanelConfig configures a titled chart widget for dashboards.
type ChartPanelConfig struct {
	Title        string
	ChartType    ChartType
	APIURL       string
	PeriodPicker bool
	Height       int
}

// ChartPanelBlock renders a card-wrapped chart with optional period picker.
func ChartPanelBlock(_ ui.UISessionContext, cfg ChartPanelConfig) ast.Node {
	chartType := string(cfg.ChartType)
	if chartType == "" {
		chartType = "bar"
	}
	height := cfg.Height
	if height == 0 {
		height = 300
	}
	var body []ast.Node
	if cfg.PeriodPicker {
		body = append(body, ast.SelectNode{
			Name:  "chart_period",
			Label: "Period",
			Options: []ast.SelectOption{
				{Label: "This Month", Value: "month"},
				{Label: "This Quarter", Value: "quarter"},
				{Label: "This Year", Value: "year"},
			},
			DefaultValue: "month",
		})
	}
	var api *ast.APISpec
	if cfg.APIURL != "" {
		api = &ast.APISpec{Method: "get", URL: cfg.APIURL}
	}
	body = append(body, ast.ChartNode{
		Config: map[string]any{"type": chartType},
		API:    api,
		Height: height,
	})
	return ast.CardNode{Title: cfg.Title, Body: body}
}
