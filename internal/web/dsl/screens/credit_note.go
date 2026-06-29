package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// CreditNoteScreenConfig controls credit note screen rendering.
type CreditNoteScreenConfig struct {
	ReadOnly bool
}

// CreditNoteScreen composes the credit note document form.
func CreditNoteScreen(sess ui.UISessionContext, cfg CreditNoteScreenConfig) ast.Node {
	lineCfg := blocks.DefaultLineItemConfig()
	lineCfg.ShowDiscount = true
	lineCfg.ShowTaxRate = true
	lineCfg.ReadOnly = cfg.ReadOnly

	return ast.PageNode{
		Title: "Credit Note",
		Body: []ast.Node{
			blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
				ShowCurrency: true, ShowStatus: true, ReadOnly: cfg.ReadOnly,
				StatusOptions: []ast.SelectOption{{Label: "Draft", Value: "draft"}, {Label: "Issued", Value: "issued"}, {Label: "Applied", Value: "applied"}},
			}),
			blocks.PartyBlock(sess, blocks.DefaultCustomerConfig()),
			blocks.ProductServiceLineBlock(sess, lineCfg),
			blocks.TaxSummaryBlock(sess),
			blocks.TotalsSummaryBlock(sess),
			blocks.ApprovalWorkflowBlock(sess),
			blocks.AttachmentsBlock(sess),
			blocks.InternalNotesBlock(sess),
		},
	}
}
