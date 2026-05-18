package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// ChartType defines the ECharts chart type for report charts.
type ChartType string

const (
	ChartTypeBar  ChartType = "bar"
	ChartTypeLine ChartType = "line"
	ChartTypePie  ChartType = "pie"
)

// ReportChartConfig configures the report chart.
type ReportChartConfig struct {
	Type   ChartType
	Series []string
	Height int
}

// ReportChartBlock renders the chart section of a financial report.
func ReportChartBlock(_ ui.UISessionContext, cfg ReportChartConfig) ast.Node {
	chartType := string(cfg.Type)
	if chartType == "" {
		chartType = "bar"
	}
	height := cfg.Height
	if height == 0 {
		height = 300
	}
	return ast.ChartNode{
		Config: map[string]any{
			"type": chartType,
		},
		Height: height,
	}
}
