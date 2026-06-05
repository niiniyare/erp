package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// RequisitionScreen composes the internal purchase requisition form.
func RequisitionScreen(sess ui.UISessionContext) ast.Node {
	lineCfg := blocks.DefaultLineItemConfig()
	lineCfg.ShowDiscount = false
	lineCfg.ShowTaxRate = false
	return ast.PageNode{
		Title: "New Purchase Requisition",
		InitAPI: &ast.APISpec{
			Method: "get",
			URL:    "/api/v1/buy/requisitions/:id",
			SendOn: "${params.id}",
		},
		Body: []ast.Node{
			ast.SectionNode{
				Title: "Request Details",
				Body: []ast.Node{
					ast.InputTextNode{Name: "ref_number", Label: "Reference #", Required: true},
					ast.InputDateNode{Name: "required_by", Label: "Required By", Required: true},
					ast.SelectNode{
						Name:     "department_id",
						Label:    "Department",
						Required: true,
						Source:   &ast.APISpec{Method: "get", URL: "/api/v1/hr/departments/options"},
					},
					ast.SelectNode{
						Name:   "status",
						Label:  "Status",
						Options: requisitionStatusOptions(),
					},
				},
			},
			blocks.ProductServiceLineBlock(sess, lineCfg),
			blocks.ApprovalWorkflowBlock(sess),
			blocks.AttachmentsBlock(sess),
			blocks.InternalNotesBlock(sess),
		},
	}
}

func requisitionStatusOptions() []ast.SelectOption {
	return []ast.SelectOption{
		{Label: "Draft", Value: "draft"},
		{Label: "Submitted", Value: "submitted"},
		{Label: "Approved", Value: "approved"},
		{Label: "Rejected", Value: "rejected"},
		{Label: "Ordered", Value: "ordered"},
	}
}

// PurchaseOrderScreen composes the purchase order document form.
func PurchaseOrderScreen(sess ui.UISessionContext) ast.Node {
	lineCfg := blocks.DefaultLineItemConfig()
	lineCfg.ShowTaxRate = true
	return ast.PageNode{
		Title: "New Purchase Order",
		InitAPI: &ast.APISpec{
			Method: "get",
			URL:    "/api/v1/buy/purchase-orders/:id",
			SendOn: "${params.id}",
		},
		Body: []ast.Node{
			blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
				ShowCurrency: true, ShowStatus: true,
				StatusOptions: purchaseOrderStatusOptions(),
			}),
			blocks.PartyBlock(sess, blocks.DefaultSupplierConfig()),
			blocks.AddressBlock(sess, blocks.AddressConfig{ShowBilling: true}),
			blocks.ProductServiceLineBlock(sess, lineCfg),
			blocks.TaxSummaryBlock(sess),
			blocks.TotalsSummaryBlock(sess),
			blocks.ApprovalWorkflowBlock(sess),
			blocks.PaymentTermsBlock(sess, blocks.PaymentTermsConfig{}),
			blocks.AttachmentsBlock(sess),
			blocks.InternalNotesBlock(sess),
		},
	}
}

func purchaseOrderStatusOptions() []ast.SelectOption {
	return []ast.SelectOption{
		{Label: "Draft", Value: "draft"},
		{Label: "Sent", Value: "sent"},
		{Label: "Confirmed", Value: "confirmed"},
		{Label: "Received", Value: "received"},
		{Label: "Invoiced", Value: "invoiced"},
		{Label: "Cancelled", Value: "cancelled"},
	}
}

