package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// TaxReportScreen renders the VAT/Tax Return report.
func TaxReportScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Tax Report",
		Body: []ast.Node{
			blocks.ReportHeaderBlock(sess, blocks.ReportHeaderConfig{Title: "Tax Report"}),
			blocks.ReportFilterBlock(sess, blocks.ReportFilterConfig{ShowPeriod: true, ShowCurrency: true}),
			blocks.ReportTableBlock(sess, blocks.ReportTableConfig{
				Columns: []blocks.ReportColumnDef{
					{Name: "tax_name", Label: "Tax", Type: "text"},
					{Name: "taxable_amount", Label: "Taxable Amount", Type: "currency"},
					{Name: "tax_collected", Label: "Tax Collected", Type: "currency"},
					{Name: "tax_paid", Label: "Tax Paid", Type: "currency"},
					{Name: "net_tax", Label: "Net Tax", Type: "currency"},
				},
				ShowGrandTotal: true,
				Exportable:     true,
			}),
		},
	}
}
