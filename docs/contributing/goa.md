# Awo ERP Goa API Implementation Guide
## *Design-First API Development with Modular Handler Architecture*

*A comprehensive guide for implementing Goa-based APIs in Awo ERP using design-first development and modular handler organization*

> **📚 Essential Reading:** This guide focuses on API layer implementation. For complementary patterns, also review:
> - `docs/contributing/service.md` - Service layer and business logic implementation
> - `docs/contributing/database-transactions.md` - Database integration with SQLC  
> - `docs/contributing/general-testing.md` - API testing strategies and patterns
> - `docs/contributing/error-handling.md` - Error handling and response patterns

## **🏗️ Handler Architecture Pattern**

Each module in Awo ERP follows a consistent handler organization pattern that promotes modularity and maintainability:

```
internal/api/handlers/
├── {module}.go                    # 🔥 MAIN MODULE ENTRY POINT - Called by main
├── {module}/                      # Module-specific handler directory
│   ├── service_handler.go         # Core service handler (implements Goa interface)
│   ├── {feature}_handler.go       # Feature-specific handlers
│   ├── types.go                   # Request/response type adapters
│   ├── validation.go              # Input validation helpers
│   └── errors.go                  # Module-specific error handling
├── common/                        # Shared handler utilities
│   ├── middleware.go              # Common middleware (auth, ABAC, tracing)
│   ├── responses.go               # Standard response helpers
│   └── validation.go              # Shared validation utilities
└── router.go                      # Main router configuration
```

**Example: Finance Module Structure**
```
internal/api/handlers/
├── finance.go                     # Entry point - exports NewFinanceHandler()
├── finance/
│   ├── service_handler.go         # Main finance service handler
│   ├── account_handler.go         # Account management endpoints
│   ├── transaction_handler.go     # Transaction processing endpoints
│   ├── report_handler.go          # Financial reporting endpoints
│   ├── types.go                   # Finance-specific type conversions
│   └── validation.go              # Finance business rule validation
```

## **📁 Complete Project Structure**

```
internal/api/
├── design/                         # Goa Design Specifications
│   ├── design.go                   # Main API definition
│   ├── services/
│   │   ├── {module}.go             # Module service design
│   │   └── types.go                # Module types and validation
│   └── types/
│       ├── common.go               # Shared types across modules
│       └── errors.go               # Standard error definitions
│
├── gen/                            # Generated Goa code
│   ├── {module}/                   # Generated service interfaces
│   └── http/{module}/              # Generated HTTP transport
│
├── handlers/                       # Handler Implementation Layer
│   ├── {module}.go                 # Module entry points
│   ├── {module}/                   # Module handler directories
│   │   ├── service_handler.go      # Core service implementation
│   │   ├── {feature}_handler.go    # Feature handlers
│   │   ├── types.go                # Type conversions
│   │   ├── validation.go           # Input validation
│   │   └── errors.go               # Error handling
│   ├── common/                     # Shared utilities
│   │   ├── middleware.go           # Authentication, ABAC, logging
│   │   ├── responses.go            # Response helpers
│   │   └── validation.go           # Common validation
│   └── router.go                   # Router setup
```

## **🎨 1. Design-First Development with Goa**

### **Main API Design Entry Point**

```go
// internal/api/design/design.go
package design

import (
    . "goa.design/goa/v3/dsl"
)

// API describes the global properties of Awo ERP API server
var _ = API("awo-erp", func() {
    Title("Awo Enterprise ERP System")
    Description("Multi-tenant ERP system with ABAC authorization and Temporal workflows")
    Version("1.0.0")
    
    Contact(func() {
        Name("Awo Development Team")
        Email("dev@awo.com")
        URL("https://awo.com")
    })
    
    License(func() {
        Name("Proprietary")
        URL("https://awo.com/license")
    })
    
    Server("awo-erp", func() {
        Host("localhost", func() {
            URI("http://localhost:8080")
            URI("https://localhost:8443")
        })
    })
    
    // Global security schemes
    JWTSecurity("jwt", func() {
        Description("JWT token authentication")
        Header("Authorization")
        Prefix("Bearer ")
    })
    
    // Global error responses
    Error("internal_error", String, "Internal server error")
    Error("bad_request", String, "Bad request")
    Error("unauthorized", String, "Unauthorized access")
    Error("forbidden", String, "Forbidden access")
    Error("not_found", String, "Resource not found")
    Error("conflict", String, "Resource conflict")
    Error("validation_error", String, "Validation failed")
})
```

### **Module Service Design Pattern**

Each module defines its own service design following consistent patterns:

```go
// internal/api/design/services/finance.go
package services

import (
    . "goa.design/goa/v3/dsl"
)

// Finance service design
var _ = Service("finance", func() {
    Description("Financial management operations including accounts, transactions, and reporting")
    
    // Common security and error handling
    Security(JWTAuth)
    Error("unauthorized", ErrorResult, "Credentials are invalid")
    Error("forbidden", ErrorResult, "Insufficient permissions")
    Error("not_found", ErrorResult, "Resource not found")
    
    // Account management endpoints
    Method("create_account", func() {
        Description("Create a new financial account")
        Payload(CreateAccountRequest)
        Result(AccountResponse)
        HTTP(func() {
            POST("/accounts")
            Response(StatusCreated)
        })
    })
    
    Method("list_accounts", func() {
        Description("List financial accounts with filtering")
        Payload(ListAccountsRequest)
        Result(CollectionOf(AccountResponse))
        HTTP(func() {
            GET("/accounts")
            Param("type")        // Query parameter for account type
            Param("status")      // Query parameter for account status
            Response(StatusOK)
        })
    })
    
    // Transaction management endpoints
    Method("create_transaction", func() {
        Description("Create a new financial transaction")
        Payload(CreateTransactionRequest)
        Result(TransactionResponse)
        HTTP(func() {
            POST("/transactions")
            Response(StatusCreated)
        })
    })
})
```

## **🔧 2. Handler Implementation Pattern**

### **Module Entry Point Pattern**

Each module must have a main entry point that exports the handler constructor:

```go
// internal/api/handlers/finance.go
package handlers

import (
    "github.com/niiniyare/erp/internal/api/handlers/finance"
    goaFinance "github.com/niiniyare/erp/internal/api/gen/finance"
    financeService "github.com/niiniyare/erp/internal/core/finance"
    "github.com/niiniyare/erp/internal/shared/metrics"
    "github.com/niiniyare/erp/internal/shared/tracing"
)

// NewFinanceHandler creates a new finance handler that implements the Goa service interface
// This is the ONLY function that main.go should call for this module
func NewFinanceHandler(
    financeService financeService.Service,
    tracing tracing.TracingService,
    metrics metrics.MetricsProvider,
) goaFinance.Service {
    return finance.NewFinanceHandler(financeService, tracing, metrics)
}
```

### **Core Service Handler Implementation**

```go
// internal/api/handlers/finance/service_handler.go
package finance

import (
    "context"
    "errors"

    goaFinance "github.com/niiniyare/erp/internal/api/gen/finance"
    financeService "github.com/niiniyare/erp/internal/core/finance"
    sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
    "github.com/niiniyare/erp/internal/shared/metrics"
    "github.com/niiniyare/erp/internal/shared/tracing"
)

// FinanceHandler implements the Goa finance service interface
type FinanceHandler struct {
    financeService financeService.Service
    tracing        tracing.TracingService
    metrics        metrics.MetricsProvider
}

// NewFinanceHandler creates a new finance handler
func NewFinanceHandler(
    financeService financeService.Service,
    tracing tracing.TracingService,
    metrics metrics.MetricsProvider,
) goaFinance.Service {
    return &FinanceHandler{
        financeService: financeService,
        tracing:        tracing,
        metrics:        metrics,
    }
}

// CreateAccount implements the create_account endpoint
func (h *FinanceHandler) CreateAccount(ctx context.Context, p *goaFinance.CreateAccountPayload) (*goaFinance.Account, error) {
    // Start tracing span
    ctx, span := h.tracing.Start(ctx, "finance.CreateAccount")
    defer span.End()
    
    // Record metrics
    h.metrics.Counter("api.finance.create_account.requests").Add(1)
    
    // Convert Goa payload to domain request
    request := h.payloadToCreateAccountRequest(p)
    
    // Call business service
    account, err := h.financeService.Account().CreateAccount(ctx, request)
    if err != nil {
        h.metrics.Counter("api.finance.create_account.errors").Add(1)
        return nil, h.handleError(err)
    }
    
    // Convert domain response to Goa response
    response := h.accountToGoaResponse(account)
    
    h.metrics.Counter("api.finance.create_account.success").Add(1)
    return response, nil
}

// handleError converts domain errors to appropriate Goa errors
func (h *FinanceHandler) handleError(err error) error {
    var businessErr *sharedErrors.BusinessError
    if errors.As(err, &businessErr) {
        switch businessErr.HTTPStatus {
        case 400:
            return goaFinance.MakeBadRequest(err)
        case 404:
            return goaFinance.MakeNotFound(err)
        case 409:
            return goaFinance.MakeConflict(err)
        case 422:
            return goaFinance.MakeUnprocessableEntity(err)
        default:
            return goaFinance.MakeBadRequest(err)
        }
    }
    return goaFinance.MakeInternalError(err)
}
```

### **Type Conversion Helpers**

```go
// internal/api/handlers/finance/types.go
package finance

import (
    goaFinance "github.com/niiniyare/erp/internal/api/gen/finance"
    "github.com/niiniyare/erp/internal/core/finance/domain"
)

// Convert Goa payload to domain request
func (h *FinanceHandler) payloadToCreateAccountRequest(p *goaFinance.CreateAccountPayload) *domain.CreateAccountRequest {
    return &domain.CreateAccountRequest{
        Code:        p.Code,
        Name:        p.Name,
        AccountType: domain.AccountType(p.Type),
        Description: p.Description,
        ParentID:    p.ParentID,
    }
}

// Convert domain entity to Goa response
func (h *FinanceHandler) accountToGoaResponse(account *domain.Account) *goaFinance.Account {
    return &goaFinance.Account{
        ID:          account.ID.String(),
        Code:        account.Code,
        Name:        account.Name,
        Type:        string(account.AccountType),
        Description: account.Description,
        Balance:     account.Balance.String(),
        Status:      string(account.Status),
        CreatedAt:   account.CreatedAt.Format(time.RFC3339),
        UpdatedAt:   account.UpdatedAt.Format(time.RFC3339),
    }
}
```

## **🔗 3. Main Application Integration**

### **Handler Registration in Main**

```go
// cmd/server/goa.go (modification)
func InitializeGOAServer(services *Services, metricsService *metrics.MetricsService, tracingService tracing.TracingService) (*GOAServer, error) {
    // Initialize GOA services using module entry points
    var (
        authSvc      auth.Service
        tenantSvc    goaTenant.Service
        financeSvc   goaFinance.Service  // New finance service
        // ... other services
    )

    // Initialize services using entry point functions
    authSvc = handlers.NewAuthHandler(services.IdentityService, tracingService, metricsService)
    tenantSvc = handlers.NewTenantGoaHandler(services.TenantService, tracingService, metricsService)
    financeSvc = handlers.NewFinanceHandler(services.FinanceService, tracingService, metricsService) // New
    
    // Create endpoints
    authEndpoints := auth.NewEndpoints(authSvc)
    tenantEndpoints := goaTenant.NewEndpoints(tenantSvc)
    financeEndpoints := goaFinance.NewEndpoints(financeSvc) // New
    
    // Create and mount servers
    mux := goahttp.NewMuxer()
    eh := createProductionErrorHandler()
    
    authServer := authsvr.New(authEndpoints, mux, dec, enc, eh, nil)
    tenantServer := tenantsvr.New(tenantEndpoints, mux, dec, enc, eh, nil)
    financeServer := financesvr.New(financeEndpoints, mux, dec, enc, eh, nil) // New
    
    // Mount servers
    authsvr.Mount(mux, authServer)
    tenantsvr.Mount(mux, tenantServer)
    financesvr.Mount(mux, financeServer) // New
    
    return &GOAServer{Handler: mux, Mux: mux}, nil
}
```

## **🧪 4. Testing Strategy**

### **Handler Unit Tests**

```go
// internal/api/handlers/finance/service_handler_test.go
package finance

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    goaFinance "github.com/niiniyare/erp/internal/api/gen/finance"
    "github.com/niiniyare/erp/internal/core/finance/domain"
    "github.com/niiniyare/erp/internal/core/finance/mocks"
)

func TestCreateAccount_Success(t *testing.T) {
    // Setup
    mockService := mocks.NewMockService(t)
    mockAccount := mocks.NewMockAccountService(t)
    mockService.On("Account").Return(mockAccount)
    
    handler := NewFinanceHandler(mockService, nil, nil)
    
    // Mock successful account creation
    expectedAccount := &domain.Account{
        Code: "1000",
        Name: "Cash",
        AccountType: domain.AssetAccount,
    }
    
    mockAccount.On("CreateAccount", mock.Anything, mock.Anything).
        Return(expectedAccount, nil)
    
    // Execute
    payload := &goaFinance.CreateAccountPayload{
        Code: "1000",
        Name: "Cash",
        Type: "asset",
    }
    
    result, err := handler.CreateAccount(context.Background(), payload)
    
    // Verify
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "1000", result.Code)
    assert.Equal(t, "Cash", result.Name)
    mockAccount.AssertExpectations(t)
}
```

## **📝 5. Development Workflow**

### **Step-by-Step Implementation**

1. **Design Phase**
   ```bash
   # Create service design
   touch internal/api/design/services/your_module.go
   
   # Generate Goa code
   make goa
   ```

2. **Handler Structure Setup**
   ```bash
   # Create module directory
   mkdir -p internal/api/handlers/your_module
   
   # Create entry point
   touch internal/api/handlers/your_module.go
   
   # Create core handler
   touch internal/api/handlers/your_module/service_handler.go
   ```

3. **Implementation**
   ```go
   // Follow the patterns shown above for:
   // - Entry point function
   // - Core service handler
   // - Type conversions
   // - Error handling
   ```

4. **Integration**
   ```go
   // Add to cmd/server/goa.go:
   // - Import module entry point
   // - Call constructor function
   // - Create endpoints
   // - Mount server
   ```

5. **Testing**
   ```bash
   # Run handler tests
   make test-unit
   
   # Test API integration
   make test-integration
   ```

## **✅ Best Practices**

1. **Module Independence**: Each module should be self-contained with its own entry point
2. **Service Interface Compliance**: All handlers must implement the generated Goa service interface
3. **Context Propagation**: Always pass context through all layers for tracing and cancellation
4. **Error Handling**: Convert domain errors to appropriate HTTP status codes
5. **Type Safety**: Use type conversion helpers to maintain separation between API and domain layers
6. **Testing**: Write comprehensive unit tests for all handlers
7. **Observability**: Include tracing spans and metrics in all handlers
8. **Validation**: Validate input at both Goa design and handler levels

## **🔄 Code Generation Commands**

```bash
# Generate all Goa code from design
make goa

# Generate specific modules (if configured)
goa gen github.com/niiniyare/erp/internal/api/design

# Clean and regenerate
make clean && make goa
```

This modular approach ensures that:
- Each module has a clear, single entry point
- Business logic stays in the service layer
- API concerns are properly separated
- Testing is straightforward and modular
- Main application only imports entry points, not internal handler details