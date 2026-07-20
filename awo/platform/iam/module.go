// Package iam is the Awo Framework Identity and Access Management subsystem.
//
// IAM is a first-class framework package. It has zero dependency on any
// ERP-specific module, business entity, workflow, or application code. It is
// designed to be portable to any application built on the Awo Framework.
//
// # Entity model
//
// IAM manages the following tenant-scoped entities:
//
//   - iam_user: authenticated human principal (one per tenant account)
//   - iam_user_role: user-to-role assignment within a tenant
//   - iam_service_account: machine principal for API-to-API authentication
//   - iam_api_token: long-lived SHA-256-hashed API key for a service account
//   - iam_session: SQL audit trail (live state is in Redis, not here)
//   - iam_login_audit: tamper-evident append-only authentication event log
//
// Three global (non-tenant-scoped) tables are managed via migrations only
// and are NOT registered as EntityDefinitions:
//
//   - iam_roles: role catalogue seeded at bootstrap
//   - iam_permissions: permission identifier registry (compiler-populated)
//   - iam_role_permissions: role-to-permission bindings for CasbinEvaluator
//
// # Usage
//
//	m := iam.New(db, redisClient)
//	m.RegisterRoutes(app)
//	rolePerms, _ := m.Auth.LoadRolePermissions(ctx)
//	evaluator, _ := auth.NewCasbinEvaluator(schema.CapabilityGrants, rolePerms)
//
// # Framework boundaries
//
// This package is application-agnostic. It MUST NOT import:
//   - ERP modules (finance, inventory, CRM, etc.)
//   - ERP entity types or services
//   - ERP configuration
//   - ERP workflows or business logic
//
// Dependencies allowed: awo/def, awo/auth, awo/filter, awo/runtime, and
// third-party infrastructure libraries (Fiber, pgx, redis, bcrypt).
//
// # Migration location
//
// SQL migrations are in ./migrations/ relative to this package.
// They follow the golang-migrate naming convention:
// YYYYMMDDNNNNNN_description.{up,down}.sql
package iam

import (
	"awo.so/awo/def"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module bundles all IAM subsystem components into a single unit for
// framework-level registration. Construct with [New], then call
// [Module.RegisterRoutes] to attach HTTP endpoints to a Fiber application.
//
// Module is safe for concurrent use after [New] returns. Its fields are
// immutable after construction.
type Module struct {
	// Auth is the primary IAM service. Exposes Login, Logout,
	// RevokeUserSessions, and LoadRolePermissions for startup wiring.
	Auth *AuthService
}

// New constructs the IAM Module with its required runtime dependencies and
// wires them into the framework entity hook instances.
//
// New MUST be called before the Fiber application starts accepting requests.
// Calling New after the first HTTP request is a data race.
//
// deps:
//   - db:    PostgreSQL connection pool. Used for user lookup, role loading,
//             and audit trail writes.
//   - redis: Redis client. Used for session storage, user session index, and
//             API token caching.
func New(db *pgxpool.Pool, redis *redis.Client) *Module {
	// Wire the UserRoleChangeHook singleton with the Redis client.
	// The singleton is referenced by UserRoleDefinition.Hooks at init() time;
	// we populate its fields here before the first request is handled.
	userRoleChangeHook.Redis = redis

	return &Module{
		Auth: &AuthService{DB: db, Redis: redis},
	}
}

// RegisterRoutes attaches the IAM HTTP endpoints to the Fiber application.
//
// Routes registered (no authentication middleware on login):
//
//	POST  /api/v1/auth/login   — issue a session token
//	POST  /api/v1/auth/logout  — revoke the current session (requires session)
//	GET   /api/v1/auth/me      — return the current viewer's identity (requires session)
func (m *Module) RegisterRoutes(app *fiber.App) {
	registerRoutes(app, m.Auth)
}

func init() {
	def.Register(&UserDefinition)
	def.Register(&UserRoleDefinition)
	def.Register(&ServiceAccountDefinition)
	def.Register(&APITokenDefinition)
	def.Register(&SessionDefinition)
	def.Register(&LoginAuditDefinition)
}
