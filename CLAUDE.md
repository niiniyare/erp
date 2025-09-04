# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a multi-project development environment containing several different applications and frameworks, primarily focused on ERP systems and business applications. The repository includes both Go-based backend services and frontend applications.

## Key Projects

### 1. Main ERP System (`/project/erp/`)
- **Language**: Go
- **Architecture**: Clean Architecture with Goa framework
- **Database**: PostgreSQL with SQLC
- **Features**: Multi-tenant ERP with ABAC (Attribute-Based Access Control)
- **Financial Module**: Production-ready double-entry bookkeeping system (~80% complete)

### 2. AWO Flight System (`/awo/`)
- **Language**: Go 
- **Architecture**: gRPC with Protocol Buffers
- **Database**: PostgreSQL with SQLC
- **Purpose**: Flight booking and management system

### 3. Frontend Project (`/project/`)
- **Framework**: Vue 3 + Vite
- **Dependencies**: Frappe UI, TailwindCSS
- **Purpose**: Frontend interface for ERP applications

## Common Commands

### ERP System (`/project/erp/`)
```bash
# Build and run
make run                    # Run the server
make ci                     # Run all essential checks (fmt, lint, test, proto, sqlc)

# Database operations
make createdb              # Create PostgreSQL database
make migrateup            # Run database migrations
make migratedown          # Rollback last migration
make sqlc                 # Generate SQLC code from SQL queries

# Code generation
make proto                # Generate gRPC and gateway files
make goa                  # Generate Goa framework code
make mock                 # Generate mocks for interfaces

# Testing
make test                 # Run all tests
make test-unit           # Run unit tests only
make test-integration    # Run integration tests
make test-coverage       # Run tests with coverage report

# Development
make fmt                 # Format Go code
make lint                # Lint Go code
make clean               # Clean generated files
```

### AWO Flight System (`/awo/`)
```bash
# Build and run
make server              # Run server on port 8080
make test                # Run all unit tests
make testdb              # Run database-related tests

# Database operations
make createdb            # Create flight database
make migrateup          # Run database migrations
make sqlc               # Generate SQLC code

# Code generation
make proto              # Generate protobuf files
make buf                # Use Buf to generate proto files
make mock               # Generate mocks

# Development
make test/html          # Generate HTML coverage report
```

### Frontend Project (`/project/`)
```bash
# Development
npm run dev             # Start Vite dev server
npm run build           # Build for production
npm run preview         # Preview production build
```

## Architecture Patterns

### ERP System Architecture
- **Clean Architecture**: Organized into layers (core, adapters, platform, shared)
- **Domain-Driven Design**: Business logic separated into domain services
- **Multi-tenant**: Row-level security (RLS) for tenant isolation
- **ABAC Authorization**: Attribute-based access control system
- **Event-driven**: Uses Temporal for workflow orchestration
- **API Design**: REST and gRPC APIs generated with Goa framework
- **Financial Module**: Enterprise-grade double-entry bookkeeping system with:
  - Complete transaction engine with state machine workflows
  - Multi-currency support with exchange rate management
  - Advanced validation framework (30+ business rules)
  -  audit trail and compliance features
  - Cost center/department/project dimensional analysis
  - Bank reconciliation and tax calculation capabilities

### Database Architecture
- **PostgreSQL**: Primary database with row-level security
- **SQLC**: Type-safe SQL query generation
- **Migrations**: Versioned database schema management
- **Multi-tenancy**: Tenant isolation at database level

### Key Directories Structure

#### ERP System (`/project/erp/`)
- `cmd/` - Application entry points (server, migrate, worker)
- `internal/api/` - API layer with Goa-generated handlers
- `internal/core/` - Business logic and domain services
  - `internal/core/finance/` - Financial module (80% complete):
    - `domain/` - Financial entities, value objects, and business rules
    - `service/` - Business logic layer with 10  services (7,256 lines)
    - `repository/` - Data access layer with full SQLC integration (4,730 lines)
- `internal/platform/` - Infrastructure concerns (database, cache, config)
- `internal/shared/` - Common utilities and types
- `db/migration/` - Database migration files (5 finance migrations complete)
- `db/queries/` - SQL query files for SQLC (3 finance query files with 40+ operations)
- `db/sqlc/` - Generated SQLC code

#### AWO System (`/awo/`)
- `cmd/` - Command-line applications
- `db/` - Database-related code and migrations
- `pkg/` - Package code including API definitions
- `util/` - Utility functions

## Development Workflow

1. **Database First**: Always run migrations before code changes
2. **Code Generation**: Regenerate code after schema or API changes using `make gen`
3. **Testing**: Run `make test-unit` during development, full `make ci` before commits
4. **Multi-tenant Context**: Always consider tenant isolation in database queries and API endpoints

## Important Notes

- The ERP system uses row-level security (RLS) for multi-tenant isolation
- ABAC system requires careful attribute management for access control
- Temporal workflows handle complex business processes
- Both systems use SQLC for type-safe database operations
- Protocol buffer definitions drive API contracts in AWO system
- Frontend connects to ERP APIs and requires CSRF token handling in development

## Financial Module Implementation Status

### ✅ **Completed Components (80% of module)**
- **Database Schema**: 5 migrations with  financial tables and constraints
- **Domain Layer**: 9 files with rich entities, value objects, and business validation
- **Service Layer**: 10 services totaling 7,256 lines including:
  - Transaction processing with state machine workflows
  - Double-entry validation engine (30+ business rules)
  - Multi-currency exchange rate engine
  - Transaction posting and reversal engines
  - Account management with hierarchical operations
- **Repository Layer**: Full SQLC integration with 4,730 lines covering:
  - Chart of accounts repository (640 lines, 14 methods)
  - Transaction repository (936 lines, 30+ methods)
  - Domain type mappers (689 lines, 15+ mapping functions)
  -  test suites (1,865 lines)

### 🚧 **Remaining Implementation (20%)**
- **API Layer**: Goa service designs and REST/gRPC handlers
- **Wire Integration**: Service dependency injection in main application
- **Advanced Modules**: AR/AP automation, customer/vendor management
- **Financial Reporting**: Standard financial statements (P&L, Balance Sheet)

### **Next Steps for Full Operation**
1. Create finance API designs (`internal/api/design/services/finance.go`)
2. Implement API handlers (`internal/api/handlers/finance_handler.go`)
3. Wire finance services into main application dependency injection
4. Set up test database for integration testing

## Testing Strategy

- **Unit Tests**: Focus on business logic in `internal/core/`
- **Integration Tests**: Test database interactions and external services
- **E2E Tests**: Full workflow testing with Docker compose
- **Coverage**: Maintain coverage reports for code quality tracking