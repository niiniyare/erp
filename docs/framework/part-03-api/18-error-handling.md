---
title: "Chapter 18: Error Handling"
part: "Part III — The API Layer"
chapter: 18
section: "18-error-handling"
related:
  - "[Chapter 17: REST API Conventions](17-rest-api-conventions.md)"
  - "[Chapter 7: The EntityRecord Lifecycle](../part-02-entity-system/07-entity-record-lifecycle.md)"
---

# Chapter 18: Error Handling

Consistent, structured error handling is one of the most visible aspects of an API's quality. Awo defines a typed error hierarchy, maps error types to HTTP status codes, and localises error messages — all configured to work seamlessly with amis form field error binding.

---

## 18.1. Error Type Hierarchy

```
AwoError (interface)
├── ValidationError      — field-level, user-correctable
├── NotFoundError        — record or resource does not exist
├── PermissionError      — RBAC denied the operation
├── ConflictError        — uniqueness violation, optimistic lock failure
├── BusinessRuleError    — hook rejected the operation with a domain reason
├── WorkflowError        — Temporal activity/workflow failure
└── InternalError        — unhandled, logged, not exposed
```

### 18.1.1. `ValidationError` — Field-Level, User-Correctable

```go
type ValidationError struct {
    Fields []FieldError `json:"errors"`
}

type FieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Code    string `json:"code,omitempty"`
}

func NewValidationError(fields ...FieldError) *ValidationError {
    return &ValidationError{Fields: fields}
}
```

Returned when: field type mismatch, required field missing, value out of range, pattern mismatch, uniqueness violation detected before DB.

### 18.1.2. `NotFoundError` — Entity or Record Does Not Exist

```go
type NotFoundError struct {
    EntityName string    `json:"entity"`
    ID         uuid.UUID `json:"id,omitempty"`
}

func NotFound(entity string, id uuid.UUID) *NotFoundError {
    return &NotFoundError{EntityName: entity, ID: id}
}
```

Also returned when a privacy policy hides a record — do not expose the fact that a record exists but is filtered.

### 18.1.3. `PermissionError` — Role Lacks Required Permission

```go
type PermissionError struct {
    Entity string `json:"entity"`
    Action string `json:"action"`
}
```

Returned by the `RequirePermission` middleware (403). Note: 401 is returned when the user is not authenticated at all (no session). 403 is returned when they are authenticated but not authorised.

### 18.1.4. `ConflictError` — Uniqueness Violation, Optimistic Lock

```go
type ConflictError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Field   string `json:"field,omitempty"`  // set for uniqueness violations
}
```

### 18.1.5. `BusinessRuleError` — Hook Rejected the Operation

```go
type BusinessRuleError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func NewBusinessError(code, format string, args ...interface{}) *BusinessRuleError {
    return &BusinessRuleError{
        Code:    code,
        Message: fmt.Sprintf(format, args...),
    }
}
```

Codes are stable machine-readable identifiers: `CREDIT_LIMIT_EXCEEDED`, `POSTED_INVOICE_IMMUTABLE`, `PERIOD_CLOSED`. Messages are user-facing and localised.

### 18.1.6. `WorkflowError` — Temporal Failure

```go
type WorkflowError struct {
    WorkflowID string `json:"workflow_id"`
    Message    string `json:"message"`
    Retryable  bool   `json:"retryable"`
}
```

When a Temporal activity fails after all retries exhausted, the error is surfaced here. Retryable errors (transient failures) return 202 + a polling URL. Non-retryable errors return 422 with the failure reason.

### 18.1.7. `InternalError` — Unhandled

Any error not matching the above types. Logged with full context, stack trace, and request ID. Response contains only:

```json
{
  "status": 500,
  "message": "an internal error occurred",
  "request_id": "req-uuid"
}
```

Never expose stack traces, SQL errors, or internal system details to API callers.

---

## 18.2. HTTP Status Code Mapping

```go
func MapErrorToHTTP(err error) (int, interface{}) {
    var ve *ValidationError
    var nfe *NotFoundError
    var pe *PermissionError
    var ce *ConflictError
    var bre *BusinessRuleError
    var we *WorkflowError

    switch {
    case errors.As(err, &ve):
        return 422, ve
    case errors.As(err, &nfe):
        return 404, nfe
    case errors.As(err, &pe):
        return 403, pe
    case errors.As(err, &ce):
        return 409, ce
    case errors.As(err, &bre):
        return 422, bre
    case errors.As(err, &we):
        if we.Retryable {
            return 202, we
        }
        return 422, we
    default:
        // Log the unhandled error
        slog.Error("unhandled error", "error", err, "stack", debug.Stack())
        return 500, InternalError()
    }
}
```

Always use `errors.As` for unwrapping — error chains (wrapped errors) must be traversed correctly. A `*BusinessRuleError` returned from three levels of nested function calls must still be correctly identified.

---

## 18.3. Field-Level Validation Errors for amis

### 18.3.1. Error Shape Expected by amis

amis form validation binds errors by the field's `name` property. The API must return errors in this shape:

```json
{
  "status": 422,
  "errors": [
    { "field": "email", "message": "must be a valid email address" },
    { "field": "custom_fields.kra_pin", "message": "must match format A000000000X" }
  ]
}
```

The amis form reads `$.errors` and maps each error's `field` to the form field with the matching `name`.

### 18.3.2. Mapping DB Constraint Errors to Field Errors

PostgreSQL constraint violations are caught in the repository layer and mapped to typed errors:

```go
func parseDBError(err error, op string) error {
    var pgErr *pgconn.PgError
    if !errors.As(err, &pgErr) {
        return err
    }

    switch pgErr.Code {
    case "23505": // unique_violation
        field := extractFieldFromConstraintName(pgErr.ConstraintName)
        return &ConflictError{
            Code:    "UNIQUE_VIOLATION",
            Message: fmt.Sprintf("a record with this %s already exists", field),
            Field:   field,
        }
    case "23503": // foreign_key_violation
        return &BusinessRuleError{
            Code:    "REFERENCE_NOT_FOUND",
            Message: "referenced record does not exist",
        }
    case "23514": // check_violation
        return &ValidationError{Fields: []FieldError{
            {Field: extractFieldFromConstraint(pgErr.ConstraintName),
             Message: "value violates a database constraint"},
        }}
    }
    return err
}
```

### 18.3.3. Top-Level (Non-Field) Errors

`BusinessRuleError` and `WorkflowError` do not have a specific field. amis renders these as a toast notification at the top of the page:

```json
{
  "status": 422,
  "code": "CREDIT_LIMIT_EXCEEDED",
  "message": "Invoice total would exceed customer credit limit of KES 45,000.00"
}
```

amis detects the absence of `errors[]` and uses `message` as the toast content.

---

## 18.4. Localisation of Error Messages

### 18.4.1. Error Codes as the Stable Key

Error messages are localised; error codes are not. Code `CREDIT_LIMIT_EXCEEDED` is the stable key used by API integrators to handle errors programmatically. The message is for human consumption and changes when translations are updated.

### 18.4.2. Message Catalogue via Go `embed`

```go
//go:embed locales/*.json
var localesFS embed.FS

type MessageCatalogue struct {
    messages map[string]map[string]string // locale → code → message
}

func (c *MessageCatalogue) Get(locale, code string, args ...interface{}) string {
    if msgs, ok := c.messages[locale]; ok {
        if msg, ok := msgs[code]; ok {
            return fmt.Sprintf(msg, args...)
        }
    }
    // Fallback to English
    return c.messages["en"][code]
}
```

```json
// locales/sw.json (Swahili)
{
  "CREDIT_LIMIT_EXCEEDED": "Jumla ya ankara inazidi kikomo cha mkopo cha KES %s",
  "REQUIRED_FIELD": "Sehemu hii inahitajika",
  "INVALID_EMAIL": "Anwani ya barua pepe si sahihi",
  "ACCOUNT_LOCKED": "Akaunti imefungwa. Jaribu tena baada ya dakika 15."
}
```

### 18.4.3. Swahili Error Messages for the Kenyan Market

Swahili is the second official language of Kenya and is the primary language of many SME operators. Error messages in Swahili reduce support calls and improve user confidence.

The locale is resolved from:
1. The request's `Accept-Language` header
2. The tenant's configured default locale
3. The user's profile locale preference
4. Fallback: English

### 18.4.4. Fallback Chain

```
Accept-Language: sw-KE, sw, en
→ Try: sw-KE → sw → en → (hardcoded English literal)
```

If a message code exists in `sw` but not in `sw-KE`, the `sw` translation is used. If it exists in neither, English is used. If not even in English, the code itself is returned as the message (signals a missing translation).
