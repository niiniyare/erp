package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// ExpenseClaimScreenConfig controls expense claim screen rendering.
type ExpenseClaimScreenConfig struct {
	ReadOnly bool
}

// ExpenseClaimScreen composes the expense claim document form.
func ExpenseClaimScreen(sess ui.UISessionContext, cfg ExpenseClaimScreenConfig) ast.Node {
	lineCfg := blocks.DefaultLineItemConfig()
	lineCfg.ShowTaxRate = true
	lineCfg.ShowUnitOfMeasure = false
	lineCfg.ReadOnly = cfg.ReadOnly

	return ast.PageNode{
		Title: "Expense Claim",
		Body: []ast.Node{
			blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
				ShowCurrency: true, ShowStatus: true, ReadOnly: cfg.ReadOnly,
				StatusOptions: []ast.SelectOption{{Label: "Draft", Value: "draft"}, {Label: "Submitted", Value: "submitted"}, {Label: "Approved", Value: "approved"}, {Label: "Paid", Value: "paid"}},
			}),
			blocks.PartyBlock(sess, blocks.DefaultEmployeeConfig()),
			blocks.ProductServiceLineBlock(sess, lineCfg),
			blocks.TaxSummaryBlock(sess),
			blocks.TotalsSummaryBlock(sess),
			blocks.ApprovalWorkflowBlock(sess),
			blocks.AttachmentsBlock(sess),
			blocks.InternalNotesBlock(sess),
		},
	}
}
