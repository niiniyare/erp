# Schema System Deep Dive

**Part of**: Schema-Driven UI Framework Technical White Paper  
**Version**: 2.0  
**Focus**: Comprehensive Schema Definition and Validation System

---

## Overview

The Schema System forms the foundation of the entire framework, providing type-safe, validated definitions for all UI components. With 900+ comprehensive JSON Schema definitions, it covers the complete spectrum of enterprise UI requirements while maintaining strict validation and inheritance patterns.

## Schema Hierarchy and Organization

### Base Schema Foundation

All schemas inherit from a common base that provides essential properties:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "$$id": {
      "type": "string",
      "description": "Unique component identifier for page designer"
    },
    "className": {
      "$ref": "#/definitions/SchemaClassName",
      "description": "CSS class names for styling"
    },
    "disabled": {
      "type": "boolean",
      "description": "Whether component is disabled"
    },
    "disabledOn": {
      "$ref": "#/definitions/SchemaExpression",
      "description": "Expression-based disable condition"
    },
    "hidden": {
      "type": "boolean",
      "description": "Whether component is hidden"
    },
    "hiddenOn": {
      "$ref": "#/definitions/SchemaExpression",
      "description": "Expression-based visibility condition"
    },
    "visible": {
      "type": "boolean",
      "description": "Whether component is visible"
    },
    "visibleOn": {
      "$ref": "#/definitions/SchemaExpression",
      "description": "Expression-based visibility condition"
    },
    "id": {
      "type": "string",
      "description": "Unique identifier for logging and tracking"
    },
    "style": {
      "type": "object",
      "description": "Inline CSS styles"
    },
    "testid": {
      "type": "string",
      "description": "Test automation identifier"
    }
  }
}
```

### Schema Categories

**Control Schemas (50+ Types)**:
Form input elements with comprehensive validation and interaction patterns.

**Action Schemas (10+ Types)**:
Interactive elements that trigger behaviors like HTTP requests, navigation, or UI state changes.

**Layout Schemas (20+ Types)**:
Container and positioning elements that organize other components.

**Data Display Schemas (30+ Types)**:
Components for presenting information including tables, lists, cards, and visualizations.

**Navigation Schemas (15+ Types)**:
Navigation elements like menus, breadcrumbs, tabs, and pagination.

**Feedback Schemas (10+ Types)**:
User feedback components including notifications, progress indicators, and status displays.

## Control Schema System

### Base Control Pattern

All input controls inherit from a common control schema:

```json
{
  "allOf": [
    {"$ref": "#/definitions/BaseSchema"},
    {
      "type": "object",
      "properties": {
        "name": {
          "type": "string",
          "description": "Form field name for data binding"
        },
        "label": {
          "type": "string",
          "description": "Human-readable field label"
        },
        "labelAlign": {
          "type": "string",
          "enum": ["left", "top", "right"],
          "description": "Label alignment relative to control"
        },
        "placeholder": {
          "type": "string",
          "description": "Placeholder text for empty field"
        },
        "description": {
          "type": "string",
          "description": "Help text displayed with field"
        },
        "required": {
          "type": "boolean",
          "description": "Whether field is required"
        },
        "requiredOn": {
          "$ref": "#/definitions/SchemaExpression",
          "description": "Expression-based required condition"
        },
        "value": {
          "description": "Default or current field value"
        },
        "submitOnChange": {
          "type": "boolean",
          "description": "Submit form when field changes"
        },
        "validateOnChange": {
          "type": "boolean",
          "description": "Validate field on every change"
        },
        "validations": {
          "type": "object",
          "description": "Client-side validation rules"
        },
        "validationErrors": {
          "type": "object",
          "description": "Custom validation error messages"
        }
      }
    }
  ]
}
```

### Specific Control Types

**Text Input Schema**:
```json
{
  "allOf": [
    {"$ref": "#/definitions/ControlSchema"},
    {
      "properties": {
        "type": {"const": "input-text"},
        "minLength": {"type": "number"},
        "maxLength": {"type": "number"},
        "pattern": {"type": "string"},
        "trimContents": {"type": "boolean"},
        "clearable": {"type": "boolean"},
        "revealPassword": {"type": "boolean"},
        "prefix": {"type": "string"},
        "suffix": {"type": "string"},
        "transform": {
          "type": "object",
          "properties": {
            "upperCase": {"type": "boolean"},
            "lowerCase": {"type": "boolean"}
          }
        }
      }
    }
  ]
}
```

**Number Input Schema**:
```json
{
  "allOf": [
    {"$ref": "#/definitions/ControlSchema"},
    {
      "properties": {
        "type": {"const": "input-number"},
        "min": {"type": "number"},
        "max": {"type": "number"},
        "step": {"type": "number"},
        "precision": {"type": "number"},
        "showSteps": {"type": "boolean"},
        "prefix": {"type": "string"},
        "suffix": {"type": "string"},
        "unitOptions": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "label": {"type": "string"},
              "value": {"type": "string"}
            }
          }
        }
      }
    }
  ]
}
```

**Select Control Schema**:
```json
{
  "allOf": [
    {"$ref": "#/definitions/ControlSchema"},
    {
      "properties": {
        "type": {"const": "select"},
        "options": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "label": {"type": "string"},
              "value": {"type": ["string", "number", "boolean"]},
              "disabled": {"type": "boolean"},
              "children": {
                "type": "array",
                "items": {"$ref": "#"}
              }
            }
          }
        },
        "source": {
          "$ref": "#/definitions/BaseApiObject",
          "description": "API endpoint for dynamic options"
        },
        "multiple": {"type": "boolean"},
        "searchable": {"type": "boolean"},
        "clearable": {"type": "boolean"},
        "checkAll": {"type": "boolean"},
        "defaultCheckAll": {"type": "boolean"},
        "joinValues": {"type": "boolean"},
        "extractValue": {"type": "boolean"},
        "valueField": {"type": "string"},
        "labelField": {"type": "string"},
        "searchPromptText": {"type": "string"},
        "noResultsText": {"type": "string"},
        "loadingPlaceholder": {"type": "string"}
      }
    }
  ]
}
```

## Action Schema System

### Action Type Resolution

Actions use conditional schema patterns to provide type-specific properties:

```json
{
  "allOf": [
    {
      "if": {
        "properties": {
          "actionType": {"const": "ajax"}
        }
      },
      "then": {"$ref": "#/definitions/AjaxActionSchema"}
    },
    {
      "if": {
        "properties": {
          "actionType": {"const": "dialog"}
        }
      },
      "then": {"$ref": "#/definitions/DialogActionSchema"}
    },
    {
      "if": {
        "properties": {
          "actionType": {"const": "drawer"}
        }
      },
      "then": {"$ref": "#/definitions/DrawerActionSchema"}
    }
  ]
}
```

### AJAX Action Schema

Comprehensive configuration for server communication:

```json
{
  "type": "object",
  "properties": {
    "actionType": {"const": "ajax"},
    "api": {
      "anyOf": [
        {"type": "string"},
        {"$ref": "#/definitions/BaseApiObject"}
      ],
      "description": "API endpoint configuration"
    },
    "method": {
      "type": "string",
      "enum": ["GET", "POST", "PUT", "PATCH", "DELETE"],
      "default": "POST"
    },
    "data": {
      "type": "object",
      "description": "Additional data to send with request"
    },
    "messages": {
      "type": "object",
      "properties": {
        "success": {"type": "string"},
        "failed": {"type": "string"}
      }
    },
    "feedback": {
      "description": "Success/failure feedback configuration"
    },
    "reload": {
      "type": "string",
      "description": "Target to reload after successful action"
    },
    "redirect": {
      "type": "string",
      "description": "URL to redirect to after successful action"
    },
    "confirmText": {
      "type": "string",
      "description": "Confirmation dialog text"
    },
    "required": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Required fields for action execution"
    }
  }
}
```

## Layout Schema System

### Grid System Schema

Comprehensive grid layout with responsive breakpoints:

```json
{
  "type": "object",
  "properties": {
    "type": {"const": "grid"},
    "columns": {
      "anyOf": [
        {"type": "number"},
        {
          "type": "object",
          "properties": {
            "xs": {"type": "number"},
            "sm": {"type": "number"},
            "md": {"type": "number"},
            "lg": {"type": "number"},
            "xl": {"type": "number"}
          }
        }
      ]
    },
    "gap": {
      "anyOf": [
        {"type": "string"},
        {"type": "number"},
        {
          "type": "object",
          "properties": {
            "x": {"type": ["string", "number"]},
            "y": {"type": ["string", "number"]}
          }
        }
      ]
    },
    "align": {
      "type": "string",
      "enum": ["start", "center", "end", "stretch"]
    },
    "justify": {
      "type": "string",
      "enum": ["start", "center", "end", "between", "around", "evenly"]
    },
    "body": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "columnSpan": {
            "anyOf": [
              {"type": "number"},
              {
                "type": "object",
                "properties": {
                  "xs": {"type": "number"},
                  "sm": {"type": "number"},
                  "md": {"type": "number"},
                  "lg": {"type": "number"},
                  "xl": {"type": "number"}
                }
              }
            ]
          },
          "rowSpan": {"type": "number"},
          "component": {"$ref": "#/definitions/SchemaNode"}
        }
      }
    }
  }
}
```

### Container Schema

Flexible container with extensive layout options:

```json
{
  "type": "object",
  "properties": {
    "type": {"const": "container"},
    "size": {
      "type": "string",
      "enum": ["xs", "sm", "md", "lg", "xl", "full", "none"]
    },
    "wrapperComponent": {
      "type": "string",
      "description": "HTML element to use as wrapper"
    },
    "bodyClassName": {
      "$ref": "#/definitions/SchemaClassName"
    },
    "wrapperBody": {
      "type": "boolean",
      "description": "Whether to wrap body content"
    },
    "style": {
      "type": "object",
      "description": "Container styling"
    },
    "body": {
      "$ref": "#/definitions/SchemaCollection",
      "description": "Container contents"
    }
  }
}
```

## Data Display Schema System

### Table Schema

Enterprise-grade data table with extensive features:

```json
{
  "type": "object",
  "properties": {
    "type": {"const": "table"},
    "source": {
      "$ref": "#/definitions/BaseApiObject",
      "description": "Data source API"
    },
    "columns": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": {"type": "string"},
          "label": {"type": "string"},
          "type": {"type": "string"},
          "width": {"type": ["string", "number"]},
          "minWidth": {"type": ["string", "number"]},
          "sortable": {"type": "boolean"},
          "searchable": {"type": "boolean"},
          "filterable": {"type": "boolean"},
          "toggled": {"type": "boolean"},
          "breakpoint": {
            "type": "string",
            "enum": ["xs", "sm", "md", "lg", "xl"]
          },
          "align": {
            "type": "string",
            "enum": ["left", "center", "right"]
          },
          "valign": {
            "type": "string",
            "enum": ["top", "middle", "bottom"]
          },
          "fixed": {
            "type": "string",
            "enum": ["left", "right"]
          },
          "template": {"type": "string"},
          "component": {"$ref": "#/definitions/SchemaNode"}
        }
      }
    },
    "rowSelection": {
      "type": "object",
      "properties": {
        "type": {
          "type": "string",
          "enum": ["checkbox", "radio"]
        },
        "keyField": {"type": "string"},
        "maxRows": {"type": "number"},
        "columnWidth": {"type": "number"}
      }
    },
    "pagination": {
      "type": "object",
      "properties": {
        "enable": {"type": "boolean"},
        "layout": {"type": "array"},
        "perPage": {"type": "number"},
        "perPageAvailable": {"type": "array"},
        "maxButtons": {"type": "number"},
        "showPageInput": {"type": "boolean"}
      }
    },
    "headerToolbar": {"$ref": "#/definitions/SchemaCollection"},
    "footerToolbar": {"$ref": "#/definitions/SchemaCollection"},
    "autoGenerateFilter": {"type": "boolean"},
    "syncLocation": {"type": "boolean"},
    "keepItemSelectionOnPageChange": {"type": "boolean"},
    "loadDataOnce": {"type": "boolean"},
    "loadDataOnceFetchOnFilter": {"type": "boolean"}
  }
}
```

## Expression System

### Expression Language

Schema expressions enable dynamic behavior through a safe expression language:

```json
{
  "$ref": "#/definitions/SchemaExpression",
  "description": "Expression types",
  "anyOf": [
    {"type": "string"},
    {
      "type": "object",
      "properties": {
        "type": {"const": "expression"},
        "expression": {"type": "string"}
      }
    },
    {
      "type": "object",
      "properties": {
        "type": {"const": "condition"},
        "test": {"$ref": "#/definitions/SchemaExpression"},
        "consequent": {"$ref": "#/definitions/SchemaExpression"},
        "alternate": {"$ref": "#/definitions/SchemaExpression"}
      }
    }
  ]
}
```

### Expression Examples

**Simple Value Reference**:
```json
{
  "visibleOn": "${user.role === 'admin'}"
}
```

**Complex Conditional**:
```json
{
  "disabledOn": {
    "type": "condition",
    "test": "${form.status === 'submitted'}",
    "consequent": true,
    "alternate": "${!form.isValid}"
  }
}
```

**Computed Property**:
```json
{
  "label": {
    "type": "expression",
    "expression": "`Welcome, ${user.firstName} ${user.lastName}`"
  }
}
```

## CSS Property Integration

### CSS Schema System

800+ CSS property schemas provide type-safe styling:

```json
{
  "type": "object",
  "properties": {
    "width": {
      "anyOf": [
        {"type": "string"},
        {"type": "number"},
        {
          "type": "object",
          "properties": {
            "xs": {"type": ["string", "number"]},
            "sm": {"type": ["string", "number"]},
            "md": {"type": ["string", "number"]},
            "lg": {"type": ["string", "number"]},
            "xl": {"type": ["string", "number"]}
          }
        }
      ]
    },
    "height": {"$ref": "#/definitions/SizeProperty"},
    "margin": {"$ref": "#/definitions/SpacingProperty"},
    "padding": {"$ref": "#/definitions/SpacingProperty"},
    "backgroundColor": {"$ref": "#/definitions/ColorProperty"},
    "border": {"$ref": "#/definitions/BorderProperty"},
    "borderRadius": {"$ref": "#/definitions/BorderRadiusProperty"},
    "boxShadow": {"$ref": "#/definitions/ShadowProperty"},
    "display": {
      "type": "string",
      "enum": ["block", "inline", "inline-block", "flex", "grid", "none"]
    },
    "position": {
      "type": "string",
      "enum": ["static", "relative", "absolute", "fixed", "sticky"]
    },
    "flexDirection": {
      "type": "string",
      "enum": ["row", "row-reverse", "column", "column-reverse"]
    },
    "justifyContent": {
      "type": "string",
      "enum": ["flex-start", "flex-end", "center", "space-between", "space-around", "space-evenly"]
    },
    "alignItems": {
      "type": "string",
      "enum": ["flex-start", "flex-end", "center", "baseline", "stretch"]
    }
  }
}
```

### Responsive Styling

```json
{
  "style": {
    "width": {
      "xs": "100%",
      "md": "50%",
      "lg": "33%"
    },
    "margin": {
      "xs": "8px",
      "md": "16px"
    },
    "display": {
      "xs": "block",
      "md": "flex"
    }
  }
}
```

## Schema Validation Pipeline

### Validation Layers

**Layer 1: JSON Schema Validation**
```go
func ValidateAgainstSchema(data interface{}, schemaPath string) error {
    schema, err := loadSchema(schemaPath)
    if err != nil {
        return fmt.Errorf("failed to load schema: %w", err)
    }
    
    result, err := schema.Validate(gojsonschema.NewGoLoader(data))
    if err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    if !result.Valid() {
        return formatValidationErrors(result.Errors())
    }
    
    return nil
}
```

**Layer 2: Business Rule Validation**
```go
func ValidateBusinessRules(schema *Schema, context *ValidationContext) error {
    validator := &BusinessValidator{
        rules: loadBusinessRules(schema.Type),
        context: context,
    }
    
    for _, rule := range validator.rules {
        if err := rule.Validate(schema, context); err != nil {
            return fmt.Errorf("business rule validation failed: %w", err)
        }
    }
    
    return nil
}
```

**Layer 3: Security Validation**
```go
func ValidateSecurityConstraints(schema *Schema, user *User) error {
    // Check component permissions
    if !hasComponentPermission(user, schema.Type) {
        return errors.New("insufficient permissions for component type")
    }
    
    // Validate API endpoints
    if schema.API != nil {
        if !hasAPIPermission(user, schema.API.URL) {
            return errors.New("insufficient API permissions")
        }
    }
    
    // Check data access
    if schema.Source != nil {
        if !hasDataAccess(user, schema.Source) {
            return errors.New("insufficient data access permissions")
        }
    }
    
    return nil
}
```

## Schema Composition Patterns

### Inheritance Patterns

**Single Inheritance**:
```json
{
  "allOf": [
    {"$ref": "#/definitions/BaseSchema"},
    {
      "type": "object",
      "properties": {
        "specificProperty": {"type": "string"}
      }
    }
  ]
}
```

**Multiple Inheritance**:
```json
{
  "allOf": [
    {"$ref": "#/definitions/BaseSchema"},
    {"$ref": "#/definitions/ControlSchema"},
    {"$ref": "#/definitions/ValidatableSchema"},
    {
      "type": "object",
      "properties": {
        "type": {"const": "input-text"}
      }
    }
  ]
}
```

**Conditional Composition**:
```json
{
  "if": {
    "properties": {
      "mode": {"const": "advanced"}
    }
  },
  "then": {
    "allOf": [
      {"$ref": "#/definitions/BaseSchema"},
      {"$ref": "#/definitions/AdvancedFeatures"}
    ]
  },
  "else": {
    "$ref": "#/definitions/BasicSchema"
  }
}
```

### Reference Resolution

```go
type SchemaResolver struct {
    cache    map[string]*Schema
    loader   SchemaLoader
    basePath string
}

func (r *SchemaResolver) ResolveReferences(schema *Schema) (*Schema, error) {
    resolved := &Schema{}
    if err := copier.Copy(resolved, schema); err != nil {
        return nil, err
    }
    
    return r.resolveNode(resolved)
}

func (r *SchemaResolver) resolveNode(node interface{}) (interface{}, error) {
    switch v := node.(type) {
    case map[string]interface{}:
        if ref, hasRef := v["$ref"]; hasRef {
            return r.resolveReference(ref.(string))
        }
        
        result := make(map[string]interface{})
        for key, value := range v {
            resolved, err := r.resolveNode(value)
            if err != nil {
                return nil, err
            }
            result[key] = resolved
        }
        return result, nil
        
    case []interface{}:
        result := make([]interface{}, len(v))
        for i, item := range v {
            resolved, err := r.resolveNode(item)
            if err != nil {
                return nil, err
            }
            result[i] = resolved
        }
        return result, nil
        
    default:
        return v, nil
    }
}
```

---

This comprehensive schema system provides the foundation for type-safe, validated UI generation while maintaining flexibility and extensibility for complex enterprise requirements.