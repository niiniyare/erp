// Package contract — re-exports from the iam bounded context.
//
// External packages MUST import this package instead of
// "awo.so/internal/core/iam" directly.
//
// Type aliases are transparent: contract.AuthzService IS iam.AuthzService,
// so wire and existing code that uses the underlying type sees no change.
package contract

import (
	"context"

	"awo.so/internal/core/iam"
)

// ============================================================================
// Service interfaces
// ============================================================================

type (
	AuthzService       = iam.AuthzService
	SessionService     = iam.SessionService
	UserService        = iam.UserService
	APIKeyService      = iam.APIKeyService
	SSOService         = iam.SSOService
	SessionInvalidator = iam.SessionInvalidator

	// Service is a backward-compatible alias for AuthzService.
	Service = iam.Service
)

// ============================================================================
// Repository interfaces
// ============================================================================

type (
	UserRepository    = iam.UserRepository
	AuthzRepository   = iam.AuthzRepository
	SessionRepository = iam.SessionRepository
	APIKeyRepository  = iam.APIKeyRepository
	SSORepository     = iam.SSORepository
)

// ============================================================================
// Domain types
// ============================================================================

type (
	// Identity
	User = iam.User

	// Authorization
	Principal      = iam.Principal
	Request        = iam.Request
	Policy         = iam.Policy
	RoleAssignment = iam.RoleAssignment

	// Session
	ResolvedSession = iam.ResolvedSession
	EntityScope     = iam.EntityScope

	// Config types
	Config        = iam.Config
	UserConfig    = iam.UserConfig
	SSOConfig     = iam.SSOConfig
	SessionConfig = iam.SessionConfig
)

// ============================================================================
// Constants
// ============================================================================

const (
	LocalsKeySession   = iam.LocalsKeySession
	LocalsKeyPrincipal = iam.LocalsKeyPrincipal
)

// ============================================================================
// Errors
// ============================================================================

var (
	ErrForbidden    = iam.ErrForbidden
	ErrUnauthorized = iam.ErrUnauthorized
)

// ============================================================================
// Constructors
// ============================================================================

var (
	// AuthzService
	New = iam.New

	// UserService
	NewUserRepository        = iam.NewUserRepository
	NewUserService           = iam.NewUserService
	NewUserServiceWithConfig = iam.NewUserServiceWithConfig

	// SessionService
	NewSessionRepository        = iam.NewSessionRepository
	NewSessionService           = iam.NewSessionService
	NewSessionServiceWithConfig = iam.NewSessionServiceWithConfig

	// SSOService
	NewSSORepository = iam.NewSSORepository
	NewSSOService    = iam.NewSSOService

	// APIKeyService
	NewAPIKeyRepository = iam.NewAPIKeyRepository
	NewAPIKeyService    = iam.NewAPIKeyService

	// Seeding helpers
	SeedDefaultRoles = iam.SeedDefaultRoles
	AssignAdminRole  = iam.AssignAdminRole
)

// ============================================================================
// PolicyChecker — for handler-layer per-resource permission filtering
//
// Handlers that need to check permissions on individual resources (e.g. the
// /schema/boot nav-filter) MUST use this interface instead of calling
// AuthzService.Enforce directly.  Direct .Enforce() calls are forbidden
// outside iam/ and middleware/ by the contract boundary guard.
// ============================================================================

// PolicyChecker checks a single permission request.
// Obtain one via [NewPolicyChecker].
type PolicyChecker interface {
	Check(ctx context.Context, r Request) (bool, error)
}

type policyCheckerAdapter struct{ svc AuthzService }

func (a *policyCheckerAdapter) Check(ctx context.Context, r Request) (bool, error) {
	return a.svc.Enforce(ctx, r)
}

// NewPolicyChecker wraps svc behind PolicyChecker.
// Returns nil when svc is nil (callers should guard).
func NewPolicyChecker(svc AuthzService) PolicyChecker {
	if svc == nil {
		return nil
	}
	return &policyCheckerAdapter{svc: svc}
}
