package ast

import (
	"fmt"
	"net/http"

	sharedErrors "awo.so/internal/shared/errors"
)

// AST error code constants. All AST-layer errors use the "AST_" prefix.
// These codes are used by CompileTree and by concrete node Validate() methods.
const (
	// Structural invariant violations
	CodeASTNodeTypeEmpty       = "AST_NODE_TYPE_EMPTY"
	CodeASTRequiredField       = "AST_REQUIRED_FIELD"
	CodeASTInvalidField        = "AST_INVALID_FIELD"
	CodeASTSyncLocationMissing = "AST_SYNC_LOCATION_MISSING"
	CodeASTAPIMethodInvalid    = "AST_API_METHOD_INVALID"
	CodeASTAPIURLEmpty         = "AST_API_URL_EMPTY"
	CodeASTChildValidation     = "AST_CHILD_VALIDATION"
	CodeASTCompileFailed       = "AST_COMPILE_FAILED"
)

// ErrRequiredField returns a BusinessError for a missing required field on a node.
func ErrRequiredField(nodeType, field string) *sharedErrors.BusinessError {
	return sharedErrors.NewBusinessError(
		CodeASTRequiredField,
		fmt.Sprintf("%s: field %q is required", nodeType, field),
	).
		WithHTTPStatus(http.StatusInternalServerError).
		WithCategory(sharedErrors.CategorySystem).
		WithDetail("node_type", nodeType).
		WithDetail("field", field).
		WithSuggestion(fmt.Sprintf("Set the %q field before compiling this node", field))
}

// ErrInvalidField returns a BusinessError for a field that fails validation.
func ErrInvalidField(nodeType, field, reason string) *sharedErrors.BusinessError {
	return sharedErrors.NewBusinessError(
		CodeASTInvalidField,
		fmt.Sprintf("%s: field %q is invalid — %s", nodeType, field, reason),
	).
		WithHTTPStatus(http.StatusInternalServerError).
		WithCategory(sharedErrors.CategorySystem).
		WithDetail("node_type", nodeType).
		WithDetail("field", field).
		WithDetail("reason", reason)
}

// ErrAPIURLEmpty returns a BusinessError when an APISpec has no URL.
func ErrAPIURLEmpty(nodeType string) *sharedErrors.BusinessError {
	return sharedErrors.NewBusinessError(
		CodeASTAPIURLEmpty,
		fmt.Sprintf("%s: APISpec URL is empty", nodeType),
	).
		WithHTTPStatus(http.StatusInternalServerError).
		WithCategory(sharedErrors.CategorySystem).
		WithDetail("node_type", nodeType).
		WithSuggestion("Set APISpec.URL to a non-empty path (e.g. \"/api/v1/resource\")")
}

// ErrAPIMethodInvalid returns a BusinessError for an unrecognised HTTP method.
func ErrAPIMethodInvalid(nodeType, method string) *sharedErrors.BusinessError {
	return sharedErrors.NewBusinessError(
		CodeASTAPIMethodInvalid,
		fmt.Sprintf("%s: APISpec method %q is not valid", nodeType, method),
	).
		WithHTTPStatus(http.StatusInternalServerError).
		WithCategory(sharedErrors.CategorySystem).
		WithDetail("node_type", nodeType).
		WithDetail("method", method).
		WithSuggestion("Use one of: get, post, put, patch, delete")
}

// ErrChildValidation wraps a child node validation error with parent context.
func ErrChildValidation(parentType string, childIndex int, childErr error) *sharedErrors.BusinessError {
	return sharedErrors.NewBusinessError(
		CodeASTChildValidation,
		fmt.Sprintf("%s: child[%d] failed validation: %v", parentType, childIndex, childErr),
	).
		WithHTTPStatus(http.StatusInternalServerError).
		WithCategory(sharedErrors.CategorySystem).
		WithDetail("parent_type", parentType).
		WithDetail("child_index", childIndex).
		WithCause(childErr)
}
