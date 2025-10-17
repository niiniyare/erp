# DateRange Component

## Overview

The DateRange component provides a read-only display for date ranges with flexible formatting options, separators, and connectors. It's designed for displaying date range data rather than input collection.

## Basic Usage

### Simple Date Range Display
```json
{
  "type": "date-range",
  "valueFormat": "YYYY-MM-DD",
  "format": "YYYY-MM-DD",
  "delimiter": " to "
}
```

### Formatted Date Range
```json
{
  "type": "date-range",
  "valueFormat": "YYYY-MM-DD",
  "displayFormat": "MMM DD, YYYY",
  "connector": " - "
}
```

### Custom Separator
```json
{
  "type": "date-range",
  "valueFormat": "YYYY-MM-DD",
  "format": "DD/MM/YYYY",
  "delimiter": " → "
}
```

## Complete Form Examples

### Report Date Range Display
```json
{
  "type": "form",
  "title": "Sales Report",
  "body": [
    {
      "type": "static",
      "label": "Report Period",
      "name": "report_period",
      "value": "2024-01-01,2024-03-31"
    },
    {
      "type": "date-range",
      "className": "mt-2",
      "valueFormat": "YYYY-MM-DD",
      "displayFormat": "MMMM DD, YYYY",
      "delimiter": " through ",
      "style": {
        "fontSize": "16px",
        "fontWeight": "500",
        "color": "#374151"
      }
    },
    {
      "type": "divider"
    },
    {
      "type": "chart",
      "api": "/api/sales-data?start=${report_period.start}&end=${report_period.end}",
      "config": {
        "type": "line",
        "xField": "date",
        "yField": "sales"
      }
    }
  ]
}
```

### Event Duration Display
```json
{
  "type": "card",
  "header": {
    "title": "Conference Schedule",
    "subTitle": "Annual Technology Summit"
  },
  "body": [
    {
      "type": "grid",
      "columns": [
        {
          "body": [
            {
              "type": "static",
              "label": "Conference Dates",
              "value": "2024-06-15,2024-06-17"
            }
          ]
        },
        {
          "body": [
            {
              "type": "date-range",
              "valueFormat": "YYYY-MM-DD",
              "displayFormat": "dddd, MMMM DD",
              "connector": " to ",
              "className": "text-lg font-semibold text-blue-600"
            }
          ]
        }
      ]
    },
    {
      "type": "table",
      "api": "/api/conference-sessions",
      "columns": [
        {"name": "title", "label": "Session"},
        {"name": "speaker", "label": "Speaker"},
        {"name": "time", "label": "Time", "type": "time"}
      ]
    }
  ]
}
```

### Project Timeline Display
```json
{
  "type": "page",
  "title": "Project Overview",
  "body": [
    {
      "type": "cards",
      "api": "/api/projects",
      "card": {
        "header": {
          "title": "${name}",
          "subTitle": "${status}"
        },
        "body": [
          {
            "type": "static",
            "label": "Project Duration",
            "value": "${start_date},${end_date}"
          },
          {
            "type": "date-range",
            "valueFormat": "YYYY-MM-DD",
            "format": "MMM DD, YYYY",
            "delimiter": " — ",
            "className": "mt-2 text-sm text-gray-600"
          },
          {
            "type": "progress",
            "value": "${completion_percentage}",
            "className": "mt-3"
          }
        ],
        "actions": [
          {
            "type": "button",
            "label": "View Details",
            "actionType": "link",
            "link": "/projects/${id}"
          }
        ]
      }
    }
  ]
}
```

### Financial Period Display
```json
{
  "type": "form",
  "title": "Financial Report Generator",
  "api": "/api/reports/generate",
  "body": [
    {
      "type": "select",
      "name": "period_type",
      "label": "Reporting Period",
      "options": [
        {"label": "Monthly", "value": "monthly"},
        {"label": "Quarterly", "value": "quarterly"},
        {"label": "Annually", "value": "annually"},
        {"label": "Custom Range", "value": "custom"}
      ]
    },
    {
      "type": "input-date-range",
      "name": "custom_period",
      "label": "Custom Date Range",
      "visibleOn": "${period_type === 'custom'}",
      "format": "YYYY-MM-DD"
    },
    {
      "type": "static",
      "label": "Selected Period",
      "value": "${custom_period}",
      "visibleOn": "${period_type === 'custom'}"
    },
    {
      "type": "date-range",
      "valueFormat": "YYYY-MM-DD",
      "displayFormat": "MMMM DD, YYYY",
      "connector": " to ",
      "visibleOn": "${period_type === 'custom'}",
      "className": "bg-blue-50 p-3 rounded-lg mt-2"
    }
  ]
}
```

### Vacation Request Display
```json
{
  "type": "table",
  "title": "Vacation Requests",
  "api": "/api/hr/vacation-requests",
  "columns": [
    {"name": "employee_name", "label": "Employee"},
    {"name": "request_type", "label": "Type"},
    {
      "name": "vacation_period",
      "label": "Vacation Period",
      "type": "date-range",
      "valueFormat": "YYYY-MM-DD",
      "displayFormat": "MMM DD",
      "delimiter": " – "
    },
    {"name": "status", "label": "Status", "type": "status"},
    {"name": "days_count", "label": "Days", "type": "number"}
  ],
  "headerToolbar": [
    {
      "type": "button",
      "label": "New Request",
      "actionType": "dialog",
      "dialog": {
        "title": "Vacation Request",
        "body": {
          "type": "form",
          "api": "/api/hr/vacation-requests",
          "body": [
            {
              "type": "input-date-range",
              "name": "vacation_dates",
              "label": "Vacation Dates",
              "required": true
            }
          ]
        }
      }
    }
  ]
}
```

## Property Reference

### Core Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `type` | `string` | - | **Required.** Must be `"date-range"` |
| `valueFormat` | `string` | `"YYYY-MM-DD"` | Format of the input date values |
| `format` | `string` | `"YYYY-MM-DD"` | Display format for the dates |
| `displayFormat` | `string` | - | Alternative display format (overrides format) |

### Separator Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `delimiter` | `string` | `" to "` | Text displayed between start and end dates |
| `connector` | `string` | - | Alternative separator (overrides delimiter) |

### Style & Appearance

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `className` | `string` | - | Additional CSS classes |
| `style` | `object` | - | Inline styles |

### State Management

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `disabled` | `boolean` | `false` | Disable the component |
| `hidden` | `boolean` | `false` | Hide the component |
| `visible` | `boolean` | `true` | Component visibility |
| `static` | `boolean` | `false` | Static display mode |

### Common Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `id` | `string` | - | Unique component identifier |
| `testid` | `string` | - | Test automation identifier |

## Date Format Reference

### Common Format Patterns

| Pattern | Example | Description |
|---------|---------|-------------|
| `YYYY-MM-DD` | `2024-03-15` | ISO date format |
| `MM/DD/YYYY` | `03/15/2024` | US date format |
| `DD/MM/YYYY` | `15/03/2024` | European date format |
| `MMM DD, YYYY` | `Mar 15, 2024` | Month abbreviated |
| `MMMM DD, YYYY` | `March 15, 2024` | Month full name |
| `dddd, MMMM DD` | `Friday, March 15` | Day and month |
| `DD MMM YYYY` | `15 Mar 2024` | Compact format |

### Separator Options

| Separator | Example | Use Case |
|-----------|---------|----------|
| `" to "` | `Mar 15 to Mar 20` | Standard range |
| `" - "` | `Mar 15 - Mar 20` | Compact range |
| `" → "` | `Mar 15 → Mar 20` | Modern arrow |
| `" — "` | `Mar 15 — Mar 20` | Em dash |
| `" through "` | `Mar 15 through Mar 20` | Formal language |
| `" – "` | `Mar 15 – Mar 20` | En dash |

## Styling & Theming

### CSS Classes

- `.date-range` - Base date range container
- `.date-range__start` - Start date display
- `.date-range__separator` - Separator element
- `.date-range__end` - End date display
- `.date-range--disabled` - Disabled state styling

### Custom Styling Examples

```json
{
  "type": "date-range",
  "valueFormat": "YYYY-MM-DD",
  "displayFormat": "MMM DD, YYYY",
  "delimiter": " to ",
  "className": "custom-date-range",
  "style": {
    "padding": "8px 12px",
    "backgroundColor": "#f8fafc",
    "borderRadius": "6px",
    "border": "1px solid #e2e8f0",
    "fontSize": "14px",
    "fontWeight": "500"
  }
}
```

### Responsive Styling
```json
{
  "type": "date-range",
  "valueFormat": "YYYY-MM-DD",
  "format": "MMM DD, YYYY",
  "delimiter": " to ",
  "className": "text-sm md:text-base lg:text-lg font-medium text-gray-700"
}
```

## Integration Patterns

### With Data Sources
```json
{
  "type": "date-range",
  "valueFormat": "YYYY-MM-DD",
  "displayFormat": "MMMM DD, YYYY",
  "connector": " through ",
  "className": "report-period"
}
```

### With Conditional Display
```json
{
  "type": "date-range",
  "valueFormat": "YYYY-MM-DD",
  "format": "MMM DD",
  "delimiter": " – ",
  "visibleOn": "${start_date && end_date}",
  "hiddenOn": "${status === 'draft'}"
}
```

### With Dynamic Formatting
```json
{
  "type": "date-range",
  "valueFormat": "YYYY-MM-DD",
  "format": "${is_same_year ? 'MMM DD' : 'MMM DD, YYYY'}",
  "delimiter": " to "
}
```

## Accessibility

### ARIA Support
- `role="text"` for semantic meaning
- `aria-label` for screen readers
- Proper date announcements

### Screen Reader Considerations
- Dates are announced in readable format
- Separators provide context
- Time relationships are clear

## Best Practices

### Date Format Selection
- Use consistent formats across the application
- Consider user locale and preferences
- Match format to data precision needed

### Separator Usage
- Use familiar separators for your audience
- Maintain consistency within views
- Consider context (formal vs. casual)

### Performance
- Format dates on the server when possible
- Cache formatted results for repeated displays
- Use appropriate precision for display

### Accessibility
- Provide clear date relationships
- Use readable format for screen readers
- Include proper semantic markup

## Go Type Definition

```go
// DateRangeComponent represents a date range display component
type DateRangeComponent struct {
    BaseComponent
    
    // Format Properties
    ValueFormat   string `json:"valueFormat,omitempty"`
    Format        string `json:"format,omitempty"`
    DisplayFormat string `json:"displayFormat,omitempty"`
    
    // Separator Properties
    Delimiter string `json:"delimiter,omitempty"`
    Connector string `json:"connector,omitempty"`
    
    // Style Properties
    ClassName string                 `json:"className,omitempty"`
    Style     map[string]interface{} `json:"style,omitempty"`
    
    // State Properties
    Disabled bool `json:"disabled,omitempty"`
    Hidden   bool `json:"hidden,omitempty"`
    Visible  bool `json:"visible,omitempty"`
    Static   bool `json:"static,omitempty"`
}

// DateRangeFactory creates DateRange components from JSON configuration
func DateRangeFactory(config map[string]interface{}) (*DateRangeComponent, error) {
    component := &DateRangeComponent{
        BaseComponent: BaseComponent{
            Type: "date-range",
        },
        ValueFormat: "YYYY-MM-DD",
        Format:      "YYYY-MM-DD",
        Delimiter:   " to ",
    }
    
    return component, mapConfig(config, component)
}

// Render generates the Templ template for the date range display
func (c *DateRangeComponent) Render() templ.Component {
    return daterange.DateRange(daterange.DateRangeProps{
        ValueFormat:   c.ValueFormat,
        Format:        c.Format,
        DisplayFormat: c.DisplayFormat,
        Delimiter:     c.Delimiter,
        Connector:     c.Connector,
        ClassName:     c.ClassName,
        Style:         c.Style,
        Disabled:      c.Disabled,
        Hidden:        c.Hidden,
        Visible:       c.Visible,
        Static:        c.Static,
    })
}

// FormatDateRange formats a date range string according to component settings
func (c *DateRangeComponent) FormatDateRange(startDate, endDate string) (string, error) {
    // Parse input dates using ValueFormat
    start, err := time.Parse(c.ValueFormat, startDate)
    if err != nil {
        return "", fmt.Errorf("invalid start date: %w", err)
    }
    
    end, err := time.Parse(c.ValueFormat, endDate)
    if err != nil {
        return "", fmt.Errorf("invalid end date: %w", err)
    }
    
    // Determine display format
    displayFormat := c.Format
    if c.DisplayFormat != "" {
        displayFormat = c.DisplayFormat
    }
    
    // Determine separator
    separator := c.Delimiter
    if c.Connector != "" {
        separator = c.Connector
    }
    
    // Format and combine dates
    formattedStart := start.Format(displayFormat)
    formattedEnd := end.Format(displayFormat)
    
    return formattedStart + separator + formattedEnd, nil
}
```

## Related Components

- **[Date](../atoms/date/)** - Single date display component
- **[Input Date Range](../atoms/input-date-range/)** - Date range input component
- **[Static](../atoms/static/)** - Static text display component
- **[Time](../atoms/time/)** - Time display component
- **[Calendar](../molecules/calendar/)** - Calendar picker component