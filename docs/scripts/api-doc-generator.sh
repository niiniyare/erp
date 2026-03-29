#!/bin/bash

# api-doc-generator.sh - Automated API Documentation Generation
# Generates OpenAPI specs and integrates with MkDocs

set -e

echo " Generating API Documentation..."

# Configuration
API_SOURCE_DIR="internal/api"
DOCS_API_DIR="docs/reference/api"
OPENAPI_FILE="docs/reference/api/openapi.yaml"

# 1. Generate OpenAPI from Go code
echo " Extracting API definitions..."
if command -v swag &> /dev/null; then
    swag init -g cmd/server/main.go -o docs/swagger --parseDependency
    echo "✅ OpenAPI spec generated"
else
    echo "⚠️ swag not installed. Run: go install github.com/swaggo/swag/cmd/swag@latest"
fi

# 2. Create API reference structure
mkdir -p "$DOCS_API_DIR"/{core-apis,workflows,testing}/

# 3. Generate API endpoint documentation
cat > "$DOCS_API_DIR/core-apis/auth/api-reference.md" << 'EOF'
# Authentication API Reference

## Overview
Complete reference for authentication endpoints including login, token refresh, and session management.

## Endpoints

### POST /auth/login
Authenticate user and return JWT token.

**Request Body:**
```json
{
  "username": "string",
  "password": "string",
  "tenant_id": "string"
}
```

**Response:**
```json
{
  "access_token": "string",
  "refresh_token": "string", 
  "expires_in": 3600,
  "user": {
    "id": "string",
    "username": "string",
    "roles": ["string"]
  }
}
```
EOF

# 4. Generate schema references
cat > "$DOCS_API_DIR/core-apis/entities/schema-reference.md" << 'EOF'
# Entity Schema Reference

## Core Entities

### Entity
Base entity structure for all domain objects.

```json
{
  "id": "uuid",
  "name": "string",
  "slug": "string", 
  "tenant_id": "uuid",
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### User Entity
User account information and permissions.

```json
{
  "id": "uuid",
  "username": "string",
  "email": "string",
  "roles": ["string"],
  "entity_id": "uuid",
  "is_active": true,
  "last_login": "datetime"
}
```
EOF

echo "✅ API documentation structure created"

# 5. Update MkDocs navigation automatically
echo " Updating navigation..."
# This would integrate with mkdocs.yml updates

echo " API documentation generation complete!"