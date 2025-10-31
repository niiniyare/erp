## Section 7: Field System

### 🎯 Learning Objectives

By the end of this section, you will understand:
- Complete Field struct definition
- All 40+ field types
- Field properties and configuration
- Field validation
- Field relationships and dependencies

### 7.1 Complete Field Struct Definition

```go
// Field represents a form input or UI component
type Field struct {
    // ═══════════════════════════════════════════════════════
    // IDENTITY - Required fields
    // ═══════════════════════════════════════════════════════
    
    // Name: Field identifier (becomes HTML name attribute)
    // Must be unique within schema
    // Format: snake_case recommended
    // Examples: "email", "first_name", "birth_date"
    Name string `json:"name" validate:"required,min=1,max=100"`
    
    // Type: Kind of input field
    // Values: "text", "email", "select", etc. (40+ types)
    Type FieldType `json:"type" validate:"required"`
    
    // Label: Human-readable label shown to users
    // Examples: "Email Address", "First Name"
    Label string `json:"label" validate:"required,min=1,max=200"`
    
    // ═══════════════════════════════════════════════════════
    // DESCRIPTION - Help text and guidance
    // ═══════════════════════════════════════════════════════
    
    // Description: Longer description of field purpose
    Description string `json:"description,omitempty" validate:"max=500"`
    
    // Placeholder: Placeholder text shown when empty
    Placeholder string `json:"placeholder,omitempty" validate:"max=200"`
    
    // Help: Help text below field
    Help string `json:"help,omitempty" validate:"max=500"`
    
    // Tooltip: Tooltip on hover
    Tooltip string `json:"tooltip,omitempty" validate:"max=200"`
    
    // Icon: Icon to display (icon name)
    Icon string `json:"icon,omitempty" validate:"icon_name"`
    
    // Error: Custom error message override
    Error string `json:"error,omitempty" validate:"max=200"`
    
    // ═══════════════════════════════════════════════════════
    // STATE FLAGS - Field behavior
    // ═══════════════════════════════════════════════════════
    
    // Required: Field must have value
    // Becomes HTML "required" attribute
    Required bool `json:"required,omitempty"`
    
    // Disabled: Field cannot be edited
    // Becomes HTML "disabled" attribute
    // Disabled fields are NOT submitted
    Disabled bool `json:"disabled,omitempty"`
    
    // Readonly: Value visible but not editable
    // Becomes HTML "readonly" attribute
    // Readonly fields ARE submitted
    Readonly bool `json:"readonly,omitempty"`
    
    // Hidden: Field not displayed
    // Still exists in DOM, still submitted
    // Different from conditional visibility
    Hidden bool `json:"hidden,omitempty"`
    
    // ═══════════════════════════════════════════════════════
    // VALUES - Field data
    // ═══════════════════════════════════════════════════════
    
    // Value: Current value (for edit forms)
    // Type depends on field type
    Value any `json:"value,omitempty"`
    
    // Default: Default value (for new forms)
    // Used when Value is not provided
    Default any `json:"default,omitempty"`
    
    // Options: For select/radio/checkbox fields
    // Array of Option structs
    Options []Option `json:"options,omitempty" validate:"dive"`
    
    // ═══════════════════════════════════════════════════════
    // BEHAVIOR - Validation and transformation
    // ═══════════════════════════════════════════════════════
    
    // Validation: Validation rules
    Validation *FieldValidation `json:"validation,omitempty"`
    
    // Transform: Value transformation
    Transform *Transform `json:"transform,omitempty"`
    
    // Mask: Input masking
    Mask *Mask `json:"mask,omitempty"`
    
    // Layout: Positioning in grid
    Layout *FieldLayout `json:"layout,omitempty"`
    
    // Style: Custom styling
    Style *Style `json:"style,omitempty"`
    
    // Config: Field-specific configuration
    Config map[string]any `json:"config,omitempty"`
    
    // Events: Event handlers
    Events *FieldEvents `json:"events,omitempty"`
    
    // Conditional: Show/hide conditions
    Conditional *Conditional `json:"conditional,omitempty"`
    
    // DataSource: Dynamic options source
    DataSource *DataSource `json:"dataSource,omitempty"`
    
    // Permissions: Access control
    Permissions *FieldPermissions `json:"permissions,omitempty"`
    
    // Dependencies: Depends on these fields
    Dependencies []string `json:"dependencies,omitempty" validate:"dive,fieldname"`
    
    // ═══════════════════════════════════════════════════════
    // FRAMEWORK INTEGRATION
    // ═══════════════════════════════════════════════════════
    
    // HTMX: HTMX attributes for this field
    HTMX *FieldHTMX `json:"htmx,omitempty"`
    
    // Alpine: Alpine.js bindings for this field
    Alpine *FieldAlpine `json:"alpine,omitempty"`
    
    // ═══════════════════════════════════════════════════════
    // ENHANCED: New fields from our discussion
    // ═══════════════════════════════════════════════════════
    
    // ShowIf: Client-side visibility (Alpine.js expression)
    // Example: "needs_shipping === true"
    // For simple UI logic only
    ShowIf string `json:"showIf,omitempty" validate:"js_expression"`
    
    // RequirePermission: Server-side permission check
    // Example: "hr.view_salary"
    // Backend decides, frontend respects
    RequirePermission string `json:"requirePermission,omitempty"`
    
    // Runtime: Runtime flags (populated by enricher)
    // Contains: Visible, Editable, Reason
    // Set by server, respected by client
    Runtime *FieldRuntime `json:"runtime,omitempty"`
    
    // ═══════════════════════════════════════════════════════
    // INTERNAL - Not in JSON
    // ═══════════════════════════════════════════════════════
    
    // compiledCondition: Pre-compiled condition for performance
    compiledCondition any `json:"-"`
}
```

### 7.2 Field Types Reference (40+ Types)

#### Basic Text Inputs

**FieldText:**
```go
FieldText FieldType = "text"
```

Single-line text input. Most common field type.

**Usage:**
```json
{
  "name": "full_name",
  "type": "text",
  "label": "Full Name",
  "placeholder": "John Doe",
  "validation": {
    "minLength": 2,
    "maxLength": 100
  }
}
```

**Renders as:**
```html
<input type="text" name="full_name" minlength="2" maxlength="100" />
```

**FieldEmail:**
```go
FieldEmail FieldType = "email"
```

Email input with built-in validation.

**Usage:**
```json
{
  "name": "email",
  "type": "email",
  "label": "Email Address",
  "required": true,
  "validation": {
    "maxLength": 255,
    "server": {
      "unique": true
    }
  }
}
```

**Renders as:**
```html
<input type="email" name="email" required maxlength="255" />
```

**FieldPassword:**
```go
FieldPassword FieldType = "password"
```

Password input (masked).

**Usage:**
```json
{
  "name": "password",
  "type": "password",
  "label": "Password",
  "required": true,
  "validation": {
    "minLength": 8,
    "pattern": "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d).+$",
    "messages": {
      "pattern": "Must contain uppercase, lowercase, and number"
    }
  }
}
```

**FieldNumber:**
```go
FieldNumber FieldType = "number"
```

Numeric input.

**Usage:**
```json
{
  "name": "age",
  "type": "number",
  "label": "Age",
  "validation": {
    "min": 18,
    "max": 120,
    "integer": true
  }
}
```

**Renders as:**
```html
<input type="number" name="age" min="18" max="120" step="1" />
```

**FieldPhone:**
```go
FieldPhone FieldType = "phone"
```

Phone number input with masking.

**Usage:**
```json
{
  "name": "phone",
  "type": "phone",
  "label": "Phone Number",
  "mask": {
    "pattern": "(999) 999-9999",
    "placeholder": "_",
    "showMask": true
  }
}
```

**FieldURL:**
```go
FieldURL FieldType = "url"
```

URL input with validation.

**Usage:**
```json
{
  "name": "website",
  "type": "url",
  "label": "Website",
  "placeholder": "https://example.com"
}
```

**FieldHidden:**
```go
FieldHidden FieldType = "hidden"
```

Hidden input (submitted but not visible).

**Usage:**
```json
{
  "name": "csrf_token",
  "type": "hidden",
  "value": "abc123"
}
```

#### Date and Time Inputs

**FieldDate:**
```go
FieldDate FieldType = "date"
```

Date picker.

**Usage:**
```json
{
  "name": "birth_date",
  "type": "date",
  "label": "Date of Birth",
  "validation": {
    "max": "2006-01-01"
  }
}
```

**FieldTime:**
```go
FieldTime FieldType = "time"
```

Time picker.

**Usage:**
```json
{
  "name": "appointment_time",
  "type": "time",
  "label": "Appointment Time"
}
```

**FieldDateTime:**
```go
FieldDateTime FieldType = "datetime"
```

Date and time combined.

**Usage:**
```json
{
  "name": "event_datetime",
  "type": "datetime",
  "label": "Event Date & Time"
}
```

**FieldDateRange:**
```go
FieldDateRange FieldType = "daterange"
```

Start and end date selection.

**Usage:**
```json
{
  "name": "date_range",
  "type": "daterange",
  "label": "Select Date Range",
  "config": {
    "format": "YYYY-MM-DD"
  }
}
```

#### Text Content Inputs

**FieldTextarea:**
```go
FieldTextarea FieldType = "textarea"
```

Multi-line text input.

**Usage:**
```json
{
  "name": "description",
  "type": "textarea",
  "label": "Description",
  "config": {
    "rows": 5
  },
  "validation": {
    "maxLength": 1000
  }
}
```

**Renders as:**
```html
<textarea name="description" rows="5" maxlength="1000"></textarea>
```

**FieldRichText:**
```go
FieldRichText FieldType = "richtext"
```

WYSIWYG editor.

**Usage:**
```json
{
  "name": "content",
  "type": "richtext",
  "label": "Content",
  "config": {
    "toolbar": ["bold", "italic", "underline", "link"],
    "height": 400
  }
}
```

**FieldCode:**
```go
FieldCode FieldType = "code"
```

Code editor with syntax highlighting.

**Usage:**
```json
{
  "name": "code_snippet",
  "type": "code",
  "label": "Code",
  "config": {
    "language": "javascript",
    "theme": "monokai",
    "lineNumbers": true
  }
}
```

**FieldJSON:**
```go
FieldJSON FieldType = "json"
```

JSON editor with validation.

**Usage:**
```json
{
  "name": "metadata",
  "type": "json",
  "label": "Metadata",
  "config": {
    "validate": true,
    "format": true
  }
}
```

#### Selection Inputs

**FieldSelect:**
```go
FieldSelect FieldType = "select"
```

Dropdown list (single selection).

**Usage:**
```json
{
  "name": "country",
  "type": "select",
  "label": "Country",
  "required": true,
  "options": [
    {"value": "us", "label": "United States"},
    {"value": "ca", "label": "Canada"},
    {"value": "uk", "label": "United Kingdom"}
  ]
}
```

**Renders as:**
```html
<select name="country" required>
  <option value="">-- Select --</option>
  <option value="us">United States</option>
  <option value="ca">Canada</option>
  <option value="uk">United Kingdom</option>
</select>
```

**FieldMultiSelect:**
```go
FieldMultiSelect FieldType = "multiselect"
```

Multiple selection dropdown.

**Usage:**
```json
{
  "name": "skills",
  "type": "multiselect",
  "label": "Skills",
  "options": [
    {"value": "go", "label": "Go"},
    {"value": "js", "label": "JavaScript"},
    {"value": "python", "label": "Python"}
  ]
}
```

**FieldRadio:**
```go
FieldRadio FieldType = "radio"
```

Radio buttons (single selection).

**Usage:**
```json
{
  "name": "gender",
  "type": "radio",
  "label": "Gender",
  "options": [
    {"value": "male", "label": "Male"},
    {"value": "female", "label": "Female"},
    {"value": "other", "label": "Other"}
  ]
}
```

**Renders as:**
```html
<div>
  <label><input type="radio" name="gender" value="male" /> Male</label>
  <label><input type="radio" name="gender" value="female" /> Female</label>
  <label><input type="radio" name="gender" value="other" /> Other</label>
</div>
```

**FieldCheckbox:**
```go
FieldCheckbox FieldType = "checkbox"
```

Checkboxes (multiple selection or single boolean).

**Usage (single boolean):**
```json
{
  "name": "agree_terms",
  "type": "checkbox",
  "label": "I agree to the terms and conditions",
  "required": true
}
```

**Usage (multiple selection):**
```json
{
  "name": "interests",
  "type": "checkbox",
  "label": "Interests",
  "options": [
    {"value": "sports", "label": "Sports"},
    {"value": "music", "label": "Music"},
    {"value": "reading", "label": "Reading"}
  ]
}
```

**FieldTreeSelect:**
```go
FieldTreeSelect FieldType = "treeselect"
```

Hierarchical selection.

**Usage:**
```json
{
  "name": "category",
  "type": "treeselect",
  "label": "Category",
  "options": [
    {
      "value": "electronics",
      "label": "Electronics",
      "children": [
        {"value": "phones", "label": "Phones"},
        {"value": "laptops", "label": "Laptops"}
      ]
    },
    {
      "value": "clothing",
      "label": "Clothing",
      "children": [
        {"value": "mens", "label": "Men's"},
        {"value": "womens", "label": "Women's"}
      ]
    }
  ]
}
```

**FieldCascader:**
```go
FieldCascader FieldType = "cascader"
```

Cascading dropdowns.

**Usage:**
```json
{
  "name": "location",
  "type": "cascader",
  "label": "Location",
  "options": [
    {
      "value": "us",
      "label": "United States",
      "children": [
        {
          "value": "ca",
          "label": "California",
          "children": [
            {"value": "sf", "label": "San Francisco"},
            {"value": "la", "label": "Los Angeles"}
          ]
        }
      ]
    }
  ]
}
```

**FieldTransfer:**
```go
FieldTransfer FieldType = "transfer"
```

Transfer list (left/right selection).

**Usage:**
```json
{
  "name": "assigned_users",
  "type": "transfer",
  "label": "Assign Users",
  "dataSource": {
    "type": "api",
    "url": "/api/users",
    "method": "GET"
  }
}
```

#### Interactive Controls

**FieldSwitch:**
```go
FieldSwitch FieldType = "switch"
```

Toggle switch (boolean).

**Usage:**
```json
{
  "name": "is_active",
  "type": "switch",
  "label": "Active",
  "default": true
}
```

**FieldSlider:**
```go
FieldSlider FieldType = "slider"
```

Range slider.

**Usage:**
```json
{
  "name": "volume",
  "type": "slider",
  "label": "Volume",
  "validation": {
    "min": 0,
    "max": 100,
    "step": 5
  },
  "default": 50
}
```

**FieldRating:**
```go
FieldRating FieldType = "rating"
```

Star rating.

**Usage:**
```json
{
  "name": "rating",
  "type": "rating",
  "label": "Rating",
  "config": {
    "max": 5,
    "allowHalf": true
  }
}
```

**FieldColor:**
```go
FieldColor FieldType = "color"
```

Color picker.

**Usage:**
```json
{
  "name": "theme_color",
  "type": "color",
  "label": "Theme Color",
  "default": "#0066cc"
}
```

#### File Uploads

**FieldFile:**
```go
FieldFile FieldType = "file"
```

Generic file upload.

**Usage:**
```json
{
  "name": "document",
  "type": "file",
  "label": "Upload Document",
  "config": {
    "accept": ".pdf,.doc,.docx",
    "maxSize": 5242880,
    "multiple": false
  }
}
```

**FieldImage:**
```go
FieldImage FieldType = "image"
```

Image upload with preview.

**Usage:**
```json
{
  "name": "avatar",
  "type": "image",
  "label": "Profile Picture",
  "config": {
    "accept": ".jpg,.jpeg,.png",
    "maxSize": 2097152,
    "preview": true,
    "crop": true,
    "aspectRatio": 1
  }
}
```

**FieldSignature:**
```go
FieldSignature FieldType = "signature"
```

Signature pad.

**Usage:**
```json
{
  "name": "signature",
  "type": "signature",
  "label": "Signature",
  "required": true,
  "config": {
    "width": 400,
    "height": 200,
    "penColor": "#000000"
  }
}
```

#### Specialized Inputs

**FieldCurrency:**
```go
FieldCurrency FieldType = "currency"
```

Money input with formatting.

**Usage:**
```json
{
  "name": "amount",
  "type": "currency",
  "label": "Amount",
  "config": {
    "currency": "USD",
    "locale": "en-US"
  },
  "validation": {
    "min": 0,
    "positive": true
  }
}
```

**FieldTags:**
```go
FieldTags FieldType = "tags"
```

Tag input (comma-separated).

**Usage:**
```json
{
  "name": "tags",
  "type": "tags",
  "label": "Tags",
  "placeholder": "Add tags...",
  "config": {
    "separator": ",",
    "allowDuplicates": false
  }
}
```

**FieldLocation:**
```go
FieldLocation FieldType = "location"
```

Address/location picker.

**Usage:**
```json
{
  "name": "location",
  "type": "location",
  "label": "Location",
  "config": {
    "enableMap": true,
    "enableAutocomplete": true
  }
}
```

**FieldRelation:**
```go
FieldRelation FieldType = "relation"
```

Foreign key relationship.

**Usage:**
```json
{
  "name": "customer_id",
  "type": "relation",
  "label": "Customer",
  "dataSource": {
    "type": "api",
    "url": "/api/customers",
    "method": "GET"
  },
  "config": {
    "valueField": "id",
    "labelField": "name",
    "searchable": true
  }
}
```

**FieldAutoComplete:**
```go
FieldAutoComplete FieldType = "autocomplete"
```

Autocomplete search.

**Usage:**
```json
{
  "name": "city",
  "type": "autocomplete",
  "label": "City",
  "dataSource": {
    "type": "api",
    "url": "/api/cities/search",
    "method": "GET"
  },
  "config": {
    "minChars": 2,
    "debounce": 300
  }
}
```

#### Display Only

**FieldDisplay:**
```go
FieldDisplay FieldType = "display"
```

Read-only display (not editable).

**Usage:**
```json
{
  "name": "created_at",
  "type": "display",
  "label": "Created",
  "value": "2024-01-15 10:30:00"
}
```

**FieldDivider:**
```go
FieldDivider FieldType = "divider"
```

Visual separator.

**Usage:**
```json
{
  "name": "divider_1",
  "type": "divider",
  "label": "Contact Information"
}
```

**FieldHTML:**
```go
FieldHTML FieldType = "html"
```

Raw HTML content.

**Usage:**
```json
{
  "name": "disclaimer",
  "type": "html",
  "value": "<p>By submitting this form, you agree to our <a href='/terms'>Terms of Service</a>.</p>"
}
```

#### Collections

**FieldRepeatable:**
```go
FieldRepeatable FieldType = "repeatable"
```

Repeatable field groups (dynamic arrays).

**Usage:**
```json
{
  "name": "addresses",
  "type": "repeatable",
  "label": "Addresses",
  "config": {
    "minItems": 1,
    "maxItems": 5,
    "itemLabel": "Address {index}",
    "addText": "Add Address",
    "removeText": "Remove",
    "template": [
      {
        "name": "street",
        "type": "text",
        "label": "Street",
        "required": true
      },
      {
        "name": "city",
        "type": "text",
        "label": "City",
        "required": true
      },
      {
        "name": "zip",
        "type": "text",
        "label": "ZIP Code",
        "required": true
      }
    ]
  }
}
```

**FieldTableRepeater:**
```go
FieldTableRepeater FieldType = "table_repeater"
```

Table-style repeatable fields.

**Usage:**
```json
{
  "name": "line_items",
  "type": "table_repeater",
  "label": "Line Items",
  "config": {
    "columns": [
      {"name": "product", "label": "Product", "type": "text"},
      {"name": "quantity", "label": "Qty", "type": "number"},
      {"name": "price", "label": "Price", "type": "currency"},
      {"name": "total", "label": "Total", "type": "currency", "readonly": true}
    ],
    "minItems": 1,
    "maxItems": 50
  }
}
```

### 7.3 Field Properties

#### Required Properties

Every field must have:
- `name` - Unique identifier
- `type` - Field type from FieldType enum
- `label` - Human-readable label

#### Optional Properties

All other properties are optional but commonly used:

**Description & Help:**
- `description` - Longer explanation
- `placeholder` - Hint text when empty
- `help` - Help text below field
- `tooltip` - Hover tooltip
- `icon` - Icon to display

**State Flags:**
- `required` - Must have value
- `disabled` - Cannot edit
- `readonly` - Can see but not edit
- `hidden` - Not displayed

**Values:**
- `value` - Current value (edit mode)
- `default` - Default value (create mode)
- `options` - For select/radio/checkbox

**Configuration:**
- `validation` - Validation rules
- `transform` - Value transformation
- `mask` - Input masking
- `layout` - Grid positioning
- `style` - Custom CSS
- `config` - Field-specific settings
- `events` - Event handlers
- `conditional` - Show/hide logic
- `dataSource` - Dynamic options
- `permissions` - Access control
- `dependencies` - Field dependencies

**Framework:**
- `htmx` - HTMX attributes
- `alpine` - Alpine.js bindings

**Enhanced (from our discussion):**
- `showIf` - Client visibility expression
- `requirePermission` - Server permission check
- `runtime` - Runtime flags (enriched)

### 7.4 Field Validation

See Section 8 for complete validation documentation.

Quick example:
```json
{
  "name": "email",
  "type": "email",
  "label": "Email",
  "required": true,
  "validation": {
    "maxLength": 255,
    "messages": {
      "required": "Email is required",
      "pattern": "Please enter a valid email"
    },
    "server": {
      "unique": true
    }
  }
}
```

### 7.5 Field Relationships

#### Dependencies

Fields can depend on other fields:

```json
{
  "name": "state",
  "type": "select",
  "label": "State",
  "dependencies": ["country"],
  "dataSource": {
    "type": "api",
    "url": "/api/states?country={country}",
    "method": "GET"
  }
}
```

When `country` changes, `state` options reload.

#### Conditional Visibility

**Client-side (simple):**
```json
{
  "name": "shipping_address",
  "type": "textarea",
  "label": "Shipping Address",
  "showIf": "needs_shipping === true"
}
```

**Server-side (permissions):**
```json
{
  "name": "employee_salary",
  "type": "currency",
  "label": "Salary",
  "requirePermission": "hr.view_salary"
}
```

**Complex (condition package):**
```json
{
  "name": "discount_code",
  "type": "text",
  "label": "Discount Code",
  "conditional": {
    "show": {
      "logic": "AND",
      "conditions": [
        {"field": "order_total", "operator": "greater", "value": 100},
        {"field": "is_premium", "operator": "equals", "value": true}
      ]
    }
  }
}
```

### 7.6 Field Methods

Your Field struct has these methods:

```go
func (f *Field) IsVisible(data map[string]any) bool
```

Checks if field should be displayed given current form data.

```go
func (f *Field) IsRequired(data map[string]any) bool
```

Checks if field is required given current form data.

```go
func (f *Field) Validate(ctx context.Context) error
```

Validates field configuration structure.

```go
func (f *Field) ValidateValue(value any) error
```

Validates a value against field rules.

```go
func (f *Field) GetDefaultValue() any
```

Returns appropriate default value for field type.

---

