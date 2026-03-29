// Package domain contains pure IAM domain models and business rules.
// No external framework dependencies — only stdlib and uuid.
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Account Status
// =============================================================================

// AccountStatus is the lifecycle state of a user's login account.
// It is stored as a typed string so the compiler rejects bare literals in
// positions that expect an account state.
// 
// State transitions

// 
// 	INACTIVE ──► ACTIVE ──► LOCKED      (too many failed login attempts)
// 	         ◄──         ──► SUSPENDED  (manual admin action)
// 	                     ◄── ACTIVE     (admin unlock / password reset)
type AccountStatus string

const (
	// AccountStatusActive — the account can authenticate and issue requests.
	AccountStatusActive AccountStatus = "ACTIVE"

	// AccountStatusInactive — the account has been created but not yet
	// activated, or has been deactivated by an administrator.
	AccountStatusInactive AccountStatus = "INACTIVE"

	// AccountStatusLocked — the account is temporarily locked after
	// exceeding the configured failed-login threshold.  It unlocks
	// automatically when LockoutUntil passes, or immediately on admin reset.
	AccountStatusLocked AccountStatus = "LOCKED"

	// AccountStatusSuspended — the account has been suspended by an
	// administrator and cannot authenticate until explicitly re-activated.
	AccountStatusSuspended AccountStatus = "SUSPENDED"
)

var validAccountStatuses = map[AccountStatus]struct{}{
	AccountStatusActive:    {},
	AccountStatusInactive:  {},
	AccountStatusLocked:    {},
	AccountStatusSuspended: {},
}

// IsValid reports whether the status is one of the defined enum values.
func (a AccountStatus) IsValid() bool {
	_, ok := validAccountStatuses[a]
	return ok
}

func (a AccountStatus) String() string { return string(a) }

// AllAccountStatuses returns every defined AccountStatus in a stable order.
// Useful for validation loops and UI enum lists.
func AllAccountStatuses() []AccountStatus {
	return []AccountStatus{
		AccountStatusActive,
		AccountStatusInactive,
		AccountStatusLocked,
		AccountStatusSuspended,
	}
}

// =============================================================================
// Employment Status
// =============================================================================

// EmploymentStatus is the HR lifecycle state of an Employee record.
type EmploymentStatus string

const (
	// EmploymentStatusActive — the employee is currently working.
	EmploymentStatusActive EmploymentStatus = "ACTIVE"

	// EmploymentStatusInactive — the employee record exists but is not
	// currently in an active employment period (e.g. pre-hire, drafted).
	EmploymentStatusInactive EmploymentStatus = "INACTIVE"

	// EmploymentStatusTerminated — employment has ended.
	// TerminationDate should be populated when this status is set.
	EmploymentStatusTerminated EmploymentStatus = "TERMINATED"

	// EmploymentStatusOnLeave — the employee is on approved leave
	// (annual, sick, maternity / paternity, etc.).
	EmploymentStatusOnLeave EmploymentStatus = "ON_LEAVE"

	// EmploymentStatusSuspended — the employee has been suspended
	// pending investigation or disciplinary action.
	EmploymentStatusSuspended EmploymentStatus = "SUSPENDED"
)

var validEmploymentStatuses = map[EmploymentStatus]struct{}{
	EmploymentStatusActive:     {},
	EmploymentStatusInactive:   {},
	EmploymentStatusTerminated: {},
	EmploymentStatusOnLeave:    {},
	EmploymentStatusSuspended:  {},
}

// IsValid reports whether the status is one of the defined enum values.
func (e EmploymentStatus) IsValid() bool {
	_, ok := validEmploymentStatuses[e]
	return ok
}

func (e EmploymentStatus) String() string { return string(e) }

// AllEmploymentStatuses returns every defined EmploymentStatus in a stable order.
func AllEmploymentStatuses() []EmploymentStatus {
	return []EmploymentStatus{
		EmploymentStatusActive,
		EmploymentStatusInactive,
		EmploymentStatusTerminated,
		EmploymentStatusOnLeave,
		EmploymentStatusSuspended,
	}
}

// =============================================================================
// User — Aggregate Root
// =============================================================================

// User is the aggregate root for a login identity in the IAM bounded context.
// 
// Conceptual model

// 
// 	A User is a credential bundle (username + password hash, MFA state,
// 	session settings) that may optionally be linked to a Person and/or an
// 	Employee record:
// 
// 	  User ──(PersonID)──► Person  (the natural person behind the login)
// 	       ──(EmployeeID)─► Employee (the HR record, if this user is staff)
// 
// 	This separation allows non-person system accounts (UserType = "API") and
// 	external portal users (UserType = "PORTAL") to have a User record without
// 	requiring a Person or Employee to exist.
// 
// TenantID vs EntityID

// 
// 	TenantID is the hard RLS boundary.  Every query against this user's data
// 	must be issued inside a transaction with SET LOCAL awo.tenant_id =
// 	'<TenantID>'.  See actor.go for the full isolation model.
// 
// 	EntityID indicates which entity (branch, department, subsidiary) within
// 	the tenant this user primarily belongs to.  It is the basis for computing
// 	EntityScope at login time; the actual access-breadth decision is made in
// 	the session service by inspecting the user's roles against their entity.
// 
// PrincipalID

// 
// 	Non-nil for portal users (UserType = "PORTAL" / "CUSTOMER") only.
// 	It identifies the external contact or party record (e.g. a supplier
// 	account, customer profile) that this login credential represents.
// 	Not related to the domain.Principal value object used in Casbin.
type User struct {
	ID                    uuid.UUID      `json:"id"`
	TenantID              uuid.UUID      `json:"tenant_id"`              // RLS key
	EntityID              uuid.UUID      `json:"entity_id"`              // application-layer entity scope anchor
	PersonID              *uuid.UUID     `json:"person_id,omitempty"`    // nil for system / API accounts
	EmployeeID            *uuid.UUID     `json:"employee_id,omitempty"`  // nil for non-staff users
	PrincipalID           *uuid.UUID     `json:"principal_id,omitempty"` // non-nil for portal users only; identifies the represented party
	Username              string         `json:"username"`
	Email                 string         `json:"email"`
	DisplayName           *string        `json:"display_name,omitempty"`
	UserType              string         `json:"user_type"` // "INTERNAL"|"SYSADMIN"|"PORTAL"|"CUSTOMER"|"API"
	AccountStatus         AccountStatus  `json:"account_status"`
	IsActive              bool           `json:"is_active"`
	LastLoginAt           *time.Time     `json:"last_login_at,omitempty"`
	PasswordChangedAt     *time.Time     `json:"password_changed_at,omitempty"`
	FailedLoginAttempts   int32          `json:"failed_login_attempts"`
	LockoutUntil          *time.Time     `json:"lockout_until,omitempty"` // nil or zero → not locked
	SessionTimeoutMinutes int32          `json:"session_timeout_minutes"`
	MfaEnabled            bool           `json:"mfa_enabled"`
	UserAttributes        map[string]any `json:"user_attributes,omitempty"` // extensible per-user metadata
	Settings              map[string]any `json:"settings,omitempty"`        // user-level preference overrides
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             *time.Time     `json:"deleted_at,omitempty"`
}

// IsLocked reports whether the account is currently in a temporary lockout.
// Alockout expires automatically when LockoutUntil passes; this method
// reflects that without requiring a status column update.
func (u *User) IsLocked() bool {
	return u.LockoutUntil != nil && !u.LockoutUntil.IsZero() && u.LockoutUntil.After(time.Now())
}

// CanAuthenticate reports whether the user is permitted to start a new
// session.  The account must be active, not soft-deleted, and not locked.
func (u *User) CanAuthenticate() bool {
	return u.IsActive &&
		u.DeletedAt == nil &&
		u.AccountStatus == AccountStatusActive &&
		!u.IsLocked()
}

// =============================================================================
// Person — Domain Entity
// =============================================================================

// Person represents a natural human being associated with a user account or
// an entity (e.g. a contact on a supplier record).
// 
// APerson record holds PII (name, national ID, tax ID, date of birth) that
// is shared across different contexts — the same Person may be linked to an
// Employee in the HR module and to a user account in the IAM module.
// 
// TenantID / EntityID note

// 
// 	Same isolation model as User.  TenantID is enforced by RLS; EntityID is
// 	the application-layer scope anchor used to build EntityScope at login.
type Person struct {
	ID                 uuid.UUID      `json:"id"`
	TenantID           uuid.UUID      `json:"tenant_id"`   // RLS key
	EntityID           uuid.UUID      `json:"entity_id"`   // application-layer entity
	PersonType         string         `json:"person_type"` // e.g. "EMPLOYEE", "CONTACT", "DEPENDENT"
	FirstName          string         `json:"first_name"`
	LastName           string         `json:"last_name"`
	MiddleName         *string        `json:"middle_name,omitempty"`
	Email              string         `json:"email,omitempty"`
	Phone              *string        `json:"phone,omitempty"`
	BirthDate          *time.Time     `json:"birth_date,omitempty"`
	NationalID         *string        `json:"national_id,omitempty"`         // Kenya: national ID number
	TaxID              *string        `json:"tax_id,omitempty"`              // Kenya: KRA PIN
	Address            []byte         `json:"address,omitempty"`             // JSONB-encoded structured address
	SecurityAttributes map[string]any `json:"security_attributes,omitempty"` // access control hints for ABAC
	Metadata           map[string]any `json:"metadata,omitempty"`            // extensible per-person metadata
	IsActive           bool           `json:"is_active"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          *time.Time     `json:"deleted_at,omitempty"`
}

// GetFullName returns the person's full name.  When a middle name is present
// it is inserted between first and last names.
func (p *Person) GetFullName() string {
	if p.MiddleName != nil && *p.MiddleName != "" {
		return p.FirstName + " " + *p.MiddleName + " " + p.LastName
	}
	return p.FirstName + " " + p.LastName
}

// =============================================================================
// Employee — Domain Entity
// =============================================================================

// Employee represents an HR employment record linked to a Person.
// 
// Relationship to User

// 
// 	User.EmployeeID → Employee  (a user who is also a member of staff)
// 	Employee.PersonID → Person  (the natural person behind the employment)
// 
// An Employee record can exist without a corresponding User record for staff
// who do not have system login access (e.g. casual labourers, historical hires).
// 
// SecurityLevel

// 
// 	A coarse clearance integer (0 = unrestricted, higher = more sensitive).
// 	Used by the ABAC condition evaluator to gate access to high-sensitivity
// 	resources without requiring explicit Casbin policies for every combination.
// 
// TenantID / EntityID note

// 
// 	Same isolation model as User and Person.
type Employee struct {
	ID               uuid.UUID        `json:"id"`
	TenantID         uuid.UUID        `json:"tenant_id"` // RLS key
	PersonID         uuid.UUID        `json:"person_id"`
	EmployeeNumber   string           `json:"employee_number"` // human-readable identifier, e.g. "EMP-0042"
	EntityID         uuid.UUID        `json:"entity_id"`       // application-layer entity (home branch / department)
	PositionTitle    *string          `json:"position_title,omitempty"`
	DepartmentID     *uuid.UUID       `json:"department_id,omitempty"`
	ManagerID        *uuid.UUID       `json:"manager_id,omitempty"` // self-referencing; nil for top-level employees
	HireDate         time.Time        `json:"hire_date"`
	TerminationDate  *time.Time       `json:"termination_date,omitempty"` // nil unless Status = TERMINATED
	SalaryInfo       map[string]any   `json:"salary_info,omitempty"`      // JSONB; not surfaced in public APIs
	Status           EmploymentStatus `json:"employment_status"`
	WorkSchedule     map[string]any   `json:"work_schedule,omitempty"`
	SecurityLevel    int32            `json:"security_level"`              // 0= default; higher = more sensitive
	AccessAttributes map[string]any   `json:"access_attributes,omitempty"` // ABAC hints forwarded to condition evaluator
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        *time.Time       `json:"deleted_at,omitempty"`
}

// IsCurrentlyEmployed reports whether the employee is in an active employment
// state (ACTIVE or ON_LEAVE) and has not been soft-deleted.
func (e *Employee) IsCurrentlyEmployed() bool {
	return (e.Status == EmploymentStatusActive || e.Status == EmploymentStatusOnLeave) &&
		e.DeletedAt == nil
}

// =============================================================================
// Composite / Read Models
// =============================================================================

// UserWithDetails embeds a User with its optionally pre-loaded Person and
// Employee records.  Used by the session service at login time to build a
// ResolvedSession, and by the user-detail API endpoint.
// 
// Because User is embedded (not a pointer), a zero-value UserWithDetails has a
// valid User sub-struct.  Person and Employee are pointers because they are
// optional — not every user has both.
type UserWithDetails struct {
	User     `json:",inline"`
	Person   *Person   `json:"person,omitempty"`
	Employee *Employee `json:"employee,omitempty"`
}

// UserRole represents a role assignment record for a user as stored in the
// identity store (distinct from the Casbin g-rules, which are the enforcement
// records — this is the human-readable audit copy).
type UserRole struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"` // human-readable role label
	UserID         uuid.UUID  `json:"user_id"`
	RoleID         uuid.UUID  `json:"role_id"`
	EntityID       uuid.UUID  `json:"entity_id"`       // the entity within which this role is effective
	AssignmentType string     `json:"assignment_type"` // e.g. "DIRECT", "DELEGATED", "TEMPORARY"
	AssignedAt     time.Time  `json:"assigned_at"`
	IsActive       bool       `json:"is_active"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"` // nil → no expiry
}

// =============================================================================
// Command Objects (Request / Input DTOs)
// =============================================================================

// CreateUserRequest is the command object for the CreateUser use case.
// 
// It contains the minimum information required to mint a new User aggregate.
// Validation is split between struct tags (for HTTP binding layers) and the
// explicit Validate() method (for domain rule enforcement that is
// framework-agnostic).
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

// Validate enforces domain invariants for user creation.
// 
// This method is intentionally minimal — it checks only the rules that live
// in the domain layer and cannot be expressed as struct tags.  The HTTP
// handler layer is responsible for running tag-based validation (e.g. via
// go-playground/validator) before calling this method.
// 
// EntityID is checked here explicitly because it is a domain invariant:
// every user must belong to an entity for EntityScope to be computable at
// login time.  Without an EntityID the session service cannot determine the
// user's application-layer access breadth.
func (r *CreateUserRequest) Validate() error {
	if r.EntityID == uuid.Nil {
		return ErrInvalidIdentity("entity_id is required")
	}
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

// UpdateUserRequest is the command object for the UpdateUser use case.
// All fields are optional pointers; only non-nil fields are applied.
// This follows a partial-update (PATCH) semantics to avoid accidental
// zero-value overwrites.
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

// ListUsersRequest carries filter and pagination parameters for the ListUsers
// query.  A nil pointer field means "no filter on this dimension".
type ListUsersRequest struct {
	UserType      *string `json:"user_type,omitempty"`
	AccountStatus *string `json:"account_status,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
	Limit         int     `json:"limit"`
	Offset        int     `json:"offset"`
}

// AuthenticateRequest is the credential bundle passed to the Authenticate
// use case.  Identifier accepts either an email address or a username so
// that a single endpoint supports both login modes.
type AuthenticateRequest struct {
	Identifier string `json:"identifier" validate:"required"` // email or username
	Password   string `json:"password"   validate:"required"`
}

// ChangePasswordRequest carries the old and new credentials for the
// ChangePassword use case.  The service must verify CurrentPassword before
// accepting NewPassword to prevent privilege escalation via stolen sessions.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password"     validate:"required,min=8"`
}

// CreatePersonRequest is the command object for the CreatePerson use case.
// 
// EntityID note

// 
// 	The EntityID here anchors the Person to a specific entity within the
// 	tenant for application-layer filtering.  It does NOT affect which tenant
// 	the row belongs to (that is always governed by TenantID via RLS).
type CreatePersonRequest struct {
	EntityID           uuid.UUID      `json:"entity_id"   validate:"required"`
	PersonType         string         `json:"person_type" validate:"required"`
	FirstName          string         `json:"first_name"  validate:"required,min=2,max=100"`
	LastName           string         `json:"last_name"   validate:"required,min=2,max=100"`
	MiddleName         *string        `json:"middle_name,omitempty"`
	Email              *string        `json:"email"       validate:"required,email"`
	Phone              *string        `json:"phone,omitempty"`
	BirthDate          *time.Time     `json:"birth_date,omitempty"`
	NationalID         *string        `json:"national_id,omitempty"` // Kenya: national ID number
	TaxID              *string        `json:"tax_id,omitempty"`      // Kenya: KRA PIN
	Address            []byte         `json:"address,omitempty"`
	SecurityAttributes map[string]any `json:"security_attributes,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

// PasswordResetToken represents a single-use password reset token.
// The raw token is emailed to the user; only the SHA-256 hash is stored.
type PasswordResetToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// IsExpired reports whether the reset token is past its expiry time.
func (t *PasswordResetToken) IsExpired() bool { return time.Now().After(t.ExpiresAt) }

// IsUsed reports whether the token has already been consumed.
func (t *PasswordResetToken) IsUsed() bool { return t.UsedAt != nil }

// MFASetup is returned by UserService.InitiateMFA.
// The Secret must be shown to the user exactly once so they can save it to
// an authenticator app (Google Authenticator, Authy, etc.).
// The QRURI encodes all setup parameters for QR-code scanners.
type MFASetup struct {
	Secret string // Base32-encoded TOTP secret (RFC 4648, no padding)
	QRURI  string // otpauth:// URI — scan with authenticator app
}

// CreateEmployeeRequest is the command object for the CreateEmployee use case.
// 
// SecurityLevel defaults to 0 (unrestricted) when not supplied.  Callers
// should explicitly set a non-zero value for employees in roles that require
// higher-sensitivity access (e.g. payroll administrators, CFO).
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
	SecurityLevel    int32            `json:"security_level"` // 0= default; higher = more sensitive
	AccessAttributes map[string]any   `json:"access_attributes,omitempty"`
}
