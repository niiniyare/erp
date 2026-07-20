> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

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
