# Month Range Control Component

**FILE PURPOSE**: Month range selection for period-based reporting and planning  
**SCOPE**: Month range picking, fiscal periods, reporting cycles, and duration-based planning  
**TARGET AUDIENCE**: Developers implementing period reporting, budget planning, and monthly range filters

## 📋 Component Overview

Month Range Control provides month range selection with start/end month picking, duration limits, shortcuts, and validation. Essential for period-based reporting and planning in ERP systems.

### Schema Reference
- **Primary Schema**: `MonthRangeControlSchema.json`
- **Related Schemas**: `FormHorizontal.json`, `ShortCuts.json`
- **Base Interface**: Form input control for month range selection

## Basic Usage

```json
{
    "type": "input-month-range",
    "name": "reporting_period",
    "label": "Reporting Period",
    "placeholder": "Select month range..."
}
```

## Go Type Definition

```go
type MonthRangeControlProps struct {
    Type              string              `json:"type"`
    Name              string              `json:"name"`
    Label             interface{}         `json:"label"`
    Placeholder       string              `json:"placeholder"`
    StartPlaceholder  string              `json:"startPlaceholder"`
    EndPlaceholder    string              `json:"endPlaceholder"`
    Format            string              `json:"format"`          // Storage format
    ValueFormat       string              `json:"valueFormat"`     // Alternative format
    InputFormat       string              `json:"inputFormat"`     // Display format (legacy)
    DisplayFormat     string              `json:"displayFormat"`   // Display format (new)
    Delimiter         string              `json:"delimiter"`       // Value separator
    JoinValues        bool                `json:"joinValues"`      // Join as string
    MinDate           string              `json:"minDate"`         // Minimum date
    MaxDate           string              `json:"maxDate"`         // Maximum date
    MinDuration       string              `json:"minDuration"`     // Minimum span
    MaxDuration       string              `json:"maxDuration"`     // Maximum span
    BorderMode        string              `json:"borderMode"`      // "full", "half", "none"
    Embed             bool                `json:"embed"`           // Inline mode
    Shortcuts         interface{}         `json:"shortcuts"`       // Range shortcuts
    Ranges            interface{}         `json:"ranges"`          // Same as shortcuts
    Animation         bool                `json:"animation"`       // Cursor animation
}
```

## Essential Variants

### Basic Month Range
```json
{
    "type": "input-month-range",
    "name": "budget_period",
    "label": "Budget Period",
    "startPlaceholder": "Start month",
    "endPlaceholder": "End month",
    "format": "YYYY-MM",
    "displayFormat": "MMM YYYY",
    "delimiter": " to ",
    "required": true
}
```

### Quarterly Planning Range
```json
{
    "type": "input-month-range",
    "name": "planning_period",
    "label": "Planning Period",
    "format": "YYYY-MM",
    "displayFormat": "MMMM YYYY",
    "minDuration": "3months",
    "maxDuration": "12months",
    "shortcuts": [
        {"label": "This Quarter", "value": "thisQuarter"},
        {"label": "Next Quarter", "value": "nextQuarter"},
        {"label": "This Year", "value": "thisYear"}
    ]
}
```

### Fiscal Period Range
```json
{
    "type": "input-month-range",
    "name": "fiscal_period",
    "label": "Fiscal Period",
    "format": "YYYY-MM",
    "displayFormat": "MMM YYYY",
    "minDate": "2020-01",
    "maxDate": "+24months",
    "shortcuts": ["lastYear", "thisYear", "nextYear"],
    "joinValues": true,
    "delimiter": ","
}
```

## Real-World Use Cases

### Budget Period Planning
```json
{
    "type": "input-month-range",
    "name": "budget_period",
    "label": "Budget Planning Period",
    "startPlaceholder": "Budget start month",
    "endPlaceholder": "Budget end month",
    "format": "YYYY-MM",
    "displayFormat": "MMMM YYYY",
    "delimiter": " to ",
    "minDuration": "1month",
    "maxDuration": "12months",
    "shortcuts": [
        {"label": "Current Quarter", "value": "thisQuarter"},
        {"label": "Next Quarter", "value": "nextQuarter"},
        {"label": "Full Year", "value": "thisYear"},
        {"label": "Next Year", "value": "nextYear"}
    ],
    "required": true,
    "validations": {
        "isDateTimeAfter": ["${currentMonth}", "month"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Budget period must start from current month or later"
    }
}
```

### Financial Reporting Range
```json
{
    "type": "input-month-range",
    "name": "reporting_range",
    "label": "Reporting Period",
    "startPlaceholder": "Report start",
    "endPlaceholder": "Report end",
    "format": "YYYY-MM",
    "displayFormat": "MMM YYYY",
    "delimiter": " - ",
    "shortcuts": [
        {"label": "Last Quarter", "value": "lastQuarter"},
        {"label": "This Quarter", "value": "thisQuarter"},
        {"label": "Last 6 Months", "value": "last6months"},
        {"label": "Year to Date", "value": "yearToDate"}
    ],
    "maxDuration": "24months",
    "joinValues": true,
    "clearable": true
}
```

### Payroll Period Selection
```json
{
    "type": "input-month-range",
    "name": "payroll_period",
    "label": "Payroll Processing Period",
    "startPlaceholder": "Payroll start month",
    "endPlaceholder": "Payroll end month",
    "format": "YYYY-MM",
    "displayFormat": "MMMM YYYY",
    "delimiter": " through ",
    "minDuration": "1month",
    "maxDuration": "6months",
    "shortcuts": [
        {"label": "This Month", "value": "thisMonth"},
        {"label": "Last Month", "value": "lastMonth"},
        {"label": "Last 3 Months", "value": "last3months"}
    ],
    "required": true,
    "autoFill": {
        "api": "/api/payroll/period-summary/${value}",
        "fillMapping": {
            "employee_count": "total_employees",
            "total_hours": "total_worked_hours",
            "overtime_hours": "total_overtime"
        }
    }
}
```

### Performance Review Cycle
```json
{
    "type": "input-month-range",
    "name": "review_cycle",
    "label": "Performance Review Cycle",
    "startPlaceholder": "Review cycle start",
    "endPlaceholder": "Review cycle end",
    "format": "YYYY-MM",
    "displayFormat": "MMMM YYYY",
    "delimiter": " to ",
    "minDuration": "3months",
    "maxDuration": "12months",
    "shortcuts": [
        {"label": "Q1 (Jan-Mar)", "value": "Q1"},
        {"label": "Q2 (Apr-Jun)", "value": "Q2"},
        {"label": "Q3 (Jul-Sep)", "value": "Q3"},
        {"label": "Q4 (Oct-Dec)", "value": "Q4"},
        {"label": "Full Year", "value": "thisYear"}
    ],
    "borderMode": "full",
    "required": true
}
```

### Sales Target Period
```json
{
    "type": "input-month-range",
    "name": "sales_target_period",
    "label": "Sales Target Period",
    "startPlaceholder": "Target start month",
    "endPlaceholder": "Target end month",
    "format": "YYYY-MM",
    "displayFormat": "MMM YYYY",
    "delimiter": " - ",
    "minDuration": "1month",
    "maxDuration": "18months",
    "shortcuts": [
        {"label": "Next Quarter", "value": "nextQuarter"},
        {"label": "Next 6 Months", "value": "next6months"},
        {"label": "Next Year", "value": "nextYear"}
    ],
    "validations": {
        "isDateTimeAfter": ["${currentMonth}", "month"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Sales targets must be set for future periods"
    }
}
```

### Inventory Analysis Period
```json
{
    "type": "input-month-range",
    "name": "inventory_analysis_period",
    "label": "Inventory Analysis Period",
    "startPlaceholder": "Analysis start",
    "endPlaceholder": "Analysis end",
    "format": "YYYY-MM",
    "displayFormat": "MMMM YYYY",
    "delimiter": " to ",
    "maxDuration": "24months",
    "shortcuts": [
        {"label": "Last Quarter", "value": "lastQuarter"},
        {"label": "Last 6 Months", "value": "last6months"},
        {"label": "Last Year", "value": "lastYear"},
        {"label": "Year to Date", "value": "yearToDate"}
    ],
    "embed": true,
    "borderMode": "half"
}
```

### Project Timeline Range
```json
{
    "type": "input-month-range",
    "name": "project_duration",
    "label": "Project Duration",
    "startPlaceholder": "Project start month",
    "endPlaceholder": "Project end month",
    "format": "YYYY-MM",
    "displayFormat": "MMMM YYYY",
    "delimiter": " to ",
    "minDuration": "1month",
    "maxDuration": "36months",
    "shortcuts": [
        {"label": "3 Months", "value": "next3months"},
        {"label": "6 Months", "value": "next6months"},
        {"label": "1 Year", "value": "next12months"}
    ],
    "required": true,
    "validations": {
        "isDateTimeSameOrAfter": ["${currentMonth}", "month"]
    },
    "validationErrors": {
        "isDateTimeSameOrAfter": "Project cannot start in the past"
    }
}
```

### Contract Duration
```json
{
    "type": "input-month-range",
    "name": "contract_period",
    "label": "Contract Period",
    "startPlaceholder": "Contract start",
    "endPlaceholder": "Contract end",
    "format": "YYYY-MM",
    "displayFormat": "MMMM YYYY",
    "delimiter": " through ",
    "minDuration": "1month",
    "maxDuration": "60months",
    "shortcuts": [
        {"label": "6 Months", "value": "next6months"},
        {"label": "1 Year", "value": "next12months"},
        {"label": "2 Years", "value": "next24months"},
        {"label": "3 Years", "value": "next36months"}
    ],
    "required": true,
    "autoFill": {
        "api": "/api/contracts/calculate-terms/${value}",
        "fillMapping": {
            "contract_length": "duration_months",
            "estimated_value": "total_value",
            "renewal_date": "auto_renewal_date"
        }
    }
}
```

### Maintenance Schedule Range
```json
{
    "type": "input-month-range",
    "name": "maintenance_schedule",
    "label": "Maintenance Schedule",
    "startPlaceholder": "Schedule start",
    "endPlaceholder": "Schedule end",
    "format": "YYYY-MM",
    "displayFormat": "MMM YYYY",
    "delimiter": " to ",
    "minDuration": "1month",
    "maxDuration": "12months",
    "shortcuts": [
        {"label": "Next Quarter", "value": "nextQuarter"},
        {"label": "Next 6 Months", "value": "next6months"},
        {"label": "Rest of Year", "value": "restOfYear"}
    ],
    "validations": {
        "isDateTimeAfter": ["${currentMonth}", "month"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Maintenance must be scheduled for future months"
    },
    "hint": "Select the period for scheduled maintenance activities"
}
```

### Fiscal Period Range
```json
{
    "type": "input-month-range",
    "name": "fiscal_range",
    "label": "Fiscal Period",
    "startPlaceholder": "Fiscal start",
    "endPlaceholder": "Fiscal end",
    "format": "YYYY-MM",
    "displayFormat": "MMMM YYYY",
    "delimiter": " - ",
    "shortcuts": [
        {"label": "Current FY", "value": "currentFiscalYear"},
        {"label": "Next FY", "value": "nextFiscalYear"},
        {"label": "Q1 FY", "value": "fiscalQ1"},
        {"label": "Q2 FY", "value": "fiscalQ2"},
        {"label": "Q3 FY", "value": "fiscalQ3"},
        {"label": "Q4 FY", "value": "fiscalQ4"}
    ],
    "minDate": "2020-01",
    "maxDate": "+60months",
    "joinValues": true,
    "delimiter": ","
}
```

This component provides essential month range selection functionality for ERP systems requiring period-based reporting, planning, and time-bound business processes.