package finance

import (
	"context"
	"fmt"

	"awo.so/awo/def"
	"awo.so/awo/filter"
	"awo.so/awo/runtime"
)

// JournalDefinition — journal master (General, Sales, Purchase, Bank, Cash).
var JournalDefinition = def.SystemDefinition{
	Name:        "journal",
	Module:      "finance",
	Label:       "Journal",
	Description: "Journal master. Groups related journal entries by type (e.g. Sales Journal, General Journal).",
	Fields: []def.FieldDef{
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Required:   true,
			Searchable: true,
			MaxLen:     255,
		},
		{
			Name:      "code",
			Type:      def.FieldTypeData,
			Required:  true,
			Unique:    true,
			Immutable: true,
			MaxLen:    10,
		},
		{
			Name:     "journal_type",
			Type:     def.FieldTypeSelect,
			Label:    "Journal Type",
			Options:  []string{"General", "Sales", "Purchase", "Bank", "Cash", "Opening"},
			Required: true,
		},
		{Name: "description", Type: def.FieldTypeSmallText},
		{
			Name:       "default_account_id",
			Type:       def.FieldTypeLink,
			Label:      "Default Account",
			LinkTarget: "finance_account",
		},
		{
			Name:    "active",
			Type:    def.FieldTypeBool,
			Default: func() any { return true },
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"role:tenant.admin", "role:finance.manager"},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer", "role:tenant.user"},
		Write:  []string{"role:tenant.admin", "role:finance.manager"},
		Delete: []string{"role:tenant.admin"},
	},
}

// JournalEntryDefinition — double-entry journal header. MANDATORY system entity.
//
// Lifecycle: draft → submitted → posted → reversed/cancelled.
// Balance validation and ledger entry creation happen in the "post" action.
var JournalEntryDefinition = def.SystemDefinition{
	Name:        "journal_entry",
	Module:      "finance",
	Label:       "Journal Entry",
	LabelPlural: "Journal Entries",
	Description: "Double-entry journal header. Lines hold the debit/credit split. Post to create immutable ledger entries.",
	Fields: []def.FieldDef{
		{
			Name:              "entry_number",
			Type:              def.FieldTypeNamingSeries,
			Label:             "Entry Number",
			Series:            "JE-{YYYY}-{SEQ:5}",
			Immutable:         true,
			TenantOverridable: false,
		},
		{
			Name:       "journal_id",
			Type:       def.FieldTypeLink,
			Label:      "Journal",
			LinkTarget: "finance_journal",
			Required:   true,
		},
		{Name: "posting_date", Type: def.FieldTypeDate, Label: "Posting Date", Required: true},
		{
			Name:       "accounting_period_id",
			Type:       def.FieldTypeLink,
			Label:      "Accounting Period",
			LinkTarget: "finance_accounting_period",
			Required:   true,
		},
		{
			Name:    "status",
			Type:    def.FieldTypeSelect,
			Options: []string{"draft", "submitted", "posted", "cancelled", "reversed"},
			Default: func() any { return "draft" },
		},
		{Name: "memo", Type: def.FieldTypeSmallText},
		{Name: "reference", Type: def.FieldTypeData, Label: "External Reference", MaxLen: 100},
		{
			Name:       "currency_id",
			Type:       def.FieldTypeLink,
			Label:      "Currency",
			LinkTarget: "finance_currency",
			Required:   true,
		},
		{
			Name:    "exchange_rate",
			Type:    def.FieldTypeCurrency,
			Label:   "Exchange Rate",
			Default: func() any { return "1.0000" },
		},
		{Name: "total_debit", Type: def.FieldTypeCurrency, Label: "Total Debit", ReadOnly: true},
		{Name: "total_credit", Type: def.FieldTypeCurrency, Label: "Total Credit", ReadOnly: true},
		{
			Name:      "is_reversal",
			Type:      def.FieldTypeBool,
			Label:     "Is Reversal",
			Immutable: true,
			Default:   func() any { return false },
		},
		{
			Name:       "reversal_of_id",
			Type:       def.FieldTypeLink,
			Label:      "Reversal Of",
			LinkTarget: "finance_journal_entry",
			Immutable:  true,
		},
	},
	Edges: []def.EdgeDef{
		{
			Name:          "lines",
			Target:        "finance_journal_entry_line",
			Type:          def.EdgeOneToMany,
			ForeignKey:    "journal_entry_id",
			CascadeDelete: true,
			Label:         "Entry Lines",
			OrderBy:       "created_at ASC",
		},
	},
	Hooks: def.HookSet{
		BeforeCreate: []def.BeforeCreateHook{&JournalEntryValidator{}},
		BeforeUpdate: []def.BeforeUpdateHook{&PostingGuard{}},
	},
	Actions: []def.ActionDef{
		{
			Name:        "post",
			Method:      def.ActionMethodPost,
			Label:       "Post Entry",
			Description: "Validate balance and create immutable ledger entries.",
			Permission:  "role:finance.accountant",
			Icon:        "fa fa-check",
			HandlerFunc: postJournalEntry,
		},
		{
			Name:           "reverse",
			Method:         def.ActionMethodPost,
			Label:          "Reverse Entry",
			Description:    "Create a mirror entry with swapped debits/credits.",
			Permission:     "role:finance.manager",
			Icon:           "fa fa-undo",
			ConfirmMessage: "This will create a reversal entry. Continue?",
			HandlerFunc:    reverseJournalEntry,
		},
	},
	WorkflowTriggers: []def.WorkflowTrigger{
		{
			On:         def.EventOnSubmit,
			WorkflowFn: "JournalEntryApprovalWorkflow",
			TaskQueue:  "finance.journal.approval",
			InputBuilder: func(r *def.EntityRecord, tc def.TriggerContext) (any, error) {
				return JournalEntryApprovalInput{
					TenantID:       r.TenantID,
					JournalEntryID: r.ID,
					ActorID:        tc.Actor.UserID,
				}, nil
			},
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant"},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer"},
		Write:  []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant"},
		Delete: []string{"role:tenant.admin", "role:finance.manager"},
		Actions: map[string][]string{
			"post":    {"role:finance.accountant", "role:finance.manager", "role:tenant.admin"},
			"reverse": {"role:finance.manager", "role:tenant.admin"},
		},
	},
}

// JournalEntryLineDefinition — individual debit or credit line within a journal entry.
var JournalEntryLineDefinition = def.SystemDefinition{
	Name:        "journal_entry_line",
	Module:      "finance",
	Label:       "Journal Entry Line",
	Description: "Debit or credit line. Exactly one of debit_amount / credit_amount must be non-zero.",
	Fields: []def.FieldDef{
		{
			Name:       "journal_entry_id",
			Type:       def.FieldTypeLink,
			Label:      "Journal Entry",
			LinkTarget: "finance_journal_entry",
			Required:   true,
			Immutable:  true,
		},
		{
			Name:       "account_id",
			Type:       def.FieldTypeLink,
			Label:      "Account",
			LinkTarget: "finance_account",
			Required:   true,
		},
		{
			Name:       "cost_center_id",
			Type:       def.FieldTypeLink,
			Label:      "Cost Center",
			LinkTarget: "finance_cost_center",
		},
		{
			Name:    "debit_amount",
			Type:    def.FieldTypeCurrency,
			Label:   "Debit",
			Default: func() any { return "0.0000" },
		},
		{
			Name:    "credit_amount",
			Type:    def.FieldTypeCurrency,
			Label:   "Credit",
			Default: func() any { return "0.0000" },
		},
		{Name: "memo", Type: def.FieldTypeSmallText},
		{Name: "reference_type", Type: def.FieldTypeData, Label: "Reference Type", MaxLen: 100},
		{Name: "reference_id", Type: def.FieldTypeData, Label: "Reference ID", MaxLen: 36},
	},
	Hooks: def.HookSet{
		BeforeCreate: []def.BeforeCreateHook{&LineBalanceValidator{}},
	},
	Permissions: def.PermissionSet{
		Create: []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant"},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer"},
		Write:  []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant"},
		Delete: []string{"role:tenant.admin", "role:finance.manager"},
	},
}

// LedgerEntryDefinition — immutable double-entry ledger record. MANDATORY system entity.
//
// Written exclusively by PostingService. Never writable via API.
var LedgerEntryDefinition = def.SystemDefinition{
	Name:        "ledger_entry",
	Module:      "finance",
	Label:       "Ledger Entry",
	LabelPlural: "Ledger Entries",
	Description: "Immutable GL ledger record created when a journal entry is posted. Never edited.",
	Fields: []def.FieldDef{
		{
			Name:      "entry_number",
			Type:      def.FieldTypeNamingSeries,
			Label:     "Entry Number",
			Series:    "LE-{YYYY}-{SEQ:7}",
			Immutable: true,
		},
		{
			Name:       "account_id",
			Type:       def.FieldTypeLink,
			Label:      "Account",
			LinkTarget: "finance_account",
			Required:   true,
			Immutable:  true,
		},
		{
			Name:       "journal_entry_id",
			Type:       def.FieldTypeLink,
			Label:      "Journal Entry",
			LinkTarget: "finance_journal_entry",
			Required:   true,
			Immutable:  true,
		},
		{
			Name:       "journal_entry_line_id",
			Type:       def.FieldTypeLink,
			Label:      "Journal Entry Line",
			LinkTarget: "finance_journal_entry_line",
			Required:   true,
			Immutable:  true,
		},
		{Name: "posting_date", Type: def.FieldTypeDate, Required: true, Immutable: true},
		{
			Name:      "debit_amount",
			Type:      def.FieldTypeCurrency,
			Label:     "Debit",
			Immutable: true,
			Default:   func() any { return "0.0000" },
		},
		{
			Name:      "credit_amount",
			Type:      def.FieldTypeCurrency,
			Label:     "Credit",
			Immutable: true,
			Default:   func() any { return "0.0000" },
		},
		{Name: "running_balance", Type: def.FieldTypeCurrency, Label: "Running Balance", ReadOnly: true},
		{
			Name:       "currency_id",
			Type:       def.FieldTypeLink,
			Label:      "Currency",
			LinkTarget: "finance_currency",
			Immutable:  true,
		},
		{
			Name:      "exchange_rate",
			Type:      def.FieldTypeCurrency,
			Label:     "Exchange Rate",
			Immutable: true,
			Default:   func() any { return "1.0000" },
		},
		{Name: "memo", Type: def.FieldTypeSmallText, Immutable: true},
		{
			Name:       "cost_center_id",
			Type:       def.FieldTypeLink,
			Label:      "Cost Center",
			LinkTarget: "finance_cost_center",
			Immutable:  true,
		},
		{
			Name:    "is_cancelled",
			Type:    def.FieldTypeBool,
			Default: func() any { return false },
		},
	},
	// Empty slices = framework writes this, API cannot
	Permissions: def.PermissionSet{
		Create: []string{},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer"},
		Write:  []string{},
		Delete: []string{},
	},
}

func postJournalEntry(ctx *def.ActionContext) (*def.ActionResult, error) {
	rt := ctx.Runtime
	entryRepo := rt.Repo("finance_journal_entry")
	lineRepo := rt.Repo("finance_journal_entry_line")
	ledgerRepo := rt.Repo("finance_ledger_entry")
	periodRepo := rt.Repo("finance_accounting_period")

	// Load the journal entry.
	entry, err := entryRepo.Get(ctx.Ctx, ctx.RecordID)
	if err != nil {
		return nil, fmt.Errorf("finance.post: load entry: %w", err)
	}

	status := entry.GetString("status")
	if status != "submitted" {
		return nil, &runtime.BusinessError{
			Code:    "finance.journal_entry.invalid_status_for_post",
			Message: fmt.Sprintf("journal entry must be in 'submitted' status to post; current status: %s", status),
			Status:  400,
		}
	}

	// Verify accounting period is open.
	periodID := entry.GetUUID("accounting_period_id")
	period, err := periodRepo.Get(ctx.Ctx, periodID)
	if err != nil {
		return nil, fmt.Errorf("finance.post: load period: %w", err)
	}
	if s := period.GetString("status"); s != "open" {
		return nil, &runtime.BusinessError{
			Code:    "finance.journal_entry.period_not_open",
			Message: fmt.Sprintf("accounting period status is '%s'; only open periods accept postings", s),
			Status:  400,
		}
	}

	// Load lines.
	lines, err := lineRepo.Query(ctx.Ctx,
		filter.Eq("journal_entry_id", ctx.RecordID),
		def.WithActionOrder("created_at ASC"),
	)
	if err != nil {
		return nil, fmt.Errorf("finance.post: load lines: %w", err)
	}
	if len(lines) < 2 {
		return nil, &runtime.BusinessError{
			Code:    "finance.journal_entry.insufficient_lines",
			Message: "journal entry must have at least two lines",
			Status:  400,
		}
	}

	// Validate double-entry balance.
	if err := ValidateBalance(lines); err != nil {
		return nil, err
	}

	// Execute within a transaction: create ledger entries + update status.
	var created int
	if err := rt.Tx(ctx.Ctx, func(txCtx context.Context) error {
		// Create one LedgerEntry per line.
		postingDate := entry.GetString("posting_date")
		currencyID := entry.GetUUID("currency_id")
		exchangeRate := entry.GetDecimal("exchange_rate")
		memo := entry.GetString("memo")

		for _, line := range lines {
			ledgerData := map[string]any{
				"account_id":            line.GetUUID("account_id"),
				"journal_entry_id":      ctx.RecordID,
				"journal_entry_line_id": line.ID,
				"posting_date":          postingDate,
				"debit_amount":          line.GetDecimal("debit_amount"),
				"credit_amount":         line.GetDecimal("credit_amount"),
				"currency_id":           currencyID,
				"exchange_rate":         exchangeRate,
				"memo":                  memo,
				"cost_center_id":        line.GetUUID("cost_center_id"),
				"is_cancelled":          false,
			}
			if _, err := ledgerRepo.Create(txCtx, ledgerData); err != nil {
				return fmt.Errorf("finance.post: create ledger entry: %w", err)
			}
			created++
		}

		// Update entry status to posted.
		_, err := entryRepo.Update(txCtx, ctx.RecordID, map[string]any{"status": "posted"})
		return err
	}); err != nil {
		return nil, fmt.Errorf("finance.post: transaction: %w", err)
	}

	// Publish domain event (outside TX).
	_ = rt.Publish(ctx.Ctx, def.ActionEvent{
		Topic:   "finance.journal_entry.posted",
		Payload: map[string]any{"entry_id": ctx.RecordID, "ledger_entries_created": created},
	})

	return &def.ActionResult{
		Message: fmt.Sprintf("Journal entry posted. %d ledger entries created.", created),
		Data:    map[string]any{"entry_id": ctx.RecordID, "status": "posted"},
	}, nil
}

func reverseJournalEntry(ctx *def.ActionContext) (*def.ActionResult, error) {
	rt := ctx.Runtime
	entryRepo := rt.Repo("finance_journal_entry")
	lineRepo := rt.Repo("finance_journal_entry_line")

	// Load the journal entry.
	entry, err := entryRepo.Get(ctx.Ctx, ctx.RecordID)
	if err != nil {
		return nil, fmt.Errorf("finance.reverse: load entry: %w", err)
	}

	if entry.GetString("status") != "posted" {
		return nil, &runtime.BusinessError{
			Code:    "finance.journal_entry.not_posted",
			Message: "only posted journal entries can be reversed",
			Status:  400,
		}
	}

	// Load original lines.
	lines, err := lineRepo.Query(ctx.Ctx, filter.Eq("journal_entry_id", ctx.RecordID))
	if err != nil {
		return nil, fmt.Errorf("finance.reverse: load lines: %w", err)
	}

	var reversalID interface{}

	if err := rt.Tx(ctx.Ctx, func(txCtx context.Context) error {
		// Create reversal journal entry.
		reversalData := map[string]any{
			"journal_id":           entry.GetUUID("journal_id"),
			"posting_date":         entry.GetString("posting_date"),
			"accounting_period_id": entry.GetUUID("accounting_period_id"),
			"status":               "submitted",
			"memo":                 "Reversal of " + entry.GetString("entry_number"),
			"currency_id":          entry.GetUUID("currency_id"),
			"exchange_rate":        entry.GetDecimal("exchange_rate"),
			"is_reversal":          true,
			"reversal_of_id":       ctx.RecordID,
		}
		reversal, err := entryRepo.Create(txCtx, reversalData)
		if err != nil {
			return fmt.Errorf("finance.reverse: create reversal entry: %w", err)
		}
		reversalID = reversal.ID

		// Mirror lines with swapped debit/credit.
		for _, line := range lines {
			mirrorData := map[string]any{
				"journal_entry_id": reversal.ID,
				"account_id":       line.GetUUID("account_id"),
				"cost_center_id":   line.GetUUID("cost_center_id"),
				// Swap debit ↔ credit.
				"debit_amount":  line.GetDecimal("credit_amount"),
				"credit_amount": line.GetDecimal("debit_amount"),
				"memo":          line.GetString("memo"),
			}
			if _, err := lineRepo.Create(txCtx, mirrorData); err != nil {
				return fmt.Errorf("finance.reverse: create mirror line: %w", err)
			}
		}

		// Mark original as reversed.
		_, err = entryRepo.Update(txCtx, ctx.RecordID, map[string]any{"status": "reversed"})
		return err
	}); err != nil {
		return nil, fmt.Errorf("finance.reverse: transaction: %w", err)
	}

	_ = rt.Publish(ctx.Ctx, def.ActionEvent{
		Topic:   "finance.journal_entry.reversed",
		Payload: map[string]any{"original_id": ctx.RecordID, "reversal_id": reversalID},
	})

	return &def.ActionResult{
		Message:    "Reversal entry created and submitted for posting.",
		Data:       map[string]any{"original_id": ctx.RecordID, "reversal_entry_id": reversalID},
		WorkflowID: "",
	}, nil
}
