### Chapter 5 — Field System

Fields are the atomic building blocks of every `EntityDefinition`. A field declaration is not merely a column specification — it is the authoritative description of an attribute across every layer of the framework: its storage type in PostgreSQL, its validation rules at the application layer, its wire format in the API response, its component hint to the page builder, and its behaviour under privacy policies. Every decision the framework makes about an attribute — whether to index it, how to serialise it, whether to log it, whether to mask it — is derived from the field declaration. This chapter documents every field type, every option, every validator, and the naming series system.

---

#### 5.1. Field Types Reference

Fields are declared using `entity.Field("name").Type(entity.FieldType)`. The type determines the PostgreSQL column type, the Go value type in `EntityRecord`, the default serialisation format in API responses, and which options are valid. Types are grouped into four families: scalar, structured, relational, and file.

##### 5.1.1. Scalar types

Scalar types map to a single PostgreSQL column and a single Go value.

###### 5.1.1.1. `Data` — UTF-8 string, configurable max length

`Data` is the standard short-string type. It maps to a `varchar(n)` column where `n` is set by the `MaxLen` option (default 140, maximum 1000). Use `Data` for human-readable identifiers, names, codes, and labels that have a meaningful length boundary. The framework creates a B-tree index on `Data` fields declared `Unique`. Full-text search on `Data` fields uses `pg_trgm`; trigger-based GIN index creation is handled automatically when `Searchable()` is declared.

```go
// Example: Data field with length constraint and uniqueness
entity.Field("account_code").
    Type(entity.Data).
    MaxLen(20).
    Unique().
    Required()
```

###### 5.1.1.2. `SmallText` — unindexed, up to 1024 characters

`SmallText` maps to `varchar(1024)`. It is intended for short descriptive strings that will never be filtered or sorted: item descriptions on a line item, a short internal note, a unit of measure label. The framework does not create any index on `SmallText` fields by default. If you find yourself needing to filter on a `SmallText` field, reconsider whether it should be a `Data` field with a `Searchable()` declaration, or whether the content should be stored differently.

###### 5.1.1.3. `LongText` — unbounded, stored as `text` column

`LongText` maps to PostgreSQL `text` with no length limit. Use it for multi-paragraph narrative content: equipment fault descriptions, approval comments, document notes, email body templates. `LongText` fields are never indexed, never sortable in the API, and are excluded from list view columns by the default page builder. The amis page builder renders `LongText` fields as `textarea` components in forms.

###### 5.1.1.4. `Int` — 64-bit signed integer

`Int` maps to `bigint` (PostgreSQL `int8`). Use it for counts, quantities where fractional values are meaningless (number of cylinders, attendance count), sequence positions, and non-financial numeric values. `Int` fields support `Min` and `Max` constraints. Do not use `Int` for monetary amounts; use `Currency` instead.

###### 5.1.1.5. `Float` — 64-bit IEEE 754

`Float` maps to `double precision`. Use it for scientific measurements, sensor readings, computed ratios, and any value where approximate representation is acceptable. `Float` is deliberately restricted from financial fields — the framework will reject a `Currency`-typed amount stored in a `Float` field at the hook layer. Fuel dip measurements in litres, GPS coordinates, and pump flow rates are appropriate `Float` uses.

> **Warning:** Never store monetary amounts in a `Float` field. IEEE 754 floating-point arithmetic is not associative for decimal values; running totals computed from `Float` columns will accumulate rounding errors that violate double-entry accounting invariants. Use `Currency` for all monetary values without exception.

###### 5.1.1.6. `Currency` — `numeric(20,4)`, never floating point

`Currency` maps to `numeric(20,4)`: exact decimal storage with 20 significant digits and 4 decimal places. This precision supports KES amounts up to 9,999,999,999,999,999.9999 — sufficient for any realistic business transaction. The framework serialises `Currency` values as strings in JSON API responses (`"4500.0000"`) to preserve precision across JavaScript clients that use IEEE 754 for all numbers. Go code receives and sends `Currency` values as `decimal.Decimal` from the `shopspring/decimal` package.

```go
// Example: Currency field for a monetary amount
entity.Field("invoice_total").
    Type(entity.Currency).
    Min(decimal.Zero).
    Required()
```

###### 5.1.1.7. `Bool` — boolean, never nullable

`Bool` maps to PostgreSQL `boolean NOT NULL`. Bool fields in Awo are never nullable — a boolean field that has not been set is false, not NULL. This eliminates three-valued logic in boolean conditions. If you need a "not yet answered" state for a boolean concept, model it as a `Select` field with `yes`, `no`, `pending` options rather than a nullable bool.

###### 5.1.1.8. `Date` — calendar date, no timezone

`Date` maps to PostgreSQL `date`. It represents a calendar date without any time-of-day or timezone component. Use `Date` for posting dates, document dates, leave dates, and any date where the time is irrelevant. Filtering on `Date` fields by range is a common pattern; the framework creates a B-tree index on `Date` fields declared `Indexed()`. The API wire format is `YYYY-MM-DD`.

###### 5.1.1.9. `DateTime` — timestamp with timezone, stored as UTC

`DateTime` maps to `timestamptz`. Values are stored in UTC and converted to East Africa Time (EAT, UTC+3) for display by the amis page builder when the tenant's locale is `ke`. The API wire format is ISO 8601 with UTC offset: `2025-06-01T09:14:33Z`. Use `DateTime` for event timestamps, workflow start times, payment timestamps, and any moment where the exact instant matters.

###### 5.1.1.10. `Time` — time of day

`Time` maps to PostgreSQL `time without time zone`. Use it for shift start times, scheduled maintenance windows, and any recurring time-of-day value where the calendar date is irrelevant. The API wire format is `HH:MM:SS`. `Time` fields are uncommon; most event recording uses `DateTime`.

###### 5.1.1.11. `UUID` — `uuid` column, auto-generated default

`UUID` maps to PostgreSQL `uuid`. The framework provides a `gen_random_uuid()` default for UUID fields declared without an explicit `Default`. UUID fields are commonly used for external reference IDs — a payment gateway transaction ID, an external system's record ID — where the value is assigned externally and not generated by Awo's naming series. The primary key `id` field present on all entities is a UUID field managed by the framework and is not declared explicitly in the field list.

---

##### 5.1.2. Structured types

Structured types store multiple values or constrained value sets.

###### 5.1.2.1. `Select` — single value from a declared option set

`Select` maps to `text` with an application-layer constraint that the value must be one of the declared options. It does not use a PostgreSQL `ENUM` type because adding an option to a PostgreSQL ENUM requires a table rewrite; changing `Options` in the `EntityDefinition` and redeploying is sufficient for the application-layer constraint, with no migration required. The `Default` option on a `Select` field must be one of the declared options; the framework validates this at startup.

```go
// Example: Select field for document status
entity.Field("status").
    Type(entity.Select).
    Options("draft", "pending_approval", "approved", "rejected", "posted").
    Default("draft").
    Required()
```

The amis page builder renders `Select` fields as dropdown components in forms and as badge/tag renderers in list columns when configured with a `Map` option.

###### 5.1.2.2. `MultiSelect` — set of values from a declared option set

`MultiSelect` maps to `text[]` (PostgreSQL text array). Values are validated individually against the declared option set at the application layer. Use `MultiSelect` for tagging, category assignment, and any attribute that is naturally a set: the payment methods a customer accepts, the product grades a tank stores, the modules enabled for a tenant. The API wire format is a JSON array of strings. GIN indexing on `MultiSelect` fields is created automatically when `Indexed()` is declared.

###### 5.1.2.3. `JSON` — arbitrary JSONB, schema-validated at the application layer

`JSON` maps to `jsonb`. Unlike custom entity JSONB storage, a `JSON` field on a system entity is a deliberately unstructured sub-document within an otherwise typed record. Use it sparingly: for configuration blobs, for external API response caching alongside the structured data derived from it, or for extensible metadata where the schema is expected to evolve. Schema validation for `JSON` fields is declared as a Go function in `Validate()` and is applied at the hook layer before persist.

> **Warning:** Do not use `JSON` fields as a backdoor for untyped storage on system entities. If you find yourself putting financial values, party identifiers, or any queryable attribute into a `JSON` field, those attributes should be promoted to typed columns. `JSON` fields cannot be efficiently filtered in list queries and cannot participate in SQL joins.

---

##### 5.1.3. Relational types

Relational types express relationships to other entities. They generate foreign key columns, join methods, and child table renderers. The full edge system built on top of these types is documented in §6.

###### 5.1.3.1. `Link` — foreign key to another EntityDefinition

`Link` maps to a `uuid` column with a foreign key constraint referencing the primary key of the named entity's table. Declaring `entity.Field("customer_id").Type(entity.Link).LinkedEntity("Customer")` generates a `customer_id uuid REFERENCES customers(id)` column. The `EntityRepository` interface exposes the linked entity via the `Include` option on `Query` and `Get`, which triggers a JOIN rather than a separate query.

```go
// Example: Link field referencing the Customer entity
entity.Field("customer_id").
    Type(entity.Link).
    LinkedEntity("Customer").
    Required()
```

`Link` fields are rendered as search-as-you-type picker components by the amis page builder, wired to the linked entity's list endpoint.

###### 5.1.3.2. `DynamicLink` — polymorphic foreign key, carries entity name + id

`DynamicLink` represents a polymorphic relationship. It generates two columns: `{field_name}_type text` (storing the entity name) and `{field_name}_id uuid` (storing the referenced record's primary key). There is no database-level foreign key constraint because PostgreSQL cannot enforce a FK across multiple tables from a single column pair; integrity is enforced at the application layer by the `before_save` hook that validates the referenced entity and record exist.

Use `DynamicLink` when a field can point to records of different entity types — for example, an `Attachment` that can be linked to either a `PurchaseOrder` or a `SalesInvoice`. The `EntityRecord` for a `DynamicLink` field exposes `GetLinkType()` and `GetLinkID()` accessors. Full DynamicLink querying patterns are in §6.5.

> **Warning:** `DynamicLink` fields have no database-level referential integrity. If the referenced record is deleted, the `DynamicLink` value becomes a dangling reference unless a `before_delete` hook on the referenced entity explicitly checks for and handles linked records. Always register such a guard hook when introducing a `DynamicLink`.

###### 5.1.3.3. `Table` — child entity inline (one-to-many in same form)

`Table` declares a one-to-many relationship where the child records are presented as an inline editable grid within the parent entity's form. It does not generate a column on the parent entity's table; instead it generates the reverse foreign key on the child entity's table. `Table` fields are rendered as `LineItemEditor` components by the default amis page builder. They are the standard mechanism for invoice lines, order lines, payroll deduction lines, and any parent-child editing pattern.

```go
// Example: Table field for invoice line items
entity.Field("line_items").
    Type(entity.Table).
    ChildEntity("SalesInvoiceLine").
    Required()
```

---

##### 5.1.4. File types

File types store references to uploaded files rather than the file data itself.

###### 5.1.4.1. `Attach` — file reference, stored path or object storage key

`Attach` maps to `text` storing either a filesystem path or an object storage key (S3-compatible). The framework does not store file data in the database. File upload is handled by a dedicated upload endpoint that returns a reference string; this string is then submitted as the `Attach` field value. The amis page builder renders `Attach` fields as file upload components with drag-and-drop support and file type restriction based on the `AllowedTypes` option.

###### 5.1.4.2. `AttachImage` — image reference with thumbnail metadata

`AttachImage` extends `Attach` with additional metadata stored as a JSON sub-document: original dimensions, thumbnail URL, MIME type, and file size. The framework generates a thumbnail automatically on upload for image types. The amis page builder renders `AttachImage` fields as image upload components with preview and optional crop functionality.

---

#### 5.2. Field Options and Constraints

Options and constraints are declared fluently after the type declaration. Multiple options can be chained. Some options are type-specific (e.g., `MaxLen` is only valid on `Data` and `SmallText`; `Options` is only valid on `Select` and `MultiSelect`); the framework validates option compatibility at startup and fails the process if an incompatible option is applied.

##### 5.2.1. `Required` — non-nullable, validated before persist

`Required()` marks a field as mandatory. At the database level, the column is created `NOT NULL`. At the application layer, the field validator checks for the field's presence and non-empty value before any hook is invoked. For `Data` and `SmallText` fields, an empty string is treated as absent and fails the required check. For `Currency` and `Int` fields, zero is a valid value and does not fail the required check; use `Min(1)` if zero is not acceptable.

`Required()` applies to both create and update operations by default. Declaring `RequiredOnCreate()` relaxes the constraint so that an update payload does not need to include the field. This is useful for fields that are set programmatically on create (naming series, created_by) and should not be re-submitted on update.

##### 5.2.2. `Unique` — unique index, validated before persist

`Unique()` creates a `UNIQUE INDEX` on the column in the tenant schema. At the application layer, the framework's async validator (§5.3.4) checks uniqueness before persist to produce a user-friendly field-level error rather than a database constraint violation. If two concurrent requests race past the async validator, the database constraint catches the conflict and the framework maps the PostgreSQL error to a `ConflictError` (§18.1.4).

Unique indexes are per-tenant schema by design. A `request_number` value that is unique within `t_acme` may exist identically within `t_other_tenant`; this is expected and correct.

##### 5.2.3. `Immutable` — set on create, rejected on update

`Immutable()` declares that the field's value is set once on create and cannot be changed thereafter. At the application layer, the `before_save` hook for update operations rejects any payload that includes an `Immutable` field with a value different from the current persisted value. Use `Immutable` for: primary identifiers set at creation (`request_number`, `invoice_number`), the `created_at` timestamp, and any field that represents the original intent of a document (a journal entry's `posting_date` is immutable; corrections require a reversal entry).

##### 5.2.4. `Sensitive` — excluded from logs, excluded from API responses unless explicitly requested

`Sensitive()` marks a field as containing confidential data. The structured logging middleware excludes all `Sensitive` fields from request and response log lines. The API response serialiser excludes `Sensitive` fields from the default response envelope; they are only included when the caller explicitly requests them via the `fields=` query parameter, and only if the caller's role has field-level read permission. Use `Sensitive` for: KRA PIN numbers, bank account numbers, salary amounts, ID document numbers, and any field that regulators or privacy policies require to be handled with care.

```go
// Example: Sensitive field for a confidential attribute
entity.Field("kra_pin").
    Type(entity.Data).
    MaxLen(11).
    Sensitive().
    Unique()
```

##### 5.2.5. `Default` — static value or Go function

`Default` sets the value used when the field is omitted from a create payload. It accepts either a static value of the field's declared type or a Go function with signature `func() T` where `T` matches the field type. The framework evaluates function defaults at record assembly time, not at package init time, ensuring that time-based defaults return the current time rather than the process start time.

```go
// Example: Static default and function default
entity.Field("status").Type(entity.Select).Default("draft")
entity.Field("created_at").Type(entity.DateTime).Default(entity.Now)
entity.Field("reference_currency").Type(entity.Data).Default("KES")
```

`entity.Now` is a framework-provided default function that returns `time.Now().UTC()`. Custom default functions must be pure and fast; they run inside the record assembly path on every create operation.

##### 5.2.6. `MaxLen` — enforced at validator, not only at DB

`MaxLen(n)` sets the maximum character length for `Data` and `SmallText` fields. It is enforced by the field validator before the database is contacted, producing a user-friendly validation error rather than a database `string_data_right_truncation` error. The PostgreSQL column is created as `varchar(n)`, providing a second enforcement at the database level as a backstop.

##### 5.2.7. `Min` / `Max` — for numeric fields

`Min(v)` and `Max(v)` set inclusive lower and upper bounds for `Int`, `Float`, and `Currency` fields. They are enforced at the field validator before persist. For `Currency` fields, `v` must be a `decimal.Decimal`; for `Int` and `Float`, `v` must be the corresponding Go numeric type. Range checks happen before any hook is invoked; a field that fails a range check returns a 422 error immediately.

##### 5.2.8. `Options` — declared option set for Select and MultiSelect

`Options(values ...string)` declares the valid values for `Select` and `MultiSelect` fields. Options are validated at field validator time; submitting a value not in the declared set returns a 422 validation error. Options are stable identifiers — they appear in database values, filter query parameters, and permission matrix rows — and must not be renamed after data has been written. To deprecate an option, add a new option and stop using the old one; never remove an option that may exist in persisted records.

Option labels for display (the human-readable string shown in the amis dropdown) are declared separately using `OptionLabel(value, label)` or via the tenant's translation catalogue. This separation ensures that changing a display label does not require a data migration.

##### 5.2.9. `Translatable` — value stored with locale key, resolved at response time

`Translatable()` marks a field whose value is a locale key rather than a raw string. At response serialisation time, the framework looks up the key in the tenant's translation catalogue and returns the resolved string for the request's locale. Use `Translatable` for system-managed string values that must appear in multiple languages: status labels, document type names, error message codes. Do not use it for user-entered content; user text is never translatable via this mechanism.

---

#### 5.3. Field Validators

Validators execute in a defined order after the field values are assembled but before any hook is invoked. A validator returns a `*entity.FieldError` to report a field-specific validation failure, or `nil` to indicate the field passes. Multiple validators can be registered on a single field; they execute in declaration order and short-circuit on the first failure by default.

##### 5.3.1. Built-in validators — email, phone (E.164), URL, regex, KES amount

The framework ships built-in validators for common formats:

- `entity.ValidateEmail()` — RFC 5322 email address format.
- `entity.ValidatePhone()` — E.164 international phone number format (`+254XXXXXXXXX` for Kenya). The validator normalises Kenyan numbers starting with `07` or `01` to E.164 automatically.
- `entity.ValidateURL()` — HTTP or HTTPS URL with a resolvable host.
- `entity.ValidateRegex(pattern)` — matches against a compiled `regexp.MustCompile` pattern. The pattern is compiled at startup, not per validation call.
- `entity.ValidateKESAmount()` — positive `numeric(20,4)` value with at most 4 decimal places. Rejects negative amounts and values with more than 4 decimal places.

```go
// Example: Applying built-in validators to fields
entity.Field("contact_email").
    Type(entity.Data).
    MaxLen(254).
    Validate(entity.ValidateEmail())

entity.Field("mpesa_phone").
    Type(entity.Data).
    MaxLen(15).
    Validate(entity.ValidatePhone())
```

##### 5.3.2. Writing a custom field validator

A custom validator is any function with signature `func(value any, record entity.EntityRecord) *entity.FieldError`. The `value` argument is the field's current value (already coerced to the field's declared Go type). The `record` argument provides read-only access to other fields in the record being validated, enabling cross-field checks.

```go
// Example: Custom validator for vehicle registration format
func validateKenyanRegistration(value any, record entity.EntityRecord) *entity.FieldError {
    reg, ok := value.(string)
    if !ok || reg == "" {
        return nil // required check is handled separately
    }
    // Kenya registration formats: KCC 123A, KAB 001Z, custom plates
    matched, _ := regexp.MatchString(`^[A-Z]{3}\s\d{3}[A-Z]$`, reg)
    if !matched {
        return entity.NewFieldError("vehicle_registration",
            "must be a valid Kenya vehicle registration (e.g. KCC 123A)")
    }
    return nil
}

// Applied in the field declaration:
entity.Field("vehicle_registration").
    Type(entity.Data).
    MaxLen(20).
    Validate(validateKenyanRegistration)
```

##### 5.3.3. Cross-field validators — validators that read sibling field values

A cross-field validator uses the `record` argument to read other field values. This is the correct mechanism for rules like "if payment method is `mpesa`, then `mpesa_phone` is required" — a rule that cannot be expressed by per-field `Required()` because the requirement is conditional.

```go
// Example: Cross-field validator enforcing conditional requirement
func validateMpesaPhoneRequired(value any, record entity.EntityRecord) *entity.FieldError {
    paymentMethod, _ := record.Get("payment_method").(string)
    if paymentMethod != "mpesa" {
        return nil
    }
    phone, _ := value.(string)
    if phone == "" {
        return entity.NewFieldError("mpesa_phone",
            "mpesa_phone is required when payment_method is mpesa")
    }
    return nil
}
```

Cross-field validators execute after all per-field type checks pass. They should not assume the sibling field has been validated; the `record` value at validation time may contain raw unvalidated input for fields that come later in the validation order.

##### 5.3.4. Async validators — validators that query the DB (uniqueness checks)

Async validators are declared with `ValidateAsync(fn)` and receive a `context.Context` and an `entity.EntityRepository` in addition to the field value and record. They are used for uniqueness checks, existence checks (verify a linked entity exists before creating the record), and any validation that requires a database round-trip.

```go
// Example: Async validator checking uniqueness of a tax PIN
func validateUniquePIN(
    ctx context.Context,
    repo entity.EntityRepository,
    value any,
    record entity.EntityRecord,
) *entity.FieldError {
    tc, err := tenant.FromContext(ctx)
    if err != nil {
        return entity.NewFieldError("kra_pin", "could not validate PIN: "+err.Error())
    }
    _ = tc
    pin, _ := value.(string)
    filter := entity.NewFilter().Eq("kra_pin", pin).Neq("id", record.ID())
    exists, err := repo.Exists(ctx, filter)
    if err != nil {
        return entity.NewFieldError("kra_pin", "validation error: "+err.Error())
    }
    if exists {
        return entity.NewFieldError("kra_pin", "a customer with this KRA PIN already exists")
    }
    return nil
}
```

Async validators run after all synchronous validators pass, and before any `before_save` hook is invoked. They execute concurrently for fields that have independent async validators; the framework collects all errors before returning to the caller.

##### 5.3.5. Validator execution order and short-circuit behaviour

The execution order for a field is: built-in type coercion → `Required` check → `MaxLen`/`Min`/`Max` checks → declared synchronous validators (in declaration order) → declared async validators (concurrently). If any step in the synchronous chain returns a non-nil `FieldError`, the chain short-circuits for that field and the async validator for that field does not run.

At the record level, validators for all fields run concurrently. The framework collects all field errors across all fields and returns them together in the 422 response, so the caller receives all validation failures in one round-trip rather than discovering them one at a time.

##### 5.3.6. Returning field-level validation errors for amis rendering

The amis form component expects field-level errors in the shape `{ name: "field_name", errors: ["message"] }` inside a top-level `errors` array in the API response. The framework's error serialiser maps `entity.FieldError` values to this exact shape automatically. A validator returning `entity.NewFieldError("mpesa_phone", "must be a valid phone number")` produces:

```json
{
  "errors": [
    { "name": "mpesa_phone", "errors": ["must be a valid phone number"] }
  ]
}
```

The amis form uses these field names to highlight the corresponding input components in the UI. Field names in the error response must exactly match the field names declared in the `EntityDefinition`; any mismatch causes the error to fall through to the form-level error display rather than the field-level highlight.

---

#### 5.4. Naming Series

Naming series generate unique, human-readable document numbers for entities that participate in business workflows. A service request with number `SR-2025-06-0042` is traceable, auditable, and communicable in a way that a raw UUID is not. Every Awo entity that represents a business document — invoices, orders, journal entries, service requests, payment entries — should have a naming series.

##### 5.4.1. What naming series are and why ERP documents need them

A naming series is a sequential numbering scheme that encodes document type, date, and a monotonically increasing counter into a human-readable string. They serve two purposes: human-facing identification (a customer can quote `INV-2025-06-0001` over the phone) and audit trail completeness (sequential gaps in an invoice series are a red flag for regulators, which the series counter makes detectable). The KRA expects Kenyan businesses to number tax invoices sequentially; Awo's naming series system satisfies this requirement structurally.

##### 5.4.2. Declaring a naming series on an EntityDefinition

A naming series is declared using `entity.NamingSeries(series)` in the `EntityDefinition`. The series is applied to the field declared `Type(entity.Data).Immutable()` that will hold the generated number. The framework assigns the series value during record assembly, before `before_save` hooks are invoked, ensuring that the number is available in hooks for cross-reference operations.

```go
// Example: Full naming series configuration for a sales invoice
var SalesInvoiceDefinition = entity.Define("SalesInvoice",
    entity.Fields(
        entity.Field("invoice_number").
            Type(entity.Data).
            MaxLen(25).
            Immutable().
            Unique().
            Required(),
        // ... other fields
    ),
    entity.NamingSeries(
        entity.Series("INV-{YYYY}-{MM}-{SEQ}").
            AppliesTo("invoice_number").
            SequenceName("sales_invoice_seq").
            ResetPolicy(entity.ResetMonthly).
            PaddedLength(4),
    ),
)
```

##### 5.4.3. Format tokens — `{PREFIX}`, `{YYYY}`, `{MM}`, `{DD}`, `{SEQ}`, `{TENANT}`

Token definitions:

- `{PREFIX}` — a static string prefix configured in `TenantConfig` as a per-tenant override. If no override is set, defaults to the value specified in `Series()`. Useful for stations that want their own prefix (`SMZ` for Shell Maanzoni, `NRB` for a Nairobi branch).
- `{YYYY}` — the four-digit calendar year of the record's creation date in EAT (East Africa Time, UTC+3).
- `{MM}` — the two-digit month (01–12) of the record's creation date in EAT.
- `{DD}` — the two-digit day (01–31) of the record's creation date in EAT.
- `{SEQ}` — the next value from the named PostgreSQL sequence, left-padded to `PaddedLength` with zeros.
- `{TENANT}` — the tenant's short code from `TenantConfig`. Useful for multi-entity deployments where documents from multiple tenants appear in the same reporting stream.

All date tokens use the tenant's configured timezone. For Kenyan deployments the default is EAT. A record created at `2025-06-01 22:30:00 UTC` has YYYY=2025, MM=06, DD=02 in EAT (UTC+3 makes it midnight of June 2nd).

##### 5.4.4. Sequence management — per-series atomic counter in PostgreSQL

Each naming series uses a PostgreSQL sequence named `{sequence_name}` in the tenant schema. The sequence is created by the Atlas migration generated from the `EntityDefinition`. The framework calls `nextval('{sequence_name}')` inside the same database transaction as the record insert, guaranteeing that the sequence value is committed with the record or rolled back with it. There is no application-level sequence management and no Redis counter; the database sequence is the sole source of truth.

Concurrent inserts each receive a unique sequence value. PostgreSQL sequences are designed for high-concurrency use and do not serialize concurrent transactions. Gaps in the sequence are possible when a transaction is rolled back after calling `nextval`; this is expected and acceptable. A gap in the numbering does not indicate data loss — it indicates a transaction that was rolled back. Auditors examining an invoice series should be informed of this.

> **Note:** The sequence `nextval` call is inside the record's INSERT transaction. If the `before_save` hook rejects the record after the sequence has been incremented, the sequence value is consumed but no record is persisted. This creates a gap. If zero gaps are an absolute compliance requirement, place the `before_save` validation as early as possible to minimise the window between sequence consumption and potential rollback.

##### 5.4.5. Reset rules — annual reset, monthly reset, never reset

Three reset policies are available:

- `entity.ResetNever` — the sequence increments indefinitely across all time. Use for entities where the absolute sequence number has legal significance (some KRA-regulated invoice types). The counter will grow large over time but PostgreSQL `bigint` sequences support up to 9,223,372,036,854,775,807.
- `entity.ResetAnnually` — the sequence resets to 1 on the first document of a new calendar year (in EAT). The series value includes `{YYYY}` to preserve uniqueness across resets. Awo uses a per-year sequence named `{sequence_name}_{YYYY}`.
- `entity.ResetMonthly` — the sequence resets to 1 on the first document of a new calendar month. The series value includes both `{YYYY}` and `{MM}`. This is the most common policy for Kenyan business documents. Awo uses a per-month sequence named `{sequence_name}_{YYYY}_{MM}`.

The framework creates reset sequences lazily: the first document in a new year or month triggers a `CREATE SEQUENCE IF NOT EXISTS {sequence_name}_{YYYY}` call before the `nextval` call.

##### 5.4.6. Tenant-specific series prefix overrides

Tenants can configure a custom prefix for any naming series via `TenantConfig`. The configuration key is `naming_series.{entity_name}.prefix`. When a custom prefix is set, the `{PREFIX}` token in the format string resolves to the tenant's value rather than the `EntityDefinition` default.

```go
// Example: Reading tenant-specific series prefix in the series assembler (internal framework)
prefix := tc.Config().GetString(
    "naming_series.SalesInvoice.prefix",
    "INV", // default if not set
)
```

This allows a multi-branch deployment where each branch is a separate tenant to have distinguishable invoice numbers (`NBI-2025-06-0001` for Nairobi, `MSA-2025-06-0001` for Mombasa) without separate `EntityDefinition` declarations.

##### 5.4.7. Retroactive renumbering — when it is safe and when it is never safe

Retroactive renumbering — changing a document's series value after it has been issued — is never safe for fiscal documents. A sales invoice or journal entry that has been issued to a counterparty or submitted to KRA eTIMS has a legal identity tied to its number. Changing the number without cancelling and reissuing the document creates an audit trail inconsistency.

For non-fiscal documents (service requests, internal maintenance logs, draft documents that have not been submitted) renumbering is safe if: the old number has not been communicated externally, no other record references the old number as a foreign key, and the renumbering is performed within a transaction that also updates any `Immutable` field override permission granted by a superuser role. The framework does not provide a built-in renumbering tool; this is deliberate. Renumbering must be an explicit, audited operation implemented as a custom Temporal workflow with approval gating, not a casual admin action.

> **Danger:** Never run a raw `UPDATE service_requests SET request_number = 'SR-...' WHERE ...` directly in the database to renumber documents. This bypasses all hook execution, does not create an audit log entry, and does not update any index or cache that references the old number. If renumbering is genuinely required, implement it as a framework workflow so that all side effects are handled correctly.

---

#### Chapter summary

Chapter 5 documents the complete field type system (§5.1), the full option and constraint vocabulary (§5.2), the four-tier validator execution pipeline including async validators (§5.3), and the naming series system with its sequence management and reset policies (§5.4). The three most critical concepts are the `Currency` type's `numeric(20,4)` storage and string serialisation (§5.1.1.6, which prevents floating-point financial bugs), the validator execution order and concurrent field error collection (§5.3.5, which determines when validation errors surface and in what shape), and the naming series sequence management inside the INSERT transaction (§5.4.4, which is the source of gap behaviour that auditors must understand).

**Next chapters to read:**

- §6 — Edges — Relationships Between EntityDefinitions (the natural continuation after mastering fields: edge declarations build on `Link`, `DynamicLink`, and `Table` field types and generate the relational structure of the data model)
- §7 — The EntityRecord Lifecycle (the hook system that executes validators, `before_save`, `after_save`, and submission hooks: understanding lifecycle ordering is essential before writing any hook that touches fields)
- §10 — Custom Fields — Runtime Schema Extension (how `CustomFieldDef` extends system entity field lists at runtime, and which field types from §5.1 are available to custom fields)
