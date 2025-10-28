#!/bin/bash

# Entity API Test Script
# This script demonstrates the entity management API endpoints

set -e

BASE_URL="http://localhost:8080"
API_PREFIX="/api/v1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Function to make API calls and handle responses
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    print_info "Testing: $description"
    
    if [ -n "$data" ]; then
        response=$(curl -s -X "$method" "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -X "$method" "$BASE_URL$endpoint")
    fi
    
    echo "Response: $response"
    echo "$response" | jq . 2>/dev/null || echo "Response is not valid JSON"
    echo ""
    
    # Return the response for further processing
    echo "$response"
}

# Start testing
print_step "Starting Entity API Tests"

# 1. Health Check
print_step "Health Checks"
api_call "GET" "/health" "" "Server health check"
api_call "GET" "/ready" "" "Server readiness check"

# 2. Create Root Entity (Company)
print_step "Creating Root Entity"
company_response=$(api_call "POST" "$API_PREFIX/entities/" '{
    "name": "Tech Solutions Corp",
    "code": "TECH001",
    "type": "account",
    "is_active": true,
    "is_hidden": false,
    "metadata": {
        "description": "Main technology solutions company",
        "industry": "Technology",
        "founded": "2020"
    }
}' "Create main company entity")

# Extract company ID from response
company_id=$(echo "$company_response" | jq -r '.id' 2>/dev/null || echo "")
if [ "$company_id" = "null" ] || [ -z "$company_id" ]; then
    print_error "Failed to create company entity"
    exit 1
fi
print_success "Created company with ID: $company_id"

# 3. Create Department (Child Entity)
print_step "Creating Department Entity"
dept_response=$(api_call "POST" "$API_PREFIX/entities/" "{
    \"parent_id\": \"$company_id\",
    \"name\": \"Engineering Department\",
    \"code\": \"ENG001\",
    \"type\": \"department\",
    \"is_active\": true,
    \"is_hidden\": false,
    \"metadata\": {
        \"description\": \"Software engineering department\",
        \"manager\": \"John Smith\",
        \"budget\": 500000
    }
}" "Create engineering department")

dept_id=$(echo "$dept_response" | jq -r '.id' 2>/dev/null || echo "")
if [ "$dept_id" = "null" ] || [ -z "$dept_id" ]; then
    print_error "Failed to create department entity"
    exit 1
fi
print_success "Created department with ID: $dept_id"

# 4. Create Product Entity
print_step "Creating Product Entity"
product_response=$(api_call "POST" "$API_PREFIX/entities/" '{
    "name": "ERP Software Suite",
    "code": "ERP-SUITE-001",
    "type": "product",
    "is_active": true,
    "is_hidden": false,
    "metadata": {
        "description": "Complete ERP software solution",
        "price": 50000,
        "currency": "USD",
        "license_type": "annual"
    }
}' "Create product entity")

product_id=$(echo "$product_response" | jq -r '.id' 2>/dev/null || echo "")
print_success "Created product with ID: $product_id"

# 5. Create Customer Entity
print_step "Creating Customer Entity"
customer_response=$(api_call "POST" "$API_PREFIX/entities/" '{
    "name": "Global Manufacturing Inc",
    "code": "CUST-GM-001",
    "type": "customer",
    "is_active": true,
    "is_hidden": false,
    "metadata": {
        "description": "Manufacturing customer",
        "contact_email": "procurement@globalmanufacturing.com",
        "phone": "+1-555-0199",
        "credit_limit": 100000
    }
}' "Create customer entity")

customer_id=$(echo "$customer_response" | jq -r '.id' 2>/dev/null || echo "")
print_success "Created customer with ID: $customer_id"

# 6. Create Supplier Entity
print_step "Creating Supplier Entity"
supplier_response=$(api_call "POST" "$API_PREFIX/entities/" '{
    "name": "Cloud Services Provider",
    "code": "SUP-CLOUD-001",
    "type": "supplier",
    "is_active": true,
    "is_hidden": false,
    "metadata": {
        "description": "Cloud infrastructure provider",
        "contact_email": "billing@cloudservices.com",
        "payment_terms": "NET15",
        "service_level": "premium"
    }
}' "Create supplier entity")

supplier_id=$(echo "$supplier_response" | jq -r '.id' 2>/dev/null || echo "")
print_success "Created supplier with ID: $supplier_id"

# 7. List All Entities
print_step "Listing All Entities"
api_call "GET" "$API_PREFIX/entities/" "" "List all entities"

# 8. Get Entity Tree
print_step "Getting Entity Tree"
api_call "GET" "$API_PREFIX/entities/tree" "" "Get entity tree structure"

# 9. Get Specific Entity
print_step "Getting Specific Entity"
api_call "GET" "$API_PREFIX/entities/$company_id" "" "Get company entity by ID"

# 10. Get Entity with Hierarchy
print_step "Getting Entity with Hierarchy Info"
api_call "GET" "$API_PREFIX/entities/$company_id/hierarchy" "" "Get entity with hierarchy information"

# 11. Get Entity Children
print_step "Getting Entity Children"
api_call "GET" "$API_PREFIX/entities/$company_id/children" "" "Get company children"

# 12. Get Entity Ancestors
print_step "Getting Entity Ancestors"
api_call "GET" "$API_PREFIX/entities/$dept_id/ancestors" "" "Get department ancestors"

# 13. Update Entity
print_step "Updating Entity"
api_call "PUT" "$API_PREFIX/entities/$company_id" '{
    "name": "Tech Solutions Corporation",
    "metadata": {
        "description": "Updated main technology solutions company",
        "industry": "Technology",
        "founded": "2020",
        "last_updated": "2024-01-01"
    }
}' "Update company entity"

# 14. Test Sequence Management
print_step "Testing Sequence Management"
api_call "POST" "$API_PREFIX/entities/sequence/next" "{
    \"entity_id\": \"$company_id\",
    \"key\": \"invoice\",
    \"fiscal_year\": 2024
}" "Get next invoice sequence"

api_call "POST" "$API_PREFIX/entities/sequence/next" "{
    \"entity_id\": \"$company_id\",
    \"key\": \"invoice\",
    \"fiscal_year\": 2024
}" "Get next invoice sequence (should increment)"

# 15. Test Error Conditions
print_step "Testing Error Conditions"

# Invalid entity type
api_call "POST" "$API_PREFIX/entities/" '{
    "name": "Invalid Type Test",
    "code": "INVALID001",
    "type": "invalid_type"
}' "Test invalid entity type (should fail)"

# Missing required fields
api_call "POST" "$API_PREFIX/entities/" '{
    "code": "INCOMPLETE001"
}' "Test missing required fields (should fail)"

# Duplicate code
api_call "POST" "$API_PREFIX/entities/" '{
    "name": "Duplicate Code Test",
    "code": "TECH001",
    "type": "account"
}' "Test duplicate code (should fail)"

# Get non-existent entity
api_call "GET" "$API_PREFIX/entities/00000000-0000-0000-0000-000000000000" "" "Test non-existent entity (should fail)"

# 16. Test Soft Delete and Restore
print_step "Testing Soft Delete and Restore"

# Create a test entity for deletion
delete_test_response=$(api_call "POST" "$API_PREFIX/entities/" '{
    "name": "Delete Test Entity",
    "code": "DELETE001",
    "type": "other",
    "is_active": true
}' "Create entity for deletion test")

delete_test_id=$(echo "$delete_test_response" | jq -r '.id' 2>/dev/null || echo "")
if [ "$delete_test_id" != "null" ] && [ -n "$delete_test_id" ]; then
    print_success "Created test entity for deletion: $delete_test_id"
    
    # Soft delete
    api_call "DELETE" "$API_PREFIX/entities/$delete_test_id" "" "Soft delete entity"
    
    # Try to get deleted entity (should fail)
    api_call "GET" "$API_PREFIX/entities/$delete_test_id" "" "Try to get deleted entity (should fail)"
    
    # Restore entity
    api_call "POST" "$API_PREFIX/entities/$delete_test_id/restore" "" "Restore deleted entity"
    
    # Get restored entity (should work)
    api_call "GET" "$API_PREFIX/entities/$delete_test_id" "" "Get restored entity (should work)"
fi

# 17. Performance Test - Create Multiple Entities
print_step "Performance Test - Creating Multiple Entities"
for i in {1..5}; do
    api_call "POST" "$API_PREFIX/entities/" "{
        \"name\": \"Test Entity $i\",
        \"code\": \"TEST$(printf '%03d' $i)\",
        \"type\": \"other\",
        \"is_active\": true,
        \"metadata\": {
            \"test_number\": $i,
            \"created_by\": \"test_script\"
        }
    }" "Create test entity $i"
done

# Final entity list
print_step "Final Entity List"
api_call "GET" "$API_PREFIX/entities/" "" "Final list of all entities"

print_step "Entity API Tests Completed"
print_success "All tests completed successfully!"

# Summary
echo ""
print_step "Test Summary"
echo "Created entities:"
echo "- Company: $company_id"
echo "- Department: $dept_id"
echo "- Product: $product_id"
echo "- Customer: $customer_id"
echo "- Supplier: $supplier_id"
echo "- Delete Test: $delete_test_id"
echo "- Performance Test: 5 additional entities"
echo ""
print_info "Check the server logs for detailed operation traces"
print_info "Use the created entity IDs for further testing"