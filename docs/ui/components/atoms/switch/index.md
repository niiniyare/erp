# Switch Component

**FILE PURPOSE**: Toggle control implementation and specifications  
**SCOPE**: All switch variants, states, and interaction patterns  
**TARGET AUDIENCE**: Developers implementing toggle controls and settings interfaces

## 📋 Component Overview

The Switch component provides intuitive toggle controls for binary settings and preferences. It offers immediate visual feedback, smooth animations, and supports various sizes and styles while maintaining full accessibility compliance.

### Schema Reference
- **Primary Schema**: `SwitchControlSchema.json`
- **Related Schemas**: `StateSchema.json`, `Options.json`
- **Base Interface**: Form control with toggle state

## 🎨 Switch Types

### Basic Switch
**Purpose**: Simple on/off toggle for boolean settings

```go
// Basic switch configuration
basicSwitch := SwitchProps{
    Name:     "notifications",
    Label:    "Enable Notifications",
    Checked:  true,
    Size:     "md",
    Variant:  "primary",
    OnChange: "handleNotificationToggle",
}

// Generated Templ component
templ BasicSwitch(props SwitchProps) {
    <div class="switch-group">
        <label class="switch-label" for={ props.ID }>
            <input 
                type="checkbox"
                id={ props.ID }
                name={ props.Name }
                class="switch-input sr-only"
                checked?={ props.Checked }
                disabled?={ props.Disabled }
                @change={ props.OnChange }
                value={ props.Value }
            />
            
            <div class={ fmt.Sprintf("switch-track switch-%s switch-%s", props.Size, props.Variant) }
                 :class="{ 'checked': $el.previousElementSibling.checked }">
                <div class="switch-thumb"></div>
                if props.ShowIcons {
                    <div class="switch-icons">
                        <span class="icon-off">
                            @Icon(IconProps{Name: props.OffIcon, Size: "xs"})
                        </span>
                        <span class="icon-on">
                            @Icon(IconProps{Name: props.OnIcon, Size: "xs"})
                        </span>
                    </div>
                }
            </div>
            
            if props.Label != "" {
                <span class="switch-text">{ props.Label }</span>
            }
        </label>
        
        if props.Description != "" {
            <div class="switch-description">{ props.Description }</div>
        }
        
        if props.Error != "" {
            <div class="error-message">{ props.Error }</div>
        }
    </div>
}
```

### Switch with Labels
**Purpose**: Toggle with on/off text labels for clarity

```go
labeledSwitch := SwitchProps{
    Name:     "autoSave",
    Label:    "Auto Save",
    OnLabel:  "Enabled",
    OffLabel: "Disabled",
    ShowLabels: true,
    Size:     "lg",
}

templ LabeledSwitch(props SwitchProps) {
    <div class="switch-group labeled-switch" x-data="{ checked: false }">
        <div class="switch-container">
            if props.Label != "" {
                <span class="switch-main-label">{ props.Label }</span>
            }
            
            <label class="switch-label">
                <input 
                    type="checkbox"
                    x-model="checked"
                    name={ props.Name }
                    class="switch-input sr-only"
                    disabled?={ props.Disabled }
                />
                
                <div class="switch-track"
                     :class="{ 'checked': checked, 'disabled': props.Disabled }">
                    
                    if props.ShowLabels {
                        <span class="switch-label-off" 
                              :class="{ 'active': !checked }">
                            { props.OffLabel }
                        </span>
                        <span class="switch-label-on" 
                              :class="{ 'active': checked }">
                            { props.OnLabel }
                        </span>
                    }
                    
                    <div class="switch-thumb"></div>
                </div>
            </label>
            
            <div class="switch-status" x-text="checked ? '{ props.OnLabel }' : '{ props.OffLabel }'"></div>
        </div>
    </div>
}
```

### Switch Group
**Purpose**: Multiple related toggle settings organized together

```go
switchGroup := SwitchGroupProps{
    Label: "Notification Settings",
    Switches: []SwitchOption{
        {
            Name:        "emailNotifications",
            Label:       "Email Notifications",
            Description: "Receive updates via email",
            Checked:     true,
            Icon:        "mail",
        },
        {
            Name:        "pushNotifications", 
            Label:       "Push Notifications",
            Description: "Browser push notifications",
            Checked:     false,
            Icon:        "bell",
        },
        {
            Name:        "smsNotifications",
            Label:       "SMS Notifications",
            Description: "Text message alerts",
            Checked:     false,
            Icon:        "phone",
            Premium:     true,
        },
    },
}

templ SwitchGroup(props SwitchGroupProps) {
    <div class="switch-group-container">
        <h3 class="group-title">{ props.Label }</h3>
        
        if props.Description != "" {
            <p class="group-description">{ props.Description }</p>
        }
        
        <div class="switch-list">
            for _, switchOpt := range props.Switches {
                <div class="switch-item" :class="{ 'premium': { fmt.Sprintf("%t", switchOpt.Premium) } }">
                    <div class="switch-content">
                        <div class="switch-header">
                            if switchOpt.Icon != "" {
                                <div class="switch-icon">
                                    @Icon(IconProps{Name: switchOpt.Icon, Size: "md"})
                                </div>
                            }
                            <div class="switch-info">
                                <span class="switch-title">{ switchOpt.Label }</span>
                                if switchOpt.Premium {
                                    @Badge(BadgeProps{Text: "Premium", Variant: "warning", Size: "xs"})
                                }
                            </div>
                        </div>
                        
                        if switchOpt.Description != "" {
                            <p class="switch-description">{ switchOpt.Description }</p>
                        }
                    </div>
                    
                    @BasicSwitch(SwitchProps{
                        Name:     switchOpt.Name,
                        Checked:  switchOpt.Checked,
                        Disabled: switchOpt.Premium && !props.HasPremium,
                        Size:     "md",
                        Variant:  "primary",
                    })
                </div>
            }
        </div>
    </div>
}
```

### Animated Switch
**Purpose**: Enhanced visual feedback with smooth transitions

```go
animatedSwitch := SwitchProps{
    Name:      "darkMode",
    Label:     "Dark Mode",
    OnIcon:    "moon",
    OffIcon:   "sun",
    ShowIcons: true,
    Animated:  true,
    Size:      "lg",
}

templ AnimatedSwitch(props SwitchProps) {
    <div class="animated-switch-group" 
         x-data={ fmt.Sprintf(`{
            checked: %t,
            toggle() {
                this.checked = !this.checked;
                $dispatch('switch-change', { 
                    name: '%s', 
                    checked: this.checked 
                });
            }
        }`, props.Checked, props.Name) }>
        
        <label class="animated-switch-label">
            <input 
                type="checkbox"
                x-model="checked"
                @change="toggle()"
                class="switch-input sr-only"
                name={ props.Name }
            />
            
            <div class="animated-switch-track" 
                 :class="{ 
                     'checked': checked,
                     'animated': true,
                     [`size-${props.Size}`]: true,
                     [`variant-${props.Variant}`]: true
                 }">
                
                <div class="switch-background"></div>
                
                <div class="switch-thumb-container">
                    <div class="switch-thumb">
                        if props.ShowIcons {
                            <div class="thumb-icon" 
                                 :class="{ 'visible': !checked }">
                                @Icon(IconProps{Name: props.OffIcon, Size: "xs"})
                            </div>
                            <div class="thumb-icon" 
                                 :class="{ 'visible': checked }">
                                @Icon(IconProps{Name: props.OnIcon, Size: "xs"})
                            </div>
                        }
                    </div>
                </div>
                
                if props.ShowLabels {
                    <div class="track-labels">
                        <span class="label-off" :class="{ 'visible': !checked }">
                            { props.OffLabel }
                        </span>
                        <span class="label-on" :class="{ 'visible': checked }">
                            { props.OnLabel }
                        </span>
                    </div>
                }
            </div>
            
            <span class="switch-text">{ props.Label }</span>
        </label>
    </div>
}
```

## 🎯 Props Interface

### Core Properties
```go
type SwitchProps struct {
    // Identity
    Name     string `json:"name"`
    ID       string `json:"id"`
    TestID   string `json:"testid"`
    
    // Content
    Label       string `json:"label"`
    Value       string `json:"value"`
    Description string `json:"description"`
    
    // Labels
    OnLabel     string `json:"onLabel"`     // "On", "Enabled", etc.
    OffLabel    string `json:"offLabel"`    // "Off", "Disabled", etc.
    ShowLabels  bool   `json:"showLabels"`  // Display labels on track
    
    // Icons
    OnIcon      string `json:"onIcon"`      // Icon for on state
    OffIcon     string `json:"offIcon"`     // Icon for off state
    ShowIcons   bool   `json:"showIcons"`   // Display icons
    
    // State
    Checked       bool   `json:"checked"`
    DefaultChecked bool  `json:"defaultChecked"`
    Required      bool   `json:"required"`
    Disabled      bool   `json:"disabled"`
    ReadOnly      bool   `json:"readonly"`
    Loading       bool   `json:"loading"`    // Show loading state
    
    // Appearance
    Size        SwitchSize    `json:"size"`        // xs, sm, md, lg, xl
    Variant     SwitchVariant `json:"variant"`     // default, primary, success, warning, danger
    Color       string        `json:"color"`       // Custom color
    Class       string        `json:"className"`
    Style       map[string]string `json:"style"`
    
    // Animation
    Animated    bool   `json:"animated"`    // Enhanced animations
    Duration    int    `json:"duration"`    // Animation duration in ms
    
    // Events
    OnChange    string `json:"onChange"`
    OnFocus     string `json:"onFocus"`
    OnBlur      string `json:"onBlur"`
    
    // Error handling
    Error string `json:"error"`
    
    // Accessibility
    AriaLabel       string `json:"ariaLabel"`
    AriaDescribedBy string `json:"ariaDescribedBy"`
    TabIndex        int    `json:"tabIndex"`
}
```

### Group Properties
```go
type SwitchGroupProps struct {
    // Identity
    Name string `json:"name"`
    ID   string `json:"id"`
    
    // Content
    Label       string         `json:"label"`
    Description string         `json:"description"`
    Switches    []SwitchOption `json:"switches"`
    
    // Features
    HasPremium   bool   `json:"hasPremium"`   // User has premium access
    AllowBulk    bool   `json:"allowBulk"`    // Enable bulk toggle
    
    // Layout
    Layout    SwitchLayout `json:"layout"`    // vertical, horizontal, grid
    Columns   int          `json:"columns"`   // Grid layout columns
    
    // Appearance
    Size      SwitchSize    `json:"size"`
    Variant   SwitchVariant `json:"variant"`
    Compact   bool          `json:"compact"`  // Compact spacing
    
    // Events
    OnChange string `json:"onChange"`
}

type SwitchOption struct {
    Name        string `json:"name"`
    Label       string `json:"label"`
    Description string `json:"description"`
    Icon        string `json:"icon"`
    Checked     bool   `json:"checked"`
    Disabled    bool   `json:"disabled"`
    Premium     bool   `json:"premium"`
    Badge       string `json:"badge"`
}
```

### Size Variants
```go
type SwitchSize string

const (
    SwitchXS SwitchSize = "xs"    // 16px width, 10px height
    SwitchSM SwitchSize = "sm"    // 24px width, 14px height
    SwitchMD SwitchSize = "md"    // 32px width, 18px height (default)
    SwitchLG SwitchSize = "lg"    // 40px width, 22px height
    SwitchXL SwitchSize = "xl"    // 48px width, 26px height
)
```

### Layout Types
```go
type SwitchLayout string

const (
    SwitchVertical   SwitchLayout = "vertical"   // Stacked vertically
    SwitchHorizontal SwitchLayout = "horizontal" // Side by side
    SwitchGrid       SwitchLayout = "grid"       // Grid layout
)
```

### Visual Variants
```go
type SwitchVariant string

const (
    SwitchDefault SwitchVariant = "default"  // Gray theme
    SwitchPrimary SwitchVariant = "primary"  // Brand color
    SwitchSuccess SwitchVariant = "success"  // Green theme
    SwitchWarning SwitchVariant = "warning"  // Yellow theme
    SwitchDanger  SwitchVariant = "danger"   // Red theme
)
```

## 🎨 Styling Implementation

### Base Switch Styles
```css
.switch-group {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
}

.switch-label {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    cursor: pointer;
    user-select: none;
    
    &:hover .switch-track {
        background: var(--color-bg-hover);
    }
}

.switch-input {
    /* Visually hidden but accessible */
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
    
    /* Focus styles */
    &:focus + .switch-track {
        outline: none;
        box-shadow: 0 0 0 3px var(--color-primary-light);
    }
    
    /* Checked state */
    &:checked + .switch-track {
        background: var(--color-primary);
        border-color: var(--color-primary);
        
        .switch-thumb {
            transform: translateX(calc(100% + 2px));
        }
        
        .icon-on {
            opacity: 1;
        }
        
        .icon-off {
            opacity: 0;
        }
    }
    
    /* Disabled state */
    &:disabled + .switch-track {
        opacity: 0.5;
        cursor: not-allowed;
        
        .switch-thumb {
            background: var(--color-bg-disabled);
        }
    }
}

.switch-track {
    position: relative;
    display: flex;
    align-items: center;
    width: 32px;
    height: 18px;
    padding: 2px;
    background: var(--color-bg-secondary);
    border: 1px solid var(--color-border-medium);
    border-radius: 12px;
    transition: var(--transition-base);
    cursor: pointer;
    
    &.checked {
        background: var(--color-primary);
        border-color: var(--color-primary);
    }
}

.switch-thumb {
    position: relative;
    width: 14px;
    height: 14px;
    background: var(--color-bg-surface);
    border-radius: 50%;
    box-shadow: var(--shadow-sm);
    transition: var(--transition-base);
    z-index: 2;
    
    /* Checked state transform applied via input:checked */
}

.switch-text {
    color: var(--color-text-primary);
    font-size: var(--font-size-base);
    font-weight: var(--font-weight-medium);
}

.switch-description {
    color: var(--color-text-secondary);
    font-size: var(--font-size-sm);
    line-height: 1.4;
    margin-top: var(--space-xs);
}

.switch-icons {
    position: absolute;
    top: 50%;
    left: 0;
    right: 0;
    transform: translateY(-50%);
    display: flex;
    justify-content: space-between;
    padding: 0 4px;
    pointer-events: none;
    
    .icon-off,
    .icon-on {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 10px;
        height: 10px;
        color: var(--color-text-tertiary);
        transition: var(--transition-quick);
        opacity: 0.7;
    }
    
    .icon-on {
        opacity: 0;
        color: white;
    }
}
```

### Size Variants
```css
/* Extra Small */
.switch-xs {
    width: 16px;
    height: 10px;
    border-radius: 6px;
    
    .switch-thumb {
        width: 8px;
        height: 8px;
    }
    
    &.checked .switch-thumb {
        transform: translateX(8px);
    }
}

/* Small */
.switch-sm {
    width: 24px;
    height: 14px;
    border-radius: 8px;
    
    .switch-thumb {
        width: 12px;
        height: 12px;
    }
    
    &.checked .switch-thumb {
        transform: translateX(12px);
    }
}

/* Medium (Default) */
.switch-md {
    width: 32px;
    height: 18px;
    border-radius: 12px;
    
    .switch-thumb {
        width: 14px;
        height: 14px;
    }
    
    &.checked .switch-thumb {
        transform: translateX(16px);
    }
}

/* Large */
.switch-lg {
    width: 40px;
    height: 22px;
    border-radius: 14px;
    
    .switch-thumb {
        width: 18px;
        height: 18px;
    }
    
    &.checked .switch-thumb {
        transform: translateX(20px);
    }
}

/* Extra Large */
.switch-xl {
    width: 48px;
    height: 26px;
    border-radius: 16px;
    
    .switch-thumb {
        width: 22px;
        height: 22px;
    }
    
    &.checked .switch-thumb {
        transform: translateX(24px);
    }
}
```

### Color Variants
```css
/* Primary variant */
.switch-primary {
    &.checked {
        background: var(--color-primary);
        border-color: var(--color-primary);
    }
}

/* Success variant */
.switch-success {
    &.checked {
        background: var(--color-success);
        border-color: var(--color-success);
    }
}

/* Warning variant */
.switch-warning {
    &.checked {
        background: var(--color-warning);
        border-color: var(--color-warning);
    }
}

/* Danger variant */
.switch-danger {
    &.checked {
        background: var(--color-danger);
        border-color: var(--color-danger);
    }
}
```

### Labeled Switch Styles
```css
.labeled-switch {
    .switch-container {
        display: flex;
        align-items: center;
        gap: var(--space-md);
    }
    
    .switch-main-label {
        font-weight: var(--font-weight-medium);
        color: var(--color-text-primary);
    }
    
    .switch-status {
        font-size: var(--font-size-sm);
        color: var(--color-text-secondary);
        min-width: 60px;
        text-align: right;
    }
    
    .switch-track {
        width: 48px;
        height: 24px;
        padding: 2px;
        
        .switch-label-off,
        .switch-label-on {
            position: absolute;
            font-size: 10px;
            font-weight: var(--font-weight-semibold);
            color: var(--color-text-tertiary);
            transition: var(--transition-quick);
            opacity: 0.7;
            
            &.active {
                opacity: 1;
            }
        }
        
        .switch-label-off {
            right: 4px;
            color: var(--color-text-secondary);
        }
        
        .switch-label-on {
            left: 4px;
            color: white;
            opacity: 0;
        }
        
        &.checked {
            .switch-label-on {
                opacity: 1;
            }
            
            .switch-label-off {
                opacity: 0;
            }
        }
        
        .switch-thumb {
            width: 20px;
            height: 20px;
        }
        
        &.checked .switch-thumb {
            transform: translateX(26px);
        }
    }
}
```

### Group Layout Styles
```css
.switch-group-container {
    .group-title {
        font-size: var(--font-size-lg);
        font-weight: var(--font-weight-semibold);
        color: var(--color-text-primary);
        margin-bottom: var(--space-sm);
    }
    
    .group-description {
        color: var(--color-text-secondary);
        font-size: var(--font-size-sm);
        line-height: 1.4;
        margin-bottom: var(--space-lg);
    }
    
    .switch-list {
        display: flex;
        flex-direction: column;
        gap: var(--space-lg);
        
        &.layout-horizontal {
            flex-direction: row;
            flex-wrap: wrap;
        }
        
        &.layout-grid {
            display: grid;
            grid-template-columns: repeat(var(--columns, 2), 1fr);
        }
        
        &.compact {
            gap: var(--space-md);
        }
    }
    
    .switch-item {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: var(--space-md);
        border: 1px solid var(--color-border-light);
        border-radius: var(--radius-md);
        background: var(--color-bg-surface);
        transition: var(--transition-base);
        
        &:hover {
            border-color: var(--color-border-medium);
            background: var(--color-bg-hover);
        }
        
        &.premium {
            position: relative;
            
            &::before {
                content: '';
                position: absolute;
                top: 0;
                right: 0;
                width: 0;
                height: 0;
                border-style: solid;
                border-width: 0 16px 16px 0;
                border-color: transparent var(--color-warning) transparent transparent;
            }
        }
        
        .switch-content {
            flex: 1;
            
            .switch-header {
                display: flex;
                align-items: center;
                gap: var(--space-sm);
                margin-bottom: var(--space-xs);
            }
            
            .switch-icon {
                width: 24px;
                height: 24px;
                display: flex;
                align-items: center;
                justify-content: center;
                color: var(--color-text-secondary);
            }
            
            .switch-info {
                display: flex;
                align-items: center;
                gap: var(--space-sm);
            }
            
            .switch-title {
                font-weight: var(--font-weight-medium);
                color: var(--color-text-primary);
            }
            
            .switch-description {
                color: var(--color-text-secondary);
                font-size: var(--font-size-sm);
                line-height: 1.4;
                margin: 0;
            }
        }
    }
}
```

### Animated Switch Styles
```css
.animated-switch-group {
    .animated-switch-track {
        position: relative;
        width: 48px;
        height: 26px;
        border-radius: 16px;
        background: var(--color-bg-secondary);
        border: 2px solid var(--color-border-medium);
        cursor: pointer;
        overflow: hidden;
        transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        
        &.animated {
            .switch-thumb-container {
                transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            }
            
            .switch-background {
                transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            }
        }
        
        &.checked {
            background: var(--color-primary);
            border-color: var(--color-primary);
            
            .switch-thumb-container {
                transform: translateX(22px);
            }
            
            .switch-background {
                transform: scale(1.1);
                opacity: 0.2;
            }
        }
        
        .switch-background {
            position: absolute;
            top: -2px;
            left: -2px;
            right: -2px;
            bottom: -2px;
            background: var(--color-primary);
            border-radius: 16px;
            transform: scale(0);
            opacity: 0;
        }
        
        .switch-thumb-container {
            position: absolute;
            top: 2px;
            left: 2px;
            width: 22px;
            height: 22px;
        }
        
        .switch-thumb {
            width: 100%;
            height: 100%;
            background: white;
            border-radius: 50%;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
            display: flex;
            align-items: center;
            justify-content: center;
            position: relative;
            
            .thumb-icon {
                position: absolute;
                opacity: 0;
                transform: scale(0.8);
                transition: all 0.2s ease-in-out;
                
                &.visible {
                    opacity: 1;
                    transform: scale(1);
                }
            }
        }
        
        .track-labels {
            position: absolute;
            top: 50%;
            left: 0;
            right: 0;
            transform: translateY(-50%);
            display: flex;
            justify-content: space-between;
            padding: 0 6px;
            pointer-events: none;
            
            .label-off,
            .label-on {
                font-size: 9px;
                font-weight: var(--font-weight-bold);
                text-transform: uppercase;
                opacity: 0;
                transition: opacity 0.2s ease-in-out;
                
                &.visible {
                    opacity: 0.8;
                }
            }
            
            .label-off {
                color: var(--color-text-secondary);
            }
            
            .label-on {
                color: white;
            }
        }
    }
}
```

## ⚙️ Advanced Features

### Loading State Switch
```go
templ LoadingSwitch(props SwitchProps) {
    <div class="loading-switch" 
         x-data={ fmt.Sprintf(`{
            loading: false,
            checked: %t,
            async toggle() {
                this.loading = true;
                try {
                    const response = await fetch('/api/settings/%s', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ checked: !this.checked })
                    });
                    if (response.ok) {
                        this.checked = !this.checked;
                    }
                } finally {
                    this.loading = false;
                }
            }
        }`, props.Checked, props.Name) }>
        
        <label class="switch-label">
            <div class="switch-track" 
                 :class="{ 
                     'checked': checked, 
                     'loading': loading,
                     'disabled': loading 
                 }"
                 @click="toggle()">
                
                <div class="switch-thumb">
                    <div x-show="!loading" class="thumb-normal"></div>
                    <div x-show="loading" class="thumb-loading">
                        @Spinner(SpinnerProps{Size: "xs"})
                    </div>
                </div>
            </div>
            
            <span class="switch-text">{ props.Label }</span>
        </label>
        
        <div x-show="loading" class="loading-message">
            Updating setting...
        </div>
    </div>
}
```

### Confirmation Switch
```go
templ ConfirmationSwitch(props SwitchProps) {
    <div class="confirmation-switch" 
         x-data={ fmt.Sprintf(`{
            checked: %t,
            showConfirm: false,
            pendingState: null,
            requestToggle() {
                this.pendingState = !this.checked;
                this.showConfirm = true;
            },
            confirm() {
                this.checked = this.pendingState;
                this.showConfirm = false;
                this.pendingState = null;
                $dispatch('switch-confirmed', { name: '%s', checked: this.checked });
            },
            cancel() {
                this.showConfirm = false;
                this.pendingState = null;
            }
        }`, props.Checked, props.Name) }>
        
        <div class="switch-container">
            <label class="switch-label">
                <div class="switch-track" 
                     :class="{ 'checked': checked }"
                     @click="requestToggle()">
                    <div class="switch-thumb"></div>
                </div>
                <span class="switch-text">{ props.Label }</span>
            </label>
            
            if props.Description != "" {
                <p class="switch-description">{ props.Description }</p>
            }
        </div>
        
        <div x-show="showConfirm" 
             x-transition
             class="confirmation-modal">
            <div class="modal-backdrop"></div>
            <div class="modal-content">
                <h4 class="modal-title">{ props.ConfirmTitle }</h4>
                <p class="modal-message">{ props.ConfirmMessage }</p>
                <div class="modal-actions">
                    <button @click="cancel()" class="btn btn-secondary">
                        Cancel
                    </button>
                    <button @click="confirm()" class="btn btn-primary">
                        Confirm
                    </button>
                </div>
            </div>
        </div>
    </div>
}
```

### Switch with Dependency
```go
templ DependentSwitch(props SwitchProps) {
    <div class="dependent-switch"
         x-data={ fmt.Sprintf(`{
            checked: %t,
            parentChecked: %t,
            get isEnabled() {
                return this.parentChecked;
            },
            toggle() {
                if (this.isEnabled) {
                    this.checked = !this.checked;
                }
            }
        }`, props.Checked, props.ParentChecked) }>
        
        <div class="dependency-info" x-show="!parentChecked">
            <div class="info-icon">
                @Icon(IconProps{Name: "info-circle", Size: "sm"})
            </div>
            <span class="info-text">
                { props.DependencyMessage }
            </span>
        </div>
        
        <label class="switch-label" 
               :class="{ 'disabled': !isEnabled }">
            <div class="switch-track" 
                 :class="{ 
                     'checked': checked && isEnabled,
                     'disabled': !isEnabled 
                 }"
                 @click="toggle()">
                <div class="switch-thumb"></div>
            </div>
            <span class="switch-text">{ props.Label }</span>
        </label>
    </div>
}
```

## 📱 Responsive Design

### Mobile Optimizations
```css
/* Mobile-specific switch styling */
@media (max-width: 479px) {
    .switch-label {
        min-height: 44px;  /* Touch target */
        align-items: center;
    }
    
    /* Larger switches on mobile */
    .switch-track {
        min-width: 44px;
        min-height: 24px;
    }
    
    .switch-group .switch-item {
        padding: var(--space-lg);
        
        .switch-content {
            .switch-title {
                font-size: var(--font-size-base);
            }
            
            .switch-description {
                font-size: var(--font-size-sm);
            }
        }
    }
    
    /* Stack group items vertically on mobile */
    .switch-list.layout-horizontal {
        flex-direction: column;
    }
    
    .switch-list.layout-grid {
        grid-template-columns: 1fr;
    }
}
```

### Touch Interactions
```css
/* Touch-friendly interactions */
@media (hover: none) {
    .switch-label {
        /* Remove hover effects on touch devices */
        &:hover .switch-track {
            background: var(--color-bg-secondary);
        }
    }
    
    .switch-track {
        /* Larger touch targets */
        min-width: 44px;
        min-height: 24px;
        
        /* Enhanced feedback for touch */
        &:active {
            transform: scale(0.95);
        }
    }
    
    .switch-item {
        &:hover {
            border-color: var(--color-border-light);
            background: var(--color-bg-surface);
        }
    }
}
```

## ♿ Accessibility Implementation

### ARIA Support
```go
func (sw SwitchProps) GetAriaAttributes() map[string]string {
    attrs := make(map[string]string)
    
    // Switch role
    attrs["role"] = "switch"
    attrs["aria-checked"] = fmt.Sprintf("%t", sw.Checked)
    
    // Required state
    if sw.Required {
        attrs["aria-required"] = "true"
    }
    
    // Invalid state
    if sw.Error != "" {
        attrs["aria-invalid"] = "true"
        if sw.ID != "" {
            attrs["aria-describedby"] = sw.ID + "-error"
        }
    }
    
    // Help text association
    if sw.Description != "" && sw.ID != "" {
        describedBy := attrs["aria-describedby"]
        if describedBy != "" {
            attrs["aria-describedby"] = describedBy + " " + sw.ID + "-desc"
        } else {
            attrs["aria-describedby"] = sw.ID + "-desc"
        }
    }
    
    // Custom label
    if sw.AriaLabel != "" {
        attrs["aria-label"] = sw.AriaLabel
    }
    
    // Loading state
    if sw.Loading {
        attrs["aria-busy"] = "true"
    }
    
    return attrs
}
```

### Screen Reader Support
```go
templ AccessibleSwitch(props SwitchProps) {
    <div class="switch-group">
        <label class="switch-label" for={ props.ID }>
            <input 
                type="checkbox"
                id={ props.ID }
                name={ props.Name }
                class="switch-input sr-only"
                checked?={ props.Checked }
                disabled?={ props.Disabled }
                for attrName, attrValue := range props.GetAriaAttributes() {
                    { attrName }={ attrValue }
                }
            />
            
            <div class="switch-track" aria-hidden="true">
                <div class="switch-thumb"></div>
            </div>
            
            if props.Label != "" {
                <span class="switch-text">{ props.Label }</span>
            }
        </label>
        
        if props.Description != "" {
            <div id={ props.ID + "-desc" } class="switch-description">
                { props.Description }
            </div>
        }
        
        if props.Error != "" {
            <div 
                id={ props.ID + "-error" }
                class="error-message"
                role="alert"
                aria-live="polite">
                { props.Error }
            </div>
        }
        
        <!-- Live region for state changes -->
        <div 
            class="sr-only" 
            aria-live="polite" 
            aria-atomic="true"
            x-text={ fmt.Sprintf("'%s is ' + (checked ? 'enabled' : 'disabled')", props.Label) }>
        </div>
    </div>
}
```

### Keyboard Navigation
```css
/* Focus management */
.switch-input:focus + .switch-track {
    outline: none;
    box-shadow: 0 0 0 3px var(--color-primary-light);
    border-color: var(--color-primary);
}

/* High contrast mode support */
@media (prefers-contrast: high) {
    .switch-track {
        border-width: 2px;
        border-color: CanvasText;
    }
    
    .switch-thumb {
        background: CanvasText;
    }
    
    .switch-input:checked + .switch-track {
        background: Highlight;
        border-color: Highlight;
        
        .switch-thumb {
            background: HighlightText;
        }
    }
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
    .switch-track,
    .switch-thumb,
    .switch-icons > * {
        transition: none;
    }
    
    .animated-switch-track,
    .animated-switch-track * {
        transition: none;
    }
}
```

## 🧪 Testing Guidelines

### Unit Tests
```go
func TestSwitchComponent(t *testing.T) {
    tests := []struct {
        name     string
        props    SwitchProps
        expected []string
    }{
        {
            name: "basic switch",
            props: SwitchProps{
                Name:  "test",
                Label: "Test Switch",
            },
            expected: []string{"switch-group", "switch-track", "Test Switch"},
        },
        {
            name: "checked switch",
            props: SwitchProps{
                Name:    "checked",
                Checked: true,
            },
            expected: []string{"checked", "aria-checked=\"true\""},
        },
        {
            name: "disabled switch",
            props: SwitchProps{
                Name:     "disabled",
                Disabled: true,
            },
            expected: []string{"disabled"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            component := renderSwitch(tt.props)
            for _, expected := range tt.expected {
                assert.Contains(t, component, expected)
            }
        })
    }
}
```

### Accessibility Tests
```javascript
describe('Switch Accessibility', () => {
    test('has proper switch role', () => {
        const switchComp = render(<Switch name="test" label="Test Switch" />);
        const switchElement = screen.getByRole('switch');
        expect(switchElement).toBeInTheDocument();
    });
    
    test('announces state changes', () => {
        const switchComp = render(<Switch name="test" label="Test Switch" />);
        const switchElement = screen.getByRole('switch');
        
        fireEvent.click(switchElement);
        expect(switchElement).toHaveAttribute('aria-checked', 'true');
    });
    
    test('supports keyboard interaction', () => {
        const switchComp = render(<Switch name="test" label="Test Switch" />);
        const switchElement = screen.getByRole('switch');
        
        switchElement.focus();
        expect(switchElement).toHaveFocus();
        
        fireEvent.keyDown(switchElement, { key: ' ' });
        expect(switchElement).toHaveAttribute('aria-checked', 'true');
        
        fireEvent.keyDown(switchElement, { key: 'Enter' });
        expect(switchElement).toHaveAttribute('aria-checked', 'false');
    });
});
```

### Visual Regression Tests
```javascript
test.describe('Switch Visual Tests', () => {
    test('all switch variants', async ({ page }) => {
        await page.goto('/components/switch');
        
        // Test all sizes
        for (const size of ['xs', 'sm', 'md', 'lg', 'xl']) {
            await expect(page.locator(`[data-size="${size}"]`)).toHaveScreenshot(`switch-${size}.png`);
        }
        
        // Test all states
        const states = ['off', 'on', 'disabled', 'loading', 'error'];
        for (const state of states) {
            await expect(page.locator(`[data-state="${state}"]`)).toHaveScreenshot(`switch-${state}.png`);
        }
        
        // Test animations
        await expect(page.locator('[data-animated="true"]')).toHaveScreenshot('switch-animated.png');
    });
});
```

## 📚 Usage Examples

### Settings Panel
```go
templ UserSettingsPanel() {
    @SwitchGroup(SwitchGroupProps{
        Label: "Privacy Settings",
        Switches: []SwitchOption{
            {
                Name:        "profileVisible",
                Label:       "Public Profile",
                Description: "Make your profile visible to other users",
                Icon:        "user",
                Checked:     true,
            },
            {
                Name:        "activityTracking",
                Label:       "Activity Tracking",
                Description: "Allow us to track your activity for analytics",
                Icon:        "activity",
                Checked:     false,
            },
            {
                Name:        "marketingEmails",
                Label:       "Marketing Emails", 
                Description: "Receive promotional emails and updates",
                Icon:        "mail",
                Checked:     false,
            },
        },
    })
}
```

### Dark Mode Toggle
```go
templ DarkModeToggle() {
    @AnimatedSwitch(SwitchProps{
        Name:      "darkMode",
        Label:     "Dark Mode",
        OnIcon:    "moon",
        OffIcon:   "sun",
        ShowIcons: true,
        Size:      "lg",
        Variant:   "primary",
        OnChange:  "toggleDarkMode",
    })
}
```

### Feature Toggles
```go
templ FeatureToggles() {
    <div class="feature-toggles">
        <h3>Experimental Features</h3>
        
        @Switch(SwitchProps{
            Name:        "betaFeatures",
            Label:       "Beta Features",
            Description: "Enable experimental features (may be unstable)",
            OnLabel:     "Enabled",
            OffLabel:    "Disabled",
            ShowLabels:  true,
            Size:        "md",
        })
        
        @ConfirmationSwitch(SwitchProps{
            Name:           "debugMode",
            Label:          "Debug Mode",
            Description:    "Enable detailed logging and debug information",
            ConfirmTitle:   "Enable Debug Mode?",
            ConfirmMessage: "This will enable detailed logging which may impact performance.",
        })
    </div>
}
```

## 🔗 Related Components

- **[Checkbox](../checkbox/)**: Multiple selections from options
- **[Radio Button](../radio/)**: Single selection from options
- **[Button](../button/)**: Action triggers and submissions
- **[Toggle Button Group](../../molecules/toggle-group/)**: Multiple toggle selections

---

**Component Status**: ✅ Production Ready  
**Schema Reference**: `SwitchControlSchema.json`  
**CSS Classes**: `.switch-group`, `.switch-{size}`, `.switch-{variant}`  
**Accessibility**: WCAG 2.1 AA Compliant