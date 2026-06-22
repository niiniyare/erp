---
title: "Chapter 5: Field System"
part: "Part II — The EntityDefinition System"
chapter: 5
section: "05-field-system"
related:
  - "[Chapter 2: The EntityDefinition](../part-01-foundations/02-entity-definition.md)"
  - "[Chapter 7: The EntityRecord Lifecycle](07-entity-record-lifecycle.md)"
  - "[Chapter 10: Custom Fields](10-custom-fields.md)"
---

# Chapter 5: Field System

Fields are the atoms of every EntityDefinition. They describe the shape of your data, drive validation, control persistence column types, and tell the SDUI layer what widget to render. Choosing the right field type at design time saves migrations, bugs, and security incidents later.

---

## 5.1. Field Types Reference

Awo provides four families of field types: scalar (primitive values), structured (constrained compound values), relational (references to other entities), and file (binary object references). Every field declaration maps to an exact PostgreSQL column type — there are no ambiguities or framework-level "auto-detect" heuristics.

### 5.1.1. Scalar Types

Scalar fields map to a single column holding a single primitive value.

#### `Data` — UTF-8 String, Configurable Max Length

The default text field for short structured strings: names, codes, identifiers, addresses. Backed by `varchar(n)` where `n` is set by `MaxLen` (default 140).

```go
field.String("full_name").
    MaxLen(255).
    NotEmpty()
```

Use `Data` when the value has bounded, known length and will be indexed or used in WHERE clauses.

Declare `Searchable()` on a `Data` field to enable substring full-text search. The framework creates a GIN index using `pg_trgm` automatically:

```go
entity.Field("customer_name").
    Type(entity.Data).
    MaxLen(100).
    Searchable()
// → creates: CREATE INDEX ... USING GIN (customer_name gin_trgm_ops)
```

This supports efficient `ILIKE '%kamau%'` queries without a sequential scan. Do not declare `Searchable()` on `LongText` fields — use PostgreSQL `tsvector` for those.

#### `SmallText` — Unindexed, Up to 1024 Characters

For medium prose that does not need a B-tree index: notes, short descriptions, rejection reasons. Stored as `varchar(1024)`. Index creation is blocked by the framework to prevent accidental oversized indexes.

```go
field.String("rejection_note").
    MaxLen(1024).
    Optional()
```

#### `LongText` — Unbounded, Stored as `text`

For free-form content: terms, comments, HTML snippets, email bodies. PostgreSQL `text` column — no length limit enforced at the DB layer. The framework enforces a soft limit via the validator if configured.

```go
field.Text("contract_body").
    Optional()
```

Never add a B-tree index to a `LongText` field. Use full-text search (`tsvector`) if searchability is required.

#### `Int` — 64-bit Signed Integer

Maps to `bigint`. Use for counts, quantities, sequence numbers. Never for monetary values.

```go
field.Int("quantity").
    Min(0).
    Default(1)
```

#### `Float` — 64-bit IEEE 754

Maps to `double precision`. Use only for scientific measurements or percentages where rounding is acceptable. **Never use Float for money.**

```go
field.Float("weight_kg").
    Min(0.0)
```

#### `Currency` — `numeric(20,4)`, Never Floating Point

The only correct type for all monetary amounts. Stored as `numeric(20,4)` (20 total digits, 4 decimal places). The framework formats output as `KES 1,234.5600` in Kenya locale contexts. Arithmetic in Go uses `github.com/shopspring/decimal` to avoid binary floating-point rounding.

```go
field.Float("unit_price").
    StorageKey("currency").  // triggers numeric(20,4) column
    Min(0.0)

// preferred idiomatic declaration:
field.Other("unit_price", schema.Currency{}).
    Default(schema.Currency{})
```

When currency fields are read from the DB, the framework returns them as `decimal.Decimal`, never as `float64`.

> **Important:** Currency values are serialised as **strings** in JSON API responses (`"4500.0000"`, not `4500.0`). This is intentional: JavaScript's `Number` type is IEEE 754 `double`, which cannot represent all `numeric(20,4)` values exactly. Clients must parse the string with a decimal library, not with `parseFloat()`.

#### `Bool` — Boolean, Never Nullable

Maps to `boolean NOT NULL DEFAULT false`. Boolean fields are never nullable in Awo — a missing value is always `false`, never `NULL`. If you need a tri-state, use a `Select` with options `yes / no / unknown`.

```go
field.Bool("is_active").
    Default(true)
```

#### `Date` — Calendar Date, No Timezone

Maps to `date`. Stores only year-month-day. Appropriate for birthdays, due dates, financial period boundaries — anything where time-of-day is irrelevant. Input accepts `YYYY-MM-DD` strings; stored without timezone.

```go
field.Time("due_date").
    StorageKey("date").
    Optional()
```

#### `DateTime` — Timestamp with Timezone, Stored as UTC

Maps to `timestamptz`. Always stored in UTC, always returned in UTC, always displayed in the tenant's configured timezone (default: `Africa/Nairobi`, EAT = UTC+3) at the presentation layer. Never store local time in a `DateTime` field.

```go
field.Time("submitted_at").
    Default(time.Now)
```

#### `Time` — Time of Day

Maps to `time without time zone`. Use for business-hours fields, shift start/end times. Stored as `HH:MM:SS`.

```go
field.Other("shift_start", schema.TimeOnly{})
```

#### `UUID` — `uuid` Column, Auto-Generated Default

Every EntityDefinition has a primary key UUID field generated automatically. You can also declare additional UUID fields for external identifiers or cross-system references.

```go
field.UUID("external_ref_id", uuid.UUID{}).
    Default(uuid.New).
    Unique()
```

---

### 5.1.2. Structured Types

Structured fields encode constrained compound values within a single column.

#### `Select` — Single Value from a Declared Option Set

Enforces that the field value is one of a declared set of options. Backed by `varchar(100)`. Option values are stored as stable keys; display labels are separate and translatable.

```go
field.Enum("status").
    Values("draft", "submitted", "approved", "rejected").
    Default("draft")
```

Never store display strings as the value — store the stable key. Labels can change; keys should not.

#### `MultiSelect` — Set of Values from a Declared Option Set

Stores multiple selections. Backed by `text[]` (PostgreSQL text array). Supports GIN indexing for `@>` containment queries.

```go
field.Strings("tags").
    Optional()
```

Use `MultiSelect` for flags, categories, and roles that a record can hold simultaneously.

#### `JSON` — Arbitrary JSONB, Schema-Validated at Application Layer

Maps to `jsonb`. Stores arbitrary structured data. The PostgreSQL layer validates only that it is valid JSON; the framework applies a Go struct validator at the application layer.

```go
field.JSON("metadata", map[string]interface{}{}).
    Optional()
```

Use sparingly. JSONB fields are opaque to many reporting tools. Prefer promoting fields to concrete columns once the shape stabilises.

---

### 5.1.3. Relational Types

Relational fields link one EntityDefinition to another.

#### `Link` — Foreign Key to Another EntityDefinition

Generates a `uuid` column with a foreign key constraint to the target entity's primary key table. Produces an index automatically.

```go
edge.From("customer", Customer.Type).
    Ref("invoices").
    Field("customer_id").
    Required()
```

#### `DynamicLink` — Polymorphic Foreign Key

Carries two columns: `{field}_type varchar(100)` (the entity name) and `{field}_id uuid`. There is **no** foreign key constraint — integrity is enforced at the application layer. Use when a record may reference any of several entity types (attachments, comments, audit entries).

```go
field.String("linked_type").  // e.g. "Invoice", "Payment", "Vendor"
    MaxLen(100).
    Optional()

field.UUID("linked_id", uuid.UUID{}).
    Optional()
```

See §6.5 for full polymorphic relationship patterns.

#### `Table` — Child Entity Inline (One-to-Many in Same Form)

Not a column type — a UI rendering hint that tells the SDUI layer to embed the child entity's list as an editable table within the parent form. The actual storage is a standard one-to-many edge.

```go
// Declared on the parent's SDUI page builder, not in the schema:
schema.TableField("items", InvoiceItem.Type)
```

---

### 5.1.4. File Types

File fields store references to binary objects, not the binary data itself.

#### `Attach` — File Reference, Stored Path or Object Storage Key

Stores a string path or object-storage key (e.g. `s3://bucket/path/file.pdf`). The framework does not store files in the database. Upload is handled by a separate presigned-URL endpoint; the resulting key is then written to this field.

```go
field.String("attachment_path").
    MaxLen(1024).
    Optional()
```

#### `AttachImage` — Image Reference with Thumbnail Metadata

Like `Attach` but stores additional JSONB metadata: original dimensions, thumbnail key, MIME type. The SDUI layer uses this metadata to render an image preview without fetching the full file.

```go
field.JSON("logo_image", schema.ImageMeta{}).
    Optional()
```

---

## 5.2. Field Options and Constraints

Every field declaration accepts a set of chainable options that control validation, storage, and API behaviour.

### 5.2.1. `Required` — Non-Nullable, Validated Before Persist

```go
field.String("company_name").
    NotEmpty()   // equivalent to Required + MinLen(1)
```

`Required` generates `NOT NULL` in SQL **and** a pre-persist validator. The validator runs before the SQL statement executes, so users receive a field-level error message, not a database constraint violation.

### 5.2.2. `Unique` — Unique Index, Validated Before Persist

```go
field.String("email").
    Unique().
    MaxLen(254)
```

Generates a unique index. The framework also runs a pre-persist uniqueness check inside the transaction to surface a user-friendly error before the DB rejects the insert. For multi-column uniqueness, declare a composite index on the entity's `Indexes()` method.

### 5.2.3. `Immutable` — Set on Create, Rejected on Update

```go
field.UUID("original_document_id", uuid.UUID{}).
    Immutable().
    Optional()
```

The framework's update path strips immutable fields from the update payload and rejects an explicit attempt to change them with `ErrImmutableField`. Use for audit keys, document lineage, and creation-time references.

### 5.2.4. `Sensitive` — Excluded from Logs, Excluded from API Responses Unless Explicitly Requested

```go
field.String("tax_pin").
    MaxLen(20).
    Sensitive()
```

Sensitive fields are:
- Redacted in all log output (`[REDACTED]` substituted)
- Excluded from default API list responses
- Excluded from webhook payloads
- Included only in explicit `GET /api/v1/{entity}/{id}?include_sensitive=true` requests by users with the `view_sensitive` permission

### 5.2.5. `Default` — Static Value or Go Function

```go
field.String("status").
    Default("draft")          // static string

field.Time("created_at").
    Default(time.Now)         // Go function, called at create time

field.UUID("id", uuid.UUID{}).
    Default(uuid.New)         // function reference
```

Function defaults are evaluated lazily at record creation, never at schema load time.

### 5.2.6. `MaxLen` — Enforced at Validator, Not Only at DB

```go
field.String("phone").
    MaxLen(20)
```

`MaxLen` generates both a `varchar(n)` column and a pre-persist validator. Users receive a `FieldTooLong` error with the field name and limit before the query reaches the DB.

### 5.2.7. `Min` / `Max` — For Numeric Fields

```go
field.Int("discount_percent").
    Min(0).
    Max(100)

field.Float("weight_kg").
    Min(0.001)
```

Both bounds are inclusive. Validation runs before persist. The framework returns a structured `FieldOutOfRange` error with the field name, value, and permitted range.

### 5.2.8. `Options` — Declared Option Set for Select and MultiSelect

```go
field.Enum("priority").
    Values("low", "medium", "high", "critical").
    Default("medium")
```

Option values are stable keys. Never rename them after data exists. To add a label, use the SDUI layer's translation keys, not the stored value.

### 5.2.9. `Translatable` — Value Stored with Locale Key, Resolved at Response Time

```go
field.String("description").
    Translatable()
```

The storage column holds the default-locale value. A companion `{field}_translations` JSONB column stores `{"sw": "...", "fr": "..."}` overrides. At response time, the framework resolves the value using the request's `Accept-Language` header and falls back to the default locale.

---

## 5.3. Field Validators

Validators are functions that inspect field values (and optionally sibling fields or the database) and return typed errors. They run after type coercion and before the persistence layer.

### 5.3.1. Built-in Validators

| Validator | Field Types | Description |
|---|---|---|
| `validate.Email()` | `Data` | RFC 5321 format |
| `validate.Phone()` | `Data` | E.164 format (+254...) |
| `validate.URL()` | `Data` | Absolute HTTP/HTTPS |
| `validate.Regex(pattern)` | `Data`, `SmallText` | Compiled at boot |
| `validate.KESAmount()` | `Currency` | Non-negative, max 4 d.p. |
| `validate.KRAPin()` | `Data` | Kenyan KRA PIN format (A000000000X) |
| `validate.NHIF()` | `Data` | NHIF member number format |

```go
field.String("kra_pin").
    MaxLen(11).
    Validate(validate.KRAPin())

field.String("phone").
    MaxLen(15).
    Validate(validate.Phone())
```

### 5.3.2. Writing a Custom Field Validator

A field validator satisfies `func(v interface{}) error`. Return `nil` for valid, `validate.Errorf("field_name", "message")` for invalid.

```go
func validateKenyanID(v interface{}) error {
    s, ok := v.(string)
    if !ok {
        return validate.Errorf("national_id", "must be a string")
    }
    if len(s) < 7 || len(s) > 8 {
        return validate.Errorf("national_id", "must be 7-8 digits")
    }
    for _, c := range s {
        if c < '0' || c > '9' {
            return validate.Errorf("national_id", "must contain only digits")
        }
    }
    return nil
}

// In schema:
field.String("national_id").
    MaxLen(8).
    Validate(validateKenyanID)
```

### 5.3.3. Cross-Field Validators — Validators That Read Sibling Field Values

Cross-field validators receive the full `EntityRecord` rather than a single value. Declare them at the entity level, not the field level:

```go
func (Invoice) Validators() []schema.EntityValidator {
    return []schema.EntityValidator{
        func(r schema.EntityRecord) error {
            start := r.GetTime("period_start")
            end := r.GetTime("period_end")
            if !end.After(start) {
                return validate.FieldErrorf("period_end",
                    "must be after period start (%s)", start.Format("2006-01-02"))
            }
            return nil
        },
    }
}
```

### 5.3.4. Async Validators — Validators That Query the DB

Async validators receive a `context.Context` and a repository reference, enabling uniqueness checks and cross-entity consistency checks that require a DB round-trip.

```go
func uniqueEmailValidator(repo CustomerRepository) schema.AsyncValidator {
    return func(ctx context.Context, r schema.EntityRecord) error {
        email := r.GetString("email")
        exists, err := repo.Exists(ctx, filter.Eq("email", email).
            And(filter.Neq("id", r.GetUUID("id"))))
        if err != nil {
            return err
        }
        if exists {
            return validate.FieldErrorf("email",
                "email address %q is already registered", email)
        }
        return nil
    }
}
```

Async validators run inside the same database transaction as the subsequent persist, so the result is consistent.

### 5.3.5. Validator Execution Order and Short-Circuit Behaviour

Validators run in this order:
1. Type coercion — rejects invalid JSON types
2. `Required` / `MaxLen` / `Min` / `Max` — field-level constraints
3. Field validators registered with `Validate()`
4. Entity-level cross-field validators
5. Async validators

By default, **all** field-level validators run (not short-circuited) so the user receives all field errors in one response. Cross-field and async validators run only after all field-level validators pass. This two-phase approach prevents confusing error messages like "period_end must be after period_start" when period_start itself is missing.

### 5.3.6. Returning Field-Level Validation Errors for amis Rendering

The framework returns validation errors in a structure that amis form components render directly:

```json
{
  "status": 422,
  "errors": [
    { "field": "email", "message": "email address \"x\" is already registered" },
    { "field": "phone", "message": "must be in E.164 format" }
  ]
}
```

Each error binds to the amis form field by the `name` property. No client-side glue code is required — amis picks up the `errors` array automatically.

---

## 5.4. Naming Series

Naming series give ERP documents human-readable, sequential identifiers: `INV-2025-00042`, `PO-2025-KE-00001`. These are distinct from UUIDs (which are the actual primary keys) and exist for compliance, communication, and operator convenience.

### 5.4.1. What Naming Series Are and Why ERP Documents Need Them

Regulatory compliance in Kenya (KRA eTIMS) and East Africa requires that tax documents have sequential, gap-free identifiers that can be audited. Naming series generate such identifiers atomically in the database, ensuring no gaps under concurrent load.

### 5.4.2. Declaring a Naming Series on an EntityDefinition

```go
func (Invoice) Config() schema.EntityConfig {
    return schema.EntityConfig{
        NamingSeries: schema.NamingSeries{
            Field:   "invoice_number",
            Format:  "INV-{YYYY}-{SEQ:05}",
            ResetOn: schema.ResetAnnually,
        },
    }
}
```

The `invoice_number` field must be declared as `Immutable` and `Unique`. The framework sets it during `Create` — it cannot be supplied by the caller.

### 5.4.3. Format Tokens

| Token | Description | Example |
|---|---|---|
| `{PREFIX}` | Module-configured string | `INV` |
| `{YYYY}` | 4-digit year (EAT timezone) | `2025` |
| `{YY}` | 2-digit year | `25` |
| `{MM}` | Zero-padded month | `07` |
| `{DD}` | Zero-padded day | `04` |
| `{SEQ:N}` | Zero-padded sequence, N digits | `00042` |
| `{TENANT}` | First 6 characters of tenant ID | `a1b2c3` |

Combined example: `PO-{YYYY}-{MM}-{SEQ:04}` → `PO-2025-07-0001`

### 5.4.4. Sequence Management — Per-Series Atomic Counter in PostgreSQL

Each naming series stores its counter in the `awo_naming_sequences` table:

```sql
CREATE TABLE awo_naming_sequences (
    series_key  varchar(200) PRIMARY KEY,  -- e.g. "Invoice:2025"
    current_seq bigint       NOT NULL DEFAULT 0
);
```

The next sequence number is obtained with:

```sql
UPDATE awo_naming_sequences
SET current_seq = current_seq + 1
WHERE series_key = $1
RETURNING current_seq;
```

This is a single atomic statement — no SELECT then UPDATE race condition. The update runs inside the same transaction as the entity insert, so a rolled-back create does not consume a sequence number (the counter is decremented by the rollback).

### 5.4.5. Reset Rules

| Rule | Behaviour |
|---|---|
| `ResetAnnually` | Sequence resets to 1 on January 1 (EAT) |
| `ResetMonthly` | Resets on the first of each month |
| `NeverReset` | Monotonically increasing forever |

Use `ResetAnnually` for invoices, purchase orders, and any document referenced in annual financial statements. Use `NeverReset` for cases, tickets, and identifiers that may be referenced years later.

### 5.4.6. Tenant-Specific Series Prefix Overrides

Tenants can configure their own prefix via the admin UI:

```go
// TenantNamingConfig is stored in the Tenant entity's config JSONB
type TenantNamingConfig struct {
    InvoicePrefix    string `json:"invoice_prefix"`    // e.g. "ACME-INV"
    PurchasePrefix   string `json:"purchase_prefix"`
}
```

The framework substitutes `{PREFIX}` with the tenant's configured value, falling back to the module default.

### 5.4.7. Retroactive Renumbering — When It Is Safe and When It Is Never Safe

**Never renumber submitted or tax-relevant documents.** Kenyan tax law and IFRS audit trails require that invoice numbers are stable, sequential, and gap-free once issued.

Retroactive renumbering is acceptable only for:
- Draft documents that have never been shared externally
- Internal reference codes that carry no legal weight

To renumber a draft series, use `awo naming resequence --entity=Invoice --status=draft --confirm`. This command requires platform-admin privilege and leaves an audit log entry.

---

## Chapter Summary

Chapter 5 documents the complete field type system (§5.1), the full option and constraint vocabulary (§5.2), the four-tier validator execution pipeline including async validators (§5.3), and the naming series system with its sequence management and reset policies (§5.4).

The three most critical concepts:

- **`Currency` type serialises as a string in JSON** (`"4500.0000"`) to preserve `numeric(20,4)` precision across JavaScript clients that use IEEE 754. Never store monetary amounts as `Float`.
- **`Searchable()` fields get a GIN index via `pg_trgm`** automatically — declare it on `Data` fields that need ILIKE-style substring search, not on `LongText` fields.
- **Naming series sequence `nextval` runs inside the INSERT transaction** — a rolled-back create consumes a sequence value, creating a gap. Gaps are normal; auditors should be informed.

**Next chapters to read:**

- [§6 — Edges](06-edges.md) — builds on `Link`, `DynamicLink`, and `Table` field types to express entity relationships
- [§7 — Entity Record Lifecycle](07-entity-record-lifecycle.md) — hook execution order and transaction boundaries that interact with field validation
- [§10 — Custom Fields](10-custom-fields.md) — how `CustomFieldDef` extends system entity field lists at runtime using the types from this chapter
