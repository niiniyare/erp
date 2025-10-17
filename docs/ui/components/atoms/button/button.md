# Button Component

**FILE PURPOSE**: Primary button component implementation and usage  
**SCOPE**: Single action buttons with all variants and states  
**TARGET AUDIENCE**: Developers implementing button functionality

## 📋 Component Overview

The Button component is the fundamental interactive element for triggering actions throughout the ERP system. It provides consistent visual feedback, accessibility support, and integrates with our schema-driven architecture.

### Basic Implementation

```go
templ Button(props ButtonProps) {
    <button 
        type={ props.Type }
        class={ props.GetClasses() }
        id={ props.ID }
        disabled?={ props.Disabled }
        data-loading?={ props.Loading }
        onclick={ props.OnClick }
        title={ props.Tooltip }>
        
        if props.Loading {
            @Spinner(SpinnerProps{Size: "sm", Class: "mr-2"})
        } else if props.Icon != "" {
            @Icon(IconProps{Name: props.Icon, Size: "16", Class: "mr-2"})
        }
        
        <span class="btn-text">{ props.Text }</span>
    </button>
}
```

## 🎨 Visual Variants

### Primary Button
**Purpose**: Main call-to-action, highest visual priority  
**Usage**: Submit forms, approve documents, primary workflows

```go
// Implementation
primaryButton := ButtonProps{
    Text:    "Submit Order",
    Variant: ButtonPrimary,
    Size:    ButtonMD,
    Type:    "submit",
}

// Generated CSS
.btn-primary {
    background: var(--color-primary);      /* #4a9b9b */
    color: white;
    border: none;
    box-shadow: var(--shadow-sm);
}

.btn-primary:hover {
    background: #3d7d7d;
    transform: translateY(-1px);
    box-shadow: var(--shadow-md);
}

.btn-primary:active {
    background: #2d5555;
    transform: scale(0.98);
}

.btn-primary:disabled {
    background: var(--color-border-medium);
    color: var(--color-text-tertiary);
    cursor: not-allowed;
    transform: none;
}
```

### Secondary Button
**Purpose**: Alternative actions, lower visual priority  
**Usage**: Cancel, Edit, alternative workflows

```go
// Implementation
secondaryButton := ButtonProps{
    Text:    "Cancel Order",
    Variant: ButtonSecondary,
    Size:    ButtonMD,
}

// Generated CSS
.btn-secondary {
    background: white;
    color: var(--color-text-secondary);
    border: 1px solid var(--color-border-medium);
}

.btn-secondary:hover {
    background: #f9fafb;
    border-color: var(--color-border-medium);
}

.btn-secondary:active {
    background: #e5e7eb;
}
```

### Success Button
**Purpose**: Positive confirmations and approvals  
**Usage**: Approve, Accept, Enable, Mark Complete

```go
// Implementation
successButton := ButtonProps{
    Text:    "Approve Request",
    Variant: ButtonSuccess,
    Size:    ButtonMD,
    Icon:    "check",
}

// Generated CSS
.btn-success {
    background: var(--color-success);      /* #10b981 */
    color: white;
    border: none;
}

.btn-success:hover {
    background: #059669;
}

.btn-success:active {
    background: #047857;
}
```

### Danger Button
**Purpose**: Destructive actions requiring caution  
**Usage**: Delete, Remove, Reject, Disable

```go
// Implementation
dangerButton := ButtonProps{
    Text:    "Delete Item",
    Variant: ButtonDanger,
    Size:    ButtonMD,
    Icon:    "trash",
    OnClick: "confirmDelete(this)",
}

// Generated CSS
.btn-danger {
    background: var(--color-error);        /* #ef4444 */
    color: white;
    border: none;
}

.btn-danger:hover {
    background: #dc2626;
}

.btn-danger:active {
    background: #b91c1c;
}
```

### Ghost Button
**Purpose**: Minimal visual impact, supplementary actions  
**Usage**: View Details, Help, Toolbar actions

```go
// Implementation
ghostButton := ButtonProps{
    Text:    "View Details",
    Variant: ButtonGhost,
    Size:    ButtonSM,
    Icon:    "eye",
}

// Generated CSS
.btn-ghost {
    background: transparent;
    color: var(--color-primary);
    border: none;
}

.btn-ghost:hover {
    background: var(--color-primary-light);
}

.btn-ghost:active {
    background: rgba(74, 155, 155, 0.2);
}
```

## 📏 Size Variants

### Size System
All sizes follow the 4px grid system and include proper touch targets:

```go
type ButtonSize string

const (
    ButtonXS ButtonSize = "xs"    // 24px height, compact interfaces
    ButtonSM ButtonSize = "sm"    // 32px height, table actions
    ButtonMD ButtonSize = "md"    // 36px height, standard forms
    ButtonLG ButtonSize = "lg"    // 44px height, primary actions
    ButtonXL ButtonSize = "xl"    // 52px height, hero sections
)
```

### Implementation Examples

```go
// Extra Small - Compact interfaces
compactButton := ButtonProps{
    Text:    "Edit",
    Variant: ButtonSecondary,
    Size:    ButtonXS,
    Icon:    "edit",
}

// Small - Table row actions
tableButton := ButtonProps{
    Text:    "Delete",
    Variant: ButtonDanger,
    Size:    ButtonSM,
    Icon:    "trash",
}

// Medium - Standard forms (default)
standardButton := ButtonProps{
    Text:    "Save Changes",
    Variant: ButtonPrimary,
    Size:    ButtonMD,
}

// Large - Primary CTAs
primaryButton := ButtonProps{
    Text:    "Submit Application",
    Variant: ButtonPrimary,
    Size:    ButtonLG,
}

// Extra Large - Hero sections
heroButton := ButtonProps{
    Text:    "Get Started",
    Variant: ButtonPrimary,
    Size:    ButtonXL,
}
```

### CSS Size Implementation

```css
/* Size-specific styles */
.btn-xs {
    padding: 4px 8px;
    font-size: var(--font-size-xs);     /* 11px */
    line-height: 1.4;
    min-width: 32px;
    height: 24px;
}

.btn-sm {
    padding: 6px 12px;
    font-size: var(--font-size-sm);     /* 12px */
    line-height: 1.4;
    min-width: 56px;
    height: 32px;
}

.btn-md {
    padding: 8px 16px;
    font-size: var(--font-size-base);   /* 13px */
    line-height: 1.5;
    min-width: 64px;
    height: 36px;
}

.btn-lg {
    padding: 12px 20px;
    font-size: var(--font-size-lg);     /* 14px */
    line-height: 1.5;
    min-width: 80px;
    height: 44px;
}

.btn-xl {
    padding: 16px 24px;
    font-size: var(--font-size-xl);     /* 16px */
    line-height: 1.5;
    min-width: 96px;
    height: 52px;
}
```

## 🔄 Interactive States

### State Management
Buttons support multiple interactive states with appropriate visual feedback:

```go
type ButtonState struct {
    Default  bool  // Base state
    Hover    bool  // Mouse over (desktop only)
    Active   bool  // Pressed/clicked
    Focus    bool  // Keyboard focus
    Disabled bool  // Non-interactive
    Loading  bool  // Processing action
}
```

### Loading State Implementation

```go
templ ButtonWithLoading(props ButtonProps) {
    <button 
        type={ props.Type }
        class={ props.GetClasses() }
        disabled?={ props.Disabled || props.Loading }
        data-loading?={ props.Loading }>
        
        if props.Loading {
            <div class="btn-spinner">
                @Spinner(SpinnerProps{
                    Size:  "sm",
                    Color: "current",
                })
            </div>
            <span class="btn-text loading">Processing...</span>
        } else {
            if props.Icon != "" {
                @Icon(IconProps{Name: props.Icon, Size: "16"})
            }
            <span class="btn-text">{ props.Text }</span>
        }
    </button>
}
```

### CSS State Implementation

```css
/* Base button styles */
.btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-sm);
    font-family: var(--font-family-ui);
    font-weight: var(--font-weight-medium);
    border-radius: var(--radius-md);
    transition: var(--transition-base);
    cursor: pointer;
    text-decoration: none;
    border: none;
    position: relative;
    overflow: hidden;
}

/* Hover state (desktop only) */
@media (hover: hover) {
    .btn:hover:not(:disabled) {
        transform: translateY(-1px);
        box-shadow: var(--shadow-md);
    }
}

/* Active state */
.btn:active:not(:disabled) {
    transform: scale(0.98);
    transition: transform 100ms ease-in-out;
}

/* Focus state */
.btn:focus-visible {
    outline: 3px solid var(--color-primary);
    outline-offset: 2px;
}

/* Disabled state */
.btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none !important;
    box-shadow: none !important;
}

/* Loading state */
.btn[data-loading="true"] {
    pointer-events: none;
    
    .btn-text {
        opacity: 0.7;
    }
    
    .btn-spinner {
        margin-right: var(--space-sm);
    }
}
```

## 🎯 Event Handling

### Click Events
Support for multiple interaction patterns:

```go
type ButtonEvents struct {
    OnClick      string  // JavaScript function call
    OnSubmit     string  // Form submission handler
    OnMouseEnter string  // Hover entry
    OnMouseLeave string  // Hover exit
    OnFocus      string  // Focus handler
    OnBlur       string  // Blur handler
}

// Usage examples
buttonProps := ButtonProps{
    Text:    "Submit Form",
    OnClick: "handleSubmit(event)",
}

// HTMX integration
htmxButton := ButtonProps{
    Text:    "Load More",
    OnClick: "hx-get='/api/items' hx-target='#items-list'",
}

// Alpine.js integration  
alpineButton := ButtonProps{
    Text:    "Toggle Menu",
    OnClick: "@click='menuOpen = !menuOpen'",
}
```

### Form Integration

```go
// Submit button
submitButton := ButtonProps{
    Text:    "Submit Application",
    Type:    "submit",
    Variant: ButtonPrimary,
    Size:    ButtonLG,
}

// Reset button
resetButton := ButtonProps{
    Text:    "Reset Form",
    Type:    "reset", 
    Variant: ButtonSecondary,
    Size:    ButtonMD,
}

// Standard button
actionButton := ButtonProps{
    Text:    "Calculate Total",
    Type:    "button",
    Variant: ButtonSecondary,
    OnClick: "calculateTotal()",
}
```

## 🏗️ Advanced Features

### Icon Integration

```go
// Icon positions
type IconPosition string

const (
    IconLeft  IconPosition = "left"   // Default
    IconRight IconPosition = "right"
    IconOnly  IconPosition = "only"   // Icon without text
)

// Implementation
iconButton := ButtonProps{
    Text:         "Download Report",
    Variant:      ButtonSecondary,
    Icon:         "download",
    IconPosition: IconLeft,
}

// Icon-only button
iconOnlyButton := ButtonProps{
    Icon:         "settings",
    Variant:      ButtonGhost,
    Size:         ButtonSM,
    IconPosition: IconOnly,
    Tooltip:      "Open Settings",
}
```

### Tooltip Support

```go
type TooltipConfig struct {
    Text      string          // Tooltip content
    Position  TooltipPosition // top, right, bottom, left
    Trigger   TooltipTrigger  // hover, focus, click
    Delay     int            // Show delay in ms
}

// Usage
tooltipButton := ButtonProps{
    Text:    "?",
    Variant: ButtonGhost,
    Size:    ButtonSM,
    Tooltip: TooltipConfig{
        Text:     "This action will permanently delete the item",
        Position: TooltipTop,
        Trigger:  TooltipHover,
        Delay:    300,
    },
}
```

### Keyboard Shortcuts

```go
type KeyboardShortcut struct {
    Key         string  // Key combination
    Description string  // Help text
    Handler     string  // Function to call
}

// Implementation
shortcutButton := ButtonProps{
    Text:     "Save",
    Variant:  ButtonPrimary,
    Shortcut: KeyboardShortcut{
        Key:         "Ctrl+S",
        Description: "Save current changes",
        Handler:     "handleSave()",
    },
}
```

## 📱 Responsive Design

### Mobile Optimizations

```go
// Mobile-specific props
type MobileConfig struct {
    TouchTarget   bool   // Ensure 44px minimum
    FullWidth     bool   // Take full container width
    StackVertical bool   // Stack in button groups
    FontSize      string // Minimum 16px to prevent zoom
}

// Implementation
mobileButton := ButtonProps{
    Text:    "Submit Order",
    Variant: ButtonPrimary,
    Size:    ButtonLG,
    Mobile: MobileConfig{
        TouchTarget:   true,
        FullWidth:     true,
        FontSize:      "16px",
    },
}
```

### CSS Responsive Implementation

```css
/* Mobile adaptations */
@media (max-width: 479px) {
    .btn {
        min-height: 44px;        /* Touch target */
        font-size: 16px !important; /* Prevent zoom */
        width: 100%;             /* Full width option */
    }
    
    .btn-group {
        flex-direction: column;
        gap: var(--space-sm);
    }
    
    .btn-xl,
    .btn-lg {
        padding: 12px 16px;      /* Reduce padding on small screens */
    }
}

/* Tablet optimizations */
@media (min-width: 480px) and (max-width: 767px) {
    .btn {
        min-height: 40px;
    }
}
```

## ♿ Accessibility Implementation

### ARIA Support

```go
type AccessibilityConfig struct {
    AriaLabel       string  // Screen reader label
    AriaDescribedBy string  // Description element ID
    AriaPressed     *bool   // Toggle state
    AriaExpanded    *bool   // Expandable content
    AriaHasPopup    string  // Popup type
    Role            string  // ARIA role override
}

// Implementation
accessibleButton := ButtonProps{
    Text:    "×",
    Variant: ButtonGhost,
    Accessibility: AccessibilityConfig{
        AriaLabel:    "Close dialog",
        Role:         "button",
    },
}
```

### Keyboard Navigation

```css
/* Focus management */
.btn:focus-visible {
    outline: 3px solid var(--color-primary);
    outline-offset: 2px;
    z-index: 1;
}

/* Skip link for keyboard users */
.btn.skip-link {
    position: absolute;
    top: -40px;
    left: 6px;
    background: var(--color-bg-surface);
    color: var(--color-text-primary);
    padding: 8px;
    text-decoration: none;
    z-index: 1000;
    
    &:focus {
        top: 6px;
    }
}
```

### Screen Reader Support

```go
// Screen reader announcements
func (b ButtonProps) GetAriaAttributes() map[string]string {
    attrs := make(map[string]string)
    
    if b.Loading {
        attrs["aria-busy"] = "true"
        attrs["aria-live"] = "polite"
    }
    
    if b.Disabled {
        attrs["aria-disabled"] = "true"
    }
    
    if b.Tooltip != "" {
        attrs["aria-describedby"] = b.ID + "-tooltip"
    }
    
    return attrs
}
```

## 🧪 Testing Guidelines

### Unit Testing

```go
func TestButtonComponent(t *testing.T) {
    tests := []struct {
        name     string
        props    ButtonProps
        expected string
    }{
        {
            name: "Primary button",
            props: ButtonProps{
                Text:    "Submit",
                Variant: ButtonPrimary,
                Size:    ButtonMD,
            },
            expected: `<button type="button" class="btn btn-primary btn-md">Submit</button>`,
        },
        {
            name: "Disabled button",
            props: ButtonProps{
                Text:     "Submit",
                Disabled: true,
            },
            expected: `<button type="button" class="btn btn-disabled" disabled>Submit</button>`,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := renderButton(tt.props)
            assert.Contains(t, result, tt.expected)
        })
    }
}
```

### Accessibility Testing

```javascript
// Automated accessibility testing
describe('Button Accessibility', () => {
    test('should have proper ARIA attributes', () => {
        const button = render(<Button text="Submit" disabled={true} />);
        expect(button).toHaveAttribute('aria-disabled', 'true');
    });
    
    test('should be keyboard navigable', () => {
        const button = render(<Button text="Submit" onClick={jest.fn()} />);
        button.focus();
        expect(button).toHaveFocus();
        
        fireEvent.keyDown(button, { key: 'Enter' });
        expect(mockOnClick).toHaveBeenCalled();
    });
    
    test('should meet color contrast requirements', () => {
        // Test contrast ratios meet WCAG AA standards
        expect(getContrastRatio('#4a9b9b', '#ffffff')).toBeGreaterThan(4.5);
    });
});
```

### Visual Regression Testing

```javascript
// Visual testing with Playwright
test.describe('Button Visual Tests', () => {
    test('all button variants', async ({ page }) => {
        await page.goto('/components/button');
        await expect(page.locator('.btn-showcase')).toHaveScreenshot('button-variants.png');
    });
    
    test('button states', async ({ page }) => {
        const button = page.locator('[data-testid="primary-button"]');
        
        // Default state
        await expect(button).toHaveScreenshot('button-default.png');
        
        // Hover state
        await button.hover();
        await expect(button).toHaveScreenshot('button-hover.png');
        
        // Focus state
        await button.focus();
        await expect(button).toHaveScreenshot('button-focus.png');
    });
});
```

## 📊 Performance Considerations

### Optimization Strategies

```go
// Button memoization for frequently used buttons
var buttonCache = make(map[string]templ.Component)

func CachedButton(key string, props ButtonProps) templ.Component {
    if cached, exists := buttonCache[key]; exists {
        return cached
    }
    
    component := Button(props)
    buttonCache[key] = component
    return component
}

// Usage for static buttons
staticButton := CachedButton("submit-primary", ButtonProps{
    Text:    "Submit",
    Variant: ButtonPrimary,
})
```

### CSS Optimization

```css
/* Minimize repaints with transform instead of changing layout properties */
.btn:hover {
    transform: translateY(-1px);  /* Better than changing top/margin */
}

/* Use will-change for animated buttons */
.btn-animated {
    will-change: transform;
}

/* Optimize for hardware acceleration */
.btn:active {
    transform: scale(0.98) translateZ(0);
}
```

## 📚 Examples and Patterns

### Common Usage Patterns

```go
// Form action buttons
templ FormActions() {
    <div class="form-actions">
        @Button(ButtonProps{
            Text:    "Save Draft",
            Variant: ButtonSecondary,
            Type:    "button",
            OnClick: "saveDraft()",
        })
        
        @Button(ButtonProps{
            Text:    "Submit for Review",
            Variant: ButtonPrimary,
            Type:    "submit",
        })
    </div>
}

// Table row actions
templ TableRowActions(itemID string) {
    <div class="row-actions">
        @Button(ButtonProps{
            Text:    "Edit",
            Variant: ButtonGhost,
            Size:    ButtonSM,
            Icon:    "edit",
            OnClick: fmt.Sprintf("editItem('%s')", itemID),
        })
        
        @Button(ButtonProps{
            Text:    "Delete", 
            Variant: ButtonDanger,
            Size:    ButtonSM,
            Icon:    "trash",
            OnClick: fmt.Sprintf("deleteItem('%s')", itemID),
        })
    </div>
}

// Loading state example
templ AsyncActionButton() {
    <div x-data="{ loading: false }">
        @Button(ButtonProps{
            Text:     "Process Order",
            Variant:  ButtonPrimary,
            Loading:  true, // Bound to Alpine.js state
            OnClick:  "processOrder()",
        })
    </div>
}
```

### Integration Examples

```go
// HTMX integration
templ HTMXButton() {
    @Button(ButtonProps{
        Text:    "Load More Items",
        Variant: ButtonSecondary,
        OnClick: `hx-get="/api/items" hx-target="#items-container" hx-swap="beforeend"`,
    })
}

// Alpine.js integration
templ AlpineButton() {
    @Button(ButtonProps{
        Text:    "Toggle Settings",
        Variant: ButtonGhost,
        OnClick: "@click='settingsOpen = !settingsOpen'",
    })
}

// Form validation integration
templ ValidatedButton() {
    @Button(ButtonProps{
        Text:     "Submit Application",
        Variant:  ButtonPrimary,
        Type:     "submit",
        Disabled: true, // Bound to form validation state
        OnClick:  "validateAndSubmit()",
    })
}
```

## 🔗 Related Documentation

- **[Button Group](button-group.md)**: Multiple button management
- **[Button Toolbar](button-toolbar.md)**: Complex button layouts
- **[Form Components](../../molecules/forms/)**: Button integration in forms
- **[Design System](../../../design_system.md)**: Visual specifications
- **[Accessibility Guide](../../../fundamentals/accessibility.md)**: WCAG compliance

---

**Component Status**: ✅ Production Ready  
**Last Updated**: October 2025  
**Schema Reference**: `ButtonGroupSchema.json`, `ActionSchema.json`  
**CSS Classes**: `.btn`, `.btn-{variant}`, `.btn-{size}`, `.btn-{state}`