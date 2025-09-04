package repo

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// userRepository implements UserRepository interface
type userRepository struct {
	store   db.Store
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewUserRepository creates a new user repository
func NewUserRepository(
	store db.Store,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) UserRepository {
	return &userRepository{
		store:   store,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	ctx, span := r.tracer.StartSpan(ctx, "user_repository.Create")
	defer span.End()

	var createdUser *model.User
	err := r.store.WithTenant(ctx, user.TenantID, func(ctx context.Context, store db.Store) error {
		// Convert user metadata to JSON bytes
		userAttributesJSON, _ := json.Marshal(user.Metadata)
		settingsJSON, _ := json.Marshal(map[string]any{})

		// Convert domain model to SQLC params
		params := db.CreateUserParams{
			EntityID:              user.TenantID, // Using TenantID as EntityID for now
			PersonID:              nil,           // Will be set when linking to Person
			EmployeeID:            nil,           // Will be set when linking to Employee
			Username:              user.Email,    // Use email as username
			Email:                 user.Email,
			PasswordHash:          &user.PasswordHash,
			UserType:              "INTERNAL", // Default user type
			AccountStatus:         stringPtr(string(user.AccountStatus)),
			SessionTimeoutMinutes: int32Ptr(480), // Default 8 hours
			MfaEnabled:            &user.MFAEnabled,
			UserAttributes:        userAttributesJSON,
			Settings:              settingsJSON,
		}

		// Create user using SQLC
		dbUser, err := store.CreateUser(ctx, params)
		if err != nil {
			// Handle database-specific errors
			if isDuplicateKeyError(err) {
				return errors.NewBusinessError("USER_EMAIL_EXISTS", "User with this email already exists")
			}
			return fmt.Errorf("failed to create user: %w", err)
		}

		// Convert back to domain model
		createdUser = &model.User{
			ID:               dbUser.ID,
			TenantID:         dbUser.TenantID,
			Email:            dbUser.Email,
			PasswordHash:     getStringValue(dbUser.PasswordHash),
			FirstName:        user.FirstName, // These aren't in the DB user table
			LastName:         user.LastName,  // They're in the persons table
			PhoneNumber:      user.PhoneNumber,
			AccountStatus:    model.UserAccountStatus(getStringValue(dbUser.AccountStatus)),
			EmailVerified:    false, // Default value
			PhoneVerified:    false, // Default value
			MFAEnabled:       getBoolValue(dbUser.MfaEnabled),
			FailedLoginCount: int(getInt32Value(dbUser.FailedLoginAttempts)),
			Metadata:         user.Metadata,
			CreatedAt:        dbUser.CreatedAt,
			UpdatedAt:        dbUser.UpdatedAt,
		}

		r.metrics.IncrementCounter("user_repository_create_success", nil)
		return nil
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("user_repository_create_error", nil)
		return nil, err
	}

	r.logger.InfoContext(ctx, "User created successfully",
		logger.Fields{"user_id": user.ID, "tenant_id": user.TenantID})

	return createdUser, nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	ctx, span := r.tracer.StartSpan(ctx, "user_repository.GetByID")
	defer span.End()

	// Get current tenant from context (this should be handled by service layer)
	tenantID := getTenantIDFromContext(ctx)

	var user *model.User
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, store db.Store) error {
		dbUser, err := store.GetUserByID(ctx, id)
		if err != nil {
			if err == db.ErrNoRows {
				return errors.NewBusinessError("USER_NOT_FOUND", "User not found")
			}
			return fmt.Errorf("failed to get user: %w", err)
		}

		// Convert to domain model
		user = &model.User{
			ID:               dbUser.ID,
			TenantID:         dbUser.TenantID,
			Email:            dbUser.Email,
			PasswordHash:     getStringValue(dbUser.PasswordHash),
			FirstName:        "",  // Will come from Person entity
			LastName:         "",  // Will come from Person entity
			PhoneNumber:      nil, // Will come from Person entity
			AccountStatus:    model.UserAccountStatus(getStringValue(dbUser.AccountStatus)),
			EmailVerified:    false, // Default value
			PhoneVerified:    false, // Default value
			MFAEnabled:       getBoolValue(dbUser.MfaEnabled),
			FailedLoginCount: int(getInt32Value(dbUser.FailedLoginAttempts)),
			Metadata:         make(map[string]any),
			CreatedAt:        dbUser.CreatedAt,
			UpdatedAt:        dbUser.UpdatedAt,
		}

		return nil
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("user_repository_get_error", nil)
		return nil, err
	}

	r.metrics.IncrementCounter("user_repository_get_success", nil)
	return user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, span := r.tracer.StartSpan(ctx, "user_repository.GetByEmail")
	defer span.End()

	tenantID := getTenantIDFromContext(ctx)

	var user *model.User
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, store db.Store) error {
		dbUser, err := store.GetUserByEmail(ctx, email)
		if err != nil {
			if err == db.ErrNoRows {
				return errors.NewBusinessError("USER_NOT_FOUND", "User not found")
			}
			return fmt.Errorf("failed to get user by email: %w", err)
		}

		// Convert to domain model
		user = &model.User{
			ID:               dbUser.ID,
			TenantID:         dbUser.TenantID,
			Email:            dbUser.Email,
			PasswordHash:     getStringValue(dbUser.PasswordHash),
			FirstName:        "",  // Will come from Person entity
			LastName:         "",  // Will come from Person entity
			PhoneNumber:      nil, // Will come from Person entity
			AccountStatus:    model.UserAccountStatus(getStringValue(dbUser.AccountStatus)),
			EmailVerified:    false, // Default value
			PhoneVerified:    false, // Default value
			MFAEnabled:       getBoolValue(dbUser.MfaEnabled),
			FailedLoginCount: int(getInt32Value(dbUser.FailedLoginAttempts)),
			Metadata:         make(map[string]any),
			CreatedAt:        dbUser.CreatedAt,
			UpdatedAt:        dbUser.UpdatedAt,
		}

		return nil
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("user_repository_get_by_email_error", nil)
		return nil, err
	}

	r.metrics.IncrementCounter("user_repository_get_by_email_success", nil)
	return user, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *model.User) (*model.User, error) {
	ctx, span := r.tracer.StartSpan(ctx, "user_repository.Update")
	defer span.End()

	var updatedUser *model.User
	err := r.store.WithTenant(ctx, user.TenantID, func(ctx context.Context, store db.Store) error {
		// Convert domain model to SQLC params
		params := db.UpdateUserParams{
			ID:                    user.ID,
			Username:              nil, // Username updates handled separately
			Email:                 &user.Email,
			UserType:              nil, // UserType updates handled separately
			AccountStatus:         stringPtr(string(user.AccountStatus)),
			SessionTimeoutMinutes: nil, // Session timeout handled separately
			MfaEnabled:            &user.MFAEnabled,
		}

		dbUser, err := store.UpdateUser(ctx, params)
		if err != nil {
			if err == db.ErrNoRows {
				return errors.NewBusinessError("USER_NOT_FOUND", "User not found")
			}
			return fmt.Errorf("failed to update user: %w", err)
		}

		// Convert back to domain model
		updatedUser = &model.User{
			ID:               dbUser.ID,
			TenantID:         dbUser.TenantID,
			Email:            dbUser.Email,
			PasswordHash:     getStringValue(dbUser.PasswordHash),
			FirstName:        "",  // These come from Person entity
			LastName:         "",  // These come from Person entity
			PhoneNumber:      nil, // These come from Person entity
			AccountStatus:    model.UserAccountStatus(getStringValue(dbUser.AccountStatus)),
			EmailVerified:    false, // Default values
			PhoneVerified:    false, // Default values
			MFAEnabled:       getBoolValue(dbUser.MfaEnabled),
			FailedLoginCount: int(getInt32Value(dbUser.FailedLoginAttempts)),
			Metadata:         user.Metadata, // Preserve original metadata
			CreatedAt:        dbUser.CreatedAt,
			UpdatedAt:        dbUser.UpdatedAt,
		}

		return nil
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("user_repository_update_error", nil)
		return nil, err
	}

	r.metrics.IncrementCounter("user_repository_update_success", nil)
	r.logger.InfoContext(ctx, "User updated successfully",
		logger.Fields{"user_id": user.ID})

	return updatedUser, nil
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.StartSpan(ctx, "user_repository.Delete")
	defer span.End()

	tenantID := getTenantIDFromContext(ctx)

	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, store db.Store) error {
		err := store.SoftDeleteUser(ctx, id)
		if err != nil {
			if err == db.ErrNoRows {
				return errors.NewBusinessError("USER_NOT_FOUND", "User not found")
			}
			return fmt.Errorf("failed to delete user: %w", err)
		}

		return nil
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("user_repository_delete_error", nil)
		return err
	}

	r.metrics.IncrementCounter("user_repository_delete_success", nil)
	r.logger.InfoContext(ctx, "User deleted successfully",
		logger.Fields{"user_id": id})

	return nil
}

// List lists users with pagination
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	ctx, span := r.tracer.StartSpan(ctx, "user_repository.List")
	defer span.End()

	tenantID := getTenantIDFromContext(ctx)

	var users []*model.User
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, store db.Store) error {
		// Validate bounds to prevent integer overflow
		if limit > 2147483647 || limit < 0 {
			return errors.NewBusinessErrorWithContext(ctx, "INVALID_LIMIT", "Limit value out of valid range")
		}
		if offset > 2147483647 || offset < 0 {
			return errors.NewBusinessErrorWithContext(ctx, "INVALID_OFFSET", "Offset value out of valid range")
		}

		// #nosec G115 - Safe conversion after bounds check
		params := db.ListUsersParams{
			Limit:         int32(limit),
			Offset:        int32(offset),
			UserType:      "", // Empty string for no filter
			AccountStatus: "", // Empty string for no filter
		}

		dbUsers, err := store.ListUsers(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}

		// Convert to domain models
		users = make([]*model.User, len(dbUsers))
		for i, dbUser := range dbUsers {
			users[i] = &model.User{
				ID:               dbUser.ID,
				TenantID:         dbUser.TenantID,
				Email:            dbUser.Email,
				PasswordHash:     getStringValue(dbUser.PasswordHash),
				FirstName:        "",  // Will come from Person entity
				LastName:         "",  // Will come from Person entity
				PhoneNumber:      nil, // Will come from Person entity
				AccountStatus:    model.UserAccountStatus(getStringValue(dbUser.AccountStatus)),
				EmailVerified:    false, // Default value
				PhoneVerified:    false, // Default value
				MFAEnabled:       getBoolValue(dbUser.MfaEnabled),
				FailedLoginCount: int(getInt32Value(dbUser.FailedLoginAttempts)),
				Metadata:         make(map[string]any),
				CreatedAt:        dbUser.CreatedAt,
				UpdatedAt:        dbUser.UpdatedAt,
			}
		}

		return nil
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, err
	}

	return users, nil
}

// Helper function implementations for remaining methods would follow the same pattern...

// GetPasswordHash retrieves user's password hash
func (r *userRepository) GetPasswordHash(ctx context.Context, userID uuid.UUID) (string, error) {
	ctx, span := r.tracer.StartSpan(ctx, "user_repository.GetPasswordHash")
	defer span.End()

	tenantID := getTenantIDFromContext(ctx)

	var hash string
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, store db.Store) error {
		var err error
		hashPtr, err := store.GetUserPasswordByID(ctx, userID)
		if err != nil {
			if err == db.ErrNoRows {
				return errors.NewBusinessError("USER_NOT_FOUND", "User not found")
			}
			return fmt.Errorf("failed to get password hash: %w", err)
		}
		hash = getStringValue(hashPtr)
		return nil
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return "", err
	}

	return hash, nil
}

// UpdatePasswordHash updates user's password hash
func (r *userRepository) UpdatePasswordHash(ctx context.Context, userID uuid.UUID, hash string) error {
	ctx, span := r.tracer.StartSpan(ctx, "user_repository.UpdatePasswordHash")
	defer span.End()

	tenantID := getTenantIDFromContext(ctx)

	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, store db.Store) error {
		params := db.UpdateUserPasswordParams{
			ID:           userID,
			PasswordHash: &hash,
		}

		err := store.UpdateUserPassword(ctx, params)
		if err != nil {
			if err == db.ErrNoRows {
				return errors.NewBusinessError("USER_NOT_FOUND", "User not found")
			}
			return fmt.Errorf("failed to update password hash: %w", err)
		}
		return nil
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return err
	}

	return nil
}

// Helper functions
func getTenantIDFromContext(ctx context.Context) uuid.UUID {
	// Extract tenant ID from context - this would be implemented properly
	// to get tenant from context set by middleware or service layer
	if tenantID, ok := ctx.Value("tenant_id").(uuid.UUID); ok {
		return tenantID
	}
	// This is a temporary placeholder - in production this should return an error
	return uuid.Nil
}

func isDuplicateKeyError(err error) bool {
	// Check if error is a duplicate key violation
	// Implementation would check against database-specific error codes
	return false
}

// Helper functions for pointer conversions
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func getBoolValue(ptr *bool) bool {
	if ptr == nil {
		return false
	}
	return *ptr
}

func getInt32Value(ptr *int32) int32 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

// ListByStatus lists users by account status
func (r *userRepository) ListByStatus(ctx context.Context, status model.UserAccountStatus) ([]*model.User, error) {
	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("ListByStatus not implemented")
}

// Count returns total number of users
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	// TODO: Implement when SQLC query is available
	return 0, fmt.Errorf("Count not implemented")
}

// UpdateLastLogin updates user's last login timestamp
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, loginTime time.Time) error {
	// TODO: Implement when SQLC query is available
	return fmt.Errorf("UpdateLastLogin not implemented")
}

// LockAccount locks a user account
func (r *userRepository) LockAccount(ctx context.Context, userID uuid.UUID, lockedUntil *time.Time, reason string) error {
	// TODO: Implement when SQLC query is available
	return fmt.Errorf("LockAccount not implemented")
}

// UnlockAccount unlocks a user account
func (r *userRepository) UnlockAccount(ctx context.Context, userID uuid.UUID) error {
	// TODO: Implement when SQLC query is available
	return fmt.Errorf("UnlockAccount not implemented")
}

// UpdateFailedLoginCount updates failed login attempt count
func (r *userRepository) UpdateFailedLoginCount(ctx context.Context, userID uuid.UUID, count int) error {
	// TODO: Implement when SQLC query is available
	return fmt.Errorf("UpdateFailedLoginCount not implemented")
}

// SetMFASecret sets MFA secret for user
func (r *userRepository) SetMFASecret(ctx context.Context, userID uuid.UUID, secret string) error {
	// TODO: Implement when SQLC query is available
	return fmt.Errorf("SetMFASecret not implemented")
}

// GetMFASecret retrieves MFA secret for user
func (r *userRepository) GetMFASecret(ctx context.Context, userID uuid.UUID) (string, error) {
	// TODO: Implement when SQLC query is available
	return "", fmt.Errorf("GetMFASecret not implemented")
}

// EnableMFA enables MFA for user
func (r *userRepository) EnableMFA(ctx context.Context, userID uuid.UUID, method model.MFAMethod) error {
	// TODO: Implement when SQLC query is available
	return fmt.Errorf("EnableMFA not implemented")
}

// DisableMFA disables MFA for user
func (r *userRepository) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	// TODO: Implement when SQLC query is available
	return fmt.Errorf("DisableMFA not implemented")
}
