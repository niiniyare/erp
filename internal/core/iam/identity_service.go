package iam

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"awo/internal/platform/cache"
	"awo/internal/shared/errors"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
	"golang.org/x/crypto/bcrypt"
)

// UserService defines the core business logic for identity operations.
// Callers should depend on this interface, not the concrete *userService.
type UserService interface {
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

	ListUsers(ctx context.Context, req *ListUsersRequest) ([]*User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*UserRole, error)
	SearchUsers(ctx context.Context, query string, limit int, offset int) ([]*User, error)
}

// UserConfig holds security thresholds for the identity service.
//
// NOTE: These are process-level defaults loaded from platform/config.AuthConfig
// at startup. Per-tenant overrides belong in the Settings module.
//
// TODO(settings): Replace hardcoded defaults with a settings.ConfigurationService lookup.
// TODO(feature-flags): MFAEnabled should be sourced from the tenant feature-flag system.
type UserConfig struct {
	MaxFailedAttempts int
	LockoutDuration   time.Duration
	SessionTTL        time.Duration
	MFAEnabled        bool
}

// DefaultUserConfig returns safe production defaults.
func DefaultUserConfig() UserConfig {
	return UserConfig{
		MaxFailedAttempts: 5,
		LockoutDuration:   15 * time.Minute,
		SessionTTL:        8 * time.Hour,
		MFAEnabled:        false,
	}
}

type userService struct {
	repo    UserRepository
	cache   cache.Service
	tracing tracing.Service
	metrics metrics.MetricsProvider
	cfg     UserConfig
}

// NewUserService creates a new identity service with default config.
func NewUserService(repo UserRepository, cache cache.Service, tracing tracing.Service, metrics metrics.MetricsProvider) UserService {
	return NewUserServiceWithConfig(repo, cache, tracing, metrics, DefaultUserConfig())
}

// NewUserServiceWithConfig creates a new identity service with explicit config.
func NewUserServiceWithConfig(repo UserRepository, cache cache.Service, tracing tracing.Service, metrics metrics.MetricsProvider, cfg UserConfig) UserService {
	return &userService{
		repo:    repo,
		cache:   cache,
		tracing: tracing,
		metrics: metrics,
		cfg:     cfg,
	}
}

func (s *userService) RegisterNewUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "iam.service.RegisterNewUser")
	defer span.End()

	if req.Password == "" || len(req.Password) < 8 {
		return nil, fmt.Errorf("validation: password must be at least 8 characters")
	}

	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash_password: failed to hash password for new user: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, req, hashedPassword)
	if err != nil {
		return nil, err
	}

	s.invalidateUserCache(ctx, user.ID, user.Email, user.Username)
	return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "iam.service.GetUserByID")
	defer span.End()

	cacheKey := fmt.Sprintf("user:id:%s", id)
	var user User
	if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil
	}

	dbUser, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.cacheUser(ctx, dbUser)
	return dbUser, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "iam.service.GetUserByEmail")
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

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "iam.service.GetUserByUsername")
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

func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "iam.service.UpdateUser")
	defer span.End()

	originalUser, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updatedUser, err := s.repo.UpdateUser(ctx, id, req)
	if err != nil {
		return nil, err
	}

	s.invalidateUserCache(ctx, originalUser.ID, originalUser.Email, originalUser.Username)
	s.cacheUser(ctx, updatedUser)
	return updatedUser, nil
}

// Authenticate handles user login with brute-force protection.
//
// NOTE(tenant-context): ctx must carry tenant_id via cache.TenantIDKey.
// TODO(settings): Load thresholds from settings service per-tenant.
// TODO(feature-flags): Add MFA check once feature-flag middleware is resolved.
func (s *userService) Authenticate(ctx context.Context, identifier, password string) (*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "iam.service.Authenticate")
	defer span.End()

	var user *User
	var err error

	if strings.Contains(identifier, "@") {
		user, err = s.repo.GetUserByEmail(ctx, identifier)
	} else {
		user, err = s.repo.GetUserByUsername(ctx, identifier)
	}
	if err != nil {
		s.metrics.IncrementCounter("iam.auth.user_not_found", nil)
		return nil, errors.ErrAuthenticationFailed
	}

	if user.LockoutUntil != nil && !user.LockoutUntil.IsZero() && user.LockoutUntil.After(time.Now()) {
		s.metrics.IncrementCounter("iam.auth.account_locked", nil)
		return nil, errors.ErrAccountLocked
	}

	hashedPassword, err := s.repo.GetUserPassword(ctx, user.ID)
	if err != nil {
		return nil, errors.ErrAuthenticationFailed
	}

	if !s.verifyPassword(hashedPassword, password) {
		s.metrics.IncrementCounter("iam.auth.failure", nil)

		if incrErr := s.repo.IncrementFailedAttempts(ctx, user.ID); incrErr != nil {
			s.metrics.IncrementCounter("iam.auth.increment_error", nil)
		}

		newAttempts := int(user.FailedLoginAttempts) + 1
		if newAttempts >= s.cfg.MaxFailedAttempts {
			lockUntil := time.Now().Add(s.cfg.LockoutDuration)
			if lockErr := s.repo.LockAccount(ctx, user.ID, lockUntil); lockErr != nil {
				s.metrics.IncrementCounter("iam.auth.lock_error", nil)
			}
		}

		return nil, errors.ErrAuthenticationFailed
	}

	if resetErr := s.repo.ResetFailedAttempts(ctx, user.ID); resetErr != nil {
		s.metrics.IncrementCounter("iam.auth.reset_error", nil)
	}

	if loginErr := s.repo.UpdateLastLogin(ctx, user.ID); loginErr != nil {
		s.metrics.IncrementCounter("iam.auth.last_login_update_error", nil)
	}

	s.invalidateUserCache(ctx, user.ID, user.Email, user.Username)
	s.metrics.IncrementCounter("iam.auth.success", nil)
	return user, nil
}

func (s *userService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	ctx, span := s.tracing.StartSpan(ctx, "iam.service.ChangePassword")
	defer span.End()

	currentHashedPassword, err := s.repo.GetUserPassword(ctx, userID)
	if err != nil {
		return errors.ErrAuthenticationFailed
	}
	if !s.verifyPassword(currentHashedPassword, oldPassword) {
		return errors.ErrAuthenticationFailed
	}

	newHashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash_password: failed to hash new password: %w", err)
	}

	return s.repo.UpdatePassword(ctx, userID, newHashedPassword)
}

func (s *userService) CreatePerson(ctx context.Context, req *CreatePersonRequest) (*Person, error) {
	return s.repo.CreatePerson(ctx, req)
}

func (s *userService) GetPersonByID(ctx context.Context, id uuid.UUID) (*Person, error) {
	return s.repo.GetPersonByID(ctx, id)
}

func (s *userService) CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*Employee, error) {
	return s.repo.CreateEmployee(ctx, req)
}

func (s *userService) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*Employee, error) {
	return s.repo.GetEmployeeByID(ctx, id)
}

func (s *userService) AssignUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	return s.repo.AssignUserRole(ctx, userID, roleID, entityID)
}

func (s *userService) RevokeUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	return s.repo.RevokeUserRole(ctx, userID, roleID, entityID)
}

func (s *userService) ListUsers(ctx context.Context, req *ListUsersRequest) ([]*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "iam.service.ListUsers")
	defer span.End()
	return s.repo.ListUsers(ctx, req)
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "iam.service.DeleteUser")
	defer span.End()
	if err := s.repo.DeleteUser(ctx, id); err != nil {
		return err
	}
	s.invalidateUserCache(ctx, id, "", "")
	return nil
}

func (s *userService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*UserRole, error) {
	// Role retrieval requires the authz Service (Casbin). Wire via handler layer.
	return []*UserRole{}, nil
}

func (s *userService) SearchUsers(ctx context.Context, query string, limit int, offset int) ([]*User, error) {
	ctx, span := s.tracing.StartSpan(ctx, "iam.service.SearchUsers")
	defer span.End()
	return s.repo.SearchUsers(ctx, query, limit, offset)
}

// --- Private Helpers ---

func (s *userService) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (s *userService) verifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *userService) cacheUser(ctx context.Context, user *User) {
	ttl := 30 * time.Minute
	s.cache.Set(ctx, fmt.Sprintf("user:id:%s", user.ID), user, ttl)
	s.cache.Set(ctx, fmt.Sprintf("user:email:%s", user.Email), user, ttl)
	if user.Username != "" {
		s.cache.Set(ctx, fmt.Sprintf("user:username:%s", user.Username), user, ttl)
	}
}

func (s *userService) invalidateUserCache(ctx context.Context, id uuid.UUID, email, username string) {
	s.cache.Delete(ctx, fmt.Sprintf("user:id:%s", id))
	s.cache.Delete(ctx, fmt.Sprintf("user:email:%s", email))
	if username != "" {
		s.cache.Delete(ctx, fmt.Sprintf("user:username:%s", username))
	}
}
