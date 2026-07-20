> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Organization Management API - curl Examples

Complete collection of curl commands for testing the Organization Management API.

##  Prerequisites

### Environment Setup
```bash
# Set base URL
export API_BASE_URL="http://localhost:8080"

# Test server connectivity
curl -X GET $API_BASE_URL/api/v1/tenants/health | jq .
```

### Required Tools
```bash
# Install jq for JSON processing
sudo apt-get install jq  # Ubuntu/Debian
brew install jq          # macOS
```

##  Organization CRUD Operations

### 1. Create Organization

#### Basic Organization Creation
```bash
curl -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corporation",
    "organization_type": "CORPORATION"
  }' | jq .
```

#### Create Different Organization Types
```bash
# Corporation
curl -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tech Solutions Inc",
    "organization_type": "CORPORATION"
  }' | jq .

# Department
curl -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Engineering Department",
    "organization_type": "DEPARTMENT"
  }' | jq .

# Team
curl -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Backend Development Team",
    "organization_type": "TEAM"
  }' | jq .

# Non-profit
curl -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Community Foundation",
    "organization_type": "NONPROFIT"
  }' | jq .

# Government
curl -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "City Planning Department",
    "organization_type": "GOVERNMENT"
  }' | jq .
```

#### Save Organization ID for Further Testing
```bash
# Create organization and save ID
ORG_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Organization",
    "organization_type": "CORPORATION"
  }' | jq -r '.id')

echo "Created organization with ID: $ORG_ID"
```

### 2. Get Organization

#### Get Specific Organization
```bash
# Using saved organization ID
curl -X GET $API_BASE_URL/api/v1/organizations/$ORG_ID | jq .

# Using explicit ID
curl -X GET $API_BASE_URL/api/v1/organizations/550e8400-e29b-41d4-a716-446655440000 | jq .
```

#### Get Organization with Error Handling
```bash
# Test with invalid UUID
curl -X GET $API_BASE_URL/api/v1/organizations/invalid-uuid | jq .

# Test with non-existent organization
curl -X GET $API_BASE_URL/api/v1/organizations/00000000-0000-0000-0000-000000000000 | jq .
```

### 3. Update Organization

#### Update Organization Name
```bash
curl -X PUT $API_BASE_URL/api/v1/organizations/$ORG_ID \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Test Organization"
  }' | jq .
```

#### Update Organization Type
```bash
curl -X PUT $API_BASE_URL/api/v1/organizations/$ORG_ID \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated LLC Organization",
    "organization_type": "LLC"
  }' | jq .
```

#### Update with Validation Tests
```bash
# Test empty name (should fail)
curl -X PUT $API_BASE_URL/api/v1/organizations/$ORG_ID \
  -H "Content-Type: application/json" \
  -d '{"name": ""}' | jq .

# Test invalid organization type (should fail)
curl -X PUT $API_BASE_URL/api/v1/organizations/$ORG_ID \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Organization",
    "organization_type": "INVALID_TYPE"
  }' | jq .
```

### 4. List Organizations

#### Basic Listing
```bash
# Get all organizations (default pagination)
curl -X GET $API_BASE_URL/api/v1/organizations | jq .
```

#### Pagination Examples
```bash
# First page with 5 items
curl -X GET "$API_BASE_URL/api/v1/organizations?limit=5&offset=0" | jq .

# Second page with 5 items
curl -X GET "$API_BASE_URL/api/v1/organizations?limit=5&offset=5" | jq .

# Get 10 items starting from item 20
curl -X GET "$API_BASE_URL/api/v1/organizations?limit=10&offset=20" | jq .
```

#### Filtered Listing
```bash
# Filter by organization type
curl -X GET "$API_BASE_URL/api/v1/organizations?organization_type=CORPORATION" | jq .
curl -X GET "$API_BASE_URL/api/v1/organizations?organization_type=DEPARTMENT" | jq .
curl -X GET "$API_BASE_URL/api/v1/organizations?organization_type=TEAM" | jq .

# Filter by status
curl -X GET "$API_BASE_URL/api/v1/organizations?status=ACTIVE" | jq .
curl -X GET "$API_BASE_URL/api/v1/organizations?status=ARCHIVED" | jq .

# Combine filters
curl -X GET "$API_BASE_URL/api/v1/organizations?organization_type=DEPARTMENT&status=ACTIVE&limit=5" | jq .
```

#### Search by Name
```bash
# Partial name matching
curl -X GET "$API_BASE_URL/api/v1/organizations?name=Test" | jq .
curl -X GET "$API_BASE_URL/api/v1/organizations?name=Engineering" | jq .
```

### 5. Organization Hierarchy

#### Get Organization Hierarchy
```bash
# Get complete hierarchy
curl -X GET $API_BASE_URL/api/v1/organizations/$ORG_ID/hierarchy | jq .

# Get hierarchy with depth limit
curl -X GET "$API_BASE_URL/api/v1/organizations/$ORG_ID/hierarchy?depth=2" | jq .

# Get only direct children
curl -X GET "$API_BASE_URL/api/v1/organizations/$ORG_ID/hierarchy?depth=1" | jq .
```

#### Build Complex Hierarchy
```bash
# Create parent corporation
CORP_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Global Tech Corporation",
    "organization_type": "CORPORATION"
  }' | jq -r '.id')

echo "Created corporation: $CORP_ID"

# Create divisions under corporation
PRODUCT_DIV_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Product Division",
    "organization_type": "DIVISION",
    "parent_id": "'$CORP_ID'"
  }' | jq -r '.id')

SERVICES_DIV_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Services Division", 
    "organization_type": "DIVISION",
    "parent_id": "'$CORP_ID'"
  }' | jq -r '.id')

# Create departments under product division
ENG_DEPT_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Engineering Department",
    "organization_type": "DEPARTMENT",
    "parent_id": "'$PRODUCT_DIV_ID'"
  }' | jq -r '.id')

DESIGN_DEPT_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Design Department",
    "organization_type": "DEPARTMENT", 
    "parent_id": "'$PRODUCT_DIV_ID'"
  }' | jq -r '.id')

# Create teams under engineering department
BACKEND_TEAM_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Backend Team",
    "organization_type": "TEAM",
    "parent_id": "'$ENG_DEPT_ID'"
  }' | jq -r '.id')

FRONTEND_TEAM_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Frontend Team",
    "organization_type": "TEAM",
    "parent_id": "'$ENG_DEPT_ID'"
  }' | jq -r '.id')

# Get complete hierarchy
echo "Getting complete hierarchy..."
curl -X GET $API_BASE_URL/api/v1/organizations/$CORP_ID/hierarchy | jq .
```

### 6. Archive Organization

#### Archive Single Organization
```bash
curl -X PATCH $API_BASE_URL/api/v1/organizations/$ORG_ID/archive \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Organizational restructuring"
  }' -v
```

#### Archive with Different Reasons
```bash
# Archive due to merger
curl -X PATCH $API_BASE_URL/api/v1/organizations/$DEPT_ID/archive \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Department merged with another unit"
  }' -v

# Archive due to closure
curl -X PATCH $API_BASE_URL/api/v1/organizations/$TEAM_ID/archive \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Project completed, team disbanded"
  }' -v
```

##   Test Scenarios

### Scenario 1: Complete Organization Lifecycle
```bash
echo "=== Complete Organization Lifecycle Test ==="

# 1. Create organization
echo "Creating organization..."
ORG_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Lifecycle Test Corp",
    "organization_type": "CORPORATION"
  }' | jq -r '.id')
echo "Created organization: $ORG_ID"

# 2. Verify creation
echo "Verifying organization creation..."
curl -s -X GET $API_BASE_URL/api/v1/organizations/$ORG_ID | jq .

# 3. Update organization
echo "Updating organization..."
curl -s -X PUT $API_BASE_URL/api/v1/organizations/$ORG_ID \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Lifecycle Test Corp",
    "organization_type": "LLC"
  }' | jq .

# 4. Create child organization
echo "Creating child organization..."
CHILD_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Child Department",
    "organization_type": "DEPARTMENT",
    "parent_id": "'$ORG_ID'"
  }' | jq -r '.id')

# 5. Get hierarchy
echo "Getting hierarchy..."
curl -s -X GET $API_BASE_URL/api/v1/organizations/$ORG_ID/hierarchy | jq .

# 6. Archive child first
echo "Archiving child organization..."
curl -s -X PATCH $API_BASE_URL/api/v1/organizations/$CHILD_ID/archive \
  -H "Content-Type: application/json" \
  -d '{"reason": "Test cleanup"}' -w "HTTP Status: %{http_code}\n"

# 7. Archive parent
echo "Archiving parent organization..."
curl -s -X PATCH $API_BASE_URL/api/v1/organizations/$ORG_ID/archive \
  -H "Content-Type: application/json" \
  -d '{"reason": "Test cleanup"}' -w "HTTP Status: %{http_code}\n"

echo "=== Lifecycle test completed ==="
```

### Scenario 2: Hierarchy Stress Test
```bash
echo "=== Hierarchy Stress Test ==="

# Create multi-level hierarchy
echo "Creating 4-level hierarchy..."

# Level 1: Corporation
L1_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{"name": "Stress Test Corp", "organization_type": "CORPORATION"}' | jq -r '.id')

# Level 2: Divisions (3 divisions)
DIV_IDS=()
for i in {1..3}; do
  DIV_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
    -H "Content-Type: application/json" \
    -d '{"name": "Division '$i'", "organization_type": "DIVISION", "parent_id": "'$L1_ID'"}' | jq -r '.id')
  DIV_IDS+=($DIV_ID)
  echo "Created Division $i: $DIV_ID"
done

# Level 3: Departments (2 per division)
DEPT_IDS=()
for DIV_ID in "${DIV_IDS[@]}"; do
  for j in {1..2}; do
    DEPT_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
      -H "Content-Type: application/json" \
      -d '{"name": "Department '$j'", "organization_type": "DEPARTMENT", "parent_id": "'$DIV_ID'"}' | jq -r '.id')
    DEPT_IDS+=($DEPT_ID)
    echo "Created Department: $DEPT_ID"
  done
done

# Level 4: Teams (2 per department)
TEAM_IDS=()
for DEPT_ID in "${DEPT_IDS[@]}"; do
  for k in {1..2}; do
    TEAM_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
      -H "Content-Type: application/json" \
      -d '{"name": "Team '$k'", "organization_type": "TEAM", "parent_id": "'$DEPT_ID'"}' | jq -r '.id')
    TEAM_IDS+=($TEAM_ID)
  done
done

echo "Created hierarchy with $(echo ${TEAM_IDS[@]} | wc -w) teams"

# Get complete hierarchy
echo "Getting complete hierarchy..."
curl -s -X GET $API_BASE_URL/api/v1/organizations/$L1_ID/hierarchy | jq '. | {id, name, children: [.children[] | {id, name, children: [.children[] | {id, name, children: [.children[] | {id, name}]}]}]}'

# Cleanup - archive all in reverse order
echo "Cleaning up..."
for TEAM_ID in "${TEAM_IDS[@]}"; do
  curl -s -X PATCH $API_BASE_URL/api/v1/organizations/$TEAM_ID/archive \
    -H "Content-Type: application/json" \
    -d '{"reason": "Stress test cleanup"}' > /dev/null
done

for DEPT_ID in "${DEPT_IDS[@]}"; do
  curl -s -X PATCH $API_BASE_URL/api/v1/organizations/$DEPT_ID/archive \
    -H "Content-Type: application/json" \
    -d '{"reason": "Stress test cleanup"}' > /dev/null
done

for DIV_ID in "${DIV_IDS[@]}"; do
  curl -s -X PATCH $API_BASE_URL/api/v1/organizations/$DIV_ID/archive \
    -H "Content-Type: application/json" \
    -d '{"reason": "Stress test cleanup"}' > /dev/null
done

curl -s -X PATCH $API_BASE_URL/api/v1/organizations/$L1_ID/archive \
  -H "Content-Type: application/json" \
  -d '{"reason": "Stress test cleanup"}' > /dev/null

echo "=== Hierarchy stress test completed ==="
```

### Scenario 3: Error Handling Test
```bash
echo "=== Error Handling Test ==="

# Test invalid JSON
echo "Testing invalid JSON..."
curl -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{"name": "Test"' | jq .

# Test missing required fields
echo "Testing missing required fields..."
curl -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{}' | jq .

# Test invalid organization type
echo "Testing invalid organization type..."
curl -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Organization",
    "organization_type": "INVALID_TYPE"
  }' | jq .

# Test circular reference
echo "Testing circular reference prevention..."
ORG1_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{"name": "Circular Test 1", "organization_type": "CORPORATION"}' | jq -r '.id')

ORG2_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{"name": "Circular Test 2", "organization_type": "DEPARTMENT", "parent_id": "'$ORG1_ID'"}' | jq -r '.id')

# Try to make ORG1 a child of ORG2 (should fail)
curl -X PUT $API_BASE_URL/api/v1/organizations/$ORG1_ID \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Circular Test 1",
    "parent_id": "'$ORG2_ID'"
  }' | jq .

# Test invalid UUID formats
echo "Testing invalid UUID formats..."
curl -X GET $API_BASE_URL/api/v1/organizations/not-a-uuid | jq .
curl -X GET $API_BASE_URL/api/v1/organizations/123 | jq .

# Test invalid pagination parameters
echo "Testing invalid pagination..."
curl -X GET "$API_BASE_URL/api/v1/organizations?limit=abc&offset=xyz" | jq .

echo "=== Error handling test completed ==="
```

### Scenario 4: Organization Types Test
```bash
echo "=== Organization Types Test ==="

# Test all valid organization types
TYPES=("CORPORATION" "LLC" "PARTNERSHIP" "NONPROFIT" "GOVERNMENT" "DEPARTMENT" "DIVISION" "TEAM")

ORG_IDS=()
for TYPE in "${TYPES[@]}"; do
  echo "Creating $TYPE organization..."
  ORG_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
    -H "Content-Type: application/json" \
    -d '{
      "name": "'$TYPE' Test Organization",
      "organization_type": "'$TYPE'"
    }' | jq -r '.id')
  
  ORG_IDS+=($ORG_ID)
  echo "Created $TYPE organization: $ORG_ID"
done

# List organizations by type
for TYPE in "${TYPES[@]}"; do
  echo "Listing $TYPE organizations..."
  curl -s -X GET "$API_BASE_URL/api/v1/organizations?organization_type=$TYPE" | jq '.data | length'
done

# Cleanup
echo "Cleaning up type test organizations..."
for ORG_ID in "${ORG_IDS[@]}"; do
  curl -s -X PATCH $API_BASE_URL/api/v1/organizations/$ORG_ID/archive \
    -H "Content-Type: application/json" \
    -d '{"reason": "Type test cleanup"}' > /dev/null
done

echo "=== Organization types test completed ==="
```

##  Performance and Monitoring

### Response Time Testing
```bash
echo "=== Performance Testing ==="

# Test creation performance
echo "Testing creation performance..."
time curl -s -X POST $API_BASE_URL/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Performance Test Organization",
    "organization_type": "CORPORATION"
  }' > /dev/null

# Test listing performance
echo "Testing listing performance..."
time curl -s -X GET $API_BASE_URL/api/v1/organizations > /dev/null

# Test hierarchy performance
echo "Testing hierarchy performance..."
time curl -s -X GET $API_BASE_URL/api/v1/organizations/$ORG_ID/hierarchy > /dev/null

echo "=== Performance testing completed ==="
```

### Load Testing
```bash
echo "=== Load Testing ==="

# Concurrent organization creation
echo "Testing concurrent creation..."
for i in {1..5}; do
  (curl -s -X POST $API_BASE_URL/api/v1/organizations \
    -H "Content-Type: application/json" \
    -d '{
      "name": "Load Test Organization '$i'",
      "organization_type": "CORPORATION"
    }' > /dev/null) &
done

wait
echo "Concurrent creation test completed"

echo "=== Load testing completed ==="
```

##  Utility Functions

### Organization Management Helper Function
```bash
#!/bin/bash

# Function to create organization hierarchy
create_organization_hierarchy() {
  local corp_name="$1"
  local dept_name="$2"
  local team_name="$3"
  
  if [ -z "$corp_name" ] || [ -z "$dept_name" ] || [ -z "$team_name" ]; then
    echo "Usage: create_organization_hierarchy <corp_name> <dept_name> <team_name>"
    return 1
  fi
  
  # Create corporation
  CORP_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
    -H "Content-Type: application/json" \
    -d "{
      \"name\": \"$corp_name\",
      \"organization_type\": \"CORPORATION\"
    }" | jq -r '.id')
  
  if [ "$CORP_ID" = "null" ]; then
    echo "Failed to create corporation"
    return 1
  fi
  
  # Create department
  DEPT_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
    -H "Content-Type: application/json" \
    -d "{
      \"name\": \"$dept_name\",
      \"organization_type\": \"DEPARTMENT\",
      \"parent_id\": \"$CORP_ID\"
    }" | jq -r '.id')
  
  # Create team
  TEAM_ID=$(curl -s -X POST $API_BASE_URL/api/v1/organizations \
    -H "Content-Type: application/json" \
    -d "{
      \"name\": \"$team_name\",
      \"organization_type\": \"TEAM\",
      \"parent_id\": \"$DEPT_ID\"
    }" | jq -r '.id')
  
  echo "Created hierarchy:"
  echo "  Corporation: $CORP_ID"
  echo "  Department: $DEPT_ID"
  echo "  Team: $TEAM_ID"
  
  # Return corporation ID for hierarchy access
  echo $CORP_ID
}

# Usage example
HIERARCHY_ROOT=$(create_organization_hierarchy "Example Corp" "Engineering Dept" "Backend Team")
if [ $? -eq 0 ]; then
  echo "Getting hierarchy:"
  curl -s -X GET $API_BASE_URL/api/v1/organizations/$HIERARCHY_ROOT/hierarchy | jq .
fi
```

### Organization Statistics Function
```bash
#!/bin/bash

# Function to get organization statistics
get_organization_stats() {
  echo "=== Organization Statistics ==="
  
  # Total count
  TOTAL=$(curl -s -X GET $API_BASE_URL/api/v1/organizations | jq '.pagination.total_items')
  echo "Total Organizations: $TOTAL"
  
  # Count by type
  TYPES=("CORPORATION" "LLC" "PARTNERSHIP" "NONPROFIT" "GOVERNMENT" "DEPARTMENT" "DIVISION" "TEAM")
  for TYPE in "${TYPES[@]}"; do
    COUNT=$(curl -s -X GET "$API_BASE_URL/api/v1/organizations?organization_type=$TYPE" | jq '.pagination.total_items')
    echo "$TYPE: $COUNT"
  done
  
  # Active vs Archived
  ACTIVE=$(curl -s -X GET "$API_BASE_URL/api/v1/organizations?status=ACTIVE" | jq '.pagination.total_items')
  ARCHIVED=$(curl -s -X GET "$API_BASE_URL/api/v1/organizations?status=ARCHIVED" | jq '.pagination.total_items')
  echo "Active: $ACTIVE"
  echo "Archived: $ARCHIVED"
  
  echo "========================"
}

# Run statistics
get_organization_stats
```

##  Common Issues and Solutions

### Issue: Circular Reference Error
```bash
# Check organization hierarchy before making changes
curl -X GET $API_BASE_URL/api/v1/organizations/$ORG_ID/hierarchy | jq .

# Verify parent-child relationships
curl -X GET $API_BASE_URL/api/v1/organizations/$PARENT_ID | jq .
curl -X GET $API_BASE_URL/api/v1/organizations/$CHILD_ID | jq .
```

### Issue: Invalid Organization Type
```bash
# Check valid organization types
echo "Valid types: CORPORATION, LLC, PARTNERSHIP, NONPROFIT, GOVERNMENT, DEPARTMENT, DIVISION, TEAM"

# Validate request before sending
REQUEST='{"name": "Test Org", "organization_type": "CORPORATION"}'
echo $REQUEST | jq empty && echo "Valid JSON" || echo "Invalid JSON"
```

### Issue: Hierarchy Depth Limit
```bash
# Check current hierarchy depth
curl -X GET $API_BASE_URL/api/v1/organizations/$ROOT_ID/hierarchy | \
  jq 'def depth: if .children then ([.children[].children | depth] | max // 0) + 1 else 0 end; depth'
```

---

**Last Updated**: 2025-07-19  
**API Version**: 1.0.0  
**Tested with**: curl 7.81.0, jq 1.6