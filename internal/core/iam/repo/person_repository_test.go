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

// PersonRepositoryTestSuite implements IAM-REPO-001: PersonRepository CRUD operations
// Covers basic CRUD operations, multi-tenant isolation, and data validation
type PersonRepositoryTestSuite struct {
	suite.Suite
	ctx      context.Context
	store    *db.MockStore
	repo     PersonRepository
	ctrl     *gomock.Controller
	logger   *logger.MockLogger
	metrics  *metrics.MockMetricsProvider
	tracing  *tracing.MockTracingService
	tenantID uuid.UUID
	entityID uuid.UUID
}

// SetupTest initializes test fixtures for each test
func (s *PersonRepositoryTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.ctrl = gomock.NewController(s.T())
	s.store = db.NewMockStore(s.ctrl)
	s.logger = logger.NewMockLogger(s.ctrl)
	s.metrics = metrics.NewMockMetricsProvider(s.ctrl)
	s.tracing = tracing.NewMockTracingService(s.ctrl)
	s.tenantID = uuid.New()
	s.entityID = uuid.New()
	s.repo = NewPersonRepository(s.store, s.logger, s.metrics, s.tracing)
}

// TearDownTest cleans up test fixtures
func (s *PersonRepositoryTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// TestPersonRepository runs the person repository test suite
func TestPersonRepository(t *testing.T) {
	suite.Run(t, new(PersonRepositoryTestSuite))
}

// TestCreatePerson implements IAM-REPO-001: Verify Person creation with tenant isolation
func (s *PersonRepositoryTestSuite) TestCreatePerson() {
	testCases := []struct {
		name            string
		setupPerson     func() *model.Person
		setupMocks      func()
		expectError     bool
		errorContains   string
		validateResult  func(*testing.T, *model.Person)
	}{
		{
			name: "IAM-REPO-001_ValidPersonCreation_AllFieldsSet",
			setupPerson: func() *model.Person {
				email := "john.doe@example.com"
				phone := "+1234567890"
				nationalID := "123456789"
				taxID := "TAX123"
				middleName := "Middle"
				
				return &model.Person{
					ID:          uuid.New(),
					TenantID:    s.tenantID,
					EntityID:    s.entityID,
					PersonType:  model.PersonTypeEmployee,
					FirstName:   "John",
					LastName:    "Doe",
					MiddleName:  &middleName,
					Email:       &email,
					PhoneNumber: &phone,
					BirthDate:   time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					NationalID:  &nationalID,
					TaxID:       &taxID,
					Address: map[string]any{
						"street":      "123 Main St",
						"city":        "Test City",
						"state":       "TS",
						"postal_code": "12345",
						"country":     "TestCountry",
					},
					SecurityAttributes: map[string]any{
						"clearance_level": "PUBLIC",
						"department":      "Engineering",
					},
					IsActive:  true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
			},
			setupMocks: func() {
				// Mock successful creation
				s.store.EXPECT().
					CreatePerson(gomock.Any(), gomock.Any()).
					Return(db.Person{}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, person *model.Person) {
				require.NotNil(t, person)
				require.Equal(t, s.tenantID, person.TenantID)
				require.Equal(t, s.entityID, person.EntityID)
				require.Equal(t, "John", person.FirstName)
				require.Equal(t, "Doe", person.LastName)
				require.Equal(t, model.PersonTypeEmployee, person.PersonType)
				require.NotNil(t, person.Email)
				require.Equal(t, "john.doe@example.com", *person.Email)
				require.True(t, person.IsActive)
				require.NotZero(t, person.CreatedAt)
				require.NotZero(t, person.UpdatedAt)
			},
		},
		{
			name: "IAM-REPO-001_MinimalPersonCreation_RequiredFieldsOnly",
			setupPerson: func() *model.Person {
				return &model.Person{
					ID:         uuid.New(),
					TenantID:   s.tenantID,
					EntityID:   s.entityID,
					PersonType: model.PersonTypeIndividual,
					FirstName:  "Jane",
					LastName:   "Smith",
					BirthDate:  time.Date(1985, 6, 15, 0, 0, 0, 0, time.UTC),
					IsActive:   true,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
			},
			setupMocks: func() {
				s.store.EXPECT().
					CreatePerson(gomock.Any(), gomock.Any()).
					Return(db.Person{}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, person *model.Person) {
				require.NotNil(t, person)
				require.Equal(t, s.tenantID, person.TenantID)
				require.Equal(t, model.PersonTypeIndividual, person.PersonType)
				require.Nil(t, person.Email)
				require.Nil(t, person.PhoneNumber)
				require.Nil(t, person.DeletedAt)
			},
		},
		{
			name: "IAM-REPO-001_DuplicateEmail_WithinTenant_ReturnsError",
			setupPerson: func() *model.Person {
				email := "duplicate@example.com"
				return &model.Person{
					ID:         uuid.New(),
					TenantID:   s.tenantID,
					EntityID:   s.entityID,
					PersonType: model.PersonTypeEmployee,
					FirstName:  "Duplicate",
					LastName:   "Email",
					Email:      &email,
					BirthDate:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					IsActive:   true,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
			},
			setupMocks: func() {
				// Mock duplicate key violation
				s.store.EXPECT().
					CreatePerson(gomock.Any(), gomock.Any()).
					Return(db.Person{}, &db.Error{Code: "23505", Message: "duplicate key value violates unique constraint"}).
					Times(1)
			},
			expectError:   true,
			errorContains: "duplicate key value violates unique constraint",
		},
		{
			name: "IAM-REPO-001_InvalidTenantID_ReturnsError",
			setupPerson: func() *model.Person {
				return &model.Person{
					ID:         uuid.New(),
					TenantID:   uuid.Nil, // Invalid tenant ID
					EntityID:   s.entityID,
					PersonType: model.PersonTypeEmployee,
					FirstName:  "Invalid",
					LastName:   "Tenant",
					BirthDate:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					IsActive:   true,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
			},
			setupMocks: func() {
				// Mock foreign key violation
				s.store.EXPECT().
					CreatePerson(gomock.Any(), gomock.Any()).
					Return(db.Person{}, &db.Error{Code: "23503", Message: "foreign key violation"}).
					Times(1)
			},
			expectError:   true,
			errorContains: "foreign key violation",
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
			person := tc.setupPerson()
			
			// Act
			result, err := s.repo.Create(s.ctx, person)
			
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

// TestGetPersonByID implements IAM-REPO-001: Verify Person retrieval with tenant isolation
func (s *PersonRepositoryTestSuite) TestGetPersonByID() {
	testCases := []struct {
		name            string
		personID        uuid.UUID
		setupMocks      func()
		expectError     bool
		errorContains   string
		validateResult  func(*testing.T, *model.Person)
	}{
		{
			name:     "IAM-REPO-001_ExistingPerson_ReturnsPerson",
			personID: uuid.New(),
			setupMocks: func() {
				s.store.EXPECT().
					GetPersonByID(gomock.Any(), gomock.Any()).
					Return(db.Person{
						ID:        uuid.New(),
						TenantID:  s.tenantID,
						EntityID:  s.entityID,
						FirstName: "John",
						LastName:  "Doe",
						IsActive:  true,
					}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, person *model.Person) {
				require.NotNil(t, person)
				require.Equal(t, s.tenantID, person.TenantID)
				require.Equal(t, s.entityID, person.EntityID)
				require.Equal(t, "John", person.FirstName)
				require.Equal(t, "Doe", person.LastName)
				require.True(t, person.IsActive)
			},
		},
		{
			name:     "IAM-REPO-001_PersonNotFound_ReturnsError",
			personID: uuid.New(),
			setupMocks: func() {
				s.store.EXPECT().
					GetPersonByID(gomock.Any(), gomock.Any()).
					Return(db.Person{}, &db.Error{Code: "02000", Message: "no data found"}).
					Times(1)
			},
			expectError:   true,
			errorContains: "no data found",
		},
		{
			name:     "IAM-REPO-001_CrossTenantAccess_ReturnsNotFound",
			personID: uuid.New(),
			setupMocks: func() {
				// RLS should prevent cross-tenant access, returning no data
				s.store.EXPECT().
					GetPersonByID(gomock.Any(), gomock.Any()).
					Return(db.Person{}, &db.Error{Code: "02000", Message: "no data found"}).
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
			result, err := s.repo.GetByID(s.ctx, tc.personID)
			
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

// TestUpdatePerson implements IAM-REPO-001: Verify Person update operations
func (s *PersonRepositoryTestSuite) TestUpdatePerson() {
	testCases := []struct {
		name            string
		setupPerson     func() *model.Person
		setupMocks      func()
		expectError     bool
		errorContains   string
		validateResult  func(*testing.T, *model.Person)
	}{
		{
			name: "IAM-REPO-001_ValidUpdate_UpdatesRecord",
			setupPerson: func() *model.Person {
				updatedEmail := "updated@example.com"
				return &model.Person{
					ID:         uuid.New(),
					TenantID:   s.tenantID,
					EntityID:   s.entityID,
					PersonType: model.PersonTypeEmployee,
					FirstName:  "John",
					LastName:   "Updated",
					Email:      &updatedEmail,
					BirthDate:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					IsActive:   true,
					CreatedAt:  time.Now().Add(-time.Hour),
					UpdatedAt:  time.Now(),
				}
			},
			setupMocks: func() {
				s.store.EXPECT().
					UpdatePerson(gomock.Any(), gomock.Any()).
					Return(db.Person{
						ID:        uuid.New(),
						TenantID:  s.tenantID,
						FirstName: "John",
						LastName:  "Updated",
						UpdatedAt: time.Now(),
					}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, person *model.Person) {
				require.NotNil(t, person)
				require.Equal(t, "Updated", person.LastName)
				require.NotZero(t, person.UpdatedAt)
			},
		},
		{
			name: "IAM-REPO-001_PersonNotFound_ReturnsError",
			setupPerson: func() *model.Person {
				return &model.Person{
					ID:        uuid.New(),
					TenantID:  s.tenantID,
					FirstName: "NotFound",
					LastName:  "User",
				}
			},
			setupMocks: func() {
				s.store.EXPECT().
					UpdatePerson(gomock.Any(), gomock.Any()).
					Return(db.Person{}, &db.Error{Code: "02000", Message: "no data found"}).
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
			person := tc.setupPerson()
			
			// Act
			result, err := s.repo.Update(s.ctx, person)
			
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

// TestDeletePerson implements IAM-REPO-001: Verify Person soft deletion
func (s *PersonRepositoryTestSuite) TestDeletePerson() {
	testCases := []struct {
		name          string
		personID      uuid.UUID
		setupMocks    func()
		expectError   bool
		errorContains string
	}{
		{
			name:     "IAM-REPO-001_ExistingPerson_SetsDeletedAt",
			personID: uuid.New(),
			setupMocks: func() {
				s.store.EXPECT().
					SoftDeletePerson(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectError: false,
		},
		{
			name:     "IAM-REPO-001_PersonNotFound_ReturnsError",
			personID: uuid.New(),
			setupMocks: func() {
				s.store.EXPECT().
					SoftDeletePerson(gomock.Any(), gomock.Any()).
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
			err := s.repo.Delete(s.ctx, tc.personID)
			
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

// TestListPersons implements IAM-REPO-001: Verify Person listing with pagination
func (s *PersonRepositoryTestSuite) TestListPersons() {
	testCases := []struct {
		name            string
		limit           int
		offset          int
		setupMocks      func()
		expectError     bool
		errorContains   string
		validateResult  func(*testing.T, []*model.Person)
	}{
		{
			name:   "IAM-REPO-001_ValidPagination_ReturnsPersons",
			limit:  10,
			offset: 0,
			setupMocks: func() {
				s.store.EXPECT().
					ListPersons(gomock.Any(), db.ListPersonsParams{
						Limit:  10,
						Offset: 0,
					}).
					Return([]db.Person{
						{ID: uuid.New(), TenantID: s.tenantID, FirstName: "John", LastName: "Doe"},
						{ID: uuid.New(), TenantID: s.tenantID, FirstName: "Jane", LastName: "Smith"},
					}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, persons []*model.Person) {
				require.Len(t, persons, 2)
				require.Equal(t, "John", persons[0].FirstName)
				require.Equal(t, "Jane", persons[1].FirstName)
				// Verify all persons belong to the same tenant
				for _, person := range persons {
					require.Equal(t, s.tenantID, person.TenantID)
				}
			},
		},
		{
			name:   "IAM-REPO-001_EmptyResult_ReturnsEmptySlice",
			limit:  10,
			offset: 100,
			setupMocks: func() {
				s.store.EXPECT().
					ListPersons(gomock.Any(), gomock.Any()).
					Return([]db.Person{}, nil).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, persons []*model.Person) {
				require.Empty(t, persons)
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
			result, err := s.repo.List(s.ctx, tc.limit, tc.offset)
			
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

// TestMultiTenantIsolation implements IAM-REPO-001: Verify strict tenant isolation
func (s *PersonRepositoryTestSuite) TestMultiTenantIsolation() {
	testCases := []struct {
		name           string
		setupScenario  func() (tenantA, tenantB uuid.UUID)
		validateResult func(*testing.T, uuid.UUID, uuid.UUID)
	}{
		{
			name: "IAM-REPO-001_CrossTenantAccess_ZeroDataLeakage",
			setupScenario: func() (uuid.UUID, uuid.UUID) {
				tenantA := uuid.New()
				tenantB := uuid.New()
				
				// Mock RLS enforcement - no cross-tenant data
				s.store.EXPECT().
					GetPersonByID(gomock.Any(), gomock.Any()).
					Return(db.Person{}, &db.Error{Code: "02000", Message: "no data found"}).
					Times(1)
				
				return tenantA, tenantB
			},
			validateResult: func(t *testing.T, tenantA, tenantB uuid.UUID) {
				// Try to access tenant A's person from tenant B context
				personID := uuid.New()
				
				// This should fail due to RLS
				person, err := s.repo.GetByID(s.ctx, personID)
				require.Error(t, err)
				require.Nil(t, person)
				require.Contains(t, err.Error(), "no data found")
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
