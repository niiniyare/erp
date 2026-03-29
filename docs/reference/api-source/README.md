# AWO ERP System API Documentation

Welcome to the AWO ERP System API documentation. This directory contains  testing resources, documentation, and utilities for all API endpoints.

##  Directory Structure

```
docs/api/
├── README.md                          # This file - main navigation
├── core-apis/                         # Core business entity APIs
│   ├── entities/                      # Entity Management API
│   ├── users/                         # User Management API
│   └── tenants/                       # Tenant Management API
├── workflows/                         # Workflow and advanced feature APIs
│   ├── access-requests/               # Access Request Workflow API
│   ├── conditional-access/            # Conditional Access Controls API
│   └── user-analytics/                # User Analytics & Behavior API
├── utilities/                         # Development utilities
│   ├── scripts/                       # Automated testing scripts
│   └── postman/                       # Postman collections
└── testing/                           # Testing resources
    ├── automated/                     # Automated test suites
    └── manual/                        # Manual testing guides
```

##  Quick Start

### 1. Setup Environment

```bash
# Database environment variables
export DB_USER=admin
export DB_PASSWORD=admin
export DB_NAME=ledger
export DB_HOST=localhost
export DB_PORT=5432

# Redis environment variables (for caching)
export REDIS_HOST=localhost
export REDIS_PORT=6379

# Start the server
go run cmd/server/main.go
```

### 2. Health Check

```bash
# Verify server is running
curl -X GET http://localhost:8080/api/v1/tenants/health | jq .

# Expected response:
# {
#   "status": "healthy",
#   "timestamp": "2024-01-01T00:00:00Z",
#   "version": "1.0.0"
# }
```

### 3. Install Dependencies

```bash
# Install jq for JSON processing
# Ubuntu/Debian: sudo apt-get install jq
# macOS: brew install jq
# Windows: choco install jq
```

##  Core APIs

### Entity Management API
**Location**: `core-apis/entities/`
- **Purpose**: Manage business entities (companies, departments, products, etc.)
- **Features**: Hierarchical structures, CRUD operations, sequence management
- **Quick Test**: `./utilities/scripts/test-entities.sh`

### User Management API
**Location**: `core-apis/users/`
- **Purpose**: User authentication, authorization, and management
- **Features**: User CRUD, authentication, role management, caching
- **Quick Test**: `./utilities/scripts/test-users.sh`

### Tenant Management API
**Location**: `core-apis/tenants/`
- **Purpose**: Multi-tenant architecture support
- **Features**: Tenant isolation, configuration, management
- **Quick Test**: `./utilities/scripts/test-tenants.sh`

##  Workflow APIs

### Access Request Workflow API
**Location**: `workflows/access-requests/`
- **Purpose**: Manage access requests for roles, permissions, and resources
- **Features**: Request creation, approval workflows, audit trails
- **Quick Test**: `./utilities/scripts/test-access-requests.sh`

### Conditional Access Controls API
**Location**: `workflows/conditional-access/`
- **Purpose**: Dynamic access control based on time, location, device, and risk
- **Features**: Rule engine, real-time evaluation, policy enforcement
- **Quick Test**: `./utilities/scripts/test-conditional-access.sh`

### User Analytics & Behavior API
**Location**: `workflows/user-analytics/`
- **Purpose**: User behavior analysis, risk assessment, and insights
- **Features**: Behavioral patterns, anomaly detection, personalized insights
- **Quick Test**: `./utilities/scripts/test-user-analytics.sh`

##  Testing Resources

### Automated Testing
**Location**: `testing/automated/`
- **Purpose**: Automated test suites for CI/CD pipelines
- **Features**: Unit tests, integration tests, performance tests
- **Run All**: `./testing/automated/run-all-tests.sh`

### Manual Testing
**Location**: `testing/manual/`
- **Purpose**: Step-by-step testing guides for manual verification
- **Features**: Curl commands, expected responses, error scenarios
- **Start Here**: `testing/manual/getting-started.md`

## ️ Development Utilities

### Testing Scripts
**Location**: `utilities/scripts/`
- **Purpose**: Automated testing scripts for quick validation
- **Features**: Colored output, error handling, cleanup
- **Run**: `./utilities/scripts/run-all-tests.sh`

### Postman Collections
**Location**: `utilities/postman/`
- **Purpose**: GUI-based API testing with Postman
- **Features**: Pre-configured requests, environments, tests
- **Import**: Import JSON files into Postman

##  Common Operations

### Full System Test
```bash
# Run  test suite
./utilities/scripts/run-all-tests.sh

# Test specific API
./utilities/scripts/test-entities.sh
./utilities/scripts/test-users.sh
./utilities/scripts/test-access-requests.sh
./utilities/scripts/test-conditional-access.sh
./utilities/scripts/test-user-analytics.sh
```

### Manual Testing
```bash
# Follow step-by-step guides
cat testing/manual/getting-started.md
cat core-apis/entities/curl-examples.md
cat workflows/access-requests/curl-examples.md
```

### Development Workflow
```bash
# 1. Start server
go run cmd/server/main.go

# 2. Run health check
curl -X GET http://localhost:8080/health

# 3. Run automated tests
./utilities/scripts/run-all-tests.sh

# 4. Manual testing as needed
```

##  API Endpoints Summary

### Health & Status
```
GET  /api/v1/tenants/health            # Tenant service health check
GET  /openapi.json                     # OpenAPI specification
GET  /swagger-ui                       # Swagger UI documentation
```

### Authentication
```
POST /api/v1/auth/login                # User login and token generation
POST /api/v1/auth/refresh              # Refresh JWT access token
POST /api/v1/auth/logout               # User logout and token invalidation
GET  /api/v1/auth/validate             # Validate JWT token
```

### Tenant Management
```
POST   /api/v1/tenants                 # Create tenant
GET    /api/v1/tenants                 # List tenants
GET    /api/v1/tenants/{id}            # Get tenant
PUT    /api/v1/tenants/{id}            # Update tenant
DELETE /api/v1/tenants/{id}            # Delete tenant
GET    /api/v1/tenants/health          # Tenant service health
```

### Organization Management
```
POST   /api/v1/organizations           # Create organization
GET    /api/v1/organizations           # List organizations
GET    /api/v1/organizations/{id}      # Get organization
PUT    /api/v1/organizations/{id}      # Update organization
GET    /api/v1/organizations/{id}/hierarchy # Get organization hierarchy
PATCH  /api/v1/organizations/{id}/archive   # Archive organization
```

### User Management
```
POST   /api/v1/users                   # Create user
GET    /api/v1/users                   # List users
GET    /api/v1/users/{id}              # Get user
PUT    /api/v1/users/{id}              # Update user
PATCH  /api/v1/users/{id}/deactivate   # Deactivate user
GET    /api/v1/users/{id}/permissions  # Get user permissions
POST   /api/v1/users/{user_id}/roles   # Assign role to user
DELETE /api/v1/users/{user_id}/roles/{role_id} # Remove role from user
```

### Note: Legacy Workflow APIs
The following APIs are documented but not yet implemented in the current GOA version:
- Access Request Workflow (/api/v1/access-requests/*)
- Conditional Access (/api/v1/conditional-access/*)  
- User Analytics (/api/v1/analytics/users/*)
- Entity Management (/api/v1/entities/*)

##  Testing Strategies

### 1. Unit Testing
- Test individual endpoints
- Validate request/response formats
- Check error handling

### 2. Integration Testing
- Test workflow sequences
- Verify data consistency
- Check cross-service interactions

### 3. Performance Testing
- Load testing with multiple requests
- Response time validation
- Cache performance verification

### 4. Security Testing
- Authentication/authorization validation
- Input validation testing
- SQL injection prevention

##  Troubleshooting Guide

### Common Issues

1. **Server not responding (Connection refused)**
   ```bash
   # Check if server is running
   curl -X GET http://localhost:8080/api/v1/tenants/health
   
   # If not running, start the server
   go run cmd/server/main.go
   ```

2. **Database connection errors**
   ```bash
   # Verify database environment variables
   echo $DB_USER $DB_PASSWORD $DB_NAME $DB_HOST $DB_PORT
   
   # Test database connection
   psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME
   ```

3. **Authentication failures**
   ```bash
   # Ensure X-User-ID header is provided
   curl -X POST http://localhost:8080/api/v1/access-requests \
     -H "X-User-ID: your-user-id" \
     -H "Content-Type: application/json" \
     -d '{...}'
   ```

4. **Invalid UUID errors**
   ```bash
   # Check UUID format (must be valid v4 UUID)
   # Valid: 550e8400-e29b-41d4-a716-446655440000
   # Invalid: invalid-uuid
   ```

5. **JSON parsing errors**
   ```bash
   # Ensure proper JSON formatting
   echo '{"key": "value"}' | jq .
   
   # Use jq to validate JSON before sending
   ```

### Getting Help

1. **Check server logs** for detailed error messages
2. **Review API documentation** in respective folders
3. **Run automated tests** to verify system health
4. **Check environment variables** and configuration
5. **Verify database migrations** are up to date

##  Security Considerations

### Headers Required
- `X-User-ID`: Required for user context
- `Content-Type: application/json`: Required for JSON requests
- `Authorization`: May be required for certain endpoints

### Data Validation
- All UUIDs must be valid v4 format
- Passwords must meet security requirements
- Email addresses must be valid format
- Input sanitization is enforced

### Rate Limiting
- API endpoints may have rate limits
- Use appropriate delays between requests
- Monitor response headers for rate limit info

##  Performance Benchmarks

### Expected Response Times
- Health check: < 10ms
- User authentication: < 100ms
- Entity operations: < 200ms
- Complex analytics: < 2s
- Bulk operations: < 5s

### Caching Strategy
- User data: 15 minutes
- Entity hierarchies: 30 minutes
- Analytics: 1 hour
- Risk assessments: 4 hours

##  Update Guide

### Adding New APIs
1. Create folder under appropriate category
2. Add curl examples and documentation
3. Create automated test script
4. Update this README.md
5. Update run-all-tests.sh

### Modifying Existing APIs
1. Update documentation files
2. Update test scripts
3. Verify backward compatibility
4. Update expected responses

---

**Last Updated**: 2024-07-18
**Version**: 1.0.0
**Maintained by**: AWO ERP System Team

For questions or issues, please refer to the troubleshooting guide or contact the development team.