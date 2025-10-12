# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

**Awo Enterprise Resource Planning System** is a sophisticated, multi-tenant ERP platform built with military-grade security and intelligent automation.

### Key Characteristics
- **Purpose**: Next-generation ERP with military-grade security
- **Architecture**: Clean Architecture with Domain-Driven Design (DDD)
- **Status**: 80% complete financial module, active development
- **Target**: Multi-tenant SaaS with white-label capabilities

## API Architecture

**Current Implementation**: Hybrid architecture combining **Fiber v2** (HTTP routing/middleware) with **Goa v3** (type generation/OpenAPI).

**Key Components:**
- **Fiber v2**: HTTP routing, middleware, content negotiation
- **Goa v3**: Type-safe request/response models, OpenAPI generation
- **SQLC**: Type-safe database queries
- **PostgreSQL RLS**: Multi-tenant data isolation

## Technology Stack

**Backend**: Go 1.25+, Fiber v2, Goa v3, PostgreSQL 15+, SQLC, Redis 7+, Temporal.io
**Frontend**: Templ, Alpine.js, HTMX, TailwindCSS, Flowbite
**Security**: PostgreSQL RLS, ABAC engine, JWT authentication
**Observability**: OpenTelemetry, structured logging

## Project Structure

```
project/erp/
├── cmd/                 # Application entry points
├── internal/
│   ├── api/             # API layer (Fiber + Goa)
│   │   ├── design/      # Goa DSL service designs
│   │   ├── gen/         # Goa-generated types
│   │   ├── handlers/    # Fiber HTTP handlers
│   │   └── middleware/  # Fiber middleware
│   ├── core/            # Business logic (domain layer)
│   │   ├── finance/     # Financial module (80% complete)
│   │   ├── abac/        # Access control (98% complete)
│   │   ├── iam/         # Identity management (95% complete)
│   │   ├── tenant/      # Multi-tenant management
│   │   └── audit/       # Audit system (100% complete)
│   ├── platform/        # Infrastructure layer
│   └── shared/          # Common utilities
├── db/                  # Database (migrations, queries, SQLC)
├── web/                 # Frontend (Templ, static assets)
└── docs/               # Documentation
```

## Essential Commands

### Development Workflow

```bash
# Complete CI pipeline (run before commits)
make ci                     # Run fmt, lint, test, goa, sqlc

# Individual checks
make fmt                    # Format Go code with gofmt
make lint                   # Lint with golangci-lint
make test                   # Run all tests
```

### Build and Run

```bash
make run                    # Start development server
make build                  # Build production binary

# ⚠️ IMPORTANT: Do not start the server in Claude Code sessions
# Let the user manually start the server to avoid port management issues
```

### Database Management

```bash
# Setup
make createdb              # Create 'erp' database
make migrateup             # Apply all pending migrations
make migratedown1          # Rollback one migration

# Query code generation
make sqlc                  # Generate type-safe Go from SQL queries
```

### Code Generation

```bash
make goa                   # Generate Goa types and OpenAPI (ACTIVE)
make sqlc                  # Generate SQLC database code
make mock                  # Generate gomock interfaces
make templ                 # Generate Templ templates
```

### Module Scaffolding

```bash
# Module generation with documentation
make scaffold-module name=inventory docs=true tests=true

# Component generation
make scaffold-component module=finance name=payment type=entity

# Standalone documentation generation
make scaffold-docs name=finance

# Show scaffolding help
make scaffold-help
```

### Direct awoctl Commands

```bash
# Module generation
awoctl new module inventory --with-docs --with-tests --verbose

# Feature generation  
awoctl new feature finance/budgets

# Component generation
awoctl new component finance payment entity

# Documentation generation
awoctl docs module finance --verbose

# Help and options
awoctl --help
awoctl new --help
awoctl docs --help
```

## Multi-Tenant Architecture

**Core Concept**: Shared database with PostgreSQL Row-Level Security (RLS) for tenant isolation.

### Key Features
- **Database-level security**: RLS prevents cross-tenant data access
- **Session context**: PostgreSQL functions manage tenant context
- **Automatic filtering**: No explicit WHERE clauses needed in queries
- **Performance**: Tenant-aware indexes optimize queries

### Request Flow
```
HTTP Request → Fiber Middleware → Extract Tenant → Set DB Context → Business Logic → SQLC Queries → RLS Filtering → Response
```

### Tenant Context Management
```sql
-- Set tenant context
SELECT set_tenant_context('tenant-uuid');

-- Get current tenant
SELECT current_tenant_id();
```

### SQLC Integration
```sql
-- Queries automatically filtered by RLS
-- name: GetAccounts :many
SELECT * FROM accounts;
-- Only returns accounts for current tenant
```

## Module Status

### Financial Module (80% Complete)
**Features**: Transaction processing, double-entry validation, multi-currency, hierarchical accounts, audit trail
**Components**: Domain layer, service layer (10 services), repository layer, API layer (31 endpoints)
**Next**: Temporal workflow integration

### ABAC Security Engine (98% Complete)
**Features**: Sub-50ms authorization, 25+ policy operators, context-aware evaluation
**Components**: Policy engine, attribute management, role-based assignment

### IAM System (95% Complete)
**Features**: JWT authentication, session management, password policies, MFA
**Components**: User management, authentication flows

### Audit & Compliance (100% Complete)
**Features**: Database change tracking, API logging, compliance reporting
**Components**: Immutable audit trail, structured logging

## API Development Workflow

**Current Pattern**: Hybrid architecture using **Fiber routing** with **Goa-generated types**.

### Adding New Endpoints

1. **Define in Goa DSL** (`/internal/api/design/services/`)
   ```go
   var _ = Service("resourceService", func() {
       Method("create", func() {
           Payload(CreateResourcePayload)
           Result(Resource)
           HTTP(func() {
               POST("/api/v1/resources")
           })
       })
   })
   ```

2. **Generate types** (`make goa`)
   - Creates types in `/internal/api/gen/`
   - Updates OpenAPI specification

3. **Implement unified handler**
   ```go
   func (h *Handler) CreateResource(c *fiber.Ctx) error {
       var payload gen.CreateResourcePayload
       if err := c.BodyParser(&payload); err != nil {
           return handleValidationError(c, err)
       }
       
       result, err := h.service.CreateResource(c.Context(), &payload)
       if err != nil {
           return handleServiceError(c, err)
       }
       
       return handleResponse(c, result)
   }
   ```

4. **Register routes** in Fiber router
   ```go
   api.Post("/api/v1/resources", handler.CreateResource)
   ```

### Benefits
- **Type safety**: Goa-generated types with validation
- **OpenAPI**: Automatic documentation generation
- **Performance**: Fiber's fast HTTP handling
- **Content negotiation**: JSON/HTML/HTMX fragment support

## Development Best Practices

### Code Style
- Use descriptive names, single responsibility functions
- Handle errors with context: `fmt.Errorf("context: %w", err)`
- Run `make fmt` before committing
- Use shared formatting functions from `internal/shared/format/`

### Architecture Patterns

**Repository Pattern**:
```go
type AccountRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
    Save(ctx context.Context, account *Account) error
}
```

**Service Pattern**:
```go
func (s *AccountService) CreateAccount(ctx context.Context, req CreateAccountRequest) (*Account, error) {
    // 1. Validate input
    // 2. Business logic
    // 3. Save via repository
    // 4. Return result
}
```

**Comment Annotations**:
- `NOTE:` Informational context
- `TODO:` Non-critical improvement needed
- `FIXME:` Critical issue requiring attention

## Testing Strategy

### Test Pyramid
```
    /\     E2E Tests (few)
   /  \    Integration Tests (some)
  /____\   Unit Tests (many)
```

### Commands
```bash
make test                  # Run all tests
make test-unit            # Unit tests only
make test-integration     # Database integration tests
make test-coverage        # Generate coverage report
```

### Test Setup
```bash
# Create test database
createdb erp_test
DATABASE_URL="postgres://user:pass@localhost/erp_test" make migrateup

# Run tests
TEST_DATABASE_URL="postgres://user:pass@localhost/erp_test" make test-integration
```

## Commit Guidelines

### Format
```
<type>(<scope>): <subject>

<body>
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `refactor`: Code refactoring
- `test`: Tests
- `chore`: Maintenance

### Example
```
feat(finance): add multi-currency transaction support

Implement currency conversion using exchange rate engine.
Supports automatic rate lookup and manual override.
```

### Guidelines
- Keep messages concise and human-readable
- Use imperative mood ("add" not "added")
- Focus on "what" and "why"
- Never mention AI assistance

## Server Management

**⚠️ IMPORTANT**: Never start the server automatically in Claude Code sessions. Let the user manage server lifecycle.

```bash
# User should manually start server
make run

# Custom port
PORT=8081 make run
```

**Default Ports**: ERP Server (8080), Frontend Dev (5173)

**Stopping Servers**:
```bash
lsof -i :8080          # Find process
kill -9 <PID>          # Kill by PID
pkill -f "server"       # Kill by name
```

## Common Tasks

### Adding a New Feature

**Before implementing any new feature:**
1. Study existing documentation in `docs/reference/modules/`
2. Follow established patterns (especially the financial module)
3. Write documentation first

**Feature Implementation Process:**
1. **Documentation**: Create module documentation structure
2. **Database**: Create migration and SQLC queries
3. **Domain**: Implement entities and business rules
4. **Service**: Business logic orchestration
5. **Repository**: Data access with SQLC
6. **API**: Goa DSL + Fiber handlers
7. **Testing**: Unit, integration, and API tests
8. **Integration**: Wire dependencies

**Reference Patterns**:
- Study `internal/core/finance/` for implementation patterns
- Follow Clean Architecture layers
- Use established error handling and validation patterns

### Adding a Database Migration

```bash
# Create migration
migrate create -ext sql -dir db/migration -seq descriptive_name

# Apply migration
make migrateup

# Rollback if needed
make migratedown1
```

## Troubleshooting

### SQLC Issues
```bash
cat sqlc.yaml                              # Check configuration
psql -d erp -f db/queries/your_file.sql    # Test query syntax
make sqlc                                  # Regenerate
```

### Migration Issues
```bash
migrate -path db/migration -database "$DATABASE_URL" version  # Check version
make dropdb createdb migrateup                               # Nuclear reset
```

### Test Database
```bash
psql -d erp_test -c "SELECT 1;"                    # Test connection
TEST_DATABASE_URL="..." make test-integration -v   # Verbose tests
```

## Documentation Resources

**Internal**: `docs/architecture/`, `docs/modules/`, `docs/api/`, `db/migration/`
**External**: [Fiber](https://docs.gofiber.io/), [SQLC](https://docs.sqlc.dev/), [Temporal](https://docs.temporal.io/), [PostgreSQL RLS](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)

## Roadmap to V1.0

### Current Status (~80% Complete)
- **Multi-Tenant Core**: ✅ Production ready
- **ABAC Engine**: 🔄 98% complete
- **IAM System**: 🔄 95% complete
- **Audit & Compliance**: ✅ 100% complete
- **Finance Module**: 🔄 80% complete

### Critical Path
1. **Workflow Integration**: Temporal workflow implementation
2. **Testing & Hardening**: >80% test coverage, performance optimization
3. **Documentation**: Complete API and module documentation

### V1.0 Success Metrics
- Complete financial transaction processing
- Sub-50ms authorization, full audit compliance
- Zero cross-tenant data leakage
- 10,000+ concurrent requests
- Complete, documented APIs

---

**This document provides essential guidance for working with the ERP codebase. For detailed architecture information, see `docs/architecture/` and `docs/modules/`.**
