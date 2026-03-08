# Awo ERP — Module Architecture & Design Reference

> **Document Purpose:** Canonical reference for system architects and developers. Covers every module's package name, services, models, methods, UI navigation label, and the inter-module dependency strategy.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Naming Conventions](#naming-conventions)
3. [Layer Map](#layer-map)
4. [UI Navigation Structure](#ui-navigation-structure)
5. [Platform Layer](#platform-layer)
6. [Business Layer — Finance](#business-layer--finance)
7. [Business Layer — People](#business-layer--people)
8. [Business Layer — Supply Chain](#business-layer--supply-chain)
9. [Business Layer — Revenue](#business-layer--revenue)
10. [Business Layer — Production](#business-layer--production)
11. [Business Layer — Projects](#business-layer--projects)
12. [Analytics Layer](#analytics-layer)
13. [Dependency Cycle Prevention](#dependency-cycle-prevention)
14. [Configuration & Feature Flags](#configuration--feature-flags)
15. [Fiber API Layer](#fiber-api-layer)
16. [Condition Evaluation Engine — `pkg/condition`](#condition-evaluation-engine)
17. [Tenant Business Rules & Workflow Decisions](#tenant-business-rules--workflow-decisions)

---

## Architecture Overview

Awo ERP is a multi-tenant, server-rendered Go application. All modules live in a single deployable binary (modular monolith). Modules are separated by Go packages, not by processes. Communication across module boundaries follows strict rules to prevent dependency cycles.

```
┌────────────────────────────────────────────────────────────┐
│                     Analytics Layer                         │
│            (read-only; consumes domain events)              │
├───────────────────────────┬────────────────────────────────┤
│    Business Layer         │   Business Layer               │
│    Finance / People       │   Supply / Revenue / Prod      │
│    Projects               │                                │
├───────────────────────────┴────────────────────────────────┤
│                     Platform Layer                          │
│     IAM · Tenant · Audit · Notify · Bridge                  │
├────────────────────────────────────────────────────────────┤
│              Infrastructure / Database                       │
│     PostgreSQL (RLS) · Go stdlib · HTMX · templ             │
└────────────────────────────────────────────────────────────┘
```

**Rule:** Business modules depend **downward only** (toward Platform). They never import each other directly. Cross-module coordination happens exclusively through the **Domain Event Bus** (in-process pub/sub) or **Workflow Orchestration** for long-running sagas.

---

## Naming Conventions

Every module has three names:

| Concept | Description | Example |
|---|---|---|
| **Package Name** | Go import path used by developers | `awo/core/ledger` |
| **Service Name** | The Go struct developers instantiate | `ledger.JournalService` |
| **UI Label** | The label visible to end-users in navigation | `Accounting → Journal Entries` |

Method names follow Go conventions: `VerbNoun` in PascalCase. Event names are past-tense domain nouns: `PaymentPosted`, `StockAdjusted`.

---

## Layer Map

```
awo/
├── pkg/
│   └── condition/    → Runtime condition evaluation engine (pure library)
│                       Used by: platform/rules · platform/flags · platform/iam
│
├── platform/
│   ├── iam/          → Identity & Access Management
│   ├── tenant/       → Organisation & Tenant Config
│   ├── audit/        → Audit & Compliance Log
│   ├── notify/       → Notification Hub
│   ├── bridge/       → External Integration Gateway
│   ├── rules/        → Business Rule Engine (wraps pkg/condition)
│   ├── flags/        → Feature Flag Service (wraps pkg/condition)
│   └── config/       → Configuration & Cascade Resolution
│
├── core/
│   ├── ledger/       → General Ledger
│   ├── payables/     → Accounts Payable
│   ├── receivables/  → Accounts Receivable
│   ├── budgets/      → Budgeting
│   ├── workforce/    → Employee Management
│   ├── payroll/      → Payroll
│   ├── attendance/   → Time & Leave
│   ├── talent/       → Performance & Growth
│   ├── catalog/      → Products & Pricing
│   ├── inventory/    → Stock Management
│   ├── procurement/  → Purchasing
│   ├── fulfillment/  → Warehousing & Shipping
│   ├── contacts/     → Customer & Vendor Data
│   ├── pipeline/     → Sales & Opportunities
│   ├── campaigns/    → Marketing Automation
│   ├── helpdesk/     → Customer Support
│   ├── planning/     → Production Planning
│   ├── bom/          → Bill of Materials
│   ├── quality/      → Quality Management
│   ├── assets/       → Asset & Maintenance
│   ├── projects/     → Project Management
│   ├── resources/    → Resource Planning
│   ├── timesheets/   → Time & Expense Tracking
│   └── projectacct/  → Project Accounting
│
└── analytics/
    ├── reports/      → Standard Reports
    ├── dashboards/   → KPI Dashboards
    └── insights/     → Advanced Analytics
```

---

## UI Navigation Structure

This is the navigation tree customers see in the Awo interface. Every leaf maps back to a package and a route.

```
Home (/)
│
├── Accounting                          [ledger · payables · receivables · budgets]
│   ├── Overview
│   ├── Journal Entries
│   ├── Chart of Accounts
│   ├── Bills to Pay
│   ├── Invoices to Collect
│   ├── Payments
│   ├── Budgets
│   └── Financial Reports
│       ├── Balance Sheet
│       ├── Income Statement
│       └── Cash Flow
│
├── People & HR                         [workforce · payroll · attendance · talent]
│   ├── Overview
│   ├── Employees
│   ├── Departments & Positions
│   ├── Payroll
│   │   ├── Run Payroll
│   │   ├── Payslips
│   │   └── Tax Reports
│   ├── Time & Leave
│   │   ├── Attendance
│   │   ├── Leave Requests
│   │   └── Work Schedules
│   └── Performance
│       ├── Reviews
│       ├── Goals
│       └── Feedback
│
├── Inventory & Supply                  [catalog · inventory · procurement · fulfillment]
│   ├── Overview
│   ├── Products
│   ├── Stock Levels
│   ├── Stock Moves
│   ├── Purchase Orders
│   ├── Suppliers
│   ├── Warehouse
│   │   ├── Receipts
│   │   ├── Transfers
│   │   └── Dispatch
│   └── Shipping & Delivery
│
├── Sales & Customers                   [contacts · pipeline · campaigns · helpdesk]
│   ├── Overview
│   ├── Contacts
│   ├── Companies
│   ├── Leads
│   ├── Opportunities
│   ├── Quotes & Orders
│   ├── Campaigns
│   │   ├── All Campaigns
│   │   ├── Email Templates
│   │   └── Mailing Lists
│   └── Support
│       ├── Tickets
│       ├── SLA Policies
│       └── Knowledge Base
│
├── Production                          [planning · bom · quality · assets]
│   ├── Overview
│   ├── Production Orders
│   ├── Work Centres
│   ├── Bill of Materials
│   ├── Quality Control
│   │   ├── Inspections
│   │   ├── Non-Conformances
│   │   └── Quality Alerts
│   └── Assets & Maintenance
│       ├── Assets
│       ├── Maintenance Schedules
│       └── Work Orders
│
├── Projects                            [projects · resources · timesheets · projectacct]
│   ├── Overview
│   ├── All Projects
│   ├── Tasks (My Tasks)
│   ├── Milestones
│   ├── Timesheets
│   ├── Expenses
│   └── Project Finance
│       ├── Budgets
│       ├── Costs
│       └── Invoices
│
├── Analytics                           [reports · dashboards · insights]
│   ├── My Dashboard
│   ├── All Dashboards
│   ├── Reports
│   │   ├── Standard Reports
│   │   └── Custom Reports
│   └── Insights
│       ├── Trends
│       └── Forecasts
│
└── Settings                            [iam · tenant · audit · notify · bridge]
    ├── Organisation
    │   ├── Company Profile
    │   ├── Fiscal Years & Periods
    │   ├── Currencies
    │   └── Branches / Locations
    ├── Users & Access
    │   ├── Users
    │   ├── Roles
    │   ├── Permissions
    │   └── API Keys
    ├── Notifications
    │   ├── Notification Rules
    │   └── Delivery Channels
    ├── Integrations
    │   ├── Connected Apps
    │   ├── Webhooks
    │   └── API Logs
    └── Audit Log
```

---

## Platform Layer

Platform modules are the foundation. They have **zero dependencies on business modules**. Every business module may depend on platform modules.

---

### IAM — Identity & Access

**Package:** `awo/platform/iam`  
**UI Label:** Settings → Users & Access  
**Purpose:** Unified authentication, authorisation, session management, and role/permission enforcement for all user types across the entire system. Implements a combined RBAC + ABAC model with a module/resource/action permission hierarchy.

#### Unified User & Session Schema

All users — Awo staff, tenant employees, portal contacts, third-party integrations — live in one table distinguished by `user_type`. There are no separate identity stores per app surface.

```sql
-- users
id            uuid        primary key default gen_random_uuid()
email         text        unique not null
user_type     text        not null              -- 'platform' | 'tenant' | 'portal' | 'third_party'
tenant_id     uuid        references tenants(id) null  -- null for platform users
principal_id  uuid        null                  -- portal users: contact_id or employee_id
display_name  text        not null
password_hash text        not null
mfa_secret    text        null
mfa_enabled   bool        not null default false
status        text        not null default 'active'   -- 'active' | 'suspended' | 'invited'
created_at    timestamptz not null default now()
last_login_at timestamptz null

-- sessions
id            uuid        primary key default gen_random_uuid()
user_id       uuid        not null references users(id) on delete cascade
user_type     text        not null              -- denormalised from user at creation; never changes
tenant_id     uuid        null                  -- denormalised; null for platform sessions
token_hash    text        unique not null
permissions   jsonb       not null default '{}'  -- pre-computed map[resource.action]bool at login
created_at    timestamptz not null default now()
expires_at    timestamptz not null
last_seen_at  timestamptz not null default now()
ip_address    text
user_agent    text
```

`user_type` is denormalised onto `sessions` so the middleware can make security namespace decisions in one query without joining back to `users`. The `permissions` JSONB column stores the pre-computed permission map built at login time. Permission checks in handlers are an in-memory map lookup — zero DB hits per request.

**Why `principal_id` on users matters:** A portal user's identity *is* their contact or employee record. When a customer logs into the portal, `principal_id` holds their `contacts.id`. Every portal data handler reads `principal_id` from the session — it never comes from a request parameter. This is the primary security control for self-service data access.

#### Module / Resource / Action Hierarchy

Permissions follow a three-level dot-notation hierarchy. This is the canonical naming for every permission in the system.

```
{module}.{resource}.{action}

Examples:
  finance.ledger.accounts.read
  finance.ledger.journal_entries.post
  finance.receivables.invoices.approve
  finance.receivables.invoices.void

  airline.reservations.bookings.read
  airline.reservations.bookings.issue_ticket
  airline.reservations.allocations.manage
  airline.reports.commissions.export

  portal.self.invoices.read          ← self.* always scoped to principal_id by RLS
  portal.self.invoices.download
  portal.self.leave.create

  platform.tenants.read              ← platform.* only assignable to platform users
  platform.tenants.provision
  platform.flags.write
```

Every nav item, every API route, and every schema endpoint is gated by exactly one permission string.

#### `iam.AuthService`

**Models:**
```
User          { ID, TenantID?, PrincipalID?, Email, UserType, DisplayName,
                PasswordHash, MFASecret, MFAEnabled, Status, LastLoginAt }

Session       { ID, UserID, UserType, TenantID?, Token, Permissions JSONB,
                ExpiresAt, LastSeenAt, IPAddress, UserAgent }

MFAChallenge  { ID, UserID, Code, Method, ExpiresAt, Used }
OAuthProvider { ID, TenantID, Provider, ClientID, ClientSecret, CallbackURL }

// ResolvedSession is what the middleware puts into c.Locals("session").
// It is constructed from the sessions row at validation time.
ResolvedSession {
    UserID       uuid.UUID
    UserType     string            // 'platform' | 'tenant' | 'portal' | 'third_party'
    TenantID     uuid.UUID         // zero value for platform sessions
    PrincipalID  uuid.UUID         // zero value for non-portal users
    DisplayName  string
    Permissions  map[string]bool   // pre-computed from sessions.permissions JSONB
}

func (s *ResolvedSession) Can(resource, action string) bool {
    return s.Permissions[resource+"."+action]
}
func (s *ResolvedSession) IsPlatform() bool { return s.UserType == "platform" }
func (s *ResolvedSession) IsPortal()   bool { return s.UserType == "portal"   }
```

**Methods:**
```go
// Authentication — one entry point for all user types
AuthService.Login(ctx, email, password string) (*ResolvedSession, string /*token*/, error)
AuthService.ValidateMFA(ctx, userID uuid.UUID, code string) (bool, error)
AuthService.ValidateSession(ctx, token string) (*ResolvedSession, error)
AuthService.Logout(ctx, token string) error
AuthService.RefreshSession(ctx, token string) (*ResolvedSession, string, error)
AuthService.RequestPasswordReset(ctx, email string) error
AuthService.ResetPassword(ctx, token, newPassword string) error
AuthService.OAuthCallback(ctx, provider, code, state string) (*ResolvedSession, string, error)

// Permission snapshot — called at login to build the sessions.permissions JSONB
AuthService.ComputePermissions(ctx, userID uuid.UUID) (map[string]bool, error)
```

#### `iam.IdentityService`

**Models:**
```
Invitation { ID, TenantID, Email, UserType, RoleIDs, Token, ExpiresAt, AcceptedAt? }
```

**Methods:**
```go
IdentityService.CreateUser(ctx, input UserCreateInput) (*User, error)
IdentityService.InviteUser(ctx, tenantID uuid.UUID, email string, userType string, roleIDs []uuid.UUID) (*Invitation, error)
IdentityService.AcceptInvitation(ctx, token, password string) (*User, error)
IdentityService.DeactivateUser(ctx, userID uuid.UUID) error
IdentityService.GetUser(ctx, userID uuid.UUID) (*User, error)
IdentityService.ListUsers(ctx, filter UserFilter) ([]*User, error)
IdentityService.UpdateUser(ctx, userID uuid.UUID, input UserUpdateInput) (*User, error)
```

#### `iam.AccessService`

**Models:**
```
Role           { ID, TenantID?, Name, Slug, Description, IsSystem, AssignableTo []string, CreatedAt }
Permission     { ID, Module, Resource, Action, FullKey string, Description }
RolePermission { RoleID, PermissionID }
UserRole       { UserID, RoleID, GrantedBy, GrantedAt, ExpiresAt? }
Policy         { ID, TenantID, Name, Effect, Conditions JSONB }
```

The `AssignableTo` field on `Role` is the solution to the platform-user-as-tenant-admin red flag. See the [Role Assignment Guard](#role-assignment-guard) section below.

**Methods:**
```go
AccessService.CreateRole(ctx, input RoleInput) (*Role, error)
AccessService.AssignRole(ctx, userID, roleID uuid.UUID, grantedBy uuid.UUID) error
AccessService.RevokeRole(ctx, userID, roleID uuid.UUID) error
AccessService.Can(ctx, userID uuid.UUID, resource, action string) (bool, error)
AccessService.CanWithContext(ctx, userID uuid.UUID, resource, action string, attrs map[string]any) (bool, error)
AccessService.ListRoles(ctx, filter RoleFilter) ([]*Role, error)
AccessService.GetPermissionMap(ctx, userID uuid.UUID) (map[string]bool, error)
AccessService.SyncRolePermissions(ctx, roleID uuid.UUID, permissionKeys []string) error

// Nav schema generation — called at session resolution
AccessService.BuildNavPermissions(ctx, userID uuid.UUID) (map[string]bool, error)
```

**Events Emitted:** `iam.UserCreated` · `iam.UserDeactivated` · `iam.SessionStarted` · `iam.RoleAssigned` · `iam.RoleRevoked`

#### Role Assignment Guard

> **Red flag:** Nothing in the basic RBAC model prevents a tenant admin from
> assigning a `platform.*` role to a tenant user, or from assigning a tenant
> role to a platform user — either of which breaks the security namespace.

**The solution is the `AssignableTo` constraint on `Role`**, enforced at the service layer before any DB write.

```go
// awo/platform/iam/access_service.go

func (s *AccessService) AssignRole(ctx context.Context, userID, roleID, grantedBy uuid.UUID) error {
    user, err := s.getUserType(ctx, userID)
    if err != nil { return err }

    role, err := s.getRole(ctx, roleID)
    if err != nil { return err }

    // ── Guard 1: namespace match ──────────────────────────────────────────
    // A role's AssignableTo declares which user_types may hold it.
    // This is set at role creation time and cannot be changed by tenant admins.
    if !slices.Contains(role.AssignableTo, user.UserType) {
        return domain.Errorf(domain.ErrForbidden,
            "role %q (assignable to: %v) cannot be assigned to a %s user",
            role.Slug, role.AssignableTo, user.UserType)
    }

    // ── Guard 2: tenant scope ─────────────────────────────────────────────
    // A tenant admin can only assign roles that belong to their own tenant
    // (or system roles with no tenant_id). They can never assign a role owned
    // by another tenant or a platform-scoped role.
    granter, err := s.getUser(ctx, grantedBy)
    if err != nil { return err }

    if granter.UserType == "tenant" {
        // Tenant admins can only grant within their own tenant
        if role.TenantID != nil && *role.TenantID != granter.TenantID {
            return domain.Errorf(domain.ErrForbidden,
                "cannot assign a role belonging to a different tenant")
        }
        // Tenant admins can never grant platform-scoped permissions
        if s.roleContainsPlatformPermissions(ctx, roleID) {
            return domain.Errorf(domain.ErrForbidden,
                "tenant administrators cannot assign roles with platform permissions")
        }
    }

    // ── Guard 3: granter must hold the role themselves (delegation rule) ──
    // You can only grant a role if you yourself hold it, OR you are a platform
    // operator with the platform.iam.roles.assign permission.
    // This prevents privilege escalation through role chaining.
    if granter.UserType != "platform" {
        if !s.userHasRole(ctx, grantedBy, roleID) {
            return domain.Errorf(domain.ErrForbidden,
                "you can only grant roles that you yourself hold")
        }
    }

    return s.repo.InsertUserRole(ctx, UserRole{
        UserID:    userID,
        RoleID:    roleID,
        GrantedBy: grantedBy,
        GrantedAt: time.Now(),
    })
}

// roleContainsPlatformPermissions checks whether any permission in the role
// belongs to the platform.* namespace.
func (s *AccessService) roleContainsPlatformPermissions(ctx context.Context, roleID uuid.UUID) bool {
    perms, _ := s.getRolePermissions(ctx, roleID)
    for _, p := range perms {
        if strings.HasPrefix(p.Module, "platform.") {
            return true
        }
    }
    return false
}
```

**System role definitions** (seeded at provisioning, immutable by tenant admins):

```go
// awo/platform/iam/seed_roles.go

var SystemRoles = []RoleSeed{
    // Platform namespace — only assignable to platform users
    {Slug: "platform.superadmin",  AssignableTo: []string{"platform"},
     Permissions: []string{"platform.tenants.*", "platform.flags.*", "platform.iam.*"}},
    {Slug: "platform.support",     AssignableTo: []string{"platform"},
     Permissions: []string{"platform.tenants.read", "platform.users.read"}},

    // Tenant namespace — only assignable to tenant users
    {Slug: "tenant.admin",         AssignableTo: []string{"tenant"},
     Permissions: []string{"finance.*", "people.*", "inventory.*", "settings.*"}},
    {Slug: "tenant.accountant",    AssignableTo: []string{"tenant"},
     Permissions: []string{"finance.*"}},
    {Slug: "tenant.hr_manager",    AssignableTo: []string{"tenant"},
     Permissions: []string{"people.*"}},

    // Airline namespace — only assignable to tenant users with airline module
    {Slug: "airline.gsa_agent",    AssignableTo: []string{"tenant"},
     Permissions: []string{"airline.reservations.*", "airline.reports.commissions.read"}},
    {Slug: "airline.ops_manager",  AssignableTo: []string{"tenant"},
     Permissions: []string{"airline.*"}},

    // Portal namespace — only assignable to portal users
    {Slug: "portal.customer",      AssignableTo: []string{"portal"},
     Permissions: []string{"portal.self.invoices.*", "portal.self.statements.read",
                            "portal.self.tickets.*"}},
    {Slug: "portal.supplier",      AssignableTo: []string{"portal"},
     Permissions: []string{"portal.self.purchase_orders.read",
                            "portal.self.bills.read", "portal.self.payments.read"}},
    {Slug: "portal.employee",      AssignableTo: []string{"portal"},
     Permissions: []string{"portal.self.payslips.read", "portal.self.leave.*",
                            "portal.self.profile.*"}},
}
```

`AssignableTo` is set once at seeding and is never exposed in tenant-facing role management UIs. A tenant admin opening Settings → Users & Access sees only roles where `AssignableTo` contains `"tenant"`.

**Events Emitted:** `iam.UserCreated` · `iam.UserDeactivated` · `iam.SessionStarted` · `iam.RoleAssigned` · `iam.RoleRevoked`

---

### Tenant — Organisation Management

**Package:** `awo/platform/tenant`  
**UI Label:** Settings → Organisation  
**Purpose:** Manages tenant provisioning, configuration, fiscal calendar, currencies, and branches.

#### `tenant.TenantService`

**Models:**
```
Tenant       { ID, Slug, Name, Plan, Status, CreatedAt, TrialEndsAt? }
TenantConfig { TenantID, DefaultCurrency, DefaultTimezone, DateFormat, FiscalYearStart, LogoURL }
Branch       { ID, TenantID, Name, Code, Address, Phone, IsHeadOffice }
```

**Methods:**
```go
TenantService.ProvisionTenant(ctx, data TenantCreateInput) (*Tenant, error)
TenantService.GetTenant(ctx, tenantID uuid.UUID) (*Tenant, error)
TenantService.UpdateConfig(ctx, tenantID uuid.UUID, config TenantConfig) error
TenantService.SuspendTenant(ctx, tenantID uuid.UUID, reason string) error
TenantService.CreateBranch(ctx, tenantID uuid.UUID, data BranchInput) (*Branch, error)
TenantService.ListBranches(ctx, tenantID uuid.UUID) ([]*Branch, error)
```

#### `tenant.CalendarService`

**Models:**
```
FiscalYear       { ID, TenantID, Name, StartDate, EndDate, Status }
AccountingPeriod { ID, FiscalYearID, Name, StartDate, EndDate, Status, ClosedAt? }
Currency         { Code, Name, Symbol, DecimalPlaces }
ExchangeRate     { ID, TenantID, FromCurrency, ToCurrency, Rate, EffectiveDate }
```

**Methods:**
```go
CalendarService.CreateFiscalYear(ctx, tenantID uuid.UUID, start, end time.Time) (*FiscalYear, error)
CalendarService.OpenPeriod(ctx, periodID uuid.UUID) error
CalendarService.ClosePeriod(ctx, periodID uuid.UUID) error
CalendarService.GetActivePeriod(ctx, tenantID uuid.UUID, date time.Time) (*AccountingPeriod, error)
CalendarService.SetExchangeRate(ctx, tenantID uuid.UUID, from, to string, rate float64, date time.Time) error
CalendarService.GetExchangeRate(ctx, tenantID uuid.UUID, from, to string, date time.Time) (float64, error)
```

**Events Emitted:** `tenant.TenantProvisioned` · `tenant.PeriodClosed` · `tenant.FiscalYearCreated`

---

### Audit — Audit & Compliance Log

**Package:** `awo/platform/audit`  
**UI Label:** Settings → Audit Log  
**Purpose:** Immutable append-only record of every state-changing operation across all modules.

#### `audit.LogService`

**Models:**
```
AuditEntry { ID, TenantID, UserID, SessionID, Module, Resource, ResourceID,
             Action, Before JSONB, After JSONB, IPAddress, OccurredAt }
```

**Methods:**
```go
LogService.Record(ctx, entry AuditInput) error
LogService.Query(ctx, filter AuditFilter) ([]*AuditEntry, Pagination, error)
LogService.GetHistory(ctx, resource string, resourceID uuid.UUID) ([]*AuditEntry, error)
LogService.ExportCompliance(ctx, tenantID uuid.UUID, dateRange DateRange, format string) ([]byte, error)
```

**Design Note:** `LogService.Record` is called by a middleware wrapper, never by application code directly. All write operations automatically emit an audit entry.

---

### Notify — Notification Hub

**Package:** `awo/platform/notify`  
**UI Label:** In-app bell icon + Settings → Notifications  
**Purpose:** Centralised delivery of in-app, email, SMS, and webhook notifications. Subscribes to domain events from all business modules.

#### `notify.NotificationService`

**Models:**
```
NotificationRule     { ID, TenantID, EventType, ChannelType, TemplateID, Recipients }
NotificationTemplate { ID, TenantID, Name, Channel, Subject?, BodyTemplate, Variables }
Notification         { ID, TenantID, UserID, Title, Body, Link?, ReadAt?, CreatedAt }
DeliveryLog          { ID, NotificationID, Channel, Status, SentAt, FailureReason? }
```

**Methods:**
```go
NotificationService.Dispatch(ctx, event DomainEvent) error
NotificationService.MarkRead(ctx, notificationID uuid.UUID) error
NotificationService.GetInbox(ctx, userID uuid.UUID) ([]*Notification, error)
NotificationService.CreateRule(ctx, rule NotificationRuleInput) (*NotificationRule, error)
NotificationService.UpdateTemplate(ctx, templateID uuid.UUID, body string) error
NotificationService.GetUnreadCount(ctx, userID uuid.UUID) (int, error)
```

---

### Bridge — Integration Gateway

**Package:** `awo/platform/bridge`  
**UI Label:** Settings → Integrations  
**Purpose:** Manages outbound webhooks, inbound API key authentication, third-party connections, and an HTTP API log.

#### `bridge.WebhookService`

**Models:**
```
WebhookEndpoint { ID, TenantID, URL, SecretKey, Events []string, Active, CreatedAt }
WebhookDelivery { ID, EndpointID, EventType, Payload JSONB, StatusCode?, SentAt, Success }
```

**Methods:**
```go
WebhookService.RegisterEndpoint(ctx, tenantID uuid.UUID, url string, events []string) (*WebhookEndpoint, error)
WebhookService.Deliver(ctx, endpointID uuid.UUID, event DomainEvent) error
WebhookService.RetryFailed(ctx, deliveryID uuid.UUID) error
WebhookService.ListDeliveries(ctx, endpointID uuid.UUID, filter DeliveryFilter) ([]*WebhookDelivery, error)
```

#### `bridge.APIKeyService`

**Models:**
```
APIKey { ID, TenantID, Name, KeyHash, Scopes []string, LastUsedAt?, ExpiresAt?, CreatedAt }
```

**Methods:**
```go
APIKeyService.Issue(ctx, tenantID uuid.UUID, name string, scopes []string) (plainKey string, *APIKey, error)
APIKeyService.Revoke(ctx, keyID uuid.UUID) error
APIKeyService.Validate(ctx, plainKey string) (*APIKey, error)
APIKeyService.List(ctx, tenantID uuid.UUID) ([]*APIKey, error)
```

---

## Identity, Database Security & Multi-App Architecture

### Overview

Every user-facing surface in Awo — the tenant ERP, the platform ops console, customer portals, supplier portals, airline GSA consoles, forecourt manager dashboards, or any future app — is served by the same database, the same Go binary, and the same `index.html`. There are no separate applications. The permission set a user holds after authentication *is* their application. A user who holds only `airline.reservations.*` permissions sees an app that looks exactly like an airline booking system. A user who holds only `portal.self.invoices.read` sees an app that looks like a customer invoice portal. Neither of those apps was explicitly built — the permission set built them.

```
One binary. One database. One index.html.
Permission set → navigation → visible schema → accessible data.
Nothing else determines what a user sees or can do.
```

---

### Database Layer — Two PostgreSQL Roles

The most fundamental security boundary in the system is at the database connection level. Two PostgreSQL roles are used, never mixed:

```sql
-- Role 1: awo_app
-- Used for all tenant and portal requests.
-- Subject to Row-Level Security on every table.
-- Cannot query cross-tenant views.
-- Cannot SET app.user_type = 'platform'.
CREATE ROLE awo_app LOGIN PASSWORD '...' NOINHERIT;
GRANT CONNECT ON DATABASE awo TO awo_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO awo_app;
REVOKE ALL ON platform_views FROM awo_app;  -- explicit deny on cross-tenant views

-- Role 2: awo_platform
-- Used exclusively for platform (Awo staff) requests.
-- BYPASSRLS: Row-Level Security is entirely skipped at the DB engine level.
-- Not a superuser — still subject to table-level GRANT/REVOKE.
-- Cross-tenant views are accessible to this role only.
CREATE ROLE awo_platform LOGIN PASSWORD '...' NOINHERIT BYPASSRLS;
GRANT CONNECT ON DATABASE awo TO awo_platform;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO awo_platform;
GRANT SELECT ON ALL TABLES IN SCHEMA platform_views TO awo_platform;
```

`BYPASSRLS` at the PostgreSQL role level means RLS policies are not evaluated at all for `awo_platform` connections — not suppressed by a policy condition, genuinely skipped by the engine. This is fundamentally more robust than a policy that checks `current_setting('app.user_type') = 'platform'` because it cannot be bypassed by setting a session variable.

**Two connection pools in Go**, one per database role:

```go
// awo/infrastructure/db/pools.go

type Pools struct {
    App      *pgxpool.Pool  // awo_app role — used by all tenant/portal handlers
    Platform *pgxpool.Pool  // awo_platform role — used by platform handlers only
}

func NewPools(cfg Config) (*Pools, error) {
    app, err := pgxpool.New(ctx, cfg.AppDSN)      // DSN uses awo_app credentials
    if err != nil { return nil, err }

    platform, err := pgxpool.New(ctx, cfg.PlatformDSN) // DSN uses awo_platform credentials
    if err != nil { return nil, err }

    return &Pools{App: app, Platform: platform}, nil
}
```

The middleware selects the pool based on `user_type` from the session. Platform handlers only ever receive the `awo_platform` pool. Tenant and portal handlers only ever receive the `awo_app` pool. No handler has access to both pools — the pool is injected through `Deps`.

```go
// awo/web/middleware/db_context.go

func SetDBPool(pools *db.Pools) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := ContextSession(c)

        if session.IsPlatform() {
            c.Locals("db", pools.Platform)
        } else {
            c.Locals("db", pools.App)
            // Set per-connection session variables for RLS policies
            conn := pools.App.AcquireFunc(c.Context(), func(conn *pgxpool.Conn) error {
                _, err := conn.Exec(c.Context(),
                    `SELECT set_config('app.tenant_id',   $1, true),
                            set_config('app.user_id',     $2, true),
                            set_config('app.user_type',   $3, true),
                            set_config('app.principal_id',$4, true)`,
                    session.TenantID.String(),
                    session.UserID.String(),
                    session.UserType,
                    session.PrincipalID.String(), // empty string for non-portal users
                )
                return err
            })
            _ = conn
        }
        return c.Next()
    }
}
```

---

### Row-Level Security Policies

RLS policies on the `awo_app` role enforce tenant isolation and portal self-isolation. The `awo_platform` role bypasses all of these at the engine level.

```sql
-- Enable RLS on all tenant-scoped tables
ALTER TABLE invoices        ENABLE ROW LEVEL SECURITY;
ALTER TABLE purchase_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE employees       ENABLE ROW LEVEL SECURITY;
-- (applied to every table in the schema)

-- Policy 1: Tenant isolation
-- Every tenant user sees only their own tenant's rows.
CREATE POLICY tenant_isolation ON invoices
    AS PERMISSIVE FOR ALL TO awo_app
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

-- Policy 2: Portal self-isolation
-- Portal users see only rows where they are the named party.
-- This is an additional AND on top of tenant_isolation.
CREATE POLICY portal_self_invoices ON invoices
    AS RESTRICTIVE FOR SELECT TO awo_app
    USING (
        current_setting('app.user_type', true) != 'portal'
        OR customer_id = current_setting('app.principal_id', true)::uuid
    );

-- Policy 3: Immutable audit entries
-- No user (including awo_app) can UPDATE or DELETE audit entries.
CREATE POLICY audit_immutable ON audit_entries
    AS RESTRICTIVE FOR ALL TO awo_app
    USING (true)
    WITH CHECK (false);  -- blocks UPDATE and DELETE entirely
```

**Platform views** — cross-tenant read-only views, accessible only to `awo_platform`:

```sql
-- Schema visible only to awo_platform
CREATE SCHEMA platform_views;

CREATE VIEW platform_views.all_invoices AS
    SELECT i.*, t.name AS tenant_name, t.slug AS tenant_slug
    FROM invoices i
    JOIN tenants t ON t.id = i.tenant_id;

CREATE VIEW platform_views.all_users AS
    SELECT u.*, t.name AS tenant_name
    FROM users u
    LEFT JOIN tenants t ON t.id = u.tenant_id;

-- Grant: awo_platform can read; awo_app cannot
GRANT USAGE ON SCHEMA platform_views TO awo_platform;
GRANT SELECT ON ALL TABLES IN SCHEMA platform_views TO awo_platform;
REVOKE ALL ON SCHEMA platform_views FROM awo_app;
```

When a platform user requests data, the service layer uses the `awo_platform` pool and queries `platform_views.*` tables directly. When a tenant user requests data, the service layer uses the `awo_app` pool and queries the base tables — RLS policies enforce that they only see their tenant's rows.

---

### X-Tenant-ID Header Contract

```
Request type          | X-Tenant-ID required? | Pool used      | Behaviour
──────────────────────┼───────────────────────┼────────────────┼────────────────────────────────────
Platform user         | Optional              | awo_platform   | If present: scopes display to that
                      |                       | (BYPASSRLS)    | tenant. If absent: global view.
──────────────────────┼───────────────────────┼────────────────┼────────────────────────────────────
Tenant user           | Required              | awo_app        | Must match session.TenantID exactly.
                      |                       | (RLS active)   | Mismatch → 403.
──────────────────────┼───────────────────────┼────────────────┼────────────────────────────────────
Portal user           | Required              | awo_app        | Same as tenant user. RLS also
                      |                       | (RLS active)   | applies portal_self_* policies.
──────────────────────┼───────────────────────┼────────────────┼────────────────────────────────────
Third-party (API key) | Required              | awo_app        | Validated against the API key's
                      |                       | (RLS active)   | registered tenant_id.
```

```go
// awo/web/middleware/tenant_context.go

func RequireTenantContext(tenantSvc *tenant.TenantService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := ContextSession(c)
        header  := c.Get("X-Tenant-ID")

        if session.IsPlatform() {
            // Platform: X-Tenant-ID is optional context, not a restriction
            if header != "" {
                t, err := tenantSvc.Get(c.Context(), uuid.MustParse(header))
                if err != nil { return fiber.ErrNotFound }
                c.Locals("tenant", t)
            }
            return c.Next()
        }

        // All other user types: header is mandatory
        if header == "" {
            return c.Status(403).JSON(response.Err("X-Tenant-ID header is required"))
        }
        if header != session.TenantID.String() {
            return c.Status(403).JSON(response.Err("Tenant mismatch"))
        }

        c.Locals("tenant", session.Tenant)
        return c.Next()
    }
}
```

---

### Single Entry Point — Boot Endpoint

There is one `index.html` for all app surfaces. It loads one schema endpoint unconditionally:

```javascript
// awo/web/static/boot.js
embed('#root',
    { type: 'service', schemaApi: '/schema/boot', id: 'root-service' },
    {}, env);
```

`/schema/boot` is the only unauthenticated schema endpoint. It checks the session cookie and returns either the login form or the full app shell built from the user's permission set:

```go
// awo/web/handlers/schema/boot.go

func (h *BootHandler) Boot(c *fiber.Ctx) error {
    token := c.Cookies("awo_session")
    if token == "" {
        return c.JSON(h.loginPage(""))
    }

    session, err := h.authSvc.ValidateSession(c.Context(), token)
    if err != nil {
        return c.JSON(h.loginPage("Your session has expired. Please sign in again."))
    }

    // Build app shell entirely from the session's permission map.
    // No app type switch — permissions drive everything.
    nav := h.navRegistry.BuildNav(session.Permissions)

    return c.JSON(map[string]any{
        "type":      "app",
        "brandName": "Awo",
        "logo":      "/static/logo.svg",
        "header":    map[string]any{"type": "tpl", "tpl": session.DisplayName},
        "pages":     nav,
    })
}
```

After successful login, the `/auth/login` handler sets the session cookie and returns a reload action targeting `root-service`. amis re-calls `/schema/boot`, which now finds a valid session and returns the app shell — the login form is replaced by the full UI with no page navigation.

---

### Permission-Driven Navigation

The nav registry maps permission keys to nav items. A section only appears if the user has at least one readable resource within it. The entire multi-app routing problem reduces to this loop:

```go
// awo/web/nav/registry.go

type NavItem struct {
    Label      string
    URL        string
    Permission string  // "module.resource.action" required to see this item
}

type NavSection struct {
    Label string
    Icon  string
    Items []NavItem
}

// All nav items across all modules and app surfaces — registered once.
// Modules call RegisterSection at startup via their Wire provider.
var defaultSections = []NavSection{
    {Label: "Accounting", Icon: "fa fa-calculator", Items: []NavItem{
        {Label: "Journal",   URL: "/accounting/journal",   Permission: "finance.ledger.journal_entries.read"},
        {Label: "Accounts",  URL: "/accounting/accounts",  Permission: "finance.ledger.accounts.read"},
        {Label: "Invoices",  URL: "/accounting/invoices",  Permission: "finance.receivables.invoices.read"},
        {Label: "Bills",     URL: "/accounting/bills",     Permission: "finance.payables.bills.read"},
    }},
    {Label: "Airline Bookings", Icon: "fa fa-plane", Items: []NavItem{
        {Label: "Bookings",    URL: "/airline/bookings",    Permission: "airline.reservations.bookings.read"},
        {Label: "Allocations", URL: "/airline/allocations", Permission: "airline.reservations.allocations.read"},
        {Label: "Commissions", URL: "/airline/commissions", Permission: "airline.reports.commissions.read"},
    }},
    {Label: "Platform", Icon: "fa fa-server", Items: []NavItem{
        {Label: "Tenants",  URL: "/platform/tenants", Permission: "platform.tenants.read"},
        {Label: "Users",    URL: "/platform/users",   Permission: "platform.users.read"},
        {Label: "Flags",    URL: "/platform/flags",   Permission: "platform.flags.read"},
        {Label: "Jobs",     URL: "/platform/jobs",    Permission: "platform.system.jobs.read"},
    }},
    {Label: "My Account", Icon: "fa fa-user", Items: []NavItem{
        {Label: "My Invoices", URL: "/portal/invoices", Permission: "portal.self.invoices.read"},
        {Label: "My Payslips", URL: "/portal/payslips", Permission: "portal.self.payslips.read"},
        {Label: "Leave",       URL: "/portal/leave",    Permission: "portal.self.leave.read"},
    }},
    // ... every module registers its section
}

func (r *NavRegistry) BuildNav(perms map[string]bool) []any {
    var result []any
    for _, section := range r.sections {
        var visible []any
        for _, item := range section.Items {
            if perms[item.Permission] {
                visible = append(visible, map[string]any{
                    "label": item.Label,
                    "url":   item.URL,
                    "schema": map[string]any{
                        "type":      "service",
                        "schemaApi": "/schema" + item.URL,
                        "fallback":  map[string]any{
                            "type": "alert", "level": "danger",
                            "body": "Failed to load " + item.Label + ". Please refresh.",
                        },
                    },
                })
            }
        }
        if len(visible) > 0 {
            result = append(result, map[string]any{
                "label":    section.Label,
                "icon":     section.Icon,
                "children": visible,
            })
        }
    }
    return result
}
```

A GSA agent whose `Permissions` map contains only `airline.*` keys gets a nav with one section: "Airline Bookings". A platform support engineer whose permissions contain only `platform.tenants.read` and `platform.users.read` gets a nav with one section: "Platform" showing only Tenants and Users. They are using the exact same code — the permission map is doing all the work.

---

### Schema & API Enforcement

Every schema endpoint and every data API route enforces the same permission:

```go
// Schema endpoint
sg.Get("/accounting/invoices",
    middleware.RequirePermission("finance.receivables.invoices", "read"),
    schema.InvoiceList(deps))

// Data API endpoint
api.Get("/invoices",
    middleware.RequirePermission("finance.receivables.invoices", "read"),
    handlers.ListInvoices(deps))

api.Post("/invoices/:id/post",
    middleware.RequirePermission("finance.receivables.invoices", "post"),
    handlers.PostInvoice(deps))
```

`RequirePermission` reads from the pre-computed `session.Permissions` map — zero DB hits:

```go
// awo/web/middleware/permission.go

func RequirePermission(resource, action string) fiber.Handler {
    key := resource + "." + action
    return func(c *fiber.Ctx) error {
        if !ContextSession(c).Can(resource, action) {
            // Return 403 — not 404. A 404 would hide that the resource exists,
            // but the permission check already reveals that. 403 is honest.
            return c.Status(403).JSON(response.Err("Access denied"))
        }
        return c.Next()
    }
}
```

If a user doesn't have a permission, they cannot discover the nav item (not rendered), cannot load the schema (403), and cannot call the API (403). Not seeing a thing is equivalent to it not existing — the architecture enforces this at three independent layers.

---

### Middleware Stack (Updated)

```go
// awo/web/router/router.go

func Build(app *fiber.App, deps *Deps) {
    // ── Infrastructure (always on) ────────────────────────────────────────
    app.Use(middleware.Recovery())
    app.Use(middleware.Logger())
    app.Use(middleware.RateLimit())

    // ── Unauthenticated routes ────────────────────────────────────────────
    app.Get("/schema/boot",   handlers.Boot(deps))      // returns login form or app shell
    app.Post("/auth/login",   handlers.Login(deps))
    app.Post("/auth/logout",  handlers.Logout(deps))
    app.Static("/static",     "./web/static")
    app.Get("/",              handlers.Index(deps))      // serves index.html

    // ── Authenticated routes ──────────────────────────────────────────────
    auth := app.Group("",
        middleware.Authenticate(deps.Auth),       // validates session cookie or Bearer token
        middleware.SetDBPool(deps.Pools),          // selects awo_app or awo_platform pool
        middleware.RequireTenantContext(deps.Tenant), // validates X-Tenant-ID header
        middleware.InjectFlags(deps.Flags),        // loads resolved feature flags into ctx
        middleware.AuditWrap(deps.Audit),          // records mutating requests
    )

    // Schema routes — one per module page
    sg := auth.Group("/schema")
    sg.Get("/app",                        handlers.AppShell(deps))
    sg.Get("/accounting/journal",
        middleware.RequirePermission("finance.ledger.journal_entries", "read"),
        schema.Journal(deps))
    sg.Get("/accounting/invoices",
        middleware.RequirePermission("finance.receivables.invoices", "read"),
        schema.InvoiceList(deps))
    sg.Get("/airline/bookings",
        middleware.RequirePermission("airline.reservations.bookings", "read"),
        schema.AirlineBookings(deps))
    sg.Get("/platform/tenants",
        middleware.RequirePermission("platform.tenants", "read"),
        schema.PlatformTenants(deps))
    // ... all modules

    // Data API routes
    api := auth.Group("/api/v1")
    registerFinanceAPI(api, deps)
    registerAirlineAPI(api, deps)
    registerPlatformAPI(api, deps)
    // ... all modules
}
```

---

### Summary: Security Layers

Five independent layers enforce access control. A request must pass all five:

```
1. DB role (awo_app vs awo_platform)
   └── Platform requests bypass RLS at the PostgreSQL engine level.
       Tenant/portal requests are subject to all RLS policies.
       This cannot be circumvented from application code.

2. X-Tenant-ID header check
   └── Non-platform users must present a header matching their session tenant.
       Mismatch → 403 before any handler runs.

3. Role assignment guards (AssignableTo + delegation rule)
   └── Platform permissions cannot be assigned to tenant users.
       Tenant permissions cannot be assigned to platform users.
       A user cannot grant a role they don't hold themselves.

4. Permission middleware on every route
   └── session.Permissions map checked on every schema + API request.
       Pre-computed at login — zero DB hits per check.

5. RLS policies on awo_app connections
   └── Even if layers 1–4 were somehow bypassed, the DB would still
       only return rows belonging to the session's tenant_id,
       and portal rows belonging to the session's principal_id.
```

---

## Business Layer — Finance

---

### Ledger — General Ledger

**Package:** `awo/core/ledger`  
**UI Label:** Accounting → Journal Entries · Chart of Accounts  
**Purpose:** Core double-entry bookkeeping engine. All financial transactions ultimately post here. Other modules post via the `ledger.PostingPort` interface — they never import this package directly.

#### `ledger.AccountService`

**Models:**
```
Account      { ID, TenantID, Code, Name, Type (asset|liability|equity|revenue|expense),
               ParentID?, CurrencyCode, Active, CreatedAt }
AccountGroup { ID, TenantID, Name, Type, SortOrder }
CostCentre   { ID, TenantID, Name, Code, ManagerID? }
```

**Methods:**
```go
AccountService.CreateAccount(ctx, input AccountInput) (*Account, error)
AccountService.UpdateAccount(ctx, accountID uuid.UUID, input AccountInput) (*Account, error)
AccountService.DeactivateAccount(ctx, accountID uuid.UUID) error
AccountService.GetAccount(ctx, accountID uuid.UUID) (*Account, error)
AccountService.ListAccounts(ctx, filter AccountFilter) ([]*Account, error)
AccountService.GetBalance(ctx, accountID uuid.UUID, asOf time.Time) (decimal.Decimal, error)
AccountService.ImportChartOfAccounts(ctx, tenantID uuid.UUID, rows []AccountImportRow) error
```

#### `ledger.JournalService`

**Models:**
```
Journal      { ID, TenantID, Name, Code, Type (general|sales|purchase|cash|payroll) }
JournalEntry { ID, TenantID, JournalID, Reference, Description, Date, PeriodID,
               Status (draft|posted|reversed), PostedAt?, ReversalOfID?, CreatedBy }
JournalLine  { ID, EntryID, AccountID, Label, Debit, Credit, CurrencyCode,
               CurrencyAmount?, ExchangeRate?, TaxID?, CostCentreID? }
```

**Methods:**
```go
JournalService.CreateEntry(ctx, input JournalEntryInput) (*JournalEntry, error)
JournalService.PostEntry(ctx, entryID uuid.UUID) error
JournalService.ReverseEntry(ctx, entryID uuid.UUID, date time.Time, reason string) (*JournalEntry, error)
JournalService.GetEntry(ctx, entryID uuid.UUID) (*JournalEntry, error)
JournalService.ListEntries(ctx, filter EntryFilter) ([]*JournalEntry, Pagination, error)
JournalService.GetTrialBalance(ctx, tenantID uuid.UUID, asOf time.Time) ([]*TrialBalanceLine, error)
```

#### `ledger.ReportingService`

**Models:**
```
FinancialReport { ID, TenantID, Type, Parameters JSONB, GeneratedAt, Data JSONB }
```

**Methods:**
```go
ReportingService.BalanceSheet(ctx, tenantID uuid.UUID, asOf time.Time) (*BalanceSheet, error)
ReportingService.IncomeStatement(ctx, tenantID uuid.UUID, period DateRange) (*IncomeStatement, error)
ReportingService.CashFlowStatement(ctx, tenantID uuid.UUID, period DateRange) (*CashFlowStatement, error)
ReportingService.AgedBalances(ctx, tenantID uuid.UUID, asOf time.Time, side string) ([]*AgedBalance, error)
```

#### `ledger.PostingPort` — Anti-Corruption Interface

This interface is defined in `awo/core/ledger/port.go` and implemented by `ledger.JournalService`. All other modules that need to post financial entries receive this interface — they never import `JournalService` directly.

```go
type PostingPort interface {
    PostLines(ctx context.Context, req PostingRequest) (entryID uuid.UUID, err error)
}

type PostingRequest struct {
    TenantID     uuid.UUID
    Journal      string        // e.g. "PAYROLL", "PURCHASE", "SALES"
    Reference    string
    Description  string
    Date         time.Time
    Lines        []PostingLine
    SourceModule string        // e.g. "payroll", "procurement"
    SourceID     uuid.UUID
}
```

**Events Emitted:** `ledger.EntryPosted` · `ledger.PeriodClosed`

---

### Payables — Accounts Payable

**Package:** `awo/core/payables`  
**UI Label:** Accounting → Bills to Pay  
**Purpose:** Manages supplier bills, payment runs, and vendor credit notes.

#### `payables.BillService`

**Models:**
```
Bill         { ID, TenantID, SupplierID, Reference, BillDate, DueDate, PeriodID,
               Status (draft|posted|partiallyPaid|paid|cancelled), Total, AmountDue,
               CurrencyCode, JournalEntryID? }
BillLine     { ID, BillID, Description, AccountID, Quantity, UnitPrice, TaxID?, Amount }
VendorCredit { ID, TenantID, SupplierID, Reference, Date, Total, Remaining }
```

**Methods:**
```go
BillService.CreateBill(ctx, input BillInput) (*Bill, error)
BillService.PostBill(ctx, billID uuid.UUID) error
BillService.GetBill(ctx, billID uuid.UUID) (*Bill, error)
BillService.ListBills(ctx, filter BillFilter) ([]*Bill, Pagination, error)
BillService.CancelBill(ctx, billID uuid.UUID, reason string) error
BillService.CreateVendorCredit(ctx, input VendorCreditInput) (*VendorCredit, error)
BillService.ApplyCreditToBill(ctx, creditID, billID uuid.UUID, amount decimal.Decimal) error
BillService.AgingReport(ctx, tenantID uuid.UUID, asOf time.Time) ([]*AgingLine, error)
```

#### `payables.PaymentService`

**Models:**
```
SupplierPayment   { ID, TenantID, SupplierID, Date, Amount, CurrencyCode,
                    PaymentMethod, Reference, JournalEntryID? }
PaymentAllocation { PaymentID, BillID, Amount }
PaymentRun        { ID, TenantID, Name, DueBy time.Time, Status, CreatedAt }
PaymentRunItem    { RunID, BillID, SupplierID, Amount, Selected }
```

**Methods:**
```go
PaymentService.RecordPayment(ctx, input PaymentInput) (*SupplierPayment, error)
PaymentService.AllocatePayment(ctx, paymentID uuid.UUID, allocations []PaymentAllocation) error
PaymentService.BuildPaymentRun(ctx, tenantID uuid.UUID, dueBy time.Time) (*PaymentRun, error)
PaymentService.ProcessPaymentRun(ctx, runID uuid.UUID) error
PaymentService.ListPayments(ctx, filter PaymentFilter) ([]*SupplierPayment, Pagination, error)
```

**Events Emitted:** `payables.BillPosted` · `payables.PaymentRecorded`

---

### Receivables — Accounts Receivable

**Package:** `awo/core/receivables`  
**UI Label:** Accounting → Invoices to Collect  
**Purpose:** Manages customer invoices, receipts, credit notes, and collections.

#### `receivables.InvoiceService`

**Models:**
```
Invoice        { ID, TenantID, CustomerID, Reference, InvoiceDate, DueDate, PeriodID,
                 Status (draft|sent|partiallyPaid|paid|overdue|void), Total, AmountDue,
                 CurrencyCode, JournalEntryID?, SentAt? }
InvoiceLine    { ID, InvoiceID, Description, AccountID, Quantity, UnitPrice, TaxID?, Amount }
CustomerCredit { ID, TenantID, CustomerID, Reference, Date, Total, Remaining }
```

**Methods:**
```go
InvoiceService.CreateInvoice(ctx, input InvoiceInput) (*Invoice, error)
InvoiceService.PostInvoice(ctx, invoiceID uuid.UUID) error
InvoiceService.SendInvoice(ctx, invoiceID uuid.UUID, recipients []string) error
InvoiceService.VoidInvoice(ctx, invoiceID uuid.UUID, reason string) error
InvoiceService.GetInvoice(ctx, invoiceID uuid.UUID) (*Invoice, error)
InvoiceService.ListInvoices(ctx, filter InvoiceFilter) ([]*Invoice, Pagination, error)
InvoiceService.SendReminder(ctx, invoiceID uuid.UUID) error
InvoiceService.CreateCustomerCredit(ctx, input CustomerCreditInput) (*CustomerCredit, error)
InvoiceService.AgingReport(ctx, tenantID uuid.UUID, asOf time.Time) ([]*AgingLine, error)
```

#### `receivables.ReceiptService`

**Models:**
```
Receipt           { ID, TenantID, CustomerID, Date, Amount, CurrencyCode,
                    PaymentMethod, Reference, JournalEntryID? }
ReceiptAllocation { ReceiptID, InvoiceID, Amount }
```

**Methods:**
```go
ReceiptService.RecordReceipt(ctx, input ReceiptInput) (*Receipt, error)
ReceiptService.AllocateReceipt(ctx, receiptID uuid.UUID, allocations []ReceiptAllocation) error
ReceiptService.ListReceipts(ctx, filter ReceiptFilter) ([]*Receipt, Pagination, error)
ReceiptService.GetReceipt(ctx, receiptID uuid.UUID) (*Receipt, error)
```

**Events Emitted:** `receivables.InvoicePosted` · `receivables.InvoiceSent` · `receivables.ReceiptRecorded` · `receivables.InvoiceOverdue`

---

### Budgets — Budgeting & Forecasting

**Package:** `awo/core/budgets`  
**UI Label:** Accounting → Budgets  
**Purpose:** Plan, allocate, and compare budgeted vs. actual financial performance.

#### `budgets.BudgetService`

**Models:**
```
Budget         { ID, TenantID, Name, FiscalYearID, Type (department|project|overall),
                 Status (draft|approved|active|closed) }
BudgetLine     { ID, BudgetID, AccountID, CostCentreID?, PeriodID, PlannedAmount, Notes }
BudgetRevision { ID, BudgetID, RevisedBy, RevisedAt, ChangeSummary }
BudgetScenario { ID, BudgetID, Name, Assumption, Lines []BudgetLine }
```

**Methods:**
```go
BudgetService.CreateBudget(ctx, input BudgetInput) (*Budget, error)
BudgetService.ApproveBudget(ctx, budgetID uuid.UUID) error
BudgetService.AllocateLine(ctx, budgetID, accountID, periodID uuid.UUID, amount decimal.Decimal) error
BudgetService.ReviseBudget(ctx, budgetID uuid.UUID, changes []BudgetLineChange, summary string) error
BudgetService.CompareActualVsBudget(ctx, budgetID uuid.UUID) ([]*BudgetVarianceLine, error)
BudgetService.CreateScenario(ctx, budgetID uuid.UUID, name string) (*BudgetScenario, error)
BudgetService.ListBudgets(ctx, filter BudgetFilter) ([]*Budget, error)
```

---

## Business Layer — People

---

### Workforce — Employee Management

**Package:** `awo/core/workforce`  
**UI Label:** People & HR → Employees  
**Purpose:** Manages the full employee lifecycle from hire to separation.

#### `workforce.EmployeeService`

**Models:**
```
Employee           { ID, TenantID, EmployeeNumber, UserID?, FirstName, LastName,
                     DateOfBirth, Gender, NationalID, Nationality, Photo?,
                     HireDate, ProbationEndDate?, Status (active|onLeave|terminated) }
Department         { ID, TenantID, Name, Code, ManagerID?, ParentID? }
JobPosition        { ID, TenantID, Title, DepartmentID, Grade, MinSalary, MaxSalary }
EmploymentContract { ID, EmployeeID, Type (permanent|contract|partTime), StartDate, EndDate?,
                     JobPositionID, DepartmentID, DirectManagerID, WorkLocationID }
EmployeeDocument   { ID, EmployeeID, Type, Name, FilePath, ExpiresAt?, UploadedAt }
Termination        { ID, EmployeeID, Date, Reason, Type (resignation|dismissal|redundancy), Notes }
```

**Methods:**
```go
EmployeeService.HireEmployee(ctx, input HireInput) (*Employee, error)
EmployeeService.GetEmployee(ctx, employeeID uuid.UUID) (*Employee, error)
EmployeeService.ListEmployees(ctx, filter EmployeeFilter) ([]*Employee, Pagination, error)
EmployeeService.UpdateEmployee(ctx, employeeID uuid.UUID, input EmployeeUpdateInput) error
EmployeeService.TerminateEmployee(ctx, employeeID uuid.UUID, input TerminationInput) error
EmployeeService.TransferEmployee(ctx, employeeID, newDeptID, newPositionID uuid.UUID, date time.Time) error
EmployeeService.UploadDocument(ctx, employeeID uuid.UUID, doc DocumentInput) (*EmployeeDocument, error)
EmployeeService.GetOrgChart(ctx, tenantID uuid.UUID) ([]*OrgNode, error)
EmployeeService.SearchDirectory(ctx, tenantID uuid.UUID, query string) ([]*Employee, error)
```

#### `workforce.PositionService`

**Methods:**
```go
PositionService.CreateDepartment(ctx, input DeptInput) (*Department, error)
PositionService.CreatePosition(ctx, input PositionInput) (*JobPosition, error)
PositionService.UpdatePosition(ctx, positionID uuid.UUID, input PositionInput) error
PositionService.ListDepartments(ctx, tenantID uuid.UUID) ([]*Department, error)
PositionService.ListPositions(ctx, tenantID uuid.UUID, filter PositionFilter) ([]*JobPosition, error)
```

**Events Emitted:** `workforce.EmployeeHired` · `workforce.EmployeeTerminated` · `workforce.EmployeeTransferred`

---

### Payroll — Payroll Processing

**Package:** `awo/core/payroll`  
**UI Label:** People & HR → Payroll  
**Purpose:** Calculates gross pay, deductions, net pay, and statutory tax submissions.

#### `payroll.RunService`

**Models:**
```
PayrollRun { ID, TenantID, Name, PeriodStart, PeriodEnd,
             Status (draft|processing|approved|paid|closed),
             TotalGross, TotalDeductions, TotalNet, ProcessedAt?, ApprovedBy? }
Payslip    { ID, RunID, EmployeeID, GrossPay, TotalDeductions, NetPay,
             Status (draft|finalised), Details JSONB }
```

**Methods:**
```go
RunService.CreateRun(ctx, tenantID uuid.UUID, periodStart, periodEnd time.Time) (*PayrollRun, error)
RunService.ProcessRun(ctx, runID uuid.UUID) error
RunService.ReviewRun(ctx, runID uuid.UUID) ([]*Payslip, error)
RunService.ApproveRun(ctx, runID, approverID uuid.UUID) error
RunService.PostToLedger(ctx, runID uuid.UUID) error
RunService.ListRuns(ctx, tenantID uuid.UUID, filter RunFilter) ([]*PayrollRun, error)
```

#### `payroll.RuleService`

**Models:**
```
SalaryComponent { ID, TenantID, Name, Code,
                  Type (earning|deduction|tax|employer),
                  CalculationBasis (fixed|percentage|formula), Value decimal, Formula? }
TaxTable        { ID, TenantID, Country, Year, Brackets []TaxBracket }
EmployeeSalary  { ID, EmployeeID, Components []ComponentValue, EffectiveFrom, EffectiveTo? }
```

**Methods:**
```go
RuleService.CreateComponent(ctx, input ComponentInput) (*SalaryComponent, error)
RuleService.AssignSalary(ctx, employeeID uuid.UUID, input SalaryInput) (*EmployeeSalary, error)
RuleService.GetCurrentSalary(ctx, employeeID uuid.UUID) (*EmployeeSalary, error)
RuleService.UpdateTaxTable(ctx, input TaxTableInput) error
RuleService.SimulatePayslip(ctx, employeeID uuid.UUID, period DateRange) (*Payslip, error)
```

#### `payroll.ComplianceService`

**Methods:**
```go
ComplianceService.GenerateTaxReport(ctx, tenantID uuid.UUID, year int) (*TaxReport, error)
ComplianceService.ExportP9Form(ctx, employeeID uuid.UUID, year int) ([]byte, error)
ComplianceService.ExportNSSFReport(ctx, tenantID uuid.UUID, period DateRange) ([]byte, error)
ComplianceService.ExportNHIFReport(ctx, tenantID uuid.UUID, period DateRange) ([]byte, error)
```

**Events Emitted:** `payroll.RunApproved` · `payroll.RunPostedToLedger` · `payroll.PayslipFinalised`

---

### Attendance — Time & Leave

**Package:** `awo/core/attendance`  
**UI Label:** People & HR → Time & Leave  
**Purpose:** Tracks daily attendance, manages leave balances, and enforces work schedules.

#### `attendance.AttendanceService`

**Models:**
```
AttendanceRecord { ID, EmployeeID, Date, ClockIn, ClockOut?, WorkedMinutes?,
                   Source (manual|biometric|web), Note? }
WorkSchedule     { ID, TenantID, Name, WeeklyHours, Shifts []ShiftRule }
```

**Methods:**
```go
AttendanceService.ClockIn(ctx, employeeID uuid.UUID, at time.Time, source string) (*AttendanceRecord, error)
AttendanceService.ClockOut(ctx, employeeID uuid.UUID, at time.Time) (*AttendanceRecord, error)
AttendanceService.ManualEntry(ctx, input AttendanceInput) (*AttendanceRecord, error)
AttendanceService.GetDailySummary(ctx, employeeID uuid.UUID, date time.Time) (*AttendanceSummary, error)
AttendanceService.GetWorkedHours(ctx, employeeID uuid.UUID, period DateRange) (decimal.Decimal, error)
AttendanceService.AssignSchedule(ctx, employeeID, scheduleID uuid.UUID) error
```

#### `attendance.LeaveService`

**Models:**
```
LeaveType    { ID, TenantID, Name, Code, AllowedDays, AccrualBasis, CarryOver, PaidLeave }
LeaveBalance { ID, EmployeeID, LeaveTypeID, Year, Allocated, Used, Remaining }
LeaveRequest { ID, EmployeeID, LeaveTypeID, StartDate, EndDate, Days,
               Status (pending|approved|rejected|cancelled),
               RequestedAt, ApprovedBy?, ApprovedAt?, Reason }
```

**Methods:**
```go
LeaveService.RequestLeave(ctx, input LeaveRequestInput) (*LeaveRequest, error)
LeaveService.ApproveLeave(ctx, requestID, approverID uuid.UUID) error
LeaveService.RejectLeave(ctx, requestID uuid.UUID, reason string) error
LeaveService.CancelLeave(ctx, requestID uuid.UUID) error
LeaveService.GetBalance(ctx, employeeID uuid.UUID, year int) ([]*LeaveBalance, error)
LeaveService.AccrueLeave(ctx, tenantID uuid.UUID, period DateRange) error
LeaveService.ListRequests(ctx, filter LeaveFilter) ([]*LeaveRequest, Pagination, error)
```

**Events Emitted:** `attendance.LeaveApproved` · `attendance.LeaveRejected`

---

### Talent — Performance & Growth

**Package:** `awo/core/talent`  
**UI Label:** People & HR → Performance  
**Purpose:** Manages goal setting, performance reviews, and continuous feedback.

#### `talent.GoalService`

**Models:**
```
Goal { ID, EmployeeID, Title, Description, Metric, Target, Current,
       DueDate, Status (onTrack|atRisk|achieved|missed), Weight }
```

**Methods:**
```go
GoalService.CreateGoal(ctx, input GoalInput) (*Goal, error)
GoalService.UpdateProgress(ctx, goalID uuid.UUID, current decimal.Decimal, note string) error
GoalService.ListGoals(ctx, employeeID, cycleID uuid.UUID) ([]*Goal, error)
GoalService.CloseGoal(ctx, goalID uuid.UUID, status string) error
```

#### `talent.ReviewService`

**Models:**
```
ReviewCycle    { ID, TenantID, Name, StartDate, EndDate, Type (annual|quarterly|probation) }
Review         { ID, CycleID, EmployeeID, ReviewerID,
                 Status (pending|inProgress|submitted|finalised),
                 OverallRating?, FinalComment?, SubmittedAt? }
ReviewQuestion { ID, CycleID, Text, Type (rating|text|multiChoice), Options? }
ReviewAnswer   { ID, ReviewID, QuestionID, Rating?, TextAnswer? }
Feedback       { ID, FromUserID, ToEmployeeID, Body, Visibility (private|manager|public), CreatedAt }
```

**Methods:**
```go
ReviewService.CreateCycle(ctx, input CycleInput) (*ReviewCycle, error)
ReviewService.LaunchCycle(ctx, cycleID uuid.UUID) error
ReviewService.SubmitReview(ctx, reviewID uuid.UUID, answers []AnswerInput, comment string) error
ReviewService.FinaliseReview(ctx, reviewID uuid.UUID, rating decimal.Decimal, comment string) error
ReviewService.GiveFeedback(ctx, input FeedbackInput) (*Feedback, error)
ReviewService.GetEmployeeReviewHistory(ctx, employeeID uuid.UUID) ([]*Review, error)
```

---

## Business Layer — Supply Chain

---

### Catalog — Products & Pricing

**Package:** `awo/core/catalog`  
**UI Label:** Inventory & Supply → Products  
**Purpose:** Central product and pricing master. All modules reference product definitions from here.

#### `catalog.ProductService`

**Models:**
```
Product         { ID, TenantID, SKU, Name, Description,
                  Type (storable|service|consumable), CategoryID,
                  UnitOfMeasure, TrackInventory bool, Photo?, Active }
ProductCategory { ID, TenantID, Name, ParentID? }
UnitOfMeasure   { ID, TenantID, Name, Symbol, Category }
ProductVariant  { ID, ProductID, SKU, Attributes JSONB, ExtraPrice }
```

**Methods:**
```go
ProductService.CreateProduct(ctx, input ProductInput) (*Product, error)
ProductService.UpdateProduct(ctx, productID uuid.UUID, input ProductInput) error
ProductService.ArchiveProduct(ctx, productID uuid.UUID) error
ProductService.GetProduct(ctx, productID uuid.UUID) (*Product, error)
ProductService.ListProducts(ctx, filter ProductFilter) ([]*Product, Pagination, error)
ProductService.CreateVariant(ctx, productID uuid.UUID, input VariantInput) (*ProductVariant, error)
```

#### `catalog.PricingService`

**Models:**
```
PriceList     { ID, TenantID, Name, CurrencyCode, StartDate, EndDate?, Active }
PriceListItem { ID, PriceListID, ProductID, VariantID?, UnitPrice, MinQty }
PriceRule     { ID, TenantID, Name, Basis (fixed|discount|markup), Conditions JSONB }
```

**Methods:**
```go
PricingService.CreatePriceList(ctx, input PriceListInput) (*PriceList, error)
PricingService.SetPrice(ctx, priceListID, productID uuid.UUID, price decimal.Decimal) error
PricingService.GetPrice(ctx, productID uuid.UUID, qty decimal.Decimal, customerID uuid.UUID, date time.Time) (decimal.Decimal, error)
PricingService.BulkImportPrices(ctx, priceListID uuid.UUID, rows []PriceRow) error
```

---

### Inventory — Stock Management

**Package:** `awo/core/inventory`  
**UI Label:** Inventory & Supply → Stock Levels · Stock Moves  
**Purpose:** Tracks real-time stock levels, movements, and adjustments across locations.

#### `inventory.StockService`

**Models:**
```
StockLocation       { ID, TenantID, Name, Code,
                      Type (warehouse|transit|virtual|customer|vendor),
                      ParentID?, Address? }
StockQuant          { ID, TenantID, ProductID, LocationID, Quantity, Reserved, Available }
StockMove           { ID, TenantID, ProductID, FromLocationID, ToLocationID, Quantity,
                      Status (draft|confirmed|done|cancelled), Reference,
                      SourceModule, SourceID, ScheduledAt, DoneAt? }
InventoryAdjustment { ID, TenantID, LocationID, Status, Lines []AdjLine,
                      ValidatedBy?, ValidatedAt? }
ReorderRule         { ID, TenantID, ProductID, LocationID, MinQty, MaxQty, RouteID }
```

**Methods:**
```go
StockService.Receive(ctx, input ReceiveInput) (*StockMove, error)
StockService.Dispatch(ctx, input DispatchInput) (*StockMove, error)
StockService.Transfer(ctx, input TransferInput) (*StockMove, error)
StockService.CreateAdjustment(ctx, locationID uuid.UUID) (*InventoryAdjustment, error)
StockService.AddAdjustmentLine(ctx, adjustmentID, productID uuid.UUID, countedQty decimal.Decimal) error
StockService.ValidateAdjustment(ctx, adjustmentID uuid.UUID) error
StockService.GetStockLevel(ctx, productID, locationID uuid.UUID) (*StockQuant, error)
StockService.GetStockByProduct(ctx, productID uuid.UUID) ([]*StockQuant, error)
StockService.ListMoves(ctx, filter MoveFilter) ([]*StockMove, Pagination, error)
StockService.SetReorderRule(ctx, input ReorderRuleInput) (*ReorderRule, error)
StockService.GetBelowReorder(ctx, tenantID uuid.UUID) ([]*ReorderRule, error)
```

**Events Emitted:** `inventory.StockReceived` · `inventory.StockDispatched` · `inventory.StockAdjusted` · `inventory.ReorderPointReached`

---

### Procurement — Purchasing

**Package:** `awo/core/procurement`  
**UI Label:** Inventory & Supply → Purchase Orders · Suppliers  
**Purpose:** Manages the full purchasing cycle from RFQ through goods receipt and bill matching.

#### `procurement.SupplierService`

**Models:**
```
Supplier         { ID, TenantID, Name, Code, TaxID, CurrencyCode,
                   PaymentTerms, PriceListID?, Active }
SupplierContact  { ID, SupplierID, Name, Email, Phone, Role }
SupplierContract { ID, SupplierID, StartDate, EndDate, Terms, FilePath? }
```

**Methods:**
```go
SupplierService.CreateSupplier(ctx, input SupplierInput) (*Supplier, error)
SupplierService.UpdateSupplier(ctx, supplierID uuid.UUID, input SupplierInput) error
SupplierService.ArchiveSupplier(ctx, supplierID uuid.UUID) error
SupplierService.GetSupplier(ctx, supplierID uuid.UUID) (*Supplier, error)
SupplierService.ListSuppliers(ctx, filter SupplierFilter) ([]*Supplier, Pagination, error)
SupplierService.EvaluateSupplier(ctx, supplierID uuid.UUID) (*SupplierScore, error)
```

#### `procurement.PurchaseOrderService`

**Models:**
```
RFQ           { ID, TenantID, SupplierID, Status (draft|sent|received), Lines []RFQLine, SentAt? }
PurchaseOrder { ID, TenantID, SupplierID, Number, OrderDate, ExpectedDate,
                Status (draft|confirmed|partiallyReceived|received|billed|cancelled),
                Total, CurrencyCode, RFQID? }
POLine        { ID, POID, ProductID, Description, Quantity, ReceivedQty, BilledQty,
                UnitPrice, TaxID?, Amount, AccountID }
GoodsReceipt  { ID, POID, Date, Lines []GRLine, CreatedBy }
```

**Methods:**
```go
PurchaseOrderService.CreateRFQ(ctx, input RFQInput) (*RFQ, error)
PurchaseOrderService.ConvertRFQToPO(ctx, rfqID uuid.UUID) (*PurchaseOrder, error)
PurchaseOrderService.CreatePO(ctx, input POInput) (*PurchaseOrder, error)
PurchaseOrderService.ConfirmPO(ctx, poID uuid.UUID) error
PurchaseOrderService.ReceiveGoods(ctx, poID uuid.UUID, lines []GRLineInput) (*GoodsReceipt, error)
PurchaseOrderService.GetPO(ctx, poID uuid.UUID) (*PurchaseOrder, error)
PurchaseOrderService.ListPOs(ctx, filter POFilter) ([]*PurchaseOrder, Pagination, error)
PurchaseOrderService.CancelPO(ctx, poID uuid.UUID, reason string) error
```

**Events Emitted:** `procurement.POConfirmed` · `procurement.GoodsReceived` · `procurement.POFullyReceived`

---

### Fulfillment — Warehousing & Shipping

**Package:** `awo/core/fulfillment`  
**UI Label:** Inventory & Supply → Warehouse · Shipping & Delivery  
**Purpose:** Manages pick-pack-ship operations and inbound receiving workflows.

#### `fulfillment.OperationsService`

**Models:**
```
Warehouse    { ID, TenantID, Name, Code, Address, BranchID? }
Bin          { ID, WarehouseID, Row, Column, Level, Barcode, ProductID?, Capacity? }
Picking      { ID, TenantID, Type (outbound|inbound|internal), SourceID,
               Status, AssignedTo?, Lines []PickLine, ScheduledAt }
Packing      { ID, PickingID, PackagedAt, PackedBy, Boxes []PackBox }
Shipment     { ID, TenantID, CarrierID, TrackingNumber?, DispatchedAt?,
               Lines []ShipLine, Status (pending|inTransit|delivered|failed) }
DeliveryNote { ID, ShipmentID, CustomerID, Lines []DNLine, SignedBy?, DeliveredAt? }
```

**Methods:**
```go
OperationsService.CreatePicking(ctx, input PickingInput) (*Picking, error)
OperationsService.AssignPicking(ctx, pickingID, workerID uuid.UUID) error
OperationsService.ValidatePicking(ctx, pickingID uuid.UUID, lines []PickDone) error
OperationsService.PackOrder(ctx, pickingID uuid.UUID, boxes []PackBoxInput) (*Packing, error)
OperationsService.CreateShipment(ctx, packingID, carrierID uuid.UUID) (*Shipment, error)
OperationsService.MarkDelivered(ctx, shipmentID uuid.UUID, proof DeliveryProof) error
OperationsService.TrackShipment(ctx, shipmentID uuid.UUID) (*ShipmentStatus, error)
```

#### `fulfillment.CarrierService`

**Models:**
```
Carrier     { ID, TenantID, Name, Code, TrackingURL?, APIConfig JSONB }
FreightRate { ID, CarrierID, ZoneFrom, ZoneTo, WeightMin, WeightMax, Rate }
```

**Methods:**
```go
CarrierService.CreateCarrier(ctx, input CarrierInput) (*Carrier, error)
CarrierService.CalculateFreight(ctx, carrierID uuid.UUID, params FreightParams) (decimal.Decimal, error)
CarrierService.ListCarriers(ctx, tenantID uuid.UUID) ([]*Carrier, error)
```

**Events Emitted:** `fulfillment.OrderDispatched` · `fulfillment.OrderDelivered`

---

## Business Layer — Revenue

---

### Contacts — Customer & Vendor Data

**Package:** `awo/core/contacts`  
**UI Label:** Sales & Customers → Contacts · Companies  
**Purpose:** Unified contact master for all customers, prospects, and companies. Referenced by pipeline, helpdesk, receivables, and payables.

#### `contacts.ContactService`

**Models:**
```
Contact        { ID, TenantID, Type (individual|company), FirstName, LastName?,
                 CompanyID?, Email, Phone?, Mobile?, Position?, Source, Tags []string }
Company        { ID, TenantID, Name, Industry, Website?, TaxID?, Address,
                 CurrencyCode, ParentID?, Segment }
InteractionLog { ID, ContactID, UserID, Type (call|email|meeting|note), Summary, OccurredAt }
ContactScore   { ContactID, Score, ScoredAt, Basis JSONB }
```

**Methods:**
```go
ContactService.CreateContact(ctx, input ContactInput) (*Contact, error)
ContactService.CreateCompany(ctx, input CompanyInput) (*Company, error)
ContactService.UpdateContact(ctx, contactID uuid.UUID, input ContactInput) error
ContactService.MergeDuplicates(ctx, primaryID uuid.UUID, duplicateIDs []uuid.UUID) error
ContactService.LogInteraction(ctx, contactID uuid.UUID, input InteractionInput) (*InteractionLog, error)
ContactService.GetContact360(ctx, contactID uuid.UUID) (*Contact360View, error)
ContactService.ListContacts(ctx, filter ContactFilter) ([]*Contact, Pagination, error)
ContactService.SearchContacts(ctx, tenantID uuid.UUID, query string) ([]*Contact, error)
ContactService.SegmentContacts(ctx, tenantID uuid.UUID, criteria SegmentCriteria) ([]*Contact, error)
```

---

### Pipeline — Sales & Opportunities

**Package:** `awo/core/pipeline`  
**UI Label:** Sales & Customers → Leads · Opportunities · Quotes & Orders  
**Purpose:** Manages the full sales cycle from lead capture through to confirmed sales order.

#### `pipeline.LeadService`

**Models:**
```
Lead { ID, TenantID, ContactID?, CompanyID?, Title, Source, AssignedTo?,
       Status (new|contacted|qualified|disqualified), Score, CreatedAt }
```

**Methods:**
```go
LeadService.CaptureLead(ctx, input LeadInput) (*Lead, error)
LeadService.QualifyLead(ctx, leadID uuid.UUID, notes string) error
LeadService.DisqualifyLead(ctx, leadID uuid.UUID, reason string) error
LeadService.AssignLead(ctx, leadID, userID uuid.UUID) error
LeadService.ConvertToOpportunity(ctx, leadID uuid.UUID) (*Opportunity, error)
LeadService.ListLeads(ctx, filter LeadFilter) ([]*Lead, Pagination, error)
```

#### `pipeline.OpportunityService`

**Models:**
```
Opportunity { ID, TenantID, LeadID?, ContactID, CompanyID?, Title, Stage,
              Probability, ExpectedRevenue, CloseDate, AssignedTo,
              Status (open|won|lost), LostReason? }
Stage       { ID, TenantID, Name, Sequence, Probability, IsFinal }
Activity    { ID, OpportunityID, Type (call|email|meeting|task), DueAt,
              DoneAt?, AssignedTo, Notes }
```

**Methods:**
```go
OpportunityService.CreateOpportunity(ctx, input OpportunityInput) (*Opportunity, error)
OpportunityService.MoveStage(ctx, opportunityID, stageID uuid.UUID) error
OpportunityService.MarkWon(ctx, opportunityID uuid.UUID) error
OpportunityService.MarkLost(ctx, opportunityID uuid.UUID, reason string) error
OpportunityService.LogActivity(ctx, opportunityID uuid.UUID, input ActivityInput) (*Activity, error)
OpportunityService.GetForecast(ctx, tenantID uuid.UUID, period DateRange) (*SalesForecast, error)
OpportunityService.ListOpportunities(ctx, filter OppFilter) ([]*Opportunity, Pagination, error)
```

#### `pipeline.SalesOrderService`

**Models:**
```
Quote     { ID, TenantID, OpportunityID?, ContactID, CompanyID, Date, ValidUntil,
            Status (draft|sent|accepted|declined|expired), Lines []QuoteLine, Total }
QuoteLine { ID, QuoteID, ProductID, Description, Qty, UnitPrice, DiscountPct, TaxID?, Amount }
SalesOrder { ID, TenantID, QuoteID?, ContactID, CompanyID, OrderDate, ConfirmedAt?,
             Status (draft|confirmed|partiallyDelivered|delivered|invoiced|cancelled),
             Lines []SOLine, Total }
SOLine    { ID, OrderID, ProductID, Description, QtyOrdered, QtyDelivered, QtyInvoiced,
            UnitPrice, DiscountPct, TaxID?, Amount }
```

**Methods:**
```go
SalesOrderService.CreateQuote(ctx, input QuoteInput) (*Quote, error)
SalesOrderService.SendQuote(ctx, quoteID uuid.UUID, recipients []string) error
SalesOrderService.ConvertQuoteToOrder(ctx, quoteID uuid.UUID) (*SalesOrder, error)
SalesOrderService.ConfirmOrder(ctx, orderID uuid.UUID) error
SalesOrderService.CancelOrder(ctx, orderID uuid.UUID, reason string) error
SalesOrderService.GetOrder(ctx, orderID uuid.UUID) (*SalesOrder, error)
SalesOrderService.ListOrders(ctx, filter SOFilter) ([]*SalesOrder, Pagination, error)
```

**Events Emitted:** `pipeline.LeadCaptured` · `pipeline.OpportunityWon` · `pipeline.SalesOrderConfirmed` · `pipeline.QuoteSent`

---

### Campaigns — Marketing Automation

**Package:** `awo/core/campaigns`  
**UI Label:** Sales & Customers → Campaigns  
**Purpose:** Plans and executes marketing campaigns, email sequences, and lead nurturing flows.

#### `campaigns.CampaignService`

**Models:**
```
Campaign         { ID, TenantID, Name, Type (email|sms|event),
                   Status (draft|scheduled|running|paused|completed),
                   AudienceSegmentID, Budget decimal?, StartDate, EndDate? }
CampaignActivity { ID, CampaignID, Type, ScheduledAt, CompletedAt?, Metrics JSONB }
EmailTemplate    { ID, TenantID, Name, Subject, BodyHTML, BodyText, Variables []string }
MailingList      { ID, TenantID, Name, SubscriberCount }
MailingListMember { ListID, ContactID, SubscribedAt, UnsubscribedAt? }
LeadScoreRule    { ID, TenantID, EventType, Points, Decay? }
```

**Methods:**
```go
CampaignService.CreateCampaign(ctx, input CampaignInput) (*Campaign, error)
CampaignService.LaunchCampaign(ctx, campaignID uuid.UUID) error
CampaignService.PauseCampaign(ctx, campaignID uuid.UUID) error
CampaignService.ScheduleEmail(ctx, campaignID, templateID uuid.UUID, sendAt time.Time) error
CampaignService.GetPerformance(ctx, campaignID uuid.UUID) (*CampaignMetrics, error)
CampaignService.CreateMailingList(ctx, tenantID uuid.UUID, name string) (*MailingList, error)
CampaignService.AddToList(ctx, listID uuid.UUID, contactIDs []uuid.UUID) error
CampaignService.ScoreLeads(ctx, tenantID uuid.UUID) error
```

---

### Helpdesk — Customer Support

**Package:** `awo/core/helpdesk`  
**UI Label:** Sales & Customers → Support  
**Purpose:** Manages customer support tickets, SLA enforcement, and knowledge base.

#### `helpdesk.TicketService`

**Models:**
```
Ticket        { ID, TenantID, ContactID, CompanyID?, Subject, Description,
                Channel (email|web|phone), Priority (low|medium|high|critical),
                Status (open|inProgress|pending|resolved|closed),
                AssignedTo?, TeamID?, SLAID?,
                OpenedAt, FirstResponseAt?, ResolvedAt? }
TicketMessage { ID, TicketID, AuthorID, Body, IsInternal, CreatedAt, Attachments []string }
SLAPolicy     { ID, TenantID, Name, ResponseTimeHours, ResolutionTimeHours, Conditions JSONB }
Team          { ID, TenantID, Name, Members []uuid.UUID }
```

**Methods:**
```go
TicketService.CreateTicket(ctx, input TicketInput) (*Ticket, error)
TicketService.AssignTicket(ctx, ticketID, agentID uuid.UUID) error
TicketService.ReplyToTicket(ctx, ticketID uuid.UUID, message TicketMessageInput) error
TicketService.EscalateTicket(ctx, ticketID uuid.UUID, reason string) error
TicketService.ResolveTicket(ctx, ticketID uuid.UUID, resolution string) error
TicketService.CloseTicket(ctx, ticketID uuid.UUID) error
TicketService.GetTicket(ctx, ticketID uuid.UUID) (*Ticket, error)
TicketService.ListTickets(ctx, filter TicketFilter) ([]*Ticket, Pagination, error)
TicketService.CheckSLABreach(ctx, tenantID uuid.UUID) ([]*Ticket, error)
```

#### `helpdesk.KnowledgeService`

**Models:**
```
Article         { ID, TenantID, Title, Body, CategoryID,
                  Status (draft|published), AuthorID,
                  PublishedAt?, Tags []string, Views int }
ArticleCategory { ID, TenantID, Name, ParentID? }
```

**Methods:**
```go
KnowledgeService.CreateArticle(ctx, input ArticleInput) (*Article, error)
KnowledgeService.PublishArticle(ctx, articleID uuid.UUID) error
KnowledgeService.SearchArticles(ctx, tenantID uuid.UUID, query string) ([]*Article, error)
KnowledgeService.SuggestForTicket(ctx, ticketID uuid.UUID) ([]*Article, error)
```

**Events Emitted:** `helpdesk.TicketCreated` · `helpdesk.TicketResolved` · `helpdesk.SLABreached`

---

## Business Layer — Production

---

### Planning — Production Planning

**Package:** `awo/core/planning`  
**UI Label:** Production → Production Orders · Work Centres  
**Purpose:** Plans and executes manufacturing production orders.

#### `planning.ProductionService`

**Models:**
```
ProductionOrder  { ID, TenantID, ProductID, VariantID?, QtyPlanned, QtyProduced,
                   BOMID, RoutingID?, Status (draft|confirmed|inProgress|done|cancelled),
                   ScheduledStart, ScheduledEnd, ActualStart?, ActualEnd?,
                   WorkCentreID?, Priority }
WorkCentre       { ID, TenantID, Name, Code, Type (machine|labour|external),
                   Capacity, CostPerHour, Active }
Routing          { ID, TenantID, ProductID, Operations []RoutingOperation }
RoutingOperation { ID, RoutingID, Name, WorkCentreID, DurationMins, Sequence }
CapacityPlan     { ID, TenantID, WorkCentreID, Date, AvailableMins, UsedMins }
```

**Methods:**
```go
ProductionService.CreateOrder(ctx, input ProductionOrderInput) (*ProductionOrder, error)
ProductionService.ConfirmOrder(ctx, orderID uuid.UUID) error
ProductionService.StartOrder(ctx, orderID uuid.UUID) error
ProductionService.ReportProgress(ctx, orderID uuid.UUID, qty decimal.Decimal, workerID uuid.UUID) error
ProductionService.CompleteOrder(ctx, orderID uuid.UUID) error
ProductionService.ScheduleOrders(ctx, tenantID uuid.UUID, horizon DateRange) error
ProductionService.GetCapacityLoad(ctx, workCentreID uuid.UUID, period DateRange) (*CapacityPlan, error)
ProductionService.ListOrders(ctx, filter ProdFilter) ([]*ProductionOrder, Pagination, error)
```

**Events Emitted:** `planning.ProductionOrderCompleted` · `planning.ProductionOrderStarted`

---

### BOM — Bill of Materials

**Package:** `awo/core/bom`  
**UI Label:** Production → Bill of Materials  
**Purpose:** Manages product component structures, versions, and engineering changes.

#### `bom.BOMService`

**Models:**
```
BOM         { ID, TenantID, ProductID, VariantID?, Version,
              Status (active|archived), Type (manufacture|phantom|subcontract),
              Qty decimal, Notes? }
BOMLine     { ID, BOMID, ComponentID, Qty, UOM, WastePercent?, OperationID? }
BOMVersion  { BOMID, Version, ChangedBy, ChangedAt, ChangeNote }
ChangeOrder { ID, TenantID, BOMID, Title, Description, Status, EffectiveDate }
```

**Methods:**
```go
BOMService.CreateBOM(ctx, input BOMInput) (*BOM, error)
BOMService.AddComponent(ctx, bomID, productID uuid.UUID, qty decimal.Decimal, uom string) (*BOMLine, error)
BOMService.RemoveComponent(ctx, lineID uuid.UUID) error
BOMService.PublishVersion(ctx, bomID uuid.UUID, note string) (*BOMVersion, error)
BOMService.ExplodeBOM(ctx, bomID uuid.UUID, qty decimal.Decimal) ([]*ExplodedComponent, error)
BOMService.CostRollUp(ctx, bomID uuid.UUID) (*BOMCost, error)
BOMService.CreateChangeOrder(ctx, input ChangeOrderInput) (*ChangeOrder, error)
BOMService.ApplyChangeOrder(ctx, changeOrderID uuid.UUID) error
```

---

### Quality — Quality Management

**Package:** `awo/core/quality`  
**UI Label:** Production → Quality Control  
**Purpose:** Manages inspection plans, test results, and non-conformance handling.

#### `quality.InspectionService`

**Models:**
```
InspectionPlan { ID, TenantID, Name, ProductID?, OperationID?,
                 Criteria []QCCriterion,
                 TriggerType (receipt|production|periodic) }
QCCriterion    { ID, PlanID, Name, Type (measure|passFail|count),
                 MinValue?, MaxValue?, Unit? }
Inspection     { ID, TenantID, PlanID, SourceType (production|receipt), SourceID,
                 Status (pending|inProgress|passed|failed),
                 InspectedBy?, CompletedAt? }
TestResult     { ID, InspectionID, CriterionID, Value, Passed bool, Notes? }
```

**Methods:**
```go
InspectionService.CreatePlan(ctx, input InspectionPlanInput) (*InspectionPlan, error)
InspectionService.TriggerInspection(ctx, planID uuid.UUID, sourceType string, sourceID uuid.UUID) (*Inspection, error)
InspectionService.RecordResults(ctx, inspectionID uuid.UUID, results []TestResultInput) error
InspectionService.CloseInspection(ctx, inspectionID uuid.UUID) (*Inspection, error)
InspectionService.ListInspections(ctx, filter InspFilter) ([]*Inspection, Pagination, error)
```

#### `quality.NonConformanceService`

**Models:**
```
NonConformance { ID, TenantID, Title, SourceType, SourceID, Description, Severity,
                 Status (open|underReview|resolved|closed), AssignedTo?, ClosedAt? }
NCAction       { ID, NCID, ActionType (rework|scrap|use|return), Description, DoneBy?, DoneAt? }
QualityAlert   { ID, TenantID, NCID?, Title, Message, Recipients []uuid.UUID, SentAt }
```

**Methods:**
```go
NonConformanceService.RaiseNC(ctx, input NCInput) (*NonConformance, error)
NonConformanceService.TakeAction(ctx, ncID uuid.UUID, input NCActionInput) error
NonConformanceService.CloseNC(ctx, ncID uuid.UUID, resolution string) error
NonConformanceService.IssueAlert(ctx, ncID uuid.UUID, title, message string) (*QualityAlert, error)
NonConformanceService.QualityReport(ctx, tenantID uuid.UUID, period DateRange) (*QualityReport, error)
```

---

### Assets — Asset & Maintenance

**Package:** `awo/core/assets`  
**UI Label:** Production → Assets & Maintenance  
**Purpose:** Tracks physical assets, schedules preventive maintenance, and manages work orders.

#### `assets.AssetService`

**Models:**
```
Asset           { ID, TenantID, Name, Code, Category, SerialNumber?,
                  Manufacturer?, Model?, PurchaseDate?, PurchaseCost?,
                  LocationID, AssignedTo?,
                  Status (active|underMaintenance|retired) }
AssetCategory   { ID, TenantID, Name, DepreciationMethod, UsefulLifeYears }
DepreciationEntry { ID, AssetID, PeriodID, Amount, BookValue, JournalEntryID? }
```

**Methods:**
```go
AssetService.CreateAsset(ctx, input AssetInput) (*Asset, error)
AssetService.TransferAsset(ctx, assetID, locationID uuid.UUID, date time.Time) error
AssetService.RetireAsset(ctx, assetID uuid.UUID, date time.Time, reason string) error
AssetService.CalculateDepreciation(ctx, assetID uuid.UUID, period DateRange) (*DepreciationEntry, error)
AssetService.ListAssets(ctx, filter AssetFilter) ([]*Asset, Pagination, error)
```

#### `assets.MaintenanceService`

**Models:**
```
MaintenanceSchedule { ID, AssetID, Type (preventive|periodic), Frequency,
                      FrequencyUnit, LastRunAt?, NextRunAt, Checklist []string }
WorkOrder           { ID, TenantID, AssetID, Title, Description, Priority, Type,
                      Status (new|assigned|inProgress|done|cancelled),
                      AssignedTo?, ScheduledAt, StartedAt?, CompletedAt? }
MaintenanceLog      { ID, WorkOrderID, Notes, PartsUsed []PartUsed, Labour decimal,
                      CompletedBy, DoneAt }
SparePart           { ID, TenantID, Name, SKU, Quantity, MinStock, SupplierID? }
```

**Methods:**
```go
MaintenanceService.CreateSchedule(ctx, input ScheduleInput) (*MaintenanceSchedule, error)
MaintenanceService.GenerateDueWorkOrders(ctx, tenantID uuid.UUID) ([]*WorkOrder, error)
MaintenanceService.CreateWorkOrder(ctx, input WorkOrderInput) (*WorkOrder, error)
MaintenanceService.AssignWorkOrder(ctx, workOrderID, technicianID uuid.UUID) error
MaintenanceService.CompleteWorkOrder(ctx, workOrderID uuid.UUID, log MaintenanceLogInput) error
MaintenanceService.TrackDowntime(ctx, assetID uuid.UUID, period DateRange) (*DowntimeReport, error)
```

---

## Business Layer — Projects

---

### Projects — Project Management

**Package:** `awo/core/projects`  
**UI Label:** Projects → All Projects · Tasks · Milestones  
**Purpose:** Core project structure, task management, and milestone tracking.

#### `projects.ProjectService`

**Models:**
```
Project         { ID, TenantID, Name, Code, CustomerID?, ManagerID,
                  Status (draft|active|onHold|closed), StartDate, EndDate,
                  Description, Tags []string, TemplateID? }
Task            { ID, ProjectID, Title, Description, AssigneeID?,
                  Status (todo|inProgress|review|done|cancelled),
                  Priority, ParentTaskID?, DueDate?,
                  EstimatedHours?, ActualHours? }
Milestone       { ID, ProjectID, Name, DueDate, Achieved bool, AchievedAt? }
ProjectTemplate { ID, TenantID, Name, Tasks []TemplateTask, Milestones []TemplateMilestone }
```

**Methods:**
```go
ProjectService.CreateProject(ctx, input ProjectInput) (*Project, error)
ProjectService.UpdateProject(ctx, projectID uuid.UUID, input ProjectInput) error
ProjectService.CloseProject(ctx, projectID uuid.UUID) error
ProjectService.AddTask(ctx, projectID uuid.UUID, input TaskInput) (*Task, error)
ProjectService.UpdateTask(ctx, taskID uuid.UUID, input TaskInput) error
ProjectService.MoveTask(ctx, taskID uuid.UUID, newStatus string) error
ProjectService.AddMilestone(ctx, projectID uuid.UUID, input MilestoneInput) (*Milestone, error)
ProjectService.MarkMilestoneAchieved(ctx, milestoneID uuid.UUID) error
ProjectService.ImportTemplate(ctx, projectID, templateID uuid.UUID) error
ProjectService.GetGantt(ctx, projectID uuid.UUID) (*GanttData, error)
ProjectService.ListProjects(ctx, filter ProjectFilter) ([]*Project, Pagination, error)
ProjectService.GetMyTasks(ctx, userID uuid.UUID) ([]*Task, error)
```

---

### Resources — Resource Planning

**Package:** `awo/core/resources`  
**UI Label:** Projects → (within project detail view)  
**Purpose:** Allocates employees to projects and checks availability.

#### `resources.AllocationService`

**Models:**
```
ResourceAllocation { ID, ProjectID, EmployeeID, RoleID, AllocationPct, StartDate, EndDate }
ResourceRequest    { ID, ProjectID, RoleID, RequiredSkills []string, StartDate, EndDate, Status }
Skill              { ID, TenantID, Name, Category }
EmployeeSkill      { EmployeeID, SkillID, ProficiencyLevel (beginner|intermediate|expert), CertifiedAt? }
```

**Methods:**
```go
AllocationService.AllocateEmployee(ctx, input AllocationInput) (*ResourceAllocation, error)
AllocationService.DeallocateEmployee(ctx, allocationID uuid.UUID) error
AllocationService.CheckAvailability(ctx, employeeID uuid.UUID, period DateRange) (*AvailabilityView, error)
AllocationService.RequestResource(ctx, input ResourceRequestInput) (*ResourceRequest, error)
AllocationService.FindAvailableBySkill(ctx, tenantID uuid.UUID, skillIDs []uuid.UUID, period DateRange) ([]*Employee, error)
AllocationService.GetProjectTeam(ctx, projectID uuid.UUID) ([]*ResourceAllocation, error)
```

---

### Timesheets — Time & Expense Tracking

**Package:** `awo/core/timesheets`  
**UI Label:** Projects → Timesheets · Expenses  
**Purpose:** Captures time worked on project tasks and employee out-of-pocket expenses.

#### `timesheets.TimesheetService`

**Models:**
```
Timesheet     { ID, EmployeeID, WeekStarting,
                Status (draft|submitted|approved|rejected),
                SubmittedAt?, ApprovedBy?, ApprovedAt? }
TimesheetLine { ID, TimesheetID, ProjectID, TaskID?, Date, Hours, Description }
```

**Methods:**
```go
TimesheetService.GetOrCreate(ctx, employeeID uuid.UUID, weekStarting time.Time) (*Timesheet, error)
TimesheetService.LogTime(ctx, timesheetID uuid.UUID, line TimesheetLineInput) (*TimesheetLine, error)
TimesheetService.UpdateLine(ctx, lineID uuid.UUID, hours decimal.Decimal, description string) error
TimesheetService.DeleteLine(ctx, lineID uuid.UUID) error
TimesheetService.SubmitTimesheet(ctx, timesheetID uuid.UUID) error
TimesheetService.ApproveTimesheet(ctx, timesheetID, approverID uuid.UUID) error
TimesheetService.RejectTimesheet(ctx, timesheetID uuid.UUID, reason string) error
TimesheetService.GetProjectHours(ctx, projectID uuid.UUID, period DateRange) (decimal.Decimal, error)
```

#### `timesheets.ExpenseService`

**Models:**
```
ExpenseReport { ID, EmployeeID, Title,
                Status (draft|submitted|approved|rejected|paid),
                SubmittedAt?, ApprovedBy?, Total }
ExpenseLine   { ID, ReportID, Date, Category, Description, Amount, CurrencyCode,
                ProjectID?, TaskID?, ReceiptPath? }
MileageLog    { ID, EmployeeID, Date, Distance, Unit, RatePerUnit, ProjectID?, Amount }
```

**Methods:**
```go
ExpenseService.CreateReport(ctx, input ExpenseReportInput) (*ExpenseReport, error)
ExpenseService.AddLine(ctx, reportID uuid.UUID, line ExpenseLineInput) (*ExpenseLine, error)
ExpenseService.SubmitReport(ctx, reportID uuid.UUID) error
ExpenseService.ApproveReport(ctx, reportID, approverID uuid.UUID) error
ExpenseService.RejectReport(ctx, reportID uuid.UUID, reason string) error
ExpenseService.LogMileage(ctx, input MileageInput) (*MileageLog, error)
ExpenseService.GetProjectExpenses(ctx, projectID uuid.UUID) ([]*ExpenseLine, error)
```

---

### ProjectAccounting — Project Finance

**Package:** `awo/core/projectacct`  
**UI Label:** Projects → Project Finance  
**Purpose:** Tracks project costs, manages project budgets, and generates project invoices.

#### `projectacct.ProjectFinanceService`

**Models:**
```
ProjectBudget  { ID, ProjectID, TotalBudget, LabourBudget, MaterialBudget,
                 ExpenseBudget, ApprovedBy?, ApprovedAt? }
ProjectCost    { ID, ProjectID, Type (labour|material|expense|overhead),
                 Description, Amount, Date,
                 SourceType, SourceID, JournalEntryID? }
ProjectInvoice { ID, ProjectID, CustomerID, BillingPeriod,
                 Lines []ProjInvLine, Status, Total, JournalEntryID? }
BillingMethod  { ProjectID, Type (fixedFee|timeAndMaterial|milestone), Config JSONB }
```

**Methods:**
```go
ProjectFinanceService.SetBudget(ctx, projectID uuid.UUID, input BudgetInput) (*ProjectBudget, error)
ProjectFinanceService.RecordCost(ctx, input ProjectCostInput) (*ProjectCost, error)
ProjectFinanceService.GetCostSummary(ctx, projectID uuid.UUID) (*CostSummary, error)
ProjectFinanceService.BudgetVsActual(ctx, projectID uuid.UUID) (*BudgetVarianceReport, error)
ProjectFinanceService.GenerateInvoice(ctx, projectID uuid.UUID, period DateRange) (*ProjectInvoice, error)
ProjectFinanceService.PostInvoice(ctx, projectInvoiceID uuid.UUID) error
ProjectFinanceService.Profitability(ctx, projectID uuid.UUID) (*ProfitabilityReport, error)
```

---

## Analytics Layer

---

### Reports — Standard Reporting

**Package:** `awo/analytics/reports`  
**UI Label:** Analytics → Reports  
**Purpose:** Generates pre-built and custom reports using read-optimised queries.

#### `reports.ReportService`

**Models:**
```
ReportDefinition { ID, TenantID, Name, Module, Query, Parameters []ParamDef,
                   DefaultFormat, IsSystem bool }
ReportExecution  { ID, DefinitionID, TenantID, TriggeredBy, Parameters JSONB,
                   Status (queued|running|done|failed), OutputPath?, CompletedAt? }
ScheduledReport  { ID, DefinitionID, CronExpr, Recipients []string, Format, Active }
```

**Methods:**
```go
ReportService.RunReport(ctx, definitionID uuid.UUID, params ReportParams) (*ReportExecution, error)
ReportService.GetExecution(ctx, executionID uuid.UUID) (*ReportExecution, error)
ReportService.ExportReport(ctx, executionID uuid.UUID, format string) ([]byte, error)
ReportService.ScheduleReport(ctx, input ScheduledReportInput) (*ScheduledReport, error)
ReportService.CreateCustomReport(ctx, input CustomReportInput) (*ReportDefinition, error)
ReportService.ListDefinitions(ctx, tenantID uuid.UUID, module string) ([]*ReportDefinition, error)
```

---

### Dashboards — KPI Dashboards

**Package:** `awo/analytics/dashboards`  
**UI Label:** Analytics → My Dashboard · All Dashboards  
**Purpose:** Configurable KPI dashboards with real-time widgets.

#### `dashboards.DashboardService`

**Models:**
```
Dashboard      { ID, TenantID, Name, OwnerID, Layout JSONB, Shared bool }
Widget         { ID, DashboardID, Type (metric|chart|table|list),
                 KPIKey, Title, Config JSONB, Position JSONB }
KPIDefinition  { Key, Name, Module, Query, Format, Unit? }
DashboardShare { DashboardID, UserID, CanEdit bool }
```

**Methods:**
```go
DashboardService.CreateDashboard(ctx, input DashboardInput) (*Dashboard, error)
DashboardService.AddWidget(ctx, dashboardID uuid.UUID, input WidgetInput) (*Widget, error)
DashboardService.RemoveWidget(ctx, widgetID uuid.UUID) error
DashboardService.UpdateLayout(ctx, dashboardID uuid.UUID, layout map[string]any) error
DashboardService.ShareDashboard(ctx, dashboardID, userID uuid.UUID, canEdit bool) error
DashboardService.RefreshWidget(ctx, widgetID uuid.UUID) (*WidgetData, error)
DashboardService.ListDashboards(ctx, userID uuid.UUID) ([]*Dashboard, error)
```

---

### Insights — Advanced Analytics

**Package:** `awo/analytics/insights`  
**UI Label:** Analytics → Insights  
**Purpose:** Trend detection, anomaly spotting, and predictive forecasts.

#### `insights.AnalyticsService`

**Models:**
```
DataSet   { ID, TenantID, Name, Module, Query, CachedAt?, CacheData JSONB? }
Trend     { ID, DataSetID, Metric, Direction (up|down|stable), Magnitude, DetectedAt }
Forecast  { ID, DataSetID, Metric, Method, Horizon,
            Predictions []ForecastPoint, GeneratedAt }
```

**Methods:**
```go
AnalyticsService.QueryDataSet(ctx, dataSetID uuid.UUID, params QueryParams) ([]map[string]any, error)
AnalyticsService.DetectTrends(ctx, dataSetID uuid.UUID) ([]*Trend, error)
AnalyticsService.GenerateForecast(ctx, dataSetID uuid.UUID, horizon int) (*Forecast, error)
AnalyticsService.ExportDataSet(ctx, dataSetID uuid.UUID, format string) ([]byte, error)
```

---

## Dependency Cycle Prevention

### The Core Rule

> **No business module imports another business module's Go package.**

This is enforced structurally and with tooling. Every cross-module interaction is mediated by one of three mechanisms below.

---

### Mechanism 1 — Ports & Adapters (synchronous reads)

When Module A needs a value owned by Module B, Module A defines a **port interface** in its own package. Module B provides an **adapter** struct that satisfies it. Wire injects the adapter at startup. Module A's source files never contain `import "awo/core/moduleB"`.

```go
// awo/core/payroll/port.go
// Defined BY payroll. Implemented BY workforce (as an adapter). Injected by Wire.

type EmployeeReader interface {
    GetPayrollEmployees(ctx context.Context, tenantID uuid.UUID) ([]*EmployeePayrollView, error)
    GetSalaryComponents(ctx context.Context, employeeID uuid.UUID) (*SalaryData, error)
}

// Defined BY payroll. Implemented BY attendance.
type WorkedHoursReader interface {
    GetWorkedHours(ctx context.Context, employeeID uuid.UUID, period DateRange) (decimal.Decimal, error)
}

// Defined BY payroll. Implemented BY ledger.
type PayrollPoster interface {
    PostPayrollRun(ctx context.Context, req PostingRequest) error
}
```

Each provider module exposes an adapter file:

```go
// awo/core/workforce/payroll_adapter.go
// This file lives in the workforce package but satisfies payroll's interface.
// The workforce package does NOT import payroll.

type PayrollAdapter struct{ svc *EmployeeService }

func NewPayrollAdapter(svc *EmployeeService) *PayrollAdapter { ... }

func (a *PayrollAdapter) GetPayrollEmployees(ctx context.Context, tenantID uuid.UUID) ([]*payroll.EmployeePayrollView, error) {
    // fetch from workforce DB, map to payroll's view type
}
```

---

### Mechanism 2 — Domain Event Bus (asynchronous reactions)

When something significant happens in Module A and Module B needs to react, Module A publishes a **domain event** onto the bus. Module B registers a handler. Neither module references the other.

```go
// awo/internal/eventbus/bus.go

type Event struct {
    Type       string
    TenantID   uuid.UUID
    OccurredAt time.Time
    Payload    any
}

type Handler func(ctx context.Context, event Event) error

type Bus interface {
    Publish(ctx context.Context, event Event) error
    Subscribe(eventType string, handler Handler)
}
```

**Canonical cross-module event flows:**

| Publisher | Event | Subscriber | Reaction |
|---|---|---|---|
| `pipeline` | `SalesOrderConfirmed` | `inventory` | Reserve stock for order |
| `pipeline` | `SalesOrderConfirmed` | `receivables` | Create draft customer invoice |
| `procurement` | `GoodsReceived` | `inventory` | Post inbound stock move |
| `procurement` | `GoodsReceived` | `payables` | Match bill lines to PO |
| `payroll` | `RunPostedToLedger` | `ledger` | Post payroll journal via PostingPort |
| `planning` | `ProductionOrderCompleted` | `inventory` | Receive finished goods to stock |
| `planning` | `ProductionOrderCompleted` | `ledger` | Post WIP → Finished Goods cost entry |
| `workforce` | `EmployeeTerminated` | `payroll` | Schedule final pay calculation |
| `workforce` | `EmployeeTerminated` | `iam` | Deactivate user account |
| `inventory` | `ReorderPointReached` | `procurement` | Auto-generate draft RFQ |
| `receivables` | `InvoiceOverdue` | `helpdesk` | Create internal collections alert |
| `helpdesk` | `SLABreached` | `notify` | Send escalation notification |

---

### Mechanism 3 — Workflow Sagas (multi-step orchestration)

For long-running processes that span multiple modules and require compensation on failure, a **saga** layer sits above all business modules. Each saga imports business modules directly but business modules never import sagas.

```go
// awo/sagas/procure_to_pay.go

// Procure-to-Pay Saga
// Step 1: procurement.ConfirmPO        → compensate: CancelPO
// Step 2: [await GoodsReceived event]  → compensate: raise exception ticket
// Step 3: inventory.Receive            → compensate: reverse stock move
// Step 4: payables.CreateBill from PO  → compensate: void bill
// Step 5: ledger.PostLines (via port)  → compensate: create reconciliation task
```

This can be implemented with simple saga structs initially and migrated to Temporal if the number of sagas or their complexity grows.

---

### Enforced Dependency Matrix

```
Module            │ platform/* │ ledger │ other core/* │ analytics/*
──────────────────┼────────────┼────────┼──────────────┼────────────
platform/*        │ limited ↓  │   ✗    │      ✗       │     ✗
core/ledger       │     ✓      │   —    │     PORT     │     ✗
core/* (others)   │     ✓      │  PORT  │     PORT     │     ✗
analytics/*       │     ✓      │  READ  │    READ      │     —
sagas/*           │     ✓      │   ✓    │      ✓       │     ✗

✓    = direct import allowed
PORT = via interface (port/adapter pattern) only; no direct package import
READ = read-only data queries; no service method calls
✗    = never import; hard rule
—    = self-reference / not applicable
```

---

### Wire Injection Pattern

The application root is the **only** place in the entire codebase that imports all modules simultaneously. This is by design — it is where the full wiring is visible and where cycles would be caught at compile time.

```go
// awo/cmd/server/wire.go

func InitializeApp(db *pgxpool.Pool, bus eventbus.Bus) (*App, error) {
    // Platform (no business deps)
    iamSvc        := iam.NewAccessService(db)
    tenantSvc     := tenant.NewTenantService(db)
    auditSvc      := audit.NewLogService(db)

    // Ledger (provides PostingPort)
    journalSvc    := ledger.NewJournalService(db, tenantSvc)

    // Workforce (provides EmployeeReader adapter for payroll)
    employeeSvc      := workforce.NewEmployeeService(db, iamSvc)
    workforceAdapter := workforce.NewPayrollAdapter(employeeSvc)

    // Attendance (provides WorkedHoursReader adapter for payroll)
    attendanceSvc    := attendance.NewAttendanceService(db)
    attendAdapter    := attendance.NewPayrollAdapter(attendanceSvc)

    // Payroll — injected with ports only; never knows about workforce or attendance packages
    payrollSvc := payroll.NewRunService(db, workforceAdapter, attendAdapter, journalSvc, bus)

    // Pipeline (provides InvoiceCreator port for receivables)
    pipelineSvc   := pipeline.NewSalesOrderService(db, bus)

    // Receivables — injected with PostingPort
    receivablesSvc := receivables.NewInvoiceService(db, journalSvc, bus)

    // ... and so on for every module

    // Subscribe event handlers
    bus.Subscribe("pipeline.SalesOrderConfirmed", inventory.HandleSalesOrderConfirmed(inventorySvc))
    bus.Subscribe("pipeline.SalesOrderConfirmed", receivables.HandleSalesOrderConfirmed(receivablesSvc))
    bus.Subscribe("workforce.EmployeeTerminated",  iam.HandleEmployeeTerminated(iamSvc))
    // ...

    return &App{ /* inject all services */ }, nil
}
```

---

### Summary: Seven Rules

1. **Platform is the bedrock.** It has no business module dependencies. Everything else may depend on it.

2. **Business modules never import each other.** The compiler enforces this if Wire generates the wiring; any accidental direct import shows up as a cycle build failure.

3. **Consumers define their ports.** If `payroll` needs employee data, `payroll` defines the `EmployeeReader` interface. `workforce` adapts to it. Ownership of the interface stays with the consumer.

4. **Side effects travel via the event bus.** Publish a domain event when something happens. Subscribers react independently. One module can never block another.

5. **Sagas orchestrate; modules execute.** Sagas know about modules. Modules never know about sagas.

6. **Analytics is purely read-side.** It queries data directly from the database (using read replicas or views) and never calls business service methods.

7. **`cmd/server` is the single wiring point.** It is the only package that sees the whole graph. All dependency setup happens there, and it is never imported by anyone else.

---

## Configuration & Feature Flags

### Overview

Awo has three tiers of configuration that cascade in priority order:

```
Global (Anthropic/Ops) → Tenant (Admin) → User (Preference)
         lowest                ↕               highest
                        feature flags
                        live at tenant
                        or global tier
```

A decision made at a lower tier can be **locked** by a higher tier, meaning the tenant cannot override a global constraint and a user cannot override a tenant policy.

---

### Package: `awo/platform/config`

**UI Label:** Settings → Organisation (tenant config) / internal ops panel (global config)

This package owns all runtime configuration. It is a platform-layer package — business modules read from it but never write to it directly.

#### Models

```
GlobalConfig   { Key, Value, DataType, Description, Locked bool, UpdatedAt, UpdatedBy }

TenantConfig   { TenantID, Key, Value, DataType, LockedByGlobal bool,
                 UpdatedAt, UpdatedBy }

UserPreference { UserID, TenantID, Key, Value, DataType, UpdatedAt }

ConfigSchema   { Key, DataType, DefaultValue, AllowedValues []string?,
                 Scope (global|tenant|user), Overridable bool,
                 Description, Module, Tags []string }
```

Every configuration key is declared once in `ConfigSchema`. The schema is the contract — if a key is not declared in the schema, it cannot be stored. This prevents configuration sprawl.

#### `config.ConfigService`

```go
// Reading — used by all business modules via injection
ConfigService.GetGlobal(ctx, key string) (ConfigValue, error)
ConfigService.GetTenant(ctx, tenantID uuid.UUID, key string) (ConfigValue, error)
ConfigService.GetUser(ctx, userID uuid.UUID, key string) (ConfigValue, error)

// Resolved — returns the effective value after cascade logic
ConfigService.Resolve(ctx, tenantID uuid.UUID, userID *uuid.UUID, key string) (ConfigValue, error)

// Writing — restricted by caller's permissions
ConfigService.SetGlobal(ctx, key string, value any) error           // ops only
ConfigService.SetTenant(ctx, tenantID uuid.UUID, key string, value any) error
ConfigService.SetUser(ctx, userID uuid.UUID, key string, value any) error

// Schema management
ConfigService.RegisterKey(ctx, schema ConfigSchema) error
ConfigService.GetSchema(ctx, key string) (*ConfigSchema, error)
ConfigService.ListSchema(ctx, module string) ([]*ConfigSchema, error)

// Bulk operations
ConfigService.GetTenantAll(ctx, tenantID uuid.UUID) (map[string]ConfigValue, error)
ConfigService.ResetTenantToDefault(ctx, tenantID uuid.UUID, key string) error
```

#### Cascade Resolution Logic

```go
// awo/platform/config/resolve.go

func (s *ConfigService) Resolve(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, key string) (ConfigValue, error) {
    schema, err := s.GetSchema(ctx, key)
    if err != nil { return ConfigValue{}, err }

    // 1. Start with the schema default
    val := schema.DefaultValue

    // 2. Apply global override (if set)
    if global, err := s.GetGlobal(ctx, key); err == nil {
        val = global
        // If global is locked, return immediately — no further overrides allowed
        if global.Locked {
            return val, nil
        }
    }

    // 3. Apply tenant override (if key is overridable at tenant scope)
    if schema.Overridable && schema.Scope != "global" {
        if tenant, err := s.GetTenant(ctx, tenantID, key); err == nil {
            val = tenant
            if tenant.Locked {
                return val, nil
            }
        }
    }

    // 4. Apply user preference (if key allows user-level override)
    if userID != nil && schema.Scope == "user" {
        if user, err := s.GetUser(ctx, *userID, key); err == nil {
            val = user
        }
    }

    return val, nil
}
```

#### Canonical Config Keys (examples)

| Key | Scope | Default | Description |
|---|---|---|---|
| `finance.default_currency` | tenant | `USD` | Default currency for new transactions |
| `finance.tax_inclusive_pricing` | tenant | `false` | Whether prices include tax by default |
| `finance.require_cost_centre` | tenant | `false` | Enforce cost centre on journal lines |
| `payroll.country` | tenant | — | Statutory compliance country |
| `payroll.payment_day` | tenant | `28` | Day of month payroll is run |
| `inventory.negative_stock_allowed` | tenant | `false` | Allow stock to go below zero |
| `inventory.auto_reorder` | tenant | `false` | Auto-create RFQs at reorder point |
| `pipeline.auto_assign_leads` | tenant | `true` | Round-robin lead assignment |
| `pipeline.require_approval_above` | tenant | `0` | Quote approval threshold amount |
| `helpdesk.default_sla_id` | tenant | — | SLA applied when no rule matches |
| `ui.date_format` | user | `YYYY-MM-DD` | Display date format preference |
| `ui.items_per_page` | user | `25` | Table pagination preference |
| `ui.sidebar_collapsed` | user | `false` | UI layout preference |

---

### Package: `awo/platform/flags`

**UI Label:** Settings → Feature Flags (visible to tenant admins for tenant-scoped flags)

Feature flags govern whether a capability is available at all, versus configuration which governs *how* a capability behaves. Every flag has an explicit owner scope.

#### Models

```
FeatureFlag  { Key, Name, Description, DefaultEnabled bool,
               Scope (global|tenant|user), Tags []string,
               RolloutStrategy (all|percentage|allowlist|denylist),
               RolloutConfig JSONB }

TenantFlag   { TenantID, FlagKey, Enabled bool, OverriddenBy, OverriddenAt,
               ExpiresAt? }

UserFlag     { UserID, TenantID, FlagKey, Enabled bool }

FlagAudit    { ID, FlagKey, TenantID?, ChangedBy, OldValue, NewValue, ChangedAt }
```

#### `flags.FlagService`

```go
FlagService.IsEnabled(ctx, key string, tenantID uuid.UUID, userID *uuid.UUID) (bool, error)
FlagService.EnableForTenant(ctx, tenantID uuid.UUID, key string) error
FlagService.DisableForTenant(ctx, tenantID uuid.UUID, key string) error
FlagService.EnableForUser(ctx, userID uuid.UUID, key string) error
FlagService.GetTenantFlags(ctx, tenantID uuid.UUID) (map[string]bool, error)
FlagService.RegisterFlag(ctx, flag FeatureFlag) error
FlagService.SetGlobalRollout(ctx, key string, strategy RolloutStrategy, config RolloutConfig) error
```

#### Flag Check Pattern in Services

Every flag check in business service code follows the same pattern. The flag key is always a string constant declared in the same package.

```go
// awo/core/planning/production_service.go

const FlagAdvancedScheduling = "manufacturing.advanced_scheduling"
const FlagMRPEnabled         = "manufacturing.mrp_enabled"

func (s *ProductionService) ScheduleOrders(ctx context.Context, tenantID uuid.UUID, horizon DateRange) error {
    mrpEnabled, _ := s.flags.IsEnabled(ctx, FlagMRPEnabled, tenantID, nil)

    if mrpEnabled {
        return s.runMRPSchedule(ctx, tenantID, horizon)
    }
    return s.runManualSchedule(ctx, tenantID, horizon)
}
```

#### Canonical Feature Flags (examples)

| Flag Key | Scope | Default | Controls |
|---|---|---|---|
| `finance.multi_currency` | tenant | `true` | Multi-currency transactions |
| `finance.budgets_module` | tenant | `true` | Budgeting UI and enforcement |
| `payroll.module` | tenant | `false` | Entire payroll module visibility |
| `manufacturing.mrp_enabled` | tenant | `false` | Material Requirements Planning |
| `manufacturing.advanced_scheduling` | tenant | `false` | Finite capacity scheduling |
| `helpdesk.module` | tenant | `true` | Customer support module |
| `helpdesk.ai_suggestions` | tenant | `false` | AI-powered article suggestions on tickets |
| `pipeline.ai_lead_scoring` | tenant | `false` | ML-based lead scoring |
| `analytics.predictive` | tenant | `false` | Forecasting and trend detection |
| `ui.beta_components` | user | `false` | Opt-in to beta UI components |

#### How Flags Drive UI Navigation

The navigation tree is not static. Every top-level section and many sub-items are conditionally rendered based on the effective flag state for the current tenant and user. This logic lives in the templ layout components.

```go
// awo/web/layout/nav.templ

templ SidebarNav(ctx context.Context, tenant *tenant.Tenant, user *iam.User, flags map[string]bool) {
    <nav>
        @NavItem("/accounting", "Accounting", IconLedger)       // always shown

        if flags["payroll.module"] {
            @NavItem("/people/payroll", "Payroll", IconPayroll)
        }

        if flags["manufacturing.module"] {
            @NavGroup("Production", IconFactory) {
                @NavItem("/production/orders", "Production Orders", IconOrders)
                if flags["manufacturing.mrp_enabled"] {
                    @NavItem("/production/mrp", "MRP", IconMRP)
                }
            }
        }

        if flags["helpdesk.module"] {
            @NavItem("/support", "Support", IconTicket)
        }

        if flags["analytics.predictive"] {
            @NavItem("/analytics/insights", "Insights", IconInsights)
        }
    </nav>
}
```

---

### How Config and Flags Are Injected into Services

Both `ConfigService` and `FlagService` are injected as interfaces into every business service that needs them. They are never called globally or via a singleton — this keeps the services testable.

```go
// awo/core/inventory/stock_service.go

type StockService struct {
    db     *pgxpool.Pool
    config config.Reader   // interface — not the concrete ConfigService
    flags  flags.Reader    // interface — not the concrete FlagService
    bus    eventbus.Bus
    poster ledger.PostingPort
}

// config.Reader interface — defined in awo/platform/config
type Reader interface {
    Resolve(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, key string) (ConfigValue, error)
}

// flags.Reader interface — defined in awo/platform/flags
type Reader interface {
    IsEnabled(ctx context.Context, key string, tenantID uuid.UUID, userID *uuid.UUID) (bool, error)
}
```

---

## Fiber API Layer

### Overview

The Awo HTTP layer is built on [Fiber](https://gofiber.io/) (v2). It sits above the service layer and is entirely responsible for HTTP concerns: routing, middleware, request parsing, response rendering, and HTMX interactions. It never contains business logic.

```
Browser / Client
      │
      ▼
┌─────────────────────────────────────────┐
│              Fiber App                   │
│  Middleware stack (applied globally)     │
│  Router groups (one per module)          │
└──────────────┬──────────────────────────┘
               │ calls
               ▼
┌─────────────────────────────────────────┐
│           Handler Layer                  │
│  Parse request → call service → render  │
│  templ component or return JSON         │
└──────────────┬──────────────────────────┘
               │ calls
               ▼
┌─────────────────────────────────────────┐
│          Service Layer (core/*)          │
│  All business logic lives here          │
│  No knowledge of HTTP                   │
└─────────────────────────────────────────┘
```

**Hard rule:** Service structs have zero imports from `github.com/gofiber/fiber`. HTTP is a delivery mechanism. The service layer is the system.

---

### Package Structure

```
awo/
└── web/
    ├── middleware/
    │   ├── auth.go          → session validation, identity extraction
    │   ├── tenant.go        → tenant resolution from hostname or header
    │   ├── rbac.go          → permission check middleware
    │   ├── audit.go         → automatic audit log recording
    │   ├── flags.go         → inject resolved flags into context
    │   ├── ratelimit.go     → per-tenant and per-user rate limiting
    │   ├── recovery.go      → panic recovery and structured error response
    │   └── logger.go        → structured request logging
    │
    ├── handlers/
    │   ├── ledger/
    │   │   ├── accounts.go
    │   │   └── journal.go
    │   ├── payroll/
    │   │   ├── runs.go
    │   │   └── payslips.go
    │   ├── inventory/
    │   │   └── stock.go
    │   └── ... (one folder per core module)
    │
    ├── router/
    │   └── router.go        → registers all route groups
    │
    ├── layout/
    │   ├── base.templ        → page shell
    │   └── nav.templ         → flag-driven navigation
    │
    └── components/           → shared UI components (ruun library)
```

---

### Middleware Stack

Middleware is applied in a deliberate order. Each layer adds information to the request context that the next layer can rely on.

```go
// awo/web/router/router.go

func New(app *fiber.App, services *Services) {
    // 1. Always-on infrastructure middleware
    app.Use(middleware.Recovery())    // catch panics, return 500
    app.Use(middleware.Logger())      // structured request log
    app.Use(middleware.RateLimit())   // global rate limit

    // 2. Tenant resolution — must come before auth
    //    Reads from subdomain (acme.awo.app) or X-Tenant-Slug header
    app.Use(middleware.ResolveTenant(services.Tenant))

    // 3. Authentication — reads session cookie or Bearer token
    //    Populates ctx with Identity + User
    app.Use(middleware.Authenticate(services.Auth))

    // 4. Flag injection — loads tenant's effective feature flags into ctx
    //    Handlers can read flags without calling FlagService directly
    app.Use(middleware.InjectFlags(services.Flags))

    // 5. Audit wrapping — records every mutating request after it completes
    app.Use(middleware.AuditWrap(services.Audit))

    // Route groups registered after middleware
    registerRoutes(app, services)
}
```

---

### Middleware Implementations

#### `middleware.ResolveTenant`

```go
func ResolveTenant(tenantSvc *tenant.TenantService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        slug := extractSlug(c) // from subdomain or header
        t, err := tenantSvc.GetBySlug(c.Context(), slug)
        if err != nil {
            return fiber.ErrNotFound
        }
        if t.Status != "active" {
            return fiber.NewError(fiber.StatusForbidden, "tenant suspended")
        }
        // Store in context — all downstream handlers use ContextTenant(c)
        c.Locals(contextKeyTenant, t)
        return c.Next()
    }
}

func ContextTenant(c *fiber.Ctx) *tenant.Tenant {
    return c.Locals(contextKeyTenant).(*tenant.Tenant)
}
```

#### `middleware.Authenticate`

```go
func Authenticate(authSvc *iam.AuthService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := extractToken(c) // cookie or Authorization: Bearer
        if token == "" {
            if isHTMXRequest(c) {
                // HTMX partial — redirect via HX-Redirect header
                c.Set("HX-Redirect", "/login")
                return c.SendStatus(fiber.StatusUnauthorized)
            }
            return c.Redirect("/login")
        }

        session, err := authSvc.ValidateSession(c.Context(), token)
        if err != nil {
            return c.Redirect("/login")
        }

        c.Locals(contextKeySession, session)
        c.Locals(contextKeyUserID, session.UserID)
        return c.Next()
    }
}
```

#### `middleware.RequirePermission` (per-route RBAC)

```go
// Applied to individual routes or route groups, not globally
func RequirePermission(action, resource string, accessSvc *iam.AccessService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := ContextUserID(c)
        tenant := ContextTenant(c)

        // ABAC context: include request-level attributes for policy evaluation
        attrs := map[string]any{
            "tenant_id": tenant.ID,
            "ip":        c.IP(),
            "method":    c.Method(),
        }

        allowed, err := accessSvc.CanWithContext(c.Context(), userID, action, resource, attrs)
        if err != nil || !allowed {
            if isHTMXRequest(c) {
                return c.Status(fiber.StatusForbidden).SendString("Access denied")
            }
            return fiber.ErrForbidden
        }
        return c.Next()
    }
}
```

#### `middleware.RequireFlag` (per-route flag gate)

```go
func RequireFlag(flagKey string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        flags := ContextFlags(c) // already resolved by InjectFlags middleware
        if !flags[flagKey] {
            // Route simply does not exist for this tenant
            return fiber.ErrNotFound
        }
        return c.Next()
    }
}
```

---

### Router Groups

Every module's routes are registered in a dedicated function. Routes are structured to mirror the UI navigation.

```go
// awo/web/router/routes.go

func registerRoutes(app *fiber.App, s *Services) {
    // ── Accounting ────────────────────────────────────────────────
    acct := app.Group("/accounting",
        middleware.RequirePermission("read", "accounting", s.Access))

    acct.Get("/",               s.LedgerHandlers.Overview)
    acct.Get("/accounts",       s.LedgerHandlers.ListAccounts)
    acct.Post("/accounts",      middleware.RequirePermission("create", "accounts", s.Access),
                                s.LedgerHandlers.CreateAccount)
    acct.Get("/journal",        s.LedgerHandlers.ListEntries)
    acct.Post("/journal",       middleware.RequirePermission("create", "journal_entries", s.Access),
                                s.LedgerHandlers.CreateEntry)
    acct.Post("/journal/:id/post", middleware.RequirePermission("post", "journal_entries", s.Access),
                                s.LedgerHandlers.PostEntry)
    acct.Get("/reports/balance-sheet",  s.LedgerHandlers.BalanceSheet)
    acct.Get("/reports/income",         s.LedgerHandlers.IncomeStatement)

    // ── Payroll (flag-gated entire group) ─────────────────────────
    payroll := app.Group("/people/payroll",
        middleware.RequireFlag("payroll.module"),
        middleware.RequirePermission("read", "payroll", s.Access))

    payroll.Get("/",             s.PayrollHandlers.Overview)
    payroll.Get("/runs",         s.PayrollHandlers.ListRuns)
    payroll.Post("/runs",        middleware.RequirePermission("create", "payroll_runs", s.Access),
                                 s.PayrollHandlers.CreateRun)
    payroll.Post("/runs/:id/process", s.PayrollHandlers.ProcessRun)
    payroll.Post("/runs/:id/approve", middleware.RequirePermission("approve", "payroll_runs", s.Access),
                                 s.PayrollHandlers.ApproveRun)
    payroll.Get("/payslips/:id", s.PayrollHandlers.GetPayslip)

    // ── Inventory ─────────────────────────────────────────────────
    inv := app.Group("/inventory",
        middleware.RequirePermission("read", "inventory", s.Access))

    inv.Get("/stock",            s.InventoryHandlers.StockLevels)
    inv.Get("/moves",            s.InventoryHandlers.ListMoves)
    inv.Post("/adjustments",     middleware.RequirePermission("create", "stock_adjustments", s.Access),
                                 s.InventoryHandlers.CreateAdjustment)

    // ── Manufacturing (flag-gated) ─────────────────────────────────
    mfg := app.Group("/production",
        middleware.RequireFlag("manufacturing.module"),
        middleware.RequirePermission("read", "production", s.Access))

    mfg.Get("/orders",           s.PlanningHandlers.ListOrders)
    mfg.Post("/orders",          s.PlanningHandlers.CreateOrder)

    // MRP sub-section — nested flag gate
    mfg.Get("/mrp",
        middleware.RequireFlag("manufacturing.mrp_enabled"),
        s.PlanningHandlers.MRPView)

    // ── REST API (for external integrations) ──────────────────────
    api := app.Group("/api/v1",
        middleware.AuthenticateAPIKey(s.Bridge))

    api.Get("/invoices",         s.ReceivablesHandlers.APIListInvoices)
    api.Post("/invoices",        s.ReceivablesHandlers.APICreateInvoice)
    api.Get("/products",         s.CatalogHandlers.APIListProducts)
    // ... etc
}
```

---

### Handler Pattern

Every handler follows the same four-step pattern: **extract → validate → call service → respond**.

```go
// awo/web/handlers/payroll/runs.go

type RunHandlers struct {
    runSvc  *payroll.RunService
    config  config.Reader
    flags   flags.Reader
}

// CreateRun handles POST /people/payroll/runs
// Supports both full-page and HTMX partial responses
func (h *RunHandlers) CreateRun(c *fiber.Ctx) error {
    tenant := middleware.ContextTenant(c)
    userID := middleware.ContextUserID(c)

    // 1. Parse and validate input
    var input struct {
        PeriodStart string `form:"period_start"`
        PeriodEnd   string `form:"period_end"`
    }
    if err := c.BodyParser(&input); err != nil {
        return renderFormError(c, "Invalid input")
    }

    start, err := time.Parse("2006-01-02", input.PeriodStart)
    if err != nil {
        return renderFormError(c, "Invalid start date")
    }
    end, err := time.Parse("2006-01-02", input.PeriodEnd)
    if err != nil {
        return renderFormError(c, "Invalid end date")
    }

    // 2. Call service — handler passes tenant/user context; service handles logic
    run, err := h.runSvc.CreateRun(c.Context(), tenant.ID, start, end)
    if err != nil {
        return renderServiceError(c, err)
    }

    // 3. Respond — HTMX partial or full redirect
    if isHTMXRequest(c) {
        // Swap the runs table row, trigger a toast notification
        c.Set("HX-Trigger", `{"showToast": "Payroll run created"}`)
        return renderComponent(c, payrollComponents.RunRow(run))
    }
    return c.Redirect("/people/payroll/runs/" + run.ID.String())
}
```

#### HTMX Response Helpers

```go
// awo/web/handlers/respond.go

func isHTMXRequest(c *fiber.Ctx) bool {
    return c.Get("HX-Request") == "true"
}

func renderComponent(c *fiber.Ctx, component templ.Component) error {
    c.Set("Content-Type", "text/html")
    return component.Render(c.Context(), c.Response().BodyWriter())
}

func renderFormError(c *fiber.Ctx, msg string) error {
    if isHTMXRequest(c) {
        return renderComponent(c, shared.FormError(msg))
    }
    return fiber.NewError(fiber.StatusBadRequest, msg)
}

func renderServiceError(c *fiber.Ctx, err error) error {
    var domainErr *domain.Error
    if errors.As(err, &domainErr) {
        return renderFormError(c, domainErr.Message)
    }
    // Unexpected error — log and return generic message
    slog.Error("service error", "err", err)
    return fiber.ErrInternalServerError
}
```

---

### Context Value Flow

The diagram below shows what each middleware layer adds to the Fiber context and what handlers/services can read from it.

```
Request arrives
    │
    ▼ ResolveTenant
    │  ctx ← Tenant{ID, Slug, Plan, Status}
    │
    ▼ Authenticate
    │  ctx ← Session{UserID, IdentityID}
    │  ctx ← User{FullName, Locale, Timezone}
    │
    ▼ InjectFlags
    │  ctx ← map[string]bool{flagKey: enabled, ...}
    │
    ▼ RequirePermission (route-level)
    │  checks AccessService.CanWithContext — blocks here if denied
    │
    ▼ Handler
    │  reads: ContextTenant(c)     → *tenant.Tenant
    │  reads: ContextUserID(c)     → uuid.UUID
    │  reads: ContextFlags(c)      → map[string]bool
    │  calls: service.Method(c.Context(), tenant.ID, ...)
    │
    ▼ Service
       receives: standard context.Context
       reads:    tenantID passed as explicit parameter
       reads:    config.Resolve(ctx, tenantID, ...)
       reads:    flags.IsEnabled(ctx, key, tenantID, nil)
       NO direct access to fiber.Ctx — services are HTTP-agnostic
```

---

### REST API vs Server-Side Rendered Routes

Awo exposes two surface areas from the same Fiber application:

| Surface | Path Prefix | Auth | Response Format | Consumer |
|---|---|---|---|---|
| SSR + HTMX | `/` | Session cookie | HTML (templ) | Browser |
| REST API | `/api/v1/` | API Key (Bearer) | JSON | External integrations, mobile |
| Webhook receiver | `/hooks/` | HMAC signature | — | Incoming webhooks |

The same service methods power all three. The difference is only in the handler's parsing and response code.

```go
// SSR handler
func (h *InvoiceHandlers) ListInvoices(c *fiber.Ctx) error {
    invoices, _, _ := h.invoiceSvc.ListInvoices(c.Context(), filter)
    return renderComponent(c, invoiceComponents.InvoiceTable(invoices))
}

// API handler — same service, different response
func (h *InvoiceHandlers) APIListInvoices(c *fiber.Ctx) error {
    invoices, pagination, _ := h.invoiceSvc.ListInvoices(c.Context(), filter)
    return c.JSON(APIResponse{Data: invoices, Meta: pagination})
}
```

---

## Condition Evaluation Engine

### Package: `pkg/condition`

**Import path:** `github.com/mustafe/awo/pkg/condition`  
**Purpose:** Production-ready, tenant-safe runtime condition evaluation. This is a pure library — it has zero knowledge of Awo's domain. It is used by `awo/platform/rules` (business rule engine), `awo/platform/flags` (flag targeting), and `awo/platform/iam` (ABAC policy evaluation).

```
awo/platform/rules
awo/platform/flags      → all import and wrap pkg/condition
awo/platform/iam
         │
         ▼
   pkg/condition        ← pure library, no ERP imports
         │
         ├── Evaluator          (thread-safe, reusable across requests)
         ├── EvalContext        (per-evaluation, holds facts + functions)
         ├── ConditionGroup     (AND/OR/NOT tree node)
         ├── ConditionRule      (single field comparison or expr formula)
         ├── Builder            (fluent API for programmatic rule construction)
         └── regexCache         (bounded LRU, ReDoS-safe via regexp2)
```

---

### Core Types

#### `condition.ConditionGroup`

The tree node. Each group has a `Conjunction` (AND or OR), an optional `Not` flag, and a `Children []any` slice that holds either `*ConditionRule` or nested `*ConditionGroup` values.

```go
type ConditionGroup struct {
    ID          string      `json:"id"`
    Conjunction Conjunction `json:"conjunction"` // "and" | "or"
    Not         bool        `json:"not,omitempty"`
    Children    []any       `json:"children"`
    If          string      `json:"if,omitempty"` // expr-lang formula override
}
```

When `If` is set on a group, the `expr-lang` formula is evaluated and its boolean result replaces the entire child evaluation. This provides a formula escape hatch for conditions that can't be expressed with the standard operators.

#### `condition.ConditionRule`

A single comparison. The `Left` expression is always a field reference; the `Right` can be a literal value, another field reference, or a custom function call.

```go
type ConditionRule struct {
    ID    string       `json:"id"`
    Left  Expression   `json:"left"`
    Op    OperatorType `json:"op"`
    Right any          `json:"right,omitempty"`
    If    string       `json:"if,omitempty"` // expr-lang formula override
}
```

#### `condition.Expression`

```go
type Expression struct {
    Type  ValueType `json:"type"`   // "value" | "field" | "func"
    Value any       `json:"value,omitempty"`
    Field string    `json:"field,omitempty"` // dot-notation e.g. "po.supplier.tier"
    Func  *FuncCall `json:"func,omitempty"`
}
```

Field expressions support arbitrary dot-notation depth. `GetValue("po.supplier.tier")` traverses nested `map[string]any` or struct fields via reflection.

#### Available Operators

| Operator constant | Symbol | Notes |
|---|---|---|
| `OpEqual` | `=` | Type-aware: numeric, time, bool, string |
| `OpNotEqual` | `!=` | |
| `OpLess` | `<` | Numeric or time |
| `OpLessOrEqual` | `<=` | |
| `OpGreater` | `>` | |
| `OpGreaterOrEqual` | `>=` | |
| `OpBetween` | between | Inclusive; right must have 2 values |
| `OpNotBetween` | not between | |
| `OpIsEmpty` | is empty | Unary: nil, `""`, `[]`, `{}` |
| `OpIsNotEmpty` | is not empty | Unary |
| `OpContains` | contains | Substring or slice membership |
| `OpNotContains` | not contains | |
| `OpStartsWith` | starts with | |
| `OpEndsWith` | ends with | |
| `OpIn` | in | `needle` found in `[]right` |
| `OpNotIn` | not in | |
| `OpMatchRegexp` | matches | `regexp2` with ReDoS timeout |

---

### `condition.Evaluator`

The evaluator is **created once per application** (or once per rule set configuration) and reused across every evaluation call. It holds the compiled regex cache and the compiled `expr-lang` program cache. It is fully thread-safe.

```go
// Created once at startup in awo/cmd/server/wire.go
evaluator := condition.NewEvaluator(nil, condition.DefaultEvalOptions())

// DefaultEvalOptions:
//   MaxDepth:      10   — nesting limit, prevents stack overflow
//   MaxConditions: 1000 — total node limit, prevents DoS
//   Timeout:       30s  — per-evaluation timeout
//   RegexTimeout:  100ms — per-regex match timeout (ReDoS protection)
//   CacheResults:  true
```

The `Config` parameter is optional. When provided it overrides `MaxDepth` if the config's `MaxLevel` is more restrictive. Awo passes `nil` and controls limits via `EvalOptions` directly.

#### `condition.EvalContext`

Created **fresh for each evaluation call**. It holds the fact data, registered functions, and collects metrics.

```go
// awo/platform/rules/engine.go — built by RuleEngine before each Evaluate call

evalCtx := condition.NewEvalContext(
    map[string]any{
        "po":        poFacts,          // nested map — dot notation traversal
        "requester": requesterFacts,
        "tenant":    tenantFacts,
    },
    condition.DefaultEvalOptions(),
)

// Register any tenant-scoped custom functions for this trigger event
for name, handler := range s.registry.FunctionsFor(triggerEvent) {
    evalCtx.RegisterFunction(name, handler)
}

matched, err := s.evaluator.Evaluate(ctx, conditionGroup, evalCtx)
metrics := evalCtx.GetMetrics() // RulesEvaluated, Duration, CacheHits, Errors
```

---

### Database Round-Trip: `UnmarshalJSON` for `ConditionGroup`

Rule trees are stored as JSONB in PostgreSQL. The standard `json.Unmarshal` cannot infer whether each element of `Children []any` is a `ConditionGroup` or a `ConditionRule` without inspection. A custom unmarshaller handles this using the presence of the `"conjunction"` key as the discriminator.

```go
// awo/platform/rules/unmarshal.go
// This lives in platform/rules, not in pkg/condition, so the library stays import-free.

func (g *condition.ConditionGroup) UnmarshalJSON(data []byte) error {
    type Alias condition.ConditionGroup
    var raw struct {
        Alias
        Children []json.RawMessage `json:"children"`
    }
    if err := json.Unmarshal(data, &raw); err != nil {
        return err
    }
    *g = condition.ConditionGroup(raw.Alias)

    for _, childData := range raw.Children {
        // Peek at the child to determine its type
        var probe struct {
            Conjunction string `json:"conjunction"`
        }
        _ = json.Unmarshal(childData, &probe)

        if probe.Conjunction != "" {
            var group condition.ConditionGroup
            if err := json.Unmarshal(childData, &group); err != nil {
                return fmt.Errorf("nested group: %w", err)
            }
            g.Children = append(g.Children, &group)
        } else {
            var rule condition.ConditionRule
            if err := json.Unmarshal(childData, &rule); err != nil {
                return fmt.Errorf("condition rule: %w", err)
            }
            g.Children = append(g.Children, &rule)
        }
    }
    return nil
}
```

Once loaded from the database this way, the evaluator's type switch hits the `*ConditionGroup` and `*ConditionRule` branches directly, avoiding the `mapToStruct` JSON round-trip fallback entirely.

---

### Function Registry

Custom functions extend the condition engine beyond static field comparisons. They are registered at the `RuleEngine` level and injected into `EvalContext` per trigger event. This keeps tenant-scoped business logic out of the library.

```go
// awo/platform/rules/registry.go

type TriggerRegistry struct {
    mu        sync.RWMutex
    fields    map[string][]condition.Field            // triggerEvent → available fields
    functions map[string]map[string]condition.FuncHandler // triggerEvent → funcName → handler
}

// Modules call RegisterTrigger at startup (in their Wire provider)
func (r *TriggerRegistry) RegisterTrigger(event string, fields []condition.Field, funcs ...TriggerFunc) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.fields[event] = fields
    for _, f := range funcs {
        r.functions[event][f.Name] = f.Handler
    }
}

// Used by RuleEngine.Evaluate to populate EvalContext before calling the evaluator
func (r *TriggerRegistry) FunctionsFor(event string) map[string]condition.FuncHandler {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.functions[event]
}

// Used by the Rules UI to populate the field picker
func (r *TriggerRegistry) FieldsFor(event string) []condition.Field {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.fields[event]
}
```

**Example trigger registration** (called from each module's Wire provider at startup):

```go
// awo/core/procurement/wire_provider.go

func RegisterProcurementRules(registry *rules.TriggerRegistry, supplierSvc *SupplierService) {
    registry.RegisterTrigger(
        "procurement.POSubmittedForApproval",

        // Fields available in the condition builder UI
        []condition.Field{
            {Name: "po.total",         Label: "PO Total Amount",    Type: condition.FieldTypeNumber},
            {Name: "po.currency",      Label: "Currency",           Type: condition.FieldTypeSelect,
             Values: []condition.SelectOption{
                 {Label: "USD", Value: "USD"}, {Label: "KES", Value: "KES"},
             }},
            {Name: "po.category",      Label: "Purchase Category",  Type: condition.FieldTypeText},
            {Name: "po.supplier_tier", Label: "Supplier Tier",      Type: condition.FieldTypeSelect,
             Values: []condition.SelectOption{
                 {Label: "Preferred", Value: "preferred"},
                 {Label: "Approved",  Value: "approved"},
                 {Label: "New",       Value: "new"},
             }},
            {Name: "requester.role",         Label: "Requester Role",       Type: condition.FieldTypeText},
            {Name: "requester.department",   Label: "Requester Department", Type: condition.FieldTypeText},
            {Name: "requester.tenure_years", Label: "Requester Tenure (yrs)", Type: condition.FieldTypeNumber},
        },

        // Custom functions available in formula rules (If: "...") for this trigger
        rules.TriggerFunc{
            Name: "supplier_score",
            Handler: func(ctx context.Context, args []any, evalCtx *condition.EvalContext) (any, error) {
                supplierID, _ := args[0].(string)
                score, err := supplierSvc.GetScore(ctx, uuid.MustParse(supplierID))
                return score, err
            },
        },
    )
}
```

---

### `condition.Builder` — Programmatic Rule Construction

The `Builder` provides a fluent API for constructing condition trees in Go code, used in tests and for seeding system default rule sets.

```go
// Building a rule set programmatically (e.g., system defaults or test fixtures)

group := condition.NewBuilder(condition.ConjunctionAnd).
    AddRule("po.total", condition.OpGreaterOrEqual, 10000.00).
    AddRule("po.currency", condition.OpEqual, "USD").
    Build()

// Nested groups
outerGroup := condition.NewBuilder(condition.ConjunctionOr).
    AddRule("requester.role", condition.OpEqual, "engineer").
    AddGroup(
        condition.NewBuilder(condition.ConjunctionAnd).
            AddRule("po.total", condition.OpGreater, 5000.00).
            AddRule("po.supplier_tier", condition.OpNotEqual, "preferred").
            Build(),
    ).
    Build()

// Formula escape hatch for expressions too complex for standard operators
// Uses expr-lang: https://expr-lang.github.io
creditCheckGroup := condition.NewBuilder(condition.ConjunctionAnd).
    AddFormula(`po_total + customer_outstanding > customer_credit_limit * 0.9`).
    Build()

// Field-to-field comparison
fieldComp := condition.NewBuilder(condition.ConjunctionAnd).
    AddFieldComparison("invoice.amount", "customer.credit_limit", condition.OpLessOrEqual).
    Build()

// Between
rangeRule := condition.NewBuilder(condition.ConjunctionAnd).
    AddBetweenRule("po.total", 500.00, 9999.99).
    Build()
```

---

### Resource Limits & Safety

| Concern | Protection | Default |
|---|---|---|
| Infinite recursion | `MaxDepth` check in `evaluateWithDepth` | 10 levels |
| DoS via rule explosion | `MaxConditions` atomic counter | 1,000 nodes |
| Slow evaluation | `context.WithTimeout` wrapping | 30 seconds |
| ReDoS via regex | `regexp2.MatchTimeout` per match | 100 ms |
| Regex cache unbounded | LRU eviction at 25% batch | 1,000 entries |
| Overly complex patterns | `MaxRegexPatternLength` check | 1,000 chars |
| Overly complex formulas | `MaxFormulaLength` check before compile | 10,000 chars |
| expr-lang compile cost | `programCache sync.Map` (compile once, run many) | unbounded |

The `EvalContext.GetMetrics()` method returns a snapshot of `RulesEvaluated`, `GroupsEvaluated`, `Duration`, `CacheHits`, `CacheMisses`, and `Errors` — all collected atomically without locking. These are logged by `RuleEngine` after every evaluation and emitted to the audit log for high-value rule executions (approvals, credit checks).

---

## Tenant Business Rules & Workflow Decisions

### Overview

Business rules are the mechanism by which tenant administrators codify their organisation's policies into Awo's runtime decision engine. Rather than hard-coding behaviour, Awo evaluates rules at the moment a workflow step reaches a decision point.

The evaluation engine is powered by **`pkg/condition`** — a purpose-built, production-ready runtime condition evaluator that lives as a standalone internal package. The `awo/platform/rules` package is a thin orchestration layer on top of it: it handles database persistence, tenant scoping, trigger registration, action execution, and approval chain management. It never re-implements condition evaluation.

```
Workflow step reaches a decision point
           │
           ▼
    rules.RuleEngine.Evaluate(ctx, tenantID, triggerEvent, facts)
           │
           ▼  loads active RuleSets from DB (ordered by priority)
           │
           ▼  for each Rule:
              builds condition.EvalContext from FactContext
              calls condition.Evaluator.Evaluate(ctx, rule.ConditionGroup, evalCtx)
           │
    ┌──────┴──────────────────────────────────────┐
    │                                              │
    ▼                                              ▼
Conditions matched                        No conditions matched
    │                                              │
    ▼                                              ▼
Execute RuleActions                     Continue to next rule / fallback
```

---

### Package: `pkg/condition`

**Import path:** `awo/pkg/condition`
**Purpose:** Standalone, HTTP-agnostic runtime condition evaluator. Contains zero ERP domain concepts. Can be used independently for form validation, feature flag evaluation, and any other rule-based decision in Awo.

This package is **not** modified per-module. Business modules interact with it only through `awo/platform/rules`.

#### Core Types

```go
// ConditionGroup — a node that groups rules with AND/OR logic and optional NOT negation.
// Children can be *ConditionRule or nested *ConditionGroup (arbitrary depth up to MaxDepth).
// When the If field is non-empty, the expr-lang formula takes precedence over Children.
type ConditionGroup struct {
    ID          string      `json:"id"`
    Conjunction Conjunction `json:"conjunction"` // "and" | "or"
    Not         bool        `json:"not,omitempty"`
    Children    []any       `json:"children"`    // []*ConditionRule | *ConditionGroup
    If          string      `json:"if,omitempty"` // expr-lang formula override
}

// ConditionRule — a single field comparison.
// When If is set, it is a raw expr-lang formula and Left/Op/Right are ignored.
type ConditionRule struct {
    ID    string       `json:"id"`
    Left  Expression   `json:"left"`
    Op    OperatorType `json:"op"`
    Right any          `json:"right,omitempty"`
    If    string       `json:"if,omitempty"` // expr-lang formula override
}

// Expression — the left or right side of a rule.
// Type "value"  → literal value
// Type "field"  → dot-notation path into EvalContext.Data  (e.g. "po.total")
// Type "func"   → registered custom function call
type Expression struct {
    Type  ValueType `json:"type"`
    Value any       `json:"value,omitempty"`
    Field string    `json:"field,omitempty"`
    Func  *FuncCall `json:"func,omitempty"`
}

// EvalContext — runtime state provided by the caller for a single evaluation.
// Data holds the FactContext. Functions holds tenant-registered custom functions.
// Thread-safe for concurrent reads.
type EvalContext struct {
    Data      map[string]any
    Functions map[string]FuncHandler
    Fields    map[string]Field
    Now       time.Time
    // internal: maxConditions, conditionCount (atomic), metrics
}
```

#### Supported Operators

| Operator constant | Meaning |
|---|---|
| `OpEqual` | `=` |
| `OpNotEqual` | `!=` |
| `OpLess` | `<` |
| `OpLessOrEqual` | `<=` |
| `OpGreater` | `>` |
| `OpGreaterOrEqual` | `>=` |
| `OpBetween` | inclusive range, requires 2 right values |
| `OpNotBetween` | outside range |
| `OpIsEmpty` | unary — nil, empty string, empty slice |
| `OpIsNotEmpty` | unary |
| `OpContains` | substring / slice membership |
| `OpNotContains` | inverse |
| `OpStartsWith` | prefix match |
| `OpEndsWith` | suffix match |
| `OpIn` | value in list |
| `OpNotIn` | value not in list |
| `OpMatchRegexp` | regexp2 with ReDoS protection and per-match timeout |

#### Resource Limits (built-in, non-negotiable)

| Limit | Constant | Default |
|---|---|---|
| Nesting depth | `DefaultMaxDepth` | 10 |
| Conditions per evaluation | `DefaultMaxConditions` | 1000 |
| Evaluation timeout | `DefaultTimeout` | 30s |
| Regex match timeout | `DefaultRegexTimeout` | 100ms |
| Regex pattern length | `MaxRegexPatternLength` | 1000 chars |
| Formula length | `MaxFormulaLength` | 10,000 chars |
| Compiled regex cache | `MaxRegexCacheSize` | 1000 entries (LRU) |

#### `condition.Evaluator`

```go
// NewEvaluator creates a thread-safe evaluator. Config should not be modified after creation.
// EvalOptions.MaxDepth, MaxConditions, Timeout, RegexTimeout all have safe defaults.
func NewEvaluator(config *Config, opts EvalOptions) *Evaluator

// Evaluate is the single entry point. Thread-safe — may be called concurrently.
// Returns (true, nil)  when the root condition is satisfied.
// Returns (false, ErrEvaluationTimeout)     if context deadline is exceeded.
// Returns (false, ErrMaxDepthExceeded)      if nesting exceeds MaxDepth.
// Returns (false, ErrResourceLimitExceeded) if condition count exceeds MaxConditions.
// Returns (false, ErrRegexTimeout)          if a regex match times out (ReDoS protection).
func (e *Evaluator) Evaluate(ctx context.Context, root any, evalCtx *EvalContext) (bool, error)
```

#### `condition.Builder` — Programmatic Rule Construction

The `Builder` provides a fluent API for constructing condition trees in Go code. Used by tests, seed data, and migration scripts — not by the rules UI, which builds JSON directly.

```go
// Simple comparison: po.total >= 10000 AND requester.role != "department_head"
group := condition.NewBuilder(condition.ConjunctionAnd).
    AddRule("po.total", condition.OpGreaterOrEqual, 10000.0).
    AddRule("requester.role", condition.OpNotEqual, "department_head").
    Build()

// BETWEEN range
group := condition.NewBuilder(condition.ConjunctionAnd).
    AddBetweenRule("po.total", 500.0, 9999.99).
    Build()

// IN list
group := condition.NewBuilder(condition.ConjunctionOr).
    AddInRule("po.category", "IT Equipment", "Software", "Cloud Services").
    Build()

// expr-lang formula for cross-field arithmetic
group := condition.NewBuilder(condition.ConjunctionAnd).
    AddFormula("po.total > customer.credit_limit * 0.9 && po.currency == tenant.default_currency").
    Build()

// Nested groups: (A AND B) OR (C AND D)
inner1 := condition.NewBuilder(condition.ConjunctionAnd).
    AddRule("po.total", condition.OpGreaterOrEqual, 10000.0).
    AddRule("po.supplier_tier", condition.OpEqual, "new").
    Build()

inner2 := condition.NewBuilder(condition.ConjunctionAnd).
    AddRule("requester.role", condition.OpEqual, "intern").
    AddRule("po.category", condition.OpEqual, "Capital Expenditure").
    Build()

root := condition.NewBuilder(condition.ConjunctionOr).
    AddGroup(inner1).
    AddGroup(inner2).
    Build()
```

#### `condition.EvalContext` — Building the Runtime Environment

```go
// NewEvalContext creates the runtime state for a single evaluation pass.
func NewEvalContext(data map[string]any, opts EvalOptions) *EvalContext

// GetValue resolves dot-notation paths from Data.
// "po.total"             → data["po"]["total"]
// "requester.department" → data["requester"]["department"]
// Supports both map[string]any nesting and Go struct fields via reflection.
func (ctx *EvalContext) GetValue(path string) (any, error)

// RegisterFunction adds a custom FuncHandler. Called per-evaluation, not globally.
func (ctx *EvalContext) RegisterFunction(name string, handler FuncHandler) error
```

#### Custom Functions

Tenant-scoped functions extend the operator set for domain-specific logic.

```go
type FuncHandler func(ctx context.Context, args []any, evalCtx *EvalContext) (any, error)

// Example: custom function checking if a supplier has an active contract
supplierContractFn := func(ctx context.Context, args []any, evalCtx *EvalContext) (any, error) {
    if len(args) == 0 {
        return false, errors.New("supplier_has_contract: supplierID arg required")
    }
    supplierID, ok := args[0].(string)
    if !ok {
        return false, errors.New("supplier_has_contract: arg must be string")
    }
    return supplierContractRepo.HasActive(ctx, supplierID)
}
evalCtx.RegisterFunction("supplier_has_contract", supplierContractFn)
```

#### JSON Deserialisation of Stored Rules

`ConditionGroup.Children` is `[]any`, which requires a custom unmarshaller when loading from PostgreSQL. The standard decoder cannot distinguish a `ConditionGroup` child from a `ConditionRule` child without inspecting for the `"conjunction"` key.

```go
// awo/pkg/condition/unmarshal.go

func (g *ConditionGroup) UnmarshalJSON(data []byte) error {
    type Alias ConditionGroup
    var raw struct {
        Alias
        Children []json.RawMessage `json:"children"`
    }
    if err := json.Unmarshal(data, &raw); err != nil {
        return err
    }
    *g = ConditionGroup(raw.Alias)
    g.Children = make([]any, 0, len(raw.Children))

    for _, child := range raw.Children {
        var probe struct {
            Conjunction string `json:"conjunction"`
        }
        _ = json.Unmarshal(child, &probe)

        if probe.Conjunction != "" {
            var group ConditionGroup
            if err := json.Unmarshal(child, &group); err != nil {
                return fmt.Errorf("child group: %w", err)
            }
            g.Children = append(g.Children, &group)
        } else {
            var rule ConditionRule
            if err := json.Unmarshal(child, &rule); err != nil {
                return fmt.Errorf("child rule: %w", err)
            }
            g.Children = append(g.Children, &rule)
        }
    }
    return nil
}
```

After this unmarshaller runs, the tree is fully typed before it reaches the evaluator's type switch — avoiding the `mapToStruct` JSON round-trip on the hot evaluation path.

#### Evaluation Metrics

Every `EvalContext` accumulates metrics for the pass. `rules.RuleEngine` reads these after each evaluation to populate `RuleExecution` audit records.

```go
type EvaluationMetrics struct {
    RulesEvaluated  int32
    GroupsEvaluated int32
    Duration        time.Duration
    CacheHits       int32   // compiled regex cache hits
    CacheMisses     int32
    Errors          int32
}

metrics := evalCtx.GetMetrics() // thread-safe snapshot
```

---

### Package: `awo/platform/rules`

**UI Label:** Settings → Business Rules
**Purpose:** Persistence, tenant scoping, trigger registry, action execution, approval chain management. Delegates all condition evaluation to `pkg/condition`. This is the only `awo/*` package that imports `pkg/condition`.

#### Database Models

```
RuleSet       { ID, TenantID, Name, Module, TriggerEvent, Active,
                Priority int, Description, CreatedBy, UpdatedAt }

Rule          { ID, RuleSetID, Name, Sequence int, StopOnMatch bool,
                ConditionTree JSONB,   ← serialised condition.ConditionGroup
                Actions []RuleAction }

RuleAction    { Type, Parameters JSONB }
// Action Types:
//   approve               → auto-approve the pending item
//   reject                → reject with a configurable message
//   require_approval_from → add a step to the approval chain
//   notify                → send notification to role/user/list
//   set_field             → mutate a field on the record
//   assign_to             → assign record to a user or team
//   add_tag               → tag the record
//   trigger_webhook       → call a registered external endpoint
//   run_workflow          → start a named saga

RuleExecution { ID, TenantID, RuleSetID, RuleID, RecordType, RecordID,
                FactSnapshot JSONB, EvalMetrics JSONB,
                ActionsExecuted []string, OccurredAt }

TriggerField  { TriggerEvent, Key, Label, FieldType, AllowedValues []string? }
// Populated by modules at startup via RuleRegistry.RegisterTrigger
```

#### `rules.RuleEngine`

```go
// Core evaluation entry point. Loads active rule sets, builds condition.EvalContext
// from facts, calls condition.Evaluator.Evaluate per rule, executes matched actions,
// records execution audit, and returns the aggregated EvaluationResult.
RuleEngine.Evaluate(ctx, tenantID uuid.UUID, triggerEvent string, facts FactContext) (*EvaluationResult, error)

// Management
RuleEngine.CreateRuleSet(ctx, tenantID uuid.UUID, input RuleSetInput) (*RuleSet, error)
RuleEngine.AddRule(ctx, ruleSetID uuid.UUID, input RuleInput) (*Rule, error)
RuleEngine.UpdateRule(ctx, ruleID uuid.UUID, input RuleInput) error
RuleEngine.DeleteRule(ctx, ruleID uuid.UUID) error
RuleEngine.ReorderRules(ctx, ruleSetID uuid.UUID, orderedIDs []uuid.UUID) error
RuleEngine.ActivateRuleSet(ctx, ruleSetID uuid.UUID) error
RuleEngine.DeactivateRuleSet(ctx, ruleSetID uuid.UUID) error

// Test — runs against sample facts without persisting anything (used by Settings UI "Test Rule" panel)
RuleEngine.TestRuleSet(ctx, ruleSetID uuid.UUID, sampleFacts FactContext) (*EvaluationResult, error)

// History
RuleEngine.GetExecutionHistory(ctx, ruleSetID uuid.UUID, filter HistoryFilter) ([]*RuleExecution, error)
```

#### `EvaluationResult`

```go
type Decision string
const (
    DecisionApprove         Decision = "approve"
    DecisionReject          Decision = "reject"
    DecisionRequireApproval Decision = "require_approval"
    DecisionNone            Decision = "none" // no rule matched
)

type EvaluationResult struct {
    Decision          Decision
    RejectReason      string
    RequiredApprovers []ApprovalChainStep
    Notifications     []PendingNotification
    FieldMutations    []FieldMutation
    Tags              []string
    MatchedRuleIDs    []uuid.UUID
    Metrics           condition.EvaluationMetrics
}
```

#### Internal: How `RuleEngine.Evaluate` Calls `pkg/condition`

```go
// awo/platform/rules/engine.go

func (e *RuleEngine) Evaluate(
    ctx context.Context,
    tenantID uuid.UUID,
    triggerEvent string,
    facts FactContext,
) (*EvaluationResult, error) {

    ruleSets, err := e.repo.LoadActive(ctx, tenantID, triggerEvent)
    if err != nil {
        return nil, fmt.Errorf("load rule sets: %w", err)
    }

    result := &EvaluationResult{}
    opts := condition.DefaultEvalOptions() // MaxDepth=10, MaxConditions=1000, Timeout=30s

    for _, rs := range ruleSets {
        for _, rule := range rs.Rules {

            // Fresh EvalContext per rule — isolated, never shared between rules
            evalCtx := condition.NewEvalContext(map[string]any(facts), opts)

            // Register tenant-scoped custom functions for this trigger
            for name, fn := range e.registry.FunctionsFor(tenantID, triggerEvent) {
                _ = evalCtx.RegisterFunction(name, fn)
            }

            // Deserialise ConditionTree — UnmarshalJSON produces a fully typed tree
            var root condition.ConditionGroup
            if err := json.Unmarshal(rule.ConditionTree, &root); err != nil {
                return nil, fmt.Errorf("rule %s: invalid condition tree: %w", rule.ID, err)
            }

            // Delegate evaluation entirely to pkg/condition
            matched, err := e.evaluator.Evaluate(ctx, &root, evalCtx)
            if err != nil {
                if errors.Is(err, condition.ErrEvaluationTimeout) ||
                    errors.Is(err, condition.ErrResourceLimitExceeded) {
                    return nil, fmt.Errorf("rule %s: %w", rule.ID, err)
                }
                slog.Warn("rule evaluation error", "rule_id", rule.ID, "err", err)
                continue
            }

            if !matched {
                continue
            }

            for _, action := range rule.Actions {
                e.applyAction(ctx, action, facts, result)
            }

            result.MatchedRuleIDs = append(result.MatchedRuleIDs, rule.ID)
            result.Metrics = evalCtx.GetMetrics()

            // Persist execution audit record asynchronously
            go e.recordExecution(context.Background(), rs, rule, facts, evalCtx.GetMetrics())

            if rule.StopOnMatch {
                return result, nil
            }
        }
    }

    if result.Decision == "" {
        result.Decision = DecisionNone
    }
    return result, nil
}
```

#### Trigger & Field Registry

Modules register their available fact fields at startup. This registry drives the UI field picker and validates rules on save.

```go
// awo/platform/rules/registry.go

func (r *RuleRegistry) RegisterTrigger(event string, fields []condition.Field)
func (r *RuleRegistry) RegisterFunction(tenantID uuid.UUID, name string, fn condition.FuncHandler)
func (r *RuleRegistry) FieldsFor(event string) []condition.Field
func (r *RuleRegistry) FunctionsFor(tenantID uuid.UUID, event string) map[string]condition.FuncHandler

// Called at startup in cmd/server/wire.go:
registry.RegisterTrigger("procurement.POSubmittedForApproval", []condition.Field{
    {Name: "po.total",             Label: "PO Total Amount",   Type: condition.FieldTypeNumber},
    {Name: "po.currency",          Label: "Currency",          Type: condition.FieldTypeSelect,
     Values: []condition.SelectOption{{Label: "USD", Value: "USD"}, {Label: "KES", Value: "KES"}}},
    {Name: "po.category",          Label: "Purchase Category", Type: condition.FieldTypeText},
    {Name: "po.supplier_tier",     Label: "Supplier Tier",     Type: condition.FieldTypeSelect,
     Values: []condition.SelectOption{
         {Label: "Preferred", Value: "preferred"},
         {Label: "Approved",  Value: "approved"},
         {Label: "New",       Value: "new"},
     }},
    {Name: "requester.role",       Label: "Requester Role",       Type: condition.FieldTypeText},
    {Name: "requester.department", Label: "Requester Department", Type: condition.FieldTypeText},
})

registry.RegisterTrigger("attendance.LeaveRequestSubmitted", []condition.Field{
    {Name: "leave.days",              Label: "Number of Days",    Type: condition.FieldTypeNumber},
    {Name: "leave.type",              Label: "Leave Type",        Type: condition.FieldTypeSelect},
    {Name: "leave.balance_remaining", Label: "Remaining Balance", Type: condition.FieldTypeNumber},
    {Name: "employee.department",     Label: "Department",        Type: condition.FieldTypeText},
    {Name: "employee.grade",          Label: "Grade",             Type: condition.FieldTypeText},
})

registry.RegisterTrigger("receivables.InvoiceAboutToPost", []condition.Field{
    {Name: "invoice.total",                Label: "Invoice Total",         Type: condition.FieldTypeNumber},
    {Name: "invoice.currency",             Label: "Currency",              Type: condition.FieldTypeSelect},
    {Name: "customer.credit_limit",        Label: "Credit Limit",          Type: condition.FieldTypeNumber},
    {Name: "customer.outstanding_balance", Label: "Outstanding Balance",   Type: condition.FieldTypeNumber},
    {Name: "customer.payment_terms",       Label: "Payment Terms",         Type: condition.FieldTypeText},
})
```

---

### Approval Workflows

When a `RuleAction` of type `require_approval_from` is executed, `ApprovalService` creates a multi-step chain. Steps come directly from the matched rule actions — no separate approval policy model needed.

#### Models

```
ApprovalRequest { ID, TenantID, RecordType, RecordID, RuleSetID?,
                  Status (pending|approved|rejected|cancelled|expired),
                  Steps []ApprovalStep, CurrentStep int,
                  RequestedBy, RequestedAt, CompletedAt?, ExpiresAt? }

ApprovalStep    { Sequence, ApproverType (user|role|direct_manager|department_head),
                  ApproverID?, RoleSlug?, Decision (pending|approved|rejected),
                  DecidedBy?, DecidedAt?, Comment? }
```

#### `approval.ApprovalService`

```go
ApprovalService.RequestApproval(ctx, tenantID uuid.UUID, recordType string, recordID uuid.UUID, requestedBy uuid.UUID, steps []ApprovalStep) (*ApprovalRequest, error)
ApprovalService.Approve(ctx, requestID, approverID uuid.UUID, comment string) error
ApprovalService.Reject(ctx, requestID, approverID uuid.UUID, reason string) error
ApprovalService.Escalate(ctx, requestID uuid.UUID) error
ApprovalService.GetPendingFor(ctx, userID uuid.UUID) ([]*ApprovalRequest, error)
ApprovalService.GetHistory(ctx, recordType string, recordID uuid.UUID) ([]*ApprovalRequest, error)
ApprovalService.CancelRequest(ctx, requestID uuid.UUID) error
```

---

### `FactContext` — The Rule Input

```go
// awo/platform/rules/facts.go
type FactContext map[string]any

// Maps directly to condition.EvalContext.Data.
// Dot-notation keys are traversed by condition.EvalContext.GetValue.
// The caller assembles all facts the rule engine might need — the engine never fetches data.

facts := rules.FactContext{
    "po.total":             po.Total,            // float64
    "po.currency":          po.CurrencyCode,     // string
    "po.category":          po.Category,         // string
    "po.supplier_tier":     supplier.Tier,       // string
    "requester.role":       requester.Role,      // string
    "requester.department": requester.Department,// string
}
```

---

### Tenant-Configured Rule Examples

The following show both what the tenant admin configures (human-readable) and the exact `condition.ConditionGroup` JSON stored in `Rule.ConditionTree`.

#### Purchase Order Approval

**Admin configures:**
```
RuleSet: "Purchase Order Approval"
Trigger: procurement.POSubmittedForApproval

Rule 1 — "Auto-approve small preferred supplier orders"
  Conditions (AND):
    po.total         less_than   500
    po.supplier_tier equal       "preferred"
  Actions: approve
  StopOnMatch: true

Rule 2 — "Manager approval for mid-range"
  Conditions (AND):
    po.total  between  500  and  9999.99
  Actions: require_approval_from { type: direct_manager }
  StopOnMatch: true

Rule 3 — "CFO + Manager for large orders"
  Conditions (AND):
    po.total  greater_or_equal  10000
  Actions:
    require_approval_from { type: direct_manager }
    require_approval_from { type: role, slug: cfo }
    notify { role: finance_team, template: large_po_alert }
  StopOnMatch: true

Rule 4 — "Block CapEx from non-department-heads"
  Conditions (AND):
    po.category    equal      "Capital Expenditure"
    requester.role not_equal  "department_head"
  Actions: reject { message: "CapEx orders must be submitted by a department head" }
  StopOnMatch: true
```

**Stored `ConditionTree` JSON for Rule 3:**
```json
{
  "id": "grp-po-large",
  "conjunction": "and",
  "children": [
    {
      "id": "rule-po-total-gte",
      "left":  { "type": "field", "field": "po.total" },
      "op":    "greater_or_equal",
      "right": { "type": "value", "value": 10000 }
    }
  ]
}
```

**Credit limit rule using Formula override** (cross-field arithmetic — not possible with basic operators):
```json
{
  "id": "grp-credit-check",
  "conjunction": "and",
  "if": "customer.outstanding_balance + invoice.total > customer.credit_limit && customer.credit_limit > 0"
}
```

The `if` formula is compiled by `expr-lang` once and cached as a `*vm.Program` in `Evaluator.programCache`. Subsequent evaluations reuse the compiled program — zero recompilation overhead.

#### Leave Request Rules

```
RuleSet: "Leave Approval"
Trigger: attendance.LeaveRequestSubmitted

Rule 1 — "Auto-approve single day (non-annual) with balance"
  Conditions (AND):
    leave.days              equal            1
    leave.balance_remaining greater_or_equal 1
    leave.type              not_equal        "annual"
  Actions: approve
  StopOnMatch: true

Rule 2 — "Manager for up to 5 days"
  Conditions (AND):
    leave.days  less_or_equal  5
  Actions: require_approval_from { type: direct_manager }
  StopOnMatch: true

Rule 3 — "HR for extended leave"
  Conditions (AND):
    leave.days  greater  5
  Actions:
    require_approval_from { type: direct_manager }
    require_approval_from { type: role, slug: hr_manager }
  StopOnMatch: true
```

#### Sales Quote Rules

```
RuleSet: "Sales Quote Approval"
Trigger: pipeline.QuoteSubmittedForApproval

Rule 1 — "Auto-approve standard quotes"
  Conditions (AND):
    quote.discount_pct  less_or_equal  10
    quote.total         less_or_equal  5000
  Actions: approve
  StopOnMatch: true

Rule 2 — "Sales manager for high discounts"
  Conditions (AND):
    quote.discount_pct  greater  10
  Actions:
    require_approval_from { type: role, slug: sales_manager }
    notify { user: requester_manager, template: high_discount_quote }
  StopOnMatch: true

Rule 3 — "Director for enterprise deals"
  Conditions (AND):
    quote.total  greater  50000
  Actions:
    require_approval_from { type: role, slug: sales_manager }
    require_approval_from { type: role, slug: sales_director }
  StopOnMatch: true
```

#### Invoice Credit Limit Rules

```
RuleSet: "Credit Limit Check"
Trigger: receivables.InvoiceAboutToPost

Rule 1 — "Block if over credit limit"  [uses formula for cross-field arithmetic]
  Formula: customer.outstanding_balance + invoice.total > customer.credit_limit && customer.credit_limit > 0
  Actions:
    reject { message: "Customer credit limit exceeded. Finance approval required." }
    notify { role: credit_controller, template: credit_limit_breach }
  StopOnMatch: true

Rule 2 — "Warn if approaching limit"  [formula: > 90% of limit]
  Formula: customer.outstanding_balance + invoice.total > customer.credit_limit * 0.9
  Actions:
    notify { role: account_manager, template: credit_limit_warning }
    add_tag { tag: "approaching_credit_limit" }
```

---

### How Business Services Call the Rule Engine

The `DecisionEngine` port interface is defined in the consuming module — services never import `platform/rules` or `pkg/condition` directly.

```go
// awo/core/procurement/port.go
// Defined BY procurement. Implemented BY rules.RuleEngine (adapter in platform/rules).
type DecisionEngine interface {
    Evaluate(ctx context.Context, tenantID uuid.UUID, trigger string, facts map[string]any) (*DecisionResult, error)
}

// awo/core/procurement/purchase_order_service.go

func (s *PurchaseOrderService) SubmitForApproval(ctx context.Context, poID uuid.UUID) error {
    po, err := s.repo.GetPO(ctx, poID)
    if err != nil { return err }

    supplier, _ := s.supplierRepo.Get(ctx, po.SupplierID)
    requester, _ := s.employeeReader.GetRequester(ctx, po.CreatedBy)

    // 1. Assemble the fact context — all facts the rule engine might need
    facts := map[string]any{
        "po.total":              po.Total,
        "po.currency":           po.CurrencyCode,
        "po.category":           po.Category,
        "po.supplier_tier":      supplier.Tier,
        "requester.role":        requester.Role,
        "requester.department":  requester.Department,
        "requester.employee_id": requester.EmployeeID.String(),
    }

    // 2. Evaluate — service never touches condition.* or rules.* directly
    decision, err := s.decisions.Evaluate(ctx, po.TenantID, "procurement.POSubmittedForApproval", facts)
    if err != nil {
        return fmt.Errorf("rule evaluation: %w", err)
    }

    // 3. Act on the decision
    switch decision.Decision {
    case "approve":
        return s.confirmPO(ctx, poID)

    case "reject":
        return s.rejectPO(ctx, poID, decision.RejectReason)

    case "require_approval":
        _, err = s.approval.RequestApproval(
            ctx, po.TenantID, "purchase_order", poID,
            po.CreatedBy, decision.RequiredApprovers,
        )
        if err != nil { return err }

        s.bus.Publish(ctx, eventbus.Event{
            Type:     "procurement.POPendingApproval",
            TenantID: po.TenantID,
            Payload:  map[string]any{"po_id": poID, "approvers": decision.RequiredApprovers},
        })
        return nil

    default: // "none" — no rule matched; apply module default behaviour
        return s.confirmPO(ctx, poID)
    }
}
```

---

### Rules UI for Tenant Administrators

The rules builder is exposed in Settings for each module that has registered triggers.

```
Settings → Accounting       → Approval Rules
Settings → Purchasing       → Purchase Order Rules
Settings → Sales            → Quote Approval Rules
Settings → People & HR      → Leave Approval Rules
Settings → Support          → Ticket Assignment Rules
Settings → Production       → Quality Hold Rules
```

Each screen provides:

- **Rule set list** with drag-to-reorder priority (higher priority = evaluated first)
- **Condition builder:** field picker (from `RuleRegistry.FieldsFor`) → operator picker (filtered by `condition.FieldType` using `condition.Config.Types`) → value input (rendered as text, number, date picker, or select based on `condition.Field.Type` and `Values`)
- **AND/OR group nesting** up to 3 levels deep in the UI (evaluator supports 10)
- **Formula mode** toggle — exposes an expr-lang text input for cross-field arithmetic or cases the operator set cannot express
- **Action builder:** action type selector → parameter form
- **Test panel:** enter sample values matching the trigger's registered fields → see which rules fire, which actions would execute, the `EvaluationMetrics` (duration, conditions evaluated) — no live data touched. Calls `RuleEngine.TestRuleSet`.
- **Execution history:** paginated log showing `FactSnapshot`, matched rule, actions taken, `EvalMetrics.Duration`, and `EvalMetrics.RulesEvaluated` per past evaluation

The operator list per field type maps to `condition.Config.Types`:

| Field type | Available operators in UI |
|---|---|
| `number` | equal, not_equal, less, less_or_equal, greater, greater_or_equal, between, not_between, is_empty, is_not_empty |
| `text` | equal, not_equal, contains, not_contains, starts_with, ends_with, is_empty, is_not_empty, match_regexp |
| `select` | equal, not_equal, select_any_in, select_not_any_in, is_empty, is_not_empty |
| `date` / `datetime` | equal, not_equal, less, less_or_equal, greater, greater_or_equal, between, is_empty, is_not_empty |
| `boolean` | equal, not_equal |

---

### Decision Flow — End to End

```
Tenant admin configures in Settings → Purchasing → Purchase Order Rules:
  Rule 3: po.total >= 10000 → require CFO + manager approval
        │
        │  stored: Rule { ConditionTree: {conjunction:"and", children:[{left:{field:"po.total"},op:"greater_or_equal",right:{value:10000}}]},
        │                 Actions: [{type:"require_approval_from",params:{type:"role",slug:"cfo"}}, ...] }
        ▼
Employee creates a $15,000 Purchase Order → clicks "Submit for Approval"
        │
        ▼  POST /purchasing/orders/:id/submit
        │
        ▼  middleware: ResolveTenant → Authenticate → InjectFlags → RequirePermission("submit","purchase_orders")
        │
        ▼  PurchaseOrderService.SubmitForApproval(ctx, poID)
        │
        ▼  assembles FactContext { "po.total": 15000, "requester.role": "Engineer", ... }
        │
        ▼  DecisionEngine.Evaluate → rules.RuleEngine.Evaluate(ctx, tenantID, event, facts)
        │
        ▼  loads 4 active rules, iterates in priority order:
           Rule 1: condition.Evaluator.Evaluate({po.total < 500 AND supplier_tier = "preferred"})
                   GetValue("po.total") = 15000 → compare(15000, 500) → OpLess = false → NO MATCH
           Rule 2: condition.Evaluator.Evaluate({po.total BETWEEN 500 AND 9999.99})
                   compare(15000, 9999.99) → OpBetween upper = false → NO MATCH
           Rule 3: condition.Evaluator.Evaluate({po.total >= 10000})
                   compare(15000, 10000) → OpGreaterOrEqual = true → MATCH  ← StopOnMatch
                   Actions executed:
                     require_approval_from { direct_manager }  → step 1
                     require_approval_from { role: cfo }       → step 2
                     notify { finance_team, large_po_alert }   → notification queued
           EvalMetrics: { RulesEvaluated: 3, GroupsEvaluated: 3, Duration: ~0.8ms, CacheHits: 0 }
        │
        ▼  ApprovalService.RequestApproval(ctx, ..., steps=[manager_step, cfo_step])
           → ApprovalRequest created, status: pending, currentStep: 0
        │
        ▼  event published: procurement.POPendingApproval
        │
        ▼  notify module → sends email to manager + CFO
        │
        ▼  PO status → "pending_approval"
        │
        ▼  Manager approves → step 0 done, currentStep → 1
        ▼  CFO approves    → step 1 done, all steps complete
        │
        ▼  event published: procurement.POApproved
        ▼  PurchaseOrderService.confirmPO → PO status → "confirmed"
        │
        ▼  RuleExecution audit record:
           { FactSnapshot: {...}, EvalMetrics: {RulesEvaluated:3, Duration:0.8ms},
             ActionsExecuted: ["require_approval_from", "require_approval_from", "notify"] }
```

The tenant admin can change the threshold, add steps, or reorder rules entirely in the Settings UI. The next evaluation uses the updated rules immediately — zero code changes, zero redeployment.

---

### Architecture Position of `pkg/condition`

```
awo/
├── pkg/
│   └── condition/          ← Standalone evaluator. Zero ERP concepts. Zero awo/* imports.
│       ├── evaluator.go    ← Evaluator, Evaluate(), evaluateGroup(), evaluateRule()
│       ├── builder.go      ← Builder fluent API
│       ├── context.go      ← EvalContext, GetValue(), RegisterFunction()
│       ├── operators.go    ← applyOperator(), compare(), matchRegexp()
│       ├── unmarshal.go    ← ConditionGroup.UnmarshalJSON() — typed child deserialisation
│       ├── cache.go        ← regexCache (LRU bounded), programCache (sync.Map)
│       └── types.go        ← All exported types and constants
│
└── platform/
    └── rules/              ← ERP orchestration. Only awo/* package that imports pkg/condition.
        ├── engine.go       ← RuleEngine.Evaluate() — loads DB, calls condition.Evaluator
        ├── registry.go     ← RuleRegistry — trigger→fields, tenant→functions
        ├── approval.go     ← ApprovalService
        ├── actions.go      ← applyAction() — interprets RuleAction types
        └── models.go       ← RuleSet, Rule, RuleAction, RuleExecution, EvaluationResult
```

**Import summary for this sub-graph:**
- `pkg/condition` has **no** `awo/*` imports — it is a pure standalone library
- `platform/rules` is the **only** `awo/*` package that imports `pkg/condition`
- All `core/*` business modules reach the rule engine through the `DecisionEngine` port interface — they never import `platform/rules` or `pkg/condition` directly

The full formal rules governing every import boundary in the codebase are specified in the next section.

---

## Import Constraints

### Purpose

Import constraints are the structural backbone of Awo's modularity. They are not conventions — they are **enforced rules**. A constraint violation is a build failure, not a code review comment. This section is the single authoritative reference for what any given package is and is not allowed to import.

---

### The Full Import Allowance Matrix

The matrix below covers every layer in the codebase. Read each row as: *"This package layer may import the columns marked ✓."*

```
                    ┌─ pkg/ ─┬─ platform/ ────────────────────────────────┬─ core/ ──┬─ analytics/ ─┬─ sagas/ ─┬─ web/ ──┬─ cmd/ ─┐
                    │condition│ iam tenant audit notify bridge rules flags  │  all     │    all       │  all     │  all   │  all   │
────────────────────┼─────────┼───────────────────────────────────────────┼──────────┼──────────────┼──────────┼────────┼────────┤
pkg/condition       │    —    │  ✗    ✗     ✗     ✗      ✗      ✗     ✗   │    ✗     │      ✗       │    ✗     │   ✗    │   ✗    │
platform/iam        │    ✗    │  —    ✗     ✗     ✗      ✗      ✗     ✗   │    ✗     │      ✗       │    ✗     │   ✗    │   ✗    │
platform/tenant     │    ✗    │  ✓    —     ✗     ✗      ✗      ✗     ✗   │    ✗     │      ✗       │    ✗     │   ✗    │   ✗    │
platform/audit      │    ✗    │  ✓    ✓     —     ✗      ✗      ✗     ✗   │    ✗     │      ✗       │    ✗     │   ✗    │   ✗    │
platform/notify     │    ✗    │  ✓    ✓     ✗     —      ✗      ✗     ✗   │    ✗     │      ✗       │    ✗     │   ✗    │   ✗    │
platform/bridge     │    ✗    │  ✓    ✓     ✓     ✗      —      ✗     ✗   │    ✗     │      ✗       │    ✗     │   ✗    │   ✗    │
platform/config     │    ✗    │  ✓    ✓     ✗     ✗      ✗      —     ✗   │    ✗     │      ✗       │    ✗     │   ✗    │   ✗    │
platform/flags      │    ✓    │  ✓    ✓     ✗     ✗      ✗      ✓     —   │    ✗     │      ✗       │    ✗     │   ✗    │   ✗    │
platform/rules      │    ✓    │  ✓    ✓     ✓     ✓      ✓      ✓     ✓   │    ✗     │      ✗       │    ✗     │   ✗    │   ✗    │
core/* (all)        │    ✗    │  ✓    ✓     ✓     ✓      ✗      ✗     ✓   │   PORT   │      ✗       │    ✗     │   ✗    │   ✗    │
analytics/*         │    ✗    │  ✓    ✓     ✗     ✗      ✗      ✗     ✗   │   READ   │      —       │    ✗     │   ✗    │   ✗    │
sagas/*             │    ✗    │  ✓    ✓     ✓     ✓      ✗      ✓     ✓   │    ✓     │      ✗       │    —     │   ✗    │   ✗    │
web/middleware      │    ✗    │  ✓    ✓     ✓     ✓      ✓      ✗     ✓   │    ✗     │      ✗       │    ✗     │   —    │   ✗    │
web/handlers        │    ✗    │  ✓    ✓     ✗     ✗      ✗      ✗     ✓   │    ✓     │      ✗       │    ✗     │   ✓    │   ✗    │
web/router          │    ✗    │  ✗    ✗     ✗     ✗      ✗      ✗     ✗   │    ✗     │      ✗       │    ✗     │   ✓    │   ✗    │
cmd/server          │    ✗    │  ✓    ✓     ✓     ✓      ✓      ✓     ✓   │    ✓     │      ✓       │    ✓     │   ✓    │   —    │

Legend
  ✓     Direct import allowed
  PORT  Via port interface only — no direct package import
  READ  Read-only DB queries or data mart — no service method calls
  ✗     Never import — hard violation
  —     Self / not applicable
```

**Key asymmetries to call out explicitly:**

- `platform/rules` may import `pkg/condition` and all other `platform/*` packages. No other `awo/*` package may import `pkg/condition`.
- `platform/flags` may import `pkg/condition` (for condition-based rollout targeting) and `platform/config` (to read flag overrides). It may not import `platform/rules`.
- `core/*` packages may import `platform/config` and `platform/flags` (as Reader interfaces). They may not import `platform/rules` directly — they use the `DecisionEngine` port.
- `web/handlers` may import `core/*` service types directly (they need to call service methods). They may not import `platform/rules`, `sagas/*`, or `analytics/*`.
- `sagas/*` may import `core/*` directly but may not be imported by `core/*` under any circumstances.
- `analytics/*` may only read data — it imports no service packages from `core/*`, only shared model types and the database pool.
- `cmd/server` is the single integration point. It is the only package permitted to import across all layers simultaneously.

---

### Platform Internal Import Order

Within `platform/`, packages have a strict partial order. A package may only import packages that appear *above* it in this list:

```
1. platform/iam        — depends on nothing inside platform/
2. platform/tenant     — may import: iam
3. platform/audit      — may import: iam, tenant
4. platform/notify     — may import: iam, tenant
5. platform/bridge     — may import: iam, tenant, audit
6. platform/config     — may import: iam, tenant
7. platform/flags      — may import: iam, tenant, config  (+pkg/condition for targeting)
8. platform/rules      — may import: all of the above     (+pkg/condition)
```

`platform/rules` sits at the top of the platform hierarchy deliberately — it depends on all other platform packages but nothing depends on it except `sagas/*`, `web/handlers`, and `cmd/server`.

---

### The `DecisionEngine` Port

`core/*` packages that reach a rule-evaluation decision point (approval gates, credit checks, assignment logic) do so through a narrow port interface defined in `awo/platform/rules/port.go`. The interface lives in `platform/rules` but business modules receive it as an injected dependency — they import the interface type, not the implementation.

```go
// awo/platform/rules/port.go
// Defined here; injected into core/* via Wire. core/* imports this file's package
// only for the interface type — they never call RuleEngine directly.

package rules

import (
    "context"
    "github.com/google/uuid"
)

// DecisionEngine is the port interface injected into business services.
// It is the only surface of platform/rules that core/* packages may reference.
type DecisionEngine interface {
    Evaluate(
        ctx          context.Context,
        tenantID     uuid.UUID,
        triggerEvent string,
        facts        map[string]any,
    ) (*EvaluationResult, error)
}
```

> **Wait — doesn't importing `platform/rules` for the interface type create a dependency from `core/*` to `platform/rules`?**
>
> Yes. This is the one deliberate exception in the matrix. The dependency is one-way and narrow: `core/*` imports `platform/rules` only for the `DecisionEngine` interface and `EvaluationResult` type. It never calls `RuleEngine` directly. The `platform/rules` package imports nothing from `core/*` in return, so no cycle exists. This is preferable to duplicating the interface definition in every business module.
>
> If this still feels too tight, the interface can be moved to a dedicated `awo/platform/rules/port` sub-package that contains only the interface and result types — no implementation, no DB access, no condition engine import.

---

### `pkg/condition` Isolation Contract

`pkg/condition` must remain completely free of `awo/*` imports at all times. This is enforced by a `depguard` linter rule (see below) and is the property that makes the package safe to extract as a standalone open-source library or to use in other Go projects.

```
pkg/condition is allowed to import:
  Standard library                (context, encoding/json, fmt, reflect, strings, sync, time, ...)
  github.com/dlclark/regexp2      (ReDoS-safe regex)
  github.com/expr-lang/expr       (formula compilation and evaluation)
  github.com/google/uuid          (Builder.AddRule generates rule IDs)

pkg/condition is NEVER allowed to import:
  github.com/mustafe/awo/*        (any Awo package whatsoever)
  github.com/gofiber/fiber        (HTTP framework)
  github.com/jackc/pgx            (database driver)
  Any other infrastructure package not listed above
```

---

### `web/*` — No Business Logic Rule

All packages under `web/` have a single non-negotiable constraint: **no business logic**. Concretely:

```
web/* MUST NOT:
  Contain conditional branches that implement business decisions
  Call two or more service methods and combine their results into a business outcome
  Import platform/rules, platform/flags, or pkg/condition directly
  Contain SQL queries of any kind
  Contain domain model structs (these belong in core/* or platform/*)

web/* IS ALLOWED TO:
  Call a single service method and render its result
  Read resolved flags from the request context (set by middleware)
  Read tenant/user from the request context (set by middleware)
  Parse and validate HTTP input before passing it to a service
  Map service errors to HTTP status codes
  Compose templ components for rendering
```

`web/middleware` is the one sub-package permitted to import `platform/iam`, `platform/tenant`, `platform/flags`, and `platform/audit` because middleware is infrastructure, not business logic. It enforces the boundary; it does not cross it.

---

### `cmd/server` — The Wiring Root

`cmd/server` is the **only** package in the codebase that sees the whole dependency graph. Its single job is to instantiate every service, wire interfaces to their implementations, and register event bus subscriptions. It contains no logic of its own.

```
cmd/server IS ALLOWED TO import:
  Every awo/* package

cmd/server MUST NOT:
  Contain any function with business logic
  Be imported by any other package (it is a leaf node in the import graph)
  Contain service method calls beyond constructor calls and bus.Subscribe()
```

Because `cmd/server` imports everything, it is also where import cycles surface first. If adding a dependency in a business module would create a cycle, the Go compiler refuses to build `cmd/server` — the violation is caught immediately, not at code review.

---

### Enforcement

#### 1. `depguard` Linter Configuration

```yaml
# .golangci.yml (excerpt)

linters-settings:
  depguard:
    rules:

      # pkg/condition must never import awo packages
      condition-isolation:
        files:
          - "pkg/condition/**/*.go"
        deny:
          - pkg: "github.com/mustafe/awo"
            desc: "pkg/condition must remain import-free of awo/* packages"
          - pkg: "github.com/gofiber/fiber"
            desc: "pkg/condition must not depend on HTTP framework"
          - pkg: "github.com/jackc/pgx"
            desc: "pkg/condition must not depend on database driver"

      # core/* must not import each other directly
      core-isolation:
        files:
          - "core/**/*.go"
        deny:
          - pkg: "github.com/mustafe/awo/pkg/condition"
            desc: "core/* must not import pkg/condition directly — use platform/rules.DecisionEngine port"

      # web/* must not contain business logic imports
      web-no-rules:
        files:
          - "web/handlers/**/*.go"
          - "web/router/**/*.go"
        deny:
          - pkg: "github.com/mustafe/awo/platform/rules"
            desc: "web handlers must not import rules engine — receive DecisionEngine via middleware context"
          - pkg: "github.com/mustafe/awo/pkg/condition"
            desc: "web handlers must not import condition engine"
          - pkg: "github.com/mustafe/awo/sagas"
            desc: "web handlers must not import sagas — sagas are triggered by service events"

      # analytics must never call service methods
      analytics-read-only:
        files:
          - "analytics/**/*.go"
        deny:
          - pkg: "github.com/mustafe/awo/core"
            desc: "analytics/* must not import core/* service packages — use read-only DB queries"
          - pkg: "github.com/mustafe/awo/sagas"
            desc: "analytics/* must not import sagas"

      # sagas must not be imported by core/*
      sagas-not-imported-by-core:
        files:
          - "core/**/*.go"
        deny:
          - pkg: "github.com/mustafe/awo/sagas"
            desc: "core/* must not import sagas — sagas orchestrate core, not the other way around"

linters:
  enable:
    - depguard
```

#### 2. `go-arch-lint` Boundaries (alternative / complementary)

For teams preferring a declarative architecture spec file:

```yaml
# .go-arch-lint.yml

version: 1
workdir: .

components:
  pkg_condition:  { in: "pkg/condition/**" }
  platform_iam:   { in: "platform/iam/**" }
  platform_rules: { in: "platform/rules/**" }
  core:           { in: "core/**" }
  analytics:      { in: "analytics/**" }
  sagas:          { in: "sagas/**" }
  web:            { in: "web/**" }
  cmd:            { in: "cmd/**" }

deps:
  pkg_condition:
    mayDependOn: []                         # nothing

  platform_iam:
    mayDependOn: []

  platform_rules:
    mayDependOn:
      - pkg_condition
      - platform_iam
      # ... other platform packages

  core:
    mayDependOn:
      - platform_iam
      - platform_tenant
      - platform_audit
      - platform_notify
      - platform_config
      - platform_flags
      - platform_rules   # interface types only (DecisionEngine port)

  analytics:
    mayDependOn:
      - platform_iam
      - platform_tenant

  sagas:
    mayDependOn:
      - core
      - platform_rules
      - platform_notify

  web:
    mayDependOn:
      - core
      - platform_iam
      - platform_tenant
      - platform_flags
      - platform_audit

  cmd:
    mayDependOn:
      - pkg_condition
      - platform_iam
      - platform_rules
      - core
      - analytics
      - sagas
      - web
```

#### 3. CI Gate

The linter and arch-lint checks run in the same CI step as `go build`. A pull request that introduces a violation in either tool cannot be merged. There is no manual override — the rules are not advisory.

```yaml
# .github/workflows/ci.yml (excerpt)

jobs:
  lint:
    steps:
      - name: Run golangci-lint (includes depguard)
        uses: golangci/golangci-lint-action@v4
        with:
          args: --timeout=5m

      - name: Run go-arch-lint
        run: go run github.com/fe3dback/go-arch-lint@latest check

      - name: Verify no import cycles
        run: go build ./...    # The Go compiler itself catches cycles
```

---

### Violations Reference

The following table catalogues the most likely accidental violations, the error they produce, and the correct fix.

| Violation | Symptom | Correct Fix |
|---|---|---|
| `core/payroll` imports `core/workforce` | `go build` import cycle error | Define `EmployeeReader` port in `payroll`, implement adapter in `workforce` |
| `core/inventory` imports `platform/rules` | `depguard` lint failure | Inject `rules.DecisionEngine` interface via Wire; do not import the package |
| `web/handlers/payroll` imports `pkg/condition` | `depguard` lint failure | Remove — handlers never evaluate conditions; services do |
| `analytics/dashboards` imports `core/ledger` | `depguard` lint failure | Replace with a direct read-model DB query against a reporting view |
| `core/pipeline` imports `sagas/order_to_cash` | `depguard` lint failure | Remove — sagas call pipeline, never the reverse |
| `pkg/condition` imports `awo/platform/audit` | `depguard` lint failure (isolation rule) | Remove — `pkg/condition` must stay import-free; caller records audit entries |
| `platform/notify` imports `platform/rules` | `go-arch-lint` failure | Remove — notify subscribes to events from the bus; it does not call rules |
| Two `platform/*` packages importing each other | `go build` import cycle error | Identify the lower-layer abstraction and extract it; follow the platform import order |

---

### Summary

The import constraint system has four physical enforcement layers:

```
1. Go compiler          → catches import cycles at go build time; no override
2. depguard linter      → enforces forbidden cross-layer imports by file glob + package path
3. go-arch-lint         → validates the declared component dependency graph
4. cmd/server structure → the single wiring point; a cycle anywhere surfaces here first

All four run in CI. A PR cannot merge if any layer reports a violation.
```

The dependency direction always flows in one way: **downward toward the foundation**.

```
cmd/server        (imports everything — wiring only)
    │
    ├── web/*         → HTTP delivery; calls services; no logic
    ├── sagas/*       → multi-step orchestration; calls core/*
    ├── analytics/*   → read-only; queries DB directly
    ├── core/*        → business logic; uses ports for cross-module calls
    │       │
    │       └── platform/*   → infrastructure; no business knowledge
    │               │
    │               └── pkg/condition   → pure evaluation library; zero awo imports
    │
    └── internal/eventbus   → in-process pub/sub; imported by core/* and platform/*
```

Nothing in this diagram ever imports something above itself. That single constraint — rigorously enforced — is what keeps the codebase navigable as it grows.
