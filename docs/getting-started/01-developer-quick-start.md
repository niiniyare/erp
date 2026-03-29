# Quick Start Guide

Get up and running with the AWO ERP System APIs in under 5 minutes.

##  1. Start the Server

```bash
# Set environment variables
export DB_USER=admin
export DB_PASSWORD=admin
export DB_NAME=ledger
export DB_HOST=localhost
export DB_PORT=5432
export REDIS_HOST=localhost
export REDIS_PORT=6379

# Start the server
go run cmd/server/main.go
```

##  2. Install Dependencies

```bash
# Install jq for JSON processing
# Ubuntu/Debian: sudo apt-get install jq
# macOS: brew install jq
# Windows: choco install jq
```

## ✅ 3. Health Check

```bash
curl -X GET http://localhost:8080/api/v1/tenants/health | jq .
```

**Expected Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0"
}
```

##  4. Run All Tests

```bash
# Navigate to API docs directory
cd docs/api

# Run  test suite
./utilities/scripts/run-all-tests.sh

# Run with verbose output
./utilities/scripts/run-all-tests.sh --verbose

# Run specific API tests
./utilities/scripts/run-all-tests.sh --category core
./utilities/scripts/run-all-tests.sh --category workflow
```

##  5. Test Individual APIs

### Entity Management API
```bash
./utilities/scripts/test-entities.sh
```

### User Management API
```bash
./utilities/scripts/test-users.sh
```

### Access Request Workflow API
```bash
./utilities/scripts/test-access-requests.sh
```

### Conditional Access Controls API
```bash
./utilities/scripts/test-conditional-access.sh
```

### User Analytics & Behavior API
```bash
./utilities/scripts/test-user-analytics.sh
```

##  6. Manual Testing

### Create Your First Tenant
```bash
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Company"
  }' | jq .
```

### Create Your First Organization
```bash
curl -X POST http://localhost:8080/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Engineering Department",
    "organization_type": "DEPARTMENT"
  }' | jq .
```

### Create Your First User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john.doe",
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "user_type": "INTERNAL"
  }' | jq .
```

### Authenticate User
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePassword123!"
  }' | jq .
```

##  7. Explore Documentation

### Core APIs
- **[Authentication](../reference/api-source/core-apis/auth/README.md)** - User authentication and JWT token management
- **[Tenant Management](../reference/api-source/core-apis/tenants/README.md)** - Multi-tenant architecture support
- **[Organization Management](../reference/api-source/core-apis/organizations/README.md)** - Organizational structures and hierarchies
- **[User Management](../reference/api-source/core-apis/users/README.md)** - User authentication and management

### Workflow APIs
- **[Access Request Workflow](../reference/api-source/workflows/access-requests/README.md)** - Access request and approval workflows
- **[Conditional Access](../reference/api-source/workflows/conditional-access/README.md)** - Dynamic access control rules
- **[User Analytics](../reference/api-source/workflows/user-analytics/README.md)** - User behavior analysis and insights

### Testing Resources
- **[Manual Testing Guide](../reference/api-source/testing/manual/getting-started.md)** - Step-by-step testing guides

##  8. Common Commands

```bash
# Test everything
cd docs/reference/api-source/utilities/scripts/ && ./run-all-tests.sh
```

##  Troubleshooting

### Server Not Running
```bash
# Check if server is running
curl -X GET http://localhost:8080/api/v1/tenants/health

# If not running, start it
go run cmd/server/main.go
```

### Database Connection Issues
```bash
# Verify environment variables
echo $DB_USER $DB_PASSWORD $DB_NAME $DB_HOST $DB_PORT

# Test database connection
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME
```

### Missing Dependencies
```bash
# Install jq
sudo apt-get install jq  # Ubuntu/Debian
brew install jq          # macOS
choco install jq          # Windows

# Check if jq is working
echo '{"test": "value"}' | jq .
```

##  Next Steps

1. **Explore the APIs** - Try the curl examples in each API's documentation.
2. **Run Performance Tests** - Use the test scripts to validate performance.
3. **Integration Testing** - Test API interactions and workflows.
4. **Security Testing** - Validate authentication and authorization.
5. **Custom Testing** - Create your own test scenarios.

##  Getting Help

- **Documentation**: Browse the `docs/` directory.
- **Examples**: Check `curl-examples.md` files in each API folder.
- **Scripts**: Use automated test scripts in `docs/reference/api-source/utilities/scripts/`.
- **Troubleshooting**: See the main [API README](../reference/api-source/README.md) for common issues.

---

**Happy Testing! **