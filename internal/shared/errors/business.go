package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ─── PREDEFINED BUSINESS ERRORS ─────────────────────────────────────────────
// These replace the simple error variables with enhanced BusinessError instances
// while maintaining backward compatibility

var (
	// ─── TENANT ERRORS ─────────────────────────────────────────────

	// ErrTenantExists indicates a tenant already exists
	ErrTenantExists = NewBusinessError(CodeTenantExists, "Tenant already exists").
			WithHTTPStatus(http.StatusConflict).
			WithCategory(CategoryTenant).
			WithSuggestion("Choose a different tenant name or slug").
			WithSuggestion("Check if you're trying to create a duplicate tenant")

	// ErrTenantNotFound indicates a tenant was not found
	ErrTenantNotFound = NewBusinessError(CodeTenantNotFound, "Tenant not found").
				WithHTTPStatus(http.StatusNotFound).
				WithCategory(CategoryTenant).
				WithSuggestion("Verify the tenant ID or slug is correct").
				WithSuggestion("Contact support if you believe this tenant should exist")

	// ErrTenantIDNotInContext indicates tenant context is missing
	ErrTenantIDNotInContext = NewBusinessError(CodeTenantContextMissing, "Tenant ID not found in request context").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryTenant).
				WithSeverity(SeverityError).
				WithSuggestion("Ensure the request includes proper tenant identification").
				WithSuggestion("Check that tenant middleware is properly configured")

	// ErrSubdomainAlreadyExists indicates subdomain conflict
	ErrSubdomainAlreadyExists = NewBusinessError(CodeSubdomainExists, "Subdomain already exists").
					WithHTTPStatus(http.StatusConflict).
					WithCategory(CategoryTenant).
					WithSuggestion("Choose a different subdomain").
					WithSuggestion("Try adding numbers or variations to make it unique")

	// ErrSubdomainSoftDeleted indicates a subdomain is taken by a soft-deleted tenant
	ErrSubdomainSoftDeleted = NewBusinessError(CodeSubdomainSoftDeleted, "Subdomain is unavailable").
				WithHTTPStatus(http.StatusConflict).
				WithCategory(CategoryTenant).
				WithSuggestion("This subdomain was recently used. Please choose a different one or try again later.").
				WithSuggestion("If you own this subdomain and want to reactivate it, please contact support.")

	// ErrFeatureNotEnabled indicates a feature is not available for the tenant
	ErrFeatureNotEnabled = NewBusinessError(CodeFeatureNotEnabled, "Feature not enabled for this tenant").
				WithHTTPStatus(http.StatusForbidden).
				WithCategory(CategoryTenant).
				WithSuggestion("Upgrade your plan to access this feature").
				WithSuggestion("Contact sales for more information about feature availability")

	// ─── USER ERRORS ─────────────────────────────────────────────

	// ErrUserNotFound indicates a user was not found
	ErrUserNotFound = NewBusinessError(CodeUserNotFound, "User not found").
			WithHTTPStatus(http.StatusNotFound).
			WithCategory(CategorySecurity).
			WithSuggestion("Verify the user ID or email is correct").
			WithSuggestion("Check if the user exists in your tenant")

	// ErrUserExists indicates a user already exists
	ErrUserExists = NewBusinessError(CodeUserExists, "User already exists").
			WithHTTPStatus(http.StatusConflict).
			WithCategory(CategorySecurity).
			WithSuggestion("Use a different email address").
			WithSuggestion("Try logging in if you already have an account")

	// ErrEmailAlreadyExists indicates email address is taken
	ErrEmailAlreadyExists = NewBusinessError(CodeEmailExists, "Email address already exists").
				WithHTTPStatus(http.StatusConflict).
				WithCategory(CategoryValidation).
				WithSuggestion("Use a different email address").
				WithSuggestion("Check if you already have an account with this email")

	// ErrUsernameAlreadyExists indicates username is taken
	ErrUsernameAlreadyExists = NewBusinessError(CodeUsernameExists, "Username already exists").
					WithHTTPStatus(http.StatusConflict).
					WithCategory(CategoryValidation).
					WithSuggestion("Choose a different username").
					WithSuggestion("Try adding numbers or variations to make it unique")

	// ErrInvalidCredentials indicates authentication failed
	ErrInvalidCredentials = NewBusinessError(CodeInvalidCredentials, "Invalid email or password").
				WithHTTPStatus(http.StatusUnauthorized).
				WithCategory(CategorySecurity).
				WithSuggestion("Check your email and password").
				WithSuggestion("Use the 'Forgot Password' option if needed").
				WithSuggestion("Ensure caps lock is not enabled")

	// ErrAuthenticationFailed indicates general authentication failure
	ErrAuthenticationFailed = NewBusinessError(CodeAuthenticationFailed, "Authentication failed").
				WithHTTPStatus(http.StatusUnauthorized).
				WithCategory(CategorySecurity).
				WithSuggestion("Verify your credentials and try again").
				WithSuggestion("Contact support if the issue persists")

	// ErrInvalidUserType indicates an invalid user type
	ErrInvalidUserType = NewBusinessError(CodeInvalidUserType, "Invalid user type").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryValidation).
				WithSuggestion("Use a valid user type (admin, user, viewer, etc.)").
				WithSuggestion("Check the API documentation for valid user types")

	// ErrInvalidAccountStatus indicates an invalid account status
	ErrInvalidAccountStatus = NewBusinessError(CodeInvalidAccountStatus, "Invalid account status").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryValidation).
				WithSuggestion("Use a valid status (active, inactive, pending, suspended)").
				WithSuggestion("Check the API documentation for valid status values")

	// ErrAccountLocked indicates account is locked
	ErrAccountLocked = NewBusinessError(CodeAccountLocked, "Account is locked").
				WithHTTPStatus(http.StatusForbidden).
				WithCategory(CategorySecurity).
				WithSeverity(SeverityWarning).
				WithSuggestion("Contact an administrator to unlock your account").
				WithSuggestion("Wait for the automatic unlock period if applicable")

	// ─── ROLE & PERMISSION ERRORS ─────────────────────────────────────────────

	// ErrRoleNotFound indicates a role was not found
	ErrRoleNotFound = NewBusinessError(CodeRoleNotFound, "Role not found").
			WithHTTPStatus(http.StatusNotFound).
			WithCategory(CategorySecurity).
			WithSuggestion("Verify the role ID or name is correct").
			WithSuggestion("Check if the role exists in your tenant")

	// ErrRoleExists indicates a role already exists
	ErrRoleExists = NewBusinessError(CodeRoleExists, "Role already exists").
			WithHTTPStatus(http.StatusConflict).
			WithCategory(CategorySecurity).
			WithSuggestion("Use a different role name").
			WithSuggestion("Check existing roles to avoid duplicates")

	// ErrUnauthorized indicates unauthorized access
	ErrUnauthorized = NewBusinessError(CodeUnauthorized, "Unauthorized access").
			WithHTTPStatus(http.StatusUnauthorized).
			WithCategory(CategorySecurity).
			WithSuggestion("Please log in to access this resource").
			WithSuggestion("Verify your authentication token is valid")

	// ErrForbidden indicates forbidden access
	ErrForbidden = NewBusinessError(CodeForbidden, "Access forbidden").
			WithHTTPStatus(http.StatusForbidden).
			WithCategory(CategorySecurity).
			WithSuggestion("Contact your administrator for access").
			WithSuggestion("Verify you have the required permissions")

	// ─── INVITATION ERRORS ─────────────────────────────────────────────

	// ErrInvitationNotFound indicates invitation was not found
	ErrInvitationNotFound = NewBusinessError(CodeInvitationNotFound, "Invitation not found").
				WithHTTPStatus(http.StatusNotFound).
				WithCategory(CategoryBusiness).
				WithSuggestion("Verify the invitation link is correct").
				WithSuggestion("Contact the person who sent the invitation")

	// ErrInvitationExpired indicates invitation has expired
	ErrInvitationExpired = NewBusinessError(CodeInvitationExpired, "Invitation has expired").
				WithHTTPStatus(http.StatusGone).
				WithCategory(CategoryBusiness).
				WithSuggestion("Request a new invitation").
				WithSuggestion("Contact an administrator for a fresh invitation link")

	// ─── API KEY ERRORS ─────────────────────────────────────────────

	// ErrAPIKeyNotFound indicates API key was not found
	ErrAPIKeyNotFound = NewBusinessError(CodeAPIKeyNotFound, "API key not found").
				WithHTTPStatus(http.StatusUnauthorized).
				WithCategory(CategorySecurity).
				WithSuggestion("Verify your API key is correct").
				WithSuggestion("Generate a new API key if the current one is invalid")

	// ─── ENTITY ERRORS ─────────────────────────────────────────────

	// ErrEntityNotFound indicates an entity was not found
	ErrEntityNotFound = NewBusinessError(CodeEntityNotFound, "Entity not found").
				WithHTTPStatus(http.StatusNotFound).
				WithCategory(CategoryBusiness).
				WithSuggestion("Verify the entity ID is correct").
				WithSuggestion("Check if the entity exists in your tenant")

	// ErrEntityNameExists indicates entity name already exists
	ErrEntityNameExists = NewBusinessError(CodeEntityNameExists, "Entity name already exists").
				WithHTTPStatus(http.StatusConflict).
				WithCategory(CategoryBusiness).
				WithSuggestion("Choose a different entity name").
				WithSuggestion("Check existing entities to avoid duplicates")

	// ErrEntityCodeExists indicates entity code already exists
	ErrEntityCodeExists = NewBusinessError(CodeEntityCodeExists, "Entity code already exists").
				WithHTTPStatus(http.StatusConflict).
				WithCategory(CategoryBusiness).
				WithSuggestion("Use a different entity code").
				WithSuggestion("Entity codes must be unique within your organization")

	// ErrInvalidParentEntity indicates invalid parent entity reference
	ErrInvalidParentEntity = NewBusinessError(CodeInvalidParentEntity, "Invalid parent entity").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryBusiness).
				WithSuggestion("Verify the parent entity exists and is valid").
				WithSuggestion("Check that the parent entity allows child entities")

	// ErrCircularReference indicates circular reference in entity hierarchy
	ErrCircularReference = NewBusinessError(CodeCircularReference, "Circular reference detected in entity hierarchy").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryBusiness).
				WithSeverity(SeverityError).
				WithSuggestion("Review the entity hierarchy to remove circular references").
				WithSuggestion("An entity cannot be a parent of itself or its ancestors")

	// ErrEntityHasChildren indicates entity has child entities
	ErrEntityHasChildren = NewBusinessError(CodeEntityHasChildren, "Entity has child entities and cannot be deleted").
				WithHTTPStatus(http.StatusConflict).
				WithCategory(CategoryBusiness).
				WithSuggestion("Remove or reassign child entities before deleting").
				WithSuggestion("Consider deactivating instead of deleting")

	// ErrInvalidEntityType indicates invalid entity type
	ErrInvalidEntityType = NewBusinessError(CodeInvalidEntityType, "Invalid entity type").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryValidation).
				WithSuggestion("Use a valid entity type (company, department, project, etc.)").
				WithSuggestion("Check the API documentation for valid entity types")

	// ─── ABAC ERRORS ─────────────────────────────────────────────

	// Policy Errors
	ErrPolicyNotFound = NewBusinessError(CodePolicyNotFound, "ABAC policy not found").
				WithHTTPStatus(http.StatusNotFound).
				WithCategory(CategorySecurity).
				WithSuggestion("Verify the policy ID is correct").
				WithSuggestion("Check if the policy exists and is active")

	ErrPolicyInvalid = NewBusinessError(CodePolicyInvalid, "ABAC policy validation failed").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryValidation).
				WithSuggestion("Review policy structure and rules").
				WithSuggestion("Ensure all required fields are provided")

	ErrPolicyConflict = NewBusinessError(CodePolicyConflict, "Conflicting ABAC policies detected").
				WithHTTPStatus(http.StatusConflict).
				WithCategory(CategorySecurity).
				WithSuggestion("Review policy priorities and combining algorithms").
				WithSuggestion("Resolve conflicting policy rules")

	ErrPolicyExpired = NewBusinessError(CodePolicyExpired, "ABAC policy has expired").
				WithHTTPStatus(http.StatusGone).
				WithCategory(CategorySecurity).
				WithSuggestion("Update the policy expiration date").
				WithSuggestion("Create a new version of the policy if needed")

	// Attribute Errors
	ErrAttributeNotFound = NewBusinessError(CodeAttributeNotFound, "Required attribute not found").
				WithHTTPStatus(http.StatusNotFound).
				WithCategory(CategorySecurity).
				WithSuggestion("Ensure all required attributes are available").
				WithSuggestion("Check attribute collection configuration")

	ErrAttributeInvalid = NewBusinessError(CodeAttributeInvalid, "Attribute value validation failed").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(CategoryValidation).
				WithSuggestion("Verify attribute value matches the expected data type").
				WithSuggestion("Check attribute constraints and allowed values")

	ErrAttributeExpired = NewBusinessError(CodeAttributeExpired, "Attribute value has expired").
				WithHTTPStatus(http.StatusGone).
				WithCategory(CategorySecurity).
				WithSuggestion("Refresh attribute values from their sources").
				WithSuggestion("Check attribute TTL configuration")

	ErrAttributeDefinitionInvalid = NewBusinessError(CodeAttributeDefinitionInvalid, "Attribute definition validation failed").
					WithHTTPStatus(http.StatusBadRequest).
					WithCategory(CategoryValidation).
					WithSuggestion("Check attribute definition structure and constraints").
					WithSuggestion("Ensure data type and category are valid")

	// Evaluation Errors
	ErrEvaluationFailed = NewBusinessError(CodeEvaluationFailed, "Policy evaluation failed").
				WithHTTPStatus(http.StatusInternalServerError).
				WithCategory(CategorySecurity).
				WithSuggestion("Check policy rules and attribute availability").
				WithSuggestion("Review evaluation context and parameters")

	ErrEvaluationTimeout = NewBusinessError(CodeEvaluationTimeout, "Policy evaluation timed out").
				WithHTTPStatus(http.StatusRequestTimeout).
				WithCategory(CategorySecurity).
				WithSuggestion("Simplify policy rules for better performance").
				WithSuggestion("Check system resources and performance")

	ErrCombiningAlgorithmFailed = NewBusinessError(CodeCombiningAlgorithmFailed, "Policy combining algorithm failed").
					WithHTTPStatus(http.StatusInternalServerError).
					WithCategory(CategorySecurity).
					WithSuggestion("Review policy combining algorithm configuration").
					WithSuggestion("Check for conflicting policy decisions")

	ErrInsufficientAttributes = NewBusinessError(CodeInsufficientAttributes, "Insufficient attributes for policy evaluation").
					WithHTTPStatus(http.StatusBadRequest).
					WithCategory(CategorySecurity).
					WithSuggestion("Provide all required attributes for evaluation").
					WithSuggestion("Check attribute collection sources")

	// ─── GENERAL ERRORS ─────────────────────────────────────────────

	// ErrInvalidInput indicates general invalid input
	ErrInvalidInput = NewBusinessError(CodeInvalidInput, "Invalid input provided").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(CategoryValidation).
			WithSuggestion("Check the request format and required fields").
			WithSuggestion("Refer to the API documentation for correct input format")

	// ErrNotFound indicates general resource not found
	ErrNotFound = NewBusinessError(CodeNotFound, "Resource not found").
			WithHTTPStatus(http.StatusNotFound).
			WithCategory(CategoryBusiness).
			WithSuggestion("Verify the resource ID is correct").
			WithSuggestion("Check if the resource exists and you have access to it")
)

// ─── ENHANCED ERROR CONSTRUCTORS ─────────────────────────────────────────────
// These functions create contextual variations of the base errors

// NewUserNotFoundError creates a user not found error with specific user context
func NewUserNotFoundError(userID string) *BusinessError {
	return NewBusinessError(CodeUserNotFound, "User not found").
		WithHTTPStatus(http.StatusNotFound).
		WithCategory(CategorySecurity).
		WithDetail("user_id", userID).
		WithSuggestion("Verify the user ID is correct").
		WithSuggestion("Check if the user exists in your tenant")
}

// NewUserNotFoundByEmailError creates a user not found error with email context
func NewUserNotFoundByEmailError(email string) *BusinessError {
	return NewBusinessError(CodeUserNotFound, "User not found").
		WithHTTPStatus(http.StatusNotFound).
		WithCategory(CategorySecurity).
		WithDetail("email", email).
		WithSuggestion("Verify the email address is correct").
		WithSuggestion("Check if the user exists in your tenant")
}

// NewEntityNotFoundError creates an entity not found error with specific entity context
func NewEntityNotFoundError(entityID string) *BusinessError {
	return NewBusinessError(CodeEntityNotFound, "Entity not found").
		WithHTTPStatus(http.StatusNotFound).
		WithCategory(CategoryBusiness).
		WithDetail("entity_id", entityID).
		WithSuggestion("Verify the entity ID is correct").
		WithSuggestion("Check if the entity exists in your tenant")
}

// NewRoleNotFoundError creates a role not found error with specific role context
func NewRoleNotFoundError(roleID string) *BusinessError {
	return NewBusinessError(CodeRoleNotFound, "Role not found").
		WithHTTPStatus(http.StatusNotFound).
		WithCategory(CategorySecurity).
		WithDetail("role_id", roleID).
		WithSuggestion("Verify the role ID is correct").
		WithSuggestion("Check if the role exists in your tenant")
}

// NewInvalidCredentialsError creates an invalid credentials error with login attempt context
func NewInvalidCredentialsError(email string, attemptCount int) *BusinessError {
	err := NewBusinessError(CodeInvalidCredentials, "Invalid email or password").
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
	return NewBusinessError(CodeEntityNameExists, "Entity name already exists").
		WithHTTPStatus(http.StatusConflict).
		WithCategory(CategoryBusiness).
		WithDetail("name", name).
		WithSuggestion("Choose a different entity name").
		WithSuggestion("Check existing entities to avoid duplicates")
}

// NewEntityCodeExistsError creates an entity code exists error with specific code context
func NewEntityCodeExistsError(code string) *BusinessError {
	return NewBusinessError(CodeEntityCodeExists, "Entity code already exists").
		WithHTTPStatus(http.StatusConflict).
		WithCategory(CategoryBusiness).
		WithDetail("code", code).
		WithSuggestion("Use a different entity code").
		WithSuggestion("Entity codes must be unique within your organization")
}

// NewCircularReferenceError creates a circular reference error with hierarchy context
func NewCircularReferenceError(entityID, parentID string) *BusinessError {
	return NewBusinessError(CodeCircularReference, "Circular reference detected in entity hierarchy").
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
	return NewBusinessError(CodeInvitationExpired, "Invitation has expired").
		WithHTTPStatus(http.StatusGone).
		WithCategory(CategoryBusiness).
		WithDetail("invitation_id", invitationID).
		WithDetail("expired_at", expiredAt).
		WithSuggestion("Request a new invitation").
		WithSuggestion("Contact an administrator for a fresh invitation link")
}

// NewFeatureNotEnabledError creates a feature not enabled error with feature context
func NewFeatureNotEnabledError(feature string, planRequired string) *BusinessError {
	return NewBusinessError(CodeFeatureNotEnabled, "Feature not enabled for this tenant").
		WithHTTPStatus(http.StatusForbidden).
		WithCategory(CategoryTenant).
		WithDetail("feature", feature).
		WithDetail("plan_required", planRequired).
		WithSuggestion(fmt.Sprintf("Upgrade to %s plan to access this feature", planRequired)).
		WithSuggestion("Contact sales for more information about feature availability")
}

// ─── DOMAIN-SPECIFIC ERROR CONSTRUCTORS ─────────────────────────────────────────────

// Accounting-specific errors
func ErrInvalidAccountingPeriod(periodStart, periodEnd time.Time) *BusinessError {
	return NewBusinessError(CodeInvalidAccountingPeriod, "Invalid accounting period").
		WithDetail("period_start", periodStart).
		WithDetail("period_end", periodEnd).
		WithHTTPStatus(http.StatusBadRequest).
		WithSuggestion("Ensure the period start date is before the end date")
}

func ErrAccountingPeriodClosed(period string) *BusinessError {
	return NewBusinessError(CodeAccountingPeriodClosed, "Accounting period is closed").
		WithDetail("period", period).
		WithHTTPStatus(http.StatusConflict).
		WithSuggestion("Contact your accountant to reopen the period if necessary")
}

func ErrInsufficientBalance(accountID string, required, available float64) *BusinessError {
	return NewBusinessError(CodeInsufficientBalance, "Insufficient account balance").
		WithDetail("account_id", accountID).
		WithDetail("required", required).
		WithDetail("available", available).
		WithHTTPStatus(http.StatusConflict).
		WithSuggestion("Ensure sufficient funds are available before proceeding")
}

func ErrDuplicateTransaction(transactionID string) *BusinessError {
	return NewBusinessError(CodeDuplicateTransaction, "Transaction already exists").
		WithDetail("transaction_id", transactionID).
		WithHTTPStatus(http.StatusConflict).
		WithSuggestion("Check if this transaction was already recorded")
}

// Integration errors
func ErrThirdPartyAPIFailure(service string, statusCode int) *BusinessError {
	return NewBusinessError(CodeThirdPartyAPIFailure, fmt.Sprintf("External service %s failed", service)).
		WithDetail("service", service).
		WithDetail("status_code", statusCode).
		WithHTTPStatus(http.StatusBadGateway).
		WithCategory(CategoryIntegration).
		WithSuggestion("Please try again later").
		WithDetail("retryable", true)
}
