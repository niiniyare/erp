package authn

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/iam/model"
)

// Service defines the authentication service interface
// Handles login, password management, MFA, and user identity
type Service interface {
	// User Management
	CreateUser(ctx context.Context, req *CreateUserRequest) (*model.User, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, req *UpdateUserRequest) (*model.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error

	// Person Management
	CreatePerson(ctx context.Context, req *CreatePersonRequest) (*model.Person, error)
	GetPerson(ctx context.Context, personID uuid.UUID) (*model.Person, error)
	UpdatePerson(ctx context.Context, req *UpdatePersonRequest) (*model.Person, error)

	// Employee Management
	CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*model.Employee, error)
	GetEmployee(ctx context.Context, employeeID uuid.UUID) (*model.Employee, error)
	UpdateEmployee(ctx context.Context, req *UpdateEmployeeRequest) (*model.Employee, error)

	// Authentication
	Authenticate(ctx context.Context, req *AuthenticationRequest) (*AuthenticationResult, error)
	ValidateToken(ctx context.Context, token string) (*TokenValidationResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenRefreshResult, error)
	Logout(ctx context.Context, userID uuid.UUID) error

	// Password Management
	ChangePassword(ctx context.Context, req *ChangePasswordRequest) error
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error
	ValidatePassword(ctx context.Context, userID uuid.UUID, password string) error

	// Multi-Factor Authentication (MFA)
	EnableMFA(ctx context.Context, req *EnableMFARequest) (*MFASetupResult, error)
	DisableMFA(ctx context.Context, userID uuid.UUID) error
	ValidateMFA(ctx context.Context, req *ValidateMFARequest) (*MFAValidationResult, error)
	GenerateMFABackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error)

	// Session Management
	CreateSession(ctx context.Context, req *CreateSessionRequest) (*model.Session, error)
	GetSession(ctx context.Context, sessionID uuid.UUID) (*model.Session, error)
	InvalidateSession(ctx context.Context, sessionID uuid.UUID) error
	InvalidateAllUserSessions(ctx context.Context, userID uuid.UUID) error

	// Role Management
	AssignRole(ctx context.Context, req *AssignRoleRequest) error
	RemoveRole(ctx context.Context, req *RemoveRoleRequest) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.Role, error)

	// Account Management
	LockAccount(ctx context.Context, userID uuid.UUID, reason string) error
	UnlockAccount(ctx context.Context, userID uuid.UUID) error
	IsAccountLocked(ctx context.Context, userID uuid.UUID) (bool, error)
}

// Request/Response types
type CreateUserRequest struct {
	Email       string                 `json:"email" validate:"required,email"`
	Password    string                 `json:"password" validate:"required,min=8"`
	FirstName   string                 `json:"first_name" validate:"required"`
	LastName    string                 `json:"last_name" validate:"required"`
	PhoneNumber *string                `json:"phone_number,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateUserRequest struct {
	UserID      uuid.UUID              `json:"user_id" validate:"required"`
	FirstName   *string                `json:"first_name,omitempty"`
	LastName    *string                `json:"last_name,omitempty"`
	PhoneNumber *string                `json:"phone_number,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type CreatePersonRequest struct {
	FirstName   string                 `json:"first_name" validate:"required"`
	LastName    string                 `json:"last_name" validate:"required"`
	Email       *string                `json:"email,omitempty" validate:"omitempty,email"`
	PhoneNumber *string                `json:"phone_number,omitempty"`
	DateOfBirth *time.Time             `json:"date_of_birth,omitempty"`
	Address     *model.Address         `json:"address,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type UpdatePersonRequest struct {
	PersonID    uuid.UUID              `json:"person_id" validate:"required"`
	FirstName   *string                `json:"first_name,omitempty"`
	LastName    *string                `json:"last_name,omitempty"`
	Email       *string                `json:"email,omitempty" validate:"omitempty,email"`
	PhoneNumber *string                `json:"phone_number,omitempty"`
	DateOfBirth *time.Time             `json:"date_of_birth,omitempty"`
	Address     *model.Address         `json:"address,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type CreateEmployeeRequest struct {
	PersonID         uuid.UUID  `json:"person_id" validate:"required"`
	EmployeeNumber   string     `json:"employee_number" validate:"required"`
	JobTitle         string     `json:"job_title" validate:"required"`
	Department       string     `json:"department" validate:"required"`
	HireDate         time.Time  `json:"hire_date" validate:"required"`
	ManagerID        *uuid.UUID `json:"manager_id,omitempty"`
	Salary           *float64   `json:"salary,omitempty"`
	EmploymentStatus string     `json:"employment_status" validate:"required"`
}

type UpdateEmployeeRequest struct {
	EmployeeID       uuid.UUID  `json:"employee_id" validate:"required"`
	JobTitle         *string    `json:"job_title,omitempty"`
	Department       *string    `json:"department,omitempty"`
	ManagerID        *uuid.UUID `json:"manager_id,omitempty"`
	Salary           *float64   `json:"salary,omitempty"`
	EmploymentStatus *string    `json:"employment_status,omitempty"`
}

type AuthenticationRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	MFACode  string `json:"mfa_code,omitempty"`
}

type AuthenticationResult struct {
	User         *model.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresAt    time.Time   `json:"expires_at"`
	MFARequired  bool        `json:"mfa_required"`
}

type TokenValidationResult struct {
	Valid  bool                   `json:"valid"`
	UserID uuid.UUID              `json:"user_id,omitempty"`
	Claims map[string]interface{} `json:"claims,omitempty"`
}

type TokenRefreshResult struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type ChangePasswordRequest struct {
	UserID          uuid.UUID `json:"user_id" validate:"required"`
	CurrentPassword string    `json:"current_password" validate:"required"`
	NewPassword     string    `json:"new_password" validate:"required,min=8"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	ResetToken  string `json:"reset_token,omitempty"`
	NewPassword string `json:"new_password,omitempty" validate:"omitempty,min=8"`
}

type EnableMFARequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Method string    `json:"method" validate:"required,oneof=totp sms email"`
}

type MFASetupResult struct {
	Secret      string   `json:"secret,omitempty"`
	QRCode      string   `json:"qr_code,omitempty"`
	BackupCodes []string `json:"backup_codes"`
}

type ValidateMFARequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Code   string    `json:"code" validate:"required"`
}

type MFAValidationResult struct {
	Valid bool `json:"valid"`
}

type CreateSessionRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	ExpiresAt time.Time `json:"expires_at"`
}

type AssignRoleRequest struct {
	UserID   uuid.UUID  `json:"user_id" validate:"required"`
	RoleID   uuid.UUID  `json:"role_id" validate:"required"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"`
}

type RemoveRoleRequest struct {
	UserID   uuid.UUID  `json:"user_id" validate:"required"`
	RoleID   uuid.UUID  `json:"role_id" validate:"required"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"`
}
