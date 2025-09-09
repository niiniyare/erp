# Entity API Testing Guide

This directory contains  testing resources for the Entity Management API.

## 📁 Files

- **`api-tests.md`** - Detailed curl commands with examples and explanations
- **`test-entities.sh`** - Automated test script for full API testing
- **`postman-collection.json`** - Postman collection for GUI testing
- **`API_TESTING_README.md`** - This file

## 🚀 Quick Start

### 1. Start the Server

```bash
# Set environment variables
export DB_USER=admin
export DB_PASSWORD=admin
export DB_NAME=ledger
export DB_HOST=localhost
export DB_PORT=5432

# Start the server
./bin/server
```

### 2. Run Automated Tests

```bash
# Make sure jq is installed for JSON processing
# sudo apt-get install jq  # On Ubuntu/Debian
# brew install jq          # On macOS

# Run the automated test suite
./test-entities.sh
```

### 3. Manual Testing with Curl

See `api-tests.md` for detailed curl commands and examples.

### 4. Postman Testing

1. Import `postman-collection.json` into Postman
2. Set the `baseUrl` variable to `http://localhost:8080`
3. Run the collection or individual requests

## 🔧 API Endpoints

### Health Checks
- `GET /health` - Server health status
- `GET /ready` - Server readiness status

### Entity Management
- `POST /api/v1/entities/` - Create entity
- `GET /api/v1/entities/` - List entities
- `GET /api/v1/entities/{id}` - Get entity by ID
- `PUT /api/v1/entities/{id}` - Update entity
- `DELETE /api/v1/entities/{id}` - Soft delete entity
- `POST /api/v1/entities/{id}/restore` - Restore entity

### Hierarchy Operations
- `GET /api/v1/entities/tree` - Get entity tree
- `GET /api/v1/entities/{id}/children` - Get entity children
- `GET /api/v1/entities/{id}/ancestors` - Get entity ancestors
- `GET /api/v1/entities/{id}/hierarchy` - Get entity with hierarchy info

### Sequence Management
- `POST /api/v1/entities/sequence/next` - Get next sequence number
- `POST /api/v1/entities/sequence/reset` - Reset sequence

## 📊 Entity Types

The API supports the following entity types:

- `account` - Chart of accounts entries
- `customer` - Customer entities
- `supplier` - Supplier/vendor entities
- `employee` - Employee entities
- `product` - Product entities
- `service` - Service entities
- `project` - Project entities
- `department` - Department entities
- `location` - Location entities
- `category` - Category entities
- `other` - Other entity types

## 🧪 Testing Scenarios

### Basic CRUD Testing
1. Create entities of different types
2. Retrieve entities by ID
3. Update entity properties
4. Soft delete and restore entities
5. List all entities

### Hierarchy Testing
1. Create parent-child relationships
2. Test multi-level hierarchies
3. Verify ancestor/descendant queries
4. Test entity tree structure

### Validation Testing
1. Test required field validation
2. Test entity type validation
3. Test duplicate code prevention
4. Test invalid UUID handling

### Sequence Testing
1. Generate sequence numbers
2. Test sequence increments
3. Test sequence resets
4. Test fiscal year separation

### Error Handling
1. Test invalid requests
2. Test non-existent entities
3. Test malformed JSON
4. Test server error responses

## 📋 Sample Test Flow

```bash
# 1. Health check
curl -X GET http://localhost:8080/health

# 2. Create parent entity
curl -X POST http://localhost:8080/api/v1/entities/ \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Company", "code": "TEST001", "type": "account"}'

# 3. Create child entity (use parent ID from step 2)
curl -X POST http://localhost:8080/api/v1/entities/ \
  -H "Content-Type: application/json" \
  -d '{"parent_id": "PARENT_UUID", "name": "Sales Dept", "code": "SALES001", "type": "department"}'

# 4. Get entity tree
curl -X GET http://localhost:8080/api/v1/entities/tree

# 5. Get sequence number
curl -X POST http://localhost:8080/api/v1/entities/sequence/next \
  -H "Content-Type: application/json" \
  -d '{"entity_id": "ENTITY_UUID", "key": "invoice", "fiscal_year": 2024}'
```

## 🎯 Expected Responses

### Success Response (201 Created)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Test Company",
  "code": "TEST001",
  "type": "account",
  "is_active": true,
  "is_hidden": false,
  "metadata": {},
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

### Error Response (400 Bad Request)
```json
{
  "error": "validation failed",
  "message": "entity name is required",
  "details": {
    "field": "name",
    "constraint": "required"
  }
}
```

### Entity List Response
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Test Company",
    "code": "TEST001",
    "type": "account",
    "is_active": true,
    "is_hidden": false,
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z"
  }
]
```

## 🔍 Debugging Tips

1. **Check server logs** for detailed error messages
2. **Verify database connection** before running tests
3. **Use jq** for JSON formatting: `curl ... | jq .`
4. **Check entity UUIDs** in responses for subsequent requests
5. **Monitor server metrics** during testing

## 📝 Notes

- All entity codes must be unique within a tenant
- Entity names must be unique within a tenant
- Metadata fields are optional and flexible
- Soft deleted entities don't appear in regular queries
- Sequence numbers are managed per entity/key/fiscal year
- Entity hierarchies use closure table pattern for performance

## 🚨 Common Issues

1. **Entity not found** - Check if entity was soft deleted
2. **Duplicate code error** - Ensure entity codes are unique
3. **Invalid parent** - Verify parent entity exists and is active
4. **Sequence errors** - Check entity ID and fiscal year parameters
5. **Database connection** - Verify environment variables are set

## 🛠️ Troubleshooting

If tests fail:

1. Check if the server is running: `curl http://localhost:8080/health`
2. Verify database connection with correct credentials
3. Check server logs for error details
4. Ensure all required environment variables are set
5. Verify no port conflicts (default: 8080)

For detailed API documentation, see `api-tests.md`.
