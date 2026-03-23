// Package service contains IAM application services.
// Services orchestrate domain rules and repository calls.
// Cache is NOT managed here — it lives entirely in repository adapters.
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"awo.so/internal/core/iam/domain"
	"awo.so/internal/core/iam/repository"
	sharedErrors "awo.so/internal/shared/errors"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// ─── Port (interface) ─────────────────────────────────────────────────────────

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
}

// ─── Config ───────────────────────────────────────────────────────────────────

// UserConfig holds brute-force protection thresholds.
// TODO(settings): Replace hardcoded defaults with a settings.ConfigurationService lookup.
type UserConfig struct {
	MaxFailedAttempts int
	LockoutDuration   time.Duration
}

// DefaultUserConfig returns safe production defaults.
func DefaultUserConfig() UserConfig {
	return UserConfig{
		MaxFailedAttempts: 5,
		LockoutDuration:   15 * time.Minute,
	}
}

// ─── Implementation ───────────────────────────────────────────────────────────

type userService struct {
	repo    repository.UserRepository
	tracer  tracing.Service
	metrics metrics.MetricsProvider
	cfg     UserConfig
}

// NewUserService constructs a UserService with default config.
func NewUserService(repo repository.UserRepository, tracer tracing.Service, m metrics.MetricsProvider) UserService {
	return NewUserServiceWithConfig(repo, tracer, m, DefaultUserConfig())
}

// NewUserServiceWithConfig constructs a UserService with explicit config.
func NewUserServiceWithConfig(
	repo repository.UserRepository,
	tracer tracing.Service,
	m metrics.MetricsProvider,
	cfg UserConfig,
) UserService {
	return &userService{repo: repo, tracer: tracer, metrics: m, cfg: cfg}
}

func (s *userService) RegisterNewUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.service.RegisterNewUser")
	defer span.End()

	// Domain validation
	if err := req.Validate(); err != nil {
		return nil, err
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("iam service: hash password: %w", err)
	}

	return s.repo.CreateUser(ctx, req, hashedPassword)
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.service.GetUserByID")
	defer span.End()
	return s.repo.GetUserByID(ctx, id)
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.service.GetUserByEmail")
	defer span.End()
	return s.repo.GetUserByEmail(ctx, email)
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.service.GetUserByUsername")
	defer span.End()
	return s.repo.GetUserByUsername(ctx, username)
}

func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, req *domain.UpdateUserRequest) (*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.service.UpdateUser")
	defer span.End()
	return s.repo.UpdateUser(ctx, id, req)
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.service.DeleteUser")
	defer span.End()
	return s.repo.DeleteUser(ctx, id)
}

func (s *userService) ListUsers(ctx context.Context, req *domain.ListUsersRequest) ([]*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.service.ListUsers")
	defer span.End()
	return s.repo.ListUsers(ctx, req)
}

func (s *userService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.service.SearchUsers")
	defer span.End()
	return s.repo.SearchUsers(ctx, query, limit, offset)
}

// Authenticate handles login with brute-force protection.
//
// NOTE(tenant-context): ctx must carry tenant_id via cache.TenantIDKey.
// TODO(settings): Load thresholds from settings service per-tenant.
func (s *userService) Authenticate(ctx context.Context, identifier, password string) (*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.service.Authenticate")
	defer span.End()

	var user *domain.User
	var err error

	if strings.Contains(identifier, "@") {
		user, err = s.repo.GetUserByEmail(ctx, identifier)
	} else {
		user, err = s.repo.GetUserByUsername(ctx, identifier)
	}
	if err != nil {
		s.metrics.IncrementCounter("iam.auth.user_not_found", nil)
		return nil, sharedErrors.ErrAuthenticationFailed
	}

	// Domain rule: lockout check
	if user.IsLocked() {
		s.metrics.IncrementCounter("iam.auth.account_locked", nil)
		return nil, sharedErrors.ErrAccountLocked
	}

	hashedPassword, err := s.repo.GetUserPassword(ctx, user.ID)
	if err != nil {
		return nil, sharedErrors.ErrAuthenticationFailed
	}

	if !verifyPassword(hashedPassword, password) {
		s.metrics.IncrementCounter("iam.auth.failure", nil)

		_ = s.repo.IncrementFailedAttempts(ctx, user.ID)

		// Domain rule: lock after max attempts
		newAttempts := int(user.FailedLoginAttempts) + 1
		if newAttempts >= s.cfg.MaxFailedAttempts {
			lockUntil := time.Now().Add(s.cfg.LockoutDuration)
			_ = s.repo.LockAccount(ctx, user.ID, lockUntil)
		}

		return nil, sharedErrors.ErrAuthenticationFailed
	}

	_ = s.repo.ResetFailedAttempts(ctx, user.ID)
	_ = s.repo.UpdateLastLogin(ctx, user.ID)

	s.metrics.IncrementCounter("iam.auth.success", nil)
	return user, nil
}

func (s *userService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.service.ChangePassword")
	defer span.End()

	currentHash, err := s.repo.GetUserPassword(ctx, userID)
	if err != nil {
		return sharedErrors.ErrAuthenticationFailed
	}
	if !verifyPassword(currentHash, oldPassword) {
		return sharedErrors.ErrAuthenticationFailed
	}

	newHash, err := hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("iam service: hash new password: %w", err)
	}

	return s.repo.UpdatePassword(ctx, userID, newHash)
}

func (s *userService) CreatePerson(ctx context.Context, req *domain.CreatePersonRequest) (*domain.Person, error) {
	return s.repo.CreatePerson(ctx, req)
}

func (s *userService) GetPersonByID(ctx context.Context, id uuid.UUID) (*domain.Person, error) {
	return s.repo.GetPersonByID(ctx, id)
}

func (s *userService) CreateEmployee(ctx context.Context, req *domain.CreateEmployeeRequest) (*domain.Employee, error) {
	return s.repo.CreateEmployee(ctx, req)
}

func (s *userService) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	return s.repo.GetEmployeeByID(ctx, id)
}

func (s *userService) AssignUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	return s.repo.AssignUserRole(ctx, userID, roleID, entityID)
}

func (s *userService) RevokeUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	return s.repo.RevokeUserRole(ctx, userID, roleID, entityID)
}

func (s *userService) GetUserRoles(_ context.Context, _ uuid.UUID) ([]*domain.UserRole, error) {
	// Role retrieval via Casbin is in AuthzService. This stub exists for interface compliance.
	// TODO: wire AuthzService here or expose via a combined query.
	return []*domain.UserRole{}, nil
}

// ─── Password helpers ─────────────────────────────────────────────────────────

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
