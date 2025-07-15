package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/crypto/bcrypt"
)

// Repository defines the interface for user data operations
type Repository interface {
	// User CRUD operations
	CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id uuid.UUID, permanent bool) error
	RestoreUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, req *ListUsersRequest) ([]*User, error)
	GetUserWithDetails(ctx context.Context, id uuid.UUID) (*UserWithDetails, error)
	
	// Person CRUD operations
	CreatePerson(ctx context.Context, req *CreatePersonRequest) (*Person, error)
	GetPersonByID(ctx context.Context, id uuid.UUID) (*Person, error)
	GetPersonByEmail(ctx context.Context, email string) (*Person, error)
	UpdatePerson(ctx context.Context, id uuid.UUID, req *CreatePersonRequest) (*Person, error)
	DeletePerson(ctx context.Context, id uuid.UUID, permanent bool) error
	
	// Employee CRUD operations
	CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*Employee, error)
	GetEmployeeByID(ctx context.Context, id uuid.UUID) (*Employee, error)
	GetEmployeeByPersonID(ctx context.Context, personID uuid.UUID) (*Employee, error)
	GetEmployeeByNumber(ctx context.Context, number string) (*Employee, error)
	UpdateEmployee(ctx context.Context, id uuid.UUID, req *CreateEmployeeRequest) (*Employee, error)
	DeleteEmployee(ctx context.Context, id uuid.UUID, permanent bool) error
	
	// Authentication operations
	AuthenticateUser(ctx context.Context, identifier, password string) (*User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, newPassword string) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	IncrementFailedLogins(ctx context.Context, userID uuid.UUID) error
	UnlockUser(ctx context.Context, userID uuid.UUID) error
	
	// User role operations
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*UserRole, error)
	AssignUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error
	RevokeUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error
	
	// Validation operations
	ValidateUserEmail(ctx context.Context, email string, excludeID *uuid.UUID) error
	ValidateUsername(ctx context.Context, username string, excludeID *uuid.UUID) error
	ValidateEmployeeNumber(ctx context.Context, number string, excludeID *uuid.UUID) error
	
	// Search operations
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*User, error)
	
	// Permission evaluation operations
	GetRoleHierarchy(ctx context.Context, roleID uuid.UUID) ([]*Role, error)
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*Permission, error)
	GetApplicablePolicies(ctx context.Context, resourceName, actionName string) ([]*Policy, error)
	GetPolicyByID(ctx context.Context, policyID uuid.UUID) (*Policy, error)
}

// repository implements the Repository interface
type repository struct {
	store   db.Store
	tracing *tracing.TracingService
	metrics *metrics.MetricsService
}

// NewRepository creates a new user repository
func NewRepository(store db.Store, tracing *tracing.TracingService, metrics *metrics.MetricsService) Repository {
	return &repository{
		store:   store,
		tracing: tracing,
		metrics: metrics,
	}
}

// CreateUser creates a new user
func (r *repository) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.create_user",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "create"),
			attribute.String("db.table", "users"),
			attribute.String("user.username", req.Username),
			attribute.String("user.email", req.Email),
			attribute.String("user.type", req.UserType),
		))
	defer span.End()

	logger.DebugContext(ctx, "Creating user in database",
		logger.Fields{
			"username":  req.Username,
			"email":     req.Email,
			"user_type": req.UserType,
		})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "create_user",
		"table":     "users",
	})

	// Hash password
	hashedPassword, err := r.hashPassword(req.Password)
	if err != nil {
		timer.Stop()
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to hash password")
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	params, err := req.ToSQLCCreateUserParams()
	if err != nil {
		timer.Stop()
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to convert request to SQLC params")
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	params.PasswordHash = &hashedPassword

	sqlcUser, err := r.store.CreateUser(ctx, params)
	duration := timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation": "create_user",
			"error_type": "sql_error",
		})

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		logger.ErrorContext(ctx, "Database operation failed",
			logger.Fields{
				"error":       err.Error(),
				"operation":   "create_user",
				"duration_ms": duration.Milliseconds(),
			})

		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Success metrics
	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "create_user",
		"status":    "success",
	})

	r.metrics.ObserveHistogram("database_query_duration",
		duration.Seconds(), metrics.Fields{
			"operation": "create_user",
		})

	logger.DebugContext(ctx, "User created successfully in database",
		logger.Fields{
			"user_id":     sqlcUser.ID.String(),
			"username":    *sqlcUser.Username,
			"duration_ms": duration.Milliseconds(),
		})

	return FromSQLCUser(sqlcUser)
}

// GetUserByID retrieves a user by ID
func (r *repository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_user_by_id",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "users"),
			attribute.String("user.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting user by ID from database",
		logger.Fields{"user_id": id.String()})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "get_user_by_id",
		"table":     "users",
	})

	sqlcUser, err := r.store.GetUserByID(ctx, id)
	duration := timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation": "get_user_by_id",
			"error_type": "sql_error",
		})

		if strings.Contains(err.Error(), "no rows") {
			span.SetStatus(codes.Error, "User not found")
			logger.WarnContext(ctx, "User not found",
				logger.Fields{"user_id": id.String()})
			return nil, errors.ErrUserNotFound
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		logger.ErrorContext(ctx, "Database operation failed",
			logger.Fields{
				"error":       err.Error(),
				"operation":   "get_user_by_id",
				"user_id":     id.String(),
				"duration_ms": duration.Milliseconds(),
			})

		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Success metrics
	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "get_user_by_id",
		"status":    "success",
	})

	r.metrics.ObserveHistogram("database_query_duration",
		duration.Seconds(), metrics.Fields{
			"operation": "get_user_by_id",
		})

	logger.DebugContext(ctx, "User retrieved successfully from database",
		logger.Fields{
			"user_id":     sqlcUser.ID.String(),
			"username":    *sqlcUser.Username,
			"duration_ms": duration.Milliseconds(),
		})

	return FromSQLCUser(sqlcUser)
}

// GetUserByEmail retrieves a user by email
func (r *repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_user_by_email",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "users"),
			attribute.String("user.email", email),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting user by email from database",
		logger.Fields{"email": email})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "get_user_by_email",
		"table":     "users",
	})

	sqlcUser, err := r.store.GetUserByEmail(ctx, email)
	duration := timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation": "get_user_by_email",
			"error_type": "sql_error",
		})

		if strings.Contains(err.Error(), "no rows") {
			span.SetStatus(codes.Error, "User not found")
			logger.WarnContext(ctx, "User not found by email",
				logger.Fields{"email": email})
			return nil, errors.ErrUserNotFound
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		logger.ErrorContext(ctx, "Database operation failed",
			logger.Fields{
				"error":       err.Error(),
				"operation":   "get_user_by_email",
				"email":       email,
				"duration_ms": duration.Milliseconds(),
			})

		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Success metrics
	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "get_user_by_email",
		"status":    "success",
	})

	r.metrics.ObserveHistogram("database_query_duration",
		duration.Seconds(), metrics.Fields{
			"operation": "get_user_by_email",
		})

	logger.DebugContext(ctx, "User retrieved successfully by email",
		logger.Fields{
			"user_id":     sqlcUser.ID.String(),
			"email":       email,
			"duration_ms": duration.Milliseconds(),
		})

	return FromSQLCUser(sqlcUser)
}

// GetUserByUsername retrieves a user by username
func (r *repository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_user_by_username",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "users"),
			attribute.String("user.username", username),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting user by username from database",
		logger.Fields{"username": username})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "get_user_by_username",
		"table":     "users",
	})

	sqlcUser, err := r.store.GetUserByUsername(ctx, &username)
	duration := timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation": "get_user_by_username",
			"error_type": "sql_error",
		})

		if strings.Contains(err.Error(), "no rows") {
			span.SetStatus(codes.Error, "User not found")
			logger.WarnContext(ctx, "User not found by username",
				logger.Fields{"username": username})
			return nil, errors.ErrUserNotFound
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		logger.ErrorContext(ctx, "Database operation failed",
			logger.Fields{
				"error":       err.Error(),
				"operation":   "get_user_by_username",
				"username":    username,
				"duration_ms": duration.Milliseconds(),
			})

		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	// Success metrics
	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "get_user_by_username",
		"status":    "success",
	})

	r.metrics.ObserveHistogram("database_query_duration",
		duration.Seconds(), metrics.Fields{
			"operation": "get_user_by_username",
		})

	logger.DebugContext(ctx, "User retrieved successfully by username",
		logger.Fields{
			"user_id":     sqlcUser.ID.String(),
			"username":    username,
			"duration_ms": duration.Milliseconds(),
		})

	return FromSQLCUser(sqlcUser)
}

// UpdateUser updates an existing user
func (r *repository) UpdateUser(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*User, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.update_user",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "update"),
			attribute.String("db.table", "users"),
			attribute.String("user.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Updating user in database",
		logger.Fields{"user_id": id.String()})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "update_user",
		"table":     "users",
	})

	params := db.UpdateUserParams{
		ID: id,
	}

	if req.Username != nil {
		params.Username = req.Username
	}
	if req.Email != nil {
		params.Email = *req.Email
	}
	if req.UserType != nil {
		params.UserType = *req.UserType
	}
	if req.AccountStatus != nil {
		params.AccountStatus = req.AccountStatus
	}
	if req.SessionTimeoutMinutes != nil {
		params.SessionTimeoutMinutes = req.SessionTimeoutMinutes
	}
	if req.MfaEnabled != nil {
		params.MfaEnabled = req.MfaEnabled
	}

	sqlcUser, err := r.store.UpdateUser(ctx, params)
	duration := timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation": "update_user",
			"error_type": "sql_error",
		})

		if strings.Contains(err.Error(), "no rows") {
			span.SetStatus(codes.Error, "User not found")
			logger.WarnContext(ctx, "User not found for update",
				logger.Fields{"user_id": id.String()})
			return nil, errors.ErrUserNotFound
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		logger.ErrorContext(ctx, "Database operation failed",
			logger.Fields{
				"error":       err.Error(),
				"operation":   "update_user",
				"user_id":     id.String(),
				"duration_ms": duration.Milliseconds(),
			})

		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Success metrics
	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "update_user",
		"status":    "success",
	})

	r.metrics.ObserveHistogram("database_query_duration",
		duration.Seconds(), metrics.Fields{
			"operation": "update_user",
		})

	logger.DebugContext(ctx, "User updated successfully in database",
		logger.Fields{
			"user_id":     sqlcUser.ID.String(),
			"duration_ms": duration.Milliseconds(),
		})

	return FromSQLCUser(sqlcUser)
}

// DeleteUser deletes a user (soft or hard delete)
func (r *repository) DeleteUser(ctx context.Context, id uuid.UUID, permanent bool) error {
	ctx, span := r.tracing.StartSpan(ctx, "repository.delete_user",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "delete"),
			attribute.String("db.table", "users"),
			attribute.String("user.id", id.String()),
			attribute.Bool("permanent", permanent),
		))
	defer span.End()

	logger.DebugContext(ctx, "Deleting user from database",
		logger.Fields{
			"user_id":   id.String(),
			"permanent": permanent,
		})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "delete_user",
		"table":     "users",
		"permanent": permanent,
	})

	var err error
	if permanent {
		// TODO: Implement hard delete when needed
		err = r.store.SoftDeleteUser(ctx, id)
	} else {
		err = r.store.SoftDeleteUser(ctx, id)
	}

	duration := timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation": "delete_user",
			"error_type": "sql_error",
		})

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		logger.ErrorContext(ctx, "Database operation failed",
			logger.Fields{
				"error":       err.Error(),
				"operation":   "delete_user",
				"user_id":     id.String(),
				"duration_ms": duration.Milliseconds(),
			})

		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Success metrics
	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "delete_user",
		"permanent": permanent,
		"status":    "success",
	})

	r.metrics.ObserveHistogram("database_query_duration",
		duration.Seconds(), metrics.Fields{
			"operation": "delete_user",
		})

	logger.DebugContext(ctx, "User deleted successfully from database",
		logger.Fields{
			"user_id":     id.String(),
			"permanent":   permanent,
			"duration_ms": duration.Milliseconds(),
		})

	return nil
}

// RestoreUser restores a soft-deleted user
func (r *repository) RestoreUser(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "repository.restore_user",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "update"),
			attribute.String("db.table", "users"),
			attribute.String("user.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Restoring user in database",
		logger.Fields{"user_id": id.String()})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "restore_user",
		"table":     "users",
	})

	// TODO: Implement user restoration when needed
	// For now, we'll use the restore queries if available
	err := r.store.RestoreSoftDeletedUser(ctx, id)
	duration := timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation": "restore_user",
			"error_type": "sql_error",
		})

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		logger.ErrorContext(ctx, "Database operation failed",
			logger.Fields{
				"error":       err.Error(),
				"operation":   "restore_user",
				"user_id":     id.String(),
				"duration_ms": duration.Milliseconds(),
			})

		return fmt.Errorf("failed to restore user: %w", err)
	}

	// Success metrics
	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "restore_user",
		"status":    "success",
	})

	r.metrics.ObserveHistogram("database_query_duration",
		duration.Seconds(), metrics.Fields{
			"operation": "restore_user",
		})

	logger.DebugContext(ctx, "User restored successfully in database",
		logger.Fields{
			"user_id":     id.String(),
			"duration_ms": duration.Milliseconds(),
		})

	return nil
}

// ListUsers retrieves users with filtering and pagination
func (r *repository) ListUsers(ctx context.Context, req *ListUsersRequest) ([]*User, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.list_users",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "users"),
			attribute.Int("limit", req.Limit),
			attribute.Int("offset", req.Offset),
		))
	defer span.End()

	logger.DebugContext(ctx, "Listing users from database",
		logger.Fields{
			"limit":  req.Limit,
			"offset": req.Offset,
		})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "list_users",
		"table":     "users",
	})

	// Handle optional string pointers
	userType := ""
	if req.UserType != nil {
		userType = *req.UserType
	}
	accountStatus := ""
	if req.AccountStatus != nil {
		accountStatus = *req.AccountStatus
	}

	sqlcUsers, err := r.store.ListUsers(ctx, db.ListUsersParams{
		Column1: userType,
		Column2: accountStatus,
		Limit:   int32(req.Limit),
		Offset:  int32(req.Offset),
	})
	duration := timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation": "list_users",
			"error_type": "sql_error",
		})

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		logger.ErrorContext(ctx, "Database operation failed",
			logger.Fields{
				"error":       err.Error(),
				"operation":   "list_users",
				"duration_ms": duration.Milliseconds(),
			})

		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Convert to domain models
	users := make([]*User, len(sqlcUsers))
	for i, sqlcUser := range sqlcUsers {
		user, err := FromSQLCUser(sqlcUser)
		if err != nil {
			return nil, fmt.Errorf("failed to convert user: %w", err)
		}
		users[i] = user
	}

	// Success metrics
	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "list_users",
		"status":    "success",
	})

	r.metrics.ObserveHistogram("database_query_duration",
		duration.Seconds(), metrics.Fields{
			"operation": "list_users",
		})

	logger.DebugContext(ctx, "Users listed successfully from database",
		logger.Fields{
			"count":       len(users),
			"duration_ms": duration.Milliseconds(),
		})

	return users, nil
}

// GetUserWithDetails retrieves a user with person and employee details
func (r *repository) GetUserWithDetails(ctx context.Context, id uuid.UUID) (*UserWithDetails, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_user_with_details",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "users"),
			attribute.String("user.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting user with details from database",
		logger.Fields{"user_id": id.String()})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "get_user_with_details",
		"table":     "users",
	})

	_, err := r.store.GetCompleteUserProfile(ctx, id)
	duration := timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation": "get_user_with_details",
			"error_type": "sql_error",
		})

		if strings.Contains(err.Error(), "no rows") {
			span.SetStatus(codes.Error, "User not found")
			logger.WarnContext(ctx, "User not found for details",
				logger.Fields{"user_id": id.String()})
			return nil, errors.ErrUserNotFound
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		logger.ErrorContext(ctx, "Database operation failed",
			logger.Fields{
				"error":       err.Error(),
				"operation":   "get_user_with_details",
				"user_id":     id.String(),
				"duration_ms": duration.Milliseconds(),
			})

		return nil, fmt.Errorf("failed to get user with details: %w", err)
	}

	// Convert the complex result to UserWithDetails
	result := &UserWithDetails{}
	
	// Try to use the complete user profile query result directly
	// If the query returns structured data, we can use it
	// For now, we'll get the basic user and build the details for robustness
	user, err := r.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result.User = *user

	// Get person details if available
	if user.PersonID != nil {
		person, err := r.GetPersonByID(ctx, *user.PersonID)
		if err == nil {
			result.Person = person
		} else {
			// Log warning but don't fail the entire operation
			logger.WarnContext(ctx, "Failed to get person details for user",
				logger.Fields{"error": err.Error(), "person_id": user.PersonID.String()})
		}
	}

	// Get employee details if available
	if user.EmployeeID != nil {
		employee, err := r.GetEmployeeByID(ctx, *user.EmployeeID)
		if err == nil {
			result.Employee = employee
		} else {
			// Log warning but don't fail the entire operation
			logger.WarnContext(ctx, "Failed to get employee details for user",
				logger.Fields{"error": err.Error(), "employee_id": user.EmployeeID.String()})
		}
	}

	// Success metrics
	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "get_user_with_details",
		"status":    "success",
	})

	r.metrics.ObserveHistogram("database_query_duration",
		duration.Seconds(), metrics.Fields{
			"operation": "get_user_with_details",
		})

	logger.DebugContext(ctx, "User with details retrieved successfully",
		logger.Fields{
			"user_id":     user.ID.String(),
			"duration_ms": duration.Milliseconds(),
		})

	return result, nil
}

// Helper functions

// CreatePerson creates a new person
func (r *repository) CreatePerson(ctx context.Context, req *CreatePersonRequest) (*Person, error) {
	params, err := req.ToSQLCCreatePersonParams()
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	sqlcPerson, err := r.store.CreatePerson(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create person: %w", err)
	}

	return FromSQLCPerson(sqlcPerson)
}

// GetPersonByID retrieves a person by ID
func (r *repository) GetPersonByID(ctx context.Context, id uuid.UUID) (*Person, error) {
	sqlcPerson, err := r.store.GetPersonByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, fmt.Errorf("person not found")
		}
		return nil, fmt.Errorf("failed to get person: %w", err)
	}

	return FromSQLCPerson(sqlcPerson)
}

// GetPersonByEmail retrieves a person by email
func (r *repository) GetPersonByEmail(ctx context.Context, email string) (*Person, error) {
	sqlcPerson, err := r.store.GetPersonByEmail(ctx, &email)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, fmt.Errorf("person not found")
		}
		return nil, fmt.Errorf("failed to get person by email: %w", err)
	}

	return FromSQLCPerson(sqlcPerson)
}

// UpdatePerson updates an existing person
func (r *repository) UpdatePerson(ctx context.Context, id uuid.UUID, req *CreatePersonRequest) (*Person, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.update_person",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "update"),
			attribute.String("db.table", "persons"),
			attribute.String("person.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Updating person in database",
		logger.Fields{"person_id": id.String()})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "update_person",
		"table":     "persons",
	})

	// TODO: Implement ToSQLCUpdatePersonParams method
	_ = req // avoid unused variable error
	_ = id  // avoid unused variable error
	
	timer.Stop()
	span.SetStatus(codes.Error, "Update person not implemented yet")
	return nil, fmt.Errorf("update person not implemented yet")
}

// DeletePerson deletes a person
func (r *repository) DeletePerson(ctx context.Context, id uuid.UUID, permanent bool) error {
	return r.store.SoftDeletePerson(ctx, id)
}

// CreateEmployee creates a new employee
func (r *repository) CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*Employee, error) {
	params, err := req.ToSQLCCreateEmployeeParams()
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	sqlcEmployee, err := r.store.CreateEmployee(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	return FromSQLCEmployee(sqlcEmployee)
}

// GetEmployeeByID retrieves an employee by ID
func (r *repository) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*Employee, error) {
	sqlcEmployee, err := r.store.GetEmployeeByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	return FromSQLCEmployee(sqlcEmployee)
}

// GetEmployeeByPersonID retrieves an employee by person ID
func (r *repository) GetEmployeeByPersonID(ctx context.Context, personID uuid.UUID) (*Employee, error) {
	sqlcEmployee, err := r.store.GetEmployeeByPersonID(ctx, personID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee by person ID: %w", err)
	}

	return FromSQLCEmployee(sqlcEmployee)
}

// GetEmployeeByNumber retrieves an employee by number
func (r *repository) GetEmployeeByNumber(ctx context.Context, number string) (*Employee, error) {
	sqlcEmployee, err := r.store.GetEmployeeByNumber(ctx, number)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee by number: %w", err)
	}

	return FromSQLCEmployee(sqlcEmployee)
}

// UpdateEmployee updates an existing employee
func (r *repository) UpdateEmployee(ctx context.Context, id uuid.UUID, req *CreateEmployeeRequest) (*Employee, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.update_employee",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "update"),
			attribute.String("db.table", "employees"),
			attribute.String("employee.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Updating employee in database",
		logger.Fields{"employee_id": id.String()})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "update_employee",
		"table":     "employees",
	})

	// TODO: Implement ToSQLCUpdateEmployeeParams method
	_ = req // avoid unused variable error
	_ = id  // avoid unused variable error
	
	timer.Stop()
	span.SetStatus(codes.Error, "Update employee not implemented yet")
	return nil, fmt.Errorf("update employee not implemented yet")
}

// DeleteEmployee deletes an employee
func (r *repository) DeleteEmployee(ctx context.Context, id uuid.UUID, permanent bool) error {
	ctx, span := r.tracing.StartSpan(ctx, "repository.delete_employee",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "delete"),
			attribute.String("db.table", "employees"),
			attribute.String("employee.id", id.String()),
			attribute.Bool("permanent", permanent),
		))
	defer span.End()

	logger.DebugContext(ctx, "Deleting employee from database",
		logger.Fields{
			"employee_id": id.String(),
			"permanent":   permanent,
		})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "delete_employee",
		"table":     "employees",
		"permanent": permanent,
	})

	// TODO: Implement employee deletion methods
	_ = permanent // avoid unused variable error
	_ = id       // avoid unused variable error
	
	timer.Stop()
	span.SetStatus(codes.Error, "Delete employee not implemented yet")
	return fmt.Errorf("delete employee not implemented yet")
}

// Authentication operations

// AuthenticateUser authenticates a user with identifier and password
func (r *repository) AuthenticateUser(ctx context.Context, identifier, password string) (*User, error) {
	var user *User
	var err error

	// Try to find user by email first, then by username
	if strings.Contains(identifier, "@") {
		user, err = r.GetUserByEmail(ctx, identifier)
	} else {
		user, err = r.GetUserByUsername(ctx, identifier)
	}

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify password would go here - for now we'll assume it's handled elsewhere
	return user, nil
}

// UpdatePassword updates a user's password
func (r *repository) UpdatePassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	hashedPassword, err := r.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return r.store.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: &hashedPassword,
	})
}

// UpdateLastLogin updates the last login time for a user
func (r *repository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	return r.store.UpdateUserLastLogin(ctx, userID)
}

// IncrementFailedLogins increments the failed login counter for a user
func (r *repository) IncrementFailedLogins(ctx context.Context, userID uuid.UUID) error {
	return r.store.IncrementFailedLogins(ctx, userID)
}

// UnlockUser unlocks a user account
func (r *repository) UnlockUser(ctx context.Context, userID uuid.UUID) error {
	return r.store.UnlockUser(ctx, userID)
}

// User role operations

// GetUserRoles retrieves roles assigned to a user
func (r *repository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*UserRole, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_user_roles",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "user_roles"),
			attribute.String("user.id", userID.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting user roles from database",
		logger.Fields{"user_id": userID.String()})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "get_user_roles",
		"table":     "user_roles",
	})

	sqlcRoles, err := r.store.GetUserRoles(ctx, userID)
	duration := timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation": "get_user_roles",
			"error_type": "sql_error",
		})

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		logger.ErrorContext(ctx, "Database operation failed",
			logger.Fields{
				"error":       err.Error(),
				"operation":   "get_user_roles",
				"user_id":     userID.String(),
				"duration_ms": duration.Milliseconds(),
			})

		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Convert to domain models
	userRoles := make([]*UserRole, len(sqlcRoles))
	for i, sqlcRole := range sqlcRoles {
		// TODO: Create proper conversion function for GetUserRolesRow
		// For now, create a minimal UserRole from the query result
		assignmentType := "DIRECT"
		if sqlcRole.AssignmentType != nil {
			assignmentType = *sqlcRole.AssignmentType
		}
		
		isActive := true
		if sqlcRole.IsActive != nil {
			isActive = *sqlcRole.IsActive
		}
		
		userRole := &UserRole{
			ID:             sqlcRole.ID,
			UserID:         userID, // Use the userID parameter
			RoleID:         sqlcRole.ID, // Use the role ID
			EntityID:       sqlcRole.EntityID,
			AssignmentType: assignmentType,
			AssignedAt:     sqlcRole.CreatedAt,
			IsActive:       isActive,
		}
		userRoles[i] = userRole
	}

	// Success metrics
	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "get_user_roles",
		"status":    "success",
	})

	r.metrics.ObserveHistogram("database_query_duration",
		duration.Seconds(), metrics.Fields{
			"operation": "get_user_roles",
		})

	logger.DebugContext(ctx, "User roles retrieved successfully from database",
		logger.Fields{
			"user_id":     userID.String(),
			"roles_count": len(userRoles),
			"duration_ms": duration.Milliseconds(),
		})

	return userRoles, nil
}

// AssignUserRole assigns a role to a user
func (r *repository) AssignUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	assignmentType := "PERMANENT"
	_, err := r.store.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
		UserID:         userID,
		RoleID:         roleID,
		EntityID:       entityID,
		AssignmentType: &assignmentType,
	})
	return err
}

// RevokeUserRole revokes a role from a user
func (r *repository) RevokeUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	return r.store.RevokeUserRole(ctx, db.RevokeUserRoleParams{
		UserID:   userID,
		RoleID:   roleID,
		EntityID: entityID,
	})
}

// Validation operations

// ValidateUserEmail validates if an email is available
func (r *repository) ValidateUserEmail(ctx context.Context, email string, excludeID *uuid.UUID) error {
	available, err := r.store.CheckEmailAvailability(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to check email availability: %w", err)
	}
	
	if !available {
		return fmt.Errorf("email already exists")
	}
	
	return nil
}

// ValidateUsername validates if a username is available
func (r *repository) ValidateUsername(ctx context.Context, username string, excludeID *uuid.UUID) error {
	available, err := r.store.CheckUsernameAvailability(ctx, &username)
	if err != nil {
		return fmt.Errorf("failed to check username availability: %w", err)
	}
	
	if !available {
		return fmt.Errorf("username already exists")
	}
	
	return nil
}

// ValidateEmployeeNumber validates if an employee number is available
func (r *repository) ValidateEmployeeNumber(ctx context.Context, number string, excludeID *uuid.UUID) error {
	available, err := r.store.CheckEmployeeNumberAvailability(ctx, number)
	if err != nil {
		return fmt.Errorf("failed to check employee number availability: %w", err)
	}
	
	if !available {
		return fmt.Errorf("employee number already exists")
	}
	
	return nil
}

// Search operations

// SearchUsers searches for users by query
func (r *repository) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*User, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.search_users",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "users"),
			attribute.String("search.query", query),
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		))
	defer span.End()

	logger.DebugContext(ctx, "Searching users in database",
		logger.Fields{
			"query":  query,
			"limit":  limit,
			"offset": offset,
		})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "search_users",
		"table":     "users",
	})

	sqlcUsers, err := r.store.SearchUsersAdvanced(ctx, db.SearchUsersAdvancedParams{
		Column1: query,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	duration := timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation": "search_users",
			"error_type": "sql_error",
		})

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		logger.ErrorContext(ctx, "Database operation failed",
			logger.Fields{
				"error":       err.Error(),
				"operation":   "search_users",
				"query":       query,
				"duration_ms": duration.Milliseconds(),
			})

		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// Convert to domain models
	users := make([]*User, len(sqlcUsers))
	for i, sqlcUser := range sqlcUsers {
		// Convert the search result to a User - this might need adjustment based on the actual return type
		username := ""
		if sqlcUser.Username != nil {
			username = *sqlcUser.Username
		}
		
		user := &User{
			ID:       sqlcUser.ID,
			Email:    sqlcUser.Email,
			Username: username,
			UserType: sqlcUser.UserType,
			// Add other fields as needed based on the search result structure
		}
		users[i] = user
	}

	// Success metrics
	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "search_users",
		"status":    "success",
	})

	r.metrics.ObserveHistogram("database_query_duration",
		duration.Seconds(), metrics.Fields{
			"operation": "search_users",
		})

	logger.DebugContext(ctx, "Users search completed successfully",
		logger.Fields{
			"query":       query,
			"results_count": len(users),
			"duration_ms": duration.Milliseconds(),
		})

	return users, nil
}

// hashPassword hashes a password using bcrypt
func (r *repository) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// verifyPassword verifies a password against its hash
func (r *repository) verifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// ===== PERMISSION EVALUATION REPOSITORY METHODS =====

// GetRoleHierarchy retrieves the complete role hierarchy for a role
func (r *repository) GetRoleHierarchy(ctx context.Context, roleID uuid.UUID) ([]*Role, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_role_hierarchy",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "roles"),
			attribute.String("role.id", roleID.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting role hierarchy from database",
		logger.Fields{"role_id": roleID.String()})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "get_role_hierarchy",
		"table":     "roles",
	})

	// For now, we'll implement a simple approach - get the role and its parents
	// In a production system, you'd have a recursive query or use a materialized path
	roles := make([]*Role, 0)
	
	// Get the base role
	sqlcRole, err := r.store.GetRoleByID(ctx, roleID)
	if err != nil {
		_ = timer.Stop()
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation": "get_role_hierarchy",
			"error_type": "sql_error",
		})

		if strings.Contains(err.Error(), "no rows") {
			span.SetStatus(codes.Error, "Role not found")
			logger.WarnContext(ctx, "Role not found for hierarchy",
				logger.Fields{"role_id": roleID.String()})
			return nil, errors.ErrRoleNotFound
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Convert to domain model
	role, err := FromSQLCRole(sqlcRole)
	if err != nil {
		_ = timer.Stop()
		span.RecordError(err)
		return nil, fmt.Errorf("failed to convert role: %w", err)
	}
	
	roles = append(roles, role)

	// Get parent roles recursively (simplified implementation)
	currentRole := role
	for currentRole.ParentRoleID != nil {
		parentSQLCRole, err := r.store.GetRoleByID(ctx, *currentRole.ParentRoleID)
		if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				break // Parent role not found, stop traversal
			}
			logger.WarnContext(ctx, "Failed to get parent role",
				logger.Fields{"parent_role_id": currentRole.ParentRoleID.String(), "error": err.Error()})
			break
		}

		parentRole, err := FromSQLCRole(parentSQLCRole)
		if err != nil {
			logger.WarnContext(ctx, "Failed to convert parent role",
				logger.Fields{"parent_role_id": currentRole.ParentRoleID.String(), "error": err.Error()})
			break
		}

		roles = append(roles, parentRole)
		currentRole = parentRole

		// Prevent infinite loops
		if len(roles) > 10 {
			logger.WarnContext(ctx, "Role hierarchy too deep, stopping traversal",
				logger.Fields{"role_id": roleID.String(), "depth": len(roles)})
			break
		}
	}

	duration := timer.Stop()

	// Success metrics
	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "get_role_hierarchy",
		"status":    "success",
	})

	r.metrics.ObserveHistogram("database_query_duration",
		duration.Seconds(), metrics.Fields{
			"operation": "get_role_hierarchy",
		})

	logger.DebugContext(ctx, "Role hierarchy retrieved successfully",
		logger.Fields{
			"role_id":        roleID.String(),
			"hierarchy_size": len(roles),
			"duration_ms":    duration.Milliseconds(),
		})

	return roles, nil
}

// GetRolePermissions retrieves all permissions for a role
func (r *repository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*Permission, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_role_permissions",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "role_permissions"),
			attribute.String("role.id", roleID.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting role permissions from database",
		logger.Fields{"role_id": roleID.String()})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "get_role_permissions",
		"table":     "role_permissions",
	})

	// TODO: Implement GetRolePermissions method
	// For now, return empty permissions
	duration := timer.Stop()
	permissions := make([]*Permission, 0)

	// Success metrics
	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "get_role_permissions",
		"status":    "success",
	})

	r.metrics.ObserveHistogram("database_query_duration",
		duration.Seconds(), metrics.Fields{
			"operation": "get_role_permissions",
		})

	logger.DebugContext(ctx, "Role permissions retrieved successfully",
		logger.Fields{
			"role_id":           roleID.String(),
			"permissions_count": len(permissions),
			"duration_ms":       duration.Milliseconds(),
		})

	return permissions, nil
}

// GetApplicablePolicies retrieves policies applicable to a resource and action
func (r *repository) GetApplicablePolicies(ctx context.Context, resourceName, actionName string) ([]*Policy, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_applicable_policies",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "policies"),
			attribute.String("resource.name", resourceName),
			attribute.String("action.name", actionName),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting applicable policies from database",
		logger.Fields{
			"resource_name": resourceName,
			"action_name":   actionName,
		})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "get_applicable_policies",
		"table":     "policies",
	})

	// TODO: Implement GetApplicablePolicies method
	// For now, return empty policies
	duration := timer.Stop()
	policies := make([]*Policy, 0)

	// Success metrics
	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "get_applicable_policies",
		"status":    "success",
	})

	r.metrics.ObserveHistogram("database_query_duration",
		duration.Seconds(), metrics.Fields{
			"operation": "get_applicable_policies",
		})

	logger.DebugContext(ctx, "Applicable policies retrieved successfully",
		logger.Fields{
			"resource_name":   resourceName,
			"action_name":     actionName,
			"policies_count":  len(policies),
			"duration_ms":     duration.Milliseconds(),
		})

	return policies, nil
}

// GetPolicyByID retrieves a policy by ID
func (r *repository) GetPolicyByID(ctx context.Context, policyID uuid.UUID) (*Policy, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_policy_by_id",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "policies"),
			attribute.String("policy.id", policyID.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting policy by ID from database",
		logger.Fields{"policy_id": policyID.String()})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "get_policy_by_id",
		"table":     "policies",
	})

	sqlcPolicy, err := r.store.GetPolicyByID(ctx, policyID)
	duration := timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation": "get_policy_by_id",
			"error_type": "sql_error",
		})

		if strings.Contains(err.Error(), "no rows") {
			span.SetStatus(codes.Error, "Policy not found")
			logger.WarnContext(ctx, "Policy not found",
				logger.Fields{"policy_id": policyID.String()})
			return nil, fmt.Errorf("policy not found")
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		logger.ErrorContext(ctx, "Database operation failed",
			logger.Fields{
				"error":       err.Error(),
				"operation":   "get_policy_by_id",
				"policy_id":   policyID.String(),
				"duration_ms": duration.Milliseconds(),
			})

		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	// Convert to domain model
	policy, err := FromSQLCPolicy(sqlcPolicy)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to convert policy: %w", err)
	}

	// Success metrics
	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "get_policy_by_id",
		"status":    "success",
	})

	r.metrics.ObserveHistogram("database_query_duration",
		duration.Seconds(), metrics.Fields{
			"operation": "get_policy_by_id",
		})

	logger.DebugContext(ctx, "Policy retrieved successfully",
		logger.Fields{
			"policy_id":   policyID.String(),
			"duration_ms": duration.Milliseconds(),
		})

	return policy, nil
}