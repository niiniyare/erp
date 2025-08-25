package abac

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// AttributeServiceTestSuite defines test suite for ABAC attribute service operations
type AttributeServiceTestSuite struct {
	suite.Suite
	ctx           context.Context
	service       AttributeService
	mockRepo      *repository.MockAttributeRepository
	ctrl          *gomock.Controller
	mockLogger    logger.Logger
	mockMetrics   metrics.MetricsProvider
	mockTracer    *tracing.MockTracingService
	encryptionKey []byte
}

// SetupTest initializes test fixtures for each test
func (s *AttributeServiceTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.ctrl = gomock.NewController(s.T())

	// Setup mocks
	s.mockRepo = repository.NewMockAttributeRepository(s.ctrl)
	s.mockLogger = logger.WithFields(logger.Fields{})
	s.mockMetrics = &metrics.MetricsService{}
	s.mockTracer = tracing.NewMockTracingService(s.ctrl)

	// Setup encryption key
	s.encryptionKey = []byte("test-key-for-encryption-32-byte")

	// Create service
	s.service = NewAttributeService(
		s.mockRepo,
		s.encryptionKey,
		s.mockLogger,
		s.mockMetrics,
		s.mockTracer,
	)
}

// TearDownTest cleans up after each test
func (s *AttributeServiceTestSuite) TearDownTest() {
	if s.ctrl != nil {
		s.ctrl.Finish()
	}
}

// TestAttributeService runs the attribute service test suite
// func TestAttributeService(t *testing.T) {
// 	suite.Run(t, new(AttributeServiceTestSuite))
// }

// TestCreateAttributeDefinition implements ABAC-ATTR-001: Attribute Definition Creation
func (s *AttributeServiceTestSuite) TestCreateAttributeDefinition() {
	testCases := []struct {
		name           string
		spec           string
		request        *CreateAttributeDefinitionRequest
		mockSetup      func()
		expectedErr    string
		validateResult func(*testing.T, *AttributeDefinitionResult)
	}{
		{
			name: "ValidRequest_ReturnsAttributeDefinition",
			spec: "ABAC-ATTR-001",
			request: &CreateAttributeDefinitionRequest{
				Name:         "department",
				DisplayName:  stringPtr("Department"),
				Description:  stringPtr("User's department in the organization"),
				DataType:     types.AttributeDataTypeString,
				Category:     types.AttributeCategoryUser,
				IsRequired:   true,
				DefaultValue: stringPtr("unassigned"),
				ValidationRules: []AttributeValidationRule{
					RuleID:      uuid.New(),
					RuleType:    AttributeRuleTypeLength,
					RuleName:    "ValidRequest_ReturnsAttributeDefinition",
					Description: "ValidRequest_ReturnsAttributeDefinition",

					Parameters: map[string]any{
						"AllowedValues": []any{"engineering", "marketing", "sales", "hr"},
					},

					ErrorMessage: "",

					IsActive: true,

					"AllowedValues": []any{"engineering", "marketing", "sales", "hr"},
					MinLength:       int32Ptr(1),
					MaxLength:       int32Ptr(50),
				},
				SecuritySettings: AttributeSecuritySettings{
					EncryptionRequired: false,
					AuditingRequired:   false,
					AccessLevel:        "standard",
				},
				Metadata: map[string]any{
					"source": "hr_system",
				},
				Tags:      []string{"user", "department", "organizational"},
				CreatedBy: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			mockSetup: func() {
				expectedResult := &AttributeDefinitionResult{
					AttributeDefinition: &models.AttributeDefinition{
						ID:          uuid.New(),
						Name:        "department",
						DisplayName: stringPtr("Department"),
						Description: stringPtr("User's department in the organization"),
						DataType:    types.AttributeDataTypeString,
						Category:    types.AttributeCategoryUser,
						IsRequired:  true,
					},
				}
				s.mockRepo.EXPECT().
					CreateAttributeDefinition(gomock.Any(), gomock.Any()).
					Return(expectedResult, nil)
			},
			validateResult: func(t *testing.T, result *AttributeDefinitionResult) {
				require.NotNil(t, result)
				require.NotNil(t, result.AttributeDefinition)
				require.Equal(t, "department", result.AttributeDefinition.Name)
				require.Equal(t, types.AttributeDataTypeString, result.AttributeDefinition.DataType)
				require.Equal(t, types.AttributeCategoryUser, result.AttributeDefinition.Category)
				require.Equal(t, true, result.AttributeDefinition.IsRequired)
			},
		},
		{
			name: "EmptyName_ReturnsValidationError",
			spec: "ABAC-ATTR-001",
			request: &CreateAttributeDefinitionRequest{
				Name:     "", // Invalid: empty name
				DataType: types.AttributeDataTypeString,
				Category: types.AttributeCategoryUser,
			},
			mockSetup:   func() {}, // No mock calls expected
			expectedErr: "validation failed",
		},
		{
			name: "InvalidDataType_ReturnsValidationError",
			spec: "ABAC-ATTR-001",
			request: &CreateAttributeDefinitionRequest{
				Name:     "test_attribute",
				DataType: "invalid_type", // Invalid data type
				Category: types.AttributeCategoryUser,
			},
			mockSetup:   func() {},
			expectedErr: "validation failed",
		},
		{
			name: "DuplicateName_ReturnsError",
			spec: "ABAC-ATTR-001",
			request: &CreateAttributeDefinitionRequest{
				Name:     "existing_attribute",
				DataType: types.AttributeDataTypeString,
				Category: types.AttributeCategoryUser,
			},
			mockSetup: func() {
				s.mockRepo.EXPECT().
					CreateAttributeDefinition(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("attribute definition with name 'existing_attribute' already exists"))
			},
			expectedErr: "already exists",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.spec+"_"+tc.name, func() {
			// Setup mocks for this test case
			tc.mockSetup()

			// Execute test
			result, err := s.service.CreateAttributeDefinition(s.ctx, tc.request)

			// Validate results
			if tc.expectedErr != "" {
				require.Error(s.T(), err)
				require.Contains(s.T(), err.Error(), tc.expectedErr)
				require.Nil(s.T(), result)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), result)
				if tc.validateResult != nil {
					tc.validateResult(s.T(), result)
				}
			}
		})
	}
}

// TestValidateAttributeValue implements ABAC-ATTR-002: Attribute Value Validation
func (s *AttributeServiceTestSuite) TestValidateAttributeValue() {
	testCases := []struct {
		name           string
		spec           string
		request        *ValidateAttributeValueRequest
		mockSetup      func()
		expectedErr    string
		validateResult func(*testing.T, *AttributeValidationResult)
	}{
		{
			name: "ValidStringValue_PassesValidation",
			spec: "ABAC-ATTR-002",
			request: &ValidateAttributeValueRequest{
				AttributeID:     uuid.New(),
				Value:           "engineering",
				ValidationLevel: "strict",
			},
			mockSetup: func() {
				// Mock setup for valid validation
			},
			validateResult: func(t *testing.T, result *AttributeValidationResult) {
				require.NotNil(t, result)
				require.True(t, result.IsValid)
				require.Empty(t, result.ValidationErrors)
			},
		},
		{
			name: "InvalidValue_FailsValidation",
			spec: "ABAC-ATTR-002",
			request: &ValidateAttributeValueRequest{
				AttributeID:     uuid.New(),
				Value:           "invalid_department",
				ValidationLevel: "strict",
			},
			mockSetup: func() {
				// Mock setup for invalid validation
			},
			validateResult: func(t *testing.T, result *AttributeValidationResult) {
				require.NotNil(t, result)
				require.False(t, result.IsValid)
				require.NotEmpty(t, result.ValidationErrors)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.spec+"_"+tc.name, func() {
			// Setup mocks for this test case
			tc.mockSetup()

			// Execute test
			result, err := s.service.ValidateAttributeValue(s.ctx, tc.request)

			// Validate results
			if tc.expectedErr != "" {
				require.Error(s.T(), err)
				require.Contains(s.T(), err.Error(), tc.expectedErr)
				require.Nil(s.T(), result)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), result)
				if tc.validateResult != nil {
					tc.validateResult(s.T(), result)
				}
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}
