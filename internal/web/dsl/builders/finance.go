package builders

import "awo.so/internal/web/ast"

// AccountPickerNode returns a Select node for choosing a GL account.
func AccountPickerNode(name, label string, required bool) ast.Node {
	return ast.SelectNode{
		Name:       name,
		Label:      label,
		Required:   required,
		Source:     &ast.APISpec{Method: "get", URL: "/api/v1/finance/accounts/options"},
		Searchable: true,
	}
}

// CurrencyPickerNode returns a Select node for choosing a currency.
func CurrencyPickerNode(name, label string, defaultCurrency string) ast.Node {
	return ast.SelectNode{
		Name:         name,
		Label:        label,
		Source:       &ast.APISpec{Method: "get", URL: "/api/v1/platform/currencies/options"},
		DefaultValue: defaultCurrency,
		Searchable:   true,
	}
}

// TaxRatePickerNode returns a Select node for choosing a tax rate.
func TaxRatePickerNode(name, label string) ast.Node {
	return ast.SelectNode{
		Name:      name,
		Label:     label,
		Source:    &ast.APISpec{Method: "get", URL: "/api/v1/finance/tax-rates/options"},
		Clearable: true,
	}
}

// PaymentMethodPickerNode returns a Select node for choosing a payment method.
func PaymentMethodPickerNode(name, label string) ast.Node {
	return ast.SelectNode{
		Name:   name,
		Label:  label,
		Source: &ast.APISpec{Method: "get", URL: "/api/v1/finance/payment-methods/options"},
	}
}
