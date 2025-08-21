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

// EmployeeRepositoryTestSuite implements IAM-REPO-001: EmployeeRepository CRUD operations
// Covers employee-specific operations, manager hierarchy validation, and department associations
type EmployeeRepositoryTestSuite struct {
	suite.Suite
	ctx      context.Context
	store    *db.MockStore
	repo     EmployeeRepository
	ctrl     *gomock.Controller
	logger   *logger.MockLogger
	metrics  *metrics.MockMetricsProvider
	tracing  *tracing.MockTracingService
	tenantID uuid.UUID
	entityID uuid.UUID
	personID uuid.UUID
}

// SetupTest initializes test fixtures for each test
func (s *EmployeeRepositoryTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.ctrl = gomock.NewController(s.T())
	s.store = db.NewMockStore(s.ctrl)
	s.logger = logger.NewMockLogger(s.ctrl)
	s.metrics = metrics.NewMockMetricsProvider(s.ctrl)
	s.tracing = tracing.NewMockTracingService(s.ctrl)
	s.tenantID = uuid.New()
	s.entityID = uuid.New()
	s.personID = uuid.New()
	s.repo = NewEmployeeRepository(s.store, s.logger, s.metrics, s.tracing)
}

// TearDownTest cleans up test fixtures
func (s *EmployeeRepositoryTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// TestEmployeeRepository runs the employee repository test suite
func TestEmployeeRepository(t *testing.T) {
	suite.Run(t, new(EmployeeRepositoryTestSuite))
}

// TestCreateEmployee implements IAM-REPO-001: Verify Employee creation with Person linking
func (s *EmployeeRepositoryTestSuite) TestCreateEmployee() {
	testCases := []struct {
		name            string
		setupEmployee   func() *model.Employee
		setupMocks      func()
		expectError     bool
		errorContains   string
		validateResult  func(*testing.T, *model.Employee)
	}{
		{
			name: "IAM-REPO-001_ValidEmployeeCreation_WithPersonLink",
			setupEmployee: func() *model.Employee {
				positionTitle := "Software Engineer"
				deptID := uuid.New()
				managerID := uuid.New()
				
				return &model.Employee{
					ID:               uuid.New(),
					TenantID:         s.tenantID,
					PersonID:         s.personID,
					EmployeeNumber:   "EMP001",
					EntityID:         s.entityID,
					PositionTitle:    &positionTitle,
					DepartmentID:     &deptID,
					ManagerID:        &managerID,
					HireDate:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					EmploymentStatus: model.EmploymentStatusActive,
					SecurityLevel:    2,
					AccessAttributes: map[string]any{
						"clearance":  "SECRET",
						"department": "Engineering",
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
			},
			setupMocks: func() {
				s.store.EXPECT().
					CreateEmployee(gomock.Any(), gomock.Any()).
					Return(db.Employee{
						ID:               uuid.New(),
						TenantID:         s.tenantID,
						PersonID:         s.personID,
						EmployeeNumber:   "EMP001",
						EntityID:         s.entityID,
						EmploymentStatus: string(model.EmploymentStatusActive),
						SecurityLevel:    2,
					}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, employee *model.Employee) {
				require.NotNil(t, employee)
				require.Equal(t, s.tenantID, employee.TenantID)
				require.Equal(t, s.personID, employee.PersonID)
				require.Equal(t, "EMP001", employee.EmployeeNumber)
				require.Equal(t, model.EmploymentStatusActive, employee.EmploymentStatus)
				require.True(t, employee.IsActive())
				require.NotNil(t, employee.ManagerID)
				require.Equal(t, 2, employee.SecurityLevel)
			},
		},
		{
			name: "IAM-REPO-001_DuplicateEmployeeNumber_WithinTenant_ReturnsError",
			setupEmployee: func() *model.Employee {
				return &model.Employee{
					ID:               uuid.New(),
					TenantID:         s.tenantID,
					PersonID:         s.personID,
					EmployeeNumber:   "DUPLICATE001",
					EntityID:         s.entityID,
					HireDate:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					EmploymentStatus: model.EmploymentStatusActive,
					SecurityLevel:    1,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				}
			},
			setupMocks: func() {
				s.store.EXPECT().
					CreateEmployee(gomock.Any(), gomock.Any()).
					Return(db.Employee{}, &db.Error{Code: "23505", Message: "duplicate key value violates unique constraint"}).
					Times(1)
			},
			expectError:   true,
			errorContains: "duplicate key value violates unique constraint",
		},
		{
			name: "IAM-REPO-001_InvalidPersonID_ReturnsError",
			setupEmployee: func() *model.Employee {
				return &model.Employee{
					ID:               uuid.New(),
					TenantID:         s.tenantID,
					PersonID:         uuid.New(), // Non-existent person
					EmployeeNumber:   "EMP002",
					EntityID:         s.entityID,
					HireDate:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					EmploymentStatus: model.EmploymentStatusActive,
					SecurityLevel:    1,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				}
			},
			setupMocks: func() {
				s.store.EXPECT().
					CreateEmployee(gomock.Any(), gomock.Any()).
					Return(db.Employee{}, &db.Error{Code: "23503", Message: "foreign key violation"}).
					Times(1)
			},
			expectError:   true,
			errorContains: "foreign key violation",
		},
		{
			name: "IAM-REPO-001_TerminatedEmployee_WithTerminationDate",
			setupEmployee: func() *model.Employee {
				terminationDate := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
				return &model.Employee{
					ID:               uuid.New(),
					TenantID:         s.tenantID,
					PersonID:         s.personID,
					EmployeeNumber:   "EMP003",
					EntityID:         s.entityID,
					HireDate:         time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
					TerminationDate:  &terminationDate,
					EmploymentStatus: model.EmploymentStatusTerminated,
					SecurityLevel:    0,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				}
			},
			setupMocks: func() {
				s.store.EXPECT().
					CreateEmployee(gomock.Any(), gomock.Any()).
					Return(db.Employee{
						ID:               uuid.New(),
						TenantID:         s.tenantID,
						PersonID:         s.personID,
						EmployeeNumber:   "EMP003",
						EmploymentStatus: string(model.EmploymentStatusTerminated),
						SecurityLevel:    0,
					}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, employee *model.Employee) {
				require.NotNil(t, employee)
				require.Equal(t, model.EmploymentStatusTerminated, employee.EmploymentStatus)
				require.False(t, employee.IsActive())
				require.Equal(t, 0, employee.SecurityLevel)
				require.NotNil(t, employee.TerminationDate)
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
			employee := tc.setupEmployee()
			
			// Act
			result, err := s.repo.Create(s.ctx, employee)
			
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

// TestGetEmployeeByPersonID implements IAM-REPO-001: Verify Employee retrieval by Person ID
func (s *EmployeeRepositoryTestSuite) TestGetEmployeeByPersonID() {
	testCases := []struct {
		name            string
		personID        uuid.UUID
		setupMocks      func()
		expectError     bool
		errorContains   string
		validateResult  func(*testing.T, *model.Employee)
	}{
		{
			name:     "IAM-REPO-001_ExistingEmployee_ReturnsEmployee",
			personID: s.personID,
			setupMocks: func() {
				s.store.EXPECT().
					GetEmployeeByPersonID(gomock.Any(), s.personID).
					Return(db.Employee{
						ID:               uuid.New(),
						TenantID:         s.tenantID,
						PersonID:         s.personID,
						EmployeeNumber:   "EMP001",
						EntityID:         s.entityID,
						EmploymentStatus: string(model.EmploymentStatusActive),
						SecurityLevel:    2,
					}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, employee *model.Employee) {
				require.NotNil(t, employee)
				require.Equal(t, s.personID, employee.PersonID)
				require.Equal(t, s.tenantID, employee.TenantID)
				require.Equal(t, "EMP001", employee.EmployeeNumber)
				require.True(t, employee.IsActive())
			},
		},
		{
			name:     "IAM-REPO-001_PersonNotEmployee_ReturnsError",
			personID: uuid.New(),
			setupMocks: func() {
				s.store.EXPECT().
					GetEmployeeByPersonID(gomock.Any(), gomock.Any()).
					Return(db.Employee{}, &db.Error{Code: "02000", Message: "no data found"}).
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
			result, err := s.repo.GetByPersonID(s.ctx, tc.personID)
			
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

// TestGetEmployeeByEmployeeNumber implements IAM-REPO-001: Verify Employee retrieval by employee number
func (s *EmployeeRepositoryTestSuite) TestGetEmployeeByEmployeeNumber() {
	testCases := []struct {
		name            string
		employeeNumber  string
		setupMocks      func()
		expectError     bool
		errorContains   string
		validateResult  func(*testing.T, *model.Employee)
	}{
		{
			name:           "IAM-REPO-001_ExistingEmployeeNumber_ReturnsEmployee",
			employeeNumber: "EMP001",
			setupMocks: func() {
				s.store.EXPECT().
					GetEmployeeByNumber(gomock.Any(), "EMP001").
					Return(db.Employee{
						ID:               uuid.New(),
						TenantID:         s.tenantID,
						PersonID:         s.personID,
						EmployeeNumber:   "EMP001",
						EntityID:         s.entityID,
						EmploymentStatus: string(model.EmploymentStatusActive),
						SecurityLevel:    2,
					}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, employee *model.Employee) {
				require.NotNil(t, employee)
				require.Equal(t, "EMP001", employee.EmployeeNumber)
				require.Equal(t, s.tenantID, employee.TenantID)
				require.True(t, employee.IsActive())
			},
		},
		{
			name:           "IAM-REPO-001_EmployeeNumberNotFound_ReturnsError",
			employeeNumber: "NOTFOUND",
			setupMocks: func() {
				s.store.EXPECT().
					GetEmployeeByNumber(gomock.Any(), "NOTFOUND").
					Return(db.Employee{}, &db.Error{Code: "02000", Message: "no data found"}).
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
			result, err := s.repo.GetByEmployeeNumber(s.ctx, tc.employeeNumber)
			
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

// TestListEmployeesByStatus implements IAM-REPO-001: Verify Employee listing by employment status
func (s *EmployeeRepositoryTestSuite) TestListEmployeesByStatus() {
	testCases := []struct {
		name            string
		status          model.EmploymentStatus
		setupMocks      func()
		expectError     bool
		errorContains   string
		validateResult  func(*testing.T, []*model.Employee)
	}{
		{
			name:   "IAM-REPO-001_ActiveEmployees_ReturnsActiveList",
			status: model.EmploymentStatusActive,
			setupMocks: func() {
				s.store.EXPECT().
					ListEmployeesByStatus(gomock.Any(), string(model.EmploymentStatusActive)).
					Return([]db.Employee{
						{
							ID:               uuid.New(),
							TenantID:         s.tenantID,
							PersonID:         uuid.New(),
							EmployeeNumber:   "EMP001",
							EmploymentStatus: string(model.EmploymentStatusActive),
							SecurityLevel:    2,
						},
						{
							ID:               uuid.New(),
							TenantID:         s.tenantID,
							PersonID:         uuid.New(),
							EmployeeNumber:   "EMP002",
							EmploymentStatus: string(model.EmploymentStatusActive),
							SecurityLevel:    1,
						},
					}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, employees []*model.Employee) {
				require.Len(t, employees, 2)
				for _, emp := range employees {
					require.Equal(t, model.EmploymentStatusActive, emp.EmploymentStatus)
					require.True(t, emp.IsActive())
					require.Equal(t, s.tenantID, emp.TenantID)
				}
			},
		},
		{
			name:   "IAM-REPO-001_TerminatedEmployees_ReturnsTerminatedList",
			status: model.EmploymentStatusTerminated,
			setupMocks: func() {
				s.store.EXPECT().
					ListEmployeesByStatus(gomock.Any(), string(model.EmploymentStatusTerminated)).
					Return([]db.Employee{
						{
							ID:               uuid.New(),
							TenantID:         s.tenantID,
							PersonID:         uuid.New(),
							EmployeeNumber:   "EMP999",
							EmploymentStatus: string(model.EmploymentStatusTerminated),
							SecurityLevel:    0,
						},
					}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, employees []*model.Employee) {
				require.Len(t, employees, 1)
				require.Equal(t, model.EmploymentStatusTerminated, employees[0].EmploymentStatus)
				require.False(t, employees[0].IsActive())
				require.Equal(t, 0, employees[0].SecurityLevel)
			},
		},
		{
			name:   "IAM-REPO-001_NoEmployeesForStatus_ReturnsEmptyList",
			status: model.EmploymentStatusOnLeave,
			setupMocks: func() {
				s.store.EXPECT().
					ListEmployeesByStatus(gomock.Any(), string(model.EmploymentStatusOnLeave)).
					Return([]db.Employee{}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, employees []*model.Employee) {
				require.Empty(t, employees)
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
			result, err := s.repo.ListByStatus(s.ctx, tc.status)
			
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

// TestUpdateEmployee implements IAM-REPO-001: Verify Employee update operations
func (s *EmployeeRepositoryTestSuite) TestUpdateEmployee() {
	testCases := []struct {
		name            string
		setupEmployee   func() *model.Employee
		setupMocks      func()
		expectError     bool
		errorContains   string
		validateResult  func(*testing.T, *model.Employee)
	}{
		{
			name: "IAM-REPO-001_ValidUpdate_UpdatesEmployeeRecord",
			setupEmployee: func() *model.Employee {
				newPosition := "Senior Software Engineer"
				newManagerID := uuid.New()
				return &model.Employee{
					ID:               uuid.New(),
					TenantID:         s.tenantID,
					PersonID:         s.personID,
					EmployeeNumber:   "EMP001",
					EntityID:         s.entityID,
					PositionTitle:    &newPosition,
					ManagerID:        &newManagerID,
					HireDate:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					EmploymentStatus: model.EmploymentStatusActive,
					SecurityLevel:    3, // Promoted
					CreatedAt:        time.Now().Add(-time.Hour),
					UpdatedAt:        time.Now(),
				}
			},
			setupMocks: func() {
				s.store.EXPECT().
					UpdateEmployee(gomock.Any(), gomock.Any()).
					Return(db.Employee{
						ID:               uuid.New(),
						TenantID:         s.tenantID,
						PersonID:         s.personID,
						EmployeeNumber:   "EMP001",
						EmploymentStatus: string(model.EmploymentStatusActive),
						SecurityLevel:    3,
						UpdatedAt:        time.Now(),
					}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, employee *model.Employee) {
				require.NotNil(t, employee)
				require.Equal(t, 3, employee.SecurityLevel)
				require.NotZero(t, employee.UpdatedAt)
			},
		},
		{
			name: "IAM-REPO-001_EmployeeNotFound_ReturnsError",
			setupEmployee: func() *model.Employee {
				return &model.Employee{
					ID:        uuid.New(),
					TenantID:  s.tenantID,
					PersonID:  uuid.New(),
					FirstName: "NotFound",
					LastName:  "Employee",
				}
			},
			setupMocks: func() {
				s.store.EXPECT().
					UpdateEmployee(gomock.Any(), gomock.Any()).
					Return(db.Employee{}, &db.Error{Code: "02000", Message: "no data found"}).
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
			
			// Arrange
			employee := tc.setupEmployee()
			
			// Act
			result, err := s.repo.Update(s.ctx, employee)
			
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

// TestEmployeeTenantIsolation implements IAM-REPO-001: Verify strict tenant isolation for employees
func (s *EmployeeRepositoryTestSuite) TestEmployeeTenantIsolation() {
	testCases := []struct {
		name           string
		setupScenario  func() (tenantA, tenantB uuid.UUID)
		validateResult func(*testing.T, uuid.UUID, uuid.UUID)
	}{
		{
			name: "IAM-REPO-001_CrossTenantEmployeeAccess_ZeroDataLeakage",
			setupScenario: func() (uuid.UUID, uuid.UUID) {
				tenantA := uuid.New()
				tenantB := uuid.New()
				
				// Mock RLS enforcement - no cross-tenant employee access
				s.store.EXPECT().
					GetEmployeeByPersonID(gomock.Any(), gomock.Any()).
					Return(db.Employee{}, &db.Error{Code: "02000", Message: "no data found"}).
					Times(1)
				
				return tenantA, tenantB
			},
			validateResult: func(t *testing.T, tenantA, tenantB uuid.UUID) {
				// Try to access tenant A's employee from tenant B context
				personID := uuid.New()
				
				// This should fail due to RLS
				employee, err := s.repo.GetByPersonID(s.ctx, personID)
				require.Error(t, err)
				require.Nil(t, employee)
				require.Contains(t, err.Error(), "no data found")
			},
		},
		{
			name: "IAM-REPO-001_SameEmployeeNumberDifferentTenants_AllowedIsolation",
			setupScenario: func() (uuid.UUID, uuid.UUID) {
				tenantA := uuid.New()
				tenantB := uuid.New()
				
				// Mock that same employee number can exist in different tenants
				s.store.EXPECT().
					GetEmployeeByNumber(gomock.Any(), "EMP001").
					Return(db.Employee{
						ID:             uuid.New(),
						TenantID:       tenantA, // Current tenant context
						EmployeeNumber: "EMP001",
					}, nil).
					Times(1)
				
				return tenantA, tenantB
			},
			validateResult: func(t *testing.T, tenantA, tenantB uuid.UUID) {
				// Should only find employee from current tenant (A), not tenant B
				employee, err := s.repo.GetByEmployeeNumber(s.ctx, "EMP001")
				require.NoError(t, err)
				require.NotNil(t, employee)
				require.Equal(t, tenantA, employee.TenantID)
				require.NotEqual(t, tenantB, employee.TenantID)
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
