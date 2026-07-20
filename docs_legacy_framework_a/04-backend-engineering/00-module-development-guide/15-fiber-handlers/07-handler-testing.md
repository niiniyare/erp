> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Handler Testing
portal: 4 — Backend Engineering
section: 00-module-development-guide/15-fiber-handlers
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-handler-struct.md
    title: Handler Struct
  - path: ../22-testing-guide/01-testing-overview.md
    title: Testing Overview
---

# Handler Testing

Handler tests use a real Fiber app with a mock service. They test the full HTTP path: parsing, validation, authentication, authorization, response mapping, and error mapping.

## Test Setup

```go
// internal/api/handlers/contracts/handler_test.go
package contracts_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsHandler "awo.so/internal/api/handlers/contracts"
	iam "awo.so/internal/core/iam"
)

func setupTestApp(svc service.ContractService) *fiber.App {
	app := fiber.New()
	api := app.Group("/api/v1")

	authConfig := &middlewarePkg.AuthConfig{
		// Test auth config with mock authz service
	}

	contractsHandler.RegisterRoutes(api, &handlers.Dependencies{
		ContractService: svc,
		AuthConfig:      authConfig,
		Logger:          testLogger(),
		Tracer:          noopTracer(),
	})

	// Inject test session via middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(iam.LocalsKeySession, testSession())
		return c.Next()
	})

	return app
}

func testSession() *iam.ResolvedSession {
	return &iam.ResolvedSession{
		TenantID:    testTenantID,
		UserID:      testUserID,
		EntityScope: iam.EntityScope{Type: iam.EntityScopeAll},
	}
}
```

## Test: Successful Create

```go
func TestHandler_Create_Success(t *testing.T) {
	mockSvc := &mockContractService{}
	mockSvc.On("Create", mock.Anything, mock.Anything).
		Return(&domain.Contract{
			ID:             testContractID,
			ContractNumber: "CONT-2025-0001",
			Status:         domain.ContractStatusDraft,
			Version:        1,
			// ...
		}, nil)

	app := setupTestApp(mockSvc)

	body, _ := json.Marshal(map[string]interface{}{
		"contract_number": "CONT-2025-0001",
		"title":           "Test Contract",
		"vendor_id":       uuid.New().String(),
		"contract_type":   "service",
		"start_date":      "2025-01-01",
		"end_date":        "2025-12-31",
		"currency":        "USD",
	})

	req := httptest.NewRequest("POST", "/api/v1/contracts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "CONT-2025-0001", result["contract_number"])
	assert.Equal(t, "draft", result["status"])
	assert.NotEmpty(t, result["version"])
}
```

## Test: Validation Failure

```go
func TestHandler_Create_ValidationFailure(t *testing.T) {
	app := setupTestApp(&mockContractService{})

	body, _ := json.Marshal(map[string]interface{}{
		"title": "",  // required field empty
	})

	req := httptest.NewRequest("POST", "/api/v1/contracts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 422, resp.StatusCode)
}
```

## Test: Not Found

```go
func TestHandler_GetByID_NotFound(t *testing.T) {
	mockSvc := &mockContractService{}
	mockSvc.On("GetByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, domain.ErrContractNotFound)

	app := setupTestApp(mockSvc)

	req := httptest.NewRequest("GET", "/api/v1/contracts/"+uuid.New().String(), nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 404, resp.StatusCode)
}
```

## Test: Conflict (version mismatch)

```go
func TestHandler_Update_Conflict(t *testing.T) {
	mockSvc := &mockContractService{}
	mockSvc.On("Update", mock.Anything, mock.Anything).
		Return(nil, domain.ErrContractConflict)

	app := setupTestApp(mockSvc)

	body, _ := json.Marshal(UpdateContractRequest{
		Title:   "Updated Title",
		Version: 1,  // stale version
		// ...
	})
	req := httptest.NewRequest("PUT", "/api/v1/contracts/"+testContractID.String(),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 409, resp.StatusCode)
}
```

## Test Coverage Requirements

Every handler method must have tests for:
- `200`/`201`/`204` success path
- `400` bad request (invalid UUID, malformed body)
- `422` validation failure (missing required field)
- `404` not found
- `409` conflict (where applicable)
- `401` missing session (inject nil session)
- `403` forbidden (inject mock authz that denies)
