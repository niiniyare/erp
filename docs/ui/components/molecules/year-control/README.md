# Year Control Component

**FILE PURPOSE**: Year selection control for annual reporting and long-term planning interfaces  
**SCOPE**: Year picking, fiscal years, annual data entry, and long-term business planning  
**TARGET AUDIENCE**: Developers implementing annual reports, long-term planning, and year-based data management features

## 📋 Component Overview

Year Control provides year-only selection functionality with fiscal year support, format customization, shortcuts, and validation. Essential for annual reporting, budgeting, and long-term planning interfaces in ERP systems.

### Schema Reference
- **Primary Schema**: `YearControlSchema.json`
- **Related Schemas**: `FormHorizontal.json`, `ShortCuts.json`
- **Base Interface**: Form input control for year selection

## Basic Usage

```json
{
    "type": "input-year",
    "name": "report_year",
    "label": "Report Year",
    "placeholder": "Select year..."
}
```

## Go Type Definition

```go
type YearControlProps struct {
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
    MinYear         int                 `json:"minYear"`         // Minimum year
    MaxYear         int                 `json:"maxYear"`         // Maximum year
}
```

## Essential Variants

### Basic Year Picker
```json
{
    "type": "input-year",
    "name": "target_year",
    "label": "Target Year",
    "placeholder": "Select target year",
    "format": "YYYY",
    "clearable": true
}
```

### Annual Report Selection
```json
{
    "type": "input-year",
    "name": "reporting_year",
    "label": "Reporting Year",
    "placeholder": "Select reporting year",
    "format": "YYYY",
    "clearable": true,
    "required": true
}
```

### Fiscal Year Picker
```json
{
    "type": "input-year",
    "name": "fiscal_year",
    "label": "Fiscal Year",
    "placeholder": "Select fiscal year",
    "format": "YYYY",
    "shortcuts": ["thisYear", "lastYear", "nextYear"],
    "clearable": true
}
```

### Year Range Selector
```json
{
    "type": "input-year",
    "name": "planning_year",
    "label": "Planning Year",
    "placeholder": "Select planning year",
    "format": "YYYY",
    "minYear": 2020,
    "maxYear": 2030,
    "clearable": true
}
```

### Embedded Year Calendar
```json
{
    "type": "input-year",
    "name": "strategic_year",
    "label": "Strategic Planning Year",
    "embed": true,
    "format": "YYYY",
    "borderMode": "none"
}
```

## Real-World Use Cases

### Annual Budget Planning
```json
{
    "type": "input-year",
    "name": "budget_year",
    "label": "Budget Year",
    "placeholder": "Select budget planning year",
    "format": "YYYY",
    "shortcuts": [
        {"label": "Current Year", "value": "thisYear"},
        {"label": "Next Year", "value": "nextYear"},
        {"label": "2 Years Ahead", "value": "+2years"}
    ],
    "clearable": true,
    "required": true,
    "validations": {
        "isDateTimeAfter": ["${currentYear}", "year"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Budget year must be current year or later"
    },
    "autoFill": {
        "api": "/api/budget/yearly-template/${value}",
        "fillMapping": {
            "budget_template": "default_allocations",
            "previous_budget": "last_year_amounts",
            "growth_targets": "projected_growth"
        }
    }
}
```

### Annual Performance Review
```json
{
    "type": "input-year",
    "name": "performance_year",
    "label": "Performance Review Year",
    "placeholder": "Select performance review year",
    "format": "YYYY",
    "shortcuts": [
        {"label": "Current Year", "value": "thisYear"},
        {"label": "Last Year", "value": "lastYear"},
        {"label": "2 Years Ago", "value": "-2years"}
    ],
    "disabledDate": "moment(currentDate).isAfter(moment())",
    "clearable": true,
    "required": true,
    "autoFill": {
        "api": "/api/hr/annual-performance/${value}",
        "fillMapping": {
            "review_cycle": "annual_cycle",
            "employee_count": "total_employees",
            "completion_rate": "reviews_completed"
        }
    }
}
```

### Financial Year Reporting
```json
{
    "type": "input-year",
    "name": "financial_year",
    "label": "Financial Year",
    "placeholder": "Select financial reporting year",
    "format": "YYYY",
    "shortcuts": [
        {"label": "Current FY", "value": "thisFiscalYear"},
        {"label": "Previous FY", "value": "lastFiscalYear"},
        {"label": "FY-2", "value": "-2fiscalYears"}
    ],
    "clearable": true,
    "required": true,
    "autoFill": {
        "api": "/api/financial/yearly-summary/${value}",
        "fillMapping": {
            "total_revenue": "annual_revenue",
            "total_expenses": "annual_expenses",
            "net_profit": "annual_profit",
            "tax_liability": "annual_tax"
        }
    }
}
```

### Strategic Planning Year
```json
{
    "type": "input-year",
    "name": "strategic_planning_year",
    "label": "Strategic Planning Year",
    "placeholder": "Select strategic planning year",
    "format": "YYYY",
    "minYear": 2024,
    "maxYear": 2035,
    "shortcuts": [
        {"label": "5-Year Plan Start", "value": "+5years"},
        {"label": "10-Year Vision", "value": "+10years"},
        {"label": "Next Decade", "value": "+10years"}
    ],
    "validations": {
        "isDateTimeAfter": ["${currentYear}", "year"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Strategic planning must be for future years"
    }
}
```

### Tax Year Selection
```json
{
    "type": "input-year",
    "name": "tax_year",
    "label": "Tax Year",
    "placeholder": "Select tax year",
    "format": "YYYY",
    "shortcuts": [
        {"label": "Current Tax Year", "value": "thisTaxYear"},
        {"label": "Previous Tax Year", "value": "lastTaxYear"},
        {"label": "2 Years Ago", "value": "-2taxYears"}
    ],
    "clearable": true,
    "required": true,
    "autoFill": {
        "api": "/api/tax/yearly-filing/${value}",
        "fillMapping": {
            "filing_deadline": "tax_deadline",
            "estimated_liability": "tax_estimate",
            "previous_payments": "payments_made",
            "required_forms": "filing_forms"
        }
    }
}
```

### Contract Year Selection
```json
{
    "type": "input-year",
    "name": "contract_year",
    "label": "Contract Year",
    "placeholder": "Select contract year",
    "format": "YYYY",
    "minYear": 2020,
    "maxYear": 2030,
    "shortcuts": [
        {"label": "Contract Start", "value": "contractStartYear"},
        {"label": "Contract End", "value": "contractEndYear"},
        {"label": "Renewal Year", "value": "renewalYear"}
    ],
    "autoFill": {
        "api": "/api/contracts/yearly-terms/${value}",
        "fillMapping": {
            "annual_value": "contract_value",
            "renewal_terms": "renewal_conditions",
            "milestones": "yearly_milestones"
        }
    }
}
```

### Inventory Planning Year
```json
{
    "type": "input-year",
    "name": "inventory_planning_year",
    "label": "Inventory Planning Year",
    "placeholder": "Select inventory planning year",
    "format": "YYYY",
    "shortcuts": [
        {"label": "Next Year", "value": "nextYear"},
        {"label": "2 Years Ahead", "value": "+2years"},
        {"label": "3 Years Ahead", "value": "+3years"}
    ],
    "validations": {
        "isDateTimeSameOrAfter": ["${currentYear}", "year"]
    },
    "validationErrors": {
        "isDateTimeSameOrAfter": "Inventory planning must be for current year or later"
    }
}
```

### Audit Year Selection
```json
{
    "type": "input-year",
    "name": "audit_year",
    "label": "Audit Year",
    "placeholder": "Select audit year",
    "format": "YYYY",
    "shortcuts": [
        {"label": "Current Year", "value": "thisYear"},
        {"label": "Last Year", "value": "lastYear"},
        {"label": "2 Years Ago", "value": "-2years"}
    ],
    "disabledDate": "moment(currentDate).isAfter(moment())",
    "clearable": true,
    "required": true,
    "hint": "Select the year for audit procedures"
}
```

### Sales Target Year
```json
{
    "type": "input-year",
    "name": "sales_target_year",
    "label": "Sales Target Year",
    "placeholder": "Select sales target year",
    "format": "YYYY",
    "shortcuts": [
        {"label": "Current Year", "value": "thisYear"},
        {"label": "Next Year", "value": "nextYear"},
        {"label": "2-Year Target", "value": "+2years"}
    ],
    "validations": {
        "isDateTimeSameOrAfter": ["${currentYear}", "year"]
    },
    "validationErrors": {
        "isDateTimeSameOrAfter": "Sales targets must be set for current year or later"
    },
    "autoFill": {
        "api": "/api/sales/yearly-targets/${value}",
        "fillMapping": {
            "target_revenue": "annual_target",
            "growth_rate": "expected_growth",
            "team_quotas": "sales_quotas"
        }
    }
}
```

### Project Year Timeline
```json
{
    "type": "input-year",
    "name": "project_completion_year",
    "label": "Project Completion Year",
    "placeholder": "Select project completion year",
    "format": "YYYY",
    "minYear": 2024,
    "maxYear": 2035,
    "shortcuts": [
        {"label": "Next Year", "value": "nextYear"},
        {"label": "2 Years", "value": "+2years"},
        {"label": "5 Years", "value": "+5years"}
    ],
    "validations": {
        "isDateTimeAfter": ["${currentYear}", "year"]
    },
    "validationErrors": {
        "isDateTimeAfter": "Project completion must be in the future"
    }
}
```

### Depreciation Schedule Year
```json
{
    "type": "input-year",
    "name": "depreciation_year",
    "label": "Depreciation Year",
    "placeholder": "Select depreciation year",
    "format": "YYYY",
    "shortcuts": [
        {"label": "Asset Purchase Year", "value": "assetPurchaseYear"},
        {"label": "Current Year", "value": "thisYear"},
        {"label": "Final Year", "value": "assetFinalYear"}
    ],
    "autoFill": {
        "api": "/api/assets/depreciation-schedule/${value}",
        "fillMapping": {
            "depreciation_amount": "annual_depreciation",
            "book_value": "year_end_value",
            "remaining_life": "years_remaining"
        }
    }
}
```

This component provides essential year selection functionality for ERP systems requiring annual planning, reporting, and long-term business management capabilities.