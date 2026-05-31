# Go Model Auto-Generation System - User Guide

## Overview

The Go Model Auto-Generation System (Phase 4.3) automatically generates complete UI schemas from Go struct definitions. This system analyzes your Go models and creates production-ready UI components including CRUD tables, forms, detail views, and specialized patterns like Kanban boards and dashboards.

## Features

- **Automatic Struct Analysis**: Parses Go files and extracts comprehensive field metadata
- **Tag-Based Customization**: Uses struct tags to control UI generation behavior
- **Intelligent Pattern Matching**: Automatically selects optimal UI patterns based on entity characteristics
- **Complete CRUD Generation**: Creates list, create, edit, detail, and filter schemas
- **Multi-Format Output**: Supports JSON, YAML, and Go code generation
- **Template System**: Extensible template-driven generation
- **CLI Tool**: Command-line interface for integration into build pipelines

## Quick Start

### 1. Basic Usage

Generate UI schemas for a specific Go file:

```bash
go run web/generator/cli/main.go -file models/user.go -model User -verbose
```

Generate schemas for an entire package:

```bash
go run web/generator/cli/main.go -package ./internal/models -verbose
```

### 2. Watch Mode

Automatically regenerate when files change:

```bash
go run web/generator/cli/main.go -package ./models -watch -verbose
```

### 3. Preview Mode

Preview generated schemas without writing files:

```bash
go run web/generator/cli/main.go -file models/user.go -preview
```

## Struct Tag System

The system supports comprehensive tag-based customization using multiple tag types:

### UI Tags

Control component rendering and behavior:

```go
type User struct {
    Email    string `ui:"component=email;label=Email Address;required=true;placeholder=Enter your email"`
    Username string `ui:"component=text;label=Username;unique=true"`
    Role     string `ui:"component=select;options=admin,user,guest"`
    IsActive bool   `ui:"component=toggle;label=Active Status"`
    Bio      string `ui:"component=textarea;rows=4;placeholder=Tell us about yourself"`
}
```

**Available UI Tag Options:**
- `component`: UI component type (text, email, select, textarea, etc.)
- `label`: Display label for the field
- `required`: Mark field as required (true/false)
- `placeholder`: Placeholder text
- `options`: Comma-separated list of options for select components
- `hidden`: Hide field from UI (true/false)
- `readonly`: Make field read-only (true/false)
- `validation`: Custom validation rules
- `help`: Help text or tooltip
- `icon`: Icon to display with field

### Validation Tags

Define validation rules:

```go
type User struct {
    Email    string `validate:"required,email"`
    Age      int    `validate:"min=18,max=100"`
    Username string `validate:"required,minlen=3,maxlen=50,alphanum"`
    Website  string `validate:"url"`
}
```

### Database Tags

Provide database context:

```go
type User struct {
    ID       uuid.UUID `db:"id,primary_key"`
    Email    string    `db:"email,unique,not null"`
    Username string    `db:"username,unique,index"`
}
```

### Form Tags

Control form layout and behavior:

```go
type User struct {
    Email     string `form:"tab=basic;section=contact"`
    Phone     string `form:"tab=basic;section=contact;inline=true"`
    Bio       string `form:"tab=profile;section=about;cols=12"`
    Interests string `form:"dependsOn=hasInterests;showWhen=true"`
}
```

### Table Tags

Configure table display:

```go
type User struct {
    Name      string    `table:"width=200px;sortable=true"`
    Email     string    `table:"width=250px;searchable=true"`
    CreatedAt time.Time `table:"width=150px;align=center;sortable=true"`
    IsActive  bool      `table:"width=100px;align=center"`
}
```

### Filter Tags

Define filtering capabilities:

```go
type User struct {
    Status    string    `filter:"type=select;preset=true"`
    CreatedAt time.Time `filter:"type=date-range;range=true"`
    Role      string    `filter:"type=select;multiple=true"`
}
```

## CLI Reference

### Basic Commands

```bash
# Generate for specific file and model
go run web/generator/cli/main.go -file user.go -model User

# Generate for entire package
go run web/generator/cli/main.go -package ./models

# Preview without writing files
go run web/generator/cli/main.go -file user.go -preview

# Watch for changes
go run web/generator/cli/main.go -package ./models -watch
```

### Configuration Options

| Flag | Description | Default |
|------|-------------|---------|
| `-file` | Input Go file to analyze | |
| `-package` | Input Go package directory | |
| `-model` | Specific struct name to generate | |
| `-output` | Output directory for schemas | `./generated` |
| `-format` | Output format (json, yaml, go) | `json` |
| `-layout` | Default layout template | `app` |
| `-api-prefix` | API endpoint prefix | `/api` |
| `-page-size` | Default table page size | `20` |
| `-patterns` | Specific UI patterns to generate | |
| `-templates` | Custom template directory | |
| `-config` | Configuration file path | |
| `-permissions` | Enable permission controls | `true` |
| `-audit` | Enable audit trail features | `true` |
| `-validate` | Validate generated schemas | `true` |
| `-force` | Force overwrite existing files | `false` |
| `-verbose` | Enable verbose logging | `false` |

### Pattern Selection

Generate specific UI patterns:

```bash
# Generate only CRUD table and create form
go run web/generator/cli/main.go -file user.go -patterns crud_table,create_form

# Generate Kanban board for task entities
go run web/generator/cli/main.go -file task.go -patterns kanban_board

# Generate dashboard for analytics entities
go run web/generator/cli/main.go -file analytics.go -patterns dashboard
```

## Supported UI Patterns

The system automatically detects and generates appropriate UI patterns:

### Basic CRUD Patterns
- **crud_table**: Standard data table with sorting, filtering, pagination
- **create_form**: Form for creating new records
- **edit_form**: Form for editing existing records
- **detail_view**: Detailed view of single records
- **search_filter**: Advanced filtering interface

### Advanced Patterns
- **kanban_board**: Workflow management with drag-and-drop
- **hierarchy_tree**: Tree structure for hierarchical data
- **wizard_form**: Multi-step forms for complex data entry
- **dashboard**: Analytics dashboard with metrics
- **calendar_view**: Calendar interface for date-based data

### Specialized Patterns
- **user_profile**: User profile management interface
- **contact_card**: Contact information display
- **invoice_form**: Financial document management
- **task_board**: Project management interface
- **notifications**: Notification center

## Entity Detection

The system automatically detects entity characteristics:

### CRUD Entities
Entities with primary keys automatically get full CRUD interfaces:

```go
type User struct {
    ID        uuid.UUID `db:"id,primary_key"`
    Email     string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Workflow Entities
Entities with status fields get workflow patterns:

```go
type Task struct {
    ID     uuid.UUID
    Status string `ui:"component=select;options=todo,in-progress,done"`
    // Automatically gets Kanban board pattern
}
```

### Hierarchical Entities
Entities with parent relationships get tree patterns:

```go
type Department struct {
    ID       uuid.UUID
    ParentID *uuid.UUID
    Name     string
    // Automatically gets hierarchy tree pattern
}
```

### Read-Only Entities
Entities without update timestamps get report patterns:

```go
type Report struct {
    ID        uuid.UUID
    CreatedAt time.Time
    // No UpdatedAt - detected as read-only
}
```

## Generated Schema Structure

The system generates complete UI schemas with the following structure:

```json
{
  "entity": "User",
  "package": "models",
  "title": "User Management",
  "version": "1.0.0",
  "listSchema": {
    "id": "user-list",
    "layout": "app",
    "title": "User List",
    "components": [...],
    "dataSources": [...],
    "actions": [...]
  },
  "createSchema": {
    "id": "user-create",
    "layout": "app", 
    "title": "Create User",
    "components": [...],
    "actions": [...]
  },
  "editSchema": {...},
  "detailSchema": {...},
  "filterSchema": {...},
  "metadata": {
    "isCrud": true,
    "hasValidation": true,
    "hasAuditTrail": true,
    "primaryKey": "ID",
    "relationships": [...]
  }
}
```

## Integration Examples

### Build Pipeline Integration

Add to your Makefile:

```makefile
.PHONY: generate-ui
generate-ui:
	go run web/generator/cli/main.go -package ./internal/models -output ./web/schemas
	
.PHONY: watch-ui
watch-ui:
	go run web/generator/cli/main.go -package ./internal/models -watch -verbose
```

### Git Hooks Integration

Pre-commit hook to regenerate schemas:

```bash
#!/bin/sh
# .git/hooks/pre-commit
go run web/generator/cli/main.go -package ./internal/models -validate
```

### CI/CD Integration

GitHub Actions workflow:

```yaml
name: Generate UI Schemas
on:
  push:
    paths:
      - 'internal/models/**/*.go'

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      - run: go run web/generator/cli/main.go -package ./internal/models
      - uses: stefanzweifel/git-auto-commit-action@v4
        with:
          commit_message: 'Auto-generate UI schemas'
```

## Configuration File

Create a configuration file for complex setups:

```yaml
# ui-generator.yml
generator:
  defaultLayout: "app"
  apiPrefix: "/api/v1"
  tablePageSize: 25
  enablePermissions: true
  enableAuditTrail: true
  
  customComponents:
    richtext: "organisms.rich-text-editor"
    fileupload: "molecules.file-upload"
    
  patterns:
    - crud_table
    - create_form
    - edit_form
    - detail_view
    
  templates:
    directory: "./custom-templates"
    
  output:
    directory: "./generated/schemas"
    format: "json"
    
  validation:
    strict: true
    rules:
      - "no-empty-labels"
      - "required-primary-key"
```

Use with CLI:

```bash
go run web/generator/cli/main.go -config ui-generator.yml -package ./models
```

## Best Practices

### 1. Struct Design

Design structs with UI generation in mind:

```go
// Good: Clear field names, appropriate tags
type User struct {
    ID        uuid.UUID `db:"id,primary_key" json:"id"`
    Email     string    `db:"email,unique" json:"email" ui:"component=email;required=true;label=Email Address"`
    FirstName string    `db:"first_name" json:"firstName" ui:"label=First Name;required=true"`
    LastName  string    `db:"last_name" json:"lastName" ui:"label=Last Name;required=true"`
    Role      string    `db:"role" json:"role" ui:"component=select;options=admin,user,guest;label=User Role"`
    IsActive  bool      `db:"is_active" json:"isActive" ui:"component=toggle;label=Active"`
    CreatedAt time.Time `db:"created_at" json:"createdAt"`
    UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}
```

### 2. Tag Organization

Organize tags consistently:

```go
type Product struct {
    // Core fields with full tag specification
    Name        string  `db:"name" json:"name" ui:"component=text;label=Product Name;required=true" validate:"required,minlen=3"`
    Description string  `db:"description" json:"description" ui:"component=textarea;label=Description;rows=4"`
    Price       float64 `db:"price" json:"price" ui:"component=number;label=Price;min=0" validate:"min=0"`
    
    // Relationship fields
    CategoryID  uuid.UUID `db:"category_id" json:"categoryId" ui:"component=select;relation=Category;label=Category"`
    
    // Metadata fields
    CreatedAt   time.Time `db:"created_at" json:"createdAt" table:"sortable=true;width=150px"`
    UpdatedAt   time.Time `db:"updated_at" json:"updatedAt" table:"sortable=true;width=150px"`
}
```

### 3. Pattern Optimization

Structure entities for optimal pattern detection:

```go
// Task entity optimized for Kanban pattern
type Task struct {
    ID          uuid.UUID `db:"id,primary_key"`
    Title       string    `ui:"label=Task Title;required=true"`
    Description string    `ui:"component=textarea;label=Description"`
    Status      string    `ui:"component=select;options=todo,in-progress,review,done;label=Status"`
    Priority    string    `ui:"component=select;options=low,medium,high,urgent;label=Priority"`
    AssigneeID  uuid.UUID `ui:"component=select;relation=User;label=Assignee"`
    DueDate     time.Time `ui:"component=date;label=Due Date"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 4. Testing Generated Schemas

Validate generated schemas in tests:

```go
func TestUserSchemaGeneration(t *testing.T) {
    analyzer := generator.NewAnalyzer()
    structInfo := analyzer.AnalyzeStruct(reflect.TypeOf(User{}))
    
    config := generator.GeneratorConfig{
        DefaultLayout: "app",
        EnablePermissions: true,
    }
    
    gen := generator.NewSchemaGenerator(config)
    schema, err := gen.GenerateUISchema(structInfo)
    
    assert.NoError(t, err)
    assert.Equal(t, "User", schema.Entity)
    assert.NotNil(t, schema.ListSchema)
    assert.NotNil(t, schema.CreateSchema)
}
```

## Troubleshooting

### Common Issues

1. **No schemas generated**
   - Check that structs have exported fields
   - Verify primary key field exists
   - Use `-verbose` flag for detailed logging

2. **Missing components**
   - Ensure UI tags are properly formatted
   - Check for typos in component names
   - Verify field types are supported

3. **Validation errors**
   - Review struct tag syntax
   - Check for conflicting tag values
   - Use `-validate` flag to identify issues

### Debug Mode

Enable debug logging:

```bash
go run web/generator/cli/main.go -file user.go -verbose -validate
```

### Schema Validation

Validate generated schemas:

```bash
go run web/generator/cli/main.go -package ./models -validate -preview
```

## Extension Points

The system is designed for extensibility:

### Custom Components

Register custom components in configuration:

```yaml
customComponents:
  geocoder: "molecules.geocoder-input"
  colorpicker: "molecules.color-picker"
  signature: "organisms.signature-pad"
```

### Custom Patterns

Add custom pattern matchers:

```go
func init() {
    matcher := generator.NewPatternMatcher()
    matcher.AddRule(generator.PatternType("custom_report"), 10.0, func(s generator.StructInfo) (bool, string) {
        return strings.Contains(s.Name, "Report"), "Custom report pattern"
    })
}
```

### Custom Templates

Create custom generation templates in your templates directory following the existing JSON schema format.

## Performance Considerations

- **Package Analysis**: Large packages may take time to analyze
- **Watch Mode**: Efficient file watching with debounced regeneration
- **Caching**: Generated schemas are cached for unchanged files
- **Parallel Processing**: Multiple files processed concurrently

## Security Notes

- Generated schemas include permission metadata
- Sensitive fields can be marked as hidden
- Validation rules are preserved in generated schemas
- Access control patterns are automatically applied

## Support

For issues, questions, or contributions:
- Review the source code in `web/generator/`
- Check existing documentation in `docs/ui/`
- Test with the demo: `go run web/generator/demo/main.go`