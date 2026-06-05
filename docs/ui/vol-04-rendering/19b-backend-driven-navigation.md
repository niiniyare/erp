---
title: "Backend-Driven Navigation"
volume: "IV — Rendering Architecture"
chapter: "19-B"
phase: 2
status: draft
audience: [Platform Engineers, Backend Engineers, Mobile Engineers, Web Engineers]
---

# Chapter 19-B — Backend-Driven Navigation

> **Volume:** IV — Rendering Architecture
> **Audience:** Platform Engineers, Backend (Go/Fiber) Engineers, Mobile/Web Engineers
> **Prerequisites:** Chapter 19 — Navigation Framework, Chapter 27 — Authorization Integration
> **Phase:** 2

---

## 19B.1 Philosophy — Navigation as a First-Class API Resource

### 19B.1.1 Why the Backend Owns Navigation

In conventional web applications, navigation is hardcoded in client code — a React router config, a Flutter `go_router` table, an Angular route manifest. This approach works when there is one application with one set of users. It fails catastrophically in a multi-tenant ERP where each of hundreds of tenants has different enabled modules, each user has different role-based permissions, feature flags gate early-access features, and the platform evolves weekly.

AwoERP makes a foundational architectural decision: **the backend is the single authoritative source of the complete navigation tree for every actor, in every context**. No navigation node is hardcoded in any client. Clients are rendering engines that receive, cache, and hydrate the navigation payload. This means that a tenant disabling the HR module instantly removes all HR navigation nodes from every session without a client deployment. A new feature gated behind a flag appears in navigation the moment the flag is enabled, without any client code change.

The navigation API endpoint `GET /api/v1/ui/navigation` returns a complete, pre-pruned tree — every item in the response is one the current user is permitted to see, given their role, their tenant's enabled modules, and the current flag state. Clients never perform client-side permission filtering on navigation data. If an item is in the response, it is visible. This eliminates an entire class of client-side permission bypass bugs.

### 19B.1.2 Navigation as a Permission Surface

Navigation is not merely a UX convenience — it is a **permission surface**. Showing a user a navigation item they cannot use creates confusion and erodes trust. Showing a navigation item that links to a resource the user could not access even if they navigated there manually is a security concern in environments subject to SOC 2 or Kenya DPA auditing, where "information disclosure" includes displaying the names of modules a user is not authorised to use.

The Casbin + CEL evaluation that prunes navigation items runs during tree construction in Go, not as a post-processing filter. This means the pruning logic has full access to the user's session context, the tenant's Casbin policy, and CEL-computed attributes (department, employment type, job grade), and produces a tree that is correct by construction rather than filtered after the fact.

### 19B.1.3 Navigation as a Tenant Configuration Surface

Tenant administrators can customise navigation beyond what the module system provides: reordering items, renaming labels (e.g. "Purchase Orders" → "Supplier Orders" for a tenant that prefers that terminology), hiding items they do not use, and adding external link items pointing to tools outside AwoERP (e.g. a link to their M-Pesa Business dashboard). These customisations are stored per-tenant in PostgreSQL and merged with the platform-defined tree during construction.

### 19B.1.4 Contrast with Client-Configured Routers

Client-configured routers (React Router, go_router) are appropriate when: the route set is static, all users see all routes, and the application has no tenancy or complex permission model. None of these conditions hold in AwoERP. Backend-driven navigation adds a network round-trip at login but eliminates the ongoing complexity of keeping client route configs synchronised with server-side permissions, tenant module configurations, and feature flags across hundreds of tenants.

---

## 19B.2 Navigation Payload Schema

### 19B.2.1 `NavigationTree` Root Object

```go
// internal/core/ui/navigation/types.go

type NavigationTree struct {
    Version    int64              `json:"version"`     // Monotonic; used for diff endpoint
    TenantID   uuid.UUID          `json:"tenant_id"`
    Portal     string             `json:"portal"`      // staff | supplier | customer | employee | platform_admin
    Locale     string             `json:"locale"`
    GeneratedAt time.Time         `json:"generated_at"`
    Groups     []NavigationGroup  `json:"groups"`
    // Metadata for observability
    PrunedCount int               `json:"pruned_count,omitempty"` // items removed by permission/flag checks
}
```

### 19B.2.2 `NavigationGroup` (Module-Level Grouping)

```go
type NavigationGroup struct {
    Key         string           `json:"key"`          // e.g. "finance", "hr", "fuel"
    Label       string           `json:"label"`        // Localised
    Icon        string           `json:"icon"`         // Icon identifier (amis icon name or custom SVG key)
    Order       int              `json:"order"`        // Display order within the tree
    Collapsible bool             `json:"collapsible"`
    DefaultOpen bool             `json:"default_open"`
    Items       []NavigationItem `json:"items"`
}
```

### 19B.2.3 `NavigationItem` Node

```go
type NavigationItem struct {
    Key                string            `json:"key"`                  // Unique within tree, e.g. "finance.accounts.list"
    Label              string            `json:"label"`                // Localised display text
    Icon               string            `json:"icon,omitempty"`       // amis icon name or SVG key
    Badge              *NavigationBadge  `json:"badge,omitempty"`      // Count, dot, or status badge
    Tooltip            string            `json:"tooltip,omitempty"`    // Hover tooltip text
    RouteKey           string            `json:"route_key"`            // Logical route identifier (NOT a URL path)
    SurfaceID          string            `json:"surface_id,omitempty"` // UI surface to render on activation
    PermissionRequired string            `json:"permission_required,omitempty"` // Already checked; informational only
    FeatureFlag        string            `json:"feature_flag,omitempty"`        // Already evaluated; informational only
    ExternalURL        string            `json:"external_url,omitempty"`        // Cross-system link
    OpenIn             string            `json:"open_in,omitempty"`             // same_tab | new_tab | modal | drawer
    Children           []NavigationItem  `json:"children,omitempty"`  // Sub-navigation
    Metadata           map[string]string `json:"metadata,omitempty"`  // Tenant-defined extension fields
    Order              int               `json:"order"`
    Disabled           bool              `json:"disabled,omitempty"`  // Visible but not clickable (e.g. coming soon)
}
```

### 19B.2.4 `NavigationBadge` Schema

```go
type NavigationBadge struct {
    Type    string `json:"type"`             // "count" | "dot" | "status"
    Count   int    `json:"count,omitempty"`  // For type=count
    Max     int    `json:"max,omitempty"`    // Display "99+" if count > max; default 99
    Color   string `json:"color,omitempty"`  // CSS color for dot/status badges
    Tooltip string `json:"tooltip,omitempty"`
    // SSE topic to subscribe to for live updates
    SSETopic string `json:"sse_topic,omitempty"` // e.g. "badge:approvals:{user_id}"
}
```

### 19B.2.5 Versioning Field on `NavigationTree`

The `version` field is a monotonically increasing integer stored per `(tenant_id, portal, user_role_hash)` combination in Redis. It is incremented whenever any of the following change:
- A permission grant/revoke for the user's roles
- A module enable/disable for the tenant
- A feature flag value change
- A tenant navigation customisation update

Clients can request a diff against their last known version using the diff endpoint (Section 19B.4.5), avoiding a full tree re-fetch on minor changes.

---

## 19B.3 Backend Construction of Navigation in Go

### 19B.3.1 `NavigationService` Interface and Implementation

```go
// internal/core/ui/navigation/service.go

type NavigationService interface {
    // GetTree returns the complete, permission-pruned navigation tree for the given context.
    GetTree(ctx context.Context, req NavigationRequest) (*NavigationTree, error)

    // GetDiff returns items changed since the given version.
    GetDiff(ctx context.Context, req NavigationRequest, sinceVersion int64) (*NavigationDiff, error)

    // InvalidateCache removes the cached tree for the given tenant/portal/role combination.
    InvalidateCache(ctx context.Context, tenantID uuid.UUID, portal string) error

    // RegisterModule registers a module's navigation items. Called at Wire wiring time.
    RegisterModule(module NavigationModule) error
}

type NavigationRequest struct {
    TenantID    uuid.UUID
    UserID      uuid.UUID
    Permissions []string   // Pre-loaded from session
    Portal      string
    Locale      string
    FlagHash    string     // Hash of current flag values for cache key
    RoleHash    string     // Hash of user's role set for cache key
}

// NavigationServiceImpl is the concrete implementation.
type NavigationServiceImpl struct {
    modules  []NavigationModule
    cache    cache.Service
    store    db.Store
    flags    FeatureFlagService
    authz    AuthzService
    audit    AuditService
}
```

### 19B.3.2 Module Registration Pattern (`RegisterNavItems()`)

Each ERP module implements the `NavigationModule` interface and registers during Wire wiring:

```go
// internal/core/ui/navigation/module.go

type NavigationModule interface {
    // Key returns the module identifier (e.g. "finance", "hr", "fuel")
    Key() string
    // NavItems returns the module's navigation contribution.
    // Items are pre-filtered by this module's internal logic but NOT by permissions —
    // permission pruning is done centrally by NavigationServiceImpl.
    NavItems(ctx context.Context, tenantCtx TenantContext) ([]NavigationGroup, error)
}

// Example: Finance module registration
// internal/core/finance/ui/nav.go

type FinanceNavModule struct {
    moduleConfig FinanceModuleConfig
}

func (m *FinanceNavModule) Key() string { return "finance" }

func (m *FinanceNavModule) NavItems(ctx context.Context, tc TenantContext) ([]NavigationGroup, error) {
    items := []NavigationItem{
        {
            Key: "finance.dashboard", Label: "Dashboard", Icon: "fa fa-chart-line",
            RouteKey: "finance.dashboard", SurfaceID: "finance.dashboard",
            PermissionRequired: "finance.dashboard.read", Order: 1,
        },
        {
            Key: "finance.accounts", Label: "Chart of Accounts", Icon: "fa fa-book",
            RouteKey: "finance.accounts.list", SurfaceID: "finance.accounts.list",
            PermissionRequired: "finance.accounts.read", Order: 2,
        },
        {
            Key: "finance.invoices", Label: "Invoices", Icon: "fa fa-file-invoice",
            RouteKey: "finance.invoices.list", SurfaceID: "finance.invoices.list",
            PermissionRequired: "finance.invoices.read", Order: 3,
            Badge: &NavigationBadge{
                Type: "count", SSETopic: "badge:invoices:overdue:{tenant_id}",
            },
        },
    }
    return []NavigationGroup{{Key: "finance", Label: "Finance", Icon: "fa fa-calculator", Order: 2, Items: items}}, nil
}
```

Wire wiring registers all modules:

```go
// internal/platform/wire/navigation.go

func ProvideNavigationService(
    financeNav *finance.FinanceNavModule,
    hrNav *hr.HRNavModule,
    procNav *procurement.ProcurementNavModule,
    fuelNav *fuel.FuelNavModule,
    cache cache.Service,
    store db.Store,
    flags FeatureFlagService,
    authz AuthzService,
    audit AuditService,
) (navigation.NavigationService, error) {
    svc := navigation.NewNavigationServiceImpl(cache, store, flags, authz, audit)
    modules := []navigation.NavigationModule{financeNav, hrNav, procNav, fuelNav}
    for _, m := range modules {
        if err := svc.RegisterModule(m); err != nil {
            return nil, fmt.Errorf("register nav module %q: %w", m.Key(), err)
        }
    }
    return svc, nil
}
```

### 19B.3.3 Permission Pruning During Construction

Permission pruning iterates every `NavigationItem` and removes items where the user's session permissions do not include the required permission:

```go
func (s *NavigationServiceImpl) pruneByPermissions(
    groups []NavigationGroup,
    permissions []string,
) []NavigationGroup {
    permSet := make(map[string]bool, len(permissions))
    for _, p := range permissions {
        permSet[p] = true
    }

    var result []NavigationGroup
    for _, group := range groups {
        pruned := pruneItems(group.Items, permSet)
        if len(pruned) > 0 {
            group.Items = pruned
            result = append(result, group)
        }
    }
    return result
}

func pruneItems(items []NavigationItem, permSet map[string]bool) []NavigationItem {
    var result []NavigationItem
    for _, item := range items {
        if item.PermissionRequired != "" && !permSet[item.PermissionRequired] {
            continue // prune
        }
        if len(item.Children) > 0 {
            item.Children = pruneItems(item.Children, permSet)
            if len(item.Children) == 0 && item.SurfaceID == "" {
                continue // parent with no children and no own surface — prune
            }
        }
        result = append(result, item)
    }
    return result
}
```

#### RLS-Aware Node Removal (Tenant Scope)

Beyond permission checks, some navigation items are only valid for tenants with specific data. For example, the "Fuel Grades" item in the Fuel module is removed if the tenant has no EPRA-registered fuel grades in their database:

```go
func (s *NavigationServiceImpl) pruneByTenantScope(
    ctx context.Context,
    groups []NavigationGroup,
    tenantID uuid.UUID,
) ([]NavigationGroup, error) {
    return store.WithTenant(ctx, tenantID, func(q *db.Queries) ([]NavigationGroup, error) {
        hasFuelGrades, _ := q.HasFuelGrades(ctx)
        if !hasFuelGrades {
            groups = removeItemByKey(groups, "fuel.grades")
        }
        return groups, nil
    })
}
```

### 19B.3.4 Feature Flag Evaluation During Construction

Items gated by feature flags are removed if the flag is disabled for the tenant's context:

```go
func (s *NavigationServiceImpl) pruneByFlags(
    ctx context.Context,
    groups []NavigationGroup,
    tenantID uuid.UUID,
) ([]NavigationGroup, error) {
    var result []NavigationGroup
    for _, group := range groups {
        var items []NavigationItem
        for _, item := range group.Items {
            if item.FeatureFlag != "" {
                enabled, err := s.flags.IsEnabled(ctx, item.FeatureFlag, tenantID)
                if err != nil || !enabled {
                    continue
                }
            }
            items = append(items, item)
        }
        if len(items) > 0 {
            group.Items = items
            result = append(result, group)
        }
    }
    return result, nil
}
```

### 19B.3.5 Tenant Navigation Overrides (Custom Items, Re-Ordering)

Tenant customisations are loaded from `tenant_nav_overrides` and merged:

```sql
CREATE TABLE tenant_nav_overrides (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    portal      TEXT NOT NULL,
    overrides   JSONB NOT NULL,
    -- overrides schema:
    -- { "reorder": {"finance": 1, "hr": 3}, "labels": {"finance.invoices": "Bills"},
    --   "hidden": ["hr.payroll.reports"], "extra_items": [{...NavigationItem}] }
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, portal)
);
```

### 19B.3.6 Portal Context Switching

The `portal` parameter in `NavigationRequest` selects which module set is active:

```go
func (s *NavigationServiceImpl) filterByPortal(groups []NavigationGroup, portal string) []NavigationGroup {
    switch portal {
    case "platform_admin":
        return filterGroupsByKeys(groups, []string{"platform"})
    case "supplier":
        return filterGroupsByKeys(groups, []string{"procurement.supplier", "account"})
    case "customer":
        return filterGroupsByKeys(groups, []string{"sales.customer", "payments", "account"})
    case "employee":
        return filterGroupsByKeys(groups, []string{"hr.self_service", "account"})
    default: // "staff"
        return groups // all enabled module groups
    }
}
```

### 19B.3.7 Wire Provider Registration for `NavigationService`

The navigation handler is registered in `routes.go` via the `Dependencies` struct:

```go
// internal/api/handlers/routes.go (relevant excerpt)
type Dependencies struct {
    // ...
    NavigationService navigation.NavigationService
    // ...
}

func RegisterRoutes(app *fiber.App, deps Dependencies) {
    ui := app.Group("/api/v1/ui", middleware.AuthenticateMiddleware(deps.SessionService))
    ui.Get("/navigation", deps.NavigationHandler.GetNavigation)
    ui.Get("/navigation/diff", deps.NavigationHandler.GetNavigationDiff)
}
```

---

## 19B.4 Navigation API Endpoints

### 19B.4.1 `GET /api/v1/ui/navigation` — Full Tree for Authenticated User

**Request:**
```http
GET /api/v1/ui/navigation HTTP/1.1
Host: app.awoerp.com
Authorization: Bearer eyJhbGciOiJSUzI1NiJ9...
X-Tenant-ID: 3f2e1d4c-5b6a-7c8d-9e0f-a1b2c3d4e5f6
Accept-Language: en-KE
X-Portal: staff
```

**Response (200 OK):**
```http
HTTP/1.1 200 OK
Content-Type: application/json
ETag: "nav-v42-hash-a3f9"
Cache-Control: private, max-age=300
Vary: Authorization, X-Tenant-ID, Accept-Language, X-Portal
```

```json
{
  "data": {
    "version": 42,
    "tenant_id": "3f2e1d4c-5b6a-7c8d-9e0f-a1b2c3d4e5f6",
    "portal": "staff",
    "locale": "en-KE",
    "generated_at": "2026-06-06T08:00:00Z",
    "groups": [
      {
        "key": "dashboard",
        "label": "Dashboard",
        "icon": "fa fa-home",
        "order": 1,
        "collapsible": false,
        "default_open": true,
        "items": [
          {
            "key": "dashboard.main",
            "label": "Overview",
            "icon": "fa fa-tachometer-alt",
            "route_key": "dashboard",
            "surface_id": "dashboard.main",
            "order": 1
          }
        ]
      }
    ],
    "pruned_count": 7
  },
  "meta": { "request_id": "req-uuid", "latency_ms": 12 }
}
```

### 19B.4.2 Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `portal` | string | `staff` | Portal context: `staff\|supplier\|customer\|employee\|platform_admin` |
| `locale` | string | Session locale | BCP 47 locale code (e.g. `en-KE`, `sw-KE`) |
| `surface_hint` | string | — | Current surface ID; used to mark the active item |

### 19B.4.4 ETag / Cache-Control Strategy

The response `ETag` encodes the navigation tree version. Clients should send `If-None-Match` on subsequent requests:

```http
GET /api/v1/ui/navigation
If-None-Match: "nav-v42-hash-a3f9"
```

If the tree has not changed, the server responds `304 Not Modified` with no body, saving bandwidth on constrained Kenyan mobile connections.

### 19B.4.5 Navigation Diff Endpoint

```http
GET /api/v1/ui/navigation/diff?since=42
```

Returns only items added, removed, or modified since version 42. The response uses the same `NavigationTree` schema but with `diff_type` annotations on each group/item: `"added" | "removed" | "modified" | "unchanged"`.

---

## 19B.5 Navigation Caching Strategy

### 19B.5.1 Cache Key Design

```go
func navCacheKey(tenantID uuid.UUID, roleHash, portal, locale, flagHash string) string {
    return fmt.Sprintf("nav:%s:%s:%s:%s:%s", tenantID, roleHash, portal, locale, flagHash)
}
```

- `roleHash`: SHA-256 of the sorted permission set. Two users with identical permissions share one cache entry.
- `flagHash`: SHA-256 of the tenant's resolved feature flag map. Changes when any flag is toggled.

### 19B.5.2 Redis TTL Policy

| Scenario | TTL |
|----------|-----|
| Standard staff session | 5 minutes |
| Portal (supplier/customer/employee) | 15 minutes (less volatile) |
| Platform admin | 1 minute (high sensitivity) |
| Warm-up entry (provisioning) | 10 minutes |

### 19B.5.3 Invalidation Triggers

```go
// Triggers that invalidate navigation cache for a tenant
func (s *NavigationServiceImpl) InvalidateCache(ctx context.Context, tenantID uuid.UUID, portal string) error {
    pattern := fmt.Sprintf("nav:%s:*:%s:*:*", tenantID, portal)
    return s.cache.DeletePattern(ctx, pattern)
}

// Called from:
// - iam.RoleAssignmentService.Assign() / Revoke()
// - tenant.ModuleService.Enable() / Disable()
// - featureflags.Service.Set() — for tenant-scoped flags
// - navigation.TenantOverrideService.Save()
```

### 19B.5.4 Warm-Up on Tenant Provisioning

When a new tenant is provisioned, the navigation cache is pre-warmed for the default staff portal and the `tenant_admin` role to ensure the first login is fast:

```go
func (s *TenantProvisioningService) warmNavigationCache(ctx context.Context, tenantID uuid.UUID) error {
    req := navigation.NavigationRequest{
        TenantID: tenantID,
        Portal:   "staff",
        Locale:   "en-KE",
        Permissions: defaultAdminPermissions,
    }
    _, err := s.navService.GetTree(ctx, req)
    return err
}
```

---

## 19B.6 Client Consumption of Navigation Payload

### 19B.6.1 Fetch-on-Login Pattern

The navigation tree is fetched immediately after successful authentication, before the first surface render:

```javascript
// web/src/bootstrap.js
async function bootstrapApp() {
    const token = await auth.getToken();
    if (!token) { redirectToLogin(); return; }

    const [navTree, firstSurface] = await Promise.all([
        api.get('/api/v1/ui/navigation', { portal: bootstrap.portalType }),
        api.get(`/api/v1/ui/surfaces/${bootstrap.defaultSurface}`)
    ]);

    store.setNavigation(navTree.data);
    amis.embed(document.getElementById('app'), firstSurface.data, env);
}
```

### 19B.6.2 Client-Side Navigation Store

The navigation tree is stored in a client-side store (plain JavaScript object in the amis `app` schema's `data` context) and rebuilt when the server signals a change via SSE `nav_diff` events.

### 19B.6.3 Route Key → Client Route Mapping Table

`route_key` values are logical identifiers independent of URL paths. The mapping from `route_key` to an actual URL path is maintained client-side:

```javascript
// web/src/navigation/route-map.js
const routeMap = {
    "dashboard":              "/app/dashboard",
    "finance.accounts.list":  "/app/finance/accounts",
    "finance.invoices.list":  "/app/finance/invoices",
    "hr.leaves.list":         "/app/hr/leaves",
    "fuel.sales.today":       "/app/fuel/sales",
};
```

### 19B.6.4 Handling Unknown `route_key` Values

Forward compatibility: when a client encounters a `route_key` it does not have a mapping for (e.g., a new module deployed to the server before the client is updated), it renders the item as visible but navigates to a "Loading new feature..." surface on click, which instructs the user to refresh.

### 19B.6.5 Badge Count Refresh (Polling vs. SSE Push)

Badge counts are refreshed via SSE. The `NavigationBadge.sse_topic` field names the SSE channel the client should subscribe to for live badge updates. When SSE is unavailable (offline, poor connectivity), the client falls back to polling at 60-second intervals.

### 19B.6.6 Navigation Re-Fetch on Permission Change

The server broadcasts an SSE event `nav_invalidated` to the user's session stream when their permissions change (role assigned/revoked). The client re-fetches the full navigation tree on receiving this event.

---

## 19B.7 Navigation for amis Web Renderer

### 19B.7.1 amis `nav` Component Binding to Navigation Payload

The amis `nav` component receives the navigation data through the `app` component's `pages` array. The `NavigationTree` is transformed into the amis `pages` format by a JavaScript transformer in `web/src/navigation/amis-adapter.js`:

```javascript
function navTreeToAmisPages(tree) {
    return tree.groups.flatMap(group => ({
        label: group.label,
        icon:  group.icon,
        children: group.items.map(item => ({
            label:   item.label,
            url:     routeMap[item.route_key] || '#',
            icon:    item.icon,
            badge:   item.badge?.count,
            schema:  { type: 'service', api: `/api/v1/ui/surfaces/${item.surface_id}` }
        }))
    }));
}
```

### 19B.7.2 `app` Component Shell Navigation Wiring

```json
{
  "type": "app",
  "brandName": "${tenantName}",
  "logo": "${tenantLogoUrl}",
  "header": { "type": "header-toolbar" },
  "pages": "${navPages}",
  "data": {
    "navPages": []
  },
  "onMount": {
    "actionType": "custom",
    "script": "fetchNavigation().then(tree => { amis.updateState({ navPages: navTreeToAmisPages(tree) }); })"
  }
}
```

### 19B.7.3 Active State Derivation from Current Surface

The `app` component marks the nav item active based on the current URL path, which maps back to a `route_key`. The `routeMap` reverse lookup provides this mapping.

### 19B.7.4 Collapsed / Expanded State Persistence

The user's collapsed/expanded state per nav group is persisted in `sessionStorage` keyed by `nav-state:{portal}`. This state is not synced to the server — it is a pure UX preference and is reset on new sessions.

---

## 19B.8 Navigation for Flutter Mobile Renderer

### 19B.8.1 Bottom Tab Bar Construction from `NavigationTree`

Flutter constructs the bottom tab bar from the first N groups (typically 4–5) in the navigation tree:

```dart
// lib/navigation/nav_builder.dart

BottomNavigationBar buildBottomNav(NavigationTree tree) {
  final topGroups = tree.groups.take(5).toList();
  return BottomNavigationBar(
    items: topGroups.map((group) => BottomNavigationBarItem(
      icon: AwoIcon.fromKey(group.icon),
      label: group.label,
    )).toList(),
    onTap: (index) => navigateToGroup(topGroups[index]),
  );
}
```

### 19B.8.2 Drawer Navigation Construction

Groups not fitting in the bottom tab bar are placed in the navigation drawer:

```dart
Drawer buildDrawer(NavigationTree tree) {
  return Drawer(
    child: ListView(
      children: tree.groups.skip(5).map((group) =>
        ExpansionTile(
          leading: AwoIcon.fromKey(group.icon),
          title: Text(group.label),
          children: group.items.map((item) =>
            ListTile(
              leading: AwoIcon.fromKey(item.icon),
              title: Text(item.label),
              trailing: item.badge != null ? Badge(count: item.badge!.count) : null,
              onTap: () => context.push(routeMap[item.routeKey]!),
            )
          ).toList(),
        )
      ).toList(),
    ),
  );
}
```

### 19B.8.3 Deep Link Routing from `route_key`

Flutter `go_router` is configured with routes derived from the navigation tree's `route_key` values, with the `route_key` serving as the `name` parameter for named navigation:

```dart
GoRouter buildRouter(NavigationTree tree) {
  final routes = tree.groups
    .expand((g) => g.items)
    .map((item) => GoRoute(
      name: item.routeKey,
      path: routePathFor(item.routeKey),
      builder: (ctx, state) => SDUISurface(surfaceId: item.surfaceId),
    )).toList();

  return GoRouter(routes: routes);
}
```

---

## 19B.9 Portal-Specific Navigation Trees

### 19B.9.1 Staff Portal Navigation (Full ERP Modules)

The staff portal shows all modules enabled for the tenant. For Shell Maanzoni fuel station (used as the canonical Kenya tenant example throughout this documentation), the staff navigation tree includes five module groups.

**Complete `NavigationTree` JSON for Shell Maanzoni (Staff Portal):**

```json
{
  "version": 15,
  "tenant_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "portal": "staff",
  "locale": "en-KE",
  "generated_at": "2026-06-06T08:00:00Z",
  "groups": [
    {
      "key": "dashboard",
      "label": "Dashboard",
      "icon": "fa fa-home",
      "order": 1,
      "collapsible": false,
      "default_open": true,
      "items": [
        {
          "key": "dashboard.main",
          "label": "Station Overview",
          "icon": "fa fa-chart-bar",
          "route_key": "dashboard",
          "surface_id": "dashboard.fuel-station",
          "order": 1
        }
      ]
    },
    {
      "key": "fuel",
      "label": "Fuel Operations",
      "icon": "fa fa-gas-pump",
      "order": 2,
      "collapsible": true,
      "default_open": true,
      "items": [
        {
          "key": "fuel.sales",
          "label": "Today's Sales",
          "icon": "fa fa-tachometer-alt",
          "route_key": "fuel.sales.today",
          "surface_id": "fuel.sales.today",
          "permission_required": "fuel.sales.read",
          "order": 1,
          "badge": {
            "type": "count",
            "sse_topic": "badge:fuel-sales:open:{tenant_id}",
            "tooltip": "Open pump sessions"
          }
        },
        {
          "key": "fuel.grades",
          "label": "Fuel Grades (EPRA)",
          "icon": "fa fa-flask",
          "route_key": "fuel.grades.list",
          "surface_id": "fuel.grades.list",
          "permission_required": "fuel.grades.read",
          "order": 2
        },
        {
          "key": "fuel.inventory",
          "label": "Tank Inventory",
          "icon": "fa fa-oil-can",
          "route_key": "fuel.inventory.view",
          "surface_id": "fuel.inventory.view",
          "permission_required": "fuel.inventory.read",
          "order": 3
        },
        {
          "key": "fuel.deliveries",
          "label": "Fuel Deliveries",
          "icon": "fa fa-truck",
          "route_key": "fuel.deliveries.list",
          "surface_id": "fuel.deliveries.list",
          "permission_required": "fuel.deliveries.read",
          "order": 4
        },
        {
          "key": "fuel.pump-readings",
          "label": "Pump Meter Readings",
          "icon": "fa fa-gauge",
          "route_key": "fuel.pump-readings.list",
          "surface_id": "fuel.pump-readings.list",
          "permission_required": "fuel.pump_readings.read",
          "order": 5
        }
      ]
    },
    {
      "key": "shop",
      "label": "Shop & Convenience",
      "icon": "fa fa-store",
      "order": 3,
      "collapsible": true,
      "default_open": false,
      "items": [
        {
          "key": "shop.pos",
          "label": "Point of Sale",
          "icon": "fa fa-cash-register",
          "route_key": "shop.pos",
          "surface_id": "shop.pos.main",
          "permission_required": "shop.pos.operate",
          "order": 1
        },
        {
          "key": "shop.inventory",
          "label": "Shop Inventory",
          "icon": "fa fa-boxes",
          "route_key": "shop.inventory.list",
          "surface_id": "shop.inventory.list",
          "permission_required": "shop.inventory.read",
          "order": 2
        },
        {
          "key": "shop.orders",
          "label": "Supplier Orders",
          "icon": "fa fa-clipboard-list",
          "route_key": "shop.orders.list",
          "surface_id": "procurement.po.list",
          "permission_required": "procurement.po.read",
          "order": 3
        }
      ]
    },
    {
      "key": "restaurant",
      "label": "Restaurant",
      "icon": "fa fa-utensils",
      "order": 4,
      "collapsible": true,
      "default_open": false,
      "items": [
        {
          "key": "restaurant.orders",
          "label": "Table Orders",
          "icon": "fa fa-clipboard",
          "route_key": "restaurant.orders.live",
          "surface_id": "restaurant.orders.live",
          "permission_required": "restaurant.orders.read",
          "order": 1
        },
        {
          "key": "restaurant.menu",
          "label": "Menu Management",
          "icon": "fa fa-book-open",
          "route_key": "restaurant.menu.list",
          "surface_id": "restaurant.menu.list",
          "permission_required": "restaurant.menu.write",
          "order": 2
        }
      ]
    },
    {
      "key": "lpg",
      "label": "LPG",
      "icon": "fa fa-fire",
      "order": 5,
      "collapsible": true,
      "default_open": false,
      "items": [
        {
          "key": "lpg.sales",
          "label": "Cylinder Sales",
          "icon": "fa fa-shopping-basket",
          "route_key": "lpg.sales.list",
          "surface_id": "lpg.sales.list",
          "permission_required": "lpg.sales.read",
          "order": 1
        },
        {
          "key": "lpg.inventory",
          "label": "Cylinder Inventory",
          "icon": "fa fa-database",
          "route_key": "lpg.inventory.view",
          "surface_id": "lpg.inventory.view",
          "permission_required": "lpg.inventory.read",
          "order": 2
        }
      ]
    },
    {
      "key": "spares",
      "label": "Spare Parts",
      "icon": "fa fa-tools",
      "order": 6,
      "collapsible": true,
      "default_open": false,
      "items": [
        {
          "key": "spares.catalogue",
          "label": "Parts Catalogue",
          "icon": "fa fa-list",
          "route_key": "spares.catalogue.list",
          "surface_id": "spares.catalogue.list",
          "permission_required": "spares.catalogue.read",
          "order": 1
        }
      ]
    },
    {
      "key": "finance",
      "label": "Finance",
      "icon": "fa fa-calculator",
      "order": 7,
      "collapsible": true,
      "default_open": false,
      "items": [
        {
          "key": "finance.invoices",
          "label": "Invoices",
          "icon": "fa fa-file-invoice",
          "route_key": "finance.invoices.list",
          "surface_id": "finance.invoices.list",
          "permission_required": "finance.invoices.read",
          "order": 1,
          "badge": { "type": "count", "sse_topic": "badge:invoices:overdue:{tenant_id}" }
        },
        {
          "key": "finance.accounts",
          "label": "Accounts",
          "icon": "fa fa-book",
          "route_key": "finance.accounts.list",
          "surface_id": "finance.accounts.list",
          "permission_required": "finance.accounts.read",
          "order": 2
        },
        {
          "key": "finance.reports",
          "label": "Reports",
          "icon": "fa fa-chart-pie",
          "route_key": "finance.reports",
          "surface_id": "finance.reports.index",
          "permission_required": "finance.reports.read",
          "order": 3
        }
      ]
    }
  ],
  "pruned_count": 3
}
```

### 19B.9.2 Supplier Portal Navigation

Scoped to procurement interactions only:

```json
{
  "portal": "supplier",
  "groups": [
    {
      "key": "procurement.supplier",
      "label": "My Orders",
      "items": [
        { "key": "supplier.pos", "label": "Purchase Orders", "route_key": "supplier.po.list", "surface_id": "portal.supplier.po.list" },
        { "key": "supplier.invoices", "label": "My Invoices", "route_key": "supplier.invoices.list", "surface_id": "portal.supplier.invoices.list" },
        { "key": "supplier.grn", "label": "Delivery Notes", "route_key": "supplier.grn.list", "surface_id": "portal.supplier.grn.list" }
      ]
    },
    {
      "key": "account",
      "label": "My Account",
      "items": [
        { "key": "account.statement", "label": "Statement", "route_key": "supplier.statement", "surface_id": "portal.supplier.statement" },
        { "key": "account.profile", "label": "Profile", "route_key": "supplier.profile", "surface_id": "portal.supplier.profile" }
      ]
    }
  ]
}
```

### 19B.9.3 Customer Portal Navigation

```json
{
  "portal": "customer",
  "groups": [
    {
      "key": "sales.customer",
      "label": "My Orders",
      "items": [
        { "key": "customer.orders", "label": "Orders", "route_key": "customer.orders.list", "surface_id": "portal.customer.orders.list" },
        { "key": "customer.invoices", "label": "Invoices", "route_key": "customer.invoices.list", "surface_id": "portal.customer.invoices.list" },
        { "key": "customer.payments", "label": "Payments", "route_key": "customer.payments.list", "surface_id": "portal.customer.payments.list" }
      ]
    },
    { "key": "account", "label": "My Account", "items": [
        { "key": "account.statement", "label": "Statement", "route_key": "customer.statement", "surface_id": "portal.customer.statement" }
      ]
    }
  ]
}
```

### 19B.9.4 Employee Self-Service Navigation

```json
{
  "portal": "employee",
  "groups": [
    {
      "key": "hr.self_service",
      "label": "My HR",
      "items": [
        { "key": "ess.leaves", "label": "Leave Requests", "route_key": "ess.leaves.list", "surface_id": "portal.ess.leaves.list",
          "badge": { "type": "count", "sse_topic": "badge:leaves:pending:{user_id}", "tooltip": "Pending requests" } },
        { "key": "ess.payslips", "label": "Payslips", "route_key": "ess.payslips.list", "surface_id": "portal.ess.payslips.list" },
        { "key": "ess.expenses", "label": "Expense Claims", "route_key": "ess.expenses.list", "surface_id": "portal.ess.expenses.list" }
      ]
    },
    { "key": "account", "label": "My Profile", "items": [
        { "key": "ess.profile", "label": "Personal Details", "route_key": "ess.profile", "surface_id": "portal.ess.profile" }
      ]
    }
  ]
}
```

### 19B.9.5 Super-Admin Navigation (Cross-Tenant)

```json
{
  "portal": "platform_admin",
  "groups": [
    {
      "key": "platform",
      "label": "Platform Management",
      "items": [
        { "key": "platform.tenants", "label": "Tenants", "route_key": "platform.tenants.list", "surface_id": "platform.tenants.list" },
        { "key": "platform.flags", "label": "Feature Flags", "route_key": "platform.flags.list", "surface_id": "platform.flags.list" },
        { "key": "platform.audit", "label": "Audit Log", "route_key": "platform.audit.list", "surface_id": "platform.audit.list" },
        { "key": "platform.billing", "label": "Billing", "route_key": "platform.billing.list", "surface_id": "platform.billing.list" }
      ]
    }
  ]
}
```

---

## 19B.10 Navigation and Multi-Tenancy

### 19B.10.1 Per-Tenant Navigation Customisation Storage

Tenant admins can configure navigation via `PUT /api/v1/tenants/{id}/navigation` with a customisation payload stored in `tenant_nav_overrides`. Supported customisations: label rename, item hiding, item reorder, external link addition.

### 19B.10.2 Tenant Admin Navigation Builder UI

A dedicated surface `platform.tenant-nav-builder` provides a drag-and-drop navigation editor for tenant admins. The builder operates on top of the live navigation API — it shows the current resolved tree and allows customisations within the permitted scope (tenants cannot add items requiring permissions they don't own).

### 19B.10.3 Module Enablement and Navigation Tree Reduction

When a tenant module is disabled via `tenant.ModuleService.Disable()`, the corresponding navigation group is removed from all cached trees via `InvalidateCache()`. The removal is immediate for new requests; existing sessions receive an SSE `nav_invalidated` event and re-fetch.

---

## 19B.11 Navigation Audit and Observability

### 19B.11.1 Navigation Fetch Latency Metrics

```go
var (
    navFetchDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "awo_navigation_fetch_duration_seconds",
        Help:    "Duration of navigation tree fetch (cache hit vs miss)",
        Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5},
    }, []string{"portal", "cache_result"}) // cache_result: hit|miss|error
)
```

P95 target: < 10ms (cache hit), < 100ms (cache miss with DB query).

### 19B.11.2 Pruned Item Audit Log

Every navigation tree construction that prunes items logs a structured audit entry for SOC 2 evidence:

```go
deps.AuditService.Record(ctx, audit.Event{
    Type:     "navigation.tree_constructed",
    ActorID:  req.UserID,
    TenantID: req.TenantID,
    Metadata: map[string]any{
        "portal":       req.Portal,
        "total_items":  totalItems,
        "pruned_count": prunedCount,
        "pruned_keys":  prunedKeys,
        "cache_result": cacheResult,
    },
})
```

### 19B.11.3 Navigation Click Telemetry

Navigation item clicks are recorded via OpenTelemetry span attributes to understand feature usage across tenants. The amis `app` component fires a custom event on nav item activation that the bootstrap init script forwards to the telemetry endpoint:

```javascript
window.addEventListener('awo:nav-click', function(e) {
    telemetry.track('navigation.item_clicked', {
        route_key:  e.detail.routeKey,
        surface_id: e.detail.surfaceId,
        portal:     bootstrap.portalType,
        tenant_id:  bootstrap.tenantId
    });
});
```

> See Chapter 38 — Observability and Chapter 39 — Telemetry for the full metrics and tracing integration.
