package ast

import (
	"errors"
	"fmt"
	"net/http"

	sharedErrors "awo.so/internal/shared/errors"
)

// CompileTree validates a Node tree and emits the root AMIS schema map.
//
// Execution contract (enforced here, not by callers):
//  1. Validate() is called on root and every reachable child (depth-first).
//  2. All validation errors are collected before any Compile() call.
//  3. If any Validate() returned an error, CompileTree returns a
//     *sharedErrors.BusinessError with code AST_COMPILE_FAILED and all
//     child errors joined in the cause chain.
//  4. Only when the tree is fully valid does CompileTree call root.Compile().
//
// Callers (CompileStage) must never call node.Compile() directly — always
// go through CompileTree to preserve the validate-before-emit guarantee.
func CompileTree(root Node) (map[string]any, error) {
	var errs []error
	collectValidationErrors(root, &errs)

	if len(errs) > 0 {
		joined := errors.Join(errs...)
		return nil, sharedErrors.NewBusinessError(
			CodeASTCompileFailed,
			fmt.Sprintf("AST validation failed with %d error(s)", len(errs)),
		).
			WithHTTPStatus(http.StatusInternalServerError).
			WithCategory(sharedErrors.CategorySystem).
			WithDetail("error_count", len(errs)).
			WithCause(joined).
			WithSuggestion("Fix all AST node validation errors before compiling")
	}

	return root.Compile(), nil
}

// collectValidationErrors performs a depth-first traversal of the node tree,
// calling Validate() on every node and appending errors to errs.
// It always descends into children even when the parent is invalid, so that
// all violations are reported in a single CompileTree call.
func collectValidationErrors(n Node, errs *[]error) {
	if err := n.Validate(); err != nil {
		*errs = append(*errs, err)
	}

	// Recurse into children if the node is a container.
	if container, ok := n.(ContainerNode); ok {
		for i, child := range container.Children() {
			if child == nil {
				*errs = append(*errs, ErrInvalidField(
					n.NodeType(),
					fmt.Sprintf("children[%d]", i),
					"child node must not be nil",
				))
				continue
			}
			collectValidationErrors(child, errs)
		}
	}
}
