package identity

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"golang.org/x/crypto/bcrypt"
)

// Service defines the core business logic for the identity domain.
// It orchestrates the repository and other services to fulfill use cases.
type Service interface {
	RegisterNewUser(ctx context.Context, req *CreateUserRequest) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*User, error)

	Authenticate(ctx context.Context, identifier, password string) (*User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error

	CreatePerson(ctx context.Context, req *CreatePersonRequest) (*Person, error)
	GetPersonByID(ctx context.Context, id uuid.UUID) (*Person, error)

	CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*Employee, error)
	GetEmployeeByID(ctx context.Context, id uuid.UUID) (*Employee, error)

	AssignUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error
	RevokeUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error

	// Additional methods needed by handlers
	ListUsers(ctx context.Context, req *ListUsersRequest) ([]*User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error)
	SearchUsers(ctx context.Context, query string, limit int, offset int) ([]*User, error)
}

// service implements the Service interface
type service struct {
	repo    Repository
	cache   cache.Service
	tracing tracing.TracingService
	metrics metrics.MetricsProvider
}

// NewService creates a new identity service
func NewService(repo Repository, cache cache.Service, tracing tracing.TracingService, metrics metrics.MetricsProvider) Service {
	return &service{
		repo:    repo,
		cache:   cache,
		tracing: tracing,
		metrics: metrics,
	}
}

// RegisterNewUser handles the business logic of creating a new user.
func (s *service) RegisterNewUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "identity.service.RegisterNewUser")
	defer span.End()

	// Basic validation
	if req.Password == "" || len(req.Password) < 8 {
		return nil, fmt.Errorf("validation: password must be at least 8 characters")
	}
	// In a real app, use a proper validation library

	// Hash the password before it goes any further
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash_password: failed to hash password for new user: %w", err)
	}

	// Call repository to create user
	user, err := s.repo.CreateUser(ctx, req, hashedPassword)
	if err != nil {
		// TODO: Check for specific repository errors (e.g., email exists) and return appropriate domain errors.
		return nil, err
	}

	// Invalidate cache for safety, though not strictly necessary on create
	s.invalidateUserCache(ctx, user.ID, user.Email, user.Username)

	return user, nil
}

// GetUserByID retrieves a user by ID, with caching.
func (s *service) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "identity.service.GetUserByID")
	defer span.End()

	// Check cache first
	cacheKey := fmt.Sprintf("user:id:%s", id)
	var user User
	if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil // Cache hit
	}

	// Cache miss, get from repository
	dbUser, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.cacheUser(ctx, dbUser)
	return dbUser, nil
}

// GetUserByEmail retrieves a user by email, with caching.
func (s *service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "identity.service.GetUserByEmail")
	defer span.End()

	cacheKey := fmt.Sprintf("user:email:%s", email)
	var user User
	if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil
	}

	dbUser, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	s.cacheUser(ctx, dbUser)
	return dbUser, nil
}

// GetUserByUsername retrieves a user by username, with caching.
func (s *service) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "identity.service.GetUserByUsername")
	defer span.End()

	cacheKey := fmt.Sprintf("user:username:%s", username)
	var user User
	if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil
	}

	dbUser, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	s.cacheUser(ctx, dbUser)
	return dbUser, nil
}

// UpdateUser updates a user's details.
func (s *service) UpdateUser(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "identity.service.UpdateUser")
	defer span.End()

	// Get the original user to invalidate cache correctly
	originalUser, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updatedUser, err := s.repo.UpdateUser(ctx, id, req)
	if err != nil {
		return nil, err
	}

	// Invalidate all possible cache entries for the old and new state
	s.invalidateUserCache(ctx, originalUser.ID, originalUser.Email, originalUser.Username)
	s.cacheUser(ctx, updatedUser)

	return updatedUser, nil
}

// Authenticate handles user login.
func (s *service) Authenticate(ctx context.Context, identifier, password string) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "identity.service.Authenticate")
	defer span.End()

	var user *User
	var err error

	// Try to find user by email first, then by username
	if strings.Contains(identifier, "@") {
		user, err = s.repo.GetUserByEmail(ctx, identifier)
	} else {
		user, err = s.repo.GetUserByUsername(ctx, identifier)
	}

	if err != nil {
		return nil, errors.ErrAuthenticationFailed
	}

	// Get the stored password hash
	hashedPassword, err := s.repo.GetUserPassword(ctx, user.ID)
	if err != nil {
		return nil, errors.ErrAuthenticationFailed
	}

	// Verify the password
	if !s.verifyPassword(hashedPassword, password) {
		// TODO: Increment failed login attempts
		return nil, errors.ErrAuthenticationFailed
	}

	// TODO: Update last login time
	return user, nil
}

// ChangePassword handles changing a user's password.
func (s *service) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	ctx, span := s.tracing.StartSpan(ctx, "identity.service.ChangePassword")
	defer span.End()

	// Verify old password
	currentHashedPassword, err := s.repo.GetUserPassword(ctx, userID)
	if err != nil {
		return errors.ErrAuthenticationFailed
	}
	if !s.verifyPassword(currentHashedPassword, oldPassword) {
		return errors.ErrAuthenticationFailed
	}

	// Hash new password
	newHashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash_password: failed to hash new password: %w", err)
	}

	// Update in repository
	return s.repo.UpdatePassword(ctx, userID, newHashedPassword)
}

// --- Person and Employee methods are pass-through for now ---

func (s *service) CreatePerson(ctx context.Context, req *CreatePersonRequest) (*Person, error) {
	return s.repo.CreatePerson(ctx, req)
}

func (s *service) GetPersonByID(ctx context.Context, id uuid.UUID) (*Person, error) {
	return s.repo.GetPersonByID(ctx, id)
}

func (s *service) CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*Employee, error) {
	return s.repo.CreateEmployee(ctx, req)
}

func (s *service) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*Employee, error) {
	return s.repo.GetEmployeeByID(ctx, id)
}

func (s *service) AssignUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	return s.repo.AssignUserRole(ctx, userID, roleID, entityID)
}

func (s *service) RevokeUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	return s.repo.RevokeUserRole(ctx, userID, roleID, entityID)
}

// --- Additional Methods ---

func (s *service) ListUsers(ctx context.Context, req *ListUsersRequest) ([]*User, error) {
	// TODO: Implement user listing with pagination
	return nil, fmt.Errorf("not implemented")
}

func (s *service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// TODO: Implement user deletion
	return fmt.Errorf("not implemented")
}

func (s *service) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error) {
	// TODO: Implement role retrieval
	return nil, fmt.Errorf("not implemented")
}

func (s *service) SearchUsers(ctx context.Context, query string, limit int, offset int) ([]*User, error) {
	// TODO: Implement user search with pagination
	return nil, fmt.Errorf("not implemented")
}


// --- Private Helpers ---

func (s *service) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (s *service) verifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *service) cacheUser(ctx context.Context, user *User) {
	ttl := 30 * time.Minute
	s.cache.Set(ctx, fmt.Sprintf("user:id:%s", user.ID), user, ttl)
	s.cache.Set(ctx, fmt.Sprintf("user:email:%s", user.Email), user, ttl)
	if user.Username != "" {
		s.cache.Set(ctx, fmt.Sprintf("user:username:%s", user.Username), user, ttl)
	}
}

func (s *service) invalidateUserCache(ctx context.Context, id uuid.UUID, email, username string) {
	s.cache.Delete(ctx, fmt.Sprintf("user:id:%s", id))
	s.cache.Delete(ctx, fmt.Sprintf("user:email:%s", email))
	if username != "" {
		s.cache.Delete(ctx, fmt.Sprintf("user:username:%s", username))
	}
}
