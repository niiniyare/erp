package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// DebitNoteScreenConfig controls debit note screen rendering.
type DebitNoteScreenConfig struct {
	ReadOnly bool
}

// DebitNoteScreen composes the debit note document form.
func DebitNoteScreen(sess ui.UISessionContext, cfg DebitNoteScreenConfig) ast.Node {
	lineCfg := blocks.DefaultLineItemConfig()
	lineCfg.ShowTaxRate = true
	lineCfg.ReadOnly = cfg.ReadOnly

	return ast.PageNode{
		Title: "Debit Note",
		Body: []ast.Node{
			blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
				ShowCurrency: true, ShowStatus: true, ReadOnly: cfg.ReadOnly,
				StatusOptions: []ast.SelectOption{{Label: "Draft", Value: "draft"}, {Label: "Issued", Value: "issued"}, {Label: "Applied", Value: "applied"}},
			}),
			blocks.PartyBlock(sess, blocks.DefaultSupplierConfig()),
			blocks.ProductServiceLineBlock(sess, lineCfg),
			blocks.TaxSummaryBlock(sess),
			blocks.TotalsSummaryBlock(sess),
			blocks.ApprovalWorkflowBlock(sess),
			blocks.AttachmentsBlock(sess),
			blocks.InternalNotesBlock(sess),
		},
	}
}
