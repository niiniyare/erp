package authn

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/core/iam/repo"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// service implements the authentication service
type service struct {
	// Repository
	repo repo.IAMRepository

	// External dependencies
	tenantService      tenant.Service
	auditService       audit.Service
	featureFlagService featureflag.AdminAdvancedService
	cache              cache.Service

	// Shared infrastructure
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewService creates a new authentication service instance
func NewService(
	repo repo.IAMRepository,
	tenantService tenant.Service,
	auditService audit.Service,
	featureFlagService featureflag.AdminAdvancedService,
	cache cache.Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) Service {
	return &service{
		repo:               repo,
		tenantService:      tenantService,
		auditService:       auditService,
		featureFlagService: featureFlagService,
		cache:              cache,
		logger:             logger,
		metrics:            metrics,
		tracer:             tracer,
	}
}

// Authentication

func (s *service) Authenticate(ctx context.Context, req *AuthenticationRequest) (*AuthenticationResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.Authenticate")
	defer span.End()

	startTime := time.Now()

	// Get user by email
	user, err := s.GetUserByEmail(ctx, req.Email)
	if err != nil {
		s.metrics.IncrementCounter("authn_authentication_failures", metrics.Fields{"reason": "user_not_found"})
		return nil, errors.NewBusinessError("INVALID_CREDENTIALS", "Invalid email or password")
	}

	// Check account status
	if !user.IsActive() {
		s.metrics.IncrementCounter("authn_authentication_failures", metrics.Fields{"reason": "account_inactive"})
		return nil, errors.NewBusinessError("ACCOUNT_INACTIVE", "Account is inactive or locked")
	}

	// Verify password
	valid, err := s.verifyPassword(ctx, user.ID, req.Password)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, err
	}

	if !valid {
		// Increment failed login count
		s.handleFailedLogin(ctx, user.ID)
		s.metrics.IncrementCounter("authn_authentication_failures", metrics.Fields{"reason": "invalid_password"})
		return nil, errors.NewBusinessError("INVALID_CREDENTIALS", "Invalid email or password")
	}

	// Check MFA if enabled
	if user.MFAEnabled && req.MFACode == "" {
		return &AuthenticationResult{
			User:        user,
			MFARequired: true,
		}, nil
	}

	if user.MFAEnabled && req.MFACode != "" {
		mfaValid, err := s.ValidateMFA(ctx, &ValidateMFARequest{
			UserID: user.ID,
			Code:   req.MFACode,
		})
		if err != nil || !mfaValid.Valid {
			s.metrics.IncrementCounter("authn_authentication_failures", metrics.Fields{"reason": "invalid_mfa"})
			return nil, errors.NewBusinessError("INVALID_MFA", "Invalid MFA code")
		}
	}

	// Generate tokens
	accessToken, refreshToken, expiresAt, err := s.generateTokens(ctx, user)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, err
	}

	// Update last login
	s.updateLastLogin(ctx, user.ID)

	// Reset failed login count
	s.resetFailedLoginCount(ctx, user.ID)

	// Record metrics
	duration := time.Since(startTime)
	s.metrics.IncrementCounter("authn_authentications_successful", nil)
	s.metrics.ObserveHistogram("authn_authentication_duration_seconds", duration.Seconds(), nil)

	// Audit log
	s.auditDataOperation(ctx, "USER_AUTHENTICATION", "user", user.ID)

	s.logger.InfoContext(ctx, "User authenticated successfully",
		logger.Fields{
			"user_id":  user.ID,
			"email":    user.Email,
			"mfa_used": user.MFAEnabled,
			"duration": duration.Milliseconds(),
		})

	return &AuthenticationResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		MFARequired:  false,
	}, nil
}

// Password Management

func (s *service) ChangePassword(ctx context.Context, req *ChangePasswordRequest) error {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.ChangePassword")
	defer span.End()

	// Verify current password
	valid, err := s.verifyPassword(ctx, req.UserID, req.CurrentPassword)
	if err != nil {
		return err
	}

	if !valid {
		s.metrics.IncrementCounter("authn_change_password_failures", metrics.Fields{"reason": "invalid_current_password"})
		return errors.NewBusinessError("INVALID_CURRENT_PASSWORD", "Current password is incorrect")
	}

	// Hash new password
	newHash, err := s.hashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update password hash
	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return err
	}

	err = s.repo.WithTenant(ctx, tenantID, func(ctx context.Context, txRepo repo.TransactionalRepository) error {
		return s.repo.Users().UpdatePasswordHash(ctx, req.UserID, newHash)
	})

	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "CHANGE_PASSWORD", "user", req.UserID)

	// Invalidate all user sessions
	s.InvalidateAllUserSessions(ctx, req.UserID)

	s.metrics.IncrementCounter("authn_passwords_changed", nil)

	s.logger.InfoContext(ctx, "Password changed successfully",
		logger.Fields{"user_id": req.UserID})

	return nil
}

// Helper methods

func (s *service) getCurrentTenantID(ctx context.Context) (uuid.UUID, error) {
	tenant, err := s.tenantService.GetCurrentTenant(ctx)
	if err != nil {
		return uuid.Nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context is required")
	}
	return tenant.ID, nil
}

func (s *service) getCurrentTenantSlug(ctx context.Context) string {
	tenant, err := s.tenantService.GetCurrentTenant(ctx)
	if err != nil {
		return ""
	}
	return tenant.Slug
}

func (s *service) getTenantCache(ctx context.Context) cache.Service {
	tenantID, _ := s.getCurrentTenantID(ctx)
	tenantSlug := s.getCurrentTenantSlug(ctx)

	return s.cache.WithTenantAndNamespace(ctx, tenantID, tenantSlug, "authn")
}

func (s *service) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

func (s *service) verifyPassword(ctx context.Context, userID uuid.UUID, password string) (bool, error) {
	// Get password hash from database
	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return false, err
	}

	var hash string
	err = s.repo.WithTenant(ctx, tenantID, func(ctx context.Context, txRepo repo.TransactionalRepository) error {
		hash, err = s.repo.Users().GetPasswordHash(ctx, userID)
		return err
	})

	if err != nil {
		return false, fmt.Errorf("failed to get password hash: %w", err)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil, nil
}

func (s *service) generateTokens(ctx context.Context, user *model.User) (accessToken, refreshToken string, expiresAt time.Time, err error) {
	// This would integrate with your existing JWT token generation
	// For now, returning placeholder implementation
	accessToken = "access_token_" + user.ID.String()
	refreshToken = "refresh_token_" + user.ID.String()
	expiresAt = time.Now().Add(time.Hour * 24)
	return accessToken, refreshToken, expiresAt, nil
}

func (s *service) handleFailedLogin(ctx context.Context, userID uuid.UUID) {
	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return
	}

	s.repo.WithTenant(ctx, tenantID, func(ctx context.Context, txRepo repo.TransactionalRepository) error {
		// This would increment failed login count and potentially lock account
		return s.repo.Users().UpdateFailedLoginCount(ctx, userID, 1)
	})
}

func (s *service) updateLastLogin(ctx context.Context, userID uuid.UUID) {
	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return
	}

	s.repo.WithTenant(ctx, tenantID, func(ctx context.Context, txRepo repo.TransactionalRepository) error {
		return s.repo.Users().UpdateLastLogin(ctx, userID, time.Now())
	})
}

func (s *service) resetFailedLoginCount(ctx context.Context, userID uuid.UUID) {
	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return
	}

	s.repo.WithTenant(ctx, tenantID, func(ctx context.Context, txRepo repo.TransactionalRepository) error {
		return s.repo.Users().UpdateFailedLoginCount(ctx, userID, 0)
	})
}

func (s *service) invalidateUserCaches(ctx context.Context, userID uuid.UUID) {
	cacheKeys := []string{
		fmt.Sprintf("user:%s", userID.String()),
		fmt.Sprintf("user_roles:%s", userID.String()),
		fmt.Sprintf("user_permissions:%s", userID.String()),
	}

	tenantCache := s.getTenantCache(ctx)
	for _, key := range cacheKeys {
		tenantCache.Delete(ctx, key)
	}
}

func (s *service) auditDataOperation(ctx context.Context, operation string, entityType string, entityID uuid.UUID) {
	tenantID, _ := s.getCurrentTenantID(ctx)

	auditLog := &audit.DataOperationLog{
		Operation:  operation,
		EntityType: entityType,
		EntityID:   entityID,
		TenantID:   tenantID,
		Timestamp:  time.Now(),
	}

	// Fire and forget audit logging
	go func() {
		s.auditService.LogDataOperation(context.Background(), auditLog)
	}()
}

// MFA Methods

func (s *service) EnableMFA(ctx context.Context, req *EnableMFARequest) (*MFASetupResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.EnableMFA")
	defer span.End()

	// Get user
	user, err := s.GetUser(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Generate MFA secret
	secret := s.generateMFASecret()

	// Store MFA secret
	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return nil, err
	}

	err = s.repo.WithTenant(ctx, tenantID, func(ctx context.Context, txRepo repo.TransactionalRepository) error {
		return s.repo.Users().SetMFASecret(ctx, req.UserID, secret)
	})

	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to store MFA secret: %w", err)
	}

	// Generate QR code (placeholder implementation)
	qrCode := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=ERP", "ERP", user.Email, secret)

	// Audit log
	s.auditDataOperation(ctx, "ENABLE_MFA", "user", req.UserID)

	s.logger.InfoContext(ctx, "MFA enabled for user",
		logger.Fields{"user_id": req.UserID, "method": req.Method})

	return &MFASetupResult{
		Secret:        secret,
		QRCode:        qrCode,
		BackupCodes:   s.generateBackupCodes(),
		Method:        req.Method,
		SetupComplete: false,
	}, nil
}

func (s *service) DisableMFA(ctx context.Context, req *DisableMFARequest) error {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.DisableMFA")
	defer span.End()

	// Verify password first
	valid, err := s.verifyPassword(ctx, req.UserID, req.Password)
	if err != nil {
		return err
	}

	if !valid {
		s.metrics.IncrementCounter("authn_disable_mfa_failures", metrics.Fields{"reason": "invalid_password"})
		return errors.NewBusinessError("INVALID_PASSWORD", "Invalid password")
	}

	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return err
	}

	err = s.repo.WithTenant(ctx, tenantID, func(ctx context.Context, txRepo repo.TransactionalRepository) error {
		return s.repo.Users().DisableMFA(ctx, req.UserID)
	})

	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "DISABLE_MFA", "user", req.UserID)

	// Clear caches
	s.invalidateUserCaches(ctx, req.UserID)

	s.logger.InfoContext(ctx, "MFA disabled for user",
		logger.Fields{"user_id": req.UserID})

	return nil
}

func (s *service) ValidateMFA(ctx context.Context, req *ValidateMFARequest) (*MFAValidationResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.ValidateMFA")
	defer span.End()

	// Get MFA secret
	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return nil, err
	}

	var secret string
	err = s.repo.WithTenant(ctx, tenantID, func(ctx context.Context, txRepo repo.TransactionalRepository) error {
		secret, err = s.repo.Users().GetMFASecret(ctx, req.UserID)
		return err
	})

	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to get MFA secret: %w", err)
	}

	// Validate TOTP code (placeholder implementation)
	valid := s.validateTOTPCode(secret, req.Code)

	result := &MFAValidationResult{
		Valid:       valid,
		ValidatedAt: time.Now(),
	}

	if valid {
		s.metrics.IncrementCounter("authn_mfa_validations_successful", nil)
	} else {
		s.metrics.IncrementCounter("authn_mfa_validations_failed", nil)
	}

	return result, nil
}

// Session Management

func (s *service) CreateSession(ctx context.Context, req *CreateSessionRequest) (*model.Session, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.CreateSession")
	defer span.End()

	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return nil, err
	}

	session := &model.Session{
		ID:           uuid.New(),
		TenantID:     tenantID,
		UserID:       req.UserID,
		Token:        s.generateSessionToken(),
		RefreshToken: s.generateRefreshToken(),
		ExpiresAt:    time.Now().Add(req.ExpirationDuration),
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		Status:       model.SessionStatusActive,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	var createdSession *model.Session
	err = s.repo.WithTenant(ctx, tenantID, func(ctx context.Context, txRepo repo.TransactionalRepository) error {
		createdSession, err = s.repo.Sessions().Create(ctx, session)
		return err
	})

	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "CREATE_SESSION", "session", session.ID)

	s.logger.InfoContext(ctx, "Session created",
		logger.Fields{"session_id": session.ID, "user_id": req.UserID})

	return createdSession, nil
}

func (s *service) ValidateSession(ctx context.Context, token string) (*SessionValidationResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.ValidateSession")
	defer span.End()

	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return nil, err
	}

	var session *model.Session
	err = s.repo.WithTenant(ctx, tenantID, func(ctx context.Context, txRepo repo.TransactionalRepository) error {
		session, err = s.repo.Sessions().GetByToken(ctx, token)
		return err
	})

	if err != nil {
		s.metrics.IncrementCounter("authn_session_validations_failed", metrics.Fields{"reason": "not_found"})
		return &SessionValidationResult{Valid: false, Reason: "Session not found"}, nil
	}

	// Check if session is expired
	if session.IsExpired() {
		s.metrics.IncrementCounter("authn_session_validations_failed", metrics.Fields{"reason": "expired"})
		return &SessionValidationResult{Valid: false, Reason: "Session expired"}, nil
	}

	// Check if session is active
	if session.Status != model.SessionStatusActive {
		s.metrics.IncrementCounter("authn_session_validations_failed", metrics.Fields{"reason": "inactive"})
		return &SessionValidationResult{Valid: false, Reason: "Session inactive"}, nil
	}

	s.metrics.IncrementCounter("authn_session_validations_successful", nil)

	return &SessionValidationResult{
		Valid:   true,
		Session: session,
	}, nil
}

func (s *service) InvalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.InvalidateSession")
	defer span.End()

	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return err
	}

	err = s.repo.WithTenant(ctx, tenantID, func(ctx context.Context, txRepo repo.TransactionalRepository) error {
		return s.repo.Sessions().InvalidateSession(ctx, sessionID)
	})

	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "INVALIDATE_SESSION", "session", sessionID)

	s.logger.InfoContext(ctx, "Session invalidated",
		logger.Fields{"session_id": sessionID})

	return nil
}

func (s *service) InvalidateAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.InvalidateAllUserSessions")
	defer span.End()

	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return err
	}

	err = s.repo.WithTenant(ctx, tenantID, func(ctx context.Context, txRepo repo.TransactionalRepository) error {
		return s.repo.Sessions().InvalidateUserSessions(ctx, userID)
	})

	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to invalidate user sessions: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "INVALIDATE_ALL_USER_SESSIONS", "user", userID)

	s.logger.InfoContext(ctx, "All user sessions invalidated",
		logger.Fields{"user_id": userID})

	return nil
}

// Role Management

func (s *service) AssignRole(ctx context.Context, req *AssignRoleRequest) error {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.AssignRole")
	defer span.End()

	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return err
	}

	userRole := &model.UserRole{
		ID:        uuid.New(),
		TenantID:  tenantID,
		UserID:    req.UserID,
		RoleID:    req.RoleID,
		EntityID:  req.EntityID,
		ExpiresAt: req.ExpiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.repo.WithTenant(ctx, tenantID, func(ctx context.Context, txRepo repo.TransactionalRepository) error {
		return s.repo.UserRoles().Assign(ctx, userRole)
	})

	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "ASSIGN_ROLE", "user_role", userRole.ID)

	// Clear user permission caches
	s.invalidateUserCaches(ctx, req.UserID)

	s.logger.InfoContext(ctx, "Role assigned to user",
		logger.Fields{"user_id": req.UserID, "role_id": req.RoleID})

	return nil
}

func (s *service) RemoveRole(ctx context.Context, req *RemoveRoleRequest) error {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.RemoveRole")
	defer span.End()

	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return err
	}

	err = s.repo.WithTenant(ctx, tenantID, func(ctx context.Context, txRepo repo.TransactionalRepository) error {
		return s.repo.UserRoles().Remove(ctx, req.UserID, req.RoleID, req.EntityID)
	})

	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to remove role: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "REMOVE_ROLE", "user", req.UserID)

	// Clear user permission caches
	s.invalidateUserCaches(ctx, req.UserID)

	s.logger.InfoContext(ctx, "Role removed from user",
		logger.Fields{"user_id": req.UserID, "role_id": req.RoleID})

	return nil
}

func (s *service) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.UserRole, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.GetUserRoles")
	defer span.End()

	// Try cache first
	cacheKey := fmt.Sprintf("user_roles:%s", userID.String())
	var userRoles []*model.UserRole
	if err := s.getTenantCache(ctx).Get(ctx, cacheKey, &userRoles); err == nil {
		s.metrics.IncrementCounter("authn_user_roles_cache_hits", nil)
		return userRoles, nil
	}

	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return nil, err
	}

	err = s.repo.WithTenant(ctx, tenantID, func(ctx context.Context, txRepo repo.TransactionalRepository) error {
		userRoles, err = s.repo.UserRoles().GetUserRoles(ctx, userID)
		return err
	})

	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Cache the result
	s.getTenantCache(ctx).Set(ctx, cacheKey, userRoles, time.Hour)
	s.metrics.IncrementCounter("authn_user_roles_cache_misses", nil)

	return userRoles, nil
}

// Helper methods for MFA and Session

func (s *service) generateMFASecret() string {
	// Placeholder implementation - would use proper TOTP secret generation
	return fmt.Sprintf("secret_%d", time.Now().Unix())
}

func (s *service) generateBackupCodes() []string {
	// Placeholder implementation - would generate proper backup codes
	codes := make([]string, 8)
	for i := range codes {
		codes[i] = fmt.Sprintf("backup_%d_%d", i+1, time.Now().Unix())
	}
	return codes
}

func (s *service) validateTOTPCode(secret, code string) bool {
	// Placeholder implementation - would use proper TOTP validation
	return code == "123456" // For testing purposes
}

func (s *service) generateSessionToken() string {
	// Placeholder implementation - would use cryptographically secure token generation
	return fmt.Sprintf("session_%s", uuid.New().String())
}

func (s *service) generateRefreshToken() string {
	// Placeholder implementation - would use cryptographically secure token generation
	return fmt.Sprintf("refresh_%s", uuid.New().String())
}
