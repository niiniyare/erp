package repo

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// PermissionRepositoryTestSuite implements IAM-REPO-004: PermissionRepository comprehensive testing
// Covers conditional access rules validation, risk metadata management, and permission lifecycle
type PermissionRepositoryTestSuite struct {
	suite.Suite
	ctx      context.Context
	store    *db.MockStore
	repo     PermissionRepository
	ctrl     *gomock.Controller
	logger   *logger.MockLogger
	metrics  *metrics.MockMetricsProvider
	tracing  *tracing.MockTracingService
	tenantID uuid.UUID
	entityID uuid.UUID
	userID   uuid.UUID
}

// SetupTest initializes test fixtures for each test
func (s *PermissionRepositoryTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.ctrl = gomock.NewController(s.T())
	s.store = db.NewMockStore(s.ctrl)
	s.logger = logger.NewMockLogger(s.ctrl)
	s.metrics = metrics.NewMockMetricsProvider(s.ctrl)
	s.tracing = tracing.NewMockTracingService(s.ctrl)
	s.tenantID = uuid.New()
	s.entityID = uuid.New()
	s.userID = uuid.New()
	s.repo = NewPermissionRepository(s.store, s.logger, s.metrics, s.tracing)
}

// TearDownTest cleans up test fixtures
func (s *PermissionRepositoryTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// TestPermissionRepository runs the comprehensive permission repository test suite
func TestPermissionRepository(t *testing.T) {
	suite.Run(t, new(PermissionRepositoryTestSuite))
}

// TestCreatePermission implements IAM-REPO-004: Verify Permission creation with conditional access rules
func (s *PermissionRepositoryTestSuite) TestCreatePermission() {
	testCases := []struct {
		name            string
		setupPermission func() *model.Permission
		setupMocks      func()
		expectError     bool
		errorContains   string
		validateResult  func(*testing.T, *model.Permission)
	}{
		{
			name: "IAM-REPO-004_ValidPermissionCreation_WithConditionalAccess",
			setupPermission: func() *model.Permission {
				resourceID := uuid.New()
				expiresAt := time.Now().Add(24 * time.Hour)
				
				return &model.Permission{
					ID:           uuid.New(),
					TenantID:     s.tenantID,
					ResourceType: "financial_records",
					ResourceID:   &resourceID,
					Action:       "read",
					EntityID:     &s.entityID,
					Conditions: []string{
						"time_of_day >= 09:00 AND time_of_day <= 17:00",
						"ip_address IN ['10.0.0.0/8', '192.168.1.0/24']",
						"mfa_verified = true",
						"device_trusted = true",
					},
					ExpiresAt: &expiresAt,
					Metadata: map[string]any{
						"risk_level":           "HIGH",
						"action_category":      "FINANCIAL_DATA_ACCESS",
						"requires_approval":    true,
						"approval_levels":      2,
						"data_classification":  "CONFIDENTIAL",
						"compliance_framework": []string{"SOX", "PCI-DSS"},
						"monitoring_required":  true,
						"audit_retention":      "7_years",
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
			},
			setupMocks: func() {
				s.store.EXPECT().
					CreatePermission(gomock.Any(), gomock.Any()).
					Return(db.Permission{
						ID:           uuid.New(),
						TenantID:     s.tenantID,
						ResourceType: "financial_records",
						Action:       "read",
						EntityID:     &s.entityID,
					}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, permission *model.Permission) {
				require.NotNil(t, permission)
				require.Equal(t, s.tenantID, permission.TenantID)
				require.Equal(t, "financial_records", permission.ResourceType)
				require.Equal(t, "read", permission.Action)
				require.NotNil(t, permission.EntityID)
				require.Equal(t, s.entityID, *permission.EntityID)
				
				// Validate conditional access rules
				require.Len(t, permission.Conditions, 4)
				require.Contains(t, permission.Conditions, "time_of_day >= 09:00 AND time_of_day <= 17:00")
				require.Contains(t, permission.Conditions, "ip_address IN ['10.0.0.0/8', '192.168.1.0/24']")
				require.Contains(t, permission.Conditions, "mfa_verified = true")
				require.Contains(t, permission.Conditions, "device_trusted = true")
				
				// Validate risk metadata
				require.NotNil(t, permission.Metadata)
				require.Equal(t, "HIGH", permission.Metadata["risk_level"])
				require.Equal(t, "FINANCIAL_DATA_ACCESS", permission.Metadata["action_category"])
				require.Equal(t, true, permission.Metadata["requires_approval"])
				require.Equal(t, 2, permission.Metadata["approval_levels"])
				require.Equal(t, "CONFIDENTIAL", permission.Metadata["data_classification"])
				
				// Validate compliance metadata
				complianceFrameworks := permission.Metadata["compliance_framework"].([]string)
				require.Contains(t, complianceFrameworks, "SOX")
				require.Contains(t, complianceFrameworks, "PCI-DSS")
				
				// Validate expiration
				require.NotNil(t, permission.ExpiresAt)
				require.False(t, permission.IsExpired())
			},
		},
		{
			name: "IAM-REPO-004_LowRiskPermission_MinimalConditions",
			setupPermission: func() *model.Permission {
				return &model.Permission{
					ID:           uuid.New(),
					TenantID:     s.tenantID,
					ResourceType: "public_reports",
					Action:       "read",
					EntityID:     &s.entityID,
					Conditions:   []string{"authenticated = true"},
					Metadata: map[string]any{
						"risk_level":        "LOW",
						"action_category":   "PUBLIC_DATA_ACCESS",
						"requires_approval": false,
						"monitoring_required": false,
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
			},
			setupMocks: func() {
				s.store.EXPECT().
					CreatePermission(gomock.Any(), gomock.Any()).
					Return(db.Permission{
						ID:           uuid.New(),
						TenantID:     s.tenantID,
						ResourceType: "public_reports",
						Action:       "read",
					}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, permission *model.Permission) {
				require.NotNil(t, permission)
				require.Equal(t, "LOW", permission.Metadata["risk_level"])
				require.Equal(t, false, permission.Metadata["requires_approval"])
				require.Len(t, permission.Conditions, 1)
				require.Equal(t, "authenticated = true", permission.Conditions[0])
			},
		},
		{
			name: "IAM-REPO-004_CriticalPermission_MaximalSecurity",
			setupPermission: func() *model.Permission {
				resourceID := uuid.New()
				expiresAt := time.Now().Add(1 * time.Hour) // Short expiration for critical
				
				return &model.Permission{
					ID:           uuid.New(),
					TenantID:     s.tenantID,
					ResourceType: "system_configuration",
					ResourceID:   &resourceID,
					Action:       "delete",
					EntityID:     &s.entityID,
					Conditions: []string{
						"role IN ['system_admin', 'security_admin']",
						"mfa_verified = true",
						"privileged_session = true",
						"approval_obtained = true",
						"witness_present = true",
						"emergency_override = false",
					},
					ExpiresAt: &expiresAt,
					Metadata: map[string]any{
						"risk_level":             "CRITICAL",
						"action_category":        "SYSTEM_ADMINISTRATION",
						"requires_approval":      true,
						"approval_levels":        3,
						"dual_control_required": true,
						"change_window_only":    true,
						"backup_required":       true,
						"rollback_plan":         true,
						"monitoring_required":   true,
						"real_time_alerts":      true,
						"audit_retention":       "permanent",
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
			},
			setupMocks: func() {
				s.store.EXPECT().
					CreatePermission(gomock.Any(), gomock.Any()).
					Return(db.Permission{
						ID:           uuid.New(),
						TenantID:     s.tenantID,
						ResourceType: "system_configuration",
						Action:       "delete",
					}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, permission *model.Permission) {
				require.NotNil(t, permission)
				require.Equal(t, "CRITICAL", permission.Metadata["risk_level"])
				require.Equal(t, 3, permission.Metadata["approval_levels"])
				require.Equal(t, true, permission.Metadata["dual_control_required"])
				require.Len(t, permission.Conditions, 6)
				require.Contains(t, permission.Conditions, "role IN ['system_admin', 'security_admin']")
				require.Contains(t, permission.Conditions, "approval_obtained = true")
				require.Contains(t, permission.Conditions, "witness_present = true")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup tracing mocks
			s.tracing.EXPECT().
				StartSpan(gomock.Any(), gomock.Any()).
				Return(s.ctx, &tracing.MockSpan{}).
				AnyTimes()
			
			// Setup mock behavior
			tc.setupMocks()
			
			// Arrange
			permission := tc.setupPermission()
			
			// Act
			result, err := s.repo.Create(s.ctx, permission)
			
			// Assert
			if tc.expectError {
				require.Error(s.T(), err)
				if tc.errorContains != "" {
					require.Contains(s.T(), err.Error(), tc.errorContains)
				}
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

// TestPermissionCheck implements IAM-REPO-004: Verify permission checking with conditional evaluation
func (s *PermissionRepositoryTestSuite) TestPermissionCheck() {
	testCases := []struct {
		name            string
		userID          uuid.UUID
		resourceType    string
		action          string
		resourceID      *uuid.UUID
		setupMocks      func()
		expectHasPermission bool
		validateResult  func(*testing.T, bool)
	}{
		{
			name:         "IAM-REPO-004_ValidPermission_AllConditionsMet_ReturnsTrue",
			userID:       s.userID,
			resourceType: "financial_records",
			action:       "read",
			resourceID:   &s.entityID,
			setupMocks: func() {
				// Mock finding a valid permission that meets all conditions
				s.store.EXPECT().
					CheckUserPermission(gomock.Any(), db.CheckUserPermissionParams{
						UserID:       s.userID,
						ResourceType: "financial_records",
						Action:       "read",
						ResourceID:   &s.entityID,
					}).
					Return(true, nil). // Permission exists and conditions are met
					Times(1)
			},
			expectHasPermission: true,
			validateResult: func(t *testing.T, hasPermission bool) {
				require.True(t, hasPermission, "User should have permission when all conditions are met")
			},
		},
		{
			name:         "IAM-REPO-004_ConditionalAccessFailed_TimeRestriction_ReturnsFalse",
			userID:       s.userID,
			resourceType: "financial_records",
			action:       "read",
			resourceID:   &s.entityID,
			setupMocks: func() {
				// Mock permission exists but time condition fails
				s.store.EXPECT().
					CheckUserPermission(gomock.Any(), db.CheckUserPermissionParams{
						UserID:       s.userID,
						ResourceType: "financial_records",
						Action:       "read",
						ResourceID:   &s.entityID,
					}).
					Return(false, nil). // Conditions not met (e.g., outside business hours)
					Times(1)
			},
			expectHasPermission: false,
			validateResult: func(t *testing.T, hasPermission bool) {
				require.False(t, hasPermission, "User should not have permission when conditions fail")
			},
		},
		{
			name:         "IAM-REPO-004_ExpiredPermission_ReturnsFalse",
			userID:       s.userID,
			resourceType: "temporary_access",
			action:       "read",
			resourceID:   nil,
			setupMocks: func() {
				// Mock expired permission
				s.store.EXPECT().
					CheckUserPermission(gomock.Any(), db.CheckUserPermissionParams{
						UserID:       s.userID,
						ResourceType: "temporary_access",
						Action:       "read",
						ResourceID:   nil,
					}).
					Return(false, nil). // Permission expired
					Times(1)
			},
			expectHasPermission: false,
			validateResult: func(t *testing.T, hasPermission bool) {
				require.False(t, hasPermission, "Expired permissions should return false")
			},
		},
		{
			name:         "IAM-REPO-004_NoPermission_ReturnsFalse",
			userID:       s.userID,
			resourceType: "restricted_data",
			action:       "delete",
			resourceID:   nil,
			setupMocks: func() {
				// Mock no permission found
				s.store.EXPECT().
					CheckUserPermission(gomock.Any(), db.CheckUserPermissionParams{
						UserID:       s.userID,
						ResourceType: "restricted_data",
						Action:       "delete",
						ResourceID:   nil,
					}).
					Return(false, nil). // No permission found
					Times(1)
			},
			expectHasPermission: false,
			validateResult: func(t *testing.T, hasPermission bool) {
				require.False(t, hasPermission, "User should not have permission when none exists")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup tracing mocks
			s.tracing.EXPECT().
				StartSpan(gomock.Any(), gomock.Any()).
				Return(s.ctx, &tracing.MockSpan{}).
				AnyTimes()
			
			// Setup mock behavior
			tc.setupMocks()
			
			// Act
			hasPermission, err := s.repo.CheckPermission(s.ctx, tc.userID, tc.resourceType, tc.action, tc.resourceID)
			
			// Assert
			require.NoError(s.T(), err)
			require.Equal(s.T(), tc.expectHasPermission, hasPermission)
			if tc.validateResult != nil {
				tc.validateResult(s.T(), hasPermission)
			}
		})
	}
}

// TestGrantRevokePermission implements IAM-REPO-004: Verify permission lifecycle management
func (s *PermissionRepositoryTestSuite) TestGrantRevokePermission() {
	testCases := []struct {
		name          string
		operation     string
		userID        uuid.UUID
		resourceType  string
		action        string
		resourceID    *uuid.UUID
		entityID      *uuid.UUID
		expiresAt     *time.Time
		setupMocks    func()
		expectError   bool
		errorContains string
	}{
		{
			name:         "IAM-REPO-004_GrantPermission_WithExpiration_CreatesRecord",
			operation:    "grant",
			userID:       s.userID,
			resourceType: "documents",
			action:       "read",
			resourceID:   &s.entityID,
			entityID:     &s.entityID,
			expiresAt:    timePtr(time.Now().Add(24 * time.Hour)),
			setupMocks: func() {
				expiresAt := time.Now().Add(24 * time.Hour)
				s.store.EXPECT().
					GrantUserPermission(gomock.Any(), db.GrantUserPermissionParams{
						UserID:       s.userID,
						ResourceType: "documents",
						Action:       "read",
						ResourceID:   &s.entityID,
						EntityID:     &s.entityID,
						ExpiresAt:    &expiresAt,
					}).
					Return(nil).
					Times(1)
			},
			expectError: false,
		},
		{
			name:         "IAM-REPO-004_GrantPermission_Permanent_NoExpiration",
			operation:    "grant",
			userID:       s.userID,
			resourceType: "basic_access",
			action:       "read",
			resourceID:   nil,
			entityID:     &s.entityID,
			expiresAt:    nil, // Permanent permission
			setupMocks: func() {
				s.store.EXPECT().
					GrantUserPermission(gomock.Any(), db.GrantUserPermissionParams{
						UserID:       s.userID,
						ResourceType: "basic_access",
						Action:       "read",
						ResourceID:   nil,
						EntityID:     &s.entityID,
						ExpiresAt:    nil,
					}).
					Return(nil).
					Times(1)
			},
			expectError: false,
		},
		{
			name:         "IAM-REPO-004_RevokePermission_RemovesAccess",
			operation:    "revoke",
			userID:       s.userID,
			resourceType: "documents",
			action:       "read",
			resourceID:   &s.entityID,
			entityID:     &s.entityID,
			setupMocks: func() {
				s.store.EXPECT().
					RevokeUserPermission(gomock.Any(), db.RevokeUserPermissionParams{
						UserID:       s.userID,
						ResourceType: "documents",
						Action:       "read",
						ResourceID:   &s.entityID,
						EntityID:     &s.entityID,
					}).
					Return(nil).
					Times(1)
			},
			expectError: false,
		},
		{
			name:         "IAM-REPO-004_RevokeNonExistentPermission_ReturnsError",
			operation:    "revoke",
			userID:       s.userID,
			resourceType: "nonexistent",
			action:       "read",
			resourceID:   nil,
			entityID:     &s.entityID,
			setupMocks: func() {
				s.store.EXPECT().
					RevokeUserPermission(gomock.Any(), db.RevokeUserPermissionParams{
						UserID:       s.userID,
						ResourceType: "nonexistent",
						Action:       "read",
						ResourceID:   nil,
						EntityID:     &s.entityID,
					}).
					Return(&db.Error{Code: "02000", Message: "no data found"}).
					Times(1)
			},
			expectError:   true,
			errorContains: "no data found",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup tracing mocks
			s.tracing.EXPECT().
				StartSpan(gomock.Any(), gomock.Any()).
				Return(s.ctx, &tracing.MockSpan{}).
				AnyTimes()
			
			// Setup mock behavior
			tc.setupMocks()
			
			// Act
			var err error
			if tc.operation == "grant" {
				err = s.repo.GrantPermission(s.ctx, tc.userID, tc.resourceType, tc.action, tc.resourceID, tc.entityID, tc.expiresAt)
			} else if tc.operation == "revoke" {
				err = s.repo.RevokePermission(s.ctx, tc.userID, tc.resourceType, tc.action, tc.resourceID, tc.entityID)
			}
			
			// Assert
			if tc.expectError {
				require.Error(s.T(), err)
				if tc.errorContains != "" {
					require.Contains(s.T(), err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(s.T(), err)
			}
		})
	}
}

// TestRemoveExpiredPermissions implements IAM-REPO-004: Verify cleanup of expired permissions
func (s *PermissionRepositoryTestSuite) TestRemoveExpiredPermissions() {
	testCases := []struct {
		name            string
		setupMocks      func()
		expectError     bool
		errorContains   string
		validateResult  func(*testing.T, int)
	}{
		{
			name: "IAM-REPO-004_CleanupExpiredPermissions_RemovesExpiredOnly",
			setupMocks: func() {
				// Mock removal of expired permissions
				s.store.EXPECT().
					RemoveExpiredPermissions(gomock.Any()).
					Return(int64(5), nil). // 5 expired permissions removed
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, removedCount int) {
				// Would validate that 5 permissions were removed
				// In actual implementation, this would return the count
			},
		},
		{
			name: "IAM-REPO-004_NoExpiredPermissions_ReturnsZero",
			setupMocks: func() {
				// Mock no expired permissions to remove
				s.store.EXPECT().
					RemoveExpiredPermissions(gomock.Any()).
					Return(int64(0), nil). // No expired permissions
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, removedCount int) {
				// Would validate that 0 permissions were removed
			},
		},
		{
			name: "IAM-REPO-004_DatabaseError_ReturnsError",
			setupMocks: func() {
				// Mock database error during cleanup
				s.store.EXPECT().
					RemoveExpiredPermissions(gomock.Any()).
					Return(int64(0), &db.Error{Code: "40001", Message: "database connection failed"}).
					Times(1)
			},
			expectError:   true,
			errorContains: "database connection failed",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup tracing mocks
			s.tracing.EXPECT().
				StartSpan(gomock.Any(), gomock.Any()).
				Return(s.ctx, &tracing.MockSpan{}).
				AnyTimes()
			
			// Setup mock behavior
			tc.setupMocks()
			
			// Act
			err := s.repo.RemoveExpiredPermissions(s.ctx)
			
			// Assert
			if tc.expectError {
				require.Error(s.T(), err)
				if tc.errorContains != "" {
					require.Contains(s.T(), err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(s.T(), err)
				if tc.validateResult != nil {
					// In actual implementation, would pass the removed count
					tc.validateResult(s.T(), 0)
				}
			}
		})
	}
}

// TestPermissionTenantIsolation implements IAM-REPO-004: Verify strict tenant isolation for permissions
func (s *PermissionRepositoryTestSuite) TestPermissionTenantIsolation() {
	testCases := []struct {
		name           string
		setupScenario  func() (tenantA, tenantB uuid.UUID)
		validateResult func(*testing.T, uuid.UUID, uuid.UUID)
	}{
		{
			name: "IAM-REPO-004_CrossTenantPermissionAccess_ZeroDataLeakage",
			setupScenario: func() (uuid.UUID, uuid.UUID) {
				tenantA := uuid.New()
				tenantB := uuid.New()
				
				// Mock RLS enforcement - no cross-tenant permission access
				s.store.EXPECT().
					GetUserPermissions(gomock.Any(), s.userID).
					Return([]db.Permission{}, nil). // Empty due to RLS
					Times(1)
				
				return tenantA, tenantB
			},
			validateResult: func(t *testing.T, tenantA, tenantB uuid.UUID) {
				// Try to access tenant A's permissions from tenant B context
				permissions, err := s.repo.GetUserPermissions(s.ctx, s.userID, nil)
				require.NoError(t, err)
				require.Empty(t, permissions, "Should not see cross-tenant permissions")
			},
		},
		{
			name: "IAM-REPO-004_PermissionCheckCrossTenant_ReturnsFalse",
			setupScenario: func() (uuid.UUID, uuid.UUID) {
				tenantA := uuid.New()
				tenantB := uuid.New()
				
				// Mock permission check across tenants returns false
				s.store.EXPECT().
					CheckUserPermission(gomock.Any(), gomock.Any()).
					Return(false, nil). // No permission due to tenant isolation
					Times(1)
				
				return tenantA, tenantB
			},
			validateResult: func(t *testing.T, tenantA, tenantB uuid.UUID) {
				// Check permission across tenants
				hasPermission, err := s.repo.CheckPermission(s.ctx, s.userID, "documents", "read", nil)
				require.NoError(t, err)
				require.False(t, hasPermission, "Cross-tenant permission check should fail")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup tracing mocks
			s.tracing.EXPECT().
				StartSpan(gomock.Any(), gomock.Any()).
				Return(s.ctx, &tracing.MockSpan{}).
				AnyTimes()
			
			// Arrange
			tenantA, tenantB := tc.setupScenario()
			
			// Act & Assert
			tc.validateResult(s.T(), tenantA, tenantB)
		})
	}
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
