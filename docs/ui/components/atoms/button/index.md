# Button Components

**FILE PURPOSE**: Complete button component system documentation  
**SCOPE**: All button variants, states, and usage patterns  
**TARGET AUDIENCE**: Developers, designers, component implementers

## 📋 Overview

Button components are fundamental interactive elements that trigger actions in the ERP system. They follow our design system principles of clarity, consistency, and accessibility while supporting all necessary business workflows.

### Available Button Types

| Component | Purpose | Documentation |
|-----------|---------|---------------|
| **[Button](button.md)** | Standard action buttons with variants and states | Primary component |
| **[Button Group](button-group.md)** | Related button collections and toggles | Group management |
| **[Button Toolbar](button-toolbar.md)** | Action toolbars and complex layouts | Advanced layouts |

## 🎨 Design Principles

### Visual Hierarchy
```
Primary → Secondary → Success → Danger → Ghost
```

- **Primary**: Main call-to-action (Submit, Save, Approve)
- **Secondary**: Alternative actions (Cancel, Reset, Edit)
- **Success**: Positive confirmations (Approve, Accept, Enable)
- **Danger**: Destructive actions (Delete, Reject, Remove)
- **Ghost**: Minimal impact actions (View, Details, Help)

### Size System
```
XS (24px) → SM (32px) → MD (36px) → LG (44px) → XL (52px)
```

## 🏗️ Architecture Integration

### Schema-Driven Generation
```go
// Button props structure
type ButtonProps struct {
    Text     string        `json:"text"`
    Type     ButtonType    `json:"type"`        // button, submit, reset
    Variant  ButtonVariant `json:"variant"`     // primary, secondary, success, danger, ghost
    Size     ButtonSize    `json:"size"`        // xs, sm, md, lg, xl
    Disabled bool          `json:"disabled"`
    Loading  bool          `json:"loading"`
    Icon     string        `json:"icon"`
    ID       string        `json:"id"`
    Class    string        `json:"className"`
    OnClick  string        `json:"onClick"`
    Tooltip  string        `json:"tooltip"`
}
```

### Component Factory Integration
```go
// Schema factory pattern
factory.RegisterRenderer("ButtonSchema", &ButtonRenderer{})
factory.RegisterRenderer("ButtonGroupSchema", &ButtonGroupRenderer{})

// Usage
component, _ := factory.RenderToTempl(ctx, "ButtonSchema", buttonProps)
```

## 🚀 Quick Start

### Basic Button
```go
templ BasicButton() {
    @Button(ButtonProps{
        Text:    "Submit Order",
        Variant: ButtonPrimary,
        Size:    ButtonMD,
        Type:    "submit",
    })
}
```

### Complete Example
```go
templ ActionButtons() {
    <div class="flex gap-3">
        @Button(ButtonProps{
            Text:    "Save Draft",
            Variant: ButtonSecondary,
            Size:    ButtonMD,
            Icon:    "save",
        })
        
        @Button(ButtonProps{
            Text:     "Submit for Approval",
            Variant:  ButtonPrimary,
            Size:     ButtonMD,
            Type:     "submit",
            OnClick:  "submitForm()",
        })
        
        @Button(ButtonProps{
            Text:     "Delete",
            Variant:  ButtonDanger,
            Size:     ButtonMD,
            Icon:     "trash",
            OnClick:  "confirmDelete()",
        })
    </div>
}
```

## 📐 Design Specifications

### Visual Properties
Based on **design_system.md** specifications:

```css
/* Size Variants */
.btn-xs { padding: 4px 8px; font-size: 11px; min-width: 32px; }
.btn-sm { padding: 6px 12px; font-size: 12px; min-width: 56px; }
.btn-md { padding: 8px 16px; font-size: 13px; min-width: 64px; }
.btn-lg { padding: 12px 20px; font-size: 14px; min-width: 80px; }
.btn-xl { padding: 16px 24px; font-size: 16px; min-width: 96px; }

/* Variant Colors */
.btn-primary { background: #4a9b9b; color: white; }
.btn-secondary { background: white; color: #374151; border: 1px solid #d1d5db; }
.btn-success { background: #10b981; color: white; }
.btn-danger { background: #ef4444; color: white; }
.btn-ghost { background: transparent; color: #4a9b9b; }
```

### State Management
- **Default**: Base styling
- **Hover**: Subtle elevation and color shift
- **Active**: Scale down 0.98, visual pressed state
- **Focus**: 3px outline with 2px offset
- **Disabled**: Reduced opacity, not-allowed cursor
- **Loading**: Spinner replaces content, disabled state

## 🔧 Technical Implementation

### JSON Schema Validation
All button components validate against these schemas:
- `ButtonGroupSchema.json` - Button group collections
- `ActionSchema.json` - Individual button actions

### CSS Class Generation
```go
func (b ButtonProps) GetClasses() string {
    classes := []string{"btn"}
    
    // Size classes
    classes = append(classes, fmt.Sprintf("btn-%s", b.Size))
    
    // Variant classes  
    classes = append(classes, fmt.Sprintf("btn-%s", b.Variant))
    
    // State classes
    if b.Disabled {
        classes = append(classes, "btn-disabled")
    }
    if b.Loading {
        classes = append(classes, "btn-loading")
    }
    
    // Custom classes
    if b.Class != "" {
        classes = append(classes, b.Class)
    }
    
    return strings.Join(classes, " ")
}
```

### Accessibility Features
- **ARIA Labels**: Automatic labeling for screen readers
- **Keyboard Support**: Full keyboard navigation
- **Focus Management**: Visible focus indicators
- **Screen Reader**: Action announcements

## 📱 Responsive Behavior

### Mobile Adaptations
- **Touch Targets**: Minimum 44px touch area
- **Font Size**: Minimum 16px to prevent zoom
- **Spacing**: Increased padding for touch interaction
- **Layout**: Stack vertically on small screens

### Breakpoint Behavior
```css
/* Mobile: Stack buttons */
@media (max-width: 479px) {
    .btn-group { flex-direction: column; gap: 8px; }
    .btn { width: 100%; min-height: 44px; }
}

/* Tablet: Maintain horizontal layout */
@media (min-width: 480px) {
    .btn-group { flex-direction: row; gap: 12px; }
}
```

## 🎯 Usage Guidelines

### When to Use Each Variant

**Primary Button**
- ✅ Main call-to-action per page/section
- ✅ Form submissions (Submit, Save, Approve)
- ✅ Progressive actions (Next, Continue, Proceed)
- ❌ More than one per visible area

**Secondary Button**
- ✅ Alternative actions (Cancel, Reset, Edit)
- ✅ Navigation (Back, Close, View Details)
- ✅ Secondary workflows

**Success Button**
- ✅ Positive confirmations (Approve, Accept, Enable)
- ✅ Successful state actions (Mark Complete)

**Danger Button**
- ✅ Destructive actions (Delete, Remove, Reject)
- ✅ Irreversible operations
- ✅ Always pair with confirmation

**Ghost Button**
- ✅ Minimal impact actions (View, Details, Help)
- ✅ Toolbar actions
- ✅ Supplementary functionality

### Content Guidelines

**Button Text**
- Use action verbs (Save, Submit, Delete, View)
- Keep text concise (1-3 words maximum)
- Use sentence case (Save draft, not Save Draft)
- Be specific (Submit order, not Submit)

**Icons**
- Support text, don't replace it
- Use consistent icon library
- Maintain 16px minimum size
- Align with text baseline

## 🧪 Testing Requirements

### Visual Testing
- [ ] All variants render correctly
- [ ] All sizes display properly
- [ ] State changes work (hover, active, disabled)
- [ ] Dark mode compatibility
- [ ] Mobile responsive behavior

### Accessibility Testing
- [ ] Keyboard navigation functional
- [ ] Screen reader announcements
- [ ] Focus indicators visible
- [ ] Color contrast meets AA standards
- [ ] Touch targets minimum 44px

### Functional Testing
- [ ] Click events trigger correctly
- [ ] Form submission works
- [ ] Loading states display
- [ ] Disabled states prevent interaction
- [ ] Tooltips appear on hover

## 📚 Related Components

### Form Integration
- **[Form Controls](../../molecules/forms/)**: Button usage in forms
- **[Validation](../../organisms/forms/)**: Submit button validation states

### Layout Integration
- **[Toolbars](../../molecules/toolbars/)**: Button toolbar layouts
- **[Navigation](../../organisms/navigation/)**: Action button placement

### Advanced Patterns
- **[Modals](../../organisms/modals/)**: Modal action buttons
- **[Tables](../../organisms/tables/)**: Row action buttons

## 🔗 External References

- **Design System**: `/docs/ui/design_system.md`
- **Schema Definitions**: `/docs/ui/Schema/definitions/components/atoms/`
- **Styling Guide**: `/docs/ui/fundamentals/styling-approach.md`
- **Component Lifecycle**: `/docs/ui/fundamentals/component-lifecycle.md`

---

**Last Updated**: October 2025  
**Maintainer**: UI Component Team  
**Status**: Active Development