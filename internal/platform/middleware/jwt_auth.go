package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/iam/authn"
	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"goa.design/goa/v3/security"
)

// JWTAuthMiddleware provides JWT authentication middleware for GOA
type JWTAuthMiddleware struct {
	iamService iam.Service
	logger     logger.Logger
	metrics    metrics.MetricsProvider
	tracer     tracing.TracingService
}

// NewJWTAuthMiddleware creates a new JWT authentication middleware
func NewJWTAuthMiddleware(
	iamService iam.Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) *JWTAuthMiddleware {
	return &JWTAuthMiddleware{
		iamService: iamService,
		logger:     logger,
		metrics:    metrics,
		tracer:     tracer,
	}
}

// JWTAuth creates a JWT authentication security middleware for GOA
func (m *JWTAuthMiddleware) JWTAuth(ctx context.Context, token string, scheme *security.JWTScheme) (context.Context, error) {
	ctx, span := m.tracer.StartSpan(ctx, "middleware.jwt_auth",
		tracing.WithAttributes(
			attribute.String("auth.type", "jwt"),
			attribute.String("token.scheme", scheme.Name),
		))
	defer span.End()

	start := time.Now()
	defer func() {
		m.recordAuthMetrics(ctx, "jwt", time.Since(start))
	}()

	// Extract token from Authorization header
	if token == "" {
		m.logger.WarnContext(ctx, "JWT token missing in request")
		span.SetStatus(codes.Error, "jwt_token_missing")
		m.recordAuthFailure(ctx, "jwt", "token_missing")
		return ctx, fmt.Errorf("authorization token required")
	}

	// Remove Bearer prefix if present
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)

	// Validate token using IAM authentication service
	result, err := m.iamService.Authentication().ValidateToken(ctx, token)
	if err != nil {
		m.logger.WarnContext(ctx, "JWT token validation failed", logger.Fields{
			"error": err.Error(),
		})
		span.SetStatus(codes.Error, "jwt_validation_failed")
		span.RecordError(err)
		m.recordAuthFailure(ctx, "jwt", "validation_failed")
		return ctx, fmt.Errorf("invalid token: %w", err)
	}

	if !result.Valid {
		m.logger.WarnContext(ctx, "JWT token invalid")
		span.SetStatus(codes.Error, "jwt_token_invalid")
		m.recordAuthFailure(ctx, "jwt", "token_invalid")
		return ctx, fmt.Errorf("invalid token")
	}

	// Get user details for context enrichment
	user, err := m.iamService.Authentication().GetUser(ctx, result.UserID)
	if err != nil {
		m.logger.WarnContext(ctx, "Failed to get user details", logger.Fields{
			"user_id": result.UserID.String(),
			"error":   err.Error(),
		})
		span.SetStatus(codes.Error, "user_lookup_failed")
		span.RecordError(err)
		m.recordAuthFailure(ctx, "jwt", "user_lookup_failed")
		return ctx, fmt.Errorf("user lookup failed: %w", err)
	}

	// Check if account is locked
	locked, err := m.iamService.Authentication().IsAccountLocked(ctx, user.ID)
	if err != nil {
		m.logger.ErrorContext(ctx, "Failed to check account lock status", logger.Fields{
			"user_id": user.ID.String(),
			"error":   err.Error(),
		})
		span.RecordError(err)
	}
	if locked {
		m.logger.WarnContext(ctx, "Account is locked", logger.Fields{
			"user_id": user.ID.String(),
		})
		span.SetStatus(codes.Error, "account_locked")
		m.recordAuthFailure(ctx, "jwt", "account_locked")
		return ctx, fmt.Errorf("account is locked")
	}

	// Enrich context with authentication information
	ctx = m.enrichContextWithAuth(ctx, user, result.Claims, token)

	// Log successful authentication
	m.logger.InfoContext(ctx, "JWT authentication successful", logger.Fields{
		"user_id":       user.ID.String(),
		"email":         user.Email,
		"authenticated": true,
	})

	span.SetAttributes(
		attribute.String("user.id", user.ID.String()),
		attribute.String("user.email", user.Email),
		attribute.Bool("auth.success", true),
	)

	m.recordAuthSuccess(ctx, "jwt")
	return ctx, nil
}

// BasicAuth creates a basic authentication security middleware for GOA
// Note: This is a placeholder implementation since BasicAuth is not currently
// used in the GOA design, only JWT authentication is implemented
func (m *JWTAuthMiddleware) BasicAuth(ctx context.Context, username, password string, scheme any) (context.Context, error) {
	ctx, span := m.tracer.StartSpan(ctx, "middleware.basic_auth",
		tracing.WithAttributes(
			attribute.String("auth.type", "basic"),
			attribute.String("username", username),
		))
	defer span.End()

	start := time.Now()
	defer func() {
		m.recordAuthMetrics(ctx, "basic", time.Since(start))
	}()

	// Authenticate using IAM service
	authReq := &authn.AuthenticationRequest{
		Email:    username,
		Password: password,
	}

	result, err := m.iamService.Authentication().Authenticate(ctx, authReq)
	if err != nil {
		m.logger.WarnContext(ctx, "Basic authentication failed", logger.Fields{
			"username": username,
			"error":    err.Error(),
		})
		span.SetStatus(codes.Error, "basic_auth_failed")
		span.RecordError(err)
		m.recordAuthFailure(ctx, "basic", "authentication_failed")
		return ctx, fmt.Errorf("authentication failed: %w", err)
	}

	// Enrich context with authentication information
	ctx = m.enrichContextWithAuth(ctx, result.User, map[string]any{"method": "basic"}, "")

	// Log successful authentication
	m.logger.InfoContext(ctx, "Basic authentication successful", logger.Fields{
		"user_id":       result.User.ID.String(),
		"username":      username,
		"authenticated": true,
	})

	span.SetAttributes(
		attribute.String("user.id", result.User.ID.String()),
		attribute.String("user.email", result.User.Email),
		attribute.Bool("auth.success", true),
	)

	m.recordAuthSuccess(ctx, "basic")
	return ctx, nil
}

// APIKeyAuth creates an API key authentication security middleware for GOA
// Note: This is a placeholder implementation since APIKeyAuth is not currently
// used in the GOA design, only JWT authentication is implemented
func (m *JWTAuthMiddleware) APIKeyAuth(ctx context.Context, key string, scheme any) (context.Context, error) {
	ctx, span := m.tracer.StartSpan(ctx, "middleware.api_key_auth",
		tracing.WithAttributes(
			attribute.String("auth.type", "api_key"),
		))
	defer span.End()

	start := time.Now()
	defer func() {
		m.recordAuthMetrics(ctx, "api_key", time.Since(start))
	}()

	// TODO: Implement API key validation
	// This would typically involve:
	// 1. Looking up the API key in the database
	// 2. Validating it's not expired
	// 3. Getting associated service account or user
	// 4. Enriching context with service account info

	// For now, reject all API key requests
	m.logger.WarnContext(ctx, "API key authentication not implemented", logger.Fields{
		"key_provided": key != "",
	})
	span.SetStatus(codes.Error, "api_key_not_implemented")
	m.recordAuthFailure(ctx, "api_key", "not_implemented")

	return ctx, fmt.Errorf("API key authentication not implemented")
}

// Helper methods

// enrichContextWithAuth adds authentication information to the request context
func (m *JWTAuthMiddleware) enrichContextWithAuth(ctx context.Context, user *model.User, claims map[string]any, token string) context.Context {
	// Add standard authentication context
	ctx = context.WithValue(ctx, "authenticated", true)
	ctx = context.WithValue(ctx, "user_id", user.ID)
	ctx = context.WithValue(ctx, "user", user)
	ctx = context.WithValue(ctx, "user_email", user.Email)
	ctx = context.WithValue(ctx, "auth_claims", claims)

	if token != "" {
		ctx = context.WithValue(ctx, "auth_token", token)
	}

	// Add authentication timestamp
	ctx = context.WithValue(ctx, "auth_time", time.Now())

	return ctx
}

// recordAuthMetrics records authentication metrics
func (m *JWTAuthMiddleware) recordAuthMetrics(ctx context.Context, method string, duration time.Duration) {
	labels := metrics.Fields{
		"method": method,
	}

	m.metrics.ObserveHistogram("auth_request_duration", duration.Seconds(), labels)
	m.metrics.IncrementCounter("auth_requests_total", labels)
}

// recordAuthSuccess records successful authentication metrics
func (m *JWTAuthMiddleware) recordAuthSuccess(ctx context.Context, method string) {
	labels := metrics.Fields{
		"method": method,
		"result": "success",
	}

	m.metrics.IncrementCounter("auth_results_total", labels)
}

// recordAuthFailure records failed authentication metrics
func (m *JWTAuthMiddleware) recordAuthFailure(ctx context.Context, method, reason string) {
	labels := metrics.Fields{
		"method": method,
		"result": "failure",
		"reason": reason,
	}

	m.metrics.IncrementCounter("auth_results_total", labels)
	m.metrics.IncrementCounter("auth_failures_total", labels)
}

// AuthContext holds authentication context information
type AuthContext struct {
	Authenticated bool
	UserID        string
	Email         string
	Claims        map[string]any
	Token         string
	AuthTime      time.Time
}

// GetAuthContext extracts authentication context from request context
func GetAuthContext(ctx context.Context) *AuthContext {
	authCtx := &AuthContext{}

	if authenticated, ok := ctx.Value("authenticated").(bool); ok {
		authCtx.Authenticated = authenticated
	}

	if userID, ok := ctx.Value("user_id").(string); ok {
		authCtx.UserID = userID
	}

	if email, ok := ctx.Value("user_email").(string); ok {
		authCtx.Email = email
	}

	if claims, ok := ctx.Value("auth_claims").(map[string]any); ok {
		authCtx.Claims = claims
	}

	if token, ok := ctx.Value("auth_token").(string); ok {
		authCtx.Token = token
	}

	if authTime, ok := ctx.Value("auth_time").(time.Time); ok {
		authCtx.AuthTime = authTime
	}

	return authCtx
}

// RequireAuthentication is a helper to check if request is authenticated
func RequireAuthentication(ctx context.Context) error {
	if authenticated, ok := ctx.Value("authenticated").(bool); !ok || !authenticated {
		return fmt.Errorf("authentication required")
	}
	return nil
}

// HTTPJWTMiddleware creates an HTTP middleware that extracts JWT from Authorization header
func (m *JWTAuthMiddleware) HTTPJWTMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// Let GOA security middleware handle the missing token
				next.ServeHTTP(w, r)
				return
			}

			// Remove Bearer prefix and trim whitespace
			token := strings.TrimPrefix(authHeader, "Bearer ")
			token = strings.TrimSpace(token)

			if token != authHeader {
				// Token was prefixed with Bearer, add it to context for GOA
				ctx := context.WithValue(r.Context(), "jwt_token", token)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}
