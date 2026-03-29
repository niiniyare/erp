// Package service contains IAM application services.
// Services orchestrate domain rules and repository calls.
// Cache is NOT managed here — it lives entirely in repository adapters.
package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/crypto/bcrypt"

	"awo.so/internal/core/iam/domain"
	"awo.so/internal/core/iam/repository"
	sharedErrors "awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// Port (interface)

// UserService defines the application service for identity operations.
type UserService interface {
	RegisterNewUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *domain.UpdateUserRequest) (*domain.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, req *domain.ListUsersRequest) ([]*domain.User, error)
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*domain.User, error)

	// Authentication with brute-force protection
	Authenticate(ctx context.Context, identifier, password string) (*domain.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error

	// Person / Employee
	CreatePerson(ctx context.Context, req *domain.CreatePersonRequest) (*domain.Person, error)
	GetPersonByID(ctx context.Context, id uuid.UUID) (*domain.Person, error)
	CreateEmployee(ctx context.Context, req *domain.CreateEmployeeRequest) (*domain.Employee, error)
	GetEmployeeByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error)

	// Role assignments (high-level — Casbin enforcement is in AuthzService)
	AssignUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error
	RevokeUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.UserRole, error)

	// Password reset flow
	// ForgotPassword generates a single-use reset token for the given email.
	// It returns the raw token (caller is responsible for emailing it) and the
	// target user ID. Returns ("", uuid.Nil, nil) when the email is not found —
	// callers must ALWAYS return 200 to the user to prevent enumeration.
	ForgotPassword(ctx context.Context, email string) (rawToken string, userID uuid.UUID, err error)
	// ResetPassword validates the token and sets a new password.
	// Returns ErrPasswordResetTokenNotFound, ErrPasswordResetTokenExpired,
	// ErrPasswordResetTokenUsed, ErrPasswordTooWeak, or ErrPasswordReused.
	ResetPassword(ctx context.Context, token, newPassword string) error

	// MFA lifecycle
	// InitiateMFA generates a fresh TOTP secret and caches it pending confirmation.
	// The returned MFASetup.Secret must be shown to the user once (for their authenticator app).
	// The secret is NOT saved to the DB until ConfirmMFA succeeds.
	InitiateMFA(ctx context.Context, userID uuid.UUID) (*domain.MFASetup, error)
	// ConfirmMFA validates the first TOTP code and if valid, persists the secret
	// in the DB and marks MFA as enabled for the user.
	ConfirmMFA(ctx context.Context, userID uuid.UUID, code string) error
	// ValidateMFACode verifies a TOTP code for an already-enrolled user.
	// Returns (false, nil) when the code is wrong or replayed.
	ValidateMFACode(ctx context.Context, userID uuid.UUID, code string) (bool, error)
	// DisableMFA requires password re-verification before clearing the secret.
	DisableMFA(ctx context.Context, userID uuid.UUID, password string) error
}

// Config

// UserConfig holds brute-force protection thresholds and MFA settings.
// TODO(settings): Replace hardcoded defaults with a settings.ConfigurationService lookup.
type UserConfig struct {
	MaxFailedAttempts int
	LockoutDuration   time.Duration

	// MFAEncryptionKey is the 32-byte AES-256 key used to encrypt TOTP secrets
	// at rest.  MFA operations fail if this is not set.
	// Set from a secure secret store (e.g. Vault, KMS) — do not hard-code.
	MFAEncryptionKey []byte

	// MFAIssuer is the issuer name shown in authenticator apps (default: "AWO ERP").
	MFAIssuer string
}

// DefaultUserConfig returns safe production defaults.
func DefaultUserConfig() UserConfig {
	return UserConfig{
		MaxFailedAttempts: 5,
		LockoutDuration:   15 * time.Minute,
		MFAIssuer:         "AWO ERP",
	}
}

// Implementation

type userService struct {
	repo    repository.UserRepository
	tracer  tracing.Service
	metrics metrics.MetricsProvider
	log     logger.Logger
	cfg     UserConfig
}

// NewUserService constructs a UserService with default config.
func NewUserService(repo repository.UserRepository, tracer tracing.Service, m metrics.MetricsProvider, log logger.Logger) UserService {
	return NewUserServiceWithConfig(repo, tracer, m, DefaultUserConfig(), log)
}

// NewUserServiceWithConfig constructs a UserService with explicit config.
func NewUserServiceWithConfig(
	repo repository.UserRepository,
	tracer tracing.Service,
	m metrics.MetricsProvider,
	cfg UserConfig,
	log logger.Logger,
) UserService {
	scopedLog := log.WithFields(logger.Fields{"component": "iam.identity"})
	return &userService{repo: repo, tracer: tracer, metrics: m, log: scopedLog, cfg: cfg}
}

func (s *userService) RegisterNewUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.RegisterNewUser")
	defer span.End()
	span.SetAttributes(attribute.String("user.email", req.Email))

	timer := s.metrics.Timer("iam_register_user_duration", nil)
	defer timer.Stop()

	if err := req.Validate(); err != nil {
		span.RecordError(err)
		return nil, err
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to hash password", logger.Fields{
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, fmt.Errorf("iam identity: hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, req, hashedPassword)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to create user", logger.Fields{
			"email":    req.Email,
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, err
	}

	s.log.DebugContext(ctx, "user registered", logger.Fields{"user_id": user.ID.String()})
	s.metrics.IncrementCounter("iam.users.registered", nil)
	return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.GetUserByID")
	defer span.End()
	span.SetAttributes(attribute.String("user.id", id.String()))

	timer := s.metrics.Timer("iam_get_user_duration", metrics.Fields{"by": "id"})
	defer timer.Stop()

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to get user by id", logger.Fields{
			"user_id":  id.String(),
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, err
	}
	return user, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.GetUserByEmail")
	defer span.End()

	timer := s.metrics.Timer("iam_get_user_duration", metrics.Fields{"by": "email"})
	defer timer.Stop()

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to get user by email", logger.Fields{
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, err
	}
	return user, nil
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.GetUserByUsername")
	defer span.End()

	timer := s.metrics.Timer("iam_get_user_duration", metrics.Fields{"by": "username"})
	defer timer.Stop()

	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to get user by username", logger.Fields{
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, err
	}
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, req *domain.UpdateUserRequest) (*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.UpdateUser")
	defer span.End()
	span.SetAttributes(attribute.String("user.id", id.String()))

	timer := s.metrics.Timer("iam_update_user_duration", nil)
	defer timer.Stop()

	user, err := s.repo.UpdateUser(ctx, id, req)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to update user", logger.Fields{
			"user_id":  id.String(),
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, err
	}

	s.log.DebugContext(ctx, "user updated", logger.Fields{"user_id": id.String()})
	return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.DeleteUser")
	defer span.End()
	span.SetAttributes(attribute.String("user.id", id.String()))

	timer := s.metrics.Timer("iam_delete_user_duration", nil)
	defer timer.Stop()

	if err := s.repo.DeleteUser(ctx, id); err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to delete user", logger.Fields{
			"user_id":  id.String(),
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return err
	}

	s.log.DebugContext(ctx, "user deleted", logger.Fields{"user_id": id.String()})
	return nil
}

func (s *userService) ListUsers(ctx context.Context, req *domain.ListUsersRequest) ([]*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.ListUsers")
	defer span.End()

	timer := s.metrics.Timer("iam_list_users_duration", nil)
	defer timer.Stop()

	users, err := s.repo.ListUsers(ctx, req)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to list users", logger.Fields{
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, err
	}
	return users, nil
}

func (s *userService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.SearchUsers")
	defer span.End()

	timer := s.metrics.Timer("iam_search_users_duration", nil)
	defer timer.Stop()

	users, err := s.repo.SearchUsers(ctx, query, limit, offset)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to search users", logger.Fields{
			"query":    query,
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, err
	}
	return users, nil
}

// Authenticate handles login with brute-force protection.
// 
// NOTE(tenant-context): ctx must carry tenant_id via cache.TenantIDKey.
// TODO(settings): Load thresholds from settings service per-tenant.
func (s *userService) Authenticate(ctx context.Context, identifier, password string) (*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.Authenticate")
	defer span.End()

	timer := s.metrics.Timer("iam_authenticate_duration", nil)
	defer timer.Stop()

	var user *domain.User
	var err error

	if strings.Contains(identifier, "@") {
		user, err = s.repo.GetUserByEmail(ctx, identifier)
	} else {
		user, err = s.repo.GetUserByUsername(ctx, identifier)
	}
	if err != nil {
		// Do not record a span error here — user-not-found is an expected
		// security event, not an infrastructure failure.
		s.metrics.IncrementCounter("iam.auth.user_not_found", nil)
		return nil, sharedErrors.ErrAuthenticationFailed
	}

	// Annotate the span with user identity now that we have it.
	span.SetAttributes(
		attribute.String("user.id", user.ID.String()),
		attribute.String("user.type", user.UserType),
	)

	// Domain rule: lockout check
	if user.IsLocked() {
		s.metrics.IncrementCounter("iam.auth.account_locked", nil)
		s.log.WarnContext(ctx, "auth blocked: account locked", logger.Fields{
			"user_id":  user.ID.String(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, sharedErrors.ErrAccountLocked
	}

	hashedPassword, err := s.repo.GetUserPassword(ctx, user.ID)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to retrieve password hash", logger.Fields{
			"user_id":  user.ID.String(),
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, sharedErrors.ErrAuthenticationFailed
	}

	if !verifyPassword(hashedPassword, password) {
		s.metrics.IncrementCounter("iam.auth.failure", nil)

		_ = s.repo.IncrementFailedAttempts(ctx, user.ID)

		newAttempts := int(user.FailedLoginAttempts) + 1
		if newAttempts >= s.cfg.MaxFailedAttempts {
			lockUntil := time.Now().Add(s.cfg.LockoutDuration)
			_ = s.repo.LockAccount(ctx, user.ID, lockUntil)
			s.log.WarnContext(ctx, "account locked after repeated failures", logger.Fields{
				"user_id":      user.ID.String(),
				"attempts":     newAttempts,
				"locked_until": lockUntil.Format(time.RFC3339),
			})
		}

		return nil, sharedErrors.ErrAuthenticationFailed
	}

	_ = s.repo.ResetFailedAttempts(ctx, user.ID)
	_ = s.repo.UpdateLastLogin(ctx, user.ID)

	s.metrics.IncrementCounter("iam.auth.success", nil)
	s.log.DebugContext(ctx, "authentication successful", logger.Fields{
		"user_id":   user.ID.String(),
		"user_type": user.UserType,
	})
	return user, nil
}

func (s *userService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.ChangePassword")
	defer span.End()
	span.SetAttributes(attribute.String("user.id", userID.String()))

	timer := s.metrics.Timer("iam_change_password_duration", nil)
	defer timer.Stop()

	if err := validatePasswordStrength(newPassword); err != nil {
		return err
	}

	currentHash, err := s.repo.GetUserPassword(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return sharedErrors.ErrAuthenticationFailed
	}
	if !verifyPassword(currentHash, oldPassword) {
		return sharedErrors.ErrAuthenticationFailed
	}

	history, err := s.repo.GetPasswordHistory(ctx, userID)
	if err == nil && isPasswordReused(newPassword, history) {
		return sharedErrors.ErrPasswordReused
	}

	newHash, err := hashPassword(newPassword)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to hash new password", logger.Fields{
			"user_id":  userID.String(),
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return fmt.Errorf("iam identity: hash new password: %w", err)
	}

	newHistory := prependHistory(newHash, history)
	if err := s.repo.UpdatePasswordAndHistory(ctx, userID, newHash, newHistory); err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to update password", logger.Fields{
			"user_id":  userID.String(),
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return err
	}

	s.log.DebugContext(ctx, "password changed", logger.Fields{"user_id": userID.String()})
	return nil
}

func (s *userService) CreatePerson(ctx context.Context, req *domain.CreatePersonRequest) (*domain.Person, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.CreatePerson")
	defer span.End()

	timer := s.metrics.Timer("iam_create_person_duration", nil)
	defer timer.Stop()

	person, err := s.repo.CreatePerson(ctx, req)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to create person", logger.Fields{
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, err
	}
	return person, nil
}

func (s *userService) GetPersonByID(ctx context.Context, id uuid.UUID) (*domain.Person, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.GetPersonByID")
	defer span.End()
	span.SetAttributes(attribute.String("person.id", id.String()))

	timer := s.metrics.Timer("iam_get_person_duration", nil)
	defer timer.Stop()

	person, err := s.repo.GetPersonByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to get person", logger.Fields{
			"person_id": id.String(),
			"error":     err.Error(),
			"trace_id":  s.tracer.GetTraceID(ctx),
		})
		return nil, err
	}
	return person, nil
}

func (s *userService) CreateEmployee(ctx context.Context, req *domain.CreateEmployeeRequest) (*domain.Employee, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.CreateEmployee")
	defer span.End()

	timer := s.metrics.Timer("iam_create_employee_duration", nil)
	defer timer.Stop()

	emp, err := s.repo.CreateEmployee(ctx, req)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to create employee", logger.Fields{
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, err
	}
	return emp, nil
}

func (s *userService) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.GetEmployeeByID")
	defer span.End()
	span.SetAttributes(attribute.String("employee.id", id.String()))

	timer := s.metrics.Timer("iam_get_employee_duration", nil)
	defer timer.Stop()

	emp, err := s.repo.GetEmployeeByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to get employee", logger.Fields{
			"employee_id": id.String(),
			"error":       err.Error(),
			"trace_id":    s.tracer.GetTraceID(ctx),
		})
		return nil, err
	}
	return emp, nil
}

func (s *userService) AssignUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.AssignUserRole")
	defer span.End()
	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("role.id", roleID.String()),
	)

	timer := s.metrics.Timer("iam_assign_role_duration", nil)
	defer timer.Stop()

	if err := s.repo.AssignUserRole(ctx, userID, roleID, entityID); err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to assign user role", logger.Fields{
			"user_id":  userID.String(),
			"role_id":  roleID.String(),
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return err
	}

	s.log.DebugContext(ctx, "user role assigned", logger.Fields{
		"user_id": userID.String(),
		"role_id": roleID.String(),
	})
	return nil
}

func (s *userService) RevokeUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.RevokeUserRole")
	defer span.End()
	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("role.id", roleID.String()),
	)

	timer := s.metrics.Timer("iam_revoke_role_duration", nil)
	defer timer.Stop()

	if err := s.repo.RevokeUserRole(ctx, userID, roleID, entityID); err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to revoke user role", logger.Fields{
			"user_id":  userID.String(),
			"role_id":  roleID.String(),
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return err
	}

	s.log.DebugContext(ctx, "user role revoked", logger.Fields{
		"user_id": userID.String(),
		"role_id": roleID.String(),
	})
	return nil
}

func (s *userService) GetUserRoles(_ context.Context, _ uuid.UUID) ([]*domain.UserRole, error) {
	// TODO(task-15): implement via UserRepository.GetUserRoleAssignments once that
	// query is added. Casbin-level roles are available via AuthzService.GetRoles.
	return []*domain.UserRole{}, nil
}

// MFA methods

func (s *userService) mfaKey() ([]byte, error) {
	if len(s.cfg.MFAEncryptionKey) != 32 {
		return nil, fmt.Errorf("iam identity: MFA not configured (encryption key must be 32 bytes)")
	}
	return s.cfg.MFAEncryptionKey, nil
}

func (s *userService) mfaIssuer() string {
	if s.cfg.MFAIssuer != "" {
		return s.cfg.MFAIssuer
	}
	return "AWO ERP"
}

func (s *userService) InitiateMFA(ctx context.Context, userID uuid.UUID) (*domain.MFASetup, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.InitiateMFA")
	defer span.End()

	key, err := s.mfaKey()
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	secret, err := generateTOTPSecret()
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("iam identity: generate totp secret: %w", err)
	}

	encrypted, err := encryptSecret(key, secret)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("iam identity: encrypt mfa secret: %w", err)
	}

	if err := s.repo.StorePendingMFASetup(ctx, userID, encrypted); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("iam identity: store pending mfa setup: %w", err)
	}

	s.metrics.IncrementCounter("iam.mfa.initiate", nil)
	return &domain.MFASetup{
		Secret: secret,
		QRURI:  buildTOTPURI(s.mfaIssuer(), user.Email, secret),
	}, nil
}

func (s *userService) ConfirmMFA(ctx context.Context, userID uuid.UUID, code string) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.ConfirmMFA")
	defer span.End()

	key, err := s.mfaKey()
	if err != nil {
		return err
	}

	encrypted, err := s.repo.GetPendingMFASetup(ctx, userID)
	if err != nil {
		return sharedErrors.ErrMFAInvalid
	}

	secret, err := decryptSecret(key, encrypted)
	if err != nil {
		span.RecordError(err)
		return sharedErrors.ErrMFAInvalid
	}

	_, ok := verifyTOTP(secret, code, 1)
	if !ok {
		return sharedErrors.ErrMFAInvalid
	}

	if err := s.repo.SetMFASecret(ctx, userID, encrypted); err != nil {
		span.RecordError(err)
		return fmt.Errorf("iam identity: save mfa secret: %w", err)
	}
	_ = s.repo.ClearPendingMFASetup(ctx, userID)

	s.metrics.IncrementCounter("iam.mfa.confirmed", nil)
	s.log.DebugContext(ctx, "MFA enabled", logger.Fields{"user_id": userID.String()})
	return nil
}

func (s *userService) ValidateMFACode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.ValidateMFACode")
	defer span.End()

	key, err := s.mfaKey()
	if err != nil {
		return false, err
	}

	encrypted, enabled, err := s.repo.GetMFASecret(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return false, err
	}
	if !enabled || encrypted == "" {
		return false, nil
	}

	secret, err := decryptSecret(key, encrypted)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("iam identity: decrypt mfa secret: %w", err)
	}

	window, ok := verifyTOTP(secret, code, 1)
	if !ok {
		s.metrics.IncrementCounter("iam.mfa.invalid", nil)
		return false, nil
	}

	// Replay prevention: reject if this window was already used.
	replayed, err := s.repo.CheckAndMarkMFAReplay(ctx, userID, window)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("iam identity: replay check: %w", err)
	}
	if replayed {
		s.metrics.IncrementCounter("iam.mfa.replay", nil)
		s.log.WarnContext(ctx, "MFA replay attempt blocked", logger.Fields{"user_id": userID.String()})
		return false, nil
	}

	s.metrics.IncrementCounter("iam.mfa.success", nil)
	return true, nil
}

func (s *userService) DisableMFA(ctx context.Context, userID uuid.UUID, password string) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.DisableMFA")
	defer span.End()

	currentHash, err := s.repo.GetUserPassword(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return sharedErrors.ErrAuthenticationFailed
	}
	if !verifyPassword(currentHash, password) {
		return sharedErrors.ErrAuthenticationFailed
	}

	if err := s.repo.ClearMFASecret(ctx, userID); err != nil {
		span.RecordError(err)
		return fmt.Errorf("iam identity: disable mfa: %w", err)
	}

	s.metrics.IncrementCounter("iam.mfa.disabled", nil)
	s.log.DebugContext(ctx, "MFA disabled", logger.Fields{"user_id": userID.String()})
	return nil
}

// Password reset

const passwordResetTTL = time.Hour

func (s *userService) ForgotPassword(ctx context.Context, email string) (string, uuid.UUID, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.ForgotPassword")
	defer span.End()

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		// User not found — return empty strings (caller always returns 200)
		return "", uuid.Nil, nil
	}

	rawToken, tokenHash, err := generatePasswordResetToken()
	if err != nil {
		span.RecordError(err)
		return "", uuid.Nil, fmt.Errorf("iam identity: generate reset token: %w", err)
	}

	expiresAt := time.Now().Add(passwordResetTTL)
	if err := s.repo.CreatePasswordResetToken(ctx, user.ID, tokenHash, expiresAt); err != nil {
		span.RecordError(err)
		return "", uuid.Nil, fmt.Errorf("iam identity: store reset token: %w", err)
	}

	s.metrics.IncrementCounter("iam.password.reset_requested", nil)
	s.log.DebugContext(ctx, "password reset token generated", logger.Fields{"user_id": user.ID.String()})
	return rawToken, user.ID, nil
}

func (s *userService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.identity.ResetPassword")
	defer span.End()

	if err := validatePasswordStrength(newPassword); err != nil {
		return err
	}

	tokenHash := sha256Hex(rawToken)
	tok, err := s.repo.GetPasswordResetToken(ctx, tokenHash)
	if err != nil {
		return sharedErrors.ErrPasswordResetTokenNotFound
	}
	if tok.IsExpired() {
		return sharedErrors.ErrPasswordResetTokenExpired
	}
	if tok.IsUsed() {
		return sharedErrors.ErrPasswordResetTokenUsed
	}

	history, err := s.repo.GetPasswordHistory(ctx, tok.UserID)
	if err != nil {
		// Non-fatal: log and continue without history check
		s.log.WarnContext(ctx, "failed to load password history, skipping reuse check", logger.Fields{
			"user_id": tok.UserID.String(),
			"error":   err.Error(),
		})
	} else {
		if isPasswordReused(newPassword, history) {
			return sharedErrors.ErrPasswordReused
		}
	}

	newHash, err := hashPassword(newPassword)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("iam identity: hash password: %w", err)
	}

	newHistory := prependHistory(newHash, history)
	if err := s.repo.UpdatePasswordAndHistory(ctx, tok.UserID, newHash, newHistory); err != nil {
		span.RecordError(err)
		return fmt.Errorf("iam identity: update password: %w", err)
	}

	// Mark token as used AFTER the password update succeeds.
	_ = s.repo.MarkPasswordResetTokenUsed(ctx, tokenHash)

	s.metrics.IncrementCounter("iam.password.reset_completed", nil)
	s.log.DebugContext(ctx, "password reset completed", logger.Fields{"user_id": tok.UserID.String()})
	return nil
}

// Password helpers

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func verifyPassword(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

// validatePasswordStrength enforces the password policy:
// minimum 12 characters, at least one uppercase, lowercase, digit, and special character.
func validatePasswordStrength(password string) error {
	if len(password) < 12 {
		return sharedErrors.ErrPasswordTooWeak
	}
	var upper, lower, digit, special bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			upper = true
		case c >= 'a' && c <= 'z':
			lower = true
		case c >= '0' && c <= '9':
			digit = true
		default:
			special = true
		}
	}
	if !upper || !lower || !digit || !special {
		return sharedErrors.ErrPasswordTooWeak
	}
	return nil
}

// isPasswordReused checks whether the plain-text password matches any of the
// stored bcrypt hashes.  Returns true if there is a match (reuse detected).
func isPasswordReused(plain string, history []string) bool {
	for _, h := range history {
		if verifyPassword(h, plain) {
			return true
		}
	}
	return false
}

// prependHistory prepends newHash to history and caps the slice at 5 entries.
func prependHistory(newHash string, existing []string) []string {
	combined := append([]string{newHash}, existing...)
	if len(combined) > 5 {
		combined = combined[:5]
	}
	return combined
}

// generatePasswordResetToken creates a random 32-byte raw token and returns
// both the hex-encoded raw token and its SHA-256 hash.
func generatePasswordResetToken() (rawToken, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	rawToken = hex.EncodeToString(b)
	hash = sha256Hex(rawToken)
	return rawToken, hash, nil
}
