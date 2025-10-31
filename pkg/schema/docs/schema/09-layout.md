## Section 9: Layout System

### 🎯 Learning Objectives

By the end of this section, you will understand:
- Complete Layout struct definition
- All layout types (grid, flex, tabs, steps, sections)
- Responsive design strategies
- Layout composition patterns
- How to position fields effectively

### 9.1 Complete Layout Struct Definition

Here is the actual Layout struct from your codebase:

```go
// Layout defines the visual structure and arrangement of form elements
type Layout struct {
    // ═══════════════════════════════════════════════════════
    // TYPE - Determines layout algorithm
    // ═══════════════════════════════════════════════════════
    
    // Type: Layout algorithm to use
    // Values: "grid", "flex", "tabs", "steps", "sections", "groups"
    Type LayoutType `json:"type" validate:"required"`
    
    // ═══════════════════════════════════════════════════════
    // GRID/FLEX CONFIGURATION - For simple layouts
    // ═══════════════════════════════════════════════════════
    
    // Columns: Number of columns (1-24)
    // Default: 2 for grid, 1 for flex
    Columns int `json:"columns,omitempty" validate:"min=1,max=24"`
    
    // Gap: Space between items
    // CSS size value: "1rem", "16px", "2em"
    Gap string `json:"gap,omitempty" validate:"css_size"`
    
    // Direction: Flex direction
    // Values: "row", "column", "row-reverse", "column-reverse"
    Direction string `json:"direction,omitempty" validate:"oneof=row column row-reverse column-reverse"`
    
    // Wrap: Allow items to wrap to next line
    Wrap bool `json:"wrap,omitempty"`
    
    // Responsive: Enable responsive behavior
    // When true, uses breakpoints
    Responsive bool `json:"responsive,omitempty"`
    
    // ═══════════════════════════════════════════════════════
    // COMPLEX LAYOUTS - For advanced organization
    // ═══════════════════════════════════════════════════════
    
    // Sections: Logical sections with titles
    Sections []Section `json:"sections,omitempty" validate:"dive"`
    
    // Groups: Visual field groups (like fieldset)
    Groups []Group `json:"groups,omitempty" validate:"dive"`
    
    // Tabs: Tabbed interface
    Tabs []Tab `json:"tabs,omitempty" validate:"dive"`
    
    // Steps: Multi-step wizard
    Steps []Step `json:"steps,omitempty" validate:"dive"`
    
    // ═══════════════════════════════════════════════════════
    // RESPONSIVE BREAKPOINTS - Different layouts per screen size
    // ═══════════════════════════════════════════════════════
    
    // Breakpoints: Screen size configurations
    Breakpoints *Breakpoints `json:"breakpoints,omitempty"`
}
```

### 9.2 Layout Types

```go
type LayoutType string

const (
    LayoutGrid     LayoutType = "grid"     // CSS Grid layout
    LayoutFlex     LayoutType = "flex"     // Flexbox layout
    LayoutTabs     LayoutType = "tabs"     // Tabbed interface
    LayoutSteps    LayoutType = "steps"    // Multi-step wizard
    LayoutSections LayoutType = "sections" // Divided into sections
    LayoutGroups   LayoutType = "groups"   // Field grouping
)
```

#### Grid Layout

**Best for:** Forms with multiple columns, structured data entry.

**Example:**
```json
{
  "layout": {
    "type": "grid",
    "columns": 2,
    "gap": "1rem",
    "responsive": true
  }
}
```

**Renders as:**
```html
<div style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 1rem;">
  <!-- Fields auto-flow into grid -->
  <div>Field 1</div>
  <div>Field 2</div>
  <div>Field 3</div>
  <div>Field 4</div>
</div>
```

**Visual:**
```
┌──────────────┬──────────────┐
│   Field 1    │   Field 2    │
├──────────────┼──────────────┤
│   Field 3    │   Field 4    │
├──────────────┼──────────────┤
│   Field 5    │   Field 6    │
└──────────────┴──────────────┘
```

**Field-Level Control:**

Fields can span multiple columns:

```json
{
  "name": "description",
  "type": "textarea",
  "layout": {
    "colSpan": 2  // Spans both columns
  }
}
```

**Visual:**
```
┌──────────────┬──────────────┐
│   Name       │   Email      │
├──────────────┴──────────────┤
│   Description (full width)  │
└─────────────────────────────┘
```

#### Flex Layout

**Best for:** Single column forms, mobile-first designs.

**Example:**
```json
{
  "layout": {
    "type": "flex",
    "direction": "column",
    "gap": "1rem"
  }
}
```

**Renders as:**
```html
<div style="display: flex; flex-direction: column; gap: 1rem;">
  <div>Field 1</div>
  <div>Field 2</div>
  <div>Field 3</div>
</div>
```

**Visual:**
```
┌─────────────────────────────┐
│        Field 1              │
├─────────────────────────────┤
│        Field 2              │
├─────────────────────────────┤
│        Field 3              │
└─────────────────────────────┘
```

#### Tabs Layout

**Best for:** Organizing many fields into logical categories.

**Example:**
```json
{
  "layout": {
    "type": "tabs",
    "tabs": [
      {
        "id": "personal",
        "label": "Personal Info",
        "icon": "user",
        "fields": ["first_name", "last_name", "email", "phone"]
      },
      {
        "id": "address",
        "label": "Address",
        "icon": "map-pin",
        "fields": ["street", "city", "state", "zip"]
      },
      {
        "id": "preferences",
        "label": "Preferences",
        "icon": "settings",
        "fields": ["language", "timezone", "notifications"]
      }
    ]
  }
}
```

**Visual:**
```
┌─────────────┬─────────────┬─────────────┐
│ Personal ✓  │  Address    │ Preferences │  ← Tab headers
├─────────────┴─────────────┴─────────────┤
│                                          │
│  First Name: [_______]                   │
│  Last Name:  [_______]                   │
│  Email:      [_______]                   │
│  Phone:      [_______]                   │
│                                          │
└──────────────────────────────────────────┘
```

**Tab struct:**
```go
type Tab struct {
    ID          string       `json:"id" validate:"required"`
    Label       string       `json:"label" validate:"required"`
    Icon        string       `json:"icon,omitempty" validate:"icon_name"`
    Description string       `json:"description,omitempty"`
    Fields      []string     `json:"fields" validate:"required,dive,fieldname"`
    Badge       string       `json:"badge,omitempty"`       // Badge text (e.g., "3")
    Disabled    bool         `json:"disabled,omitempty"`    // Tab disabled
    Order       int          `json:"order,omitempty"`       // Tab order
    Conditional *Conditional `json:"conditional,omitempty"` // Show/hide tab
    Style       *Style       `json:"style,omitempty"`
}
```

#### Steps Layout

**Best for:** Multi-step wizards, onboarding flows, checkout processes.

**Example:**
```json
{
  "layout": {
    "type": "steps",
    "steps": [
      {
        "id": "account",
        "title": "Account Details",
        "description": "Create your account",
        "icon": "user-plus",
        "fields": ["username", "email", "password"],
        "order": 0,
        "validation": true
      },
      {
        "id": "profile",
        "title": "Profile Information",
        "description": "Tell us about yourself",
        "icon": "id-card",
        "fields": ["first_name", "last_name", "bio"],
        "order": 1,
        "validation": true
      },
      {
        "id": "preferences",
        "title": "Preferences",
        "description": "Customize your experience",
        "icon": "settings",
        "fields": ["language", "timezone", "notifications"],
        "order": 2,
        "skippable": true
      }
    ]
  }
}
```

**Visual:**
```
   1. Account    →    2. Profile    →    3. Preferences
   ━━━━━━━━━━         ─────────          ─────────
   (completed)        (current)          (pending)

┌────────────────────────────────────────────────┐
│  Step 2 of 3: Profile Information              │
│  Tell us about yourself                        │
├────────────────────────────────────────────────┤
│                                                │
│  First Name: [_______]                         │
│  Last Name:  [_______]                         │
│  Bio:        [________________]                │
│                                                │
│  [← Back]              [Next →]                │
└────────────────────────────────────────────────┘
```

**Step struct:**
```go
type Step struct {
    ID          string       `json:"id" validate:"required"`
    Title       string       `json:"title" validate:"required"`
    Description string       `json:"description,omitempty"`
    Icon        string       `json:"icon,omitempty" validate:"icon_name"`
    Fields      []string     `json:"fields" validate:"required,dive,fieldname"`
    Order       int          `json:"order" validate:"min=0"`
    Skippable   bool         `json:"skippable,omitempty"`   // Can skip this step
    Validation  bool         `json:"validation,omitempty"`  // Validate before next
    Conditional *Conditional `json:"conditional,omitempty"` // Show/hide step
    Style       *Style       `json:"style,omitempty"`
}
```

**Navigation:**
- Steps must be completed in order (unless `skippable: true`)
- If `validation: true`, validates current step before allowing "Next"
- Progress indicator shows completion status
- Back button allows returning to previous steps

#### Sections Layout

**Best for:** Long forms with logical groupings, collapsible sections.

**Example:**
```json
{
  "layout": {
    "type": "sections",
    "sections": [
      {
        "id": "personal",
        "title": "Personal Information",
        "description": "Basic personal details",
        "icon": "user",
        "fields": ["first_name", "last_name", "birth_date"],
        "collapsible": true,
        "collapsed": false,
        "columns": 2
      },
      {
        "id": "contact",
        "title": "Contact Information",
        "description": "How we can reach you",
        "icon": "mail",
        "fields": ["email", "phone", "address"],
        "collapsible": true,
        "collapsed": false,
        "columns": 1
      }
    ]
  }
}
```

**Visual:**
```
┌────────────────────────────────────────┐
│ 👤 Personal Information            [▼] │  ← Collapsible header
│    Basic personal details              │
├────────────────────────────────────────┤
│  First Name: [______]  Last Name: [____]│
│  Birth Date: [__/__/__]                │
└────────────────────────────────────────┘

┌────────────────────────────────────────┐
│ ✉️  Contact Information            [▼] │
│    How we can reach you                │
├────────────────────────────────────────┤
│  Email:   [_______________________]    │
│  Phone:   [_______________________]    │
│  Address: [_______________________]    │
└────────────────────────────────────────┘
```

**Section struct:**
```go
type Section struct {
    ID          string       `json:"id" validate:"required"`
    Title       string       `json:"title,omitempty"`
    Description string       `json:"description,omitempty"`
    Icon        string       `json:"icon,omitempty" validate:"icon_name"`
    Fields      []string     `json:"fields" validate:"required,dive,fieldname"`
    Collapsible bool         `json:"collapsible,omitempty"` // Can collapse
    Collapsed   bool         `json:"collapsed,omitempty"`   // Initially collapsed
    Columns     int          `json:"columns,omitempty" validate:"min=1,max=12"`
    Order       int          `json:"order,omitempty"`
    Conditional *Conditional `json:"conditional,omitempty"`
    Style       *Style       `json:"style,omitempty"`
}
```

#### Groups Layout

**Best for:** Visual grouping like HTML fieldset, related fields.

**Example:**
```json
{
  "layout": {
    "type": "groups",
    "groups": [
      {
        "id": "name_group",
        "label": "Full Name",
        "fields": ["first_name", "last_name"],
        "border": true,
        "columns": 2
      },
      {
        "id": "address_group",
        "label": "Mailing Address",
        "description": "Where should we send correspondence?",
        "fields": ["street", "city", "state", "zip"],
        "border": true,
        "columns": 2
      }
    ]
  }
}
```

**Visual:**
```
┌─ Full Name ──────────────────────────┐
│  First Name: [______]  Last Name: [__]│
└──────────────────────────────────────┘

┌─ Mailing Address ────────────────────┐
│  Where should we send correspondence? │
│  Street: [________________________]  │
│  City: [_______]  State: [__]        │
│  ZIP: [_____]                         │
└──────────────────────────────────────┘
```

**Group struct:**
```go
type Group struct {
    ID          string       `json:"id" validate:"required"`
    Label       string       `json:"label,omitempty"`
    Description string       `json:"description,omitempty"`
    Fields      []string     `json:"fields" validate:"required,dive,fieldname"`
    Border      bool         `json:"border,omitempty"`    // Show border
    Columns     int          `json:"columns,omitempty" validate:"min=1,max=12"`
    Order       int          `json:"order,omitempty"`
    Conditional *Conditional `json:"conditional,omitempty"`
    Style       *Style       `json:"style,omitempty"`
}
```

### 9.3 Responsive Design

#### Breakpoints Configuration

```go
type Breakpoints struct {
    Mobile  *BreakpointConfig `json:"mobile,omitempty"`  // < 640px
    Tablet  *BreakpointConfig `json:"tablet,omitempty"`  // 640px - 1024px
    Desktop *BreakpointConfig `json:"desktop,omitempty"` // > 1024px
}

type BreakpointConfig struct {
    Columns    int      `json:"columns,omitempty" validate:"min=1,max=24"`
    Gap        string   `json:"gap,omitempty" validate:"css_size"`
    Direction  string   `json:"direction,omitempty" validate:"oneof=row column"`
    HideFields []string `json:"hideFields,omitempty"` // Fields to hide at this size
}
```

**Example:**
```json
{
  "layout": {
    "type": "grid",
    "columns": 3,
    "gap": "1.5rem",
    "responsive": true,
    "breakpoints": {
      "mobile": {
        "columns": 1,
        "gap": "1rem",
        "hideFields": ["optional_field"]
      },
      "tablet": {
        "columns": 2,
        "gap": "1.25rem"
      },
      "desktop": {
        "columns": 3,
        "gap": "1.5rem"
      }
    }
  }
}
```

**Behavior:**
- **Desktop (>1024px)**: 3 columns, 1.5rem gap
- **Tablet (640-1024px)**: 2 columns, 1.25rem gap
- **Mobile (<640px)**: 1 column, 1rem gap, hides "optional_field"

**Rendered CSS:**
```css
/* Desktop (default) */
.form-layout {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1.5rem;
}

/* Tablet */
@media (max-width: 1024px) {
  .form-layout {
    grid-template-columns: repeat(2, 1fr);
    gap: 1.25rem;
  }
}

/* Mobile */
@media (max-width: 640px) {
  .form-layout {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
  .field[data-name="optional_field"] {
    display: none;
  }
}
```

### 9.4 Field-Level Layout Control

Individual fields can control their positioning:

```go
type FieldLayout struct {
    Row     int    `json:"row,omitempty"`     // Grid row
    Column  int    `json:"column,omitempty"`  // Grid column
    ColSpan int    `json:"colSpan,omitempty"` // Columns to span
    RowSpan int    `json:"rowSpan,omitempty"` // Rows to span
    Order   int    `json:"order,omitempty"`   // Display order
    Width   string `json:"width,omitempty"`   // Custom width
    Offset  int    `json:"offset,omitempty"`  // Column offset
    Class   string `json:"class,omitempty"`   // CSS classes
}
```

**Example:**
```json
{
  "fields": [
    {
      "name": "first_name",
      "type": "text",
      "label": "First Name"
      // Default: auto-placement
    },
    {
      "name": "last_name",
      "type": "text",
      "label": "Last Name"
      // Default: auto-placement
    },
    {
      "name": "bio",
      "type": "textarea",
      "label": "Biography",
      "layout": {
        "colSpan": 2,  // Spans both columns
        "order": 10    // Shows later
      }
    },
    {
      "name": "profile_picture",
      "type": "image",
      "label": "Profile Picture",
      "layout": {
        "row": 1,
        "column": 3,
        "rowSpan": 2  // Spans 2 rows
      }
    }
  ]
}
```

**Visual Result (3-column grid):**
```
┌──────────────┬──────────────┬──────────────┐
│ First Name   │ Last Name    │              │
├──────────────┼──────────────┤  Profile     │
│                              │  Picture     │
│  Biography (spans 2 cols)    │  (spans 2    │
│                              │   rows)      │
└──────────────────────────────┴──────────────┘
```

### 9.5 Layout Validation

Your Layout struct has a validation method:

```go
func (l *Layout) ValidateLayout(schema *Schema) error
```

**Checks:**
1. All field names in sections/tabs/steps/groups exist in schema
2. No duplicate field references
3. All fields are assigned to exactly one section/tab/step (if using those layouts)
4. Valid column counts (1-24)
5. Valid direction values
6. Valid gap values (CSS size)

**Example validation errors:**
```go
// Error: Field doesn't exist
"Section.personal references non-existent field: middle_name"

// Error: Field used in multiple tabs
"Field 'email' appears in multiple tabs: personal, contact"

// Error: Invalid columns
"Columns must be between 1 and 24, got: 30"
```

### 9.6 Complete Layout Examples

#### Example 1: Simple Two-Column Form

```json
{
  "id": "contact-form",
  "type": "form",
  "title": "Contact Information",
  "layout": {
    "type": "grid",
    "columns": 2,
    "gap": "1rem",
    "responsive": true
  },
  "fields": [
    {"name": "first_name", "type": "text", "label": "First Name"},
    {"name": "last_name", "type": "text", "label": "Last Name"},
    {"name": "email", "type": "email", "label": "Email"},
    {"name": "phone", "type": "phone", "label": "Phone"},
    {
      "name": "message",
      "type": "textarea",
      "label": "Message",
      "layout": {"colSpan": 2}
    }
  ]
}
```

#### Example 2: Tabbed Form

```json
{
  "id": "employee-form",
  "type": "form",
  "title": "Employee Information",
  "layout": {
    "type": "tabs",
    "tabs": [
      {
        "id": "personal",
        "label": "Personal",
        "icon": "user",
        "fields": ["first_name", "last_name", "birth_date", "ssn"]
      },
      {
        "id": "employment",
        "label": "Employment",
        "icon": "briefcase",
        "fields": ["hire_date", "department", "position", "salary"]
      },
      {
        "id": "contact",
        "label": "Contact",
        "icon": "mail",
        "fields": ["email", "phone", "address", "emergency_contact"]
      }
    ]
  },
  "fields": [
    {"name": "first_name", "type": "text", "label": "First Name"},
    {"name": "last_name", "type": "text", "label": "Last Name"},
    {"name": "birth_date", "type": "date", "label": "Birth Date"},
    {"name": "ssn", "type": "text", "label": "SSN", "requirePermission": "hr.view_ssn"},
    {"name": "hire_date", "type": "date", "label": "Hire Date"},
    {"name": "department", "type": "select", "label": "Department"},
    {"name": "position", "type": "text", "label": "Position"},
    {"name": "salary", "type": "currency", "label": "Salary", "requirePermission": "hr.view_salary"},
    {"name": "email", "type": "email", "label": "Email"},
    {"name": "phone", "type": "phone", "label": "Phone"},
    {"name": "address", "type": "textarea", "label": "Address"},
    {"name": "emergency_contact", "type": "text", "label": "Emergency Contact"}
  ]
}
```

#### Example 3: Multi-Step Wizard

```json
{
  "id": "registration-wizard",
  "type": "workflow",
  "title": "User Registration",
  "layout": {
    "type": "steps",
    "steps": [
      {
        "id": "account",
        "title": "Create Account",
        "icon": "user-plus",
        "fields": ["username", "email", "password", "confirm_password"],
        "order": 0,
        "validation": true
      },
      {
        "id": "profile",
        "title": "Profile Info",
        "icon": "id-card",
        "fields": ["first_name", "last_name", "birth_date", "avatar"],
        "order": 1,
        "validation": true
      },
      {
        "id": "preferences",
        "title": "Preferences",
        "icon": "settings",
        "fields": ["language", "timezone", "newsletter"],
        "order": 2,
        "skippable": true
      },
      {
        "id": "confirm",
        "title": "Confirm",
        "icon": "check-circle",
        "fields": ["terms_accepted"],
        "order": 3,
        "validation": true
      }
    ]
  },
  "fields": [
    {"name": "username", "type": "text", "label": "Username", "required": true},
    {"name": "email", "type": "email", "label": "Email", "required": true},
    {"name": "password", "type": "password", "label": "Password", "required": true},
    {"name": "confirm_password", "type": "password", "label": "Confirm Password", "required": true},
    {"name": "first_name", "type": "text", "label": "First Name", "required": true},
    {"name": "last_name", "type": "text", "label": "Last Name", "required": true},
    {"name": "birth_date", "type": "date", "label": "Birth Date"},
    {"name": "avatar", "type": "image", "label": "Profile Picture"},
    {"name": "language", "type": "select", "label": "Language"},
    {"name": "timezone", "type": "select", "label": "Timezone"},
    {"name": "newsletter", "type": "checkbox", "label": "Subscribe to newsletter"},
    {"name": "terms_accepted", "type": "checkbox", "label": "I accept the terms", "required": true}
  ]
}
```

#### Example 4: Sectioned Form with Responsive Grid

```json
{
  "id": "invoice-form",
  "type": "form",
  "title": "Create Invoice",
  "layout": {
    "type": "sections",
    "columns": 2,
    "gap": "1.5rem",
    "responsive": true,
    "sections": [
      {
        "id": "customer",
        "title": "Customer Information",
        "icon": "user",
        "fields": ["customer_name", "customer_email", "customer_address"],
        "collapsible": true,
        "columns": 2
      },
      {
        "id": "invoice_details",
        "title": "Invoice Details",
        "icon": "file-text",
        "fields": ["invoice_number", "invoice_date", "due_date", "terms"],
        "collapsible": true,
        "columns": 2
      },
      {
        "id": "line_items",
        "title": "Line Items",
        "icon": "list",
        "fields": ["line_items"],
        "collapsible": false,
        "columns": 1
      },
      {
        "id": "totals",
        "title": "Totals",
        "icon": "dollar-sign",
        "fields": ["subtotal", "tax", "total"],
        "collapsible": false,
        "columns": 2
      }
    ],
    "breakpoints": {
      "mobile": {
        "columns": 1,
        "gap": "1rem"
      },
      "tablet": {
        "columns": 2,
        "gap": "1.25rem"
      }
    }
  },
  "fields": [
    {"name": "customer_name", "type": "text", "label": "Customer Name"},
    {"name": "customer_email", "type": "email", "label": "Email"},
    {"name": "customer_address", "type": "textarea", "label": "Address", "layout": {"colSpan": 2}},
    {"name": "invoice_number", "type": "text", "label": "Invoice #", "readonly": true},
    {"name": "invoice_date", "type": "date", "label": "Invoice Date"},
    {"name": "due_date", "type": "date", "label": "Due Date"},
    {"name": "terms", "type": "select", "label": "Terms"},
    {"name": "line_items", "type": "table_repeater", "label": "Items"},
    {"name": "subtotal", "type": "currency", "label": "Subtotal", "readonly": true},
    {"name": "tax", "type": "currency", "label": "Tax", "readonly": true},
    {"name": "total", "type": "currency", "label": "Total", "readonly": true}
  ]
}
```

### 9.7 Best Practices

#### Choose the Right Layout Type

| Use Case | Layout Type | Why |
|----------|------------|-----|
| Simple form (< 10 fields) | Grid (2 cols) | Clean, organized |
| Long form (> 10 fields) | Sections | Reduces cognitive load |
| Many categories | Tabs | Separates concerns |
| Sequential process | Steps | Guides user flow |
| Mobile-first | Flex (column) | Works on all screens |
| Complex data entry | Grid with sections | Structure + flexibility |

#### Keep It Simple

```json
// ✅ GOOD: Simple, clear
{
  "layout": {
    "type": "grid",
    "columns": 2,
    "gap": "1rem"
  }
}

// ❌ BAD: Over-complicated
{
  "layout": {
    "type": "grid",
    "columns": 12,
    "sections": [...], // Don't mix grid with sections
    "tabs": [...]      // Don't mix tabs with grid
  }
}
```

#### Use Responsive Breakpoints

Always enable responsive:
```json
{
  "layout": {
    "type": "grid",
    "columns": 2,
    "responsive": true  // ← Always true for production
  }
}
```

#### Group Related Fields

```json
// ✅ GOOD: Related fields together
{
  "sections": [
    {
      "id": "name",
      "title": "Name",
      "fields": ["first_name", "middle_name", "last_name"]
    },
    {
      "id": "contact",
      "title": "Contact",
      "fields": ["email", "phone", "address"]
    }
  ]
}

// ❌ BAD: Mixed unrelated fields
{
  "sections": [
    {
      "id": "misc",
      "title": "Information",
      "fields": ["first_name", "email", "payment_method", "shoe_size"]
    }
  ]
}
```

---

