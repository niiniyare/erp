# Range Control Component

**FILE PURPOSE**: Slider control for numeric value selection with visual range feedback  
**SCOPE**: Single and dual sliders, value ranges, step controls, and custom marks  
**TARGET AUDIENCE**: Developers implementing numeric inputs, filters, and range selection interfaces

## 📋 Component Overview

Range Control provides interactive slider functionality for selecting numeric values or ranges with visual feedback, step controls, marks, and value display options. Essential for numeric input scenarios in ERP forms.

### Schema Reference
- **Primary Schema**: `RangeControlSchema.json`
- **Related Schemas**: `FormHorizontal.json`, `TooltipPosType.json`
- **Base Interface**: Form input control for numeric range selection

## Basic Usage

```json
{
    "type": "input-range",
    "name": "quantity",
    "label": "Quantity",
    "min": 0,
    "max": 100,
    "step": 1
}
```

## Go Type Definition

```go
type RangeControlProps struct {
    Type              string              `json:"type"`
    Name              string              `json:"name"`
    Label             interface{}         `json:"label"`
    Min               interface{}         `json:"min"`              // number or expression
    Max               interface{}         `json:"max"`              // number or expression  
    Step              interface{}         `json:"step"`             // number or expression
    Value             interface{}         `json:"value"`
    Multiple          bool                `json:"multiple"`         // dual slider
    Unit              string              `json:"unit"`
    ShowSteps         bool                `json:"showSteps"`
    ShowInput         bool                `json:"showInput"`
    Clearable         bool                `json:"clearable"`
    Parts             interface{}         `json:"parts"`            // number or array
    Marks             map[string]interface{} `json:"marks"`
    TooltipVisible    bool                `json:"tooltipVisible"`
    TooltipPlacement  string              `json:"tooltipPlacement"`
    JoinValues        bool                `json:"joinValues"`
    Delimiter         string              `json:"delimiter"`
}
```

## Essential Variants

### Basic Range Slider
```json
{
    "type": "input-range",
    "name": "price",
    "label": "Price Range",
    "min": 0,
    "max": 1000,
    "step": 10,
    "unit": "$",
    "showInput": true,
    "tooltipVisible": true
}
```

### Dual Range Slider
```json
{
    "type": "input-range",
    "name": "price_range",
    "label": "Price Range",
    "min": 0,
    "max": 10000,
    "step": 100,
    "multiple": true,
    "unit": "$",
    "joinValues": true,
    "delimiter": "-",
    "showInput": true,
    "tooltipVisible": true
}
```

### Percentage Slider
```json
{
    "type": "input-range",
    "name": "completion",
    "label": "Completion Percentage",
    "min": 0,
    "max": 100,
    "step": 5,
    "unit": "%",
    "showSteps": true,
    "tooltipVisible": true,
    "marks": {
        "0": "0%",
        "25": "25%", 
        "50": "50%",
        "75": "75%",
        "100": "100%"
    }
}
```

### Custom Marked Slider
```json
{
    "type": "input-range",
    "name": "priority",
    "label": "Priority Level",
    "min": 1,
    "max": 5,
    "step": 1,
    "tooltipVisible": true,
    "marks": {
        "1": "Low",
        "2": "Medium-Low",
        "3": "Medium", 
        "4": "Medium-High",
        "5": "High"
    }
}
```

### Quantity Selector
```json
{
    "type": "input-range",
    "name": "quantity",
    "label": "Order Quantity",
    "min": 1,
    "max": 100,
    "step": 1,
    "showInput": true,
    "clearable": true,
    "tooltipVisible": true,
    "showSteps": false
}
```

## Real-World Use Cases

### Budget Range Filter
```json
{
    "type": "input-range",
    "name": "budget_range",
    "label": "Budget Range",
    "min": 0,
    "max": 100000,
    "step": 1000,
    "multiple": true,
    "unit": "$",
    "joinValues": true,
    "delimiter": " - ",
    "showInput": true,
    "tooltipVisible": true,
    "marks": {
        "0": "$0",
        "25000": "$25K",
        "50000": "$50K", 
        "75000": "$75K",
        "100000": "$100K"
    }
}
```

### Employee Performance Rating
```json
{
    "type": "input-range",
    "name": "performance_rating",
    "label": "Performance Rating",
    "min": 1,
    "max": 10,
    "step": 0.5,
    "unit": "/10",
    "tooltipVisible": true,
    "showInput": false,
    "marks": {
        "1": "Poor",
        "3": "Below Average",
        "5": "Average",
        "7": "Good", 
        "9": "Excellent",
        "10": "Outstanding"
    }
}
```

### Project Timeline Allocation
```json
{
    "type": "input-range",
    "name": "timeline_allocation",
    "label": "Timeline Allocation",
    "min": 0,
    "max": 365,
    "step": 7,
    "unit": " days",
    "multiple": true,
    "joinValues": false,
    "showInput": true,
    "tooltipVisible": true,
    "marks": {
        "0": "Start",
        "90": "Q1",
        "180": "Q2",
        "270": "Q3",
        "365": "Q4"
    }
}
```

### Inventory Level Control
```json
{
    "type": "input-range",
    "name": "stock_level",
    "label": "Stock Level Alert",
    "min": 0,
    "max": 1000,
    "step": 10,
    "unit": " units",
    "multiple": true,
    "joinValues": true,
    "delimiter": " to ",
    "showInput": true,
    "tooltipVisible": true,
    "validations": {
        "minimum": 10
    },
    "validationErrors": {
        "minimum": "Minimum stock level must be at least 10 units"
    }
}
```

### Discount Percentage
```json
{
    "type": "input-range",
    "name": "discount_percent",
    "label": "Discount Percentage",
    "min": 0,
    "max": 50,
    "step": 2.5,
    "unit": "%",
    "showInput": true,
    "tooltipVisible": true,
    "showSteps": true,
    "marks": {
        "0": "0%",
        "10": "10%",
        "20": "20%",
        "30": "30%",
        "40": "40%",
        "50": "50%"
    }
}
```

### Age Range Filter
```json
{
    "type": "input-range",
    "name": "age_range",
    "label": "Age Range",
    "min": 18,
    "max": 65,
    "step": 1,
    "multiple": true,
    "unit": " years",
    "joinValues": true,
    "delimiter": "-",
    "showInput": true,
    "tooltipVisible": true,
    "marks": {
        "18": "18",
        "25": "25",
        "35": "35",
        "45": "45",
        "55": "55",
        "65": "65"
    }
}
```

This component provides essential range selection functionality for ERP forms requiring numeric value input with visual feedback and precise control.