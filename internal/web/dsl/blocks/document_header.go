package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// DocumentHeaderConfig controls which fields appear in the document header section.
type DocumentHeaderConfig struct {
	ResourceURL   string // e.g. "/api/v1/finance/invoices"
	ShowCurrency  bool
	ShowStatus    bool
	StatusOptions []ast.SelectOption
	ReadOnly      bool
}

// DocumentHeaderBlock renders the top section of any document form:
// reference number, date, and optional currency/status fields.
func DocumentHeaderBlock(sess ui.UISessionContext, cfg DocumentHeaderConfig) ast.Node {
	fields := []ast.Node{
		ast.InputTextNode{Name: "ref_number", Label: "Reference #", Required: true, DisabledOn: boolExpr(cfg.ReadOnly)},
		ast.InputDateNode{Name: "document_date", Label: "Date", Required: true, DisabledOn: boolExpr(cfg.ReadOnly)},
		ast.InputDateNode{Name: "due_date", Label: "Due Date", DisabledOn: boolExpr(cfg.ReadOnly)},
	}
	if cfg.ShowCurrency {
		currency := sess.Currency
		if currency == "" {
			currency = "USD"
		}
		fields = append(fields, ast.SelectNode{
			Name:         "currency",
			Label:        "Currency",
			Source:       &ast.APISpec{Method: "get", URL: "/api/v1/platform/currencies/options"},
			DefaultValue: currency,
			DisabledOn:   boolExpr(cfg.ReadOnly),
		})
	}
	if cfg.ShowStatus && len(cfg.StatusOptions) > 0 {
		fields = append(fields, ast.SelectNode{
			Name:    "status",
			Label:   "Status",
			Options: cfg.StatusOptions,
		})
	}
	return ast.SectionNode{
		Title: "Document Details",
		Body:  fields,
	}
}
