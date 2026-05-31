# Modal Organism

## Overview

The Modal organism provides a comprehensive dialog system for overlaying content on top of the main interface. It supports various modal types including standard modals, confirmation dialogs, and form modals with full HTMX integration.

## Component Files

- **Template**: `web/components/organisms/modal/modal.templ`
- **Types**: `web/components/organisms/modal/types.go`
- **Helpers**: `web/components/organisms/modal/helpers.go`

## Basic Usage

### Standard Modal

```go
@modal.Modal(modal.ModalProps{
    ID:   "example-modal",
    Size: modal.ModalSizeMD,
    Type: modal.ModalTypeDefault,
    Open: true,
    Header: &modal.ModalHeader{
        Title:    "Modal Title",
        Subtitle: "Optional subtitle",
        Icon:     "info",
        Closable: true,
    },
    Body: modal.ModalBody{
        Content:    "This is the modal content.",
        Scrollable: false,
    },
    Footer: &modal.ModalFooter{
        Actions: []modal.ModalAction{
            {
                Text:    "Save",
                Variant: atoms.ButtonPrimary,
                Type:    "submit",
            },
            {
                Text:    "Cancel",
                Variant: atoms.ButtonLight,
                OnClick: "closeModal()",
            },
        },
        Alignment: "right",
    },
    Closable:        true,
    CloseOnBackdrop: true,
    CloseOnEscape:   true,
})
```

### Confirmation Modal

```go
@modal.ConfirmModal(modal.ConfirmModalProps{
    Title:       "Delete Item",
    Message:     "Are you sure you want to delete this item? This action cannot be undone.",
    ConfirmText: "Delete",
    CancelText:  "Cancel",
    Variant:     "danger",
    Icon:        "trash",
    OnConfirm:   "deleteItem()",
    OnCancel:    "console.log('Cancelled')",
})
```

### Form Modal

```go
@modal.FormModal(modal.FormModalProps{
    Title:      "Create New User",
    FormID:     "user-form",
    HxPost:     "/api/users",
    HxTarget:   "#user-table",
    HxSwap:     "outerHTML",
    SubmitText: "Create User",
    CancelText: "Cancel",
    Size:       modal.ModalSizeLG,
    Validate:   true,
    Content:    userFormHTML,
})
```

## Configuration

### Modal Sizes

```go
type ModalSize string

const (
    ModalSizeSM   ModalSize = "sm"   // Small: 320px
    ModalSizeMD   ModalSize = "md"   // Medium: 512px (default)
    ModalSizeLG   ModalSize = "lg"   // Large: 640px
    ModalSizeXL   ModalSize = "xl"   // Extra Large: 768px
    ModalSize2XL  ModalSize = "2xl"  // 2X Large: 896px
    ModalSizeFull ModalSize = "full" // Full screen
)
```

### Modal Types

```go
type ModalType string

const (
    ModalTypeDefault ModalType = "default" // Standard modal
    ModalTypeForm    ModalType = "form"    // Form modal with validation
    ModalTypeConfirm ModalType = "confirm" // Confirmation dialog
    ModalTypeAlert   ModalType = "alert"   // Alert/info modal
)
```

### Header Configuration

```go
type ModalHeader struct {
    Title         string // Main title
    Subtitle      string // Optional subtitle
    Icon          string // Icon name
    Closable      bool   // Show close button
    Level         int    // Heading level (1-6)
    CustomContent string // Raw HTML content
}
```

### Body Configuration

```go
type ModalBody struct {
    Content    string // Text content
    HTML       string // Raw HTML content
    Scrollable bool   // Enable scrolling
    Padding    string // Custom padding classes
    FormID     string // Form element ID
    Loading    bool   // Show loading state
}
```

### Footer Configuration

```go
type ModalFooter struct {
    Actions       []ModalAction // Action buttons
    Alignment     string        // left, center, right, between
    Compact       bool          // Reduce spacing
    CustomContent string        // Raw HTML content
}
```

## HTMX Integration

### Dynamic Content Loading

```go
@modal.Modal(modal.ModalProps{
    ID:        "dynamic-modal",
    HxGet:     "/api/modal-content",
    HxTarget:  "this",
    HxSwap:    "innerHTML",
    HxTrigger: "shown",
})
```

### Form Submission

```go
@modal.FormModal(modal.FormModalProps{
    Title:    "Edit Profile",
    HxPost:   "/api/profile/update",
    HxTarget: "#profile-section",
    HxSwap:   "outerHTML",
})
```

## Alpine.js Integration

### State Management

The modal includes built-in Alpine.js data and methods:

```javascript
// Generated Alpine.js data
{
    open: false,
    loading: false,
    
    // Methods
    openModal() {
        this.open = true;
        // Focus management
        // Callback execution
    },
    
    closeModal() {
        this.open = false;
        // Cleanup
    },
    
    handleEscapeKey() {
        if (this.closeOnEscape) {
            this.closeModal();
        }
    },
    
    handleBackdropClick() {
        if (this.closeOnBackdrop) {
            this.closeModal();
        }
    }
}
```

### Custom Callbacks

```go
@modal.Modal(modal.ModalProps{
    OnOpen:  "console.log('Modal opened')",
    OnClose: "console.log('Modal closed')",
})
```

## Styling Options

### Backdrop Variants

```go
props.Backdrop = "blur"   // Blurred backdrop
props.Backdrop = "dark"   // Dark backdrop
props.Backdrop = "light"  // Light backdrop
```

### Positioning

```go
props.Position = "center" // Center (default)
props.Position = "top"    // Top aligned
props.Position = "bottom" // Bottom aligned
```

### Animations

```go
props.Animation = "fade"  // Fade in/out
props.Animation = "scale" // Scale in/out
props.Animation = "slide" // Slide in/out
```

## Accessibility Features

### ARIA Support

```go
@modal.Modal(modal.ModalProps{
    AriaLabel:       "User settings modal",
    AriaDescribedBy: "modal-description",
    Role:            "dialog",
})
```

### Focus Management

- Automatic focus trapping within modal
- Focus returns to trigger element on close
- ESC key support for closing
- Tab navigation within modal content

### Screen Reader Support

- Proper ARIA roles and labels
- Hidden backdrop from screen readers
- Descriptive button labels
- Semantic heading structure

## Advanced Examples

### Multi-Step Form Modal

```go
@modal.Modal(modal.ModalProps{
    ID:   "wizard-modal",
    Size: modal.ModalSizeXL,
    Header: &modal.ModalHeader{
        Title:    "Setup Wizard",
        Closable: false, // Prevent closing during setup
    },
    Body: modal.ModalBody{
        HTML:       wizardStepsHTML,
        Scrollable: true,
    },
    Footer: &modal.ModalFooter{
        Actions: []modal.ModalAction{
            {
                Text:    "Previous",
                Variant: atoms.ButtonLight,
                OnClick: "previousStep()",
            },
            {
                Text:    "Next",
                Variant: atoms.ButtonPrimary,
                OnClick: "nextStep()",
            },
        },
        Alignment: "between",
    },
})
```

### Confirmation with Custom Icon

```go
@modal.ConfirmModal(modal.ConfirmModalProps{
    Title:       "Publish Article",
    Message:     "Are you ready to publish this article? It will be visible to all users.",
    ConfirmText: "Publish",
    CancelText:  "Keep Draft",
    Variant:     "info",
    Icon:        "globe",
    OnConfirm:   "publishArticle()",
})
```

### Loading Modal with HTMX

```go
@modal.Modal(modal.ModalProps{
    ID:     "loading-modal",
    HxGet:  "/api/slow-content",
    Body: modal.ModalBody{
        Loading: true,
        LoadingText: "Fetching data...",
    },
})
```

## Best Practices

### Content Guidelines

1. **Keep titles concise** - Use clear, descriptive titles
2. **Provide context** - Use subtitles for additional information
3. **Progressive disclosure** - Don't overwhelm with too much content
4. **Clear actions** - Make primary and secondary actions obvious

### UX Considerations

1. **Appropriate sizing** - Choose modal size based on content needs
2. **Escape routes** - Always provide a way to close the modal
3. **Loading states** - Show feedback for async operations
4. **Mobile responsive** - Test on various screen sizes

### Performance

1. **Lazy loading** - Use HTMX for dynamic content
2. **Conditional rendering** - Only render when needed
3. **Cleanup** - Properly dispose of event listeners
4. **Keyboard navigation** - Ensure full keyboard accessibility

### Security

1. **Content sanitization** - Sanitize any user-provided HTML
2. **CSRF protection** - Include CSRF tokens in forms
3. **Input validation** - Validate all form inputs
4. **XSS prevention** - Escape dynamic content properly

## Component Hierarchy

```
Modal (Organism)
├── Backdrop (Element)
├── Container (Element)
│   ├── Header (Section)
│   │   ├── Icon (Atom)
│   │   ├── Title/Subtitle (Text)
│   │   └── Close Button (Atom)
│   ├── Body (Section)
│   │   └── Content (Mixed)
│   └── Footer (Section)
│       └── Actions (Button Atoms)
└── Loading State (Molecule)
    ├── Spinner (Atom)
    └── Loading Text (Text)
```

This modal organism provides a complete dialog solution with excellent accessibility, HTMX integration, and Alpine.js state management, following modern web development best practices.