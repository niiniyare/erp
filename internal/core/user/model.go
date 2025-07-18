package user

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
)

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

func (a AccountStatus) String() string {
	return string(a)
}

func AllAccountStatuses() []AccountStatus {
	return []AccountStatus{
		AccountStatusActive,
		AccountStatusInactive,
		AccountStatusLocked,
		AccountStatusSuspended,
	}
}

// User represents a user in the system (domain model)
type User struct {
	ID                    uuid.UUID              `json:"id"`
	TenantID              uuid.UUID              `json:"tenant_id"`
	EntityID              uuid.UUID              `json:"entity_id"`
	PersonID              *uuid.UUID             `json:"person_id,omitempty"`
	EmployeeID            *uuid.UUID             `json:"employee_id,omitempty"`
	Username              string                 `json:"username"`
	Email                 string                 `json:"email"`
	UserType              string                 `json:"user_type"`
	AccountStatus         AccountStatus          `json:"account_status"`
	IsActive              bool                   `json:"is_active"`
	LastLoginAt           *time.Time             `json:"last_login_at,omitempty"`
	PasswordChangedAt     *time.Time             `json:"password_changed_at,omitempty"`
	FailedLoginAttempts   int32                  `json:"failed_login_attempts"`
	LockoutUntil          *time.Time             `json:"lockout_until,omitempty"`
	SessionTimeoutMinutes int32                  `json:"session_timeout_minutes"`
	MfaEnabled            bool                   `json:"mfa_enabled"`
	UserAttributes        map[string]any `json:"user_attributes,omitempty"`
	Settings              map[string]any `json:"settings,omitempty"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
	DeletedAt             *time.Time             `json:"deleted_at,omitempty"`
}

// Person represents a person in the system (domain model)
type Person struct {
	ID                 uuid.UUID              `json:"id"`
	TenantID           uuid.UUID              `json:"tenant_id"`
	EntityID           uuid.UUID              `json:"entity_id"`
	PersonType         string                 `json:"person_type"`
	FirstName          string                 `json:"first_name"`
	LastName           string                 `json:"last_name"`
	MiddleName         *string                `json:"middle_name,omitempty"`
	Email              string                 `json:"email,omitempty"`
	Phone              *string                `json:"phone,omitempty"`
	BirthDate          *time.Time             `json:"birth_date,omitempty"`
	NationalID         *string                `json:"national_id,omitempty"`
	TaxID              *string                `json:"tax_id,omitempty"`
	Address            []byte                 `json:"address,omitempty"`
	SecurityAttributes map[string]any `json:"security_attributes,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	IsActive           bool                   `json:"is_active"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	DeletedAt          *time.Time             `json:"deleted_at,omitempty"`
}

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

// AllEmploymentStatuses returns all valid enum values.
func AllEmploymentStatuses() []EmploymentStatus {
	return []EmploymentStatus{
		EmploymentStatusActive,
		EmploymentStatusInactive,
		EmploymentStatusTerminated,
		EmploymentStatusOnLeave,
		EmploymentStatusSuspended,
	}
}

// IsValid checks if the value is a valid enum.
func (e EmploymentStatus) IsValid() bool {
	_, ok := validEmploymentStatuses[e]
	return ok
}

func (e EmploymentStatus) String() string {
	return string(e)
}

// Employee represents an employee in the system (domain model)
type Employee struct {
	ID               uuid.UUID              `json:"id"`
	TenantID         uuid.UUID              `json:"tenant_id"`
	PersonID         uuid.UUID              `json:"person_id"`
	EmployeeNumber   string                 `json:"employee_number"`
	EntityID         uuid.UUID              `json:"entity_id"`
	PositionTitle    *string                `json:"position_title,omite/mpty"`
	DepartmentID     *uuid.UUID             `json:"department_id,omitempty"`
	ManagerID        *uuid.UUID             `json:"manager_id,omitempty"`
	HireDate         time.Time              `json:"hire_date"`
	TerminationDate  *time.Time             `json:"termination_date,omitempty"`
	SalaryInfo       map[string]any `json:"salary_info,omitempty"`
	Status           EmploymentStatus       `json:"employment_status"`
	WorkSchedule     map[string]any `json:"work_schedule,omitempty"`
	SecurityLevel    int32                  `json:"security_level"`
	AccessAttributes map[string]any `json:"access_attributes,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	DeletedAt        *time.Time             `json:"deleted_at,omitempty"`
}

// UserWithDetails represents user with person and employee details
type UserWithDetails struct {
	User     `json:",inline"`
	Person   *Person   `json:"person,omitempty"`
	Employee *Employee `json:"employee,omitempty"`
}

// UserRole represents a role assigned to a user
type UserRole struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	RoleID         uuid.UUID  `json:"role_id"`
	EntityID       uuid.UUID  `json:"entity_id"`
	AssignmentType string     `json:"assignment_type"`
	AssignedAt     time.Time  `json:"assigned_at"`
	AssignedBy     *uuid.UUID `json:"assigned_by,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	IsActive       bool       `json:"is_active"`
}

// UserSession represents a user session
type UserSession struct {
	ID             uuid.UUID `json:"id"`
	TenantID       int32     `json:"tenant_id"`
	UserID         uuid.UUID `json:"user_id"`
	SessionToken   string    `json:"session_token"`
	RefreshToken   string    `json:"refresh_token"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	DeviceInfo     string    `json:"device_info"`
	LocationInfo   string    `json:"location_info"`
	IsActive       bool      `json:"is_active"`
	ExpiresAt      time.Time `json:"expires_at"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// Request/Response types

// CreateUserRequest represents user creation request
type CreateUserRequest struct {
	EntityID              uuid.UUID              `json:"entity_id" validate:"required"`
	PersonID              *uuid.UUID             `json:"person_id,omitempty"`
	EmployeeID            *uuid.UUID             `json:"employee_id,omitempty"`
	Username              string                 `json:"username" validate:"required,min=3,max=50"`
	Email                 string                 `json:"email" validate:"required,email"`
	Password              string                 `json:"password" validate:"required,min=8"`
	UserType              string                 `json:"user_type" validate:"required"`
	AccountStatus         string                 `json:"account_status"`
	SessionTimeoutMinutes int32                  `json:"session_timeout_minutes"`
	MfaEnabled            bool                   `json:"mfa_enabled"`
	UserAttributes        map[string]any `json:"user_attributes,omitempty"`
	Settings              map[string]any `json:"settings,omitempty"`
}

// CreatePersonRequest represents person creation request
type CreatePersonRequest struct {
	EntityID           uuid.UUID              `json:"entity_id" validate:"required"`
	PersonType         string                 `json:"person_type" validate:"required"`
	FirstName          string                 `json:"first_name" validate:"required,min=2,max=100"`
	LastName           string                 `json:"last_name" validate:"required,min=2,max=100"`
	MiddleName         *string                `json:"middle_name,omitempty"`
	Email              *string                `json:"email" validate:"required,email"`
	Phone              *string                `json:"phone,omitempty"`
	BirthDate          *time.Time             `json:"birth_date,omitempty"`
	NationalID         *string                `json:"national_id,omitempty"`
	TaxID              *string                `json:"tax_id,omitempty"`
	Address            []byte                 `json:"address,omitempty"`
	SecurityAttributes map[string]any `json:"security_attributes,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

// CreateEmployeeRequest represents employee creation request
type CreateEmployeeRequest struct {
	PersonID         uuid.UUID              `json:"person_id" validate:"required"`
	EmployeeNumber   string                 `json:"employee_number" validate:"required,min=2,max=50"`
	EntityID         uuid.UUID              `json:"entity_id" validate:"required"`
	PositionTitle    *string                `json:"position_title,omitempty"`
	DepartmentID     *uuid.UUID             `json:"department_id,omitempty"`
	ManagerID        *uuid.UUID             `json:"manager_id,omitempty"`
	HireDate         time.Time              `json:"hire_date"`
	SalaryInfo       map[string]any `json:"salary_info,omitempty"`
	Status           EmploymentStatus       `json:"employment_status"`
	WorkSchedule     map[string]any `json:"work_schedule,omitempty"`
	SecurityLevel    int32                  `json:"security_level"`
	AccessAttributes map[string]any `json:"access_attributes,omitempty"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	Username              *string                `json:"username,omitempty"`
	Email                 *string                `json:"email,omitempty"`
	UserType              *string                `json:"user_type,omitempty"`
	AccountStatus         *string                `json:"account_status,omitempty"`
	SessionTimeoutMinutes *int32                 `json:"session_timeout_minutes,omitempty"`
	MfaEnabled            *bool                  `json:"mfa_enabled,omitempty"`
	UserAttributes        map[string]any `json:"user_attributes,omitempty"`
	Settings              map[string]any `json:"settings,omitempty"`
}

// ListUsersRequest represents user list request with filters
type ListUsersRequest struct {
	UserType      *string `json:"user_type,omitempty"`
	AccountStatus *string `json:"account_status,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
	Limit         int     `json:"limit"`
	Offset        int     `json:"offset"`
}

// AuthenticateRequest represents authentication request
type AuthenticateRequest struct {
	Identifier string `json:"identifier" validate:"required"` // username or email
	Password   string `json:"password" validate:"required"`
}

// AuthenticateResponse represents authentication response
type AuthenticateResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// Conversion functions

// FromSQLCUser converts SQLC User to domain User
func FromSQLCUser(sqlcUser *db.User) (*User, error) {
	var userAttributes map[string]any
	if len(sqlcUser.UserAttributes) > 0 {
		if err := json.Unmarshal(sqlcUser.UserAttributes, &userAttributes); err != nil {
			return nil, err
		}
	}

	var settings map[string]any
	if len(sqlcUser.Settings) > 0 {
		if err := json.Unmarshal(sqlcUser.Settings, &settings); err != nil {
			return nil, err
		}
	}

	var deletedAt *time.Time
	if sqlcUser.DeletedAt.Valid {
		deletedAt = &sqlcUser.DeletedAt.Time
	}

	var lastLoginAt *time.Time
	if sqlcUser.LastLoginAt.Valid {
		lastLoginAt = &sqlcUser.LastLoginAt.Time
	}

	var passwordChangedAt *time.Time
	if sqlcUser.PasswordChangedAt.Valid {
		passwordChangedAt = &sqlcUser.PasswordChangedAt.Time
	}

	var lockoutUntil *time.Time
	if sqlcUser.LockoutUntil.Valid {
		lockoutUntil = &sqlcUser.LockoutUntil.Time
	}

	// Handle optional pointers from SQLC
	username := ""
	if sqlcUser.Username != nil {
		username = *sqlcUser.Username
	}

	accountStatus := AccountStatusActive
	if sqlcUser.AccountStatus != nil {
		accountStatus = AccountStatus(*sqlcUser.AccountStatus)
	}

	failedLoginAttempts := int32(0)
	if sqlcUser.FailedLoginAttempts != nil {
		failedLoginAttempts = *sqlcUser.FailedLoginAttempts
	}

	sessionTimeoutMinutes := int32(480) // default 8 hours
	if sqlcUser.SessionTimeoutMinutes != nil {
		sessionTimeoutMinutes = *sqlcUser.SessionTimeoutMinutes
	}

	mfaEnabled := false
	if sqlcUser.MfaEnabled != nil {
		mfaEnabled = *sqlcUser.MfaEnabled
	}

	return &User{
		ID:                    sqlcUser.ID,
		TenantID:              sqlcUser.TenantID,
		EntityID:              sqlcUser.EntityID,
		PersonID:              sqlcUser.PersonID,
		EmployeeID:            sqlcUser.EmployeeID,
		Username:              username,
		Email:                 sqlcUser.Email,
		UserType:              sqlcUser.UserType,
		AccountStatus:         accountStatus,
		IsActive:              sqlcUser.IsActive,
		LastLoginAt:           lastLoginAt,
		PasswordChangedAt:     passwordChangedAt,
		FailedLoginAttempts:   failedLoginAttempts,
		LockoutUntil:          lockoutUntil,
		SessionTimeoutMinutes: sessionTimeoutMinutes,
		MfaEnabled:            mfaEnabled,
		UserAttributes:        userAttributes,
		Settings:              settings,
		CreatedAt:             sqlcUser.CreatedAt,
		UpdatedAt:             sqlcUser.UpdatedAt,
		DeletedAt:             deletedAt,
	}, nil
}

// FromSQLCPerson converts SQLC Person to domain Person
func FromSQLCPerson(sqlcPerson *db.Person) (*Person, error) {
	var securityAttributes map[string]any
	if len(sqlcPerson.SecurityAttributes) > 0 {
		if err := json.Unmarshal(sqlcPerson.SecurityAttributes, &securityAttributes); err != nil {
			return nil, err
		}
	}

	var metadata map[string]any
	if len(sqlcPerson.Metadata) > 0 {
		if err := json.Unmarshal(sqlcPerson.Metadata, &metadata); err != nil {
			return nil, err
		}
	}
	var deletedAt *time.Time
	if !sqlcPerson.DeletedAt.Time.IsZero() && sqlcPerson.DeletedAt.Valid {
		deletedAt = &sqlcPerson.DeletedAt.Time
	}

	var birthDate *time.Time
	if !sqlcPerson.BirthDate.IsZero() {
		birthDate = &sqlcPerson.BirthDate
	}

	// Handle optional email pointer from SQLC
	email := ""
	if sqlcPerson.Email != nil {
		email = *sqlcPerson.Email
	}

	return &Person{
		ID:                 sqlcPerson.ID,
		TenantID:           sqlcPerson.TenantID,
		EntityID:           sqlcPerson.EntityID,
		PersonType:         sqlcPerson.PersonType,
		FirstName:          sqlcPerson.FirstName,
		LastName:           sqlcPerson.LastName,
		MiddleName:         sqlcPerson.MiddleName,
		Email:              email,
		Phone:              sqlcPerson.Phone,
		BirthDate:          birthDate,
		NationalID:         sqlcPerson.NationalID,
		TaxID:              sqlcPerson.TaxID,
		Address:            sqlcPerson.Address,
		SecurityAttributes: securityAttributes,
		Metadata:           metadata,
		IsActive:           sqlcPerson.IsActive,
		CreatedAt:          sqlcPerson.CreatedAt,
		UpdatedAt:          sqlcPerson.UpdatedAt,
		DeletedAt:          deletedAt,
	}, nil
}

// FromSQLCEmployee converts SQLC Employee to domain Employee
func FromSQLCEmployee(sqlcEmployee *db.Employee) (*Employee, error) {
	var salaryInfo map[string]any
	if len(sqlcEmployee.SalaryInfo) > 0 {
		if err := json.Unmarshal(sqlcEmployee.SalaryInfo, &salaryInfo); err != nil {
			return nil, err
		}
	}

	var workSchedule map[string]any
	if len(sqlcEmployee.WorkSchedule) > 0 {
		if err := json.Unmarshal(sqlcEmployee.WorkSchedule, &workSchedule); err != nil {
			return nil, err
		}
	}

	var accessAttributes map[string]any
	if len(sqlcEmployee.AccessAttributes) > 0 {
		if err := json.Unmarshal(sqlcEmployee.AccessAttributes, &accessAttributes); err != nil {
			return nil, err
		}
	}

	var deletedAt *time.Time
	if sqlcEmployee.DeletedAt.Valid {
		deletedAt = &sqlcEmployee.DeletedAt.Time
	}

	var terminationDate *time.Time
	if !sqlcEmployee.TerminationDate.IsZero() {
		terminationDate = &sqlcEmployee.TerminationDate
	}

	// Handle optional pointers from SQLC
	status := EmploymentStatusActive
	if sqlcEmployee.EmploymentStatus != nil {
		status = EmploymentStatus(*sqlcEmployee.EmploymentStatus)
	}

	securityLevel := int32(0)
	if sqlcEmployee.SecurityLevel != nil {
		securityLevel = *sqlcEmployee.SecurityLevel
	}

	return &Employee{
		ID:               sqlcEmployee.ID,
		TenantID:         sqlcEmployee.TenantID,
		PersonID:         sqlcEmployee.PersonID,
		EmployeeNumber:   sqlcEmployee.EmployeeNumber,
		EntityID:         sqlcEmployee.EntityID,
		PositionTitle:    sqlcEmployee.PositionTitle,
		DepartmentID:     sqlcEmployee.DepartmentID,
		ManagerID:        sqlcEmployee.ManagerID,
		HireDate:         sqlcEmployee.HireDate,
		TerminationDate:  terminationDate,
		SalaryInfo:       salaryInfo,
		Status:           status,
		WorkSchedule:     workSchedule,
		SecurityLevel:    securityLevel,
		AccessAttributes: accessAttributes,
		CreatedAt:        sqlcEmployee.CreatedAt,
		UpdatedAt:        sqlcEmployee.UpdatedAt,
		DeletedAt:        deletedAt,
	}, nil
}

// ToSQLCCreateUserParams converts domain CreateUserRequest to SQLC params
func (req *CreateUserRequest) ToSQLCCreateUserParams() (db.CreateUserParams, error) {
	var userAttributes []byte
	if req.UserAttributes != nil {
		var err error
		userAttributes, err = json.Marshal(req.UserAttributes)
		if err != nil {
			return db.CreateUserParams{}, err
		}
	}

	var settings []byte
	if req.Settings != nil {
		var err error
		settings, err = json.Marshal(req.Settings)
		if err != nil {
			return db.CreateUserParams{}, err
		}
	}

	params := db.CreateUserParams{
		EntityID:              req.EntityID,
		PersonID:              req.PersonID,
		EmployeeID:            req.EmployeeID,
		Username:              &req.Username,
		Email:                 req.Email,
		UserType:              req.UserType,
		AccountStatus:         &req.AccountStatus,
		SessionTimeoutMinutes: &req.SessionTimeoutMinutes,
		MfaEnabled:            &req.MfaEnabled,
		UserAttributes:        userAttributes,
		Settings:              settings,
	}

	return params, nil
}

// ToSQLCCreatePersonParams converts domain CreatePersonRequest to SQLC params
func (req *CreatePersonRequest) ToSQLCCreatePersonParams() (db.CreatePersonParams, error) {
	var securityAttributes []byte
	if req.SecurityAttributes != nil {
		var err error
		securityAttributes, err = json.Marshal(req.SecurityAttributes)
		if err != nil {
			return db.CreatePersonParams{}, err
		}
	}

	var metadata []byte
	if req.Metadata != nil {
		var err error
		metadata, err = json.Marshal(req.Metadata)
		if err != nil {
			return db.CreatePersonParams{}, err
		}
	}

	params := db.CreatePersonParams{
		EntityID:           req.EntityID,
		PersonType:         req.PersonType,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		MiddleName:         req.MiddleName,
		Email:              req.Email,
		Phone:              req.Phone,
		NationalID:         req.NationalID,
		TaxID:              req.TaxID,
		Address:            req.Address,
		SecurityAttributes: securityAttributes,
		Metadata:           metadata,
	}

	return params, nil
}

// ToSQLCCreateEmployeeParams converts domain CreateEmployeeRequest to SQLC params
func (req *CreateEmployeeRequest) ToSQLCCreateEmployeeParams() (db.CreateEmployeeParams, error) {
	var salaryInfo []byte
	if req.SalaryInfo != nil {
		var err error
		salaryInfo, err = json.Marshal(req.SalaryInfo)
		if err != nil {
			return db.CreateEmployeeParams{}, err
		}
	}

	var workSchedule []byte
	if req.WorkSchedule != nil {
		var err error
		workSchedule, err = json.Marshal(req.WorkSchedule)
		if err != nil {
			return db.CreateEmployeeParams{}, err
		}
	}

	var accessAttributes []byte
	if req.AccessAttributes != nil {
		var err error
		accessAttributes, err = json.Marshal(req.AccessAttributes)
		if err != nil {
			return db.CreateEmployeeParams{}, err
		}
	}

	params := db.CreateEmployeeParams{
		PersonID:       req.PersonID,
		EmployeeNumber: req.EmployeeNumber,
		EntityID:       req.EntityID,
		PositionTitle:  req.PositionTitle,
		DepartmentID:   req.DepartmentID,
		ManagerID:      req.ManagerID,
		HireDate:       req.HireDate,
		SalaryInfo:     salaryInfo,
		// Status:           req.EmploymentStatusActive,
		WorkSchedule:     workSchedule,
		SecurityLevel:    &req.SecurityLevel,
		AccessAttributes: accessAttributes,
	}

	return params, nil
}

// GetFullName returns the full name of a person
func (p *Person) GetFullName() string {
	if p.MiddleName != nil && *p.MiddleName != "" {
		return p.FirstName + " " + *p.MiddleName + " " + p.LastName
	}
	return p.FirstName + " " + p.LastName
}

// IsValidUserType checks if the user type is valid
func IsValidUserType(userType string) bool {
	validTypes := []string{"ADMIN", "INTERNAL", "CUSTOMER", "VENDOR"}
	for _, t := range validTypes {
		if t == userType {
			return true
		}
	}
	return false
}

// IsValidAccountStatus checks if the account status is valid
func IsValidAccountStatus(status string) bool {
	validStatuses := []string{"ACTIVE", "INACTIVE", "LOCKED", "SUSPENDED"}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// IsValidEmploymentStatus checks if the employment status is valid
func IsValidEmploymentStatus(status string) bool {
	validStatuses := []string{"ACTIVE", "TERMINATED", "ON_LEAVE", "SUSPENDED"}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// ===== PERMISSION EVALUATION MODELS =====

// Role represents a role in the system
type Role struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	EntityID     uuid.UUID              `json:"entity_id"`
	Name         string                 `json:"name"`
	DisplayName  *string                `json:"display_name,omitempty"`
	Description  *string                `json:"description,omitempty"`
	RoleType     string                 `json:"role_type"`
	ParentRoleID *uuid.UUID             `json:"parent_role_id,omitempty"`
	Level        int32                  `json:"level"`
	Permissions  map[string]any `json:"permissions"`
	EntityScope  map[string]any `json:"entity_scope"`
	Conditions   map[string]any `json:"conditions"`
	IsActive     bool                   `json:"is_active"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// Permission represents a permission in the system
type Permission struct {
	ID                uuid.UUID              `json:"id"`
	TenantID          uuid.UUID              `json:"tenant_id"`
	ResourceID        uuid.UUID              `json:"resource_id"`
	ActionID          uuid.UUID              `json:"action_id"`
	Name              string                 `json:"name"`
	DisplayName       *string                `json:"display_name,omitempty"`
	Description       *string                `json:"description,omitempty"`
	Effect            string                 `json:"effect"` // ALLOW or DENY
	Conditions        map[string]any `json:"conditions"`
	DataFilters       map[string]any `json:"data_filters"`
	FieldRestrictions map[string]any `json:"field_restrictions"`
	IsActive          bool                   `json:"is_active"`
	CreatedAt         time.Time              `json:"created_at"`
}

// Policy represents an ABAC policy
type Policy struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	EntityID    *uuid.UUID             `json:"entity_id,omitempty"`
	Name        string                 `json:"name"`
	DisplayName *string                `json:"display_name,omitempty"`
	Description *string                `json:"description,omitempty"`
	PolicyType  string                 `json:"policy_type"`
	Effect      string                 `json:"effect"` // ALLOW or DENY
	Priority    int32                  `json:"priority"`
	Category    string                 `json:"category"`
	Target      map[string]any `json:"target"`
	Rule        map[string]any `json:"rule"`
	Obligations map[string]any `json:"obligations"`
	Advice      map[string]any `json:"advice"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// EffectivePermission represents a permission granted to a user through roles
type EffectivePermission struct {
	Permission     *Permission `json:"permission"`
	GrantedByRole  *Role       `json:"granted_by_role"`
	EntityID       uuid.UUID   `json:"entity_id"`
	AssignmentType string      `json:"assignment_type"`
	ExpiresAt      *time.Time  `json:"expires_at,omitempty"`
}

// PermissionEvaluationRequest represents a permission evaluation request
type PermissionEvaluationRequest struct {
	UserID       uuid.UUID              `json:"user_id" validate:"required"`
	ResourceName string                 `json:"resource_name" validate:"required"`
	ActionName   string                 `json:"action_name" validate:"required"`
	EntityID     *uuid.UUID             `json:"entity_id,omitempty"`
	Context      map[string]any `json:"context,omitempty"`
}

// PermissionEvaluationResult represents the result of permission evaluation
type PermissionEvaluationResult struct {
	Allowed           bool                  `json:"allowed"`
	PolicyDecisions   []string              `json:"policy_decisions"`
	EffectiveRoles    []string              `json:"effective_roles"`
	EvaluationTimeMS  int                   `json:"evaluation_time_ms"`
	CacheHit          bool                  `json:"cache_hit"`
	RBACResult        *RBACEvaluationResult `json:"rbac_result,omitempty"`
	ABACResult        *ABACEvaluationResult `json:"abac_result,omitempty"`
}

// RBACEvaluationResult represents RBAC evaluation result
type RBACEvaluationResult struct {
	Allowed         bool     `json:"allowed"`
	PolicyDecisions []string `json:"policy_decisions"`
}

// ABACEvaluationRequest represents ABAC evaluation request
type ABACEvaluationRequest struct {
	UserID       uuid.UUID              `json:"user_id" validate:"required"`
	ResourceName string                 `json:"resource_name" validate:"required"`
	ActionName   string                 `json:"action_name" validate:"required"`
	EntityID     *uuid.UUID             `json:"entity_id,omitempty"`
	Context      map[string]any `json:"context,omitempty"`
}

// ABACEvaluationResult represents ABAC evaluation result
type ABACEvaluationResult struct {
	Allowed            bool                   `json:"allowed"`
	PolicyDecisions    []string               `json:"policy_decisions"`
	ApplicablePolicies []string               `json:"applicable_policies"`
	EvaluationDetails  map[string]any `json:"evaluation_details,omitempty"`
}

// BulkPermissionEvaluationRequest represents multiple permission evaluations
type BulkPermissionEvaluationRequest struct {
	Requests []*PermissionEvaluationRequest `json:"requests" validate:"required,min=1"`
}

// PolicyTestRequest represents a policy test request
type PolicyTestRequest struct {
	UserID       uuid.UUID              `json:"user_id" validate:"required"`
	ResourceName string                 `json:"resource_name" validate:"required"`
	ActionName   string                 `json:"action_name" validate:"required"`
	EntityID     *uuid.UUID             `json:"entity_id,omitempty"`
	Context      map[string]any `json:"context,omitempty"`
}

// PolicyTestResult represents the result of policy testing
type PolicyTestResult struct {
	PolicyID      uuid.UUID              `json:"policy_id"`
	PolicyName    string                 `json:"policy_name"`
	TargetMatches bool                   `json:"target_matches"`
	RuleResult    bool                   `json:"rule_result"`
	Effect        string                 `json:"effect"`
	Details       map[string]any `json:"details"`
}

// ===== CONVERSION FUNCTIONS FOR PERMISSION MODELS =====

// FromSQLCRole converts SQLC Role to domain Role
func FromSQLCRole(sqlcRole *db.Role) (*Role, error) {
	var permissions map[string]any
	if len(sqlcRole.Permissions) > 0 {
		if err := json.Unmarshal(sqlcRole.Permissions, &permissions); err != nil {
			return nil, err
		}
	}

	var entityScope map[string]any
	if len(sqlcRole.EntityScope) > 0 {
		if err := json.Unmarshal(sqlcRole.EntityScope, &entityScope); err != nil {
			return nil, err
		}
	}

	var conditions map[string]any
	if len(sqlcRole.Conditions) > 0 {
		if err := json.Unmarshal(sqlcRole.Conditions, &conditions); err != nil {
			return nil, err
		}
	}

	// Handle optional pointers from SQLC
	description := &sqlcRole.Description

	roleType := ""
	if sqlcRole.RoleType != nil {
		roleType = *sqlcRole.RoleType
	}

	level := int32(0)
	if sqlcRole.Level != nil {
		level = *sqlcRole.Level
	}

	isActive := false
	if sqlcRole.IsActive != nil {
		isActive = *sqlcRole.IsActive
	}

	return &Role{
		ID:           sqlcRole.ID,
		TenantID:     sqlcRole.TenantID,
		EntityID:     sqlcRole.EntityID,
		Name:         sqlcRole.Name,
		DisplayName:  sqlcRole.DisplayName,
		Description:  description,
		RoleType:     roleType,
		ParentRoleID: sqlcRole.ParentRoleID,
		Level:        level,
		Permissions:  permissions,
		EntityScope:  entityScope,
		Conditions:   conditions,
		IsActive:     isActive,
		CreatedAt:    sqlcRole.CreatedAt,
		UpdatedAt:    sqlcRole.UpdatedAt,
	}, nil
}

// FromSQLCPermission converts SQLC Permission to domain Permission
func FromSQLCPermission(sqlcPermission *db.Permission) (*Permission, error) {
	var conditions map[string]any
	if len(sqlcPermission.Conditions) > 0 {
		if err := json.Unmarshal(sqlcPermission.Conditions, &conditions); err != nil {
			return nil, err
		}
	}

	var dataFilters map[string]any
	if len(sqlcPermission.DataFilters) > 0 {
		if err := json.Unmarshal(sqlcPermission.DataFilters, &dataFilters); err != nil {
			return nil, err
		}
	}

	var fieldRestrictions map[string]any
	if len(sqlcPermission.FieldRestrictions) > 0 {
		if err := json.Unmarshal(sqlcPermission.FieldRestrictions, &fieldRestrictions); err != nil {
			return nil, err
		}
	}

	// Handle optional pointers from SQLC
	description := &sqlcPermission.Description

	effect := ""
	if sqlcPermission.Effect != nil {
		effect = *sqlcPermission.Effect
	}

	isActive := false
	if sqlcPermission.IsActive != nil {
		isActive = *sqlcPermission.IsActive
	}

	createdAt := time.Time{}
	if sqlcPermission.CreatedAt.Valid {
		createdAt = sqlcPermission.CreatedAt.Time
	}

	return &Permission{
		ID:                sqlcPermission.ID,
		TenantID:          sqlcPermission.TenantID,
		ResourceID:        sqlcPermission.ResourceID,
		ActionID:          sqlcPermission.ActionID,
		Name:              sqlcPermission.Name,
		DisplayName:       sqlcPermission.DisplayName,
		Description:       description,
		Effect:            effect,
		Conditions:        conditions,
		DataFilters:       dataFilters,
		FieldRestrictions: fieldRestrictions,
		IsActive:          isActive,
		CreatedAt:         createdAt,
	}, nil
}

// FromSQLCPolicy converts SQLC Policy to domain Policy
func FromSQLCPolicy(sqlcPolicy *db.Policy) (*Policy, error) {
	var target map[string]any
	if len(sqlcPolicy.Target) > 0 {
		if err := json.Unmarshal(sqlcPolicy.Target, &target); err != nil {
			return nil, err
		}
	}

	var rule map[string]any
	if len(sqlcPolicy.Rule) > 0 {
		if err := json.Unmarshal(sqlcPolicy.Rule, &rule); err != nil {
			return nil, err
		}
	}

	var obligations map[string]any
	if len(sqlcPolicy.Obligations) > 0 {
		if err := json.Unmarshal(sqlcPolicy.Obligations, &obligations); err != nil {
			return nil, err
		}
	}

	var advice map[string]any
	if len(sqlcPolicy.Advice) > 0 {
		if err := json.Unmarshal(sqlcPolicy.Advice, &advice); err != nil {
			return nil, err
		}
	}

	// Handle optional pointers from SQLC
	description := &sqlcPolicy.Description

	policyType := ""
	if sqlcPolicy.PolicyType != nil {
		policyType = *sqlcPolicy.PolicyType
	}

	effect := ""
	if sqlcPolicy.Effect != nil {
		effect = *sqlcPolicy.Effect
	}

	priority := int32(0)
	if sqlcPolicy.Priority != nil {
		priority = *sqlcPolicy.Priority
	}

	category := ""
	if sqlcPolicy.Category != nil {
		category = *sqlcPolicy.Category
	}

	isActive := false
	if sqlcPolicy.IsActive != nil {
		isActive = *sqlcPolicy.IsActive
	}

	createdAt := time.Time{}
	if sqlcPolicy.CreatedAt.Valid {
		createdAt = sqlcPolicy.CreatedAt.Time
	}

	updatedAt := time.Time{}
	if sqlcPolicy.UpdatedAt.Valid {
		updatedAt = sqlcPolicy.UpdatedAt.Time
	}

	return &Policy{
		ID:          sqlcPolicy.ID,
		TenantID:    sqlcPolicy.TenantID,
		EntityID:    sqlcPolicy.EntityID,
		Name:        sqlcPolicy.Name,
		DisplayName: sqlcPolicy.DisplayName,
		Description: description,
		PolicyType:  policyType,
		Effect:      effect,
		Priority:    priority,
		Category:    category,
		Target:      target,
		Rule:        rule,
		Obligations: obligations,
		Advice:      advice,
		IsActive:    isActive,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// FromSQLCUserRole converts SQLC UserRole to domain UserRole
func FromSQLCUserRole(sqlcUserRole *db.UserRole) (*UserRole, error) {
	var expiresAt *time.Time
	if sqlcUserRole.ExpiresAt.Valid {
		expiresAt = &sqlcUserRole.ExpiresAt.Time
	}

	// Handle optional pointers from SQLC
	assignmentType := ""
	if sqlcUserRole.AssignmentType != nil {
		assignmentType = *sqlcUserRole.AssignmentType
	}

	isActive := false
	if sqlcUserRole.IsActive != nil {
		isActive = *sqlcUserRole.IsActive
	}

	return &UserRole{
		ID:             sqlcUserRole.ID,
		UserID:         sqlcUserRole.UserID,
		RoleID:         sqlcUserRole.RoleID,
		EntityID:       sqlcUserRole.EntityID,
		AssignmentType: assignmentType,
		AssignedAt:     sqlcUserRole.AssignedAt,
		AssignedBy:     sqlcUserRole.AssignedBy,
		ExpiresAt:      expiresAt,
		IsActive:       isActive,
	}, nil
}
