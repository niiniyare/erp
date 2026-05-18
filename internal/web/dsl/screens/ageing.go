package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// AgeingScreenConfig controls AR vs AP ageing report.
type AgeingScreenConfig struct {
	IsAP bool // false = AR ageing, true = AP ageing
}

// AgeingScreen renders the Accounts Receivable or Payable Ageing report.
func AgeingScreen(sess ui.UISessionContext, cfg AgeingScreenConfig) ast.Node {
	title := "AR Ageing"
	if cfg.IsAP {
		title = "AP Ageing"
	}
	return ast.PageNode{
		Title: title,
		Body: []ast.Node{
			blocks.ReportHeaderBlock(sess, blocks.ReportHeaderConfig{Title: title}),
			blocks.ReportFilterBlock(sess, blocks.ReportFilterConfig{ShowPeriod: true, ShowCurrency: true}),
			blocks.ReportTableBlock(sess, blocks.ReportTableConfig{
				Columns: []blocks.ReportColumnDef{
					{Name: "party_name", Label: "Name", Type: "text"},
					{Name: "current", Label: "Current", Type: "currency"},
					{Name: "days_30", Label: "1-30 Days", Type: "currency"},
					{Name: "days_60", Label: "31-60 Days", Type: "currency"},
					{Name: "days_90", Label: "61-90 Days", Type: "currency"},
					{Name: "over_90", Label: "Over 90", Type: "currency"},
					{Name: "total", Label: "Total", Type: "currency"},
				},
				ShowGrandTotal: true,
				Exportable:     true,
			}),
		},
	}
}
