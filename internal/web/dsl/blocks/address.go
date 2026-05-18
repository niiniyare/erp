package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// AddressConfig controls which address sections to render.
type AddressConfig struct {
	ShowBilling  bool
	ShowShipping bool
	ReadOnly     bool
}

// AddressBlock renders billing and/or shipping address sections.
func AddressBlock(sess ui.UISessionContext, cfg AddressConfig) ast.Node {
	var sections []ast.Node
	if cfg.ShowBilling {
		sections = append(sections, addressSection("Billing Address", "billing", cfg.ReadOnly))
	}
	if cfg.ShowShipping {
		sections = append(sections, addressSection("Shipping Address", "shipping", cfg.ReadOnly))
	}
	if len(sections) == 0 {
		sections = append(sections, addressSection("Address", "", cfg.ReadOnly))
	}
	return ast.SectionNode{Title: "Address", Body: sections}
}

func addressSection(title, prefix string, readOnly bool) ast.Node {
	p := ""
	if prefix != "" {
		p = prefix + "_"
	}
	return ast.SectionNode{
		Title: title,
		Body: []ast.Node{
			ast.InputTextNode{Name: p + "street", Label: "Street", DisabledOn: boolExpr(readOnly)},
			ast.InputTextNode{Name: p + "city", Label: "City", DisabledOn: boolExpr(readOnly)},
			ast.InputTextNode{Name: p + "state", Label: "State/Province", DisabledOn: boolExpr(readOnly)},
			ast.InputTextNode{Name: p + "postal_code", Label: "Postal Code", DisabledOn: boolExpr(readOnly)},
			ast.SelectNode{
				Name:       p + "country",
				Label:      "Country",
				Source:     &ast.APISpec{Method: "get", URL: "/api/v1/platform/countries/options"},
				Searchable: true,
				DisabledOn: boolExpr(readOnly),
			},
		},
	}
}
