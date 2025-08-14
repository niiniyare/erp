package repo

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/iam/model"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/errors"
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
		// Convert domain model to SQLC params
		params := db.CreateUserParams{
			ID:               user.ID,
			TenantID:         user.TenantID,
			Email:            user.Email,
			PasswordHash:     user.PasswordHash,
			FirstName:        user.FirstName,
			LastName:         user.LastName,
			PhoneNumber:      user.PhoneNumber,
			AccountStatus:    string(user.AccountStatus),
			EmailVerified:    user.EmailVerified,
			PhoneVerified:    user.PhoneVerified,
			MfaEnabled:       user.MFAEnabled,
			FailedLoginCount: int32(user.FailedLoginCount),
			Metadata:         user.Metadata,
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
			PasswordHash:     dbUser.PasswordHash,
			FirstName:        dbUser.FirstName,
			LastName:         dbUser.LastName,
			PhoneNumber:      dbUser.PhoneNumber,
			AccountStatus:    model.UserAccountStatus(dbUser.AccountStatus),
			EmailVerified:    dbUser.EmailVerified,
			PhoneVerified:    dbUser.PhoneVerified,
			MFAEnabled:       dbUser.MfaEnabled,
			FailedLoginCount: int(dbUser.FailedLoginCount),
			Metadata:         dbUser.Metadata,
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
			PasswordHash:     dbUser.PasswordHash,
			FirstName:        dbUser.FirstName,
			LastName:         dbUser.LastName,
			PhoneNumber:      dbUser.PhoneNumber,
			AccountStatus:    model.UserAccountStatus(dbUser.AccountStatus),
			EmailVerified:    dbUser.EmailVerified,
			PhoneVerified:    dbUser.PhoneVerified,
			MFAEnabled:       dbUser.MfaEnabled,
			FailedLoginCount: int(dbUser.FailedLoginCount),
			Metadata:         dbUser.Metadata,
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
			PasswordHash:     dbUser.PasswordHash,
			FirstName:        dbUser.FirstName,
			LastName:         dbUser.LastName,
			PhoneNumber:      dbUser.PhoneNumber,
			AccountStatus:    model.UserAccountStatus(dbUser.AccountStatus),
			EmailVerified:    dbUser.EmailVerified,
			PhoneVerified:    dbUser.PhoneVerified,
			MFAEnabled:       dbUser.MfaEnabled,
			FailedLoginCount: int(dbUser.FailedLoginCount),
			Metadata:         dbUser.Metadata,
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
			ID:               user.ID,
			FirstName:        user.FirstName,
			LastName:         user.LastName,
			PhoneNumber:      user.PhoneNumber,
			AccountStatus:    string(user.AccountStatus),
			EmailVerified:    user.EmailVerified,
			PhoneVerified:    user.PhoneVerified,
			MfaEnabled:       user.MFAEnabled,
			FailedLoginCount: int32(user.FailedLoginCount),
			Metadata:         user.Metadata,
			UpdatedAt:        time.Now(),
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
			PasswordHash:     dbUser.PasswordHash,
			FirstName:        dbUser.FirstName,
			LastName:         dbUser.LastName,
			PhoneNumber:      dbUser.PhoneNumber,
			AccountStatus:    model.UserAccountStatus(dbUser.AccountStatus),
			EmailVerified:    dbUser.EmailVerified,
			PhoneVerified:    dbUser.PhoneVerified,
			MFAEnabled:       dbUser.MfaEnabled,
			FailedLoginCount: int(dbUser.FailedLoginCount),
			Metadata:         dbUser.Metadata,
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
		err := store.DeleteUser(ctx, id)
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
		params := db.ListUsersParams{
			Limit:  int32(limit),
			Offset: int32(offset),
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
				PasswordHash:     dbUser.PasswordHash,
				FirstName:        dbUser.FirstName,
				LastName:         dbUser.LastName,
				PhoneNumber:      dbUser.PhoneNumber,
				AccountStatus:    model.UserAccountStatus(dbUser.AccountStatus),
				EmailVerified:    dbUser.EmailVerified,
				PhoneVerified:    dbUser.PhoneVerified,
				MFAEnabled:       dbUser.MfaEnabled,
				FailedLoginCount: int(dbUser.FailedLoginCount),
				Metadata:         dbUser.Metadata,
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
		hash, err = store.GetUserPasswordHash(ctx, userID)
		if err != nil {
			if err == db.ErrNoRows {
				return errors.NewBusinessError("USER_NOT_FOUND", "User not found")
			}
			return fmt.Errorf("failed to get password hash: %w", err)
		}
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
			PasswordHash: hash,
			UpdatedAt:    time.Now(),
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

// Additional methods would be implemented following the same pattern:
// - ListByStatus
// - Count
// - UpdateLastLogin
// - LockAccount
// - UnlockAccount
// - UpdateFailedLoginCount
// - SetMFASecret
// - GetMFASecret
// - EnableMFA
// - DisableMFA