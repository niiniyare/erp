# AWO ERP System Developer Guide

Welcome to the ERP system development guide. This documentation helps new developers understand our architecture, patterns, and best practices.

## 🚀 Quick Start

**New to the project?** Start here:
1. [Architecture Overview](./architecture.md) - Understand our Clean Architecture pattern
2. [Data Flow Patterns](./data-flow.md) - See how data moves through layers
3. [Code Examples](./code-examples.md) - Practical implementation examples

## 📚 Documentation Structure

### Core Architecture
- **[Architecture Overview](./architecture.md)** - Clean Architecture layers and responsibilities
- **[Data Flow Patterns](./data-flow.md)** - Request/response flow with sequence diagrams

### Technical Integration
- **[SQLC Integration](./sqlc-integration.md)** - Database layer patterns with SQLC
- **[Observability](./observability.md)** - Logging, tracing, and metrics integration

### Development Guidance
- **[Code Examples](./code-examples.md)** - Real-world implementation examples
- **[Best Practices](./best-practices.md)** - Development guidelines and common patterns
- **[Error Handling](./error-handling.md)** - Error flow and handling strategies

## 🎯 Quick Reference

### Common Tasks
| Task | Reference |
|------|-----------|
| Creating a new API endpoint | [Code Examples → Handler Layer](./dev/code-examples.md#handler-layer-example) |
| Adding business logic | [Code Examples → Service Layer](./dev/code-examples.md#service-layer-example) |
| Database operations | [SQLC Integration](./dev/sqlc-integration.md) |
| Adding logging/tracing | [Observability → Layer Integration](./dev/observability.md#layer-specific-observability-patterns) |
| Error handling | [Error Handling Guide](./dev/error-handling.md) |
| Caching patterns | [Best Practices → Caching](./dev/best-practices.md#caching-pattern) |

### Layer Responsibilities Quick Reference
```
API Layer      → Handlers, Middleware, Validation
Core Layer     → Business Logic, Domain Models, Interfaces  
Repository     → Data Access, SQLC Integration, Conversions
Infrastructure → Database, Cache, External Services
```

## 🔧 Development Workflow

1. **API First**: Start with handler definition and request/response models
2. **Service Logic**: Implement business logic in service layer
3. **Repository**: Add data access layer with SQLC
4. **Observability**: Add logging, tracing, and metrics
5. **Testing**: Write unit and integration tests

## 📊 Key Patterns

- **Request Flow**: Client → Handler → Service → Repository → Database
- **Error Flow**: Database → Repository (Wrap) → Service (Handle) → Handler (Format)
- **Dependency Direction**: Higher layers import lower layers, never reverse
- **Observability**: Context-aware logging, distributed tracing, metrics at each layer

## 🚨 Important Guidelines

- ❌ **Never** import lower layers from higher layers
- ✅ **Always** use domain-specific errors
- ✅ **Always** convert models at repository boundaries  
- ✅ **Always** include observability in new code
- ✅ **Always** handle tenant context for multi-tenant operations

## 🆘 Need Help?

- **Architecture questions**: Check [Architecture Overview](./dev/architecture.md)
- **Implementation help**: See [Code Examples](./dev/code-examples.md)
- **Performance concerns**: Review [Observability Guide](./dev/observability.md)
- **Error debugging**: Consult [Error Handling](./dev/error-handling.md)

---

> 💡 **Tip**: Use Ctrl+F to quickly search for specific patterns or concepts across these guides.
