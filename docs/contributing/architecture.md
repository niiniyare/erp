# Architecture Overview

Our ERP system follows a **Hexagonal (Ports & Adapters) Architecture**. This is a specific implementation of Clean Architecture that emphasizes the isolation of the application core from external technologies and frameworks.

## 🏗️ System Architecture

The architecture is designed to protect the core business logic from outside concerns. External actors (like users or other systems) interact with the core through specific "Ports," and "Adapters" provide the concrete implementation for these interactions.

```
                                 ┌──────────────────┐
                                 │   Web Browser    │
                                 │ (External Actor) │
                                 └──────────────────┘
                                         │
                                         ▼
                          ┌─────────────────────────────┐
                          │   Transport/API Layer       │
                          │ (internal/api/handlers)     │
                          └─────────────────────────────┘
                                         │ (Primary Port)
                                         ▼
  ┌──────────────────────────────────────────────────────────────────────────┐
  │                            Application Core                              │
  │                          (internal/core/{domain})                          │
  │                                                                          │
  │   ┌────────────────┐   ┌──────────────────┐   ┌───────────────────────┐  │
  │   │    Services    │──▶│ Domain Models &  │◀──│ Repository Interfaces │  │
  │   │ (Business Logic) │   │    Interfaces    │   │    (Secondary Port)   │  │
  │   └────────────────┘   └──────────────────┘   └───────────────────────┘  │
  │                                                                          │
  └────────────────────────────────────┬─────────────────────────────────────┘
                                       │ (Secondary Adapters)
           ┌───────────────────────────┴───────────────────────────┐
           │                                                       │
┌──────────────────────────┐                          ┌──────────────────────────┐
│  Persistence Adapter     │                          │  Platform Infrastructure │
│ (internal/core/{domain}/ │                          │   (internal/platform)    │
│      repository)         │                          │                          │
└──────────────────────────┘                          └──────────────────────────┘
           │                                                       │
           ▼                                                       ▼
┌──────────────────────────┐                          ┌──────────────────────────┐
│   Database (Postgres)    │                          │      Cache (Redis)       │
│       (sqlc)             │                          │      Middleware, etc.    │
└──────────────────────────┘                          └──────────────────────────┘
```

## 📁 Directory Structure

The directory structure strictly follows the architectural pattern, separating concerns into distinct modules.

```
internal/
├── api/
│   ├── design/           # API Design (Goa DSL) - The Source of Truth
│   ├── middleware/       # Tenant context, auth, logging middleware
│   └── handlers/         # HTTP Handlers (Adapter for API input)
├── core/
│   ├── {domain}/         # Self-contained domain module (e.g., tenant, abac)
│   │   ├── models/       # Domain models for the module
│   │   ├── services/     # Business logic implementation (Application Core)
│   │   ├── repository/   # Data persistence IMPLEMENTATION (Adapter)
│   │   └── *.go          # Interfaces (Ports) for services/repositories
│   └── ...
├── platform/             # Cross-cutting infrastructure concerns (Adapters)
│   ├── cache/            # Caching implementation (e.g., Redis)
│   └── config/           # Application configuration
└── shared/               # Utilities used by all other layers
    ├── logger/
    ├── errors/
    └── context.go
```

## 🎯 Layer Responsibilities

### 1. API Layer (`internal/api/`)

**Purpose**: To act as an adapter for external API calls, translating them into calls to the application core.

#### Components:
- **`design/`**: The declarative source of truth for the API using the **Goa** DSL.
- **`handlers/`**: The concrete implementation that handles HTTP requests, extracts data, and calls the appropriate core service.

#### Responsibilities:
- ✅ Define API contracts, endpoints, and data structures in `design/`.
- ✅ Handle HTTP request/response lifecycle.
- ✅ Perform initial input validation and sanitization.
- ✅ Call the appropriate service in the core layer, passing a `context`.
- ✅ Format responses (JSON, status codes) based on the results from the core.
- ❌ **No business logic.**
- ❌ **No direct database or repository access.**

### 2. Application Core (`internal/core/{domain}/`)

**Purpose**: To contain all business logic and domain knowledge, completely independent of any external framework or technology.

#### Components:
- **Service Interfaces (Ports)**: Defines the contract for the business logic (e.g., `TenantService` interface).
- **`services/`**: The implementation of the service interfaces, orchestrating business operations.
- **`models/`**: The rich domain entities, value objects, and data structures that represent business concepts.
- **Repository Interfaces (Ports)**: Defines the contracts for data persistence (e.g., `TenantRepository` interface). The service depends on this, not the implementation.

#### Responsibilities:
- ✅ Implement all business rules and validation.
- ✅ Manage the state and lifecycle of domain entities.
- ✅ Orchestrate complex operations and workflows.
- ✅ Define the requirements for persistence through repository interfaces.
- ❌ **No knowledge of HTTP, SQL, or specific databases.**
- ❌ **Depends only on interfaces (ports), never concrete implementations (adapters).**

### 3. Persistence Adapter (`internal/core/{domain}/repository/`)

**Purpose**: To implement the data persistence port defined in the application core. It adapts the domain's needs to the specific database technology.

#### Components:
- **Repository Implementation**: A struct that implements the repository interface from the core.
- **Model Converters**: Functions to transform data between the core domain models and the `sqlc`-generated database models.
- **Store Integration**: Uses the `sqlc`-generated `Store` to execute database queries.

#### Responsibilities:
- ✅ Implement the repository interfaces defined in the core.
- ✅ Convert domain models to database models (and vice-versa).
- ✅ Execute queries using the type-safe `sqlc` store.
- ✅ Handle database-specific errors and translate them into domain-agnostic errors.
- ✅ Manage database transactions.
- ❌ **No business logic.**

### 4. Platform Infrastructure (`internal/platform/`)

**Purpose**: To provide concrete implementations for all other cross-cutting technical concerns.

#### Components:
- **`database/`**: Connection management, pooling, and migration logic.
- **`middleware/`**: HTTP middleware adapters for auth, logging, and injecting the tenant context.
- **`cache/`**: A Redis client adapter that implements a generic caching interface.
- **`config/`**: Environment-based configuration loading.

#### Responsibilities:
- ✅ Manage the lifecycle of infrastructure components.
- ✅ Provide concrete tools for the rest of the application to use.
- ✅ Integrate with external third-party services.
- ❌ **No business logic or domain-specific knowledge.**

## 🔄 Dependency Flow

The dependency rule is strict: **all dependencies must point inward, toward the application core.**

### Correct Dependency Direction
```
API Handlers (`adapters`) ──depends on──▶ Core Service Interfaces (`ports`)
                                                  │
                                                  ▼
Core Services (`core`)    ──depends on──▶ Core Repository Interfaces (`ports`)
                                                                      ▲
                                                                      │ (implements)
Persistence (`adapters`)  ◀───uses───── `sqlc` Generated Store & `platform`
```

### Key Rules:
1.  The `internal/core` modules must not import any other layer (`api`, `platform`).
2.  The `internal/api` and `internal/platform` layers depend on the interfaces defined in `internal/core`.
3.  The repository *implementations* in `internal/core/{domain}/repository` are adapters and are allowed to depend on `platform` and the `db/sqlc` store.

### Example Dependency Injection (Conceptual `main.go`):
```go
// main.go - Wire dependencies together
func main() {
    // 1. Platform & Infrastructure Adapters
    dbConn := database.Connect(cfg.Database)
    store := sqlc.NewStore(dbConn) // sqlc-generated DAL
    
    // 2. Persistence Adapter (implements the core port)
    // This NewRepository lives in internal/core/tenant/repository/
    tenantRepo := tenant_repository.New(store)
    
    // 3. Core Service (depends on the repository PORT/interface)
    // This NewService lives in internal/core/tenant/services/
    tenantService := tenant_service.New(tenantRepo)
    
    // 4. API Adapter (depends on the service PORT/interface)
    // This NewHandler lives in internal/api/handlers/
    tenantHandler := handlers.NewTenantHandler(tenantService)
    
    // 5. Setup router with the handler
    router.Setup(tenantHandler)
}
```

## 🏛️ Hexagonal Architecture Benefits

### 1. **Maximum Testability**
- The application core is completely isolated and can be tested without any database or API, leading to extremely fast and reliable unit tests.
- Adapters can be tested independently against their respective ports.

### 2. **Framework Independence**
- The core business logic is not tied to Goa, Gin, or any other web framework. The entire API layer could be swapped out with minimal impact on the core.

### 3. **Technology Independence**
- The persistence logic is abstracted away. The database could be swapped from Postgres to another SQL database by changing only the `sqlc` queries and the repository implementation, with zero changes to the business logic.

### 4. **Clear Boundaries & Maintainability**
- The strict separation of concerns makes the codebase easy to navigate and understand. It's always clear where a specific piece of logic should live.

## 🚨 Common Anti-Patterns to Avoid

### ❌ **Wrong Dependency Direction**
```go
// In internal/core/tenant/services/service.go
import "internal/api/handlers" // ❌ WRONG! The core cannot know about the API layer.
import "internal/platform/database" // ❌ WRONG! The core cannot know about platform specifics.
```

### ❌ **Business Logic in Adapters**
```go
// In internal/api/handlers/tenant.go
func (h *Handler) CreateTenant(c *gin.Context) {
    // ❌ Business validation should be in the core service, not the handler.
    if len(req.Name) < 3 {
        c.JSON(400, gin.H{"error": "Name too short"})
        return
    }
}
```

### ❌ **Leaking Database Models**
```go
// In internal/core/tenant/services/service.go
import "project/db/sqlc" // ❌ WRONG! The service must not know about sqlc models.
func (s *Service) CreateTenant(req CreateRequest) (*sqlc.Tenant, error) {
    // ❌ The service must always use and return its own domain models.
}
```

## ✅ Best Practices

1.  **Define Ports First**: Always start by defining the interfaces in the core that represent the application's needs.
2.  **Implement Adapters Second**: Write the concrete implementations for those interfaces in the outer layers (`repository`, `handlers`).
3.  **Convert Models at the Boundary**: The adapter's primary job is to convert data between the external world's format and the core domain's format.
4.  **Dependency Injection**: Wire everything together at the application's entry point (`main.go`).

---

📚 **Next Steps**: 
- [Data Flow Patterns](./data-flow.md) - See how requests flow through layers
- [Code Examples](./code-examples.md) - Practical implementation examples
- [Best Practices](./01-best-practices.md) - Development guidelines
