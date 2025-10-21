# AWO ERP UI Component System

**FILE PURPOSE**: Main entry point for the AWO ERP UI component documentation system  
**SCOPE**: Complete guide to building enterprise ERP interfaces with JSON-driven architecture  
**TARGET AUDIENCE**: Go developers, frontend engineers, AI assistants, and system architects

## 🎯 Quick Navigation

### For Developers
- 🚀 **[Getting Started](getting-started.md)** - Setup, installation, and first component
- 🏗️ **[Architecture](Schema-Driven-Architecture.md)** - JSON-driven UI system design
- 🧩 **[Components](components/)** - Atomic design component library (140+ components)
- 📋 **[Schema Definitions](schema/)** - JSON schema specifications (900+ schemas)
- 📚 **[Guides](guides/)** - ERP workflow patterns and practical implementation

### For AI Assistants
- 📖 **[Schema Reference](schema/definitions/)** - Component schema definitions for code generation
- 🔧 **[Templ Advanced Features](templ-llms.md)** - Streaming, suspense, and optimization patterns
- 🎨 **[Design Patterns](Design-Pattern-Reference-Guide.md)** - Complete UI pattern reference
- 💼 **[Business Workflows](guides/erp-workflow-patterns.md)** - Real-world ERP implementation patterns

## 🏗️ System Architecture

### Core Technology Stack
```
┌─────────────────┬─────────────────┬─────────────────┐
│   Templates     │   Interaction   │    Styling      │
├─────────────────┼─────────────────┼─────────────────┤
│ Go Templ        │ HTMX            │ Flowbite        │
│ Type-safe HTML  │ Server-driven   │ Component lib   │
│ Hot reload      │ No JavaScript   │ TailwindCSS     │
└─────────────────┴─────────────────┴─────────────────┘
                          │
                          ▼
               ┌─────────────────────┐
               │   Alpine.js Store   │
               │   Client State      │
               │   Reactive UI       │
               └─────────────────────┘
```

### JSON-Driven UI System
The AWO ERP uses a sophisticated **schema-driven architecture** where:
- **JSON schemas** define UI component structure and behavior
- **Pattern Renderer** interprets schemas and generates live UI
- **Go models** automatically generate UI schemas via reflection
- **Visual Builder** provides drag-and-drop interface creation

## 📋 Component Library Structure

### Atomic Design Hierarchy
```
Templates/       - Application layout and page structure
  ├── Root       - Application container and global configuration
  ├── Page       - Basic page layouts with sidebars and toolbars  
  ├── Service    - Data fetching and orchestration
  ├── Operation  - Action grouping and button toolbars
  ├── Each       - Dynamic list rendering and iteration
  └── Switch     - Conditional rendering based on state

Organisms/       - Complex business components
  ├── CRUD       - Data tables with full CRUD operations
  ├── Forms      - Multi-step forms with validation
  ├── Navigation - Sidebars, breadcrumbs, and menus
  └── Modals     - Dialog boxes and overlay components

Molecules/       - Composite UI components  
  ├── Cards      - Content containers with headers/actions
  ├── Fields     - Form inputs with labels and validation
  ├── Alerts     - Notifications and status messages
  └── Dropdowns  - Selection and action menus

Atoms/          - Basic UI elements
  ├── Buttons    - Actions in multiple variants and states
  ├── Inputs     - Text fields, checkboxes, radios
  ├── Icons      - SVG icon system with size variants
  └── Text       - Typography and content display
```

## 🎯 Key Features

### Enterprise-Grade Capabilities
- ✅ **Multi-tenant architecture** with tenant isolation
- ✅ **Role-based access control** with UI component permissions  
- ✅ **Real-time form validation** with server-side logic
- ✅ **Responsive design** optimized for desktop and mobile
- ✅ **Accessibility compliance** (WCAG 2.1 AA)
- ✅ **International localization** with Go i18n integration
- ✅ **Performance optimization** with 37.1KB JavaScript bundle

### Developer Experience
- 🔥 **Hot reload development** with `templ generate --watch`
- 🛡️ **Type safety** across Go templates and data structures
- 📦 **Component composition** following atomic design principles
- 🧪 **Testing strategies** for template and integration testing
- 📊 **Schema validation** preventing UI/backend misalignment
- 🎨 **Visual debugging** with component inspection tools

## 🚀 Quick Start

### 1. Prerequisites
```bash
# Install required tools
go install github.com/a-h/templ/cmd/templ@latest

# Verify installation
go version  # Requires Go 1.21+
templ --help
```

### 2. Basic Component Example
```go
// hello.templ
package components

templ HelloWorld(name string) {
    <div class="p-4 bg-blue-50 rounded-lg">
        <h1 class="text-xl font-bold text-blue-900">
            Hello, { name }!
        </h1>
    </div>
}
```

### 3. Development Workflow
```bash
# Start hot reload
templ generate --watch

# In another terminal, run server
go run main.go

# Open browser to see live updates
open http://localhost:8080
```

## 📚 Documentation Structure

### Essential Reading Order
1. **[Getting Started](getting-started.md)** - Set up development environment
2. **[Schema-Driven Architecture](Schema-Driven-Architecture.md)** - Understand the JSON-UI system
3. **[Component Fundamentals](components/atoms/)** - Learn basic building blocks
4. **[ERP Workflow Patterns](guides/erp-workflow-patterns.md)** - Implement business processes
5. **[Component Integration](guides/component-integration-guide.md)** - Build complete applications

### Advanced Topics
- **[Templ Advanced Features](templ-llms.md)** - Streaming, optimization, and performance
- **[Design Pattern Reference](Design-Pattern-Reference-Guide.md)** - Complete UI pattern catalog
- **[Schema Definitions](schema/definitions/)** - Component specification reference

## 🔧 Development Commands

```bash
# Component development
templ generate                    # Generate templates
templ generate --watch           # Hot reload development
templ fmt                        # Format templates

# Application development  
go run ./cmd/server              # Start ERP server
go test ./...                    # Run tests
go build -o bin/server ./cmd/server  # Build for production

# UI development
npm run build:css               # Build TailwindCSS
npm run watch:css               # Watch CSS changes
```

## 🏢 Real-World Usage

### ERP Module Examples
The component system supports complete ERP workflows:

- **Customer Management** - Lead capture, customer profiles, relationship tracking
- **Order Processing** - Quote generation, order fulfillment, shipping management  
- **Financial Management** - Invoicing, payment processing, financial reporting
- **Inventory Control** - Stock management, reorder automation, warehouse operations
- **Manufacturing** - Production planning, quality control, maintenance scheduling

### Integration Capabilities
- 🔌 **REST API integration** via HTMX and Go handlers
- 🔄 **Real-time updates** with WebSocket and Server-Sent Events
- 📊 **Data visualization** with Chart.js and custom components
- 📄 **PDF generation** for reports and documents
- 📧 **Email integration** for notifications and workflows

## 📖 External References

### Official Documentation
- [Templ Guide](https://templ.guide) - Official templating language documentation
- [HTMX Documentation](https://htmx.org/docs/) - Hypermedia-driven interactions
- [Alpine.js Guide](https://alpinejs.dev/start-here) - Reactive JavaScript framework
- [Flowbite Components](https://flowbite.com/docs/components/) - UI component library
- [TailwindCSS](https://tailwindcss.com/docs) - Utility-first CSS framework

### Community Resources
- [Go Templates Best Practices](https://golang.org/pkg/html/template/)
- [HTMX Examples](https://htmx.org/examples/)
- [Alpine.js Patterns](https://alpinejs.dev/start-here)

## 🤝 Contributing

### Documentation Standards
- **LLM-Friendly**: Clear section markers and semantic structure
- **Progressive Complexity**: Basic → Intermediate → Advanced learning paths
- **Practical Examples**: Working code snippets and real-world scenarios
- **Cross-Referenced**: Consistent linking between related sections

### Code Standards
- **Type Safety**: All templates must compile without errors
- **Performance**: JavaScript bundle maintained under 50KB (currently 37.1KB)
- **Accessibility**: WCAG 2.1 AA compliance required
- **Testing**: Unit tests for components and integration tests for workflows

---

**VERSION**: 2024.3  
**LAST UPDATED**: December 2024  
**MAINTAINER**: AWO ERP Development Team  
**LICENSE**: Proprietary - AWO Enterprise Solutions