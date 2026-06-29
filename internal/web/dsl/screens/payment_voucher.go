package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// PaymentVoucherScreenConfig controls payment voucher screen rendering.
type PaymentVoucherScreenConfig struct {
	IsReceipt bool // true = receipt voucher, false = payment voucher
	ReadOnly  bool
}

// PaymentVoucherScreen composes the payment/receipt voucher form.
func PaymentVoucherScreen(sess ui.UISessionContext, cfg PaymentVoucherScreenConfig) ast.Node {
	title := "Payment Voucher"
	if cfg.IsReceipt {
		title = "Receipt Voucher"
	}
	partyCfg := blocks.DefaultSupplierConfig()
	if cfg.IsReceipt {
		partyCfg = blocks.DefaultCustomerConfig()
	}
	return ast.PageNode{
		Title: title,
		Body: []ast.Node{
			blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
				ShowCurrency: true, ShowStatus: true, ReadOnly: cfg.ReadOnly,
				StatusOptions: []ast.SelectOption{{Label: "Draft", Value: "draft"}, {Label: "Posted", Value: "posted"}, {Label: "Cancelled", Value: "cancelled"}},
			}),
			blocks.PartyBlock(sess, partyCfg),
			blocks.TotalsSummaryBlock(sess),
			blocks.PaymentTermsBlock(sess, blocks.PaymentTermsConfig{ReadOnly: cfg.ReadOnly}),
			blocks.ApprovalWorkflowBlock(sess),
			blocks.AttachmentsBlock(sess),
			blocks.InternalNotesBlock(sess),
		},
	}
}
