package workflow_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/tenant/domain"
	"github.com/niiniyare/erp/internal/core/tenant/workflow"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

// ---------------------------------------------------------------------------
// BulkOperationWorkflowSuite
// ---------------------------------------------------------------------------

type BulkOperationWorkflowSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func TestBulkOperationWorkflowSuite(t *testing.T) {
	suite.Run(t, new(BulkOperationWorkflowSuite))
}

func (s *BulkOperationWorkflowSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivityWithOptions(
		func(ctx context.Context, ids []uuid.UUID, status domain.TenantStatus) error { return nil },
		activity.RegisterOptions{Name: "BulkUpdateStatusActivity"},
	)
	s.env.RegisterActivityWithOptions(
		func(ctx context.Context, ids []uuid.UUID) error { return nil },
		activity.RegisterOptions{Name: "BulkSoftDeleteActivity"},
	)
}

func (s *BulkOperationWorkflowSuite) AfterTest(_, _ string) {
	s.env.AssertExpectations(s.T())
}

// TN-WF-010: Bulk status update — suspend
func (s *BulkOperationWorkflowSuite) TestBulkSuspend_Success() {
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	input := workflow.BulkOperationInput{
		TenantIDs: ids,
		Operation: "suspend",
		Status:    domain.StatusSuspended,
	}

	s.env.OnActivity("BulkUpdateStatusActivity", mock.Anything, ids, domain.StatusSuspended).Return(nil)

	s.env.ExecuteWorkflow(workflow.BulkOperationWorkflow, input)

	require.True(s.T(), s.env.IsWorkflowCompleted())
	require.NoError(s.T(), s.env.GetWorkflowError())

	var result workflow.BulkOperationResult
	require.NoError(s.T(), s.env.GetWorkflowResult(&result))
	require.Equal(s.T(), 3, result.Succeeded)
	require.Equal(s.T(), 0, result.Failed)
}

// TN-WF-010: Bulk status update — activate
func (s *BulkOperationWorkflowSuite) TestBulkActivate_Success() {
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	input := workflow.BulkOperationInput{
		TenantIDs: ids,
		Operation: "activate",
		Status:    domain.StatusActive,
	}

	s.env.OnActivity("BulkUpdateStatusActivity", mock.Anything, ids, domain.StatusActive).Return(nil)

	s.env.ExecuteWorkflow(workflow.BulkOperationWorkflow, input)

	require.True(s.T(), s.env.IsWorkflowCompleted())
	require.NoError(s.T(), s.env.GetWorkflowError())

	var result workflow.BulkOperationResult
	require.NoError(s.T(), s.env.GetWorkflowResult(&result))
	require.Equal(s.T(), 2, result.Succeeded)
}

// TN-WF-010: Bulk status update — failure
func (s *BulkOperationWorkflowSuite) TestBulkSuspend_Error() {
	ids := []uuid.UUID{uuid.New()}
	input := workflow.BulkOperationInput{
		TenantIDs: ids,
		Operation: "suspend",
		Status:    domain.StatusSuspended,
	}

	s.env.OnActivity("BulkUpdateStatusActivity", mock.Anything, ids, domain.StatusSuspended).Return(fmt.Errorf("db error"))

	s.env.ExecuteWorkflow(workflow.BulkOperationWorkflow, input)

	require.True(s.T(), s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "bulk suspend failed")
}

// TN-WF-011: Bulk soft delete — success
func (s *BulkOperationWorkflowSuite) TestBulkDelete_Success() {
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	input := workflow.BulkOperationInput{
		TenantIDs: ids,
		Operation: "delete",
	}

	s.env.OnActivity("BulkSoftDeleteActivity", mock.Anything, ids).Return(nil)

	s.env.ExecuteWorkflow(workflow.BulkOperationWorkflow, input)

	require.True(s.T(), s.env.IsWorkflowCompleted())
	require.NoError(s.T(), s.env.GetWorkflowError())

	var result workflow.BulkOperationResult
	require.NoError(s.T(), s.env.GetWorkflowResult(&result))
	require.Equal(s.T(), 5, result.Succeeded)
}

// TN-WF-011: Bulk soft delete — failure
func (s *BulkOperationWorkflowSuite) TestBulkDelete_Error() {
	ids := []uuid.UUID{uuid.New()}
	input := workflow.BulkOperationInput{
		TenantIDs: ids,
		Operation: "delete",
	}

	s.env.OnActivity("BulkSoftDeleteActivity", mock.Anything, ids).Return(fmt.Errorf("delete failed"))

	s.env.ExecuteWorkflow(workflow.BulkOperationWorkflow, input)

	require.True(s.T(), s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "bulk delete failed")
}

// Unknown operation
func (s *BulkOperationWorkflowSuite) TestBulkOperation_UnknownOperation() {
	input := workflow.BulkOperationInput{
		TenantIDs: []uuid.UUID{uuid.New()},
		Operation: "nuke",
	}

	s.env.ExecuteWorkflow(workflow.BulkOperationWorkflow, input)

	require.True(s.T(), s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "unknown operation")
}
