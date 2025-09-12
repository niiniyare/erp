package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/api/gen/auth"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/iam/authn"
	"github.com/niiniyare/erp/internal/core/iam/authz"
	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/core/iam/policy"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"goa.design/goa/v3/security"
)

// AuthHandler implements the GOA auth service following the data flow pattern
type AuthHandler struct {
	iamService    iam.Service
	tenantService tenant.Service
	tracing       tracing.TracingService
	metrics       *metrics.MetricsService
}

// NewAuthHandler creates a new auth handler following Clean Architecture pattern
func NewAuthHandler(iamSvc iam.Service, tenantSvc tenant.Service, tracing tracing.TracingService, metrics *metrics.MetricsService) auth.Service {
	return &AuthHandler{
		iamService:    iamSvc,
		tenantService: tenantSvc,
		tracing:       tracing,
		metrics:       metrics,
	}
}

// NewAuthHandlerWithIdentity creates a temporary auth handler using identity service
// TODO: Remove this when full IAM service is available
func NewAuthHandlerWithIdentity(identitySvc interface{}, tenantSvc tenant.Service, tracing tracing.TracingService, metrics *metrics.MetricsService) auth.Service {
	// Create a simple IAM adapter that wraps the identity service
	iamAdapter := &identityToIAMAdapter{identityService: identitySvc}

	return &AuthHandler{
		iamService:    iamAdapter,
		tenantService: tenantSvc,
		tracing:       tracing,
		metrics:       metrics,
	}
}

// JWTAuth implements the authorization logic for the JWT security scheme
func (h *AuthHandler) JWTAuth(ctx context.Context, token string, scheme *security.JWTScheme) (context.Context, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "auth.jwt_auth",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("security.scheme", "jwt"),
			attribute.String("token.prefix", token[:min(len(token), 10)]),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("auth_jwt_validation_duration", metrics.Fields{})
	defer timer.Stop()

	// Validate token using IAM authentication service
	result, err := h.iamService.Authentication().ValidateToken(ctx, token)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("auth_jwt_validation_failed_total", metrics.Fields{
			"reason": "validation_error",
		})

		logger.WarnContext(ctx, "JWT token validation failed", logger.Fields{
			"token_prefix": token[:min(len(token), 10)],
			"error":        err.Error(),
		})

		return nil, fmt.Errorf("invalid token")
	}

	if !result.Valid {
		h.metrics.IncrementCounter("auth_jwt_validation_failed_total", metrics.Fields{
			"reason": "invalid_token",
		})

		logger.WarnContext(ctx, "JWT token is invalid", logger.Fields{
			"token_prefix": token[:min(len(token), 10)],
		})

		return nil, fmt.Errorf("invalid token")
	}

	// Add user ID to context for downstream services
	ctx = context.WithValue(ctx, "user_id", result.UserID)

	logger.InfoContext(ctx, "JWT token validated successfully", logger.Fields{
		"user_id":      result.UserID.String(),
		"token_prefix": token[:min(len(token), 10)],
	})

	h.metrics.IncrementCounter("auth_jwt_validations_total", metrics.Fields{
		"status": "success",
	})
	return ctx, nil
}

// Login authenticates user and returns JWT token following data flow pattern
func (h *AuthHandler) Login(ctx context.Context, p *auth.LoginPayload) (*auth.Auth, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "auth.login",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user.email", p.Email),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("auth_login_duration", metrics.Fields{})
	defer timer.Stop()

	logger.InfoContext(ctx, "Processing login request", logger.Fields{
		"email": p.Email,
	})

	// Authenticate user with IAM service
	authReq := &authn.AuthenticationRequest{
		Email:    p.Email,
		Password: p.Password,
		MFACode:  "", // TODO: Add MFA support from payload if needed
	}

	authResult, err := h.iamService.Authentication().Authenticate(ctx, authReq)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("auth_login_failed_total", metrics.Fields{
			"reason": "authentication_failed",
		})

		logger.WarnContext(ctx, "Authentication failed", logger.Fields{
			"email": p.Email,
			"error": err.Error(),
		})

		// Return authentication error - GOA will convert to HTTP 401
		return nil, auth.MakeUnauthorized(err)
	}

	user := authResult.User

	// Check if account is active
	if !user.IsActive() {
		h.metrics.IncrementCounter("auth_login_failed_total", metrics.Fields{
			"reason": "account_inactive",
		})

		logger.WarnContext(ctx, "Login attempt for inactive account", logger.Fields{
			"user_id":        user.ID.String(),
			"email":          user.Email,
			"account_status": string(user.AccountStatus),
			"is_active":      user.IsActive(),
		})

		return nil, auth.MakeUnauthorized(fmt.Errorf("account is not active"))
	}

	// Use tokens from authentication result or generate if not present
	accessToken := authResult.AccessToken
	refreshToken := authResult.RefreshToken

	if accessToken == "" {
		// Fallback to mock token if not generated
		accessToken = "jwt-token-" + user.ID.String()
	}
	if refreshToken == "" {
		// Fallback to mock refresh token if not generated
		refreshToken = "refresh-token-" + user.ID.String()
	}

	// Get user roles from IAM service
	userRoles, err := h.iamService.Authentication().GetUserRoles(ctx, user.ID)
	if err != nil {
		logger.WarnContext(ctx, "Failed to get user roles", logger.Fields{
			"user_id": user.ID.String(),
			"error":   err.Error(),
		})
		userRoles = []*model.Role{} // Default to empty roles
	}

	// Get user effective permissions from authorization service
	effectivePerms, err := h.iamService.Authorization().GetUserEffectivePermissions(ctx, user.ID, &user.EntityID)
	if err != nil {
		logger.WarnContext(ctx, "Failed to get user permissions", logger.Fields{
			"user_id": user.ID.String(),
			"error":   err.Error(),
		})
		effectivePerms = &authz.UserEffectivePermissions{Permissions: make(map[string][]string)}
	}

	// Convert permissions to string array for API response
	var permissions []string
	for resourceType, actions := range effectivePerms.Permissions {
		for _, action := range actions {
			permissions = append(permissions, resourceType+":"+action)
		}
	}

	// Get tenant information
	tenantInfo, err := h.getTenantInfo(ctx, user.TenantID)
	if err != nil {
		logger.WarnContext(ctx, "Failed to get tenant info", logger.Fields{
			"tenant_id": user.TenantID.String(),
			"error":     err.Error(),
		})
		// Use default tenant info on error
		tenantInfo = &auth.TenantInfo{
			ID:        user.TenantID.String(),
			Name:      "Default Tenant",
			Subdomain: "default",
			Status:    "active",
		}
	}

	// Get person details if available
	var firstName, lastName string
	if user.PersonID != nil {
		if person, err := h.iamService.Authentication().GetPerson(ctx, *user.PersonID); err == nil {
			firstName = person.FirstName
			lastName = person.LastName
		}
	}

	// Fallback to user's own name fields if person not found
	if firstName == "" {
		firstName = user.FirstName
	}
	if lastName == "" {
		lastName = user.LastName
	}

	// Build role name for response (use first role or default)
	var primaryRole string
	if len(userRoles) > 0 {
		primaryRole = userRoles[0].Name
	} else {
		primaryRole = "user" // Default role
	}

	// Convert domain user to GOA response
	result := &auth.Auth{
		AccessToken:  accessToken,
		RefreshToken: &refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    uint(authResult.ExpiresAt.Unix() - time.Now().Unix()), // Calculate remaining time
		User: &auth.UserInfo{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: firstName,
			LastName:  lastName,
			Role:      &primaryRole,
		},
		Tenant:      tenantInfo,
		Permissions: permissions,
	}

	// TODO: Update last login timestamp
	// h.userService.UpdateLastLogin(ctx, user.ID)

	h.metrics.IncrementCounter("auth_login_total", metrics.Fields{
		"status": "success",
	})
	span.SetAttributes(
		attribute.String("result.user_id", result.User.ID),
		attribute.String("result.tenant_id", result.Tenant.ID),
	)

	logger.InfoContext(ctx, "Login successful", logger.Fields{
		"user_id":   user.ID.String(),
		"tenant_id": user.TenantID.String(),
		"email":     user.Email,
	})

	return result, nil
}

// getTenantInfo helper method to retrieve tenant information
func (h *AuthHandler) getTenantInfo(ctx context.Context, tenantID uuid.UUID) (*auth.TenantInfo, error) {
	tenant, err := h.tenantService.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return &auth.TenantInfo{
		ID:        tenant.ID.String(),
		Name:      tenant.Name,
		Subdomain: tenant.Slug,
		Status:    "active",
	}, nil
}

// Refresh refreshes JWT token following data flow pattern
func (h *AuthHandler) Refresh(ctx context.Context, p *auth.RefreshPayload) (*auth.Auth, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "auth.refresh",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("auth_refresh_duration", metrics.Fields{})
	defer timer.Stop()

	// TODO: Implement token refresh logic using existing user service
	result := &auth.Auth{
		AccessToken: "refreshed-jwt-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}

	h.metrics.IncrementCounter("auth_refresh_total", metrics.Fields{})
	return result, nil
}

// Logout logs out user and invalidates token following data flow pattern
func (h *AuthHandler) Logout(ctx context.Context, p *auth.LogoutPayload) error {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "auth.logout",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("auth_logout_duration", metrics.Fields{})
	defer timer.Stop()

	// TODO: Implement logout logic using existing user service
	logger.Info("Auth logout called", logger.Fields{})

	h.metrics.IncrementCounter("auth_logout_total", metrics.Fields{})
	return nil
}

// Validate validates JWT token following data flow pattern
func (h *AuthHandler) Validate(ctx context.Context, p *auth.ValidatePayload) (*auth.TokenValidation, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "auth.validate",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("auth_validate_duration", metrics.Fields{})
	defer timer.Stop()

	logger.InfoContext(ctx, "Validating JWT token", logger.Fields{
		"token_prefix": (*p.Token)[:min(len(*p.Token), 10)],
	})

	// TODO: Implement proper JWT token validation
	// For now, extracting user ID from mock token format: "jwt-token-<user-id>"
	var userID string
	if len(*p.Token) > 10 && (*p.Token)[:10] == "jwt-token-" {
		userID = (*p.Token)[10:] // Extract user ID from mock token
	} else {
		h.metrics.IncrementCounter("auth_validate_failed_total", metrics.Fields{
			"reason": "invalid_token_format",
		})

		logger.WarnContext(ctx, "Invalid token format", logger.Fields{
			"token_prefix": (*p.Token)[:min(len(*p.Token), 10)],
		})

		return &auth.TokenValidation{
			Valid: false,
		}, nil
	}

	// Parse user ID and validate user exists
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.metrics.IncrementCounter("auth_validate_failed_total", metrics.Fields{
			"reason": "invalid_user_id",
		})

		return &auth.TokenValidation{
			Valid: false,
		}, nil
	}

	// Get user from IAM authentication service
	user, err := h.iamService.Authentication().GetUser(ctx, userUUID)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("auth_validate_failed_total", metrics.Fields{
			"reason": "user_not_found",
		})

		logger.WarnContext(ctx, "User not found for token validation", logger.Fields{
			"user_id": userID,
			"error":   err.Error(),
		})

		return &auth.TokenValidation{
			Valid: false,
		}, nil
	}

	// Check if user account is still active
	if !user.IsActive() {
		h.metrics.IncrementCounter("auth_validate_failed_total", metrics.Fields{
			"reason": "account_inactive",
		})

		logger.WarnContext(ctx, "Token validation failed - account inactive", logger.Fields{
			"user_id":        user.ID.String(),
			"account_status": string(user.AccountStatus),
			"is_active":      user.IsActive(),
		})

		return &auth.TokenValidation{
			Valid: false,
		}, nil
	}

	// Get user roles from IAM service
	userRoles, err := h.iamService.Authentication().GetUserRoles(ctx, user.ID)
	if err != nil {
		logger.WarnContext(ctx, "Failed to get user roles for token validation", logger.Fields{
			"user_id": user.ID.String(),
			"error":   err.Error(),
		})
		userRoles = []*model.Role{} // Default to empty roles
	}

	// Get user effective permissions from authorization service
	effectivePerms, err := h.iamService.Authorization().GetUserEffectivePermissions(ctx, user.ID, &user.EntityID)
	if err != nil {
		logger.WarnContext(ctx, "Failed to get user permissions for token validation", logger.Fields{
			"user_id": user.ID.String(),
			"error":   err.Error(),
		})
		effectivePerms = &authz.UserEffectivePermissions{Permissions: make(map[string][]string)}
	}

	// Convert permissions to string array for API response
	var permissions []string
	for resourceType, actions := range effectivePerms.Permissions {
		for _, action := range actions {
			permissions = append(permissions, resourceType+":"+action)
		}
	}

	// Get tenant information
	tenantInfo, err := h.getTenantInfo(ctx, user.TenantID)
	if err != nil {
		logger.WarnContext(ctx, "Failed to get tenant info for token validation", logger.Fields{
			"tenant_id": user.TenantID.String(),
			"error":     err.Error(),
		})
		// Use default tenant info on error
		tenantInfo = &auth.TenantInfo{
			ID:        user.TenantID.String(),
			Name:      "Default Tenant",
			Subdomain: "default",
			Status:    "active",
		}
	}

	// Get person details if available
	var firstName, lastName string
	if user.PersonID != nil {
		if person, err := h.iamService.Authentication().GetPerson(ctx, *user.PersonID); err == nil {
			firstName = person.FirstName
			lastName = person.LastName
		}
	}

	// Fallback to user's own name fields if person not found
	if firstName == "" {
		firstName = user.FirstName
	}
	if lastName == "" {
		lastName = user.LastName
	}

	// Build role name for response (use first role or default)
	var primaryRole string
	if len(userRoles) > 0 {
		primaryRole = userRoles[0].Name
	} else {
		primaryRole = "user" // Default role
	}

	// Token is valid, return user info
	result := &auth.TokenValidation{
		Valid: true,
		User: &auth.UserInfo{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: firstName,
			LastName:  lastName,
			Role:      &primaryRole,
		},
		Tenant:      tenantInfo,
		Permissions: permissions,
		ExpiresAt:   nil, // TODO: Calculate actual expiration from token
	}

	h.metrics.IncrementCounter("auth_validate_total", metrics.Fields{
		"status": "success",
	})

	span.SetAttributes(
		attribute.String("result.user_id", result.User.ID),
		attribute.Bool("result.valid", result.Valid),
	)

	logger.InfoContext(ctx, "Token validation successful", logger.Fields{
		"user_id":   user.ID.String(),
		"tenant_id": user.TenantID.String(),
		"email":     user.Email,
	})

	return result, nil
}

// ─── TEMPORARY IAM ADAPTER ──────────────────────────────────────────────────

// identityToIAMAdapter temporarily adapts the identity service to IAM interface
// TODO: Remove this when full IAM service is implemented
type identityToIAMAdapter struct {
	identityService interface{}
}

// Authentication returns a temporary authn service adapter
func (a *identityToIAMAdapter) Authentication() authn.Service {
	return &identityAuthnAdapter{identityService: a.identityService}
}

// Authorization returns a temporary authz service adapter
func (a *identityToIAMAdapter) Authorization() authz.Service {
	return &identityAuthzAdapter{}
}

// Policy returns a temporary policy service adapter
func (a *identityToIAMAdapter) Policy() policy.Service {
	return &dummyPolicyService{}
}

// dummyPolicyService implements policy.Service with minimal functionality
type dummyPolicyService struct{}

// Implement all required policy service methods
func (d *dummyPolicyService) CreatePolicy(ctx context.Context, req *policy.CreatePolicyRequest) (*model.Policy, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) GetPolicy(ctx context.Context, policyID uuid.UUID) (*model.Policy, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) UpdatePolicy(ctx context.Context, req *policy.UpdatePolicyRequest) (*model.Policy, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) DeletePolicy(ctx context.Context, policyID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) ListPolicies(ctx context.Context, req *policy.ListPoliciesRequest) (*policy.ListPoliciesResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) ValidatePolicy(ctx context.Context, policy *model.Policy) error {
	return fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) EvaluatePolicy(ctx context.Context, req *policy.PolicyEvaluationRequest) (*policy.PolicyEvaluationResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) BulkEvaluatePolicies(ctx context.Context, req *policy.BulkPolicyEvaluationRequest) (*policy.BulkPolicyEvaluationResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) TestPolicy(ctx context.Context, req *policy.PolicyTestRequest) (*policy.PolicyTestResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) CreatePolicyTemplate(ctx context.Context, req *policy.CreatePolicyTemplateRequest) (*model.PolicyTemplate, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) GetPolicyTemplate(ctx context.Context, templateID uuid.UUID) (*model.PolicyTemplate, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) ListPolicyTemplates(ctx context.Context) ([]*model.PolicyTemplate, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) InstantiatePolicyFromTemplate(ctx context.Context, req *policy.InstantiatePolicyRequest) (*model.Policy, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) CreatePolicyVersion(ctx context.Context, req *policy.CreatePolicyVersionRequest) (*model.PolicyVersion, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) GetPolicyVersion(ctx context.Context, policyID uuid.UUID, version int) (*model.PolicyVersion, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) ListPolicyVersions(ctx context.Context, policyID uuid.UUID) ([]*model.PolicyVersion, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) PromotePolicyVersion(ctx context.Context, policyID uuid.UUID, version int) error {
	return fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) CreateAttribute(ctx context.Context, req *policy.CreateAttributeRequest) (*model.Attribute, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) GetAttribute(ctx context.Context, attributeID uuid.UUID) (*model.Attribute, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) UpdateAttribute(ctx context.Context, req *policy.UpdateAttributeRequest) (*model.Attribute, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) DeleteAttribute(ctx context.Context, attributeID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) ListAttributes(ctx context.Context, req *policy.ListAttributesRequest) (*policy.ListAttributesResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) CombinePolicyDecisions(ctx context.Context, decisions []*model.PolicyDecision, algorithm string) (*policy.CombinedPolicyDecision, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) AnalyzePolicyConflicts(ctx context.Context, policyIDs []uuid.UUID) (*policy.PolicyConflictAnalysis, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) GetPolicyImpactAnalysis(ctx context.Context, policyID uuid.UUID) (*policy.PolicyImpactAnalysis, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) InvalidatePolicyCache(ctx context.Context, policyIDs []uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) WarmupPolicyCache(ctx context.Context, policyIDs []uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (d *dummyPolicyService) GetPolicyCacheStats(ctx context.Context) (*policy.PolicyCacheStats, error) {
	return nil, fmt.Errorf("not implemented")
}

// identityAuthnAdapter adapts identity service to authn interface
type identityAuthnAdapter struct {
	identityService interface{}
}

// Authenticate method adapter
func (a *identityAuthnAdapter) Authenticate(ctx context.Context, req *authn.AuthenticationRequest) (*authn.AuthenticationResult, error) {
	// Convert identity service interface to concrete type
	identitySvc, ok := a.identityService.(identity.Service)
	if !ok {
		return nil, fmt.Errorf("invalid identity service type")
	}

	// Use identity service authenticate method
	user, err := identitySvc.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	// Convert identity user to model user
	modelUser := &model.User{
		ID:            user.ID,
		TenantID:      user.TenantID,
		EntityID:      user.EntityID,
		PersonID:      user.PersonID,
		Email:         user.Email,
		FirstName:     "", // Identity user doesn't have FirstName, get from Person if needed
		LastName:      "", // Identity user doesn't have LastName, get from Person if needed
		AccountStatus: model.UserAccountStatus(user.AccountStatus),
	}

	return &authn.AuthenticationResult{
		User:         modelUser,
		AccessToken:  "jwt-token-" + user.ID.String(), // Generate simple token
		RefreshToken: "",                              // Identity service doesn't provide refresh token
		ExpiresAt:    time.Now().Add(24 * time.Hour),  // Default 24 hour expiry
		MFARequired:  false,
	}, nil
}

// ValidateToken method adapter
func (a *identityAuthnAdapter) ValidateToken(ctx context.Context, token string) (*authn.TokenValidationResult, error) {
	// For mock implementation, extract user ID from token
	if len(token) > 10 && token[:10] == "jwt-token-" {
		userIDStr := token[10:]
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return &authn.TokenValidationResult{Valid: false}, nil
		}
		return &authn.TokenValidationResult{
			Valid:  true,
			UserID: userID,
		}, nil
	}
	return &authn.TokenValidationResult{Valid: false}, nil
}

// GetUser method adapter
func (a *identityAuthnAdapter) GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	identitySvc, ok := a.identityService.(identity.Service)
	if !ok {
		return nil, fmt.Errorf("invalid identity service type")
	}

	user, err := identitySvc.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:            user.ID,
		TenantID:      user.TenantID,
		EntityID:      user.EntityID,
		PersonID:      user.PersonID,
		Email:         user.Email,
		FirstName:     "", // Identity user doesn't have FirstName
		LastName:      "", // Identity user doesn't have LastName
		AccountStatus: model.UserAccountStatus(user.AccountStatus),
	}, nil
}

// GetUserRoles method adapter - returns empty for now
func (a *identityAuthnAdapter) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.Role, error) {
	// Return empty roles for now - TODO: implement when role system is ready
	return []*model.Role{}, nil
}

// GetPerson method adapter - returns nil for now
func (a *identityAuthnAdapter) GetPerson(ctx context.Context, personID uuid.UUID) (*model.Person, error) {
	// Return nil for now - TODO: implement when person system is ready
	return nil, fmt.Errorf("person service not implemented")
}

// Implement other required methods with minimal implementations
func (a *identityAuthnAdapter) CreateUser(ctx context.Context, req *authn.CreateUserRequest) (*model.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) UpdateUser(ctx context.Context, req *authn.UpdateUserRequest) (*model.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) CreatePerson(ctx context.Context, req *authn.CreatePersonRequest) (*model.Person, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) UpdatePerson(ctx context.Context, req *authn.UpdatePersonRequest) (*model.Person, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) CreateEmployee(ctx context.Context, req *authn.CreateEmployeeRequest) (*model.Employee, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) GetEmployee(ctx context.Context, employeeID uuid.UUID) (*model.Employee, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) UpdateEmployee(ctx context.Context, req *authn.UpdateEmployeeRequest) (*model.Employee, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) RefreshToken(ctx context.Context, refreshToken string) (*authn.TokenRefreshResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) Logout(ctx context.Context, userID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) ChangePassword(ctx context.Context, req *authn.ChangePasswordRequest) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) ResetPassword(ctx context.Context, req *authn.ResetPasswordRequest) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) ValidatePassword(ctx context.Context, userID uuid.UUID, password string) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) EnableMFA(ctx context.Context, req *authn.EnableMFARequest) (*authn.MFASetupResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) ValidateMFA(ctx context.Context, req *authn.ValidateMFARequest) (*authn.MFAValidationResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) GenerateMFABackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) CreateSession(ctx context.Context, req *authn.CreateSessionRequest) (*model.Session, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) GetSession(ctx context.Context, sessionID uuid.UUID) (*model.Session, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) InvalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) InvalidateAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) AssignRole(ctx context.Context, req *authn.AssignRoleRequest) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) RemoveRole(ctx context.Context, req *authn.RemoveRoleRequest) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) LockAccount(ctx context.Context, userID uuid.UUID, reason string) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) UnlockAccount(ctx context.Context, userID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthnAdapter) IsAccountLocked(ctx context.Context, userID uuid.UUID) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

// identityAuthzAdapter provides minimal authorization adapter
type identityAuthzAdapter struct{}

// GetUserEffectivePermissions returns empty permissions for now
func (a *identityAuthzAdapter) GetUserEffectivePermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (*authz.UserEffectivePermissions, error) {
	return &authz.UserEffectivePermissions{
		UserID:      userID,
		EntityID:    entityID,
		Permissions: make(map[string][]string),
		Roles:       []string{},
		Timestamp:   time.Now(),
	}, nil
}

// Implement other required methods with minimal implementations
func (a *identityAuthzAdapter) EvaluatePermission(ctx context.Context, req *authz.PermissionEvaluationRequest) (*authz.PermissionEvaluationResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) BulkEvaluatePermissions(ctx context.Context, req *authz.BulkPermissionEvaluationRequest) (*authz.BulkPermissionEvaluationResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) CalculateRoleHierarchy(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (*authz.RoleHierarchy, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) CreateAccessRequest(ctx context.Context, req *authz.CreateAccessRequestRequest) (*model.AccessRequest, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) GetAccessRequest(ctx context.Context, requestID uuid.UUID) (*model.AccessRequest, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) ProcessAccessRequest(ctx context.Context, req *authz.ProcessAccessRequestRequest) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) ListAccessRequests(ctx context.Context, req *authz.ListAccessRequestsRequest) (*authz.ListAccessRequestsResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) CreateApprovalWorkflow(ctx context.Context, req *authz.CreateApprovalWorkflowRequest) (*model.ApprovalWorkflow, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) GetApprovalWorkflow(ctx context.Context, workflowID uuid.UUID) (*model.ApprovalWorkflow, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) UpdateApprovalWorkflow(ctx context.Context, req *authz.UpdateApprovalWorkflowRequest) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) EvaluateConditionalAccess(ctx context.Context, req *authz.ConditionalAccessRequest) (*authz.ConditionalAccessResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) CreateConditionalAccessPolicy(ctx context.Context, req *authz.CreateConditionalAccessPolicyRequest) (*model.ConditionalAccessPolicy, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) UpdateConditionalAccessPolicy(ctx context.Context, req *authz.UpdateConditionalAccessPolicyRequest) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) GrantPermission(ctx context.Context, req *authz.GrantPermissionRequest) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) RevokePermission(ctx context.Context, req *authz.RevokePermissionRequest) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) ListUserPermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) ([]*model.Permission, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) GetDecisionHistory(ctx context.Context, userID uuid.UUID, limit int) ([]*authz.DecisionHistoryEntry, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) InvalidateUserCache(ctx context.Context, userID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) InvalidatePolicyCache(ctx context.Context, policyIDs []uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (a *identityAuthzAdapter) GetCacheStatistics(ctx context.Context) (*authz.CacheStatistics, error) {
	return nil, fmt.Errorf("not implemented")
}
