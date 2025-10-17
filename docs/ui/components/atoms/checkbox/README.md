# Checkbox Component

**FILE PURPOSE**: Boolean selection control implementation and specifications  
**SCOPE**: All checkbox variants, validation, and interaction patterns  
**TARGET AUDIENCE**: Developers implementing form controls and selection interfaces

## 📋 Component Overview

The Checkbox component provides intuitive boolean selection for forms, lists, and data tables. It supports individual selections, group selections, and indeterminate states with full accessibility compliance.

### Schema Reference
- **Primary Schema**: `CheckboxControlSchema.json`
- **Related Schemas**: `IconCheckedSchema.json`, `Options.json`
- **Base Interface**: Form control with validation

## 🎨 Checkbox Types

### Basic Checkbox
**Purpose**: Simple boolean selection for forms and agreements

```go
// Basic checkbox configuration
basicCheckbox := CheckboxProps{
    Name:     "terms",
    Label:    "I agree to the Terms of Service",
    Value:    "accepted",
    Required: true,
    Checked:  false,
}

// Generated Templ component
templ BasicCheckbox(props CheckboxProps) {
    <div class="checkbox-group">
        <label class="checkbox-label" for={ props.ID }>
            <input 
                type="checkbox"
                id={ props.ID }
                name={ props.Name }
                value={ props.Value }
                class={ props.GetClasses() }
                checked?={ props.Checked }
                required?={ props.Required }
                disabled?={ props.Disabled }
                @change={ props.OnChange }
            />
            <span class="checkbox-indicator"></span>
            <span class="checkbox-text">{ props.Label }</span>
        </label>
        
        if props.HelpText != "" {
            <div class="help-text">{ props.HelpText }</div>
        }
        
        if props.Error != "" {
            <div class="error-message">{ props.Error }</div>
        }
    </div>
}
```

### Checkbox with Custom Icon
**Purpose**: Enhanced visual feedback with custom check icons

```go
iconCheckbox := CheckboxProps{
    Name:        "notifications",
    Label:       "Enable notifications", 
    Value:       "enabled",
    CheckedIcon: "check-circle",
    UncheckedIcon: "circle",
    Size:        "lg",
    Variant:     "primary",
}

templ IconCheckbox(props CheckboxProps) {
    <div class="checkbox-group icon-checkbox" x-data="{ checked: false }">
        <label class="checkbox-label">
            <input 
                type="checkbox"
                x-model="checked"
                class="sr-only"
                name={ props.Name }
                value={ props.Value }
            />
            
            <div class="checkbox-icon-container">
                if props.UncheckedIcon != "" {
                    <span x-show="!checked" class="unchecked-icon">
                        @Icon(IconProps{Name: props.UncheckedIcon, Size: props.Size})
                    </span>
                }
                if props.CheckedIcon != "" {
                    <span x-show="checked" class="checked-icon">
                        @Icon(IconProps{Name: props.CheckedIcon, Size: props.Size})
                    </span>
                }
            </div>
            
            <span class="checkbox-text">{ props.Label }</span>
        </label>
    </div>
}
```

### Checkbox Group
**Purpose**: Multiple related selections with group validation

```go
checkboxGroup := CheckboxGroupProps{
    Name:     "features",
    Label:    "Select Features",
    Required: true,
    MinItems: 1,
    MaxItems: 3,
    Options: []CheckboxOption{
        {Value: "sms", Label: "SMS Notifications", Description: "Receive text messages"},
        {Value: "email", Label: "Email Alerts", Description: "Receive email notifications"},
        {Value: "push", Label: "Push Notifications", Description: "Mobile app notifications"},
        {Value: "slack", Label: "Slack Integration", Description: "Connect to Slack workspace"},
    },
}

templ CheckboxGroup(props CheckboxGroupProps) {
    <fieldset class="checkbox-group-container" x-data={ fmt.Sprintf(`{
        selected: [],
        get isValid() {
            return this.selected.length >= %d && this.selected.length <= %d;
        },
        toggle(value) {
            if (this.selected.includes(value)) {
                this.selected = this.selected.filter(v => v !== value);
            } else {
                this.selected.push(value);
            }
        }
    }`, props.MinItems, props.MaxItems) }>
        
        <legend class="group-legend">
            { props.Label }
            if props.Required {
                <span class="required">*</span>
            }
        </legend>
        
        <div class="checkbox-options">
            for _, option := range props.Options {
                <div class="checkbox-option">
                    <label class="checkbox-label">
                        <input 
                            type="checkbox"
                            name={ props.Name }
                            value={ option.Value }
                            @change={ fmt.Sprintf("toggle('%s')", option.Value) }
                            :checked="selected.includes('"+ option.Value +"')"
                            class="checkbox-input"
                        />
                        <span class="checkbox-indicator"></span>
                        <div class="checkbox-content">
                            <span class="checkbox-title">{ option.Label }</span>
                            if option.Description != "" {
                                <span class="checkbox-description">{ option.Description }</span>
                            }
                        </div>
                    </label>
                </div>
            }
        </div>
        
        <div class="group-validation">
            <span x-show="!isValid" class="validation-error">
                Select between { string(props.MinItems) } and { string(props.MaxItems) } options
            </span>
            <span class="selection-count" x-text="`${selected.length} selected`"></span>
        </div>
    </fieldset>
}
```

### Indeterminate Checkbox
**Purpose**: Parent checkbox for managing child selections

```go
indeterminateCheckbox := CheckboxProps{
    Name:          "selectAll",
    Label:         "Select All Items",
    Indeterminate: true,
    OnChange:      "handleSelectAll",
}

templ IndeterminateCheckbox(props CheckboxProps) {
    <div class="checkbox-group indeterminate-checkbox" 
         x-data={ fmt.Sprintf(`{
            childrenSelected: 0,
            totalChildren: %d,
            get state() {
                if (this.childrenSelected === 0) return 'none';
                if (this.childrenSelected === this.totalChildren) return 'all';
                return 'some';
            },
            get checked() {
                return this.state === 'all';
            },
            get indeterminate() {
                return this.state === 'some';
            },
            toggleAll() {
                const newState = this.state === 'all' ? false : true;
                // Emit event to toggle all children
                $dispatch('toggle-all', { checked: newState });
            }
        }`, props.TotalChildren) }>
        
        <label class="checkbox-label">
            <input 
                type="checkbox"
                :checked="checked"
                :indeterminate="indeterminate"
                @change="toggleAll()"
                class="checkbox-input"
            />
            <span class="checkbox-indicator" 
                  :class="{
                      'indeterminate': indeterminate,
                      'checked': checked
                  }"></span>
            <span class="checkbox-text">{ props.Label }</span>
        </label>
        
        <div class="selection-summary" x-show="childrenSelected > 0">
            <span x-text="`${childrenSelected} of ${totalChildren} selected`"></span>
        </div>
    </div>
}
```

## 🎯 Props Interface

### Core Properties
```go
type CheckboxProps struct {
    // Identity
    Name     string `json:"name"`
    ID       string `json:"id"`
    TestID   string `json:"testid"`
    
    // Content
    Label       string `json:"label"`
    Value       string `json:"value"`
    HelpText    string `json:"helpText"`
    Error       string `json:"error"`
    Description string `json:"description"`
    
    // State
    Checked       bool   `json:"checked"`
    DefaultChecked bool  `json:"defaultChecked"`
    Indeterminate bool   `json:"indeterminate"`
    Required      bool   `json:"required"`
    Disabled      bool   `json:"disabled"`
    ReadOnly      bool   `json:"readonly"`
    
    // Appearance
    Size        CheckboxSize    `json:"size"`        // xs, sm, md, lg, xl
    Variant     CheckboxVariant `json:"variant"`     // default, primary, success, error
    Color       string          `json:"color"`       // Custom color
    Class       string          `json:"className"`
    Style       map[string]string `json:"style"`
    
    // Icons
    CheckedIcon   string `json:"checkedIcon"`     // Custom checked icon
    UncheckedIcon string `json:"uncheckedIcon"`   // Custom unchecked icon
    IconPosition  string `json:"iconPosition"`    // left, right
    
    // Group properties
    TotalChildren int    `json:"totalChildren"`   // For indeterminate checkboxes
    
    // Events
    OnChange    string `json:"onChange"`
    OnFocus     string `json:"onFocus"`
    OnBlur      string `json:"onBlur"`
    
    // Accessibility
    AriaLabel       string `json:"ariaLabel"`
    AriaDescribedBy string `json:"ariaDescribedBy"`
    AriaInvalid     bool   `json:"ariaInvalid"`
    TabIndex        int    `json:"tabIndex"`
}
```

### Group Properties
```go
type CheckboxGroupProps struct {
    // Identity
    Name   string `json:"name"`
    ID     string `json:"id"`
    
    // Content
    Label       string           `json:"label"`
    HelpText    string           `json:"helpText"`
    Error       string           `json:"error"`
    Options     []CheckboxOption `json:"options"`
    
    // Validation
    Required bool `json:"required"`
    MinItems int  `json:"minItems"`
    MaxItems int  `json:"maxItems"`
    
    // Layout
    Direction string `json:"direction"`  // vertical, horizontal
    Columns   int    `json:"columns"`    // Grid layout columns
    
    // Appearance
    Size    CheckboxSize    `json:"size"`
    Variant CheckboxVariant `json:"variant"`
    
    // Events
    OnChange string `json:"onChange"`
}

type CheckboxOption struct {
    Value       string `json:"value"`
    Label       string `json:"label"`
    Description string `json:"description"`
    Disabled    bool   `json:"disabled"`
    Checked     bool   `json:"checked"`
}
```

### Size Variants
```go
type CheckboxSize string

const (
    CheckboxXS CheckboxSize = "xs"    // 14px indicator
    CheckboxSM CheckboxSize = "sm"    // 16px indicator  
    CheckboxMD CheckboxSize = "md"    // 18px indicator (default)
    CheckboxLG CheckboxSize = "lg"    // 20px indicator
    CheckboxXL CheckboxSize = "xl"    // 24px indicator
)
```

### Visual Variants
```go
type CheckboxVariant string

const (
    CheckboxDefault CheckboxVariant = "default"  // Gray theme
    CheckboxPrimary CheckboxVariant = "primary"  // Brand color
    CheckboxSuccess CheckboxVariant = "success"  // Green theme
    CheckboxWarning CheckboxVariant = "warning"  // Yellow theme
    CheckboxDanger  CheckboxVariant = "danger"   // Red theme
)
```

## 🎨 Styling Implementation

### Base Checkbox Styles
```css
.checkbox-group {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
}

.checkbox-label {
    display: flex;
    align-items: flex-start;
    gap: var(--space-sm);
    cursor: pointer;
    user-select: none;
    line-height: 1.5;
    
    &:hover .checkbox-indicator {
        border-color: var(--color-border-dark);
        background: var(--color-bg-hover);
    }
}

.checkbox-input {
    position: absolute;
    opacity: 0;
    width: 0;
    height: 0;
    
    /* Focus styles */
    &:focus + .checkbox-indicator {
        outline: none;
        border-color: var(--color-primary);
        box-shadow: 0 0 0 3px var(--color-primary-light);
    }
    
    /* Checked state */
    &:checked + .checkbox-indicator {
        background: var(--color-primary);
        border-color: var(--color-primary);
        color: white;
        
        &::after {
            opacity: 1;
            transform: rotate(45deg) scale(1);
        }
    }
    
    /* Indeterminate state */
    &:indeterminate + .checkbox-indicator {
        background: var(--color-primary);
        border-color: var(--color-primary);
        
        &::after {
            opacity: 1;
            transform: scale(1);
        }
    }
    
    /* Disabled state */
    &:disabled + .checkbox-indicator {
        background: var(--color-bg-disabled);
        border-color: var(--color-border-disabled);
        cursor: not-allowed;
    }
}

.checkbox-indicator {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    border: 2px solid var(--color-border-medium);
    border-radius: var(--radius-sm);
    background: var(--color-bg-surface);
    transition: var(--transition-base);
    flex-shrink: 0;
    
    /* Checkmark */
    &::after {
        content: '';
        position: absolute;
        opacity: 0;
        transition: var(--transition-quick);
    }
    
    /* Standard checkmark */
    &:not(.indeterminate)::after {
        width: 4px;
        height: 8px;
        border: solid currentColor;
        border-width: 0 2px 2px 0;
        transform: rotate(45deg) scale(0.8);
    }
    
    /* Indeterminate dash */
    &.indeterminate::after {
        width: 8px;
        height: 2px;
        background: currentColor;
        border-radius: 1px;
        transform: scale(0.8);
    }
}

.checkbox-text {
    color: var(--color-text-primary);
    font-size: var(--font-size-base);
}

.checkbox-description {
    color: var(--color-text-secondary);
    font-size: var(--font-size-sm);
    margin-top: var(--space-xs);
}
```

### Size Variants
```css
/* Size-specific styling */
.checkbox-xs .checkbox-indicator {
    width: 14px;
    height: 14px;
    
    &::after {
        width: 3px;
        height: 6px;
        border-width: 0 1.5px 1.5px 0;
    }
    
    &.indeterminate::after {
        width: 6px;
        height: 1.5px;
    }
}

.checkbox-sm .checkbox-indicator {
    width: 16px;
    height: 16px;
    
    &::after {
        width: 3.5px;
        height: 7px;
        border-width: 0 1.5px 1.5px 0;
    }
    
    &.indeterminate::after {
        width: 7px;
        height: 1.5px;
    }
}

.checkbox-lg .checkbox-indicator {
    width: 20px;
    height: 20px;
    
    &::after {
        width: 5px;
        height: 9px;
        border-width: 0 2px 2px 0;
    }
    
    &.indeterminate::after {
        width: 9px;
        height: 2px;
    }
}

.checkbox-xl .checkbox-indicator {
    width: 24px;
    height: 24px;
    
    &::after {
        width: 6px;
        height: 11px;
        border-width: 0 2.5px 2.5px 0;
    }
    
    &.indeterminate::after {
        width: 11px;
        height: 2.5px;
    }
}
```

### Color Variants
```css
/* Primary variant */
.checkbox-primary .checkbox-input:checked + .checkbox-indicator,
.checkbox-primary .checkbox-input:indeterminate + .checkbox-indicator {
    background: var(--color-primary);
    border-color: var(--color-primary);
}

/* Success variant */
.checkbox-success .checkbox-input:checked + .checkbox-indicator,
.checkbox-success .checkbox-input:indeterminate + .checkbox-indicator {
    background: var(--color-success);
    border-color: var(--color-success);
}

/* Warning variant */
.checkbox-warning .checkbox-input:checked + .checkbox-indicator,
.checkbox-warning .checkbox-input:indeterminate + .checkbox-indicator {
    background: var(--color-warning);
    border-color: var(--color-warning);
}

/* Danger variant */
.checkbox-danger .checkbox-input:checked + .checkbox-indicator,
.checkbox-danger .checkbox-input:indeterminate + .checkbox-indicator {
    background: var(--color-danger);
    border-color: var(--color-danger);
}
```

### Group Layout Styles
```css
.checkbox-group-container {
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

.checkbox-options {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
    
    /* Horizontal layout */
    &.horizontal {
        flex-direction: row;
        flex-wrap: wrap;
    }
    
    /* Grid layout */
    &.grid {
        display: grid;
        grid-template-columns: repeat(var(--columns), 1fr);
        gap: var(--space-md);
    }
}

.checkbox-option {
    .checkbox-content {
        display: flex;
        flex-direction: column;
        gap: var(--space-xs);
    }
    
    .checkbox-title {
        font-weight: var(--font-weight-medium);
        color: var(--color-text-primary);
    }
    
    .checkbox-description {
        font-size: var(--font-size-sm);
        color: var(--color-text-secondary);
        line-height: 1.4;
    }
}

.group-validation {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: var(--space-sm);
    font-size: var(--font-size-sm);
    
    .validation-error {
        color: var(--color-error);
    }
    
    .selection-count {
        color: var(--color-text-secondary);
    }
}
```

## ⚙️ Advanced Features

### Data Binding and Validation
```go
templ ValidatedCheckboxGroup(props CheckboxGroupProps) {
    <div class="checkbox-group-validation" 
         x-data={ fmt.Sprintf(`{
            selected: %s,
            error: '',
            get isValid() {
                return this.selected.length >= %d && this.selected.length <= %d;
            },
            validate() {
                if (!this.isValid) {
                    this.error = 'Please select between %d and %d options';
                } else {
                    this.error = '';
                }
                return this.isValid;
            },
            toggle(value) {
                if (this.selected.includes(value)) {
                    this.selected = this.selected.filter(v => v !== value);
                } else {
                    this.selected.push(value);
                }
                this.validate();
            }
        }`, toJSON(props.DefaultSelected), props.MinItems, props.MaxItems, props.MinItems, props.MaxItems) }>
        
        @CheckboxGroup(props)
        
        <div x-show="error" x-text="error" class="error-message"></div>
    </div>
}
```

### Table Row Selection
```go
templ TableRowCheckbox(props TableRowCheckboxProps) {
    <td class="select-cell">
        <label class="checkbox-label table-checkbox">
            <input 
                type="checkbox"
                class="checkbox-input"
                :checked="selectedRows.includes('{ props.RowID }')"
                @change="toggleRowSelection('{ props.RowID }', $event.target.checked)"
                value={ props.RowID }
            />
            <span class="checkbox-indicator"></span>
            <span class="sr-only">Select row</span>
        </label>
    </td>
}
```

### Nested Checkbox Tree
```go
templ CheckboxTree(props CheckboxTreeProps) {
    <div class="checkbox-tree" x-data="checkboxTree()">
        for _, node := range props.Nodes {
            @CheckboxTreeNode(node)
        }
    </div>
}

templ CheckboxTreeNode(node TreeNode) {
    <div class="tree-node" x-data={ fmt.Sprintf(`{
        nodeId: '%s',
        expanded: %t,
        get allChildrenSelected() {
            return this.children.every(child => child.selected);
        },
        get someChildrenSelected() {
            return this.children.some(child => child.selected);
        },
        get indeterminate() {
            return this.someChildrenSelected && !this.allChildrenSelected;
        }
    }`, node.ID, node.Expanded) }>
        
        <div class="node-header">
            if len(node.Children) > 0 {
                <button 
                    @click="expanded = !expanded"
                    class="expand-toggle"
                    :class="{ 'expanded': expanded }">
                    @Icon(IconProps{Name: "chevron-right", Size: "sm"})
                </button>
            }
            
            <label class="checkbox-label">
                <input 
                    type="checkbox"
                    :checked="allChildrenSelected"
                    :indeterminate="indeterminate"
                    @change="toggleNodeAndChildren($event.target.checked)"
                    class="checkbox-input"
                />
                <span class="checkbox-indicator"></span>
                <span class="node-label">{ node.Label }</span>
            </label>
        </div>
        
        if len(node.Children) > 0 {
            <div x-show="expanded" class="node-children">
                for _, child := range node.Children {
                    @CheckboxTreeNode(child)
                }
            </div>
        }
    </div>
}
```

## 📱 Responsive Design

### Mobile Optimizations
```css
/* Mobile-specific checkbox styling */
@media (max-width: 479px) {
    .checkbox-label {
        min-height: 44px;  /* Touch target */
        align-items: center;
    }
    
    .checkbox-indicator {
        width: 20px;       /* Larger touch target */
        height: 20px;
    }
    
    /* Stack group options vertically on mobile */
    .checkbox-options.horizontal {
        flex-direction: column;
    }
    
    .checkbox-options.grid {
        grid-template-columns: 1fr;
    }
    
    /* Simplify descriptions on mobile */
    .checkbox-description {
        font-size: var(--font-size-xs);
    }
}
```

### Touch Interactions
```css
/* Touch-friendly interactions */
@media (hover: none) {
    .checkbox-label {
        /* Remove hover effects on touch devices */
        &:hover .checkbox-indicator {
            border-color: var(--color-border-medium);
            background: var(--color-bg-surface);
        }
    }
    
    /* Larger tap targets */
    .checkbox-indicator {
        min-width: 44px;
        min-height: 44px;
    }
}
```

## ♿ Accessibility Implementation

### ARIA Support
```go
func (checkbox CheckboxProps) GetAriaAttributes() map[string]string {
    attrs := make(map[string]string)
    
    // Required state
    if checkbox.Required {
        attrs["aria-required"] = "true"
    }
    
    // Invalid state
    if checkbox.Error != "" {
        attrs["aria-invalid"] = "true"
        if checkbox.ID != "" {
            attrs["aria-describedby"] = checkbox.ID + "-error"
        }
    }
    
    // Help text association
    if checkbox.HelpText != "" && checkbox.ID != "" {
        describedBy := attrs["aria-describedby"]
        if describedBy != "" {
            attrs["aria-describedby"] = describedBy + " " + checkbox.ID + "-help"
        } else {
            attrs["aria-describedby"] = checkbox.ID + "-help"
        }
    }
    
    // Custom label
    if checkbox.AriaLabel != "" {
        attrs["aria-label"] = checkbox.AriaLabel
    }
    
    return attrs
}
```

### Screen Reader Support
```go
templ AccessibleCheckbox(props CheckboxProps) {
    <div class="checkbox-group" role="group">
        <label for={ props.ID } class="checkbox-label">
            <input 
                id={ props.ID }
                type="checkbox"
                name={ props.Name }
                value={ props.Value }
                class="checkbox-input"
                for attrName, attrValue := range props.GetAriaAttributes() {
                    { attrName }={ attrValue }
                }
                checked?={ props.Checked }
                required?={ props.Required }
                disabled?={ props.Disabled }
            />
            <span class="checkbox-indicator" aria-hidden="true"></span>
            <span class="checkbox-text">{ props.Label }</span>
        </label>
        
        if props.HelpText != "" {
            <div id={ props.ID + "-help" } class="help-text">
                { props.HelpText }
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
    </div>
}
```

### Keyboard Navigation
```css
/* Focus management */
.checkbox-input:focus + .checkbox-indicator {
    outline: none;
    border-color: var(--color-primary);
    box-shadow: 0 0 0 3px var(--color-primary-light);
    z-index: 1;
    position: relative;
}

/* High contrast mode support */
@media (prefers-contrast: high) {
    .checkbox-indicator {
        border-width: 2px;
        border-color: CanvasText;
    }
    
    .checkbox-input:checked + .checkbox-indicator {
        background: CanvasText;
        color: Canvas;
    }
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
    .checkbox-indicator,
    .checkbox-indicator::after {
        transition: none;
    }
}
```

## 🧪 Testing Guidelines

### Unit Tests
```go
func TestCheckboxComponent(t *testing.T) {
    tests := []struct {
        name     string
        props    CheckboxProps
        expected []string
    }{
        {
            name: "basic checkbox",
            props: CheckboxProps{
                Name:  "test",
                Label: "Test Checkbox",
                Value: "test-value",
            },
            expected: []string{"checkbox-group", "checkbox-label", "Test Checkbox"},
        },
        {
            name: "required checkbox",
            props: CheckboxProps{
                Name:     "required",
                Required: true,
            },
            expected: []string{"required", "aria-required=\"true\""},
        },
        {
            name: "checked checkbox",
            props: CheckboxProps{
                Name:    "checked",
                Checked: true,
            },
            expected: []string{"checked"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            component := renderCheckbox(tt.props)
            for _, expected := range tt.expected {
                assert.Contains(t, component, expected)
            }
        })
    }
}
```

### Accessibility Tests
```javascript
describe('Checkbox Accessibility', () => {
    test('has proper label association', () => {
        const checkbox = render(<Checkbox name="test" label="Test Label" />);
        const checkboxElement = screen.getByLabelText('Test Label');
        expect(checkboxElement).toBeInTheDocument();
    });
    
    test('announces state changes to screen readers', () => {
        const checkbox = render(<Checkbox name="test" label="Test" />);
        const checkboxElement = screen.getByRole('checkbox');
        
        fireEvent.click(checkboxElement);
        expect(checkboxElement).toBeChecked();
    });
    
    test('supports keyboard navigation', () => {
        const checkbox = render(<Checkbox name="test" label="Test" />);
        const checkboxElement = screen.getByRole('checkbox');
        
        checkboxElement.focus();
        expect(checkboxElement).toHaveFocus();
        
        fireEvent.keyDown(checkboxElement, { key: ' ' });
        expect(checkboxElement).toBeChecked();
    });
});
```

### Visual Regression Tests
```javascript
test.describe('Checkbox Visual Tests', () => {
    test('all checkbox states', async ({ page }) => {
        await page.goto('/components/checkbox');
        
        // Test all sizes
        for (const size of ['xs', 'sm', 'md', 'lg', 'xl']) {
            await expect(page.locator(`[data-size="${size}"]`)).toHaveScreenshot(`checkbox-${size}.png`);
        }
        
        // Test all states
        const states = ['default', 'checked', 'indeterminate', 'disabled', 'error'];
        for (const state of states) {
            await expect(page.locator(`[data-state="${state}"]`)).toHaveScreenshot(`checkbox-${state}.png`);
        }
    });
});
```

## 📚 Usage Examples

### Simple Agreement Checkbox
```go
templ TermsAgreement() {
    @Checkbox(CheckboxProps{
        Name:     "terms",
        Label:    "I agree to the Terms of Service and Privacy Policy",
        Required: true,
        OnChange: "validateForm()",
    })
}
```

### Feature Selection Form
```go
templ FeatureSelection() {
    @CheckboxGroup(CheckboxGroupProps{
        Name:     "features",
        Label:    "Select Additional Features",
        MinItems: 1,
        Options: []CheckboxOption{
            {Value: "advanced-analytics", Label: "Advanced Analytics", Description: "Detailed reporting and insights"},
            {Value: "api-access", Label: "API Access", Description: "Programmatic access to your data"},
            {Value: "priority-support", Label: "Priority Support", Description: "24/7 dedicated support team"},
        },
    })
}
```

### Table Selection Pattern
```go
templ SelectableTable() {
    <table class="data-table">
        <thead>
            <tr>
                <th>
                    @IndeterminateCheckbox(CheckboxProps{
                        Name:          "selectAll",
                        Label:         "Select All",
                        TotalChildren: len(tableData),
                        OnChange:      "handleSelectAll",
                    })
                </th>
                <th>Name</th>
                <th>Email</th>
                <th>Status</th>
            </tr>
        </thead>
        <tbody>
            for _, row := range tableData {
                <tr>
                    <td>
                        @Checkbox(CheckboxProps{
                            Name:     "selectedRows",
                            Value:    row.ID,
                            Label:    fmt.Sprintf("Select %s", row.Name),
                            OnChange: "handleRowSelection",
                        })
                    </td>
                    <td>{ row.Name }</td>
                    <td>{ row.Email }</td>
                    <td>{ row.Status }</td>
                </tr>
            }
        </tbody>
    </table>
}
```

## 🔗 Related Components

- **[Radio Button](../radio/)**: Single selection from options
- **[Switch](../switch/)**: Toggle controls for settings
- **[Button](../button/)**: Action triggers and submissions
- **[Form Group](../../molecules/form-group/)**: Enhanced form layouts

---

**Component Status**: ✅ Production Ready  
**Schema Reference**: `CheckboxControlSchema.json`  
**CSS Classes**: `.checkbox-group`, `.checkbox-{size}`, `.checkbox-{variant}`  
**Accessibility**: WCAG 2.1 AA Compliant