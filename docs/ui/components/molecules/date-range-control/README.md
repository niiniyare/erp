# Date Range Control Component

**FILE PURPOSE**: Date range selection control for period-based data filtering and reporting  
**SCOPE**: Date range picking, period selection, reporting filters, and temporal data management  
**TARGET AUDIENCE**: Developers implementing date range filters, reporting interfaces, and temporal data selection

## 📋 Component Overview

Date Range Control provides comprehensive date range selection functionality with start/end date picking, predefined ranges, shortcuts, and validation. Essential for reporting, filtering, and temporal data management in ERP systems.

### Schema Reference
- **Primary Schema**: `DateRangeControlSchema.json` 
- **Related Schemas**: `FormHorizontal.json`, `ShortCuts.json`, `DateRangeSchema.json`
- **Base Interface**: Form input control for date range selection

## Basic Usage

```json
{
    "type": "input-date-range",
    "name": "reporting_period",
    "label": "Reporting Period",
    "placeholder": "Select date range..."
}
```

## Go Type Definition

```go
type DateRangeControlProps struct {
    Type               string              `json:"type"`
    Name               string              `json:"name"`
    Label              interface{}         `json:"label"`
    Placeholder        string              `json:"placeholder"`
    StartPlaceholder   string              `json:"startPlaceholder"`
    EndPlaceholder     string              `json:"endPlaceholder"`
    Format             string              `json:"format"`           // Storage format
    ValueFormat        string              `json:"valueFormat"`      // Alternative format
    InputFormat        string              `json:"inputFormat"`      // Display format (legacy)
    DisplayFormat      string              `json:"displayFormat"`    // Display format (new)
    Delimiter          string              `json:"delimiter"`        // Value separator
    JoinValues         bool                `json:"joinValues"`       // Join as string
    MinDate            string              `json:"minDate"`          // Minimum date
    MaxDate            string              `json:"maxDate"`          // Maximum date
    MinDuration        string              `json:"minDuration"`      // Minimum span
    MaxDuration        string              `json:"maxDuration"`      // Maximum span
    BorderMode         string              `json:"borderMode"`       // "full", "half", "none"
    Embed              bool                `json:"embed"`            // Inline mode
    Shortcuts          interface{}         `json:"shortcuts"`        // Range shortcuts
    Ranges             interface{}         `json:"ranges"`           // Same as shortcuts
    Animation          bool                `json:"animation"`        // Cursor animation
    UTC                bool                `json:"utc"`              // Store as UTC
    ClearButton        bool                `json:"clearButton"`      // Clear button
    CloseOnSelect      bool                `json:"closeOnSelect"`    // Auto close
    Transform          string              `json:"transform"`        // Date transform function
}
```

## Essential Variants

### Basic Date Range
```json
{
    "type": "input-date-range",
    "name": "period_filter",
    "label": "Period Filter",
    "startPlaceholder": "Start date",
    "endPlaceholder": "End date",
    "format": "YYYY-MM-DD",
    "displayFormat": "MMM DD, YYYY",
    "delimiter": " to ",
    "clearButton": true
}
```

### Reporting Date Range
```json
{
    "type": "input-date-range",
    "name": "report_range",
    "label": "Report Range",
    "format": "YYYY-MM-DD",
    "displayFormat": "MM/DD/YYYY",
    "shortcuts": [
        {"label": "Today", "value": "today"},
        {"label": "This Week", "value": "thisWeek"},
        {"label": "This Month", "value": "thisMonth"},
        {"label": "Last 30 Days", "value": "last30days"}
    ],
    "maxDuration": "1year"
}
```

### Financial Period Range
```json
{
    "type": "input-date-range",
    "name": "financial_period",
    "label": "Financial Period",
    "format": "YYYY-MM-DD",
    "displayFormat": "MMM DD, YYYY",
    "shortcuts": [
        {"label": "Current Quarter", "value": "thisQuarter"},
        {"label": "Last Quarter", "value": "lastQuarter"},
        {"label": "This Year", "value": "thisYear"},
        {"label": "Last Year", "value": "lastYear"}
    ],
    "minDate": "2020-01-01",
    "joinValues": true
}
```

### Project Timeline Range
```json
{
    "type": "input-date-range",
    "name": "project_timeline",
    "label": "Project Timeline",
    "format": "YYYY-MM-DD",
    "displayFormat": "MMM DD, YYYY",
    "minDuration": "1day",
    "maxDuration": "2years",
    "embed": true,
    "animation": true,
    "closeOnSelect": false
}
```

## Real-World Use Cases

### Sales Report Date Range
```json
{
    "type": "input-date-range",
    "name": "sales_report_period",
    "label": "Sales Report Period",
    "startPlaceholder": "Report start date",
    "endPlaceholder": "Report end date",
    "format": "YYYY-MM-DD",
    "displayFormat": "MMM DD, YYYY",
    "delimiter": " - ",
    "shortcuts": [
        {"label": "Today", "value": "today"},
        {"label": "Yesterday", "value": "yesterday"},
        {"label": "This Week", "value": "thisWeek"},
        {"label": "Last Week", "value": "lastWeek"},
        {"label": "This Month", "value": "thisMonth"},
        {"label": "Last Month", "value": "lastMonth"},
        {"label": "This Quarter", "value": "thisQuarter"},
        {"label": "Last Quarter", "value": "lastQuarter"},
        {"label": "This Year", "value": "thisYear"},
        {"label": "Last Year", "value": "lastYear"},
        {"label": "Last 7 Days", "value": "last7days"},
        {"label": "Last 30 Days", "value": "last30days"},
        {"label": "Last 90 Days", "value": "last90days"}
    ],
    "maxDuration": "2years",
    "clearButton": true,
    "required": true,
    "validations": {
        "isDateTimeBefore": ["${end_date}", "${current_date}", "day"]
    },
    "validationErrors": {
        "isDateTimeBefore": "Report period cannot extend into the future"
    },
    "autoFill": {
        "api": "/api/sales/period-summary/${value}",
        "fillMapping": {
            "total_sales": "period_sales",
            "order_count": "total_orders",
            "avg_order_value": "average_order",
            "growth_rate": "period_growth"
        }
    }
}
```

### Employee Timesheet Period
```json
{
    "type": "input-date-range",
    "name": "timesheet_period",
    "label": "Timesheet Period",
    "startPlaceholder": "Period start",
    "endPlaceholder": "Period end",
    "format": "YYYY-MM-DD",
    "displayFormat": "ddd, MMM DD",
    "delimiter": " to ",
    "minDuration": "1day",
    "maxDuration": "14days",
    "shortcuts": [
        {"label": "This Week", "value": "thisWeek"},
        {"label": "Last Week", "value": "lastWeek"},
        {"label": "Current Pay Period", "value": "currentPayPeriod"},
        {"label": "Previous Pay Period", "value": "previousPayPeriod"}
    ],
    "clearButton": true,
    "required": true,
    "autoFill": {
        "api": "/api/timesheets/period-data/${employee_id}/${value}",
        "fillMapping": {
            "total_hours": "hours_worked",
            "overtime_hours": "overtime",
            "vacation_hours": "vacation_used",
            "sick_hours": "sick_used"
        }
    }
}
```

### Project Milestone Range
```json
{
    "type": "input-date-range",
    "name": "milestone_period",
    "label": "Milestone Period",
    "startPlaceholder": "Milestone start",
    "endPlaceholder": "Milestone end",
    "format": "YYYY-MM-DD",
    "displayFormat": "MMM DD, YYYY",
    "delimiter": " → ",
    "minDuration": "1day",
    "maxDuration": "6months",
    "shortcuts": [
        {"label": "Next Week", "value": "nextWeek"},
        {"label": "Next Month", "value": "nextMonth"},
        {"label": "Next Quarter", "value": "nextQuarter"},
        {"label": "Sprint 1 (2 weeks)", "value": "next14days"},
        {"label": "Sprint 2 (4 weeks)", "value": "next28days"}
    ],
    "minDate": "${project_start_date}",
    "maxDate": "${project_end_date}",
    "validations": {
        "isDateTimeAfter": ["${start_date}", "${current_date}", "day"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Milestone period must be in the future"
    }
}
```

### Inventory Analysis Period
```json
{
    "type": "input-date-range",
    "name": "inventory_analysis_period",
    "label": "Inventory Analysis Period",
    "startPlaceholder": "Analysis start",
    "endPlaceholder": "Analysis end",
    "format": "YYYY-MM-DD",
    "displayFormat": "MMM DD, YYYY",
    "delimiter": " to ",
    "shortcuts": [
        {"label": "Last 7 Days", "value": "last7days"},
        {"label": "Last 30 Days", "value": "last30days"},
        {"label": "Last Quarter", "value": "lastQuarter"},
        {"label": "Last 6 Months", "value": "last6months"},
        {"label": "Last Year", "value": "lastYear"}
    ],
    "maxDuration": "2years",
    "embed": true,
    "borderMode": "half",
    "autoFill": {
        "api": "/api/inventory/analysis/${value}",
        "fillMapping": {
            "turnover_rate": "inventory_turnover",
            "stockout_days": "stockout_count",
            "carrying_cost": "storage_cost",
            "reorder_frequency": "reorder_count"
        }
    }
}
```

### Financial Statement Period
```json
{
    "type": "input-date-range",
    "name": "financial_statement_period",
    "label": "Financial Statement Period",
    "startPlaceholder": "Statement start",
    "endPlaceholder": "Statement end",
    "format": "YYYY-MM-DD",
    "displayFormat": "MMM DD, YYYY",
    "delimiter": " through ",
    "shortcuts": [
        {"label": "Current Month", "value": "thisMonth"},
        {"label": "Current Quarter", "value": "thisQuarter"},
        {"label": "Current Year", "value": "thisYear"},
        {"label": "Previous Month", "value": "lastMonth"},
        {"label": "Previous Quarter", "value": "lastQuarter"},
        {"label": "Previous Year", "value": "lastYear"},
        {"label": "Year to Date", "value": "yearToDate"}
    ],
    "minDate": "${fiscal_year_start}",
    "maxDate": "${current_date}",
    "joinValues": true,
    "delimiter": ",",
    "required": true,
    "autoFill": {
        "api": "/api/financial/statement-data/${value}",
        "fillMapping": {
            "total_revenue": "revenue",
            "total_expenses": "expenses",
            "net_income": "profit_loss",
            "cash_flow": "cash_flow"
        }
    }
}
```

### Campaign Performance Period
```json
{
    "type": "input-date-range",
    "name": "campaign_period",
    "label": "Campaign Period",
    "startPlaceholder": "Campaign start",
    "endPlaceholder": "Campaign end",
    "format": "YYYY-MM-DD",
    "displayFormat": "MMM DD, YYYY",
    "delimiter": " to ",
    "minDuration": "1day",
    "maxDuration": "1year",
    "shortcuts": [
        {"label": "Last 7 Days", "value": "last7days"},
        {"label": "Last 30 Days", "value": "last30days"},
        {"label": "This Month", "value": "thisMonth"},
        {"label": "Last Month", "value": "lastMonth"},
        {"label": "This Quarter", "value": "thisQuarter"}
    ],
    "clearButton": true,
    "autoFill": {
        "api": "/api/marketing/campaign-metrics/${value}",
        "fillMapping": {
            "impressions": "total_impressions",
            "clicks": "total_clicks",
            "conversions": "total_conversions",
            "cost": "total_cost",
            "roi": "return_on_investment"
        }
    }
}
```

### Maintenance Schedule Range
```json
{
    "type": "input-date-range",
    "name": "maintenance_schedule",
    "label": "Maintenance Schedule",
    "startPlaceholder": "Schedule start",
    "endPlaceholder": "Schedule end",
    "format": "YYYY-MM-DD",
    "displayFormat": "ddd, MMM DD",
    "delimiter": " - ",
    "minDuration": "1day",
    "maxDuration": "3months",
    "shortcuts": [
        {"label": "Next Week", "value": "nextWeek"},
        {"label": "Next Month", "value": "nextMonth"},
        {"label": "Next Quarter", "value": "nextQuarter"}
    ],
    "minDate": "${current_date}",
    "validations": {
        "isDateTimeAfter": ["${start_date}", "${current_date}", "day"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Maintenance must be scheduled for future dates"
    }
}
```

### Audit Period Selection
```json
{
    "type": "input-date-range",
    "name": "audit_period",
    "label": "Audit Period",
    "startPlaceholder": "Audit start date",
    "endPlaceholder": "Audit end date",
    "format": "YYYY-MM-DD",
    "displayFormat": "MMM DD, YYYY",
    "delimiter": " to ",
    "shortcuts": [
        {"label": "Last Month", "value": "lastMonth"},
        {"label": "Last Quarter", "value": "lastQuarter"},
        {"label": "Last 6 Months", "value": "last6months"},
        {"label": "Last Year", "value": "lastYear"},
        {"label": "Custom Period", "value": "custom"}
    ],
    "maxDate": "${current_date}",
    "maxDuration": "2years",
    "required": true,
    "autoFill": {
        "api": "/api/audit/scope-analysis/${value}",
        "fillMapping": {
            "transaction_count": "total_transactions",
            "departments_affected": "department_count",
            "estimated_hours": "audit_hours",
            "risk_level": "risk_assessment"
        }
    }
}
```

### Training Program Duration
```json
{
    "type": "input-date-range",
    "name": "training_duration",
    "label": "Training Program Duration",
    "startPlaceholder": "Program start",
    "endPlaceholder": "Program end",
    "format": "YYYY-MM-DD",
    "displayFormat": "MMM DD, YYYY",
    "delimiter": " to ",
    "minDuration": "1day",
    "maxDuration": "1year",
    "shortcuts": [
        {"label": "1 Week Program", "value": "next7days"},
        {"label": "2 Week Program", "value": "next14days"},
        {"label": "1 Month Program", "value": "next30days"},
        {"label": "3 Month Program", "value": "next90days"}
    ],
    "minDate": "${current_date}",
    "validations": {
        "isDateTimeAfter": ["${start_date}", "${current_date}", "day"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Training program must start in the future"
    }
}
```

This component provides essential date range selection functionality for ERP systems requiring temporal data filtering, period-based reporting, and time-bound business process management.