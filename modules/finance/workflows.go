package finance

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// ─── Workflow input/output types ──────────────────────────────────────────────

// InvoiceApprovalInput is the payload passed to InvoiceApprovalWorkflow.
type InvoiceApprovalInput struct {
	TenantID  uuid.UUID
	InvoiceID uuid.UUID
	ActorID   uuid.UUID
}

// PaymentProcessingInput is the payload passed to PaymentProcessingWorkflow.
type PaymentProcessingInput struct {
	TenantID  uuid.UUID
	PaymentID uuid.UUID
	ActorID   uuid.UUID
}

// JournalEntryApprovalInput is the payload passed to JournalEntryApprovalWorkflow.
type JournalEntryApprovalInput struct {
	TenantID       uuid.UUID
	JournalEntryID uuid.UUID
	ActorID        uuid.UUID
}

// ─── Workflow functions ───────────────────────────────────────────────────────
//
// Temporal determinism rules:
//   - No time.Now()    → use workflow.Now(ctx)
//   - No time.Sleep()  → use workflow.Sleep(ctx, d)
//   - No rand          → use workflow.SideEffect
//   - No direct I/O    → delegate to activities
//   - No goroutines    → use workflow.Go

// InvoiceApprovalWorkflow manages the invoice approval gate.
// Waits for an "approve" or "reject" signal (7-day timeout), then
// creates the GL journal entry and sends notification.
func InvoiceApprovalWorkflow(ctx workflow.Context, input InvoiceApprovalInput) error {
	// TODO: set workflow.WithActivityOptions (retry policy, start-to-close timeout)
	// TODO: create signal channel: approveSignal := workflow.GetSignalChannel(ctx, "approve")
	// TODO: create signal channel: rejectSignal := workflow.GetSignalChannel(ctx, "reject")
	// TODO: selector := workflow.NewSelector(ctx)
	// TODO: selector.AddReceive(approveSignal, ...) → call CreateInvoiceJournalEntryActivity
	// TODO: selector.AddReceive(rejectSignal, ...) → call SendInvoiceRejectedNotificationActivity
	// TODO: selector.AddFuture(workflow.NewTimer(ctx, 7*24*time.Hour), ...) → escalate
	// TODO: selector.Select(ctx)
	_ = ctx
	_ = input
	return nil
}

// PaymentProcessingWorkflow processes a submitted payment end-to-end.
// Creates the GL journal entry and confirms payment.
func PaymentProcessingWorkflow(ctx workflow.Context, input PaymentProcessingInput) error {
	// TODO: call ValidatePaymentActivity
	// TODO: call CreatePaymentJournalEntryActivity
	// TODO: call UpdatePaymentStatusActivity (status → "processed")
	// TODO: call SendPaymentConfirmationActivity
	_ = ctx
	_ = input
	return nil
}

// JournalEntryApprovalWorkflow handles multi-step journal entry approval.
// Entries above a configurable threshold require manager sign-off.
func JournalEntryApprovalWorkflow(ctx workflow.Context, input JournalEntryApprovalInput) error {
	// TODO: call GetJournalEntryTotalActivity to read total_debit
	// TODO: call GetApprovalThresholdActivity from tenant settings
	// TODO: if total < threshold → call PostJournalEntryActivity directly
	// TODO: if total >= threshold → wait for approval signal (manager sign-off)
	_ = ctx
	_ = input
	return nil
}

// ─── Activity implementations ─────────────────────────────────────────────────

// FinanceActivities holds Finance module activity dependencies.
// Inject at worker startup via wire before registering with Temporal worker.
type FinanceActivities struct {
	Posting *PostingService
	// TODO: add NotificationClient, EmailClient when Notification module is ready
}

// NewFinanceActivities constructs FinanceActivities with required dependencies.
func NewFinanceActivities(posting *PostingService) *FinanceActivities {
	return &FinanceActivities{Posting: posting}
}

// CreateInvoiceJournalEntryActivity creates the accounting journal entry for an approved invoice.
// Debits Accounts Receivable; credits Revenue accounts per invoice line.
func (a *FinanceActivities) CreateInvoiceJournalEntryActivity(ctx context.Context, invoiceID uuid.UUID) error {
	_ = activity.GetInfo(ctx)
	// TODO: load invoice and its lines from EntityRepository
	// TODO: load revenue GL accounts per line (account_id field on InvoiceLine)
	// TODO: load AR account from tenant settings
	// TODO: create JournalEntry with journal_type="Sales", status="draft"
	// TODO: create JournalEntryLine: debit AR account for invoice total
	// TODO: create JournalEntryLines: credit revenue accounts per line
	// TODO: create TaxEntry records for each line with tax
	// TODO: call Posting.PostJournalEntry to finalise
	// TODO: update invoice.journal_entry_id
	return fmt.Errorf("CreateInvoiceJournalEntryActivity: not yet implemented")
}

// CreatePaymentJournalEntryActivity creates the accounting journal entry for a payment.
func (a *FinanceActivities) CreatePaymentJournalEntryActivity(ctx context.Context, paymentID uuid.UUID) error {
	_ = activity.GetInfo(ctx)
	// TODO: load payment from EntityRepository
	// TODO: load payment method GL account
	// TODO: determine debit/credit split based on payment_type (receive=debit bank, credit AR)
	// TODO: create JournalEntry with journal_type based on payment method type
	// TODO: create JournalEntryLines
	// TODO: call Posting.PostJournalEntry
	// TODO: update payment.journal_entry_id and status="processed"
	return fmt.Errorf("CreatePaymentJournalEntryActivity: not yet implemented")
}

// PostJournalEntryActivity calls PostingService to post a prepared journal entry.
func (a *FinanceActivities) PostJournalEntryActivity(ctx context.Context, entryID uuid.UUID) error {
	_ = activity.GetInfo(ctx)
	return a.Posting.PostJournalEntry(ctx, entryID)
}
