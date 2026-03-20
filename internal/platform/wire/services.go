// Package wire - Service layer providers
package wire

import (
	"context"

	"github.com/google/uuid"
	temporalclient "go.temporal.io/sdk/client"

	"awo/internal/platform/cache"
	"awo/internal/shared/errors"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"

	// Core services
	financeService "awo/internal/core/finance/service"
	"awo/internal/core/iam"
	"awo/internal/core/iam/authn"
	"awo/internal/core/iam/authz"
	"awo/internal/core/iam/model"
	"awo/internal/core/iam/policy"
	"awo/internal/core/tenant"

	db "awo/db/sqlc"
)

// ============================================================================
// TENANT SERVICE
// ============================================================================

// NewTenantService creates a new tenant service
func NewTenantService(
	store db.Store,
	cache cache.Service,
	tracer tracing.Service,
	log logger.Logger,
) tenant.Service {
	return tenant.NewService(tenant.Dependencies{
		Store:  store,
		Cache:  cache,
		Tracer: tracer,
		Logger: log,
	})
}

// ============================================================================
// SIMPLIFIED IAM SERVICE FOR STARTUP
// ============================================================================

// NewSimpleAuthenticationService creates a minimal authentication service for startup
func NewSimpleAuthenticationService() authn.Service {
	// Return a mock/minimal implementation for now to get the app running
	// This should be replaced with proper implementation later
	return &simpleAuthService{}
}

// simpleAuthService is a minimal implementation just to satisfy the interface
type simpleAuthService struct{}

// Implement all required authn.Service methods minimally (just return errors for now)
func (s *simpleAuthService) CreateUser(ctx context.Context, req *authn.CreateUserRequest) (*model.User, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "User creation not implemented yet")
}

func (s *simpleAuthService) GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "Get user not implemented yet")
}

func (s *simpleAuthService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "Get user by email not implemented yet")
}

func (s *simpleAuthService) UpdateUser(ctx context.Context, req *authn.UpdateUserRequest) (*model.User, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "Update user not implemented yet")
}

func (s *simpleAuthService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return errors.NewBusinessError("NOT_IMPLEMENTED", "Delete user not implemented yet")
}

func (s *simpleAuthService) CreatePerson(ctx context.Context, req *authn.CreatePersonRequest) (*model.Person, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "Create person not implemented yet")
}

func (s *simpleAuthService) GetPerson(ctx context.Context, personID uuid.UUID) (*model.Person, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "Get person not implemented yet")
}

func (s *simpleAuthService) UpdatePerson(ctx context.Context, req *authn.UpdatePersonRequest) (*model.Person, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "Update person not implemented yet")
}

func (s *simpleAuthService) CreateEmployee(ctx context.Context, req *authn.CreateEmployeeRequest) (*model.Employee, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "Create employee not implemented yet")
}

func (s *simpleAuthService) GetEmployee(ctx context.Context, employeeID uuid.UUID) (*model.Employee, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "Get employee not implemented yet")
}

func (s *simpleAuthService) UpdateEmployee(ctx context.Context, req *authn.UpdateEmployeeRequest) (*model.Employee, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "Update employee not implemented yet")
}

func (s *simpleAuthService) Authenticate(ctx context.Context, req *authn.AuthenticationRequest) (*authn.AuthenticationResult, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "Authentication not implemented yet")
}

func (s *simpleAuthService) ValidateToken(ctx context.Context, token string) (*authn.TokenValidationResult, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "Token validation not implemented yet")
}

func (s *simpleAuthService) RefreshToken(ctx context.Context, refreshToken string) (*authn.TokenRefreshResult, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "Token refresh not implemented yet")
}

func (s *simpleAuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return errors.NewBusinessError("NOT_IMPLEMENTED", "Logout not implemented yet")
}

func (s *simpleAuthService) ChangePassword(ctx context.Context, req *authn.ChangePasswordRequest) error {
	return errors.NewBusinessError("NOT_IMPLEMENTED", "Change password not implemented yet")
}

func (s *simpleAuthService) ResetPassword(ctx context.Context, req *authn.ResetPasswordRequest) error {
	return errors.NewBusinessError("NOT_IMPLEMENTED", "Reset password not implemented yet")
}

func (s *simpleAuthService) ValidatePassword(ctx context.Context, userID uuid.UUID, password string) error {
	return errors.NewBusinessError("NOT_IMPLEMENTED", "Password validation not implemented yet")
}

func (s *simpleAuthService) EnableMFA(ctx context.Context, req *authn.EnableMFARequest) (*authn.MFASetupResult, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "MFA not implemented yet")
}

func (s *simpleAuthService) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	return errors.NewBusinessError("NOT_IMPLEMENTED", "MFA not implemented yet")
}

func (s *simpleAuthService) ValidateMFA(ctx context.Context, req *authn.ValidateMFARequest) (*authn.MFAValidationResult, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "MFA not implemented yet")
}

func (s *simpleAuthService) GenerateMFABackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "MFA not implemented yet")
}

// Session Management methods (missing from interface)
func (s *simpleAuthService) CreateSession(ctx context.Context, req *authn.CreateSessionRequest) (*model.Session, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "Create session not implemented yet")
}

func (s *simpleAuthService) GetSession(ctx context.Context, sessionID uuid.UUID) (*model.Session, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "Get session not implemented yet")
}

func (s *simpleAuthService) InvalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	return errors.NewBusinessError("NOT_IMPLEMENTED", "Invalidate session not implemented yet")
}

func (s *simpleAuthService) InvalidateAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	return errors.NewBusinessError("NOT_IMPLEMENTED", "Invalidate all user sessions not implemented yet")
}

// Role Management methods (missing from interface)
func (s *simpleAuthService) AssignRole(ctx context.Context, req *authn.AssignRoleRequest) error {
	return errors.NewBusinessError("NOT_IMPLEMENTED", "Assign role not implemented yet")
}

func (s *simpleAuthService) RemoveRole(ctx context.Context, req *authn.RemoveRoleRequest) error {
	return errors.NewBusinessError("NOT_IMPLEMENTED", "Remove role not implemented yet")
}

func (s *simpleAuthService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.Role, error) {
	return nil, errors.NewBusinessError("NOT_IMPLEMENTED", "Get user roles not implemented yet")
}

// Account Management methods (missing from interface)
func (s *simpleAuthService) LockAccount(ctx context.Context, userID uuid.UUID, reason string) error {
	return errors.NewBusinessError("NOT_IMPLEMENTED", "Lock account not implemented yet")
}

func (s *simpleAuthService) UnlockAccount(ctx context.Context, userID uuid.UUID) error {
	return errors.NewBusinessError("NOT_IMPLEMENTED", "Unlock account not implemented yet")
}

func (s *simpleAuthService) IsAccountLocked(ctx context.Context, userID uuid.UUID) (bool, error) {
	return false, errors.NewBusinessError("NOT_IMPLEMENTED", "Is account locked not implemented yet")
}

// NewSimpleIAMService creates a minimal IAM service
func NewSimpleIAMService(authnService authn.Service) iam.Service {
	return &simpleIAMService{authService: authnService}
}

type simpleIAMService struct {
	authService authn.Service
}

func (s *simpleIAMService) Authentication() authn.Service {
	return s.authService
}

func (s *simpleIAMService) Authorization() authz.Service {
	return nil // Not implemented for startup
}

func (s *simpleIAMService) Policy() policy.Service {
	return nil // Not implemented for startup
}

// ============================================================================
// TENANT TEMPORAL INTEGRATION
// ============================================================================

// NewTenantTemporalIntegration creates the tenant Temporal integration
func NewTenantTemporalIntegration(
	tenantService tenant.Service,
	temporalClient temporalclient.Client,
	log logger.Logger,
) (*tenant.TemporalIntegration, error) {
	return tenant.NewTemporalIntegration(tenant.TemporalConfig{
		Service:        tenantService,
		TemporalClient: temporalClient,
		Logger:         log,
	})
}

// ============================================================================
// FINANCE SERVICES
// ============================================================================

// NewFinanceServices creates finance service collection
func NewFinanceServices(
	store db.Store,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) *financeService.Services {
	// For now, return nil to bypass finance services during startup
	// TODO: Implement proper finance services after getting basic app running
	return nil
}
