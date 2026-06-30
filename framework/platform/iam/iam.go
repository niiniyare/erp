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
// # Authentication Flow (v2.1 — opaque sessions)
//
//  1. Client sends POST /auth/login with {email, password} + X-Awo-Tenant header.
//  2. AuthService verifies Argon2id password hash from tenant_users.
//  3. computeSnapshot walks active role assignments to build a permission snapshot.
//  4. Two opaque tokens issued: access (15 min) + refresh (7 days).
//     Format: awosess_tnt_<hex(32-byte CSPRNG)>
//  5. Access token stored in Redis (hot path); both tokens hashed and written
//     to tenant_sessions in PostgreSQL (durable audit copy).
//  6. Subsequent requests send "Authorization: Bearer <access_token>".
//  7. AuthMiddleware resolves token via Redis and sets a SessionViewer.
//  8. Framework handlers call ViewerFromCtx to get the authenticated principal.
//  9. POST /auth/refresh rotates both tokens and recomputes the permission snapshot.
//
// # Wiring
//
// Simplest — pass RedisClient to bootstrap.Mount and everything is automatic:
//
//	bootstrap.Mount(app, bootstrap.Options{
//	    Pool:        pool,
//	    RedisClient: redisClient,
//	})
//
// Manual wiring (if you need custom middleware ordering):
//
//	svc := iam.NewAuthService(pool, redisClient)
//	iam.NewHandler(svc).Mount(app)
//	app.Use(iam.AuthMiddleware(redisClient))
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
