# Radio Button Component

**FILE PURPOSE**: Single selection control implementation and specifications  
**SCOPE**: All radio button variants, groups, and interaction patterns  
**TARGET AUDIENCE**: Developers implementing form controls and selection interfaces

## 📋 Component Overview

The Radio Button component provides intuitive single selection from a group of mutually exclusive options. It supports various layouts, validation, and enhanced visual feedback while maintaining full accessibility compliance.

### Schema Reference
- **Primary Schema**: `RadioControlSchema.json`
- **Related Schemas**: `Options.json`, `Option.json`
- **Base Interface**: Form control with single selection

## 🎨 Radio Button Types

### Basic Radio Button
**Purpose**: Simple single selection from a group of options

```go
// Basic radio button configuration
basicRadio := RadioProps{
    Name:     "paymentMethod",
    Value:    "credit",
    Label:    "Credit Card",
    Checked:  false,
    Required: true,
}

// Generated Templ component
templ BasicRadio(props RadioProps) {
    <div class="radio-group">
        <label class="radio-label" for={ props.ID }>
            <input 
                type="radio"
                id={ props.ID }
                name={ props.Name }
                value={ props.Value }
                class={ props.GetClasses() }
                checked?={ props.Checked }
                required?={ props.Required }
                disabled?={ props.Disabled }
                @change={ props.OnChange }
            />
            <span class="radio-indicator"></span>
            <span class="radio-text">{ props.Label }</span>
        </label>
        
        if props.Description != "" {
            <div class="radio-description">{ props.Description }</div>
        }
        
        if props.Error != "" {
            <div class="error-message">{ props.Error }</div>
        }
    </div>
}
```

### Radio Group
**Purpose**: Complete single-selection interface with multiple options

```go
radioGroup := RadioGroupProps{
    Name:     "subscription",
    Label:    "Select Subscription Plan",
    Required: true,
    Value:    "pro", // Selected value
    Options: []RadioOption{
        {
            Value:       "basic",
            Label:       "Basic Plan",
            Description: "$9/month - Essential features",
            Price:       "$9",
            Badge:       "Popular",
        },
        {
            Value:       "pro",
            Label:       "Pro Plan", 
            Description: "$19/month - Advanced features",
            Price:       "$19",
            Badge:       "Recommended",
        },
        {
            Value:       "enterprise",
            Label:       "Enterprise Plan",
            Description: "$49/month - All features",
            Price:       "$49",
            Disabled:    false,
        },
    },
}

templ RadioGroup(props RadioGroupProps) {
    <fieldset class="radio-group-container" x-data={ fmt.Sprintf(`{
        selected: '%s',
        change(value) {
            this.selected = value;
            $dispatch('radio-change', { name: '%s', value: value });
        }
    }`, props.Value, props.Name) }>
        
        <legend class="group-legend">
            { props.Label }
            if props.Required {
                <span class="required" aria-label="required">*</span>
            }
        </legend>
        
        if props.HelpText != "" {
            <div class="group-help-text">{ props.HelpText }</div>
        }
        
        <div class="radio-options" :class="{ [`layout-${props.Layout}`]: true }">
            for _, option := range props.Options {
                <div class="radio-option" :class="{ 'selected': selected === '{ option.Value }' }">
                    <label class="radio-label">
                        <input 
                            type="radio"
                            name={ props.Name }
                            value={ option.Value }
                            :checked="selected === '{ option.Value }'"
                            @change={ fmt.Sprintf("change('%s')", option.Value) }
                            disabled?={ option.Disabled }
                            class="radio-input"
                        />
                        <span class="radio-indicator"></span>
                        
                        <div class="radio-content">
                            <div class="radio-header">
                                <span class="radio-title">{ option.Label }</span>
                                if option.Badge != "" {
                                    @Badge(BadgeProps{
                                        Text:    option.Badge,
                                        Variant: "primary",
                                        Size:    "sm",
                                    })
                                }
                                if option.Price != "" {
                                    <span class="radio-price">{ option.Price }</span>
                                }
                            </div>
                            
                            if option.Description != "" {
                                <div class="radio-description">{ option.Description }</div>
                            }
                            
                            if len(option.Features) > 0 {
                                <ul class="radio-features">
                                    for _, feature := range option.Features {
                                        <li class="feature-item">
                                            @Icon(IconProps{Name: "check", Size: "xs"})
                                            <span>{ feature }</span>
                                        </li>
                                    }
                                </ul>
                            }
                        </div>
                    </label>
                </div>
            }
        </div>
        
        if props.Error != "" {
            <div class="error-message" role="alert">{ props.Error }</div>
        }
    </fieldset>
}
```

### Card-Style Radio Group
**Purpose**: Enhanced visual selection with card layouts

```go
templ CardRadioGroup(props RadioGroupProps) {
    <div class="card-radio-group" x-data={ fmt.Sprintf(`{
        selected: '%s',
        change(value) {
            this.selected = value;
        }
    }`, props.Value) }>
        
        <h3 class="group-title">{ props.Label }</h3>
        
        <div class="card-options">
            for _, option := range props.Options {
                <div class="radio-card" 
                     :class="{ 
                         'selected': selected === '{ option.Value }',
                         'disabled': { fmt.Sprintf("%t", option.Disabled) }
                     }">
                    <label class="card-label">
                        <input 
                            type="radio"
                            name={ props.Name }
                            value={ option.Value }
                            :checked="selected === '{ option.Value }'"
                            @change={ fmt.Sprintf("change('%s')", option.Value) }
                            disabled?={ option.Disabled }
                            class="card-input sr-only"
                        />
                        
                        <div class="card-header">
                            if option.Icon != "" {
                                <div class="card-icon">
                                    @Icon(IconProps{Name: option.Icon, Size: "lg"})
                                </div>
                            }
                            <div class="card-indicator">
                                <span class="radio-dot"></span>
                            </div>
                        </div>
                        
                        <div class="card-body">
                            <h4 class="card-title">{ option.Label }</h4>
                            if option.Description != "" {
                                <p class="card-description">{ option.Description }</p>
                            }
                            if option.Price != "" {
                                <div class="card-price">{ option.Price }</div>
                            }
                        </div>
                        
                        if option.Badge != "" {
                            <div class="card-badge">
                                @Badge(BadgeProps{Text: option.Badge, Variant: "success"})
                            </div>
                        }
                    </label>
                </div>
            }
        </div>
    </div>
}
```

### Image Radio Group
**Purpose**: Visual selection with image previews

```go
templ ImageRadioGroup(props RadioGroupProps) {
    <div class="image-radio-group" x-data={ fmt.Sprintf(`{
        selected: '%s'
    }`, props.Value) }>
        
        <h3 class="group-title">{ props.Label }</h3>
        
        <div class="image-options">
            for _, option := range props.Options {
                <div class="image-radio-option">
                    <label class="image-label">
                        <input 
                            type="radio"
                            name={ props.Name }
                            value={ option.Value }
                            x-model="selected"
                            class="image-input sr-only"
                        />
                        
                        <div class="image-container" 
                             :class="{ 'selected': selected === '{ option.Value }' }">
                            if option.ImageSrc != "" {
                                <img 
                                    src={ option.ImageSrc }
                                    alt={ option.Label }
                                    class="option-image"
                                />
                            }
                            <div class="image-overlay">
                                <div class="selection-indicator">
                                    @Icon(IconProps{Name: "check-circle", Size: "lg"})
                                </div>
                            </div>
                        </div>
                        
                        <div class="image-label-text">
                            <span class="image-title">{ option.Label }</span>
                            if option.Description != "" {
                                <span class="image-description">{ option.Description }</span>
                            }
                        </div>
                    </label>
                </div>
            }
        </div>
    </div>
}
```

## 🎯 Props Interface

### Core Properties
```go
type RadioProps struct {
    // Identity
    Name     string `json:"name"`
    ID       string `json:"id"`
    TestID   string `json:"testid"`
    
    // Content
    Label       string `json:"label"`
    Value       string `json:"value"`
    Description string `json:"description"`
    
    // State
    Checked   bool `json:"checked"`
    Required  bool `json:"required"`
    Disabled  bool `json:"disabled"`
    ReadOnly  bool `json:"readonly"`
    
    // Appearance
    Size    RadioSize    `json:"size"`        // xs, sm, md, lg, xl
    Variant RadioVariant `json:"variant"`     // default, primary, success, error
    Class   string       `json:"className"`
    Style   map[string]string `json:"style"`
    
    // Events
    OnChange string `json:"onChange"`
    OnFocus  string `json:"onFocus"`
    OnBlur   string `json:"onBlur"`
    
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
type RadioGroupProps struct {
    // Identity
    Name   string `json:"name"`
    ID     string `json:"id"`
    
    // Content
    Label    string        `json:"label"`
    HelpText string        `json:"helpText"`
    Error    string        `json:"error"`
    Options  []RadioOption `json:"options"`
    
    // State
    Value    string `json:"value"`     // Selected value
    Required bool   `json:"required"`
    Disabled bool   `json:"disabled"`
    
    // Layout
    Layout    RadioLayout `json:"layout"`    // vertical, horizontal, grid, cards
    Columns   int         `json:"columns"`   // Grid layout columns
    Direction string      `json:"direction"` // row, column
    
    // Appearance
    Size    RadioSize    `json:"size"`
    Variant RadioVariant `json:"variant"`
    
    // Events
    OnChange string `json:"onChange"`
    
    // Validation
    ValidationMessage string `json:"validationMessage"`
}

type RadioOption struct {
    Value       string   `json:"value"`
    Label       string   `json:"label"`
    Description string   `json:"description"`
    Price       string   `json:"price"`
    Badge       string   `json:"badge"`
    Icon        string   `json:"icon"`
    ImageSrc    string   `json:"imageSrc"`
    Features    []string `json:"features"`
    Disabled    bool     `json:"disabled"`
    Checked     bool     `json:"checked"`
}
```

### Size Variants
```go
type RadioSize string

const (
    RadioXS RadioSize = "xs"    // 14px indicator
    RadioSM RadioSize = "sm"    // 16px indicator  
    RadioMD RadioSize = "md"    // 18px indicator (default)
    RadioLG RadioSize = "lg"    // 20px indicator
    RadioXL RadioSize = "xl"    // 24px indicator
)
```

### Layout Types
```go
type RadioLayout string

const (
    RadioVertical   RadioLayout = "vertical"   // Stacked vertically
    RadioHorizontal RadioLayout = "horizontal" // Side by side
    RadioGrid       RadioLayout = "grid"       // Grid layout
    RadioCards      RadioLayout = "cards"      // Card-style layout
    RadioImages     RadioLayout = "images"     // Image-based layout
)
```

### Visual Variants
```go
type RadioVariant string

const (
    RadioDefault RadioVariant = "default"  // Gray theme
    RadioPrimary RadioVariant = "primary"  // Brand color
    RadioSuccess RadioVariant = "success"  // Green theme
    RadioWarning RadioVariant = "warning"  // Yellow theme
    RadioDanger  RadioVariant = "danger"   // Red theme
)
```

## 🎨 Styling Implementation

### Base Radio Styles
```css
.radio-group {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
}

.radio-label {
    display: flex;
    align-items: flex-start;
    gap: var(--space-sm);
    cursor: pointer;
    user-select: none;
    line-height: 1.5;
    
    &:hover .radio-indicator {
        border-color: var(--color-border-dark);
        background: var(--color-bg-hover);
    }
}

.radio-input {
    position: absolute;
    opacity: 0;
    width: 0;
    height: 0;
    
    /* Focus styles */
    &:focus + .radio-indicator {
        outline: none;
        border-color: var(--color-primary);
        box-shadow: 0 0 0 3px var(--color-primary-light);
    }
    
    /* Checked state */
    &:checked + .radio-indicator {
        border-color: var(--color-primary);
        
        &::after {
            opacity: 1;
            transform: scale(1);
        }
    }
    
    /* Disabled state */
    &:disabled + .radio-indicator {
        background: var(--color-bg-disabled);
        border-color: var(--color-border-disabled);
        cursor: not-allowed;
        
        &::after {
            background: var(--color-text-disabled);
        }
    }
}

.radio-indicator {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    border: 2px solid var(--color-border-medium);
    border-radius: 50%;
    background: var(--color-bg-surface);
    transition: var(--transition-base);
    flex-shrink: 0;
    
    /* Radio dot */
    &::after {
        content: '';
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: var(--color-primary);
        opacity: 0;
        transform: scale(0);
        transition: var(--transition-quick);
    }
}

.radio-text {
    color: var(--color-text-primary);
    font-size: var(--font-size-base);
}

.radio-description {
    color: var(--color-text-secondary);
    font-size: var(--font-size-sm);
    margin-top: var(--space-xs);
    line-height: 1.4;
}
```

### Group Layout Styles
```css
.radio-group-container {
    border: none;
    padding: 0;
    margin: 0;
}

.group-legend {
    font-weight: var(--font-weight-semibold);
    color: var(--color-text-primary);
    margin-bottom: var(--space-md);
    padding: 0;
}

.group-help-text {
    color: var(--color-text-secondary);
    font-size: var(--font-size-sm);
    margin-bottom: var(--space-md);
    line-height: 1.4;
}

.radio-options {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
    
    /* Horizontal layout */
    &.layout-horizontal {
        flex-direction: row;
        flex-wrap: wrap;
    }
    
    /* Grid layout */
    &.layout-grid {
        display: grid;
        grid-template-columns: repeat(var(--columns, 2), 1fr);
        gap: var(--space-md);
    }
    
    /* Card layout */
    &.layout-cards {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
        gap: var(--space-lg);
    }
}

.radio-option {
    .radio-content {
        display: flex;
        flex-direction: column;
        gap: var(--space-xs);
    }
    
    .radio-header {
        display: flex;
        align-items: center;
        gap: var(--space-sm);
    }
    
    .radio-title {
        font-weight: var(--font-weight-medium);
        color: var(--color-text-primary);
    }
    
    .radio-price {
        font-weight: var(--font-weight-semibold);
        color: var(--color-primary);
        margin-left: auto;
    }
    
    .radio-features {
        list-style: none;
        padding: 0;
        margin: var(--space-sm) 0 0 0;
        
        .feature-item {
            display: flex;
            align-items: center;
            gap: var(--space-xs);
            color: var(--color-text-secondary);
            font-size: var(--font-size-sm);
            
            svg {
                color: var(--color-success);
            }
        }
    }
}
```

### Card Style Variants
```css
.card-radio-group {
    .group-title {
        font-size: var(--font-size-lg);
        font-weight: var(--font-weight-semibold);
        margin-bottom: var(--space-lg);
        color: var(--color-text-primary);
    }
    
    .card-options {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
        gap: var(--space-lg);
    }
}

.radio-card {
    position: relative;
    border: 2px solid var(--color-border-light);
    border-radius: var(--radius-lg);
    background: var(--color-bg-surface);
    transition: var(--transition-base);
    cursor: pointer;
    
    &:hover {
        border-color: var(--color-border-medium);
        box-shadow: var(--shadow-sm);
    }
    
    &.selected {
        border-color: var(--color-primary);
        box-shadow: 0 0 0 1px var(--color-primary-light);
    }
    
    &.disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }
    
    .card-label {
        display: block;
        padding: var(--space-lg);
        height: 100%;
        cursor: inherit;
    }
    
    .card-header {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        margin-bottom: var(--space-md);
    }
    
    .card-icon {
        width: 48px;
        height: 48px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--color-bg-secondary);
        border-radius: var(--radius-md);
        color: var(--color-text-secondary);
    }
    
    .card-indicator {
        .radio-dot {
            width: 20px;
            height: 20px;
            border: 2px solid var(--color-border-medium);
            border-radius: 50%;
            background: var(--color-bg-surface);
            position: relative;
            
            &::after {
                content: '';
                position: absolute;
                top: 50%;
                left: 50%;
                transform: translate(-50%, -50%) scale(0);
                width: 8px;
                height: 8px;
                border-radius: 50%;
                background: var(--color-primary);
                transition: var(--transition-quick);
            }
        }
    }
    
    &.selected .radio-dot {
        border-color: var(--color-primary);
        
        &::after {
            transform: translate(-50%, -50%) scale(1);
        }
    }
    
    .card-body {
        .card-title {
            font-size: var(--font-size-lg);
            font-weight: var(--font-weight-semibold);
            color: var(--color-text-primary);
            margin: 0 0 var(--space-sm) 0;
        }
        
        .card-description {
            color: var(--color-text-secondary);
            font-size: var(--font-size-sm);
            line-height: 1.4;
            margin: 0;
        }
        
        .card-price {
            font-size: var(--font-size-xl);
            font-weight: var(--font-weight-bold);
            color: var(--color-primary);
            margin-top: var(--space-md);
        }
    }
    
    .card-badge {
        position: absolute;
        top: var(--space-md);
        right: var(--space-md);
    }
}
```

### Image Style Variants
```css
.image-radio-group {
    .image-options {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
        gap: var(--space-lg);
    }
}

.image-radio-option {
    .image-label {
        display: block;
        cursor: pointer;
    }
    
    .image-container {
        position: relative;
        border-radius: var(--radius-lg);
        overflow: hidden;
        border: 3px solid transparent;
        transition: var(--transition-base);
        
        &:hover {
            transform: translateY(-2px);
            box-shadow: var(--shadow-lg);
        }
        
        &.selected {
            border-color: var(--color-primary);
            
            .image-overlay {
                opacity: 1;
            }
        }
    }
    
    .option-image {
        width: 100%;
        height: 160px;
        object-fit: cover;
        display: block;
    }
    
    .image-overlay {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(var(--color-primary-rgb), 0.8);
        display: flex;
        align-items: center;
        justify-content: center;
        opacity: 0;
        transition: var(--transition-base);
        
        .selection-indicator {
            color: white;
        }
    }
    
    .image-label-text {
        padding: var(--space-md) 0;
        text-align: center;
        
        .image-title {
            display: block;
            font-weight: var(--font-weight-medium);
            color: var(--color-text-primary);
            margin-bottom: var(--space-xs);
        }
        
        .image-description {
            color: var(--color-text-secondary);
            font-size: var(--font-size-sm);
        }
    }
}
```

### Size Variants
```css
/* Size-specific styling */
.radio-xs .radio-indicator {
    width: 14px;
    height: 14px;
    
    &::after {
        width: 6px;
        height: 6px;
    }
}

.radio-sm .radio-indicator {
    width: 16px;
    height: 16px;
    
    &::after {
        width: 7px;
        height: 7px;
    }
}

.radio-lg .radio-indicator {
    width: 20px;
    height: 20px;
    
    &::after {
        width: 9px;
        height: 9px;
    }
}

.radio-xl .radio-indicator {
    width: 24px;
    height: 24px;
    
    &::after {
        width: 11px;
        height: 11px;
    }
}
```

### Color Variants
```css
/* Primary variant */
.radio-primary .radio-input:checked + .radio-indicator {
    border-color: var(--color-primary);
    
    &::after {
        background: var(--color-primary);
    }
}

/* Success variant */
.radio-success .radio-input:checked + .radio-indicator {
    border-color: var(--color-success);
    
    &::after {
        background: var(--color-success);
    }
}

/* Warning variant */
.radio-warning .radio-input:checked + .radio-indicator {
    border-color: var(--color-warning);
    
    &::after {
        background: var(--color-warning);
    }
}

/* Danger variant */
.radio-danger .radio-input:checked + .radio-indicator {
    border-color: var(--color-danger);
    
    &::after {
        background: var(--color-danger);
    }
}
```

## ⚙️ Advanced Features

### Dynamic Option Loading
```go
templ DynamicRadioGroup(props RadioGroupProps) {
    <div class="dynamic-radio-group" 
         x-data={ fmt.Sprintf(`{
            loading: false,
            options: %s,
            selected: '%s',
            async loadOptions() {
                this.loading = true;
                try {
                    const response = await fetch('/api/options/%s');
                    this.options = await response.json();
                } finally {
                    this.loading = false;
                }
            }
        }`, toJSON(props.Options), props.Value, props.Name) }>
        
        <h3 class="group-title">{ props.Label }</h3>
        
        <div x-show="loading" class="loading-state">
            @Spinner(SpinnerProps{Size: "md"})
            <span>Loading options...</span>
        </div>
        
        <div x-show="!loading" class="radio-options">
            <template x-for="option in options" :key="option.value">
                <div class="radio-option">
                    <label class="radio-label">
                        <input 
                            type="radio"
                            :name="name"
                            :value="option.value"
                            x-model="selected"
                            class="radio-input"
                        />
                        <span class="radio-indicator"></span>
                        <div class="radio-content">
                            <span class="radio-title" x-text="option.label"></span>
                            <span class="radio-description" x-text="option.description"></span>
                        </div>
                    </label>
                </div>
            </template>
        </div>
    </div>
}
```

### Conditional Radio Options
```go
templ ConditionalRadioGroup(props RadioGroupProps) {
    <div class="conditional-radio-group" 
         x-data={ fmt.Sprintf(`{
            selected: '%s',
            conditions: %s,
            isOptionVisible(option) {
                if (!option.condition) return true;
                return this.evaluateCondition(option.condition);
            },
            evaluateCondition(condition) {
                // Implement condition evaluation logic
                return true;
            }
        }`, props.Value, toJSON(props.Conditions)) }>
        
        <fieldset class="radio-group-container">
            <legend class="group-legend">{ props.Label }</legend>
            
            <div class="radio-options">
                for _, option := range props.Options {
                    <div class="radio-option" 
                         x-show={ fmt.Sprintf("isOptionVisible(%s)", toJSON(option)) }>
                        @RadioOption(option)
                    </div>
                }
            </div>
        </fieldset>
    </div>
}
```

### Radio with Custom Validation
```go
templ ValidatedRadioGroup(props RadioGroupProps) {
    <div class="validated-radio-group" 
         x-data={ fmt.Sprintf(`{
            selected: '%s',
            error: '',
            touched: false,
            get isValid() {
                if (!this.touched) return true;
                return this.selected !== '';
            },
            validate() {
                this.touched = true;
                if (!this.isValid) {
                    this.error = '%s';
                } else {
                    this.error = '';
                }
                return this.isValid;
            },
            change(value) {
                this.selected = value;
                this.validate();
                $dispatch('radio-change', { name: '%s', value: value, valid: this.isValid });
            }
        }`, props.Value, props.ValidationMessage, props.Name) }>
        
        @RadioGroup(props)
        
        <div x-show="error" x-text="error" class="error-message" role="alert"></div>
    </div>
}
```

## 📱 Responsive Design

### Mobile Optimizations
```css
/* Mobile-specific radio styling */
@media (max-width: 479px) {
    .radio-label {
        min-height: 44px;  /* Touch target */
        align-items: center;
    }
    
    .radio-indicator {
        width: 20px;       /* Larger touch target */
        height: 20px;
        
        &::after {
            width: 9px;
            height: 9px;
        }
    }
    
    /* Stack options vertically on mobile */
    .radio-options.layout-horizontal {
        flex-direction: column;
    }
    
    .radio-options.layout-grid {
        grid-template-columns: 1fr;
    }
    
    /* Simplify card layout on mobile */
    .card-options {
        grid-template-columns: 1fr;
    }
    
    .radio-card .card-label {
        padding: var(--space-md);
    }
    
    /* Stack image options on mobile */
    .image-options {
        grid-template-columns: repeat(2, 1fr);
    }
}

/* Very small screens */
@media (max-width: 320px) {
    .image-options {
        grid-template-columns: 1fr;
    }
}
```

### Touch Interactions
```css
/* Touch-friendly interactions */
@media (hover: none) {
    .radio-label {
        /* Remove hover effects on touch devices */
        &:hover .radio-indicator {
            border-color: var(--color-border-medium);
            background: var(--color-bg-surface);
        }
    }
    
    .radio-card {
        &:hover {
            border-color: var(--color-border-light);
            box-shadow: none;
        }
    }
    
    /* Larger tap targets */
    .radio-indicator {
        min-width: 44px;
        min-height: 44px;
    }
}
```

## ♿ Accessibility Implementation

### ARIA Support
```go
func (radio RadioGroupProps) GetAriaAttributes() map[string]string {
    attrs := make(map[string]string)
    
    // Required state
    if radio.Required {
        attrs["aria-required"] = "true"
    }
    
    // Invalid state
    if radio.Error != "" {
        attrs["aria-invalid"] = "true"
        if radio.ID != "" {
            attrs["aria-describedby"] = radio.ID + "-error"
        }
    }
    
    // Help text association
    if radio.HelpText != "" && radio.ID != "" {
        describedBy := attrs["aria-describedby"]
        if describedBy != "" {
            attrs["aria-describedby"] = describedBy + " " + radio.ID + "-help"
        } else {
            attrs["aria-describedby"] = radio.ID + "-help"
        }
    }
    
    return attrs
}
```

### Screen Reader Support
```go
templ AccessibleRadioGroup(props RadioGroupProps) {
    <fieldset class="radio-group-container" 
              role="radiogroup"
              for attrName, attrValue := range props.GetAriaAttributes() {
                  { attrName }={ attrValue }
              }>
        
        <legend class="group-legend">
            { props.Label }
            if props.Required {
                <span class="required" aria-label="required">*</span>
            }
        </legend>
        
        if props.HelpText != "" {
            <div id={ props.ID + "-help" } class="group-help-text">
                { props.HelpText }
            </div>
        }
        
        <div class="radio-options">
            for i, option := range props.Options {
                <div class="radio-option">
                    <label class="radio-label">
                        <input 
                            type="radio"
                            id={ fmt.Sprintf("%s-%d", props.ID, i) }
                            name={ props.Name }
                            value={ option.Value }
                            checked?={ option.Value == props.Value }
                            disabled?={ option.Disabled }
                            class="radio-input"
                            aria-describedby?={ option.Description != "" ? fmt.Sprintf("%s-%d-desc", props.ID, i) : "" }
                        />
                        <span class="radio-indicator" aria-hidden="true"></span>
                        <span class="radio-text">{ option.Label }</span>
                    </label>
                    
                    if option.Description != "" {
                        <div id={ fmt.Sprintf("%s-%d-desc", props.ID, i) } class="radio-description">
                            { option.Description }
                        </div>
                    }
                </div>
            }
        </div>
        
        if props.Error != "" {
            <div 
                id={ props.ID + "-error" }
                class="error-message"
                role="alert"
                aria-live="polite">
                { props.Error }
            </div>
        }
    </fieldset>
}
```

### Keyboard Navigation
```css
/* Focus management */
.radio-input:focus + .radio-indicator {
    outline: none;
    border-color: var(--color-primary);
    box-shadow: 0 0 0 3px var(--color-primary-light);
    z-index: 1;
    position: relative;
}

/* High contrast mode support */
@media (prefers-contrast: high) {
    .radio-indicator {
        border-width: 2px;
        border-color: CanvasText;
    }
    
    .radio-input:checked + .radio-indicator::after {
        background: CanvasText;
    }
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
    .radio-indicator,
    .radio-indicator::after {
        transition: none;
    }
    
    .radio-card {
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
func TestRadioComponent(t *testing.T) {
    tests := []struct {
        name     string
        props    RadioGroupProps
        expected []string
    }{
        {
            name: "basic radio group",
            props: RadioGroupProps{
                Name:  "test",
                Label: "Test Options",
                Options: []RadioOption{
                    {Value: "option1", Label: "Option 1"},
                    {Value: "option2", Label: "Option 2"},
                },
            },
            expected: []string{"radio-group-container", "Test Options", "Option 1", "Option 2"},
        },
        {
            name: "required radio group",
            props: RadioGroupProps{
                Name:     "required",
                Required: true,
            },
            expected: []string{"required", "aria-required=\"true\""},
        },
        {
            name: "selected option",
            props: RadioGroupProps{
                Name:  "selected",
                Value: "option1",
                Options: []RadioOption{
                    {Value: "option1", Label: "Option 1"},
                },
            },
            expected: []string{"checked"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            component := renderRadioGroup(tt.props)
            for _, expected := range tt.expected {
                assert.Contains(t, component, expected)
            }
        })
    }
}
```

### Accessibility Tests
```javascript
describe('Radio Button Accessibility', () => {
    test('has proper radiogroup role', () => {
        const radio = render(
            <RadioGroup 
                name="test" 
                label="Test Options" 
                options={[
                    { value: "1", label: "Option 1" },
                    { value: "2", label: "Option 2" }
                ]} 
            />
        );
        const radiogroup = screen.getByRole('radiogroup');
        expect(radiogroup).toBeInTheDocument();
    });
    
    test('supports keyboard navigation', () => {
        const radio = render(<RadioGroup name="test" options={options} />);
        const radioButtons = screen.getAllByRole('radio');
        
        radioButtons[0].focus();
        expect(radioButtons[0]).toHaveFocus();
        
        fireEvent.keyDown(radioButtons[0], { key: 'ArrowDown' });
        expect(radioButtons[1]).toHaveFocus();
        expect(radioButtons[1]).toBeChecked();
    });
    
    test('announces selection changes', () => {
        const radio = render(<RadioGroup name="test" options={options} />);
        const radioButton = screen.getByLabelText('Option 1');
        
        fireEvent.click(radioButton);
        expect(radioButton).toBeChecked();
    });
});
```

### Visual Regression Tests
```javascript
test.describe('Radio Button Visual Tests', () => {
    test('all radio variants', async ({ page }) => {
        await page.goto('/components/radio');
        
        // Test all layouts
        for (const layout of ['vertical', 'horizontal', 'grid', 'cards']) {
            await expect(page.locator(`[data-layout="${layout}"]`)).toHaveScreenshot(`radio-${layout}.png`);
        }
        
        // Test all states
        const states = ['default', 'checked', 'disabled', 'error'];
        for (const state of states) {
            await expect(page.locator(`[data-state="${state}"]`)).toHaveScreenshot(`radio-${state}.png`);
        }
    });
});
```

## 📚 Usage Examples

### Payment Method Selection
```go
templ PaymentMethodSelection() {
    @RadioGroup(RadioGroupProps{
        Name:     "paymentMethod",
        Label:    "Select Payment Method",
        Required: true,
        Options: []RadioOption{
            {
                Value:       "credit",
                Label:       "Credit Card",
                Description: "Visa, MasterCard, American Express",
                Icon:        "credit-card",
            },
            {
                Value:       "paypal",
                Label:       "PayPal",
                Description: "Pay with your PayPal account",
                Icon:        "paypal",
            },
            {
                Value:       "bank",
                Label:       "Bank Transfer",
                Description: "Direct bank account transfer",
                Icon:        "bank",
            },
        },
    })
}
```

### Subscription Plan Cards
```go
templ SubscriptionPlans() {
    @CardRadioGroup(RadioGroupProps{
        Name:   "subscription",
        Label:  "Choose Your Plan",
        Layout: RadioCards,
        Options: []RadioOption{
            {
                Value:       "starter",
                Label:       "Starter",
                Description: "Perfect for individuals",
                Price:       "$9/month",
                Badge:       "Popular",
                Features: []string{
                    "5 Projects",
                    "10GB Storage",
                    "Email Support",
                },
            },
            {
                Value:       "professional",
                Label:       "Professional",
                Description: "For growing teams",
                Price:       "$29/month",
                Badge:       "Recommended",
                Features: []string{
                    "Unlimited Projects",
                    "100GB Storage",
                    "Priority Support",
                    "Advanced Analytics",
                },
            },
        },
    })
}
```

### Theme Selection with Images
```go
templ ThemeSelection() {
    @ImageRadioGroup(RadioGroupProps{
        Name:   "theme",
        Label:  "Select Application Theme",
        Layout: RadioImages,
        Options: []RadioOption{
            {
                Value:       "light",
                Label:       "Light Theme",
                Description: "Clean and bright interface",
                ImageSrc:    "/images/theme-light.png",
            },
            {
                Value:       "dark",
                Label:       "Dark Theme", 
                Description: "Easy on the eyes",
                ImageSrc:    "/images/theme-dark.png",
            },
            {
                Value:       "auto",
                Label:       "Auto Theme",
                Description: "Follows system preference",
                ImageSrc:    "/images/theme-auto.png",
            },
        },
    })
}
```

## 🔗 Related Components

- **[Checkbox](../checkbox/)**: Multiple selections from options
- **[Switch](../switch/)**: Toggle controls for settings
- **[Dropdown](../../molecules/dropdown/)**: Selection menus with search
- **[Button Group](../../molecules/button-group/)**: Action button selections

---

**Component Status**: ✅ Production Ready  
**Schema Reference**: `RadioControlSchema.json`  
**CSS Classes**: `.radio-group`, `.radio-{size}`, `.radio-{variant}`  
**Accessibility**: WCAG 2.1 AA Compliant