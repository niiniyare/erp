# Schema System Extensions - Implementation Plan

**Selected Categories for Implementation**

1. Schema Management & Organization
2. Field Types & Input Controls
3. Validation & Business Rules
4. Data Transformation & Processing

---

## 📋 Executive Summary

This plan outlines the implementation of 40+ features across 4 foundational categories. These extensions will transform your schema system into a production-grade, enterprise-ready solution for Awo ERP.

**Implementation Timeline**: 2-3 weeks
**Priority Level**: Critical for production readiness
**Complexity**: Medium to High

---

## 1. Schema Management & Organization

### 1.1 Schema Composition & Inheritance 🔥

**Priority**: Critical
**Complexity**: 🟡 Medium (4-5 days)
**Impact**: Eliminates code duplication, enables reusable patterns

#### Features to Implement:

##### A. Base Schema Extension
```go
type BaseSchema struct {
    ID          string
    Name        string
    Description string
    Fields      []Field
    Metadata    map[string]any
}

type ExtendedSchema struct {
    Base       *BaseSchema
    Overrides  map[string]Field  // Override specific fields
    Additional []Field            // Add new fields
}
```

**Use Cases**:
- Common customer schema extended for B2B vs B2C
- Base invoice schema extended for sales vs purchase invoices
- Employee schema extended for contractors vs full-time

##### B. Schema Mixins
```go
type Mixin struct {
    ID          string
    Name        string
    Fields      []Field
    Validation  []ValidationRule
    Metadata    map[string]any
}

type Schema struct {
    // ... existing fields
    Mixins      []string  // Mixin IDs to include
}
```

**Pre-built Mixins**:
- `audit_fields`: created_at, updated_at, created_by, updated_by
- `address_fields`: street, city, state, postal_code, country
- `contact_fields`: phone, email, fax, website
- `financial_fields`: currency, amount, tax_rate, tax_amount
- `status_fields`: status, is_active, is_deleted

##### C. Schema Composition
```go
type ComposedSchema struct {
    ID         string
    Name       string
    Components []SchemaComponent
}

type SchemaComponent struct {
    SchemaID   string
    Prefix     string  // Namespace for fields
    Required   bool    // Is this component required?
    Condition  *Condition  // When to include this component
}
```

**Example**:
```json
{
  "id": "sales_order",
  "name": "Sales Order",
  "components": [
    {
      "schema_id": "customer_info",
      "prefix": "customer",
      "required": true
    },
    {
      "schema_id": "shipping_info",
      "prefix": "shipping",
      "required": false,
      "condition": {
        "field": "requires_shipping",
        "operator": "equals",
        "value": true
      }
    }
  ]
}
```

#### Implementation Files:
- `pkg/schema/composition/base.go`
- `pkg/schema/composition/mixin.go`
- `pkg/schema/composition/composer.go`
- `pkg/schema/composition/resolver.go`

---

### 1.2 Schema Versioning 🔥

**Priority**: Critical
**Complexity**: 🟠 Hard (1-2 weeks)
**Impact**: Safe schema evolution, backward compatibility

#### Features to Implement:

##### A. Version Management
```go
type SchemaVersion struct {
    SchemaID      string
    Version       string  // semver: 1.2.3
    CreatedAt     time.Time
    CreatedBy     string
    ChangeLog     []Change
    Deprecated    bool
    DeprecatedAt  *time.Time
    Schema        *Schema
}

type Change struct {
    Type        ChangeType  // added, modified, removed, deprecated
    Target      string      // field name or schema component
    Description string
    Breaking    bool
}
```

##### B. Migration System
```go
type Migration struct {
    FromVersion   string
    ToVersion     string
    Up            MigrationFunc
    Down          MigrationFunc
    DataTransform func(map[string]any) (map[string]any, error)
}

type MigrationFunc func(ctx context.Context, old *Schema) (*Schema, error)
```

**Example Migration**:
```go
// Rename field from "customer_name" to "customer_full_name"
migration := Migration{
    FromVersion: "1.0.0",
    ToVersion: "1.1.0",
    Up: func(ctx context.Context, old *Schema) (*Schema, error) {
        // Schema transformation logic
    },
    DataTransform: func(data map[string]any) (map[string]any, error) {
        if name, ok := data["customer_name"]; ok {
            data["customer_full_name"] = name
            delete(data, "customer_name")
        }
        return data, nil
    },
}
```

##### C. Version Compatibility Checker
```go
type CompatibilityChecker struct {
    registry *SchemaRegistry
}

func (c *CompatibilityChecker) Check(old, new *Schema) (*CompatibilityReport, error)

type CompatibilityReport struct {
    Compatible      bool
    Breaking        []BreakingChange
    Warnings        []Warning
    Recommendations []string
}
```

#### Implementation Files:
- `pkg/schema/version/version.go`
- `pkg/schema/version/migration.go`
- `pkg/schema/version/compatibility.go`
- `pkg/schema/version/changelog.go`

---

### 1.3 Schema Registry & Discovery ⭐

**Priority**: High
**Complexity**: 🟡 Medium (3-4 days)
**Impact**: Centralized schema management, improved discoverability

#### Features to Implement:

##### A. Schema Registry
```go
type SchemaRegistry struct {
    store    SchemaStore
    cache    Cache
    index    SearchIndex
    hooks    []RegistryHook
}

type SchemaStore interface {
    Save(ctx context.Context, schema *Schema) error
    Get(ctx context.Context, id string, version ...string) (*Schema, error)
    List(ctx context.Context, filter *Filter) ([]*Schema, error)
    Delete(ctx context.Context, id string) error
    Versions(ctx context.Context, id string) ([]*SchemaVersion, error)
}
```

##### B. Schema Search & Discovery
```go
type SearchIndex interface {
    Index(schema *Schema) error
    Search(query *SearchQuery) (*SearchResults, error)
    Suggest(prefix string) ([]string, error)
}

type SearchQuery struct {
    Text       string
    Tags       []string
    Category   string
    FieldTypes []FieldType
    TenantMode TenantMode
    Limit      int
    Offset     int
}
```

##### C. Schema Relationships
```go
type SchemaGraph struct {
    nodes map[string]*SchemaNode
    edges map[string][]*SchemaEdge
}

type SchemaNode struct {
    Schema     *Schema
    References []*SchemaReference
    Referenced []*SchemaReference
    Mixins     []*Mixin
    Extends    *Schema
}

type SchemaReference struct {
    FromSchema string
    FromField  string
    ToSchema   string
    Type       ReferenceType  // lookup, composition, inheritance
}
```

#### Implementation Files:
- `pkg/schema/registry/registry.go`
- `pkg/schema/registry/store/postgres.go`
- `pkg/schema/registry/store/memory.go`
- `pkg/schema/registry/search/index.go`
- `pkg/schema/registry/graph/graph.go`

---

### 1.4 Schema Organization 💡

**Priority**: Medium
**Complexity**: 🟢 Easy (2-3 days)
**Impact**: Better organization for large projects

#### Features to Implement:

##### A. Namespaces
```go
type Namespace struct {
    ID          string
    Name        string
    Description string
    Parent      *Namespace
    Children    []*Namespace
    Schemas     []string  // Schema IDs
}

// Example: com.awoerp.finance.invoice
```

##### B. Schema Categories & Tags
```go
type Category struct {
    ID          string
    Name        string
    Description string
    Icon        string
    Color       string
}

type Tag struct {
    Name  string
    Color string
}

type Schema struct {
    // ... existing fields
    Category    string
    Tags        []string
    Namespace   string
}
```

**Pre-defined Categories**:
- `financial`: Finance-related schemas
- `inventory`: Inventory management
- `sales`: Sales & CRM
- `purchasing`: Procurement
- `hr`: Human resources
- `manufacturing`: Production
- `system`: System/admin schemas

#### Implementation Files:
- `pkg/schema/organization/namespace.go`
- `pkg/schema/organization/category.go`
- `pkg/schema/organization/tags.go`

---

## 2. Field Types & Input Controls

### 2.1 Advanced Field Types 🔥

**Priority**: Critical
**Complexity**: 🟠 Hard (1-2 weeks)
**Impact**: Covers 90% of ERP use cases

#### Features to Implement:

##### A. Rich Text Editor Field
```go
type RichTextField struct {
    BaseField
    Editor      RichTextEditor  // tinymce, quill, ckeditor
    Toolbar     []string        // bold, italic, link, image, etc.
    MaxLength   int
    AllowImages bool
    AllowTables bool
    AllowHTML   bool
    Sanitize    bool
}
```

##### B. Code Editor Field
```go
type CodeEditorField struct {
    BaseField
    Language    string  // go, javascript, sql, json, yaml
    Theme       string  // vs-dark, vs-light
    LineNumbers bool
    TabSize     int
    ReadOnly    bool
    Validate    bool  // Syntax validation
}
```

##### C. JSON Editor Field
```go
type JSONEditorField struct {
    BaseField
    Schema      *JSONSchema  // JSON Schema for validation
    Mode        string       // tree, code, form
    MaxDepth    int
    Sortable    bool
    Searchable  bool
}
```

##### D. Barcode/QR Scanner Field
```go
type BarcodeScannerField struct {
    BaseField
    Format      []BarcodeFormat  // qr, ean13, code128, etc.
    Camera      bool             // Enable camera scanning
    Manual      bool             // Allow manual entry
    Validate    func(string) bool
}
```

##### E. Signature Pad Field
```go
type SignatureField struct {
    BaseField
    Width       int
    Height      int
    PenColor    string
    Background  string
    Format      string  // png, jpg, svg
    Required    bool
}
```

##### F. Location/Map Field
```go
type LocationField struct {
    BaseField
    Provider    string  // google, mapbox, openstreetmap
    DefaultZoom int
    AllowDrag   bool
    ShowAddress bool
    Geocode     bool  // Convert to lat/lng
}
```

##### G. Rating Field
```go
type RatingField struct {
    BaseField
    MaxRating   int
    Icon        string  // star, heart, thumbs
    Size        string  // small, medium, large
    AllowHalf   bool
    ShowLabel   bool
}
```

##### H. Slider Range Field
```go
type SliderRangeField struct {
    BaseField
    Min         float64
    Max         float64
    Step        float64
    DualHandle  bool
    ShowTicks   bool
    ShowTooltip bool
    Unit        string
}
```

##### I. Tag/Chip Input Field
```go
type TagInputField struct {
    BaseField
    MaxTags     int
    Suggestions []string
    Autocomplete bool
    AllowCustom  bool
    CaseSensitive bool
    Delimiter    string
}
```

##### J. Transfer List Field
```go
type TransferListField struct {
    BaseField
    Available   []Option
    ShowSearch  bool
    Pagination  bool
    ItemsPerPage int
    Sortable    bool
}
```

#### Implementation Files:
- `pkg/schema/fields/richtext/richtext.go`
- `pkg/schema/fields/code/editor.go`
- `pkg/schema/fields/json/editor.go`
- `pkg/schema/fields/barcode/scanner.go`
- `pkg/schema/fields/signature/pad.go`
- `pkg/schema/fields/location/map.go`
- `pkg/schema/fields/rating/rating.go`
- `pkg/schema/fields/slider/range.go`
- `pkg/schema/fields/tag/input.go`
- `pkg/schema/fields/transfer/list.go`

---

### 2.2 Field Collections & Repeaters 🔥

**Priority**: Critical (Essential for invoices, line items)
**Complexity**: 🟠 Hard (1 week)
**Impact**: Core ERP functionality

#### Features to Implement:

##### A. Repeatable Field Groups
```go
type RepeatableField struct {
    BaseField
    Template    []Field         // Fields to repeat
    MinItems    int
    MaxItems    int
    Sortable    bool
    Collapsible bool
    ItemLabel   string          // Template for item labels
    DefaultItem map[string]any  // Default values for new items
}
```

**Example - Invoice Line Items**:
```json
{
  "type": "repeatable",
  "name": "line_items",
  "label": "Line Items",
  "template": [
    {
      "type": "lookup",
      "name": "product_id",
      "label": "Product",
      "required": true
    },
    {
      "type": "number",
      "name": "quantity",
      "label": "Quantity",
      "default": 1
    },
    {
      "type": "currency",
      "name": "unit_price",
      "label": "Unit Price"
    },
    {
      "type": "calculated",
      "name": "line_total",
      "label": "Total",
      "formula": "quantity * unit_price"
    }
  ],
  "min_items": 1,
  "sortable": true,
  "item_label": "Product {product_name}"
}
```

##### B. Table Repeater
```go
type TableRepeaterField struct {
    RepeatableField
    Layout      string  // table, cards, accordion
    Editable    bool    // Inline editing
    BulkActions []Action
    RowActions  []Action
    Footer      bool    // Show totals footer
    Aggregates  map[string]AggregateFunc
}
```

##### C. Nested Repeaters
```go
type NestedRepeaterField struct {
    RepeatableField
    MaxDepth    int
    ChildField  string  // Field name for child items
}
```

**Example - Project with Tasks and Subtasks**:
```json
{
  "type": "nested_repeatable",
  "name": "tasks",
  "label": "Project Tasks",
  "template": [
    {
      "type": "text",
      "name": "task_name",
      "label": "Task"
    },
    {
      "type": "date",
      "name": "due_date",
      "label": "Due Date"
    }
  ],
  "child_field": "subtasks",
  "max_depth": 3
}
```

##### D. Repeater Operations
```go
type RepeaterOperations interface {
    Add(ctx context.Context, item map[string]any) error
    Remove(ctx context.Context, index int) error
    Update(ctx context.Context, index int, item map[string]any) error
    Move(ctx context.Context, from, to int) error
    Clear(ctx context.Context) error
    Import(ctx context.Context, items []map[string]any) error
    Export(ctx context.Context, format string) ([]byte, error)
}
```

#### Implementation Files:
- `pkg/schema/fields/repeatable/repeatable.go`
- `pkg/schema/fields/repeatable/table.go`
- `pkg/schema/fields/repeatable/nested.go`
- `pkg/schema/fields/repeatable/operations.go`
- `pkg/schema/fields/repeatable/validation.go`

---

### 2.3 Dynamic Fields ⭐

**Priority**: High
**Complexity**: 🟡 Medium (4-5 days)
**Impact**: Advanced user experience

#### Features to Implement:

##### A. Calculated Fields
```go
type CalculatedField struct {
    BaseField
    Formula     string
    Dependencies []string
    Precision   int
    UpdateOn    string  // change, blur, manual
    ReadOnly    bool
}

type FormulaEngine interface {
    Evaluate(formula string, context map[string]any) (any, error)
    Validate(formula string) error
    GetDependencies(formula string) []string
}
```

**Supported Formulas**:
- Arithmetic: `+, -, *, /, %, ^`
- Functions: `SUM(), AVG(), MIN(), MAX(), COUNT(), IF(), ROUND()`
- Date functions: `TODAY(), NOW(), DATEADD(), DATEDIFF()`
- Text functions: `CONCAT(), UPPER(), LOWER(), TRIM()`
- Lookup functions: `LOOKUP(), VLOOKUP()`

##### B. Cascading Dropdowns
```go
type CascadingDropdownField struct {
    SelectField
    ParentField string
    DataSource  DataSourceFunc
    CacheKey    string
}

type DataSourceFunc func(ctx context.Context, parentValue any) ([]Option, error)
```

**Example - Country → State → City**:
```json
{
  "type": "select",
  "name": "country",
  "label": "Country",
  "options_source": "countries_table"
},
{
  "type": "cascading_select",
  "name": "state",
  "label": "State/Province",
  "parent_field": "country",
  "data_source": {
    "type": "database",
    "query": "SELECT id, name FROM states WHERE country_id = {{parent_value}}"
  }
},
{
  "type": "cascading_select",
  "name": "city",
  "label": "City",
  "parent_field": "state",
  "data_source": {
    "type": "database",
    "query": "SELECT id, name FROM cities WHERE state_id = {{parent_value}}"
  }
}
```

##### C. Lookup Fields
```go
type LookupField struct {
    BaseField
    TargetSchema string
    TargetField  string
    DisplayFields []string
    SearchFields  []string
    Filters      map[string]any
    MinSearch    int
    MaxResults   int
    CreateNew    bool
}
```

##### D. Polymorphic Fields
```go
type PolymorphicField struct {
    BaseField
    TypeField    string
    TypeMap      map[string][]Field
}
```

**Example - Contact (Person or Company)**:
```json
{
  "type": "select",
  "name": "contact_type",
  "options": ["person", "company"]
},
{
  "type": "polymorphic",
  "name": "contact_details",
  "type_field": "contact_type",
  "type_map": {
    "person": [
      {"type": "text", "name": "first_name"},
      {"type": "text", "name": "last_name"},
      {"type": "date", "name": "birth_date"}
    ],
    "company": [
      {"type": "text", "name": "company_name"},
      {"type": "text", "name": "tax_id"},
      {"type": "number", "name": "employees"}
    ]
  }
}
```

#### Implementation Files:
- `pkg/schema/fields/calculated/calculated.go`
- `pkg/schema/fields/calculated/formula.go`
- `pkg/schema/fields/cascading/cascading.go`
- `pkg/schema/fields/lookup/lookup.go`
- `pkg/schema/fields/polymorphic/polymorphic.go`

---

## 3. Validation & Business Rules

### 3.1 Advanced Validation 🔥

**Priority**: Critical
**Complexity**: 🟠 Hard (1 week)
**Impact**: Data integrity, business logic enforcement

#### Features to Implement:

##### A. Custom Validator Registration
```go
type ValidatorRegistry struct {
    validators map[string]ValidatorFunc
    mu         sync.RWMutex
}

type ValidatorFunc func(ctx context.Context, value any, params map[string]any) error

func (r *ValidatorRegistry) Register(name string, fn ValidatorFunc)
func (r *ValidatorRegistry) Validate(ctx context.Context, name string, value any, params map[string]any) error
```

**Built-in Validators**:
- `required`: Field is required
- `email`: Valid email address
- `url`: Valid URL
- `phone`: Valid phone number (international)
- `min/max`: Min/max value or length
- `regex`: Regular expression match
- `unique`: Unique value in database
- `exists`: Value exists in another table
- `luhn`: Credit card Luhn validation
- `iban`: IBAN validation
- `tax_id`: Country-specific tax ID validation

##### B. Async Validators
```go
type AsyncValidator struct {
    Name       string
    Validate   func(ctx context.Context, value any) error
    Debounce   time.Duration
    Cache      bool
    CacheTTL   time.Duration
}

// Example: Check if email is already registered
asyncValidator := AsyncValidator{
    Name: "unique_email",
    Validate: func(ctx context.Context, value any) error {
        email := value.(string)
        exists, err := db.EmailExists(ctx, email)
        if err != nil {
            return err
        }
        if exists {
            return errors.New("email already registered")
        }
        return nil
    },
    Debounce: 500 * time.Millisecond,
    Cache: true,
    CacheTTL: 5 * time.Minute,
}
```

##### C. Cross-Field Validation
```go
type CrossFieldValidator struct {
    Name       string
    Fields     []string
    Validate   func(ctx context.Context, values map[string]any) error
}

// Example: Start date must be before end date
crossValidator := CrossFieldValidator{
    Name: "date_range",
    Fields: []string{"start_date", "end_date"},
    Validate: func(ctx context.Context, values map[string]any) error {
        start := values["start_date"].(time.Time)
        end := values["end_date"].(time.Time)
        if start.After(end) {
            return errors.New("start date must be before end date")
        }
        return nil
    },
}
```

##### D. Conditional Validation
```go
type ConditionalValidation struct {
    Condition  *Condition
    Rules      []ValidationRule
}

// Example: Require shipping address only if requires_shipping is true
conditionalValidation := ConditionalValidation{
    Condition: &Condition{
        Field: "requires_shipping",
        Operator: "equals",
        Value: true,
    },
    Rules: []ValidationRule{
        {Field: "shipping_address", Rule: "required"},
        {Field: "shipping_city", Rule: "required"},
        {Field: "shipping_postal_code", Rule: "required"},
    },
}
```

##### E. Database Validation
```go
type DatabaseValidator struct {
    Name       string
    Query      string
    Params     []string
    ErrorMsg   string
}

// Example: Check if product is in stock
dbValidator := DatabaseValidator{
    Name: "product_in_stock",
    Query: "SELECT stock_quantity FROM products WHERE id = $1 AND stock_quantity >= $2",
    Params: []string{"product_id", "quantity"},
    ErrorMsg: "insufficient stock for product",
}
```

##### F. Business Rule Engine
```go
type BusinessRule struct {
    ID          string
    Name        string
    Description string
    Condition   *Condition
    Actions     []RuleAction
    Priority    int
    Enabled     bool
}

type RuleAction struct {
    Type    ActionType  // validate, transform, notify, execute
    Config  map[string]any
}

type RuleEngine interface {
    AddRule(rule *BusinessRule) error
    RemoveRule(id string) error
    Evaluate(ctx context.Context, data map[string]any) (*RuleResult, error)
}
```

#### Implementation Files:
- `pkg/schema/validation/registry.go`
- `pkg/schema/validation/async.go`
- `pkg/schema/validation/cross_field.go`
- `pkg/schema/validation/conditional.go`
- `pkg/schema/validation/database.go`
- `pkg/schema/validation/rules/engine.go`
- `pkg/schema/validation/rules/rule.go`

---

### 3.2 Validation Rules Library ⭐

**Priority**: High
**Complexity**: 🟡 Medium (3-4 days)
**Impact**: Industry-specific validation

#### Pre-built Rule Sets:

##### A. Financial Rules
```go
// Amount within limits
AmountRange(min, max decimal.Decimal)

// Currency format validation
CurrencyFormat(currency string)

// Accounting period validation
AccountingPeriod(fiscalYear int)

// GL account validation
GLAccountExists(accountCode string)

// Tax calculation validation
TaxCalculation(taxRate decimal.Decimal)

// Credit limit check
CreditLimit(customerID string, amount decimal.Decimal)

// Payment terms validation
PaymentTerms(terms string)
```

##### B. Inventory Rules
```go
// Stock availability
StockAvailable(productID string, quantity int)

// Reorder point check
BelowReorderPoint(productID string)

// Lot number validation
LotNumberFormat(lotNumber string)

// Serial number validation
SerialNumberUnique(serialNumber string)

// Expiry date validation
ExpiryDate(expiryDate time.Time)

// Warehouse location validation
WarehouseLocation(location string)

// UOM validation
UnitOfMeasure(uom string)
```

##### C. HR Rules
```go
// Leave balance check
LeaveBalance(employeeID string, days int)

// Working hours validation
WorkingHours(hours float64)

// Salary range validation
SalaryRange(position string, salary decimal.Decimal)

// Employee age validation
EmployeeAge(birthDate time.Time, minAge, maxAge int)

// Department validation
DepartmentExists(deptID string)

// Manager validation
ManagerReports(managerID, employeeID string)
```

##### D. Compliance Rules
```go
// GDPR consent validation
GDPRConsent(purpose string)

// Data retention validation
DataRetention(date time.Time, retentionPeriod time.Duration)

// PII validation
PIIMasking(fieldType string)

// Audit trail validation
AuditTrail(action string)

// Access control validation
AccessControl(userID, resource string)
```

#### Implementation Files:
- `pkg/schema/validation/rules/financial/financial.go`
- `pkg/schema/validation/rules/inventory/inventory.go`
- `pkg/schema/validation/rules/hr/hr.go`
- `pkg/schema/validation/rules/compliance/compliance.go`

---

### 3.3 Enhanced Error Handling 💡

**Priority**: Medium
**Complexity**: 🟢 Easy (2-3 days)
**Impact**: Better user experience

#### Features to Implement:

##### A. Validation Levels
```go
type ValidationLevel string

const (
    ValidationError   ValidationLevel = "error"    // Blocks submission
    ValidationWarning ValidationLevel = "warning"  // Shows warning, allows submission
    ValidationInfo    ValidationLevel = "info"     // Informational message
    ValidationSuccess ValidationLevel = "success"  // Positive feedback
)

type ValidationResult struct {
    Level   ValidationLevel
    Field   string
    Message string
    Details map[string]any
}
```

##### B. Soft Validation
```go
type SoftValidation struct {
    Rule        ValidationRule
    Message     string
    Overridable bool
    RequireReason bool
}

// Example: Warn if discount exceeds 20% but allow override
softValidation := SoftValidation{
    Rule: ValidationRule{
        Field: "discount_percent",
        Operator: "greater_than",
        Value: 20,
    },
    Message: "Discount exceeds 20%. Are you sure?",
    Overridable: true,
    RequireReason: true,
}
```

##### C. Error Recovery Suggestions
```go
type ErrorRecovery struct {
    Error       error
    Suggestions []RecoverySuggestion
}

type RecoverySuggestion struct {
    Type        SuggestionType  // fix, alternative, help
    Description string
    Action      *Action
}

// Example: Invalid email format
recovery := ErrorRecovery{
    Error: errors.New("invalid email format"),
    Suggestions: []RecoverySuggestion{
        {
            Type: SuggestionTypeFix,
            Description: "Did you mean: user@example.com?",
            Action: &Action{
                Type: "replace_value",
                Value: "user@example.com",
            },
        },
        {
            Type: SuggestionTypeHelp,
            Description: "Email must include @ and domain",
        },
    },
}
```

##### D. Error Templates
```go
type ErrorTemplate struct {
    Code     string
    Template string
    Severity ValidationLevel
    Category string
}

// Example templates
templates := []ErrorTemplate{
    {
        Code: "required_field",
        Template: "{{.FieldLabel}} is required",
        Severity: ValidationError,
        Category: "required",
    },
    {
        Code: "invalid_format",
        Template: "{{.FieldLabel}} has invalid format. Expected: {{.ExpectedFormat}}",
        Severity: ValidationError,
        Category: "format",
    },
    {
        Code: "out_of_range",
        Template: "{{.FieldLabel}} must be between {{.Min}} and {{.Max}}",
        Severity: ValidationError,
        Category: "range",
    },
}
```

#### Implementation Files:
- `pkg/schema/validation/levels.go`
- `pkg/schema/validation/soft.go`
- `pkg/schema/validation/recovery.go`
- `pkg/schema/validation/templates.go`

---

## 4. Data Transformation & Processing

### 4.1 Field Transformers 🔥

**Priority**: Critical
**Complexity**: 🟡 Medium (4-5 days)
**Impact**: Data quality, consistency

#### Features to Implement:

##### A. Input Transformers (Pre-validation)
```go
type InputTransformer interface {
    Transform(ctx context.Context, value any) (any, error)
    Name() string
}

// Built-in transformers
type TrimTransformer struct {}
func (t *TrimTransformer) Transform(ctx context.Context, value any) (any, error) {
    if s, ok := value.(string); ok {
        return strings.TrimSpace(s), nil
    }
    return value, nil
}

type LowercaseTransformer struct {}
type UppercaseTransformer struct {}
type SlugifyTransformer struct {}
type RemoveSpecialCharsTransformer struct {}
type NormalizeWhitespaceTransformer struct {}
type RemoveDiacriticsTransformer struct {}
type ParseNumberTransformer struct {}
type ParseDateTransformer struct {
    Format string
}
type FormatPhoneTransformer struct {
    Country string
}
```

##### B. Output Transformers (Post-validation)
```go
type OutputTransformer interface {
    Transform(ctx context.Context, value any) (any, error)
    Name() string
}

// Built-in transformers
type EncryptTransformer struct {
    Key []byte
}

type HashTransformer struct {
    Algorithm string  // bcrypt, argon2, sha256
}

type FormatCurrencyTransformer struct {
    Currency string
    Locale   string
}

type FormatDateTransformer struct {
    Format string
    Timezone string
}

type RedactPIITransformer struct {
    FieldTypes []string
}

type ThumbnailTransformer struct {
    Width  int
    Height int
    Quality int
}
```

##### C. Transformer Registry
```go
type TransformerRegistry struct {
    transformers map[string]Transformer
    mu           sync.RWMutex
}

func (r *TransformerRegistry) Register(name string, t Transformer)
func (r *TransformerRegistry) Get(name string) (Transformer, error)
func (r *TransformerRegistry) Apply(ctx context.Context, name string, value any) (any, error)
```

#### Implementation Files:
- `pkg/schema/transform/transformer.go`
- `pkg/schema/transform/input/trim.go`
- `pkg/schema/transform/input/case.go`
- `pkg/schema/transform/input/normalize.go`
- `pkg/schema/transform/output/encrypt.go`
- `pkg/schema/transform/output/hash.go`
- `pkg/schema/transform/output/format.go`
- `pkg/schema/transform/registry.go`

---

### 4.2 Data Pipeline ⭐

**Priority**: High
**Complexity**: 🟡 Medium (3-4 days)
**Impact**: Complex data processing

#### Features to Implement:

##### A. Transformation Pipeline
```go
type Pipeline struct {
    Name        string
    Steps       []PipelineStep
    ErrorMode   ErrorMode  // stop, continue, collect
}

type PipelineStep struct {
    Name        string
    Transformer Transformer
    Condition   *Condition
    OnError     func(error) error
}

type PipelineEngine interface {
    Execute(ctx context.Context, data map[string]any) (map[string]any, error)
    Validate(pipeline *Pipeline) error
}
```

**Example Pipeline**:
```json
{
  "name": "customer_data_pipeline",
  "steps": [
    {
      "name": "trim_whitespace",
      "transformer": "trim",
      "fields": ["first_name", "last_name", "email"]
    },
    {
      "name": "normalize_email",
      "transformer": "lowercase",
      "fields": ["email"]
    },
    {
      "name": "format_phone",
      "transformer": "format_phone",
      "fields": ["phone"],
      "params": {"country": "US"}
    },
    {
      "name": "hash_password",
      "transformer": "bcrypt",
      "fields": ["password"],
      "condition": {
        "field": "password",
        "operator": "not_empty"
      }
    }
  ]
}
```

##### B. Conditional Transforms
```go
type ConditionalTransform struct {
    Condition   *Condition
    Transformer Transformer
    Else        Transformer
}
```

##### C. Batch Transforms
```go
type BatchTransformer interface {
    TransformBatch(ctx context.Context, values []any) ([]any, error)
}

type BatchPipeline struct {
    Pipeline    *Pipeline
    BatchSize   int
    Parallel    bool
    Workers     int
}
```

#### Implementation Files:
- `pkg/schema/transform/pipeline/pipeline.go`
- `pkg/schema/transform/pipeline/step.go`
- `pkg/schema/transform/pipeline/engine.go`
- `pkg/schema/transform/pipeline/batch.go`

---

### 4.3 Data Enrichment 💡

**Priority**: Medium
**Complexity**: 🟡 Medium (4-5 days)
**Impact**: Enhanced data quality

#### Features to Implement:

##### A. Auto-complete Data
```go
type DataEnricher interface {
    Enrich(ctx context.Context, field string, value any) (map[string]any, error)
}

// Example: Address enrichment
type AddressEnricher struct {
    Provider string  // google, mapbox
    APIKey   string
}

func (e *AddressEnricher) Enrich(ctx context.Context, field string, address any) (map[string]any, error) {
    // Returns: street, city, state, postal_code, country, lat, lng
}

// Example: Company lookup
type CompanyEnricher struct {
    Provider string  // clearbit, fullcontact
    APIKey   string
}

func (e *CompanyEnricher) Enrich(ctx context.Context, field string, domain any) (map[string]any, error) {
    // Returns: company_name, industry, employee_count, revenue, etc.
}
```

##### B. Validation Enrichment
```go
// Email validation
type EmailValidator struct {
    Provider string  // zerobounce, mailgun
    APIKey   string
}

func (v *EmailValidator) Validate(ctx context.Context, email string) (*EmailValidation, error)

type EmailValidation struct {
    Valid       bool
    Deliverable bool
    Disposable  bool
    Role        bool
    Suggestion  string
}

// Phone validation
type PhoneValidator struct {
    Provider string  // twilio, numverify
    APIKey   string
}

func (v *PhoneValidator) Validate(ctx context.Context, phone string) (*PhoneValidation, error)

type PhoneValidation struct {
    Valid       bool
    CountryCode string
    Carrier     string
    LineType    string  // mobile, landline, voip
    Location    string
}
```

##### C. Currency Conversion
```go
type CurrencyConverter struct {
    Provider string  // openexchangerates, fixer
    APIKey   string
    CacheTTL time.Duration
}

func (c *CurrencyConverter) Convert(ctx context.Context, amount decimal.Decimal, from, to string) (decimal.Decimal, error)
func (c *CurrencyConverter) GetRate(ctx context.Context, from, to string) (decimal.Decimal, error)
```

##### D. Tax Calculation
```go
type TaxCalculator struct {
    Provider string  // taxjar, avalara
    APIKey   string
}

func (t *TaxCalculator) Calculate(ctx context.Context, req *TaxRequest) (*TaxResponse, error)

type TaxRequest struct {
    Amount      decimal.Decimal
    FromAddress Address
    ToAddress   Address
    ProductCode string
}

type TaxResponse struct {
    TaxAmount    decimal.Decimal
    TaxRate      decimal.Decimal
    Breakdown    []TaxBreakdown
}
```

#### Implementation Files:
- `pkg/schema/enrichment/enricher.go`
- `pkg/schema/enrichment/address.go`
- `pkg/schema/enrichment/company.go`
- `pkg/schema/enrichment/email.go`
- `pkg/schema/enrichment/phone.go`
- `pkg/schema/enrichment/currency.go`
- `pkg/schema/enrichment/tax.go`

---

## 5. Implementation Strategy

### Phase 1: Foundation (Week 1)
**Days 1-3**: Schema Composition & Inheritance
- Base schema extension
- Mixins
- Composer and resolver

**Days 4-5**: Field Collections & Repeaters
- Repeatable field groups
- Table repeater
- Basic operations

**Days 6-7**: Custom Validators & Registry
- Validator registry
- Built-in validators
- Cross-field validation

### Phase 2: Core Features (Week 2)
**Days 1-3**: Advanced Field Types
- Rich text, code editor, JSON editor
- Barcode scanner, signature pad
- Rating, slider, tag input

**Days 3-5**: Calculated Fields & Cascading
- Formula engine
- Calculated fields
- Cascading dropdowns

**Days 6-7**: Field Transformers & Pipeline
- Input/output transformers
- Transformer registry
- Basic pipeline

### Phase 3: Advanced Features (Week 3)
**Days 1-2**: Schema Versioning
- Version management
- Basic migrations
- Compatibility checking

**Days 3-4**: Schema Registry
- Registry implementation
- PostgreSQL store
- Basic search

**Days 5-7**: Business Rules & Enrichment
- Rule engine
- Industry-specific rules
- Data enrichment services

---

## 6. Testing Strategy

### Unit Tests
- Test each field type independently
- Test validators in isolation
- Test transformers with edge cases
- Test formula engine calculations

### Integration Tests
- Test schema composition end-to-end
- Test repeatable fields with validation
- Test pipeline with multiple steps
- Test enrichment with mock services

### Performance Tests
- Benchmark validator performance
- Benchmark transformer performance
- Test with large datasets (1000+ fields)
- Test concurrent operations

### E2E Tests
- Test complete form workflows
- Test invoice creation with line items
- Test cascading dropdowns
- Test calculated fields updates

---

## 7. Documentation Plan

### Developer Documentation
- API reference for each component
- Code examples for common use cases
- Migration guides from old system
- Best practices and patterns

### User Documentation
- Field type catalog with screenshots
- Validation rule reference
- Formula syntax guide
- Troubleshooting guide

### Video Tutorials
- Creating schemas with composition
- Using repeatable fields for line items
- Building calculated fields
- Setting up validation rules

---

## 8. Success Metrics

### Code Quality
- ✅ 80%+ test coverage
- ✅ Zero critical security issues
- ✅ All linting checks pass
- ✅ Documentation for all public APIs

### Performance
- ✅ Schema validation < 100ms (99th percentile)
- ✅ Field rendering < 50ms (99th percentile)
- ✅ Transformer pipeline < 200ms (99th percentile)
- ✅ Support 1000+ field schemas

### Usability
- ✅ Reduce form creation time by 50%
- ✅ Zero-config for 80% of use cases
- ✅ Clear error messages
- ✅ Comprehensive examples

---

## Next Steps

1. **Review this plan** and confirm priorities
2. **Approve the approach** for each category
3. **Start implementation** with Phase 1
4. **Iterate and refine** based on feedback

Ready to begin implementation? Let me know which phase you'd like to start with!
