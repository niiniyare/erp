# HR & Payroll Module — Awo ERP

**Version:** 2.0.0  
**Status:** Architecture Draft  
**Audience:** Engineering · Product · Finance · Compliance  
**Coverage:** Global — any jurisdiction via configuration

---

## Table of Contents

1. [Overview & Design Philosophy](#1-overview--design-philosophy)
2. [Jurisdiction Configuration Framework](#2-jurisdiction-configuration-framework)
   - 2.1 [Architecture](#21-architecture)
   - 2.2 [Platform-Managed Jurisdiction Registry](#22-platform-managed-jurisdiction-registry)
   - 2.3 [JSON Jurisdiction Schema](#23-json-jurisdiction-schema)
   - 2.4 [Tenant Administrator UI](#24-tenant-administrator-ui)
   - 2.5 [Rate Versioning & Effective Dates](#25-rate-versioning--effective-dates)
3. [Domain Model](#3-domain-model)
4. [Module Architecture](#4-module-architecture)
5. [Sub-Modules](#5-sub-modules)
   - 5.1 [Employee Management](#51-employee-management)
   - 5.2 [Organisation Structure](#52-organisation-structure)
   - 5.3 [Leave & Absence](#53-leave--absence)
   - 5.4 [Time & Attendance](#54-time--attendance)
   - 5.5 [Shift Management](#55-shift-management)
   - 5.6 [Payroll Engine](#56-payroll-engine)
   - 5.7 [Statutory Compliance](#57-statutory-compliance)
   - 5.8 [Benefits & Allowances](#58-benefits--allowances)
   - 5.9 [Employee Accountability](#59-employee-accountability)
   - 5.10 [Offboarding & Separation](#510-offboarding--separation)
6. [Attendance Integration Layer](#6-attendance-integration-layer)
   - 6.1 [Biometric Devices](#61-biometric-devices)
   - 6.2 [Mobile Clock-In/Out](#62-mobile-clock-inout)
   - 6.3 [API Integration](#63-api-integration)
   - 6.4 [Reconciliation & Conflict Resolution](#64-reconciliation--conflict-resolution)
7. [Payroll Pipeline Architecture](#7-payroll-pipeline-architecture)
8. [General Ledger Integration](#8-general-ledger-integration)
   - 8.1 [Design Principles](#81-design-principles)
   - 8.2 [Chart of Accounts Structure](#82-chart-of-accounts-structure)
   - 8.3 [Employee Sub-Ledger](#83-employee-sub-ledger)
   - 8.4 [Payroll Posting Flow](#84-payroll-posting-flow)
   - 8.5 [Statutory Remittance Posting](#85-statutory-remittance-posting)
   - 8.6 [Employee Accountability Posting](#86-employee-accountability-posting)
   - 8.7 [Payroll Clearance Account](#87-payroll-clearance-account)
   - 8.8 [P&L Cleanliness Strategy](#88-pl-cleanliness-strategy)
   - 8.9 [Event Contract with Finance Module](#89-event-contract-with-finance-module)
9. [API Reference](#9-api-reference)
10. [Business Rules & Validation](#10-business-rules--validation)
11. [Authorization Model](#11-authorization-model)
12. [Database Schema](#12-database-schema)
13. [Workflow Orchestration (Temporal)](#13-workflow-orchestration-temporal)
14. [Integration Points](#14-integration-points)
15. [Configuration Flags](#15-configuration-flags)
16. [Future Roadmap Considerations](#16-future-roadmap-considerations)

---

## 1. Overview & Design Philosophy

The HR & Payroll module manages the complete employee lifecycle — from onboarding through separation — together with a fully configurable payroll engine capable of operating in any jurisdiction worldwide. The system ships with pre-built jurisdiction packages for common markets but is not hard-coded to any of them. New jurisdictions are added through configuration, not code.

### Core Design Principles

**Jurisdiction-agnostic computation engine.** The payroll pipeline operates entirely on abstract concepts: earnings, deductions, tax bands, contribution rates. All jurisdiction-specific values are loaded at runtime from a versioned configuration store. Adding a new country requires no code deployment.

**Configuration over code.** Tax bands, statutory contribution rates, minimum wages, statutory leave entitlements, and filing calendars are all stored as structured data. Platform administrators manage them through a dedicated UI or by uploading a JSON jurisdiction package. Tenants customise within the envelope their jurisdiction allows.

**Pipeline-based, deterministic payroll.** Payroll computation is a pure, reproducible Go pipeline. Given the same inputs and the same configuration snapshot, it will always produce the same output. Temporal orchestrates execution at scale; the pipeline itself never touches the database.

**Immutable audit trail.** Payslips, payroll runs, and all statutory rate snapshots are append-only. Corrections require a reversal followed by a new run. GL entries are never edited — only reversed.

**Shift-aware scheduling.** The module natively supports complex shift patterns required by retail, healthcare, hospitality, forecourt, and manufacturing operations, including rotating rosters, split shifts, and on-call scheduling.

**Employee accountability.** Cash and stock discrepancies attributable to specific employees are tracked in an employee sub-ledger, enabling recovery via payroll deduction or separate posting, with clean GL entries that keep the P&L accurate.

**GL integration through events.** The HR module never writes journal entries directly. It emits structured `PayrollPosted` events consumed by the Finance module, which owns the ledger. This hard boundary prevents the HR module from ever putting the accounts into an inconsistent state.

### Scope Boundary

| In Scope | Out of Scope |
|---|---|
| Employee records & employment contracts | Recruitment / ATS |
| Organisation hierarchy | Learning Management System |
| Leave management & approval workflows | Performance appraisals |
| Shift scheduling & roster management | Long-term manpower planning |
| Time tracking & attendance (biometric, mobile, API) | Health & safety incident management |
| Payroll computation & payslip generation | Expense claims (Finance module) |
| Jurisdiction-configurable statutory deductions | General Ledger posting (Finance module) |
| Benefits & allowances administration | Medical insurance claims processing |
| Employee accountability & discrepancy recovery | Asset management (Fixed Assets module) |
| Separation & terminal benefits | Pension fund investment management |

---

## 2. Jurisdiction Configuration Framework

### 2.1 Architecture

The system separates jurisdiction knowledge into three tiers:

```
┌─────────────────────────────────────────────────────────────┐
│  TIER 1 — PLATFORM                                          │
│  Platform administrators manage jurisdiction packages.      │
│  Each package is a versioned JSON file covering all         │
│  statutory rules for one country.                           │
│  Loaded once; shared across all tenants in that country.    │
├─────────────────────────────────────────────────────────────┤
│  TIER 2 — TENANT DEFAULTS                                   │
│  When a tenant selects a jurisdiction, they inherit the     │
│  platform defaults. They may override values that are        │
│  legally permissible to customise (e.g. higher leave        │
│  entitlements than the statutory minimum).                  │
├─────────────────────────────────────────────────────────────┤
│  TIER 3 — EMPLOYEE OVERRIDES                                │
│  Individual employees may have contract-level overrides     │
│  (e.g. a foreign national subject to a different tax        │
│  treaty, or an executive with a bespoke pension rate).      │
└─────────────────────────────────────────────────────────────┘
```

The payroll engine resolves values in priority order: Employee Override → Tenant Default → Platform Jurisdiction Package.

### 2.2 Platform-Managed Jurisdiction Registry

The platform ships with built-in jurisdiction packages for common markets and an open schema for new ones.

```
Platform Jurisdiction Registry
  └── jurisdiction_packages/
        ├── KE.json     (Kenya)
        ├── UG.json     (Uganda)
        ├── TZ.json     (Tanzania)
        ├── NG.json     (Nigeria)
        ├── ZA.json     (South Africa)
        ├── GH.json     (Ghana)
        ├── GB.json     (United Kingdom)
        ├── AE.json     (UAE)
        └── custom/     (platform-admin-uploaded packages)
```

Platform administrators access the Jurisdiction Registry via the Platform Administration console (`/platform/jurisdictions`). From here they can:

- **View** all active and historical jurisdiction packages
- **Upload** a new package (JSON validation runs on upload)
- **Edit** individual fields through a structured form UI (for simple changes like updating a tax band)
- **Activate / Deactivate** a package version
- **Schedule** a package version to become active on a future effective date (e.g. loading next financial year's rates before they take effect)
- **Audit** all changes with author, timestamp, and diff view

### 2.3 JSON Jurisdiction Schema

Each jurisdiction is described by a single JSON document. The schema is validated against a JSON Schema definition at upload time.

```json
{
  "$schema": "https://awoerp.com/schemas/jurisdiction/v2.json",
  "code": "KE",
  "name": "Kenya",
  "currency": "KES",
  "fiscal_year": {
    "start_month": 7,
    "start_day": 1
  },
  "leave_year": {
    "start_month": 1,
    "start_day": 1
  },
  "work_week": {
    "standard_hours_per_week": 45,
    "standard_days": ["Mon","Tue","Wed","Thu","Fri"],
    "overtime_threshold_daily_hours": 8,
    "overtime_threshold_weekly_hours": 45
  },
  "minimum_wage": {
    "effective_from": "2024-05-01",
    "amount": 16200,
    "currency": "KES",
    "period": "monthly",
    "notes": "General unskilled worker — Wages Order 2024"
  },
  "statutory_leave": [
    {
      "code": "ANNUAL",
      "name": "Annual Leave",
      "paid": true,
      "min_days_per_year": 21,
      "accrual_method": "monthly_accrual",
      "max_carry_forward_days": 10,
      "legal_reference": "Employment Act 2007, Section 28"
    },
    {
      "code": "SICK",
      "name": "Sick Leave",
      "paid": true,
      "min_days_per_year": 14,
      "accrual_method": "fixed_annual",
      "legal_reference": "Employment Act 2007, Section 30"
    },
    {
      "code": "MATERNITY",
      "name": "Maternity Leave",
      "paid": true,
      "min_days_per_year": 90,
      "accrual_method": "none",
      "gender_restricted": "female",
      "legal_reference": "Employment Act 2007, Section 29"
    },
    {
      "code": "PATERNITY",
      "name": "Paternity Leave",
      "paid": true,
      "min_days_per_year": 14,
      "accrual_method": "none",
      "gender_restricted": "male",
      "legal_reference": "Employment (Amendment) Act 2022"
    }
  ],
  "tax_components": [
    {
      "code": "PAYE",
      "name": "Pay As You Earn",
      "description": "Income tax withheld by employer",
      "computation_type": "banded_progressive",
      "applies_to": "taxable_income",
      "paid_by": "employee",
      "remittance_frequency": "monthly",
      "remittance_due_day": 9,
      "legal_reference": "Income Tax Act Cap 470",
      "bands": [
        { "from": 0,      "to": 24000,  "rate": 0.10, "flat_tax": 0 },
        { "from": 24001,  "to": 32333,  "rate": 0.25, "flat_tax": 0 },
        { "from": 32334,  "to": 500000, "rate": 0.30, "flat_tax": 0 },
        { "from": 500001, "to": 800000, "rate": 0.325,"flat_tax": 0 },
        { "from": 800001, "to": null,   "rate": 0.35, "flat_tax": 0 }
      ],
      "reliefs": [
        {
          "code": "PERSONAL_RELIEF",
          "name": "Personal Relief",
          "amount": 2400,
          "currency": "KES",
          "period": "monthly",
          "auto_apply": true
        },
        {
          "code": "INSURANCE_RELIEF",
          "name": "Insurance Relief",
          "rate": 0.15,
          "max_amount": 5000,
          "currency": "KES",
          "period": "monthly",
          "auto_apply": false,
          "requires_declaration": true
        }
      ]
    },
    {
      "code": "NSSF_TIER1",
      "name": "NSSF Tier I",
      "computation_type": "percentage_of_ceiling",
      "ceiling_amount": 7000,
      "employee_rate": 0.06,
      "employer_rate": 0.06,
      "paid_by": "both",
      "pre_tax": true,
      "pensionable": true,
      "remittance_frequency": "monthly",
      "remittance_due_day": 9,
      "legal_reference": "NSSF Act 2013"
    },
    {
      "code": "NSSF_TIER2",
      "name": "NSSF Tier II",
      "computation_type": "percentage_of_band",
      "band_floor": 7001,
      "band_ceiling": 36000,
      "employee_rate": 0.06,
      "employer_rate": 0.06,
      "paid_by": "both",
      "pre_tax": true,
      "pensionable": true,
      "remittance_frequency": "monthly",
      "remittance_due_day": 9,
      "legal_reference": "NSSF Act 2013"
    },
    {
      "code": "SHIF",
      "name": "Social Health Insurance Fund",
      "computation_type": "flat_percentage",
      "employee_rate": 0.0275,
      "employer_rate": 0.0,
      "paid_by": "employee",
      "pre_tax": false,
      "remittance_frequency": "monthly",
      "remittance_due_day": 9,
      "legal_reference": "Social Health Insurance Act 2023"
    },
    {
      "code": "AHL",
      "name": "Affordable Housing Levy",
      "computation_type": "flat_percentage",
      "employee_rate": 0.015,
      "employer_rate": 0.015,
      "paid_by": "both",
      "pre_tax": false,
      "remittance_frequency": "monthly",
      "remittance_due_day": 9,
      "legal_reference": "Finance Act 2023"
    },
    {
      "code": "NITA",
      "name": "National Industrial Training Authority",
      "computation_type": "flat_amount",
      "flat_amount_per_employee": 50,
      "currency": "KES",
      "paid_by": "employer",
      "remittance_frequency": "monthly",
      "remittance_due_day": 9
    }
  ],
  "terminal_benefits": {
    "notice_pay_rule": "max(contract_notice_period, 1_month_basic)",
    "leave_encashment_formula": "unused_days * (basic_salary / 30)",
    "gratuity_formula": "(years_of_service * basic_salary) / 12",
    "gratuity_applies_when": "no_registered_pension_scheme",
    "severance_formula": "15_days_basic_per_completed_year",
    "severance_applies_when": "redundancy_only"
  },
  "filing_calendar": [
    {
      "component": "PAYE",
      "form": "P10",
      "frequency": "monthly",
      "due_day_of_next_month": 9
    },
    {
      "component": "PAYE",
      "form": "P9A",
      "frequency": "annual",
      "due_month": 2,
      "due_day": 28
    }
  ],
  "identifier_fields": [
    { "code": "national_id", "name": "National ID", "required": true, "validation_regex": "^[0-9]{7,8}$" },
    { "code": "kra_pin",     "name": "KRA PIN",     "required": true, "validation_regex": "^[A-Z][0-9]{9}[A-Z]$" },
    { "code": "nssf_number", "name": "NSSF Number", "required": false }
  ],
  "effective_from": "2024-07-01",
  "notes": "Finance Act 2023 — SHIF replaces NHIF. AHL introduced July 2023."
}
```

#### Generic Jurisdiction Schema for a New Country

For a country not yet in the registry, a platform administrator fills in a template with no pre-existing knowledge assumed. The minimum viable fields required to process payroll are:

- `currency`, `work_week.standard_hours_per_week`
- At least one `tax_component` with `computation_type` and either `bands` (progressive) or `flat_percentage` (simple)
- `statutory_leave` (can be an empty array if none exist — the tenant may still define their own)
- `minimum_wage` (can be null if none mandated)

### 2.4 Tenant Administrator UI

The tenant-facing settings panel (`/settings/payroll/jurisdiction`) exposes a subset of the jurisdiction configuration that is legally permissible to customise.

**Readable (display only)**

- Active jurisdiction package name and version
- Statutory tax bands and rates (reference view)
- Statutory leave minimums
- Filing calendar and remittance deadlines

**Configurable by Tenant**

| Setting | Constraint |
|---|---|
| Annual leave entitlement | Must be ≥ statutory minimum |
| Sick leave entitlement | Must be ≥ statutory minimum |
| Carry-forward days | Must be ≥ 0; tenant may reduce max carry-forward (cannot go negative) |
| Overtime policy (multipliers, cap) | No statutory constraint — tenant sets |
| Allowance types | Fully tenant-defined within flags system |
| Deduction types | Fully tenant-defined |
| Payroll cut-off day | Day of month after which next month's run is used for joiners |
| Tax reliefs to auto-apply | Tenant can opt insurance relief in/out globally |
| Pension scheme type | Override NSSF with approved occupational pension scheme |
| Bank payment file format | From platform-supported list |

Any tenant-configured value that would breach the statutory minimum generates a validation error that cannot be bypassed, with a link to the relevant legal reference from the jurisdiction package.

### 2.5 Rate Versioning & Effective Dates

Every tax component record carries `effective_from` and `effective_to` dates. The payroll engine loads rates using the last day of the payroll period as the lookup key.

```go
// domain/statutory/rates.go

func (r *RateRegistry) Resolve(jurisdiction, componentCode string, asOf time.Time) (*TaxComponent, error) {
    // Walk versions for this component, return the one whose
    // effective_from <= asOf <= effective_to (or effective_to is null)
    for _, v := range r.versions[jurisdiction][componentCode] {
        if !v.EffectiveFrom.After(asOf) && (v.EffectiveTo == nil || !v.EffectiveTo.Before(asOf)) {
            return &v, nil
        }
    }
    return nil, ErrNoActiveRate{Jurisdiction: jurisdiction, Component: componentCode, AsOf: asOf}
}
```

When a government publishes new rates (e.g. budget changes), a platform admin loads a new version with the future `effective_from` date. Existing payroll runs are not affected — they always re-read from their own snapshot, not the live registry.

---

## 3. Domain Model

### Aggregate Roots

```
Employee          — central entity; owns contract, bank details, documents, accountability balance
PayrollRun        — immutable record of a payroll computation for a period
LeaveRequest      — request for absence with state machine
TimeSheet         — daily attendance records for a period
ShiftRoster       — assigned shift schedules for a team or employee over a planning horizon
DiscrepancyRecord — employee accountability event (cash/stock over/under)
Organisation      — tenant's legal entity; owns departments, grades, positions
```

### Core Entities

```
EmploymentContract      — terms of employment attached to an Employee
Department              — organisational unit with parent-child hierarchy
Position                — a role slot within a department
Grade                   — pay band defining salary range and step increments
ShiftPattern            — template defining a working schedule (start/end/breaks/days)
ShiftAssignment         — assignment of a ShiftPattern to an Employee for a date range
Roster                  — a collection of ShiftAssignments for a team and planning period
AttendanceRecord        — a single day's clock-in/out record for one employee
AttendanceSource        — enum: biometric | mobile | api | manual
AllowanceType           — tenant-defined earning category
DeductionType           — tenant-defined deduction category
LeaveType               — tenant-defined absence category with accrual rules
PayslipLine             — a single line on a payslip (earning or deduction)
TaxComponent            — a statutory deduction rule from the jurisdiction package
DiscrepancyType         — cash_short | cash_over | stock_shortage | stock_surplus | damage
EmployeeReceivable      — running balance of amounts owed by an employee
```

### Value Objects

```
Money           — amount + ISO 4217 currency code
DateRange       — inclusive start and end date
EmployeeID      — tenant-scoped UUID
ShiftTime       — time-of-day (HH:MM) without timezone; shift dates supply the date component
OvertimeResult  — computed overtime minutes and applicable multiplier
DiscrepancyAmount — quantity + unit + monetary value at standard cost
```

---

## 4. Module Architecture

### Package Layout

```
internal/hr/
├── domain/
│   ├── employee/
│   │   ├── employee.go
│   │   ├── contract.go
│   │   ├── accountability.go     # DiscrepancyRecord, EmployeeReceivable
│   │   └── events.go
│   ├── organisation/
│   │   ├── department.go
│   │   ├── grade.go
│   │   └── position.go
│   ├── leave/
│   │   ├── leave_request.go
│   │   ├── leave_balance.go
│   │   └── accrual.go
│   ├── shift/
│   │   ├── pattern.go            # ShiftPattern, break rules
│   │   ├── roster.go             # Roster, ShiftAssignment
│   │   └── swap.go               # ShiftSwap request & approval
│   ├── attendance/
│   │   ├── record.go             # AttendanceRecord
│   │   ├── source.go             # AttendanceSource, ingestion metadata
│   │   └── reconciler.go        # conflict resolution between sources
│   ├── payroll/
│   │   ├── run.go
│   │   ├── payslip.go
│   │   ├── pipeline.go
│   │   └── stages/
│   └── statutory/
│       ├── jurisdiction.go
│       ├── registry.go           # JurisdictionRegistry, JSON loader
│       ├── rate_resolver.go
│       └── computation/
│           ├── progressive.go    # banded_progressive computation
│           ├── flat.go           # flat_percentage computation
│           ├── band.go           # percentage_of_band computation
│           └── flat_amount.go    # flat_amount_per_employee
├── application/
│   ├── employee_service.go
│   ├── payroll_service.go
│   ├── leave_service.go
│   ├── shift_service.go
│   ├── attendance_service.go
│   └── accountability_service.go
├── infrastructure/
│   ├── postgres/queries/
│   ├── temporal/
│   │   ├── payroll_workflow.go
│   │   ├── leave_accrual_workflow.go
│   │   ├── offboarding_workflow.go
│   │   └── shift_publish_workflow.go
│   └── attendance/
│       ├── biometric/
│       │   ├── zkteco_adapter.go
│       │   └── suprema_adapter.go
│       ├── mobile/
│       │   └── push_handler.go
│       └── api/
│           └── webhook_handler.go
└── transport/http/
    ├── employee_handler.go
    ├── payroll_handler.go
    ├── leave_handler.go
    ├── shift_handler.go
    ├── attendance_handler.go
    └── accountability_handler.go
```

### Layer Responsibilities

| Layer | Responsibility |
|---|---|
| `domain/` | Pure business logic, invariants, domain events. No I/O. |
| `application/` | Use-case orchestration. Calls domain, emits events, coordinates repos. |
| `infrastructure/` | SQLC queries, Temporal workflows, biometric device adapters, webhook handlers. |
| `transport/` | Fiber HTTP handlers, request parsing, response serialisation. |

---

## 5. Sub-Modules

### 5.1 Employee Management

Manages the authoritative record of each employee within a tenant.

#### Employee Status State Machine

```
Draft ──► Pending ──► Active ──► Suspended ──► Active (reinstatement)
                         │
                         └──► Terminated (terminal)
```

#### Key Operations

| Operation | Permission | Outcome |
|---|---|---|
| `OnboardEmployee` | `hr.employee.create` | Record created in `Draft`; contract drafted |
| `ActivateEmployee` | `hr.employee.activate` | Status → `Active`; leave balances initialised; system access provisioned |
| `UpdatePersonalDetails` | `hr.employee.update` | Audit trail entry created |
| `ChangeGrade` | `hr.employee.update` | New contract effective date; salary updated |
| `SuspendEmployee` | `hr.employee.suspend` | Status → `Suspended`; payroll blocked |
| `TerminateEmployee` | `hr.employee.terminate` | Status → `Terminated`; offboarding workflow triggered |

#### Go Domain Model (excerpt)

```go
// domain/employee/employee.go

type Employee struct {
    ID             EmployeeID
    TenantID       TenantID
    EmployeeNumber string
    FullName       string
    DateOfBirth    time.Time
    NationalID     NationalID
    TaxID          TaxID
    Gender         Gender
    Status         Status
    DepartmentID   uuid.UUID
    PositionID     uuid.UUID
    GradeID        uuid.UUID
    ManagerID      *EmployeeID
    JoinDate       time.Time
    Contract       *EmploymentContract
    BankAccounts   []BankAccount
    Identifiers    map[string]string  // jurisdiction-specific: {"kra_pin": "A123...", "nssf_number": "..."}
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

### 5.2 Organisation Structure

#### Hierarchy

```
Organisation (legal entity / tenant)
  └── Department (recursive, N levels)
        ├── Sub-Department
        │     └── Team
        └── Position  (named slot, has Grade, may be Vacant or Filled)
```

#### Grade Structure

Each grade defines a pay band and optional step increments within that band. Positions reference a grade; employees assigned to a position must have salaries within that band unless an override justification is provided.

```go
type Grade struct {
    ID          uuid.UUID
    TenantID    TenantID
    Code        string           // e.g. "G1", "M2", "E1"
    Name        string
    SalaryMin   decimal.Decimal
    SalaryMax   decimal.Decimal
    Currency    string
    Steps       []SalaryStep     // optional step increments
    CreatedAt   time.Time
}
```

### 5.3 Leave & Absence

#### Leave Type Configuration

Each `LeaveType` is either inherited from the jurisdiction package (statutory) or defined by the tenant (contractual). Statutory leave types cannot be deleted and their `min_days_per_year` cannot be reduced below the jurisdiction minimum.

| Field | Type | Description |
|---|---|---|
| `code` | string | Unique per tenant (e.g. `ANNUAL`, `SICK`, `COMPASSIONATE`) |
| `source` | enum | `statutory` \| `contractual` |
| `paid` | bool | Whether leave days receive pay |
| `accrual_method` | enum | `fixed_annual` \| `monthly_accrual` \| `none` |
| `accrual_rate` | decimal | Days per month (monthly_accrual only) |
| `max_days_per_year` | decimal | Cap on entitlement |
| `carry_forward_days` | decimal | Max days carried to next leave year |
| `requires_approval` | bool | Requires manager approval before granting |
| `min_notice_days` | int | Minimum advance notice required |
| `gender_restricted` | enum | `none` \| `male` \| `female` |
| `requires_documentation` | bool | e.g. sick note for sick leave beyond N days |
| `documentation_threshold_days` | int | Days beyond which documentation is mandatory |
| `jurisdiction_codes` | []string | Jurisdictions where this type is legally mandated |

#### Leave Request State Machine

```
Draft ──► Submitted ──► Pending Approval ──► Approved ──► Active ──► Completed
                              │
                              └──► Rejected

Approved ──► Cancelled (employee cancels before start date)
```

#### Accrual Engine

Monthly accrual runs as a Temporal cron activity on the last day of each month, pro-rating for mid-month joiners and accounting for unpaid absence days.

### 5.4 Time & Attendance

Attendance records are sourced from multiple channels (see §6 for full integration detail). The attendance sub-module normalises all sources into a single `AttendanceRecord` per employee per day.

#### AttendanceRecord

```go
type AttendanceRecord struct {
    ID                uuid.UUID
    TenantID          TenantID
    EmployeeID        EmployeeID
    Date              time.Time
    Source            AttendanceSource   // biometric | mobile | api | manual
    SourceRef         string             // device ID, mobile session ID, API request ID
    ScheduledStart    *time.Time         // from shift assignment
    ScheduledEnd      *time.Time
    ActualClockIn     *time.Time
    ActualClockOut    *time.Time
    BreakMinutes      int
    OvertimeMinutes   int
    AbsenceType       AbsenceType        // present | absent | half_day | public_holiday | leave
    LeaveRequestID    *uuid.UUID
    ApprovedOT        bool               // overtime approved by manager
    Notes             string
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

#### TimeSheet Lifecycle

```
Open ──► Submitted ──► Approved ──► Locked (consumed by payroll run)
              │
              └──► Rejected ──► Open (correction)
```

#### Overtime Policy (fully configurable)

```go
type OvertimePolicy struct {
    WeeklyThresholdHours    float64
    DailyThresholdHours     float64
    WeekdayMultiplier       float64
    WeekendMultiplier       float64
    PublicHolidayMultiplier float64
    CapHoursPerMonth        *float64  // nil = uncapped
    RequiresPreApproval     bool
    RequiresManagerApproval bool
}
```

### 5.5 Shift Management

Shift management supports industries where employees work non-standard, rotating, or 24/7 schedules. This sub-module is used by the Retail, Hospital, Forecourt, and Restaurant verticals of Awo ERP.

#### Shift Pattern

A `ShiftPattern` is a reusable template that defines a single working shift.

```go
type ShiftPattern struct {
    ID               uuid.UUID
    TenantID         TenantID
    Code             string            // e.g. "MORNING", "NIGHT", "SPLIT-A"
    Name             string
    StartTime        ShiftTime         // e.g. 06:00
    EndTime          ShiftTime         // e.g. 14:00; may be next calendar day
    CrossesMidnight  bool
    TotalMinutes     int               // computed: end - start - breaks
    Breaks           []BreakRule
    DaysOfWeek       []time.Weekday    // applicable days
    IsOnCall         bool
    DepartmentID     *uuid.UUID        // nil = applicable to all departments
}

type BreakRule struct {
    StartOffset  int   // minutes from shift start
    Duration     int   // minutes
    Paid         bool
}
```

#### Roster

A `Roster` is a planning document that assigns shift patterns to employees over a defined horizon (typically a week or a month).

```go
type Roster struct {
    ID            uuid.UUID
    TenantID      TenantID
    DepartmentID  uuid.UUID
    PeriodStart   time.Time
    PeriodEnd     time.Time
    Status        RosterStatus   // draft | published | locked
    Assignments   []ShiftAssignment
    CreatedBy     uuid.UUID
    PublishedAt   *time.Time
}

type ShiftAssignment struct {
    ID             uuid.UUID
    RosterID       uuid.UUID
    EmployeeID     EmployeeID
    ShiftPatternID uuid.UUID
    Date           time.Time
    Override       *ShiftOverride   // ad-hoc changes to start/end for this specific assignment
}
```

#### Roster Lifecycle

```
Draft ──► Published ──► Locked (attendance records created)
              │
              └──► Recalled ──► Draft (for correction before lock date)
```

A roster is locked automatically by a Temporal scheduled activity at a configurable number of days before the period starts (default: 2 days). Once locked, changes require a shift swap request.

#### Shift Swap

Employees may request to swap shifts with a colleague. The workflow:

```
SwapRequest (employee A proposes swap to employee B)
  ├── Employee B accepts / declines
  ├── If accepted → Manager approves / declines
  └── If approved → RosterAssignments updated; both employees notified
```

Constraints checked:
- Both employees must be in compatible roles or positions
- Swap must not cause an employee to exceed overtime cap
- Swap must not create a rest period violation (minimum hours between shifts)

#### On-Call Management

Employees assigned an `IsOnCall = true` shift are included in the roster as standby. If activated:
- An on-call activation record is created with actual start/end times
- Compensation is computed based on a configurable on-call rate (flat fee, hourly rate, or percentage of basic)
- The activation is included as an allowance line in the next payroll run

#### Industry-Specific Patterns

| Vertical | Typical Pattern | Notes |
|---|---|---|
| Retail | Morning / Afternoon / Evening splits | Peaks on weekends; Sunday premium rate |
| Hospital | Day / Night / Long Day rotation | 12-hour shifts; mandatory rest periods |
| Forecourt | 3-shift 24-hour rotation | Overnight premium; no shift < 8h gap |
| Restaurant | Split shifts (lunch + dinner) | Unpaid mid-day break; flexible scheduling |
| Warehouse | Day / Night with rotating rest days | Weekly rotation prevents fatigue burnout |

#### Minimum Rest Period Validation

The system enforces a configurable minimum rest period between the end of one shift and the start of the next (default: 11 hours, matching most labour law minimums). The roster publisher checks this before allowing a roster to move from `Draft` to `Published`.

### 5.6 Payroll Engine

The payroll engine is the computational core. It processes employees in parallel batches and produces immutable payslips. See §7 for the full pipeline architecture.

#### PayrollRun Status Machine

```
Draft ──► Processing ──► Review ──► Approved ──► Posted ──► Reversed (terminal)
              │
              └──► Failed (employees with errors excluded; run can be resumed)
```

#### Payslip Structure

```go
type Payslip struct {
    ID                  uuid.UUID
    RunID               uuid.UUID
    EmployeeID          EmployeeID
    Period              DateRange
    Currency            string
    WorkingDaysInPeriod int
    DaysPaid            int

    // Earnings
    BasicSalary         decimal.Decimal
    Allowances          []PayslipLine
    GrossEarnings       decimal.Decimal

    // Pre-tax adjustments
    PreTaxDeductions    []PayslipLine
    TaxableIncome       decimal.Decimal

    // Tax and statutory
    StatutoryDeductions []PayslipLine
    TotalStatutory      decimal.Decimal

    // Other post-tax deductions
    OtherDeductions     []PayslipLine
    TotalOtherDeductions decimal.Decimal

    TotalDeductions     decimal.Decimal
    NetPay              decimal.Decimal

    // Employer cost (not deducted from employee — shown for HR/Finance)
    EmployerContributions []PayslipLine
    TotalEmployerCost   decimal.Decimal

    YTDGross            decimal.Decimal
    YTDTax              decimal.Decimal
}
```

#### Multi-Currency Support

Basic salary and contractual allowances are stored in the contract currency. At run time, all amounts are converted to the run currency using the exchange rate snapshot captured at run creation. Exchange rates are immutable after the run is created.

### 5.7 Statutory Compliance

The statutory compliance sub-module does not hard-code any jurisdiction rules. It reads from the jurisdiction registry (§2) and dispatches computation to the appropriate strategy based on the `computation_type` field of each `TaxComponent`.

#### Computation Strategies

| Computation Type | Description | Example |
|---|---|---|
| `banded_progressive` | Tax bands applied incrementally; marginal rate per band | PAYE in most jurisdictions |
| `flat_percentage` | Single rate applied to the entire base | SHIF Kenya, UIF South Africa |
| `percentage_of_ceiling` | Rate applied to the lower of base and ceiling | NSSF Tier I Kenya |
| `percentage_of_band` | Rate applied to the amount between floor and ceiling | NSSF Tier II Kenya |
| `flat_amount_per_employee` | Fixed amount regardless of salary | NITA Kenya |
| `formula` | Arbitrary formula evaluated by `pkg/condition` expression engine | Complex reliefs, CRA in Nigeria |

```go
// domain/statutory/computation/dispatcher.go

type Dispatcher struct {
    strategies map[string]ComputationStrategy
}

func (d *Dispatcher) Compute(component TaxComponent, base decimal.Decimal) (decimal.Decimal, error) {
    s, ok := d.strategies[component.ComputationType]
    if !ok {
        return decimal.Zero, ErrUnknownComputationType{Type: component.ComputationType}
    }
    return s.Compute(component, base)
}
```

This design means any future computation type can be supported by implementing the `ComputationStrategy` interface and registering it — without touching existing strategies.

#### Rate Snapshot

When a payroll run transitions from `Draft` to `Processing`, the system snapshots all resolved `TaxComponent` records for the run's jurisdiction and period into the `payroll_runs.rate_snapshot` JSONB column. This snapshot is immutable and is the single source of truth for any audit or re-computation query related to that run.

### 5.8 Benefits & Allowances

#### AllowanceType Flags

| Flag | Effect |
|---|---|
| `taxable` | Added to taxable income before PAYE computation |
| `pensionable` | Included in pensionable pay for contribution computation |
| `prorated_on_absence` | Reduced proportionally for unpaid absence days |
| `fixed_amount` | Fixed value; otherwise a formula or percentage of basic |
| `payable_in_arrears` | Paid in period after earned |
| `included_in_gross` | Shown in gross earnings on payslip |
| `on_call_compensation` | Only paid when an on-call shift is activated |

#### DeductionType Flags

| Flag | Effect |
|---|---|
| `statutory` | System-managed; HR cannot edit |
| `pre_tax` | Deducted from gross before PAYE (reduces tax liability) |
| `loan_recovery` | Linked to a loan schedule; auto-closes when balance reaches zero |
| `accountability_recovery` | Recovery of an employee discrepancy receivable |
| `requires_approval` | Requires manager or Finance sign-off |
| `capped_at_net` | Cannot reduce net pay below zero |

### 5.9 Employee Accountability

This sub-module handles situations where an employee is responsible for a cash register, a stock station, a fuel nozzle, or any other asset where a quantifiable discrepancy can be attributed to them. Common scenarios:

- **Cashier short/over** — the till balance at end of shift differs from the expected balance
- **Fuel/stock variance** — a forecourt attendant's shift-end wetstock figure shows a loss not explained by legitimate sales
- **Product damage** — an employee causes damage to goods whose cost can be quantified
- **Key fob / equipment loss** — an employee loses company property with a known replacement cost

#### Discrepancy Lifecycle

```
Identified ──► Under Review ──► Disputed ──► Resolved
                    │                │
                    │                └──► Waived (management decision)
                    │
                    └──► Confirmed ──► Recovery Scheduled ──► Recovered (cleared)
                                │
                                └──► Written Off
```

#### DiscrepancyRecord

```go
type DiscrepancyRecord struct {
    ID              uuid.UUID
    TenantID        TenantID
    EmployeeID      EmployeeID
    Type            DiscrepancyType      // cash_short | cash_over | stock_shortage | damage | equipment_loss
    IncidentDate    time.Time
    ShiftID         *uuid.UUID
    Description     string
    ExpectedValue   decimal.Decimal
    ActualValue     decimal.Decimal
    VarianceAmount  decimal.Decimal      // always positive; sign determined by Type
    Currency        string
    Status          DiscrepancyStatus
    RecoveryMethod  RecoveryMethod       // payroll_deduction | direct_payment | waived | write_off
    RecoveryPlanID  *uuid.UUID
    EvidenceRefs    []string             // document storage keys
    RaisedBy        uuid.UUID
    ReviewedBy      *uuid.UUID
    ResolvedAt      *time.Time
    Notes           string
    CreatedAt       time.Time
}
```

#### Recovery Plan

When a discrepancy is confirmed and the recovery method is `payroll_deduction`, a `RecoveryPlan` is created:

```go
type RecoveryPlan struct {
    ID                 uuid.UUID
    DiscrepancyID      uuid.UUID
    EmployeeID         EmployeeID
    TotalAmount        decimal.Decimal
    RecoveredAmount    decimal.Decimal
    RemainingAmount    decimal.Decimal
    Instalments        []RecoveryInstalment
    Status             RecoveryPlanStatus  // active | completed | cancelled
}

type RecoveryInstalment struct {
    Period          DateRange
    Amount          decimal.Decimal
    DeductedRunID   *uuid.UUID   // set when payroll run has consumed this instalment
    Status          InstalmentStatus  // pending | deducted | skipped
}
```

The payroll pipeline's `ApplyOtherDeductions` stage queries active recovery plans for each employee and includes due instalments as `accountability_recovery` deduction lines on the payslip.

#### Cash Over Handling

A cash surplus (`cash_over`) can be handled in two ways, configurable per tenant:

1. **Remit to company** — surplus is treated as company income; no employee benefit
2. **Credit to employee** — surplus is held in a credit balance against the employee; future shortfalls are first offset against this credit before raising a deduction

#### Integration with Shift Module

When a discrepancy is linked to a `ShiftID`, the HR module can pull the assigned employee directly from the roster. In forecourt operations, this enables automatic variance attribution when the Inventory / Wetstock module reports a pump-level variance.

### 5.10 Offboarding & Separation

When an employee is terminated, a Temporal workflow orchestrates the full separation process:

```
TerminationInitiated
  ├── ComputeTerminalBenefits     (reads jurisdiction terminal_benefits rules)
  ├── FinalPayrollRun             (pro-rated last month + terminal benefits)
  ├── ClearOutstandingDeductions  (close loan schedules; mark unsettled)
  ├── SettleDiscrepancies         (flag any open accountability records)
  ├── RevokeSystemAccess          (IAM event)
  ├── IssueServiceCertificate     (document generation)
  └── ArchiveEmployeeRecord       (status → Terminated)
```

Terminal benefit formulas are read from the active jurisdiction package's `terminal_benefits` block and evaluated by the `pkg/condition` expression engine.

---

## 6. Attendance Integration Layer

The attendance integration layer ingests clock events from multiple sources and normalises them into `AttendanceRecord` entries. All raw events are stored in an append-only `attendance_raw_events` table before normalisation, enabling full auditability and re-processing.

### 6.1 Biometric Devices

#### Supported Protocols

| Protocol | Devices | Description |
|---|---|---|
| ZKTeco SDK / PUSH | ZKTeco series (K20, F18, SpeedFace) | Device pushes events to a configured server endpoint |
| BioConnect REST API | Suprema BioStar 2, BioEntry | Cloud-managed; Awo polls or subscribes to webhooks |
| ADMS (Attendance Data Management Service) | ZKTeco ADMS-compatible devices | UDP/HTTP pull from device |
| CSV Import | Any device with CSV export | Batch import via scheduled upload or SFTP |
| ONVIF / proprietary | Hikvision, Dahua | Event subscription via device-specific adapter |

#### Device Registration

Each physical device is registered as a `BiometricDevice` record:

```go
type BiometricDevice struct {
    ID           uuid.UUID
    TenantID     TenantID
    DeviceCode   string          // unique per tenant, printed on device
    Name         string          // e.g. "Main Entrance Scanner"
    Location     string
    Protocol     BiometricProtocol
    IPAddress    string
    Port         int
    SerialNumber string
    DepartmentID *uuid.UUID      // if nil, events apply to any employee
    IsActive     bool
    LastHeartbeat time.Time
}
```

#### ZKTeco PUSH Adapter (example)

The ZKTeco SDK pushes attendance logs to a configurable HTTP endpoint. Awo registers a webhook receiver at `/integrations/biometric/zkteco/{device_id}/push`.

```go
// infrastructure/attendance/biometric/zkteco_adapter.go

type ZKTecoPushPayload struct {
    SN         string `json:"sn"`         // device serial number
    Ret        int    `json:"ret"`         // 200 = success
    AttStamp   []struct {
        Pin    string `json:"pin"`         // maps to employee's biometric_pin
        Time   string `json:"time"`        // "2024-11-14 08:03:17"
        Status int    `json:"status"`      // 0=check-in, 1=check-out, 4=OT-in, 5=OT-out
        Verify int    `json:"verify"`      // 1=fingerprint, 4=face, 15=card
    } `json:"AttStamp"`
}

func (a *ZKTecoAdapter) HandlePush(ctx context.Context, payload ZKTecoPushPayload) error {
    for _, stamp := range payload.AttStamp {
        employee, err := a.employeeRepo.FindByBiometricPin(ctx, stamp.Pin)
        if err != nil {
            a.log.Warn("unknown biometric pin", "pin", stamp.Pin, "device", payload.SN)
            continue
        }
        raw := RawAttendanceEvent{
            TenantID:    employee.TenantID,
            EmployeeID:  employee.ID,
            Source:      SourceBiometric,
            SourceRef:   payload.SN,
            EventTime:   parseZKTime(stamp.Time),
            EventType:   mapZKStatus(stamp.Status),
            VerifyMethod: mapZKVerify(stamp.Verify),
            RawPayload:  stamp,
        }
        return a.rawEventRepo.Insert(ctx, raw)
    }
    return nil
}
```

#### Employee Biometric Enrollment

Each employee has a `biometric_pin` field (the ID programmed into the device). Enrollment is done at device level (fingerprint / face scan) and associated to the employee record in Awo via the biometric PIN.

### 6.2 Mobile Clock-In/Out

Mobile attendance is available in the Awo mobile app (iOS / Android) and the Progressive Web App. It is intended for field employees, remote workers, or sites without biometric hardware.

#### Features

| Feature | Description |
|---|---|
| GPS geofencing | Clock-in only permitted within a configurable radius of a registered work location |
| Selfie verification | Optional photo capture at clock-in; stored for audit |
| Offline mode | Events queued locally when offline; synced when connectivity resumes |
| QR code clock-in | Employee scans a static QR code at a location (alternative to GPS) |
| NFC tap | Tap NFC-enabled phone on a registered NFC tag at location |

#### Mobile Event Payload

```json
{
  "employee_id": "uuid",
  "event_type": "clock_in",
  "event_time": "2024-11-14T08:02:45Z",
  "location": {
    "latitude": -1.286389,
    "longitude": 36.817223,
    "accuracy_meters": 8
  },
  "work_location_id": "uuid",
  "geofence_check": {
    "passed": true,
    "distance_meters": 34,
    "radius_meters": 100
  },
  "selfie_key": "uploads/selfies/uuid.jpg",
  "device_id": "device-fingerprint",
  "app_version": "2.3.1",
  "offline_queued": false
}
```

#### Geofence Configuration

```go
type WorkLocation struct {
    ID              uuid.UUID
    TenantID        TenantID
    Name            string
    Latitude        float64
    Longitude       float64
    RadiusMeters    int
    AllowMobile     bool
    RequireSelfie   bool
    DepartmentIDs   []uuid.UUID  // which departments use this location
}
```

### 6.3 API Integration

Third-party workforce management systems, POS systems, or custom hardware can push attendance events directly via a REST API or webhook.

#### Push Endpoint

```
POST /api/v1/integrations/attendance
Authorization: Bearer <api_key>   (tenant-scoped API key)
X-Tenant: <slug>
```

```json
{
  "events": [
    {
      "employee_id": "uuid-or-null",
      "employee_external_ref": "EMP-0042",   // alternative to employee_id
      "event_type": "clock_in",              // clock_in | clock_out | break_start | break_end
      "event_time": "2024-11-14T08:05:00Z",
      "source_system": "SIMBA_POS",
      "source_ref": "SIMBA-TXN-98234",
      "metadata": {}
    }
  ]
}
```

The endpoint validates each event, resolves `employee_external_ref` to an internal `employee_id`, persists raw events, and returns a summary:

```json
{
  "accepted": 47,
  "rejected": 2,
  "errors": [
    { "index": 5, "code": "EMPLOYEE_NOT_FOUND", "ref": "EMP-9999" },
    { "index": 23, "code": "FUTURE_EVENT_TIME", "event_time": "2024-11-15T08:00:00Z" }
  ]
}
```

#### Webhook Subscription (Pull Model Alternative)

For systems that prefer to subscribe, Awo exposes a webhook consumer model. The external system registers a webhook URL with Awo; when Awo needs to confirm attendance data (e.g. from a POS system that controls access), it calls that URL.

### 6.4 Reconciliation & Conflict Resolution

Multiple sources may deliver events for the same employee on the same day. The reconciliation engine runs nightly (Temporal cron) and applies the following priority rules:

| Priority | Source | Rationale |
|---|---|---|
| 1 (highest) | `biometric` | Hardware-captured; tamper-resistant |
| 2 | `api` | System-integrated; trusted source |
| 3 | `mobile` | GPS-verified but device-controlled |
| 4 (lowest) | `manual` | Human-entered; requires approval |

When conflicting events exist:

1. The highest-priority source wins for the primary `AttendanceRecord`.
2. All other events are preserved in `attendance_raw_events` with a `conflict_resolution` status.
3. The reconciliation result is flagged for HR review if the difference between sources exceeds a configurable threshold (default: 15 minutes).
4. HR can override the automatic resolution manually.

---

## 7. Payroll Pipeline Architecture

### Stage Interface

```go
// domain/payroll/pipeline.go

type Stage interface {
    Name() string
    // Enabled returns false when a feature flag or configuration excludes this stage
    Enabled(cfg PayrollConfig) bool
    Process(ctx context.Context, in EmployeePayrollInput) (EmployeePayrollOutput, error)
}

type Runner struct {
    stages []Stage
    cfg    PayrollConfig
}

func (r *Runner) Run(ctx context.Context, in EmployeePayrollInput) (EmployeePayrollOutput, error) {
    out := EmployeePayrollOutput{Input: in}
    for _, stage := range r.stages {
        if !stage.Enabled(r.cfg) {
            continue
        }
        var err error
        out, err = stage.Process(ctx, out.AsInput())
        if err != nil {
            return out, fmt.Errorf("stage %s: %w", stage.Name(), err)
        }
    }
    return out, nil
}
```

### Pipeline Stages (ordered)

| # | Stage | Description |
|---|---|---|
| 1 | `LoadEmployee` | Fetch active contract, grade, allowances, deduction entries, open recovery plans |
| 2 | `ComputeWorkingDays` | Determine paid days from timesheet and leave records; cross-reference shift roster |
| 3 | `ProrateSalary` | Reduce basic salary for joiners, leavers, or unpaid absence days |
| 4 | `ApplyAllowances` | Add all active allowance entries; evaluate formula-based allowances; prorate where flagged |
| 5 | `ApplyOnCallCompensation` | Add compensation for any on-call activations in the period |
| 6 | `ApplyPreTaxDeductions` | Subtract pre-tax deductions (pension, approved pre-tax schemes) |
| 7 | `ComputeTaxableIncome` | `gross - pre_tax_deductions - applicable_reliefs` (reliefs from jurisdiction package) |
| 8 | `ComputeIncomeTax` | Dispatch to jurisdiction computation strategy; apply reliefs |
| 9 | `ComputeStatutoryContributions` | Iterate all active `TaxComponent` records for the jurisdiction; compute each |
| 10 | `ApplyAccountabilityRecoveries` | Insert due instalment lines from active recovery plans |
| 11 | `ApplyOtherDeductions` | Loan recoveries, advances, other post-tax deductions in priority order |
| 12 | `ComputeNetPay` | `gross - total_deductions`; enforce `capped_at_net` flag |
| 13 | `UpdateYTD` | Accumulate year-to-date gross, tax, and contributions |
| 14 | `ComputeEmployerCost` | Separately sum employer-side contributions for Finance reporting |
| 15 | `ValidateOutput` | Net ≥ 0; totals balance; all required fields present; statutory minimums met |

### Data Flow

```
EmployeePayrollInput
  ├── employee_id, run_id, period
  ├── basic_salary, currency, contract_type
  ├── working_days_in_period, days_paid
  ├── allowance_entries[]
  ├── deduction_entries[]
  ├── recovery_plan_instalments[]
  ├── on_call_activations[]
  ├── tax_components[]          ← from rate_snapshot
  ├── reliefs[]                 ← from rate_snapshot
  └── ytd_snapshot

        ──[15-stage pipeline]──►

EmployeePayrollOutput
  ├── gross_earnings
  ├── pre_tax_deductions{}
  ├── taxable_income
  ├── income_tax
  ├── statutory_deductions{}
  ├── accountability_recoveries{}
  ├── other_deductions{}
  ├── net_pay
  ├── employer_contributions{}
  ├── ytd_gross, ytd_tax
  └── payslip_lines[]
```

---

## 8. General Ledger Integration

### 8.1 Design Principles

The HR module has a hard boundary with the Finance module. It never writes journal entries, never references GL account codes directly in its domain logic, and never reads from the ledger. Instead:

1. The HR module produces a structured `PayrollJournalInstruction` after a run is approved.
2. This instruction is emitted as a `PayrollPosted` domain event.
3. The Finance module consumes the event and creates the journal entries using its own account mapping configuration.
4. The Finance module emits a `JournalCreated` event back to confirm posting; HR updates the run status to `posted`.

This boundary ensures that a misconfigured payroll can never corrupt the General Ledger, and the Finance module remains the single source of truth for all financial data.

### 8.2 Chart of Accounts Structure

The Finance module maintains a payroll-specific section in the Chart of Accounts. The recommended structure follows a separation between expense accounts (P&L) and liability/clearing accounts (Balance Sheet).

```
INCOME STATEMENT (P&L)
├── 5000  Personnel Expenses
│     ├── 5100  Salaries & Wages
│     │     ├── 5110  Basic Salaries
│     │     ├── 5120  Overtime Pay
│     │     └── 5130  Casual / Contract Labour
│     ├── 5200  Allowances
│     │     ├── 5210  Housing Allowance
│     │     ├── 5220  Transport Allowance
│     │     └── 5230  Other Allowances
│     ├── 5300  Employer Statutory Contributions
│     │     ├── 5310  Employer NSSF / Pension
│     │     ├── 5320  Employer Health Levy
│     │     ├── 5330  Affordable Housing Levy (Employer)
│     │     └── 5340  Skills Development Levy
│     └── 5400  Other Staff Costs
│           ├── 5410  Staff Training
│           └── 5420  Employee Welfare

BALANCE SHEET (Liabilities)
├── 2200  Payroll Liabilities
│     ├── 2210  Payroll Clearing Account       ← KEY ACCOUNT (see §8.7)
│     ├── 2220  PAYE Payable
│     ├── 2230  NSSF Payable (Employee)
│     ├── 2231  NSSF Payable (Employer)
│     ├── 2240  Health Insurance Payable
│     ├── 2250  Housing Levy Payable
│     └── 2260  Other Statutory Payables

BALANCE SHEET (Assets)
├── 1300  Employee Receivables (Sub-Ledger Control)
│     └── 1310  Employee Accountability Receivable  ← for discrepancy recoveries
```

Cost centres can be appended as dimension codes to expense accounts, enabling per-department P&L breakdown without multiplying the account structure:

```
5110-FORECOURT    Basic Salaries — Forecourt
5110-RETAIL       Basic Salaries — Retail
5110-ADMIN        Basic Salaries — Administration
```

### 8.3 Employee Sub-Ledger

**Employees do not each need a dedicated GL account.** The General Ledger operates at the control account level; the detail lives in the HR module's sub-ledger.

**Rule:** The GL sees control accounts; the HR module sees individual employee balances.

| GL Control Account | HR Sub-Ledger | Detail |
|---|---|---|
| `2210 Payroll Clearing` | `payroll_runs.net_pay` per employee | Per-run net pay obligation |
| `1310 Employee Accountability Receivable` | `employee_receivables` table | Per-employee discrepancy balance |
| `2230 NSSF Payable (Employee)` | `payslip_lines` (NSSF employee) | Per-employee deduction per run |

The Finance module's reconciliation report can drill down from any control account balance to the underlying HR sub-ledger entries. This keeps the GL clean (one line per account per posting) while preserving full employee-level audit detail.

**Exception — Employee Loan Advances:** If the company extends personal loans or salary advances to employees, these are tracked per-employee as `EmployeeReceivable` records in the HR module, with `1310` as the GL control. Recovery via payroll reduces the HR sub-ledger balance and posts:

```
Dr 2210 Payroll Clearing          (reducing the net pay obligation)
Cr 1310 Employee Receivable Ctrl  (clearing the asset)
```

### 8.4 Payroll Posting Flow

When a payroll run is approved and posted, the following journal entries are generated.

#### Step 1 — Payroll Accrual (on approval)

This records the payroll expense and the obligation to pay employees and statutory bodies.

```
Dr  5110  Basic Salaries                     [gross basic per department]
Dr  5220  Allowances                         [gross allowances per type]
Dr  5310  Employer NSSF                      [employer contribution]
Dr  5330  Employer Housing Levy              [employer AHL]
Dr  5340  Skills Development Levy            [SDL where applicable]

Cr  2210  Payroll Clearing Account           [total net pay to employees]
Cr  2220  PAYE Payable                       [total PAYE withheld]
Cr  2230  NSSF Payable (Employee)            [employee NSSF]
Cr  2231  NSSF Payable (Employer)            [employer NSSF]
Cr  2240  Health Insurance Payable           [SHIF / NHIF]
Cr  2250  Housing Levy Payable               [employee + employer AHL]
Cr  1310  Employee Accountability Rec.Ctrl   [accountability recovery deductions]
Cr  [Loan Liability / Advance Account]       [loan recovery deductions]
```

The debit side hits P&L expense accounts (split by cost centre / department).  
The credit side hits Balance Sheet liability accounts.

**The P&L sees the full cost of employing someone — gross salary plus employer contributions.**  
Net pay and tax are balance sheet movements only.

#### Step 2 — Bank Transfer (on payment)

When the bank file is processed and confirmed, the net pay obligation is settled:

```
Dr  2210  Payroll Clearing Account           [total net pay]
Cr  1010  Bank Account (Operating)           [amount transferred]
```

The Payroll Clearing Account now has a zero balance (assuming all employees are paid in the same run).

### 8.5 Statutory Remittance Posting

Statutory deductions accumulate as liabilities until remitted to the tax authority. Remittance is a Finance module operation triggered either manually or on a scheduled basis.

#### PAYE Remittance

```
Dr  2220  PAYE Payable
Cr  1010  Bank Account (Operating)
```

#### NSSF Remittance (combined employee + employer)

```
Dr  2230  NSSF Payable (Employee)
Dr  2231  NSSF Payable (Employer)
Cr  1010  Bank Account (Operating)
```

#### Remittance Reconciliation

The Finance module generates a statutory remittance schedule for each liability account showing:
- Opening balance (from prior period)
- Additions this period (from payroll posting)
- Payments made
- Closing balance (should match the tax authority's records)

### 8.6 Employee Accountability Posting

#### When a discrepancy is confirmed

```
Dr  1310  Employee Accountability Receivable Ctrl   [amount of discrepancy]
Cr  5XXX  Cash Shortage / Inventory Variance        [expense / contra account]
```

The debit creates an asset (money owed to the company). The credit hits an expense account (e.g. `5450 Cash Shortage` or is netted against the relevant Inventory Variance account from the Inventory module).

#### When a recovery instalment is deducted via payroll

The payroll posting already includes the deduction (§8.4 Step 1). The specific accounting is:

```
[Part of Step 1 above]
Cr  1310  Employee Accountability Rec.Ctrl    [instalment amount]
```

This reduces the asset as recovery progresses.

#### When a discrepancy is waived or written off

```
Dr  5460  Discrepancy Write-Offs
Cr  1310  Employee Accountability Receivable Ctrl
```

#### Cash Over Handling (where surplus is remitted to company)

```
Dr  1010  Bank / Petty Cash
Cr  4XXX  Cash Surplus Income             [or contra to Cost of Sales]
```

### 8.7 Payroll Clearance Account

The `2210 Payroll Clearing Account` is a balance sheet liability account that acts as a bridge between the payroll accrual and the bank payment. It should ideally carry a zero balance at month-end (all employees paid).

**Monitoring:**

A non-zero balance at month-end indicates:
- Employees not yet paid (bank file not processed)
- Reversal runs not matched to original runs
- Unclaimed salaries (employee bank account rejected)

The Finance module flags a non-zero month-end clearing balance for investigation. Unclaimed amounts are reclassified to `2290 Unclaimed Wages Payable` after a configurable aging period.

**Multi-run scenarios:**  
A company running a monthly staff payroll and a separate weekly casual payroll will have one Payroll Clearing Account entry per run. The account is reconciled per run using the `run_id` tag on each journal entry line.

### 8.8 P&L Cleanliness Strategy

A well-structured payroll GL integration produces a P&L where:

1. **Personnel costs are visible by type and department.** Using the account structure in §8.2 plus cost centre dimensions, management can see exactly what the payroll costs are by grade, department, or location — without needing to drill into the payroll system.

2. **Employer contributions are separated from gross wages.** `5310 Employer NSSF` and `5330 Employer Housing Levy` are distinct lines, making it easy to see the true cost of employment beyond the salary line.

3. **Discrepancies are separate from trading revenue/costs.** `5450 Cash Shortage` and `5460 Discrepancy Write-Offs` are distinct P&L lines, not netted against sales or inventory, so they are clearly visible in variance analysis.

4. **No net pay appears on the P&L.** Net pay is entirely a balance sheet movement (Clearing → Bank). The P&L sees gross salary + employer contributions. This is correct accounting — tax and employee contributions are not costs to the company (they are withheld funds).

5. **Reversal runs produce equal and opposite entries.** Reversals are mirror journal entries; they do not create new expense lines. The correction run's entries are the only ones that appear as a replacement expense.

6. **Period allocation is clean.** Payroll accrual (Step 1) is posted in the period the salary is earned. The bank transfer (Step 2) may happen in the next period. This matches the accrual basis of accounting and keeps the P&L period-accurate.

### 8.9 Event Contract with Finance Module

```go
// Emitted by HR module when a run is posted
type PayrollPosted struct {
    RunID          uuid.UUID
    TenantID       TenantID
    Period         DateRange
    Currency       string
    PostedAt       time.Time
    JournalLines   []JournalInstruction
}

type JournalInstruction struct {
    AccountCode    string               // HR module uses semantic codes, not GL IDs
    AccountType    string               // expense | liability | asset
    CostCentre     string               // department code
    Debit          *decimal.Decimal
    Credit         *decimal.Decimal
    Description    string               // e.g. "Basic Salaries — November 2024"
    Reference      string               // run_id for traceability
    EmployeeCount  int                  // for control total reconciliation
}
```

The Finance module maintains an `AccountMapping` table that resolves the HR module's semantic account codes (e.g. `BASIC_SALARY`, `EMPLOYER_NSSF`) to the tenant's actual GL account IDs. This mapping is configured once during system setup and can be updated by the Finance administrator.

---

## 9. API Reference

All endpoints require `Authorization: Bearer <token>` and `X-Tenant: <slug>`.

### 9.1 Jurisdiction Management (Platform Admin)

```
GET    /platform/jurisdictions                      List all jurisdiction packages
POST   /platform/jurisdictions                      Upload a new jurisdiction package (JSON)
GET    /platform/jurisdictions/{code}               Get current active package for a jurisdiction
GET    /platform/jurisdictions/{code}/versions      List all versions
POST   /platform/jurisdictions/{code}/activate      Activate a specific version
PUT    /platform/jurisdictions/{code}/components/{component_code}   Edit a single tax component
```

### 9.2 Employee Endpoints

```
POST   /employees                                   Create employee (status: Draft)
GET    /employees                                   List employees (filterable)
GET    /employees/{id}                              Get employee detail
PATCH  /employees/{id}                              Update employee details
PATCH  /employees/{id}/activate                     Activate employee
PATCH  /employees/{id}/suspend                      Suspend employee
PATCH  /employees/{id}/terminate                    Terminate employee (triggers offboarding)
GET    /employees/{id}/payslips                     List payslips
GET    /employees/{id}/leave/balances               Get leave balances
GET    /employees/{id}/accountability               Get accountability history and open balance
```

#### `POST /employees` Request

```json
{
  "first_name": "Amina",
  "last_name": "Odhiambo",
  "date_of_birth": "1990-05-14",
  "gender": "female",
  "department_id": "uuid",
  "position_id": "uuid",
  "grade_id": "uuid",
  "manager_id": "uuid",
  "join_date": "2024-12-01",
  "identifiers": {
    "national_id": "30491234",
    "kra_pin": "A123456789Z"
  },
  "contract": {
    "type": "permanent",
    "basic_salary": 120000,
    "currency": "KES",
    "effective_from": "2024-12-01"
  }
}
```

**Error Codes**

| Code | Condition |
|---|---|
| `DUPLICATE_IDENTIFIER` | A jurisdiction-required identifier already exists in this tenant |
| `GRADE_SALARY_OUT_OF_BAND` | Basic salary outside grade range |
| `POSITION_ALREADY_FILLED` | Position occupied by an active employee |
| `BELOW_MINIMUM_WAGE` | Basic salary below jurisdiction's minimum wage |

### 9.3 Shift Management Endpoints

```
POST   /shifts/patterns                             Create shift pattern
GET    /shifts/patterns                             List shift patterns
PATCH  /shifts/patterns/{id}                        Update shift pattern

POST   /shifts/rosters                              Create roster
GET    /shifts/rosters/{id}                         Get roster with assignments
PATCH  /shifts/rosters/{id}/publish                 Publish roster (validates rest periods)
PATCH  /shifts/rosters/{id}/recall                  Recall published roster
POST   /shifts/rosters/{id}/assignments             Add / update assignments
DELETE /shifts/rosters/{id}/assignments/{assignment_id}

POST   /shifts/swaps                                Request shift swap
PATCH  /shifts/swaps/{id}/accept                    Target employee accepts
PATCH  /shifts/swaps/{id}/approve                   Manager approves
PATCH  /shifts/swaps/{id}/decline                   Target employee or manager declines

GET    /employees/{id}/shifts                       Get employee's shift assignments
```

### 9.4 Attendance Endpoints

```
POST   /attendance/events                           Push attendance event(s) (API integration)
GET    /attendance/records                          List attendance records (filterable by date, employee, department)
GET    /attendance/records/{id}                     Get single record
PATCH  /attendance/records/{id}                     Manual override (requires hr.attendance.manage)
GET    /attendance/conflicts                        List records flagged for reconciliation review

GET    /integrations/biometric/devices              List registered biometric devices
POST   /integrations/biometric/devices              Register a new device
GET    /integrations/biometric/devices/{id}/status  Device heartbeat / connectivity status
POST   /integrations/biometric/zkteco/{device_id}/push   ZKTeco push receiver (device calls this)
```

### 9.5 Payroll Endpoints

```
POST   /payroll/runs                                Initiate payroll run
GET    /payroll/runs                                List runs (filterable)
GET    /payroll/runs/{id}                           Run summary: status, totals, error count
GET    /payroll/runs/{id}/payslips                  All payslips for a run
GET    /payroll/runs/{id}/errors                    Employees excluded due to errors
PATCH  /payroll/runs/{id}/approve                   Approve run (review → approved)
PATCH  /payroll/runs/{id}/post                      Post run (approved → posted; emits PayrollPosted)
PATCH  /payroll/runs/{id}/reverse                   Reverse a posted run
GET    /payroll/payslips/{id}                       Get individual payslip
GET    /payroll/payslips/{id}/pdf                   Download payslip as PDF
```

#### `POST /payroll/runs` Request

```json
{
  "period_start": "2024-11-01",
  "period_end": "2024-11-30",
  "pay_date": "2024-11-29",
  "currency": "KES",
  "employee_ids": null,
  "department_ids": null,
  "include_on_call_compensation": true,
  "include_accountability_recoveries": true
}
```

**Response** `202 Accepted`

```json
{
  "run_id": "uuid",
  "status": "processing",
  "workflow_id": "payroll-run-uuid",
  "jurisdiction": "KE",
  "rate_snapshot_version": "KE-2024-07-01"
}
```

### 9.6 Accountability Endpoints

```
POST   /accountability/discrepancies                Raise a discrepancy record
GET    /accountability/discrepancies                List discrepancies (filterable)
GET    /accountability/discrepancies/{id}           Get discrepancy detail
PATCH  /accountability/discrepancies/{id}/confirm   Confirm discrepancy
PATCH  /accountability/discrepancies/{id}/waive     Waive discrepancy
PATCH  /accountability/discrepancies/{id}/dispute   Mark as disputed (employee challenge)
POST   /accountability/discrepancies/{id}/recovery-plan   Create recovery plan
GET    /employees/{id}/accountability/balance       Current outstanding receivable balance
```

### 9.7 Statutory Report Endpoints

```
GET    /payroll/reports/income-tax-return           Jurisdiction-specific income tax return
GET    /payroll/reports/social-contribution-schedule  NSSF / pension remittance schedule
GET    /payroll/reports/payroll-journal             Journal entry instructions for Finance
GET    /payroll/reports/employer-cost               Total cost of employment per department
GET    /payroll/reports/ytd-summary                 Year-to-date summary per employee
GET    /payroll/reports/bank-file                   Bank payment instruction file
```

**Query Parameters** (all report endpoints):

| Param | Description |
|---|---|
| `run_id` | Scope to a specific run |
| `period_start` / `period_end` | Date range filter |
| `department_id` | Scope to a department |
| `format` | `json` \| `csv` \| `pdf` \| jurisdiction-specific form codes |

---

## 10. Business Rules & Validation

### 10.1 Employee Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `HR-EMP-001` | `employee_number` unique within tenant | Error |
| `HR-EMP-002` | All jurisdiction-required identifiers (`identifier_fields[required=true]`) must be present before activation | Error |
| `HR-EMP-003` | Age must be ≥ 16 on `join_date` (minimum working age) | Error |
| `HR-EMP-004` | `basic_salary` ≥ jurisdiction minimum wage as of `join_date` | Error |
| `HR-EMP-005` | `basic_salary` within grade salary band (configurable: hard error or warning) | Configurable |
| `HR-EMP-006` | Only one active employment contract per employee at a time | Error |
| `HR-EMP-007` | Suspended employee's payroll must be blocked until reinstatement | Error |
| `HR-EMP-008` | Terminated employee cannot appear in a new payroll run | Error |
| `HR-EMP-009` | Biometric PIN must be unique within a tenant | Error |

### 10.2 Shift & Roster Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `HR-SHF-001` | Minimum rest period between shifts must be ≥ `jurisdiction.work_week.min_rest_hours` (default 11) | Error |
| `HR-SHF-002` | An employee cannot be assigned two shifts on the same date | Error |
| `HR-SHF-003` | Shifts crossing midnight must set `crosses_midnight = true`; `end_time < start_time` is only valid when this flag is set | Error |
| `HR-SHF-004` | Roster cannot be published while any assignment violates rest period rules | Error |
| `HR-SHF-005` | Swap requests must be between employees with compatible positions | Error |
| `HR-SHF-006` | A locked roster's assignments cannot be changed directly; swap requests must be used | Error |
| `HR-SHF-007` | On-call shift activation duration must be within the on-call window | Warning |

### 10.3 Attendance Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `HR-ATT-001` | Clock-out time must be after clock-in time | Error |
| `HR-ATT-002` | Event time must not be in the future (± configurable tolerance, default 5 min) | Error |
| `HR-ATT-003` | Mobile clock-in must pass geofence check unless `geofence_bypass` is enabled for the employee | Error |
| `HR-ATT-004` | A day with no clock-in or approved leave is marked absent automatically | Automatic |
| `HR-ATT-005` | Overnight shifts must have their clock-out on the day after clock-in | Automatic |
| `HR-ATT-006` | Manual attendance overrides require `hr.attendance.manage` permission | Error |

### 10.4 Payroll Run Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `HR-PAY-001` | Run period cannot span two calendar years | Error |
| `HR-PAY-002` | Two non-reversed runs for the same tenant and period cannot coexist | Error |
| `HR-PAY-003` | `pay_date` ≥ `period_end` | Error |
| `HR-PAY-004` | Net pay for any employee must be ≥ 0 (unless `capped_at_net` flags still leave a deficit — block that deduction instead) | Error |
| `HR-PAY-005` | A `posted` run cannot be edited; only reversal is permitted | Error |
| `HR-PAY-006` | Rate snapshot must exist and be locked before `processing` begins | Error |
| `HR-PAY-007` | Employees without approved timesheets are excluded when `require_timesheet_approval = true` | Configurable |
| `HR-PAY-008` | Employer contributions must appear as separate payslip lines | Error |
| `HR-PAY-009` | Exchange rates are snapshotted at run creation and immutable thereafter | Error |

### 10.5 Accountability Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `HR-ACC-001` | `variance_amount` must be positive; type (short/over) captures the direction | Error |
| `HR-ACC-002` | A discrepancy cannot be confirmed if it is in `Disputed` status without an overriding manager approval | Error |
| `HR-ACC-003` | Recovery instalment amount must not cause net pay to go below zero; excess is carried to the next period | Automatic |
| `HR-ACC-004` | Cash over credits are held per employee and offset against future shortfalls before a deduction is raised | Configurable |
| `HR-ACC-005` | Written-off discrepancies generate a write-off GL instruction to the Finance module | Error |
| `HR-ACC-006` | A waived discrepancy still generates a journal entry (Dr Discrepancy Write-Off, Cr Employee Receivable) | Error |

### 10.6 Leave Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `HR-LVE-001` | Available balance ≥ requested days unless `allow_negative_leave_balance = true` | Configurable |
| `HR-LVE-002` | Leave requests must not overlap with an already-approved request | Error |
| `HR-LVE-003` | Minimum notice days must be met | Error |
| `HR-LVE-004` | Gender-restricted leave types enforce gender matching | Error |
| `HR-LVE-005` | Carry-forward in excess of `carry_forward_days` is forfeited at leave year end | Automatic |
| `HR-LVE-006` | Public holidays falling within a leave request do not consume leave days | Automatic |
| `HR-LVE-007` | Leave cannot be approved for Suspended or Terminated employees | Error |

### 10.7 Statutory Compliance Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `HR-STC-001` | Tax components must be resolved using rates effective on `period_end` | Error |
| `HR-STC-002` | Employer contributions must not reduce employee net pay | Error |
| `HR-STC-003` | Only one active version per `(jurisdiction, component_code)` at any effective date | Error |
| `HR-STC-004` | `basic_salary` must be ≥ jurisdiction minimum wage on pay date | Error |
| `HR-STC-005` | Tenant leave entitlements may not be configured below jurisdiction statutory minimums | Error |

---

## 11. Authorization Model

### Permission Keys

```
# Employee lifecycle
hr.employee.create
hr.employee.read
hr.employee.update
hr.employee.activate
hr.employee.suspend
hr.employee.terminate

# Organisation
hr.organisation.department.manage
hr.organisation.grade.manage
hr.organisation.position.manage

# Shift management
hr.shift.pattern.manage
hr.shift.roster.create
hr.shift.roster.publish
hr.shift.roster.manage
hr.shift.swap.request            # employee self-service
hr.shift.swap.approve            # manager

# Attendance
hr.attendance.read
hr.attendance.manage             # manual override
hr.attendance.biometric.manage   # device registration

# Leave
hr.leave.request.create          # employee self-service
hr.leave.request.approve         # manager / HR
hr.leave.balance.read
hr.leave.type.manage             # HR admin

# Timesheet
hr.timesheet.submit              # employee self-service
hr.timesheet.approve             # manager / HR

# Payroll
hr.payroll.run.create
hr.payroll.run.read
hr.payroll.run.approve
hr.payroll.run.post
hr.payroll.run.reverse
hr.payroll.payslip.read.own      # employee reads own payslip
hr.payroll.payslip.read.all      # HR / Finance

# Accountability
hr.accountability.discrepancy.raise
hr.accountability.discrepancy.review
hr.accountability.discrepancy.confirm
hr.accountability.discrepancy.waive
hr.accountability.recovery.manage

# Statutory & Reporting
hr.statutory.rate.manage         # platform admin only
hr.report.payroll.read
hr.report.statutory.read

# Platform jurisdiction management
platform.jurisdiction.manage     # platform admin only
```

### Default Roles

| Role | Key Permissions |
|---|---|
| `platform_admin` | All `platform.*` permissions; `hr.statutory.rate.manage` |
| `hr_admin` | All `hr.*` permissions |
| `hr_manager` | Employee CRUD, leave approve, timesheet approve, payroll run create/read, accountability review |
| `finance_manager` | Payroll run approve/post/reverse; payslip read all; all report permissions |
| `line_manager` | Leave approve (own team), timesheet approve (own team), roster publish, shift swap approve, employee read |
| `shift_supervisor` | Roster create/manage, attendance read, shift swap approve |
| `employee` | `hr.leave.request.create`, `hr.timesheet.submit`, `hr.payroll.payslip.read.own`, `hr.shift.swap.request` |

### Casbin Model

```ini
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _   # user, role, domain (tenant)

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && r.act == p.act
```

---

## 12. Database Schema

### Core Tables

```sql
-- Employees
CREATE TABLE employees (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    employee_number TEXT NOT NULL,
    first_name      TEXT NOT NULL,
    last_name       TEXT NOT NULL,
    date_of_birth   DATE NOT NULL,
    gender          TEXT NOT NULL CHECK (gender IN ('male','female','other')),
    status          TEXT NOT NULL DEFAULT 'draft'
                        CHECK (status IN ('draft','pending','active','suspended','terminated')),
    department_id   UUID REFERENCES departments(id),
    position_id     UUID REFERENCES positions(id),
    grade_id        UUID REFERENCES grades(id),
    manager_id      UUID REFERENCES employees(id),
    biometric_pin   TEXT,
    join_date       DATE NOT NULL,
    identifiers     JSONB NOT NULL DEFAULT '{}',  -- {"kra_pin": "...", "nssf_number": "..."}
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, employee_number),
    UNIQUE (tenant_id, biometric_pin)
);

-- Shift Patterns
CREATE TABLE shift_patterns (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    code            TEXT NOT NULL,
    name            TEXT NOT NULL,
    start_time      TIME NOT NULL,
    end_time        TIME NOT NULL,
    crosses_midnight BOOLEAN NOT NULL DEFAULT FALSE,
    total_minutes   INT NOT NULL,
    breaks          JSONB NOT NULL DEFAULT '[]',
    days_of_week    INT[] NOT NULL,   -- 0=Sun .. 6=Sat
    is_on_call      BOOLEAN NOT NULL DEFAULT FALSE,
    department_id   UUID REFERENCES departments(id),
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (tenant_id, code)
);

-- Rosters
CREATE TABLE rosters (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    department_id   UUID NOT NULL REFERENCES departments(id),
    period_start    DATE NOT NULL,
    period_end      DATE NOT NULL,
    status          TEXT NOT NULL DEFAULT 'draft'
                        CHECK (status IN ('draft','published','locked','recalled')),
    created_by      UUID NOT NULL,
    published_at    TIMESTAMPTZ,
    locked_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Shift Assignments
CREATE TABLE shift_assignments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL REFERENCES tenants(id),
    roster_id         UUID NOT NULL REFERENCES rosters(id),
    employee_id       UUID NOT NULL REFERENCES employees(id),
    shift_pattern_id  UUID NOT NULL REFERENCES shift_patterns(id),
    date              DATE NOT NULL,
    override_start    TIME,
    override_end      TIME,
    UNIQUE (roster_id, employee_id, date)
);

-- Raw Attendance Events (append-only)
CREATE TABLE attendance_raw_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    employee_id     UUID NOT NULL REFERENCES employees(id),
    source          TEXT NOT NULL CHECK (source IN ('biometric','mobile','api','manual')),
    source_ref      TEXT,
    event_type      TEXT NOT NULL CHECK (event_type IN ('clock_in','clock_out','break_start','break_end')),
    event_time      TIMESTAMPTZ NOT NULL,
    verify_method   TEXT,
    location        JSONB,
    raw_payload     JSONB,
    conflict_status TEXT DEFAULT 'pending' CHECK (conflict_status IN ('pending','accepted','overridden')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Attendance Records (normalised, one per employee per day)
CREATE TABLE attendance_records (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL REFERENCES tenants(id),
    employee_id       UUID NOT NULL REFERENCES employees(id),
    date              DATE NOT NULL,
    source            TEXT NOT NULL,
    scheduled_start   TIMESTAMPTZ,
    scheduled_end     TIMESTAMPTZ,
    actual_clock_in   TIMESTAMPTZ,
    actual_clock_out  TIMESTAMPTZ,
    break_minutes     INT NOT NULL DEFAULT 0,
    overtime_minutes  INT NOT NULL DEFAULT 0,
    absence_type      TEXT NOT NULL DEFAULT 'present'
                          CHECK (absence_type IN ('present','absent','half_day','public_holiday','leave')),
    leave_request_id  UUID REFERENCES leave_requests(id),
    approved_ot       BOOLEAN NOT NULL DEFAULT FALSE,
    notes             TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, employee_id, date)
);

-- Payroll Runs
CREATE TABLE payroll_runs (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        UUID NOT NULL REFERENCES tenants(id),
    period_start     DATE NOT NULL,
    period_end       DATE NOT NULL,
    pay_date         DATE NOT NULL,
    jurisdiction     CHAR(2) NOT NULL,
    currency         CHAR(3) NOT NULL,
    status           TEXT NOT NULL DEFAULT 'draft'
                         CHECK (status IN ('draft','processing','review','approved','posted','reversed','failed')),
    rate_snapshot    JSONB NOT NULL DEFAULT '{}',   -- immutable at processing start
    exchange_rates   JSONB NOT NULL DEFAULT '{}',   -- immutable at run creation
    employee_count   INT NOT NULL DEFAULT 0,
    total_gross      NUMERIC(18,4),
    total_net        NUMERIC(18,4),
    total_tax        NUMERIC(18,4),
    total_employer_cost NUMERIC(18,4),
    created_by       UUID NOT NULL,
    approved_by      UUID,
    posted_at        TIMESTAMPTZ,
    finalized_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Discrepancy Records (Employee Accountability)
CREATE TABLE discrepancy_records (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        UUID NOT NULL REFERENCES tenants(id),
    employee_id      UUID NOT NULL REFERENCES employees(id),
    type             TEXT NOT NULL
                         CHECK (type IN ('cash_short','cash_over','stock_shortage','stock_surplus','damage','equipment_loss')),
    incident_date    DATE NOT NULL,
    shift_id         UUID REFERENCES shift_assignments(id),
    description      TEXT NOT NULL,
    expected_value   NUMERIC(18,4) NOT NULL,
    actual_value     NUMERIC(18,4) NOT NULL,
    variance_amount  NUMERIC(18,4) NOT NULL,
    currency         CHAR(3) NOT NULL,
    status           TEXT NOT NULL DEFAULT 'identified'
                         CHECK (status IN ('identified','under_review','disputed','confirmed','waived','recovered','written_off')),
    recovery_method  TEXT CHECK (recovery_method IN ('payroll_deduction','direct_payment','waived','write_off')),
    evidence_refs    JSONB NOT NULL DEFAULT '[]',
    raised_by        UUID NOT NULL,
    reviewed_by      UUID,
    resolved_at      TIMESTAMPTZ,
    notes            TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Employee Receivables (sub-ledger control)
CREATE TABLE employee_receivables (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        UUID NOT NULL REFERENCES tenants(id),
    employee_id      UUID NOT NULL REFERENCES employees(id),
    discrepancy_id   UUID REFERENCES discrepancy_records(id),
    opening_balance  NUMERIC(18,4) NOT NULL,
    recovered_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
    balance          NUMERIC(18,4) NOT NULL,
    currency         CHAR(3) NOT NULL,
    status           TEXT NOT NULL DEFAULT 'active'
                         CHECK (status IN ('active','settled','written_off','waived')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Jurisdiction Packages (loaded from JSON)
CREATE TABLE jurisdiction_packages (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code             CHAR(2) NOT NULL,
    version_tag      TEXT NOT NULL,
    package_data     JSONB NOT NULL,       -- full JSON jurisdiction definition
    effective_from   DATE NOT NULL,
    effective_to     DATE,
    is_active        BOOLEAN NOT NULL DEFAULT FALSE,
    loaded_by        UUID NOT NULL,        -- platform admin
    loaded_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (code, effective_from)
);
```

### Row-Level Security

```sql
-- Standard tenant isolation on all HR tables
ALTER TABLE employees ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON employees
    USING (tenant_id = current_setting('app.tenant_id')::UUID);

-- Employee payslip self-read
CREATE POLICY payslip_self_read ON payslips
    AS PERMISSIVE FOR SELECT
    USING (
        tenant_id = current_setting('app.tenant_id')::UUID
        AND (
            current_setting('app.permission_hr_payroll_payslip_read_all', TRUE) = 'true'
            OR employee_id = current_setting('app.employee_id', TRUE)::UUID
        )
    );

-- Jurisdiction packages — no tenant_id; accessible to all
CREATE POLICY platform_read_jurisdiction ON jurisdiction_packages
    FOR SELECT USING (TRUE);

CREATE POLICY platform_write_jurisdiction ON jurisdiction_packages
    FOR ALL USING (current_setting('app.role') = 'platform_admin');
```

---

## 13. Workflow Orchestration (Temporal)

### Payroll Run Workflow

```go
func PayrollRunWorkflow(ctx workflow.Context, input PayrollRunInput) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 3},
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    // 1. Snapshot jurisdiction rates
    if err := workflow.ExecuteActivity(ctx, SnapshotJurisdictionRatesActivity, input).Get(ctx, nil); err != nil {
        return err
    }

    // 2. Load eligible employees
    var employees []EmployeePayrollInput
    if err := workflow.ExecuteActivity(ctx, LoadEligibleEmployeesActivity, input).Get(ctx, &employees); err != nil {
        return err
    }

    // 3. Process in parallel batches of 50
    var futures []workflow.Future
    for i := 0; i < len(employees); i += 50 {
        batch := employees[i:min(i+50, len(employees))]
        futures = append(futures, workflow.ExecuteActivity(ctx, ProcessPayrollBatchActivity, input.RunID, batch))
    }
    for _, f := range futures {
        if err := f.Get(ctx, nil); err != nil {
            return err
        }
    }

    // 4. Aggregate totals
    if err := workflow.ExecuteActivity(ctx, AggregateRunTotalsActivity, input.RunID).Get(ctx, nil); err != nil {
        return err
    }

    // 5. Transition to review
    return workflow.ExecuteActivity(ctx, TransitionRunStatusActivity, input.RunID, RunStatusReview).Get(ctx, nil)
}
```

### Roster Publish & Lock Workflow

```go
// Triggered when a roster is published.
// Schedules a delayed lock at (period_start - lock_lead_days).
func RosterLifecycleWorkflow(ctx workflow.Context, input RosterInput) error {
    lockAt := input.PeriodStart.AddDate(0, 0, -input.LockLeadDays)
    _ = workflow.Sleep(ctx, time.Until(lockAt))

    ao := workflow.ActivityOptions{StartToCloseTimeout: 5 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)

    // Lock roster and create attendance records from shift assignments
    return workflow.ExecuteActivity(ctx, LockRosterAndCreateAttendanceActivity, input.RosterID).Get(ctx, nil)
}
```

### Attendance Reconciliation Cron

```go
// Runs nightly at 01:00. Resolves multi-source conflicts for the previous day.
func AttendanceReconciliationWorkflow(ctx workflow.Context) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)
    yesterday := time.Now().AddDate(0, 0, -1)
    return workflow.ExecuteActivity(ctx, ReconcileAttendanceActivity, yesterday).Get(ctx, nil)
}
```

### Leave Accrual Cron

```go
// Runs on the last day of each month (Temporal deduplication via workflow ID).
func LeaveAccrualCronWorkflow(ctx workflow.Context) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)
    return workflow.ExecuteActivity(ctx, AccrueMonthlyLeaveActivity, time.Now()).Get(ctx, nil)
}
```

### Offboarding Workflow

```go
func OffboardingWorkflow(ctx workflow.Context, input OffboardingInput) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 15 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)

    steps := []struct{ name string; fn interface{} }{
        {"ComputeTerminalBenefits",    ComputeTerminalBenefitsActivity},
        {"CreateFinalPayrollRun",      CreateFinalPayrollRunActivity},
        {"ClearOutstandingDeductions", ClearOutstandingDeductionsActivity},
        {"SettleDiscrepancies",        SettleDiscrepanciesActivity},
        {"RevokeSystemAccess",         RevokeSystemAccessActivity},
        {"IssueServiceCertificate",    IssueServiceCertificateActivity},
        {"ArchiveEmployeeRecord",      ArchiveEmployeeRecordActivity},
    }
    for _, step := range steps {
        if err := workflow.ExecuteActivity(ctx, step.fn, input).Get(ctx, nil); err != nil {
            return fmt.Errorf("offboarding step %s: %w", step.name, err)
        }
    }
    return nil
}
```

---

## 14. Integration Points

### Finance Module (Internal)

| Event | Direction | Payload |
|---|---|---|
| `PayrollPosted` | HR → Finance | Run ID, period, `[]JournalInstruction`, total employer cost |
| `PayrollReversed` | HR → Finance | Original run ID, reversal run ID, mirror instructions |
| `DiscrepancyConfirmed` | HR → Finance | Discrepancy ID, employee ID, GL instruction (Dr Employee Receivable, Cr Variance) |
| `DiscrepancyWaived` | HR → Finance | GL write-off instruction |
| `TerminalBenefitsComputed` | HR → Finance | Employee ID, breakdown, GL instructions |
| `JournalCreated` | Finance → HR | Confirms successful posting; HR marks run `posted` |

### Inventory / Wetstock Module (Internal)

For forecourt and retail operations, the Inventory module emits variance events that the HR module can consume to auto-raise discrepancy records:

| Event | Direction | Description |
|---|---|---|
| `ShiftVarianceReported` | Inventory → HR | Pump/till variance with linked shift ID; HR creates a `DiscrepancyRecord` in `under_review` status |
| `DiscrepancyAcknowledged` | HR → Inventory | HR confirms discrepancy; Inventory closes its variance record |

### IAM Module (Internal)

| Event | Direction | Description |
|---|---|---|
| `EmployeeActivated` | HR → IAM | Provision system user account for employee |
| `EmployeeSuspended` | HR → IAM | Disable user account |
| `EmployeeTerminated` | HR → IAM | Revoke all sessions; delete or archive account |

### KRA eTIMS (Kenya)

```go
type KRAeTIMSClient interface {
    SubmitPAYEReturn(ctx context.Context, period time.Month, year int, entries []PAYEEntry) (SubmissionRef, error)
    ValidateTaxID(ctx context.Context, pin string) (TaxIDValidationResult, error)
    FetchComplianceCertificate(ctx context.Context, pin string) (ComplianceCert, error)
}
```

The platform ships a built-in eTIMS adapter. Other jurisdictions can add their own filing adapters by implementing the `StatutoryFilingAdapter` interface:

```go
type StatutoryFilingAdapter interface {
    JurisdictionCode() string
    SubmitReturn(ctx context.Context, component string, period DateRange, data interface{}) (SubmissionRef, error)
    ValidateIdentifier(ctx context.Context, identifierCode, value string) (ValidationResult, error)
}
```

### Bank Payment Files

| Format | Supported Institutions |
|---|---|
| ISO 20022 pain.001 XML | Stanbic, KCB, Absa, Standard Chartered (most markets) |
| NIBSS bulk transfer | Nigerian banks |
| MT103 SWIFT | International / correspondent banking |
| ACB (Authenticated Collections Batch) | South Africa — SARS / commercial banks |
| Generic CSV | Configurable column mapping for any bank |

---

## 15. Configuration Flags

Stored per tenant in `tenant_config.hr`. All flags have defaults that can be overridden by the tenant administrator via the Settings UI.

### Payroll

| Flag | Default | Description |
|---|---|---|
| `require_timesheet_approval` | `true` | Block payroll for employees with unapproved timesheets |
| `enforce_grade_salary_band` | `false` | Hard-block salary outside grade band (default is warning) |
| `allow_negative_leave_balance` | `false` | Permit leave beyond available balance |
| `multi_currency_allowances` | `false` | Allow allowance entries in non-base currency |
| `require_separation_workflow` | `true` | Enforce Temporal offboarding on termination |
| `payslip_self_service` | `true` | Employees can download own payslips |

### Attendance & Shifts

| Flag | Default | Description |
|---|---|---|
| `mobile_attendance_enabled` | `true` | Enable mobile clock-in/out |
| `require_geofence_mobile` | `true` | GPS must be within work location radius |
| `require_selfie_clockin` | `false` | Selfie capture required on mobile clock-in |
| `biometric_attendance_enabled` | `false` | Enable biometric device integration |
| `attendance_conflict_tolerance_minutes` | `15` | Threshold for flagging multi-source conflicts |
| `roster_lock_lead_days` | `2` | Days before period start at which roster auto-locks |
| `min_rest_hours_between_shifts` | `11` | Minimum rest period; can exceed but not go below jurisdiction minimum |
| `on_call_compensation_method` | `hourly` | `flat_fee` \| `hourly` \| `percentage_of_basic` |
| `require_manager_ot_approval` | `true` | Overtime must be pre-approved to be paid |

### Accountability

| Flag | Default | Description |
|---|---|---|
| `accountability_enabled` | `true` | Enable discrepancy tracking sub-module |
| `cash_over_credit_employee` | `false` | Credit cash surpluses to employee against future shortfalls |
| `max_recovery_percentage_of_net` | `50` | Max % of net pay that can be recovered in a single period |
| `auto_raise_from_inventory_variance` | `false` | Auto-create discrepancy records from Inventory module variance events |

### Statutory (jurisdiction-level, platform-managed)

These flags override individual components within the active jurisdiction package for a tenant, where the jurisdiction allows variation:

| Flag | Default | Description |
|---|---|---|
| `nssf_use_tier_system` | `true` | Kenya: use NSSF Act 2013 tier rates vs legacy flat rate |
| `shif_enabled` | `true` | Kenya: SHIF (2.75%) instead of old NHIF bands |
| `affordable_housing_levy_enabled` | `true` | Kenya: AHL 1.5% employee + 1.5% employer |
| `pension_scheme_type` | `nssf` | `nssf` \| `occupational` \| `provident_fund` |

---

## 16. Future Roadmap Considerations

### 16.1 Advanced Roster Optimisation

A constraint-satisfaction or linear programming solver (e.g. Google OR-Tools via a sidecar service) could automate roster generation given:
- Employee preferences and availability
- Minimum staffing levels per shift and department
- Labour cost optimisation constraints
- Rest period and overtime cap constraints

The `ShiftPattern` and `Roster` domain model supports this without schema changes.

### 16.2 Employee Self-Service Portal

The mobile and web apps can expose an employee-facing portal for:
- Leave request submission and balance viewing
- Payslip downloads
- Shift viewing and swap requests
- Discrepancy disputes
- Update personal and bank details (subject to approval)

### 16.3 Performance-Linked Pay

A future `PerformanceLink` configuration on `AllowanceType` would allow an allowance to be computed against a performance score sourced from a future Performance module, using the `pkg/condition` expression engine to evaluate the formula.

### 16.4 Multi-Entity Payroll

Where a tenant operates multiple legal entities (e.g. Shell Maanzoni Service Station as an entity within a larger group), the model supports this by allowing multiple `Organisation` records under one tenant. Each organisation can have its own jurisdiction and currency, with shared employee pools where employees are contracted to one entity but may work across others (inter-entity secondment).

### 16.5 AI-Assisted Compliance Monitoring

The platform could flag compliance risks by analysing:
- Employees approaching statutory overtime limits
- Departments where shift patterns violate rest period rules more than N% of the time
- Discrepancy patterns suggesting systemic issues (specific cashiers, specific shifts)

### 16.6 Filing Automation

As `StatutoryFilingAdapter` implementations mature, the module can automate:
- Monthly PAYE submissions to tax authorities
- NSSF / pension remittance scheduling
- Annual reconciliation reports

Each adapter is independently deployable as a platform plugin, allowing new jurisdictions' filing integrations to be added without core module changes.

---

*This document is the authoritative architecture specification for the Awo ERP HR & Payroll module. Jurisdiction-specific statutory rates and legal references should always be verified against the official gazette of the relevant authority before production use. Rate data in the jurisdiction package registry takes precedence over any examples in this document.*
