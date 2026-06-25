---
title: "Chapter 25: Print and PDF Generation"
part: "Part IV — The SDUI Layer"
chapter: 25
section: "25-print-pdf"
related:
  - "[Chapter 2: The EntityDefinition](../part-01-foundations/02-entity-definition.md)"
  - "[Chapter 29: Defining Activities](../part-05-workflow/29-defining-activities.md)"
  - "[Chapter 40: Metadata and Reporting](../part-06-modules/platform-entities.md)"
---

# Chapter 25: Print and PDF Generation

ERP systems live and die on paper: invoices, delivery notes, payslips, purchase orders, statements of account. Awo's PDF generation stack is built on headless Chrome controlled from Go via `chromedp`. HTML templates written with Go's `html/template` define each document layout, and a CSS `@media print` stylesheet handles pagination, headers, and footers. The result is pixel-accurate PDFs that use the full power of modern CSS.

---

## 25.1 Architecture and Technology Choices

### 25.1.1 Why chromedp

`chromedp` drives a headless Chrome or Chromium instance from Go. It navigates to a URL, waits for the page to load, and prints to PDF using Chrome's built-in print engine. This approach gives you:

- **Modern CSS support**: Flexbox, Grid, custom properties, `@media print`, `@page`, `page-break-*` — everything Chrome supports.
- **Web fonts**: Load Google Fonts or embed base64 fonts in the template.
- **JavaScript**: If needed, client-side chart libraries can render into the PDF (not recommended for performance, but possible).
- **Pixel accuracy**: The same rendering engine as the browser the user is using.

### 25.1.2 Why Not wkhtmltopdf

wkhtmltopdf is abandoned software. Its last release used WebKit from 2018. It does not support Flexbox reliably, has security vulnerabilities with no patch path, and is no longer maintained. Do not use it.

### 25.1.3 Why Not a Pure Go PDF Library

Libraries like `fpdf` or `gofpdf` generate PDFs directly without an HTML rendering engine. They require you to manually calculate coordinates for every element. Tables that span multiple pages, headers that repeat, complex layouts — all must be implemented by hand. For ERP documents that need to match a business's letterhead precisely, this approach is impractical. Use the HTML template path.

### 25.1.4 chromedp Instance Management

Starting a Chrome process for every PDF request is too slow (Chrome start time is 200–500ms). Instead, Awo maintains a pool of long-lived headless Chrome instances:

```go
type ChromePool struct {
    allocator context.Context  // chromedp browser allocator
    cancel    context.CancelFunc
    semaphore chan struct{}     // Limits concurrent PDF renders
}

func NewChromePool(maxConcurrency int) *ChromePool {
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("no-sandbox", true),
        chromedp.Flag("disable-gpu", true),
        chromedp.Flag("disable-dev-shm-usage", true),
    )
    allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
    return &ChromePool{
        allocator: allocCtx,
        cancel:    cancel,
        semaphore: make(chan struct{}, maxConcurrency),
    }
}
```

Each PDF render acquires a semaphore slot, runs in a new browser context (tab), and releases the slot when done. The browser process itself stays alive between renders.

---

## 25.2 HTML Print Templates

### 25.2.1 Template System

Print templates use Go's `html/template` package. Each document type has its own template file:

```
templates/print/
  invoice.html
  delivery_note.html
  payslip.html
  purchase_order.html
  statement_of_account.html
  _layout.html         // Shared header, footer, CSS variables
  _base.css            // Base print stylesheet
```

The `_layout.html` template defines the outer document structure that all document templates extend:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .DocumentTitle }}</title>
    <style>
        {{ template "base_css" . }}
        {{ template "doc_css" . }}
    </style>
</head>
<body>
    <div class="document">
        {{ template "header" . }}
        {{ template "body" . }}
        {{ template "footer" . }}
    </div>
</body>
</html>
```

### 25.2.2 Template Data Binding

Each template receives a typed struct:

```go
type InvoicePrintData struct {
    // Tenant branding
    Tenant       *tenant.Tenant
    TenantConfig *config.TenantConfig
    LogoURL      string
    PrimaryColor string

    // Document data
    Invoice      *finance.SalesInvoice
    Lines        []*finance.SalesInvoiceLine
    Customer     *crm.Customer
    BillingAddr  *crm.Address
    ShippingAddr *crm.Address

    // Computed fields
    SubTotal     decimal.Decimal
    TaxTotal     decimal.Decimal
    GrandTotal   decimal.Decimal
    TaxLines     []TaxLine

    // Formatting helpers
    DateFormat   string  // From TenantConfig: "DD/MM/YYYY"
    Currency     string  // "KES"
}
```

The template accesses this data with standard Go template syntax:

```html
{{ define "body" }}
<table class="invoice-lines">
    <thead>
        <tr>
            <th>#</th>
            <th>Item</th>
            <th>Qty</th>
            <th>Unit Price</th>
            <th>Amount</th>
        </tr>
    </thead>
    <tbody>
        {{ range $i, $line := .Lines }}
        <tr>
            <td>{{ add $i 1 }}</td>
            <td>
                <strong>{{ $line.ItemName }}</strong>
                {{ if $line.Description }}
                <br><small>{{ $line.Description }}</small>
                {{ end }}
            </td>
            <td class="text-right">{{ $line.Qty }}</td>
            <td class="text-right">{{ formatCurrency $line.UnitPrice $.Currency }}</td>
            <td class="text-right">{{ formatCurrency $line.Amount $.Currency }}</td>
        </tr>
        {{ end }}
    </tbody>
    <tfoot>
        <tr class="subtotal-row">
            <td colspan="4" class="text-right">Sub-total</td>
            <td class="text-right">{{ formatCurrency .SubTotal .Currency }}</td>
        </tr>
        {{ range .TaxLines }}
        <tr>
            <td colspan="4" class="text-right">{{ .Name }} ({{ .Rate }}%)</td>
            <td class="text-right">{{ formatCurrency .Amount $.Currency }}</td>
        </tr>
        {{ end }}
        <tr class="grand-total-row">
            <td colspan="4" class="text-right"><strong>Total</strong></td>
            <td class="text-right"><strong>{{ formatCurrency .GrandTotal .Currency }}</strong></td>
        </tr>
    </tfoot>
</table>
{{ end }}
```

Template functions (`formatCurrency`, `add`, `formatDate`, etc.) are registered globally when the template set is parsed at startup.

### 25.2.3 Tenant Branding Injection

Every print template receives the tenant's branding from `TenantConfig`:

```go
func BuildInvoicePrintData(ctx context.Context, invoice *finance.SalesInvoice) (*InvoicePrintData, error) {
    cfg := config.FromContext(ctx)
    t   := tenant.FromContext(ctx)

    logoURL := cfg.GetString(ctx, "branding.logo_url")
    if logoURL == "" {
        logoURL = "/assets/default-logo.png"
    }

    return &InvoicePrintData{
        Tenant:       t,
        TenantConfig: cfg,
        LogoURL:      logoURL,
        PrimaryColor: cfg.GetString(ctx, "branding.primary_colour"),
        // ... load invoice, lines, customer ...
    }, nil
}
```

The CSS uses a custom property for the primary colour, set from the template data:

```html
<style>
:root {
    --primary: {{ .PrimaryColor }};
}
</style>
```

### 25.2.4 PrintTemplate Field on EntityDefinition

EntityDefinitions that support PDF output declare a `PrintTemplate` field:

```go
var SalesInvoiceDefinition = definition.EntityDefinition{
    Name:          "sales_invoice",
    // ...
    PrintTemplate: "invoice",  // Matches templates/print/invoice.html
    PrintDataBuilder: BuildInvoicePrintData,
}
```

If `PrintTemplate` is empty, the entity has no PDF endpoint. The framework will return 404 for `GET /api/v1/{entity}/{id}/pdf` if no print template is registered.

---

## 25.3 The PDF Endpoint

### 25.3.1 Route

```
GET /api/v1/{entity}/{id}/pdf
```

The route is auto-registered for any EntityDefinition with a non-empty `PrintTemplate` field.

### 25.3.2 Handler Flow

```go
func (h *PDFHandler) Handle(c *fiber.Ctx) error {
    entityName := c.Params("entity")
    id         := c.Params("id")

    def, ok := h.registry.Get(entityName)
    if !ok || def.PrintTemplate == "" {
        return fiber.ErrNotFound
    }

    // Permission check — same as read
    if err := h.permissions.Require(c, entityName, "read"); err != nil {
        return err
    }

    // Check cache first
    cacheKey := pdfCacheKey(c.TenantID(), entityName, id)
    if pdfBytes, ok := h.cache.Get(cacheKey); ok {
        return streamPDF(c, pdfBytes)
    }

    // Build template data
    data, err := def.PrintDataBuilder(c.UserContext(), id)
    if err != nil {
        return err
    }

    // Render HTML
    var buf bytes.Buffer
    if err := h.templates.ExecuteTemplate(&buf, def.PrintTemplate+".html", data); err != nil {
        return fmt.Errorf("template render: %w", err)
    }

    // Convert to PDF via chromedp
    pdfBytes, err := h.chrome.PrintToPDF(c.UserContext(), buf.String())
    if err != nil {
        return fmt.Errorf("pdf generation: %w", err)
    }

    // Store in object storage
    if err := h.storage.Put(cacheKey, pdfBytes); err != nil {
        h.log.Warn("pdf cache write failed", "key", cacheKey, "err", err)
        // Continue — we have the bytes, just serve without caching
    }

    return streamPDF(c, pdfBytes)
}

func streamPDF(c *fiber.Ctx, b []byte) error {
    c.Set("Content-Type", "application/pdf")
    c.Set("Content-Disposition", `inline; filename="document.pdf"`)
    c.Set("Content-Length", strconv.Itoa(len(b)))
    return c.Send(b)
}
```

### 25.3.3 chromedp Print Logic

```go
func (p *ChromePool) PrintToPDF(ctx context.Context, html string) ([]byte, error) {
    // Acquire semaphore slot
    select {
    case p.semaphore <- struct{}{}:
        defer func() { <-p.semaphore }()
    case <-ctx.Done():
        return nil, ctx.Err()
    }

    // Create a new browser tab
    tabCtx, cancel := chromedp.NewContext(p.allocator)
    defer cancel()

    // Load the HTML string directly as a data URL
    dataURL := "data:text/html;charset=utf-8," + url.QueryEscape(html)

    var pdfBuf []byte
    if err := chromedp.Run(tabCtx,
        chromedp.Navigate(dataURL),
        chromedp.WaitReady("body"),
        chromedp.ActionFunc(func(ctx context.Context) error {
            var err error
            pdfBuf, _, err = page.PrintToPDF().
                WithPrintBackground(true).
                WithPreferCSSPageSize(true).
                Do(ctx)
            return err
        }),
    ); err != nil {
        return nil, err
    }

    return pdfBuf, nil
}
```

`WithPreferCSSPageSize(true)` means the page dimensions declared in the CSS `@page` rule take precedence over any Chrome defaults. This is how A4 sizing is enforced.

---

## 25.4 CSS Page Layout

### 25.4.1 Page Size and Margins

```css
@page {
    size: A4 portrait;  /* 210mm × 297mm */
    margin: 15mm 20mm 20mm 20mm;  /* top right bottom left */
}

@page :first {
    margin-top: 10mm;  /* Less top margin on first page (logo takes space) */
}
```

### 25.4.2 Header and Footer via CSS

Chrome supports `@page` margin boxes for repeating headers and footers:

```css
@page {
    @top-right {
        content: "Page " counter(page) " of " counter(pages);
        font-size: 9pt;
        color: #666;
    }
    @bottom-center {
        content: "Confidential — {{ .Tenant.Name }}";
        font-size: 8pt;
        color: #999;
    }
}
```

The `counter(page)` and `counter(pages)` CSS variables are injected by the browser's print engine. No JavaScript is needed.

### 25.4.3 Page Breaks

Control where pages break within multi-page documents:

```css
/* Never break inside an invoice line table */
.invoice-lines tr {
    break-inside: avoid;
}

/* Always start a new section on a fresh page */
.new-section {
    break-before: page;
}

/* Repeat the table header on every page */
thead {
    display: table-header-group;
}
```

### 25.4.4 Fonts

System fonts are available in the headless Chrome environment, but they vary by OS. For consistent output across development (macOS), CI (Linux), and production (Linux), embed fonts as base64 in the template:

```css
@font-face {
    font-family: 'Inter';
    src: url('data:font/woff2;base64,AAABAA...') format('woff2');
    font-weight: 400;
}
```

The `cmd/embed-fonts` tool in the framework repository converts font files to base64 CSS rules. Run it when updating the font set. The output is committed to `templates/print/_fonts.css` and included in `_base.css`.

### 25.4.5 Complete Base Print CSS

```css
/* templates/print/_base.css */

*, *::before, *::after {
    box-sizing: border-box;
}

body {
    font-family: 'Inter', 'Helvetica Neue', Arial, sans-serif;
    font-size: 10pt;
    line-height: 1.4;
    color: #1a1a1a;
    background: white;
}

.document {
    width: 100%;
    max-width: 100%;
}

/* Tables */
table {
    width: 100%;
    border-collapse: collapse;
}
th {
    text-align: left;
    background: var(--primary, #1a56db);
    color: white;
    padding: 6pt 8pt;
    font-size: 9pt;
    font-weight: 600;
}
td {
    padding: 5pt 8pt;
    border-bottom: 0.5pt solid #e5e7eb;
    vertical-align: top;
}
.text-right {
    text-align: right;
}

/* Grand total row */
.grand-total-row td {
    border-top: 1.5pt solid #1a1a1a;
    border-bottom: 1.5pt solid #1a1a1a;
    font-size: 11pt;
}

/* Print-only: hide elements not needed in PDF */
.no-print {
    display: none !important;
}
```

---

## 25.5 PDF Caching

### 25.5.1 Cache Key

```go
func pdfCacheKey(tenantID uuid.UUID, entity, id string) string {
    return fmt.Sprintf("pdf/%s/%s/%s.pdf", tenantID, entity, id)
}
```

This key is used in both the object storage path and as the basis for the ETag header.

### 25.5.2 Cache Invalidation

When a record is updated, its cached PDF is stale. The `after_save` hook on any entity with a `PrintTemplate` deletes the cached PDF from object storage:

```go
func InvalidatePDFCache(ctx context.Context, record *EntityRecord) error {
    def := registry.MustGet(record.EntityName)
    if def.PrintTemplate == "" {
        return nil
    }
    key := pdfCacheKey(record.TenantID, record.EntityName, record.ID.String())
    return storage.Delete(ctx, key)
}
```

The next request for this record's PDF will regenerate it.

### 25.5.3 Serving from Cache

On cache hit, the handler checks if the object storage key exists and returns a redirect to a pre-signed URL (for S3-compatible storage):

```go
if url, err := h.storage.PresignedURL(cacheKey, 15*time.Minute); err == nil {
    return c.Redirect(url, fiber.StatusTemporaryRedirect)
}
```

If the storage layer doesn't support pre-signed URLs, proxy the bytes through the API handler instead.

---

## 25.6 Background PDF Generation

For bulk operations — payroll runs generating 200 payslips, month-end statements for 500 customers — generating PDFs synchronously in an HTTP handler is inappropriate. Use a Temporal activity.

### 25.6.1 GeneratePDFActivity

```go
type GeneratePDFInput struct {
    TenantID   uuid.UUID
    EntityName string
    RecordID   uuid.UUID
}

type GeneratePDFOutput struct {
    StorageKey string
    URL        string  // Pre-signed download URL
}

func (a *PrintActivities) GeneratePDFActivity(
    ctx context.Context,
    input GeneratePDFInput,
) (GeneratePDFOutput, error) {
    // Heartbeat for long-running batches
    activity.RecordHeartbeat(ctx, "starting")

    def := a.registry.MustGet(input.EntityName)
    data, err := def.PrintDataBuilder(ctx, input.RecordID.String())
    if err != nil {
        return GeneratePDFOutput{}, err
    }

    var buf bytes.Buffer
    if err := a.templates.ExecuteTemplate(&buf, def.PrintTemplate+".html", data); err != nil {
        return GeneratePDFOutput{}, err
    }

    activity.RecordHeartbeat(ctx, "rendering")
    pdfBytes, err := a.chrome.PrintToPDF(ctx, buf.String())
    if err != nil {
        return GeneratePDFOutput{}, err
    }

    key := pdfCacheKey(input.TenantID, input.EntityName, input.RecordID.String())
    if err := a.storage.Put(key, pdfBytes); err != nil {
        return GeneratePDFOutput{}, err
    }

    url, _ := a.storage.PresignedURL(key, 7*24*time.Hour)
    return GeneratePDFOutput{StorageKey: key, URL: url}, nil
}
```

### 25.6.2 Payroll Run PDF Workflow

A payroll run workflow generates all payslip PDFs in parallel using child workflows or activity batches:

```go
func PayrollPDFGenerationWorkflow(ctx workflow.Context, params PayrollPDFParams) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        HeartbeatTimeout:    30 * time.Second,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    // Generate PDFs for all payslips in the run
    var futures []workflow.Future
    for _, payslipID := range params.PayslipIDs {
        f := workflow.ExecuteActivity(ctx, activities.GeneratePDFActivity, GeneratePDFInput{
            TenantID:   params.TenantID,
            EntityName: "payslip",
            RecordID:   payslipID,
        })
        futures = append(futures, f)
    }

    // Wait for all PDFs
    var results []GeneratePDFOutput
    for _, f := range futures {
        var out GeneratePDFOutput
        if err := f.Get(ctx, &out); err != nil {
            return err
        }
        results = append(results, out)
    }

    // Bundle into a ZIP and notify
    workflow.ExecuteActivity(ctx, activities.BundlePDFsAndNotify,
        BundleInput{TenantID: params.TenantID, Files: results, NotifyUserID: params.RequestedBy})

    return nil
}
```

### 25.6.3 User Notification on Completion

When the bulk PDF generation workflow completes, it sends an in-app notification with a download link:

```go
func (a *PrintActivities) BundlePDFsAndNotify(ctx context.Context, input BundleInput) error {
    // Zip all PDFs
    zipKey := fmt.Sprintf("pdf/%s/payroll-run-%s.zip", input.TenantID, time.Now().Format("2006-01"))
    if err := a.bundleAndUpload(ctx, input.Files, zipKey); err != nil {
        return err
    }

    url, _ := a.storage.PresignedURL(zipKey, 7*24*time.Hour)

    // Send in-app notification
    return a.notifications.Send(ctx, &Notification{
        TenantID: input.TenantID,
        UserID:   input.NotifyUserID,
        Title:    "Payslips ready for download",
        Body:     fmt.Sprintf("%d payslips have been generated.", len(input.Files)),
        Type:     "success",
        Channel:  "in_app",
        Metadata: map[string]any{
            "download_url": url,
            "expires_in":   "7 days",
        },
    })
}
```
