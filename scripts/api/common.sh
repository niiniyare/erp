#!/bin/bash

# Common utilities for API test scripts

# Function to check server connectivity
check_server_connectivity() {
    local server_url="$1"
    local health_endpoint="${server_url}/health"
    
    echo "🔍 Checking server connectivity at $server_url..."
    
    # Try to connect to the health endpoint
    if curl -s -f "$health_endpoint" > /dev/null 2>&1; then
        echo "✅ Server is accessible"
        return 0
    else
        echo "❌ Server is not accessible at $server_url"
        echo "💡 Make sure the server is running with: go run ./cmd/server/"
        return 1
    fi
}

# Function to validate JSON response
validate_json() {
    local response="$1"
    echo "$response" | jq . > /dev/null 2>&1
}

# Function to extract field from JSON
extract_json_field() {
    local json="$1"
    local field="$2"
    echo "$json" | jq -r ".$field // empty"
}

# Function to make HTTP request with proper error handling
make_request() {
    local method="$1"
    local url="$2"
    local headers="$3"
    local data="$4"
    
    local curl_args=(-s -w "%{http_code}")
    
    # Add headers if provided
    if [ -n "$headers" ]; then
        while IFS= read -r header; do
            [ -n "$header" ] && curl_args+=(-H "$header")
        done <<< "$headers"
    fi
    
    # Add data for POST/PUT requests
    if [ -n "$data" ]; then
        curl_args+=(-d "$data")
    fi
    
    # Add method if not GET
    if [ "$method" != "GET" ]; then
        curl_args+=(-X "$method")
    fi
    
    # Make the request
    curl "${curl_args[@]}" "$url"
}

# Function to parse HTTP response
parse_response() {
    local response="$1"
    local http_code="${response: -3}"
    local body="${response%???}"
    
    echo "HTTP_CODE:$http_code"
    echo "BODY:$body"
}

# Function to wait for server to be ready
wait_for_server() {
    local server_url="$1"
    local max_attempts="${2:-30}"
    local sleep_time="${3:-2}"
    
    echo "⏳ Waiting for server to be ready (max ${max_attempts} attempts)..."
    
    for i in $(seq 1 $max_attempts); do
        if check_server_connectivity "$server_url" >/dev/null 2>&1; then
            echo "✅ Server is ready after $i attempts"
            return 0
        fi
        
        echo "  Attempt $i/$max_attempts failed, waiting ${sleep_time}s..."
        sleep "$sleep_time"
    done
    
    echo "❌ Server failed to become ready after $max_attempts attempts"
    return 1
}