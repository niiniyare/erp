package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// JournalEntryScreenConfig controls journal entry screen rendering.
type JournalEntryScreenConfig struct {
	ReadOnly bool
}

// JournalEntryScreen composes the journal entry document form.
func JournalEntryScreen(sess ui.UISessionContext, cfg JournalEntryScreenConfig) ast.Node {
	lineCfg := blocks.JournalLineItemConfig()
	lineCfg.ReadOnly = cfg.ReadOnly

	return ast.PageNode{
		Title: "Journal Entry",
		InitAPI: &ast.APISpec{
			Method: "get",
			URL:    "/api/v1/finance/journal-entries/:id",
			SendOn: "${params.id}",
			ResponseData: map[string]any{
				"can_approve":     "${data.permissions.can_approve|default:false}",
				"tenant_currency": "${data.tenant_currency|default:''}",
			},
		},
		Body: []ast.Node{
			blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{ShowCurrency: true, ShowStatus: true, ReadOnly: cfg.ReadOnly,
				StatusOptions: []ast.SelectOption{{Label: "Draft", Value: "draft"}, {Label: "Posted", Value: "posted"}, {Label: "Reversed", Value: "reversed"}}}),
			blocks.ProductServiceLineBlock(sess, lineCfg),
			blocks.TotalsSummaryBlock(sess),
			blocks.ApprovalWorkflowBlock(sess),
			blocks.AttachmentsBlock(sess),
			blocks.InternalNotesBlock(sess),
		},
	}
}
