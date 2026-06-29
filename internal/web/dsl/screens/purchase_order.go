package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// POScreenConfig controls purchase order screen behaviour.
type POScreenConfig struct {
	ShowAttachments   bool
	ShowInternalNotes bool
	ReadOnly          bool
}

// POScreen composes the purchase order document form.
func POScreen(sess ui.UISessionContext, cfg POScreenConfig) ast.Node {
	lineCfg := blocks.DefaultLineItemConfig()
	lineCfg.ShowTaxRate = true
	lineCfg.ReadOnly = cfg.ReadOnly

	body := []ast.Node{
		blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
			ShowCurrency: true, ShowStatus: true, ReadOnly: cfg.ReadOnly,
			StatusOptions: poStatusOptions(),
		}),
		blocks.PartyBlock(sess, blocks.DefaultSupplierConfig()),
		blocks.AddressBlock(sess, blocks.AddressConfig{ShowShipping: true, ReadOnly: cfg.ReadOnly}),
		blocks.ProductServiceLineBlock(sess, lineCfg),
		blocks.TaxSummaryBlock(sess),
		blocks.TotalsSummaryBlock(sess),
		blocks.PaymentTermsBlock(sess, blocks.PaymentTermsConfig{ReadOnly: cfg.ReadOnly}),
		blocks.ApprovalWorkflowBlock(sess),
	}
	if cfg.ShowAttachments {
		body = append(body, blocks.AttachmentsBlock(sess))
	}
	if cfg.ShowInternalNotes {
		body = append(body, blocks.InternalNotesBlock(sess))
	}
	return ast.PageNode{Title: "Purchase Order", Body: body}
}

func poStatusOptions() []ast.SelectOption {
	return []ast.SelectOption{
		{Label: "Draft", Value: "draft"},
		{Label: "Submitted", Value: "submitted"},
		{Label: "Approved", Value: "approved"},
		{Label: "Received", Value: "received"},
		{Label: "Cancelled", Value: "cancelled"},
	}
}
