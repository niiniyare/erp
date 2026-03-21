[<-- Back to Index](README.md)

## Code Architecture & Conventions

### Repository Layout

```
internal/
├── api/
│   ├── handlers/
│   │   ├── auth_handler.go        # POST /auth/login, /logout, /forgot-password
│   │   ├── user_handler.go        # CRUD /api/v1/users
│   │   ├── role_handler.go        # CRUD /api/v1/roles
│   │   ├── flag_handler.go        # GET/PATCH /api/v1/settings/flags
│   │   ├── settings_handler.go    # GET/PATCH /api/v1/settings/{module}
│   │   └── schema_handler.go      # GET /schema/boot, /schema/*
│   └── middleware/
│       ├── auth.go                # Authenticate — session cookie or Bearer
│       ├── tenant.go              # ResolveTenant — X-Tenant-ID or subdomain
│       ├── permission.go          # RequirePermission, RequireFlag
│       ├── db_pool.go             # SetDBPool
│       ├── audit.go               # AuditWrap
│       └── context.go             # ContextSession, ContextTenantID, helpers
│
└── platform/
    ├── service.go                 # Facade: Platform struct — single import for all modules
    ├── iam.go                     # IAM service composition
    ├── flags.go                   # FlagService — flag catalogue + tenant values
    ├── settings.go                # SettingService — setting catalogue + tenant values
    ├── boot.go                    # BootService — builds app shell from MRA+flags+perms
    │
    ├── domain/
    │   ├── user.go                # User, UserType, UserCreateParams
    │   ├── session.go             # ResolvedSession, SessionConfiguration
    │   ├── role.go                # Role, RoleAssignment, AssignOptions
    │   ├── permission.go          # Permission, PermissionKey
    │   ├── policy.go              # Policy, Effect
    │   ├── module.go              # Module, Resource, Action (MRA domain types)
    │   ├── flag.go                # FlagDefinition, TenantFlag
    │   ├── setting.go             # SettingDefinition, TenantSetting, UserPreference
    │   ├── entity.go              # Entity, EntityScope, EntityType
    │   └── events.go              # All platform domain events
    │
    └── repo/
        ├── user_repo.go           # interface + impl (sqlc + cache)
        ├── session_repo.go
        ├── role_repo.go
        ├── module_repo.go         # MRA queries — list enabled, nav data
        ├── flag_repo.go           # flag definitions + tenant values
        ├── setting_repo.go        # setting definitions + tenant values
        ├── entity_repo.go
        └── casbin_adapter.go      # Casbin ↔ PostgreSQL bridge

internal/core/authz/
├── authz.go        Service interface — the only file callers need to know
├── types.go        ActorType, Request, Policy, RoleAssignment, Principal
├── errors.go       Self-contained *Error type
├── model.go        Casbin CONF string
├── adapter.go      pgxAdapter — implements persist.BatchAdapter
├── service.go      New() constructor, Enforce, EnforceBatch, InvalidateCache
├── roles.go        AssignRole, RevokeRole, GetRoles, HasRole, revokeExpiredRoles
├── policies.go     AddPolicy, RemovePolicy, GetPolicies
└── middleware.go   Fiber handler factory (svc.Middleware)

db/
├── migration/                     # Sequential migration files
├── queries/                       # SQL for sqlc
│   ├── users.sql
│   ├── modules.sql                # MRA + nav queries
│   ├── flags.sql
│   ├── settings.sql
│   └── ...
└── sqlc/                          # Generated code (package: db)
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
