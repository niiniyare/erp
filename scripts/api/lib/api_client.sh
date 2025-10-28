#!/bin/bash

# API Client Library
# Provides functions for interacting with the ERP API.

# Fetches a variable from the environment or uses a default.
# Usage: local value=$(env_get "VAR_NAME" "default_value")
env_get() {
    local var_name="$1"
    local default_value="$2"
    if [ -n "${!var_name}" ]; then
        echo "${!var_name}"
    else
        echo "$default_value"
    fi
}

# Global server configuration
SERVER_PORT=$(env_get "SERVER_PORT" "8080")
SERVER_URL=$(env_get "SERVER_URL" "http://localhost:${SERVER_PORT}")
TENANT_HEADER="X-Tenant-ID"

# Performs a POST request.
# Usage: api::post "path" "tenant_id" "json_payload"
# Returns the full curl response (body + http_code).
api::post() {
    local path="$1"
    local tenant_id="$2"
    local data="$3"
    
    curl -s -w "%{http_code}" \
        -H "Content-Type: application/json" \
        -H "$TENANT_HEADER: $tenant_id" \
        -d "$data" \
        "$SERVER_URL$path"
}

# Performs a GET request.
# Usage: api::get "path" "tenant_id"
# Returns the full curl response (body + http_code).
api::get() {
    local path="$1"
    local tenant_id="$2"
    
    curl -s -w "%{http_code}" \
        -H "$TENANT_HEADER: $tenant_id" \
        "$SERVER_URL$path"
}

# Performs a DELETE request.
# Usage: api::delete "path" "tenant_id"
# Returns the full curl response (body + http_code).
api::delete() {
    local path="$1"
    local tenant_id="$2"
    
    curl -s -w "%{http_code}" -X DELETE \
        -H "$TENANT_HEADER: $tenant_id" \
        "$SERVER_URL$path"
}

# Performs a request to a public endpoint (no tenant ID).
# Usage: api::get_public "path"
api::get_public() {
    local path="$1"
    curl -s -w "%{http_code}" "$SERVER_URL$path"
}
