# Entity Management API

The Entity Management API provides  CRUD operations for managing business entities in a hierarchical structure.

## 📋 Overview

Entities represent core business objects like companies, departments, products, customers, suppliers, and more. The API supports:

- **Hierarchical Structure**: Parent-child relationships with unlimited nesting
- **Entity Types**: Pre-defined types (account, customer, supplier, employee, etc.)
- **Metadata Support**: Flexible JSON metadata for custom fields
- **Sequence Management**: Auto-increment sequences for numbering
- **Soft Deletion**: Entities can be deleted and restored
- **Audit Trail**: Full creation/modification tracking

## 🔧 Base URL

```
http://localhost:8080/api/v1/entities
```

## 📊 Entity Types

| Type | Description | Use Case |
|------|-------------|----------|
| `account` | Chart of accounts entries | Financial accounting |
| `customer` | Customer entities | CRM and sales |
| `supplier` | Supplier/vendor entities | Procurement |
| `employee` | Employee entities | HR management |
| `product` | Product entities | Inventory management |
| `service` | Service entities | Service management |
| `project` | Project entities | Project management |
| `department` | Department entities | Organizational structure |
| `location` | Location entities | Geographic management |
| `category` | Category entities | Classification |
| `other` | Other entity types | Custom use cases |

## 🚀 Quick Start

### 1. Health Check
```bash
curl -X GET http://localhost:8080/health | jq .
```

### 2. Create Root Entity
```bash
curl -X POST http://localhost:8080/api/v1/entities \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Main Company",
    "code": "MAIN001",
    "type": "account",
    "is_active": true,
    "metadata": {
      "description": "Main company entity"
    }
  }' | jq .
```

### 3. List Entities
```bash
curl -X GET http://localhost:8080/api/v1/entities | jq .
```

### 4. Get Entity Tree
```bash
curl -X GET http://localhost:8080/api/v1/entities/tree | jq .
```

## 📚 Documentation Files

- **[curl-examples.md](curl-examples.md)** -  curl command examples
- **[API Reference](api-reference.md)** - Detailed API specification
- **[Schema Reference](schema-reference.md)** - Request/response schemas

## 🧪 Testing

### Automated Testing
```bash
# Run entity API tests
./utilities/scripts/test-entities.sh

# Run with verbose output
./utilities/scripts/test-entities.sh -v
```

### Manual Testing
```bash
# Follow step-by-step examples
cat curl-examples.md
```

## 🔍 Common Use Cases

### Creating Hierarchical Structure
```bash
# 1. Create parent company
PARENT_ID=$(curl -s -X POST http://localhost:8080/api/v1/entities \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Global Corp",
    "code": "GLOBAL",
    "type": "account"
  }' | jq -r '.id')

# 2. Create subsidiary
curl -X POST http://localhost:8080/api/v1/entities \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "'$PARENT_ID'",
    "name": "Regional Office",
    "code": "REGION",
    "type": "department"
  }' | jq .
```

### Sequence Management
```bash
# Get next sequence number
curl -X POST http://localhost:8080/api/v1/entities/sequence/next \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "'$ENTITY_ID'",
    "key": "invoice",
    "fiscal_year": 2024
  }' | jq .
```

## 🎯 Expected Responses

### Success Response (201 Created)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Main Company",
  "code": "MAIN001",
  "type": "account",
  "is_active": true,
  "is_hidden": false,
  "metadata": {
    "description": "Main company entity"
  },
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

## 🚨 Common Issues

1. **Duplicate code error** - Entity codes must be unique
2. **Invalid parent** - Parent entity must exist and be active
3. **Entity not found** - Check if entity was soft deleted
4. **Validation errors** - Check required fields and data types

## 📈 Performance Notes

- Entity hierarchies use closure table pattern for performance
- Caching is implemented for frequently accessed entities
- Tree operations are optimized for large hierarchies
- Soft deletion preserves referential integrity

## 🔗 Related APIs

- **User Management**: Users belong to entities
- **Access Requests**: Requests are scoped to entities
- **Tenant Management**: Entities belong to tenants