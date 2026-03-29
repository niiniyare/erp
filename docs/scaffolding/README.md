# awoctl - ERP Scaffolding Tool

The `awoctl` (AWO Control) tool is a sophisticated code scaffolding system for the AWO ERP platform. It generates production-ready code following clean architecture patterns, ERP best practices, and organizational standards.

##  Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Commands](#commands)
- [Templates](#templates)
- [Integration](#integration)
- [Examples](#examples)
- [Best Practices](#best-practices)

##  Overview

`awoctl` generates boilerplate code for:

- **Complete Modules**: Domain, Service, Repository, and API layers
- **Individual Components**: Entities, services, repositories, handlers, DTOs, middleware, validators, mappers
- **Features**: Additions to existing modules
- **Test Scaffolds**: Unit, integration, and API tests

### Key Features

- **ERP-Aware**: Built-in understanding of financial modules, multi-tenancy, ABAC security
- **Production-Ready**: Generates code that compiles and follows Go best practices
- **Safety First**: Dry-run mode, validation, conflict detection
- **Template-Driven**: Comprehensive template system with business logic
- **Clean Architecture**: Enforces domain-driven design and layered architecture

##  Installation

### Method 1: Using Makefile (Recommended)

```bash
# Build and install the tool
make scaffold-install

# Verify installation
./awoctl --version
```

### Method 2: Manual Build

```bash
# Build from source
go build -o awoctl ./cmd/awoctl/

# Add to PATH (optional)
sudo mv awoctl /usr/local/bin/
```

### Method 3: Development Setup

```bash
# Full development environment setup (includes awoctl)
make dev-setup
```

## ⚡ Quick Start

### 1. Generate a Complete Module

```bash
# Generate inventory management module
make scaffold-module name=inventory

# Or use awoctl directly
./awoctl new module inventory --verbose
```

This creates:
```
internal/core/inventory/
├── domain/
│   ├── inventory.go         # Main entity
│   ├── types.go            # Status enums, types
│   ├── errors.go           # Domain errors
│   ├── validation.go       # Validation logic
│   ├── repository.go       # Repository interface
│   └── requests.go         # Request DTOs
├── inventory_service.go     # Business logic
├── settings_helper.go      # Configuration
└── repository/
    ├── interface.go        # Repository contract
    ├── inventory_repository.go  # Implementation
    ├── mappers/
    │   └── inventory_mappers.go  # Data mapping
    └── filters.go          # Query filters
```

### 2. Generate Individual Components

```bash
# Generate a payment entity in finance module
make scaffold-component module=finance name=payment type=entity

# Generate a service with tests
./awoctl new component finance payment_processor service --with-tests
```

### 3. Add Features to Existing Modules

```bash
# Add budgeting feature to finance module
make scaffold-feature path=finance/budgets

# Add inventory adjustments
./awoctl new feature inventory/adjustments --verbose
```

##  Commands

### Module Generation

```bash
awoctl new module <name> [flags]
```

**Flags:**
- `--dry-run`: Preview generation without creating files
- `--with-tests`: Generate test scaffolds
- `--verbose`: Enable detailed output

**Examples:**
```bash
# Basic module
awoctl new module payroll

# With tests and preview
awoctl new module hr --with-tests --dry-run --verbose
```

### Component Generation

```bash
awoctl new component <module> <component_name> <type> [flags]
```

**Component Types:**
- `entity`: Domain entity with validation
- `service`: Business logic service  
- `repository`: Data access layer
- `handler`: API handler
- `dto`: Data transfer object
- `middleware`: HTTP middleware
- `validator`: Input validator
- `mapper`: Data mapper

**Examples:**
```bash
# Generate payment entity
awoctl new component finance payment entity

# Generate service with tests
awoctl new component inventory stock_adjustment service --with-tests

# Generate API handler
awoctl new component hr employee_profile handler --verbose
```

### Feature Generation

```bash
awoctl new feature <module/feature> [flags]
```

**Examples:**
```bash
# Add reconciliation to finance
awoctl new feature finance/reconciliation

# Add reports to inventory
awoctl new feature inventory/reports --with-tests
```

##  Templates

### Template Categories

1. **Domain Layer Templates**
   - `entity.go.tmpl`: Core domain entities
   - `types.go.tmpl`: Enums, value objects
   - `validation.go.tmpl`: Business rules
   - `repository.go.tmpl`: Repository interfaces

2. **Service Layer Templates**
   - `service.go.tmpl`: Business logic orchestration
   - `interface.go.tmpl`: Service contracts
   - `settings_helper.go.tmpl`: Configuration

3. **Repository Layer Templates**
   - `implementation.go.tmpl`: Data access implementation
   - `mappers.go.tmpl`: Domain ↔ Database mapping
   - `filters.go.tmpl`: Query building

4. **API Layer Templates**
   - `handler.go.tmpl`: HTTP request handling
   - `types.go.tmpl`: Request/response models

5. **Component Templates**
   - Individual component generation
   - Specialized for specific patterns

6. **Test Templates**
   - `domain_test.go.tmpl`: Entity and validation tests
   - `service_test.go.tmpl`: Service layer tests
   - `repository_test.go.tmpl`: Data access tests
   - `integration_test.go.tmpl`: Integration tests

### Template Features

- **ERP-Specific Logic**: Financial modules, multi-tenancy, ABAC
- **Smart Naming**: Automatic case conversion (camelCase, PascalCase, snake_case)
- **Conditional Generation**: Context-aware code generation
- **Validation**: Built-in business rules and constraints
- **Observability**: Metrics, tracing, structured logging

##  Integration

### Makefile Integration

The tool is fully integrated with the project Makefile:

```bash
# Install tool
make scaffold-install

# Generate modules
make scaffold-module name=inventory

# Generate components  
make scaffold-component module=finance name=payment type=entity

# Generate features
make scaffold-feature path=finance/budgets

# Show help
make scaffold-help
```

### Development Workflow

```bash
# 1. Generate module
make scaffold-module name=inventory

# 2. Update domain models
# Edit internal/core/inventory/domain/ files

# 3. Generate code
make sqlc goa

# 4. Implement business logic
# Edit service and repository files

# 5. Test
make test
```

### Build System Integration

The tool is included in:
- `make dev-setup`: Includes tool installation
- `make check-tools`: Validates tool availability
- CI/CD pipelines: For automated scaffolding

##  Examples

### Example 1: Customer Management Module

```bash
# Generate customer module
make scaffold-module name=customer

# Add customer segments feature
make scaffold-feature path=customer/segments

# Add specific components
make scaffold-component module=customer name=credit_score entity
make scaffold-component module=customer name=notification service --with-tests
```

### Example 2: Financial Reporting

```bash
# Generate reporting module
make scaffold-module name=reporting --with-tests

# Add specific report types
make scaffold-component module=reporting name=profit_loss_report entity
make scaffold-component module=reporting name=balance_sheet_report entity
make scaffold-component module=reporting name=report_generator service

# Add export functionality
make scaffold-component module=reporting name=pdf_exporter service
```

### Example 3: Inventory Management

```bash
# Generate inventory module
make scaffold-module name=inventory

# Add specialized components
make scaffold-component module=inventory name=stock_movement entity
make scaffold-component module=inventory name=warehouse_location entity
make scaffold-component module=inventory name=stock_validator validator

# Add features
make scaffold-feature path=inventory/adjustments
make scaffold-feature path=inventory/transfers
```

## ✅ Best Practices

### 1. Module Organization

- **Start with modules**: Always begin with `new module` for new domains
- **Logical grouping**: Group related entities within modules
- **Feature separation**: Use features for distinct functionality

### 2. Naming Conventions

- **Module names**: Use singular nouns (`finance`, `inventory`, `payroll`)
- **Component names**: Use descriptive names (`payment_method`, `stock_adjustment`)
- **Feature names**: Use plural or action-oriented names (`budgets`, `reconciliation`)

### 3. Development Workflow

1. **Generate first**: Use scaffolding before manual coding
2. **Review generated code**: Understand the patterns before modification
3. **Implement incrementally**: Start with domain, then service, then API
4. **Test early**: Use `--with-tests` flag and implement tests

### 4. Template Customization

- **Study existing templates**: Understand patterns before modification
- **Follow conventions**: Maintain consistency with generated code
- **Test template changes**: Use dry-run mode extensively

### 5. Integration

- **Use Makefile targets**: Prefer `make scaffold-*` commands
- **Integrate with workflows**: Include in CI/CD processes
- **Document customizations**: Maintain team documentation

## ️ Troubleshooting

### Common Issues

**Tool not found:**
```bash
# Ensure tool is built
make scaffold-install

# Check if in PATH
which awoctl
```

**Permission errors:**
```bash
# Check file permissions
ls -la awoctl

# Make executable
chmod +x awoctl
```

**Template errors:**
```bash
# Validate project structure
./awoctl --help

# Use dry-run to debug
./awoctl new module test --dry-run --verbose
```

**Generated code doesn't compile:**
```bash
# Update dependencies
go mod tidy

# Regenerate shared code
make sqlc goa

# Check imports
go build ./internal/core/[module]/...
```

### Getting Help

```bash
# General help
./awoctl --help

# Command-specific help
./awoctl new module --help
./awoctl new component --help

# Makefile help
make scaffold-help
make help
```

##  Additional Resources

- [Architecture Documentation](../architecture/README.md)
- [Module Development Guide](../development/modules.md)
- [Testing Guidelines](../development/testing.md)
- [API Development](../development/api.md)
- [Template Development](./templates.md)

---

The `awoctl` tool is designed to accelerate development while maintaining code quality and architectural consistency. Use it to bootstrap new functionality and focus on implementing business logic rather than boilerplate code.