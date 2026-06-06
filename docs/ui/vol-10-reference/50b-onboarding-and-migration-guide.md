---
title: "Onboarding and Migration Guide for Implementers"
volume: "X — Reference & Guidance"
chapter: "50-B"
phase: 4
status: draft
audience: [ERP Implementers, Solution Architects]
---

# Chapter 50-B — Onboarding and Migration Guide for Implementers

> **Volume:** X — Reference & Guidance
> **Audience:** ERP Implementers, Solution Architects
> **Prerequisites:** Chapter 28-B — Platform/Tenant/Portal Views, Chapter 19-B — Backend-Driven Navigation
> **Phase:** 4

---

## 50B.1 Onboarding Overview

Onboarding a new tenant to AwoERP is a structured process that moves from platform provisioning through module configuration to go-live. For a typical Kenya SME (a fuel station, a logistics company, a retail chain), the full onboarding process takes 2–5 days of implementer time, spread over a 2–4 week period to accommodate data migration, user training, and parallel running.

The onboarding process has seven stages:

1. **Pre-onboarding**: Infrastructure readiness, data collection, project kick-off
2. **Tenant provisioning**: Create tenant record, assign modules, configure environment
3. **Navigation and theme**: Configure navigation, upload branding, set portal configurations
4. **Data migration**: Import chart of accounts, opening balances, master data
5. **User onboarding**: Create users, assign roles, train key users
6. **Portal activation**: Enable supplier/customer/employee portals, onboard portal users
7. **Go-live**: Parallel run, cutover, hypercare

---

## 50B.2 Pre-Onboarding Checklist

Before tenant provisioning begins, collect and verify the following:

```markdown
## Pre-Onboarding Checklist

### Business Information
- [ ] Legal company name (exact, for KRA eTIMS compliance)
- [ ] KRA PIN (validated against KRA portal)
- [ ] VAT Registration Number (if VAT-registered)
- [ ] Physical address (for letterhead and KRA records)
- [ ] P.O. Box and postal code
- [ ] Primary contact phone number
- [ ] Primary business email

### Technical Requirements
- [ ] Confirmed subscription plan and module set
- [ ] Data residency region confirmed (Kenya DPA compliance)
- [ ] Custom domain configured (e.g. erp.clientcompany.co.ke) and DNS delegated
- [ ] SSL certificate issued for custom domain
- [ ] M-Pesa Paybill / Till Number provided (if customer portal payments enabled)
- [ ] Bank details for payment instructions on invoices

### Branding Assets
- [ ] Company logo (PNG, min 200×200px, transparent background preferred)
- [ ] Favicon (ICO or PNG, 32×32px)
- [ ] Primary brand colour (hex code)
- [ ] Accent colour (hex code, optional)
- [ ] Preferred font (Google Fonts name, optional)

### Legacy System Data
- [ ] Chart of accounts (Excel or CSV)
- [ ] Opening balances as of cutover date
- [ ] Customer master data (name, email, phone, KRA PIN, credit terms)
- [ ] Supplier master data
- [ ] Employee master data (for HR module)
- [ ] Historical invoices for carry-forward (optional)
```

---

## 50B.3 Tenant Provisioning Steps

### 50B.3.1 Tenant Creation via Platform Admin Surface

Using the platform admin surface or the API directly:

```bash
# Create tenant via API
curl -X POST https://app.awoerp.com/api/v1/platform/tenants \
  -H "Authorization: Bearer $PLATFORM_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "legal_name":        "Shell Maanzoni Limited",
    "trading_name":      "Shell Maanzoni",
    "kra_pin":           "P051234567A",
    "vat_reg_number":    "V01234567A",
    "address_line1":     "Shell Maanzoni Service Station",
    "address_line2":     "Mombasa-Nairobi Highway, Maanzoni",
    "city":              "Machakos",
    "country":           "KE",
    "phone":             "+254712345678",
    "email":             "admin@shellmaanzoni.co.ke",
    "company_size":      "SMALL",
    "subscription_plan": "PROFESSIONAL"
  }'
```

**Response:**
```json
{
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "legal_name": "Shell Maanzoni Limited",
    "status": "PENDING",
    "created_at": "2026-06-06T09:00:00Z"
  }
}
```

Activate the tenant to move from PENDING to ACTIVE:

```bash
curl -X POST https://app.awoerp.com/api/v1/platform/tenants/a1b2c3d4.../activate \
  -H "Authorization: Bearer $PLATFORM_ADMIN_TOKEN"
```

### 50B.3.2 Module Enablement

Enable the modules for this tenant based on their subscription:

```bash
curl -X PUT https://app.awoerp.com/api/v1/platform/tenants/a1b2c3d4.../modules \
  -H "Authorization: Bearer $PLATFORM_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled_modules": [
      "finance",
      "hr",
      "procurement",
      "fuel",
      "shop",
      "restaurant",
      "lpg",
      "spares"
    ]
  }'
```

After enabling modules, the navigation service automatically includes the corresponding navigation groups in the tenant's navigation tree. Verify with:

```bash
# Get navigation tree for tenant admin
curl "https://app.awoerp.com/api/v1/ui/navigation?portal=staff" \
  -H "Authorization: Bearer $TENANT_ADMIN_TOKEN" \
  -H "X-Tenant-ID: a1b2c3d4..."
```

### 50B.3.3 Navigation Configuration

Customise the navigation for the tenant's specific terminology and workflow preferences:

```bash
curl -X PUT https://app.awoerp.com/api/v1/tenants/a1b2c3d4.../navigation \
  -H "Authorization: Bearer $TENANT_ADMIN_TOKEN" \
  -H "X-Tenant-ID: a1b2c3d4..." \
  -H "Content-Type: application/json" \
  -d '{
    "portal": "staff",
    "overrides": {
      "labels": {
        "procurement": "Purchasing",
        "procurement.po.list": "Purchase Orders",
        "fuel.deliveries.list": "Fuel Truck Deliveries"
      },
      "reorder": {
        "fuel": 1,
        "finance": 7,
        "shop": 2
      },
      "hidden": [],
      "extra_items": [
        {
          "key": "external.mpesa-dashboard",
          "label": "M-Pesa Business",
          "icon": "fa fa-mobile",
          "external_url": "https://business.safaricom.co.ke",
          "open_in": "new_tab",
          "order": 99
        }
      ]
    }
  }'
```

### 50B.3.4 Theme and Branding Upload

```bash
# Upload logo
curl -X POST https://app.awoerp.com/api/v1/tenants/a1b2c3d4.../theme/logo \
  -H "Authorization: Bearer $TENANT_ADMIN_TOKEN" \
  -H "X-Tenant-ID: a1b2c3d4..." \
  -F "file=@/path/to/shell-maanzoni-logo.png"

# Set brand colours and font
curl -X PATCH https://app.awoerp.com/api/v1/tenants/a1b2c3d4.../theme \
  -H "Authorization: Bearer $TENANT_ADMIN_TOKEN" \
  -H "X-Tenant-ID: a1b2c3d4..." \
  -H "Content-Type: application/json" \
  -d '{
    "tokens": {
      "--colors-brand-primary":       "#DD1700",
      "--colors-brand-primary-light": "#E53935",
      "--colors-brand-primary-dark":  "#B71C1C",
      "--colors-brand-accent":        "#FFCC00"
    }
  }'
```

### 50B.3.5 Portal Activation

Activate each portal the tenant will use:

```bash
# Activate Supplier Portal
curl -X POST https://app.awoerp.com/api/v1/tenants/a1b2c3d4.../portals/supplier/activate \
  -H "Authorization: Bearer $TENANT_ADMIN_TOKEN" \
  -H "X-Tenant-ID: a1b2c3d4..." \
  -H "Content-Type: application/json" \
  -d '{
    "subdomain": "suppliers.shellmaanzoni.co.ke",
    "welcome_message": "Welcome to Shell Maanzoni Supplier Portal",
    "default_locale": "en-KE"
  }'

# Activate Customer Portal
curl -X POST https://app.awoerp.com/api/v1/tenants/a1b2c3d4.../portals/customer/activate \
  -H "Authorization: Bearer $TENANT_ADMIN_TOKEN" \
  -H "X-Tenant-ID: a1b2c3d4..." \
  -H "Content-Type: application/json" \
  -d '{
    "subdomain": "pay.shellmaanzoni.co.ke",
    "mpesa_paybill": "123456",
    "default_locale": "en-KE"
  }'
```

---

## 50B.4 Data Migration for UI Customisations

When migrating a tenant from a previous AwoERP installation or from a different platform, any UI customisations (tenant-specific surface overrides, saved fragments, navigation overrides) must be migrated separately from business data.

### Export from source installation:

```bash
# Export UI customisations
curl "https://old.awoerp.com/api/v1/tenants/old-tenant-id/ui/export" \
  -H "Authorization: Bearer $OLD_PLATFORM_ADMIN_TOKEN" \
  -o ui-customisations.json

# Export theme
curl "https://old.awoerp.com/api/v1/tenants/old-tenant-id/theme/export" \
  -H "Authorization: Bearer $OLD_PLATFORM_ADMIN_TOKEN" \
  -o theme-bundle.json
```

### Import to target installation:

```bash
# Import theme
curl -X POST "https://new.awoerp.com/api/v1/tenants/new-tenant-id/theme/import" \
  -H "Authorization: Bearer $PLATFORM_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d @theme-bundle.json

# Import UI customisations
curl -X POST "https://new.awoerp.com/api/v1/tenants/new-tenant-id/ui/import" \
  -H "Authorization: Bearer $PLATFORM_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d @ui-customisations.json
```

> **⚠ Warning:** UI customisation imports must be validated against the target installation's template versions. If the source installation used template version 1 and the target is on version 2 with breaking changes, the import will fail validation and report which customisations need manual remediation.

---

## 50B.5 User Onboarding Flows

Users are created via the IAM module. For a typical fuel station, create these role groups:

```bash
# Create the initial tenant admin user
curl -X POST "https://app.awoerp.com/api/v1/iam/users" \
  -H "X-Tenant-ID: a1b2c3d4..." \
  -H "Authorization: Bearer $TENANT_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name":  "John Mwangi",
    "email":      "john.mwangi@shellmaanzoni.co.ke",
    "phone":      "+254712345678",
    "roles":      ["tenant_admin"],
    "send_invite": true
  }'
```

New users receive an email invite with a one-time password setup link. The initial password must be changed on first login.

**Recommended role assignments for a fuel station:**

| User type | Roles |
|-----------|-------|
| Station Manager | `tenant_admin`, `operations_manager`, `finance_manager` |
| Finance Manager | `finance_manager` |
| Senior Attendant | `operations_supervisor`, `shop_operator` |
| Pump Attendant | `pump_attendant` |
| Cashier | `shop_operator`, `restaurant_operator` |

---

## 50B.6 Migration from Legacy ERP UI

### 50B.6.1 Field Mapping from Legacy to AwoERP Components

Common field mapping patterns when migrating from legacy ERPs (QuickBooks, Sage, Tally):

| Legacy Concept | AwoERP Component | Migration Note |
|----------------|-----------------|----------------|
| Report screen | `service` + `table` | Port column definitions; add `sortable` |
| Data entry grid | `form` + `input-*` | Map validation rules to amis `validations` |
| Dropdown list | `select` + `source` API | Move hardcoded options to a data API |
| Print preview | `printable: true` surface | See Chapter 12-B for print template setup |
| Dashboard widget | `statistic` or `chart` | Connect to aggregation API endpoint |
| Multi-step wizard | `wizard` | Map each step to a form `body` array |

### 50B.6.2 Print Template Migration

Legacy ERP print templates (Crystal Reports, RDLC, Word templates) must be recreated as AwoERP print surface definitions. The migration process:

1. **Inventory existing templates**: List all print documents (invoices, POs, reports, payslips)
2. **Map fields**: For each field in the legacy template, identify the corresponding AwoERP API field
3. **Recreate as AST**: Build the print surface JSON definition using `platform.print-letterhead` template
4. **Validate eTIMS compliance**: Run `awo ui validate --compliance` to check KRA-required fields
5. **Test PDF output**: Generate test PDFs and compare with legacy output

```bash
# Validate a migrated invoice template for eTIMS compliance
awo ui validate --compliance --document-type tax_invoice \
  web/schemas/pages/finance/invoice-print.json
```

---

## 50B.7 Go-Live Checklist

```markdown
## Go-Live Checklist — AwoERP Tenant Onboarding

**Tenant**: Shell Maanzoni Limited
**Go-Live Date**: ___________
**Implementer**: ___________

### Infrastructure
- [ ] Custom domain DNS resolves correctly
- [ ] SSL certificate valid and auto-renewing
- [ ] Database backups configured and tested
- [ ] Monitoring alerts configured (PagerDuty/Slack)

### Data Completeness
- [ ] Chart of accounts imported and reviewed
- [ ] Opening balances entered and signed off by Finance Manager
- [ ] Customer master data imported (min. 90% of active customers)
- [ ] Supplier master data imported
- [ ] Employee records imported (if HR module active)
- [ ] Fuel grade and tank configuration complete (EPRA grades verified)

### KRA eTIMS
- [ ] KRA PIN configured and validated
- [ ] eTIMS credentials provided and tested
- [ ] Test invoice signed and QR code verified on KRA portal
- [ ] Print template reviewed and approved by client

### User Readiness
- [ ] All users created and invite emails accepted
- [ ] Role assignments reviewed and approved
- [ ] Key users completed training (min. 2 hours hands-on)
- [ ] Quick reference guides distributed

### Portal Readiness (if applicable)
- [ ] Supplier portal tested with at least 1 supplier account
- [ ] Customer portal tested with at least 1 customer account (M-Pesa payment tested)
- [ ] Employee portal tested with at least 5 employee accounts

### Branding
- [ ] Logo displays correctly in browser and print
- [ ] Theme colours match client brand guidelines
- [ ] Invoice letterhead approved by client

### Sign-off
- [ ] Client sign-off on go-live readiness
- [ ] Hypercare support plan confirmed (2 weeks post go-live)
- [ ] Rollback plan documented
```
