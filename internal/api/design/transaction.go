package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// FINANCIAL TRANSACTION TYPES
// ============================================================================

// FinanceTransaction represents a financial transaction with double-entry bookkeeping.
// Supports multi-currency operations, approval workflows, and comprehensive audit trails.
var FinanceTransaction = ResultType("application/vnd.erp.finance.transaction", func() {
	Description("Financial transaction with double-entry bookkeeping support, multi-currency capabilities, and comprehensive workflow management")

	Attributes(func() {
		Field(1, "id", String, "Unique transaction identifier", func() {
			Format(FormatUUID)
			Example("txn-123e4567-e89b-12d3-a456-426614174000")
			Description("Primary key for transaction references")
		})

		Field(2, "tenant_id", String, "Associated tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
			Description("Tenant isolation boundary")
		})

		Field(3, "entity_id", String, "Associated entity identifier", func() {
			Format(FormatUUID)
			Example("987fcdeb-51d2-43b8-a456-426614174000")
			Description("Entity that owns this transaction")
		})

		Field(4, "transaction_number", String, "Human-readable transaction number", func() {
			Pattern("^[A-Z0-9-]{6,20}$")
			Example("JE-2023-001234")
			Description("Sequential or formatted transaction identifier")
		})

		Field(5, "transaction_type", String, "Classification of transaction", func() {
			Enum("JOURNAL_ENTRY", "PAYMENT", "RECEIPT", "TRANSFER", "ADJUSTMENT",
				"ACCRUAL", "REVERSAL", "RECURRING", "ALLOCATION", "REVALUATION")
			Example("JOURNAL_ENTRY")
			Description("Business type determining transaction behavior")
		})

		Field(6, "transaction_status", String, "Current processing status", func() {
			Enum("DRAFT", "PENDING_APPROVAL", "APPROVED", "POSTED", "VOID",
				"REVERSED", "FAILED", "CANCELLED")
			Default("DRAFT")
			Example("POSTED")
			Description("Workflow state and posting status")
		})

		Field(7, "transaction_date", String, "Business transaction date", func() {
			Format(FormatDateTime)
			Example("2023-12-07T00:00:00Z")
			Description("Date when business event occurred")
		})

		Field(8, "posting_date", String, "General ledger posting date", func() {
			Format(FormatDateTime)
			Example("2023-12-07T15:30:00Z")
			Description("When transaction was posted to GL")
		})

		Field(9, "value_date", String, "Economic value date", func() {
			Format(FormatDateTime)
			Example("2023-12-07T00:00:00Z")
			Description("Date for financial calculations and interest")
		})

		Field(10, "description", String, "Transaction description", func() {
			MinLength(3)
			MaxLength(500)
			Example("Monthly office rent payment for December 2023")
			Description("Business purpose and details")
		})

		Field(11, "reference_number", String, "External reference identifier", func() {
			MaxLength(50)
			Example("INV-2023-5678")
			Description("Related document or external system reference")
		})

		Field(12, "source_document", String, "Originating document type", func() {
			Enum("INVOICE", "RECEIPT", "CONTRACT", "BANK_STATEMENT", "EXPENSE_REPORT",
				"PAYROLL", "MANUAL_ENTRY", "SYSTEM_GENERATED", "IMPORT")
			Example("INVOICE")
			Description("Type of source document")
		})

		Field(13, "currency_code", String, "Transaction currency", func() {
			Pattern("^[A-Z]{3}$")
			Default("USD")
			Example("USD")
			Description("ISO 4217 currency code for amounts")
		})

		Field(14, "exchange_rate", String, "Currency exchange rate", func() {
			Pattern("^\\d+(\\.\\d{1,6})?$")
			Default("1.000000")
			Example("1.085420")
			Description("Rate to functional currency (if different)")
		})

		Field(15, "functional_currency", String, "Entity's functional currency", func() {
			Pattern("^[A-Z]{3}$")
			Default("USD")
			Example("USD")
			Description("Entity's reporting currency")
		})

		Field(16, "total_debit_amount", String, "Total debit entries", func() {
			Pattern("^\\d+(\\.\\d{1,4})?$")
			Example("5000.00")
			Description("Sum of all debit amounts in transaction currency")
		})

		Field(17, "total_credit_amount", String, "Total credit entries", func() {
			Pattern("^\\d+(\\.\\d{1,4})?$")
			Example("5000.00")
			Description("Sum of all credit amounts in transaction currency")
		})

		Field(18, "total_functional_debit", String, "Total debits in functional currency", func() {
			Pattern("^\\d+(\\.\\d{1,4})?$")
			Example("5427.10")
			Description("Debit total converted to functional currency")
		})

		Field(19, "total_functional_credit", String, "Total credits in functional currency", func() {
			Pattern("^\\d+(\\.\\d{1,4})?$")
			Example("5427.10")
			Description("Credit total converted to functional currency")
		})

		Field(20, "batch_id", String, "Transaction batch identifier", func() {
			Format(FormatUUID)
			Example("batch-456e7890-e89b-12d3-a456-426614174000")
			Description("Groups related transactions for processing")
		})

		Field(21, "approval_required", Boolean, "Requires approval workflow", func() {
			Default(false)
			Example(true)
			Description("Whether transaction needs approval before posting")
		})

		Field(22, "approval_status", String, "Approval workflow status", func() {
			Enum("NOT_REQUIRED", "PENDING", "APPROVED", "REJECTED", "EXPIRED")
			Default("NOT_REQUIRED")
			Example("APPROVED")
			Description("Current approval state")
		})

		Field(23, "approved_by", String, "Approver user identifier", func() {
			Format(FormatUUID)
			Example("approver-789abc12-def3-4567-890a-bcdef1234567")
			Description("User who approved the transaction")
		})

		Field(24, "approved_at", String, "Approval timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-07T14:00:00Z")
			Description("When transaction was approved")
		})

		Field(25, "is_reversed", Boolean, "Reversal status flag", func() {
			Default(false)
			Example(false)
			Description("Whether this transaction has been reversed")
		})

		Field(26, "reversed_by_transaction_id", String, "Reversing transaction reference", func() {
			Format(FormatUUID)
			Example("reverse-txn-abc123de-f456-7890-abc1-23def4567890")
			Description("Transaction that reversed this one")
		})

		Field(27, "reversal_reason", String, "Reason for reversal", func() {
			MaxLength(500)
			Example("Correcting data entry error in account allocation")
			Description("Business justification for reversal")
		})

		Field(28, "validation_status", String, "Data validation state", func() {
			Enum("VALID", "INVALID", "WARNINGS", "PENDING_VALIDATION")
			Default("PENDING_VALIDATION")
			Example("VALID")
			Description("Result of business rule validation")
		})

		Field(29, "validation_errors", ArrayOf(String), "Validation error messages", func() {
			Example([]string{
				"Debit and credit totals do not balance",
				"Transaction date is in a closed period",
			})
			Description("List of validation issues")
		})

		Field(30, "reconciliation_status", String, "Bank reconciliation status", func() {
			Enum("NOT_APPLICABLE", "PENDING", "RECONCILED", "EXCEPTION")
			Default("NOT_APPLICABLE")
			Example("RECONCILED")
			Description("Status for bank reconciliation process")
		})

		Field(31, "reconciled_at", String, "Reconciliation timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-08T10:00:00Z")
			Description("When transaction was reconciled")
		})

		Field(32, "cost_center", String, "Cost center allocation", func() {
			Pattern("^[A-Z0-9-]{3,20}$")
			Example("CC-ENG-001")
			Description("Primary cost center for transaction")
		})

		Field(33, "project_id", String, "Associated project identifier", func() {
			Format(FormatUUID)
			Example("project-123e4567-e89b-12d3-a456-426614174000")
			Description("Project tracking for transaction")
		})

		Field(34, "department_id", String, "Associated department identifier", func() {
			Format(FormatUUID)
			Example("dept-456e7890-e89b-12d3-a456-426614174000")
			Description("Department allocation for transaction")
		})

		Field(35, "tags", ArrayOf(String), "Transaction classification tags", func() {
			Example([]string{"rent", "facilities", "monthly_recurring"})
			Description("Searchable tags for categorization")
		})

		Field(36, "attachments", ArrayOf(String), "Document attachment references", func() {
			Elem(func() {
				Format(FormatUUID)
			})
			Example([]string{
				"doc-123e4567-e89b-12d3-a456-426614174000",
				"doc-456e7890-e89b-12d3-a456-426614174000",
			})
			Description("Supporting documents and receipts")
		})

		Field(37, "notes", String, "Additional transaction notes", func() {
			MaxLength(1000)
			Example("Processed by automated rent payment system. Next payment due 2024-01-07.")
			Description("Free-form notes and comments")
		})

		Field(38, "external_references", MapOf(String, String), "External system mappings", func() {
			Example(map[string]any{
				"quickbooks_id": "QB-TXN-789123",
				"bank_ref":      "BANK-REF-456789",
				"erp_import_id": "IMP-2023-001234",
			})
			Description("Integration references to external systems")
		})

		// Audit fields from common.go
		AuditFields()

		Required("id", "tenant_id", "entity_id", "transaction_number", "transaction_type",
			"transaction_status", "transaction_date", "description", "currency_code",
			"total_debit_amount", "total_credit_amount", "created_at")
	})

	View("default", func() {
		Description("Standard transaction view for listings and general display")
		Attribute("id")
		Attribute("transaction_number")
		Attribute("transaction_type")
		Attribute("transaction_status")
		Attribute("transaction_date")
		Attribute("description")
		Attribute("total_debit_amount")
		Attribute("total_credit_amount")
		Attribute("currency_code")
		Attribute("created_at")
	})

	View("detailed", func() {
		Description("Complete transaction view with all workflow and audit information")
		Attribute("id")
		Attribute("tenant_id")
		Attribute("entity_id")
		Attribute("transaction_number")
		Attribute("transaction_type")
		Attribute("transaction_status")
		Attribute("transaction_date")
		Attribute("posting_date")
		Attribute("value_date")
		Attribute("description")
		Attribute("reference_number")
		Attribute("source_document")
		Attribute("currency_code")
		Attribute("exchange_rate")
		Attribute("functional_currency")
		Attribute("total_debit_amount")
		Attribute("total_credit_amount")
		Attribute("total_functional_debit")
		Attribute("total_functional_credit")
		Attribute("batch_id")
		Attribute("approval_required")
		Attribute("approval_status")
		Attribute("approved_by")
		Attribute("approved_at")
		Attribute("is_reversed")
		Attribute("validation_status")
		Attribute("reconciliation_status")
		Attribute("cost_center")
		Attribute("project_id")
		Attribute("department_id")
		Attribute("tags")
		Attribute("attachments")
		Attribute("notes")
		Attribute("external_references")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})

	View("summary", func() {
		Description("Minimal view for references and quick lookups")
		Attribute("id")
		Attribute("transaction_number")
		Attribute("transaction_type")
		Attribute("transaction_date")
		Attribute("description")
		Attribute("total_debit_amount")
		Attribute("currency_code")
	})

	View("approval", func() {
		Description("Approval workflow view for managers and reviewers")
		Attribute("id")
		Attribute("transaction_number")
		Attribute("transaction_type")
		Attribute("transaction_date")
		Attribute("description")
		Attribute("total_debit_amount")
		Attribute("currency_code")
		Attribute("approval_required")
		Attribute("approval_status")
		Attribute("created_by")
		Attribute("created_at")
	})
})

// TransactionEntry represents individual debit/credit entries within a transaction.
var TransactionEntry = Type("TransactionEntry", func() {
	Description("Individual journal entry line item with account allocation and analytical dimensions")

	Field(1, "id", String, "Unique entry identifier", func() {
		Format(FormatUUID)
		Example("entry-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for entry record")
	})

	Field(2, "transaction_id", String, "Parent transaction identifier", func() {
		Format(FormatUUID)
		Example("txn-456e7890-e89b-12d3-a456-426614174000")
		Description("Transaction containing this entry")
	})

	Field(3, "entry_number", UInt, "Entry sequence within transaction", func() {
		Minimum(1)
		Maximum(1000)
		Example(1)
		Description("Ordering of entries within transaction")
	})

	Field(4, "account_id", String, "General ledger account identifier", func() {
		Format(FormatUUID)
		Example("account-789abc12-def3-4567-890a-bcdef1234567")
		Description("Account receiving the debit or credit")
	})

	Field(5, "account_code", String, "Account code for reference", func() {
		Pattern("^[0-9]{1,4}(-[0-9]{1,4})*$")
		Example("1100-001")
		Description("Human-readable account identifier")
	})

	Field(6, "account_name", String, "Account name for display", func() {
		MaxLength(100)
		Example("Cash - Operating Account")
		Description("Account description for UI display")
	})

	Field(7, "debit_amount", String, "Debit amount in transaction currency", func() {
		Pattern("^\\d+(\\.\\d{1,4})?$")
		Example("5000.00")
		Description("Amount if entry is a debit (positive)")
	})

	Field(8, "credit_amount", String, "Credit amount in transaction currency", func() {
		Pattern("^\\d+(\\.\\d{1,4})?$")
		Example("0.00")
		Description("Amount if entry is a credit (positive)")
	})

	Field(9, "functional_debit_amount", String, "Debit in functional currency", func() {
		Pattern("^\\d+(\\.\\d{1,4})?$")
		Example("5427.10")
		Description("Debit amount converted to functional currency")
	})

	Field(10, "functional_credit_amount", String, "Credit in functional currency", func() {
		Pattern("^\\d+(\\.\\d{1,4})?$")
		Example("0.00")
		Description("Credit amount converted to functional currency")
	})

	Field(11, "description", String, "Entry-specific description", func() {
		MaxLength(500)
		Example("Office rent payment - December 2023")
		Description("Detailed description for this entry")
	})

	Field(12, "reference", String, "Entry reference identifier", func() {
		MaxLength(100)
		Example("RENT-DEC-2023")
		Description("Additional reference for entry")
	})

	Field(13, "cost_center", String, "Cost center allocation", func() {
		Pattern("^[A-Z0-9-]{3,20}$")
		Example("CC-ADMIN-001")
		Description("Cost center for management reporting")
	})

	Field(14, "department", String, "Department allocation", func() {
		Pattern("^[A-Z0-9-]{3,20}$")
		Example("DEPT-ADMIN")
		Description("Department code for entry")
	})

	Field(15, "project_id", String, "Project identifier", func() {
		Format(FormatUUID)
		Example("project-123e4567-e89b-12d3-a456-426614174000")
		Description("Project tracking for entry")
	})

	Field(16, "product_line", String, "Product line allocation", func() {
		Pattern("^[A-Z0-9-]{3,20}$")
		Example("PROD-SOFTWARE")
		Description("Product line for profitability analysis")
	})

	Field(17, "customer_id", String, "Customer reference", func() {
		Format(FormatUUID)
		Example("customer-456e7890-e89b-12d3-a456-426614174000")
		Description("Customer for revenue/receivable entries")
	})

	Field(18, "vendor_id", String, "Vendor reference", func() {
		Format(FormatUUID)
		Example("vendor-789abc12-def3-4567-890a-bcdef1234567")
		Description("Vendor for expense/payable entries")
	})

	Field(19, "tax_code", String, "Tax code for entry", func() {
		Pattern("^[A-Z0-9-]{2,10}$")
		Example("VAT-STD")
		Description("Tax treatment for this entry")
	})

	Field(20, "tax_amount", String, "Tax amount", func() {
		Pattern("^\\d+(\\.\\d{1,4})?$")
		Example("800.00")
		Description("Tax portion of entry amount")
	})

	Field(21, "quantity", String, "Quantity for unit-based entries", func() {
		Pattern("^\\d+(\\.\\d{1,6})?$")
		Example("100.000000")
		Description("Quantity for inventory or unit-based transactions")
	})

	Field(22, "unit_price", String, "Unit price for quantity-based entries", func() {
		Pattern("^\\d+(\\.\\d{1,4})?$")
		Example("50.00")
		Description("Price per unit for calculations")
	})

	Field(23, "due_date", String, "Payment or collection due date", func() {
		Format(FormatDateTime)
		Example("2024-01-07T00:00:00Z")
		Description("When payment is due (for payables/receivables)")
	})

	Field(24, "analytical_tags", ArrayOf(String), "Analytical dimension tags", func() {
		Example([]string{"facilities", "recurring", "essential"})
		Description("Tags for multi-dimensional analysis")
	})

	Field(25, "custom_fields", MapOf(String, Any), "Entry-specific custom data", func() {
		Example(map[string]any{
			"lease_period":   "2023-12",
			"square_footage": 2500,
			"rate_per_sqft":  2.00,
		})
		Description("Flexible fields for business-specific data")
	})

	Field(26, "reconciliation_ref", String, "Bank reconciliation reference", func() {
		MaxLength(100)
		Example("BANK-STMT-20231207-001")
		Description("Reference for bank reconciliation matching")
	})

	// Audit fields
	AuditFields()

	Required("id", "transaction_id", "entry_number", "account_id", "description", "created_at")
})

// TransactionBatch represents a group of related transactions processed together.
var TransactionBatch = Type("TransactionBatch", func() {
	Description("Batch processing container for related financial transactions")

	Field(1, "id", String, "Unique batch identifier", func() {
		Format(FormatUUID)
		Example("batch-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for batch record")
	})

	Field(2, "tenant_id", String, "Associated tenant identifier", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Description("Tenant isolation boundary")
	})

	Field(3, "batch_number", String, "Human-readable batch number", func() {
		Pattern("^[A-Z0-9-]{6,20}$")
		Example("BATCH-2023-001234")
		Description("Sequential batch identifier")
	})

	Field(4, "batch_type", String, "Type of batch processing", func() {
		Enum("MANUAL", "AUTOMATED", "IMPORT", "RECURRING", "PAYROLL", "ALLOCATION", "ADJUSTMENT")
		Example("IMPORT")
		Description("Source and nature of batch")
	})

	Field(5, "batch_status", String, "Current batch processing status", func() {
		Enum("PENDING", "PROCESSING", "COMPLETED", "FAILED", "CANCELLED", "PARTIALLY_PROCESSED")
		Default("PENDING")
		Example("COMPLETED")
		Description("Overall batch processing state")
	})

	Field(6, "description", String, "Batch description", func() {
		MaxLength(500)
		Example("Monthly payroll processing for December 2023")
		Description("Business purpose of batch")
	})

	Field(7, "total_transactions", UInt, "Number of transactions in batch", func() {
		Example(247)
		Description("Count of transactions processed")
	})

	Field(8, "successful_transactions", UInt, "Successfully processed transactions", func() {
		Example(245)
		Description("Count of transactions completed without errors")
	})

	Field(9, "failed_transactions", UInt, "Failed transaction count", func() {
		Example(2)
		Description("Count of transactions with processing errors")
	})

	Field(10, "total_debit_amount", String, "Total debit amount for batch", func() {
		Pattern("^\\d+(\\.\\d{1,4})?$")
		Example("125000.00")
		Description("Sum of all debit amounts in batch")
	})

	Field(11, "total_credit_amount", String, "Total credit amount for batch", func() {
		Pattern("^\\d+(\\.\\d{1,4})?$")
		Example("125000.00")
		Description("Sum of all credit amounts in batch")
	})

	Field(12, "currency", String, "Batch currency", func() {
		Pattern("^[A-Z]{3}$")
		Example("USD")
		Description("Primary currency for batch totals")
	})

	Field(13, "processing_started_at", String, "Batch processing start time", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:00:00Z")
		Description("When batch processing began")
	})

	Field(14, "processing_completed_at", String, "Batch processing completion time", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:15:30Z")
		Description("When batch processing finished")
	})

	Field(15, "error_summary", ArrayOf(String), "Processing error messages", func() {
		Example([]string{
			"Transaction TXN-001: Invalid account code",
			"Transaction TXN-047: Amount exceeds approval limit",
		})
		Description("List of errors encountered during processing")
	})

	// Audit fields
	AuditFields()

	Required("id", "tenant_id", "batch_number", "batch_type", "batch_status",
		"description", "total_transactions", "currency", "created_at")
})
