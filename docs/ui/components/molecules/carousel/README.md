# Carousel Component

## Overview

The Carousel component provides a slideshow interface for displaying multiple items in a rotating fashion. It supports automatic playback, various animation types, navigation controls, and multi-item display modes.

## Basic Usage

### Simple Image Carousel
```json
{
  "type": "carousel",
  "options": [
    {"image": "/images/slide1.jpg", "title": "Slide 1"},
    {"image": "/images/slide2.jpg", "title": "Slide 2"},
    {"image": "/images/slide3.jpg", "title": "Slide 3"}
  ]
}
```

### Auto-playing Carousel
```json
{
  "type": "carousel",
  "auto": true,
  "interval": 3000,
  "animation": "fade",
  "options": [
    {"image": "/images/banner1.jpg", "title": "Welcome"},
    {"image": "/images/banner2.jpg", "title": "Features"},
    {"image": "/images/banner3.jpg", "title": "Get Started"}
  ]
}
```

### Carousel with Controls
```json
{
  "type": "carousel",
  "controls": ["dots", "arrows"],
  "controlsTheme": "dark",
  "alwaysShowArrow": true,
  "options": [
    {"image": "/images/product1.jpg", "caption": "Product A"},
    {"image": "/images/product2.jpg", "caption": "Product B"},
    {"image": "/images/product3.jpg", "caption": "Product C"}
  ]
}
```

## Complete Form Examples

### Product Showcase Carousel
```json
{
  "type": "page",
  "title": "Product Showcase",
  "body": [
    {
      "type": "carousel",
      "height": 400,
      "auto": true,
      "interval": 5000,
      "animation": "slide",
      "controls": ["dots", "arrows"],
      "controlsTheme": "light",
      "itemSchema": {
        "type": "container",
        "style": {
          "position": "relative",
          "height": "400px",
          "backgroundImage": "url(${image})",
          "backgroundSize": "cover",
          "backgroundPosition": "center"
        },
        "body": [
          {
            "type": "container",
            "className": "absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/60 to-transparent p-8",
            "body": [
              {
                "type": "tpl",
                "tpl": "<h2 class='text-3xl font-bold text-white mb-2'>${title}</h2>"
              },
              {
                "type": "tpl",
                "tpl": "<p class='text-lg text-gray-200'>${description}</p>"
              },
              {
                "type": "button",
                "label": "Learn More",
                "level": "primary",
                "className": "mt-4",
                "actionType": "link",
                "link": "${link}"
              }
            ]
          }
        ]
      },
      "options": [
        {
          "image": "/images/products/laptop.jpg",
          "title": "Premium Laptops",
          "description": "High-performance computing for professionals",
          "link": "/products/laptops"
        },
        {
          "image": "/images/products/tablet.jpg",
          "title": "Smart Tablets",
          "description": "Portable productivity and entertainment",
          "link": "/products/tablets"
        },
        {
          "image": "/images/products/phone.jpg",
          "title": "Latest Smartphones",
          "description": "Stay connected with cutting-edge technology",
          "link": "/products/phones"
        }
      ]
    }
  ]
}
```

### Team Member Carousel
```json
{
  "type": "carousel",
  "multiple": {"count": 3},
  "controls": ["arrows"],
  "alwaysShowArrow": true,
  "itemSchema": {
    "type": "card",
    "className": "mx-2",
    "body": [
      {
        "type": "image",
        "src": "${photo}",
        "className": "w-24 h-24 rounded-full mx-auto mb-4"
      },
      {
        "type": "tpl",
        "tpl": "<h3 class='text-lg font-semibold text-center'>${name}</h3>"
      },
      {
        "type": "tpl",
        "tpl": "<p class='text-sm text-gray-600 text-center'>${position}</p>"
      },
      {
        "type": "tpl",
        "tpl": "<p class='text-xs text-gray-500 text-center mt-2'>${department}</p>"
      }
    ]
  },
  "api": "/api/team-members"
}
```

### News & Updates Carousel
```json
{
  "type": "page",
  "title": "Latest News",
  "body": [
    {
      "type": "carousel",
      "height": 300,
      "auto": true,
      "interval": 4000,
      "animation": "slide",
      "controls": ["dots"],
      "controlsTheme": "dark",
      "thumbMode": "cover",
      "itemSchema": {
        "type": "container",
        "className": "relative h-full",
        "body": [
          {
            "type": "image",
            "src": "${featured_image}",
            "className": "w-full h-full object-cover"
          },
          {
            "type": "container",
            "className": "absolute bottom-0 left-0 right-0 bg-black/75 text-white p-6",
            "body": [
              {
                "type": "tpl",
                "tpl": "<span class='inline-block bg-blue-600 text-xs px-2 py-1 rounded mb-2'>${category}</span>"
              },
              {
                "type": "tpl",
                "tpl": "<h3 class='text-xl font-bold mb-2'>${title}</h3>"
              },
              {
                "type": "tpl",
                "tpl": "<p class='text-sm text-gray-300'>${excerpt}</p>"
              },
              {
                "type": "tpl",
                "tpl": "<div class='flex justify-between items-center mt-4'><span class='text-xs text-gray-400'>${published_date}</span><a href='${url}' class='text-blue-400 text-sm hover:underline'>Read More →</a></div>"
              }
            ]
          }
        ]
      },
      "api": "/api/news/featured"
    }
  ]
}
```

### Client Testimonials Carousel
```json
{
  "type": "carousel",
  "auto": true,
  "interval": 6000,
  "animation": "fade",
  "controls": ["dots"],
  "height": 250,
  "itemSchema": {
    "type": "container",
    "className": "text-center p-8",
    "body": [
      {
        "type": "tpl",
        "tpl": "<div class='text-4xl text-gray-300 mb-4'>\"</div>"
      },
      {
        "type": "tpl",
        "tpl": "<p class='text-lg italic text-gray-700 mb-6'>${testimonial}</p>"
      },
      {
        "type": "container",
        "className": "flex items-center justify-center space-x-4",
        "body": [
          {
            "type": "image",
            "src": "${client_photo}",
            "className": "w-12 h-12 rounded-full"
          },
          {
            "type": "container",
            "body": [
              {
                "type": "tpl",
                "tpl": "<p class='font-semibold text-gray-900'>${client_name}</p>"
              },
              {
                "type": "tpl",
                "tpl": "<p class='text-sm text-gray-600'>${client_title}, ${company}</p>"
              }
            ]
          }
        ]
      }
    ]
  },
  "options": [
    {
      "testimonial": "This product has transformed our workflow completely. Highly recommended!",
      "client_name": "Sarah Johnson",
      "client_title": "CTO",
      "company": "TechCorp Inc.",
      "client_photo": "/images/clients/sarah.jpg"
    },
    {
      "testimonial": "Outstanding customer service and excellent product quality.",
      "client_name": "Michael Chen",
      "client_title": "Product Manager",
      "company": "Innovation Labs",
      "client_photo": "/images/clients/michael.jpg"
    }
  ]
}
```

### Feature Highlights Carousel
```json
{
  "type": "carousel",
  "controls": ["arrows"],
  "alwaysShowArrow": false,
  "animation": "slide",
  "duration": 300,
  "icons": {
    "prev": {
      "type": "icon",
      "icon": "fa fa-chevron-left",
      "className": "text-2xl"
    },
    "next": {
      "type": "icon", 
      "icon": "fa fa-chevron-right",
      "className": "text-2xl"
    }
  },
  "itemSchema": {
    "type": "container",
    "className": "p-8 text-center",
    "body": [
      {
        "type": "icon",
        "icon": "${icon}",
        "className": "text-6xl text-blue-600 mb-4"
      },
      {
        "type": "tpl",
        "tpl": "<h3 class='text-2xl font-bold mb-4'>${title}</h3>"
      },
      {
        "type": "tpl",
        "tpl": "<p class='text-gray-600 leading-relaxed'>${description}</p>"
      },
      {
        "type": "button",
        "label": "Learn More",
        "level": "link",
        "className": "mt-4",
        "actionType": "dialog",
        "dialog": {
          "title": "${title}",
          "body": {
            "type": "tpl",
            "tpl": "<div class='prose'>${detailed_description}</div>"
          }
        }
      }
    ]
  },
  "options": [
    {
      "icon": "fa fa-shield-alt",
      "title": "Security First",
      "description": "Enterprise-grade security with end-to-end encryption",
      "detailed_description": "<p>Our security measures include...</p>"
    },
    {
      "icon": "fa fa-chart-line",
      "title": "Advanced Analytics",
      "description": "Real-time insights and comprehensive reporting",
      "detailed_description": "<p>Our analytics platform provides...</p>"
    }
  ]
}
```

## Property Reference

### Core Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `type` | `string` | - | **Required.** Must be `"carousel"` |
| `options` | `array` | `[]` | Array of slide data objects |
| `itemSchema` | `object` | - | Template for rendering each slide |

### Playback Control

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `auto` | `boolean` | `false` | Enable automatic slideshow |
| `interval` | `number/string` | `3000` | Auto-play interval in milliseconds |
| `duration` | `number` | `500` | Animation transition duration |

### Animation & Display

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `animation` | `string` | `"slide"` | Animation type: `"slide"` or `"fade"` |
| `width` | `number` | - | Carousel width in pixels |
| `height` | `number` | - | Carousel height in pixels |
| `thumbMode` | `string` | `"contain"` | Image display mode: `"contain"` or `"cover"` |

### Navigation Controls

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `controls` | `array` | `["dots"]` | Control types: `["dots", "arrows"]` |
| `controlsTheme` | `string` | `"light"` | Theme: `"light"` or `"dark"` |
| `alwaysShowArrow` | `boolean` | `false` | Always show navigation arrows |

### Multi-Item Display

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `multiple` | `object` | - | Multi-item configuration |
| `multiple.count` | `number` | - | Number of items to show simultaneously |

### Custom Icons

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `icons` | `object` | - | Custom navigation icons |
| `icons.prev` | `object` | - | Previous button icon schema |
| `icons.next` | `object` | - | Next button icon schema |

### Data Source

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `name` | `string` | - | Data source field name |
| `api` | `string/object` | - | API endpoint for dynamic data |

### Common Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `className` | `string` | - | Additional CSS classes |
| `style` | `object` | - | Inline styles |
| `id` | `string` | - | Unique component identifier |
| `testid` | `string` | - | Test automation identifier |
| `disabled` | `boolean` | `false` | Disable carousel interactions |
| `hidden` | `boolean` | `false` | Hide the component |
| `visible` | `boolean` | `true` | Component visibility |
| `placeholder` | `string` | - | Placeholder text when no data |

## Event Handling

### Available Events

| Event | Description | Data |
|-------|-------------|------|
| `change` | Slide changed | `{index: number, item: object}` |
| `click` | Slide clicked | `{index: number, item: object}` |
| `mouseenter` | Mouse enters carousel | `{index: number}` |
| `mouseleave` | Mouse leaves carousel | `{index: number}` |

### Event Configuration Examples

```json
{
  "type": "carousel",
  "options": [...],
  "onEvent": {
    "change": {
      "actions": [
        {
          "actionType": "setValue",
          "componentId": "current-slide",
          "value": "${index}"
        }
      ]
    },
    "click": {
      "actions": [
        {
          "actionType": "dialog",
          "dialog": {
            "title": "Slide Details",
            "body": "Clicked on slide ${index}: ${item.title}"
          }
        }
      ]
    }
  }
}
```

## Item Schema Patterns

### Basic Item Template
```json
{
  "itemSchema": {
    "type": "container",
    "body": [
      {"type": "image", "src": "${image}"},
      {"type": "tpl", "tpl": "<h3>${title}</h3>"},
      {"type": "tpl", "tpl": "<p>${description}</p>"}
    ]
  }
}
```

### Card-based Template
```json
{
  "itemSchema": {
    "type": "card",
    "header": {
      "title": "${title}",
      "subTitle": "${subtitle}"
    },
    "body": [
      {"type": "image", "src": "${image}"},
      {"type": "tpl", "tpl": "${content}"}
    ],
    "actions": [
      {
        "type": "button",
        "label": "View Details",
        "actionType": "link",
        "link": "${detailUrl}"
      }
    ]
  }
}
```

### Custom Layout Template
```json
{
  "itemSchema": {
    "type": "grid",
    "columns": [
      {
        "md": 6,
        "body": [
          {"type": "image", "src": "${image}", "className": "w-full h-full object-cover"}
        ]
      },
      {
        "md": 6,
        "body": [
          {"type": "tpl", "tpl": "<h2 class='text-2xl font-bold'>${title}</h2>"},
          {"type": "tpl", "tpl": "<p class='text-gray-600 mt-4'>${description}</p>"},
          {
            "type": "button",
            "label": "${buttonText}",
            "level": "primary",
            "className": "mt-6",
            "actionType": "link",
            "link": "${actionUrl}"
          }
        ]
      }
    ]
  }
}
```

## Styling & Theming

### CSS Classes

- `.carousel` - Base carousel container
- `.carousel__viewport` - Slide viewport area
- `.carousel__track` - Slide track container
- `.carousel__slide` - Individual slide wrapper
- `.carousel__controls` - Navigation controls container
- `.carousel__dots` - Dot indicators
- `.carousel__arrows` - Arrow navigation
- `.carousel--auto` - Auto-playing state
- `.carousel--multiple` - Multi-item display

### Custom Styling Examples

```json
{
  "type": "carousel",
  "className": "custom-carousel",
  "style": {
    "borderRadius": "12px",
    "overflow": "hidden",
    "boxShadow": "0 10px 25px rgba(0,0,0,0.1)"
  },
  "controlsTheme": "dark",
  "options": [...]
}
```

### Responsive Configuration
```json
{
  "type": "carousel",
  "className": "responsive-carousel",
  "height": 300,
  "multiple": {
    "count": "${device === 'mobile' ? 1 : device === 'tablet' ? 2 : 3}"
  },
  "options": [...]
}
```

## Accessibility

### ARIA Support
- `role="region"` for carousel container
- `aria-label` for navigation controls
- `aria-current` for active slide
- Keyboard navigation support

### Keyboard Navigation
- `Arrow Left/Right` - Navigate slides
- `Space/Enter` - Activate controls
- `Tab` - Focus navigation elements

### Best Practices
- Provide pause controls for auto-playing carousels
- Include alternative navigation methods
- Ensure sufficient color contrast for controls
- Support reduced motion preferences

## Integration Patterns

### With API Data Sources
```json
{
  "type": "carousel",
  "api": "/api/featured-content",
  "itemSchema": {
    "type": "card",
    "body": [
      {"type": "image", "src": "${thumbnail}"},
      {"type": "tpl", "tpl": "<h3>${title}</h3>"}
    ]
  }
}
```

### With Form Controls
```json
{
  "type": "form",
  "body": [
    {
      "type": "select",
      "name": "category",
      "label": "Category",
      "options": [
        {"label": "All", "value": "all"},
        {"label": "Featured", "value": "featured"}
      ]
    },
    {
      "type": "carousel",
      "api": "/api/content?category=${category}",
      "trackExpression": "${category}",
      "itemSchema": {...}
    }
  ]
}
```

## Best Practices

### Performance
- Optimize images for web delivery
- Lazy load non-visible slides
- Limit the number of simultaneous slides
- Use appropriate image compression

### User Experience
- Provide clear navigation indicators
- Include pause/play controls for auto-play
- Ensure smooth transitions
- Support touch/swipe gestures on mobile

### Accessibility
- Include proper ARIA labels
- Support keyboard navigation
- Provide alternative content access
- Respect user motion preferences

## Go Type Definition

```go
// CarouselComponent represents a slideshow carousel
type CarouselComponent struct {
    BaseComponent
    
    // Core Properties
    Options    []map[string]interface{} `json:"options,omitempty"`
    ItemSchema *SchemaNode             `json:"itemSchema,omitempty"`
    
    // Playback Control
    Auto     bool        `json:"auto,omitempty"`
    Interval interface{} `json:"interval,omitempty"`
    Duration int         `json:"duration,omitempty"`
    
    // Animation & Display
    Animation string `json:"animation,omitempty"`
    Width     int    `json:"width,omitempty"`
    Height    int    `json:"height,omitempty"`
    ThumbMode string `json:"thumbMode,omitempty"`
    
    // Navigation Controls
    Controls        []string `json:"controls,omitempty"`
    ControlsTheme   string   `json:"controlsTheme,omitempty"`
    AlwaysShowArrow bool     `json:"alwaysShowArrow,omitempty"`
    
    // Multi-Item Display
    Multiple *MultipleConfig `json:"multiple,omitempty"`
    
    // Custom Icons
    Icons *IconsConfig `json:"icons,omitempty"`
    
    // Data Source
    Name string     `json:"name,omitempty"`
    API  *APIConfig `json:"api,omitempty"`
    
    // Style Properties
    ClassName   string                 `json:"className,omitempty"`
    Style       map[string]interface{} `json:"style,omitempty"`
    Placeholder string                 `json:"placeholder,omitempty"`
    
    // Event Configuration
    OnEvent map[string]EventConfig `json:"onEvent,omitempty"`
}

// MultipleConfig for multi-item display
type MultipleConfig struct {
    Count int `json:"count"`
}

// IconsConfig for custom navigation icons
type IconsConfig struct {
    Prev *SchemaNode `json:"prev,omitempty"`
    Next *SchemaNode `json:"next,omitempty"`
}

// CarouselFactory creates Carousel components from JSON configuration
func CarouselFactory(config map[string]interface{}) (*CarouselComponent, error) {
    component := &CarouselComponent{
        BaseComponent: BaseComponent{
            Type: "carousel",
        },
        Animation:     "slide",
        Duration:      500,
        Controls:      []string{"dots"},
        ControlsTheme: "light",
        ThumbMode:     "contain",
    }
    
    return component, mapConfig(config, component)
}

// Render generates the Templ template for the carousel
func (c *CarouselComponent) Render() templ.Component {
    return carousel.Carousel(carousel.CarouselProps{
        Options:         c.Options,
        ItemSchema:      c.ItemSchema,
        Auto:            c.Auto,
        Interval:        c.Interval,
        Duration:        c.Duration,
        Animation:       c.Animation,
        Width:           c.Width,
        Height:          c.Height,
        ThumbMode:       c.ThumbMode,
        Controls:        c.Controls,
        ControlsTheme:   c.ControlsTheme,
        AlwaysShowArrow: c.AlwaysShowArrow,
        Multiple:        c.Multiple,
        Icons:           c.Icons,
        Name:            c.Name,
        API:             c.API,
        ClassName:       c.ClassName,
        Style:           c.Style,
        Placeholder:     c.Placeholder,
        OnEvent:         c.OnEvent,
    })
}
```

## Related Components

- **[Image](../atoms/image/)** - Image display component
- **[Card](../card/)** - Card container component
- **[Button](../atoms/button/)** - Navigation button component
- **[Container](../atoms/container/)** - Layout container component
- **[Grid](../molecules/grid/)** - Grid layout component