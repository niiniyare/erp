---
title: "Chapter 22: Foundation Components Reference"
part: "Part IV — The SDUI Layer"
chapter: 22
section: "22-foundation-components"
related:
  - "[Chapter 21: SDUI Philosophy](21-sdui-philosophy.md)"
  - "[Chapter 23: Composite Blocks](23-composite-blocks.md)"
  - "[Chapter 5: Field System](../part-02-entity-system/05-field-system.md)"
---

# Chapter 22: Foundation Components Reference

Foundation components are the individual form fields and display widgets that composite blocks (Chapter 23) are assembled from. This chapter documents the amis-backed component for each Awo field type, with configuration options and ERP-specific usage patterns.

---

## 22.1. Text and Data Fields

### 22.1.1. `TextField`

```go
amis.TextField("full_name").
    Label("Full Name").
    Placeholder("Enter full name").
    MaxLength(255).
    Required(true).
    ClearButton(true)
```

amis schema output:
```json
{
  "type": "input-text",
  "name": "full_name",
  "label": "Full Name",
  "placeholder": "Enter full name",
  "maxLength": 255,
  "required": true,
  "clearable": true
}
```

**Read-only mode** (for detail views):
```go
amis.TextField("full_name").Label("Full Name").ReadOnly(true)
// renders as: "type": "static"
```

**Input mask** (for structured data like phone numbers):
```go
amis.TextField("phone").
    Mask("+254 ### ### ###").
    Placeholder("+254 700 000 000")
```

### 22.1.2. `NumberField`

```go
amis.NumberField("unit_price").
    Label("Unit Price (KES)").
    Prefix("KES").
    Precision(2).
    ThousandSeparator(true).
    Min(0)
```

For currency fields, always use `Precision(2)` for display (2 decimal places for user input) while the underlying `Currency` field type stores 4 decimal places (`numeric(20,4)`).

### 22.1.3. `TextAreaField`

```go
amis.TextAreaField("notes").
    Label("Notes").
    Rows(4).
    MaxLength(1024).
    Resizable(true)
```

### 22.1.4. `RichTextField` — When to Use

Use `RichTextField` only for content that will be rendered as HTML (email templates, document templates, help text). For ERP forms, prefer `TextAreaField` — rich text editors are heavy and unnecessary for most data entry.

---

## 22.2. Choice Fields

### 22.2.1. `SelectField` — Static Options

```go
amis.SelectField("status").
    Label("Status").
    Options([]amis.Option{
        {Value: "draft", Label: "Draft"},
        {Value: "submitted", Label: "Submitted"},
        {Value: "approved", Label: "Approved"},
    }).
    Default("draft").
    Clearable(false)
```

**API-sourced options** (when the option list is large or dynamic):
```go
amis.SelectField("branch_id").
    Label("Branch").
    Source("GET /api/v1/branches?fields=id,name").
    ValueField("id").
    LabelField("name").
    Searchable(true)
```

### 22.2.2. `MultiSelectField`

```go
amis.MultiSelectField("tags").
    Label("Tags").
    Options(predefinedTags).
    MaxSelected(5).
    TagMode(true)  // renders as pill tags, not checkboxes
```

### 22.2.3. `RadioGroupField` — When to Prefer Over Select

Use `RadioGroupField` when there are 2-4 mutually exclusive options that the user should be able to evaluate at a glance without opening a dropdown:

```go
amis.RadioGroupField("priority").
    Label("Priority").
    Options([]amis.Option{
        {Value: "low", Label: "Low"},
        {Value: "medium", Label: "Medium"},
        {Value: "high", Label: "High"},
    }).
    Inline(true)
```

### 22.2.4. `SwitchField` and `CheckboxField`

`SwitchField` for single boolean options that have a clear "on/off" semantic:
```go
amis.SwitchField("is_active").Label("Active").Default(true)
```

`CheckboxField` when part of a list of options or when the semantic is "agree/acknowledge":
```go
amis.CheckboxField("terms_accepted").
    Label("I accept the terms and conditions").
    Required(true)
```

---

## 22.3. Date and Time Fields

### 22.3.1. `DateField`

```go
amis.DateField("due_date").
    Label("Due Date").
    Format("YYYY-MM-DD").
    DisplayFormat("DD MMM YYYY").   // "04 Jul 2025" — Kenyan convention
    MinDate("today").
    Required(true)
```

### 22.3.2. `DateRangeField`

```go
amis.DateRangeField("report_period").
    Label("Report Period").
    StartName("period_start").
    EndName("period_end").
    Format("YYYY-MM-DD").
    Shortcuts([]string{"thisMonth", "lastMonth", "thisQuarter", "thisYear"})
```

### 22.3.3. `DateTimeField` — Timezone Handling for EAT

Always store datetimes as UTC. Display in EAT (UTC+3) for Kenyan tenants:

```go
amis.DateTimeField("submitted_at").
    Label("Submitted At").
    Format("YYYY-MM-DDTHH:mm:ssZ").  // UTC for API
    DisplayFormat("DD MMM YYYY HH:mm").  // "04 Jul 2025 15:30"
    Timezone("Africa/Nairobi").
    ReadOnly(true)
```

---

## 22.4. Relational Fields

### 22.4.1. `LinkField` — Entity Picker with Search

```go
amis.LinkField("customer_id").
    Label("Customer").
    Source("GET /api/v1/customers?q=${keywords}&limit=20").
    ValueField("id").
    LabelField("name").
    SearchMinLength(2).  // start searching after 2 characters
    Required(true)
```

The `q=${keywords}` pattern uses the amis variable injection — when the user types, amis substitutes `keywords` with the input value and calls the API.

### 22.4.2. `LinkField` API Source Configuration

The search API endpoint must:
- Accept `q` as a search parameter
- Return the standard paginated envelope `{ data: [{ id, name, ... }], meta: {...} }`
- Filter by tenant automatically (via tenant middleware)
- Respect the caller's permissions (privacy policy)

```go
// Auto-generated search endpoint for Customer entity
// GET /api/v1/customers?q=acme&fields=id,name,email&limit=20
```

### 22.4.3. `DynamicLinkField`

For polymorphic references where the entity type is user-selected:

```go
amis.Group([
    amis.SelectField("linked_type").
        Label("Link To").
        Options([]amis.Option{
            {Value: "Invoice", Label: "Invoice"},
            {Value: "PurchaseOrder", Label: "Purchase Order"},
            {Value: "Customer", Label: "Customer"},
        }),
    amis.LinkField("linked_id").
        Label("Record").
        Source("GET /api/v1/${linked_type_lower}?q=${keywords}").
        VisibleOn("${linked_type}"),
])
```

---

## 22.5. File and Media Fields

### 22.5.1. `FileUploadField`

```go
amis.FileUploadField("attachment").
    Label("Attachment").
    Accept(".pdf,.docx,.xlsx,.jpg,.png").
    MaxSize(10 * 1024 * 1024).  // 10MB
    Multiple(false).
    UploadAPI("POST /api/v1/attachments/upload").
    ValueField("file_path")
```

Upload flow:
1. User selects a file
2. amis calls `POST /api/v1/attachments/upload` with multipart form data
3. Server stores file in object storage, returns `{ file_path, file_name, file_size }`
4. amis stores `file_path` as the field value
5. Form submit saves `file_path` to the entity record

### 22.5.2. `ImageUploadField`

```go
amis.ImageUploadField("logo").
    Label("Company Logo").
    Accept("image/*").
    MaxSize(2 * 1024 * 1024).   // 2MB
    CropAspectRatio(16, 9).
    PreviewWidth(160).
    PreviewHeight(90)
```

---

## 22.6. ERP-Specific Components

### 22.6.1. `LineItemEditor` — Child Table Editor

The line item editor is used for invoice lines, purchase order lines, and any one-to-many relationship where child records are entered inline in the parent form:

```go
amis.LineItemEditor("items").
    Label("Line Items").
    Endpoint("POST /api/v1/invoice-items").
    Columns([]amis.LineItemColumn{
        amis.LinkColumn("product_id", "Product").Required(true),
        amis.NumberColumn("quantity", "Qty").Min(1).Default(1),
        amis.CurrencyColumn("unit_price", "Unit Price (KES)").Required(true),
        amis.CurrencyColumn("line_total", "Total").Formula("${quantity} * ${unit_price}").ReadOnly(true),
    }).
    FooterRow(amis.SummaryRow().
        Label("Subtotal").
        Field("subtotal").
        Formula("SUM(items.line_total)"))
```

The line item editor saves child records on parent form submit, inside the same transaction.

### 22.6.2. `CurrencyField`

```go
amis.CurrencyField("total_amount").
    Label("Total Amount").
    Currency("KES").
    Precision(2).
    ReadOnly(true)
```

Displays values formatted as `KES 45,000.00`. The underlying JSON value is the numeric string `"45000.0000"` (4 decimal places from `numeric(20,4)`).

### 22.6.3. `NamingSeriesField`

```go
amis.NamingSeriesField("invoice_number").
    Label("Invoice #").
    HelpText("Auto-generated on save").
    ReadOnly(true)
```

Displayed as read-only on the form. Shows the format pattern (e.g. `INV-2025-XXXXX`) before save, and the generated value after save.

### 22.6.4. `BarcodeField`

```go
amis.BarcodeField("serial_number").
    Label("Serial Number").
    Formats([]string{"CODE_128", "QR_CODE"}).
    CameraEnabled(true).
    ManualEntry(true).
    Placeholder("Scan or enter serial number")
```

When `CameraEnabled(true)`, amis renders a camera trigger button that activates the device camera for scanning. On mobile devices this opens the native camera; on desktop it opens a webcam scanner.

### 22.6.5. `SerialBatchPicker`

For inventory operations requiring serial number or batch selection:

```go
amis.SerialBatchPicker("serial_numbers").
    Label("Serial Numbers").
    ProductField("product_id").
    WarehouseField("warehouse_id").
    Mode("serial").  // "serial" | "batch"
    Required(true).
    MinSelections(1)
```

This component fetches available serial numbers for the selected product and warehouse from `GET /api/v1/serial-numbers?product_id=${product_id}&warehouse_id=${warehouse_id}&status=available`.
