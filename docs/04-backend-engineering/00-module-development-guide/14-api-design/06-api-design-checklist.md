---
title: API Design Checklist
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[API Design Overview](01-api-design-overview.md)"
  - "[URL Conventions](02-url-conventions.md)"
  - "[Request/Response](03-request-response.md)"
  - "[Pagination](04-pagination.md)"
---

# API Design Checklist

Use before opening a PR for new API endpoints.

## URL Design

- [ ] No tenant ID in URL — tenant comes from session
- [ ] Resource names are lowercase plural nouns (`/contracts`, `/finance/accounts`)
- [ ] IDs are UUID path params (`/:id`)
- [ ] Action endpoints use `POST /:id/{action}` (`/contracts/:id/submit`)
- [ ] No verbs in resource URLs (`/contracts`, not `/getContracts`)
- [ ] Nested resources used only for tight ownership (`/contracts/:id/lines`)

## Request Format

- [ ] `Content-Type: application/json` for all write endpoints
- [ ] UUID parameters parsed with proper error (400 on invalid format)
- [ ] Request body validated before calling service
- [ ] Required fields enforced with 422 + field-level detail
- [ ] Dates in `YYYY-MM-DD` format
- [ ] Monetary values as strings in JSON (not float)
- [ ] `version` included in all update/delete/action request bodies

## Response Format

- [ ] 200 for success, 201 for creates (with `Location` header), 204 for deletes
- [ ] Error envelope: `{"error": {"code": "...", "message": "...", "details": [...]}}`
- [ ] List endpoints return `{"data": [...], "pagination": {...}}`
- [ ] Empty list returns `"data": []` (not null)
- [ ] `total_pages` calculated with `math.Ceil`
- [ ] Timestamps in RFC3339 format (`2025-01-15T10:30:00Z`)
- [ ] Monetary values as decimal strings (not float64)
- [ ] Deleted records return 404 (not 200 with `deleted_at` set)

## Status Codes

- [ ] 400 — malformed JSON, invalid UUID format
- [ ] 401 — missing or expired token
- [ ] 403 — authenticated but not authorized
- [ ] 404 — resource not found (or belongs to another tenant — same response)
- [ ] 409 — optimistic lock conflict or duplicate key
- [ ] 422 — business rule violation or validation error
- [ ] 429 — rate limit exceeded
- [ ] 500 — unexpected server error (no internal details leaked)

## Security

- [ ] All endpoints protected by `Authenticate` middleware
- [ ] All endpoints protected by `Authorize` middleware with specific permission string
- [ ] No tenant ID accepted from URL or request body
- [ ] No sensitive data in error messages (no SQL, no stack traces)
- [ ] Rate limit appropriate for endpoint type (auth: 10/min, standard: 60/min, bulk: 5/min)

## Pagination

- [ ] List endpoints accept `page` and `page_size` query params
- [ ] `page_size` max enforced at 100
- [ ] Default `page_size` is 20
- [ ] `total` returned in pagination object
- [ ] `total_pages` returned and calculated correctly

## Documentation

- [ ] New endpoint documented in Portal 8 (API Reference)
- [ ] Request body with all fields documented
- [ ] All possible response codes documented
- [ ] Example request and response JSON included
- [ ] Related endpoints cross-linked

## Breaking Change Assessment

Does this change:
- Remove a field from the response? → **Breaking**
- Change a field type (e.g., int → string)? → **Breaking**
- Remove or rename a query parameter? → **Breaking**
- Change a status code for an existing response? → **Breaking**
- Add a required request field? → **Breaking**

If any box is checked → requires API versioning discussion before merge.
