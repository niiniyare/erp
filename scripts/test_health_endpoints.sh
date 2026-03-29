#!/bin/bash

# Test script for health endpoints - GOA vs Gin comparison
# This script tests both GOA and Gin health endpoints to validate migration

echo " Testing Health Endpoints - GOA vs Gin Migration"
echo "================================================="

# Set test configuration
export SERVER_PORT="8081"
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="admin"
export DB_PASSWORD="admin"
export DB_NAME="ledger"
export REDIS_HOST="localhost"
export REDIS_PORT="6379"

# Start server in background
echo "Starting server on port $SERVER_PORT..."
cd cmd/server
./bin/server &
SERVER_PID=$!

# Wait for server to start
echo "Waiting for server to start..."
sleep 5

# Function to test endpoint
test_endpoint() {
    local url=$1
    local description=$2
    local expected_status=$3
    
    echo ""
    echo "Testing: $description"
    echo "URL: $url"
    echo "----------------------------------------"
    
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$url")
    http_status=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    response_body=$(echo "$response" | sed -e 's/HTTPSTATUS\:.*//g')
    
    if [ "$http_status" -eq "$expected_status" ]; then
        echo "✅ Status: $http_status (Expected: $expected_status)"
        echo " Response: $response_body"
    else
        echo "❌ Status: $http_status (Expected: $expected_status)"
        echo " Response: $response_body"
    fi
}

# Test GOA health endpoints
echo ""
echo " Testing GOA Health Endpoints"
echo "================================"
test_endpoint "http://localhost:$SERVER_PORT/health" "GOA Health Check" 200
test_endpoint "http://localhost:$SERVER_PORT/ready" "GOA Readiness Check" 200

# Test Gin health endpoints (via combined handler)
echo ""
echo " Testing Gin Health Endpoints"
echo "================================"
test_endpoint "http://localhost:$SERVER_PORT/api/v1/health" "Gin Health Check" 200
test_endpoint "http://localhost:$SERVER_PORT/api/v1/ready" "Gin Readiness Check" 200

# Test OpenAPI endpoint
echo ""
echo " Testing OpenAPI Endpoint"
echo "============================"
test_endpoint "http://localhost:$SERVER_PORT/openapi.json" "OpenAPI Specification" 200

# Test Access Request endpoints
echo ""
echo " Testing Access Request Endpoints"
echo "===================================="
test_endpoint "http://localhost:$SERVER_PORT/api/v1/access-requests/stats" "Access Request Stats" 200

echo ""
echo " Cleanup"
echo "==========="
echo "Stopping server (PID: $SERVER_PID)..."
kill $SERVER_PID 2>/dev/null || echo "Server already stopped"

echo ""
echo "✅ Health endpoint testing completed!"
echo "Review the results above to validate GOA migration"