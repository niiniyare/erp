package abac

import (
	"context"
	"fmt"
	"sync"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// AdvancedRuleEngine implements the RuleEngine interface using the 'expr' library.
type AdvancedRuleEngine struct {
	programCache map[string]*vm.Program
	mutex        sync.RWMutex
}

// NewAdvancedRuleEngine creates a new instance of the rule engine.
func NewAdvancedRuleEngine() *AdvancedRuleEngine {
	return &AdvancedRuleEngine{
		programCache: make(map[string]*vm.Program),
	}
}

// EvaluateExpression evaluates a given expression string against a context of attributes.
func (are *AdvancedRuleEngine) EvaluateExpression(
	ctx context.Context,
	expression string,
	attributeCtx *EvaluationAttributeContext,
) (*ExpressionEvaluationResult, error) {
	// Check cache for a compiled program
	are.mutex.RLock()
	program, exists := are.programCache[expression]
	are.mutex.RUnlock()

	var err error
	if !exists {
		// Create the environment map for 'expr' from our context.
		// This allows expressions like "subject.roles.contains('admin')"
		env := are.createEnv(attributeCtx)

		// Compile the expression.
		program, err = expr.Compile(expression, expr.Env(env), expr.AsBool())
		if err != nil {
			return nil, fmt.Errorf("failed to compile expression '%s': %w", expression, err)
		}

		// Cache the compiled program.
		are.mutex.Lock()
		are.programCache[expression] = program
		are.mutex.Unlock()
	}

	// Create the environment again for running the program.
	// This is necessary because the values might be different for each evaluation.
	env := are.createEnv(attributeCtx)

	// Run the compiled program.
	result, err := expr.Run(program, env)
	if err != nil {
		return nil, fmt.Errorf("failed to run expression '%s': %w", expression, err)
	}

	boolResult, ok := result.(bool)
	if !ok {
		return nil, fmt.Errorf("expression '%s' did not return a boolean result, got %T", expression, result)
	}

	return &ExpressionEvaluationResult{
		Result:      boolResult,
		Explanation: "Expression evaluated successfully",
	}, nil
}

// createEnv converts our EvaluationAttributeContext into a map suitable for the expr library.
func (are *AdvancedRuleEngine) createEnv(attributeCtx *EvaluationAttributeContext) map[string]any {
	return map[string]any{
		"subject":     attributeCtx.Subject,
		"resource":    attributeCtx.Resource,
		"action":      attributeCtx.Action,
		"environment": attributeCtx.Environment,
		// You can also add custom functions here if needed in the future
	}
}

// // ValidationResult represents the result of a validation check.
// // This might be used by other services, so we keep the definition.
// type ValidationResult struct {
// 	Valid  bool     `json:"valid"`
// 	Errors []string `json:"errors,omitempty"`
// }
