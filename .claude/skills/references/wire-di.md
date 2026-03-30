# Wire Dependency Injection Reference

## Key facts

- Framework: `github.com/google/wire` v0.7.0 (compile-time codegen)
- Source file: `wire.go` (has `//go:build wireinject` tag — NEVER remove it)
- Generated file: `wire_gen.go` — do NOT edit manually, do NOT commit
- Regenerate: `make wire` → calls `./scripts/generate-wire.sh`

## Top-level Dependencies struct  (internal/api/handlers/routes.go)

```go
type Dependencies struct {
    Logger           logger.Logger
    Metrics          metrics.MetricsProvider
    Tracer           tracing.Service
    TenantService    coreTenant.Service
    UserService      iam.UserService
    FinanceServices  *financeService.Services
    TenantMiddleware fiber.Handler
    SecurityManager  *middlewarePkg.RouteSecurityManager
    SessionService   iam.SessionService
    AuthConfig       *middlewarePkg.AuthConfig
    SSOService       iam.SSOService
    APIKeyService    iam.APIKeyService
    Store            db.Store
    AuditService     audit.Service
}
```

`Dependencies.Validate()` panics if Logger, Metrics, or Tracer are nil.

## Provider set map

| Set name | Provides |
|---|---|
| `ObservabilitySet` | `logger.Logger`, `metrics.MetricsProvider`, `tracing.Service` |
| `DatabaseSet` | `db.Store` (pgx pool + SQLC wrapper) |
| `CacheSet` | `cache.Service` (Redis) |
| `IAMRepositorySet` | UserRepository, SessionRepository, AuthzRepository, SSORepository, APIKeyRepository |
| `IAMServiceSet` | UserService, SessionService, SSOService, APIKeyService |
| `TenantSet` | TenantRepository, TenantService |
| `AuditSet` | AuditRepository, AuditService |
| `FinanceSet` | AccountService, TransactionService, ReportingService → `*financeService.Services` |
| `MiddlewareSet` | TenantMiddleware, SecurityManager, AuthConfig |

## Adding a new domain (step-by-step)

### 1. Write the interface

```go
// internal/core/<domain>/repository/<noun>.go
type FooRepository interface {
    Create(ctx context.Context, params CreateFooParams) (*Foo, error)
    GetByID(ctx context.Context, id uuid.UUID) (*Foo, error)
    List(ctx context.Context, tenantID uuid.UUID) ([]*Foo, error)
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### 2. Write the SQLC adapter

```go
// internal/core/<domain>/repository/<noun>_sqlc.go
type fooSQLCRepository struct{ store db.Store }

func NewFooRepository(store db.Store) FooRepository {
    return &fooSQLCRepository{store: store}
}

func (r *fooSQLCRepository) Create(ctx context.Context, params CreateFooParams) (*Foo, error) {
    tenantID, ok := shared.GetTenantID(ctx)
    if !ok { return nil, ErrMissingTenant }
    var result *Foo
    err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
        row, err := s.CreateFoo(ctx, db.CreateFooParams{...})
        if err != nil { return parseFooDBError(err, "Create") }
        result = mapFoo(row)
        return nil
    })
    return result, err
}
```

### 3. Write the service

```go
// internal/core/<domain>/service.go
type FooService interface {
    Create(ctx context.Context, req CreateFooRequest) (*Foo, error)
}

type fooService struct {
    repo   FooRepository
    audit  audit.Service
    logger logger.Logger
}

func NewFooService(repo FooRepository, audit audit.Service, log logger.Logger) FooService {
    return &fooService{repo: repo, audit: audit, logger: log}
}
```

### 4. Create a provider set

```go
// internal/core/<domain>/wire.go  (no build tag — this is just providers)
var FooSet = wire.NewSet(
    NewFooRepository,
    wire.Bind(new(FooRepository), new(*fooSQLCRepository)),
    NewFooService,
    wire.Bind(new(FooService), new(*fooService)),
)
```

### 5. Add to Dependencies

```go
// internal/api/handlers/routes.go
type Dependencies struct {
    // ... existing fields
    FooService domain.FooService  // add here
}
```

### 6. Add to top-level injector in wire.go

```go
func InitializeApplication(ctx context.Context, cfg *config.Config) (*Application, func(), error) {
    wire.Build(
        // ... existing sets
        domain.FooSet,  // add here
        NewApplication,
    )
    return nil, nil, nil
}
```

### 7. Regenerate

```bash
make wire
```

## Wire rules

- Every `wire.Build` call must be in a file with `//go:build wireinject`
- Provider functions must return `(T, error)` or `(T, func(), error)` for cleanup
- Use `wire.Bind(new(Interface), new(*ConcreteType))` when injecting interfaces
- Use `wire.NewSet(...)` to group related providers — never wire individual funcs directly into Build
- Circular dependencies will fail at `make wire` — not at runtime
