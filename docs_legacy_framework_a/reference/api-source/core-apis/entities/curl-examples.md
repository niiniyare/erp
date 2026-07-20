> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Entity API Test Queries

This document contains structured curl commands for testing the entity management endpoints.

## Prerequisites

Make sure the server is running:
```bash
export DB_USER=admin && export DB_PASSWORD=admin && export DB_NAME=ledger && export DB_HOST=localhost && export DB_PORT=5432 && ./bin/server
```

## Health Check

### Check Server Health
```bash
curl -X GET http://localhost:8080/health
```

### Check Server Readiness
```bash
curl -X GET http://localhost:8080/ready
```

## Entity Management

### 1. Create Entity (Root Level)
```bash
curl -X POST http://localhost:8080/api/v1/entities/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Main Company",
    "code": "MAIN001",
    "type": "account",
    "is_active": true,
    "is_hidden": false,
    "metadata": {
      "description": "Main company entity",
      "industry": "Technology"
    }
  }'
```

### 2. Create Child Entity
```bash
curl -X POST http://localhost:8080/api/v1/entities/ \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "REPLACE_WITH_PARENT_UUID",
    "name": "Sales Department",
    "code": "SALES001",
    "type": "department",
    "is_active": true,
    "is_hidden": false,
    "metadata": {
      "description": "Sales department under main company",
      "manager": "John Doe"
    }
  }'
```

### 3. Create Product Entity
```bash
curl -X POST http://localhost:8080/api/v1/entities/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ERP Software License",
    "code": "ERP-LIC-001",
    "type": "product",
    "is_active": true,
    "is_hidden": false,
    "metadata": {
      "description": "Annual ERP software license",
      "price": 10000,
      "currency": "USD"
    }
  }'
```

### 4. Create Customer Entity
```bash
curl -X POST http://localhost:8080/api/v1/entities/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tech Solutions Inc",
    "code": "CUST-001",
    "type": "customer",
    "is_active": true,
    "is_hidden": false,
    "metadata": {
      "description": "Technology solutions customer",
      "contact_email": "contact@techsolutions.com",
      "phone": "+1-555-0123"
    }
  }'
```

### 5. Create Supplier Entity
```bash
curl -X POST http://localhost:8080/api/v1/entities/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Office Supplies Co",
    "code": "SUP-001",
    "type": "supplier",
    "is_active": true,
    "is_hidden": false,
    "metadata": {
      "description": "Office supplies vendor",
      "contact_email": "orders@officesupplies.com",
      "payment_terms": "NET30"
    }
  }'
```

## Entity Retrieval

### 6. Get Entity by ID
```bash
curl -X GET http://localhost:8080/api/v1/entities/REPLACE_WITH_ENTITY_UUID
```

### 7. List All Entities
```bash
curl -X GET http://localhost:8080/api/v1/entities/
```

### 8. Get Entity Tree Structure
```bash
curl -X GET http://localhost:8080/api/v1/entities/tree
```

### 9. Get Entity with Hierarchy Information
```bash
curl -X GET http://localhost:8080/api/v1/entities/REPLACE_WITH_ENTITY_UUID/hierarchy
```

### 10. Get Entity Children
```bash
curl -X GET http://localhost:8080/api/v1/entities/REPLACE_WITH_PARENT_UUID/children
```

### 11. Get Entity Ancestors
```bash
curl -X GET http://localhost:8080/api/v1/entities/REPLACE_WITH_CHILD_UUID/ancestors
```

## Entity Updates

### 12. Update Entity
```bash
curl -X PUT http://localhost:8080/api/v1/entities/REPLACE_WITH_ENTITY_UUID \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Company Name",
    "is_active": false,
    "metadata": {
      "description": "Updated company description",
      "last_modified": "2024-01-01"
    }
  }'
```

### 13. Update Entity Type
```bash
curl -X PUT http://localhost:8080/api/v1/entities/REPLACE_WITH_ENTITY_UUID \
  -H "Content-Type: application/json" \
  -d '{
    "type": "location",
    "metadata": {
      "address": "123 Main Street",
      "city": "New York"
    }
  }'
```

## Entity Deletion and Restoration

### 14. Soft Delete Entity
```bash
curl -X DELETE http://localhost:8080/api/v1/entities/REPLACE_WITH_ENTITY_UUID
```

### 15. Restore Deleted Entity
```bash
curl -X POST http://localhost:8080/api/v1/entities/REPLACE_WITH_ENTITY_UUID/restore
```

## Sequence Management

### 16. Get Next Sequence Number
```bash
curl -X POST http://localhost:8080/api/v1/entities/sequence/next \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "REPLACE_WITH_ENTITY_UUID",
    "key": "invoice",
    "fiscal_year": 2024
  }'
```

### 17. Reset Sequence
```bash
curl -X POST http://localhost:8080/api/v1/entities/sequence/reset \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "REPLACE_WITH_ENTITY_UUID",
    "key": "invoice",
    "fiscal_year": 2024
  }'
```

## Entity Types

The following entity types are supported:
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
- `other` - Other types

## Error Handling Examples

### 18. Create Entity with Invalid Data
```bash
curl -X POST http://localhost:8080/api/v1/entities/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "",
    "code": "INVALID",
    "type": "invalid_type"
  }'
```

### 19. Get Non-existent Entity
```bash
curl -X GET http://localhost:8080/api/v1/entities/00000000-0000-0000-0000-000000000000
```

### 20. Create Entity with Duplicate Code
```bash
curl -X POST http://localhost:8080/api/v1/entities/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Duplicate Test",
    "code": "MAIN001",
    "type": "account"
  }'
```

## Advanced Queries

### 21. Create Multi-level Hierarchy
```bash
# Create parent company
curl -X POST http://localhost:8080/api/v1/entities/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Global Corp",
    "code": "GLOBAL",
    "type": "account",
    "is_active": true
  }'

# Create subsidiary (use parent ID from above)
curl -X POST http://localhost:8080/api/v1/entities/ \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "REPLACE_WITH_PARENT_UUID",
    "name": "Regional Office",
    "code": "REGION",
    "type": "department",
    "is_active": true
  }'

# Create department under subsidiary
curl -X POST http://localhost:8080/api/v1/entities/ \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "REPLACE_WITH_REGIONAL_UUID",
    "name": "IT Department",
    "code": "IT-DEPT",
    "type": "department",
    "is_active": true
  }'
```

### 22. Test Entity Validation
```bash
# Test name validation (too short)
curl -X POST http://localhost:8080/api/v1/entities/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "A",
    "code": "SHORT",
    "type": "account"
  }'

# Test code validation (too short)
curl -X POST http://localhost:8080/api/v1/entities/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Valid Name",
    "code": "A",
    "type": "account"
  }'
```

## Response Examples

### Successful Entity Creation Response
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Main Company",
  "code": "MAIN001",
  "type": "account",
  "is_active": true,
  "is_hidden": false,
  "metadata": {
    "description": "Main company entity",
    "industry": "Technology"
  },
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

### Entity List Response
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Main Company",
    "code": "MAIN001",
    "type": "account",
    "is_active": true,
    "is_hidden": false,
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z"
  }
]
```

### Entity Tree Response
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Main Company",
    "code": "MAIN001",
    "type": "account",
    "level": 0,
    "path": "Main Company",
    "has_children": true,
    "children": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440001",
        "name": "Sales Department",
        "code": "SALES001",
        "type": "department",
        "level": 1,
        "path": "Main Company/Sales Department",
        "has_children": false
      }
    ]
  }
]
```

### Error Response
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

## Testing Workflow

1. **Start with health checks** to ensure server is running
2. **Create root entities** (companies, accounts)
3. **Create child entities** (departments, products, customers)
4. **Test hierarchy operations** (tree, children, ancestors)
5. **Test updates and validation**
6. **Test deletion and restoration**
7. **Test sequence management**
8. **Test error conditions**

## Notes

- Replace `REPLACE_WITH_ENTITY_UUID` with actual UUIDs from creation responses
- Replace `REPLACE_WITH_PARENT_UUID` with parent entity UUIDs
- All timestamps are in ISO 8601 format
- Entity codes must be unique within the tenant
- Entity names must be unique within the tenant
- Metadata is optional and can contain any valid JSON
- Soft deleted entities can be restored
- Sequence numbers are managed per entity, key, and fiscal year

## Environment Setup

Before running tests, ensure your environment is set up:

```bash
# Database environment variables
export DB_USER=admin
export DB_PASSWORD=admin
export DB_NAME=ledger
export DB_HOST=localhost
export DB_PORT=5432

# Start the server
./bin/server
```

The server should be running on `http://localhost:8080` for all these curl commands to work.