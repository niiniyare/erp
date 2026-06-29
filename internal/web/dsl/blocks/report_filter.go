package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// ReportFilterConfig controls which filter inputs appear on report pages.
type ReportFilterConfig struct {
	ShowPeriod       bool
	ShowEntityPicker bool
	EntityURL        string
	ShowCurrency     bool
	ShowComparison   bool
}

// ReportFilterBlock renders the filter form for financial reports.
func ReportFilterBlock(_ ui.UISessionContext, cfg ReportFilterConfig) ast.Node {
	var fields []ast.Node
	if cfg.ShowPeriod {
		fields = append(fields, ast.InputDateRangeNode{Name: "period", Label: "Period", Required: true})
	}
	if cfg.ShowEntityPicker && cfg.EntityURL != "" {
		fields = append(fields, ast.SelectNode{
			Name:       "entity_id",
			Label:      "Entity",
			Source:     &ast.APISpec{Method: "get", URL: cfg.EntityURL},
			Clearable:  true,
			Searchable: true,
		})
	}
	if cfg.ShowCurrency {
		fields = append(fields, ast.SelectNode{
			Name:      "currency",
			Label:     "Currency",
			Source:    &ast.APISpec{Method: "get", URL: "/api/v1/platform/currencies/options"},
			Clearable: true,
		})
	}
	if cfg.ShowComparison {
		fields = append(fields, ast.SelectNode{
			Name:  "compare_period",
			Label: "Compare With",
			Options: []ast.SelectOption{
				{Label: "Prior Period", Value: "prior"},
				{Label: "Prior Year", Value: "prior_year"},
				{Label: "None", Value: "none"},
			},
			DefaultValue: "none",
		})
	}
	if len(fields) == 0 {
		fields = append(fields, ast.InputDateRangeNode{Name: "period", Label: "Period", Required: true})
	}
	return ast.FilterBarNode{Body: fields}
}
