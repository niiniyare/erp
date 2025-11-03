# Renderer (templ)

## Overview

Converts schemas to HTML using type-safe templ components.

## Basic Structure

```go
// views/form.templ
package views

import "github.com/niiniyare/erp/pkg/schema"

templ FormPage(s *schema.Schema, data map[string]any) {
    <!DOCTYPE html>
    <html>
    <head>
        <title>{ s.Title }</title>
    </head>
    <body>
        <h1>{ s.Title }</h1>
        @FormRenderer(s, data)
    </body>
    </html>
}

templ FormRenderer(s *schema.Schema, data map[string]any) {
    <form action={ s.Config.Action } method={ s.Config.Method }>
        for _, field := range s.Fields {
            @FieldRenderer(&field, data[field.Name])
        }
    </form>
}

templ FieldRenderer(field *schema.Field, value any) {
    switch field.Type {
        case "text", "email":
            <input type={ string(field.Type) } name={ field.Name } />
        case "textarea":
            <textarea name={ field.Name }></textarea>
        case "select":
            @SelectField(field, value)
    }
}
```

## Field Type Mapping

| Field Type | templ Component |
|-----------|-----------------|
| `text`, `email`, `password` | `TextInput` |
| `textarea` | `TextareaInput` |
| `select` | `SelectInput` |
| `radio` | `RadioInput` |
| `checkbox` | `CheckboxInput` |
| `date` | `DateInput` |
| `file` | `FileInput` |

## HTMX Integration

```go
templ FormRenderer(s *schema.Schema, data map[string]any) {
    <form
        if s.HTMX != nil && s.HTMX.Enabled {
            hx-post={ s.HTMX.Post }
            hx-target={ s.HTMX.Target }
            hx-swap={ s.HTMX.Swap }
        }
    >
        @FieldsRenderer(s.Fields, data)
    </form>
}
```

## Validation Attributes

```go
templ TextInput(field *schema.Field, value string) {
    <input
        type={ string(field.Type) }
        name={ field.Name }
        value={ value }
        if field.Required {
            required
        }
        if field.Validation != nil {
            if field.Validation.MinLength != nil {
                minlength={ strconv.Itoa(*field.Validation.MinLength) }
            }
            if field.Validation.MaxLength != nil {
                maxlength={ strconv.Itoa(*field.Validation.MaxLength) }
            }
        }
    />
}
```

## Generate and Use

```bash
# Generate Go code from templ
cd views
templ generate

# Use in handler
return views.FormPage(schema, data).Render(ctx, w)
```

[← Back](11-enricher.md) | [Next: Testing →](13-testing.md)