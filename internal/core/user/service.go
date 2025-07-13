package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Service defines the user service interface
type Service interface {
	// User CRUD operations
	CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id uuid.UUID, permanent bool) error
	ListUsers(ctx context.Context, req *ListUsersRequest) ([]*User, error)
	GetUserWithDetails(ctx context.Context, id uuid.UUID) (*UserWithDetails, error)
	
	// Person operations
	CreatePerson(ctx context.Context, req *CreatePersonRequest) (*Person, error)
	GetPersonByID(ctx context.Context, id uuid.UUID) (*Person, error)
	UpdatePerson(ctx context.Context, id uuid.UUID, req *CreatePersonRequest) (*Person, error)
	
	// Employee operations
	CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*Employee, error)
	GetEmployeeByID(ctx context.Context, id uuid.UUID) (*Employee, error)
	UpdateEmployee(ctx context.Context, id uuid.UUID, req *CreateEmployeeRequest) (*Employee, error)
	
	// Authentication operations
	AuthenticateUser(ctx context.Context, identifier, password string) (*User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, newPassword string) error
	
	// Role operations
	AssignUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error
	RevokeUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*UserRole, error)
	
	// Validation operations
	ValidateUserEmail(ctx context.Context, email string, excludeID *uuid.UUID) error
	ValidateUsername(ctx context.Context, username string, excludeID *uuid.UUID) error
	ValidateEmployeeNumber(ctx context.Context, number string, excludeID *uuid.UUID) error
	
	// Search operations
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*User, error)
}

// service implements the Service interface
type service struct {
	repo    Repository
	cache   cache.Service
	tracing *tracing.TracingService
	metrics *metrics.MetricsService
}

// NewService creates a new user service
func NewService(repo Repository, cache cache.Service, tracing *tracing.TracingService, metrics *metrics.MetricsService) Service {
	return &service{
		repo:    repo,
		cache:   cache,
		tracing: tracing,
		metrics: metrics,
	}
}

// CreateUser creates a new user with business logic validation and caching
func (s *service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.create_user",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user.email", req.Email),
			attribute.String("user.username", req.Username),
			attribute.String("user.type", req.UserType),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting user creation",
		logger.Fields{
			"email":     req.Email,
			"username":  req.Username,
			"user_type": req.UserType,
		})

	// Business validation
	if err := s.validateCreateUserRequest(ctx, req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Validation failed")
		return nil, err
	}

	// Check email availability
	if err := s.repo.ValidateUserEmail(ctx, req.Email, nil); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Email validation failed")
		s.metrics.IncrementCounter("user_creation_errors", metrics.Fields{
			"error_type": "email_exists",
		})
		return nil, fmt.Errorf("email validation failed: %w", err)
	}

	// Check username availability if provided
	if req.Username != "" {
		if err := s.repo.ValidateUsername(ctx, req.Username, nil); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Username validation failed")
			s.metrics.IncrementCounter("user_creation_errors", metrics.Fields{
				"error_type": "username_exists",
			})
			return nil, fmt.Errorf("username validation failed: %w", err)
		}
	}

	// Repository operation with metrics
	timer := s.metrics.Timer("service_operation_duration", metrics.Fields{
		"operation": "create_user",
	})

	user, err := s.repo.CreateUser(ctx, req)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("service_errors", metrics.Fields{
			"operation": "create_user",
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to create user",
			logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Cache the user
	if err := s.cacheUser(ctx, user); err != nil {
		logger.WarnContext(ctx, "Failed to cache user",
			logger.Fields{"error": err.Error(), "user_id": user.ID.String()})
	}

	// Success metrics
	s.metrics.IncrementCounter("user_created_total", metrics.Fields{
		"status": "success",
	})

	s.metrics.ObserveHistogram("user_creation_duration",
		duration.Seconds(), metrics.Fields{})

	logger.InfoContext(ctx, "User created successfully",
		logger.Fields{
			"user_id":     user.ID.String(),
			"email":       user.Email,
			"duration_ms": duration.Milliseconds(),
		})

	return user, nil
}

// GetUserByID retrieves a user by ID with caching
func (s *service) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_user_by_id",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting user by ID",
		logger.Fields{"user_id": id.String()})

	// Check cache first
	cacheKey := fmt.Sprintf("user:id:%s", id.String())
	var user User

	cacheTimer := s.metrics.Timer("cache_operation_duration", metrics.Fields{
		"operation": "get",
		"cache_key": "user_id",
	})

	if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
		cacheTimer.Stop()
		// Cache hit
		s.metrics.IncrementCounter("cache_hits_total", metrics.Fields{
			"cache_key": "user_id",
		})
		logger.DebugContext(ctx, "Cache hit for user ID",
			logger.Fields{"user_id": id.String()})
		return &user, nil
	}
	cacheTimer.Stop()

	// Cache miss
	s.metrics.IncrementCounter("cache_misses_total", metrics.Fields{
		"cache_key": "user_id",
	})
	logger.DebugContext(ctx, "Cache miss for user ID",
		logger.Fields{"user_id": id.String()})

	// Get from repository
	dbUser, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to get user by ID",
			logger.Fields{"error": err.Error(), "user_id": id.String()})
		return nil, err
	}

	// Cache the result
	if err := s.cacheUser(ctx, dbUser); err != nil {
		logger.WarnContext(ctx, "Failed to cache user",
			logger.Fields{"error": err.Error(), "user_id": dbUser.ID.String()})
	}

	logger.DebugContext(ctx, "User retrieved from database",
		logger.Fields{"user_id": dbUser.ID.String()})

	return dbUser, nil
}

// GetUserByEmail retrieves a user by email with caching
func (s *service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_user_by_email",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user.email", email),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting user by email",
		logger.Fields{"email": email})

	// Check cache first
	cacheKey := fmt.Sprintf("user:email:%s", email)
	var user User

	cacheTimer := s.metrics.Timer("cache_operation_duration", metrics.Fields{
		"operation": "get",
		"cache_key": "user_email",
	})

	if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
		cacheTimer.Stop()
		// Cache hit
		s.metrics.IncrementCounter("cache_hits_total", metrics.Fields{
			"cache_key": "user_email",
		})
		logger.DebugContext(ctx, "Cache hit for user email",
			logger.Fields{"email": email})
		return &user, nil
	}
	cacheTimer.Stop()

	// Cache miss
	s.metrics.IncrementCounter("cache_misses_total", metrics.Fields{
		"cache_key": "user_email",
	})
	logger.DebugContext(ctx, "Cache miss for user email",
		logger.Fields{"email": email})

	// Get from repository
	dbUser, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to get user by email",
			logger.Fields{"error": err.Error(), "email": email})
		return nil, err
	}

	// Cache the result
	if err := s.cacheUser(ctx, dbUser); err != nil {
		logger.WarnContext(ctx, "Failed to cache user",
			logger.Fields{"error": err.Error(), "user_id": dbUser.ID.String()})
	}

	logger.DebugContext(ctx, "User retrieved from database",
		logger.Fields{"user_id": dbUser.ID.String(), "email": email})

	return dbUser, nil
}

// GetUserByUsername retrieves a user by username with caching
func (s *service) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_user_by_username",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user.username", username),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting user by username",
		logger.Fields{"username": username})

	// Check cache first
	cacheKey := fmt.Sprintf("user:username:%s", username)
	var user User

	cacheTimer := s.metrics.Timer("cache_operation_duration", metrics.Fields{
		"operation": "get",
		"cache_key": "user_username",
	})

	if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
		cacheTimer.Stop()
		// Cache hit
		s.metrics.IncrementCounter("cache_hits_total", metrics.Fields{
			"cache_key": "user_username",
		})
		logger.DebugContext(ctx, "Cache hit for user username",
			logger.Fields{"username": username})
		return &user, nil
	}
	cacheTimer.Stop()

	// Cache miss
	s.metrics.IncrementCounter("cache_misses_total", metrics.Fields{
		"cache_key": "user_username",
	})
	logger.DebugContext(ctx, "Cache miss for user username",
		logger.Fields{"username": username})

	// Get from repository
	dbUser, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to get user by username",
			logger.Fields{"error": err.Error(), "username": username})
		return nil, err
	}

	// Cache the result
	if err := s.cacheUser(ctx, dbUser); err != nil {
		logger.WarnContext(ctx, "Failed to cache user",
			logger.Fields{"error": err.Error(), "user_id": dbUser.ID.String()})
	}

	logger.DebugContext(ctx, "User retrieved from database",
		logger.Fields{"user_id": dbUser.ID.String(), "username": username})

	return dbUser, nil
}

// UpdateUser updates an existing user with business logic and cache invalidation
func (s *service) UpdateUser(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.update_user",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Updating user",
		logger.Fields{"user_id": id.String()})

	// Business validation
	if err := s.validateUpdateUserRequest(ctx, req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Validation failed")
		return nil, err
	}

	// Validate email if being updated
	if req.Email != nil {
		if err := s.repo.ValidateUserEmail(ctx, *req.Email, &id); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Email validation failed")
			return nil, fmt.Errorf("email validation failed: %w", err)
		}
	}

	// Validate username if being updated
	if req.Username != nil {
		if err := s.repo.ValidateUsername(ctx, *req.Username, &id); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Username validation failed")
			return nil, fmt.Errorf("username validation failed: %w", err)
		}
	}

	// Repository operation
	user, err := s.repo.UpdateUser(ctx, id, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to update user",
			logger.Fields{"error": err.Error(), "user_id": id.String()})
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Invalidate cache
	if err := s.invalidateUserCache(ctx, user); err != nil {
		logger.WarnContext(ctx, "Failed to invalidate user cache",
			logger.Fields{"error": err.Error(), "user_id": user.ID.String()})
	}

	// Cache the updated user
	if err := s.cacheUser(ctx, user); err != nil {
		logger.WarnContext(ctx, "Failed to cache updated user",
			logger.Fields{"error": err.Error(), "user_id": user.ID.String()})
	}

	logger.InfoContext(ctx, "User updated successfully",
		logger.Fields{"user_id": user.ID.String()})

	return user, nil
}

// DeleteUser deletes a user and invalidates cache
func (s *service) DeleteUser(ctx context.Context, id uuid.UUID, permanent bool) error {
	ctx, span := s.tracing.StartSpan(ctx, "service.delete_user",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user.id", id.String()),
			attribute.Bool("permanent", permanent),
		))
	defer span.End()

	logger.InfoContext(ctx, "Deleting user",
		logger.Fields{"user_id": id.String(), "permanent": permanent})

	// Get user first for cache invalidation
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "User not found")
		return err
	}

	// Repository operation
	if err := s.repo.DeleteUser(ctx, id, permanent); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to delete user",
			logger.Fields{"error": err.Error(), "user_id": id.String()})
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Invalidate cache
	if err := s.invalidateUserCache(ctx, user); err != nil {
		logger.WarnContext(ctx, "Failed to invalidate user cache",
			logger.Fields{"error": err.Error(), "user_id": user.ID.String()})
	}

	logger.InfoContext(ctx, "User deleted successfully",
		logger.Fields{"user_id": id.String(), "permanent": permanent})

	return nil
}

// ListUsers lists users with pagination and caching
func (s *service) ListUsers(ctx context.Context, req *ListUsersRequest) ([]*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.list_users",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.Int("limit", req.Limit),
			attribute.Int("offset", req.Offset),
		))
	defer span.End()

	logger.DebugContext(ctx, "Listing users",
		logger.Fields{"limit": req.Limit, "offset": req.Offset})

	// Generate cache key based on filters
	cacheKey := s.generateListCacheKey(req)
	var users []*User

	// Check cache first
	if err := s.cache.Get(ctx, cacheKey, &users); err == nil {
		s.metrics.IncrementCounter("cache_hits_total", metrics.Fields{
			"cache_key": "user_list",
		})
		logger.DebugContext(ctx, "Cache hit for user list")
		return users, nil
	}

	// Cache miss
	s.metrics.IncrementCounter("cache_misses_total", metrics.Fields{
		"cache_key": "user_list",
	})

	// Get from repository
	users, err := s.repo.ListUsers(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to list users",
			logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Cache the result (shorter TTL for lists)
	if err := s.cache.Set(ctx, cacheKey, users, 5*time.Minute); err != nil {
		logger.WarnContext(ctx, "Failed to cache user list",
			logger.Fields{"error": err.Error()})
	}

	logger.DebugContext(ctx, "Users listed successfully",
		logger.Fields{"count": len(users)})

	return users, nil
}

// GetUserWithDetails retrieves a user with complete details
func (s *service) GetUserWithDetails(ctx context.Context, id uuid.UUID) (*UserWithDetails, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_user_with_details",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting user with details",
		logger.Fields{"user_id": id.String()})

	// Repository operation
	userDetails, err := s.repo.GetUserWithDetails(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to get user with details",
			logger.Fields{"error": err.Error(), "user_id": id.String()})
		return nil, fmt.Errorf("failed to get user with details: %w", err)
	}

	logger.DebugContext(ctx, "User with details retrieved successfully",
		logger.Fields{"user_id": userDetails.User.ID.String()})

	return userDetails, nil
}

// CreatePerson creates a new person record
func (s *service) CreatePerson(ctx context.Context, req *CreatePersonRequest) (*Person, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.create_person",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("person.email", req.Email),
			attribute.String("person.first_name", req.FirstName),
			attribute.String("person.last_name", req.LastName),
		))
	defer span.End()

	logger.InfoContext(ctx, "Creating person",
		logger.Fields{
			"email":      req.Email,
			"first_name": req.FirstName,
			"last_name":  req.LastName,
		})

	// Business validation
	if err := s.validateCreatePersonRequest(ctx, req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Validation failed")
		return nil, err
	}

	// Repository operation
	person, err := s.repo.CreatePerson(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to create person",
			logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to create person: %w", err)
	}

	logger.InfoContext(ctx, "Person created successfully",
		logger.Fields{"person_id": person.ID.String()})

	return person, nil
}

// GetPersonByID retrieves a person by ID
func (s *service) GetPersonByID(ctx context.Context, id uuid.UUID) (*Person, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_person_by_id",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("person.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting person by ID",
		logger.Fields{"person_id": id.String()})

	// Repository operation
	person, err := s.repo.GetPersonByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to get person by ID",
			logger.Fields{"error": err.Error(), "person_id": id.String()})
		return nil, fmt.Errorf("failed to get person: %w", err)
	}

	logger.DebugContext(ctx, "Person retrieved successfully",
		logger.Fields{"person_id": person.ID.String()})

	return person, nil
}

// UpdatePerson updates an existing person
func (s *service) UpdatePerson(ctx context.Context, id uuid.UUID, req *CreatePersonRequest) (*Person, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.update_person",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("person.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Updating person",
		logger.Fields{"person_id": id.String()})

	// Repository operation
	person, err := s.repo.UpdatePerson(ctx, id, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to update person",
			logger.Fields{"error": err.Error(), "person_id": id.String()})
		return nil, fmt.Errorf("failed to update person: %w", err)
	}

	logger.InfoContext(ctx, "Person updated successfully",
		logger.Fields{"person_id": person.ID.String()})

	return person, nil
}

// CreateEmployee creates a new employee record
func (s *service) CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*Employee, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.create_employee",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("employee.number", req.EmployeeNumber),
			attribute.String("employee.position", req.PositionTitle),
		))
	defer span.End()

	logger.InfoContext(ctx, "Creating employee",
		logger.Fields{
			"employee_number": req.EmployeeNumber,
			"position_title":  req.PositionTitle,
		})

	// Business validation
	if err := s.validateCreateEmployeeRequest(ctx, req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Validation failed")
		return nil, err
	}

	// Check employee number availability
	if err := s.repo.ValidateEmployeeNumber(ctx, req.EmployeeNumber, nil); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Employee number validation failed")
		return nil, fmt.Errorf("employee number validation failed: %w", err)
	}

	// Repository operation
	employee, err := s.repo.CreateEmployee(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to create employee",
			logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	logger.InfoContext(ctx, "Employee created successfully",
		logger.Fields{"employee_id": employee.ID.String()})

	return employee, nil
}

// GetEmployeeByID retrieves an employee by ID
func (s *service) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*Employee, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_employee_by_id",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("employee.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting employee by ID",
		logger.Fields{"employee_id": id.String()})

	// Repository operation
	employee, err := s.repo.GetEmployeeByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to get employee by ID",
			logger.Fields{"error": err.Error(), "employee_id": id.String()})
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	logger.DebugContext(ctx, "Employee retrieved successfully",
		logger.Fields{"employee_id": employee.ID.String()})

	return employee, nil
}

// UpdateEmployee updates an existing employee
func (s *service) UpdateEmployee(ctx context.Context, id uuid.UUID, req *CreateEmployeeRequest) (*Employee, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.update_employee",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("employee.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Updating employee",
		logger.Fields{"employee_id": id.String()})

	// Repository operation
	employee, err := s.repo.UpdateEmployee(ctx, id, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to update employee",
			logger.Fields{"error": err.Error(), "employee_id": id.String()})
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	logger.InfoContext(ctx, "Employee updated successfully",
		logger.Fields{"employee_id": employee.ID.String()})

	return employee, nil
}

// AuthenticateUser authenticates a user by identifier and password
func (s *service) AuthenticateUser(ctx context.Context, identifier, password string) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.authenticate_user",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user.identifier", identifier),
		))
	defer span.End()

	logger.InfoContext(ctx, "Authenticating user",
		logger.Fields{"identifier": identifier})

	// Repository operation
	user, err := s.repo.AuthenticateUser(ctx, identifier, password)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Authentication failed")
		logger.WarnContext(ctx, "User authentication failed",
			logger.Fields{"error": err.Error(), "identifier": identifier})
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	logger.InfoContext(ctx, "User authenticated successfully",
		logger.Fields{"user_id": user.ID.String()})

	return user, nil
}

// UpdatePassword updates a user's password
func (s *service) UpdatePassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	ctx, span := s.tracing.StartSpan(ctx, "service.update_password",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user.id", userID.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Updating user password",
		logger.Fields{"user_id": userID.String()})

	// Repository operation
	if err := s.repo.UpdatePassword(ctx, userID, newPassword); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to update password",
			logger.Fields{"error": err.Error(), "user_id": userID.String()})
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate user cache after password change
	user, err := s.repo.GetUserByID(ctx, userID)
	if err == nil {
		if err := s.invalidateUserCache(ctx, user); err != nil {
			logger.WarnContext(ctx, "Failed to invalidate user cache after password change",
				logger.Fields{"error": err.Error(), "user_id": userID.String()})
		}
	}

	logger.InfoContext(ctx, "User password updated successfully",
		logger.Fields{"user_id": userID.String()})

	return nil
}

// GetUserRoles retrieves roles assigned to a user
func (s *service) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*UserRole, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_user_roles",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user.id", userID.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting user roles",
		logger.Fields{"user_id": userID.String()})

	// Repository operation
	roles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to get user roles",
			logger.Fields{"error": err.Error(), "user_id": userID.String()})
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	logger.DebugContext(ctx, "User roles retrieved successfully",
		logger.Fields{"user_id": userID.String(), "roles_count": len(roles)})

	return roles, nil
}

// AssignUserRole assigns a role to a user
func (s *service) AssignUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "service.assign_user_role",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user.id", userID.String()),
			attribute.String("role.id", roleID.String()),
			attribute.String("entity.id", entityID.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Assigning role to user",
		logger.Fields{
			"user_id":   userID.String(),
			"role_id":   roleID.String(),
			"entity_id": entityID.String(),
		})

	// Repository operation
	if err := s.repo.AssignUserRole(ctx, userID, roleID, entityID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to assign user role",
			logger.Fields{"error": err.Error()})
		return fmt.Errorf("failed to assign user role: %w", err)
	}

	logger.InfoContext(ctx, "Role assigned to user successfully",
		logger.Fields{
			"user_id":   userID.String(),
			"role_id":   roleID.String(),
			"entity_id": entityID.String(),
		})

	return nil
}

// RevokeUserRole revokes a role from a user
func (s *service) RevokeUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "service.revoke_user_role",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user.id", userID.String()),
			attribute.String("role.id", roleID.String()),
			attribute.String("entity.id", entityID.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Revoking role from user",
		logger.Fields{
			"user_id":   userID.String(),
			"role_id":   roleID.String(),
			"entity_id": entityID.String(),
		})

	// Repository operation
	if err := s.repo.RevokeUserRole(ctx, userID, roleID, entityID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to revoke user role",
			logger.Fields{"error": err.Error()})
		return fmt.Errorf("failed to revoke user role: %w", err)
	}

	logger.InfoContext(ctx, "Role revoked from user successfully",
		logger.Fields{
			"user_id":   userID.String(),
			"role_id":   roleID.String(),
			"entity_id": entityID.String(),
		})

	return nil
}

// ValidateUserEmail validates if an email is available
func (s *service) ValidateUserEmail(ctx context.Context, email string, excludeID *uuid.UUID) error {
	return s.repo.ValidateUserEmail(ctx, email, excludeID)
}

// ValidateUsername validates if a username is available
func (s *service) ValidateUsername(ctx context.Context, username string, excludeID *uuid.UUID) error {
	return s.repo.ValidateUsername(ctx, username, excludeID)
}

// ValidateEmployeeNumber validates if an employee number is available
func (s *service) ValidateEmployeeNumber(ctx context.Context, number string, excludeID *uuid.UUID) error {
	return s.repo.ValidateEmployeeNumber(ctx, number, excludeID)
}

// SearchUsers searches for users by query
func (s *service) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.search_users",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("search.query", query),
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		))
	defer span.End()

	logger.DebugContext(ctx, "Searching users",
		logger.Fields{
			"query":  query,
			"limit":  limit,
			"offset": offset,
		})

	// Repository operation
	users, err := s.repo.SearchUsers(ctx, query, limit, offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to search users",
			logger.Fields{"error": err.Error(), "query": query})
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	logger.DebugContext(ctx, "User search completed successfully",
		logger.Fields{
			"query":  query,
			"count":  len(users),
			"limit":  limit,
			"offset": offset,
		})

	return users, nil
}

// Helper functions for caching and validation

// cacheUser caches a user with multiple cache keys
func (s *service) cacheUser(ctx context.Context, user *User) error {
	cacheTTL := 30 * time.Minute

	// Cache by ID
	idKey := fmt.Sprintf("user:id:%s", user.ID.String())
	if err := s.cache.Set(ctx, idKey, user, cacheTTL); err != nil {
		return fmt.Errorf("failed to cache user by ID: %w", err)
	}

	// Cache by email
	emailKey := fmt.Sprintf("user:email:%s", user.Email)
	if err := s.cache.Set(ctx, emailKey, user, cacheTTL); err != nil {
		return fmt.Errorf("failed to cache user by email: %w", err)
	}

	// Cache by username if available
	if user.Username != nil && *user.Username != "" {
		usernameKey := fmt.Sprintf("user:username:%s", *user.Username)
		if err := s.cache.Set(ctx, usernameKey, user, cacheTTL); err != nil {
			return fmt.Errorf("failed to cache user by username: %w", err)
		}
	}

	return nil
}

// invalidateUserCache invalidates all cache entries for a user
func (s *service) invalidateUserCache(ctx context.Context, user *User) error {
	// Delete cache by ID
	idKey := fmt.Sprintf("user:id:%s", user.ID.String())
	if err := s.cache.Delete(ctx, idKey); err != nil {
		logger.WarnContext(ctx, "Failed to delete user cache by ID",
			logger.Fields{"error": err.Error(), "key": idKey})
	}

	// Delete cache by email
	emailKey := fmt.Sprintf("user:email:%s", user.Email)
	if err := s.cache.Delete(ctx, emailKey); err != nil {
		logger.WarnContext(ctx, "Failed to delete user cache by email",
			logger.Fields{"error": err.Error(), "key": emailKey})
	}

	// Delete cache by username if available
	if user.Username != nil && *user.Username != "" {
		usernameKey := fmt.Sprintf("user:username:%s", *user.Username)
		if err := s.cache.Delete(ctx, usernameKey); err != nil {
			logger.WarnContext(ctx, "Failed to delete user cache by username",
				logger.Fields{"error": err.Error(), "key": usernameKey})
		}
	}

	return nil
}

// generateListCacheKey generates a cache key for list operations
func (s *service) generateListCacheKey(req *ListUsersRequest) string {
	key := fmt.Sprintf("user:list:%d:%d", req.Limit, req.Offset)
	if req.UserType != nil {
		key += fmt.Sprintf(":type:%s", *req.UserType)
	}
	if req.AccountStatus != nil {
		key += fmt.Sprintf(":status:%s", *req.AccountStatus)
	}
	return key
}

// validateCreateUserRequest validates a create user request
func (s *service) validateCreateUserRequest(ctx context.Context, req *CreateUserRequest) error {
	if req.Email == "" {
		return errors.ErrInvalidInput.WithMessage("email is required")
	}
	if req.UserType == "" {
		return errors.ErrInvalidInput.WithMessage("user type is required")
	}
	if req.Password == "" {
		return errors.ErrInvalidInput.WithMessage("password is required")
	}
	if len(req.Password) < 8 {
		return errors.ErrInvalidInput.WithMessage("password must be at least 8 characters")
	}
	return nil
}

// validateUpdateUserRequest validates an update user request
func (s *service) validateUpdateUserRequest(ctx context.Context, req *UpdateUserRequest) error {
	if req.Email != nil && *req.Email == "" {
		return errors.ErrInvalidInput.WithMessage("email cannot be empty")
	}
	if req.UserType != nil && *req.UserType == "" {
		return errors.ErrInvalidInput.WithMessage("user type cannot be empty")
	}
	return nil
}

// validateCreatePersonRequest validates a create person request
func (s *service) validateCreatePersonRequest(ctx context.Context, req *CreatePersonRequest) error {
	if req.FirstName == "" {
		return errors.ErrInvalidInput.WithMessage("first name is required")
	}
	if req.LastName == "" {
		return errors.ErrInvalidInput.WithMessage("last name is required")
	}
	if req.Email == "" {
		return errors.ErrInvalidInput.WithMessage("email is required")
	}
	return nil
}

// validateCreateEmployeeRequest validates a create employee request
func (s *service) validateCreateEmployeeRequest(ctx context.Context, req *CreateEmployeeRequest) error {
	if req.EmployeeNumber == "" {
		return errors.ErrInvalidInput.WithMessage("employee number is required")
	}
	if req.PersonID == uuid.Nil {
		return errors.ErrInvalidInput.WithMessage("person ID is required")
	}
	return nil
}
