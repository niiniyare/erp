# Quick Start Guide

Get up and running with the ERP System APIs in under 5 minutes.

## 🚀 1. Start the Server

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

## 🔧 2. Install Dependencies

```bash
# Install jq for JSON processing
# Ubuntu/Debian: sudo apt-get install jq
# macOS: brew install jq
# Windows: choco install jq
```

## ✅ 3. Health Check

```bash
curl -X GET http://localhost:8080/health | jq .
```

**Expected Response:**
```json
{
  "service": "awo",
  "status": "ok",
  "version": "1.0.0"
}
```

## 🧪 4. Run All Tests

```bash
# Navigate to API docs directory
cd docs/api

# Run comprehensive test suite
./utilities/scripts/run-all-tests.sh

# Run with verbose output
./utilities/scripts/run-all-tests.sh --verbose

# Run specific API tests
./utilities/scripts/run-all-tests.sh --category core
./utilities/scripts/run-all-tests.sh --category workflow
```

## 📊 5. Test Individual APIs

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

## 🔍 6. Manual Testing

### Create Your First Entity
```bash
curl -X POST http://localhost:8080/api/v1/entities \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Company",
    "code": "COMPANY001",
    "type": "account",
    "is_active": true
  }' | jq .
```

### Create Your First User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "YOUR_ENTITY_ID",
    "username": "john.doe",
    "email": "john.doe@example.com",
    "password": "SecurePassword123!",
    "user_type": "INTERNAL",
    "account_status": "ACTIVE"
  }' | jq .
```

### Create Your First Access Request
```bash
curl -X POST http://localhost:8080/api/v1/access-requests \
  -H "Content-Type: application/json" \
  -H "X-User-ID: YOUR_USER_ID" \
  -d '{
    "entity_id": "YOUR_ENTITY_ID",
    "request_type": "ROLE_ASSIGNMENT",
    "justification": "Need admin access for testing",
    "business_reason": "API Testing",
    "duration_hours": 24
  }' | jq .
```

## 📚 7. Explore Documentation

### Core APIs
- **[Entity Management](core-apis/entities/README.md)** - Business entities and hierarchies
- **[User Management](core-apis/users/README.md)** - User authentication and management

### Workflow APIs
- **[Access Request Workflow](workflows/access-requests/README.md)** - Access request and approval workflows
- **[Conditional Access](workflows/conditional-access/README.md)** - Dynamic access control rules
- **[User Analytics](workflows/user-analytics/README.md)** - User behavior analysis and insights

### Testing Resources
- **[Manual Testing](testing/manual/getting-started.md)** - Step-by-step testing guides
- **[Automated Testing](utilities/scripts/)** - Automated test scripts

## 🎯 8. Common Commands

```bash
# Test everything
./utilities/scripts/run-all-tests.sh

# Test with verbose output
./utilities/scripts/run-all-tests.sh --verbose

# Test specific category
./utilities/scripts/run-all-tests.sh --category core

# Test in parallel
./utilities/scripts/run-all-tests.sh --parallel

# Preserve test data
./utilities/scripts/run-all-tests.sh --no-cleanup

# Get help
./utilities/scripts/run-all-tests.sh --help
```

## 🚨 Troubleshooting

### Server Not Running
```bash
# Check if server is running
curl -X GET http://localhost:8080/health

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

## 🔗 Next Steps

1. **Explore the APIs** - Try the curl examples in each API's documentation
2. **Run Performance Tests** - Use the test scripts to validate performance
3. **Integration Testing** - Test API interactions and workflows
4. **Security Testing** - Validate authentication and authorization
5. **Custom Testing** - Create your own test scenarios

## 📞 Getting Help

- **Documentation**: Browse the `docs/api/` directory
- **Examples**: Check `curl-examples.md` files in each API folder
- **Scripts**: Use automated test scripts in `utilities/scripts/`
- **Troubleshooting**: See the main [README.md](README.md) for common issues

---

**Happy Testing! 🎉**