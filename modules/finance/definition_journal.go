package finance

import "awo.so/awo/def"

// JournalDefinition — journal master (GL, Sales, Purchase, Bank, Cash, Opening).
var JournalDefinition = def.SystemDefinition{
	Name:        "journal",
	Module:      "finance",
	Label:       "Journal",
	LabelPlural: "Journals",
	Description: "Journal master. Groups journal entries by type (GL, Sales, Purchase, Bank, Cash, Opening).",
	Icon:        "book",
	Scope:       def.ScopeTenant,

	Fields: []def.FieldDef{
		{
			Name:       "code",
			Type:       def.FieldTypeData,
			Label:      "Journal Code",
			Required:   true,
			Unique:     true,
			MaxLen:     20,
			Searchable: true,
		},
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Label:      "Journal Name",
			Required:   true,
			Searchable: true,
			MaxLen:     255,
		},
		{
			Name:     "journal_type",
			Type:     def.FieldTypeSelect,
			Label:    "Journal Type",
			Required: true,
			Immutable: true,
			Options:  []string{"general", "sales", "purchase", "bank", "cash", "opening"},
		},
		{
			Name:       "default_account",
			Type:       def.FieldTypeLink,
			Label:      "Default Account",
			Description: "Default GL account for entries in this journal.",
			LinkTarget: "finance_account",
		},
		{
			Name:       "currency",
			Type:       def.FieldTypeLink,
			Label:      "Currency",
			LinkTarget: "finance_currency",
		},
		{
			Name:    "is_active",
			Type:    def.FieldTypeBool,
			Label:   "Active",
			Default: func() any { return true },
		},
		{
			Name:    "sequence_prefix",
			Type:    def.FieldTypeData,
			Label:   "Sequence Prefix",
			Description: "Prefix for auto-generated entry sequence numbers (e.g., 'JE', 'INV', 'BILL').",
			MaxLen:  20,
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"finance.journal.create"},
		Read:   []string{"finance.journal.read"},
		Write:  []string{"finance.journal.update"},
		Delete: []string{"finance.journal.delete"},
	},
}

func init() {
	def.Register(&JournalDefinition)
}

// JournalEntryDefinition — double-entry header with state machine.
// State: draft → submitted → posted → reversed
// Posted entries are immutable. Reversal creates a mirror entry.
var JournalEntryDefinition = def.SystemDefinition{
	Name:        "journal_entry",
	Module:      "finance",
	Label:       "Journal Entry",
	LabelPlural: "Journal Entries",
	Description: "Double-entry journal entry header. Transitions: draft → submitted → posted → reversed. Posted entries are immutable.",
	Icon:        "file-text",
	Scope:       def.ScopeTenant,

	Fields: []def.FieldDef{
		{
			Name:      "entry_number",
			Type:      def.FieldTypeNamingSeries,
			Label:     "Entry Number",
			Series:    "JE-{YYYY}-{SEQ:5}",
			Required:  true,
			Immutable: true,
			ReadOnly:  true,
		},
		{
			Name:       "journal",
			Type:       def.FieldTypeLink,
			Label:      "Journal",
			LinkTarget: "finance_journal",
			Required:   true,
			Immutable:  true,
		},
		{
			Name:      "entry_date",
			Type:      def.FieldTypeDate,
			Label:     "Entry Date",
			Required:  true,
		},
		{
			Name:        "posting_date",
			Type:        def.FieldTypeDate,
			Label:       "Posting Date",
			Description: "Date the entry affects GL balances. Set on post action.",
			ReadOnly:    true,
		},
		{
			Name:     "status",
			Type:     def.FieldTypeSelect,
			Label:    "Status",
			Required: true,
			Options:  []string{"draft", "submitted", "posted", "reversed", "cancelled"},
			Default:  func() any { return "draft" },
			ReadOnly: true,
		},
		{
			Name:        "accounting_period",
			Type:        def.FieldTypeLink,
			Label:       "Accounting Period",
			LinkTarget:  "finance_accounting_period",
		},
		{
			Name:        "currency",
			Type:        def.FieldTypeLink,
			Label:       "Currency",
			LinkTarget:  "finance_currency",
			Required:    true,
		},
		{
			Name:     "total_debit",
			Type:     def.FieldTypeCurrency,
			Label:    "Total Debit",
			ReadOnly: true,
			Computed: true,
		},
		{
			Name:     "total_credit",
			Type:     def.FieldTypeCurrency,
			Label:    "Total Credit",
			ReadOnly: true,
			Computed: true,
		},
		{
			Name:    "reference",
			Type:    def.FieldTypeData,
			Label:   "Reference",
			MaxLen:  255,
		},
		{
			Name:    "narration",
			Type:    def.FieldTypeSmallText,
			Label:   "Narration",
		},
		{
			Name:       "reversal_of",
			Type:       def.FieldTypeLink,
			Label:      "Reversal Of",
			Description: "Source entry that this entry reverses. Set by the reverse action.",
			LinkTarget: "finance_journal_entry",
			ReadOnly:   true,
		},
	},

	Actions: []def.ActionDef{
		{
			Name:           "submit",
			Label:          "Submit",
			Description:    "Submit entry for review. Transitions from draft to submitted.",
			ConfirmMessage: "Submit this journal entry for review?",
			Icon:           "send",
			WorkflowEvent:  def.EventType("submit"),
			HandlerFunc:    stubAction("submit"),
		},
		{
			Name:           "post",
			Label:          "Post",
			Description:    "Post entry to the General Ledger. Transitions from submitted to posted. Entry becomes immutable.",
			ConfirmMessage: "Post this journal entry to the General Ledger? This action cannot be undone.",
			Icon:           "check",
			WorkflowEvent:  def.EventType("post"),
			HandlerFunc:    stubAction("post"),
		},
		{
			Name:           "reverse",
			Label:          "Reverse",
			Description:    "Create a reversing entry with opposite debit/credit amounts.",
			ConfirmMessage: "Create a reversal entry for this journal entry?",
			Icon:           "rotate-ccw",
			HandlerFunc:    stubAction("reverse"),
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"finance.journal_entry.create"},
		Read:   []string{"finance.journal_entry.read"},
		Write:  []string{"finance.journal_entry.update"},
		Delete: []string{"finance.journal_entry.delete"},
		Actions: map[string][]string{
			"submit":  {"finance.journal_entry.submit"},
			"post":    {"finance.journal_entry.post"},
			"reverse": {"finance.journal_entry.reverse"},
		},
	},
}

func init() {
	def.Register(&JournalEntryDefinition)
}
