package user

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
)

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
	AccountStatus         string                 `json:"account_status"`
	IsActive              bool                   `json:"is_active"`
	LastLoginAt           *time.Time             `json:"last_login_at,omitempty"`
	PasswordChangedAt     *time.Time             `json:"password_changed_at,omitempty"`
	FailedLoginAttempts   int32                  `json:"failed_login_attempts"`
	LockoutUntil          *time.Time             `json:"lockout_until,omitempty"`
	SessionTimeoutMinutes int32                  `json:"session_timeout_minutes"`
	MfaEnabled            bool                   `json:"mfa_enabled"`
	UserAttributes        map[string]interface{} `json:"user_attributes,omitempty"`
	Settings              map[string]interface{} `json:"settings,omitempty"`
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
	Email              string                 `json:"email"`
	Phone              *string                `json:"phone,omitempty"`
	BirthDate          *time.Time             `json:"birth_date,omitempty"`
	NationalID         *string                `json:"national_id,omitempty"`
	TaxID              *string                `json:"tax_id,omitempty"`
	Address            *string                `json:"address,omitempty"`
	SecurityAttributes map[string]interface{} `json:"security_attributes,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	IsActive           bool                   `json:"is_active"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	DeletedAt          *time.Time             `json:"deleted_at,omitempty"`
}

// Employee represents an employee in the system (domain model)
type Employee struct {
	ID               uuid.UUID              `json:"id"`
	TenantID         int32                  `json:"tenant_id"`
	PersonID         uuid.UUID              `json:"person_id"`
	EmployeeNumber   string                 `json:"employee_number"`
	EntityID         uuid.UUID              `json:"entity_id"`
	PositionTitle    *string                `json:"position_title,omitempty"`
	DepartmentID     *uuid.UUID             `json:"department_id,omitempty"`
	ManagerID        *uuid.UUID             `json:"manager_id,omitempty"`
	HireDate         time.Time              `json:"hire_date"`
	TerminationDate  *time.Time             `json:"termination_date,omitempty"`
	SalaryInfo       map[string]interface{} `json:"salary_info,omitempty"`
	EmploymentStatus string                 `json:"employment_status"`
	WorkSchedule     map[string]interface{} `json:"work_schedule,omitempty"`
	SecurityLevel    int32                  `json:"security_level"`
	AccessAttributes map[string]interface{} `json:"access_attributes,omitempty"`
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
	UserAttributes        map[string]interface{} `json:"user_attributes,omitempty"`
	Settings              map[string]interface{} `json:"settings,omitempty"`
}

// CreatePersonRequest represents person creation request
type CreatePersonRequest struct {
	EntityID           uuid.UUID              `json:"entity_id" validate:"required"`
	PersonType         string                 `json:"person_type" validate:"required"`
	FirstName          string                 `json:"first_name" validate:"required,min=2,max=100"`
	LastName           string                 `json:"last_name" validate:"required,min=2,max=100"`
	MiddleName         *string                `json:"middle_name,omitempty"`
	Email              string                 `json:"email" validate:"required,email"`
	Phone              *string                `json:"phone,omitempty"`
	BirthDate          *time.Time             `json:"birth_date,omitempty"`
	NationalID         *string                `json:"national_id,omitempty"`
	TaxID              *string                `json:"tax_id,omitempty"`
	Address            *string                `json:"address,omitempty"`
	SecurityAttributes map[string]interface{} `json:"security_attributes,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
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
	SalaryInfo       map[string]interface{} `json:"salary_info,omitempty"`
	EmploymentStatus string                 `json:"employment_status"`
	WorkSchedule     map[string]interface{} `json:"work_schedule,omitempty"`
	SecurityLevel    int32                  `json:"security_level"`
	AccessAttributes map[string]interface{} `json:"access_attributes,omitempty"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	Username              *string                `json:"username,omitempty"`
	Email                 *string                `json:"email,omitempty"`
	UserType              *string                `json:"user_type,omitempty"`
	AccountStatus         *string                `json:"account_status,omitempty"`
	SessionTimeoutMinutes *int32                 `json:"session_timeout_minutes,omitempty"`
	MfaEnabled            *bool                  `json:"mfa_enabled,omitempty"`
	UserAttributes        map[string]interface{} `json:"user_attributes,omitempty"`
	Settings              map[string]interface{} `json:"settings,omitempty"`
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
	var userAttributes map[string]interface{}
	if len(sqlcUser.UserAttributes) > 0 {
		if err := json.Unmarshal(sqlcUser.UserAttributes, &userAttributes); err != nil {
			return nil, err
		}
	}

	var settings map[string]interface{}
	if len(sqlcUser.Settings) > 0 {
		if err := json.Unmarshal(sqlcUser.Settings, &settings); err != nil {
			return nil, err
		}
	}

	var deletedAt *time.Time
	if sqlcUser.DeletedAt != nil && sqlcUser.DeletedAt.Valid {
		deletedAt = &sqlcUser.DeletedAt.Time
	}

	var lastLoginAt *time.Time
	if sqlcUser.LastLoginAt != nil && sqlcUser.LastLoginAt.Valid {
		lastLoginAt = &sqlcUser.LastLoginAt.Time
	}

	var passwordChangedAt *time.Time
	if sqlcUser.PasswordChangedAt != nil && sqlcUser.PasswordChangedAt.Valid {
		passwordChangedAt = &sqlcUser.PasswordChangedAt.Time
	}

	var lockoutUntil *time.Time
	if sqlcUser.LockoutUntil != nil && sqlcUser.LockoutUntil.Valid {
		lockoutUntil = &sqlcUser.LockoutUntil.Time
	}

	return &User{
		ID:                    sqlcUser.ID,
		TenantID:              sqlcUser.TenantID,
		EntityID:              sqlcUser.EntityID,
		PersonID:              sqlcUser.PersonID,
		EmployeeID:            sqlcUser.EmployeeID,
		Username:              *sqlcUser.Username,
		Email:                 sqlcUser.Email,
		UserType:              sqlcUser.UserType,
		AccountStatus:         *sqlcUser.AccountStatus,
		IsActive:              sqlcUser.IsActive,
		LastLoginAt:           lastLoginAt,
		PasswordChangedAt:     passwordChangedAt,
		FailedLoginAttempts:   *sqlcUser.FailedLoginAttempts,
		LockoutUntil:          lockoutUntil,
		SessionTimeoutMinutes: *sqlcUser.SessionTimeoutMinutes,
		MfaEnabled:            *sqlcUser.MfaEnabled,
		UserAttributes:        userAttributes,
		Settings:              settings,
		CreatedAt:             sqlcUser.CreatedAt,
		UpdatedAt:             sqlcUser.UpdatedAt,
		DeletedAt:             deletedAt,
	}, nil
}

// FromSQLCPerson converts SQLC Person to domain Person
func FromSQLCPerson(sqlcPerson *db.Person) (*Person, error) {
	var securityAttributes map[string]interface{}
	if len(sqlcPerson.SecurityAttributes) > 0 {
		if err := json.Unmarshal(sqlcPerson.SecurityAttributes, &securityAttributes); err != nil {
			return nil, err
		}
	}

	var metadata map[string]interface{}
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

	return &Person{
		ID:                 sqlcPerson.ID,
		TenantID:           sqlcPerson.TenantID,
		EntityID:           sqlcPerson.EntityID,
		PersonType:         sqlcPerson.PersonType,
		FirstName:          sqlcPerson.FirstName,
		LastName:           sqlcPerson.LastName,
		MiddleName:         sqlcPerson.MiddleName,
		Email:              *sqlcPerson.Email,
		Phone:              sqlcPerson.Phone,
		BirthDate:          birthDate,
		NationalID:         sqlcPerson.NationalID,
		TaxID:              sqlcPerson.TaxID,
		Address:            *string(&sqlcPerson.Address),
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
	var salaryInfo map[string]interface{}
	if len(sqlcEmployee.SalaryInfo) > 0 {
		if err := json.Unmarshal(sqlcEmployee.SalaryInfo, &salaryInfo); err != nil {
			return nil, err
		}
	}

	var workSchedule map[string]interface{}
	if len(sqlcEmployee.WorkSchedule) > 0 {
		if err := json.Unmarshal(sqlcEmployee.WorkSchedule, &workSchedule); err != nil {
			return nil, err
		}
	}

	var accessAttributes map[string]interface{}
	if len(sqlcEmployee.AccessAttributes) > 0 {
		if err := json.Unmarshal(sqlcEmployee.AccessAttributes, &accessAttributes); err != nil {
			return nil, err
		}
	}

	var deletedAt *time.Time
	if sqlcEmployee.DeletedAt != nil && sqlcEmployee.DeletedAt.Valid {
		deletedAt = &sqlcEmployee.DeletedAt.Time
	}

	var terminationDate *time.Time
	if sqlcEmployee.TerminationDate != nil && sqlcEmployee.TerminationDate.Valid {
		terminationDate = &sqlcEmployee.TerminationDate.Time
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
		HireDate:         sqlcEmployee.HireDate.Time,
		TerminationDate:  terminationDate,
		SalaryInfo:       salaryInfo,
		EmploymentStatus: sqlcEmployee.EmploymentStatus,
		WorkSchedule:     workSchedule,
		SecurityLevel:    sqlcEmployee.SecurityLevel,
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
		PersonID:         req.PersonID,
		EmployeeNumber:   req.EmployeeNumber,
		EntityID:         req.EntityID,
		PositionTitle:    req.PositionTitle,
		DepartmentID:     req.DepartmentID,
		ManagerID:        req.ManagerID,
		HireDate:         db.Date{Time: req.HireDate, Valid: true},
		SalaryInfo:       salaryInfo,
		EmploymentStatus: req.EmploymentStatus,
		WorkSchedule:     workSchedule,
		SecurityLevel:    req.SecurityLevel,
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
