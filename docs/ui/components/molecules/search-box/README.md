# SearchBox Component

## Overview

The SearchBox component provides a comprehensive search interface with immediate search capabilities, loading states, and enhanced styling options. It supports both mini and full-size variations with configurable search behavior.

## Basic Usage

### Simple Search Box
```json
{
  "type": "search-box",
  "name": "query",
  "placeholder": "Search products..."
}
```

### Enhanced Search Box
```json
{
  "type": "search-box",
  "name": "user_search",
  "placeholder": "Search users by name or email",
  "enhance": true,
  "clearable": true,
  "searchImediately": true
}
```

### Mini Search Box
```json
{
  "type": "search-box",
  "name": "quick_search",
  "placeholder": "Quick search",
  "mini": true,
  "clearable": true
}
```

## Complete Form Examples

### Product Search Interface
```json
{
  "type": "form",
  "api": "/api/products/search",
  "body": [
    {
      "type": "search-box",
      "name": "product_query",
      "placeholder": "Search products by name, SKU, or category",
      "enhance": true,
      "clearable": true,
      "searchImediately": true,
      "clearAndSubmit": true,
      "className": "mb-4"
    },
    {
      "type": "table",
      "source": "${products}",
      "columns": [
        {"name": "name", "label": "Product Name"},
        {"name": "sku", "label": "SKU"},
        {"name": "category", "label": "Category"},
        {"name": "price", "label": "Price", "type": "number"}
      ]
    }
  ]
}
```

### User Directory Search
```json
{
  "type": "page",
  "body": [
    {
      "type": "search-box",
      "name": "user_search",
      "placeholder": "Search users by name, email, or department",
      "enhance": true,
      "clearable": true,
      "searchImediately": false,
      "className": "search-header",
      "onEvent": {
        "search": {
          "actions": [
            {
              "actionType": "reload",
              "componentId": "user-list",
              "data": {
                "query": "${user_search}"
              }
            }
          ]
        }
      }
    },
    {
      "type": "cards",
      "id": "user-list",
      "api": "/api/users/search?query=${user_search}",
      "card": {
        "header": {
          "title": "${name}",
          "subTitle": "${email}"
        },
        "body": "${department}"
      }
    }
  ]
}
```

### Advanced Search with Loading States
```json
{
  "type": "search-box",
  "name": "advanced_search",
  "placeholder": "Search across all documents",
  "enhance": true,
  "clearable": true,
  "searchImediately": true,
  "loading": "${searchInProgress}",
  "disabled": "${searchDisabled}",
  "className": "advanced-search-box",
  "style": {
    "marginBottom": "20px"
  },
  "onEvent": {
    "search": {
      "actions": [
        {
          "actionType": "ajax",
          "api": "/api/search/documents",
          "data": {
            "query": "${advanced_search}",
            "type": "full_text"
          }
        }
      ],
      "debounce": {
        "wait": 300
      }
    },
    "clear": {
      "actions": [
        {
          "actionType": "setValue",
          "componentId": "search-results",
          "value": []
        }
      ]
    }
  }
}
```

### Multi-Field Search Form
```json
{
  "type": "form",
  "title": "Advanced Search",
  "api": "/api/search",
  "body": [
    {
      "type": "grid",
      "columns": [
        {
          "body": [
            {
              "type": "search-box",
              "name": "global_search",
              "label": "Global Search",
              "placeholder": "Search everything...",
              "enhance": true,
              "clearable": true,
              "searchImediately": true
            }
          ]
        },
        {
          "body": [
            {
              "type": "search-box",
              "name": "category_search",
              "label": "Category Search",
              "placeholder": "Search categories...",
              "mini": true,
              "clearable": true
            }
          ]
        }
      ]
    },
    {
      "type": "divider"
    },
    {
      "type": "search-box",
      "name": "quick_filter",
      "placeholder": "Quick filter results",
      "mini": true,
      "clearable": true,
      "clearAndSubmit": true,
      "className": "mt-3"
    }
  ]
}
```

## Property Reference

### Core Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `type` | `string` | - | **Required.** Must be `"search-box"` |
| `name` | `string` | - | Field name for form submission |
| `placeholder` | `string` | - | Placeholder text displayed in search input |

### Style & Appearance

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `mini` | `boolean` | `false` | Enable compact mini style |
| `enhance` | `boolean` | `false` | Enable enhanced visual styling |
| `className` | `string` | - | Additional CSS classes |
| `style` | `object` | - | Inline styles |

### Behavior Control

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `clearable` | `boolean` | `false` | Show clear button when input has value |
| `searchImediately` | `boolean` | `false` | Trigger search on every keystroke |
| `clearAndSubmit` | `boolean` | `false` | Auto-search again after clearing content |

### State Management

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `loading` | `boolean` | `false` | Display loading spinner |
| `disabled` | `boolean` | `false` | Disable the search input |
| `hidden` | `boolean` | `false` | Hide the component |

### Common Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `id` | `string` | - | Unique component identifier |
| `testid` | `string` | - | Test automation identifier |
| `visible` | `boolean` | `true` | Component visibility |
| `static` | `boolean` | `false` | Static display mode |

## Event Handling

### Available Events

| Event | Description | Data |
|-------|-------------|------|
| `search` | Triggered when search is performed | `{value: string}` |
| `clear` | Triggered when clear button is clicked | `{value: ""}` |
| `change` | Triggered when input value changes | `{value: string}` |
| `focus` | Triggered when input gains focus | `{value: string}` |
| `blur` | Triggered when input loses focus | `{value: string}` |

### Event Configuration Examples

```json
{
  "type": "search-box",
  "name": "product_search",
  "onEvent": {
    "search": {
      "actions": [
        {
          "actionType": "reload",
          "componentId": "product-table",
          "data": {"query": "${product_search}"}
        }
      ],
      "debounce": {"wait": 500}
    },
    "clear": {
      "actions": [
        {
          "actionType": "setValue",
          "componentId": "product-table",
          "value": {"products": []}
        }
      ]
    }
  }
}
```

## Styling & Theming

### CSS Classes

- `.search-box` - Base search box container
- `.search-box--mini` - Mini variant styling
- `.search-box--enhanced` - Enhanced variant styling
- `.search-box__input` - Search input field
- `.search-box__clear` - Clear button
- `.search-box__loading` - Loading indicator

### Custom Styling Example

```json
{
  "type": "search-box",
  "name": "styled_search",
  "placeholder": "Custom styled search",
  "enhance": true,
  "className": "custom-search-box",
  "style": {
    "borderRadius": "8px",
    "boxShadow": "0 2px 4px rgba(0,0,0,0.1)",
    "fontSize": "16px"
  }
}
```

## Integration Patterns

### With Data Tables
```json
{
  "type": "search-box",
  "name": "table_search",
  "placeholder": "Search table data...",
  "searchImediately": true,
  "clearable": true,
  "onEvent": {
    "search": {
      "actions": [
        {
          "actionType": "reload",
          "componentId": "data-table",
          "data": {"search": "${table_search}"}
        }
      ],
      "debounce": {"wait": 300}
    }
  }
}
```

### With API Endpoints
```json
{
  "type": "search-box",
  "name": "api_search",
  "placeholder": "Search API data...",
  "enhance": true,
  "onEvent": {
    "search": {
      "actions": [
        {
          "actionType": "ajax",
          "api": {
            "url": "/api/search",
            "method": "GET",
            "data": {"q": "${api_search}"}
          }
        }
      ]
    }
  }
}
```

## Accessibility

### ARIA Support
- `aria-label` for screen readers
- `role="searchbox"` for semantic meaning
- Keyboard navigation support
- Focus management

### Keyboard Shortcuts
- `Enter` - Trigger search
- `Escape` - Clear search (when clearable)
- `Tab` - Navigate to next element

## Best Practices

### Performance
- Use `debounce` for immediate search to reduce API calls
- Implement proper loading states for better UX
- Consider pagination for large result sets

### User Experience
- Provide clear placeholder text
- Show loading indicators during search
- Enable clearing for easy reset
- Use appropriate sizing (mini vs. full)

### Accessibility
- Always provide meaningful labels
- Support keyboard navigation
- Maintain focus management
- Use proper ARIA attributes

## Go Type Definition

```go
// SearchBoxComponent represents a search box input with enhanced features
type SearchBoxComponent struct {
    BaseComponent
    
    // Core Properties
    Name        string `json:"name,omitempty"`
    Placeholder string `json:"placeholder,omitempty"`
    
    // Style Properties
    Mini      bool   `json:"mini,omitempty"`
    Enhance   bool   `json:"enhance,omitempty"`
    ClassName string `json:"className,omitempty"`
    Style     map[string]interface{} `json:"style,omitempty"`
    
    // Behavior Properties
    Clearable        bool `json:"clearable,omitempty"`
    SearchImediately bool `json:"searchImediately,omitempty"`
    ClearAndSubmit   bool `json:"clearAndSubmit,omitempty"`
    
    // State Properties
    Loading  bool `json:"loading,omitempty"`
    Disabled bool `json:"disabled,omitempty"`
    Hidden   bool `json:"hidden,omitempty"`
    
    // Event Configuration
    OnEvent map[string]EventConfig `json:"onEvent,omitempty"`
}

// SearchBoxFactory creates SearchBox components from JSON configuration
func SearchBoxFactory(config map[string]interface{}) (*SearchBoxComponent, error) {
    component := &SearchBoxComponent{
        BaseComponent: BaseComponent{
            Type: "search-box",
        },
    }
    
    return component, mapConfig(config, component)
}

// Render generates the Templ template for the search box
func (c *SearchBoxComponent) Render() templ.Component {
    return searchbox.SearchBox(searchbox.SearchBoxProps{
        Name:             c.Name,
        Placeholder:      c.Placeholder,
        Mini:             c.Mini,
        Enhance:          c.Enhance,
        Clearable:        c.Clearable,
        SearchImediately: c.SearchImediately,
        ClearAndSubmit:   c.ClearAndSubmit,
        Loading:          c.Loading,
        Disabled:         c.Disabled,
        ClassName:        c.ClassName,
        Style:           c.Style,
        OnEvent:         c.OnEvent,
    })
}
```

## Related Components

- **[Input](../atoms/input/)** - Basic text input component
- **[Select](../atoms/select/)** - Dropdown selection component  
- **[Button](../atoms/button/)** - Action button component
- **[Form](../../organisms/form/)** - Complete form container
- **[Table](../../organisms/table/)** - Data table component