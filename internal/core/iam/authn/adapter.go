package authn

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// adapter implements the Service interface by wrapping the existing identity service
// This follows the adapter pattern to preserve existing mature implementations
type adapter struct {
	identityService identity.Service
	logger          logger.Logger
	metrics         metrics.MetricsProvider
	tracer          tracing.Service
}

// NewAdapterService creates a new authentication service that wraps the existing identity service
func NewAdapterService(
	identityService identity.Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) Service {
	return &adapter{
		identityService: identityService,
		logger:          logger,
		metrics:         metrics,
		tracer:          tracer,
	}
}

// User Management methods

func (a *adapter) CreateUser(ctx context.Context, req *CreateUserRequest) (*model.User, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authn.CreateUser")
	defer span.End()

	// Convert IAM request to identity request
	// Generate UUID for entity_id since it's required in identity service
	entityID := uuid.New() // TODO: This should come from tenant context or be configurable

	identityReq := &identity.CreateUserRequest{
		EntityID:              entityID,
		Username:              req.Email, // Use email as username for now
		Email:                 req.Email,
		Password:              req.Password,
		UserType:              "INTERNAL", // Default user type, should be configurable
		AccountStatus:         "ACTIVE",   // Default status
		SessionTimeoutMinutes: 30,         // Default timeout
		MfaEnabled:            false,      // Default MFA off
		UserAttributes: map[string]any{
			"first_name": req.FirstName,
			"last_name":  req.LastName,
		},
		Settings: req.Metadata,
	}

	// Call the existing mature identity service
	identityUser, err := a.identityService.RegisterNewUser(ctx, identityReq)
	if err != nil {
		a.metrics.IncrementCounter("authn.create_user.error", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Convert identity user to IAM user model
	iamUser := a.convertIdentityUserToIAMUser(identityUser)

	a.metrics.IncrementCounter("authn.create_user.success", nil)
	a.logger.Info("User created successfully", map[string]any{
		"user_id": iamUser.ID,
		"email":   iamUser.Email,
	})

	return iamUser, nil
}

func (a *adapter) GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authn.GetUser")
	defer span.End()

	identityUser, err := a.identityService.GetUserByID(ctx, userID)
	if err != nil {
		a.metrics.IncrementCounter("authn.get_user.error", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	iamUser := a.convertIdentityUserToIAMUser(identityUser)
	a.metrics.IncrementCounter("authn.get_user.success", nil)

	return iamUser, nil
}

func (a *adapter) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authn.GetUserByEmail")
	defer span.End()

	identityUser, err := a.identityService.GetUserByEmail(ctx, email)
	if err != nil {
		a.metrics.IncrementCounter("authn.get_user_by_email.error", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	iamUser := a.convertIdentityUserToIAMUser(identityUser)
	a.metrics.IncrementCounter("authn.get_user_by_email.success", nil)

	return iamUser, nil
}

func (a *adapter) UpdateUser(ctx context.Context, req *UpdateUserRequest) (*model.User, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authn.UpdateUser")
	defer span.End()

	// Convert IAM request to identity request
	identityReq := &identity.UpdateUserRequest{
		UserAttributes: map[string]any{
			"first_name": req.FirstName,
			"last_name":  req.LastName,
		},
		Settings: req.Metadata,
	}

	identityUser, err := a.identityService.UpdateUser(ctx, req.UserID, identityReq)
	if err != nil {
		a.metrics.IncrementCounter("authn.update_user.error", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	iamUser := a.convertIdentityUserToIAMUser(identityUser)
	a.metrics.IncrementCounter("authn.update_user.success", nil)

	return iamUser, nil
}

func (a *adapter) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	ctx, span := a.tracer.StartSpan(ctx, "authn.DeleteUser")
	defer span.End()

	err := a.identityService.DeleteUser(ctx, userID)
	if err != nil {
		a.metrics.IncrementCounter("authn.delete_user.error", map[string]any{"error": err.Error()})
		return fmt.Errorf("failed to delete user: %w", err)
	}

	a.metrics.IncrementCounter("authn.delete_user.success", nil)
	return nil
}

// Person Management methods

func (a *adapter) CreatePerson(ctx context.Context, req *CreatePersonRequest) (*model.Person, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authn.CreatePerson")
	defer span.End()

	// Convert IAM request to identity request
	entityID := uuid.New() // TODO: This should come from tenant context or be configurable

	identityReq := &identity.CreatePersonRequest{
		EntityID:   entityID,
		PersonType: "INDIVIDUAL", // Default person type
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		Phone:      req.PhoneNumber,
		BirthDate:  req.DateOfBirth,
		Metadata:   req.Metadata,
	}

	identityPerson, err := a.identityService.CreatePerson(ctx, identityReq)
	if err != nil {
		a.metrics.IncrementCounter("authn.create_person.error", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("failed to create person: %w", err)
	}

	iamPerson := a.convertIdentityPersonToIAMPerson(identityPerson)
	a.metrics.IncrementCounter("authn.create_person.success", nil)

	return iamPerson, nil
}

func (a *adapter) GetPerson(ctx context.Context, personID uuid.UUID) (*model.Person, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authn.GetPerson")
	defer span.End()

	identityPerson, err := a.identityService.GetPersonByID(ctx, personID)
	if err != nil {
		a.metrics.IncrementCounter("authn.get_person.error", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("failed to get person: %w", err)
	}

	iamPerson := a.convertIdentityPersonToIAMPerson(identityPerson)
	a.metrics.IncrementCounter("authn.get_person.success", nil)

	return iamPerson, nil
}

func (a *adapter) UpdatePerson(ctx context.Context, req *UpdatePersonRequest) (*model.Person, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authn.UpdatePerson")
	defer span.End()

	// For now, return not implemented as the identity service doesn't have UpdatePerson
	return nil, newNotImplementedError("UpdatePerson", "Person update not yet implemented in identity service")
}

// Employee Management methods

func (a *adapter) CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*model.Employee, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authn.CreateEmployee")
	defer span.End()

	// Convert IAM request to identity request
	entityID := uuid.New() // TODO: This should come from tenant context or be configurable

	identityReq := &identity.CreateEmployeeRequest{
		PersonID:       req.PersonID,
		EmployeeNumber: req.EmployeeNumber,
		EntityID:       entityID,
		PositionTitle:  &req.JobTitle,
		ManagerID:      req.ManagerID,
		HireDate:       req.HireDate,
		SalaryInfo: map[string]any{
			"salary": req.Salary,
		},
		Status: identity.EmploymentStatus(req.EmploymentStatus),
	}

	identityEmployee, err := a.identityService.CreateEmployee(ctx, identityReq)
	if err != nil {
		a.metrics.IncrementCounter("authn.create_employee.error", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	iamEmployee := a.convertIdentityEmployeeToIAMEmployee(identityEmployee)
	a.metrics.IncrementCounter("authn.create_employee.success", nil)

	return iamEmployee, nil
}

func (a *adapter) GetEmployee(ctx context.Context, employeeID uuid.UUID) (*model.Employee, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authn.GetEmployee")
	defer span.End()

	identityEmployee, err := a.identityService.GetEmployeeByID(ctx, employeeID)
	if err != nil {
		a.metrics.IncrementCounter("authn.get_employee.error", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	iamEmployee := a.convertIdentityEmployeeToIAMEmployee(identityEmployee)
	a.metrics.IncrementCounter("authn.get_employee.success", nil)

	return iamEmployee, nil
}

func (a *adapter) UpdateEmployee(ctx context.Context, req *UpdateEmployeeRequest) (*model.Employee, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authn.UpdateEmployee")
	defer span.End()

	// For now, return not implemented as the identity service doesn't have UpdateEmployee
	return nil, newNotImplementedError("UpdateEmployee", "Employee update not yet implemented in identity service")
}

// Authentication methods

func (a *adapter) Authenticate(ctx context.Context, req *AuthenticationRequest) (*AuthenticationResult, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authn.Authenticate")
	defer span.End()

	// Use the existing mature authentication logic
	identityUser, err := a.identityService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		a.metrics.IncrementCounter("authn.authenticate.error", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Convert to IAM user model
	iamUser := a.convertIdentityUserToIAMUser(identityUser)

	// For now, return basic authentication result
	// TODO: Implement proper JWT token generation and MFA handling
	result := &AuthenticationResult{
		User:         iamUser,
		AccessToken:  "TODO_IMPLEMENT_JWT_GENERATION",
		RefreshToken: "TODO_IMPLEMENT_REFRESH_TOKEN",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		MFARequired:  false, // TODO: Check if user has MFA enabled
	}

	a.metrics.IncrementCounter("authn.authenticate.success", nil)
	a.logger.Info("User authenticated successfully", map[string]any{
		"user_id": iamUser.ID,
		"email":   iamUser.Email,
	})

	return result, nil
}

// Password Management methods

func (a *adapter) ChangePassword(ctx context.Context, req *ChangePasswordRequest) error {
	ctx, span := a.tracer.StartSpan(ctx, "authn.ChangePassword")
	defer span.End()

	err := a.identityService.ChangePassword(ctx, req.UserID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		a.metrics.IncrementCounter("authn.change_password.error", map[string]any{"error": err.Error()})
		return fmt.Errorf("failed to change password: %w", err)
	}

	a.metrics.IncrementCounter("authn.change_password.success", nil)
	return nil
}

// Role Management methods

func (a *adapter) AssignRole(ctx context.Context, req *AssignRoleRequest) error {
	ctx, span := a.tracer.StartSpan(ctx, "authn.AssignRole")
	defer span.End()

	// Use existing role assignment logic
	entityID := uuid.Nil
	if req.EntityID != nil {
		entityID = *req.EntityID
	}

	err := a.identityService.AssignUserRole(ctx, req.UserID, req.RoleID, entityID)
	if err != nil {
		a.metrics.IncrementCounter("authn.assign_role.error", map[string]any{"error": err.Error()})
		return fmt.Errorf("failed to assign role: %w", err)
	}

	a.metrics.IncrementCounter("authn.assign_role.success", nil)
	return nil
}

func (a *adapter) RemoveRole(ctx context.Context, req *RemoveRoleRequest) error {
	ctx, span := a.tracer.StartSpan(ctx, "authn.RemoveRole")
	defer span.End()

	entityID := uuid.Nil
	if req.EntityID != nil {
		entityID = *req.EntityID
	}

	err := a.identityService.RevokeUserRole(ctx, req.UserID, req.RoleID, entityID)
	if err != nil {
		a.metrics.IncrementCounter("authn.remove_role.error", map[string]any{"error": err.Error()})
		return fmt.Errorf("failed to remove role: %w", err)
	}

	a.metrics.IncrementCounter("authn.remove_role.success", nil)
	return nil
}

func (a *adapter) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.Role, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authn.GetUserRoles")
	defer span.End()

	identityRoles, err := a.identityService.GetUserRoles(ctx, userID)
	if err != nil {
		a.metrics.IncrementCounter("authn.get_user_roles.error", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Convert identity roles to IAM roles
	iamRoles := make([]*model.Role, len(identityRoles))
	for i, identityRole := range identityRoles {
		iamRoles[i] = a.convertIdentityRoleToIAMRole(identityRole)
	}

	a.metrics.IncrementCounter("authn.get_user_roles.success", nil)
	return iamRoles, nil
}

// Stub implementations for methods not available in legacy identity service

func (a *adapter) ValidateToken(ctx context.Context, token string) (*TokenValidationResult, error) {
	return nil, newNotImplementedError("ValidateToken", "Token validation not yet implemented")
}

func (a *adapter) RefreshToken(ctx context.Context, refreshToken string) (*TokenRefreshResult, error) {
	return nil, newNotImplementedError("RefreshToken", "Token refresh not yet implemented")
}

func (a *adapter) Logout(ctx context.Context, userID uuid.UUID) error {
	return newNotImplementedError("Logout", "Logout not yet implemented")
}

func (a *adapter) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	return newNotImplementedError("ResetPassword", "Password reset not yet implemented")
}

func (a *adapter) ValidatePassword(ctx context.Context, userID uuid.UUID, password string) error {
	return newNotImplementedError("ValidatePassword", "Password validation not yet implemented")
}

func (a *adapter) EnableMFA(ctx context.Context, req *EnableMFARequest) (*MFASetupResult, error) {
	return nil, newNotImplementedError("EnableMFA", "MFA not yet implemented in identity service")
}

func (a *adapter) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	return newNotImplementedError("DisableMFA", "MFA not yet implemented in identity service")
}

func (a *adapter) ValidateMFA(ctx context.Context, req *ValidateMFARequest) (*MFAValidationResult, error) {
	return nil, newNotImplementedError("ValidateMFA", "MFA not yet implemented in identity service")
}

func (a *adapter) GenerateMFABackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return nil, newNotImplementedError("GenerateMFABackupCodes", "MFA not yet implemented in identity service")
}

func (a *adapter) CreateSession(ctx context.Context, req *CreateSessionRequest) (*model.Session, error) {
	return nil, newNotImplementedError("CreateSession", "Session management not yet implemented in identity service")
}

func (a *adapter) GetSession(ctx context.Context, sessionID uuid.UUID) (*model.Session, error) {
	return nil, newNotImplementedError("GetSession", "Session management not yet implemented in identity service")
}

func (a *adapter) InvalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	return newNotImplementedError("InvalidateSession", "Session management not yet implemented in identity service")
}

func (a *adapter) InvalidateAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	return newNotImplementedError("InvalidateAllUserSessions", "Session management not yet implemented in identity service")
}

func (a *adapter) LockAccount(ctx context.Context, userID uuid.UUID, reason string) error {
	return newNotImplementedError("LockAccount", "Account locking not yet implemented in identity service")
}

func (a *adapter) UnlockAccount(ctx context.Context, userID uuid.UUID) error {
	return newNotImplementedError("UnlockAccount", "Account unlocking not yet implemented in identity service")
}

func (a *adapter) IsAccountLocked(ctx context.Context, userID uuid.UUID) (bool, error) {
	return false, newNotImplementedError("IsAccountLocked", "Account lock status not yet implemented in identity service")
}

// Conversion helper methods

func (a *adapter) convertIdentityUserToIAMUser(identityUser *identity.User) *model.User {
	if identityUser == nil {
		return nil
	}

	// Extract first and last name from user attributes
	var firstName, lastName string
	if attrs := identityUser.UserAttributes; attrs != nil {
		if fn, ok := attrs["first_name"].(string); ok {
			firstName = fn
		}
		if ln, ok := attrs["last_name"].(string); ok {
			lastName = ln
		}
	}

	return &model.User{
		ID:               identityUser.ID,
		TenantID:         identityUser.TenantID,
		Email:            identityUser.Email,
		PasswordHash:     "", // Don't expose password hash
		FirstName:        firstName,
		LastName:         lastName,
		PhoneNumber:      nil, // Not directly available in identity user
		AccountStatus:    model.UserAccountStatus(identityUser.AccountStatus),
		EmailVerified:    true,  // Default assumption
		PhoneVerified:    false, // Default assumption
		MFAEnabled:       identityUser.MfaEnabled,
		MFAMethod:        nil, // Not available in identity service
		MFASecret:        nil, // Never expose MFA secret
		LastLoginAt:      identityUser.LastLoginAt,
		PasswordExpired:  false, // Default assumption
		FailedLoginCount: int(identityUser.FailedLoginAttempts),
		LockedUntil:      identityUser.LockoutUntil,
		Metadata:         identityUser.Settings,
		CreatedAt:        identityUser.CreatedAt,
		UpdatedAt:        identityUser.UpdatedAt,
		DeletedAt:        identityUser.DeletedAt,
	}
}

func (a *adapter) convertIdentityPersonToIAMPerson(identityPerson *identity.Person) *model.Person {
	if identityPerson == nil {
		return nil
	}

	// Convert email to pointer
	var email *string
	if identityPerson.Email != "" {
		email = &identityPerson.Email
	}

	// Convert birth date (handle potential nil)
	var birthDate time.Time
	if identityPerson.BirthDate != nil {
		birthDate = *identityPerson.BirthDate
	}

	return &model.Person{
		ID:                 identityPerson.ID,
		TenantID:           identityPerson.TenantID,
		EntityID:           identityPerson.EntityID,
		PersonType:         model.PersonType(identityPerson.PersonType),
		FirstName:          identityPerson.FirstName,
		LastName:           identityPerson.LastName,
		MiddleName:         identityPerson.MiddleName,
		Email:              email,
		PhoneNumber:        identityPerson.Phone,
		BirthDate:          birthDate,
		NationalID:         identityPerson.NationalID,
		TaxID:              identityPerson.TaxID,
		Address:            nil, // Address conversion may need additional logic for []byte to map[string]any
		SecurityAttributes: identityPerson.SecurityAttributes,
		Metadata:           identityPerson.Metadata,
		IsActive:           identityPerson.IsActive,
		CreatedAt:          identityPerson.CreatedAt,
		UpdatedAt:          identityPerson.UpdatedAt,
		DeletedAt:          identityPerson.DeletedAt,
	}
}

func (a *adapter) convertIdentityEmployeeToIAMEmployee(identityEmployee *identity.Employee) *model.Employee {
	if identityEmployee == nil {
		return nil
	}

	return &model.Employee{
		ID:               identityEmployee.ID,
		TenantID:         identityEmployee.TenantID,
		PersonID:         identityEmployee.PersonID,
		EmployeeNumber:   identityEmployee.EmployeeNumber,
		EntityID:         identityEmployee.EntityID,
		PositionTitle:    identityEmployee.PositionTitle,
		DepartmentID:     identityEmployee.DepartmentID,
		ManagerID:        identityEmployee.ManagerID,
		HireDate:         identityEmployee.HireDate,
		TerminationDate:  identityEmployee.TerminationDate,
		SalaryInfo:       identityEmployee.SalaryInfo,
		EmploymentStatus: model.EmploymentStatus(identityEmployee.Status),
		WorkSchedule:     identityEmployee.WorkSchedule,
		SecurityLevel:    int(identityEmployee.SecurityLevel),
		AccessAttributes: identityEmployee.AccessAttributes,
		CreatedAt:        identityEmployee.CreatedAt,
		UpdatedAt:        identityEmployee.UpdatedAt,
		DeletedAt:        identityEmployee.DeletedAt,
	}
}

func (a *adapter) convertIdentityRoleToIAMRole(identityRole *identity.Role) *model.Role {
	if identityRole == nil {
		return nil
	}

	return &model.Role{
		ID:          identityRole.ID,
		TenantID:    identityRole.UserID, // Note: This mapping may need adjustment based on actual identity role structure
		Name:        identityRole.Name,
		Description: "",  // Default empty description
		ParentID:    nil, // Not available in identity role model
		EntityID:    nil, // Not available in identity role model
		Metadata:    nil, // Not available in identity role model
		CreatedAt:   identityRole.AssignedAt,
		UpdatedAt:   identityRole.AssignedAt,
		DeletedAt:   nil,
	}
}

// Helper function to create not implemented errors
func newNotImplementedError(operation, details string) error {
	return errors.NewBusinessError("NOT_IMPLEMENTED", fmt.Sprintf("%s not implemented", operation)).
		WithDetail("operation", operation).
		WithDetail("details", details).
		WithHTTPStatus(501)
}
