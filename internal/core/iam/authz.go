package iam

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

// Service is the single entry point for all authorization operations.
// Callers should depend on this interface, not the concrete *service.
type Service interface {
	// Enforcement

	// Enforce checks whether subject may perform action on object within domain.
	Enforce(ctx context.Context, r Request) (bool, error)

	// EnforceBatch checks multiple requests in one call.
	EnforceBatch(ctx context.Context, reqs []Request) ([]bool, error)

	// Role management

	// AssignRole grants subject the named role in domain, recording metadata
	// in role_assignments and adding the g-rule to Casbin.
	AssignRole(ctx context.Context, tenantID, subject, role, domain string, opts ...AssignOpt) error

	// RevokeRole removes the role from Casbin and marks the assignment inactive.
	RevokeRole(ctx context.Context, subject, role, domain string) error

	// GetRoles returns all roles subject holds in domain.
	GetRoles(ctx context.Context, subject, domain string) ([]string, error)

	// HasRole reports whether subject holds role in domain.
	HasRole(ctx context.Context, subject, role, domain string) (bool, error)

	// GetAssignments returns the role_assignments metadata rows for subject+domain.
	GetAssignments(ctx context.Context, subject, domain string) ([]RoleAssignment, error)

	// Policy management (admin)

	// AddPolicy inserts a p-rule (allow or deny).
	AddPolicy(ctx context.Context, p Policy) error

	// RemovePolicy deletes a p-rule.
	RemovePolicy(ctx context.Context, p Policy) error

	// GetPolicies returns all p-rules for a given domain.
	GetPolicies(ctx context.Context, domain string) ([]Policy, error)

	// HTTP

	// Middleware returns a Fiber handler that enforces object+action for the
	// authenticated Principal stored in c.Locals(LocalsKeyPrincipal).
	Middleware(object, action string) fiber.Handler

	// Cache

	// InvalidateCache reloads the policy from the database, clearing any
	// in-memory cache that Casbin holds.
	InvalidateCache(ctx context.Context) error
}
