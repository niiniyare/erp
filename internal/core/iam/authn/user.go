package authn

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/core/iam/repo"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// User Management

func (s *service) CreateUser(ctx context.Context, req *CreateUserRequest) (*model.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.CreateUser")
	defer span.End()

	// Get current tenant
	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return nil, err
	}

	// Hash password
	passwordHash, err := s.hashPassword(req.Password)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessError("PASSWORD_HASH_FAILED", "Failed to hash password")
	}

	// Create user model
	user := &model.User{
		ID:               uuid.New(),
		TenantID:         tenantID,
		Email:            req.Email,
		PasswordHash:     passwordHash,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		PhoneNumber:      req.PhoneNumber,
		AccountStatus:    model.UserAccountStatusActive,
		EmailVerified:    false,
		PhoneVerified:    false,
		MFAEnabled:       false,
		FailedLoginCount: 0,
		Metadata:         req.Metadata,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Create user in database
	createdUser, err := s.repo.Users().Create(ctx, user)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		s.metrics.IncrementCounter("authn_create_user_errors", metrics.Fields{"reason": "database_error"})
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "CREATE_USER", "user", user.ID)

	// Clear relevant caches
	s.invalidateUserCaches(ctx, user.ID)

	// Update metrics
	s.metrics.IncrementCounter("authn_users_created", nil)

	s.logger.InfoContext(ctx, "User created successfully",
		logger.Fields{
			"user_id":   user.ID,
			"email":     user.Email,
			"tenant_id": tenantID,
		})

	return createdUser, nil
}

func (s *service) GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.GetUser")
	defer span.End()

	// Try cache first
	cacheKey := fmt.Sprintf("user:%s", userID.String())
	var user model.User
	if err := s.getTenantCache(ctx).Get(ctx, cacheKey, &user); err == nil {
		s.metrics.IncrementCounter("authn_user_cache_hits", nil)
		return &user, nil
	}

	// Cache miss - get from database
	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return nil, err
	}

	foundUser, err := s.repo.Users().GetByID(ctx, userID)

	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		s.metrics.IncrementCounter("authn_get_user_errors", metrics.Fields{"reason": "not_found"})
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Cache the result
	s.getTenantCache(ctx).Set(ctx, cacheKey, foundUser, time.Hour)
	s.metrics.IncrementCounter("authn_user_cache_misses", nil)

	return foundUser, nil
}

func (s *service) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.GetUserByEmail")
	defer span.End()

	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.Users().GetByEmail(ctx, email)

	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

func (s *service) UpdateUser(ctx context.Context, req *UpdateUserRequest) (*model.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.UpdateUser")
	defer span.End()

	// Get existing user
	user, err := s.GetUser(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.PhoneNumber != nil {
		user.PhoneNumber = req.PhoneNumber
	}
	if req.Metadata != nil {
		user.Metadata = req.Metadata
	}
	user.UpdatedAt = time.Now()

	// Update in database
	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return nil, err
	}

	updatedUser, err := s.repo.Users().Update(ctx, user)

	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "UPDATE_USER", "user", user.ID)

	// Clear caches
	s.invalidateUserCaches(ctx, user.ID)

	s.logger.InfoContext(ctx, "User updated successfully",
		logger.Fields{"user_id": user.ID})

	return updatedUser, nil
}

func (s *service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.DeleteUser")
	defer span.End()

	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return err
	}

	err := s.repo.Users().Delete(ctx, userID)

	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "DELETE_USER", "user", userID)

	// Clear caches
	s.invalidateUserCaches(ctx, userID)

	s.logger.InfoContext(ctx, "User deleted successfully",
		logger.Fields{"user_id": userID})

	return nil
}
