package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// InvoiceScreenConfig controls invoice/bill screen behaviour.
type InvoiceScreenConfig struct {
	IsPurchase        bool
	ShowPaymentTerms  bool
	ShowAttachments   bool
	ShowInternalNotes bool
	ReadOnly          bool
}

// InvoiceScreen composes a full invoice or bill document form.
func InvoiceScreen(sess ui.UISessionContext, cfg InvoiceScreenConfig) ast.Node {
	lineCfg := blocks.DefaultLineItemConfig()
	lineCfg.ShowDiscount = !cfg.IsPurchase
	lineCfg.ShowTaxRate = true
	lineCfg.ReadOnly = cfg.ReadOnly

	partyCfg := blocks.DefaultCustomerConfig()
	if cfg.IsPurchase {
		partyCfg = blocks.DefaultSupplierConfig()
	}

	body := []ast.Node{
		blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{ShowCurrency: true, ShowStatus: true, ReadOnly: cfg.ReadOnly,
			StatusOptions: invoiceStatusOptions()}),
		blocks.PartyBlock(sess, partyCfg),
		blocks.AddressBlock(sess, blocks.AddressConfig{ShowBilling: true, ShowShipping: !cfg.IsPurchase, ReadOnly: cfg.ReadOnly}),
		blocks.ProductServiceLineBlock(sess, lineCfg),
		blocks.TaxSummaryBlock(sess),
		blocks.TotalsSummaryBlock(sess),
		blocks.ApprovalWorkflowBlock(sess),
	}
	if cfg.ShowPaymentTerms {
		body = append(body, blocks.PaymentTermsBlock(sess, blocks.PaymentTermsConfig{ReadOnly: cfg.ReadOnly}))
	}
	if cfg.ShowAttachments {
		body = append(body, blocks.AttachmentsBlock(sess))
	}
	if cfg.ShowInternalNotes {
		body = append(body, blocks.InternalNotesBlock(sess))
	}
	docType := "invoices"
	if cfg.IsPurchase {
		docType = "bills"
	}
	return ast.PageNode{
		Title: pageTitle(cfg.IsPurchase, "Invoice", "Bill"),
		// InitAPI loads the document record when editing (:id in URL).
		// SendOn skips the call on the new-document route (no id param).
		InitAPI: &ast.APISpec{
			Method: "get",
			URL:    "/api/v1/finance/" + docType + "/:id",
			SendOn: "${params.id}",
			// Hoist can_approve and tenant_currency into page scope.
			ResponseData: map[string]any{
				"can_approve":     "${data.permissions.can_approve|default:false}",
				"tenant_currency": "${data.tenant_currency|default:''}",
			},
		},
		Body: body,
	}
}

func invoiceStatusOptions() []ast.SelectOption {
	return []ast.SelectOption{
		{Label: "Draft", Value: "draft"},
		{Label: "Sent", Value: "sent"},
		{Label: "Paid", Value: "paid"},
		{Label: "Overdue", Value: "overdue"},
		{Label: "Cancelled", Value: "cancelled"},
	}
}

func pageTitle(isPurchase bool, salesTitle, purchaseTitle string) string {
	if isPurchase {
		return purchaseTitle
	}
	return salesTitle
}
