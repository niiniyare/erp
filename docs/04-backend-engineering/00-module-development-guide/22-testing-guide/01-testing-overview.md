---
title: Testing Overview
portal: 4 — Backend Engineering
section: 00-module-development-guide/22-testing-guide
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-domain-tests.md
    title: Domain Tests
  - path: ./03-repository-tests.md
    title: Repository Tests
  - path: ./04-service-tests.md
    title: Service Tests
---

# Testing Overview

AwoERP uses a layered testing strategy. Each layer tests what it owns — no layer's tests depend on another layer's implementation details.

## Test Layers

| Layer | Test type | What is mocked | Coverage target |
|-------|----------|----------------|----------------|
| Domain | Unit | Nothing | State machine, value objects, entity helpers |
| Repository | Integration | Nothing (real DB) | All repo methods against real PostgreSQL |
| Service | Unit | Repository (mock), platform services (mock) | Business rules, auth checks, state transitions |
| Handler | Integration | Service (mock) | HTTP parsing, validation, auth, error mapping |

**Key constraint: never mock the database in repository tests.** Use a real test database. This is stated in the project's memory and enforced in code review. Mocked DB tests have historically masked migration bugs.

## Test Database Setup

Repository and handler integration tests use a real PostgreSQL test database:

```bash
# Start test DB
make test-db-up

# Apply migrations
make migrate-test-up

# Run tests
make test

# Tear down
make test-db-down
```

The test DB URL is set via the environment variable `TEST_DATABASE_URL`. CI uses Docker Compose to spin up a fresh PostgreSQL instance per test run.

## Test File Conventions

| File | Location | Pattern |
|------|----------|---------|
| Domain unit tests | `internal/core/contracts/domain/` | `*_test.go` |
| Repository integration tests | `internal/core/contracts/repository/` | `*_integration_test.go` |
| Service unit tests | `internal/core/contracts/service/` | `*_test.go` |
| Handler integration tests | `internal/api/handlers/contracts/` | `*_test.go` |

## Test Coverage Requirements

| Layer | Minimum coverage |
|-------|-----------------|
| Domain (state machine) | 100% of transition matrix |
| Repository | All happy paths + all error paths (not found, conflict) |
| Service | 80% of service package |
| Handler | All HTTP status code paths (200/201/400/401/403/404/409/422/500) |

## Test Helpers Location

Shared test helpers live in `internal/testutil/`:

```
internal/testutil/
├── db.go         # test database setup + teardown
├── fixtures.go   # data builders for test contracts, users, tenants
├── session.go    # build a test ResolvedSession
└── assert.go     # custom assertions
```

## Running Tests

```bash
make test           # all tests
make test-unit      # unit tests only (no DB)
make test-int       # integration tests only (requires DB)
make test-coverage  # with coverage report
make lint           # lint + vet
```

Never skip tests to make CI pass. If a test fails, fix the code, not the test.
