# AWO ERP Financial Module

> **Enterprise-grade double-entry bookkeeping system with military-grade security and real-time reporting capabilities.**

## Overview

The Financial Module is the core accounting engine of the AWO ERP platform, providing production-ready financial management with sophisticated transaction processing, multi-currency support, and comprehensive compliance frameworks.

### Key Features

✅ **Double-Entry Bookkeeping** - Database-enforced accounting principles  
✅ **Multi-Tenant Security** - Automatic tenant isolation via Row Level Security  
✅ **Real-Time Processing** - Atomic transactions with optimistic locking  
✅ **Regulatory Compliance** - SOX, GAAP, IFRS audit trail requirements  
✅ **Multi-Currency Support** - Global operations with automatic conversion  
✅ **Advanced Reporting** - Financial statements with drill-down capabilities  

## Quick Start

### Prerequisites

- PostgreSQL 15+ with UUID extension
- Go 1.21+ for SQLC code generation
- Redis for caching (optional)

### Installation

```bash
# Clone and setup
git clone [repository]
cd erp/financial-module

# Install dependencies
go mod download

# Run database migrations
make migrate-up

# Start the service
make run-finance-service
```

### Basic Usage

```go
// Create an account
account, err := accountService.Create(ctx, &domain.CreateAccountRequest{
    AccountCode: "1000",
    AccountName: "Cash",
    RootType:    domain.RootTypeAsset,
})

// Create a transaction
transaction, err := transactionService.Create(ctx, &domain.CreateTransactionRequest{
    TransactionNumber: "TXN-001",
    Description:       "Initial cash deposit",
    Entries: []domain.CreateEntryRequest{
        {AccountID: cashAccountID, DebitAmount: decimal.NewFromFloat(1000.00)},
        {AccountID: equityAccountID, CreditAmount: decimal.NewFromFloat(1000.00)},
    },
})
```

## Architecture Overview

The Financial Module implements **Clean Architecture** with **Domain-Driven Design** patterns:

```mermaid
graph TB
    A[HTTP API Layer] --> B[Service Layer]
    B --> C[Domain Layer]
    B --> D[Repository Layer]
    D --> E[Database Layer]
    
    subgraph "Core Entities"
        F[Accounts]
        G[Transactions]
        H[Entries]
        I[Account Groups]
    end
    
    C --> F
    C --> G
    C --> H
    C --> I
```

### Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Database** | PostgreSQL 15+ | ACID transactions, RLS, advanced constraints |
| **Code Generation** | SQLC | Type-safe database operations |
| **Decimal Precision** | pgtype.Numeric | Financial calculations without floating-point errors |
| **Caching** | Redis | High-performance query optimization |
| **Observability** | OpenTelemetry | Distributed tracing and metrics |

## Core Concepts

### Chart of Accounts
Hierarchical account structure supporting unlimited depth with materialized path optimization for fast queries.

### Double-Entry Transactions
Every transaction maintains accounting equation: **Assets = Liabilities + Equity**
- Enforced at database level through CHECK constraints
- Automatic balance validation before posting

### Multi-Tenant Security
- **Row Level Security (RLS)**: Automatic tenant isolation
- **Comprehensive Audit Trails**: Full user tracking and change history
- **Soft Deletes**: Compliance-friendly data retention

### State Management
Transactions follow a strict state machine:
```
DRAFT → PENDING_APPROVAL → APPROVED → POSTED
```
Each transition validated through database constraints.

## Documentation Guide

| Document | Purpose | Audience |
|----------|---------|----------|
| **[Technical Architecture](./technical-architecture.md)** | Deep technical implementation, SQL schema, SQLC patterns | Developers, Architects |
| **[Business Domain Guide](./business-domain-guide.md)** | Financial concepts, business rules, domain model | Business Analysts, Product Teams |
| **[API & Integration](./api-integration.md)** | REST API reference and integration patterns | API Consumers, Integration Teams |
| **[Operations & Security](./operations-security.md)** | Security, compliance, deployment, testing | DevOps, Security Teams |
| **[Specialized Guides](./guides/)** | Currency management, reporting, workflows | Feature-specific guidance |

## Getting Help

### Development Support
- **Issues**: [GitHub Issues](./TASK.md) for bug reports
- **Architecture Questions**: See [Technical Architecture](./technical-architecture.md)
- **Business Logic**: See [Business Domain Guide](./business-domain-guide.md)

### Production Support
- **Security**: [Operations & Security Guide](./operations-security.md)
- **Performance**: Database optimization and query patterns
- **Compliance**: Audit requirements and regulatory guidance

## Contributing

1. Follow [Clean Architecture](./technical-architecture.md#clean-architecture) patterns
2. Maintain [financial integrity](./business-domain-guide.md#double-entry-principles) constraints
3. Add [comprehensive tests](./operations-security.md#testing-strategy)
4. Update relevant documentation

## Status & Roadmap

**Current Status**: 80% complete financial module with active development

| Component | Status | Coverage |
|-----------|--------|----------|
| **Core Accounting** | ✅ Complete | Chart of accounts, transactions, entries |
| **Multi-Currency** | ✅ Complete | Exchange rates, conversion, revaluation |
| **Reporting** | 🚧 In Progress | Trial balance, P&L, balance sheet |
| **Advanced Features** | 📅 Planned | Budgeting, forecasting, analytics |

See **[ROADMAP.md](./ROADMAP.md)** for detailed development timeline.

---

**Version**: 4.0 | **Status**: Production Ready | **Last Updated**: January 2025