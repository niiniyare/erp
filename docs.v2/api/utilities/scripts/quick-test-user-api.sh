#!/bin/bash

# Quick User API Test - Basic functionality demonstration
# This script provides simple curl commands to test the User API

BASE_URL="http://localhost:8080"

echo "🚀 Quick User API Test Suite"
echo "=============================="
echo

# Check if server is running
echo "1. Health Check:"
echo "curl -X GET $BASE_URL/health"
curl -X GET $BASE_URL/health 2>/dev/null && echo || echo "❌ Server not running"
echo

# Example: Create Entity (required for users)
echo "2. Create Test Entity:"
echo "curl -X POST $BASE_URL/api/v1/entities \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{\"name\":\"Test Company\",\"code\":\"TEST001\",\"type\":\"account\"}'"
echo

# Example: Create User
echo "3. Create User:"
echo "curl -X POST $BASE_URL/api/v1/users \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{"
echo "    \"entity_id\": \"YOUR_ENTITY_ID\","
echo "    \"username\": \"john.doe\","
echo "    \"email\": \"john.doe@example.com\","
echo "    \"password\": \"SecurePassword123!\","
echo "    \"user_type\": \"INTERNAL\","
echo "    \"account_status\": \"ACTIVE\""
echo "  }'"
echo

# Example: Authenticate User
echo "4. Authenticate User:"
echo "curl -X POST $BASE_URL/api/v1/users/auth \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{"
echo "    \"identifier\": \"john.doe@example.com\","
echo "    \"password\": \"SecurePassword123!\""
echo "  }'"
echo

# Example: Get User
echo "5. Get User by ID:"
echo "curl -X GET $BASE_URL/api/v1/users/YOUR_USER_ID"
echo

# Example: List Users
echo "6. List Users:"
echo "curl -X GET \"$BASE_URL/api/v1/users?limit=10&offset=0\""
echo

# Example: Search Users
echo "7. Search Users:"
echo "curl -X GET \"$BASE_URL/api/v1/users/search?q=john&limit=5\""
echo

# Example: Update User
echo "8. Update User:"
echo "curl -X PUT $BASE_URL/api/v1/users/YOUR_USER_ID \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{\"mfa_enabled\": true}'"
echo

# Example: Update Password
echo "9. Update Password:"
echo "curl -X PUT $BASE_URL/api/v1/users/YOUR_USER_ID/password \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{"
echo "    \"current_password\": \"SecurePassword123!\","
echo "    \"new_password\": \"NewPassword456!\""
echo "  }'"
echo

# Example: Get User Roles
echo "10. Get User Roles:"
echo "curl -X GET $BASE_URL/api/v1/users/YOUR_USER_ID/roles"
echo

# Example: Delete User
echo "11. Delete User (Soft Delete):"
echo "curl -X DELETE $BASE_URL/api/v1/users/YOUR_USER_ID"
echo

echo "📝 Complete Test Suite:"
echo "For comprehensive testing, run: ./test-user-api.sh"
echo "For detailed documentation, see: user-api-tests.md"
echo

echo "🛠️ To start the server:"
echo "go run cmd/server/main.go"
echo

echo "🔧 Required Environment Variables:"
echo "export DB_USER=admin"
echo "export DB_PASSWORD=admin"
echo "export DB_NAME=ledger"
echo "export DB_HOST=localhost"
echo "export DB_PORT=5432"
echo "export REDIS_HOST=localhost"
echo "export REDIS_PORT=6379"