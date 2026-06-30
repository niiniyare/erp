// Package iam implements Identity and Access Management for the tenant plane.
//
// # Two-Plane Architecture
//
// AwoERP operates two completely separate identity planes (see docs/framework/iam.md).
// This package implements PLANE 2 — Tenant Users only. Platform users (PLANE 1)
// are not yet implemented and are not required for tenant ERP functionality.
//
// # Entities
//
//   - tenant_user         — employees and service accounts belonging to one tenant
//   - role                — named permission bundle (system or tenant-defined)
//   - permission          — a permission string bound to a role (granted/denied)
//   - user_role_assignment — binds a user to a role, optionally within an OU
//
// All four entities are registered with the framework registry so that CRUD routes,
// SDUI pages, and audit logging are auto-generated.
//
// # Authentication Flow
//
//  1. Client sends POST /auth/login with {email, password} + X-Awo-Tenant header.
//  2. AuthService verifies Argon2id password hash from tenant_users.
//  3. computeSnapshot walks active role assignments to build a permission snapshot.
//  4. Session is created in Redis (TTL 8 hours).
//  5. Client receives an opaque 32-byte hex token.
//  6. Subsequent requests send "Authorization: Bearer <token>".
//  7. AuthMiddleware loads the session from Redis and sets a SessionViewer.
//  8. Framework handlers call ViewerFromCtx to get the authenticated principal.
//
// # Wiring
//
// Mount the middleware and handlers in your bootstrap:
//
//	svc := iam.NewAuthService(pool, redisClient)
//	iam.NewHandler(svc).Mount(app)
//	app.Use(iam.AuthMiddleware(redisClient))
//
//	// Replace the anonymous viewer with the IAM-backed one:
//	bootstrap.Mount(app, bootstrap.Options{
//	    Pool:     pool,
//	    ViewerFn: iam.ViewerFromCtx,
//	})
package iam

import (
	"awo.so/framework/def"
)

func init() {
	def.Register(&TenantUserDefinition)
	def.Register(&RoleDefinition)
	def.Register(&PermissionDefinition)
	def.Register(&UserRoleAssignmentDefinition)
}
