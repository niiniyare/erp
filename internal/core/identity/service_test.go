//go:build unit
// +build unit

package identity

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/cache"
	shared_errors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

// IdentityServiceTestSuite is the test suite for the identity service
type IdentityServiceTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	service    Service
	mockRepo   *MockRepository
	mockCache  *cache.MockService
	mockTracer *tracing.MockTracingService
	mockMetric *metrics.MockMetricsService
}

func (s *IdentityServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = NewMockRepository(s.ctrl)
	s.mockCache = cache.NewMockService(s.ctrl)
	s.mockTracer = tracing.NewMockTracingService()
	s.mockMetric = metrics.NewMockMetricsService()

	s.service = NewService(s.mockRepo, s.mockCache, s.mockTracer, s.mockMetric)
}

func (s *IdentityServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestIdentityServiceTestSuite(t *testing.T) {
	suite.Run(t, new(IdentityServiceTestSuite))
}

func (s *IdentityServiceTestSuite) TestRegisterNewUser() {
	tests := []struct {
		name             string
		req              *CreateUserRequest
		expectedUser     *User
		expectedError    error
		mockExpectations func(repo *MockRepository, cache *cache.MockService, user *User)
	}{
		{
			name: "Success",
			req: &CreateUserRequest{
				Email:    "test.user@example.com",
				Username: "test.user",
				Password: "a-very-secure-password",
			},
			expectedUser:  &User{ID: uuid.New(), Email: "test.user@example.com", Username: "test.user"},
			expectedError: nil,
			mockExpectations: func(repo *MockRepository, cache *cache.MockService, user *User) {
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(user, nil)
				cache.EXPECT().Delete(gomock.Any(), fmt.Sprintf("user:id:%s", user.ID)).Return(nil).AnyTimes()
				cache.EXPECT().Delete(gomock.Any(), fmt.Sprintf("user:email:%s", user.Email)).Return(nil).AnyTimes()
				cache.EXPECT().Delete(gomock.Any(), fmt.Sprintf("user:username:%s", user.Username)).Return(nil).AnyTimes()
			},
		},
		{
			name: "Failure - Empty Password",
			req: &CreateUserRequest{
				Email:    "test.user@example.com",
				Username: "test.user",
				Password: "",
			},
			expectedUser:  nil,
			expectedError: fmt.Errorf("validation: password must be at least 8 characters"),
			mockExpectations: func(repo *MockRepository, cache *cache.MockService, user *User) {
				// No repository or cache calls expected
			},
		},
		{
			name: "Failure - Password Too Short",
			req: &CreateUserRequest{
				Email:    "test.user@example.com",
				Username: "test.user",
				Password: "short",
			},
			expectedUser:  nil,
			expectedError: fmt.Errorf("validation: password must be at least 8 characters"),
			mockExpectations: func(repo *MockRepository, cache *cache.MockService, user *User) {
				// No repository or cache calls expected
			},
		},
		{
			name: "Failure - CreateUser Repository Error",
			req: &CreateUserRequest{
				Email:    "test.user@example.com",
				Username: "test.user",
				Password: "a-very-secure-password",
			},
			expectedUser:  nil,
			expectedError: errors.New("repository error"),
			mockExpectations: func(repo *MockRepository, cache *cache.MockService, user *User) {
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("repository error"))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			defer s.TearDownTest()

			if tc.mockExpectations != nil {
				tc.mockExpectations(s.mockRepo, s.mockCache, tc.expectedUser)
			}

			user, err := s.service.RegisterNewUser(context.Background(), tc.req)

			if tc.expectedError != nil {
				require.Error(s.T(), err)
				require.Equal(s.T(), tc.expectedError.Error(), err.Error()) // Compare error strings for fmt.Errorf
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), user)
				require.Equal(s.T(), tc.expectedUser.Email, user.Email)
				require.Equal(s.T(), tc.expectedUser.Username, user.Username)
			}
		})
	}
}

func (s *IdentityServiceTestSuite) TestGetUserByID() {
	userID := uuid.New()
	expectedUser := &User{ID: userID, Email: "cache.miss@example.com"}

	tests := []struct {
		name             string
		userID           uuid.UUID
		expectedUser     *User
		expectedError    error
		mockExpectations func(repo *MockRepository, cache *cache.MockService, userID uuid.UUID, user *User)
	}{
		{
			name:          "Success - Cache Miss",
			userID:        userID,
			expectedUser:  expectedUser,
			expectedError: nil,
			mockExpectations: func(repo *MockRepository, cache *cache.MockService, userID uuid.UUID, user *User) {
				cache.EXPECT().Get(gomock.Any(), fmt.Sprintf("user:id:%s", userID), gomock.Any()).Return(errors.New("cache miss"))
				repo.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil)
				cache.EXPECT().Set(gomock.Any(), fmt.Sprintf("user:id:%s", user.ID), user, 30*time.Minute).Return(nil)
				cache.EXPECT().Set(gomock.Any(), fmt.Sprintf("user:email:%s", user.Email), user, 30*time.Minute).Return(nil)
			},
		},
		{
			name:          "Success - Cache Hit",
			userID:        userID,
			expectedUser:  expectedUser,
			expectedError: nil,
			mockExpectations: func(repo *MockRepository, cache *cache.MockService, userID uuid.UUID, user *User) {
				cache.EXPECT().Get(gomock.Any(), fmt.Sprintf("user:id:%s", userID), gomock.Any()).DoAndReturn(func(_ context.Context, _ string, dest interface{}) error {
					destPtr := dest.(*User)
					*destPtr = *user
					return nil
				})
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset mocks for each sub-test
			defer s.TearDownTest()

			if tc.mockExpectations != nil {
				tc.mockExpectations(s.mockRepo, s.mockCache, tc.userID, tc.expectedUser)
			}

			user, err := s.service.GetUserByID(context.Background(), tc.userID)

			if tc.expectedError != nil {
				require.Error(s.T(), err)
				require.Equal(s.T(), tc.expectedError, err)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), user)
				require.Equal(s.T(), tc.expectedUser, user)
			}
		})
	}
}

func (s *IdentityServiceTestSuite) TestGetUserByEmail() {
	tests := []struct {
		name             string
		email            string
		expectedUser     *User
		expectedError    error
		mockExpectations func(repo *MockRepository, cache *cache.MockService, email string, user *User)
	}{
		{
			name:          "Success - Cache Miss",
			email:         "email.miss@example.com",
			expectedUser:  &User{ID: uuid.New(), Email: "email.miss@example.com"},
			expectedError: nil,
			mockExpectations: func(repo *MockRepository, cache *cache.MockService, email string, user *User) {
				cache.EXPECT().Get(gomock.Any(), fmt.Sprintf("user:email:%s", email), gomock.Any()).Return(errors.New("cache miss"))
				repo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(user, nil)
				cache.EXPECT().Set(gomock.Any(), fmt.Sprintf("user:email:%s", email), user, 30*time.Minute).Return(nil)
				cache.EXPECT().Set(gomock.Any(), fmt.Sprintf("user:id:%s", user.ID), user, 30*time.Minute).Return(nil)
			},
		},
		{
			name:          "Success - Cache Hit",
			email:         "email.hit@example.com",
			expectedUser:  &User{ID: uuid.New(), Email: "email.hit@example.com"},
			expectedError: nil,
			mockExpectations: func(repo *MockRepository, cache *cache.MockService, email string, user *User) {
				cache.EXPECT().Get(gomock.Any(), fmt.Sprintf("user:email:%s", email), gomock.Any()).DoAndReturn(func(_ context.Context, _ string, dest interface{}) error {
					destPtr := dest.(*User)
					*destPtr = *user
					return nil
				})
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset mocks for each sub-test
			defer s.TearDownTest()

			if tc.mockExpectations != nil {
				tc.mockExpectations(s.mockRepo, s.mockCache, tc.email, tc.expectedUser)
			}

			user, err := s.service.GetUserByEmail(context.Background(), tc.email)

			if tc.expectedError != nil {
				require.Error(s.T(), err)
				require.Equal(s.T(), tc.expectedError, err)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), user)
				require.Equal(s.T(), tc.expectedUser, user)
			}
		})
	}
}

func (s *IdentityServiceTestSuite) TestGetUserByUsername() {
	tests := []struct {
		name             string
		username         string
		expectedUser     *User
		expectedError    error
		mockExpectations func(repo *MockRepository, cache *cache.MockService, username string, user *User)
	}{
		{
			name:          "Success - Cache Miss",
			username:      "username.miss",
			expectedUser:  &User{ID: uuid.New(), Username: "username.miss", Email: "user@name.com"},
			expectedError: nil,
			mockExpectations: func(repo *MockRepository, cache *cache.MockService, username string, user *User) {
				cache.EXPECT().Get(gomock.Any(), fmt.Sprintf("user:username:%s", username), gomock.Any()).Return(errors.New("cache miss"))
				repo.EXPECT().GetUserByUsername(gomock.Any(), username).Return(user, nil)
				cache.EXPECT().Set(gomock.Any(), fmt.Sprintf("user:username:%s", username), user, 30*time.Minute).Return(nil)
				cache.EXPECT().Set(gomock.Any(), fmt.Sprintf("user:id:%s", user.ID), user, 30*time.Minute).Return(nil)
				cache.EXPECT().Set(gomock.Any(), fmt.Sprintf("user:email:%s", user.Email), user, 30*time.Minute).Return(nil)
			},
		},
		{
			name:          "Success - Cache Hit",
			username:      "username.hit",
			expectedUser:  &User{ID: uuid.New(), Username: "username.hit", Email: "user@name.com"},
			expectedError: nil,
			mockExpectations: func(repo *MockRepository, cache *cache.MockService, username string, user *User) {
				cache.EXPECT().Get(gomock.Any(), fmt.Sprintf("user:username:%s", username), gomock.Any()).DoAndReturn(func(_ context.Context, _ string, dest interface{}) error {
					destPtr := dest.(*User)
					*destPtr = *user
					return nil
				})
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset mocks for each sub-test
			defer s.TearDownTest()

			if tc.mockExpectations != nil {
				tc.mockExpectations(s.mockRepo, s.mockCache, tc.username, tc.expectedUser)
			}

			user, err := s.service.GetUserByUsername(context.Background(), tc.username)

			if tc.expectedError != nil {
				require.Error(s.T(), err)
				require.Equal(s.T(), tc.expectedError, err)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), user)
				require.Equal(s.T(), tc.expectedUser, user)
			}
		})
	}
}

func (s *IdentityServiceTestSuite) TestUpdateUser() {
	userID := uuid.New()
	originalUser := &User{ID: userID, Email: "original@example.com", Username: "originaluser"}
	updatedEmail := "updated@example.com"
	updatedUsername := "updateduser"
	req := &UpdateUserRequest{Email: &updatedEmail, Username: &updatedUsername}
	expectedUser := &User{ID: userID, Email: updatedEmail, Username: updatedUsername}

	tests := []struct {
		name             string
		userID           uuid.UUID
		req              *UpdateUserRequest
		originalUser     *User
		expectedUser     *User
		expectedError    error
		mockExpectations func(repo *MockRepository, cache *cache.MockService, userID uuid.UUID, originalUser, updatedUser *User, req *UpdateUserRequest)
	}{
		{
			name:          "Success",
			userID:        userID,
			req:           req,
			originalUser:  originalUser,
			expectedUser:  expectedUser,
			expectedError: nil,
			mockExpectations: func(repo *MockRepository, cache *cache.MockService, userID uuid.UUID, originalUser, updatedUser *User, req *UpdateUserRequest) {
				repo.EXPECT().GetUserByID(gomock.Any(), userID).Return(originalUser, nil)
				repo.EXPECT().UpdateUser(gomock.Any(), userID, req).Return(updatedUser, nil)
				cache.EXPECT().Delete(gomock.Any(), fmt.Sprintf("user:id:%s", originalUser.ID)).Return(nil).AnyTimes()
				cache.EXPECT().Delete(gomock.Any(), fmt.Sprintf("user:email:%s", originalUser.Email)).Return(nil).AnyTimes()
				cache.EXPECT().Delete(gomock.Any(), fmt.Sprintf("user:username:%s", originalUser.Username)).Return(nil).AnyTimes()
				cache.EXPECT().Set(gomock.Any(), fmt.Sprintf("user:id:%s", updatedUser.ID), updatedUser, 30*time.Minute).Return(nil).AnyTimes()
				cache.EXPECT().Set(gomock.Any(), fmt.Sprintf("user:email:%s", updatedUser.Email), updatedUser, 30*time.Minute).Return(nil).AnyTimes()
				cache.EXPECT().Set(gomock.Any(), fmt.Sprintf("user:username:%s", updatedUser.Username), updatedUser, 30*time.Minute).Return(nil).AnyTimes()
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset mocks for each sub-test
			defer s.TearDownTest()

			if tc.mockExpectations != nil {
				tc.mockExpectations(s.mockRepo, s.mockCache, tc.userID, tc.originalUser, tc.expectedUser, tc.req)
			}

			user, err := s.service.UpdateUser(context.Background(), tc.userID, tc.req)

			if tc.expectedError != nil {
				require.Error(s.T(), err)
				require.Equal(s.T(), tc.expectedError, err)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), user)
				require.Equal(s.T(), tc.expectedUser.Email, user.Email)
				require.Equal(s.T(), tc.expectedUser.Username, user.Username)
			}
		})
	}
}

func (s *IdentityServiceTestSuite) TestAuthenticate() {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	user := &User{ID: uuid.New(), Email: "auth@example.com", Username: "authuser"}

	tests := []struct {
		name             string
		identifier       string
		password         string
		hashedPassword   string
		expectedUser     *User
		expectedError    error
		mockExpectations func(repo *MockRepository, identifier string, user *User, hashedPassword string)
	}{
		{
			name:           "Success - By Email",
			identifier:     "auth@example.com",
			password:       "correctpassword",
			hashedPassword: string(hashedPassword),
			expectedUser:   user,
			expectedError:  nil,
			mockExpectations: func(repo *MockRepository, identifier string, user *User, hashedPassword string) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), identifier).Return(user, nil)
				repo.EXPECT().GetUserPassword(gomock.Any(), user.ID).Return(hashedPassword, nil)
			},
		},
		{
			name:           "Success - By Username",
			identifier:     "authuser",
			password:       "correctpassword",
			hashedPassword: string(hashedPassword),
			expectedUser:   user,
			expectedError:  nil,
			mockExpectations: func(repo *MockRepository, identifier string, user *User, hashedPassword string) {
				repo.EXPECT().GetUserByUsername(gomock.Any(), identifier).Return(user, nil)
				repo.EXPECT().GetUserPassword(gomock.Any(), user.ID).Return(hashedPassword, nil)
			},
		},
		{
			name:           "Failure - Incorrect Password",
			identifier:     "auth@example.com",
			password:       "wrongpassword",
			hashedPassword: string(hashedPassword),
			expectedUser:   user,
			expectedError:  shared_errors.ErrAuthenticationFailed,
			mockExpectations: func(repo *MockRepository, identifier string, user *User, hashedPassword string) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), identifier).Return(user, nil)
				repo.EXPECT().GetUserPassword(gomock.Any(), user.ID).Return(hashedPassword, nil)
			},
		},
		{
			name:           "Failure - User Not Found By Email",
			identifier:     "nonexistent@example.com",
			password:       "anypassword",
			hashedPassword: "",
			expectedUser:   nil,
			expectedError:  shared_errors.ErrAuthenticationFailed,
			mockExpectations: func(repo *MockRepository, identifier string, user *User, hashedPassword string) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), identifier).Return(nil, errors.New("not found"))
			},
		},
		{
			name:           "Failure - User Not Found By Username",
			identifier:     "nonexistentuser",
			password:       "anypassword",
			hashedPassword: "",
			expectedUser:   nil,
			expectedError:  shared_errors.ErrAuthenticationFailed,
			mockExpectations: func(repo *MockRepository, identifier string, user *User, hashedPassword string) {
				repo.EXPECT().GetUserByUsername(gomock.Any(), identifier).Return(nil, errors.New("not found"))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset mocks for each sub-test
			defer s.TearDownTest()

			if tc.mockExpectations != nil {
				tc.mockExpectations(s.mockRepo, tc.identifier, tc.expectedUser, tc.hashedPassword)
			}

			user, err := s.service.Authenticate(context.Background(), tc.identifier, tc.password)

			require.ErrorIs(s.T(), err, tc.expectedError)
			if tc.expectedError == nil {
				require.NotNil(s.T(), user)
				require.Equal(s.T(), tc.expectedUser, user)
			} else {
				require.Nil(s.T(), user)
			}
		})
	}
}

func (s *IdentityServiceTestSuite) TestChangePassword() {
	userID := uuid.New()
	oldHashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
	wrongHashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	tests := []struct {
		name              string
		userID            uuid.UUID
		oldPassword       string
		newPassword       string
		oldHashedPassword string
		expectedError     error
		mockExpectations  func(repo *MockRepository, userID uuid.UUID, oldHashedPassword string, newPassword string)
	}{
		{
			name:              "Success",
			userID:            userID,
			oldPassword:       "oldpassword",
			newPassword:       "newpassword",
			oldHashedPassword: string(oldHashedPassword),
			expectedError:     nil,
			mockExpectations: func(repo *MockRepository, userID uuid.UUID, oldHashedPassword string, newPassword string) {
				repo.EXPECT().GetUserPassword(gomock.Any(), userID).Return(oldHashedPassword, nil)
				repo.EXPECT().UpdatePassword(gomock.Any(), userID, gomock.Any()).Return(nil)
			},
		},
		{
			name:              "Failure - Incorrect Old Password",
			userID:            userID,
			oldPassword:       "wrongpassword",
			newPassword:       "newpassword",
			oldHashedPassword: string(wrongHashedPassword),
			expectedError:     shared_errors.ErrAuthenticationFailed,
			mockExpectations: func(repo *MockRepository, userID uuid.UUID, oldHashedPassword string, newPassword string) {
				repo.EXPECT().GetUserPassword(gomock.Any(), userID).Return(oldHashedPassword, nil)
			},
		},
		{
			name:              "Failure - GetUserPassword Error",
			userID:            userID,
			oldPassword:       "oldpassword",
			newPassword:       "newpassword",
			oldHashedPassword: "",
			expectedError:     shared_errors.ErrAuthenticationFailed,
			mockExpectations: func(repo *MockRepository, userID uuid.UUID, oldHashedPassword string, newPassword string) {
				repo.EXPECT().GetUserPassword(gomock.Any(), userID).Return("", errors.New("db error"))
			},
		},
		{
			name:              "Failure - UpdatePassword Error",
			userID:            userID,
			oldPassword:       "oldpassword",
			newPassword:       "newpassword",
			oldHashedPassword: string(oldHashedPassword),
			expectedError:     errors.New("db update error"),
			mockExpectations: func(repo *MockRepository, userID uuid.UUID, oldHashedPassword string, newPassword string) {
				repo.EXPECT().GetUserPassword(gomock.Any(), userID).Return(oldHashedPassword, nil)
				repo.EXPECT().UpdatePassword(gomock.Any(), userID, gomock.Any()).Return(errors.New("db update error"))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset mocks for each sub-test
			defer s.TearDownTest()

			if tc.mockExpectations != nil {
				tc.mockExpectations(s.mockRepo, tc.userID, tc.oldHashedPassword, tc.newPassword)
			}

			err := s.service.ChangePassword(context.Background(), tc.userID, tc.oldPassword, tc.newPassword)

			if tc.expectedError != nil {
				require.Error(s.T(), err)
				require.Equal(s.T(), tc.expectedError, err)
			} else {
				require.NoError(s.T(), err)
			}
		})
	}
}

func (s *IdentityServiceTestSuite) TestAssignUserRole() {
	tests := []struct {
		name             string
		userID           uuid.UUID
		roleID           uuid.UUID
		entityID         uuid.UUID
		expectedError    error
		mockExpectations func(repo *MockRepository, userID, roleID, entityID uuid.UUID)
	}{
		{
			name:          "Success",
			userID:        uuid.New(),
			roleID:        uuid.New(),
			entityID:      uuid.New(),
			expectedError: nil,
			mockExpectations: func(repo *MockRepository, userID, roleID, entityID uuid.UUID) {
				// repo.EXPECT().AssignUserRole(gomock.Any(), userID, roleID, entityID).Return(nil)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset mocks for each sub-test
			defer s.TearDownTest()

			if tc.mockExpectations != nil {
				tc.mockExpectations(s.mockRepo, tc.userID, tc.roleID, tc.entityID)
			}

			err := s.service.AssignUserRole(context.Background(), tc.userID, tc.roleID, tc.entityID)

			if tc.expectedError != nil {
				require.Error(s.T(), err)
				require.Equal(s.T(), tc.expectedError, err)
			} else {
				require.NoError(s.T(), err)
			}
		})
	}
}

func (s *IdentityServiceTestSuite) TestRevokeUserRole() {
	tests := []struct {
		name             string
		userID           uuid.UUID
		roleID           uuid.UUID
		entityID         uuid.UUID
		expectedError    error
		mockExpectations func(repo *MockRepository, userID, roleID, entityID uuid.UUID)
	}{
		{
			name:          "Success",
			userID:        uuid.New(),
			roleID:        uuid.New(),
			entityID:      uuid.New(),
			expectedError: nil,
			mockExpectations: func(repo *MockRepository, userID, roleID, entityID uuid.UUID) {
				// repo.EXPECT().RevokeUserRole(gomock.Any(), userID, roleID, entityID).Return(nil)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset mocks for each sub-test
			defer s.TearDownTest()

			if tc.mockExpectations != nil {
				tc.mockExpectations(s.mockRepo, tc.userID, tc.roleID, tc.entityID)
			}

			err := s.service.RevokeUserRole(context.Background(), tc.userID, tc.roleID, tc.entityID)

			if tc.expectedError != nil {
				require.Error(s.T(), err)
				require.Equal(s.T(), tc.expectedError, err)
			} else {
				require.NoError(s.T(), err)
			}
		})
	}
}

func (s *IdentityServiceTestSuite) TestCreatePerson() {
	tests := []struct {
		name             string
		req              *CreatePersonRequest
		expectedPerson   *Person
		expectedError    error
		mockExpectations func(repo *MockRepository, person *Person, req *CreatePersonRequest)
	}{
		{
			name: "Success",
			req: &CreatePersonRequest{
				FirstName: "John",
				LastName:  "Doe",
			},
			expectedPerson: &Person{ID: uuid.New(), FirstName: "John", LastName: "Doe"},
			expectedError:  nil,
			mockExpectations: func(repo *MockRepository, person *Person, req *CreatePersonRequest) {
				repo.EXPECT().CreatePerson(gomock.Any(), req).Return(person, nil)
			},
		},
		{
			name: "Failure - Repo Error",
			req: &CreatePersonRequest{
				FirstName: "Jane",
				LastName:  "Doe",
			},
			expectedPerson: nil,
			expectedError:  errors.New("repository error"),
			mockExpectations: func(repo *MockRepository, person *Person, req *CreatePersonRequest) {
				repo.EXPECT().CreatePerson(gomock.Any(), req).Return(nil, errors.New("repository error"))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset mocks for each sub-test
			defer s.TearDownTest()

			if tc.mockExpectations != nil {
				tc.mockExpectations(s.mockRepo, tc.expectedPerson, tc.req)
			}

			person, err := s.service.CreatePerson(context.Background(), tc.req)

			if tc.expectedError != nil {
				require.Error(s.T(), err)
				require.Equal(s.T(), tc.expectedError, err)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), person)
				require.Equal(s.T(), tc.expectedPerson, person)
			}
		})
	}
}

func (s *IdentityServiceTestSuite) TestGetPersonByID() {
	tests := []struct {
		name             string
		personID         uuid.UUID
		expectedPerson   *Person
		expectedError    error
		mockExpectations func(repo *MockRepository, personID uuid.UUID, person *Person)
	}{
		{
			name:           "Success",
			personID:       uuid.New(),
			expectedPerson: &Person{ID: uuid.New(), FirstName: "Jane", LastName: "Doe"},
			expectedError:  nil,
			mockExpectations: func(repo *MockRepository, personID uuid.UUID, person *Person) {
				repo.EXPECT().GetPersonByID(gomock.Any(), personID).Return(person, nil)
			},
		},
		{
			name:           "Failure - Repo Error",
			personID:       uuid.New(),
			expectedPerson: nil,
			expectedError:  errors.New("repository error"),
			mockExpectations: func(repo *MockRepository, personID uuid.UUID, person *Person) {
				repo.EXPECT().GetPersonByID(gomock.Any(), personID).Return(nil, errors.New("repository error"))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset mocks for each sub-test
			defer s.TearDownTest()

			if tc.mockExpectations != nil {
				tc.mockExpectations(s.mockRepo, tc.personID, tc.expectedPerson)
			}

			person, err := s.service.GetPersonByID(context.Background(), tc.personID)

			if tc.expectedError != nil {
				require.Error(s.T(), err)
				require.Equal(s.T(), tc.expectedError, err)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), person)
				require.Equal(s.T(), tc.expectedPerson, person)
			}
		})
	}
}

func (s *IdentityServiceTestSuite) TestCreateEmployee() {
	tests := []struct {
		name             string
		req              *CreateEmployeeRequest
		expectedEmployee *Employee
		expectedError    error
		mockExpectations func(repo *MockRepository, employee *Employee, req *CreateEmployeeRequest)
	}{
		{
			name: "Success",
			req: &CreateEmployeeRequest{
				PersonID:       uuid.New(),
				EmployeeNumber: "EMP001",
			},
			expectedEmployee: &Employee{ID: uuid.New(), PersonID: uuid.New(), EmployeeNumber: "EMP001"},
			expectedError:    nil,
			mockExpectations: func(repo *MockRepository, employee *Employee, req *CreateEmployeeRequest) {
				repo.EXPECT().CreateEmployee(gomock.Any(), req).Return(employee, nil)
			},
		},
		{
			name: "Failure - Repo Error",
			req: &CreateEmployeeRequest{
				PersonID:       uuid.New(),
				EmployeeNumber: "EMP002",
			},
			expectedEmployee: nil,
			expectedError:    errors.New("repository error"),
			mockExpectations: func(repo *MockRepository, employee *Employee, req *CreateEmployeeRequest) {
				repo.EXPECT().CreateEmployee(gomock.Any(), req).Return(nil, errors.New("repository error"))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset mocks for each sub-test
			defer s.TearDownTest()

			if tc.mockExpectations != nil {
				tc.mockExpectations(s.mockRepo, tc.expectedEmployee, tc.req)
			}

			employee, err := s.service.CreateEmployee(context.Background(), tc.req)

			if tc.expectedError != nil {
				require.Error(s.T(), err)
				require.Equal(s.T(), tc.expectedError, err)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), employee)
				require.Equal(s.T(), tc.expectedEmployee, employee)
			}
		})
	}
}

func (s *IdentityServiceTestSuite) TestGetEmployeeByID() {
	tests := []struct {
		name             string
		employeeID       uuid.UUID
		expectedEmployee *Employee
		expectedError    error
		mockExpectations func(repo *MockRepository, employeeID uuid.UUID, employee *Employee)
	}{
		{
			name:             "Success",
			employeeID:       uuid.New(),
			expectedEmployee: &Employee{ID: uuid.New(), PersonID: uuid.New(), EmployeeNumber: "EMP003"},
			expectedError:    nil,
			mockExpectations: func(repo *MockRepository, employeeID uuid.UUID, employee *Employee) {
				repo.EXPECT().GetEmployeeByID(gomock.Any(), employeeID).Return(employee, nil)
			},
		},
		{
			name:             "Failure - Repo Error",
			employeeID:       uuid.New(),
			expectedEmployee: nil,
			expectedError:    errors.New("repository error"),
			mockExpectations: func(repo *MockRepository, employeeID uuid.UUID, employee *Employee) {
				repo.EXPECT().GetEmployeeByID(gomock.Any(), employeeID).Return(nil, errors.New("repository error"))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset mocks for each sub-test
			defer s.TearDownTest()

			if tc.mockExpectations != nil {
				tc.mockExpectations(s.mockRepo, tc.employeeID, tc.expectedEmployee)
			}

			employee, err := s.service.GetEmployeeByID(context.Background(), tc.employeeID)

			if tc.expectedError != nil {
				require.Error(s.T(), err)
				require.Equal(s.T(), tc.expectedError, err)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), employee)
				require.Equal(s.T(), tc.expectedEmployee, employee)
			}
		})
	}
}
