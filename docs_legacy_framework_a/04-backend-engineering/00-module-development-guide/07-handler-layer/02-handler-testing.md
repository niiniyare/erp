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
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Handler Layer Overview](01-handler-overview.md)"
  - "[Testing Overview](../18-testing/01-testing-overview.md)"
  - "[Middleware Testing](../16-middleware-chain/07-middleware-testing.md)"
---

# Handler Testing

## Testing with `app.Test()`

Fiber's `app.Test()` exercises the full HTTP stack without starting a real network server:

```go
func TestContractHandler_Create_Success(t *testing.T) {
    app := setupTestApp(t)

    body := `{
        "contract_number": "CONT-001",
        "title": "Test Contract",
        "contract_type": "service",
        "total_value": "10000",
        "currency": "USD",
        "start_date": "2025-01-01",
        "end_date": "2025-12-31"
    }`

    req := httptest.NewRequest("POST", "/api/v1/contracts", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    resp, err := app.Test(req)
    require.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
    assert.NotEmpty(t, resp.Header.Get("Location"))

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    assert.Equal(t, "CONT-001", result["contract_number"])
}
```

## Test App Setup

```go
func setupTestApp(t *testing.T) *fiber.App {
    t.Helper()

    mockSvc := &mockContractService{
        createResult: &domain.Contract{
            ID:             uuid.New(),
            ContractNumber: "CONT-001",
            Title:          "Test Contract",
            Status:         domain.StatusDraft,
            TotalValue:     decimal.NewFromFloat(10000),
            Version:        1,
        },
    }

    app := fiber.New(fiber.Config{
        ErrorHandler: customErrorHandler,
    })

    // Inject test session — bypasses auth middleware
    sess := testSession()
    app.Use(func(c *fiber.Ctx) error {
        c.Locals("resolved_session", sess)
        return c.Next()
    })

    h := NewContractHandler(mockSvc)
    api := app.Group("/api/v1")
    api.Post("/contracts", h.Create)
    api.Get("/contracts/:id", h.GetByID)
    api.Post("/contracts/:id/submit", h.Submit)

    return app
}
```

## Error Response Tests

```go
func TestContractHandler_GetByID_NotFound(t *testing.T) {
    app := setupTestAppWithService(t, &mockContractService{
        getErr: domain.ErrContractNotFound,
    })

    req := httptest.NewRequest("GET", "/api/v1/contracts/"+uuid.New().String(), nil)
    resp, err := app.Test(req)
    require.NoError(t, err)
    assert.Equal(t, http.StatusNotFound, resp.StatusCode)

    var body map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&body)
    errObj := body["error"].(map[string]interface{})
    assert.Equal(t, "NOT_FOUND", errObj["code"])
}

func TestContractHandler_Create_VersionConflict(t *testing.T) {
    app := setupTestAppWithService(t, &mockContractService{
        updateErr: domain.ErrVersionConflict,
    })

    body := `{"title": "Updated", "version": 1}`
    req := httptest.NewRequest("PUT", "/api/v1/contracts/"+uuid.New().String(), strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    resp, _ := app.Test(req)
    assert.Equal(t, http.StatusConflict, resp.StatusCode)
}
```

## Validation Tests

```go
func TestContractHandler_Create_MissingFields(t *testing.T) {
    app := setupTestApp(t)

    cases := []struct {
        name string
        body string
        expectedField string
    }{
        {"missing contract_number", `{"title":"Test","total_value":"1000","currency":"USD","start_date":"2025-01-01","end_date":"2025-12-31"}`, "contract_number"},
        {"missing title", `{"contract_number":"C-1","total_value":"1000","currency":"USD","start_date":"2025-01-01","end_date":"2025-12-31"}`, "title"},
        {"invalid currency", `{"contract_number":"C-1","title":"T","total_value":"1000","currency":"USDD","start_date":"2025-01-01","end_date":"2025-12-31"}`, "currency"},
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/api/v1/contracts", strings.NewReader(tc.body))
            req.Header.Set("Content-Type", "application/json")

            resp, _ := app.Test(req)
            assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

            var body map[string]interface{}
            json.NewDecoder(resp.Body).Decode(&body)
            errObj := body["error"].(map[string]interface{})
            assert.Equal(t, "VALIDATION_ERROR", errObj["code"])

            details := errObj["details"].([]interface{})
            fields := make([]string, len(details))
            for i, d := range details {
                fields[i] = d.(map[string]interface{})["field"].(string)
            }
            assert.Contains(t, fields, tc.expectedField)
        })
    }
}
```

## UUID Parsing Test

```go
func TestContractHandler_GetByID_InvalidUUID(t *testing.T) {
    app := setupTestApp(t)

    req := httptest.NewRequest("GET", "/api/v1/contracts/not-a-uuid", nil)
    resp, _ := app.Test(req)
    assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
```

## Service Mock

```go
type mockContractService struct {
    createResult *domain.Contract
    createErr    error
    getResult    *domain.Contract
    getErr       error
    submitErr    error
    updateErr    error
}

func (m *mockContractService) Create(_ context.Context, _ iam.ResolvedSession, _ CreateParams) (*domain.Contract, error) {
    return m.createResult, m.createErr
}
func (m *mockContractService) GetByID(_ context.Context, _ iam.ResolvedSession, _ uuid.UUID) (*domain.Contract, error) {
    return m.getResult, m.getErr
}
func (m *mockContractService) Submit(_ context.Context, _ iam.ResolvedSession, _ uuid.UUID, _ int) error {
    return m.submitErr
}
// ... implement remaining interface methods
```
