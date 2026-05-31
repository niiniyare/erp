# Tenant Management API - curl Examples

Complete collection of curl commands for testing the Tenant Management API.

##  Prerequisites

### Environment Setup
```bash
# Set base URL
export API_BASE_URL="http://localhost:8080"

# Ensure server is running
curl -X GET $API_BASE_URL/api/v1/tenants/health | jq .
```

### Required Tools
```bash
# Install jq for JSON processing
sudo apt-get install jq  # Ubuntu/Debian
brew install jq          # macOS
```

##  Service Health Check

### Health Status
```bash
# Check tenant service health
curl -X GET $API_BASE_URL/api/v1/tenants/health | jq .

# Expected Response:
# {
#   "status": "healthy",
#   "timestamp": "2024-01-01T00:00:00Z",
#   "version": "1.0.0"
# }
```

##  Tenant CRUD Operations

### 1. Create Tenant

#### Basic Tenant Creation
```bash
curl -X POST $API_BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corporation"
  }' | jq .
```

#### Multiple Tenant Creation
```bash
# Create multiple tenants for testing
curl -X POST $API_BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "Tech Solutions Inc"}' | jq .

curl -X POST $API_BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "Global Manufacturing"}' | jq .

curl -X POST $API_BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "Financial Services Ltd"}' | jq .
```

#### Save Tenant ID for Further Testing
```bash
# Create tenant and save ID
TENANT_ID=$(curl -s -X POST $API_BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Tenant for Operations"}' | jq -r '.id')

echo "Created tenant with ID: $TENANT_ID"
```

### 2. Get Tenant

#### Get Specific Tenant
```bash
# Using saved tenant ID
curl -X GET $API_BASE_URL/api/v1/tenants/$TENANT_ID | jq .

# Using explicit ID
curl -X GET $API_BASE_URL/api/v1/tenants/550e8400-e29b-41d4-a716-446655440000 | jq .
```

#### Get Tenant with Error Handling
```bash
# Test with invalid UUID
curl -X GET $API_BASE_URL/api/v1/tenants/invalid-uuid | jq .

# Test with non-existent tenant
curl -X GET $API_BASE_URL/api/v1/tenants/00000000-0000-0000-0000-000000000000 | jq .
```

### 3. Update Tenant

#### Update Tenant Name
```bash
curl -X PUT $API_BASE_URL/api/v1/tenants/$TENANT_ID \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Test Tenant Name"
  }' | jq .
```

#### Update with Validation Test
```bash
# Test empty name (should fail)
curl -X PUT $API_BASE_URL/api/v1/tenants/$TENANT_ID \
  -H "Content-Type: application/json" \
  -d '{"name": ""}' | jq .

# Test very long name (should fail)
curl -X PUT $API_BASE_URL/api/v1/tenants/$TENANT_ID \
  -H "Content-Type: application/json" \
  -d '{
    "name": "This is a very long tenant name that exceeds the maximum allowed length and should fail validation testing the limits of the system"
  }' | jq .
```

### 4. List Tenants

#### Basic Listing
```bash
# Get all tenants (default pagination)
curl -X GET $API_BASE_URL/api/v1/tenants | jq .
```

#### Pagination Examples
```bash
# First page with 5 items
curl -X GET "$API_BASE_URL/api/v1/tenants?limit=5&offset=0" | jq .

# Second page with 5 items  
curl -X GET "$API_BASE_URL/api/v1/tenants?limit=5&offset=5" | jq .

# Get 10 items starting from item 20
curl -X GET "$API_BASE_URL/api/v1/tenants?limit=10&offset=20" | jq .
```

#### Large Dataset Testing
```bash
# Test pagination edge cases
curl -X GET "$API_BASE_URL/api/v1/tenants?limit=1000&offset=0" | jq .
curl -X GET "$API_BASE_URL/api/v1/tenants?limit=0&offset=0" | jq .
curl -X GET "$API_BASE_URL/api/v1/tenants?limit=-1&offset=-1" | jq .
```

### 5. Delete Tenant

#### Delete Specific Tenant
```bash
# Delete the test tenant
curl -X DELETE $API_BASE_URL/api/v1/tenants/$TENANT_ID -v

# Verify deletion
curl -X GET $API_BASE_URL/api/v1/tenants/$TENANT_ID | jq .
```

#### Delete with Error Handling
```bash
# Test deleting non-existent tenant
curl -X DELETE $API_BASE_URL/api/v1/tenants/00000000-0000-0000-0000-000000000000 -v

# Test deleting with invalid UUID
curl -X DELETE $API_BASE_URL/api/v1/tenants/invalid-uuid -v
```

##   Test Scenarios

### Scenario 1: Complete Tenant Lifecycle
```bash
echo "=== Complete Tenant Lifecycle Test ==="

# 1. Create tenant
echo "Creating tenant..."
TENANT_ID=$(curl -s -X POST $API_BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "Lifecycle Test Tenant"}' | jq -r '.id')
echo "Created tenant: $TENANT_ID"

# 2. Verify creation
echo "Verifying tenant creation..."
curl -s -X GET $API_BASE_URL/api/v1/tenants/$TENANT_ID | jq .

# 3. Update tenant
echo "Updating tenant..."
curl -s -X PUT $API_BASE_URL/api/v1/tenants/$TENANT_ID \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Lifecycle Test Tenant"}' | jq .

# 4. List tenants to verify update
echo "Listing tenants..."
curl -s -X GET $API_BASE_URL/api/v1/tenants | jq '.data[] | select(.id == "'$TENANT_ID'")'

# 5. Delete tenant
echo "Deleting tenant..."
curl -s -X DELETE $API_BASE_URL/api/v1/tenants/$TENANT_ID -w "HTTP Status: %{http_code}\n"

# 6. Verify deletion
echo "Verifying deletion..."
curl -s -X GET $API_BASE_URL/api/v1/tenants/$TENANT_ID | jq .

echo "=== Lifecycle test completed ==="
```

### Scenario 2: Bulk Operations Testing
```bash
echo "=== Bulk Operations Test ==="

# Create multiple tenants
TENANT_IDS=()
for i in {1..5}; do
  TENANT_ID=$(curl -s -X POST $API_BASE_URL/api/v1/tenants \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"Bulk Test Tenant $i\"}" | jq -r '.id')
  TENANT_IDS+=($TENANT_ID)
  echo "Created tenant $i: $TENANT_ID"
done

# List all tenants
echo "Listing all tenants:"
curl -s -X GET $API_BASE_URL/api/v1/tenants | jq '.data | length'

# Update all created tenants
for i in "${!TENANT_IDS[@]}"; do
  curl -s -X PUT $API_BASE_URL/api/v1/tenants/${TENANT_IDS[$i]} \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"Updated Bulk Test Tenant $((i+1))\"}" | jq .
done

# Delete all created tenants
for TENANT_ID in "${TENANT_IDS[@]}"; do
  curl -s -X DELETE $API_BASE_URL/api/v1/tenants/$TENANT_ID
  echo "Deleted tenant: $TENANT_ID"
done

echo "=== Bulk operations test completed ==="
```

### Scenario 3: Error Handling Test
```bash
echo "=== Error Handling Test ==="

# Test invalid JSON
echo "Testing invalid JSON..."
curl -X POST $API_BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "Test"' | jq .

# Test missing required fields
echo "Testing missing required fields..."
curl -X POST $API_BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{}' | jq .

# Test invalid UUID formats
echo "Testing invalid UUID formats..."
curl -X GET $API_BASE_URL/api/v1/tenants/not-a-uuid | jq .
curl -X GET $API_BASE_URL/api/v1/tenants/123 | jq .

# Test invalid pagination parameters
echo "Testing invalid pagination..."
curl -X GET "$API_BASE_URL/api/v1/tenants?limit=abc&offset=xyz" | jq .

echo "=== Error handling test completed ==="
```

### Scenario 4: Performance Test
```bash
echo "=== Performance Test ==="

# Test response times
echo "Testing response times..."

# Health check performance
echo "Health check timing:"
time curl -s -X GET $API_BASE_URL/api/v1/tenants/health > /dev/null

# Create operation performance
echo "Create operation timing:"
time curl -s -X POST $API_BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "Performance Test Tenant"}' > /dev/null

# List operation performance
echo "List operation timing:"
time curl -s -X GET $API_BASE_URL/api/v1/tenants > /dev/null

echo "=== Performance test completed ==="
```

##  Monitoring and Debugging

### Request/Response Logging
```bash
# Enable verbose output for debugging
curl -X GET $API_BASE_URL/api/v1/tenants -v

# Log request/response with timing
curl -X GET $API_BASE_URL/api/v1/tenants \
  -w "Response Code: %{http_code}\nTime Total: %{time_total}s\nTime Connect: %{time_connect}s\n"
```

### Header Analysis
```bash
# Check response headers
curl -I -X GET $API_BASE_URL/api/v1/tenants

# Include headers in response
curl -i -X GET $API_BASE_URL/api/v1/tenants
```

##  Utility Functions

### Tenant Cleanup Function
```bash
#!/bin/bash
# Function to clean up test tenants
cleanup_test_tenants() {
  echo "Cleaning up test tenants..."
  
  # Get all tenants with "Test" in the name
  TEST_TENANT_IDS=$(curl -s -X GET $API_BASE_URL/api/v1/tenants | \
    jq -r '.data[] | select(.name | contains("Test")) | .id')
  
  # Delete each test tenant
  for TENANT_ID in $TEST_TENANT_IDS; do
    curl -s -X DELETE $API_BASE_URL/api/v1/tenants/$TENANT_ID
    echo "Deleted test tenant: $TENANT_ID"
  done
  
  echo "Cleanup completed"
}

# Run cleanup
cleanup_test_tenants
```

### Tenant Statistics Function
```bash
#!/bin/bash
# Function to get tenant statistics
get_tenant_stats() {
  echo "=== Tenant Statistics ==="
  
  # Total count
  TOTAL=$(curl -s -X GET $API_BASE_URL/api/v1/tenants | jq '.pagination.total_items')
  echo "Total Tenants: $TOTAL"
  
  # Active tenants
  ACTIVE=$(curl -s -X GET $API_BASE_URL/api/v1/tenants | \
    jq '.data | map(select(.status == "active")) | length')
  echo "Active Tenants: $ACTIVE"
  
  echo "========================"
}

# Run statistics
get_tenant_stats
```

##  Common Issues and Solutions

### Issue: Connection Refused
```bash
# Check if server is running
curl -X GET $API_BASE_URL/api/v1/tenants/health

# If failed, start the server
# go run cmd/server/main.go
```

### Issue: Invalid JSON Response
```bash
# Verify response format
curl -X GET $API_BASE_URL/api/v1/tenants | python -m json.tool

# Or with jq validation
curl -X GET $API_BASE_URL/api/v1/tenants | jq empty
```

### Issue: Permission Errors
```bash
# Check if authentication headers are required
curl -X GET $API_BASE_URL/api/v1/tenants -H "Authorization: Bearer <your-token>"
```

---

**Last Updated**: 2025-07-19  
**API Version**: 1.0.0  
**Tested with**: curl 7.81.0, jq 1.6