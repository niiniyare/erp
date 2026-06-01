---
title: Workflow Design
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Temporal Overview](01-temporal-overview.md)"
  - "[Activity Design](03-activity-design.md)"
---

# Workflow Design

## ContractApprovalWorkflow

The approval workflow orchestrates the multi-step review process after a contract is submitted.

```go
// internal/core/contracts/workflows/approval_workflow.go
package workflows

import (
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "awo.so/internal/core/contracts/activities"
)

// ApprovalWorkflowInput is passed when the workflow is started.
type ApprovalWorkflowInput struct {
    TenantID       string
    ContractID     string
    ContractNumber string
    SubmitterID    string
}

// ApprovalDecision is sent via Signal when a reviewer acts.
type ApprovalDecision struct {
    ReviewerID string
    Decision   string // "approve" | "reject"
    Comment    string
}

const (
    SignalApprovalDecision = "approval_decision"
    QueryApprovalStatus   = "approval_status"
)

func ContractApprovalWorkflow(ctx workflow.Context, input ApprovalWorkflowInput) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("ContractApprovalWorkflow started", "contractID", input.ContractID)

    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Second,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 3,
            InitialInterval: 2 * time.Second,
        },
    }
    actCtx := workflow.WithActivityOptions(ctx, ao)

    // Step 1: Notify reviewers
    if err := workflow.ExecuteActivity(actCtx,
        activities.NotifyReviewers,
        activities.NotifyReviewersInput{
            TenantID:       input.TenantID,
            ContractID:     input.ContractID,
            ContractNumber: input.ContractNumber,
        },
    ).Get(ctx, nil); err != nil {
        logger.Error("NotifyReviewers activity failed", "error", err)
        // Non-fatal — workflow continues even if notification fails
    }

    // Step 2: Wait for reviewer decision (up to 7 days)
    var decision ApprovalDecision
    decisionChannel := workflow.GetSignalChannel(ctx, SignalApprovalDecision)

    selector := workflow.NewSelector(ctx)
    selector.AddReceive(decisionChannel, func(ch workflow.ReceiveChannel, more bool) {
        ch.Receive(ctx, &decision)
    })

    // 7-day timeout
    timerFuture := workflow.NewTimer(ctx, 7*24*time.Hour)
    selector.AddFuture(timerFuture, func(f workflow.Future) {
        // Timed out — escalate
        decision = ApprovalDecision{Decision: "timeout"}
    })

    selector.Select(ctx)

    // Step 3: Apply decision
    switch decision.Decision {
    case "approve":
        return workflow.ExecuteActivity(actCtx,
            activities.ApproveContract,
            activities.ApproveContractInput{
                TenantID:   input.TenantID,
                ContractID: input.ContractID,
                ReviewerID: decision.ReviewerID,
                Comment:    decision.Comment,
            },
        ).Get(ctx, nil)

    case "reject":
        return workflow.ExecuteActivity(actCtx,
            activities.RejectContract,
            activities.RejectContractInput{
                TenantID:    input.TenantID,
                ContractID:  input.ContractID,
                ReviewerID:  decision.ReviewerID,
                Comment:     decision.Comment,
                SubmitterID: input.SubmitterID,
            },
        ).Get(ctx, nil)

    case "timeout":
        return workflow.ExecuteActivity(actCtx,
            activities.EscalateTimedOut,
            activities.EscalateTimedOutInput{
                TenantID:   input.TenantID,
                ContractID: input.ContractID,
            },
        ).Get(ctx, nil)
    }

    return nil
}
```

## ContractExpiryWorkflow

Runs nightly via a cron schedule:

```go
// internal/core/contracts/workflows/expiry_workflow.go
package workflows

import (
    "time"

    "go.temporal.io/sdk/workflow"

    "awo.so/internal/core/contracts/activities"
)

type ExpiryWorkflowInput struct {
    TenantID  string
    DaysAhead int // check contracts expiring within this many days
}

func ContractExpiryWorkflow(ctx workflow.Context, input ExpiryWorkflowInput) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
    }
    actCtx := workflow.WithActivityOptions(ctx, ao)

    var contractIDs []string
    if err := workflow.ExecuteActivity(actCtx,
        activities.FindExpiringContracts,
        activities.FindExpiringContractsInput{
            TenantID:  input.TenantID,
            DaysAhead: input.DaysAhead,
        },
    ).Get(ctx, &contractIDs); err != nil {
        return err
    }

    for _, id := range contractIDs {
        // Fire-and-forget reminder — do not fail workflow if one reminder fails
        _ = workflow.ExecuteActivity(actCtx,
            activities.SendExpiryReminder,
            activities.SendExpiryReminderInput{
                TenantID:   input.TenantID,
                ContractID: id,
            },
        ).Get(ctx, nil)
    }

    return nil
}
```

## Starting a Workflow from Service Layer

```go
// In contractService.Submit
func (s *contractService) Submit(ctx context.Context, req SubmitContractRequest) (*domain.Contract, error) {
    // ... authorize, validate, persist ...

    // Start approval workflow
    workflowID := fmt.Sprintf("contract-approval-%s", contractID)
    _, err = s.temporalClient.ExecuteWorkflow(ctx,
        client.StartWorkflowOptions{
            ID:        workflowID,
            TaskQueue: workflows.ContractsTaskQueue,
        },
        workflows.ContractApprovalWorkflow,
        workflows.ApprovalWorkflowInput{
            TenantID:       tenantID.String(),
            ContractID:     contractID.String(),
            ContractNumber: updated.ContractNumber,
            SubmitterID:    req.UserID.String(),
        },
    )
    if err != nil {
        // Workflow start failure is non-fatal — log and continue
        s.log.Error().Err(err).Str("contract_id", contractID.String()).
            Msg("failed to start approval workflow")
    }

    return updated, nil
}
```

## Signaling a Workflow

```go
// Called from the reviewer's HTTP action handler
func (s *contractService) RecordApprovalDecision(ctx context.Context, req ApprovalDecisionRequest) error {
    workflowID := fmt.Sprintf("contract-approval-%s", req.ContractID)
    return s.temporalClient.SignalWorkflow(ctx,
        workflowID,
        "",  // runID empty = latest run
        workflows.SignalApprovalDecision,
        workflows.ApprovalDecision{
            ReviewerID: req.ReviewerID.String(),
            Decision:   req.Decision,
            Comment:    req.Comment,
        },
    )
}
```
