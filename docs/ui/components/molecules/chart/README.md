# Chart Component

## Overview

The Chart component provides powerful data visualization capabilities using ECharts integration. It supports various chart types, real-time data updates, interactive features, and comprehensive theming options.

## Basic Usage

### Simple Line Chart
```json
{
  "type": "chart",
  "api": "/api/sales-data",
  "config": {
    "type": "line",
    "xField": "date",
    "yField": "sales"
  }
}
```

### Bar Chart with Data
```json
{
  "type": "chart",
  "source": "${chartData}",
  "config": {
    "type": "bar",
    "xField": "category",
    "yField": "value",
    "color": ["#1f77b4", "#ff7f0e", "#2ca02c"]
  }
}
```

### Pie Chart
```json
{
  "type": "chart",
  "api": "/api/revenue-breakdown",
  "config": {
    "type": "pie",
    "angleField": "value",
    "colorField": "category",
    "radius": 0.8
  }
}
```

## Complete Form Examples

### Sales Dashboard
```json
{
  "type": "page",
  "title": "Sales Dashboard",
  "body": [
    {
      "type": "grid",
      "columns": [
        {
          "md": 8,
          "body": [
            {
              "type": "chart",
              "api": "/api/sales/monthly-trend",
              "height": 400,
              "config": {
                "type": "line",
                "xField": "month",
                "yField": "sales",
                "smooth": true,
                "point": {
                  "size": 5,
                  "shape": "diamond"
                }
              },
              "interval": 30000
            }
          ]
        },
        {
          "md": 4,
          "body": [
            {
              "type": "chart",
              "api": "/api/sales/by-region",
              "height": 200,
              "config": {
                "type": "pie",
                "angleField": "sales",
                "colorField": "region",
                "radius": 0.9,
                "label": {
                  "type": "inner",
                  "offset": "-30%",
                  "content": "{value}",
                  "style": {
                    "fill": "white",
                    "fontSize": 14,
                    "fontWeight": "bold"
                  }
                }
              }
            }
          ]
        }
      ]
    }
  ]
}
```

### Financial Analytics
```json
{
  "type": "form",
  "title": "Financial Analytics",
  "body": [
    {
      "type": "input-date-range",
      "name": "date_range",
      "label": "Date Range",
      "value": "2024-01-01,2024-12-31"
    },
    {
      "type": "select",
      "name": "chart_type",
      "label": "Chart Type",
      "value": "line",
      "options": [
        {"label": "Line Chart", "value": "line"},
        {"label": "Bar Chart", "value": "bar"},
        {"label": "Area Chart", "value": "area"}
      ]
    },
    {
      "type": "chart",
      "api": "/api/financial/data?start=${date_range.start}&end=${date_range.end}",
      "height": 500,
      "config": {
        "type": "${chart_type}",
        "xField": "date",
        "yField": "amount",
        "seriesField": "category",
        "smooth": true
      },
      "trackExpression": "${date_range} ${chart_type}",
      "initFetch": true,
      "onEvent": {
        "click": {
          "actions": [
            {
              "actionType": "dialog",
              "dialog": {
                "title": "Data Point Details",
                "body": {
                  "type": "table",
                  "api": "/api/financial/details?date=${date}&category=${category}",
                  "columns": [
                    {"name": "transaction_id", "label": "ID"},
                    {"name": "description", "label": "Description"},
                    {"name": "amount", "label": "Amount", "type": "number"}
                  ]
                }
              }
            }
          ]
        }
      }
    }
  ]
}
```

### Real-time Monitoring
```json
{
  "type": "page",
  "title": "System Monitoring",
  "body": [
    {
      "type": "grid",
      "columns": [
        {
          "md": 6,
          "body": [
            {
              "type": "chart",
              "name": "cpu_chart",
              "api": "/api/monitoring/cpu",
              "interval": 5000,
              "height": 300,
              "config": {
                "type": "line",
                "xField": "timestamp",
                "yField": "usage",
                "smooth": true,
                "color": "#ff6b6b",
                "yAxis": {
                  "min": 0,
                  "max": 100,
                  "label": {
                    "formatter": "{value}%"
                  }
                }
              },
              "replaceChartOption": true
            }
          ]
        },
        {
          "md": 6,
          "body": [
            {
              "type": "chart",
              "name": "memory_chart",
              "api": "/api/monitoring/memory",
              "interval": 5000,
              "height": 300,
              "config": {
                "type": "area",
                "xField": "timestamp",
                "yField": "usage",
                "color": "#4ecdc4",
                "yAxis": {
                  "min": 0,
                  "max": 100,
                  "label": {
                    "formatter": "{value}%"
                  }
                }
              }
            }
          ]
        }
      ]
    }
  ]
}
```

### Interactive Analytics
```json
{
  "type": "page",
  "title": "Product Analytics",
  "body": [
    {
      "type": "chart",
      "api": "/api/products/performance",
      "height": 600,
      "config": {
        "type": "scatter",
        "xField": "price",
        "yField": "sales_volume",
        "colorField": "category",
        "sizeField": "profit_margin",
        "size": [4, 30],
        "shape": "circle",
        "pointStyle": {
          "fillOpacity": 0.8,
          "stroke": "#bbb",
          "lineWidth": 1
        }
      },
      "clickAction": {
        "actionType": "drawer",
        "drawer": {
          "title": "Product Details: ${name}",
          "body": {
            "type": "form",
            "api": "/api/products/${product_id}",
            "body": [
              {
                "type": "static",
                "name": "name",
                "label": "Product Name"
              },
              {
                "type": "static",
                "name": "category",
                "label": "Category"
              },
              {
                "type": "static",
                "name": "price",
                "label": "Price",
                "tpl": "$${price}"
              },
              {
                "type": "chart",
                "api": "/api/products/${product_id}/trend",
                "height": 200,
                "config": {
                  "type": "line",
                  "xField": "month",
                  "yField": "sales"
                }
              }
            ]
          }
        }
      }
    }
  ]
}
```

### Geographic Data Visualization
```json
{
  "type": "chart",
  "api": "/api/sales/by-location",
  "mapURL": "/static/maps/world.json",
  "mapName": "world",
  "height": 500,
  "config": {
    "type": "map",
    "map": "world",
    "nameField": "country",
    "valueField": "sales",
    "roam": true,
    "zoom": 1.2,
    "itemStyle": {
      "areaColor": "#e6f3ff",
      "borderColor": "#999"
    },
    "emphasis": {
      "itemStyle": {
        "areaColor": "#4dabf7"
      }
    },
    "visualMap": {
      "min": 0,
      "max": 1000000,
      "left": "left",
      "top": "bottom",
      "text": ["High", "Low"],
      "calculable": true,
      "inRange": {
        "color": ["#e6f3ff", "#1c7ed6"]
      }
    }
  }
}
```

## Property Reference

### Core Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `type` | `string` | - | **Required.** Must be `"chart"` |
| `config` | `object` | - | **Required.** ECharts configuration object |
| `api` | `string/object` | - | Data source API endpoint |
| `source` | `string` | - | Data source from context variables |

### Data Management

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `initFetch` | `boolean` | `true` | Load data on component initialization |
| `initFetchOn` | `string` | - | Conditional expression for initial load |
| `interval` | `number` | - | Auto-refresh interval in milliseconds |
| `dataFilter` | `function` | - | Data transformation function |
| `disableDataMapping` | `boolean` | `false` | Disable automatic data mapping |

### Chart Configuration

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `chartTheme` | `object` | - | Chart theme configuration |
| `width` | `number/string` | - | Chart width (pixels or percentage) |
| `height` | `number/string` | `400` | Chart height (pixels or percentage) |
| `trackExpression` | `string` | - | Expression to track for updates |
| `replaceChartOption` | `boolean` | `false` | Replace vs. merge chart options |

### Interactive Features

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `clickAction` | `object` | - | Action configuration for chart clicks |
| `onEvent` | `object` | - | Event handlers for chart interactions |

### Geographic Charts

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `mapURL` | `string/object` | - | GeoJSON map data source |
| `mapName` | `string` | - | Map identifier for registration |
| `loadBaiduMap` | `boolean` | `false` | Enable Baidu map integration |

### Performance & Behavior

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `unMountOnHidden` | `boolean` | `false` | Destroy chart when hidden |
| `name` | `string` | - | Component identifier for targeting |

### Common Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `className` | `string` | - | Additional CSS classes |
| `style` | `object` | - | Inline styles |
| `id` | `string` | - | Unique component identifier |
| `testid` | `string` | - | Test automation identifier |
| `disabled` | `boolean` | `false` | Disable chart interactions |
| `hidden` | `boolean` | `false` | Hide the component |
| `visible` | `boolean` | `true` | Component visibility |

## Chart Configuration Reference

### Chart Types

| Type | Description | Key Fields |
|------|-------------|------------|
| `line` | Line chart | `xField`, `yField`, `seriesField` |
| `bar` | Bar chart | `xField`, `yField`, `seriesField` |
| `area` | Area chart | `xField`, `yField`, `seriesField` |
| `pie` | Pie chart | `angleField`, `colorField` |
| `scatter` | Scatter plot | `xField`, `yField`, `sizeField` |
| `map` | Geographic map | `nameField`, `valueField` |
| `gauge` | Gauge chart | `percent`, `range` |
| `radar` | Radar chart | `xField`, `yField`, `seriesField` |

### Common Configuration Options

```json
{
  "config": {
    "type": "line",
    "xField": "date",
    "yField": "value",
    "seriesField": "category",
    "smooth": true,
    "point": {
      "size": 4,
      "shape": "circle"
    },
    "color": ["#1890ff", "#52c41a", "#faad14"],
    "legend": {
      "position": "top"
    },
    "tooltip": {
      "showMarkers": true
    },
    "annotations": [
      {
        "type": "line",
        "start": ["min", "median"],
        "end": ["max", "median"],
        "style": {
          "stroke": "#ff4d4f",
          "lineDash": [2, 2]
        }
      }
    ]
  }
}
```

## Event Handling

### Available Events

| Event | Description | Data |
|-------|-------------|------|
| `click` | Chart element clicked | `{data: object, event: Event}` |
| `dblclick` | Chart element double-clicked | `{data: object, event: Event}` |
| `mouseover` | Mouse enters chart element | `{data: object, event: Event}` |
| `mouseout` | Mouse leaves chart element | `{data: object, event: Event}` |
| `legendclick` | Legend item clicked | `{data: object, event: Event}` |

### Event Configuration Examples

```json
{
  "type": "chart",
  "config": {...},
  "onEvent": {
    "click": {
      "actions": [
        {
          "actionType": "toast",
          "msg": "Clicked: ${data.name} - ${data.value}"
        }
      ]
    },
    "dblclick": {
      "actions": [
        {
          "actionType": "dialog",
          "dialog": {
            "title": "Details for ${data.name}",
            "body": "Value: ${data.value}"
          }
        }
      ]
    }
  }
}
```

## Theming & Styling

### Chart Themes

```json
{
  "type": "chart",
  "chartTheme": {
    "color": ["#5B8FF9", "#5AD8A6", "#5D7092", "#F6BD16"],
    "backgroundColor": "transparent",
    "textStyle": {
      "fontFamily": "Inter, sans-serif",
      "fontSize": 12,
      "color": "#6b7280"
    }
  },
  "config": {...}
}
```

### Custom Styling

```json
{
  "type": "chart",
  "className": "custom-chart",
  "style": {
    "border": "1px solid #e5e7eb",
    "borderRadius": "8px",
    "padding": "16px",
    "backgroundColor": "#ffffff"
  },
  "config": {...}
}
```

## Data Integration Patterns

### API Data Source
```json
{
  "type": "chart",
  "api": {
    "url": "/api/analytics/data",
    "method": "GET",
    "data": {
      "start_date": "${start_date}",
      "end_date": "${end_date}",
      "filters": "${filters}"
    }
  },
  "config": {...}
}
```

### Context Data Source
```json
{
  "type": "chart",
  "source": "${dashboardData.charts.sales}",
  "config": {...}
}
```

### Data Transformation
```json
{
  "type": "chart",
  "api": "/api/raw-data",
  "dataFilter": "return data.map(item => ({...item, value: item.amount / 1000}))",
  "config": {...}
}
```

## Accessibility

### ARIA Support
- Chart container has appropriate role
- Data tables for screen readers
- Keyboard navigation support
- Alternative text descriptions

### Screen Reader Features
```json
{
  "type": "chart",
  "config": {
    "accessibility": {
      "enabled": true,
      "description": "Sales performance chart showing monthly trends",
      "keyboardNavigation": {
        "enabled": true
      }
    }
  }
}
```

## Best Practices

### Performance
- Use appropriate chart types for data size
- Implement data sampling for large datasets
- Configure reasonable refresh intervals
- Optimize data queries on the server

### User Experience
- Provide loading states during data fetch
- Include meaningful tooltips and legends
- Support responsive design for mobile
- Implement proper error handling

### Accessibility
- Include chart descriptions for screen readers
- Provide alternative data representations
- Support keyboard navigation
- Use colorblind-friendly color schemes

## Go Type Definition

```go
// ChartComponent represents a data visualization chart
type ChartComponent struct {
    BaseComponent
    
    // Core Properties
    Config     map[string]interface{} `json:"config"`
    API        *APIConfig            `json:"api,omitempty"`
    Source     string                `json:"source,omitempty"`
    
    // Data Management
    InitFetch           bool   `json:"initFetch,omitempty"`
    InitFetchOn         string `json:"initFetchOn,omitempty"`
    Interval            int    `json:"interval,omitempty"`
    DataFilter          string `json:"dataFilter,omitempty"`
    DisableDataMapping  bool   `json:"disableDataMapping,omitempty"`
    
    // Chart Configuration  
    ChartTheme          map[string]interface{} `json:"chartTheme,omitempty"`
    Width               interface{}           `json:"width,omitempty"`
    Height              interface{}           `json:"height,omitempty"`
    TrackExpression     string                `json:"trackExpression,omitempty"`
    ReplaceChartOption  bool                  `json:"replaceChartOption,omitempty"`
    
    // Interactive Features
    ClickAction *ActionConfig                `json:"clickAction,omitempty"`
    OnEvent     map[string]EventConfig       `json:"onEvent,omitempty"`
    
    // Geographic Charts
    MapURL      *APIConfig `json:"mapURL,omitempty"`
    MapName     string     `json:"mapName,omitempty"`
    LoadBaiduMap bool      `json:"loadBaiduMap,omitempty"`
    
    // Performance & Behavior
    UnMountOnHidden bool   `json:"unMountOnHidden,omitempty"`
    Name            string `json:"name,omitempty"`
    
    // Style Properties
    ClassName string                 `json:"className,omitempty"`
    Style     map[string]interface{} `json:"style,omitempty"`
}

// ChartFactory creates Chart components from JSON configuration
func ChartFactory(config map[string]interface{}) (*ChartComponent, error) {
    component := &ChartComponent{
        BaseComponent: BaseComponent{
            Type: "chart",
        },
        InitFetch: true,
        Height:    400,
    }
    
    return component, mapConfig(config, component)
}

// Render generates the Templ template for the chart
func (c *ChartComponent) Render() templ.Component {
    return chart.Chart(chart.ChartProps{
        Config:             c.Config,
        API:                c.API,
        Source:             c.Source,
        InitFetch:          c.InitFetch,
        InitFetchOn:        c.InitFetchOn,
        Interval:           c.Interval,
        ChartTheme:         c.ChartTheme,
        Width:              c.Width,
        Height:             c.Height,
        TrackExpression:    c.TrackExpression,
        ReplaceChartOption: c.ReplaceChartOption,
        ClickAction:        c.ClickAction,
        OnEvent:            c.OnEvent,
        MapURL:             c.MapURL,
        MapName:            c.MapName,
        LoadBaiduMap:       c.LoadBaiduMap,
        UnMountOnHidden:    c.UnMountOnHidden,
        Name:               c.Name,
        ClassName:          c.ClassName,
        Style:              c.Style,
    })
}

// ECharts configuration types
type ChartConfig struct {
    Type        string                 `json:"type"`
    XField      string                 `json:"xField,omitempty"`
    YField      string                 `json:"yField,omitempty"`
    SeriesField string                 `json:"seriesField,omitempty"`
    Color       []string               `json:"color,omitempty"`
    Smooth      bool                   `json:"smooth,omitempty"`
    Legend      map[string]interface{} `json:"legend,omitempty"`
    Tooltip     map[string]interface{} `json:"tooltip,omitempty"`
    XAxis       map[string]interface{} `json:"xAxis,omitempty"`
    YAxis       map[string]interface{} `json:"yAxis,omitempty"`
}
```

## Related Components

- **[Table](../../organisms/table/)** - Data table component
- **[Card](../card/)** - Container for chart displays
- **[Stats](../stats/)** - Statistical summary component
- **[Progress](../atoms/progress/)** - Progress indicator component
- **[Form](../../organisms/form/)** - Form container with charts