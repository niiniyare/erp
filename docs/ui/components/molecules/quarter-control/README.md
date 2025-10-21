# Quarter Control Component

**FILE PURPOSE**: Quarter selection control for quarterly reporting and planning interfaces  
**SCOPE**: Quarter picking, fiscal quarters, quarterly data entry, and business reporting cycles  
**TARGET AUDIENCE**: Developers implementing quarterly reports, business planning, and quarterly data management features

## 📋 Component Overview

Quarter Control provides quarter-only selection functionality with fiscal year support, format customization, shortcuts, and validation. Essential for quarterly business reporting and planning interfaces in ERP systems.

### Schema Reference
- **Primary Schema**: `QuarterControlSchema.json`
- **Related Schemas**: `FormHorizontal.json`, `ShortCuts.json`
- **Base Interface**: Form input control for quarter selection

## Basic Usage

```json
{
    "type": "input-quarter",
    "name": "report_quarter",
    "label": "Report Quarter",
    "placeholder": "Select quarter..."
}
```

## Go Type Definition

```go
type QuarterControlProps struct {
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
    Embed           bool                `json:"embed"`           // Inline mode
    BorderMode      string              `json:"borderMode"`      // "full", "half", "none"
    Shortcuts       interface{}         `json:"shortcuts"`       // string or array
    DisabledDate    string              `json:"disabledDate"`    // Date disable function
    InputForbid     bool                `json:"inputForbid"`     // Forbid manual input
}
```

## Essential Variants

### Basic Quarter Picker
```json
{
    "type": "input-quarter",
    "name": "target_quarter",
    "label": "Target Quarter",
    "placeholder": "Select target quarter",
    "format": "X",
    "clearable": true
}
```

### Quarterly Report Selection
```json
{
    "type": "input-quarter",
    "name": "reporting_quarter",
    "label": "Reporting Quarter",
    "placeholder": "Select reporting quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "clearable": true,
    "required": true
}
```

### Fiscal Quarter Picker
```json
{
    "type": "input-quarter",
    "name": "fiscal_quarter",
    "label": "Fiscal Quarter",
    "placeholder": "Select fiscal quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "shortcuts": ["thisQuarter", "lastQuarter", "nextQuarter"],
    "clearable": true
}
```

### Embedded Quarter Calendar
```json
{
    "type": "input-quarter",
    "name": "planning_quarter",
    "label": "Planning Quarter",
    "embed": true,
    "format": "X",
    "borderMode": "none"
}
```

### UTC Quarter Storage
```json
{
    "type": "input-quarter",
    "name": "global_quarter",
    "label": "Global Reporting Quarter",
    "placeholder": "Select quarter for global report",
    "format": "YYYY-[Q]Q",
    "displayFormat": "[Q]Q YYYY",
    "utc": true,
    "clearable": true
}
```

## Real-World Use Cases

### Quarterly Business Planning
```json
{
    "type": "input-quarter",
    "name": "planning_quarter",
    "label": "Planning Quarter",
    "placeholder": "Select planning quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "shortcuts": [
        {"label": "Current Quarter", "value": "thisQuarter"},
        {"label": "Next Quarter", "value": "nextQuarter"},
        {"label": "Next Year Q1", "value": "+1year,Q1"}
    ],
    "clearable": true,
    "required": true,
    "validations": {
        "isDateTimeAfter": ["${currentQuarter}", "quarter"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Planning quarter must be current quarter or later"
    }
}
```

### Quarterly Financial Reporting
```json
{
    "type": "input-quarter",
    "name": "financial_quarter",
    "label": "Financial Quarter",
    "placeholder": "Select financial reporting quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "shortcuts": [
        {"label": "Last Quarter", "value": "lastQuarter"},
        {"label": "Current Quarter", "value": "thisQuarter"},
        {"label": "Year to Date", "value": "yearToDate"}
    ],
    "clearable": true,
    "required": true,
    "autoFill": {
        "api": "/api/financial/quarterly-summary/${value}",
        "fillMapping": {
            "revenue": "total_revenue",
            "expenses": "total_expenses",
            "profit": "net_profit"
        }
    }
}
```

### Sales Performance Quarter
```json
{
    "type": "input-quarter",
    "name": "sales_quarter",
    "label": "Sales Performance Quarter",
    "placeholder": "Select sales quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "shortcuts": [
        {"label": "Q1", "value": "thisYear,Q1"},
        {"label": "Q2", "value": "thisYear,Q2"},
        {"label": "Q3", "value": "thisYear,Q3"},
        {"label": "Q4", "value": "thisYear,Q4"}
    ],
    "clearable": true,
    "required": true,
    "autoFill": {
        "api": "/api/sales/quarterly-stats/${value}",
        "fillMapping": {
            "total_sales": "sales_amount",
            "deals_closed": "deals_count",
            "targets_met": "target_achievement"
        }
    }
}
```

### Budget Review Quarter
```json
{
    "type": "input-quarter",
    "name": "budget_review_quarter",
    "label": "Budget Review Quarter",
    "placeholder": "Select budget review quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "shortcuts": [
        {"label": "Current Quarter", "value": "thisQuarter"},
        {"label": "Next Quarter", "value": "nextQuarter"},
        {"label": "Last Quarter", "value": "lastQuarter"}
    ],
    "clearable": true,
    "validations": {
        "isDateTimeBetween": ["lastYear", "nextYear", "quarter"]
    },
    "validationErrors": {
        "isDateTimeBetween": "Budget review must be within reasonable timeframe"
    }
}
```

### Inventory Audit Quarter
```json
{
    "type": "input-quarter",
    "name": "audit_quarter",
    "label": "Inventory Audit Quarter",
    "placeholder": "Select audit quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "shortcuts": [
        {"label": "Last Quarter", "value": "lastQuarter"},
        {"label": "Current Quarter", "value": "thisQuarter"}
    ],
    "disabledDate": "moment(currentDate).isAfter(moment())",
    "clearable": true,
    "required": true,
    "hint": "Select the quarter for inventory audit"
}
```

### Performance Review Quarter
```json
{
    "type": "input-quarter",
    "name": "review_quarter",
    "label": "Performance Review Quarter",
    "placeholder": "Select review quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "shortcuts": [
        {"label": "Last Quarter", "value": "lastQuarter"},
        {"label": "2 Quarters Ago", "value": "-2quarters"},
        {"label": "3 Quarters Ago", "value": "-3quarters"}
    ],
    "disabledDate": "moment(currentDate).isAfter(moment())",
    "clearable": true,
    "required": true
}
```

### Project Milestone Quarter
```json
{
    "type": "input-quarter",
    "name": "milestone_quarter",
    "label": "Milestone Quarter",
    "placeholder": "Select milestone quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "shortcuts": [
        {"label": "Next Quarter", "value": "nextQuarter"},
        {"label": "Q2 Next Year", "value": "+1year,Q2"},
        {"label": "Q4 Next Year", "value": "+1year,Q4"}
    ],
    "validations": {
        "isDateTimeAfter": ["${currentQuarter}", "quarter"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Milestone must be set for future quarters"
    }
}
```

### Tax Filing Quarter
```json
{
    "type": "input-quarter",
    "name": "tax_quarter",
    "label": "Tax Filing Quarter",
    "placeholder": "Select tax filing quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "shortcuts": [
        {"label": "Q1 (Jan-Mar)", "value": "thisYear,Q1"},
        {"label": "Q2 (Apr-Jun)", "value": "thisYear,Q2"},
        {"label": "Q3 (Jul-Sep)", "value": "thisYear,Q3"},
        {"label": "Q4 (Oct-Dec)", "value": "thisYear,Q4"}
    ],
    "clearable": true,
    "required": true,
    "autoFill": {
        "api": "/api/tax/quarterly-filing/${value}",
        "fillMapping": {
            "filing_deadline": "deadline",
            "estimated_tax": "estimated_amount",
            "previous_payments": "payments_made"
        }
    }
}
```

### Subscription Billing Quarter
```json
{
    "type": "input-quarter",
    "name": "billing_quarter",
    "label": "Billing Quarter",
    "placeholder": "Select billing quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "shortcuts": [
        {"label": "Current Quarter", "value": "thisQuarter"},
        {"label": "Next Quarter", "value": "nextQuarter"},
        {"label": "Next Year Q1", "value": "+1year,Q1"}
    ],
    "clearable": true,
    "required": true,
    "validations": {
        "isDateTimeSameOrAfter": ["${currentQuarter}", "quarter"]
    },
    "validationErrors": {
        "isDateTimeSameOrAfter": "Billing quarter cannot be in the past"
    }
}
```

### Market Analysis Quarter
```json
{
    "type": "input-quarter",
    "name": "analysis_quarter",
    "label": "Market Analysis Quarter",
    "placeholder": "Select analysis quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "embed": true,
    "borderMode": "full",
    "shortcuts": [
        {"label": "Last Quarter", "value": "lastQuarter"},
        {"label": "Same Quarter Last Year", "value": "-1year,sameQuarter"}
    ],
    "hint": "Select quarter for market analysis comparison"
}
```

This component provides essential quarter selection functionality for ERP systems requiring quarterly business cycles, reporting, and planning capabilities.