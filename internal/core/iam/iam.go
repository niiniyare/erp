// Package iam is the IAM bounded-context facade.
// External callers import only this package; internal sub-packages are an
// implementation detail. Following the same pattern as internal/core/tenant.
package iam

import (
	db "awo.so/db/sqlc"
	"awo.so/internal/core/iam/domain"
	"awo.so/internal/core/iam/repository"
	iamservice "awo.so/internal/core/iam/service"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// Re-export: Domain Types

type (
	// Identity
	User                  = domain.User
	Person                = domain.Person
	Employee              = domain.Employee
	UserWithDetails       = domain.UserWithDetails
	UserRole              = domain.UserRole
	AccountStatus         = domain.AccountStatus
	EmploymentStatus      = domain.EmploymentStatus
	CreateUserRequest     = domain.CreateUserRequest
	UpdateUserRequest     = domain.UpdateUserRequest
	ListUsersRequest      = domain.ListUsersRequest
	AuthenticateRequest   = domain.AuthenticateRequest
	ChangePasswordRequest = domain.ChangePasswordRequest
	CreatePersonRequest   = domain.CreatePersonRequest
	CreateEmployeeRequest = domain.CreateEmployeeRequest

	// Authorization
	ActorType      = domain.ActorType
	Principal      = domain.Principal
	Request        = domain.Request
	Policy         = domain.Policy
	RoleAssignment = domain.RoleAssignment
	AssignOpt      = domain.AssignOpt
	AssignOpts     = domain.AssignOpts

	// Password reset
	PasswordResetToken = domain.PasswordResetToken

	// MFA
	MFASetup = domain.MFASetup

	// SSO
	OAuthProvider            = domain.OAuthProvider
	SSOProvider              = domain.SSOProvider
	SSOUserInfo              = domain.SSOUserInfo
	CreateSSOProviderRequest = domain.CreateSSOProviderRequest

	// API Keys
	APIKey              = domain.APIKey
	CreateAPIKeyRequest = domain.CreateAPIKeyRequest

	// Session
	SessionConfig   = domain.SessionConfig
	EntityScopeType = domain.EntityScopeType
	EntityScope     = domain.EntityScope
	Configuration   = domain.Configuration
	Session         = domain.Session
	ResolvedSession = domain.ResolvedSession

	// Errors
	Error = domain.Error
)

// Re-export: Constants

// casbinModel exposes the Casbin CONF model string for package-level tests.
const casbinModel = domain.CasbinModel

const (
	// Account statuses
	AccountStatusActive    = domain.AccountStatusActive
	AccountStatusInactive  = domain.AccountStatusInactive
	AccountStatusLocked    = domain.AccountStatusLocked
	AccountStatusSuspended = domain.AccountStatusSuspended

	// Actor types
	ActorPlatform = domain.ActorPlatform
	ActorTenant   = domain.ActorTenant
	ActorPortal   = domain.ActorPortal
	ActorAPI      = domain.ActorAPI

	// Domains
	DomainPlatform = domain.DomainPlatform

	// Entity scope types
	EntityScopeAll     = domain.EntityScopeAll
	EntityScopeSubtree = domain.EntityScopeSubtree
	EntityScopeEntity  = domain.EntityScopeEntity

	// OAuth providers
	OAuthProviderGoogle    = domain.OAuthProviderGoogle
	OAuthProviderMicrosoft = domain.OAuthProviderMicrosoft

	// Locals keys (Fiber context)
	LocalsKeySession   = domain.LocalsKeySession
	LocalsKeyPrincipal = domain.LocalsKeyPrincipal
)

// Re-export: Errors

var (
	ErrForbidden      = domain.ErrForbidden
	ErrUnauthorized   = domain.ErrUnauthorized
	ErrInvalidRequest = domain.ErrInvalidRequest
	ErrPolicyConflict = domain.ErrPolicyConflict
)

// Re-export: Functions

var (
	// Subject helpers
	PlatformSubject = domain.PlatformSubject
	TenantSubject   = domain.TenantSubject
	PortalSubject   = domain.PortalSubject
	APISubject      = domain.APISubject

	// Domain helpers
	TenantDomain = domain.TenantDomain
	PortalDomain = domain.PortalDomain
	APIDomain    = domain.APIDomain

	// AssignOpt constructors
	WithExpiry      = domain.WithExpiry
	WithAssignedBy  = domain.WithAssignedBy
	WithDelegatedBy = domain.WithDelegatedBy

	// Config defaults
	DefaultSessionConfig  = domain.DefaultSessionConfig
	DefaultConfiguration  = domain.DefaultConfiguration
	AllAccountStatuses    = domain.AllAccountStatuses
	AllEmploymentStatuses = domain.AllEmploymentStatuses
)

// Re-export: Service Interfaces

type (
	UserService        = iamservice.UserService
	AuthzService       = iamservice.AuthzService
	SessionService     = iamservice.SessionService
	SSOService         = iamservice.SSOService
	APIKeyService      = iamservice.APIKeyService
	SessionInvalidator = iamservice.SessionInvalidator

	// Service is a backward-compatible alias for AuthzService.
	// Prefer AuthzService in new code.
	Service = iamservice.AuthzService
)

// Re-export: Repository Interfaces

type (
	UserRepository    = repository.UserRepository
	AuthzRepository   = repository.AuthzRepository
	SessionRepository = repository.SessionRepository
	SSORepository     = repository.SSORepository
	APIKeyRepository  = repository.APIKeyRepository
)

// Re-export: AuthzConfig

// Config is the constructor config for the AuthzService (Casbin).
type Config = iamservice.AuthzConfig

// UserConfig holds brute-force protection thresholds for the UserService.
type UserConfig = iamservice.UserConfig

// SSOConfig is the constructor config for the SSOService.
type SSOConfig = iamservice.SSOConfig

// Constructors (wire entry points)

// NewUserRepository constructs a cache-backed Postgres UserRepository.
func NewUserRepository(store db.Store, cacheSvc cache.Service, tracer tracing.Service, m metrics.MetricsProvider) UserRepository {
	return repository.NewUserRepository(store, cacheSvc, tracer, m)
}

// NewUserService constructs a UserService with default brute-force config.
// Cache is handled by the repository — the service receives no cache dependency.
func NewUserService(repo UserRepository, tracer tracing.Service, m metrics.MetricsProvider, log logger.Logger) UserService {
	return iamservice.NewUserService(repo, tracer, m, log)
}

// NewUserServiceWithConfig constructs a UserService with explicit brute-force config.
// Pass authz to enable automatic tenant_admin bootstrap on user creation; pass nil to skip.
func NewUserServiceWithConfig(repo UserRepository, authz AuthzService, tracer tracing.Service, m metrics.MetricsProvider, cfg UserConfig, log logger.Logger) UserService {
	return iamservice.NewUserServiceWithConfig(repo, authz, tracer, m, cfg, log)
}

// New constructs a fully initialised AuthzService backed by PostgreSQL via Casbin.
// Config must include Store, Cache, and Logger.
func New(cfg Config) (AuthzService, error) {
	return iamservice.NewAuthzService(cfg)
}

// NewInMemoryAuthzService creates an in-memory AuthzService for unit tests.
// See service.NewInMemoryAuthzService for details.
func NewInMemoryAuthzService(repo AuthzRepository, log logger.Logger) (AuthzService, error) {
	return iamservice.NewInMemoryAuthzService(repo, log)
}

// NewSessionRepository constructs a cache-backed Postgres SessionRepository.
func NewSessionRepository(store db.Store, cacheSvc cache.Service, tracer tracing.Service, m metrics.MetricsProvider) SessionRepository {
	return repository.NewSessionRepository(store, cacheSvc, tracer, m)
}

// NewSessionService constructs a SessionService with default session config.
// Cache is handled by the repository — the service receives no cache dependency.
func NewSessionService(
	identity UserService,
	repo SessionRepository,
	tracer tracing.Service,
	m metrics.MetricsProvider,
	log logger.Logger,
) SessionService {
	return iamservice.NewSessionService(identity, repo, tracer, m, log)
}

// NewSessionServiceWithConfig constructs a SessionService with explicit session config.
func NewSessionServiceWithConfig(
	identity UserService,
	repo SessionRepository,
	tracer tracing.Service,
	m metrics.MetricsProvider,
	log logger.Logger,
	cfg SessionConfig,
) SessionService {
	return iamservice.NewSessionServiceWithConfig(identity, repo, tracer, m, log, cfg)
}

// NewSSORepository constructs a Postgres-backed SSORepository.
func NewSSORepository(store db.Store) SSORepository {
	return repository.NewSSORepository(store)
}

// NewSSOService constructs an SSOService for OAuth/OIDC login flows.
func NewSSOService(
	repo SSORepository,
	identity UserService,
	tracer tracing.Service,
	m metrics.MetricsProvider,
	log logger.Logger,
	cfg SSOConfig,
) SSOService {
	return iamservice.NewSSOService(repo, identity, tracer, m, log, cfg)
}

// NewAPIKeyRepository constructs a Postgres-backed APIKeyRepository.
// NOTE: Run `make sqlc` after applying migration 000310 to generate the
// required Store methods for this repository.
func NewAPIKeyRepository(store db.Store) APIKeyRepository {
	return repository.NewAPIKeyRepository(store)
}

// NewAPIKeyService constructs an APIKeyService.
func NewAPIKeyService(
	repo APIKeyRepository,
	cacheSvc cache.Service,
	tracer tracing.Service,
	m metrics.MetricsProvider,
) APIKeyService {
	return iamservice.NewAPIKeyService(repo, cacheSvc, tracer, m)
}
