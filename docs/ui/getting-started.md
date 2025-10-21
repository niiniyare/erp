# Getting Started with AWO ERP UI

**FILE PURPOSE**: Quick setup guide for the AWO ERP UI component system  
**PREREQUISITES**: Go 1.21+, basic HTML/CSS knowledge  
**ESTIMATED TIME**: 10-15 minutes  
**OUTCOME**: Working development environment with hot reload

## Quick Setup

### 1. Install Prerequisites
```bash
# Verify Go installation (required: 1.21+)
go version

# Install Templ CLI
go install github.com/a-h/templ/cmd/templ@latest

# Verify installation
templ --help
```

### 2. Clone AWO ERP Project
```bash
# Clone the repository
git clone <awo-erp-repo-url>
cd awo-erp

# Generate templates
templ generate

# Start development server
go run ./cmd/server
```

### 3. Technology Stack
- **Go Templ** - Type-safe HTML templating
- **HTMX** - Server-driven interactions  
- **Alpine.js** - Client-side reactivity
- **Flowbite** - UI component library
- **TailwindCSS** - Utility-first styling

## Development Workflow

### Hot Reload Development
```bash
# Terminal 1: Watch templates
templ generate --watch

# Terminal 2: Run server
go run ./cmd/server

# Terminal 3: Watch CSS (if needed)
npm run watch:css
```

### Project Structure
```
project/
├── cmd/server/          # Application entry point
├── web/
│   ├── components/      # UI components (atomic design)
│   │   ├── atoms/       # Basic elements (buttons, inputs)
│   │   ├── molecules/   # Composite components (cards, fields)
│   │   ├── organisms/   # Complex components (tables, forms)
│   │   └── templates/   # Page layouts and structure
│   ├── layouts/         # Base layouts and shells
│   ├── pages/           # Application pages
│   └── static/          # CSS, JS, images
└── docs/ui/             # This documentation
```

## Your First Component

### 1. Create a Simple Component
Create `web/components/atoms/greeting.templ`:

```go
package atoms

templ Greeting(name string) {
    <div class="p-4 bg-blue-50 border border-blue-200 rounded-lg">
        <h2 class="text-lg font-semibold text-blue-900">
            Hello, { name }!
        </h2>
        <p class="text-blue-700 mt-2">
            Welcome to AWO ERP UI System
        </p>
    </div>
}
```

### 2. Use the Component
In a page file `web/pages/example.templ`:

```go
package pages

import "your-project/web/components/atoms"
import "your-project/web/layouts"

templ ExamplePage() {
    @layouts.App("Example") {
        <div class="container mx-auto py-8">
            @atoms.Greeting("Developer")
        </div>
    }
}
```

### 3. Generate and Test
```bash
# Generate templates
templ generate

# Start server and visit /example
go run ./cmd/server
```

## Core Concepts

### 1. Atomic Design
Components are organized in a hierarchy:
- **Atoms** - Basic elements (buttons, inputs, icons)
- **Molecules** - Simple combinations (search box, card header)
- **Organisms** - Complex UI sections (navigation, data tables)
- **Templates** - Page layouts and structure

### 2. JSON Schema Integration
Components can be generated from JSON schemas:

```json
{
    "type": "card",
    "header": {
        "title": "Customer Details",
        "subtitle": "John Doe"
    },
    "body": [
        {"type": "text", "value": "Email: john@example.com"},
        {"type": "text", "value": "Phone: (555) 123-4567"}
    ]
}
```

### 3. HTMX Integration
Server interactions without JavaScript:

```go
templ UserForm() {
    <form hx-post="/api/users" hx-target="#result">
        <input type="text" name="name" required />
        <button type="submit">Save User</button>
    </form>
    <div id="result"></div>
}
```

## Next Steps

### Essential Reading
1. **[Schema-Driven Architecture](Schema-Driven-Architecture.md)** - Understand the JSON-UI system
2. **[Component Library](components/)** - Browse available components  
3. **[ERP Workflow Patterns](guides/erp-workflow-patterns.md)** - Real-world examples
4. **[Templ Advanced Features](templ-llms.md)** - Performance and optimization

### Common Tasks
- **Add new components** → Follow atomic design structure
- **Implement forms** → Use validation patterns in `components/molecules/`
- **Create data tables** → Use CRUD organisms in `components/organisms/`
- **Build workflows** → Reference `guides/erp-workflow-patterns.md`

### Development Commands
```bash
# Essential commands
templ generate --watch    # Hot reload development
templ fmt                 # Format templates  
go test ./...            # Run tests
go build                 # Build application

# Debugging
templ generate --log-level=debug  # Verbose output
go run -race ./cmd/server         # Race detection
```

## Troubleshooting

### Common Issues

**Templates not generating:**
```bash
# Check syntax
templ generate

# Clear cache and regenerate
rm -rf web/*_templ.go
templ generate
```

**Styling not applied:**
```bash
# Ensure Flowbite CSS is loaded
# Check browser dev tools for CSS errors
# Verify TailwindCSS classes are correct
```

**HTMX not working:**
```bash
# Verify HTMX is loaded in layout
# Check browser network tab for requests
# Ensure server returns proper HTML fragments
```

### Getting Help
- **Documentation Issues** - Check component README files
- **Code Examples** - Browse `docs/ui/guides/` directory
- **Schema Reference** - See `docs/ui/schema/definitions/`
- **Advanced Patterns** - Review `templ-llms.md`

---

**You're now ready to build sophisticated ERP interfaces with the AWO UI system!**