package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// DeliveryNoteScreenConfig controls delivery note screen rendering.
type DeliveryNoteScreenConfig struct {
	ReadOnly bool
}

// DeliveryNoteScreen composes the delivery note document form.
func DeliveryNoteScreen(sess ui.UISessionContext, cfg DeliveryNoteScreenConfig) ast.Node {
	lineCfg := blocks.GRNLineItemConfig()
	lineCfg.ReadOnly = cfg.ReadOnly

	return ast.PageNode{
		Title: "Delivery Note",
		Body: []ast.Node{
			blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
				ShowStatus: true, ReadOnly: cfg.ReadOnly,
				StatusOptions: []ast.SelectOption{{Label: "Draft", Value: "draft"}, {Label: "Dispatched", Value: "dispatched"}, {Label: "Delivered", Value: "delivered"}, {Label: "Cancelled", Value: "cancelled"}},
			}),
			blocks.PartyBlock(sess, blocks.DefaultCustomerConfig()),
			blocks.AddressBlock(sess, blocks.AddressConfig{ShowShipping: true, ReadOnly: cfg.ReadOnly}),
			blocks.ProductServiceLineBlock(sess, lineCfg),
			blocks.TotalsSummaryBlock(sess),
			blocks.ApprovalWorkflowBlock(sess),
			blocks.AttachmentsBlock(sess),
		},
	}
}
