# UI Elements Reference

<!-- LLM-CONTEXT-START -->
**FILE PURPOSE**: Comprehensive reference for foundational UI elements (buttons, inputs, cards, etc.)
**DEPENDENCIES**: Templ + Flowbite + TailwindCSS, optional Alpine.js
**TARGET AUDIENCE**: Developers building atomic UI components
<!-- **RELATED FILES**: `templ-llms.md` (advanced features), `flowbite-llms-full.txt` (styling reference) -->
<!-- LLM-CONTEXT-END -->

## Overview

This reference documents foundational UI elements built with Templ, styled with Flowbite, and optionally enhanced with Alpine.js. These atomic components serve as building blocks for complex applications.

<!-- LLM-TECH-STACK-START -->
**TECHNOLOGY STACK:**
- **[Templ](https://templ.guide)** - Type-safe Go templating language (Required)
- **[Flowbite](https://flowbite.com)** - UI component library based on TailwindCSS (Required) 
- **[Alpine.js](https://alpinejs.dev)** - Lightweight reactive JavaScript framework (Optional)
- **[TailwindCSS](https://tailwindcss.com)** - Utility-first CSS framework (Included with Flowbite)
<!-- LLM-TECH-STACK-END -->

## Design Principles

<!-- LLM-PRINCIPLES-START -->
**COMPONENT DESIGN PRINCIPLES:**
1. **Atomic Design**: Each element serves a single, well-defined purpose
2. **Composability**: Elements combine to create complex components
3. **Accessibility**: Built-in WCAG 2.1 AA compliance
4. **Progressive Enhancement**: Core functionality works without JavaScript
5. **Type Safety**: Leverages Go's type system through Templ
6. **Consistency**: Unified design system via Flowbite
<!-- LLM-PRINCIPLES-END -->

## Props Pattern

All elements follow a consistent props structure:

<!-- LLM-PROPS-PATTERN-START -->
```go
type ElementProps struct {
    // Content Properties
    Text        string    // Display text or content
    Value       string    // Form field value
    Placeholder string    // Input placeholder text
    
    // Styling Properties  
    Variant     string    // Visual style ("primary", "secondary", etc.)
    Size        string    // Size modifier ("sm", "base", "lg")
    Color       string    // Color theme ("blue", "red", "green")
    
    // Behavior Properties
    Disabled    bool      // Disable user interaction
    Loading     bool      // Show loading state
    Required    bool      // Form validation requirement
    OnClick     string    // JavaScript click handler
    OnChange    string    // JavaScript change handler
    
    // HTML Properties
    ID          string    // HTML id attribute
    Class       string    // Additional CSS classes
    Name        string    // Form field name
    AriaLabel   string    // Accessibility label
}
```
<!-- LLM-PROPS-PATTERN-END -->

---

## Button Element

<!-- LLM-BUTTON-START -->
**PURPOSE**: Interactive element for user actions and form submissions

**FUNCTION SIGNATURE:**
```go
templ Button(props ButtonProps) {}
```

**PROPS SPECIFICATION:**
```go
type ButtonProps struct {
    Text        string  // Button text content
    Variant     string  // "primary", "secondary", "success", "danger", "warning", "info", "light", "dark"
    Size        string  // "xs", "sm", "base", "lg", "xl"
    Type        string  // "button", "submit", "reset"
    Disabled    bool    // Disable interaction
    Loading     bool    // Show spinner, disable interaction
    FullWidth   bool    // Expand to container width
    OnClick     string  // JavaScript click handler
    ID          string  // HTML id attribute
    Class       string  // Additional CSS classes
    AriaLabel   string  // Accessibility label
}
```

**USAGE EXAMPLES:**
```go
// Primary action button
Button(ButtonProps{
    Text: "Save Changes",
    Variant: "primary",
    Type: "submit",
})

// Loading state button
Button(ButtonProps{
    Text: "Processing...",
    Variant: "primary", 
    Loading: true,
    Disabled: true,
})

// Destructive action button
Button(ButtonProps{
    Text: "Delete Account",
    Variant: "danger",
    OnClick: "confirmDelete()",
    AriaLabel: "Delete your account permanently",
})

// Full-width mobile button
Button(ButtonProps{
    Text: "Continue",
    Variant: "primary",
    Size: "lg",
    FullWidth: true,
})
```

**ACCESSIBILITY FEATURES:**
- Automatic focus states and keyboard navigation
- Loading state disables button to prevent double-submission
- ARIA attributes for screen readers
- Semantic button types for form integration

**DEPENDENCIES:**
- Required: Flowbite CSS framework
- Optional: Alpine.js for enhanced interactivity
<!-- LLM-BUTTON-END -->

---

## Input Element

<!-- LLM-INPUT-START -->
**PURPOSE**: Single-line text input with validation and accessibility features

**FUNCTION SIGNATURE:**
```go
templ Input(props InputProps) {}
```

**PROPS SPECIFICATION:**
```go
type InputProps struct {
    Type         string  // "text", "email", "password", "number", "tel", "url", "search"
    Value        string  // Input field value
    Placeholder  string  // Placeholder text
    Name         string  // Form field name
    Label        string  // Associated label text
    HelperText   string  // Help or error message below input
    Size         string  // "sm", "base", "lg"
    Validation   string  // "none", "success", "error"
    Required     bool    // HTML5 required attribute
    Disabled     bool    // Disable input
    ReadOnly     bool    // Make read-only
    MaxLength    int     // Character limit
    ID           string  // HTML id attribute
    Class        string  // Additional CSS classes
    OnChange     string  // JavaScript change handler
    OnFocus      string  // JavaScript focus handler
    OnBlur       string  // JavaScript blur handler
}
```

**USAGE EXAMPLES:**
```go
// Basic text input with label
Input(InputProps{
    Label: "Full Name",
    Type: "text",
    Name: "fullName",
    Required: true,
    Placeholder: "Enter your full name",
})

// Email input with validation
Input(InputProps{
    Label: "Email Address",
    Type: "email", 
    Name: "email",
    Validation: "error",
    HelperText: "Please enter a valid email address",
    Required: true,
})

// Search input with real-time filtering
Input(InputProps{
    Type: "search",
    Placeholder: "Search products...",
    OnChange: "filterProducts(this.value)",
    Size: "lg",
})

// Password input with character limit
Input(InputProps{
    Label: "Password",
    Type: "password",
    Name: "password",
    Required: true,
    MaxLength: 50,
    HelperText: "Must be 8+ characters",
})
```

**ACCESSIBILITY FEATURES:**
- Proper label association using `for` attribute
- Error states communicated to screen readers
- HTML5 validation attributes
- Focus management and keyboard navigation

**DEPENDENCIES:**
- Required: Flowbite CSS framework
- Optional: Alpine.js for real-time validation
<!-- LLM-INPUT-END -->

---

## Card Element

<!-- LLM-CARD-START -->
**PURPOSE**: Flexible container component for grouping related content

**FUNCTION SIGNATURE:**
```go
templ Card(props CardProps) {}
```

**PROPS SPECIFICATION:**
```go
type CardProps struct {
    Title       string           // Header title
    Subtitle    string           // Header subtitle
    Content     templ.Component  // Main content area
    Footer      templ.Component  // Footer content
    Image       string           // Header image URL
    ImageAlt    string           // Image alt text
    Bordered    bool             // Show border (default: true)
    Shadow      string           // "none", "sm", "base", "md", "lg", "xl"
    Padding     string           // "none", "sm", "base", "lg", "xl"
    Clickable   bool             // Make entire card clickable
    OnClick     string           // JavaScript click handler
    MaxWidth    string           // Maximum width constraint
    ID          string           // HTML id attribute
    Class       string           // Additional CSS classes
}
```

**USAGE EXAMPLES:**
```go
// Product card with image and actions
Card(CardProps{
    Title: "Premium Wireless Headphones",
    Subtitle: "$299.99",
    Image: "/images/headphones.jpg",
    ImageAlt: "Premium wireless headphones in black",
    Content: productDescription(),
    Footer: productActions(),
    Shadow: "lg",
    Clickable: true,
    OnClick: "viewProduct('headphones-123')",
})

// Simple content card
Card(CardProps{
    Title: "Getting Started",
    Content: helpContent(),
    Bordered: true,
    Padding: "lg",
    MaxWidth: "md",
})

// Minimal card without border
Card(CardProps{
    Content: notificationContent(),
    Shadow: "sm",
    Bordered: false,
    Padding: "sm",
})
```

**COMPOSITION PATTERN:**
Cards follow a hierarchical content structure: optional image → title/subtitle header → main content area → optional footer. Each section can be independently controlled through props.

**DEPENDENCIES:**
- Required: Flowbite CSS framework
- Optional: Alpine.js for dynamic content updates
<!-- LLM-CARD-END -->

---

## Alert Element

<!-- LLM-ALERT-START -->
**PURPOSE**: Display important messages and notifications with semantic styling

**FUNCTION SIGNATURE:**
```go
templ Alert(props AlertProps) {}
```

**PROPS SPECIFICATION:**
```go
type AlertProps struct {
    Message     string           // Alert message text
    Title       string           // Optional alert title
    Variant     string           // "info", "success", "warning", "error"
    Dismissible bool             // Show close button
    Icon        bool             // Show status icon (default: true)
    Border      bool             // Show colored border
    Actions     templ.Component  // Custom action buttons
    OnDismiss   string           // JavaScript dismiss handler
    ID          string           // HTML id attribute
    Class       string           // Additional CSS classes
}
```

**USAGE EXAMPLES:**
```go
// Success notification
Alert(AlertProps{
    Title: "Success!",
    Message: "Your changes have been saved successfully.",
    Variant: "success",
    Dismissible: true,
    Icon: true,
})

// Error alert with retry action
Alert(AlertProps{
    Title: "Connection Failed",
    Message: "Unable to save your changes. Please check your connection and try again.",
    Variant: "error",
    Actions: retryButton(),
    OnDismiss: "hideAlert('connection-error')",
})

// Warning with border styling
Alert(AlertProps{
    Message: "Your session will expire in 5 minutes.",
    Variant: "warning",
    Border: true,
    Dismissible: true,
})

// Info alert without icon
Alert(AlertProps{
    Title: "New Feature Available",
    Message: "Check out our new dashboard design in settings.",
    Variant: "info",
    Icon: false,
    Actions: learnMoreButton(),
})
```

**SEMANTIC COLOR SYSTEM:**
- **Info**: Blue styling for informational messages
- **Success**: Green styling for positive feedback
- **Warning**: Yellow styling for cautionary messages  
- **Error**: Red styling for error states

**DEPENDENCIES:**
- Required: Flowbite CSS framework
- Required: Flowbite JavaScript (for dismissible functionality)
- Optional: Alpine.js for state management
<!-- LLM-ALERT-END -->

---

## Modal Element

<!-- LLM-MODAL-START -->
**PURPOSE**: Overlay dialog component for focused content and user interactions

**FUNCTION SIGNATURE:**
```go
templ Modal(props ModalProps) {}
```

**PROPS SPECIFICATION:**
```go
type ModalProps struct {
    Title       string           // Header title
    Content     templ.Component  // Modal body content
    Footer      templ.Component  // Action buttons area
    Size        string           // "xs", "sm", "default", "lg", "xl", "2xl"
    Position    string           // "center", "top"
    Backdrop    string           // "default", "static", "none"
    Keyboard    bool             // ESC key closes modal (default: true)
    Persistent  bool             // Prevent closing (overrides backdrop/keyboard)
    ID          string           // Required for targeting
    Class       string           // Additional CSS classes
}
```

**USAGE EXAMPLES:**
```go
// Confirmation dialog
Modal(ModalProps{
    ID: "delete-confirmation",
    Title: "Delete Account",
    Content: deleteWarningContent(),
    Footer: confirmationButtons(),
    Size: "sm",
    Backdrop: "static",  // Prevent clicking outside to close
})

// Large content modal
Modal(ModalProps{
    ID: "terms-modal", 
    Title: "Terms of Service",
    Content: termsContent(),
    Size: "xl",
    Position: "center",
    Keyboard: true,
})

// Form modal
Modal(ModalProps{
    ID: "edit-profile",
    Title: "Edit Profile",
    Content: profileForm(),
    Footer: formActions(),
    Size: "lg",
})
```

**INTERACTION BEHAVIOR:**
- **Backdrop**: Controls modal dismissal behavior when clicking outside
- **Keyboard**: ESC key support for accessibility
- **Focus Management**: Automatically traps focus within modal content
- **Persistent**: Prevents all closing methods for critical dialogs

**DEPENDENCIES:**
- Required: Flowbite CSS framework
- Required: Flowbite JavaScript (for modal functionality and focus management)
- Optional: Alpine.js for state management
<!-- LLM-MODAL-END -->

---

## Form Elements

### Select Element

<!-- LLM-SELECT-START -->
**PURPOSE**: Dropdown selection component with single or multiple selection

**FUNCTION SIGNATURE:**
```go
templ Select(props SelectProps) {}
```

**PROPS SPECIFICATION:**
```go
type SelectOption struct {
    Value    string  // Option value
    Label    string  // Display text
    Selected bool    // Pre-selected state
    Disabled bool    // Disable option
}

type SelectProps struct {
    Options     []SelectOption  // Available options
    Value       string          // Selected value
    Placeholder string          // Placeholder text
    Name        string          // Form field name
    Label       string          // Associated label
    Multiple    bool            // Allow multiple selection
    Searchable  bool            // Enable search filtering (requires Alpine.js)
    Size        string          // "sm", "base", "lg"
    Validation  string          // "none", "success", "error"
    Disabled    bool            // Disable select
    Required    bool            // HTML5 required attribute
    HelperText  string          // Help or error text
    ID          string          // HTML id attribute
    Class       string          // Additional CSS classes
    OnChange    string          // JavaScript change handler
}
```

**USAGE EXAMPLES:**
```go
// Basic single selection
Select(SelectProps{
    Label: "Country",
    Name: "country",
    Placeholder: "Select your country",
    Options: countryOptions,
    Required: true,
})

// Multiple selection with search
Select(SelectProps{
    Label: "Skills",
    Name: "skills",
    Multiple: true,
    Searchable: true,
    Options: skillOptions,
    HelperText: "Select all applicable skills",
})
```
<!-- LLM-SELECT-END -->

### Checkbox Element

<!-- LLM-CHECKBOX-START -->
**PURPOSE**: Binary or multi-state selection component

**FUNCTION SIGNATURE:**
```go
templ Checkbox(props CheckboxProps) {}
```

**PROPS SPECIFICATION:**
```go
type CheckboxProps struct {
    Checked       bool    // Whether checkbox is checked
    Label         string  // Checkbox label text
    Name          string  // Form field name
    Value         string  // Checkbox value
    Disabled      bool    // Whether checkbox is disabled
    Indeterminate bool    // Three-state checkbox (checked/unchecked/indeterminate)
    Color         string  // "blue", "red", "green", "yellow", "purple"
    Size          string  // "sm", "base", "lg"
    HelperText    string  // Helper text below checkbox
    Required      bool    // Form validation requirement
    ID            string  // HTML id attribute
    Class         string  // Additional CSS classes
    OnChange      string  // JavaScript change handler
}
```

**USAGE EXAMPLES:**
```go
// Basic checkbox
Checkbox(CheckboxProps{
    Label: "I agree to the terms and conditions",
    Name: "terms",
    Value: "accepted",
    Required: true,
})

// Indeterminate checkbox for "select all"
Checkbox(CheckboxProps{
    Label: "Select All",
    Name: "selectAll",
    Indeterminate: true,
    OnChange: "toggleSelectAll()",
})
```
<!-- LLM-CHECKBOX-END -->

### Textarea Element

<!-- LLM-TEXTAREA-START -->
**PURPOSE**: Multi-line text input for longer content

**FUNCTION SIGNATURE:**
```go
templ Textarea(props TextareaProps) {}
```

**PROPS SPECIFICATION:**
```go
type TextareaProps struct {
    Value       string  // Textarea content
    Placeholder string  // Placeholder text
    Name        string  // Form field name
    Label       string  // Associated label
    Rows        int     // Number of visible rows (default: 4)
    Resize      string  // "none", "both", "horizontal", "vertical"
    MaxLength   int     // Character limit
    Validation  string  // "none", "success", "error"
    HelperText  string  // Help or error text
    Disabled    bool    // Disable textarea
    ReadOnly    bool    // Make read-only
    Required    bool    // HTML5 required attribute
    ID          string  // HTML id attribute
    Class       string  // Additional CSS classes
    OnChange    string  // JavaScript change handler
}
```

**USAGE EXAMPLES:**
```go
// Basic textarea
Textarea(TextareaProps{
    Label: "Message",
    Name: "message",
    Placeholder: "Enter your message...",
    Rows: 4,
    Required: true,
})

// Character-limited textarea
Textarea(TextareaProps{
    Label: "Bio",
    Name: "bio",
    MaxLength: 500,
    HelperText: "Describe yourself in 500 characters or less",
    Resize: "vertical",
})
```
<!-- LLM-TEXTAREA-END -->

---

## Typography Elements

### Heading Element

<!-- LLM-HEADING-START -->
**PURPOSE**: Semantic heading component with customizable styling

**FUNCTION SIGNATURE:**
```go
templ Heading(props HeadingProps) {}
```

**PROPS SPECIFICATION:**
```go
type HeadingProps struct {
    Level   int     // Heading level (1-6)
    Text    string  // Heading text content
    Size    string  // "auto", "xs", "sm", "base", "lg", "xl", "2xl", "3xl", "4xl", "5xl", "6xl", "7xl", "8xl", "9xl"
    Weight  string  // "auto", "thin", "light", "normal", "medium", "semibold", "bold", "extrabold", "black"
    Color   string  // Text color variant
    Align   string  // "left", "center", "right", "justify"
    ID      string  // HTML id attribute
    Class   string  // Additional CSS classes
}
```

**USAGE EXAMPLES:**
```go
// Page title (h1)
Heading(HeadingProps{
    Level: 1,
    Text: "Dashboard",
    Size: "4xl",
    Weight: "bold",
})

// Section heading (h2)
Heading(HeadingProps{
    Level: 2,
    Text: "Recent Activity",
    Size: "2xl",
    Weight: "semibold",
    Color: "gray-800",
})
```
<!-- LLM-HEADING-END -->

### Link Element

<!-- LLM-LINK-START -->
**PURPOSE**: Accessible hyperlink component with styling options

**FUNCTION SIGNATURE:**
```go
templ Link(props LinkProps) {}
```

**PROPS SPECIFICATION:**
```go
type LinkProps struct {
    Href      string  // Link destination URL
    Text      string  // Link text content
    Target    string  // Link target ("_blank", "_self", etc.)
    Rel       string  // Link relationship
    Underline bool    // Whether link is underlined (default: true)
    Color     string  // Link color variant ("blue", "red", "green")
    External  bool    // Whether link is external (adds security attributes)
    ID        string  // HTML id attribute
    Class     string  // Additional CSS classes
}
```

**USAGE EXAMPLES:**
```go
// Basic internal link
Link(LinkProps{
    Href: "/about",
    Text: "Learn More",
    Underline: true,
})

// External link with security
Link(LinkProps{
    Href: "https://example.com",
    Text: "Visit Example",
    External: true,
    Target: "_blank",
})
```
<!-- LLM-LINK-END -->

---

## Layout Elements

### Container Element

<!-- LLM-CONTAINER-START -->
**PURPOSE**: Responsive container with max-width constraints and centering

**FUNCTION SIGNATURE:**
```go
templ Container(props ContainerProps) {}
```

**PROPS SPECIFICATION:**
```go
type ContainerProps struct {
    Content  templ.Component  // Container content
    MaxWidth string          // "sm", "md", "lg", "xl", "2xl", "3xl", "4xl", "5xl", "6xl", "7xl", "full"
    Padding  string          // "none", "sm", "base", "lg", "xl"
    Center   bool            // Whether container is centered (default: true)
    ID       string          // HTML id attribute
    Class    string          // Additional CSS classes
}
```

**USAGE EXAMPLES:**
```go
// Main content container
Container(ContainerProps{
    Content: pageContent(),
    MaxWidth: "7xl",
    Padding: "lg",
    Center: true,
})

// Narrow form container
Container(ContainerProps{
    Content: loginForm(),
    MaxWidth: "md",
    Padding: "xl",
})
```
<!-- LLM-CONTAINER-END -->

### Grid Element

<!-- LLM-GRID-START -->
**PURPOSE**: CSS Grid layout component with responsive column support

**FUNCTION SIGNATURE:**
```go
templ Grid(props GridProps) {}
```

**PROPS SPECIFICATION:**
```go
type GridProps struct {
    Content    templ.Component     // Grid content
    Columns    string              // "1", "2", "3", "4", "5", "6", "12"
    Gap        string              // "0", "1", "2", "3", "4", "5", "6", "8", "10", "12"
    Responsive map[string]string   // Responsive column settings {"md": "2", "lg": "3"}
    ID         string              // HTML id attribute
    Class      string              // Additional CSS classes
}
```

**USAGE EXAMPLES:**
```go
// Responsive product grid
Grid(GridProps{
    Content: productCards(),
    Columns: "1",
    Responsive: map[string]string{
        "md": "2",
        "lg": "3",
        "xl": "4",
    },
    Gap: "6",
})

// Fixed layout grid
Grid(GridProps{
    Content: dashboardWidgets(),
    Columns: "3",
    Gap: "4",
})
```
<!-- LLM-GRID-END -->

---

## Composition Guidelines

<!-- LLM-COMPOSITION-START -->
**COMPOSITION PATTERNS:**

### 1. Container-Content Pattern
Use Container as a wrapper for consistent spacing and layout:
```
Container
├── Heading (page title)
├── Grid
│   ├── Card (item 1)
│   ├── Card (item 2)
│   └── Card (item 3)
```

### 2. Form Composition Pattern
Combine form elements with validation and layout:
```
Card
├── Heading (form title)
├── Form Container
│   ├── Input (field 1)
│   ├── Select (field 2)
│   ├── Textarea (field 3)
│   └── Button (submit)
└── Alert (validation messages)
```

### 3. Interactive Card Pattern
Create actionable content cards:
```
Card
├── Image (optional)
├── Heading (title)
├── Content (description)
└── Footer
    ├── Button (primary action)
    └── Link (secondary action)
```

**ATTRIBUTE INHERITANCE:**
1. Default values from component definitions
2. Props passed to complex components
3. Direct attributes on elements (highest priority)
4. CSS classes are appended, not replaced
<!-- LLM-COMPOSITION-END -->

## Accessibility Guidelines

<!-- LLM-ACCESSIBILITY-START -->
**BUILT-IN ACCESSIBILITY FEATURES:**
- Semantic HTML structure for all elements
- Proper label association for form controls
- ARIA attributes where semantic HTML is insufficient
- Keyboard navigation support
- Focus management for interactive elements
- Color contrast compliance (WCAG AA)

**ACCESSIBILITY CHECKLIST:**
- [ ] All interactive elements reachable via Tab key
- [ ] Form fields have associated labels
- [ ] Images include meaningful alt text
- [ ] Color is not the only means of conveying information
- [ ] Focus indicators are visible
- [ ] Screen reader announcements are clear

**COMMON PATTERNS:**
```go
// ✅ Good: Proper label association
Input(InputProps{
    Label: "Email Address",
    Name: "email",
    Type: "email",
    Required: true,
})

// ✅ Good: Descriptive button text
Button(ButtonProps{
    Text: "Save Profile Changes",
    AriaLabel: "Save changes to your user profile",
})

// ✅ Good: Meaningful link text
Link(LinkProps{
    Text: "View full product details",
    Href: "/products/123",
})
```
<!-- LLM-ACCESSIBILITY-END -->

## Performance Considerations

<!-- LLM-PERFORMANCE-START -->
**OPTIMIZATION STRATEGIES:**
1. **Component Reuse**: Define variants through props rather than duplicate components
2. **Conditional Rendering**: Use Templ's conditional syntax to avoid unnecessary HTML
3. **CSS Classes**: Reuse class combinations through helper functions
4. **Memory Management**: Pool frequently allocated objects in Go handlers

**EXAMPLE OPTIMIZATIONS:**
```go
// ✅ Efficient: Reuse class combinations
var buttonBaseClasses = "font-medium rounded-lg focus:ring-4 focus:outline-none transition-colors duration-200"

func getButtonClasses(variant, size string) string {
    sizeClass := buttonSizeClasses[size]
    variantClass := buttonVariantClasses[variant] 
    return strings.Join([]string{buttonBaseClasses, sizeClass, variantClass}, " ")
}

// ✅ Efficient: Conditional rendering
templ Button(props ButtonProps) {
    <button class={ getButtonClasses(props.Variant, props.Size) }>
        if props.Loading {
            @LoadingSpinner()
        }
        { props.Text }
    </button>
}
```
<!-- LLM-PERFORMANCE-END -->

## References

<!-- LLM-REFERENCES-START -->
**EXTERNAL REFERENCE FILES:**
- `templ-llms.md` - Advanced Templ features (streaming, suspense patterns)
- `flowbite-llms-full.txt` - Complete Flowbite component catalog

**OFFICIAL DOCUMENTATION:**
- [Templ Guide](https://templ.guide) - Go templating language reference
- [Flowbite Components](https://flowbite.com/docs/components/) - UI component library
- [Alpine.js Documentation](https://alpinejs.dev) - Reactive JavaScript framework
- [TailwindCSS](https://tailwindcss.com/docs) - Utility-first CSS framework

**RELATED DOCUMENTATION:**
- [Forms Guide](forms.md) - Advanced form patterns and validation
- [Composition Patterns](../patterns/composition.md) - Building complex UIs
- [HTMX Integration](../patterns/htmx-integration.md) - Server interaction patterns
<!-- LLM-REFERENCES-END -->

<!-- LLM-METADATA-START -->
**METADATA FOR AI ASSISTANTS:**
- File Type: Component Reference Documentation
- Scope: Foundational UI elements only
- Dependencies: Templ + Flowbite + TailwindCSS
- Optional: Alpine.js for enhanced interactivity  
- Target: Atomic design components
- Complexity: Beginner to Intermediate
<!-- LLM-METADATA-END -->
