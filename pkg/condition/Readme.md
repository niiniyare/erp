# Condition Package Documentation

## Overview

The `condition` package provides a production-ready runtime condition evaluation engine for feature flags, attribute-based access control (ABAC) policies, and customer workflows. It offers a safe, performant, and flexible way to evaluate complex business logic without executing arbitrary code.

**Key Features:**
- Safe condition evaluation with resource limits and timeout protection
- Support for complex nested logic with AND/OR/NOT operations
- 17+ comparison operators (equality, range, pattern matching, membership tests)
- Custom function support for domain-specific logic
- expr-lang formula expressions for advanced workflows
- Thread-safe concurrent evaluation
- Performance optimization through intelligent caching
- Protection against ReDoS attacks and resource exhaustion
- Comprehensive error handling and observability metrics

## Installation

```bash
go get github.com/niiniyare/erp/pkg/condition
```

## Quick Start

### Basic Usage

```go
package main

import (
	"context"
	"github.com/niiniyare/erp/condition"
)

func main() {
	// Create evaluation context with data
	ctx := condition.NewEvalContext(
		map[string]any{
			"age":           25,
			"premium":       true,
			"country":       "US",
		},
		condition.DefaultEvalOptions(),
	)

	// Build a simple condition
	builder := condition.NewBuilder(condition.ConjunctionAnd)
	builder.AddRule("age", condition.OpGreaterOrEqual, 18)
	builder.AddRule("premium", condition.OpEqual, true)
	group := builder.Build()

	// Evaluate
	evaluator := condition.NewEvaluator(nil, condition.DefaultEvalOptions())
	result, err := evaluator.Evaluate(context.Background(), group, ctx)
	if err != nil {
		panic(err)
	}
	
	println(result) // true
}
```

### Using Formulas

```go
builder := condition.NewBuilder(condition.ConjunctionAnd)
builder.AddFormula("age >= 18 && (premium || vip)")
group := builder.Build()

result, err := evaluator.Evaluate(context.Background(), group, ctx)
```

## Theoretical Foundation

### Condition Evaluation as Decision Logic

Condition evaluation is fundamentally about converting data into decisions through logical inference. This is rooted in formal logic and constraint satisfaction:

**Boolean Algebra Foundations:**
Conditions use Boolean operations (AND, OR, NOT) which follow De Morgan's laws and other logical equivalences. The engine efficiently evaluates conditions by:
- Short-circuiting: stopping evaluation once the result is determined
- Lazy evaluation: only computing branches necessary to reach a conclusion
- Precedence: respecting operator precedence in formulas

Example: `(A AND B) OR (NOT C)` stops evaluating once either the AND succeeds or C is proven false.

**Temporal Logic:**
Many business conditions involve time: "users over 18", "events after timestamp T", "recurring weekly actions". The engine treats `Now` as a special variable in evaluation context, allowing reproducible time-based decisions. This differs from real-time systems that continuously react to time changes—the condition engine evaluates at a snapshot.

**Attribute-Based Access Control (ABAC):**
The condition engine implements ABAC, which uses attributes (user properties, resource properties, environment) rather than roles to make authorization decisions. This is more flexible than role-based access control (RBAC) because policies can combine arbitrary attributes:
- RBAC: "admins can delete documents" (coarse-grained)
- ABAC: "users in engineering dept with 5+ years exp can delete documents they own or created before this quarter" (fine-grained)

The trade-off: ABAC is more powerful but harder to audit and understand at scale.

**Complexity Theory:**
Evaluating arbitrary nested conditions is NP-complete in the general case. The engine mitigates this through:
- Depth limits: preventing exponential branching
- Condition count limits: preventing explosion of sub-expressions
- Caching: avoiding re-evaluation of identical branches
- Timeouts: preventing worst-case scenarios from blocking

These limits are not eliminating the problem but making it practical and predictable for business logic.

### Storage and Persistence Considerations

The condition package is **stateless and runtime-only**. It performs no persistence operations:

**What It Does NOT Do:**
- Write to databases or files
- Maintain permanent audit logs
- Store evaluation history
- Persist cache between process restarts
- Track state across requests

**Why This Design:**
1. **Statelessness enables horizontal scaling**: Multiple instances can evaluate conditions independently
2. **Decoupling concerns**: Storage/persistence is the responsibility of integrating systems
3. **Flexibility**: Each system decides its own persistence strategy
4. **Security**: No sensitive data accidentally stored in logs

**Caching Behavior:**
Caching occurs only in-memory during the process lifetime:
- Regex cache: LRU-bounded to 1000 compiled patterns
- Formula cache: Unbounded (potential memory issue with high formula diversity)
- Cache invalidation: Only when process restarts (no TTL or manual invalidation)

This design is appropriate for:
- Short-lived processes (request handlers, cron jobs)
- Conditions that change infrequently
- Systems where memory pressure is monitored

It's problematic for:
- Long-running services with evolving conditions
- Memory-constrained environments
- Systems needing formula cache invalidation

**When You Need Storage:**

Store externally if you need to:

1. **Version conditions over time**: Who changed what policy and when?
   - Solution: Store conditions in a database with Git-like versioning
   
2. **Persist evaluation results**: Did user X have access on date Y?
   - Solution: Write results to an audit log or data warehouse
   
3. **A/B test conditions**: Run different rules for different user cohorts?
   - Solution: Store condition variants in database, select at runtime
   
4. **Reload conditions without restart**: Update rules without process restart?
   - Solution: Periodically fetch conditions from database, rebuild Evaluator

### Formula Cache Memory Growth Issue

**Known limitation:** The formula compilation cache in `Evaluator` has no bounds or eviction. In systems evaluating many unique formulas, this causes unbounded memory growth:

```go
// This leaks memory - each unique formula stays forever
for i := 0; i < 1_000_000; i++ {
	formula := fmt.Sprintf("value == %d", i)
	evaluator.evaluateFormula(ctx, formula, evalCtx) // Formula cached forever
}
```

**Workaround:** Implement periodic Evaluator recreation:
```go
// Recreate evaluator every N evaluations or time period
if time.Since(lastRecreate) > 1*time.Hour {
	evaluator = condition.NewEvaluator(config, opts)
	lastRecreate = time.Now()
}
```

Or pre-compile known formulas at startup rather than runtime.

## Core Concepts

### Field Types

Fields define the data types that conditions can operate on:

- `FieldTypeText`: String values
- `FieldTypeNumber`: Numeric values (integers, floats)
- `FieldTypeBoolean`: True/false values
- `FieldTypeDate`: Date values (YYYY-MM-DD)
- `FieldTypeDateTime`: Date and time (RFC3339)
- `FieldTypeTime`: Time values (HH:MM:SS)
- `FieldTypeSelect`: Pre-defined options from a list

### Operators

Conditions use operators to compare values. Available operators:

**Comparison Operators:**
- `OpEqual` (`equal`): Exact match
- `OpNotEqual` (`not_equal`): Not equal
- `OpLess` (`less`): Less than
- `OpLessOrEqual` (`less_or_equal`): Less than or equal
- `OpGreater` (`greater`): Greater than
- `OpGreaterOrEqual` (`greater_or_equal`): Greater than or equal

**Range Operators:**
- `OpBetween` (`between`): Within range (inclusive)
- `OpNotBetween` (`not_between`): Outside range

**String Operators:**
- `OpContains` (`contains`): Substring match
- `OpNotContains` (`not_contains`): Substring not present
- `OpStartsWith` (`starts_with`): Starts with prefix
- `OpEndsWith` (`ends_with`): Ends with suffix

**Membership Operators:**
- `OpIn` (`select_any_in`): In list
- `OpNotIn` (`select_not_any_in`): Not in list

**Null Operators:**
- `OpIsEmpty` (`is_empty`): Value is null/empty
- `OpIsNotEmpty` (`is_not_empty`): Value is not null/empty

**Pattern Operators:**
- `OpMatchRegexp` (`match_regexp`): Regex pattern match

### Value Types

Values in expressions can be provided as:

- `ValueTypeValue`: Literal value (number, string, boolean, etc.)
- `ValueTypeField`: Reference to a field in the evaluation context
- `ValueTypeFunc`: Result of a custom function call

### Conjunctions

Conditions are combined using logical operators:

- `ConjunctionAnd`: All conditions must be true
- `ConjunctionOr`: At least one condition must be true

Groups can be negated with the `Not` flag.

## Usage Patterns

### Pattern 1: Programmatic Rule Building

```go
builder := condition.NewBuilder(condition.ConjunctionAnd)
builder.
	AddRule("status", condition.OpEqual, "active").
	AddRule("age", condition.OpGreaterOrEqual, 18).
	AddRule("country", condition.OpIn, "US", "CA", "MX")

group := builder.Build()
```

### Pattern 2: Field Comparison

Compare two fields against each other:

```go
builder := condition.NewBuilder(condition.ConjunctionAnd)
builder.AddFieldComparison("salary", "budget", condition.OpLessOrEqual)
group := builder.Build()
```

### Pattern 3: Range Checks

```go
builder := condition.NewBuilder(condition.ConjunctionAnd)
builder.AddBetweenRule("age", 18, 65)
group := builder.Build()
```

### Pattern 4: Complex Nested Logic

```go
// Inner group: premium OR vip
innerBuilder := condition.NewBuilder(condition.ConjunctionOr)
innerBuilder.
	AddRule("premium", condition.OpEqual, true).
	AddRule("vip", condition.OpEqual, true)

// Outer group: active AND (premium OR vip)
outerBuilder := condition.NewBuilder(condition.ConjunctionAnd)
outerBuilder.
	AddRule("status", condition.OpEqual, "active").
	AddGroup(innerBuilder.Build())

group := outerBuilder.Build()
```

### Pattern 5: Formula-Based Rules

For complex business logic, use expr-lang formulas:

```go
builder := condition.NewBuilder(condition.ConjunctionAnd)
builder.AddFormula(`
	(age >= 18 && age <= 65) &&
	(status == "active" || status == "pending") &&
	(discount > 0 || vip)
`)
group := builder.Build()
```

Formula expressions have access to all fields in the evaluation context and support standard operators, functions, and variables.

### Pattern 6: Custom Functions

```go
// Register custom function
ctx.RegisterFunction("isWeekend", func(ctx context.Context, args []any, evalCtx *EvalContext) (any, error) {
	if len(args) == 0 {
		return false, errors.New("date required")
	}
	
	dateStr, ok := args[0].(string)
	if !ok {
		return false, errors.New("invalid date type")
	}
	
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false, err
	}
	
	day := date.Weekday()
	return day == time.Saturday || day == time.Sunday, nil
})

// Use in formula
builder.AddFormula("isWeekend(bookingDate)")
```

## Configuration

### EvalOptions

Configure evaluation behavior:

```go
opts := condition.EvalOptions{
	MaxDepth:      10,                        // Max nesting depth
	MaxConditions: 1000,                      // Max total conditions
	Timeout:       30 * time.Second,          // Evaluation timeout
	CacheResults:  true,                      // Cache compiled formulas
	RegexTimeout:  100 * time.Millisecond,    // Regex match timeout
}

evaluator := condition.NewEvaluator(config, opts)
```

**Defaults:**
- `MaxDepth`: 10 levels (prevents stack overflow)
- `MaxConditions`: 1000 (prevents resource exhaustion)
- `Timeout`: 30 seconds
- `CacheResults`: true
- `RegexTimeout`: 100 milliseconds

### Config Structure

```go
config := &condition.Config{
	Fields: []any{
		condition.Field{
			Type:  condition.FieldTypeNumber,
			Label: "User Age",
			Name:  "age",
		},
		// ... more fields
	},
	Funcs: []any{
		// Custom function definitions
	},
	MaxLevel: 10,
}
```

## Evaluation Context

The `EvalContext` holds runtime data and state for condition evaluation:

```go
data := map[string]any{
	"user": map[string]any{
		"id":    "user-123",
		"age":   25,
		"email": "user@example.com",
	},
	"status": "active",
}

ctx := condition.NewEvalContext(data, condition.DefaultEvalOptions())

// Nested field access using dot notation
value, err := ctx.GetValue("user.email") // "user@example.com"
```

### Thread Safety

`EvalContext` is thread-safe for concurrent reads. However, data modifications and function registration must be externally synchronized:

```go
// Safe: concurrent evaluations
go evaluator.Evaluate(bgCtx, group, ctx)
go evaluator.Evaluate(bgCtx, group, ctx)

// Unsafe: modifying data during evaluation
go ctx.Data["key"] = "value"  // Race condition!

// Safe: modify before evaluation starts
ctx.Data["key"] = "value"
// Now safe to evaluate concurrently
go evaluator.Evaluate(bgCtx, group, ctx)
```

### Metrics

Each evaluation produces metrics:

```go
metrics := ctx.GetMetrics()
// metrics.RulesEvaluated  - Number of rules evaluated
// metrics.GroupsEvaluated - Number of groups evaluated
// metrics.Duration        - Total evaluation time
// metrics.CacheHits       - Formula cache hits
// metrics.CacheMisses     - Formula cache misses
// metrics.Errors          - Number of errors encountered
```

## Security Considerations

### Resource Limits

The engine enforces strict resource limits to prevent abuse:

1. **Max Depth**: Prevents stack overflow from deeply nested conditions
2. **Max Conditions**: Prevents condition explosion and resource exhaustion
3. **Max Formula Length**: Prevents compilation of overly complex expressions
4. **Regex Timeout**: Prevents ReDoS (Regular Expression Denial of Service) attacks
5. **Evaluation Timeout**: Prevents long-running evaluations from blocking

### Timeout Configuration

```go
// Short timeout for user-facing operations
opts := condition.EvalOptions{
	Timeout: 100 * time.Millisecond,
}

// Longer timeout for background jobs
opts := condition.EvalOptions{
	Timeout: 5 * time.Second,
}
```

### Input Validation

All structures validate inputs:

```go
rule := &condition.ConditionRule{
	ID: "rule-1",
	Left: condition.Expression{
		Type:  condition.ValueTypeField,
		Field: "age",
	},
	Op: condition.OpGreaterOrEqual,
	Right: 18,
}

if err := rule.Validate(); err != nil {
	log.Fatal(err)
}
```

## Error Handling

The package defines specific error types for different failure modes:

```go
if err != nil {
	switch err {
	case condition.ErrFieldNotFound:
		// Handle missing field
	case condition.ErrFunctionNotFound:
		// Handle missing function
	case condition.ErrMaxDepthExceeded:
		// Condition too deeply nested
	case condition.ErrResourceLimitExceeded:
		// Too many conditions
	case condition.ErrEvaluationTimeout:
		// Evaluation took too long
	case condition.ErrRegexTimeout:
		// Regex match timed out
	case condition.ErrValidation:
		// Invalid structure
	default:
		// Other errors
	}
}
```

## Advanced Features

### Custom Function Handlers

Functions are powerful for domain-specific logic:

```go
type UserService struct {
	db *database.Connection
}

func (us *UserService) HasPermission(ctx context.Context, args []any, evalCtx *EvalContext) (any, error) {
	if len(args) < 2 {
		return false, errors.New("requires userId and permission")
	}
	
	userID, ok := args[0].(string)
	if !ok {
		return false, errors.New("userId must be string")
	}
	
	permission, ok := args[1].(string)
	if !ok {
		return false, errors.New("permission must be string")
	}
	
	// Query database with context awareness
	return us.db.CheckPermission(ctx, userID, permission)
}

// Register
evalCtx.RegisterFunction("hasPermission", us.HasPermission)

// Use in formula
builder.AddFormula("hasPermission(userId, 'admin')")
```

### Performance Optimization

The engine automatically caches:

1. **Compiled Regex Patterns**: LRU cache with automatic eviction
2. **Compiled Formulas**: Cache compiled expr-lang programs

```go
// First call compiles and caches
result1, _ := evaluator.Evaluate(ctx1, group, evalCtx1)

// Second call reuses cached formula (if identical)
result2, _ := evaluator.Evaluate(ctx2, group, evalCtx2)
```

### Observability

Track performance and behavior:

```go
result, err := evaluator.Evaluate(context.Background(), group, ctx)

metrics := ctx.GetMetrics()
log.Printf("Evaluated %d rules in %v",
	metrics.RulesEvaluated,
	metrics.Duration,
)

if metrics.CacheMisses > 0 {
	hitRate := float64(metrics.CacheHits) / float64(metrics.CacheMisses + metrics.CacheHits)
	log.Printf("Cache hit rate: %.2f%%", hitRate*100)
}
```

## Data Type Conversions

The engine handles automatic type conversions:

### Numeric Comparison

All numeric types are converted to float64:

```go
// These all work
condition.Compare(int(5), float64(5.0))      // 0
condition.Compare(uint32(10), int64(10))     // 0
condition.Compare(float32(3.14), float64(3.14)) // 0
```

### Time Parsing

Dates and times support multiple formats:

```go
// All automatically parsed
"2024-01-15"                     // Date only
"2024-01-15T10:30:45Z"           // RFC3339
"2024-01-15T10:30:45.999999Z"    // RFC3339 with nanoseconds
"2024-01-15 10:30:45"            // Common database format
"10:30:45"                       // Time only
```

### String Fallback

When types don't match, values are converted to strings:

```go
ctx.Data = map[string]any{"value": 42}

// "42" == "42" (both converted to strings)
// Note: "10" < "2" as strings!
```

## Common Patterns

### Feature Flags

```go
group := condition.NewBuilder(condition.ConjunctionAnd).
	AddRule("featureFlag", condition.OpEqual, true).
	AddRule("betaUser", condition.OpEqual, true).
	Build()

enabled, _ := evaluator.Evaluate(ctx, group, evalCtx)
```

### User Segmentation

```go
group := condition.NewBuilder(condition.ConjunctionAnd).
	AddRule("country", condition.OpIn, "US", "CA", "GB").
	AddBetweenRule("age", 18, 65).
	AddRule("signupDate", condition.OpLess, "2024-01-01").
	Build()
```

### ABAC (Attribute-Based Access Control)

```go
group := condition.NewBuilder(condition.ConjunctionAnd).
	AddRule("department", condition.OpEqual, "engineering").
	AddFormula("role == 'lead' || (role == 'engineer' && yearsExp >= 3)").
	Build()
```

### Time-Based Rules

```go
group := condition.NewBuilder(condition.ConjunctionAnd).
	AddRule("currentHour", condition.OpGreaterOrEqual, 9).
	AddRule("currentHour", condition.OpLess, 17).
	Build()
```

## Best Practices

1. **Validate on input**: Call `Validate()` on rules and groups before evaluation
2. **Set appropriate timeouts**: Use shorter timeouts for user-facing operations
3. **Cache evaluators**: Reuse `Evaluator` instances across evaluations
4. **Monitor metrics**: Track evaluation performance and cache effectiveness
5. **Use formulas for complex logic**: Easier to read than deeply nested groups
6. **Test custom functions**: Ensure handlers work with context cancellation
7. **Handle errors gracefully**: Don't crash on validation or evaluation errors
8. **Document field mappings**: Maintain clear documentation of available fields
9. **Limit formula complexity**: Break very complex logic into multiple evaluations
10. **Consider performance**: Monitor regex patterns and formula length

## Troubleshooting

### "Maximum nesting depth exceeded"

The condition is too deeply nested. Either:
- Flatten the structure using `OR` at higher levels
- Increase `MaxDepth` if it's a legitimate use case
- Refactor using formulas

### "Resource limit exceeded"

Too many conditions are being evaluated. Either:
- Simplify the condition structure
- Increase `MaxConditions`
- Use short-circuit evaluation with `AND` where possible

### "Regex timeout"

A regex pattern is taking too long. Either:
- Simplify the pattern
- Increase `RegexTimeout`
- Check for ReDoS vulnerabilities in the pattern

### "Field not found"

The evaluation context doesn't contain expected data. Either:
- Verify field names (case-sensitive)
- Check nested path syntax (use dot notation: "user.email")
- Ensure data is populated before evaluation

## Constants Reference

```go
const (
	// Resource Limits
	DefaultMaxDepth       = 10
	DefaultMaxConditions  = 1000
	DefaultTimeout        = 30 * time.Second
	DefaultRegexTimeout   = 100 * time.Millisecond
	MaxRegexCacheSize     = 1000
	MaxRegexPatternLength = 1000
	MaxFormulaLength      = 10000
	MaxLRUEvictionBatch   = 100
)
```

## Contributing

When extending the package:

1. Add new operators to `OperatorType` enum
2. Implement handler in `applyOperator()` or `applyUnaryOperator()`
3. Add validation in `isValidOperator()`
4. Add tests for edge cases
5. Update documentation

## License

[Your License Here]
