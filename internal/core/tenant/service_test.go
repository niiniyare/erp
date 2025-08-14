package tenant

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/cache"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// ServiceTestSuite is the test suite for the tenant service
type ServiceTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	service Service
	repo    *MockRepository
	cache   *cache.MockService // Corrected mock name
	tracer  *tracing.MockTracingService
	ctx     context.Context
}

// SetupTest runs before each test in the suite
func (s *ServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.repo = NewMockRepository(s.ctrl)
	s.cache = cache.NewMockService(s.ctrl) // Corrected constructor
	s.tracer = tracing.NewMockTracingService(s.ctrl)

	// Mock the tracer to return a mock span
	mockSpan := tracing.NewMockSpan(s.ctrl)
	mockSpan.EXPECT().End(gomock.Any()).AnyTimes() // Correct End signature
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetStatus(gomock.Any(), gomock.Any()).AnyTimes()
	mockSpan.EXPECT().RecordError(gomock.Any()).AnyTimes()

	s.tracer.EXPECT().StartSpan(gomock.Any(), gomock.Any(), gomock.Any()).Return(context.Background(), mockSpan).AnyTimes()

	s.service = NewService(s.repo, s.cache, s.tracer)
	s.ctx = context.Background()
}

// TearDownTest runs after each test
func (s *ServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// TestTenantService runs the test suite
func TestTenantService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

// TestCreateTenant covers test cases MT-CORE-001, MT-CORE-002, MT-CORE-003
func (s *ServiceTestSuite) TestCreateTenant() {
	subdomain := "acme"

	testCases := []struct {
		name          string
		request       CreateTenantRequest
		mockSetup     func()
		expectError   bool
		expectedError error
		checkResult   func(t *Tenant)
	}{
		{
			name: "MT-CORE-001: Valid Tenant Creation",
			request: CreateTenantRequest{
				Name:         "ACME Corporation",
				Email:        "admin@acme.com",
				Subdomain:    &subdomain,
				CountryCode:  "US",
				CurrencyCode: "USD",
			},
			mockSetup: func() {
				s.repo.EXPECT().GetBySubdomain(s.ctx, subdomain).Return(nil, sharedErrors.ErrTenantNotFound).Times(1)
				s.repo.EXPECT().Create(s.ctx, gomock.Any()).Return(nil).Times(1)
				s.cache.EXPECT().Set(s.ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			expectError: false,
			checkResult: func(t *Tenant) {
				s.Require().NotNil(t)
				s.Require().Equal("ACME Corporation", t.Name)
				s.Require().Equal("admin@acme.com", t.Email)
				s.Require().Equal(StatusActive, t.Status)
				s.Require().NotEqual(uuid.Nil, t.ID)
			},
		},
		{
			name: "MT-CORE-002: Tenant Slug Generation",
			request: CreateTenantRequest{
				Name:         "ACME Corporation & Co.",
				Email:        "admin@acme-co.com",
				CountryCode:  "US",
				CurrencyCode: "USD",
			},
			mockSetup: func() {
				s.repo.EXPECT().Create(s.ctx, gomock.Any()).Return(nil).Times(1)
				s.cache.EXPECT().Set(s.ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			expectError: false,
			checkResult: func(t *Tenant) {
				s.Require().NotNil(t)
				s.Require().Equal("acme-corporation-and-co", t.Slug)
			},
		},
		{
			name: "Subdomain Already Exists",
			request: CreateTenantRequest{
				Name:         "Another ACME",
				Email:        "admin@another-acme.com",
				Subdomain:    &subdomain,
				CountryCode:  "US",
				CurrencyCode: "USD",
			},
			mockSetup: func() {
				s.repo.EXPECT().GetBySubdomain(s.ctx, subdomain).Return(&Tenant{ID: uuid.New(), Name: "Original ACME"}, nil).Times(1)
			},
			expectError:   true,
			expectedError: sharedErrors.ErrSubdomainAlreadyExists,
		},
		{
			name: "Name Validation (Required)",
			request: CreateTenantRequest{
				Email: "admin@no-name.com",
			},
			mockSetup:     func() {},
			expectError:   true,
			expectedError: sharedErrors.NewBusinessError("INVALID_TENANT_NAME", "Tenant name is required"),
		},
		{
			name: "Email Validation (Invalid Format)",
			request: CreateTenantRequest{
				Name:  "Valid Name",
				Email: "not-a-valid-email",
			},
			mockSetup:     func() {},
			expectError:   true,
			expectedError: sharedErrors.NewBusinessError("INVALID_EMAIL", "Invalid email format"),
		},
		{
			name: "Email Validation (Required)",
			request: CreateTenantRequest{
				Name: "Valid Name",
			},
			mockSetup:     func() {},
			expectError:   true,
			expectedError: sharedErrors.NewBusinessError("INVALID_EMAIL", "Email is required"),
		},
		{
			name: "Country Code Validation (Invalid)",
			request: CreateTenantRequest{
				Name:        "Valid Name",
				Email:       "valid@email.com",
				CountryCode: "USA",
			},
			mockSetup:     func() {},
			expectError:   true,
			expectedError: sharedErrors.NewBusinessError("INVALID_COUNTRY_CODE", "Invalid country code format"),
		},
		{
			name: "Currency Code Validation (Invalid)",
			request: CreateTenantRequest{
				Name:         "Valid Name",
				Email:        "valid@email.com",
				CountryCode:  "US",
				CurrencyCode: "USAD",
			},
			mockSetup:     func() {},
			expectError:   true,
			expectedError: sharedErrors.NewBusinessError("INVALID_CURRENCY_CODE", "Invalid currency code format"),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset controller and mocks for each run
			tc.mockSetup()

			newTenant, err := s.service.CreateTenant(s.ctx, tc.request)

			if tc.expectError {
				s.Require().Error(err)
				if tc.expectedError != nil {
					s.Require().Equal(tc.expectedError, err)
				}
			} else {
				s.Require().NoError(err)
				if tc.checkResult != nil {
					tc.checkResult(newTenant)
				}
			}
			s.TearDownTest()
		})
	}
}

func (s *ServiceTestSuite) TestUpdateTenantStatus() {
	tenantID := uuid.New()
	suspendedStatus := StatusSuspended

	testCases := []struct {
		name          string
		serviceCall   func() error
		mockSetup     func()
		expectError   bool
		expectedError error
	}{
		{
			name:        "MT-CORE-005: Deactivate an active tenant",
			serviceCall: func() error { return s.service.DeactivateTenant(s.ctx, tenantID) },
			mockSetup: func() {
				gomock.InOrder(
					s.repo.EXPECT().GetByID(s.ctx, tenantID).Return(&Tenant{ID: tenantID, Status: StatusActive}, nil),
					s.cache.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes(),
					s.repo.EXPECT().Update(s.ctx, tenantID, UpdateTenantRequest{Status: &suspendedStatus}).Return(nil),
					s.repo.EXPECT().GetByID(s.ctx, tenantID).Return(&Tenant{ID: tenantID, Status: StatusSuspended}, nil),
					s.cache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes(),
				)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.mockSetup()
			err := tc.serviceCall()

			if tc.expectError {
				s.Require().Error(err)
				if tc.expectedError != nil {
					s.Require().ErrorIs(err, tc.expectedError)
				}
			} else {
				s.Require().NoError(err)
			}
			s.TearDownTest()
		})
	}
}

func (s *ServiceTestSuite) TestSoftDeleteTenant() {
	tenantID := uuid.New()

	testCases := []struct {
		name          string
		mockSetup     func()
		expectError   bool
		expectedError error
	}{
		{
			name: "MT-CORE-006: Soft delete an existing tenant",
			mockSetup: func() {
				gomock.InOrder(
					s.repo.EXPECT().GetByID(s.ctx, tenantID).Return(&Tenant{ID: tenantID}, nil),
					s.cache.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes(),
					s.repo.EXPECT().Delete(s.ctx, tenantID).Return(nil),
				)
			},
		},
		{
			name: "DB error on soft delete",
			mockSetup: func() {
				dbError := fmt.Errorf("db error")
				gomock.InOrder(
					s.repo.EXPECT().GetByID(s.ctx, tenantID).Return(&Tenant{ID: tenantID}, nil),
					s.cache.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes(),
					s.repo.EXPECT().Delete(s.ctx, tenantID).Return(dbError),
				)
			},
			expectError:   true,
			expectedError: fmt.Errorf("failed to delete tenant: db error"),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.mockSetup()

			err := s.service.DeleteTenant(s.ctx, tenantID)

			if tc.expectError {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectedError.Error())
			} else {
				s.Require().NoError(err)
			}
			s.TearDownTest()
		})
	}
}
