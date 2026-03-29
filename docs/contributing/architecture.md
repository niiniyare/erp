# Awo ERP Architecture Overview
## *Clean Architecture + Temporal Workflows + Module Facades*

*Modern ERP system architecture combining Clean Architecture principles with Temporal workflow orchestration and modular service design*

> ** Related Documentation:**  
> - `docs/contributing/service.md` - Service implementation patterns and module development  
> - `docs/contributing/goa.md` - API design and handler architecture  
> - `docs/contributing/database-transactions.md` - Database patterns and tenant isolation

## ️ System Architecture

Awo ERP follows **Clean Architecture with Temporal-First design**, emphasizing workflow orchestration, multi-tenant isolation, and modular service facades.

```
                    ┌─────────────────────────────────────────┐
                    │           External Actors               │
                    │  Web Browser  │  API Clients  │  CLI    │
                    └─────────────────┬───────────────────────┘
                                      │
                                      ▼
                    ┌─────────────────────────────────────────┐
                    │          API Layer (Goa)                │
                    │    internal/api/handlers/               │
                    │  ┌─────────────────────────────────────┐│
                    │  │ Module Entry Points                 ││
                    │  │ finance.go, iam.go, tenant.go      ││
                    │  │         │                           ││
                    │  │         ▼                           ││
                    │  │ ┌─────────────────┐ ┌─────────────┐ ││
                    │  │ │ finance/        │ │ iam/        │ ││
                    │  │ │ service_handler │ │ auth_handler│ ││
                    │  │ └─────────────────┘ └─────────────┘ ││
                    │  └─────────────────────────────────────┘│
                    └─────────────────┬───────────────────────┘
                                      │ (Primary Ports)
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Application Core                                     │
│                     internal/core/{module}/                                 │
│                                                                             │
│  ┌─────────────────┐ ┌─────────────────────────────────────────────────────┐ │
│  │ Module Facade   │ │                Service Layer                        │ │
│  │ service.go      │ │ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐   │ │
│  │                 │ │ │   Domain    │ │  Feature    │ │ Workflows & │   │ │
│  │ finance.Service │ │ │   Models    │ │  Services   │ │ Activities  │   │ │
│  │ iam.Service     ├─┼▶│ Entities &  │◀┤ Business   │◀┤  (Temporal) │   │ │
│  │ tenant.Service  │ │ │ Value Objs  │ │  Logic      │ │             │   │ │
│  │                 │ │ └─────────────┘ └─────────────┘ └─────────────┘   │ │
│  └─────────────────┘ └─────────────────────────────────────────────────────┘ │
│                                         │ (Secondary Ports)                 │
└─────────────────────────────────────────┼─────────────────────────────────────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    │                     │                     │
         ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
         │  Persistence     │  │   Temporal       │  │   Infrastructure │
         │  (Repository)    │  │   Workers        │  │   (Platform)     │
         │                  │  │                  │  │                  │
         │ SQLC + Context   │  │ Workflow/        │  │ Cache, Config,   │
         │ Tenant Isolation │  │ Activity Engine  │  │ Metrics, Tracing │
         └──────────────────┘  └──────────────────┘  └──────────────────┘
                    │                     │                     │
                    ▼                     ▼                     ▼
         ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
         │   PostgreSQL     │  │    Temporal      │  │ Redis, Observ.   │
         │   (RLS + SQLC)   │  │    Server        │  │ External APIs    │
         └──────────────────┘  └──────────────────┘  └──────────────────┘
```

##  Directory Structure

The architecture follows strict modular organization with clear separation of concerns:

```
internal/
├── api/                                # API Layer - External Interface
│   ├── design/                         # Goa API specifications
│   │   ├── design.go                   # Main API definition  
│   │   └── services/                   # Module service designs
│   ├── gen/                            # Generated Goa code
│   └── handlers/                       # HTTP request handlers
│       ├── {module}.go                 #  Module entry points (main imports)
│       ├── {module}/                   # Module handler implementations
│       │   ├── service_handler.go      # Core service handler
│       │   ├── types.go                # Request/response adapters
│       │   └── validation.go           # Input validation
│       └── common/                     # Shared handler utilities
│
├── core/                               # Application Core - Business Logic
│   ├── {module}/                       # Self-contained domain module
│   │   ├── service.go                  #  Module facade (external interface)
│   │   ├── domain/                     # Business entities and rules
│   │   │   ├── entities.go             # Core business entities
│   │   │   ├── types.go                # Value objects and enums
│   │   │   ├── constants.go            # Business configuration
│   │   │   └── validation.go           # Business validation rules
│   │   ├── {feature}/                  # Feature-specific submodules  
│   │   │   ├── service.go              # Feature service implementation
│   │   │   └── models.go               # Feature-specific models
│   │   ├── repository/                 # Data access layer
│   │   │   ├── interfaces.go           # Repository contracts
│   │   │   ├── {entity}.go             # SQLC-based implementations
│   │   │   └── mappers.go              # Domain ↔ DB conversions
│   │   ├── activities/                 # Temporal activities
│   │   │   ├── activity_registry.go    # Activity registration
│   │   │   └── {domain}_activities.go  # Domain-specific activities
│   │   └── workflows/                  # Temporal workflows
│   │       └── {process}_workflow.go   # Business process workflows
│   └── ...
│
├── platform/                           # Infrastructure Layer - External Concerns
│   ├── cache/                          # Redis caching implementation
│   ├── config/                         # Configuration management
│   ├── database/                       # Database connection & migrations
│   └── temporal/                       # Temporal client configuration
│
└── shared/                             # Cross-cutting Utilities
    ├── errors/                         # Error handling framework
    ├── logger/                         # Structured logging
    ├── metrics/                        # Metrics collection
    └── tracing/                        # Distributed tracing
```

##  Layer Responsibilities

### 1. **API Layer (`internal/api/`)**
*External interface for HTTP requests*

**Key Pattern: Modular Handler Architecture**
- **Entry Points**: `{module}.go` files export constructor functions
- **Handler Directories**: `{module}/` contains implementation details
- **Main Integration**: Only imports entry point functions

```go
// internal/api/handlers/finance.go
func NewFinanceHandler(
    financeService finance.Service,
    tracing tracing.TracingService,
    metrics metrics.MetricsProvider,
) goaFinance.Service {
    return finance.NewFinanceHandler(financeService, tracing, metrics)
}
```

**Responsibilities:**
- ✅ Handle HTTP request/response lifecycle
- ✅ Convert between API and domain models  
- ✅ Input validation and sanitization
- ✅ Call appropriate core services
- ✅ Format responses with proper status codes
- ❌ No business logic or data access

### 2. **Application Core (`internal/core/{module}/`)**
*Business logic and domain knowledge*

**Key Pattern: Module Facade Architecture**
- **Module Facade**: `service.go` provides unified external interface
- **Feature Submodules**: Organized by business capability
- **Temporal Integration**: Workflows and activities for complex processes

```go
// internal/core/finance/service.go
type Service interface {
    Account() AccountService
    Transaction() TransactionService
    Reports() ReportService
}

type service struct {
    accountSvc     AccountService
    transactionSvc TransactionService
    reportSvc      ReportService
    
    // Essential dependencies
    tenantService      tenant.Service
    auditService       audit.Service
    featureFlagService featureflag.Service
}
```

**Responsibilities:**
- ✅ Implement business rules and validation
- ✅ Orchestrate complex workflows (Temporal)
- ✅ Manage domain entity lifecycle
- ✅ Define repository contracts (interfaces)
- ✅ Maintain module boundaries
- ❌ No knowledge of HTTP, database specifics, or frameworks

### 3. **Repository Layer (`internal/core/{module}/repository/`)**
*Data persistence with tenant isolation*

**Key Pattern: Context-Based Tenant Resolution**
- Uses `shared.GetTenantID(ctx)` for automatic tenant isolation
- SQLC queries with `current_tenant_id()` function
- No explicit tenant ID parameters needed

```go
// All queries automatically tenant-isolated
func (r *repository) CreateAccount(ctx context.Context, req *domain.CreateAccountRequest) error {
    // Tenant ID resolved from context automatically
    return r.store.CreateAccount(ctx, sqlc.CreateAccountParams{
        Code: req.Code,
        Name: req.Name,
        // current_tenant_id() used in SQL query
    })
}
```

**Responsibilities:**
- ✅ Implement repository interfaces from core
- ✅ Convert between domain and database models
- ✅ Execute type-safe SQLC queries
- ✅ Handle tenant isolation automatically
- ✅ Manage database transactions
- ❌ No business logic or validation

### 4. **Temporal Layer (Workflows & Activities)**
*Long-running business processes*

**Key Pattern: Configuration-Driven Workflows**
- Activities depend on domain services via DI
- Workflows orchestrate complex business processes
- Configuration constants prevent hard-coded values

```go
// internal/core/finance/workflows/transaction_approval_workflow.go
func TransactionApprovalWorkflow(ctx workflow.Context, input *TransactionApprovalInput) error {
    // Configure timeouts from domain constants
    activityOptions := workflow.ActivityOptions{
        ScheduleToCloseTimeout: domain.DefaultApprovalTimeout,
        RetryPolicy: domain.DefaultRetryPolicy,
    }
    
    // Execute activities with dependency injection
    err := workflow.ExecuteActivity(activityCtx, a.ValidateTransaction, input.TransactionID).Get(ctx, nil)
    // ... orchestrate approval process
}
```

**Responsibilities:**
- ✅ Orchestrate long-running business processes
- ✅ Handle async operations and timeouts
- ✅ Manage complex state transitions
- ✅ Integrate with external systems
- ✅ Ensure process reliability and recovery
- ❌ No direct database access (use activities)

### 5. **Infrastructure Layer (`internal/platform/`)**
*Technical concerns and external integrations*

**Responsibilities:**
- ✅ Database connection management
- ✅ Caching infrastructure (Redis)
- ✅ Configuration loading and validation
- ✅ Observability (metrics, tracing, logging)
- ✅ External service integrations
- ❌ No business logic or domain knowledge

##  Dependency Flow & Module Isolation

### **Strict Dependency Direction**
All dependencies must point inward toward the application core:

```
External → API Handlers → Core Services → Repository Interfaces
                           ↑                        ↓
                    Module Facades           Repository Implementations
                           ↑                        ↓
                Essential Services            SQLC + Database
```

### **Module Isolation Rules**
1. **External Imports**: Only import module facades (`{module}.Service`)
2. **Internal Structure**: Features organized in submodules
3. **Essential Dependencies**: Each module has tenant, audit, featureflag services
4. **Cross-Module Communication**: Via events or direct service calls

### **Context Propagation Pattern**
Every operation includes context with tenant/entity/user isolation:

```go
func (s *service) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*Account, error) {
    // Context automatically contains:
    // - Tenant ID (for RLS)
    // - Entity ID (for company isolation)  
    // - User ID (for audit trails)
    // - Trace spans (for observability)
    
    return s.accountService.Create(ctx, req)
}
```

## ️ Architecture Benefits

### **1. Workflow-First Design**
- Complex business processes are explicitly modeled as workflows
- Async operations handled reliably with Temporal
- Clear process visibility and monitoring

### **2. Module Autonomy**
- Each module is self-contained with clear boundaries
- Easy to test, develop, and maintain independently  
- Consistent patterns across all business domains

### **3. Multi-Tenant by Design**
- Automatic tenant isolation at database level
- Context-driven data access with RLS policies
- No risk of cross-tenant data leakage

### **4. Configuration-Driven**
- Business rules and timeouts externalized
- Feature flags for controlled rollouts
- Environment-specific behavior without code changes

### **5. Enterprise-Grade Observability**
- Distributed tracing across all layers
- Structured logging with context
- Business metrics and operational dashboards

##  Anti-Patterns to Avoid

### ❌ **Breaking Module Boundaries**
```go
// DON'T: Direct import of internal module structure
import "internal/core/finance/account" // ❌ WRONG

// DO: Import module facade only  
import "internal/core/finance" // ✅ CORRECT
```

### ❌ **Hard-Coded Configuration**
```go
// DON'T: Hard-coded timeouts and values
timeout := 5 * time.Minute // ❌ WRONG

// DO: Use domain constants
timeout := domain.DefaultApprovalTimeout // ✅ CORRECT
```

### ❌ **Bypassing Workflows for Complex Operations**
```go
// DON'T: Synchronous multi-step operations
func (s *service) ProcessMonthEnd(ctx context.Context) error {
    s.calculateBalances(ctx)   // This could fail after 20 minutes
    s.generateReports(ctx)     // Leaving system in inconsistent state
    s.sendNotifications(ctx)   // ❌ WRONG
}

// DO: Use workflows for complex processes  
func (s *service) ProcessMonthEnd(ctx context.Context) error {
    return s.temporal.ExecuteWorkflow(ctx, workflows.MonthEndProcess, input)
} // ✅ CORRECT
```

### ❌ **Context-Free Operations**
```go
// DON'T: Skip context propagation
func (r *repo) GetAccount(id uuid.UUID) (*Account, error) // ❌ WRONG

// DO: Always include context
func (r *repo) GetAccount(ctx context.Context, id uuid.UUID) (*Account, error) // ✅ CORRECT
```

## ✅ Best Practices

1. **Start with Domain Design**: Define entities, value objects, and business rules first
2. **Implement Service Facades**: Create clean module interfaces before internal complexity
3. **Use Workflows for Complex Processes**: Model multi-step operations as explicit workflows
4. **Context Everywhere**: Every operation must accept and propagate context
5. **Configuration Over Code**: Externalize business parameters and timeouts
6. **Test Module Boundaries**: Verify modules can be tested in isolation
7. **Monitor Workflows**: Add observability to understand business process health

---

 **Next Steps**:
- [Service Implementation Guide](./service.md) - Module development patterns
- [Goa API Development](./goa.md) - Handler and API design patterns  
- [Database Transactions](./database-transactions.md) - Tenant isolation and SQLC patterns
- [Best Practices](./01-best-practices.md) - Development guidelines and standards