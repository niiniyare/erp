---
title: "Projects Module Patterns"
id: mod-010
status: accepted
category: GUIDE
stability: STABLE
audience: [module-authors]
since: "1.0"
normative-level: informative
related:
  - "[Finance Patterns](finance-patterns.md)"
  - "[HR Patterns](hr-patterns.md)"
  - "[Business Module Catalog](business-modules.md)"
  - "[Awo Glossary](../GLOSSARY.md)"
---

# Projects Module Patterns

**MOD-010 | Status: Accepted | Stability: Stable**

Patterns for project management: project lifecycle, task tracking, time logging, expense allocation, and project billing.

---

## 1. Core Entities

| Entity | Type | Purpose |
|---|---|---|
| `project` | System | Project header: client, budget, dates, status |
| `project_task` | System | Work item with assignee, estimate, actual hours |
| `project_timesheet` | System | Time log entries (hours × billing rate) |
| `project_expense` | Custom | Out-of-pocket expenses charged to a project |
| `project_milestone` | Custom | Key dates and completion criteria |

System entities: timesheets feed revenue recognition and billing — financial integrity required.

---

## 2. Project Entity

```go
var ProjectDefinition = definition.SystemDefinition{
    Name:   "project",
    Module: "projects",
    Fields: []definition.FieldDef{
        {Name: "code",           Type: definition.FieldNamingSeries, Series: "PROJ-{YYYY}-{SEQ:4}"},
        {Name: "name",           Type: definition.FieldData, Required: true, Searchable: true},
        {Name: "client",         Type: definition.FieldLink, LinkTarget: "crm_customer", Required: true},
        {Name: "status",         Type: definition.FieldSelect,
            Options: []string{"Planning", "Active", "On Hold", "Completed", "Cancelled"}, Default: "Planning"},
        {Name: "start_date",     Type: definition.FieldDate, Required: true},
        {Name: "end_date",       Type: definition.FieldDate},
        {Name: "budget_kes",     Type: definition.FieldCurrency},
        {Name: "billed_kes",     Type: definition.FieldCurrency},  // computed: sum of invoiced timesheets
        {Name: "billing_type",   Type: definition.FieldSelect,
            Options: []string{"Fixed Price", "Time & Materials", "Retainer"}},
        {Name: "project_manager", Type: definition.FieldLink, LinkTarget: "iam_user"},
        {Name: "description",    Type: definition.FieldLongText},
    },
    Edges: []definition.EdgeDef{
        {Name: "tasks",       Target: "project_task",      Type: definition.EdgeOneToMany, CascadeDelete: true},
        {Name: "timesheets",  Target: "project_timesheet",  Type: definition.EdgeOneToMany},
        {Name: "milestones",  Target: "project_milestone",  Type: definition.EdgeOneToMany, CascadeDelete: true},
    },
    Permissions: definition.PermissionSet{
        Create: []string{"role:projects.manager", "role:tenant.admin"},
        Read:   []string{"role:projects.member", "role:projects.manager", "role:tenant.admin"},
        Write:  []string{"role:projects.manager", "role:tenant.admin"},
        Delete: []string{"role:tenant.admin"},
    },
    Policy: definition.PolicyFunc(func(ctx context.Context) definition.Filter {
        actor := session.ActorFromContext(ctx)
        if actor.HasRole("role:projects.manager") || actor.HasRole("role:tenant.admin") {
            return filter.All()  // managers see all projects
        }
        // Members see only projects they're assigned to (via tasks)
        return filter.In("id", subquery.ProjectsForMember(actor.UserID))
    }),
}
```

---

## 3. Task Entity with Progress Tracking

```go
var ProjectTaskDefinition = definition.SystemDefinition{
    Name:   "project_task",
    Module: "projects",
    Fields: []definition.FieldDef{
        {Name: "project",       Type: definition.FieldLink, LinkTarget: "project",
            Required: true, Immutable: true},
        {Name: "title",         Type: definition.FieldData, Required: true, Searchable: true},
        {Name: "assignee",      Type: definition.FieldLink, LinkTarget: "iam_user"},
        {Name: "status",        Type: definition.FieldSelect,
            Options: []string{"Todo", "In Progress", "Review", "Done"}, Default: "Todo"},
        {Name: "priority",      Type: definition.FieldSelect,
            Options: []string{"Low", "Medium", "High", "Critical"}, Default: "Medium"},
        {Name: "estimated_hours", Type: definition.FieldFloat},
        {Name: "logged_hours",  Type: definition.FieldFloat},  // computed from timesheets
        {Name: "due_date",      Type: definition.FieldDate},
        {Name: "completed_at",  Type: definition.FieldDateTime},
        {Name: "description",   Type: definition.FieldLongText},
    },
    Hooks: definition.HookSet{
        BeforeSave: []definition.BeforeSaveHook{&TaskCompletionHook{}},
    },
}

type TaskCompletionHook struct{}

func (h *TaskCompletionHook) BeforeSave(ctx context.Context, record *definition.EntityRecord, isUpdate bool) error {
    if !isUpdate {
        return nil
    }
    status, _ := record.Fields["status"].(string)
    if status == "Done" {
        if record.Fields["completed_at"] == nil {
            record.Fields["completed_at"] = time.Now().UTC()
        }
    } else {
        record.Fields["completed_at"] = nil  // clear if moved back
    }
    return nil
}
```

---

## 4. Timesheet Logging

```go
var TimesheetDefinition = definition.SystemDefinition{
    Name:   "project_timesheet",
    Module: "projects",
    Fields: []definition.FieldDef{
        {Name: "project",      Type: definition.FieldLink, LinkTarget: "project",
            Required: true, Immutable: true},
        {Name: "task",         Type: definition.FieldLink, LinkTarget: "project_task"},
        {Name: "employee",     Type: definition.FieldLink, LinkTarget: "hr_employee",
            Required: true, Immutable: true},
        {Name: "date",         Type: definition.FieldDate, Required: true},
        {Name: "hours",        Type: definition.FieldFloat, Required: true},
        {Name: "description",  Type: definition.FieldSmallText, Required: true},
        {Name: "billing_rate", Type: definition.FieldCurrency},  // KES/hour at time of logging
        {Name: "amount_kes",   Type: definition.FieldCurrency},  // hours × billing_rate
        {Name: "billable",     Type: definition.FieldBool, Default: true},
        {Name: "invoiced",     Type: definition.FieldBool, Default: false},
        {Name: "invoice",      Type: definition.FieldLink, LinkTarget: "finance_invoice"},
    },
    Hooks: definition.HookSet{
        BeforeCreate: []definition.BeforeCreateHook{&TimesheetRateHook{}},
    },
    Policy: definition.PolicyFunc(func(ctx context.Context) definition.Filter {
        actor := session.ActorFromContext(ctx)
        if actor.HasRole("role:projects.manager") || actor.HasRole("role:tenant.admin") {
            return filter.All()
        }
        // Members see only their own timesheets
        return filter.Eq("employee", actor.LinkedEmployeeID)
    }),
}

// TimesheetRateHook auto-fills billing rate from employee's current rate
type TimesheetRateHook struct {
    EmployeeRepo definition.EntityRepository[HREmployee]
}

func (h *TimesheetRateHook) BeforeCreate(ctx context.Context, record *definition.EntityRecord) error {
    hours, _ := record.Fields["hours"].(float64)
    if hours <= 0 {
        return &definition.ValidationError{
            Fields: map[string]string{"hours": "Hours must be greater than zero"},
        }
    }

    // Fill billing rate from employee record if not provided
    if record.Fields["billing_rate"] == nil {
        empID, _ := record.Fields["employee"].(uuid.UUID)
        emp, err := h.EmployeeRepo.Get(ctx, empID)
        if err != nil {
            return fmt.Errorf("TimesheetRateHook.BeforeCreate: get employee: %w", err)
        }
        rate, _ := emp.Fields["billing_rate_kes"].(decimal.Decimal)
        record.Fields["billing_rate"] = rate
    }

    rate, _ := record.Fields["billing_rate"].(decimal.Decimal)
    hoursDecimal := decimal.NewFromFloat(hours)
    record.Fields["amount_kes"] = rate.Mul(hoursDecimal)

    return nil
}
```

---

## 5. Project Billing Action

For Time & Materials projects, billing creates a Finance invoice from unbilled timesheets:

```go
Actions: []definition.ActionDef{
    {
        Name:        "bill",
        Label:       "Create Invoice from Timesheets",
        Permission:  "role:projects.manager",
        HandlerFunc: BillProjectAction,
    },
},
```

```go
func BillProjectAction(ctx context.Context, action definition.ActionContext) (*definition.ActionResult, error) {
    project, err := action.Repo.Get(ctx, action.RecordID)
    if err != nil {
        return nil, fmt.Errorf("BillProjectAction: get project: %w", err)
    }

    billingType, _ := project.Fields["billing_type"].(string)
    if billingType != "Time & Materials" {
        return nil, &definition.BusinessError{
            Code:    "projects.not_t_and_m",
            Message: "Billing action is only for Time & Materials projects",
            Status:  400,
        }
    }

    // Fetch unbilled billable timesheets
    timesheets, _, err := action.Services.TimesheetRepo.Query(ctx, filter.And(
        filter.Eq("project", action.RecordID),
        filter.Eq("billable", true),
        filter.Eq("invoiced", false),
    ))
    if err != nil {
        return nil, fmt.Errorf("BillProjectAction: fetch timesheets: %w", err)
    }
    if len(timesheets) == 0 {
        return nil, &definition.BusinessError{
            Code:    "projects.no_unbilled_timesheets",
            Message: "No unbilled timesheets found for this project",
            Status:  400,
        }
    }

    // Sum total
    total := decimal.Zero
    for _, ts := range timesheets {
        amount, _ := ts.Fields["amount_kes"].(decimal.Decimal)
        total = total.Add(amount)
    }

    // Create Finance invoice via FinanceService interface (cross-module boundary)
    invoiceID, err := action.Services.FinanceService.CreateProjectInvoice(ctx, finance.ProjectInvoiceInput{
        CustomerID:  project.Fields["client"].(uuid.UUID),
        ProjectID:   action.RecordID,
        TotalKES:    total,
        Timesheets:  timesheets,
    })
    if err != nil {
        return nil, fmt.Errorf("BillProjectAction: create invoice: %w", err)
    }

    // Mark timesheets as invoiced
    _, err = action.Services.TimesheetRepo.BulkUpdate(ctx, filter.And(
        filter.Eq("project", action.RecordID),
        filter.Eq("billable", true),
        filter.Eq("invoiced", false),
    ), definition.Patch{"invoiced": true, "invoice": invoiceID})
    if err != nil {
        return nil, fmt.Errorf("BillProjectAction: mark invoiced: %w", err)
    }

    return &definition.ActionResult{
        Message: fmt.Sprintf("Invoice created for KES %.2f", total),
        Data:    map[string]any{"invoice_id": invoiceID},
    }, nil
}
```

---

## 6. Project Dashboard SDUI

```go
func BuildProjectDashboard(ctx context.Context, actor session.Actor) ([]byte, error) {
    psc := sdui.NewPageSchemaContext(ctx, actor)

    return psc.Marshal(amis.Page(amis.PageProps{
        Title: "Project Overview",
        Body: amis.Grid([]amis.Panel{
            // Status breakdown
            amis.Panel("Project Status", amis.Chart(amis.ChartProps{
                Type: "pie",
                API:  "GET /api/v1/entities/project/aggregate?group_by=status",
            })),
            // My tasks
            amis.Panel("My Tasks", amis.CRUD(amis.CRUDProps{
                API: "GET /api/v1/entities/project_task?assignee=__current_user__&status=In Progress,Todo",
                Columns: []amis.Column{
                    {Name: "project.name", Label: "Project"},
                    {Name: "title", Label: "Task"},
                    {Name: "priority", Label: "Priority"},
                    {Name: "due_date", Label: "Due", Type: "date"},
                },
            })),
            // Hours logged this month
            amis.Panel("Hours This Month", amis.Stat(amis.StatProps{
                API:   "GET /api/v1/entities/project_timesheet/aggregate?op=sum&field=hours&period=this_month",
                Label: "Total Hours",
            })),
        }),
    }))
}
```

---

## Related Documents

- [Finance Patterns](finance-patterns.md) — invoice creation from project billing
- [HR Patterns](hr-patterns.md) — employee-user link, billing rates
- [Business Module Catalog](business-modules.md) — projects module dependencies
- [Dashboard Patterns](../08-sdui/dashboard-patterns.md) — stat cards and chart patterns
