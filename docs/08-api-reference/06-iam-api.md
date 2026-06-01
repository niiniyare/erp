---
title: IAM API
portal: 8 — API Reference
section: 08-api-reference
audience: [backend-engineer, frontend-engineer, integrator]
related:
  - "[API Reference Overview](01-api-reference-overview.md)"
  - "[Authentication](02-authentication.md)"
  - "[User Roles Reference](../02-product-overview/03-user-roles.md)"
---

# IAM API

Manage users, roles, permissions, and role assignments within a tenant.

## Users

### List Users

```
GET /api/v1/tenants/{tenant_id}/users
```

**Query parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | Page number (default: 1) |
| `page_size` | int | Results per page (default: 20, max: 100) |
| `status` | string | Filter: `active`, `inactive`, `pending` |
| `role` | string | Filter by role name |
| `search` | string | Search by name or email |

**Response 200:**

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "jane.smith@example.com",
      "display_name": "Jane Smith",
      "status": "active",
      "roles": ["contracts.editor", "finance.viewer"],
      "entity_scope": {
        "type": "subtree",
        "entity_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
      },
      "created_at": "2025-01-10T08:00:00Z",
      "last_login_at": "2025-05-20T14:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 42,
    "total_pages": 3
  }
}
```

### Get User

```
GET /api/v1/tenants/{tenant_id}/users/{user_id}
```

**Response 200:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "jane.smith@example.com",
  "display_name": "Jane Smith",
  "status": "active",
  "roles": ["contracts.editor", "finance.viewer"],
  "entity_scope": {
    "type": "subtree",
    "entity_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
  },
  "permissions": [
    "contracts.contract.create",
    "contracts.contract.read",
    "contracts.contract.update",
    "contracts.contract.submit",
    "finance.account.read",
    "finance.transaction.read",
    "finance.report.read"
  ],
  "created_at": "2025-01-10T08:00:00Z",
  "updated_at": "2025-03-15T10:00:00Z",
  "last_login_at": "2025-05-20T14:30:00Z"
}
```

### Create User

```
POST /api/v1/tenants/{tenant_id}/users
```

**Request body:**

```json
{
  "email": "john.doe@example.com",
  "display_name": "John Doe",
  "roles": ["contracts.viewer"],
  "entity_scope": {
    "type": "all"
  }
}
```

**Entity scope types:**

| Type | Additional field | Meaning |
|------|-----------------|---------|
| `all` | — | Access to all tenant data |
| `subtree` | `entity_id` | Entity + all descendants |
| `entity` | `entity_id` | Single entity only |

**Response 201:**

```json
{
  "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "email": "john.doe@example.com",
  "display_name": "John Doe",
  "status": "pending",
  "roles": ["contracts.viewer"],
  "entity_scope": { "type": "all" },
  "created_at": "2025-05-31T12:00:00Z"
}
```

An invitation email is sent to the user. Status moves from `pending` → `active` on first login.

### Update User

```
PATCH /api/v1/tenants/{tenant_id}/users/{user_id}
```

**Request body** (all fields optional):

```json
{
  "display_name": "Jane Smith-Jones",
  "entity_scope": {
    "type": "subtree",
    "entity_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
  }
}
```

Cannot update `email` or `roles` via this endpoint — use dedicated role assignment endpoints.

### Deactivate User

```
POST /api/v1/tenants/{tenant_id}/users/{user_id}/deactivate
```

Sets status to `inactive`. Revokes all active sessions. User cannot log in until reactivated.

**Response 200:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "inactive"
}
```

### Reactivate User

```
POST /api/v1/tenants/{tenant_id}/users/{user_id}/reactivate
```

Sets status back to `active`.

---

## Role Assignments

### Assign Role

```
POST /api/v1/tenants/{tenant_id}/users/{user_id}/roles
```

**Request body:**

```json
{
  "role": "contracts.approver"
}
```

**Response 200:**

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "roles": ["contracts.editor", "contracts.approver", "finance.viewer"]
}
```

### Remove Role

```
DELETE /api/v1/tenants/{tenant_id}/users/{user_id}/roles/{role_name}
```

**Response 200:**

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "roles": ["contracts.editor", "finance.viewer"]
}
```

---

## Roles

### List Roles

```
GET /api/v1/tenants/{tenant_id}/roles
```

**Response 200:**

```json
{
  "data": [
    {
      "id": "a3f8e2b1-...",
      "name": "contracts.editor",
      "display_name": "Contracts Editor",
      "is_system": true,
      "permissions": [
        "contracts.contract.create",
        "contracts.contract.read",
        "contracts.contract.update",
        "contracts.contract.submit"
      ],
      "created_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": "b4c7d3e2-...",
      "name": "procurement.reviewer",
      "display_name": "Procurement Reviewer",
      "is_system": false,
      "permissions": [
        "contracts.contract.read",
        "contracts.contract.approve"
      ],
      "created_at": "2025-03-10T09:00:00Z"
    }
  ],
  "pagination": { "page": 1, "page_size": 20, "total": 12, "total_pages": 1 }
}
```

### Get Role

```
GET /api/v1/tenants/{tenant_id}/roles/{role_name}
```

Returns same shape as list item, plus `assigned_users_count`.

### Create Role

```
POST /api/v1/tenants/{tenant_id}/roles
```

**Request body:**

```json
{
  "name": "procurement.reviewer",
  "display_name": "Procurement Reviewer",
  "permissions": [
    "contracts.contract.read",
    "contracts.contract.approve",
    "finance.account.read"
  ]
}
```

Role name must be `{prefix}.{verb}` format, unique within the tenant. System roles cannot be created via API.

**Response 201:** Role object.

### Update Role

```
PATCH /api/v1/tenants/{tenant_id}/roles/{role_name}
```

**Request body:**

```json
{
  "display_name": "Senior Procurement Reviewer",
  "permissions": [
    "contracts.contract.read",
    "contracts.contract.approve",
    "contracts.contract.update",
    "finance.account.read"
  ]
}
```

Cannot update system roles (`is_system: true`). Permission update replaces the full permission set.

### Delete Role

```
DELETE /api/v1/tenants/{tenant_id}/roles/{role_name}
```

Fails with `409 Conflict` if role is currently assigned to any users.

---

## Permissions

### List Available Permissions

```
GET /api/v1/tenants/{tenant_id}/permissions
```

Returns all permission strings defined by enabled modules.

**Response 200:**

```json
{
  "data": [
    {
      "name": "contracts.contract.create",
      "module": "contracts",
      "resource": "contract",
      "action": "create",
      "description": "Create new contracts"
    },
    {
      "name": "contracts.contract.read",
      "module": "contracts",
      "resource": "contract",
      "action": "read",
      "description": "Read contracts and related data"
    }
  ]
}
```

### Check Permission

```
POST /api/v1/tenants/{tenant_id}/users/{user_id}/check-permission
```

Evaluate whether a user has a specific permission (respects entity scope).

**Request body:**

```json
{
  "permission": "contracts.contract.approve",
  "resource_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response 200:**

```json
{
  "allowed": true,
  "reason": "role:contracts.approver"
}
```

---

## Error Responses

| Code | Error | Cause |
|------|-------|-------|
| 404 | `USER_NOT_FOUND` | User ID not in tenant |
| 404 | `ROLE_NOT_FOUND` | Role name not in tenant |
| 409 | `EMAIL_ALREADY_EXISTS` | Duplicate email within tenant |
| 409 | `ROLE_NAME_TAKEN` | Role name already used |
| 409 | `ROLE_IN_USE` | Delete role with active assignments |
| 422 | `INVALID_ROLE_NAME` | Name doesn't match `{prefix}.{verb}` pattern |
| 422 | `UNKNOWN_PERMISSION` | Permission name not in enabled modules |
| 403 | `CANNOT_MODIFY_SYSTEM_ROLE` | Attempt to modify/delete system role |
