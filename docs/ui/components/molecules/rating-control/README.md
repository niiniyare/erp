# Rating Control Component

**FILE PURPOSE**: Star rating input control for user feedback and quality assessment  
**SCOPE**: Star ratings, half-star support, custom characters, colors, and text feedback  
**TARGET AUDIENCE**: Developers implementing review systems, feedback forms, and quality assessment interfaces

## 📋 Component Overview

Rating Control provides interactive star rating functionality with support for half-stars, custom characters, colors, text feedback, and read-only display modes. Essential for collecting user ratings and feedback in ERP systems.

### Schema Reference
- **Primary Schema**: `RatingControlSchema.json`
- **Related Schemas**: `FormHorizontal.json`, `textPositionType.json`
- **Base Interface**: Form input control for star rating selection

## Basic Usage

```json
{
    "type": "input-rating",
    "name": "rating",
    "label": "Rating",
    "count": 5
}
```

## Go Type Definition

```go
type RatingControlProps struct {
    Type            string              `json:"type"`
    Name            string              `json:"name"`
    Label           interface{}         `json:"label"`
    Count           int                 `json:"count"`           // Number of stars
    Half            bool                `json:"half"`            // Allow half stars
    AllowClear      bool                `json:"allowClear"`      // Allow clearing rating
    ReadOnly        bool                `json:"readonly"`        // Read-only mode
    Value           interface{}         `json:"value"`
    Colors          interface{}         `json:"colors"`          // string or map
    InactiveColor   string              `json:"inactiveColor"`   // Unselected star color
    Texts           map[string]string   `json:"texts"`           // Rating text labels
    TextPosition    string              `json:"textPosition"`    // Text position
    Char            string              `json:"char"`            // Custom character
    CharClassName   string              `json:"charClassName"`   // Custom character CSS
    TextClassName   string              `json:"textClassName"`   // Text CSS class
}
```

## Essential Variants

### Basic Star Rating
```json
{
    "type": "input-rating",
    "name": "product_rating",
    "label": "Product Rating",
    "count": 5,
    "allowClear": true,
    "required": true
}
```

### Half-Star Rating
```json
{
    "type": "input-rating",
    "name": "service_rating",
    "label": "Service Quality",
    "count": 5,
    "half": true,
    "allowClear": true,
    "texts": {
        "1": "Poor",
        "2": "Fair", 
        "3": "Good",
        "4": "Very Good",
        "5": "Excellent"
    },
    "textPosition": "right"
}
```

### Custom Colors Rating
```json
{
    "type": "input-rating",
    "name": "priority_rating",
    "label": "Priority Level",
    "count": 3,
    "colors": {
        "1": "#52c41a",
        "2": "#faad14", 
        "3": "#f5222d"
    },
    "inactiveColor": "#d9d9d9",
    "texts": {
        "1": "Low Priority",
        "2": "Medium Priority",
        "3": "High Priority"
    }
}
```

### Custom Character Rating
```json
{
    "type": "input-rating",
    "name": "satisfaction",
    "label": "Satisfaction Level",
    "count": 5,
    "char": "😊",
    "allowClear": true,
    "texts": {
        "1": "Very Unsatisfied",
        "2": "Unsatisfied",
        "3": "Neutral", 
        "4": "Satisfied",
        "5": "Very Satisfied"
    }
}
```

### Read-Only Display
```json
{
    "type": "input-rating",
    "name": "average_rating",
    "label": "Average Rating",
    "count": 5,
    "half": true,
    "readonly": true,
    "value": 4.3,
    "colors": "#faad14"
}
```

## Real-World Use Cases

### Employee Performance Review
```json
{
    "type": "input-rating",
    "name": "performance_score",
    "label": "Overall Performance",
    "count": 10,
    "half": true,
    "allowClear": false,
    "required": true,
    "colors": {
        "1": "#ff4d4f",
        "2": "#ff4d4f", 
        "3": "#ff7a45",
        "4": "#ffa940",
        "5": "#ffc53d",
        "6": "#fadb14",
        "7": "#a0d911",
        "8": "#52c41a",
        "9": "#13c2c2",
        "10": "#1890ff"
    },
    "texts": {
        "1": "Needs Significant Improvement",
        "2": "Below Expectations",
        "3": "Approaching Expectations",
        "4": "Meets Some Expectations",
        "5": "Meets Expectations",
        "6": "Above Average",
        "7": "Good Performance",
        "8": "Very Good Performance", 
        "9": "Excellent Performance",
        "10": "Outstanding Performance"
    },
    "textPosition": "bottom"
}
```

### Product Quality Assessment
```json
{
    "type": "input-rating",
    "name": "quality_rating",
    "label": "Product Quality",
    "count": 5,
    "half": true,
    "allowClear": true,
    "colors": "#faad14",
    "inactiveColor": "#f0f0f0",
    "texts": {
        "0.5": "Terrible",
        "1": "Poor",
        "1.5": "Below Average",
        "2": "Fair",
        "2.5": "Average",
        "3": "Good",
        "3.5": "Very Good",
        "4": "Great",
        "4.5": "Nearly Perfect",
        "5": "Perfect"
    },
    "textPosition": "right",
    "hint": "Rate the overall quality of this product"
}
```

### Customer Service Feedback
```json
{
    "type": "input-rating",
    "name": "service_feedback",
    "label": "How was our service?",
    "count": 5,
    "char": "⭐",
    "allowClear": true,
    "colors": "#ff6b6b",
    "texts": {
        "1": "Very Poor Service",
        "2": "Poor Service",
        "3": "Average Service",
        "4": "Good Service", 
        "5": "Excellent Service"
    },
    "textPosition": "bottom",
    "textClassName": "text-sm text-gray-600"
}
```

### Difficulty Level Rating
```json
{
    "type": "input-rating",
    "name": "difficulty_level",
    "label": "Task Difficulty",
    "count": 5,
    "char": "🔥",
    "allowClear": true,
    "colors": {
        "1": "#52c41a",
        "2": "#87d068",
        "3": "#faad14",
        "4": "#ff7a45",
        "5": "#f5222d"
    },
    "texts": {
        "1": "Very Easy",
        "2": "Easy", 
        "3": "Moderate",
        "4": "Hard",
        "5": "Very Hard"
    },
    "textPosition": "left"
}
```

### Training Course Evaluation
```json
{
    "type": "input-rating",
    "name": "course_rating",
    "label": "Course Rating",
    "count": 5,
    "half": true,
    "allowClear": true,
    "required": true,
    "colors": "#1890ff",
    "inactiveColor": "#d9d9d9",
    "texts": {
        "1": "Not Helpful",
        "2": "Somewhat Helpful",
        "3": "Moderately Helpful",
        "4": "Very Helpful",
        "5": "Extremely Helpful"
    },
    "validations": {
        "minimum": 1
    },
    "validationErrors": {
        "minimum": "Please provide a rating for this course"
    }
}
```

### Vendor Performance Score
```json
{
    "type": "input-rating",
    "name": "vendor_score",
    "label": "Vendor Performance",
    "count": 4,
    "allowClear": false,
    "colors": {
        "1": "#f5222d",
        "2": "#faad14", 
        "3": "#52c41a",
        "4": "#1890ff"
    },
    "texts": {
        "1": "Poor Performance",
        "2": "Needs Improvement",
        "3": "Good Performance", 
        "4": "Excellent Performance"
    },
    "textPosition": "right",
    "hint": "Rate overall vendor performance including delivery, quality, and communication"
}
```

This component provides essential rating input functionality for ERP feedback systems, performance reviews, and quality assessment scenarios.