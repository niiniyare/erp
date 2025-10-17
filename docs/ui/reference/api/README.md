# API Reference

<!-- LLM-CONTEXT-START -->
**FILE PURPOSE**: Complete API reference for Templ components, handlers, and integration patterns
**SCOPE**: Function signatures, type definitions, configuration options, server endpoints
**TARGET AUDIENCE**: Developers implementing UI components and server handlers
**RELATED FILES**: `components/elements.md` (component usage), `patterns/htmx-integration.md` (server patterns)
<!-- LLM-CONTEXT-END -->

## Overview

This reference provides complete API documentation for the ERP UI system built with **Templ + HTMX + Alpine.js + Flowbite**. It covers component signatures, server handler patterns, and configuration options.

<!-- LLM-API-STRUCTURE-START -->
**API REFERENCE STRUCTURE:**
```
├── Component APIs       # Templ component function signatures
├── Server Handler APIs  # Go HTTP handler patterns  
├── HTMX Endpoint APIs   # Server endpoint specifications
├── Alpine.js APIs       # JavaScript component interfaces
├── Configuration APIs   # System configuration options
└── Type Definitions     # Complete type reference
```
<!-- LLM-API-STRUCTURE-END -->

---

## Component APIs

### Button Component

<!-- LLM-BUTTON-API-START -->
**FUNCTION SIGNATURE:**
```go
func Button(props ButtonProps) templ.Component
```

**TYPE DEFINITION:**
```go
type ButtonProps struct {
    Text        string            // Button text content
    Variant     ButtonVariant     // Visual style variant
    Size        ButtonSize        // Size variant
    Type        ButtonType        // HTML button type
    Disabled    bool              // Disable interaction
    Loading     bool              // Show loading state
    FullWidth   bool              // Expand to container width
    Icon        IconProps         // Optional icon configuration
    OnClick     string            // JavaScript click handler
    ID          string            // HTML id attribute
    Class       string            // Additional CSS classes
    AriaLabel   string            // Accessibility label
    DataAttrs   map[string]string // Custom data attributes
}

type ButtonVariant string
const (
    ButtonPrimary   ButtonVariant = "primary"
    ButtonSecondary ButtonVariant = "secondary"
    ButtonSuccess   ButtonVariant = "success"
    ButtonDanger    ButtonVariant = "danger"
    ButtonWarning   ButtonVariant = "warning"
    ButtonInfo      ButtonVariant = "info"
    ButtonLight     ButtonVariant = "light"
    ButtonDark      ButtonVariant = "dark"
)

type ButtonSize string
const (
    ButtonXS   ButtonSize = "xs"
    ButtonSM   ButtonSize = "sm"
    ButtonBase ButtonSize = "base"
    ButtonLG   ButtonSize = "lg"
    ButtonXL   ButtonSize = "xl"
)

type ButtonType string
const (
    ButtonTypeButton ButtonType = "button"
    ButtonTypeSubmit ButtonType = "submit"
    ButtonTypeReset  ButtonType = "reset"
)
```

**USAGE EXAMPLES:**
```go
// Basic usage
@Button(ButtonProps{
    Text: "Save",
    Variant: ButtonPrimary,
    Type: ButtonTypeSubmit,
})

// With loading state
@Button(ButtonProps{
    Text: "Processing...",
    Variant: ButtonPrimary,
    Loading: true,
    Disabled: true,
})

// With custom attributes
@Button(ButtonProps{
    Text: "Delete",
    Variant: ButtonDanger,
    OnClick: "confirmDelete()",
    DataAttrs: map[string]string{
        "hx-delete": "/api/users/123",
        "hx-confirm": "Are you sure?",
    },
})
```
<!-- LLM-BUTTON-API-END -->

### Input Component

<!-- LLM-INPUT-API-START -->
**FUNCTION SIGNATURE:**
```go
func Input(props InputProps) templ.Component
```

**TYPE DEFINITION:**
```go
type InputProps struct {
    Type         InputType         // Input field type
    Value        string            // Input field value
    Placeholder  string            // Placeholder text
    Name         string            // Form field name
    Label        string            // Associated label text
    HelperText   string            // Help or error message
    Size         InputSize         // Size variant
    Validation   ValidationState   // Validation state
    Required     bool              // HTML5 required attribute
    Disabled     bool              // Disable input
    ReadOnly     bool              // Make read-only
    MaxLength    int               // Character limit
    MinLength    int               // Minimum character requirement
    Pattern      string            // HTML5 pattern attribute
    ID           string            // HTML id attribute
    Class        string            // Additional CSS classes
    Autocomplete string            // Autocomplete hint
    OnChange     string            // JavaScript change handler
    OnFocus      string            // JavaScript focus handler
    OnBlur       string            // JavaScript blur handler
    OnInput      string            // JavaScript input handler
    DataAttrs    map[string]string // Custom data attributes
}

type InputType string
const (
    InputText     InputType = "text"
    InputEmail    InputType = "email"
    InputPassword InputType = "password"
    InputNumber   InputType = "number"
    InputTel      InputType = "tel"
    InputURL      InputType = "url"
    InputSearch   InputType = "search"
    InputDate     InputType = "date"
    InputTime     InputType = "time"
    InputDateTime InputType = "datetime-local"
)

type InputSize string
const (
    InputSM   InputSize = "sm"
    InputBase InputSize = "base"
    InputLG   InputSize = "lg"
)

type ValidationState string
const (
    ValidationNone    ValidationState = "none"
    ValidationSuccess ValidationState = "success"
    ValidationError   ValidationState = "error"
)
```
<!-- LLM-INPUT-API-END -->

### Form Component

<!-- LLM-FORM-API-START -->
**FUNCTION SIGNATURE:**
```go
func Form(props FormProps) templ.Component
```

**TYPE DEFINITION:**
```go
type FormProps struct {
    Action      string            // Form action URL
    Method      string            // HTTP method
    Fields      []FormField       // Form field definitions
    Validation  ValidationConfig  // Validation configuration
    CSRF        string            // CSRF token
    Encoding    string            // Form encoding type
    NoValidate  bool              // Disable HTML5 validation
    OnSubmit    string            // JavaScript submit handler
    OnReset     string            // JavaScript reset handler
    ID          string            // HTML id attribute
    Class       string            // Additional CSS classes
    DataAttrs   map[string]string // Custom data attributes
}

type FormField struct {
    Type       string                 // Field type
    Name       string                 // Field name
    Label      string                 // Field label
    Value      interface{}            // Field value
    Props      map[string]interface{} // Field-specific properties
    Validation FieldValidation        // Field validation rules
}

type ValidationConfig struct {
    Enabled     bool                    // Enable validation
    Strategy    string                  // "submit", "blur", "change"
    Rules       map[string][]Rule       // Validation rules by field
    Messages    map[string]string       // Custom error messages
    OnValidate  string                  // JavaScript validation handler
}

type FieldValidation struct {
    Required bool     // Field is required
    MinLength int     // Minimum length
    MaxLength int     // Maximum length
    Pattern   string  // Regex pattern
    Custom    string  // Custom validation function
}
```
<!-- LLM-FORM-API-END -->

---

## Server Handler APIs

### HTMX Response Handlers

<!-- LLM-HTMX-HANDLERS-START -->
**RESPONSE HELPER FUNCTIONS:**
```go
// HTMX response helpers
func HTMXResponse(w http.ResponseWriter, component templ.Component) error
func HTMXTrigger(w http.ResponseWriter, events map[string]interface{})
func HTMXRedirect(w http.ResponseWriter, url string)
func HTMXRefresh(w http.ResponseWriter)

// Validation response helpers
func ValidationErrorResponse(w http.ResponseWriter, errors map[string]string) error
func SuccessResponse(w http.ResponseWriter, message string) error
func ErrorResponse(w http.ResponseWriter, message string, code int) error
```

**HANDLER PATTERN:**
```go
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func WithHTMX(handler HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := handler(w, r); err != nil {
            log.Printf("Handler error: %v", err)
            ErrorResponse(w, "Internal server error", 500)
        }
    }
}

// Example usage
func CreateUserHandler(w http.ResponseWriter, r *http.Request) error {
    if r.Method != "POST" {
        return errors.New("method not allowed")
    }
    
    var user User
    if err := parseForm(r, &user); err != nil {
        return ValidationErrorResponse(w, map[string]string{
            "email": "Invalid email format",
        })
    }
    
    if err := userService.Create(user); err != nil {
        return err
    }
    
    HTMXTrigger(w, map[string]interface{}{
        "userCreated": map[string]string{
            "message": "User created successfully",
            "type": "success",
        },
    })
    
    return HTMXResponse(w, UserCard(user))
}
```
<!-- LLM-HTMX-HANDLERS-END -->

### Validation Handlers

<!-- LLM-VALIDATION-HANDLERS-START -->
**VALIDATION API:**
```go
type Validator struct {
    rules map[string][]ValidationRule
    lang  string
}

type ValidationRule interface {
    Validate(value interface{}) ValidationResult
    Message() string
}

type ValidationResult struct {
    Valid   bool   `json:"valid"`
    Message string `json:"message,omitempty"`
    Code    string `json:"code,omitempty"`
}

// Built-in validation rules
func Required() ValidationRule
func MinLength(min int) ValidationRule
func MaxLength(max int) ValidationRule
func Email() ValidationRule
func Pattern(regex string) ValidationRule
func Unique(repo Repository, field string) ValidationRule

// Validation middleware
func ValidationMiddleware(validator *Validator) func(http.Handler) http.Handler

// Field validation endpoint
func ValidateFieldHandler(w http.ResponseWriter, r *http.Request) error {
    field := r.URL.Query().Get("field")
    value := r.FormValue(field)
    
    result := validator.ValidateField(field, value)
    
    if !result.Valid {
        w.Header().Set("HX-Retarget", fmt.Sprintf("#%s-error", field))
        return HTMXResponse(w, ErrorMessage(result.Message))
    }
    
    return HTMXResponse(w, SuccessMessage(""))
}
```
<!-- LLM-VALIDATION-HANDLERS-END -->

---

## HTMX Endpoint APIs

### Standard Endpoints

<!-- LLM-ENDPOINTS-START -->
**ENDPOINT SPECIFICATIONS:**

#### GET /api/{entity}/search
```go
// Query parameters
type SearchParams struct {
    Q        string            `query:"q"`         // Search query
    Page     int               `query:"page"`      // Page number
    Limit    int               `query:"limit"`     // Items per page
    Sort     string            `query:"sort"`      // Sort field
    Order    string            `query:"order"`     // Sort direction
    Filters  map[string]string `query:"filters"`   // Additional filters
}

// Response: HTML table rows or cards
```

#### POST /api/{entity}
```go
// Request: Form data or JSON
// Response: Created entity HTML + HX-Trigger events
type CreateResponse struct {
    Entity    interface{}       `json:"entity"`
    Triggers  map[string]interface{} `json:"triggers"`
}
```

#### PUT /api/{entity}/{id}
```go
// Request: Form data or JSON
// Response: Updated entity HTML + HX-Trigger events
```

#### DELETE /api/{entity}/{id}
```go
// Response: Empty with HX-Trigger events
type DeleteResponse struct {
    Triggers map[string]interface{} `json:"triggers"`
}
```

#### POST /api/{entity}/bulk
```go
type BulkRequest struct {
    Action string   `json:"action"`  // "delete", "update", "export"
    IDs    []string `json:"ids"`     // Entity IDs
    Data   map[string]interface{} `json:"data,omitempty"` // Update data
}

// Response: Updated list HTML + HX-Trigger events
```

#### GET /api/{entity}/{field}/validate
```go
type ValidationRequest struct {
    Field string      `query:"field"` // Field name
    Value interface{} `form:"value"`  // Field value
}

type ValidationResponse struct {
    Valid   bool   `json:"valid"`
    Message string `json:"message,omitempty"`
}
```
<!-- LLM-ENDPOINTS-END -->

---

## Alpine.js APIs

### Global Stores

<!-- LLM-ALPINE-STORES-START -->
**STORE DEFINITIONS:**
```javascript
// Application store
Alpine.store('app', {
    notifications: [],
    loading: false,
    user: null,
    
    // Methods
    addNotification(type, message, duration = 5000) {},
    removeNotification(id) {},
    setLoading(state) {},
    setUser(user) {}
})

// Form store
Alpine.store('forms', {
    validation: {},
    dirty: {},
    
    // Methods  
    setFieldError(field, error) {},
    clearFieldError(field) {},
    setFieldDirty(field, dirty) {},
    isFormValid(formId) {},
    resetForm(formId) {}
})

// UI store
Alpine.store('ui', {
    sidebar: false,
    theme: 'light',
    modals: {},
    
    // Methods
    toggleSidebar() {},
    setTheme(theme) {},
    openModal(id) {},
    closeModal(id) {}
})
```

**COMPONENT PATTERNS:**
```javascript
// List manager component
function listManager() {
    return {
        selectedItems: [],
        selectAll: false,
        loading: false,
        
        init() {
            this.$watch('selectAll', (value) => {
                this.toggleSelectAll(value)
            })
        },
        
        toggleSelectAll(state) {},
        updateSelectedItems() {},
        bulkAction(action) {},
        deleteSelected() {},
        exportSelected() {}
    }
}

// Form handler component  
function formHandler(config = {}) {
    return {
        submitting: false,
        dirty: false,
        validation: {},
        
        init() {
            this.setupEventListeners()
        },
        
        setupEventListeners() {},
        validateField(field, value) {},
        submit() {},
        reset() {}
    }
}

// Modal component
function modalHandler(id) {
    return {
        open: false,
        
        init() {
            this.$watch('open', (value) => {
                if (value) {
                    this.onOpen()
                } else {
                    this.onClose()
                }
            })
        },
        
        show() {},
        hide() {},
        onOpen() {},
        onClose() {}
    }
}
```
<!-- LLM-ALPINE-STORES-END -->

---

## Configuration APIs

### System Configuration

<!-- LLM-CONFIG-API-START -->
**CONFIGURATION STRUCTURE:**
```go
type UIConfig struct {
    Theme        ThemeConfig        `yaml:"theme"`
    Components   ComponentConfig    `yaml:"components"`
    Validation   ValidationConfig   `yaml:"validation"`
    HTMX         HTMXConfig         `yaml:"htmx"`
    Alpine       AlpineConfig       `yaml:"alpine"`
    Performance  PerformanceConfig  `yaml:"performance"`
}

type ThemeConfig struct {
    DefaultTheme string            `yaml:"default_theme"`
    Themes       map[string]Theme  `yaml:"themes"`
    DarkMode     bool              `yaml:"dark_mode"`
}

type ComponentConfig struct {
    DefaultSize    string                    `yaml:"default_size"`
    DefaultVariant string                    `yaml:"default_variant"`
    ClassOverrides map[string]string         `yaml:"class_overrides"`
    Presets        map[string]ComponentPreset `yaml:"presets"`
}

type ValidationConfig struct {
    Strategy      string            `yaml:"strategy"`
    DebounceMs    int               `yaml:"debounce_ms"`
    ShowErrors    bool              `yaml:"show_errors"`
    ErrorDisplay  string            `yaml:"error_display"`
    CustomRules   map[string]string `yaml:"custom_rules"`
}

type HTMXConfig struct {
    Timeout       int               `yaml:"timeout"`
    RetryCount    int               `yaml:"retry_count"`
    Headers       map[string]string `yaml:"headers"`
    ErrorHandler  string            `yaml:"error_handler"`
    LoadingClass  string            `yaml:"loading_class"`
}

type PerformanceConfig struct {
    CacheTemplates bool `yaml:"cache_templates"`
    MinifyHTML     bool `yaml:"minify_html"`
    CompressAssets bool `yaml:"compress_assets"`
    LazyLoading    bool `yaml:"lazy_loading"`
}
```

**CONFIGURATION METHODS:**
```go
// Load configuration
func LoadUIConfig(path string) (*UIConfig, error)

// Apply configuration
func ApplyConfig(config *UIConfig) error

// Get configuration values
func GetTheme() string
func GetComponentDefault(component, property string) string
func GetValidationStrategy() string

// Update configuration at runtime
func SetTheme(theme string) error
func UpdateComponentDefaults(overrides map[string]string) error
```
<!-- LLM-CONFIG-API-END -->

---

## Type Definitions

### Core Types

<!-- LLM-CORE-TYPES-START -->
**FUNDAMENTAL TYPES:**
```go
// Base component interface
type Component interface {
    Render() templ.Component
    Validate() error
    Props() interface{}
}

// Common property types
type CommonProps struct {
    ID        string            `json:"id"`
    Class     string            `json:"class"`
    Style     string            `json:"style"`
    DataAttrs map[string]string `json:"data_attrs"`
    AriaAttrs map[string]string `json:"aria_attrs"`
}

// Event handler types
type EventHandler string
type EventHandlers map[string]EventHandler

// Icon configuration
type IconProps struct {
    Name     string `json:"name"`
    Size     string `json:"size"`
    Color    string `json:"color"`
    Position string `json:"position"` // "left", "right", "top", "bottom"
}

// Layout types
type LayoutProps struct {
    Container bool   `json:"container"`
    MaxWidth  string `json:"max_width"`
    Padding   string `json:"padding"`
    Margin    string `json:"margin"`
}

// Responsive configuration
type ResponsiveConfig map[string]string // breakpoint -> value

// Color system
type ColorVariant string
const (
    ColorPrimary   ColorVariant = "primary"
    ColorSecondary ColorVariant = "secondary"
    ColorSuccess   ColorVariant = "success"
    ColorDanger    ColorVariant = "danger"
    ColorWarning   ColorVariant = "warning"
    ColorInfo      ColorVariant = "info"
    ColorLight     ColorVariant = "light"
    ColorDark      ColorVariant = "dark"
)

// Size system
type SizeVariant string
const (
    SizeXS   SizeVariant = "xs"
    SizeSM   SizeVariant = "sm"
    SizeBase SizeVariant = "base"
    SizeLG   SizeVariant = "lg"
    SizeXL   SizeVariant = "xl"
    Size2XL  SizeVariant = "2xl"
    Size3XL  SizeVariant = "3xl"
)
```
<!-- LLM-CORE-TYPES-END -->

### Advanced Types

<!-- LLM-ADVANCED-TYPES-START -->
**COMPLEX TYPE DEFINITIONS:**
```go
// Table configuration
type TableConfig struct {
    Columns     []TableColumn     `json:"columns"`
    Sortable    bool              `json:"sortable"`
    Searchable  bool              `json:"searchable"`
    Pagination  PaginationConfig  `json:"pagination"`
    Selection   SelectionConfig   `json:"selection"`
    Actions     []TableAction     `json:"actions"`
}

type TableColumn struct {
    Key        string `json:"key"`
    Label      string `json:"label"`
    Sortable   bool   `json:"sortable"`
    Searchable bool   `json:"searchable"`
    Width      string `json:"width"`
    Align      string `json:"align"`
    Render     string `json:"render"` // Custom render function
}

type PaginationConfig struct {
    Enabled   bool `json:"enabled"`
    PageSize  int  `json:"page_size"`
    ShowTotal bool `json:"show_total"`
    ShowSizer bool `json:"show_sizer"`
}

type SelectionConfig struct {
    Enabled  bool   `json:"enabled"`
    Multiple bool   `json:"multiple"`
    OnSelect string `json:"on_select"`
}

type TableAction struct {
    Label   string `json:"label"`
    Icon    string `json:"icon"`
    Action  string `json:"action"`
    Confirm bool   `json:"confirm"`
    Bulk    bool   `json:"bulk"`
}

// Navigation types
type NavigationConfig struct {
    Items       []NavigationItem `json:"items"`
    Collapsible bool             `json:"collapsible"`
    Theme       string           `json:"theme"`
    Position    string           `json:"position"`
}

type NavigationItem struct {
    Label    string           `json:"label"`
    URL      string           `json:"url"`
    Icon     string           `json:"icon"`
    Active   bool             `json:"active"`
    Children []NavigationItem `json:"children,omitempty"`
    Roles    []string         `json:"roles,omitempty"`
}

// Form schema types
type FormSchema struct {
    Fields  []FieldSchema   `json:"fields"`
    Layout  LayoutSchema    `json:"layout"`
    Actions []ActionSchema  `json:"actions"`
    Rules   []ValidationRule `json:"rules"`
}

type FieldSchema struct {
    Name        string                 `json:"name"`
    Type        string                 `json:"type"`
    Label       string                 `json:"label"`
    Required    bool                   `json:"required"`
    Placeholder string                 `json:"placeholder"`
    Options     []OptionSchema         `json:"options,omitempty"`
    Props       map[string]interface{} `json:"props,omitempty"`
    Conditions  []ConditionSchema      `json:"conditions,omitempty"`
}

type LayoutSchema struct {
    Type    string                 `json:"type"` // "grid", "flex", "stack"
    Columns int                    `json:"columns,omitempty"`
    Gap     string                 `json:"gap,omitempty"`
    Props   map[string]interface{} `json:"props,omitempty"`
}

type ActionSchema struct {
    Type    string `json:"type"`    // "submit", "reset", "button"
    Label   string `json:"label"`
    Variant string `json:"variant"`
    Action  string `json:"action"`
}
```
<!-- LLM-ADVANCED-TYPES-END -->

---

## Error Handling

<!-- LLM-ERROR-HANDLING-START -->
**ERROR TYPES:**
```go
// Custom error types
type UIError struct {
    Type    string `json:"type"`
    Message string `json:"message"`
    Field   string `json:"field,omitempty"`
    Code    string `json:"code"`
}

type ValidationErrors []UIError

func (ve ValidationErrors) HasField(field string) bool
func (ve ValidationErrors) GetFieldError(field string) string
func (ve ValidationErrors) GetFieldErrors(field string) []string

// Error constants
const (
    ErrorTypeValidation = "validation"
    ErrorTypeServer     = "server"
    ErrorTypeNetwork    = "network"
    ErrorTypeAuth       = "auth"
    ErrorTypeNotFound   = "not_found"
)

// Error handling functions
func HandleValidationError(w http.ResponseWriter, errors ValidationErrors)
func HandleServerError(w http.ResponseWriter, err error)
func HandleNotFoundError(w http.ResponseWriter, resource string)
func HandleAuthError(w http.ResponseWriter, message string)
```

**ERROR RESPONSE FORMAT:**
```go
type ErrorResponse struct {
    Success bool     `json:"success"`
    Error   UIError  `json:"error"`
    Errors  []UIError `json:"errors,omitempty"`
}

type SuccessResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Message string      `json:"message,omitempty"`
}
```
<!-- LLM-ERROR-HANDLING-END -->

---

## References

<!-- LLM-API-REFERENCES-START -->
**RELATED DOCUMENTATION:**
- `components/elements.md` - Component usage examples and patterns
- `patterns/htmx-integration.md` - Server interaction patterns and HTMX implementation
- `guides/validation-guide.md` - Complete validation system implementation
- `fundamentals/architecture.md` - System architecture and design principles

**EXTERNAL REFERENCES:**
- `templ-llms.md` - Advanced Templ features and optimization patterns
- `flowbite-llms-full.txt` - Complete Flowbite component reference and styling guide

**OFFICIAL DOCUMENTATION:**
- [Templ Documentation](https://templ.guide) - Official Templ language reference
- [HTMX Reference](https://htmx.org/reference/) - Complete HTMX attribute and event reference
- [Alpine.js API](https://alpinejs.dev/reference/) - Alpine.js directives and methods
- [Flowbite API](https://flowbite.com/docs/components/) - Flowbite component library reference
<!-- LLM-API-REFERENCES-END -->

<!-- LLM-METADATA-START -->
**METADATA FOR AI ASSISTANTS:**
- File Type: Complete API Reference Documentation
- Scope: All component APIs, handler patterns, and type definitions
- Dependencies: Go + Templ + HTMX + Alpine.js + Flowbite
- Target: Developers implementing UI components and server integration
- Complexity: Intermediate to Advanced
- Usage: Function signatures, type definitions, configuration options
<!-- LLM-METADATA-END -->