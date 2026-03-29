# API Testing Scripts

This directory contains organized test scripts for the ERP API endpoints and middleware validation.

## Quick Start

```bash
# Run the interactive test menu
./docs/scripts/api/test.sh

# View test summary  
./docs/scripts/api/test.sh --summary
```

## Test Categories

###  Tenant API Tests
**File:** `test_tenant_api.sh`
- Tests all tenant CRUD operations
- Validates native middleware (header/subdomain extraction)
- Tests public endpoint bypass functionality
- Covers error scenarios and edge cases

### ️ Middleware Integration Tests
**Built into:** `test.sh` (option 4)
- Tests native HTTP middleware chain
- Validates tenant context injection
- Tests public endpoint whitelisting
- Error handling validation

###  Health Check Tests  
**Built into:** `test.sh` (option 5)
- Tests all service health endpoints
- System readiness validation
- Service dependency checks

## Prerequisites

1. **Server Running**: Start the ERP server first
   ```bash
   go run ./cmd/server/
   ```

2. **Dependencies**: Ensure curl is available
   ```bash
   which curl  # Should show curl location
   ```

## Test Scenarios

### Tenant API Test Scenarios
1. **Health Check** (Public endpoint)
2. **List without tenant** (Should fail with 400)  
3. **List with tenant header** (Should succeed)
4. **Get specific tenant** (Tests UUID validation)
5. **Create tenant** (Tests payload validation)
6. **Subdomain extraction** (Tests bo.tenant.domain.com pattern)

### Middleware Test Scenarios
1. **Public endpoint bypass** (No tenant required)
2. **Missing tenant context** (Should return 400)
3. **Invalid UUID format** (Should return 400)
4. **Tenant validation** (Should validate against database)

## Expected Results

| Test | Expected Status | Description |
|------|----------------|-------------|
| Health endpoints | 200 | Public endpoints should always work |
| No tenant header | 400 | Protected endpoints require tenant |
| Valid tenant header | 200/404 | Depends on tenant existence |
| Invalid UUID | 400 | Validation should catch malformed UUIDs |
| Subdomain extraction | 200 | Should extract tenant from subdomain |

## Architecture Validation

These tests validate the Day 4 milestone achievements:
- ✅ **Native HTTP Middleware**: No Gin dependencies
- ✅ **Tenant Isolation**: Context-based RLS 
- ✅ **Public Endpoint Bypass**: Configurable whitelist
- ✅ **Error Handling**: Proper JSON responses
- ✅ **GOA Integration**: Clean API layer

## Adding New Tests

1. **Create test script** in this directory
2. **Add to test.sh menu** with new option
3. **Update README.md** with test details
4. **Follow naming convention**: `test_<service>_api.sh`

## File Structure
```
docs/scripts/api/
├── README.md              # This documentation
├── test.sh               # Main test entry point
├── test_tenant_api.sh    # Tenant API tests
└── test_entity_api.sh    # Entity API tests (coming soon)
```

## Integration with Development Workflow

These scripts are designed to be used during:
- **Development**: Quick validation of API changes
- **Migration**: Validation of Gin to GOA conversion  
- **CI/CD**: Automated testing pipeline (future)
- **Documentation**: Living examples of API usage