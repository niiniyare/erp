package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// GoodsReceiptScreenConfig controls GRN screen rendering.
type GoodsReceiptScreenConfig struct {
	ReadOnly bool
}

// GoodsReceiptScreen composes the goods receipt note document form.
func GoodsReceiptScreen(sess ui.UISessionContext, cfg GoodsReceiptScreenConfig) ast.Node {
	lineCfg := blocks.GRNLineItemConfig()
	lineCfg.ReadOnly = cfg.ReadOnly

	return ast.PageNode{
		Title: "Goods Receipt Note",
		Body: []ast.Node{
			blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{ShowStatus: true, ReadOnly: cfg.ReadOnly,
				StatusOptions: []ast.SelectOption{{Label: "Draft", Value: "draft"}, {Label: "Received", Value: "received"}, {Label: "Cancelled", Value: "cancelled"}}}),
			blocks.PartyBlock(sess, blocks.DefaultSupplierConfig()),
			blocks.AddressBlock(sess, blocks.AddressConfig{ShowShipping: true, ReadOnly: cfg.ReadOnly}),
			blocks.ProductServiceLineBlock(sess, lineCfg),
			blocks.TotalsSummaryBlock(sess),
			blocks.ApprovalWorkflowBlock(sess),
			blocks.AttachmentsBlock(sess),
			blocks.InternalNotesBlock(sess),
		},
	}
}
