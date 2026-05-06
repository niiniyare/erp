package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"awo.so/internal/core/finance/domain"
	platformTemporal "awo.so/internal/platform/temporal"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
)

// ── TemporalIntegration ────────────────────────────────────────────────────────

// TemporalIntegration wires finance service methods into Temporal activities and
// workflows, then registers them with the platform WorkerManager.
type TemporalIntegration struct {
	txnService  TransactionService
	workerMgr   *platformTemporal.WorkerManager
}

// NewTemporalIntegration creates a TemporalIntegration ready to be registered.
func NewTemporalIntegration(txnService TransactionService, workerMgr *platformTemporal.WorkerManager) *TemporalIntegration {
	return &TemporalIntegration{
		txnService: txnService,
		workerMgr:  workerMgr,
	}
}

// Register pushes all finance activities and workflows into the platform's
// WorkerManager so they are picked up when workers start.
func (ti *TemporalIntegration) Register() {
	acts := &financeActivities{txnService: ti.txnService}

	ti.workerMgr.RegisterModuleActivities("finance", map[string]any{
		domain.ActivityTypeTransactionPosting:  acts.PostTransactionActivity,
		domain.ActivityTypeTransactionReversal: acts.ReversalActivity,
		domain.ActivityTypeTransactionCreation: acts.GenerateRecurringTransactionActivity,
		domain.ActivityTypeNotification:        acts.EscalateApprovalActivity,
	})

	ti.workerMgr.RegisterModuleWorkflows("finance", map[string]any{
		domain.WorkflowTypeTransactionPosting:  RecurringTransactionWorkflow,
		domain.WorkflowTypeTransactionApproval: ApprovalEscalationWorkflow,
		domain.WorkflowTypePeriodClosing:       PeriodEndWorkflow,
	})
}

// ── Activities ────────────────────────────────────────────────────────────────

// financeActivities holds dependencies for finance Temporal activities.
// Each exported method is a Temporal activity.
type financeActivities struct {
	txnService TransactionService
}

// PostTransactionInput is the input for PostTransactionActivity.
type PostTransactionActivityInput struct {
	TransactionID uuid.UUID  `json:"transaction_id"`
	PostingDate   *time.Time `json:"posting_date,omitempty"`
}

// PostTransactionActivity wraps TransactionService.PostTransaction for use as a
// Temporal activity. It inherits the configured retry policy from the workflow.
func (a *financeActivities) PostTransactionActivity(ctx context.Context, input PostTransactionActivityInput) error {
	logger.InfoContext(ctx, "PostTransactionActivity started",
		logger.Fields{"transaction_id": input.TransactionID.String()})

	_, err := a.txnService.PostTransaction(ctx, input.TransactionID, input.PostingDate)
	if err != nil {
		return fmt.Errorf("PostTransactionActivity: %w", err)
	}

	logger.InfoContext(ctx, "PostTransactionActivity completed",
		logger.Fields{"transaction_id": input.TransactionID.String()})
	return nil
}

// GenerateRecurringTransactionInput is the input for GenerateRecurringTransactionActivity.
type GenerateRecurringTransactionInput struct {
	TemplateTransactionID uuid.UUID `json:"template_transaction_id"`
	GenerationDate        time.Time `json:"generation_date"`
}

// GenerateRecurringTransactionActivity creates a new transaction from a recurring
// template and immediately posts it. Called by RecurringTransactionWorkflow.
func (a *financeActivities) GenerateRecurringTransactionActivity(ctx context.Context, input GenerateRecurringTransactionInput) (uuid.UUID, error) {
	activity.RecordHeartbeat(ctx, "generating")

	logger.InfoContext(ctx, "GenerateRecurringTransactionActivity started",
		logger.Fields{
			"template_id":     input.TemplateTransactionID.String(),
			"generation_date": input.GenerationDate.Format("2006-01-02"),
		})

	newTxn, err := a.txnService.CreateRecurringTransaction(ctx, input.TemplateTransactionID, input.GenerationDate)
	if err != nil {
		return uuid.Nil, fmt.Errorf("GenerateRecurringTransactionActivity: create: %w", err)
	}

	activity.RecordHeartbeat(ctx, "posting")

	generationDate := input.GenerationDate
	if _, err := a.txnService.PostTransaction(ctx, newTxn.ID, &generationDate); err != nil {
		return newTxn.ID, fmt.Errorf("GenerateRecurringTransactionActivity: post: %w", err)
	}

	logger.InfoContext(ctx, "GenerateRecurringTransactionActivity completed",
		logger.Fields{"new_transaction_id": newTxn.ID.String()})

	return newTxn.ID, nil
}

// EscalateApprovalInput is the input for EscalateApprovalActivity.
type EscalateApprovalInput struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	// SubmittedAt is the time the transaction was submitted for approval;
	// used to compute elapsed SLA time in the log.
	SubmittedAt time.Time `json:"submitted_at"`
}

// EscalateApprovalActivity is triggered by ApprovalEscalationWorkflow when the
// approval SLA (default 48 h) lapses. It rejects the transaction so it reverts
// to DRAFT, then records a structured log entry (the actual notification channel
// — email, Slack, etc. — is owned by the notification service).
func (a *financeActivities) EscalateApprovalActivity(ctx context.Context, input EscalateApprovalInput) error {
	logger.InfoContext(ctx, "EscalateApprovalActivity: SLA lapsed — rejecting approval",
		logger.Fields{
			"transaction_id": input.TransactionID.String(),
			"submitted_at":   input.SubmittedAt.Format(time.RFC3339),
			"elapsed":        time.Since(input.SubmittedAt).String(),
		})

	// Inject system identity so RejectTransaction's auth + SOD guards pass.
	// This rejection is system-initiated (SLA timer), not a human approver action.
	sysCtx := shared.WithUserID(ctx, shared.SystemUserID)

	// Reject the transaction with a system-generated SLA-expiry reason.
	// RejectTransaction moves the transaction back to DRAFT status.
	_, err := a.txnService.RejectTransaction(sysCtx, input.TransactionID, domain.RejectionReasonExpired, "Approval SLA expired — transaction automatically rejected and returned to DRAFT")
	if err != nil {
		return fmt.Errorf("EscalateApprovalActivity: reject: %w", err)
	}

	logger.InfoContext(ctx, "EscalateApprovalActivity completed — transaction returned to DRAFT",
		logger.Fields{"transaction_id": input.TransactionID.String()})

	return nil
}

// ReversalActivityInput is the input for ReversalActivity.
type ReversalActivityInput struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Reason        string    `json:"reason"`
}

// ReversalActivity performs an auto-reversal (e.g. accrual reversal on period
// open). Called by PeriodEndWorkflow.
func (a *financeActivities) ReversalActivity(ctx context.Context, input ReversalActivityInput) (uuid.UUID, error) {
	activity.RecordHeartbeat(ctx, "reversing")

	logger.InfoContext(ctx, "ReversalActivity started",
		logger.Fields{
			"transaction_id": input.TransactionID.String(),
			"reason":         input.Reason,
		})

	reversal, err := a.txnService.ReverseTransaction(ctx, input.TransactionID, input.Reason)
	if err != nil {
		return uuid.Nil, fmt.Errorf("ReversalActivity: %w", err)
	}

	logger.InfoContext(ctx, "ReversalActivity completed",
		logger.Fields{
			"original_id": input.TransactionID.String(),
			"reversal_id": reversal.ID.String(),
		})

	return reversal.ID, nil
}

// ── Workflows ─────────────────────────────────────────────────────────────────

// RecurringTransactionWorkflowInput is the cron workflow input.
type RecurringTransactionWorkflowInput struct {
	TemplateTransactionID uuid.UUID `json:"template_transaction_id"`
}

// RecurringTransactionWorkflow is a cron-scheduled workflow that generates a
// transaction from a recurring template and posts it.
// Schedule this workflow with a CronSchedule expression (e.g. "0 0 1 * *" for
// monthly on the 1st).
func RecurringTransactionWorkflow(ctx workflow.Context, input RecurringTransactionWorkflowInput) error {
	ao := workflow.ActivityOptions{
		TaskQueue:           domain.TemporalTaskQueueFinance,
		StartToCloseTimeout: domain.TransactionWorkflowTimeout,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: domain.DefaultRetryAttempts,
			InitialInterval: domain.StandardRetryBackoff,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	generationDate := workflow.Now(ctx)

	var newID uuid.UUID
	err := workflow.ExecuteActivity(ctx,
		(*financeActivities)(nil).GenerateRecurringTransactionActivity,
		GenerateRecurringTransactionInput{
			TemplateTransactionID: input.TemplateTransactionID,
			GenerationDate:        generationDate,
		}).Get(ctx, &newID)
	if err != nil {
		return fmt.Errorf("RecurringTransactionWorkflow: %w", err)
	}

	workflow.GetLogger(ctx).Info("RecurringTransactionWorkflow completed",
		"template_id", input.TemplateTransactionID.String(),
		"new_transaction_id", newID.String(),
		"generation_date", generationDate.Format("2006-01-02"),
	)
	return nil
}

// ApprovalEscalationWorkflowInput is the input for the approval escalation workflow.
type ApprovalEscalationWorkflowInput struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	SubmittedAt   time.Time `json:"submitted_at"`
	// SLADuration is the approval window; defaults to 48 h if zero.
	SLADuration time.Duration `json:"sla_duration,omitempty"`
}

// ApprovalEscalationWorkflow waits for the approval SLA to lapse then escalates.
// The workflow should be started when a transaction enters PENDING_APPROVAL.
// If the approver signals approval or rejection before SLA elapses, the workflow
// exits cleanly via the SignalApprovalGranted / SignalApprovalRejected channels.
func ApprovalEscalationWorkflow(ctx workflow.Context, input ApprovalEscalationWorkflowInput) error {
	sla := input.SLADuration
	if sla == 0 {
		sla = 48 * time.Hour
	}

	// Listen for early-resolution signals from the approval handler.
	approvedCh := workflow.GetSignalChannel(ctx, domain.SignalApprovalGranted)
	rejectedCh := workflow.GetSignalChannel(ctx, domain.SignalApprovalRejected)

	timerFired := false
	timerCtx, cancelTimer := workflow.WithCancel(ctx)
	timerCh := workflow.NewTimer(timerCtx, sla)

	selector := workflow.NewSelector(ctx)

	selector.AddReceive(approvedCh, func(c workflow.ReceiveChannel, more bool) {
		cancelTimer()
		workflow.GetLogger(ctx).Info("ApprovalEscalationWorkflow: approved before SLA",
			"transaction_id", input.TransactionID.String())
	})

	selector.AddReceive(rejectedCh, func(c workflow.ReceiveChannel, more bool) {
		cancelTimer()
		workflow.GetLogger(ctx).Info("ApprovalEscalationWorkflow: rejected before SLA",
			"transaction_id", input.TransactionID.String())
	})

	selector.AddFuture(timerCh, func(f workflow.Future) {
		if err := f.Get(ctx, nil); err == nil {
			timerFired = true
		}
	})

	selector.Select(ctx)

	if !timerFired {
		return nil // Resolved before SLA — nothing to escalate.
	}

	// SLA lapsed — run the escalation activity.
	ao := workflow.ActivityOptions{
		TaskQueue:           domain.TemporalTaskQueueFinanceHighPriority,
		StartToCloseTimeout: domain.NotificationActivityTimeout * 6, // 60 s
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: domain.NetworkRetryAttempts,
			InitialInterval: domain.NetworkRetryBackoff,
		},
	}
	actCtx := workflow.WithActivityOptions(ctx, ao)

	return workflow.ExecuteActivity(actCtx,
		(*financeActivities)(nil).EscalateApprovalActivity,
		EscalateApprovalInput{
			TransactionID: input.TransactionID,
			SubmittedAt:   input.SubmittedAt,
		}).Get(ctx, nil)
}

// PeriodEndWorkflowInput is the input for the period-end automation workflow.
type PeriodEndWorkflowInput struct {
	PeriodID uuid.UUID `json:"period_id"`
	TenantID uuid.UUID `json:"tenant_id"`
	// AccrualTransactionIDs lists accrual transactions that must be auto-reversed
	// on the first day of the new period.
	AccrualTransactionIDs []uuid.UUID `json:"accrual_transaction_ids,omitempty"`
}

// PeriodEndWorkflow is triggered when a period is soft-closed. It auto-reverses
// accrual entries (accrual reversal) in a fan-out pattern.
// Depreciation and other period-end calculations can be added as additional
// activity steps.
func PeriodEndWorkflow(ctx workflow.Context, input PeriodEndWorkflowInput) error {
	ao := workflow.ActivityOptions{
		TaskQueue:           domain.TemporalTaskQueueFinanceLongRunning,
		StartToCloseTimeout: domain.PeriodClosingWorkflowTimeout,
		HeartbeatTimeout:    1 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: domain.DefaultRetryAttempts,
			InitialInterval: domain.BulkRetryBackoff,
		},
	}
	actCtx := workflow.WithActivityOptions(ctx, ao)

	log := workflow.GetLogger(ctx)
	log.Info("PeriodEndWorkflow: starting accrual reversals",
		"period_id", input.PeriodID.String(),
		"accrual_count", len(input.AccrualTransactionIDs),
	)

	// Fan-out: reverse all accruals concurrently.
	futures := make([]workflow.Future, 0, len(input.AccrualTransactionIDs))
	for _, txnID := range input.AccrualTransactionIDs {
		f := workflow.ExecuteActivity(actCtx,
			(*financeActivities)(nil).ReversalActivity,
			ReversalActivityInput{
				TransactionID: txnID,
				Reason:        fmt.Sprintf("auto-reversal: period %s accrual", input.PeriodID),
			})
		futures = append(futures, f)
	}

	var errs []error
	for i, f := range futures {
		var reversalID uuid.UUID
		if err := f.Get(ctx, &reversalID); err != nil {
			errs = append(errs, fmt.Errorf("accrual[%d] %s: %w", i, input.AccrualTransactionIDs[i], err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("PeriodEndWorkflow: %d reversal(s) failed: %v", len(errs), errs)
	}

	log.Info("PeriodEndWorkflow completed",
		"period_id", input.PeriodID.String(),
		"reversed_count", len(input.AccrualTransactionIDs),
	)
	return nil
}
