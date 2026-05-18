package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// PaymentTermsConfig controls payment terms block rendering.
type PaymentTermsConfig struct {
	ReadOnly bool
}

// PaymentTermsBlock renders the payment terms section (net days, due date).
func PaymentTermsBlock(sess ui.UISessionContext, cfg PaymentTermsConfig) ast.Node {
	return ast.SectionNode{
		Title: "Payment Terms",
		Body: []ast.Node{
			ast.SelectNode{
				Name:       "payment_terms_id",
				Label:      "Payment Terms",
				Source:     &ast.APISpec{Method: "get", URL: "/api/v1/finance/payment-terms/options"},
				DisabledOn: boolExpr(cfg.ReadOnly),
			},
			ast.InputDateNode{
				Name:       "payment_due_date",
				Label:      "Due Date",
				DisabledOn: boolExpr(cfg.ReadOnly),
			},
		},
	}
}
