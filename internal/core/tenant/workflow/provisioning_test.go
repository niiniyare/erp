package workflow_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"awo/internal/core/tenant/domain"
	"awo/internal/core/tenant/workflow"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

// ---------------------------------------------------------------------------
// ProvisioningWorkflowSuite — uses Temporal test environment
// ---------------------------------------------------------------------------

type ProvisioningWorkflowSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func TestProvisioningWorkflowSuite(t *testing.T) {
	suite.Run(t, new(ProvisioningWorkflowSuite))
}

func (s *ProvisioningWorkflowSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivityWithOptions(
		func(ctx context.Context, input domain.ProvisioningInput) (*domain.ProvisioningResult, error) { return nil, nil },
		activity.RegisterOptions{Name: "ProvisionTenantActivity"},
	)
	s.env.RegisterActivityWithOptions(
		func(ctx context.Context, tenantID uuid.UUID) error { return nil },
		activity.RegisterOptions{Name: "CreateDefaultConfigActivity"},
	)
	s.env.RegisterActivityWithOptions(
		func(ctx context.Context, tenantID uuid.UUID) error { return nil },
		activity.RegisterOptions{Name: "InitUsageActivity"},
	)
	s.env.RegisterActivityWithOptions(
		func(ctx context.Context, tenantID uuid.UUID) error { return nil },
		activity.RegisterOptions{Name: "ActivateTenantActivity"},
	)
	s.env.RegisterActivityWithOptions(
		func(ctx context.Context, tenantID uuid.UUID) error { return nil },
		activity.RegisterOptions{Name: "SendWelcomeNotificationActivity"},
	)
	s.env.RegisterActivityWithOptions(
		func(ctx context.Context, tenantID uuid.UUID) error { return nil },
		activity.RegisterOptions{Name: "CleanupTenantActivity"},
	)
}

func (s *ProvisioningWorkflowSuite) AfterTest(_, _ string) {
	s.env.AssertExpectations(s.T())
}

// TN-WF-001: Successful provisioning workflow
func (s *ProvisioningWorkflowSuite) TestProvisioningWorkflow_Success() {
	tenantID := uuid.New()
	input := domain.ProvisioningInput{
		Name:         "Acme Corp",
		Email:        "admin@acme.com",
		CountryCode:  "US",
		CurrencyCode: "USD",
	}
	provResult := domain.ProvisioningResult{
		TenantID: tenantID,
		Slug:     "acme-corp",
		Status:   "pending",
	}

	// Register activity mocks in order
	s.env.OnActivity("ProvisionTenantActivity", mock.Anything, input).Return(&provResult, nil)
	s.env.OnActivity("CreateDefaultConfigActivity", mock.Anything, tenantID).Return(nil)
	s.env.OnActivity("InitUsageActivity", mock.Anything, tenantID).Return(nil)
	s.env.OnActivity("ActivateTenantActivity", mock.Anything, tenantID).Return(nil)
	s.env.OnActivity("SendWelcomeNotificationActivity", mock.Anything, tenantID).Return(nil)

	s.env.ExecuteWorkflow(workflow.ProvisioningWorkflow, input)

	require.True(s.T(), s.env.IsWorkflowCompleted())
	require.NoError(s.T(), s.env.GetWorkflowError())

	var result domain.ProvisioningResult
	require.NoError(s.T(), s.env.GetWorkflowResult(&result))
	require.Equal(s.T(), tenantID, result.TenantID)
}

// TN-WF-002: Rollback on config creation failure
func (s *ProvisioningWorkflowSuite) TestProvisioningWorkflow_RollbackOnConfigFailure() {
	tenantID := uuid.New()
	input := domain.ProvisioningInput{
		Name:         "Fail Corp",
		Email:        "fail@test.com",
		CountryCode:  "US",
		CurrencyCode: "USD",
	}
	provResult := domain.ProvisioningResult{
		TenantID: tenantID,
		Slug:     "fail-corp",
		Status:   "pending",
	}

	s.env.OnActivity("ProvisionTenantActivity", mock.Anything, input).Return(&provResult, nil)
	s.env.OnActivity("CreateDefaultConfigActivity", mock.Anything, tenantID).Return(fmt.Errorf("config creation failed"))
	// Compensation: cleanup should be called
	s.env.OnActivity("CleanupTenantActivity", mock.Anything, tenantID).Return(nil)

	s.env.ExecuteWorkflow(workflow.ProvisioningWorkflow, input)

	require.True(s.T(), s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "create config failed")
}

// TN-WF-002 variant: Rollback on init usage failure
func (s *ProvisioningWorkflowSuite) TestProvisioningWorkflow_RollbackOnInitUsageFailure() {
	tenantID := uuid.New()
	input := domain.ProvisioningInput{
		Name:         "Usage Fail",
		Email:        "usage@test.com",
		CountryCode:  "US",
		CurrencyCode: "USD",
	}
	provResult := domain.ProvisioningResult{
		TenantID: tenantID,
		Slug:     "usage-fail",
		Status:   "pending",
	}

	s.env.OnActivity("ProvisionTenantActivity", mock.Anything, input).Return(&provResult, nil)
	s.env.OnActivity("CreateDefaultConfigActivity", mock.Anything, tenantID).Return(nil)
	s.env.OnActivity("InitUsageActivity", mock.Anything, tenantID).Return(fmt.Errorf("init usage failed"))
	s.env.OnActivity("CleanupTenantActivity", mock.Anything, tenantID).Return(nil)

	s.env.ExecuteWorkflow(workflow.ProvisioningWorkflow, input)

	require.True(s.T(), s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "init usage failed")
}

// TN-WF-002 variant: Rollback on activation failure
func (s *ProvisioningWorkflowSuite) TestProvisioningWorkflow_RollbackOnActivationFailure() {
	tenantID := uuid.New()
	input := domain.ProvisioningInput{
		Name:         "Activate Fail",
		Email:        "act@test.com",
		CountryCode:  "US",
		CurrencyCode: "USD",
	}
	provResult := domain.ProvisioningResult{
		TenantID: tenantID,
		Slug:     "activate-fail",
		Status:   "pending",
	}

	s.env.OnActivity("ProvisionTenantActivity", mock.Anything, input).Return(&provResult, nil)
	s.env.OnActivity("CreateDefaultConfigActivity", mock.Anything, tenantID).Return(nil)
	s.env.OnActivity("InitUsageActivity", mock.Anything, tenantID).Return(nil)
	s.env.OnActivity("ActivateTenantActivity", mock.Anything, tenantID).Return(fmt.Errorf("activation failed"))
	s.env.OnActivity("CleanupTenantActivity", mock.Anything, tenantID).Return(nil)

	s.env.ExecuteWorkflow(workflow.ProvisioningWorkflow, input)

	require.True(s.T(), s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "activate tenant failed")
}

// TN-WF-001 variant: Welcome notification failure is non-fatal
func (s *ProvisioningWorkflowSuite) TestProvisioningWorkflow_WelcomeFailureNonFatal() {
	tenantID := uuid.New()
	input := domain.ProvisioningInput{
		Name:         "Welcome Fail",
		Email:        "welcome@test.com",
		CountryCode:  "US",
		CurrencyCode: "USD",
	}
	provResult := domain.ProvisioningResult{
		TenantID: tenantID,
		Slug:     "welcome-fail",
		Status:   "pending",
	}

	s.env.OnActivity("ProvisionTenantActivity", mock.Anything, input).Return(&provResult, nil)
	s.env.OnActivity("CreateDefaultConfigActivity", mock.Anything, tenantID).Return(nil)
	s.env.OnActivity("InitUsageActivity", mock.Anything, tenantID).Return(nil)
	s.env.OnActivity("ActivateTenantActivity", mock.Anything, tenantID).Return(nil)
	s.env.OnActivity("SendWelcomeNotificationActivity", mock.Anything, tenantID).Return(fmt.Errorf("email service down"))

	s.env.ExecuteWorkflow(workflow.ProvisioningWorkflow, input)

	// Workflow should still succeed — welcome is best-effort
	require.True(s.T(), s.env.IsWorkflowCompleted())
	require.NoError(s.T(), s.env.GetWorkflowError())

	var result domain.ProvisioningResult
	require.NoError(s.T(), s.env.GetWorkflowResult(&result))
	require.Equal(s.T(), tenantID, result.TenantID)
}

// TN-WF-001 variant: Provision step failure (no compensation needed)
func (s *ProvisioningWorkflowSuite) TestProvisioningWorkflow_ProvisionFailure() {
	input := domain.ProvisioningInput{
		Name:         "DB Fail",
		Email:        "db@test.com",
		CountryCode:  "US",
		CurrencyCode: "USD",
	}

	s.env.OnActivity("ProvisionTenantActivity", mock.Anything, input).Return(nil, fmt.Errorf("db connection refused"))

	s.env.ExecuteWorkflow(workflow.ProvisioningWorkflow, input)

	require.True(s.T(), s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "provision tenant failed")
}
