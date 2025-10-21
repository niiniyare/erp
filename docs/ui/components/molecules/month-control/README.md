# Month Control Component

**FILE PURPOSE**: Month selection control for monthly reporting and planning interfaces  
**SCOPE**: Month picking, fiscal periods, reporting cycles, and monthly data entry  
**TARGET AUDIENCE**: Developers implementing monthly reports, budgets, planning, and recurring monthly tasks

## 📋 Component Overview

Month Control provides month-only selection functionality with format customization, validation, shortcuts, and UTC support. Essential for monthly reporting, budgeting, and planning interfaces in ERP systems.

### Schema Reference
- **Primary Schema**: `MonthControlSchema.json`
- **Related Schemas**: `FormHorizontal.json`, `ShortCuts.json`
- **Base Interface**: Form input control for month selection

## Basic Usage

```json
{
    "type": "input-month",
    "name": "report_month",
    "label": "Report Month",
    "placeholder": "Select month..."
}
```

## Go Type Definition

```go
type MonthControlProps struct {
    Type            string              `json:"type"`
    Name            string              `json:"name"`
    Label           interface{}         `json:"label"`
    Placeholder     string              `json:"placeholder"`
    Value           interface{}         `json:"value"`
    Format          string              `json:"format"`          // Storage format
    ValueFormat     string              `json:"valueFormat"`     // Alternative format
    InputFormat     string              `json:"inputFormat"`     // Display format (legacy)
    DisplayFormat   string              `json:"displayFormat"`   // Display format (new)
    UTC             bool                `json:"utc"`             // Store as UTC
    Clearable       bool                `json:"clearable"`
    Disabled        bool                `json:"disabled"`
    ReadOnly        bool                `json:"readOnly"`
    Required        bool                `json:"required"`
    Embed           bool                `json:"emebed"`          // Inline mode
    BorderMode      string              `json:"borderMode"`      // "full", "half", "none"
    Shortcuts       interface{}         `json:"shortcuts"`       // string or array
    DisabledDate    string              `json:"disabledDate"`    // Date disable function
    InputForbid     bool                `json:"inputForbid"`     // Forbid manual input
}
```

## Essential Variants

### Basic Month Picker
```json
{
    "type": "input-month",
    "name": "target_month",
    "label": "Target Month",
    "placeholder": "Select target month",
    "format": "YYYY-MM",
    "clearable": true
}
```

### Monthly Report Selection
```json
{
    "type": "input-month",
    "name": "reporting_month",
    "label": "Reporting Month",
    "placeholder": "Select reporting month",
    "format": "YYYY-MM",
    "displayFormat": "MMMM YYYY",
    "clearable": true,
    "required": true
}
```

### Fiscal Month Picker
```json
{
    "type": "input-month",
    "name": "fiscal_month",
    "label": "Fiscal Month",
    "placeholder": "Select fiscal month",
    "format": "YYYY-MM",
    "displayFormat": "MMM YYYY",
    "shortcuts": ["thisMonth", "lastMonth", "nextMonth"],
    "clearable": true
}
```

### Embedded Month Calendar
```json
{
    "type": "input-month",
    "name": "planning_month",
    "label": "Planning Month",
    "emebed": true,
    "format": "YYYY-MM",
    "borderMode": "none"
}
```

### UTC Month Storage
```json
{
    "type": "input-month",
    "name": "global_month",
    "label": "Global Reporting Month",
    "placeholder": "Select month for global report",
    "format": "YYYY-MM-DD[T]HH:mm:ss[Z]",
    "displayFormat": "MMMM YYYY",
    "utc": true,
    "clearable": true
}
```

## Real-World Use Cases

### Monthly Budget Planning
```json
{
    "type": "input-month",
    "name": "budget_month",
    "label": "Budget Month",
    "placeholder": "Select budget planning month",
    "format": "YYYY-MM",
    "displayFormat": "MMMM YYYY",
    "shortcuts": [
        {"label": "Current Month", "value": "thisMonth"},
        {"label": "Next Month", "value": "nextMonth"},
        {"label": "Next Quarter Start", "value": "+3months"}
    ],
    "clearable": true,
    "required": true,
    "validations": {
        "isDateTimeAfter": ["${currentMonth}", "month"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Budget month must be current month or later"
    }
}
```

### Payroll Period Selection
```json
{
    "type": "input-month",
    "name": "payroll_month",
    "label": "Payroll Month",
    "placeholder": "Select payroll month",
    "format": "YYYY-MM",
    "displayFormat": "MMM YYYY",
    "shortcuts": ["thisMonth", "lastMonth"],
    "disabledDate": "moment().add(1, 'month').isBefore(moment(currentDate))",
    "clearable": true,
    "required": true,
    "hint": "Select the month for payroll processing"
}
```

### Monthly Performance Review
```json
{
    "type": "input-month",
    "name": "review_month",
    "label": "Review Month",
    "placeholder": "Select review month",
    "format": "YYYY-MM",
    "displayFormat": "MMMM YYYY",
    "shortcuts": [
        {"label": "Last Month", "value": "lastMonth"},
        {"label": "2 Months Ago", "value": "-2months"},
        {"label": "3 Months Ago", "value": "-3months"}
    ],
    "disabledDate": "moment(currentDate).isAfter(moment())",
    "clearable": true,
    "required": true
}
```

### Inventory Report Month
```json
{
    "type": "input-month",
    "name": "inventory_month",
    "label": "Inventory Report Month",
    "placeholder": "Select inventory reporting month",
    "format": "YYYY-MM",
    "displayFormat": "MMMM YYYY",
    "shortcuts": ["thisMonth", "lastMonth", "lastQuarter"],
    "clearable": true,
    "borderMode": "half",
    "autoFill": {
        "api": "/api/inventory/monthly-stats/${value}",
        "fillMapping": {
            "inventory_count": "total_items",
            "inventory_value": "total_value"
        }
    }
}
```

### Subscription Billing Month
```json
{
    "type": "input-month",
    "name": "billing_month",
    "label": "Billing Month",
    "placeholder": "Select billing month",
    "format": "YYYY-MM",
    "displayFormat": "MMM YYYY",
    "shortcuts": [
        {"label": "This Month", "value": "thisMonth"},
        {"label": "Next Month", "value": "nextMonth"},
        {"label": "Next Quarter", "value": "+3months"}
    ],
    "clearable": true,
    "required": true,
    "validations": {
        "isDateTimeSameOrAfter": ["${currentMonth}", "month"]
    },
    "validationErrors": {
        "isDateTimeSameOrAfter": "Billing month cannot be in the past"
    }
}
```

### Project Timeline Month
```json
{
    "type": "input-month",
    "name": "project_month",
    "label": "Project Month",
    "placeholder": "Select project month",
    "format": "YYYY-MM",
    "displayFormat": "MMMM YYYY",
    "emebed": true,
    "borderMode": "full",
    "disabledDate": "moment(currentDate).isBefore(moment('${project_start_date}'), 'month')",
    "hint": "Select a month within the project timeline"
}
```

### Maintenance Schedule Month
```json
{
    "type": "input-month",
    "name": "maintenance_month",
    "label": "Maintenance Month",
    "placeholder": "Schedule maintenance month",
    "format": "YYYY-MM",
    "displayFormat": "MMMM YYYY",
    "shortcuts": [
        {"label": "Next Month", "value": "nextMonth"},
        {"label": "In 2 Months", "value": "+2months"},
        {"label": "In 3 Months", "value": "+3months"},
        {"label": "In 6 Months", "value": "+6months"}
    ],
    "clearable": true,
    "required": true,
    "validations": {
        "isDateTimeAfter": ["${currentMonth}", "month"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Maintenance must be scheduled for future months"
    }
}
```

This component provides essential month selection functionality for ERP systems requiring monthly planning, reporting, and scheduling capabilities.