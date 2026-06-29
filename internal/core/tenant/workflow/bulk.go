package workflow

import (
	"fmt"
	"time"

	"awo.so/internal/core/tenant/domain"
	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const BulkOperationWorkflowID = "TenantBulkOperationWorkflow"

// BulkOperationInput describes a bulk tenant operation.
type BulkOperationInput struct {
	TenantIDs []uuid.UUID         `json:"tenant_ids"`
	Operation string              `json:"operation"` // "suspend", "activate", "archive", "delete"
	Status    domain.TenantStatus `json:"status,omitempty"`
	Reason    string              `json:"reason,omitempty"`
}

// BulkOperationResult contains the outcome of a bulk operation.
type BulkOperationResult struct {
	Succeeded int      `json:"succeeded"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors,omitempty"`
}

// BulkOperationWorkflow processes bulk tenant operations with parallel activities.
func BulkOperationWorkflow(ctx workflow.Context, input BulkOperationInput) (*BulkOperationResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting bulk tenant operation", "operation", input.Operation, "count", len(input.TenantIDs))

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 60 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	switch input.Operation {
	case "suspend", "activate":
		if err := workflow.ExecuteActivity(ctx, "BulkUpdateStatusActivity", input.TenantIDs, input.Status).Get(ctx, nil); err != nil {
			return nil, fmt.Errorf("bulk %s failed: %w", input.Operation, err)
		}
	case "delete":
		if err := workflow.ExecuteActivity(ctx, "BulkSoftDeleteActivity", input.TenantIDs).Get(ctx, nil); err != nil {
			return nil, fmt.Errorf("bulk delete failed: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown operation: %s", input.Operation)
	}

	return &BulkOperationResult{
		Succeeded: len(input.TenantIDs),
	}, nil
}
