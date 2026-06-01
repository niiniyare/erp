---
title: IAM API
portal: 8 — API Reference
section: 08-api-reference
audience: [backend-engineer, frontend-engineer, integrator]
related:
  - "[API Reference Overview](01-api-reference-overview.md)"
  - "[Authentication](02-authentication.md)"
  - "[Authorization Model](../03-platform-architecture/02-iam/03-authorization-model.md)"
---

# IAM API

## Users

### GET /api/v1/iam/users

List users in the tenant.

**Permission required**: `iam.user.list`

#### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | — |
| `page_size` | int | — |
| `role` | string | Filter by role name |
| `entity_id` | uuid | Filter by entity scope |
| `active` | bool | Default: true |

#### Response 200

```json
{
  "data": [
    {
      "id": "...",
      "email": "jane@example.com",
      "name": "Jane Smith",
      "is_active": true,
      "roles": ["contracts.editor", "finance.viewer"],
      "entity_scope": {
        "type": "entity",
        "entity_id": "..."
      },
      "created_at": "2025-01-01T00:00:00Z",
      "last_login_at": "2025-01-15T08:30:00Z"
    }
  ],
  "pagination": { "page": 1, "page_size": 20, "total": 8, "total_pages": 1 }
}
```

---

### POST /api/v1/iam/users

Create a new user in the tenant.

**Permission required**: `iam.user.create`

```json
{
  "email": "john@example.com",
  "name": "John Doe",
  "password": "TemporaryPass123!",
  "roles": ["contracts.viewer"],
  "entity_scope": {
    "type": "all"
  }
}
```

**Response 201** — returns created user (password not returned).

**Response 409** — email already exists in tenant.

---

### GET /api/v1/iam/users/:id

Get user by ID.

---

### PUT /api/v1/iam/users/:id

Update user. Accepts same fields as POST. Include `version`.

**Response 422** — cannot change own roles (must be done by another admin).

---

### DELETE /api/v1/iam/users/:id

Soft-deactivate user. All sessions revoked immediately.

```json
{ "version": 3 }
```

---

### POST /api/v1/iam/users/:id/reset-password

Trigger password reset email.

**Permission required**: `iam.user.reset_password`

---

## Roles

### GET /api/v1/iam/roles

List all roles defined for the tenant.

#### Response 200

```json
{
  "data": [
    {
      "id": "...",
      "name": "contracts.editor",
      "display_name": "Contracts Editor",
      "description": "Can create and edit contracts",
      "permissions": [
        "contracts.contract.create",
        "contracts.contract.update",
        "contracts.contract.read",
        "contracts.contract.submit"
      ],
      "user_count": 3
    }
  ]
}
```

---

### POST /api/v1/iam/roles

Create custom role.

**Permission required**: `iam.role.create`

```json
{
  "name": "contracts.approver",
  "display_name": "Contracts Approver",
  "description": "Can approve submitted contracts",
  "permissions": [
    "contracts.contract.read",
    "contracts.contract.approve"
  ]
}
```

---

### PUT /api/v1/iam/roles/:id

Update role permissions. Include `version`.

**Response 422** — cannot modify system-defined roles.

---

### DELETE /api/v1/iam/roles/:id

Delete custom role.

**Response 422** — role still assigned to users. Unassign first.

---

## Permissions

### GET /api/v1/iam/permissions

List all available permissions in the system.

#### Response 200

```json
{
  "data": [
    {
      "key": "contracts.contract.create",
      "module": "contracts",
      "resource": "contract",
      "action": "create",
      "description": "Create new contracts"
    },
    {
      "key": "contracts.contract.approve",
      "module": "contracts",
      "resource": "contract",
      "action": "approve",
      "description": "Approve submitted contracts"
    }
  ]
}
```

---

## User Role Assignments

### POST /api/v1/iam/users/:id/roles

Assign roles to user.

**Permission required**: `iam.user.assign_roles`

```json
{
  "roles": ["contracts.editor", "finance.viewer"],
  "version": 2
}
```

Replaces existing role assignments entirely.

---

### GET /api/v1/iam/users/:id/permissions

Get effective permissions for user (union of all role permissions).

#### Response 200

```json
{
  "user_id": "...",
  "permissions": [
    "contracts.contract.create",
    "contracts.contract.read",
    "contracts.contract.update",
    "contracts.contract.submit",
    "finance.account.read"
  ]
}
```

---

## Sessions

### GET /api/v1/iam/users/:id/sessions

List active sessions for user.

**Permission required**: `iam.user.manage_sessions`

#### Response 200

```json
{
  "data": [
    {
      "token_prefix": "550e8400",
      "created_at": "2025-01-15T08:30:00Z",
      "expires_at": "2025-01-16T08:30:00Z",
      "last_active_at": "2025-01-15T14:22:00Z",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0..."
    }
  ]
}
```

---

### DELETE /api/v1/iam/users/:id/sessions

Revoke all sessions for user. Forces re-login on all devices.

**Response 204** — sessions revoked.
