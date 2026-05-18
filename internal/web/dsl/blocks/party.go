package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// PartyConfig controls which party type to show and the API endpoint.
type PartyConfig struct {
	Label      string // "Customer" | "Supplier" | "Employee"
	FieldName  string // "customer_id" | "supplier_id" | "employee_id"
	OptionsURL string // API for party picker options
	Required   bool
	ReadOnly   bool
}

// DefaultCustomerConfig returns party config for customer selection.
func DefaultCustomerConfig() PartyConfig {
	return PartyConfig{Label: "Customer", FieldName: "customer_id", OptionsURL: "/api/v1/crm/customers/options", Required: true}
}

// DefaultSupplierConfig returns party config for supplier selection.
func DefaultSupplierConfig() PartyConfig {
	return PartyConfig{Label: "Supplier", FieldName: "supplier_id", OptionsURL: "/api/v1/procurement/suppliers/options", Required: true}
}

// DefaultEmployeeConfig returns party config for employee selection.
func DefaultEmployeeConfig() PartyConfig {
	return PartyConfig{Label: "Employee", FieldName: "employee_id", OptionsURL: "/api/v1/hr/employees/options", Required: true}
}

// PartyBlock renders the party (customer/supplier/employee) selector section.
func PartyBlock(sess ui.UISessionContext, cfg PartyConfig) ast.Node {
	label := cfg.Label
	if label == "" {
		label = "Party"
	}
	return ast.SectionNode{
		Title: label + " Details",
		Body: []ast.Node{
			ast.SelectNode{
				Name:       cfg.FieldName,
				Label:      label,
				Required:   cfg.Required,
				Source:     &ast.APISpec{Method: "get", URL: cfg.OptionsURL},
				Searchable: true,
				DisabledOn: boolExpr(cfg.ReadOnly),
			},
		},
	}
}
