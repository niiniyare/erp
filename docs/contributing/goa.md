# ERP GOA API Implementation Guide
## *Design-First API Development with Modular Type Architecture*

*A comprehensive guide for implementing GOA-based APIs in ERP using design-first development, feature-specific type organization, and reusable domain models*

> **📚 Essential Reading:** This guide focuses on API layer implementation. For complementary patterns, also review:
> - `docs/contributing/service.md` - Service layer and business logic implementation
> - `docs/contributing/database-transactions.md` - Database integration with SQLC  
> - `docs/contributing/general-testing.md` - API testing strategies and patterns
> - `docs/contributing/error-handling.md` - Error handling and response patterns

## **🏗️ Modern API Architecture Pattern**

Our ERP system follows a sophisticated multi-layered architecture that promotes modularity, reusability, and maintainability:

```
internal/api/
├── design/                               # 🎨 GOA Design Layer (Design-First)
│   ├── design.go                         # Main API definition
│   ├── services/                         # Service definitions by domain
│   │   └── {domain}/                     # Domain-specific service files
│   │       ├── {main_service}.go         # Main service (e.g., transactions.go)
│   │       ├── {feature}_service.go      # Feature services (e.g., accounts.go) 
│   │       ├── types_{feature}.go        # Feature-specific payload/result types
│   │       ├── types_{workflow}.go       # Workflow-related types
│   │       ├── types_{reporting}.go      # Reporting types
│   │       └── types.go                  # Legacy/backward compatibility types
│   └── types/                            # 🔄 Reusable Domain Types (Cross-System)
│       ├── common.go                     # Universal patterns, pagination, audit
│       ├── {domain}.go                   # Domain models & enums (e.g., finance.go)
│       └── abac.go                       # ABAC-specific types
│
├── gen/                                  # 🤖 Generated GOA Code
│   ├── {service}/                        # Generated service interfaces
│   └── http/{service}/                   # Generated HTTP transport
│
└── handlers/                             # 🛠️ Handler Implementation Layer
    ├── {domain}.go                       # Domain entry points
    ├── {domain}/                         # Domain handler directories
    │   ├── service_handler.go            # Main service implementation
    │   ├── {feature}_handler.go          # Feature-specific handlers
    │   ├── types.go                      # Type conversions
    │   ├── validation.go                 # Business validation
    │   └── errors.go                     # Error handling
    └── common/                           # Shared utilities
```

## **📁 Type Organization Architecture**

### **1. Reusable Domain Types (`/types/`)**
Common types shared across multiple systems to prevent circular dependencies:

```go
// internal/api/design/types/finance.go - Reusable Financial Domain Types
package types

// Financial Domain Enums - Reusable across all financial services
var AccountRootType = func() {
    Enum("ASSET", "LIABILITY", "EQUITY", "REVENUE", "EXPENSE")
}

var TransactionStatus = func() {
    Enum("DRAFT", "PENDING_APPROVAL", "APPROVED", "POSTED", "REVERSED")
}

// Core Financial Domain Models - For cross-system usage
var FinancialAmount = Type("FinancialAmount", func() {
    Description("Financial amount with currency information")
    Attribute("amount", String, "Amount as decimal string", func() {
        Pattern(DecimalPattern)
        Example("1234.56")
    })
    Attribute("currency", String, "Currency code (ISO 4217)", func() {
        Pattern(CurrencyCodePattern)
        Example("USD")
    })
    Required("amount", "currency")
})

var AccountReference = Type("AccountReference", func() {
    Description("Lightweight reference to an account")
    Attribute("id", String, "Account ID", func() {
        Format(FormatUUID)
    })
    Attribute("code", String, "Account code", func() {
        Example("1100")
    })
    Attribute("name", String, "Account name", func() {
        Example("Cash - Operating Account")
    })
    Required("id", "code", "name")
})
```

### **2. Feature-Specific Service Types**
Each feature maintains its own payload and result types:

```go
// internal/api/design/services/finance/types_transactions.go
package finance

import (
    . "awo.so/internal/api/design/types"
    . "goa.design/goa/v3/dsl"
)

var CreateTransactionPayload = Type("CreateTransactionPayload", func() {
    Description("Payload for creating a new transaction")
    
    Attribute("transaction_type", String, "Transaction type", func() {
        Enum("GENERAL_JOURNAL", "ACCOUNTS_PAYABLE", "ACCOUNTS_RECEIVABLE")
    })
    Attribute("amount", FinancialAmount, "Transaction amount") // Reusing domain type
    Attribute("entries", ArrayOf(TransactionEntryPayload), "Journal entries")
    
    Required("transaction_type", "amount", "entries")
})

// Inline payloads from service definitions
var GetTransactionByIdPayload = Type("GetTransactionByIdPayload", func() {
    Description("Payload for getting transaction by ID")
    Attribute("id", String, "Transaction ID", func() {
        Format(FormatUUID)
    })
    Required("id")
})
```

### **3. Service Definition Pattern**
Clean, minimal service definitions that reference organized types:

```go
// internal/api/design/services/finance/transactions.go
package finance

import (
    . "goa.design/goa/v3/dsl"
)

// Service describes the finance management service
var _ = Service("finance", func() {
    Description("Financial management service for double-entry bookkeeping and accounting")

    HTTP(func() {
        Path("/api/v1/finance")
    })

    // Clean method definitions referencing types
    Method("createTransaction", func() {
        Description("Create a new financial transaction")
        Payload(CreateTransactionPayload)        // From types_transactions.go
        Result(TransactionResult)                // From types_transactions.go
        Error("bad_request")
        Error("unauthorized")
        Error("unprocessable_entity")
        HTTP(func() {
            POST("/transactions")
            Response(StatusCreated)
        })
    })

    Method("getTransaction", func() {
        Description("Get transaction by ID with entries")
        Payload(GetTransactionByIdPayload)       // Moved from inline definition
        Result(TransactionWithEntriesResult)
        Error("not_found")
        Error("unauthorized")
        HTTP(func() {
            GET("/transactions/{id}")
            Param("id")
        })
    })
})
```

## **🎯 Type Organization Best Practices**

### **1. Domain Types Placement Rules**

- **`/types/finance.go`**: Core financial models, enums, validation types that other systems might need
- **`/services/finance/types_transactions.go`**: Transaction-specific payloads and results
- **`/services/finance/types_accounts.go`**: Account management payloads and results  
- **`/services/finance/types_workflow.go`**: Workflow and approval types
- **`/services/finance/types_reporting.go`**: Financial reporting types
- **`/services/finance/types.go`**: Legacy compatibility types

### **2. Import Patterns**

```go
// In reusable types files (types/finance.go)
package types
import (
    . "goa.design/goa/v3/dsl"
)

// In feature-specific types files (services/finance/types_transactions.go)
package finance
import (
    . "awo.so/internal/api/design/types"  // Access to reusable types
    . "goa.design/goa/v3/dsl"
)

// In service definition files (services/finance/transactions.go)
package finance
import (
    . "goa.design/goa/v3/dsl"  // Only GOA DSL needed
)
// Note: Types are automatically accessible due to same package
```

### **3. Type Naming Conventions**

| Type Category | Naming Pattern | Example | Location |
|---------------|----------------|---------|-----------|
| **Payloads** | `{Action}{Entity}Payload` | `CreateTransactionPayload` | `types_{feature}.go` |
| **Results** | `{Entity}Result` | `TransactionResult` | `types_{feature}.go` |  
| **Lists** | `{Entity}ListResult` | `AccountListResult` | `types_{feature}.go` |
| **Domain Models** | `{Entity}` | `FinancialAmount` | `types/{domain}.go` |
| **Enums** | `{Entity}Type` or `{Entity}Status` | `AccountType`, `TransactionStatus` | `types/{domain}.go` |
| **Inline Service Payloads** | `Get{Entity}ByIdPayload` | `GetTransactionByIdPayload` | `types_{feature}.go` |

## **🔧 Service Design Patterns**

### **Single Service Architecture**
All related functionality consolidated into one service for better organization:

```go
// ✅ GOOD: Single finance service with multiple feature areas
var _ = Service("finance", func() {
    Description("Financial management service")
    
    // Legacy account management (backward compatibility)
    Method("createAccount", func() { /* ... */ })
    Method("getAccount", func() { /* ... */ })
    
    // Modern unified account/group management  
    Method("createAccountNode", func() { /* ... */ })
    Method("listAccountNodes", func() { /* ... */ })
    
    // Transaction management
    Method("createTransaction", func() { /* ... */ })
    Method("postTransaction", func() { /* ... */ })
    
    // Workflow management
    Method("getTransactionStatus", func() { /* ... */ })
    Method("submitApprovalDecision", func() { /* ... */ })
    
    // Financial reporting
    Method("getTrialBalance", func() { /* ... */ })
})
```

### **Method Organization Patterns**

```go
// Group related methods with clear comments
var _ = Service("finance", func() {
    
    // Legacy Account Management Methods (for backward compatibility)
    Method("createAccount", func() { /* ... */ })
    Method("getAccount", func() { /* ... */ })
    Method("deleteAccount", func() { /* ... */ })
    
    // Modern Unified Account/Group Management Methods
    Method("createAccountNode", func() { /* ... */ })
    Method("getAccountNode", func() { /* ... */ })
    Method("listAccountNodes", func() { /* ... */ })
    
    // Transaction Management Methods
    Method("createTransaction", func() { /* ... */ })
    Method("getTransaction", func() { /* ... */ })
    Method("postTransaction", func() { /* ... */ })
    
    // Transaction Workflow Management Methods
    Method("getTransactionStatus", func() { /* ... */ })
    Method("submitApprovalDecision", func() { /* ... */ })
    
    // Advanced Financial Reporting Methods
    Method("getTrialBalance", func() { /* ... */ })
})
```

## **📐 Advanced Type Composition**

### **Reusable Pattern Usage**

```go
// Using common patterns from types/common.go
var ListTransactionsPayload = Type("ListTransactionsPayload", func() {
    Description("Payload for listing transactions")
    
    // Business-specific filters
    Attribute("status", String, "Filter by status", TransactionStatus)
    Attribute("account_id", String, "Filter by account", func() {
        Format(FormatUUID)
    })
    
    // Reuse common patterns
    Attribute("date_range", TimeRange, "Date range filter")         // From common.go
    Attribute("pagination", Pagination, "Pagination parameters")     // From common.go
})

var TransactionListResult = Type("TransactionListResult", func() {
    Description("List of transactions with pagination")
    
    Attribute("transactions", ArrayOf(TransactionResult), "Transaction list")
    Attribute("pagination", PaginationMeta, "Pagination metadata")   // From common.go
    
    Required("transactions", "pagination")
})
```

### **Domain Type Composition**

```go
// Compose complex types from domain building blocks
var JournalEntry = Type("JournalEntry", func() {
    Description("Individual journal entry")
    
    Attribute("account", AccountReference, "Account being debited/credited")  // Domain type
    Attribute("debit_amount", FinancialAmount, "Debit amount")               // Domain type
    Attribute("credit_amount", FinancialAmount, "Credit amount")             // Domain type
    Attribute("description", String, "Entry description")
    
    Required("account", "debit_amount", "credit_amount", "description")
})
```

## **🧪 Testing Patterns**

### **Type Validation Testing**

```go
func TestCreateTransactionPayload_Validation(t *testing.T) {
    tests := []struct {
        name    string
        payload *goaFinance.CreateTransactionPayload
        wantErr bool
    }{
        {
            name: "valid_payload",
            payload: &goaFinance.CreateTransactionPayload{
                TransactionType: "GENERAL_JOURNAL",
                Amount: &goaFinance.FinancialAmount{
                    Amount:   "1234.56",
                    Currency: "USD",
                },
                Entries: []*goaFinance.TransactionEntryPayload{
                    {Description: "Test entry 1"},
                    {Description: "Test entry 2"},
                },
            },
            wantErr: false,
        },
        {
            name: "missing_required_fields",
            payload: &goaFinance.CreateTransactionPayload{
                TransactionType: "GENERAL_JOURNAL",
                // Missing Amount and Entries
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Validation testing logic
        })
    }
}
```

## **🔄 Development Workflow**

### **Adding New Features**

1. **Design Phase**
   ```bash
   # 1. Add domain types (if needed) to types/{domain}.go
   # 2. Create feature-specific types in services/{domain}/types_{feature}.go
   # 3. Update service definition to use new types
   # 4. Generate GOA code
   make goa
   ```

2. **Type Organization Steps**
   ```go
   // Step 1: Define reusable domain types (types/finance.go)
   var NewDomainType = Type("NewDomainType", func() {
       // Common fields that other systems might use
   })
   
   // Step 2: Define feature payloads (services/finance/types_newfeature.go)
   var CreateNewFeaturePayload = Type("CreateNewFeaturePayload", func() {
       // Feature-specific request fields
       Attribute("domain_field", NewDomainType, "Reused domain type")
   })
   
   // Step 3: Reference in service (services/finance/main_service.go)
   Method("createNewFeature", func() {
       Payload(CreateNewFeaturePayload)  // Clean reference
       Result(NewFeatureResult)
   })
   ```

3. **Type Migration Process**
   ```go
   // Move inline payloads to appropriate types files:
   
   // ❌ Before (inline definition)
   Method("getItem", func() {
       Payload(func() {
           Attribute("id", String, "Item ID", func() {
               Format(FormatUUID)
           })
           Required("id")
       })
   })
   
   // ✅ After (extracted to types file)
   Method("getItem", func() {
       Payload(GetItemByIdPayload)  // Defined in types_{feature}.go
   })
   ```

### **Code Generation Commands**

```bash
# Generate all GOA code
make goa

# Clean and regenerate
make clean && make goa

# Verify no generation errors
make goa 2>&1 | grep -i error
```

## **✅ Architecture Benefits**

1. **🎯 Clear Separation of Concerns**: Types organized by feature and reusability
2. **🔄 Reusability**: Domain types prevent duplication across services  
3. **🚫 No Circular Dependencies**: Clean import hierarchy with `/types/` at the base
4. **🧹 Minimal Service Definitions**: Services focus on method definitions, not type declarations
5. **🔍 Easy Navigation**: Developers can quickly find feature-specific types
6. **📈 Scalability**: New features can easily reuse existing domain types
7. **🤝 Cross-System Integration**: Other services can import domain types from `/types/`
8. **🧪 Testability**: Clear type boundaries make testing straightforward

## **⚠️ Common Pitfalls to Avoid**

1. **❌ Don't** put business-specific payloads in `/types/common.go`
2. **❌ Don't** create circular imports between service types and domain types  
3. **❌ Don't** use inline type definitions in service files
4. **❌ Don't** mix feature-specific types in wrong files (e.g., transaction types in account types file)
5. **❌ Don't** forget to use domain enums for consistent validation
6. **✅ Do** keep service definitions minimal and clean
7. **✅ Do** reuse common patterns (pagination, audit fields, time ranges)
8. **✅ Do** follow consistent naming conventions
9. **✅ Do** organize types by feature and logical grouping
10. **✅ Do** validate GOA generation after every change

This architecture ensures maintainable, scalable, and well-organized API definitions that support complex business domains while remaining developer-friendly and preventing common architectural issues.