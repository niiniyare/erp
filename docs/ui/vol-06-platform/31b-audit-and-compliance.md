---
title: "Audit and Compliance"
volume: "VI — Platform & API"
chapter: "31-B"
phase: 3
status: draft
audience: [Platform Engineers, Security Engineers, ERP Implementers]
---

# Chapter 31-B — Audit and Compliance

> **Volume:** VI — Platform & API
> **Audience:** Platform Engineers, Security Engineers, ERP Implementers
> **Prerequisites:** Chapter 26 — Security Model, Chapter 28-B — Platform/Tenant/Portal Views
> **Phase:** 3

---

## 31B.1 Compliance Scope for AwoERP UI Platform

AwoERP targets the East African market with Kenya as its primary jurisdiction. The UI platform's compliance obligations arise from multiple overlapping regulatory frameworks: the EU General Data Protection Regulation (GDPR) applies to tenants that process data of EU data subjects; the Kenya Data Protection Act 2019 applies to all tenants operating in Kenya; KRA eTIMS regulations apply to all invoicing activities in Kenya; and SOC 2 Type II is a target certification for the platform operator.

Beyond these regulatory requirements, AwoERP's position as a multi-tenant ERP means that **one tenant's compliance failure must never contaminate other tenants**. Audit logs, PII data, and compliance artefacts are strictly tenant-scoped. The platform's own audit logs (platform admin actions) are maintained separately from tenant business audit logs.

The UI layer has specific compliance obligations that are distinct from backend data compliance. These include: what data fields appear in UI payloads (PII minimisation), how audit events are generated for UI actions, how the KRA eTIMS QR code appears on printed invoices, and how user consent is collected for data processing activities.

---

## 31B.2 GDPR Considerations

### 31B.2.1 PII in UI Payloads — Identification and Minimisation

UI payloads (compiled surface definitions and their data responses) must not include PII that is not necessary for the surface's function. The following fields are classified as PII in AwoERP:

| Field | Classification | UI Minimisation Rule |
|-------|---------------|---------------------|
| National ID number | High sensitivity | Show only last 4 digits in list views: `****1234` |
| KRA PIN | Medium (business) | Show full in invoice surfaces; mask in list views |
| Phone number | Medium | Show for HR/contact surfaces; mask in report exports |
| Bank account number | High sensitivity | Never shown in UI (shown only during payment setup, then masked) |
| Employee salary | High sensitivity | Only visible on payslip surface to the employee themselves or HR admin |
| Date of birth | Medium | Only visible on HR profile surface |

The surface definition declares its PII fields in the `pii_fields` metadata array:

```json
{
  "surface_id": "hr.employee.profile",
  "pii_fields": ["national_id", "date_of_birth", "salary", "bank_account"],
  "pii_minimisation": {
    "national_id": { "mask": "last_4", "display_context": ["hr.admin", "hr.payroll"] }
  }
}
```

The compilation pipeline reads `pii_fields` and applies masking transforms based on the `ViewContext.Permissions` — if the requesting user does not have the listed permission, the field value is masked before being included in the surface data response.

### 31B.2.2 Right to Erasure — Impact on UI Definitions and Cached Payloads

When a data subject invokes their right to erasure (GDPR Article 17 / Kenya DPA Section 38), the following UI-layer artefacts must be handled:

1. **Cached compiled surfaces**: Any Redis-cached surface data containing the data subject's PII must be invalidated. The `UICompilationService` exposes a `InvalidateSubjectData(ctx, subjectID uuid.UUID)` method that scans and purges affected cache keys.

2. **Audit log entries**: Audit log entries that reference the data subject's identifier are anonymised (the `actor_id` field is replaced with a deterministic hash of the original ID) rather than deleted — audit logs must remain complete for SOC 2 evidence.

3. **Exported documents**: PDFs and CSV exports stored in object storage that contain PII must be deleted. A Temporal workflow handles this systematically:

```go
// internal/core/compliance/workflow/erasure.go
func (w *ErasureWorkflow) Execute(ctx workflow.Context, params ErasureParams) error {
    // 1. Mark records as erased in PostgreSQL (handled by data team)
    // 2. Invalidate UI caches
    err := workflow.ExecuteActivity(ctx, activities.InvalidateUICache, params.SubjectID).Get(ctx, nil)
    // 3. Delete stored exports
    err = workflow.ExecuteActivity(ctx, activities.DeleteStoredExports, params.SubjectID).Get(ctx, nil)
    // 4. Anonymise audit log references
    err = workflow.ExecuteActivity(ctx, activities.AnonymiseAuditLogs, params.SubjectID).Get(ctx, nil)
    // 5. Record erasure completion
    return workflow.ExecuteActivity(ctx, activities.RecordErasureCompletion, params).Get(ctx, nil)
}
```

### 31B.2.3 Data Retention for UI Audit Logs

UI audit logs (surface access events, action execution events) are retained for:
- 12 months in hot storage (PostgreSQL, queryable)
- 7 years in cold storage (compressed, object storage, for legal/tax requirements)

Retention is enforced by a nightly Temporal cron workflow that archives and purges records based on `created_at` timestamp.

---

## 31B.3 SOC 2 Controls

### 31B.3.1 Access Control Evidence via Navigation Pruning Logs

SOC 2 Trust Service Criteria CC6.1 requires evidence that access to systems is limited to authorised users. The navigation pruning audit log (Section 19B.11.2) provides this evidence: every navigation tree construction records which items were pruned and why (`permission_missing`, `flag_disabled`, `module_disabled`). This log is queryable by auditors.

The audit query surface `platform.audit.navigation-pruning` provides a compliance-ready view:

```sql
-- Query for auditors: show all navigation access decisions for a user in the audit period
SELECT
    generated_at,
    actor_id,
    portal,
    total_items,
    pruned_count,
    pruned_keys
FROM ui_audit_nav_construction
WHERE actor_id = $1
  AND generated_at BETWEEN $2 AND $3
ORDER BY generated_at DESC;
```

### 31B.3.2 Audit Trail Completeness

Every UI mutation action (form submit, button click that calls an API) generates an audit event via `deps.AuditService.Record()`. The audit record includes:

```go
deps.AuditService.Record(c.UserContext(), audit.Event{
    Type:       "ui.action.executed",
    ActorType:  viewCtx.ActorType,
    ActorID:    viewCtx.UserID,
    TenantID:   viewCtx.TenantID,
    SurfaceID:  surfaceID,
    ActionType: actionType,      // "form_submit", "button_click", "export"
    ResourceID: resourceID,
    Changes:    changedFields,   // field-level diff for mutations
    TraceID:    viewCtx.TraceID,
})
```

SOC 2 CC7.2 requires that all security events are logged. The `ui.action.executed` events for security-sensitive surfaces (user management, permission assignment, tenant configuration) are additionally flagged with `security_relevant: true` for faster auditor review.

### 31B.3.3 Change Management for UI Definitions

Changes to surface definitions (stored in `ui_surface_definitions`) are tracked in a `ui_surface_definition_versions` table:

```sql
CREATE TABLE ui_surface_definition_versions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID REFERENCES tenants(id),  -- NULL for platform surfaces
    surface_id   TEXT NOT NULL,
    version      INTEGER NOT NULL,
    definition   JSONB NOT NULL,
    changed_by   UUID REFERENCES users(id),
    change_note  TEXT,
    changed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Platform surface definition changes go through the standard Git + PR + CI pipeline, providing a complete change history in version control.

---

## 31B.4 Kenya Data Protection Act (DPA 2019) Considerations

### 31B.4.1 Data Localisation Requirements

The Kenya DPA 2019 requires that personal data about Kenyan data subjects be processed and stored in Kenya or in a jurisdiction with adequate data protection laws. AwoERP's production deployment for Kenya tenants uses AWS Africa (Cape Town) `af-south-1` or, when available, AWS Kenya region. The `storage.region` configuration option is validated at tenant provisioning time:

```go
var kenyaApprovedRegions = []string{"af-south-1", "eu-west-1", "eu-central-1"}

func validateDataResidency(region string) error {
    if !slices.Contains(kenyaApprovedRegions, region) {
        return fmt.Errorf("region %q is not approved for Kenya DPA data residency", region)
    }
    return nil
}
```

### 31B.4.2 Consent Management in UI

The Kenya DPA requires explicit consent for collection and processing of personal data. The employee self-service portal and customer portal display consent forms at first login:

```json
{
  "type": "form",
  "title": "Data Processing Consent",
  "api": "POST /api/v1/portal/consent",
  "body": [
    { "type": "html", "html": "<p>By using this portal, you consent to <a href='/privacy-policy'>our data processing policy</a>...</p>" },
    { "type": "checkbox", "name": "consent_given", "label": "I consent to data processing as described", "required": true,
      "requiredOn": true }
  ],
  "submitText": "Accept and Continue"
}
```

Consent records are stored in `portal_consent_records` with timestamp, IP address, and the policy version consented to.

---

## 31B.5 KRA eTIMS Integration Compliance

### 31B.5.1 Invoice UI Fields Required by eTIMS

KRA eTIMS mandates that every tax invoice display the following fields. These fields are validated at the surface compilation stage for any surface with `document_type: "tax_invoice"`:

| Field | Source | UI Display |
|-------|--------|------------|
| Seller KRA PIN | Tenant config | Always visible, non-editable |
| Buyer KRA PIN | Customer record | Required; validated format `A000000000A` |
| CU Invoice Number | eTIMS API response | Displayed after eTIMS signing |
| Invoice Date | Invoice record | Cannot be future-dated |
| Line items with VAT rate | Invoice lines | VAT breakdown required |
| Total excl. VAT | Computed | Must match sum of lines |
| VAT amount | Computed | Must match rate × excl. amount |
| Total incl. VAT | Computed | Must equal excl. + VAT |
| eTIMS verification QR code | eTIMS API | Must appear on print/PDF |
| eTIMS timestamp | eTIMS API | Unix timestamp of KRA signing |

The `UICompilationService` validates these fields are present in any `tax_invoice` surface definition at compile time — missing required eTIMS fields cause a compilation error, not a runtime validation error.

### 31B.5.2 QR Code Display in Invoice Surface

The QR code must appear in the **upper right corner** of every page of the printed invoice (this is a KRA mandate, not an aesthetic choice):

```json
{
  "type": "container",
  "className": "etims-qr-block",
  "style": { "position": "absolute", "top": "10mm", "right": "10mm" },
  "display_modes": ["print"],
  "children": [
    { "type": "image", "src": "${document.etims_qr_url}", "width": 80, "height": 80,
      "alt": "KRA eTIMS Verification QR Code" },
    { "type": "tpl", "tpl": "<small>CU: ${document.etims_cu_invoice_number}</small>" }
  ]
}
```

> **⚠ Warning:** An invoice surface must display the QR code only after eTIMS signing is complete. If `document.etims_status !== 'SIGNED'`, the invoice print must be blocked or watermarked "UNVERIFIED". See Chapter 12-B Section 12B.9 for the full eTIMS PDF pipeline.

---

## 31B.6 UI Audit Log Schema

The `ui_audit_events` table stores all UI-layer audit events:

```sql
CREATE TABLE ui_audit_events (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        UUID NOT NULL REFERENCES tenants(id),
    event_type       TEXT NOT NULL,       -- see EventType enum below
    actor_type       TEXT NOT NULL,       -- platform_admin | tenant_staff | portal_*
    actor_id         UUID NOT NULL,
    portal_type      TEXT,               -- NULL for staff events
    surface_id       TEXT,
    action_type      TEXT,               -- form_submit | button_click | export | print | nav_click
    resource_type    TEXT,               -- invoice | po | employee | etc.
    resource_id      UUID,
    http_method      TEXT,
    http_path        TEXT,
    http_status      INTEGER,
    duration_ms      INTEGER,
    trace_id         TEXT NOT NULL,
    security_relevant BOOLEAN DEFAULT FALSE,
    metadata         JSONB,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Partition by month for performance
CREATE INDEX ui_audit_events_tenant_created ON ui_audit_events (tenant_id, created_at DESC);
CREATE INDEX ui_audit_events_actor ON ui_audit_events (actor_id, created_at DESC);
CREATE INDEX ui_audit_events_trace ON ui_audit_events (trace_id);
```

### 31B.6.1 UI Surface Access Events

Generated when a user loads a surface:
```json
{
  "event_type": "ui.surface.accessed",
  "actor_type": "tenant_staff",
  "surface_id": "finance.invoices.list",
  "http_method": "GET",
  "http_path": "/api/v1/ui/surfaces/finance.invoices.list",
  "http_status": 200,
  "duration_ms": 45
}
```

### 31B.6.2 Action Execution Events

Generated when a user submits a form or executes a mutation action:
```json
{
  "event_type": "ui.action.executed",
  "actor_type": "tenant_staff",
  "action_type": "form_submit",
  "surface_id": "finance.invoices.new",
  "resource_type": "invoice",
  "resource_id": "new-invoice-uuid",
  "http_method": "POST",
  "http_path": "/api/v1/finance/invoices",
  "http_status": 201,
  "security_relevant": false,
  "metadata": { "invoice_number": "INV-2026-001", "amount": 150000 }
}
```

### 31B.6.3 Navigation Pruning Decision Logs

```json
{
  "event_type": "ui.navigation.constructed",
  "actor_type": "tenant_staff",
  "metadata": {
    "portal": "staff",
    "total_items": 24,
    "pruned_count": 3,
    "pruned_keys": [
      { "key": "hr.payroll.reports", "reason": "permission_missing", "permission": "hr.payroll.read" },
      { "key": "finance.settings", "reason": "permission_missing", "permission": "finance.settings.admin" }
    ]
  }
}
```

---

## 31B.7 Compliance Review Checklist for New UI Surfaces

Every new UI surface must pass this checklist before being promoted to production:

```markdown
## UI Surface Compliance Checklist

**Surface ID**: `_______________________________`
**Reviewed by**: `_______________________________`
**Date**: `_______________________________`

### Data Protection
- [ ] PII fields identified and `pii_fields` metadata declared
- [ ] PII masking rules defined for list/summary views
- [ ] Data export (CSV/XLSX/PDF) does not include unexpected PII
- [ ] Surface is tenant-scoped (no cross-tenant data exposure possible)

### Authorization
- [ ] `required_permission` set on surface definition
- [ ] All mutation actions have `visibleOn` permission checks
- [ ] Audit `deps.AuditService.Record()` called after every mutation

### KRA eTIMS (for invoice/tax document surfaces only)
- [ ] All required eTIMS fields present in surface definition
- [ ] QR code displayed in mandated position on print layout
- [ ] `etims_status === 'SIGNED'` guard applied before PDF generation
- [ ] Unsigned invoice watermark active

### Logging & Observability
- [ ] Prometheus metrics instrument the new endpoint
- [ ] OpenTelemetry span created for surface fetch handler
- [ ] Error responses include `trace_id`

### Accessibility
- [ ] All form fields have associated `label` elements
- [ ] Images have `alt` text
- [ ] Colour contrast ≥ 4.5:1 for normal text (WCAG AA)

### Portal Actor Restrictions (if portal-accessible)
- [ ] Surface `allowed_actor_types` list declared
- [ ] Portal user can only see their own records (CEL rule validated)
- [ ] No internal ERP data exposed through portal surface
```

> See Chapter 42 — CI/CD Considerations for how this checklist is partially automated in the CI pipeline via `awo ui validate --compliance`.
