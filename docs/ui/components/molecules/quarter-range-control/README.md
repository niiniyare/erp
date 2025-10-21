# Quarter Range Control Component

**FILE PURPOSE**: Quarter range selection for multi-quarter reporting and planning  
**SCOPE**: Quarter range picking, fiscal periods, multi-quarter analysis, and business cycle planning  
**TARGET AUDIENCE**: Developers implementing multi-quarter reports, business planning, and quarterly comparison features

## 📋 Component Overview

Quarter Range Control provides quarter range selection with start/end quarter picking, duration limits, fiscal year support, shortcuts, and validation. Essential for multi-quarter reporting and business cycle analysis in ERP systems.

### Schema Reference
- **Primary Schema**: `QuarterRangeControlSchema.json`
- **Related Schemas**: `FormHorizontal.json`, `ShortCuts.json`
- **Base Interface**: Form input control for quarter range selection

## Basic Usage

```json
{
    "type": "input-quarter-range",
    "name": "reporting_period",
    "label": "Reporting Period",
    "placeholder": "Select quarter range..."
}
```

## Go Type Definition

```go
type QuarterRangeControlProps struct {
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

### Basic Quarter Range
```json
{
    "type": "input-quarter-range",
    "name": "business_period",
    "label": "Business Period",
    "startPlaceholder": "Start quarter",
    "endPlaceholder": "End quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "delimiter": " to ",
    "required": true
}
```

### Fiscal Year Range
```json
{
    "type": "input-quarter-range",
    "name": "fiscal_period",
    "label": "Fiscal Period",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "minDuration": "1quarter",
    "maxDuration": "8quarters",
    "shortcuts": [
        {"label": "This Fiscal Year", "value": "thisFiscalYear"},
        {"label": "Last Fiscal Year", "value": "lastFiscalYear"},
        {"label": "Last 2 Years", "value": "last8quarters"}
    ]
}
```

### Comparison Period Range
```json
{
    "type": "input-quarter-range",
    "name": "comparison_period",
    "label": "Comparison Period",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "shortcuts": ["lastYear", "thisYear", "last2quarters"],
    "joinValues": true,
    "delimiter": ","
}
```

## Real-World Use Cases

### Multi-Quarter Budget Planning
```json
{
    "type": "input-quarter-range",
    "name": "budget_planning_period",
    "label": "Budget Planning Period",
    "startPlaceholder": "Planning start quarter",
    "endPlaceholder": "Planning end quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "delimiter": " through ",
    "minDuration": "2quarters",
    "maxDuration": "12quarters",
    "shortcuts": [
        {"label": "Next 2 Quarters", "value": "next2quarters"},
        {"label": "Next 4 Quarters", "value": "next4quarters"},
        {"label": "Next Fiscal Year", "value": "nextFiscalYear"},
        {"label": "Next 2 Years", "value": "next8quarters"}
    ],
    "required": true,
    "validations": {
        "isDateTimeAfter": ["${currentQuarter}", "quarter"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Budget planning must start from current quarter or later"
    }
}
```

### Performance Analysis Range
```json
{
    "type": "input-quarter-range",
    "name": "performance_analysis_period",
    "label": "Performance Analysis Period",
    "startPlaceholder": "Analysis start",
    "endPlaceholder": "Analysis end",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "delimiter": " - ",
    "maxDuration": "12quarters",
    "shortcuts": [
        {"label": "Last 2 Quarters", "value": "last2quarters"},
        {"label": "Last 4 Quarters", "value": "last4quarters"},
        {"label": "Last Year", "value": "lastYear"},
        {"label": "Last 2 Years", "value": "last8quarters"}
    ],
    "embed": true,
    "borderMode": "half"
}
```

### Sales Target Period
```json
{
    "type": "input-quarter-range",
    "name": "sales_target_period",
    "label": "Sales Target Period",
    "startPlaceholder": "Target start quarter",
    "endPlaceholder": "Target end quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "delimiter": " to ",
    "minDuration": "1quarter",
    "maxDuration": "8quarters",
    "shortcuts": [
        {"label": "Rest of Year", "value": "restOfYear"},
        {"label": "Next Year", "value": "nextYear"},
        {"label": "Next 6 Quarters", "value": "next6quarters"}
    ],
    "validations": {
        "isDateTimeAfter": ["${currentQuarter}", "quarter"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Sales targets must be set for future periods"
    }
}
```

### Financial Comparison Range
```json
{
    "type": "input-quarter-range",
    "name": "financial_comparison",
    "label": "Financial Comparison Period",
    "startPlaceholder": "Comparison start",
    "endPlaceholder": "Comparison end",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "delimiter": " vs ",
    "maxDuration": "8quarters",
    "shortcuts": [
        {"label": "YoY Same Quarters", "value": "yoyComparison"},
        {"label": "Last 4 vs This 4", "value": "rolling4quarters"},
        {"label": "Last Year vs This Year", "value": "yearComparison"}
    ],
    "autoFill": {
        "api": "/api/financial/quarter-comparison/${value}",
        "fillMapping": {
            "revenue_growth": "revenue_change_percent",
            "profit_growth": "profit_change_percent",
            "cost_variance": "cost_change_percent"
        }
    }
}
```

### Project Timeline Range
```json
{
    "type": "input-quarter-range",
    "name": "project_timeline",
    "label": "Project Timeline",
    "startPlaceholder": "Project start quarter",
    "endPlaceholder": "Project end quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "delimiter": " to ",
    "minDuration": "1quarter",
    "maxDuration": "16quarters",
    "shortcuts": [
        {"label": "2 Quarters", "value": "next2quarters"},
        {"label": "1 Year", "value": "next4quarters"},
        {"label": "2 Years", "value": "next8quarters"}
    ],
    "required": true,
    "validations": {
        "isDateTimeSameOrAfter": ["${currentQuarter}", "quarter"]
    },
    "validationErrors": {
        "isDateTimeSameOrAfter": "Project cannot start in the past"
    }
}
```

### Market Research Period
```json
{
    "type": "input-quarter-range",
    "name": "market_research_period",
    "label": "Market Research Period",
    "startPlaceholder": "Research start",
    "endPlaceholder": "Research end",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "delimiter": " through ",
    "minDuration": "2quarters",
    "maxDuration": "12quarters",
    "shortcuts": [
        {"label": "Last 4 Quarters", "value": "last4quarters"},
        {"label": "Last 6 Quarters", "value": "last6quarters"},
        {"label": "Last 2 Years", "value": "last8quarters"}
    ],
    "hint": "Select quarters for market trend analysis"
}
```

### Audit Period Range
```json
{
    "type": "input-quarter-range",
    "name": "audit_period",
    "label": "Audit Period",
    "startPlaceholder": "Audit start quarter",
    "endPlaceholder": "Audit end quarter",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "delimiter": " to ",
    "maxDuration": "8quarters",
    "shortcuts": [
        {"label": "Last Quarter", "value": "lastQuarter"},
        {"label": "Last 2 Quarters", "value": "last2quarters"},
        {"label": "Last Year", "value": "lastYear"}
    ],
    "required": true,
    "autoFill": {
        "api": "/api/audit/period-scope/${value}",
        "fillMapping": {
            "transaction_count": "total_transactions",
            "departments_involved": "department_list",
            "compliance_areas": "areas_to_audit"
        }
    }
}
```

### Revenue Recognition Period
```json
{
    "type": "input-quarter-range",
    "name": "revenue_recognition_period",
    "label": "Revenue Recognition Period",
    "startPlaceholder": "Recognition start",
    "endPlaceholder": "Recognition end",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "delimiter": " - ",
    "minDuration": "1quarter",
    "maxDuration": "16quarters",
    "shortcuts": [
        {"label": "Current FY", "value": "currentFiscalYear"},
        {"label": "Next FY", "value": "nextFiscalYear"},
        {"label": "Multi-Year (8Q)", "value": "next8quarters"}
    ],
    "required": true
}
```

### Strategic Planning Horizon
```json
{
    "type": "input-quarter-range",
    "name": "strategic_planning_horizon",
    "label": "Strategic Planning Horizon",
    "startPlaceholder": "Planning start",
    "endPlaceholder": "Planning end",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "delimiter": " to ",
    "minDuration": "4quarters",
    "maxDuration": "20quarters",
    "shortcuts": [
        {"label": "1 Year Plan", "value": "next4quarters"},
        {"label": "3 Year Plan", "value": "next12quarters"},
        {"label": "5 Year Plan", "value": "next20quarters"}
    ],
    "validations": {
        "isDateTimeAfter": ["${currentQuarter}", "quarter"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Strategic planning must be for future quarters"
    }
}
```

### Contract Performance Period
```json
{
    "type": "input-quarter-range",
    "name": "contract_performance_period",
    "label": "Contract Performance Period",
    "startPlaceholder": "Performance start",
    "endPlaceholder": "Performance end",
    "format": "X",
    "displayFormat": "[Q]Q YYYY",
    "delimiter": " through ",
    "minDuration": "1quarter",
    "maxDuration": "24quarters",
    "shortcuts": [
        {"label": "Contract Year 1", "value": "contractYear1"},
        {"label": "Contract Year 2", "value": "contractYear2"},
        {"label": "Full Contract", "value": "fullContract"}
    ],
    "autoFill": {
        "api": "/api/contracts/performance-metrics/${value}",
        "fillMapping": {
            "milestones": "quarterly_milestones",
            "deliverables": "expected_deliverables",
            "payment_schedule": "payment_quarters"
        }
    }
}
```

This component provides essential quarter range selection functionality for ERP systems requiring multi-quarter business analysis, planning, and comparative reporting capabilities.