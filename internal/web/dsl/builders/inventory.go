package builders

import "awo.so/internal/web/ast"

// ProductPickerNode returns a Select node for choosing a product or service.
func ProductPickerNode(name, label string, required bool) ast.Node {
	return ast.SelectNode{
		Name:       name,
		Label:      label,
		Required:   required,
		Source:     &ast.APISpec{Method: "get", URL: "/api/v1/inventory/products/options"},
		Searchable: true,
	}
}

// UOMPickerNode returns a Select node for choosing a unit of measure.
func UOMPickerNode(name, label string) ast.Node {
	return ast.SelectNode{
		Name:      name,
		Label:     label,
		Source:    &ast.APISpec{Method: "get", URL: "/api/v1/inventory/uom/options"},
		Clearable: true,
	}
}

// WarehousePickerNode returns a Select node for choosing a warehouse/location.
func WarehousePickerNode(name, label string) ast.Node {
	return ast.SelectNode{
		Name:   name,
		Label:  label,
		Source: &ast.APISpec{Method: "get", URL: "/api/v1/inventory/warehouses/options"},
	}
}
