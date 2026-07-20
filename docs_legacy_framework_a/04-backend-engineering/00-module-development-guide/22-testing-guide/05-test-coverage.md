> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Test Coverage Standards
portal: 4 — Backend Engineering
section: 00-module-development-guide/22-testing-guide
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-testing-overview.md
    title: Testing Overview
  - path: ./04-service-tests.md
    title: Service Tests
---

# Test Coverage Standards

Coverage requirements by module layer.

## Required Test Cases Per Layer

### Domain

- [ ] `CanTransitionTo`: every row in the transition table (valid + invalid)
- [ ] Every value object constructor: valid input, each invalid input type
- [ ] Every entity helper method (`IsDeleted`, `IsEditable`, `IsActive`)
- [ ] Each sentinel error is a distinct `errors.New` (not aliased)

### Repository (Integration — Real DB)

- [ ] `Create`: success → returns persisted entity
- [ ] `Create`: duplicate unique key → `ErrContractAlreadyExists`
- [ ] `GetByID`: success
- [ ] `GetByID`: not found → `ErrContractNotFound`
- [ ] `GetByID`: cross-tenant → `ErrContractNotFound` (RLS check)
- [ ] `Update`: success → version incremented
- [ ] `Update`: stale version → `ErrContractConflict`
- [ ] `Delete`: success → `deleted_at` set
- [ ] `Delete`: already deleted → `ErrContractNotFound`
- [ ] `List`: returns empty slice (not nil) when no results
- [ ] `List`: filters work (status, entity_id)
- [ ] `Count`: matches `List` result count

### Service (Unit — Mocked Dependencies)

- [ ] Every method: authorization check runs first (deny path → `ErrForbidden`)
- [ ] Every method: authz error propagates as error (not 403)
- [ ] Every status transition: invalid transition → `ErrContractInvalidTransition`
- [ ] Every status transition: `UpdateStatus` not called on invalid transition
- [ ] `Update`: `ErrContractNotEditable` when status != draft
- [ ] `Update`: `ErrContractConflict` propagates from repo
- [ ] `Create`: `ErrContractAlreadyExists` when number exists
- [ ] `Create`: repository not called after auth denial
- [ ] `Create`: audit event recorded after success
- [ ] `Create`: event published after success

### Handler (Integration — Mock Service, Real Fiber)

- [ ] `POST /contracts`: 201 with valid body
- [ ] `POST /contracts`: 400 with invalid JSON
- [ ] `POST /contracts`: 422 with missing required fields
- [ ] `POST /contracts`: 401 without session
- [ ] `POST /contracts`: 403 when service returns `ErrForbidden`
- [ ] `GET /contracts/:id`: 200 with valid ID
- [ ] `GET /contracts/:id`: 400 with non-UUID ID
- [ ] `GET /contracts/:id`: 404 when not found
- [ ] `PUT /contracts/:id`: 409 on conflict
- [ ] `POST /contracts/:id/submit`: 422 on invalid transition
- [ ] `DELETE /contracts/:id`: 204 on success
- [ ] `DELETE /contracts/:id`: 404 on not found

## Coverage Command

```bash
make test-coverage
# opens HTML report at coverage/index.html
```

```bash
# Check coverage threshold (fails if below 80%)
go test -coverprofile=coverage.out ./internal/core/contracts/...
go tool cover -func=coverage.out | grep total
```

## What Coverage Doesn't Measure

Coverage measures line execution, not correctness. A 100% coverage score with weak assertions is worthless. Each test must have meaningful assertions:

```go
// WEAK — covered but asserts nothing useful
assert.NotNil(t, contract)

// STRONG — asserts the important properties
assert.Equal(t, domain.ContractStatusDraft, contract.Status)
assert.Equal(t, 1, contract.Version)
assert.NotEqual(t, uuid.Nil, contract.ID)
assert.Equal(t, params.ContractNumber, contract.ContractNumber)
```
