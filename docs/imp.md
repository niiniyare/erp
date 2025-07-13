# UsersPermision Service Implementation Request

## 📋 Overview
Implement the **UsersPermision** feature following our Clean Architecture patterns. Review existing SQLC queries and database schema to understand the service requirements.

> **📚 Reference Documentation**: Before starting, review our [Developer Guide](./dev/README.md) for architecture patterns and implementation examples.

## 🎯 Implementation Checklist

### 1. **API Contract Design**
- [ ] Define Goa DSL design in `design/{user}.go`
- [ ] Include required endpoints: `{list_main_endpoints}`
- [ ] Follow our [API Layer patterns](./dev/architecture.md#api-layer-responsibilities)
- [ ] Generate interfaces: `goa gen ./design`

### 2. **Service Layer Implementation**
- [ ] Create service in `internal/core/{user}/`
- [ ] Implement business logic following [Service Layer patterns](./dev/best-practices.md#service-layer-responsibilities)
- [ ] Add authorization checks using tenant/entity context
- [ ] Include feature flag gates where applicable
- [ ] Follow [Error Handling patterns](./dev/error-handling.md)

### 3. **Repository Integration**
- [ ] Review existing SQLC queries for data requirements
- [ ] Implement repository following [Repository patterns](./dev/sqlc-integration.md)
- [ ] Ensure proper model conversion between domain and database models
- [ ] Verify Row Level Security (RLS) compliance

### 4. **Observability Integration**
- [ ] Add structured logging using [Logging patterns](./dev/observability.md#structured-logging)
- [ ] Implement distributed tracing following [Tracing patterns](./dev/observability.md#distributed-tracing)
- [ ] Include metrics collection using [Metrics patterns](./dev/observability.md#metrics-collection)

### 5. **UI Layer Configuration**
- [ ] Create AMIS JSON configuration in `web-ui/{user}/`
- [ ] Include context-sensitive elements for multi-tenant support
- [ ] Add client-side validation rules
- [ ] Follow UI/UX patterns established in existing modules

### 6. **Testing Implementation**
- [ ] Unit tests for service layer following [Testing patterns](./dev/best-practices.md#testing-patterns)
- [ ] Integration tests for API endpoints
- [ ] Repository tests with test database
- [ ] Manual verification in UI

## 🔄 Implementation Workflow

1. **Preparation**
   - Review [Architecture Overview](./dev/architecture.md) and [Data Flow patterns](./dev/data-flow.md)
   - Study existing SQLC queries to understand data model
   - Check current database schema for required fields

2. **Development**
   - Follow [Code Examples](./dev/code-examples.md) for complete implementation patterns
   - Use [Best Practices](./dev/best-practices.md) for development guidelines
   - Implement layer by layer: Repository → Service → API → UI

3. **Integration**
   - Add authorization middleware following existing patterns
   - Configure context propagation (tenant/user/entity)
   - Verify RLS policy enforcement

4. **Verification**
   - Test with cURL/Postman using tenant context headers
   - Verify UI functionality with different user roles
   - Check database records respect tenant isolation

## ⚠️ Important Considerations

### **If Service Already Exists**
- Review current implementation against our [Architecture patterns](./dev/architecture.md)
- **Add missing components**: Follow the checklist above for any missing layers
- **Remove anti-patterns**: Check [Common Anti-Patterns](./dev/best-practices.md#common-anti-patterns-to-avoid)
- **Update observability**: Ensure [Observability integration](./dev/observability.md) is complete

### **Multi-Tenant Requirements**
- Always verify tenant context in service layer
- Use RLS as final enforcement layer
- Include entity hierarchy checks where applicable
- Follow [Security Best Practices](./dev/best-practices.md#security-best-practices)

### **Context Propagation**
- Pass context through all layers
- Store tenant/user/entity in context values
- Use context for timeouts and cancellation
- Reference [Error Handling](./dev/error-handling.md) for context-aware error logging

## 🚀 Deployment Steps

1. **Code Generation**
   ```bash
   goa gen ./design
   sqlc generate  # if queries were modified
   ```

2. **Feature Flag Setup**
   - Add feature toggle in admin UI if needed
   - Configure environment-specific settings

3. **Testing**
   - Run test suite: `go test ./...`
   - Manual verification with UI
   - Verify Temporal workflows if applicable

4. **Monitoring**
   - Check metrics dashboard for new service
   - Verify distributed tracing spans
   - Monitor error rates and performance

## 📖 Key Reference Links

- **[Service Documentation](./module/user-permissions.md)** - read user-permissions Documentation
- **[Main Developer Guide](./data-flow-pattern.md)** - Start here for quick navigation
- **[Architecture Overview](./dev/architecture.md)** - Understanding Clean Architecture layers
- **[Complete Code Examples](./dev/code-examples.md)** - Full implementation examples
- **[Best Practices](./dev/best-practices.md)** - Development guidelines and patterns
- **[Error Handling](./dev/error-handling.md)** - Comprehensive error strategies
- **[Observability Guide](./dev/observability.md)** - Logging, tracing, and metrics

