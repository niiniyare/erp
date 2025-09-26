# CLAUDE.md - UI Documentation Guide

This file provides guidance to Claude Code and other AI assistants when working with the ERP UI documentation and components in this repository.

## Documentation Overview

This UI documentation system is designed to be **LLM-friendly** and provides comprehensive guidance for building ERP interfaces using **Templ + HTMX + Alpine.js + Flowbite**.

### Documentation Structure

```
docs/ui/
├── README.md                     # 🎯 LLM-optimized entry point and navigation
├── getting-started.md            # 🚀 Complete setup and installation guide
├── CLAUDE.md                     # 🤖 This file - AI assistant guidance
├── templ-llms.md                 # 📚 Advanced Templ features reference
├── flowbite-llms-full.txt        # 🎨 Complete Flowbite styling reference
├── fundamentals/
│   └── architecture.md           # 🏗️ Core system architecture and design principles
├── components/
│   ├── elements.md               # 🧱 Foundational UI elements (buttons, inputs, cards)
│   └── forms.md                  # 📝 Form components and interactive patterns
├── patterns/
│   └── htmx-integration.md       # 🔄 Server interaction patterns and HTMX implementation
├── guides/
│   └── validation-guide.md       # ✅ Complete validation system implementation
└── reference/
    └── api-reference.md          # 📖 Complete API documentation and type definitions
```

## Key Technologies

### Core Stack
- **[Templ](https://templ.guide)** - Type-safe Go templating language that compiles to Go code
- **[HTMX](https://htmx.org)** - Server-driven UI interactions without complex JavaScript
- **[Alpine.js](https://alpinejs.dev)** - Lightweight reactive JavaScript framework
- **[Flowbite](https://flowbite.com)** - UI component library built on TailwindCSS

### Architecture Pattern
- **Server-First** - Business logic and validation on the server
- **Progressive Enhancement** - Works without JavaScript, enhanced with it
- **Component-Based** - Atomic design with reusable Templ components
- **Type-Safe** - Leverages Go's type system for templates and props

## Navigation Guide for AI Assistants

### 🎯 Quick Start Locations

1. **New to the system?** → Start with `README.md` for overview and navigation
2. **Setting up a project?** → Go to `getting-started.md` for complete setup instructions
3. **Building components?** → Use `components/elements.md` for foundational UI elements
4. **Creating forms?** → Reference `components/forms.md` for form patterns
5. **Server integration?** → Check `patterns/htmx-integration.md` for HTMX patterns
6. **Implementation details?** → Use `guides/validation-guide.md` for complete validation system
7. **API reference needed?** → Consult `reference/api-reference.md` for complete type definitions

### 🔍 LLM Navigation Markers

All documentation files include semantic markers for AI navigation:

```html
<!-- LLM-CONTEXT-START -->
**FILE PURPOSE**: Brief description of file purpose
**SCOPE**: What this file covers
**TARGET AUDIENCE**: Who should use this file
<!-- LLM-CONTEXT-END -->

<!-- LLM-SECTION-NAME-START -->
Content organized by semantic sections...
<!-- LLM-SECTION-NAME-END -->
```

These markers help AI assistants quickly locate relevant information.

## Common AI Assistant Tasks

### 1. Component Implementation
**When asked to create UI components:**
1. Start with `components/elements.md` for foundational patterns
2. Reference `components/forms.md` for interactive elements
3. Check `reference/api-reference.md` for exact type definitions
4. Use `templ-llms.md` for advanced Templ features

### 2. Server Integration
**When implementing HTMX interactions:**
1. Use `patterns/htmx-integration.md` for request/response patterns
2. Reference `guides/validation-guide.md` for server-side validation
3. Check `fundamentals/architecture.md` for system design principles

### 3. Form Development
**When building forms and validation:**
1. Start with `components/forms.md` for component patterns
2. Use `guides/validation-guide.md` for complete validation implementation
3. Reference `patterns/htmx-integration.md` for server communication
4. Check `reference/api-reference.md` for validation types

### 4. Styling and Design
**When applying styling and design:**
1. Reference `flowbite-llms-full.txt` for complete component styling
2. Use `components/elements.md` for design system patterns
3. Check `fundamentals/architecture.md` for design principles

## Code Patterns to Follow

### 1. Templ Component Structure
```go
// Always follow this pattern for components
templ ComponentName(props ComponentProps) {
    <element class={ getComponentClasses(props) }>
        if props.ShowLabel {
            <label>{ props.Label }</label>
        }
        { children... }
    </element>
}
```

### 2. Props Pattern
```go
// Consistent props structure across all components
type ComponentProps struct {
    // Content properties
    Text        string
    Value       string
    
    // Styling properties
    Variant     string
    Size        string
    
    // Behavior properties
    Disabled    bool
    OnClick     string
    
    // HTML properties
    ID          string
    Class       string
    AriaLabel   string
}
```

### 3. HTMX Integration Pattern
```html
<!-- Follow this pattern for server interactions -->
<form 
    hx-post="/api/endpoint"
    hx-target="#results"
    hx-swap="innerHTML"
    hx-on::after-request="handleResponse(event)">
    <!-- Form content -->
</form>
```

## Important Implementation Notes

### Security Requirements
- **Always validate inputs server-side** - Never trust client data
- **Include CSRF protection** for state-changing requests
- **Sanitize HTML output** to prevent XSS attacks
- **Use proper HTTP status codes** for different error scenarios

### Performance Considerations
- **Debounce rapid requests** like search with `delay:300ms`
- **Use lazy loading** for expensive content with `intersect once`
- **Return minimal HTML** from server endpoints
- **Cache responses** when appropriate with ETags

### Accessibility Requirements
- **Include proper ARIA labels** for all interactive elements
- **Ensure keyboard navigation** works for all components
- **Provide meaningful error messages** for form validation
- **Use semantic HTML** structure throughout

## File-Specific Guidance

### `README.md`
- **Purpose**: Main entry point with comprehensive navigation
- **Use when**: AI needs overview or navigation assistance
- **Contains**: Project overview, quick start, technology explanations

### `getting-started.md`
- **Purpose**: Complete setup and project initialization
- **Use when**: AI needs to help users set up new projects
- **Contains**: Installation, configuration, project structure, examples

### `components/elements.md`
- **Purpose**: Foundational UI element reference
- **Use when**: AI needs to create basic components (buttons, inputs, cards)
- **Contains**: Component props, usage patterns, accessibility features

### `components/forms.md`
- **Purpose**: Form component patterns and interactions
- **Use when**: AI needs to build forms or interactive elements
- **Contains**: Form components, validation patterns, multi-step forms

### `patterns/htmx-integration.md`
- **Purpose**: Server interaction patterns and HTMX implementation
- **Use when**: AI needs to implement server-client communication
- **Contains**: Request patterns, response handling, error management

### `guides/validation-guide.md`
- **Purpose**: Complete validation system implementation
- **Use when**: AI needs to implement form validation or server-side validation
- **Contains**: Validation architecture, server implementation, client integration

### `reference/api-reference.md`
- **Purpose**: Complete API documentation and type definitions
- **Use when**: AI needs exact function signatures or type information
- **Contains**: Component APIs, handler patterns, type definitions

## External Reference Files

### `templ-llms.md`
- **Purpose**: Advanced Templ features and optimization patterns
- **Use when**: AI needs advanced templating features
- **Contains**: Streaming, suspense, performance optimization

### `flowbite-llms-full.txt`
- **Purpose**: Complete Flowbite component catalog and styling reference
- **Use when**: AI needs specific styling or component examples
- **Contains**: Full Flowbite documentation and examples

## Best Practices for AI Assistants

### 1. Always Start with Documentation
- Read relevant documentation files before implementing
- Use the navigation markers to find specific information quickly
- Reference multiple files for complete understanding

### 2. Follow Established Patterns
- Use the props patterns consistently across components
- Follow the HTMX integration patterns for server communication
- Implement validation using the three-layer approach

### 3. Provide Complete Solutions
- Include proper error handling in all implementations
- Add accessibility features to all components
- Implement security measures (CSRF, validation, sanitization)

### 4. Reference External Resources
- Point users to `templ-llms.md` for advanced Templ features
- Reference `flowbite-llms-full.txt` for styling questions
- Include links to official documentation

## Troubleshooting Common Issues

### Component Not Rendering
1. Check Templ syntax in the component file
2. Verify props are being passed correctly
3. Ensure `templ generate` has been run
4. Check for compilation errors in Go code

### HTMX Not Working
1. Verify HTMX is loaded in the page
2. Check server endpoints are returning proper HTML
3. Validate HTMX attributes are correct
4. Check browser network tab for request/response

### Styling Issues
1. Verify Flowbite CSS is loaded
2. Check class names match Flowbite documentation
3. Ensure TailwindCSS is properly configured
4. Reference `flowbite-llms-full.txt` for examples

### Validation Problems
1. Check server-side validation logic
2. Verify client-side Alpine.js validation
3. Ensure proper error message display
4. Reference `guides/validation-guide.md` for complete patterns

## Version and Compatibility

This documentation is designed for:
- **Templ**: Latest version (1.0+)
- **HTMX**: Version 1.9+
- **Alpine.js**: Version 3.13+
- **Flowbite**: Version 2.0+
- **Go**: Version 1.21+

Always check the official documentation for the latest features and compatibility information.

---

**For AI Assistants**: This documentation system is designed to provide comprehensive, LLM-friendly guidance for building robust ERP interfaces. Use the semantic markers and structured navigation to quickly find relevant information and provide accurate, complete solutions to users.