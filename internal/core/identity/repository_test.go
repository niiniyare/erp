//go:build unit
// +build unit

package identity

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// RepositoryTestSuite is the test suite for the identity repository
type RepositoryTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	repo       Repository
	mockStore  *db.MockStore
	mockTracer *tracing.MockService
	mockMetric *metrics.MockMetricsService
}

func (s *RepositoryTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockStore = db.NewMockStore(s.ctrl)
	s.mockTracer = tracing.NewMockService()
	s.mockMetric = metrics.NewMockMetricsService()

	s.repo = NewRepository(s.mockStore, s.mockTracer, s.mockMetric)
}

func (s *RepositoryTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func (s *RepositoryTestSuite) TestCreateUser() {
	// Define test cases
	tests := []struct {
		name             string
		req              *CreateUserRequest
		hashedPassword   string
		expectedUser     *User
		expectedError    error
		mockExpectations func(store *db.MockStore, req *CreateUserRequest, hashedPassword string, expectedUser *User)
	}{
		{
			name: "Success",
			req: &CreateUserRequest{
				Email:    "test@example.com",
				Username: "testuser",
				UserType: "ADMIN",
				EntityID: uuid.New(),
			},
			hashedPassword: "hashedpassword",
			expectedUser: &User{
				Email:    "test@example.com",
				Username: "testuser",
				UserType: "ADMIN",
			},
			expectedError: nil,
			mockExpectations: func(store *db.MockStore, req *CreateUserRequest, hashedPassword string, expectedUser *User) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, arg db.CreateUserParams) (*db.User, error) {
					// Simulate the database returning a user with a generated ID
					generatedID := uuid.New()
					return &db.User{
						ID:           generatedID,
						Email:        arg.Email,
						Username:     arg.Username,
						UserType:     arg.UserType,
						EntityID:     arg.EntityID,
						PasswordHash: arg.PasswordHash,
					}, nil
				})
			},
		},
		{
			name: "Store Error",
			req: &CreateUserRequest{
				Email:    "error@example.com",
				Username: "erroruser",
				UserType: "ADMIN",
				EntityID: uuid.New(),
			},
			hashedPassword: "hashedpassword",
			expectedUser:   nil,
			expectedError:  errors.New("database error"),
			mockExpectations: func(store *db.MockStore, req *CreateUserRequest, hashedPassword string, expectedUser *User) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil, errors.New("database error"))
			},
		},
	}

	// Run test cases
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset mocks for each sub-test
			defer s.TearDownTest()

			if tc.mockExpectations != nil {
				tc.mockExpectations(s.mockStore, tc.req, tc.hashedPassword, tc.expectedUser)
			}

			user, err := s.repo.CreateUser(context.Background(), tc.req, tc.hashedPassword)

			if tc.expectedError != nil {
				require.Error(s.T(), err)
				require.Contains(s.T(), err.Error(), tc.expectedError.Error())
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), user)
				require.Equal(s.T(), tc.expectedUser.Email, user.Email)
				require.Equal(s.T(), tc.expectedUser.Username, user.Username)
				require.Equal(s.T(), tc.expectedUser.UserType, user.UserType)
				require.NotNil(s.T(), user.ID)
				require.NotNil(s.T(), user.EntityID)
			}
		})
	}
}
