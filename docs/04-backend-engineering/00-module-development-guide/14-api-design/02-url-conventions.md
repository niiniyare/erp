---
title: URL Conventions
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[API Design Overview](01-api-design-overview.md)"
  - "[Pagination](04-pagination.md)"
---

# URL Conventions

## Resource Naming

- Plural nouns for collections: `/contracts`, `/contract-lines`
- Kebab-case for multi-word resources: `/contract-lines`, `/access-requests`
- No verbs in URLs: use HTTP method to express the action
- Subresources via nesting (max 2 levels): `/contracts/{id}/lines`

## Contracts Endpoints

```
GET    /api/v1/contracts                     List contracts (paginated)
POST   /api/v1/contracts                     Create contract
GET    /api/v1/contracts/:id                 Get contract by ID
PUT    /api/v1/contracts/:id                 Update contract (full replace)
PATCH  /api/v1/contracts/:id                 Partial update (optional)
DELETE /api/v1/contracts/:id                 Soft delete contract

POST   /api/v1/contracts/:id/submit          Submit for review
POST   /api/v1/contracts/:id/approve         Approve
POST   /api/v1/contracts/:id/reject          Reject → draft
POST   /api/v1/contracts/:id/activate        Activate
POST   /api/v1/contracts/:id/suspend         Suspend
POST   /api/v1/contracts/:id/terminate       Terminate

GET    /api/v1/contracts/:id/lines           List contract lines
POST   /api/v1/contracts/:id/lines           Add contract line
PUT    /api/v1/contracts/:id/lines/:lineId   Update contract line
DELETE /api/v1/contracts/:id/lines/:lineId   Delete contract line

POST   /api/v1/contracts/import              Upload CSV (multipart)
GET    /api/v1/contracts/export              Download CSV
```

## Action Endpoints (Status Transitions)

Status transitions use sub-resource POST endpoints, not query parameters:

```
✓  POST /contracts/:id/submit
✗  POST /contracts/:id?action=submit
✗  PUT  /contracts/:id  { "status": "submitted" }
```

The POST body carries only transition-specific data (version for optimistic lock, optional comment):

```json
{
  "version": 3,
  "comment": "Approved after legal review"
}
```

## ID Format

All resource IDs are UUIDs in path parameters. Reject non-UUID IDs with 400:

```go
id, err := uuid.Parse(c.Params("id"))
if err != nil {
    return fiber.NewError(fiber.StatusBadRequest, "id must be a valid UUID")
}
```

## Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | Page number, 1-based (default: 1) |
| `page_size` | int | Items per page (default: 20, max: 100) |
| `status` | string | Filter by status |
| `entity_id` | uuid | Filter by entity |
| `from` | date | Filter by start_date >= from (YYYY-MM-DD) |
| `to` | date | Filter by end_date <= to (YYYY-MM-DD) |
| `q` | string | Full-text search on contract_number and title |

## Versioning

Breaking changes require a new version prefix:

```
/api/v1/contracts  → stable
/api/v2/contracts  → new version with breaking changes
```

v1 continues to work until a deprecation window (minimum 6 months) expires.
