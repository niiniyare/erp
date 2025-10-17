# Icon Component

**FILE PURPOSE**: Visual symbol and iconography implementation and specifications  
**SCOPE**: All icon variants, sizes, and usage patterns  
**TARGET AUDIENCE**: Developers implementing visual symbols, indicators, and decorative elements

## 📋 Component Overview

The Icon component provides a consistent system for displaying visual symbols throughout the ERP interface. It supports multiple icon libraries, various sizes, colors, and interactive states while maintaining accessibility and performance standards.

### Schema Reference
- **Primary Schema**: `IconSchema.json`
- **Related Schemas**: `IconCheckedSchema.json`, `IconItemSchema.json`
- **Base Interface**: Visual element with semantic meaning

## 🎨 Icon Types

### Basic Icon
**Purpose**: Simple visual symbols for UI elements

```go
// Basic icon configuration
basicIcon := IconProps{
    Name:  "user",
    Size:  "md",
    Color: "current",
}

// Generated Templ component
templ BasicIcon(props IconProps) {
    <span class={ fmt.Sprintf("icon icon-%s", props.Size) }
          role={ getIconRole(props) }
          aria-label={ props.AriaLabel }
          aria-hidden={ fmt.Sprintf("%t", props.Decorative) }>
        
        switch props.Library {
        case "heroicons":
            @HeroIcon(props)
        case "lucide":
            @LucideIcon(props)
        case "feather":
            @FeatherIcon(props)
        case "custom":
            @CustomIcon(props)
        default:
            @HeroIcon(props) // Default to Heroicons
        }
    </span>
}

// Hero Icons implementation
templ HeroIcon(props IconProps) {
    switch props.Name {
    case "user":
        <svg class="icon-svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
        </svg>
    case "envelope":
        <svg class="icon-svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75" />
        </svg>
    case "check":
        <svg class="icon-svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
        </svg>
    case "x":
        <svg class="icon-svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
        </svg>
    default:
        @UnknownIcon(props.Name)
    }
}
```

### Interactive Icon
**Purpose**: Clickable icons with hover states and actions

```go
interactiveIcon := IconProps{
    Name:        "heart",
    Size:        "lg",
    Interactive: true,
    OnClick:     "toggleFavorite",
    AriaLabel:   "Add to favorites",
}

templ InteractiveIcon(props IconProps) {
    <button 
        type="button"
        class={ fmt.Sprintf("icon-button icon-button-%s", props.Size) }
        @click={ props.OnClick }
        aria-label={ props.AriaLabel }
        :class="{ 'icon-active': active }"
        x-data="{ active: false }"
        @click="active = !active">
        
        <span class="icon-container">
            @BasicIcon(IconProps{
                Name:       props.Name,
                Size:       props.Size,
                Library:    props.Library,
                Decorative: true,
            })
        </span>
        
        if props.Badge != "" {
            <span class="icon-badge">{ props.Badge }</span>
        }
        
        if props.Count > 0 {
            <span class="icon-count">{ formatCount(props.Count) }</span>
        }
    </button>
}
```

### Status Icon
**Purpose**: Icons with semantic colors for status indication

```go
statusIcon := IconProps{
    Name:   "check-circle",
    Status: "success",
    Size:   "md",
}

templ StatusIcon(props IconProps) {
    <span class={ fmt.Sprintf("status-icon status-%s icon-%s", props.Status, props.Size) }
          role="img"
          aria-label={ fmt.Sprintf("%s status", props.Status) }>
        
        @BasicIcon(IconProps{
            Name:       props.Name,
            Size:       props.Size,
            Library:    props.Library,
            Decorative: true,
        })
    </span>
}

// Status types
type IconStatus string

const (
    StatusSuccess IconStatus = "success"  // Green - completed, valid
    StatusWarning IconStatus = "warning"  // Yellow - caution, pending
    StatusError   IconStatus = "error"    // Red - failed, invalid
    StatusInfo    IconStatus = "info"     // Blue - informational
    StatusNeutral IconStatus = "neutral"  // Gray - default, inactive
)
```

### Animated Icon
**Purpose**: Icons with loading states and transitions

```go
animatedIcon := IconProps{
    Name:      "refresh",
    Size:      "md",
    Animation: "spin",
    Duration:  1000,
}

templ AnimatedIcon(props IconProps) {
    <span class={ fmt.Sprintf("animated-icon icon-%s animation-%s", props.Size, props.Animation) }
          style={ fmt.Sprintf("--animation-duration: %dms", props.Duration) }
          role="img"
          aria-label={ props.AriaLabel }>
        
        @BasicIcon(IconProps{
            Name:       props.Name,
            Size:       props.Size,
            Library:    props.Library,
            Decorative: true,
        })
    </span>
}

// Animation types
type IconAnimation string

const (
    AnimationSpin   IconAnimation = "spin"    // Continuous rotation
    AnimationPulse  IconAnimation = "pulse"   // Opacity pulsing
    AnimationBounce IconAnimation = "bounce"  // Bounce effect
    AnimationPing   IconAnimation = "ping"    // Ping/ripple effect
    AnimationNone   IconAnimation = "none"    // No animation
)
```

### Icon with Background
**Purpose**: Icons with colored backgrounds and shapes

```go
backgroundIcon := IconProps{
    Name:       "user",
    Size:       "lg",
    Background: "circle",
    Color:      "white",
    BgColor:    "primary",
}

templ BackgroundIcon(props IconProps) {
    <div class={ fmt.Sprintf("icon-with-background bg-%s bg-shape-%s icon-%s", props.BgColor, props.Background, props.Size) }
         role="img"
         aria-label={ props.AriaLabel }>
        
        @BasicIcon(IconProps{
            Name:       props.Name,
            Size:       props.Size,
            Color:      props.Color,
            Library:    props.Library,
            Decorative: true,
        })
    </div>
}

// Background shapes
type IconBackground string

const (
    BgCircle  IconBackground = "circle"   // Circular background
    BgSquare  IconBackground = "square"   // Square background
    BgRounded IconBackground = "rounded"  // Rounded square
    BgNone    IconBackground = "none"     // No background
)
```

## 🎯 Props Interface

### Core Properties
```go
type IconProps struct {
    // Identity
    Name   string `json:"name"`     // Icon identifier
    ID     string `json:"id"`
    TestID string `json:"testid"`
    
    // Appearance
    Size     IconSize      `json:"size"`      // xs, sm, md, lg, xl, 2xl, 3xl
    Color    string        `json:"color"`     // Icon color
    Library  IconLibrary   `json:"library"`   // heroicons, lucide, feather, custom
    Variant  IconVariant   `json:"variant"`   // outline, solid, mini
    
    // Background
    Background IconBackground `json:"background"` // circle, square, rounded, none
    BgColor    string         `json:"bgColor"`    // Background color
    BgSize     string         `json:"bgSize"`     // Background size
    
    // Status
    Status IconStatus `json:"status"` // success, warning, error, info, neutral
    
    // Animation
    Animation IconAnimation `json:"animation"` // spin, pulse, bounce, ping, none
    Duration  int           `json:"duration"`  // Animation duration in ms
    
    // Interaction
    Interactive bool   `json:"interactive"` // Make icon clickable
    OnClick     string `json:"onClick"`
    OnHover     string `json:"onHover"`
    
    // Badges and counts
    Badge string `json:"badge"` // Badge text
    Count int    `json:"count"` // Notification count
    
    // States
    Active   bool `json:"active"`   // Active state
    Disabled bool `json:"disabled"` // Disabled state
    Loading  bool `json:"loading"`  // Loading state
    
    // Styling
    Class string            `json:"className"`
    Style map[string]string `json:"style"`
    
    // Accessibility
    AriaLabel   string `json:"ariaLabel"`
    AriaHidden  bool   `json:"ariaHidden"`
    Decorative  bool   `json:"decorative"`  // Purely decorative (aria-hidden="true")
    Role        string `json:"role"`
    Title       string `json:"title"`       // Tooltip text
}
```

### Size Variants
```go
type IconSize string

const (
    IconXS  IconSize = "xs"   // 12px
    IconSM  IconSize = "sm"   // 16px
    IconMD  IconSize = "md"   // 20px (default)
    IconLG  IconSize = "lg"   // 24px
    IconXL  IconSize = "xl"   // 32px
    Icon2XL IconSize = "2xl"  // 48px
    Icon3XL IconSize = "3xl"  // 64px
)
```

### Icon Libraries
```go
type IconLibrary string

const (
    LibraryHeroicons IconLibrary = "heroicons" // Heroicons (default)
    LibraryLucide    IconLibrary = "lucide"    // Lucide React
    LibraryFeather   IconLibrary = "feather"   // Feather Icons
    LibraryCustom    IconLibrary = "custom"    // Custom SVG icons
)
```

### Icon Variants
```go
type IconVariant string

const (
    VariantOutline IconVariant = "outline" // Outline style (default)
    VariantSolid   IconVariant = "solid"   // Filled style
    VariantMini    IconVariant = "mini"    // Mini size style
)
```

## 🎨 Styling Implementation

### Base Icon Styles
```css
.icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    color: currentColor;
    
    .icon-svg {
        width: 100%;
        height: 100%;
        display: block;
        fill: currentColor;
        stroke: currentColor;
    }
}

/* Size variants */
.icon-xs {
    width: 12px;
    height: 12px;
}

.icon-sm {
    width: 16px;
    height: 16px;
}

.icon-md {
    width: 20px;
    height: 20px;
}

.icon-lg {
    width: 24px;
    height: 24px;
}

.icon-xl {
    width: 32px;
    height: 32px;
}

.icon-2xl {
    width: 48px;
    height: 48px;
}

.icon-3xl {
    width: 64px;
    height: 64px;
}
```

### Interactive Icon Styles
```css
.icon-button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    position: relative;
    background: none;
    border: none;
    cursor: pointer;
    border-radius: var(--radius-md);
    transition: var(--transition-base);
    color: var(--color-text-secondary);
    
    &:hover {
        color: var(--color-text-primary);
        background: var(--color-bg-hover);
    }
    
    &:focus {
        outline: none;
        box-shadow: 0 0 0 2px var(--color-primary-light);
    }
    
    &:active {
        background: var(--color-bg-pressed);
        transform: scale(0.95);
    }
    
    &.icon-active {
        color: var(--color-primary);
        background: var(--color-primary-light);
    }
    
    &:disabled {
        opacity: 0.5;
        cursor: not-allowed;
        
        &:hover {
            color: var(--color-text-secondary);
            background: none;
        }
    }
    
    .icon-container {
        display: flex;
        align-items: center;
        justify-content: center;
    }
    
    .icon-badge {
        position: absolute;
        top: -4px;
        right: -4px;
        background: var(--color-primary);
        color: white;
        font-size: 10px;
        font-weight: var(--font-weight-bold);
        padding: 1px 4px;
        border-radius: var(--radius-full);
        min-width: 16px;
        height: 16px;
        display: flex;
        align-items: center;
        justify-content: center;
    }
    
    .icon-count {
        position: absolute;
        top: -8px;
        right: -8px;
        background: var(--color-danger);
        color: white;
        font-size: 11px;
        font-weight: var(--font-weight-bold);
        padding: 2px 6px;
        border-radius: var(--radius-full);
        min-width: 18px;
        height: 18px;
        display: flex;
        align-items: center;
        justify-content: center;
    }
}

/* Size-specific button padding */
.icon-button-xs {
    padding: 2px;
}

.icon-button-sm {
    padding: 4px;
}

.icon-button-md {
    padding: 6px;
}

.icon-button-lg {
    padding: 8px;
}

.icon-button-xl {
    padding: 12px;
}
```

### Status Icon Styles
```css
.status-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    
    &.status-success {
        color: var(--color-success);
    }
    
    &.status-warning {
        color: var(--color-warning);
    }
    
    &.status-error {
        color: var(--color-error);
    }
    
    &.status-info {
        color: var(--color-info);
    }
    
    &.status-neutral {
        color: var(--color-text-secondary);
    }
}
```

### Background Icon Styles
```css
.icon-with-background {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    position: relative;
    
    /* Background shapes */
    &.bg-shape-circle {
        border-radius: 50%;
    }
    
    &.bg-shape-square {
        border-radius: 0;
    }
    
    &.bg-shape-rounded {
        border-radius: var(--radius-md);
    }
    
    /* Background colors */
    &.bg-primary {
        background: var(--color-primary);
        color: white;
    }
    
    &.bg-secondary {
        background: var(--color-bg-secondary);
        color: var(--color-text-secondary);
    }
    
    &.bg-success {
        background: var(--color-success);
        color: white;
    }
    
    &.bg-warning {
        background: var(--color-warning);
        color: var(--color-text-primary);
    }
    
    &.bg-error {
        background: var(--color-error);
        color: white;
    }
    
    &.bg-info {
        background: var(--color-info);
        color: white;
    }
    
    /* Size-specific background dimensions */
    &.icon-xs {
        width: 24px;
        height: 24px;
    }
    
    &.icon-sm {
        width: 32px;
        height: 32px;
    }
    
    &.icon-md {
        width: 40px;
        height: 40px;
    }
    
    &.icon-lg {
        width: 48px;
        height: 48px;
    }
    
    &.icon-xl {
        width: 64px;
        height: 64px;
    }
}
```

### Animation Styles
```css
.animated-icon {
    &.animation-spin {
        animation: spin var(--animation-duration, 1s) linear infinite;
    }
    
    &.animation-pulse {
        animation: pulse var(--animation-duration, 2s) cubic-bezier(0.4, 0, 0.6, 1) infinite;
    }
    
    &.animation-bounce {
        animation: bounce var(--animation-duration, 1s) infinite;
    }
    
    &.animation-ping {
        animation: ping var(--animation-duration, 1s) cubic-bezier(0, 0, 0.2, 1) infinite;
    }
}

@keyframes spin {
    from {
        transform: rotate(0deg);
    }
    to {
        transform: rotate(360deg);
    }
}

@keyframes pulse {
    0%, 100% {
        opacity: 1;
    }
    50% {
        opacity: 0.5;
    }
}

@keyframes bounce {
    0%, 100% {
        transform: translateY(-25%);
        animation-timing-function: cubic-bezier(0.8, 0, 1, 1);
    }
    50% {
        transform: translateY(0);
        animation-timing-function: cubic-bezier(0, 0, 0.2, 1);
    }
}

@keyframes ping {
    75%, 100% {
        transform: scale(2);
        opacity: 0;
    }
}
```

## ⚙️ Advanced Features

### Dynamic Icon Loading
```go
templ DynamicIcon(props IconProps) {
    <div class="dynamic-icon" 
         x-data={ fmt.Sprintf(`{
            loading: true,
            iconName: '%s',
            iconLibrary: '%s',
            async loadIcon() {
                try {
                    const response = await fetch('/api/icons/' + this.iconLibrary + '/' + this.iconName);
                    const iconSvg = await response.text();
                    this.$el.querySelector('.icon-placeholder').innerHTML = iconSvg;
                } catch (error) {
                    console.error('Failed to load icon:', error);
                    this.$el.querySelector('.icon-placeholder').innerHTML = this.getFallbackIcon();
                } finally {
                    this.loading = false;
                }
            },
            getFallbackIcon() {
                return '<svg class="icon-svg" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z" /></svg>';
            }
        }`, props.Name, props.Library) }
         x-init="loadIcon()">
        
        <div class="icon-placeholder" 
             :class="{ 'loading': loading }"
             role="img"
             :aria-label="iconName">
            
            <div x-show="loading" class="icon-skeleton">
                @AnimatedIcon(IconProps{
                    Name:      "refresh",
                    Size:      props.Size,
                    Animation: "spin",
                })
            </div>
        </div>
    </div>
}
```

### Icon Picker
```go
templ IconPicker(props IconPickerProps) {
    <div class="icon-picker" 
         x-data={ fmt.Sprintf(`{
            selectedIcon: '%s',
            searchTerm: '',
            category: 'all',
            icons: %s,
            get filteredIcons() {
                return this.icons.filter(icon => {
                    const matchesSearch = icon.name.toLowerCase().includes(this.searchTerm.toLowerCase()) ||
                                        icon.tags.some(tag => tag.toLowerCase().includes(this.searchTerm.toLowerCase()));
                    const matchesCategory = this.category === 'all' || icon.category === this.category;
                    return matchesSearch && matchesCategory;
                });
            },
            selectIcon(iconName) {
                this.selectedIcon = iconName;
                $dispatch('icon-selected', { icon: iconName });
            }
        }`, props.SelectedIcon, toJSON(props.AvailableIcons)) }>
        
        <div class="picker-header">
            <div class="search-input">
                <input 
                    type="text"
                    x-model="searchTerm"
                    placeholder="Search icons..."
                    class="form-input"
                />
                @Icon(IconProps{Name: "search", Size: "sm"})
            </div>
            
            <div class="category-filter">
                <select x-model="category" class="form-select">
                    <option value="all">All Categories</option>
                    <option value="ui">UI Elements</option>
                    <option value="navigation">Navigation</option>
                    <option value="communication">Communication</option>
                    <option value="media">Media</option>
                    <option value="system">System</option>
                </select>
            </div>
        </div>
        
        <div class="picker-grid">
            <template x-for="icon in filteredIcons" :key="icon.name">
                <button 
                    type="button"
                    @click="selectIcon(icon.name)"
                    :class="{ 'selected': selectedIcon === icon.name }"
                    class="icon-option"
                    :title="icon.name">
                    
                    <div class="option-icon" x-html="icon.svg"></div>
                    <span class="option-name" x-text="icon.name"></span>
                </button>
            </template>
        </div>
        
        <div x-show="filteredIcons.length === 0" class="no-results">
            <div class="no-results-icon">
                @Icon(IconProps{Name: "search-x", Size: "lg"})
            </div>
            <p>No icons found matching your search.</p>
        </div>
    </div>
}
```

### Icon Set Management
```go
templ IconSet(props IconSetProps) {
    <div class="icon-set" x-data={ fmt.Sprintf(`{
        icons: %s,
        selectedIcons: [],
        get allSelected() {
            return this.selectedIcons.length === this.icons.length;
        },
        get someSelected() {
            return this.selectedIcons.length > 0 && this.selectedIcons.length < this.icons.length;
        },
        toggleAll() {
            if (this.allSelected) {
                this.selectedIcons = [];
            } else {
                this.selectedIcons = this.icons.map(icon => icon.name);
            }
        },
        toggleIcon(iconName) {
            if (this.selectedIcons.includes(iconName)) {
                this.selectedIcons = this.selectedIcons.filter(name => name !== iconName);
            } else {
                this.selectedIcons.push(iconName);
            }
        }
    }`, toJSON(props.Icons)) }>
        
        <div class="set-header">
            <h3 class="set-title">{ props.Title }</h3>
            
            <div class="set-actions">
                <label class="select-all">
                    <input 
                        type="checkbox"
                        :checked="allSelected"
                        :indeterminate="someSelected"
                        @change="toggleAll()"
                    />
                    <span>Select All</span>
                </label>
                
                <span class="selection-count" x-text="`${selectedIcons.length} selected`"></span>
            </div>
        </div>
        
        <div class="icon-grid">
            <template x-for="icon in icons" :key="icon.name">
                <div class="icon-item" 
                     :class="{ 'selected': selectedIcons.includes(icon.name) }">
                    
                    <label class="icon-checkbox">
                        <input 
                            type="checkbox"
                            :checked="selectedIcons.includes(icon.name)"
                            @change="toggleIcon(icon.name)"
                        />
                        <span class="checkbox-indicator"></span>
                    </label>
                    
                    <div class="icon-display">
                        <div class="icon-preview" x-html="icon.svg"></div>
                        <span class="icon-name" x-text="icon.name"></span>
                    </div>
                    
                    <div class="icon-info">
                        <span class="icon-size" x-text="icon.size"></span>
                        <span class="icon-format" x-text="icon.format"></span>
                    </div>
                </div>
            </template>
        </div>
    </div>
}
```

## 📱 Responsive Design

### Mobile Optimizations
```css
/* Mobile-specific icon styling */
@media (max-width: 479px) {
    .icon-button {
        min-width: 44px;  /* Touch target */
        min-height: 44px;
        padding: 8px;
        
        .icon-count,
        .icon-badge {
            font-size: 12px;
            min-width: 20px;
            height: 20px;
        }
    }
    
    /* Larger icons for better visibility */
    .icon-xs { width: 14px; height: 14px; }
    .icon-sm { width: 18px; height: 18px; }
    .icon-md { width: 22px; height: 22px; }
    .icon-lg { width: 26px; height: 26px; }
    
    /* Simplify icon picker on mobile */
    .icon-picker {
        .picker-grid {
            grid-template-columns: repeat(4, 1fr);
            gap: var(--space-sm);
        }
        
        .icon-option {
            flex-direction: column;
            padding: var(--space-sm);
            
            .option-name {
                font-size: var(--font-size-xs);
                margin-top: var(--space-xs);
            }
        }
    }
}
```

### Touch Interactions
```css
/* Touch-friendly interactions */
@media (hover: none) {
    .icon-button {
        /* Remove hover effects on touch devices */
        &:hover {
            color: var(--color-text-secondary);
            background: none;
        }
        
        /* Enhanced tap feedback */
        &:active {
            background: var(--color-bg-pressed);
            transform: scale(0.9);
        }
    }
    
    .icon-option {
        &:hover {
            background: var(--color-bg-surface);
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
func (icon IconProps) GetAriaAttributes() map[string]string {
    attrs := make(map[string]string)
    
    // Decorative icons should be hidden from screen readers
    if icon.Decorative {
        attrs["aria-hidden"] = "true"
        return attrs
    }
    
    // Role based on usage
    if icon.Interactive {
        attrs["role"] = "button"
        attrs["tabindex"] = "0"
    } else {
        attrs["role"] = "img"
    }
    
    // Accessible name
    if icon.AriaLabel != "" {
        attrs["aria-label"] = icon.AriaLabel
    } else if icon.Title != "" {
        attrs["aria-label"] = icon.Title
    } else if !icon.Decorative {
        attrs["aria-label"] = icon.Name
    }
    
    // Loading state
    if icon.Loading {
        attrs["aria-busy"] = "true"
    }
    
    // Disabled state
    if icon.Disabled {
        attrs["aria-disabled"] = "true"
    }
    
    return attrs
}

func getIconRole(props IconProps) string {
    if props.Interactive {
        return "button"
    }
    if props.Decorative {
        return ""
    }
    return "img"
}
```

### Screen Reader Support
```go
templ AccessibleIcon(props IconProps) {
    <span class={ props.GetClasses() }
          for attrName, attrValue := range props.GetAriaAttributes() {
              { attrName }={ attrValue }
          }>
        
        @BasicIcon(props)
        
        <!-- Provide text alternative for complex icons -->
        if !props.Decorative && props.Description != "" {
            <span class="sr-only">{ props.Description }</span>
        }
        
        <!-- Loading announcement -->
        if props.Loading {
            <span class="sr-only" aria-live="polite">Loading</span>
        }
    </span>
}
```

### Keyboard Navigation
```css
/* Focus management */
.icon-button {
    &:focus {
        outline: none;
        box-shadow: 0 0 0 2px var(--color-primary-light);
        border-radius: var(--radius-md);
    }
    
    &:focus-visible {
        box-shadow: 0 0 0 2px var(--color-primary);
    }
}

/* High contrast mode support */
@media (prefers-contrast: high) {
    .icon {
        .icon-svg {
            stroke-width: 2;
        }
    }
    
    .status-icon {
        &.status-success { color: #00FF00; }
        &.status-error { color: #FF0000; }
        &.status-warning { color: #FFFF00; }
        &.status-info { color: #0000FF; }
    }
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
    .animated-icon {
        animation: none;
    }
    
    .icon-button {
        transition: none;
        
        &:active {
            transform: none;
        }
    }
}
```

## 🧪 Testing Guidelines

### Unit Tests
```go
func TestIconComponent(t *testing.T) {
    tests := []struct {
        name     string
        props    IconProps
        expected []string
    }{
        {
            name: "basic icon",
            props: IconProps{
                Name: "user",
                Size: "md",
            },
            expected: []string{"icon", "icon-md", "user"},
        },
        {
            name: "interactive icon",
            props: IconProps{
                Name:        "heart",
                Interactive: true,
                OnClick:     "toggleFavorite",
            },
            expected: []string{"icon-button", "role=\"button\""},
        },
        {
            name: "decorative icon",
            props: IconProps{
                Name:       "decoration",
                Decorative: true,
            },
            expected: []string{"aria-hidden=\"true\""},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            component := renderIcon(tt.props)
            for _, expected := range tt.expected {
                assert.Contains(t, component, expected)
            }
        })
    }
}
```

### Accessibility Tests
```javascript
describe('Icon Accessibility', () => {
    test('decorative icons are hidden from screen readers', () => {
        const icon = render(<Icon name="decoration" decorative />);
        const iconElement = screen.getByRole('img', { hidden: true });
        expect(iconElement).toHaveAttribute('aria-hidden', 'true');
    });
    
    test('interactive icons have proper role', () => {
        const icon = render(<Icon name="heart" interactive onClick={() => {}} />);
        const iconElement = screen.getByRole('button');
        expect(iconElement).toBeInTheDocument();
    });
    
    test('status icons announce their meaning', () => {
        const icon = render(<Icon name="check" status="success" />);
        const iconElement = screen.getByRole('img');
        expect(iconElement).toHaveAttribute('aria-label', 'success status');
    });
    
    test('supports keyboard interaction', () => {
        const handleClick = jest.fn();
        const icon = render(<Icon name="star" interactive onClick={handleClick} />);
        const iconElement = screen.getByRole('button');
        
        iconElement.focus();
        expect(iconElement).toHaveFocus();
        
        fireEvent.keyDown(iconElement, { key: 'Enter' });
        expect(handleClick).toHaveBeenCalled();
    });
});
```

### Visual Regression Tests
```javascript
test.describe('Icon Visual Tests', () => {
    test('all icon sizes and states', async ({ page }) => {
        await page.goto('/components/icon');
        
        // Test all sizes
        for (const size of ['xs', 'sm', 'md', 'lg', 'xl', '2xl', '3xl']) {
            await expect(page.locator(`[data-size="${size}"]`)).toHaveScreenshot(`icon-${size}.png`);
        }
        
        // Test all states
        const states = ['default', 'interactive', 'status', 'animated', 'background'];
        for (const state of states) {
            await expect(page.locator(`[data-state="${state}"]`)).toHaveScreenshot(`icon-${state}.png`);
        }
        
        // Test animations
        const animations = ['spin', 'pulse', 'bounce', 'ping'];
        for (const animation of animations) {
            await expect(page.locator(`[data-animation="${animation}"]`)).toHaveScreenshot(`icon-${animation}.png`);
        }
    });
});
```

## 📚 Usage Examples

### Navigation Icons
```go
templ NavigationMenu() {
    <nav class="main-nav">
        @InteractiveIcon(IconProps{
            Name:      "home",
            Size:      "md",
            AriaLabel: "Home",
            OnClick:   "navigateHome",
        })
        
        @InteractiveIcon(IconProps{
            Name:      "user",
            Size:      "md",
            AriaLabel: "Profile",
            OnClick:   "navigateProfile",
        })
        
        @InteractiveIcon(IconProps{
            Name:      "settings",
            Size:      "md",
            AriaLabel: "Settings",
            OnClick:   "navigateSettings",
        })
    </nav>
}
```

### Status Indicators
```go
templ ConnectionStatus() {
    <div class="connection-status">
        <span>Connection Status:</span>
        @StatusIcon(IconProps{
            Name:   "wifi",
            Status: "success",
            Size:   "sm",
        })
        <span>Connected</span>
    </div>
}

templ LoadingIndicator() {
    <div class="loading-indicator">
        @AnimatedIcon(IconProps{
            Name:      "refresh",
            Size:      "lg",
            Animation: "spin",
            AriaLabel: "Loading content",
        })
        <span>Loading...</span>
    </div>
}
```

### Action Buttons
```go
templ ActionButtons() {
    <div class="action-buttons">
        @InteractiveIcon(IconProps{
            Name:      "heart",
            Size:      "md",
            AriaLabel: "Add to favorites",
            OnClick:   "toggleFavorite",
        })
        
        @InteractiveIcon(IconProps{
            Name:      "share",
            Size:      "md",
            AriaLabel: "Share item",
            OnClick:   "shareItem",
        })
        
        @InteractiveIcon(IconProps{
            Name:      "bookmark",
            Size:      "md",
            AriaLabel: "Bookmark item",
            OnClick:   "bookmarkItem",
            Count:     3,
        })
    </div>
}
```

### User Avatar with Status
```go
templ UserAvatar() {
    <div class="user-avatar">
        @BackgroundIcon(IconProps{
            Name:       "user",
            Size:       "xl",
            Background: "circle",
            BgColor:    "primary",
            Color:      "white",
        })
        
        @StatusIcon(IconProps{
            Name:   "circle",
            Status: "success",
            Size:   "sm",
        })
    </div>
}
```

## 🔗 Related Components

- **[Button](../button/)** - Action triggers using icons
- **[Badge](../badge/)** - Status indicators with icons
- **[Status](../status/)** - State indicators with icons
- **[Avatar](../../molecules/avatar/)** - User representations with icons

---

**Component Status**: ✅ Production Ready  
**Schema Reference**: `IconSchema.json`  
**CSS Classes**: `.icon`, `.icon-{size}`, `.status-{status}`  
**Accessibility**: WCAG 2.1 AA Compliant