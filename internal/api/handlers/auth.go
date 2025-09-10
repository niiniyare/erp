package handlers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/api/gen/auth"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"goa.design/goa/v3/security"
)

// AuthHandler implements the GOA auth service following the data flow pattern
type AuthHandler struct {
	userService identity.Service
	tracing     tracing.TracingService
	metrics     *metrics.MetricsService
}

// NewAuthHandler creates a new auth handler following Clean Architecture pattern
func NewAuthHandler(userSvc identity.Service, tracing tracing.TracingService, metrics *metrics.MetricsService) auth.Service {
	return &AuthHandler{
		userService: userSvc,
		tracing:     tracing,
		metrics:     metrics,
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

	// TODO: Implement JWT token validation using existing user service
	// This should integrate with h.userService to validate tokens
	logger.Info("JWT Auth called", logger.Fields{"token_prefix": token[:min(len(token), 10)]})

	h.metrics.IncrementCounter("auth_jwt_validations_total", metrics.Fields{})
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

	// Authenticate user with identity service
	user, err := h.userService.Authenticate(ctx, p.Email, p.Password)
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

	// Check if account is active
	if !user.IsActive || user.AccountStatus != identity.AccountStatusActive {
		h.metrics.IncrementCounter("auth_login_failed_total", metrics.Fields{
			"reason": "account_inactive",
		})

		logger.WarnContext(ctx, "Login attempt for inactive account", logger.Fields{
			"user_id":        user.ID.String(),
			"email":          user.Email,
			"account_status": string(user.AccountStatus),
			"is_active":      user.IsActive,
		})

		return nil, auth.MakeUnauthorized(fmt.Errorf("account is not active"))
	}

	// TODO: Generate actual JWT tokens - for now using mock tokens
	// This should integrate with a JWT service to generate real tokens
	accessToken := "jwt-token-" + user.ID.String()
	refreshToken := "refresh-token-" + user.ID.String()

	// Convert domain user to GOA response
	result := &auth.Auth{
		AccessToken:  accessToken,
		RefreshToken: &refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		User: &auth.UserInfo{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: "", // Will be populated from Person if available
			LastName:  "", // Will be populated from Person if available
			Role:      &user.UserType,
		},
		Tenant: &auth.TenantInfo{
			ID:        user.TenantID.String(),
			Name:      "Default Tenant", // TODO: Get actual tenant info
			Subdomain: "default",        // TODO: Get actual subdomain
			Status:    "active",         // TODO: Get actual status
		},
		Permissions: []string{}, // TODO: Get user permissions
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

	// Get user from identity service
	user, err := h.userService.GetUserByID(ctx, userUUID)
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
	if !user.IsActive || user.AccountStatus != identity.AccountStatusActive {
		h.metrics.IncrementCounter("auth_validate_failed_total", metrics.Fields{
			"reason": "account_inactive",
		})

		logger.WarnContext(ctx, "Token validation failed - account inactive", logger.Fields{
			"user_id":        user.ID.String(),
			"account_status": string(user.AccountStatus),
			"is_active":      user.IsActive,
		})

		return &auth.TokenValidation{
			Valid: false,
		}, nil
	}

	// Token is valid, return user info
	result := &auth.TokenValidation{
		Valid: true,
		User: &auth.UserInfo{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: "", // Will be populated from Person if available
			LastName:  "", // Will be populated from Person if available
			Role:      &user.UserType,
		},
		Tenant: &auth.TenantInfo{
			ID:        user.TenantID.String(),
			Name:      "Default Tenant", // TODO: Get actual tenant info
			Subdomain: "default",        // TODO: Get actual subdomain
			Status:    "active",         // TODO: Get actual status
		},
		Permissions: []string{}, // TODO: Get user permissions
		ExpiresAt:   nil,        // TODO: Calculate actual expiration from token
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
