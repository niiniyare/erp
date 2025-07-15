package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/cache"
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
	
	// Permission evaluation operations
	EvaluatePermission(ctx context.Context, req *PermissionEvaluationRequest) (*PermissionEvaluationResult, error)
	GetUserEffectivePermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) ([]*EffectivePermission, error)
	BulkEvaluatePermissions(ctx context.Context, req *BulkPermissionEvaluationRequest) ([]*PermissionEvaluationResult, error)
	
	// Role hierarchy operations
	CalculateRoleHierarchy(ctx context.Context, roleID uuid.UUID) ([]*Role, error)
	GetInheritedPermissions(ctx context.Context, roleID uuid.UUID) ([]*Permission, error)
	
	// ABAC policy operations
	EvaluateABACPolicies(ctx context.Context, req *ABACEvaluationRequest) (*ABACEvaluationResult, error)
	TestPolicy(ctx context.Context, policyID uuid.UUID, req *PolicyTestRequest) (*PolicyTestResult, error)
}

// service implements the Service interface
type service struct {
	repo           Repository
	cache          cache.Service
	permissionCache *PermissionCacheService
	tracing        *tracing.TracingService
	metrics        *metrics.MetricsService
}

// NewService creates a new user service
func NewService(repo Repository, cache cache.Service, tracing *tracing.TracingService, metrics *metrics.MetricsService) Service {
	permissionCache := NewPermissionCacheService(cache, metrics, tracing)
	return &service{
		repo:           repo,
		cache:          cache,
		permissionCache: permissionCache,
		tracing:        tracing,
		metrics:        metrics,
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
			attribute.String("person.email", func() string {
				if req.Email != nil { return *req.Email }
				return ""
			}()),
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
			attribute.String("employee.position", func() string {
				if req.PositionTitle != nil { return *req.PositionTitle }
				return ""
			}()),
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

	// Check cache first using the permission cache service
	if cachedRoles, found, err := s.permissionCache.GetUserRoles(ctx, userID); err == nil && found {
		logger.DebugContext(ctx, "User roles retrieved from cache",
			logger.Fields{"user_id": userID.String(), "roles_count": len(cachedRoles)})
		return cachedRoles, nil
	}

	// Cache miss - get from repository
	roles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Repository operation failed")
		logger.ErrorContext(ctx, "Failed to get user roles",
			logger.Fields{"error": err.Error(), "user_id": userID.String()})
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Cache the result using the permission cache service
	if err := s.permissionCache.SetUserRoles(ctx, userID, roles); err != nil {
		logger.WarnContext(ctx, "Failed to cache user roles",
			logger.Fields{"error": err.Error(), "user_id": userID.String()})
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

	// Invalidate user permissions and role cache after role assignment
	s.invalidateUserAndRoleCache(ctx, userID, roleID)

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

	// Invalidate user permissions and role cache after role revocation
	s.invalidateUserAndRoleCache(ctx, userID, roleID)

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

/*
CACHE INVALIDATION GUIDE FOR ROLE/PERMISSION OPERATIONS:

When implementing new methods that modify roles or permissions, use these guidelines:

1. USER ROLE ASSIGNMENT/REVOCATION:
   - Call invalidateUserAndRoleCache(ctx, userID, roleID)
   - This invalidates both user permission cache and role cache

2. ROLE HIERARCHY MODIFICATIONS (CreateRole, UpdateRole, DeleteRole):
   - Call invalidateRoleHierarchyCache(ctx, roleID)
   - Also call invalidateUserPermissionsForRole(ctx, roleID) if the role has users

3. PERMISSION MODIFICATIONS (CreatePermission, UpdatePermission, DeletePermission):
   - Call invalidatePermissionCache(ctx, permissionID)
   - This flushes all permission caches as permissions affect multiple users/roles

4. ROLE PERMISSION ASSIGNMENTS:
   - Call invalidateRoleHierarchyCache(ctx, roleID)
   - Call invalidateUserPermissionsForRole(ctx, roleID)

5. POLICY MODIFICATIONS:
   - Call permissionCache.FlushPermissionCache(ctx) as policies affect evaluation

Examples of methods that would need cache invalidation:
- CreateRole() -> invalidateRoleHierarchyCache + invalidateUserPermissionsForRole
- UpdateRole() -> invalidateRoleHierarchyCache + invalidateUserPermissionsForRole
- DeleteRole() -> invalidateRoleHierarchyCache + invalidateUserPermissionsForRole
- CreatePermission() -> invalidatePermissionCache
- UpdatePermission() -> invalidatePermissionCache
- DeletePermission() -> invalidatePermissionCache
- AssignRolePermission() -> invalidateRoleHierarchyCache + invalidateUserPermissionsForRole
- RevokeRolePermission() -> invalidateRoleHierarchyCache + invalidateUserPermissionsForRole
*/

// invalidateUserAndRoleCache invalidates both user permissions and role cache
func (s *service) invalidateUserAndRoleCache(ctx context.Context, userID, roleID uuid.UUID) {
	// Invalidate user permissions cache
	if err := s.permissionCache.InvalidateUserPermissions(ctx, userID); err != nil {
		logger.WarnContext(ctx, "Failed to invalidate user permissions cache",
			logger.Fields{"error": err.Error(), "user_id": userID.String()})
	}

	// Invalidate role cache
	if err := s.permissionCache.InvalidateRoleCache(ctx, roleID); err != nil {
		logger.WarnContext(ctx, "Failed to invalidate role cache",
			logger.Fields{"error": err.Error(), "role_id": roleID.String()})
	}
}

// invalidateRoleHierarchyCache invalidates role hierarchy cache for a role and its related roles
func (s *service) invalidateRoleHierarchyCache(ctx context.Context, roleID uuid.UUID) {
	// Invalidate the role's cache
	if err := s.permissionCache.InvalidateRoleCache(ctx, roleID); err != nil {
		logger.WarnContext(ctx, "Failed to invalidate role cache",
			logger.Fields{"error": err.Error(), "role_id": roleID.String()})
	}

	// TODO: When implementing role hierarchy modifications, also invalidate parent/child role caches
	// This would require getting the role hierarchy and invalidating related roles
}

// invalidatePermissionCache invalidates permission-related caches
func (s *service) invalidatePermissionCache(ctx context.Context, permissionID uuid.UUID) {
	// TODO: When implementing permission modification methods, add specific permission cache invalidation
	// For now, we flush all permission caches as a fallback
	if err := s.permissionCache.FlushPermissionCache(ctx); err != nil {
		logger.WarnContext(ctx, "Failed to flush permission cache",
			logger.Fields{"error": err.Error(), "permission_id": permissionID.String()})
	}
}

// invalidateUserPermissionsForRole invalidates user permissions for all users with a specific role
func (s *service) invalidateUserPermissionsForRole(ctx context.Context, roleID uuid.UUID) {
	// TODO: When implementing role modification methods, this should:
	// 1. Get all users with the role
	// 2. Invalidate their permission caches
	// For now, we'll flush all user permission caches
	if err := s.permissionCache.FlushPermissionCache(ctx); err != nil {
		logger.WarnContext(ctx, "Failed to flush permission cache after role modification",
			logger.Fields{"error": err.Error(), "role_id": roleID.String()})
	}
}

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
	if user.Username != "" {
		usernameKey := fmt.Sprintf("user:username:%s", user.Username)
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
	if user.Username != "" {
		usernameKey := fmt.Sprintf("user:username:%s", user.Username)
		if err := s.cache.Delete(ctx, usernameKey); err != nil {
			logger.WarnContext(ctx, "Failed to delete user cache by username",
				logger.Fields{"error": err.Error(), "key": usernameKey})
		}
	}

	// Invalidate permission-related cache using the specialized permission cache service
	if err := s.permissionCache.InvalidateUserPermissions(ctx, user.ID); err != nil {
		logger.WarnContext(ctx, "Failed to invalidate user permissions cache",
			logger.Fields{"error": err.Error(), "user_id": user.ID.String()})
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
		return fmt.Errorf("email is required")
	}
	if req.UserType == "" {
		return fmt.Errorf("user type is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

// validateUpdateUserRequest validates an update user request
func (s *service) validateUpdateUserRequest(ctx context.Context, req *UpdateUserRequest) error {
	if req.Email != nil && *req.Email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if req.UserType != nil && *req.UserType == "" {
		return fmt.Errorf("user type cannot be empty")
	}
	return nil
}

// validateCreatePersonRequest validates a create person request
func (s *service) validateCreatePersonRequest(ctx context.Context, req *CreatePersonRequest) error {
	if req.FirstName == "" {
		return fmt.Errorf("first name is required")
	}
	if req.LastName == "" {
		return fmt.Errorf("last name is required")
	}
	if req.Email == nil || *req.Email == "" {
		return fmt.Errorf("email is required")
	}
	return nil
}

// validateCreateEmployeeRequest validates a create employee request
func (s *service) validateCreateEmployeeRequest(ctx context.Context, req *CreateEmployeeRequest) error {
	if req.EmployeeNumber == "" {
		return fmt.Errorf("employee number is required")
	}
	if req.PersonID == uuid.Nil {
		return fmt.Errorf("person ID is required")
	}
	return nil
}

// ===== PERMISSION EVALUATION METHODS =====

// EvaluatePermission evaluates a permission request using RBAC and ABAC
func (s *service) EvaluatePermission(ctx context.Context, req *PermissionEvaluationRequest) (*PermissionEvaluationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.evaluate_permission",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user.id", req.UserID.String()),
			attribute.String("resource.name", req.ResourceName),
			attribute.String("action.name", req.ActionName),
		))
	defer span.End()

	logger.DebugContext(ctx, "Evaluating permission",
		logger.Fields{
			"user_id":       req.UserID.String(),
			"resource_name": req.ResourceName,
			"action_name":   req.ActionName,
		})

	startTime := time.Now()

	// Check cache first using the specialized permission cache service
	if cachedResult, found, err := s.permissionCache.GetPermissionEvaluationResult(ctx, req); err == nil && found {
		// Cache hit - return cached result
		cachedResult.EvaluationTimeMS = int(time.Since(startTime).Milliseconds())
		return cachedResult, nil
	}

	// Cache miss - evaluate permission

	// Step 1: Get user's roles and permissions
	userRoles, err := s.repo.GetUserRoles(ctx, req.UserID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Step 2: Calculate role hierarchy and inherited permissions
	allRoles := make([]*Role, 0)
	allPermissions := make([]*Permission, 0)
	
	for _, userRole := range userRoles {
		// Get role hierarchy
		roleHierarchy, err := s.CalculateRoleHierarchy(ctx, userRole.RoleID)
		if err != nil {
			logger.WarnContext(ctx, "Failed to calculate role hierarchy",
				logger.Fields{"role_id": userRole.RoleID.String(), "error": err.Error()})
			continue
		}
		allRoles = append(allRoles, roleHierarchy...)

		// Get inherited permissions for each role
		for _, role := range roleHierarchy {
			permissions, err := s.GetInheritedPermissions(ctx, role.ID)
			if err != nil {
				logger.WarnContext(ctx, "Failed to get inherited permissions",
					logger.Fields{"role_id": role.ID.String(), "error": err.Error()})
				continue
			}
			allPermissions = append(allPermissions, permissions...)
		}
	}

	// Step 3: Check RBAC permissions
	rbacResult := s.evaluateRBACPermissions(ctx, req, allPermissions)

	// Step 4: Evaluate ABAC policies if needed
	abacResult := &ABACEvaluationResult{Allowed: true} // Default allow if no policies
	if rbacResult.Allowed || s.shouldEvaluateABACForDeny(req) {
		abacReq := &ABACEvaluationRequest{
			UserID:       req.UserID,
			ResourceName: req.ResourceName,
			ActionName:   req.ActionName,
			EntityID:     req.EntityID,
			Context:      req.Context,
		}
		abacResult, err = s.EvaluateABACPolicies(ctx, abacReq)
		if err != nil {
			logger.WarnContext(ctx, "ABAC evaluation failed, falling back to RBAC",
				logger.Fields{"error": err.Error()})
			abacResult = &ABACEvaluationResult{Allowed: rbacResult.Allowed}
		}
	}

	// Step 5: Combine RBAC and ABAC results
	finalDecision := rbacResult.Allowed && abacResult.Allowed
	
	// Build result
	result := PermissionEvaluationResult{
		Allowed:           finalDecision,
		PolicyDecisions:   append(rbacResult.PolicyDecisions, abacResult.PolicyDecisions...),
		EffectiveRoles:    s.extractRoleNames(allRoles),
		EvaluationTimeMS:  int(time.Since(startTime).Milliseconds()),
		CacheHit:          false,
		RBACResult:        rbacResult,
		ABACResult:        abacResult,
	}

	// Cache the result using the specialized permission cache service
	if err := s.permissionCache.SetPermissionEvaluationResult(ctx, req, &result); err != nil {
		logger.WarnContext(ctx, "Failed to cache permission evaluation",
			logger.Fields{"error": err.Error()})
	}

	// Metrics
	s.metrics.IncrementCounter("permission_evaluations_total", metrics.Fields{
		"decision": fmt.Sprintf("%t", finalDecision),
	})
	s.metrics.ObserveHistogram("permission_evaluation_duration",
		float64(result.EvaluationTimeMS)/1000, metrics.Fields{})

	logger.DebugContext(ctx, "Permission evaluation completed",
		logger.Fields{
			"user_id":        req.UserID.String(),
			"resource_name":  req.ResourceName,
			"action_name":    req.ActionName,
			"allowed":        finalDecision,
			"duration_ms":    result.EvaluationTimeMS,
		})

	return &result, nil
}

// GetUserEffectivePermissions retrieves all effective permissions for a user
func (s *service) GetUserEffectivePermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) ([]*EffectivePermission, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_user_effective_permissions",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user.id", userID.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting user effective permissions",
		logger.Fields{"user_id": userID.String()})

	// Check cache first using the specialized permission cache service
	if cachedPermissions, found, err := s.permissionCache.GetUserPermissions(ctx, userID, entityID); err == nil && found {
		return cachedPermissions, nil
	}

	// Get all user roles
	userRoles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	effectivePermissions := make([]*EffectivePermission, 0)
	processedPermissions := make(map[string]bool) // Deduplication

	for _, userRole := range userRoles {
		// Skip if entity filtering is requested and doesn't match
		if entityID != nil && userRole.EntityID != *entityID {
			continue
		}

		// Get role hierarchy
		roleHierarchy, err := s.CalculateRoleHierarchy(ctx, userRole.RoleID)
		if err != nil {
			logger.WarnContext(ctx, "Failed to calculate role hierarchy",
				logger.Fields{"role_id": userRole.RoleID.String(), "error": err.Error()})
			continue
		}

		// Process each role in hierarchy
		for _, role := range roleHierarchy {
			permissions, err := s.GetInheritedPermissions(ctx, role.ID)
			if err != nil {
				logger.WarnContext(ctx, "Failed to get role permissions",
					logger.Fields{"role_id": role.ID.String(), "error": err.Error()})
				continue
			}

			for _, permission := range permissions {
				// Create unique key for deduplication
				key := fmt.Sprintf("%s:%s:%s", permission.ResourceID, permission.ActionID, userRole.EntityID)
				if processedPermissions[key] {
					continue
				}
				processedPermissions[key] = true

				effectivePermission := &EffectivePermission{
					Permission:     permission,
					GrantedByRole:  role,
					EntityID:       userRole.EntityID,
					AssignmentType: userRole.AssignmentType,
					ExpiresAt:      userRole.ExpiresAt,
				}
				effectivePermissions = append(effectivePermissions, effectivePermission)
			}
		}
	}

	// Cache the result using the specialized permission cache service
	if err := s.permissionCache.SetUserPermissions(ctx, userID, entityID, effectivePermissions); err != nil {
		logger.WarnContext(ctx, "Failed to cache user permissions",
			logger.Fields{"error": err.Error()})
	}

	logger.DebugContext(ctx, "User effective permissions retrieved",
		logger.Fields{
			"user_id":           userID.String(),
			"permissions_count": len(effectivePermissions),
		})

	return effectivePermissions, nil
}

// BulkEvaluatePermissions evaluates multiple permissions efficiently
func (s *service) BulkEvaluatePermissions(ctx context.Context, req *BulkPermissionEvaluationRequest) ([]*PermissionEvaluationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.bulk_evaluate_permissions",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.Int("requests.count", len(req.Requests)),
		))
	defer span.End()

	logger.DebugContext(ctx, "Bulk evaluating permissions",
		logger.Fields{"requests_count": len(req.Requests)})

	results := make([]*PermissionEvaluationResult, len(req.Requests))
	
	// Process in parallel for better performance
	type evalResult struct {
		index  int
		result *PermissionEvaluationResult
		err    error
	}
	
	resultChan := make(chan evalResult, len(req.Requests))
	
	// Launch goroutines for parallel evaluation
	for i, permReq := range req.Requests {
		go func(idx int, request *PermissionEvaluationRequest) {
			result, err := s.EvaluatePermission(ctx, request)
			resultChan <- evalResult{index: idx, result: result, err: err}
		}(i, permReq)
	}
	
	// Collect results
	for i := 0; i < len(req.Requests); i++ {
		evalRes := <-resultChan
		if evalRes.err != nil {
			logger.WarnContext(ctx, "Permission evaluation failed in bulk operation",
				logger.Fields{"index": evalRes.index, "error": evalRes.err.Error()})
			// Create a denied result for failed evaluations
			results[evalRes.index] = &PermissionEvaluationResult{
				Allowed:           false,
				PolicyDecisions:   []string{"evaluation_failed"},
				EffectiveRoles:    []string{},
				EvaluationTimeMS:  0,
				CacheHit:          false,
			}
		} else {
			results[evalRes.index] = evalRes.result
		}
	}

	logger.DebugContext(ctx, "Bulk permission evaluation completed",
		logger.Fields{"requests_count": len(req.Requests)})

	return results, nil
}

// CalculateRoleHierarchy calculates the complete role hierarchy for a role
func (s *service) CalculateRoleHierarchy(ctx context.Context, roleID uuid.UUID) ([]*Role, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.calculate_role_hierarchy",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("role.id", roleID.String()),
		))
	defer span.End()

	// Check cache first using the specialized permission cache service
	if cachedRoles, found, err := s.permissionCache.GetRoleHierarchy(ctx, roleID); err == nil && found {
		return cachedRoles, nil
	}

	// Cache miss - calculate hierarchy

	roles, err := s.repo.GetRoleHierarchy(ctx, roleID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get role hierarchy: %w", err)
	}

	// Cache the result using the specialized permission cache service
	if err := s.permissionCache.SetRoleHierarchy(ctx, roleID, roles); err != nil {
		logger.WarnContext(ctx, "Failed to cache role hierarchy",
			logger.Fields{"error": err.Error()})
	}

	return roles, nil
}

// GetInheritedPermissions gets all permissions for a role including inherited ones
func (s *service) GetInheritedPermissions(ctx context.Context, roleID uuid.UUID) ([]*Permission, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_inherited_permissions",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("role.id", roleID.String()),
		))
	defer span.End()

	// Check cache first
	cacheKey := fmt.Sprintf("role:permissions:%s", roleID.String())
	var permissions []*Permission
	
	if err := s.cache.Get(ctx, cacheKey, &permissions); err == nil {
		s.metrics.IncrementCounter("role_permissions_cache_hits_total", metrics.Fields{})
		return permissions, nil
	}

	// Cache miss - get permissions
	s.metrics.IncrementCounter("role_permissions_cache_misses_total", metrics.Fields{})

	permissions, err := s.repo.GetRolePermissions(ctx, roleID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	// Cache the result
	if err := s.cache.Set(ctx, cacheKey, permissions, 15*time.Minute); err != nil {
		logger.WarnContext(ctx, "Failed to cache role permissions",
			logger.Fields{"error": err.Error()})
	}

	return permissions, nil
}

// EvaluateABACPolicies evaluates ABAC policies for a request
func (s *service) EvaluateABACPolicies(ctx context.Context, req *ABACEvaluationRequest) (*ABACEvaluationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.evaluate_abac_policies",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user.id", req.UserID.String()),
			attribute.String("resource.name", req.ResourceName),
			attribute.String("action.name", req.ActionName),
		))
	defer span.End()

	logger.DebugContext(ctx, "Evaluating ABAC policies",
		logger.Fields{
			"user_id":       req.UserID.String(),
			"resource_name": req.ResourceName,
			"action_name":   req.ActionName,
		})

	// Get applicable policies
	policies, err := s.repo.GetApplicablePolicies(ctx, req.ResourceName, req.ActionName)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get applicable policies: %w", err)
	}

	// If no policies, default to allow
	if len(policies) == 0 {
		return &ABACEvaluationResult{
			Allowed:           true,
			PolicyDecisions:   []string{"no_policies_applicable"},
			ApplicablePolicies: []string{},
		}, nil
	}

	// Get user context for evaluation
	userContext, err := s.buildUserContext(ctx, req.UserID, req.EntityID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to build user context: %w", err)
	}

	// Evaluate each policy
	policyDecisions := make([]string, 0)
	applicablePolicies := make([]string, 0)
	allowCount := 0
	denyCount := 0

	for _, policy := range policies {
		// Check if policy target matches
		if !s.policyTargetMatches(policy, req, userContext) {
			continue
		}
		
		applicablePolicies = append(applicablePolicies, policy.Name)
		
		// Evaluate policy rule
		allowed, err := s.evaluatePolicyRule(ctx, policy, req, userContext)
		if err != nil {
			logger.WarnContext(ctx, "Policy evaluation failed",
				logger.Fields{"policy_id": policy.ID.String(), "error": err.Error()})
			continue
		}

		decision := fmt.Sprintf("policy_%s_%s", policy.Name, policy.Effect)
		if allowed {
			if policy.Effect == "ALLOW" {
				allowCount++
				decision += "_granted"
			} else if policy.Effect == "DENY" {
				denyCount++
				decision += "_denied"
			}
		} else {
			decision += "_not_applicable"
		}
		
		policyDecisions = append(policyDecisions, decision)
	}

	// Determine final decision (DENY takes precedence)
	finalDecision := allowCount > 0 && denyCount == 0

	result := &ABACEvaluationResult{
		Allowed:            finalDecision,
		PolicyDecisions:    policyDecisions,
		ApplicablePolicies: applicablePolicies,
		EvaluationDetails: map[string]interface{}{
			"allow_count": allowCount,
			"deny_count":  denyCount,
			"total_policies": len(policies),
		},
	}

	logger.DebugContext(ctx, "ABAC policy evaluation completed",
		logger.Fields{
			"user_id":        req.UserID.String(),
			"resource_name":  req.ResourceName,
			"action_name":    req.ActionName,
			"allowed":        finalDecision,
			"policies_count": len(applicablePolicies),
		})

	return result, nil
}

// TestPolicy tests a specific policy against a request
func (s *service) TestPolicy(ctx context.Context, policyID uuid.UUID, req *PolicyTestRequest) (*PolicyTestResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.test_policy",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("policy.id", policyID.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Testing policy",
		logger.Fields{"policy_id": policyID.String()})

	// Get the policy
	policy, err := s.repo.GetPolicyByID(ctx, policyID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	// Build evaluation request
	evalReq := &ABACEvaluationRequest{
		UserID:       req.UserID,
		ResourceName: req.ResourceName,
		ActionName:   req.ActionName,
		EntityID:     req.EntityID,
		Context:      req.Context,
	}

	// Get user context
	userContext, err := s.buildUserContext(ctx, req.UserID, req.EntityID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to build user context: %w", err)
	}

	// Test target matching
	targetMatches := s.policyTargetMatches(policy, evalReq, userContext)
	
	var ruleResult bool
	var ruleError error
	
	if targetMatches {
		// Test rule evaluation
		ruleResult, ruleError = s.evaluatePolicyRule(ctx, policy, evalReq, userContext)
	}

	result := &PolicyTestResult{
		PolicyID:      policyID,
		PolicyName:    policy.Name,
		TargetMatches: targetMatches,
		RuleResult:    ruleResult && targetMatches,
		Effect:        policy.Effect,
		Details: map[string]interface{}{
			"target_matches": targetMatches,
			"rule_result":    ruleResult,
			"rule_error":     ruleError,
		},
	}

	logger.DebugContext(ctx, "Policy test completed",
		logger.Fields{
			"policy_id":      policyID.String(),
			"target_matches": targetMatches,
			"rule_result":    ruleResult,
		})

	return result, nil
}

// Helper methods for permission evaluation

// generatePermissionCacheKey is now deprecated - use PermissionCacheService.GeneratePermissionEvaluationKey instead
// This method is kept for backwards compatibility but will be removed in a future version

func (s *service) evaluateRBACPermissions(ctx context.Context, req *PermissionEvaluationRequest, permissions []*Permission) *RBACEvaluationResult {
	decisions := make([]string, 0)
	allowed := false

	// Find matching permissions
	for _, permission := range permissions {
		// This would need to check resource and action matching
		// For now, simplified check based on names
		if s.permissionMatches(permission, req.ResourceName, req.ActionName) {
			if permission.Effect == "ALLOW" {
				allowed = true
				decisions = append(decisions, fmt.Sprintf("rbac_allow_%s", permission.Name))
			} else if permission.Effect == "DENY" {
				allowed = false
				decisions = append(decisions, fmt.Sprintf("rbac_deny_%s", permission.Name))
				break // DENY takes precedence
			}
		}
	}

	if len(decisions) == 0 {
		decisions = append(decisions, "rbac_no_matching_permissions")
	}

	return &RBACEvaluationResult{
		Allowed:         allowed,
		PolicyDecisions: decisions,
	}
}

func (s *service) shouldEvaluateABACForDeny(req *PermissionEvaluationRequest) bool {
	// Always evaluate ABAC for comprehensive security
	return true
}

func (s *service) extractRoleNames(roles []*Role) []string {
	names := make([]string, len(roles))
	for i, role := range roles {
		names[i] = role.Name
	}
	return names
}

func (s *service) buildUserContext(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (map[string]interface{}, error) {
	// Get user details with person and employee info
	userDetails, err := s.repo.GetUserWithDetails(ctx, userID)
	if err != nil {
		return nil, err
	}

	context := map[string]interface{}{
		"user_id":    userID.String(),
		"user_type":  userDetails.User.UserType,
		"email":      userDetails.User.Email,
		"attributes": userDetails.User.UserAttributes,
	}

	if userDetails.Person != nil {
		context["person_type"] = userDetails.Person.PersonType
		context["person_attributes"] = userDetails.Person.SecurityAttributes
	}

	if userDetails.Employee != nil {
		context["security_level"] = userDetails.Employee.SecurityLevel
		context["department_id"] = userDetails.Employee.DepartmentID
		context["manager_id"] = userDetails.Employee.ManagerID
		context["access_attributes"] = userDetails.Employee.AccessAttributes
	}

	if entityID != nil {
		context["entity_id"] = entityID.String()
	}

	// Add current time for time-based policies
	context["current_time"] = time.Now()
	context["current_hour"] = time.Now().Hour()
	context["current_day"] = int(time.Now().Weekday())

	return context, nil
}

func (s *service) permissionMatches(permission *Permission, resourceName, actionName string) bool {
	// This is a simplified check - in production you'd want more sophisticated matching
	// that considers the actual resource and action IDs from the database
	return true // Placeholder - implement proper matching logic
}

func (s *service) policyTargetMatches(policy *Policy, req *ABACEvaluationRequest, userContext map[string]interface{}) bool {
	// Simplified target matching - implement full JSON target evaluation
	// This would parse the policy.Target JSON and match against user context
	return true // Placeholder - implement proper target matching
}

func (s *service) evaluatePolicyRule(ctx context.Context, policy *Policy, req *ABACEvaluationRequest, userContext map[string]interface{}) (bool, error) {
	// Simplified rule evaluation - implement full JSON rule evaluation engine
	// This would parse the policy.Rule JSON and evaluate against user context
	return true, nil // Placeholder - implement proper rule evaluation
}
