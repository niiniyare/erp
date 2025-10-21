# CLAUDE.md - AI Assistant Guide

This file provides guidance to Claude Code and other AI assistants when working with the AWO ERP UI documentation and components.

## 📋 Quick Navigation for AI Assistants

### Essential Files
- **[README.md](README.md)** - Main documentation entry point
- **[getting-started.md](getting-started.md)** - Setup and basic examples
- **[Schema-Driven-Architecture.md](Schema-Driven-Architecture.md)** - Core system architecture
- **[templ-llms.md](templ-llms.md)** - Advanced Templ features and optimization

### Component Documentation
- **[components/](components/)** - Complete component library (140+ components)
- **[schema/definitions/](schema/definitions/)** - JSON schema specifications (900+ schemas)
- **[guides/](guides/)** - ERP workflow patterns and integration examples

## 🏗️ System Architecture Overview

### Technology Stack
```
Go Templ (Templates) + HTMX (Interactions) + Alpine.js (State) + Flowbite (Styling)
                                    ↓
                        JSON Schema-Driven UI System
                                    ↓
                     Pattern Renderer + Visual Builder
```

### Key Technologies
- **[Templ](https://templ.guide)** - Type-safe Go templating language (PRIMARY)
- **[HTMX](https://htmx.org)** - Server-driven UI interactions
- **[Alpine.js](https://alpinejs.dev)** - Lightweight client-side reactivity
- **[Flowbite](https://flowbite.com)** - Production-ready UI components
- **[TailwindCSS](https://tailwindcss.com)** - Utility-first CSS framework

## 🧩 Component Architecture

### Atomic Design Hierarchy
```
Templates/    - Application layouts (Root, Page, Service, Operation, Each, Switch)
Organisms/    - Complex business components (CRUD, Forms, Navigation, Modals)
Molecules/    - Composite UI components (Cards, Fields, Alerts, Dropdowns)
Atoms/        - Basic UI elements (Buttons, Inputs, Icons, Text)
```

### Schema Integration
The system uses JSON schemas to define UI components:
- **Schema definitions** in `schema/definitions/`
- **Pattern Renderer** interprets schemas and generates UI
- **Go models** automatically generate UI schemas
- **Visual Builder** provides drag-and-drop interface creation

## 📝 Writing Code with AWO ERP UI

### Basic Templ Component Pattern
```go
// Component with props
templ ComponentName(props ComponentProps) {
    <div class={ getComponentClasses(props) }>
        if props.ShowLabel {
            <label>{ props.Label }</label>
        }
        { children... }
    </div>
}

// Props structure
type ComponentProps struct {
    // Content
    Text        string
    Value       string
    
    // Styling  
    Variant     string
    Size        string
    
    // Behavior
    Disabled    bool
    OnClick     string
    
    // HTML
    ID          string
    Class       string
    AriaLabel   string
}
```

### HTMX Integration Pattern
```html
<form 
    hx-post="/api/endpoint"
    hx-target="#results" 
    hx-swap="innerHTML"
    hx-on::after-request="handleResponse(event)">
    <!-- Form content -->
</form>
```

### Alpine.js State Management
```html
<div x-data="{ 
    open: false, 
    loading: false,
    data: []
}">
    <button @click="open = !open">Toggle</button>
    <div x-show="open" x-transition>Content</div>
</div>
```

## 🎯 AI Assistant Guidelines

### When Creating Components
1. **Check existing components first** - Browse `components/` directory
2. **Follow atomic design** - Use appropriate component level
3. **Reference schemas** - Check `schema/definitions/` for specifications
4. **Use Flowbite patterns** - Leverage existing Flowbite components
5. **Ensure accessibility** - Include ARIA labels and keyboard navigation

### When Implementing Business Logic
1. **Server-first approach** - Business logic stays on Go backend
2. **Progressive enhancement** - Core functionality works without JavaScript
3. **HTMX for interactions** - Use server-driven UI updates
4. **Alpine.js for state** - Handle client-side reactive state only
5. **Type safety** - Leverage Go's type system throughout

### Code Examples to Reference
- **Basic components** - See `components/atoms/` for foundational patterns
- **Complex components** - See `components/organisms/` for business logic
- **Real-world workflows** - See `guides/erp-workflow-patterns.md`
- **Integration patterns** - See `guides/component-integration-guide.md`

## 📚 Common AI Assistant Tasks

### 1. Component Creation
**For new UI components:**
1. Check `components/` for existing similar components
2. Reference appropriate schema in `schema/definitions/`
3. Follow atomic design principles
4. Use Templ best practices from `templ-llms.md`

### 2. ERP Workflow Implementation
**For business process implementation:**
1. Start with `guides/erp-workflow-patterns.md`
2. Use organism-level components for complex business logic
3. Reference `guides/component-integration-guide.md` for composition
4. Follow multi-step form patterns in `components/organisms/`

### 3. Schema Integration
**For JSON schema work:**
1. Reference existing schemas in `schema/definitions/`
2. Follow schema patterns established in documentation
3. Ensure schema-component alignment
4. Test with Pattern Renderer system

### 4. Performance Optimization
**For performance concerns:**
1. Reference `templ-llms.md` for advanced optimization
2. Maintain JavaScript bundle under 50KB (currently 37.1KB)
3. Use server-side rendering for core functionality
4. Implement progressive enhancement patterns

## 🚨 Important Constraints

### Security Requirements
- **Server-side validation** - Never trust client-side data
- **CSRF protection** - Include tokens in state-changing requests
- **XSS prevention** - Sanitize all user-generated content
- **Access control** - Check permissions before rendering components

### Performance Requirements
- **JavaScript bundle** - Must stay under 50KB (currently 37.1KB)
- **Server-first** - Core functionality without JavaScript
- **Progressive enhancement** - Layer interactive features
- **Type safety** - Leverage Go's type system

### Accessibility Requirements
- **WCAG 2.1 AA compliance** - All components must be accessible
- **Keyboard navigation** - Full keyboard access to all functionality
- **Screen reader support** - Proper ARIA labels and structure
- **Color contrast** - Meet accessibility color requirements

## 🔧 Development Workflow

### Essential Commands
```bash
# Template development
templ generate --watch           # Hot reload
templ fmt                       # Format templates
templ generate                  # One-time generation

# Server development
go run ./cmd/server            # Start server
go test ./...                  # Run tests
go build                       # Build application

# Frontend development
npm run build:css              # Build TailwindCSS
npm run watch:css              # Watch CSS changes
```

### File Structure to Follow
```
web/components/
├── atoms/          # Basic elements (button, input, icon)
├── molecules/      # Composite components (card, field, alert)
├── organisms/      # Complex components (table, form, modal)
└── templates/      # Layout templates (page, service, operation)

docs/ui/
├── components/     # Component documentation
├── guides/         # Implementation guides
├── schema/         # Schema definitions
└── reference/      # API and design reference
```

## 📖 External References

### Advanced Features
- **`templ-llms.md`** - Streaming, suspense, optimization patterns
- **`Design-Pattern-Reference-Guide.md`** - Complete UI pattern catalog

### Official Documentation
- [Templ Guide](https://templ.guide) - Template language documentation
- [HTMX Docs](https://htmx.org/docs/) - Server interaction patterns
- [Alpine.js Guide](https://alpinejs.dev/start-here) - Reactive framework
- [Flowbite Components](https://flowbite.com/docs/components/) - UI library

---

**For AI Assistants**: This system prioritizes type safety, performance, and accessibility. Always check existing components before creating new ones, and ensure all implementations follow the established patterns and constraints.