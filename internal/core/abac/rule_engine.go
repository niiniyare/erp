package abac

import (
	"context"
)

// ExpressionEvaluationResult represents the result of an expression evaluation.
type ExpressionEvaluationResult struct {
	Result      bool   `json:"result"`
	Explanation string `json:"explanation,omitempty"`
}

// RuleEngine defines the interface for the expression evaluation component.
type RuleEngine interface {
	// EvaluateExpression evaluates a given expression string against a context of attributes.
	EvaluateExpression(
		ctx context.Context,
		expression string,
		attributeCtx *EvaluationAttributeContext,
	) (*ExpressionEvaluationResult, error)
}
