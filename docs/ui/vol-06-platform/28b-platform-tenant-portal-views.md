---
title: "Platform / Tenant / Portal View Differentiation"
volume: "VI — Platform & API"
chapter: "28-B"
phase: 3
status: draft
audience: [Platform Engineers, Solution Architects, Backend Engineers, Product Owners]
---

# Chapter 28-B — Platform / Tenant / Portal View Differentiation

> **Volume:** VI — Platform & API
> **Audience:** Platform Engineers, Solution Architects, Backend Engineers, Product Owners
> **Prerequisites:** Chapter 28 — Multi-Tenant Architecture, Chapter 26 — Security Model, Chapter 19-B — Backend-Driven Navigation
> **Phase:** 3

---

## 28B.1 The Three-Layer View Model

AwoERP serves **fundamentally different UI experiences** to fundamentally different actors — the operator running the platform, a business using the ERP, and external parties accessing a business's self-service portals — all from a single compiled SDUI backend. Understanding this three-layer model is prerequisite to understanding every design decision about navigation, surface access, data scoping, and portal configuration.

```
┌─────────────────────────────────────────────────────────────────────┐
│                     LAYER 1: PLATFORM                               │
│  AWO/operator admin view. Cross-tenant. No RLS. Manages tenants,    │
│  global flags, billing, platform health.                            │
│  Actor: platform_admin                                              │
└───────────────────────────┬─────────────────────────────────────────┘
                            │  provisions ↓
┌───────────────────────────▼─────────────────────────────────────────┐
│                     LAYER 2: TENANT                                  │
│  Single-tenant ERP view. Full RLS. Role-based surface access.        │
│  Actor: tenant_staff (many roles: admin, manager, clerk, operator)  │
│  Example: Shell Maanzoni fuel station staff                          │
└───────────────────────────┬─────────────────────────────────────────┘
                            │  exposes portals ↓
┌───────────────────────────▼─────────────────────────────────────────┐
│                     LAYER 3: PORTALS                                 │
│  Tenant-scoped sub-applications for external actors. Separate        │
│  identity. Minimal navigation. Permission-constrained surfaces.      │
│  Actors: portal_supplier | portal_customer | portal_employee         │
└─────────────────────────────────────────────────────────────────────┘
```

### 28B.1.1 Layer 1 — Platform Layer (AWO Operator View)

The platform layer is accessed only by AWO employees operating the multi-tenant infrastructure. It provides cross-tenant visibility: a single view of all tenants, their module configurations, subscription status, and system health. No tenant RLS applies — the platform admin can see tenant records, but specific tenant business data (invoices, employees, transactions) remains scoped and is never visible through the platform layer.

### 28B.1.2 Layer 2 — Tenant Layer

The tenant layer is the core ERP experience. A business (a fuel station, a logistics company, a SACCO) accesses the full ERP through this layer with all modules enabled for their subscription. Every user in this layer is a tenant employee or system user. PostgreSQL RLS ensures they can only access data belonging to their tenant.

### 28B.1.3 Layer 3 — Portal Layer

Portal users are not tenant employees. They are **external actors** — a fuel supplier delivering to a station, a corporate customer paying an invoice, an employee checking their payslip from home. Portal users have a separate identity (separate login, separate JWT claims, separate session table), access only a narrow subset of ERP data, and cannot navigate to internal ERP surfaces.

### 28B.1.4 Isolation Guarantees Between Layers

- Platform admin cannot access tenant business data records
- Tenant staff cannot access other tenants' data (RLS enforcement)
- Portal users cannot access internal tenant staff surfaces
- Portal users from one tenant cannot access another tenant's portal
- Layer membership is encoded in the JWT `actor_type` claim and validated on every request

### 28B.1.5 How RLS and Casbin Enforce Layer Boundaries

PostgreSQL RLS enforces **tenant data isolation** (Layer 2 and 3 cannot see other tenants' data). Casbin enforces **surface and action access** (which surfaces a given `actor_type` can access, which actions they can execute). CEL expressions on Casbin rules enable fine-grained attribute-based access (e.g. an employee portal user can only see their own payslips).

---

## 28B.2 Platform (Super-Admin) View

### 28B.2.1 Scope: Cross-Tenant Visibility

Platform admin surfaces use a special database connection that bypasses tenant RLS. This is achieved by connecting as the `platform_admin` PostgreSQL role:

```go
// internal/core/platform/store.go
func WithPlatformAdmin(ctx context.Context, q db.Queries, fn func(*db.Queries) error) error {
    // Uses platform_admin DB role — RLS policies do not apply
    return fn(q.WithRole("platform_admin"))
}
```

> **⚠ Warning:** The `platform_admin` DB role bypasses all RLS. Never use it in tenant-scoped handlers. It is exclusively for platform management operations: tenant provisioning, billing, global configuration.

### 28B.2.2 Platform Admin Navigation Tree

Platform admin navigation is minimal — it covers only platform management surfaces:

```json
{
  "portal": "platform_admin",
  "groups": [
    { "key": "platform", "label": "Platform", "items": [
      { "key": "platform.tenants", "label": "Tenants", "surface_id": "platform.tenants.list" },
      { "key": "platform.billing", "label": "Billing", "surface_id": "platform.billing.list" },
      { "key": "platform.flags", "label": "Feature Flags", "surface_id": "platform.flags.list" },
      { "key": "platform.audit", "label": "Audit Log", "surface_id": "platform.audit.list" },
      { "key": "platform.health", "label": "System Health", "surface_id": "platform.health.dashboard" }
    ]}
  ]
}
```

### 28B.2.3 Tenant Management Surface

The tenant management surface provides full CRUD over tenants plus lifecycle management actions (PENDING → ACTIVE → SUSPENDED → ARCHIVED):

```json
{
  "type": "crud",
  "api": "GET /api/v1/platform/tenants",
  "columns": [
    { "name": "legal_name", "label": "Tenant Name" },
    { "name": "status", "label": "Status", "type": "tag" },
    { "name": "plan", "label": "Plan" },
    { "name": "enabled_modules", "label": "Modules", "type": "json" },
    { "name": "created_at", "label": "Created", "type": "datetime" }
  ]
}
```

### 28B.2.7 Security: Platform Admin Identity Verification

Platform admins must authenticate with hardware MFA (FIDO2/WebAuthn). The login flow for `platform_admin` actors uses a separate endpoint `/api/v1/platform/auth/login` that enforces MFA and issues platform-scoped JWTs with `actor_type: "platform_admin"` and a 2-hour maximum session.

### 28B.2.8 Platform Admin amis Surface Schema Example

```json
{
  "type": "page",
  "title": "Tenant Management",
  "toolbar": [
    { "type": "button", "label": "Provision New Tenant", "actionType": "dialog",
      "dialog": { "type": "dialog", "title": "Provision Tenant",
        "body": { "type": "form", "api": "POST /api/v1/platform/tenants" }
      }
    }
  ],
  "body": {
    "type": "crud",
    "api": "GET /api/v1/platform/tenants",
    "quickSaveApi": "PATCH /api/v1/platform/tenants/${id}"
  }
}
```

---

## 28B.3 Tenant (Staff) View

### 28B.3.1 Scope: Single-Tenant, Full ERP Module Access

The tenant staff view provides access to all ERP modules enabled for the tenant, subject to the user's role permissions. A Finance Manager sees the Finance module fully; a Pump Attendant at a fuel station sees only the Fuel Sales surface and cannot navigate to HR or Finance.

### 28B.3.2 Tenant Admin Sub-View vs. Regular Staff View

Tenant admins (users with `tenant.admin` role) see an additional "Settings" navigation group that provides access to: user management, role management, module configuration, theme settings, portal activation, and navigation customisation. Regular staff do not see the Settings group — it is pruned from their navigation tree.

### 28B.3.3 Module Enablement and Navigation Reduction

The `tenant.enabled_modules` array controls which navigation groups are present in the staff view. For a tenant with only Finance and HR enabled:

```go
enabledModules := []string{"finance", "hr"}
// fuel, procurement, shop, restaurant, lpg, spares groups → pruned from navigation
```

### 28B.3.4 Role-Based Surface Differentiation

**Operations Manager (Shell Maanzoni):**
Sees full Fuel Operations group (all sub-items), Dashboard with operations KPIs, Shop & Convenience (read/write), LPG (read/write). Cannot see Finance module (separate Finance Manager role).

**Pump Attendant (Shell Maanzoni):**
Sees only: "Today's Pump Sessions" surface. No access to inventory, deliveries, financials, or management surfaces. The `fuel.sales.today` surface shows a simplified view with only their assigned pumps.

**Finance Manager:**
Sees Finance module (all sub-items), Dashboard (finance KPIs), Reports. Cannot see operational surfaces (fuel sales, pump readings).

This differentiation is achieved purely through Casbin role policies + navigation pruning. No separate surface definitions exist per role — the same surface is compiled differently based on the `ViewContext.Permissions` set:

```go
// Finance surface "finance.invoices.list" compiled for Finance Manager vs Pump Attendant:
// Finance Manager:    approveAction, editAction, deleteAction, exportAction visible
// Pump Attendant:     receives HTTP 403 on surface fetch — no invoice permission at all
```

### 28B.3.5 Tenant Branding Application

In the staff view, the tenant's full branding is applied: logo in the sidebar, primary colour for buttons and active navigation state, custom font if configured. The `window.__AWO_BOOTSTRAP__.themeTokens` contains the tenant's resolved token set.

### 28B.3.6 Tenant Staff View amis Surface Schema Example

```json
{
  "type": "app",
  "brandName": "${tenantName}",
  "logo": "${tenantLogoUrl}",
  "header": {
    "brand": "${tenantName}",
    "children": [
      { "type": "nav", "links": [
        { "label": "${currentUserName}", "icon": "fa fa-user",
          "children": [
            { "label": "Profile", "href": "/app/profile" },
            { "label": "Logout", "href": "/api/v1/auth/logout" }
          ]
        }
      ]}
    ]
  },
  "pages": "${navPages}"
}
```

---

## 28B.4 Portal Views (External Actor Access)

### 28B.4.1 Portal Architecture Overview

Portals are **tenant-scoped sub-applications**. They share the tenant's branding and data, but are accessed by external parties through a separate identity system. Key architectural properties:

- Portal users authenticate at `/portal/{portal_type}/auth/login`, not the staff login endpoint
- Portal JWTs carry `actor_type`, `tenant_id`, `portal_id`, and `portal_user_id` claims
- Portal users are stored in a separate `portal_users` table, not the `users` table
- Portal sessions are shorter-lived (8 hours for supplier/customer, 12 hours for employee)
- Portal navigation is minimal and curated — no access to internal ERP surfaces

### 28B.4.2 Supplier Portal

**Scope**: Purchase order acknowledgement, invoice submission, GRN confirmation, account statement view.

**Permission model**: `portal.supplier.*` CEL-scoped — all checks include `actor.supplier_id == document.supplier_id` to ensure a supplier only sees their own documents.

**amis Surface Schema Example (Supplier PO List)**:
```json
{
  "type": "page",
  "title": "My Purchase Orders",
  "body": {
    "type": "crud",
    "api": "GET /api/v1/portal/supplier/purchase-orders",
    "columns": [
      { "name": "po_number", "label": "PO Number" },
      { "name": "order_date", "label": "Date", "type": "date" },
      { "name": "total_amount", "label": "Amount (KES)", "type": "number" },
      { "name": "status", "label": "Status", "type": "tag" },
      { "name": "actions", "type": "operation", "buttons": [
        { "label": "View", "actionType": "link", "link": "/portal/supplier/po/${id}" },
        { "label": "Acknowledge", "actionType": "ajax", "api": "POST /api/v1/portal/supplier/purchase-orders/${id}/acknowledge",
          "visibleOn": "status === 'PENDING_ACKNOWLEDGEMENT'" }
      ]}
    ]
  }
}
```

### 28B.4.3 Customer Portal

**Scope**: Sales order status, invoice payment, statement download, support tickets.

**M-Pesa Payment Integration**: The customer portal payment surface is the primary M-Pesa STK Push integration point. Kenyan customers pay invoices directly from the portal using their M-Pesa number.

**amis Surface Schema Example (Customer Invoice with M-Pesa)**:
```json
{
  "type": "page",
  "title": "Invoice ${invoice_number}",
  "body": [
    { "type": "service", "api": "GET /api/v1/portal/customer/invoices/${id}",
      "body": [
        { "type": "static", "name": "amount_due", "label": "Amount Due (KES)" },
        { "type": "static", "name": "due_date", "label": "Due Date" },
        { "type": "divider" },
        {
          "type": "form",
          "title": "Pay via M-Pesa",
          "api": "POST /api/v1/payments/mpesa/stk-push",
          "visibleOn": "status !== 'PAID'",
          "body": [
            { "type": "input-text", "name": "phone", "label": "M-Pesa Phone Number", "required": true,
              "validations": { "matchRegexp": "^(07|01)[0-9]{8}$" },
              "validationErrors": { "matchRegexp": "Enter a valid Kenyan phone number" }
            }
          ],
          "submitText": "Send M-Pesa Request",
          "onSuccess": { "actionType": "toast", "args": { "msg": "Check your phone for the M-Pesa prompt", "msgType": "info" } }
        }
      ]
    }
  ]
}
```

### 28B.4.4 Employee Self-Service Portal

**Scope**: Leave applications, payslip view, expense claims, personal details.

**Permission model**: `portal.employee.*` CEL-scoped — all checks include `actor.employee_id == record.employee_id`. Employees can only see and submit their own records.

**Integration with HR Module**: Portal employee actions (leave approval requests, expense claims) create records in the HR module's tables but are visible to both the portal user (status updates) and internal HR staff (approval workflows).

**amis Surface Schema Example (Leave Application)**:
```json
{
  "type": "form",
  "title": "Apply for Leave",
  "api": "POST /api/v1/portal/employee/leaves",
  "body": [
    { "type": "select", "name": "leave_type", "label": "Leave Type",
      "source": "GET /api/v1/portal/employee/leave-types" },
    { "type": "input-date-range", "name": "date_range", "label": "Leave Dates", "required": true },
    { "type": "textarea", "name": "reason", "label": "Reason", "maxLength": 500 },
    { "type": "input-file", "name": "supporting_doc", "label": "Supporting Document (optional)",
      "receiver": "/api/v1/documents/upload", "accept": ".pdf,.jpg" }
  ],
  "onSuccess": {
    "actionType": "toast",
    "args": { "msg": "Leave application submitted successfully", "msgType": "success" }
  }
}
```

### 28B.4.5 Future Franchise Reporting Portal

A future portal variant for brand owners / franchisor companies (e.g. a fuel brand like Shell) to view aggregate performance across their franchisee tenants. This requires cross-tenant consent: the franchisee tenant must explicitly grant read access to their aggregate (not transactional) data to the franchisor's portal account. The `DisclosureConsent` model governs this and is tracked in a separate consent table.

---

## 28B.5 View Context Resolution in the Compilation Pipeline

### 28B.5.1 `ViewContext` Object

`ViewContext` is the complete context object passed through every stage of the UI compilation pipeline. It determines which navigation items are shown, which surface components are visible, and which data is returned.

```go
// internal/core/ui/context.go

// ViewContext carries the full actor context for UI compilation and navigation construction.
// It is derived from the authenticated session and the request headers.
type ViewContext struct {
    // Identity
    ActorType  string    `json:"actor_type"`  // platform_admin | tenant_staff | portal_supplier | portal_customer | portal_employee
    UserID     uuid.UUID `json:"user_id"`     // Internal user ID (from users or portal_users table)
    TenantID   uuid.UUID `json:"tenant_id"`   // Tenant this view is scoped to

    // Portal (set only for portal actors)
    PortalID   *uuid.UUID `json:"portal_id,omitempty"`
    PortalType string     `json:"portal_type,omitempty"` // supplier | customer | employee

    // Permissions (pre-loaded at login from Casbin evaluation)
    // Format: "<module>.<resource>.<action>" e.g. "finance.accounts.read"
    Permissions []string `json:"permissions"`

    // Feature flags (resolved for this tenant at request time)
    FeatureFlags map[string]bool `json:"feature_flags"`

    // Localisation
    Locale string `json:"locale"` // BCP 47 e.g. "en-KE", "sw-KE"

    // Request metadata
    RequestID string `json:"request_id"`
    TraceID   string `json:"trace_id"`

    // Tenant module enablement
    EnabledModules []string `json:"enabled_modules"`

    // Session metadata
    SessionID  uuid.UUID `json:"session_id"`
    IssuedAt   time.Time `json:"issued_at"`
    ExpiresAt  time.Time `json:"expires_at"`
}
```

### 28B.5.2 How `ViewContext` Flows Through Pipeline Stages

```
HTTP Request
  → TenantMiddleware: extracts TenantID from X-Tenant-ID header
  → AuthMiddleware:   validates JWT, extracts ActorType, UserID, SessionID
  → SessionMiddleware: loads Permissions, FeatureFlags, EnabledModules from session cache
  → Handler: constructs ViewContext from middleware-populated request context
  → NavigationService.GetTree(ctx, NavigationRequest{...viewCtx})
  → UICompilationService.Compile(ctx, surfaceID, viewCtx)
  → ResponseBuilder: serialises compiled surface with viewCtx-derived data
```

### 28B.5.3 Navigation Tree Selection by `actor_type`

```go
func (s *NavigationServiceImpl) selectBaseTree(actorType string) []NavigationGroup {
    switch actorType {
    case "platform_admin":
        return s.platformAdminGroups
    case "portal_supplier":
        return s.supplierPortalGroups
    case "portal_customer":
        return s.customerPortalGroups
    case "portal_employee":
        return s.employeePortalGroups
    default: // tenant_staff
        return s.allModuleGroups
    }
}
```

### 28B.5.4 Surface Access Guard by `actor_type`

The `UICompilationService` checks surface access by `actor_type` before compilation:

```go
func (s *UICompilationService) Compile(ctx context.Context, surfaceID string, viewCtx ViewContext) (*CompiledSurface, error) {
    surface, err := s.surfaceRepo.GetByID(ctx, surfaceID)
    if err != nil {
        return nil, err
    }

    if !surface.IsAccessibleTo(viewCtx.ActorType) {
        return nil, &shared.BusinessError{
            Code:    "permission_denied",
            Message: "Surface not accessible to your account type",
            Status:  403,
        }
    }

    if surface.RequiredPermission != "" && !slices.Contains(viewCtx.Permissions, surface.RequiredPermission) {
        return nil, &shared.BusinessError{
            Code:    "permission_denied",
            Message: "You don't have permission to view this surface",
            Status:  403,
        }
    }

    return s.compiler.Compile(surface, &CompileContext{ViewContext: viewCtx})
}
```

---

## 28B.6 Portal Authentication and Identity

### 28B.6.1 Separate Auth Endpoint for Portal Users

Portal users authenticate at a portal-specific endpoint:

```
POST /portal/supplier/auth/login
POST /portal/customer/auth/login
POST /portal/employee/auth/login
```

These endpoints use the same `AuthHandler` infrastructure but validate credentials against the `portal_users` table and issue portal-scoped JWTs.

### 28B.6.2 Portal JWT Claims

```json
{
  "sub": "portal-user-uuid",
  "tenant_id": "tenant-uuid",
  "portal_id": "portal-uuid",
  "portal_type": "supplier",
  "actor_type": "portal_supplier",
  "supplier_id": "supplier-uuid",
  "permissions": ["portal.supplier.po.read", "portal.supplier.invoices.write"],
  "iat": 1749196800,
  "exp": 1749225600
}
```

The `supplier_id` claim is used by CEL rules to enforce that a portal supplier can only access documents belonging to their supplier entity.

### 28B.6.3 Portal Session Lifetime Policy

| Portal Type | Session Duration | Refresh Token |
|-------------|-----------------|---------------|
| Supplier | 8 hours | Yes, up to 7 days |
| Customer | 8 hours | Yes, up to 30 days |
| Employee | 12 hours | Yes, up to 14 days |
| Platform Admin | 2 hours | No — re-auth required |

### 28B.6.4 Portal SSO Options

Customer and Employee portals support SAML 2.0 SSO for enterprise tenants that manage their own identity provider. The portal login endpoint accepts a SAML assertion in addition to username/password. Supplier portals support basic username/password only.

---

## 28B.7 amis Multi-View Shell Configuration

### 28B.7.1 Shell Selection by `actor_type` at Bootstrap Time

The `index.html` bootstrap script selects the shell based on `window.__AWO_BOOTSTRAP__.portalType`:

```javascript
var portalType = window.__AWO_BOOTSTRAP__.portalType;
var shellSchema = portalType === 'staff' ? STAFF_SHELL : PORTAL_SHELL;

amis.embed(document.getElementById('app'), shellSchema, {
    theme: 'cxd',
    locale: window.__AWO_BOOTSTRAP__.locale,
    fetcher: awoFetcher,
    requestAdaptor: requestAdaptor,
    responseAdaptor: responseAdaptor
});
```

### 28B.7.2 Portal Shell vs. Staff Shell amis `app` Config

**Staff Shell** (full sidebar + header):
```json
{
  "type": "app",
  "brandName": "${tenantName}",
  "logo": "${tenantLogoUrl}",
  "header": { "brand": "${tenantName}", "children": ["notifications", "user-menu"] },
  "aside": true,
  "pages": "${navPages}"
}
```

**Supplier Portal Shell** (minimal header, no sidebar):
```json
{
  "type": "page",
  "className": "portal-shell portal-shell--supplier",
  "header": {
    "type": "wrapper",
    "className": "portal-header",
    "body": [
      { "type": "image", "src": "${tenantLogoUrl}", "height": 40, "className": "portal-logo" },
      { "type": "tpl", "tpl": "Supplier Portal", "className": "portal-title" },
      { "type": "nav", "stacked": false, "links": [
        { "label": "Purchase Orders", "to": "/portal/supplier/po" },
        { "label": "Invoices", "to": "/portal/supplier/invoices" },
        { "label": "My Account", "to": "/portal/supplier/account" },
        { "label": "Logout", "to": "/portal/supplier/auth/logout" }
      ]}
    ]
  },
  "body": { "type": "container", "id": "portal-content-area" }
}
```

The supplier portal shell deliberately omits a sidebar to keep the interface clean for external users who are not ERP power users. Navigation is a simple top-level tab bar.

### 28B.7.3 Theme Token Swap on Shell Init

When the shell initialises, the portal type is written to the `html` element's `data-portal` attribute, triggering the portal-specific CSS custom property overrides:

```javascript
document.documentElement.setAttribute('data-portal', window.__AWO_BOOTSTRAP__.portalType);
```

---

## 28B.8 Flutter Multi-View Shell Configuration

### 28B.8.1 Portal-Aware Navigation Tree Consumer

Flutter reads the `actor_type` from the stored JWT and selects the appropriate shell widget:

```dart
Widget buildShell(BuildContext context, NavigationTree navTree) {
    var actorType = AuthProvider.of(context).actorType;
    switch (actorType) {
        case 'tenant_staff':    return StaffShell(navTree: navTree);
        case 'portal_supplier': return PortalShell(navTree: navTree, type: 'supplier');
        case 'portal_customer': return PortalShell(navTree: navTree, type: 'customer');
        case 'portal_employee': return PortalShell(navTree: navTree, type: 'employee');
        default:                return ErrorShell(message: 'Unknown actor type');
    }
}
```

### 28B.8.2 Deep Link Namespace Separation

Deep link routes are namespaced to prevent routing collisions:

```
/app/*           → Tenant staff routes
/portal/supplier/*  → Supplier portal routes
/portal/customer/*  → Customer portal routes
/portal/employee/*  → Employee portal routes
/platform/*      → Platform admin routes (restricted network only)
```

```dart
final router = GoRouter(routes: [
    GoRoute(path: '/app/:surface', builder: (ctx, state) =>
        SDUISurface(surfaceId: state.pathParameters['surface']!, actorType: 'tenant_staff')),
    GoRoute(path: '/portal/supplier/:surface', builder: (ctx, state) =>
        SDUISurface(surfaceId: state.pathParameters['surface']!, actorType: 'portal_supplier')),
    // ...
]);
```

---

## 28B.9 Cross-Portal Concerns

### 28B.9.1 Shared Component Library

All portals and the staff view use the same amis `cxd` component library. Portal-specific visual differences are achieved through theme token overrides (`data-portal` CSS scoping), not different component sets. This means a `crud` component on the supplier portal renders the same way as on the staff ERP — only colours, spacing, and typography differ.

### 28B.9.2 Localisation Per Portal

Portal users may prefer a different locale than staff users of the same tenant. The supplier portal defaults to `en-KE` but allows per-user locale preference stored in `portal_users.preferred_locale`. The Employee portal supports `sw-KE` (Swahili) as a first-class option since many Kenyan workers prefer Swahili-language interfaces.

### 28B.9.3 Accessibility Per Portal

All portals target WCAG 2.1 AA compliance. The customer portal additionally targets WCAG 2.1 AAA for colour contrast since it is a public-facing application. The employee portal is designed with touch-first interactions (larger tap targets, `--spacing-4: 18px`) for mobile access.

### 28B.9.4 Analytics Separation Between Portals

OpenTelemetry spans and Prometheus metrics are tagged with `portal_type` and `actor_type` to enable separate dashboards for staff ERP usage vs. portal usage. Portal conversion metrics (e.g. "what % of customer portal invoice views result in M-Pesa payment within 24h") are tracked separately from staff operational metrics.
