package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// LineItemConfig controls columns visible on the line item combo per document type.
type LineItemConfig struct {
	ShowProductCode   bool
	ShowDescription   bool
	ShowQty           bool
	ShowUnitOfMeasure bool
	ShowUnitPrice     bool
	ShowDiscount      bool
	ShowTaxRate       bool
	ShowSubtotal      bool
	ShowAccount       bool
	AllowFreeTextItem bool
	DefaultCurrency   string
	MaxLines          int
	ReadOnly          bool
}

// DefaultLineItemConfig returns config for standard commercial documents (invoice, PO, bill).
func DefaultLineItemConfig() LineItemConfig {
	return LineItemConfig{
		ShowProductCode:   true,
		ShowDescription:   true,
		ShowQty:           true,
		ShowUnitOfMeasure: true,
		ShowUnitPrice:     true,
		ShowSubtotal:      true,
	}
}

// GRNLineItemConfig returns config for goods receipt notes (no tax, no price editing).
func GRNLineItemConfig() LineItemConfig {
	return LineItemConfig{
		ShowProductCode:   true,
		ShowDescription:   true,
		ShowQty:           true,
		ShowUnitOfMeasure: true,
	}
}

// JournalLineItemConfig returns config for journal entries (account + debit/credit).
func JournalLineItemConfig() LineItemConfig {
	return LineItemConfig{
		ShowDescription: true,
		ShowAccount:     true,
		ShowSubtotal:    true,
	}
}

// ProductServiceLineBlock returns the shared line items combo node.
// Every document form with line items MUST use this. Never define a custom
// line item table in a screen file.
func ProductServiceLineBlock(sess ui.UISessionContext, cfg LineItemConfig) ast.Node {
	items := buildLineItemFields(cfg)
	combo := ast.ComboNode{
		Name:           "line_items",
		Label:          "Line Items",
		Items:          items,
		Multiple:       true,
		AddButtonLabel: "Add Line",
		DisabledOn:     boolExpr(cfg.ReadOnly),
	}
	if cfg.MaxLines > 0 {
		combo.MaxLength = cfg.MaxLines
	}
	return combo
}

func buildLineItemFields(cfg LineItemConfig) []ast.Node {
	var fields []ast.Node
	if cfg.ShowProductCode {
		fields = append(fields, ast.InputTextNode{Name: "product_code", Label: "Code", DisabledOn: boolExpr(cfg.ReadOnly)})
	}
	if cfg.ShowDescription {
		fields = append(fields, ast.InputTextNode{Name: "description", Label: "Description", Required: true, DisabledOn: boolExpr(cfg.ReadOnly)})
	}
	if cfg.ShowAccount {
		fields = append(fields, ast.SelectNode{
			Name: "account_id", Label: "Account", Required: true,
			Source: &ast.APISpec{Method: "get", URL: "/api/v1/finance/accounts/options"},
		})
	}
	if cfg.ShowQty {
		fields = append(fields, ast.InputNumberNode{Name: "qty", Label: "Qty", Required: true, DisabledOn: boolExpr(cfg.ReadOnly)})
	}
	if cfg.ShowUnitOfMeasure {
		fields = append(fields, ast.SelectNode{
			Name: "uom", Label: "UOM",
			Source: &ast.APISpec{Method: "get", URL: "/api/v1/inventory/uom/options"},
		})
	}
	if cfg.ShowUnitPrice {
		fields = append(fields, ast.InputNumberNode{Name: "unit_price", Label: "Unit Price", Required: true, Precision: 2, DisabledOn: boolExpr(cfg.ReadOnly)})
	}
	if cfg.ShowDiscount {
		fields = append(fields, ast.InputNumberNode{Name: "discount_pct", Label: "Disc %", Precision: 2})
	}
	if cfg.ShowTaxRate {
		fields = append(fields, ast.SelectNode{
			Name:   "tax_rate_id",
			Label:  "Tax",
			Source: &ast.APISpec{Method: "get", URL: "/api/v1/finance/tax-rates/options"},
		})
	}
	if cfg.ShowSubtotal {
		fields = append(fields, ast.InputNumberNode{Name: "subtotal", Label: "Subtotal", Precision: 2, DisabledOn: "true"})
	}
	return fields
}
