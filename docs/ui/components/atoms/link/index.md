# Link Component

**FILE PURPOSE**: Navigation element implementation and specifications  
**SCOPE**: All link variants, states, and accessibility patterns  
**TARGET AUDIENCE**: Developers implementing navigation, references, and interactive text elements

## 📋 Component Overview

The Link component provides accessible navigation elements for both internal routing and external references. It supports various styles, states, and behaviors while maintaining semantic HTML and accessibility standards.

### Schema Reference
- **Primary Schema**: `LinkSchema.json`
- **Related Schemas**: `ActionSchema.json`, `ButtonGroupSchema.json`
- **Base Interface**: Navigation element with semantic meaning

## 🎨 Link Types

### Basic Link
**Purpose**: Standard navigation and reference links

```go
// Basic link configuration
basicLink := LinkProps{
    Text:   "Learn more",
    Href:   "/documentation",
    Target: "_self",
}

// Generated Templ component
templ BasicLink(props LinkProps) {
    <a href={ props.Href }
       target={ props.Target }
       rel={ props.Rel }
       class={ fmt.Sprintf("link link-%s link-%s", props.Size, props.Variant) }
       role={ props.Role }
       aria-label={ props.AriaLabel }
       aria-describedby={ props.AriaDescribedBy }>
        
        if props.Icon != "" && props.IconPosition == "left" {
            <span class="link-icon link-icon-left">
                @Icon(IconProps{Name: props.Icon, Size: getLinkIconSize(props.Size)})
            </span>
        }
        
        <span class="link-text">{ props.Text }</span>
        
        if props.Icon != "" && props.IconPosition == "right" {
            <span class="link-icon link-icon-right">
                @Icon(IconProps{Name: props.Icon, Size: getLinkIconSize(props.Size)})
            </span>
        }
        
        if props.External {
            <span class="link-external-indicator" aria-label="Opens in new window">
                @Icon(IconProps{Name: "external-link", Size: "xs"})
            </span>
        }
    </a>
}
```

### Button Link
**Purpose**: Links styled as buttons for call-to-action elements

```go
buttonLink := LinkProps{
    Text:    "Get Started",
    Href:    "/signup",
    Variant: "button",
    Color:   "primary",
    Size:    "lg",
}

templ ButtonLink(props LinkProps) {
    <a href={ props.Href }
       target={ props.Target }
       class={ fmt.Sprintf("link-button btn btn-%s btn-%s", props.Color, props.Size) }
       role="button"
       aria-label={ props.AriaLabel }>
        
        if props.Icon != "" && props.IconPosition == "left" {
            <span class="btn-icon btn-icon-left">
                @Icon(IconProps{Name: props.Icon, Size: getLinkIconSize(props.Size)})
            </span>
        }
        
        <span class="btn-text">{ props.Text }</span>
        
        if props.Icon != "" && props.IconPosition == "right" {
            <span class="btn-icon btn-icon-right">
                @Icon(IconProps{Name: props.Icon, Size: getLinkIconSize(props.Size)})
            </span>
        }
        
        if props.Loading {
            <span class="btn-loading">
                @Icon(IconProps{Name: "loader", Size: getLinkIconSize(props.Size)})
            </span>
        }
    </a>
}
```

### Breadcrumb Link
**Purpose**: Navigation hierarchy links

```go
breadcrumbLink := LinkProps{
    Text:      "Products",
    Href:      "/products",
    Variant:   "breadcrumb",
    Separator: "/",
}

templ BreadcrumbLink(props LinkProps) {
    <span class="breadcrumb-item">
        <a href={ props.Href }
           class="breadcrumb-link"
           aria-label={ fmt.Sprintf("Navigate to %s", props.Text) }>
            { props.Text }
        </a>
        
        if props.Separator != "" && !props.IsLast {
            <span class="breadcrumb-separator" aria-hidden="true">
                { props.Separator }
            </span>
        }
    </span>
}

templ BreadcrumbNavigation(props BreadcrumbProps) {
    <nav class="breadcrumb-nav" aria-label="Breadcrumb navigation">
        <ol class="breadcrumb-list">
            for i, item := range props.Items {
                <li class="breadcrumb-item">
                    if i == len(props.Items)-1 {
                        <span class="breadcrumb-current" aria-current="page">
                            { item.Text }
                        </span>
                    } else {
                        @BreadcrumbLink(LinkProps{
                            Text:      item.Text,
                            Href:      item.Href,
                            Separator: props.Separator,
                            IsLast:    false,
                        })
                    }
                </li>
            }
        </ol>
    </nav>
}
```

### Card Link
**Purpose**: Large clickable areas for cards and tiles

```go
cardLink := LinkProps{
    Text:        "View Project",
    Href:        "/projects/123",
    Variant:     "card",
    Description: "Full stack web application with React and Node.js",
    Image:       "/images/project-thumb.jpg",
}

templ CardLink(props LinkProps) {
    <a href={ props.Href }
       class="card-link"
       aria-label={ fmt.Sprintf("View details for %s", props.Text) }>
        
        <article class="link-card">
            if props.Image != "" {
                <div class="card-image">
                    <img src={ props.Image } alt="" loading="lazy" />
                </div>
            }
            
            <div class="card-content">
                <h3 class="card-title">{ props.Text }</h3>
                
                if props.Description != "" {
                    <p class="card-description">{ props.Description }</p>
                }
                
                if props.Meta != "" {
                    <div class="card-meta">{ props.Meta }</div>
                }
                
                <div class="card-action">
                    <span class="action-text">{ props.ActionText }</span>
                    @Icon(IconProps{Name: "arrow-right", Size: "sm"})
                </div>
            </div>
        </article>
    </a>
}
```

### Tab Link
**Purpose**: Navigation tabs with active states

```go
tabLink := LinkProps{
    Text:   "Overview",
    Href:   "/dashboard/overview",
    Variant: "tab",
    Active:  true,
}

templ TabLink(props LinkProps) {
    <a href={ props.Href }
       class={ fmt.Sprintf("tab-link %s", ternary(props.Active, "tab-active", "")) }
       role="tab"
       aria-selected={ fmt.Sprintf("%t", props.Active) }
       aria-controls={ props.AriaControls }
       tabindex={ ternary(props.Active, 0, -1) }>
        
        if props.Icon != "" {
            <span class="tab-icon">
                @Icon(IconProps{Name: props.Icon, Size: "sm"})
            </span>
        }
        
        <span class="tab-text">{ props.Text }</span>
        
        if props.Count > 0 {
            <span class="tab-count">{ formatCount(props.Count) }</span>
        }
        
        if props.Badge != "" {
            @Badge(BadgeProps{Text: props.Badge, Size: "xs", Variant: "danger"})
        }
    </a>
}

templ TabNavigation(props TabNavigationProps) {
    <nav class="tab-navigation" role="tablist" aria-label={ props.AriaLabel }>
        for _, tab := range props.Tabs {
            @TabLink(LinkProps{
                Text:         tab.Text,
                Href:         tab.Href,
                Icon:         tab.Icon,
                Count:        tab.Count,
                Badge:        tab.Badge,
                Active:       tab.Active,
                AriaControls: tab.PanelID,
            })
        }
    </nav>
}
```

## 🎯 Props Interface

### Core Properties
```go
type LinkProps struct {
    // Identity
    ID     string `json:"id"`
    TestID string `json:"testid"`
    
    // Content
    Text        string `json:"text"`        // Link text
    Href        string `json:"href"`        // URL or route
    Description string `json:"description"` // Additional description
    
    // Navigation
    Target  string `json:"target"`  // _self, _blank, _parent, _top
    Rel     string `json:"rel"`     // noopener, noreferrer, etc.
    Replace bool   `json:"replace"` // Replace history entry
    
    // Appearance
    Variant  LinkVariant `json:"variant"`  // text, button, card, tab, breadcrumb
    Size     LinkSize    `json:"size"`     // xs, sm, md, lg, xl
    Color    LinkColor   `json:"color"`    // primary, secondary, success, etc.
    
    // Icons and Media
    Icon         string       `json:"icon"`         // Icon name
    IconPosition IconPosition `json:"iconPosition"` // left, right
    Image        string       `json:"image"`        // Image URL for cards
    
    // States
    Active    bool `json:"active"`    // Active state for tabs/nav
    Disabled  bool `json:"disabled"`  // Disabled state
    Loading   bool `json:"loading"`   // Loading state
    External  bool `json:"external"`  // External link indicator
    Download  bool `json:"download"`  // Download link
    
    // Additional Content
    Count      int    `json:"count"`      // Count for tabs/nav
    Badge      string `json:"badge"`      // Badge text
    Meta       string `json:"meta"`       // Metadata text
    ActionText string `json:"actionText"` // Call-to-action text
    
    // Navigation specific
    Separator    string `json:"separator"`    // Breadcrumb separator
    IsLast       bool   `json:"isLast"`       // Last breadcrumb item
    AriaControls string `json:"ariaControls"` // Tab panel ID
    
    // Styling
    Class     string            `json:"className"`
    Style     map[string]string `json:"style"`
    UnderLine bool              `json:"underline"` // Force underline
    
    // Events
    OnClick string `json:"onClick"`
    OnHover string `json:"onHover"`
    OnFocus string `json:"onFocus"`
    
    // Accessibility
    AriaLabel       string `json:"ariaLabel"`
    AriaDescribedBy string `json:"ariaDescribedBy"`
    Role            string `json:"role"`
    Title           string `json:"title"`
}
```

### Navigation Properties
```go
type BreadcrumbProps struct {
    Items     []BreadcrumbItem `json:"items"`
    Separator string           `json:"separator"` // Default "/"
    MaxItems  int              `json:"maxItems"`  // Collapse after N items
}

type BreadcrumbItem struct {
    Text string `json:"text"`
    Href string `json:"href"`
}

type TabNavigationProps struct {
    Tabs      []TabItem `json:"tabs"`
    AriaLabel string    `json:"ariaLabel"`
    Vertical  bool      `json:"vertical"`
}

type TabItem struct {
    Text     string `json:"text"`
    Href     string `json:"href"`
    Icon     string `json:"icon"`
    Count    int    `json:"count"`
    Badge    string `json:"badge"`
    Active   bool   `json:"active"`
    PanelID  string `json:"panelId"`
    Disabled bool   `json:"disabled"`
}
```

### Size Variants
```go
type LinkSize string

const (
    LinkXS LinkSize = "xs"    // 12px text, small icons
    LinkSM LinkSize = "sm"    // 14px text, small icons
    LinkMD LinkSize = "md"    // 16px text, medium icons (default)
    LinkLG LinkSize = "lg"    // 18px text, large icons
    LinkXL LinkSize = "xl"    // 20px text, extra large icons
)
```

### Visual Variants
```go
type LinkVariant string

const (
    LinkText       LinkVariant = "text"       // Standard text link
    LinkButton     LinkVariant = "button"     // Button appearance
    LinkCard       LinkVariant = "card"       // Card/tile link
    LinkTab        LinkVariant = "tab"        // Tab navigation
    LinkBreadcrumb LinkVariant = "breadcrumb" // Breadcrumb navigation
    LinkNav        LinkVariant = "nav"        // Navigation link
)
```

### Color Options
```go
type LinkColor string

const (
    LinkPrimary   LinkColor = "primary"   // Brand color
    LinkSecondary LinkColor = "secondary" // Neutral color
    LinkSuccess   LinkColor = "success"   // Green color
    LinkWarning   LinkColor = "warning"   // Yellow color
    LinkDanger    LinkColor = "danger"    // Red color
    LinkInfo      LinkColor = "info"      // Blue color
    LinkMuted     LinkColor = "muted"     // Subdued color
)
```

## 🎨 Styling Implementation

### Base Link Styles
```css
.link {
    display: inline-flex;
    align-items: center;
    gap: var(--space-xs);
    color: var(--color-primary);
    text-decoration: none;
    font-weight: var(--font-weight-medium);
    transition: var(--transition-base);
    cursor: pointer;
    
    &:hover {
        color: var(--color-primary-dark);
        text-decoration: underline;
    }
    
    &:focus {
        outline: none;
        box-shadow: 0 0 0 2px var(--color-primary-light);
        border-radius: var(--radius-sm);
    }
    
    &:active {
        color: var(--color-primary-darker);
    }
    
    &:visited {
        color: var(--color-primary-visited);
    }
    
    &[aria-disabled="true"] {
        color: var(--color-text-disabled);
        cursor: not-allowed;
        text-decoration: none;
        
        &:hover {
            color: var(--color-text-disabled);
            text-decoration: none;
        }
    }
    
    .link-text {
        line-height: 1.4;
    }
    
    .link-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
        
        &.link-icon-left {
            margin-right: var(--space-xs);
        }
        
        &.link-icon-right {
            margin-left: var(--space-xs);
        }
    }
    
    .link-external-indicator {
        opacity: 0.7;
        margin-left: var(--space-xs);
    }
}
```

### Size Variants
```css
/* Extra Small */
.link-xs {
    font-size: var(--font-size-xs);
    gap: 2px;
    
    .link-icon {
        width: 12px;
        height: 12px;
    }
}

/* Small */
.link-sm {
    font-size: var(--font-size-sm);
    gap: var(--space-xs);
    
    .link-icon {
        width: 14px;
        height: 14px;
    }
}

/* Medium (Default) */
.link-md {
    font-size: var(--font-size-base);
    gap: var(--space-xs);
    
    .link-icon {
        width: 16px;
        height: 16px;
    }
}

/* Large */
.link-lg {
    font-size: var(--font-size-lg);
    gap: var(--space-sm);
    
    .link-icon {
        width: 18px;
        height: 18px;
    }
}

/* Extra Large */
.link-xl {
    font-size: var(--font-size-xl);
    gap: var(--space-sm);
    
    .link-icon {
        width: 20px;
        height: 20px;
    }
}
```

### Color Variants
```css
/* Primary (Default) */
.link-primary {
    color: var(--color-primary);
    
    &:hover {
        color: var(--color-primary-dark);
    }
    
    &:visited {
        color: var(--color-primary-visited);
    }
}

/* Secondary */
.link-secondary {
    color: var(--color-text-secondary);
    
    &:hover {
        color: var(--color-text-primary);
    }
}

/* Success */
.link-success {
    color: var(--color-success);
    
    &:hover {
        color: var(--color-success-dark);
    }
}

/* Warning */
.link-warning {
    color: var(--color-warning);
    
    &:hover {
        color: var(--color-warning-dark);
    }
}

/* Danger */
.link-danger {
    color: var(--color-danger);
    
    &:hover {
        color: var(--color-danger-dark);
    }
}

/* Info */
.link-info {
    color: var(--color-info);
    
    &:hover {
        color: var(--color-info-dark);
    }
}

/* Muted */
.link-muted {
    color: var(--color-text-tertiary);
    
    &:hover {
        color: var(--color-text-secondary);
    }
}
```

### Button Link Styles
```css
.link-button {
    /* Inherits all button styles */
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-sm);
    padding: var(--space-sm) var(--space-md);
    background: var(--color-primary);
    color: white;
    border: 1px solid var(--color-primary);
    border-radius: var(--radius-md);
    font-weight: var(--font-weight-medium);
    text-decoration: none;
    transition: var(--transition-base);
    
    &:hover {
        background: var(--color-primary-dark);
        border-color: var(--color-primary-dark);
        color: white;
        text-decoration: none;
        transform: translateY(-1px);
        box-shadow: var(--shadow-md);
    }
    
    &:focus {
        outline: none;
        box-shadow: 0 0 0 3px var(--color-primary-light);
    }
    
    &:active {
        transform: translateY(0);
        box-shadow: var(--shadow-sm);
    }
    
    .btn-loading {
        animation: spin 1s linear infinite;
    }
}
```

### Card Link Styles
```css
.card-link {
    display: block;
    text-decoration: none;
    color: inherit;
    border-radius: var(--radius-lg);
    transition: var(--transition-base);
    
    &:hover {
        text-decoration: none;
        transform: translateY(-2px);
        box-shadow: var(--shadow-lg);
        
        .card-action {
            color: var(--color-primary);
        }
    }
    
    &:focus {
        outline: none;
        box-shadow: 0 0 0 3px var(--color-primary-light);
    }
    
    .link-card {
        border: 1px solid var(--color-border-light);
        border-radius: var(--radius-lg);
        background: var(--color-bg-surface);
        overflow: hidden;
        transition: var(--transition-base);
        
        .card-image {
            aspect-ratio: 16 / 9;
            overflow: hidden;
            
            img {
                width: 100%;
                height: 100%;
                object-fit: cover;
                transition: var(--transition-base);
            }
        }
        
        .card-content {
            padding: var(--space-lg);
            
            .card-title {
                font-size: var(--font-size-lg);
                font-weight: var(--font-weight-semibold);
                color: var(--color-text-primary);
                margin: 0 0 var(--space-sm) 0;
                line-height: 1.3;
            }
            
            .card-description {
                color: var(--color-text-secondary);
                font-size: var(--font-size-sm);
                line-height: 1.5;
                margin: 0 0 var(--space-md) 0;
            }
            
            .card-meta {
                color: var(--color-text-tertiary);
                font-size: var(--font-size-xs);
                margin: 0 0 var(--space-md) 0;
            }
            
            .card-action {
                display: flex;
                align-items: center;
                gap: var(--space-xs);
                color: var(--color-text-secondary);
                font-size: var(--font-size-sm);
                font-weight: var(--font-weight-medium);
                transition: var(--transition-base);
                
                .action-text {
                    flex: 1;
                }
            }
        }
    }
    
    &:hover .link-card {
        .card-image img {
            transform: scale(1.05);
        }
    }
}
```

### Tab Link Styles
```css
.tab-navigation {
    border-bottom: 1px solid var(--color-border-light);
    display: flex;
    gap: 0;
    
    &[aria-orientation="vertical"] {
        flex-direction: column;
        border-bottom: none;
        border-right: 1px solid var(--color-border-light);
    }
}

.tab-link {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    padding: var(--space-md) var(--space-lg);
    color: var(--color-text-secondary);
    text-decoration: none;
    border-bottom: 2px solid transparent;
    font-weight: var(--font-weight-medium);
    transition: var(--transition-base);
    white-space: nowrap;
    
    &:hover {
        color: var(--color-text-primary);
        text-decoration: none;
        background: var(--color-bg-hover);
    }
    
    &:focus {
        outline: none;
        box-shadow: inset 0 0 0 2px var(--color-primary-light);
    }
    
    &.tab-active {
        color: var(--color-primary);
        border-bottom-color: var(--color-primary);
        background: var(--color-primary-light);
    }
    
    &[aria-disabled="true"] {
        color: var(--color-text-disabled);
        cursor: not-allowed;
        
        &:hover {
            color: var(--color-text-disabled);
            background: none;
        }
    }
    
    .tab-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }
    
    .tab-text {
        flex: 1;
        line-height: 1.4;
    }
    
    .tab-count {
        background: var(--color-bg-secondary);
        color: var(--color-text-secondary);
        font-size: var(--font-size-xs);
        padding: 2px 6px;
        border-radius: var(--radius-sm);
        font-weight: var(--font-weight-semibold);
    }
    
    &.tab-active .tab-count {
        background: var(--color-primary);
        color: white;
    }
}

/* Vertical tabs */
.tab-navigation[aria-orientation="vertical"] {
    .tab-link {
        border-bottom: none;
        border-right: 2px solid transparent;
        
        &.tab-active {
            border-right-color: var(--color-primary);
        }
    }
}
```

### Breadcrumb Styles
```css
.breadcrumb-nav {
    .breadcrumb-list {
        display: flex;
        align-items: center;
        gap: var(--space-xs);
        list-style: none;
        margin: 0;
        padding: 0;
        flex-wrap: wrap;
    }
    
    .breadcrumb-item {
        display: flex;
        align-items: center;
        gap: var(--space-xs);
        
        .breadcrumb-link {
            color: var(--color-text-secondary);
            text-decoration: none;
            font-size: var(--font-size-sm);
            transition: var(--transition-base);
            
            &:hover {
                color: var(--color-primary);
                text-decoration: underline;
            }
            
            &:focus {
                outline: none;
                box-shadow: 0 0 0 2px var(--color-primary-light);
                border-radius: var(--radius-sm);
            }
        }
        
        .breadcrumb-current {
            color: var(--color-text-primary);
            font-size: var(--font-size-sm);
            font-weight: var(--font-weight-medium);
        }
        
        .breadcrumb-separator {
            color: var(--color-text-tertiary);
            font-size: var(--font-size-sm);
            user-select: none;
        }
    }
}
```

## ⚙️ Advanced Features

### Link Prefetching
```go
templ PrefetchLink(props LinkProps) {
    <a href={ props.Href }
       class={ props.GetClasses() }
       @mouseenter={ fmt.Sprintf("prefetchPage('%s')", props.Href) }
       x-data="{ prefetched: false }"
       @mouseenter="if (!prefetched) { 
           fetch($el.href, { method: 'HEAD' }); 
           prefetched = true; 
       }">
        
        <span class="link-text">{ props.Text }</span>
    </a>
}
```

### Link with Loading State
```go
templ LoadingLink(props LinkProps) {
    <a href={ props.Href }
       class={ props.GetClasses() }
       x-data="{ loading: false }"
       @click="loading = true"
       :class="{ 'link-loading': loading }"
       :aria-busy="loading">
        
        <span x-show="!loading" class="link-content">
            if props.Icon != "" {
                @Icon(IconProps{Name: props.Icon, Size: getLinkIconSize(props.Size)})
            }
            <span class="link-text">{ props.Text }</span>
        </span>
        
        <span x-show="loading" class="link-loading-content">
            @Icon(IconProps{Name: "loader", Size: getLinkIconSize(props.Size)})
            <span class="link-text">Loading...</span>
        </span>
    </a>
}
```

### Smart External Link
```go
templ SmartExternalLink(props LinkProps) {
    <a href={ props.Href }
       target="_blank"
       rel="noopener noreferrer"
       class={ fmt.Sprintf("%s external-link", props.GetClasses()) }
       aria-label={ fmt.Sprintf("%s (opens in new window)", props.Text) }>
        
        <span class="link-text">{ props.Text }</span>
        
        <span class="external-indicator" aria-hidden="true">
            @Icon(IconProps{Name: "external-link", Size: "xs"})
        </span>
        
        <!-- Screen reader warning -->
        <span class="sr-only">(opens in new window)</span>
    </a>
}
```

### Download Link
```go
templ DownloadLink(props LinkProps) {
    <a href={ props.Href }
       download={ props.Filename }
       class={ fmt.Sprintf("%s download-link", props.GetClasses()) }
       aria-label={ fmt.Sprintf("Download %s", props.Text) }>
        
        <span class="link-icon">
            @Icon(IconProps{Name: "download", Size: getLinkIconSize(props.Size)})
        </span>
        
        <span class="link-content">
            <span class="link-text">{ props.Text }</span>
            if props.FileSize != "" {
                <span class="file-size">({ props.FileSize })</span>
            }
        </span>
        
        <span class="download-indicator" aria-hidden="true">
            @Icon(IconProps{Name: "arrow-down", Size: "xs"})
        </span>
    </a>
}
```

## 📱 Responsive Design

### Mobile Optimizations
```css
/* Mobile-specific link styling */
@media (max-width: 479px) {
    .link {
        min-height: 44px;  /* Touch target */
        padding: var(--space-xs) 0;
        
        .link-icon {
            width: 20px;
            height: 20px;
        }
    }
    
    .link-button {
        min-height: 44px;
        padding: var(--space-md) var(--space-lg);
        font-size: var(--font-size-base);
    }
    
    .card-link {
        .link-card {
            .card-content {
                padding: var(--space-md);
                
                .card-title {
                    font-size: var(--font-size-base);
                }
            }
        }
    }
    
    .tab-navigation {
        overflow-x: auto;
        scrollbar-width: none;
        -ms-overflow-style: none;
        
        &::-webkit-scrollbar {
            display: none;
        }
        
        .tab-link {
            padding: var(--space-md);
            min-width: max-content;
        }
    }
    
    .breadcrumb-nav {
        .breadcrumb-list {
            font-size: var(--font-size-xs);
        }
    }
}
```

### Touch Interactions
```css
/* Touch-friendly interactions */
@media (hover: none) {
    .link {
        /* Remove hover effects on touch devices */
        &:hover {
            text-decoration: none;
        }
        
        /* Enhanced tap feedback */
        &:active {
            opacity: 0.7;
        }
    }
    
    .link-button {
        &:hover {
            transform: none;
            box-shadow: none;
        }
        
        &:active {
            transform: scale(0.98);
        }
    }
    
    .card-link {
        &:hover {
            transform: none;
            box-shadow: var(--shadow-md);
            
            .card-image img {
                transform: none;
            }
        }
        
        &:active {
            transform: scale(0.98);
        }
    }
    
    .tab-link {
        &:hover {
            background: none;
        }
        
        &:active {
            background: var(--color-bg-pressed);
        }
    }
}
```

## ♿ Accessibility Implementation

### ARIA Support
```go
func (link LinkProps) GetAriaAttributes() map[string]string {
    attrs := make(map[string]string)
    
    // Role based on usage
    if link.Variant == "button" {
        attrs["role"] = "button"
    } else if link.Variant == "tab" {
        attrs["role"] = "tab"
        attrs["aria-selected"] = fmt.Sprintf("%t", link.Active)
        if link.AriaControls != "" {
            attrs["aria-controls"] = link.AriaControls
        }
    }
    
    // Accessible name
    if link.AriaLabel != "" {
        attrs["aria-label"] = link.AriaLabel
    } else if link.External {
        attrs["aria-label"] = fmt.Sprintf("%s (opens in new window)", link.Text)
    } else if link.Download {
        attrs["aria-label"] = fmt.Sprintf("Download %s", link.Text)
    }
    
    // Description
    if link.AriaDescribedBy != "" {
        attrs["aria-describedby"] = link.AriaDescribedBy
    }
    
    // Current page
    if link.Active && link.Variant == "breadcrumb" {
        attrs["aria-current"] = "page"
    } else if link.Active {
        attrs["aria-current"] = "true"
    }
    
    // Disabled state
    if link.Disabled {
        attrs["aria-disabled"] = "true"
        attrs["tabindex"] = "-1"
    }
    
    return attrs
}
```

### Screen Reader Support
```go
templ AccessibleLink(props LinkProps) {
    <a href={ props.Href }
       target={ props.Target }
       rel={ props.Rel }
       class={ props.GetClasses() }
       for attrName, attrValue := range props.GetAriaAttributes() {
           { attrName }={ attrValue }
       }>
        
        <span class="link-text">{ props.Text }</span>
        
        if props.External {
            <span class="sr-only">(opens in new window)</span>
        }
        
        if props.Download {
            <span class="sr-only">(download file)</span>
        }
        
        if props.Loading {
            <span class="sr-only" aria-live="polite">Loading</span>
        }
    </a>
}
```

### Keyboard Navigation
```css
/* Focus management */
.link {
    &:focus {
        outline: none;
        box-shadow: 0 0 0 2px var(--color-primary-light);
        border-radius: var(--radius-sm);
    }
    
    &:focus-visible {
        box-shadow: 0 0 0 2px var(--color-primary);
    }
}

/* Skip links for keyboard users */
.skip-link {
    position: absolute;
    top: -40px;
    left: 6px;
    background: var(--color-bg-surface);
    color: var(--color-primary);
    padding: 8px;
    text-decoration: none;
    z-index: 1000;
    border-radius: var(--radius-sm);
    
    &:focus {
        top: 6px;
    }
}

/* High contrast mode support */
@media (prefers-contrast: high) {
    .link {
        text-decoration: underline;
        
        &:focus {
            outline: 2px solid currentColor;
            outline-offset: 2px;
        }
    }
    
    .link-button {
        border-width: 2px;
        border-style: solid;
    }
    
    .tab-link {
        border-width: 2px;
        
        &.tab-active {
            border-bottom-width: 4px;
        }
    }
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
    .link,
    .link-button,
    .card-link,
    .tab-link {
        transition: none;
        
        &:hover {
            transform: none;
        }
    }
}
```

## 🧪 Testing Guidelines

### Unit Tests
```go
func TestLinkComponent(t *testing.T) {
    tests := []struct {
        name     string
        props    LinkProps
        expected []string
    }{
        {
            name: "basic link",
            props: LinkProps{
                Text: "Learn more",
                Href: "/docs",
            },
            expected: []string{"link", "Learn more", "href=\"/docs\""},
        },
        {
            name: "external link",
            props: LinkProps{
                Text:     "External site",
                Href:     "https://example.com",
                External: true,
            },
            expected: []string{"external-link", "target=\"_blank\"", "rel=\"noopener noreferrer\""},
        },
        {
            name: "button link",
            props: LinkProps{
                Text:    "Get Started",
                Href:    "/signup",
                Variant: "button",
            },
            expected: []string{"link-button", "role=\"button\""},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            component := renderLink(tt.props)
            for _, expected := range tt.expected {
                assert.Contains(t, component, expected)
            }
        })
    }
}
```

### Accessibility Tests
```javascript
describe('Link Accessibility', () => {
    test('external links announce they open in new window', () => {
        const link = render(<Link text="External" href="https://example.com" external />);
        const linkElement = screen.getByRole('link');
        expect(linkElement).toHaveAttribute('aria-label', 'External (opens in new window)');
    });
    
    test('tab links have proper ARIA attributes', () => {
        const link = render(<Link text="Tab 1" href="#panel1" variant="tab" active />);
        const linkElement = screen.getByRole('tab');
        expect(linkElement).toHaveAttribute('aria-selected', 'true');
    });
    
    test('disabled links are not focusable', () => {
        const link = render(<Link text="Disabled" href="/test" disabled />);
        const linkElement = screen.getByRole('link');
        expect(linkElement).toHaveAttribute('aria-disabled', 'true');
        expect(linkElement).toHaveAttribute('tabindex', '-1');
    });
    
    test('supports keyboard navigation', () => {
        const link = render(<Link text="Test" href="/test" />);
        const linkElement = screen.getByRole('link');
        
        linkElement.focus();
        expect(linkElement).toHaveFocus();
        
        fireEvent.keyDown(linkElement, { key: 'Enter' });
        // Should navigate
    });
});
```

### Visual Regression Tests
```javascript
test.describe('Link Visual Tests', () => {
    test('all link variants and states', async ({ page }) => {
        await page.goto('/components/link');
        
        // Test all variants
        const variants = ['text', 'button', 'card', 'tab', 'breadcrumb'];
        for (const variant of variants) {
            await expect(page.locator(`[data-variant="${variant}"]`)).toHaveScreenshot(`link-${variant}.png`);
        }
        
        // Test all sizes
        for (const size of ['xs', 'sm', 'md', 'lg', 'xl']) {
            await expect(page.locator(`[data-size="${size}"]`)).toHaveScreenshot(`link-${size}.png`);
        }
        
        // Test all colors
        const colors = ['primary', 'secondary', 'success', 'warning', 'danger'];
        for (const color of colors) {
            await expect(page.locator(`[data-color="${color}"]`)).toHaveScreenshot(`link-${color}.png`);
        }
        
        // Test special states
        await expect(page.locator('[data-state="external"]')).toHaveScreenshot('link-external.png');
        await expect(page.locator('[data-state="download"]')).toHaveScreenshot('link-download.png');
    });
});
```

## 📚 Usage Examples

### Navigation Menu
```go
templ NavigationMenu() {
    <nav class="main-nav">
        @BasicLink(LinkProps{
            Text: "Home",
            Href: "/",
            Icon: "home",
        })
        
        @BasicLink(LinkProps{
            Text: "Products",
            Href: "/products",
            Icon: "package",
        })
        
        @BasicLink(LinkProps{
            Text: "Documentation",
            Href: "https://docs.example.com",
            Icon: "book",
            External: true,
        })
    </nav>
}
```

### Call-to-Action Buttons
```go
templ CTASection() {
    <section class="cta-section">
        @ButtonLink(LinkProps{
            Text:    "Start Free Trial",
            Href:    "/signup",
            Variant: "button",
            Color:   "primary",
            Size:    "lg",
            Icon:    "arrow-right",
            IconPosition: "right",
        })
        
        @BasicLink(LinkProps{
            Text:  "Learn more",
            Href:  "/features",
            Color: "secondary",
        })
    </section>
}
```

### Product Cards
```go
templ ProductGrid() {
    <div class="product-grid">
        @CardLink(LinkProps{
            Text:        "Premium Package",
            Href:        "/products/premium",
            Description: "Everything you need for professional projects",
            Image:       "/images/premium.jpg",
            ActionText:  "View details",
            Meta:        "$99/month",
        })
    </div>
}
```

### Tab Navigation
```go
templ DashboardTabs() {
    @TabNavigation(TabNavigationProps{
        AriaLabel: "Dashboard sections",
        Tabs: []TabItem{
            {Text: "Overview", Href: "/dashboard", Active: true, Icon: "home"},
            {Text: "Analytics", Href: "/dashboard/analytics", Icon: "bar-chart", Count: 5},
            {Text: "Settings", Href: "/dashboard/settings", Icon: "settings"},
        },
    })
}
```

## 🔗 Related Components

- **[Button](../button/)** - Action triggers and form submissions
- **[Navigation](../../organisms/navigation/)** - Complex navigation systems
- **[Breadcrumb](../../molecules/breadcrumb/)** - Navigation hierarchy
- **[Menu](../../molecules/menu/)** - Dropdown navigation menus

---

**Component Status**: ✅ Production Ready  
**Schema Reference**: `LinkSchema.json`  
**CSS Classes**: `.link`, `.link-{variant}`, `.link-{size}`, `.link-{color}`  
**Accessibility**: WCAG 2.1 AA Compliant