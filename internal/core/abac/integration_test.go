package abac

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// ABACIntegrationTestSuite provides a comprehensive integration test suite
type ABACIntegrationTestSuite struct {
	suite.Suite

	// Test infrastructure
	ctx     context.Context
	db      *sql.DB
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService

	// ABAC components
	policyManager      PolicyManager
	evaluationEngine   PolicyEvaluationEngine
	attributeService   AttributeService
	attributeCollector AttributeCollector
	attributeResolver  AttributeResolver
	externalSources    ExternalAttributeSourceManager
	performanceOpt     PerformanceOptimizer
	monitoringService  MonitoringService
	securityCompliance SecurityComplianceManager

	// Test data
	testTenantID   uuid.UUID
	testUserID     uuid.UUID
	testResourceID uuid.UUID
	testPolicyID   uuid.UUID
}

// SetupSuite initializes the test environment
func (suite *ABACIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Initialize test infrastructure
	suite.logger = logger.NewMockLogger()
	suite.metrics = metrics.NewMockMetricsProvider()
	suite.tracer = tracing.NewMockTracingService()

	// Initialize ABAC components with mock implementations
	suite.setupABACComponents()

	// Create test data
	suite.createTestData()
}

func (suite *ABACIntegrationTestSuite) setupABACComponents() {
	// Mock repositories for integration testing
	policyRepo := &MockPolicyRepository{}
	attributeRepo := &MockAttributeRepository{}
	auditRepo := &MockAuditLogRepository{}

	// Initialize ABAC services
	suite.attributeService = NewAttributeService(
		attributeRepo,
		[]byte("test-key-for-encryption-32-bytes!"),
		suite.logger,
		suite.metrics,
		suite.tracer,
	)

	suite.attributeCollector = NewAttributeCollector(
		attributeRepo,
		nil, // sourceRepo
		suite.logger,
		suite.metrics,
		suite.tracer,
	)

	suite.attributeResolver = NewAttributeResolver(
		attributeRepo,
		suite.attributeCollector,
		suite.logger,
		suite.metrics,
		suite.tracer,
	)

	suite.externalSources = NewExternalAttributeSourceManager(
		nil, // sourceRepo
		suite.logger,
		suite.metrics,
		suite.tracer,
	)

	suite.evaluationEngine = NewPolicyEvaluationEngine(
		policyRepo,
		suite.attributeResolver,
		suite.externalSources,
		suite.logger,
		suite.metrics,
		suite.tracer,
	)

	suite.policyManager = NewPolicyManager(
		policyRepo,
		suite.evaluationEngine,
		suite.logger,
		suite.metrics,
		suite.tracer,
	)

	suite.performanceOpt = NewPerformanceOptimizer(
		policyRepo,
		nil, // evaluationRepo
		attributeRepo,
		suite.logger,
		suite.metrics,
		suite.tracer,
	)

	suite.monitoringService = NewMonitoringService(
		policyRepo,
		nil, // evaluationRepo
		attributeRepo,
		suite.logger,
		suite.metrics,
		suite.tracer,
	)

	suite.securityCompliance = NewSecurityComplianceManager(
		auditRepo,
		[]byte("test-key-for-encryption-32-bytes!"),
		suite.logger,
		suite.metrics,
		suite.tracer,
	)
}

func (suite *ABACIntegrationTestSuite) createTestData() {
	suite.testTenantID = uuid.New()
	suite.testUserID = uuid.New()
	suite.testResourceID = uuid.New()
	suite.testPolicyID = uuid.New()
}

// TestEndToEndPolicyEvaluation tests the complete policy evaluation flow
func (suite *ABACIntegrationTestSuite) TestEndToEndPolicyEvaluation() {
	// Step 1: Create attribute definitions
	suite.T().Log("Step 1: Creating attribute definitions")

	userDeptAttr, err := suite.attributeService.CreateAttributeDefinition(suite.ctx, &CreateAttributeDefinitionRequest{
		Name:          "user_department",
		DisplayName:   stringPtr("User Department"),
		Description:   stringPtr("The department the user belongs to"),
		DataType:      types.AttributeDataTypeString,
		Category:      types.AttributeCategoryUser,
		IsRequired:    true,
		AllowedValues: []interface{}{"engineering", "sales", "marketing", "hr"},
		SecuritySettings: AttributeSecuritySettings{
			IsEncrypted: false,
			IsSensitive: false,
		},
		CreatedBy: &suite.testUserID,
	})
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), userDeptAttr)

	// Step 2: Create a policy
	suite.T().Log("Step 2: Creating ABAC policy")

	policy, err := suite.policyManager.CreatePolicy(suite.ctx, &CreatePolicyRequest{
		Name:        "Engineering Document Access",
		DisplayName: stringPtr("Engineering Document Access Policy"),
		Description: stringPtr("Allows engineering users to read engineering documents"),
		PolicyType:  types.PolicyTypeABAC,
		Effect:      types.PolicyEffectPermit,
		Priority:    100,
		Category:    stringPtr("document_access"),
		Target: &models.PolicyTarget{
			Subjects: []models.PolicyTargetMatch{
				{
					AttributeID: userDeptAttr.ID.String(),
					Match:       models.PolicyMatchExact,
					Values:      []interface{}{"engineering"},
				},
			},
			Resources: []models.PolicyTargetMatch{
				{
					AttributeID: "resource.type",
					Match:       models.PolicyMatchExact,
					Values:      []interface{}{"document"},
				},
			},
			Actions: []models.PolicyTargetMatch{
				{
					AttributeID: "action.name",
					Match:       models.PolicyMatchExact,
					Values:      []interface{}{"read"},
				},
			},
		},
		Rules: []*models.PolicyRule{
			{
				ID:       uuid.New(),
				Name:     "Engineering Access Rule",
				Effect:   types.PolicyEffectPermit,
				Priority: 1,
				Condition: &models.PolicyCondition{
					Expression: "subject.department == 'engineering' AND resource.type == 'document' AND action.name == 'read'",
				},
			},
		},
		CombiningAlgorithm: types.CombiningAlgorithmDenyOverrides,
		Status:             types.PolicyStatusActive,
		CreatedBy:          &suite.testUserID,
	})
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), policy)
	suite.testPolicyID = policy.ID

	// Step 3: Test policy evaluation with matching attributes
	suite.T().Log("Step 3: Testing policy evaluation with matching attributes")

	evaluationRequest := &PolicyEvaluationRequest{
		RequestID: uuid.New(),
		PolicyID:  suite.testPolicyID,
		Subject: SubjectContext{
			UserID: suite.testUserID,
			Roles:  []string{"engineer"},
			Attributes: map[string]interface{}{
				"department": "engineering",
				"clearance":  "standard",
			},
		},
		Resource: ResourceContext{
			ResourceID:   suite.testResourceID,
			ResourceType: "document",
			Attributes: map[string]interface{}{
				"classification": "internal",
				"department":     "engineering",
			},
		},
		Action: ActionContext{
			Action: "read",
		},
		Environment: EnvironmentContext{
			Timestamp: time.Now(),
		},
		EvaluationMode: EvaluationModeStandard,
		Options: PolicyEvaluationOptions{
			IncludeExplanation:  true,
			TrackAttributeUsage: true,
			EnableCaching:       true,
		},
	}

	result, err := suite.evaluationEngine.EvaluatePolicy(suite.ctx, evaluationRequest)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), types.PolicyDecisionPermit, result.Decision)
	assert.NotNil(suite.T(), result.Explanation)
	assert.NotEmpty(suite.T(), result.ApplicableRules)

	// Step 4: Test policy evaluation with non-matching attributes
	suite.T().Log("Step 4: Testing policy evaluation with non-matching attributes")

	evaluationRequest.Subject.Attributes["department"] = "sales"

	result, err = suite.evaluationEngine.EvaluatePolicy(suite.ctx, evaluationRequest)
	require.NoError(t, err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), types.PolicyDecisionDeny, result.Decision)

	// Step 5: Verify audit trail
	suite.T().Log("Step 5: Verifying audit trail")

	auditQuery := &AuditQueryRequest{
		ActorID:      &suite.testUserID,
		ResourceType: "document",
		StartTime:    &time.Time{},
		EndTime:      timePtr(time.Now()),
		Page:         1,
		PageSize:     10,
	}

	auditResult, err := suite.securityCompliance.QueryAuditLog(suite.ctx, auditQuery)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), auditResult)
}

// TestAttributeManagementFlow tests the complete attribute management workflow
func (suite *ABACIntegrationTestSuite) TestAttributeManagementFlow() {
	// Step 1: Create sensitive attribute definition
	suite.T().Log("Step 1: Creating sensitive attribute definition")

	ssnAttr, err := suite.attributeService.CreateAttributeDefinition(suite.ctx, &CreateAttributeDefinitionRequest{
		Name:        "social_security_number",
		DisplayName: stringPtr("Social Security Number"),
		Description: stringPtr("Employee's social security number"),
		DataType:    types.AttributeDataTypeString,
		Category:    types.AttributeCategoryUser,
		IsRequired:  false,
		ValidationRules: []AttributeValidationRule{
			{
				Type: "regex",
				Parameters: map[string]interface{}{
					"pattern": `^\d{3}-\d{2}-\d{4}$`,
				},
			},
		},
		SecuritySettings: AttributeSecuritySettings{
			IsEncrypted: true,
			IsSensitive: true,
			AccessLevel: "restricted",
		},
		CreatedBy: &suite.testUserID,
	})
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ssnAttr)

	// Step 2: Test attribute validation
	suite.T().Log("Step 2: Testing attribute validation")

	// Valid SSN
	validationResult, err := suite.attributeService.ValidateAttributeValue(suite.ctx, &ValidateAttributeValueRequest{
		AttributeName: "social_security_number",
		Value:         "123-45-6789",
		DataType:      types.AttributeDataTypeString,
		ValidationRules: []AttributeValidationRule{
			{
				Type: "regex",
				Parameters: map[string]interface{}{
					"pattern": `^\d{3}-\d{2}-\d{4}$`,
				},
			},
		},
	})
	require.NoError(suite.T(), err)
	assert.True(suite.T(), validationResult.IsValid)
	assert.Empty(suite.T(), validationResult.ValidationErrors)

	// Invalid SSN
	validationResult, err = suite.attributeService.ValidateAttributeValue(suite.ctx, &ValidateAttributeValueRequest{
		AttributeName: "social_security_number",
		Value:         "invalid-ssn",
		DataType:      types.AttributeDataTypeString,
		ValidationRules: []AttributeValidationRule{
			{
				Type: "regex",
				Parameters: map[string]interface{}{
					"pattern": `^\d{3}-\d{2}-\d{4}$`,
				},
			},
		},
	})
	require.NoError(suite.T(), err)
	assert.False(suite.T(), validationResult.IsValid)
	assert.NotEmpty(suite.T(), validationResult.ValidationErrors)

	// Step 3: Test attribute encryption
	suite.T().Log("Step 3: Testing attribute encryption")

	encryptResult, err := suite.attributeService.EncryptAttributeValue(suite.ctx, &EncryptAttributeValueRequest{
		AttributeName: "social_security_number",
		Value:         "123-45-6789",
		EncryptionKey: "user-encryption-key",
	})
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), encryptResult)
	assert.NotEmpty(suite.T(), encryptResult.EncryptedValue)
	assert.NotEqual(suite.T(), "123-45-6789", encryptResult.EncryptedValue)

	// Step 4: Test attribute decryption
	suite.T().Log("Step 4: Testing attribute decryption")

	decryptResult, err := suite.attributeService.DecryptAttributeValue(suite.ctx, &DecryptAttributeValueRequest{
		AttributeName:  "social_security_number",
		EncryptedValue: encryptResult.EncryptedValue,
		EncryptionKey:  "user-encryption-key",
	})
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), decryptResult)
	assert.Equal(suite.T(), "123-45-6789", decryptResult.DecryptedValue)

	// Step 5: Test attribute collection
	suite.T().Log("Step 5: Testing attribute collection")

	collectionRequest := &AttributeCollectionRequest{
		RequestID:          uuid.New(),
		TargetType:         AttributeTargetTypeUser,
		TargetID:           suite.testUserID,
		RequiredAttributes: []string{"social_security_number", "department"},
		CollectionOptions: AttributeCollectionOptions{
			IncludeEncrypted: true,
			ValidateValues:   true,
		},
		Priority: CollectionPriorityHigh,
		Timeout:  30 * time.Second,
	}

	collectionResult, err := suite.attributeCollector.CollectAttributes(suite.ctx, collectionRequest)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), collectionResult)
}

// TestSecurityComplianceFlow tests the security and compliance workflow
func (suite *ABACIntegrationTestSuite) TestSecurityComplianceFlow() {
	// Step 1: Test data encryption
	suite.T().Log("Step 1: Testing data encryption")

	encryptRequest := &EncryptDataRequest{
		DataID:        &suite.testUserID,
		Data:          "sensitive personal information",
		DataType:      "personal_data",
		SecurityLevel: SecurityLevelHigh,
		ActorID:       &suite.testUserID,
	}

	encryptResult, err := suite.securityCompliance.EncryptSensitiveData(suite.ctx, encryptRequest)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), encryptResult)
	assert.True(suite.T(), encryptResult.IsEncrypted)
	assert.Equal(suite.T(), DataClassificationPersonal, encryptResult.Classification)

	// Step 2: Test data decryption
	suite.T().Log("Step 2: Testing data decryption")

	decryptRequest := &DecryptDataRequest{
		DataID:        &suite.testUserID,
		EncryptedData: encryptResult.EncryptedData,
		Purpose:       "authorized access for compliance review",
		ActorID:       &suite.testUserID,
	}

	decryptResult, err := suite.securityCompliance.DecryptSensitiveData(suite.ctx, decryptRequest)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), decryptResult)
	assert.Equal(suite.T(), "sensitive personal information", decryptResult.DecryptedData)
	assert.True(suite.T(), decryptResult.AccessLogged)

	// Step 3: Test audit event recording
	suite.T().Log("Step 3: Testing audit event recording")

	auditRequest := &AuditEventRequest{
		EventType:    AuditEventTypePolicyEvaluation,
		ActorID:      &suite.testUserID,
		ResourceType: "document",
		ResourceID:   &suite.testResourceID,
		Action:       "read",
		Result:       stringPtr("PERMIT"),
		Details: map[string]interface{}{
			"policy_id": suite.testPolicyID.String(),
			"decision":  "PERMIT",
		},
		Timestamp: time.Now(),
		IPAddress: stringPtr("192.168.1.100"),
		UserAgent: stringPtr("Test User Agent"),
	}

	err = suite.securityCompliance.RecordAuditEvent(suite.ctx, auditRequest)
	require.NoError(suite.T(), err)

	// Step 4: Test compliance report generation
	suite.T().Log("Step 4: Testing compliance report generation")

	reportRequest := &ComplianceReportRequest{
		Framework:   ComplianceFrameworkGDPR,
		PeriodStart: time.Now().Add(-30 * 24 * time.Hour),
		PeriodEnd:   time.Now(),
		RequestedBy: &suite.testUserID,
		Scope:       []string{"data_protection", "user_rights"},
	}

	report, err := suite.securityCompliance.GenerateComplianceReport(suite.ctx, reportRequest)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), report)
	assert.Equal(suite.T(), ComplianceFrameworkGDPR, report.Framework)
	assert.Equal(suite.T(), ComplianceStatusCompliant, report.Status)
	assert.Greater(suite.T(), report.Score, 0.0)

	// Step 5: Test data subject request processing
	suite.T().Log("Step 5: Testing data subject request processing")

	dataSubjectRequest := &DataSubjectRequest{
		SubjectID:   suite.testUserID,
		RequestType: DataSubjectRequestTypeAccess,
	}

	dataSubjectResponse, err := suite.securityCompliance.ProcessDataSubjectRequest(suite.ctx, dataSubjectRequest)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), dataSubjectResponse)
	assert.Equal(suite.T(), "completed", dataSubjectResponse.Status)
}

// TestPerformanceOptimizationFlow tests the performance optimization features
func (suite *ABACIntegrationTestSuite) TestPerformanceOptimizationFlow() {
	// Step 1: Test policy compilation
	suite.T().Log("Step 1: Testing policy compilation")

	compilationRequest := &PolicyCompilationRequest{
		PolicyID: suite.testPolicyID,
		Options: map[string]interface{}{
			"optimization_level": "high",
			"cache_rules":        true,
		},
	}

	compiledPolicy, err := suite.performanceOpt.CompilePolicy(suite.ctx, compilationRequest)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), compiledPolicy)

	// Step 2: Test evaluation caching
	suite.T().Log("Step 2: Testing evaluation caching")

	cacheRequest := &CacheRequest{
		Key: "policy_evaluation_cache_test",
		Parameters: map[string]interface{}{
			"policy_id": suite.testPolicyID.String(),
			"user_id":   suite.testUserID.String(),
			"action":    "read",
		},
	}

	// Should return cache miss first time
	cachedResult, err := suite.performanceOpt.GetCachedEvaluation(suite.ctx, cacheRequest)
	if err != nil {
		// Cache miss is expected for first access
		assert.Contains(suite.T(), err.Error(), "cache miss")
	}

	// Step 3: Test batch evaluation
	suite.T().Log("Step 3: Testing batch evaluation")

	batchRequest := &BatchEvaluationRequest{
		RequestID: uuid.New(),
		Evaluations: []PolicyEvaluationRequest{
			{
				PolicyID: suite.testPolicyID,
				Subject: SubjectContext{
					UserID: suite.testUserID,
					Attributes: map[string]interface{}{
						"department": "engineering",
					},
				},
				Resource: ResourceContext{
					ResourceID:   suite.testResourceID,
					ResourceType: "document",
				},
				Action: ActionContext{
					Action: "read",
				},
			},
		},
		Options: BatchEvaluationOptions{
			Parallel:    true,
			MaxWorkers:  4,
			EnableCache: true,
		},
	}

	batchResult, err := suite.performanceOpt.BatchEvaluate(suite.ctx, batchRequest)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), batchResult)
	assert.Len(suite.T(), batchResult.Results, 1)
}

// TestMonitoringAndAnalyticsFlow tests the monitoring and analytics features
func (suite *ABACIntegrationTestSuite) TestMonitoringAndAnalyticsFlow() {
	// Step 1: Record evaluation metrics
	suite.T().Log("Step 1: Recording evaluation metrics")

	metricsRequest := &EvaluationMetricsRequest{
		EvaluationID:   uuid.New(),
		UserID:         suite.testUserID,
		PolicyID:       &suite.testPolicyID,
		ResourceType:   "document",
		ResourceID:     &suite.testResourceID,
		Action:         "read",
		Decision:       types.PolicyDecisionPermit,
		ExecutionTime:  25 * time.Millisecond,
		CacheHit:       false,
		AttributeCount: 5,
		Timestamp:      time.Now(),
	}

	err := suite.monitoringService.RecordEvaluationMetrics(suite.ctx, metricsRequest)
	require.NoError(suite.T(), err)

	// Step 2: Get evaluation metrics
	suite.T().Log("Step 2: Retrieving evaluation metrics")

	metricsQuery := &MetricsQueryRequest{
		StartTime: timePtr(time.Now().Add(-1 * time.Hour)),
		EndTime:   timePtr(time.Now()),
		PolicyID:  &suite.testPolicyID,
		UserID:    &suite.testUserID,
	}

	metrics, err := suite.monitoringService.GetEvaluationMetrics(suite.ctx, metricsQuery)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), metrics)

	// Step 3: Test anomaly detection
	suite.T().Log("Step 3: Testing anomaly detection")

	anomalyRequest := &AnomalyDetectionRequest{
		TimeWindow: 1 * time.Hour,
		Context: map[string]interface{}{
			"user_id": suite.testUserID.String(),
		},
	}

	anomalies, err := suite.monitoringService.DetectAnomalies(suite.ctx, anomalyRequest)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), anomalies)

	// Step 4: Get system health
	suite.T().Log("Step 4: Checking system health")

	health, err := suite.monitoringService.GetSystemHealth(suite.ctx)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), health)
}

// TestErrorHandlingAndRecovery tests error scenarios and recovery mechanisms
func (suite *ABACIntegrationTestSuite) TestErrorHandlingAndRecovery() {
	// Step 1: Test policy evaluation with missing policy
	suite.T().Log("Step 1: Testing evaluation with non-existent policy")

	nonExistentPolicyID := uuid.New()
	evaluationRequest := &PolicyEvaluationRequest{
		RequestID: uuid.New(),
		PolicyID:  nonExistentPolicyID,
		Subject: SubjectContext{
			UserID: suite.testUserID,
		},
		Resource: ResourceContext{
			ResourceID:   suite.testResourceID,
			ResourceType: "document",
		},
		Action: ActionContext{
			Action: "read",
		},
	}

	result, err := suite.evaluationEngine.EvaluatePolicy(suite.ctx, evaluationRequest)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to retrieve policy")

	// Step 2: Test attribute validation with invalid data
	suite.T().Log("Step 2: Testing attribute validation with invalid data")

	validationRequest := &ValidateAttributeValueRequest{
		AttributeName: "email",
		Value:         "invalid-email-format",
		DataType:      types.AttributeDataTypeString,
		ValidationRules: []AttributeValidationRule{
			{
				Type: "regex",
				Parameters: map[string]interface{}{
					"pattern": `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
				},
			},
		},
	}

	validationResult, err := suite.attributeService.ValidateAttributeValue(suite.ctx, validationRequest)
	require.NoError(suite.T(), err)
	assert.False(suite.T(), validationResult.IsValid)
	assert.NotEmpty(suite.T(), validationResult.ValidationErrors)

	// Step 3: Test encryption with invalid key
	suite.T().Log("Step 3: Testing decryption with wrong key")

	// First encrypt with one key
	encryptResult, err := suite.attributeService.EncryptAttributeValue(suite.ctx, &EncryptAttributeValueRequest{
		AttributeName: "test_field",
		Value:         "test value",
		EncryptionKey: "correct-key",
	})
	require.NoError(suite.T(), err)

	// Try to decrypt with different key (this should fail)
	_, err = suite.attributeService.DecryptAttributeValue(suite.ctx, &DecryptAttributeValueRequest{
		AttributeName:  "test_field",
		EncryptedValue: encryptResult.EncryptedValue,
		EncryptionKey:  "wrong-key",
	})
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "decryption failed")
}

// TestConcurrentOperations tests concurrent access patterns
func (suite *ABACIntegrationTestSuite) TestConcurrentOperations() {
	// Test concurrent policy evaluations
	suite.T().Log("Testing concurrent policy evaluations")

	const numConcurrentEvaluations = 10
	results := make(chan *PolicyEvaluationResult, numConcurrentEvaluations)
	errors := make(chan error, numConcurrentEvaluations)

	for i := 0; i < numConcurrentEvaluations; i++ {
		go func(index int) {
			evaluationRequest := &PolicyEvaluationRequest{
				RequestID: uuid.New(),
				PolicyID:  suite.testPolicyID,
				Subject: SubjectContext{
					UserID: uuid.New(), // Different user for each evaluation
					Attributes: map[string]interface{}{
						"department": "engineering",
					},
				},
				Resource: ResourceContext{
					ResourceID:   uuid.New(), // Different resource for each evaluation
					ResourceType: "document",
				},
				Action: ActionContext{
					Action: "read",
				},
				Environment: EnvironmentContext{
					Timestamp: time.Now(),
				},
			}

			result, err := suite.evaluationEngine.EvaluatePolicy(suite.ctx, evaluationRequest)
			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}(i)
	}

	// Collect results
	successCount := 0
	errorCount := 0

	for i := 0; i < numConcurrentEvaluations; i++ {
		select {
		case result := <-results:
			assert.NotNil(suite.T(), result)
			successCount++
		case err := <-errors:
			suite.T().Logf("Concurrent evaluation error: %v", err)
			errorCount++
		case <-time.After(5 * time.Second):
			suite.T().Fatal("Timeout waiting for concurrent evaluations")
		}
	}

	suite.T().Logf("Concurrent evaluations: %d successful, %d errors", successCount, errorCount)
	assert.Equal(suite.T(), numConcurrentEvaluations, successCount+errorCount)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// Run the test suite
func TestABACIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ABACIntegrationTestSuite))
}

// Additional integration tests for specific scenarios

func TestRealWorldScenarios(t *testing.T) {
	t.Run("HealthcareDataAccess", func(t *testing.T) {
		testHealthcareDataAccessScenario(t)
	})

	t.Run("FinancialRecordsAccess", func(t *testing.T) {
		testFinancialRecordsAccessScenario(t)
	})

	t.Run("MultiTenantAccess", func(t *testing.T) {
		testMultiTenantAccessScenario(t)
	})
}

func testHealthcareDataAccessScenario(t *testing.T) {
	// Test scenario: Doctor accessing patient medical records
	// Only doctors in the same department can access patient records
	// Emergency personnel can access any records during emergencies

	ctx := context.Background()
	logger := logger.NewMockLogger()
	metrics := metrics.NewMockMetricsProvider()
	tracer := tracing.NewMockTracingService()

	// Mock components
	policyRepo := &MockPolicyRepository{}
	attributeRepo := &MockAttributeRepository{}
	auditRepo := &MockAuditLogRepository{}

	evaluationEngine := NewPolicyEvaluationEngine(
		policyRepo,
		NewAttributeResolver(attributeRepo, nil, logger, metrics, tracer),
		NewExternalAttributeSourceManager(nil, logger, metrics, tracer),
		logger,
		metrics,
		tracer,
	)

	// Test data
	doctorID := uuid.New()
	patientID := uuid.New()
	medicalRecordID := uuid.New()
	policyID := uuid.New()

	// Create healthcare access policy
	healthcarePolicy := &models.Policy{
		ID:                 policyID,
		Name:               "Healthcare Data Access Policy",
		Status:             types.PolicyStatusActive,
		Effect:             types.PolicyEffectPermit,
		CombiningAlgorithm: types.CombiningAlgorithmDenyOverrides,
		Rules: []*models.PolicyRule{
			{
				ID:     uuid.New(),
				Name:   "Doctor Same Department Access",
				Effect: types.PolicyEffectPermit,
				Condition: &models.PolicyCondition{
					Expression: "subject.role == 'doctor' AND subject.department == resource.department",
				},
			},
			{
				ID:     uuid.New(),
				Name:   "Emergency Access",
				Effect: types.PolicyEffectPermit,
				Condition: &models.PolicyCondition{
					Expression: "subject.role == 'emergency_personnel' OR environment.emergency_status == 'active'",
				},
			},
		},
	}

	// Mock policy retrieval
	policyRepo.On("GetPolicyByID", mock.Anything, policyID).Return(healthcarePolicy, nil)

	// Test 1: Doctor accessing patient in same department
	t.Run("DoctorSameDepartmentAccess", func(t *testing.T) {
		request := &PolicyEvaluationRequest{
			RequestID: uuid.New(),
			PolicyID:  policyID,
			Subject: SubjectContext{
				UserID: doctorID,
				Roles:  []string{"doctor"},
				Attributes: map[string]interface{}{
					"role":       "doctor",
					"department": "cardiology",
					"license":    "MD12345",
				},
			},
			Resource: ResourceContext{
				ResourceID:   medicalRecordID,
				ResourceType: "medical_record",
				Owner:        &patientID,
				Attributes: map[string]interface{}{
					"patient_id":  patientID.String(),
					"department":  "cardiology",
					"sensitivity": "high",
				},
			},
			Action: ActionContext{
				Action: "read",
			},
			Environment: EnvironmentContext{
				Timestamp: time.Now(),
				Attributes: map[string]interface{}{
					"emergency_status": "normal",
				},
			},
		}

		result, err := evaluationEngine.EvaluatePolicy(ctx, request)
		require.NoError(t, err)
		assert.Equal(t, types.PolicyDecisionPermit, result.Decision)
	})

	// Test 2: Doctor accessing patient in different department (should be denied)
	t.Run("DoctorDifferentDepartmentAccess", func(t *testing.T) {
		request := &PolicyEvaluationRequest{
			RequestID: uuid.New(),
			PolicyID:  policyID,
			Subject: SubjectContext{
				UserID: doctorID,
				Roles:  []string{"doctor"},
				Attributes: map[string]interface{}{
					"role":       "doctor",
					"department": "cardiology",
				},
			},
			Resource: ResourceContext{
				ResourceID:   medicalRecordID,
				ResourceType: "medical_record",
				Attributes: map[string]interface{}{
					"department": "neurology", // Different department
				},
			},
			Action: ActionContext{
				Action: "read",
			},
			Environment: EnvironmentContext{
				Timestamp: time.Now(),
				Attributes: map[string]interface{}{
					"emergency_status": "normal",
				},
			},
		}

		result, err := evaluationEngine.EvaluatePolicy(ctx, request)
		require.NoError(t, err)
		assert.Equal(t, types.PolicyDecisionDeny, result.Decision)
	})

	// Test 3: Emergency access (should be permitted regardless of department)
	t.Run("EmergencyAccess", func(t *testing.T) {
		request := &PolicyEvaluationRequest{
			RequestID: uuid.New(),
			PolicyID:  policyID,
			Subject: SubjectContext{
				UserID: uuid.New(),
				Roles:  []string{"emergency_personnel"},
				Attributes: map[string]interface{}{
					"role":       "emergency_personnel",
					"department": "emergency",
				},
			},
			Resource: ResourceContext{
				ResourceID:   medicalRecordID,
				ResourceType: "medical_record",
				Attributes: map[string]interface{}{
					"department": "cardiology", // Different department
				},
			},
			Action: ActionContext{
				Action: "read",
			},
			Environment: EnvironmentContext{
				Timestamp: time.Now(),
				Attributes: map[string]interface{}{
					"emergency_status": "active",
				},
			},
		}

		result, err := evaluationEngine.EvaluatePolicy(ctx, request)
		require.NoError(t, err)
		assert.Equal(t, types.PolicyDecisionPermit, result.Decision)
	})
}

func testFinancialRecordsAccessScenario(t *testing.T) {
	// Test scenario: Financial services with SOX compliance
	// Only authorized financial analysts can access financial records
	// Auditors have read-only access to all records
	// Managers can access records of their direct reports

	t.Log("Testing financial records access scenario with SOX compliance")

	ctx := context.Background()
	logger := logger.NewMockLogger()
	metrics := metrics.NewMockMetricsProvider()
	tracer := tracing.NewMockTracingService()

	auditRepo := &MockAuditLogRepository{}
	securityCompliance := NewSecurityComplianceManager(
		auditRepo,
		[]byte("financial-encryption-key-32-bytes!"),
		logger,
		metrics,
		tracer,
	)

	// Test compliance validation for financial framework
	complianceRequest := &ComplianceValidationRequest{
		Framework: ComplianceFrameworkSOX,
		Context: map[string]interface{}{
			"data_type":     "financial_record",
			"access_method": "api",
			"user_role":     "financial_analyst",
			"audit_logging": true,
			"encryption":    true,
		},
	}

	result, err := securityCompliance.ValidateComplianceRequirements(ctx, complianceRequest)
	require.NoError(t, err)
	assert.True(t, result.IsCompliant)
	assert.Equal(t, ComplianceFrameworkSOX, result.Framework)
	assert.Greater(t, result.ComplianceScore, 80.0) // High compliance score expected

	t.Logf("SOX compliance validation completed with score: %.2f", result.ComplianceScore)
}

func testMultiTenantAccessScenario(t *testing.T) {
	// Test scenario: Multi-tenant SaaS application
	// Users can only access resources within their tenant
	// Tenant admins can access all resources within their tenant
	// System admins can access resources across tenants

	t.Log("Testing multi-tenant access scenario")

	ctx := context.Background()
	logger := logger.NewMockLogger()
	metrics := metrics.NewMockMetricsProvider()
	tracer := tracing.NewMockTracingService()

	// Test data for different tenants
	tenant1ID := uuid.New()
	tenant2ID := uuid.New()
	user1ID := uuid.New()
	user2ID := uuid.New()
	systemAdminID := uuid.New()
	resource1ID := uuid.New()
	resource2ID := uuid.New()

	policyRepo := &MockPolicyRepository{}
	attributeRepo := &MockAttributeRepository{}

	evaluationEngine := NewPolicyEvaluationEngine(
		policyRepo,
		NewAttributeResolver(attributeRepo, nil, logger, metrics, tracer),
		NewExternalAttributeSourceManager(nil, logger, metrics, tracer),
		logger,
		metrics,
		tracer,
	)

	// Multi-tenant policy
	multiTenantPolicy := &models.Policy{
		ID:                 uuid.New(),
		Name:               "Multi-Tenant Access Policy",
		Status:             types.PolicyStatusActive,
		Effect:             types.PolicyEffectPermit,
		CombiningAlgorithm: types.CombiningAlgorithmDenyOverrides,
		Rules: []*models.PolicyRule{
			{
				ID:     uuid.New(),
				Name:   "Same Tenant Access",
				Effect: types.PolicyEffectPermit,
				Condition: &models.PolicyCondition{
					Expression: "subject.tenant_id == resource.tenant_id",
				},
			},
			{
				ID:     uuid.New(),
				Name:   "System Admin Access",
				Effect: types.PolicyEffectPermit,
				Condition: &models.PolicyCondition{
					Expression: "subject.role == 'system_admin'",
				},
			},
		},
	}

	policyRepo.On("GetPolicyByID", mock.Anything, mock.Anything).Return(multiTenantPolicy, nil)

	// Test 1: User accessing resource in same tenant
	t.Run("SameTenantAccess", func(t *testing.T) {
		request := &PolicyEvaluationRequest{
			RequestID: uuid.New(),
			PolicyID:  multiTenantPolicy.ID,
			Subject: SubjectContext{
				UserID: user1ID,
				Attributes: map[string]interface{}{
					"tenant_id": tenant1ID.String(),
					"role":      "user",
				},
			},
			Resource: ResourceContext{
				ResourceID: resource1ID,
				Attributes: map[string]interface{}{
					"tenant_id": tenant1ID.String(),
				},
			},
			Action: ActionContext{
				Action: "read",
			},
		}

		result, err := evaluationEngine.EvaluatePolicy(ctx, request)
		require.NoError(t, err)
		assert.Equal(t, types.PolicyDecisionPermit, result.Decision)
	})

	// Test 2: User accessing resource in different tenant (should be denied)
	t.Run("CrossTenantAccess", func(t *testing.T) {
		request := &PolicyEvaluationRequest{
			RequestID: uuid.New(),
			PolicyID:  multiTenantPolicy.ID,
			Subject: SubjectContext{
				UserID: user1ID,
				Attributes: map[string]interface{}{
					"tenant_id": tenant1ID.String(),
					"role":      "user",
				},
			},
			Resource: ResourceContext{
				ResourceID: resource2ID,
				Attributes: map[string]interface{}{
					"tenant_id": tenant2ID.String(), // Different tenant
				},
			},
			Action: ActionContext{
				Action: "read",
			},
		}

		result, err := evaluationEngine.EvaluatePolicy(ctx, request)
		require.NoError(t, err)
		assert.Equal(t, types.PolicyDecisionDeny, result.Decision)
	})

	// Test 3: System admin accessing any resource (should be permitted)
	t.Run("SystemAdminAccess", func(t *testing.T) {
		request := &PolicyEvaluationRequest{
			RequestID: uuid.New(),
			PolicyID:  multiTenantPolicy.ID,
			Subject: SubjectContext{
				UserID: systemAdminID,
				Attributes: map[string]interface{}{
					"tenant_id": "system", // System admin belongs to system tenant
					"role":      "system_admin",
				},
			},
			Resource: ResourceContext{
				ResourceID: resource2ID,
				Attributes: map[string]interface{}{
					"tenant_id": tenant2ID.String(), // Can access any tenant
				},
			},
			Action: ActionContext{
				Action: "read",
			},
		}

		result, err := evaluationEngine.EvaluatePolicy(ctx, request)
		require.NoError(t, err)
		assert.Equal(t, types.PolicyDecisionPermit, result.Decision)
	})

	t.Log("Multi-tenant access scenario completed successfully")
}
