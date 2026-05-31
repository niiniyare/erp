[<-- Back to Index](README.md)

## Code Architecture & Conventions

### Repository Layout

> **Note:** The `internal/platform/` facade described in earlier docs **does not exist**. The actual code lives in `internal/core/iam/` (IAM bounded context) and separate sibling bounded contexts. Import from `internal/core/iam` — it re-exports everything callers need.

```
internal/
├── api/
│   ├── handlers/
│   │   ├── auth/           # login, logout, MFA, password reset, SSO, API keys
│   │   ├── audit/          # GET /api/v1/audit-logs
│   │   ├── schema/         # GET /api/v1/schema/boot (AMIS app shell)
│   │   ├── tenant/         # CRUD /api/v1/tenants
│   │   ├── user/           # CRUD /api/v1/users
│   │   ├── finance/        # /api/v1/finance/*
│   │   └── routes.go       # Router, Dependencies wiring
│   └── middleware/
│       ├── session_middleware.go  # Authenticate — session cookie or Bearer eak_ token
│       ├── whitelist.go           # Public endpoints (no session required)
│       ├── observability.go       # Logging + tracing + metrics per request
│       └── cors.go
│
└── core/
    ├── iam/                # IAM bounded context facade (import this)
    │   ├── iam.go          # Re-exports: UserService, SessionService, AuthzService, etc.
    │   ├── domain/
    │   │   ├── authz.go    # ActorType, CasbinModel, Policy, RoleAssignment
    │   │   ├── session.go  # ResolvedSession, Can(), CanDo(), FeatureEnabled()
    │   │   ├── identity.go # User, AccountStatus, MFASetup
    │   │   └── apikey.go   # APIKey, CreateAPIKeyRequest
    │   ├── repository/     # authz, user, session, sso, apikey repos
    │   └── service/        # authz, user, session, sso, mfa_totp, apikey services
    │
    ├── entity/             # Entity hierarchy bounded context (separate from IAM)
    │   ├── entity.go       # Facade: NewService, NewRepository
    │   ├── domain/
    │   ├── repository/
    │   └── service/
    │
    ├── audit/              # Audit log bounded context (separate from IAM)
    │   ├── interface.go    # Service + Repository interfaces
    │   ├── model.go        # AuditEvent, CreateAuditEventRequest, analytics types
    │   ├── service.go      # CreateAuditEvent, GetAuditEvents, analytics
    │   ├── repository.go   # DB adapter using sqlc
    │   └── validation.go   # Field-level validation helpers
    │
    └── tenant/             # Tenant bounded context
        └── ...

db/
├── migration/             # Sequential numbered migration files (000001_…)
├── queries/               # SQL source files for sqlc
│   ├── sessions.sql
│   ├── authz.sql
│   ├── users.sql
│   ├── api_keys.sql
│   ├── audit.sql
│   ├── boot.sql           # MRA nav queries (ListActiveSystemModules, etc.)
│   └── ...
└── sqlc/                  # Generated Go code (package: db) — DO NOT EDIT
```

Unit tests live next to every file they test.

---

### Repository Layer Convention

```
Service  →  Repository interface  →  Repository implementation  →  sqlc generated code
                                         (owns cache logic)
```

Services call repository methods. Repositories use sqlc internally and own all cache logic. Services never write SQL or set cache keys.

```go
// internal/platform/repo/user_repo_impl.go

type userRepoImpl struct {
    q     *db.Queries    // sqlc — only place raw queries appear
    cache *redis.Client  // cache: the repo's responsibility, not the service's
}

func (r *userRepoImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    cacheKey := "user:" + id.String()
    if raw, err := r.cache.Get(ctx, cacheKey).Bytes(); err == nil {
        var u domain.User
        json.Unmarshal(raw, &u)
        return &u, nil
    }
    row, err := r.q.GetUserByID(ctx, id) // sqlc call
    if err == pgx.ErrNoRows { return nil, domain.ErrNotFound }
    user := mapRowToUser(row)
    data, _ := json.Marshal(user)
    r.cache.Set(ctx, cacheKey, data, 5*time.Minute)
    return user, nil
}
```

---

### Function Parameter Convention

All service and repository methods use a single params struct:

```go
// DO — struct parameter; adding a field never breaks existing callers
func (s *IAMService) CreateUser(ctx context.Context,
    params domain.UserCreateParams) (*domain.User, error)

// DON'T — breaks call sites when new fields are needed
func (s *IAMService) CreateUser(ctx context.Context,
    email, displayName string, userType domain.UserType) (*domain.User, error)
```

---

### Context Carries Tenant and User Identity

Middleware injects `tenant_id`, `user_id`, and `entity_id` via `context.WithValue`. Services read them without explicit parameters:

```go
// Set by middleware
ctx = context.WithValue(ctx, contextKeyTenantID, tenantID)
ctx = context.WithValue(ctx, contextKeyUserID,   userID)
ctx = context.WithValue(ctx, contextKeyEntityID, entityID)

// Typed accessors (domain package)
func TenantIDFromContext(ctx context.Context) (uuid.UUID, bool)
func UserIDFromContext(ctx context.Context)   (uuid.UUID, bool)
func EntityIDFromContext(ctx context.Context) (uuid.UUID, bool)

// Service reads without explicit parameter
func (s *IAMService) ListUsers(ctx context.Context,
    params domain.UserListParams) ([]*domain.User, int, error) {
    tenantID, ok := domain.TenantIDFromContext(ctx)
    if !ok { return nil, 0, domain.ErrMissingTenantContext }
    params.TenantID = tenantID
    return s.repo.List(ctx, params)
}
```

| Data | How to pass |
|---|---|
| `tenant_id` for query scoping | Context — set by auth middleware |
| `user_id` of the *acting* user (for audit) | Context — set by auth middleware |
| `user_id` of the *target* user (e.g., assign role to this user) | Explicit in params struct |
| Business parameters | Explicit params struct |

---

### Tenant Identification

Tenants have three identifiers. UUID is canonical. Slug and subdomain resolve to UUID at the edge:

```go
func ResolveTenant(tenantRepo TenantRepository) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Path 1: X-Tenant-ID header (API clients, frontend)
        if headerID := c.Get("X-Tenant-ID"); headerID != "" {
            id, err := uuid.Parse(headerID)
            if err != nil { return c.Status(400).JSON(response.Err("invalid X-Tenant-ID")) }
            return injectAndContinue(c, id)
        }
        // Path 2: Subdomain (browser acme.awoerp.com)
        if t, err := tenantRepo.GetBySubdomain(c.Context(), c.Hostname()); err == nil {
            return injectAndContinue(c, t.ID)
        }
        // Path 3: Platform — no tenant required
        if s := middleware.ContextSession(c); s != nil && s.IsPlatform() {
            return c.Next()
        }
        return c.Status(403).JSON(response.Err("X-Tenant-ID header is required"))
    }
}
```

---

Next: [MRA Registry](./03b-mra-registry.md)
