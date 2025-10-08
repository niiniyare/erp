# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

**Awo Enterprise Resource Planning System** is a sophisticated, multi-tenant ERP platform built with military-grade security and intelligent automation. The project combines enterprise security requirements with modern cloud architecture, delivering unparalleled access control, regulatory compliance, and operational excellence.

### Key Characteristics
- **Purpose**: Next-generation ERP with military-grade security
- **Architecture**: Clean Architecture with Domain-Driven Design (DDD)
- **Status**: 80% complete financial module, active development
- **Target**: Multi-tenant SaaS with white-label capabilities

## Important: API Framework Migration Status

⚠️ **CRITICAL: The project is transitioning from Goa framework to Fiber framework**

**Current State:**
- Goa v3 types and generated code are still being used
- Fiber framework is handling HTTP routing and middleware
- Hybrid architecture using Goa-generated types with Fiber handlers

**Questions for Assistant:**
1. **API Implementation Analysis**: Please analyze the current API implementation in `/internal/api/` to understand:
   - How Fiber is integrated with Goa-generated types
   - Current handler patterns and middleware setup
   - Request/response transformation patterns
   - Where Goa code generation is still utilized vs. pure Fiber implementation

2. **Migration Strategy**: Based on your analysis, document:
   - Which Goa components are still in use (types, validation, OpenAPI generation)?
   - How are Goa service designs being used with Fiber?
   - What migration patterns should be followed for new endpoints?
   - Should we continue using `make goa` or transition fully to Fiber?

3. **Documentation Updates**: After analysis, update this documentation to reflect:
   - Accurate command references (remove `make goa` if not used)
   - Correct API development workflow
   - Fiber-specific patterns and best practices
   - How to handle Goa-generated types in Fiber handlers

**Please analyze the codebase and provide clarification on these points before making significant API-related changes.**

## Technology Stack & Architectural Decisions

### Backend Architecture

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Language | Go 1.25+ | High performance, excellent concurrency, strong typing |
| API Framework | **Fiber v2** | High performance, Express-like API, efficient routing |
| Type System | **Goa v3 (Generated Types)** | Type-safe request/response models, OpenAPI compliance |
| Database | PostgreSQL 15+ | ACID compliance, Row-Level Security, advanced features |
| Query Builder | SQLC | Type safety, performance, compile-time validation |
| Caching | Redis 7+ | Session management, performance optimization |
| Workflows | Temporal.io | Reliable business process automation |
| Observability | OpenTelemetry | Distributed tracing, metrics, monitoring |

### Frontend Architecture

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Templates | Templ | Type-safe HTML generation, compile-time validation |
| JavaScript | Alpine.js | Lightweight reactivity, server-first approach |
| HTMX | HTMX | Server-driven interactions, minimal JavaScript |
| CSS Framework | TailwindCSS | Utility-first, design consistency |
| Components | Flowbite | Professional UI components, design system |

## Project Structure

### Directory Layout

```
project/erp/
├── cmd/
│   ├── server/          # Main application entry point
│   ├── migrate/         # Database migration CLI
│   └── worker/          # Background job worker (Temporal)
├── internal/
│   ├── api/             # API layer
│   │   ├── design/      # Goa DSL service designs (if still used)
│   │   ├── gen/         # Goa-generated types and code
│   │   ├── handlers/    # Fiber HTTP handlers
│   │   └── middleware/  # Fiber middleware (auth, tenant, logging)
│   ├── core/            # Business logic (domain layer)
│   │   ├── finance/     # Financial module (80% complete)
│   │   │   ├── domain/      # Entities, value objects, business rules
│   │   │   ├── service/     # Business logic (10 services, 7,256 lines)
│   │   │   └── repository/  # Data access (4,730 lines)
│   │   ├── abac/        # Attribute-Based Access Control (98% complete)
│   │   ├── iam/         # Identity & Access Management (95% complete)
│   │   ├── tenant/      # Multi-tenant management
│   │   └── audit/       # Audit & compliance (100% complete)
│   ├── platform/        # Infrastructure layer
│   │   ├── database/    # PostgreSQL connection, transactions
│   │   ├── cache/       # Redis caching
│   │   ├── temporal/    # Workflow client
│   │   └── observability/ # Tracing, metrics, logging
│   └── shared/          # Common utilities
│       ├── errors/      # Error types and handling
│       ├── format/      # Data formatting utilities
│       └── validation/  # Input validation
├── db/
│   ├── migration/       # SQL migration files (versioned)
│   ├── queries/         # SQLC query definitions
│   └── sqlc/           # Generated SQLC code
├── web/                 # Frontend components
│   ├── components/      # Templ components (atomic design)
│   ├── static/         # Static assets (CSS, JS, images)
│   └── templates/      # Page templates
└── docs/               # Comprehensive documentation
    ├── architecture/   # Architecture guides
    ├── modules/        # Module specifications
    └── api/           # API documentation
```

## Essential Commands

### Development Workflow

```bash
# Complete CI pipeline (run before commits)
make ci                     # Run fmt, lint, test, proto, sqlc

# Individual checks
make fmt                    # Format Go code with gofmt
make lint                   # Lint with golangci-lint
make vet                    # Run go vet
```

### Build and Run

```bash
make run                    # Start development server
make build                  # Build production binary
make server                 # Run compiled server binary

# ⚠️ IMPORTANT: Do not start the server in Claude Code sessions
# Let the user manually start the server to avoid port management issues
```

### Database Management

```bash
# Setup
make createdb              # Create 'erp' database
make dropdb                # Drop 'erp' database

# Migrations
make migrateup             # Apply all pending migrations
make migratedown           # Rollback last migration
make migrateup1            # Apply next migration only
make migratedown1          # Rollback one migration only

# Query code generation
make sqlc                  # Generate type-safe Go from SQL queries
```

### Code Generation

```bash
# ⚠️ NOTE: Verify which of these commands are still in use after Fiber migration
make proto                 # Generate gRPC, gateway files (if used)
make goa                   # Generate Goa service code (verify usage)
make mock                  # Generate gomock interfaces
make gen                   # Run all generators
```

### Testing

```bash
make test                  # Run all tests
make test-unit            # Unit tests only (no integration)
make test-integration     # Integration tests (requires database)
make test-coverage        # Generate coverage report (HTML)
make test-verbose         # Run tests with verbose output
```

### Maintenance

```bash
make clean                # Remove generated files and binaries
make deps                 # Download and verify dependencies
make tidy                 # Clean up go.mod and go.sum
```

## Architecture Deep Dive

### 1. Multi-Tenant Architecture with PostgreSQL RLS

#### Overview

The ERP system implements **shared database, shared schema** multi-tenancy with PostgreSQL Row-Level Security (RLS) as the primary isolation mechanism.

**Key Benefits:**
- Database-level security enforcement
- Automatic tenant filtering without application-level logic
- Zero-trust security model
- Performance optimization through intelligent indexing

#### Session Context Management

**PostgreSQL Session Variables:**
```sql
-- Set tenant context with validation
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_id UUID, user_role TEXT DEFAULT 'application_role')
  RETURNS VOID AS $$
DECLARE
  tenant_status TEXT;
BEGIN
  -- Validate tenant exists and is active
  SELECT status INTO tenant_status
  FROM tenants
  WHERE id = tenant_id AND deleted_at IS NULL;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Tenant not found: %', tenant_id;
  END IF;

  IF tenant_status != 'ACTIVE' THEN
    RAISE EXCEPTION 'Tenant is not active: % (status: %)', tenant_id, tenant_status;
  END IF;

  -- Set session variables
  PERFORM set_config('app.current_tenant_id', tenant_id::text, true);
  PERFORM set_config('app.tenant_status', tenant_status, true);
  PERFORM set_config('app.context_set_at', NOW()::text, true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Retrieve current tenant
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
BEGIN
  RETURN COALESCE(
    nullif(current_setting('app.current_tenant_id', FALSE), ''),
    NULL
  )::UUID;
EXCEPTION
  WHEN OTHERS THEN
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
```

#### RLS Policy Framework

**Policy Hierarchy:**
```sql
-- Enable RLS on tenant-related tables
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;

-- Application role: Tenant isolation
CREATE POLICY tenant_isolation_policy ON tenants
  FOR ALL TO application_role
  USING (id = current_tenant_id() OR current_tenant_id() IS NULL)
  WITH CHECK (id = current_tenant_id());

-- Admin role: Full cross-tenant access
CREATE POLICY admin_full_access_policy ON tenants
  FOR ALL TO admin_role
  USING (true)
  WITH CHECK (true);

-- Readonly role: Read access for reporting
CREATE POLICY readonly_access_policy ON tenants
  FOR SELECT TO readonly_role
  USING (true);
```

#### Role Management

**Database Roles:**
- `application_role`: Standard tenant-isolated access
- `admin_role`: Cross-tenant administrative access  
- `readonly_role`: Read-only access for analytics

```sql
-- Role creation with permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON tenants TO application_role;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO admin_role;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly_role;

-- Function execution permissions
GRANT EXECUTE ON FUNCTION set_tenant_context(UUID, TEXT) TO application_role;
GRANT EXECUTE ON FUNCTION current_tenant_id() TO application_role;
```

#### Go Service Integration

**Tenant Service (`internal/core/tenant/service.go`):**
```go
// Service interface
type Service interface {
    // Context management
    SetTenant(ctx context.Context, tenantID uuid.UUID) error
    ResetTenant(ctx context.Context) error
    ValidateCurrentTenant(ctx context.Context) error
    WithTenantContext(ctx context.Context, tenantID uuid.UUID, fn func(context.Context) error) error
    
    // Tenant operations
    GetCurrentTenant(ctx context.Context) (*Tenant, error)
    CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error)
}

// SetTenant establishes database tenant context
func (s *service) SetTenant(ctx context.Context, tenantID uuid.UUID) error {
    // Validate tenant exists
    exists, err := s.ExistsTenant(ctx, tenantID)
    if err != nil {
        return fmt.Errorf("failed to validate tenant: %w", err)
    }
    if !exists {
        return &sharedErrors.NotFoundError{Resource: "tenant", ID: tenantID.String()}
    }
    
    // Set database session context
    if err := s.repo.SetTenantContext(ctx, tenantID); err != nil {
        return fmt.Errorf("failed to set tenant context: %w", err)
    }
    
    return nil
}
```

#### SQLC Integration

**Type-Safe Queries (`db/queries/tenant.sql`):**
```sql
-- name: SetTenantContext :exec
SELECT set_tenant_context($1);

-- name: GetCurrentTenant :one
SELECT * FROM tenants
WHERE id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetTenantConfiguration :one
SELECT * FROM tenant_configurations
WHERE tenant_id = current_tenant_id();
```

**Generated Go Code:**
```go
// SQLC generates type-safe repository methods
func (q *Queries) SetTenantContext(ctx context.Context, tenantID uuid.UUID) error
func (q *Queries) GetCurrentTenant(ctx context.Context) (Tenant, error)
func (q *Queries) GetTenantConfiguration(ctx context.Context) (TenantConfiguration, error)
```

#### Request Flow

```
HTTP Request
    ↓
Fiber Middleware: Extract Tenant (header/subdomain)
    ↓
Validate Tenant ID
    ↓
Set Database Context (set_tenant_context)
    ↓
Execute Business Logic
    ↓
SQLC Queries (automatically filtered by RLS)
    ↓
PostgreSQL RLS Filtering
    ↓
Return Filtered Results
    ↓
HTTP Response
```

**Middleware Implementation:**
```go
// Fiber middleware for tenant context
func TenantMiddleware(tenantService tenant.Service) fiber.Handler {
    return func(c *fiber.Ctx) error {
        ctx := c.UserContext()
        
        // Extract tenant from header or subdomain
        tenantIdentifier, err := extractTenantID(c)
        if err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": "Invalid tenant context",
            })
        }
        
        // Resolve and set tenant context
        tenantID, err := tenantService.ResolveTenantID(ctx, tenantIdentifier)
        if err != nil {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "Tenant not found",
            })
        }
        
        if err := tenantService.SetTenant(ctx, tenantID); err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "Failed to set tenant context",
            })
        }
        
        // Store in locals for downstream handlers
        c.Locals("tenant_id", tenantID)
        
        return c.Next()
    }
}

func extractTenantID(c *fiber.Ctx) (string, error) {
    // Priority 1: X-Tenant-ID header
    if tenantID := c.Get("X-Tenant-ID"); tenantID != "" {
        return tenantID, nil
    }
    
    // Priority 2: Subdomain extraction
    return extractFromSubdomain(c.Hostname())
}
```

### 2. Financial Module Architecture (80% Complete)

#### Implementation Status

**✅ Completed Components:**

1. **Database Schema** (5 migrations, ~1,200 lines SQL)
   - Chart of accounts with hierarchical structure
   - Transaction tables with double-entry constraints
   - Multi-currency support with exchange rates
   - Audit trail and versioning

2. **Domain Layer** (9 files, ~2,800 lines)
   - Rich entities: `Account`, `Transaction`, `JournalEntry`
   - Value objects: `Money`, `ExchangeRate`, `FiscalPeriod`
   - Business rule validators (30+ rules)

3. **Service Layer** (10 services, 7,256 lines)
   - `AccountService`: Chart of accounts management
   - `TransactionService`: Transaction lifecycle with state machine
   - `PostingService`: Double-entry posting engine
   - `ValidationService`: 30+ business rule validators
   - `ExchangeRateService`: Currency conversion engine
   - `ReconciliationService`: Bank reconciliation
   - `TaxService`: Tax calculation
   - `AuditService`: Comprehensive audit trail
   - `ReportingService`: Financial statement generation
   - `WorkflowService`: Temporal workflow integration

4. **Repository Layer** (4,730 lines)
   - Full SQLC integration with type safety
   - `ChartOfAccountsRepository`: 14 methods, 640 lines
   - `TransactionRepository`: 30+ methods, 936 lines
   - Domain type mappers: 15+ functions, 689 lines
   - Test suite: 1,865 lines

5. **API Layer** (100% complete)
   - 31 Finance endpoints implemented
   - ⚠️ **Question for Assistant**: Are these using Fiber handlers with Goa types?

**🚧 Remaining Work (20%):**

1. **Workflow Integration** (Next Priority)
   - Service layer integration with WorkflowOrchestrator
   - Temporal workflow implementation
   - Business process automation

2. **Testing & Production Readiness**
   - Comprehensive integration tests
   - Performance optimization
   - Security hardening

#### Features Implemented

- ✅ Transaction processing with state transitions
- ✅ Double-entry validation (debits = credits)
- ✅ Multi-currency with automatic conversion
- ✅ Transaction reversal and correction
- ✅ Hierarchical account structures
- ✅ Cost center/department/project dimensions
- ✅ Fiscal period management
- ✅ Audit trail with full history
- ✅ Bank reconciliation support
- ✅ Tax calculation framework

### 3. ABAC Security Engine (98% Complete)

**Attribute-Based Access Control** provides:
- Sub-50ms authorization decisions
- 25+ Policy Operators for complex logic
- Context-aware evaluation with risk assessment
- Temporal workflows for approval processes

**Key Components:**
- Policy engine with expression evaluation
- Attribute management (user, resource, environment)
- Role-based attribute assignment
- Policy evaluation caching

### 4. IAM System (95% Complete)

**Identity & Access Management:**
- User authentication (JWT-based)
- Session management with Redis
- Password policies and enforcement
- Multi-factor authentication support
- User provisioning and deprovisioning

### 5. Audit & Compliance (100% Complete)

**Comprehensive Audit System:**
- Database-level change tracking
- API request/response logging
- User action auditing
- Compliance report generation
- Immutable audit trail

## API Development Workflow

⚠️ **IMPORTANT: This section needs clarification after analyzing the Fiber/Goa integration**

### Questions for Assistant:

1. **Current API Pattern**: How are new endpoints currently being created?
   - Pure Fiber handlers?
   - Goa-generated types with Fiber routing?
   - Hybrid approach?

2. **Type Generation**: 
   - Are Goa DSL files still being maintained?
   - How are request/response types defined?
   - Is OpenAPI still generated from Goa?

3. **Adding New Endpoints**: What's the current pattern?
   ```go
   // Option A: Pure Fiber
   app.Post("/api/v1/resource", func(c *fiber.Ctx) error {
       // Handler implementation
   })
   
   // Option B: Goa types with Fiber
   app.Post("/api/v1/resource", func(c *fiber.Ctx) error {
       var req *gen.CreateResourceRequest // Goa-generated type
       // Handler implementation
   })
   
   // Option C: Something else?
   ```

**Please analyze the codebase and document the correct pattern here.**

## Development Best Practices

### Go Code Style

- **Naming**: Use descriptive names (avoid single letters except loop indices)
- **Functions**: Single responsibility, focused purpose
- **Interfaces**: Small, focused interfaces
- **Errors**: Always handle, wrap with context using `fmt.Errorf("context: %w", err)`
- **Comments**: Godoc format for exported symbols
- **Formatting**: Run `make fmt` before committing

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create account: %w", err)
}

// Good: Use custom error types for business logic
if account.IsArchived {
    return &sharedErrors.BusinessRuleError{
        Rule:    "account_active",
        Message: "Cannot modify archived account",
    }
}
```

### Repository Pattern

```go
// Repository interface in domain layer
type AccountRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
    Save(ctx context.Context, account *Account) error
}

// Implementation uses SQLC
type accountRepository struct {
    queries *db.Queries
}

func (r *accountRepository) GetByID(ctx context.Context, id uuid.UUID) (*Account, error) {
    dbAccount, err := r.queries.GetAccountByID(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, &sharedErrors.NotFoundError{Resource: "account", ID: id.String()}
        }
        return nil, fmt.Errorf("database error: %w", err)
    }
    
    return mapToAccount(dbAccount), nil
}
```

### Service Layer Pattern

```go
type AccountService struct {
    repo   AccountRepository
    logger *zap.Logger
    tracer trace.Tracer
}

func (s *AccountService) CreateAccount(ctx context.Context, req CreateAccountRequest) (*Account, error) {
    ctx, span := s.tracer.Start(ctx, "AccountService.CreateAccount")
    defer span.End()
    
    // 1. Validate input
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // 2. Business logic
    account := NewAccount(req)
    
    // 3. Save via repository
    if err := s.repo.Save(ctx, account); err != nil {
        return nil, fmt.Errorf("failed to save account: %w", err)
    }
    
    // 4. Return result
    return account, nil
}
```

### Comment Annotations

Use these when implementation is incomplete:

```go
// NOTE: This function assumes accounts are already validated
// TODO: Implement balance calculation for multi-currency accounts  
// FIXME: Race condition when updating concurrent transactions - needs transaction isolation
```

- **NOTE**: Informational context for future developers
- **TODO**: Non-critical feature or improvement needed
- **FIXME**: Critical issue requiring immediate attention

Always provide detailed expectations or fixes when using these annotations.

### Formatting Utilities

Always use shared formatting functions from `internal/shared/format/`:

```go
import "project/erp/internal/shared/format"

// Format money
formatted := format.Money(amount, currency)

// Format dates
dateStr := format.Date(time.Now())

// Format decimal numbers
decimalStr := format.Decimal(value, 2)
```

## Testing Strategy

### Test Pyramid

```
         /\        E2E Tests (few)
        /  \       - Full workflow tests
       /____\      - Docker compose environment
      /      \     Integration Tests (some)
     /________\    - Database interactions
    /          \   - External service calls
   /____________\  Unit Tests (many)
  /              \ - Business logic
 /________________\- Domain validation
```

### Unit Tests

```bash
make test-unit
```

- **Location**: `*_test.go` files alongside source
- **Focus**: Domain logic, business rules, pure functions
- **Mocking**: Use gomock for interfaces
- **Coverage**: Aim for >80% on critical paths

### Integration Tests

```bash
make test-integration
```

- **Location**: `*_integration_test.go` files
- **Build Tag**: `// +build integration`
- **Requirements**: Running PostgreSQL instance
- **Setup**: Use test fixtures and migrations

### Coverage Reports

```bash
make test-coverage  # Generates coverage.html
open coverage.html  # View in browser
```

### Test Database Setup

```bash
# Create test database
createdb erp_test

# Run migrations
DATABASE_URL="postgres://user:pass@localhost/erp_test?sslmode=disable" make migrateup

# Run tests
TEST_DATABASE_URL="postgres://user:pass@localhost/erp_test?sslmode=disable" make test-integration
```

## Commit Message Guidelines

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style (formatting, no logic change)
- `refactor`: Code refactoring (no functional change)
- `test`: Adding or updating tests
- `chore`: Maintenance, dependency updates

### Examples

```
feat(finance): add multi-currency transaction support

Implement currency conversion in transaction service using
exchange rate engine. Supports automatic rate lookup and
manual rate override.

Closes #123
```

```
fix(rls): resolve tenant isolation in account queries

Add missing tenant_id filter in SQLC query to ensure
proper multi-tenant data isolation.

Fixes #456
```

### Guidelines

- Write concise, human-readable messages
- Focus on "what" and "why", not "how"
- Use imperative mood ("add feature" not "added feature")
- Keep subject under 72 characters
- **Never mention AI assistance** (e.g., "claude")
- Reference issue numbers when applicable

## Server Management

### Important: Manual Server Control

**⚠️ NEVER start the server automatically in Claude Code sessions.**

```bash
# User should manually start server
cd project/erp
make run

# Or with custom port
PORT=8081 make run
```

**Reason**: In Termux and similar environments, killing processes on occupied ports is challenging. Always let the user manage server lifecycle.

### Default Ports

- ERP Server: 8080
- Frontend Dev: 5173
- gRPC Server: 9090 (if used)

### Stopping Servers

```bash
# Find process by port
lsof -i :8080

# Kill by PID
kill -9 <PID>

# Or use pkill
pkill -f "server"
```

## Common Tasks

### Adding a New Feature

**⚠️ IMPORTANT: Before implementing any new feature, you MUST:**

1. **Study Existing Documentation** - Read the comprehensive documentation in `docs/reference/modules/` for similar features
2. **Follow Established Patterns** - Use existing modules as reference (especially the financial module)
3. **Write Documentation First** - Create proper documentation before writing code

#### Example: Learning from the Financial Module

The financial module (`internal/core/finance/`) is the gold standard example with complete documentation:

**Available Documentation:**
```
docs/reference/modules/financial/
├── ACCOUNT_SERVICE_INTEGRATION_SUMMARY.md  # Service integration patterns
├── api-reference.md                        # Complete API documentation
├── architecture-guide.md                   # Architecture and design patterns
├── currency-management.md                  # Complex feature implementation
├── deployment-guide.md                     # Deployment considerations
├── financial-management.md                 # Business logic documentation
├── financial.sql                           # Database schema reference
├── integration-guide.md                    # Integration with other modules
├── PRD.md                                  # Product Requirements Document
├── README.md                               # Module overview
├── ROADMAP.md                              # Implementation roadmap
├── security-compliance-guide.md            # Security requirements
├── settings-integration.md                 # Configuration management
├── sql-ledger-user-guideV306.md           # User guide
├── TASK.md                                 # Task breakdown
├── testing-strategy.md                     # Testing approach
└── testing.md                              # Test implementation guide
```

**📚 MANDATORY: When implementing a new feature module, you MUST:**

1. **Read and understand** the financial module documentation structure
2. **Create similar documentation** for your new feature module
3. **Follow the same patterns** demonstrated in the financial module

#### Step-by-Step Feature Implementation

**Phase 1: Documentation and Planning (DO THIS FIRST!)**

Before writing any code, create these documents in `docs/reference/modules/{your_module}/`:

```bash
# Create module documentation directory
mkdir -p docs/reference/modules/{your_module}

# Required documents (based on financial module structure):
touch docs/reference/modules/{your_module}/README.md
touch docs/reference/modules/{your_module}/PRD.md
touch docs/reference/modules/{your_module}/architecture-guide.md
touch docs/reference/modules/{your_module}/ROADMAP.md
touch docs/reference/modules/{your_module}/api-reference.md
touch docs/reference/modules/{your_module}/integration-guide.md
touch docs/reference/modules/{your_module}/testing-strategy.md
touch docs/reference/modules/{your_module}/security-compliance-guide.md
```

**Document Contents (Reference Financial Module):**

1. **README.md** - Module overview, purpose, features
   - Study: `docs/reference/modules/financial/README.md`
   - Include: Overview, key features, quick start, architecture diagram

2. **PRD.md** - Product Requirements Document
   - Study: `docs/reference/modules/financial/PRD.md`
   - Include: Business requirements, user stories, acceptance criteria

3. **architecture-guide.md** - Technical architecture
   - Study: `docs/reference/modules/financial/architecture-guide.md`
   - Include: Layer architecture, design patterns, data flow, component diagram

4. **ROADMAP.md** - Implementation plan
   - Study: `docs/reference/modules/financial/ROADMAP.md`
   - Include: Milestones, phases, dependencies, completion criteria

5. **api-reference.md** - API specification
   - Study: `docs/reference/modules/financial/api-reference.md`
   - Include: Endpoints, request/response formats, error codes

6. **integration-guide.md** - Integration patterns
   - Study: `docs/reference/modules/financial/integration-guide.md`
   - Include: Dependencies, service interactions, event flows

7. **testing-strategy.md** - Testing approach
   - Study: `docs/reference/modules/financial/testing-strategy.md`
   - Include: Test pyramid, coverage goals, testing patterns

8. **security-compliance-guide.md** - Security requirements
   - Study: `docs/reference/modules/financial/security-compliance-guide.md`
   - Include: ABAC policies, data protection, audit requirements

**Phase 2: Database Schema**

```bash
# Create migration
migrate create -ext sql -dir db/migration -seq add_{your_feature}

# Document schema in docs/reference/modules/{your_module}/{your_module}.sql
# Reference: docs/reference/modules/financial/financial.sql
```

**Phase 3: SQLC Queries**

```bash
# Add queries to db/queries/{your_module}.sql
# Generate code
make sqlc

# Document queries in api-reference.md
```

**Phase 4: Domain Layer**

```bash
# Create directory structure
mkdir -p internal/core/{your_module}/domain
mkdir -p internal/core/{your_module}/service  
mkdir -p internal/core/{your_module}/repository
```

**Reference Implementation:**
```
internal/core/finance/
├── domain/
│   ├── account.go              # Entity with business logic
│   ├── transaction.go          # Aggregate root
│   ├── money.go                # Value object
│   ├── validation_rules.go     # Business rules
│   └── errors.go               # Domain-specific errors
├── service/
│   ├── account_service.go      # Business logic orchestration
│   ├── transaction_service.go  # Transaction management
│   └── validation_service.go   # Validation engine
└── repository/
    ├── account_repository.go   # Data access interface
    ├── transaction_repository.go
    └── mappers.go              # Domain/DB mapping
```

**Phase 5: Service Layer**

- Study: `internal/core/finance/service/` implementation
- Implement business logic following finance module patterns
- Use dependency injection
- Write comprehensive tests

**Phase 6: Repository Layer**

- Study: `internal/core/finance/repository/` implementation
- Implement repository interfaces
- Use SQLC-generated queries
- Add proper error handling

**Phase 7: API Layer**

- ⚠️ **Follow current Fiber/Goa pattern** (needs clarification)
- Study: `internal/api/handlers/` for existing patterns
- Implement handlers
- Add middleware as needed
- Update api-reference.md with actual endpoints

**Phase 8: Testing**

- Follow patterns in `internal/core/finance/service/*_test.go`
- Write unit tests for domain logic
- Write integration tests for repositories
- Write API tests for handlers
- Document in testing.md

**Phase 9: Integration**

- Wire dependencies in `cmd/server/`
- Update integration-guide.md with actual integration steps
- Add configuration documentation

**Phase 10: Documentation Review**

Before considering the feature complete:
- [ ] All documentation files created and complete
- [ ] Code matches architecture-guide.md
- [ ] API documentation matches implementation  
- [ ] Testing strategy implemented
- [ ] Security requirements addressed
- [ ] Integration guide validated

### Assistant Prompt: Feature Documentation Request

**When implementing a new feature, Claude should:**

```
Assistant, I need to implement a new feature module: [{MODULE_NAME}]

Before writing any code, please:

1. **Study the Financial Module Documentation**
   - Read all files in docs/reference/modules/financial/
   - Understand the documentation structure and depth
   - Note the patterns used for architecture, API, testing, etc.

2. **Create Initial Documentation Set**
   Based on the financial module structure, create:
   - README.md with module overview
   - PRD.md with business requirements  
   - architecture-guide.md with technical design
   - ROADMAP.md with implementation phases
   - api-reference.md with endpoint specifications
   - integration-guide.md with dependency map
   - testing-strategy.md with test approach
   - security-compliance-guide.md with security requirements

3. **Provide Implementation Plan**
   After documentation is ready:
   - Break down into implementation phases
   - Identify dependencies on existing modules
   - Estimate completion percentage for each phase
   - Suggest testing approach

4. **Code Implementation**
   Only after documentation approval:
   - Follow patterns from internal/core/finance/
   - Implement with same level of detail and testing
   - Update documentation as implementation progresses

Please start by analyzing the financial module documentation and 
proposing the documentation structure for [{MODULE_NAME}].
```

### Quick Reference: Financial Module Patterns

**Study these files for implementation patterns:**

| Pattern | Reference File | Key Concepts |
|---------|---------------|--------------|
| Entity Design | `internal/core/finance/domain/account.go` | Rich entities, encapsulation |
| Value Objects | `internal/core/finance/domain/money.go` | Immutability, validation |
| Business Rules | `internal/core/finance/domain/validation_rules.go` | Rule engine, validators |
| Service Layer | `internal/core/finance/service/account_service.go` | Orchestration, transactions |
| Repository | `internal/core/finance/repository/account_repository.go` | Data access, mapping |
| Testing | `internal/core/finance/service/*_test.go` | Mocking, test patterns |

**Documentation to Study:**

| Document | Purpose | Key Sections |
|----------|---------|--------------|
| architecture-guide.md | System design | Layers, patterns, diagrams |
| integration-guide.md | Module connections | Dependencies, events |
| testing-strategy.md | Test approach | Coverage, patterns |
| ROADMAP.md | Implementation plan | Phases, milestones |

### Example: Creating a New Inventory Module

**Step 1: Documentation Analysis**
```bash
# Read financial module docs
cat docs/reference/modules/financial/README.md
cat docs/reference/modules/financial/architecture-guide.md
cat docs/reference/modules/financial/PRD.md
# ... study all documents
```

**Step 2: Create Documentation Structure**
```bash
mkdir -p docs/reference/modules/inventory
cd docs/reference/modules/inventory

# Create all required documentation files
# Use financial module as template
```

**Step 3: Request Implementation Approval**
```
Show documentation to user for review before coding.
Get approval on architecture and roadmap.
```

**Step 4: Implement Following Financial Patterns**
```bash
# Only after documentation is approved
mkdir -p internal/core/inventory/{domain,service,repository}
# Follow exact patterns from internal/core/finance/
```

### Adding a Database Migration

```bash
# Create migration files
migrate create -ext sql -dir db/migration -seq descriptive_name

# Edit the generated files:
# - db/migration/XXXXXX_descriptive_name.up.sql
# - db/migration/XXXXXX_descriptive_name.down.sql

# Apply migration
make migrateup

# If needed, rollback
make migratedown1
```

## Troubleshooting

### SQLC Generation Fails

```bash
# Check sqlc.yaml configuration
cat sqlc.yaml

# Test query syntax
psql -d erp -f db/queries/your_file.sql

# Regenerate
make sqlc
```

### Migration Conflicts

```bash
# Check current version
migrate -path db/migration -database "$DATABASE_URL" version

# Force version (use with caution!)
migrate -path db/migration -database "$DATABASE_URL" force <version>

# Nuclear option: recreate database
make dropdb createdb migrateup
```

### Test Database Connection

```bash
# Test connection
psql -d erp_test -c "SELECT 1;"

# Run tests with verbose output
TEST_DATABASE_URL="..." make test-integration -v
```

### API Issues

⚠️ **Document Fiber-specific troubleshooting after analyzing implementation**

## Documentation Resources

### Internal Documentation

- **Architecture**: `docs/architecture/` - System design documents
- **Modules**: `docs/modules/` - Module specifications
- **API**: `docs/api/` - API documentation
- **Database**: Review migrations in `db/migration/`

### External Resources

- [Fiber Documentation](https://docs.gofiber.io/)
- [SQLC Documentation](https://docs.sqlc.dev/)
- [Temporal Documentation](https://docs.temporal.io/)
- [PostgreSQL RLS](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

## Roadmap to Version 1.0

### Current Status

- **Overall Progress**: ~80%
- **Multi-Tenant Core**: ✅ Production ready
- **ABAC Engine**: 🔄 98% complete (Phase 8 of 9)
- **IAM System**: 🔄 95% complete (Phase 2.2 of 3)
- **Audit & Compliance**: ✅ 100% complete
- **Finance Module**: 🔄 80% complete
- **Finance API**: ✅ 100% complete (needs Fiber integration verification)

### Critical Path to V1.0

**Phase 1: Workflow Integration** (Current Priority)
- Service layer integration with WorkflowOrchestrator
- Temporal workflow implementation for financial processes
- Business process automation
- Testing and validation

**Phase 2: API Migration Completion**
- ⚠️ Clarify Fiber/Goa integration strategy
- Complete any remaining endpoint migrations
- Update OpenAPI documentation
- API integration testing

**Phase 3: Testing & Hardening**
- Comprehensive test coverage (>80%)
- Performance optimization and load testing
- Security audit and penetration testing
- Documentation completion

### Success Metrics for V1.0

- ✅ Financial Module: Complete transaction processing
- ✅ Security: Sub-50ms authorization, full audit compliance
- ✅ Multi-Tenancy: Zero cross-tenant data leakage
- ✅ Performance: 10,000+ concurrent requests
- ✅ API: Complete, documented, business-focused endpoints

## Performance Optimization

### Database Performance

**Indexing Strategy for RLS:**
```sql
-- Optimize RLS policy performance
CREATE INDEX CONCURRENTLY idx_tenants_rls_active
  ON tenants(id)
  WHERE deleted_at IS NULL AND status = 'ACTIVE';

CREATE INDEX CONCURRENTLY idx_tenant_configs_rls
  ON tenant_configurations(tenant_id)
  WHERE tenant_id IS NOT NULL;

-- Partial indexes for common patterns
CREATE INDEX CONCURRENTLY idx_usage_stats_current_month
  ON tenant_usage_stats(tenant_id, period_start)
  WHERE period_start >= date_trunc('month', CURRENT_DATE);
```

**Connection Pool Configuration:**
```go
func configureDatabasePool() *pgxpool.Config {
    config, _ := pgxpool.ParseConfig(databaseURL)
    
    // Optimize for RLS workloads
    config.MaxConns = 50                    // Adequate pool size
    config.MinConns = 10                    // Maintain warm connections
    config.MaxConnLifetime = time.Hour      // Prevent context leakage
    config.MaxConnIdleTime = time.Minute*30 // Clean up idle connections
    
    // Ensure clean tenant context on new connections
    config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
        _, err := conn.Exec(ctx, "SELECT set_config('app.current_tenant_id', '', false)")
        return err
    }
    
    return config
}
```

### Caching Strategy

**Tenant Context Caching:**
```go
// Cache tenant resolution for 30 minutes
func (s *service) ResolveTenantID(ctx context.Context, identifier string) (uuid.UUID, error) {
    cacheKey := fmt.Sprintf("tenant:resolve:%s", identifier)
    
    // Check cache first
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
        if tenantID, err := uuid.Parse(cached); err == nil {
            return tenantID, nil
        }
    }
    
    // Resolve from database
    tenantID, err := s.repo.ResolveSubdomainToID(ctx, identifier)
    if err != nil {
        return uuid.Nil, err
    }
    
    // Cache successful resolution
    s.cache.Set(ctx, cacheKey, tenantID.String(), time.Minute*30)
    
    return tenantID, nil
}
```

## Security Considerations

### Injection Prevention

PostgreSQL functions use parameterized queries to prevent SQL injection:
```sql
-- Safe: Direct parameter binding
CREATE OR REPLACE FUNCTION secure_set_tenant_context(tenant_id_param UUID)
  RETURNS VOID AS $
BEGIN
  PERFORM set_config('app.current_tenant_id', tenant_id_param::text, true);
END;
$ LANGUAGE plpgsql SECURITY DEFINER;
```

### Context Tampering Prevention

```go
// Validate tenant context matches expected tenant
func (s *service) ValidateTenantContext(ctx context.Context, expectedTenant uuid.UUID) error {
    currentTenant, err := s.getCurrentTenantFromDB(ctx)
    if err != nil {
        return fmt.Errorf("failed to retrieve current tenant: %w", err)
    }
    
    if currentTenant == nil {
        return errors.New("no tenant context established")
    }
    
    if *currentTenant != expectedTenant {
        // Log potential security violation
        logger.WarnContext(ctx, "Tenant context mismatch detected", logger.Fields{
            "expected": expectedTenant,
            "actual":   *currentTenant,
        })
        
        return &sharedErrors.UnauthorizedError{
            Message: "Tenant context validation failed",
        }
    }
    
    return nil
}
```

### Audit Trail

```sql
-- Comprehensive audit logging
CREATE TABLE tenant_audit_log (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id TEXT NOT NULL,
    operation TEXT NOT NULL,
    table_name TEXT NOT NULL,
    old_values JSONB,
    new_values JSONB,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT
);

-- Audit trigger function
CREATE OR REPLACE FUNCTION audit_tenant_operations()
  RETURNS TRIGGER AS $
BEGIN
  INSERT INTO tenant_audit_log (
    tenant_id, user_id, operation, table_name,
    old_values, new_values
  ) VALUES (
    current_tenant_id(),
    current_user,
    TG_OP,
    TG_TABLE_NAME,
    CASE WHEN TG_OP = 'DELETE' THEN to_jsonb(OLD) ELSE NULL END,
    CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN to_jsonb(NEW) ELSE NULL END
  );
  
  RETURN CASE WHEN TG_OP = 'DELETE' THEN OLD ELSE NEW END;
END;
$ LANGUAGE plpgsql;
```

## Monitoring and Observability

### Metrics Collection

```go
type TenantMetrics struct {
    contextSetDuration    prometheus.Histogram
    contextValidationTime prometheus.Histogram
    rlsPolicyViolations   prometheus.Counter
    tenantSwitchFrequency prometheus.Counter
}

func (s *service) SetTenantWithMetrics(ctx context.Context, tenantID uuid.UUID) error {
    start := time.Now()
    defer func() {
        s.metrics.contextSetDuration.Observe(time.Since(start).Seconds())
    }()
    
    s.metrics.tenantSwitchFrequency.WithLabelValues(tenantID.String()).Inc()
    
    return s.SetTenant(ctx, tenantID)
}
```

### Distributed Tracing

```go
func (s *service) SetTenant(ctx context.Context, tenantID uuid.UUID) error {
    ctx, span := s.tracer.Start(ctx, "tenant.service.SetTenant")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("tenant.id", tenantID.String()),
        attribute.String("operation", "set_tenant_context"),
    )
    
    // Implementation...
    
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "Failed to set database context")
        return err
    }
    
    span.SetStatus(codes.Ok, "Tenant context set successfully")
    return nil
}
```

## Advanced Patterns

### Scoped Tenant Context

```go
// WithTenantContext provides scoped tenant context execution
func (s *service) WithTenantContext(ctx context.Context, tenantID uuid.UUID, fn func(context.Context) error) error {
    // Store current context for restoration
    currentTenantID, err := s.getCurrentTenantFromDB(ctx)
    if err != nil {
        logger.WarnContext(ctx, "Could not retrieve current tenant context", logger.Fields{
            "error": err.Error(),
        })
    }
    
    // Set new tenant context
    if err := s.SetTenant(ctx, tenantID); err != nil {
        return fmt.Errorf("failed to set tenant context: %w", err)
    }
    
    // Execute function with tenant context
    defer func() {
        // Restore previous context or reset
        if currentTenantID != nil {
            if restoreErr := s.SetTenant(ctx, *currentTenantID); restoreErr != nil {
                logger.ErrorContext(ctx, "Failed to restore tenant context", logger.Fields{
                    "previous_tenant": currentTenantID,
                    "error":          restoreErr.Error(),
                })
            }
        } else {
            if resetErr := s.ResetTenant(ctx); resetErr != nil {
                logger.ErrorContext(ctx, "Failed to reset tenant context", logger.Fields{
                    "error": resetErr.Error(),
                })
            }
        }
    }()
    
    return fn(ctx)
}
```

### Multi-Tenant Batch Operations

```go
// ProcessTenantsInBatch executes operations for multiple tenants
func (s *service) ProcessTenantsInBatch(ctx context.Context, tenantIDs []uuid.UUID, fn func(context.Context, uuid.UUID) error) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(tenantIDs))
    
    for _, tenantID := range tenantIDs {
        wg.Add(1)
        go func(tid uuid.UUID) {
            defer wg.Done()
            
            // Execute with scoped tenant context
            if err := s.WithTenantContext(ctx, tid, func(tenantCtx context.Context) error {
                return fn(tenantCtx, tid)
            }); err != nil {
                errChan <- fmt.Errorf("tenant %s: %w", tid, err)
            }
        }(tenantID)
    }
    
    wg.Wait()
    close(errChan)
    
    // Collect errors
    var errs []error
    for err := range errChan {
        errs = append(errs, err)
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("batch operation failed: %v", errs)
    }
    
    return nil
}
```

## Migration Guides

### Adding RLS to Existing Tables

```sql
-- Step 1: Enable RLS
ALTER TABLE your_table ENABLE ROW LEVEL SECURITY;

-- Step 2: Add tenant_id column if not exists
ALTER TABLE your_table 
ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id);

-- Step 3: Backfill tenant_id for existing data (if applicable)
UPDATE your_table 
SET tenant_id = (SELECT id FROM tenants WHERE /* your criteria */);

-- Step 4: Make tenant_id NOT NULL
ALTER TABLE your_table 
ALTER COLUMN tenant_id SET NOT NULL;

-- Step 5: Create index
CREATE INDEX idx_your_table_tenant_id ON your_table(tenant_id);

-- Step 6: Create RLS policy
CREATE POLICY tenant_isolation_policy ON your_table
  FOR ALL TO application_role
  USING (tenant_id = current_tenant_id())
  WITH CHECK (tenant_id = current_tenant_id());
```

### Migrating from Goa to Fiber

⚠️ **This section should be completed after analyzing the codebase**

**Typical Migration Pattern:**
```go
// Before (Goa handler)
func (h *Handler) CreateResource(ctx context.Context, req *gen.CreateResourcePayload) (*gen.Resource, error) {
    // Implementation
}

// After (Fiber handler with Goa types)
func (h *Handler) CreateResource(c *fiber.Ctx) error {
    var req gen.CreateResourcePayload // Still using Goa-generated types
    
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    // Business logic
    result, err := h.service.CreateResource(c.UserContext(), &req)
    if err != nil {
        return handleError(c, err)
    }
    
    return c.Status(fiber.StatusCreated).JSON(result)
}
```

## FAQ

### Why use RLS instead of application-level filtering?

**Benefits of RLS:**
- **Security**: Database-level enforcement prevents bypass
- **Simplicity**: No need for explicit WHERE clauses in every query
- **Performance**: Database can optimize with tenant-aware indexes
- **Audit**: PostgreSQL logs provide complete access audit trail
- **Trust**: Zero-trust security model with defense in depth

### How do I test multi-tenant isolation?

```go
func TestTenantIsolation(t *testing.T) {
    // Create two tenants
    tenant1 := createTestTenant(t, "tenant1")
    tenant2 := createTestTenant(t, "tenant2")
    
    // Set context to tenant1
    err := service.SetTenant(ctx, tenant1.ID)
    require.NoError(t, err)
    
    // Create resource for tenant1
    resource1 := createTestResource(t, "resource1")
    
    // Switch to tenant2
    err = service.SetTenant(ctx, tenant2.ID)
    require.NoError(t, err)
    
    // Verify tenant2 cannot access tenant1's resource
    _, err = service.GetResource(ctx, resource1.ID)
    assert.Error(t, err)
    assert.True(t, errors.Is(err, sql.ErrNoRows))
}
```

### What happens if tenant context is not set?

- RLS policies with `current_tenant_id() IS NULL` check will allow access
- Most policies are designed to deny access when context is missing
- Middleware ensures context is always set for authenticated requests
- System operations can explicitly bypass RLS for maintenance

### How do I debug RLS issues?

```sql
-- Check current tenant context
SELECT current_tenant_id();

-- Check active RLS policies
SELECT schemaname, tablename, policyname, permissive, roles, qual, with_check
FROM pg_policies
WHERE tablename = 'your_table';

-- Disable RLS temporarily for debugging (superuser only)
SET row_security = OFF;
SELECT * FROM your_table; -- See all rows
SET row_security = ON;
```

### Can I use RLS with connection pooling?

**Yes, but with considerations:**
- Each connection maintains its own session variables
- Use `AfterConnect` hook to reset tenant context on new connections
- Alternatively, set context at the beginning of each transaction
- Current implementation uses transaction-scoped context (`true` parameter in `set_config`)

### Performance impact of RLS?

**Minimal with proper optimization:**
- RLS checks add ~1-2ms per query on average
- Proper indexing on `tenant_id` is critical
- PostgreSQL optimizer is aware of RLS and can use indexes efficiently
- Benchmark shows <5% overhead compared to explicit WHERE clauses

---

**Last Updated**: This document reflects the current state of the ERP project. Please verify API implementation details (Goa/Fiber integration) before making significant changes.

**Action Items for Assistant:**

1. **Critical Analysis Required**:
   - Analyze `/internal/api/` implementation thoroughly
   - Document the actual Fiber/Goa integration pattern being used
   - Identify which Goa components are still active (types, validation, OpenAPI generation)

2. **Documentation Updates**:
   - Update "API Development Workflow" section with accurate patterns
   - Clarify which `make` commands are still needed (proto, goa, gen)
   - Add Fiber-specific middleware and handler patterns
   - Document request/response transformation between Goa types and Fiber

3. **Migration Guide**:
   - Create clear guidelines for adding new Fiber endpoints
   - Document how to use Goa-generated types in Fiber handlers
   - Explain OpenAPI documentation generation (if still using Goa for this)

4. **Code Examples**:
   - Add actual examples from the codebase showing Fiber/Goa integration
   - Document error handling patterns specific to Fiber
   - Show middleware chain setup for tenant context, auth, etc.

**Questions to Answer**:
- Is `make goa` still generating code that's being used?
- Are Goa DSL files in `/internal/api/design/` still maintained?
- How is OpenAPI documentation being generated?
- What's the pattern for request validation (Goa's validation or custom)?
- Are gRPC endpoints still relevant or is it HTTP-only now?

Please analyze the codebase and update this documentation with your findings.
