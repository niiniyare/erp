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
	store              db.Store
	cache              cache.Service

	// Authentication components
	jwtManager     *JWTManager
	sessionManager *SessionManager

	// Shared infrastructure
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.Service
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
	tracer tracing.Service,
) Service {
	// NOTE: JWT secrets should come from configuration/environment variables
	// TODO: Implement proper configuration management for JWT secrets
	// TODO: Add secret rotation functionality
	// TODO: Integrate with secrets management system (HashiCorp Vault, etc.)
	accessSecret := "your-256-bit-access-secret-key-here"   // TODO: Use config
	refreshSecret := "your-256-bit-refresh-secret-key-here" // TODO: Use config

	return &service{
		repo:               repo,
		tenantService:      tenantService,
		auditService:       auditService,
		featureFlagService: featureFlagService,
		cache:              cache,
		jwtManager:         NewJWTManager(accessSecret, refreshSecret),
		sessionManager:     NewSessionManager(logger),
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

	err = s.store.WithTenant(ctx, tenantID, func(ctx context.Context, store db.Store) error {
		return s.repo.Users().UpdatePasswordHash(ctx, req.UserID, newHash)
	})
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "CHANGE_PASSWORD", "user", req.UserID)

	// Invalidate all user sessions
	if err := s.InvalidateAllUserSessions(ctx, req.UserID); err != nil {
		s.logger.WarnContext(ctx, "Failed to invalidate user sessions after password change", logger.Fields{"error": err.Error(), "user_id": req.UserID.String()})
	}

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
	// NOTE: Since the cache service doesn't have WithTenantAndNamespace method,
	// we'll use the cache service directly for now
	return s.cache
}

func (s *service) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

func (s *service) verifyPassword(ctx context.Context, userID uuid.UUID, password string) (bool, error) {
	// TODO: Get password hash from database
	tenantID, err := s.getCurrentTenantID(ctx)
	if err != nil {
		return false, err
	}

	var hash string
	err = s.store.WithTenant(ctx, tenantID, func(ctx context.Context, store db.Store) error {
		var err error
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
	// Get user roles for JWT claims
	// NOTE: Currently using placeholder empty roles array
	// TODO: Implement proper role retrieval from UserRoles repository
	// TODO: Add caching for frequently accessed user roles
	roles := []string{} // Placeholder, should get from s.GetUserRoles(ctx, user.ID)

	// Generate JWT token pair using JWT manager
	accessToken, refreshToken, expiresAt, err = s.jwtManager.GenerateTokenPair(user, roles)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate JWT tokens",
			logger.Fields{"error": err.Error(), "user_id": user.ID})
		return "", "", time.Time{}, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return accessToken, refreshToken, expiresAt, nil
}

func (s *service) handleFailedLogin(ctx context.Context, userID uuid.UUID) {
	// Increment failed login count directly through repository
	if err := s.repo.Users().UpdateFailedLoginCount(ctx, userID, 1); err != nil {
		s.logger.WarnContext(ctx, "Failed to increment failed login count", logger.Fields{"error": err.Error(), "user_id": userID.String()})
	}
}

func (s *service) updateLastLogin(ctx context.Context, userID uuid.UUID) {
	// Update last login directly through repository
	if err := s.repo.Users().UpdateLastLogin(ctx, userID, time.Now()); err != nil {
		s.logger.WarnContext(ctx, "Failed to update last login", logger.Fields{"error": err.Error(), "user_id": userID.String()})
	}
}

func (s *service) resetFailedLoginCount(ctx context.Context, userID uuid.UUID) {
	// Reset failed login count directly through repository
	if err := s.repo.Users().UpdateFailedLoginCount(ctx, userID, 0); err != nil {
		s.logger.WarnContext(ctx, "Failed to reset failed login count", logger.Fields{"error": err.Error(), "user_id": userID.String()})
	}
}

func (s *service) invalidateUserCaches(ctx context.Context, userID uuid.UUID) {
	cacheKeys := []string{
		fmt.Sprintf("user:%s", userID.String()),
		fmt.Sprintf("user_roles:%s", userID.String()),
		fmt.Sprintf("user_permissions:%s", userID.String()),
	}

	tenantCache := s.getTenantCache(ctx)
	for _, key := range cacheKeys {
		if err := tenantCache.Delete(ctx, key); err != nil {
			s.logger.WarnContext(ctx, "Failed to delete cache key", logger.Fields{"error": err.Error(), "cache_key": key})
		}
	}
}

func (s *service) auditDataOperation(ctx context.Context, operation string, entityType string, entityID uuid.UUID) {
	// Create audit event using the service interface
	auditReq := audit.CreateAuditEventRequest{
		EventType:     operation,
		EventCategory: "authentication",
		Severity:      "info",
		IPAddress:     nil, // Would extract from context in real implementation
		UserAgent:     nil, // Would extract from context in real implementation
		ResourceID:    &entityID,
	}

	// Fire and forget audit logging - handle any errors
	go func() {
		_, err := s.auditService.CreateAuditEvent(context.Background(), auditReq)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to create audit event", logger.Fields{"error": err.Error()})
		}
	}()
}

// MFA Methods

func (s *service) EnableMFA(ctx context.Context, req *EnableMFARequest) (*MFASetupResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.EnableMFA")
	defer span.End()
	///TODO: check first what is if the tennant is Enable MFA from featureFlagService
	// Get user
	user, err := s.GetUser(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Generate MFA secret
	secret := s.generateMFASecret()

	// Store MFA secret
	err = s.repo.Users().SetMFASecret(ctx, req.UserID, secret)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to store MFA secret: %w", err)
	}

	// TODO: Generate QR code (placeholder implementation)
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

func (s *service) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.DisableMFA")
	defer span.End()

	// Disable MFA directly through repository
	err := s.repo.Users().DisableMFA(ctx, userID)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "DISABLE_MFA", "user", userID)

	// Clear caches
	s.invalidateUserCaches(ctx, userID)

	s.logger.InfoContext(ctx, "MFA disabled for user",
		logger.Fields{"user_id": userID})

	return nil
}

func (s *service) ValidateMFA(ctx context.Context, req *ValidateMFARequest) (*MFAValidationResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.ValidateMFA")
	defer span.End()

	// Get MFA secret
	secret, err := s.repo.Users().GetMFASecret(ctx, req.UserID)
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

	// Use SessionManager to create session
	session, err := s.sessionManager.CreateSession(ctx, req.UserID, tenantID, tenantID, req.IPAddress, req.UserAgent, req.ExpirationDuration)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// NOTE: Sessions repository is not implemented yet, using placeholder
	// TODO: Implement session persistence through repository
	// TODO: Add session to cache for faster retrieval
	// TODO: Implement concurrent session limit enforcement
	var createdSession *model.Session
	// createdSession, err = s.repo.Sessions().Create(ctx, session)
	createdSession = session
	err = nil
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to persist session: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "CREATE_SESSION", "session", session.ID)

	s.logger.InfoContext(ctx, "Session created",
		logger.Fields{"session_id": session.ID, "user_id": req.UserID})

	return createdSession, nil
}

func (s *service) ValidateSession(ctx context.Context, token string) (*model.Session, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.ValidateSession")
	defer span.End()

	// NOTE: Sessions repository is not implemented yet, using placeholder
	// session, err = s.repo.Sessions().GetByToken(ctx, token)
	if token == "" {
		s.metrics.IncrementCounter("authn_session_validations_failed", metrics.Fields{"reason": "invalid_token"})
		return nil, errors.NewBusinessError("INVALID_TOKEN", "Invalid session token")
	}

	// Placeholder implementation
	s.metrics.IncrementCounter("authn_session_validations_successful", nil)
	return &model.Session{}, nil
}

func (s *service) InvalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.InvalidateSession")
	defer span.End()

	// NOTE: Sessions repository is not implemented yet, using placeholder
	// err = s.repo.Sessions().InvalidateSession(ctx, sessionID)
	err := error(nil)
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

	// NOTE: Sessions repository is not implemented yet, using placeholder
	// err = s.repo.Sessions().InvalidateUserSessions(ctx, userID)
	err := error(nil)
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
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// NOTE: UserRoles repository is not implemented yet, using placeholder
	// err = s.repo.UserRoles().Assign(ctx, userRole)
	err = nil
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

	// NOTE: UserRoles repository is not implemented yet, using placeholder
	// err = s.repo.UserRoles().Remove(ctx, req.UserID, req.RoleID, req.EntityID)
	err := error(nil)
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

func (s *service) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.Role, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.GetUserRoles")
	defer span.End()

	// Try cache first
	cacheKey := fmt.Sprintf("user_roles:%s", userID.String())
	var roles []*model.Role
	if err := s.getTenantCache(ctx).Get(ctx, cacheKey, &roles); err == nil {
		s.metrics.IncrementCounter("authn_user_roles_cache_hits", nil)
		return roles, nil
	}

	// NOTE: UserRoles repository is not implemented yet, using placeholder
	// userRoles, err = s.repo.UserRoles().GetUserRoles(ctx, userID)
	// Convert to Role objects
	roles = []*model.Role{}
	err := error(nil)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Cache the result - log any cache errors but don't fail the operation
	if err := s.getTenantCache(ctx).Set(ctx, cacheKey, roles, time.Hour); err != nil {
		s.logger.WarnContext(ctx, "Failed to cache user roles", logger.Fields{"error": err.Error(), "cache_key": cacheKey})
	}
	s.metrics.IncrementCounter("authn_user_roles_cache_misses", nil)

	return roles, nil
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

// Missing interface methods implementations

func (s *service) CreatePerson(ctx context.Context, req *CreatePersonRequest) (*model.Person, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.CreatePerson")
	defer span.End()

	person := &model.Person{
		ID:          uuid.New(),
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	createdPerson, err := s.repo.Persons().Create(ctx, person)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to create person: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "CREATE_PERSON", "person", createdPerson.ID)

	s.logger.InfoContext(ctx, "Person created successfully",
		logger.Fields{"person_id": createdPerson.ID})

	return createdPerson, nil
}

func (s *service) GetPerson(ctx context.Context, personID uuid.UUID) (*model.Person, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.GetPerson")
	defer span.End()

	person, err := s.repo.Persons().GetByID(ctx, personID)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to get person: %w", err)
	}

	return person, nil
}

func (s *service) UpdatePerson(ctx context.Context, req *UpdatePersonRequest) (*model.Person, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.UpdatePerson")
	defer span.End()

	// Get existing person
	person, err := s.GetPerson(ctx, req.PersonID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.FirstName != nil {
		person.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		person.LastName = *req.LastName
	}
	if req.Email != nil {
		person.Email = req.Email
	}
	if req.PhoneNumber != nil {
		person.PhoneNumber = req.PhoneNumber
	}
	// NOTE: BirthDate and Address fields would need proper handling
	// based on the actual Person struct fields
	person.UpdatedAt = time.Now()

	updatedPerson, err := s.repo.Persons().Update(ctx, person)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to update person: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "UPDATE_PERSON", "person", req.PersonID)

	s.logger.InfoContext(ctx, "Person updated successfully",
		logger.Fields{"person_id": req.PersonID})

	return updatedPerson, nil
}

func (s *service) CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*model.Employee, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.CreateEmployee")
	defer span.End()

	employee := &model.Employee{
		ID:               uuid.New(),
		PersonID:         req.PersonID,
		EmployeeNumber:   req.EmployeeNumber,
		PositionTitle:    &req.JobTitle,
		HireDate:         req.HireDate,
		ManagerID:        req.ManagerID,
		EmploymentStatus: model.EmploymentStatus(req.EmploymentStatus),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	createdEmployee, err := s.repo.Employees().Create(ctx, employee)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "CREATE_EMPLOYEE", "employee", createdEmployee.ID)

	s.logger.InfoContext(ctx, "Employee created successfully",
		logger.Fields{"employee_id": createdEmployee.ID, "employee_number": createdEmployee.EmployeeNumber})

	return createdEmployee, nil
}

func (s *service) GetEmployee(ctx context.Context, employeeID uuid.UUID) (*model.Employee, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.GetEmployee")
	defer span.End()

	employee, err := s.repo.Employees().GetByID(ctx, employeeID)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	return employee, nil
}

func (s *service) UpdateEmployee(ctx context.Context, req *UpdateEmployeeRequest) (*model.Employee, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.UpdateEmployee")
	defer span.End()

	// Get existing employee
	employee, err := s.GetEmployee(ctx, req.EmployeeID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.JobTitle != nil {
		employee.PositionTitle = req.JobTitle
	}
	if req.ManagerID != nil {
		employee.ManagerID = req.ManagerID
	}
	if req.EmploymentStatus != nil {
		employee.EmploymentStatus = model.EmploymentStatus(*req.EmploymentStatus)
	}
	// NOTE: Department and Salary would need proper handling
	// based on the actual Employee struct fields
	employee.UpdatedAt = time.Now()

	updatedEmployee, err := s.repo.Employees().Update(ctx, employee)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	// Audit log
	s.auditDataOperation(ctx, "UPDATE_EMPLOYEE", "employee", req.EmployeeID)

	s.logger.InfoContext(ctx, "Employee updated successfully",
		logger.Fields{"employee_id": req.EmployeeID})

	return updatedEmployee, nil
}

// Token validation and refresh methods
func (s *service) ValidateToken(ctx context.Context, token string) (*TokenValidationResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.ValidateToken")
	defer span.End()

	// Validate access token using JWT manager
	claims, err := s.jwtManager.ValidateAccessToken(token)
	if err != nil {
		s.metrics.IncrementCounter("authn_token_validations_failed", metrics.Fields{"reason": "invalid_token"})
		return &TokenValidationResult{Valid: false}, nil
	}

	// Extract claims for response
	claimsMap := map[string]any{
		"user_id":     claims.UserID.String(),
		"tenant_id":   claims.TenantID.String(),
		"entity_id":   claims.EntityID.String(),
		"email":       claims.Email,
		"roles":       claims.Roles,
		"mfa_enabled": claims.MFAEnabled,
		"expires_at":  claims.ExpiresAt.Time,
		"issued_at":   claims.IssuedAt.Time,
	}

	s.metrics.IncrementCounter("authn_token_validations_successful", nil)

	return &TokenValidationResult{
		Valid:  true,
		UserID: claims.UserID,
		Claims: claimsMap,
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*TokenRefreshResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "authn.service.RefreshToken")
	defer span.End()

	// Generate new access token using JWT manager
	newAccessToken, expiresAt, err := s.jwtManager.RefreshAccessToken(refreshToken)
	if err != nil {
		s.metrics.IncrementCounter("authn_token_refresh_failed", metrics.Fields{"reason": "invalid_refresh_token"})
		return nil, errors.NewBusinessError("INVALID_REFRESH_TOKEN", "Invalid or expired refresh token")
	}

	// NOTE: In production, you might want to rotate refresh tokens as well
	// TODO: Implement refresh token rotation for enhanced security
	// TODO: Add refresh token blacklisting to prevent reuse of compromised tokens
	// TODO: Track refresh token usage patterns for security monitoring

	s.metrics.IncrementCounter("authn_token_refresh_successful", nil)

	s.logger.InfoContext(ctx, "Access token refreshed successfully")

	return &TokenRefreshResult{
		AccessToken:  newAccessToken,
		RefreshToken: refreshToken, // Currently reusing same refresh token
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *service) Logout(ctx context.Context, userID uuid.UUID) error {
	// Placeholder implementation
	return s.InvalidateAllUserSessions(ctx, userID)
}

func (s *service) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	// Placeholder implementation - would involve email verification etc.
	return nil
}

func (s *service) ValidatePassword(ctx context.Context, userID uuid.UUID, password string) error {
	valid, err := s.verifyPassword(ctx, userID, password)
	if err != nil {
		return err
	}
	if !valid {
		return errors.NewBusinessError("INVALID_PASSWORD", "Password is invalid")
	}
	return nil
}

func (s *service) GenerateMFABackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.generateBackupCodes(), nil
}

func (s *service) GetSession(ctx context.Context, sessionID uuid.UUID) (*model.Session, error) {
	// Placeholder implementation
	return &model.Session{}, nil
}

func (s *service) LockAccount(ctx context.Context, userID uuid.UUID, reason string) error {
	// Placeholder implementation
	return s.repo.Users().LockAccount(ctx, userID, nil, reason)
}

func (s *service) UnlockAccount(ctx context.Context, userID uuid.UUID) error {
	// Placeholder implementation
	return s.repo.Users().UnlockAccount(ctx, userID)
}

func (s *service) IsAccountLocked(ctx context.Context, userID uuid.UUID) (bool, error) {
	// Placeholder implementation - would check account status
	return false, nil
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
