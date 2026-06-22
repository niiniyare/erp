---
title: "Chapter 19: API Versioning"
part: "Part III — The API Layer"
chapter: 19
section: "19-api-versioning"
related:
  - "[Chapter 17: REST API Conventions](17-rest-api-conventions.md)"
  - "[Chapter 18: Error Handling](18-error-handling.md)"
---

# Chapter 19: API Versioning

API versioning is a commitment to your API consumers. Breaking their integrations without warning destroys trust and causes real business harm. Awo's versioning strategy prioritises stability for existing integrations while enabling the framework to evolve.

---

## 19.1. Versioning Strategy

### 19.1.1. URI Prefix Versioning — `/api/v1/`, `/api/v2/`

```
/api/v1/invoices          ← current stable version
/api/v2/invoices          ← new major version (when v1 breaks)
/api/v1/invoices/preview  ← preview endpoint for future behaviour (opt-in)
```

URI versioning is explicit and cache-friendly. Header versioning (`API-Version: 2`) is harder to cache and debug.

### 19.1.2. Version Lifetime Policy

- `v1` is supported for at minimum **24 months** after `v2` is released
- Deprecation notice period: **6 months** minimum
- No version is removed without operator notification via email, changelog, and in-app notice

### 19.1.3. What Constitutes a Breaking Change

**Breaking changes require a new major version (`v2`):**
- Removing a response field
- Changing a field's type (e.g. `string` → `int`)
- Changing an error code that integrations may handle programmatically
- Changing authentication requirements on an existing endpoint
- Removing an endpoint
- Changing pagination cursor format

**Non-breaking changes (no version bump required):**
- Adding a new optional field to a response
- Adding a new endpoint
- Adding a new optional request field
- Adding new error codes (existing codes unchanged)
- Improving error messages (as long as codes are unchanged)
- Performance improvements with identical semantics

---

## 19.2. Deprecation Lifecycle

### 19.2.1. `Deprecation` and `Sunset` Response Headers

When an endpoint or field is deprecated, responses include:

```
Deprecation: Mon, 04 Jul 2025 00:00:00 GMT
Sunset: Sat, 04 Jul 2026 00:00:00 GMT
Link: <https://docs.awo.so/api/migration/v1-to-v2>; rel="successor-version"
```

- `Deprecation`: when the deprecation was announced
- `Sunset`: when the endpoint will be removed (callers have until this date)
- `Link`: migration guide

### 19.2.2. Deprecation Notice Period

Minimum **6 months** between `Deprecation` header appearing and `Sunset` date. For endpoints used by many integrations (detected via usage metrics), extend to 12 months.

### 19.2.3. Communication Channels

1. `Deprecation`/`Sunset` headers on every response (automated, always present)
2. Changelog entry on the public developer portal
3. Email to all tenant admin contacts with active API usage
4. In-app banner for tenants using deprecated endpoints (SDUI layer shows it in the developer console)
5. Webhook event `api.deprecation.warning` sent 30 days before sunset

### 19.2.4. Removal Process

On the sunset date:
1. Endpoint returns 410 Gone with migration guide URL
2. Retained in the codebase for 30 more days (as 410 handler) before deletion
3. Metrics monitoring for 30 days post-removal to catch missed migrations

---

## 19.3. Backward Compatibility Rules

### 19.3.1. Safe Changes — Can Deploy Without Version Bump

```go
// SAFE: adding an optional field to a response
type InvoiceResponse struct {
    // existing fields...
    EtimsTrackingNumber string `json:"etims_tracking_number,omitempty"` // NEW — omitempty is key
}

// SAFE: adding a new endpoint
app.Get("/api/v1/invoices/:id/etims-status", handler.EtimsStatus)

// SAFE: adding an optional request field
type CreateInvoiceInput struct {
    // existing fields...
    ExternalReference string `json:"external_reference,omitempty"` // NEW
}
```

### 19.3.2. Unsafe Changes — Require Version Bump or Feature Flag

```go
// UNSAFE: removing a field
type InvoiceResponse struct {
    // ContactEmail string — REMOVED — callers may depend on this
}

// UNSAFE: changing a field type
type InvoiceResponse struct {
    TotalAmount float64 `json:"total_amount"` // UNSAFE — was string "45000.0000"
}

// UNSAFE: changing an error code
// Was: CREDIT_LIMIT_EXCEEDED
// Changed to: CREDIT_INSUFFICIENT — breaks integrations handling the old code
```

### 19.3.3. The Contract Test Suite

A contract test suite verifies backward compatibility automatically in CI:

```go
// contracts/v1_invoice_test.go
func TestInvoiceContractV1(t *testing.T) {
    resp := apiTest.GET("/api/v1/invoices/test-fixture-id")

    // Assert all v1 contract fields are present
    resp.JSON().Object().
        ContainsKey("data").
        Value("data").Object().
        ContainsKey("id").
        ContainsKey("invoice_number").
        ContainsKey("total_amount").
        ContainsKey("status").
        ContainsKey("created_at")

    // Assert field types
    totalAmount := resp.JSON().Path("$.data.total_amount").String()
    // total_amount must remain a string (not float) in v1
    assert.Regexp(t, `^\d+\.\d{4}$`, totalAmount)
}
```

Contract tests run in CI against every PR. A PR that removes or changes a contracted field fails CI immediately, before any review.
