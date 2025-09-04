package featureflag

import (
	"context"
	"fmt"
	"hash/fnv"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/access/conditional"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/patrickmn/go-cache"
)

// Rule types and operators - these need to be defined or imported
// type RuleType string
// type LogicalOperator string
// type ComparisonOperator string
//
// const (
// 	// Rule types
// 	RuleTypeCondition  RuleType = "condition"
// 	RuleTypeLogical    RuleType = "logical"
// 	RuleTypePercentage RuleType = "percentage"
// 	RuleTypeCustom     RuleType = "custom"
//
// 	// Logical operators
// 	OperatorAND LogicalOperator = "AND"
// 	OperatorOR  LogicalOperator = "OR"
// 	OperatorNOT LogicalOperator = "NOT"
//
// 	// Comparison operators
// 	OpEquals       ComparisonOperator = "equals"
// 	OpNotEquals    ComparisonOperator = "not_equals"
// 	OpContains     ComparisonOperator = "contains"
// 	OpStartsWith   ComparisonOperator = "starts_with"
// 	OpEndsWith     ComparisonOperator = "ends_with"
// 	OpGreaterThan  ComparisonOperator = "greater_than"
// 	OpLessThan     ComparisonOperator = "less_than"
// 	OpGreaterEqual ComparisonOperator = "greater_equal"
// 	OpLessEqual    ComparisonOperator = "less_equal"
// 	OpIn           ComparisonOperator = "in"
// 	OpNotIn        ComparisonOperator = "not_in"
// 	OpExists       ComparisonOperator = "exists"
// 	OpNotExists    ComparisonOperator = "not_exists"
// 	OpRegexMatch   ComparisonOperator = "regex_match"
// )

// Required structs that are referenced in the code
// type EvaluationRule struct {
// 	ID          string             `json:"id"`
// 	Type        RuleType           `json:"type"`
// 	Operator    LogicalOperator    `json:"operator,omitempty"`
// 	Condition   *RuleCondition     `json:"condition,omitempty"`
// 	Children    []*EvaluationRule  `json:"children,omitempty"`
// 	Weight      *float64           `json:"weight,omitempty"`
// 	Description string             `json:"description,omitempty"`
// }
//
// type RuleCondition struct {
// 	Field    string             `json:"field"`
// 	Operator ComparisonOperator `json:"operator"`
// 	Value    any        `json:"value,omitempty"`
// 	Values   []any      `json:"values,omitempty"`
// }
//
// type ComplexEvaluationRules struct {
// 	RootRule *EvaluationRule `json:"root_rule"`
// }
//
// type AdvancedEvaluationContext struct {
// 	UserID      uuid.UUID          `json:"user_id"`
// 	UserData    map[string]any     `json:"user_data,omitempty"`
// 	SessionData map[string]any     `json:"session_data,omitempty"`
// 	RequestData map[string]any     `json:"request_data,omitempty"`
// 	CustomData  map[string]any     `json:"custom_data,omitempty"`
// 	DeviceInfo  *DeviceInfo        `json:"device_info,omitempty"`
// 	LocationInfo *LocationInfo     `json:"location_info,omitempty"`
// 	TimeContext *TimeContext      `json:"time_context,omitempty"`
// }

type DeviceInfo struct {
	DeviceType      string `json:"device_type"`
	OperatingSystem string `json:"operating_system"`
	Browser         string `json:"browser"`
	IsManaged       bool   `json:"is_managed"`
	IsCompliant     bool   `json:"is_compliant"`
}

type LocationInfo struct {
	Country   string `json:"country"`
	Region    string `json:"region"`
	City      string `json:"city"`
	IsTrusted bool   `json:"is_trusted"`
}

type TimeContext struct {
	IsWorkingHours bool `json:"is_working_hours"`
	IsWeekend      bool `json:"is_weekend"`
}

// type AdvancedEvaluationResult struct {
// 	FlagID         uuid.UUID `json:"flag_id"`
// 	UserID         uuid.UUID `json:"user_id"`
// 	Enabled        bool      `json:"enabled"`
// 	Reason         string    `json:"reason"`
// 	EvaluationPath []string  `json:"evaluation_path,omitempty"`
// 	EvaluatedAt    time.Time `json:"evaluated_at"`
// }
//
// // Service interface placeholder
// type Service interface {
// 	// Add methods as needed
// }
//
// // advancedEvaluationEngine placeholder - this should be your actual implementation
// type advancedEvaluationEngine struct {
// 	conditionalAccessService conditional.ConditionalAccessService
// 	simpleService            Service
// 	logger                   logger.Logger
// 	metrics                  metrics.MetricsProvider
// 	tracing                  tracing.TracingService
// }

// evaluateRule placeholder - implement this method based on your actual logic
// func (e *advancedEvaluationEngine) evaluateRule(ctx context.Context, rule *EvaluationRule, context *AdvancedEvaluationContext) (bool, []string) {
// 	// Placeholder implementation
// 	return true, []string{"placeholder_evaluation"}
// }

// ExpressionEngine represents the interface for expression evaluation engines
type ExpressionEngine interface {
	// Evaluate a single expression with context
	EvaluateExpression(ctx context.Context, expression string, context map[string]any) (any, error)

	// Evaluate boolean expression (for conditional logic)
	EvaluateBooleanExpression(ctx context.Context, expression string, context map[string]any) (bool, error)

	// Get engine type for logging/metrics
	GetEngineType() string

	// Validate expression syntax
	ValidateExpression(expression string, context map[string]any) error

	// Get available functions and variables
	GetCapabilities() EngineCapabilities
}

// EngineCapabilities describes what the engine supports
type EngineCapabilities struct {
	SupportedFunctions []string        `json:"supported_functions"`
	SupportedTypes     []string        `json:"supported_types"`
	MaxComplexity      int             `json:"max_complexity"`
	Features           map[string]bool `json:"features"`
}

// EngineType represents the type of expression engine
type EngineType string

const (
	EngineTypeCEL  EngineType = "cel"
	EngineTypeExpr EngineType = "expr"
)

// EngineConfig holds configuration for expression engines
type EngineConfig struct {
	Type            EngineType     `json:"type"`
	CacheTTL        time.Duration  `json:"cache_ttl"`
	MaxCacheSize    int            `json:"max_cache_size"`
	EnableMetrics   bool           `json:"enable_metrics"`
	CustomFunctions map[string]any `json:"custom_functions,omitempty"`
}

// DefaultEngineConfigs provides sensible defaults
var DefaultEngineConfigs = map[EngineType]EngineConfig{
	EngineTypeCEL: {
		Type:          EngineTypeCEL,
		CacheTTL:      15 * time.Minute,
		MaxCacheSize:  1000,
		EnableMetrics: true,
	},
	EngineTypeExpr: {
		Type:          EngineTypeExpr,
		CacheTTL:      15 * time.Minute,
		MaxCacheSize:  1000,
		EnableMetrics: true,
	},
}

// celExpressionEngine implements ExpressionEngine using CEL-Go
type celExpressionEngine struct {
	env           *cel.Env
	programCache  *cache.Cache
	compiledCache sync.Map
	logger        *slog.Logger
	config        EngineConfig
	mutex         sync.RWMutex
}

// NewCELExpressionEngine creates a new CEL-based expression engine
func NewCELExpressionEngine(logger *slog.Logger, config ...EngineConfig) (ExpressionEngine, error) {
	cfg := DefaultEngineConfigs[EngineTypeCEL]
	if len(config) > 0 {
		cfg = config[0]
	}

	// Create CEL environment with comprehensive type system
	env, err := cel.NewEnv(
		// Core variable declarations for feature flag evaluation
		cel.Variable("user", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("session", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("request", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("device", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("location", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("network", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("time", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("risk", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("custom", cel.MapType(cel.StringType, cel.AnyType)),

		// Built-in context variables
		cel.Variable("now", cel.TimestampType),
		cel.Variable("current_hour", cel.IntType),
		cel.Variable("current_day", cel.StringType),
		cel.Variable("is_weekend", cel.BoolType),
		cel.Variable("is_business_hours", cel.BoolType),
		cel.Variable("user_id", cel.StringType),
		cel.Variable("tenant_id", cel.StringType),

		// Rule evaluation context
		cel.Variable("rules", cel.MapType(cel.StringType, cel.BoolType)),

		// Custom functions for feature flag evaluation
		cel.Function("inList",
			cel.Overload("in_list_string", []*cel.Type{cel.StringType, cel.ListType(cel.StringType)}, cel.BoolType),
			cel.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				// Convert to native Go types for easier handling
				lhsValue := lhs.Value()
				rhsValue := rhs.Value()

				// Handle list conversion
				if rhsList, ok := rhsValue.([]any); ok {
					for _, item := range rhsList {
						if lhsValue == item {
							return types.True
						}
					}
				} else if rhsList, ok := rhsValue.([]string); ok {
					lhsStr, ok := lhsValue.(string)
					if !ok {
						return types.False
					}
					for _, item := range rhsList {
						if lhsStr == item {
							return types.True
						}
					}
				}
				return types.False
			})),

		cel.Function("between",
			cel.Overload("between_int", []*cel.Type{cel.IntType, cel.IntType, cel.IntType}, cel.BoolType),
			cel.SingletonFunctionBinding(func(args ...ref.Val) ref.Val {
				if len(args) != 3 {
					return types.False
				}

				// Convert to native Go types
				val := args[0].Value()
				min := args[1].Value()
				max := args[2].Value()

				// Type check and convert
				valInt, ok1 := val.(int64)
				minInt, ok2 := min.(int64)
				maxInt, ok3 := max.(int64)

				if !ok1 || !ok2 || !ok3 {
					return types.False
				}

				return types.Bool(valInt >= minInt && valInt <= maxInt)
			})),

		cel.Function("hasRole",
			cel.Overload("has_role", []*cel.Type{cel.StringType}, cel.BoolType),
			cel.SingletonUnaryBinding(func(val ref.Val) ref.Val {
				// This would integrate with your role system
				return types.Bool(true) // Placeholder
			})),

		cel.Function("percentage",
			cel.Overload("percentage_user", []*cel.Type{cel.StringType, cel.IntType}, cel.BoolType),
			cel.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				// Convert to native Go types
				userIDValue := lhs.Value()
				percentValue := rhs.Value()

				userID, ok1 := userIDValue.(string)
				percent, ok2 := percentValue.(int64)

				if !ok1 || !ok2 {
					return types.False
				}

				h := fnv.New32a()
				h.Write([]byte(userID))
				hashVal := h.Sum32()
				return types.Bool(int64(hashVal%100) < percent)
			})),

		cel.Function("riskScore",
			cel.Overload("risk_score", []*cel.Type{cel.MapType(cel.StringType, cel.AnyType)}, cel.IntType),
			cel.SingletonUnaryBinding(func(val ref.Val) ref.Val {
				// Calculate risk score from context
				return types.Int(25) // Placeholder
			})),

		cel.Function("geoDistance",
			cel.Overload("geo_distance", []*cel.Type{cel.DoubleType, cel.DoubleType, cel.DoubleType, cel.DoubleType}, cel.DoubleType),
			cel.SingletonFunctionBinding(func(args ...ref.Val) ref.Val {
				if len(args) != 4 {
					return types.Double(0)
				}

				// Convert to native Go types
				lat1Value := args[0].Value()
				lon1Value := args[1].Value()
				lat2Value := args[2].Value()
				lon2Value := args[3].Value()

				lat1, ok1 := lat1Value.(float64)
				lon1, ok2 := lon1Value.(float64)
				lat2, ok3 := lat2Value.(float64)
				lon2, ok4 := lon2Value.(float64)

				if !ok1 || !ok2 || !ok3 || !ok4 {
					return types.Double(0)
				}

				// Simplified distance calculation
				dist := ((lat2-lat1)*(lat2-lat1) + (lon2-lon1)*(lon2-lon1))
				return types.Double(dist)
			})),

		// Add contains function for string operations
		cel.Function("contains",
			cel.Overload("contains_string", []*cel.Type{cel.StringType, cel.StringType}, cel.BoolType),
			cel.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				strValue := lhs.Value()
				substrValue := rhs.Value()

				str, ok1 := strValue.(string)
				substr, ok2 := substrValue.(string)

				if !ok1 || !ok2 {
					return types.False
				}
				return types.Bool(strings.Contains(str, substr))
			})),

		// Add startsWith function
		cel.Function("startsWith",
			cel.Overload("starts_with_string", []*cel.Type{cel.StringType, cel.StringType}, cel.BoolType),
			cel.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				strValue := lhs.Value()
				prefixValue := rhs.Value()

				str, ok1 := strValue.(string)
				prefix, ok2 := prefixValue.(string)

				if !ok1 || !ok2 {
					return types.False
				}
				return types.Bool(strings.HasPrefix(str, prefix))
			})),

		// Add endsWith function
		cel.Function("endsWith",
			cel.Overload("ends_with_string", []*cel.Type{cel.StringType, cel.StringType}, cel.BoolType),
			cel.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				strValue := lhs.Value()
				suffixValue := rhs.Value()

				str, ok1 := strValue.(string)
				suffix, ok2 := suffixValue.(string)

				if !ok1 || !ok2 {
					return types.False
				}
				return types.Bool(strings.HasSuffix(str, suffix))
			})),

		// Add has function to check if field exists
		cel.Function("has",
			cel.Overload("has_field", []*cel.Type{cel.StringType}, cel.BoolType),
			cel.SingletonUnaryBinding(func(val ref.Val) ref.Val {
				// This is a placeholder - in practice you'd check the context
				return types.Bool(true)
			})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}

	return &celExpressionEngine{
		env:          env,
		programCache: cache.New(cfg.CacheTTL, cfg.CacheTTL*2),
		logger:       logger,
		config:       cfg,
	}, nil
}

func (e *celExpressionEngine) EvaluateExpression(ctx context.Context, expression string, context map[string]any) (any, error) {
	cacheKey := fmt.Sprintf("cel_%s", expression)

	var program cel.Program
	if cached, found := e.compiledCache.Load(cacheKey); found {
		program = cached.(cel.Program)
	} else {
		ast, issues := e.env.Compile(expression)
		if issues != nil && issues.Err() != nil {
			return nil, fmt.Errorf("CEL compilation failed: %w", issues.Err())
		}

		prog, err := e.env.Program(ast)
		if err != nil {
			return nil, fmt.Errorf("CEL program creation failed: %w", err)
		}

		program = prog
		e.compiledCache.Store(cacheKey, program)
	}

	// Evaluate expression
	out, _, err := program.Eval(context)
	if err != nil {
		return nil, fmt.Errorf("CEL evaluation failed: %w", err)
	}

	return out.Value(), nil
}

func (e *celExpressionEngine) EvaluateBooleanExpression(ctx context.Context, expression string, context map[string]any) (bool, error) {
	result, err := e.EvaluateExpression(ctx, expression, context)
	if err != nil {
		return false, err
	}

	if boolResult, ok := result.(bool); ok {
		return boolResult, nil
	}

	return false, fmt.Errorf("expression did not return boolean value: %T", result)
}

func (e *celExpressionEngine) GetEngineType() string {
	return string(EngineTypeCEL)
}

func (e *celExpressionEngine) ValidateExpression(expression string, context map[string]any) error {
	ast, issues := e.env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return issues.Err()
	}

	_, err := e.env.Program(ast)
	return err
}

func (e *celExpressionEngine) GetCapabilities() EngineCapabilities {
	return EngineCapabilities{
		SupportedFunctions: []string{"inList", "between", "hasRole", "percentage", "riskScore", "geoDistance", "contains", "startsWith", "endsWith", "has"},
		SupportedTypes:     []string{"bool", "int", "double", "string", "bytes", "list", "map", "timestamp", "duration"},
		MaxComplexity:      1000,
		Features: map[string]bool{
			"type_checking":    true,
			"custom_functions": true,
			"regex":            true,
			"datetime":         true,
			"math":             true,
		},
	}
}

// exprExpressionEngine implements ExpressionEngine using expr-lang
type exprExpressionEngine struct {
	programCache  *cache.Cache
	compiledCache sync.Map
	logger        *slog.Logger
	config        EngineConfig
	mutex         sync.RWMutex
}

// NewExprExpressionEngine creates a new expr-lang based expression engine
func NewExprExpressionEngine(logger *slog.Logger, config ...EngineConfig) ExpressionEngine {
	cfg := DefaultEngineConfigs[EngineTypeExpr]
	if len(config) > 0 {
		cfg = config[0]
	}

	return &exprExpressionEngine{
		programCache: cache.New(cfg.CacheTTL, cfg.CacheTTL*2),
		logger:       logger,
		config:       cfg,
	}
}

func (e *exprExpressionEngine) EvaluateExpression(ctx context.Context, expression string, context map[string]any) (any, error) {
	cacheKey := fmt.Sprintf("expr_%s", expression)

	var program *vm.Program
	if cached, found := e.compiledCache.Load(cacheKey); found {
		program = cached.(*vm.Program)
	} else {
		// Create environment with custom functions
		env := map[string]any{
			"inList": func(value any, list []any) bool {
				for _, item := range list {
					if value == item {
						return true
					}
				}
				return false
			},
			"between": func(value, min, max int) bool {
				return value >= min && value <= max
			},
			"hasRole": func(role string) bool {
				// Integrate with your role system
				return true // Placeholder
			},
			"percentage": func(userID string, percent int) bool {
				h := fnv.New32a()
				h.Write([]byte(userID))
				hashVal := h.Sum32()
				return int(hashVal%100) < percent
			},
			"riskScore": func(context map[string]any) int {
				// Calculate risk score from context
				return 25 // Placeholder
			},
			"geoDistance": func(lat1, lon1, lat2, lon2 float64) float64 {
				// Simplified distance calculation
				return ((lat2-lat1)*(lat2-lat1) + (lon2-lon1)*(lon2-lon1))
			},
			"contains": func(str, substr string) bool {
				return strings.Contains(str, substr)
			},
			"startsWith": func(str, prefix string) bool {
				return strings.HasPrefix(str, prefix)
			},
			"endsWith": func(str, suffix string) bool {
				return strings.HasSuffix(str, suffix)
			},
			"has": func(field string) bool {
				// Check if field exists in context
				_, exists := context[field]
				return exists
			},
		}

		// Merge context with environment
		for k, v := range context {
			env[k] = v
		}

		compiled, err := expr.Compile(expression, expr.Env(env))
		if err != nil {
			return nil, fmt.Errorf("expr compilation failed: %w", err)
		}

		program = compiled
		e.compiledCache.Store(cacheKey, program)
	}

	// Execute expression
	output, err := expr.Run(program, context)
	if err != nil {
		return nil, fmt.Errorf("expr evaluation failed: %w", err)
	}

	return output, nil
}

func (e *exprExpressionEngine) EvaluateBooleanExpression(ctx context.Context, expression string, context map[string]any) (bool, error) {
	result, err := e.EvaluateExpression(ctx, expression, context)
	if err != nil {
		return false, err
	}

	// Convert result to boolean
	switch v := result.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case string:
		return v != "", nil
	default:
		return result != nil, nil
	}
}

func (e *exprExpressionEngine) GetEngineType() string {
	return string(EngineTypeExpr)
}

func (e *exprExpressionEngine) ValidateExpression(expression string, context map[string]any) error {
	_, err := expr.Compile(expression, expr.Env(context))
	return err
}

func (e *exprExpressionEngine) GetCapabilities() EngineCapabilities {
	return EngineCapabilities{
		SupportedFunctions: []string{"inList", "between", "hasRole", "percentage", "riskScore", "geoDistance", "contains", "startsWith", "endsWith", "has"},
		SupportedTypes:     []string{"bool", "int", "float", "string", "array", "map"},
		MaxComplexity:      500,
		Features: map[string]bool{
			"type_checking":    false,
			"custom_functions": true,
			"regex":            true,
			"datetime":         false,
			"math":             true,
		},
	}
}

// EnhancedAdvancedEvaluationEngine extends the original with pluggable expression engines
type EnhancedAdvancedEvaluationEngine struct {
	// Embed original implementation
	*advancedEvaluationEngine

	// Expression engines
	celEngine  ExpressionEngine
	exprEngine ExpressionEngine

	// Default engine preference
	defaultEngine EngineType

	// Engine selection strategy
	engineSelector EngineSelector
}

// EngineSelector determines which engine to use for a given expression
type EngineSelector interface {
	SelectEngine(ctx context.Context, expression string, context map[string]any) EngineType
}

// SimpleEngineSelector always returns the default engine
type SimpleEngineSelector struct {
	defaultEngine EngineType
}

func (s *SimpleEngineSelector) SelectEngine(ctx context.Context, expression string, context map[string]any) EngineType {
	return s.defaultEngine
}

// SmartEngineSelector chooses engine based on expression complexity and features
type SmartEngineSelector struct {
	celThreshold  int
	exprThreshold int
}

func (s *SmartEngineSelector) SelectEngine(ctx context.Context, expression string, context map[string]any) EngineType {
	// Simple heuristics for engine selection
	if strings.Contains(expression, "timestamp") || strings.Contains(expression, "duration") {
		return EngineTypeCEL // CEL better for time operations
	}

	if len(expression) > 200 || strings.Count(expression, "&&") > 5 || strings.Count(expression, "||") > 5 {
		return EngineTypeCEL // CEL better for complex expressions
	}

	return EngineTypeExpr // Default to expr for simple expressions
}

// NewEnhancedAdvancedEvaluationEngine creates an enhanced evaluation engine with both expression engines
func NewEnhancedAdvancedEvaluationEngine(
	conditionalAccessService conditional.ConditionalAccessService,
	simpleService Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracing tracing.TracingService,
	defaultEngine EngineType,
) (*EnhancedAdvancedEvaluationEngine, error) {
	// Create original implementation
	originalEngine := &advancedEvaluationEngine{
		conditionalAccessService: conditionalAccessService,
		simpleService:            simpleService,
		logger:                   logger,
		metrics:                  metrics,
		tracing:                  tracing,
	}

	// Create expression engines
	slogLogger := slog.Default()

	celEngine, err := NewCELExpressionEngine(slogLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL engine: %w", err)
	}

	exprEngine := NewExprExpressionEngine(slogLogger)

	// Create engine selector
	var selector EngineSelector
	switch defaultEngine {
	case EngineTypeCEL:
		selector = &SimpleEngineSelector{defaultEngine: EngineTypeCEL}
	case EngineTypeExpr:
		selector = &SimpleEngineSelector{defaultEngine: EngineTypeExpr}
	default:
		selector = &SmartEngineSelector{celThreshold: 100, exprThreshold: 50}
	}

	return &EnhancedAdvancedEvaluationEngine{
		advancedEvaluationEngine: originalEngine,
		celEngine:                celEngine,
		exprEngine:               exprEngine,
		defaultEngine:            defaultEngine,
		engineSelector:           selector,
	}, nil
}

// GetExpressionEngine returns the appropriate expression engine
func (e *EnhancedAdvancedEvaluationEngine) GetExpressionEngine(engineType EngineType) ExpressionEngine {
	switch engineType {
	case EngineTypeCEL:
		return e.celEngine
	case EngineTypeExpr:
		return e.exprEngine
	default:
		return e.exprEngine
	}
}

// EvaluateExpressionWithEngine evaluates an expression using the specified engine
func (e *EnhancedAdvancedEvaluationEngine) EvaluateExpressionWithEngine(ctx context.Context, expression string, context map[string]any, engineType EngineType) (any, error) {
	engine := e.GetExpressionEngine(engineType)
	return engine.EvaluateExpression(ctx, expression, context)
}

// EvaluateExpressionAuto automatically selects the best engine for the expression
func (e *EnhancedAdvancedEvaluationEngine) EvaluateExpressionAuto(ctx context.Context, expression string, context map[string]any) (any, error) {
	engineType := e.engineSelector.SelectEngine(ctx, expression, context)
	engine := e.GetExpressionEngine(engineType)

	// Record metrics about engine selection
	e.metrics.IncrementCounter("expression_engine_selection", map[string]any{
		"engine_type":       string(engineType),
		"expression_length": len(expression),
	})

	return engine.EvaluateExpression(ctx, expression, context)
}

// Enhanced rule evaluation using pluggable expression engines
func (e *EnhancedAdvancedEvaluationEngine) EvaluateComplexRulesWithEngine(
	ctx context.Context,
	flagID uuid.UUID,
	rules *ComplexEvaluationRules,
	context *AdvancedEvaluationContext,
	engineType EngineType,
) (*AdvancedEvaluationResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "enhancedAdvancedEvaluationEngine.EvaluateComplexRulesWithEngine")
	defer span.End()

	if rules == nil || rules.RootRule == nil {
		return &AdvancedEvaluationResult{
			FlagID:      flagID,
			UserID:      context.UserID,
			Enabled:     false,
			Reason:      "No evaluation rules provided",
			EvaluatedAt: time.Now(),
		}, nil
	}

	// Build expression context
	expressionContext := e.buildExpressionContext(context)

	// Evaluate rules using the specified engine
	result, evaluationPath := e.evaluateRuleWithEngine(ctx, rules.RootRule, expressionContext, engineType)

	// Record engine usage metrics
	e.metrics.IncrementCounter("complex_rule_evaluation", map[string]any{
		"engine_type": string(engineType),
		"rule_count":  e.countRules(rules.RootRule),
		"result":      result,
	})

	return &AdvancedEvaluationResult{
		FlagID:         flagID,
		UserID:         context.UserID,
		Enabled:        result,
		Reason:         strings.Join(evaluationPath, " -> "),
		EvaluationPath: evaluationPath,
		EvaluatedAt:    time.Now(),
	}, nil
}

// Helper methods for enhanced engine

func (e *EnhancedAdvancedEvaluationEngine) buildExpressionContext(context *AdvancedEvaluationContext) map[string]any {
	expressionContext := make(map[string]any)

	// Add structured context data
	expressionContext["user"] = context.UserData
	expressionContext["session"] = context.SessionData
	expressionContext["request"] = context.RequestData
	expressionContext["custom"] = context.CustomData

	// Add user ID
	expressionContext["user_id"] = context.UserID.String()

	// Add device context
	if context.DeviceInfo != nil {
		expressionContext["device"] = map[string]any{
			"type":         context.DeviceInfo.DeviceType,
			"os":           context.DeviceInfo.OperatingSystem,
			"browser":      context.DeviceInfo.Browser,
			"is_managed":   context.DeviceInfo.IsManaged,
			"is_compliant": context.DeviceInfo.IsCompliant,
		}
	}

	// Add location context
	if context.LocationInfo != nil {
		expressionContext["location"] = map[string]any{
			"country":    context.LocationInfo.Country,
			"region":     context.LocationInfo.Region,
			"city":       context.LocationInfo.City,
			"is_trusted": context.LocationInfo.IsTrusted,
		}
	}

	// Add time context
	if context.TimeContext != nil {
		expressionContext["time"] = map[string]any{
			"is_working_hours": context.TimeContext.IsWorkingHours,
			"is_weekend":       context.TimeContext.IsWeekend,
		}
		expressionContext["is_business_hours"] = context.TimeContext.IsWorkingHours
		expressionContext["is_weekend"] = context.TimeContext.IsWeekend
	}

	// Add current time
	now := time.Now()
	expressionContext["now"] = now
	expressionContext["current_hour"] = now.Hour()
	expressionContext["current_day"] = now.Weekday().String()

	return expressionContext
}

func (e *EnhancedAdvancedEvaluationEngine) evaluateRuleWithEngine(
	ctx context.Context,
	rule *EvaluationRule,
	context map[string]any,
	engineType EngineType,
) (bool, []string) {
	engine := e.GetExpressionEngine(engineType)

	switch rule.Type {
	case RuleTypeCondition:
		if rule.Condition != nil {
			// Build expression from condition
			expression := e.buildConditionExpression(rule.Condition)
			result, err := engine.EvaluateBooleanExpression(ctx, expression, context)
			if err != nil {
				e.logger.Error("Expression evaluation failed", logger.Fields{
					"error":      err.Error(),
					"rule_id":    rule.ID,
					"engine":     string(engineType),
					"expression": expression,
				})
				return false, []string{fmt.Sprintf("error:%s", err.Error())}
			}
			return result, []string{fmt.Sprintf("condition:%s=%t", expression, result)}
		}
		return false, []string{"condition_missing"}

	case RuleTypeLogical:
		return e.evaluateLogicalRuleWithEngine(ctx, rule, context, engine)

	case RuleTypePercentage:
		if rule.Weight != nil {
			expression := fmt.Sprintf("percentage(user_id, %d)", int(*rule.Weight*100))
			result, err := engine.EvaluateBooleanExpression(ctx, expression, context)
			if err != nil {
				return false, []string{fmt.Sprintf("percentage_error:%s", err.Error())}
			}
			return result, []string{fmt.Sprintf("percentage:%.2f=%t", *rule.Weight, result)}
		}
		return false, []string{"percentage_weight_missing"}

	case RuleTypeCustom:
		// For custom rules, you can define expressions directly in the rule
		if rule.Description != "" {
			// Use description as custom expression
			result, err := engine.EvaluateBooleanExpression(ctx, rule.Description, context)
			if err != nil {
				return false, []string{fmt.Sprintf("custom_error:%s", err.Error())}
			}
			return result, []string{fmt.Sprintf("custom:%s=%t", rule.Description, result)}
		}
		return false, []string{"custom_expression_missing"}

	default:
		// Fall back to original implementation for other rule types
		return e.evaluateRule(ctx, rule, &AdvancedEvaluationContext{
			UserID:      uuid.MustParse(context["user_id"].(string)),
			UserData:    getMapFromInterface(context["user"]),
			SessionData: getMapFromInterface(context["session"]),
			RequestData: getMapFromInterface(context["request"]),
			CustomData:  getMapFromInterface(context["custom"]),
		})
	}
}

func (e *EnhancedAdvancedEvaluationEngine) evaluateLogicalRuleWithEngine(
	ctx context.Context,
	rule *EvaluationRule,
	context map[string]any,
	engine ExpressionEngine,
) (bool, []string) {
	var expressions []string
	var results []bool
	var paths []string

	for _, child := range rule.Children {
		result, childPaths := e.evaluateRuleWithEngine(ctx, child, context, EngineType(engine.GetEngineType()))
		results = append(results, result)
		paths = append(paths, childPaths...)

		// Build expression parts for logical operations
		if child.Type == RuleTypeCondition && child.Condition != nil {
			expressions = append(expressions, e.buildConditionExpression(child.Condition))
		}
	}

	// Apply logical operator
	var finalResult bool
	var operation string

	switch rule.Operator {
	case OperatorAND:
		finalResult = true
		for _, result := range results {
			finalResult = finalResult && result
		}
		operation = "AND"

	case OperatorOR:
		finalResult = false
		for _, result := range results {
			finalResult = finalResult || result
		}
		operation = "OR"

	case OperatorNOT:
		if len(results) > 0 {
			finalResult = !results[0]
			operation = "NOT"
		}
	}

	return finalResult, append(paths, fmt.Sprintf("%s:%t", operation, finalResult))
}

func (e *EnhancedAdvancedEvaluationEngine) buildConditionExpression(condition *RuleCondition) string {
	field := condition.Field
	operator := condition.Operator
	value := condition.Value

	// Build expression based on operator
	switch operator {
	case OpEquals:
		return fmt.Sprintf(`%s == "%v"`, field, value)
	case OpNotEquals:
		return fmt.Sprintf(`%s != "%v"`, field, value)
	case OpContains:
		return fmt.Sprintf(`contains(%s, "%v")`, field, value)
	case OpStartsWith:
		return fmt.Sprintf(`startsWith(%s, "%v")`, field, value)
	case OpEndsWith:
		return fmt.Sprintf(`endsWith(%s, "%v")`, field, value)
	case OpGreaterThan:
		return fmt.Sprintf(`%s > %v`, field, value)
	case OpLessThan:
		return fmt.Sprintf(`%s < %v`, field, value)
	case OpGreaterEqual:
		return fmt.Sprintf(`%s >= %v`, field, value)
	case OpLessEqual:
		return fmt.Sprintf(`%s <= %v`, field, value)
	case OpIn:
		if condition.Values != nil {
			valueList := make([]string, len(condition.Values))
			for i, v := range condition.Values {
				valueList[i] = fmt.Sprintf(`"%v"`, v)
			}
			return fmt.Sprintf(`inList(%s, [%s])`, field, strings.Join(valueList, ", "))
		}
		return fmt.Sprintf(`%s == "%v"`, field, value)
	case OpNotIn:
		if condition.Values != nil {
			valueList := make([]string, len(condition.Values))
			for i, v := range condition.Values {
				valueList[i] = fmt.Sprintf(`"%v"`, v)
			}
			return fmt.Sprintf(`!inList(%s, [%s])`, field, strings.Join(valueList, ", "))
		}
		return fmt.Sprintf(`%s != "%v"`, field, value)
	case OpExists:
		return fmt.Sprintf(`has(%s)`, field)
	case OpNotExists:
		return fmt.Sprintf(`!has(%s)`, field)
	case OpRegexMatch:
		return fmt.Sprintf(`%s.matches("%v")`, field, value)
	default:
		return fmt.Sprintf(`%s == "%v"`, field, value)
	}
}

func (e *EnhancedAdvancedEvaluationEngine) countRules(rule *EvaluationRule) int {
	count := 1
	for _, child := range rule.Children {
		count += e.countRules(child)
	}
	return count
}

// Utility function to safely convert any to map[string]any
func getMapFromInterface(value any) map[string]any {
	if value == nil {
		return make(map[string]any)
	}
	if m, ok := value.(map[string]any); ok {
		return m
	}
	if m, ok := value.(map[string]any); ok {
		result := make(map[string]any)
		for k, v := range m {
			result[k] = v
		}
		return result
	}
	return make(map[string]any)
}

// Factory function to create evaluation engines with different configurations
type EngineFactory struct {
	logger  *slog.Logger
	metrics metrics.MetricsProvider
}

func NewEngineFactory(logger *slog.Logger, metrics metrics.MetricsProvider) *EngineFactory {
	return &EngineFactory{
		logger:  logger,
		metrics: metrics,
	}
}

func (f *EngineFactory) CreateEngine(engineType EngineType, config ...EngineConfig) (ExpressionEngine, error) {
	switch engineType {
	case EngineTypeCEL:
		return NewCELExpressionEngine(f.logger, config...)
	case EngineTypeExpr:
		return NewExprExpressionEngine(f.logger, config...), nil
	default:
		return nil, fmt.Errorf("unsupported engine type: %s", engineType)
	}
}

// ExpressionEngineComparator for testing and benchmarking both engines
type ExpressionEngineComparator struct {
	celEngine  ExpressionEngine
	exprEngine ExpressionEngine
	logger     *slog.Logger
}

func NewExpressionEngineComparator(logger *slog.Logger) (*ExpressionEngineComparator, error) {
	celEngine, err := NewCELExpressionEngine(logger)
	if err != nil {
		return nil, err
	}

	exprEngine := NewExprExpressionEngine(logger)

	return &ExpressionEngineComparator{
		celEngine:  celEngine,
		exprEngine: exprEngine,
		logger:     logger,
	}, nil
}

// CompareEvaluation runs the same expression on both engines and compares results
func (c *ExpressionEngineComparator) CompareEvaluation(ctx context.Context, expression string, context map[string]any) (*ComparisonResult, error) {
	start := time.Now()

	// Evaluate with CEL
	celStart := time.Now()
	celResult, celErr := c.celEngine.EvaluateExpression(ctx, expression, context)
	celDuration := time.Since(celStart)

	// Evaluate with expr
	exprStart := time.Now()
	exprResult, exprErr := c.exprEngine.EvaluateExpression(ctx, expression, context)
	exprDuration := time.Since(exprStart)

	result := &ComparisonResult{
		Expression:    expression,
		CELResult:     celResult,
		CELError:      celErr,
		CELDuration:   celDuration,
		ExprResult:    exprResult,
		ExprError:     exprErr,
		ExprDuration:  exprDuration,
		ResultsMatch:  compareResults(celResult, exprResult, celErr, exprErr),
		TotalDuration: time.Since(start),
	}

	c.logger.Debug("Expression comparison completed",
		"expression", expression,
		"cel_duration", celDuration,
		"expr_duration", exprDuration,
		"results_match", result.ResultsMatch)

	return result, nil
}

type ComparisonResult struct {
	Expression    string        `json:"expression"`
	CELResult     any           `json:"cel_result"`
	CELError      error         `json:"cel_error,omitempty"`
	CELDuration   time.Duration `json:"cel_duration"`
	ExprResult    any           `json:"expr_result"`
	ExprError     error         `json:"expr_error,omitempty"`
	ExprDuration  time.Duration `json:"expr_duration"`
	ResultsMatch  bool          `json:"results_match"`
	TotalDuration time.Duration `json:"total_duration"`
}

func compareResults(celResult, exprResult any, celErr, exprErr error) bool {
	// If both have errors, consider them matching
	if celErr != nil && exprErr != nil {
		return true
	}

	// If only one has an error, they don't match
	if (celErr == nil) != (exprErr == nil) {
		return false
	}

	// Compare actual results
	return fmt.Sprintf("%v", celResult) == fmt.Sprintf("%v", exprResult)
}

// EngineMetrics provides detailed metrics about engine usage
type EngineMetrics struct {
	EngineType       string        `json:"engine_type"`
	EvaluationsCount int64         `json:"evaluations_count"`
	AverageLatency   time.Duration `json:"average_latency"`
	ErrorRate        float64       `json:"error_rate"`
	CacheHitRate     float64       `json:"cache_hit_rate"`
}

// MetricsCollector collects and aggregates metrics from expression engines
type MetricsCollector struct {
	metrics map[string]*EngineMetrics
	mutex   sync.RWMutex
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*EngineMetrics),
	}
}

func (m *MetricsCollector) RecordEvaluation(engineType string, duration time.Duration, success bool, cacheHit bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.metrics[engineType]; !exists {
		m.metrics[engineType] = &EngineMetrics{
			EngineType: engineType,
		}
	}

	metric := m.metrics[engineType]
	metric.EvaluationsCount++

	// Update average latency (simple moving average)
	if metric.AverageLatency == 0 {
		metric.AverageLatency = duration
	} else {
		metric.AverageLatency = (metric.AverageLatency + duration) / 2
	}

	// Update error rate
	if !success {
		metric.ErrorRate = (metric.ErrorRate + 1.0) / 2.0
	} else {
		metric.ErrorRate = metric.ErrorRate * 0.95 // Decay error rate for successful evaluations
	}

	// Update cache hit rate
	if cacheHit {
		metric.CacheHitRate = (metric.CacheHitRate + 1.0) / 2.0
	} else {
		metric.CacheHitRate = metric.CacheHitRate * 0.95
	}
}

func (m *MetricsCollector) GetMetrics() map[string]*EngineMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*EngineMetrics)
	for k, v := range m.metrics {
		result[k] = &EngineMetrics{
			EngineType:       v.EngineType,
			EvaluationsCount: v.EvaluationsCount,
			AverageLatency:   v.AverageLatency,
			ErrorRate:        v.ErrorRate,
			CacheHitRate:     v.CacheHitRate,
		}
	}
	return result
}

// Example usage and integration helpers

// ExpressionMigrator helps migrate expressions between engines
type ExpressionMigrator struct {
	logger *slog.Logger
}

func NewExpressionMigrator(logger *slog.Logger) *ExpressionMigrator {
	return &ExpressionMigrator{logger: logger}
}

// MigrateToCEL converts expr-lang expressions to CEL format
func (m *ExpressionMigrator) MigrateToCEL(exprExpression string) (string, error) {
	// Basic migration patterns
	celExpression := exprExpression

	// Replace expr-lang specific syntax with CEL equivalents
	migrations := map[string]string{
		" and ":     " && ",
		" or ":      " || ",
		" not ":     " ! ",
		"len(":      "size(",
		"contains(": "contains(",
		"matches(":  ".matches(",
	}

	for exprSyntax, celSyntax := range migrations {
		celExpression = strings.ReplaceAll(celExpression, exprSyntax, celSyntax)
	}

	return celExpression, nil
}

// MigrateToExpr converts CEL expressions to expr-lang format
func (m *ExpressionMigrator) MigrateToExpr(celExpression string) (string, error) {
	// Basic migration patterns
	exprExpression := celExpression

	// Replace CEL specific syntax with expr-lang equivalents
	migrations := map[string]string{
		" && ":      " and ",
		" || ":      " or ",
		" ! ":       " not ",
		"size(":     "len(",
		".matches(": " matches ",
	}

	for celSyntax, exprSyntax := range migrations {
		exprExpression = strings.ReplaceAll(exprExpression, celSyntax, exprSyntax)
	}

	return exprExpression, nil
}

// Configuration helpers

// DefaultCELConfig returns a default configuration for CEL engine
func DefaultCELConfig() EngineConfig {
	return EngineConfig{
		Type:          EngineTypeCEL,
		CacheTTL:      15 * time.Minute,
		MaxCacheSize:  1000,
		EnableMetrics: true,
	}
}

// DefaultExprConfig returns a default configuration for expr-lang engine
func DefaultExprConfig() EngineConfig {
	return EngineConfig{
		Type:          EngineTypeExpr,
		CacheTTL:      15 * time.Minute,
		MaxCacheSize:  1000,
		EnableMetrics: true,
	}
}

// PerformanceConfig returns a configuration optimized for performance
func PerformanceConfig(engineType EngineType) EngineConfig {
	return EngineConfig{
		Type:          engineType,
		CacheTTL:      30 * time.Minute, // Longer cache TTL
		MaxCacheSize:  5000,             // Larger cache
		EnableMetrics: true,
	}
}

// SecurityConfig returns a configuration optimized for security
func SecurityConfig(engineType EngineType) EngineConfig {
	return EngineConfig{
		Type:          engineType,
		CacheTTL:      5 * time.Minute, // Shorter cache TTL
		MaxCacheSize:  500,             // Smaller cache
		EnableMetrics: true,
	}
}

// Example benchmark test helper
type BenchmarkResult struct {
	EngineType    string        `json:"engine_type"`
	Expression    string        `json:"expression"`
	ExecutionTime time.Duration `json:"execution_time"`
	MemoryUsage   int64         `json:"memory_usage"`
	CacheHit      bool          `json:"cache_hit"`
	Success       bool          `json:"success"`
	Error         error         `json:"error,omitempty"`
}

// BenchmarkSuite runs performance tests against both engines
type BenchmarkSuite struct {
	comparator *ExpressionEngineComparator
	logger     *slog.Logger
}

func NewBenchmarkSuite(logger *slog.Logger) (*BenchmarkSuite, error) {
	comparator, err := NewExpressionEngineComparator(logger)
	if err != nil {
		return nil, err
	}

	return &BenchmarkSuite{
		comparator: comparator,
		logger:     logger,
	}, nil
}

func (b *BenchmarkSuite) RunBenchmark(ctx context.Context, expressions []string, context map[string]any, iterations int) ([]*BenchmarkResult, error) {
	var results []*BenchmarkResult

	for _, expression := range expressions {
		// Benchmark CEL
		celResult := b.benchmarkEngine(ctx, b.comparator.celEngine, expression, context, iterations)
		results = append(results, celResult)

		// Benchmark expr
		exprResult := b.benchmarkEngine(ctx, b.comparator.exprEngine, expression, context, iterations)
		results = append(results, exprResult)
	}

	return results, nil
}

func (b *BenchmarkSuite) benchmarkEngine(ctx context.Context, engine ExpressionEngine, expression string, context map[string]any, iterations int) *BenchmarkResult {
	var totalTime time.Duration
	var successCount int
	var lastError error

	// Run multiple iterations
	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, err := engine.EvaluateExpression(ctx, expression, context)
		duration := time.Since(start)

		totalTime += duration
		if err == nil {
			successCount++
		} else {
			lastError = err
		}
	}

	return &BenchmarkResult{
		EngineType:    engine.GetEngineType(),
		Expression:    expression,
		ExecutionTime: totalTime / time.Duration(iterations),
		Success:       successCount == iterations,
		Error:         lastError,
	}
}

// Additional utility functions for debugging and monitoring

// ExpressionDebugger provides debugging capabilities
type ExpressionDebugger struct {
	logger *slog.Logger
}

func NewExpressionDebugger(logger *slog.Logger) *ExpressionDebugger {
	return &ExpressionDebugger{logger: logger}
}

// DebugExpression provides detailed information about expression evaluation
func (d *ExpressionDebugger) DebugExpression(ctx context.Context, engine ExpressionEngine, expression string, context map[string]any) *DebugResult {
	start := time.Now()

	// Validate expression first
	validationErr := engine.ValidateExpression(expression, context)

	// Evaluate expression
	result, evalErr := engine.EvaluateExpression(ctx, expression, context)

	duration := time.Since(start)

	debugResult := &DebugResult{
		Expression:      expression,
		EngineType:      engine.GetEngineType(),
		Context:         context,
		Result:          result,
		ValidationError: validationErr,
		EvaluationError: evalErr,
		ExecutionTime:   duration,
		Capabilities:    engine.GetCapabilities(),
	}

	d.logger.Debug("Expression debug completed",
		"expression", expression,
		"engine", engine.GetEngineType(),
		"valid", validationErr == nil,
		"success", evalErr == nil,
		"duration", duration)

	return debugResult
}

type DebugResult struct {
	Expression      string             `json:"expression"`
	EngineType      string             `json:"engine_type"`
	Context         map[string]any     `json:"context"`
	Result          any                `json:"result"`
	ValidationError error              `json:"validation_error,omitempty"`
	EvaluationError error              `json:"evaluation_error,omitempty"`
	ExecutionTime   time.Duration      `json:"execution_time"`
	Capabilities    EngineCapabilities `json:"capabilities"`
}
