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
	// ShowSubtotal renders a computed subtotal column (qty × unit_price × discount).
	// Use for commercial documents (invoice, bill, PO, SO).
	ShowSubtotal bool
	// ShowDebit renders a debit amount input. Use for journal entry lines only.
	ShowDebit bool
	// ShowCredit renders a credit amount input. Use for journal entry lines only.
	ShowCredit        bool
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

// JournalLineItemConfig returns config for journal entries.
// Journal lines use explicit debit/credit fields — not subtotal.
// Server validates that sum(debit) == sum(credit) before posting.
func JournalLineItemConfig() LineItemConfig {
	return LineItemConfig{
		ShowDescription: true,
		ShowAccount:     true,
		ShowDebit:       true,
		ShowCredit:      true,
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
			Name:   "uom",
			Label:  "UOM",
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
		// FormulaNode writes the computed value into "subtotal".
		// The disabled InputNumberNode reads the same field for display.
		// Condition guards against computing before required fields are entered.
		fields = append(fields,
			ast.FormulaNode{
				Name:      "subtotal",
				Formula:   "qty * unit_price * (1 - (discount_pct || 0) / 100)",
				InitSet:   true,
				Condition: "${qty && unit_price}",
			},
			ast.InputNumberNode{Name: "subtotal", Label: "Subtotal", Precision: 2, DisabledOn: "true"},
		)
	}
	if cfg.ShowDebit {
		fields = append(fields, ast.InputNumberNode{
			Name:       "debit",
			Label:      "Debit",
			Precision:  2,
			DisabledOn: boolExpr(cfg.ReadOnly),
		})
	}
	if cfg.ShowCredit {
		fields = append(fields, ast.InputNumberNode{
			Name:       "credit",
			Label:      "Credit",
			Precision:  2,
			DisabledOn: boolExpr(cfg.ReadOnly),
		})
	}
	return fields
}
