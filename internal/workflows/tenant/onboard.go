// Package tenant provides workflow starters for tenant lifecycle operations.
// Starters are thin HTTP-facing clients that enqueue Temporal workflows;
// the workflow logic lives in internal/core/tenant/workflow/.
package tenant

import (
	"context"
	"fmt"
	"time"

	temporalclient "go.temporal.io/sdk/client"

	"awo.so/internal/core/tenant/domain"
	workflowpkg "awo.so/internal/core/tenant/workflow"
)

// Starter enqueues tenant Temporal workflows via the Temporal client.
type Starter struct {
	client temporalclient.Client
}

// NewStarter creates a workflow starter backed by the given Temporal client.
func NewStarter(client temporalclient.Client) *Starter {
	return &Starter{client: client}
}

// OnboardResult is returned after a provisioning workflow is accepted.
type OnboardResult struct {
	WorkflowID string `json:"workflow_id"`
	RunID      string `json:"run_id"`
}

// StartOnboarding enqueues a TenantProvisioningWorkflow and returns the workflow
// and run IDs so the caller can poll for completion. The workflow runs
// asynchronously — this call returns as soon as Temporal accepts the task.
func (s *Starter) StartOnboarding(ctx context.Context, input domain.ProvisioningInput) (*OnboardResult, error) {
	workflowID := fmt.Sprintf("tenant-onboard-%s-%d", input.Email, time.Now().UnixMilli())

	opts := temporalclient.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflowpkg.TaskQueue,
	}

	run, err := s.client.ExecuteWorkflow(ctx, opts, workflowpkg.ProvisioningWorkflowID, input)
	if err != nil {
		return nil, fmt.Errorf("enqueue provisioning workflow: %w", err)
	}

	return &OnboardResult{
		WorkflowID: run.GetID(),
		RunID:      run.GetRunID(),
	}, nil
}
