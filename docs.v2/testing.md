# Testing Guide

## Overview
This project uses Go's built-in testing framework with testify for assertions and mocking.

## Test Structure
```
├── test/
│   └── testutil/          # Testing utilities
├── internal/domain/
│   └── */                 # Domain service tests (*_test.go)
└── db/sqlc/              # Database layer tests
```

## Running Tests

### All Tests
```bash
make test
```

### Database Tests Only
```bash
make testdb
```

### HTML Coverage Report
```bash
make test/html
```

## Test Types

### Unit Tests
- Location: `internal/domain/*/service_test.go`
- Use mocks for dependencies
- Test business logic in isolation

### Integration Tests
- Location: `db/sqlc/*_test.go`
- Test database interactions
- Use test database

## Test Utilities

### Database Testing
```go
import "github.com/niiniyare/erp/test/testutil"

func TestWithDB(t *testing.T) {
    tdb := testutil.NewTestDB(t)
    defer tdb.Close()
    
    // Use tdb.Pool or tdb.DB for testing
}
```

### Mocking
- Use testify/mock for repository mocks
- See `internal/domain/tenant/service_test.go` for examples

## Environment Variables
- `TEST_DATABASE_URL`: Test database connection string
- Default: `postgresql://admin:admin@localhost:5432/ledger_test?sslmode=disable`