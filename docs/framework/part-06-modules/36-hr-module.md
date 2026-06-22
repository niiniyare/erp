---
title: "Chapter 36: HR Module"
part: "Part VI — Built-In Modules"
chapter: 36
section: "36-hr-module"
related:
  - "[Chapter 33: Finance Module](33-finance-module.md)"
  - "[Chapter 30: Signals, Queries](../part-05-workflow/30-signals-queries.md)"
---

# Chapter 36: HR Module

The HR module covers employee records, leave management, attendance, payroll (with Kenya-specific PAYE/NHIF/NSSF calculations), and disciplinary workflows.

---

## 36.1. Employee and Organisation

### 36.1.1. Employee Entity

```go
type Employee struct {
    EmployeeNumber string     // EMP-0001
    FirstName      string
    LastName       string
    NationalID     string     // Kenya National ID
    KraPin         string     // KRA Personal PIN for PAYE
    NhifNumber     string
    NssfNumber     string
    DateOfBirth    time.Time
    DateOfJoining  time.Time
    DateOfLeaving  *time.Time
    DepartmentID   uuid.UUID
    DesignationID  uuid.UUID
    ReportsToID    *uuid.UUID // direct manager
    EmploymentType string     // permanent | contract | casual
    PayrollGroup   string     // monthly | weekly | casual-daily
    BankAccount    EncryptedBankDetails
    Status         string     // active | on_leave | terminated
}
```

### 36.1.2. Department Hierarchy

Departments use the materialised path pattern. An employee's department determines:
- Which GL cost centre their salary is posted to
- Which manager approves their leave requests
- Their access to department-specific reports

---

## 36.2. Leave Management

### 36.2.1. Leave Types

```go
// Kenyan Employment Act 2007 minimum leave entitlements:
// Annual Leave: 21 days per year
// Sick Leave: 7 days fully paid, 7 days half-pay per year
// Maternity: 3 months fully paid
// Paternity: 2 weeks (not statutory, employer discretion)
// Compassionate: 3 days (death of immediate family)

type LeaveType struct {
    Name            string
    CarryForward    bool           // unused leave carries to next year
    MaxCarryDays    int
    IsPayable       bool           // paid leave
    RequiresApproval bool
    EligibilityDays int            // employment days before eligible
}
```

### 36.2.2. Leave Allocation — Per Year, Per Employee

Leave is allocated at the start of each year (or pro-rated for new joiners):

```go
type LeaveAllocation struct {
    EmployeeID    uuid.UUID
    LeaveTypeID   uuid.UUID
    FiscalYear    string    // "2025-26"
    AllocatedDays decimal.Decimal
    UsedDays      decimal.Decimal
    BalanceDays   decimal.Decimal  // computed: allocated - used
    CarriedDays   decimal.Decimal  // brought forward from last year
}
```

### 36.2.3. Leave Request Workflow

```go
func LeaveApprovalWorkflow(ctx workflow.Context, params LeaveParams) error {
    // Step 1: Validate leave balance
    workflow.ExecuteActivity(ctx, activities.ValidateLeaveBalance, params)

    // Step 2: Notify manager
    workflow.ExecuteActivity(ctx, activities.NotifyLeaveApprover, params)

    // Step 3: Wait for manager decision (5 business days timeout)
    decision, _ := waitForApprovalGate(ctx, ApprovalGate{
        Stage:         "manager",
        TimeoutHours:  5 * 24,
        OnTimeout:     "escalate",
    })

    if decision.Action == "approved" {
        workflow.ExecuteActivity(ctx, activities.DeductLeaveBalance, params)
        workflow.ExecuteActivity(ctx, activities.UpdateLeaveStatus, params, "approved")
        workflow.ExecuteActivity(ctx, activities.NotifyEmployeeApproved, params)
    } else {
        workflow.ExecuteActivity(ctx, activities.UpdateLeaveStatus, params, "rejected")
        workflow.ExecuteActivity(ctx, activities.NotifyEmployeeRejected, params, decision.Reason)
    }
    return nil
}
```

---

## 36.3. Attendance and Payroll

### 36.3.1. Attendance Record Entity

```go
type Attendance struct {
    EmployeeID    uuid.UUID
    AttendanceDate time.Time
    InTime        *time.Time
    OutTime       *time.Time
    WorkingHours  decimal.Decimal
    Status        string     // present | absent | half_day | on_leave | holiday
    LeaveTypeID   *uuid.UUID // if on leave
}
```

### 36.3.2. Kenya PAYE Structure

Kenya PAYE (Pay As You Earn) uses graduated tax bands (2025 rates):

```go
type PAYEBand struct {
    MinMonthly   decimal.Decimal
    MaxMonthly   decimal.Decimal  // nil for last band
    Rate         decimal.Decimal  // percentage
}

var KenyaPAYEBands2025 = []PAYEBand{
    {Min: 0, Max: 24000, Rate: 10},
    {Min: 24001, Max: 32333, Rate: 25},
    {Min: 32334, Max: 500000, Rate: 30},
    {Min: 500001, Max: 800000, Rate: 32.5},
    {Min: 800001, Max: nil, Rate: 35},
}

// Personal relief: KES 2,400/month
// NHIF relief: NHIF contribution is a tax relief
// Affordable Housing Levy: 1.5% of gross pay (employer + employee each)
```

### 36.3.3. Payroll Run Workflow

```go
func PayrollRunWorkflow(ctx workflow.Context, params PayrollRunParams) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)

    // 1. Fetch all employees in payroll group
    var employees []Employee
    workflow.ExecuteActivity(ctx, activities.FetchPayrollEmployees, params).Get(ctx, &employees)

    // 2. Process each employee
    for _, emp := range employees {
        workflow.ExecuteActivity(ctx, activities.ComputePayslip,
            PayslipInput{
                EmployeeID:   emp.ID,
                PayrollMonth: params.PayrollMonth,
                AttendanceSummary: params.AttendanceSummary[emp.ID],
            })
    }

    // 3. Generate payroll journal entry
    workflow.ExecuteActivity(ctx, activities.PostPayrollJournalEntry, params.RunID)

    // 4. Generate PAYE remittance report for KRA
    workflow.ExecuteActivity(ctx, activities.GeneratePAYEReport, params.RunID)

    // 5. Generate NHIF and NSSF remittance reports
    workflow.ExecuteActivity(ctx, activities.GenerateStatutoryReports, params.RunID)

    return nil
}
```

### 36.3.4. Payslip Entity

```go
type Payslip struct {
    EmployeeID       uuid.UUID
    PayrollMonth     string          // "2025-07"
    GrossSalary      decimal.Decimal
    // Deductions
    PAYE             decimal.Decimal
    NHIFEmployee     decimal.Decimal // employee's share
    NSSFEmployee     decimal.Decimal // employee's share (Tier I + Tier II)
    AHL              decimal.Decimal // Affordable Housing Levy (1.5%)
    // Employer contributions (not deducted from employee, shown for costing)
    NHIFEmployer     decimal.Decimal
    NSSFEmployer     decimal.Decimal
    // Net
    NetPay           decimal.Decimal
    // GL posting reference
    JournalEntryID   uuid.UUID
}
```

---

## 36.4. Disciplinary Workflow

### 36.4.1. Warning Entity

```go
type DisciplinaryWarning struct {
    EmployeeID    uuid.UUID
    WarningLevel  string    // "verbal" | "written_1" | "written_2" | "final"
    Incident      string
    IssuedDate    time.Time
    IssuedBy      uuid.UUID
    AcknowledgedAt *time.Time
}
```

### 36.4.2. Show-Cause Notice Workflow

When a disciplinary matter requires a formal response:

```go
func ShowCauseWorkflow(ctx workflow.Context, params ShowCauseParams) error {
    // Issue the show-cause notice
    workflow.ExecuteActivity(ctx, activities.IssueShowCauseNotice, params)

    // Wait for employee's written response (7 days per Employment Act)
    responseCh := workflow.GetSignalChannel(ctx, "employee-response")
    var response EmployeeResponse

    selector := workflow.NewSelector(ctx)
    selector.AddReceive(responseCh, func(c workflow.ReceiveChannel, _bool) {
        c.Receive(ctx, &response)
    })
    selector.AddFuture(workflow.NewTimer(ctx, 7*24*time.Hour), func(f workflow.Future) {
        response = EmployeeResponse{Type: "no_response"}
    })
    selector.Select(ctx)

    // HR review with response
    decision, _ := waitForApprovalGate(ctx, ApprovalGate{
        Stage:        "hr_review",
        RequiredRoles: []string{"hr_manager"},
        TimeoutHours: 48,
    })

    workflow.ExecuteActivity(ctx, activities.RecordDisciplinaryOutcome, params, decision, response)
    return nil
}
```

### 36.4.4. Termination Workflow

The termination workflow ensures proper computation of final dues:
1. Calculate outstanding leave balance payout
2. Calculate notice period payment or deduction
3. Calculate any gratuity (service of >5 years)
4. Generate termination letter
5. Post final payroll GL entries
6. Deactivate employee's system access (signals IAM service)
7. Generate statutory notification records (NHIF exit, NSSF transfer)
