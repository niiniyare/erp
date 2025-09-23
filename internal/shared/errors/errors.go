package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ─── PREDEFINED ERRORS ─────────────────────────────────────────────
// These replace the simple error variables with enhanced BusinessError instances
// while maintaining backward compatibility

var (
	// ─── TENANT ERRORS ─────────────────────────────────────────────

	// ErrTenantExists indicates a tenant already exists
	ErrTenantExists = NewBusinessError("TENANT_EXISTS", "Tenant already exists").
			WithHTTPStatus(http.StatusConflict).
			WithCategory(CategoryTenant).
			WithSuggestion("Choose a different tenant name or slug").
			WithSuggestion("Check if you're trying to create a duplicate tenant")

	// ErrTenantNotFound indicates a tenant was not found
	ErrTenantNotFound = NewBusinessError("TENANT_NOT_FOUND", "Tenant not found").
				WithHTTPStatus(http.StatusNotFound).
				WithCategory(CategoryTenant).
				WithSuggestion("Verify the tenant ID or slug is correct").
				WithSuggestion("Contact support if you believe this tenant should exist")

	// ErrTenantIDNotInContext indicates tenant context is missing
	ErrTenantIDNotInContext = NewBusinessError("TENANT_CONTEXT_MISSING", "Tenant ID not found in request context").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryTenant).
				WithSeverity(SeverityError).
				WithSuggestion("Ensure the request includes proper tenant identification").
				WithSuggestion("Check that tenant middleware is properly configured")

	// ErrSubdomainAlreadyExists indicates subdomain conflict
	ErrSubdomainAlreadyExists = NewBusinessError("SUBDOMAIN_EXISTS", "Subdomain already exists").
					WithHTTPStatus(http.StatusConflict).
					WithCategory(CategoryTenant).
					WithSuggestion("Choose a different subdomain").
					WithSuggestion("Try adding numbers or variations to make it unique")

	// ErrSubdomainSoftDeleted indicates a subdomain is taken by a soft-deleted tenant
	ErrSubdomainSoftDeleted = NewBusinessError("SUBDOMAIN_SOFT_DELETED", "Subdomain is unavailable").
				WithHTTPStatus(http.StatusConflict).
				WithCategory(CategoryTenant).
				WithSuggestion("This subdomain was recently used. Please choose a different one or try again later.").
				WithSuggestion("If you own this subdomain and want to reactivate it, please contact support.")

	// ErrFeatureNotEnabled indicates a feature is not available for the tenant
	ErrFeatureNotEnabled = NewBusinessError("FEATURE_NOT_ENABLED", "Feature not enabled for this tenant").
				WithHTTPStatus(http.StatusForbidden).
				WithCategory(CategoryTenant).
				WithSuggestion("Upgrade your plan to access this feature").
				WithSuggestion("Contact sales for more information about feature availability")

	// ─── USER ERRORS ─────────────────────────────────────────────

	// ErrUserNotFound indicates a user was not found
	ErrUserNotFound = NewBusinessError("USER_NOT_FOUND", "User not found").
			WithHTTPStatus(http.StatusNotFound).
			WithCategory(CategorySecurity).
			WithSuggestion("Verify the user ID or email is correct").
			WithSuggestion("Check if the user exists in your tenant")

	// ErrUserExists indicates a user already exists
	ErrUserExists = NewBusinessError("USER_EXISTS", "User already exists").
			WithHTTPStatus(http.StatusConflict).
			WithCategory(CategorySecurity).
			WithSuggestion("Use a different email address").
			WithSuggestion("Try logging in if you already have an account")

	// ErrEmailAlreadyExists indicates email address is taken
	ErrEmailAlreadyExists = NewBusinessError("EMAIL_EXISTS", "Email address already exists").
				WithHTTPStatus(http.StatusConflict).
				WithCategory(CategoryValidation).
				WithSuggestion("Use a different email address").
				WithSuggestion("Check if you already have an account with this email")

	// ErrUsernameAlreadyExists indicates username is taken
	ErrUsernameAlreadyExists = NewBusinessError("USERNAME_EXISTS", "Username already exists").
					WithHTTPStatus(http.StatusConflict).
					WithCategory(CategoryValidation).
					WithSuggestion("Choose a different username").
					WithSuggestion("Try adding numbers or variations to make it unique")

	// ErrInvalidCredentials indicates authentication failed
	ErrInvalidCredentials = NewBusinessError("INVALID_CREDENTIALS", "Invalid email or password").
				WithHTTPStatus(http.StatusUnauthorized).
				WithCategory(CategorySecurity).
				WithSuggestion("Check your email and password").
				WithSuggestion("Use the 'Forgot Password' option if needed").
				WithSuggestion("Ensure caps lock is not enabled")

	// ErrAuthenticationFailed indicates general authentication failure
	ErrAuthenticationFailed = NewBusinessError("AUTHENTICATION_FAILED", "Authentication failed").
				WithHTTPStatus(http.StatusUnauthorized).
				WithCategory(CategorySecurity).
				WithSuggestion("Verify your credentials and try again").
				WithSuggestion("Contact support if the issue persists")

	// ErrInvalidUserType indicates an invalid user type
	ErrInvalidUserType = NewBusinessError("INVALID_USER_TYPE", "Invalid user type").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryValidation).
				WithSuggestion("Use a valid user type (admin, user, viewer, etc.)").
				WithSuggestion("Check the API documentation for valid user types")

	// ErrInvalidAccountStatus indicates an invalid account status
	ErrInvalidAccountStatus = NewBusinessError("INVALID_ACCOUNT_STATUS", "Invalid account status").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryValidation).
				WithSuggestion("Use a valid status (active, inactive, pending, suspended)").
				WithSuggestion("Check the API documentation for valid status values")

	// ErrAccountLocked indicates account is locked
	ErrAccountLocked = NewBusinessError("ACCOUNT_LOCKED", "Account is locked").
				WithHTTPStatus(http.StatusForbidden).
				WithCategory(CategorySecurity).
				WithSeverity(SeverityWarning).
				WithSuggestion("Contact an administrator to unlock your account").
				WithSuggestion("Wait for the automatic unlock period if applicable")

	// ─── ROLE & PERMISSION ERRORS ─────────────────────────────────────────────

	// ErrRoleNotFound indicates a role was not found
	ErrRoleNotFound = NewBusinessError("ROLE_NOT_FOUND", "Role not found").
			WithHTTPStatus(http.StatusNotFound).
			WithCategory(CategorySecurity).
			WithSuggestion("Verify the role ID or name is correct").
			WithSuggestion("Check if the role exists in your tenant")

	// ErrRoleExists indicates a role already exists
	ErrRoleExists = NewBusinessError("ROLE_EXISTS", "Role already exists").
			WithHTTPStatus(http.StatusConflict).
			WithCategory(CategorySecurity).
			WithSuggestion("Use a different role name").
			WithSuggestion("Check existing roles to avoid duplicates")

	// ErrUnauthorized indicates unauthorized access (commented out in original)
	ErrUnauthorized = NewBusinessError("UNAUTHORIZED", "Unauthorized access").
			WithHTTPStatus(http.StatusUnauthorized).
			WithCategory(CategorySecurity).
			WithSuggestion("Please log in to access this resource").
			WithSuggestion("Verify your authentication token is valid")

	// ErrForbidden indicates forbidden access (commented out in original)
	ErrForbidden = NewBusinessError("FORBIDDEN", "Access forbidden").
			WithHTTPStatus(http.StatusForbidden).
			WithCategory(CategorySecurity).
			WithSuggestion("Contact your administrator for access").
			WithSuggestion("Verify you have the required permissions")

	// ─── INVITATION ERRORS ─────────────────────────────────────────────

	// ErrInvitationNotFound indicates invitation was not found
	ErrInvitationNotFound = NewBusinessError("INVITATION_NOT_FOUND", "Invitation not found").
				WithHTTPStatus(http.StatusNotFound).
				WithCategory(CategoryBusiness).
				WithSuggestion("Verify the invitation link is correct").
				WithSuggestion("Contact the person who sent the invitation")

	// ErrInvitationExpired indicates invitation has expired
	ErrInvitationExpired = NewBusinessError("INVITATION_EXPIRED", "Invitation has expired").
				WithHTTPStatus(http.StatusGone).
				WithCategory(CategoryBusiness).
				WithSuggestion("Request a new invitation").
				WithSuggestion("Contact an administrator for a fresh invitation link")

	// ─── API KEY ERRORS ─────────────────────────────────────────────

	// ErrAPIKeyNotFound indicates API key was not found
	ErrAPIKeyNotFound = NewBusinessError("API_KEY_NOT_FOUND", "API key not found").
				WithHTTPStatus(http.StatusUnauthorized).
				WithCategory(CategorySecurity).
				WithSuggestion("Verify your API key is correct").
				WithSuggestion("Generate a new API key if the current one is invalid")

	// ─── ENTITY ERRORS ─────────────────────────────────────────────

	// ErrEntityNotFound indicates an entity was not found
	ErrEntityNotFound = NewBusinessError("ENTITY_NOT_FOUND", "Entity not found").
				WithHTTPStatus(http.StatusNotFound).
				WithCategory(CategoryBusiness).
				WithSuggestion("Verify the entity ID is correct").
				WithSuggestion("Check if the entity exists in your tenant")

	// ErrEntityNameExists indicates entity name already exists
	ErrEntityNameExists = NewBusinessError("ENTITY_NAME_EXISTS", "Entity name already exists").
				WithHTTPStatus(http.StatusConflict).
				WithCategory(CategoryBusiness).
				WithSuggestion("Choose a different entity name").
				WithSuggestion("Check existing entities to avoid duplicates")

	// ErrEntityCodeExists indicates entity code already exists
	ErrEntityCodeExists = NewBusinessError("ENTITY_CODE_EXISTS", "Entity code already exists").
				WithHTTPStatus(http.StatusConflict).
				WithCategory(CategoryBusiness).
				WithSuggestion("Use a different entity code").
				WithSuggestion("Entity codes must be unique within your organization")

	// ErrInvalidParentEntity indicates invalid parent entity reference
	ErrInvalidParentEntity = NewBusinessError("INVALID_PARENT_ENTITY", "Invalid parent entity").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryBusiness).
				WithSuggestion("Verify the parent entity exists and is valid").
				WithSuggestion("Check that the parent entity allows child entities")

	// ErrCircularReference indicates circular reference in entity hierarchy
	ErrCircularReference = NewBusinessError("CIRCULAR_REFERENCE", "Circular reference detected in entity hierarchy").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryBusiness).
				WithSeverity(SeverityError).
				WithSuggestion("Review the entity hierarchy to remove circular references").
				WithSuggestion("An entity cannot be a parent of itself or its ancestors")

	// ─── ABAC ERRORS ─────────────────────────────────────────────

	// Policy Errors
	ErrPolicyNotFound = NewBusinessError("POLICY_NOT_FOUND", "ABAC policy not found").
				WithHTTPStatus(http.StatusNotFound).
				WithCategory(CategorySecurity).
				WithSuggestion("Verify the policy ID is correct").
				WithSuggestion("Check if the policy exists and is active")

	ErrPolicyInvalid = NewBusinessError("POLICY_INVALID", "ABAC policy validation failed").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryValidation).
				WithSuggestion("Review policy structure and rules").
				WithSuggestion("Ensure all required fields are provided")

	ErrPolicyConflict = NewBusinessError("POLICY_CONFLICT", "Conflicting ABAC policies detected").
				WithHTTPStatus(http.StatusConflict).
				WithCategory(CategorySecurity).
				WithSuggestion("Review policy priorities and combining algorithms").
				WithSuggestion("Resolve conflicting policy rules")

	ErrPolicyExpired = NewBusinessError("POLICY_EXPIRED", "ABAC policy has expired").
				WithHTTPStatus(http.StatusGone).
				WithCategory(CategorySecurity).
				WithSuggestion("Update the policy expiration date").
				WithSuggestion("Create a new version of the policy if needed")

	// Attribute Errors
	ErrAttributeNotFound = NewBusinessError("ATTRIBUTE_NOT_FOUND", "Required attribute not found").
				WithHTTPStatus(http.StatusNotFound).
				WithCategory(CategorySecurity).
				WithSuggestion("Ensure all required attributes are available").
				WithSuggestion("Check attribute collection configuration")

	ErrAttributeInvalid = NewBusinessError("ATTRIBUTE_INVALID", "Attribute value validation failed").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryValidation).
				WithSuggestion("Verify attribute value matches the expected data type").
				WithSuggestion("Check attribute constraints and allowed values")

	ErrAttributeExpired = NewBusinessError("ATTRIBUTE_EXPIRED", "Attribute value has expired").
				WithHTTPStatus(http.StatusGone).
				WithCategory(CategorySecurity).
				WithSuggestion("Refresh attribute values from their sources").
				WithSuggestion("Check attribute TTL configuration")

	ErrAttributeDefinitionInvalid = NewBusinessError("ATTRIBUTE_DEFINITION_INVALID", "Attribute definition validation failed").
					WithHTTPStatus(http.StatusBadRequest).
					WithCategory(CategoryValidation).
					WithSuggestion("Check attribute definition structure and constraints").
					WithSuggestion("Ensure data type and category are valid")

	// Evaluation Errors
	ErrEvaluationFailed = NewBusinessError("EVALUATION_FAILED", "Policy evaluation failed").
				WithHTTPStatus(http.StatusInternalServerError).
				WithCategory(CategorySecurity).
				WithSuggestion("Check policy rules and attribute availability").
				WithSuggestion("Review evaluation context and parameters")

	ErrEvaluationTimeout = NewBusinessError("EVALUATION_TIMEOUT", "Policy evaluation timed out").
				WithHTTPStatus(http.StatusRequestTimeout).
				WithCategory(CategorySecurity).
				WithSuggestion("Simplify policy rules for better performance").
				WithSuggestion("Check system resources and performance")

	ErrCombiningAlgorithmFailed = NewBusinessError("COMBINING_ALGORITHM_FAILED", "Policy combining algorithm failed").
					WithHTTPStatus(http.StatusInternalServerError).
					WithCategory(CategorySecurity).
					WithSuggestion("Review policy combining algorithm configuration").
					WithSuggestion("Check for conflicting policy decisions")

	ErrInsufficientAttributes = NewBusinessError("INSUFFICIENT_ATTRIBUTES", "Insufficient attributes for policy evaluation").
					WithHTTPStatus(http.StatusBadRequest).
					WithCategory(CategorySecurity).
					WithSuggestion("Provide all required attributes for evaluation").
					WithSuggestion("Check attribute collection sources")

	// ErrEntityHasChildren indicates entity has child entities
	ErrEntityHasChildren = NewBusinessError("ENTITY_HAS_CHILDREN", "Entity has child entities and cannot be deleted").
				WithHTTPStatus(http.StatusConflict).
				WithCategory(CategoryBusiness).
				WithSuggestion("Remove or reassign child entities before deleting").
				WithSuggestion("Consider deactivating instead of deleting")

	// ErrInvalidEntityType indicates invalid entity type
	ErrInvalidEntityType = NewBusinessError("INVALID_ENTITY_TYPE", "Invalid entity type").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryValidation).
				WithSuggestion("Use a valid entity type (company, department, project, etc.)").
				WithSuggestion("Check the API documentation for valid entity types")

	// ─── GENERAL ERRORS ─────────────────────────────────────────────

	// ErrInvalidInput indicates general invalid input
	ErrInvalidInput = NewBusinessError("INVALID_INPUT", "Invalid input provided").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(CategoryValidation).
			WithSuggestion("Check the request format and required fields").
			WithSuggestion("Refer to the API documentation for correct input format")

	// ErrNotFound indicates general resource not found
	ErrNotFound = NewBusinessError("NOT_FOUND", "Resource not found").
			WithHTTPStatus(http.StatusNotFound).
			WithCategory(CategoryBusiness).
			WithSuggestion("Verify the resource ID is correct").
			WithSuggestion("Check if the resource exists and you have access to it")
)

// ─── ENHANCED ERROR CONSTRUCTORS ─────────────────────────────────────────────
// These functions create contextual variations of the base errors

// NewUserNotFoundError creates a user not found error with specific user context
func NewUserNotFoundError(userID string) *BusinessError {
	return NewBusinessError("USER_NOT_FOUND", "User not found").
		WithHTTPStatus(http.StatusNotFound).
		WithCategory(CategorySecurity).
		WithDetail("user_id", userID).
		WithSuggestion("Verify the user ID is correct").
		WithSuggestion("Check if the user exists in your tenant")
}

// NewUserNotFoundByEmailError creates a user not found error with email context
func NewUserNotFoundByEmailError(email string) *BusinessError {
	return NewBusinessError("USER_NOT_FOUND", "User not found").
		WithHTTPStatus(http.StatusNotFound).
		WithCategory(CategorySecurity).
		WithDetail("email", email).
		WithSuggestion("Verify the email address is correct").
		WithSuggestion("Check if the user exists in your tenant")
}

// NewEntityNotFoundError creates an entity not found error with specific entity context
func NewEntityNotFoundError(entityID string) *BusinessError {
	return NewBusinessError("ENTITY_NOT_FOUND", "Entity not found").
		WithHTTPStatus(http.StatusNotFound).
		WithCategory(CategoryBusiness).
		WithDetail("entity_id", entityID).
		WithSuggestion("Verify the entity ID is correct").
		WithSuggestion("Check if the entity exists in your tenant")
}

// NewRoleNotFoundError creates a role not found error with specific role context
func NewRoleNotFoundError(roleID string) *BusinessError {
	return NewBusinessError("ROLE_NOT_FOUND", "Role not found").
		WithHTTPStatus(http.StatusNotFound).
		WithCategory(CategorySecurity).
		WithDetail("role_id", roleID).
		WithSuggestion("Verify the role ID is correct").
		WithSuggestion("Check if the role exists in your tenant")
}

// NewInvalidCredentialsError creates an invalid credentials error with login attempt context
func NewInvalidCredentialsError(email string, attemptCount int) *BusinessError {
	err := NewBusinessError("INVALID_CREDENTIALS", "Invalid email or password").
		WithHTTPStatus(http.StatusUnauthorized).
		WithCategory(CategorySecurity).
		WithDetail("email", email).
		WithDetail("attempt_count", attemptCount).
		WithSuggestion("Check your email and password").
		WithSuggestion("Use the 'Forgot Password' option if needed")

	// Add account lockout warning for multiple attempts
	if attemptCount >= 3 {
		err.WithSuggestion("Multiple failed attempts detected - account may be locked").
			WithSeverity(SeverityWarning)
	}

	return err
}

// NewEntityNameExistsError creates an entity name exists error with specific name context
func NewEntityNameExistsError(name string) *BusinessError {
	return NewBusinessError("ENTITY_NAME_EXISTS", "Entity name already exists").
		WithHTTPStatus(http.StatusConflict).
		WithCategory(CategoryBusiness).
		WithDetail("name", name).
		WithSuggestion("Choose a different entity name").
		WithSuggestion("Check existing entities to avoid duplicates")
}

// NewEntityCodeExistsError creates an entity code exists error with specific code context
func NewEntityCodeExistsError(code string) *BusinessError {
	return NewBusinessError("ENTITY_CODE_EXISTS", "Entity code already exists").
		WithHTTPStatus(http.StatusConflict).
		WithCategory(CategoryBusiness).
		WithDetail("code", code).
		WithSuggestion("Use a different entity code").
		WithSuggestion("Entity codes must be unique within your organization")
}

// NewCircularReferenceError creates a circular reference error with hierarchy context
func NewCircularReferenceError(entityID, parentID string) *BusinessError {
	return NewBusinessError("CIRCULAR_REFERENCE", "Circular reference detected in entity hierarchy").
		WithHTTPStatus(http.StatusBadRequest).
		WithCategory(CategoryBusiness).
		WithSeverity(SeverityError).
		WithDetail("entity_id", entityID).
		WithDetail("parent_id", parentID).
		WithSuggestion("Review the entity hierarchy to remove circular references").
		WithSuggestion("An entity cannot be a parent of itself or its ancestors")
}

// NewInvitationExpiredError creates an invitation expired error with expiration context
func NewInvitationExpiredError(invitationID string, expiredAt string) *BusinessError {
	return NewBusinessError("INVITATION_EXPIRED", "Invitation has expired").
		WithHTTPStatus(http.StatusGone).
		WithCategory(CategoryBusiness).
		WithDetail("invitation_id", invitationID).
		WithDetail("expired_at", expiredAt).
		WithSuggestion("Request a new invitation").
		WithSuggestion("Contact an administrator for a fresh invitation link")
}

// NewFeatureNotEnabledError creates a feature not enabled error with feature context
func NewFeatureNotEnabledError(feature string, planRequired string) *BusinessError {
	return NewBusinessError("FEATURE_NOT_ENABLED", "Feature not enabled for this tenant").
		WithHTTPStatus(http.StatusForbidden).
		WithCategory(CategoryTenant).
		WithDetail("feature", feature).
		WithDetail("plan_required", planRequired).
		WithSuggestion(fmt.Sprintf("Upgrade to %s plan to access this feature", planRequired)).
		WithSuggestion("Contact sales for more information about feature availability")
}

// ─── ERROR TYPE CHECKING HELPERS ─────────────────────────────────────────────
// These helpers maintain backward compatibility for error checking

// IsUserNotFound checks if error is user not found
func IsUserNotFound(err error) bool {
	return IsBusinessErrorCode(err, "USER_NOT_FOUND")
}

// IsEntityNotFound checks if error is entity not found
func IsEntityNotFound(err error) bool {
	return IsBusinessErrorCode(err, "ENTITY_NOT_FOUND")
}

// IsRoleNotFound checks if error is role not found
func IsRoleNotFound(err error) bool {
	return IsBusinessErrorCode(err, "ROLE_NOT_FOUND")
}

// IsTenantNotFound checks if error is tenant not found
func IsTenantNotFound(err error) bool {
	return IsBusinessErrorCode(err, "TENANT_NOT_FOUND")
}

// IsInvalidCredentials checks if error is invalid credentials
func IsInvalidCredentials(err error) bool {
	return IsBusinessErrorCode(err, "INVALID_CREDENTIALS")
}

// IsUnauthorized checks if error is unauthorized
func IsUnauthorized(err error) bool {
	return IsBusinessErrorCode(err, "UNAUTHORIZED")
}

// IsForbidden checks if error is forbidden
func IsForbidden(err error) bool {
	return IsBusinessErrorCode(err, "FORBIDDEN")
}

// IsConflict checks if error is a conflict (exists) error
func IsConflict(err error) bool {
	conflictCodes := []string{
		"USER_EXISTS", "ENTITY_NAME_EXISTS", "ENTITY_CODE_EXISTS",
		"ROLE_EXISTS", "TENANT_EXISTS", "EMAIL_EXISTS", "USERNAME_EXISTS",
		"SUBDOMAIN_EXISTS",
	}

	for _, code := range conflictCodes {
		if IsBusinessErrorCode(err, code) {
			return true
		}
	}
	return false
}

// IsValidationError checks if error is a validation error
func IsValidationError(err error) bool {
	if _, ok := err.(ValidationErrors); ok {
		return true
	}
	if _, ok := err.(ValidationError); ok {
		return true
	}
	if be, ok := err.(*BusinessError); ok {
		return be.Category == CategoryValidation
	}
	return false
}

// IsTemporaryError checks if error is temporary/retryable
func IsTemporaryError(err error) bool {
	return IsTemporary(err)
}

// IsBusinessErrorCode checks if error is a BusinessError with specific code
func IsBusinessErrorCode(err error, code string) bool {
	var be *BusinessError
	if errors.As(err, &be) {
		return be.Code == code
	}
	return false
}

// ─── MIGRATION HELPERS ─────────────────────────────────────────────
// These help migrate from simple errors to enhanced errors

// WrapSimpleError wraps a simple error with enhanced context
func WrapSimpleError(simpleErr error, code string, httpStatus int) *BusinessError {
	return NewBusinessError(code, simpleErr.Error()).
		WithHTTPStatus(httpStatus).
		WithDetail("original_error", simpleErr.Error())
}

// UpgradeError upgrades a simple error to enhanced error if possible
func UpgradeError(err error) error {
	if err == nil {
		return nil
	}

	// Already enhanced
	if _, ok := err.(*BusinessError); ok {
		return err
	}
	if _, ok := err.(*RepositoryError); ok {
		return err
	}
	if _, ok := err.(ValidationErrors); ok {
		return err
	}

	// Map common simple errors to enhanced ones
	errMsg := err.Error()
	switch errMsg {
	case "user not found":
		return ErrUserNotFound
	case "entity not found":
		return ErrEntityNotFound
	case "role not found":
		return ErrRoleNotFound
	case "tenant not found":
		return ErrTenantNotFound
	case "invalid credentials":
		return ErrInvalidCredentials
	case "unauthorized":
		return ErrUnauthorized
	case "forbidden":
		return ErrForbidden
	default:
		// Generic upgrade
		return WrapSimpleError(err, "UNKNOWN_ERROR", http.StatusInternalServerError)
	}
}
