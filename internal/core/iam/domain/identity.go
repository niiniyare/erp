// Package domain contains pure IAM domain models and business rules.
// No external framework dependencies — only stdlib and uuid.
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// ─── Account Status ───────────────────────────────────────────────────────────

type AccountStatus string

const (
	AccountStatusActive    AccountStatus = "ACTIVE"
	AccountStatusInactive  AccountStatus = "INACTIVE"
	AccountStatusLocked    AccountStatus = "LOCKED"
	AccountStatusSuspended AccountStatus = "SUSPENDED"
)

var validAccountStatuses = map[AccountStatus]struct{}{
	AccountStatusActive:    {},
	AccountStatusInactive:  {},
	AccountStatusLocked:    {},
	AccountStatusSuspended: {},
}

func (a AccountStatus) IsValid() bool {
	_, ok := validAccountStatuses[a]
	return ok
}

func (a AccountStatus) String() string { return string(a) }

func AllAccountStatuses() []AccountStatus {
	return []AccountStatus{
		AccountStatusActive,
		AccountStatusInactive,
		AccountStatusLocked,
		AccountStatusSuspended,
	}
}

// ─── Employment Status ────────────────────────────────────────────────────────

type EmploymentStatus string

const (
	EmploymentStatusActive     EmploymentStatus = "ACTIVE"
	EmploymentStatusInactive   EmploymentStatus = "INACTIVE"
	EmploymentStatusTerminated EmploymentStatus = "TERMINATED"
	EmploymentStatusOnLeave    EmploymentStatus = "ON_LEAVE"
	EmploymentStatusSuspended  EmploymentStatus = "SUSPENDED"
)

var validEmploymentStatuses = map[EmploymentStatus]struct{}{
	EmploymentStatusActive:     {},
	EmploymentStatusInactive:   {},
	EmploymentStatusTerminated: {},
	EmploymentStatusOnLeave:    {},
	EmploymentStatusSuspended:  {},
}

func (e EmploymentStatus) IsValid() bool {
	_, ok := validEmploymentStatuses[e]
	return ok
}

func (e EmploymentStatus) String() string { return string(e) }

func AllEmploymentStatuses() []EmploymentStatus {
	return []EmploymentStatus{
		EmploymentStatusActive,
		EmploymentStatusInactive,
		EmploymentStatusTerminated,
		EmploymentStatusOnLeave,
		EmploymentStatusSuspended,
	}
}

// ─── Aggregate Roots ─────────────────────────────────────────────────────────

// User is the IAM aggregate root for identity.
type User struct {
	ID                    uuid.UUID      `json:"id"`
	TenantID              uuid.UUID      `json:"tenant_id"`
	EntityID              uuid.UUID      `json:"entity_id"`
	PersonID              *uuid.UUID     `json:"person_id,omitempty"`
	EmployeeID            *uuid.UUID     `json:"employee_id,omitempty"`
	PrincipalID           *uuid.UUID     `json:"principal_id,omitempty"`
	Username              string         `json:"username"`
	Email                 string         `json:"email"`
	DisplayName           *string        `json:"display_name,omitempty"`
	UserType              string         `json:"user_type"`
	AccountStatus         AccountStatus  `json:"account_status"`
	IsActive              bool           `json:"is_active"`
	LastLoginAt           *time.Time     `json:"last_login_at,omitempty"`
	PasswordChangedAt     *time.Time     `json:"password_changed_at,omitempty"`
	FailedLoginAttempts   int32          `json:"failed_login_attempts"`
	LockoutUntil          *time.Time     `json:"lockout_until,omitempty"`
	SessionTimeoutMinutes int32          `json:"session_timeout_minutes"`
	MfaEnabled            bool           `json:"mfa_enabled"`
	UserAttributes        map[string]any `json:"user_attributes,omitempty"`
	Settings              map[string]any `json:"settings,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             *time.Time     `json:"deleted_at,omitempty"`
}

// IsLocked reports whether the account is currently under lockout.
func (u *User) IsLocked() bool {
	return u.LockoutUntil != nil && !u.LockoutUntil.IsZero() && u.LockoutUntil.After(time.Now())
}

// Person represents a natural person attached to a user or entity.
type Person struct {
	ID                 uuid.UUID      `json:"id"`
	TenantID           uuid.UUID      `json:"tenant_id"`
	EntityID           uuid.UUID      `json:"entity_id"`
	PersonType         string         `json:"person_type"`
	FirstName          string         `json:"first_name"`
	LastName           string         `json:"last_name"`
	MiddleName         *string        `json:"middle_name,omitempty"`
	Email              string         `json:"email,omitempty"`
	Phone              *string        `json:"phone,omitempty"`
	BirthDate          *time.Time     `json:"birth_date,omitempty"`
	NationalID         *string        `json:"national_id,omitempty"`
	TaxID              *string        `json:"tax_id,omitempty"`
	Address            []byte         `json:"address,omitempty"`
	SecurityAttributes map[string]any `json:"security_attributes,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	IsActive           bool           `json:"is_active"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          *time.Time     `json:"deleted_at,omitempty"`
}

// GetFullName returns the full name, including optional middle name.
func (p *Person) GetFullName() string {
	if p.MiddleName != nil && *p.MiddleName != "" {
		return p.FirstName + " " + *p.MiddleName + " " + p.LastName
	}
	return p.FirstName + " " + p.LastName
}

// Employee represents an employee record linked to a Person.
type Employee struct {
	ID               uuid.UUID        `json:"id"`
	TenantID         uuid.UUID        `json:"tenant_id"`
	PersonID         uuid.UUID        `json:"person_id"`
	EmployeeNumber   string           `json:"employee_number"`
	EntityID         uuid.UUID        `json:"entity_id"`
	PositionTitle    *string          `json:"position_title,omitempty"`
	DepartmentID     *uuid.UUID       `json:"department_id,omitempty"`
	ManagerID        *uuid.UUID       `json:"manager_id,omitempty"`
	HireDate         time.Time        `json:"hire_date"`
	TerminationDate  *time.Time       `json:"termination_date,omitempty"`
	SalaryInfo       map[string]any   `json:"salary_info,omitempty"`
	Status           EmploymentStatus `json:"employment_status"`
	WorkSchedule     map[string]any   `json:"work_schedule,omitempty"`
	SecurityLevel    int32            `json:"security_level"`
	AccessAttributes map[string]any   `json:"access_attributes,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        *time.Time       `json:"deleted_at,omitempty"`
}

// UserWithDetails embeds User with optional Person and Employee records.
type UserWithDetails struct {
	User     `json:",inline"`
	Person   *Person   `json:"person,omitempty"`
	Employee *Employee `json:"employee,omitempty"`
}

// UserRole represents a role assignment record for a user.
type UserRole struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	UserID         uuid.UUID  `json:"user_id"`
	RoleID         uuid.UUID  `json:"role_id"`
	EntityID       uuid.UUID  `json:"entity_id"`
	AssignmentType string     `json:"assignment_type"`
	AssignedAt     time.Time  `json:"assigned_at"`
	IsActive       bool       `json:"is_active"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// ─── Request / Command Objects ─────────────────────────────────────────────

type CreateUserRequest struct {
	EntityID              uuid.UUID      `json:"entity_id"              validate:"required"`
	PersonID              *uuid.UUID     `json:"person_id,omitempty"`
	EmployeeID            *uuid.UUID     `json:"employee_id,omitempty"`
	Username              string         `json:"username"               validate:"required,min=3,max=50"`
	Email                 string         `json:"email"                  validate:"required,email"`
	DisplayName           *string        `json:"display_name,omitempty"`
	Password              string         `json:"password"               validate:"required,min=8"`
	UserType              string         `json:"user_type"              validate:"required"`
	AccountStatus         string         `json:"account_status"`
	SessionTimeoutMinutes int32          `json:"session_timeout_minutes"`
	MfaEnabled            bool           `json:"mfa_enabled"`
	UserAttributes        map[string]any `json:"user_attributes,omitempty"`
	Settings              map[string]any `json:"settings,omitempty"`
}

// Validate enforces basic domain rules on user creation.
func (r *CreateUserRequest) Validate() error {
	if r.Username == "" {
		return ErrInvalidIdentity("username is required")
	}
	if !strings.Contains(r.Email, "@") {
		return ErrInvalidIdentity("invalid email address")
	}
	if len(r.Password) < 8 {
		return ErrInvalidIdentity("password must be at least 8 characters")
	}
	if r.UserType == "" {
		return ErrInvalidIdentity("user_type is required")
	}
	return nil
}

type UpdateUserRequest struct {
	Username              *string        `json:"username,omitempty"`
	Email                 *string        `json:"email,omitempty"`
	UserType              *string        `json:"user_type,omitempty"`
	AccountStatus         *string        `json:"account_status,omitempty"`
	SessionTimeoutMinutes *int32         `json:"session_timeout_minutes,omitempty"`
	MfaEnabled            *bool          `json:"mfa_enabled,omitempty"`
	UserAttributes        map[string]any `json:"user_attributes,omitempty"`
	Settings              map[string]any `json:"settings,omitempty"`
}

type ListUsersRequest struct {
	UserType      *string `json:"user_type,omitempty"`
	AccountStatus *string `json:"account_status,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
	Limit         int     `json:"limit"`
	Offset        int     `json:"offset"`
}

// AuthenticateRequest is the credential bundle passed to Authenticate.
type AuthenticateRequest struct {
	Identifier string `json:"identifier" validate:"required"` // email or username
	Password   string `json:"password"   validate:"required"`
}

// ChangePasswordRequest carries old and new passwords for ChangePassword.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password"     validate:"required,min=8"`
}

type CreatePersonRequest struct {
	EntityID           uuid.UUID      `json:"entity_id"  validate:"required"`
	PersonType         string         `json:"person_type" validate:"required"`
	FirstName          string         `json:"first_name"  validate:"required,min=2,max=100"`
	LastName           string         `json:"last_name"   validate:"required,min=2,max=100"`
	MiddleName         *string        `json:"middle_name,omitempty"`
	Email              *string        `json:"email"       validate:"required,email"`
	Phone              *string        `json:"phone,omitempty"`
	BirthDate          *time.Time     `json:"birth_date,omitempty"`
	NationalID         *string        `json:"national_id,omitempty"`
	TaxID              *string        `json:"tax_id,omitempty"`
	Address            []byte         `json:"address,omitempty"`
	SecurityAttributes map[string]any `json:"security_attributes,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

type CreateEmployeeRequest struct {
	PersonID         uuid.UUID        `json:"person_id"        validate:"required"`
	EmployeeNumber   string           `json:"employee_number"  validate:"required,min=2,max=50"`
	EntityID         uuid.UUID        `json:"entity_id"        validate:"required"`
	PositionTitle    *string          `json:"position_title,omitempty"`
	DepartmentID     *uuid.UUID       `json:"department_id,omitempty"`
	ManagerID        *uuid.UUID       `json:"manager_id,omitempty"`
	HireDate         time.Time        `json:"hire_date"`
	SalaryInfo       map[string]any   `json:"salary_info,omitempty"`
	Status           EmploymentStatus `json:"employment_status"`
	WorkSchedule     map[string]any   `json:"work_schedule,omitempty"`
	SecurityLevel    int32            `json:"security_level"`
	AccessAttributes map[string]any   `json:"access_attributes,omitempty"`
}
