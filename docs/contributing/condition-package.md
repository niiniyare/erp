# Condition Package Documentation

## Overview

The `condition` package provides a production-ready runtime condition evaluation engine for feature flags, ABAC policies, and customer workflows. It enables dynamic evaluation of complex business rules with high performance and comprehensive type safety.

## Table of Contents

- [Architecture](#architecture)
- [Core Concepts](#core-concepts)
- [API Reference](#api-reference)
- [Usage Examples](#usage-examples)
- [JSON Schema Integration](#json-schema-integration)
- [Performance & Security](#performance--security)
- [Testing Strategy](#testing-strategy)
- [Development Guidelines](#development-guidelines)

## Architecture

### Clean Architecture Layers

```
┌─────────────────────────────────────────┐
│              API Layer                  │
│   (Builder, Fluent Interface)          │
├─────────────────────────────────────────┤
│             Domain Layer                │
│  (Evaluator, Rules, Groups, Types)     │
├─────────────────────────────────────────┤
│           Infrastructure                │
│    (Regex Cache, Metrics, Validation)  │
└─────────────────────────────────────────┘
```

### Key Components

1. **Evaluator** - Core evaluation engine with context management
2. **Builder** - Fluent interface for programmatic rule construction
3. **Expressions** - Type-safe value and field access system
4. **Operators** - Comprehensive comparison and logical operators
5. **Context** - Thread-safe evaluation context with metrics
6. **Cache** - Performance optimization for regex and compiled expressions

## Core Concepts

### Condition Rules

A `ConditionRule` represents a single condition with:
- **ID**: Unique identifier for tracking and debugging
- **Left Expression**: Left-hand side value or field reference
- **Operator**: Comparison or logical operator
- **Right Expression**: Right-hand side value(s) for comparison
- **Formula**: Optional expr-lang formula for complex logic

```go
type ConditionRule struct {
    ID    string       `json:"id"`
    Left  Expression   `json:"left"`
    Op    OperatorType `json:"op"`
    Right any          `json:"right,omitempty"`
    If    string       `json:"if,omitempty"` // Formula expression
}
```

### Condition Groups

A `ConditionGroup` represents logical groupings with:
- **Conjunction**: AND/OR logic between children
- **Children**: Rules or nested groups
- **Negation**: NOT operator for entire group
- **Formula**: Optional group-level formula override

```go
type ConditionGroup struct {
    ID          string         `json:"id"`
    Conjunction ConjunctionType `json:"conjunction"`
    Children    []any          `json:"children"`
    Not         bool           `json:"not,omitempty"`
    If          string         `json:"if,omitempty"`
}
```

### Expressions

Type-safe value access supporting:

```go
type Expression struct {
    Type  ValueType `json:"type"`
    Value any       `json:"value,omitempty"`
    Field string    `json:"field,omitempty"`
    Func  *FuncCall `json:"func,omitempty"`
}
```

**Expression Types:**
- `ValueTypeValue` - Static values (strings, numbers, booleans)
- `ValueTypeField` - Dynamic field references with dot notation
- `ValueTypeFunc` - Custom function calls with arguments

### Operators

Comprehensive operator support:

**Comparison Operators:**
- `OpEqual`, `OpNotEqual`
- `OpGreater`, `OpGreaterEqual`, `OpLess`, `OpLessEqual`
- `OpBetween`, `OpNotBetween`

**String Operators:**
- `OpContains`, `OpStartsWith`, `OpEndsWith`
- `OpMatchRegexp` (with ReDoS protection)

**Collection Operators:**
- `OpIn`, `OpNotIn`

**Unary Operators:**
- `OpIsEmpty`, `OpIsNotEmpty`

### Field Access

Supports nested field access via dot notation:
```go
// Access nested object properties
"user.profile.email"
"order.items[0].price"
"settings.notifications.enabled"
```

## API Reference

### Core Evaluation

#### NewEvaluator
```go
func NewEvaluator(config *Config, opts EvalOptions) *Evaluator
```
Creates a new evaluator instance with configuration and options.

#### Evaluate
```go
func (e *Evaluator) Evaluate(ctx context.Context, condition any, evalCtx *EvalContext) (bool, error)
```
Evaluates a condition (rule, group, or map) and returns boolean result.

### Context Management

#### NewEvalContext
```go
func NewEvalContext(data map[string]any, opts EvalOptions) *EvalContext
```
Creates evaluation context with data and options.

#### RegisterFunction
```go
func (ctx *EvalContext) RegisterFunction(name string, fn CustomFunction) error
```
Registers custom functions for expression evaluation.

#### RegisterField
```go
func (ctx *EvalContext) RegisterField(field Field) error
```
Registers field definitions for validation and type checking.

### Builder API

#### NewBuilder
```go
func NewBuilder(conjunction ConjunctionType) *Builder
```
Creates fluent builder for programmatic rule construction.

#### Fluent Methods
```go
// Rule construction
builder.AddRule(field string, op OperatorType, value any) *Builder
builder.AddFieldComparison(leftField, rightField string, op OperatorType) *Builder
builder.AddBetweenRule(field string, min, max any) *Builder
builder.AddInRule(field string, values []any) *Builder

// Formula support
builder.AddFormula(formula string) *Builder
builder.SetFormula(formula string) *Builder

// Group operations
builder.AddGroup(group *ConditionGroup) *Builder
builder.Not() *Builder

// Build result
builder.Build() *ConditionGroup
builder.Validate() error
```

## Usage Examples

### Basic Rule Evaluation

```go
// Create evaluator
config := &Config{
    Types: map[string]TypeConfig{
        "text": {Operators: []OperatorType{OpEqual, OpContains}},
        "number": {Operators: []OperatorType{OpEqual, OpGreater}},
    },
}
evaluator := NewEvaluator(config, DefaultEvalOptions())

// Simple rule
rule := ConditionRule{
    ID: "admin-check",
    Left: Expression{
        Type:  ValueTypeField,
        Field: "user.role",
    },
    Op: OpEqual,
    Right: Expression{
        Type:  ValueTypeValue,
        Value: "admin",
    },
}

// Evaluation context
data := map[string]any{
    "user": map[string]any{
        "role": "admin",
        "active": true,
    },
}
evalCtx := NewEvalContext(data, DefaultEvalOptions())

// Evaluate
result, err := evaluator.Evaluate(context.Background(), &rule, evalCtx)
// result: true, err: nil
```

### Complex Group Logic

```go
// (role = admin) OR ((role = user) AND (premium = true))
group := &ConditionGroup{
    ID:          "access-control",
    Conjunction: ConjunctionOr,
    Children: []any{
        ConditionRule{
            ID: "admin-check",
            Left: Expression{Type: ValueTypeField, Field: "role"},
            Op:   OpEqual,
            Right: Expression{Type: ValueTypeValue, Value: "admin"},
        },
        &ConditionGroup{
            ID:          "premium-user",
            Conjunction: ConjunctionAnd,
            Children: []any{
                ConditionRule{
                    ID: "user-check",
                    Left: Expression{Type: ValueTypeField, Field: "role"},
                    Op:   OpEqual,
                    Right: Expression{Type: ValueTypeValue, Value: "user"},
                },
                ConditionRule{
                    ID: "premium-check",
                    Left: Expression{Type: ValueTypeField, Field: "premium"},
                    Op:   OpEqual,
                    Right: Expression{Type: ValueTypeValue, Value: true},
                },
            },
        },
    },
}
```

### Builder Pattern

```go
// Fluent builder interface
condition := NewBuilder(ConjunctionAnd).
    AddRule("age", OpGreaterEqual, 18).
    AddRule("verified", OpEqual, true).
    AddInRule("country", []any{"US", "CA", "UK"}).
    Build()

// Advanced builder with groups
condition := NewBuilder(ConjunctionOr).
    AddRule("role", OpEqual, "admin").
    AddGroup(
        NewBuilder(ConjunctionAnd).
            AddRule("role", OpEqual, "user").
            AddRule("premium", OpEqual, true).
            Build(),
    ).
    Build()
```

### Formula Expressions

```go
// Using expr-lang formulas for complex logic
rule := ConditionRule{
    ID: "complex-rule",
    If: `age >= 18 && verified && contains(permissions, "admin")`,
}

// Group-level formula
group := &ConditionGroup{
    ID: "formula-group",
    If: `user.role == "admin" || (user.premium && user.credits > 100)`,
    // Children are ignored when formula is present
}
```

### Custom Functions

```go
// Register custom function
evalCtx.RegisterFunction("daysSince", func(ctx context.Context, args []any, evalCtx *EvalContext) (any, error) {
    if len(args) != 1 {
        return nil, errors.New("daysSince requires exactly one argument")
    }
    
    dateStr, ok := args[0].(string)
    if !ok {
        return nil, errors.New("daysSince argument must be a string")
    }
    
    date, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        return nil, err
    }
    
    return int(time.Since(date).Hours() / 24), nil
})

// Use in formula
rule := ConditionRule{
    ID: "date-check",
    If: `daysSince(user.lastLogin) < 30`,
}
```

### Dynamic JSON Evaluation

```go
// Evaluate from JSON/map structures
jsonRule := map[string]any{
    "id": "json-rule",
    "left": map[string]any{
        "type": "field",
        "field": "status",
    },
    "op": "equal",
    "right": map[string]any{
        "type": "value",
        "value": "active",
    },
}

result, err := evaluator.Evaluate(ctx, jsonRule, evalCtx)
```

## JSON Schema Integration

The condition package integrates with comprehensive JSON schemas for frontend components:

### Schema Files Structure

```
docs/ui/Schema/definitions/
├── ConditionRule.json              # Rule definition schema
├── ConditionGroupValue.json        # Group definition schema
├── ConditionBuilderField.json      # Field configuration schema
├── ConditionBuilderConfig.json     # Builder configuration schema
├── OperatorType.json              # Operator definitions
└── ExpressionComplex.json         # Expression schemas
```

### ConditionRule Schema

```json
{
  "type": "object",
  "properties": {
    "id": {},
    "left": {"$ref": "#/definitions/ExpressionComplex"},
    "op": {"$ref": "#/definitions/OperatorType"},
    "right": {
      "anyOf": [
        {"$ref": "#/definitions/ExpressionComplex"},
        {
          "type": "array",
          "items": {"$ref": "#/definitions/ExpressionComplex"}
        }
      ]
    },
    "if": {"type": "string"}
  },
  "required": ["id"],
  "additionalProperties": false
}
```

### ConditionGroup Schema

```json
{
  "type": "object",
  "properties": {
    "id": {"type": "string"},
    "conjunction": {
      "type": "string",
      "enum": ["and", "or"]
    },
    "not": {"type": "boolean"},
    "children": {
      "type": "array",
      "items": {
        "anyOf": [
          {"$ref": "#/definitions/ConditionRule"},
          {"$ref": "#/definitions/ConditionGroupValue"}
        ]
      }
    },
    "if": {"type": "string"}
  },
  "required": ["id", "conjunction"]
}
```

### Operator Types

Supported operators in JSON schema:
- `"equal"`, `"not_equal"`
- `"greater"`, `"greater_or_equal"`, `"less"`, `"less_or_equal"`
- `"between"`, `"not_between"`
- `"like"`, `"not_like"`, `"starts_with"`, `"ends_with"`
- `"is_empty"`, `"is_not_empty"`
- `"select_equals"`, `"select_not_equals"`
- `"select_any_in"`, `"select_not_any_in"`

## Performance & Security

### Performance Features

1. **Regex Caching**: Compiled regex patterns cached with LRU eviction
2. **Expression Compilation**: Pre-compiled expr-lang formulas
3. **Concurrent Safety**: Thread-safe evaluation contexts
4. **Resource Limits**: Configurable depth, condition count, and timeout limits
5. **Metrics Collection**: Performance monitoring with atomic counters

```go
type EvalOptions struct {
    MaxDepth      int           // Maximum nesting depth (default: 10)
    MaxConditions int           // Maximum conditions per evaluation (default: 100)
    Timeout       time.Duration // Overall evaluation timeout (default: 5s)
    RegexTimeout  time.Duration // Regex match timeout (default: 100ms)
}
```

### Security Features

1. **ReDoS Protection**: Regex timeouts prevent catastrophic backtracking
2. **Input Validation**: Comprehensive validation of all inputs
3. **Resource Limits**: Prevent memory/CPU exhaustion attacks
4. **Type Safety**: Strong typing prevents injection attacks
5. **Audit Trail**: Comprehensive logging and metrics

### Performance Benchmarks

```
BenchmarkEvaluator_SimpleRule-8        79204    15402 ns/op
BenchmarkEvaluator_ComplexGroup-8      56959    25416 ns/op  
BenchmarkEvaluator_Formula-8          180090     7747 ns/op
BenchmarkEvaluator_NestedGroups-8      55039    25270 ns/op
```

## Testing Strategy

### Test Coverage: 79.0%

Comprehensive test suites covering:

1. **BasicComparisonTestSuite** - Core operator functionality
2. **StringOperatorsTestSuite** - String manipulation operators
3. **UnaryOperatorsTestSuite** - IsEmpty/IsNotEmpty operators
4. **BetweenOperatorTestSuite** - Range comparisons
5. **InOperatorTestSuite** - Collection membership
6. **RegexOperatorTestSuite** - Pattern matching with ReDoS protection
7. **GroupsTestSuite** - Logical grouping and nesting
8. **FormulaTestSuite** - expr-lang formula evaluation
9. **ResourceLimitsTestSuite** - Security and performance limits
10. **CustomFunctionsTestSuite** - Function registration and execution
11. **MetricsTestSuite** - Performance monitoring
12. **BuilderTestSuite** - Fluent interface functionality
13. **ValidationTestSuite** - Input validation and error handling
14. **FieldTestSuite** - Field management and registration
15. **DynamicDataTestSuite** - JSON/map evaluation
16. **TypeConversionTestSuite** - Type coercion edge cases
17. **ErrorHandlingTestSuite** - Error scenarios and recovery

### Test Examples

```go
func TestComplexCondition(t *testing.T) {
    evaluator := NewEvaluator(testConfig, DefaultEvalOptions())
    
    // Test complex nested logic
    condition := NewBuilder(ConjunctionOr).
        AddRule("role", OpEqual, "admin").
        AddGroup(
            NewBuilder(ConjunctionAnd).
                AddRule("role", OpEqual, "user").
                AddRule("premium", OpEqual, true).
                Build(),
        ).
        Build()
    
    tests := []struct {
        name string
        data map[string]any
        want bool
    }{
        {"admin user", map[string]any{"role": "admin"}, true},
        {"premium user", map[string]any{"role": "user", "premium": true}, true},
        {"regular user", map[string]any{"role": "user", "premium": false}, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            evalCtx := NewEvalContext(tt.data, DefaultEvalOptions())
            result, err := evaluator.Evaluate(context.Background(), condition, evalCtx)
            assert.NoError(t, err)
            assert.Equal(t, tt.want, result)
        })
    }
}
```

## Development Guidelines

### Adding New Operators

1. **Define Operator Constant**
```go
const (
    OpNewOperator OperatorType = "new_operator"
)
```

2. **Update Operator Mapping**
```go
func (e *Evaluator) applyOperator(left, right any, op OperatorType) (bool, error) {
    switch op {
    case OpNewOperator:
        return e.handleNewOperator(left, right)
    // ... existing cases
    }
}
```

3. **Implement Handler**
```go
func (e *Evaluator) handleNewOperator(left, right any) (bool, error) {
    // Implementation with proper type checking and error handling
}
```

4. **Add Tests**
```go
func (s *OperatorTestSuite) TestNewOperator() {
    // Comprehensive test cases including edge cases
}
```

5. **Update Documentation**
- Add to operator list in this documentation
- Update JSON schema if needed
- Add usage examples

### Error Handling Patterns

```go
// Use context for error details
return fmt.Errorf("field access failed for %q: %w", fieldPath, err)

// Validate inputs early
if rule.ID == "" {
    return fmt.Errorf("%w: rule ID is required", ErrValidation)
}

// Handle type conversion gracefully
func toFloat64(v any) (float64, error) {
    switch val := v.(type) {
    case float64:
        return val, nil
    case int:
        return float64(val), nil
    case string:
        return strconv.ParseFloat(val, 64)
    default:
        return 0, fmt.Errorf("cannot convert %T to float64", v)
    }
}
```

### Performance Considerations

1. **Use Atomic Operations for Metrics**
```go
atomic.AddInt32(&ctx.metrics.RulesEvaluated, 1)
```

2. **Implement Caching Strategically**
```go
// Cache expensive operations like regex compilation
if cached, exists := e.regexCache.Get(pattern); exists {
    return cached.(*regexp2.Regexp), nil
}
```

3. **Validate Resource Limits**
```go
if depth > e.opts.MaxDepth {
    return false, fmt.Errorf("%w: depth %d exceeds limit %d", 
        ErrMaxDepthExceeded, depth, e.opts.MaxDepth)
}
```

### Security Best Practices

1. **Always Set Timeouts**
```go
re.MatchTimeout = e.opts.RegexTimeout
```

2. **Validate Inputs Thoroughly**
```go
func (r *ConditionRule) Validate() error {
    if r.If != "" {
        return nil // Formula rules skip traditional validation
    }
    // ... validation logic
}
```

3. **Use Context for Cancellation**
```go
func (e *Evaluator) Evaluate(ctx context.Context, condition any, evalCtx *EvalContext) (bool, error) {
    select {
    case <-ctx.Done():
        return false, ctx.Err()
    default:
        // Continue evaluation
    }
}
```

## Integration Examples

### Feature Flags

```go
// Feature flag evaluation
featureFlag := ConditionRule{
    ID: "new-dashboard",
    Left: Expression{Type: ValueTypeField, Field: "user.beta_tester"},
    Op:   OpEqual,
    Right: Expression{Type: ValueTypeValue, Value: true},
}

enabled, _ := evaluator.Evaluate(ctx, &featureFlag, userContext)
if enabled {
    // Show new dashboard
}
```

### ABAC Policies

```go
// Attribute-based access control
policy := &ConditionGroup{
    ID:          "document-access",
    Conjunction: ConjunctionAnd,
    Children: []any{
        ConditionRule{
            ID: "department-match",
            Left: Expression{Type: ValueTypeField, Field: "user.department"},
            Op:   OpEqual,
            Right: Expression{Type: ValueTypeField, Field: "document.department"},
        },
        ConditionRule{
            ID: "security-clearance",
            Left: Expression{Type: ValueTypeField, Field: "user.clearance_level"},
            Op:   OpGreaterEqual,
            Right: Expression{Type: ValueTypeField, Field: "document.required_clearance"},
        },
    },
}
```

### Workflow Rules

```go
// Business workflow conditions
approvalRule := NewBuilder(ConjunctionOr).
    AddRule("amount", OpLess, 1000).  // Auto-approve small amounts
    AddGroup(
        NewBuilder(ConjunctionAnd).
            AddRule("approver.role", OpEqual, "manager").
            AddRule("approver.department", OpEqual, "finance").
            Build(),
    ).
    Build()
```

## Conclusion

The condition package provides a robust, secure, and performant foundation for dynamic rule evaluation in enterprise applications. Its clean architecture, comprehensive test coverage, and JSON schema integration make it suitable for production use in feature flags, access control, and workflow automation systems.

For questions or contributions, please refer to the project's contributing guidelines and open an issue for discussion.