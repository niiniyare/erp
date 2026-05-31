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
