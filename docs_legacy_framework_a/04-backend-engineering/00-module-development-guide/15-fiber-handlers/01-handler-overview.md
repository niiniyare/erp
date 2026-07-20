> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Fiber Handler Overview
portal: 4 — Backend Engineering
section: 00-module-development-guide/15-fiber-handlers
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-handler-struct.md
    title: Handler Struct
  - path: ./03-request-dtos.md
    title: Request DTOs
  - path: ./04-response-dtos.md
    title: Response DTOs
---

# Fiber Handler Overview

The handler layer bridges HTTP and the service layer. It parses requests, validates input, extracts the session, calls the service, maps responses, and maps errors to HTTP responses.

## Handler Package Structure

```
internal/api/handlers/contracts/
├── handler.go    # Handler struct + constructor + all handler methods
├── routes.go     # RegisterRoutes function
├── request.go    # Request DTOs with validate tags
├── response.go   # Response DTOs + MapXxxToResponse functions
└── errors.go     # mapError(err error) error
```

## Handler Responsibilities

| Responsibility | Handler does |
|---------------|-------------|
| Parse request body | `c.BodyParser(&req)` |
| Parse URL params | `c.Params("id")` |
| Parse query params | `c.QueryInt("page_size", 20)` |
| Validate input | `validate.Struct(req)` |
| Extract session | `c.Locals(iam.LocalsKeySession)` |
| Call service | `h.service.Create(ctx, ...)` |
| Map domain → response DTO | `MapContractToResponse(contract)` |
| Map errors to HTTP | `mapError(err)` |
| Return response | `c.Status(201).JSON(resp)` |

## Handler Does NOT Do

| Handler does NOT do | Where it belongs |
|--------------------|-----------------|
| Business logic | Service layer |
| Permission checks (beyond middleware) | Service layer |
| Database queries | Repository layer |
| State machine checks | Domain entity |
| Audit recording | Service layer |
| Event publishing | Service layer |

## Handler → Service Data Flow

```
HTTP Request
    │
    ▼
Handler parses body + params
    │
    ▼
Handler validates request DTO
    │
    ▼
Handler extracts session (sess.TenantID, sess.UserID, etc.)
    │
    ▼
Handler calls service with typed Go structs (no HTTP types)
    │
    ▼
Service returns domain entity or error
    │
    ├── Success → handler maps to response DTO → JSON
    └── Error   → mapError(err) → fiber.Error → JSON
```

## HTTP Status Codes

| Operation | Success | Created by |
|-----------|---------|------------|
| POST (create) | `201 Created` | Handler |
| GET (detail) | `200 OK` | Handler |
| GET (list) | `200 OK` | Handler |
| PUT/PATCH (update) | `200 OK` | Handler |
| POST (status change) | `200 OK` | Handler |
| DELETE | `204 No Content` | Handler |

## Validation Library

AwoERP uses `github.com/go-playground/validator/v10` for request struct validation. The validator instance is typically shared application-wide and injected into the handler struct or accessed via a package-level singleton.

```go
import "github.com/go-playground/validator/v10"

var validate = validator.New()

// In handler
if err := validate.Struct(req); err != nil {
	return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
}
```
