# Badge Component

**FILE PURPOSE**: Status indicator and label implementation and specifications  
**SCOPE**: All badge variants, styles, and usage patterns  
**TARGET AUDIENCE**: Developers implementing status indicators, labels, and notifications

## 📋 Component Overview

The Badge component provides visual indicators for status, counts, labels, and notifications. It supports various styles, colors, and positioning options while maintaining consistent design and accessibility standards.

### Schema Reference
- **Primary Schema**: `BadgeObject.json`
- **Related Schemas**: `StatusSchema.json`, `TagSchema.json`
- **Base Interface**: Display element with semantic meaning

## 🎨 Badge Types

### Basic Badge
**Purpose**: Simple status indicators and labels

```go
// Basic badge configuration
basicBadge := BadgeProps{
    Text:    "New",
    Variant: "primary",
    Size:    "md",
}

// Generated Templ component
templ BasicBadge(props BadgeProps) {
    <span class={ fmt.Sprintf("badge badge-%s badge-%s", props.Size, props.Variant) }
          id={ props.ID }
          role="status"
          aria-label={ props.AriaLabel }>
        
        if props.Icon != "" {
            @Icon(IconProps{Name: props.Icon, Size: getBadgeIconSize(props.Size)})
        }
        
        if props.Text != "" {
            <span class="badge-text">{ props.Text }</span>
        }
        
        if props.Count > 0 {
            <span class="badge-count">{ formatCount(props.Count) }</span>
        }
        
        if props.Dismissible {
            <button 
                type="button"
                class="badge-dismiss"
                @click={ props.OnDismiss }
                aria-label="Remove badge">
                @Icon(IconProps{Name: "x", Size: "xs"})
            </button>
        }
    </span>
}
```

### Status Badge
**Purpose**: Semantic status indicators with consistent colors

```go
statusBadge := BadgeProps{
    Text:    "Active",
    Status:  "success",
    WithDot: true,
}

templ StatusBadge(props BadgeProps) {
    <span class={ fmt.Sprintf("badge status-badge status-%s", props.Status) }
          role="status"
          aria-label={ fmt.Sprintf("Status: %s", props.Text) }>
        
        if props.WithDot {
            <span class="status-dot" aria-hidden="true"></span>
        }
        
        <span class="badge-text">{ props.Text }</span>
    </span>
}

// Status options
type BadgeStatus string

const (
    StatusSuccess BadgeStatus = "success"  // Green - operational, active
    StatusWarning BadgeStatus = "warning"  // Yellow - needs attention
    StatusDanger  BadgeStatus = "danger"   // Red - error, critical
    StatusInfo    BadgeStatus = "info"     // Blue - informational
    StatusNeutral BadgeStatus = "neutral"  // Gray - inactive, draft
)
```

### Count Badge
**Purpose**: Notification counts and numeric indicators

```go
countBadge := BadgeProps{
    Count:    42,
    Variant:  "danger",
    Position: "top-right",
    Max:      99,
}

templ CountBadge(props BadgeProps) {
    <span class={ fmt.Sprintf("badge count-badge badge-%s", props.Variant) }
          role="status"
          aria-label={ fmt.Sprintf("%d notifications", props.Count) }>
        
        <span class="count-text">
            if props.Count > props.Max {
                { fmt.Sprintf("%d+", props.Max) }
            } else {
                { fmt.Sprintf("%d", props.Count) }
            }
        </span>
    </span>
}

// Usage with positioned badges
templ BadgeWrapper(children templ.Component, badge BadgeProps) {
    <div class="badge-wrapper" style="position: relative; display: inline-block;">
        @children
        
        if badge.Count > 0 || badge.Text != "" {
            <span class={ fmt.Sprintf("badge-positioned badge-position-%s", badge.Position) }>
                @CountBadge(badge)
            </span>
        }
    </div>
}
```

### Interactive Badge
**Purpose**: Clickable badges with actions and states

```go
interactiveBadge := BadgeProps{
    Text:        "Filter: Active",
    Variant:     "secondary",
    Clickable:   true,
    Dismissible: true,
    OnClick:     "handleFilterClick",
    OnDismiss:   "removeFilter",
}

templ InteractiveBadge(props BadgeProps) {
    <span class={ fmt.Sprintf("badge interactive-badge badge-%s", props.Variant) }
          x-data="{ hover: false }"
          @mouseenter="hover = true"
          @mouseleave="hover = false"
          role={ templ.Attributes{"button": props.Clickable, "status": !props.Clickable} }
          :class="{ 'badge-hover': hover }"
          tabindex={ templ.Attributes{"0": props.Clickable, "-1": !props.Clickable} }>
        
        if props.Clickable {
            <button 
                type="button"
                class="badge-action"
                @click={ props.OnClick }
                @keydown.enter={ props.OnClick }
                @keydown.space.prevent={ props.OnClick }>
        }
        
        if props.Icon != "" {
            @Icon(IconProps{Name: props.Icon, Size: getBadgeIconSize(props.Size)})
        }
        
        <span class="badge-text">{ props.Text }</span>
        
        if props.Clickable {
            </button>
        }
        
        if props.Dismissible {
            <button 
                type="button"
                class="badge-dismiss"
                @click={ props.OnDismiss }
                aria-label="Remove badge"
                @keydown.enter={ props.OnDismiss }
                @keydown.space.prevent={ props.OnDismiss }>
                @Icon(IconProps{Name: "x", Size: "xs"})
            </button>
        }
    </span>
}
```

### Badge Group
**Purpose**: Multiple related badges organized together

```go
badgeGroup := BadgeGroupProps{
    Title:  "Tags",
    Badges: []BadgeProps{
        {Text: "React", Variant: "primary"},
        {Text: "TypeScript", Variant: "info"},
        {Text: "Tailwind", Variant: "success"},
    },
    Layout:      "wrap",
    Dismissible: true,
}

templ BadgeGroup(props BadgeGroupProps) {
    <div class="badge-group" x-data="{ badges: JSON.parse('{{ toJSON(props.Badges) }}') }">
        if props.Title != "" {
            <span class="badge-group-title">{ props.Title }</span>
        }
        
        <div class={ fmt.Sprintf("badge-list badge-layout-%s", props.Layout) }>
            <template x-for="(badge, index) in badges" :key="index">
                <span class="badge badge-secondary badge-dismissible">
                    <span class="badge-text" x-text="badge.text"></span>
                    if props.Dismissible {
                        <button 
                            type="button"
                            class="badge-dismiss"
                            @click="badges.splice(index, 1)"
                            aria-label="Remove badge">
                            @Icon(IconProps{Name: "x", Size: "xs"})
                        </button>
                    }
                </span>
            </template>
        </div>
        
        if props.AddMore {
            <button 
                type="button"
                class="badge badge-outline badge-add"
                @click="$dispatch('badge-add')"
                aria-label="Add new badge">
                @Icon(IconProps{Name: "plus", Size: "xs"})
                <span>Add</span>
            </button>
        }
    </div>
}
```

## 🎯 Props Interface

### Core Properties
```go
type BadgeProps struct {
    // Identity
    ID     string `json:"id"`
    TestID string `json:"testid"`
    
    // Content
    Text  string `json:"text"`           // Display text
    Count int    `json:"count"`          // Numeric count
    Icon  string `json:"icon"`           // Icon name
    
    // Appearance
    Size     BadgeSize     `json:"size"`     // xs, sm, md, lg
    Variant  BadgeVariant  `json:"variant"`  // primary, secondary, success, etc.
    Status   BadgeStatus   `json:"status"`   // success, warning, danger, info, neutral
    Color    string        `json:"color"`    // Custom color
    Outline  bool          `json:"outline"`  // Outline style
    Rounded  bool          `json:"rounded"`  // Fully rounded corners
    
    // Features
    WithDot      bool   `json:"withDot"`      // Show status dot
    Dismissible  bool   `json:"dismissible"`  // Show dismiss button
    Clickable    bool   `json:"clickable"`    // Make badge clickable
    Pulsing      bool   `json:"pulsing"`      // Pulsing animation
    
    // Positioning
    Position BadgePosition `json:"position"`  // top-left, top-right, etc.
    
    // Count specific
    Max       int  `json:"max"`        // Maximum count before showing "+"
    ShowZero  bool `json:"showZero"`   // Show badge when count is 0
    
    // Styling
    Class string            `json:"className"`
    Style map[string]string `json:"style"`
    
    // Events
    OnClick   string `json:"onClick"`
    OnDismiss string `json:"onDismiss"`
    
    // Accessibility
    AriaLabel string `json:"ariaLabel"`
    Role      string `json:"role"`
}
```

### Group Properties
```go
type BadgeGroupProps struct {
    // Identity
    ID string `json:"id"`
    
    // Content
    Title       string      `json:"title"`
    Description string      `json:"description"`
    Badges      []BadgeProps `json:"badges"`
    
    // Layout
    Layout      BadgeLayout `json:"layout"`      // horizontal, vertical, wrap, grid
    Columns     int         `json:"columns"`     // Grid layout columns
    Spacing     string      `json:"spacing"`     // xs, sm, md, lg
    
    // Features
    Dismissible bool `json:"dismissible"`  // Allow removing badges
    AddMore     bool `json:"addMore"`      // Show add button
    Limit       int  `json:"limit"`        // Maximum number of badges
    
    // Appearance
    Size    BadgeSize    `json:"size"`
    Variant BadgeVariant `json:"variant"`
    
    // Events
    OnAdd    string `json:"onAdd"`
    OnRemove string `json:"onRemove"`
    OnChange string `json:"onChange"`
}
```

### Size Variants
```go
type BadgeSize string

const (
    BadgeXS BadgeSize = "xs"    // 12px height, tiny text
    BadgeSM BadgeSize = "sm"    // 16px height, small text
    BadgeMD BadgeSize = "md"    // 20px height, normal text (default)
    BadgeLG BadgeSize = "lg"    // 24px height, larger text
)
```

### Visual Variants
```go
type BadgeVariant string

const (
    BadgePrimary   BadgeVariant = "primary"    // Brand color
    BadgeSecondary BadgeVariant = "secondary"  // Neutral gray
    BadgeSuccess   BadgeVariant = "success"    // Green
    BadgeWarning   BadgeVariant = "warning"    // Yellow/Orange
    BadgeDanger    BadgeVariant = "danger"     // Red
    BadgeInfo      BadgeVariant = "info"       // Blue
    BadgeLight     BadgeVariant = "light"      // Light gray
    BadgeDark      BadgeVariant = "dark"       // Dark gray
)
```

### Position Options
```go
type BadgePosition string

const (
    PositionTopLeft     BadgePosition = "top-left"
    PositionTopRight    BadgePosition = "top-right"
    PositionBottomLeft  BadgePosition = "bottom-left" 
    PositionBottomRight BadgePosition = "bottom-right"
    PositionInline      BadgePosition = "inline"
)
```

### Layout Types
```go
type BadgeLayout string

const (
    LayoutHorizontal BadgeLayout = "horizontal"  // Side by side
    LayoutVertical   BadgeLayout = "vertical"    // Stacked
    LayoutWrap       BadgeLayout = "wrap"        // Wrap to new lines
    LayoutGrid       BadgeLayout = "grid"        // Grid layout
)
```

## 🎨 Styling Implementation

### Base Badge Styles
```css
.badge {
    display: inline-flex;
    align-items: center;
    gap: var(--space-xs);
    padding: 2px 8px;
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-medium);
    line-height: 1;
    border-radius: var(--radius-sm);
    border: 1px solid transparent;
    white-space: nowrap;
    vertical-align: middle;
    transition: var(--transition-base);
    
    /* Base colors */
    background: var(--color-bg-secondary);
    color: var(--color-text-secondary);
    
    &:empty {
        display: none;
    }
}

.badge-text {
    font-size: inherit;
    font-weight: inherit;
    line-height: inherit;
}

.badge-dismiss {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 14px;
    height: 14px;
    margin-left: var(--space-xs);
    background: none;
    border: none;
    color: currentColor;
    cursor: pointer;
    border-radius: var(--radius-xs);
    opacity: 0.7;
    transition: var(--transition-quick);
    
    &:hover {
        opacity: 1;
        background: rgba(255, 255, 255, 0.2);
    }
    
    &:focus {
        outline: none;
        box-shadow: 0 0 0 2px currentColor;
    }
}
```

### Size Variants
```css
/* Extra Small */
.badge-xs {
    padding: 1px 4px;
    font-size: 10px;
    min-height: 12px;
    
    .badge-dismiss {
        width: 10px;
        height: 10px;
        margin-left: 2px;
    }
}

/* Small */
.badge-sm {
    padding: 1px 6px;
    font-size: 11px;
    min-height: 16px;
    
    .badge-dismiss {
        width: 12px;
        height: 12px;
        margin-left: 3px;
    }
}

/* Medium (Default) */
.badge-md {
    padding: 2px 8px;
    font-size: var(--font-size-xs);
    min-height: 20px;
    
    .badge-dismiss {
        width: 14px;
        height: 14px;
        margin-left: var(--space-xs);
    }
}

/* Large */
.badge-lg {
    padding: 4px 12px;
    font-size: var(--font-size-sm);
    min-height: 24px;
    
    .badge-dismiss {
        width: 16px;
        height: 16px;
        margin-left: var(--space-sm);
    }
}
```

### Color Variants
```css
/* Primary */
.badge-primary {
    background: var(--color-primary);
    color: white;
    border-color: var(--color-primary);
    
    &.badge-outline {
        background: transparent;
        color: var(--color-primary);
        border-color: var(--color-primary);
    }
}

/* Secondary */
.badge-secondary {
    background: var(--color-bg-secondary);
    color: var(--color-text-secondary);
    border-color: var(--color-border-medium);
    
    &.badge-outline {
        background: transparent;
        color: var(--color-text-secondary);
        border-color: var(--color-border-medium);
    }
}

/* Success */
.badge-success {
    background: var(--color-success);
    color: white;
    border-color: var(--color-success);
    
    &.badge-outline {
        background: var(--color-success-bg);
        color: var(--color-success);
        border-color: var(--color-success);
    }
}

/* Warning */
.badge-warning {
    background: var(--color-warning);
    color: var(--color-text-primary);
    border-color: var(--color-warning);
    
    &.badge-outline {
        background: var(--color-warning-bg);
        color: var(--color-warning-dark);
        border-color: var(--color-warning);
    }
}

/* Danger */
.badge-danger {
    background: var(--color-danger);
    color: white;
    border-color: var(--color-danger);
    
    &.badge-outline {
        background: var(--color-danger-bg);
        color: var(--color-danger);
        border-color: var(--color-danger);
    }
}

/* Info */
.badge-info {
    background: var(--color-info);
    color: white;
    border-color: var(--color-info);
    
    &.badge-outline {
        background: var(--color-info-bg);
        color: var(--color-info);
        border-color: var(--color-info);
    }
}

/* Light */
.badge-light {
    background: var(--color-bg-surface);
    color: var(--color-text-primary);
    border-color: var(--color-border-light);
    
    &.badge-outline {
        background: transparent;
        color: var(--color-text-primary);
        border-color: var(--color-border-medium);
    }
}

/* Dark */
.badge-dark {
    background: var(--color-text-primary);
    color: var(--color-bg-surface);
    border-color: var(--color-text-primary);
    
    &.badge-outline {
        background: transparent;
        color: var(--color-text-primary);
        border-color: var(--color-text-primary);
    }
}
```

### Status Badge Styles
```css
.status-badge {
    display: inline-flex;
    align-items: center;
    gap: var(--space-xs);
    padding: 2px 8px;
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-medium);
    border-radius: var(--radius-full);
    
    .status-dot {
        width: 6px;
        height: 6px;
        border-radius: 50%;
        flex-shrink: 0;
    }
}

.status-success {
    background: var(--color-success-bg);
    color: var(--color-success-dark);
    
    .status-dot {
        background: var(--color-success);
    }
}

.status-warning {
    background: var(--color-warning-bg);
    color: var(--color-warning-dark);
    
    .status-dot {
        background: var(--color-warning);
    }
}

.status-danger {
    background: var(--color-danger-bg);
    color: var(--color-danger-dark);
    
    .status-dot {
        background: var(--color-danger);
    }
}

.status-info {
    background: var(--color-info-bg);
    color: var(--color-info-dark);
    
    .status-dot {
        background: var(--color-info);
    }
}

.status-neutral {
    background: var(--color-bg-secondary);
    color: var(--color-text-secondary);
    
    .status-dot {
        background: var(--color-border-dark);
    }
}
```

### Interactive Badge Styles
```css
.interactive-badge {
    cursor: pointer;
    transition: var(--transition-base);
    
    &:hover {
        transform: translateY(-1px);
        box-shadow: var(--shadow-sm);
    }
    
    &:focus {
        outline: none;
        box-shadow: 0 0 0 3px var(--color-primary-light);
    }
    
    &:active {
        transform: translateY(0);
    }
    
    .badge-action {
        display: flex;
        align-items: center;
        gap: var(--space-xs);
        background: none;
        border: none;
        color: inherit;
        font: inherit;
        cursor: pointer;
        
        &:focus {
            outline: none;
        }
    }
}

.badge-dismissible {
    .badge-dismiss {
        &:hover {
            background: rgba(255, 255, 255, 0.2);
        }
        
        &:focus {
            outline: none;
            box-shadow: 0 0 0 1px currentColor;
        }
    }
}
```

### Count Badge Styles
```css
.count-badge {
    min-width: 20px;
    height: 20px;
    padding: 0 6px;
    border-radius: var(--radius-full);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 11px;
    font-weight: var(--font-weight-bold);
    line-height: 1;
    
    .count-text {
        font-size: inherit;
        font-weight: inherit;
        line-height: inherit;
    }
    
    /* Single digit optimization */
    &[data-count="1"],
    &[data-count="2"],
    &[data-count="3"],
    &[data-count="4"],
    &[data-count="5"],
    &[data-count="6"],
    &[data-count="7"],
    &[data-count="8"],
    &[data-count="9"] {
        min-width: 20px;
        width: 20px;
        padding: 0;
    }
}
```

### Positioned Badge Styles
```css
.badge-wrapper {
    position: relative;
    display: inline-block;
}

.badge-positioned {
    position: absolute;
    z-index: 1;
    
    &.badge-position-top-right {
        top: -8px;
        right: -8px;
    }
    
    &.badge-position-top-left {
        top: -8px;
        left: -8px;
    }
    
    &.badge-position-bottom-right {
        bottom: -8px;
        right: -8px;
    }
    
    &.badge-position-bottom-left {
        bottom: -8px;
        left: -8px;
    }
    
    &.badge-position-inline {
        position: static;
        display: inline-flex;
    }
}
```

### Badge Group Styles
```css
.badge-group {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
    
    .badge-group-title {
        font-size: var(--font-size-sm);
        font-weight: var(--font-weight-medium);
        color: var(--color-text-secondary);
        margin-bottom: var(--space-xs);
    }
    
    .badge-list {
        display: flex;
        gap: var(--space-sm);
        
        &.badge-layout-horizontal {
            flex-direction: row;
            flex-wrap: nowrap;
            overflow-x: auto;
        }
        
        &.badge-layout-vertical {
            flex-direction: column;
        }
        
        &.badge-layout-wrap {
            flex-direction: row;
            flex-wrap: wrap;
        }
        
        &.badge-layout-grid {
            display: grid;
            grid-template-columns: repeat(var(--columns, 3), 1fr);
        }
    }
    
    .badge-add {
        display: inline-flex;
        align-items: center;
        gap: var(--space-xs);
        cursor: pointer;
        
        &:hover {
            background: var(--color-bg-hover);
        }
    }
}
```

### Animation Styles
```css
/* Pulsing animation */
.badge-pulsing {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse {
    0%, 100% {
        opacity: 1;
    }
    50% {
        opacity: 0.7;
    }
}

/* Rounded corners */
.badge-rounded {
    border-radius: var(--radius-full);
}

/* Outline style */
.badge-outline {
    background: transparent !important;
    border-width: 1px;
    border-style: solid;
}
```

## ⚙️ Advanced Features

### Dynamic Count Badge
```go
templ DynamicCountBadge(props BadgeProps) {
    <div class="dynamic-count-badge" 
         x-data={ fmt.Sprintf(`{
            count: %d,
            max: %d,
            get displayCount() {
                if (this.count === 0 && !%t) return '';
                return this.count > this.max ? this.max + '+' : this.count.toString();
            },
            get shouldShow() {
                return this.count > 0 || %t;
            },
            increment() {
                this.count++;
                $dispatch('count-changed', { count: this.count });
            },
            decrement() {
                if (this.count > 0) {
                    this.count--;
                    $dispatch('count-changed', { count: this.count });
                }
            },
            reset() {
                this.count = 0;
                $dispatch('count-changed', { count: this.count });
            }
        }`, props.Count, props.Max, props.ShowZero, props.ShowZero) }>
        
        <span x-show="shouldShow" 
              x-text="displayCount"
              x-transition
              class={ fmt.Sprintf("badge count-badge badge-%s", props.Variant) }
              :class="{ 'badge-pulsing': count > 0 }"
              role="status"
              :aria-label="`${count} notifications`">
        </span>
    </div>
}
```

### Progress Badge
```go
templ ProgressBadge(props BadgeProps) {
    <div class="progress-badge" 
         x-data={ fmt.Sprintf(`{
            progress: %f,
            get percentage() {
                return Math.round(this.progress * 100);
            },
            get status() {
                if (this.progress < 0.3) return 'danger';
                if (this.progress < 0.7) return 'warning';
                return 'success';
            }
        }`, props.Progress) }>
        
        <span class="badge badge-outline"
              :class="`badge-${status}`"
              role="progressbar"
              :aria-valuenow="percentage"
              aria-valuemin="0"
              aria-valuemax="100"
              :aria-label="`Progress: ${percentage}%`">
            
            <div class="progress-bar">
                <div class="progress-fill" 
                     :style="`width: ${percentage}%`"
                     :class="`bg-${status}`"></div>
            </div>
            
            <span class="progress-text" x-text="`${percentage}%`"></span>
        </span>
    </div>
}
```

### Editable Badge
```go
templ EditableBadge(props BadgeProps) {
    <div class="editable-badge" 
         x-data={ fmt.Sprintf(`{
            editing: false,
            text: '%s',
            originalText: '%s',
            startEdit() {
                this.editing = true;
                this.$nextTick(() => {
                    this.$refs.input.focus();
                    this.$refs.input.select();
                });
            },
            saveEdit() {
                this.editing = false;
                this.originalText = this.text;
                $dispatch('badge-edited', { text: this.text });
            },
            cancelEdit() {
                this.editing = false;
                this.text = this.originalText;
            }
        }`, props.Text, props.Text) }>
        
        <span x-show="!editing" 
              class={ fmt.Sprintf("badge badge-%s badge-editable", props.Variant) }
              @click="startEdit()"
              role="button"
              tabindex="0"
              @keydown.enter="startEdit()"
              :aria-label="`Edit badge: ${text}`">
            
            <span x-text="text"></span>
            
            if props.Editable {
                <span class="edit-icon">
                    @Icon(IconProps{Name: "edit", Size: "xs"})
                </span>
            }
        </span>
        
        <div x-show="editing" 
             x-transition
             class="badge-editor">
            <input 
                type="text"
                x-model="text"
                x-ref="input"
                class="badge-input"
                @keydown.enter="saveEdit()"
                @keydown.escape="cancelEdit()"
                @blur="saveEdit()"
                maxlength="20"
            />
        </div>
    </div>
}
```

### Badge with Tooltip
```go
templ TooltipBadge(props BadgeProps) {
    <div class="tooltip-badge" 
         x-data="{ showTooltip: false }"
         @mouseenter="showTooltip = true"
         @mouseleave="showTooltip = false">
        
        @BasicBadge(props)
        
        if props.Tooltip != "" {
            <div x-show="showTooltip"
                 x-transition
                 class="badge-tooltip"
                 role="tooltip">
                { props.Tooltip }
                <div class="tooltip-arrow"></div>
            </div>
        }
    </div>
}
```

## 📱 Responsive Design

### Mobile Optimizations
```css
/* Mobile-specific badge styling */
@media (max-width: 479px) {
    .badge {
        font-size: var(--font-size-xs);
        padding: 2px 6px;
        
        /* Larger touch targets for interactive badges */
        &.interactive-badge {
            min-height: 32px;
            padding: 4px 8px;
        }
    }
    
    .badge-group {
        .badge-list {
            &.badge-layout-horizontal {
                flex-wrap: wrap;
            }
            
            &.badge-layout-grid {
                grid-template-columns: repeat(2, 1fr);
            }
        }
    }
    
    /* Stack badges vertically on very small screens */
    .badge-positioned {
        position: static;
        display: inline-flex;
        margin: var(--space-xs);
    }
}
```

### Touch Interactions
```css
/* Touch-friendly interactions */
@media (hover: none) {
    .interactive-badge {
        /* Remove hover effects on touch devices */
        &:hover {
            transform: none;
            box-shadow: none;
        }
        
        /* Enhance tap feedback */
        &:active {
            transform: scale(0.95);
        }
    }
    
    .badge-dismiss {
        min-width: 24px;
        min-height: 24px;
        
        &:hover {
            background: none;
        }
    }
}
```

## ♿ Accessibility Implementation

### ARIA Support
```go
func (badge BadgeProps) GetAriaAttributes() map[string]string {
    attrs := make(map[string]string)
    
    // Role based on usage
    if badge.Clickable {
        attrs["role"] = "button"
        attrs["tabindex"] = "0"
    } else if badge.Count > 0 {
        attrs["role"] = "status"
    } else {
        attrs["role"] = "img"
    }
    
    // Accessible name
    if badge.AriaLabel != "" {
        attrs["aria-label"] = badge.AriaLabel
    } else if badge.Count > 0 {
        attrs["aria-label"] = fmt.Sprintf("%d notifications", badge.Count)
    } else if badge.Text != "" {
        attrs["aria-label"] = badge.Text
    }
    
    // Live region for dynamic counts
    if badge.Count > 0 {
        attrs["aria-live"] = "polite"
        attrs["aria-atomic"] = "true"
    }
    
    return attrs
}
```

### Screen Reader Support
```go
templ AccessibleBadge(props BadgeProps) {
    <span class={ props.GetClasses() }
          for attrName, attrValue := range props.GetAriaAttributes() {
              { attrName }={ attrValue }
          }>
        
        if props.Text != "" {
            <span class="badge-text">{ props.Text }</span>
        }
        
        if props.Count > 0 {
            <span class="badge-count">{ formatCount(props.Count) }</span>
        }
        
        if props.Dismissible {
            <button 
                type="button"
                class="badge-dismiss"
                aria-label={ fmt.Sprintf("Remove %s badge", props.Text) }
                @click={ props.OnDismiss }>
                @Icon(IconProps{Name: "x", Size: "xs"})
            </button>
        }
        
        <!-- Screen reader only description -->
        if props.Description != "" {
            <span class="sr-only">{ props.Description }</span>
        }
    </span>
}
```

### Keyboard Navigation
```css
/* Focus management */
.badge {
    &[role="button"] {
        cursor: pointer;
        
        &:focus {
            outline: none;
            box-shadow: 0 0 0 2px var(--color-primary-light);
            border-radius: var(--radius-sm);
        }
        
        &:focus-visible {
            box-shadow: 0 0 0 2px var(--color-primary);
        }
    }
}

.badge-dismiss {
    &:focus {
        outline: none;
        box-shadow: 0 0 0 1px currentColor;
        border-radius: var(--radius-xs);
    }
}

/* High contrast mode support */
@media (prefers-contrast: high) {
    .badge {
        border-width: 2px;
        border-style: solid;
        border-color: currentColor;
    }
    
    .status-dot {
        border: 1px solid currentColor;
    }
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
    .badge-pulsing {
        animation: none;
    }
    
    .interactive-badge {
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
func TestBadgeComponent(t *testing.T) {
    tests := []struct {
        name     string
        props    BadgeProps
        expected []string
    }{
        {
            name: "basic badge",
            props: BadgeProps{
                Text:    "New",
                Variant: "primary",
            },
            expected: []string{"badge", "badge-primary", "New"},
        },
        {
            name: "count badge",
            props: BadgeProps{
                Count:   5,
                Variant: "danger",
            },
            expected: []string{"badge", "count-badge", "5"},
        },
        {
            name: "dismissible badge",
            props: BadgeProps{
                Text:        "Tag",
                Dismissible: true,
            },
            expected: []string{"badge-dismiss", "aria-label"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            component := renderBadge(tt.props)
            for _, expected := range tt.expected {
                assert.Contains(t, component, expected)
            }
        })
    }
}
```

### Accessibility Tests
```javascript
describe('Badge Accessibility', () => {
    test('has proper role for count badges', () => {
        const badge = render(<Badge count={5} />);
        const badgeElement = screen.getByRole('status');
        expect(badgeElement).toBeInTheDocument();
        expect(badgeElement).toHaveAttribute('aria-label', '5 notifications');
    });
    
    test('clickable badges have button role', () => {
        const badge = render(<Badge text="Filter" clickable onClick={() => {}} />);
        const badgeElement = screen.getByRole('button');
        expect(badgeElement).toBeInTheDocument();
    });
    
    test('supports keyboard interaction', () => {
        const handleClick = jest.fn();
        const badge = render(<Badge text="Tag" clickable onClick={handleClick} />);
        const badgeElement = screen.getByRole('button');
        
        badgeElement.focus();
        expect(badgeElement).toHaveFocus();
        
        fireEvent.keyDown(badgeElement, { key: 'Enter' });
        expect(handleClick).toHaveBeenCalled();
    });
});
```

### Visual Regression Tests
```javascript
test.describe('Badge Visual Tests', () => {
    test('all badge variants', async ({ page }) => {
        await page.goto('/components/badge');
        
        // Test all sizes
        for (const size of ['xs', 'sm', 'md', 'lg']) {
            await expect(page.locator(`[data-size="${size}"]`)).toHaveScreenshot(`badge-${size}.png`);
        }
        
        // Test all variants
        const variants = ['primary', 'secondary', 'success', 'warning', 'danger', 'info'];
        for (const variant of variants) {
            await expect(page.locator(`[data-variant="${variant}"]`)).toHaveScreenshot(`badge-${variant}.png`);
        }
        
        // Test special states
        await expect(page.locator('[data-state="outline"]')).toHaveScreenshot('badge-outline.png');
        await expect(page.locator('[data-state="dismissible"]')).toHaveScreenshot('badge-dismissible.png');
    });
});
```

## 📚 Usage Examples

### Notification Badge
```go
templ NotificationBell() {
    @BadgeWrapper(
        @Icon(IconProps{Name: "bell", Size: "lg"}),
        BadgeProps{
            Count:    3,
            Variant:  "danger",
            Position: "top-right",
            Max:      99,
        },
    )
}
```

### Status Indicators
```go
templ UserStatus() {
    <div class="user-info">
        <span class="user-name">John Doe</span>
        @StatusBadge(BadgeProps{
            Text:    "Online",
            Status:  "success",
            WithDot: true,
        })
    </div>
}

templ OrderStatus() {
    @StatusBadge(BadgeProps{
        Text:   "Processing",
        Status: "warning",
    })
}
```

### Filter Tags
```go
templ FilterTags() {
    @BadgeGroup(BadgeGroupProps{
        Title:       "Active Filters",
        Dismissible: true,
        AddMore:     true,
        Badges: []BadgeProps{
            {Text: "Category: Electronics", Variant: "secondary"},
            {Text: "Price: $100-500", Variant: "secondary"},
            {Text: "Brand: Apple", Variant: "secondary"},
        },
    })
}
```

### Progress Indicator
```go
templ TaskProgress() {
    <div class="task-item">
        <span class="task-name">Project Setup</span>
        @ProgressBadge(BadgeProps{
            Progress: 0.75,
            Text:     "75%",
        })
    </div>
}
```

## 🔗 Related Components

- **[Tag](../tag/)**: Categorization and metadata labels
- **[Status](../status/)**: State indicators with icons
- **[Button](../button/)**: Interactive action triggers
- **[Tooltip](../../molecules/tooltip/)**: Enhanced badge descriptions

---

**Component Status**: ✅ Production Ready  
**Schema Reference**: `BadgeObject.json`  
**CSS Classes**: `.badge`, `.badge-{size}`, `.badge-{variant}`  
**Accessibility**: WCAG 2.1 AA Compliant