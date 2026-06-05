package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// QuotationScreen composes the sales quotation document form.
func QuotationScreen(sess ui.UISessionContext) ast.Node {
	lineCfg := blocks.DefaultLineItemConfig()
	lineCfg.ShowDiscount = true
	lineCfg.ShowTaxRate = true
	return ast.PageNode{
		Title: "New Quotation",
		InitAPI: &ast.APISpec{
			Method: "get",
			URL:    "/api/v1/sell/quotations/:id",
			SendOn: "${params.id}",
		},
		Body: []ast.Node{
			blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
				ShowCurrency: true, ShowStatus: true,
				StatusOptions: quotationStatusOptions(),
			}),
			blocks.PartyBlock(sess, blocks.DefaultCustomerConfig()),
			blocks.AddressBlock(sess, blocks.AddressConfig{ShowBilling: true, ShowShipping: true}),
			blocks.ProductServiceLineBlock(sess, lineCfg),
			blocks.TaxSummaryBlock(sess),
			blocks.TotalsSummaryBlock(sess),
			blocks.PaymentTermsBlock(sess, blocks.PaymentTermsConfig{}),
			blocks.AttachmentsBlock(sess),
			blocks.InternalNotesBlock(sess),
		},
	}
}

func quotationStatusOptions() []ast.SelectOption {
	return []ast.SelectOption{
		{Label: "Draft", Value: "draft"},
		{Label: "Sent", Value: "sent"},
		{Label: "Accepted", Value: "accepted"},
		{Label: "Rejected", Value: "rejected"},
		{Label: "Expired", Value: "expired"},
	}
}

// SalesOrderScreen composes the sales order document form.
func SalesOrderScreen(sess ui.UISessionContext) ast.Node {
	lineCfg := blocks.DefaultLineItemConfig()
	lineCfg.ShowDiscount = true
	lineCfg.ShowTaxRate = true
	return ast.PageNode{
		Title: "New Sales Order",
		InitAPI: &ast.APISpec{
			Method: "get",
			URL:    "/api/v1/sell/orders/:id",
			SendOn: "${params.id}",
		},
		Body: []ast.Node{
			blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
				ShowCurrency: true, ShowStatus: true,
				StatusOptions: salesOrderStatusOptions(),
			}),
			blocks.PartyBlock(sess, blocks.DefaultCustomerConfig()),
			blocks.AddressBlock(sess, blocks.AddressConfig{ShowBilling: true, ShowShipping: true}),
			blocks.ProductServiceLineBlock(sess, lineCfg),
			blocks.TaxSummaryBlock(sess),
			blocks.TotalsSummaryBlock(sess),
			blocks.ApprovalWorkflowBlock(sess),
			blocks.AttachmentsBlock(sess),
			blocks.InternalNotesBlock(sess),
		},
	}
}

func salesOrderStatusOptions() []ast.SelectOption {
	return []ast.SelectOption{
		{Label: "Draft", Value: "draft"},
		{Label: "Confirmed", Value: "confirmed"},
		{Label: "Delivered", Value: "delivered"},
		{Label: "Invoiced", Value: "invoiced"},
		{Label: "Cancelled", Value: "cancelled"},
	}
}

// DeliveryNoteScreen composes the goods delivery confirmation form.
func DeliveryNoteScreen(sess ui.UISessionContext) ast.Node {
	lineCfg := blocks.GRNLineItemConfig()
	return ast.PageNode{
		Title: "New Delivery Note",
		InitAPI: &ast.APISpec{
			Method: "get",
			URL:    "/api/v1/sell/delivery-notes/:id",
			SendOn: "${params.id}",
		},
		Body: []ast.Node{
			blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
				ShowStatus: true,
				StatusOptions: []ast.SelectOption{
					{Label: "Draft", Value: "draft"},
					{Label: "Dispatched", Value: "dispatched"},
					{Label: "Delivered", Value: "delivered"},
				},
			}),
			blocks.PartyBlock(sess, blocks.DefaultCustomerConfig()),
			blocks.AddressBlock(sess, blocks.AddressConfig{ShowShipping: true}),
			blocks.ProductServiceLineBlock(sess, lineCfg),
			blocks.AttachmentsBlock(sess),
			blocks.InternalNotesBlock(sess),
		},
	}
}
