> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Workflow Testing
portal: 4 — Backend Engineering
section: 00-module-development-guide/11-temporal-workflows
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-workflow-design.md
    title: Workflow Design
  - path: ./03-activity-design.md
    title: Activity Design
---

# Workflow Testing

## Test Environment

Use Temporal's `testsuite` package — it runs workflows synchronously without a real Temporal server:

```go
// internal/core/contracts/workflows/approval_workflow_test.go
package workflows_test

import (
    "testing"

    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
    "go.temporal.io/sdk/testsuite"

    "awo.so/internal/core/contracts/activities"
    "awo.so/internal/core/contracts/workflows"
)

type ApprovalWorkflowTestSuite struct {
    suite.Suite
    testsuite.WorkflowTestSuite
    env *testsuite.TestWorkflowEnvironment
}

func (s *ApprovalWorkflowTestSuite) SetupTest() {
    s.env = s.NewTestWorkflowEnvironment()
    s.env.RegisterWorkflow(workflows.ContractApprovalWorkflow)
}

func (s *ApprovalWorkflowTestSuite) TearDownTest() {
    s.env.AssertExpectations(s.T())
}

func TestApprovalWorkflow(t *testing.T) {
    suite.Run(t, new(ApprovalWorkflowTestSuite))
}
```

## Test: Happy Path — Reviewer Approves

```go
func (s *ApprovalWorkflowTestSuite) TestApproveHappyPath() {
    input := workflows.ApprovalWorkflowInput{
        TenantID:       "aaaaaaaa-0000-0000-0000-000000000001",
        ContractID:     "bbbbbbbb-0000-0000-0000-000000000001",
        ContractNumber: "CONT-2025-0001",
        SubmitterID:    "cccccccc-0000-0000-0000-000000000001",
    }

    // Mock NotifyReviewers activity
    s.env.OnActivity(activities.NotifyReviewers, mock.Anything, mock.Anything).
        Return(nil)

    // Mock ApproveContract activity
    s.env.OnActivity(activities.ApproveContract, mock.Anything, mock.MatchedBy(func(i activities.ApproveContractInput) bool {
        return i.ContractID == input.ContractID
    })).Return(nil)

    // Send signal BEFORE starting — test env queues it
    s.env.RegisterDelayedCallback(func() {
        s.env.SignalWorkflow(workflows.SignalApprovalDecision, workflows.ApprovalDecision{
            ReviewerID: "dddddddd-0000-0000-0000-000000000001",
            Decision:   "approve",
            Comment:    "Looks good",
        })
    }, 0)

    s.env.ExecuteWorkflow(workflows.ContractApprovalWorkflow, input)

    s.True(s.env.IsWorkflowCompleted())
    s.NoError(s.env.GetWorkflowError())
}
```

## Test: Reviewer Rejects

```go
func (s *ApprovalWorkflowTestSuite) TestRejectPath() {
    // ... setup input ...

    s.env.OnActivity(activities.NotifyReviewers, mock.Anything, mock.Anything).Return(nil)
    s.env.OnActivity(activities.RejectContract, mock.Anything, mock.Anything).Return(nil)

    s.env.RegisterDelayedCallback(func() {
        s.env.SignalWorkflow(workflows.SignalApprovalDecision, workflows.ApprovalDecision{
            Decision: "reject",
            Comment:  "Needs revision",
        })
    }, 0)

    s.env.ExecuteWorkflow(workflows.ContractApprovalWorkflow, input)

    s.True(s.env.IsWorkflowCompleted())
    s.NoError(s.env.GetWorkflowError())
}
```

## Test: Timeout — Escalation

```go
func (s *ApprovalWorkflowTestSuite) TestTimeout_EscalatesAfter7Days() {
    // ... setup input ...

    s.env.OnActivity(activities.NotifyReviewers, mock.Anything, mock.Anything).Return(nil)
    s.env.OnActivity(activities.EscalateTimedOut, mock.Anything, mock.Anything).Return(nil)

    // Do NOT send signal — let the timer fire
    // TestWorkflowEnvironment skips time automatically when no callbacks remain

    s.env.ExecuteWorkflow(workflows.ContractApprovalWorkflow, input)

    s.True(s.env.IsWorkflowCompleted())
    s.NoError(s.env.GetWorkflowError())

    // EscalateTimedOut must have been called
    s.env.AssertActivityCalled(s.T(), "EscalateTimedOut", mock.Anything, mock.Anything)
}
```

## Activity Testing

Activities are regular functions — test them with a real test database or mock dependencies:

```go
func TestContractActivities_ApproveContract(t *testing.T) {
    store := testutil.NewTestStore(t)
    repo  := repository.NewContractRepository(store)
    acts  := activities.NewContractActivities(repo, &noopNotifSvc{})

    // Create a submitted contract
    tenantID  := testutil.CreateTestTenant(t, store)
    entityID  := testutil.CreateTestEntity(t, store, tenantID)
    contract  := testutil.CreateTestContractWithStatus(t, store, tenantID, entityID, domain.ContractStatusSubmitted)

    err := acts.ApproveContract(context.Background(), activities.ApproveContractInput{
        TenantID:   tenantID.String(),
        ContractID: contract.ID.String(),
        ReviewerID: uuid.New().String(),
    })
    require.NoError(t, err)

    // Verify status changed
    fetched, err := repo.GetByID(context.Background(), contract.ID, tenantID)
    require.NoError(t, err)
    assert.Equal(t, domain.ContractStatusApproved, fetched.Status)
}
```
