package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// FilterBarConfig controls which filter inputs appear in the listing filter bar.
type FilterBarConfig struct {
	ShowDateRange     bool
	ShowStatus        bool
	StatusOptions     []ast.SelectOption
	ShowEntityPicker  bool
	EntityURL         string // API URL for entity picker options
	EntityFieldName   string // field name for entity filter
	EntityLabel       string
	ShowAmountRange   bool
	ShowCurrency      bool
	ShowTypeFilter    bool
	TypeOptions       []ast.SelectOption
	TypeFieldName     string
	TypeLabel         string
	ShowSearch        bool
	SearchPlaceholder string
	Collapsible       bool
}

// FilterBarBlock builds the filter header used on every listing page.
// Never write a custom filter form in a screen file.
func FilterBarBlock(sess ui.UISessionContext, cfg FilterBarConfig) ast.Node {
	var fields []ast.Node

	if cfg.ShowSearch {
		ph := cfg.SearchPlaceholder
		if ph == "" {
			ph = "Search…"
		}
		fields = append(fields, ast.InputTextNode{Name: "keywords", Label: "", Placeholder: ph, Trim: true})
	}
	if cfg.ShowDateRange {
		fields = append(fields, ast.InputDateRangeNode{Name: "date_range", Label: "Date Range"})
	}
	if cfg.ShowStatus && len(cfg.StatusOptions) > 0 {
		fields = append(fields, ast.SelectNode{
			Name:      "status",
			Label:     "Status",
			Options:   cfg.StatusOptions,
			Clearable: true,
			Multiple:  true,
		})
	}
	if cfg.ShowEntityPicker && cfg.EntityURL != "" {
		fieldName := cfg.EntityFieldName
		if fieldName == "" {
			fieldName = "entity_id"
		}
		label := cfg.EntityLabel
		if label == "" {
			label = "Entity"
		}
		fields = append(fields, ast.SelectNode{
			Name:       fieldName,
			Label:      label,
			Source:     &ast.APISpec{Method: "get", URL: cfg.EntityURL},
			Searchable: true,
			Clearable:  true,
		})
	}
	if cfg.ShowTypeFilter && len(cfg.TypeOptions) > 0 {
		fieldName := cfg.TypeFieldName
		if fieldName == "" {
			fieldName = "type"
		}
		label := cfg.TypeLabel
		if label == "" {
			label = "Type"
		}
		fields = append(fields, ast.SelectNode{
			Name:      fieldName,
			Label:     label,
			Options:   cfg.TypeOptions,
			Clearable: true,
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
	if cfg.ShowAmountRange {
		fields = append(fields, ast.InputNumberNode{Name: "amount_min", Label: "Amount From", Precision: 2})
		fields = append(fields, ast.InputNumberNode{Name: "amount_max", Label: "Amount To", Precision: 2})
	}

	if len(fields) == 0 {
		fields = append(fields, ast.InputTextNode{Name: "keywords", Label: "", Placeholder: "Search…", Trim: true})
	}

	return ast.FilterBarNode{Body: fields, Collapsible: cfg.Collapsible, ShowCount: true}
}
