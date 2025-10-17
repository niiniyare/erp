# Templ Integration Guide

**FILE PURPOSE**: Complete guide to Templ templating in ERP UI system  
**SCOPE**: Templ syntax, integration patterns, performance optimization  
**TARGET AUDIENCE**: Go developers, template authors, component builders

## 🎯 Templ Overview

[Templ](https://templ.guide) is a type-safe, compile-time Go templating language that generates Go code from template files. It provides the foundation for our server-side rendering architecture.

### Why Templ for ERP?
- **Type Safety**: Compile-time validation of templates and props
- **Performance**: No runtime template parsing, compiled Go code
- **Integration**: Seamless integration with Go backend
- **Security**: Automatic HTML escaping and XSS prevention
- **Tooling**: IDE support, hot reload, error detection

## 🏗️ Templ Architecture in ERP

### Template Compilation Pipeline
```
.templ files → templ generate → .go files → go build → Binary
```

### Integration with Schema System
```go
// Schema factory generates Templ components
factory, _ := engine.NewSchemaFactory("docs/ui/Schema")
component, _ := factory.RenderToTempl(ctx, "ButtonSchema", props)

// Use in Templ template
templ Page() {
    <div>
        @component  // Schema-generated component
    </div>
}
```

## 📝 Templ Syntax Guide

### Basic Template Structure
```go
// Define a template with typed parameters
templ ComponentName(props ComponentProps) {
    <div class={ getClasses(props) }>
        { props.Title }
    </div>
}

// Template with conditional rendering
templ Button(text string, disabled bool) {
    <button 
        disabled?={ disabled }
        class="btn btn-primary">
        { text }
    </button>
}
```

### Props and Type Safety
```go
// Define props structure
type ButtonProps struct {
    Text     string
    Variant  ButtonVariant
    Size     ButtonSize
    Disabled bool
    OnClick  string
    ID       string
    Class    string
}

// Use in template
templ Button(props ButtonProps) {
    <button 
        type="button"
        id={ props.ID }
        class={ buildButtonClasses(props) }
        disabled?={ props.Disabled }
        onclick={ props.OnClick }>
        { props.Text }
    </button>
}
```

### Conditional Rendering
```go
templ UserProfile(user User, isOwner bool) {
    <div class="profile">
        <h1>{ user.Name }</h1>
        
        // Conditional block
        if isOwner {
            <div class="admin-actions">
                <button>Edit Profile</button>
                <button>Delete Account</button>
            </div>
        }
        
        // Conditional attribute
        <img 
            src={ user.Avatar }
            class={ "avatar", templ.KV("admin", isOwner) }>
    </div>
}
```

### Loops and Iteration
```go
templ TableRows(items []TableItem) {
    for _, item := range items {
        <tr>
            <td>{ item.Name }</td>
            <td>{ item.Value }</td>
            <td>
                if item.Actions != nil {
                    for _, action := range item.Actions {
                        @ActionButton(action)
                    }
                }
            </td>
        </tr>
    }
}
```

## 🔧 Component Composition Patterns

### Slot Pattern (Children)
```go
// Parent component that accepts children
templ Card(title string) {
    <div class="card">
        <div class="card-header">
            <h3>{ title }</h3>
        </div>
        <div class="card-body">
            { children... }  // Slot for child content
        </div>
    </div>
}

// Usage with children
templ ProfileCard(user User) {
    @Card("User Profile") {
        <p>Name: { user.Name }</p>
        <p>Email: { user.Email }</p>
        @EditButton(user.ID)
    }
}
```

### Layout Pattern
```go
// Base layout template
templ BaseLayout(title, description string) {
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <title>{ title }</title>
        <meta name="description" content={ description }>
        <link href="/assets/styles.css" rel="stylesheet">
        <script src="https://unpkg.com/htmx.org@1.9.10"></script>
        <script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
    </head>
    <body>
        @Header()
        <main>
            { children... }
        </main>
        @Footer()
    </body>
    </html>
}

// Page using layout
templ DashboardPage(data DashboardData) {
    @BaseLayout("Dashboard", "ERP Dashboard Overview") {
        <div class="dashboard">
            @StatsCards(data.Stats)
            @RecentActivity(data.Activities)
        </div>
    }
}
```

### Component Factory Pattern
```go
// Dynamic component rendering
templ RenderComponent(componentType string, props map[string]interface{}) {
    switch componentType {
        case "button":
            @Button(props["text"].(string), props["variant"].(string))
        case "input":
            @Input(props["name"].(string), props["value"].(string))
        case "card":
            @Card(props["title"].(string)) {
                { templ.Raw(props["content"].(string)) }
            }
        default:
            <div class="error">Unknown component: { componentType }</div>
    }
}
```

## 🚀 Performance Optimization

### HTTP Streaming
```go
// Enable streaming for faster TTFB
func handleDashboard(w http.ResponseWriter, r *http.Request) {
    // Use streaming handler
    templ.Handler(DashboardPage(data), templ.WithStreaming()).ServeHTTP(w, r)
}

// Template with streaming
templ DashboardPage(userID string) {
    @BaseLayout("Dashboard", "ERP Dashboard") {
        // Render immediately available content
        @Header()
        @Navigation()
        
        // Force flush to client
        @templ.Flush()
        
        // Render data that requires database queries
        @LoadUserData(userID)
        @LoadRecentActivities(userID)
    }
}
```

### Suspense Pattern
```go
// Render placeholder while loading data
templ UserDashboard(userID string) {
    <div class="dashboard">
        @templ.Flush()  // Send initial content
        
        // Show loading state
        <div id="user-stats" class="loading">
            <div class="spinner"></div>
            <p>Loading user statistics...</p>
        </div>
        
        @templ.Flush()  // Send loading state
        
        // Replace with actual data (using HTMX)
        <div 
            hx-get={ "/api/users/" + userID + "/stats" }
            hx-target="#user-stats"
            hx-swap="outerHTML"
            hx-trigger="load">
        </div>
    </div>
}
```

### Component Memoization
```go
// Cache expensive components
var componentCache = make(map[string]templ.Component)

func CachedComponent(key string, generator func() templ.Component) templ.Component {
    if cached, exists := componentCache[key]; exists {
        return cached
    }
    
    component := generator()
    componentCache[key] = component
    return component
}

// Usage
templ ExpensiveChart(data ChartData) {
    @CachedComponent("chart-"+data.ID, func() templ.Component {
        return RenderChart(data)
    })
}
```

## 🔗 HTMX Integration Patterns

### Form Handling
```go
templ LoginForm() {
    <form 
        hx-post="/api/auth/login"
        hx-target="#login-result"
        hx-swap="innerHTML"
        hx-indicator="#login-spinner">
        
        @Input(InputProps{
            Name:        "username",
            Type:        "text",
            Placeholder: "Username",
            Required:    true,
        })
        
        @Input(InputProps{
            Name:        "password",
            Type:        "password",
            Placeholder: "Password",
            Required:    true,
        })
        
        @Button(ButtonProps{
            Text:    "Login",
            Type:    "submit",
            Variant: ButtonPrimary,
        })
        
        <div id="login-spinner" class="htmx-indicator">
            @Spinner()
        </div>
    </form>
    
    <div id="login-result"></div>
}
```

### Dynamic Content Updates
```go
templ UserList() {
    <div id="user-list">
        @UserTable()
        
        // Auto-refresh every 30 seconds
        <div 
            hx-get="/api/users/table"
            hx-target="#user-list"
            hx-swap="innerHTML"
            hx-trigger="every 30s">
        </div>
    </div>
}

templ UserTable() {
    <table class="users-table">
        <thead>
            <tr>
                <th>Name</th>
                <th>Email</th>
                <th>Status</th>
                <th>Actions</th>
            </tr>
        </thead>
        <tbody>
            // Table content with HTMX actions
            for _, user := range users {
                @UserRow(user)
            }
        </tbody>
    </table>
}
```

### Modal Dialogs
```go
templ UserEditModal(user User) {
    <div 
        class="modal-overlay"
        x-data="{ open: true }"
        x-show="open"
        @click.away="open = false">
        
        <div class="modal">
            <div class="modal-header">
                <h2>Edit User</h2>
                <button @click="open = false">&times;</button>
            </div>
            
            <form 
                hx-put={ "/api/users/" + user.ID }
                hx-target="#user-list"
                hx-swap="innerHTML">
                
                @Input(InputProps{
                    Name:  "name",
                    Value: user.Name,
                    Label: "Name",
                })
                
                @Input(InputProps{
                    Name:  "email",
                    Value: user.Email,
                    Label: "Email",
                    Type:  "email",
                })
                
                <div class="modal-actions">
                    @Button(ButtonProps{
                        Text:    "Save",
                        Type:    "submit",
                        Variant: ButtonPrimary,
                    })
                    
                    @Button(ButtonProps{
                        Text:    "Cancel",
                        Variant: ButtonSecondary,
                        OnClick: "open = false",
                    })
                </div>
            </form>
        </div>
    </div>
}
```

## 🎨 Alpine.js Integration

### Client State Management
```go
templ SearchableTable() {
    <div x-data="{ 
        search: '', 
        sortBy: 'name', 
        sortDesc: false,
        filteredItems: [] 
    }">
        
        // Search input
        @Input(InputProps{
            Name:        "search",
            Placeholder: "Search users...",
            XModel:      "search",
        })
        
        // Sort controls
        <div class="sort-controls">
            <select x-model="sortBy">
                <option value="name">Name</option>
                <option value="email">Email</option>
                <option value="created">Created</option>
            </select>
            
            <button @click="sortDesc = !sortDesc">
                <span x-text="sortDesc ? '↓' : '↑'"></span>
            </button>
        </div>
        
        // Table with Alpine.js filtering
        <table>
            <tbody>
                <template x-for="item in filteredItems" :key="item.id">
                    @UserRowTemplate()
                </template>
            </tbody>
        </table>
    </div>
}
```

### Form Validation
```go
templ UserForm() {
    <form x-data="userForm()">
        @Input(InputProps{
            Name:     "name",
            Label:    "Name",
            Required: true,
            XModel:   "name",
            XBind:    ":class": "{ 'error': errors.name }",
        })
        
        <div x-show="errors.name" class="error-message">
            <span x-text="errors.name"></span>
        </div>
        
        @Button(ButtonProps{
            Text:     "Submit",
            Type:     "submit",
            Disabled: "!isValid",
            XBind:    ":disabled": "!isValid",
        })
    </form>
    
    <script>
        function userForm() {
            return {
                name: '',
                errors: {},
                get isValid() {
                    return this.name.length >= 2 && Object.keys(this.errors).length === 0;
                },
                validateName() {
                    if (this.name.length < 2) {
                        this.errors.name = 'Name must be at least 2 characters';
                    } else {
                        delete this.errors.name;
                    }
                }
            }
        }
    </script>
}
```

## 🔒 Security and Safety

### XSS Prevention
```go
// Automatic HTML escaping
templ UserProfile(user User) {
    <div>
        // Automatically escaped
        <h1>{ user.Name }</h1>
        
        // Raw HTML (use with caution)
        <div>
            { templ.Raw(user.SafeHTMLBio) }
        </div>
        
        // URL escaping
        <a href={ templ.SafeURL(user.Website) }>Website</a>
    </div>
}
```

### CSRF Protection
```go
templ SecureForm(csrfToken string) {
    <form method="POST" action="/api/users">
        <input type="hidden" name="csrf_token" value={ csrfToken }>
        
        @Input(InputProps{
            Name: "name",
            Label: "Name",
        })
        
        @Button(ButtonProps{
            Text: "Submit",
            Type: "submit",
        })
    </form>
}
```

## 🧪 Testing Templ Components

### Unit Testing
```go
func TestButtonComponent(t *testing.T) {
    props := ButtonProps{
        Text:    "Test Button",
        Variant: ButtonPrimary,
        Size:    ButtonMD,
    }
    
    // Render to string
    html, err := templ.ToGoHTML(context.Background(), Button(props))
    require.NoError(t, err)
    
    // Assertions
    assert.Contains(t, html, "Test Button")
    assert.Contains(t, html, "btn-primary")
    assert.Contains(t, html, `type="button"`)
}
```

### Integration Testing
```go
func TestUserForm(t *testing.T) {
    user := User{
        ID:    "123",
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    // Render form
    form := UserEditForm(user)
    html, _ := templ.ToGoHTML(context.Background(), form)
    
    // Test form structure
    assert.Contains(t, html, `name="name"`)
    assert.Contains(t, html, `value="John Doe"`)
    assert.Contains(t, html, `hx-put="/api/users/123"`)
}
```

## 🛠️ Development Workflow

### Live Reload Setup
```bash
# Install templ CLI
go install github.com/a-h/templ/cmd/templ@latest

# Watch for changes and regenerate
templ generate --watch

# Or use with air for full hot reload
air
```

### Build Integration
```makefile
# Makefile targets
.PHONY: templ
templ:
	templ generate

.PHONY: build
build: templ
	go build -o bin/server cmd/server/main.go

.PHONY: dev
dev: templ
	air
```

## 📚 Best Practices

### Template Organization
```
web/
├── components/
│   ├── atoms/          # Basic elements
│   ├── molecules/      # Composite components  
│   ├── organisms/      # Complex components
│   └── templates/      # Page layouts
├── pages/              # Full page templates
└── layouts/            # Base layouts
```

### Error Handling
```go
templ SafeComponent(data Data, err error) {
    if err != nil {
        <div class="error">
            <p>Error loading component</p>
            <details>
                <summary>Details</summary>
                <pre>{ err.Error() }</pre>
            </details>
        </div>
    } else {
        @ActualComponent(data)
    }
}
```

### Performance Tips
1. **Use streaming** for long-running operations
2. **Cache expensive components** with memoization
3. **Minimize template complexity** - move logic to Go functions
4. **Use proper HTMX patterns** for efficient updates
5. **Profile template rendering** in production

## 📚 Related Documentation

- **[Architecture](architecture.md)**: System design principles
- **[Schema System](schema-system.md)**: Type-safe component generation
- **[Styling Approach](styling-approach.md)**: CSS and styling patterns
- **[Component Lifecycle](component-lifecycle.md)**: Component creation process

Templ provides the foundation for type-safe, performant server-side rendering in our ERP system, enabling rapid development while maintaining security and reliability.