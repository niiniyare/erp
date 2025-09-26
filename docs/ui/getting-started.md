# Getting Started with ERP UI Components

<!-- LLM-CONTEXT-START -->
**FILE PURPOSE**: Complete setup guide for Templ + HTMX + Alpine.js + Flowbite stack
**PREREQUISITES**: Go 1.21+, Node.js 18+, basic HTML/CSS knowledge
**ESTIMATED TIME**: 15-30 minutes
**OUTCOME**: Working component development environment with hot reload
**REFERENCE FILES**: `templ-llms.md` for advanced features, `flowbite-llms-full.txt` for component catalog
<!-- LLM-CONTEXT-END -->

## Quick Setup Checklist

<!-- LLM-CHECKLIST-START -->
**SETUP CHECKLIST FOR LLMs:**
- [ ] Install Go 1.21+
- [ ] Install Templ CLI: `go install github.com/a-h/templ/cmd/templ@latest`
- [ ] Create project structure
- [ ] Add dependencies (Flowbite, HTMX, Alpine.js)
- [ ] Create first component
- [ ] Test hot reload workflow
<!-- LLM-CHECKLIST-END -->

## 1. Prerequisites & Installation

### System Requirements

<!-- LLM-REQUIREMENTS-START -->
```bash
# Check Go version (required: 1.21+)
go version

# Check Node.js version (optional, for npm packages)
node --version

# Install Templ CLI
go install github.com/a-h/templ/cmd/templ@latest

# Verify Templ installation
templ --help
```
<!-- LLM-REQUIREMENTS-END -->

### Technology Stack Overview

| Technology | Purpose | Required | CDN Available |
|------------|---------|----------|---------------|
| **[Templ](https://templ.guide)** | Go templating language | ✅ Yes | N/A |
| **[HTMX](https://htmx.org)** | Hypermedia interactions | ✅ Yes | ✅ Yes |
| **[Alpine.js](https://alpinejs.dev)** | Reactive JavaScript | ❌ Optional | ✅ Yes |
| **[Flowbite](https://flowbite.com)** | UI component library | ✅ Yes | ✅ Yes |
| **[TailwindCSS](https://tailwindcss.com)** | Utility-first CSS | ✅ Yes | ✅ Via Flowbite |

## 2. Project Structure Setup

### Create New Project

<!-- LLM-PROJECT-SETUP-START -->
```bash
# Create project directory
mkdir erp-ui-project && cd erp-ui-project

# Initialize Go module
go mod init erp-ui-project

# Add Templ dependency
go get github.com/a-h/templ

# Create directory structure
mkdir -p {components/{elements,forms,layout,advanced},templates/{layouts,pages},static/{css,js},cmd/server}
```

**RECOMMENDED PROJECT STRUCTURE:**
```
erp-ui-project/
├── go.mod
├── go.sum
├── components/                 # Templ components
│   ├── elements/              # Basic components (buttons, inputs)
│   ├── forms/                 # Form components with validation
│   ├── layout/                # Layout components (nav, containers)
│   └── advanced/              # Complex components (tables, modals)
├── templates/                 # Page templates
│   ├── layouts/              # Layout templates
│   └── pages/                # Page-specific templates
├── static/                   # Static assets
│   ├── css/                  # Custom CSS
│   └── js/                   # Custom JavaScript
├── cmd/server/              # Main application
│   └── main.go
└── README.md
```
<!-- LLM-PROJECT-SETUP-END -->

## 3. Base HTML Template

Create the foundation HTML template with all required dependencies:

<!-- LLM-BASE-TEMPLATE-START -->
```html
<!-- templates/layouts/base.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - ERP System</title>
    
    <!-- REQUIRED: Flowbite CSS (includes TailwindCSS) -->
    <link href="https://cdn.jsdelivr.net/npm/flowbite@2.5.1/dist/flowbite.min.css" rel="stylesheet">
    
    <!-- REQUIRED: HTMX for server interactions -->
    <script src="https://unpkg.com/htmx.org@1.9.12"></script>
    
    <!-- OPTIONAL: Alpine.js for client-side reactivity -->
    <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
    
    <!-- HTMX & Alpine.js integration styles -->
    <style>
        /* HTMX loading states */
        .htmx-request { opacity: 0.5; }
        .htmx-swapping { opacity: 0; }
        .htmx-settling { opacity: 1; }
        .htmx-indicator { display: none; }
        .htmx-request .htmx-indicator { display: inline; }
        
        /* Alpine.js transitions */
        [x-cloak] { display: none !important; }
    </style>
</head>
<body class="bg-gray-50">
    <!-- Content will be injected here -->
    <div id="app">{{.Content}}</div>
    
    <!-- REQUIRED: Flowbite JavaScript for interactive components -->
    <script src="https://cdn.jsdelivr.net/npm/flowbite@2.5.1/dist/flowbite.min.js"></script>
</body>
</html>
```
<!-- LLM-BASE-TEMPLATE-END -->

## 4. Your First Component

Create a basic button component to test the setup:

<!-- LLM-FIRST-COMPONENT-START -->
```go
// components/elements/button.templ
package elements

import "strings"

// ButtonProps defines the properties for the Button component
type ButtonProps struct {
    Text     string `json:"text" validate:"required,min=1,max=50"`
    Variant  string `json:"variant" validate:"oneof=primary secondary success danger warning info light dark"`
    Size     string `json:"size" validate:"oneof=xs sm base lg xl"`
    Disabled bool   `json:"disabled"`
    Loading  bool   `json:"loading"`
    OnClick  string `json:"onclick"`
    Type     string `json:"type" validate:"oneof=button submit reset"`
    ID       string `json:"id"`
    Class    string `json:"class"`
}

// Button renders a Flowbite-styled button component
templ Button(props ButtonProps) {
    <button 
        type={ getButtonType(props.Type) }
        if props.ID != "" {
            id={ props.ID }
        }
        class={ getButtonClasses(props.Variant, props.Size) + " " + props.Class }
        if props.Disabled || props.Loading {
            disabled
        }
        if props.OnClick != "" {
            onclick={ props.OnClick }
        }
    >
        if props.Loading {
            <!-- Loading spinner -->
            <svg class="animate-spin -ml-1 mr-3 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Loading...
        } else {
            { props.Text }
        }
    </button>
}

// Helper functions for button styling (Flowbite classes)
func getButtonClasses(variant, size string) string {
    base := "font-medium rounded-lg focus:ring-4 focus:outline-none transition-colors duration-200"
    
    // Size classes
    var sizeClass string
    switch size {
    case "xs":
        sizeClass = "px-3 py-2 text-xs"
    case "sm":
        sizeClass = "px-3 py-2 text-sm"
    case "lg":
        sizeClass = "px-5 py-3 text-base"
    case "xl":
        sizeClass = "px-6 py-3.5 text-base"
    default: // base
        sizeClass = "px-5 py-2.5 text-sm"
    }
    
    // Variant classes (using Flowbite design system)
    var variantClass string
    switch variant {
    case "primary":
        variantClass = "text-white bg-blue-700 hover:bg-blue-800 focus:ring-blue-300 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
    case "secondary":
        variantClass = "text-gray-900 bg-white border border-gray-300 hover:bg-gray-100 focus:ring-gray-200 dark:bg-gray-800 dark:text-white dark:border-gray-600 dark:hover:bg-gray-700 dark:focus:ring-gray-700"
    case "success":
        variantClass = "text-white bg-green-700 hover:bg-green-800 focus:ring-green-300 dark:bg-green-600 dark:hover:bg-green-700 dark:focus:ring-green-800"
    case "danger":
        variantClass = "text-white bg-red-700 hover:bg-red-800 focus:ring-red-300 dark:bg-red-600 dark:hover:bg-red-700 dark:focus:ring-red-900"
    case "warning":
        variantClass = "text-white bg-yellow-400 hover:bg-yellow-500 focus:ring-yellow-300 dark:focus:ring-yellow-900"
    case "info":
        variantClass = "text-white bg-cyan-700 hover:bg-cyan-800 focus:ring-cyan-300 dark:bg-cyan-600 dark:hover:bg-cyan-700 dark:focus:ring-cyan-800"
    case "light":
        variantClass = "text-gray-900 bg-white border border-gray-300 hover:bg-gray-100 focus:ring-gray-200 dark:bg-gray-800 dark:text-white dark:border-gray-600 dark:hover:bg-gray-700 dark:focus:ring-gray-700"
    case "dark":
        variantClass = "text-white bg-gray-800 hover:bg-gray-900 focus:ring-gray-300 dark:bg-gray-800 dark:hover:bg-gray-700 dark:focus:ring-gray-700"
    default:
        variantClass = "text-white bg-blue-700 hover:bg-blue-800 focus:ring-blue-300 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
    }
    
    return strings.Join([]string{base, sizeClass, variantClass}, " ")
}

func getButtonType(btnType string) string {
    if btnType == "" {
        return "button"
    }
    return btnType
}
```
<!-- LLM-FIRST-COMPONENT-END -->

## 5. Test Server Setup

Create a minimal server to test your components:

<!-- LLM-SERVER-SETUP-START -->
```go
// cmd/server/main.go
package main

import (
    "context"
    "log"
    "net/http"
    
    "erp-ui-project/components/elements"
    "github.com/a-h/templ"
)

func main() {
    // Create a test page component
    component := templ.ComponentFunc(func(ctx context.Context, w templ.ComponentWriter) error {
        // Write HTML document structure
        _, err := w.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ERP UI Components Test</title>
    <link href="https://cdn.jsdelivr.net/npm/flowbite@2.5.1/dist/flowbite.min.css" rel="stylesheet">
    <script src="https://unpkg.com/htmx.org@1.9.12"></script>
    <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body class="bg-gray-50 p-8">
    <div class="max-w-4xl mx-auto">
        <h1 class="text-3xl font-bold text-gray-900 mb-8">ERP UI Components Test</h1>
        <div class="space-y-4">
`)
        if err != nil {
            return err
        }
        
        // Test different button variants
        buttons := []elements.ButtonProps{
            {Text: "Primary Button", Variant: "primary", Size: "base"},
            {Text: "Secondary Button", Variant: "secondary", Size: "base"},
            {Text: "Success Button", Variant: "success", Size: "sm"},
            {Text: "Danger Button", Variant: "danger", Size: "lg"},
            {Text: "Loading Button", Variant: "primary", Size: "base", Loading: true},
            {Text: "Disabled Button", Variant: "primary", Size: "base", Disabled: true},
        }
        
        for _, btnProps := range buttons {
            err = elements.Button(btnProps).Render(ctx, w)
            if err != nil {
                return err
            }
        }
        
        // Close HTML document
        _, err = w.WriteString(`
        </div>
        
        <!-- HTMX Test Section -->
        <div class="mt-8">
            <h2 class="text-2xl font-bold text-gray-900 mb-4">HTMX Test</h2>
            <button 
                hx-get="/test-htmx" 
                hx-target="#htmx-result"
                class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">
                Test HTMX
            </button>
            <div id="htmx-result" class="mt-4 p-4 bg-gray-100 rounded-lg min-h-[50px]">
                <!-- HTMX response will appear here -->
            </div>
        </div>
        
        <!-- Alpine.js Test Section -->
        <div class="mt-8" x-data="{ count: 0 }">
            <h2 class="text-2xl font-bold text-gray-900 mb-4">Alpine.js Test</h2>
            <button 
                @click="count++" 
                class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700">
                Increment
            </button>
            <p class="mt-2 text-lg">Count: <span x-text="count" class="font-bold"></span></p>
        </div>
    </div>
    <script src="https://cdn.jsdelivr.net/npm/flowbite@2.5.1/dist/flowbite.min.js"></script>
</body>
</html>`)
        return err
    })

    // HTMX test endpoint
    http.HandleFunc("/test-htmx", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        w.WriteString(`<p class="text-green-600 font-semibold">✅ HTMX is working! Request received at ` + 
            r.Header.Get("X-Forwarded-For") + `</p>`)
    })

    // Main page
    http.Handle("/", templ.Handler(component))
    
    log.Println("🚀 Server starting on :8080")
    log.Println("📱 Visit: http://localhost:8080")
    log.Println("📚 Documentation: docs/ui/README.md")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```
<!-- LLM-SERVER-SETUP-END -->

## 6. Development Workflow

### Hot Reload Setup

<!-- LLM-WORKFLOW-START -->
```bash
# Terminal 1: Start Templ hot reload (watches .templ files)
templ generate --watch

# Terminal 2: Run your server with auto-restart (install air first)
go install github.com/cosmtrek/air@latest
air

# OR: Manual server restart
go run cmd/server/main.go
```

**DEVELOPMENT COMMANDS:**
- `templ generate` - Generate Go code from .templ files
- `templ generate --watch` - Auto-regenerate on file changes
- `go run cmd/server/main.go` - Run development server
- `go mod tidy` - Clean up dependencies
- `go test ./...` - Run tests
<!-- LLM-WORKFLOW-END -->

### Testing Your Setup

1. **Start Templ watcher**: Open terminal → `templ generate --watch`
2. **Start server**: Open another terminal → `go run cmd/server/main.go`
3. **Open browser**: Navigate to `http://localhost:8080`
4. **Verify functionality**:
   - ✅ Buttons styled with Flowbite
   - ✅ HTMX button makes server request
   - ✅ Alpine.js counter increments
   - ✅ No console errors

<!-- LLM-TROUBLESHOOTING-START -->
## Troubleshooting

**COMMON ISSUES & SOLUTIONS:**

1. **"templ: command not found"**
   ```bash
   # Add Go bin to PATH
   export PATH=$PATH:$(go env GOPATH)/bin
   # Add to ~/.bashrc or ~/.zshrc for persistence
   ```

2. **"Package not found" errors**
   ```bash
   # Ensure module is initialized
   go mod init your-project-name
   go mod tidy
   ```

3. **Templates not updating**
   ```bash
   # Make sure templ generate is running
   templ generate
   # Check for syntax errors in .templ files
   ```

4. **Styles not loading**
   - Check browser network tab for 404 errors
   - Verify Flowbite CDN is accessible
   - Ensure correct CSS class names (see `flowbite-llms-full.txt`)

5. **HTMX not working**
   - Check browser console for JavaScript errors
   - Verify HTMX script is loaded
   - Check network tab for failed requests

6. **Alpine.js not reactive**
   - Ensure Alpine.js script has `defer` attribute
   - Check for JavaScript syntax errors
   - Verify `x-data` initialization
<!-- LLM-TROUBLESHOOTING-END -->

## 7. Next Steps

<!-- LLM-NEXT-STEPS-START -->
**LEARNING PROGRESSION:**
1. ✅ **Basic Setup** (you are here)
2. 📚 **[Basic Components](components/elements.md)** - Learn all available elements
3. 🏗️ **[Architecture](fundamentals/architecture.md)** - Understand system design
4. 📋 **[Form Components](components/forms.md)** - Build interactive forms with validation
5. 🔧 **[Patterns](patterns/composition.md)** - Advanced composition patterns
6. 🚀 **[Deployment](guides/deployment.md)** - Production deployment strategies

**IMMEDIATE ACTIONS:**
- Explore more components in `components/elements.md`
- Learn form validation in `components/forms.md`
- Study composition patterns in `patterns/composition.md`
- Check out real examples in `reference/examples/`
<!-- LLM-NEXT-STEPS-END -->

## 8. Development Best Practices

<!-- LLM-BEST-PRACTICES-START -->
**FILE ORGANIZATION:**
- Keep components small and focused (single responsibility)
- Use consistent naming: `ButtonProps`, `InputProps`, etc.
- Group related components in same package
- Document component props with struct tags

**PERFORMANCE TIPS:**
- Use `templ generate --watch` during development
- Leverage browser caching for static assets
- Minimize Alpine.js state complexity
- Use HTMX for server interactions instead of large JavaScript

**CODE QUALITY:**
- Validate props with struct tags
- Use helper functions for complex logic
- Follow Go naming conventions
- Add comments for complex components
<!-- LLM-BEST-PRACTICES-END -->

## References

<!-- LLM-REFERENCES-START -->
**INCLUDED REFERENCE FILES:**
- `templ-llms.md` - Advanced Templ features (streaming, suspense patterns)
- `flowbite-llms-full.txt` - Complete Flowbite component catalog with examples

**OFFICIAL DOCUMENTATION:**
- [Templ Guide](https://templ.guide) - Complete templating language reference
- [HTMX Documentation](https://htmx.org/docs/) - Hypermedia interaction patterns
- [Alpine.js Guide](https://alpinejs.dev/start-here) - Reactive JavaScript framework
- [Flowbite Components](https://flowbite.com/docs/getting-started/introduction/) - UI component library
- [TailwindCSS](https://tailwindcss.com/docs) - Utility-first CSS framework

**COMMUNITY RESOURCES:**
- [Templ Examples](https://github.com/a-h/templ/tree/main/examples) - Official examples
- [HTMX Examples](https://htmx.org/examples/) - Common interaction patterns
- [Alpine.js Examples](https://alpinejs.dev/start-here) - Reactive patterns
<!-- LLM-REFERENCES-END -->