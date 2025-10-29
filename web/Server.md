# AWO ERP UI System - Server Integration Guide

**Version**: 1.0.0  
**Last Updated**: October 2024  
**Technology Stack**: Go Templ + Fiber + HTMX + Alpine.js + Flowbite  

---

## 🎯 Overview

This document provides a comprehensive guide for integrating the AWO ERP UI system with your existing Fiber server infrastructure. The UI system is built using **templ** (type-safe Go templates) with **HTMX** for server-driven interactions.

## 📁 Current UI System Structure

```
web/
├── assets/js/           # JavaScript components (37.1KB total)
├── components/          # 40+ reusable UI components
│   ├── atoms/          # Basic elements (button, input, icon)
│   ├── molecules/      # Composite components (card, form, alert)
│   └── organisms/      # Complex components (table, dropdown, modal)
├── layout/             # Layout system (base, nav, header, footer, app)
├── pages/              # Page templates
│   ├── auth/          # Login, signup pages
│   ├── dashboard.templ # Main dashboard
│   └── tenant/        # Tenant CRUD pages (list, create, view)
└── utils/             # Helper utilities
```

## 🏗️ Architecture Principles

### **1. Server-First Approach**
- **Business logic stays on Go backend**
- **Progressive enhancement** - works without JavaScript
- **HTMX** for server-driven UI updates
- **Alpine.js** for minimal client-side reactivity

### **2. Type Safety Throughout**
- **Templ** provides compile-time template validation
- **Go structs** define component props with full IDE support
- **Zero runtime template errors**

### **3. Component-Based Design**
- **Atomic Design**: Atoms → Molecules → Organisms → Templates → Pages
- **Reusable components** with consistent APIs
- **Props-driven configuration** - no hardcoded values

### **4. Enterprise Features**
- **Multi-tenant ready** - context-aware components
- **Accessibility** (WCAG 2.1 AA compliant)
- **Security** - XSS prevention, CSP compliance
- **Performance** - 37.1KB JS bundle, server-side rendering

---

## 🔧 Integration Requirements

### **1. Static Asset Serving**

Your Fiber app needs to serve CSS and JavaScript assets:

```go
// Add to your Fiber app configuration
app.Static("/assets", "./web/assets", fiber.Static{
    Compress:      true,
    ByteRange:     true,
    Browse:        false,
    CacheDuration: 24 * time.Hour,
})

// Serve compiled CSS (you'll need to generate this)
app.Static("/css", "./web/dist/css")
```

### **2. Required Assets**

**CSS:**
- TailwindCSS compilation required
- Flowbite CSS for component styling

**JavaScript:**
- Alpine.js for reactivity
- HTMX for server interactions  
- Component-specific JS files (optional)

### **3. Content-Type Headers**

Ensure proper content types for assets:

```go
app.Use(func(c *fiber.Ctx) error {
    path := c.Path()
    if strings.HasSuffix(path, ".js") {
        c.Set("Content-Type", "application/javascript")
    } else if strings.HasSuffix(path, ".css") {
        c.Set("Content-Type", "text/css")
    }
    return c.Next()
})
```

---

## 🎮 Page Examples & Route Integration

### **1. Dashboard Page**

**Template**: `web/pages/dashboard.templ`

**Route Handler Example**:
```go
func DashboardHandler(c *fiber.Ctx) error {
    // Get user context
    user := getUserFromContext(c)
    
    // Prepare dashboard data
    props := pages.DashboardProps{
        UserName:    user.Name,
        CompanyName: user.Company,
        Metrics: []pages.DashboardMetric{
            {
                Title:       "Revenue",
                Value:       "$125,430",
                Change:      "+12.5%",
                ChangeType:  "positive",
                Icon:        "dollar-sign",
                Description: "vs last month",
            },
            // ... more metrics
        },
        Activities: getRecentActivities(),
    }
    
    // Render template
    return pages.Dashboard(props).Render(c.Context(), c.Response().BodyWriter())
}
```

### **2. Tenant CRUD Pages**

**Templates**: 
- `web/pages/tenant/list.templ` - List tenants with pagination
- `web/pages/tenant/create.templ` - Create new tenant
- `web/pages/tenant/view.templ` - View tenant details

**Route Handlers**:
```go
// List tenants
func TenantsListHandler(c *fiber.Ctx) error {
    // Parse pagination and filters
    page := c.QueryInt("page", 1)
    status := c.Query("status")
    name := c.Query("name")
    
    // Get tenants from service
    tenants, pagination := tenantService.GetTenants(c.Context(), 
        page, 20, status, name)
    
    props := tenant.TenantListProps{
        Tenants:       tenants,
        Pagination:    pagination,
        StatusFilter:  status,
        NameFilter:    name,
        UserName:      getUserName(c),
        CompanyName:   getCompanyName(c),
        CanCreate:     hasPermission(c, "tenant:create"),
        CanEdit:       hasPermission(c, "tenant:edit"),
        CanDelete:     hasPermission(c, "tenant:delete"),
    }
    
    return tenant.TenantList(props).Render(c.Context(), c.Response().BodyWriter())
}

// Create tenant (GET - show form)
func TenantCreateGetHandler(c *fiber.Ctx) error {
    props := tenant.TenantCreateProps{
        UserName:    getUserName(c),
        CompanyName: getCompanyName(c),
        FormData:    tenant.TenantFormData{}, // Empty form
    }
    
    return tenant.TenantCreate(props).Render(c.Context(), c.Response().BodyWriter())
}

// Create tenant (POST - process form)
func TenantCreatePostHandler(c *fiber.Ctx) error {
    var req tenant.TenantFormData
    if err := c.BodyParser(&req); err != nil {
        // Return form with errors
        return renderTenantCreateWithErrors(c, req, err)
    }
    
    // Validate and create tenant
    createdTenant, err := tenantService.CreateTenant(c.Context(), req)
    if err != nil {
        return renderTenantCreateWithErrors(c, req, err)
    }
    
    // Redirect to tenant view
    return c.Redirect("/tenants/" + createdTenant.ID)
}
```

### **3. Authentication Pages**

**Templates**:
- `web/pages/auth/login.templ` - Login form
- `web/pages/auth/signup.templ` - Registration form

**Route Handler Example**:
```go
func LoginHandler(c *fiber.Ctx) error {
    props := auth.LoginProps{
        CompanyName: "AWO ERP",
        Logo:        "/assets/images/logo.png",
        RedirectURL: c.Query("redirect"),
        ShowSignup:  true,
        SignupURL:   "/auth/signup",
        ForgotURL:   "/auth/forgot-password",
    }
    
    return auth.Login(props).Render(c.Context(), c.Response().BodyWriter())
}
```

---

## 🔄 HTMX Integration Patterns

### **1. Form Submissions**

Templates include HTMX attributes for dynamic form handling:

```html
<form 
    hx-post="/tenants"
    hx-target="#form-container"
    hx-swap="outerHTML"
    hx-indicator="#submit-spinner"
>
    <!-- Form fields -->
</form>
```

**Server Response**:
```go
func CreateTenantHTMX(c *fiber.Ctx) error {
    // Process form
    if err := processForm(c); err != nil {
        // Return form with errors (HTMX will replace the form)
        return renderFormWithErrors(c, err)
    }
    
    // Success - redirect or return success message
    c.Set("HX-Redirect", "/tenants")
    return c.SendStatus(200)
}
```

### **2. Table Updates**

Tables can be updated dynamically:

```html
<div id="tenant-table">
    <!-- Table content -->
</div>

<!-- Pagination links update the table -->
<button hx-get="/tenants?page=2" hx-target="#tenant-table" hx-swap="outerHTML">
    Next Page
</button>
```

### **3. Search and Filters**

Real-time search without page reload:

```html
<form hx-get="/tenants" hx-target="#tenant-table" hx-trigger="change, submit">
    <input name="search" hx-trigger="keyup changed delay:300ms" />
</form>
```

---

## 🎨 Styling and Assets

### **1. TailwindCSS Configuration**

You'll need to compile TailwindCSS with your components:

```js
// tailwind.config.js
module.exports = {
  content: [
    "./web/**/*.templ",
    "./web/**/*.go",
  ],
  theme: {
    extend: {
      colors: {
        border: "hsl(var(--border))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        // ... design system colors
      }
    }
  },
  plugins: [
    require("@tailwindcss/forms"),
    require("flowbite/plugin")
  ]
}
```

### **2. CSS Build Process**

```bash
# Install dependencies
npm install -D tailwindcss @tailwindcss/forms flowbite

# Build CSS
npx tailwindcss -i ./web/assets/css/input.css -o ./web/dist/css/tailwind.css --watch
```

### **3. Theme System**

The UI supports light/dark themes with CSS variables:

```css
:root {
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
  --primary: 222.2 47.4% 11.2%;
  /* ... more variables */
}

.dark {
  --background: 222.2 84% 4.9%;
  --foreground: 210 40% 98%;
  --primary: 210 40% 98%;
  /* ... dark mode variables */
}
```

---

## 🔒 Security Considerations

### **1. CSRF Protection**

Include CSRF tokens in forms:

```go
// Middleware to inject CSRF token
func CSRFMiddleware(c *fiber.Ctx) error {
    token := generateCSRFToken()
    c.Locals("csrf_token", token)
    return c.Next()
}

// In templates
<input type="hidden" name="_token" value="{{ .csrf_token }}">
```

### **2. Content Security Policy**

Configure CSP headers for HTMX and Alpine.js:

```go
app.Use(func(c *fiber.Ctx) error {
    c.Set("Content-Security-Policy", 
        "default-src 'self'; " +
        "script-src 'self' 'unsafe-inline'; " + // Alpine.js needs inline scripts
        "style-src 'self' 'unsafe-inline'; " +  // Tailwind needs inline styles
        "img-src 'self' data: https:;")
    return c.Next()
})
```

### **3. XSS Prevention**

Templ automatically escapes content:

```go
templ UserProfile(user User) {
    <h1>{ user.Name }</h1>  <!-- Automatically escaped -->
}
```

---

## 📊 Performance Optimizations

### **1. Asset Optimization**

- **JavaScript**: 37.1KB total (compressed)
- **CSS**: Use PurgeCSS with TailwindCSS
- **Images**: Optimize and use WebP format
- **Caching**: Set appropriate cache headers

### **2. Server-Side Rendering**

Templates render on the server:

```go
// Pre-compile templates for better performance
//go:embed web/pages/*.templ
var templateFS embed.FS

// Use with Fiber
func RenderTemplate(name string, data interface{}) fiber.Handler {
    return func(c *fiber.Ctx) error {
        return templates[name].Render(c.Context(), c.Response().BodyWriter())
    }
}
```

### **3. Partial Updates**

Use HTMX for partial page updates:

- Update table rows instead of entire tables
- Replace form sections on validation errors
- Update navigation badges dynamically

---

## 🚀 Getting Started Checklist

### **Phase 1: Basic Setup**
- [ ] Add static asset serving to Fiber
- [ ] Set up TailwindCSS compilation
- [ ] Create basic route handlers
- [ ] Test dashboard page

### **Phase 2: Authentication**
- [ ] Implement login/signup handlers
- [ ] Add session middleware
- [ ] Configure CSRF protection
- [ ] Test authentication flow

### **Phase 3: CRUD Operations**
- [ ] Implement tenant CRUD handlers
- [ ] Add HTMX endpoints
- [ ] Test form submissions
- [ ] Add validation and error handling

### **Phase 4: Advanced Features**
- [ ] Add search and filtering
- [ ] Implement pagination
- [ ] Add real-time updates
- [ ] Optimize performance

---

## 🔧 Example Route Registration

Here's how to register UI routes in your existing Fiber setup:

```go
// In your routes.go file
func RegisterUIRoutes(app *fiber.App) {
    // Static assets
    app.Static("/assets", "./web/assets")
    app.Static("/css", "./web/dist/css")
    
    // Authentication routes
    auth := app.Group("/auth")
    auth.Get("/login", handlers.LoginGetHandler)
    auth.Post("/login", handlers.LoginPostHandler)
    auth.Get("/signup", handlers.SignupGetHandler)
    auth.Post("/signup", handlers.SignupPostHandler)
    
    // Protected routes
    protected := app.Group("/", middleware.AuthRequired())
    
    // Dashboard
    protected.Get("/", handlers.DashboardHandler)
    protected.Get("/dashboard", handlers.DashboardHandler)
    
    // Tenants
    tenants := protected.Group("/tenants")
    tenants.Get("/", handlers.TenantsListHandler)
    tenants.Get("/new", handlers.TenantCreateGetHandler)
    tenants.Post("/", handlers.TenantCreatePostHandler)
    tenants.Get("/:id", handlers.TenantViewHandler)
    tenants.Get("/:id/edit", handlers.TenantEditGetHandler)
    tenants.Put("/:id", handlers.TenantUpdateHandler)
    tenants.Delete("/:id", handlers.TenantDeleteHandler)
}
```

---

## 📚 Additional Resources

### **Components Reference**
- **Layout System**: `web/layout/` - Base, navigation, header, footer
- **UI Components**: `web/components/` - 40+ reusable components
- **Page Templates**: `web/pages/` - Complete page examples

### **Documentation**
- **[Templ Guide](https://templ.guide)** - Template language documentation
- **[HTMX Docs](https://htmx.org/docs/)** - Server interaction patterns
- **[Alpine.js Guide](https://alpinejs.dev)** - Client-side reactivity
- **[Flowbite Components](https://flowbite.com/docs/components/)** - UI library

### **Template Generation**
```bash
# Generate Go code from templates
templ generate web/

# Watch for changes during development
templ generate --watch web/
```

---

**This UI system is production-ready and enterprise-grade, designed to integrate seamlessly with your existing Fiber architecture while providing modern, accessible, and performant user interfaces.**