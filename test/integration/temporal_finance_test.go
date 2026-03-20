package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"

	"awo/internal/core/finance"
	"awo/internal/core/finance/domain"
	"awo/internal/core/finance/service"
	"awo/internal/core/finance/workflows"
	"awo/internal/shared/logger"
)

// IntegrationTestSuite provides utilities for integration testing of Temporal workflows
type IntegrationTestSuite struct {
	testsuite.WorkflowTestSuite
	logger logger.Logger
}

func (s *IntegrationTestSuite) SetupTest() {
	config := logger.DefaultConfig()
	config.Level = logger.InfoLevel
	factory := &logger.LoggerFactory{}
	log, _ := factory.NewLogger(config)
	s.logger = log
}

// TestFinanceWorkflowExecution tests the complete finance workflow execution pipeline
func TestFinanceWorkflowExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := &IntegrationTestSuite{}
	suite.SetupTest()

	t.Run("TransactionApprovalWorkflow", func(t *testing.T) {
		suite.testTransactionApprovalWorkflow(t)
	})

	t.Run("TransactionProcessingWorkflow", func(t *testing.T) {
		suite.testTransactionProcessingWorkflow(t)
	})

	t.Run("AccountCreationWorkflow", func(t *testing.T) {
		suite.testAccountCreationWorkflow(t)
	})
}

func (s *IntegrationTestSuite) testTransactionApprovalWorkflow(t *testing.T) {
	// Create test environment
	env := s.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{
		EnableLoggingInReplay: true,
	})

	// Register workflows and activities
	transactionWorkflows := &workflows.TransactionWorkflows{}
	env.RegisterWorkflow(transactionWorkflows.TransactionApprovalWorkflow)

	// Mock activities for this test
	s.mockApprovalActivities(env)

	// Create test data
	transactionID := uuid.New()
	testInput := domain.TransactionApprovalWorkflowInput{
		TransactionID: transactionID,
		Amount:        decimal.NewFromFloat(5000.00),
		AccountType:   "EXPENSE",
		SubmittedBy:   "john.doe@company.com",
		SubmittedAt:   time.Now(),
	}

	// Execute workflow
	env.ExecuteWorkflow(transactionWorkflows.TransactionApprovalWorkflow, testInput)

	// Verify completion
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// Get result
	var result domain.TransactionApprovalWorkflowResult
	err := env.GetWorkflowResult(&result)
	require.NoError(t, err)

	// Verify workflow executed correctly
	assert.Equal(t, transactionID, result.TransactionID)
	assert.NotEmpty(t, result.Status)
	assert.NotZero(t, result.StartTime)
	assert.Positive(t, result.Duration)

	s.logger.Info("Transaction approval workflow test passed", logger.Fields{
		"transaction_id": result.TransactionID,
		"status":         result.Status,
		"duration":       result.Duration,
	})
}

func (s *IntegrationTestSuite) testTransactionProcessingWorkflow(t *testing.T) {
	env := s.NewTestWorkflowEnvironment()

	transactionWorkflows := &workflows.TransactionWorkflows{}
	env.RegisterWorkflow(transactionWorkflows.TransactionProcessingWorkflow)

	// Mock activities
	s.mockProcessingActivities(env)

	transactionID := uuid.New()
	testInput := domain.TransactionProcessingWorkflowInput{
		TransactionID: transactionID,
	}

	env.ExecuteWorkflow(transactionWorkflows.TransactionProcessingWorkflow, testInput)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result domain.TransactionProcessingWorkflowResult
	err := env.GetWorkflowResult(&result)
	require.NoError(t, err)

	assert.Equal(t, transactionID, result.TransactionID)
	assert.Equal(t, domain.ProcessingStatusCompleted, result.Status)
	assert.NotEmpty(t, result.PostingReference)

	s.logger.Info("Transaction processing workflow test passed", logger.Fields{
		"transaction_id":    result.TransactionID,
		"status":            result.Status,
		"posting_reference": result.PostingReference,
	})
}

func (s *IntegrationTestSuite) testAccountCreationWorkflow(t *testing.T) {
	env := s.NewTestWorkflowEnvironment()

	accountWorkflows := &workflows.AccountWorkflows{}
	env.RegisterWorkflow(accountWorkflows.AccountCreationWorkflow)

	// Mock activities
	s.mockAccountActivities(env)

	testInput := domain.AccountCreationWorkflowInput{
		AccountCode: "1000-001",
		AccountName: "Test Cash Account",
		AccountType: "ASSET",
		CreatedBy:   "admin@company.com",
		Recipients:  []string{"accounting@company.com"},
	}

	env.ExecuteWorkflow(accountWorkflows.AccountCreationWorkflow, testInput)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result domain.AccountCreationWorkflowResult
	err := env.GetWorkflowResult(&result)
	require.NoError(t, err)

	assert.NotNil(t, result.AccountID)
	assert.Equal(t, domain.AccountCreationStatusCompleted, result.Status)

	s.logger.Info("Account creation workflow test passed", logger.Fields{
		"account_id": result.AccountID,
		"status":     result.Status,
	})
}

// Mock activity implementations

func (s *IntegrationTestSuite) mockApprovalActivities(env *testsuite.TestWorkflowEnvironment) {
	// Mock validation activity
	env.OnActivity("ValidateTransactionRequest", domain.TransactionValidationInput{}).Return(
		&domain.TransactionValidationResult{
			IsValid:          true,
			ValidationErrors: []string{},
		}, nil).Times(0) // Allow any number of calls with any arguments

	// Mock approval requirements activity
	env.OnActivity("DetermineApprovalRequirements", domain.ApprovalRequirementsInput{}).Return(
		&domain.ApprovalRequirementsResult{
			RequiredApprovers:     []string{"manager@company.com"},
			RequiredApprovalCount: 1,
			ApprovalTimeout:       time.Hour,
		}, nil).Times(0)

	// Mock notification activity
	env.OnActivity("SendApprovalNotification", domain.ApprovalNotificationInput{}).Return(
		nil, nil).Times(0)
}

func (s *IntegrationTestSuite) mockProcessingActivities(env *testsuite.TestWorkflowEnvironment) {
	// Mock double entry validation
	env.OnActivity("ValidateDoubleEntry", domain.DoubleEntryValidationInput{}).Return(
		&domain.DoubleEntryValidationResult{
			IsValid: true,
		}, nil).Times(0)

	// Mock ledger posting
	env.OnActivity("PostTransactionToLedger", domain.LedgerPostingInput{}).Return(
		&domain.LedgerPostingResult{
			PostingReference: "POST-12345",
			ProcessedEntries: []domain.TransactionEntry{
				{
					ID:          uuid.New(),
					AccountID:   uuid.New(),
					DebitAmount: decimal.NewFromFloat(1000),
					Description: "Test entry",
				},
			},
		}, nil).Times(0)

	// Mock balance update
	env.OnActivity("UpdateAccountBalance", domain.BalanceUpdateInput{}).Return(
		nil, nil).Times(0)
}

func (s *IntegrationTestSuite) mockAccountActivities(env *testsuite.TestWorkflowEnvironment) {
	// Mock account validation
	env.OnActivity("ValidateAccountCreationActivity", domain.AccountHierarchyValidationInput{}).Return(
		&domain.AccountHierarchyValidationResult{
			IsValid: true,
		}, nil).Times(0)

	// Mock account creation
	env.OnActivity("CreateAccountActivity", domain.AccountCreationInput{}).Return(
		&domain.AccountCreationResult{
			AccountID: uuid.New(),
		}, nil).Times(0)

	// Mock notification
	env.OnActivity("SendAccountCreatedNotificationActivity", domain.AccountNotificationInput{}).Return(
		nil, nil).Times(0)
}

// TestTemporalIntegrationSetup tests that the temporal integration can be properly set up
func TestTemporalIntegrationSetup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test verifies the temporal integration setup process
	loggerConfig := logger.DefaultConfig()
	loggerConfig.Level = logger.InfoLevel
	factory := &logger.LoggerFactory{}
	testLogger, _ := factory.NewLogger(loggerConfig)

	// Create a minimal configuration that should pass validation
	config := finance.TemporalIntegrationConfig{
		Services: &service.Services{
			Account:          nil, // Would be initialized with proper mock
			Transaction:      nil, // Would be initialized with proper mock
			TransactionEntry: nil, // Would be initialized with proper mock
		},
		IAMService:          nil, // Optional for this test
		AuditService:        nil, // Optional for this test
		FeatureFlagService:  nil, // Optional for this test
		SettingsService:     nil, // Optional for this test
		NotificationService: nil, // Optional for this test
		CacheService:        nil, // Optional for this test
		TemporalClient:      nil, // Would provide test client in full integration
		Logger:              testLogger,
		Metrics:             nil, // Optional for this test
		Tracer:              nil, // Optional for this test
	}

	// Test that the integration can be created (even if it fails due to missing client)
	_, err := finance.NewTemporalIntegration(config)

	// We expect this to fail due to missing temporal client, but not due to structural issues
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "temporal client is required")

	t.Logf("✅ Temporal integration setup test completed - structure validation passed")
}

// BenchmarkWorkflowExecution benchmarks workflow execution performance
func BenchmarkWorkflowExecution(b *testing.B) {
	suite := &IntegrationTestSuite{}
	suite.SetupTest()

	b.Run("TransactionApproval", func(b *testing.B) {
		env := suite.NewTestWorkflowEnvironment()
		transactionWorkflows := &workflows.TransactionWorkflows{}
		env.RegisterWorkflow(transactionWorkflows.TransactionApprovalWorkflow)
		suite.mockApprovalActivities(env)

		testInput := domain.TransactionApprovalWorkflowInput{
			TransactionID: uuid.New(),
			Amount:        decimal.NewFromFloat(1000.00),
			AccountType:   "EXPENSE",
			SubmittedBy:   "user@company.com",
			SubmittedAt:   time.Now(),
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			testInput.TransactionID = uuid.New() // Unique ID for each iteration
			env.ExecuteWorkflow(transactionWorkflows.TransactionApprovalWorkflow, testInput)
		}
	})
}
